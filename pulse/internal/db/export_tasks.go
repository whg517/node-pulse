package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// ExportTaskRepository durably persists export tasks so they survive server restarts.
// The in-memory map in internal/export/service.go remains the source of truth for
// hot tasks, but every state change is mirrored here, and pending/processing tasks
// are reloaded on startup.
type ExportTaskRepository interface {
	Create(ctx context.Context, task *models.ExportTask) error
	GetByID(ctx context.Context, id string) (*models.ExportTask, error)
	ListByUser(ctx context.Context, userID string, limit int) ([]*models.ExportTask, error)
	Update(ctx context.Context, task *models.ExportTask) error
	ListByStatuses(ctx context.Context, statuses []string) ([]*models.ExportTask, error)
	// Delete removes an export task record by id. Callers should remove the
	// associated file first (see ExportService.DeleteExport).
	Delete(ctx context.Context, id string) error
}

type exportTaskRepository struct {
	pool *pgxpool.Pool
}

// NewExportTaskRepository creates a new ExportTaskRepository backed by PostgreSQL.
func NewExportTaskRepository(pool *pgxpool.Pool) ExportTaskRepository {
	return &exportTaskRepository{pool: pool}
}

// Create persists a new export task.
func (r *exportTaskRepository) Create(ctx context.Context, task *models.ExportTask) error {
	nodeIDs, err := json.Marshal(task.NodeIDs)
	if err != nil {
		return errors.New("failed to marshal node_ids: " + err.Error())
	}
	metrics, err := json.Marshal(task.Metrics)
	if err != nil {
		return errors.New("failed to marshal metrics: " + err.Error())
	}

	query := `
		INSERT INTO export_tasks
			(id, user_id, node_ids, start_time, end_time, metrics, format, status,
			 file_path, file_size, record_count, error_message, created_at, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
	`

	var completedAt interface{}
	if task.CompletedAt != nil {
		completedAt = *task.CompletedAt
	}
	filePath := nullableString(task.FilePath)
	errMsg := nullableString(task.Error)

	_, err = r.pool.Exec(ctx, query,
		task.ID, task.UserID, nodeIDs, task.StartTime, task.EndTime, metrics,
		task.Format, task.Status, filePath, task.FileSize, task.RecordCount, errMsg,
		task.CreatedAt, completedAt,
	)
	if err != nil {
		return errors.New("failed to create export task: " + err.Error())
	}
	return nil
}

// GetByID retrieves an export task by ID.
func (r *exportTaskRepository) GetByID(ctx context.Context, id string) (*models.ExportTask, error) {
	query := `
		SELECT id, user_id, node_ids, start_time, end_time, metrics, format, status,
		       file_path, file_size, record_count, error_message, created_at, completed_at
		FROM export_tasks WHERE id = $1
	`
	return r.scanTask(ctx, query, id)
}

// ListByUser lists export tasks for a user, newest first.
func (r *exportTaskRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*models.ExportTask, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, user_id, node_ids, start_time, end_time, metrics, format, status,
		       file_path, file_size, record_count, error_message, created_at, completed_at
		FROM export_tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, errors.New("failed to list export tasks: " + err.Error())
	}
	defer rows.Close()
	return r.scanTasks(rows)
}

// ListByStatuses returns tasks in the given statuses (used for recovery on startup).
func (r *exportTaskRepository) ListByStatuses(ctx context.Context, statuses []string) ([]*models.ExportTask, error) {
	if len(statuses) == 0 {
		return []*models.ExportTask{}, nil
	}
	query := `
		SELECT id, user_id, node_ids, start_time, end_time, metrics, format, status,
		       file_path, file_size, record_count, error_message, created_at, completed_at
		FROM export_tasks
		WHERE status = ANY($1)
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, statuses)
	if err != nil {
		return nil, errors.New("failed to list export tasks by status: " + err.Error())
	}
	defer rows.Close()
	return r.scanTasks(rows)
}

// Update mirrors an export task's current state to durable storage.
func (r *exportTaskRepository) Update(ctx context.Context, task *models.ExportTask) error {
	nodeIDs, err := json.Marshal(task.NodeIDs)
	if err != nil {
		return errors.New("failed to marshal node_ids: " + err.Error())
	}
	metrics, err := json.Marshal(task.Metrics)
	if err != nil {
		return errors.New("failed to marshal metrics: " + err.Error())
	}

	query := `
		UPDATE export_tasks SET
			node_ids = $2,
			metrics = $3,
			format = $4,
			status = $5,
			file_path = $6,
			file_size = $7,
			record_count = $8,
			error_message = $9,
			completed_at = $10,
			updated_at = NOW()
		WHERE id = $1
	`

	var completedAt interface{}
	if task.CompletedAt != nil {
		completedAt = *task.CompletedAt
	}
	filePath := nullableString(task.FilePath)
	errMsg := nullableString(task.Error)

	_, err = r.pool.Exec(ctx, query,
		task.ID, nodeIDs, metrics, task.Format, task.Status,
		filePath, task.FileSize, task.RecordCount, errMsg, completedAt,
	)
	if err != nil {
		return errors.New("failed to update export task: " + err.Error())
	}
	return nil
}

// Delete removes an export task row by id.
func (r *exportTaskRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM export_tasks WHERE id = $1`, id)
	if err != nil {
		return errors.New("failed to delete export task: " + err.Error())
	}
	return nil
}

func (r *exportTaskRepository) scanTask(ctx context.Context, query string, args ...interface{}) (*models.ExportTask, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	task, err := scanExportTaskRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("export task not found")
		}
		return nil, err
	}
	return task, nil
}

func (r *exportTaskRepository) scanTasks(rows pgx.Rows) ([]*models.ExportTask, error) {
	tasks := make([]*models.ExportTask, 0)
	for rows.Next() {
		task, err := scanExportTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// scannable is satisfied by both *pgx.Row and pgx.Rows.
type scannable interface {
	Scan(dest ...interface{}) error
}

func scanExportTaskRow(row scannable) (*models.ExportTask, error) {
	var (
		task        models.ExportTask
		nodeIDsJSON []byte
		metricsJSON []byte
		filePath    *string
		errMsg      *string
		completedAt *time.Time
	)
	err := row.Scan(
		&task.ID, &task.UserID, &nodeIDsJSON, &task.StartTime, &task.EndTime,
		&metricsJSON, &task.Format, &task.Status, &filePath, &task.FileSize,
		&task.RecordCount, &errMsg, &task.CreatedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(nodeIDsJSON, &task.NodeIDs); err != nil {
		return nil, errors.New("failed to unmarshal node_ids: " + err.Error())
	}
	if err := json.Unmarshal(metricsJSON, &task.Metrics); err != nil {
		return nil, errors.New("failed to unmarshal metrics: " + err.Error())
	}
	if filePath != nil {
		task.FilePath = *filePath
	}
	if errMsg != nil {
		task.Error = *errMsg
	}
	task.CompletedAt = completedAt
	return &task, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

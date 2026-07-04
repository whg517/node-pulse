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

// ReportScheduleRepository stores recurring report schedules (ADR-001).
type ReportScheduleRepository interface {
	Create(ctx context.Context, s *models.ReportSchedule) error
	ListByOwner(ctx context.Context, ownerUserID string) ([]*models.ReportSchedule, error)
	Update(ctx context.Context, s *models.ReportSchedule) error
	Delete(ctx context.Context, id, ownerUserID string) error
	// ListDue returns enabled schedules whose next_run_at has passed (or is null).
	ListDue(ctx context.Context, now time.Time) ([]*models.ReportSchedule, error)
	// MarkRun records the last/next run timestamps after execution.
	MarkRun(ctx context.Context, id string, lastRun, nextRun time.Time) error
}

type reportScheduleRepo struct {
	pool *pgxpool.Pool
}

// NewReportScheduleRepository creates a repository backed by PostgreSQL.
func NewReportScheduleRepository(pool *pgxpool.Pool) ReportScheduleRepository {
	return &reportScheduleRepo{pool: pool}
}

func (r *reportScheduleRepo) Create(ctx context.Context, s *models.ReportSchedule) error {
	nodeIDs, _ := json.Marshal(s.NodeIDs)
	metrics, _ := json.Marshal(s.Metrics)
	return r.pool.QueryRow(ctx, `
		INSERT INTO report_schedules (owner_user_id, name, frequency, time_of_day, node_ids, metrics, format, recipient_email, enabled, next_run_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`, s.OwnerUserID, s.Name, s.Frequency, s.TimeOfDay, nodeIDs, metrics, s.Format, nullable(s.RecipientEmail), s.Enabled, time.Now()).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *reportScheduleRepo) ListByOwner(ctx context.Context, ownerUserID string) ([]*models.ReportSchedule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_user_id, name, frequency, time_of_day, node_ids, metrics, format, COALESCE(recipient_email,''), enabled, last_run_at, next_run_at, created_at, updated_at
		FROM report_schedules WHERE owner_user_id = $1 ORDER BY created_at DESC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectSchedules(rows)
}

func (r *reportScheduleRepo) ListDue(ctx context.Context, now time.Time) ([]*models.ReportSchedule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_user_id, name, frequency, time_of_day, node_ids, metrics, format, COALESCE(recipient_email,''), enabled, last_run_at, next_run_at, created_at, updated_at
		FROM report_schedules
		WHERE enabled = true AND (next_run_at IS NULL OR next_run_at <= $1)
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectSchedules(rows)
}

func (r *reportScheduleRepo) Update(ctx context.Context, s *models.ReportSchedule) error {
	nodeIDs, _ := json.Marshal(s.NodeIDs)
	metrics, _ := json.Marshal(s.Metrics)
	tag, err := r.pool.Exec(ctx, `
		UPDATE report_schedules SET
			name=$3, frequency=$4, time_of_day=$5, node_ids=$6, metrics=$7, format=$8,
			recipient_email=$9, enabled=$10, updated_at=NOW()
		WHERE id=$1 AND owner_user_id=$2
	`, s.ID, s.OwnerUserID, s.Name, s.Frequency, s.TimeOfDay, nodeIDs, metrics, s.Format, nullable(s.RecipientEmail), s.Enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *reportScheduleRepo) Delete(ctx context.Context, id, ownerUserID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM report_schedules WHERE id=$1 AND owner_user_id=$2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *reportScheduleRepo) MarkRun(ctx context.Context, id string, lastRun, nextRun time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE report_schedules SET last_run_at=$2, next_run_at=$3, updated_at=NOW() WHERE id=$1`, id, lastRun, nextRun)
	return err
}

func collectSchedules(rows pgx.Rows) ([]*models.ReportSchedule, error) {
	var out []*models.ReportSchedule
	for rows.Next() {
		var (
			s       models.ReportSchedule
			nodeIDs []byte
			metrics []byte
			lastRun *time.Time
			nextRun *time.Time
		)
		if err := rows.Scan(&s.ID, &s.OwnerUserID, &s.Name, &s.Frequency, &s.TimeOfDay, &nodeIDs, &metrics, &s.Format, &s.RecipientEmail, &s.Enabled, &lastRun, &nextRun, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if len(nodeIDs) > 0 && string(nodeIDs) != "null" {
			if err := json.Unmarshal(nodeIDs, &s.NodeIDs); err != nil {
				return nil, errors.New("unmarshal node_ids: " + err.Error())
			}
		}
		if len(metrics) > 0 && string(metrics) != "null" {
			if err := json.Unmarshal(metrics, &s.Metrics); err != nil {
				return nil, errors.New("unmarshal metrics: " + err.Error())
			}
		}
		s.LastRunAt = lastRun
		s.NextRunAt = nextRun
		out = append(out, &s)
	}
	return out, rows.Err()
}

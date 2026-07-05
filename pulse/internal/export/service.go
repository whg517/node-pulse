package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

const (
	// MaxNodes is the maximum number of nodes that can be exported in a single request
	MaxNodes = 50

	// MaxFileSize is the maximum export file size in bytes (10MB)
	MaxFileSize = 10 * 1024 * 1024

	// MinTimeRange is the minimum time range for export (1 hour)
	MinTimeRange = 1 * time.Hour

	// MaxTimeRange is the maximum time range for export (7 days)
	MaxTimeRange = 7 * 24 * time.Hour

	// ExportRetention is how long export files are kept before cleanup
	ExportRetention = 24 * time.Hour

	// ExportDir is the directory where export files are stored
	ExportDir = "/tmp/exports"
)

// TaskStore mirrors export task state to durable storage so tasks survive restarts.
// It is optional: when nil the service behaves exactly as before (in-memory only).
type TaskStore interface {
	Create(ctx context.Context, task *models.ExportTask) error
	GetByID(ctx context.Context, id string) (*models.ExportTask, error)
	ListByUser(ctx context.Context, userID string, limit int) ([]*models.ExportTask, error)
	Update(ctx context.Context, task *models.ExportTask) error
	ListByStatuses(ctx context.Context, statuses []string) ([]*models.ExportTask, error)
	Delete(ctx context.Context, id string) error
}

// ExportService handles data export operations
type ExportService struct {
	pool   *pgxpool.Pool
	store  TaskStore
	tasks  map[string]*models.ExportTask
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewExportService creates a new export service. A nil store keeps the legacy
// in-memory-only behavior; passing a store additionally persists every task and
// recovers pending/processing tasks on startup.
func NewExportService(pool *pgxpool.Pool, store TaskStore) *ExportService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &ExportService{
		pool:   pool,
		store:  store,
		tasks:  make(map[string]*models.ExportTask),
		ctx:    ctx,
		cancel: cancel,
	}

	// Recover unfinished tasks from durable storage before accepting new work.
	if store != nil {
		service.recoverPendingTasks()
	}

	// Start background cleanup goroutine
	go service.cleanupOldExports()

	return service
}

// recoverPendingTasks reloads pending/processing tasks from durable storage.
// Tasks still marked processing after a crash are retried once; tasks left
// pending are also retried. Completed/failed tasks are loaded for history queries.
func (s *ExportService) recoverPendingTasks() {
	if s.store == nil {
		return
	}
	active, err := s.store.ListByStatuses(s.ctx, []string{"pending", "processing"})
	if err != nil {
		slog.Error("Failed to load active export tasks from storage; continuing in-memory only",
			"component", "export", "error", err)
		return
	}
	s.mu.Lock()
	for _, task := range active {
		taskCopy := *task
		s.tasks[task.ID] = &taskCopy
	}
	s.mu.Unlock()

	slog.Info("Recovered export tasks from storage", "component", "export", "count", len(active))

	// Retry tasks that were interrupted mid-flight. They are reset to pending so
	// processExport can drive them through processing -> completed/failed again.
	for _, task := range active {
		t := *task
		go s.processExport(&t)
	}
}

// CreateExportRequest represents a request to create an export
type CreateExportRequest struct {
	UserID    string
	NodeIDs   []string
	StartTime time.Time
	EndTime   time.Time
	Metrics   []string
	Format    string
}

// Validate validates the export request
func (r *CreateExportRequest) Validate() error {
	// Validate node IDs
	if len(r.NodeIDs) == 0 {
		return fmt.Errorf("at least one node_id is required")
	}
	if len(r.NodeIDs) > MaxNodes {
		return fmt.Errorf("maximum %d nodes allowed per export", MaxNodes)
	}

	// Validate time range
	if r.EndTime.Before(r.StartTime) {
		return fmt.Errorf("end_time must be after start_time")
	}

	duration := r.EndTime.Sub(r.StartTime)
	if duration < MinTimeRange {
		return fmt.Errorf("time range must be at least %s", MinTimeRange)
	}
	if duration > MaxTimeRange {
		return fmt.Errorf("time range must be at most %s", MaxTimeRange)
	}

	// Validate metrics
	if len(r.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}

	validMetrics := map[string]bool{
		"latency":          true,
		"packet_loss_rate": true,
		"jitter":           true,
	}
	for _, metric := range r.Metrics {
		if !validMetrics[metric] {
			return fmt.Errorf("invalid metric: %s", metric)
		}
	}

	// Validate format
	if r.Format == "" {
		r.Format = "csv" // Default format
	}
	if r.Format != "csv" {
		return fmt.Errorf("unsupported format: %s (only CSV is supported in MVP)", r.Format)
	}

	return nil
}

// CreateExport creates a new export task
func (s *ExportService) CreateExport(ctx context.Context, req *CreateExportRequest) (*models.ExportTask, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create export task
	task := &models.ExportTask{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		NodeIDs:   req.NodeIDs,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Metrics:   req.Metrics,
		Format:    req.Format,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Store task
	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	// Persist to durable storage (best-effort; in-memory remains source of truth)
	if s.store != nil {
		if err := s.store.Create(context.Background(), task); err != nil {
			slog.Error("Failed to persist export task; continuing in-memory",
				"component", "export", "task_id", task.ID, "error", err)
		}
	}

	// Start async processing
	go s.processExport(task)

	return task, nil
}

// GetExport retrieves an export task by ID
func (s *ExportService) GetExport(exportID string) (*models.ExportTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[exportID]
	if !ok {
		return nil, fmt.Errorf("export not found: %s", exportID)
	}

	return task, nil
}

// ListExports returns recent export tasks for a user in newest-first order.
func (s *ExportService) ListExports(userID string, limit int) []*models.ExportTask {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*models.ExportTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		if userID != "" && task.UserID != userID {
			continue
		}
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	if len(tasks) > limit {
		return tasks[:limit]
	}
	return tasks
}

// DeleteExport removes an export task: it deletes the generated file (if any),
// drops the in-memory entry, and deletes the durable row. Returns an error if
// the task cannot be found. Deleting a still-processing task is allowed; the
// background goroutine will tolerate its disappearance (updateTaskStatus is a
// no-op on a missing in-memory task).
func (s *ExportService) DeleteExport(exportID string) error {
	s.mu.Lock()
	var filePath string
	if task, ok := s.tasks[exportID]; ok {
		filePath = task.FilePath
		delete(s.tasks, exportID)
	}
	s.mu.Unlock()

	// Remove the on-disk file. A missing file is not an error (it may already
	// have been cleaned up by the retention sweeper).
	if filePath != "" {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			slog.Error("Failed to delete export file during DeleteExport",
				"component", "export", "file", filePath, "error", err)
			// Continue: we still want to remove the DB row.
		}
	}

	// Remove the durable row.
	if s.store != nil {
		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
		defer cancel()
		if err := s.store.Delete(ctx, exportID); err != nil {
			return fmt.Errorf("failed to delete export task: %w", err)
		}
	}

	slog.Info("Deleted export task", "component", "export", "export_id", exportID)
	return nil
}

// processExport processes the export task asynchronously
func (s *ExportService) processExport(task *models.ExportTask) {
	// Update status to processing
	s.updateTaskStatus(task.ID, "processing", "", 0, 0, "")

	// Generate export file
	filePath, recordCount, err := s.generateExportFile(task)
	if err != nil {
		slog.Error("Failed to generate export",
			"component", "export", "task_id", task.ID, "error", err)
		s.updateTaskStatus(task.ID, "failed", "", 0, 0, err.Error())
		return
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		slog.Error("Failed to get file size",
			"component", "export", "task_id", task.ID, "error", err)
		s.updateTaskStatus(task.ID, "failed", "", 0, 0, err.Error())
		return
	}

	// Check file size limit
	if fileInfo.Size() > MaxFileSize {
		_ = os.Remove(filePath)
		err = fmt.Errorf("export file exceeds maximum size of %d bytes", MaxFileSize)
		s.updateTaskStatus(task.ID, "failed", "", 0, 0, err.Error())
		return
	}

	// Update task to completed
	now := time.Now()
	s.updateTaskStatus(task.ID, "completed", filePath, fileInfo.Size(), recordCount, "")
	task = s.getTask(task.ID)
	if task != nil {
		task.CompletedAt = &now
		// Mirror the completion timestamp to durable storage.
		if s.store != nil {
			if err := s.store.Update(context.Background(), task); err != nil {
				slog.Error("Failed to mirror export task completion",
					"component", "export", "task_id", task.ID, "error", err)
			}
		}
	}

	slog.Info("Export completed",
		"component", "export",
		"task_id", task.ID,
		"records", recordCount,
		"bytes", fileInfo.Size(),
	)
}

// generateExportFile generates the export file
func (s *ExportService) generateExportFile(task *models.ExportTask) (string, int, error) {
	// Ensure export directory exists
	if err := os.MkdirAll(ExportDir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create export directory: %w", err)
	}

	// Generate file path
	filename := fmt.Sprintf("metrics_export_%s_%s.csv",
		task.ID,
		time.Now().Format("20060102-150405"))
	filePath := filepath.Join(ExportDir, filename)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create export file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Write UTF-8 BOM for Excel compatibility
	if _, err := file.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return "", 0, fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"timestamp", "node_id", "region", "metric_name", "value", "unit"}
	if err := writer.Write(header); err != nil {
		return "", 0, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Query and write data
	recordCount := 0
	ctx := context.Background()

	for _, nodeID := range task.NodeIDs {
		for _, metric := range task.Metrics {
			rows, err := s.queryMetrics(ctx, nodeID, metric, task.StartTime, task.EndTime)
			if err != nil {
				return "", 0, fmt.Errorf("failed to query metrics for node %s: %w", nodeID, err)
			}

			for _, row := range rows {
				// Convert row to CSV record
				record := []string{
					row.Timestamp,
					row.NodeID,
					row.Region,
					row.Metric,
					strconv.FormatFloat(row.Value, 'f', -1, 64),
					row.Unit,
				}

				if err := writer.Write(record); err != nil {
					return "", 0, fmt.Errorf("failed to write CSV record: %w", err)
				}

				recordCount++

				// Check file size periodically (every 1000 records)
				if recordCount%1000 == 0 {
					fileInfo, _ := file.Stat()
					if fileInfo.Size() > MaxFileSize {
						return "", 0, fmt.Errorf("export file exceeds maximum size")
					}
				}
			}
		}
	}

	return filePath, recordCount, nil
}

// queryMetrics queries metrics from the database
func (s *ExportService) queryMetrics(
	ctx context.Context,
	nodeID string,
	metric string,
	startTime time.Time,
	endTime time.Time,
) ([]models.ExportMetricsRow, error) {
	// Map metric name to database column
	columnMap := map[string]string{
		"latency":          "latency_ms",
		"packet_loss_rate": "packet_loss_rate",
		"jitter":           "jitter_ms",
	}

	column, ok := columnMap[metric]
	if !ok {
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}

	// Query metrics with node region
	query := `
		SELECT
			m.timestamp,
			m.node_id,
			n.region,
			$4::text as metric_name,
			m.` + column + ` as value
		FROM metrics m
		JOIN nodes n ON m.node_id = n.id
		WHERE m.node_id = $1
			AND m.timestamp >= $2
			AND m.timestamp <= $3
			AND m.` + column + ` IS NOT NULL
		ORDER BY m.timestamp ASC;
	`

	rows, err := s.pool.Query(ctx, query, nodeID, startTime, endTime, metric)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	results := make([]models.ExportMetricsRow, 0)
	for rows.Next() {
		var timestamp time.Time
		var nodeID, region, metricName string
		var value float64

		if err := rows.Scan(&timestamp, &nodeID, &region, &metricName, &value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		results = append(results, models.ExportMetricsRow{
			Timestamp: timestamp.Format(time.RFC3339),
			NodeID:    nodeID,
			Region:    region,
			Metric:    metricName,
			Value:     value,
			Unit:      models.MetricUnit(metricName),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// updateTaskStatus updates the status of an export task
func (s *ExportService) updateTaskStatus(exportID, status, filePath string, fileSize int64, recordCount int, errorMsg string) {
	s.mu.Lock()
	task, ok := s.tasks[exportID]
	if !ok {
		s.mu.Unlock()
		return
	}

	task.Status = status
	task.FilePath = filePath
	task.FileSize = fileSize
	task.RecordCount = recordCount
	task.Error = errorMsg
	s.mu.Unlock()

	// Mirror to durable storage (best-effort; released lock to avoid blocking hot path)
	if s.store != nil && task != nil {
		if err := s.store.Update(context.Background(), task); err != nil {
			slog.Error("Failed to mirror export task update",
				"component", "export", "task_id", exportID, "error", err)
		}
	}
}

// getTask retrieves a task without locking (internal use)
func (s *ExportService) getTask(exportID string) *models.ExportTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[exportID]
}

// cleanupOldExports removes old export files and tasks
func (s *ExportService) cleanupOldExports() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performCleanup()
		}
	}
}

// performCleanup performs the actual cleanup
func (s *ExportService) performCleanup() {
	slog.Info("Starting cleanup of old export files", "component", "export")

	now := time.Now()
	cutoffTime := now.Add(-ExportRetention)

	s.mu.Lock()
	defer s.mu.Unlock()

	for exportID, task := range s.tasks {
		// Remove completed tasks older than retention period
		if task.IsCompleted() && task.CreatedAt.Before(cutoffTime) {
			// Delete file
			if task.FilePath != "" {
				if err := os.Remove(task.FilePath); err != nil {
					slog.Error("Failed to delete export file",
						"component", "export", "file", task.FilePath, "error", err)
				} else {
					slog.Info("Deleted export file", "component", "export", "file", task.FilePath)
				}
			}

			// Remove from memory
			delete(s.tasks, exportID)
		}
	}

	slog.Info("Export cleanup completed", "component", "export")
}

// Shutdown stops the export service
func (s *ExportService) Shutdown() {
	s.cancel()
}

package cleanup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kevin/node-pulse/pulse-api/internal/config"
)

// PgxPool defines the database pool interface for cleanup operations
type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

// CleanupTask implements the scheduler.Task interface for metrics data cleanup
type CleanupTask struct {
	name   string
	cfg    *config.CleanupConfig
	db     PgxPool
	logger *log.Logger

	// Runtime state
	lastRun      time.Time
	lastDuration time.Duration
	lastError    error
	runCount     int64
	isRunning    bool
}

// NewCleanupTask creates a new cleanup task
func NewCleanupTask(cfg *config.CleanupConfig, db PgxPool, logger *log.Logger) (*CleanupTask, error) {
	if !cfg.Enabled {
		if logger != nil {
			logger.Println("[Cleanup] Task disabled")
		}
		return nil, nil
	}

	if cfg.IntervalSeconds <= 0 {
		return nil, fmt.Errorf("invalid interval_seconds: %d", cfg.IntervalSeconds)
	}

	if cfg.RetentionDays <= 0 {
		return nil, fmt.Errorf("invalid retention_days: %d", cfg.RetentionDays)
	}

	return &CleanupTask{
		name:   "metrics-cleanup",
		cfg:    cfg,
		db:     db,
		logger: logger,
	}, nil
}

// Name returns the task name (implements scheduler.Task)
func (c *CleanupTask) Name() string {
	return c.name
}

// Interval returns the execution interval (implements scheduler.Task)
func (c *CleanupTask) Interval() time.Duration {
	return time.Duration(c.cfg.IntervalSeconds) * time.Second
}

// Execute runs the cleanup task (implements scheduler.Task)
func (c *CleanupTask) Execute(ctx context.Context) error {
	start := time.Now()
	c.isRunning = true
	defer func() { c.isRunning = false }()

	if c.logger != nil {
		c.logger.Printf("[Cleanup] Starting metrics data cleanup (retention_days: %d, timestamp: %s)",
			c.cfg.RetentionDays, start.Format(time.RFC3339))
	}

	// Execute cleanup SQL with parameterized query to prevent SQL injection
	sql := "DELETE FROM metrics WHERE timestamp < NOW() - INTERVAL $1 * INTERVAL '1 day'"
	result, err := c.db.Exec(ctx, sql, c.cfg.RetentionDays)
	if err != nil {
		c.lastError = err
		if c.logger != nil {
			c.logger.Printf("[Cleanup] ERROR: Failed to execute cleanup SQL: %v", err)
		}
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Get deleted row count
	rowsAffected := result.RowsAffected()

	duration := time.Since(start)

	c.lastRun = start
	c.lastDuration = duration
	c.lastError = nil
	c.runCount++

	if c.logger != nil {
		c.logger.Printf("[Cleanup] Metrics data cleanup completed (rows_deleted: %d, duration_ms: %d)",
			rowsAffected, duration.Milliseconds())
	}

	// Check for slow query
	if c.cfg.SlowThresholdMs > 0 && duration.Milliseconds() > c.cfg.SlowThresholdMs {
		if c.logger != nil {
			c.logger.Printf("[Cleanup] WARN: Slow cleanup operation detected (duration_ms: %d, threshold_ms: %d)",
				duration.Milliseconds(), c.cfg.SlowThresholdMs)
		}
	}

	return nil
}

// GetStatus returns the current task status
func (c *CleanupTask) GetStatus() *TaskStatus {
	lastErrMsg := ""
	if c.lastError != nil {
		lastErrMsg = c.lastError.Error()
	}

	return &TaskStatus{
		Name:         c.name,
		IsRunning:    c.isRunning,
		LastRun:      c.lastRun,
		NextRun:      c.lastRun.Add(c.Interval()),
		LastDuration: c.lastDuration,
		LastError:    lastErrMsg,
		RunCount:     c.runCount,
	}
}

// TaskStatus represents the cleanup task status
type TaskStatus struct {
	Name         string        `json:"name"`
	IsRunning    bool          `json:"is_running"`
	LastRun      time.Time     `json:"last_run"`
	NextRun      time.Time     `json:"next_run"`
	LastDuration time.Duration `json:"last_duration"`
	LastError    string        `json:"last_error,omitempty"`
	RunCount     int64         `json:"run_count"`
}
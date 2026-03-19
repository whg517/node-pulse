package cleanup

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/whg517/node-pulse/pulse/internal/config"
)

// PgxPool defines the database pool interface for cleanup operations
type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

// CleanupTask implements the scheduler.Task interface for metrics data cleanup
type CleanupTask struct {
	name string
	cfg  *config.CleanupConfig
	db   PgxPool

	// Runtime state
	lastRun      time.Time
	lastDuration time.Duration
	lastError    error
	runCount     int64
	isRunning    bool
}

// NewCleanupTask creates a new cleanup task.
// Returns nil (without error) when cleanup is disabled in config.
func NewCleanupTask(cfg *config.CleanupConfig, db PgxPool) (*CleanupTask, error) {
	if !cfg.Enabled {
		slog.Info("Cleanup task disabled", "component", "cleanup")
		return nil, nil
	}

	if cfg.IntervalSeconds <= 0 {
		return nil, fmt.Errorf("invalid interval_seconds: %d", cfg.IntervalSeconds)
	}

	if cfg.RetentionDays <= 0 {
		return nil, fmt.Errorf("invalid retention_days: %d", cfg.RetentionDays)
	}

	return &CleanupTask{
		name: "metrics-cleanup",
		cfg:  cfg,
		db:   db,
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

	slog.Info("Starting metrics data cleanup",
		"component", "cleanup",
		"retention_days", c.cfg.RetentionDays,
	)

	// Execute cleanup SQL with parameterized query to prevent SQL injection
	sql := "DELETE FROM metrics WHERE timestamp < NOW() - INTERVAL '1 day' * $1"
	result, err := c.db.Exec(ctx, sql, c.cfg.RetentionDays)
	if err != nil {
		c.lastError = err
		slog.Error("Failed to execute cleanup SQL", "component", "cleanup", "error", err)
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Get deleted row count
	rowsAffected := result.RowsAffected()

	duration := time.Since(start)

	c.lastRun = start
	c.lastDuration = duration
	c.lastError = nil
	c.runCount++

	slog.Info("Metrics data cleanup completed",
		"component", "cleanup",
		"rows_deleted", rowsAffected,
		"duration_ms", duration.Milliseconds(),
	)

	// Check for slow query
	if c.cfg.SlowThresholdMs > 0 && duration.Milliseconds() > c.cfg.SlowThresholdMs {
		slog.Warn("Slow cleanup operation detected",
			"component", "cleanup",
			"duration_ms", duration.Milliseconds(),
			"threshold_ms", c.cfg.SlowThresholdMs,
		)
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

package server

import (
	"log/slog"

	"github.com/whg517/node-pulse/pulse/internal/cleanup"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/scheduler"
	"github.com/whg517/node-pulse/pulse/internal/suppression"
)

// TaskRegistry manages scheduled task registration
type TaskRegistry struct {
	scheduler   scheduler.Scheduler
	database    *db.Database
	cleanupTask *cleanup.CleanupTask
}

// NewTaskRegistry creates a new task registry
func NewTaskRegistry(sched scheduler.Scheduler, database *db.Database) *TaskRegistry {
	return &TaskRegistry{
		scheduler: sched,
		database:  database,
	}
}

// RegisterAll registers all scheduled tasks
func (r *TaskRegistry) RegisterAll() error {
	if r.database == nil || r.database.Pool == nil {
		slog.Warn("Skipping task registration (no database)", "component", "registry")
		return nil
	}

	if err := r.registerCleanupTask(); err != nil {
		return err
	}

	if err := r.registerSuppressionCleanup(); err != nil {
		return err
	}

	return nil
}

// registerCleanupTask creates and registers the data cleanup task
func (r *TaskRegistry) registerCleanupTask() error {
	cfg := config.MustLoad()
	cleanupConfig := &config.CleanupConfig{
		Enabled:         cfg.Cleanup.Enabled,
		IntervalSeconds: cfg.Cleanup.IntervalSeconds,
		RetentionDays:   cfg.Cleanup.RetentionDays,
		SlowThresholdMs: cfg.Cleanup.SlowThresholdMs,
	}

	if !cleanupConfig.Enabled {
		slog.Warn("Cleanup task disabled in configuration", "component", "registry")
		return nil
	}

	var err error
	r.cleanupTask, err = cleanup.NewCleanupTask(cleanupConfig, r.database.Pool)
	if err != nil {
		return err
	}

	if err := r.scheduler.RegisterTask(r.cleanupTask); err != nil {
		return err
	}

	slog.Info("Cleanup task registered",
		"component", "registry",
		"interval_seconds", cleanupConfig.IntervalSeconds,
		"retention_days", cleanupConfig.RetentionDays,
	)
	return nil
}

// registerSuppressionCleanup creates and registers the suppression cleanup task
func (r *TaskRegistry) registerSuppressionCleanup() error {
	suppressionCleanupTask := suppression.NewCleanupTask(db.NewAlertSuppressionsQuerier(r.database.Pool))
	if err := r.scheduler.RegisterTask(suppressionCleanupTask); err != nil {
		return err
	}

	slog.Info("Suppression cleanup task registered", "component", "registry", "interval", "1h")
	return nil
}

package server

import (
	"log"

	"github.com/kevin/node-pulse/pulse-api/internal/cleanup"
	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/scheduler"
	"github.com/kevin/node-pulse/pulse-api/internal/suppression"
)

// TaskRegistry manages scheduled task registration
type TaskRegistry struct {
	scheduler  scheduler.Scheduler
	database   *db.Database
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
		log.Println("[WARN] [Registry] Skipping task registration (no database)")
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
		log.Println("[WARN] [Registry] Cleanup task disabled in configuration")
		return nil
	}

	var err error
	r.cleanupTask, err = cleanup.NewCleanupTask(cleanupConfig, r.database.Pool, log.Default())
	if err != nil {
		return err
	}

	if err := r.scheduler.RegisterTask(r.cleanupTask); err != nil {
		return err
	}

	log.Printf("[INFO] [Registry] Cleanup task registered (interval: %ds, retention: %ddays)",
		cleanupConfig.IntervalSeconds, cleanupConfig.RetentionDays)
	return nil
}

// registerSuppressionCleanup creates and registers the suppression cleanup task
func (r *TaskRegistry) registerSuppressionCleanup() error {
	suppressionCleanupTask := suppression.NewCleanupTask(db.NewAlertSuppressionsQuerier(r.database.Pool))
	if err := r.scheduler.RegisterTask(suppressionCleanupTask); err != nil {
		return err
	}

	log.Println("[INFO] [Registry] Suppression cleanup task registered (interval: 1h)")
	return nil
}

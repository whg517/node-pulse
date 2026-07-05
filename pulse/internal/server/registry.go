package server

import (
	"log/slog"

	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/cleanup"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/export"
	"github.com/whg517/node-pulse/pulse/internal/notify"
	"github.com/whg517/node-pulse/pulse/internal/scheduler"
	"github.com/whg517/node-pulse/pulse/internal/suppression"
)

// TaskRegistry manages scheduled task registration
type TaskRegistry struct {
	scheduler     scheduler.Scheduler
	database      *db.Database
	cleanupTask   *cleanup.CleanupTask
	exportService *export.ExportService
	mailer        notify.Sender
}

// NewTaskRegistry creates a new task registry
func NewTaskRegistry(sched scheduler.Scheduler, database *db.Database) *TaskRegistry {
	return &TaskRegistry{
		scheduler: sched,
		database:  database,
	}
}

// WithExportService attaches the export service (for report scheduling).
func (r *TaskRegistry) WithExportService(svc *export.ExportService) *TaskRegistry {
	r.exportService = svc
	return r
}

// WithMailer attaches the email sender (for report delivery).
func (r *TaskRegistry) WithMailer(m notify.Sender) *TaskRegistry {
	r.mailer = m
	return r
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

	if err := r.registerReportScheduleTask(); err != nil {
		return err
	}

	if err := r.registerAuthCleanupTask(); err != nil {
		return err
	}

	return nil
}

// registerAuthCleanupTask wires the periodic cleanup of auth-related tables
// (refresh_tokens, token_blacklist, rate_limits, auth_audit_logs, api_keys).
// This closes the O-G1 gap from docs/user-journey.md §23.1: these tables
// previously grew without bound while authentication.md claimed daily cleanup.
func (r *TaskRegistry) registerAuthCleanupTask() error {
	cfg := config.MustLoad()
	intervalSeconds := cfg.Cleanup.IntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400 // daily, matches authentication.md
	}
	job := auth.NewCleanupJob(r.database.Pool, intervalSeconds, cfg.Cleanup.RetentionDays)
	task := newAuthCleanupTask(job, intervalSeconds)
	if err := r.scheduler.RegisterTask(task); err != nil {
		return err
	}
	slog.Info("Auth cleanup task registered",
		"component", "registry",
		"interval_seconds", intervalSeconds,
		"retention_days", cfg.Cleanup.RetentionDays,
	)
	return nil
}

// registerReportScheduleTask wires the recurring-report runner (ADR-001).
func (r *TaskRegistry) registerReportScheduleTask() error {
	if r.exportService == nil || r.mailer == nil {
		slog.Info("Skipping report schedule task (export service or mailer not configured)", "component", "registry")
		return nil
	}
	store := db.NewReportScheduleRepository(r.database.Pool)
	runner := export.NewReportScheduleRunner(
		store,
		r.exportService,
		r.mailer,
		nil, // owner-email lookup optional; recipient_email is preferred
		r.exportService.GetExport,
	)
	if err := r.scheduler.RegisterTask(runner); err != nil {
		return err
	}
	slog.Info("Registered report schedule task", "component", "registry", "interval", runner.Interval())
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

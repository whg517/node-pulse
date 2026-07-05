package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/auth"
)

// authCleanupTask adapts *auth.CleanupJob to the scheduler.Task interface so
// the periodic cleanup of refresh tokens, token blacklist, rate limits, audit
// logs and expired API keys runs alongside the other scheduled tasks.
//
// This closes the O-G1 gap documented in docs/user-journey.md: the cleanup
// functions existed in auth/cleanup_job.go but were never wired to the
// scheduler, so auth_audit_logs / refresh_tokens / sessions / api_keys grew
// without bound. The scheduler itself tracks lastRun/lastError/runCount, so
// this adapter only needs to invoke RunAll on each tick.
type authCleanupTask struct {
	job      *auth.CleanupJob
	interval time.Duration
}

// newAuthCleanupTask constructs the task. intervalSeconds governs the cadence
// (default 24h, matching the "daily cleanup" claim in authentication.md).
func newAuthCleanupTask(job *auth.CleanupJob, intervalSeconds int) *authCleanupTask {
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}
	return &authCleanupTask{
		job:      job,
		interval: time.Duration(intervalSeconds) * time.Second,
	}
}

// Name implements scheduler.Task.
func (t *authCleanupTask) Name() string { return "auth-cleanup" }

// Interval implements scheduler.Task.
func (t *authCleanupTask) Interval() time.Duration { return t.interval }

// Execute implements scheduler.Task. RunAll performs the five cleanup steps
// (tokens, blacklist, rate limits, audit logs, API keys) and logs each failure
// internally; it does not return an error, so neither do we. A non-nil return
// would otherwise be captured by the scheduler as taskState.lastError.
func (t *authCleanupTask) Execute(_ context.Context) error {
	t.job.RunAll()
	slog.Debug("auth cleanup tick finished", "component", "auth_cleanup", "interval", t.interval)
	return nil
}

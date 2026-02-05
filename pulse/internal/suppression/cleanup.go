package suppression

import (
	"context"
	"log/slog"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/db"
)

// CleanupTask handles cleanup of expired alert suppressions
type CleanupTask struct {
	querier db.AlertSuppressionsQuerier
}

// NewCleanupTask creates a new SuppressionCleanupTask
func NewCleanupTask(querier db.AlertSuppressionsQuerier) *CleanupTask {
	return &CleanupTask{
		querier: querier,
	}
}

// Name returns the task name
func (t *CleanupTask) Name() string {
	return "suppression-cleanup"
}

// Execute runs the cleanup task
func (t *CleanupTask) Execute(ctx context.Context) error {
	deleted, err := t.querier.DeleteExpiredSuppressions(ctx)
	if err != nil {
		slog.Error("Failed to cleanup expired suppressions", "error", err)
		return err
	}

	if deleted > 0 {
		slog.Info("Cleaned up expired suppressions",
			"count", deleted,
			"timestamp", time.Now().Format(time.RFC3339))
	}

	return nil
}

// Interval returns the task execution interval (hourly)
func (t *CleanupTask) Interval() time.Duration {
	return 1 * time.Hour
}

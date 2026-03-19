package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CleanupJob handles periodic cleanup of expired refresh tokens
type CleanupJob struct {
	pool          *pgxpool.Pool
	interval      time.Duration
	retentionDays int
	stopChan      chan struct{}
	batchSize     int
}

// NewCleanupJob creates a new cleanup job for refresh tokens
func NewCleanupJob(pool *pgxpool.Pool, intervalSeconds, retentionDays int) *CleanupJob {
	return &CleanupJob{
		pool:          pool,
		interval:      time.Duration(intervalSeconds) * time.Second,
		retentionDays: retentionDays,
		stopChan:      make(chan struct{}),
		batchSize:     1000, // Delete 1000 records at a time to avoid table locking
	}
}

// Start begins the cleanup job in a background goroutine
func (j *CleanupJob) Start() {
	slog.Info("Starting refresh token cleanup job",
		"component", "cleanup_job",
		"interval", j.interval,
		"retention_days", j.retentionDays,
	)

	go func() {
		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()

		// Run immediately on start
		j.Run()

		for {
			select {
			case <-ticker.C:
				j.Run()
			case <-j.stopChan:
				slog.Info("Stopping refresh token cleanup job", "component", "cleanup_job")
				return
			}
		}
	}()
}

// Stop gracefully stops the cleanup job
func (j *CleanupJob) Stop() {
	close(j.stopChan)
}

// Run executes the cleanup job once
func (j *CleanupJob) Run() {
	startTime := time.Now()
	ctx := context.Background()

	totalDeleted, err := j.cleanupExpiredTokens(ctx)
	elapsed := time.Since(startTime)

	if err != nil {
		slog.Error("Error cleaning up expired tokens", "component", "cleanup_job", "error", err)
		return
	}

	if totalDeleted > 0 {
		slog.Info("Cleaned up expired tokens",
			"component", "cleanup_job",
			"count", totalDeleted,
			"duration", elapsed,
		)
	}
}

// cleanupExpiredTokens deletes expired refresh tokens in batches
func (j *CleanupJob) cleanupExpiredTokens(ctx context.Context) (int, error) {
	totalDeleted := 0

	for {
		// Delete expired tokens in batches
		result, err := j.pool.Exec(ctx, `
			DELETE FROM refresh_tokens
			WHERE token_id IN (
				SELECT token_id
				FROM refresh_tokens
				WHERE expires_at <= NOW()
				LIMIT $1
			)
		`, j.batchSize)

		if err != nil {
			return totalDeleted, err
		}

		rowsDeleted := result.RowsAffected()
		totalDeleted += int(rowsDeleted)

		// If fewer than batchSize rows were deleted, we're done
		if rowsDeleted < int64(j.batchSize) {
			break
		}

		// Add a small delay to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	return totalDeleted, nil
}

// DeleteAllTokensForUser deletes all refresh tokens for a specific user
func (j *CleanupJob) DeleteAllTokensForUser(ctx context.Context, userID string) error {
	_, err := j.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

// CleanupTokenBlacklist removes expired entries from token_blacklist
func (j *CleanupJob) CleanupTokenBlacklist(ctx context.Context) error {
	result, err := j.pool.Exec(ctx, `
		DELETE FROM token_blacklist
		WHERE expires_at <= NOW()
	`)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		slog.Info("Cleaned up expired blacklist entries",
			"component", "cleanup_job", "count", rowsAffected)
	}
	return nil
}

// CleanupRateLimits removes old rate limit entries
func (j *CleanupJob) CleanupRateLimits(ctx context.Context, retentionHours int) error {
	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour)
	result, err := j.pool.Exec(ctx, `
		DELETE FROM rate_limits
		WHERE window_start < $1
	`, cutoff)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		slog.Info("Cleaned up old rate limit entries",
			"component", "cleanup_job", "count", rowsAffected)
	}
	return nil
}

// CleanupAuditLogs removes old audit log entries
func (j *CleanupJob) CleanupAuditLogs(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	result, err := j.pool.Exec(ctx, `
		DELETE FROM auth_audit_logs
		WHERE created_at < $1
	`, cutoff)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		slog.Info("Cleaned up old audit log entries",
			"component", "cleanup_job", "count", rowsAffected)
	}
	return nil
}

// CleanupExpiredAPIKeys removes expired/inactive API keys
func (j *CleanupJob) CleanupExpiredAPIKeys(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	result, err := j.pool.Exec(ctx, `
		DELETE FROM api_keys
		WHERE (is_active = false AND created_at < $1)
		OR (expires_at <= NOW())
	`, cutoff)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		slog.Info("Cleaned up expired/inactive API keys",
			"component", "cleanup_job", "count", rowsAffected)
	}
	return nil
}

// RunAll executes all cleanup jobs
func (j *CleanupJob) RunAll() {
	startTime := time.Now()
	ctx := context.Background()

	slog.Info("Starting comprehensive cleanup", "component", "cleanup_job")

	// Cleanup refresh tokens
	if _, err := j.cleanupExpiredTokens(ctx); err != nil {
		slog.Error("Error cleaning up refresh tokens", "component", "cleanup_job", "error", err)
	}

	// Cleanup token blacklist
	if err := j.CleanupTokenBlacklist(ctx); err != nil {
		slog.Error("Error cleaning up token blacklist", "component", "cleanup_job", "error", err)
	}

	// Cleanup rate limits (keep for 24 hours)
	if err := j.CleanupRateLimits(ctx, 24); err != nil {
		slog.Error("Error cleaning up rate limits", "component", "cleanup_job", "error", err)
	}

	// Cleanup audit logs (keep for 90 days)
	if err := j.CleanupAuditLogs(ctx, 90); err != nil {
		slog.Error("Error cleaning up audit logs", "component", "cleanup_job", "error", err)
	}

	// Cleanup expired API keys (keep inactive keys for 30 days)
	if err := j.CleanupExpiredAPIKeys(ctx, 30); err != nil {
		slog.Error("Error cleaning up API keys", "component", "cleanup_job", "error", err)
	}

	slog.Info("Comprehensive cleanup completed",
		"component", "cleanup_job",
		"duration", time.Since(startTime),
	)
}

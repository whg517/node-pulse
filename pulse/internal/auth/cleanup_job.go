package auth

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CleanupJob handles periodic cleanup of expired refresh tokens
type CleanupJob struct {
	pool           *pgxpool.Pool
	interval       time.Duration
	retentionDays  int
	stopChan       chan struct{}
	batchSize      int
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
	log.Printf("[CleanupJob] Starting refresh token cleanup job (interval: %v, retention: %d days)",
		j.interval, j.retentionDays)

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
				log.Printf("[CleanupJob] Stopping refresh token cleanup job")
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
		log.Printf("[CleanupJob] Error cleaning up expired tokens: %v", err)
		return
	}

	if totalDeleted > 0 {
		log.Printf("[CleanupJob] Cleaned up %d expired tokens in %v", totalDeleted, elapsed)
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
		log.Printf("[CleanupJob] Cleaned up %d expired blacklist entries", rowsAffected)
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
		log.Printf("[CleanupJob] Cleaned up %d old rate limit entries", rowsAffected)
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
		log.Printf("[CleanupJob] Cleaned up %d old audit log entries", rowsAffected)
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
		log.Printf("[CleanupJob] Cleaned up %d expired/inactive API keys", rowsAffected)
	}
	return nil
}

// RunAll executes all cleanup jobs
func (j *CleanupJob) RunAll() {
	startTime := time.Now()
	ctx := context.Background()

	log.Printf("[CleanupJob] Starting comprehensive cleanup...")

	// Cleanup refresh tokens
	if _, err := j.cleanupExpiredTokens(ctx); err != nil {
		log.Printf("[CleanupJob] Error cleaning up refresh tokens: %v", err)
	}

	// Cleanup token blacklist
	if err := j.CleanupTokenBlacklist(ctx); err != nil {
		log.Printf("[CleanupJob] Error cleaning up token blacklist: %v", err)
	}

	// Cleanup rate limits (keep for 24 hours)
	if err := j.CleanupRateLimits(ctx, 24); err != nil {
		log.Printf("[CleanupJob] Error cleaning up rate limits: %v", err)
	}

	// Cleanup audit logs (keep for 90 days)
	if err := j.CleanupAuditLogs(ctx, 90); err != nil {
		log.Printf("[CleanupJob] Error cleaning up audit logs: %v", err)
	}

	// Cleanup expired API keys (keep inactive keys for 30 days)
	if err := j.CleanupExpiredAPIKeys(ctx, 30); err != nil {
		log.Printf("[CleanupJob] Error cleaning up API keys: %v", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("[CleanupJob] Comprehensive cleanup completed in %v", elapsed)
}

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

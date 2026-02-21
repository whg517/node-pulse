package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RateLimiter provides database-backed rate limiting
type RateLimiter struct {
	pool *pgxpool.Pool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(pool *pgxpool.Pool) *RateLimiter {
	return &RateLimiter{pool: pool}
}

// ClearRateLimitStore clears all rate limit entries (for testing)
func ClearRateLimitStore(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `DELETE FROM rate_limits`)
	return err
}

// InitRateLimiter initializes the rate limiter (for compatibility)
func InitRateLimiter() {
	// No-op for database-backed rate limiter
}

// WindowType defines the rate limit window type
type WindowType string

const (
	WindowPerMinute WindowType = "minute"
	WindowPerHour   WindowType = "hour"
	WindowPerDay    WindowType = "day"
)

// CheckRateLimit checks if the request is within rate limits
// Returns (allowed, remaining, resetTime, error)
func (r *RateLimiter) CheckRateLimit(ctx context.Context, key string, windowType WindowType, maxCount int) (bool, int, time.Time, error) {
	now := time.Now()
	windowStart := r.getWindowStart(now, windowType)

	// Try to increment or create rate limit entry
	var currentCount int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO rate_limits (key, window_type, window_start, request_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (key, window_type, window_start)
		DO UPDATE SET request_count = rate_limits.request_count + 1
		RETURNING request_count
	`, key, windowType, windowStart).Scan(&currentCount)

	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to check rate limit: %w", err)
	}

	allowed := currentCount <= maxCount
	remaining := max(0, maxCount-currentCount)
	resetTime := r.getNextWindowStart(now, windowType)

	return allowed, remaining, resetTime, nil
}

// getWindowStart calculates the start of the current time window
func (r *RateLimiter) getWindowStart(now time.Time, windowType WindowType) time.Time {
	switch windowType {
	case WindowPerMinute:
		return now.Truncate(time.Minute)
	case WindowPerHour:
		return now.Truncate(time.Hour)
	case WindowPerDay:
		return now.Truncate(24 * time.Hour)
	default:
		return now.Truncate(time.Minute)
	}
}

// getNextWindowStart calculates the start of the next time window
func (r *RateLimiter) getNextWindowStart(now time.Time, windowType WindowType) time.Time {
	switch windowType {
	case WindowPerMinute:
		return now.Truncate(time.Minute).Add(time.Minute)
	case WindowPerHour:
		return now.Truncate(time.Hour).Add(time.Hour)
	case WindowPerDay:
		return now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	default:
		return now.Truncate(time.Minute).Add(time.Minute)
	}
}

// CleanupOldEntries removes old rate limit entries
func (r *RateLimiter) CleanupOldEntries(ctx context.Context, retentionHours int) error {
	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour)

	_, err := r.pool.Exec(ctx, `
		DELETE FROM rate_limits
		WHERE window_start < $1
	`, cutoff)

	if err != nil {
		return fmt.Errorf("failed to cleanup old rate limit entries: %w", err)
	}

	return nil
}

// ResetRateLimit resets the rate limit for a key (for testing/admin)
func (r *RateLimiter) ResetRateLimit(ctx context.Context, key string, windowType WindowType) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM rate_limits
		WHERE key = $1 AND window_type = $2
	`, key, windowType)

	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	return nil
}

// GetCurrentCount gets the current count for a key in the current window
func (r *RateLimiter) GetCurrentCount(ctx context.Context, key string, windowType WindowType) (int, error) {
	now := time.Now()
	windowStart := r.getWindowStart(now, windowType)

	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(request_count, 0)
		FROM rate_limits
		WHERE key = $1 AND window_type = $2 AND window_start = $3
	`, key, windowType, windowStart).Scan(&count)

	if err != nil {
		// No rows means no requests yet, return 0
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current count: %w", err)
	}

	return count, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

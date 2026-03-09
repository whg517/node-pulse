package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProgressiveRateLimiter implements progressive rate limiting with increasing penalties
type ProgressiveRateLimiter struct {
	pool *pgxpool.Pool
}

// NewProgressiveRateLimiter creates a new progressive rate limiter
func NewProgressiveRateLimiter(pool *pgxpool.Pool) *ProgressiveRateLimiter {
	return &ProgressiveRateLimiter{pool: pool}
}

// ViolationLevel represents the severity level of rate limit violations
type ViolationLevel int

const (
	NoViolation      ViolationLevel = 0
	FirstViolation   ViolationLevel = 1
	SecondViolation  ViolationLevel = 2
	ThirdViolation   ViolationLevel = 3
	FourthPlusViolation ViolationLevel = 4
)

// CheckProgressiveRateLimit checks rate limit with progressive penalties
// Returns (allowed, violationLevel, retryAfter, error)
//
// Progressive penalties:
// - First violation: 60 second ban
// - Second violation: 5 minute ban
// - Third violation: 1 hour ban
// - Fourth+ violation: 24 hour ban (requires manual review)
func (r *ProgressiveRateLimiter) CheckProgressiveRateLimit(
	ctx context.Context,
	key string,
	windowType WindowType,
	maxCount int,
) (allowed bool, level ViolationLevel, retryAfter time.Time, err error) {

	now := time.Now()
	windowStart := getWindowStart(now, windowType)

	// Check current count and violation history
	var currentCount int
	var violationCount int
	var lastViolationTime *time.Time

	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(request_count, 0),
		       COALESCE(violation_count, 0),
		       last_violation_at
		FROM rate_limits
		WHERE key = $1 AND window_type = $2 AND window_start = $3
	`, key, windowType, windowStart).Scan(&currentCount, &violationCount, &lastViolationTime)

	if err != nil {
		// No existing record, this is the first request
		currentCount = 0
		violationCount = 0
		lastViolationTime = nil
	}

	// Check if currently in violation period
	if lastViolationTime != nil {
		violationDuration := getViolationDuration(ViolationLevel(violationCount))
		violationEnd := lastViolationTime.Add(violationDuration)

		if now.Before(violationEnd) {
			// Still in violation period
			return false, ViolationLevel(violationCount), violationEnd, nil
		}
	}

	// Check if this request exceeds the limit
	currentCount++
	allowed = currentCount <= maxCount

	if !allowed {
		// New violation - increment violation count
		violationCount++
		if violationCount > 4 {
			violationCount = 4 // Cap at 4 (24 hour ban)
		}

		retryAfter = now.Add(getViolationDuration(ViolationLevel(violationCount)))

		// Update with new violation
		_, err = r.pool.Exec(ctx, `
			INSERT INTO rate_limits (key, window_type, window_start, request_count, violation_count, last_violation_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (key, window_type, window_start)
			DO UPDATE SET
				request_count = $4,
				violation_count = $5,
				last_violation_at = NOW()
		`, key, windowType, windowStart, currentCount, violationCount)

		return false, ViolationLevel(violationCount), retryAfter, err
	}

	// No violation - update request count, reset violation count if we're past the violation period
	resetViolations := lastViolationTime == nil || now.After(lastViolationTime.Add(getViolationDuration(ViolationLevel(violationCount))))

	if resetViolations {
		violationCount = 0
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO rate_limits (key, window_type, window_start, request_count, violation_count)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key, window_type, window_start)
		DO UPDATE SET
			request_count = $4,
			violation_count = $5
	`, key, windowType, windowStart, currentCount, violationCount)

	return true, NoViolation, time.Time{}, err
}

// getViolationDuration returns the ban duration for a given violation level
func getViolationDuration(level ViolationLevel) time.Duration {
	switch level {
	case FirstViolation:
		return 60 * time.Second
	case SecondViolation:
		return 5 * time.Minute
	case ThirdViolation:
		return 1 * time.Hour
	case FourthPlusViolation:
		return 24 * time.Hour
	default:
		return 0
	}
}

// getWindowStart calculates the start of the current time window
func getWindowStart(now time.Time, windowType WindowType) time.Time {
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

// ResetViolations resets violation count for a key (admin function)
func (r *ProgressiveRateLimiter) ResetViolations(ctx context.Context, key string, windowType WindowType) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE rate_limits
		SET violation_count = 0, last_violation_at = NULL
		WHERE key = $1 AND window_type = $2
	`, key, windowType)

	if err != nil {
		return fmt.Errorf("failed to reset violations: %w", err)
	}

	return nil
}

// GetViolationInfo returns current violation status for a key
func (r *ProgressiveRateLimiter) GetViolationInfo(
	ctx context.Context,
	key string,
	windowType WindowType,
) (level ViolationLevel, lastViolation *time.Time, retryAfter *time.Time, err error) {

	var violationCount int
	var lastViolationTime *time.Time

	err = r.pool.QueryRow(ctx, `
		SELECT violation_count, last_violation_at
		FROM rate_limits
		WHERE key = $1 AND window_type = $2
		ORDER BY window_start DESC
		LIMIT 1
	`, key, windowType).Scan(&violationCount, &lastViolationTime)

	if err != nil {
		// No records
		return NoViolation, nil, nil, nil
	}

	level = ViolationLevel(violationCount)

	// Calculate retry-after time if in violation period
	if lastViolationTime != nil && violationCount > 0 {
		violationEnd := lastViolationTime.Add(getViolationDuration(level))
		if time.Now().Before(violationEnd) {
			retryAfter = &violationEnd
		}
	}

	return level, lastViolationTime, retryAfter, nil
}

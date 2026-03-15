package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimiter_CheckRateLimit tests basic rate limiting
func TestRateLimiter_CheckRateLimit(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	key := "test-user-192.168.1.1"
	maxCount := 5

	// First request should be allowed
	allowed, remaining, resetTime, err := limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 4, remaining)
	assert.True(t, resetTime.After(time.Now()))
}

// TestRateLimiter_CheckRateLimit_Exhausted tests rate limit exhaustion
func TestRateLimiter_CheckRateLimit_Exhausted(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	key := "test-user-exhaust"
	maxCount := 3

	// Exhaust the limit
	for i := 0; i < maxCount; i++ {
		allowed, _, _, err := limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Next request should be denied
	allowed, remaining, _, err := limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
}

// TestRateLimiter_WindowReset tests that limits reset after window expires
func TestRateLimiter_WindowReset(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	key := "test-user-reset"
	maxCount := 2

	// Exhaust the limit
	_, _, _, err := limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	require.NoError(t, err)
	_, _, _, err = limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	require.NoError(t, err)

	// Third request should be denied
	allowed, _, _, err := limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Note: Can't easily test actual window reset in unit test without time mocking
	// Integration tests would verify this works correctly with real time passage
}

// TestRateLimiter_DifferentKeys tests that different keys have independent limits
func TestRateLimiter_DifferentKeys(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	maxCount := 2

	// Exhaust limit for key1
	key1 := "user-1"
	_, _, _, _ = limiter.CheckRateLimit(ctx, key1, WindowPerMinute, maxCount)
	_, _, _, _ = limiter.CheckRateLimit(ctx, key1, WindowPerMinute, maxCount)

	// key1 should be exhausted
	allowed, _, _, _ := limiter.CheckRateLimit(ctx, key1, WindowPerMinute, maxCount)
	assert.False(t, allowed)

	// key2 should still have full quota
	key2 := "user-2"
	allowed, remaining, _, _ := limiter.CheckRateLimit(ctx, key2, WindowPerMinute, maxCount)
	assert.True(t, allowed)
	assert.Equal(t, 1, remaining)
}

// TestRateLimiter_DifferentWindowTypes tests independent limits for different window types
func TestRateLimiter_DifferentWindowTypes(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	key := "test-user-windows"

	// Exhaust per-minute limit
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, 2)
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, 2)

	// Per-minute should be exhausted
	allowed, _, _, _ := limiter.CheckRateLimit(ctx, key, WindowPerMinute, 2)
	assert.False(t, allowed)

	// Per-hour should still be available
	allowed, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerHour, 2)
	assert.True(t, allowed)
}

// TestRateLimiter_GetCurrentCount tests getting current count
func TestRateLimiter_GetCurrentCount(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	key := "test-user-count"

	// Initial count should be 0
	count, err := limiter.GetCurrentCount(ctx, key, WindowPerMinute)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Make some requests
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, 10)
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, 10)
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, 10)

	// Count should be 3
	count, err = limiter.GetCurrentCount(ctx, key, WindowPerMinute)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

// TestRateLimiter_ResetRateLimit tests manual reset
func TestRateLimiter_ResetRateLimit(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	key := "test-user-reset"
	maxCount := 2

	// Exhaust limit
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)

	// Should be exhausted
	allowed, _, _, _ := limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	assert.False(t, allowed)

	// Reset
	err := limiter.ResetRateLimit(ctx, key, WindowPerMinute)
	require.NoError(t, err)

	// Should be allowed again
	allowed, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	assert.True(t, allowed)
}

// TestRateLimiter_CleanupOldEntries tests cleanup of old entries
func TestRateLimiter_CleanupOldEntries(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	limiter := NewRateLimiter(pool)
	ctx := context.Background()

	key := "test-user-cleanup"

	// Create some entries
	_, _, _, _ = limiter.CheckRateLimit(ctx, key, WindowPerMinute, 10)

	// Cleanup with 0 retention (delete all)
	err := limiter.CleanupOldEntries(ctx, 0)
	require.NoError(t, err)

	// Count should be 0 after cleanup
	count, err := limiter.GetCurrentCount(ctx, key, WindowPerMinute)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestRateLimiter_PersistenceAcrossRestart tests that limits persist (simulated)
func TestRateLimiter_PersistenceAcrossRestart(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	ctx := context.Background()

	key := "test-user-persist"
	maxCount := 3

	// Simulate "first instance" - create some requests
	limiter1 := NewRateLimiter(pool)
	_, _, _, _ = limiter1.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	_, _, _, _ = limiter1.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)

	// Simulate "server restart" - create new limiter instance
	limiter2 := NewRateLimiter(pool)

	// Count should persist across "restart"
	count, err := limiter2.GetCurrentCount(ctx, key, WindowPerMinute)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Next request should consume the 3rd slot
	allowed, remaining, _, _ := limiter2.CheckRateLimit(ctx, key, WindowPerMinute, maxCount)
	assert.True(t, allowed)
	assert.Equal(t, 0, remaining)
}

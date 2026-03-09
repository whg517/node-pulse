package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	// Use environment variable for test database URL
	connString := "postgres://nodepulse:testpass@localhost:5432/nodepulse_test?sslmode=disable"

	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err, "Failed to connect to test database")

	// Clean up any existing test data
	_, _ = pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id LIKE 'test-%'")
	_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id LIKE 'test-%'")

	t.Cleanup(func() {
		// Clean up test data after test
		_, _ = pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id LIKE 'test-%'")
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id LIKE 'test-%'")
		pool.Close()
	})

	return pool
}

// TestCreateSession_Success tests successful session creation
func TestCreateSession_Success(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0 Test Browser"

	session, refreshToken, refreshTokenRecord, err := sessionService.CreateSession(ctx, userID, ipAddress, userAgent, false)

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotEmpty(t, refreshToken)
	require.NotNil(t, refreshTokenRecord)

	// Verify session fields
	assert.Equal(t, userID, session.UserID.String())
	assert.NotNil(t, session.DeviceID)
	assert.NotNil(t, session.IPAddress)
	assert.Equal(t, ipAddress, *session.IPAddress)
	assert.NotNil(t, session.UserAgent)
	assert.Equal(t, userAgent, *session.UserAgent)
	assert.False(t, session.RememberMe)
	assert.True(t, session.ExpiresAt.Time.After(time.Now()))
	assert.True(t, session.MaxValidUntil.Time.After(time.Now()))

	// Verify refresh token
	assert.Equal(t, session.SessionID, *refreshTokenRecord.SessionID)
}

// TestCreateSession_WithRememberMe tests session creation with remember me
func TestCreateSession_WithRememberMe(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	session, _, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Test Browser", true)

	require.NoError(t, err)
	require.NotNil(t, session)
	assert.True(t, session.RememberMe)
	// With remember me, expiry should be 90 days instead of 30
	expectedExpiry := time.Now().Add(89 * 24 * time.Hour) // Allow 1 day tolerance
	assert.True(t, session.ExpiresAt.Time.After(expectedExpiry))
}

// TestSessionLimitEnforcement tests that max 10 sessions are allowed per user
func TestSessionLimitEnforcement(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create 11 sessions (max is 10)
	var sessions []*models.Session
	for i := 0; i < 11; i++ {
		session, _, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Test Browser", false)
		require.NoError(t, err, "Session %d should succeed", i)
		sessions = append(sessions, session)
	}

	// Verify only 10 active sessions exist
	activeSessions, err := sessionService.GetUserSessions(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, activeSessions, 10, "Should have exactly 10 active sessions")

	// Verify the oldest session was evicted
	sessionIDs := make(map[uuid.UUID]bool)
	for _, s := range activeSessions {
		sessionIDs[s.SessionID] = true
	}

	// The first session should have been evicted
	assert.False(t, sessionIDs[sessions[0].SessionID], "Oldest session should be evicted")
	// All other sessions should still exist
	for i := 1; i < 11; i++ {
		assert.True(t, sessionIDs[sessions[i].SessionID], "Session %d should still exist", i)
	}
}

// TestRefreshTokenRotation_Success tests successful token rotation
func TestRefreshTokenRotation_Success(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create initial session
	_, oldRefreshToken, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Test Browser", false)
	require.NoError(t, err)

	// Rotate the token
	newRefreshToken, newTokenRecord, err := sessionService.RefreshSession(ctx, oldRefreshToken, "192.168.1.1", "Test Browser")

	require.NoError(t, err)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotNil(t, newTokenRecord)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken, "New token should be different")
}

// TestConcurrentUseDetection tests that concurrent token use is detected
func TestConcurrentUseDetection(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create initial session
	_, oldRefreshToken, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Test Browser", false)
	require.NoError(t, err)

	// First refresh should succeed
	_, _, err = sessionService.RefreshSession(ctx, oldRefreshToken, "192.168.1.1", "Test Browser")
	require.NoError(t, err)

	// Second refresh with same token should fail (already used)
	_, _, err = sessionService.RefreshSession(ctx, oldRefreshToken, "192.168.1.1", "Test Browser")
	assert.Error(t, err, "Should reject reused token")
	assert.Contains(t, err.Error(), "already used")
}

// TestGracePeriodHandling tests grace period for concurrent refresh attempts
func TestGracePeriodHandling(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()
	ipAddress := "192.168.1.1"

	// Create initial session
	_, oldRefreshToken, _, err := sessionService.CreateSession(ctx, userID, ipAddress, "Test Browser", false)
	require.NoError(t, err)

	// First refresh should succeed
	newRefreshToken1, _, err := sessionService.RefreshSession(ctx, oldRefreshToken, ipAddress, "Test Browser")
	require.NoError(t, err)

	// Immediate second refresh from same IP within grace period should succeed (race condition handling)
	// This simulates two tabs refreshing at the same time
	newRefreshToken2, _, err := sessionService.RefreshSession(ctx, oldRefreshToken, ipAddress, "Test Browser")
	require.NoError(t, err, "Should allow grace period refresh from same IP")
	assert.NotEmpty(t, newRefreshToken2)

	// The tokens should be different (new token issued)
	assert.NotEqual(t, newRefreshToken1, newRefreshToken2)
}

// TestGracePeriodDifferentIP tests that grace period doesn't apply to different IPs
func TestGracePeriodDifferentIP(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create initial session
	_, oldRefreshToken, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Test Browser", false)
	require.NoError(t, err)

	// First refresh should succeed
	_, _, err = sessionService.RefreshSession(ctx, oldRefreshToken, "192.168.1.1", "Test Browser")
	require.NoError(t, err)

	// Second refresh from different IP should fail even within grace period (potential attack)
	_, _, err = sessionService.RefreshSession(ctx, oldRefreshToken, "192.168.1.100", "Test Browser")
	assert.Error(t, err, "Should reject token reuse from different IP")
	assert.Contains(t, err.Error(), "reuse detected")
}

// TestRevokeSingleSession tests revoking a single session
func TestRevokeSingleSession(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create two sessions
	session1, _, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Browser 1", false)
	require.NoError(t, err)

	session2, _, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Browser 2", false)
	require.NoError(t, err)

	// Revoke first session
	err = sessionService.RevokeSession(ctx, userID, session1.SessionID)
	require.NoError(t, err)

	// Verify only first session is revoked
	sessions, err := sessionService.GetUserSessions(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, session2.SessionID, sessions[0].SessionID)
}

// TestRevokeAllSessions tests revoking all user sessions
func TestRevokeAllSessions(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		_, _, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Test Browser", false)
		require.NoError(t, err)
	}

	// Verify all sessions exist
	sessions, err := sessionService.GetUserSessions(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Revoke all sessions
	err = sessionService.RevokeAllSessions(ctx, userID)
	require.NoError(t, err)

	// Verify no active sessions remain
	sessions, err = sessionService.GetUserSessions(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, sessions, 0)
}

// TestSessionActivityUpdate tests that session activity is updated on refresh
func TestSessionActivityUpdate(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create session
	session, refreshToken, _, err := sessionService.CreateSession(ctx, userID, "192.168.1.1", "Test Browser", false)
	require.NoError(t, err)

	originalActivityTime := session.LastActivityAt.Time

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Refresh token (should update activity)
	_, _, err = sessionService.RefreshSession(ctx, refreshToken, "192.168.1.1", "Test Browser")
	require.NoError(t, err)

	// Verify activity was updated
	updatedSession, err := sessionService.ValidateSession(ctx, session.SessionID)
	require.NoError(t, err)
	assert.True(t, updatedSession.LastActivityAt.Time.After(originalActivityTime))
}

// TestExpiredSessionCleanup tests cleanup of expired sessions
func TestExpiredSessionCleanup(t *testing.T) {
	pool := setupTestDB(t)
	refreshTokenService := NewRefreshTokenService(pool)
	sessionService := NewSessionService(pool, refreshTokenService)

	ctx := context.Background()
	userID := "test-" + uuid.New().String()

	// Create a session with manual expiry (we'll insert it directly with past expiry)
	expiredSessionID := uuid.New()
	pastTime := time.Now().Add(-24 * time.Hour)

	_, err := pool.Exec(ctx, `
		INSERT INTO sessions (session_id, user_id, device_id, ip_address, user_agent, remember_me, expires_at, max_valid_until, last_activity_at, created_at)
		VALUES ($1, $2, 'device123', '192.168.1.1', 'Test', false, $3, $4, $5, $6)
	`, expiredSessionID, userID, pastTime, pastTime, pastTime, pastTime)
	require.NoError(t, err)

	// Run cleanup
	err = sessionService.CleanupExpiredSessions(ctx, 1)
	require.NoError(t, err)

	// Verify expired session was deleted
	_, err = sessionService.ValidateSession(ctx, expiredSessionID)
	assert.Error(t, err, "Expired session should be deleted")
}

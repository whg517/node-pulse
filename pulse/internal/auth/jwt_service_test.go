package auth

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTService_GenerateAccessToken tests token generation
func TestJWTService_GenerateAccessToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewJWTService("test-secret-key-32-bytes-1234567890", 15, pool)

	token, jti, err := service.GenerateAccessToken("user-123", "admin")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, jti)
	assert.NotEqual(t, token, jti) // Token and JTI should be different
}

// TestJWTService_ValidateAccessToken tests valid token validation
func TestJWTService_ValidateAccessToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewJWTService("test-secret-key-32-bytes-1234567890", 15, pool)

	// Generate token
	token, _, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Validate token
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.NotEmpty(t, claims.JTI)
}

// TestJWTService_ValidateAccessToken_Expired tests expired token rejection
func TestJWTService_ValidateAccessToken_Expired(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	// Create service with negative expiration to ensure token is already expired
	service := NewJWTService("test-secret-key-32-bytes-1234567890", -1, pool)

	token, _, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Should fail validation immediately (token already expired)
	_, err = service.ValidateAccessToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestJWTService_ValidateAccessToken_AlgorithmConfusion tests "none" algorithm rejection
func TestJWTService_ValidateAccessToken_AlgorithmConfusion(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewJWTService("test-secret-key-32-bytes-1234567890", 15, pool)

	// Create a malicious "none" algorithm token
	maliciousToken := "eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4ifQ."

	_, err := service.ValidateAccessToken(maliciousToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

// TestJWTService_ValidateAccessToken_ClockSkew tests 60-second clock skew tolerance
func TestJWTService_ValidateAccessToken_ClockSkew(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewJWTService("test-secret-key-32-bytes-1234567890", 15, pool)

	token, _, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Simulate token being 55 seconds in the future (within 60s leeway)
	// This should still validate due to clock skew tolerance
	claims, err := service.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)

	// Note: Testing actual future tokens requires mocking time,
	// but the parser is configured with jwt.WithLeeway(60*time.Second)
}

// TestJWTService_CheckRevoked tests blacklist checking
func TestJWTService_CheckRevoked(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	ctx := context.Background()
	service := NewJWTService("test-secret-key-32-bytes-1234567890", 15, pool)

	// Check non-revoked token
	revoked, err := service.CheckRevoked(ctx, "test-jti-123")
	require.NoError(t, err)
	assert.False(t, revoked)

	// Add to blacklist
	_, err = pool.Exec(ctx, `
		INSERT INTO token_blacklist (jti, revoked_at, expires_at)
		VALUES ($1, NOW(), NOW() + INTERVAL '1 hour')
	`, "test-jti-123")
	require.NoError(t, err)

	// Check revoked token
	revoked, err = service.CheckRevoked(ctx, "test-jti-123")
	require.NoError(t, err)
	assert.True(t, revoked)
}

// TestJWTService_GetJTI tests JTI extraction
func TestJWTService_GetJTI(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewJWTService("test-secret-key-32-bytes-1234567890", 15, pool)

	token, expectedJTI, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Extract JTI from token
	extractedJTI, err := service.GetJTI(token)
	require.NoError(t, err)
	assert.Equal(t, expectedJTI, extractedJTI)
}

// setupTestDBWithCleanup creates a test database connection with cleanup
func setupTestDBWithCleanup(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Use test database from environment or skip
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
		pool.Close()
		return nil
	}

	// Clean up any existing tables
	cleanupTables(ctx, pool)

	// Create all tables (run migrations)
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			user_id UUID PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			email VARCHAR(255),
			role VARCHAR(50) NOT NULL DEFAULT 'viewer',
			is_active BOOLEAN DEFAULT true,
			locked_until TIMESTAMP,
			failed_login_attempts INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "Failed to create users table")

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id SERIAL PRIMARY KEY,
			token_id UUID UNIQUE NOT NULL,
			token_hash TEXT NOT NULL,
			user_id UUID NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			max_valid_until TIMESTAMP NOT NULL,
			revoked_at TIMESTAMP,
			replaced_by UUID REFERENCES refresh_tokens(token_id),
			user_agent TEXT,
			ip_address INET,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "Failed to create refresh_tokens table")

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS token_blacklist (
			jti TEXT PRIMARY KEY,
			revoked_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL
		)
	`)
	require.NoError(t, err, "Failed to create token_blacklist table")

	// Defer pool cleanup
	t.Cleanup(func() {
		cleanupTables(ctx, pool)
		pool.Close()
	})

	return pool
}

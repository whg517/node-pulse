package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCleanupJobTest creates a test database and cleanup job
func setupCleanupJobTest(t *testing.T) (*pgxpool.Pool, *CleanupJob, func()) {
	t.Helper()

	// Use test database from environment or skip
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil, nil, nil
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
		pool.Close()
		return nil, nil, nil
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
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
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

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS rate_limits (
			id SERIAL PRIMARY KEY,
			key VARCHAR(255) NOT NULL,
			window_type VARCHAR(10) NOT NULL,
			window_start TIMESTAMP NOT NULL,
			request_count INTEGER DEFAULT 1,
			UNIQUE(key, window_type, window_start)
		)
	`)
	require.NoError(t, err, "Failed to create rate_limits table")

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS auth_audit_logs (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			user_id UUID,
			ip_address INET,
			details JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "Failed to create auth_audit_logs table")

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			key_hash TEXT UNIQUE NOT NULL,
			key_prefix TEXT NOT NULL,
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			is_active BOOLEAN DEFAULT true,
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW(),
			last_used_at TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create api_keys table")

	// Create cleanup job with short intervals for testing
	job := NewCleanupJob(pool, 3600, 90) // 1 hour interval, 90 days retention

	cleanupFunc := func() {
		cleanupTables(ctx, pool)
		pool.Close()
	}

	return pool, job, cleanupFunc
}

// TestCleanupJob_NewCleanupJob tests cleanup job creation
func TestCleanupJob_NewCleanupJob(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	job := NewCleanupJob(pool, 3600, 90)

	assert.NotNil(t, job, "CleanupJob should be created")
	assert.Equal(t, pool, job.pool, "Pool should be set")
	assert.Equal(t, 3600*time.Second, job.interval, "Interval should be set")
	assert.Equal(t, 90, job.retentionDays, "Retention days should be set")
	assert.NotNil(t, job.stopChan, "Stop channel should be created")
	assert.Equal(t, 1000, job.batchSize, "Batch size should be 1000")
}

// TestCleanupJob_CleanupExpiredTokens tests cleanup of expired refresh tokens
func TestCleanupJob_CleanupExpiredTokens(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()

	// Create the user first (foreign key requirement)
	hashedPassword, _ := HashPassword("testpass")
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err, "Failed to create test user")

	// Insert test data - expired and non-expired tokens
	_, err = pool.Exec(ctx, `
		INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at)
		VALUES
			($1, $2, $3, NOW() - INTERVAL '24 hours', NOW() - INTERVAL '30 days', NOW()),
			($4, $5, $6, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 days', NOW()),
			($7, $8, $9, NOW() + INTERVAL '24 hours', NOW() + INTERVAL '30 days', NOW())
		`,
		uuid.New(), HashTokenSHA256("expired1"), userID,
		uuid.New(), HashTokenSHA256("expired2"), userID,
		uuid.New(), HashTokenSHA256("valid"), userID,
	)
	require.NoError(t, err, "Failed to insert test tokens")

	// Run cleanup
	job := NewCleanupJob(pool, 3600, 90)
	deleted, err := job.cleanupExpiredTokens(ctx)
	assert.NoError(t, err, "Cleanup should not error")
	assert.Equal(t, 2, deleted, "Should delete 2 expired tokens")

	// Verify only valid token remains
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", userID).Scan(&count)
	assert.NoError(t, err, "Failed to count remaining tokens")
	assert.Equal(t, 1, count, "Only 1 valid token should remain")
}

// TestCleanupJob_CleanupTokenBlacklist tests cleanup of expired blacklist entries
func TestCleanupJob_CleanupTokenBlacklist(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	ctx := context.Background()

	// Insert test blacklist entries
	_, err := pool.Exec(ctx, `
		INSERT INTO token_blacklist (jti, revoked_at, expires_at)
		VALUES
			('jti-expired-1', NOW() - INTERVAL '24 hours', NOW() - INTERVAL '1 hour'),  -- Expired
			('jti-expired-2', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '30 minutes'),  -- Expired
			('jti-valid', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '1 hour')  -- Still valid
	`)
	require.NoError(t, err, "Failed to insert test blacklist entries")

	// Run cleanup
	job := NewCleanupJob(pool, 3600, 90)
	err = job.CleanupTokenBlacklist(ctx)
	assert.NoError(t, err, "Cleanup should not error")

	// Verify only valid entry remains
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM token_blacklist WHERE expires_at > NOW()").Scan(&count)
	assert.NoError(t, err, "Failed to count remaining entries")
	assert.Equal(t, 1, count, "Only 1 valid blacklist entry should remain")
}

// TestCleanupJob_CleanupRateLimits tests cleanup of old rate limit entries
func TestCleanupJob_CleanupRateLimits(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	ctx := context.Background()

	// Insert test rate limit entries
	_, err := pool.Exec(ctx, `
		INSERT INTO rate_limits (key, window_type, window_start, request_count)
		VALUES
			('ip:old1', 'MINUTE', NOW() - INTERVAL '48 hours', 5),  -- Old
			('ip:old2', 'MINUTE', NOW() - INTERVAL '25 hours', 3),  -- Old
			('ip:recent', 'MINUTE', NOW() - INTERVAL '1 hour', 2)  -- Recent (within 24h)
	`)
	require.NoError(t, err, "Failed to insert test rate limit entries")

	// Run cleanup with 24 hour retention
	job := NewCleanupJob(pool, 3600, 90)
	err = job.CleanupRateLimits(ctx, 24)
	assert.NoError(t, err, "Cleanup should not error")

	// Verify only recent entry remains
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM rate_limits").Scan(&count)
	assert.NoError(t, err, "Failed to count remaining entries")
	assert.Equal(t, 1, count, "Only 1 recent rate limit entry should remain")
}

// TestCleanupJob_CleanupAuditLogs tests cleanup of old audit log entries
func TestCleanupJob_CleanupAuditLogs(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	ctx := context.Background()

	// Insert test audit log entries
	_, err := pool.Exec(ctx, `
		INSERT INTO auth_audit_logs (event_type, user_id, ip_address, details, created_at)
		VALUES
			('login', $1, '127.0.0.1', '{}', NOW() - INTERVAL '100 days'),  -- Old
			('logout', $1, '127.0.0.1', '{}', NOW() - INTERVAL '95 days'),  -- Old
			('login', $1, '127.0.0.1', '{}', NOW() - INTERVAL '10 days')  -- Recent (within 90 days)
	`, uuid.New())
	require.NoError(t, err, "Failed to insert test audit log entries")

	// Run cleanup with 90 day retention
	job := NewCleanupJob(pool, 3600, 90)
	err = job.CleanupAuditLogs(ctx, 90)
	assert.NoError(t, err, "Cleanup should not error")

	// Verify only recent entry remains
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM auth_audit_logs").Scan(&count)
	assert.NoError(t, err, "Failed to count remaining entries")
	assert.Equal(t, 1, count, "Only 1 recent audit log entry should remain")
}

// TestCleanupJob_DeleteAllTokensForUser tests deletion of all user tokens
func TestCleanupJob_DeleteAllTokensForUser(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()

	// Create the user first (foreign key requirement)
	hashedPassword, _ := HashPassword("testpass")
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err, "Failed to create test user")

	// Insert multiple tokens for user
	_, err = pool.Exec(ctx, `
		INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at)
		VALUES
			($1, $2, $3, NOW() + INTERVAL '1 day', NOW() + INTERVAL '30 days', NOW()),
			($4, $5, $6, NOW() + INTERVAL '2 days', NOW() + INTERVAL '30 days', NOW()),
			($7, $8, $9, NOW() + INTERVAL '3 days', NOW() + INTERVAL '30 days', NOW())
	`,
		uuid.New(), HashTokenSHA256("token1"), userID,
		uuid.New(), HashTokenSHA256("token2"), userID,
		uuid.New(), HashTokenSHA256("token3"), userID,
	)
	require.NoError(t, err, "Failed to insert test tokens")

	// Delete all tokens for user
	job := NewCleanupJob(pool, 3600, 90)
	err = job.DeleteAllTokensForUser(ctx, userID.String())
	assert.NoError(t, err, "DeleteAllTokensForUser should not error")

	// Verify all tokens deleted
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", userID).Scan(&count)
	assert.NoError(t, err, "Failed to count remaining tokens")
	assert.Equal(t, 0, count, "All tokens should be deleted")
}

// TestCleanupJob_StartStop tests starting and stopping cleanup job
func TestCleanupJob_StartStop(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	job := NewCleanupJob(pool, 1, 90) // 1 second interval for testing

	// Start job
	job.Start()

	// Wait a bit to ensure goroutine started
	time.Sleep(100 * time.Millisecond)

	// Stop job (should not panic)
	assert.NotPanics(t, func() {
		job.Stop()
	}, "Stop should not panic")

	// Wait for goroutine to finish
	time.Sleep(200 * time.Millisecond)
}

// TestCleanupJob_RunAll tests running all cleanup jobs
func TestCleanupJob_RunAll(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	job := NewCleanupJob(pool, 3600, 90)

	// Run all cleanup jobs (should not panic)
	assert.NotPanics(t, func() {
		job.RunAll()
	}, "RunAll should not panic")
}

// TestCleanupJob_BatchDeletion tests that large deletions are batched
func TestCleanupJob_BatchDeletion(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()

	// Create the user first (foreign key requirement)
	hashedPassword, _ := HashPassword("testpass")
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err, "Failed to create test user")

	// Insert 2500 expired tokens (more than batch size of 1000)
	// Use PostgreSQL time arithmetic to avoid timezone issues
	for i := 0; i < 2500; i++ {
		_, err := pool.Exec(ctx, `
			INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at)
			VALUES ($1, $2, $3, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 days', NOW())
		`, uuid.New(), HashTokenSHA256(fmt.Sprintf("token%d", i)), userID)
		require.NoError(t, err, "Failed to insert test token")
	}

	// Run cleanup
	job := NewCleanupJob(pool, 3600, 90)
	deleted, err := job.cleanupExpiredTokens(ctx)
	assert.NoError(t, err, "Cleanup should not error")
	assert.Equal(t, 2500, deleted, "Should delete all 2500 expired tokens")

	// Verify all tokens deleted
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", userID).Scan(&count)
	assert.NoError(t, err, "Failed to count remaining tokens")
	assert.Equal(t, 0, count, "All expired tokens should be deleted")
}

// TestCleanupJob_CleanupExpiredAPIKeys tests cleanup of expired and inactive API keys
func TestCleanupJob_CleanupExpiredAPIKeys(t *testing.T) {
	pool, _, cleanup := setupCleanupJobTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New()

	// Create the user first
	hashedPassword, _ := HashPassword("testpass")
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err, "Failed to create test user")

	// Insert test API keys using PostgreSQL INTERVAL for consistent timezone handling
	_, err = pool.Exec(ctx, `
		INSERT INTO api_keys (key_hash, key_prefix, user_id, name, is_active, expires_at, created_at)
		VALUES
			($1, $2, $3, $4, false, NOW() + INTERVAL '30 days', NOW() - INTERVAL '35 days'),
			($5, $6, $7, $8, false, NOW() + INTERVAL '30 days', NOW() - INTERVAL '40 days'),
			($9, $10, $11, $12, true, NOW() - INTERVAL '1 hour', NOW()),
			($13, $14, $15, $16, true, NOW() + INTERVAL '30 days', NOW()),
			($17, $18, $19, $20, false, NOW() + INTERVAL '30 days', NOW() - INTERVAL '10 days')
		`,
		HashTokenSHA256("inactive-old-1"), "old1", userID, "Old Inactive Key 1",
		HashTokenSHA256("inactive-old-2"), "old2", userID, "Old Inactive Key 2",
		HashTokenSHA256("expired"), "exp", userID, "Expired Key",
		HashTokenSHA256("valid"), "val", userID, "Valid Active Key",
		HashTokenSHA256("inactive-recent"), "rec", userID, "Recent Inactive Key")
	require.NoError(t, err, "Failed to insert test API keys")

	// Run cleanup with 30 day retention
	job := NewCleanupJob(pool, 3600, 90)
	err = job.CleanupExpiredAPIKeys(ctx, 30)
	assert.NoError(t, err, "Cleanup should not error")

	// Verify only valid and recent inactive keys remain (3 keys should be deleted)
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM api_keys").Scan(&count)
	assert.NoError(t, err, "Failed to count remaining API keys")
	assert.Equal(t, 2, count, "Only 2 API keys should remain (valid active and recent inactive)")
}

// BenchmarkCleanupJob_CleanupExpiredTokens benchmarks cleanup performance
func BenchmarkCleanupJob_CleanupExpiredTokens(b *testing.B) {
	// This benchmark requires a real database connection
	// Skip if database not available
	pool, err := pgxpool.New(context.Background(), "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable")
	if err != nil {
		b.Skip("Database not available")
		return
	}
	defer pool.Close()

	// Setup test data
	ctx := context.Background()
	userID := uuid.New()

	// Create the user first
	hashedPassword, _ := HashPassword("testpass")
	_, _ = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")

	for i := 0; i < 10000; i++ {
		_, _ = pool.Exec(ctx, `
			INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at)
			VALUES ($1, $2, $3, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 days', NOW())
		`, uuid.New(), HashTokenSHA256(fmt.Sprintf("bench-token%d", i)), userID)
	}

	job := NewCleanupJob(pool, 3600, 90)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = job.cleanupExpiredTokens(ctx)
	}
}

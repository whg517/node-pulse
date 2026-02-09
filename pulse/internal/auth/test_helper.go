package auth

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// cleanupTables drops all test tables
func cleanupTables(ctx context.Context, pool *pgxpool.Pool) {
	pool.Exec(ctx, "DROP TABLE IF EXISTS api_keys CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS auth_audit_logs CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS rate_limits CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS token_blacklist CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS refresh_tokens CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
}

// createTestTables creates all required test database tables
func createTestTables(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	// Create tables in correct order to satisfy foreign key constraints
	tables := []string{
		// Users table must be created first
		`CREATE TABLE IF NOT EXISTS users (
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
		)`,
		// Other tables can reference users
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
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
		)`,
		`CREATE TABLE IF NOT EXISTS token_blacklist (
			jti TEXT PRIMARY KEY,
			revoked_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS rate_limits (
			id SERIAL PRIMARY KEY,
			key VARCHAR(255) NOT NULL,
			window_type VARCHAR(10) NOT NULL,
			window_start TIMESTAMP NOT NULL,
			request_count INTEGER DEFAULT 1,
			UNIQUE(key, window_type, window_start)
		)`,
		`CREATE TABLE IF NOT EXISTS auth_audit_logs (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			user_id UUID,
			ip_address INET,
			details JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			key_hash TEXT UNIQUE NOT NULL,
			key_prefix TEXT NOT NULL,
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			is_active BOOLEAN DEFAULT true,
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW(),
			last_used_at TIMESTAMP
		)`,
	}

	for _, tableSQL := range tables {
		_, err := pool.Exec(ctx, tableSQL)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}
	}
}

// setupTestDBWithCleanup creates a test database connection with cleanup
// This function is exported for use across all auth test files
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

	// Create all tables
	createTestTables(ctx, t, pool)

	// Register cleanup function
	t.Cleanup(func() {
		cleanupTables(ctx, pool)
		pool.Close()
	})

	return pool
}

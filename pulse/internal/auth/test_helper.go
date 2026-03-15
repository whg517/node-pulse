package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// cleanupTables drops all test tables
func cleanupTables(ctx context.Context, pool *pgxpool.Pool) {
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS api_keys CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS auth_audit_logs CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS rate_limits CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS token_blacklist CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS refresh_tokens CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
}

// createTestTables creates all required test database tables
func createTestTables(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// Verify DB connectivity; skip the test when the database is not reachable
	// (pgxpool.New is lazy and succeeds even without an actual connection)
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
		return
	}

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
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			device_id VARCHAR(255),
			ip_address INET,
			user_agent TEXT,
			remember_me BOOLEAN DEFAULT false,
			expires_at TIMESTAMP NOT NULL,
			max_valid_until TIMESTAMP NOT NULL,
			last_activity_at TIMESTAMP DEFAULT NOW(),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id SERIAL PRIMARY KEY,
			token_id UUID UNIQUE NOT NULL,
			token_hash TEXT NOT NULL,
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			session_id UUID REFERENCES sessions(session_id) ON DELETE CASCADE,
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

// CreateTestUser creates a test user and returns the UUID string
func CreateTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()

	userID := uuid.New()
	hashedPassword, _ := HashPassword("testpass")

	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")

	require.NoError(t, err, "Failed to create test user")
	return userID.String()
}

// GenerateTestRSAKeyPair generates an RSA-2048 key pair for testing
// Returns private key and public key in PEM format
func GenerateTestRSAKeyPair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate RSA private key")

	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Failed to marshal public key")
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM)
}

// NewTestJWTService creates a JWT service with test RSA keys
func NewTestJWTService(t *testing.T, accessExpirationMinutes int, pool *pgxpool.Pool) *JWTService {
	t.Helper()

	privateKeyPEM, publicKeyPEM := GenerateTestRSAKeyPair(t)
	return NewJWTService(privateKeyPEM, publicKeyPEM, "test-key-id", accessExpirationMinutes, pool)
}

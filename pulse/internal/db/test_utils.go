package db

import (
	"context"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

// SetupTestDB creates a test database connection pool
// This function is exported for use in other test packages
func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	// Setup test config
	testutil.SetupTestConfig()

	// Load config before migrations (seedAdminUser needs config)
	config.MustLoad()

	ctx := context.Background()

	// Use test database from environment or default
	testDSN := testutil.GetTestDBURL()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	// Verify actual connectivity (pgxpool.New is lazy)
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping test: cannot ping test database: %v", err)
	}

	// Reset the schema to a clean state before each test run. Because
	// golang-migrate records applied versions in schema_migrations, dropping
	// tables manually would leave the version stale and Migrate() would skip
	// re-creation. Rolling down to version 0 then up guarantees a fresh schema
	// regardless of prior test state.
	if err := resetTestSchema(ctx, pool); err != nil {
		t.Skipf("Skipping: database not available - schema reset failed: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		pool.Close()
		testutil.TeardownTestConfig()
	}

	return pool, cleanup
}

// resetTestSchema tears down any existing schema (via the migration down path
// when possible) and rebuilds it via Up, yielding a deterministic fresh schema
// for each test. It tolerates both fresh databases (nil version) and databases
// left at version 1 by a prior run.
func resetTestSchema(ctx context.Context, pool *pgxpool.Pool) error {
	dsn := pool.Config().ConnString()
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// Roll back any applied migrations to reach a clean (nil) version state.
	// Down returns an error once there is nothing left to roll back, which we
	// ignore. Force(-1) then guarantees the version row is removed even if the
	// database was left dirty by a prior, crashed run.
	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return err
	}
	_ = m.Down()
	if err := m.Force(-1); err != nil { // -1 = remove version row entirely (clean slate)
		return err
	}
	if err := func() error { _, _ = m.Close(); return nil }(); err != nil {
		return err
	}

	// Re-apply from a clean state.
	m2, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return err
	}
	defer func() { _, _ = m2.Close() }()
	if err := m2.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	// Run data seeding (admin user) that migrations cannot express in SQL.
	return Seed(ctx, pool)
}

// CreateTestUser creates a test user in the database and returns the user ID
func CreateTestUser(ctx context.Context, pool *pgxpool.Pool, username, password, role string) uuid.UUID {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	userID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
	`, userID, username, hashedPassword, username+"@test.com", role)
	if err != nil {
		panic(err)
	}

	return userID
}

// CreateTestAPIKey creates a test API key and returns the plain text key and key ID
func CreateTestAPIKey(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, name string) (string, int) {
	// Generate a random API key
	keyPlain := uuid.New().String()

	// Hash the key
	hash := bcryptHash(keyPlain)

	// Extract key prefix (first 8 chars)
	keyPrefix := keyPlain[:8]

	// Insert into database
	var keyID int
	err := pool.QueryRow(ctx, `
		INSERT INTO api_keys (key_hash, key_prefix, user_id, name, is_active, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		RETURNING id
	`, hash, keyPrefix, userID, name).Scan(&keyID)
	if err != nil {
		panic(err)
	}

	return keyPlain, keyID
}

// bcryptHash is a helper to hash passwords/API keys
func bcryptHash(plain string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

// CleanupTestTables drops all test tables (useful for cleanup between tests)
func CleanupTestTables(ctx context.Context, pool *pgxpool.Pool) {
	// Drop auth-related tables first (due to foreign keys)
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS rate_limits CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS auth_audit_logs CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS token_blacklist CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS refresh_tokens CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS api_keys CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")

	// Drop other tables
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alert_records CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alert_events CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alert_suppressions CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alerts CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS webhook_logs CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS webhooks CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
}

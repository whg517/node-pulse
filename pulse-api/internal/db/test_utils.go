package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
)

// setupTestDB creates a test database connection pool
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
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

	// Clean up any existing tables from previous tests
	pool.Exec(ctx, "DROP TABLE IF EXISTS alerts CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")

	// Create all tables
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		pool.Close()
		testutil.TeardownTestConfig()
	}

	return pool, cleanup
}

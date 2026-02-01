package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// setupTestDB creates a test database connection pool
func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	// Use test database from environment or default
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"

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

	return pool
}

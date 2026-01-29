package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestCreateProbesTable tests probes table creation
func TestCreateProbesTable(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	defer pool.Close()

	// Run migration
	if err := createProbesTable(ctx, pool); err != nil {
		t.Fatalf("Failed to create probes table: %v", err)
	}

	// Verify table exists
	var tableName string
	err := pool.QueryRow(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = 'probes'
	`).Scan(&tableName)

	if err != nil {
		t.Fatalf("Probes table was not created: %v", err)
	}

	// Verify columns exist
	requiredColumns := []string{
		"id", "node_id", "type", "target", "port",
		"interval_seconds", "count", "timeout_seconds",
		"created_at", "updated_at",
	}

	for _, col := range requiredColumns {
		var columnName string
		err := pool.QueryRow(ctx, `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_name = 'probes' AND column_name = $1
		`, col).Scan(&columnName)

		if err != nil {
			t.Errorf("Required column '%s' was not created: %v", col, err)
		}
	}

	// Verify indexes exist
	requiredIndexes := []string{
		"idx_probes_node_id",
		"idx_probes_type",
	}

	for _, idx := range requiredIndexes {
		var indexName string
		err := pool.QueryRow(ctx, `
			SELECT indexname
			FROM pg_indexes
			WHERE indexname = $1
		`, idx).Scan(&indexName)

		if err != nil {
			t.Errorf("Required index '%s' was not created: %v", idx, err)
		}
	}

	// Verify foreign key constraint to nodes table
	var constraintName string
	err = pool.QueryRow(ctx, `
		SELECT
			tc.constraint_name
		FROM
			information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
		WHERE
			tc.table_name = 'probes'
			AND tc.constraint_type = 'FOREIGN KEY'
			AND kcu.column_name = 'node_id'
	`).Scan(&constraintName)

	if err != nil {
		t.Errorf("Foreign key constraint on node_id was not created: %v", err)
	}
}

// TestCreateMetricsTable tests metrics table creation
func TestCreateMetricsTable(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	defer pool.Close()

	// Create probes table first (required for foreign key)
	if err := createProbesTable(ctx, pool); err != nil {
		t.Fatalf("Failed to create probes table: %v", err)
	}

	// Run migration
	if err := createMetricsTable(ctx, pool); err != nil {
		t.Fatalf("Failed to create metrics table: %v", err)
	}

	// Verify table exists
	var tableName string
	err := pool.QueryRow(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = 'metrics'
	`).Scan(&tableName)

	if err != nil {
		t.Fatalf("Metrics table was not created: %v", err)
	}

	// Verify columns exist
	requiredColumns := []string{
		"id", "node_id", "probe_id", "timestamp",
		"latency_ms", "packet_loss_rate", "jitter_ms",
		"is_aggregated", "created_at",
	}

	for _, col := range requiredColumns {
		var columnName string
		err := pool.QueryRow(ctx, `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_name = 'metrics' AND column_name = $1
		`, col).Scan(&columnName)

		if err != nil {
			t.Errorf("Required column '%s' was not created: %v", col, err)
		}
	}

	// Verify indexes exist
	requiredIndexes := []string{
		"idx_metrics_node_timestamp",
		"idx_metrics_probe_timestamp",
		"idx_metrics_timestamp",
		"idx_metrics_aggregated",
	}

	for _, idx := range requiredIndexes {
		var indexName string
		err := pool.QueryRow(ctx, `
			SELECT indexname
			FROM pg_indexes
			WHERE indexname = $1
		`, idx).Scan(&indexName)

		if err != nil {
			t.Errorf("Required index '%s' was not created: %v", idx, err)
		}
	}

	// Verify foreign key constraints
	constraints := map[string]string{
		"node_id": "nodes",
		"probe_id": "probes",
	}

	for col, refTable := range constraints {
		var constraintName string
		err := pool.QueryRow(ctx, `
			SELECT
				tc.constraint_name
			FROM
				information_schema.table_constraints AS tc
				JOIN information_schema.key_column_usage AS kcu
					ON tc.constraint_name = kcu.constraint_name
				JOIN information_schema.constraint_column_usage AS ccu
					ON ccu.constraint_name = tc.constraint_name
			WHERE
				tc.table_name = 'metrics'
				AND tc.constraint_type = 'FOREIGN KEY'
				AND kcu.column_name = $1
				AND ccu.table_name = $2
		`, col, refTable).Scan(&constraintName)

		if err != nil {
			t.Errorf("Foreign key constraint on %s referencing %s was not created: %v", col, refTable, err)
		}
	}
}

// setupTestDB creates a test database connection pool
func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	// Use test database from environment or default
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	// Clean up any existing probes/metrics tables from previous tests
	pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")

	// Create dependent tables first
	if err := createUsersTable(ctx, pool); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	if err := createNodesTable(ctx, pool); err != nil {
		t.Fatalf("Failed to create nodes table: %v", err)
	}

	return pool
}

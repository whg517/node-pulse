package db

import (
	"context"
	"testing"
)

// TestCreateProbesTable tests the probes table created by the baseline migration.
func TestCreateProbesTable(t *testing.T) {
	ctx := context.Background()
	pool, _ := SetupTestDB(t)
	defer pool.Close()

	// SetupTestDB applies the baseline migration, which creates the probes
	// table; verify its structure rather than re-running the step.
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

// TestCreateMetricsTable tests the metrics table created by the baseline migration.
func TestCreateMetricsTable(t *testing.T) {
	ctx := context.Background()
	pool, _ := SetupTestDB(t)
	defer pool.Close()

	// SetupTestDB applies the baseline migration, which creates the metrics
	// table (probes is created in the same migration); verify its structure.
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

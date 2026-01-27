package db

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNodesTableCreation tests that nodes table is created with correct structure
func TestNodesTableCreation(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Run migration
	err = Migrate(ctx, pool)
	require.NoError(t, err, "Migration should succeed")

	// Verify nodes table exists and has correct columns
	var tableName string
	err = pool.QueryRow(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'nodes'
	`).Scan(&tableName)
	assert.NoError(t, err, "nodes table should exist")
	assert.Equal(t, "nodes", tableName, "Table name should be nodes")

	// Verify columns
	expectedColumns := map[string]bool{
		"id":         false,
		"name":       false,
		"ip":         false,
		"region":     false,
		"tags":       false,
		"created_at": false,
		"updated_at": false,
	}

	rows, err := pool.Query(ctx, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'nodes' AND table_schema = 'public'
		ORDER BY column_name
	`)
	require.NoError(t, err, "Should query columns successfully")

	defer rows.Close()
	var foundColumns []string
	for rows.Next() {
		var colName string
		err := rows.Scan(&colName)
		require.NoError(t, err)
		foundColumns = append(foundColumns, colName)
		if _, exists := expectedColumns[colName]; exists {
			expectedColumns[colName] = true
		}
	}

	require.NoError(t, rows.Err())
	assert.ElementsMatch(t, []string{"created_at", "id", "ip", "name", "region", "tags", "updated_at"}, foundColumns)

	// Verify all expected columns exist
	for col, found := range expectedColumns {
		assert.True(t, found, "Column %s should exist", col)
	}
}

// TestNodesIndexCreation tests that idx_nodes_region index is created
func TestNodesIndexCreation(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Run migration
	err = Migrate(ctx, pool)
	require.NoError(t, err)

	// Verify index exists
	var indexName string
	err = pool.QueryRow(ctx, `
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = 'nodes' AND schemaname = 'public' AND indexname = 'idx_nodes_region'
	`).Scan(&indexName)

	assert.NoError(t, err, "idx_nodes_region index should exist")
	assert.Equal(t, "idx_nodes_region", indexName, "Index name should be idx_nodes_region")
}

// TestNodesTableConstraints tests that nodes table has correct constraints
func TestNodesTableConstraints(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Run migration
	err = Migrate(ctx, pool)
	require.NoError(t, err)

	// Test NOT NULL constraint on name
	_, err = pool.Exec(ctx, `INSERT INTO nodes (name, ip, region) VALUES (NULL, '192.168.1.1', 'us-east')`)
	assert.Error(t, err, "Should reject NULL name")
	assert.Contains(t, err.Error(), "null value in column \"name\"", "Error should mention null constraint")

	// Test NOT NULL constraint on ip
	_, err = pool.Exec(ctx, `INSERT INTO nodes (name, ip, region) VALUES ('Test Node', NULL, 'us-east')`)
	assert.Error(t, err, "Should reject NULL ip")
	assert.Contains(t, err.Error(), "null value in column \"ip\"", "Error should mention null constraint")

	// Test NOT NULL constraint on region
	_, err = pool.Exec(ctx, `INSERT INTO nodes (name, ip, region) VALUES ('Test Node', '192.168.1.1', NULL)`)
	assert.Error(t, err, "Should reject NULL region")
	assert.Contains(t, err.Error(), "null value in column \"region\"", "Error should mention null constraint")
}

// TestNodesTableUUIDGeneration tests that id is auto-generated as UUID
func TestNodesTableUUIDGeneration(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Run migration
	err = Migrate(ctx, pool)
	require.NoError(t, err)

	// Insert node without specifying id (should auto-generate)
	var insertedID string
	err = pool.QueryRow(ctx, `
		INSERT INTO nodes (name, ip, region)
		VALUES ('Test Node', '192.168.1.1', 'us-east')
		RETURNING id
	`).Scan(&insertedID)
	require.NoError(t, err)

	// Verify it's a valid UUID format (should be 36 characters with dashes)
	assert.Len(t, insertedID, 36, "ID should be UUID format with dashes")
	assert.Contains(t, insertedID, "-", "ID should contain UUID dashes")
}

// TestNodesTableTimestampDefaults tests that created_at and updated_at have default values
func TestNodesTableTimestampDefaults(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Run migration
	err = Migrate(ctx, pool)
	require.NoError(t, err)

	// Insert node without specifying timestamps
	var createdAt, updatedAt string
	err = pool.QueryRow(ctx, `
		INSERT INTO nodes (name, ip, region)
		VALUES ('Test Node', '192.168.1.1', 'us-east')
		RETURNING created_at::text, updated_at::text
	`).Scan(&createdAt, &updatedAt)
	require.NoError(t, err)

	// Verify timestamps are not empty
	assert.NotEmpty(t, createdAt, "created_at should have default value")
	assert.NotEmpty(t, updatedAt, "updated_at should have default value")
}

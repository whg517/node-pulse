package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/cleanup"
	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
)

// TestCleanupTask_Integration tests the cleanup task with real database
func TestCleanupTask_Integration(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), testutil.GetTestDBURL())
	if err != nil {
		t.Skip("No database connection")
		return
	}
	defer pool.Close()

	// Run migrations
	ctx := context.Background()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Database not ready: %v", err)
		return
	}

	// Create test node and probe
	testNodeID := uuid.New()
	testProbeID := uuid.New()

	// Clean up test data first
	pool.Exec(ctx, "DELETE FROM metrics WHERE probe_id = $1", testProbeID)
	pool.Exec(ctx, "DELETE FROM probes WHERE id = $1", testProbeID)
	pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", testNodeID)

	// Insert test node
	_, err = pool.Exec(ctx, `
		INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, testNodeID, "cleanup-test-node", "192.168.1.100", "us-east", "{}")
	require.NoError(t, err)

	// Insert test probe
	_, err = pool.Exec(ctx, `
		INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`, testProbeID, testNodeID, "TCP", "example.com", 80, 60, 5, 5)
	require.NoError(t, err)

	// Insert old metrics (8 days ago - should be deleted)
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	for i := 0; i < 5; i++ {
		_, err := pool.Exec(ctx, `
			INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, testNodeID, testProbeID, oldTime, 50.0, 0.01, 2.5)
		require.NoError(t, err)
	}

	// Insert recent metrics (1 day ago - should be kept)
	recentTime := time.Now().Add(-1 * 24 * time.Hour)
	for i := 0; i < 3; i++ {
		_, err := pool.Exec(ctx, `
			INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, testNodeID, testProbeID, recentTime, 45.0, 0.005, 1.8)
		require.NoError(t, err)
	}

	// Verify initial data count
	var initialCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM metrics WHERE probe_id = $1", testProbeID).Scan(&initialCount)
	require.NoError(t, err)
	assert.Equal(t, 8, initialCount, "Should have 8 total metrics (5 old + 3 recent)")

	// Create cleanup task with 7 days retention
	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
		SlowThresholdMs: 30000,
	}

	task, err := cleanup.NewCleanupTask(cfg, pool, nil)
	require.NoError(t, err)
	require.NotNil(t, task)

	// Execute cleanup
	err = task.Execute(ctx)
	assert.NoError(t, err, "Cleanup should succeed")

	// Verify old data was deleted and recent data kept
	var finalCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM metrics WHERE probe_id = $1", testProbeID).Scan(&finalCount)
	require.NoError(t, err)
	assert.Equal(t, 3, finalCount, "Should have 3 metrics remaining (recent only)")

	// Verify task status
	status := task.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "metrics-cleanup", status.Name)
	assert.Equal(t, int64(1), status.RunCount)
	assert.False(t, status.IsRunning)
	assert.Empty(t, status.LastError)

	// Cleanup test data
	pool.Exec(ctx, "DELETE FROM metrics WHERE probe_id = $1", testProbeID)
	pool.Exec(ctx, "DELETE FROM probes WHERE id = $1", testProbeID)
	pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", testNodeID)
}

// TestCleanupTask_ZeroRows_Integration tests cleanup when no data needs to be deleted
func TestCleanupTask_ZeroRows_Integration(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), testutil.GetTestDBURL())
	if err != nil {
		t.Skip("No database connection")
		return
	}
	defer pool.Close()

	ctx := context.Background()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Database not ready: %v", err)
		return
	}

	// Create cleanup task
	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
	}

	task, err := cleanup.NewCleanupTask(cfg, pool, nil)
	require.NoError(t, err)

	// Execute cleanup - should succeed even with no data to delete
	err = task.Execute(ctx)
	assert.NoError(t, err, "Cleanup should succeed even with no data")

	// Verify status
	status := task.GetStatus()
	assert.Equal(t, int64(1), status.RunCount)
	assert.Empty(t, status.LastError)
}

// TestCleanupTask_CustomRetention_Integration tests cleanup with custom retention period
func TestCleanupTask_CustomRetention_Integration(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), testutil.GetTestDBURL())
	if err != nil {
		t.Skip("No database connection")
		return
	}
	defer pool.Close()

	ctx := context.Background()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Database not ready: %v", err)
		return
	}

	// Create test node and probe
	testNodeID := uuid.New()
	testProbeID := uuid.New()

	// Clean up
	defer pool.Exec(ctx, "DELETE FROM metrics WHERE probe_id = $1", testProbeID)
	defer pool.Exec(ctx, "DELETE FROM probes WHERE id = $1", testProbeID)
	defer pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", testNodeID)

	// Insert test data
	pool.Exec(ctx, `
		INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, testNodeID, "cleanup-test-node-2", "192.168.1.101", "us-east", "{}")

	pool.Exec(ctx, `
		INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`, testProbeID, testNodeID, "TCP", "example.com", 80, 60, 5, 5)

	// Insert data 15 days old (should be deleted with 7-day retention)
	oldTime := time.Now().Add(-15 * 24 * time.Hour)
	pool.Exec(ctx, `
		INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, testNodeID, testProbeID, oldTime, 100.0, 0.02, 5.0)

	// Insert data 5 days old (should be kept with 7-day retention)
	recentTime := time.Now().Add(-5 * 24 * time.Hour)
	pool.Exec(ctx, `
		INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, testNodeID, testProbeID, recentTime, 200.0, 0.01, 3.0)

	// Test with 7-day retention (default)
	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
	}

	task, err := cleanup.NewCleanupTask(cfg, pool, nil)
	require.NoError(t, err)

	err = task.Execute(ctx)
	assert.NoError(t, err)

	var count int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM metrics WHERE probe_id = $1", testProbeID).Scan(&count)
	assert.Equal(t, 1, count, "Should keep 5-day-old data, delete 15-day-old data")
}

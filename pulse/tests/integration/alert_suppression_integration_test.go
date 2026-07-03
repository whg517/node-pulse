package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/alert"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/suppression"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

func setupSuppressionTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	testutil.SetupTestConfig()

	// Load configuration
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Connect to test database (skips cleanly when DB is unreachable or -short)
	pool := testutil.RequireDB(t)

	// Run migrations
	err = db.Migrate(ctx, pool)
	require.NoError(t, err, "Failed to run migrations")

	// Cleanup function
	cleanup := func() {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alert_suppressions CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alert_events CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS webhooks CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alerts CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
		pool.Close()
		testutil.TeardownTestConfig()
	}

	return pool, cleanup
}

func TestAlertSuppression_Integration(t *testing.T) {
	pool, cleanup := setupSuppressionTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test node
	nodeQuerier := db.NewPoolQuerier(pool)
	nodeID := uuid.New()
	err := nodeQuerier.CreateNode(ctx, nodeID, "test-node-1", "192.168.1.100", "us-west", map[string]any{
		"test": "true",
	})
	require.NoError(t, err)

	// Create alert rule
	alertQuerier := db.NewAlertQuerier(pool)
	nodeIDStr := nodeID.String()
	latencyRule := &models.Alert{
		Metric:    "latency",
		Threshold: 100.0,
		Level:     "P0",
		NodeID:    &nodeIDStr,
		Enabled:   true,
	}
	err = alertQuerier.CreateAlert(ctx, latencyRule)
	require.NoError(t, err)

	// Initialize alert engine with suppression
	config := alert.DefaultEngineConfig()
	config.WorkerPoolSize = 2
	config.MetricChannelBufferSize = 100
	alertEngine := alert.NewAlertEngine(pool, alertQuerier, config)
	alertEngine.Start()
	defer alertEngine.Stop()

	// Give workers time to initialize
	time.Sleep(100 * time.Millisecond)

	// Helper function to clean alert data between sub-tests
	cleanAlertData := func() {
		_, _ = pool.Exec(ctx, "DELETE FROM alert_suppressions")
		_, _ = pool.Exec(ctx, "DELETE FROM alert_events")
	}

	t.Run("First alert should not be suppressed", func(t *testing.T) {
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      150.0, // Exceeds 100ms threshold
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for evaluation
		time.Sleep(500 * time.Millisecond)

		// Verify suppression was recorded
		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
		suppression, err := suppressionQuerier.CheckSuppression(ctx, nodeID.String(), "latency")
		require.NoError(t, err)
		assert.NotNil(t, suppression, "Suppression record should exist")
		assert.True(t, time.Now().Before(suppression.SuppressedUntil), "Suppression should be active")
	})

	t.Run("Second alert within suppression window should be suppressed", func(t *testing.T) {
		// Don't clean - this test depends on previous test's suppression
		// Send another alert immediately (within 5 minute window)
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      160.0, // Also exceeds threshold
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for evaluation
		time.Sleep(500 * time.Millisecond)

		// Note: We can't directly verify suppression from the test,
		// but the suppression record should still exist with the same suppressed_until
		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
		suppression, err := suppressionQuerier.CheckSuppression(ctx, nodeID.String(), "latency")
		require.NoError(t, err)
		assert.NotNil(t, suppression, "Suppression record should still exist")
	})

	t.Run("Cleanup expired suppressions", func(t *testing.T) {
		cleanAlertData() // Clean data from previous tests
		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)

		// Create an expired suppression record manually
		err := suppressionQuerier.CreateOrUpdateSuppression(ctx, nodeID.String(), "jitter", time.Now().Add(-1*time.Minute))
		require.NoError(t, err)

		// Run cleanup
		deleted, err := suppressionQuerier.DeleteExpiredSuppressions(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), deleted, "Should delete 1 expired suppression")

		// Verify expired suppression is gone
		_, err = suppressionQuerier.CheckSuppression(ctx, nodeID.String(), "jitter")
		assert.Error(t, err, "Expired suppression should be deleted")
		assert.Equal(t, db.ErrSuppressionNotFound, err, "Should return ErrSuppressionNotFound")
	})

	t.Run("Different metric should not be suppressed", func(t *testing.T) {
		cleanAlertData() // Clean data from previous tests

		// Create alert rule for packet loss
		packetLossRule := &models.Alert{
			Metric:    "packet_loss_rate",
			Threshold: 5.0,
			Level:     "P1",
			NodeID:    &nodeIDStr,
			Enabled:   true,
		}
		err = alertQuerier.CreateAlert(ctx, packetLossRule)
		require.NoError(t, err)

		// Create a fresh alert engine to pick up the new packet_loss rule
		config := alert.DefaultEngineConfig()
		config.WorkerPoolSize = 2
		config.MetricChannelBufferSize = 100
		freshAlertEngine := alert.NewAlertEngine(pool, alertQuerier, config)
		freshAlertEngine.Start()
		defer freshAlertEngine.Stop()
		time.Sleep(100 * time.Millisecond)

		// Wait for rule cache refresh
		time.Sleep(200 * time.Millisecond)

		// Send packet loss alert (different metric)
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      50.0,
			PacketLossRate: 10.0, // Exceeds 5% threshold
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := freshAlertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for evaluation
		time.Sleep(500 * time.Millisecond)

		// Verify packet loss suppression was created
		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
		suppression, err := suppressionQuerier.CheckSuppression(ctx, nodeID.String(), "packet_loss_rate")
		require.NoError(t, err)
		assert.NotNil(t, suppression, "Packet loss suppression should exist")
	})

	t.Run("Different node should not be suppressed", func(t *testing.T) {
		cleanAlertData() // Clean data from previous tests

		// Create second node
		nodeID2 := uuid.New()
		err := nodeQuerier.CreateNode(ctx, nodeID2, "test-node-2", "192.168.1.101", "us-east", map[string]any{
			"test": "true",
		})
		require.NoError(t, err)

		// Create latency rule for node2
		nodeID2Str := nodeID2.String()
		node2LatencyRule := &models.Alert{
			Metric:    "latency",
			Threshold: 100.0,
			Level:     "P0",
			NodeID:    &nodeID2Str,
			Enabled:   true,
		}
		err = alertQuerier.CreateAlert(ctx, node2LatencyRule)
		require.NoError(t, err)

		// Create a fresh alert engine to pick up the new rule for node2
		config := alert.DefaultEngineConfig()
		config.WorkerPoolSize = 2
		config.MetricChannelBufferSize = 100
		freshAlertEngine := alert.NewAlertEngine(pool, alertQuerier, config)
		freshAlertEngine.Start()
		defer freshAlertEngine.Stop()
		time.Sleep(100 * time.Millisecond)

		// Wait for rule cache refresh
		time.Sleep(200 * time.Millisecond)

		// Send alert for node2 (different node, same metric)
		metricData := &alert.MetricData{
			NodeID:         nodeID2.String(),
			LatencyMs:      150.0,
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := freshAlertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for evaluation
		time.Sleep(500 * time.Millisecond)

		// Verify node2 has its own suppression record
		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
		suppression, err := suppressionQuerier.CheckSuppression(ctx, nodeID2.String(), "latency")
		require.NoError(t, err)
		assert.NotNil(t, suppression, "Node2 should have its own suppression record")
	})
}

func TestSuppressionService_Unit(t *testing.T) {
	t.Run("ShouldSuppress with no record", func(t *testing.T) {
		ctx := context.Background()
		pool, cleanup := setupSuppressionTestDB(t)
		defer cleanup()

		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
		service := suppression.NewService(suppressionQuerier)

		testNodeID := uuid.New().String()
		suppressed, err := service.ShouldSuppress(ctx, testNodeID, "latency")
		require.NoError(t, err)
		assert.False(t, suppressed, "Should not suppress when no record exists")
	})

	t.Run("RecordDefaultSuppression", func(t *testing.T) {
		ctx := context.Background()
		pool, cleanup := setupSuppressionTestDB(t)
		defer cleanup()

		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
		service := suppression.NewService(suppressionQuerier)

		// Create a test node first (required for foreign key constraint)
		nodeQuerier := db.NewPoolQuerier(pool)
		testNodeID := uuid.New()
		err := nodeQuerier.CreateNode(ctx, testNodeID, "test-node", "192.168.1.1", "us-west", nil)
		require.NoError(t, err)

		err = service.RecordDefaultSuppression(ctx, testNodeID.String(), "latency")
		require.NoError(t, err)

		// Verify suppression was created
		suppression, err := suppressionQuerier.CheckSuppression(ctx, testNodeID.String(), "latency")
		require.NoError(t, err)
		assert.NotNil(t, suppression, "Suppression should be created")
		assert.True(t, time.Now().Before(suppression.SuppressedUntil), "Suppression should be active")
		assert.True(t, time.Until(suppression.SuppressedUntil) < 6*time.Minute && time.Until(suppression.SuppressedUntil) > 4*time.Minute, "Suppression window should be ~5 minutes")
	})

	t.Run("ShouldSuppress within window", func(t *testing.T) {
		ctx := context.Background()
		pool, cleanup := setupSuppressionTestDB(t)
		defer cleanup()

		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
		service := suppression.NewService(suppressionQuerier)

		// Create a test node first (required for foreign key constraint)
		nodeQuerier := db.NewPoolQuerier(pool)
		testNodeID := uuid.New()
		err := nodeQuerier.CreateNode(ctx, testNodeID, "test-node", "192.168.1.1", "us-west", nil)
		require.NoError(t, err)

		// Create suppression
		err = service.RecordDefaultSuppression(ctx, testNodeID.String(), "latency")
		require.NoError(t, err)

		// Check suppression
		suppressed, err := service.ShouldSuppress(ctx, testNodeID.String(), "latency")
		require.NoError(t, err)
		assert.True(t, suppressed, "Should suppress within window")
	})

	t.Run("DeleteExpiredSuppressions", func(t *testing.T) {
		ctx := context.Background()
		pool, cleanup := setupSuppressionTestDB(t)
		defer cleanup()

		suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)

		// Create test nodes first (required for foreign key constraint)
		nodeQuerier := db.NewPoolQuerier(pool)
		testNodeID := uuid.New()
		err := nodeQuerier.CreateNode(ctx, testNodeID, "test-node-1", "192.168.1.1", "us-west", nil)
		require.NoError(t, err)

		testNodeID2 := uuid.New()
		err = nodeQuerier.CreateNode(ctx, testNodeID2, "test-node-2", "192.168.1.2", "us-west", nil)
		require.NoError(t, err)

		// Create expired suppression
		err = suppressionQuerier.CreateOrUpdateSuppression(ctx, testNodeID.String(), "latency", time.Now().Add(-1*time.Minute))
		require.NoError(t, err)

		// Create active suppression
		err = suppressionQuerier.CreateOrUpdateSuppression(ctx, testNodeID2.String(), "latency", time.Now().Add(5*time.Minute))
		require.NoError(t, err)

		// Run cleanup
		deleted, err := suppressionQuerier.DeleteExpiredSuppressions(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), deleted, "Should delete 1 expired suppression")

		// Verify expired is gone, active remains
		_, err = suppressionQuerier.CheckSuppression(ctx, testNodeID.String(), "latency")
		assert.Error(t, err, "Expired suppression should be deleted")

		suppression, err := suppressionQuerier.CheckSuppression(ctx, testNodeID2.String(), "latency")
		require.NoError(t, err)
		assert.NotNil(t, suppression, "Active suppression should remain")
	})
}

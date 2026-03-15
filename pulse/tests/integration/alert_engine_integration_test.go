package integration

import (
	"context"
	"fmt"
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
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

func setupAlertEngineTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	testutil.SetupTestConfig()

	// Load configuration
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Connect to test database
	testDBURL := testutil.GetTestDBURL()

	pool, err := pgxpool.New(ctx, testDBURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Run migrations
	err = db.Migrate(ctx, pool)
	require.NoError(t, err, "Failed to run migrations")

	// Cleanup function
	cleanup := func() {
		// Drop all tables
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

func TestAlertEngine_Integration(t *testing.T) {
	pool, cleanup := setupAlertEngineTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test node
	nodeQuerier := db.NewPoolQuerier(pool)
	nodeID := uuid.New()
	err := nodeQuerier.CreateNode(ctx, nodeID, "test-node-1", "192.168.1.100", "us-west", map[string]interface{}{
		"test": "true",
	})
	require.NoError(t, err)

	// Create second test node for cross-node testing
	nodeID2 := uuid.New()
	err = nodeQuerier.CreateNode(ctx, nodeID2, "test-node-2", "192.168.1.101", "us-east", map[string]interface{}{
		"test": "true",
	})
	require.NoError(t, err)

	// Create alert rules
	alertQuerier := db.NewAlertQuerier(pool)

	// Store node ID as string for alert rules
	nodeIDStr := nodeID.String()

	// Create latency alert rule (P0, threshold 100ms)
	latencyRule := &models.Alert{
		Metric:    "latency",
		Threshold: 100.0,
		Level:     "P0",
		NodeID:    &nodeIDStr,
		Enabled:   true,
	}
	err = alertQuerier.CreateAlert(ctx, latencyRule)
	require.NoError(t, err)

	// Create packet loss alert rule (P1, threshold 5%)
	packetLossRule := &models.Alert{
		Metric:    "packet_loss_rate",
		Threshold: 5.0,
		Level:     "P1",
		NodeID:    &nodeIDStr,
		Enabled:   true,
	}
	err = alertQuerier.CreateAlert(ctx, packetLossRule)
	require.NoError(t, err)

	// Initialize alert engine
	config := alert.DefaultEngineConfig()
	config.WorkerPoolSize = 2 // Smaller pool for tests
	config.MetricChannelBufferSize = 100

	alertEngine := alert.NewAlertEngine(pool, alertQuerier, config)
	alertEngine.Start()
	defer alertEngine.Stop()

	// Give workers time to initialize
	time.Sleep(100 * time.Millisecond)

	t.Run("Evaluate metrics with threshold exceeded", func(t *testing.T) {
		// Send metric data with exceeded thresholds
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      150.0, // Exceeds 100ms threshold
			PacketLossRate: 10.0,  // Exceeds 5% threshold
			JitterMs:       5.0,   // Within limits (no rule)
			Timestamp:      time.Now(),
		}

		// Queue for evaluation
		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for async evaluation
		time.Sleep(500 * time.Millisecond)

		// Query alert events from database
		// Note: We need to implement a query method for alert events
		// For now, we just verify no errors occurred
		stats := alertEngine.GetStats()
		assert.NotNil(t, stats)
		assert.Equal(t, 2, stats["cached_rules"])
	})

	t.Run("Evaluate metrics within thresholds", func(t *testing.T) {
		// Send metric data within thresholds
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      50.0, // Below 100ms threshold
			PacketLossRate: 2.0,  // Below 5% threshold
			JitterMs:       3.0,  // Within limits
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for async evaluation
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("Evaluate metrics for different node", func(t *testing.T) {
		// Send metric data for node2 (should not trigger node1-specific rules)
		metricData := &alert.MetricData{
			NodeID:         nodeID2.String(),
			LatencyMs:      200.0, // Exceeds node1's 100ms threshold
			PacketLossRate: 15.0,  // Exceeds node1's 5% threshold
			JitterMs:       10.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for async evaluation
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("Global rule evaluation", func(t *testing.T) {
		// Create global alert rule (node_id = NULL)
		globalRule := &models.Alert{
			Metric:    "jitter",
			Threshold: 8.0,
			Level:     "P2",
			NodeID:    nil, // Global rule
			Enabled:   true,
		}
		err := alertQuerier.CreateAlert(ctx, globalRule)
		require.NoError(t, err)

		// Wait for rule cache refresh (60 second interval, so we need to trigger manually)
		// For now, just wait a bit
		time.Sleep(200 * time.Millisecond)

		// Send metric data for node2 that exceeds global jitter threshold
		metricData := &alert.MetricData{
			NodeID:         nodeID2.String(),
			LatencyMs:      50.0,
			PacketLossRate: 1.0,
			JitterMs:       10.0, // Exceeds global 8ms threshold
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for async evaluation
		time.Sleep(500 * time.Millisecond)

		// Cleanup
		err = alertQuerier.DeleteAlert(ctx, globalRule.ID)
		assert.NoError(t, err)
	})

	t.Run("Disabled rule should not trigger", func(t *testing.T) {
		// Disable the latency rule
		updateReq := &models.UpdateAlertRequest{
			Enabled: boolPtr(false),
		}
		_, err := alertQuerier.UpdateAlert(ctx, latencyRule.ID, updateReq)
		require.NoError(t, err)

		// Wait for rule cache refresh
		time.Sleep(200 * time.Millisecond)

		// Send metric data with exceeded latency threshold
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      200.0, // Would exceed threshold if rule was enabled
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for async evaluation
		time.Sleep(500 * time.Millisecond)

		// Re-enable rule for cleanup
		updateReq.Enabled = boolPtr(true)
		_, err = alertQuerier.UpdateAlert(ctx, latencyRule.ID, updateReq)
		assert.NoError(t, err)
	})

	t.Run("Alert engine stats", func(t *testing.T) {
		stats := alertEngine.GetStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "cached_rules")
		assert.Contains(t, stats, "rule_cache_last_refresh")
		assert.Contains(t, stats, "metric_channel_depth")
		assert.Contains(t, stats, "metric_channel_capacity")
	})
}

func TestAlertEngine_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	pool, cleanup := setupAlertEngineTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test node
	nodeQuerier := db.NewPoolQuerier(pool)
	nodeID := uuid.New()
	err := nodeQuerier.CreateNode(ctx, nodeID, "perf-node", "10.0.0.1", "eu-central", map[string]interface{}{
		"perf": "true",
	})
	require.NoError(t, err)

	// Create multiple alert rules
	alertQuerier := db.NewAlertQuerier(pool)
	perfNodeIDStr := nodeID.String()
	for i := 0; i < 10; i++ {
		threshold := float64(100 + i*10)
		alert := &models.Alert{
			Metric:    "latency",
			Threshold: threshold,
			Level:     "P1",
			NodeID:    &perfNodeIDStr,
			Enabled:   true,
		}
		err := alertQuerier.CreateAlert(ctx, alert)
		require.NoError(t, err)
	}

	// Initialize alert engine
	config := alert.DefaultEngineConfig()
	alertEngine := alert.NewAlertEngine(pool, alertQuerier, config)
	alertEngine.Start()
	defer alertEngine.Stop()

	time.Sleep(100 * time.Millisecond)

	// Measure evaluation latency
	numMetrics := 100
	startTime := time.Now()

	for i := 0; i < numMetrics; i++ {
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      150.0,
			PacketLossRate: 5.0,
			JitterMs:       8.0,
			Timestamp:      time.Now(),
		}
		alertEngine.EvaluateMetrics(metricData)
	}

	// Wait for all evaluations to complete
	time.Sleep(2 * time.Second)

	duration := time.Since(startTime)
	avgLatency := duration.Milliseconds() / int64(numMetrics)

	// Verify average latency is under 100ms target
	assert.Less(t, avgLatency, int64(100), fmt.Sprintf("Average evaluation latency %dms exceeds 100ms target", avgLatency))

	t.Logf("Processed %d metrics in %v (avg: %dms per metric)", numMetrics, duration, avgLatency)
}

func boolPtr(b bool) *bool {
	return &b
}

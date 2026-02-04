package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/alert"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
	"github.com/kevin/node-pulse/pulse-api/internal/webhook"
)

func setupWebhookPushTestDB(t *testing.T) (*pgxpool.Pool, func()) {
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
		pool.Exec(ctx, "DROP TABLE IF EXISTS webhook_logs CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS webhooks CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS alert_suppressions CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS alert_events CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS alerts CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
		pool.Close()
	}

	return pool, cleanup
}

func TestWebhookPush_Integration(t *testing.T) {
	pool, cleanup := setupWebhookPushTestDB(t)
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

	// Track webhook deliveries
	var webhookDeliveries []map[string]any
	var webhookMutex sync.Mutex

	// Helper function to clean alert data between sub-tests
	cleanAlertData := func() {
		pool.Exec(ctx, "DELETE FROM alert_suppressions")
		pool.Exec(ctx, "DELETE FROM alert_events")
		pool.Exec(ctx, "DELETE FROM webhook_logs")
		webhookMutex.Lock()
		webhookDeliveries = nil
		webhookMutex.Unlock()
	}

	// Create test webhook server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)

		// Verify payload structure
		assert.Equal(t, "1.0", payload["version"])
		alertData, ok := payload["alert"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "latency", alertData["metric"])
		assert.Equal(t, nodeIDStr, alertData["node_id"])

		// Track delivery
		webhookMutex.Lock()
		webhookDeliveries = append(webhookDeliveries, payload)
		webhookMutex.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook in database
	webhooksQuerier := db.NewWebhookQuerier(pool)
	testWebhook := &models.Webhook{
		URL:     server.URL,
		Enabled: true,
	}
	err = webhooksQuerier.CreateWebhook(ctx, testWebhook)
	require.NoError(t, err)

	// Initialize alert engine
	config := alert.DefaultEngineConfig()
	config.WorkerPoolSize = 2
	config.MetricChannelBufferSize = 100
	alertEngine := alert.NewAlertEngine(pool, alertQuerier, config)
	alertEngine.Start()
	defer alertEngine.Stop()

	// Give workers time to initialize
	time.Sleep(100 * time.Millisecond)

	t.Run("Webhook should be triggered on alert event creation", func(t *testing.T) {
		// Send metric that exceeds threshold
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      150.0, // Exceeds 100ms threshold
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for alert evaluation and webhook delivery
		time.Sleep(1 * time.Second)

		// Verify webhook was called
		webhookMutex.Lock()
		assert.GreaterOrEqual(t, len(webhookDeliveries), 1, "Webhook should be called at least once")
		webhookMutex.Unlock()

		// Verify alert event was created
		var eventCount int
		err = pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM alert_events
			WHERE node_id = $1
		`, nodeIDStr).Scan(&eventCount)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, eventCount, 1, "Alert event should be created")

		// Verify webhook log was created
		// Note: GetWebhookLogs doesn't exist yet, so we'll verify via direct query
		rows, err := pool.Query(ctx, `
			SELECT id, webhook_id, alert_event_id, status, retry_count, error_message, sent_at
			FROM webhook_logs
			WHERE webhook_id = $1
			ORDER BY sent_at DESC
			LIMIT 1
		`, testWebhook.ID)
		require.NoError(t, err)
		defer rows.Close()

		if rows.Next() {
			var log struct {
				ID           string
				WebhookID    string
				AlertEventID string
				Status       string
				RetryCount   int
				ErrorMessage string
				SentAt       time.Time
			}
			err = rows.Scan(
				&log.ID,
				&log.WebhookID,
				&log.AlertEventID,
				&log.Status,
				&log.RetryCount,
				&log.ErrorMessage,
				&log.SentAt,
			)
			require.NoError(t, err)

			assert.Equal(t, testWebhook.ID, log.WebhookID)
			assert.Equal(t, "success", log.Status, "Webhook delivery should succeed")
			assert.Equal(t, 0, log.RetryCount, "First attempt should succeed")
			assert.Equal(t, "", log.ErrorMessage, "No error message on success")
		} else {
			t.Fatal("Webhook log should be created")
		}
	})

	t.Run("Multiple webhooks should all be triggered", func(t *testing.T) {
		cleanAlertData() // Clean data from previous test
		// Create additional webhooks
		webhook2 := &models.Webhook{
			URL:     server.URL,
			Enabled: true,
		}
		err = webhooksQuerier.CreateWebhook(ctx, webhook2)
		require.NoError(t, err)

		webhook3 := &models.Webhook{
			URL:     server.URL,
			Enabled: true,
		}
		err = webhooksQuerier.CreateWebhook(ctx, webhook3)
		require.NoError(t, err)

		// Reset webhook deliveries counter
		webhookMutex.Lock()
		webhookDeliveries = nil
		webhookMutex.Unlock()

		// Wait for rule cache refresh
		time.Sleep(200 * time.Millisecond)

		// Send metric that exceeds threshold
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      160.0, // Exceeds 100ms threshold
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for alert evaluation and webhook delivery
		time.Sleep(1 * time.Second)

		// Verify webhook was called at least 3 times (once per webhook)
		webhookMutex.Lock()
		assert.GreaterOrEqual(t, len(webhookDeliveries), 3, "All 3 webhooks should be called")
		webhookMutex.Unlock()
	})

	t.Run("Disabled webhook should not be triggered", func(t *testing.T) {
		cleanAlertData() // Clean data from previous test
		// Disable the first webhook
		enabled := false
		update := &models.UpdateWebhookRequest{
			Enabled: &enabled,
		}
		_, err = webhooksQuerier.UpdateWebhook(ctx, testWebhook.ID, update)
		require.NoError(t, err)

		// Reset webhook deliveries counter
		webhookMutex.Lock()
		prevCount := len(webhookDeliveries)
		webhookMutex.Unlock()

		// Wait for cache refresh
		time.Sleep(200 * time.Millisecond)

		// Send metric that exceeds threshold
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      170.0, // Exceeds 100ms threshold
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for alert evaluation and webhook delivery
		time.Sleep(1 * time.Second)

		// Verify only 2 webhooks were called (webhook1 is disabled)
		webhookMutex.Lock()
		newCount := len(webhookDeliveries)
		webhookMutex.Unlock()

		// Should have 2 more deliveries (webhook2 and webhook3)
		assert.Equal(t, prevCount+2, newCount, "Only enabled webhooks should be called")
	})

	t.Run("Suppressed alerts should not trigger webhooks", func(t *testing.T) {
		cleanAlertData() // Clean data from previous test
		// Re-enable webhook1
		enabled := true
		update := &models.UpdateWebhookRequest{
			Enabled: &enabled,
		}
		_, err = webhooksQuerier.UpdateWebhook(ctx, testWebhook.ID, update)
		require.NoError(t, err)

		// Wait for cache refresh
		time.Sleep(200 * time.Millisecond)

		// Reset webhook deliveries counter
		webhookMutex.Lock()
		webhookDeliveries = nil
		webhookMutex.Unlock()

		// Send first metric
		metricData := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      180.0,
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success := alertEngine.EvaluateMetrics(metricData)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for alert evaluation and webhook delivery
		time.Sleep(1 * time.Second)

		webhookMutex.Lock()
		firstCount := len(webhookDeliveries)
		webhookMutex.Unlock()

		assert.GreaterOrEqual(t, firstCount, 3, "All webhooks should be called for first alert")

		// Send second metric immediately (within 5-minute suppression window)
		metricData2 := &alert.MetricData{
			NodeID:         nodeID.String(),
			LatencyMs:      190.0,
			PacketLossRate: 1.0,
			JitterMs:       2.0,
			Timestamp:      time.Now(),
		}

		success = alertEngine.EvaluateMetrics(metricData2)
		assert.True(t, success, "Metric should be queued successfully")

		// Wait for alert evaluation
		time.Sleep(500 * time.Millisecond)

		webhookMutex.Lock()
		secondCount := len(webhookDeliveries)
		webhookMutex.Unlock()

		// Webhook count should not increase (alert was suppressed)
		assert.Equal(t, firstCount, secondCount, "Suppressed alert should not trigger webhooks")
	})
}

func TestWebhookPush_RetryLogic(t *testing.T) {
	pool, cleanup := setupWebhookPushTestDB(t)
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

	// Track attempt count
	attempts := 0

	// Create test webhook server that fails first 3 times, succeeds on 4th
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook in database
	webhooksQuerier := db.NewWebhookQuerier(pool)
	testWebhook := &models.Webhook{
		URL:     server.URL,
		Enabled: true,
	}
	err = webhooksQuerier.CreateWebhook(ctx, testWebhook)
	require.NoError(t, err)

	// Initialize push service
	webhookLogsQuerier := db.NewWebhookLogsQuerier(pool)
	pushService := webhook.NewPushService(webhooksQuerier, webhookLogsQuerier, "http://localhost:8080")

	// Create alert event in database (required for foreign key constraint)
	alertEventsQuerier := db.NewAlertEventsQuerier(pool)
	alertEvent := &models.AlertEvent{
		ID:           uuid.New().String(),
		NodeID:       nodeID.String(),
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	t.Run("Webhook should succeed after retries", func(t *testing.T) {
		startTime := time.Now()

		// Send webhook (will retry and eventually succeed)
		err = pushService.SendAlert(ctx, alertEvent)
		require.NoError(t, err, "Webhook should succeed after retries")

		duration := time.Since(startTime)
		assert.GreaterOrEqual(t, duration, 7*time.Second, "Should take at least 7 seconds for retries")
		assert.Less(t, duration, 10*time.Second, "Should complete within 10 seconds")

		// Verify 4 attempts were made (1 initial + 3 retries)
		assert.Equal(t, 4, attempts)

		// Verify success log was created
		rows, err := pool.Query(ctx, `
			SELECT status, retry_count
			FROM webhook_logs
			WHERE webhook_id = $1
			ORDER BY sent_at DESC
			LIMIT 1
		`, testWebhook.ID)
		require.NoError(t, err)
		defer rows.Close()

		if rows.Next() {
			var status string
			var retryCount int
			err = rows.Scan(&status, &retryCount)
			require.NoError(t, err)

			assert.Equal(t, "success", status)
			// retry_count should be 3 (indicating 3 retries were needed)
			assert.Equal(t, 3, retryCount)
		}
	})
}

func TestWebhookPush_ConcurrentDelivery(t *testing.T) {
	pool, cleanup := setupWebhookPushTestDB(t)
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

	// Track delivery order and timing
	var deliveries []string
	var deliveryTimes []time.Time
	var deliveryMutex sync.Mutex

	// Create test webhook server with variable delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deliveryMutex.Lock()
		deliveries = append(deliveries, r.URL.String())
		deliveryTimes = append(deliveryTimes, time.Now())
		deliveryMutex.Unlock()

		// Add small delay to simulate processing
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create multiple webhooks
	webhooksQuerier := db.NewWebhookQuerier(pool)
	for i := 1; i <= 5; i++ {
		webhook := &models.Webhook{
			URL:     server.URL,
			Enabled: true,
		}
		err = webhooksQuerier.CreateWebhook(ctx, webhook)
		require.NoError(t, err)
	}

	// Initialize push service
	webhookLogsQuerier := db.NewWebhookLogsQuerier(pool)
	pushService := webhook.NewPushService(webhooksQuerier, webhookLogsQuerier, "http://localhost:8080")

	// Create alert event in database (required for foreign key constraint)
	alertEventsQuerier := db.NewAlertEventsQuerier(pool)
	alertEvent := &models.AlertEvent{
		ID:           uuid.New().String(),
		NodeID:       nodeID.String(),
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	t.Run("Multiple webhooks should be delivered concurrently", func(t *testing.T) {
		startTime := time.Now()

		err = pushService.SendAlert(ctx, alertEvent)
		require.NoError(t, err)

		duration := time.Since(startTime)

		// With 5 webhooks each taking 100ms:
		// Serial delivery would take ~500ms
		// Concurrent delivery should take ~100-150ms
		assert.Less(t, duration, 250*time.Millisecond, "Webhooks should be delivered concurrently")

		// Verify all 5 webhooks were called
		assert.Len(t, deliveries, 5)

		// Verify logs were created for all webhooks
		rows, err := pool.Query(ctx, `
			SELECT COUNT(*)
			FROM webhook_logs
			WHERE alert_event_id = $1
		`, alertEvent.ID)
		require.NoError(t, err)
		defer rows.Close()

		if rows.Next() {
			var count int64
			err = rows.Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, int64(5), count)
		}
	})
}

package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

// Mock WebhookLogsQuerier for testing
type mockWebhookLogsQuerier struct {
	logsCreated []*models.WebhookLog
}

func (m *mockWebhookLogsQuerier) CreateWebhookLog(ctx context.Context, log *models.WebhookLog) error {
	m.logsCreated = append(m.logsCreated, log)
	return nil
}

// Mock WebhookQuerier for testing
type mockWebhookQuerier struct {
	webhooks []*models.Webhook
}

func (m *mockWebhookQuerier) GetWebhooks(ctx context.Context) ([]*models.Webhook, error) {
	return m.webhooks, nil
}

func (m *mockWebhookQuerier) CreateWebhook(ctx context.Context, webhook *models.Webhook) error {
	return nil
}

func (m *mockWebhookQuerier) GetWebhookByID(ctx context.Context, id string) (*models.Webhook, error) {
	for _, wh := range m.webhooks {
		if wh.ID == id {
			return wh, nil
		}
	}
	return nil, nil
}

func (m *mockWebhookQuerier) UpdateWebhook(ctx context.Context, id string, update *models.UpdateWebhookRequest) (*models.Webhook, error) {
	return nil, nil
}

func (m *mockWebhookQuerier) DeleteWebhook(ctx context.Context, id string) error {
	return nil
}

func TestPushService_SendWebhook_Success(t *testing.T) {
	// Create test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify payload structure
		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)

		assert.Equal(t, "1.0", payload["version"])
		alert, ok := payload["alert"].(map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, alert["id"])
		assert.Equal(t, "latency", alert["metric"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier:    &mockWebhookQuerier{},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx := context.Background()
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	webhook := &models.Webhook{
		ID:      "webhook-1",
		URL:     server.URL,
		Enabled: true,
	}

	err := service.SendWebhook(ctx, alertEvent, webhook)
	require.NoError(t, err)

	// Verify success log was created
	require.Len(t, mockLogs.logsCreated, 1)
	assert.Equal(t, "success", mockLogs.logsCreated[0].Status)
	assert.Equal(t, 0, mockLogs.logsCreated[0].RetryCount)
	assert.Equal(t, "", mockLogs.logsCreated[0].ErrorMessage)
}

func TestPushService_SendWebhook_RetrySuccess(t *testing.T) {
	attempts := 0

	// Create test server that fails first attempt, succeeds on second
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier:    &mockWebhookQuerier{},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx := context.Background()
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	webhook := &models.Webhook{
		ID:      "webhook-1",
		URL:     server.URL,
		Enabled: true,
	}

	startTime := time.Now()
	err := service.SendWebhook(ctx, alertEvent, webhook)
	require.NoError(t, err)

	// Should have taken at least 1 second (backoff)
	duration := time.Since(startTime)
	assert.GreaterOrEqual(t, duration, 1*time.Second)

	// Verify success log was created
	require.Len(t, mockLogs.logsCreated, 1)
	assert.Equal(t, "success", mockLogs.logsCreated[0].Status)
}

func TestPushService_SendWebhook_MaxRetriesExceeded(t *testing.T) {
	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier:    &mockWebhookQuerier{},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx := context.Background()
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	webhook := &models.Webhook{
		ID:      "webhook-1",
		URL:     server.URL,
		Enabled: true,
	}

	startTime := time.Now()
	err := service.SendWebhook(ctx, alertEvent, webhook)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook delivery failed after 4 attempts")

	// Should have taken at least 7 seconds (1s + 2s + 4s backoffs)
	duration := time.Since(startTime)
	assert.GreaterOrEqual(t, duration, 7*time.Second)

	// Verify failure log was created with max retries
	require.Len(t, mockLogs.logsCreated, 1)
	assert.Equal(t, "failure", mockLogs.logsCreated[0].Status)
	assert.Equal(t, 3, mockLogs.logsCreated[0].RetryCount)
	assert.NotEmpty(t, mockLogs.logsCreated[0].ErrorMessage)
}

func TestPushService_SendWebhook_ContextCancellation(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier:    &mockWebhookQuerier{},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	webhook := &models.Webhook{
		ID:      "webhook-1",
		URL:     server.URL,
		Enabled: true,
	}

	err := service.SendWebhook(ctx, alertEvent, webhook)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestPushService_SendAlert_NoWebhooks(t *testing.T) {
	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier:    &mockWebhookQuerier{webhooks: []*models.Webhook{}},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx := context.Background()
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	err := service.SendAlert(ctx, alertEvent)
	require.NoError(t, err, "Should return nil when no webhooks configured")

	// No logs should be created
	assert.Len(t, mockLogs.logsCreated, 0)
}

func TestPushService_SendAlert_MultipleWebhooks(t *testing.T) {
	// Create test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier: &mockWebhookQuerier{
			webhooks: []*models.Webhook{
				{ID: "webhook-1", URL: server.URL, Enabled: true},
				{ID: "webhook-2", URL: server.URL, Enabled: true},
				{ID: "webhook-3", URL: server.URL, Enabled: true},
			},
		},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx := context.Background()
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	err := service.SendAlert(ctx, alertEvent)
	require.NoError(t, err)

	// Verify all 3 webhooks were logged
	assert.Len(t, mockLogs.logsCreated, 3)
	for _, log := range mockLogs.logsCreated {
		assert.Equal(t, "success", log.Status)
	}
}

func TestPushService_SendAlert_OnlyEnabledWebhooks(t *testing.T) {
	// Create test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier: &mockWebhookQuerier{
			webhooks: []*models.Webhook{
				{ID: "webhook-1", URL: server.URL, Enabled: true},
				{ID: "webhook-2", URL: server.URL, Enabled: false}, // Disabled
				{ID: "webhook-3", URL: server.URL, Enabled: true},
			},
		},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx := context.Background()
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	err := service.SendAlert(ctx, alertEvent)
	require.NoError(t, err)

	// Verify only 2 enabled webhooks were logged
	assert.Len(t, mockLogs.logsCreated, 2)
}

func TestPushService_SendAlert_PartialFailure(t *testing.T) {
	// Create test server that fails for webhook-2
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	failureServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failureServer.Close()

	mockLogs := &mockWebhookLogsQuerier{}
	service := &PushService{
		webhookQuerier: &mockWebhookQuerier{
			webhooks: []*models.Webhook{
				{ID: "webhook-1", URL: server.URL, Enabled: true},
				{ID: "webhook-2", URL: failureServer.URL, Enabled: true}, // Will fail
				{ID: "webhook-3", URL: server.URL, Enabled: true},
			},
		},
		webhookLogsQuerier: mockLogs,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		baseURL:            "http://localhost:8080",
	}

	ctx := context.Background()
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	// This will take ~7 seconds due to retries
	startTime := time.Now()
	err := service.SendAlert(ctx, alertEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook push completed with 1 failures out of 3 webhooks")
	duration := time.Since(startTime)
	assert.GreaterOrEqual(t, duration, 7*time.Second)

	// Verify all 3 webhooks were logged
	assert.Len(t, mockLogs.logsCreated, 3)

	// webhook-1 and webhook-3 should succeed
	assert.Equal(t, "success", mockLogs.logsCreated[0].Status)
	assert.Equal(t, "success", mockLogs.logsCreated[2].Status)

	// webhook-2 should fail
	assert.Equal(t, "failure", mockLogs.logsCreated[1].Status)
	assert.Equal(t, 3, mockLogs.logsCreated[1].RetryCount)
}

func TestPushService_formatAlertEvent(t *testing.T) {
	service := &PushService{
		baseURL: "http://localhost:8080",
	}

	alertEvent := &models.AlertEvent{
		ID:           "test-alert-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	webhook := &models.Webhook{
		ID:      "webhook-1",
		URL:     "https://example.com/webhook",
		Enabled: true,
	}

	formatted, err := service.formatAlertEvent(alertEvent, webhook)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "1.0", formatted["version"])

	alert, ok := formatted["alert"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test-alert-1", alert["id"])
	assert.Equal(t, "latency", alert["metric"])
	assert.Equal(t, 100.0, alert["threshold"])
	assert.Equal(t, 150.0, alert["current_value"])
	assert.Equal(t, "P0", alert["level"])
	assert.Equal(t, "node-1", alert["node_id"])
	assert.Equal(t, "2024-01-15T10:30:00Z", alert["triggered_at"])

	links, ok := formatted["links"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "http://localhost:8080/nodes/node-1", links["alert_details"])
	assert.Equal(t, "http://localhost:8080", links["dashboard"])
}

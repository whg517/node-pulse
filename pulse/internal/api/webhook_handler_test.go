package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

func setupWebhookHandlerTest() (*gin.Engine, *db.MockWebhookQuerier) {
	router, mockQuerier, _ := setupWebhookHandlerTestWithHandler()
	return router, mockQuerier
}

func setupWebhookHandlerTestWithHandler() (*gin.Engine, *db.MockWebhookQuerier, *WebhookHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockQuerier := &db.MockWebhookQuerier{
		Webhooks: make(map[string]*models.Webhook),
	}

	handler := NewWebhookHandler(mockQuerier)

	router.POST("/webhooks", handler.CreateWebhookHandler)
	router.GET("/webhooks", handler.GetWebhooksHandler)
	router.GET("/webhooks/:id", handler.GetWebhookByIDHandler)
	router.PUT("/webhooks/:id", handler.UpdateWebhookHandler)
	router.DELETE("/webhooks/:id", handler.DeleteWebhookHandler)
	router.POST("/webhooks/preview", handler.PreviewWebhookEventHandler)
	router.POST("/webhooks/:id/test", handler.TestWebhookHandler)

	return router, mockQuerier, handler
}

func TestCreateWebhookHandler_ValidHTTPSURL(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	reqBody := models.CreateWebhookRequest{
		URL: "https://example.com/webhook",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.CreateWebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Data.Webhook.ID)
	assert.Equal(t, "https://example.com/webhook", response.Data.Webhook.URL)
	assert.True(t, response.Data.Webhook.Enabled)
}

func TestCreateWebhookHandler_HTTPURLRejected(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	reqBody := models.CreateWebhookRequest{
		URL: "http://example.com/webhook", // Not HTTPS
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ERR_INVALID_URL", response["code"])
	assert.Contains(t, response["details"].(string), "HTTPS")
}

func TestCreateWebhookHandler_InvalidURLFormat(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	reqBody := models.CreateWebhookRequest{
		URL: "not-a-valid-url",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateWebhookHandler_CustomEventFormat(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	customFormat := map[string]interface{}{
		"version": "2.0",
		"custom":  "format",
	}

	reqBody := models.CreateWebhookRequest{
		URL:         "https://example.com/webhook",
		EventFormat: customFormat,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.CreateWebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Data.Webhook.EventFormat)
}

func TestCreateWebhookHandler_DefaultEnabled(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	reqBody := models.CreateWebhookRequest{
		URL: "https://example.com/webhook",
		// Enabled not provided, should default to true
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.CreateWebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Data.Webhook.Enabled)
}

func TestGetWebhooksHandler_Success(t *testing.T) {
	router, mockQuerier := setupWebhookHandlerTest()

	// Create test webhooks
	webhook1 := &models.Webhook{
		ID:      "test-id-1",
		URL:     "https://example.com/hook1",
		Enabled: true,
	}
	webhook2 := &models.Webhook{
		ID:      "test-id-2",
		URL:     "https://example.com/hook2",
		Enabled: false,
	}
	mockQuerier.Webhooks[webhook1.ID] = webhook1
	mockQuerier.Webhooks[webhook2.ID] = webhook2

	req := httptest.NewRequest("GET", "/webhooks", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetWebhooksResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data.Webhooks, 2)
}

func TestPreviewWebhookEventHandler_DefaultFormat(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	req := httptest.NewRequest("POST", "/webhooks/preview", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "pulse.example.com"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PreviewWebhookEventResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	alert, ok := response.Data.Payload["alert"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "preview-alert-1", alert["id"])
	assert.Equal(t, "latency", alert["metric"])
	assert.Equal(t, float64(100), alert["threshold"])

	links, ok := response.Data.Payload["links"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http://pulse.example.com/nodes/preview-node-1", links["alert_details"])
}

func TestPreviewWebhookEventHandler_CustomFormat(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	reqBody := models.PreviewWebhookEventRequest{
		EventFormat: map[string]interface{}{
			"text":     "Alert {{.AlertID}} on {{.NodeID}}",
			"severity": "{{.Level}}",
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/webhooks/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PreviewWebhookEventResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Alert preview-alert-1 on preview-node-1", response.Data.Payload["text"])
	assert.Equal(t, "P1", response.Data.Payload["severity"])
}

func TestTestWebhookHandler_Success(t *testing.T) {
	router, mockQuerier, handler := setupWebhookHandlerTestWithHandler()

	webhook := &models.Webhook{
		ID:      "test-id",
		URL:     "https://example.com/webhook",
		Enabled: true,
	}
	mockQuerier.Webhooks[webhook.ID] = webhook

	handler.testSender = func(ctx context.Context, event *models.AlertEvent, target *models.Webhook, baseURL string) error {
		assert.Equal(t, webhook.ID, target.ID)
		assert.NotEmpty(t, event.ID)
		assert.NotEmpty(t, baseURL)
		return nil
	}

	req := httptest.NewRequest("POST", "/webhooks/test-id/test", nil)
	req.Host = "pulse.example.com"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TestWebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-id", response.Data.WebhookID)
	assert.Equal(t, "success", response.Data.Status)
}

func TestTestWebhookHandler_DeliveryFailure(t *testing.T) {
	router, mockQuerier, handler := setupWebhookHandlerTestWithHandler()

	webhook := &models.Webhook{
		ID:      "test-id",
		URL:     "https://example.com/webhook",
		Enabled: true,
	}
	mockQuerier.Webhooks[webhook.ID] = webhook
	handler.testSender = func(context.Context, *models.AlertEvent, *models.Webhook, string) error {
		return errors.New("endpoint rejected request")
	}

	req := httptest.NewRequest("POST", "/webhooks/test-id/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var response models.TestWebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-id", response.Data.WebhookID)
	assert.Equal(t, "failure", response.Data.Status)
	assert.Contains(t, response.Data.Error, "endpoint rejected request")
}

func TestTestWebhookHandler_NotFound(t *testing.T) {
	router, _, _ := setupWebhookHandlerTestWithHandler()

	req := httptest.NewRequest("POST", "/webhooks/non-existent/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetWebhookByIDHandler_Success(t *testing.T) {
	router, mockQuerier := setupWebhookHandlerTest()

	webhook := &models.Webhook{
		ID:      "test-id",
		URL:     "https://example.com/webhook",
		Enabled: true,
	}
	mockQuerier.Webhooks[webhook.ID] = webhook

	req := httptest.NewRequest("GET", "/webhooks/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UpdateWebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, webhook.ID, response.Data.Webhook.ID)
	assert.Equal(t, webhook.URL, response.Data.Webhook.URL)
}

func TestGetWebhookByIDHandler_NotFound(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	req := httptest.NewRequest("GET", "/webhooks/non-existent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ERR_NOT_FOUND", response["code"])
}

func TestUpdateWebhookHandler_ValidHTTPSURL(t *testing.T) {
	router, mockQuerier := setupWebhookHandlerTest()

	webhook := &models.Webhook{
		ID:      "test-id",
		URL:     "https://example.com/old",
		Enabled: true,
	}
	mockQuerier.Webhooks[webhook.ID] = webhook

	newURL := "https://example.com/new"
	reqBody := models.UpdateWebhookRequest{
		URL: &newURL,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/webhooks/test-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UpdateWebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, newURL, response.Data.Webhook.URL)
}

func TestUpdateWebhookHandler_HTTPURLRejected(t *testing.T) {
	router, mockQuerier := setupWebhookHandlerTest()

	webhook := &models.Webhook{
		ID:      "test-id",
		URL:     "https://example.com/old",
		Enabled: true,
	}
	mockQuerier.Webhooks[webhook.ID] = webhook

	newURL := "http://example.com/new" // Not HTTPS
	reqBody := models.UpdateWebhookRequest{
		URL: &newURL,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/webhooks/test-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ERR_INVALID_URL", response["code"])
}

func TestUpdateWebhookHandler_NotFound(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	newURL := "https://example.com/new"
	reqBody := models.UpdateWebhookRequest{
		URL: &newURL,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/webhooks/non-existent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteWebhookHandler_Success(t *testing.T) {
	router, mockQuerier := setupWebhookHandlerTest()

	webhook := &models.Webhook{
		ID:      "test-id",
		URL:     "https://example.com/webhook",
		Enabled: true,
	}
	mockQuerier.Webhooks[webhook.ID] = webhook

	req := httptest.NewRequest("DELETE", "/webhooks/test-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify webhook was deleted
	_, exists := mockQuerier.Webhooks["test-id"]
	assert.False(t, exists)
}

func TestDeleteWebhookHandler_NotFound(t *testing.T) {
	router, _ := setupWebhookHandlerTest()

	req := httptest.NewRequest("DELETE", "/webhooks/non-existent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ERR_NOT_FOUND", response["code"])
}

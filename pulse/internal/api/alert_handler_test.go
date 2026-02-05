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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// MockAlertQuerier is a mock for AlertQuerier interface
type MockAlertQuerier struct {
	createAlertFunc  func(context.Context, *models.Alert) error
	getAlertsFunc    func(context.Context, *string) ([]*models.Alert, error)
	getAlertByIDFunc func(context.Context, string) (*models.Alert, error)
	updateAlertFunc  func(context.Context, string, *models.UpdateAlertRequest) (*models.Alert, error)
	deleteAlertFunc  func(context.Context, string) error
}

func (m *MockAlertQuerier) CreateAlert(ctx context.Context, alert *models.Alert) error {
	if m.createAlertFunc != nil {
		return m.createAlertFunc(ctx, alert)
	}
	// Default: generate ID and return success
	alert.ID = uuid.New().String()
	return nil
}

func (m *MockAlertQuerier) GetAlerts(ctx context.Context, nodeID *string) ([]*models.Alert, error) {
	if m.getAlertsFunc != nil {
		return m.getAlertsFunc(ctx, nodeID)
	}
	return nil, nil
}

func (m *MockAlertQuerier) GetAlertByID(ctx context.Context, id string) (*models.Alert, error) {
	if m.getAlertByIDFunc != nil {
		return m.getAlertByIDFunc(ctx, id)
	}
	return nil, errors.New("alert not found")
}

func (m *MockAlertQuerier) UpdateAlert(ctx context.Context, id string, req *models.UpdateAlertRequest) (*models.Alert, error) {
	if m.updateAlertFunc != nil {
		return m.updateAlertFunc(ctx, id, req)
	}
	return nil, errors.New("alert not found")
}

func (m *MockAlertQuerier) DeleteAlert(ctx context.Context, id string) error {
	if m.deleteAlertFunc != nil {
		return m.deleteAlertFunc(ctx, id)
	}
	return errors.New("alert not found")
}

func setupAlertTestRouter(t *testing.T) (*gin.Engine, *MockAlertQuerier) {
	gin.SetMode(gin.TestMode)

	mockQuerier := &MockAlertQuerier{}
	handler := NewAlertHandler(mockQuerier)

	router := gin.New()
	router.POST("/api/v1/alerts/rules", handler.CreateAlertRuleHandler)
	router.GET("/api/v1/alerts/rules", handler.GetAlertRulesHandler)
	router.GET("/api/v1/alerts/rules/:id", handler.GetAlertRuleByIDHandler)
	router.PUT("/api/v1/alerts/rules/:id", handler.UpdateAlertRuleHandler)
	router.DELETE("/api/v1/alerts/rules/:id", handler.DeleteAlertRuleHandler)

	return router, mockQuerier
}

func TestCreateAlertRuleHandler(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	// Mock successful creation
	mock.createAlertFunc = func(ctx context.Context, alert *models.Alert) error {
		alert.ID = uuid.New().String()
		return nil
	}

	// Test creating a valid alert
	body := map[string]interface{}{
		"metric":    "latency",
		"threshold": 100.5,
		"level":     "P1",
		"enabled":   true,
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/v1/alerts/rules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.CreateAlertResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Data.Alert.ID)
	assert.Equal(t, "latency", response.Data.Alert.Metric)
	assert.Equal(t, 100.5, response.Data.Alert.Threshold)
	assert.Equal(t, "P1", response.Data.Alert.Level)
	assert.True(t, response.Data.Alert.Enabled)
}

func TestCreateAlertRuleHandlerValidation(t *testing.T) {
	router, _ := setupAlertTestRouter(t)

	tests := []struct {
		name       string
		body       map[string]interface{}
		expectCode int
	}{
		{
			name: "Invalid metric",
			body: map[string]interface{}{
				"metric":    "invalid_metric",
				"threshold": 100.0,
				"level":     "P1",
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Invalid level",
			body: map[string]interface{}{
				"metric":    "latency",
				"threshold": 100.0,
				"level":     "P4",
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Missing required field",
			body: map[string]interface{}{
				"metric": "latency",
				// missing threshold and level
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Invalid threshold (negative)",
			body: map[string]interface{}{
				"metric":    "latency",
				"threshold": -10.0,
				"level":     "P1",
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/v1/alerts/rules", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "ERR_VALIDATION", response["code"])
		})
	}
}

func TestCreateAlertRuleHandlerDefaultsEnabled(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	// Mock successful creation
	mock.createAlertFunc = func(ctx context.Context, alert *models.Alert) error {
		alert.ID = uuid.New().String()
		return nil
	}

	// Test that enabled defaults to true when not provided
	body := map[string]interface{}{
		"metric":    "latency",
		"threshold": 100.0,
		"level":     "P1",
		// enabled not provided
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/v1/alerts/rules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.CreateAlertResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Data.Alert.Enabled, "enabled should default to true")
}

func TestGetAlertRulesHandler(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	// Mock successful alerts retrieval
	mock.getAlertsFunc = func(ctx context.Context, nodeID *string) ([]*models.Alert, error) {
		alert1 := &models.Alert{
			ID:        uuid.New().String(),
			Metric:    "latency",
			Threshold: 100.0,
			Level:     "P1",
			NodeID:    nil,
			Enabled:   true,
		}
		alert2 := &models.Alert{
			ID:        uuid.New().String(),
			Metric:    "packet_loss_rate",
			Threshold: 5.0,
			Level:     "P0",
			NodeID:    nil,
			Enabled:   true,
		}
		return []*models.Alert{alert1, alert2}, nil
	}

	// Test getting all alerts
	req, _ := http.NewRequest("GET", "/api/v1/alerts/rules", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetAlertsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data.Alerts, 2)
}

func TestGetAlertRuleByIDHandler(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	alertID := uuid.New().String()

	// Mock successful alert retrieval
	mock.getAlertByIDFunc = func(ctx context.Context, id string) (*models.Alert, error) {
		if id == alertID {
			return &models.Alert{
				ID:        alertID,
				Metric:    "latency",
				Threshold: 100.0,
				Level:     "P1",
				NodeID:    nil,
				Enabled:   true,
			}, nil
		}
		return nil, errors.New("alert not found")
	}

	// Test getting by ID
	req, _ := http.NewRequest("GET", "/api/v1/alerts/rules/"+alertID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetAlertByIDResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, alertID, response.Data.Alert.ID)
	assert.Equal(t, "latency", response.Data.Alert.Metric)
}

func TestGetAlertRuleByIDHandlerNotFound(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	// Mock not found
	mock.getAlertByIDFunc = func(ctx context.Context, id string) (*models.Alert, error) {
		return nil, errors.New("alert not found")
	}

	req, _ := http.NewRequest("GET", "/api/v1/alerts/rules/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ERR_NOT_FOUND", response["code"])
}

func TestUpdateAlertRuleHandler(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	alertID := uuid.New().String()

	// Mock successful update
	mock.updateAlertFunc = func(ctx context.Context, id string, req *models.UpdateAlertRequest) (*models.Alert, error) {
		if id == alertID {
			threshold := 200.0
			if req.Threshold != nil {
				threshold = *req.Threshold
			}
			return &models.Alert{
				ID:        alertID,
				Metric:    "latency",
				Threshold: threshold,
				Level:     "P1",
				NodeID:    nil,
				Enabled:   true,
			}, nil
		}
		return nil, errors.New("alert not found")
	}

	// Update threshold
	newThreshold := 200.0
	body := map[string]interface{}{
		"threshold": newThreshold,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/v1/alerts/rules/"+alertID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UpdateAlertResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, newThreshold, response.Data.Alert.Threshold)
}

func TestUpdateAlertRuleHandlerNotFound(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	// Mock not found
	mock.updateAlertFunc = func(ctx context.Context, id string, req *models.UpdateAlertRequest) (*models.Alert, error) {
		return nil, errors.New("alert not found")
	}

	body := map[string]interface{}{
		"threshold": 200.0,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/v1/alerts/rules/non-existent-id", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ERR_NOT_FOUND", response["code"])
}

func TestDeleteAlertRuleHandler(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	alertID := uuid.New().String()

	// Mock successful deletion
	mock.deleteAlertFunc = func(ctx context.Context, id string) error {
		if id == alertID {
			return nil
		}
		return errors.New("alert not found")
	}

	// Delete the alert
	req, _ := http.NewRequest("DELETE", "/api/v1/alerts/rules/"+alertID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeleteAlertResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Message, "deleted successfully")
}

func TestDeleteAlertRuleHandlerNotFound(t *testing.T) {
	router, mock := setupAlertTestRouter(t)

	// Mock not found
	mock.deleteAlertFunc = func(ctx context.Context, id string) error {
		return errors.New("alert not found")
	}

	req, _ := http.NewRequest("DELETE", "/api/v1/alerts/rules/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ERR_NOT_FOUND", response["code"])
}

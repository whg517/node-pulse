package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/auth"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

func setupAlertRecordsAPITest(t *testing.T) (*gin.Engine, *pgxpool.Pool, *auth.SessionService, func()) {
	pool, cleanup := setupTestDB(t)

	// Create tables
	ctx := context.Background()
	err := db.Migrate(ctx, pool)
	require.NoError(t, err)

	// Create test session service
	sessionService := auth.NewSessionService(pool)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAlertRecordHandler(pool)

	// Create authenticated test route
	router.GET("/api/v1/alerts/records", func(c *gin.Context) {
		// Mock authenticated user
		c.Set("user_id", "test-user-id")
		c.Set("role", "admin")
		handler.GetAlertRecordsHandler(c)
	})

	router.PUT("/api/v1/alerts/records/:id/status", func(c *gin.Context) {
		// Mock authenticated user
		c.Set("user_id", "test-user-id")
		c.Set("role", "admin")
		handler.UpdateAlertRecordStatusHandler(c)
	})

	return router, pool, sessionService, cleanup
}

func TestGetAlertRecordsHandler(t *testing.T) {
	router, pool, _, cleanup := setupAlertRecordsAPITest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data
	node := &models.Node{
		ID:     "test-node-id",
		Name:   "Test Node",
		IP:     "192.168.1.1",
		Region: "us-east",
	}
	err := db.CreateNode(ctx, pool, node)
	require.NoError(t, err)

	// Create alert event
	alertEvent := &models.AlertEvent{
		ID:           "test-alert-event-id",
		NodeID:       node.ID,
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}
	alertEventsQuerier := db.NewAlertEventsQuerier(pool)
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	// Create alert record
	record := &models.AlertRecord{
		AlertEventID: alertEvent.ID,
		NodeID:       node.ID,
		Metric:       alertEvent.Metric,
		Level:        alertEvent.Level,
		Status:       "pending",
	}
	err = db.CreateAlertRecord(ctx, pool, record)
	require.NoError(t, err)

	// Test getting all records
	req, _ := http.NewRequest("GET", "/api/v1/alerts/records", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Alert records retrieved successfully", response["message"])
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 1, "Should return at least 1 record")

	// Test filtering by node_id
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?node_id="+node.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response2 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response2)
	require.NoError(t, err)

	data2 := response2["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data2), 1, "Should return records for node")

	// Test filtering by status
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?status=pending", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response3 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response3)
	require.NoError(t, err)

	data3 := response3["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data3), 1, "Should return pending records")

	// Test filtering by level
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?level=P0", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response4 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response4)
	require.NoError(t, err)

	data4 := response4["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data4), 1, "Should return P0 records")

	// Test invalid level
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?level=INVALID", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response5 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response5)
	require.NoError(t, err)

	assert.Equal(t, "ERR_INVALID_LEVEL", response5["code"])

	// Test invalid status
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?status=INVALID", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response6 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response6)
	require.NoError(t, err)

	assert.Equal(t, "ERR_INVALID_STATUS", response6["code"])

	// Test invalid start_time format
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?start_time=invalid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response7 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response7)
	require.NoError(t, err)

	assert.Equal(t, "ERR_INVALID_START_TIME", response7["code"])

	// Test valid time range filtering
	startTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?start_time="+startTime+"&end_time="+endTime, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test pagination
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?limit=1&offset=0", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test invalid limit
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?limit=invalid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response8 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response8)
	require.NoError(t, err)

	assert.Equal(t, "ERR_INVALID_LIMIT", response8["code"])
}

func TestUpdateAlertRecordStatusHandler(t *testing.T) {
	router, pool, _, cleanup := setupAlertRecordsAPITest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data
	node := &models.Node{
		ID:     "test-node-id-2",
		Name:   "Test Node 2",
		IP:     "192.168.1.2",
		Region: "us-east",
	}
	err := db.CreateNode(ctx, pool, node)
	require.NoError(t, err)

	alertEvent := &models.AlertEvent{
		ID:           "test-alert-event-id-2",
		NodeID:       node.ID,
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}
	alertEventsQuerier := db.NewAlertEventsQuerier(pool)
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	record := &models.AlertRecord{
		AlertEventID: alertEvent.ID,
		NodeID:       node.ID,
		Metric:       alertEvent.Metric,
		Level:        alertEvent.Level,
		Status:       "pending",
	}
	err = db.CreateAlertRecord(ctx, pool, record)
	require.NoError(t, err)

	// Test valid status update: pending -> in_progress
	reqBody := map[string]string{"status": "in_progress"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/records/"+record.ID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Alert record status updated successfully", response["message"])
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "in_progress", data["status"], "Status should be updated to in_progress")

	// Test invalid status
	reqBody = map[string]string{"status": "invalid_status"}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/v1/alerts/records/"+record.ID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response2 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response2)
	require.NoError(t, err)

	assert.Equal(t, "ERR_INVALID_STATUS", response2["code"])

	// Test missing status field
	reqBody = map[string]string{}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/v1/alerts/records/"+record.ID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test updating non-existent record
	reqBody = map[string]string{"status": "resolved"}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/v1/alerts/records/non-existent-id/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response3 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response3)
	require.NoError(t, err)

	assert.Equal(t, "ERR_RECORD_NOT_FOUND", response3["code"])

	// Test invalid status transition (resolved -> in_progress not allowed)
	// First update to resolved
	reqBody = map[string]string{"status": "resolved"}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/v1/alerts/records/"+record.ID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Now try to update back to in_progress (should fail)
	reqBody = map[string]string{"status": "in_progress"}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/v1/alerts/records/"+record.ID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response4 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response4)
	require.NoError(t, err)

	assert.Equal(t, "ERR_INVALID_STATUS_TRANSITION", response4["code"])
}

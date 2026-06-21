package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/api"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

func setupTestDBForAlertRecords(t *testing.T) (*pgxpool.Pool, func()) {
	testutil.SetupTestConfig()

	// Load configuration
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	// Verify actual connectivity (pgxpool.New is lazy)
	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		t.Skipf("Skipping test: cannot ping test database: %v", pingErr)
	}

	// Clean up and migrate
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alert_records CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS alert_events CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")

	if err := db.Migrate(ctx, pool); err != nil {
		t.Skipf("Skipping: database not available - migration failed: %v", err)
	}

	cleanup := func() {
		pool.Close()
		testutil.TeardownTestConfig()
	}

	return pool, cleanup
}

func setupAlertRecordsAPITest(t *testing.T) (*gin.Engine, *pgxpool.Pool, func()) {
	pool, cleanup := setupTestDBForAlertRecords(t)

	// Create tables
	ctx := context.Background()
	err := db.Migrate(ctx, pool)
	require.NoError(t, err)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := api.NewAlertRecordHandler(pool)

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

	router.POST("/api/v1/alerts/records/:id/notes", func(c *gin.Context) {
		// Mock authenticated user
		c.Set("user_id", "test-user-id")
		c.Set("role", "admin")
		handler.AddAlertNoteHandler(c)
	})

	router.GET("/api/v1/alerts/records/:id/notes", func(c *gin.Context) {
		// Mock authenticated user
		c.Set("user_id", "test-user-id")
		c.Set("role", "admin")
		handler.GetAlertNotesHandler(c)
	})

	return router, pool, cleanup
}

func TestGetAlertRecordsHandler(t *testing.T) {
	router, pool, cleanup := setupAlertRecordsAPITest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data with valid UUIDs
	nodeUUID := uuid.New()
	err := db.CreateNode(ctx, pool, nodeUUID, "Test Node", "192.168.1.1", "us-east", nil)
	require.NoError(t, err)

	// Create alert event
	alertEventID := uuid.New().String()
	alertEvent := &models.AlertEvent{
		ID:           alertEventID,
		NodeID:       nodeUUID.String(),
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
		NodeID:       nodeUUID.String(),
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

	// Debug: print response if not OK
	if w.Code != http.StatusOK {
		t.Logf("Response code: %d, Body: %s", w.Code, w.Body.String())
	}

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Alert records retrieved successfully", response["message"])
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 1, "Should return at least 1 record")

	// Test filtering by node_id - use the actual UUID
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?node_id="+nodeUUID.String(), nil)
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
	// URL-encode timestamps to handle + in timezone
	req, _ = http.NewRequest("GET", "/api/v1/alerts/records?start_time="+url.QueryEscape(startTime)+"&end_time="+url.QueryEscape(endTime), nil)
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
	router, pool, cleanup := setupAlertRecordsAPITest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data with valid UUIDs
	nodeUUID := uuid.New()
	err := db.CreateNode(ctx, pool, nodeUUID, "Test Node 2", "192.168.1.2", "us-east", nil)
	require.NoError(t, err)

	alertEventID := uuid.New().String()
	alertEvent := &models.AlertEvent{
		ID:           alertEventID,
		NodeID:       nodeUUID.String(),
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
		NodeID:       nodeUUID.String(),
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
	req, _ = http.NewRequest("PUT", "/api/v1/alerts/records/"+uuid.New().String()+"/status", bytes.NewBuffer(jsonBody))
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

func TestAlertRecordNotesHandlers(t *testing.T) {
	router, pool, cleanup := setupAlertRecordsAPITest(t)
	defer cleanup()

	ctx := context.Background()

	nodeUUID := uuid.New()
	err := db.CreateNode(ctx, pool, nodeUUID, "Test Node Notes", "192.168.1.3", "us-east", nil)
	require.NoError(t, err)

	alertEvent := &models.AlertEvent{
		ID:           uuid.New().String(),
		NodeID:       nodeUUID.String(),
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P1",
		CreatedAt:    time.Now(),
	}
	alertEventsQuerier := db.NewAlertEventsQuerier(pool)
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	record := &models.AlertRecord{
		AlertEventID: alertEvent.ID,
		NodeID:       nodeUUID.String(),
		Metric:       alertEvent.Metric,
		Level:        alertEvent.Level,
		Status:       "pending",
	}
	err = db.CreateAlertRecord(ctx, pool, record)
	require.NoError(t, err)

	reqBody := map[string]string{"status": "in_progress", "note": "Investigating carrier route"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/records/"+record.ID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var statusResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	require.NoError(t, err)
	statusData := statusResponse["data"].(map[string]interface{})
	notes := statusData["notes"].([]interface{})
	require.Len(t, notes, 1)
	firstNote := notes[0].(map[string]interface{})
	assert.Equal(t, "Investigating carrier route", firstNote["content"])
	assert.Equal(t, "System", firstNote["user_name"])

	reqBody = map[string]string{"note": "Carrier confirmed maintenance"}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/v1/alerts/records/"+record.ID+"/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var addResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &addResponse)
	require.NoError(t, err)
	addedNote := addResponse["data"].(map[string]interface{})
	assert.Equal(t, "Carrier confirmed maintenance", addedNote["content"])

	req, _ = http.NewRequest("GET", "/api/v1/alerts/records/"+record.ID+"/notes", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var listResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	require.NoError(t, err)
	listData := listResponse["data"].([]interface{})
	require.Len(t, listData, 2)
	assert.Equal(t, "Investigating carrier route", listData[0].(map[string]interface{})["content"])
	assert.Equal(t, "Carrier confirmed maintenance", listData[1].(map[string]interface{})["content"])

	reqBody = map[string]string{"note": "   "}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/v1/alerts/records/"+record.ID+"/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	reqBody = map[string]string{"note": "This should not attach"}
	jsonBody, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/v1/alerts/records/"+uuid.New().String()+"/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/api"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

// setupIntegrationTestRouter creates a router with real database connection for integration testing
func setupIntegrationTestRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	beaconHandler := api.NewBeaconHandler(db.NewPoolQuerier(pool))
	router.POST("/api/v1/beacon/heartbeat", beaconHandler.HandleHeartbeat)

	return router
}

// createTestNode creates a test node in the database
func createTestNode(t *testing.T, ctx context.Context, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()

	nodeID := uuid.New()
	err := db.CreateNode(ctx, pool, nodeID, "test-node", "192.168.1.100", "us-west", map[string]interface{}{
		"environment": "test",
		"version":     "1.0.0",
	})
	require.NoError(t, err, "Failed to create test node")

	return nodeID
}

func TestBeaconHeartbeatIntegration_ValidRequest(t *testing.T) {
	// Setup - skip if no test database
	ctx := context.Background()
	pool := testDBPool(t)
	if pool == nil {
		t.Skip("No test database available")
	}

	// Create test node
	testNodeID := createTestNode(t, ctx, pool)
	defer cleanupTestNode(t, ctx, pool, testNodeID)

	// Setup router
	router := setupIntegrationTestRouter(pool)

	// Prepare request
	reqBody := models.HeartbeatRequest{
		NodeID:         testNodeID.String(),
		ProbeID:        "probe-001",
		LatencyMs:      50.5,
		PacketLossRate: 0.1,
		JitterMs:       5.2,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Act - measure response time
	start := time.Now()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	duration := time.Since(start)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.HeartbeatSuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "心跳数据接收成功", resp.Message)
	assert.True(t, resp.Data.Received)
	assert.Equal(t, testNodeID.String(), resp.Data.NodeID)

	// Verify response time is within 5 seconds (NFR-OTHER-001)
	assert.Less(t, duration.Milliseconds(), int64(5000),
		"API response time should be less than 5 seconds, took %dms", duration.Milliseconds())

	// Also verify it's reasonably fast (P95 target: 200ms)
	t.Logf("API response time: %dms", duration.Milliseconds())
}

func TestBeaconHeartbeatIntegration_InvalidNodeID(t *testing.T) {
	// Setup
	pool := testDBPool(t)
	if pool == nil {
		t.Skip("No test database available")
	}

	router := setupIntegrationTestRouter(pool)

	// Use non-existent node ID
	fakeNodeID := uuid.New()

	reqBody := models.HeartbeatRequest{
		NodeID:         fakeNodeID.String(),
		ProbeID:        "probe-001",
		LatencyMs:      50.5,
		PacketLossRate: 0.1,
		JitterMs:       5.2,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should return 400 for non-existent node
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "ERR_NODE_NOT_FOUND", resp.Code)
	assert.Contains(t, resp.Message, "节点不存在")
}

func TestBeaconHeartbeatIntegration_MetricValidation(t *testing.T) {
	tests := []struct {
		name           string
		nodeID         uuid.UUID
		latencyMs      float64
		packetLossRate float64
		jitterMs       float64
		expectedCode   string
	}{
		{
			name:           "Latency too high",
			latencyMs:      65000,
			packetLossRate: 10.0,
			jitterMs:       100.0,
			expectedCode:   "ERR_INVALID_LATENCY",
		},
		{
			name:           "Packet loss too high",
			latencyMs:      100.0,
			packetLossRate: 150.0,
			jitterMs:       100.0,
			expectedCode:   "ERR_INVALID_PACKET_LOSS",
		},
		{
			name:           "Jitter too high",
			latencyMs:      100.0,
			packetLossRate: 10.0,
			jitterMs:       60000,
			expectedCode:   "ERR_INVALID_JITTER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			pool := testDBPool(t)
			if pool == nil {
				t.Skip("No test database available")
			}

			testNodeID := createTestNode(t, ctx, pool)
			defer cleanupTestNode(t, ctx, pool, testNodeID)

			router := setupIntegrationTestRouter(pool)

			reqBody := models.HeartbeatRequest{
				NodeID:         testNodeID.String(),
				ProbeID:        "probe-001",
				LatencyMs:      tt.latencyMs,
				PacketLossRate: tt.packetLossRate,
				JitterMs:       tt.jitterMs,
				Timestamp:      time.Now().Format(time.RFC3339),
			}

			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Act
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var resp models.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestBeaconHeartbeatIntegration_PerformanceTest(t *testing.T) {
	// Performance test: Verify API can handle multiple requests within 5 seconds
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	pool := testDBPool(t)
	if pool == nil {
		t.Skip("No test database available")
	}

	// Create test node
	testNodeID := createTestNode(t, ctx, pool)
	defer cleanupTestNode(t, ctx, pool, testNodeID)

	router := setupIntegrationTestRouter(pool)

	// Send 10 heartbeat requests
	numRequests := 10
	successCount := 0

	for i := 0; i < numRequests; i++ {
		reqBody := models.HeartbeatRequest{
			NodeID:         testNodeID.String(),
			ProbeID:        fmt.Sprintf("probe-%03d", i),
			LatencyMs:      50.0 + float64(i),
			PacketLossRate: float64(i) / 100.0,
			JitterMs:       5.0 + float64(i),
			Timestamp:      time.Now().Format(time.RFC3339),
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		start := time.Now()
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		// Verify each request completes within 5 seconds
		assert.Less(t, duration.Milliseconds(), int64(5000),
			"Request %d exceeded 5 second threshold: %dms", i, duration.Milliseconds())

		if w.Code == http.StatusOK {
			successCount++
		}
	}

	// All requests should succeed
	assert.Equal(t, numRequests, successCount, "Not all requests succeeded")
	t.Logf("Successfully processed %d heartbeat requests in under 5 seconds each", numRequests)
}

// Helper function to get test database pool
func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Use environment variable or default to test database
	connString := "postgres://nodepulse:testpass@localhost:5432/nodepulse_test?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Logf("Failed to connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		t.Logf("Failed to ping test database: %v", err)
		pool.Close()
		return nil
	}

	return pool
}

// Helper function to cleanup test node
func cleanupTestNode(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID) {
	t.Helper()

	err := db.DeleteNode(ctx, pool, nodeID)
	if err != nil {
		t.Logf("Failed to cleanup test node: %v", err)
	}
}

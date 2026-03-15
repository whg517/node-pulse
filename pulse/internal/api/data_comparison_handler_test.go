package api

import (
	"context"
	"encoding/json"
	"fmt"
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

	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

func TestGetComparisonHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	// Setup test database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Run migrations to ensure tables exist
	if err := db.Migrate(ctx, pool); err != nil {
		t.Skipf("Skipping: database not available - migration failed: %v", err)
	}

	// Create test nodes
	nodeIDs := createTestNodes(t, ctx, pool, 3)

	// Insert test metrics data
	insertTestMetrics(t, ctx, pool, nodeIDs)

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	startTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)

	// Build URL with repeated node_ids parameters for Gin []string binding
	// URL-encode timestamps to preserve + in timezone offset
	url := "/api/v1/data/comparison?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime) + "&metrics=latency"
	for _, nodeID := range []string{nodeIDs[0], nodeIDs[1]} {
		url += "&node_ids=" + nodeID
	}
	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response ComparisonResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate response structure
	assert.NotEmpty(t, response.Data.TimeRange.Start)
	assert.NotEmpty(t, response.Data.TimeRange.End)
	assert.NotEmpty(t, response.Data.Nodes)
	assert.NotEmpty(t, response.Data.Statistics)
	assert.NotEmpty(t, response.Message)
	assert.NotEmpty(t, response.Timestamp)

	// Validate nodes data
	assert.GreaterOrEqual(t, len(response.Data.Nodes), 2)

	for _, node := range response.Data.Nodes {
		assert.NotEmpty(t, node.NodeID)
		assert.NotEmpty(t, node.Name)
		assert.Contains(t, node.Metrics, "latency")

		latencyData := node.Metrics["latency"]
		assert.Greater(t, len(latencyData.DataPoints), 0)
		assert.GreaterOrEqual(t, latencyData.Avg, 0.0)
		assert.GreaterOrEqual(t, latencyData.Max, 0.0)
		assert.GreaterOrEqual(t, latencyData.Min, 0.0)
		assert.GreaterOrEqual(t, latencyData.Max, latencyData.Min)
	}

	// Validate statistics
	assert.Contains(t, response.Data.Statistics, "latency")
	latencyStats := response.Data.Statistics["latency"]
	assert.GreaterOrEqual(t, latencyStats.OverallAvg, 0.0)
	assert.GreaterOrEqual(t, latencyStats.OverallMax, 0.0)
	assert.GreaterOrEqual(t, latencyStats.OverallMin, 0.0)
	assert.NotEmpty(t, latencyStats.Differences)

	// Cleanup
	cleanupTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetComparisonHandler_MultipleMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Create test nodes
	nodeIDs := createTestNodes(t, ctx, pool, 2)

	// Insert test metrics data
	insertTestMetrics(t, ctx, pool, nodeIDs)

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request with multiple metrics
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	startTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)

	// Build URL with repeated node_ids and metrics parameters for Gin []string binding
	// URL-encode timestamps to preserve + in timezone offset
	url := "/api/v1/data/comparison?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime)
	for _, nodeID := range []string{nodeIDs[0], nodeIDs[1]} {
		url += "&node_ids=" + nodeID
	}
	for _, metric := range []string{"latency", "packet_loss_rate", "jitter"} {
		url += "&metrics=" + metric
	}
	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response ComparisonResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate that all metrics are present
	for _, node := range response.Data.Nodes {
		assert.Contains(t, node.Metrics, "latency")
		assert.Contains(t, node.Metrics, "packet_loss_rate")
		assert.Contains(t, node.Metrics, "jitter")
	}

	// Validate statistics for all metrics
	assert.Contains(t, response.Data.Statistics, "latency")
	assert.Contains(t, response.Data.Statistics, "packet_loss_rate")
	assert.Contains(t, response.Data.Statistics, "jitter")

	// Cleanup
	cleanupTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetComparisonHandler_MaxFiveNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Create test nodes
	nodeIDs := createTestNodes(t, ctx, pool, 5)

	// Insert test metrics data
	insertTestMetrics(t, ctx, pool, nodeIDs)

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request with 5 nodes (should succeed)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	startTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)

	// Build URL with repeated node_ids parameters for Gin []string binding
	// URL-encode timestamps to preserve + in timezone offset
	url := "/api/v1/data/comparison?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime) + "&metrics=latency"
	for _, nodeID := range nodeIDs {
		url += "&node_ids=" + nodeID
	}
	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response - should succeed
	assert.Equal(t, http.StatusOK, w.Code)

	// Cleanup
	cleanupTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetComparisonHandler_TooManyNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database (nil cache for error case testing)
	handler := NewDataHandler(nil, nil)

	// Create request with 6 node IDs (should fail validation)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	startTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)

	// Use repeated node_ids parameters for correct []string binding (6 nodes should fail max=5 validation)
	url := "/api/v1/data/comparison?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime) + "&metrics=latency"
	url += "&node_ids=node1&node_ids=node2&node_ids=node3&node_ids=node4&node_ids=node5&node_ids=node6"
	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response - should fail with validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetComparisonHandler_MinTwoNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database (nil cache for error case testing)
	handler := NewDataHandler(nil, nil)

	// Create request with only 1 node (should fail)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	startTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)

	req, _ := http.NewRequest("GET", "/api/v1/data/comparison?node_ids=node1&start_time="+url.QueryEscape(startTime)+"&end_time="+url.QueryEscape(endTime)+"&metrics=latency", nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response - should fail with validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetComparisonHandler_InvalidTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database (nil cache for error case testing)
	handler := NewDataHandler(nil, nil)

	// Create request with end_time before start_time
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	startTime := time.Now().Format(time.RFC3339)
	endTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

	// Use repeated node_ids parameters for correct []string binding
	url := "/api/v1/data/comparison?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime) + "&metrics=latency"
	url += "&node_ids=node1&node_ids=node2"
	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response - should fail
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["details"], "end_time must be after start_time")
}

func TestGetComparisonHandler_InvalidMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database (nil cache for error case testing)
	handler := NewDataHandler(nil, nil)

	// Create request with invalid metric
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	startTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)

	// Use repeated node_ids parameters for correct []string binding
	url := "/api/v1/data/comparison?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime) + "&metrics=invalid_metric"
	url += "&node_ids=node1&node_ids=node2"
	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response - should fail
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["details"], "not valid")
}

func TestGetComparisonHandler_InvalidStartTimeFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database (nil cache for error case testing)
	handler := NewDataHandler(nil, nil)

	// Create request with invalid start_time format
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Use repeated node_ids parameters for correct []string binding
	url := "/api/v1/data/comparison?start_time=invalid&end_time=" + url.QueryEscape("2024-01-01T00:00:00Z") + "&metrics=latency"
	url += "&node_ids=node1&node_ids=node2"
	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	// Execute handler
	handler.GetComparisonHandler(c)

	// Check response - should fail
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["details"], "ISO 8601 format")
}

func TestGetComparisonHandler_MissingRequiredParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database (nil cache for error case testing)
	handler := NewDataHandler(nil, nil)

	tests := []struct {
		name       string
		url        string
		errorField string
	}{
		{
			name:       "Missing node_ids",
			url:        "/api/v1/data/comparison?start_time=2024-01-01T00:00:00Z&end_time=2024-01-01T01:00:00Z&metrics=latency",
			errorField: "node_ids",
		},
		{
			name:       "Missing start_time",
			url:        "/api/v1/data/comparison?node_ids=node1&node_ids=node2&end_time=2024-01-01T01:00:00Z&metrics=latency",
			errorField: "start_time",
		},
		{
			name:       "Missing end_time",
			url:        "/api/v1/data/comparison?node_ids=node1&node_ids=node2&start_time=2024-01-01T00:00:00Z&metrics=latency",
			errorField: "end_time",
		},
		{
			name:       "Missing metrics",
			url:        "/api/v1/data/comparison?node_ids=node1&node_ids=node2&start_time=2024-01-01T00:00:00Z&end_time=2024-01-01T01:00:00Z",
			errorField: "metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest("GET", tt.url, nil)
			c.Request = req

			// Execute handler
			handler.GetComparisonHandler(c)

			// Check response - should fail
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestCalculateStatistics(t *testing.T) {
	tests := []struct {
		name        string
		dataPoints  []DataPoint
		expectedAvg float64
		expectedMax float64
		expectedMin float64
	}{
		{
			name: "Normal data",
			dataPoints: []DataPoint{
				{Timestamp: "2024-01-01T00:00:00Z", Value: 10.0},
				{Timestamp: "2024-01-01T00:01:00Z", Value: 20.0},
				{Timestamp: "2024-01-01T00:02:00Z", Value: 30.0},
			},
			expectedAvg: 20.0,
			expectedMax: 30.0,
			expectedMin: 10.0,
		},
		{
			name: "Single data point",
			dataPoints: []DataPoint{
				{Timestamp: "2024-01-01T00:00:00Z", Value: 15.0},
			},
			expectedAvg: 15.0,
			expectedMax: 15.0,
			expectedMin: 15.0,
		},
		{
			name:        "Empty data",
			dataPoints:  []DataPoint{},
			expectedAvg: 0.0,
			expectedMax: 0.0,
			expectedMin: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, max, min := calculateStatistics(tt.dataPoints)
			assert.Equal(t, tt.expectedAvg, avg)
			assert.Equal(t, tt.expectedMax, max)
			assert.Equal(t, tt.expectedMin, min)
		})
	}
}

func TestFindOverlapTimeRange(t *testing.T) {
	tests := []struct {
		name      string
		nodesData []ComparisonNodeData
		wantStart string
		wantEnd   string
	}{
		{
			name: "Full overlap",
			nodesData: []ComparisonNodeData{
				{
					NodeID: "node1",
					Metrics: map[string]ComparisonMetricData{
						"latency": {
							DataPoints: []DataPoint{
								{Timestamp: "2024-01-01T00:00:00Z", Value: 10.0},
								{Timestamp: "2024-01-01T01:00:00Z", Value: 20.0},
							},
						},
					},
				},
				{
					NodeID: "node2",
					Metrics: map[string]ComparisonMetricData{
						"latency": {
							DataPoints: []DataPoint{
								{Timestamp: "2024-01-01T00:30:00Z", Value: 15.0},
								{Timestamp: "2024-01-01T01:30:00Z", Value: 25.0},
							},
						},
					},
				},
			},
			wantStart: "2024-01-01T00:30:00Z",
			wantEnd:   "2024-01-01T01:00:00Z",
		},
		{
			name: "No overlap",
			nodesData: []ComparisonNodeData{
				{
					NodeID: "node1",
					Metrics: map[string]ComparisonMetricData{
						"latency": {
							DataPoints: []DataPoint{
								{Timestamp: "2024-01-01T00:00:00Z", Value: 10.0},
								{Timestamp: "2024-01-01T01:00:00Z", Value: 20.0},
							},
						},
					},
				},
				{
					NodeID: "node2",
					Metrics: map[string]ComparisonMetricData{
						"latency": {
							DataPoints: []DataPoint{
								{Timestamp: "2024-01-01T02:00:00Z", Value: 30.0},
								{Timestamp: "2024-01-01T03:00:00Z", Value: 40.0},
							},
						},
					},
				},
			},
			wantStart: "2024-01-01T02:00:00Z",
			wantEnd:   "2024-01-01T02:00:00Z", // Zero duration
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := findOverlapTimeRange(tt.nodesData)
			assert.Equal(t, tt.wantStart, gotStart.Format(time.RFC3339))
			assert.Equal(t, tt.wantEnd, gotEnd.Format(time.RFC3339))
		})
	}
}

// Helper functions

func createTestNodes(t *testing.T, ctx context.Context, pool *pgxpool.Pool, count int) []string {
	t.Helper()

	// Verify DB connectivity before using it
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping: database not available: %v", err)
		return nil
	}

	nodeIDs := make([]string, count)

	for i := 0; i < count; i++ {
		nodeID := uuid.New()
		nodeName := fmt.Sprintf("test-node-%d", i)

		query := `
			INSERT INTO nodes (id, name, ip, region, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`
		_, err := pool.Exec(ctx, query, nodeID, nodeName, "192.168.1."+fmt.Sprintf("%d", 100+i), "us-east", "online")
		require.NoError(t, err)
		nodeIDs[i] = nodeID.String()
	}

	return nodeIDs
}

func insertTestMetrics(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeIDs []string) {
	t.Helper()

	now := time.Now()
	baseTime := now.Add(-2 * time.Hour)

	for _, nodeID := range nodeIDs {
		// Create a probe for this node first (required by foreign key constraint)
		probeID := uuid.New()
		nodeUUID, err := uuid.Parse(nodeID)
		require.NoError(t, err)

		probeQuery := `
			INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds)
			VALUES ($1, $2, 'TCP', 'example.com', 443, 60, 10, 5)
		`
		_, err = pool.Exec(ctx, probeQuery, probeID, nodeUUID)
		require.NoError(t, err)

		// Now insert metrics with the probe_id
		for i := 0; i < 60; i++ { // 60 data points over 2 hours
			timestamp := baseTime.Add(time.Duration(i) * 2 * time.Minute)

			query := `
				INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms, is_aggregated, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, false, NOW())
			`

			latency := 50.0 + float64(i%10)*5.0
			packetLoss := 0.01 + float64(i%5)*0.005
			jitter := 2.0 + float64(i%3)*1.0

			_, err := pool.Exec(ctx, query, nodeUUID, probeID, timestamp, latency, packetLoss, jitter)
			require.NoError(t, err)
		}
	}
}

func cleanupTestNodes(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeIDs []string) {
	t.Helper()

	// Delete metrics first (foreign key constraint)
	for _, nodeID := range nodeIDs {
		_, err := pool.Exec(ctx, "DELETE FROM metrics WHERE node_id = $1", nodeID)
		require.NoError(t, err)
	}

	// Delete nodes
	for _, nodeID := range nodeIDs {
		_, err := pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", nodeID)
		require.NoError(t, err)
	}
}

package api

import (
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

	"github.com/kevin/node-pulse/pulse-api/internal/diagnostic"
	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
)

func TestGetDiagnosisHandler_Success_NodeLocalFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Create test nodes in different regions
	nodeIDs := createTestNodesForDiagnosis(t, ctx, pool)

	// Create a test probe for the metrics (use first node)
	probeID := createTestProbe(t, ctx, pool, nodeIDs[0])

	// Insert test metrics to simulate node local failure
	// node1 in us-east has high latency, others normal
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, probeID, "node_local_failure")

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Build query with array format: ?node_ids=id1&node_ids=id2&node_ids=id3&node_ids=id4
	req, _ := http.NewRequest("GET",
		"/api/v1/data/diagnosis?node_ids="+nodeIDs[0]+"&node_ids="+nodeIDs[1]+"&node_ids="+nodeIDs[2]+"&node_ids="+nodeIDs[3],
		nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response
	if w.Code != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResp)
		t.Logf("Error response: %+v", errorResp)
	}

	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosisResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate response structure
	assert.NotEmpty(t, response.Data.ProblemType)
	assert.NotEmpty(t, response.Data.Confidence)
	assert.NotEmpty(t, response.Data.Analysis.NodesAnalyzed)
	assert.NotEmpty(t, response.Data.Analysis.RegionsAnalyzed)
	assert.NotEmpty(t, response.Data.Recommendation)
	assert.NotEmpty(t, response.Timestamp)

	// Verify diagnosis detected node local failure
	assert.Equal(t, diagnostic.ProblemTypeNodeLocalFailure, response.Data.ProblemType)
	assert.Equal(t, diagnostic.ConfidenceHigh, response.Data.Confidence)

	// Cleanup
	cleanupDiagnosisTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetDiagnosisHandler_Success_CrossBorderLink(t *testing.T) {
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
	nodeIDs := createTestNodesForDiagnosis(t, ctx, pool)

	// Create a test probe for the metrics (use first node)
	probeID := createTestProbe(t, ctx, pool, nodeIDs[0])

	// Insert test metrics to simulate cross-border link issue
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, probeID, "cross_border_link")

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Build query with array format: ?node_ids=id1&node_ids=id2&node_ids=id3&node_ids=id4
	req, _ := http.NewRequest("GET",
		"/api/v1/data/diagnosis?node_ids="+nodeIDs[0]+"&node_ids="+nodeIDs[1]+"&node_ids="+nodeIDs[2]+"&node_ids="+nodeIDs[3],
		nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response
	if w.Code != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResp)
		t.Logf("Error response: %+v", errorResp)
	}

	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosisResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify diagnosis detected cross-border link issue
	assert.Equal(t, diagnostic.ProblemTypeCrossBorderLink, response.Data.ProblemType)
	// Confidence should be at least medium (with variance-based calculation, medium is realistic)
	assert.True(t, response.Data.Confidence == diagnostic.ConfidenceHigh || response.Data.Confidence == diagnostic.ConfidenceMedium)

	// Cleanup
	cleanupDiagnosisTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetDiagnosisHandler_Success_ISPRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Create test nodes (6 nodes for ISP pattern)
	nodeIDs := createTestNodesForDiagnosis(t, ctx, pool)

	// Create a test probe for the metrics (use first node)
	probeID := createTestProbe(t, ctx, pool, nodeIDs[0])

	// Insert test metrics to simulate ISP routing issue
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, probeID, "isp_routing")

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Build query with all 6 nodes
	req, _ := http.NewRequest("GET",
		"/api/v1/data/diagnosis?node_ids="+nodeIDs[0]+"&node_ids="+nodeIDs[1]+"&node_ids="+nodeIDs[2]+"&node_ids="+nodeIDs[3]+"&node_ids="+nodeIDs[4]+"&node_ids="+nodeIDs[5],
		nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response
	if w.Code != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResp)
		t.Logf("Error response: %+v", errorResp)
	}

	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosisResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify diagnosis detected ISP routing issue
	// Note: ISP routing is detected when one ISP shows issues while others are normal
	assert.Equal(t, diagnostic.ProblemTypeISPRouting, response.Data.ProblemType)

	// Cleanup
	cleanupDiagnosisTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetDiagnosisHandler_MinThreeNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Create only 2 test nodes (insufficient for diagnosis)
	allNodeIDs := createTestNodesForDiagnosis(t, ctx, pool)
	nodeIDs := allNodeIDs[:2]

	// Create a test probe for the metrics (use first node)
	probeID := createTestProbe(t, ctx, pool, nodeIDs[0])

	// Insert test metrics
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, probeID, "node_local_failure")

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Build query with array format: ?node_ids=id1&node_ids=id2
	req, _ := http.NewRequest("GET",
		"/api/v1/data/diagnosis?node_ids="+nodeIDs[0]+"&node_ids="+nodeIDs[1],
		nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response - should fail with insufficient nodes error (Gin validation)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// New format includes code, message, details
	assert.Contains(t, response["code"], "ERR_VALIDATION")
	assert.Contains(t, response["message"], "Invalid query parameters")

	// Cleanup
	cleanupDiagnosisTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetDiagnosisHandler_MissingNodeIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database (nil cache for error case testing)
	handler := NewDataHandler(nil, nil)

	// Create request without node_ids
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/api/v1/data/diagnosis", nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response - should fail validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDiagnosisHandler_NoDataFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Create test nodes but don't insert any metrics
	nodeIDs := createTestNodesForDiagnosis(t, ctx, pool)

	// Create handler with nil cache (tests use PostgreSQL only)
	handler := NewDataHandler(pool, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Build query with array format: ?node_ids=id1&node_ids=id2&node_ids=id3
	req, _ := http.NewRequest("GET",
		"/api/v1/data/diagnosis?node_ids="+nodeIDs[0]+"&node_ids="+nodeIDs[1]+"&node_ids="+nodeIDs[2],
		nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response - should fail with insufficient data
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// New format includes code, message, details
	assert.Contains(t, response["code"], "ERR_INSUFFICIENT_DATA")
	assert.Contains(t, response["message"], "Insufficient data for diagnosis")

	// Cleanup
	cleanupDiagnosisTestNodes(t, ctx, pool, nodeIDs)
}

// Helper functions

func createTestProbe(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeID string) string {
	t.Helper()

	probeID := uuid.New()
	query := `
		INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err := pool.Exec(ctx, query, probeID, nodeID, "TCP", "example.com", 443, 60, 5, 10)
	require.NoError(t, err)
	return probeID.String()
}

func createTestNodesForDiagnosis(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []string {
	t.Helper()

	// Create 6 nodes in different regions with ISP tags
	// ISP routing test needs same ISP in multiple regions
	nodeConfigs := []struct {
		name   string
		ip     string
		region string
		isp    string
	}{
		{"test-node-1", "192.168.1.101", "us-east", "ISP-A"},
		{"test-node-2", "192.168.1.102", "us-east", "ISP-A"},
		{"test-node-3", "192.168.1.103", "us-east", "ISP-B"},
		{"test-node-4", "192.168.1.104", "eu-west", "ISP-A"},
		{"test-node-5", "192.168.1.105", "eu-west", "ISP-A"},
		{"test-node-6", "192.168.1.106", "eu-west", "ISP-B"},
	}

	nodeIDs := make([]string, len(nodeConfigs))

	for i, config := range nodeConfigs {
		nodeID := uuid.New()

		query := `
			INSERT INTO nodes (id, name, ip, region, status, tags, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		`
		// Add ISP tag as JSONB
		tags := fmt.Sprintf(`{"isp": "%s"}`, config.isp)
		_, err := pool.Exec(ctx, query, nodeID, config.name, config.ip, config.region, "online", tags)
		require.NoError(t, err)
		nodeIDs[i] = nodeID.String()
	}

	return nodeIDs
}

func insertDiagnosisTestMetrics(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeIDs []string, probeID string, scenario string) {
	t.Helper()

	now := time.Now()
	baseTime := now.Add(-1 * time.Hour)

	var metricConfigs []struct {
		nodeID        string
		latency       float64
		packetLoss    float64
		jitter        float64
	}

	switch scenario {
	case "node_local_failure":
		// node1: high latency, rest: normal
		metricConfigs = []struct {
			nodeID     string
			latency    float64
			packetLoss float64
			jitter     float64
		}{
			{nodeIDs[0], 200.0, 0.05, 5.0}, // High latency
		}
		// Add remaining nodes with normal metrics
		for i := 1; i < len(nodeIDs); i++ {
			metricConfigs = append(metricConfigs, struct {
				nodeID     string
				latency    float64
				packetLoss float64
				jitter     float64
			}{
				nodeIDs[i], 50.0 + float64(i)*2.0, 0.01, 2.0 + float64(i)*0.5,
			})
		}

	case "cross_border_link":
		// Cross-border link: us-east region (ISP-A nodes) normal, eu-west region (ISP-A nodes) high latency
		// This uses only ISP-A nodes to avoid ISP pattern detection
		metricConfigs = []struct {
			nodeID     string
			latency    float64
			packetLoss float64
			jitter     float64
		}{
			{nodeIDs[0], 50.0, 0.01, 2.0},   // us-east, ISP-A: Normal
			{nodeIDs[1], 55.0, 0.01, 2.5},   // us-east, ISP-A: Normal
			{nodeIDs[3], 180.0, 0.05, 8.0},  // eu-west, ISP-A: High latency
			{nodeIDs[4], 190.0, 0.06, 9.0},  // eu-west, ISP-A: High latency
		}

	case "isp_routing":
		// ISP routing scenario: ISP-A nodes show issues across regions, ISP-B nodes normal
		// This simulates an ISP-specific routing problem
		metricConfigs = []struct {
			nodeID     string
			latency    float64
			packetLoss float64
			jitter     float64
		}{
			{nodeIDs[0], 150.0, 0.04, 6.0}, // us-east, ISP-A: high latency
			{nodeIDs[1], 160.0, 0.05, 7.0}, // us-east, ISP-A: high latency
			{nodeIDs[2], 50.0, 0.01, 2.0},  // us-east, ISP-B: normal
			{nodeIDs[3], 155.0, 0.045, 6.5}, // eu-west, ISP-A: high latency
			{nodeIDs[4], 165.0, 0.055, 7.5}, // eu-west, ISP-A: high latency
			{nodeIDs[5], 55.0, 0.01, 2.5},  // eu-west, ISP-B: normal
		}
	}

	// Insert metrics for each node
	for _, config := range metricConfigs {
		for i := 0; i < 30; i++ { // 30 data points over 1 hour
			timestamp := baseTime.Add(time.Duration(i) * 2 * time.Minute)

			// Add some variance to make it realistic
			latencyVariance := float64(i%5) * 5.0
			packetLossVariance := float64(i%3) * 0.005
			jitterVariance := float64(i%2) * 1.0

			query := `
				INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms, is_aggregated, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, false, NOW())
			`

			_, err := pool.Exec(ctx, query,
				config.nodeID,
				probeID,
				timestamp,
				config.latency+latencyVariance,
				config.packetLoss+packetLossVariance,
				config.jitter+jitterVariance,
			)
			require.NoError(t, err)
		}
	}
}

func cleanupDiagnosisTestNodes(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeIDs []string) {
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

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	// Insert test metrics to simulate node local failure
	// node1 in us-east has high latency, others normal
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, "node_local_failure")

	// Create handler
	handler := NewDataHandler(pool)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	nodeIDsParam := nodeIDs[0] + "," + nodeIDs[1] + "," + nodeIDs[2] + "," + nodeIDs[3]
	req, _ := http.NewRequest("GET", "/api/v1/data/diagnosis?node_ids="+nodeIDsParam, nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response
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
	assert.Equal(t, "node_local_failure", response.Data.ProblemType)
	assert.Equal(t, "high", response.Data.Confidence)

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

	// Insert test metrics to simulate cross-border link issue
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, "cross_border_link")

	// Create handler
	handler := NewDataHandler(pool)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	nodeIDsParam := nodeIDs[0] + "," + nodeIDs[1] + "," + nodeIDs[2] + "," + nodeIDs[3]
	req, _ := http.NewRequest("GET", "/api/v1/data/diagnosis?node_ids="+nodeIDsParam, nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosisResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify diagnosis detected cross-border link issue
	assert.Equal(t, "cross_border_link", response.Data.ProblemType)
	assert.Equal(t, "high", response.Data.Confidence)

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

	// Create test nodes
	nodeIDs := createTestNodesForDiagnosis(t, ctx, pool)

	// Insert test metrics to simulate ISP routing issue
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, "isp_routing")

	// Create handler
	handler := NewDataHandler(pool)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	nodeIDsParam := nodeIDs[0] + "," + nodeIDs[1] + "," + nodeIDs[2] + "," + nodeIDs[3]
	req, _ := http.NewRequest("GET", "/api/v1/data/diagnosis?node_ids="+nodeIDsParam, nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosisResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify diagnosis detected ISP routing issue
	assert.Contains(t, response.Data.ProblemType, "isp_routing")

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
	nodeIDs := createTestNodesForDiagnosis(t, ctx, pool)[:2]

	// Insert test metrics
	insertDiagnosisTestMetrics(t, ctx, pool, nodeIDs, "node_local_failure")

	// Create handler
	handler := NewDataHandler(pool)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	nodeIDsParam := nodeIDs[0] + "," + nodeIDs[1]
	req, _ := http.NewRequest("GET", "/api/v1/data/diagnosis?node_ids="+nodeIDsParam, nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response - should fail with insufficient nodes error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Insufficient data for diagnosis")

	// Cleanup
	cleanupDiagnosisTestNodes(t, ctx, pool, nodeIDs)
}

func TestGetDiagnosisHandler_MissingNodeIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler without database
	handler := NewDataHandler(nil)

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

	// Create handler
	handler := NewDataHandler(pool)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	nodeIDsParam := nodeIDs[0] + "," + nodeIDs[1] + "," + nodeIDs[2]
	req, _ := http.NewRequest("GET", "/api/v1/data/diagnosis?node_ids="+nodeIDsParam, nil)
	c.Request = req

	// Execute handler
	handler.GetDiagnosisHandler(c)

	// Check response - should fail with insufficient data
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Insufficient data for diagnosis")

	// Cleanup
	cleanupDiagnosisTestNodes(t, ctx, pool, nodeIDs)
}

// Helper functions

func createTestNodesForDiagnosis(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []string {
	t.Helper()

	// Create 4 nodes in different regions
	nodeConfigs := []struct {
		name   string
		ip     string
		region string
	}{
		{"test-node-1", "192.168.1.101", "us-east"},
		{"test-node-2", "192.168.1.102", "us-east"},
		{"test-node-3", "192.168.1.103", "eu-west"},
		{"test-node-4", "192.168.1.104", "eu-west"},
	}

	nodeIDs := make([]string, len(nodeConfigs))

	for i, config := range nodeConfigs {
		nodeID := uuid.New()

		query := `
			INSERT INTO nodes (id, name, ip, region, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`
		_, err := pool.Exec(ctx, query, nodeID, config.name, config.ip, config.region, "online")
		require.NoError(t, err)
		nodeIDs[i] = nodeID.String()
	}

	return nodeIDs
}

func insertDiagnosisTestMetrics(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nodeIDs []string, scenario string) {
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
		// node1: high latency, node2-4: normal
		metricConfigs = []struct {
			nodeID     string
			latency    float64
			packetLoss float64
			jitter     float64
		}{
			{nodeIDs[0], 200.0, 0.05, 5.0},  // High latency
			{nodeIDs[1], 50.0, 0.01, 2.0},   // Normal
			{nodeIDs[2], 52.0, 0.01, 2.0},   // Normal
			{nodeIDs[3], 55.0, 0.01, 2.5},   // Normal
		}

	case "cross_border_link":
		// us-east nodes: normal, eu-west nodes: high latency
		metricConfigs = []struct {
			nodeID     string
			latency    float64
			packetLoss float64
			jitter     float64
		}{
			{nodeIDs[0], 50.0, 0.01, 2.0},   // Normal
			{nodeIDs[1], 55.0, 0.01, 2.5},   // Normal
			{nodeIDs[2], 180.0, 0.05, 8.0},  // High latency
			{nodeIDs[3], 190.0, 0.06, 9.0},  // High latency
		}

	case "isp_routing":
		// Multiple regions with similar high latency
		metricConfigs = []struct {
			nodeID     string
			latency    float64
			packetLoss float64
			jitter     float64
		}{
			{nodeIDs[0], 150.0, 0.04, 6.0},  // High latency
			{nodeIDs[1], 160.0, 0.05, 7.0},  // High latency
			{nodeIDs[2], 155.0, 0.045, 6.5}, // High latency
			{nodeIDs[3], 165.0, 0.055, 7.5}, // High latency
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
				VALUES ($1, gen_random_uuid(), $2, $3, $4, $5, false, NOW())
			`

			_, err := pool.Exec(ctx, query,
				config.nodeID,
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

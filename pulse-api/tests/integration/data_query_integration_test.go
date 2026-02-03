package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/auth"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

// TestDataQueryEndpoints_Integration tests the data query API endpoints
// This includes GET /api/v1/data/metrics, /history, /comparison, and /diagnosis
func TestDataQueryEndpoints_Integration(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	if router == nil {
		return
	}
	defer pool.Close()

	// Create test user and login
	username := fmt.Sprintf("dataquery_%s", uuid.New().String()[:8])
	password := "testpass123"

	// Clean up any existing test user
	pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)

	// Create test user
	userID := uuid.New()
	hashedPassword, _ := auth.HashPassword(password)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at) VALUES ($1, $2, $3, $4, 0, NULL, NOW(), NOW())",
		userID, username, hashedPassword, "admin",
	)
	require.NoError(t, err)

	// Login to get session cookie
	loginReq := models.LoginRequest{
		Username: username,
		Password: password,
	}
	loginReqBody, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginReqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Extract session_id cookie
	cookies := w.Result().Cookies()
	var sessionID string
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionID = cookie.Value
			break
		}
	}
	require.NotEmpty(t, sessionID, "Failed to get session_id cookie")

	// Create test nodes
	now := time.Now()
	node1ID := uuid.New()
	node2ID := uuid.New()
	node3ID := uuid.New()

	// Insert test nodes
	for _, nodeID := range []uuid.UUID{node1ID, node2ID, node3ID} {
		nodeName := fmt.Sprintf("test-node-%s", nodeID.String()[:8])
		_, err = pool.Exec(context.Background(),
			"INSERT INTO nodes (id, name, ip, region, tags, last_heartbeat, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
			nodeID, nodeName, "192.168.1.100", "us-east", "{}", now.Add(-2*time.Minute), "online",
		)
		require.NoError(t, err)
	}

	// Insert test metrics data
	baseTime := now.Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)

		// Insert metrics for each node
		for _, nodeID := range []uuid.UUID{node1ID, node2ID, node3ID} {
			latency := 50.0 + float64(i%50)
			packetLoss := 0.01 + float64(i%10)*0.001
			jitter := 2.0 + float64(i%10)*0.5

			_, err = pool.Exec(context.Background(),
				`INSERT INTO metrics (node_id, timestamp, latency_ms, packet_loss_rate, jitter_ms)
				VALUES ($1, $2, $3, $4, $5)`,
				nodeID, timestamp, latency, packetLoss, jitter,
			)
			require.NoError(t, err)
		}
	}

	// Test 1: Get real-time metrics
	t.Run("get_realtime_metrics", func(t *testing.T) {
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/metrics?node_id=%s&node_id=%s", node1ID.String(), node2ID.String())
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data, ok := resp["data"].([]interface{})
		assert.True(t, ok, "Expected data array in response")
		assert.Greater(t, len(data), 0, "Expected metrics data")
	})

	// Test 2: Get real-time metrics without node_id (should fail)
	t.Run("get_metrics_missing_node_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/data/metrics", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errorMsg, ok := resp["error"].(string)
		assert.True(t, ok, "Expected error in response")
		assert.Contains(t, errorMsg, "node_id is required")
	})

	// Test 3: Get historical data
	t.Run("get_historical_data", func(t *testing.T) {
		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/history?node_id=%s&node_id=%s&start_time=%s&end_time=%s&metric=latency&metric=jitter",
			node1ID.String(), node2ID.String(), startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data, ok := resp["data"].([]interface{})
		assert.True(t, ok, "Expected data array in response")
		assert.Greater(t, len(data), 0, "Expected historical data")

		aggregation, ok := resp["aggregation"].(string)
		assert.True(t, ok, "Expected aggregation in response")
		assert.Equal(t, "1m", aggregation, "Expected default aggregation")
	})

	// Test 4: Get historical data with invalid time range
	t.Run("get_historical_data_invalid_time_range", func(t *testing.T) {
		startTime := now.Format(time.RFC3339)
		endTime := now.Add(-1 * time.Hour).Format(time.RFC3339)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/history?node_id=%s&start_time=%s&end_time=%s&metric=latency",
			node1ID.String(), startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errorMsg, ok := resp["error"].(string)
		assert.True(t, ok, "Expected error in response")
		assert.Contains(t, errorMsg, "Invalid time range")
	})

	// Test 5: Get historical data with invalid metric
	t.Run("get_historical_data_invalid_metric", func(t *testing.T) {
		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/history?node_id=%s&start_time=%s&end_time=%s&metric=invalid_metric",
			node1ID.String(), startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errorMsg, ok := resp["error"].(string)
		assert.True(t, ok, "Expected error in response")
		assert.Contains(t, errorMsg, "Invalid metric")
	})

	// Test 6: Get historical data with invalid aggregation
	t.Run("get_historical_data_invalid_aggregation", func(t *testing.T) {
		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/history?node_id=%s&start_time=%s&end_time=%s&metric=latency&aggregation=10m",
			node1ID.String(), startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errorMsg, ok := resp["error"].(string)
		assert.True(t, ok, "Expected error in response")
		assert.Contains(t, errorMsg, "Invalid aggregation")
	})

	// Test 7: Get comparison data
	t.Run("get_comparison_data", func(t *testing.T) {
		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		nodeIDs := fmt.Sprintf("%s,%s,%s", node1ID.String(), node2ID.String(), node3ID.String())
		url := fmt.Sprintf("/api/v1/data/comparison?node_ids=%s&start_time=%s&end_time=%s&metrics=latency&metrics=jitter",
			nodeIDs, startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok, "Expected data object in response")

		nodes, ok := data["nodes"].([]interface{})
		assert.True(t, ok, "Expected nodes array in data")
		assert.Greater(t, len(nodes), 0, "Expected nodes in comparison")

		stats, ok := data["statistics"].(map[string]interface{})
		assert.True(t, ok, "Expected statistics in data")
		assert.Greater(t, len(stats), 0, "Expected statistics for metrics")
	})

	// Test 8: Get comparison with insufficient nodes (should fail)
	t.Run("get_comparison_insufficient_nodes", func(t *testing.T) {
		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/comparison?node_ids=%s&start_time=%s&end_time=%s&metrics=latency",
			node1ID.String(), startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Should get validation error for min=2 nodes
		assert.Contains(t, w.Body.String(), "Invalid query parameters")
	})

	// Test 9: Get diagnosis data
	t.Run("get_diagnosis_data", func(t *testing.T) {
		w := httptest.NewRecorder()
		nodeIDs := fmt.Sprintf("%s,%s,%s", node1ID.String(), node2ID.String(), node3ID.String())
		url := fmt.Sprintf("/api/v1/data/diagnosis?node_ids=%s", nodeIDs)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok, "Expected data object in response")

		// Check for diagnosis result fields
		problemType, ok := data["problem_type"].(string)
		assert.True(t, ok, "Expected problem_type in data")
		assert.NotEmpty(t, problemType, "Expected problem type to be set")

		confidence, ok := data["confidence"].(float64)
		assert.True(t, ok, "Expected confidence in data")
		assert.GreaterOrEqual(t, confidence, 0.0, "Expected confidence >= 0")
		assert.LessOrEqual(t, confidence, 1.0, "Expected confidence <= 1")
	})

	// Test 10: Get diagnosis with insufficient nodes (should fail)
	t.Run("get_diagnosis_insufficient_nodes", func(t *testing.T) {
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/diagnosis?node_ids=%s,%s", node1ID.String(), node2ID.String())
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Should get validation error for min=3 nodes
		assert.Contains(t, w.Body.String(), "Invalid query parameters")
	})

	// Test 11: Unauthorized access (no session cookie)
	t.Run("unauthorized_access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/data/metrics?node_id="+node1ID.String(), nil)
		// No session cookie added

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test 12: Historical data with 5m aggregation
	t.Run("get_historical_data_5m_aggregation", func(t *testing.T) {
		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/history?node_id=%s&start_time=%s&end_time=%s&metric=latency&aggregation=5m",
			node1ID.String(), startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		aggregation, ok := resp["aggregation"].(string)
		assert.True(t, ok, "Expected aggregation in response")
		assert.Equal(t, "5m", aggregation, "Expected 5m aggregation")
	})

	// Test 13: Historical data with 1h aggregation
	t.Run("get_historical_data_1h_aggregation", func(t *testing.T) {
		startTime := now.Add(-24 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/history?node_id=%s&start_time=%s&end_time=%s&metric=latency&aggregation=1h",
			node1ID.String(), startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		aggregation, ok := resp["aggregation"].(string)
		assert.True(t, ok, "Expected aggregation in response")
		assert.Equal(t, "1h", aggregation, "Expected 1h aggregation")
	})

	// Test 14: Comparison with max nodes (5 nodes)
	t.Run("get_comparison_max_nodes", func(t *testing.T) {
		// Create 2 more nodes to reach max of 5
		node4ID := uuid.New()
		node5ID := uuid.New()

		for _, nodeID := range []uuid.UUID{node4ID, node5ID} {
			nodeName := fmt.Sprintf("test-node-%s", nodeID.String()[:8])
			_, err = pool.Exec(context.Background(),
				"INSERT INTO nodes (id, name, ip, region, tags, last_heartbeat, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
				nodeID, nodeName, "192.168.1.100", "us-east", "{}", now.Add(-2*time.Minute), "online",
			)
			require.NoError(t, err)

			// Insert metrics for new node
			baseTime := now.Add(-1 * time.Hour)
			for i := 0; i < 30; i++ {
				timestamp := baseTime.Add(time.Duration(i) * 2 * time.Minute)
				latency := 50.0 + float64(i%50)
				_, err = pool.Exec(context.Background(),
					`INSERT INTO metrics (node_id, timestamp, latency_ms, packet_loss_rate, jitter_ms)
					VALUES ($1, $2, $3, 0.01, 2.0)`,
					nodeID, timestamp, latency,
				)
				require.NoError(t, err)
			}
		}

		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		nodeIDs := fmt.Sprintf("%s,%s,%s,%s,%s", node1ID.String(), node2ID.String(), node3ID.String(), node4ID.String(), node5ID.String())
		url := fmt.Sprintf("/api/v1/data/comparison?node_ids=%s&start_time=%s&end_time=%s&metrics=latency",
			nodeIDs, startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok, "Expected data object in response")

		nodes, ok := data["nodes"].([]interface{})
		assert.True(t, ok, "Expected nodes array in data")
		assert.Equal(t, 5, len(nodes), "Expected 5 nodes in comparison")
	})

	// Test 15: Comparison with too many nodes (should fail with max=5)
	t.Run("get_comparison_too_many_nodes", func(t *testing.T) {
		node6ID := uuid.New()
		nodeName := fmt.Sprintf("test-node-%s", node6ID.String()[:8])
		_, err = pool.Exec(context.Background(),
			"INSERT INTO nodes (id, name, ip, region, tags, last_heartbeat, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
			node6ID, nodeName, "192.168.1.100", "us-east", "{}", now.Add(-2*time.Minute), "online",
		)
		require.NoError(t, err)

		startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endTime := now.Format(time.RFC3339)

		w := httptest.NewRecorder()
		nodeIDs := fmt.Sprintf("%s,%s,%s,%s,%s,%s", node1ID.String(), node2ID.String(), node3ID.String(), uuid.New().String(), uuid.New().String(), node6ID.String())
		url := fmt.Sprintf("/api/v1/data/comparison?node_ids=%s&start_time=%s&end_time=%s&metrics=latency",
			nodeIDs, startTime, endTime)
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Should get validation error for max=5 nodes
		assert.Contains(t, w.Body.String(), "Invalid query parameters")
	})

	// Test 16: Get historical data with invalid timestamp format
	t.Run("get_historical_data_invalid_timestamp", func(t *testing.T) {
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/v1/data/history?node_id=%s&start_time=invalid&end_time=invalid&metric=latency",
			node1ID.String())
		req, _ := http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errorMsg, ok := resp["error"].(string)
		assert.True(t, ok, "Expected error in response")
		assert.Contains(t, errorMsg, "Invalid start_time format")
	})

	// Cleanup
	t.Cleanup(func() {
		// Clean up all test data
		pool.Exec(context.Background(), "DELETE FROM metrics WHERE node_id = ANY($1)", []uuid.UUID{node1ID, node2ID, node3ID})
		pool.Exec(context.Background(), "DELETE FROM nodes WHERE id = ANY($1)", []uuid.UUID{node1ID, node2ID, node3ID})
		pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)
	})
}

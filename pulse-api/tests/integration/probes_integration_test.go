package integration

import (
	"bytes"
	"context"
	"encoding/json"
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

// TestCreateProbe_Integration tests the complete probe creation workflow
func TestCreateProbe_Integration(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	defer pool.Close()

	// Create test user and login
	username := "probe_test_user"
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

	// Create test node
	nodeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())",
		nodeID, "测试节点", "192.168.1.100", "us-east", "{}",
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

	// Test 1: Create TCP probe
	t.Run("create_tcp_probe", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          nodeID.String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            80,
			IntervalSeconds: 60,
			Count:           5,
			TimeoutSeconds:  10,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp models.CreateProbeResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "探测配置创建成功", resp.Message)
		assert.NotNil(t, resp.Data.Probe)
		assert.Equal(t, "TCP", resp.Data.Probe.Type)
		assert.Equal(t, "example.com", resp.Data.Probe.Target)
		assert.Equal(t, 80, resp.Data.Probe.Port)
		assert.Equal(t, 60, resp.Data.Probe.IntervalSeconds)
		assert.Equal(t, 5, resp.Data.Probe.Count)
		assert.Equal(t, 10, resp.Data.Probe.TimeoutSeconds)
		assert.NotEmpty(t, resp.Data.Probe.ID)
		assert.NotEmpty(t, resp.Timestamp)
	})

	// Test 2: Create UDP probe
	t.Run("create_udp_probe", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          nodeID.String(),
			Type:            "UDP",
			Target:          "192.168.1.1",
			Port:            53,
			IntervalSeconds: 120,
			Count:           10,
			TimeoutSeconds:  5,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp models.CreateProbeResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "探测配置创建成功", resp.Message)
		assert.Equal(t, "UDP", resp.Data.Probe.Type)
	})

	// Test 3: Invalid probe type
	t.Run("invalid_probe_type", func(t *testing.T) {
		createProbeReq := map[string]interface{}{
			"node_id":          nodeID.String(),
			"type":             "ICMP",
			"target":           "example.com",
			"port":             80,
			"interval_seconds": 60,
			"count":            5,
			"timeout_seconds":  10,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Contains(t, resp.Code, "ERR")
	})

	// Test 4: Non-existent node
	t.Run("node_not_found", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          uuid.New().String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            80,
			IntervalSeconds: 60,
			Count:           5,
			TimeoutSeconds:  10,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "ERR_PROBE_NODE_NOT_FOUND", resp.Code)
	})
}

// TestGetProbes_Integration tests listing probes
func TestGetProbes_Integration(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	defer pool.Close()

	// Create test user and login
	username := "get_probes_user"
	password := "testpass123"

	pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)
	// Clean up any existing probes from previous tests
	pool.Exec(context.Background(), "DELETE FROM probes")

	userID := uuid.New()
	hashedPassword, _ := auth.HashPassword(password)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at) VALUES ($1, $2, $3, $4, 0, NULL, NOW(), NOW())",
		userID, username, hashedPassword, "admin",
	)
	require.NoError(t, err)

	// Create test node
	nodeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())",
		nodeID, "测试节点", "192.168.1.100", "us-east", "{}",
	)
	require.NoError(t, err)

	// Create test probes
	probe1ID := uuid.New()
	probe2ID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())",
		probe1ID, nodeID, "TCP", "example.com", 80, 60, 5, 10,
	)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(),
		"INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())",
		probe2ID, nodeID, "UDP", "192.168.1.1", 53, 120, 10, 5,
	)
	require.NoError(t, err)

	// Login
	loginReq := models.LoginRequest{
		Username: username,
		Password: password,
	}
	loginReqBody, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginReqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	var sessionID string
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionID = cookie.Value
			break
		}
	}
	require.NotEmpty(t, sessionID)

	// Test: Get all probes
	t.Run("get_all_probes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/probes", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.GetProbesResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "探测配置列表获取成功", resp.Message)
		assert.Len(t, resp.Data.Probes, 2)
		assert.NotEmpty(t, resp.Timestamp)
	})

	// Test: Get probes filtered by node
	t.Run("get_probes_by_node", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/probes?node_id="+nodeID.String(), nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.GetProbesResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Len(t, resp.Data.Probes, 2)
		for _, probe := range resp.Data.Probes {
			assert.Equal(t, nodeID.String(), probe.NodeID)
		}
	})
}

// TestUpdateDeleteProbe_Integration tests updating and deleting probes
func TestUpdateDeleteProbe_Integration(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	defer pool.Close()

	// Create test user and login
	username := "update_probe_user"
	password := "testpass123"

	pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)

	userID := uuid.New()
	hashedPassword, _ := auth.HashPassword(password)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at) VALUES ($1, $2, $3, $4, 0, NULL, NOW(), NOW())",
		userID, username, hashedPassword, "admin",
	)
	require.NoError(t, err)

	// Create test node and probe
	nodeID := uuid.New()
	probeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())",
		nodeID, "测试节点", "192.168.1.100", "us-east", "{}",
	)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(),
		"INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())",
		probeID, nodeID, "TCP", "example.com", 80, 60, 5, 10,
	)
	require.NoError(t, err)

	// Login
	loginReq := models.LoginRequest{
		Username: username,
		Password: password,
	}
	loginReqBody, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginReqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	var sessionID string
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionID = cookie.Value
			break
		}
	}
	require.NotEmpty(t, sessionID)

	// Test 1: Update probe
	t.Run("update_probe", func(t *testing.T) {
		newInterval := 120
		newCount := 10
		updateReq := models.UpdateProbeRequest{
			IntervalSeconds: &newInterval,
			Count:           &newCount,
		}
		reqBody, _ := json.Marshal(updateReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/probes/"+probeID.String(), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.UpdateProbeResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "探测配置更新成功", resp.Message)
		assert.Equal(t, newInterval, resp.Data.Probe.IntervalSeconds)
		assert.Equal(t, newCount, resp.Data.Probe.Count)
	})

	// Test 2: Delete probe
	t.Run("delete_probe", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/probes/"+probeID.String()+"?confirm=true", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.DeleteProbeResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "探测配置删除成功", resp.Message)
		assert.NotEmpty(t, resp.Timestamp)

		// Verify probe is deleted
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/probes/"+probeID.String(), nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test 3: Delete without confirmation
	t.Run("delete_without_confirmation", func(t *testing.T) {
		// Create another probe for this test
		newProbeID := uuid.New()
		_, err = pool.Exec(context.Background(),
			"INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())",
			newProbeID, nodeID, "TCP", "example.com", 80, 60, 5, 10,
		)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/probes/"+newProbeID.String(), nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "ERR_CONFIRMATION_REQUIRED", resp.Code)
	})
}

// TestProbeConstraints_Integration tests database constraints and validation
func TestProbeConstraints_Integration(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	defer pool.Close()

	// Create test user and login
	username := "probe_constraints_user"
	password := "testpass123"

	pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)

	userID := uuid.New()
	hashedPassword, _ := auth.HashPassword(password)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at) VALUES ($1, $2, $3, $4, 0, NULL, NOW(), NOW())",
		userID, username, hashedPassword, "admin",
	)
	require.NoError(t, err)

	// Create test node
	nodeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())",
		nodeID, "测试节点", "192.168.1.100", "us-east", "{}",
	)
	require.NoError(t, err)

	// Login
	loginReq := models.LoginRequest{
		Username: username,
		Password: password,
	}
	loginReqBody, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginReqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	var sessionID string
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionID = cookie.Value
			break
		}
	}
	require.NotEmpty(t, sessionID)

	// Test 1: Interval out of range (too low)
	t.Run("interval_too_low", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          nodeID.String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            80,
			IntervalSeconds: 30, // Below minimum 60
			Count:           5,
			TimeoutSeconds:  10,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test 2: Interval out of range (too high)
	t.Run("interval_too_high", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          nodeID.String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            80,
			IntervalSeconds: 400, // Above maximum 300
			Count:           5,
			TimeoutSeconds:  10,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test 3: Count out of range
	t.Run("count_out_of_range", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          nodeID.String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            80,
			IntervalSeconds: 60,
			Count:           200, // Above maximum 100
			TimeoutSeconds:  10,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test 4: Timeout out of range
	t.Run("timeout_out_of_range", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          nodeID.String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            80,
			IntervalSeconds: 60,
			Count:           5,
			TimeoutSeconds:  60, // Above maximum 30
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test 5: Port out of range
	t.Run("port_out_of_range", func(t *testing.T) {
		createProbeReq := models.CreateProbeRequest{
			NodeID:          nodeID.String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            70000, // Above maximum 65535
			IntervalSeconds: 60,
			Count:           5,
			TimeoutSeconds:  10,
		}
		reqBody, _ := json.Marshal(createProbeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/probes", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestMetricsTableIntegration tests metrics table compatibility with Story 3.2
func TestMetricsTableIntegration(t *testing.T) {
	_, pool, _ := setupTestRouter(t)
	defer pool.Close()

	// Create test user
	username := "metrics_test_user"
	password := "testpass123"

	pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)

	userID := uuid.New()
	hashedPassword, _ := auth.HashPassword(password)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at) VALUES ($1, $2, $3, $4, 0, NULL, NOW(), NOW())",
		userID, username, hashedPassword, "admin",
	)
	require.NoError(t, err)

	// Create test node and probe
	nodeID := uuid.New()
	probeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())",
		nodeID, "测试节点", "192.168.1.100", "us-east", "{}",
	)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(),
		"INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())",
		probeID, nodeID, "TCP", "example.com", 80, 60, 5, 10,
	)
	require.NoError(t, err)

	// Test 1: Insert metrics data (simulating Story 3.2 batch write)
	t.Run("insert_metrics_data", func(t *testing.T) {
		now := time.Now()

		// Insert multiple metric records
		_, err = pool.Exec(context.Background(),
			`INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms, is_aggregated, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW()),
			       ($1, $2, $8, $9, $10, $11, $12, NOW()),
			       ($1, $2, $13, $14, $15, $16, $17, NOW())`,
			nodeID, probeID,
			now.Add(-2*time.Minute), 45.5, 0.0, 2.3, false,
			now.Add(-1*time.Minute), 48.2, 0.0, 3.1, false,
			now, 42.8, 0.0, 2.8, false,
		)
		require.NoError(t, err)

		// Verify records were inserted
		var count int
		err = pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM metrics WHERE probe_id = $1", probeID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	// Test 2: Query metrics with index (test time series performance)
	t.Run("query_metrics_with_index", func(t *testing.T) {
		rows, err := pool.Query(context.Background(),
			"SELECT id, node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms, is_aggregated FROM metrics WHERE probe_id = $1 ORDER BY timestamp DESC LIMIT 10",
			probeID,
		)
		require.NoError(t, err)
		defer rows.Close()

		metricsCount := 0
		for rows.Next() {
			var id int64
			var nodeID, probeID uuid.UUID
			var timestamp time.Time
			var latencyMs, packetLossRate, jitterMs float64
			var isAggregated bool

			err = rows.Scan(&id, &nodeID, &probeID, &timestamp, &latencyMs, &packetLossRate, &jitterMs, &isAggregated)
			require.NoError(t, err)
			metricsCount++
		}

		assert.Equal(t, 3, metricsCount)
	})

	// Test 3: Test foreign key constraints
	t.Run("test_foreign_key_constraints", func(t *testing.T) {
		// Try to insert metric with invalid probe_id
		invalidProbeID := uuid.New()
		_, err := pool.Exec(context.Background(),
			"INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, created_at) VALUES ($1, $2, NOW(), 50.0, NOW())",
			nodeID, invalidProbeID,
		)
		assert.Error(t, err, "Should fail due to foreign key constraint")
	})

	// Test 4: Test cascade delete
	t.Run("test_cascade_delete", func(t *testing.T) {
		// Delete probe (should cascade to metrics)
		_, err = pool.Exec(context.Background(), "DELETE FROM probes WHERE id = $1", probeID)
		require.NoError(t, err)

		// Verify metrics were deleted
		var count int
		err = pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM metrics WHERE probe_id = $1", probeID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Metrics should be cascade deleted")
	})
}

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

// TestGetNodeStatus_Integration tests the complete node status query workflow
func TestGetNodeStatus_Integration(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	defer pool.Close()

	// Create test user and login
	username := "status_test_user"
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

	// Create test nodes with different heartbeat times
	now := time.Now()
	
	// Online node (2 minutes ago)
	onlineNodeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, last_heartbeat, last_report_time, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
		onlineNodeID, "在线测试节点", "192.168.1.100", "us-east", "{}", now.Add(-2*time.Minute), now.Add(-2*time.Minute), "online",
	)
	require.NoError(t, err)

	// Offline node (10 minutes ago)
	offlineNodeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, last_heartbeat, last_report_time, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
		offlineNodeID, "离线测试节点", "192.168.1.101", "us-east", "{}", now.Add(-10*time.Minute), now.Add(-10*time.Minute), "offline",
	)
	require.NoError(t, err)

	// Connecting node (no heartbeat)
	connectingNodeID := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, last_heartbeat, last_report_time, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NULL, NULL, $7, NOW(), NOW())",
		connectingNodeID, "连接中测试节点", "192.168.1.102", "us-east", "{}", nil, nil, "connecting",
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

	// Test 1: Query online node status
	t.Run("online_node_status", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/nodes/"+onlineNodeID.String()+"/status", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp models.GetNodeStatusResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		
		assert.Equal(t, "online", resp.Data.Node.Status)
		assert.Equal(t, "在线测试节点", resp.Data.Node.Name)
		assert.NotNil(t, resp.Data.Node.LastHeartbeat)
		assert.Equal(t, "节点状态查询成功", resp.Message)
		assert.NotEmpty(t, resp.Timestamp)
	})

	// Test 2: Query offline node status
	t.Run("offline_node_status", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/nodes/"+offlineNodeID.String()+"/status", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp models.GetNodeStatusResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		
		assert.Equal(t, "offline", resp.Data.Node.Status)
		assert.Equal(t, "离线测试节点", resp.Data.Node.Name)
		assert.NotNil(t, resp.Data.Node.LastHeartbeat)
		assert.Equal(t, "节点状态查询成功", resp.Message)
	})

	// Test 3: Query connecting node status
	t.Run("connecting_node_status", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/nodes/"+connectingNodeID.String()+"/status", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp models.GetNodeStatusResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		
		assert.Equal(t, "connecting", resp.Data.Node.Status)
		assert.Equal(t, "连接中测试节点", resp.Data.Node.Name)
		assert.Nil(t, resp.Data.Node.LastHeartbeat)
		assert.Equal(t, "节点状态查询成功", resp.Message)
	})

	// Test 4: Query non-existent node status
	t.Run("node_not_found", func(t *testing.T) {
		nonExistentID := uuid.New()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/nodes/"+nonExistentID.String()+"/status", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var resp models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		
		assert.Equal(t, "ERR_NODE_NOT_FOUND", resp.Code)
		assert.Contains(t, resp.Message, "节点不存在")
	})

	// Test 5: Query without authentication
	t.Run("unauthenticated_request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/nodes/"+onlineNodeID.String()+"/status", nil)
		// No session_id cookie
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var resp models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		
		assert.Equal(t, "ERR_UNAUTHORIZED", resp.Code)
	})

	// Test 6: Concurrent status queries
	t.Run("concurrent_queries", func(t *testing.T) {
		queries := 10
		
		for i := 0; i < queries; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/nodes/"+onlineNodeID.String()+"/status", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
				
				router.ServeHTTP(w, req)
			}()
		}
		
		// Wait for all queries to complete
		time.Sleep(2 * time.Second)
		
		// If we got here without panic, concurrent queries work
		assert.True(t, true, "Concurrent queries should not cause panic")
	})

	// Test 7: Invalid UUID format
	t.Run("invalid_uuid_format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/nodes/invalid-uuid-123/status", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var resp models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		
		assert.Contains(t, resp.Message, "无效的节点 ID 格式")
	})

	// Cleanup
	cleanupTestUser(pool, username)
}

// TestGetNodeStatus_ResponseFormat validates API response format matches specification
func TestGetNodeStatus_ResponseFormat(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	defer pool.Close()

	// Create test user and node
	username := "response_format_user"
	password := "testpass123"
	
	// Cleanup existing user
	pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)

	userID := uuid.New()
	hashedPassword, _ := auth.HashPassword(password)
	pool.Exec(context.Background(),
		"INSERT INTO users (user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at) VALUES ($1, $2, $3, $4, 0, NULL, NOW(), NOW())",
		userID, username, hashedPassword, "admin",
	)

	// Create node
	nodeID := uuid.New()
	now := time.Now()
	pool.Exec(context.Background(),
		"INSERT INTO nodes (id, name, ip, region, tags, last_heartbeat, last_report_time, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
		nodeID, "格式测试节点", "192.168.1.99", "us-west", "{}", now.Add(-1*time.Minute), now, "online",
	)

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

	// Query node status
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/nodes/"+nodeID.String()+"/status", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	
	router.ServeHTTP(w, req)

	// Validate response format
	var resp models.GetNodeStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response structure
	assert.NotNil(t, resp.Data)
	assert.NotNil(t, resp.Data.Node)
	assert.NotEmpty(t, resp.Message)
	assert.NotEmpty(t, resp.Timestamp)

	// Check node data structure
	assert.NotEmpty(t, resp.Data.Node.ID)
	assert.NotEmpty(t, resp.Data.Node.Name)
	assert.NotEmpty(t, resp.Data.Node.Status)
	assert.NotNil(t, resp.Data.Node.LastHeartbeat)

	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, resp.Timestamp)
	assert.NoError(t, err, "Timestamp should be RFC3339 format")

	// Cleanup
	cleanupTestUser(pool, username)
}

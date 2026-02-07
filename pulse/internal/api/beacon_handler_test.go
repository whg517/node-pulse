package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/cache"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/pkg/middleware"
)

func TestMain(m *testing.M) {
	// Set up test config environment variables before running tests
	os.Setenv("PULSE_JWT_SECRET", "test-secret-key-for-jwt-token-generation-in-tests-min-64-bytes-long-for-security")
	os.Setenv("PULSE_SERVER_PORT", "8080")
	os.Setenv("PULSE_DB_HOST", "localhost")
	os.Setenv("PULSE_DB_PORT", "5432")
	os.Setenv("PULSE_DB_NAME", "test")
	os.Setenv("PULSE_DB_USER", "test")
	os.Setenv("PULSE_DB_PASSWORD", "test")

	// Load config (will use env vars if config file doesn't exist)
	_, err := config.Load()
	if err != nil {
		// If config file doesn't exist, that's okay for tests - env vars will be used
		fmt.Printf("Warning: Config file not found, using environment variables: %v\n", err)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// setupTestRouter creates a test router with beacon heartbeat endpoint and JWT auth
// Returns: router, authHeader for given nodeID
func setupTestRouter(nodeQuerier db.NodesQuerier, nodeID string) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create JWT service for testing (config loaded in TestMain)
	jwtService, err := auth.NewJWTService()
	if err != nil {
		panic(fmt.Sprintf("Failed to create JWT service: %v", err))
	}

	// Generate a beacon JWT token for testing with the given nodeID
	token, _, err := jwtService.GenerateAccessToken(nodeID, "beacon")
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test token: %v", err))
	}
	authHeader := fmt.Sprintf("Bearer %s", token)

	// Create memory cache and batch writer for testing
	memoryCache := cache.NewMemoryCache()
	batchWriter := cache.NewBatchWriter(nil, 1000, 100) // nil DB for testing

	beaconHandler := NewBeaconHandler(nodeQuerier, memoryCache, batchWriter, nil) // nil alert engine for tests
	router.POST("/api/v1/beacon/heartbeat", middleware.JWTAuthMiddleware(), beaconHandler.HandleHeartbeat)

	return router, authHeader
}

func TestHandleHeartbeat_ValidNodeAndValidMetrics(t *testing.T) {
	// Arrange
	testNodeID := uuid.New()
	mockQuerier := &MockNodesQuerier{
		getNodeByIDFunc: func(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
			return &models.Node{
				ID:     testNodeID.String(),
				Name:   "test-node",
				IP:     "192.168.1.1",
				Region: "us-west",
			}, nil
		},
	}

	router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

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
		req.Header.Set("Authorization", authHeader)
	req.Header.Set("Authorization", authHeader)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.HeartbeatSuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "心跳数据接收成功", resp.Message)
	assert.True(t, resp.Data.Received)
}

func TestHandleHeartbeat_InvalidNodeID_Returns400(t *testing.T) {
	// Arrange
	testNodeID := uuid.New() // For auth token
	mockQuerier := &MockNodesQuerier{}
	router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

	reqBody := models.HeartbeatRequest{
		NodeID:         "invalid-uuid-format",
		ProbeID:        "probe-001",
		LatencyMs:      50.5,
		PacketLossRate: 0.1,
		JitterMs:       5.2,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ERR_INVALID_NODE_ID", resp.Code)
	assert.Contains(t, resp.Message, "节点 ID 格式无效")
}

func TestHandleHeartbeat_NodeNotFound_Returns400(t *testing.T) {
	// Arrange
	testNodeID := uuid.New()
	mockQuerier := &MockNodesQuerier{
		getNodeByIDFunc: func(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
			return nil, db.ErrNodeNotFound
		},
	}

	router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

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
		req.Header.Set("Authorization", authHeader)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, ErrNodeNotFound, resp.Code)
	assert.Contains(t, resp.Message, "节点不存在")
}

func TestHandleHeartbeat_LatencyOutOfRange_Returns400(t *testing.T) {
	tests := []struct {
		name      string
		latencyMs float64
	}{
		{"Negative latency", -1.0},
		{"Latency too high", 60001.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			testNodeID := uuid.New()
			mockQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:     testNodeID.String(),
						Name:   "test-node",
						IP:     "192.168.1.1",
						Region: "us-west",
					}, nil
				},
			}

			router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

			reqBody := models.HeartbeatRequest{
				NodeID:         testNodeID.String(),
				ProbeID:        "probe-001",
				LatencyMs:      tt.latencyMs,
				PacketLossRate: 0.1,
				JitterMs:       5.2,
				Timestamp:      time.Now().Format(time.RFC3339),
			}

			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)

			// Act
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var resp models.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, ErrInvalidLatency, resp.Code)
			assert.Contains(t, resp.Message, "时延超出范围")
		})
	}
}

func TestHandleHeartbeat_PacketLossOutOfRange_Returns400(t *testing.T) {
	tests := []struct {
		name           string
		packetLossRate float64
	}{
		{"Negative packet loss", -1.0},
		{"Packet loss too high", 101.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			testNodeID := uuid.New()
			mockQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:     testNodeID.String(),
						Name:   "test-node",
						IP:     "192.168.1.1",
						Region: "us-west",
					}, nil
				},
			}

			router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

			reqBody := models.HeartbeatRequest{
				NodeID:         testNodeID.String(),
				ProbeID:        "probe-001",
				LatencyMs:      50.5,
				PacketLossRate: tt.packetLossRate,
				JitterMs:       5.2,
				Timestamp:      time.Now().Format(time.RFC3339),
			}

			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)

			// Act
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var resp models.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, ErrInvalidPacketLoss, resp.Code)
			assert.Contains(t, resp.Message, "丢包率超出范围")
		})
	}
}

func TestHandleHeartbeat_JitterOutOfRange_Returns400(t *testing.T) {
	tests := []struct {
		name     string
		jitterMs float64
	}{
		{"Negative jitter", -1.0},
		{"Jitter too high", 50001.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			testNodeID := uuid.New()
			mockQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:     testNodeID.String(),
						Name:   "test-node",
						IP:     "192.168.1.1",
						Region: "us-west",
					}, nil
				},
			}

			router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

			reqBody := models.HeartbeatRequest{
				NodeID:         testNodeID.String(),
				ProbeID:        "probe-001",
				LatencyMs:      50.5,
				PacketLossRate: 0.1,
				JitterMs:       tt.jitterMs,
				Timestamp:      time.Now().Format(time.RFC3339),
			}

			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)

			// Act
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var resp models.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, ErrInvalidJitter, resp.Code)
			assert.Contains(t, resp.Message, "抖动超出范围")
		})
	}
}

func TestHandleHeartbeat_MissingRequiredFields_Returns400(t *testing.T) {
	// Arrange
	testNodeID := uuid.New() // For auth token
	mockQuerier := &MockNodesQuerier{}
	router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

	// Missing required fields
	reqBody := map[string]interface{}{
		"node_id": uuid.New().String(),
		// Missing probe_id, latency_ms, packet_loss_rate, jitter_ms, timestamp
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ERR_INVALID_REQUEST", resp.Code)
	assert.Contains(t, resp.Message, "请求参数无效")
}

func TestHandleHeartbeat_InvalidTimestampFormat_Returns400(t *testing.T) {
	// Arrange
	testNodeID := uuid.New()
	mockQuerier := &MockNodesQuerier{
		getNodeByIDFunc: func(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
			return &models.Node{
				ID:     testNodeID.String(),
				Name:   "test-node",
				IP:     "192.168.1.1",
				Region: "us-west",
			}, nil
		},
	}

	router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

	reqBody := models.HeartbeatRequest{
		NodeID:         testNodeID.String(),
		ProbeID:        "probe-001",
		LatencyMs:      50.5,
		PacketLossRate: 0.1,
		JitterMs:       5.2,
		Timestamp:      "invalid-timestamp-format",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, ErrInvalidTimestamp, resp.Code)
	assert.Contains(t, resp.Message, "时间戳格式无效")
}

func TestHandleHeartbeat_InvalidProbeID_Returns400(t *testing.T) {
	// Test that probe_id longer than 255 characters is rejected
	t.Run("Too long probe_id", func(t *testing.T) {
		// Arrange
		testNodeID := uuid.New()
		mockQuerier := &MockNodesQuerier{
			getNodeByIDFunc: func(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
				return &models.Node{
					ID:     testNodeID.String(),
					Name:   "test-node",
					IP:     "192.168.1.1",
					Region: "us-west",
				}, nil
			},
		}

		router, authHeader := setupTestRouter(mockQuerier, testNodeID.String())

		// Create a probe_id that exceeds 255 characters
		longProbeID := string(make([]byte, 256))

		reqBody := models.HeartbeatRequest{
			NodeID:         testNodeID.String(),
			ProbeID:        longProbeID,
			LatencyMs:      50.5,
			PacketLossRate: 0.1,
			JitterMs:       5.2,
			Timestamp:      time.Now().Format(time.RFC3339),
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/beacon/heartbeat", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "ERR_INVALID_PROBE_ID", resp.Code)
		assert.Contains(t, resp.Message, "探针 ID 格式无效")
	})
}

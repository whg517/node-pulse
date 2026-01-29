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
	"github.com/google/uuid"
	"github.com/kevin/node-pulse/pulse-api/internal/cache"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a test router with beacon heartbeat endpoint
func setupTestRouter(nodeQuerier db.NodesQuerier) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create memory cache and batch writer for testing
	memoryCache := cache.NewMemoryCache()
	batchWriter := cache.NewBatchWriter(nil, 1000, 100) // nil DB for testing

	beaconHandler := NewBeaconHandler(nodeQuerier, memoryCache, batchWriter)
	router.POST("/api/v1/beacon/heartbeat", beaconHandler.HandleHeartbeat)

	return router
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

	router := setupTestRouter(mockQuerier)

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
	mockQuerier := &MockNodesQuerier{}
	router := setupTestRouter(mockQuerier)

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

	router := setupTestRouter(mockQuerier)

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

			router := setupTestRouter(mockQuerier)

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

			router := setupTestRouter(mockQuerier)

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

			router := setupTestRouter(mockQuerier)

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
	mockQuerier := &MockNodesQuerier{}
	router := setupTestRouter(mockQuerier)

	// Missing required fields
	reqBody := map[string]interface{}{
		"node_id": uuid.New().String(),
		// Missing probe_id, latency_ms, packet_loss_rate, jitter_ms, timestamp
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

	router := setupTestRouter(mockQuerier)

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

		router := setupTestRouter(mockQuerier)

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

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

var (
	ErrInvalidLatency    = "ERR_INVALID_LATENCY"
	ErrInvalidPacketLoss = "ERR_INVALID_PACKET_LOSS"
	ErrInvalidJitter     = "ERR_INVALID_JITTER"
	ErrInvalidTimestamp  = "ERR_INVALID_TIMESTAMP"
	ErrRateLimitExceeded = "ERR_RATE_LIMIT_EXCEEDED"
)

// BeaconHandler handles beacon heartbeat API requests
type BeaconHandler struct {
	nodeQuerier db.NodesQuerier
}

// NewBeaconHandler creates a new BeaconHandler
func NewBeaconHandler(nodeQuerier db.NodesQuerier) *BeaconHandler {
	return &BeaconHandler{
		nodeQuerier: nodeQuerier,
	}
}

// HandleHeartbeat handles POST /api/v1/beacon/heartbeat
func (h *BeaconHandler) HandleHeartbeat(c *gin.Context) {
	// Parse request body
	var req models.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate node ID format
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_NODE_ID",
			Message: "节点 ID 格式无效",
			Details: map[string]interface{}{
				"node_id": req.NodeID,
				"error":   err.Error(),
			},
		})
		return
	}

	// Validate probe_id format (max length check)
	if len(req.ProbeID) > 255 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_PROBE_ID",
			Message: "探针 ID 格式无效",
			Details: map[string]interface{}{
				"probe_id": req.ProbeID,
				"reason":   "probe_id must be <= 255 characters",
			},
		})
		return
	}

	// Validate node ID exists
	ctx := context.Background()
	_, err = h.nodeQuerier.GetNodeByID(ctx, nodeID)
	if err != nil {
		if err == db.ErrNodeNotFound {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrNodeNotFound,
				Message: "节点不存在",
				Details: map[string]interface{}{
					"node_id": req.NodeID,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "节点查询失败",
			Details: err.Error(),
		})
		return
	}

	// Validate latency range (0-60000ms)
	if req.LatencyMs < 0 || req.LatencyMs > 60000 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidLatency,
			Message: "时延超出范围",
			Details: map[string]interface{}{
				"field":     "latency_ms",
				"value":     req.LatencyMs,
				"min":       0,
				"max":       60000,
				"unit":      "ms",
			},
		})
		return
	}

	// Validate packet loss rate range (0-100%)
	if req.PacketLossRate < 0 || req.PacketLossRate > 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidPacketLoss,
			Message: "丢包率超出范围",
			Details: map[string]interface{}{
				"field":     "packet_loss_rate",
				"value":     req.PacketLossRate,
				"min":       0,
				"max":       100,
				"unit":      "%",
			},
		})
		return
	}

	// Validate jitter range (0-50000ms)
	if req.JitterMs < 0 || req.JitterMs > 50000 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidJitter,
			Message: "抖动超出范围",
			Details: map[string]interface{}{
				"field":     "jitter_ms",
				"value":     req.JitterMs,
				"min":       0,
				"max":       50000,
				"unit":      "ms",
			},
		})
		return
	}

	// Validate timestamp format
	_, err = time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidTimestamp,
			Message: "时间戳格式无效",
			Details: map[string]interface{}{
				"field":     "timestamp",
				"value":     req.Timestamp,
				"expected":  "ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
				"error":     err.Error(),
			},
		})
		return
	}

	// TODO: Story 3.2 will implement memory cache and async write
	// For now, just validate and return success

	c.JSON(http.StatusOK, models.HeartbeatSuccessResponse{
		Data: models.HeartbeatData{
			Received:  true,
			NodeID:    req.NodeID,
			Timestamp: time.Now(),
		},
		Message:   "心跳数据接收成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"hash/crc32"
	"io"
	"net/http"
	"sync"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whg517/node-pulse/pulse/internal/alert"
	"github.com/whg517/node-pulse/pulse/internal/cache"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/pkg/middleware"
)

var (
	ErrInvalidLatency              = "ERR_INVALID_LATENCY"
	ErrInvalidPacketLoss           = "ERR_INVALID_PACKET_LOSS"
	ErrInvalidJitter               = "ERR_INVALID_JITTER"
	ErrInvalidTimestamp            = "ERR_INVALID_TIMESTAMP"
	ErrRateLimitExceeded           = "ERR_RATE_LIMIT_EXCEEDED"
	ErrUnauthorizedNode            = "ERR_UNAUTHORIZED_NODE"
	ErrCompressionCorrupted        = "ERR_COMPRESSION_CORRUPTED"
	ErrInvalidConfig               = "ERR_INVALID_CONFIG"
	ErrConfigConflict              = "ERR_CONFIG_CONFLICT"
	ErrBeaconConfigNotFound        = "ERR_BEACON_CONFIG_NOT_FOUND"
	ErrBeaconGroupNotFound         = "ERR_BEACON_GROUP_NOT_FOUND"
)

// BeaconHandler handles beacon heartbeat API requests
type BeaconHandler struct {
	nodeQuerier db.NodesQuerier
	memoryCache *cache.MemoryCache
	batchWriter *cache.BatchWriter
	alertEngine *alert.AlertEngine
}

// NewBeaconHandler creates a new BeaconHandler
func NewBeaconHandler(nodeQuerier db.NodesQuerier, memoryCache *cache.MemoryCache, batchWriter *cache.BatchWriter, alertEngine *alert.AlertEngine) *BeaconHandler {
	return &BeaconHandler{
		nodeQuerier: nodeQuerier,
		memoryCache: memoryCache,
		batchWriter: batchWriter,
		alertEngine: alertEngine,
	}
}

// HandleHeartbeat handles POST /api/v1/beacon/heartbeat
func (h *BeaconHandler) HandleHeartbeat(c *gin.Context) {
	// Get user info from context (set by JWTAuthMiddleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "Unauthorized",
		})
		return
	}

	role, err := middleware.GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "Unauthorized",
		})
		return
	}

	// Verify role is "beacon"
	if role != "beacon" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "权限不足：需要 beacon 角色",
			Details: map[string]interface{}{
				"role":      role,
				"required":  "beacon",
			},
		})
		return
	}

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

	// Verify JWT token's user_id matches the requested node_id
	// This ensures a beacon can only report for its own node
	if userID != req.NodeID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "权限不足：Token 节点 ID 与请求不匹配",
			Details: map[string]interface{}{
				"token_node_id": userID,
				"request_node_id": req.NodeID,
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
				"field": "latency_ms",
				"value": req.LatencyMs,
				"min":   0,
				"max":   60000,
				"unit":  "ms",
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
				"field": "packet_loss_rate",
				"value": req.PacketLossRate,
				"min":   0,
				"max":   100,
				"unit":  "%",
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
				"field": "jitter_ms",
				"value": req.JitterMs,
				"min":   0,
				"max":   50000,
				"unit":  "ms",
			},
		})
		return
	}

	// Validate timestamp format
	parsedTime, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidTimestamp,
			Message: "时间戳格式无效",
			Details: map[string]interface{}{
				"field":    "timestamp",
				"value":    req.Timestamp,
				"expected": "ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
				"error":    err.Error(),
			},
		})
		return
	}

	// Write to memory cache (Story 3.2 implementation)
	metricPoint := &cache.MetricPoint{
		Timestamp:      parsedTime,
		LatencyMs:      req.LatencyMs,
		PacketLossRate: req.PacketLossRate,
		JitterMs:       req.JitterMs,
	}

	if err := h.memoryCache.Store(req.NodeID, metricPoint); err != nil {
		slog.Error("Failed to write to memory cache",
			"node_id", req.NodeID,
			"error", err)
		// Don't return error to avoid affecting Beacon reporting
	}

	// Send to batch writer buffer (non-blocking)
	metricRecord := &cache.MetricRecord{
		NodeID:         req.NodeID,
		ProbeID:        req.ProbeID,
		Timestamp:      parsedTime,
		LatencyMs:      req.LatencyMs,
		PacketLossRate: req.PacketLossRate,
		JitterMs:       req.JitterMs,
		IsAggregated:   false,
	}

	if err := h.batchWriter.Write(metricRecord); err != nil {
		if err == cache.ErrBufferFull {
			slog.Warn("Batch writer buffer full, dropping metric",
				"node_id", req.NodeID,
				"probe_id", req.ProbeID)
		} else {
			slog.Error("Failed to write to batch buffer",
				"node_id", req.NodeID,
				"error", err)
		}
		// Don't return error to avoid affecting Beacon reporting
	}

	// Trigger alert evaluation (async, non-blocking)
	if h.alertEngine != nil {
		metricData := &alert.MetricData{
			NodeID:         req.NodeID,
			LatencyMs:      req.LatencyMs,
			PacketLossRate: req.PacketLossRate,
			JitterMs:       req.JitterMs,
			Timestamp:      parsedTime,
		}
		if !h.alertEngine.EvaluateMetrics(metricData) {
			slog.Warn("Alert engine metric channel full, skipping evaluation",
				"node_id", req.NodeID,
				"timestamp", parsedTime)
		}
	}

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

// CompressedHeartbeatRequest represents a compressed heartbeat payload
type CompressedHeartbeatRequest struct {
	Data     []byte `json:"data" binding:"required"`      // Gzip compressed JSON data
	Checksum uint32 `json:"checksum" binding:"required"`  // CRC32 checksum
}

// HandleCompressedHeartbeat handles POST /api/v1/beacon/heartbeat/compressed
// Supports gzip-compressed heartbeat data (FR-4.1.5)
func (h *BeaconHandler) HandleCompressedHeartbeat(c *gin.Context) {
	// Get user info from context (set by JWTAuthMiddleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "Unauthorized",
		})
		return
	}

	role, err := middleware.GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "Unauthorized",
		})
		return
	}

	// Verify role is "beacon"
	if role != "beacon" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "权限不足：需要 beacon 角色",
		})
		return
	}

	// Parse compressed request
	var compReq CompressedHeartbeatRequest
	if err := c.ShouldBindJSON(&compReq); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate CRC32 checksum
	computedChecksum := crc32.ChecksumIEEE(compReq.Data)
	if computedChecksum != compReq.Checksum {
		RecordCompressionCorruption()
		slog.Error("CRC32 checksum mismatch",
			"expected", compReq.Checksum,
			"computed", computedChecksum)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrCompressionCorrupted,
			Message: "压缩数据校验失败",
			Details: map[string]interface{}{
				"expected_checksum": compReq.Checksum,
				"computed_checksum": computedChecksum,
			},
		})
		return
	}

	// Decompress gzip data
	reader, err := gzip.NewReader(bytes.NewReader(compReq.Data))
	if err != nil {
		RecordCompressionCorruption()
		slog.Error("Failed to decompress gzip data", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrCompressionCorrupted,
			Message: "解压缩失败",
			Details: err.Error(),
		})
		return
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		RecordCompressionCorruption()
		slog.Error("Failed to read decompressed data", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrCompressionCorrupted,
			Message: "读取解压数据失败",
			Details: err.Error(),
		})
		return
	}

	// Parse the decompressed JSON into heartbeat request
	var req models.HeartbeatRequest
	if err := json.Unmarshal(decompressed, &req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Verify JWT token's user_id matches the requested node_id
	if userID != req.NodeID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "权限不足：Token 节点 ID 与请求不匹配",
		})
		return
	}

	// Validate node ID format
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_NODE_ID",
			Message: "节点 ID 格式无效",
		})
		return
	}

	// Validate node exists
	ctx := context.Background()
	_, err = h.nodeQuerier.GetNodeByID(ctx, nodeID)
	if err != nil {
		if err == db.ErrNodeNotFound {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrNodeNotFound,
				Message: "节点不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "节点查询失败",
		})
		return
	}

	// Validate timestamp
	parsedTime, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidTimestamp,
			Message: "时间戳格式无效",
		})
		return
	}

	// Write to memory cache
	metricPoint := &cache.MetricPoint{
		Timestamp:      parsedTime,
		LatencyMs:      req.LatencyMs,
		PacketLossRate: req.PacketLossRate,
		JitterMs:       req.JitterMs,
	}
	h.memoryCache.Store(req.NodeID, metricPoint)

	// Send to batch writer
	metricRecord := &cache.MetricRecord{
		NodeID:         req.NodeID,
		ProbeID:        req.ProbeID,
		Timestamp:      parsedTime,
		LatencyMs:      req.LatencyMs,
		PacketLossRate: req.PacketLossRate,
		JitterMs:       req.JitterMs,
		IsAggregated:   false,
	}
	h.batchWriter.Write(metricRecord)

	// Trigger alert evaluation
	if h.alertEngine != nil {
		metricData := &alert.MetricData{
			NodeID:         req.NodeID,
			LatencyMs:      req.LatencyMs,
			PacketLossRate: req.PacketLossRate,
			JitterMs:       req.JitterMs,
			Timestamp:      parsedTime,
		}
		h.alertEngine.EvaluateMetrics(metricData)
	}

	c.JSON(http.StatusOK, models.HeartbeatSuccessResponse{
		Data: models.HeartbeatData{
			Received:  true,
			NodeID:    req.NodeID,
			Timestamp: time.Now(),
		},
		Message:   "压缩心跳数据接收成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// BeaconConfig represents a beacon's probe configuration
type BeaconConfig struct {
	Probes          []ProbeConfig `json:"probes"`
	IntervalSeconds int           `json:"interval_seconds"`
	TimeoutSeconds  int           `json:"timeout_seconds"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Version         int           `json:"version"`
}

// ProbeConfig represents a single probe configuration
type ProbeConfig struct {
	ID              string `json:"id"`
	Type            string `json:"type"`       // TCP or UDP
	Target          string `json:"target"`
	Port            int    `json:"port"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	Count           int    `json:"count"`
}

// BeaconConfigUpdateRequest represents a request to update beacon config
type BeaconConfigUpdateRequest struct {
	Probes          *[]ProbeConfig `json:"probes,omitempty"`
	IntervalSeconds *int           `json:"interval_seconds,omitempty"`
	TimeoutSeconds  *int           `json:"timeout_seconds,omitempty"`
}

// BeaconConfigResponse represents beacon config response
type BeaconConfigResponse struct {
	Data      BeaconConfig `json:"data"`
	Message   string       `json:"message"`
	Timestamp string       `json:"timestamp"`
}

// ConfigHistoryEntry represents a config history entry
type ConfigHistoryEntry struct {
	Version   int           `json:"version"`
	Config    BeaconConfig  `json:"config"`
	ChangedAt time.Time     `json:"changed_at"`
	ChangedBy string        `json:"changed_by"`
}

// ConfigHistoryResponse represents config history response
type ConfigHistoryResponse struct {
	Data      []ConfigHistoryEntry `json:"data"`
	Message   string               `json:"message"`
	Timestamp string               `json:"timestamp"`
}

// In-memory beacon config store (for MVP - in production use database)
var (
	beaconConfigStore     = make(map[string]*BeaconConfig)
	beaconConfigHistory   = make(map[string][]ConfigHistoryEntry)
	beaconConfigMutex     sync.RWMutex
)

// GetBeaconConfig handles GET /api/v1/beacons/:id/config
func (h *BeaconHandler) GetBeaconConfig(c *gin.Context) {
	beaconID := c.Param("id")

	// Validate beacon exists
	_, err := uuid.Parse(beaconID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_BEACON_ID",
			Message: "Beacon ID 格式无效",
		})
		return
	}

	beaconConfigMutex.RLock()
	config, exists := beaconConfigStore[beaconID]
	beaconConfigMutex.RUnlock()

	if !exists {
		// Return default config
		config = &BeaconConfig{
			Probes:          []ProbeConfig{},
			IntervalSeconds: 60,
			TimeoutSeconds:  5,
			UpdatedAt:       time.Now(),
			Version:         1,
		}
	}

	c.JSON(http.StatusOK, BeaconConfigResponse{
		Data:      *config,
		Message:   "Beacon 配置获取成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateBeaconConfig handles POST /api/v1/beacons/:id/config
func (h *BeaconHandler) UpdateBeaconConfig(c *gin.Context) {
	beaconID := c.Param("id")

	// Validate beacon ID
	_, err := uuid.Parse(beaconID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_BEACON_ID",
			Message: "Beacon ID 格式无效",
		})
		return
	}

	// Parse request
	var req BeaconConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate config
	if err := validateBeaconConfig(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidConfig,
			Message: "配置验证失败",
			Details: err.Error(),
		})
		return
	}

	beaconConfigMutex.Lock()
	defer beaconConfigMutex.Unlock()

	// Get existing config or create new
	existing, exists := beaconConfigStore[beaconID]
	if !exists {
		existing = &BeaconConfig{
			Probes:          []ProbeConfig{},
			IntervalSeconds: 60,
			TimeoutSeconds:  5,
			Version:         0,
		}
	}

	// Store history before update
	historyEntry := ConfigHistoryEntry{
		Version:   existing.Version,
		Config:    *existing,
		ChangedAt: time.Now(),
		ChangedBy: "system", // In production, get from auth context
	}
	beaconConfigHistory[beaconID] = append(beaconConfigHistory[beaconID], historyEntry)

	// Apply updates
	if req.Probes != nil {
		existing.Probes = *req.Probes
	}
	if req.IntervalSeconds != nil {
		existing.IntervalSeconds = *req.IntervalSeconds
	}
	if req.TimeoutSeconds != nil {
		existing.TimeoutSeconds = *req.TimeoutSeconds
	}
	existing.Version++
	existing.UpdatedAt = time.Now()

	beaconConfigStore[beaconID] = existing

	c.JSON(http.StatusOK, BeaconConfigResponse{
		Data:      *existing,
		Message:   "Beacon 配置更新成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetBeaconConfigHistory handles GET /api/v1/beacons/:id/config/history
func (h *BeaconHandler) GetBeaconConfigHistory(c *gin.Context) {
	beaconID := c.Param("id")

	// Validate beacon ID
	_, err := uuid.Parse(beaconID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_BEACON_ID",
			Message: "Beacon ID 格式无效",
		})
		return
	}

	beaconConfigMutex.RLock()
	history := beaconConfigHistory[beaconID]
	beaconConfigMutex.RUnlock()

	// Limit to last 50 entries
	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	c.JSON(http.StatusOK, ConfigHistoryResponse{
		Data:      history,
		Message:   "配置历史获取成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// BatchConfigUpdateRequest represents a batch config update request
type BatchConfigUpdateRequest struct {
	BeaconIDs       []string               `json:"beacon_ids" binding:"required"`
	Config          BeaconConfigUpdateRequest `json:"config" binding:"required"`
}

// BatchConfigUpdateResponse represents batch config update response
type BatchConfigUpdateResponse struct {
	Data      BatchConfigResult `json:"data"`
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
}

// BatchConfigResult represents the result of batch config update
type BatchConfigResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// BatchUpdateBeaconGroupConfig handles POST /api/v1/beacon-groups/:gid/config
func (h *BeaconHandler) BatchUpdateBeaconGroupConfig(c *gin.Context) {
	groupID := c.Param("gid")

	// Parse request
	var req BatchConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate config
	if err := validateBeaconConfig(&req.Config); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidConfig,
			Message: "配置验证失败",
			Details: err.Error(),
		})
		return
	}

	result := BatchConfigResult{
		SuccessCount: 0,
		FailedCount:  0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	// Apply config to each beacon
	for _, beaconID := range req.BeaconIDs {
		// Validate beacon ID format
		if _, err := uuid.Parse(beaconID); err != nil {
			result.FailedCount++
			result.FailedIDs = append(result.FailedIDs, beaconID)
			result.Errors = append(result.Errors, "invalid beacon ID: "+beaconID)
			continue
		}

		beaconConfigMutex.Lock()

		existing, exists := beaconConfigStore[beaconID]
		if !exists {
			existing = &BeaconConfig{
				Probes:          []ProbeConfig{},
				IntervalSeconds: 60,
				TimeoutSeconds:  5,
				Version:         0,
			}
		}

		// Store history
		historyEntry := ConfigHistoryEntry{
			Version:   existing.Version,
			Config:    *existing,
			ChangedAt: time.Now(),
			ChangedBy: "batch:" + groupID,
		}
		beaconConfigHistory[beaconID] = append(beaconConfigHistory[beaconID], historyEntry)

		// Apply updates
		if req.Config.Probes != nil {
			existing.Probes = *req.Config.Probes
		}
		if req.Config.IntervalSeconds != nil {
			existing.IntervalSeconds = *req.Config.IntervalSeconds
		}
		if req.Config.TimeoutSeconds != nil {
			existing.TimeoutSeconds = *req.Config.TimeoutSeconds
		}
		existing.Version++
		existing.UpdatedAt = time.Now()

		beaconConfigStore[beaconID] = existing
		beaconConfigMutex.Unlock()

		result.SuccessCount++
	}

	c.JSON(http.StatusOK, BatchConfigUpdateResponse{
		Data:      result,
		Message:   "批量配置更新完成",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// validateBeaconConfig validates beacon configuration
func validateBeaconConfig(req *BeaconConfigUpdateRequest) error {
	// Check interval is not too short (minimum 5 seconds)
	if req.IntervalSeconds != nil && *req.IntervalSeconds < 5 {
		return errors.New("interval too short: minimum 5 seconds required")
	}

	// Check timeout is reasonable
	if req.TimeoutSeconds != nil && *req.TimeoutSeconds < 1 {
		return errors.New("timeout too short: minimum 1 second required")
	}

	// Validate probes if provided
	if req.Probes != nil {
		for _, probe := range *req.Probes {
			// Check probe interval vs global interval conflict
			if probe.IntervalSeconds < 5 {
				return errors.New("probe interval too short: minimum 5 seconds required")
			}
			if probe.TimeoutSeconds < 1 {
				return errors.New("probe timeout too short: minimum 1 second required")
			}
			if probe.Type != "TCP" && probe.Type != "UDP" {
				return errors.New("invalid probe type: must be TCP or UDP")
			}
			if probe.Target == "" {
				return errors.New("probe target cannot be empty")
			}
			if probe.Port < 1 || probe.Port > 65535 {
				return errors.New("probe port must be between 1 and 65535")
			}
		}
	}

	return nil
}

// GetConfigPreview handles POST /api/v1/beacons/:id/config/preview
// Returns validation result without applying changes
func (h *BeaconHandler) GetConfigPreview(c *gin.Context) {
	beaconID := c.Param("id")

	// Validate beacon ID
	_, err := uuid.Parse(beaconID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_BEACON_ID",
			Message: "Beacon ID 格式无效",
		})
		return
	}

	// Parse request
	var req BeaconConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate config
	validationErr := validateBeaconConfig(&req)

	warnings := []string{}
	conflicts := []string{}

	// Check for potential conflicts
	if req.IntervalSeconds != nil && *req.IntervalSeconds < 30 {
		warnings = append(warnings, "间隔时间小于 30 秒可能导致性能问题")
	}
	if req.Probes != nil && len(*req.Probes) > 50 {
		warnings = append(warnings, "探测配置数量超过 50 个可能导致资源消耗过高")
	}

	if validationErr != nil {
		conflicts = append(conflicts, validationErr.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"valid":      validationErr == nil,
			"warnings":   warnings,
			"conflicts":  conflicts,
			"preview_of": req,
		},
		"message":   "配置预览生成成功",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

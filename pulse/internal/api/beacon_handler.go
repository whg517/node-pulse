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
	ErrInvalidLatency       = "ERR_INVALID_LATENCY"
	ErrInvalidPacketLoss    = "ERR_INVALID_PACKET_LOSS"
	ErrInvalidJitter        = "ERR_INVALID_JITTER"
	ErrInvalidTimestamp     = "ERR_INVALID_TIMESTAMP"
	ErrRateLimitExceeded    = "ERR_RATE_LIMIT_EXCEEDED"
	ErrUnauthorizedNode     = "ERR_UNAUTHORIZED_NODE"
	ErrCompressionCorrupted = "ERR_COMPRESSION_CORRUPTED"
	ErrInvalidConfig        = "ERR_INVALID_CONFIG"
	ErrConfigConflict       = "ERR_CONFIG_CONFLICT"
	ErrBeaconConfigNotFound = "ERR_BEACON_CONFIG_NOT_FOUND"
	ErrBeaconGroupNotFound  = "ERR_BEACON_GROUP_NOT_FOUND"
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

func (h *BeaconHandler) beaconConfigQuerier() db.BeaconConfigsQuerier {
	if h == nil || h.nodeQuerier == nil {
		return nil
	}
	querier, ok := h.nodeQuerier.(db.BeaconConfigsQuerier)
	if !ok {
		return nil
	}
	return querier
}

func (h *BeaconHandler) mtrResultsQuerier() db.MTRResultsQuerier {
	if h == nil || h.nodeQuerier == nil {
		return nil
	}
	querier, ok := h.nodeQuerier.(db.MTRResultsQuerier)
	if !ok {
		return nil
	}
	return querier
}

// HandleHeartbeat handles POST /api/v1/beacon/heartbeat
// @Summary		Submit beacon heartbeat
// @Description	Receives a heartbeat from a beacon node. Requires JWT authentication with beacon role and mTLS.
// @Tags			Beacon
// @Accept			json
// @Produce		json
// @Param			request	body		models.HeartbeatRequest			true	"Heartbeat data"
// @Success		200		{object}	models.HeartbeatSuccessResponse	"Heartbeat received successfully"
// @Failure		400		{object}	models.ErrorResponse			"Invalid request parameters"
// @Failure		401		{object}	models.ErrorResponse			"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse			"Forbidden (requires beacon role)"
// @Failure		500		{object}	models.ErrorResponse			"Internal server error"
// @Security		BearerAuth
// @Router			/beacon/heartbeat [post]
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
			Message: "Insufficient permissions: beacon role required",
			Details: map[string]interface{}{
				"role":     role,
				"required": "beacon",
			},
		})
		return
	}

	// Parse request body
	var req models.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// Validate node ID format
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_NODE_ID",
			Message: "Invalid node ID format",
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
			Message: "Insufficient permissions: token node ID does not match request",
			Details: map[string]interface{}{
				"token_node_id":   userID,
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
				Message: "Node not found",
				Details: map[string]interface{}{
					"node_id": req.NodeID,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to query node",
			Details: err.Error(),
		})
		return
	}

	// Validate latency range (0-60000ms)
	if req.LatencyMs < 0 || req.LatencyMs > 60000 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidLatency,
			Message: "Latency out of range",
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
			Message: "Packet loss rate out of range",
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
			Message: "Jitter out of range",
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
			Message: "Invalid timestamp format",
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
		Message:   "Heartbeat received successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// CompressedHeartbeatRequest represents a compressed heartbeat payload
type CompressedHeartbeatRequest struct {
	Data     []byte `json:"data" binding:"required"`     // Gzip compressed JSON data
	Checksum uint32 `json:"checksum" binding:"required"` // CRC32 checksum
}

// HandleCompressedHeartbeat handles POST /api/v1/beacon/heartbeat/compressed
// Supports gzip-compressed heartbeat data (FR-4.1.5)
// @Summary		Submit compressed beacon heartbeat
// @Description	Receives a gzip-compressed heartbeat from a beacon node. Requires JWT authentication with beacon role and mTLS.
// @Tags			Beacon
// @Accept			json
// @Produce		json
// @Param			request	body		CompressedHeartbeatRequest		true	"Compressed heartbeat data with CRC32 checksum"
// @Success		200		{object}	models.HeartbeatSuccessResponse	"Compressed heartbeat received successfully"
// @Failure		400		{object}	models.ErrorResponse			"Invalid or corrupted compressed data"
// @Failure		401		{object}	models.ErrorResponse			"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse			"Forbidden (requires beacon role)"
// @Failure		500		{object}	models.ErrorResponse			"Internal server error"
// @Security		BearerAuth
// @Router			/beacon/heartbeat/compressed [post]
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
			Message: "Insufficient permissions: beacon role required",
		})
		return
	}

	// Parse compressed request
	var compReq CompressedHeartbeatRequest
	if err := c.ShouldBindJSON(&compReq); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// Validate compressed data size (max 1MB to prevent memory exhaustion)
	const maxCompressedSize = 1024 * 1024
	if len(compReq.Data) > maxCompressedSize {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_PAYLOAD_TOO_LARGE",
			Message: "Compressed data exceeds maximum size (1MB)",
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
			Message: "Compressed data checksum failed",
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
			Message: "Decompression failed",
			Details: err.Error(),
		})
		return
	}
	defer func() { _ = reader.Close() }()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		RecordCompressionCorruption()
		slog.Error("Failed to read decompressed data", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrCompressionCorrupted,
			Message: "Failed to read decompressed data",
			Details: err.Error(),
		})
		return
	}

	// Parse the decompressed JSON into heartbeat request
	var req models.HeartbeatRequest
	if err := json.Unmarshal(decompressed, &req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// Verify JWT token's user_id matches the requested node_id
	if userID != req.NodeID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "Insufficient permissions: token node ID does not match request",
		})
		return
	}

	// Validate node ID format
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_NODE_ID",
			Message: "Invalid node ID format",
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
				Message: "Node not found",
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
	if err := h.memoryCache.Store(req.NodeID, metricPoint); err != nil {
		slog.Error("Failed to write to memory cache",
			"node_id", req.NodeID,
			"error", err)
	}

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
	_ = h.batchWriter.Write(metricRecord)

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
		Message:   "Compressed heartbeat received successfully",
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
	Type            string `json:"type"` // TCP, UDP, or MTR
	Target          string `json:"target"`
	Port            int    `json:"port"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	Count           int    `json:"count"`
	MaxHops         int    `json:"max_hops,omitempty"`
	PacketSize      int    `json:"packet_size,omitempty"`
}

// BeaconConfigUpdateRequest represents a request to update beacon config
type BeaconConfigUpdateRequest struct {
	Probes          *[]ProbeConfig `json:"probes,omitempty"`
	IntervalSeconds *int           `json:"interval_seconds,omitempty"`
	TimeoutSeconds  *int           `json:"timeout_seconds,omitempty"`
}

// BeaconConfigAckRequest represents a beacon config apply acknowledgement.
type BeaconConfigAckRequest struct {
	NodeID       string `json:"node_id" binding:"required"`
	Version      int    `json:"version" binding:"required"`
	Status       string `json:"status" binding:"required"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// MTRHopRequest represents a single MTR hop reported by a beacon.
type MTRHopRequest struct {
	HopNumber  int     `json:"hop_number"`
	IP         string  `json:"ip"`
	Hostname   string  `json:"hostname,omitempty"`
	ASNumber   string  `json:"as_number,omitempty"`
	Sent       int     `json:"sent"`
	Received   int     `json:"received"`
	LossRate   float64 `json:"loss_rate"`
	LastRTTMs  float64 `json:"last_rtt_ms"`
	AvgRTTMs   float64 `json:"avg_rtt_ms"`
	BestRTTMs  float64 `json:"best_rtt_ms"`
	WorstRTTMs float64 `json:"worst_rtt_ms"`
	StdDevMs   float64 `json:"std_dev_ms"`
	Location   string  `json:"location,omitempty"`
}

// MTRResultRequest represents an MTR result reported by a beacon.
type MTRResultRequest struct {
	NodeID       string          `json:"node_id" binding:"required"`
	ProbeID      string          `json:"probe_id,omitempty"`
	Target       string          `json:"target" binding:"required"`
	TotalHops    int             `json:"total_hops"`
	Hops         []MTRHopRequest `json:"hops" binding:"required"`
	CompletedAt  string          `json:"completed_at" binding:"required"`
	Success      bool            `json:"success"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// BeaconConfigResponse represents beacon config response
type BeaconConfigResponse struct {
	Data      BeaconConfig `json:"data"`
	Message   string       `json:"message"`
	Timestamp string       `json:"timestamp"`
}

// ConfigHistoryEntry represents a config history entry
type ConfigHistoryEntry struct {
	Version   int          `json:"version"`
	Config    BeaconConfig `json:"config"`
	ChangedAt time.Time    `json:"changed_at"`
	ChangedBy string       `json:"changed_by"`
}

// ConfigHistoryResponse represents config history response
type ConfigHistoryResponse struct {
	Data      []ConfigHistoryEntry `json:"data"`
	Message   string               `json:"message"`
	Timestamp string               `json:"timestamp"`
}

// In-memory beacon config store used only when a database-backed querier is not
// wired, mainly for narrow handler tests.
var (
	beaconConfigStore   = make(map[string]*BeaconConfig)
	beaconConfigHistory = make(map[string][]ConfigHistoryEntry)
	beaconConfigMutex   sync.RWMutex
)

// GetBeaconConfig handles GET /api/v1/beacons/:id/config
// @Summary		Get beacon configuration
// @Description	Retrieves the current probe configuration for a beacon.
// @Tags			Beacon
// @Accept			json
// @Produce		json
// @Param			id	path		string				true	"Beacon UUID"
// @Success		200	{object}	BeaconConfigResponse	"Beacon configuration"
// @Failure		400	{object}	models.ErrorResponse	"Invalid beacon ID"
// @Failure		401	{object}	models.ErrorResponse	"Unauthorized"
// @Security		BearerAuth
// @Router			/beacons/{id}/config [get]
func (h *BeaconHandler) GetBeaconConfig(c *gin.Context) {
	beaconID := c.Param("id")

	// Validate beacon exists
	parsedBeaconID, err := uuid.Parse(beaconID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_BEACON_ID",
			Message: "Invalid beacon ID format",
		})
		return
	}

	if querier := h.beaconConfigQuerier(); querier != nil {
		config, err := querier.GetBeaconConfig(c.Request.Context(), parsedBeaconID)
		if err != nil && !errors.Is(err, db.ErrBeaconConfigNotFound) {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    "ERR_INTERNAL_SERVER",
				Message: "Failed to get beacon config",
				Details: err.Error(),
			})
			return
		}
		if errors.Is(err, db.ErrBeaconConfigNotFound) {
			config = defaultDBBeaconConfig(parsedBeaconID)
		}

		c.JSON(http.StatusOK, BeaconConfigResponse{
			Data:      fromDBBeaconConfig(config),
			Message:   "Beacon config retrieved successfully",
			Timestamp: time.Now().Format(time.RFC3339),
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
		Message:   "Beacon config retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateBeaconConfig handles POST /api/v1/beacons/:id/config
// @Summary		Update beacon configuration
// @Description	Updates the probe configuration for a beacon. Requires admin or operator role.
// @Tags			Beacon
// @Accept			json
// @Produce		json
// @Param			id		path		string						true	"Beacon UUID"
// @Param			request	body		BeaconConfigUpdateRequest	true	"Configuration update request"
// @Success		200		{object}	BeaconConfigResponse		"Beacon configuration updated"
// @Failure		400		{object}	models.ErrorResponse		"Invalid request or config validation failed"
// @Failure		401		{object}	models.ErrorResponse		"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse		"Forbidden (requires admin or operator role)"
// @Security		BearerAuth
// @Router			/beacons/{id}/config [post]
func (h *BeaconHandler) UpdateBeaconConfig(c *gin.Context) {
	beaconID := c.Param("id")

	// Validate beacon ID
	parsedBeaconID, err := uuid.Parse(beaconID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_BEACON_ID",
			Message: "Invalid beacon ID format",
		})
		return
	}

	// Parse request
	var req BeaconConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// Validate config
	if err := validateBeaconConfig(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidConfig,
			Message: "Config validation failed",
			Details: err.Error(),
		})
		return
	}

	if querier := h.beaconConfigQuerier(); querier != nil {
		config, err := querier.UpsertBeaconConfig(c.Request.Context(), parsedBeaconID, toDBBeaconConfigUpdate(req, "system"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    "ERR_INTERNAL_SERVER",
				Message: "Failed to update beacon config",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, BeaconConfigResponse{
			Data:      fromDBBeaconConfig(config),
			Message:   "Beacon config updated successfully",
			Timestamp: time.Now().Format(time.RFC3339),
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
		Message:   "Beacon config updated successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetBeaconConfigHistory handles GET /api/v1/beacons/:id/config/history
// @Summary		Get beacon configuration history
// @Description	Retrieves the configuration change history for a beacon (last 50 entries).
// @Tags			Beacon
// @Accept			json
// @Produce		json
// @Param			id	path		string					true	"Beacon UUID"
// @Success		200	{object}	ConfigHistoryResponse	"Configuration history"
// @Failure		400	{object}	models.ErrorResponse	"Invalid beacon ID"
// @Failure		401	{object}	models.ErrorResponse	"Unauthorized"
// @Security		BearerAuth
// @Router			/beacons/{id}/config/history [get]
func (h *BeaconHandler) GetBeaconConfigHistory(c *gin.Context) {
	beaconID := c.Param("id")

	// Validate beacon ID
	parsedBeaconID, err := uuid.Parse(beaconID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_BEACON_ID",
			Message: "Invalid beacon ID format",
		})
		return
	}

	if querier := h.beaconConfigQuerier(); querier != nil {
		dbHistory, err := querier.GetBeaconConfigHistory(c.Request.Context(), parsedBeaconID, 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    "ERR_INTERNAL_SERVER",
				Message: "Failed to get beacon config history",
				Details: err.Error(),
			})
			return
		}

		history := make([]ConfigHistoryEntry, 0, len(dbHistory))
		for _, entry := range dbHistory {
			history = append(history, ConfigHistoryEntry{
				Version:   entry.Version,
				Config:    fromDBBeaconConfig(&entry.Config),
				ChangedAt: entry.ChangedAt,
				ChangedBy: entry.ChangedBy,
			})
		}

		c.JSON(http.StatusOK, ConfigHistoryResponse{
			Data:      history,
			Message:   "Config history retrieved successfully",
			Timestamp: time.Now().Format(time.RFC3339),
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
		Message:   "Config history retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// BatchConfigUpdateRequest represents a batch config update request
type BatchConfigUpdateRequest struct {
	BeaconIDs []string                  `json:"beacon_ids" binding:"required"`
	Config    BeaconConfigUpdateRequest `json:"config" binding:"required"`
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
// @Summary		Batch update beacon group configuration
// @Description	Applies a configuration update to all beacons in a group. Requires admin or operator role.
// @Tags			Beacon
// @Accept			json
// @Produce		json
// @Param			gid		path		string						true	"Beacon group ID"
// @Param			request	body		BatchConfigUpdateRequest	true	"Batch configuration update request"
// @Success		200		{object}	BatchConfigUpdateResponse	"Batch update completed"
// @Failure		400		{object}	models.ErrorResponse		"Invalid request or config validation failed"
// @Failure		401		{object}	models.ErrorResponse		"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse		"Forbidden (requires admin or operator role)"
// @Security		BearerAuth
// @Router			/beacon-groups/{gid}/config [post]
func (h *BeaconHandler) BatchUpdateBeaconGroupConfig(c *gin.Context) {
	groupID := c.Param("gid")

	// Parse request
	var req BatchConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// Validate config
	if err := validateBeaconConfig(&req.Config); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidConfig,
			Message: "Config validation failed",
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

	if querier := h.beaconConfigQuerier(); querier != nil {
		for _, beaconID := range req.BeaconIDs {
			parsedBeaconID, err := uuid.Parse(beaconID)
			if err != nil {
				result.FailedCount++
				result.FailedIDs = append(result.FailedIDs, beaconID)
				result.Errors = append(result.Errors, "invalid beacon ID: "+beaconID)
				continue
			}

			if _, err := querier.UpsertBeaconConfig(c.Request.Context(), parsedBeaconID, toDBBeaconConfigUpdate(req.Config, "batch:"+groupID)); err != nil {
				result.FailedCount++
				result.FailedIDs = append(result.FailedIDs, beaconID)
				result.Errors = append(result.Errors, err.Error())
				continue
			}

			result.SuccessCount++
		}

		c.JSON(http.StatusOK, BatchConfigUpdateResponse{
			Data:      result,
			Message:   "Batch config update completed",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// Hold lock for entire batch operation to prevent race conditions
	beaconConfigMutex.Lock()
	defer beaconConfigMutex.Unlock()

	// Apply config to each beacon
	for _, beaconID := range req.BeaconIDs {
		// Validate beacon ID format
		if _, err := uuid.Parse(beaconID); err != nil {
			result.FailedCount++
			result.FailedIDs = append(result.FailedIDs, beaconID)
			result.Errors = append(result.Errors, "invalid beacon ID: "+beaconID)
			continue
		}

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

		result.SuccessCount++
	}

	c.JSON(http.StatusOK, BatchConfigUpdateResponse{
		Data:      result,
		Message:   "Batch config update completed",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// AcknowledgeBeaconConfig handles POST /api/v1/beacon/config/ack.
func (h *BeaconHandler) AcknowledgeBeaconConfig(c *gin.Context) {
	var req BeaconConfigAckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_NODE_ID",
			Message: "Invalid node ID format",
		})
		return
	}
	if req.Version < 1 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidConfig,
			Message: "Config version must be >= 1",
		})
		return
	}
	if req.Status != "applied" && req.Status != "failed" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidConfig,
			Message: "Config ack status must be applied or failed",
		})
		return
	}

	if userID, err := middleware.GetUserID(c); err == nil && userID != "" && userID != req.NodeID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "Insufficient permissions: token node ID does not match request",
			Details: map[string]interface{}{
				"token_node_id":   userID,
				"request_node_id": req.NodeID,
			},
		})
		return
	}

	querier := h.beaconConfigQuerier()
	if querier == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Beacon config acknowledgement received",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	if err := querier.AcknowledgeBeaconConfig(c.Request.Context(), nodeID, req.Version, req.Status, req.ErrorMessage); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to acknowledge beacon config",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Beacon config acknowledgement received",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// HandleMTRResult handles POST /api/v1/beacon/mtr.
func (h *BeaconHandler) HandleMTRResult(c *gin.Context) {
	var req MTRResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_NODE_ID",
			Message: "Invalid node ID format",
		})
		return
	}

	if userID, err := middleware.GetUserID(c); err == nil && userID != "" && userID != req.NodeID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    ErrUnauthorizedNode,
			Message: "Insufficient permissions: token node ID does not match request",
			Details: map[string]interface{}{
				"token_node_id":   userID,
				"request_node_id": req.NodeID,
			},
		})
		return
	}

	completedAt, err := time.Parse(time.RFC3339, req.CompletedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_TIMESTAMP",
			Message: "Invalid completed_at format",
			Details: "Must be RFC3339 format",
		})
		return
	}

	hops := make([]db.MTRHop, 0, len(req.Hops))
	for _, hop := range req.Hops {
		hops = append(hops, db.MTRHop{
			HopNumber:  hop.HopNumber,
			IP:         hop.IP,
			Hostname:   hop.Hostname,
			ASNumber:   hop.ASNumber,
			Sent:       hop.Sent,
			Received:   hop.Received,
			LossRate:   hop.LossRate,
			LastRTTMs:  hop.LastRTTMs,
			AvgRTTMs:   hop.AvgRTTMs,
			BestRTTMs:  hop.BestRTTMs,
			WorstRTTMs: hop.WorstRTTMs,
			StdDevMs:   hop.StdDevMs,
			Location:   hop.Location,
		})
	}

	querier := h.mtrResultsQuerier()
	if querier == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "MTR result received",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	result, err := querier.SaveMTRResult(c.Request.Context(), db.MTRResultInput{
		NodeID:       nodeID,
		ProbeID:      req.ProbeID,
		Target:       req.Target,
		Success:      req.Success,
		TotalHops:    req.TotalHops,
		Hops:         hops,
		CompletedAt:  completedAt,
		ErrorMessage: req.ErrorMessage,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to save MTR result",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         result.ID.String(),
			"node_id":    result.NodeID.String(),
			"target":     result.Target,
			"total_hops": result.TotalHops,
		},
		"message":   "MTR result received",
		"timestamp": time.Now().Format(time.RFC3339),
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
			if probe.Type != "TCP" && probe.Type != "UDP" && probe.Type != "MTR" {
				return errors.New("invalid probe type: must be TCP, UDP, or MTR")
			}
			if probe.Target == "" {
				return errors.New("probe target cannot be empty")
			}
			if probe.Type != "MTR" && (probe.Port < 1 || probe.Port > 65535) {
				return errors.New("probe port must be between 1 and 65535")
			}
			if probe.Type == "MTR" {
				if probe.MaxHops != 0 && (probe.MaxHops < 1 || probe.MaxHops > 64) {
					return errors.New("mtr max_hops must be between 1 and 64")
				}
				if probe.PacketSize != 0 && (probe.PacketSize < 64 || probe.PacketSize > 1500) {
					return errors.New("mtr packet_size must be between 64 and 1500")
				}
			}
		}
	}

	return nil
}

func defaultDBBeaconConfig(beaconID uuid.UUID) *db.BeaconConfig {
	return &db.BeaconConfig{
		BeaconID:        beaconID,
		Probes:          []db.BeaconProbeConfig{},
		IntervalSeconds: 60,
		TimeoutSeconds:  5,
		Version:         1,
		UpdatedAt:       time.Now(),
	}
}

func fromDBBeaconConfig(config *db.BeaconConfig) BeaconConfig {
	if config == nil {
		return BeaconConfig{
			Probes:          []ProbeConfig{},
			IntervalSeconds: 60,
			TimeoutSeconds:  5,
			UpdatedAt:       time.Now(),
			Version:         1,
		}
	}

	probes := make([]ProbeConfig, 0, len(config.Probes))
	for _, probe := range config.Probes {
		probes = append(probes, ProbeConfig{
			ID:              probe.ID,
			Type:            probe.Type,
			Target:          probe.Target,
			Port:            probe.Port,
			IntervalSeconds: probe.IntervalSeconds,
			TimeoutSeconds:  probe.TimeoutSeconds,
			Count:           probe.Count,
			MaxHops:         probe.MaxHops,
			PacketSize:      probe.PacketSize,
		})
	}

	return BeaconConfig{
		Probes:          probes,
		IntervalSeconds: config.IntervalSeconds,
		TimeoutSeconds:  config.TimeoutSeconds,
		UpdatedAt:       config.UpdatedAt,
		Version:         config.Version,
	}
}

func toDBBeaconConfigUpdate(req BeaconConfigUpdateRequest, changedBy string) db.BeaconConfigUpdate {
	update := db.BeaconConfigUpdate{
		IntervalSeconds: req.IntervalSeconds,
		TimeoutSeconds:  req.TimeoutSeconds,
		ChangedBy:       changedBy,
	}

	if req.Probes != nil {
		probes := make([]db.BeaconProbeConfig, 0, len(*req.Probes))
		for _, probe := range *req.Probes {
			probes = append(probes, db.BeaconProbeConfig{
				ID:              probe.ID,
				Type:            probe.Type,
				Target:          probe.Target,
				Port:            probe.Port,
				IntervalSeconds: probe.IntervalSeconds,
				TimeoutSeconds:  probe.TimeoutSeconds,
				Count:           probe.Count,
				MaxHops:         probe.MaxHops,
				PacketSize:      probe.PacketSize,
			})
		}
		update.Probes = &probes
	}

	return update
}

// GetConfigPreview handles POST /api/v1/beacons/:id/config/preview
// Returns validation result without applying changes
// @Summary		Preview beacon configuration changes
// @Description	Validates a configuration update and returns a preview without applying changes.
// @Tags			Beacon
// @Accept			json
// @Produce		json
// @Param			id		path		string						true	"Beacon UUID"
// @Param			request	body		BeaconConfigUpdateRequest	true	"Configuration to preview"
// @Success		200		{object}	map[string]interface{}		"Configuration preview result"
// @Failure		400		{object}	models.ErrorResponse		"Invalid request"
// @Failure		401		{object}	models.ErrorResponse		"Unauthorized"
// @Security		BearerAuth
// @Router			/beacons/{id}/config/preview [post]
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

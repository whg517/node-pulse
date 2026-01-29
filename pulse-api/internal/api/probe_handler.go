package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
	"github.com/kevin/node-pulse/pulse-api/pkg/middleware"
)

var (
	ErrProbeTypeRequired      = "ERR_PROBE_TYPE_REQUIRED"
	ErrProbeTypeInvalid       = "ERR_PROBE_TYPE_INVALID"
	ErrProbeTargetRequired    = "ERR_PROBE_TARGET_REQUIRED"
	ErrProbeTargetInvalid     = "ERR_PROBE_TARGET_INVALID"
	ErrProbePortRequired      = "ERR_PROBE_PORT_REQUIRED"
	ErrProbePortInvalid       = "ERR_PROBE_PORT_INVALID"
	ErrProbeIntervalRequired  = "ERR_PROBE_INTERVAL_REQUIRED"
	ErrProbeIntervalInvalid   = "ERR_PROBE_INTERVAL_INVALID"
	ErrProbeCountRequired     = "ERR_PROBE_COUNT_REQUIRED"
	ErrProbeCountInvalid      = "ERR_PROBE_COUNT_INVALID"
	ErrProbeTimeoutRequired   = "ERR_PROBE_TIMEOUT_REQUIRED"
	ErrProbeTimeoutInvalid    = "ERR_PROBE_TIMEOUT_INVALID"
	ErrProbeNotFound          = "ERR_PROBE_NOT_FOUND"
	ErrProbeNodeIDRequired    = "ERR_PROBE_NODE_ID_REQUIRED"
	ErrProbeNodeIDInvalid     = "ERR_PROBE_NODE_ID_INVALID"
	ErrProbeNodeNotFound      = "ERR_PROBE_NODE_NOT_FOUND"
)

// ProbeHandler handles probe API requests
type ProbeHandler struct {
	probeQuerier db.ProbesQuerier
	nodeQuerier  db.NodesQuerier
}

// NewProbeHandler creates a new ProbeHandler
func NewProbeHandler(probeQuerier db.ProbesQuerier, nodeQuerier db.NodesQuerier) *ProbeHandler {
	return &ProbeHandler{
		probeQuerier: probeQuerier,
		nodeQuerier:  nodeQuerier,
	}
}

// CreateProbeHandler handles POST /api/v1/probes
func (h *ProbeHandler) CreateProbeHandler(c *gin.Context) {
	// Parse request body
	var req models.CreateProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Normalize type to uppercase for case-insensitive validation
	req.Type = strings.ToUpper(req.Type)

	// Validate probe type
	if req.Type != "TCP" && req.Type != "UDP" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrProbeTypeInvalid,
			Message: "探测类型无效（必须是 TCP 或 UDP）",
			Details: map[string]interface{}{
				"field":  "type",
				"value":  req.Type,
			},
		})
		return
	}

	// Validate node_id exists
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrProbeNodeIDInvalid,
			Message: "节点 ID 格式无效",
			Details: map[string]interface{}{
				"field":  "node_id",
				"value":  req.NodeID,
				"error":  err.Error(),
			},
		})
		return
	}

	// Check if node exists
	ctx := context.Background()
	_, err = h.nodeQuerier.GetNodeByID(ctx, nodeID)
	if err != nil {
		if errors.Is(err, db.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrProbeNodeNotFound,
				Message: "节点不存在",
				Details: map[string]interface{}{
					"node_id": req.NodeID,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点查询失败",
			Details: err.Error(),
		})
		return
	}

	// Validate target (IP or domain)
	if !isValidTarget(req.Target) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrProbeTargetInvalid,
			Message: "目标地址无效（必须是有效的 IP 地址或域名）",
			Details: map[string]interface{}{
				"field": "target",
				"value": req.Target,
			},
		})
		return
	}

	// Generate UUID for new probe
	probeID := uuid.New()

	// Create probe in database
	err = h.probeQuerier.CreateProbe(
		ctx,
		probeID,
		nodeID,
		req.Type,
		req.Target,
		req.Port,
		req.IntervalSeconds,
		req.Count,
		req.TimeoutSeconds,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "探测配置创建失败",
			Details: err.Error(),
		})
		return
	}

	// Fetch created probe from database to return
	probe, err := h.probeQuerier.GetProbeByID(ctx, probeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "探测配置创建失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.CreateProbeResponse{
		Data: models.ProbeData{
			Probe: probe,
		},
		Message:   "探测配置创建成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetProbesHandler handles GET /api/v1/probes
func (h *ProbeHandler) GetProbesHandler(c *gin.Context) {
	ctx := context.Background()

	// Parse query parameters
	nodeIDParam := c.Query("node_id")

	var probes []*models.Probe
	var err error

	if nodeIDParam != "" {
		// Filter by node
		nodeID, err := uuid.Parse(nodeIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrProbeNodeIDInvalid,
				Message: "节点 ID 格式无效",
				Details: map[string]interface{}{
					"node_id": nodeIDParam,
				},
			})
			return
		}
		probes, err = h.probeQuerier.GetProbesByNode(ctx, nodeID)
	} else {
		// Get all probes
		probes, err = h.probeQuerier.GetProbes(ctx)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "探测配置列表获取失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GetProbesResponse{
		Data: models.ProbesListData{
			Probes: probes,
		},
		Message:   "探测配置列表获取成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetProbeByIDHandler handles GET /api/v1/probes/:id
func (h *ProbeHandler) GetProbeByIDHandler(c *gin.Context) {
	// Parse UUID from path parameter
	idParam := c.Param("id")
	probeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "无效的探测配置 ID 格式",
			Details: map[string]interface{}{
				"probe_id": idParam,
				"error":    err.Error(),
			},
		})
		return
	}

	// Fetch probe from database
	ctx := context.Background()
	probe, err := h.probeQuerier.GetProbeByID(ctx, probeID)
	if err != nil {
		if errors.Is(err, db.ErrProbeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrProbeNotFound,
				Message: "探测配置不存在",
				Details: map[string]interface{}{
					"probe_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "探测配置查询失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GetProbeResponse{
		Data: models.ProbeData{
			Probe: probe,
		},
		Message:   "探测配置查询成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateProbeHandler handles PUT /api/v1/probes/:id
func (h *ProbeHandler) UpdateProbeHandler(c *gin.Context) {
	// Parse UUID from path parameter
	idParam := c.Param("id")
	probeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "无效的探测配置 ID 格式",
			Details: map[string]interface{}{
				"probe_id": idParam,
				"error":    err.Error(),
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate target if provided
	if req.Target != nil && !isValidTarget(*req.Target) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrProbeTargetInvalid,
			Message: "目标地址无效（必须是有效的 IP 地址或域名）",
			Details: map[string]interface{}{
				"field": "target",
				"value": *req.Target,
			},
		})
		return
	}

	// Build updates map (only include non-nil values)
	updates := make(map[string]interface{})
	if req.Type != nil {
		// Normalize type to uppercase for case-insensitive validation
		normalizedType := strings.ToUpper(*req.Type)

		// Validate probe type
		if normalizedType != "TCP" && normalizedType != "UDP" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrProbeTypeInvalid,
				Message: "探测类型无效（必须是 TCP 或 UDP）",
				Details: map[string]interface{}{
					"field":  "type",
					"value":  normalizedType,
				},
			})
			return
		}

		updates["type"] = normalizedType
	}
	if req.Target != nil {
		updates["target"] = *req.Target
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.IntervalSeconds != nil {
		updates["interval_seconds"] = *req.IntervalSeconds
	}
	if req.Count != nil {
		updates["count"] = *req.Count
	}
	if req.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *req.TimeoutSeconds
	}

	// Update probe in database
	ctx := context.Background()
	err = h.probeQuerier.UpdateProbe(ctx, probeID, updates)
	if err != nil {
		if errors.Is(err, db.ErrProbeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrProbeNotFound,
				Message: "探测配置不存在",
				Details: map[string]interface{}{
					"probe_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "探测配置更新失败",
			Details: err.Error(),
		})
		return
	}

	// Fetch updated probe from database to return
	probe, err := h.probeQuerier.GetProbeByID(ctx, probeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "探测配置更新失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UpdateProbeResponse{
		Data: models.ProbeData{
			Probe: probe,
		},
		Message:   "探测配置更新成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DeleteProbeHandler handles DELETE /api/v1/probes/:id
func (h *ProbeHandler) DeleteProbeHandler(c *gin.Context) {
	// Parse UUID from path parameter
	idParam := c.Param("id")
	probeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "无效的探测配置 ID 格式",
			Details: map[string]interface{}{
				"probe_id": idParam,
				"error":    err.Error(),
			},
		})
		return
	}

	// Check for confirmation parameter (required by AC)
	confirm := c.Query("confirm")
	if confirm != "true" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_CONFIRMATION_REQUIRED",
			Message: "删除操作需要确认",
			Details: map[string]interface{}{
				"probe_id":     idParam,
				"action":       "DELETE",
				"requirement":  "Add ?confirm=true query parameter to confirm deletion",
			},
		})
		return
	}

	// Delete probe from database
	ctx := context.Background()
	err = h.probeQuerier.DeleteProbe(ctx, probeID)
	if err != nil {
		if errors.Is(err, db.ErrProbeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrProbeNotFound,
				Message: "探测配置不存在",
				Details: map[string]interface{}{
					"probe_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "探测配置删除失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DeleteProbeResponse{
		Message:   "探测配置删除成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// isValidTarget validates if target is a valid IP address or domain name
func isValidTarget(target string) bool {
	// Check if it's a valid IPv4 or IPv6 address
	if parsedIP := net.ParseIP(target); parsedIP != nil {
		return true
	}

	// Domain validation
	if target == "" || len(target) > 255 {
		return false
	}

	// Cannot start or end with hyphen or dot
	if target[0] == '-' || target[0] == '.' || target[len(target)-1] == '-' || target[len(target)-1] == '.' {
		return false
	}

	// Split by dots and validate each label
	labels := strings.Split(target, ".")
	if len(labels) < 2 {
		return false // Must have at least 2 labels (e.g., example.com)
	}

	for _, label := range labels {
		// Each label must be 1-63 characters
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Label cannot start or end with hyphen
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}

		// Label can only contain alphanumeric and hyphens
		for i, c := range label {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
			// Cannot have consecutive hyphens
			if c == '-' && i > 0 && label[i-1] == '-' {
				return false
			}
		}
	}

	// TLD (last label) must be at least 2 characters and all letters
	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return false
	}
	for _, c := range tld {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}

	return true
}

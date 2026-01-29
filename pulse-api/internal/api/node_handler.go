package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
	"github.com/kevin/node-pulse/pulse-api/pkg/middleware"
)

var (
	ErrNodeNameRequired    = "ERR_NODE_NAME_REQUIRED"
	ErrNodeIPRequired      = "ERR_NODE_IP_REQUIRED"
	ErrNodeIPInvalid       = "ERR_NODE_IP_INVALID"
	ErrNodeRegionRequired = "ERR_NODE_REGION_REQUIRED"
	ErrNodeRegionInvalid  = "ERR_NODE_REGION_INVALID"
	ErrNodeNotFound       = "ERR_NODE_NOT_FOUND"
	ErrPermissionDenied    = "ERR_PERMISSION_DENIED"
	ErrDatabaseError      = "ERR_DATABASE_ERROR"
)

// NodeHandler handles node API requests
type NodeHandler struct {
	nodeQuerier db.NodesQuerier
}

// NewNodeHandler creates a new NodeHandler
func NewNodeHandler(nodeQuerier db.NodesQuerier) *NodeHandler {
	return &NodeHandler{
		nodeQuerier: nodeQuerier,
	}
}

// CreateNodeHandler handles POST /api/v1/nodes
func (h *NodeHandler) CreateNodeHandler(c *gin.Context) {
	// RBAC is handled by middleware - only admin/operator can reach this handler

	// Parse request body
	var req models.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrNodeNameRequired,
			Message: "节点名称不能为空",
			Details: map[string]interface{}{
				"field":      "name",
				"constraint": "NOT NULL",
			},
		})
		return
	}

	if req.IP == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrNodeIPRequired,
			Message: "节点 IP 地址不能为空",
			Details: map[string]interface{}{
				"field":      "ip",
				"constraint": "NOT NULL",
			},
		})
		return
	}

	if req.Region == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrNodeRegionRequired,
			Message: "地区标签不能为空",
			Details: map[string]interface{}{
				"field":      "region",
				"constraint": "NOT NULL",
			},
		})
		return
	}

	// Validate IP format (basic IPv4 validation)
	if !isValidIPv4(req.IP) && !isValidIPv6(req.IP) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrNodeIPInvalid,
			Message: "节点 IP 地址格式无效",
			Details: map[string]interface{}{
				"field":  "ip",
				"value": req.IP,
			},
		})
		return
	}

	// Validate name length
	if len(req.Name) > 255 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "节点名称长度不能超过 255 个字符",
			Details: map[string]interface{}{
				"field":   "name",
				"value":   req.Name,
				"max":     255,
			},
		})
		return
	}

	ctx := context.Background()

	// Check for duplicate node (name+ip combination) using efficient query
	existingNode, err := h.nodeQuerier.GetNodeByNameAndIP(ctx, req.Name, req.IP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点查询失败",
			Details: err.Error(),
		})
		return
	}

	// If duplicate found, update and return existing node
	if existingNode != nil {
		// Convert string ID to UUID for update
		existingNodeID, err := uuid.Parse(existingNode.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    ErrDatabaseError,
				Message: "节点 ID 格式错误",
				Details: err.Error(),
			})
			return
		}

		// Build updates map
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.IP != "" {
			updates["ip"] = req.IP
		}
		if req.Region != "" {
			updates["region"] = req.Region
		}
		if req.Tags != nil {
			updates["tags"] = req.Tags
		}

		// Update existing node
		err = h.nodeQuerier.UpdateNode(ctx, existingNodeID, updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    ErrDatabaseError,
				Message: "节点更新失败",
				Details: err.Error(),
			})
			return
		}

		// Fetch updated node from database
		updatedNode, err := h.nodeQuerier.GetNodeByID(ctx, existingNodeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    ErrDatabaseError,
				Message: "节点更新失败",
				Details: err.Error(),
			})
			return
		}

		// Return existing node with 200 status (AC #3)
		c.JSON(http.StatusOK, models.CreateNodeResponse{
			Data: models.CreateNodeData{
				ID:        updatedNode.ID,
				Name:      updatedNode.Name,
				IP:        updatedNode.IP,
				Region:    updatedNode.Region,
				Tags:      updatedNode.Tags,
				CreatedAt: updatedNode.CreatedAt,
				UpdatedAt: updatedNode.UpdatedAt,
			},
			Message:   "节点已存在，已更新信息",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// Generate UUID for new node
	nodeID := uuid.New()

	// Create node in database
	err = h.nodeQuerier.CreateNode(ctx, nodeID, req.Name, req.IP, req.Region, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点创建失败",
			Details: err.Error(),
		})
		return
	}

	// Fetch created node from database to return
	node, err := h.nodeQuerier.GetNodeByID(ctx, nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点创建失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.CreateNodeResponse{
		Data: models.CreateNodeData{
			ID:        node.ID,
			Name:      node.Name,
			IP:        node.IP,
			Region:    node.Region,
			Tags:      node.Tags,
			CreatedAt: node.CreatedAt,
			UpdatedAt: node.UpdatedAt,
		},
		Message:   "节点创建成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetNodesHandler handles GET /api/v1/nodes
func (h *NodeHandler) GetNodesHandler(c *gin.Context) {
	// All roles can view nodes (admin, operator, viewer) - auth is handled by middleware

	// Parse query parameters with validation
	region := c.Query("region")

	// Execute query
	ctx := context.Background()
	var nodes []*models.Node
	var err error

	if region != "" {
		// Filter by region
		nodes, err = h.nodeQuerier.GetNodesByRegion(ctx, region)
	} else {
		// Get all nodes
		nodes, err = h.nodeQuerier.GetNodes(ctx)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点列表获取失败",
			Details: err.Error(),
		})
		return
	}

	// Parse and validate pagination parameters
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	// Validate limit (default: 100, max: 1000)
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Validate offset (default: 0)
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Apply pagination at application level
	totalNodes := len(nodes)
	start := offset
	end := offset + limit

	// Clamp bounds
	if start > totalNodes {
		nodes = []*models.Node{}
	} else if end > totalNodes {
		nodes = nodes[start:]
	} else {
		nodes = nodes[start:end]
	}

	c.JSON(http.StatusOK, models.GetNodesResponse{
		Data: models.GetNodesData{
			Nodes: nodes,
		},
		Message:   "节点列表获取成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetNodeByIDHandler handles GET /api/v1/nodes/:id
func (h *NodeHandler) GetNodeByIDHandler(c *gin.Context) {
	// All roles can view nodes (admin, operator, viewer) - auth is handled by middleware

	// Parse UUID from path parameter
	idParam := c.Param("id")
	nodeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "无效的节点 ID 格式",
			Details: map[string]interface{}{
				"node_id": idParam,
				"error":    err.Error(),
			},
		})
		return
	}

	// Fetch node from database
	ctx := context.Background()
	node, err := h.nodeQuerier.GetNodeByID(ctx, nodeID)
	if err != nil {
		if errors.Is(err, db.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrNodeNotFound,
				Message: "节点不存在",
				Details: map[string]interface{}{
					"node_id": idParam,
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

	c.JSON(http.StatusOK, models.GetNodeResponse{
		Data: models.GetNodeData{
			Node: node,
		},
		Message:   "节点查询成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateNodeHandler handles PUT /api/v1/nodes/:id
func (h *NodeHandler) UpdateNodeHandler(c *gin.Context) {
	// RBAC is handled by middleware - only admin/operator can reach this handler

	// Parse UUID from path parameter
	idParam := c.Param("id")
	nodeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "无效的节点 ID 格式",
			Details: map[string]interface{}{
				"node_id": idParam,
				"error":    err.Error(),
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate IP format if provided
	if req.IP != nil && *req.IP != "" {
		if !isValidIPv4(*req.IP) && !isValidIPv6(*req.IP) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrNodeIPInvalid,
				Message: "节点 IP 地址格式无效",
				Details: map[string]interface{}{
					"field":  "ip",
					"value": *req.IP,
				},
			})
			return
		}
	}

	// Validate name length if provided
	if req.Name != nil && len(*req.Name) > 255 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "节点名称长度不能超过 255 个字符",
			Details: map[string]interface{}{
				"field": "name",
				"value": *req.Name,
				"max":   255,
			},
		})
		return
	}

	// Validate region length if provided
	if req.Region != nil && len(*req.Region) > 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "地区标签长度不能超过 100 个字符",
			Details: map[string]interface{}{
				"field":  "region",
				"value": *req.Region,
				"max":   100,
			},
		})
		return
	}

	// Build updates map (only include non-nil values)
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IP != nil {
		updates["ip"] = *req.IP
	}
	if req.Region != nil {
		updates["region"] = *req.Region
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}

	// Update node in database
	ctx := context.Background()
	err = h.nodeQuerier.UpdateNode(ctx, nodeID, updates)
	if err != nil {
		if errors.Is(err, db.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrNodeNotFound,
				Message: "节点不存在",
				Details: map[string]interface{}{
					"node_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点更新失败",
			Details: err.Error(),
		})
		return
	}

	// Fetch updated node from database to return
	node, err := h.nodeQuerier.GetNodeByID(ctx, nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点更新失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UpdateNodeResponse{
		Data: struct {
			Node *models.Node `json:"node"`
		}{
			Node: node,
		},
		Message:   "节点更新成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DeleteNodeHandler handles DELETE /api/v1/nodes/:id
func (h *NodeHandler) DeleteNodeHandler(c *gin.Context) {
	// RBAC is handled by middleware - only admin/operator can reach this handler

	// Parse UUID from path parameter
	idParam := c.Param("id")
	nodeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "无效的节点 ID 格式",
			Details: map[string]interface{}{
				"node_id": idParam,
				"error":    err.Error(),
			},
		})
		return
	}

	// Check for confirmation parameter (required by AC #4)
	confirm := c.Query("confirm")
	if confirm != "true" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_CONFIRMATION_REQUIRED",
			Message: "删除操作需要确认",
			Details: map[string]interface{}{
				"node_id":    idParam,
				"action":     "DELETE",
				"requirement": "Add ?confirm=true query parameter to confirm deletion",
			},
		})
		return
	}

	// Delete node from database
	ctx := context.Background()
	err = h.nodeQuerier.DeleteNode(ctx, nodeID)
	if err != nil {
		if errors.Is(err, db.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrNodeNotFound,
				Message: "节点不存在",
				Details: map[string]interface{}{
					"node_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点删除失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DeleteNodeResponse{
		Message:   "节点删除成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// isValidIPv4 validates IPv4 address format
func isValidIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// Must be a valid IPv4 address (To4() returns non-nil for IPv4)
	if parsed.To4() == nil {
		return false
	}
	// Ensure the string representation matches exactly (reject partial IPs like "192.168")
	return parsed.String() == ip
}

// isValidIPv6 validates IPv6 address format
func isValidIPv6(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// Must be a valid IPv6 address (To16() returns non-nil and To4() returns nil for IPv6)
	if parsed.To4() != nil {
		return false
	}
	if parsed.To16() == nil {
		return false
	}
	// Ensure the string representation matches exactly
	return parsed.String() == ip
}

// GetNodeStatusHandler handles GET /api/v1/nodes/:id/status
func (h *NodeHandler) GetNodeStatusHandler(c *gin.Context) {
	// All roles can view node status (admin, operator, viewer) - auth is handled by middleware

	// Parse UUID from path parameter
	idParam := c.Param("id")
	nodeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "无效的节点 ID 格式",
			Details: map[string]interface{}{
				"node_id": idParam,
				"error":    err.Error(),
			},
		})
		return
	}

	// Fetch node status from database
	// Use request context with timeout (NFR-PERF-003: P99 ≤ 500ms)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
	defer cancel()

	nodeStatus, err := h.nodeQuerier.GetNodeStatus(ctx, nodeID)
	if err != nil {
		if errors.Is(err, db.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrNodeNotFound,
				Message: "节点不存在",
				Details: map[string]interface{}{
					"node_id": idParam,
				},
			})
			return
		}

		// Don't expose internal error details to client (security best practice)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    ErrDatabaseError,
			Message: "节点状态查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.GetNodeStatusResponse{
		Data: models.NodeStatusData{
			Node: nodeStatus,
		},
		Message:   "节点状态查询成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

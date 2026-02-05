package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// AlertHandler handles alert-related HTTP requests
type AlertHandler struct {
	querier db.AlertQuerier
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(querier db.AlertQuerier) *AlertHandler {
	return &AlertHandler{
		querier: querier,
	}
}

// CreateAlertRuleHandler handles POST /api/v1/alerts/rules
func (h *AlertHandler) CreateAlertRuleHandler(c *gin.Context) {
	var req models.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Set default enabled to true if not provided
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}

	alert := &models.Alert{
		Metric:    req.Metric,
		Threshold: req.Threshold,
		Level:     req.Level,
		NodeID:    req.NodeID,
		Enabled:   *req.Enabled,
	}

	if err := h.querier.CreateAlert(c.Request.Context(), alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to create alert rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.CreateAlertResponse{
		Data: models.AlertData{
			Alert: alert,
		},
		Message:   "Alert rule created successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetAlertRulesHandler handles GET /api/v1/alerts/rules
func (h *AlertHandler) GetAlertRulesHandler(c *gin.Context) {
	nodeID := c.Query("node_id")
	var nodeIDPtr *string

	if nodeID != "" {
		nodeIDPtr = &nodeID
	}

	alerts, err := h.querier.GetAlerts(c.Request.Context(), nodeIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to retrieve alert rules",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GetAlertsResponse{
		Data: models.AlertsListData{
			Alerts: alerts,
		},
		Message:   "Alert rules retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetAlertRuleByIDHandler handles GET /api/v1/alerts/rules/:id
func (h *AlertHandler) GetAlertRuleByIDHandler(c *gin.Context) {
	id := c.Param("id")

	alert, err := h.querier.GetAlertByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "alert not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_NOT_FOUND",
				"message": "Alert rule not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to retrieve alert rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GetAlertByIDResponse{
		Data: models.AlertData{
			Alert: alert,
		},
		Message:   "Alert rule retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateAlertRuleHandler handles PUT /api/v1/alerts/rules/:id
func (h *AlertHandler) UpdateAlertRuleHandler(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Validation failed",
			"details": err.Error(),
		})
		return
	}

	alert, err := h.querier.UpdateAlert(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "alert not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_NOT_FOUND",
				"message": "Alert rule not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to update alert rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UpdateAlertResponse{
		Data: models.AlertData{
			Alert: alert,
		},
		Message:   "Alert rule updated successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DeleteAlertRuleHandler handles DELETE /api/v1/alerts/rules/:id
func (h *AlertHandler) DeleteAlertRuleHandler(c *gin.Context) {
	id := c.Param("id")

	if err := h.querier.DeleteAlert(c.Request.Context(), id); err != nil {
		if err.Error() == "alert not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_NOT_FOUND",
				"message": "Alert rule not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to delete alert rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DeleteAlertResponse{
		Message:   "Alert rule deleted successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

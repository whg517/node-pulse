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
// @Summary		Create an alert rule
// @Description	Creates a new alert rule. Requires admin or operator role.
// @Tags			Alerts
// @Accept			json
// @Produce		json
// @Param			request	body		models.CreateAlertRequest	true	"Alert rule creation request"
// @Success		200		{object}	models.CreateAlertResponse	"Alert rule created successfully"
// @Failure		400		{object}	map[string]interface{}		"Validation failed"
// @Failure		401		{object}	map[string]interface{}		"Unauthorized"
// @Failure		403		{object}	map[string]interface{}		"Forbidden (requires admin or operator role)"
// @Failure		500		{object}	map[string]interface{}		"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/rules [post]
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
// @Summary		List alert rules
// @Description	Retrieves all alert rules. Optionally filtered by node_id.
// @Tags			Alerts
// @Accept			json
// @Produce		json
// @Param			node_id	query		string					false	"Filter by node UUID"
// @Success		200		{object}	models.GetAlertsResponse	"List of alert rules"
// @Failure		401		{object}	map[string]interface{}		"Unauthorized"
// @Failure		500		{object}	map[string]interface{}		"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/rules [get]
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
// @Summary		Get alert rule by ID
// @Description	Retrieves an alert rule by its ID.
// @Tags			Alerts
// @Accept			json
// @Produce		json
// @Param			id	path		string						true	"Alert rule ID"
// @Success		200	{object}	models.GetAlertByIDResponse	"Alert rule details"
// @Failure		401	{object}	map[string]interface{}		"Unauthorized"
// @Failure		404	{object}	map[string]interface{}		"Alert rule not found"
// @Failure		500	{object}	map[string]interface{}		"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/rules/{id} [get]
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
// @Summary		Update an alert rule
// @Description	Updates an existing alert rule. Requires admin or operator role.
// @Tags			Alerts
// @Accept			json
// @Produce		json
// @Param			id		path		string						true	"Alert rule ID"
// @Param			request	body		models.UpdateAlertRequest	true	"Alert rule update request"
// @Success		200		{object}	models.UpdateAlertResponse	"Alert rule updated successfully"
// @Failure		400		{object}	map[string]interface{}		"Validation failed"
// @Failure		401		{object}	map[string]interface{}		"Unauthorized"
// @Failure		403		{object}	map[string]interface{}		"Forbidden (requires admin or operator role)"
// @Failure		404		{object}	map[string]interface{}		"Alert rule not found"
// @Failure		500		{object}	map[string]interface{}		"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/rules/{id} [put]
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
// @Summary		Delete an alert rule
// @Description	Deletes an alert rule by its ID. Requires admin or operator role.
// @Tags			Alerts
// @Accept			json
// @Produce		json
// @Param			id	path		string						true	"Alert rule ID"
// @Success		200	{object}	models.DeleteAlertResponse	"Alert rule deleted successfully"
// @Failure		401	{object}	map[string]interface{}		"Unauthorized"
// @Failure		403	{object}	map[string]interface{}		"Forbidden (requires admin or operator role)"
// @Failure		404	{object}	map[string]interface{}		"Alert rule not found"
// @Failure		500	{object}	map[string]interface{}		"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/rules/{id} [delete]
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

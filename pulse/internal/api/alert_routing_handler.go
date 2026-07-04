package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// AlertRoutingHandler serves CRUD for per-webhook alert routing rules (ADR-002).
type AlertRoutingHandler struct {
	repo db.AlertRoutingRulesRepository
}

// NewAlertRoutingHandler constructs the handler. repo may be nil without a DB.
func NewAlertRoutingHandler(repo db.AlertRoutingRulesRepository) *AlertRoutingHandler {
	return &AlertRoutingHandler{repo: repo}
}

type alertRoutingRuleRequest struct {
	Name       string   `json:"name" binding:"required"`
	Enabled    bool     `json:"enabled"`
	Metric     string   `json:"metric"`
	Severities []string `json:"severities"`
	NodeID     string   `json:"node_id"`
	WebhookID  string   `json:"webhook_id" binding:"required"`
}

// ListRoutingRulesHandler handles GET /api/v1/alerts/routing-rules
// @Summary		List alert routing rules
// @Description	Lists routing rules owned by the current user.
// @Tags			Alert Routing
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Failure		401	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/alerts/routing-rules [get]
func (h *AlertRoutingHandler) ListRoutingRulesHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Routing store unavailable"})
		return
	}
	uid := userIDFromContext(c)
	rules, err := h.repo.ListByOwner(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to list routing rules", Details: err.Error()})
		return
	}
	if rules == nil {
		rules = []*models.AlertRoutingRule{}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"rules": rules}, "message": "Routing rules retrieved", "timestamp": time.Now().Format(time.RFC3339)})
}

// CreateRoutingRuleHandler handles POST /api/v1/alerts/routing-rules
// @Summary		Create an alert routing rule
// @Description	Creates a per-webhook routing rule. Requires admin or operator role.
// @Tags			Alert Routing
// @Accept			json
// @Produce		json
// @Param			request	body		alertRoutingRuleRequest	true	"Routing rule fields"
// @Success		201		{object}	map[string]interface{}
// @Failure		400		{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/alerts/routing-rules [post]
func (h *AlertRoutingHandler) CreateRoutingRuleHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Routing store unavailable"})
		return
	}
	var req alertRoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: "Invalid request", Details: err.Error()})
		return
	}
	rule := &models.AlertRoutingRule{
		OwnerUserID: userIDFromContext(c),
		Name:        req.Name,
		Enabled:     req.Enabled,
		Metric:      req.Metric,
		Severities:  req.Severities,
		NodeID:      req.NodeID,
		WebhookID:   req.WebhookID,
	}
	if err := h.repo.Create(c.Request.Context(), rule); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to create routing rule", Details: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rule, "message": "Routing rule created", "timestamp": time.Now().Format(time.RFC3339)})
}

// UpdateRoutingRuleHandler handles PUT /api/v1/alerts/routing-rules/:id
// @Summary		Update an alert routing rule
// @Tags			Alert Routing
// @Accept			json
// @Produce		json
// @Param			id		path		string					true	"Rule ID"
// @Param			request	body		alertRoutingRuleRequest	true	"Routing rule fields"
// @Success		200		{object}	map[string]interface{}
// @Failure		404		{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/alerts/routing-rules/{id} [put]
func (h *AlertRoutingHandler) UpdateRoutingRuleHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Routing store unavailable"})
		return
	}
	var req alertRoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: "Invalid request", Details: err.Error()})
		return
	}
	rule := &models.AlertRoutingRule{
		ID:          c.Param("id"),
		OwnerUserID: userIDFromContext(c),
		Name:        req.Name,
		Enabled:     req.Enabled,
		Metric:      req.Metric,
		Severities:  req.Severities,
		NodeID:      req.NodeID,
		WebhookID:   req.WebhookID,
	}
	if err := h.repo.Update(c.Request.Context(), rule); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Routing rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to update routing rule", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule, "message": "Routing rule updated", "timestamp": time.Now().Format(time.RFC3339)})
}

// DeleteRoutingRuleHandler handles DELETE /api/v1/alerts/routing-rules/:id
// @Summary		Delete an alert routing rule
// @Tags			Alert Routing
// @Produce		json
// @Param			id	path		string	true	"Rule ID"
// @Success		200	{object}	map[string]interface{}
// @Failure		404	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/alerts/routing-rules/{id} [delete]
func (h *AlertRoutingHandler) DeleteRoutingRuleHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Routing store unavailable"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), c.Param("id"), userIDFromContext(c)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Routing rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to delete routing rule", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Routing rule deleted", "timestamp": time.Now().Format(time.RFC3339)})
}

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

// BeaconConfigTemplateHandler serves CRUD for reusable beacon config templates.
type BeaconConfigTemplateHandler struct {
	repo db.BeaconConfigTemplatesRepository
}

// NewBeaconConfigTemplateHandler constructs the handler. repo may be nil when
// the DB pool is unavailable; handlers then return 503.
func NewBeaconConfigTemplateHandler(repo db.BeaconConfigTemplatesRepository) *BeaconConfigTemplateHandler {
	return &BeaconConfigTemplateHandler{repo: repo}
}

type beaconConfigTemplateRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	Probes          []any  `json:"probes" binding:"required"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
}

func userIDFromContext(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ListBeaconConfigTemplatesHandler handles GET /api/v1/beacon-config-templates
// @Summary		List beacon config templates
// @Description	Lists templates owned by the current user. Requires auth.
// @Tags			Beacon Templates
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Failure		401	{object}	models.ErrorResponse
// @Failure		500	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/beacon-config-templates [get]
func (h *BeaconConfigTemplateHandler) ListBeaconConfigTemplatesHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Template store unavailable"})
		return
	}
	uid := userIDFromContext(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Unauthorized"})
		return
	}
	templates, err := h.repo.ListByOwner(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to list templates", Details: err.Error()})
		return
	}
	if templates == nil {
		templates = []*models.BeaconConfigTemplate{}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      gin.H{"templates": templates},
		"message":   "Templates retrieved successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// CreateBeaconConfigTemplateHandler handles POST /api/v1/beacon-config-templates
// @Summary		Create a beacon config template
// @Description	Creates a reusable template owned by the current user. Requires admin or operator role.
// @Tags			Beacon Templates
// @Accept			json
// @Produce		json
// @Param			request	body		beaconConfigTemplateRequest	true	"Template fields"
// @Success		201		{object}	map[string]interface{}
// @Failure		400		{object}	models.ErrorResponse
// @Failure		401		{object}	models.ErrorResponse
// @Failure		403		{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/beacon-config-templates [post]
func (h *BeaconConfigTemplateHandler) CreateBeaconConfigTemplateHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Template store unavailable"})
		return
	}
	uid := userIDFromContext(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Unauthorized"})
		return
	}
	var req beaconConfigTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: "Invalid request", Details: err.Error()})
		return
	}
	t := &models.BeaconConfigTemplate{
		OwnerUserID:     uid,
		Name:            req.Name,
		Description:     req.Description,
		Probes:          req.Probes,
		IntervalSeconds: req.IntervalSeconds,
		TimeoutSeconds:  req.TimeoutSeconds,
	}
	if err := h.repo.Create(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to create template", Details: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"data":      t,
		"message":   "Template created successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// UpdateBeaconConfigTemplateHandler handles PUT /api/v1/beacon-config-templates/:id
// @Summary		Update a beacon config template
// @Description	Updates a template owned by the current user.
// @Tags			Beacon Templates
// @Accept			json
// @Produce		json
// @Param			id		path		string						true	"Template ID"
// @Param			request	body		beaconConfigTemplateRequest	true	"Template fields"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	models.ErrorResponse
// @Failure		404		{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/beacon-config-templates/{id} [put]
func (h *BeaconConfigTemplateHandler) UpdateBeaconConfigTemplateHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Template store unavailable"})
		return
	}
	uid := userIDFromContext(c)
	id := c.Param("id")
	var req beaconConfigTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: "Invalid request", Details: err.Error()})
		return
	}
	t := &models.BeaconConfigTemplate{
		ID:              id,
		OwnerUserID:     uid,
		Name:            req.Name,
		Description:     req.Description,
		Probes:          req.Probes,
		IntervalSeconds: req.IntervalSeconds,
		TimeoutSeconds:  req.TimeoutSeconds,
	}
	if err := h.repo.Update(c.Request.Context(), t); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to update template", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": t, "message": "Template updated successfully", "timestamp": time.Now().Format(time.RFC3339),
	})
}

// DeleteBeaconConfigTemplateHandler handles DELETE /api/v1/beacon-config-templates/:id
// @Summary		Delete a beacon config template
// @Description	Deletes a template owned by the current user.
// @Tags			Beacon Templates
// @Produce		json
// @Param			id	path		string	true	"Template ID"
// @Success		200	{object}	map[string]interface{}
// @Failure		404	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/beacon-config-templates/{id} [delete]
func (h *BeaconConfigTemplateHandler) DeleteBeaconConfigTemplateHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Template store unavailable"})
		return
	}
	uid := userIDFromContext(c)
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id, uid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to delete template", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully", "timestamp": time.Now().Format(time.RFC3339)})
}

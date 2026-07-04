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

// ReportScheduleHandler serves CRUD for recurring report schedules (ADR-001).
type ReportScheduleHandler struct {
	repo db.ReportScheduleRepository
}

// NewReportScheduleHandler constructs the handler.
func NewReportScheduleHandler(repo db.ReportScheduleRepository) *ReportScheduleHandler {
	return &ReportScheduleHandler{repo: repo}
}

type reportScheduleRequest struct {
	Name           string   `json:"name" binding:"required"`
	Frequency      string   `json:"frequency" binding:"required"`
	TimeOfDay      string   `json:"time_of_day"`
	NodeIDs        []string `json:"node_ids" binding:"required"`
	Metrics        []string `json:"metrics"`
	Format         string   `json:"format"`
	RecipientEmail string   `json:"recipient_email"`
	Enabled        bool     `json:"enabled"`
}

func validateSchedule(req *reportScheduleRequest) error {
	switch req.Frequency {
	case "daily", "weekly", "monthly":
	default:
		return errors.New("frequency must be daily, weekly, or monthly")
	}
	if req.TimeOfDay == "" {
		req.TimeOfDay = "09:00"
	}
	if _, err := time.Parse("15:04", req.TimeOfDay); err != nil {
		return errors.New("time_of_day must be HH:MM")
	}
	if req.Format == "" {
		req.Format = "csv"
	}
	if req.Format != "csv" && req.Format != "pdf" {
		return errors.New("format must be csv or pdf")
	}
	return nil
}

// ListReportSchedulesHandler handles GET /api/v1/reports/schedules
// @Summary		List report schedules
// @Description	Lists schedules owned by the current user.
// @Tags			Report Schedules
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Failure		401	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/reports/schedules [get]
func (h *ReportScheduleHandler) ListReportSchedulesHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Schedule store unavailable"})
		return
	}
	list, err := h.repo.ListByOwner(c.Request.Context(), userIDFromContext(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to list schedules", Details: err.Error()})
		return
	}
	if list == nil {
		list = []*models.ReportSchedule{}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"schedules": list}, "message": "Schedules retrieved", "timestamp": time.Now().Format(time.RFC3339)})
}

// CreateReportScheduleHandler handles POST /api/v1/reports/schedules
// @Summary		Create a report schedule
// @Description	Creates a recurring report schedule owned by the current user. Requires admin role.
// @Tags			Report Schedules
// @Accept			json
// @Produce		json
// @Param			request	body		reportScheduleRequest	true	"Schedule fields"
// @Success		201		{object}	map[string]interface{}
// @Failure		400		{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/reports/schedules [post]
func (h *ReportScheduleHandler) CreateReportScheduleHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Schedule store unavailable"})
		return
	}
	var req reportScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: "Invalid request", Details: err.Error()})
		return
	}
	if err := validateSchedule(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: err.Error()})
		return
	}
	s := &models.ReportSchedule{
		OwnerUserID:    userIDFromContext(c),
		Name:           req.Name,
		Frequency:      req.Frequency,
		TimeOfDay:      req.TimeOfDay,
		NodeIDs:        req.NodeIDs,
		Metrics:        req.Metrics,
		Format:         req.Format,
		RecipientEmail: req.RecipientEmail,
		Enabled:        req.Enabled,
	}
	if err := h.repo.Create(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to create schedule", Details: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": s, "message": "Schedule created", "timestamp": time.Now().Format(time.RFC3339)})
}

// UpdateReportScheduleHandler handles PUT /api/v1/reports/schedules/:id
// @Summary		Update a report schedule
// @Tags			Report Schedules
// @Accept			json
// @Produce		json
// @Param			id		path		string					true	"Schedule ID"
// @Param			request	body		reportScheduleRequest	true	"Schedule fields"
// @Success		200		{object}	map[string]interface{}
// @Failure		404		{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/reports/schedules/{id} [put]
func (h *ReportScheduleHandler) UpdateReportScheduleHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Schedule store unavailable"})
		return
	}
	var req reportScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: "Invalid request", Details: err.Error()})
		return
	}
	if err := validateSchedule(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_VALIDATION", Message: err.Error()})
		return
	}
	s := &models.ReportSchedule{
		ID:             c.Param("id"),
		OwnerUserID:    userIDFromContext(c),
		Name:           req.Name,
		Frequency:      req.Frequency,
		TimeOfDay:      req.TimeOfDay,
		NodeIDs:        req.NodeIDs,
		Metrics:        req.Metrics,
		Format:         req.Format,
		RecipientEmail: req.RecipientEmail,
		Enabled:        req.Enabled,
	}
	if err := h.repo.Update(c.Request.Context(), s); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to update schedule", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s, "message": "Schedule updated", "timestamp": time.Now().Format(time.RFC3339)})
}

// DeleteReportScheduleHandler handles DELETE /api/v1/reports/schedules/:id
// @Summary		Delete a report schedule
// @Tags			Report Schedules
// @Produce		json
// @Param			id	path		string	true	"Schedule ID"
// @Success		200	{object}	map[string]interface{}
// @Failure		404	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/reports/schedules/{id} [delete]
func (h *ReportScheduleHandler) DeleteReportScheduleHandler(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "ERR_UNAVAILABLE", Message: "Schedule store unavailable"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), c.Param("id"), userIDFromContext(c)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "ERR_INTERNAL", Message: "Failed to delete schedule", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted", "timestamp": time.Now().Format(time.RFC3339)})
}

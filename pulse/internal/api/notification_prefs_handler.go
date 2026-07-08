package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// NotificationPrefsHandler exposes per-user notification preferences (F4 P2).
type NotificationPrefsHandler struct {
	repo db.NotificationPrefsRepository
}

// NewNotificationPrefsHandler constructs the handler.
func NewNotificationPrefsHandler(repo db.NotificationPrefsRepository) *NotificationPrefsHandler {
	return &NotificationPrefsHandler{repo: repo}
}

// GetNotificationPrefsHandler returns the current user's prefs.
// Defaults are returned (not 404) when the user has no row yet.
//
// @Summary		Get notification preferences
// @Description	Per-user email-notification preferences (F4 Phase 2)
// @Tags			auth,notifications
// @Produce		json
// @Success		200	{object}	models.NotificationPrefs
// @Security		BearerAuth
// @Router			/auth/notification-prefs [get]
func (h *NotificationPrefsHandler) GetNotificationPrefsHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	prefs, err := h.repo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "DB_ERROR", Message: "Failed to load preferences", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": prefs, "timestamp": time.Now().Format(time.RFC3339)})
}

// UpdateNotificationPrefsHandler updates the current user's prefs (partial update).
//
// @Summary		Update notification preferences
// @Description	Update email-notification floor / enable / override address
// @Tags			auth,notifications
// @Accept		json
// @Produce		json
// @Param			body	body		models.UpdateNotificationPrefsRequest	true	"fields to update"
// @Success		200	{object}	models.NotificationPrefs
// @Failure		400	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router			/auth/notification-prefs [put]
func (h *NotificationPrefsHandler) UpdateNotificationPrefsHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.UpdateNotificationPrefsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "INVALID_REQUEST", Message: "Invalid request body", Details: err.Error()})
		return
	}
	if req.MinAlertLevel != nil && !models.IsValidAlertLevel(*req.MinAlertLevel) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_INVALID_LEVEL", Message: "min_alert_level must be P0, P1, or P2"})
		return
	}
	prefs, err := h.repo.Upsert(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "DB_ERROR", Message: "Failed to save preferences", Details: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": prefs, "message": "Preferences updated", "timestamp": time.Now().Format(time.RFC3339)})
}

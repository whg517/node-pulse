package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	webhooksvc "github.com/whg517/node-pulse/pulse/internal/webhook"
)

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	querier db.WebhookQuerier
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(querier db.WebhookQuerier) *WebhookHandler {
	return &WebhookHandler{
		querier: querier,
	}
}

// ValidateHTTPSURL validates that URL is a valid HTTPS URL
func ValidateHTTPSURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme != "https" {
		return errors.New("URL must use HTTPS scheme for security (NFR-SEC-003)")
	}

	if u.Host == "" {
		return errors.New("URL must have a valid host")
	}

	return nil
}

// CreateWebhookHandler handles POST /api/v1/webhooks
// @Summary		Create a webhook configuration
// @Description	Creates a new webhook configuration. URL must use HTTPS. Admin role required.
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Param			request	body		models.CreateWebhookRequest		true	"Webhook creation request"
// @Success		200		{object}	models.CreateWebhookResponse	"Webhook created successfully"
// @Failure		400		{object}	map[string]interface{}			"Validation failed or invalid HTTPS URL"
// @Failure		401		{object}	map[string]interface{}			"Unauthorized"
// @Failure		403		{object}	map[string]interface{}			"Forbidden (requires admin role)"
// @Failure		500		{object}	map[string]interface{}			"Internal server error"
// @Security		BearerAuth
// @Router			/webhooks [post]
func (h *WebhookHandler) CreateWebhookHandler(c *gin.Context) {
	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Validate HTTPS URL
	if err := ValidateHTTPSURL(req.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_URL",
			"message": "URL validation failed",
			"details": err.Error(),
		})
		return
	}

	// Set default enabled to true if not provided
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}

	// Set default event format if not provided
	if req.EventFormat == nil {
		req.EventFormat = models.DefaultEventFormat
	}

	webhook := &models.Webhook{
		URL:         req.URL,
		EventFormat: req.EventFormat,
		Enabled:     *req.Enabled,
	}

	if err := h.querier.CreateWebhook(c.Request.Context(), webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to create webhook configuration",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.CreateWebhookResponse{
		Data: models.WebhookData{
			Webhook: webhook,
		},
		Message:   "Webhook configuration created successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetWebhooksHandler handles GET /api/v1/webhooks
// @Summary		List all webhooks
// @Description	Retrieves all webhook configurations. Admin role required.
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Success		200	{object}	models.GetWebhooksResponse	"List of webhook configurations"
// @Failure		401	{object}	map[string]interface{}		"Unauthorized"
// @Failure		403	{object}	map[string]interface{}		"Forbidden (requires admin role)"
// @Failure		500	{object}	map[string]interface{}		"Internal server error"
// @Security		BearerAuth
// @Router			/webhooks [get]
func (h *WebhookHandler) GetWebhooksHandler(c *gin.Context) {
	webhooks, err := h.querier.GetWebhooks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to retrieve webhook configurations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GetWebhooksResponse{
		Data: models.WebhooksListData{
			Webhooks: webhooks,
		},
		Message:   "Webhook configurations retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// PreviewWebhookEventHandler handles POST /api/v1/webhooks/preview
// @Summary		Preview webhook payload
// @Description	Renders a webhook event format with a sample alert event. Admin role required.
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Param			request	body		models.PreviewWebhookEventRequest	true	"Webhook preview request"
// @Success		200		{object}	models.PreviewWebhookEventResponse	"Rendered webhook payload"
// @Failure		400		{object}	map[string]interface{}				"Validation failed"
// @Failure		401		{object}	map[string]interface{}				"Unauthorized"
// @Failure		403		{object}	map[string]interface{}				"Forbidden (requires admin role)"
// @Security		BearerAuth
// @Router			/webhooks/preview [post]
func (h *WebhookHandler) PreviewWebhookEventHandler(c *gin.Context) {
	var req models.PreviewWebhookEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Validation failed",
			"details": err.Error(),
		})
		return
	}

	payload, err := webhooksvc.RenderAlertEvent(requestBaseURL(c), sampleWebhookAlertEvent(), req.EventFormat)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_EVENT_FORMAT",
			"message": "Invalid webhook event format",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PreviewWebhookEventResponse{
		Data: models.WebhookPreviewData{
			Payload: payload,
		},
		Message:   "Webhook payload preview rendered successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetWebhookByIDHandler handles GET /api/v1/webhooks/:id
// @Summary		Get webhook by ID
// @Description	Retrieves a webhook configuration by its ID. Admin role required.
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Param			id	path		string						true	"Webhook ID"
// @Success		200	{object}	models.UpdateWebhookResponse	"Webhook configuration"
// @Failure		401	{object}	map[string]interface{}		"Unauthorized"
// @Failure		403	{object}	map[string]interface{}		"Forbidden (requires admin role)"
// @Failure		404	{object}	map[string]interface{}		"Webhook not found"
// @Failure		500	{object}	map[string]interface{}		"Internal server error"
// @Security		BearerAuth
// @Router			/webhooks/{id} [get]
func (h *WebhookHandler) GetWebhookByIDHandler(c *gin.Context) {
	id := c.Param("id")

	webhook, err := h.querier.GetWebhookByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "webhook not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_NOT_FOUND",
				"message": "Webhook configuration not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to retrieve webhook configuration",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UpdateWebhookResponse{
		Data: models.WebhookData{
			Webhook: webhook,
		},
		Message:   "Webhook configuration retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateWebhookHandler handles PUT /api/v1/webhooks/:id
// @Summary		Update a webhook
// @Description	Updates an existing webhook configuration. URL must use HTTPS if provided. Admin role required.
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Param			id		path		string						true	"Webhook ID"
// @Param			request	body		models.UpdateWebhookRequest	true	"Webhook update request"
// @Success		200		{object}	models.UpdateWebhookResponse	"Webhook updated successfully"
// @Failure		400		{object}	map[string]interface{}			"Validation failed or invalid HTTPS URL"
// @Failure		401		{object}	map[string]interface{}			"Unauthorized"
// @Failure		403		{object}	map[string]interface{}			"Forbidden (requires admin role)"
// @Failure		404		{object}	map[string]interface{}			"Webhook not found"
// @Failure		500		{object}	map[string]interface{}			"Internal server error"
// @Security		BearerAuth
// @Router			/webhooks/{id} [put]
func (h *WebhookHandler) UpdateWebhookHandler(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Validate HTTPS URL if provided
	if req.URL != nil {
		if err := ValidateHTTPSURL(*req.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_URL",
				"message": "URL validation failed",
				"details": err.Error(),
			})
			return
		}
	}

	webhook, err := h.querier.UpdateWebhook(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "webhook not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_NOT_FOUND",
				"message": "Webhook configuration not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to update webhook configuration",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UpdateWebhookResponse{
		Data: models.WebhookData{
			Webhook: webhook,
		},
		Message:   "Webhook configuration updated successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DeleteWebhookHandler handles DELETE /api/v1/webhooks/:id
// @Summary		Delete a webhook
// @Description	Deletes a webhook configuration by its ID. Admin role required.
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Param			id	path		string						true	"Webhook ID"
// @Success		200	{object}	models.DeleteWebhookResponse	"Webhook deleted successfully"
// @Failure		401	{object}	map[string]interface{}			"Unauthorized"
// @Failure		403	{object}	map[string]interface{}			"Forbidden (requires admin role)"
// @Failure		404	{object}	map[string]interface{}			"Webhook not found"
// @Failure		500	{object}	map[string]interface{}			"Internal server error"
// @Security		BearerAuth
// @Router			/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhookHandler(c *gin.Context) {
	id := c.Param("id")

	if err := h.querier.DeleteWebhook(c.Request.Context(), id); err != nil {
		if err.Error() == "webhook not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_NOT_FOUND",
				"message": "Webhook configuration not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to delete webhook configuration",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DeleteWebhookResponse{
		Message:   "Webhook configuration deleted successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func sampleWebhookAlertEvent() *models.AlertEvent {
	return &models.AlertEvent{
		ID:           "preview-alert-1",
		NodeID:       "preview-node-1",
		Metric:       "latency",
		Threshold:    100,
		CurrentValue: 150,
		Level:        "P1",
		CreatedAt:    time.Date(2026, 6, 21, 9, 30, 0, 0, time.UTC),
	}
}

func requestBaseURL(c *gin.Context) string {
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		host = "localhost"
	}

	return fmt.Sprintf("%s://%s", proto, host)
}

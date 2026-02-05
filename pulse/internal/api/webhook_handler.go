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

// GetWebhookByIDHandler handles GET /api/v1/webhooks/:id
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

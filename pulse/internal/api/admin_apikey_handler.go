package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/auth"
)

// AdminAPIKeyHandler handles admin API key management requests
type AdminAPIKeyHandler struct {
	apiKeyService *auth.APIKeyService
}

// NewAdminAPIKeyHandler creates a new admin API key handler
func NewAdminAPIKeyHandler(apiKeyService *auth.APIKeyService) *AdminAPIKeyHandler {
	return &AdminAPIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// ListAPIKeysHandler handles GET /api/v1/admin/apikeys
// @Summary		List all API keys
// @Description	Retrieves all API keys. Admin role required.
// @Tags			Admin API Keys
// @Accept			json
// @Produce		json
// @Success		200	{object}	map[string]interface{}	"List of API keys"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Failure		500	{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/apikeys [get]
func (h *AdminAPIKeyHandler) ListAPIKeysHandler(c *gin.Context) {
	keys, err := h.apiKeyService.ListAllAPIKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to retrieve API keys",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"keys": keys,
		},
		"message":   "API keys retrieved successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetAPIKeyByIDHandler handles GET /api/v1/admin/apikeys/:id
// @Summary		Get API key by ID
// @Description	Retrieves an API key by its integer ID. Admin role required.
// @Tags			Admin API Keys
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"API key ID"
// @Success		200	{object}	map[string]interface{}	"API key details"
// @Failure		400	{object}	map[string]interface{}	"Invalid ID"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Failure		404	{object}	map[string]interface{}	"API key not found"
// @Security		BearerAuth
// @Router			/admin/apikeys/{id} [get]
func (h *AdminAPIKeyHandler) GetAPIKeyByIDHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_ID",
			"message": "Invalid API key ID",
			"details": "ID must be a valid integer",
		})
		return
	}

	key, err := h.apiKeyService.GetAPIKeyByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "ERR_NOT_FOUND",
			"message": "API key not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"key": key,
		},
		"message":   "API key retrieved successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name             string  `json:"name" binding:"required"`
	UserID           *string `json:"user_id"`
	ServiceAccountID *string `json:"service_account_id"`
}

// CreateAPIKeyHandler handles POST /api/v1/admin/apikeys
// @Summary		Create a new API key
// @Description	Creates a new API key. The full key is only shown once on creation. Admin role required.
// @Tags			Admin API Keys
// @Accept			json
// @Produce		json
// @Param			request	body		CreateAPIKeyRequest		true	"API key creation request"
// @Success		201		{object}	map[string]interface{}	"API key created successfully"
// @Failure		400		{object}	map[string]interface{}	"Validation failed"
// @Failure		401		{object}	map[string]interface{}	"Unauthorized"
// @Failure		403		{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Failure		500		{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/apikeys [post]
func (h *AdminAPIKeyHandler) CreateAPIKeyHandler(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Generate new key with np_live_ prefix
	fullToken, dbKey, err := h.apiKeyService.GenerateAPIKey(c.Request.Context(), req.UserID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to create API key",
			"details": err.Error(),
		})
		return
	}

	// Add np_live_ prefix to the token
	fullToken = "np_live_" + fullToken

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"key":      dbKey,
			"full_key": fullToken, // Only shown on creation
		},
		"message":   "API key created successfully. Save the full_key now, it will not be shown again.",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// RotateAPIKeyHandler handles POST /api/v1/admin/apikeys/:id/rotate
// @Summary		Rotate an API key
// @Description	Rotates an API key, generating a new key value. The old key remains valid for 24 hours. Admin role required.
// @Tags			Admin API Keys
// @Accept			json
// @Produce		json
// @Param			id	path		int						true	"API key ID"
// @Success		200	{object}	map[string]interface{}	"API key rotated successfully"
// @Failure		400	{object}	map[string]interface{}	"Invalid ID"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Failure		500	{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/apikeys/{id}/rotate [post]
func (h *AdminAPIKeyHandler) RotateAPIKeyHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_ID",
			"message": "Invalid API key ID",
			"details": "ID must be a valid integer",
		})
		return
	}

	fullToken, oldKey, newKey, err := h.apiKeyService.RotateAPIKey(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL",
			"message": "Failed to rotate API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"old_key":      oldKey,
			"new_key":      newKey,
			"full_new_key": fullToken, // Only shown on rotation
		},
		"message":   "API key rotated successfully. Old key remains valid for 24 hours. Save the full_new_key now, it will not be shown again.",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// RevokeAPIKeyHandler handles DELETE /api/v1/admin/apikeys/:id
// @Summary		Revoke an API key
// @Description	Revokes (deletes) an API key by its integer ID. Admin role required.
// @Tags			Admin API Keys
// @Accept			json
// @Produce		json
// @Param			id	path		int						true	"API key ID"
// @Success		200	{object}	map[string]interface{}	"API key revoked successfully"
// @Failure		400	{object}	map[string]interface{}	"Invalid ID"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Failure		404	{object}	map[string]interface{}	"API key not found"
// @Security		BearerAuth
// @Router			/admin/apikeys/{id} [delete]
func (h *AdminAPIKeyHandler) RevokeAPIKeyHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_ID",
			"message": "Invalid API key ID",
			"details": "ID must be a valid integer",
		})
		return
	}

	// Admin can revoke any key (pass nil for userID)
	err = h.apiKeyService.RevokeAPIKey(c.Request.Context(), nil, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "ERR_NOT_FOUND",
			"message": "API key not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "API key revoked successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

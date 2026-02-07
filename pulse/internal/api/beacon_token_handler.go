package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// BeaconTokenRequest represents the request to get a beacon token
type BeaconTokenRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// BeaconTokenResponse represents the beacon token response
type BeaconTokenResponse struct {
	Data      BeaconTokenData `json:"data"`
	Message   string          `json:"message"`
	Timestamp string          `json:"timestamp"`
}

// BeaconTokenData contains the access token and expiration
type BeaconTokenData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
	NodeID      string `json:"node_id"`
}

// BeaconTokenHandler handles beacon token API requests
type BeaconTokenHandler struct {
	jwtService *auth.JWTService
	db         BeaconTokenDB
}

// BeaconTokenDB defines the database interface for beacon tokens
type BeaconTokenDB interface {
	GetNodeIDByAPIKey(ctx context.Context, apiKeyHash string) (uuid.UUID, error)
	UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error
}

// NewBeaconTokenHandler creates a new BeaconTokenHandler
func NewBeaconTokenHandler(jwtService *auth.JWTService, db BeaconTokenDB) *BeaconTokenHandler {
	return &BeaconTokenHandler{
		jwtService: jwtService,
		db:         db,
	}
}

// HandleGetToken handles POST /api/v1/beacon/token
// Authenticates beacon using API key and returns JWT access token
func (h *BeaconTokenHandler) HandleGetToken(c *gin.Context) {
	// Parse request body
	var req BeaconTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// Validate API key length (should be at least 32 characters)
	if len(req.APIKey) < 32 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_API_KEY",
			Message: "API Key 格式无效",
			Details: map[string]interface{}{
				"reason": "api_key must be at least 32 characters",
			},
		})
		return
	}

	// Hash API key (SHA-256)
	hasher := sha256.New()
	hasher.Write([]byte(req.APIKey))
	apiKeyHash := hex.EncodeToString(hasher.Sum(nil))

	// Look up node ID by API key hash
	ctx := context.Background()
	nodeID, err := h.db.GetNodeIDByAPIKey(ctx, apiKeyHash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_UNAUTHORIZED",
			Message: "API Key 无效或已过期",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	// Generate JWT access token with role="beacon"
	// Use node ID as user ID for beacon service account
	accessToken, _, err := h.jwtService.GenerateAccessToken(nodeID.String(), "beacon")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Token 生成失败",
			Details: err.Error(),
		})
		return
	}

	// Get JWT expiration time from config
	cfg := config.Get()
	expiresIn := cfg.JWT.AccessTokenExpirationMinutes * 60 // convert to seconds

	// Set security headers to prevent token caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")

	c.JSON(http.StatusOK, BeaconTokenResponse{
		Data: BeaconTokenData{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
			NodeID:      nodeID.String(),
		},
		Message:   "Beacon Token 获取成功",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

const (
	// Security constants (F29: Moved from magic numbers to named constants)
	MaxFailedLoginAttempts    = 5
	AccountLockDuration        = 10 * time.Minute
	MaxSessionsPerUser         = 10
	MaxLoginAttemptsPerMinute  = 5
	MaxRefreshAttemptsPerMinute = 10
	MaxLogoutAttemptsPerMinute = 10
	MaxAPIKeyAttemptsPerMinute = 11
	ConstantAuthDelay          = 150 * time.Millisecond
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	pool              *pgxpool.Pool
	jwtService        *JWTService
	refreshTokenService *RefreshTokenService
	apiKeyService     *APIKeyService
	rateLimiter       *RateLimiter
	auditLogger       *AuditLogger
	accessExpirationMinutes int
	refreshExpirationDays  int
	maxValidityDays        int
	cookieSecure           bool
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	pool *pgxpool.Pool,
	jwtPrivateKey string,
	jwtPublicKey string,
	jwtKeyID string,
	accessExpirationMinutes int,
	refreshExpirationDays int,
	maxValidityDays int,
	cookieSecure bool,
) *AuthHandler {
	jwtService := NewJWTService(jwtPrivateKey, jwtPublicKey, jwtKeyID, accessExpirationMinutes, pool)
	refreshTokenService := NewRefreshTokenService(pool)
	apiKeyService := NewAPIKeyService(pool)
	rateLimiter := NewRateLimiter(pool)
	auditLogger := NewAuditLogger(pool)

	return &AuthHandler{
		pool:                  pool,
		jwtService:           jwtService,
		refreshTokenService:  refreshTokenService,
		apiKeyService:        apiKeyService,
		rateLimiter:          rateLimiter,
		auditLogger:          auditLogger,
		accessExpirationMinutes: accessExpirationMinutes,
		refreshExpirationDays:   refreshExpirationDays,
		maxValidityDays:         maxValidityDays,
		cookieSecure:            cookieSecure,
	}
}

// Login handles user authentication
// @Summary Login
// @Description Authenticate user with username/password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.TokenResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 423 {object} models.ErrorResponse "Account locked"
// @Failure 429 {object} models.ErrorResponse "Rate limit exceeded"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate Content-Type header (prevent content-type confusion attacks)
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, models.ErrorResponse{
			Code:    "UNSUPPORTED_MEDIA_TYPE",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
		})
		return
	}

	// Extract IP for rate limiting
	ipAddress := c.ClientIP()
	rateLimitKey := fmt.Sprintf("ip:%s", ipAddress)

	// Check rate limit (uses MaxLoginAttemptsPerMinute constant)
	allowed, _, resetTime, err := h.rateLimiter.CheckRateLimit(ctx, rateLimitKey, WindowPerMinute, MaxLoginAttemptsPerMinute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "RATE_LIMIT_ERROR",
			Message: "Failed to check rate limit",
		})
		return
	}

	if !allowed {
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Code:    "ERR_RATE_LIMIT_EXCEEDED",
			Message: "Too many login attempts",
			Details: resetTime.Format(time.RFC3339),
		})
		return
	}

	// Get user from database with account lock check (database-side time check)
	var userID, role, passwordHash string
	var failedAttempts int
	err = h.pool.QueryRow(ctx, `
		SELECT user_id, password_hash, role, failed_login_attempts
		FROM users
		WHERE username = $1
		  AND (locked_until IS NULL OR locked_until <= NOW())
	`, req.Username).Scan(&userID, &passwordHash, &role, &failedAttempts)

	if err != nil {
		// User not found, locked, OR account is still locked (DB filters locked accounts)
		// Add constant delay before response (timing attack prevention)
		constantAuthDelay()

		// Generic error message (user enumeration prevention)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_INVALID_CREDENTIALS",
			Message: "Invalid username or password",
		})
		return
	}

	// Verify password using bcrypt directly (constant-time comparison)
		err = VerifyPassword(req.Password, passwordHash)
	if err != nil {
		// Password incorrect - increment failed attempts and log security event
		newAttempts := failedAttempts + 1
		if newAttempts >= MaxFailedLoginAttempts {
			// Lock account for AccountLockDuration
			lockUntil := time.Now().Add(AccountLockDuration)
			_, _ = h.pool.Exec(ctx, `
				UPDATE users
				SET failed_login_attempts = $1, locked_until = $2
				WHERE user_id = $3
			`, newAttempts, lockUntil, userID)
		} else {
			_, _ = h.pool.Exec(ctx, `
				UPDATE users
				SET failed_login_attempts = $1
				WHERE user_id = $2
			`, newAttempts, userID)
		}

		// Log failed login attempt (F30: Add audit logging for failed logins)
		h.logAuditEvent(ctx, "login_failed", &userID, ipAddress, map[string]interface{}{
			"failed_attempts": newAttempts,
			"account_locked":  newAttempts >= 5,
		})

		// Add constant delay to prevent timing attacks
		constantAuthDelay()

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_INVALID_CREDENTIALS",
			Message: "Invalid username or password",
		})
		return
	}

	// Reset failed attempts on successful login
	_, _ = h.pool.Exec(ctx, `
		UPDATE users
		SET failed_login_attempts = 0, locked_until = NULL
		WHERE user_id = $1
	`, userID)

	// Generate tokens
	accessToken, jti, err := h.jwtService.GenerateAccessToken(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate access token",
		})
		return
	}

	userAgent := c.GetHeader("User-Agent")
	refreshToken, _, err := h.refreshTokenService.CreateRefreshToken(
		ctx, userID, userAgent, ipAddress, h.maxValidityDays,
	)
	if err != nil {
		// Log error for debugging (temporary)
		fmt.Printf("[ERROR] Failed to create refresh token for user: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate refresh token",
		})
		return
	}

	// Log successful login
	h.logAuditEvent(ctx, "login", &userID, ipAddress, map[string]interface{}{
		"jti":        jti,
		"user_agent": userAgent,
	})

	// Set refresh token as HTTP-only cookie (must be set before JSON response)
	// Use Lax mode to allow cookies with proxied requests in development
	sameSiteMode := http.SameSiteLaxMode
	if h.cookieSecure {
		sameSiteMode = http.SameSiteStrictMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   h.refreshExpirationDays * 86400, // Convert days to seconds
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: sameSiteMode,
		Path:     "/",
	})

	c.JSON(http.StatusOK, models.LoginResponse{
		Message: "Login successful",
		Data: struct {
			UserID      string `json:"user_id"`
			Username    string `json:"username"`
			Role        string `json:"role"`
			AccessToken string `json:"access_token"`
		}{
			UserID:      userID,
			Username:    req.Username,
			Role:        role,
			AccessToken: accessToken,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Refresh handles token refresh
// @Summary Refresh token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshRequest true "Refresh token"
// @Success 200 {object} models.TokenResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse "Token already used"
// @Failure 429 {object} models.ErrorResponse "Rate limit exceeded"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate Content-Type header (prevent content-type confusion attacks)
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, models.ErrorResponse{
			Code:    "UNSUPPORTED_MEDIA_TYPE",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		// If no token in body, try to get from cookie
		cookie, err := c.Cookie("refresh_token")
		if err != nil || cookie == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Refresh token required in body or cookie",
			})
			return
		}
		req.RefreshToken = cookie
	}

	// Hash the refresh token for rate limit key
	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])
	rateLimitKey := fmt.Sprintf("token:%s", tokenHash)

	// Check rate limit (uses MaxRefreshAttemptsPerMinute constant)
	allowed, _, resetTime, err := h.rateLimiter.CheckRateLimit(ctx, rateLimitKey, WindowPerMinute, MaxRefreshAttemptsPerMinute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "RATE_LIMIT_ERROR",
			Message: "Failed to check rate limit",
		})
		return
	}

	if !allowed {
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Code:    "ERR_RATE_LIMIT_EXCEEDED",
			Message: "Too many refresh requests",
		})
		return
	}

	// Rotate refresh token
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	newRefreshToken, dbToken, err := h.refreshTokenService.RotateRefreshToken(
		ctx, req.RefreshToken, userAgent, ipAddress, h.maxValidityDays,
	)

	if err != nil {
		// Check if token was already used (concurrent request)
		if err.Error() == "token already used" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Code:    "TOKEN_ALREADY_USED",
				Message: "Refresh token has already been used",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "Invalid or expired refresh token",
		})
		return
	}

	// Generate new access token
	// Need to get user role from database
	var role string
	err = h.pool.QueryRow(ctx, `
		SELECT role FROM users WHERE user_id = $1
	`, dbToken.UserID).Scan(&role)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "USER_LOOKUP_FAILED",
			Message: "Failed to lookup user",
		})
		return
	}

	accessToken, jti, err := h.jwtService.GenerateAccessToken(dbToken.UserID.String(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate access token",
		})
		return
	}

	// Log token refresh
	userIDStr := dbToken.UserID.String()
	h.logAuditEvent(ctx, "token_refresh", &userIDStr, ipAddress, map[string]interface{}{
		"jti":        jti,
		"user_agent": userAgent,
	})

	// Set refresh token as HTTP-only cookie (secure, not exposed in response body)
	// Use Lax mode to allow cookies with proxied requests in development
	sameSiteMode := http.SameSiteLaxMode
	if h.cookieSecure {
		sameSiteMode = http.SameSiteStrictMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		MaxAge:   h.refreshExpirationDays * 86400, // Convert days to seconds
		HttpOnly: true,                          // Prevent XSS
		Secure:   h.cookieSecure,                // HTTPS only (configurable)
		SameSite: sameSiteMode,                  // Lax for dev, Strict for prod
		Path:     "/",
	})

	// Return response in same format as Login for frontend consistency
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"data": gin.H{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   h.accessExpirationMinutes * 60,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Logout handles user logout
// @Summary Logout
// @Description Logout user and revoke tokens
// @Tags auth
// @Produce json
// @Security Bearer
// @Success 200
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// Rate limit logout requests per user to prevent DoS
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authorization required",
		})
		return
	}

	rateLimitKey := fmt.Sprintf("logout:%s", userID)
	// Rate limit logout requests (uses MaxLogoutAttemptsPerMinute constant)
	allowed, _, resetTime, err := h.rateLimiter.CheckRateLimit(ctx, rateLimitKey, WindowPerMinute, MaxLogoutAttemptsPerMinute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "RATE_LIMIT_ERROR",
			Message: "Failed to check rate limit",
		})
		return
	}

	if !allowed {
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Code:    "ERR_RATE_LIMIT_EXCEEDED",
			Message: "Too many logout requests",
		})
		return
	}

	// Get JTI from access token for blacklist
	accessToken := c.GetHeader("Authorization")
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	jti, err := h.jwtService.GetJTI(accessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_TOKEN",
			Message: "Failed to parse token",
		})
		return
	}

	// Add access token JTI to blacklist
	expiresAt := time.Now().Add(time.Duration(h.accessExpirationMinutes) * time.Minute)
	_, err = h.pool.Exec(ctx, `
		INSERT INTO token_blacklist (jti, revoked_at, expires_at)
		VALUES ($1, NOW(), $2)
	`, jti, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "BLACKLIST_FAILED",
			Message: "Failed to revoke access token",
		})
		return
	}

	// Revoke refresh token (if provided in body)
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.ShouldBindJSON(&req)

	if req.RefreshToken != "" {
		// Hash and revoke
		hash := sha256.Sum256([]byte(req.RefreshToken))
		tokenHash := hex.EncodeToString(hash[:])

		_, err = h.pool.Exec(ctx, `
			UPDATE refresh_tokens
			SET revoked_at = NOW(), updated_at = NOW()
			WHERE token_id = $1 AND user_id = $2
		`, tokenHash, userID)

		// Don't fail if refresh token revocation fails
		_ = err
	}

	// Log logout
	h.logAuditEvent(ctx, "logout", &userID, c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// GetMe returns the current authenticated user's information
// @Summary Get current user
// @Description Get information about the currently authenticated user
// @Tags auth
// @Produce json
// @Security Bearer
// @Success 200 {object} models.GetMeResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user from context (set by JWT middleware)
	userID := c.GetString("user_id")
	role := c.GetString("role")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	// Query database for username
	var username string
	err := h.pool.QueryRow(ctx, `
		SELECT username FROM users WHERE user_id = $1
	`, userID).Scan(&username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "USER_LOOKUP_FAILED",
			Message: "Failed to retrieve user information",
		})
		return
	}

	c.JSON(http.StatusOK, models.GetMeResponse{
		Message: "Success",
		Data: struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		}{
			UserID:   userID,
			Username: username,
			Role:     role,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetSessions returns user's active sessions
// @Summary Get sessions
// @Description Get all active refresh tokens for current user
// @Tags auth
// @Produce json
// @Security Bearer
// @Success 200 {array} models.SessionResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authorization required",
		})
		return
	}

	tokens, err := h.refreshTokenService.GetUserRefreshTokens(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "SESSIONS_FETCH_FAILED",
			Message: "Failed to fetch sessions",
		})
		return
	}

	// Convert to response format
	sessions := make([]models.SessionResponse, len(tokens))
	for i, token := range tokens {
		var ipAddress *string
		if token.IPAddress.IsValid() {
			ipStr := token.IPAddress.String()
			ipAddress = &ipStr
		}
		sessions[i] = models.SessionResponse{
			SessionID:     token.TokenID.String(),
			CreatedAt:     token.CreatedAt.Time.Format(time.RFC3339),
			ExpiresAt:     token.ExpiresAt.Time.Format(time.RFC3339),
			MaxValidUntil: token.MaxValidUntil.Time.Format(time.RFC3339),
			UserAgent:     token.UserAgent,
			IPAddress:     ipAddress,
		}
	}

	c.JSON(http.StatusOK, sessions)
}

// DeleteSession revokes a specific session
// @Summary Delete session
// @Description Revoke a specific refresh token
// @Tags auth
// @Produce json
// @Security Bearer
// @Param id path string true "Session ID"
// @Success 200
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/auth/sessions/:id [delete]
func (h *AuthHandler) DeleteSession(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authorization required",
		})
		return
	}

	sessionID := c.Param("id")

	// Verify ownership and revoke
	err := h.refreshTokenService.RevokeRefreshToken(ctx, userID, sessionID)
	if err != nil {
		if err.Error() == "token not found or already revoked" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "SESSION_REVOKE_FAILED",
			Message: "Failed to revoke session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session revoked successfully",
	})
}

// GetSessionInfo returns session expiration info
// @Summary Get session info
// @Description Get current session expiration information
// @Tags auth
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/session-info [get]
func (h *AuthHandler) GetSessionInfo(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authorization required",
		})
		return
	}

	tokens, err := h.refreshTokenService.GetUserRefreshTokens(ctx, userID)
	if err != nil || len(tokens) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"active_sessions": 0,
		})
		return
	}

	// Return the most recent session info
	token := tokens[0]
	c.JSON(http.StatusOK, gin.H{
		"expires_at":      token.ExpiresAt.Time.Format(time.RFC3339),
		"max_valid_until": token.MaxValidUntil.Time.Format(time.RFC3339),
		"session_id":      token.TokenID,
	})
}

// RevokeAllSessions revokes all user sessions (admin endpoint)
// @Summary Revoke all sessions (Admin)
// @Description Revoke all refresh tokens for a user
// @Tags auth
// @Produce json
// @Security Bearer
// @Param userId path string true "User ID"
// @Success 200
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/admin/auth/revoke-all/:userId [post]
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	ctx := c.Request.Context()

	// Check admin permission
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "Admin access required",
		})
		return
	}

	targetUserID := c.Param("userId")

	// Revoke all user's refresh tokens
	err := h.refreshTokenService.RevokeAllUserTokens(ctx, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "REVOKE_FAILED",
			Message: "Failed to revoke sessions",
		})
		return
	}

	// Blacklist all their access tokens (this is done on-demand in middleware)
	// But we should log this event
	h.logAuditEvent(ctx, "admin_revoke_all", &targetUserID, c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "All sessions revoked successfully",
	})
}

// ExchangeAPIKey exchanges an API key for JWT tokens (for beacon/device auth)
// @Summary Exchange API key
// @Description Exchange API key for access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "API key"
// @Success 200 {object} models.TokenResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/beacon/token [post]
func (h *AuthHandler) ExchangeAPIKey(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate Content-Type header (prevent content-type confusion attacks)
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, models.ErrorResponse{
			Code:    "UNSUPPORTED_MEDIA_TYPE",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var req struct {
		APIKey string `json:"api_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
		})
		return
	}

	ipAddress := c.ClientIP()
	rateLimitKey := fmt.Sprintf("apikey:%s", ipAddress)

	// Check rate limit (11 requests per minute)
	allowed, _, resetTime, err := h.rateLimiter.CheckRateLimit(ctx, rateLimitKey, WindowPerMinute, 11)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "RATE_LIMIT_ERROR",
			Message: "Failed to check rate limit",
		})
		return
	}

	if !allowed {
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Code:    "ERR_RATE_LIMIT_EXCEEDED",
			Message: "Too many API key exchange attempts",
		})
		return
	}

	// Validate API key
	apiKey, err := h.apiKeyService.ValidateAPIKey(ctx, req.APIKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "INVALID_API_KEY",
			Message: "Invalid API key",
		})
		return
	}

	// Generate tokens
	// API keys don't have a user, so we use a special beacon user
	// The role will be "beacon"
	userID := apiKey.UserID
	if userID == nil {
		// Create a beacon-specific user ID
		userID = &apiKey.KeyPrefix
	}

	accessToken, jti, err := h.jwtService.GenerateAccessToken(*userID, "beacon")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate access token",
		})
		return
	}

	userAgent := c.GetHeader("User-Agent")
	refreshToken, _, err := h.refreshTokenService.CreateRefreshToken(
		ctx, *userID, userAgent, ipAddress, h.maxValidityDays,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate refresh token",
		})
		return
	}

	// Log API key exchange
	h.logAuditEvent(ctx, "api_key_exchange", userID, ipAddress, map[string]interface{}{
		"jti":          jti,
		"api_key_id":   apiKey.ID,
		"api_key_name": apiKey.Name,
	})

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        h.accessExpirationMinutes * 60,
		RefreshExpiresIn: h.refreshExpirationDays * 86400,
	})
}

// logAuditEvent logs security events to the audit log table
func (h *AuthHandler) logAuditEvent(ctx context.Context, eventType string, userID *string, ipAddress string, details map[string]interface{}) {
	if userID == nil {
		uid := ""
		userID = &uid
	}

	// Handle empty IP address (set to NULL)
	var ipAddressPtr *string
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	_, err := h.pool.Exec(ctx, `
		INSERT INTO auth_audit_logs (event_type, user_id, ip_address, details)
		VALUES ($1, $2, $3, $4)
	`, eventType, *userID, ipAddressPtr, details)

	if err != nil {
		// Log sanitized error (no sensitive details)
		fmt.Printf("WARN: failed to log audit event: %s\n", eventType)
	}
}

// constantAuthDelay provides a constant 150ms delay to prevent timing attacks
// on authentication failures. This ensures all failed auth paths take the same time.
func constantAuthDelay() {
	time.Sleep(150 * time.Millisecond)
}

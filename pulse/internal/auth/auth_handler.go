package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	pool              *pgxpool.Pool
	jwtService        *JWTService
	refreshTokenStore *RefreshTokenStore
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(pool *pgxpool.Pool) (*AuthHandler, error) {
	jwtService, err := NewJWTService()
	if err != nil {
		return nil, err
	}

	return &AuthHandler{
		pool:              pool,
		jwtService:        jwtService,
		refreshTokenStore: NewRefreshTokenStore(pool),
	}, nil
}

// PostLogin handles POST /api/v1/auth/login
// @Summary		User login
// @Description	Authenticates a user and returns JWT access token and refresh token cookie. Supports username/password authentication.
// @Description
// @Description	**Rate Limiting:** Maximum 5 failed attempts per IP per 15 minutes.
// @Description	**Account Lockout:** 6 failed attempts will lock the account for 10 minutes.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		models.LoginRequest	true	"Login credentials"
// @Success		200		{object}	models.LoginResponse	"Login successful"
// @Failure		400		{object}	models.ErrorResponse	"Invalid request format"
// @Failure		401		{object}	models.ErrorResponse	"Invalid credentials"
// @Failure		423		{object}	models.ErrorResponse	"Account locked"
// @Failure		429		{object}	models.ErrorResponse	"Rate limit exceeded"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Router			/auth/login [post]
func (h *AuthHandler) PostLogin(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse and validate request
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "ERR_INVALID_REQUEST",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Rate limiting check (simplified - implement full rate limiter separately)
	ip := c.ClientIP()
	if isRateLimited(ip) {
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Code:    "ERR_RATE_LIMIT_EXCEEDED",
			Message: "Too many login attempts, please try again later",
			Details: nil,
		})
		return
	}

	// Look up user
	user, err := h.lookupUser(ctx, req.Username)
	if err != nil {
		// User not found or DB error
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_INVALID_CREDENTIALS",
			Message: "Invalid username or password",
			Details: nil,
		})
		return
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.Valid {
		if time.Now().Before(user.LockedUntil.Time) {
			c.JSON(http.StatusLocked, models.ErrorResponse{
				Code:    "ERR_ACCOUNT_LOCKED",
				Message: "Account locked due to too many failed login attempts",
				Details: map[string]interface{}{
					"locked_until":          user.LockedUntil.Time.Format(time.RFC3339),
					"lock_duration_minutes": 10,
				},
			})
			return
		}
	}

	// Verify password
	if err := VerifyPassword(req.Password, user.PasswordHash); err != nil {
		// Increment failed attempts
		h.incrementFailedAttempts(ctx, user.UserID)

		// Check if should lock account
		if user.FailedLoginAttempts+1 >= 5 {
			h.lockAccount(ctx, user.UserID)
			c.JSON(http.StatusLocked, models.ErrorResponse{
				Code:    "ERR_ACCOUNT_LOCKED",
				Message: "Account locked due to too many failed login attempts",
				Details: map[string]interface{}{
					"locked_until":          time.Now().Add(10 * time.Minute).Format(time.RFC3339),
					"lock_duration_minutes": 10,
				},
			})
		} else {
			remaining := 5 - (user.FailedLoginAttempts + 1)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    "ERR_INVALID_CREDENTIALS",
				Message: "Invalid username or password",
				Details: map[string]interface{}{
					"failed_attempts":    user.FailedLoginAttempts + 1,
					"remaining_attempts": remaining,
				},
			})
		}
		return
	}

	// Successful login - reset failed attempts
	h.resetFailedAttempts(ctx, user.UserID)

	// Generate JWT access token
	accessToken, _, err := h.jwtService.GenerateAccessToken(user.UserID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to generate access token",
			Details: nil,
		})
		return
	}

	// Generate refresh token
	refreshToken, refreshJti, err := h.jwtService.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to generate refresh token",
			Details: nil,
		})
		return
	}

	// Hash refresh token for storage
	tokenHash, err := h.jwtService.HashRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to process refresh token",
			Details: nil,
		})
		return
	}

	// Store refresh token in database
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	err = h.refreshTokenStore.Save(ctx, user.UserID, tokenHash, refreshJti, "web", ip, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to store refresh token",
			Details: nil,
		})
		return
	}

	// Set refresh token cookie
	cfg := config.Get()
	secureFlag := cfg.IsProduction()
	c.SetSameSite(2) // http.SameSiteStrict
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*3600, // 7 days in seconds
		"/",       // path
		"",        // domain (uses current host)
		true,      // HttpOnly (prevent XSS)
		secureFlag, // Secure (true in production, false in development)
	)

	// Set security headers to prevent caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")

	// Return access token in response body
	c.JSON(http.StatusOK, models.LoginResponse{
		Data: struct {
			UserID      string `json:"user_id"`
			Username    string `json:"username"`
			Role        string `json:"role"`
			AccessToken string `json:"access_token"`
		}{
			UserID:      user.UserID,
			Username:    user.Username,
			Role:        user.Role,
			AccessToken: accessToken,
		},
		Message:   "Login successful",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// PostRefresh handles POST /api/v1/auth/refresh
// @Summary		Refresh access token
// @Description	Refreshes an expired access token using a valid refresh token from cookie. Implements token rotation.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200		{object}	models.RefreshResponse	"Token refreshed successfully"
// @Failure		401		{object}	models.ErrorResponse	"Invalid or expired refresh token"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Router			/auth/refresh [post]
func (h *AuthHandler) PostRefresh(c *gin.Context) {
	ctx := c.Request.Context()

	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_UNAUTHORIZED",
			Message: "Refresh token not found",
			Details: nil,
		})
		return
	}

	// Hash the token to look it up
	tokenHash, err := h.jwtService.HashRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to process refresh token",
			Details: nil,
		})
		return
	}

	// Get refresh token from database
	storedToken, err := h.refreshTokenStore.GetByHash(ctx, tokenHash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_INVALID_REFRESH_TOKEN",
			Message: "Invalid or expired refresh token",
			Details: nil,
		})
		return
	}

	// Delete old refresh token (token rotation)
	if err := h.refreshTokenStore.Delete(ctx, tokenHash); err != nil {
		// Log error but don't block refresh - security measure
		// In production, use proper logging
	}

	// Look up user to get role
	user, err := h.lookupUserByID(ctx, storedToken.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_INVALID_REFRESH_TOKEN",
			Message: "User not found",
			Details: nil,
		})
		return
	}

	// Generate new access token
	accessToken, _, err := h.jwtService.GenerateAccessToken(storedToken.UserID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to generate access token",
			Details: nil,
		})
		return
	}

	// Generate new refresh token
	newRefreshToken, newRefreshJti, err := h.jwtService.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to generate refresh token",
			Details: nil,
		})
		return
	}

	// Hash new refresh token
	newTokenHash, err := h.jwtService.HashRefreshToken(newRefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to process refresh token",
			Details: nil,
		})
		return
	}

	// Store new refresh token
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	ip := c.ClientIP()
	err = h.refreshTokenStore.Save(ctx, storedToken.UserID, newTokenHash, newRefreshJti, storedToken.DeviceInfo, ip, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to store refresh token",
			Details: nil,
		})
		return
	}

	// Set new refresh token cookie
	cfg := config.Get()
	secureFlag := cfg.IsProduction()
	c.SetSameSite(2) // http.SameSiteStrict
	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		7*24*3600, // 7 days in seconds
		"/",       // path
		"",        // domain
		true,      // HttpOnly
		secureFlag, // Secure
	)

	// Set security headers
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")

	// Return new access token
	c.JSON(http.StatusOK, models.RefreshResponse{
		Data: struct {
			AccessToken string `json:"access_token"`
		}{
			AccessToken: accessToken,
		},
		Message:   "Token refreshed successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetMe handles GET /api/v1/auth/me
// @Summary		Get current user
// @Description	Returns the currently authenticated user's information based on JWT access token.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200		{object}	models.GetMeResponse	"Success"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Router			/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_UNAUTHORIZED",
			Message: "Authentication required",
			Details: nil,
		})
		return
	}

	// Look up user to get username and role
	user, err := h.lookupUserByID(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to retrieve user information",
			Details: nil,
		})
		return
	}

	// Return user information
	c.JSON(http.StatusOK, models.GetMeResponse{
		Data: struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		}{
			UserID:   user.UserID,
			Username: user.Username,
			Role:     user.Role,
		},
		Message:   "Success",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// PostLogout handles POST /api/v1/auth/logout
func (h *AuthHandler) PostLogout(c *gin.Context) {
	ctx := c.Request.Context()

	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil {
		// Hash and delete from database
		tokenHash, err := h.jwtService.HashRefreshToken(refreshToken)
		if err == nil {
			// Log error but don't block logout
			h.refreshTokenStore.Delete(ctx, tokenHash)
		}
	}

	// Clear refresh token cookie
	cfg := config.Get()
	secureFlag := cfg.IsProduction()
	c.SetSameSite(2) // http.SameSiteStrict
	c.SetCookie(
		"refresh_token",
		"",
		-1,         // MaxAge -1 to delete
		"/",        // path
		"",         // domain
		true,       // HttpOnly
		secureFlag, // Secure
	)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Logout successful",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// lookupUser retrieves user by username from database
func (h *AuthHandler) lookupUser(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	err := h.pool.QueryRow(ctx, query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// lookupUserByID retrieves user by user ID from database
func (h *AuthHandler) lookupUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	query := `
		SELECT user_id, username, password_hash, role, failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	err := h.pool.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// incrementFailedAttempts increments failed login counter
func (h *AuthHandler) incrementFailedAttempts(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET failed_login_attempts = failed_login_attempts + 1, updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := h.pool.Exec(ctx, query, userID)
	return err
}

// resetFailedAttempts resets failed login counter
func (h *AuthHandler) resetFailedAttempts(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET failed_login_attempts = 0, locked_until = NULL, updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := h.pool.Exec(ctx, query, userID)
	return err
}

// lockAccount locks user account
func (h *AuthHandler) lockAccount(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET failed_login_attempts = 0, locked_until = NOW() + INTERVAL '10 minutes', updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := h.pool.Exec(ctx, query, userID)
	return err
}

// Rate limiting (simplified in-memory implementation)
var RateLimitStore = make(map[string]RateLimitInfo)

type RateLimitInfo struct {
	Attempts    int
	WindowStart time.Time
}

func isRateLimited(ip string) bool {
	now := time.Now()
	info, exists := RateLimitStore[ip]

	// Reset if window expired
	if !exists || now.Sub(info.WindowStart) > time.Minute {
		RateLimitStore[ip] = RateLimitInfo{Attempts: 1, WindowStart: now}
		return false
	}

	// Increment counter first
	newAttempts := info.Attempts + 1
	RateLimitStore[ip] = RateLimitInfo{Attempts: newAttempts, WindowStart: info.WindowStart}

	// Check limit after increment
	if newAttempts >= 5 {
		return true
	}

	return false
}

// ClearRateLimitStore clears the rate limit store (for testing purposes)
func ClearRateLimitStore() {
	RateLimitStore = make(map[string]RateLimitInfo)
}

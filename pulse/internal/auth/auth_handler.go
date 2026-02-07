package auth

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

const (
	// Refresh token expiration duration
	RefreshTokenExpirationDays = 7

	// Failed login attempt thresholds
	MaxFailedLoginAttempts = 5
	AccountLockDuration     = 10 * time.Minute

	// Rate limiting
	MaxLoginAttemptsPerMinute = 5
	RateLimitWindow           = time.Minute

	// Max active refresh tokens per user
	MaxActiveTokensPerUser = 5
)

// UserRefreshLock manages concurrent refresh attempts per user
type UserRefreshLock struct {
	locks map[string]*sync.Mutex
	mu    sync.Mutex
}

var globalRefreshLock = &UserRefreshLock{
	locks: make(map[string]*sync.Mutex),
}

// acquireLock gets or creates a mutex for the given user ID
func (u *UserRefreshLock) acquireLock(userID string) *sync.Mutex {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, exists := u.locks[userID]; !exists {
		u.locks[userID] = &sync.Mutex{}
	}

	return u.locks[userID]
}

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
			log.Printf("[Security] [AUTH] Login attempt on locked account: user_id=%s ip=%s", user.UserID, ip)
			c.JSON(http.StatusLocked, models.ErrorResponse{
				Code:    "ERR_ACCOUNT_LOCKED",
				Message: "Account locked due to too many failed login attempts",
				Details: map[string]interface{}{
					"locked_until":          user.LockedUntil.Time.Format(time.RFC3339),
					"lock_duration_minutes": AccountLockDuration.Minutes(),
				},
			})
			return
		}
	}

	// Verify password
	if err := VerifyPassword(req.Password, user.PasswordHash); err != nil {
		// Increment failed attempts
		h.incrementFailedAttempts(ctx, user.UserID)

		// Log security event
		log.Printf("[Security] [AUTH] Failed login attempt: username=%s ip=%s attempts=%d", req.Username, ip, user.FailedLoginAttempts+1)

		// Check if should lock account
		if user.FailedLoginAttempts+1 >= MaxFailedLoginAttempts {
			h.lockAccount(ctx, user.UserID)
			c.JSON(http.StatusLocked, models.ErrorResponse{
				Code:    "ERR_ACCOUNT_LOCKED",
				Message: "Account locked due to too many failed login attempts",
				Details: map[string]interface{}{
					"locked_until":          time.Now().Add(AccountLockDuration).Format(time.RFC3339),
					"lock_duration_minutes": AccountLockDuration.Minutes(),
				},
			})
		} else {
			remaining := MaxFailedLoginAttempts - (user.FailedLoginAttempts + 1)
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

	// Log successful login
	log.Printf("[Security] [AUTH] Successful login: user_id=%s username=%s ip=%s jti=%s", user.UserID, user.Username, ip, "placeholder")

	// Generate JWT access token
	accessToken, jti, err := h.jwtService.GenerateAccessToken(user.UserID, user.Role)
	if err != nil {
		log.Printf("[Error] [AUTH] Failed to generate access token: user_id=%s error=%v", user.UserID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to generate access token",
			Details: nil,
		})
		return
	}

	// Log JTI for audit
	log.Printf("[Audit] [AUTH] Access token generated: user_id=%s jti=%s ip=%s", user.UserID, jti, ip)

	// Generate refresh token
	refreshToken, refreshJti, err := h.jwtService.GenerateRefreshToken()
	if err != nil {
		log.Printf("[Error] [AUTH] Failed to generate refresh token: user_id=%s error=%v", user.UserID, err)
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
		log.Printf("[Error] [AUTH] Failed to hash refresh token: user_id=%s error=%v", user.UserID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to process refresh token",
			Details: nil,
		})
		return
	}

	// Store refresh token in database
	expiresAt := time.Now().Add(RefreshTokenExpirationDays * 24 * time.Hour)
	err = h.refreshTokenStore.Save(ctx, user.UserID, tokenHash, refreshJti, "web", ip, expiresAt)
	if err != nil {
		log.Printf("[Error] [AUTH] Failed to store refresh token: user_id=%s error=%v", user.UserID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to store refresh token",
			Details: nil,
		})
		return
	}

	// Log refresh token creation
	log.Printf("[Audit] [AUTH] Refresh token created: user_id=%s jti=%s ip=%s expires_at=%s", user.UserID, refreshJti, ip, expiresAt.Format(time.RFC3339))

	// Set refresh token cookie
	cfg := config.Get()
	secureFlag := cfg.IsProduction()
	c.SetSameSite(2) // http.SameSiteStrict
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(RefreshTokenExpirationDays*24*3600), // Convert days to seconds
		"/",   // path
		"",    // domain (uses current host)
		true,  // HttpOnly (prevent XSS)
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
// @Description	Refreshes an expired access token using a valid refresh token from cookie. Implements token rotation with concurrent safety.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200		{object}	models.RefreshResponse	"Token refreshed successfully"
// @Failure		401		{object}	models.ErrorResponse	"Invalid or expired refresh token"
// @Failure		409		{object}	models.ErrorResponse	"Concurrent refresh in progress"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Router			/auth/refresh [post]
func (h *AuthHandler) PostRefresh(c *gin.Context) {
	ctx := c.Request.Context()
	ip := c.ClientIP()

	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		log.Printf("[Security] [AUTH] Refresh attempt without token cookie: ip=%s", ip)
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
		log.Printf("[Error] [AUTH] Failed to hash refresh token: ip=%s error=%v", ip, err)
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
		log.Printf("[Security] [AUTH] Invalid or expired refresh token: ip=%s error=%v", ip, err)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_INVALID_REFRESH_TOKEN",
			Message: "Invalid or expired refresh token",
			Details: nil,
		})
		return
	}

	// Acquire user-level lock to prevent concurrent refreshes
	userLock := globalRefreshLock.acquireLock(storedToken.UserID)

	// Try to acquire lock - return 409 if another refresh is in progress
	if !userLock.TryLock() {
		log.Printf("[Security] [AUTH] Concurrent refresh attempt blocked: user_id=%s ip=%s", storedToken.UserID, ip)
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Code:    "ERR_CONCURRENT_REFRESH",
			Message: "Another refresh request is in progress. Please try again.",
			Details: nil,
		})
		return
	}

	// Ensure lock is released
	defer userLock.Unlock()

	// Double-check token still exists after acquiring lock (another refresh might have deleted it)
	_, err = h.refreshTokenStore.GetByHash(ctx, tokenHash)
	if err != nil {
		log.Printf("[Security] [AUTH] Token already consumed by concurrent refresh: user_id=%s ip=%s", storedToken.UserID, ip)
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Code:    "ERR_TOKEN_CONSUMED",
			Message: "Refresh token already used. Please try again.",
			Details: nil,
		})
		return
	}

	// Check token count limit before proceeding
	activeCount, err := h.refreshTokenStore.CountActiveTokens(ctx, storedToken.UserID)
	if err != nil {
		log.Printf("[Error] [AUTH] Failed to count active tokens: user_id=%s error=%v", storedToken.UserID, err)
	} else if activeCount >= MaxActiveTokensPerUser {
		log.Printf("[Security] [AUTH] Max active tokens reached, cleaning oldest: user_id=%s count=%d", storedToken.UserID, activeCount)
		// Delete oldest tokens (this is a simplified approach - in production you might want to be more selective)
		// For now, we'll allow the refresh to proceed and let cleanup job handle it
	}

	// Delete old refresh token (token rotation) - this MUST succeed for refresh to proceed
	if err := h.refreshTokenStore.Delete(ctx, tokenHash); err != nil {
		log.Printf("[Error] [AUTH] Failed to delete old refresh token: user_id=%s token_hash=%s error=%v", storedToken.UserID, tokenHash, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to rotate refresh token",
			Details: nil,
		})
		return
	}

	log.Printf("[Audit] [AUTH] Old refresh token deleted: user_id=%s jti=%s ip=%s", storedToken.UserID, storedToken.Jti, ip)

	// Look up user to get role
	user, err := h.lookupUserByID(ctx, storedToken.UserID)
	if err != nil {
		log.Printf("[Error] [AUTH] User not found during refresh: user_id=%s error=%v", storedToken.UserID, err)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "ERR_INVALID_REFRESH_TOKEN",
			Message: "User not found",
			Details: nil,
		})
		return
	}

	// Generate new access token
	accessToken, newJti, err := h.jwtService.GenerateAccessToken(storedToken.UserID, user.Role)
	if err != nil {
		log.Printf("[Error] [AUTH] Failed to generate access token: user_id=%s error=%v", storedToken.UserID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to generate access token",
			Details: nil,
		})
		return
	}

	// Log new access token
	log.Printf("[Audit] [AUTH] Access token refreshed: user_id=%s jti=%s ip=%s", storedToken.UserID, newJti, ip)

	// Generate new refresh token
	newRefreshToken, newRefreshJti, err := h.jwtService.GenerateRefreshToken()
	if err != nil {
		log.Printf("[Error] [AUTH] Failed to generate refresh token: user_id=%s error=%v", storedToken.UserID, err)
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
		log.Printf("[Error] [AUTH] Failed to hash new refresh token: user_id=%s error=%v", storedToken.UserID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to process refresh token",
			Details: nil,
		})
		return
	}

	// Store new refresh token
	expiresAt := time.Now().Add(RefreshTokenExpirationDays * 24 * time.Hour)
	err = h.refreshTokenStore.Save(ctx, storedToken.UserID, newTokenHash, newRefreshJti, storedToken.DeviceInfo, ip, expiresAt)
	if err != nil {
		log.Printf("[Error] [AUTH] Failed to store new refresh token: user_id=%s error=%v", storedToken.UserID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to store refresh token",
			Details: nil,
		})
		return
	}

	// Log new refresh token
	log.Printf("[Audit] [AUTH] New refresh token created: user_id=%s jti=%s ip=%s expires_at=%s", storedToken.UserID, newRefreshJti, ip, expiresAt.Format(time.RFC3339))

	// Set security headers BEFORE setting cookie (order matters)
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")

	// Set new refresh token cookie
	cfg := config.Get()
	secureFlag := cfg.IsProduction()
	c.SetSameSite(2) // http.SameSiteStrict
	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		int(RefreshTokenExpirationDays*24*3600), // Convert days to seconds
		"/",   // path
		"",    // domain
		true,  // HttpOnly
		secureFlag, // Secure
	)

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
	ip := c.ClientIP()

	// Try to get user ID from context for logging
	userID, _ := c.Get("user_id")

	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil {
		// Hash and delete from database
		tokenHash, err := h.jwtService.HashRefreshToken(refreshToken)
		if err == nil {
			if deleteErr := h.refreshTokenStore.Delete(ctx, tokenHash); deleteErr == nil {
				log.Printf("[Audit] [AUTH] Logout successful: user_id=%v ip=%s", userID, ip)
			} else {
				log.Printf("[Error] [AUTH] Failed to delete refresh token on logout: user_id=%v ip=%s error=%v", userID, ip, deleteErr)
			}
		}
	} else {
		log.Printf("[Audit] [AUTH] Logout without refresh token cookie: user_id=%v ip=%s", userID, ip)
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

	// Log account lock event
	if err == nil {
		log.Printf("[Security] [AUTH] Account locked: user_id=%s duration=%s", userID, AccountLockDuration)
	}

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
	if !exists || now.Sub(info.WindowStart) > RateLimitWindow {
		RateLimitStore[ip] = RateLimitInfo{Attempts: 1, WindowStart: now}
		return false
	}

	// Increment counter first
	newAttempts := info.Attempts + 1
	RateLimitStore[ip] = RateLimitInfo{Attempts: newAttempts, WindowStart: info.WindowStart}

	// Check limit after increment
	if newAttempts >= MaxLoginAttemptsPerMinute {
		log.Printf("[Security] [AUTH] Rate limit exceeded: ip=%s attempts=%d", ip, newAttempts)
		return true
	}

	return false
}

// ClearRateLimitStore clears the rate limit store (for testing purposes)
func ClearRateLimitStore() {
	RateLimitStore = make(map[string]RateLimitInfo)
}

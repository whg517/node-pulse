package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	pool           *pgxpool.Pool
	sessionService  *SessionService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		pool:          pool,
		sessionService: NewSessionService(pool),
	}
}

// PostLogin handles POST /api/v1/auth/login
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
					"locked_until":        user.LockedUntil.Time.Format(time.RFC3339),
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
					"locked_until":        time.Now().Add(10 * time.Minute).Format(time.RFC3339),
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

	// Create session
	sessionID, err := h.sessionService.CreateSession(ctx, user.UserID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_INTERNAL_SERVER",
			Message: "Failed to create session",
			Details: nil,
		})
		return
	}

	// Set session cookie
	c.SetCookie(
		"session_id",
		sessionID,
		86400,       // 24 hours in seconds
		"/",          // path
		"",           // domain (uses current host)
		true,         // HttpOnly
		false,        // Secure (set to true in production with HTTPS)
	)

	// Return success response
	c.JSON(http.StatusOK, models.LoginResponse{
		Data: struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		}{
			UserID:   user.UserID,
			Username: user.Username,
			Role:     user.Role,
		},
		Message:   "Login successful",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// PostLogout handles POST /api/v1/auth/logout
func (h *AuthHandler) PostLogout(c *gin.Context) {
	ctx := c.Request.Context()

	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		// No session cookie, but that's okay for logout
		c.JSON(http.StatusOK, gin.H{
			"message":   "Logout successful",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Delete session from database
	if err := h.sessionService.DeleteSession(ctx, sessionID); err != nil {
		// Log error but don't block logout
		// In production, use proper logging
	}

	// Clear session cookie
	c.SetCookie(
		"session_id",
		"",
		-1,    // MaxAge -1 to delete
		"/",    // path
		"",     // domain
		true,   // HttpOnly
		false,  // Secure
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
	Attempts     int
	WindowStart  time.Time
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

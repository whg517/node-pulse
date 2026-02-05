package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SessionValidator defines the interface for session validation
// This allows the middleware to be decoupled from the concrete auth service
type SessionValidator interface {
	GetSession(ctx context.Context, sessionID string) (userID string, role string, err error)
}

// AuthMiddleware validates session cookie and sets user context
func AuthMiddleware(validator SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract session ID from cookie
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			// No session cookie = unauthenticated
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			})
			return
		}

		// Validate session
		userID, role, err := validator.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			// Invalid or expired session
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_INVALID_SESSION",
				"message": "Invalid or expired session",
			})
			return
		}

		// Set user context for protected routes
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// GetUserRole retrieves the user role from the Gin context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	r, ok := role.(string)
	return r, ok
}

// RequireAuth is a shorthand to get user ID and role, aborting if not authenticated
func RequireAuth(c *gin.Context) (userID, role string, ok bool) {
	userID, ok = GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_UNAUTHORIZED",
			"message": "Authentication required",
		})
		return "", "", false
	}

	role, ok = GetUserRole(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_UNAUTHORIZED",
			"message": "Authentication required",
		})
		return "", "", false
	}

	return userID, role, true
}

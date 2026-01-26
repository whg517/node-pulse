package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates session cookie and sets user context
func AuthMiddleware(sessionService *SessionService) gin.HandlerFunc {
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
		userID, role, err := sessionService.GetSession(c.Request.Context(), sessionID)
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

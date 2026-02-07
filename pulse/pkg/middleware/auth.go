package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/whg517/node-pulse/pulse/internal/auth"
)

var (
	jwtService     *auth.JWTService
	jwtServiceOnce sync.Once
	jwtServiceErr  error
)

// initJWTService initializes the JWT service singleton
func initJWTService() error {
	var initErr error
	jwtServiceOnce.Do(func() {
		jwtService, initErr = auth.NewJWTService()
	})
	return initErr
}

// JWTAuthMiddleware validates JWT access token from Authorization header
func JWTAuthMiddleware() gin.HandlerFunc {
	// Initialize JWT service once
	if err := initJWTService(); err != nil {
		// If we can't initialize JWT service, return a middleware that always fails
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    "ERR_INTERNAL_SERVER",
				"message": "Failed to initialize JWT service",
			})
		}
	}

	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authorization header required",
			})
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_INVALID_TOKEN_FORMAT",
				"message": "Invalid authorization header format. Expected: Bearer <token>",
			})
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_INVALID_TOKEN",
				"message": fmt.Sprintf("Invalid or expired access token: %v", err),
			})
			return
		}

		// Set user context for protected routes
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("jti", claims.Jti)

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

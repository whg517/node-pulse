package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/auth"
)

// JWTAuthMiddleware validates JWT tokens and checks blacklist
func JWTAuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_INVALID_TOKEN_FORMAT",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract token
		tokenString := authHeader[7:]

		// Validate token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_INVALID_TOKEN",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Check blacklist
		ctx := c.Request.Context()
		revoked, err := jwtService.CheckRevoked(ctx, claims.JTI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "BLACKLIST_CHECK_FAILED",
				"message": "Failed to verify token status",
			})
			c.Abort()
			return
		}

		if revoked {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "TOKEN_REVOKED",
				"message": "Token has been revoked",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("jti", claims.JTI)

		c.Next()
	}
}

// GetUserID retrieves user_id from context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("user_id not found in context")
	}
	return userID.(string), nil
}

// GetUserRole retrieves role from context
func GetUserRole(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errors.New("role not found in context")
	}
	return role.(string), nil
}

// GetJTI retrieves JTI from context
func GetJTI(c *gin.Context) (string, error) {
	jti, exists := c.Get("jti")
	if !exists {
		return "", errors.New("jti not found in context")
	}
	return jti.(string), nil
}

// RequireAuth is a helper that checks if user is authenticated
func RequireAuth() gin.HandlerFunc {
	return JWTAuthMiddleware(nil) // Will be configured with actual service
}

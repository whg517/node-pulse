package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kevin/node-pulse/pulse-api/internal/config"
)

// CORSMiddleware provides CORS support with configurable origins
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()

		// Get allowed origins from config
		allowedOrigins := cfg.CORS.AllowedOrigins
		if allowedOrigins == "" {
			// Default to localhost for development
			allowedOrigins = "http://localhost:8080"
		}

		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		isAllowed := false
		if origin == "" {
			// No Origin header (same-origin or non-browser request)
			isAllowed = true
		} else {
			for _, allowed := range strings.Split(allowedOrigins, ",") {
				if strings.TrimSpace(allowed) == "*" || strings.TrimSpace(allowed) == origin {
					isAllowed = true
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		if isAllowed {
			// Set CORS headers from config
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

			// Use config for headers, with fallback to default
			allowedHeaders := cfg.CORS.AllowedHeaders
			if allowedHeaders == "" {
				allowedHeaders = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
			}
			c.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			// Use config for methods, with fallback to default
			allowedMethods := cfg.CORS.AllowedMethods
			if allowedMethods == "" {
				allowedMethods = "POST, OPTIONS, GET, PUT, DELETE, PATCH"
			}
			c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)

			// Handle preflight requests
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}

		c.Next()
	}
}

package api

import (
	"github.com/gin-gonic/gin"

	"github.com/kevin/node-pulse/pulse-api/internal/health"
	"github.com/kevin/node-pulse/pulse-api/pkg/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, healthChecker *health.HealthChecker) {
	// Apply error handling middleware
	router.Use(middleware.ErrorHandler())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", healthChecker.Handler)
	}
}

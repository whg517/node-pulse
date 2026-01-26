package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin/node-pulse/pulse-api/internal/health"
	"github.com/kevin/node-pulse/pulse-api/internal/auth"
	"github.com/kevin/node-pulse/pulse-api/pkg/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, healthChecker *health.HealthChecker, pool *pgxpool.Pool) {
	// Apply error handling middleware
	router.Use(middleware.ErrorHandler())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint (public)
		v1.GET("/health", healthChecker.Handler)

		// Auth endpoints (public)
		authHandler := auth.NewAuthHandler(pool)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.PostLogin)
			authGroup.POST("/logout", authHandler.PostLogout)
		}

		// Protected routes example (require auth)
		sessionService := auth.NewSessionService(pool)
		protected := v1.Group("/")
		protected.Use(auth.AuthMiddleware(sessionService))
		{
			// Future protected routes can be added here
			// Example: protected.GET("/nodes", handler)
		}
	}
}

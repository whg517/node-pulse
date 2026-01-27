package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/health"
	"github.com/kevin/node-pulse/pulse-api/internal/auth"
	"github.com/kevin/node-pulse/pulse-api/pkg/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, healthChecker *health.HealthChecker, pool *pgxpool.Pool) {
	// Initialize rate limiter
	middleware.InitRateLimiter()

	// Apply error handling and rate limiting middleware
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware())

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

		// Node management routes (require auth)
		sessionService := auth.NewSessionService(pool)
		nodeQuerier := db.NewPoolQuerier(pool)
		nodeHandler := NewNodeHandler(nodeQuerier)

		// Nodes group with auth middleware
		nodes := v1.Group("/nodes")
		nodes.Use(auth.AuthMiddleware(sessionService))

		// GET /api/v1/nodes - Get all nodes (all roles)
		nodes.GET("", nodeHandler.GetNodesHandler)

		// GET /api/v1/nodes/:id/status - Get node status (all roles)
		// CRITICAL: Specific route must come before generic /:id route
		nodes.GET("/:id/status", nodeHandler.GetNodeStatusHandler)

		// GET /api/v1/nodes/:id - Get node by ID (all roles)
		nodes.GET("/:id", nodeHandler.GetNodeByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		nodes.Use(auth.RBACMiddleware([]string{"admin", "operator"}))

		// POST /api/v1/nodes - Create node (admin/operator only)
		nodes.POST("", nodeHandler.CreateNodeHandler)

		// PUT /api/v1/nodes/:id - Update node (admin/operator only)
		nodes.PUT("/:id", nodeHandler.UpdateNodeHandler)

		// DELETE /api/v1/nodes/:id - Delete node (admin/operator only)
		nodes.DELETE("/:id", nodeHandler.DeleteNodeHandler)
	}
}

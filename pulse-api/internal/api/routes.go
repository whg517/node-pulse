package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin/node-pulse/pulse-api/internal/cache"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/health"
	"github.com/kevin/node-pulse/pulse-api/internal/auth"
	"github.com/kevin/node-pulse/pulse-api/pkg/middleware"
)

// CacheManager holds cache instances that need cleanup on shutdown
type CacheManager struct {
	MemoryCache *cache.MemoryCache
	BatchWriter *cache.BatchWriter
}

// SetupRoutes configures all API routes and returns cache manager for shutdown
func SetupRoutes(router *gin.Engine, healthChecker *health.HealthChecker, pool *pgxpool.Pool) *CacheManager {
	// Initialize rate limiter
	middleware.InitRateLimiter()

	// Apply error handling and rate limiting middleware
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware())

	// Initialize memory cache and batch writer (Story 3.2)
	memoryCache := cache.NewMemoryCache()
	batchWriter := cache.NewBatchWriter(pool, 1000, 100) // Buffer size 1000, batch size 100
	batchWriter.Start()

	// Create cache manager for graceful shutdown
	cacheManager := &CacheManager{
		MemoryCache: memoryCache,
		BatchWriter: batchWriter,
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint (public)
		v1.GET("/health", healthChecker.Handler)

		// Beacon endpoints (public - no auth required for MVP)
		beaconHandler := NewBeaconHandler(db.NewPoolQuerier(pool), memoryCache, batchWriter)
		beacon := v1.Group("/beacon")
		{
			// POST /api/v1/beacon/heartbeat - Receive heartbeat data (public)
			beacon.POST("/heartbeat", beaconHandler.HandleHeartbeat)
		}

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

		// Probe management routes (require auth)
		probeQuerier := db.NewPoolQuerier(pool)
		probeHandler := NewProbeHandler(probeQuerier, nodeQuerier)

		// Probes group with auth middleware
		probes := v1.Group("/probes")
		probes.Use(auth.AuthMiddleware(sessionService))

		// GET /api/v1/probes - Get all probes (all roles)
		probes.GET("", probeHandler.GetProbesHandler)

		// GET /api/v1/probes/:id - Get probe by ID (all roles)
		probes.GET("/:id", probeHandler.GetProbeByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		probes.Use(auth.RBACMiddleware([]string{"admin", "operator"}))

		// POST /api/v1/probes - Create probe (admin/operator only)
		probes.POST("", probeHandler.CreateProbeHandler)

		// PUT /api/v1/probes/:id - Update probe (admin/operator only)
		probes.PUT("/:id", probeHandler.UpdateProbeHandler)

		// DELETE /api/v1/probes/:id - Delete probe (admin/operator only)
		probes.DELETE("/:id", probeHandler.DeleteProbeHandler)
	}

	// Return cache manager for graceful shutdown
	return cacheManager
}

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/whg517/node-pulse/pulse/internal/alert"
	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/cache"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/export"
	"github.com/whg517/node-pulse/pulse/internal/health"
	"github.com/whg517/node-pulse/pulse/pkg/metrics"
	"github.com/whg517/node-pulse/pulse/pkg/middleware"
)

// CacheManager holds cache instances that need cleanup on shutdown
type CacheManager struct {
	MemoryCache      *cache.MemoryCache
	BatchWriter      *cache.BatchWriter
	AlertEngine      *alert.AlertEngine
	ExportService    *export.ExportService
	MetricsCollector *metrics.Collector
}

// SetupRoutes configures all API routes and returns cache manager for shutdown
func SetupRoutes(router *gin.Engine, healthChecker *health.HealthChecker, pool *pgxpool.Pool) *CacheManager {
	// Apply CORS middleware (must be first)
	router.Use(middleware.CORSMiddleware())

	// Initialize rate limiter
	middleware.InitRateLimiter()

	// Apply error handling and rate limiting middleware
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware())

	// Initialize memory cache and batch writer (Story 3.2)
	memoryCache := cache.NewMemoryCache()
	batchWriter := cache.NewBatchWriter(pool, 1000, 100) // Buffer size 1000, batch size 100
	batchWriter.Start()

	// Initialize alert engine (Story 5.5)
	alertQuerier := db.NewAlertQuerier(pool)
	alertEngineConfig := alert.DefaultEngineConfig()
	alertEngine := alert.NewAlertEngine(pool, alertQuerier, alertEngineConfig)
	alertEngine.Start()

	// Initialize export service (Story 8.1)
	exportService := export.NewExportService(pool)

	// Initialize metrics collector (Story 8.3)
	metricsCollector := metrics.NewCollector()
	metricsCollector.Start()

	// Initialize handlers that depend on metrics collector
	metricsHandler := NewMetricsHandler(metricsCollector)

	// Create cache manager for graceful shutdown
	cacheManager := &CacheManager{
		MemoryCache:      memoryCache,
		BatchWriter:      batchWriter,
		AlertEngine:      alertEngine,
		ExportService:    exportService,
		MetricsCollector: metricsCollector,
	}

	// Apply performance tracking middleware (Story 8.3)
	perfConfig := middleware.DefaultPerformanceConfig(metricsCollector)
	router.Use(middleware.PerformanceMiddleware(perfConfig))
	router.Use(middleware.InjectCollector(metricsCollector))

	// Swagger documentation route (public)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint (public)
		v1.GET("/health", healthChecker.Handler)

		// Beacon endpoints (public - no auth required for MVP)
		beaconHandler := NewBeaconHandler(db.NewPoolQuerier(pool), memoryCache, batchWriter, alertEngine)
		beacon := v1.Group("/beacon")
		{
			// POST /api/v1/beacon/heartbeat - Receive heartbeat data (public)
			beacon.POST("/heartbeat", beaconHandler.HandleHeartbeat)
		}

		// Auth endpoints (public)
		authHandler := auth.NewAuthHandler(pool)
		sessionService := auth.NewSessionService(pool)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.PostLogin)
			authGroup.POST("/logout", authHandler.PostLogout)
			authGroup.GET("/me", middleware.AuthMiddleware(sessionService), authHandler.GetMe)
		}

		// Config management routes (require admin auth only)
		configGroup := v1.Group("/config")
		configGroup.Use(middleware.AuthMiddleware(sessionService))
		configGroup.Use(middleware.RBACMiddleware([]string{"admin"}))
		{
			// GET /api/v1/config - Get current configuration (admin only, passwords redacted)
			configGroup.GET("", GetConfigHandler)

			// GET /api/v1/config/validate - Validate configuration (admin only)
			configGroup.GET("/validate", ValidateConfigHandler)
		}

		// Node management routes (require auth)
		nodeQuerier := db.NewPoolQuerier(pool)
		nodeHandler := NewNodeHandler(nodeQuerier)

		// Nodes group with auth middleware
		nodes := v1.Group("/nodes")
		nodes.Use(middleware.AuthMiddleware(sessionService))

		// GET /api/v1/nodes - Get all nodes (all roles)
		nodes.GET("", nodeHandler.GetNodesHandler)

		// GET /api/v1/nodes/:id/status - Get node status (all roles)
		// CRITICAL: Specific route must come before generic /:id route
		nodes.GET("/:id/status", nodeHandler.GetNodeStatusHandler)

		// GET /api/v1/nodes/:id - Get node by ID (all roles)
		nodes.GET("/:id", nodeHandler.GetNodeByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		nodes.Use(middleware.RBACMiddleware([]string{"admin", "operator"}))

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
		probes.Use(middleware.AuthMiddleware(sessionService))

		// GET /api/v1/probes - Get all probes (all roles)
		probes.GET("", probeHandler.GetProbesHandler)

		// GET /api/v1/probes/:id - Get probe by ID (all roles)
		probes.GET("/:id", probeHandler.GetProbeByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		probes.Use(middleware.RBACMiddleware([]string{"admin", "operator"}))

		// POST /api/v1/probes - Create probe (admin/operator only)
		probes.POST("", probeHandler.CreateProbeHandler)

		// PUT /api/v1/probes/:id - Update probe (admin/operator only)
		probes.PUT("/:id", probeHandler.UpdateProbeHandler)

		// DELETE /api/v1/probes/:id - Delete probe (admin/operator only)
		probes.DELETE("/:id", probeHandler.DeleteProbeHandler)

		// Data query routes (require auth)
		dataHandler := NewDataHandler(pool, memoryCache)
		data := v1.Group("/data")
		data.Use(middleware.AuthMiddleware(sessionService))

		// GET /api/v1/data/metrics - Get real-time metrics (all roles)
		data.GET("/metrics", dataHandler.GetMetricsHandler)

		// GET /api/v1/data/history - Get historical data (all roles)
		data.GET("/history", dataHandler.GetHistoryHandler)

		// GET /api/v1/data/comparison - Get node comparison data (all roles) (Story 7.2)
		data.GET("/comparison", dataHandler.GetComparisonHandler)

		// GET /api/v1/data/diagnosis - Get problem type diagnosis (all roles) (Story 7.4)
		data.GET("/diagnosis", dataHandler.GetDiagnosisHandler)

		// GET /api/v1/data/performance - Get performance metrics with targets (all roles) (Story 8.4)
		data.GET("/performance", metricsHandler.GetPerformanceData)

		// Export management routes (require admin auth only) (Story 8.1)
		exportHandler := NewExportHandler(exportService)

		// Export group with auth and RBAC middleware (admin only)
		exports := v1.Group("/data/export")
		exports.Use(middleware.AuthMiddleware(sessionService))
		exports.Use(middleware.RBACMiddleware([]string{"admin"}))
		{
			// POST /api/v1/data/export - Create export task (admin only)
			exports.POST("", exportHandler.CreateExportHandler)

			// GET /api/v1/data/export/:id - Get export status (admin only)
			exports.GET("/:id", exportHandler.GetExportStatusHandler)

			// GET /api/v1/data/export/:id/download - Download export file (admin only)
			exports.GET("/:id/download", exportHandler.DownloadExportHandler)
		}

		// Alert management routes (require auth) (Story 5.1)
		alertHandler := NewAlertHandler(alertQuerier)

		// Alerts group with auth middleware
		alerts := v1.Group("/alerts")
		alerts.Use(middleware.AuthMiddleware(sessionService))

		// GET /api/v1/alerts/rules - Get all alert rules (all roles)
		alerts.GET("/rules", alertHandler.GetAlertRulesHandler)

		// GET /api/v1/alerts/rules/:id - Get alert rule by ID (all roles)
		alerts.GET("/rules/:id", alertHandler.GetAlertRuleByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		alerts.Use(middleware.RBACMiddleware([]string{"admin", "operator"}))

		// POST /api/v1/alerts/rules - Create alert rule (admin/operator only)
		alerts.POST("/rules", alertHandler.CreateAlertRuleHandler)

		// PUT /api/v1/alerts/rules/:id - Update alert rule (admin/operator only)
		alerts.PUT("/rules/:id", alertHandler.UpdateAlertRuleHandler)

		// DELETE /api/v1/alerts/rules/:id - Delete alert rule (admin/operator only)
		alerts.DELETE("/rules/:id", alertHandler.DeleteAlertRuleHandler)

		// Alert record management routes (require auth) (Story 6.1)
		alertRecordHandler := NewAlertRecordHandler(pool)

		// Alert records group with auth middleware
		alertRecords := v1.Group("/alerts/records")
		alertRecords.Use(middleware.AuthMiddleware(sessionService))

		// GET /api/v1/alerts/records - Get alert records with filtering (all roles)
		alertRecords.GET("", alertRecordHandler.GetAlertRecordsHandler)

		// PUT /api/v1/alerts/records/:id/status - Update alert record status (all roles)
		alertRecords.PUT("/:id/status", alertRecordHandler.UpdateAlertRecordStatusHandler)

		// Webhook management routes (require admin auth only) (Story 5.2)
		webhookQuerier := db.NewWebhookQuerier(pool)
		webhookHandler := NewWebhookHandler(webhookQuerier)

		// Webhooks group with auth and RBAC middleware (admin only)
		webhooks := v1.Group("/webhooks")
		webhooks.Use(middleware.AuthMiddleware(sessionService))
		webhooks.Use(middleware.RBACMiddleware([]string{"admin"}))

		// GET /api/v1/webhooks - Get all webhook configurations (admin only)
		webhooks.GET("", webhookHandler.GetWebhooksHandler)

		// GET /api/v1/webhooks/:id - Get webhook configuration by ID (admin only)
		webhooks.GET("/:id", webhookHandler.GetWebhookByIDHandler)

		// POST /api/v1/webhooks - Create webhook configuration (admin only)
		webhooks.POST("", webhookHandler.CreateWebhookHandler)

		// PUT /api/v1/webhooks/:id - Update webhook configuration (admin only)
		webhooks.PUT("/:id", webhookHandler.UpdateWebhookHandler)

		// DELETE /api/v1/webhooks/:id - Delete webhook configuration (admin only)
		webhooks.DELETE("/:id", webhookHandler.DeleteWebhookHandler)

		// Performance metrics routes (require auth) (Story 8.3, 8.4)
		// Metrics group with auth middleware (all roles)
		metricsGroup := v1.Group("/metrics")
		metricsGroup.Use(middleware.AuthMiddleware(sessionService))
		{
			// GET /api/v1/metrics/performance - Get performance metrics (all roles)
			metricsGroup.GET("/performance", metricsHandler.GetPerformanceMetrics)

			// GET /api/v1/metrics/stats - Get collector statistics (all roles)
			metricsGroup.GET("/stats", metricsHandler.GetCollectorStats)
		}
	}

	// Return cache manager for graceful shutdown
	return cacheManager
}

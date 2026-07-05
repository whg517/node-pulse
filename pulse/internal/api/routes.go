package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/whg517/node-pulse/pulse/internal/alert"
	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/cache"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/csrf"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/export"
	"github.com/whg517/node-pulse/pulse/internal/health"
	"github.com/whg517/node-pulse/pulse/internal/notify"
	"github.com/whg517/node-pulse/pulse/internal/realtime"
	"github.com/whg517/node-pulse/pulse/pkg/metrics"
	"github.com/whg517/node-pulse/pulse/pkg/middleware"
	"github.com/whg517/node-pulse/pulse/web"
)

// CacheManager holds cache instances that need cleanup on shutdown
type CacheManager struct {
	MemoryCache      *cache.MemoryCache
	BatchWriter      *cache.BatchWriter
	AlertEngine      *alert.AlertEngine
	ExportService    *export.ExportService
	MetricsCollector *metrics.Collector
}

// mailAdapter bridges notify.Sender (notify.Attachment) to auth.Mailer
// (auth.MailerAttachment) so the auth package stays free of a notify import.
type mailAdapter struct {
	sender notify.Sender
}

func (m mailAdapter) Send(ctx context.Context, to, subject, body string, attachments ...auth.MailerAttachment) error {
	conv := make([]notify.Attachment, 0, len(attachments))
	for _, a := range attachments {
		conv = append(conv, notify.Attachment{Filename: a.Filename, Content: a.Content, ContentType: a.ContentType})
	}
	return m.sender.Send(ctx, to, subject, body, conv...)
}

// SetupRoutes configures all API routes and returns cache manager for shutdown.
// realtimeHub is the server-owned event hub (also used by the node-status
// sweeper); SetupRoutes wires it into the alert engine and /ws handler.
// cfg provides app config (notify/reset URL); mailer sends outbound email and
// may be a NoopSender when SMTP is unconfigured.
func SetupRoutes(router *gin.Engine, healthChecker *health.HealthChecker, pool *pgxpool.Pool, realtimeHub *realtime.Hub, cfg *config.Config, mailer notify.Sender) *CacheManager {
	// Apply CORS middleware (must be first)
	router.Use(middleware.CORSMiddleware())

	// Initialize rate limiter
	middleware.InitRateLimiter()

	// Initialize mTLS configuration (mandatory for production beacons)
	middleware.InitMTLSConfig()

	// Apply error handling and rate limiting middleware
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware())

	// Apply distributed tracing middleware (otelgin creates a span per request;
	// TraceIDMiddleware injects the trace ID into the response header and log prefix).
	router.Use(middleware.OtelGinMiddleware("pulse"))
	router.Use(middleware.TraceIDMiddleware())

	// Get JWT configuration from environment or use defaults
	jwtPrivateKey := getEnvOrDefault("PULSE_JWT_PRIVATE_KEY", "")
	jwtPublicKey := getEnvOrDefault("PULSE_JWT_PUBLIC_KEY", "")
	jwtKeyID := getEnvOrDefault("PULSE_JWT_KEY_ID", "")

	// If no RSA keys provided, generate them for RS256
	if jwtPrivateKey == "" || jwtPublicKey == "" {
		jwtPrivateKey, jwtPublicKey = generateRSAKeyPair()
		if jwtKeyID == "" {
			jwtKeyID = "default-key"
		}
	}

	// Get cookie secure setting from environment (default false for development)
	cookieSecure := getEnvOrDefault("PULSE_SESSION_COOKIE_SECURE", "false") == "true"

	// Load rate limit configuration (falls back to defaults if config unavailable)
	var rateLimitOpts auth.RateLimitOptions
	if cfg, err := config.Load(); err == nil {
		rateLimitOpts = auth.RateLimitOptions{
			LoginPerMinute:   cfg.RateLimit.LoginMaxPerMinute,
			RefreshPerMinute: cfg.RateLimit.RefreshMaxPerMinute,
			LogoutPerMinute:  cfg.RateLimit.RefreshMaxPerMinute, // reuse refresh limit for logout
			APIKeyPerMinute:  cfg.RateLimit.APIKeyMaxPerMinute,
		}
	}

	// Initialize JWT service with RS256
	jwtService := auth.NewJWTService(jwtPrivateKey, jwtPublicKey, jwtKeyID, 15, pool)

	// realtimeHub is provided by the caller (server-owned); wire it into the
	// alert engine and the /ws handler below.

	// Initialize auth handler
	authHandler := auth.NewAuthHandler(
		pool,
		jwtPrivateKey,
		jwtPublicKey,
		jwtKeyID,
		15, // 15 minutes access token
		7,  // 7 days refresh token
		30, // 30 days max validity
		cookieSecure,
		rateLimitOpts,
		auth.WithPasswordResetMailer(mailAdapter{sender: mailer}, cfg.Notify.PasswordResetURL),
	)

	// Initialize memory cache and batch writer (Story 3.2)
	memoryCache := cache.NewMemoryCache()
	batchWriter := cache.NewBatchWriter(pool, 1000, 100) // Buffer size 1000, batch size 100
	if pool != nil {
		batchWriter.Start()
	}

	// Initialize alert engine (Story 5.5)
	var alertEngine *alert.AlertEngine
	alertQuerier := db.NewAlertQuerier(pool)
	if pool != nil {
		alertEngineConfig := alert.DefaultEngineConfig()
		alertEngine = alert.NewAlertEngine(pool, alertQuerier, alertEngineConfig)
		alertEngine.WithRealtimeHub(realtimeHub)
		// Wire configured server.base_url so webhook payloads carry correct
		// absolute links instead of the localhost default.
		if cfg, err := config.Load(); err == nil {
			alertEngine.WithWebhookBaseURL(cfg.Server.BaseURL)
		}
		alertEngine.Start()
	}

	// Initialize export service (Story 8.1). Pass a durable task store so export
	// tasks survive server restarts (see docs/user-journey.md §17 G3).
	exportService := export.NewExportService(pool, db.NewExportTaskRepository(pool))

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

	// Prometheus metrics endpoint (public)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Realtime websocket endpoint. The frontend passes JWT access tokens as a
	// query parameter because browser WebSocket APIs cannot set auth headers.
	realtimeHandler := realtime.NewHandler(realtimeHub, jwtService)
	router.GET("/ws", realtimeHandler.ServeWS)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint (public)
		v1.GET("/health", healthChecker.Handler)

		// Beacon endpoints (JWT auth for beacons)
		beaconHandler := NewBeaconHandler(db.NewPoolQuerier(pool), memoryCache, batchWriter, alertEngine, realtimeHub)

		beacon := v1.Group("/beacon")
		{
			// POST /api/v1/beacon/token - Exchange API key for JWT token (public, API key auth)
			// This is the entry point for beacons to obtain JWT tokens
			beacon.POST("/token", authHandler.ExchangeAPIKey)

			// POST /api/v1/beacon/heartbeat - Receive heartbeat data (JWT + mTLS auth required)
			beacon.POST("/heartbeat", middleware.MTLSAuthMiddleware(), middleware.JWTAuthMiddleware(jwtService), beaconHandler.HandleHeartbeat)
			// POST /api/v1/beacon/heartbeat/compressed - Receive compressed heartbeat data (FR-4.1.5)
			beacon.POST("/heartbeat/compressed", middleware.MTLSAuthMiddleware(), middleware.JWTAuthMiddleware(jwtService), beaconHandler.HandleCompressedHeartbeat)
			// POST /api/v1/beacon/config/ack - Receive server config apply acknowledgement
			beacon.POST("/config/ack", middleware.MTLSAuthMiddleware(), middleware.JWTAuthMiddleware(jwtService), beaconHandler.AcknowledgeBeaconConfig)
			// POST /api/v1/beacon/mtr - Receive MTR route-hop result
			beacon.POST("/mtr", middleware.MTLSAuthMiddleware(), middleware.JWTAuthMiddleware(jwtService), beaconHandler.HandleMTRResult)
		}

		// Beacon config management routes (require auth) (FR-4.2.4)
		beacons := v1.Group("/beacons")
		beacons.Use(middleware.JWTAuthMiddleware(jwtService))
		{
			// GET /api/v1/beacons/:id/config - Get beacon config
			beacons.GET("/:id/config", beaconHandler.GetBeaconConfig)
			// POST /api/v1/beacons/:id/config - Update beacon config
			beacons.POST("/:id/config", middleware.RBACMiddleware([]string{"admin", "operator"}), beaconHandler.UpdateBeaconConfig)
			// GET /api/v1/beacons/:id/config/history - Get beacon config history
			beacons.GET("/:id/config/history", beaconHandler.GetBeaconConfigHistory)
			// POST /api/v1/beacons/:id/config/preview - Preview config changes
			beacons.POST("/:id/config/preview", beaconHandler.GetConfigPreview)
			// POST /api/v1/beacons/:id/config/rollback - Roll back config to a prior version (admin/operator)
			beacons.POST("/:id/config/rollback", middleware.RBACMiddleware([]string{"admin", "operator"}), beaconHandler.RollbackBeaconConfigHandler)
		}

		// Beacon group config management routes (admin/operator only)
		beaconGroups := v1.Group("/beacon-groups")
		beaconGroups.Use(middleware.JWTAuthMiddleware(jwtService))
		beaconGroups.Use(middleware.RBACMiddleware([]string{"admin", "operator"}))
		{
			// POST /api/v1/beacon-groups/:gid/config - Batch update beacon configs
			beaconGroups.POST("/:gid/config", beaconHandler.BatchUpdateBeaconGroupConfig)
		}

		// Beacon config templates (ADR-003). Owned by user; admin/operator write.
		beaconConfigTemplateHandler := NewBeaconConfigTemplateHandler(db.NewBeaconConfigTemplatesRepository(pool))
		templates := v1.Group("/beacon-config-templates")
		templates.Use(middleware.JWTAuthMiddleware(jwtService))
		{
			templates.GET("", beaconConfigTemplateHandler.ListBeaconConfigTemplatesHandler)
			templates.POST("", middleware.RBACMiddleware([]string{"admin", "operator"}), beaconConfigTemplateHandler.CreateBeaconConfigTemplateHandler)
			templates.PUT("/:id", middleware.RBACMiddleware([]string{"admin", "operator"}), beaconConfigTemplateHandler.UpdateBeaconConfigTemplateHandler)
			templates.DELETE("/:id", middleware.RBACMiddleware([]string{"admin", "operator"}), beaconConfigTemplateHandler.DeleteBeaconConfigTemplateHandler)
		}

		// Auth endpoints (public)
		authGroup := v1.Group("/auth")
		{
			// POST /api/v1/auth/login - User login
			authGroup.POST("/login", authHandler.Login)
			// POST /api/v1/auth/refresh - Refresh access token
			authGroup.POST("/refresh", authHandler.Refresh)
			// POST /api/v1/auth/logout - Logout (requires valid token to blacklist it)
			authGroup.POST("/logout", middleware.JWTAuthMiddleware(jwtService), authHandler.Logout)
			// GET /api/v1/auth/me - Get current user info (requires auth)
			authGroup.GET("/me", middleware.JWTAuthMiddleware(jwtService), authHandler.GetMe)
			// GET /api/v1/auth/sessions - Get user sessions (requires auth)
			authGroup.GET("/sessions", middleware.JWTAuthMiddleware(jwtService), authHandler.GetSessions)
			// DELETE /api/v1/auth/sessions/:id - Revoke specific session (requires auth)
			authGroup.DELETE("/sessions/:id", middleware.JWTAuthMiddleware(jwtService), authHandler.DeleteSession)
			// POST /api/v1/auth/sessions/revoke-all - Revoke all own sessions (requires auth)
			authGroup.POST("/sessions/revoke-all", middleware.JWTAuthMiddleware(jwtService), authHandler.RevokeAllMySessions)
			// GET /api/v1/auth/session-info - Get session expiration info (requires auth)
			authGroup.GET("/session-info", middleware.JWTAuthMiddleware(jwtService), authHandler.GetSessionInfo)
			// GET /api/v1/auth/verify - Validate current token and return claims (requires auth)
			authGroup.GET("/verify", middleware.JWTAuthMiddleware(jwtService), authHandler.GetMe)
			// POST /api/v1/auth/password/reset/request - Request password reset
			authGroup.POST("/password/reset/request", authHandler.RequestPasswordReset)
			// POST /api/v1/auth/password/reset/confirm - Confirm password reset
			authGroup.POST("/password/reset/confirm", authHandler.ConfirmPasswordReset)
			// POST /api/v1/auth/password/change - Change password (requires auth + CSRF)
			authGroup.POST("/password/change", middleware.JWTAuthMiddleware(jwtService), csrf.CSRFMiddleware(), authHandler.ChangePassword)
		}

		// Admin auth routes (admin only)
		adminAuth := v1.Group("/admin/auth")
		adminAuth.Use(middleware.JWTAuthMiddleware(jwtService))
		adminAuth.Use(middleware.RBACMiddleware([]string{"admin"}))
		{
			// POST /api/v1/admin/auth/revoke-all/:userId - Revoke all user sessions
			adminAuth.POST("/revoke-all/:userId", authHandler.RevokeAllSessions)
		}

		// Admin audit log routes (admin only)
		adminAuditHandler := NewAdminAuditHandler(pool)
		adminAudit := v1.Group("/admin/audit")
		adminAudit.Use(middleware.JWTAuthMiddleware(jwtService))
		adminAudit.Use(middleware.RBACMiddleware([]string{"admin"}))
		{
			// GET /api/v1/admin/audit/logs - Query audit logs (admin only)
			adminAudit.GET("/logs", adminAuditHandler.GetAuditLogs)
			// GET /api/v1/admin/audit/logs/:id - Get audit log by ID (admin only)
			adminAudit.GET("/logs/:id", adminAuditHandler.GetAuditLogByID)
		}

		// Admin user management routes (admin only)
		auditLogger := auth.NewAuditLogger(pool)
		adminUserHandler := NewAdminUserHandler(db.NewUserQuerier(pool), auditLogger)

		adminUsers := v1.Group("/admin/users")
		adminUsers.Use(middleware.JWTAuthMiddleware(jwtService))
		adminUsers.Use(middleware.RBACMiddleware([]string{"admin"}))
		{
			// GET /api/v1/admin/users - List all users (admin only)
			adminUsers.GET("", adminUserHandler.ListUsers)

			// GET /api/v1/admin/users/:id - Get user by ID (admin only)
			adminUsers.GET("/:id", adminUserHandler.GetUser)

			// POST /api/v1/admin/users - Create new user (admin only)
			adminUsers.POST("", adminUserHandler.CreateUser)

			// PUT /api/v1/admin/users/:id - Update user (admin only)
			adminUsers.PUT("/:id", adminUserHandler.UpdateUser)

			// DELETE /api/v1/admin/users/:id - Delete user (admin only)
			adminUsers.DELETE("/:id", adminUserHandler.DeleteUser)
		}

		// Admin API key management routes (admin only)
		apiKeyService := auth.NewAPIKeyService(pool)
		adminAPIKeyHandler := NewAdminAPIKeyHandler(apiKeyService)

		adminAPIKeys := v1.Group("/admin/apikeys")
		adminAPIKeys.Use(middleware.JWTAuthMiddleware(jwtService))
		adminAPIKeys.Use(middleware.RBACMiddleware([]string{"admin"}))
		{
			// GET /api/v1/admin/apikeys - List all API keys (admin only)
			adminAPIKeys.GET("", adminAPIKeyHandler.ListAPIKeysHandler)

			// GET /api/v1/admin/apikeys/:id - Get API key by ID (admin only)
			adminAPIKeys.GET("/:id", adminAPIKeyHandler.GetAPIKeyByIDHandler)

			// POST /api/v1/admin/apikeys - Create new API key (admin only)
			adminAPIKeys.POST("", adminAPIKeyHandler.CreateAPIKeyHandler)

			// POST /api/v1/admin/apikeys/:id/rotate - Rotate API key (admin only)
			adminAPIKeys.POST("/:id/rotate", adminAPIKeyHandler.RotateAPIKeyHandler)

			// DELETE /api/v1/admin/apikeys/:id - Revoke API key (admin only)
			adminAPIKeys.DELETE("/:id", adminAPIKeyHandler.RevokeAPIKeyHandler)
		}

		// Config management routes (require admin auth only)
		configGroup := v1.Group("/config")
		configGroup.Use(middleware.JWTAuthMiddleware(jwtService))
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
		nodes.Use(middleware.JWTAuthMiddleware(jwtService))

		// GET /api/v1/nodes - Get all nodes (all roles)
		nodes.GET("", nodeHandler.GetNodesHandler)

		// GET /api/v1/nodes/:id/status - Get node status (all roles)
		// CRITICAL: Specific route must come before generic /:id route
		nodes.GET("/:id/status", nodeHandler.GetNodeStatusHandler)

		// GET /api/v1/nodes/:id - Get node by ID (all roles)
		nodes.GET("/:id", nodeHandler.GetNodeByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		// Add CSRF protection for state-changing operations
		nodes.Use(csrf.CSRFMiddleware())
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
		probes.Use(middleware.JWTAuthMiddleware(jwtService))

		// GET /api/v1/probes - Get all probes (all roles)
		probes.GET("", probeHandler.GetProbesHandler)

		// GET /api/v1/probes/:id - Get probe by ID (all roles)
		probes.GET("/:id", probeHandler.GetProbeByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		// Add CSRF protection for state-changing operations
		nodes.Use(csrf.CSRFMiddleware())
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
		data.Use(middleware.JWTAuthMiddleware(jwtService))

		// GET /api/v1/data/metrics - Get real-time metrics (all roles)
		data.GET("/metrics", dataHandler.GetMetricsHandler)

		// GET /api/v1/data/history - Get historical data (all roles)
		data.GET("/history", dataHandler.GetHistoryHandler)

		// GET /api/v1/data/comparison - Get node comparison data (all roles) (Story 7.2)
		data.GET("/comparison", dataHandler.GetComparisonHandler)

		// GET /api/v1/data/diagnosis - Get problem type diagnosis (all roles) (Story 7.4)
		data.GET("/diagnosis", dataHandler.GetDiagnosisHandler)

		// GET /api/v1/data/mtr - Get latest MTR route-hop result for a node
		data.GET("/mtr", dataHandler.GetLatestMTRHandler)

		// GET /api/v1/data/mtr/history - Get historical MTR route-hop results for a node
		data.GET("/mtr/history", dataHandler.GetMTRHistoryHandler)

		// GET /api/v1/data/performance - Get performance metrics with targets (all roles) (Story 8.4)
		data.GET("/performance", metricsHandler.GetPerformanceData)

		// Export management routes (require admin auth only) (Story 8.1)
		exportHandler := NewExportHandler(exportService)

		// Export group with auth and RBAC middleware (admin only)
		exports := v1.Group("/data/export")
		exports.Use(middleware.JWTAuthMiddleware(jwtService))
		exports.Use(middleware.RBACMiddleware([]string{"admin"}))
		{
			// GET /api/v1/data/export - List export tasks (admin only)
			exports.GET("", exportHandler.ListExportsHandler)

			// POST /api/v1/data/export - Create export task (admin only)
			exports.POST("", exportHandler.CreateExportHandler)

			// GET /api/v1/data/export/:id - Get export status (admin only)
			exports.GET("/:id", exportHandler.GetExportStatusHandler)

			// GET /api/v1/data/export/:id/download - Download export file (admin only)
			exports.GET("/:id/download", exportHandler.DownloadExportHandler)
		}

		// Report schedules (ADR-001). Admin-managed recurring reports.
		reportScheduleHandler := NewReportScheduleHandler(db.NewReportScheduleRepository(pool))
		reportsSched := v1.Group("/reports/schedules")
		reportsSched.Use(middleware.JWTAuthMiddleware(jwtService))
		{
			reportsSched.GET("", reportScheduleHandler.ListReportSchedulesHandler)
			reportsSched.POST("", middleware.RBACMiddleware([]string{"admin"}), reportScheduleHandler.CreateReportScheduleHandler)
			reportsSched.PUT("/:id", middleware.RBACMiddleware([]string{"admin"}), reportScheduleHandler.UpdateReportScheduleHandler)
			reportsSched.DELETE("/:id", middleware.RBACMiddleware([]string{"admin"}), reportScheduleHandler.DeleteReportScheduleHandler)
		}

		// Alert management routes (require auth) (Story 5.1)
		alertHandler := NewAlertHandler(alertQuerier)

		// Alerts group with auth middleware
		alerts := v1.Group("/alerts")
		alerts.Use(middleware.JWTAuthMiddleware(jwtService))

		// Alert routing rules (ADR-002). CRUD for per-webhook routing.
		alertRoutingHandler := NewAlertRoutingHandler(db.NewAlertRoutingRulesRepository(pool))
		alerts.GET("/routing-rules", alertRoutingHandler.ListRoutingRulesHandler)
		alerts.POST("/routing-rules", middleware.RBACMiddleware([]string{"admin", "operator"}), alertRoutingHandler.CreateRoutingRuleHandler)
		alerts.PUT("/routing-rules/:id", middleware.RBACMiddleware([]string{"admin", "operator"}), alertRoutingHandler.UpdateRoutingRuleHandler)
		alerts.DELETE("/routing-rules/:id", middleware.RBACMiddleware([]string{"admin", "operator"}), alertRoutingHandler.DeleteRoutingRuleHandler)

		// GET /api/v1/alerts/rules - Get all alert rules (all roles)
		alerts.GET("/rules", alertHandler.GetAlertRulesHandler)

		// GET /api/v1/alerts/rules/:id - Get alert rule by ID (all roles)
		alerts.GET("/rules/:id", alertHandler.GetAlertRuleByIDHandler)

		// Create/Update/Delete routes require RBAC (admin or operator)
		// Add CSRF protection for state-changing operations
		nodes.Use(csrf.CSRFMiddleware())
		alerts.Use(middleware.RBACMiddleware([]string{"admin", "operator"}))

		// POST /api/v1/alerts/rules - Create alert rule (admin/operator only)
		alerts.POST("/rules", alertHandler.CreateAlertRuleHandler)

		// PUT /api/v1/alerts/rules/:id - Update alert rule (admin/operator only)
		alerts.PUT("/rules/:id", alertHandler.UpdateAlertRuleHandler)

		// DELETE /api/v1/alerts/rules/:id - Delete alert rule (admin/operator only)
		alerts.DELETE("/rules/:id", alertHandler.DeleteAlertRuleHandler)

		// Alert record management routes (require auth) (Story 6.1)
		alertRecordHandler := NewAlertRecordHandler(pool, realtimeHub)

		// Alert records group with auth middleware
		alertRecords := v1.Group("/alerts/records")
		alertRecords.Use(middleware.JWTAuthMiddleware(jwtService))

		// GET /api/v1/alerts/records - Get alert records with filtering (all roles)
		alertRecords.GET("", alertRecordHandler.GetAlertRecordsHandler)

		// PUT /api/v1/alerts/records/:id/status - Update alert record status (all roles)
		alertRecords.PUT("/:id/status", alertRecordHandler.UpdateAlertRecordStatusHandler)

		// POST /api/v1/alerts/records/:id/notes - Add alert investigation note (all roles)
		alertRecords.POST("/:id/notes", alertRecordHandler.AddAlertNoteHandler)

		// GET /api/v1/alerts/records/:id/notes - List alert investigation notes (all roles)
		alertRecords.GET("/:id/notes", alertRecordHandler.GetAlertNotesHandler)

		// GET /api/v1/alerts/records/:id/timeline - List merged alert lifecycle timeline (all roles)
		alertRecords.GET("/:id/timeline", alertRecordHandler.GetAlertTimelineHandler)

		// Webhook management routes (require admin auth only) (Story 5.2)
		webhookQuerier := db.NewWebhookQuerier(pool)
		var webhookLogsQuerier db.WebhookLogsQuerier
		if pool != nil {
			webhookLogsQuerier = db.NewWebhookLogsQuerier(pool)
		}
		webhookHandler := NewWebhookHandler(webhookQuerier, webhookLogsQuerier)

		// Webhooks group with auth and RBAC middleware (admin only)
		webhooks := v1.Group("/webhooks")
		webhooks.Use(middleware.JWTAuthMiddleware(jwtService))
		webhooks.Use(middleware.RBACMiddleware([]string{"admin"}))

		// GET /api/v1/webhooks - Get all webhook configurations (admin only)
		webhooks.GET("", webhookHandler.GetWebhooksHandler)

		// GET /api/v1/webhooks/:id - Get webhook configuration by ID (admin only)
		webhooks.GET("/:id", webhookHandler.GetWebhookByIDHandler)

		// GET /api/v1/webhooks/:id/logs - List delivery logs for a webhook (admin only)
		webhooks.GET("/:id/logs", webhookHandler.GetWebhookLogsHandler)

		// POST /api/v1/webhooks - Create webhook configuration (admin only)
		webhooks.POST("", webhookHandler.CreateWebhookHandler)

		// POST /api/v1/webhooks/preview - Preview rendered webhook payload (admin only)
		webhooks.POST("/preview", webhookHandler.PreviewWebhookEventHandler)

		// POST /api/v1/webhooks/:id/test - Send a sample delivery to one webhook (admin only)
		webhooks.POST("/:id/test", webhookHandler.TestWebhookHandler)

		// PUT /api/v1/webhooks/:id - Update webhook configuration (admin only)
		webhooks.PUT("/:id", webhookHandler.UpdateWebhookHandler)

		// DELETE /api/v1/webhooks/:id - Delete webhook configuration (admin only)
		webhooks.DELETE("/:id", webhookHandler.DeleteWebhookHandler)

		// Performance metrics routes (require auth) (Story 8.3, 8.4)
		// Metrics group with auth middleware (all roles)
		metricsGroup := v1.Group("/metrics")
		metricsGroup.Use(middleware.JWTAuthMiddleware(jwtService))
		{
			// GET /api/v1/metrics/performance - Get performance metrics (all roles)
			metricsGroup.GET("/performance", metricsHandler.GetPerformanceMetrics)

			// GET /api/v1/metrics/stats - Get collector statistics (all roles)
			metricsGroup.GET("/stats", metricsHandler.GetCollectorStats)
		}
	}

	// Serve the embedded frontend (Vite production build). The SPA is served
	// from the same origin as the API, so the browser uses relative URLs and
	// no CORS/reverse-proxy is needed.
	//
	// Order matters: this is registered AFTER all /api/v1, /swagger, /metrics
	// and /ws routes, so those take precedence. Static assets under /assets
	// are served with long cache headers; any other unmatched path falls back
	// to index.html so BrowserRouter (history mode) deep links resolve.
	registerFrontend(router)

	// Return cache manager for graceful shutdown
	return cacheManager
}

// registerFrontend wires the embedded frontend SPA onto the router.
func registerFrontend(router *gin.Engine) {
	dist := web.DistFS()
	fileServer := http.FileServer(http.FS(dist))

	// Static assets are fingerprinted by Vite, so they can be cached forever.
	// The embedded FS is rooted at dist/, so asset paths like /assets/app.js
	// map directly to "assets/app.js" inside the FS (no prefix stripping needed).
	router.GET("/assets/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// Other root-level static files (e.g. /vite.svg, /favicon.ico).
	router.StaticFileFS("/vite.svg", "vite.svg", http.FS(dist))

	// SPA fallback: any path not matched by the API or static handlers serves
	// index.html so client-side routing owns the URL. This must be a NoRoute
	// handler so it only catches genuinely unmatched paths.
	indexHTML, err := web.IndexHTML()
	if err != nil {
		slog.Error("Failed to load embedded index.html; SPA fallback disabled", "error", err)
		return
	}
	router.NoRoute(func(c *gin.Context) {
		// Do not hijack API/infra routes that were simply not found — return 404
		// JSON for those so clients get a clear signal instead of an HTML page.
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" || path == "/metrics" || path == "/swagger" || path == "/ws" {
			c.JSON(http.StatusNotFound, gin.H{"code": "ERR_NOT_FOUND", "message": "resource not found"})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// generateRSAKeyPair generates an RSA-2048 key pair for JWT RS256 signing
// Returns private key and public key in PEM format
func generateRSAKeyPair() (string, string) {
	// Generate 2048-bit RSA private key (minimum per design spec)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate RSA private key: %v", err))
	}

	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal public key: %v", err))
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM)
}

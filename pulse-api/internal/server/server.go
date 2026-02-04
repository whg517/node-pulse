package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin/node-pulse/pulse-api/internal/api"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/health"
	"github.com/kevin/node-pulse/pulse-api/internal/scheduler"
)

// Server represents the HTTP server with all its dependencies
type Server struct {
	config         *Config
	router         *gin.Engine
	httpServer     *http.Server
	database       *db.Database
	healthChecker  *health.HealthChecker
	scheduler      scheduler.Scheduler
	cacheManager   *api.CacheManager
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// Start starts the server and all its dependencies
func (s *Server) Start() error {
	log.Printf("[INFO] [Server] Starting Node Pulse API on port %s...", s.config.Server.Port)

	// Start scheduler
	if err := s.scheduler.Start(s.shutdownCtx); err != nil {
		return err
	}
	log.Println("[INFO] [Server] Scheduler started")

	// Start HTTP server in background
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] [Server] Failed to start: %v", err)
		}
	}()

	log.Println("[INFO] [Server] Node Pulse API started successfully")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	log.Println("[INFO] [Server] Shutting down...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop cache components
	if s.cacheManager != nil {
		s.stopCacheComponents()
	}

	// Stop scheduler
	if err := s.scheduler.Stop(); err != nil {
		log.Printf("[WARN] [Server] Error stopping scheduler: %v", err)
	}

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[WARN] [Server] Forced to shutdown: %v", err)
		return err
	}

	// Close database
	if s.database != nil {
		s.database.Close()
	}

	// Cancel shutdown context
	if s.shutdownCancel != nil {
		s.shutdownCancel()
	}

	log.Println("[INFO] [Server] Shutdown complete")
	return nil
}

// WaitForShutdown waits for interrupt signal and triggers graceful shutdown
func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := s.Shutdown(); err != nil {
		log.Printf("[ERROR] [Server] Shutdown error: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// stopCacheComponents stops all cache-related components
func (s *Server) stopCacheComponents() {
	log.Println("[INFO] [Server] Stopping cache components...")

	if s.cacheManager.AlertEngine != nil {
		log.Println("  -> Alert engine")
		s.cacheManager.AlertEngine.Stop()
	}

	if s.cacheManager.BatchWriter != nil {
		log.Println("  -> Batch writer")
		s.cacheManager.BatchWriter.Stop()
	}

	if s.cacheManager.MemoryCache != nil {
		log.Println("  -> Memory cache")
		s.cacheManager.MemoryCache.Stop()
	}

	if s.cacheManager.ExportService != nil {
		log.Println("  -> Export service")
		s.cacheManager.ExportService.Shutdown()
	}

	if s.cacheManager.MetricsCollector != nil {
		log.Println("  -> Metrics collector")
		s.cacheManager.MetricsCollector.Stop()
	}
}

// setupHealthChecker configures the health checker with all dependencies
func (s *Server) setupHealthChecker() {
	// Create alert system checker if database is available
	var alertSystemChecker *health.AlertSystemChecker
	if s.database != nil && s.database.Pool != nil && s.cacheManager != nil && s.cacheManager.AlertEngine != nil {
		webhookLogsQuerier := db.NewWebhookLogsQuerier(s.database.Pool)
		alertSuppressionsQuerier := db.NewAlertSuppressionsQuerier(s.database.Pool)
		alertSystemChecker = health.NewAlertSystemChecker(
			s.cacheManager.AlertEngine,
			webhookLogsQuerier,
			alertSuppressionsQuerier,
		)
		log.Println("[INFO] [Server] Alert system health checker initialized")
	}

	// Create health checker with all components
	var dbChecker health.Checker
	if s.database != nil {
		dbChecker = s.database
	}

	s.healthChecker = health.New(dbChecker, s.scheduler, alertSystemChecker)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	var dbPool *pgxpool.Pool
	if s.database != nil {
		dbPool = s.database.Pool
	}

	s.cacheManager = api.SetupRoutes(s.router, s.healthChecker, dbPool)
	log.Println("[INFO] [Server] Routes configured")
}

// setupSchedulerTasks registers all scheduled tasks
func (s *Server) setupSchedulerTasks() error {
	registry := NewTaskRegistry(s.scheduler, s.database)
	return registry.RegisterAll()
}

// setupHTTPServer creates and configures the HTTP server
func (s *Server) setupHTTPServer() {
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.Server.Port,
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}
}

// setupContext creates the shutdown context
func (s *Server) setupContext() {
	s.shutdownCtx, s.shutdownCancel = context.WithCancel(context.Background())
}

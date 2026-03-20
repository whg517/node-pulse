package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/api"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/health"
	"github.com/whg517/node-pulse/pulse/internal/scheduler"
	"github.com/whg517/node-pulse/pulse/pkg/telemetry"
)

// Server represents the HTTP server with all its dependencies
type Server struct {
	config            *Config
	router            *gin.Engine
	httpServer        *http.Server
	database          *db.Database
	healthChecker     *health.HealthChecker
	scheduler         scheduler.Scheduler
	cacheManager      *api.CacheManager
	telemetryProvider *telemetry.Provider
	shutdownCtx       context.Context
	shutdownCancel    context.CancelFunc
}

// Start starts the server and all its dependencies
func (s *Server) Start() error {
	slog.Info("Starting Node Pulse API", "component", "server", "port", s.config.Server.Port)

	// Start scheduler
	if err := s.scheduler.Start(s.shutdownCtx); err != nil {
		return err
	}
	slog.Info("Scheduler started", "component", "server")

	// Start HTTP server in background
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start HTTP server", "component", "server", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Node Pulse API started successfully", "component", "server")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	slog.Info("Shutting down", "component", "server")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop cache components
	if s.cacheManager != nil {
		s.stopCacheComponents()
	}

	// Stop scheduler
	if err := s.scheduler.Stop(); err != nil {
		slog.Warn("Error stopping scheduler", "component", "server", "error", err)
	}

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Warn("Forced to shutdown", "component", "server", "error", err)
		return err
	}

	// Close database
	if s.database != nil {
		s.database.Close()
	}

	// Flush and shutdown telemetry (must be after HTTP server to export remaining spans)
	if s.telemetryProvider != nil {
		s.telemetryProvider.Shutdown(ctx)
	}

	// Cancel shutdown context
	if s.shutdownCancel != nil {
		s.shutdownCancel()
	}

	slog.Info("Shutdown complete", "component", "server")
	return nil
}

// WaitForShutdown waits for interrupt signal and triggers graceful shutdown
func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := s.Shutdown(); err != nil {
		slog.Error("Shutdown error", "component", "server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// stopCacheComponents stops all cache-related components
func (s *Server) stopCacheComponents() {
	slog.Info("Stopping cache components", "component", "server")

	if s.cacheManager.AlertEngine != nil {
		slog.Debug("Stopping alert engine", "component", "server")
		s.cacheManager.AlertEngine.Stop()
	}

	if s.cacheManager.BatchWriter != nil {
		slog.Debug("Stopping batch writer", "component", "server")
		s.cacheManager.BatchWriter.Stop()
	}

	if s.cacheManager.MemoryCache != nil {
		slog.Debug("Stopping memory cache", "component", "server")
		s.cacheManager.MemoryCache.Stop()
	}

	if s.cacheManager.ExportService != nil {
		slog.Debug("Stopping export service", "component", "server")
		s.cacheManager.ExportService.Shutdown()
	}

	if s.cacheManager.MetricsCollector != nil {
		slog.Debug("Stopping metrics collector", "component", "server")
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
		slog.Info("Alert system health checker initialized", "component", "server")
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
	slog.Info("Routes configured", "component", "server")
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

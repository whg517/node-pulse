package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kevin/node-pulse/pulse-api/internal/api"
	"github.com/kevin/node-pulse/pulse-api/internal/cleanup"
	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/health"
	"github.com/kevin/node-pulse/pulse-api/internal/scheduler"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PULSE_PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	database, err := db.New(os.Getenv("DATABASE_URL"))
	var healthChecker *health.HealthChecker
	if err != nil {
		log.Printf("[WARN] Database connection failed: %v", err)
		// Continue without database for now, health check will report disabled
		// Pass nil directly to avoid interface nil behavior issues
		// Note: scheduler not created yet, will update health checker after scheduler init
		healthChecker = health.New(nil, nil)
	} else {
		defer database.Close()

		// Run migrations
		log.Println("[Migration] Running database migrations...")
		ctx := context.Background()
		if err := db.Migrate(ctx, database.Pool); err != nil {
			log.Fatalf("[Migration] Failed to run migrations: %v", err)
		}
		log.Println("[Migration] Database migrations completed successfully")

		// Note: scheduler not created yet, will update health checker after scheduler init
		healthChecker = health.New(database, nil)
	}

	// Initialize Gin router
	router := gin.Default()

	// Setup routes and get cache manager for shutdown
	cacheManager := api.SetupRoutes(router, healthChecker, database.Pool)

	// Initialize scheduler for background tasks (Story 3.12)
	sched, err := scheduler.NewScheduler()
	if err != nil {
		log.Fatalf("[Pulse] Failed to create scheduler: %v", err)
	}

	// Load cleanup configuration
	cleanupConfig, err := config.LoadCleanupConfig()
	if err != nil {
		log.Fatalf("[Pulse] Failed to load cleanup config: %v", err)
	}

	// Create and register cleanup task if enabled
	var cleanupTask *cleanup.CleanupTask
	if cleanupConfig.Enabled && database != nil && database.Pool != nil {
		cleanupTask, err = cleanup.NewCleanupTask(cleanupConfig, database.Pool, log.Default())
		if err != nil {
			log.Fatalf("[Pulse] Failed to create cleanup task: %v", err)
		}

		if cleanupTask != nil {
			if err := sched.RegisterTask(cleanupTask); err != nil {
				log.Fatalf("[Pulse] Failed to register cleanup task: %v", err)
			}
			log.Printf("[Pulse] Cleanup task registered (interval: %ds, retention: %ddays)",
				cleanupConfig.IntervalSeconds, cleanupConfig.RetentionDays)
		}
	}

	// Start scheduler in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		log.Fatalf("[Pulse] Failed to start scheduler: %v", err)
	}
	log.Println("[Pulse] Scheduler started")

	// Update health checker with scheduler reference
	healthChecker = health.New(
		func() health.Checker {
			if database != nil {
				return database
			}
			return nil
		}(),
		sched,
	)

	// Create server with timeout configuration
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("[Pulse] API server starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Pulse] Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[Pulse] Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop cache components (Story 3.2)
	if cacheManager != nil {
		log.Println("[Pulse] Stopping batch writer...")
		cacheManager.BatchWriter.Stop()
		log.Println("[Pulse] Stopping memory cache...")
		cacheManager.MemoryCache.Stop()
	}

	// Stop scheduler and cleanup task (Story 3.12)
	log.Println("[Pulse] Stopping scheduler...")
	if err := sched.Stop(); err != nil {
		log.Printf("[Pulse] Error stopping scheduler: %v", err)
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Pulse] Server forced to shutdown: %v", err)
	}

	log.Println("[Pulse] Server exited")
}

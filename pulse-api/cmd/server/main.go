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
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/health"
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
		healthChecker = health.New(nil)
	} else {
		defer database.Close()
		healthChecker = health.New(database)

		// Run migrations
		log.Println("[Migration] Running database migrations...")
		ctx := context.Background()
		if err := db.Migrate(ctx, database.Pool); err != nil {
			log.Fatalf("[Migration] Failed to run migrations: %v", err)
		}
		log.Println("[Migration] Database migrations completed successfully")
	}

	// Initialize Gin router
	router := gin.Default()

	// Setup routes and get cache manager for shutdown
	cacheManager := api.SetupRoutes(router, healthChecker, database.Pool)

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop cache components (Story 3.2)
	if cacheManager != nil {
		log.Println("[Pulse] Stopping batch writer...")
		cacheManager.BatchWriter.Stop()
		log.Println("[Pulse] Stopping memory cache...")
		cacheManager.MemoryCache.Stop()
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[Pulse] Server forced to shutdown: %v", err)
	}

	log.Println("[Pulse] Server exited")
}

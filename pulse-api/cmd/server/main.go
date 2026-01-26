package main

import (
	"context"
	"log"
	"os"

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

	// Setup routes
	api.SetupRoutes(router, healthChecker, database.Pool)

	// Start server
	log.Printf("[Pulse] API server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[Pulse] Failed to start server: %v", err)
	}
}

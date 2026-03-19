package main

import (
	"log/slog"
	"os"

	_ "github.com/whg517/node-pulse/pulse/api/docs" // Swagger docs
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/logger"
	"github.com/whg517/node-pulse/pulse/internal/server"
)

// @title			Node Pulse API
// @version		1.0
// @description	Node Pulse is a distributed monitoring system for tracking node health, metrics, and alerts in real-time.
// @description
// @description	## Features
// @description	- Real-time beacon heartbeat monitoring
// @description	- Multi-metric data collection and analysis
// @description	- Configurable alert rules with suppression
// @description	- Webhook notifications for alerts
// @description	- Historical data queries and export
// @description	- Role-based access control (RBAC)
// @description
// @description	## Authentication
// @description	Most endpoints require authentication. Use the `/api/v1/auth/login` endpoint to obtain a session token.
// @description	Include the token in the `Authorization` header: `Bearer <your-token>`
// @termsOfService	http://swagger.io/terms/

// @contact.name	API Support
// @contact.url		https://github.com/kevin/node-pulse
// @contact.email	support@example.com

// @license.name	MIT
// @license.url		https://opensource.org/licenses/MIT

// @host			localhost:6532
// @BasePath		/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name							Authorization
// @description					Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"

func main() {
	// Initialize structured logger from application configuration
	cfg := config.MustLoad()
	if err := logger.InitLogger(&cfg.Log); err != nil {
		// Fallback: log to stderr and continue with default slog settings
		slog.Error("Failed to initialize logger", "error", err)
		os.Exit(1)
	}

	// Build and start server using Builder pattern
	srv, err := server.NewBuilder().
		Build()
	if err != nil {
		slog.Error("Failed to build server", "error", err)
		os.Exit(1)
	}

	// Start the server
	if err := srv.Start(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	srv.WaitForShutdown()
}

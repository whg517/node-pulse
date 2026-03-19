package server

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/health"
	"github.com/whg517/node-pulse/pulse/internal/scheduler"
)

// Builder provides a fluent interface for building a Server
type Builder struct {
	config *Config
}

// NewBuilder creates a new Server builder
func NewBuilder() *Builder {
	return &Builder{
		config: DefaultConfig(),
	}
}

// WithPort sets the server port
func (b *Builder) WithPort(port string) *Builder {
	// Override port directly in config
	b.config.Server.Port = port
	return b
}

// WithDatabase sets up the database connection and migrations
func (b *Builder) WithDatabase(databaseURL string) *Builder {
	// Override database URL directly in config
	b.config.DB.URL = databaseURL
	return b
}

// Build constructs and initializes the Server
func (b *Builder) Build() (*Server, error) {
	// Initialize server
	srv := &Server{
		config: b.config,
		router: gin.Default(),
	}

	// Setup shutdown context
	srv.setupContext()

	// Initialize database
	if err := srv.initDatabase(); err != nil {
		slog.Warn("Database initialization failed, starting in DEGRADED MODE",
			"component", "server", "error", err)
		srv.healthChecker = health.New(nil, nil, nil)
	} else {
		// Run migrations
		if err := srv.runMigrations(); err != nil {
			return nil, err
		}

		// Initial health checker (will be updated later)
		srv.healthChecker = health.New(srv.database, nil, nil)
	}

	// Initialize scheduler
	if err := srv.initScheduler(); err != nil {
		return nil, err
	}

	// Setup routes
	srv.setupRoutes()

	// Setup health checker with all dependencies
	srv.setupHealthChecker()

	// Register scheduled tasks
	if err := srv.setupSchedulerTasks(); err != nil {
		return nil, err
	}

	// Setup HTTP server
	srv.setupHTTPServer()

	return srv, nil
}

// initDatabase initializes the database connection
func (s *Server) initDatabase() error {
	database, err := db.New(s.config.DB.URL)
	if err != nil {
		return err
	}

	s.database = database
	slog.Info("Database initialized", "component", "server")
	return nil
}

// runMigrations runs database migrations
func (s *Server) runMigrations() error {
	slog.Info("Running database migrations", "component", "server")
	ctx := context.Background()
	if err := db.Migrate(ctx, s.database.Pool); err != nil {
		slog.Error("Migration failed", "component", "server", "error", err)
		return err
	}
	slog.Info("Database migrations completed", "component", "server")
	return nil
}

// initScheduler initializes the task scheduler
func (s *Server) initScheduler() error {
	sched, err := scheduler.NewScheduler()
	if err != nil {
		return err
	}

	s.scheduler = sched
	slog.Info("Scheduler initialized", "component", "server")
	return nil
}

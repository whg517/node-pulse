package server

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/health"
	"github.com/kevin/node-pulse/pulse-api/internal/scheduler"
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
		log.Printf("[WARN] [Server] Database initialization failed: %v", err)
		log.Println("[WARN] [Server] Starting in DEGRADED MODE")
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
	log.Println("[INFO] [Server] Database initialized")
	return nil
}

// runMigrations runs database migrations
func (s *Server) runMigrations() error {
	log.Println("[INFO] [Server] Running database migrations...")
	ctx := context.Background()
	if err := db.Migrate(ctx, s.database.Pool); err != nil {
		log.Fatalf("[ERROR] [Server] Migration failed: %v", err)
		return err
	}
	log.Println("[INFO] [Server] Database migrations completed")
	return nil
}

// initScheduler initializes the task scheduler
func (s *Server) initScheduler() error {
	sched, err := scheduler.NewScheduler()
	if err != nil {
		return err
	}

	s.scheduler = sched
	log.Println("[INFO] [Server] Scheduler initialized")
	return nil
}

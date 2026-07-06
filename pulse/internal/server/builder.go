package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/health"
	"github.com/whg517/node-pulse/pulse/internal/scheduler"
	"github.com/whg517/node-pulse/pulse/pkg/telemetry"
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

	// Configure trusted proxies (O-G6). gin.Default() trusts everything by
	// default, which is wrong behind a reverse proxy — it would let any client
	// spoof its IP via X-Forwarded-For. When the operator lists the proxy
	// CIDRs (e.g. the nginx/Caddy host), only those parse the header and
	// c.ClientIP() / audit-log IPs reflect the real caller. An empty/nil list
	// keeps the legacy "trust all" behavior (fine for direct exposure).
	if proxies := b.config.Server.TrustedProxies; len(proxies) > 0 {
		if err := srv.router.SetTrustedProxies(proxies); err != nil {
			return nil, fmt.Errorf("invalid trusted_proxies: %w", err)
		}
		slog.Info("Configured trusted proxies", "component", "server", "proxies", proxies)
	} else {
		// Explicit nil → gin trusts all remote addrs (legacy default).
		_ = srv.router.SetTrustedProxies(nil)
	}

	// Setup shutdown context
	srv.setupContext()

	// Initialize OpenTelemetry tracing (must happen before routes are registered
	// so that the otelgin middleware picks up the correct global TracerProvider).
	if err := srv.initTelemetry(); err != nil {
		// Non-fatal: log a warning and continue without tracing
		slog.Warn("Telemetry initialization failed", "component", "server", "error", err)
	}

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

// initTelemetry initialises the global OpenTelemetry TracerProvider.
func (s *Server) initTelemetry() error {
	cfg := telemetry.Config{
		Enabled:        s.config.Telemetry.Enabled,
		ServiceName:    s.config.Telemetry.ServiceName,
		ServiceVersion: s.config.Telemetry.ServiceVersion,
		Environment:    s.config.Telemetry.Environment,
		OTLPEndpoint:   s.config.Telemetry.OTLPEndpoint,
		SamplingRate:   s.config.Telemetry.SamplingRate,
	}

	provider, err := telemetry.Init(s.shutdownCtx, cfg)
	if err != nil {
		return err
	}
	s.telemetryProvider = provider
	slog.Info("Telemetry initialized", "component", "server")
	return nil
}

// initDatabase initializes the database connection
func (s *Server) initDatabase() error {
	database, err := db.New(s.config.DB.URL, db.PoolOptions{
		MaxConnections:         s.config.DB.MaxConnections,
		MinConnections:         s.config.DB.MinConnections,
		MaxConnLifetimeSeconds: s.config.DB.ConnMaxLifetime,
		MaxConnIdleSeconds:     s.config.DB.ConnMaxIdleTime,
	})
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


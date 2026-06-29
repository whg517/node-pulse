package db

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database holds of database connection pool
type Database struct {
	Pool *pgxpool.Pool
}

// Check implements health.Checker interface
func (db *Database) Check(ctx context.Context) error {
	if db.Pool == nil {
		return os.ErrClosed
	}
	return db.Pool.Ping(ctx)
}

// PoolOptions configures the connection pool. Zero values fall back to sane
// defaults, so callers that don't care can pass PoolOptions{}.
type PoolOptions struct {
	// MaxConnections caps the pool size. Defaults to 25.
	MaxConnections int
	// MinConnections is the pool's warm size. Defaults to 2.
	MinConnections int
	// MaxConnLifetimeSeconds bounds how long a connection is reused. Defaults
	// to 1 hour.
	MaxConnLifetimeSeconds int
	// MaxConnIdleSeconds bounds how long an idle connection is kept. Defaults
	// to 5 minutes.
	MaxConnIdleSeconds int
}

// New creates a new database connection pool configured from opts.
func New(databaseURL string, opts PoolOptions) (*Database, error) {
	if databaseURL == "" {
		slog.Error("PULSE_DATABASE_URL is not set; server will start in DEGRADED MODE without database functionality",
			"component", "db",
			"hint", `set: export PULSE_DATABASE_URL="postgres://user:password@localhost:5432/dbname"`,
		)
		return nil, os.ErrInvalid
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		slog.Error("Failed to parse database URL; server will start in DEGRADED MODE",
			"component", "db", "error", err)
		return nil, err
	}

	// Apply options with sensible defaults. The previous code called
	// pgxpool.New(url) with no config, so pgx defaulted to
	// max(4, runtime.NumCPU()) — small enough to be exhausted under modest
	// concurrency (e.g. a websocket + several data queries in flight).
	if opts.MaxConnections > 0 {
		config.MaxConns = int32(opts.MaxConnections)
	} else {
		config.MaxConns = 25
	}
	if opts.MinConnections > 0 {
		config.MinConns = int32(opts.MinConnections)
	} else {
		config.MinConns = 2
	}
	if opts.MaxConnLifetimeSeconds > 0 {
		config.MaxConnLifetime = time.Duration(opts.MaxConnLifetimeSeconds) * time.Second
	} else {
		config.MaxConnLifetime = time.Hour
	}
	if opts.MaxConnIdleSeconds > 0 {
		config.MaxConnIdleTime = time.Duration(opts.MaxConnIdleSeconds) * time.Second
	} else {
		config.MaxConnIdleTime = 5 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		slog.Error("Failed to create connection pool; server will start in DEGRADED MODE",
			"component", "db", "error", err)
		return nil, err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("Failed to connect to database; check PULSE_DATABASE_URL and ensure the database server is running",
			"component", "db", "error", err)
		return nil, err
	}

	slog.Info("Database connection pool initialized",
		"component", "db",
		"max_conns", config.MaxConns,
		"min_conns", config.MinConns,
		"max_lifetime", config.MaxConnLifetime.String(),
		"max_idle", config.MaxConnIdleTime.String())
	return &Database{Pool: pool}, nil
}

// Close closes database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		slog.Info("Database connection pool closed", "component", "db")
	}
}

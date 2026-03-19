package db

import (
	"context"
	"log/slog"
	"os"

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

// New creates a new database connection
func New(databaseURL string) (*Database, error) {
	if databaseURL == "" {
		slog.Error("PULSE_DATABASE_URL is not set; server will start in DEGRADED MODE without database functionality",
			"component", "db",
			"hint", `set: export PULSE_DATABASE_URL="postgres://user:password@localhost:5432/dbname"`,
		)
		return nil, os.ErrInvalid
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
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

	slog.Info("Database connection pool initialized", "component", "db")
	return &Database{Pool: pool}, nil
}

// Close closes database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		slog.Info("Database connection pool closed", "component", "db")
	}
}

package db

import (
	"context"
	"log"
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
		log.Println("[DB] DATABASE_URL environment variable is not set")
		return nil, os.ErrInvalid
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("[DB] Database connection pool initialized successfully")
	return &Database{Pool: pool}, nil
}

// Close closes database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("[DB] Database connection pool closed")
	}
}

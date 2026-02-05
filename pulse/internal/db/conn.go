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
		log.Println("[ERROR] [DB] DATABASE_URL environment variable is not set")
		log.Println("[WARN] [DB] The server will start in DEGRADED MODE without database functionality")
		log.Println("[INFO] [DB] To enable database, set: export DATABASE_URL=\"postgres://user:password@localhost:5432/dbname\"")
		return nil, os.ErrInvalid
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Printf("[ERROR] [DB] Failed to create connection pool: %v", err)
		log.Println("[WARN] [DB] The server will start in DEGRADED MODE without database functionality")
		return nil, err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Printf("[ERROR] [DB] Failed to connect to database: %v", err)
		log.Println("[WARN] [DB] Please check your DATABASE_URL and ensure the database server is running")
		return nil, err
	}

	log.Println("[INFO] [DB] Database connection pool initialized successfully")
	return &Database{Pool: pool}, nil
}

// Close closes database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("[DB] Database connection pool closed")
	}
}

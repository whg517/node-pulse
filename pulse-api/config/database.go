package config

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DatabaseURL string
}

// NewDatabaseConfig creates a new database configuration from environment variables
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

// InitDB initializes a database connection pool
func InitDB(cfg *DatabaseConfig) (*pgxpool.Pool, error) {
	if cfg.DatabaseURL == "" {
		log.Println("[DB] DATABASE_URL environment variable is not set")
		return nil, os.ErrInvalid
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("Database connection pool initialized successfully")
	return pool, nil
}

// CloseDB closes database connection pool
func CloseDB(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
		log.Println("Database connection pool closed")
	}
}

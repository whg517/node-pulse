package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

// Migrate creates all database tables and indexes
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := createUsersTable(ctx, pool); err != nil {
		return err
	}

	if err := createSessionsTable(ctx, pool); err != nil {
		return err
	}

	if err := seedAdminUser(ctx, pool); err != nil {
		return err
	}

	return nil
}

// createUsersTable creates the users table with indexes
func createUsersTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			user_id UUID PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash VARCHAR(100) NOT NULL,
			role VARCHAR(20) NOT NULL,
			failed_login_attempts INTEGER DEFAULT 0,
			locked_until TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createSessionsTable creates the sessions table with indexes and foreign keys
func createSessionsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS sessions (
			session_id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			role VARCHAR(20) NOT NULL,
			expired_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_expired_at ON sessions(expired_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_expired ON sessions(user_id, expired_at DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// seedAdminUser creates the default admin user
func seedAdminUser(ctx context.Context, pool *pgxpool.Pool) error {
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // Default password for development
	}

	// Hash password with bcrypt (cost factor 12)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), 12)
	if err != nil {
		return err
	}

	// Check if admin user already exists
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE username = $1
		)
	`, adminUsername).Scan(&exists)

	if err != nil {
		return err
	}

	// Only create if admin user doesn't exist
	if !exists {
		adminUserID := uuid.New()
		query := `
			INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`

		_, err := pool.Exec(ctx, query, adminUserID, adminUsername, passwordHash, "admin")
		if err != nil {
			return err
		}

		log.Printf("[Migration] Admin user created: %s", adminUsername)
	} else {
		log.Printf("[Migration] Admin user already exists: %s", adminUsername)
	}

	return nil
}

package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver for golang-migrate
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"golang.org/x/crypto/bcrypt"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate runs the versioned SQL migrations embedded under migrations/, then
// performs idempotent data seeding that cannot be expressed in pure SQL
// (admin password hashing depends on runtime config + bcrypt).
//
// The migration history is tracked in the schema_migrations table managed by
// golang-migrate. This replaces the previous 27-step inline Go migration list
// with a versioned, reversible baseline (0001_init).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := runVersionedMigrations(pool); err != nil {
		return fmt.Errorf("versioned migrations: %w", err)
	}
	if err := Seed(ctx, pool); err != nil {
		return fmt.Errorf("data seeding: %w", err)
	}
	return nil
}

// runVersionedMigrations applies the embedded SQL migration files via
// golang-migrate, using the pool's DSN for a dedicated connection.
func runVersionedMigrations(pool *pgxpool.Pool) error {
	dsn := pool.Config().ConnString()

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			slog.Warn("Closing migrate instance failed", "component", "migration", "source_error", srcErr, "database_error", dbErr)
		}
	}()

	// Up applies all pending migrations; a no-op when already latest.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	version, dirty, vErr := m.Version()
	if vErr != nil && vErr != migrate.ErrNilVersion {
		slog.Warn("Could not read migration version", "component", "migration", "error", vErr)
	} else {
		slog.Info("Migrations applied", "component", "migration", "version", version, "dirty", dirty)
	}
	return nil
}

// Seed performs idempotent data seeding that depends on runtime configuration.
// It is separate from schema migrations because the admin password must be
// hashed with bcrypt from config, which cannot be done in a pure SQL file.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	return seedAdminUser(ctx, pool)
}

// seedAdminUser creates the initial admin user from configuration if it does
// not already exist. Idempotent across runs.
func seedAdminUser(ctx context.Context, pool *pgxpool.Pool) error {
	cfg := config.Get()

	adminUsername := cfg.Admin.Username
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminPassword := cfg.Admin.Password
	if adminPassword == "" {
		adminPassword = "Admin123" // Default password for development
	}

	// Validate admin password meets security requirements
	if err := auth.ValidatePassword(adminPassword); err != nil {
		slog.Warn("Admin password validation failed; using default admin password is NOT recommended for production",
			"component", "migration", "error", err)
		// Note: We don't fail here to allow development setups, but log a warning
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

		slog.Info("Admin user created", "component", "migration", "username", adminUsername)
	} else {
		slog.Info("Admin user already exists", "component", "migration", "username", adminUsername)
	}

	return nil
}

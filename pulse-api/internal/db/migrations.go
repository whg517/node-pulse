package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/auth"
)

// Migrate creates all database tables and indexes
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := createUsersTable(ctx, pool); err != nil {
		return err
	}

	if err := createSessionsTable(ctx, pool); err != nil {
		return err
	}

	if err := createNodesTable(ctx, pool); err != nil {
		return err
	}

	if err := addNodeStatusFields(ctx, pool); err != nil {
		return err
	}

	if err := createProbesTable(ctx, pool); err != nil {
		return err
	}

	if err := createProbesTrigger(ctx, pool); err != nil {
		return err
	}

	if err := createMetricsTable(ctx, pool); err != nil {
		return err
	}

	if err := createAlertsTable(ctx, pool); err != nil {
		return err
	}

	if err := createWebhooksTable(ctx, pool); err != nil {
		return err
	}

	if err := createAlertEventsTable(ctx, pool); err != nil {
		return err
	}

	if err := createAlertSuppressionsTable(ctx, pool); err != nil {
		return err
	}

	if err := createWebhookLogsTable(ctx, pool); err != nil {
		return err
	}

	if err := createAlertRecordsTable(ctx, pool); err != nil {
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

// createNodesTable creates nodes table with indexes
func createNodesTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS nodes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			ip VARCHAR(45) NOT NULL,
			region VARCHAR(100) NOT NULL,
			tags JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_nodes_region ON nodes(region);
		CREATE INDEX IF NOT EXISTS idx_nodes_created_at ON nodes(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_nodes_ip ON nodes(ip);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// seedAdminUser creates the default admin user
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
		log.Printf("[WARN] [Migration] Admin password validation failed: %v", err)
		log.Printf("[WARN] [Migration] Using default admin password is NOT recommended for production!")
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

		log.Printf("[Migration] Admin user created: %s", adminUsername)
	} else {
		log.Printf("[Migration] Admin user already exists: %s", adminUsername)
	}

	return nil
}

// addNodeStatusFields adds status tracking fields to nodes table
func addNodeStatusFields(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		DO $$
		BEGIN
			-- Add last_heartbeat column
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='nodes' AND column_name='last_heartbeat'
			) THEN
				ALTER TABLE nodes ADD COLUMN last_heartbeat TIMESTAMPTZ;
			END IF;

			-- Add last_report_time column
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='nodes' AND column_name='last_report_time'
			) THEN
				ALTER TABLE nodes ADD COLUMN last_report_time TIMESTAMPTZ;
			END IF;

			-- Add status column
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='nodes' AND column_name='status'
			) THEN
				ALTER TABLE nodes ADD COLUMN status VARCHAR(20) DEFAULT 'connecting';
				ALTER TABLE nodes ADD CONSTRAINT chk_node_status
					CHECK (status IN ('online', 'offline', 'connecting'));
			END IF;

			-- Add index on last_heartbeat for status queries
			IF NOT EXISTS (
				SELECT 1 FROM pg_indexes
				WHERE indexname = 'idx_nodes_last_heartbeat'
			) THEN
				CREATE INDEX idx_nodes_last_heartbeat ON nodes(last_heartbeat);
			END IF;
		END $$;
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createProbesTable creates probes table with indexes and foreign keys
func createProbesTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS probes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			type VARCHAR(10) NOT NULL CHECK (type IN ('TCP', 'UDP')),
			target VARCHAR(255) NOT NULL,
			port INTEGER NOT NULL CHECK (port >= 1 AND port <= 65535),
			interval_seconds INTEGER NOT NULL CHECK (interval_seconds >= 60 AND interval_seconds <= 300),
			count INTEGER NOT NULL CHECK (count >= 1 AND count <= 100),
			timeout_seconds INTEGER NOT NULL CHECK (timeout_seconds >= 1 AND timeout_seconds <= 30),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_probes_node_id ON probes(node_id);
		CREATE INDEX IF NOT EXISTS idx_probes_type ON probes(type);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createMetricsTable creates metrics table with indexes and foreign keys
func createMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS metrics (
			id BIGSERIAL PRIMARY KEY,
			node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			probe_id UUID NOT NULL REFERENCES probes(id) ON DELETE CASCADE,
			timestamp TIMESTAMPTZ NOT NULL,
			latency_ms DECIMAL(10,2),
			packet_loss_rate DECIMAL(5,4),
			jitter_ms DECIMAL(10,2),
			is_aggregated BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_metrics_node_timestamp ON metrics(node_id, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_metrics_probe_timestamp ON metrics(probe_id, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_metrics_aggregated ON metrics(is_aggregated, timestamp DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createProbesTrigger creates a trigger to auto-update updated_at on probes table
func createProbesTrigger(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		-- Drop trigger if exists
		DROP TRIGGER IF EXISTS update_probes_updated_at ON probes;

		-- Create trigger function
		CREATE OR REPLACE FUNCTION update_probes_updated_at_func()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		-- Create trigger
		CREATE TRIGGER update_probes_updated_at
		BEFORE UPDATE ON probes
		FOR EACH ROW
		EXECUTE FUNCTION update_probes_updated_at_func();
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createAlertsTable creates alerts table with indexes and foreign keys (Story 5.1)
func createAlertsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS alerts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			metric VARCHAR NOT NULL CHECK (metric IN ('latency', 'packet_loss_rate', 'jitter')),
			threshold DECIMAL(10,2) NOT NULL CHECK (threshold > 0),
			level VARCHAR NOT NULL CHECK (level IN ('P0', 'P1', 'P2')),
			node_id UUID REFERENCES nodes(id) ON DELETE CASCADE,
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_alerts_node_id ON alerts(node_id);
		CREATE INDEX IF NOT EXISTS idx_alerts_enabled ON alerts(enabled);
		CREATE INDEX IF NOT EXISTS idx_alerts_metric ON alerts(metric);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createWebhooksTable creates webhooks table with indexes (Story 5.2)
// Note: HTTPS validation is handled at application layer (api/webhook_handler.go)
func createWebhooksTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS webhooks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			url VARCHAR NOT NULL,
			event_format JSONB,
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON webhooks(enabled);
		CREATE INDEX IF NOT EXISTS idx_webhooks_url ON webhooks(url);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createAlertEventsTable creates alert_events table with indexes (Story 5.5)
func createAlertEventsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS alert_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			metric VARCHAR NOT NULL,
			threshold DECIMAL NOT NULL,
			current_value DECIMAL NOT NULL,
			level VARCHAR NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_alert_events_node_id ON alert_events(node_id);
		CREATE INDEX IF NOT EXISTS idx_alert_events_metric ON alert_events(metric);
		CREATE INDEX IF NOT EXISTS idx_alert_events_created_at ON alert_events(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_alert_events_node_created ON alert_events(node_id, created_at DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createAlertSuppressionsTable creates alert_suppressions table with indexes (Story 5.6)
func createAlertSuppressionsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS alert_suppressions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			metric VARCHAR NOT NULL,
			suppressed_until TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(node_id, metric)
		);

		CREATE INDEX IF NOT EXISTS idx_alert_suppressions_node_metric ON alert_suppressions(node_id, metric);
		CREATE INDEX IF NOT EXISTS idx_alert_suppressions_until ON alert_suppressions(suppressed_until);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createWebhookLogsTable creates webhook_logs table with indexes (Story 5.7)
func createWebhookLogsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS webhook_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
			alert_event_id UUID NOT NULL REFERENCES alert_events(id) ON DELETE CASCADE,
			status VARCHAR NOT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			error_message TEXT,
			sent_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_webhook_logs_webhook_id ON webhook_logs(webhook_id);
		CREATE INDEX IF NOT EXISTS idx_webhook_logs_alert_event_id ON webhook_logs(alert_event_id);
		CREATE INDEX IF NOT EXISTS idx_webhook_logs_status ON webhook_logs(status);
		CREATE INDEX IF NOT EXISTS idx_webhook_logs_created_at ON webhook_logs(created_at DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createAlertRecordsTable creates alert_records table with indexes (Story 6.1)
func createAlertRecordsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS alert_records (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			alert_event_id UUID NOT NULL REFERENCES alert_events(id) ON DELETE CASCADE,
			node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			metric VARCHAR NOT NULL,
			level VARCHAR NOT NULL,
			status VARCHAR NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			CONSTRAINT chk_alert_record_status CHECK (status IN ('pending', 'in_progress', 'resolved'))
		);

		CREATE INDEX IF NOT EXISTS idx_alert_records_node_id ON alert_records(node_id);
		CREATE INDEX IF NOT EXISTS idx_alert_records_level ON alert_records(level);
		CREATE INDEX IF NOT EXISTS idx_alert_records_status ON alert_records(status);
		CREATE INDEX IF NOT EXISTS idx_alert_records_created_at ON alert_records(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_alert_records_node_created ON alert_records(node_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_alert_records_status_created ON alert_records(status, created_at DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// Migrate creates all database tables and indexes
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	steps := []struct {
		name string
		fn   func(context.Context, *pgxpool.Pool) error
	}{
		{"createUsersTable", createUsersTable},
		{"createSessionsTable", createSessionsTable},
		{"dropAndRecreateRefreshTokensTable", dropAndRecreateRefreshTokensTable},
		{"dropAndRecreateBeaconTokensAsAPIKeys", dropAndRecreateBeaconTokensAsAPIKeys},
		{"createTokenBlacklistTable", createTokenBlacklistTable},
		{"createAuthAuditLogsTable", createAuthAuditLogsTable},
		{"createRateLimitsTable", createRateLimitsTable},
		{"createPasswordResetTokensTable", createPasswordResetTokensTable},
		{"createRolesTable", createRolesTable},
		{"createPermissionsTable", createPermissionsTable},
		{"createRolePermissionsTable", createRolePermissionsTable},
		{"createNodesTable", createNodesTable},
		{"addNodeStatusFields", addNodeStatusFields},
		{"createProbesTable", createProbesTable},
		{"createProbesTrigger", createProbesTrigger},
		{"createBeaconConfigTables", createBeaconConfigTables},
		{"createMTRResultsTable", createMTRResultsTable},
		{"createMetricsTable", createMetricsTable},
		{"createAlertsTable", createAlertsTable},
		{"createWebhooksTable", createWebhooksTable},
		{"createAlertEventsTable", createAlertEventsTable},
		{"createAlertSuppressionsTable", createAlertSuppressionsTable},
		{"createWebhookLogsTable", createWebhookLogsTable},
		{"createAlertRecordsTable", createAlertRecordsTable},
		{"createAlertStatusHistoryTable", createAlertStatusHistoryTable},
		{"createAlertNotesTable", createAlertNotesTable},
		{"seedAdminUser", seedAdminUser},
	}

	for _, step := range steps {
		slog.Info("Running migration step", "component", "migration", "step", step.name)
		if err := step.fn(ctx, pool); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}
	}

	return nil
}

// createUsersTable creates the users table with indexes
func createUsersTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			user_id UUID PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(255) UNIQUE,
			password_hash VARCHAR(100) NOT NULL,
			role VARCHAR(20) NOT NULL,
			failed_login_attempts INTEGER DEFAULT 0,
			locked_until TIMESTAMPTZ,
			mfa_enabled BOOLEAN DEFAULT false,
			mfa_secret TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		-- Add mfa_enabled/mfa_secret columns if missing (schema upgrade from older test-schema versions)
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='users' AND column_name='mfa_enabled'
			) THEN
				ALTER TABLE users ADD COLUMN mfa_enabled BOOLEAN DEFAULT false;
			END IF;

			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='users' AND column_name='mfa_secret'
			) THEN
				ALTER TABLE users ADD COLUMN mfa_secret TEXT;
			END IF;
		END $$;

		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_locked ON users(locked_until) WHERE locked_until IS NOT NULL;
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createSessionsTable creates the sessions table with indexes and foreign keys
func createSessionsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS sessions (
			session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			device_id VARCHAR(255),
			ip_address INET,
			user_agent TEXT,
			remember_me BOOLEAN DEFAULT false,
			expires_at TIMESTAMPTZ NOT NULL,
			max_valid_until TIMESTAMPTZ NOT NULL,
			last_activity_at TIMESTAMPTZ DEFAULT NOW(),
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_expired ON sessions(user_id, expires_at DESC);
		CREATE INDEX IF NOT EXISTS idx_sessions_device ON sessions(device_id) WHERE device_id IS NOT NULL;
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// dropAndRecreateRefreshTokensTable drops and recreates refresh_tokens with new schema for JWT rewrite
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

// createMTRResultsTable creates tables for route-hop/MTR snapshots reported by beacons.
func createMTRResultsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS mtr_results (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			probe_id VARCHAR(255),
			target VARCHAR(255) NOT NULL,
			success BOOLEAN NOT NULL,
			total_hops INTEGER NOT NULL DEFAULT 0 CHECK (total_hops >= 0),
			hops JSONB NOT NULL DEFAULT '[]',
			completed_at TIMESTAMPTZ NOT NULL,
			error_message TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_mtr_results_node_completed
			ON mtr_results(node_id, completed_at DESC);
		CREATE INDEX IF NOT EXISTS idx_mtr_results_target
			ON mtr_results(target);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createBeaconConfigTables creates persistent beacon config and history tables.
func createBeaconConfigTables(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS beacon_configs (
			beacon_id UUID PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
			probes JSONB NOT NULL DEFAULT '[]',
			interval_seconds INTEGER NOT NULL DEFAULT 60 CHECK (interval_seconds >= 5),
			timeout_seconds INTEGER NOT NULL DEFAULT 5 CHECK (timeout_seconds >= 1),
			version INTEGER NOT NULL DEFAULT 1 CHECK (version >= 1),
			last_ack_version INTEGER,
			last_ack_at TIMESTAMPTZ,
			last_ack_status VARCHAR(20),
			last_ack_error TEXT,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='beacon_configs' AND column_name='last_ack_version'
			) THEN
				ALTER TABLE beacon_configs ADD COLUMN last_ack_version INTEGER;
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='beacon_configs' AND column_name='last_ack_at'
			) THEN
				ALTER TABLE beacon_configs ADD COLUMN last_ack_at TIMESTAMPTZ;
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='beacon_configs' AND column_name='last_ack_status'
			) THEN
				ALTER TABLE beacon_configs ADD COLUMN last_ack_status VARCHAR(20);
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='beacon_configs' AND column_name='last_ack_error'
			) THEN
				ALTER TABLE beacon_configs ADD COLUMN last_ack_error TEXT;
			END IF;
		END $$;

		CREATE TABLE IF NOT EXISTS beacon_config_history (
			id BIGSERIAL PRIMARY KEY,
			beacon_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			version INTEGER NOT NULL,
			config JSONB NOT NULL,
			changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			changed_by VARCHAR(255) NOT NULL DEFAULT 'system'
		);

		CREATE INDEX IF NOT EXISTS idx_beacon_config_history_beacon_changed
			ON beacon_config_history(beacon_id, changed_at DESC);
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

// createAlertStatusHistoryTable creates alert_status_history for lifecycle timelines.
func createAlertStatusHistoryTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS alert_status_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			alert_id UUID NOT NULL REFERENCES alert_records(id) ON DELETE CASCADE,
			from_status VARCHAR,
			to_status VARCHAR NOT NULL,
			user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
			user_name VARCHAR(255) NOT NULL DEFAULT 'System',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			CONSTRAINT chk_alert_status_history_from_status CHECK (from_status IS NULL OR from_status IN ('pending', 'in_progress', 'resolved')),
			CONSTRAINT chk_alert_status_history_to_status CHECK (to_status IN ('pending', 'in_progress', 'resolved'))
		);

		CREATE INDEX IF NOT EXISTS idx_alert_status_history_alert_id ON alert_status_history(alert_id, created_at ASC);
		CREATE INDEX IF NOT EXISTS idx_alert_status_history_user_id ON alert_status_history(user_id);
		CREATE INDEX IF NOT EXISTS idx_alert_status_history_created_at ON alert_status_history(created_at DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createAlertNotesTable creates alert_notes table for operator investigation notes.
func createAlertNotesTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS alert_notes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			alert_id UUID NOT NULL REFERENCES alert_records(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
			user_name VARCHAR(255) NOT NULL DEFAULT 'System',
			content TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			CONSTRAINT chk_alert_note_content_not_empty CHECK (length(trim(content)) > 0)
		);

		CREATE INDEX IF NOT EXISTS idx_alert_notes_alert_id ON alert_notes(alert_id, created_at ASC);
		CREATE INDEX IF NOT EXISTS idx_alert_notes_user_id ON alert_notes(user_id);
		CREATE INDEX IF NOT EXISTS idx_alert_notes_created_at ON alert_notes(created_at DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// dropAndRecreateRefreshTokensTable drops and recreates refresh_tokens with new schema for JWT rewrite
func dropAndRecreateRefreshTokensTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		-- Drop existing table
		DROP TABLE IF EXISTS refresh_tokens CASCADE;

		-- Create new table with sliding + absolute expiration support
		CREATE TABLE refresh_tokens (
			id SERIAL PRIMARY KEY,
			token_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
			token_hash TEXT NOT NULL UNIQUE,
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			session_id UUID REFERENCES sessions(session_id) ON DELETE CASCADE,
			expires_at TIMESTAMPTZ NOT NULL,
			max_valid_until TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ,
			replaced_by UUID REFERENCES refresh_tokens(token_id),
			user_agent TEXT,
			ip_address INET,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_revoked ON refresh_tokens(user_id, revoked_at);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_id ON refresh_tokens(token_id);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session ON refresh_tokens(session_id) WHERE session_id IS NOT NULL;
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// dropAndRecreateBeaconTokensAsAPIKeys drops beacon_tokens and creates api_keys table
func dropAndRecreateBeaconTokensAsAPIKeys(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		-- Drop existing table
		DROP TABLE IF EXISTS beacon_tokens CASCADE;
		DROP TABLE IF EXISTS api_keys CASCADE;
		DROP TABLE IF EXISTS service_accounts CASCADE;

		-- Create service_accounts table first
		CREATE TABLE service_accounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			scopes JSONB NOT NULL DEFAULT '[]',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			expires_at TIMESTAMPTZ
		);

		-- Create renamed table with improved schema
		CREATE TABLE api_keys (
			id SERIAL PRIMARY KEY,
			key_id VARCHAR(20) NOT NULL UNIQUE,
			key_hash TEXT NOT NULL UNIQUE,
			key_prefix VARCHAR(20) NOT NULL,
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			service_account_id UUID REFERENCES service_accounts(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			is_active BOOLEAN DEFAULT true,
			expires_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			last_used_at TIMESTAMPTZ,
			CONSTRAINT chk_owner_xor CHECK (
				(user_id IS NOT NULL)::integer + (service_account_id IS NOT NULL)::integer = 1
			)
		);

		CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
		CREATE INDEX IF NOT EXISTS idx_api_keys_service_account_id ON api_keys(service_account_id);
		CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
		CREATE INDEX IF NOT EXISTS idx_api_keys_active_expires ON api_keys(is_active, expires_at);
		CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createTokenBlacklistTable creates token_blacklist for immediate access token revocation
func createTokenBlacklistTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS token_blacklist (
			jti TEXT PRIMARY KEY,
			user_id UUID REFERENCES users(user_id),
			revoked_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			reason TEXT
		);

		-- Add user_id column if missing (schema upgrade from older versions)
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='token_blacklist' AND column_name='user_id'
			) THEN
				ALTER TABLE token_blacklist ADD COLUMN user_id UUID REFERENCES users(user_id);
			END IF;
		END $$;

		CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires ON token_blacklist(expires_at);
		CREATE INDEX IF NOT EXISTS idx_token_blacklist_user ON token_blacklist(user_id, expires_at);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createAuthAuditLogsTable creates auth_audit_logs for security event tracking
func createAuthAuditLogsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS auth_audit_logs (
			id BIGSERIAL PRIMARY KEY,
			event_type VARCHAR(100) NOT NULL,
			user_id UUID REFERENCES users(user_id),
			service_account_id UUID REFERENCES service_accounts(id),
			session_id UUID REFERENCES sessions(session_id),
			ip_address INET,
			user_agent TEXT,
			details JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		-- Add service_account_id column if missing (schema upgrade from older versions)
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='auth_audit_logs' AND column_name='service_account_id'
			) THEN
				ALTER TABLE auth_audit_logs ADD COLUMN service_account_id UUID REFERENCES service_accounts(id);
			END IF;
		END $$;

		CREATE INDEX IF NOT EXISTS idx_audit_events ON auth_audit_logs(event_type, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_users ON auth_audit_logs(user_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_ips ON auth_audit_logs(ip_address, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_service_accounts ON auth_audit_logs(service_account_id, created_at DESC);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createRateLimitsTable creates rate_limits for database-backed rate limiting
func createRateLimitsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS rate_limits (
			id BIGSERIAL PRIMARY KEY,
			key TEXT NOT NULL,
			window_type VARCHAR(10) NOT NULL,
			window_start TIMESTAMPTZ NOT NULL,
			request_count INTEGER DEFAULT 1,
			UNIQUE(key, window_type, window_start)
		);

		CREATE INDEX IF NOT EXISTS idx_rate_limits_lookup ON rate_limits(key, window_type, window_start);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createRolesTable creates the roles table for RBAC
func createRolesTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS roles (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			is_system_role BOOLEAN DEFAULT false,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		-- Seed default roles
		INSERT INTO roles (name, description, is_system_role) VALUES
			('admin', 'Full system access', true),
			('operator', 'Can manage nodes and view alerts', true),
			('viewer', 'Read-only access', true)
		ON CONFLICT (name) DO NOTHING;
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createPermissionsTable creates the permissions table for RBAC
func createPermissionsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS permissions (
			id SERIAL PRIMARY KEY,
			resource VARCHAR(50) NOT NULL,
			action VARCHAR(20) NOT NULL,
			description TEXT,
			UNIQUE(resource, action)
		);

		-- Seed default permissions
		INSERT INTO permissions (resource, action, description) VALUES
			('nodes', 'read', 'View nodes'),
			('nodes', 'write', 'Create and update nodes'),
			('nodes', 'delete', 'Delete nodes'),
			('probes', 'read', 'View probes'),
			('probes', 'write', 'Create and update probes'),
			('probes', 'delete', 'Delete probes'),
			('alerts', 'read', 'View alerts'),
			('alerts', 'write', 'Create and update alerts'),
			('alerts', 'delete', 'Delete alerts'),
			('webhooks', 'read', 'View webhooks'),
			('webhooks', 'write', 'Create and update webhooks'),
			('webhooks', 'delete', 'Delete webhooks'),
			('users', 'read', 'View users'),
			('users', 'write', 'Create and update users'),
			('users', 'delete', 'Delete users'),
			('settings', 'read', 'View system settings'),
			('settings', 'write', 'Update system settings'),
			('reports', 'read', 'View reports'),
			('reports', 'write', 'Generate reports'),
			('api_keys', 'read', 'View API keys'),
			('api_keys', 'write', 'Create and update API keys'),
			('api_keys', 'delete', 'Delete API keys')
		ON CONFLICT (resource, action) DO NOTHING;
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createRolePermissionsTable creates the role_permissions junction table for RBAC
func createRolePermissionsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS role_permissions (
			role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
			permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
			granted_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (role_id, permission_id)
		);

		-- Seed default role permissions
		-- Admin: all permissions
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
		WHERE r.name = 'admin'
		ON CONFLICT DO NOTHING;

		-- Operator: nodes, probes, alerts, webhooks (read/write), reports (read/write)
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r
		CROSS JOIN permissions p
		WHERE r.name = 'operator'
			AND p.resource IN ('nodes', 'probes', 'alerts', 'webhooks', 'reports')
			AND p.action IN ('read', 'write')
		ON CONFLICT DO NOTHING;

		-- Viewer: read-only access to all resources except users and settings
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r
		CROSS JOIN permissions p
		WHERE r.name = 'viewer'
			AND p.action = 'read'
			AND p.resource NOT IN ('users', 'settings')
		ON CONFLICT DO NOTHING;
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// createPasswordResetTokensTable creates password_reset_tokens table with indexes
func createPasswordResetTokensTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			used_at TIMESTAMPTZ,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_hash ON password_reset_tokens(token_hash);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

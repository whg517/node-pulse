package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditLogger records security events to the database
type AuditLogger struct {
	pool *pgxpool.Pool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(pool *pgxpool.Pool) *AuditLogger {
	return &AuditLogger{pool: pool}
}

// LogEvent records an audit event to the database
func (a *AuditLogger) LogEvent(ctx context.Context, eventType string, userID *uuid.UUID, ipAddress string, details map[string]interface{}) error {
	// Serialize details to JSONB
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	// Insert audit log entry
	_, err = a.pool.Exec(ctx, `
		INSERT INTO auth_audit_logs (event_type, user_id, ip_address, details, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, eventType, userID, ipAddress, detailsJSON)

	return err
}

// Audit event types
const (
	// Login events
	EventLoginSuccess      = "login_success"
	EventLoginFailed       = "login_failed"
	EventLoginLocked       = "login_locked"
	EventRateLimitExceeded = "rate_limit_exceeded"

	// Token events
	EventTokenGenerated   = "token_generated"
	EventTokenRefreshed   = "token_refreshed"
	EventTokenRevoked     = "token_revoked"
	EventTokenBlacklisted = "token_blacklisted"

	// Session events
	EventSessionCreated  = "session_created"
	EventSessionRevoked  = "session_revoked"
	EventSessionListed   = "session_listed"
	EventAllSessionsRevoked = "all_sessions_revoked"

	// API key events
	EventAPIKeyGenerated = "api_key_generated"
	EventAPIKeyUsed      = "api_key_used"
	EventAPIKeyRevoked   = "api_key_revoked"

	// Admin events
	EventAdminRevokeAll = "admin_revoke_all"
)

// Helper functions for common audit events

// LogLoginSuccess records a successful login
func (a *AuditLogger) LogLoginSuccess(ctx context.Context, userID uuid.UUID, ipAddress string) {
	details := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	a.LogEvent(ctx, EventLoginSuccess, &userID, ipAddress, details)
}

// LogLoginFailed records a failed login attempt
func (a *AuditLogger) LogLoginFailed(ctx context.Context, username, ipAddress string) {
	details := map[string]interface{}{
		"username":  username,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	// userID is nil for failed logins
	a.LogEvent(ctx, EventLoginFailed, nil, ipAddress, details)
}

// LogLoginLocked records an account lockout
func (a *AuditLogger) LogLoginLocked(ctx context.Context, userID uuid.UUID, ipAddress string, failedAttempts int) {
	details := map[string]interface{}{
		"failed_attempts": failedAttempts,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}
	a.LogEvent(ctx, EventLoginLocked, &userID, ipAddress, details)
}

// LogTokenRefreshed records a token refresh
func (a *AuditLogger) LogTokenRefreshed(ctx context.Context, userID uuid.UUID, ipAddress string) {
	details := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	a.LogEvent(ctx, EventTokenRefreshed, &userID, ipAddress, details)
}

// LogTokenRevoked records a token revocation (logout)
func (a *AuditLogger) LogTokenRevoked(ctx context.Context, userID uuid.UUID, ipAddress string, jti string) {
	details := map[string]interface{}{
		"jti":       jti,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	a.LogEvent(ctx, EventTokenRevoked, &userID, ipAddress, details)
}

// LogAllSessionsRevoked records admin revocation of all user sessions
func (a *AuditLogger) LogAllSessionsRevoked(ctx context.Context, targetUserID uuid.UUID, adminUserID uuid.UUID, ipAddress string) {
	details := map[string]interface{}{
		"target_user_id": targetUserID.String(),
		"admin_user_id":  adminUserID.String(),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
	a.LogEvent(ctx, EventAdminRevokeAll, &adminUserID, ipAddress, details)
}

// LogRateLimitExceeded records a rate limit violation
func (a *AuditLogger) LogRateLimitExceeded(ctx context.Context, userID *uuid.UUID, ipAddress, endpoint string) {
	details := map[string]interface{}{
		"endpoint":  endpoint,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	a.LogEvent(ctx, EventRateLimitExceeded, userID, ipAddress, details)
}

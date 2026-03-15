package auth

import (
	"context"
	"encoding/json"
	"fmt"
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
	_ = a.LogEvent(ctx, EventLoginSuccess, &userID, ipAddress, details)
}

// LogLoginFailed records a failed login attempt
func (a *AuditLogger) LogLoginFailed(ctx context.Context, username, ipAddress string) {
	details := map[string]interface{}{
		"username":  username,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	// userID is nil for failed logins
	_ = a.LogEvent(ctx, EventLoginFailed, nil, ipAddress, details)
}

// LogLoginLocked records an account lockout
func (a *AuditLogger) LogLoginLocked(ctx context.Context, userID uuid.UUID, ipAddress string, failedAttempts int) {
	details := map[string]interface{}{
		"failed_attempts": failedAttempts,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}
	_ = a.LogEvent(ctx, EventLoginLocked, &userID, ipAddress, details)
}

// LogTokenRefreshed records a token refresh
func (a *AuditLogger) LogTokenRefreshed(ctx context.Context, userID uuid.UUID, ipAddress string) {
	details := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	_ = a.LogEvent(ctx, EventTokenRefreshed, &userID, ipAddress, details)
}

// LogTokenRevoked records a token revocation (logout)
func (a *AuditLogger) LogTokenRevoked(ctx context.Context, userID uuid.UUID, ipAddress string, jti string) {
	details := map[string]interface{}{
		"jti":       jti,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	_ = a.LogEvent(ctx, EventTokenRevoked, &userID, ipAddress, details)
}

// LogAllSessionsRevoked records admin revocation of all user sessions
func (a *AuditLogger) LogAllSessionsRevoked(ctx context.Context, targetUserID uuid.UUID, adminUserID uuid.UUID, ipAddress string) {
	details := map[string]interface{}{
		"target_user_id": targetUserID.String(),
		"admin_user_id":  adminUserID.String(),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
	_ = a.LogEvent(ctx, EventAdminRevokeAll, &adminUserID, ipAddress, details)
}

// LogRateLimitExceeded records a rate limit violation
func (a *AuditLogger) LogRateLimitExceeded(ctx context.Context, userID *uuid.UUID, ipAddress, endpoint string) {
	details := map[string]interface{}{
		"endpoint":  endpoint,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	_ = a.LogEvent(ctx, EventRateLimitExceeded, userID, ipAddress, details)
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	EventType string
	UserID    *string
	IPAddress *string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

// QueryAuditLogs retrieves audit logs with filtering and pagination
func (a *AuditLogger) QueryAuditLogs(ctx context.Context, filters AuditLogFilter) ([]map[string]interface{}, int, error) {
	// Build dynamic query
	baseQuery := `
		SELECT id, event_type, user_id, ip_address, details, created_at
		FROM auth_audit_logs
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM auth_audit_logs WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if filters.EventType != "" {
		baseQuery += fmt.Sprintf(" AND event_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, filters.EventType)
		argCount++
	}

	if filters.UserID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filters.UserID)
		argCount++
	}

	if filters.IPAddress != nil {
		baseQuery += fmt.Sprintf(" AND ip_address = $%d", argCount)
		countQuery += fmt.Sprintf(" AND ip_address = $%d", argCount)
		args = append(args, *filters.IPAddress)
		argCount++
	}

	if filters.StartTime != nil {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.StartTime)
		argCount++
	}

	if filters.EndTime != nil {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.EndTime)
		argCount++
	}

	// Get total count
	var totalCount int
	err := a.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	baseQuery += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
		argCount++
	}

	if filters.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	rows, err := a.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var eventType, userID string
		var ipAddress *string
		var detailsJSON []byte
		var createdAt time.Time

		err := rows.Scan(&id, &eventType, &userID, &ipAddress, &detailsJSON, &createdAt)
		if err != nil {
			return nil, 0, err
		}

		// Parse JSONB details
		var details map[string]interface{}
		if detailsJSON != nil {
			err = json.Unmarshal(detailsJSON, &details)
			if err != nil {
				return nil, 0, err
			}
		}

		result := map[string]interface{}{
			"id":         id,
			"event_type": eventType,
			"user_id":    userID,
			"ip_address": ipAddress,
			"details":    details,
			"created_at": createdAt,
		}
		results = append(results, result)
	}

	return results, totalCount, nil
}

// GetAuditLogByID retrieves a single audit log entry by ID
func (a *AuditLogger) GetAuditLogByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	var eventType, userID string
	var ipAddress *string
	var detailsJSON []byte
	var createdAt time.Time

	err := a.pool.QueryRow(ctx, `
		SELECT event_type, user_id, ip_address, details, created_at
		FROM auth_audit_logs
		WHERE id = $1
	`, id).Scan(&eventType, &userID, &ipAddress, &detailsJSON, &createdAt)

	if err != nil {
		return nil, err
	}

	// Parse JSONB details
	var details map[string]interface{}
	if detailsJSON != nil {
		err = json.Unmarshal(detailsJSON, &details)
		if err != nil {
			return nil, err
		}
	}

	log := map[string]interface{}{
		"id":         id,
		"event_type": eventType,
		"user_id":    userID,
		"ip_address": ipAddress,
		"details":    details,
		"created_at": createdAt,
	}

	return log, nil
}

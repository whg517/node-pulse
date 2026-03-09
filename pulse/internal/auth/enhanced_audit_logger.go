package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnhancedAuditLogger provides comprehensive audit logging with additional event types
type EnhancedAuditLogger struct {
	pool *pgxpool.Pool
}

// NewEnhancedAuditLogger creates a new enhanced audit logger
func NewEnhancedAuditLogger(pool *pgxpool.Pool) *EnhancedAuditLogger {
	return &EnhancedAuditLogger{pool: pool}
}

// Additional audit event types beyond the base AuditLogger
const (
	// Authorization events
	EventAccessDenied         = "access_denied"
	EventPrivilegeEscalation  = "privilege_escalation_attempt"
	EventUnauthorizedAction    = "unauthorized_action_attempt"

	// Security events
	EventSuspiciousActivity    = "suspicious_activity"
	EventAnomalousLocation     = "anomalous_location_login"
	EventConcurrentSession     = "concurrent_session_detected"
	EventTokenReuse            = "token_reuse_detected"

	// Admin events
	EventUserCreated           = "admin_user_created"
	EventUserDeleted           = "admin_user_deleted"
	EventRoleChanged           = "admin_role_changed"
	EventConfigModified        = "admin_config_modified"
	EventPasswordReset         = "admin_password_reset"
	EventRateLimitReset        = "admin_rate_limit_reset"

	// API key events
	EventAPIKeyCreated         = "admin_api_key_created"
	EventAPIKeyDeleted         = "admin_api_key_deleted"
	EventAPIKeyRotated         = "admin_api_key_rotated"
	EventAPIKeyCompromised     = "api_key_compromised"

	// Beacon events
	EventBeaconRegistered      = "beacon_registered"
	EventBeaconUnregistered    = "beacon_unregistered"
	EventBeaconCertExpiring    = "beacon_certificate_expiring"
	EventBeaconCertExpired     = "beacon_certificate_expired"

	// System events
	EventSecurityAlert         = "security_alert"
	EventSystemBreach          = "system_breach_detected"
	EventDataExfil             = "data_exfiltration_attempt"
)

// LogAccessDenied records an access denied event
func (e *EnhancedAuditLogger) LogAccessDenied(
	ctx context.Context,
	userID *uuid.UUID,
	ipAddress string,
	resource string,
	action string,
	reason string,
) error {
	details := map[string]interface{}{
		"resource":   resource,
		"action":     action,
		"reason":     reason,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	return e.logEvent(ctx, EventAccessDenied, userID, ipAddress, details)
}

// LogPrivilegeEscalationAttempt records a privilege escalation attempt
func (e *EnhancedAuditLogger) LogPrivilegeEscalationAttempt(
	ctx context.Context,
	userID *uuid.UUID,
	ipAddress string,
	attemptedRole string,
	currentRole string,
) error {
	details := map[string]interface{}{
		"attempted_role": attemptedRole,
		"current_role":   currentRole,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"severity":       "high",
	}
	return e.logEvent(ctx, EventPrivilegeEscalation, userID, ipAddress, details)
}

// LogSuspiciousActivity records suspicious activity
func (e *EnhancedAuditLogger) LogSuspiciousActivity(
	ctx context.Context,
	userID *uuid.UUID,
	ipAddress string,
	activityType string,
	details map[string]interface{},
) error {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["activity_type"] = activityType
	details["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	details["severity"] = "medium"
	return e.logEvent(ctx, EventSuspiciousActivity, userID, ipAddress, details)
}

// LogTokenReuseDetected records a token reuse attack
func (e *EnhancedAuditLogger) LogTokenReuseDetected(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress string,
	tokenHash string,
	familyRevoked bool,
) error {
	details := map[string]interface{}{
		"token_hash":      tokenHash,
		"family_revoked":  familyRevoked,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"severity":        "critical",
	}
	return e.logEvent(ctx, EventTokenReuse, &userID, ipAddress, details)
}

// LogConcurrentSession records concurrent session detection
func (e *EnhancedAuditLogger) LogConcurrentSession(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress string,
	sessionCount int,
	maxAllowed int,
) error {
	details := map[string]interface{}{
		"session_count": sessionCount,
		"max_allowed":   maxAllowed,
		"action_taken":  "denied",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}
	return e.logEvent(ctx, EventConcurrentSession, &userID, ipAddress, details)
}

// LogSecurityAlert records a high-priority security alert
func (e *EnhancedAuditLogger) LogSecurityAlert(
	ctx context.Context,
	userID *uuid.UUID,
	ipAddress string,
	alertType string,
	severity string,
	details map[string]interface{},
) error {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["alert_type"] = alertType
	details["severity"] = severity
	details["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	details["requires_review"] = severity == "critical" || severity == "high"
	return e.logEvent(ctx, EventSecurityAlert, userID, ipAddress, details)
}

// LogAdminAction records an administrative action
func (e *EnhancedAuditLogger) LogAdminAction(
	ctx context.Context,
	adminUserID uuid.UUID,
	ipAddress string,
	actionType string,
	targetUser *uuid.UUID,
	details map[string]interface{},
) error {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["admin_user_id"] = adminUserID.String()
	details["action"] = actionType
	details["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	if targetUser != nil {
		details["target_user_id"] = targetUser.String()
	}
	return e.logEvent(ctx, actionType, &adminUserID, ipAddress, details)
}

// LogBeaconEvent records beacon-related security events
func (e *EnhancedAuditLogger) LogBeaconEvent(
	ctx context.Context,
	beaconID uuid.UUID,
	eventType string,
	ipAddress string,
	details map[string]interface{},
) error {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["beacon_id"] = beaconID.String()
	details["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	return e.logEvent(ctx, eventType, nil, ipAddress, details)
}

// logEvent is the internal method that writes to the audit log
func (e *EnhancedAuditLogger) logEvent(
	ctx context.Context,
	eventType string,
	userID *uuid.UUID,
	ipAddress string,
	details map[string]interface{},
) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	var userIDPtr *string
	if userID != nil {
		uid := userID.String()
		userIDPtr = &uid
	}

	var ipAddressPtr *string
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	_, err = e.pool.Exec(ctx, `
		INSERT INTO auth_audit_logs (event_type, user_id, ip_address, details, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, eventType, userIDPtr, ipAddressPtr, detailsJSON)

	return err
}

// QueryAuditLogs retrieves audit logs with filtering
func (e *EnhancedAuditLogger) QueryAuditLogs(
	ctx context.Context,
	filter AuditLogFilter,
) ([]map[string]interface{}, error) {
	query := `
		SELECT id, event_type, user_id, ip_address, details, created_at
		FROM auth_audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filter.EventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, filter.EventType)
		argCount++
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filter.StartTime)
		argCount++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filter.EndTime)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var eventType string
		var userID *string
		var ipAddress *string
		var detailsJSON []byte
		var createdAt time.Time

		err = rows.Scan(&id, &eventType, &userID, &ipAddress, &detailsJSON, &createdAt)
		if err != nil {
			return nil, err
		}

		var details map[string]interface{}
		if detailsJSON != nil {
			err = json.Unmarshal(detailsJSON, &details)
			if err != nil {
				return nil, err
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

	return results, nil
}

// AuditLogFilter provides filtering options for audit log queries
type AuditLogFilter struct {
	EventType string
	UserID    *string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
}

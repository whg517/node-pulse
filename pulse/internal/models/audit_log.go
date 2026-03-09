package models

import "github.com/jackc/pgx/v5/pgtype"

// AuditLog represents an authentication/security event in the audit log
type AuditLog struct {
	ID               int64            `json:"id" db:"id"`
	EventType        string           `json:"event_type" db:"event_type"` // login, logout, token_refresh, etc.
	UserID           *string          `json:"user_id,omitempty" db:"user_id"`
	ServiceAccountID *string          `json:"service_account_id,omitempty" db:"service_account_id"`
	SessionID        *string          `json:"session_id,omitempty" db:"session_id"`
	IPAddress        *string          `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent        *string          `json:"user_agent,omitempty" db:"user_agent"`
	Details          map[string]any   `json:"details,omitempty" db:"details"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
}

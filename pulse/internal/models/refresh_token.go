package models

import (
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Session represents a user session
type Session struct {
	SessionID      uuid.UUID        `json:"session_id" db:"session_id"`
	UserID         uuid.UUID        `json:"user_id" db:"user_id"`
	DeviceID       *string          `json:"device_id,omitempty" db:"device_id"`
	IPAddress      *netip.Addr      `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string          `json:"user_agent,omitempty" db:"user_agent"`
	RememberMe     bool             `json:"remember_me" db:"remember_me"`
	ExpiresAt      pgtype.Timestamp `json:"expires_at" db:"expires_at"`
	MaxValidUntil  pgtype.Timestamp `json:"max_valid_until" db:"max_valid_until"`
	LastActivityAt pgtype.Timestamp `json:"last_activity_at" db:"last_activity_at"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
}

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID            int               `json:"id" db:"id"`
	TokenID       uuid.UUID         `json:"token_id" db:"token_id"`
	TokenHash     string            `json:"-" db:"token_hash"` // Never expose in JSON
	UserID        uuid.UUID         `json:"user_id" db:"user_id"`
	SessionID     *uuid.UUID        `json:"session_id,omitempty" db:"session_id"`
	ExpiresAt     pgtype.Timestamp  `json:"expires_at" db:"expires_at"`
	MaxValidUntil pgtype.Timestamp  `json:"max_valid_until" db:"max_valid_until"`
	RevokedAt     *pgtype.Timestamp `json:"revoked_at,omitempty" db:"revoked_at"`
	ReplacedBy    *uuid.UUID        `json:"replaced_by,omitempty" db:"replaced_by"`
	UserAgent     *string           `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress     *netip.Addr       `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt     pgtype.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt     pgtype.Timestamp  `json:"updated_at" db:"updated_at"`
}

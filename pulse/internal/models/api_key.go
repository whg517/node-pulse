package models

import "github.com/jackc/pgx/v5/pgtype"

// APIKey represents an API key for device/beacon authentication
type APIKey struct {
	ID               int               `json:"id" db:"id"`
	KeyID            string            `json:"key_id" db:"key_id"`         // Lookup identifier
	KeyHash          string            `json:"-" db:"key_hash"`            // SHA-256 hash, never expose
	KeyPrefix        string            `json:"key_prefix" db:"key_prefix"` // First 8 chars for identification
	UserID           *string           `json:"user_id,omitempty" db:"user_id"`
	ServiceAccountID *string           `json:"service_account_id,omitempty" db:"service_account_id"`
	Name             string            `json:"name" db:"name"`
	IsActive         bool              `json:"is_active" db:"is_active"`
	ExpiresAt        *pgtype.Timestamp `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt        pgtype.Timestamp  `json:"created_at" db:"created_at"`
	LastUsedAt       *pgtype.Timestamp `json:"last_used_at,omitempty" db:"last_used_at"`
}

// ServiceAccount represents a service account for automated API access
type ServiceAccount struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description *string           `json:"description,omitempty" db:"description"`
	Scopes      []string          `json:"scopes" db:"scopes"`
	IsActive    bool              `json:"is_active" db:"is_active"`
	CreatedAt   pgtype.Timestamp  `json:"created_at" db:"created_at"`
	ExpiresAt   *pgtype.Timestamp `json:"expires_at,omitempty" db:"expires_at"`
}

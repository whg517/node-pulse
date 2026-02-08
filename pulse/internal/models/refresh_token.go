package models

import "github.com/jackc/pgx/v5/pgtype"

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID           int                `json:"id" db:"id"`
	TokenID      string             `json:"token_id" db:"token_id"`
	UserID       string             `json:"user_id" db:"user_id"`
	ExpiresAt    pgtype.Timestamp   `json:"expires_at" db:"expires_at"`
	MaxValidUntil pgtype.Timestamp  `json:"max_valid_until" db:"max_valid_until"`
	RevokedAt    *pgtype.Timestamp  `json:"revoked_at,omitempty" db:"revoked_at"`
	ReplacedBy   *string            `json:"replaced_by,omitempty" db:"replaced_by"`
	UserAgent    *string            `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress    *string            `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt    pgtype.Timestamp   `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp   `json:"updated_at" db:"updated_at"`
}

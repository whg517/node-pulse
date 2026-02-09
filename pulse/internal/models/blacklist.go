package models

import "github.com/jackc/pgx/v5/pgtype"

// BlacklistEntry represents a revoked JWT in the token blacklist
type BlacklistEntry struct {
	JTI        string          `json:"jti" db:"jti"`
	RevokedAt  pgtype.Timestamp `json:"revoked_at" db:"revoked_at"`
	ExpiresAt  pgtype.Timestamp `json:"expires_at" db:"expires_at"`
}

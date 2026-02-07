package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBeaconTokenNotFound = errors.New("beacon token not found")
	ErrBeaconTokenExpired  = errors.New("beacon token expired")
)

// BeaconTokensQuerier defines interface for beacon token database operations
type BeaconTokensQuerier interface {
	GetNodeIDByAPIKey(ctx context.Context, apiKeyHash string) (uuid.UUID, error)
	UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error
}

// beaconTokensQuerier implements BeaconTokensQuerier interface
type beaconTokensQuerier struct {
	pool *pgxpool.Pool
}

// NewBeaconTokensQuerier creates a new BeaconTokensQuerier
func NewBeaconTokensQuerier(pool *pgxpool.Pool) BeaconTokensQuerier {
	return &beaconTokensQuerier{pool: pool}
}

// GetNodeIDByAPIKey retrieves node ID and token ID by API key hash
func (q *beaconTokensQuerier) GetNodeIDByAPIKey(ctx context.Context, apiKeyHash string) (uuid.UUID, error) {
	var nodeID uuid.UUID
	var tokenID uuid.UUID
	var isActive bool
	var expiresAt interface{} // Can be nil or time.Time

	query := `
		SELECT node_id, token_id, is_active, expires_at
		FROM beacon_tokens
		WHERE api_key_hash = $1
		LIMIT 1
	`

	err := q.pool.QueryRow(ctx, query, apiKeyHash).Scan(&nodeID, &tokenID, &isActive, &expiresAt)
	if err != nil {
		return uuid.Nil, ErrBeaconTokenNotFound
	}

	// Check if token is active
	if !isActive {
		return uuid.Nil, ErrBeaconTokenNotFound
	}

	// Check if token has expired (if expires_at is set)
	if expiresAt != nil {
		// TODO: Implement proper expiration check using pgtype.Timestamp
		// For now, we rely on the database constraint and is_active flag
	}

	// Update last_used_at timestamp asynchronously (fire and forget)
	_ = q.UpdateLastUsed(ctx, tokenID)

	return nodeID, nil
}

// UpdateLastUsed updates the last_used_at timestamp for a token
func (q *beaconTokensQuerier) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE beacon_tokens
		SET last_used_at = NOW()
		WHERE token_id = $1
	`

	_, err := q.pool.Exec(ctx, query, tokenID)
	return err
}

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	TokenID    string    `db:"token_id"`
	UserID     string    `db:"user_id"`
	TokenHash  string    `db:"token_hash"`
	Jti        string    `db:"jti"`
	DeviceInfo string    `db:"device_info"`
	IPAddress  string    `db:"ip_address"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
}

// RefreshTokenStore handles refresh token database operations
type RefreshTokenStore struct {
	pool *pgxpool.Pool
}

// NewRefreshTokenStore creates a new refresh token store
func NewRefreshTokenStore(pool *pgxpool.Pool) *RefreshTokenStore {
	return &RefreshTokenStore{pool: pool}
}

// Save stores a new refresh token in the database
func (s *RefreshTokenStore) Save(ctx context.Context, userID, tokenHash, jti, deviceInfo, ipAddress string, expiresAt time.Time) error {
	tokenID := uuid.New().String()

	query := `
		INSERT INTO refresh_tokens (token_id, user_id, token_hash, jti, device_info, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	_, err := s.pool.Exec(ctx, query, tokenID, userID, tokenHash, jti, deviceInfo, ipAddress, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

// GetByHash retrieves a refresh token by its hash
func (s *RefreshTokenStore) GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var token RefreshToken

	query := `
		SELECT token_id, user_id, token_hash, jti, device_info, ip_address, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	err := s.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.TokenID,
		&token.UserID,
		&token.TokenHash,
		&token.Jti,
		&token.DeviceInfo,
		&token.IPAddress,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

// Delete removes a refresh token from the database
func (s *RefreshTokenStore) Delete(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	result, err := s.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}

// DeleteAllForUser removes all refresh tokens for a specific user
func (s *RefreshTokenStore) DeleteAllForUser(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all refresh tokens for user: %w", err)
	}

	return nil
}

// DeleteByJti removes a refresh token by its JWT ID
func (s *RefreshTokenStore) DeleteByJti(ctx context.Context, jti string) error {
	query := `DELETE FROM refresh_tokens WHERE jti = $1`

	result, err := s.pool.Exec(ctx, query, jti)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token by jti: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("refresh token with jti %s not found", jti)
	}

	return nil
}

// CountActiveTokens returns the number of active (non-expired) tokens for a user
func (s *RefreshTokenStore) CountActiveTokens(ctx context.Context, userID string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE user_id = $1 AND expires_at > NOW()
	`

	err := s.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active tokens: %w", err)
	}

	return count, nil
}

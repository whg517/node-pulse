package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// RefreshTokenService manages refresh tokens with concurrency protection
type RefreshTokenService struct {
	pool    *pgxpool.Pool
	mutexes map[string]*sync.Mutex
	mu      sync.Mutex
}

// NewRefreshTokenService creates a new refresh token service
func NewRefreshTokenService(pool *pgxpool.Pool) *RefreshTokenService {
	return &RefreshTokenService{
		pool:    pool,
		mutexes: make(map[string]*sync.Mutex),
	}
}

// getMutex returns a mutex for the given user ID (with cleanup)
func (s *RefreshTokenService) getMutex(userID string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.mutexes[userID]; !exists {
		s.mutexes[userID] = &sync.Mutex{}
	}

	return s.mutexes[userID]
}

// CreateRefreshToken creates a new refresh token for a user
func (s *RefreshTokenService) CreateRefreshToken(ctx context.Context, userID string, userAgent, ipAddress string, maxValidityDays int) (string, *models.RefreshToken, error) {
	tokenPlain := uuid.New().String()

	// Hash the token for storage
	hash := sha256.Sum256([]byte(tokenPlain))
	tokenHash := hex.EncodeToString(hash[:])

	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days
	maxValidUntil := now.Add(time.Duration(maxValidityDays) * 24 * time.Hour)

	// Handle empty IP address (set to NULL)
	var ipAddressPtr *string
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	// Handle empty user agent (set to NULL)
	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	// Insert into database
	var dbToken models.RefreshToken
	err := s.pool.QueryRow(ctx, `
		INSERT INTO refresh_tokens (token_id, user_id, expires_at, max_valid_until, user_agent, ip_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, token_id, user_id, expires_at, max_valid_until, revoked_at, replaced_by, user_agent, ip_address, created_at, updated_at
	`, tokenHash, userID, expiresAt, maxValidUntil, userAgentPtr, ipAddressPtr).Scan(
		&dbToken.ID,
		&dbToken.TokenID,
		&dbToken.UserID,
		&dbToken.ExpiresAt,
		&dbToken.MaxValidUntil,
		&dbToken.RevokedAt,
		&dbToken.ReplacedBy,
		&dbToken.UserAgent,
		&dbToken.IPAddress,
		&dbToken.CreatedAt,
		&dbToken.UpdatedAt,
	)

	if err != nil {
		return "", nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return tokenPlain, &dbToken, nil
}

// ValidateRefreshToken validates a refresh token and returns the database record
func (s *RefreshTokenService) ValidateRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	// Hash the token to compare with database
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Query the token (using a transaction with SELECT FOR UPDATE for concurrency protection)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var dbToken models.RefreshToken
	err = tx.QueryRow(ctx, `
		SELECT id, token_id, user_id, expires_at, max_valid_until, revoked_at, replaced_by, user_agent, ip_address, created_at, updated_at
		FROM refresh_tokens
		WHERE token_id = $1
		FOR UPDATE
	`, tokenHash).Scan(
		&dbToken.ID,
		&dbToken.TokenID,
		&dbToken.UserID,
		&dbToken.ExpiresAt,
		&dbToken.MaxValidUntil,
		&dbToken.RevokedAt,
		&dbToken.ReplacedBy,
		&dbToken.UserAgent,
		&dbToken.IPAddress,
		&dbToken.CreatedAt,
		&dbToken.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if token is revoked
	if dbToken.RevokedAt != nil {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Check expiration
	now := time.Now()
	if dbToken.ExpiresAt.Time.Before(now) {
		return nil, fmt.Errorf("token has expired")
	}

	// Check max valid until
	if dbToken.MaxValidUntil.Time.Before(now) {
		return nil, fmt.Errorf("token has exceeded maximum validity period")
	}

	return &dbToken, nil
}

// RotateRefreshToken rotates a refresh token (one-time use, returns new token)
func (s *RefreshTokenService) RotateRefreshToken(ctx context.Context, oldToken string, userAgent, ipAddress string, maxValidityDays int) (string, *models.RefreshToken, error) {
	// Hash the old token
	hash := sha256.Sum256([]byte(oldToken))
	tokenHash := hex.EncodeToString(hash[:])

	// First, validate and get the old token info to get userID
	oldTokenInfo, err := s.ValidateRefreshToken(ctx, oldToken)
	if err != nil {
		return "", nil, err
	}

	// Get mutex for this user
	userMutex := s.getMutex(oldTokenInfo.UserID)
	userMutex.Lock()
	defer userMutex.Unlock()

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Double-check token hasn't been used (concurrent request)
	var revokedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT revoked_at FROM refresh_tokens WHERE token_id = $1 FOR UPDATE
	`, tokenHash).Scan(&revokedAt)

	if err != nil {
		return "", nil, fmt.Errorf("token not found: %w", err)
	}

	if revokedAt != nil {
		return "", nil, fmt.Errorf("token already used") // Will return 409 Conflict
	}

	// Mark old token as revoked
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $1, updated_at = $1
		WHERE token_id = $2
	`, now, tokenHash)

	if err != nil {
		return "", nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Calculate new expiration (sliding window)
	newExpiresAt := now.Add(7 * 24 * time.Hour)
	if newExpiresAt.After(oldTokenInfo.MaxValidUntil.Time) {
		newExpiresAt = oldTokenInfo.MaxValidUntil.Time
	}

	// Create new token
	newTokenPlain := uuid.New().String()
	newHash := sha256.Sum256([]byte(newTokenPlain))
	newTokenHash := hex.EncodeToString(newHash[:])

	var newToken models.RefreshToken
	err = tx.QueryRow(ctx, `
		INSERT INTO refresh_tokens (token_id, user_id, expires_at, max_valid_until, user_agent, ip_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, token_id, user_id, expires_at, max_valid_until, revoked_at, replaced_by, user_agent, ip_address, created_at, updated_at
	`, newTokenHash, oldTokenInfo.UserID, newExpiresAt, oldTokenInfo.MaxValidUntil, userAgent, ipAddress).Scan(
		&newToken.ID,
		&newToken.TokenID,
		&newToken.UserID,
		&newToken.ExpiresAt,
		&newToken.MaxValidUntil,
		&newToken.RevokedAt,
		&newToken.ReplacedBy,
		&newToken.UserAgent,
		&newToken.IPAddress,
		&newToken.CreatedAt,
		&newToken.UpdatedAt,
	)

	if err != nil {
		return "", nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newTokenPlain, &newToken, nil
}

// RevokeRefreshToken revokes a refresh token by token ID
func (s *RefreshTokenService) RevokeRefreshToken(ctx context.Context, userID, tokenID string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, tokenID, userID)

	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("token not found or already revoked")
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *RefreshTokenService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}

	return nil
}

// GetUserRefreshTokens returns all non-revoked refresh tokens for a user
func (s *RefreshTokenService) GetUserRefreshTokens(ctx context.Context, userID string) ([]models.RefreshToken, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, token_id, user_id, expires_at, max_valid_until, revoked_at, replaced_by, user_agent, ip_address, created_at, updated_at
		FROM refresh_tokens
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to query refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.RefreshToken
	for rows.Next() {
		var token models.RefreshToken
		err := rows.Scan(
			&token.ID,
			&token.TokenID,
			&token.UserID,
			&token.ExpiresAt,
			&token.MaxValidUntil,
			&token.RevokedAt,
			&token.ReplacedBy,
			&token.UserAgent,
			&token.IPAddress,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// CleanupExpiredTokens removes expired and revoked tokens from the database
func (s *RefreshTokenService) CleanupExpiredTokens(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	_, err := s.pool.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE (revoked_at IS NOT NULL AND revoked_at < $1)
		OR (expires_at < $1)
	`, cutoff)

	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}

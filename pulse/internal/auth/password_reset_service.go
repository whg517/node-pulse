package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PasswordResetService handles password reset token generation and validation
type PasswordResetService struct {
	pool *pgxpool.Pool
}

// NewPasswordResetService creates a new password reset service
func NewPasswordResetService(pool *pgxpool.Pool) *PasswordResetService {
	return &PasswordResetService{
		pool: pool,
	}
}

// GenerateResetToken creates a 256-bit random token, stores SHA-256 hash
func (s *PasswordResetService) GenerateResetToken(ctx context.Context, userID string, ipAddress, userAgent string) (string, error) {
	// Generate 256-bit (32 byte) random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Encode as URL-safe base64
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash token with SHA-256 for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	// Token expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour)

	// Store token hash in database
	_, err := s.pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, tokenHash, expiresAt, ipAddress, userAgent)

	if err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	return token, nil
}

// ValidateResetToken checks token hash and expiry, returns userID
func (s *PasswordResetService) ValidateResetToken(ctx context.Context, token string) (*uuid.UUID, error) {
	// Hash token to compare with stored hash
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	// Query for valid token (not used, not expired)
	var userID uuid.UUID
	var usedAt *time.Time

	err := s.pool.QueryRow(ctx, `
		SELECT user_id, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1
		  AND expires_at > NOW()
		  AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, tokenHash).Scan(&userID, &usedAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	return &userID, nil
}

// ConsumeResetToken marks token as used (single-use)
func (s *PasswordResetService) ConsumeResetToken(ctx context.Context, token string) error {
	// Hash token to find and mark as used
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	// Mark token as used
	result, err := s.pool.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE token_hash = $1
		  AND used_at IS NULL
	`, tokenHash)

	if err != nil {
		return fmt.Errorf("failed to consume token: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("token not found or already used")
	}

	return nil
}

// CleanupExpiredTokens removes tokens older than 24 hours
func (s *PasswordResetService) CleanupExpiredTokens(ctx context.Context) error {
	// Delete tokens that are either:
	// 1. Expired and created more than 24 hours ago
	// 2. Already used and created more than 24 hours ago
	cutoff := time.Now().Add(-24 * time.Hour)

	_, err := s.pool.Exec(ctx, `
		DELETE FROM password_reset_tokens
		WHERE (expires_at < NOW() OR used_at IS NOT NULL)
		  AND created_at < $1
	`, cutoff)

	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}

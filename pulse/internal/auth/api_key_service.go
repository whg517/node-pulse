package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// APIKeyService manages API keys for device/beacon authentication
type APIKeyService struct {
	pool *pgxpool.Pool
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(pool *pgxpool.Pool) *APIKeyService {
	return &APIKeyService{pool: pool}
}

// GenerateAPIKey generates a new API key (256-bit random, base64 URL-encoded)
func (s *APIKeyService) GenerateAPIKey(ctx context.Context, userID *string, name string) (string, *models.APIKey, error) {
	// Generate 256-bit random token (32 bytes)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64 URL-safe string
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token for storage
	hash := sha256.Sum256(tokenBytes)
	tokenHash := hex.EncodeToString(hash[:])

	// Extract key prefix (first 8 chars of base64) for identification
	keyPrefix := token[:8]

	// Insert into database
	var dbKey models.APIKey
	err := s.pool.QueryRow(ctx, `
		INSERT INTO api_keys (key_hash, key_prefix, user_id, name, is_active, created_at)
		VALUES ($1, $2, $3, $4, true, NOW())
		RETURNING id, key_hash, key_prefix, user_id, name, is_active, expires_at, created_at, last_used_at
	`, tokenHash, keyPrefix, userID, name).Scan(
		&dbKey.ID,
		&dbKey.KeyHash,
		&dbKey.KeyPrefix,
		&dbKey.UserID,
		&dbKey.Name,
		&dbKey.IsActive,
		&dbKey.ExpiresAt,
		&dbKey.CreatedAt,
		&dbKey.LastUsedAt,
	)

	if err != nil {
		return "", nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return token, &dbKey, nil
}

// ValidateAPIKey validates an API key and returns the database record
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, token string) (*models.APIKey, error) {
	// Strip np_live_ prefix if present (tokens may be provided with or without prefix)
	const keyPrefix = "np_live_"
	if strings.HasPrefix(token, keyPrefix) {
		token = strings.TrimPrefix(token, keyPrefix)
	}

	// Decode the token to get raw bytes
	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid API key format: %w", err)
	}

	// Hash the token to compare with database
	hash := sha256.Sum256(tokenBytes)
	tokenHash := hex.EncodeToString(hash[:])

	var dbKey models.APIKey
	err = s.pool.QueryRow(ctx, `
		SELECT id, key_hash, key_prefix, user_id, name, is_active, expires_at, created_at, last_used_at
		FROM api_keys
		WHERE key_hash = $1
	`, tokenHash).Scan(
		&dbKey.ID,
		&dbKey.KeyHash,
		&dbKey.KeyPrefix,
		&dbKey.UserID,
		&dbKey.Name,
		&dbKey.IsActive,
		&dbKey.ExpiresAt,
		&dbKey.CreatedAt,
		&dbKey.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Check if key is active
	if !dbKey.IsActive {
		return nil, fmt.Errorf("API key is inactive")
	}

	// Check expiration
	if dbKey.ExpiresAt != nil && dbKey.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Update last_used_at
	_, err = s.pool.Exec(ctx, `
		UPDATE api_keys
		SET last_used_at = NOW()
		WHERE id = $1
	`, dbKey.ID)

	if err != nil {
		// Don't fail validation if we can't update last_used_at
		// Just log the error (in production, use proper logging)
		fmt.Printf("WARN: failed to update last_used_at for API key: %v\n", err)
	}

	return &dbKey, nil
}

// RevokeAPIKey revokes an API key by ID
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, userID *string, keyID int) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE api_keys
		SET is_active = false
		WHERE id = $1 AND ($2::uuid IS NULL OR user_id = $2)
	`, keyID, userID)

	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// GetUserAPIKeys returns all API keys for a user
func (s *APIKeyService) GetUserAPIKeys(ctx context.Context, userID string) ([]models.APIKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, key_hash, key_prefix, user_id, name, is_active, expires_at, created_at, last_used_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var key models.APIKey
		err := rows.Scan(
			&key.ID,
			&key.KeyHash,
			&key.KeyPrefix,
			&key.UserID,
			&key.Name,
			&key.IsActive,
			&key.ExpiresAt,
			&key.CreatedAt,
			&key.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// CleanupExpiredKeys removes expired keys from the database
func (s *APIKeyService) CleanupExpiredKeys(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	_, err := s.pool.Exec(ctx, `
		DELETE FROM api_keys
		WHERE is_active = false AND created_at < $1
	`, cutoff)

	if err != nil {
		return fmt.Errorf("failed to cleanup expired keys: %w", err)
	}

	return nil
}

// ListAllAPIKeys returns all API keys in the system (admin only)
func (s *APIKeyService) ListAllAPIKeys(ctx context.Context) ([]models.APIKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, key_id, key_hash, key_prefix, user_id, service_account_id, name, is_active, expires_at, created_at, last_used_at
		FROM api_keys
		ORDER BY created_at DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query all API keys: %w", err)
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var key models.APIKey
		err := rows.Scan(
			&key.ID,
			&key.KeyID,
			&key.KeyHash,
			&key.KeyPrefix,
			&key.UserID,
			&key.ServiceAccountID,
			&key.Name,
			&key.IsActive,
			&key.ExpiresAt,
			&key.CreatedAt,
			&key.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// GetAPIKeyByID retrieves an API key by its database ID
func (s *APIKeyService) GetAPIKeyByID(ctx context.Context, keyID int) (*models.APIKey, error) {
	var key models.APIKey
	err := s.pool.QueryRow(ctx, `
		SELECT id, key_id, key_hash, key_prefix, user_id, service_account_id, name, is_active, expires_at, created_at, last_used_at
		FROM api_keys
		WHERE id = $1
	`, keyID).Scan(
		&key.ID,
		&key.KeyID,
		&key.KeyHash,
		&key.KeyPrefix,
		&key.UserID,
		&key.ServiceAccountID,
		&key.Name,
		&key.IsActive,
		&key.ExpiresAt,
		&key.CreatedAt,
		&key.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	return &key, nil
}

// RotateAPIKey creates a new key while keeping the old key valid for 24 hours
func (s *APIKeyService) RotateAPIKey(ctx context.Context, oldKeyID int) (string, *models.APIKey, *models.APIKey, error) {
	// Get the old key
	oldKey, err := s.GetAPIKeyByID(ctx, oldKeyID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get old key: %w", err)
	}

	// Set old key to expire in 24 hours
	expiryTime := time.Now().Add(24 * time.Hour)
	_, err = s.pool.Exec(ctx, `
		UPDATE api_keys
		SET expires_at = $1
		WHERE id = $2
	`, expiryTime, oldKeyID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to update old key expiry: %w", err)
	}

	// Generate new key with np_live_ prefix
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64 URL-safe string
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Add np_live_ prefix
	fullToken := "np_live_" + token

	// Hash the token for storage (hash only the base64 part, not the prefix)
	hash := sha256.Sum256(tokenBytes)
	tokenHash := hex.EncodeToString(hash[:])

	// Generate key_id (20 char unique identifier)
	keyIDBytes := make([]byte, 10)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return "", nil, nil, fmt.Errorf("failed to generate key_id: %w", err)
	}
	keyIDStr := hex.EncodeToString(keyIDBytes)

	// Extract key prefix (first 8 chars of base64) for identification
	keyPrefix := token[:8]

	// Insert new key into database
	var newKey models.APIKey
	err = s.pool.QueryRow(ctx, `
		INSERT INTO api_keys (key_id, key_hash, key_prefix, user_id, service_account_id, name, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW())
		RETURNING id, key_id, key_hash, key_prefix, user_id, service_account_id, name, is_active, expires_at, created_at, last_used_at
	`, keyIDStr, tokenHash, keyPrefix, oldKey.UserID, oldKey.ServiceAccountID, oldKey.Name).Scan(
		&newKey.ID,
		&newKey.KeyID,
		&newKey.KeyHash,
		&newKey.KeyPrefix,
		&newKey.UserID,
		&newKey.ServiceAccountID,
		&newKey.Name,
		&newKey.IsActive,
		&newKey.ExpiresAt,
		&newKey.CreatedAt,
		&newKey.LastUsedAt,
	)

	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create new API key: %w", err)
	}

	// Refresh old key with updated expiry
	oldKey.ExpiresAt.Time = expiryTime
	oldKey.ExpiresAt.Valid = true

	return fullToken, oldKey, &newKey, nil
}

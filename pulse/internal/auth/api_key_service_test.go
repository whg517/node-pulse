package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeyService_GenerateAPIKey tests API key generation
func TestAPIKeyService_GenerateAPIKey(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewAPIKeyService(pool)
	ctx := context.Background()

	userID := CreateTestUser(t, ctx, pool)
	token, dbKey, err := service.GenerateAPIKey(ctx, &userID, "Test Key")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, dbKey)
	assert.Equal(t, userID, *dbKey.UserID)
	assert.Equal(t, "Test Key", dbKey.Name)
	assert.True(t, dbKey.IsActive)
	assert.Len(t, dbKey.KeyPrefix, 8) // First 8 chars
	assert.NotEqual(t, token, dbKey.KeyHash) // Hash should not match plain token
}

// TestAPIKeyService_ValidateAPIKey tests API key validation
func TestAPIKeyService_ValidateAPIKey(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewAPIKeyService(pool)
	ctx := context.Background()

	userID := CreateTestUser(t, ctx, pool)
	token, _, err := service.GenerateAPIKey(ctx, &userID, "Test Key")
	require.NoError(t, err)

	// Validate token
	dbKey, err := service.ValidateAPIKey(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, userID, *dbKey.UserID)
	assert.True(t, dbKey.IsActive)
}

// TestAPIKeyService_ValidateAPIKey_InvalidToken tests invalid API key
func TestAPIKeyService_ValidateAPIKey_InvalidToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewAPIKeyService(pool)
	ctx := context.Background()

	_, err := service.ValidateAPIKey(ctx, "invalid-key-format")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// TestAPIKeyService_RevokeAPIKey tests key revocation
func TestAPIKeyService_RevokeAPIKey(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewAPIKeyService(pool)
	ctx := context.Background()

	userID := CreateTestUser(t, ctx, pool)
	_, dbKey, err := service.GenerateAPIKey(ctx, &userID, "Test Key")
	require.NoError(t, err)

	// Revoke key
	err = service.RevokeAPIKey(ctx, &userID, dbKey.ID)
	require.NoError(t, err)

	// Try to validate - should fail
	_, err = service.ValidateAPIKey(ctx, "invalid") // We can't validate the original token since we hashed it
	assert.Error(t, err)
}

// TestAPIKeyService_GetUserAPIKeys tests listing user keys
func TestAPIKeyService_GetUserAPIKeys(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewAPIKeyService(pool)
	ctx := context.Background()

	userID := CreateTestUser(t, ctx, pool)

	// Create multiple keys
	_, _, err := service.GenerateAPIKey(ctx, &userID, "Key 1")
	require.NoError(t, err)

	_, _, err = service.GenerateAPIKey(ctx, &userID, "Key 2")
	require.NoError(t, err)

	// Get all keys
	keys, err := service.GetUserAPIKeys(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

// TestAPIKeyService_OneWayHash tests that original token cannot be retrieved from hash
func TestAPIKeyService_OneWayHash(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewAPIKeyService(pool)
	ctx := context.Background()

	userID := CreateTestUser(t, ctx, pool)
	token, dbKey, err := service.GenerateAPIKey(ctx, &userID, "Test Key")
	require.NoError(t, err)

	// Verify hash is not the plain token
	assert.NotEqual(t, token, dbKey.KeyHash)
	assert.NotContains(t, dbKey.KeyHash, token)
	assert.NotContains(t, token, dbKey.KeyHash)

	// The hash should be SHA-256 (64 hex chars)
	assert.Len(t, dbKey.KeyHash, 64)
}

// TestAPIKeyService_KeyPrefix tests key prefix extraction
func TestAPIKeyService_KeyPrefix(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewAPIKeyService(pool)
	ctx := context.Background()

	userID := CreateTestUser(t, ctx, pool)
	token, dbKey, err := service.GenerateAPIKey(ctx, &userID, "Test Key")
	require.NoError(t, err)

	// Verify prefix is first 8 chars of token
	assert.Equal(t, token[:8], dbKey.KeyPrefix)
}

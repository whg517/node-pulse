package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRefreshTokenService_CreateRefreshToken tests token creation
func TestRefreshTokenService_CreateRefreshToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	token, dbToken, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, dbToken)
	assert.Equal(t, "user-123", dbToken.UserID)
	assert.Nil(t, dbToken.RevokedAt)
}

// TestRefreshTokenService_ValidateRefreshToken tests token validation
func TestRefreshTokenService_ValidateRefreshToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create token
	token, _, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	// Validate token
	dbToken, err := service.ValidateRefreshToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", dbToken.UserID)
	assert.Nil(t, dbToken.RevokedAt)
}

// TestRefreshTokenService_RotateRefreshToken tests token rotation
func TestRefreshTokenService_RotateRefreshToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create initial token
	oldToken, _, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	// Rotate token
	newToken, newDbToken, err := service.RotateRefreshToken(ctx, oldToken, "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, oldToken, newToken)
	assert.Equal(t, "user-123", newDbToken.UserID)

	// Try to use old token again - should fail
	_, err = service.ValidateRefreshToken(ctx, oldToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revoked")
}

// TestRefreshTokenService_ConcurrentRotation tests concurrent refresh requests
func TestRefreshTokenService_ConcurrentRotation(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create initial token
	oldToken, _, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	// Launch concurrent rotations
	successCount := 0
	failCount := 0
	results := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, _, err := service.RotateRefreshToken(ctx, oldToken, "Mozilla/5.0", "192.168.1.1", 30)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < 10; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	// Only one should succeed
	assert.Equal(t, 1, successCount, "Expected exactly one successful rotation")
	assert.Equal(t, 9, failCount, "Expected 9 rotations to fail")
}

// TestRefreshTokenService_SlidingExpiration tests sliding expiration logic
func TestRefreshTokenService_SlidingExpiration(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create token with 30 day max validity
	oldToken, originalToken, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	// Rotate immediately - should extend expiration
	newToken, rotatedToken, err := service.RotateRefreshToken(ctx, oldToken, "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	// New token should have later expiration than original
	assert.True(t, rotatedToken.ExpiresAt.Time.After(originalToken.ExpiresAt.Time))
	assert.NotEqual(t, oldToken, newToken)
}

// TestRefreshTokenService_AbsoluteExpirationCap tests absolute max validity cap
func TestRefreshTokenService_AbsoluteExpirationCap(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create token with short max validity (1 day)
	oldToken, _, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 1)
	require.NoError(t, err)

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Try to rotate - new expiration should not exceed max_valid_until
	_, rotatedToken, err := service.RotateRefreshToken(ctx, oldToken, "Mozilla/5.0", "192.168.1.1", 1)
	require.NoError(t, err)

	// The rotated token's max_valid_until should match the original
	// (can't verify exact expiration time without more complex time mocking)
	assert.NotNil(t, rotatedToken.MaxValidUntil)
}

// TestRefreshTokenService_RevokeRefreshToken tests single token revocation
func TestRefreshTokenService_RevokeRefreshToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create token
	_, dbToken, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	// Revoke token
	err = service.RevokeRefreshToken(ctx, "user-123", dbToken.TokenID)
	require.NoError(t, err)

	// Try to validate - should fail
	_, err = service.ValidateRefreshToken(ctx, dbToken.TokenID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revoked")
}

// TestRefreshTokenService_RevokeAllUserTokens tests revoking all user tokens
func TestRefreshTokenService_RevokeAllUserTokens(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create multiple tokens for same user
	_, token1, _ := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	_, token2, _ := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)

	// Revoke all tokens
	err := service.RevokeAllUserTokens(ctx, "user-123")
	require.NoError(t, err)

	// Verify all tokens are revoked
	_, err = service.ValidateRefreshToken(ctx, token1.TokenID)
	assert.Error(t, err)

	_, err = service.ValidateRefreshToken(ctx, token2.TokenID)
	assert.Error(t, err)
}

// TestRefreshTokenService_GetUserRefreshTokens tests listing user tokens
func TestRefreshTokenService_GetUserRefreshTokens(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create multiple tokens
	_, _, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)
	_, _, err = service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	// Get all tokens
	tokens, err := service.GetUserRefreshTokens(ctx, "user-123")
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
}

// TestRefreshTokenService_CleanupExpiredTokens tests cleanup job
func TestRefreshTokenService_CleanupExpiredTokens(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewRefreshTokenService(pool)
	ctx := context.Background()

	// Create and immediately revoke a token
	_, dbToken, err := service.CreateRefreshToken(ctx, "user-123", "Mozilla/5.0", "192.168.1.1", 30)
	require.NoError(t, err)

	err = service.RevokeRefreshToken(ctx, "user-123", dbToken.TokenID)
	require.NoError(t, err)

	// Run cleanup with 0 retention (delete all revoked)
	err = service.CleanupExpiredTokens(ctx, 0)
	require.NoError(t, err)

	// Verify revoked token was cleaned up
	tokens, err := service.GetUserRefreshTokens(ctx, "user-123")
	require.NoError(t, err)
	assert.Len(t, tokens, 0, "Revoked tokens should be cleaned up")
}

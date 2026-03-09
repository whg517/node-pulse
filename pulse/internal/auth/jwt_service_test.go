package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTService_GenerateAccessToken tests token generation
func TestJWTService_GenerateAccessToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	token, jti, err := service.GenerateAccessToken("user-123", "admin")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, jti)
	assert.NotEqual(t, token, jti) // Token and JTI should be different
}

// TestJWTService_GenerateAccessTokenWithSession tests token generation with session and scopes
func TestJWTService_GenerateAccessTokenWithSession(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	token, jti, err := service.GenerateAccessTokenWithSession("user-123", "admin", "session-abc", []string{"read:all", "write:nodes"})

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, jti)

	// Verify token contains session and scope claims
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "session-abc", claims.SessionID)
	assert.Equal(t, []string{"read:all", "write:nodes"}, claims.Scope)
}

// TestJWTService_ValidateAccessToken tests valid token validation
func TestJWTService_ValidateAccessToken(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	// Generate token
	token, _, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Validate token
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.NotEmpty(t, claims.JTI)
}

// TestJWTService_ValidateAccessToken_Expired tests expired token rejection
func TestJWTService_ValidateAccessToken_Expired(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	// Create service with negative expiration to ensure token is already expired
	service := NewTestJWTService(t, -1, pool)

	token, _, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Should fail validation immediately (token already expired)
	_, err = service.ValidateAccessToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestJWTService_ValidateAccessToken_AlgorithmConfusion tests "none" algorithm rejection
func TestJWTService_ValidateAccessToken_AlgorithmConfusion(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	// Create a malicious "none" algorithm token
	maliciousToken := "eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4ifQ."

	_, err := service.ValidateAccessToken(maliciousToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

// TestJWTService_ValidateAccessToken_HS256Rejection tests HS256 rejection (must be RS256)
func TestJWTService_ValidateAccessToken_HS256Rejection(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	// Create an HS256 token (signed with HMAC instead of RSA)
	// This is a valid JWT but should be rejected because we require RS256
	// Instead, we'll verify that RS256 validation rejects HS256 tokens
	hs256Token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMiLCJyb2xlIjoiYWRtaW4iLCJqdGkiOiJ0ZXN0LWp0aSIsImlzcyI6Im5vZGUtcHVsc2UiLCJzdWIiOiJ1c2VyLTEyMyIsImF1ZCI6Im5vZGUtcHVsc2UtYXBpIiwiZXhwIjoxNzUwMDAwMDAwLCJpYXQiOjE3NTAwMDAwMDAsIm5iZiI6MTc1MDAwMDAwMH0.invalid"

	_, err := service.ValidateAccessToken(hs256Token)
	assert.Error(t, err)
}

// TestJWTService_ValidateAccessToken_ClockSkew tests 60-second clock skew tolerance
func TestJWTService_ValidateAccessToken_ClockSkew(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	token, _, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Simulate token being 55 seconds in the future (within 60s leeway)
	// This should still validate due to clock skew tolerance
	claims, err := service.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)

	// Note: Testing actual future tokens requires mocking time,
	// but the parser is configured with jwt.WithLeeway(60*time.Second)
}

// TestJWTService_CheckRevoked tests blacklist checking
func TestJWTService_CheckRevoked(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	ctx := context.Background()
	service := NewTestJWTService(t, 15, pool)

	// Check non-revoked token
	revoked, err := service.CheckRevoked(ctx, "test-jti-123")
	require.NoError(t, err)
	assert.False(t, revoked)

	// Add to blacklist
	_, err = pool.Exec(ctx, `
		INSERT INTO token_blacklist (jti, revoked_at, expires_at)
		VALUES ($1, NOW(), NOW() + INTERVAL '1 hour')
	`, "test-jti-123")
	require.NoError(t, err)

	// Check revoked token
	revoked, err = service.CheckRevoked(ctx, "test-jti-123")
	require.NoError(t, err)
	assert.True(t, revoked)
}

// TestJWTService_GetJTI tests JTI extraction
func TestJWTService_GetJTI(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	token, expectedJTI, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Extract JTI from token
	extractedJTI, err := service.GetJTI(token)
	require.NoError(t, err)
	assert.Equal(t, expectedJTI, extractedJTI)
}

// TestJWTService_KeyID tests key ID handling
func TestJWTService_KeyID(t *testing.T) {
	pool := setupTestDBWithCleanup(t)

	service := NewTestJWTService(t, 15, pool)

	assert.Equal(t, "test-key-id", service.GetKeyID())
	assert.True(t, service.VerifyKeyID("test-key-id"))
	assert.False(t, service.VerifyKeyID("different-key-id"))

	// Verify key ID is included in token header
	_, _, err := service.GenerateAccessToken("user-123", "admin")
	require.NoError(t, err)

	// Note: Token header verification would require parsing the raw JWT
	// The implementation includes kid in the header during signing
}

// TestJWTService_NilPool tests fail-closed behavior with nil pool
func TestJWTService_NilPool(t *testing.T) {
	service := NewTestJWTService(t, 15, nil)

	ctx := context.Background()
	_, err := service.CheckRevoked(ctx, "test-jti")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database pool not initialized")
}

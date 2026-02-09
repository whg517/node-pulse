package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/db"
)

// TestDiagnosis_DirectServiceCall tests the service directly without HTTP layer
func TestDiagnosis_DirectServiceCall(t *testing.T) {
	pool, cleanup := db.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user first
	testUserID := uuid.New()
	passwordHash, _ := auth.HashPassword("testpass")
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, 'operator', NOW(), NOW())
	`, testUserID, "service_test_user", passwordHash)
	require.NoError(t, err, "Failed to create test user")

	// Create a refresh token directly
	service := auth.NewRefreshTokenService(pool)
	testUserIDStr := testUserID.String()

	plainToken, dbToken, err := service.CreateRefreshToken(
		ctx, testUserIDStr, "TestAgent", "", 30,  // Empty IP address
	)
	require.NoError(t, err, "CreateRefreshToken should succeed")

	t.Logf("=== DIRECT SERVICE TEST ===")
	t.Logf("Plain token: %s", plainToken)
	t.Logf("Token ID: %s", dbToken.TokenID)
	t.Logf("User ID: %s", dbToken.UserID)
	t.Logf("Expires: %v", dbToken.ExpiresAt)
	t.Logf("Revoked: %v", dbToken.RevokedAt)

	// Test 1: Validate the token
	t.Run("ValidateToken", func(t *testing.T) {
		validatedToken, err := service.ValidateRefreshToken(ctx, plainToken)
		if err != nil {
			t.Logf("✗ ValidateRefreshToken failed: %v", err)
		} else {
			t.Logf("✓ ValidateRefreshToken succeeded")
			t.Logf("  Token ID: %s", validatedToken.TokenID)
			t.Logf("  User ID: %s", validatedToken.UserID)
		}
	})

	// Test 2: Rotate the token
	t.Run("RotateToken", func(t *testing.T) {
		newToken, rotatedDBToken, err := service.RotateRefreshToken(
			ctx, plainToken, "TestAgent", "", 30,  // Empty IP address
		)
		if err != nil {
			t.Logf("✗ RotateRefreshToken failed: %v", err)
		} else {
			t.Logf("✓ RotateRefreshToken succeeded")
			t.Logf("  New token: %s", newToken)
			t.Logf("  Old token revoked: %v", rotatedDBToken.RevokedAt)
		}
	})

	// Test 3: Try to rotate again (should fail with "token already used")
	t.Run("RotateAgain_ShouldFail", func(t *testing.T) {
		_, _, err := service.RotateRefreshToken(
			ctx, plainToken, "TestAgent", "", 30,  // Empty IP address
		)
		if err != nil {
			if err.Error() == "token already used" {
				t.Logf("✓ Correctly returned 'token already used'")
			} else {
				t.Logf("✗ Wrong error: %v", err)
			}
		} else {
			t.Logf("✗ Should have failed but didn't!")
		}
	})
}

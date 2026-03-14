package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// TestDiagnosis_TokenHash verifies token hashing logic
func TestDiagnosis_TokenHash(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	if router == nil {
		return
	}
	defer pool.Close()

	auth.ClearRateLimitStore(context.Background(), pool)

	// Create test user and login
	testUserID := uuid.New()
	testPassword := "testPassword"
	passwordHash, _ := auth.HashPassword(testPassword)
	testUsername := fmt.Sprintf("hash_%s", testUserID.String()[:8])

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, testUserID, testUsername, passwordHash, "operator")
	require.NoError(t, err)
	defer cleanupTestUser(pool, testUsername)

	// Login
	loginReq := models.LoginRequest{
		Username: testUsername,
		Password: testPassword,
	}
	loginBody, _ := json.Marshal(loginReq)
	loginHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, loginHTTPReq)

	require.Equal(t, http.StatusOK, wLogin.Code)

	// Get refresh token from cookie
	var refreshToken string
	for _, c := range wLogin.Result().Cookies() {
		if c.Name == "refresh_token" {
			refreshToken = c.Value
			break
		}
	}
	require.NotEmpty(t, refreshToken, "Should have refresh_token cookie")

	t.Logf("=== TOKEN ANALYSIS ===")
	t.Logf("Refresh token from cookie: %s", refreshToken)
	t.Logf("Token length: %d", len(refreshToken))

	// Hash the token the same way the code does
	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])
	t.Logf("Token hash (SHA256): %s", tokenHash)

	// Query the database to see what's stored
	var dbTokenID uuid.UUID
	var dbTokenHash string
	var dbUserID string
	err = pool.QueryRow(context.Background(), `
		SELECT token_id, token_hash, user_id::text
		FROM refresh_tokens
		WHERE user_id = $1
	`, testUserID).Scan(&dbTokenID, &dbTokenHash, &dbUserID)

	if err != nil {
		t.Logf("✗ ERROR: Token not found in DB: %v", err)
	} else {
		t.Logf("✓ Token found in DB:")
		t.Logf("  - token_id: %s", dbTokenID)
		t.Logf("  - token_hash: %s", dbTokenHash)
		t.Logf("  - user_id: %s", dbUserID)

		// Compare hashes
		if tokenHash == dbTokenHash {
			t.Logf("✓ HASHES MATCH - Token should validate correctly")
		} else {
			t.Logf("✗ HASH MISMATCH - This is the bug!")
			t.Logf("  Expected: %s", dbTokenHash)
			t.Logf("  Got:      %s", tokenHash)
		}
	}

	// Now try to refresh
	refreshReq := models.RefreshRequest{
		RefreshToken: refreshToken,
	}
	reqBody, _ := json.Marshal(refreshReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("\n=== REFRESH REQUEST ===")
	t.Logf("Request body: %s", reqBody)
	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())
}

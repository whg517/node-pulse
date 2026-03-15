package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// TestDiagnosis_SequentialRefresh diagnoses concurrent refresh test failure
// by testing sequential requests first to identify if it's a code bug or test issue
func TestDiagnosis_SequentialRefresh(t *testing.T) {
	router, pool, _ := setupTestRouter(t)
	if router == nil {
		return
	}
	defer pool.Close()

	// Clear rate limit store for clean test
	_ = auth.ClearRateLimitStore(context.Background(), pool)

	// Arrange - Create test user and login
	testUserID := uuid.New()
	testPassword := "testPassword"
	passwordHash, _ := auth.HashPassword(testPassword)
	testUsername := fmt.Sprintf("diag_%s", testUserID.String()[:8])

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, testUserID, testUsername, passwordHash, "operator")
	require.NoError(t, err)
	defer cleanupTestUser(pool, testUsername)

	// Login to get initial tokens
	loginReq := models.LoginRequest{
		Username: testUsername,
		Password: testPassword,
	}
	loginBody, _ := json.Marshal(loginReq)
	loginHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, loginHTTPReq)

	require.Equal(t, http.StatusOK, wLogin.Code, "Login should succeed")

	// Get refresh token from cookie
	var refreshToken string
	for _, c := range wLogin.Result().Cookies() {
		if c.Name == "refresh_token" {
			refreshToken = c.Value
			break
		}
	}
	require.NotEmpty(t, refreshToken, "Should have refresh_token cookie")

	// Verify token exists in database before tests
	var tokenCountBefore int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", testUserID).Scan(&tokenCountBefore)
	require.NoError(t, err)
	require.Equal(t, 1, tokenCountBefore, "Should have 1 refresh token after login")
	t.Logf("✓ Refresh token created in DB (count: %d)", tokenCountBefore)

	// ============================================================================
	// TEST 1: First refresh request - should succeed
	// ============================================================================
	t.Run("Test1_FirstRefresh_ShouldSucceed", func(t *testing.T) {
		refreshReq := models.RefreshRequest{
			RefreshToken: refreshToken,
		}
		reqBody, _ := json.Marshal(refreshReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf("Test1 Response Status: %d", w.Code)
		t.Logf("Test1 Response Body: %s", w.Body.String())

		// Assert first refresh succeeds
		assert.Equal(t, http.StatusOK, w.Code, "First refresh should succeed (200 OK)")

		if w.Code == http.StatusOK {
			// The refresh endpoint returns a wrapped response: {"data": {"access_token": ...}}
			var resp struct {
				Message string `json:"message"`
				Data    struct {
					AccessToken string `json:"access_token"`
					TokenType   string `json:"token_type"`
					ExpiresIn   int    `json:"expires_in"`
				} `json:"data"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.NotEmpty(t, resp.Data.AccessToken, "Should have new access token")
			t.Logf("✓ First refresh succeeded, got new access token")
		}
	})

	// Small delay to ensure first request completes
	time.Sleep(100 * time.Millisecond)

	// Check database state after first refresh
	var tokenCountAfter1 int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", testUserID).Scan(&tokenCountAfter1)
	require.NoError(t, err)
	t.Logf("DB token count after first refresh: %d (should be 1 - old revoked, new added)", tokenCountAfter1)

	// ============================================================================
	// TEST 2: Second refresh with SAME token - should fail with 409
	// ============================================================================
	t.Run("Test2_SecondRefresh_SameToken_ShouldFail409", func(t *testing.T) {
		refreshReq := models.RefreshRequest{
			RefreshToken: refreshToken, // SAME token as test1
		}
		reqBody, _ := json.Marshal(refreshReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf("Test2 Response Status: %d", w.Code)
		t.Logf("Test2 Response Body: %s", w.Body.String())

		// DIAGNOSIS: This tells us if code is working correctly
		if w.Code == http.StatusConflict {
			t.Logf("✓ CODE IS CORRECT: Token was already used, returned 409 Conflict")
			t.Logf("  → Concurrent test failure is likely a TEST issue (httptest/goroutine timing)")
		} else if w.Code == http.StatusUnauthorized {
			t.Logf("✗ CODE BUG: Returned 401 instead of 409")
			t.Logf("  → Token validation logic has a bug")
			t.Logf("  → THIS MUST BE FIXED (security issue)")
		} else {
			t.Logf("✗ UNEXPECTED: Returned status %d", w.Code)
		}

		// Expected behavior: 409 Conflict
		assert.Equal(t, http.StatusConflict, w.Code, "Second refresh with same token should return 409 Conflict")
	})

	// Check final database state
	var tokenCountAfter2 int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", testUserID).Scan(&tokenCountAfter2)
	require.NoError(t, err)
	t.Logf("Final DB token count: %d", tokenCountAfter2)

	// Check revoked tokens
	var revokedCount int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NOT NULL", testUserID).Scan(&revokedCount)
	require.NoError(t, err)
	t.Logf("Revoked tokens in DB: %d", revokedCount)

	// ============================================================================
	// DIAGNOSIS SUMMARY
	// ============================================================================
	t.Log("\n=== DIAGNOSIS SUMMARY ===")
	if tokenCountBefore == 1 {
		t.Log("✓ Token created successfully")
	} else {
		t.Logf("✗ Token creation issue (count: %d)", tokenCountBefore)
	}

	t.Log("\nBased on test results:")
	t.Log("- If Test1=200 AND Test2=409 → Code correct, fix test")
	t.Log("- If Test1=200 AND Test2=401 → Code bug, must fix")
	t.Log("- If Test1=401          → Code bug in token validation")
}

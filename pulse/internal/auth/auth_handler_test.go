package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/config"
)

// setupAuthHandlerTest creates test database, config, and auth handler
func setupAuthHandlerTest(t *testing.T) (*pgxpool.Pool, *AuthHandler, *config.Config, func()) {
	t.Helper()

	// Use test database
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil, nil, nil, nil
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
		pool.Close()
		return nil, nil, nil, nil
	}

	// Clean up and create tables
	cleanupTables(ctx, pool)
	createTestTables(ctx, t, pool)

	// Create test config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:                       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			AccessTokenExpirationMinutes: 15,
			RefreshTokenExpirationDays:   7,
			RefreshTokenMaxValidityDays:  30,
		},
		RateLimit: config.RateLimitConfig{
			LoginMaxPerMinute:   5,
			LoginMaxPerDay:      100,
			RefreshMaxPerMinute: 10,
			RefreshMaxPerDay:    200,
			APIKeyMaxPerMinute:  11,
		},
	}

	// Create auth handler using correct constructor with RSA keys
	privateKeyPEM, publicKeyPEM := GenerateTestRSAKeyPair(t)
	handler := NewAuthHandler(
		pool,
		privateKeyPEM,
		publicKeyPEM,
		"test-key-id",
		cfg.JWT.AccessTokenExpirationMinutes,
		cfg.JWT.RefreshTokenExpirationDays,
		cfg.JWT.RefreshTokenMaxValidityDays,
		false, // cookieSecure false for tests
		RateLimitOptions{
			LoginPerMinute:   cfg.RateLimit.LoginMaxPerMinute,
			RefreshPerMinute: cfg.RateLimit.RefreshMaxPerMinute,
			LogoutPerMinute:  cfg.RateLimit.RefreshMaxPerMinute,
			APIKeyPerMinute:  cfg.RateLimit.APIKeyMaxPerMinute,
		},
	)

	cleanup := func() {
		cleanupTables(ctx, pool)
		pool.Close()
	}

	return pool, handler, cfg, cleanup
}

// TestAuthHandler_Login_Success tests successful login
func TestAuthHandler_Login_Success(t *testing.T) {
	_, handler, _, cleanup := setupAuthHandlerTest(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	// Create test user with hashed password
	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, err := HashPassword("TestPass123")
	require.NoError(t, err)

	_, err = handler.pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err, "Failed to create test user")

	// Create login request
	loginReq := map[string]string{
		"username": "testuser",
		"password": "TestPass123",
	}
	reqBody, _ := json.Marshal(loginReq)

	// Create request
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/auth/login", handler.Login)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code, "Login should succeed")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse response")

	// Check response structure
	assert.Contains(t, response["message"], "Login successful", "Should show success message")
	assert.NotNil(t, response["data"], "Response should have data field")

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "user_id", "Data should contain user_id")
	assert.Contains(t, data, "access_token", "Data should contain access_token")
	assert.Contains(t, data, "role", "Data should contain role")
}

// TestAuthHandler_Login_InvalidCredentials tests login with invalid credentials
func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	_, handler, _, cleanup := setupAuthHandlerTest(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	// Create test user
	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, _ := HashPassword("TestPass123")
	_, err := handler.pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err)

	// Create login request with wrong password
	loginReq := map[string]string{
		"username": "testuser",
		"password": "WrongPassword",
	}
	reqBody, _ := json.Marshal(loginReq)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/auth/login", handler.Login)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Login should fail with invalid credentials")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "ERR_INVALID_CREDENTIALS", response["code"], "Should return ERR_INVALID_CREDENTIALS error")
	assert.Contains(t, response["message"], "Invalid", "Should return error message")
}

// TestAuthHandler_DeleteSession_NotFound tests deleting non-existent session
func TestAuthHandler_DeleteSession_NotFound(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_DeleteSession_OwnerVerification tests user can only delete own sessions
func TestAuthHandler_DeleteSession_OwnerVerification(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_GetSessionInfo_Success tests getting session expiration info
func TestAuthHandler_GetSessionInfo_Success(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_RevokeAllSessions_Success tests admin revoking all user sessions
func TestAuthHandler_RevokeAllSessions_Success(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_RevokeAllSessions_AccessControl tests non-admin cannot revoke
func TestAuthHandler_RevokeAllSessions_AccessControl(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_ExchangeAPIKey_Success tests API key exchange
func TestAuthHandler_ExchangeAPIKey_Success(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_ExchangeAPIKey_InvalidKey tests with invalid API key
func TestAuthHandler_ExchangeAPIKey_InvalidKey(t *testing.T) {
	t.Skip("Requires test database - to be implemented")

	/* Future implementation with test DB:
	gin.SetMode(gin.TestMode)
	router := gin.New()

	pool := setupTestDB(t)
	privateKeyPEM, publicKeyPEM := GenerateTestRSAKeyPair(t)
	handler := auth.NewAuthHandler(pool, privateKeyPEM, publicKeyPEM, "test-key-id", 15, 7, 30, false, auth.RateLimitOptions{})
	router.POST("/api/v1/beacon/token", handler.ExchangeAPIKey)

	reqBody := map[string]string{"api_key": "invalid-key-123"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/beacon/token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 401 for invalid API key
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	*/
}

// TestAuthHandler_ExchangeAPIKey_RateLimited tests API key rate limiting
func TestAuthHandler_ExchangeAPIKey_RateLimited(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_DatabaseFailure_Tests graceful degradation on DB failure
func TestAuthHandler_DatabaseFailure(t *testing.T) {
	t.Skip("Requires test database - to be implemented")
}

// TestAuthHandler_MalformedInput tests handling of malformed JSON input
func TestAuthHandler_MalformedInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/api/v1/auth/login", handler.Login)

	// Send malformed JSON
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 for malformed input
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuthHandler_MissingFields tests handling of missing required fields
func TestAuthHandler_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/api/v1/auth/login", handler.Login)

	// Send empty request body
	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 for missing fields
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuthHandler_TimingAttackResistance tests timing attack prevention
func TestAuthHandler_TimingAttackResistance(t *testing.T) {
	t.Skip("Requires test database - to be implemented with timing measurements")

	// Future implementation:
	// 1. Measure response time for valid username/wrong password
	// 2. Measure response time for invalid username/wrong password
	// 3. Assert times are within acceptable range (100-200ms with artificial delay)
	// 4. Assert no significant timing difference between the two cases
}

// Helper function to create test context
func createTestContext() (context.Context, *gin.Context) {
	ctx := context.Background()
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = httptest.NewRequest("POST", "/test", nil)
	return ctx, ginCtx
}

// Helper function to measure execution time
func measureTime(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// ============ Tests for bug fixes ============

// TestAuthHandler_ExchangeAPIKey_MissingAuthorizationHeader verifies that ExchangeAPIKey
// returns 401 when the Authorization header is absent (no DB access needed for this path).
func TestAuthHandler_ExchangeAPIKey_MissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &AuthHandler{}
	router := gin.New()
	router.POST("/api/v1/beacon/token", handler.ExchangeAPIKey)

	req, _ := http.NewRequest("POST", "/api/v1/beacon/token", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "MISSING_AUTHORIZATION", resp["code"])
}

// TestAuthHandler_ExchangeAPIKey_InvalidBearerFormat verifies that ExchangeAPIKey
// returns 401 when Authorization header is not in "Bearer <token>" format.
func TestAuthHandler_ExchangeAPIKey_InvalidBearerFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &AuthHandler{}
	router := gin.New()
	router.POST("/api/v1/beacon/token", handler.ExchangeAPIKey)

	req, _ := http.NewRequest("POST", "/api/v1/beacon/token", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // Basic auth, not Bearer
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "INVALID_AUTHORIZATION_FORMAT", resp["code"])
}

// TestAuthHandler_ExchangeAPIKey_NoAssociation_Returns401 tests that ExchangeAPIKey returns 401
// when an API key exists in the DB but has no associated user_id or service_account_id.
// This tests the fix ensuring unassociated API keys cannot generate tokens.
func TestAuthHandler_ExchangeAPIKey_NoAssociation_Returns401(t *testing.T) {
	_, handler, _, cleanup := setupAuthHandlerTest(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// Generate a valid RFC 6750-style API key (32 raw bytes, base64 URL-encoded)
	rawKey := make([]byte, 32)
	_, err := rand.Read(rawKey)
	require.NoError(t, err)
	apiKeyStr := base64.URLEncoding.EncodeToString(rawKey)

	// Hash the raw bytes for DB storage (matches ValidateAPIKey logic)
	hashArr := sha256.Sum256(rawKey)
	tokenHash := hex.EncodeToString(hashArr[:])
	keyPrefix := apiKeyStr[:8]

	// Insert API key with user_id = NULL (no user or service account association)
	_, err = handler.pool.Exec(ctx, `
		INSERT INTO api_keys (key_hash, key_prefix, user_id, name, is_active, created_at)
		VALUES ($1, $2, NULL, 'orphan-key', true, NOW())
	`, tokenHash, keyPrefix)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/beacon/token", handler.ExchangeAPIKey)

	req, _ := http.NewRequest("POST", "/api/v1/beacon/token", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKeyStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "API_KEY_NOT_ASSOCIATED", resp["code"])
}

// TestAuthHandler_Logout_RefreshTokenRevocationByHash tests that Logout correctly revokes
// the refresh token by token_hash (not token_id), which was the bug that was fixed.
func TestAuthHandler_Logout_RefreshTokenRevocationByHash(t *testing.T) {
	pool, handler, _, cleanup := setupAuthHandlerTest(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// Create a test user
	userID := uuid.New()
	hashedPassword, _ := HashPassword("TestPass123")
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "logouttest", hashedPassword, "logout@test.com", "admin")
	require.NoError(t, err)

	// Generate a JWT access token for the user
	accessToken, _, err := handler.jwtService.GenerateAccessToken(userID.String(), "admin")
	require.NoError(t, err)

	// Insert a refresh token into DB (raw token string → hash stored)
	refreshTokenStr := "test-refresh-" + uuid.NewString()
	hashArr := sha256.Sum256([]byte(refreshTokenStr))
	tokenHash := hex.EncodeToString(hashArr[:])
	tokenID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at, updated_at)
		VALUES ($1, $2, $3, NOW() + interval '7 days', NOW() + interval '30 days', NOW(), NOW())
	`, tokenID, tokenHash, userID)
	require.NoError(t, err)

	// Call Logout with middleware simulation (user_id set in gin context)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	userIDStr := userID.String()
	router.POST("/api/v1/auth/logout", func(c *gin.Context) {
		c.Set("user_id", userIDStr)
		handler.Logout(c)
	})

	reqBody, _ := json.Marshal(map[string]string{"refresh_token": refreshTokenStr})
	req, _ := http.NewRequest("POST", "/api/v1/auth/logout", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the refresh token was revoked (revoked_at is now set for that token_hash)
	var revokedCount int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NOT NULL`,
		tokenHash,
	).Scan(&revokedCount)
	require.NoError(t, err)
	assert.Equal(t, 1, revokedCount, "refresh token should be marked revoked via token_hash lookup")
}

// TestAuthHandler_RevokeAllMySessions_Success tests the POST /auth/sessions/revoke-all endpoint
// (new route added in the fix). Verifies all of a user's sessions are revoked.
func TestAuthHandler_RevokeAllMySessions_Success(t *testing.T) {
	pool, handler, _, cleanup := setupAuthHandlerTest(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// Create a test user
	userID := uuid.New()
	hashedPassword, _ := HashPassword("TestPass123")
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "revoketest", hashedPassword, "revoke@test.com", "admin")
	require.NoError(t, err)

	// Insert 3 active refresh tokens
	for i := 0; i < 3; i++ {
		tokenID := uuid.New()
		rawToken := fmt.Sprintf("token-%d-%s", i, uuid.NewString())
		hashArr := sha256.Sum256([]byte(rawToken))
		tokenHash := hex.EncodeToString(hashArr[:])
		_, err = pool.Exec(ctx, `
			INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at, updated_at)
			VALUES ($1, $2, $3, NOW() + interval '7 days', NOW() + interval '30 days', NOW(), NOW())
		`, tokenID, tokenHash, userID)
		require.NoError(t, err)
	}

	// Confirm 3 active tokens before revoking
	var beforeCount int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL`, userID).Scan(&beforeCount)
	assert.Equal(t, 3, beforeCount)

	// Call RevokeAllMySessions with middleware simulation
	gin.SetMode(gin.TestMode)
	router := gin.New()
	userIDStr := userID.String()
	router.POST("/api/v1/auth/sessions/revoke-all", func(c *gin.Context) {
		c.Set("user_id", userIDStr)
		handler.RevokeAllMySessions(c)
	})

	req, _ := http.NewRequest("POST", "/api/v1/auth/sessions/revoke-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["message"], "revoked")

	// Verify all tokens are now revoked
	var activeCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL`, userID).Scan(&activeCount)
	require.NoError(t, err)
	assert.Equal(t, 0, activeCount, "all sessions should be revoked")
}

// TestAuthHandler_LogAuditEvent_NullOrEmptyUserID verifies that logAuditEvent correctly
// handles empty or nil userID by inserting NULL (not an empty string) into the UUID column.
// This prevents a PostgreSQL UUID constraint violation that existed before the fix.
func TestAuthHandler_LogAuditEvent_NullOrEmptyUserID(t *testing.T) {
	pool, handler, _, cleanup := setupAuthHandlerTest(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// Test with empty string pointer — should not panic or error
	emptyUserID := ""
	assert.NotPanics(t, func() {
		handler.logAuditEvent(ctx, "test_empty_userid", &emptyUserID, "127.0.0.1", nil)
	})

	var emptyCount int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM auth_audit_logs WHERE event_type = 'test_empty_userid' AND user_id IS NULL
	`).Scan(&emptyCount)
	require.NoError(t, err)
	assert.Equal(t, 1, emptyCount, "audit log for empty userID should have NULL user_id column")

	// Test with nil pointer — should also insert NULL user_id
	assert.NotPanics(t, func() {
		handler.logAuditEvent(ctx, "test_nil_userid", nil, "127.0.0.1", nil)
	})

	var nilCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM auth_audit_logs WHERE event_type = 'test_nil_userid' AND user_id IS NULL
	`).Scan(&nilCount)
	require.NoError(t, err)
	assert.Equal(t, 1, nilCount, "audit log for nil userID should have NULL user_id column")
}

// Benchmark login handler performance
func BenchmarkAuthHandler_Login(b *testing.B) {
	gin.SetMode(gin.TestMode)

	_, handler, _, cleanup := setupAuthHandlerTest(&testing.T{})
	if cleanup == nil {
		b.Skip("Database not available")
		return
	}
	defer cleanup()

	router := gin.New()
	router.POST("/api/v1/auth/login", handler.Login)

	// Create test user
	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, err := HashPassword("correct-password")
	require.NoError(&testing.T{}, err)

	_, err = handler.pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "benchmarkuser", hashedPassword, "benchmark@example.com", "admin")
	if err != nil {
		b.Fatalf("Failed to create test user: %v", err)
	}

	reqBody := map[string]string{"username": "benchmarkuser", "password": "correct-password"}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}

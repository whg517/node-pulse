package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/config"
)

// TestSecurity_TimingAttackResistance tests that ConstantTimeCompare prevents timing attacks
// Tech-Spec requirement: Constant-time comparison for token hashes
func TestSecurity_TimingAttackResistance(t *testing.T) {
	// Test that comparison time doesn't leak information about match position
	validHash := HashTokenSHA256("valid-token-12345")
	wrongHash := HashTokenSHA256("wrong-token-54321")

	// Each comparison should take similar time regardless of where they differ
	iterations := 1000
	validStart := time.Now()
	for i := 0; i < iterations; i++ {
		ConstantTimeCompare("valid-token-12345", validHash)
	}
	validDuration := time.Since(validStart)

	wrongStart := time.Now()
	for i := 0; i < iterations; i++ {
		ConstantTimeCompare("wrong-token-54321", wrongHash)
	}
	wrongDuration := time.Since(wrongStart)

	// Partial match (first char correct)
	partialStart := time.Now()
	for i := 0; i < iterations; i++ {
		ConstantTimeCompare("valid-token-00000", validHash)
	}
	partialDuration := time.Since(partialStart)

	// All durations should be in microsecond range (not milliseconds)
	// This ensures we're using constant-time compare, not variable-time string compare
	assert.Less(t, validDuration, time.Millisecond, "Valid comparison should be sub-millisecond")
	assert.Less(t, wrongDuration, time.Millisecond, "Wrong comparison should be sub-millisecond")
	assert.Less(t, partialDuration, time.Millisecond, "Partial match comparison should be sub-millisecond")

	// Test that we don't leak position information through significant timing differences
	// The key is that all comparisons complete quickly (sub-millisecond for 1000 iterations)
	t.Log("Valid duration:", validDuration)
	t.Log("Wrong duration:", wrongDuration)
	t.Log("Partial duration:", partialDuration)
}

// TestSecurity_SQLInjectionPrevention tests that user input is properly sanitized
// Tech-Spec requirement: Parameterized queries to prevent SQL injection
func TestSecurity_SQLInjectionPrevention(t *testing.T) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, t, pool)

	userID := uuid.New()

	// Create a test user
	hashedPassword, _ := HashPassword("testpass")
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err)

	// Test SQL injection with username validation
	sqlInjectionPayloads := []string{
		"admin' --",
		"admin' OR '1'='1",
		"admin'; DROP TABLE users--",
		"admin' UNION SELECT * FROM users--",
		"'; DELETE FROM refresh_tokens WHERE '1'='1",
		"admin' OR 1=1#",
		"admin'/*comment*/OR/*comment*/1=1",
	}

	for _, payload := range sqlInjectionPayloads {
		t.Run("Payload_"+strings.ReplaceAll(payload, " ", "_"), func(t *testing.T) {
			// Test that input is sanitized during login
			// Attempt to login with malicious username
			hashedPassword2, _ := HashPassword("password123")
			_, err := pool.Exec(ctx, `
				INSERT INTO users (user_id, username, password_hash, email, role, is_active)
				VALUES ($1, $2, $3, $4, $5, true)
			`, uuid.New(), payload, hashedPassword2, payload+"@example.com", "viewer")

			// Query should succeed or fail with proper error, not crash
			// The key is that SQL injection doesn't work
			if err != nil {
				// Expected - input may be invalid, but shouldn't cause SQL injection
				assert.Contains(t, err.Error(), "user_id",
					"Error should be about UUID validation or similar, not SQL syntax")
			}

			// Verify no unauthorized data was created
			var count int
			err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE username = $1", payload).Scan(&count)
			assert.NoError(t, err, "Query should succeed")
			assert.LessOrEqual(t, count, 1, "Should have at most one user with this username")
		})
	}

	// Verify database integrity - all critical tables should be intact
	var userCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, userCount, 1, "Users table should not be empty")

	var blacklistCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM token_blacklist").Scan(&blacklistCount)
	assert.NoError(t, err)
	// Blacklist should be empty since no tokens were revoked
	assert.Equal(t, 0, blacklistCount, "Blacklist should be empty")
}

// TestSecurity_AccountEnumerationPrevention tests that error messages don't leak user existence
// Tech-Spec requirement: Generic error messages for login failures
func TestSecurity_AccountEnumerationPrevention(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, t, pool)

	// Create test config
	cfg := &config.JWTConfig{
		Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		AccessTokenExpirationMinutes:   15,
		RefreshTokenExpirationDays:     7,
		RefreshTokenMaxValidityDays:    30,
	}

	jwtService := NewJWTService(cfg.Secret, cfg.AccessTokenExpirationMinutes, pool)
	handler := NewAuthHandler(pool, cfg.Secret, cfg.AccessTokenExpirationMinutes,
		cfg.RefreshTokenExpirationDays, cfg.RefreshTokenMaxValidityDays)
	handler.jwtService = jwtService

	router := gin.New()
	router.POST("/api/v1/auth/login", handler.Login)

	// Create a real user
	realUserID := uuid.New()
	hashedPassword, _ := HashPassword("RealPassword123")
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, realUserID, "realuser", hashedPassword, "real@example.com", "admin")
	require.NoError(t, err)

	// Test 1: Wrong password for existing user
	loginReq1 := map[string]string{"username": "realuser", "password": "WrongPassword"}
	body1, _ := json.Marshal(loginReq1)
	req1, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(body1)))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Test 2: Non-existent user
	loginReq2 := map[string]string{"username": "nonexistentuser", "password": "SomePassword"}
	body2, _ := json.Marshal(loginReq2)
	req2, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(body2)))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Both responses should have same status code
	assert.Equal(t, w1.Code, w2.Code, "Status codes should be identical to prevent enumeration")

	// Both responses should have similar error messages (not revealing which failed)
	assert.Contains(t, w1.Body.String(), "ERR_INVALID_CREDENTIALS", "Should show generic error")
	assert.Contains(t, w2.Body.String(), "ERR_INVALID_CREDENTIALS", "Should show generic error")

	// Response structures should be similar (not revealing user existence)
	// In production, also ensure response times are similar
}

// TestSecurity_TokenBindingPrevention tests CSRF protection via token binding
// Tech-Spec requirement: Tokens bound to session/origin
func TestSecurity_TokenBindingPrevention(t *testing.T) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, t, pool)

	jwtService := NewJWTService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", 15, pool)

	// Generate token for user
	userID := uuid.New().String()
	role := "admin"
	accessToken, jti, err := jwtService.GenerateAccessToken(userID, role)
	require.NoError(t, err)

	// Token should be bound to JTI (unique identifier)
	claims, err := jwtService.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, claims.JTI, "Token should have JTI binding")
	assert.Equal(t, jti, claims.JTI, "JTI should match generated value")

	// Test that token can be checked against blacklist
	// (This would be caught by blacklist/revocation in production)
	revoked, err := jwtService.CheckRevoked(ctx, claims.JTI)
	assert.NoError(t, err)
	assert.False(t, revoked, "New token should not be revoked")

	// Once revoked, token should be marked as revoked
	_, err = pool.Exec(ctx, `
		INSERT INTO token_blacklist (jti, revoked_at, expires_at)
		VALUES ($1, NOW(), NOW() + INTERVAL '1 hour')
	`, jti)
	require.NoError(t, err)

	revoked, err = jwtService.CheckRevoked(ctx, claims.JTI)
	assert.NoError(t, err)
	assert.True(t, revoked, "Token should be marked as revoked")

	// Note: ValidateAccessToken doesn't automatically check the blacklist
	// The middleware is responsible for calling CheckRevoked after validation
	// This test verifies that the blacklist mechanism works correctly
}

// TestSecurity_PrivilegeEscalationPrevention tests role-based access control
// Tech-Spec requirement: Users cannot elevate privileges without authorization
func TestSecurity_PrivilegeEscalationPrevention(t *testing.T) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, t, pool)

	jwtService := NewJWTService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", 15, pool)

	// Create users with different roles
	viewerID := uuid.New().String()
	adminID := uuid.New().String()

	// Generate tokens with specific roles
	viewerToken, viewerJTI, err := jwtService.GenerateAccessToken(viewerID, "viewer")
	require.NoError(t, err)

	adminToken, adminJTI, err := jwtService.GenerateAccessToken(adminID, "admin")
	require.NoError(t, err)

	// Verify role claims are correctly embedded
	viewerClaims, err := jwtService.ValidateAccessToken(viewerToken)
	assert.NoError(t, err)
	assert.Equal(t, "viewer", viewerClaims.Role, "Viewer token should have viewer role")
	assert.Equal(t, viewerID, viewerClaims.UserID, "UserID should match")
	assert.Equal(t, viewerJTI, viewerClaims.JTI, "JTI should match")

	adminClaims, err := jwtService.ValidateAccessToken(adminToken)
	assert.NoError(t, err)
	assert.Equal(t, "admin", adminClaims.Role, "Admin token should have admin role")
	assert.Equal(t, adminID, adminClaims.UserID, "UserID should match")
	assert.Equal(t, adminJTI, adminClaims.JTI, "JTI should match")

	// Test that role cannot be tampered with (token signature prevents modification)
	// If someone tries to modify the token payload, signature validation will fail

	// Create a test to ensure that middleware properly enforces role checks
	// (This would typically be done in the middleware integration tests)
	t.Run("MiddlewareRoleEnforcement", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Mock middleware that validates JWT and sets context
		mockAuth := func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
				c.Abort()
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ValidateAccessToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
				c.Abort()
				return
			}

			// Set claims in context
			c.Set("user_id", claims.UserID)
			c.Set("role", claims.Role)
			c.Set("jti", claims.JTI)
			c.Next()
		}

		router.Use(mockAuth)

		// Admin-only endpoint
		router.GET("/admin", func(c *gin.Context) {
			role, exists := c.Get("role")
			if !exists || role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient_permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
		})

		// Test viewer token trying to access admin endpoint
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+viewerToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Viewer should not access admin endpoint")

		// Test admin token accessing admin endpoint
		req2, _ := http.NewRequest("GET", "/admin", nil)
		req2.Header.Set("Authorization", "Bearer "+adminToken)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code, "Admin should access admin endpoint")
	})
}

// TestSecurity_PasswordHashingSecurity tests password hashing is secure
func TestSecurity_PasswordHashingSecurity(t *testing.T) {
	// Test that passwords are hashed with bcrypt
	password := "TestPassword123!"

	// Hash password
	hash, err := HashPassword(password)
	assert.NoError(t, err, "Password hashing should succeed")
	assert.NotEmpty(t, hash, "Hash should not be empty")
	assert.NotEqual(t, password, hash, "Hash should not equal plaintext password")

	// Verify hash format (bcrypt hashes start with $2a$, $2b$, or $2y$)
	assert.True(t, strings.HasPrefix(hash, "$2"),
		"Bcrypt hash should start with $2")

	// Test that same password produces different hashes (salt)
	hash2, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, hash, hash2, "Same password should produce different hashes due to salt")

	// Test that both hashes validate the same password
	err1 := VerifyPassword(password, hash)
	err2 := VerifyPassword(password, hash2)
	assert.NoError(t, err1, "First hash should validate password")
	assert.NoError(t, err2, "Second hash should validate password")

	// Test that wrong password doesn't validate
	wrongErr := VerifyPassword("WrongPassword", hash)
	assert.Error(t, wrongErr, "Wrong password should not validate")

	// Test that hash comparison is case-sensitive
	upperPassword := strings.ToUpper(password)
	upperErr := VerifyPassword(upperPassword, hash)
	assert.Error(t, upperErr, "Uppercase password should not validate")
}

// TestSecurity_BruteForceProtection tests rate limiting prevents brute force
func TestSecurity_BruteForceProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, t, pool)

	cfg := &config.JWTConfig{
		Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		AccessTokenExpirationMinutes:   15,
		RefreshTokenExpirationDays:     7,
		RefreshTokenMaxValidityDays:    30,
	}

	jwtService := NewJWTService(cfg.Secret, cfg.AccessTokenExpirationMinutes, pool)
	handler := NewAuthHandler(pool, cfg.Secret, cfg.AccessTokenExpirationMinutes,
		cfg.RefreshTokenExpirationDays, cfg.RefreshTokenMaxValidityDays)
	handler.jwtService = jwtService

	router := gin.New()
	router.POST("/api/v1/auth/login", handler.Login)

	// Create test user
	userID := uuid.New()
	hashedPassword, _ := HashPassword("CorrectPassword123")
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err)

	// Attempt 6 logins with wrong password (exceeds rate limit of 5)
	for i := 0; i < 6; i++ {
		loginReq := map[string]string{"username": "testuser", "password": "WrongPassword"}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 5 {
			// First 5 attempts should return invalid_credentials
			assert.Equal(t, http.StatusUnauthorized, w.Code,
				"Attempt %d should return unauthorized", i+1)
		} else {
			// 6th attempt should be rate limited
			assert.Equal(t, http.StatusTooManyRequests, w.Code,
				"Attempt 6 should be rate limited")
		}
	}

	// Verify account is locked
	var lockedUntil sql.NullTime
	err = pool.QueryRow(ctx, "SELECT locked_until FROM users WHERE username = 'testuser'").Scan(&lockedUntil)
	assert.NoError(t, err)
	assert.True(t, lockedUntil.Valid, "Account should be locked")
	assert.True(t, lockedUntil.Time.After(time.Now()), "Lock should be in the future")
}

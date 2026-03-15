package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/api"
	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/health"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

// setupTestRouterForRegression creates a test router with database connection
func setupTestRouterForRegression(t *testing.T) (*gin.Engine, *pgxpool.Pool) {
	testutil.SetupTestConfig()

	// Load configuration
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), testutil.GetTestDBURL())
	if err != nil {
		t.Skip("No database connection")
		return nil, nil
	}

	// Run database migrations
	if err := db.Migrate(context.Background(), pool); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	router := gin.New()
	healthChecker := health.New(nil, nil, nil) // No scheduler or alert system in tests
	cacheManager := api.SetupRoutes(router, healthChecker, pool)

	// Defer cache cleanup for test cleanup
	t.Cleanup(func() {
		if cacheManager != nil {
			cacheManager.BatchWriter.Stop()
			cacheManager.MemoryCache.Stop()
		}
		testutil.TeardownTestConfig()
		pool.Close()
	})

	return router, pool
}

// cleanupTestUser removes test user from database
func cleanupTestUserForRegression(pool *pgxpool.Pool, username string) {
	_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE username = $1", username)
}

// TestRegression_NodeEndpointsWithNewAuth verifies that existing node management endpoints
// work correctly with the new JWT authentication system (Tech-Spec Task 10.3, AC 8.1)
func TestRegression_NodeEndpointsWithNewAuth(t *testing.T) {
	router, pool := setupTestRouterForRegression(t)
	if router == nil {
		return
	}

	// Clear rate limit store for clean test
	_ = auth.ClearRateLimitStore(context.Background(), pool)

	// Arrange - Create test user and login
	testUserID := uuid.New()
	testPassword := "testPassword"
	passwordHash, _ := auth.HashPassword(testPassword)
	testUsername := fmt.Sprintf("node_regression_%s", testUserID.String()[:8])

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, testUserID, testUsername, passwordHash, "admin")
	require.NoError(t, err)
	defer cleanupTestUserForRegression(pool, testUsername)

	// Login to get JWT token
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

	var loginResp models.LoginResponse
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.Data.AccessToken, "Should have access token")

	accessToken := loginResp.Data.AccessToken

	// Extract CSRF token from login cookies (needed for mutation requests)
	var csrfToken string
	for _, c := range wLogin.Result().Cookies() {
		if c.Name == "csrf_token" {
			csrfToken = c.Value
			break
		}
	}

	// Test 1: GET /api/v1/nodes - List nodes (should work with JWT)
	t.Run("ListNodes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/nodes", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 200 OK with nodes list (empty array is fine)
		assert.Equal(t, http.StatusOK, w.Code, "GET /api/v1/nodes should work with JWT")
	})

	// Test 2: POST /api/v1/nodes - Create node (admin only)
	t.Run("CreateNode", func(t *testing.T) {
		newNode := models.CreateNodeRequest{
			Name:   "test-node-regression",
			IP:     "192.168.1.100",
			Region: "us-west-1",
			Tags:   map[string]interface{}{"env": "test", "tier": "frontend"},
		}
		nodeBody, _ := json.Marshal(newNode)
		req, _ := http.NewRequest("POST", "/api/v1/nodes", bytes.NewBuffer(nodeBody))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 201 Created or 200 OK
		assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusOK,
			"POST /api/v1/nodes should work with JWT, got %d: %s", w.Code, w.Body.String())
	})
}

// TestRegression_ProbeEndpointsWithNewAuth verifies that probe endpoints work with JWT
// (Tech-Spec Task 10.3, AC 8.1)
func TestRegression_ProbeEndpointsWithNewAuth(t *testing.T) {
	router, pool := setupTestRouterForRegression(t)
	if router == nil {
		return
	}

	_ = auth.ClearRateLimitStore(context.Background(), pool)

	// Arrange - Create operator user (can manage probes)
	testUserID := uuid.New()
	testPassword := "testPassword"
	passwordHash, _ := auth.HashPassword(testPassword)
	testUsername := fmt.Sprintf("probe_regression_%s", testUserID.String()[:8])

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, testUserID, testUsername, passwordHash, "operator")
	require.NoError(t, err)
	defer cleanupTestUserForRegression(pool, testUsername)

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

	var loginResp models.LoginResponse
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	accessToken := loginResp.Data.AccessToken

	// Test: GET /api/v1/probes - List probes
	t.Run("ListProbes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/probes", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GET /api/v1/probes should work with JWT")
	})
}

// TestRegression_AlertEndpointsWithNewAuth verifies that alert endpoints work with JWT
// (Tech-Spec Task 10.3, AC 8.1)
func TestRegression_AlertEndpointsWithNewAuth(t *testing.T) {
	router, pool := setupTestRouterForRegression(t)
	if router == nil {
		return
	}

	_ = auth.ClearRateLimitStore(context.Background(), pool)

	// Arrange - Create operator user
	testUserID := uuid.New()
	testPassword := "testPassword"
	passwordHash, _ := auth.HashPassword(testPassword)
	testUsername := fmt.Sprintf("alert_regression_%s", testUserID.String()[:8])

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, testUserID, testUsername, passwordHash, "operator")
	require.NoError(t, err)
	defer cleanupTestUserForRegression(pool, testUsername)

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

	var loginResp models.LoginResponse
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	accessToken := loginResp.Data.AccessToken

	// Test: GET /api/v1/alerts/rules - List alert rules
	t.Run("ListAlertRules", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/alerts/rules", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GET /api/v1/alerts/rules should work with JWT")
	})
}

// TestRegression_RBACWithNewJWT verifies that RBAC (Role-Based Access Control) works correctly
// with the new JWT authentication system (Tech-Spec Task 10.3, AC 8.2)
func TestRegression_RBACWithNewJWT(t *testing.T) {
	router, pool := setupTestRouterForRegression(t)
	if router == nil {
		return
	}

	_ = auth.ClearRateLimitStore(context.Background(), pool)

	// Test 1: Viewer role should NOT be able to create nodes
	t.Run("ViewerCannotCreateNodes", func(t *testing.T) {
		testUserID := uuid.New()
		testPassword := "testPassword"
		passwordHash, _ := auth.HashPassword(testPassword)
		testUsername := fmt.Sprintf("viewer_rbac_%s", testUserID.String()[:8])

		_, err := pool.Exec(context.Background(), `
			INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`, testUserID, testUsername, passwordHash, "viewer")
		require.NoError(t, err)
		defer cleanupTestUserForRegression(pool, testUsername)

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

		var loginResp models.LoginResponse
		err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
		require.NoError(t, err)

		// Extract CSRF token from login cookies
		var viewerCSRF string
		for _, c := range wLogin.Result().Cookies() {
			if c.Name == "csrf_token" {
				viewerCSRF = c.Value
				break
			}
		}

		// Try to create node (should be forbidden by RBAC)
		newNode := models.CreateNodeRequest{
			Name:   "unauthorized-node",
			IP:     "10.0.0.1",
			Region: "us-west-1",
		}
		nodeBody, _ := json.Marshal(newNode)
		req, _ := http.NewRequest("POST", "/api/v1/nodes", bytes.NewBuffer(nodeBody))
		req.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", viewerCSRF)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: viewerCSRF})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 403 Forbidden or 401 Unauthorized
		assert.True(t, w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized,
			"Viewer should not be able to create nodes, got %d", w.Code)
	})

	// Test 2: Admin role SHOULD be able to create nodes
	t.Run("AdminCanCreateNodes", func(t *testing.T) {
		testUserID := uuid.New()
		testPassword := "testPassword"
		passwordHash, _ := auth.HashPassword(testPassword)
		testUsername := fmt.Sprintf("admin_rbac_%s", testUserID.String()[:8])

		_, err := pool.Exec(context.Background(), `
			INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`, testUserID, testUsername, passwordHash, "admin")
		require.NoError(t, err)
		defer cleanupTestUserForRegression(pool, testUsername)

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

		var loginResp models.LoginResponse
		err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
		require.NoError(t, err)

		// Extract CSRF token from login cookies
		var adminCSRF string
		for _, c := range wLogin.Result().Cookies() {
			if c.Name == "csrf_token" {
				adminCSRF = c.Value
				break
			}
		}

		// Try to create node (should succeed)
		newNode := models.CreateNodeRequest{
			Name:   "authorized-admin-node",
			IP:     "10.0.0.2",
			Region: "us-east-1",
		}
		nodeBody, _ := json.Marshal(newNode)
		req, _ := http.NewRequest("POST", "/api/v1/nodes", bytes.NewBuffer(nodeBody))
		req.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", adminCSRF)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: adminCSRF})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should succeed
		assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusOK,
			"Admin should be able to create nodes, got %d: %s", w.Code, w.Body.String())
	})
}

// TestRegression_UserContextPreservation verifies that middleware correctly sets user context
// and handlers can retrieve user information (Tech-Spec Task 10.3)
func TestRegression_UserContextPreservation(t *testing.T) {
	router, pool := setupTestRouterForRegression(t)
	if router == nil {
		return
	}

	_ = auth.ClearRateLimitStore(context.Background(), pool)

	// Arrange - Create test user
	testUserID := uuid.New()
	testPassword := "testPassword"
	passwordHash, _ := auth.HashPassword(testPassword)
	testUsername := fmt.Sprintf("context_%s", testUserID.String()[:8])

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, testUserID, testUsername, passwordHash, "admin")
	require.NoError(t, err)
	defer cleanupTestUserForRegression(pool, testUsername)

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

	var loginResp models.LoginResponse
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	require.NoError(t, err)

	// Test: GET /api/v1/auth/me should return correct user info
	t.Run("GetMeReturnsCorrectUser", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GET /api/v1/auth/me should return 200")

		var resp models.GetMeResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Verify user context is preserved
		assert.Equal(t, testUsername, resp.Data.Username, "Username should match")
		assert.Equal(t, "admin", resp.Data.Role, "Role should be admin")
		assert.NotEmpty(t, resp.Data.UserID, "UserID should not be empty")
	})
}

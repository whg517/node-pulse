package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/pkg/metrics"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---- RateLimit tests ----

func TestInitRateLimiter_Enabled(t *testing.T) {
	// Ensure rate limit enabled
	_ = os.Unsetenv("PULSE_RATE_LIMIT_ENABLED")

	// Reset global state
	rateLimiter = nil
	rateLimiterCancel = nil
	rateLimitEnabled = true

	InitRateLimiter()
	defer ShutdownRateLimiter()

	assert.NotNil(t, rateLimiter)
	assert.True(t, rateLimitEnabled)
}

func TestInitRateLimiter_Disabled(t *testing.T) {
	_ = os.Setenv("PULSE_RATE_LIMIT_ENABLED", "false")
	defer func() {
		_ = os.Unsetenv("PULSE_RATE_LIMIT_ENABLED")
		// Restore to enabled
		rateLimitEnabled = true
		rateLimiter = nil
	}()

	rateLimiter = nil
	rateLimitEnabled = true

	InitRateLimiter()

	assert.False(t, rateLimitEnabled)
	assert.Nil(t, rateLimiter)
}

func TestShutdownRateLimiter_NilCancel(t *testing.T) {
	// Should not panic when cancel is nil
	origCancel := rateLimiterCancel
	rateLimiterCancel = nil
	defer func() { rateLimiterCancel = origCancel }()

	assert.NotPanics(t, func() {
		ShutdownRateLimiter()
	})
}

func TestRateLimitMiddleware_Disabled(t *testing.T) {
	origEnabled := rateLimitEnabled
	origLimiter := rateLimiter
	defer func() {
		rateLimitEnabled = origEnabled
		rateLimiter = origLimiter
	}()

	rateLimitEnabled = false
	rateLimiter = nil

	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_AllowsUnderLimit(t *testing.T) {
	origEnabled := rateLimitEnabled
	origLimiter := rateLimiter
	defer func() {
		rateLimitEnabled = origEnabled
		rateLimiter = origLimiter
	}()

	rateLimitEnabled = true
	rateLimiter = &RateLimiter{
		visitors: make(map[string]*visitor),
		window:   time.Minute,
	}

	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_ExceedsLimit_Unauthenticated(t *testing.T) {
	origEnabled := rateLimitEnabled
	origLimiter := rateLimiter
	defer func() {
		rateLimitEnabled = origEnabled
		rateLimiter = origLimiter
	}()

	rateLimitEnabled = true
	// Pre-populate with the key that will match the request (empty ClientIP when no RemoteAddr)
	rateLimiter = &RateLimiter{
		visitors: map[string]*visitor{
			"ip:": {
				requests: unauthLimit + 1, // already over limit
				timer:    time.Now(),
			},
		},
		window: time.Minute,
	}

	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimitMiddleware_AuthenticatedUser(t *testing.T) {
	origEnabled := rateLimitEnabled
	origLimiter := rateLimiter
	defer func() {
		rateLimitEnabled = origEnabled
		rateLimiter = origLimiter
	}()

	rateLimitEnabled = true
	rateLimiter = &RateLimiter{
		visitors: make(map[string]*visitor),
		window:   time.Minute,
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_WindowExpiry(t *testing.T) {
	origEnabled := rateLimitEnabled
	origLimiter := rateLimiter
	defer func() {
		rateLimitEnabled = origEnabled
		rateLimiter = origLimiter
	}()

	rateLimitEnabled = true
	rateLimiter = &RateLimiter{
		visitors: map[string]*visitor{
			"ip:": {
				requests: unauthLimit + 100,                // way over limit
				timer:    time.Now().Add(-2 * time.Minute), // but window expired
			},
		},
		window: time.Minute,
	}

	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Window expired, so counter reset → allowed
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractUserIDFromBearer(t *testing.T) {
	// A real JWT has three base64url parts: header.payload.signature
	// Construct a minimal valid-structure token with user_id in payload
	// header: {"alg":"RS256","typ":"JWT"}
	// payload: {"user_id":"user-abc","role":"admin"}
	// signature: arbitrary (not verified by extractUserIDFromBearer)
	header := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"
	payload := "eyJ1c2VyX2lkIjoidXNlci1hYmMiLCJyb2xlIjoiYWRtaW4ifQ" // {"user_id":"user-abc","role":"admin"}
	sig := "FAKESIGNATURE"
	token := header + "." + payload + "." + sig

	tests := []struct {
		name       string
		authHeader string
		wantUserID string
	}{
		{
			name:       "valid Bearer token with user_id",
			authHeader: "Bearer " + token,
			wantUserID: "user-abc",
		},
		{
			name:       "no Authorization header",
			authHeader: "",
			wantUserID: "",
		},
		{
			name:       "Basic auth (not Bearer)",
			authHeader: "Basic dXNlcjpwYXNz",
			wantUserID: "",
		},
		{
			name:       "Bearer with invalid JWT structure",
			authHeader: "Bearer notajwt",
			wantUserID: "",
		},
		{
			name:       "Bearer with invalid base64 payload",
			authHeader: "Bearer header.!!!invalid!!!.sig",
			wantUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUserIDFromBearer(tt.authHeader)
			assert.Equal(t, tt.wantUserID, got)
		})
	}
}

func TestRateLimitMiddleware_BearerTokenBeforeJWT(t *testing.T) {
	// Regression test: rate limiter runs BEFORE JWTAuthMiddleware (router-level registration).
	// Authenticated requests must use per-user limit (authLimit=600), not per-IP (unauthLimit=100).
	origEnabled := rateLimitEnabled
	origLimiter := rateLimiter
	defer func() {
		rateLimitEnabled = origEnabled
		rateLimiter = origLimiter
	}()

	rateLimitEnabled = true
	rateLimiter = &RateLimiter{
		visitors: make(map[string]*visitor),
		window:   time.Minute,
	}

	// Pre-fill IP counter beyond unauthLimit but below authLimit
	// to prove requests are bucketed per user, not per IP
	rateLimiter.visitors["ip:"] = &visitor{
		requests: unauthLimit + 10, // over IP limit, but under user limit
		timer:    time.Now(),
	}

	router := gin.New()
	// Rate limiter registered BEFORE JWT (as in routes.go) — no user_id in context yet
	router.Use(RateLimitMiddleware())
	router.GET("/api/v1/webhooks", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Construct a Bearer token with user_id in payload (signature not checked here)
	payload := "eyJ1c2VyX2lkIjoidXNlci1hYmMiLCJyb2xlIjoiYWRtaW4ifQ"
	token := "eyJhbGciOiJSUzI1NiJ9." + payload + ".FAKESIG"

	req, _ := http.NewRequest("GET", "/api/v1/webhooks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be allowed: bucketed as user:user-abc (authLimit=600), not as IP (over unauthLimit)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the user bucket was created (not IP bucket)
	rateLimiter.mu.RLock()
	_, userBucketExists := rateLimiter.visitors["user:user-abc"]
	rateLimiter.mu.RUnlock()
	assert.True(t, userBucketExists, "expected a user:user-abc bucket to be created")
}

func TestIsRateLimitEnabled(t *testing.T) {
	tests := []struct {
		env     string
		enabled bool
	}{
		{"", true},
		{"true", true},
		{"1", true},
		{"yes", true},
		{"false", false},
		{"0", false},
		{"off", false},
		{"no", false},
		{"FALSE", false},
	}

	for _, tt := range tests {
		_ = os.Setenv("PULSE_RATE_LIMIT_ENABLED", tt.env)
		result := isRateLimitEnabled()
		assert.Equal(t, tt.enabled, result, "PULSE_RATE_LIMIT_ENABLED=%q", tt.env)
	}
	_ = os.Unsetenv("PULSE_RATE_LIMIT_ENABLED")
}

// ---- Performance middleware tests ----

func TestDefaultPerformanceConfig(t *testing.T) {
	collector := metrics.NewCollector()
	cfg := DefaultPerformanceConfig(collector)

	assert.Equal(t, collector, cfg.Collector)
	assert.NotEmpty(t, cfg.DashboardEndpoints)
}

func TestPerformanceMiddleware_NilCollector(t *testing.T) {
	cfg := PerformanceConfig{Collector: nil}
	router := gin.New()
	router.Use(PerformanceMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPerformanceMiddleware_APIRequest(t *testing.T) {
	collector := metrics.NewCollector()
	cfg := DefaultPerformanceConfig(collector)

	router := gin.New()
	router.Use(PerformanceMiddleware(cfg))
	router.GET("/api/v1/nodes", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/api/v1/nodes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPerformanceMiddleware_DashboardEndpoint(t *testing.T) {
	collector := metrics.NewCollector()
	cfg := DefaultPerformanceConfig(collector)

	router := gin.New()
	router.Use(PerformanceMiddleware(cfg))
	router.GET("/api/v1/data/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/api/v1/data/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInjectCollector(t *testing.T) {
	collector := metrics.NewCollector()

	router := gin.New()
	router.Use(InjectCollector(collector))
	router.GET("/test", func(c *gin.Context) {
		val, exists := c.Get("metrics_collector")
		require.True(t, exists)
		assert.Equal(t, collector, val)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInjectCollector_Nil(t *testing.T) {
	router := gin.New()
	router.Use(InjectCollector(nil))
	router.GET("/test", func(c *gin.Context) {
		_, exists := c.Get("metrics_collector")
		assert.False(t, exists)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetDBQueryDuration(t *testing.T) {
	collector := metrics.NewCollector()

	router := gin.New()
	router.Use(InjectCollector(collector))
	router.GET("/test", func(c *gin.Context) {
		// Should not panic when collector is properly set
		SetDBQueryDuration(c, "SELECT", 10*time.Millisecond, true)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetDBQueryDuration_NilContext(t *testing.T) {
	// Should not panic with nil context
	assert.NotPanics(t, func() {
		SetDBQueryDuration(nil, "SELECT", 10*time.Millisecond, true)
	})
}

func TestSetDBQueryDuration_NoCollector(t *testing.T) {
	router := gin.New()
	// No InjectCollector middleware
	router.GET("/test", func(c *gin.Context) {
		// Should not panic when no collector in context
		SetDBQueryDuration(c, "SELECT", 10*time.Millisecond, false)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- Auth middleware tests ----

type mockJWTService struct {
	claims   *JWTClaims
	err      error
	revoked  bool
	checkErr error
}

func (m *mockJWTService) ValidateAccessToken(token string) (*JWTClaims, error) {
	return m.claims, m.err
}

func (m *mockJWTService) CheckRevoked(_ context.Context, _ string) (bool, error) {
	return m.revoked, m.checkErr
}

func TestJWTAuthMiddlewareWithInterface_NoHeader(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(&mockJWTService{}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddlewareWithInterface_InvalidFormat(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(&mockJWTService{}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token invalid-format")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddlewareWithInterface_InvalidToken(t *testing.T) {
	svc := &mockJWTService{claims: nil, err: assert.AnError}
	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddlewareWithInterface_BlacklistCheckFails(t *testing.T) {
	svc := &mockJWTService{
		claims:   &JWTClaims{UserID: "user-1", Role: "admin", JTI: "jti-1"},
		err:      nil,
		revoked:  false,
		checkErr: assert.AnError, // blacklist check fails
	}
	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestJWTAuthMiddlewareWithInterface_RevokedToken(t *testing.T) {
	svc := &mockJWTService{
		claims:  &JWTClaims{UserID: "user-1", Role: "admin", JTI: "jti-1"},
		revoked: true,
	}
	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddlewareWithInterface_Success(t *testing.T) {
	svc := &mockJWTService{
		claims:  &JWTClaims{UserID: "user-1", Role: "admin", JTI: "jti-1"},
		revoked: false,
	}
	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(svc))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		role, _ := GetUserRole(c)
		jti, _ := GetJTI(c)
		c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role, "jti": jti})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user-1")
	assert.Contains(t, w.Body.String(), "admin")
}

func TestGetUserID_NotFound(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := GetUserID(c)
	assert.Error(t, err)
}

func TestGetUserRole_NotFound(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := GetUserRole(c)
	assert.Error(t, err)
}

func TestGetJTI_NotFound(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := GetJTI(c)
	assert.Error(t, err)
}

// ---- Error middleware tests ----

func TestErrorResponse_Error(t *testing.T) {
	e := &ErrorResponse{Code: "ERR_TEST", Message: "test error"}
	assert.Equal(t, "[ERR_TEST] test error", e.Error())
}

func TestAppError_Error(t *testing.T) {
	e := &AppError{Code: "APP_ERR", Message: "app error"}
	assert.Equal(t, "[APP_ERR] app error", e.Error())
}

func TestErrorHandler_WithErrorResponse(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		_ = c.Error(&ErrorResponse{Code: "ERR_TEST", Message: "test"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_TEST")
}

func TestErrorHandler_WithAppError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		_ = c.Error(&AppError{Code: "APP_ERR", Message: "app error"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "APP_ERR")
}

// ---- MTLs config tests ----

func TestInitMTLSConfig_Disabled(t *testing.T) {
	_ = os.Setenv("PULSE_MTLS_ENABLED", "disabled")
	defer os.Unsetenv("PULSE_MTLS_ENABLED")

	// Reset once
	mTLSConfigOnce = *new(sync.Once)
	mTLSConfig = nil

	InitMTLSConfig()

	cfg := GetMTLSConfig()
	assert.Equal(t, "disabled", cfg.Mode)
}

func TestInitMTLSConfig_BooleanTrue(t *testing.T) {
	_ = os.Setenv("PULSE_MTLS_ENABLED", "true")
	defer os.Unsetenv("PULSE_MTLS_ENABLED")

	mTLSConfigOnce = *new(sync.Once)
	mTLSConfig = nil

	InitMTLSConfig()

	cfg := GetMTLSConfig()
	assert.Equal(t, "strict", cfg.Mode)
}

func TestInitMTLSConfig_BooleanFalse(t *testing.T) {
	_ = os.Setenv("PULSE_MTLS_ENABLED", "false")
	defer os.Unsetenv("PULSE_MTLS_ENABLED")

	mTLSConfigOnce = *new(sync.Once)
	mTLSConfig = nil

	InitMTLSConfig()

	cfg := GetMTLSConfig()
	assert.Equal(t, "disabled", cfg.Mode)
}

func TestGetEnvOrDefault(t *testing.T) {
	_ = os.Setenv("TEST_ENV_VAR", "testvalue")
	defer os.Unsetenv("TEST_ENV_VAR")

	assert.Equal(t, "testvalue", getEnvOrDefault("TEST_ENV_VAR", "default"))
	assert.Equal(t, "default", getEnvOrDefault("NONEXISTENT_ENV_VAR", "default"))
}

func TestGetEnvInt(t *testing.T) {
	_ = os.Setenv("TEST_INT_ENV", "42")
	defer os.Unsetenv("TEST_INT_ENV")

	assert.Equal(t, 42, getEnvInt("TEST_INT_ENV", 0))
	assert.Equal(t, 10, getEnvInt("NONEXISTENT_INT_ENV", 10))
}

func TestParseIntSafe(t *testing.T) {
	val, err := parseIntSafe("42")
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	_, err = parseIntSafe("not-a-number")
	assert.Error(t, err)
}

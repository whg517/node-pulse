package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

func TestCORSMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		expectCORS     bool
	}{
		{
			name:           "Allowed origin from default list",
			origin:         "http://localhost:3000",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectCORS:     true,
		},
		{
			name:           "Allowed origin from default list - port 5173",
			origin:         "http://localhost:5173",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectCORS:     true,
		},
		{
			name:           "Preflight OPTIONS request",
			origin:         "http://localhost:3000",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			expectCORS:     true,
		},
		{
			name:           "No origin header (same-origin)",
			origin:         "",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectCORS:     false,
		},
		{
			name:           "Disallowed origin",
			origin:         "http://evil.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectCORS:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test config
			testutil.SetupTestConfig()
			defer testutil.TeardownTestConfig()

			// Load config
			config.MustLoad()

			// Unset CORS env to use defaults
			os.Unsetenv("PULSE_CORS_ALLOWED_ORIGINS")
			os.Unsetenv("CORS_ALLOWED_ORIGINS")

			router := gin.New()
			router.Use(CORSMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectCORS {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
			} else if tt.origin != "" {
				// If origin is set but not allowed, no CORS headers should be present
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSMiddleware_CustomOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test config
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Set custom allowed origins
	os.Setenv("PULSE_CORS_ALLOWED_ORIGINS", "https://example.com,https://app.example.com")
	defer os.Unsetenv("PULSE_CORS_ALLOWED_ORIGINS")

	// Load config with custom env vars
	config.MustLoad()

	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tests := []struct {
		name       string
		origin     string
		expectCORS bool
	}{
		{
			name:       "Allowed custom origin",
			origin:     "https://example.com",
			expectCORS: true,
		},
		{
			name:       "Allowed custom origin with subdomain",
			origin:     "https://app.example.com",
			expectCORS: true,
		},
		{
			name:       "Disallowed origin (localhost not in custom list)",
			origin:     "http://localhost:3000",
			expectCORS: false,
		},
		{
			name:       "Disallowed origin (different domain)",
			origin:     "https://evil.com",
			expectCORS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectCORS {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSMiddleware_Wildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test config
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Set wildcard to allow all origins
	os.Setenv("PULSE_CORS_ALLOWED_ORIGINS", "*")
	defer os.Unsetenv("PULSE_CORS_ALLOWED_ORIGINS")

	// Load config with custom env vars
	config.MustLoad()

	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// With wildcard, any origin should be allowed
	assert.Equal(t, "http://any-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

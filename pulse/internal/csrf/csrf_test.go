package csrf

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestGenerateCSRFToken_LengthAndFormat verifies the generated token is the
// expected length and contains only hex-encoded characters.
func TestGenerateCSRFToken_LengthAndFormat(t *testing.T) {
	token, err := GenerateCSRFToken()
	require.NoError(t, err, "GenerateCSRFToken should not return an error")

	// hex.EncodeToString doubles the byte length
	assert.Len(t, token, CSRFTokenLength*2, "token should be %d hex chars", CSRFTokenLength*2)
	_, decodeErr := hex.DecodeString(token)
	assert.NoError(t, decodeErr, "token should be valid hex")
}

// TestGenerateCSRFToken_Uniqueness verifies two generated tokens are different,
// confirming the use of a cryptographically secure random source.
func TestGenerateCSRFToken_Uniqueness(t *testing.T) {
	t1, err := GenerateCSRFToken()
	require.NoError(t, err)
	t2, err := GenerateCSRFToken()
	require.NoError(t, err)

	assert.NotEqual(t, t1, t2, "two generated tokens must differ")
}

// TestSetCSRFCookie verifies the cookie attributes that defend against XSS
// (HttpOnly) and cross-site request forgery (SameSite=Strict).
func TestSetCSRFCookie(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		secure      bool
		wantSecure  bool
	}{
		{"secure flag true", "abc123", true, true},
		{"secure flag false", "abc123", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			SetCSRFCookie(c, tt.token, tt.secure)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()
			cookies := resp.Cookies()
			require.Len(t, cookies, 1, "exactly one cookie should be set")

			cookie := cookies[0]
			assert.Equal(t, CSRFCookieName, cookie.Name)
			assert.Equal(t, tt.token, cookie.Value)
			assert.True(t, cookie.HttpOnly, "cookie must be HttpOnly")
			assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite, "cookie must be SameSite=Strict")
			assert.Equal(t, tt.wantSecure, cookie.Secure)
			assert.Equal(t, 86400*7, cookie.MaxAge, "cookie MaxAge should be 7 days")
			assert.Equal(t, "/", cookie.Path)
		})
	}
}

// TestCSRFMiddleware_SafeMethods verifies GET/HEAD/OPTIONS bypass validation.
func TestCSRFMiddleware_SafeMethods(t *testing.T) {
	safeMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}

	for _, method := range safeMethods {
		t.Run(method, func(t *testing.T) {
			router := gin.New()
			router.Use(CSRFMiddleware())
			reached := false
			router.Handle(method, "/", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/", nil)
			router.ServeHTTP(w, req)

			assert.True(t, reached, "handler should be reached for safe method %s", method)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestCSRFMiddleware_StateChangingMethods covers POST/PUT/DELETE/PATCH paths:
// missing token with valid origin fallback, missing cookie, token mismatch,
// and successful validation.
func TestCSRFMiddleware_StateChangingMethods(t *testing.T) {
	stateChanging := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range stateChanging {
		t.Run(method+"_missing_token_with_valid_origin_passes", func(t *testing.T) {
			router := gin.New()
			router.Use(CSRFMiddleware())
			reached := false
			router.Handle(method, "/", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/", nil)
			req.Host = "example.com"
			req.Header.Set("Origin", "http://example.com")
			router.ServeHTTP(w, req)

			assert.True(t, reached, "handler should be reached via origin fallback")
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run(method+"_missing_token_and_no_origin_rejected", func(t *testing.T) {
			router := gin.New()
			router.Use(CSRFMiddleware())
			reached := false
			router.Handle(method, "/", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/", nil)
			router.ServeHTTP(w, req)

			assert.False(t, reached, "handler must NOT be reached without token or origin")
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), "ERR_CSRF_MISSING")
		})

		t.Run(method+"_token_without_cookie_rejected", func(t *testing.T) {
			router := gin.New()
			router.Use(CSRFMiddleware())
			reached := false
			router.Handle(method, "/", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/", nil)
			req.Header.Set(CSRFHeaderName, "some-token")
			router.ServeHTTP(w, req)

			assert.False(t, reached)
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), "ERR_CSRF_COOKIE_MISSING")
		})

		t.Run(method+"_token_mismatch_rejected", func(t *testing.T) {
			router := gin.New()
			router.Use(CSRFMiddleware())
			reached := false
			router.Handle(method, "/", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/", nil)
			req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "cookie-token"})
			req.Header.Set(CSRFHeaderName, "different-token")
			router.ServeHTTP(w, req)

			assert.False(t, reached)
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), "ERR_CSRF_INVALID")
		})

		t.Run(method+"_matching_token_passes", func(t *testing.T) {
			router := gin.New()
			router.Use(CSRFMiddleware())
			reached := false
			router.Handle(method, "/", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/", nil)
			req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "shared-token"})
			req.Header.Set(CSRFHeaderName, "shared-token")
			router.ServeHTTP(w, req)

			assert.True(t, reached, "matching token should reach handler")
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestValidateOrigin is a table-driven test of the highest-risk function in
// the package: the origin/referer host-extraction logic.
func TestValidateOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		referer string
		host    string
		want    bool
	}{
		// Both headers missing -> reject
		{"empty origin and referer", "", "", "example.com", false},
		// Exact host match via Origin
		{"origin exact host match", "http://example.com", "", "example.com", true},
		{"https origin exact host match", "https://example.com", "", "example.com", true},
		// Origin with path -> host stripped then matched
		{"origin with path", "http://example.com/some/path", "", "example.com", true},
		// Origin host mismatch -> reject
		{"origin host mismatch", "http://evil.com", "", "example.com", false},
		// Referer fallback when Origin empty
		{"referer exact host match", "", "http://example.com/dashboard", "example.com", true},
		{"referer host mismatch", "", "http://evil.com", "example.com", false},
		// Malformed origin (no scheme separator) -> reject origin branch
		{"origin missing scheme separator", "example.com", "", "example.com", false},
		// Referer takes effect even when origin is malformed (mismatch)
		{"malformed origin but valid referer", "noscheme", "http://example.com", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			c.Request.Host = tt.host
			if tt.origin != "" {
				c.Request.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				c.Request.Header.Set("Referer", tt.referer)
			}

			got := validateOrigin(c)
			assert.Equal(t, tt.want, got)
		})
	}
}

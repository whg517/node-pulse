package csrf

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// CSRFTokenLength is the length of the CSRF token in bytes (32 bytes = 256 bits)
	CSRFTokenLength = 32
	// CSRFCookieName is the name of the CSRF cookie
	CSRFCookieName = "csrf_token"
	// CSRFHeaderName is the name of the CSRF header
	CSRFHeaderName = "X-CSRF-Token"
)

// CSRFMiddleware generates and validates CSRF tokens for state-changing operations
// Tokens are generated on login and validated on POST/PUT/DELETE requests
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF validation for safe methods
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// For state-changing methods (POST, PUT, DELETE, PATCH), validate CSRF token
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
			// Get token from header
			token := c.GetHeader(CSRFHeaderName)

			// If no token in header, check origin/referer for fallback validation
			if token == "" {
				if !validateOrigin(c) {
					c.JSON(http.StatusForbidden, gin.H{
						"code":    "ERR_CSRF_MISSING",
						"message": "CSRF token required. Please include X-CSRF-Token header.",
					})
					c.Abort()
					return
				}
				// Origin validation passed, allow request
				c.Next()
				return
			}

			// Get token from cookie
			cookieToken, err := c.Cookie(CSRFCookieName)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    "ERR_CSRF_COOKIE_MISSING",
					"message": "CSRF cookie not found. Please login again.",
				})
				c.Abort()
				return
			}

			// Validate tokens match
			if token != cookieToken {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    "ERR_CSRF_INVALID",
					"message": "Invalid CSRF token. Please login again.",
				})
				c.Abort()
				return
			}

			// Token validated successfully
			c.Next()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken generates a cryptographically secure random CSRF token
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SetCSRFCookie sets the CSRF token as an httpOnly, SameSite=Strict cookie
func SetCSRFCookie(c *gin.Context, token string, secure bool) {
	sameSite := http.SameSiteStrictMode
	c.SetSameSite(sameSite)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
	})
}

// validateOrigin performs fallback validation using Origin/Referer headers
// This provides additional protection when CSRF token is not available
func validateOrigin(c *gin.Context) bool {
	origin := c.Request.Header.Get("Origin")
	referer := c.Request.Header.Get("Referer")

	// If both headers are missing, reject the request
	if origin == "" && referer == "" {
		return false
	}

	// Get the host from the request
	host := c.Request.Host

	// Check Origin header if present
	if origin != "" {
		// Parse origin URL and check for exact host match
		if idx := strings.Index(origin, "://"); idx != -1 {
			originHost := origin[idx+3:]
			if slashIdx := strings.Index(originHost, "/"); slashIdx != -1 {
				originHost = originHost[:slashIdx]
			}
			if originHost == host {
				return true
			}
		}
	}

	// Check Referer header if present
	if referer != "" {
		// Parse referer URL and check for exact host match
		if idx := strings.Index(referer, "://"); idx != -1 {
			refererHost := referer[idx+3:]
			if slashIdx := strings.Index(refererHost, "/"); slashIdx != -1 {
				refererHost = refererHost[:slashIdx]
			}
			if refererHost == host {
				return true
			}
		}
	}

	return false
}

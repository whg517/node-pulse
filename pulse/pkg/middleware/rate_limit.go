package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// unauthLimit is the per-IP limit for unauthenticated requests (anti-abuse).
	unauthLimit = 100
	// authLimit is the per-user limit for authenticated requests.
	// Dashboard + node detail + alerts can easily generate 10+ requests/page-load;
	// 600/min gives ~10 req/s headroom for normal multi-tab usage.
	authLimit       = 600
	rateLimitWindow = time.Minute
)

var (
	rateLimiter       *RateLimiter
	rateLimiterCancel context.CancelFunc
	rateLimitEnabled  = true
)

type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	window   time.Duration
}

type visitor struct {
	requests int
	timer    time.Time
}

func InitRateLimiter() {
	rateLimitEnabled = isRateLimitEnabled()
	if !rateLimitEnabled {
		return
	}

	rateLimiter = &RateLimiter{
		visitors: make(map[string]*visitor),
		window:   rateLimitWindow,
	}

	// Start cleanup goroutine with cancel support
	ctx, cancel := context.WithCancel(context.Background())
	rateLimiterCancel = cancel
	go cleanupStaleVisitors(ctx)
}

func cleanupStaleVisitors(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return // Exit gracefully when context is cancelled
		case <-time.After(time.Minute):
			rateLimiter.mu.Lock()
			for key, v := range rateLimiter.visitors {
				if time.Since(v.timer) > rateLimiter.window {
					delete(rateLimiter.visitors, key)
				}
			}
			rateLimiter.mu.Unlock()
		}
	}
}

func ShutdownRateLimiter() {
	if rateLimiterCancel != nil {
		rateLimiterCancel()
	}
}

// extractUserIDFromBearer extracts the user_id claim from a JWT Bearer token
// without performing signature verification. This is intentionally unverified -
// it is only used to bucket rate limit keys so that authenticated users get the
// higher per-user limit instead of the per-IP limit. The actual token validity
// is enforced by JWTAuthMiddleware downstream.
func extractUserIDFromBearer(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	parts := strings.Split(authHeader[7:], ".")
	if len(parts) != 3 {
		return ""
	}
	// JWT payload is base64url-encoded without padding
	payload := parts[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Try without padding correction
		decoded, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return ""
		}
	}
	var claims struct {
		UserID string `json:"user_id"`
		Sub    string `json:"sub"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return ""
	}
	if claims.UserID != "" {
		return claims.UserID
	}
	return claims.Sub
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rateLimitEnabled || rateLimiter == nil {
			c.Next()
			return
		}

		// Prefer user_id set by JWTAuthMiddleware (when rate limiter runs after JWT).
		// If not set yet (rate limiter runs before JWT at router level), extract it
		// from the Bearer token payload without signature verification — this is only
		// used for key bucketing; actual auth is enforced by JWTAuthMiddleware.
		userID := c.GetString("user_id")
		if userID == "" {
			userID = extractUserIDFromBearer(c.GetHeader("Authorization"))
		}

		var key string
		var limit int
		if userID != "" {
			key = fmt.Sprintf("user:%s", userID)
			limit = authLimit
		} else {
			key = fmt.Sprintf("ip:%s", c.ClientIP())
			limit = unauthLimit
		}

		rateLimiter.mu.Lock()
		defer rateLimiter.mu.Unlock()

		v, exists := rateLimiter.visitors[key]
		if !exists {
			v = &visitor{
				requests: 1,
				timer:    time.Now(),
			}
			rateLimiter.visitors[key] = v
		} else {
			if time.Since(v.timer) > rateLimiter.window {
				v.requests = 1
				v.timer = time.Now()
			} else {
				v.requests++
			}

			if v.requests > limit {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    "ERR_RATE_LIMIT_EXCEEDED",
					"message": "请求过于频繁，请稍后再试",
					"details": map[string]interface{}{
						"limit":    limit,
						"window":   rateLimiter.window.String(),
						"requests": v.requests,
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func isRateLimitEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("PULSE_RATE_LIMIT_ENABLED")))
	return v != "false" && v != "0" && v != "off" && v != "no"
}

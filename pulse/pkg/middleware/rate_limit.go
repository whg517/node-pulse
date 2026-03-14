package middleware

import (
	"context"
	"fmt"
	"net/http"
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
	authLimit     = 600
	rateLimitWindow = time.Minute
)

var (
	rateLimiter       *RateLimiter
	rateLimiterCancel context.CancelFunc
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

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prefer user_id set by JWTAuthMiddleware; fall back to IP for unauthed requests.
		userID := c.GetString("user_id")

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

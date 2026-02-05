package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	defaultLimit = 100
	defaultWindow = time.Minute
	rateLimiter   *RateLimiter
	rateLimiterCancel context.CancelFunc
)

type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	requests int
	timer    time.Time
}

func InitRateLimiter() {
	rateLimiter = &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    defaultLimit,
		window:   defaultWindow,
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
			for ip, v := range rateLimiter.visitors {
				if time.Since(v.timer) > rateLimiter.window {
					delete(rateLimiter.visitors, ip)
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
		ip := c.ClientIP()

		rateLimiter.mu.Lock()
		defer rateLimiter.mu.Unlock()

		v, exists := rateLimiter.visitors[ip]
		if !exists {
			v = &visitor{
				requests: 1,
				timer:    time.Now(),
			}
			rateLimiter.visitors[ip] = v
		} else {
			if time.Since(v.timer) > rateLimiter.window {
				v.requests = 1
				v.timer = time.Now()
			} else {
				v.requests++
			}

			if v.requests > rateLimiter.limit {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    "ERR_RATE_LIMIT_EXCEEDED",
					"message": "请求过于频繁，请稍后再试",
					"details": map[string]interface{}{
						"limit":    rateLimiter.limit,
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

package auth

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all authentication-related Prometheus metrics
var (
	// loginAttemptsTotal tracks total login attempts by result (success/failure/locked)
	loginAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"result"}, // success, failure, locked
	)

	// loginDurationSeconds tracks login request duration
	loginDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_login_duration_seconds",
			Help:    "Duration of login requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"}, // success, failure, locked
	)

	// refreshTokenRotationsTotal tracks refresh token rotations
	refreshTokenRotationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_refresh_token_rotations_total",
			Help: "Total number of refresh token rotations",
		},
		[]string{"result"}, // success, conflict, expired
	)

	// refreshTokenDurationSeconds tracks refresh request duration
	refreshTokenDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_refresh_token_duration_seconds",
			Help:    "Duration of refresh token requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"}, // success, conflict, expired
	)

	// authValidationDurationSeconds tracks JWT validation duration
	authValidationDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_validation_duration_seconds",
			Help:    "Duration of JWT validation in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"}, // success, invalid, expired, revoked
	)

	// blacklistChecksTotal tracks blacklist check operations
	blacklistChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_blacklist_checks_total",
			Help: "Total number of token blacklist checks",
		},
		[]string{"result"}, // found, not_found
	)

	// blacklistSize tracks the current size of the token blacklist
	blacklistSize prometheus.Gauge

	// apiKeyExchangesTotal tracks API key exchange attempts
	apiKeyExchangesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_api_key_exchanges_total",
			Help: "Total number of API key exchange attempts",
		},
		[]string{"result"}, // success, invalid, rate_limited
	)

	// rateLimitChecksTotal tracks rate limit check operations
	rateLimitChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_rate_limit_checks_total",
			Help: "Total number of rate limit checks",
		},
		[]string{"endpoint", "result"}, // endpoint: login, refresh, apikey; result: allowed, denied
	)

	// activeRefreshTokens tracks the number of active refresh tokens
	activeRefreshTokens prometheus.Gauge

	// activeUsers tracks the number of users with active sessions
	activeUsers prometheus.Gauge

	// tokenGenerationTotal tracks token generation operations
	tokenGenerationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_token_generation_total",
			Help: "Total number of tokens generated",
		},
		[]string{"type"}, // access, refresh
	)

	// sessionOperationsTotal tracks session management operations
	sessionOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_session_operations_total",
			Help: "Total number of session management operations",
		},
		[]string{"operation"}, // list, revoke, revoke_all
	)
)

func init() {
	// Initialize gauges that need Get() method
	blacklistSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "auth_blacklist_size",
		Help: "Current number of entries in the token blacklist",
	})
	prometheus.MustRegister(blacklistSize)

	activeRefreshTokens = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "auth_active_refresh_tokens",
		Help: "Current number of active (non-revoked) refresh tokens",
	})
	prometheus.MustRegister(activeRefreshTokens)

	activeUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "auth_active_users",
		Help: "Current number of users with active refresh tokens",
	})
	prometheus.MustRegister(activeUsers)
}

// RecordLoginAttempt records a login attempt with result and duration
func RecordLoginAttempt(result string, duration time.Duration) {
	loginAttemptsTotal.WithLabelValues(result).Inc()
	loginDurationSeconds.WithLabelValues(result).Observe(duration.Seconds())
}

// RecordRefreshRotation records a refresh token rotation with result and duration
func RecordRefreshRotation(result string, duration time.Duration) {
	refreshTokenRotationsTotal.WithLabelValues(result).Inc()
	refreshTokenDurationSeconds.WithLabelValues(result).Observe(duration.Seconds())
}

// RecordAuthValidation records a JWT validation with result and duration
func RecordAuthValidation(result string, duration time.Duration) {
	authValidationDurationSeconds.WithLabelValues(result).Observe(duration.Seconds())
}

// RecordBlacklistCheck records a blacklist check operation
func RecordBlacklistCheck(found bool) {
	result := "not_found"
	if found {
		result = "found"
	}
	blacklistChecksTotal.WithLabelValues(result).Inc()
}

// UpdateBlacklistSize updates the current blacklist size
func UpdateBlacklistSize(size float64) {
	blacklistSize.Set(size)
}

// RecordAPIKeyExchange records an API key exchange attempt
func RecordAPIKeyExchange(result string) {
	apiKeyExchangesTotal.WithLabelValues(result).Inc()
}

// RecordRateLimitCheck records a rate limit check
func RecordRateLimitCheck(endpoint, result string) {
	rateLimitChecksTotal.WithLabelValues(endpoint, result).Inc()
}

// UpdateActiveRefreshTokens updates the count of active refresh tokens
func UpdateActiveRefreshTokens(count float64) {
	activeRefreshTokens.Set(count)
}

// UpdateActiveUsers updates the count of active users
func UpdateActiveUsers(count float64) {
	activeUsers.Set(count)
}

// RecordTokenGeneration records a token generation operation
func RecordTokenGeneration(tokenType string) {
	tokenGenerationTotal.WithLabelValues(tokenType).Inc()
}

// RecordSessionOperation records a session management operation
func RecordSessionOperation(operation string) {
	sessionOperationsTotal.WithLabelValues(operation).Inc()
}

// StartMetricsCollection starts periodic metrics collection (e.g., blacklist size)
func StartMetricsCollection(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Collect immediately on start
	collectMetrics(ctx, pool)

	for {
		select {
		case <-ticker.C:
			collectMetrics(ctx, pool)
		case <-ctx.Done():
			log.Println("[Metrics] Stopping metrics collection")
			return
		}
	}
}

// collectMetrics gathers current metrics from the database
func collectMetrics(ctx context.Context, pool *pgxpool.Pool) {
	// Update blacklist size
	var blacklistCount int64
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM token_blacklist").Scan(&blacklistCount)
	if err != nil {
		log.Printf("[Metrics] Error querying blacklist size: %v", err)
	} else {
		UpdateBlacklistSize(float64(blacklistCount))
	}

	// Update active refresh tokens count
	var activeTokenCount int64
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM refresh_tokens WHERE revoked_at IS NULL AND expires_at > NOW()").Scan(&activeTokenCount)
	if err != nil {
		log.Printf("[Metrics] Error querying active refresh tokens: %v", err)
	} else {
		UpdateActiveRefreshTokens(float64(activeTokenCount))
	}

	// Update active users count
	var userCount int64
	err = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT user_id) FROM refresh_tokens WHERE revoked_at IS NULL AND expires_at > NOW()").Scan(&userCount)
	if err != nil {
		log.Printf("[Metrics] Error querying active users: %v", err)
	} else {
		UpdateActiveUsers(float64(userCount))
	}
}

// GetMetricsAsJSON returns metrics summary as JSON string for debugging
// Note: This uses Write() which is the proper way to read metric values
func GetMetricsAsJSON() string {
	// This is a simplified version for debugging
	// In production, use Prometheus HTTP endpoint to scrape metrics
	return `{
		"note": "Use /metrics endpoint for full Prometheus scraping",
		"login_attempts": "enabled",
		"refresh_rotations": "enabled",
		"blacklist_checks": "enabled",
		"api_key_exchanges": "enabled",
		"rate_limiting": "enabled"
	}`
}

// Note: To expose metrics for Prometheus scraping, use:
// import "github.com/prometheus/client_golang/prometheus/promhttp"
// router.GET("/metrics", gin.WrapH(promhttp.Handler()))


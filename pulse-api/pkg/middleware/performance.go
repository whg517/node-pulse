package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kevin/node-pulse/pulse-api/pkg/metrics"
)

// PerformanceConfig configures the performance middleware
type PerformanceConfig struct {
	// Collector is the metrics collector to record metrics
	Collector *metrics.Collector
	// DashboardEndpoints defines which endpoints are considered dashboard endpoints
	DashboardEndpoints []string
}

// DefaultPerformanceConfig returns default performance middleware configuration
func DefaultPerformanceConfig(collector *metrics.Collector) PerformanceConfig {
	return PerformanceConfig{
		Collector: collector,
		DashboardEndpoints: []string{
			"/api/v1/data/metrics",
			"/api/v1/data/history",
			"/api/v1/data/comparison",
			"/api/v1/data/diagnosis",
		},
	}
}

// PerformanceMiddleware creates a middleware that tracks API performance metrics
func PerformanceMiddleware(config PerformanceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if collector is not configured
		if config.Collector == nil {
			c.Next()
			return
		}

		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Determine metric type
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		metricType := metrics.MetricTypeAPI
		for _, dashEndpoint := range config.DashboardEndpoints {
			if endpoint == dashEndpoint {
				metricType = metrics.MetricTypeDashboard
				break
			}
		}

		// Record metric based on type
		if metricType == metrics.MetricTypeDashboard {
			config.Collector.RecordDashboardLoad(
				endpoint,
				c.Request.Method,
				duration,
				c.Writer.Status(),
			)
		} else {
			config.Collector.RecordAPIRequest(
				endpoint,
				c.Request.Method,
				duration,
				c.Writer.Status(),
			)
		}
	}
}

// SetDBQueryDuration is a helper function to record database query metrics
// This should be called from database query methods
func SetDBQueryDuration(c *gin.Context, queryType string, duration time.Duration, success bool) {
	if c == nil {
		return
	}

	// Get collector from context
	collector, exists := c.Get("metrics_collector")
	if !exists {
		return
	}

	metricsCollector, ok := collector.(*metrics.Collector)
	if !ok {
		return
	}

	metricsCollector.RecordDatabaseQuery(queryType, duration, success)
}

// InjectCollector injects the metrics collector into the Gin context
// This allows database query methods to access the collector
func InjectCollector(collector *metrics.Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		if collector != nil {
			c.Set("metrics_collector", collector)
		}
		c.Next()
	}
}

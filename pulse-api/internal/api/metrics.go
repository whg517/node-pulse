package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kevin/node-pulse/pulse-api/pkg/metrics"
)

// MetricsHandler handles performance metrics endpoints
type MetricsHandler struct {
	collector *metrics.Collector
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(collector *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{
		collector: collector,
	}
}

// GetPerformanceMetrics handles GET /api/v1/metrics/performance
func (h *MetricsHandler) GetPerformanceMetrics(c *gin.Context) {
	// Parse query parameters
	filter := h.parseFilterParams(c)

	// Get metrics
	aggregated, summary := h.collector.GetMetrics(filter)

	// Build response
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"metrics": aggregated,
			"summary": summary,
		},
		"message":   "Performance metrics retrieved successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetCollectorStats handles GET /api/v1/metrics/stats
func (h *MetricsHandler) GetCollectorStats(c *gin.Context) {
	stats := h.collector.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"data":      stats,
		"message":   "Collector statistics retrieved successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// parseFilterParams parses filter parameters from the request
func (h *MetricsHandler) parseFilterParams(c *gin.Context) metrics.MetricsFilter {
	filter := metrics.MetricsFilter{}

	// Parse metric_type
	if metricTypeStr := c.Query("metric_type"); metricTypeStr != "" {
		var metricType metrics.MetricType
		switch metricTypeStr {
		case "api":
			metricType = metrics.MetricTypeAPI
		case "dashboard":
			metricType = metrics.MetricTypeDashboard
		case "database":
			metricType = metrics.MetricTypeDatabase
		default:
			// Invalid metric type, will return no results
			metricType = metrics.MetricType("")
		}
		filter.MetricType = &metricType
	}

	// Parse endpoint
	filter.Endpoint = c.Query("endpoint")

	// Parse start_time (default: 1 hour ago)
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = startTime
		} else {
			// Invalid format, use default
			filter.StartTime = time.Now().Add(-1 * time.Hour)
		}
	} else {
		filter.StartTime = time.Now().Add(-1 * time.Hour)
	}

	// Parse end_time (default: now)
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = endTime
		} else {
			// Invalid format, use default
			filter.EndTime = time.Now()
		}
	} else {
		filter.EndTime = time.Now()
	}

	// Parse aggregation (default: 1m)
	if aggregationStr := c.Query("aggregation"); aggregationStr != "" {
		if aggregation, err := parseAggregation(aggregationStr); err == nil {
			filter.Aggregation = aggregation
		} else {
			// Invalid format, use default
			filter.Aggregation = time.Minute
		}
	} else {
		filter.Aggregation = time.Minute
	}

	return filter
}

// parseAggregation parses aggregation duration string
func parseAggregation(s string) (time.Duration, error) {
	// Parse format like "1m", "5m", "15m"
	val, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return 0, err
	}

	unit := s[len(s)-1:]
	switch unit {
	case "m":
		return time.Duration(val) * time.Minute, nil
	case "s":
		return time.Duration(val) * time.Second, nil
	case "h":
		return time.Duration(val) * time.Hour, nil
	default:
		return 0, err
	}
}

// GetPerformanceMetricsLegacy is a legacy endpoint that returns metrics in a different format
// This is kept for backward compatibility
func (h *MetricsHandler) GetPerformanceMetricsLegacy(c *gin.Context) {
	// Parse query parameters
	filter := h.parseFilterParams(c)

	// Get metrics
	aggregated, summary := h.collector.GetMetrics(filter)

	// Format response differently for legacy format
	type LegacyMetric struct {
		TimeWindow   string  `json:"time_window"`
		Endpoint     string  `json:"endpoint"`
		MetricType   string  `json:"metric_type"`
		Count        int64   `json:"count"`
		AvgDuration  float64 `json:"avg_duration_ms"`
		P95          float64 `json:"p95_ms"`
		P99          float64 `json:"p99_ms"`
		SuccessRate  float64 `json:"success_rate"`
	}

	var legacyMetrics []LegacyMetric
	for _, agg := range aggregated {
		legacyMetrics = append(legacyMetrics, LegacyMetric{
			TimeWindow:  agg.TimeWindow.Format(time.RFC3339),
			Endpoint:    agg.Endpoint,
			MetricType:  agg.MetricType,
			Count:       agg.Count,
			AvgDuration: agg.AvgDuration,
			P95:         agg.P95,
			P99:         agg.P99,
			SuccessRate: agg.SuccessRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "OK",
		"data": gin.H{
			"metrics": legacyMetrics,
			"summary": summary,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

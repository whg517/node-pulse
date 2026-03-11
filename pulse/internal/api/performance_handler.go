package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/pkg/metrics"
)

// GetPerformanceData handles GET /api/v1/data/performance
// Returns performance metrics with targets, anomaly detection, and trend data
// @Summary		Get performance data with targets
// @Description	Returns performance metrics with SLA targets, trend data, anomaly detection, and overall system health status.
// @Tags			Metrics
// @Accept			json
// @Produce		json
// @Param			time_range	query		string					false	"Time range (e.g. 24h, 7d)"	default(24h)
// @Success		200	{object}	map[string]interface{}	"Performance data with targets and health status"
// @Failure		400	{object}	map[string]interface{}	"Invalid time_range parameter"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Security		BearerAuth
// @Router			/data/performance [get]
func (h *MetricsHandler) GetPerformanceData(c *gin.Context) {
	// Parse time range parameter (default 24 hours)
	timeRange := c.DefaultQuery("time_range", "24h")
	duration, err := parseTimeRange(timeRange)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_TIME_RANGE",
			"message": "Invalid time_range parameter. Use format like '24h', '7d'",
		})
		return
	}

	// Create filter for metrics collector
	filter := metrics.MetricsFilter{
		StartTime:   time.Now().Add(-duration),
		EndTime:     time.Now(),
		Aggregation: time.Minute,
	}

	// Get metrics from collector
	aggregated, summary := h.collector.GetMetrics(filter)

	// Build performance metrics response
	perfMetrics := buildPerformanceMetrics(aggregated)
	trendData := buildTrendData(aggregated)
	systemHealth, anomalies := evaluateSystemHealth(perfMetrics)

	// Build summary
	perfSummary := models.PerformanceSummary{
		TotalRequests:   summary.TotalRequests,
		AvgResponseTime: summary.OverallAvg,
		MaxResponseTime: summary.OverallP99,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": models.PerformanceDataResponse{
			Metrics:      perfMetrics,
			TrendData:    trendData,
			SystemHealth: systemHealth,
			Anomalies:    anomalies,
			Summary:      perfSummary,
		},
		"message":   "性能数据查询成功",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// parseTimeRange parses time range string like "24h", "7d" into duration
func parseTimeRange(s string) (time.Duration, error) {
	// Default to 24 hours if empty
	if s == "" {
		return 24 * time.Hour, nil
	}

	if len(s) < 2 {
		return 0, &time.ParseError{}
	}

	// Parse numeric value
	var val int
	for i := 0; i < len(s)-1; i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, &time.ParseError{}
		}
		val = val*10 + int(s[i]-'0')
	}

	// Parse unit
	unit := s[len(s)-1:]
	switch unit {
	case "h":
		return time.Duration(val) * time.Hour, nil
	case "d":
		return time.Duration(val) * 24 * time.Hour, nil
	default:
		return 0, &time.ParseError{}
	}
}

// buildPerformanceMetrics converts aggregated metrics to performance metrics with targets
func buildPerformanceMetrics(aggregated []metrics.AggregatedMetrics) []models.PerformanceMetric {
	// Aggregate by metric type
	metricGroups := make(map[string][]metrics.AggregatedMetrics)
	for _, agg := range aggregated {
		metricGroups[agg.MetricType] = append(metricGroups[agg.MetricType], agg)
	}

	// Map metric types to target names
	metricTypeMapping := map[string]string{
		"dashboard": "dashboard_load_time",
		"api":       "api_response_time",
		"database":  "data_query_time",
	}

	displayNameMapping := map[string]string{
		"dashboard": "仪表盘加载时间",
		"api":       "API 响应时间",
		"database":  "数据查询时间",
	}

	var perfMetrics []models.PerformanceMetric

	for metricType, aggs := range metricGroups {
		targetName := metricTypeMapping[metricType]
		target := models.GetTarget(targetName)
		if target == nil {
			continue
		}

		// Calculate latest P99 and P95
		var latestP99, latestP95 float64
		if len(aggs) > 0 {
			latest := aggs[len(aggs)-1]
			latestP99 = latest.P99
			latestP95 = latest.P95
		}

		// Determine status and check for anomalies
		status := "healthy"
		var anomaly string

		if latestP99 > target.TargetP99 {
			status = "unhealthy"
			anomaly = "P99 超过目标值"
		} else if latestP95 > target.TargetP95 {
			status = "unhealthy"
			anomaly = "P95 超过目标值"
		}

		perfMetric := models.PerformanceMetric{
			MetricName:  targetName,
			DisplayName: displayNameMapping[metricType],
			CurrentP99:  latestP99,
			CurrentP95:  latestP95,
			TargetP99:   target.TargetP99,
			TargetP95:   target.TargetP95,
			Unit:        target.Unit,
			Status:      status,
		}

		if anomaly != "" {
			perfMetric.Anomaly = anomaly
		}

		perfMetrics = append(perfMetrics, perfMetric)
	}

	// Ensure all target metrics are present, even with no data
	for _, target := range models.PerformanceTargets {
		found := false
		for _, pm := range perfMetrics {
			if pm.MetricName == target.MetricName {
				found = true
				break
			}
		}
		if !found {
			perfMetrics = append(perfMetrics, models.PerformanceMetric{
				MetricName:  target.MetricName,
				DisplayName: target.DisplayName,
				CurrentP99:  0,
				CurrentP95:  0,
				TargetP99:   target.TargetP99,
				TargetP95:   target.TargetP95,
				Unit:        target.Unit,
				Status:      "healthy",
			})
		}
	}

	return perfMetrics
}

// buildTrendData creates trend data for charts
func buildTrendData(aggregated []metrics.AggregatedMetrics) []models.MetricTrendData {
	// Group by metric type
	metricGroups := make(map[string][]metrics.AggregatedMetrics)
	for _, agg := range aggregated {
		metricGroups[agg.MetricType] = append(metricGroups[agg.MetricType], agg)
	}

	metricTypeMapping := map[string]string{
		"dashboard": "dashboard_load_time",
		"api":       "api_response_time",
		"database":  "data_query_time",
	}

	var trendData []models.MetricTrendData

	for metricType, aggs := range metricGroups {
		targetName := metricTypeMapping[metricType]
		if targetName == "" {
			continue
		}

		var dataPoints []models.TrendDataPoint
		for _, agg := range aggs {
			dataPoints = append(dataPoints, models.TrendDataPoint{
				Timestamp: agg.TimeWindow.Format(time.RFC3339),
				P99:       agg.P99,
				P95:       agg.P95,
			})
		}

		trendData = append(trendData, models.MetricTrendData{
			MetricName: targetName,
			DataPoints: dataPoints,
		})
	}

	return trendData
}

// evaluateSystemHealth determines overall system health and identifies anomalies
func evaluateSystemHealth(perfMetrics []models.PerformanceMetric) (string, []models.Anomaly) {
	systemHealth := "healthy"
	var anomalies []models.Anomaly

	// Check if we have any actual data (all metrics showing 0 means no data yet)
	hasData := false
	for _, metric := range perfMetrics {
		if metric.CurrentP99 > 0 || metric.CurrentP95 > 0 {
			hasData = true
			break
		}
	}

	// If no data yet, system is healthy (default state)
	if !hasData {
		return "healthy", anomalies
	}

	for _, metric := range perfMetrics {
		if metric.Status == "unhealthy" {
			systemHealth = "unhealthy"

			// Determine severity based on PRD thresholds
			// P0 (Critical): Dashboard P99 > 5s (5000ms) OR API P99 > 1000ms OR Query P99 > 600ms
			// P1 (Warning): Exceeds target but below P0 threshold
			severity := "P1" // Warning
			if (metric.MetricName == "dashboard_load_time" && metric.CurrentP99 > 5000) ||
				(metric.MetricName == "api_response_time" && metric.CurrentP99 > 1000) ||
				(metric.MetricName == "data_query_time" && metric.CurrentP99 > 600) {
				severity = "P0" // Critical
			}

			anomalies = append(anomalies, models.Anomaly{
				MetricName: metric.MetricName,
				Severity:   severity,
				Message:    metric.DisplayName + " " + metric.Anomaly,
			})
		}
	}

	return systemHealth, anomalies
}

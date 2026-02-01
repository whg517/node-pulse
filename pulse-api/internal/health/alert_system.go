package health

import (
	"context"
	"time"

	"github.com/kevin/node-pulse/pulse-api/internal/db"
)

// AlertSystemChecker provides alert system health check functionality
type AlertSystemChecker struct {
	alertEngine            AlertEngineStats
	webhookLogsQuerier     db.WebhookLogsQuerier
	alertSuppressionsQuerier db.AlertSuppressionsQuerier
}

// AlertEngineStats defines interface for getting alert engine statistics
type AlertEngineStats interface {
	GetStats() map[string]interface{}
}

// NewAlertSystemChecker creates a new alert system health checker
func NewAlertSystemChecker(
	alertEngine AlertEngineStats,
	webhookLogsQuerier db.WebhookLogsQuerier,
	alertSuppressionsQuerier db.AlertSuppressionsQuerier,
) *AlertSystemChecker {
	return &AlertSystemChecker{
		alertEngine:            alertEngine,
		webhookLogsQuerier:     webhookLogsQuerier,
		alertSuppressionsQuerier: alertSuppressionsQuerier,
	}
}

// CheckAlertEngine checks the alert engine health
func (c *AlertSystemChecker) CheckAlertEngine(ctx context.Context) (*AlertEngineStatus, error) {
	// Get stats from alert engine
	stats := c.alertEngine.GetStats()

	status := &AlertEngineStatus{
		Status: "ok",
	}

	// Extract cached rules count
	if cachedRules, ok := stats["cached_rules"].(int); ok {
		status.CachedRules = cachedRules
	}

	// Extract rule cache last refresh
	if lastRefresh, ok := stats["rule_cache_last_refresh"].(string); ok {
		status.RuleCacheLastRefresh = lastRefresh

		// Check if cache is stale (> 5 minutes)
		if lastRefreshTime, err := time.Parse(time.RFC3339, lastRefresh); err == nil {
			if time.Since(lastRefreshTime) > 5*time.Minute {
				status.Status = "stale"
			}
		}
	}

	// Extract metric channel depth and capacity
	channelDepth := 0
	if depth, ok := stats["metric_channel_depth"].(int); ok {
		channelDepth = depth
		status.MetricChannelDepth = depth
	}
	if channelCapacity, ok := stats["metric_channel_capacity"].(int); ok {
		status.MetricChannelCapacity = channelCapacity

		// Check if channel is full
		if channelDepth >= channelCapacity {
			status.Status = "full"
		}
	}

	return status, nil
}

// CheckWebhookDelivery checks webhook delivery health based on recent logs
func (c *AlertSystemChecker) CheckWebhookDelivery(ctx context.Context) (*WebhookDeliveryStatus, error) {
	status := &WebhookDeliveryStatus{
		Status:     "nodata",
		SuccessRate: 0.0,
		TotalCount:  0,
		SuccessCount: 0,
	}

	// Query last 100 webhook logs
	// Note: For MVP, we'll use a simple count query
	// In production, you might want to add a GetRecentWebhookLogs method
	var totalCount, successCount int64

	// Get total count from last 100 records
	err := c.webhookLogsQuerier.CountRecentWebhookLogs(ctx, &totalCount, &successCount, 100)
	if err != nil {
		// If query fails, return nodata (fail-open)
		return status, nil
	}

	status.TotalCount = int(totalCount)
	status.SuccessCount = int(successCount)

	if totalCount == 0 {
		// No webhook logs yet
		return status, nil
	}

	// Calculate success rate
	status.SuccessRate = float64(successCount) / float64(totalCount) * 100.0

	// Determine health status
	if status.SuccessRate >= 95.0 {
		status.Status = "healthy"
	} else if status.SuccessRate >= 80.0 {
		status.Status = "degraded"
	} else {
		status.Status = "unhealthy"
	}

	return status, nil
}

// CheckAlertSuppression checks alert suppression service health
func (c *AlertSystemChecker) CheckAlertSuppression(ctx context.Context) (*AlertSuppressionStatus, error) {
	status := &AlertSuppressionStatus{
		Status: "ok",
	}

	// Get count of active suppressions (suppressed_until > now)
	activeCount, err := c.alertSuppressionsQuerier.CountActiveSuppressions(ctx)
	if err != nil {
		// Fail-open: return error status but don't fail health check
		status.Status = "error"
		status.ActiveSuppressionCount = 0
		return status, nil
	}

	status.ActiveSuppressionCount = activeCount
	return status, nil
}

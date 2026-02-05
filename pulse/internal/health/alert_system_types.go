package health

// AlertSystemStatus represents the overall alert system health status
type AlertSystemStatus struct {
	AlertEngine      *AlertEngineStatus      `json:"alert_engine,omitempty"`
	WebhookDelivery  *WebhookDeliveryStatus  `json:"webhook_delivery,omitempty"`
	AlertSuppression *AlertSuppressionStatus `json:"alert_suppression,omitempty"`
}

// AlertEngineStatus represents alert engine health status
type AlertEngineStatus struct {
	Status                 string `json:"status"` // ok, stale, full
	CachedRules            int    `json:"cached_rules"`
	RuleCacheLastRefresh   string `json:"rule_cache_last_refresh"`
	MetricChannelDepth     int    `json:"metric_channel_depth"`
	MetricChannelCapacity  int    `json:"metric_channel_capacity"`
}

// WebhookDeliveryStatus represents webhook delivery health status
type WebhookDeliveryStatus struct {
	Status       string  `json:"status"` // healthy, degraded, unhealthy, nodata
	SuccessRate  float64 `json:"success_rate"` // 0-100
	TotalCount   int     `json:"total_count"`
	SuccessCount int     `json:"success_count"`
}

// AlertSuppressionStatus represents alert suppression service status
type AlertSuppressionStatus struct {
	Status                 string `json:"status"` // ok, error
	ActiveSuppressionCount int64  `json:"active_suppression_count"`
}

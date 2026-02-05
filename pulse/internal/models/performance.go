package models

// PerformanceTarget defines performance targets for a metric
type PerformanceTarget struct {
	MetricName string  `json:"metric_name"`
	DisplayName string  `json:"display_name"`
	TargetP99   float64 `json:"target_p99"`
	TargetP95   float64 `json:"target_p95"`
	Unit        string  `json:"unit"`
}

// PerformanceMetric represents a single performance metric with current values and status
type PerformanceMetric struct {
	MetricName    string  `json:"metric_name"`
	DisplayName   string  `json:"display_name"`
	CurrentP99    float64 `json:"current_p99"`
	CurrentP95    float64 `json:"current_p95"`
	TargetP99     float64 `json:"target_p99"`
	TargetP95     float64 `json:"target_p95"`
	Unit          string  `json:"unit"`
	Status        string  `json:"status"`        // "healthy" or "unhealthy"
	Anomaly       string  `json:"anomaly,omitempty"`
}

// TrendDataPoint represents a single data point in the trend chart
type TrendDataPoint struct {
	Timestamp string  `json:"timestamp"`
	P99       float64 `json:"p99"`
	P95       float64 `json:"p95"`
}

// MetricTrendData contains trend data for a specific metric
type MetricTrendData struct {
	MetricName string         `json:"metric_name"`
	DataPoints []TrendDataPoint `json:"data_points"`
}

// Anomaly represents a performance anomaly
type Anomaly struct {
	MetricName string `json:"metric_name"`
	Severity   string `json:"severity"`   // "P0", "P1"
	Message    string `json:"message"`
}

// PerformanceSummary contains summary statistics
type PerformanceSummary struct {
	TotalRequests   int64   `json:"total_requests"`
	AvgResponseTime float64 `json:"avg_response_time"`
	MaxResponseTime float64 `json:"max_response_time"`
}

// PerformanceDataResponse is the response structure for performance data API
type PerformanceDataResponse struct {
	Metrics      []PerformanceMetric `json:"metrics"`
	TrendData    []MetricTrendData   `json:"trend_data"`
	SystemHealth string             `json:"system_health"`
	Anomalies    []Anomaly          `json:"anomalies"`
	Summary      PerformanceSummary `json:"summary"`
}

// Performance targets defined in the story (FR21)
var PerformanceTargets = []PerformanceTarget{
	{
		MetricName: "dashboard_load_time",
		DisplayName: "仪表盘加载时间",
		TargetP99:   3000, // 3 seconds
		TargetP95:   2000, // 2 seconds
		Unit:        "ms",
	},
	{
		MetricName: "api_response_time",
		DisplayName: "API 响应时间",
		TargetP99:   500, // 500ms
		TargetP95:   200, // 200ms
		Unit:        "ms",
	},
	{
		MetricName: "data_query_time",
		DisplayName: "数据查询时间",
		TargetP99:   300, // 300ms
		TargetP95:   200, // 200ms
		Unit:        "ms",
	},
}

// GetTarget retrieves performance target for a given metric name
func GetTarget(metricName string) *PerformanceTarget {
	for _, target := range PerformanceTargets {
		if target.MetricName == metricName {
			return &target
		}
	}
	return nil
}

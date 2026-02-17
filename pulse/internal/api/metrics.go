package api

import (
	"context"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/whg517/node-pulse/pulse/pkg/metrics"
)

// Prometheus metrics for FR-4.2.3 System Self-Observability
var (
	// System metrics
	pulseMemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pulse_server_memory_usage_bytes",
		Help: "Current memory usage in bytes",
	})

	pulseCPUUsagePercent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pulse_server_cpu_usage_percent",
		Help: "Current CPU usage percentage",
	})

	pulseGoroutines = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pulse_server_goroutines",
		Help: "Current number of goroutines",
	})

	pulseHeapAllocBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pulse_server_heap_alloc_bytes",
		Help: "Current heap allocation in bytes",
	})

	// API metrics
	pulseAPIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulse_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint", "status_code"},
	)

	pulseAPIResponseTimeSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulse_api_response_time_seconds",
			Help:    "API response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	pulseAPIActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pulse_api_active_connections",
		Help: "Current number of active connections",
	})

	// Webhook metrics
	pulseWebhookQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pulse_webhook_queue_depth",
		Help: "Current webhook queue depth",
	})

	pulseWebhookDeliverySuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pulse_webhook_delivery_success_total",
		Help: "Total number of successful webhook deliveries",
	})

	pulseWebhookDeliveryFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pulse_webhook_delivery_failed_total",
		Help: "Total number of failed webhook deliveries",
	})

	// Beacon metrics
	pulseBeaconsConnected = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pulse_beacons_connected",
		Help: "Current number of connected beacons",
	})

	pulseBeaconsDisconnectedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pulse_beacons_disconnected_total",
		Help: "Total number of beacon disconnections",
	})

	// Compression metrics (FR-4.1.5)
	pulseCompressionCorruptionTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pulse_compression_corruption_total",
		Help: "Total number of compression corruption errors",
	})
)

// SystemMetricsCollector manages system metrics collection
type SystemMetricsCollector struct {
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	started  bool
	interval time.Duration
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector(interval time.Duration) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		interval: interval,
	}
}

// Start begins periodic system metrics collection
func (c *SystemMetricsCollector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return
	}

	c.started = true

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel

	c.wg.Add(1)
	go c.collectionLoop()
}

// Stop stops the system metrics collector
func (c *SystemMetricsCollector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return
	}

	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	c.started = false
}

// collectionLoop periodically collects system metrics
func (c *SystemMetricsCollector) collectionLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect immediately on start
	c.collectSystemMetrics()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collectSystemMetrics()
		}
	}
}

// collectSystemMetrics collects current system metrics
func (c *SystemMetricsCollector) collectSystemMetrics() {
	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pulseMemoryUsageBytes.Set(float64(m.Sys))
	pulseHeapAllocBytes.Set(float64(m.HeapAlloc))

	// Goroutines
	pulseGoroutines.Set(float64(runtime.NumGoroutine()))

	// CPU usage is approximated by GC CPU fraction
	// Note: For more accurate CPU usage, use golang.org/x/sys/cpu or external tools
	// This is a simplified approximation
	cpuUsage := float64(m.GCCPUFraction) * 100
	pulseCPUUsagePercent.Set(cpuUsage)
}

// RecordAPIRequest records an API request metric
func RecordAPIRequest(endpoint string, statusCode int, duration time.Duration) {
	statusCodeStr := strconv.Itoa(statusCode)
	pulseAPIRequestsTotal.WithLabelValues(endpoint, statusCodeStr).Inc()
	pulseAPIResponseTimeSeconds.WithLabelValues(endpoint).Observe(duration.Seconds())
}

// IncrementActiveConnections increments the active connections counter
func IncrementActiveConnections() {
	pulseAPIActiveConnections.Inc()
}

// DecrementActiveConnections decrements the active connections counter
func DecrementActiveConnections() {
	pulseAPIActiveConnections.Dec()
}

// SetWebhookQueueDepth sets the current webhook queue depth
func SetWebhookQueueDepth(depth float64) {
	pulseWebhookQueueDepth.Set(depth)
}

// RecordWebhookSuccess records a successful webhook delivery
func RecordWebhookSuccess() {
	pulseWebhookDeliverySuccessTotal.Inc()
}

// RecordWebhookFailure records a failed webhook delivery
func RecordWebhookFailure() {
	pulseWebhookDeliveryFailedTotal.Inc()
}

// SetBeaconsConnected sets the current number of connected beacons
func SetBeaconsConnected(count float64) {
	pulseBeaconsConnected.Set(count)
}

// RecordBeaconDisconnection records a beacon disconnection
func RecordBeaconDisconnection() {
	pulseBeaconsDisconnectedTotal.Inc()
}

// RecordCompressionCorruption records a compression corruption error
func RecordCompressionCorruption() {
	pulseCompressionCorruptionTotal.Inc()
}

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
		TimeWindow  string  `json:"time_window"`
		Endpoint    string  `json:"endpoint"`
		MetricType  string  `json:"metric_type"`
		Count       int64   `json:"count"`
		AvgDuration float64 `json:"avg_duration_ms"`
		P95         float64 `json:"p95_ms"`
		P99         float64 `json:"p99_ms"`
		SuccessRate float64 `json:"success_rate"`
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

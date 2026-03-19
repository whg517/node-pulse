package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	// DefaultBufferCapacity is the default capacity of the ring buffer (24 hours at 1 min intervals)
	DefaultBufferCapacity = 1440
	// DefaultAggregationInterval is the default interval for aggregating metrics
	DefaultAggregationInterval = time.Minute
	// DefaultRetentionPeriod is the default retention period for raw metrics
	DefaultRetentionPeriod = 24 * time.Hour
)

// AlertThresholds defines thresholds for performance alerts
type AlertThresholds struct {
	APIResponseP99      time.Duration
	DashboardLoadP99    time.Duration
	DatabaseQueryP99    time.Duration
	ErrorRate           float64
}

// DefaultAlertThresholds returns default alert thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		APIResponseP99:   500 * time.Millisecond,
		DashboardLoadP99: 1000 * time.Millisecond,
		DatabaseQueryP99: 100 * time.Millisecond,
		ErrorRate:        0.01, // 1%
	}
}

// Collector collects and aggregates performance metrics
type Collector struct {
	mu                  sync.RWMutex
	buffer              *RingBuffer
	aggregated          []AggregatedMetrics
	capacity            int
	aggregationInterval time.Duration
	retentionPeriod     time.Duration
	alertThresholds     AlertThresholds
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
	started             bool
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return NewCollectorWithConfig(DefaultBufferCapacity, DefaultAggregationInterval, DefaultRetentionPeriod)
}

// NewCollectorWithConfig creates a new metrics collector with custom configuration
func NewCollectorWithConfig(capacity int, aggregationInterval, retentionPeriod time.Duration) *Collector {
	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		buffer:              NewRingBuffer(capacity),
		aggregated:          make([]AggregatedMetrics, 0),
		capacity:            capacity,
		aggregationInterval: aggregationInterval,
		retentionPeriod:     retentionPeriod,
		alertThresholds:     DefaultAlertThresholds(),
		ctx:                 ctx,
		cancel:              cancel,
		started:             false,
	}
}

// Start begins the metrics collection and aggregation process
func (c *Collector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return
	}

	c.started = true

	// Start aggregation goroutine
	c.wg.Add(1)
	go c.aggregationLoop()

	slog.Info("Metrics collector started",
		"component", "metrics",
		"capacity", c.capacity,
		"aggregation_interval", c.aggregationInterval,
		"retention_period", c.retentionPeriod,
	)
}

// Stop gracefully stops the metrics collector
func (c *Collector) Stop() {
	c.mu.Lock()
	if !c.started {
		c.mu.Unlock()
		return
	}
	c.cancel()
	c.started = false
	c.mu.Unlock()

	// Wait outside the lock so in-progress aggregation goroutines can acquire
	// the mutex to complete their current work before exiting.
	c.wg.Wait()

	slog.Info("Metrics collector stopped", "component", "metrics")
}

// RecordMetric records a single metric
func (c *Collector) RecordMetric(record MetricRecord) {
	c.buffer.Push(record)
}

// RecordAPIRequest records an API request metric
func (c *Collector) RecordAPIRequest(endpoint, method string, duration time.Duration, status int) {
	c.RecordMetric(MetricRecord{
		Timestamp:  time.Now(),
		Endpoint:   endpoint,
		Method:     method,
		Duration:   duration,
		Status:     status,
		MetricType: MetricTypeAPI,
	})
}

// RecordDashboardLoad records a dashboard load metric
func (c *Collector) RecordDashboardLoad(endpoint, method string, duration time.Duration, status int) {
	c.RecordMetric(MetricRecord{
		Timestamp:  time.Now(),
		Endpoint:   endpoint,
		Method:     method,
		Duration:   duration,
		Status:     status,
		MetricType: MetricTypeDashboard,
	})
}

// RecordDatabaseQuery records a database query metric
func (c *Collector) RecordDatabaseQuery(queryType string, duration time.Duration, success bool) {
	status := 200
	if !success {
		status = 500
	}

	c.RecordMetric(MetricRecord{
		Timestamp:  time.Now(),
		Endpoint:   queryType,
		Method:     "DB",
		Duration:   duration,
		Status:     status,
		MetricType: MetricTypeDatabase,
	})
}

// GetMetrics retrieves metrics based on the provided filter
func (c *Collector) GetMetrics(filter MetricsFilter) ([]AggregatedMetrics, MetricsSummary) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get filtered raw metrics
	records := c.buffer.GetFiltered(filter)

	// Aggregate metrics
	aggregationInterval := filter.Aggregation
	if aggregationInterval == 0 {
		aggregationInterval = c.aggregationInterval
	}

	aggregated := AggregateMetrics(records, aggregationInterval)

	// Calculate summary
	summary := CalculateSummary(aggregated)

	return aggregated, summary
}

// GetAggregatedMetrics returns the pre-aggregated metrics
func (c *Collector) GetAggregatedMetrics() []AggregatedMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]AggregatedMetrics, len(c.aggregated))
	copy(result, c.aggregated)
	return result
}

// aggregationLoop runs periodic aggregation of metrics
func (c *Collector) aggregationLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.aggregationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.aggregateMetrics()
			c.cleanupOldMetrics()
			c.checkAlertThresholds()
		}
	}
}

// aggregateMetrics aggregates raw metrics and stores them
func (c *Collector) aggregateMetrics() {
	// Get all raw metrics
	records := c.buffer.GetAll()

	if len(records) == 0 {
		return
	}

	// Aggregate by minute
	aggregated := AggregateMetrics(records, time.Minute)

	c.mu.Lock()
	// Append new aggregated metrics
	c.aggregated = append(c.aggregated, aggregated...)
	c.mu.Unlock()

	slog.Debug("Aggregated metrics",
		"component", "metrics",
		"raw_count", len(records),
		"window_count", len(aggregated),
	)
}

// cleanupOldMetrics removes raw metrics older than the retention period
func (c *Collector) cleanupOldMetrics() {
	cutoff := time.Now().Add(-c.retentionPeriod)

	// Note: RingBuffer automatically overwrites old data, so this is mainly
	// for cleanup of aggregated metrics if needed
	c.mu.Lock()
	defer c.mu.Unlock()

	var filtered []AggregatedMetrics
	for _, agg := range c.aggregated {
		if agg.TimeWindow.After(cutoff) {
			filtered = append(filtered, agg)
		}
	}

	if len(filtered) < len(c.aggregated) {
		removed := len(c.aggregated) - len(filtered)
		c.aggregated = filtered
		slog.Debug("Cleaned up old aggregated metrics", "component", "metrics", "removed", removed)
	}
}

// checkAlertThresholds checks if any performance metrics exceed alert thresholds
func (c *Collector) checkAlertThresholds() {
	records := c.buffer.GetAll()

	if len(records) == 0 {
		return
	}

	// Group by metric type
	byType := make(map[MetricType][]time.Duration)
	errorCounts := make(map[MetricType]int)

	for _, record := range records {
		byType[record.MetricType] = append(byType[record.MetricType], record.Duration)
		if record.Status >= 400 {
			errorCounts[record.MetricType]++
		}
	}

	// Check thresholds
	for metricType, durations := range byType {
		if len(durations) == 0 {
			continue
		}

		_, _, p99 := CalculatePercentiles(durations)
		errorRate := float64(errorCounts[metricType]) / float64(len(durations))

		var threshold time.Duration
		var metricName string

		switch metricType {
		case MetricTypeAPI:
			threshold = c.alertThresholds.APIResponseP99
			metricName = "API Response"
		case MetricTypeDashboard:
			threshold = c.alertThresholds.DashboardLoadP99
			metricName = "Dashboard Load"
		case MetricTypeDatabase:
			threshold = c.alertThresholds.DatabaseQueryP99
			metricName = "Database Query"
		}

		if p99 > threshold {
			slog.Warn("Performance alert: P99 exceeds threshold",
				"component", "metrics",
				"metric", metricName,
				"p99", p99,
				"threshold", threshold,
			)
		}

		if errorRate > c.alertThresholds.ErrorRate {
			slog.Warn("Performance alert: error rate exceeds threshold",
				"component", "metrics",
				"metric", metricName,
				"error_rate_pct", errorRate*100,
				"threshold_pct", c.alertThresholds.ErrorRate*100,
			)
		}
	}
}

// SetAlertThresholds sets custom alert thresholds
func (c *Collector) SetAlertThresholds(thresholds AlertThresholds) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.alertThresholds = thresholds
	slog.Debug("Alert thresholds updated", "component", "metrics")
}

// GetStats returns collector statistics
func (c *Collector) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"started":             c.started,
		"buffer_size":         c.buffer.Size(),
		"buffer_capacity":     c.capacity,
		"aggregated_metrics":  len(c.aggregated),
		"aggregation_interval": c.aggregationInterval.String(),
		"retention_period":     c.retentionPeriod.String(),
	}
}

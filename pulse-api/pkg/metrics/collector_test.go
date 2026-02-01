package metrics

import (
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	if collector == nil {
		t.Fatal("Expected non-nil collector")
	}

	if collector.started {
		t.Error("Expected collector to be not started")
	}

	if collector.capacity != DefaultBufferCapacity {
		t.Errorf("Expected capacity %d, got %d", DefaultBufferCapacity, collector.capacity)
	}
}

func TestNewCollectorWithConfig(t *testing.T) {
	capacity := 100
	interval := 30 * time.Second
	retention := 12 * time.Hour

	collector := NewCollectorWithConfig(capacity, interval, retention)

	if collector == nil {
		t.Fatal("Expected non-nil collector")
	}

	if collector.capacity != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, collector.capacity)
	}

	if collector.aggregationInterval != interval {
		t.Errorf("Expected aggregation interval %v, got %v", interval, collector.aggregationInterval)
	}

	if collector.retentionPeriod != retention {
		t.Errorf("Expected retention period %v, got %v", retention, collector.retentionPeriod)
	}
}

func TestCollectorStartStop(t *testing.T) {
	collector := NewCollector()

	collector.Start()

	if !collector.started {
		t.Error("Expected collector to be started")
	}

	// Start again should be idempotent
	collector.Start()

	if !collector.started {
		t.Error("Expected collector to still be started")
	}

	collector.Stop()

	if collector.started {
		t.Error("Expected collector to be stopped")
	}
}

func TestRecordMetric(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	record := MetricRecord{
		Timestamp:  time.Now(),
		Endpoint:   "/api/test",
		Method:     "GET",
		Duration:   100 * time.Millisecond,
		Status:     200,
		MetricType: MetricTypeAPI,
	}

	collector.RecordMetric(record)

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	stats := collector.GetStats()
	bufferSize, ok := stats["buffer_size"].(int)
	if !ok {
		t.Fatal("Expected buffer_size to be int")
	}

	if bufferSize != 1 {
		t.Errorf("Expected buffer size 1, got %d", bufferSize)
	}
}

func TestRecordAPIRequest(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	collector.RecordAPIRequest("/api/test", "GET", 100*time.Millisecond, 200)

	time.Sleep(100 * time.Millisecond)

	stats := collector.GetStats()
	bufferSize, _ := stats["buffer_size"].(int)

	if bufferSize != 1 {
		t.Errorf("Expected buffer size 1, got %d", bufferSize)
	}
}

func TestRecordDashboardLoad(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	collector.RecordDashboardLoad("/api/v1/data/metrics", "GET", 200*time.Millisecond, 200)

	time.Sleep(100 * time.Millisecond)

	stats := collector.GetStats()
	bufferSize, _ := stats["buffer_size"].(int)

	if bufferSize != 1 {
		t.Errorf("Expected buffer size 1, got %d", bufferSize)
	}
}

func TestRecordDatabaseQuery(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	collector.RecordDatabaseQuery("SELECT", 50*time.Millisecond, true)

	time.Sleep(100 * time.Millisecond)

	stats := collector.GetStats()
	bufferSize, _ := stats["buffer_size"].(int)

	if bufferSize != 1 {
		t.Errorf("Expected buffer size 1, got %d", bufferSize)
	}
}

func TestGetMetrics(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	// Record some metrics
	now := time.Now()
	for i := 0; i < 10; i++ {
		collector.RecordAPIRequest("/api/test", "GET", time.Duration(i)*10*time.Millisecond, 200)
	}

	time.Sleep(200 * time.Millisecond)

	// Get all metrics
	filter := MetricsFilter{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
	}

	aggregated, summary := collector.GetMetrics(filter)

	if len(aggregated) == 0 {
		t.Error("Expected at least one aggregated metric")
	}

	if summary.TotalRequests != 10 {
		t.Errorf("Expected total requests 10, got %d", summary.TotalRequests)
	}
}

func TestGetMetricsWithFilter(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	// Record different types of metrics
	collector.RecordAPIRequest("/api/test1", "GET", 100*time.Millisecond, 200)
	collector.RecordDashboardLoad("/api/v1/data/metrics", "GET", 200*time.Millisecond, 200)
	collector.RecordDatabaseQuery("SELECT", 50*time.Millisecond, true)

	time.Sleep(200 * time.Millisecond)

	// Filter by API type only
	metricType := MetricTypeAPI
	filter := MetricsFilter{
		MetricType: &metricType,
	}

	aggregated, _ := collector.GetMetrics(filter)

	// Should only have API metrics
	for _, agg := range aggregated {
		if agg.MetricType != "api" {
			t.Errorf("Expected metric type 'api', got '%s'", agg.MetricType)
		}
	}
}

func TestDefaultAlertThresholds(t *testing.T) {
	thresholds := DefaultAlertThresholds()

	if thresholds.APIResponseP99 != 500*time.Millisecond {
		t.Errorf("Expected API P99 threshold 500ms, got %v", thresholds.APIResponseP99)
	}

	if thresholds.DashboardLoadP99 != 1000*time.Millisecond {
		t.Errorf("Expected Dashboard P99 threshold 1000ms, got %v", thresholds.DashboardLoadP99)
	}

	if thresholds.DatabaseQueryP99 != 100*time.Millisecond {
		t.Errorf("Expected Database P99 threshold 100ms, got %v", thresholds.DatabaseQueryP99)
	}

	if thresholds.ErrorRate != 0.01 {
		t.Errorf("Expected error rate threshold 0.01, got %f", thresholds.ErrorRate)
	}
}

func TestSetAlertThresholds(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	newThresholds := AlertThresholds{
		APIResponseP99:   1000 * time.Millisecond,
		DashboardLoadP99: 2000 * time.Millisecond,
		DatabaseQueryP99: 200 * time.Millisecond,
		ErrorRate:        0.02,
	}

	collector.SetAlertThresholds(newThresholds)

	time.Sleep(200 * time.Millisecond)

	// Verify thresholds were set (we can't directly access them, but the call shouldn't panic)
}

func TestGetStats(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	stats := collector.GetStats()

	started, ok := stats["started"].(bool)
	if !ok {
		t.Fatal("Expected started to be bool")
	}

	if !started {
		t.Error("Expected started to be true")
	}

	bufferCapacity, ok := stats["buffer_capacity"].(int)
	if !ok {
		t.Fatal("Expected buffer_capacity to be int")
	}

	if bufferCapacity != DefaultBufferCapacity {
		t.Errorf("Expected buffer_capacity %d, got %d", DefaultBufferCapacity, bufferCapacity)
	}

	aggregationInterval, ok := stats["aggregation_interval"].(string)
	if !ok {
		t.Fatal("Expected aggregation_interval to be string")
	}

	if aggregationInterval != "1m0s" {
		t.Errorf("Expected aggregation_interval '1m0s', got '%s'", aggregationInterval)
	}

	retentionPeriod, ok := stats["retention_period"].(string)
	if !ok {
		t.Fatal("Expected retention_period to be string")
	}

	if retentionPeriod != "24h0m0s" {
		t.Errorf("Expected retention_period '24h0m0s', got '%s'", retentionPeriod)
	}
}

func TestCollectorBufferOverflow(t *testing.T) {
	capacity := 10
	collector := NewCollectorWithConfig(capacity, time.Minute, time.Hour)
	collector.Start()
	defer collector.Stop()

	// Record more metrics than capacity
	for i := 0; i < 20; i++ {
		collector.RecordAPIRequest("/api/test", "GET", time.Duration(i)*time.Millisecond, 200)
	}

	time.Sleep(200 * time.Millisecond)

	stats := collector.GetStats()
	bufferSize, _ := stats["buffer_size"].(int)

	// Buffer should be at capacity
	if bufferSize != capacity {
		t.Errorf("Expected buffer size %d, got %d", capacity, bufferSize)
	}
}

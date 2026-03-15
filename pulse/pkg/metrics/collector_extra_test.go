package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAggregatedMetrics_Empty(t *testing.T) {
	collector := NewCollector()

	result := collector.GetAggregatedMetrics()
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestGetAggregatedMetrics_WithData(t *testing.T) {
	collector := NewCollectorWithConfig(100, 50*time.Millisecond, time.Hour)
	collector.Start()
	defer collector.Stop()

	// Record some metrics
	for i := 0; i < 5; i++ {
		collector.RecordAPIRequest("/api/test", "GET", 100*time.Millisecond, 200)
	}

	// Wait for aggregation
	time.Sleep(200 * time.Millisecond)

	// The aggregated slice is populated after aggregation runs
	// GetAggregatedMetrics should return a copy
	result := collector.GetAggregatedMetrics()
	assert.NotNil(t, result)
}

func TestAggregateMetrics_Direct(t *testing.T) {
	collector := NewCollectorWithConfig(100, time.Hour, time.Hour)

	// Manually add records to buffer
	now := time.Now()
	for i := 0; i < 10; i++ {
		collector.RecordAPIRequest("/api/test", "GET",
			time.Duration(i+1)*10*time.Millisecond, 200)
	}

	// Call aggregateMetrics directly
	assert.NotPanics(t, func() {
		collector.aggregateMetrics()
	})

	// After aggregation, aggregated slice should have data
	result := collector.GetAggregatedMetrics()
	assert.NotEmpty(t, result)
	_ = now
}

func TestAggregateMetrics_EmptyBuffer(t *testing.T) {
	collector := NewCollector()

	// Should not panic when buffer is empty
	assert.NotPanics(t, func() {
		collector.aggregateMetrics()
	})

	result := collector.GetAggregatedMetrics()
	assert.Empty(t, result)
}

func TestCleanupOldMetrics_RemovesStaleData(t *testing.T) {
	// Use a very short retention period
	collector := NewCollectorWithConfig(100, time.Hour, 1*time.Millisecond)

	// Manually add stale aggregated data
	staleTime := time.Now().Add(-10 * time.Second)
	freshTime := time.Now().Add(10 * time.Second)

	collector.mu.Lock()
	collector.aggregated = []AggregatedMetrics{
		{TimeWindow: staleTime},
		{TimeWindow: freshTime},
	}
	collector.mu.Unlock()

	// Run cleanup
	assert.NotPanics(t, func() {
		collector.cleanupOldMetrics()
	})

	// Stale metrics should be removed
	result := collector.GetAggregatedMetrics()
	// Only the fresh metric should remain
	assert.Len(t, result, 1)
	assert.Equal(t, freshTime, result[0].TimeWindow)
}

func TestCleanupOldMetrics_AllFresh(t *testing.T) {
	collector := NewCollectorWithConfig(100, time.Hour, 24*time.Hour)

	freshTime := time.Now()
	collector.mu.Lock()
	collector.aggregated = []AggregatedMetrics{
		{TimeWindow: freshTime},
		{TimeWindow: freshTime.Add(time.Minute)},
	}
	collector.mu.Unlock()

	collector.cleanupOldMetrics()

	result := collector.GetAggregatedMetrics()
	assert.Len(t, result, 2)
}

func TestCheckAlertThresholds_WithAlerts(t *testing.T) {
	collector := NewCollectorWithConfig(100, time.Hour, time.Hour)

	// Set very low thresholds to trigger alerts
	collector.SetAlertThresholds(AlertThresholds{
		APIResponseP99:   1 * time.Millisecond,
		DashboardLoadP99: 1 * time.Millisecond,
		DatabaseQueryP99: 1 * time.Millisecond,
		ErrorRate:        0.001, // 0.1%
	})

	// Add API metrics that will exceed threshold
	collector.RecordAPIRequest("/api/test", "GET", 100*time.Millisecond, 200)
	collector.RecordAPIRequest("/api/test", "GET", 200*time.Millisecond, 500) // error
	collector.RecordAPIRequest("/api/test", "GET", 300*time.Millisecond, 200)

	// Add dashboard metrics
	collector.RecordDashboardLoad("/dashboard", "GET", 500*time.Millisecond, 200)

	// Add database metrics
	collector.RecordDatabaseQuery("SELECT", 50*time.Millisecond, true)
	collector.RecordDatabaseQuery("SELECT", 50*time.Millisecond, false) // failure

	// checkAlertThresholds should not panic
	assert.NotPanics(t, func() {
		collector.checkAlertThresholds()
	})
}

func TestCheckAlertThresholds_EmptyBuffer(t *testing.T) {
	collector := NewCollector()

	// Should not panic with empty buffer
	assert.NotPanics(t, func() {
		collector.checkAlertThresholds()
	})
}

func TestRecordDatabaseQuery_Failure(t *testing.T) {
	collector := NewCollector()

	// Record a failed query
	assert.NotPanics(t, func() {
		collector.RecordDatabaseQuery("INSERT", 50*time.Millisecond, false)
	})

	stats := collector.GetStats()
	assert.Equal(t, 1, stats["buffer_size"])
}


func TestCollector_StopNotStarted(t *testing.T) {
	collector := NewCollector()

	// Stop without starting should not panic
	assert.NotPanics(t, func() {
		collector.Stop()
	})
}

func TestCollector_StartTwice(t *testing.T) {
	collector := NewCollector()
	collector.Start()
	defer collector.Stop()

	// Starting again should not panic or create duplicate goroutines
	assert.NotPanics(t, func() {
		collector.Start()
	})

	stats := collector.GetStats()
	assert.True(t, stats["started"].(bool))
}

func TestGetMetrics_WithFilter(t *testing.T) {
	collector := NewCollector()

	now := time.Now()
	collector.RecordAPIRequest("/api/nodes", "GET", 100*time.Millisecond, 200)
	collector.RecordAPIRequest("/api/metrics", "GET", 50*time.Millisecond, 200)
	collector.RecordDashboardLoad("/dashboard", "GET", 200*time.Millisecond, 200)

	filter := MetricsFilter{
		StartTime:   now.Add(-time.Minute),
		EndTime:     now.Add(time.Minute),
		Aggregation: time.Minute,
	}

	aggregated, summary := collector.GetMetrics(filter)
	require.NotNil(t, aggregated)
	require.NotNil(t, summary)
}

func TestAggregationLoop_Runs(t *testing.T) {
	collector := NewCollectorWithConfig(100, 50*time.Millisecond, time.Hour)
	collector.Start()

	// Add some metrics
	collector.RecordAPIRequest("/api/test", "GET", 100*time.Millisecond, 200)

	// Wait for at least one aggregation cycle
	time.Sleep(200 * time.Millisecond)

	collector.Stop()

	// Stats should be available
	stats := collector.GetStats()
	assert.NotNil(t, stats)
}

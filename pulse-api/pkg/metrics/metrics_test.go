package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestRingBuffer(t *testing.T) {
	t.Run("Push and size", func(t *testing.T) {
		rb := NewRingBuffer(10)

		for i := 0; i < 5; i++ {
			rb.Push(MetricRecord{
				Timestamp: time.Now(),
				Duration:  time.Duration(i) * time.Millisecond,
			})
		}

		if rb.Size() != 5 {
			t.Errorf("Expected size 5, got %d", rb.Size())
		}
	})

	t.Run("Overwrite when full", func(t *testing.T) {
		rb := NewRingBuffer(5)

		for i := 0; i < 10; i++ {
			rb.Push(MetricRecord{
				Timestamp: time.Now(),
				Duration:  time.Duration(i) * time.Millisecond,
			})
		}

		if rb.Size() != 5 {
			t.Errorf("Expected size 5, got %d", rb.Size())
		}

		// Check that oldest entries were overwritten
		records := rb.GetAll()
		if len(records) != 5 {
			t.Errorf("Expected 5 records, got %d", len(records))
		}

		// The oldest record should be 5 (first 5 were overwritten)
		if records[0].Duration != 5*time.Millisecond {
			t.Errorf("Expected oldest duration 5ms, got %v", records[0].Duration)
		}
	})

	t.Run("GetFiltered", func(t *testing.T) {
		rb := NewRingBuffer(100)

		now := time.Now()
		for i := 0; i < 10; i++ {
			rb.Push(MetricRecord{
				Timestamp:  now.Add(time.Duration(i) * time.Minute),
				Endpoint:   "/api/test",
				Duration:   time.Duration(i) * time.Millisecond,
				MetricType: MetricTypeAPI,
			})
		}

		// Filter by time range
		filter := MetricsFilter{
			StartTime: now.Add(5 * time.Minute),
			EndTime:   now.Add(10 * time.Minute),
		}

		records := rb.GetFiltered(filter)
		if len(records) != 5 {
			t.Errorf("Expected 5 records after filtering, got %d", len(records))
		}

		// Filter by metric type
		filter = MetricsFilter{
			MetricType: func() *MetricType { mt := MetricTypeAPI; return &mt }(),
		}

		records = rb.GetFiltered(filter)
		if len(records) != 10 {
			t.Errorf("Expected 10 API records, got %d", len(records))
		}
	})

	t.Run("Clear", func(t *testing.T) {
		rb := NewRingBuffer(10)

		for i := 0; i < 5; i++ {
			rb.Push(MetricRecord{
				Timestamp: time.Now(),
				Duration:  time.Duration(i) * time.Millisecond,
			})
		}

		rb.Clear()

		if rb.Size() != 0 {
			t.Errorf("Expected size 0 after clear, got %d", rb.Size())
		}
	})
}

func TestCalculatePercentiles(t *testing.T) {
	t.Run("Basic percentiles", func(t *testing.T) {
		durations := []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			30 * time.Millisecond,
			40 * time.Millisecond,
			50 * time.Millisecond,
			60 * time.Millisecond,
			70 * time.Millisecond,
			80 * time.Millisecond,
			90 * time.Millisecond,
			100 * time.Millisecond,
		}

		p50, p95, p99 := CalculatePercentiles(durations)

		// P50 should be around 50ms (median)
		if p50 < 45*time.Millisecond || p50 > 55*time.Millisecond {
			t.Errorf("P50 out of expected range: %v", p50)
		}

		// P95 should be around 95ms
		if p95 < 90*time.Millisecond || p95 > 100*time.Millisecond {
			t.Errorf("P95 out of expected range: %v", p95)
		}

		// P99 should be 100ms (max)
		if p99 != 100*time.Millisecond {
			t.Errorf("P99 should be 100ms, got %v", p99)
		}
	})

	t.Run("Empty slice", func(t *testing.T) {
		durations := []time.Duration{}
		p50, p95, p99 := CalculatePercentiles(durations)

		if p50 != 0 || p95 != 0 || p99 != 0 {
			t.Errorf("Expected 0 for all percentiles with empty slice, got %v, %v, %v", p50, p95, p99)
		}
	})

	t.Run("Single value", func(t *testing.T) {
		durations := []time.Duration{50 * time.Millisecond}
		p50, p95, p99 := CalculatePercentiles(durations)

		if p50 != 50*time.Millisecond || p95 != 50*time.Millisecond || p99 != 50*time.Millisecond {
			t.Errorf("Expected all percentiles to be 50ms, got %v, %v, %v", p50, p95, p99)
		}
	})
}

func TestAggregateMetrics(t *testing.T) {
	t.Run("Basic aggregation", func(t *testing.T) {
		now := time.Now()
		records := []MetricRecord{
			{
				Timestamp:  now,
				Endpoint:   "/api/test",
				Method:     "GET",
				Duration:   100 * time.Millisecond,
				Status:     200,
				MetricType: MetricTypeAPI,
			},
			{
				Timestamp:  now.Add(10 * time.Second),
				Endpoint:   "/api/test",
				Method:     "GET",
				Duration:   150 * time.Millisecond,
				Status:     200,
				MetricType: MetricTypeAPI,
			},
			{
				Timestamp:  now.Add(20 * time.Second),
				Endpoint:   "/api/test",
				Method:     "GET",
				Duration:   200 * time.Millisecond,
				Status:     200,
				MetricType: MetricTypeAPI,
			},
		}

		aggregated := AggregateMetrics(records, time.Minute)

		if len(aggregated) == 0 {
			t.Fatal("Expected at least one aggregated metric")
		}

		agg := aggregated[0]
		if agg.Count != 3 {
			t.Errorf("Expected count 3, got %d", agg.Count)
		}

		if agg.AvgDuration < 149 || agg.AvgDuration > 151 {
			t.Errorf("Expected avg duration around 150ms, got %f", agg.AvgDuration)
		}

		if agg.MinDuration != 100 {
			t.Errorf("Expected min duration 100ms, got %f", agg.MinDuration)
		}

		if agg.MaxDuration != 200 {
			t.Errorf("Expected max duration 200ms, got %f", agg.MaxDuration)
		}
	})

	t.Run("Empty records", func(t *testing.T) {
		records := []MetricRecord{}
		aggregated := AggregateMetrics(records, time.Minute)

		if len(aggregated) != 0 {
			t.Errorf("Expected empty aggregation for empty records, got %d", len(aggregated))
		}
	})

	t.Run("Success rate calculation", func(t *testing.T) {
		now := time.Now()
		records := []MetricRecord{
			{
				Timestamp:  now,
				Endpoint:   "/api/test",
				Duration:   100 * time.Millisecond,
				Status:     200,
				MetricType: MetricTypeAPI,
			},
			{
				Timestamp:  now,
				Endpoint:   "/api/test",
				Duration:   100 * time.Millisecond,
				Status:     200,
				MetricType: MetricTypeAPI,
			},
			{
				Timestamp:  now,
				Endpoint:   "/api/test",
				Duration:   100 * time.Millisecond,
				Status:     500,
				MetricType: MetricTypeAPI,
			},
		}

		aggregated := AggregateMetrics(records, time.Minute)

		if len(aggregated) == 0 {
			t.Fatal("Expected at least one aggregated metric")
		}

		agg := aggregated[0]
		expectedSuccessRate := 2.0 / 3.0
		if agg.SuccessRate != expectedSuccessRate {
			t.Errorf("Expected success rate %f, got %f", expectedSuccessRate, agg.SuccessRate)
		}
	})
}

func TestCalculateSummary(t *testing.T) {
	t.Run("Basic summary", func(t *testing.T) {
		aggregated := []AggregatedMetrics{
			{
				Count:        100,
				AvgDuration:  50.0,
				P95:          90.0,
				P99:          95.0,
				SuccessRate:  0.99,
			},
			{
				Count:        200,
				AvgDuration:  60.0,
				P95:          100.0,
				P99:          110.0,
				SuccessRate:  0.98,
			},
		}

		summary := CalculateSummary(aggregated)

		if summary.TotalRequests != 300 {
			t.Errorf("Expected total requests 300, got %d", summary.TotalRequests)
		}

		// Overall avg should be (100*50 + 200*60) / 300 = 17000/300 = 56.67
		if summary.OverallAvg < 56.6 || summary.OverallAvg > 56.7 {
			t.Errorf("Expected overall avg around 56.67, got %f", summary.OverallAvg)
		}

		// Overall success rate should be (100*0.99 + 200*0.98) / 300 = 295/300 = 0.983
		if summary.OverallSuccessRate < 0.983 || summary.OverallSuccessRate > 0.984 {
			t.Errorf("Expected overall success rate around 0.983, got %f", summary.OverallSuccessRate)
		}
	})

	t.Run("Empty aggregation", func(t *testing.T) {
		aggregated := []AggregatedMetrics{}
		summary := CalculateSummary(aggregated)

		if summary.TotalRequests != 0 {
			t.Errorf("Expected total requests 0, got %d", summary.TotalRequests)
		}
	})
}

func TestRingBufferConcurrency(t *testing.T) {
	t.Run("Concurrent push", func(t *testing.T) {
		rb := NewRingBuffer(1000)
		var wg sync.WaitGroup

		// Launch 10 goroutines, each pushing 100 records
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					rb.Push(MetricRecord{
						Timestamp: time.Now(),
						Duration:  time.Duration(id*100+j) * time.Millisecond,
					})
				}
			}(i)
		}

		wg.Wait()

		if rb.Size() != 1000 {
			t.Errorf("Expected size 1000, got %d", rb.Size())
		}
	})

	t.Run("Concurrent push and read", func(t *testing.T) {
		rb := NewRingBuffer(1000)
		var wg sync.WaitGroup

		// Launch writers
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					rb.Push(MetricRecord{
						Timestamp: time.Now(),
						Duration:  time.Duration(id*100+j) * time.Millisecond,
					})
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		// Launch readers
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					_ = rb.GetAll()
					_ = rb.Size()
					time.Sleep(time.Microsecond)
				}
			}()
		}

		wg.Wait()

		if rb.Size() != 500 {
			t.Errorf("Expected size 500, got %d", rb.Size())
		}
	})
}

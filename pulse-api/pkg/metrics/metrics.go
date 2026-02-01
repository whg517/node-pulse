package metrics

import (
	"sort"
	"sync"
	"time"
)

// MetricType represents the type of metric being collected
type MetricType string

const (
	MetricTypeAPI       MetricType = "api"
	MetricTypeDashboard MetricType = "dashboard"
	MetricTypeDatabase  MetricType = "database"
)

// MetricRecord represents a single performance metric
type MetricRecord struct {
	Timestamp  time.Time
	Endpoint   string
	Method     string
	Duration   time.Duration
	Status     int
	MetricType MetricType
}

// AggregatedMetrics represents aggregated metrics for a time window
type AggregatedMetrics struct {
	TimeWindow   time.Time `json:"time_window"`
	Endpoint     string    `json:"endpoint"`
	MetricType   string    `json:"metric_type"`
	Count        int64     `json:"count"`
	AvgDuration  float64   `json:"avg_duration_ms"`
	MinDuration  float64   `json:"min_duration_ms"`
	MaxDuration  float64   `json:"max_duration_ms"`
	P50          float64   `json:"p50_ms"`
	P95          float64   `json:"p95_ms"`
	P99          float64   `json:"p99_ms"`
	SuccessRate  float64   `json:"success_rate"`
}

// MetricsSummary represents overall metrics summary
type MetricsSummary struct {
	TotalRequests      int64   `json:"total_requests"`
	OverallAvg         float64 `json:"overall_avg_ms"`
	OverallP95         float64 `json:"overall_p95_ms"`
	OverallP99         float64 `json:"overall_p99_ms"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
}

// MetricsFilter represents filter options for querying metrics
type MetricsFilter struct {
	MetricType    *MetricType
	Endpoint      string
	StartTime     time.Time
	EndTime       time.Time
	Aggregation   time.Duration // 1m, 5m, 15m, etc.
}

// RingBuffer implements a thread-safe circular buffer for metric records
type RingBuffer struct {
	mu       sync.RWMutex
	buffer   []MetricRecord
	capacity int
	size     int
	head     int
	tail     int
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buffer:   make([]MetricRecord, capacity),
		capacity: capacity,
		size:     0,
		head:     0,
		tail:     0,
	}
}

// Push adds a metric record to the ring buffer
func (rb *RingBuffer) Push(record MetricRecord) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.tail] = record
	rb.tail = (rb.tail + 1) % rb.capacity

	if rb.size < rb.capacity {
		rb.size++
	} else {
		// Buffer is full, move head forward (overwrite oldest)
		rb.head = (rb.head + 1) % rb.capacity
	}
}

// GetAll returns all metric records in the buffer
func (rb *RingBuffer) GetAll() []MetricRecord {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]MetricRecord, rb.size)
	for i := 0; i < rb.size; i++ {
		idx := (rb.head + i) % rb.capacity
		result[i] = rb.buffer[idx]
	}
	return result
}

// GetFiltered returns metric records filtered by the given criteria
func (rb *RingBuffer) GetFiltered(filter MetricsFilter) []MetricRecord {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var result []MetricRecord
	for i := 0; i < rb.size; i++ {
		idx := (rb.head + i) % rb.capacity
		record := rb.buffer[idx]

		// Apply filters
		if filter.MetricType != nil && record.MetricType != *filter.MetricType {
			continue
		}
		if filter.Endpoint != "" && record.Endpoint != filter.Endpoint {
			continue
		}
		if !filter.StartTime.IsZero() && record.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && record.Timestamp.After(filter.EndTime) {
			continue
		}

		result = append(result, record)
	}
	return result
}

// Clear removes all records from the buffer
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.size = 0
	rb.head = 0
	rb.tail = 0
}

// Size returns the current number of records in the buffer
func (rb *RingBuffer) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// CalculatePercentiles calculates P50, P95, P99 percentiles from durations
func CalculatePercentiles(durations []time.Duration) (p50, p95, p99 time.Duration) {
	if len(durations) == 0 {
		return 0, 0, 0
	}

	// Sort durations
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	n := len(durations)

	// Helper function to get value at percentile
	getPercentile := func(p float64) time.Duration {
		idx := int(float64(n) * p / 100)
		if idx >= n {
			idx = n - 1
		}
		return durations[idx]
	}

	return getPercentile(50), getPercentile(95), getPercentile(99)
}

// AggregateMetrics aggregates raw metrics into time windows
func AggregateMetrics(records []MetricRecord, windowSize time.Duration) []AggregatedMetrics {
	if len(records) == 0 {
		return []AggregatedMetrics{}
	}

	// Group records by endpoint and time window
	groups := make(map[string]map[int64][]MetricRecord)

	for _, record := range records {
		// Calculate window key (truncated to window size)
		windowTime := record.Timestamp.Truncate(windowSize)
		key := record.Endpoint + "|" + record.MetricType.String() + "|" + windowTime.Format(time.RFC3339)

		if _, exists := groups[key]; !exists {
			groups[key] = make(map[int64][]MetricRecord)
		}
		groups[key][windowTime.Unix()] = append(groups[key][windowTime.Unix()], record)
	}

	// Calculate aggregated metrics for each group
	var result []AggregatedMetrics
	for _, group := range groups {
		for _, windowRecords := range group {
			if len(windowRecords) == 0 {
				continue
			}

			agg := calculateAggregation(windowRecords)
			result = append(result, agg)
		}
	}

	// Sort by time window
	sort.Slice(result, func(i, j int) bool {
		return result[i].TimeWindow.Before(result[j].TimeWindow)
	})

	return result
}

// calculateAggregation calculates aggregated metrics for a set of records
func calculateAggregation(records []MetricRecord) AggregatedMetrics {
	if len(records) == 0 {
		return AggregatedMetrics{}
	}

	var totalDuration time.Duration
	var minDuration time.Duration = records[0].Duration
	var maxDuration time.Duration = records[0].Duration
	var successCount int
	durations := make([]time.Duration, 0, len(records))

	for _, record := range records {
		totalDuration += record.Duration
		durations = append(durations, record.Duration)

		if record.Duration < minDuration {
			minDuration = record.Duration
		}
		if record.Duration > maxDuration {
			maxDuration = record.Duration
		}

		if record.Status >= 200 && record.Status < 400 {
			successCount++
		}
	}

	p50, p95, p99 := CalculatePercentiles(durations)

	return AggregatedMetrics{
		TimeWindow:  records[0].Timestamp,
		Endpoint:    records[0].Endpoint,
		MetricType:  records[0].MetricType.String(),
		Count:       int64(len(records)),
		AvgDuration: float64(totalDuration) / float64(len(records)) / float64(time.Millisecond),
		MinDuration: float64(minDuration) / float64(time.Millisecond),
		MaxDuration: float64(maxDuration) / float64(time.Millisecond),
		P50:         float64(p50) / float64(time.Millisecond),
		P95:         float64(p95) / float64(time.Millisecond),
		P99:         float64(p99) / float64(time.Millisecond),
		SuccessRate: float64(successCount) / float64(len(records)),
	}
}

// CalculateSummary calculates overall summary from aggregated metrics
func CalculateSummary(aggregated []AggregatedMetrics) MetricsSummary {
	if len(aggregated) == 0 {
		return MetricsSummary{}
	}

	var totalRequests int64
	var totalDuration time.Duration
	var successCount int

	for _, agg := range aggregated {
		totalRequests += agg.Count
		totalDuration += time.Duration(agg.AvgDuration*float64(agg.Count)) * time.Millisecond
		successCount += int(float64(agg.Count) * agg.SuccessRate)
	}

	// For percentiles, we'd need raw data. This is an approximation.
	// A more accurate implementation would track all durations.
	return MetricsSummary{
		TotalRequests:      totalRequests,
		OverallAvg:         float64(totalDuration) / float64(totalRequests) / float64(time.Millisecond),
		OverallP95:         calculateOverallPercentile(aggregated, 95),
		OverallP99:         calculateOverallPercentile(aggregated, 99),
		OverallSuccessRate: float64(successCount) / float64(totalRequests),
	}
}

// calculateOverallPercentile calculates overall percentile from aggregated metrics
// This is an approximation - accurate calculation requires raw data
func calculateOverallPercentile(aggregated []AggregatedMetrics, percentile float64) float64 {
	if len(aggregated) == 0 {
		return 0
	}

	// Weighted average of percentiles by request count
	var weightedSum float64
	var totalRequests int64

	for _, agg := range aggregated {
		var p float64
		switch percentile {
		case 95:
			p = agg.P95
		case 99:
			p = agg.P99
		default:
			p = agg.P50
		}
		weightedSum += p * float64(agg.Count)
		totalRequests += agg.Count
	}

	if totalRequests == 0 {
		return 0
	}

	return weightedSum / float64(totalRequests)
}

// String returns the string representation of MetricType
func (mt MetricType) String() string {
	return string(mt)
}

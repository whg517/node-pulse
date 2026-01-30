package probe

import (
	"fmt"
	"math"
	"testing"
)

// TestCalculateMean tests mean calculation
func TestCalculateMean(t *testing.T) {
	collector := NewCoreMetricsCollector()

	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{
			name:     "normal samples",
			samples:  []float64{1.0, 2.0, 3.0},
			expected: 2.0,
		},
		{
			name:     "decimal values",
			samples:  []float64{100.5, 200.3, 300.2},
			expected: 200.33,
		},
		{
			name:     "single value",
			samples:  []float64{42.5},
			expected: 42.5,
		},
		{
			name:     "empty samples",
			samples:  []float64{},
			expected: 0,
		},
		{
			name:     "RTT samples",
			samples:  []float64{100.0, 102.5, 98.3, 105.2, 99.1},
			expected: 101.02,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.CalculateMean(tt.samples)
			if result != tt.expected {
				t.Errorf("CalculateMean(%v) = %v, want %v", tt.samples, result, tt.expected)
			}
		})
	}
}

// TestCalculateMedian tests median calculation
func TestCalculateMedian(t *testing.T) {
	collector := NewCoreMetricsCollector()

	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{
			name:     "odd number of samples",
			samples:  []float64{1.0, 2.0, 3.0},
			expected: 2.0,
		},
		{
			name:     "even number of samples",
			samples:  []float64{1.0, 2.0, 3.0, 4.0},
			expected: 2.5,
		},
		{
			name:     "unsorted samples",
			samples:  []float64{3.0, 1.0, 4.0, 2.0},
			expected: 2.5,
		},
		{
			name:     "single value",
			samples:  []float64{42.5},
			expected: 42.5,
		},
		{
			name:     "empty samples",
			samples:  []float64{},
			expected: 0,
		},
		{
			name:     "RTT samples (odd)",
			samples:  []float64{100.0, 102.5, 98.3, 105.2, 99.1},
			expected: 100.0,
		},
		{
			name:     "RTT samples (even)",
			samples:  []float64{100.0, 102.5, 98.3, 105.2},
			expected: 101.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.CalculateMedian(tt.samples)
			if result != tt.expected {
				t.Errorf("CalculateMedian(%v) = %v, want %v", tt.samples, result, tt.expected)
			}
		})
	}
}

// TestCalculateVariance tests variance calculation
func TestCalculateVariance(t *testing.T) {
	collector := NewCoreMetricsCollector()

	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{
			name:     "constant values",
			samples:  []float64{5.0, 5.0, 5.0},
			expected: 0,
		},
		{
			name:     "varying values",
			samples:  []float64{1.0, 2.0, 3.0},
			expected: 0.67,
		},
		{
			name:     "RTT samples",
			samples:  []float64{10.0, 12.0, 8.0, 15.0, 11.0},
			expected: 5.36,
		},
		{
			name:     "empty samples",
			samples:  []float64{},
			expected: 0,
		},
		{
			name:     "single value",
			samples:  []float64{42.5},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.CalculateVariance(tt.samples)
			if result != tt.expected {
				t.Errorf("CalculateVariance(%v) = %v, want %v", tt.samples, result, tt.expected)
			}
		})
	}
}

// TestCalculateJitter tests jitter calculation
func TestCalculateJitter(t *testing.T) {
	collector := NewCoreMetricsCollector()

	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{
			name:     "constant values",
			samples:  []float64{10.0, 10.0, 10.0},
			expected: 0,
		},
		{
			name:     "varying values",
			samples:  []float64{10.0, 12.0, 8.0, 15.0, 11.0},
			expected: 4.25,
		},
		{
			name:     "two samples",
			samples:  []float64{10.0, 12.0},
			expected: 2.0,
		},
		{
			name:     "empty samples",
			samples:  []float64{},
			expected: 0,
		},
		{
			name:     "single sample",
			samples:  []float64{42.5},
			expected: 0,
		},
		{
			name:     "RTT fluctuations",
			samples:  []float64{100.0, 102.5, 98.3, 105.2, 99.1},
			expected: 4.93,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.CalculateJitter(tt.samples)
			if result != tt.expected {
				t.Errorf("CalculateJitter(%v) = %v, want %v", tt.samples, result, tt.expected)
			}
		})
	}
}

// TestCalculatePacketLossRate tests packet loss rate calculation
func TestCalculatePacketLossRate(t *testing.T) {
	collector := NewCoreMetricsCollector()

	tests := []struct {
		name     string
		sent     int
		received int
		expected float64
	}{
		{
			name:     "no loss",
			sent:     10,
			received: 10,
			expected: 0,
		},
		{
			name:     "50% loss",
			sent:     10,
			received: 5,
			expected: 50.0,
		},
		{
			name:     "100% loss",
			sent:     10,
			received: 0,
			expected: 100.0,
		},
		{
			name:     "20% loss",
			sent:     10,
			received: 8,
			expected: 20.0,
		},
		{
			name:     "partial loss",
			sent:     100,
			received: 95,
			expected: 5.0,
		},
		{
			name:     "no packets sent",
			sent:     0,
			received: 0,
			expected: 0,
		},
		{
			name:     "decimal loss rate",
			sent:     3,
			received: 2,
			expected: 33.33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.CalculatePacketLossRate(tt.sent, tt.received)
			if result != tt.expected {
				t.Errorf("CalculatePacketLossRate(%d, %d) = %v, want %v", tt.sent, tt.received, result, tt.expected)
			}
		})
	}
}

// TestCalculateFromSamples tests complete metrics calculation
func TestCalculateFromSamples(t *testing.T) {
	collector := NewCoreMetricsCollector()

	tests := []struct {
		name     string
		samples  []SamplePoint
		sent     int
		received int
		expected CoreMetrics
	}{
		{
			name: "all successful",
			samples: []SamplePoint{
				{RTTMs: 100.0, Success: true},
				{RTTMs: 102.5, Success: true},
				{RTTMs: 98.3, Success: true},
			},
			sent:     3,
			received: 3,
			expected: CoreMetrics{
				RTTMs:          100.27,
				RTTMedianMs:    100.0,
				RTTVarianceMs:  2.98,
				JitterMs:       3.35,
				PacketLossRate: 0,
				SampleCount:    3,
			},
		},
		{
			name: "partial loss",
			samples: []SamplePoint{
				{RTTMs: 100.0, Success: true},
				{RTTMs: 0, Success: false},
				{RTTMs: 102.5, Success: true},
			},
			sent:     3,
			received: 2,
			expected: CoreMetrics{
				RTTMs:          101.25,
				RTTMedianMs:    101.25,
				RTTVarianceMs:  1.56,
				JitterMs:       2.5,
				PacketLossRate: 33.33,
				SampleCount:    3,
			},
		},
		{
			name:     "empty samples",
			samples:  []SamplePoint{},
			sent:     0,
			received: 0,
			expected: CoreMetrics{
				RTTMs:          0,
				RTTMedianMs:    0,
				RTTVarianceMs:  0,
				JitterMs:       0,
				PacketLossRate: 0,
				SampleCount:    0,
			},
		},
		{
			name: "all failed",
			samples: []SamplePoint{
				{RTTMs: 0, Success: false},
				{RTTMs: 0, Success: false},
				{RTTMs: 0, Success: false},
			},
			sent:     3,
			received: 0,
			expected: CoreMetrics{
				RTTMs:          0,
				RTTMedianMs:    0,
				RTTVarianceMs:  0,
				JitterMs:       0,
				PacketLossRate: 100.0,
				SampleCount:    3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.CalculateFromSamples(tt.samples, tt.sent, tt.received)

			// Use approximate comparison for floating point values
			if !almostEqualInThousandths(result.RTTMs, tt.expected.RTTMs) {
				t.Errorf("RTTMs = %v, want %v", result.RTTMs, tt.expected.RTTMs)
			}
			if !almostEqualInThousandths(result.RTTMedianMs, tt.expected.RTTMedianMs) {
				t.Errorf("RTTMedianMs = %v, want %v", result.RTTMedianMs, tt.expected.RTTMedianMs)
			}
			if !almostEqualInThousandths(result.RTTVarianceMs, tt.expected.RTTVarianceMs) {
				t.Errorf("RTTVarianceMs = %v, want %v", result.RTTVarianceMs, tt.expected.RTTVarianceMs)
			}
			if !almostEqualInThousandths(result.JitterMs, tt.expected.JitterMs) {
				t.Errorf("JitterMs = %v, want %v", result.JitterMs, tt.expected.JitterMs)
			}
			if !almostEqualInThousandths(result.PacketLossRate, tt.expected.PacketLossRate) {
				t.Errorf("PacketLossRate = %v, want %v", result.PacketLossRate, tt.expected.PacketLossRate)
			}
			if result.SampleCount != tt.expected.SampleCount {
				t.Errorf("SampleCount = %v, want %v", result.SampleCount, tt.expected.SampleCount)
			}
		})
	}
}

// TestMeasurementPrecision tests that measurements meet precision requirements (≤1ms)
func TestMeasurementPrecision(t *testing.T) {
	collector := NewCoreMetricsCollector()

	// Test RTT precision
	rttSamples := []float64{123.456, 124.789, 125.234}
	mean := collector.CalculateMean(rttSamples)

	// Check that precision is 2 decimal places (better than ≤1ms requirement)
	strValue := fmt.Sprintf("%.2f", mean)
	if len(strValue) > 0 {
		// Verify we can represent sub-millisecond precision
		if mean == math.Trunc(mean) {
			t.Error("Expected sub-millisecond precision, got integer value")
		}
	}

	// Test that variance and jitter also maintain precision
	jitter := collector.CalculateJitter(rttSamples)
	if jitter == math.Trunc(jitter) {
		t.Error("Expected sub-millisecond jitter precision, got integer value")
	}

	variance := collector.CalculateVariance(rttSamples)
	if variance == math.Trunc(variance) && variance > 0 {
		t.Error("Expected sub-millisecond variance precision, got integer value")
	}
}

// almostEqualInThousandths checks if two float64 values are almost equal (within 0.01)
func almostEqualInThousandths(a, b float64) bool {
	const epsilon = 0.01
	return math.Abs(a-b) <= epsilon
}

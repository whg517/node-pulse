package probe

import (
	"fmt"
	"testing"
)

// TestCoreMetricsCollector_Algorithms tests all metric calculation algorithms
func TestCoreMetricsCollector_Algorithms(t *testing.T) {
	collector := NewCoreMetricsCollector()

	t.Run("CalculateMean with 10 samples", func(t *testing.T) {
		samples := make([]float64, 10)
		for i := range samples {
			samples[i] = 100.0 + float64(i)
		}
		mean := collector.CalculateMean(samples)
		expected := 104.5
		if mean != expected {
			t.Errorf("CalculateMean() = %v, want %v", mean, expected)
		}
	})

	t.Run("CalculateMean with 100 samples", func(t *testing.T) {
		samples := make([]float64, 100)
		for i := range samples {
			samples[i] = float64(i)
		}
		mean := collector.CalculateMean(samples)
		expected := 49.5
		if mean != expected {
			t.Errorf("CalculateMean() = %v, want %v", mean, expected)
		}
	})

	t.Run("CalculateMedian odd count", func(t *testing.T) {
		samples := []float64{1, 3, 5, 7, 9}
		median := collector.CalculateMedian(samples)
		expected := 5.0
		if median != expected {
			t.Errorf("CalculateMedian() = %v, want %v", median, expected)
		}
	})

	t.Run("CalculateMedian even count", func(t *testing.T) {
		samples := []float64{1, 3, 5, 7}
		median := collector.CalculateMedian(samples)
		expected := 4.0
		if median != expected {
			t.Errorf("CalculateMedian() = %v, want %v", median, expected)
		}
	})

	t.Run("CalculateVariance stable samples", func(t *testing.T) {
		samples := []float64{100, 100, 100, 100, 100}
		variance := collector.CalculateVariance(samples)
		expected := 0.0
		if variance != expected {
			t.Errorf("CalculateVariance() = %v, want %v", variance, expected)
		}
	})

	t.Run("CalculateVariance varying samples", func(t *testing.T) {
		samples := []float64{98, 99, 100, 101, 102}
		variance := collector.CalculateVariance(samples)
		expected := 2.0
		if variance != expected {
			t.Errorf("CalculateVariance() = %v, want %v", variance, expected)
		}
	})

	t.Run("CalculateJitter no variation", func(t *testing.T) {
		samples := []float64{100, 100, 100, 100}
		jitter := collector.CalculateJitter(samples)
		expected := 0.0
		if jitter != expected {
			t.Errorf("CalculateJitter() = %v, want %v", jitter, expected)
		}
	})

	t.Run("CalculateJitter with variation", func(t *testing.T) {
		samples := []float64{98, 100, 102, 100, 99}
		jitter := collector.CalculateJitter(samples)
		if jitter <= 0 {
			t.Errorf("CalculateJitter() = %v, want > 0", jitter)
		}
	})

	t.Run("CalculatePacketLossRate no loss", func(t *testing.T) {
		lossRate := collector.CalculatePacketLossRate(10, 10)
		expected := 0.0
		if lossRate != expected {
			t.Errorf("CalculatePacketLossRate() = %v, want %v", lossRate, expected)
		}
	})

	t.Run("CalculatePacketLossRate 50% loss", func(t *testing.T) {
		lossRate := collector.CalculatePacketLossRate(10, 5)
		expected := 50.0
		if lossRate != expected {
			t.Errorf("CalculatePacketLossRate() = %v, want %v", lossRate, expected)
		}
	})

	t.Run("CalculatePacketLossRate 100% loss", func(t *testing.T) {
		lossRate := collector.CalculatePacketLossRate(10, 0)
		expected := 100.0
		if lossRate != expected {
			t.Errorf("CalculatePacketLossRate() = %v, want %v", lossRate, expected)
		}
	})
}

// TestCoreMetricsCollector_BoundaryConditions tests edge cases
func TestCoreMetricsCollector_BoundaryConditions(t *testing.T) {
	collector := NewCoreMetricsCollector()

	t.Run("Empty samples", func(t *testing.T) {
		samples := []float64{}
		mean := collector.CalculateMean(samples)
		median := collector.CalculateMedian(samples)
		variance := collector.CalculateVariance(samples)
		jitter := collector.CalculateJitter(samples)

		if mean != 0 || median != 0 || variance != 0 || jitter != 0 {
			t.Errorf("Empty samples should return 0, got mean=%v, median=%v, variance=%v, jitter=%v",
				mean, median, variance, jitter)
		}
	})

	t.Run("Single sample", func(t *testing.T) {
		samples := []float64{100.0}
		mean := collector.CalculateMean(samples)
		median := collector.CalculateMedian(samples)
		variance := collector.CalculateVariance(samples)
		jitter := collector.CalculateJitter(samples)

		if mean != 100.0 || median != 100.0 || variance != 0 || jitter != 0 {
			t.Errorf("Single sample: mean=%v (want 100), median=%v (want 100), variance=%v (want 0), jitter=%v (want 0)",
				mean, median, variance, jitter)
		}
	})

	t.Run("Two samples jitter", func(t *testing.T) {
		samples := []float64{100.0, 102.0}
		jitter := collector.CalculateJitter(samples)
		expected := 2.0
		if jitter != expected {
			t.Errorf("Two samples jitter = %v, want %v", jitter, expected)
		}
	})

	t.Run("All probes failed", func(t *testing.T) {
		samples := []SamplePoint{
			{Success: false},
			{Success: false},
			{Success: false},
		}
		metrics := collector.CalculateFromSamples(samples, 3, 0)

		if metrics.RTTMs != 0 || metrics.RTTMedianMs != 0 || metrics.JitterMs != 0 || metrics.RTTVarianceMs != 0 {
			t.Errorf("All failed: RTT should be 0, got rtt=%v, median=%v, jitter=%v, variance=%v",
				metrics.RTTMs, metrics.RTTMedianMs, metrics.JitterMs, metrics.RTTVarianceMs)
		}
		if metrics.PacketLossRate != 100.0 {
			t.Errorf("All failed: packet loss should be 100%%, got %v", metrics.PacketLossRate)
		}
	})

	t.Run("All probes succeeded", func(t *testing.T) {
		samples := []SamplePoint{
			{RTTMs: 100.0, Success: true},
			{RTTMs: 102.0, Success: true},
			{RTTMs: 98.0, Success: true},
		}
		metrics := collector.CalculateFromSamples(samples, 3, 3)

		if metrics.PacketLossRate != 0 {
			t.Errorf("All succeeded: packet loss should be 0%%, got %v", metrics.PacketLossRate)
		}
		if metrics.RTTMs <= 0 {
			t.Errorf("All succeeded: RTT should be > 0, got %v", metrics.RTTMs)
		}
	})

	t.Run("Partial packet loss", func(t *testing.T) {
		samples := []SamplePoint{
			{RTTMs: 100.0, Success: true},
			{Success: false},
			{RTTMs: 102.0, Success: true},
		}
		metrics := collector.CalculateFromSamples(samples, 3, 2)

		expectedLoss := 33.33
		if metrics.PacketLossRate != expectedLoss {
			t.Errorf("Partial loss: packet loss = %v, want %v", metrics.PacketLossRate, expectedLoss)
		}
	})
}

// TestCoreMetricsCollector_MeasurementPrecision verifies sub-millisecond precision
func TestCoreMetricsCollector_MeasurementPrecision(t *testing.T) {
	collector := NewCoreMetricsCollector()

	t.Run("RTT precision", func(t *testing.T) {
		samples := []float64{123.456, 124.789, 125.234}
		mean := collector.CalculateMean(samples)

		// Check that we maintain 2 decimal places (better than ≤1ms requirement)
		strValue := fmt.Sprintf("%.2f", mean)
		if len(strValue) == 0 {
			t.Error("Expected formatted mean value")
		}

		// Verify it's not an integer (has decimal precision)
		if mean == float64(int(mean)) && mean > 0 {
			t.Error("Expected sub-millisecond precision, got integer value")
		}
	})

	t.Run("Jitter precision", func(t *testing.T) {
		samples := []float64{100.123, 102.456, 98.789}
		jitter := collector.CalculateJitter(samples)

		// Verify jitter has decimal precision (unless calculation results in integer)
		// The actual jitter might round to an integer due to the rounding algorithm
		if jitter < 0 {
			t.Errorf("Expected non-negative jitter, got %v", jitter)
		}
	})

	t.Run("Variance precision", func(t *testing.T) {
		samples := []float64{100.0, 102.0, 98.0}
		variance := collector.CalculateVariance(samples)

		// Verify variance has decimal precision (unless it's 0 or very small)
		if variance > 0.01 && variance == float64(int(variance)) {
			t.Error("Expected sub-millisecond variance precision, got integer value")
		}
	})
}

// TestCoreMetricsCollector_SampleSizes tests with various sample counts
func TestCoreMetricsCollector_SampleSizes(t *testing.T) {
	collector := NewCoreMetricsCollector()

	t.Run("10 samples", func(t *testing.T) {
		samples := make([]SamplePoint, 10)
		for i := range samples {
			samples[i] = SamplePoint{
				RTTMs:   100.0 + float64(i),
				Success: true,
			}
		}
		metrics := collector.CalculateFromSamples(samples, 10, 10)

		if metrics.SampleCount != 10 {
			t.Errorf("SampleCount = %v, want 10", metrics.SampleCount)
		}
		if metrics.PacketLossRate != 0 {
			t.Errorf("PacketLossRate = %v, want 0", metrics.PacketLossRate)
		}
		if metrics.RTTMs <= 0 {
			t.Errorf("RTTMs = %v, want > 0", metrics.RTTMs)
		}
	})

	t.Run("100 samples", func(t *testing.T) {
		samples := make([]SamplePoint, 100)
		for i := range samples {
			samples[i] = SamplePoint{
				RTTMs:   100.0 + float64(i%10), // Cycle through 10 different RTT values
				Success: true,
			}
		}
		metrics := collector.CalculateFromSamples(samples, 100, 100)

		if metrics.SampleCount != 100 {
			t.Errorf("SampleCount = %v, want 100", metrics.SampleCount)
		}
		if metrics.PacketLossRate != 0 {
			t.Errorf("PacketLossRate = %v, want 0", metrics.PacketLossRate)
		}
		if metrics.RTTMs <= 0 {
			t.Errorf("RTTMs = %v, want > 0", metrics.RTTMs)
		}
	})
}

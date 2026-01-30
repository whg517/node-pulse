package probe

import (
	"math"
	"sort"
)

// CoreMetrics represents calculated core network quality metrics
type CoreMetrics struct {
	RTTMs          float64 `json:"rtt_ms"`           // RTT mean (milliseconds)
	RTTMedianMs    float64 `json:"rtt_median_ms"`    // RTT median (milliseconds)
	RTTVarianceMs  float64 `json:"rtt_variance_ms"`  // RTT variance (milliseconds²)
	JitterMs       float64 `json:"jitter_ms"`        // Delay jitter (milliseconds)
	PacketLossRate float64 `json:"packet_loss_rate"` // Packet loss rate (%)
	SampleCount    int     `json:"sample_count"`     // Number of sample points
}

// SamplePoint represents a single probe sample point
type SamplePoint struct {
	RTTMs     float64 `json:"rtt_ms"`
	Timestamp string  `json:"timestamp"`
	Success   bool    `json:"success"`
}

// CoreMetricsCollector calculates core metrics from sample points
type CoreMetricsCollector struct{}

// NewCoreMetricsCollector creates a new CoreMetricsCollector
func NewCoreMetricsCollector() *CoreMetricsCollector {
	return &CoreMetricsCollector{}
}

// CalculateMean calculates the mean of a sample set
func (c *CoreMetricsCollector) CalculateMean(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range samples {
		sum += v
	}
	return math.Round((sum/float64(len(samples)))*100) / 100
}

// CalculateMedian calculates the median of a sample set
func (c *CoreMetricsCollector) CalculateMedian(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	// Create a copy to avoid modifying the original slice
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)

	// Get median
	n := len(sorted)
	if n%2 == 0 {
		return math.Round(((sorted[n/2-1] + sorted[n/2]) / 2) * 100) / 100
	}
	return math.Round(sorted[n/2] * 100) / 100
}

// CalculateVariance calculates the variance of a sample set
func (c *CoreMetricsCollector) CalculateVariance(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	mean := c.CalculateMean(samples)
	sum := 0.0
	for _, v := range samples {
		diff := v - mean
		sum += diff * diff
	}
	variance := sum / float64(len(samples))
	return math.Round(variance*100) / 100
}

// CalculateJitter calculates the jitter (adjacent sample difference)
func (c *CoreMetricsCollector) CalculateJitter(samples []float64) float64 {
	if len(samples) < 2 {
		return 0
	}
	sum := 0.0
	for i := 1; i < len(samples); i++ {
		diff := samples[i] - samples[i-1]
		if diff < 0 {
			diff = -diff // Absolute value
		}
		sum += diff
	}
	jitter := sum / float64(len(samples)-1)
	return math.Round(jitter*100) / 100
}

// CalculatePacketLossRate calculates the packet loss rate
func (c *CoreMetricsCollector) CalculatePacketLossRate(sent, received int) float64 {
	if sent == 0 {
		return 0
	}
	lossRate := (1.0 - float64(received)/float64(sent)) * 100
	return math.Round(lossRate*100) / 100
}

// CalculateFromSamples calculates core metrics from sample points
func (c *CoreMetricsCollector) CalculateFromSamples(samples []SamplePoint, sent, received int) CoreMetrics {
	// Extract successful RTT samples
	rttSamples := make([]float64, 0, len(samples))
	for _, s := range samples {
		if s.Success {
			rttSamples = append(rttSamples, s.RTTMs)
		}
	}

	// Calculate metrics
	rttMs := c.CalculateMean(rttSamples)
	rttMedianMs := c.CalculateMedian(rttSamples)
	varianceMs := c.CalculateVariance(rttSamples)
	jitterMs := c.CalculateJitter(rttSamples)
	packetLossRate := c.CalculatePacketLossRate(sent, received)

	return CoreMetrics{
		RTTMs:          rttMs,
		RTTMedianMs:    rttMedianMs,
		RTTVarianceMs:  varianceMs,
		JitterMs:       jitterMs,
		PacketLossRate: packetLossRate,
		SampleCount:    len(samples),
	}
}

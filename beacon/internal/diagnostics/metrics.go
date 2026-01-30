package diagnostics

// PrometheusMetrics contains Prometheus metrics summary
type PrometheusMetrics struct {
	BeaconUp       float64 `json:"beacon_up"`
	RTTSeconds     float64 `json:"beacon_rtt_seconds"`
	PacketLossRate float64 `json:"beacon_packet_loss_rate"`
	JitterMs       float64 `json:"beacon_jitter_ms"`
}

// collectPrometheusMetrics collects Prometheus metrics summary
// NOTE: For Story 3.10, metrics are derived from network diagnostics.
// Future work: Integrate with actual metrics.MetricsCollector to get real-time Prometheus metrics.
func (c *collector) collectPrometheusMetrics() (*PrometheusMetrics, error) {
	// Collect network status to derive metrics
	networkStatus, err := c.collectNetworkStatus()
	if err != nil {
		return nil, err
	}

	// Convert RTT from milliseconds to seconds for Prometheus format
	rttSeconds := networkStatus.RTTMs.Avg / 1000.0

	metrics := &PrometheusMetrics{
		BeaconUp:       1,                           // Beacon is running
		RTTSeconds:     rttSeconds,                  // Actual RTT from network check
		PacketLossRate: networkStatus.PacketLossRate, // Actual packet loss rate
		JitterMs:       0,                           // Requires multiple samples to calculate (not implemented yet)
	}

	return metrics, nil
}

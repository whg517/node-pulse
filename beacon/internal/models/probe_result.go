package models

import "time"

// TCPProbeResult represents the result of a TCP probe operation.
// Use this struct for TCP-specific probe results that include RTT measurements.
type TCPProbeResult struct {
	Success        bool    `json:"success"`         // Connectivity (success/failure)
	RTTMs          float64 `json:"rtt_ms"`          // Round-trip time in milliseconds (mean)
	RTTMedianMs    float64 `json:"rtt_median_ms"`   // RTT median in milliseconds
	JitterMs       float64 `json:"jitter_ms"`       // Delay jitter in milliseconds
	VarianceMs     float64 `json:"variance_ms"`     // RTT variance in milliseconds²
	PacketLossRate float64 `json:"packet_loss_rate"` // Packet loss rate (0-100%)
	SampleCount    int     `json:"sample_count"`    // Number of sample points
	ErrorMessage   string  `json:"error_message"`   // Error message if failed
	Timestamp      string  `json:"timestamp"`       // Probe timestamp (ISO 8601)
}

// UDPProbeResult represents the result of a UDP probe operation.
// Use this struct for UDP-specific probe results that include packet loss measurements.
type UDPProbeResult struct {
	Success         bool    `json:"success"`           // Connectivity (success/failure)
	PacketLossRate  float64 `json:"packet_loss_rate"`  // Packet loss rate (0-100%)
	RTTMs           float64 `json:"rtt_ms"`            // Average round-trip time in milliseconds
	RTTMedianMs     float64 `json:"rtt_median_ms"`     // RTT median in milliseconds
	JitterMs        float64 `json:"jitter_ms"`         // Delay jitter in milliseconds
	VarianceMs      float64 `json:"variance_ms"`       // RTT variance in milliseconds²
	SentPackets     int     `json:"sent_packets"`      // Number of packets sent
	ReceivedPackets int     `json:"received_packets"`  // Number of packets received
	SampleCount     int     `json:"sample_count"`      // Number of sample points
	ErrorMessage    string  `json:"error_message"`     // Error message if failed
	Timestamp       string  `json:"timestamp"`         // Probe timestamp (ISO 8601)
}

// ProbeResult represents a generic probe result for both TCP and UDP probes.
// Use this struct when you need a unified format for multiple probe types.
// The Metrics field contains type-specific measurements (RTT, packet loss, etc.).
type ProbeResult struct {
	Type         string                 `json:"type"`          // "tcp_ping" or "udp_ping"
	Target       string                 `json:"target"`        // Target IP:Port
	Success      bool                   `json:"success"`       // Probe success status
	Metrics      map[string]interface{} `json:"metrics"`       // RTT, packet loss, etc.
	ErrorMessage string                 `json:"error_message"` // Error message if failed
	Timestamp    string                 `json:"timestamp"`     // Probe timestamp (ISO 8601)
}

// NewTCPProbeResult creates a new TCP probe result with current timestamp (backward compatible)
func NewTCPProbeResult(success bool, rttMs float64, errorMessage string) *TCPProbeResult {
	return &TCPProbeResult{
		Success:      success,
		RTTMs:        rttMs,
		ErrorMessage: errorMessage,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
}

// NewTCPProbeResultWithMetrics creates a new TCP probe result with core metrics
func NewTCPProbeResultWithMetrics(success bool, rttMs, rttMedianMs, jitterMs, varianceMs, packetLossRate float64, sampleCount int, errorMessage string) *TCPProbeResult {
	return &TCPProbeResult{
		Success:        success,
		RTTMs:          rttMs,
		RTTMedianMs:    rttMedianMs,
		JitterMs:       jitterMs,
		VarianceMs:     varianceMs,
		PacketLossRate: packetLossRate,
		SampleCount:    sampleCount,
		ErrorMessage:   errorMessage,
		Timestamp:      time.Now().Format(time.RFC3339),
	}
}

// NewUDPProbeResult creates a new UDP probe result with current timestamp (backward compatible)
func NewUDPProbeResult(success bool, packetLossRate, rttMs float64, sent, received int, errorMessage string) *UDPProbeResult {
	return &UDPProbeResult{
		Success:         success,
		PacketLossRate:  packetLossRate,
		RTTMs:           rttMs,
		SentPackets:     sent,
		ReceivedPackets: received,
		ErrorMessage:    errorMessage,
		Timestamp:       time.Now().Format(time.RFC3339),
	}
}

// NewUDPProbeResultWithMetrics creates a new UDP probe result with core metrics
func NewUDPProbeResultWithMetrics(success bool, packetLossRate, rttMs, rttMedianMs, jitterMs, varianceMs float64, sent, received, sampleCount int, errorMessage string) *UDPProbeResult {
	return &UDPProbeResult{
		Success:         success,
		PacketLossRate:  packetLossRate,
		RTTMs:           rttMs,
		RTTMedianMs:     rttMedianMs,
		JitterMs:        jitterMs,
		VarianceMs:      varianceMs,
		SentPackets:     sent,
		ReceivedPackets: received,
		SampleCount:     sampleCount,
		ErrorMessage:    errorMessage,
		Timestamp:       time.Now().Format(time.RFC3339),
	}
}

// NewProbeResult creates a new generic probe result with current timestamp
func NewProbeResult(probeType, target string, success bool, metrics map[string]interface{}, errorMessage string) *ProbeResult {
	return &ProbeResult{
		Type:         probeType,
		Target:       target,
		Success:      success,
		Metrics:      metrics,
		ErrorMessage: errorMessage,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
}

// ToGenericResult converts a TCPProbeResult to a generic ProbeResult
func (r *TCPProbeResult) ToGenericResult(target string) *ProbeResult {
	metrics := map[string]interface{}{
		"rtt_ms": r.RTTMs,
	}

	return &ProbeResult{
		Type:         "tcp_ping",
		Target:       target,
		Success:      r.Success,
		Metrics:      metrics,
		ErrorMessage: r.ErrorMessage,
		Timestamp:    r.Timestamp,
	}
}

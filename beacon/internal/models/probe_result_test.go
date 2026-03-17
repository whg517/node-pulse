package models

import (
	"testing"
	"time"
)

// TestNewTCPProbeResult tests basic TCP probe result creation
func TestNewTCPProbeResult(t *testing.T) {
	result := NewTCPProbeResult(true, 1.5, "")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.RTTMs != 1.5 {
		t.Errorf("Expected RTTMs=1.5, got %f", result.RTTMs)
	}
	if result.ErrorMessage != "" {
		t.Errorf("Expected empty ErrorMessage, got %s", result.ErrorMessage)
	}
	if result.Timestamp == "" {
		t.Error("Expected non-empty Timestamp")
	}
	// Verify timestamp is valid RFC3339
	_, err := time.Parse(time.RFC3339, result.Timestamp)
	if err != nil {
		t.Errorf("Timestamp is not valid RFC3339: %v", err)
	}
}

// TestNewTCPProbeResult_Failure tests failed TCP probe result creation
func TestNewTCPProbeResult_Failure(t *testing.T) {
	result := NewTCPProbeResult(false, 0, "connection refused")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Success {
		t.Error("Expected Success=false")
	}
	if result.RTTMs != 0 {
		t.Errorf("Expected RTTMs=0, got %f", result.RTTMs)
	}
	if result.ErrorMessage != "connection refused" {
		t.Errorf("Expected 'connection refused', got '%s'", result.ErrorMessage)
	}
}

// TestNewTCPProbeResultWithMetrics tests TCP probe result with full metrics
func TestNewTCPProbeResultWithMetrics(t *testing.T) {
	result := NewTCPProbeResultWithMetrics(true, 2.5, 2.3, 0.1, 0.01, 0.0, 10, "")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.RTTMs != 2.5 {
		t.Errorf("Expected RTTMs=2.5, got %f", result.RTTMs)
	}
	if result.RTTMedianMs != 2.3 {
		t.Errorf("Expected RTTMedianMs=2.3, got %f", result.RTTMedianMs)
	}
	if result.JitterMs != 0.1 {
		t.Errorf("Expected JitterMs=0.1, got %f", result.JitterMs)
	}
	if result.VarianceMs != 0.01 {
		t.Errorf("Expected VarianceMs=0.01, got %f", result.VarianceMs)
	}
	if result.PacketLossRate != 0.0 {
		t.Errorf("Expected PacketLossRate=0.0, got %f", result.PacketLossRate)
	}
	if result.SampleCount != 10 {
		t.Errorf("Expected SampleCount=10, got %d", result.SampleCount)
	}
	if result.Timestamp == "" {
		t.Error("Expected non-empty Timestamp")
	}
}

// TestNewTCPProbeResultWithMetrics_WithError tests TCP probe result with error
func TestNewTCPProbeResultWithMetrics_WithError(t *testing.T) {
	result := NewTCPProbeResultWithMetrics(false, 0, 0, 0, 0, 100.0, 0, "timeout")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Success {
		t.Error("Expected Success=false")
	}
	if result.PacketLossRate != 100.0 {
		t.Errorf("Expected PacketLossRate=100.0, got %f", result.PacketLossRate)
	}
	if result.ErrorMessage != "timeout" {
		t.Errorf("Expected 'timeout', got '%s'", result.ErrorMessage)
	}
}

// TestNewUDPProbeResult tests basic UDP probe result creation
func TestNewUDPProbeResult(t *testing.T) {
	result := NewUDPProbeResult(true, 5.0, 1.2, 10, 10, "")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.PacketLossRate != 5.0 {
		t.Errorf("Expected PacketLossRate=5.0, got %f", result.PacketLossRate)
	}
	if result.RTTMs != 1.2 {
		t.Errorf("Expected RTTMs=1.2, got %f", result.RTTMs)
	}
	if result.SentPackets != 10 {
		t.Errorf("Expected SentPackets=10, got %d", result.SentPackets)
	}
	if result.ReceivedPackets != 10 {
		t.Errorf("Expected ReceivedPackets=10, got %d", result.ReceivedPackets)
	}
	if result.ErrorMessage != "" {
		t.Errorf("Expected empty ErrorMessage, got %s", result.ErrorMessage)
	}
	if result.Timestamp == "" {
		t.Error("Expected non-empty Timestamp")
	}
}

// TestNewUDPProbeResult_Failure tests failed UDP probe result creation
func TestNewUDPProbeResult_Failure(t *testing.T) {
	result := NewUDPProbeResult(false, 100.0, 0, 10, 0, "all packets lost")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Success {
		t.Error("Expected Success=false")
	}
	if result.PacketLossRate != 100.0 {
		t.Errorf("Expected PacketLossRate=100.0, got %f", result.PacketLossRate)
	}
	if result.ErrorMessage != "all packets lost" {
		t.Errorf("Expected 'all packets lost', got '%s'", result.ErrorMessage)
	}
}

// TestNewUDPProbeResultWithMetrics tests UDP probe result with full metrics
func TestNewUDPProbeResultWithMetrics(t *testing.T) {
	result := NewUDPProbeResultWithMetrics(true, 0.0, 1.5, 1.4, 0.05, 0.002, 10, 10, 10, "")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.PacketLossRate != 0.0 {
		t.Errorf("Expected PacketLossRate=0.0, got %f", result.PacketLossRate)
	}
	if result.RTTMs != 1.5 {
		t.Errorf("Expected RTTMs=1.5, got %f", result.RTTMs)
	}
	if result.RTTMedianMs != 1.4 {
		t.Errorf("Expected RTTMedianMs=1.4, got %f", result.RTTMedianMs)
	}
	if result.JitterMs != 0.05 {
		t.Errorf("Expected JitterMs=0.05, got %f", result.JitterMs)
	}
	if result.VarianceMs != 0.002 {
		t.Errorf("Expected VarianceMs=0.002, got %f", result.VarianceMs)
	}
	if result.SentPackets != 10 {
		t.Errorf("Expected SentPackets=10, got %d", result.SentPackets)
	}
	if result.ReceivedPackets != 10 {
		t.Errorf("Expected ReceivedPackets=10, got %d", result.ReceivedPackets)
	}
	if result.SampleCount != 10 {
		t.Errorf("Expected SampleCount=10, got %d", result.SampleCount)
	}
	if result.Timestamp == "" {
		t.Error("Expected non-empty Timestamp")
	}
}

// TestNewUDPProbeResultWithMetrics_WithError tests UDP probe with error
func TestNewUDPProbeResultWithMetrics_WithError(t *testing.T) {
	result := NewUDPProbeResultWithMetrics(false, 100.0, 0, 0, 0, 0, 10, 0, 10, "host unreachable")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Success {
		t.Error("Expected Success=false")
	}
	if result.ErrorMessage != "host unreachable" {
		t.Errorf("Expected 'host unreachable', got '%s'", result.ErrorMessage)
	}
}

// TestNewProbeResult tests generic probe result creation
func TestNewProbeResult(t *testing.T) {
	metrics := map[string]interface{}{
		"rtt_ms":      1.5,
		"packet_loss": 0.0,
	}
	result := NewProbeResult("tcp_ping", "localhost:8080", true, metrics, "")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Type != "tcp_ping" {
		t.Errorf("Expected Type='tcp_ping', got '%s'", result.Type)
	}
	if result.Target != "localhost:8080" {
		t.Errorf("Expected Target='localhost:8080', got '%s'", result.Target)
	}
	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.Metrics == nil {
		t.Error("Expected non-nil Metrics")
	}
	if result.ErrorMessage != "" {
		t.Errorf("Expected empty ErrorMessage, got '%s'", result.ErrorMessage)
	}
	if result.Timestamp == "" {
		t.Error("Expected non-empty Timestamp")
	}
}

// TestNewProbeResult_Failure tests generic probe result failure
func TestNewProbeResult_Failure(t *testing.T) {
	result := NewProbeResult("udp_ping", "localhost:9090", false, nil, "timeout")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Success {
		t.Error("Expected Success=false")
	}
	if result.Metrics != nil {
		t.Error("Expected nil Metrics for failure")
	}
	if result.ErrorMessage != "timeout" {
		t.Errorf("Expected 'timeout', got '%s'", result.ErrorMessage)
	}
}

// TestTCPProbeResult_ToGenericResult tests converting TCPProbeResult to ProbeResult
func TestTCPProbeResult_ToGenericResult(t *testing.T) {
	tcpResult := &TCPProbeResult{
		Success:      true,
		RTTMs:        2.5,
		ErrorMessage: "",
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	generic := tcpResult.ToGenericResult("localhost:80")
	if generic == nil {
		t.Fatal("Expected non-nil generic result")
	}
	if generic.Type != "tcp_ping" {
		t.Errorf("Expected Type='tcp_ping', got '%s'", generic.Type)
	}
	if generic.Target != "localhost:80" {
		t.Errorf("Expected Target='localhost:80', got '%s'", generic.Target)
	}
	if !generic.Success {
		t.Error("Expected Success=true")
	}
	if generic.Metrics == nil {
		t.Fatal("Expected non-nil Metrics")
	}
	if generic.Metrics["rtt_ms"] != 2.5 {
		t.Errorf("Expected rtt_ms=2.5, got %v", generic.Metrics["rtt_ms"])
	}
	if generic.Timestamp != tcpResult.Timestamp {
		t.Errorf("Expected Timestamp to be preserved")
	}
}

// TestTCPProbeResult_ToGenericResult_Failure tests converting failed TCPProbeResult
func TestTCPProbeResult_ToGenericResult_Failure(t *testing.T) {
	tcpResult := &TCPProbeResult{
		Success:      false,
		RTTMs:        0,
		ErrorMessage: "connection refused",
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	generic := tcpResult.ToGenericResult("localhost:443")
	if generic == nil {
		t.Fatal("Expected non-nil generic result")
	}
	if generic.Success {
		t.Error("Expected Success=false")
	}
	if generic.ErrorMessage != "connection refused" {
		t.Errorf("Expected 'connection refused', got '%s'", generic.ErrorMessage)
	}
}

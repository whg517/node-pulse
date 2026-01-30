package reporter

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/models"
)

// TestHeartbeatDataSerialization tests that HeartbeatData can be serialized to JSON correctly
func TestHeartbeatDataSerialization(t *testing.T) {
	// Arrange
	data := &HeartbeatData{
		NodeID:         "test-node-uuid-123",
		LatencyMs:      123.45,
		PacketLossRate: 0.5,
		JitterMs:       5.67,
		Timestamp:      "2026-01-30T12:34:56Z",
	}

	// Act
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal heartbeat data: %v", err)
	}

	// Assert
	var decoded HeartbeatData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal heartbeat data: %v", err)
	}

	if decoded.NodeID != data.NodeID {
		t.Errorf("Expected NodeID %s, got %s", data.NodeID, decoded.NodeID)
	}
	if decoded.LatencyMs != data.LatencyMs {
		t.Errorf("Expected LatencyMs %f, got %f", data.LatencyMs, decoded.LatencyMs)
	}
	if decoded.PacketLossRate != data.PacketLossRate {
		t.Errorf("Expected PacketLossRate %f, got %f", data.PacketLossRate, decoded.PacketLossRate)
	}
	if decoded.JitterMs != data.JitterMs {
		t.Errorf("Expected JitterMs %f, got %f", data.JitterMs, decoded.JitterMs)
	}
	if decoded.Timestamp != data.Timestamp {
		t.Errorf("Expected Timestamp %s, got %s", data.Timestamp, decoded.Timestamp)
	}
}

// TestHeartbeatDataJSONFormat tests the JSON field names match the API specification
func TestHeartbeatDataJSONFormat(t *testing.T) {
	// Arrange
	data := &HeartbeatData{
		NodeID:         "test-node-uuid",
		LatencyMs:      100.0,
		PacketLossRate: 1.0,
		JitterMs:       2.5,
		Timestamp:      "2026-01-30T12:34:56Z",
	}

	// Act
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal heartbeat data: %v", err)
	}

	// Assert - verify JSON field names
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Check required fields exist
	if _, ok := raw["node_id"]; !ok {
		t.Error("Missing required field: node_id")
	}
	if _, ok := raw["latency_ms"]; !ok {
		t.Error("Missing required field: latency_ms")
	}
	if _, ok := raw["packet_loss_rate"]; !ok {
		t.Error("Missing required field: packet_loss_rate")
	}
	if _, ok := raw["jitter_ms"]; !ok {
		t.Error("Missing required field: jitter_ms")
	}
	if _, ok := raw["timestamp"]; !ok {
		t.Error("Missing required field: timestamp")
	}
}

// TestNewHeartbeatData tests creating heartbeat data with current timestamp
func TestNewHeartbeatData(t *testing.T) {
	// Arrange & Act
	data := NewHeartbeatData("test-node-id", 100.0, 0.5, 5.0)

	// Assert
	if data.NodeID != "test-node-id" {
		t.Errorf("Expected NodeID 'test-node-id', got %s", data.NodeID)
	}
	if data.LatencyMs != 100.0 {
		t.Errorf("Expected LatencyMs 100.0, got %f", data.LatencyMs)
	}
	if data.PacketLossRate != 0.5 {
		t.Errorf("Expected PacketLossRate 0.5, got %f", data.PacketLossRate)
	}
	if data.JitterMs != 5.0 {
		t.Errorf("Expected JitterMs 5.0, got %f", data.JitterMs)
	}

	// Verify timestamp is valid ISO 8601 format
	_, err := time.Parse(time.RFC3339, data.Timestamp)
	if err != nil {
		t.Errorf("Timestamp is not valid ISO 8601 format: %v", err)
	}
}

// TestPulseAPIClientCreation tests creating a new Pulse API client
func TestPulseAPIClientCreation(t *testing.T) {
	// Arrange & Act
	client := NewPulseAPIClient("https://pulse.example.com", 5*time.Second)

	// Assert
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.serverURL != "https://pulse.example.com" {
		t.Errorf("Expected serverURL 'https://pulse.example.com', got %s", client.serverURL)
	}
	if client.timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", client.timeout)
	}
	if client.httpClient == nil {
		t.Error("Expected non-nil httpClient")
	}
}

// TestPulseAPIClientTLSSupport tests that the client supports TLS/HTTPS
func TestPulseAPIClientTLSSupport(t *testing.T) {
	// Arrange & Act
	client := NewPulseAPIClient("https://pulse.example.com", 5*time.Second)

	// Assert - verify TLS config is present
	transport := client.httpClient.Transport.(*http.Transport)
	if transport == nil {
		t.Fatal("Expected http.Transport, got nil")
	}

	if transport.TLSClientConfig == nil {
		t.Error("Expected TLSClientConfig to be non-nil for HTTPS support")
	}
}

// TestHeartbeatReporterCreation tests creating a new HeartbeatReporter
func TestHeartbeatReporterCreation(t *testing.T) {
	// Arrange
	apiClient := NewPulseAPIClient("https://pulse.example.com", 5*time.Second)
	mockScheduler := &mockProbeScheduler{}

	// Act
	reporter := NewHeartbeatReporter(apiClient, "test-node-id", mockScheduler)

	// Assert
	if reporter == nil {
		t.Fatal("Expected non-nil reporter")
	}
	if reporter.apiClient != apiClient {
		t.Error("Expected apiClient to match")
	}
	if reporter.nodeID != "test-node-id" {
		t.Errorf("Expected nodeID 'test-node-id', got %s", reporter.nodeID)
	}
	if reporter.reporting {
		t.Error("Expected reporting to be false initially")
	}
	if reporter.scheduler != mockScheduler {
		t.Error("Expected scheduler to match")
	}
}

// Mock scheduler for testing
type mockProbeScheduler struct {
	tcpResults []*models.TCPProbeResult
	udpResults []*models.UDPProbeResult
}

func (m *mockProbeScheduler) GetLatestResults() ([]*models.TCPProbeResult, []*models.UDPProbeResult) {
	return m.tcpResults, m.udpResults
}

// TestAggregateMetricsFromTCPProbes tests aggregating metrics from multiple successful TCP probes
func TestAggregateMetricsFromTCPProbes(t *testing.T) {
	// Arrange
	mockScheduler := &mockProbeScheduler{}
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", mockScheduler)

	// Create mock TCP probe results (3 successful probes)
	tcpResults := []*models.TCPProbeResult{
		{
			Success:        true,
			RTTMs:          100.0,
			PacketLossRate: 0.0,
			JitterMs:       2.0,
		},
		{
			Success:        true,
			RTTMs:          200.0,
			PacketLossRate: 1.0,
			JitterMs:       3.0,
		},
		{
			Success:        true,
			RTTMs:          150.0,
			PacketLossRate: 0.5,
			JitterMs:       2.5,
		},
	}

	// Act
	data := reporter.AggregateMetrics(tcpResults, nil)

	// Assert - averages should be (100+200+150)/3, (0+1+0.5)/3, (2+3+2.5)/3
	if data.NodeID != "test-node-id" {
		t.Errorf("Expected NodeID 'test-node-id', got %s", data.NodeID)
	}
	if data.LatencyMs != 150.0 { // (100 + 200 + 150) / 3
		t.Errorf("Expected LatencyMs 150.0, got %f", data.LatencyMs)
	}
	if data.PacketLossRate != 0.5 { // (0 + 1 + 0.5) / 3
		t.Errorf("Expected PacketLossRate 0.5, got %f", data.PacketLossRate)
	}
	if data.JitterMs != 2.5 { // (2 + 3 + 2.5) / 3
		t.Errorf("Expected JitterMs 2.5, got %f", data.JitterMs)
	}
}

// TestAggregateMetricsNoProbes tests aggregating when no probe results are available
func TestAggregateMetricsNoProbes(t *testing.T) {
	// Arrange
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", &mockProbeScheduler{})

	// Act
	data := reporter.AggregateMetrics(nil, nil)

	// Assert - default values should be used
	if data.NodeID != "test-node-id" {
		t.Errorf("Expected NodeID 'test-node-id', got %s", data.NodeID)
	}
	if data.LatencyMs != 0 {
		t.Errorf("Expected LatencyMs 0 (default), got %f", data.LatencyMs)
	}
	if data.PacketLossRate != 100 {
		t.Errorf("Expected PacketLossRate 100 (default - all failed), got %f", data.PacketLossRate)
	}
	if data.JitterMs != 0 {
		t.Errorf("Expected JitterMs 0 (default), got %f", data.JitterMs)
	}
}

// TestAggregateMetricsPartialFailures tests aggregating when some probes failed
func TestAggregateMetricsPartialFailures(t *testing.T) {
	// Arrange
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", &mockProbeScheduler{})

	// Create mock TCP probe results (2 successful, 1 failed)
	tcpResults := []*models.TCPProbeResult{
		{
			Success:        true,
			RTTMs:          100.0,
			PacketLossRate: 0.0,
			JitterMs:       2.0,
		},
		{
			Success:        false, // Failed probe - should be excluded
		},
		{
			Success:        true,
			RTTMs:          200.0,
			PacketLossRate: 1.0,
			JitterMs:       3.0,
		},
	}

	// Act
	data := reporter.AggregateMetrics(tcpResults, nil)

	// Assert - only successful probes should be averaged
	if data.LatencyMs != 150.0 { // (100 + 200) / 2 (excluding failed probe)
		t.Errorf("Expected LatencyMs 150.0, got %f", data.LatencyMs)
	}
	if data.PacketLossRate != 0.5 { // (0 + 1) / 2
		t.Errorf("Expected PacketLossRate 0.5, got %f", data.PacketLossRate)
	}
	if data.JitterMs != 2.5 { // (2 + 3) / 2
		t.Errorf("Expected JitterMs 2.5, got %f", data.JitterMs)
	}
}

// TestAggregateMetricsUDPProbes tests aggregating metrics from UDP probes
func TestAggregateMetricsUDPProbes(t *testing.T) {
	// Arrange
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", &mockProbeScheduler{})

	// Create mock UDP probe results
	udpResults := []*models.UDPProbeResult{
		{
			Success:        true,
			RTTMs:          120.0,
			PacketLossRate: 2.0,
			JitterMs:       4.0,
		},
		{
			Success:        true,
			RTTMs:          180.0,
			PacketLossRate: 3.0,
			JitterMs:       6.0,
		},
	}

	// Act
	data := reporter.AggregateMetrics(nil, udpResults)

	// Assert
	if data.LatencyMs != 150.0 { // (120 + 180) / 2
		t.Errorf("Expected LatencyMs 150.0, got %f", data.LatencyMs)
	}
	if data.PacketLossRate != 2.5 { // (2 + 3) / 2
		t.Errorf("Expected PacketLossRate 2.5, got %f", data.PacketLossRate)
	}
	if data.JitterMs != 5.0 { // (4 + 6) / 2
		t.Errorf("Expected JitterMs 5.0, got %f", data.JitterMs)
	}
}

// TestAggregateMetricsMixedProbes tests aggregating metrics from both TCP and UDP probes
func TestAggregateMetricsMixedProbes(t *testing.T) {
	// Arrange
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", &mockProbeScheduler{})

	// Create mock TCP and UDP probe results
	tcpResults := []*models.TCPProbeResult{
		{
			Success:        true,
			RTTMs:          100.0,
			PacketLossRate: 0.0,
			JitterMs:       2.0,
		},
	}

	udpResults := []*models.UDPProbeResult{
		{
			Success:        true,
			RTTMs:          200.0,
			PacketLossRate: 1.0,
			JitterMs:       3.0,
		},
	}

	// Act
	data := reporter.AggregateMetrics(tcpResults, udpResults)

	// Assert - should average across both TCP and UDP results
	if data.LatencyMs != 150.0 { // (100 + 200) / 2
		t.Errorf("Expected LatencyMs 150.0, got %f", data.LatencyMs)
	}
	if data.PacketLossRate != 0.5 { // (0 + 1) / 2
		t.Errorf("Expected PacketLossRate 0.5, got %f", data.PacketLossRate)
	}
	if data.JitterMs != 2.5 { // (2 + 3) / 2
		t.Errorf("Expected JitterMs 2.5, got %f", data.JitterMs)
	}
}

// TestAggregateMetricsAllFailed tests aggregating when all probes failed
func TestAggregateMetricsAllFailed(t *testing.T) {
	// Arrange
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", &mockProbeScheduler{})

	// Create mock TCP probe results (all failed)
	tcpResults := []*models.TCPProbeResult{
		{Success: false},
		{Success: false},
		{Success: false},
	}

	// Act
	data := reporter.AggregateMetrics(tcpResults, nil)

	// Assert - should use default values for all failed probes
	if data.LatencyMs != 0 {
		t.Errorf("Expected LatencyMs 0 (default), got %f", data.LatencyMs)
	}
	if data.PacketLossRate != 100 {
		t.Errorf("Expected PacketLossRate 100 (default - all failed), got %f", data.PacketLossRate)
	}
	if data.JitterMs != 0 {
		t.Errorf("Expected JitterMs 0 (default), got %f", data.JitterMs)
	}
}

// TestStartReporting tests starting the heartbeat reporter
func TestStartReporting(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger(&config.Config{
		LogLevel:     "INFO",
		LogFile:      "/tmp/test-reporter.log",
		LogMaxSize:   10,
		LogMaxAge:    7,
		LogMaxBackups: 3,
		LogCompress:  false,
		LogToConsole: false,
	})
	defer logger.Close()

	// Arrange
	mockScheduler := &mockProbeScheduler{}
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", mockScheduler)

	// Act
	ctx := context.Background()
	reporter.StartReporting(ctx)
	defer reporter.StopReporting() // Ensure cleanup

	// Wait a moment for goroutine to start
	time.Sleep(50 * time.Millisecond)

	// Assert
	reporter.mu.Lock()
	isReporting := reporter.reporting
	reporter.mu.Unlock()

	if !isReporting {
		t.Error("Expected reporting to be true after StartReporting")
	}
	if reporter.ticker == nil {
		t.Error("Expected ticker to be initialized")
	}
}

// TestStopReporting tests stopping the heartbeat reporter
func TestStopReporting(t *testing.T) {
	// Arrange
	mockScheduler := &mockProbeScheduler{}
	reporter := NewHeartbeatReporter(NewPulseAPIClient("https://pulse.example.com", 5*time.Second), "test-node-id", mockScheduler)
	ctx := context.Background()
	reporter.StartReporting(ctx)

	// Wait a moment
	time.Sleep(50 * time.Millisecond)

	// Act
	reporter.StopReporting()

	// Wait for goroutine to finish
	time.Sleep(50 * time.Millisecond)

	// Assert
	reporter.mu.Lock()
	isReporting := reporter.reporting
	reporter.mu.Unlock()

	if isReporting {
		t.Error("Expected reporting to be false after StopReporting")
	}
}

// TestReportWithRetrySuccess tests successful heartbeat reporting without retries
func TestReportWithRetrySuccess(t *testing.T) {
	// Arrange - start mock Pulse server
	mockServer := NewMockPulseServer()
	defer mockServer.Close()

	apiClient := NewPulseAPIClient(mockServer.GetURL(), 5*time.Second)
	mockScheduler := &mockProbeScheduler{
		tcpResults: []*models.TCPProbeResult{
			{Success: true, RTTMs: 100.0, PacketLossRate: 0.0, JitterMs: 2.0},
		},
	}
	reporter := NewHeartbeatReporter(apiClient, "test-node-uuid", mockScheduler)

	// Act - trigger report
	reporter.reportWithRetry()

	// Assert - server should receive exactly 1 heartbeat (no retries)
	if mockServer.GetHeartbeatCount() != 1 {
		t.Errorf("Expected 1 heartbeat (no retries), got %d", mockServer.GetHeartbeatCount())
	}
}

// TestReportWithRetryFailure tests retry mechanism on failure
func TestReportWithRetryFailure(t *testing.T) {
	// Arrange - start mock Pulse server configured to fail
	mockServer := NewMockPulseServer()
	mockServer.SetResponseStatusCode(http.StatusInternalServerError)
	defer mockServer.Close()

	apiClient := NewPulseAPIClient(mockServer.GetURL(), 5*time.Second)
	mockScheduler := &mockProbeScheduler{
		tcpResults: []*models.TCPProbeResult{
			{Success: true, RTTMs: 100.0, PacketLossRate: 0.0, JitterMs: 2.0},
		},
	}
	reporter := NewHeartbeatReporter(apiClient, "test-node-uuid", mockScheduler)

	// Act - trigger report (should retry 3 times)
	reporter.reportWithRetry()

	// Assert - server should receive 3 requests (MaxRetries)
	if mockServer.GetHeartbeatCount() != 3 {
		t.Errorf("Expected 3 heartbeat requests (MaxRetries), got %d", mockServer.GetHeartbeatCount())
	}
}

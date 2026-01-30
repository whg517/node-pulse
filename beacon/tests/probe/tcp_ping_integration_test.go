package probe

import (
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/probe"
)

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// initTestLogger initializes logger for integration tests
func initTestLogger(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")
	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 10,
		LogCompress:   false,
		LogToConsole:  false,
	}
	if err := logger.InitLogger(cfg); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
}

// TestIntegration_TCPProbeWithRealServer tests complete TCP probe flow with real server
func TestIntegration_TCPProbeWithRealServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start a test TCP server
	server := startTestTCPServer(t, "localhost:19001")
	defer server.Close()

	// Create probe configuration
	cfg := probe.TCPProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           19001,
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          10,
	}

	// Create pinger
	pinger := probe.NewTCPPinger(cfg)

	// Execute batch probes with core metrics
	result, err := pinger.ExecuteBatch(cfg.Count)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	// Verify probe succeeded
	if !result.Success {
		t.Errorf("Probe failed: %s", result.ErrorMessage)
	}
	if result.RTTMs <= 0 {
		t.Errorf("Expected RTT > 0, got %f", result.RTTMs)
	}
	if result.SampleCount != cfg.Count {
		t.Errorf("Expected SampleCount=%d, got %d", cfg.Count, result.SampleCount)
	}
	if result.Timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}

	// Verify core metrics are calculated
	if result.RTTMedianMs <= 0 {
		t.Errorf("Expected RTTMedianMs > 0, got %f", result.RTTMedianMs)
	}
	if result.JitterMs < 0 {
		t.Errorf("Expected JitterMs >= 0, got %f", result.JitterMs)
	}
	if result.VarianceMs < 0 {
		t.Errorf("Expected VarianceMs >= 0, got %f", result.VarianceMs)
	}

	if result.PacketLossRate != 0 {
		t.Errorf("Expected 0%% packet loss, got %.2f%%", result.PacketLossRate)
	}
}

// TestIntegration_ProbeSchedulerWithMultipleTargets tests scheduler with multiple probe targets
func TestIntegration_ProbeSchedulerWithMultipleTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Initialize logger for this test
	initTestLogger(t)
	defer logger.Close()

	// Start multiple test servers
	server1 := startTestTCPServer(t, "localhost:19002")
	defer server1.Close()

	server2 := startTestTCPServer(t, "localhost:19003")
	defer server2.Close()

	// Create probe configs
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     19002,
			TimeoutSeconds:  5,
			Interval: 60,
			Count:    10,
		},
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     19003,
			TimeoutSeconds:  5,
			Interval: 60,
			Count:    10,
		},
	}

	// Create scheduler
	scheduler, err := probe.NewProbeScheduler(probeConfigs)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Verify probe count
	if scheduler.GetProbeCount() != 2 {
		t.Errorf("Expected 2 probes, got %d", scheduler.GetProbeCount())
	}

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for probes to execute
	time.Sleep(2 * time.Second)

	// Stop scheduler
	scheduler.Stop()

	// Verify scheduler stopped
	if scheduler.IsRunning() {
		t.Error("Expected scheduler to be stopped")
	}
}

// TestIntegration_ProbeSchedulerWithInvalidConfig tests scheduler error handling
func TestIntegration_ProbeSchedulerWithInvalidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create invalid probe configs
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     99999, // Invalid port
			TimeoutSeconds:  5,
			Interval: 60,
			Count:    10,
		},
	}

	// Create scheduler - should fail validation
	_, err := probe.NewProbeScheduler(probeConfigs)
	if err == nil {
		t.Error("Expected error when creating scheduler with invalid config")
	}
}

// TestIntegration_ProbeSchedulerWithInsufficientCount tests scheduler count validation
func TestIntegration_ProbeSchedulerWithInsufficientCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create probe configs with count < 10 (should fail core metrics calculation)
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     19001,
			TimeoutSeconds:  5,
			Interval: 60,
			Count:    5, // Invalid: must be ≥ 10 for core metrics
		},
	}

	// Create scheduler - should fail with count ≥ 10 validation error
	_, err := probe.NewProbeScheduler(probeConfigs)
	if err == nil {
		t.Error("Expected error when creating scheduler with count < 10")
	}

	// Verify error message mentions the count requirement
	if err != nil && !containsString(err.Error(), "must be ≥ 10") {
		t.Errorf("Expected error message to contain 'must be ≥ 10', got: %v", err)
	}

	t.Logf("✅ Count validation working correctly: %v", err)
}

// TestIntegration_ProbeSchedulerWithMixedTargets tests scheduler with valid and invalid targets
func TestIntegration_ProbeSchedulerWithMixedTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start one valid server
	server := startTestTCPServer(t, "localhost:19004")
	defer server.Close()

	// Create probe configs with mix of valid and invalid targets
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     19004, // Valid
			TimeoutSeconds:  2,
			Interval: 60,
			Count:    10,
		},
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     19999, // Invalid (port not listening)
			TimeoutSeconds:  1,
			Interval: 60,
			Count:    10,
		},
	}

	// Create scheduler
	scheduler, err := probe.NewProbeScheduler(probeConfigs)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for probes to execute
	time.Sleep(2 * time.Second)

	// Stop scheduler
	scheduler.Stop()

	// Verify scheduler executed without crashing
	if !scheduler.IsRunning() {
		t.Log("Scheduler stopped successfully after handling mixed targets")
	}
}

// TestIntegration_ProbeGracefulShutdown tests graceful shutdown of scheduler
func TestIntegration_ProbeGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start test server
	server := startTestTCPServer(t, "localhost:19005")
	defer server.Close()

	// Create probe config with minimum valid interval for testing
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     19005,
			TimeoutSeconds:  5,
			Interval: 60, // Minimum valid interval (60 seconds)
			Count:    10,
		},
	}

	// Create and start scheduler
	scheduler, err := probe.NewProbeScheduler(probeConfigs)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Let scheduler run for a bit
	time.Sleep(2 * time.Second)

	// Send interrupt signal to simulate graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	// Stop scheduler in background
	go func() {
		time.Sleep(500 * time.Millisecond)
		scheduler.Stop()
	}()

	// Wait a bit for shutdown
	time.Sleep(1 * time.Second)

	// Verify scheduler stopped gracefully
	if scheduler.IsRunning() {
		t.Error("Expected scheduler to be stopped after graceful shutdown")
	}
}

// TestIntegration_ExecuteProbeNow tests manual probe execution
func TestIntegration_ExecuteProbeNow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start test server
	server := startTestTCPServer(t, "localhost:19006")
	defer server.Close()

	// Create scheduler
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "tcp_ping",
			Target:   "localhost",
			Port:     19006,
			TimeoutSeconds:  5,
			Interval: 60,
			Count:    10,
		},
	}

	scheduler, err := probe.NewProbeScheduler(probeConfigs)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Execute probe manually without starting scheduler
	result, err := scheduler.ExecuteProbeNow(0)
	if err != nil {
		t.Fatalf("ExecuteProbeNow failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful probe, got error: %s", result.ErrorMessage)
	}

	if result.RTTMs <= 0 {
		t.Errorf("Expected RTT > 0, got %f", result.RTTMs)
	}

	// Test invalid index
	_, err = scheduler.ExecuteProbeNow(99)
	if err == nil {
		t.Error("Expected error for invalid probe index")
	}
}

// Helper function to start a test TCP server
func startTestTCPServer(t *testing.T, addr string) *testTCPServer {
	t.Helper()
	server := &testTCPServer{addr: addr}
	go server.start()
	time.Sleep(100 * time.Millisecond) // Give server time to start
	return server
}

type testTCPServer struct {
	addr string
	lis  net.Listener
}

func (s *testTCPServer) start() {
	var err error
	s.lis, err = net.Listen("tcp", s.addr)
	if err != nil {
		return
	}
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			break
		}
		conn.Close()
	}
}

func (s *testTCPServer) Close() {
	if s.lis != nil {
		s.lis.Close()
	}
}

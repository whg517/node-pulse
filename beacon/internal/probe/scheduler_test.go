package probe

import (
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/whg517/node-pulse/beacon/internal/config"
	"github.com/whg517/node-pulse/beacon/internal/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize a simple test logger to avoid nil pointer panics
	logger.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	os.Exit(m.Run())
}

// startTCPServerForScheduler starts a test TCP server for scheduler tests
func startTCPServerForScheduler(t *testing.T, addr string) net.Listener {
	t.Helper()
	server, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	return server
}

// makeValidProbeConfig creates a valid TCP probe config
func makeValidTCPProbeConfig(target string, port int) config.ProbeConfig {
	return config.ProbeConfig{
		Type:           "tcp_ping",
		Target:         target,
		Port:           port,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          10,
	}
}

// makeValidUDPProbeConfig creates a valid UDP probe config
func makeValidUDPProbeConfig(target string, port int) config.ProbeConfig {
	return config.ProbeConfig{
		Type:           "udp_ping",
		Target:         target,
		Port:           port,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          10,
	}
}

func makeValidMTRProbeConfig(target string) config.ProbeConfig {
	return config.ProbeConfig{
		Type:           "mtr",
		Target:         target,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          3,
		MaxHops:        8,
		PacketSize:     128,
	}
}

// TestNewProbeScheduler_Empty tests creating a scheduler with no probes
func TestNewProbeScheduler_Empty(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}
	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}
	if scheduler.GetProbeCount() != 0 {
		t.Errorf("Expected 0 probes, got %d", scheduler.GetProbeCount())
	}
}

// TestNewProbeScheduler_WithTCPProbe tests creating a scheduler with TCP probe
func TestNewProbeScheduler_WithTCPProbe(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18770")
	defer func() { _ = server.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18770)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}
	if scheduler.GetProbeCount() != 1 {
		t.Errorf("Expected 1 probe, got %d", scheduler.GetProbeCount())
	}
}

// TestNewProbeScheduler_WithUDPProbe tests creating a scheduler with UDP probe
func TestNewProbeScheduler_WithUDPProbe(t *testing.T) {
	cfg := makeValidUDPProbeConfig("localhost", 18771)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}
	if scheduler.GetProbeCount() != 1 {
		t.Errorf("Expected 1 probe, got %d", scheduler.GetProbeCount())
	}
}

func TestNewProbeScheduler_WithMTRProbe(t *testing.T) {
	cfg := makeValidMTRProbeConfig("localhost")
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}
	if scheduler.GetProbeCount() != 1 {
		t.Errorf("Expected 1 probe, got %d", scheduler.GetProbeCount())
	}
}

// TestNewProbeScheduler_MixedProbes tests creating a scheduler with mixed probe types
func TestNewProbeScheduler_MixedProbes(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18772")
	defer func() { _ = server.Close() }()

	probes := []config.ProbeConfig{
		makeValidTCPProbeConfig("localhost", 18772),
		makeValidUDPProbeConfig("localhost", 18773),
		makeValidMTRProbeConfig("localhost"),
	}
	scheduler, err := NewProbeScheduler(probes)
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}
	if scheduler.GetProbeCount() != 3 {
		t.Errorf("Expected 3 probes, got %d", scheduler.GetProbeCount())
	}
}

// TestNewProbeScheduler_InvalidTCPConfig tests that invalid TCP config returns error
func TestNewProbeScheduler_InvalidTCPConfig(t *testing.T) {
	// Invalid: port 0
	cfg := config.ProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           0,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          10,
	}
	_, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err == nil {
		t.Error("Expected error for invalid probe config")
	}
}

// TestNewProbeScheduler_InvalidUDPConfig tests that invalid UDP config returns error
func TestNewProbeScheduler_InvalidUDPConfig(t *testing.T) {
	// Invalid: port 0
	cfg := config.ProbeConfig{
		Type:           "udp_ping",
		Target:         "localhost",
		Port:           0,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          10,
	}
	_, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err == nil {
		t.Error("Expected error for invalid probe config")
	}
}

// TestNewProbeScheduler_CountTooLow_TCP tests that count < 10 returns error for TCP
func TestNewProbeScheduler_CountTooLow_TCP(t *testing.T) {
	cfg := config.ProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           18774,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          5, // Below minimum of 10
	}
	_, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err == nil {
		t.Error("Expected error for count < 10")
	}
}

// TestNewProbeScheduler_CountTooLow_UDP tests that count < 10 returns error for UDP
func TestNewProbeScheduler_CountTooLow_UDP(t *testing.T) {
	cfg := config.ProbeConfig{
		Type:           "udp_ping",
		Target:         "localhost",
		Port:           18775,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          5, // Below minimum of 10
	}
	_, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err == nil {
		t.Error("Expected error for count < 10")
	}
}

// TestProbeScheduler_Start_NoProbes tests starting scheduler with no probes
func TestProbeScheduler_Start_NoProbes(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Should not be "running" in background goroutine sense since no probes
	// But should not error
	scheduler.Stop()
}

// TestProbeScheduler_Start_AlreadyRunning tests double-start returns error
func TestProbeScheduler_Start_AlreadyRunning(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18776")
	defer func() { _ = server.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18776)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("First Start() failed: %v", err)
	}
	defer scheduler.Stop()

	// Manually set running state to true to test double-start
	scheduler.mu.Lock()
	scheduler.running = true
	scheduler.mu.Unlock()

	if err := scheduler.Start(); err == nil {
		t.Error("Expected error when starting already running scheduler")
	}
}

// TestProbeScheduler_IsRunning tests IsRunning function
func TestProbeScheduler_IsRunning(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	if scheduler.IsRunning() {
		t.Error("Expected scheduler to not be running initially")
	}

	// Start with no probes - should set running=true, then return early
	// Note: with no probes, Start returns early without starting goroutine
	_ = scheduler.Start()
	// Running is set to true, but no goroutine
	// Stop won't work normally since there's no goroutine, but IsRunning should still work
}

// TestProbeScheduler_Stop_NotRunning tests stopping a non-running scheduler
func TestProbeScheduler_Stop_NotRunning(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}
	// Should not panic or error
	scheduler.Stop()
}

// TestProbeScheduler_GetProbeCount tests GetProbeCount
func TestProbeScheduler_GetProbeCount(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18777")
	defer func() { _ = server.Close() }()

	probes := []config.ProbeConfig{
		makeValidTCPProbeConfig("localhost", 18777),
		makeValidUDPProbeConfig("localhost", 18778),
	}
	scheduler, err := NewProbeScheduler(probes)
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}
	if scheduler.GetProbeCount() != 2 {
		t.Errorf("Expected 2, got %d", scheduler.GetProbeCount())
	}
}

// TestProbeScheduler_ExecuteProbeNow tests ExecuteProbeNow
func TestProbeScheduler_ExecuteProbeNow(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18779")
	defer func() { _ = server.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18779)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	result, err := scheduler.ExecuteProbeNow(0)
	if err != nil {
		t.Fatalf("ExecuteProbeNow failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

// TestProbeScheduler_ExecuteProbeNow_OutOfRange tests ExecuteProbeNow out of range
func TestProbeScheduler_ExecuteProbeNow_OutOfRange(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	_, err = scheduler.ExecuteProbeNow(0)
	if err == nil {
		t.Error("Expected error for out of range index")
	}

	_, err = scheduler.ExecuteProbeNow(-1)
	if err == nil {
		t.Error("Expected error for negative index")
	}
}

// TestProbeScheduler_GetLatestResults_Empty tests GetLatestResults with no results
func TestProbeScheduler_GetLatestResults_Empty(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	tcpResults, udpResults := scheduler.GetLatestResults()
	if tcpResults == nil {
		t.Error("Expected non-nil TCP results slice")
	}
	if udpResults == nil {
		t.Error("Expected non-nil UDP results slice")
	}
	if len(tcpResults) != 0 || len(udpResults) != 0 {
		t.Errorf("Expected empty results, got TCP=%d, UDP=%d", len(tcpResults), len(udpResults))
	}
}

// TestProbeScheduler_UpdateProbeInterval tests UpdateProbeInterval
func TestProbeScheduler_UpdateProbeInterval(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18780")
	defer func() { _ = server.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18780)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	// Start to initialize interval
	_ = scheduler.Start()
	defer scheduler.Stop()

	// Update interval with multiplier 2
	if err := scheduler.UpdateProbeInterval(2); err != nil {
		t.Fatalf("UpdateProbeInterval failed: %v", err)
	}
}

// TestProbeScheduler_UpdateProbeInterval_InvalidMultiplier tests invalid multiplier
func TestProbeScheduler_UpdateProbeInterval_InvalidMultiplier(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	// Test invalid multipliers
	if err := scheduler.UpdateProbeInterval(0); err == nil {
		t.Error("Expected error for multiplier 0")
	}
	if err := scheduler.UpdateProbeInterval(11); err == nil {
		t.Error("Expected error for multiplier > 10")
	}
	if err := scheduler.UpdateProbeInterval(-1); err == nil {
		t.Error("Expected error for negative multiplier")
	}
}

// TestProbeScheduler_GetInterval tests GetInterval
func TestProbeScheduler_GetInterval(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18781")
	defer func() { _ = server.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18781)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	// Before start, interval is 0
	interval := scheduler.GetInterval()
	if interval != 0 {
		t.Logf("Interval before start: %v", interval)
	}

	// Start to set interval
	_ = scheduler.Start()
	defer scheduler.Stop()

	interval = scheduler.GetInterval()
	if interval <= 0 {
		t.Errorf("Expected positive interval after start, got %v", interval)
	}
}

// TestProbeScheduler_ReloadConfig tests ReloadConfig
func TestProbeScheduler_ReloadConfig(t *testing.T) {
	server1 := startTCPServerForScheduler(t, "localhost:18782")
	defer func() { _ = server1.Close() }()
	server2 := startTCPServerForScheduler(t, "localhost:18783")
	defer func() { _ = server2.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18782)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	// Reload with different config
	newCfg := makeValidTCPProbeConfig("localhost", 18783)
	if err := scheduler.ReloadConfig([]config.ProbeConfig{newCfg}); err != nil {
		t.Fatalf("ReloadConfig failed: %v", err)
	}

	if scheduler.GetProbeCount() != 1 {
		t.Errorf("Expected 1 probe after reload, got %d", scheduler.GetProbeCount())
	}
}

// TestProbeScheduler_ReloadConfig_Empty tests ReloadConfig with empty config
func TestProbeScheduler_ReloadConfig_Empty(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18784")
	defer func() { _ = server.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18784)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	if err := scheduler.ReloadConfig([]config.ProbeConfig{}); err != nil {
		t.Fatalf("ReloadConfig failed: %v", err)
	}

	if scheduler.GetProbeCount() != 0 {
		t.Errorf("Expected 0 probes after reload, got %d", scheduler.GetProbeCount())
	}
}

// TestProbeScheduler_ReloadConfig_InvalidConfig tests ReloadConfig with invalid config
func TestProbeScheduler_ReloadConfig_InvalidConfig(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	// Invalid config - port 0
	invalidCfg := config.ProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           0,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          10,
	}
	if err := scheduler.ReloadConfig([]config.ProbeConfig{invalidCfg}); err == nil {
		t.Error("Expected error for invalid config")
	}
}

// TestProbeScheduler_ReloadConfig_CountTooLow tests ReloadConfig with count too low
func TestProbeScheduler_ReloadConfig_CountTooLow(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	lowCountCfg := config.ProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           18785,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          5,
	}
	if err := scheduler.ReloadConfig([]config.ProbeConfig{lowCountCfg}); err == nil {
		t.Error("Expected error for count < 10")
	}
}

// TestProbeScheduler_ReloadConfig_WithUDP tests ReloadConfig with UDP probes
func TestProbeScheduler_ReloadConfig_WithUDP(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	udpCfg := makeValidUDPProbeConfig("localhost", 18786)
	if err := scheduler.ReloadConfig([]config.ProbeConfig{udpCfg}); err != nil {
		t.Fatalf("ReloadConfig with UDP failed: %v", err)
	}

	if scheduler.GetProbeCount() != 1 {
		t.Errorf("Expected 1 probe after UDP reload, got %d", scheduler.GetProbeCount())
	}
}

// TestProbeScheduler_ReloadConfig_InvalidUDP tests ReloadConfig with invalid UDP config
func TestProbeScheduler_ReloadConfig_InvalidUDP(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	invalidUDPCfg := config.ProbeConfig{
		Type:           "udp_ping",
		Target:         "localhost",
		Port:           0, // invalid
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          10,
	}
	if err := scheduler.ReloadConfig([]config.ProbeConfig{invalidUDPCfg}); err == nil {
		t.Error("Expected error for invalid UDP config")
	}
}

// TestProbeScheduler_ReloadConfig_UDPCountTooLow tests ReloadConfig with UDP count too low
func TestProbeScheduler_ReloadConfig_UDPCountTooLow(t *testing.T) {
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	lowCountUDPCfg := config.ProbeConfig{
		Type:           "udp_ping",
		Target:         "localhost",
		Port:           18787,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          5, // Below minimum
	}
	if err := scheduler.ReloadConfig([]config.ProbeConfig{lowCountUDPCfg}); err == nil {
		t.Error("Expected error for UDP count < 10")
	}
}

// TestProbeScheduler_StartStop_WithProbes tests full start/stop cycle
func TestProbeScheduler_StartStop_WithProbes(t *testing.T) {
	server := startTCPServerForScheduler(t, "localhost:18788")
	defer func() { _ = server.Close() }()

	cfg := makeValidTCPProbeConfig("localhost", 18788)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("Expected scheduler to be running after Start")
	}

	// Wait a moment for the initial probe execution
	time.Sleep(200 * time.Millisecond)

	scheduler.Stop()

	if scheduler.IsRunning() {
		t.Error("Expected scheduler to not be running after Stop")
	}
}

// TestProbeScheduler_StartWithUDPFirst tests start with UDP as first probe
func TestProbeScheduler_StartWithUDPFirst(t *testing.T) {
	cfg := makeValidUDPProbeConfig("localhost", 18789)
	scheduler, err := NewProbeScheduler([]config.ProbeConfig{cfg})
	if err != nil {
		t.Fatalf("NewProbeScheduler failed: %v", err)
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer scheduler.Stop()

	interval := scheduler.GetInterval()
	if interval != 60*time.Second {
		t.Errorf("Expected 60s interval from UDP config, got %v", interval)
	}
}

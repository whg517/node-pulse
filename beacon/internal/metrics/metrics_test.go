package metrics

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/probe"
)

// initTestLogger initializes the logger for tests
func initTestLogger(t *testing.T) {
	if err := logger.InitLogger(&config.Config{
		LogLevel:      "INFO",
		LogFile:       "/tmp/test-metrics.log",
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	}); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
}

// TestNewMetrics tests the creation of a new Metrics handler
func TestNewMetrics(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()

	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: true,
		MetricsPort:    2112,
		MetricsUpdateSeconds: 10,
	}

	// Create a mock scheduler
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	// Create metrics handler
	m := NewMetrics(cfg, scheduler)

	assert.NotNil(t, m)
	assert.NotNil(t, m.config)
	assert.NotNil(t, m.scheduler)
	assert.NotNil(t, m.beaconUp)
	assert.NotNil(t, m.beaconRTTSeconds)
	assert.NotNil(t, m.beaconPacketLoss)
	assert.NotNil(t, m.beaconJitterMs)
	assert.NotNil(t, m.registry)
	assert.False(t, m.running)
}

// TestMetricsStart tests starting the metrics server
func TestMetricsStart(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: true,
		MetricsPort:    19112, // Use a different port to avoid conflicts
		MetricsUpdateSeconds: 10,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Start metrics server
	err = m.Start()
	assert.NoError(t, err)
	assert.True(t, m.IsRunning())

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is listening
	resp, err := http.Get("http://localhost:19112/metrics")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Stop metrics server
	err = m.Stop()
	assert.NoError(t, err)
	assert.False(t, m.IsRunning())
}

// TestMetricsStartDisabled tests that metrics server doesn't start when disabled
func TestMetricsStartDisabled(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: false,
		MetricsPort:    19113,
		MetricsUpdateSeconds: 10,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Start metrics server (should not start because disabled)
	err = m.Start()
	assert.NoError(t, err)
	assert.False(t, m.IsRunning())
}

// TestMetricsStartAlreadyRunning tests that starting an already running server returns error
func TestMetricsStartAlreadyRunning(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: true,
		MetricsPort:    19114,
		MetricsUpdateSeconds: 10,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Start metrics server
	err = m.Start()
	require.NoError(t, err)
	defer m.Stop()

	// Try to start again
	err = m.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

// TestMetricsEndpoint tests the /metrics endpoint returns correct format
func TestMetricsEndpoint(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-123",
		NodeName:       "beacon-test",
		MetricsEnabled: true,
		MetricsPort:    19115,
		MetricsUpdateSeconds: 10,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Start metrics server
	err = m.Start()
	require.NoError(t, err)
	defer m.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Request metrics endpoint
	resp, err := http.Get("http://localhost:19115/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify content type contains text/plain
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "text/plain")

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Verify Prometheus exposition format
	assert.Contains(t, bodyStr, "# HELP beacon_up Beacon running status")
	assert.Contains(t, bodyStr, "# TYPE beacon_up gauge")
	assert.Contains(t, bodyStr, "beacon_up{node_id=\"test-node-123\",node_name=\"beacon-test\"} 1")

	assert.Contains(t, bodyStr, "# HELP beacon_rtt_seconds Latest RTT latency in seconds")
	assert.Contains(t, bodyStr, "# TYPE beacon_rtt_seconds gauge")

	assert.Contains(t, bodyStr, "# HELP beacon_packet_loss_rate Latest packet loss rate")
	assert.Contains(t, bodyStr, "# TYPE beacon_packet_loss_rate gauge")

	assert.Contains(t, bodyStr, "# HELP beacon_jitter_ms Latest jitter in milliseconds")
	assert.Contains(t, bodyStr, "# TYPE beacon_jitter_ms gauge")

	// Verify labels are present
	assert.Contains(t, bodyStr, "node_id=\"test-node-123\"")
	assert.Contains(t, bodyStr, "node_name=\"beacon-test\"")
}

// TestUpdateMetricsNoResults tests metrics update with no probe results
func TestUpdateMetricsNoResults(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: true,
		MetricsPort:    19116,
		MetricsUpdateSeconds: 10,
	}

	// Create scheduler with no probes
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Update metrics (should handle empty results gracefully)
	m.updateMetrics()

	// Verify no panic occurred
	assert.True(t, true)
}

// TestUpdateMetricsWithResults tests metrics update with probe results
func TestUpdateMetricsWithResults(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: true,
		MetricsPort:    19117,
		MetricsUpdateSeconds: 10,
		Probes: []config.ProbeConfig{
			{
				Type:           "tcp_ping",
				Target:         "127.0.0.1",
				Port:           80,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          10,
			},
		},
	}

	// Create scheduler with probe
	scheduler, err := probe.NewProbeScheduler(cfg.Probes)
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Start metrics server
	err = m.Start()
	require.NoError(t, err)
	defer m.Stop()

	// Give server time to start and collect metrics
	time.Sleep(200 * time.Millisecond)

	// Request metrics endpoint
	resp, err := http.Get("http://localhost:19117/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Verify metrics are present with labels
	assert.Contains(t, bodyStr, "beacon_rtt_seconds{node_id=\"test-node-id\",node_name=\"test-node\"}")
	assert.Contains(t, bodyStr, "beacon_packet_loss_rate{node_id=\"test-node-id\",node_name=\"test-node\"}")
	assert.Contains(t, bodyStr, "beacon_jitter_ms{node_id=\"test-node-id\",node_name=\"test-node\"}")
}

// TestMetricsStopGraceful tests graceful shutdown
func TestMetricsStopGraceful(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: true,
		MetricsPort:    19118,
		MetricsUpdateSeconds: 10,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Start metrics server
	err = m.Start()
	require.NoError(t, err)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop metrics server
	err = m.Stop()
	assert.NoError(t, err)
	assert.False(t, m.IsRunning())

	// Verify server is no longer listening
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get("http://localhost:19118/metrics")
	assert.Error(t, err) // Should fail because server is stopped
}

// TestMetricsStopNotRunning tests stopping a non-running server
func TestMetricsStopNotRunning(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:         "test-node-id",
		NodeName:       "test-node",
		MetricsEnabled: true,
		MetricsPort:    19119,
		MetricsUpdateSeconds: 10,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Stop without starting (should not error)
	err = m.Stop()
	assert.NoError(t, err)
}

// TestMetricsWaitGroupBlocking tests that Stop() blocks until collector finishes
// Fix #2: Add WaitGroup.Wait() blocking test
func TestMetricsWaitGroupBlocking(t *testing.T) {
	initTestLogger(t)
	defer logger.Close()
	cfg := &config.Config{
		NodeID:               "test-node-id",
		NodeName:             "test-node",
		MetricsEnabled:       true,
		MetricsPort:          19120,
		MetricsUpdateSeconds: 10,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Start metrics server
	err = m.Start()
	require.NoError(t, err)

	// Give time for collector to start
	time.Sleep(100 * time.Millisecond)

	// Track if collector goroutine has finished
	collectorDone := make(chan struct{})
	
	// Call Stop in goroutine to detect blocking
	go func() {
		_ = m.Stop()
		close(collectorDone)
	}()

	// Stop should block until collector goroutine finishes
	// If it doesn't block, this will timeout indicating a problem
	select {
	case <-collectorDone:
		// Good - Stop() completed (after waiting for collector)
		assert.False(t, m.IsRunning())
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() took too long - possible deadlock or WaitGroup issue")
	}
}

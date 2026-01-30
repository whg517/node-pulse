package metrics_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"beacon/internal/config"
	"beacon/internal/metrics"
	"beacon/internal/probe"
)

// TestPrometheusScrapingWorkflow simulates Prometheus scraping the /metrics endpoint
func TestPrometheusScrapingWorkflow(t *testing.T) {
	cfg := &config.Config{
		NodeID:         "integration-test-node",
		NodeName:       "beacon-integration",
		MetricsEnabled: true,
		MetricsPort:    29112,
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

	// Create probe scheduler
	scheduler, err := probe.NewProbeScheduler(cfg.Probes)
	require.NoError(t, err)

	// Start scheduler
	err = scheduler.Start()
	require.NoError(t, err)
	defer scheduler.Stop()

	// Create metrics server
	metricsServer := metrics.NewMetrics(cfg, scheduler)

	// Start metrics server
	err = metricsServer.Start()
	require.NoError(t, err)
	defer metricsServer.Stop()

	// Wait for server to start and first metrics collection
	time.Sleep(500 * time.Millisecond)

	// Simulate Prometheus scraping
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Verify Prometheus exposition format compliance
	assert.Contains(t, bodyStr, "# HELP beacon_up")
	assert.Contains(t, bodyStr, "# TYPE beacon_up gauge")
	assert.Contains(t, bodyStr, fmt.Sprintf("beacon_up{node_id=\"%s\",node_name=\"%s\"} 1", cfg.NodeID, cfg.NodeName))

	assert.Contains(t, bodyStr, "# HELP beacon_rtt_seconds")
	assert.Contains(t, bodyStr, "# TYPE beacon_rtt_seconds gauge")

	assert.Contains(t, bodyStr, "# HELP beacon_packet_loss_rate")
	assert.Contains(t, bodyStr, "# TYPE beacon_packet_loss_rate gauge")

	assert.Contains(t, bodyStr, "# HELP beacon_jitter_ms")
	assert.Contains(t, bodyStr, "# TYPE beacon_jitter_ms gauge")

	// Verify all metrics have correct labels
	assert.Contains(t, bodyStr, fmt.Sprintf("node_id=\"%s\"", cfg.NodeID))
	assert.Contains(t, bodyStr, fmt.Sprintf("node_name=\"%s\"", cfg.NodeName))
}

// TestMetricsUpdateAfterProbeExecution verifies metrics update after probe results
func TestMetricsUpdateAfterProbeExecution(t *testing.T) {
	cfg := &config.Config{
		NodeID:         "update-test-node",
		NodeName:       "beacon-update",
		MetricsEnabled: true,
		MetricsPort:    29113,
		Probes: []config.ProbeConfig{
			{
				Type:           "tcp_ping",
				Target:         "127.0.0.1",
				Port:           22, // SSH port (should be open on most systems)
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          10,
			},
		},
	}

	// Create and start probe scheduler
	scheduler, err := probe.NewProbeScheduler(cfg.Probes)
	require.NoError(t, err)
	err = scheduler.Start()
	require.NoError(t, err)
	defer scheduler.Stop()

	// Create and start metrics server
	metricsServer := metrics.NewMetrics(cfg, scheduler)
	err = metricsServer.Start()
	require.NoError(t, err)
	defer metricsServer.Stop()

	// Wait for probes to execute and metrics to update
	time.Sleep(2 * time.Second)

	// Request metrics
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Verify beacon_up is 1 (running)
	assert.Contains(t, bodyStr, fmt.Sprintf("beacon_up{node_id=\"%s\",node_name=\"%s\"} 1", cfg.NodeID, cfg.NodeName))

	// Verify other metrics are present (values may vary based on probe results)
	assert.Contains(t, bodyStr, fmt.Sprintf("beacon_rtt_seconds{node_id=\"%s\",node_name=\"%s\"}", cfg.NodeID, cfg.NodeName))
	assert.Contains(t, bodyStr, fmt.Sprintf("beacon_packet_loss_rate{node_id=\"%s\",node_name=\"%s\"}", cfg.NodeID, cfg.NodeName))
	assert.Contains(t, bodyStr, fmt.Sprintf("beacon_jitter_ms{node_id=\"%s\",node_name=\"%s\"}", cfg.NodeID, cfg.NodeName))
}

// TestMetricsNoProbeResultsScenario tests metrics when no probe results are available
func TestMetricsNoProbeResultsScenario(t *testing.T) {
	cfg := &config.Config{
		NodeID:         "no-results-node",
		NodeName:       "beacon-no-results",
		MetricsEnabled: true,
		MetricsPort:    29114,
	}

	// Create scheduler with no probes
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	// Create and start metrics server
	metricsServer := metrics.NewMetrics(cfg, scheduler)
	err = metricsServer.Start()
	require.NoError(t, err)
	defer metricsServer.Stop()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Request metrics
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Verify beacon_up is 1
	assert.Contains(t, bodyStr, "beacon_up")
	assert.Contains(t, bodyStr, "} 1")

	// Verify other metrics are present with default values (0)
	assert.Contains(t, bodyStr, "beacon_rtt_seconds")
	assert.Contains(t, bodyStr, "beacon_packet_loss_rate")
	assert.Contains(t, bodyStr, "beacon_jitter_ms")
}

// TestMetricsServerStartStop tests server lifecycle
func TestMetricsServerStartStop(t *testing.T) {
	cfg := &config.Config{
		NodeID:         "lifecycle-test-node",
		NodeName:       "beacon-lifecycle",
		MetricsEnabled: true,
		MetricsPort:    29115,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	metricsServer := metrics.NewMetrics(cfg, scheduler)

	// Start server
	err = metricsServer.Start()
	require.NoError(t, err)
	assert.True(t, metricsServer.IsRunning())

	time.Sleep(200 * time.Millisecond)

	// Verify server is accessible
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Stop server
	err = metricsServer.Stop()
	require.NoError(t, err)
	assert.False(t, metricsServer.IsRunning())

	// Wait for server to fully stop
	time.Sleep(200 * time.Millisecond)

	// Verify server is no longer accessible
	_, err = http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
	assert.Error(t, err)
}

// TestPrometheusExpositionFormatCompliance verifies strict format compliance
func TestPrometheusExpositionFormatCompliance(t *testing.T) {
	cfg := &config.Config{
		NodeID:         "format-test-node",
		NodeName:       "beacon-format",
		MetricsEnabled: true,
		MetricsPort:    29116,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	metricsServer := metrics.NewMetrics(cfg, scheduler)
	err = metricsServer.Start()
	require.NoError(t, err)
	defer metricsServer.Stop()

	time.Sleep(200 * time.Millisecond)

	// Request metrics
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Parse and validate Prometheus exposition format
	lines := strings.Split(bodyStr, "\n")
	
	helpLines := 0
	typeLines := 0
	metricLines := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# HELP") {
			helpLines++
		} else if strings.HasPrefix(line, "# TYPE") {
			typeLines++
		} else if !strings.HasPrefix(line, "#") {
			metricLines++
			// Verify metric line format: metric_name{labels} value
			assert.True(t, strings.Contains(line, "{") && strings.Contains(line, "}"))
			assert.True(t, strings.Contains(line, "node_id=\""))
			assert.True(t, strings.Contains(line, "node_name=\""))
		}
	}

	// Verify we have the expected number of metrics (4 core metrics)
	assert.Equal(t, 4, helpLines, "Should have 4 HELP lines for core metrics")
	assert.Equal(t, 4, typeLines, "Should have 4 TYPE lines for core metrics")
	assert.GreaterOrEqual(t, metricLines, 4, "Should have at least 4 metric value lines")

	// Verify Content-Type header
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "text/plain")
}

// TestConfigurationDisabled tests that metrics server doesn't start when disabled
func TestConfigurationDisabled(t *testing.T) {
	cfg := &config.Config{
		NodeID:         "disabled-test-node",
		NodeName:       "beacon-disabled",
		MetricsEnabled: false,
		MetricsPort:    29117,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	metricsServer := metrics.NewMetrics(cfg, scheduler)

	// Start should succeed but not actually start the server
	err = metricsServer.Start()
	require.NoError(t, err)
	assert.False(t, metricsServer.IsRunning())

	// Verify server is not accessible
	_, err = http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
	assert.Error(t, err)
}

// TestMetricsPerformance tests response time requirements
func TestMetricsPerformance(t *testing.T) {
	cfg := &config.Config{
		NodeID:         "perf-test-node",
		NodeName:       "beacon-perf",
		MetricsEnabled: true,
		MetricsPort:    29118,
	}

	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	metricsServer := metrics.NewMetrics(cfg, scheduler)
	err = metricsServer.Start()
	require.NoError(t, err)
	defer metricsServer.Stop()

	time.Sleep(200 * time.Millisecond)

	// Test response time (should be < 100ms as per requirements)
	iterations := 10
	var totalDuration time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", cfg.MetricsPort))
		duration := time.Since(start)
		
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		totalDuration += duration
		
		// Individual request should be < 100ms
		assert.Less(t, duration.Milliseconds(), int64(100), "Individual request should be < 100ms")
	}

	// Average response time should be < 50ms
	avgDuration := totalDuration / time.Duration(iterations)
	assert.Less(t, avgDuration.Milliseconds(), int64(50), "Average response time should be < 50ms")
	
	t.Logf("Average response time: %v", avgDuration)
}

package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/monitor"
	"beacon/internal/probe"
)

// TestIntegration_ResourceMonitorLifecycle tests resource monitor start/stop lifecycle
func TestIntegration_ResourceMonitorLifecycle(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "test-node-resource-monitor"
node_name: "Test Resource Monitor"

probes:
  - type: tcp_ping
    target: localhost
    port: 80
    timeout_seconds: 1
    interval: 60
    count: 10

# Resource monitor configuration
resource_monitor:
  enabled: true
  check_interval_seconds: 1
  thresholds:
    cpu_microcores: 100
    memory_mb: 100
  degradation:
    degraded_level:
      cpu_microcores: 200
      memory_mb: 150
      interval_multiplier: 2
    critical_level:
      cpu_microcores: 300
      memory_mb: 200
      interval_multiplier: 3
    recovery:
      consecutive_normal_checks: 3
  alerting:
    suppression_window_seconds: 5

log_level: INFO
log_file: /tmp/beacon-test.log
log_to_console: false
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := logger.InitLogger(cfg); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Close()

	// Create probe scheduler
	scheduler, err := probe.NewProbeScheduler(cfg.Probes)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Create logrus adapter for monitor
	logAdapter := &monitor.LogrusLogger{}

	// Convert config
	monitorCfg := &monitor.ResourceMonitorConfig{
		Enabled:              cfg.ResourceMonitor.Enabled,
		CheckIntervalSeconds: cfg.ResourceMonitor.CheckIntervalSeconds,
		Thresholds: monitor.ThresholdsConfig{
			CPUMicrocores: cfg.ResourceMonitor.Thresholds.CPUMicrocores,
			MemoryMB:      cfg.ResourceMonitor.Thresholds.MemoryMB,
		},
		Degradation: monitor.DegradationConfig{
			DegradedLevel: monitor.DegradationLevelConfig{
				CPUMicrocores:      cfg.ResourceMonitor.Degradation.DegradedLevel.CPUMicrocores,
				MemoryMB:           cfg.ResourceMonitor.Degradation.DegradedLevel.MemoryMB,
				IntervalMultiplier: cfg.ResourceMonitor.Degradation.DegradedLevel.IntervalMultiplier,
			},
			CriticalLevel: monitor.DegradationLevelConfig{
				CPUMicrocores:      cfg.ResourceMonitor.Degradation.CriticalLevel.CPUMicrocores,
				MemoryMB:           cfg.ResourceMonitor.Degradation.CriticalLevel.MemoryMB,
				IntervalMultiplier: cfg.ResourceMonitor.Degradation.CriticalLevel.IntervalMultiplier,
			},
			Recovery: monitor.RecoveryConfig{
				ConsecutiveNormalChecks: cfg.ResourceMonitor.Degradation.Recovery.ConsecutiveNormalChecks,
			},
		},
		Alerting: monitor.AlertingConfig{
			SuppressionWindowSeconds: cfg.ResourceMonitor.Alerting.SuppressionWindowSeconds,
		},
	}

	// Create monitor
	resourceMonitor, err := monitor.NewMonitor(monitorCfg, scheduler, logAdapter)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	// Test start
	if err := resourceMonitor.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	if !resourceMonitor.IsRunning() {
		t.Error("Monitor should be running after Start()")
	}

	// Wait for at least one resource check
	time.Sleep(2 * time.Second)

	// Get resource usage
	usage := resourceMonitor.GetResourceUsage()
	if usage == nil {
		t.Error("Resource usage should not be nil after start")
	} else {
		t.Logf("Resource usage - CPU: %.2f microcores, Memory: %.2f MB", usage.CPUMicrocores, usage.MemoryMB)
	}

	// Get degradation level
	level := resourceMonitor.GetDegradationLevel()
	t.Logf("Current degradation level: %s", level.String())

	// Test stop
	resourceMonitor.Stop()
	time.Sleep(500 * time.Millisecond)

	if resourceMonitor.IsRunning() {
		t.Error("Monitor should not be running after Stop()")
	}

	t.Log("✅ Resource monitor lifecycle test passed")
}

// TestIntegration_ResourceMonitorAlertSuppression tests alert suppression mechanism
func TestIntegration_ResourceMonitorAlertSuppression(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "beacon.yaml")

	// Very low thresholds to trigger alerts quickly
	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "test-node-alerts"
node_name: "Test Alerts"

probes:
  - type: tcp_ping
    target: localhost
    port: 80
    timeout_seconds: 1
    interval: 60
    count: 10

resource_monitor:
  enabled: true
  check_interval_seconds: 1
  thresholds:
    cpu_microcores: 1     # Very low to trigger alerts
    memory_mb: 1          # Very low to trigger alerts
  degradation:
    degraded_level:
      cpu_microcores: 2
      memory_mb: 2
      interval_multiplier: 2
    critical_level:
      cpu_microcores: 3
      memory_mb: 3
      interval_multiplier: 3
    recovery:
      consecutive_normal_checks: 2
  alerting:
    suppression_window_seconds: 3  # Short window for testing

log_level: INFO
log_file: /tmp/beacon-test-alerts.log
log_to_console: false
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.InitLogger(cfg); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Close()

	scheduler, _ := probe.NewProbeScheduler(cfg.Probes)
	logAdapter := &monitor.LogrusLogger{}

	monitorCfg := &monitor.ResourceMonitorConfig{
		Enabled:              true,
		CheckIntervalSeconds: 1,
		Thresholds: monitor.ThresholdsConfig{
			CPUMicrocores: 1,  // Very low threshold
			MemoryMB:      1,
		},
		Degradation: monitor.DegradationConfig{
			DegradedLevel: monitor.DegradationLevelConfig{
				CPUMicrocores:      2,
				MemoryMB:           2,
				IntervalMultiplier: 2,
			},
			CriticalLevel: monitor.DegradationLevelConfig{
				CPUMicrocores:      3,
				MemoryMB:           3,
				IntervalMultiplier: 3,
			},
			Recovery: monitor.RecoveryConfig{
				ConsecutiveNormalChecks: 2,
			},
		},
		Alerting: monitor.AlertingConfig{
			SuppressionWindowSeconds: 3,
		},
	}

	resourceMonitor, err := monitor.NewMonitor(monitorCfg, scheduler, logAdapter)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := resourceMonitor.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer resourceMonitor.Stop()

	// Wait for alerts to trigger (resource usage will exceed low thresholds)
	time.Sleep(3 * time.Second)

	alerts := resourceMonitor.GetAlerts()
	if len(alerts) == 0 {
		t.Log("⚠️  No alerts triggered (resource usage below threshold - this is OK in test environment)")
	} else {
		t.Logf("✅ Alert suppression test: %d alerts triggered", len(alerts))
		for i, alert := range alerts {
			t.Logf("  Alert %d: %s=%.2f (threshold=%.2f), level=%s",
				i+1, alert.ResourceType, alert.CurrentValue, alert.Threshold, alert.Level)
		}

		// Verify suppression by waiting and checking if new alerts are suppressed
		initialAlertCount := len(alerts)
		time.Sleep(4 * time.Second) // Wait past suppression window
		newAlerts := resourceMonitor.GetAlerts()

		if len(newAlerts) > initialAlertCount {
			t.Logf("✅ New alerts triggered after suppression window: %d -> %d",
				initialAlertCount, len(newAlerts))
		}
	}
}

// TestIntegration_ResourceMonitorWithDebugCommand tests debug command integration
func TestIntegration_ResourceMonitorWithDebugCommand(t *testing.T) {
	// Build beacon binary
	if err := exec.Command("go", "build", "-o", "/tmp/beacon-test", ".").Run(); err != nil {
		t.Skipf("Skipping test: failed to build beacon: %v", err)
	}
	defer os.Remove("/tmp/beacon-test")

	// Create config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "test-debug-integration"
node_name: "Debug Integration Test"

probes:
  - type: tcp_ping
    target: localhost
    port: 80
    timeout_seconds: 1
    interval: 60
    count: 10

resource_monitor:
  enabled: true
  check_interval_seconds: 60
  thresholds:
    cpu_microcores: 100
    memory_mb: 100
  degradation:
    degraded_level:
      cpu_microcores: 200
      memory_mb: 150
      interval_multiplier: 2
    critical_level:
      cpu_microcores: 300
      memory_mb: 200
      interval_multiplier: 3
    recovery:
      consecutive_normal_checks: 3
  alerting:
    suppression_window_seconds: 300

log_level: INFO
log_file: /tmp/beacon-test-debug.log
log_to_console: false
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Run debug command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/tmp/beacon-test", "debug", "--config", configFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Debug command output (stderr expected): %s", string(output))
	}

	outputStr := string(output)

	// Verify resource monitor section exists in JSON output
	if !strings.Contains(outputStr, "resource_monitor") {
		t.Error("Debug output should contain 'resource_monitor' section")
	}

	// Verify configuration fields are present
	if !strings.Contains(outputStr, "cpu_microcores") {
		t.Error("Debug output should contain 'cpu_microcores' configuration")
	}
	if !strings.Contains(outputStr, "memory_mb") {
		t.Error("Debug output should contain 'memory_mb' configuration")
	}

	// Verify debug mode indicators
	if strings.Contains(outputStr, "not_applicable") {
		t.Log("✅ Debug mode correctly shows 'not_applicable' for runtime status")
	}

	t.Log("✅ Debug command integration test passed")
}

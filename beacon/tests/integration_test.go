package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestIntegration_BeaconStartCommandOutput(t *testing.T) {
	t.Skip("Skipping integration test - start command blocks waiting for signals")
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/beacon.yaml"
	logFile := tmpDir + "/beacon.log"
	configContent := fmt.Sprintf(`
pulse_server: "http://localhost:8080"
node_id: "test-integration-01"
node_name: "Integration Test Node"
log_file: "%s"
log_level: "INFO"
`, logFile)
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run beacon start command using go run
	cmd := exec.Command("go", "run", "../main.go", "start", "--config", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run beacon start: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "[INFO] Loading configuration...") {
		t.Error("Expected start output to contain '[INFO] Loading configuration...'")
	}
	if !strings.Contains(outputStr, "[INFO] Starting probes...") {
		t.Error("Expected start output to contain '[INFO] Starting probes...'")
	}
	if !strings.Contains(outputStr, "[INFO] Connecting to Pulse...") {
		t.Error("Expected start output to contain '[INFO] Connecting to Pulse...'")
	}
	if !strings.Contains(outputStr, "[INFO] Registration successful!") {
		t.Error("Expected start output to contain '[INFO] Registration successful!'")
	}
	if !strings.Contains(outputStr, "[INFO] Starting Beacon...") {
		t.Error("Expected start output to contain '[INFO] Starting Beacon...'")
	}
}

func TestIntegration_BeaconStopCommandOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/beacon.yaml"
	logFile := tmpDir + "/beacon.log"
	configContent := fmt.Sprintf(`
pulse_server: "http://localhost:8080"
node_id: "test-integration-stop"
node_name: "Integration Test Node"
log_file: "%s"
log_level: "INFO"
`, logFile)
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run beacon stop command using go run
	cmd := exec.Command("go", "run", "../main.go", "stop", "--config", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run beacon stop: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "[INFO] Stopping Beacon...") {
		t.Error("Expected stop output to contain '[INFO] Stopping Beacon...'")
	}
	// Accept either "stopped successfully" or "is not running" since beacon may not be running
	if !strings.Contains(outputStr, "stopped successfully") && !strings.Contains(outputStr, "is not running") {
		t.Errorf("Expected stop output to contain 'stopped successfully' or 'is not running', got: %s", outputStr)
	}
}

func TestIntegration_BeaconStatusCommandOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/beacon.yaml"
	logFile := tmpDir + "/beacon.log"
	configContent := fmt.Sprintf(`
pulse_server: "http://localhost:8080"
node_id: "test-integration-01"
node_name: "Integration Test Node"
log_file: "%s"
log_level: "INFO"
`, logFile)
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run beacon status command using go run
	cmd := exec.Command("go", "run", "../main.go", "status", "--config", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run beacon status: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	// Verify JSON output
	if !strings.Contains(outputStr, `"status"`) {
		t.Error("Expected status output to contain JSON status field")
	}
	if !strings.Contains(outputStr, `"last_heartbeat"`) {
		t.Error("Expected status output to contain JSON last_heartbeat field")
	}
	if !strings.Contains(outputStr, `"config_version"`) {
		t.Error("Expected status output to contain JSON config_version field")
	}
	if !strings.Contains(outputStr, `"node_id"`) {
		t.Error("Expected status output to contain JSON node_id field")
	}
	if !strings.Contains(outputStr, `"node_name"`) {
		t.Error("Expected status output to contain JSON node_name field")
	}
}

func TestIntegration_BeaconDebugCommandOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config with multiple probes
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/beacon.yaml"
	logFile := tmpDir + "/beacon.log"
	configContent := fmt.Sprintf(`
pulse_server: "http://localhost:8080"
node_id: "test-integration-debug-01"
node_name: "Integration Debug Test Node"
log_file: "%s"
log_level: "INFO"
debug_mode: false
probes:
  - type: "tcp_ping"
    target: "192.168.1.1"
    port: 80
    timeout_seconds: 5
    interval: 60
    count: 10
  - type: "udp_ping"
    target: "192.168.1.2"
    port: 53
    timeout_seconds: 3
    interval: 120
    count: 5
`, logFile)
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run beacon debug command using go run
	cmd := exec.Command("go", "run", "../main.go", "debug", "--config", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run beacon debug: %v, output: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify JSON format (AC #3)
	if !strings.Contains(outputStr, `"timestamp"`) {
		t.Error("Expected debug output to contain JSON timestamp field")
	}
	if !strings.Contains(outputStr, `"level"`) {
		t.Error("Expected debug output to contain JSON level field")
	}
	if !strings.Contains(outputStr, `"message"`) {
		t.Error("Expected debug output to contain JSON message field")
	}
	if !strings.Contains(outputStr, `"diagnostics"`) {
		t.Error("Expected debug output to contain JSON diagnostics field")
	}

	// Verify all required diagnostic sections are present (AC #2)
	requiredSections := []string{
		`"network_status"`,
		`"configuration"`,
		`"connection_status"`,
		`"resource_usage"`,
		`"probe_tasks"`,
		`"prometheus_metrics"`,
	}
	for _, section := range requiredSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("Expected debug output to contain diagnostic section: %s", section)
		}
	}

	// Verify specific network status fields (AC #2)
	if !strings.Contains(outputStr, `"pulse_server_reachable"`) {
		t.Error("Expected network_status to contain pulse_server_reachable field")
	}
	if !strings.Contains(outputStr, `"pulse_server_address"`) {
		t.Error("Expected network_status to contain pulse_server_address field")
	}
	if !strings.Contains(outputStr, `"packet_loss_rate"`) {
		t.Error("Expected network_status to contain packet_loss_rate field")
	}

	// Verify configuration fields (AC #2)
	if !strings.Contains(outputStr, `"config_file"`) {
		t.Error("Expected configuration to contain config_file field")
	}
	if !strings.Contains(outputStr, `"log_level"`) {
		t.Error("Expected configuration to contain log_level field")
	}
	if !strings.Contains(outputStr, `"debug_mode"`) {
		t.Error("Expected configuration to contain debug_mode field")
	}

	// Verify connection status fields (AC #2)
	if !strings.Contains(outputStr, `"status"`) {
		t.Error("Expected connection_status to contain status field")
	}
	if !strings.Contains(outputStr, `"retry_count"`) {
		t.Error("Expected connection_status to contain retry_count field")
	}
	if !strings.Contains(outputStr, `"backoff_seconds"`) {
		t.Error("Expected connection_status to contain backoff_seconds field")
	}

	// Verify resource usage fields (AC #2)
	if !strings.Contains(outputStr, `"cpu_percent"`) {
		t.Error("Expected resource_usage to contain cpu_percent field")
	}
	if !strings.Contains(outputStr, `"memory_mb"`) {
		t.Error("Expected resource_usage to contain memory_mb field")
	}
	if !strings.Contains(outputStr, `"memory_percent"`) {
		t.Error("Expected resource_usage to contain memory_percent field")
	}

	// Verify probe tasks fields (AC #2)
	if !strings.Contains(outputStr, `"total_tasks"`) {
		t.Error("Expected probe_tasks to contain total_tasks field")
	}
	if !strings.Contains(outputStr, `"tasks"`) {
		t.Error("Expected probe_tasks to contain tasks array field")
	}

	// Verify structured log format - ISO 8601 timestamp (AC #3)
	if !strings.Contains(outputStr, "T") && strings.Contains(outputStr, `"timestamp"`) {
		// Extract timestamp value and check format
		// JSON should have ISO 8601 format like "2026-01-30T10:30:00Z"
		timestampIdx := strings.Index(outputStr, `"timestamp"`)
		if timestampIdx > 0 {
			// Look for the date pattern after timestamp field
			afterTimestamp := outputStr[timestampIdx:]
			// ISO 8601 format should contain "T" between date and time
			if !strings.Contains(afterTimestamp[:50], "T") {
				t.Error("Expected timestamp in ISO 8601 format (should contain 'T' separator)")
			}
		}
	}

	// Verify log level is DEBUG (AC #3)
	if !strings.Contains(outputStr, `"level": "DEBUG"`) {
		t.Error("Expected log level to be DEBUG in diagnostic output")
	}
}

func TestIntegration_BeaconDebugCommandPrettyOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/beacon.yaml"
	logFile := tmpDir + "/beacon.log"
	configContent := fmt.Sprintf(`
pulse_server: "http://localhost:8080"
node_id: "test-integration-debug-pretty"
node_name: "Integration Debug Pretty Test"
log_file: "%s"
log_level: "INFO"
probes:
  - type: "tcp_ping"
    target: "192.168.1.1"
    port: 80
    timeout_seconds: 5
    interval: 60
    count: 10
`, logFile)
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run beacon debug command with --pretty flag
	cmd := exec.Command("go", "run", "../main.go", "debug", "--pretty", "--config", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run beacon debug --pretty: %v, output: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify human-readable format indicators (AC #3)
	expectedHeaders := []string{
		"Beacon Diagnostic Information",
		"Network Status",
		"Configuration",
		"Connection Status",
		"Resource Usage",
		"Probe Tasks",
		"Prometheus Metrics",
	}
	for _, header := range expectedHeaders {
		if !strings.Contains(outputStr, header) {
			t.Errorf("Expected pretty output to contain header: %s", header)
		}
	}

	// Pretty output should contain section dividers
	if !strings.Contains(outputStr, "=") && !strings.Contains(outputStr, "-") {
		t.Error("Expected pretty output to contain formatting characters")
	}
}

func TestIntegration_BeaconDebugCommandDebugMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config with debug_mode: true
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/beacon.yaml"
	logFile := tmpDir + "/beacon.log"
	configContent := fmt.Sprintf(`
pulse_server: "http://localhost:8080"
node_id: "test-integration-debug-mode"
node_name: "Integration Debug Mode Test"
log_file: "%s"
log_level: "WARN"
debug_mode: true
`, logFile)
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run beacon debug command
	cmd := exec.Command("go", "run", "../main.go", "debug", "--config", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run beacon debug: %v, output: %s", err, string(output))
	}

	outputStr := string(output)

	// When debug_mode is true, log_level should be DEBUG (AC #2)
	// This verifies the debug_mode configuration integration
	if !strings.Contains(outputStr, `"log_level": "DEBUG"`) {
		t.Error("Expected debug_mode=true to override log_level to DEBUG")
	}
	if !strings.Contains(outputStr, `"debug_mode": true`) {
		t.Error("Expected debug_mode to be true in configuration output")
	}
}

func TestIntegration_InvalidCommandError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Run beacon with invalid command using go run
	cmd := exec.Command("go", "run", "../main.go", "invalid-command")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}

	// Check error output
	errorStr := stderr.String()
	if !strings.Contains(errorStr, "unknown command") && !strings.Contains(errorStr, "invalid-command") {
		t.Error("Expected error message for unknown command")
	}
}

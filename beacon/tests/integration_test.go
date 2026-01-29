package tests

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestIntegration_BeaconStartCommandOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config
	tmpFile := t.TempDir() + "/beacon.yaml"
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-integration-01"
node_name: "Integration Test Node"
`
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
	tmpFile := t.TempDir() + "/beacon.yaml"
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-integration-stop"
node_name: "Integration Test Node"
`
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
	if !strings.Contains(outputStr, "[INFO] Beacon stopped successfully") {
		t.Error("Expected stop output to contain '[INFO] Beacon stopped successfully'")
	}
}

func TestIntegration_BeaconStatusCommandOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test config
	tmpFile := t.TempDir() + "/beacon.yaml"
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-integration-01"
node_name: "Integration Test Node"
`
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

	// Create test config
	tmpFile := t.TempDir() + "/beacon.yaml"
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-integration-01"
node_name: "Integration Test Node"
`
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
	if !strings.Contains(outputStr, "[DEBUG] Debug mode enabled") {
		t.Error("Expected debug output to contain '[DEBUG] Debug mode enabled'")
	}
	if !strings.Contains(outputStr, "[DEBUG] (1/4)") {
		t.Error("Expected debug output to contain step indicator '(1/4)'")
	}
	if !strings.Contains(outputStr, "[DEBUG] (2/4)") {
		t.Error("Expected debug output to contain step indicator '(2/4)'")
	}
	if !strings.Contains(outputStr, "[DEBUG] (3/4)") {
		t.Error("Expected debug output to contain step indicator '(3/4)'")
	}
	if !strings.Contains(outputStr, "[DEBUG] (4/4)") {
		t.Error("Expected debug output to contain step indicator '(4/4)'")
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

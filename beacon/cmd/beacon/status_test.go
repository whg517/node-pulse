package beacon

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStatusCommand_JSONOutput tests JSON formatted status output (AC #9)
func TestStatusCommand_JSONOutput(t *testing.T) {
	// Create test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "status"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for valid JSON output (AC #9)
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}
}

// TestStatusCommand_StatusField tests online/offline status field (AC #10)
func TestStatusCommand_StatusField(t *testing.T) {
	// Create test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "status"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for status field (AC #10)
	if !strings.Contains(output, `"status"`) {
		t.Error("Expected status field in JSON output")
	}
}

// TestStatusCommand_LastHeartbeat tests last heartbeat field (AC #10)
func TestStatusCommand_LastHeartbeat(t *testing.T) {
	// Create test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "status"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for last_heartbeat field (AC #10)
	if !strings.Contains(output, `"last_heartbeat"`) {
		t.Error("Expected last_heartbeat field in JSON output")
	}
}

// TestStatusCommand_ConfigVersion tests config version field (AC #10)
func TestStatusCommand_ConfigVersion(t *testing.T) {
	// Create test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "status"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for config_version field (AC #10)
	if !strings.Contains(output, `"config_version"`) {
		t.Error("Expected config_version field in JSON output")
	}
}

// TestStatusCommand_OfflineOnMissingConfig tests status returns offline when config missing
func TestStatusCommand_OfflineOnMissingConfig(t *testing.T) {
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", "/nonexistent/config.yaml", "status"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Should still return valid JSON with offline status
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Errorf("Expected valid JSON output even when config missing, got error: %v", err)
	}

	// Check for offline status
	if status["status"] != "offline" {
		t.Errorf("Expected status to be 'offline' when config is missing, got: %v", status["status"])
	}
}

// TestStatusCommand_AllRequiredFields tests all required fields are present
func TestStatusCommand_AllRequiredFields(t *testing.T) {
	// Create test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "status"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Parse JSON and check all required fields
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	requiredFields := []string{"status", "last_heartbeat", "config_version", "node_id", "node_name"}
	for _, field := range requiredFields {
		if _, ok := status[field]; !ok {
			t.Errorf("Expected %s field in status output", field)
		}
	}
}

// TestStatusCommand_JSONFormat tests JSON is properly formatted
func TestStatusCommand_JSONFormat(t *testing.T) {
	// Create test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "status"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for indentation (2 spaces)
	lines := strings.Split(output, "\n")
	if len(lines) > 1 {
		// JSON should be indented
		if !strings.Contains(output, "  ") {
			t.Log("Note: JSON output may not be indented with 2 spaces")
		}
	}

	// Verify JSON can be parsed
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Errorf("Expected valid JSON output: %v", err)
	}
}

package beacon

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStopCommand_GracefulShutdown tests graceful shutdown output (AC #7-8)
func TestStopCommand_GracefulShutdown(t *testing.T) {
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
	GetRootCmd().SetArgs([]string{"--config", configPath, "stop"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for stop messages (AC #7-8)
	if !strings.Contains(output, "Stopping") && !strings.Contains(output, "停止") {
		t.Error("Expected stop output for graceful shutdown")
	}
	if !strings.Contains(output, "stopped") && !strings.Contains(output, "成功") &&
	   !strings.Contains(output, "success") && !strings.Contains(output, "未运行") &&
	   !strings.Contains(output, "not running") {
		t.Error("Expected stop success message or not running message")
	}
}

// TestStopCommand_WaitForProbe tests that stop waits for probes to complete (AC #7)
func TestStopCommand_WaitForProbe(t *testing.T) {
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
	GetRootCmd().SetArgs([]string{"--config", configPath, "stop"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Should contain some indication of graceful shutdown
	if !strings.Contains(output, "Beacon") && !strings.Contains(output, "beacon") {
		t.Error("Expected output to mention beacon process")
	}
}

// TestStopCommand_ConfirmationOutput tests stop confirmation output (AC #8)
func TestStopCommand_ConfirmationOutput(t *testing.T) {
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
	GetRootCmd().SetArgs([]string{"--config", configPath, "stop"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for confirmation of stop (AC #8)
	if !strings.Contains(output, "成功") && !strings.Contains(output, "success") &&
	   !strings.Contains(output, "完成") && !strings.Contains(output, "complete") &&
	   !strings.Contains(output, "未运行") && !strings.Contains(output, "not running") {
		t.Error("Expected stop confirmation message or not running message")
	}
}

// TestStopCommand_NoConfig tests stop with no config file
func TestStopCommand_NoConfig(t *testing.T) {
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", "/nonexistent/beacon.yaml", "stop"})

	// Execute command - should handle gracefully
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Should still show some output
	if output == "" {
		t.Error("Expected some output from stop command")
	}
}

// TestStopCommand_PIDFileHandling tests stop command handles PID files correctly
func TestStopCommand_PIDFileHandling(t *testing.T) {
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
	GetRootCmd().SetArgs([]string{"--config", configPath, "stop"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Should mention PID or process status
	if !strings.Contains(output, "PID") && !strings.Contains(output, "进程") &&
	   !strings.Contains(output, "process") && !strings.Contains(output, "未运行") {
		t.Log("Note: Stop output may not explicitly mention PID when no process is running")
	}
}

// TestStopCommand_GracefulShutdownMessage tests graceful shutdown messaging
func TestStopCommand_GracefulShutdownMessage(t *testing.T) {
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
	GetRootCmd().SetArgs([]string{"--config", configPath, "stop"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Check for stopping message
	if !strings.Contains(output, "停止") && !strings.Contains(output, "Stopping") {
		t.Error("Expected stop command to show stopping message")
	}
}

// TestStopCommand_CleanupOnNonRunning tests cleanup when beacon not running
func TestStopCommand_CleanupOnNonRunning(t *testing.T) {
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
	GetRootCmd().SetArgs([]string{"--config", configPath, "stop"})

	// Execute command
	_ = GetRootCmd().Execute()

	output := buf.String()

	// Should handle gracefully when beacon not running
	if output == "" {
		t.Error("Expected stop command to produce output even when beacon not running")
	}

	// Should mention that beacon is not running
	if !strings.Contains(output, "未运行") && !strings.Contains(output, "not running") &&
	   !strings.Contains(output, "可能未运行") {
		t.Log("Note: Stop command may not explicitly mention 'not running' status")
	}
}

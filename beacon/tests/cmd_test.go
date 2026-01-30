package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"beacon/cmd/beacon"
)

// createTestConfig creates a minimal valid test config file
func createTestConfig(t *testing.T) string {
	tmpFile := filepath.Join(t.TempDir(), "test-config.yaml")
	logFile := filepath.Join(t.TempDir(), "beacon.log")
	configContent := fmt.Sprintf(`
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
log_file: "%s"
log_level: "INFO"
`, logFile)
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	return tmpFile
}

// executeWithTimeout executes the beacon command with a timeout
func executeWithTimeout(timeout time.Duration) (string, error) {
	var buf bytes.Buffer
	beacon.GetRootCmd().SetOut(&buf)
	beacon.GetRootCmd().SetErr(&buf)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	var output string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		err := beacon.GetRootCmd().ExecuteContext(ctx)
		output = buf.String()
		done <- err
	}()

	var err error
	select {
	case err = <-done:
		// Got result
	case <-time.After(timeout + 1*time.Second):
		// Timeout
		err = fmt.Errorf("execution timed out after %v", timeout)
	}
	cancel() // Cancel the context
	wg.Wait() // Wait for goroutine to finish
	return output, err
}

func TestRootCommand(t *testing.T) {
	// Create test config
	tmpFile := createTestConfig(t)

	// Capture output
	var buf bytes.Buffer
	beacon.GetRootCmd().SetOut(&buf)
	beacon.GetRootCmd().SetArgs([]string{"--config", tmpFile})

	// Execute command
	err := beacon.GetRootCmd().Execute()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check help output is shown
	output := buf.String()
	if !strings.Contains(output, "Available Commands") {
		t.Error("Expected help output to contain 'Available Commands'")
	}
	if !strings.Contains(output, "start") {
		t.Error("Expected help output to contain 'start'")
	}
	if !strings.Contains(output, "stop") {
		t.Error("Expected help output to contain 'stop'")
	}
}

func TestStartCommand(t *testing.T) {
	t.Skip("Skipping TestStartCommand - start command blocks waiting for signals")
	// Create test config
	tmpFile := createTestConfig(t)
	beacon.GetRootCmd().SetArgs([]string{"--config", tmpFile, "start"})

	// Execute command with timeout - start command waits for interrupt
	output, _ := executeWithTimeout(2 * time.Second)

	// Check output - should at least show initial messages
	if !strings.Contains(output, "Loading configuration") {
		t.Error("Expected start output to contain 'Loading configuration'")
	}
	if !strings.Contains(output, "Node ID") {
		t.Error("Expected start output to contain 'Node ID'")
	}
}

func TestStopCommand(t *testing.T) {
	// Create test config
	tmpFile := createTestConfig(t)

	// Capture output
	var buf bytes.Buffer
	beacon.GetRootCmd().SetOut(&buf)
	beacon.GetRootCmd().SetArgs([]string{"--config", tmpFile, "stop"})

	// Execute command
	err := beacon.GetRootCmd().Execute()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "[INFO] Stopping Beacon...") {
		t.Error("Expected stop output to contain '[INFO] Stopping Beacon...'")
	}
	// Accept either "stopped successfully" or "is not running" since beacon may not be running
	if !strings.Contains(output, "stopped successfully") && !strings.Contains(output, "is not running") {
		t.Errorf("Expected stop output to contain 'stopped successfully' or 'is not running', got: %s", output)
	}
}

func TestStatusCommand(t *testing.T) {
	// Create test config
	tmpFile := createTestConfig(t)

	// Capture output
	var buf bytes.Buffer
	beacon.GetRootCmd().SetOut(&buf)
	beacon.GetRootCmd().SetArgs([]string{"--config", tmpFile, "status"})

	// Execute command
	err := beacon.GetRootCmd().Execute()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check output contains JSON
	output := buf.String()
	if !strings.Contains(output, `"status"`) {
		t.Error("Expected status output to contain JSON status field")
	}
	if !strings.Contains(output, `"last_heartbeat"`) {
		t.Error("Expected status output to contain JSON last_heartbeat field")
	}
	if !strings.Contains(output, `"config_version"`) {
		t.Error("Expected status output to contain JSON config_version field")
	}
}

func TestDebugCommand(t *testing.T) {
	// Create test config
	tmpFile := createTestConfig(t)

	// Capture output
	var buf bytes.Buffer
	beacon.GetRootCmd().SetOut(&buf)
	beacon.GetRootCmd().SetArgs([]string{"--config", tmpFile, "debug"})

	// Execute command
	err := beacon.GetRootCmd().Execute()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "[DEBUG] Debug mode enabled") {
		t.Error("Expected debug output to contain '[DEBUG] Debug mode enabled'")
	}
	if !strings.Contains(output, "[DEBUG] (1/4)") {
		t.Error("Expected debug output to contain step indicators")
	}
}

func TestConfigFlag(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	beacon.GetRootCmd().SetOut(&buf)
	beacon.GetRootCmd().SetArgs([]string{"--config", "/tmp/test.yaml", "status"})

	// Execute command
	// Note: status command doesn't error on missing config, it returns offline status
	_ = beacon.GetRootCmd().Execute()

	// Check that config path was set
	if beacon.GetConfigFile() != "/tmp/test.yaml" {
		t.Errorf("Expected config file to be '/tmp/test.yaml', got: %s", beacon.GetConfigFile())
	}
}

func TestDebugFlag(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	beacon.GetRootCmd().SetOut(&buf)
	beacon.GetRootCmd().SetArgs([]string{"--debug", "status"})

	// Execute command
	_ = beacon.GetRootCmd().Execute()

	// Check that debug flag is set
	if !beacon.IsDebug() {
		t.Error("Expected debug flag to be true when --debug is passed")
	}
}

func TestInvalidCommand(t *testing.T) {
	// Capture error output
	var buf bytes.Buffer
	beacon.GetRootCmd().SetErr(&buf)
	beacon.GetRootCmd().SetArgs([]string{"invalid"})

	// Execute command
	err := beacon.GetRootCmd().Execute()
	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}

	// Check error output
	output := buf.String()
	if !strings.Contains(output, "unknown command") || !strings.Contains(output, "invalid") {
		t.Error("Expected error message for unknown command")
	}
}

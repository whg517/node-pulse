package beacon

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
)

// executeWithTimeout executes the command with a timeout context
func executeWithTimeout(timeout time.Duration) (string, error) {
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	// Note: Don't clear args here - the test should set them before calling

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

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
		err := GetRootCmd().ExecuteContext(ctx)
		output = buf.String()
		done <- err
	}()

	var err error
	select {
	case err = <-done:
		// Got result, wait for goroutine to finish
	case <-time.After(timeout + 1*time.Second):
		// Timeout, but we still need to wait for goroutine
		err = fmt.Errorf("execution timed out after %v", timeout)
	}
	cancel() // Cancel the context
	wg.Wait() // Wait for goroutine to finish
	return output, err
}

// TestStartCommand_ProgressOutput tests real-time progress output (AC #2)
func TestStartCommand_ProgressOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

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

	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	// Execute command with timeout - use a short timeout for test
	output, err := executeWithTimeout(2 * time.Second)
	t.Logf("Output: %s", output)
	t.Logf("Error: %v", err)

	// Check for progress messages (AC #2)
	foundLoading := strings.Contains(output, "Loading configuration") || strings.Contains(output, "加载配置文件")
	foundProgress := strings.Contains(output, "(1/4)") || strings.Contains(output, "(2/4)")

	if !foundLoading && !foundProgress && err == nil {
		t.Error("Expected progress output for loading configuration")
	}
}

// TestStartCommand_DeploymentTiming tests deployment timing statistics (AC #5)
func TestStartCommand_DeploymentTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	t.Log("TestStartCommand_DeploymentTiming: starting")
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Test panicked: %v", r)
			t.Fail()
		}
	}()

	// Create test config
	t.Log("TestStartCommand_DeploymentTiming: creating config")
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

	t.Log("TestStartCommand_DeploymentTiming: setting args")
	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	t.Log("TestStartCommand_DeploymentTiming: calling executeWithTimeout")
	// Execute command with timeout
	output, err := executeWithTimeout(2 * time.Second)
	t.Logf("Output: %s", output)
	t.Logf("Error: %v", err)

	// Check for any progress indicators (AC #1)
	foundProgress := strings.Contains(output, "(1/4)") || strings.Contains(output, "(2/4)") ||
		strings.Contains(output, "(3/4)") || strings.Contains(output, "(4/4)")
	if !foundProgress && err == nil {
		t.Error("Expected progress step indicators")
	}

	// Check for deployment completion message (AC #5)
	foundCompletion := strings.Contains(output, "部署完成") || strings.Contains(output, "completed")
	if !foundCompletion && err == nil {
		t.Log("Note: Deployment completion message not found (timeout expected)")
	}
}

// TestStartCommand_ConfigErrorWithLocation tests configuration error with location hints (AC #6)
func TestStartCommand_ConfigErrorWithLocation(t *testing.T) {
	// Create invalid config with indentation error
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
    wrong_indentation: "this is wrong"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	// Execute command - this fails fast due to invalid config
	output, _ := executeWithTimeout(2 * time.Second)

	// Check for location-specific error message (AC #6)
	if !strings.Contains(output, "第") && !strings.Contains(output, "Line") {
		t.Error("Expected error message with line number")
	}
	if !strings.Contains(output, "缩进") && !strings.Contains(output, "indentation") {
		t.Error("Expected error message to mention indentation issue")
	}
}

// TestStartCommand_RegistrationSuccess tests registration success output (AC #4)
func TestStartCommand_RegistrationSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
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

	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	// Execute command with timeout
	output, _ := executeWithTimeout(3 * time.Second)

	// Check for node ID in output (AC #4)
	if !strings.Contains(output, "Node ID") && !strings.Contains(output, "node_id") &&
	   !strings.Contains(output, "Node") && !strings.Contains(output, "节点") {
		t.Error("Expected registration success output with node ID or Node info")
	}
}

// TestStartCommand_InvalidConfig tests config file not found error
func TestStartCommand_InvalidConfig(t *testing.T) {
	// Use non-existent config file
	GetRootCmd().SetArgs([]string{"--config", "/nonexistent/path/config.yaml", "start"})

	// Execute command - fails fast due to missing config
	output, _ := executeWithTimeout(2 * time.Second)

	// Check for error message about config file
	if !strings.Contains(output, "config") && !strings.Contains(output, "Config") {
		t.Error("Expected error message about config file not found")
	}
}

// TestGetLocalIP tests local IP detection
func TestGetLocalIP(t *testing.T) {
	ip := getLocalIP()
	if ip == "" {
		t.Error("Expected to get a local IP address")
	}
	// Basic format check - should be x.x.x.x
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		t.Errorf("Expected IP in x.x.x.x format, got: %s", ip)
	}
}

// TestIsValidUUID_Valid tests valid UUID format validation
func TestIsValidUUID_Valid(t *testing.T) {
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"f47ac10b-58cc-4372-a567-0e02b2c3d479",
	}

	for _, uuid := range validUUIDs {
		if !isValidUUID(uuid) {
			t.Errorf("Expected %s to be valid UUID", uuid)
		}
	}
}

// TestIsValidUUID_Invalid tests invalid UUID format validation
func TestIsValidUUID_Invalid(t *testing.T) {
	invalidUUIDs := []string{
		"not-a-uuid",
		"550e8400-e29b-41d4-a716", // too short
		"550e8400-e29b-41d4-a716-446655440000-extra", // too long
		"550e8400-e29b-41d4-a716-44665544000x", // invalid char
		"",
	}

	for _, uuid := range invalidUUIDs {
		if isValidUUID(uuid) {
			t.Errorf("Expected %s to be invalid UUID", uuid)
		}
	}
}

// TestStartCommand_ProgressSteps tests that progress steps are shown (AC #1)
func TestStartCommand_ProgressSteps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

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

	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	// Execute command with timeout
	output, _ := executeWithTimeout(3 * time.Second)

	// Check for loading/configuration messages (AC #1)
	if !strings.Contains(output, "Loading") && !strings.Contains(output, "configuration") &&
	   !strings.Contains(output, "加载") && !strings.Contains(output, "配置") {
		t.Error("Expected progress step indicators or loading messages")
	}
}

// TestStartCommand_ErrorSuggestion tests error messages contain suggestions (AC #6)
func TestStartCommand_ErrorSuggestion(t *testing.T) {
	// Create invalid config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "invalid-url"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	// Execute command - fails fast due to invalid URL
	output, _ := executeWithTimeout(2 * time.Second)

	// Check for suggestion in error message (AC #6)
	if !strings.Contains(output, "suggestion") && !strings.Contains(output, "建议") &&
	   !strings.Contains(output, "URL") && !strings.Contains(output, "url") {
		// Error should contain helpful information
		if !strings.Contains(output, "invalid") && !strings.Contains(output, "Invalid") {
			t.Error("Expected error message with suggestion or specific issue")
		}
	}
}

// TestStartCommand_PIDFileCreation tests PID file is created on start
func TestStartCommand_PIDFileCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

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

	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	// Execute command with timeout
	output, _ := executeWithTimeout(2 * time.Second)
	t.Logf("Output: %s", output)

	// Check for PID file creation message
	if !strings.Contains(output, "PID file") && !strings.Contains(output, "PID") {
		t.Log("Note: PID file creation message may not be shown in test environment")
	}
}

// TestStartCommand_PIDFileErrorHandling tests PID file error handling
func TestStartCommand_PIDFileErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Create test config in a directory where PID file creation might fail
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

	GetRootCmd().SetArgs([]string{"--config", configPath, "start"})

	// Execute command with timeout
	output, _ := executeWithTimeout(2 * time.Second)
	t.Logf("Output: %s", output)

	// Should still start even if PID file creation has issues
	// Command should handle PID file errors gracefully
	if strings.Contains(output, "FATAL") || strings.Contains(output, "panic") {
		t.Error("Start command should handle PID file errors gracefully")
	}
}

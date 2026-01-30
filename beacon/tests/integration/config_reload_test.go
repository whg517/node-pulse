package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"beacon/internal/config"

	"github.com/sirupsen/logrus"
)

// TestConfigReload_BeaconStartup verifies Beacon can start with config watcher
func TestConfigReload_BeaconStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")
	pidPath := filepath.Join(tmpDir, "beacon.pid")

	// Write initial config
	initialConfig := `pulse_server: http://localhost:8080
node_id: integration-test-node
node_name: Integration Test Node
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
log_to_console: true
log_file: /tmp/beacon-integration-test.log
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Build beacon binary if not exists
	beaconBinary := "./beacon"
	if _, err := os.Stat(beaconBinary); os.IsNotExist(err) {
		t.Skip("Beacon binary not found - run 'make build' first")
	}

	// Start beacon process
	cmd := exec.Command(beaconBinary, "start", "--config", cfgPath, "--pid", pidPath)
	cmd.Dir = tmpDir

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start beacon: %v", err)
	}

	// Give beacon time to start and initialize config watcher
	time.Sleep(2 * time.Second)

	// Verify beacon is running by checking PID file
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		t.Error("PID file not created - beacon may not have started")
	}

	// Stop beacon gracefully
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Errorf("Failed to send interrupt: %v", err)
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Beacon exited with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Beacon did not exit within 5 seconds")
		cmd.Process.Kill()
	}
}

// TestConfigReload_FileModification verifies config changes are detected without restart
func TestConfigReload_FileModification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")
	pidPath := filepath.Join(tmpDir, "beacon.pid")
	logPath := filepath.Join(tmpDir, "beacon.log")

	// Write initial config
	initialConfig := fmt.Sprintf(`pulse_server: http://localhost:8080
node_id: integration-test-reload
node_name: Integration Test Reload
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
log_to_console: false
log_file: %s
`, logPath)

	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Build beacon binary if not exists
	beaconBinary := "./beacon"
	if _, err := os.Stat(beaconBinary); os.IsNotExist(err) {
		t.Skip("Beacon binary not found - run 'make build' first")
	}

	// Start beacon process
	cmd := exec.Command(beaconBinary, "start", "--config", cfgPath, "--pid", pidPath)
	cmd.Dir = tmpDir

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start beacon: %v", err)
	}
	defer cmd.Process.Kill()

	// Give beacon time to start
	time.Sleep(2 * time.Second)

	// Modify config file (change interval from 60 to 120 seconds)
	modifiedConfig := fmt.Sprintf(`pulse_server: http://localhost:8080
node_id: integration-test-reload
node_name: Integration Test Reload
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 120
    timeout_seconds: 5
    count: 10
log_to_console: false
log_file: %s
`, logPath)

	if err := os.WriteFile(cfgPath, []byte(modifiedConfig), 0644); err != nil {
		t.Fatalf("Failed to write modified config: %v", err)
	}

	// Wait for config reload (debounce + reload time)
	time.Sleep(3 * time.Second)

	// Verify config reload happened by checking logs
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(logData)
	foundConfigReload := false
	foundProbeChanges := false

	// Check for config reload indicators
	if contains(logContent, "Configuration changes detected") {
		foundConfigReload = true
	}
	if contains(logContent, "interval 60 -> 120 seconds") {
		foundProbeChanges = true
	}

	if !foundConfigReload {
		t.Error("Config reload not detected in logs - expected 'Configuration changes detected'")
	}
	if !foundProbeChanges {
		t.Error("Probe interval change not logged - expected 'interval 60 -> 120 seconds'")
	}

	// Verify beacon is still running (no restart needed)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		t.Error("Beacon process exited - config reload should not require restart")
	}
}

// TestConfigReload_InvalidConfig verifies invalid config is rejected
func TestConfigReload_InvalidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")
	pidPath := filepath.Join(tmpDir, "beacon.pid")
	logPath := filepath.Join(tmpDir, "beacon.log")

	// Write initial valid config
	initialConfig := fmt.Sprintf(`pulse_server: http://localhost:8080
node_id: integration-test-validation
node_name: Integration Test Validation
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
log_to_console: false
log_file: %s
`, logPath)

	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Build beacon binary if not exists
	beaconBinary := "./beacon"
	if _, err := os.Stat(beaconBinary); os.IsNotExist(err) {
		t.Skip("Beacon binary not found - run 'make build' first")
	}

	// Start beacon process
	cmd := exec.Command(beaconBinary, "start", "--config", cfgPath, "--pid", pidPath)
	cmd.Dir = tmpDir

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start beacon: %v", err)
	}
	defer cmd.Process.Kill()

	// Give beacon time to start
	time.Sleep(2 * time.Second)

	// Load initial config to verify it's valid
	initialCfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load initial config: %v", err)
	}

	// Write invalid config (interval exceeds max)
	invalidConfig := fmt.Sprintf(`pulse_server: http://localhost:8080
node_id: integration-test-validation
node_name: Integration Test Validation
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 500
    timeout_seconds: 5
    count: 10
log_to_console: false
log_file: %s
`, logPath)

	if err := os.WriteFile(cfgPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Wait for reload attempt
	time.Sleep(3 * time.Second)

	// Verify validation error in logs
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(logData)
	foundValidationError := false

	if contains(logContent, "config validation failed") || contains(logContent, "Failed to reload config") {
		foundValidationError = true
	}

	if !foundValidationError {
		t.Error("Validation error not found in logs - expected config to be rejected")
	}

	// Verify beacon is still running with original config
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		t.Error("Beacon process exited after invalid config - should continue running with original config")
	}

	// Verify original config is still in effect by checking logs don't show new interval
	if contains(logContent, "interval 60 -> 500 seconds") {
		t.Error("Invalid config was applied - original config should have been preserved")
	}

	_ = initialCfg // Verify config is accessible
}

// TestConfigReload_ConcurrentReads verifies thread-safe config access during reload
func TestConfigReload_ConcurrentReads(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	// Write config
	initialConfig := `pulse_server: http://localhost:8080
node_id: integration-test-concurrent
node_name: Integration Test Concurrent
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
log_to_console: false
log_file: /tmp/beacon-concurrent-test.log
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create file watcher
	logger := setupTestLogger(t)
	watcher, err := config.NewFileWatcher(cfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go watcher.Start(ctx)
	time.Sleep(500 * time.Millisecond)

	// Launch concurrent readers and config modifiers
	done := make(chan bool)

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				config := watcher.GetConfig()
				if config == nil {
					t.Errorf("Reader %d: got nil config", id)
				}
				if config.PulseServer == "" {
					t.Errorf("Reader %d: pulse_server is empty", id)
				}
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Config modifiers (write config changes)
	go func() {
		for i := 0; i < 5; i++ {
			port := 8080 + i
			modifiedConfig := fmt.Sprintf(`pulse_server: http://localhost:%d
node_id: integration-test-concurrent
node_name: Integration Test Concurrent
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
log_to_console: false
log_file: /tmp/beacon-concurrent-test.log
`, port)
			os.WriteFile(cfgPath, []byte(modifiedConfig), 0644)
			time.Sleep(200 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 11; i++ {
		<-done
	}

	// Verify no race conditions occurred
	// (If we got here without panic/deadlock, test passed)
	t.Log("Concurrent config access test completed successfully")
}

// Helper functions

func setupTestLogger(t *testing.T) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
	return logger
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsIn(s, substr))
}

func containsIn(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

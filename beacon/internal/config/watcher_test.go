package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// TestFileWatcher_StartStop tests starting and stopping the file watcher
func TestFileWatcher_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	// Write initial config
	initialConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize logger for testing
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	watcher, err := NewFileWatcher(cfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Test Start
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start watcher in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- watcher.Start(ctx)
	}()

	// Wait a bit for startup
	time.Sleep(500 * time.Millisecond)

	// Cancel context to stop watcher
	cancel()

	// Wait for Start to return
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Start() returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start() did not return after context cancellation")
	}
}

// TestFileWatcher_ConfigReload tests configuration file change detection and reload
func TestFileWatcher_ConfigReload(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	// Write initial config
	initialConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize logger for testing
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)

	watcher, err := NewFileWatcher(cfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Track reload callbacks
	var mu sync.Mutex
	reloadCount := 0
	var lastConfig *Config

	watcher.OnReload(func(newConfig *Config, changes []string) error {
		mu.Lock()
		defer mu.Unlock()
		reloadCount++
		lastConfig = newConfig
		return nil
	})

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go watcher.Start(ctx)

	// Wait for watcher to start
	time.Sleep(500 * time.Millisecond)

	// Modify config file
	newConfig := `pulse_server: http://localhost:9090
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 120
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(newConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for reload (debounce delay is 1 second)
	time.Sleep(3 * time.Second)

	// Verify reload occurred
	mu.Lock()
	if reloadCount != 1 {
		t.Errorf("Expected 1 reload, got %d", reloadCount)
	}
	if lastConfig == nil {
		t.Fatal("Last config is nil")
	}
	if lastConfig.PulseServer != "http://localhost:9090" {
		t.Errorf("Expected pulse_server http://localhost:9090, got %s", lastConfig.PulseServer)
	}
	mu.Unlock()
}

// TestFileWatcher_ValidationFailure tests that invalid config doesn't replace valid config
func TestFileWatcher_ValidationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	// Write initial valid config
	initialConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)

	watcher, err := NewFileWatcher(cfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	var mu sync.Mutex
	reloadCount := 0

	watcher.OnReload(func(newConfig *Config, changes []string) error {
		mu.Lock()
		defer mu.Unlock()
		reloadCount++
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go watcher.Start(ctx)
	time.Sleep(500 * time.Millisecond)

	// Write invalid config (interval_seconds exceeds limit)
	invalidConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 500
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for reload attempt
	time.Sleep(3 * time.Second)

	// Verify no reload occurred (validation should have failed)
	mu.Lock()
	if reloadCount != 0 {
		t.Errorf("Expected 0 reloads (validation failed), got %d", reloadCount)
	}
	mu.Unlock()

	// Verify original config is still in place
	currentConfig := watcher.GetConfig()
	if currentConfig.PulseServer != "http://localhost:8080" {
		t.Errorf("Original config should remain unchanged after validation failure, got pulse_server: %s", currentConfig.PulseServer)
	}
}

// TestFileWatcher_Debounce tests that rapid file changes only trigger one reload
func TestFileWatcher_Debounce(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	initialConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)

	watcher, err := NewFileWatcher(cfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	var mu sync.Mutex
	reloadCount := 0

	watcher.OnReload(func(newConfig *Config, changes []string) error {
		mu.Lock()
		defer mu.Unlock()
		reloadCount++
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go watcher.Start(ctx)
	time.Sleep(500 * time.Millisecond)

	// Rapidly modify config file 3 times (each with a different pulse_server)
	for i := 0; i < 3; i++ {
		port := 8080 + i
		newConfig := fmt.Sprintf(`pulse_server: http://localhost:%d
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`, port)
		if err := os.WriteFile(cfgPath, []byte(newConfig), 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond) // 100ms between changes
	}

	// Wait for debounce delay
	time.Sleep(3 * time.Second)

	// Verify only one reload occurred
	mu.Lock()
	if reloadCount != 1 {
		t.Errorf("Expected 1 reload (debounced), got %d", reloadCount)
	}
	mu.Unlock()
}

// TestFileWatcher_ConcurrentReads tests thread-safe config reads
func TestFileWatcher_ConcurrentReads(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	initialConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)

	watcher, err := NewFileWatcher(cfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go watcher.Start(ctx)
	time.Sleep(500 * time.Millisecond)

	// Launch multiple goroutines reading config concurrently
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				config := watcher.GetConfig()
				if config == nil {
					t.Error("GetConfig returned nil")
				}
				if config.PulseServer == "" {
					t.Error("PulseServer is empty")
				}
			}
		}()
	}

	wg.Wait()
}

// TestFileWatcher_Rollback tests config rollback on callback failure
func TestFileWatcher_Rollback(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	initialConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)

	watcher, err := NewFileWatcher(cfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Callback that fails on second call
	callCount := 0
	watcher.OnReload(func(newConfig *Config, changes []string) error {
		callCount++
		if callCount == 2 {
			return fmt.Errorf("simulated callback failure")
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go watcher.Start(ctx)
	time.Sleep(500 * time.Millisecond)

	// First successful reload
	newConfig1 := `pulse_server: http://localhost:9090
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	os.WriteFile(cfgPath, []byte(newConfig1), 0644)
	time.Sleep(2 * time.Second)

	// Second reload that will fail
	newConfig2 := `pulse_server: http://localhost:10100
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	os.WriteFile(cfgPath, []byte(newConfig2), 0644)
	time.Sleep(2 * time.Second)

	// Verify config rolled back to previous state
	currentConfig := watcher.GetConfig()
	if currentConfig.PulseServer != "http://localhost:9090" {
		t.Errorf("Config should have rolled back to http://localhost:9090, got %s", currentConfig.PulseServer)
	}
}

// TestFileWatcher_CustomConfigPath tests custom config file path support
func TestFileWatcher_CustomConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	customCfgPath := filepath.Join(tmpDir, "custom-beacon.yaml")

	initialConfig := `pulse_server: http://localhost:8080
node_id: test-node-1
node_name: Test Node 1
probes:
  - type: tcp_ping
    target: example.com
    port: 443
    interval: 60
    timeout_seconds: 5
    count: 10
`
	if err := os.WriteFile(customCfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(customCfgPath)
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)

	watcher, err := NewFileWatcher(customCfgPath, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file watcher with custom path: %v", err)
	}

	if watcher.GetConfigPath() != customCfgPath {
		t.Errorf("Expected config path %s, got %s", customCfgPath, watcher.GetConfigPath())
	}

	// Verify watcher works with custom path
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- watcher.Start(ctx)
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Start() with custom path failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start() did not return")
	}
}

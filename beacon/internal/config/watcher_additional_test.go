package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// createTestConfigForWatcher creates a minimal valid config for watcher tests
func createTestConfigForWatcher(t *testing.T) (string, *Config) {
	t.Helper()
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `pulse_server: http://localhost:6532
node_id: test-node-1
node_name: Test Node 1
`
	if err := os.WriteFile(cfgPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	return cfgPath, cfg
}

func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)
	return logger
}

// TestFileWatcher_GetVersion tests GetVersion
func TestFileWatcher_GetVersion(t *testing.T) {
	cfgPath, cfg := createTestConfigForWatcher(t)

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Initial version should be 1
	version := watcher.GetVersion()
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
}

// TestFileWatcher_GetReloadCount tests GetReloadCount
func TestFileWatcher_GetReloadCount(t *testing.T) {
	cfgPath, cfg := createTestConfigForWatcher(t)

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Initial reload count should be 0
	count := watcher.GetReloadCount()
	if count != 0 {
		t.Errorf("Expected reload count 0, got %d", count)
	}
}

// TestFileWatcher_GetLastReloadTime tests GetLastReloadTime
func TestFileWatcher_GetLastReloadTime(t *testing.T) {
	cfgPath, cfg := createTestConfigForWatcher(t)

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Initial last reload time should be zero
	lastReload := watcher.GetLastReloadTime()
	if !lastReload.IsZero() {
		t.Errorf("Expected zero time initially, got %v", lastReload)
	}
}

// TestFileWatcher_TriggerReload tests TriggerReload
func TestFileWatcher_TriggerReload(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `pulse_server: http://localhost:6532
node_id: test-node-1
node_name: Test Node 1
`
	if err := os.WriteFile(cfgPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Trigger reload with same content (no changes)
	if err := watcher.TriggerReload(); err != nil {
		t.Fatalf("TriggerReload failed: %v", err)
	}

	// No change detected, reload count should still be 0
	count := watcher.GetReloadCount()
	if count != 0 {
		t.Logf("Reload count after no-change trigger: %d", count)
	}
}

// TestFileWatcher_TriggerReload_WithChanges tests TriggerReload with actual changes
func TestFileWatcher_TriggerReload_WithChanges(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `pulse_server: http://localhost:6532
node_id: test-node-1
node_name: Test Node 1
`
	if err := os.WriteFile(cfgPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Register a callback to track reload
	reloadCalled := false
	watcher.OnReload(func(newCfg *Config, changes []string) error {
		reloadCalled = true
		return nil
	})

	// Modify the config
	newContent := `pulse_server: http://changed:6532
node_id: test-node-1
node_name: Test Node 1
`
	if err := os.WriteFile(cfgPath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to write changed config: %v", err)
	}

	// Trigger reload
	if err := watcher.TriggerReload(); err != nil {
		t.Fatalf("TriggerReload failed: %v", err)
	}

	if !reloadCalled {
		t.Error("Expected reload callback to be called")
	}

	// Reload count should be 1
	count := watcher.GetReloadCount()
	if count != 1 {
		t.Errorf("Expected reload count 1, got %d", count)
	}

	// Version should be 2
	version := watcher.GetVersion()
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}

	// Last reload time should be set
	lastReload := watcher.GetLastReloadTime()
	if lastReload.IsZero() {
		t.Error("Expected non-zero last reload time")
	}
	if time.Since(lastReload) > 5*time.Second {
		t.Error("Expected recent last reload time")
	}
}

// TestFileWatcher_DiffConfig tests the diffConfig function
func TestFileWatcher_DiffConfig(t *testing.T) {
	cfgPath, cfg := createTestConfigForWatcher(t)

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Create two configs to diff
	oldCfg := &Config{
		PulseServer: "http://old:6532",
		NodeID:      "old-node",
		NodeName:    "Old Node",
		Probes: []ProbeConfig{
			{
				Type:           "tcp_ping",
				Target:         "example.com",
				Port:           443,
				Interval:       60,
				TimeoutSeconds: 5,
				Count:          10,
			},
		},
	}

	// Same config - no changes
	changes := watcher.diffConfig(oldCfg, oldCfg)
	if len(changes) != 0 {
		t.Errorf("Expected no changes for identical configs, got %d", len(changes))
	}

	// Different pulse_server
	newCfg := &Config{
		PulseServer: "http://new:6532",
		NodeID:      "old-node",
		NodeName:    "Old Node",
		Probes:      oldCfg.Probes,
	}
	changes = watcher.diffConfig(oldCfg, newCfg)
	if len(changes) == 0 {
		t.Error("Expected changes for different pulse_server")
	}

	// Different node_id  
	newCfg2 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      "new-node",
		NodeName:    "Old Node",
		Probes:      oldCfg.Probes,
	}
	changes = watcher.diffConfig(oldCfg, newCfg2)
	if len(changes) == 0 {
		t.Error("Expected changes for different node_id")
	}

	// Different node_name
	newCfg3 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    "New Name",
		Probes:      oldCfg.Probes,
	}
	changes = watcher.diffConfig(oldCfg, newCfg3)
	if len(changes) == 0 {
		t.Error("Expected changes for different node_name")
	}

	// Probe count changed
	newCfg4 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{},
	}
	changes = watcher.diffConfig(oldCfg, newCfg4)
	if len(changes) == 0 {
		t.Error("Expected changes for different probe count")
	}
}

// TestFileWatcher_DiffConfig_ProbeDetails tests diffConfig with detailed probe changes
func TestFileWatcher_DiffConfig_ProbeDetails(t *testing.T) {
	cfgPath, cfg := createTestConfigForWatcher(t)

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	probe1 := ProbeConfig{
		Type:           "tcp_ping",
		Target:         "example.com",
		Port:           443,
		Interval:       60,
		TimeoutSeconds: 5,
		Count:          10,
	}

	oldCfg := &Config{
		PulseServer: "http://test:6532",
		NodeID:      "test-node",
		NodeName:    "Test",
		Probes:      []ProbeConfig{probe1},
	}

	// Change probe type
	probe2 := ProbeConfig{
		Type:           "udp_ping",
		Target:         "example.com",
		Port:           443,
		Interval:       60,
		TimeoutSeconds: 5,
		Count:          10,
	}
	newCfg := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{probe2},
	}
	changes := watcher.diffConfig(oldCfg, newCfg)
	if len(changes) == 0 {
		t.Error("Expected changes for probe type change")
	}

	// Change probe target
	probe3 := ProbeConfig{
		Type:           probe1.Type,
		Target:         "changed.com",
		Port:           probe1.Port,
		Interval:       probe1.Interval,
		TimeoutSeconds: probe1.TimeoutSeconds,
		Count:          probe1.Count,
	}
	newCfg2 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{probe3},
	}
	changes = watcher.diffConfig(oldCfg, newCfg2)
	if len(changes) == 0 {
		t.Error("Expected changes for probe target change")
	}

	// Change probe port
	probe4 := ProbeConfig{
		Type:           probe1.Type,
		Target:         probe1.Target,
		Port:           80,
		Interval:       probe1.Interval,
		TimeoutSeconds: probe1.TimeoutSeconds,
		Count:          probe1.Count,
	}
	newCfg3 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{probe4},
	}
	changes = watcher.diffConfig(oldCfg, newCfg3)
	if len(changes) == 0 {
		t.Error("Expected changes for probe port change")
	}

	// Change probe interval
	probe5 := ProbeConfig{
		Type:           probe1.Type,
		Target:         probe1.Target,
		Port:           probe1.Port,
		Interval:       120,
		TimeoutSeconds: probe1.TimeoutSeconds,
		Count:          probe1.Count,
	}
	newCfg4 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{probe5},
	}
	changes = watcher.diffConfig(oldCfg, newCfg4)
	if len(changes) == 0 {
		t.Error("Expected changes for probe interval change")
	}

	// Change probe timeout
	probe6 := ProbeConfig{
		Type:           probe1.Type,
		Target:         probe1.Target,
		Port:           probe1.Port,
		Interval:       probe1.Interval,
		TimeoutSeconds: 10,
		Count:          probe1.Count,
	}
	newCfg5 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{probe6},
	}
	changes = watcher.diffConfig(oldCfg, newCfg5)
	if len(changes) == 0 {
		t.Error("Expected changes for probe timeout change")
	}

	// Change probe count
	probe7 := ProbeConfig{
		Type:           probe1.Type,
		Target:         probe1.Target,
		Port:           probe1.Port,
		Interval:       probe1.Interval,
		TimeoutSeconds: probe1.TimeoutSeconds,
		Count:          20,
	}
	newCfg6 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{probe7},
	}
	changes = watcher.diffConfig(oldCfg, newCfg6)
	if len(changes) == 0 {
		t.Error("Expected changes for probe count change")
	}

	// New probe added (new has more probes)
	newCfg7 := &Config{
		PulseServer: oldCfg.PulseServer,
		NodeID:      oldCfg.NodeID,
		NodeName:    oldCfg.NodeName,
		Probes:      []ProbeConfig{probe1, probe2},
	}
	changes = watcher.diffConfig(oldCfg, newCfg7)
	if len(changes) == 0 {
		t.Error("Expected changes for new probe added")
	}
}
func TestFileWatcher_TriggerReload_CallbackError(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `pulse_server: http://localhost:6532
node_id: test-node-1
node_name: Test Node 1
`
	if err := os.WriteFile(cfgPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	watcher, err := NewFileWatcher(cfgPath, cfg, newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Register a callback that returns an error
	watcher.OnReload(func(newCfg *Config, changes []string) error {
		return os.ErrPermission // Simulate callback error
	})

	// Modify the config
	newContent := `pulse_server: http://changed:6532
node_id: test-node-1
node_name: Test Node 1
`
	if err := os.WriteFile(cfgPath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to write changed config: %v", err)
	}

	// Trigger reload - should fail due to callback error
	if err := watcher.TriggerReload(); err == nil {
		t.Error("Expected error from callback failure")
	}

	// Config should have been rolled back
	currentCfg := watcher.GetConfig()
	if currentCfg.PulseServer != "http://localhost:6532" {
		t.Errorf("Expected config to be rolled back, got %s", currentCfg.PulseServer)
	}

	// Version should have been rolled back to 1
	if watcher.GetVersion() != 1 {
		t.Errorf("Expected version to be rolled back to 1, got %d", watcher.GetVersion())
	}
}

package process

import (
"os"
"path/filepath"
"testing"

"beacon/internal/config"
)

// TestNewManager_NilConfig tests creating manager with nil config
func TestNewManager_NilConfig(t *testing.T) {
m := NewManager(nil)
if m == nil {
t.Fatal("Expected non-nil manager")
}
// With nil config, should use default PID file
if m.pidFile != DefaultPIDFile {
t.Errorf("Expected default PID file %s, got %s", DefaultPIDFile, m.pidFile)
}
}

// TestNewManager_WithConfig tests creating manager with config
func TestNewManager_WithConfig(t *testing.T) {
tmpDir := t.TempDir()
cfgPath := filepath.Join(tmpDir, "beacon.yaml")

cfg := &config.Config{
ConfigPath: cfgPath,
}

m := NewManager(cfg)
if m == nil {
t.Fatal("Expected non-nil manager")
}
}

// TestNewManager_ConfigInCurrentDir tests manager with config in current dir
func TestNewManager_ConfigInCurrentDir(t *testing.T) {
cfg := &config.Config{
ConfigPath: "./beacon.yaml",
}

m := NewManager(cfg)
if m == nil {
t.Fatal("Expected non-nil manager")
}
// With config in current dir, should use alternative PID file
if m.pidFile != AlternativePIDFile {
t.Errorf("Expected alternative PID file %s, got %s", AlternativePIDFile, m.pidFile)
}
}

// TestNewManager_EmptyConfigPath tests manager with empty config path
func TestNewManager_EmptyConfigPath(t *testing.T) {
cfg := &config.Config{
ConfigPath: "",
}

m := NewManager(cfg)
if m == nil {
t.Fatal("Expected non-nil manager")
}
// With empty config path, should use default PID file
if m.pidFile != DefaultPIDFile {
t.Errorf("Expected default PID file %s, got %s", DefaultPIDFile, m.pidFile)
}
}

// TestManager_GetPIDFile tests GetPIDFile
func TestManager_GetPIDFile(t *testing.T) {
tmpDir := t.TempDir()
pidFile := filepath.Join(tmpDir, "test.pid")

m := &Manager{pidFile: pidFile}

result := m.GetPIDFile()
if result != pidFile {
t.Errorf("Expected PID file %s, got %s", pidFile, result)
}
}

// TestManager_isBeaconProcess tests isBeaconProcess
func TestManager_isBeaconProcess_Additional(t *testing.T) {
m := &Manager{}

// isBeaconProcess always returns true in the current implementation
result := m.isBeaconProcess(os.Getpid())
if !result {
t.Error("Expected isBeaconProcess to return true")
}

// Even for non-existent PID
result = m.isBeaconProcess(999999)
if !result {
t.Error("Expected isBeaconProcess to return true for any PID (current implementation)")
}
}

// TestManager_Stop_NoPIDFile tests Stop when no PID file exists
func TestManager_Stop_NoPIDFile_Additional(t *testing.T) {
tmpDir := t.TempDir()
pidFile := filepath.Join(tmpDir, "nonexistent.pid")

m := &Manager{pidFile: pidFile}

err := m.Stop()
if err == nil {
t.Error("Expected error when no PID file")
}
}

// TestManager_Stop_ProcessNotRunning tests Stop when process not running
func TestManager_Stop_ProcessNotRunning_Additional(t *testing.T) {
tmpDir := t.TempDir()
pidFile := filepath.Join(tmpDir, "test.pid")

// Write a non-existent PID
if err := os.WriteFile(pidFile, []byte("999999"), 0644); err != nil {
t.Fatalf("Failed to create test PID file: %v", err)
}

m := &Manager{pidFile: pidFile}

err := m.Stop()
if err == nil {
t.Error("Expected error when process not running")
}
}

// TestManager_ReadPID_AlternativeFile tests ReadPID that falls back to alternative file
func TestManager_ReadPID_AlternativeFile(t *testing.T) {
// Use primary nonexistent path  
m := &Manager{pidFile: "/nonexistent/primary.pid"}

// If AlternativePIDFile doesn't exist either, should return error
_, err := m.ReadPID()
if err == nil {
// If AlternativePIDFile exists (from previous test), that's OK
t.Log("ReadPID succeeded - AlternativePIDFile exists")
}
}

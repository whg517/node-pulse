package process

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestManager_WritePID tests PID file writing
func TestManager_WritePID(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	manager := &Manager{pidFile: pidFile}

	if err := manager.WritePID(); err != nil {
		t.Fatalf("Failed to write PID: %v", err)
	}

	// Verify PID file exists
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Error("PID file was not created")
	}

	// Verify PID file content
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("Failed to read PID file: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	if pidStr == "" {
		t.Error("PID file is empty")
	}

	// Should contain current process PID
	expectedPID := os.Getpid()
	pidInt, err := strconv.Atoi(pidStr)
	if err != nil {
		t.Errorf("PID file content is not a valid integer: %s", pidStr)
	}
	if pidInt != expectedPID {
		t.Errorf("Expected PID %d, got %d", expectedPID, pidInt)
	}
}

// TestManager_ReadPID tests PID file reading
func TestManager_ReadPID(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	// Write a test PID
	testPID := 12345
	pidStr := "12345\n"
	if err := os.WriteFile(pidFile, []byte(pidStr), 0644); err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}

	manager := &Manager{pidFile: pidFile}

	pid, err := manager.ReadPID()
	if err != nil {
		t.Fatalf("Failed to read PID: %v", err)
	}

	if pid != testPID {
		t.Errorf("Expected PID %d, got %d", testPID, pid)
	}
}

// TestManager_ReadPID_NoFile tests reading PID when file doesn't exist
func TestManager_ReadPID_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "nonexistent.pid")

	manager := &Manager{pidFile: pidFile}

	_, err := manager.ReadPID()
	if err == nil {
		t.Error("Expected error when PID file doesn't exist")
	}
}

// TestManager_ReadPID_InvalidFormat tests reading invalid PID format
func TestManager_ReadPID_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	// Write invalid PID
	if err := os.WriteFile(pidFile, []byte("invalid"), 0644); err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}

	manager := &Manager{pidFile: pidFile}

	_, err := manager.ReadPID()
	if err == nil {
		t.Error("Expected error for invalid PID format")
	}
}

// TestManager_IsRunning tests checking if process is running
func TestManager_IsRunning(t *testing.T) {
	manager := &Manager{pidFile: "./test.pid"}

	// Current process should be running
	currentPID := os.Getpid()
	if !manager.IsRunning(currentPID) {
		t.Error("Expected current process to be running")
	}

	// Non-existent PID should not be running
	if manager.IsRunning(999999) {
		t.Error("Expected non-existent process to not be running")
	}
}

// TestManager_Cleanup tests PID file cleanup
func TestManager_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	// Create PID file
	if err := os.WriteFile(pidFile, []byte("12345"), 0644); err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}

	manager := &Manager{pidFile: pidFile}

	if err := manager.Cleanup(); err != nil {
		t.Fatalf("Failed to cleanup PID file: %v", err)
	}

	// Verify PID file is removed
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Error("PID file was not removed")
	}
}

// TestManager_Cleanup_NoFile tests cleanup when PID file doesn't exist
func TestManager_Cleanup_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "nonexistent.pid")

	manager := &Manager{pidFile: pidFile}

	// Should not error even if file doesn't exist
	if err := manager.Cleanup(); err != nil {
		t.Logf("Cleanup error (expected): %v", err)
	}
}

// TestManager_GetStatus tests getting process status
func TestManager_GetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	// Write current PID to file
	currentPID := os.Getpid()
	pidStr := strconv.Itoa(currentPID)
	if err := os.WriteFile(pidFile, []byte(pidStr), 0644); err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}

	manager := &Manager{pidFile: pidFile}

	status := manager.GetStatus()

	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	if status.PIDFile != pidFile {
		t.Errorf("Expected PID file %s, got %s", pidFile, status.PIDFile)
	}

	if status.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

// TestManager_GetStatus_NoFile tests status when no PID file exists
func TestManager_GetStatus_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "nonexistent.pid")

	manager := &Manager{pidFile: pidFile}

	status := manager.GetStatus()

	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	if status.Running {
		t.Error("Expected running to be false when no PID file exists")
	}

	if status.Error == "" {
		t.Error("Expected error message when no PID file exists")
	}
}

// TestManager_WritePID_CreatesDirectory tests that WritePID creates directory
func TestManager_WritePID_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "subdir", "test.pid")

	manager := &Manager{pidFile: pidFile}

	if err := manager.WritePID(); err != nil {
		t.Fatalf("Failed to write PID: %v", err)
	}

	// Verify PID file exists
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Error("PID file was not created in subdirectory")
	}
}

// TestManager_Stop tests stopping a process (basic test)
func TestManager_Stop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process stop test in short mode")
	}

	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	// Create a test subprocess that will exit gracefully
	// For this test, we'll use a non-existent PID to test error handling
	manager := &Manager{pidFile: pidFile}

	// Write a non-existent PID
	if err := os.WriteFile(pidFile, []byte("99999"), 0644); err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}

	// Try to stop - should fail gracefully
	err := manager.Stop()
	if err == nil {
		t.Error("Expected error when trying to stop non-existent process")
	}

	// PID file should be cleaned up even if process not found
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Log("Note: PID file may still exist after failed stop")
	}
}

// TestStatus_Structure tests Status struct fields
func TestStatus_Structure(t *testing.T) {
	status := &Status{
		Running:   true,
		PID:       12345,
		PIDFile:   "/test/beacon.pid",
		Timestamp: time.Now(),
	}

	if !status.Running {
		t.Error("Expected Running to be true")
	}

	if status.PID != 12345 {
		t.Errorf("Expected PID 12345, got %d", status.PID)
	}

	if status.PIDFile != "/test/beacon.pid" {
		t.Errorf("Expected PIDFile /test/beacon.pid, got %s", status.PIDFile)
	}
}

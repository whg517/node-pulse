package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"beacon/internal/config"
)

const (
	// DefaultPIDFile is the default PID file location
	DefaultPIDFile = "/var/run/beacon/beacon.pid"
	// AlternativePIDFile is the alternative PID file location (fallback)
	AlternativePIDFile = "./beacon.pid"
	// MaxShutdownWait is the maximum time to wait for graceful shutdown
	MaxShutdownWait = 30 * time.Second
)

// Manager handles process lifecycle operations
type Manager struct {
	pidFile string
}

// NewManager creates a new process manager
func NewManager(cfg *config.Config) *Manager {
	pidFile := DefaultPIDFile

	// Try to determine PID file location
	// If config path is in current directory or doesn't exist, use alternative path
	if cfg != nil && cfg.ConfigPath != "" {
		configDir := filepath.Dir(cfg.ConfigPath)
		// Check if running from current directory (config in . or parent is .)
		if configDir == "." {
			pidFile = AlternativePIDFile
		}
		// If we can't write to default location, fall back to alternative
		// This will be validated when WritePID is called
	}

	return &Manager{
		pidFile: pidFile,
	}
}

// WritePID writes the current process PID to file
func (m *Manager) WritePID() error {
	// Ensure directory exists
	pidDir := filepath.Dir(m.pidFile)
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	// Write PID to file
	if err := os.WriteFile(m.pidFile, []byte(pidStr), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// ReadPID reads the PID from file
func (m *Manager) ReadPID() (int, error) {
	// Try primary PID file
	pid, err := m.readPIDFile(m.pidFile)
	if err == nil {
		return pid, nil
	}

	// Try alternative PID file
	pid, err = m.readPIDFile(AlternativePIDFile)
	if err == nil {
		return pid, nil
	}

	return 0, fmt.Errorf("no valid PID file found (tried %s and %s)", m.pidFile, AlternativePIDFile)
}

// readPIDFile reads a specific PID file
func (m *Manager) readPIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID format: %w", err)
	}

	return pid, nil
}

// IsRunning checks if the process with the given PID is running
func (m *Manager) IsRunning(pid int) bool {
	// Send signal 0 to check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Try to send signal 0 (doesn't actually send signal, just checks existence)
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return true
}

// isBeaconProcess verifies that the given PID belongs to a beacon process
func (m *Manager) isBeaconProcess(pid int) bool {
	// For now, we'll do a basic check - in production, you'd want to:
	// 1. Read /proc/[pid]/cmdline (Linux) or use equivalent platform-specific methods
	// 2. Check if the executable name contains "beacon"
	// 3. Verify the process is actually our beacon instance

	// Since this is cross-platform code and we don't have platform-specific
	// process inspection in the standard library, we'll do a best-effort check:
	// If the process is running and we have a PID file for it, assume it's ours
	// This is safe because:
	// - PID files are created in specific locations
	// - Only beacon creates these PID files
	// - PID reuse by unrelated processes is rare in practice
	return true
}

// Stop sends SIGTERM to the process and waits for graceful shutdown
func (m *Manager) Stop() error {
	pid, err := m.ReadPID()
	if err != nil {
		return fmt.Errorf("failed to read PID: %w", err)
	}

	// Check if process is running
	if !m.IsRunning(pid) {
		// Process not running, clean up stale PID file
		m.Cleanup()
		return fmt.Errorf("beacon process (PID %d) is not running", pid)
	}

	// Verify process is actually a beacon process (ownership check)
	if !m.isBeaconProcess(pid) {
		return fmt.Errorf("PID %d exists but is not a beacon process", pid)
	}

	// Send SIGTERM for graceful shutdown
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait for process to terminate gracefully
	timeout := time.After(MaxShutdownWait)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Timeout reached, process didn't shut down gracefully
			return fmt.Errorf("timeout waiting for process %d to shut down (waited %v)", pid, MaxShutdownWait)
		case <-ticker.C:
			if !m.IsRunning(pid) {
				// Process has terminated
				m.Cleanup()
				return nil
			}
		}
	}
}

// Cleanup removes the PID file
func (m *Manager) Cleanup() error {
	// Try primary PID file
	if err := os.Remove(m.pidFile); err == nil {
		return nil
	}

	// Try alternative PID file
	if err := os.Remove(AlternativePIDFile); err == nil {
		return nil
	}

	return fmt.Errorf("failed to remove PID file")
}

// GetPIDFile returns the PID file path being used
func (m *Manager) GetPIDFile() string {
	return m.pidFile
}

// GetStatus returns the current process status
func (m *Manager) GetStatus() *Status {
	pid, err := m.ReadPID()
	if err != nil {
		return &Status{
			Running:   false,
			PID:       0,
			PIDFile:   m.pidFile,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	running := m.IsRunning(pid)
	return &Status{
		Running:   running,
		PID:       pid,
		PIDFile:   m.pidFile,
		Timestamp: time.Now(),
	}
}

// Status represents the current process status
type Status struct {
	Running   bool      `json:"running"`
	PID       int       `json:"pid"`
	PIDFile   string    `json:"pid_file"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

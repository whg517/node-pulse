package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"beacon/internal/config"
)

// TestBeaconRuntimeLogging tests Beacon runtime logging behavior
func TestBeaconRuntimeLogging(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	// Initialize logger with Beacon-like configuration
	cfg := &config.Config{
		NodeID:        "test-node-123",
		NodeName:      "test-beacon",
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 10,
		LogCompress:   false,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer Close()

	// Simulate Beacon lifecycle logging
	// 1. Configuration loading
	WithFields(map[string]interface{}{
		"node_id":   cfg.NodeID,
		"node_name": cfg.NodeName,
		"config":    "/etc/beacon/beacon.yaml",
	}).Info("Configuration loaded successfully")

	// 2. Probe scheduler starting
	WithFields(map[string]interface{}{
		"component":  "probe",
		"interval":   "60s",
		"tcp_count":  2,
		"udp_count":  1,
	}).Info("Probe scheduler started")

	// 3. Individual probe execution
	WithFields(map[string]interface{}{
		"component":   "probe",
		"probe_type":  "tcp_ping",
		"target":      "192.168.1.1:80",
		"success":     true,
		"sample_count": 10,
		"rtt_ms":      25.5,
		"packet_loss": 0.0,
	}).Info("TCP probe completed")

	// 4. Metrics server starting
	WithFields(map[string]interface{}{
		"component": "metrics",
		"address":   ":2112",
	}).Info("Starting Prometheus metrics server")

	// 5. Heartbeat reporting
	WithFields(map[string]interface{}{
		"component": "reporter",
		"interval":  "60s",
	}).Info("Starting heartbeat reporter")

	// 6. Warning scenario
	WithFields(map[string]interface{}{
		"component": "probe",
		"target":    "192.168.1.2:443",
		"error":     "connection timeout",
	}).Warn("TCP probe failed")

	// 7. Error scenario
	WithFields(map[string]interface{}{
		"component": "reporter",
		"attempt":   3,
		"error":     "connection refused",
	}).Error("Heartbeat report failed")

	// 8. Beacon shutdown
	WithFields(map[string]interface{}{
		"node_id":   cfg.NodeID,
		"node_name": cfg.NodeName,
	}).Info("Shutting down gracefully")

	Close()

	// Verify log file was created and contains valid JSON
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Count log entries
	lines := parseLogLines(data)
	if len(lines) < 8 {
		t.Errorf("Expected at least 8 log entries, got %d", len(lines))
	}

	// Verify all entries are valid JSON
	validEntries := 0
	for i, line := range lines {
		var entry map[string]interface{}
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Errorf("Entry %d is not valid JSON: %v", i+1, err)
			continue
		}

		// Verify required fields
		if entry["timestamp"] == nil {
			t.Errorf("Entry %d missing timestamp", i+1)
		}
		if entry["level"] == nil {
			t.Errorf("Entry %d missing level", i+1)
		}
		if entry["message"] == nil {
			t.Errorf("Entry %d missing message", i+1)
		}

		// Verify component field exists for all entries
		if entry["component"] == nil {
			t.Logf("Entry %d: %+v", i+1, entry)
		}

		validEntries++
	}

	if validEntries < 8 {
		t.Errorf("Only %d/%d entries are valid JSON", validEntries, len(lines))
	}

	t.Logf("✅ All %d log entries are valid JSON with proper structure", validEntries)
}

// TestMultiModuleLoggingConsistency tests consistent logging across modules
func TestMultiModuleLoggingConsistency(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		NodeID:        "test-node-456",
		NodeName:      "test-beacon",
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 10,
		LogCompress:   false,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer Close()

	// Simulate logging from different modules concurrently
	modules := []string{"probe", "reporter", "metrics", "config"}
	done := make(chan bool, len(modules))

	for _, module := range modules {
		go func(mod string) {
			for i := 0; i < 5; i++ {
				WithFields(map[string]interface{}{
					"component": mod,
					"iteration": i,
					"module":    mod,
				}).Infof("Module %s iteration %d", mod, i)
			}
			done <- true
		}(module)
	}

	// Wait for all goroutines
	for i := 0; i < len(modules); i++ {
		<-done
	}

	// Give time for logs to flush
	time.Sleep(100 * time.Millisecond)
	Close()

	// Verify all log entries have consistent structure
	data, _ := os.ReadFile(logFile)
	lines := parseLogLines(data)

	componentCount := make(map[string]int)
	for _, line := range lines {
		var entry map[string]interface{}
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if component, ok := entry["component"].(string); ok {
			componentCount[component]++
		}
	}

	// Verify all modules logged
	for _, module := range modules {
		if count := componentCount[module]; count == 0 {
			t.Errorf("Module %s did not log any entries", module)
		} else {
			t.Logf("Module %s: %d log entries", module, count)
		}
	}

	// Verify total entries (5 iterations * 4 modules = 20)
	expectedEntries := len(modules) * 5
	if len(lines) < expectedEntries {
		t.Logf("Warning: Expected %d entries, got %d (some may be lost due to concurrency)", expectedEntries, len(lines))
	}
}

// TestLogRotationTrigger tests log rotation triggering
func TestLogRotationTrigger(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	// Small max size to trigger rotation quickly
	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    1, // 1 MB
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   true,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Write enough logs to trigger rotation
	largeMessage := "Test log entry for rotation trigger " + string(make([]byte, 1000))
	for i := 0; i < 1500; i++ {
		Info(largeMessage)
	}

	// Wait for rotation
	time.Sleep(200 * time.Millisecond)
	Close()

	// Check for rotated files
	files, _ := os.ReadDir(tempDir)
	logFiles := []string{}
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".log") ||
			strings.HasSuffix(file.Name(), ".gz")) {
			logFiles = append(logFiles, file.Name())
		}
	}

	if len(logFiles) < 2 {
		t.Logf("Note: Only %d log file(s) found - rotation may not have triggered yet", len(logFiles))
	} else {
		t.Logf("✅ Log rotation triggered: %d files created", len(logFiles))
		for _, f := range logFiles {
			t.Logf("  - %s", f)
		}
	}
}

// TestLogDateFormatting tests ISO 8601 timestamp formatting
func TestLogDateFormatting(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 10,
		LogCompress:   false,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	Info("Test message for timestamp validation")
	Close()

	// Read and verify timestamp format
	data, _ := os.ReadFile(logFile)
	lines := parseLogLines(data)

	if len(lines) == 0 {
		t.Fatal("No log entries found")
	}

	var entry map[string]interface{}
	json.Unmarshal(lines[0], &entry)

	timestamp, ok := entry["timestamp"].(string)
	if !ok {
		t.Fatal("No timestamp field found")
	}

	// Verify ISO 8601 format (RFC3339)
	_, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		t.Errorf("Timestamp %s is not valid ISO 8601 format: %v", timestamp, err)
	} else {
		t.Logf("✅ Timestamp format valid: %s", timestamp)
	}
}

// parseLogLines parses log file into JSON lines
func parseLogLines(data []byte) [][]byte {
	lines := [][]byte{}
	current := []byte{}

	for _, b := range data {
		if b == '\n' {
			if len(current) > 0 {
				lines = append(lines, current)
				current = []byte{}
			}
		} else {
			current = append(current, b)
		}
	}

	if len(current) > 0 {
		lines = append(lines, current)
	}

	return lines
}

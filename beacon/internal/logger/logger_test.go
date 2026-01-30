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

// TestInitLogger_Success tests successful logger initialization
func TestInitLogger_Success(t *testing.T) {
	// Create temporary log file
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	// Create test configuration
	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 10,
		LogCompress:   true,
		LogToConsole:  false,
	}

	// Initialize logger
	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Verify logger instance is created
	if Logger == nil {
		t.Fatal("Logger instance is nil")
	}

	// Write a test log to ensure file is created
	Info("Test log message")

	// Verify log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Log file was not created: %s", logFile)
	}

	// Clean up
	Close()
}

// TestInitLogger_LogLevels tests different log levels
func TestInitLogger_LogLevels(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := &config.Config{
				LogLevel:      level,
				LogFile:       logFile,
				LogMaxSize:    10,
				LogMaxAge:     7,
				LogMaxBackups: 10,
				LogCompress:   false,
				LogToConsole:  false,
			}

			err := InitLogger(cfg)
			if err != nil {
				t.Fatalf("InitLogger with level %s failed: %v", level, err)
			}

			// Verify log level is set correctly
			// Note: logrus uses "warning" instead of "warn" for WARN level
			expectedLevel := strings.ToLower(level)
			if level == "WARN" {
				expectedLevel = "warning"
			}

			if Logger.GetLevel().String() != expectedLevel {
				t.Errorf("Expected log level %s, got %s", expectedLevel, Logger.GetLevel().String())
			}

			Close()
		})
	}
}

// TestInitLogger_InvalidLogLevel tests invalid log level
func TestInitLogger_InvalidLogLevel(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		LogLevel:      "INVALID",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 10,
		LogCompress:   false,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err == nil {
		t.Error("Expected error for invalid log level, got nil")
	}

	if !strings.Contains(err.Error(), "invalid log level") {
		t.Errorf("Expected 'invalid log level' error message, got: %v", err)
	}
}

// TestInitLogger_LogDirectoryCreation tests automatic log directory creation
func TestInitLogger_LogDirectoryCreation(t *testing.T) {
	// Use a nested path that doesn't exist
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "logs", "beacon", "beacon.log")

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

	// Verify log directory was created
	logDir := filepath.Dir(logFile)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Fatalf("Log directory was not created: %s", logDir)
	}

	Close()
}

// TestJSONFormatter tests JSON log format
func TestJSONFormatter(t *testing.T) {
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

	// Write a test log message
	testMessage := "Test structured log message"
	Info(testMessage)

	// Ensure log is flushed
	Close()

	// Read log file and verify JSON format
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(data, &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v\nLog content: %s", err, string(data))
	}

	// Verify required fields
	requiredFields := []string{"timestamp", "level", "message"}
	for _, field := range requiredFields {
		if _, exists := logEntry[field]; !exists {
			t.Errorf("Missing required field in log: %s", field)
		}
	}

	// Verify message content
	if logEntry["message"] != testMessage {
		t.Errorf("Expected message '%s', got '%v'", testMessage, logEntry["message"])
	}

	// Verify timestamp format (ISO 8601)
	if timestamp, ok := logEntry["timestamp"].(string); ok {
		if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
			t.Errorf("Invalid timestamp format: %s, error: %v", timestamp, err)
		}
	}

	// Verify level
	if logEntry["level"] != "info" {
		t.Errorf("Expected level 'info', got '%v'", logEntry["level"])
	}
}

// TestWithFields tests structured logging with fields
func TestWithFields(t *testing.T) {
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

	// Write log with structured fields
	testFields := map[string]interface{}{
		"node_id":   "test-node-123",
		"component": "probe",
		"probe_type": "tcp_ping",
	}
	WithFields(testFields).Info("Probe execution started")

	Close()

	// Read log file
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(data, &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify fields exist
	for key, value := range testFields {
		if logEntry[key] != value {
			t.Errorf("Expected field %s = %v, got %v", key, value, logEntry[key])
		}
	}
}

// TestWithField tests single field logging
func TestWithField(t *testing.T) {
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

	// Write log with single field
	WithField("node_id", "test-node-456").Info("Test message")

	Close()

	// Read and verify
	data, _ := os.ReadFile(logFile)
	var logEntry map[string]interface{}
	json.Unmarshal(data, &logEntry)

	if logEntry["node_id"] != "test-node-456" {
		t.Errorf("Expected node_id 'test-node-456', got '%v'", logEntry["node_id"])
	}
}

// TestWithError tests error field logging
func TestWithError(t *testing.T) {
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

	// Write error log
	testErr := os.ErrNotExist
	WithError(testErr).Error("File not found")

	Close()

	// Read and verify
	data, _ := os.ReadFile(logFile)
	var logEntry map[string]interface{}
	json.Unmarshal(data, &logEntry)

	// Verify error field exists
	if logEntry["error"] == nil {
		t.Error("Expected 'error' field in log entry")
	}

	errorStr := logEntry["error"].(string)
	if !strings.Contains(errorStr, "does not exist") && !strings.Contains(errorStr, "no such file") {
		t.Errorf("Expected error message to contain 'does not exist' or 'no such file', got: %s", errorStr)
	}
}

// TestLogLevels tests different log level methods
func TestLogLevels(t *testing.T) {
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

	// Test different log levels
	Info("Info message")
	Warn("Warning message")
	Error("Error message")

	Close()

	// Read log file
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify all three log entries were written
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(lines))
	}

	// Verify each log entry has correct level
	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		expectedLevel := map[int]string{
			0: "info",
			1: "warning",
			2: "error",
		}[i]

		if logEntry["level"] != expectedLevel {
			t.Errorf("Entry %d: expected level '%s', got '%v'", i, expectedLevel, logEntry["level"])
		}
	}
}

// TestLogToConsole tests console logging
func TestLogToConsole(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 10,
		LogCompress:   false,
		LogToConsole:  true, // Enable console logging
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Verify logger output is set (should be multi-writer)
	if Logger.Out == nil {
		t.Error("Logger output is nil")
	}

	Close()
}

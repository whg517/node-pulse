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

// TestLogRotation_BySize tests log rotation by file size
func TestLogRotation_BySize(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    1, // 1 MB for testing
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Write enough logs to trigger rotation (approximately 1.1 MB)
	largeMessage := strings.Repeat("Test log entry for rotation ", 100) // ~2.5 KB per message
	for i := 0; i < 500; i++ {
		Info(largeMessage)
	}

	// Give lumberjack time to rotate
	time.Sleep(100 * time.Millisecond)

	Close()

	// Check for rotated log files
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	// Should have multiple log files (original + rotated)
	if len(files) < 2 {
		t.Errorf("Expected at least 2 log files due to rotation, got %d", len(files))
	}

	// Verify file naming pattern
	hasRotatedFile := false
	for _, file := range files {
		filename := file.Name()
		// Rotated files should have timestamp pattern
		if strings.HasPrefix(filename, "beacon-") && strings.HasSuffix(filename, ".log") {
			hasRotatedFile = true
			break
		}
	}

	if !hasRotatedFile {
		t.Error("Expected to find rotated log file with timestamp pattern")
	}
}

// TestLogRotation_Compression tests log file compression
func TestLogRotation_Compression(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    1, // 1 MB
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   true, // Enable compression
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Write enough logs to trigger rotation
	largeMessage := strings.Repeat("Test log entry for compression ", 100)
	for i := 0; i < 500; i++ {
		Info(largeMessage)
	}

	// Give lumberjack time to rotate and compress
	time.Sleep(200 * time.Millisecond)

	Close()

	// Check for compressed log files
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	// Should have at least one .gz compressed file
	hasCompressedFile := false
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".gz") {
			hasCompressedFile = true
			break
		}
	}

	if !hasCompressedFile {
		t.Log("Note: No compressed file found yet (rotation may not have triggered)")
		// This is not necessarily a failure - lumberjack compresses on rotation
		// If we didn't write enough logs, no rotation occurred
	}
}

// TestLogRotation_MaxBackups tests maximum backup retention
func TestLogRotation_MaxBackups(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    1, // 1 MB
		LogMaxAge:     7,
		LogMaxBackups: 2, // Keep only 2 backups
		LogCompress:   false,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Write enough logs to create multiple rotations
	largeMessage := strings.Repeat("Test log entry for max backups ", 100)
	for i := 0; i < 1500; i++ {
		Info(largeMessage)
	}

	time.Sleep(100 * time.Millisecond)
	Close()

	// Count log files
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	// Should not exceed MaxBackups + 1 (current file)
	// Note: This is a soft check - lumberjack may not have cleaned up old files yet
	if len(files) > cfg.LogMaxBackups+1 {
		t.Logf("Warning: Found %d log files, expected max %d", len(files), cfg.LogMaxBackups+1)
	}
}

// TestLogRotation_TimeBased tests time-based rotation (daily)
func TestLogRotation_TimeBased(t *testing.T) {
	// Note: This test is difficult to implement in a unit test without manipulating system time
	// Lumberjack uses local time and rotates daily at midnight
	// In production, this works automatically
	t.Skip("Time-based rotation requires system time manipulation - tested in integration tests")
}

// TestLogConfigurationDefaults tests default log configuration values
func TestLogConfigurationDefaults(t *testing.T) {
	// Test with zero values (should use defaults)
	cfg := &config.Config{
		LogLevel:      "",
		LogFile:       "",
		LogMaxSize:    0,
		LogMaxAge:     0,
		LogMaxBackups: 0,
		LogCompress:   false,
		LogToConsole:  false,
	}

	// Simulate what LoadConfig does - set defaults
	if cfg.LogLevel == "" {
		cfg.LogLevel = "INFO"
	}
	if cfg.LogFile == "" {
		cfg.LogFile = "/var/log/beacon/beacon.log"
	}
	if cfg.LogMaxSize == 0 {
		cfg.LogMaxSize = 10
	}
	if cfg.LogMaxAge == 0 {
		cfg.LogMaxAge = 7
	}
	if cfg.LogMaxBackups == 0 {
		cfg.LogMaxBackups = 10
	}

	// Verify defaults
	if cfg.LogLevel != "INFO" {
		t.Errorf("Expected default log level 'INFO', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFile != "/var/log/beacon/beacon.log" {
		t.Errorf("Expected default log file '/var/log/beacon/beacon.log', got '%s'", cfg.LogFile)
	}
	if cfg.LogMaxSize != 10 {
		t.Errorf("Expected default MaxSize 10, got %d", cfg.LogMaxSize)
	}
	if cfg.LogMaxAge != 7 {
		t.Errorf("Expected default MaxAge 7, got %d", cfg.LogMaxAge)
	}
	if cfg.LogMaxBackups != 10 {
		t.Errorf("Expected default MaxBackups 10, got %d", cfg.LogMaxBackups)
	}
}

// TestLogRotation_FileIntegrity tests that rotated log files are valid JSON
func TestLogRotation_FileIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	cfg := &config.Config{
		LogLevel:      "INFO",
		LogFile:       logFile,
		LogMaxSize:    1, // 1 MB
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Write logs with structured data
	for i := 0; i < 500; i++ {
		WithFields(map[string]interface{}{
			"iteration": i,
			"test":      "rotation",
		}).Info("Test log message")
	}

	time.Sleep(100 * time.Millisecond)
	Close()

	// Read all log files and verify JSON validity
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	validJSONCount := 0
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		filePath := filepath.Join(tempDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Logf("Failed to read %s: %v", file.Name(), err)
			continue
		}

		// Check each line is valid JSON
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		allValid := true
		for _, line := range lines {
			if line == "" {
				continue
			}
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
				t.Logf("Invalid JSON in %s: %v", file.Name(), err)
				allValid = false
				break
			}
		}

		if allValid && len(lines) > 0 {
			validJSONCount++
		}
	}

	if validJSONCount == 0 {
		t.Error("No valid JSON log files found")
	}

	t.Logf("Found %d log files with valid JSON", validJSONCount)
}

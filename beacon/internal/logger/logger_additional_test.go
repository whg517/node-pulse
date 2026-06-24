package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/whg517/node-pulse/beacon/internal/config"
)

// initAdditionalLogger creates a test logger
func initAdditionalLogger(t *testing.T) {
	t.Helper()
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "beacon.log")

	if err := InitLogger(&config.Config{
		LogLevel:      "DEBUG",
		LogFile:       logFile,
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	}); err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
}

// TestLogger_Infof tests Infof function
func TestLogger_Infof(t *testing.T) {
	initAdditionalLogger(t)
	defer func() { _ = Close() }()

	// Should not panic
	Infof("test info %s %d", "message", 42)
}

// TestLogger_Warnf tests Warnf function
func TestLogger_Warnf(t *testing.T) {
	initAdditionalLogger(t)
	defer func() { _ = Close() }()

	// Should not panic
	Warnf("test warn %s", "message")
}

// TestLogger_Errorf tests Errorf function
func TestLogger_Errorf(t *testing.T) {
	initAdditionalLogger(t)
	defer func() { _ = Close() }()

	// Should not panic
	Errorf("test error %s %v", "message", 42)
}

// TestLogger_GetLogger tests GetLogger function
func TestLogger_GetLogger(t *testing.T) {
	initAdditionalLogger(t)
	defer func() { _ = Close() }()

	l := GetLogger()
	if l == nil {
		t.Fatal("Expected non-nil logger")
	}
}

// TestLogger_GetLogger_SetDirectly tests GetLogger after direct set
func TestLogger_GetLogger_SetDirectly(t *testing.T) {
	// Set logger directly
	testLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	Logger = testLogger

	l := GetLogger()
	if l == nil {
		t.Fatal("Expected non-nil logger")
	}
	if l != testLogger {
		t.Error("Expected same logger instance")
	}
}

// TestLogger_Close tests Close
func TestLogger_Close(t *testing.T) {
	initAdditionalLogger(t)

	// Close should not error
	if err := Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// TestLogger_Close_NoFile tests Close when no file writer is set
func TestLogger_Close_NoFile(t *testing.T) {
	// Reset closer to nil
	closer = nil
	Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Close should not error
	if err := Close(); err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}

package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"

	"beacon/internal/config"
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
	// Set logger directly (as done in some tests)
	testLogger := logrus.New()
	testLogger.SetOutput(os.Stderr)
	Logger = testLogger

	l := GetLogger()
	if l == nil {
		t.Fatal("Expected non-nil logger")
	}
	if l != testLogger {
		t.Error("Expected same logger instance")
	}
}

// TestLogger_Close_MultiWriter tests Close with multi-writer
func TestLogger_Close_MultiWriter(t *testing.T) {
	initAdditionalLogger(t)

	// Close should not error
	if err := Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// TestLogger_Close_NonCloser tests Close when output is not a Closer
func TestLogger_Close_NonCloser(t *testing.T) {
	Logger = logrus.New()
	Logger.SetOutput(os.Stderr) // os.Stderr is not an io.Closer for this

	// Close should not error (returns nil when not a Closer)
	if err := Close(); err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}

package monitor

import (
	"log/slog"
	"os"
	"testing"

	"beacon/internal/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	logger.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	os.Exit(m.Run())
}

// TestSlogLogger_Info tests the Info method
func TestSlogLogger_Info(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Info("test info message")
	l.Info("multiple", "args", 123)
}

// TestSlogLogger_Infof tests the Infof method
func TestSlogLogger_Infof(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Infof("test info %s %d", "message", 42)
}

// TestSlogLogger_Warn tests the Warn method
func TestSlogLogger_Warn(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Warn("test warning message")
}

// TestSlogLogger_Warnf tests the Warnf method
func TestSlogLogger_Warnf(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Warnf("test warning %s", "format")
}

// TestSlogLogger_Error tests the Error method
func TestSlogLogger_Error(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Error("test error message")
}

// TestSlogLogger_Errorf tests the Errorf method
func TestSlogLogger_Errorf(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Errorf("test error %s %v", "format", 42)
}

// TestSlogLogger_Debug tests the Debug method
func TestSlogLogger_Debug(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Debug("test debug message")
}

// TestSlogLogger_Debugf tests the Debugf method
func TestSlogLogger_Debugf(t *testing.T) {
	l := NewSlogLogger(slog.Default())
	// Should not panic
	l.Debugf("test debug %s", "format")
}

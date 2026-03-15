package monitor

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"beacon/internal/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	logger.Logger = logrus.New()
	logger.Logger.SetOutput(os.Stderr)
	logger.Logger.SetLevel(logrus.WarnLevel)
	os.Exit(m.Run())
}

// TestLogrusLogger_Info tests the Info method
func TestLogrusLogger_Info(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic
	logger.Info("test info message")
	logger.Info("multiple", "args", 123)
}

// TestLogrusLogger_Infof tests the Infof method
func TestLogrusLogger_Infof(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic
	logger.Infof("test info %s %d", "message", 42)
}

// TestLogrusLogger_Warn tests the Warn method
func TestLogrusLogger_Warn(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic
	logger.Warn("test warning message")
}

// TestLogrusLogger_Warnf tests the Warnf method
func TestLogrusLogger_Warnf(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic
	logger.Warnf("test warning %s", "format")
}

// TestLogrusLogger_Error tests the Error method
func TestLogrusLogger_Error(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic
	logger.Error("test error message")
}

// TestLogrusLogger_Errorf tests the Errorf method
func TestLogrusLogger_Errorf(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic
	logger.Errorf("test error %s %v", "format", 42)
}

// TestLogrusLogger_Debug tests the Debug method
func TestLogrusLogger_Debug(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic (uses Info as fallback)
	logger.Debug("test debug message")
}

// TestLogrusLogger_Debugf tests the Debugf method
func TestLogrusLogger_Debugf(t *testing.T) {
	logger := &LogrusLogger{}
	// Should not panic (uses Infof as fallback)
	logger.Debugf("test debug %s", "format")
}

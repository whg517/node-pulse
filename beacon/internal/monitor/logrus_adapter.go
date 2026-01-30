package monitor

import (
	"beacon/internal/logger"
)

// LogrusLogger adapts the logrus logger to the monitor.Logger interface
type LogrusLogger struct{}

// Info logs at info level
func (l *LogrusLogger) Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof logs formatted info message
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warn logs at warning level
func (l *LogrusLogger) Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnf logs formatted warning message
func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Error logs at error level
func (l *LogrusLogger) Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf logs formatted error message
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Debug logs at debug level (using Info as fallback)
func (l *LogrusLogger) Debug(args ...interface{}) {
	logger.Info(args...)
}

// Debugf logs formatted debug message (using Infof as fallback)
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

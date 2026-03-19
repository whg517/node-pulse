package monitor

import (
	"fmt"
	"log/slog"
)

// SlogLogger adapts the global slog logger to the monitor.Logger interface.
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a SlogLogger wrapping the provided slog.Logger.
func NewSlogLogger(l *slog.Logger) *SlogLogger {
	return &SlogLogger{logger: l}
}

func (l *SlogLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *SlogLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *SlogLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *SlogLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *SlogLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *SlogLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *SlogLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *SlogLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

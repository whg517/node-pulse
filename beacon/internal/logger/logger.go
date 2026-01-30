package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"beacon/internal/config"
)

var (
	// Logger is the global logger instance
	Logger *logrus.Logger
)

// InitLogger initializes the global logger with configuration
func InitLogger(cfg *config.Config) error {
	Logger = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %w", cfg.LogLevel, err)
	}
	Logger.SetLevel(level)

	// Set JSON formatter
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Create log directory if not exists
	logDir := filepath.Dir(cfg.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Configure lumberjack for log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.LogFile,        // Log file path
		MaxSize:    cfg.LogMaxSize,     // Max size in MB
		MaxAge:     cfg.LogMaxAge,      // Max age in days
		MaxBackups: cfg.LogMaxBackups,  // Max number of old log files
		Compress:   cfg.LogCompress,    // Compress rotated files
		LocalTime:  true,               // Use local time for file names
	}

	// Multi-writer: file + console (if enabled)
	var writers []io.Writer
	writers = append(writers, lumberjackLogger)

	if cfg.LogToConsole {
		writers = append(writers, os.Stdout)
	}

	// Set output to multi-writer
	Logger.SetOutput(io.MultiWriter(writers...))

	return nil
}

// WithFields creates a logger entry with structured fields
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithField creates a logger entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

// WithError creates a logger entry with an error field
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

// Info logs an info message
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// Close flushes any buffered log entries
func Close() error {
	// lumberjack.Logger implements io.WriteCloser
	if closer, ok := Logger.Out.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

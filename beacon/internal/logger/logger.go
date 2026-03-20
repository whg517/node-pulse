package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"

	"beacon/internal/config"
)

var (
	// Logger is the global structured logger instance.
	Logger *slog.Logger

	// currentLevel tracks the configured log level for inspection.
	currentLevel slog.Level

	// closer holds the log file writer for cleanup.
	closer io.Closer
)

// InitLogger initializes the global logger with configuration.
// It uses log/slog with a JSON handler, supporting log rotation via lumberjack.
func InitLogger(cfg *config.Config) error {
	// Parse log level
	level, err := parseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %w", cfg.LogLevel, err)
	}
	currentLevel = level

	// Create log directory if it does not exist
	logDir := filepath.Dir(cfg.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Configure lumberjack for automatic log rotation
	lj := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    cfg.LogMaxSize,
		MaxAge:     cfg.LogMaxAge,
		MaxBackups: cfg.LogMaxBackups,
		Compress:   cfg.LogCompress,
		LocalTime:  true,
	}
	closer = lj

	// Build output writer (file + optional console)
	var writers []io.Writer
	writers = append(writers, lj)
	if cfg.LogToConsole {
		writers = append(writers, os.Stdout)
	}
	output := io.MultiWriter(writers...)

	// Create JSON handler with custom attribute names to match existing log schema:
	//   "time" -> "timestamp", "msg" -> "message"
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "timestamp"
			case slog.MessageKey:
				a.Key = "message"
			}
			return a
		},
	}

	Logger = slog.New(slog.NewJSONHandler(output, opts))
	// Set as global default so slog.Info() etc. use this logger
	slog.SetDefault(Logger)

	return nil
}

// parseLevel converts a level string to slog.Level.
func parseLevel(levelStr string) (slog.Level, error) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown level %q", levelStr)
	}
}

// GetLevel returns the configured log level.
func GetLevel() slog.Level {
	return currentLevel
}

// WithFields returns a child logger that includes the provided key-value pairs.
func WithFields(fields map[string]interface{}) *slog.Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return Logger.With(args...)
}

// WithField returns a child logger with a single additional field.
func WithField(key string, value interface{}) *slog.Logger {
	return Logger.With(key, value)
}

// WithError returns a child logger with an "error" field.
func WithError(err error) *slog.Logger {
	return Logger.With("error", err)
}

// Info logs a message at INFO level.
func Info(args ...interface{}) {
	Logger.Info(fmt.Sprint(args...))
}

// Infof logs a formatted message at INFO level.
func Infof(format string, args ...interface{}) {
	Logger.Info(fmt.Sprintf(format, args...))
}

// Warn logs a message at WARN level.
func Warn(args ...interface{}) {
	Logger.Warn(fmt.Sprint(args...))
}

// Warnf logs a formatted message at WARN level.
func Warnf(format string, args ...interface{}) {
	Logger.Warn(fmt.Sprintf(format, args...))
}

// Error logs a message at ERROR level.
func Error(args ...interface{}) {
	Logger.Error(fmt.Sprint(args...))
}

// Errorf logs a formatted message at ERROR level.
func Errorf(format string, args ...interface{}) {
	Logger.Error(fmt.Sprintf(format, args...))
}

// Fatal logs a message at ERROR level, then exits with status 1.
func Fatal(args ...interface{}) {
	Logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf logs a formatted message at ERROR level, then exits with status 1.
func Fatalf(format string, args ...interface{}) {
	Logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Close flushes any buffered log entries and releases file resources.
func Close() error {
	if closer != nil {
		return closer.Close()
	}
	return nil
}

// GetLogger returns the global slog.Logger instance.
func GetLogger() *slog.Logger {
	return Logger
}

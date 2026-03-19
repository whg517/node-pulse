// Package logger provides a structured logging setup for the Pulse server.
//
// It wraps the standard log/slog package and configures a global default logger
// from application configuration. All packages should use slog.Info(), slog.Warn(),
// etc. directly after calling InitLogger in main.
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/whg517/node-pulse/pulse/internal/config"
)

// InitLogger configures the global slog default logger from the application config.
//
// Supported log levels (case-insensitive): DEBUG, INFO, WARN/WARNING, ERROR.
// Supported formats: "json" (default for production), "text" (human-readable).
// Output is always written to stdout.
func InitLogger(cfg *config.LogConfig) error {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// Map standard slog keys to the project's log schema
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

	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// Default to JSON for structured production logging
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
	return nil
}

// parseLevel converts a level string to slog.Level.
func parseLevel(levelStr string) (slog.Level, error) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO", "":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q: expected DEBUG, INFO, WARN, or ERROR", levelStr)
	}
}

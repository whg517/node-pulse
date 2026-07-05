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
	"sync"

	"github.com/whg517/node-pulse/pulse/internal/config"
)

// handlerOpts is the shared options backing the default logger. It is kept at
// package level so SetLevel can mutate it in place for hot-reload (O-G4),
// avoiding the cost of rebuilding the handler on every reload.
var (
	handlerOpts   *slog.HandlerOptions
	handlerOptsMu sync.RWMutex
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

	handlerOptsMu.Lock()
	defer handlerOptsMu.Unlock()
	handlerOpts = &slog.HandlerOptions{
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
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	} else {
		// Default to JSON for structured production logging
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	}

	slog.SetDefault(slog.New(handler))
	return nil
}

// SetLevel hot-reloads the global logger's level (O-G4). It mutates the
// backing handler options in place, so all subsequent log calls observe the
// new level without rebuilding the handler. Called from the SIGHUP reload
// path in server.WaitForShutdown.
func SetLevel(levelStr string) error {
	level, err := parseLevel(levelStr)
	if err != nil {
		return err
	}
	handlerOptsMu.Lock()
	defer handlerOptsMu.Unlock()
	if handlerOpts == nil {
		return fmt.Errorf("logger not initialized")
	}
	handlerOpts.Level = level
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

package server

import (
	"github.com/kevin/node-pulse/pulse-api/internal/config"
)

// Config holds server configuration
// This is an adapter that delegates to the global config
type Config struct {
	*config.Config
}

// DefaultConfig returns default server configuration by loading global config
func DefaultConfig() *Config {
	cfg := config.MustLoad()
	return &Config{Config: cfg}
}

// GetPort returns the server port
func (c *Config) GetPort() string {
	return c.Server.Port
}

// GetDatabaseURL returns the database URL
func (c *Config) GetDatabaseURL() string {
	return c.DB.URL
}

// GetReadTimeout returns the read timeout in seconds
func (c *Config) GetReadTimeout() int {
	return c.Server.ReadTimeout
}

// GetWriteTimeout returns the write timeout in seconds
func (c *Config) GetWriteTimeout() int {
	return c.Server.WriteTimeout
}

// GetIdleTimeout returns the idle timeout in seconds
func (c *Config) GetIdleTimeout() int {
	return c.Server.IdleTimeout
}

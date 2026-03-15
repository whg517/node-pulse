package server

import (
	"github.com/whg517/node-pulse/pulse/internal/config"
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

// GetCORSConfig returns the CORS configuration
func (c *Config) GetCORSConfig() *config.CORSConfig {
	return &c.CORS
}

// GetAdminConfig returns the admin configuration
func (c *Config) GetAdminConfig() *config.AdminConfig {
	return &c.Admin
}

// GetSessionConfig returns the session configuration
func (c *Config) GetSessionConfig() *config.SessionConfig {
	return &c.Session
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Config.IsProduction()
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Config.IsDevelopment()
}

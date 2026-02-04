package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config holds all application configuration
type Config struct {
	Server  ServerConfig
	DB      DatabaseConfig
	Cleanup CleanupConfig
	Log     LogConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  int // seconds
	WriteTimeout int // seconds
	IdleTimeout  int // seconds
	Mode         string // debug, release, test
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string
	MaxConnections  int
	MinConnections  int
	ConnMaxLifetime int // seconds
	ConnMaxIdleTime int // seconds
}

// CleanupConfig holds cleanup task configuration
type CleanupConfig struct {
	Enabled         bool
	IntervalSeconds int
	RetentionDays   int
	SlowThresholdMs int64
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, file path
}

var (
	globalConfig *Config
	once         sync.Once
	initError    error
)

// Load loads configuration from multiple sources with priority
// Priority: Environment Variables > Config File > Default Values
func Load() (*Config, error) {
	once.Do(func() {
		globalConfig, initError = loadConfig()
	})
	return globalConfig, initError
}

// MustLoad loads configuration or panics
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}
	return cfg
}

// Get returns the global configuration instance
// Should be called after Load() or MustLoad()
func Get() *Config {
	if globalConfig == nil {
		panic("Configuration not loaded. Call Load() or MustLoad() first")
	}
	return globalConfig
}

// loadConfig implements the configuration loading logic
func loadConfig() (*Config, error) {
	// Start with defaults
	cfg := defaultConfig()

	// Load from config file if exists (not implemented yet, can be added later)
	// cfg = mergeFromFile(cfg, "config.yaml")

	// Override with environment variables
	cfg = mergeFromEnv(cfg)

	// Validate final configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// defaultConfig returns configuration with default values
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8080",
			ReadTimeout:  15,
			WriteTimeout: 15,
			IdleTimeout:  60,
			Mode:         "debug",
		},
		DB: DatabaseConfig{
			MaxConnections:  10,
			MinConnections:  1,
			ConnMaxLifetime: 3600, // 1 hour
			ConnMaxIdleTime: 300,  // 5 minutes
		},
		Cleanup: CleanupConfig{
			Enabled:         true,
			IntervalSeconds: 3600, // 1 hour
			RetentionDays:   7,
			SlowThresholdMs: 30000, // 30 seconds
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}
}

// mergeFromEnv overrides configuration with environment variables
func mergeFromEnv(cfg *Config) *Config {
	// Server configuration
	if v := os.Getenv("PULSE_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("PULSE_READ_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.ReadTimeout); i > 0 {
			cfg.Server.ReadTimeout = i
		}
	}
	if v := os.Getenv("PULSE_WRITE_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.WriteTimeout); i > 0 {
			cfg.Server.WriteTimeout = i
		}
	}
	if v := os.Getenv("PULSE_IDLE_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.IdleTimeout); i > 0 {
			cfg.Server.IdleTimeout = i
		}
	}
	if v := os.Getenv("PULSE_MODE"); v != "" {
		cfg.Server.Mode = v
	}

	// Database configuration
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.DB.URL = v
	}
	if v := os.Getenv("DB_MAX_CONNECTIONS"); v != "" {
		if i := parseInt(v, cfg.DB.MaxConnections); i > 0 {
			cfg.DB.MaxConnections = i
		}
	}
	if v := os.Getenv("DB_MIN_CONNECTIONS"); v != "" {
		if i := parseInt(v, cfg.DB.MinConnections); i >= 0 {
			cfg.DB.MinConnections = i
		}
	}
	if v := os.Getenv("DB_CONN_MAX_LIFETIME"); v != "" {
		if i := parseInt(v, cfg.DB.ConnMaxLifetime); i > 0 {
			cfg.DB.ConnMaxLifetime = i
		}
	}
	if v := os.Getenv("DB_CONN_MAX_IDLE_TIME"); v != "" {
		if i := parseInt(v, cfg.DB.ConnMaxIdleTime); i >= 0 {
			cfg.DB.ConnMaxIdleTime = i
		}
	}

	// Cleanup configuration
	if v := os.Getenv("CLEANUP_ENABLED"); v != "" {
		cfg.Cleanup.Enabled = parseBool(v, cfg.Cleanup.Enabled)
	}
	if v := os.Getenv("CLEANUP_INTERVAL"); v != "" {
		if i := parseInt(v, cfg.Cleanup.IntervalSeconds); i > 0 {
			cfg.Cleanup.IntervalSeconds = i
		}
	}
	if v := os.Getenv("CLEANUP_RETENTION_DAYS"); v != "" {
		if i := parseInt(v, cfg.Cleanup.RetentionDays); i > 0 {
			cfg.Cleanup.RetentionDays = i
		}
	}
	if v := os.Getenv("CLEANUP_SLOW_THRESHOLD"); v != "" {
		if i := parseInt64(v, cfg.Cleanup.SlowThresholdMs); i >= 0 {
			cfg.Cleanup.SlowThresholdMs = i
		}
	}

	// Log configuration
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		cfg.Log.Output = v
	}

	return cfg
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config invalid: %w", err)
	}
	if err := c.DB.Validate(); err != nil {
		return fmt.Errorf("database config invalid: %w", err)
	}
	if err := c.Cleanup.Validate(); err != nil {
		return fmt.Errorf("cleanup config invalid: %w", err)
	}
	return nil
}

// Validate validates server configuration
func (c *ServerConfig) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive, got %d", c.ReadTimeout)
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive, got %d", c.WriteTimeout)
	}
	if c.IdleTimeout <= 0 {
		return fmt.Errorf("idle_timeout must be positive, got %d", c.IdleTimeout)
	}
	if c.Mode != "" && c.Mode != "debug" && c.Mode != "release" && c.Mode != "test" {
		return fmt.Errorf("mode must be one of: debug, release, test, got %s", c.Mode)
	}
	return nil
}

// Validate validates database configuration
func (c *DatabaseConfig) Validate() error {
	if c.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive, got %d", c.MaxConnections)
	}
	if c.MinConnections < 0 {
		return fmt.Errorf("min_connections cannot be negative, got %d", c.MinConnections)
	}
	if c.MinConnections > c.MaxConnections {
		return fmt.Errorf("min_connections (%d) cannot be greater than max_connections (%d)",
			c.MinConnections, c.MaxConnections)
	}
	if c.ConnMaxLifetime <= 0 {
		return fmt.Errorf("conn_max_lifetime must be positive, got %d", c.ConnMaxLifetime)
	}
	if c.ConnMaxIdleTime < 0 {
		return fmt.Errorf("conn_max_idle_time cannot be negative, got %d", c.ConnMaxIdleTime)
	}
	return nil
}

// Validate validates cleanup configuration
func (c *CleanupConfig) Validate() error {
	if c.IntervalSeconds <= 0 {
		return fmt.Errorf("interval_seconds must be positive, got %d", c.IntervalSeconds)
	}
	if c.RetentionDays <= 0 {
		return fmt.Errorf("retention_days must be positive, got %d", c.RetentionDays)
	}
	if c.SlowThresholdMs < 0 {
		return fmt.Errorf("slow_threshold_ms cannot be negative, got %d", c.SlowThresholdMs)
	}
	return nil
}

// Helper functions

func parseInt(s string, defaultValue int) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}

func parseInt64(s string, defaultValue int64) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

func parseBool(s string, defaultValue bool) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return defaultValue
	}
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

// Reset resets the global configuration (mainly for testing)
func Reset() {
	globalConfig = nil
	initError = nil
	once = sync.Once{}
}

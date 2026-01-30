package config

import (
	"fmt"
	"os"
	"strconv"
)

// CleanupConfig defines the configuration for cleanup task
type CleanupConfig struct {
	Enabled         bool  `yaml:"enabled" env:"CLEANUP_ENABLED" default:"true"`
	IntervalSeconds int   `yaml:"interval_seconds" env:"CLEANUP_INTERVAL" default:"3600"`
	RetentionDays   int   `yaml:"retention_days" env:"CLEANUP_RETENTION_DAYS" default:"7"`
	SlowThresholdMs int64 `yaml:"slow_threshold_ms" env:"CLEANUP_SLOW_THRESHOLD" default:"30000"`
}

// LoadCleanupConfig loads cleanup configuration from environment variables
func LoadCleanupConfig() (*CleanupConfig, error) {
	cfg := &CleanupConfig{
		Enabled:         getEnvBool("CLEANUP_ENABLED", true),
		IntervalSeconds: getEnvInt("CLEANUP_INTERVAL", 3600),
		RetentionDays:   getEnvInt("CLEANUP_RETENTION_DAYS", 7),
		SlowThresholdMs: int64(getEnvInt("CLEANUP_SLOW_THRESHOLD", 30000)),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cleanup config: %w", err)
	}

	return cfg, nil
}

// Validate validates the cleanup configuration
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

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1" || val == "yes" || val == "on"
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

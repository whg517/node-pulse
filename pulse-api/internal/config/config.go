package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server  ServerConfig
	DB      DatabaseConfig
	Cleanup CleanupConfig
	Log     LogConfig
	CORS    CORSConfig
	Admin   AdminConfig
	Session SessionConfig
	JWT     JWTConfig
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

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
	MaxAge         int // seconds
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	Username string
	Password string
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Secret          string
	ExpirationHours int
	CookieSecure    bool
	CookieSameSite  string // Strict, Lax, None
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
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

	// Load from config file if exists
	fileCfg, err := loadFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}
	if fileCfg != nil {
		// Merge file config over defaults (non-zero values only)
		mergeConfig(cfg, fileCfg)
	}

	// Override with environment variables
	cfg = mergeFromEnv(cfg)

	// Auto-generate secrets if needed
	if err := generateSecrets(cfg); err != nil {
		return nil, fmt.Errorf("failed to generate secrets: %w", err)
	}

	// Validate final configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadFromFile attempts to load configuration from YAML file
// Search order: CONFIG_PATH env var → ./config.yaml → /etc/node-pulse/config.yaml
// Returns nil if no file found (not an error)
func loadFromFile() (*Config, error) {
	var configPaths []string

	// Check CONFIG_PATH environment variable first
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		configPaths = append(configPaths, configPath)
	}

	// Check current working directory
	configPaths = append(configPaths, "./config.yaml")

	// Check system-wide config directory
	configPaths = append(configPaths, "/etc/node-pulse/config.yaml")

	// Try each path
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, try to load it
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
			}

			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("failed to parse YAML from %s: %w", path, err)
			}

			return &cfg, nil
		}
	}

	// No config file found - this is OK, will use env vars and defaults
	return nil, nil
}

// mergeConfig merges src config into dst config (overwrites non-zero values)
func mergeConfig(dst, src *Config) {
	if src.Server.Port != "" {
		dst.Server.Port = src.Server.Port
	}
	if src.Server.ReadTimeout != 0 {
		dst.Server.ReadTimeout = src.Server.ReadTimeout
	}
	if src.Server.WriteTimeout != 0 {
		dst.Server.WriteTimeout = src.Server.WriteTimeout
	}
	if src.Server.IdleTimeout != 0 {
		dst.Server.IdleTimeout = src.Server.IdleTimeout
	}
	if src.Server.Mode != "" {
		dst.Server.Mode = src.Server.Mode
	}

	if src.DB.URL != "" {
		dst.DB.URL = src.DB.URL
	}
	if src.DB.MaxConnections != 0 {
		dst.DB.MaxConnections = src.DB.MaxConnections
	}
	if src.DB.MinConnections != 0 {
		dst.DB.MinConnections = src.DB.MinConnections
	}
	if src.DB.ConnMaxLifetime != 0 {
		dst.DB.ConnMaxLifetime = src.DB.ConnMaxLifetime
	}
	if src.DB.ConnMaxIdleTime != 0 {
		dst.DB.ConnMaxIdleTime = src.DB.ConnMaxIdleTime
	}

	if src.Cleanup.Enabled {
		dst.Cleanup.Enabled = src.Cleanup.Enabled
	}
	if src.Cleanup.IntervalSeconds != 0 {
		dst.Cleanup.IntervalSeconds = src.Cleanup.IntervalSeconds
	}
	if src.Cleanup.RetentionDays != 0 {
		dst.Cleanup.RetentionDays = src.Cleanup.RetentionDays
	}
	if src.Cleanup.SlowThresholdMs != 0 {
		dst.Cleanup.SlowThresholdMs = src.Cleanup.SlowThresholdMs
	}

	if src.Log.Level != "" {
		dst.Log.Level = src.Log.Level
	}
	if src.Log.Format != "" {
		dst.Log.Format = src.Log.Format
	}
	if src.Log.Output != "" {
		dst.Log.Output = src.Log.Output
	}

	if src.CORS.AllowedOrigins != "" {
		dst.CORS.AllowedOrigins = src.CORS.AllowedOrigins
	}
	if src.CORS.AllowedMethods != "" {
		dst.CORS.AllowedMethods = src.CORS.AllowedMethods
	}
	if src.CORS.AllowedHeaders != "" {
		dst.CORS.AllowedHeaders = src.CORS.AllowedHeaders
	}
	if src.CORS.MaxAge != 0 {
		dst.CORS.MaxAge = src.CORS.MaxAge
	}

	if src.Admin.Username != "" {
		dst.Admin.Username = src.Admin.Username
	}
	if src.Admin.Password != "" {
		dst.Admin.Password = src.Admin.Password
	}

	if src.Session.Secret != "" {
		dst.Session.Secret = src.Session.Secret
	}
	if src.Session.ExpirationHours != 0 {
		dst.Session.ExpirationHours = src.Session.ExpirationHours
	}
	if src.Session.CookieSecure {
		dst.Session.CookieSecure = src.Session.CookieSecure
	}
	if src.Session.CookieSameSite != "" {
		dst.Session.CookieSameSite = src.Session.CookieSameSite
	}

	if src.JWT.Secret != "" {
		dst.JWT.Secret = src.JWT.Secret
	}
	if src.JWT.ExpirationHours != 0 {
		dst.JWT.ExpirationHours = src.JWT.ExpirationHours
	}
}

// generateSecrets auto-generates secrets if not provided
func generateSecrets(cfg *Config) error {
	// Generate session secret if not provided
	if cfg.Session.Secret == "" {
		secret, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("failed to generate session secret: %w", err)
		}
		cfg.Session.Secret = secret
	}

	// Generate JWT secret if not provided
	if cfg.JWT.Secret == "" {
		secret, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		cfg.JWT.Secret = secret
	}

	return nil
}

// generateRandomSecret generates a cryptographically secure random secret
func generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
		CORS: CORSConfig{
			AllowedOrigins: "http://localhost:3000,http://localhost:5173,http://localhost:8080",
			AllowedMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowedHeaders: "Content-Type,Authorization",
			MaxAge:         86400, // 24 hours
		},
		Admin: AdminConfig{
			Username: "admin",
			Password: "Admin123",
		},
		Session: SessionConfig{
			Secret:          "", // Will be auto-generated if empty
			ExpirationHours: 24,
			CookieSecure:    false, // Derived from Mode in mergeFromEnv
			CookieSameSite:  "Lax",
		},
		JWT: JWTConfig{
			Secret:          "", // Will be auto-generated if empty
			ExpirationHours: 24,
		},
	}
}

// mergeFromEnv overrides configuration with environment variables
func mergeFromEnv(cfg *Config) *Config {
	// Server configuration
	if v := os.Getenv("PULSE_SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	} else if v := os.Getenv("PULSE_PORT"); v != "" {
		// Legacy fallback
		cfg.Server.Port = v
	}
	if v := os.Getenv("PULSE_SERVER_READ_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.ReadTimeout); i > 0 {
			cfg.Server.ReadTimeout = i
		}
	} else if v := os.Getenv("PULSE_READ_TIMEOUT"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.Server.ReadTimeout); i > 0 {
			cfg.Server.ReadTimeout = i
		}
	}
	if v := os.Getenv("PULSE_SERVER_WRITE_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.WriteTimeout); i > 0 {
			cfg.Server.WriteTimeout = i
		}
	} else if v := os.Getenv("PULSE_WRITE_TIMEOUT"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.Server.WriteTimeout); i > 0 {
			cfg.Server.WriteTimeout = i
		}
	}
	if v := os.Getenv("PULSE_SERVER_IDLE_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.IdleTimeout); i > 0 {
			cfg.Server.IdleTimeout = i
		}
	} else if v := os.Getenv("PULSE_IDLE_TIMEOUT"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.Server.IdleTimeout); i > 0 {
			cfg.Server.IdleTimeout = i
		}
	}
	if v := os.Getenv("PULSE_SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	} else if v := os.Getenv("PULSE_MODE"); v != "" {
		// Legacy fallback
		cfg.Server.Mode = v
	}

	// Database configuration
	if v := os.Getenv("PULSE_DATABASE_URL"); v != "" {
		cfg.DB.URL = v
	} else if v := os.Getenv("DATABASE_URL"); v != "" {
		// Legacy fallback
		cfg.DB.URL = v
	}
	if v := os.Getenv("PULSE_DATABASE_MAX_CONNECTIONS"); v != "" {
		if i := parseInt(v, cfg.DB.MaxConnections); i > 0 {
			cfg.DB.MaxConnections = i
		}
	} else if v := os.Getenv("DB_MAX_CONNECTIONS"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.DB.MaxConnections); i > 0 {
			cfg.DB.MaxConnections = i
		}
	}
	if v := os.Getenv("PULSE_DATABASE_MIN_CONNECTIONS"); v != "" {
		if i := parseInt(v, cfg.DB.MinConnections); i >= 0 {
			cfg.DB.MinConnections = i
		}
	} else if v := os.Getenv("DB_MIN_CONNECTIONS"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.DB.MinConnections); i >= 0 {
			cfg.DB.MinConnections = i
		}
	}
	if v := os.Getenv("PULSE_DATABASE_CONN_MAX_LIFETIME"); v != "" {
		if i := parseInt(v, cfg.DB.ConnMaxLifetime); i > 0 {
			cfg.DB.ConnMaxLifetime = i
		}
	} else if v := os.Getenv("DB_CONN_MAX_LIFETIME"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.DB.ConnMaxLifetime); i > 0 {
			cfg.DB.ConnMaxLifetime = i
		}
	}
	if v := os.Getenv("PULSE_DATABASE_CONN_MAX_IDLE_TIME"); v != "" {
		if i := parseInt(v, cfg.DB.ConnMaxIdleTime); i >= 0 {
			cfg.DB.ConnMaxIdleTime = i
		}
	} else if v := os.Getenv("DB_CONN_MAX_IDLE_TIME"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.DB.ConnMaxIdleTime); i >= 0 {
			cfg.DB.ConnMaxIdleTime = i
		}
	}

	// Cleanup configuration
	if v := os.Getenv("PULSE_CLEANUP_ENABLED"); v != "" {
		cfg.Cleanup.Enabled = parseBool(v, cfg.Cleanup.Enabled)
	} else if v := os.Getenv("CLEANUP_ENABLED"); v != "" {
		// Legacy fallback
		cfg.Cleanup.Enabled = parseBool(v, cfg.Cleanup.Enabled)
	}
	if v := os.Getenv("PULSE_CLEANUP_INTERVAL"); v != "" {
		if i := parseInt(v, cfg.Cleanup.IntervalSeconds); i > 0 {
			cfg.Cleanup.IntervalSeconds = i
		}
	} else if v := os.Getenv("CLEANUP_INTERVAL"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.Cleanup.IntervalSeconds); i > 0 {
			cfg.Cleanup.IntervalSeconds = i
		}
	}
	if v := os.Getenv("PULSE_CLEANUP_RETENTION_DAYS"); v != "" {
		if i := parseInt(v, cfg.Cleanup.RetentionDays); i > 0 {
			cfg.Cleanup.RetentionDays = i
		}
	} else if v := os.Getenv("CLEANUP_RETENTION_DAYS"); v != "" {
		// Legacy fallback
		if i := parseInt(v, cfg.Cleanup.RetentionDays); i > 0 {
			cfg.Cleanup.RetentionDays = i
		}
	}
	if v := os.Getenv("PULSE_CLEANUP_SLOW_THRESHOLD"); v != "" {
		if i := parseInt64(v, cfg.Cleanup.SlowThresholdMs); i >= 0 {
			cfg.Cleanup.SlowThresholdMs = i
		}
	} else if v := os.Getenv("CLEANUP_SLOW_THRESHOLD"); v != "" {
		// Legacy fallback
		if i := parseInt64(v, cfg.Cleanup.SlowThresholdMs); i >= 0 {
			cfg.Cleanup.SlowThresholdMs = i
		}
	}

	// Log configuration
	if v := os.Getenv("PULSE_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	} else if v := os.Getenv("LOG_LEVEL"); v != "" {
		// Legacy fallback
		cfg.Log.Level = v
	}
	if v := os.Getenv("PULSE_LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	} else if v := os.Getenv("LOG_FORMAT"); v != "" {
		// Legacy fallback
		cfg.Log.Format = v
	}
	if v := os.Getenv("PULSE_LOG_OUTPUT"); v != "" {
		cfg.Log.Output = v
	} else if v := os.Getenv("LOG_OUTPUT"); v != "" {
		// Legacy fallback
		cfg.Log.Output = v
	}

	// CORS configuration
	if v := os.Getenv("PULSE_CORS_ALLOWED_ORIGINS"); v != "" {
		cfg.CORS.AllowedOrigins = v
	} else if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		// Legacy fallback
		cfg.CORS.AllowedOrigins = v
	}
	if v := os.Getenv("PULSE_CORS_ALLOWED_METHODS"); v != "" {
		cfg.CORS.AllowedMethods = v
	}
	if v := os.Getenv("PULSE_CORS_ALLOWED_HEADERS"); v != "" {
		cfg.CORS.AllowedHeaders = v
	}
	if v := os.Getenv("PULSE_CORS_MAX_AGE"); v != "" {
		if i := parseInt(v, cfg.CORS.MaxAge); i > 0 {
			cfg.CORS.MaxAge = i
		}
	}

	// Admin configuration
	if v := os.Getenv("PULSE_ADMIN_USERNAME"); v != "" {
		cfg.Admin.Username = v
	} else if v := os.Getenv("ADMIN_USERNAME"); v != "" {
		// Legacy fallback
		cfg.Admin.Username = v
	}
	if v := os.Getenv("PULSE_ADMIN_PASSWORD"); v != "" {
		cfg.Admin.Password = v
	} else if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		// Legacy fallback
		cfg.Admin.Password = v
	}

	// Session configuration
	if v := os.Getenv("PULSE_SESSION_SECRET"); v != "" {
		cfg.Session.Secret = v
	} else if v := os.Getenv("SESSION_SECRET"); v != "" {
		// Legacy fallback
		cfg.Session.Secret = v
	}
	if v := os.Getenv("PULSE_SESSION_EXPIRATION_HOURS"); v != "" {
		if i := parseInt(v, cfg.Session.ExpirationHours); i > 0 {
			cfg.Session.ExpirationHours = i
		}
	}
	if v := os.Getenv("PULSE_SESSION_COOKIE_SECURE"); v != "" {
		cfg.Session.CookieSecure = parseBool(v, cfg.Session.CookieSecure)
	}
	if v := os.Getenv("PULSE_SESSION_COOKIE_SAMESITE"); v != "" {
		cfg.Session.CookieSameSite = v
	}

	// JWT configuration
	if v := os.Getenv("PULSE_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	} else if v := os.Getenv("JWT_SECRET"); v != "" {
		// Legacy fallback
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("PULSE_JWT_EXPIRATION_HOURS"); v != "" {
		if i := parseInt(v, cfg.JWT.ExpirationHours); i > 0 {
			cfg.JWT.ExpirationHours = i
		}
	}

	// Derive CookieSecure from Mode if not explicitly set
	if !cfg.Session.CookieSecure {
		cfg.Session.CookieSecure = cfg.IsProduction()
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
	if err := c.Log.Validate(); err != nil {
		return fmt.Errorf("log config invalid: %w", err)
	}
	if err := c.CORS.Validate(); err != nil {
		return fmt.Errorf("cors config invalid: %w", err)
	}
	if err := c.Admin.Validate(); err != nil {
		return fmt.Errorf("admin config invalid: %w", err)
	}
	if err := c.Session.Validate(); err != nil {
		return fmt.Errorf("session config invalid: %w", err)
	}
	if err := c.JWT.Validate(); err != nil {
		return fmt.Errorf("jwt config invalid: %w", err)
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

// Validate validates log configuration
func (c *LogConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if c.Level != "" && !validLevels[c.Level] {
		return fmt.Errorf("level must be one of: debug, info, warn, error, got %s", c.Level)
	}
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if c.Format != "" && !validFormats[c.Format] {
		return fmt.Errorf("format must be one of: json, text, got %s", c.Format)
	}
	return nil
}

// Validate validates CORS configuration
func (c *CORSConfig) Validate() error {
	if c.AllowedOrigins == "" {
		return fmt.Errorf("allowed_origins cannot be empty")
	}
	if c.AllowedMethods == "" {
		return fmt.Errorf("allowed_methods cannot be empty")
	}
	if c.MaxAge < 0 {
		return fmt.Errorf("max_age cannot be negative, got %d", c.MaxAge)
	}
	return nil
}

// Validate validates admin configuration
func (c *AdminConfig) Validate() error {
	if c.Username == "" {
		return fmt.Errorf("admin username cannot be empty")
	}
	if c.Password == "" {
		return fmt.Errorf("admin password cannot be empty")
	}
	return nil
}

// Validate validates session configuration
func (c *SessionConfig) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("session secret cannot be empty")
	}
	if c.ExpirationHours <= 0 {
		return fmt.Errorf("session expiration_hours must be positive, got %d", c.ExpirationHours)
	}
	validSameSite := map[string]bool{
		"Strict": true,
		"Lax":    true,
		"None":   true,
	}
	if c.CookieSameSite != "" && !validSameSite[c.CookieSameSite] {
		return fmt.Errorf("cookie_samesite must be one of: Strict, Lax, None, got %s", c.CookieSameSite)
	}
	return nil
}

// Validate validates JWT configuration
func (c *JWTConfig) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("jwt secret cannot be empty")
	}
	if c.ExpirationHours <= 0 {
		return fmt.Errorf("jwt expiration_hours must be positive, got %d", c.ExpirationHours)
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

// String returns a safe string representation of the configuration with credentials redacted
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Server{Port:%s,Mode:%s} DB{URL:%s,MaxConn:%d} "+
			"Cleanup{Enabled:%t} Log{Level:%s} CORS{Origins:%s} "+
			"Admin{Username:%s,Password:***REDACTED***} "+
			"Session{Secret:***REDACTED***,ExpirationHours:%d} "+
			"JWT{Secret:***REDACTED***,ExpirationHours:%d}}",
		c.Server.Port, c.Server.Mode,
		maskURL(c.DB.URL), c.DB.MaxConnections,
		c.Cleanup.Enabled, c.Log.Level,
		c.CORS.AllowedOrigins, c.Admin.Username,
		c.Session.ExpirationHours,
		c.JWT.ExpirationHours,
	)
}

// maskURL masks sensitive parts of a URL (e.g., password in postgres://user:pass@host)
func maskURL(url string) string {
	if url == "" {
		return ""
	}
	// If URL contains password, mask it
	// Format: postgres://user:password@host:port/db
	if strings.Contains(url, "://") && strings.Contains(url, "@") {
		parts := strings.Split(url, "://")
		if len(parts) == 2 {
			rest := parts[1]
			if strings.Contains(rest, "@") {
				credParts := strings.SplitN(rest, "@", 2)
				if len(credParts) == 2 && strings.Contains(credParts[0], ":") {
					userParts := strings.SplitN(credParts[0], ":", 2)
					maskedCreds := userParts[0] + ":***@"
					return parts[0] + "://" + maskedCreds + credParts[1]
				}
			}
		}
	}
	// For safety, if URL is long enough, truncate it
	if len(url) > 50 {
		return url[:47] + "..."
	}
	return url
}

// IsProduction returns true if the server is running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}

// IsDevelopment returns true if the server is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "debug"
}

package config

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	DB         DatabaseConfig   `yaml:"database"`
	Cleanup    CleanupConfig    `yaml:"cleanup"`
	Log        LogConfig        `yaml:"log"`
	CORS       CORSConfig       `yaml:"cors"`
	Admin      AdminConfig      `yaml:"admin"`
	Session    SessionConfig    `yaml:"session"`
	JWT        JWTConfig        `yaml:"jwt"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
	Telemetry  TelemetryConfig  `yaml:"telemetry"`
}

// TelemetryConfig holds OpenTelemetry / distributed-tracing configuration.
type TelemetryConfig struct {
	// Enabled controls whether distributed tracing is active.
	Enabled bool `yaml:"enabled"`

	// ServiceName is the logical service name reported in traces (default: "pulse").
	ServiceName string `yaml:"service_name"`

	// ServiceVersion is the deployed version string (default: "unknown").
	ServiceVersion string `yaml:"service_version"`

	// Environment is the deployment environment, e.g. "production", "staging", "development".
	Environment string `yaml:"environment"`

	// OTLPEndpoint is the gRPC address of an OTLP-compatible collector,
	// e.g. "localhost:4317".  When empty, traces are written to stdout.
	OTLPEndpoint string `yaml:"otlp_endpoint"`

	// SamplingRate controls what fraction of requests are traced (0.0 – 1.0, default 1.0).
	SamplingRate float64 `yaml:"sampling_rate"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
	Mode         string `yaml:"mode"`
	// BaseURL is the externally reachable URL of this Pulse instance (no trailing
	// slash), used to render absolute links in webhook payloads and exported data.
	// Defaults to http://localhost:<port> for local development.
	BaseURL string `yaml:"base_url"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string `yaml:"url"`
	MaxConnections  int    `yaml:"max_connections"`
	MinConnections  int    `yaml:"min_connections"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time"`
}

// CleanupConfig holds cleanup task configuration
type CleanupConfig struct {
	Enabled         bool  `yaml:"enabled"`
	IntervalSeconds int   `yaml:"interval_seconds"`
	RetentionDays   int   `yaml:"retention_days"`
	SlowThresholdMs int64 `yaml:"slow_threshold_ms"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string `yaml:"allowed_origins"`
	AllowedMethods string `yaml:"allowed_methods"`
	AllowedHeaders string `yaml:"allowed_headers"`
	MaxAge         int    `yaml:"max_age"`
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
	CookieSecure    bool   `yaml:"cookie_secure"`
	CookieSameSite  string `yaml:"cookie_samesite"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret                       string `yaml:"secret"`
	PrivateKey                   string `yaml:"private_key"`
	PublicKey                    string `yaml:"public_key"`
	AccessTokenExpirationMinutes int    `yaml:"access_token_expiration_minutes"`
	RefreshTokenExpirationDays   int    `yaml:"refresh_token_expiration_days"`
	RefreshTokenMaxValidityDays  int    `yaml:"refresh_token_max_validity_days"`
	KeyID                        string `yaml:"key_id"`
}

// RateLimitConfig holds rate limiting configuration for auth endpoints
type RateLimitConfig struct {
	LoginMaxPerMinute   int `yaml:"login_max_per_minute"`
	LoginMaxPerDay      int `yaml:"login_max_per_day"`
	RefreshMaxPerMinute int `yaml:"refresh_max_per_minute"`
	RefreshMaxPerDay    int `yaml:"refresh_max_per_day"`
	APIKeyMaxPerMinute  int `yaml:"apikey_max_per_minute"`
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

	// Load from config file if exists (merged onto defaults)
	if err := loadFromFile(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
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

// loadFromFile attempts to load configuration from YAML file.
// Search order: PULSE_CONFIG_PATH env var → ./pulse.yaml → /etc/node-pulse/pulse.yaml
// Returns nil if no file found (not an error)
func loadFromFile(cfg *Config) error {
	var configPaths []string

	// Check PULSE_CONFIG_PATH environment variable first
	if configPath := os.Getenv("PULSE_CONFIG_PATH"); configPath != "" {
		configPaths = append(configPaths, configPath)
	}

	// Check current working directory
	configPaths = append(configPaths, "./pulse.yaml")

	// Check system-wide config directory
	configPaths = append(configPaths, "/etc/node-pulse/pulse.yaml")

	// Try each path
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, try to load it
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read config file %s: %w", path, err)
			}

			decoder := yaml.NewDecoder(bytes.NewReader(data))
			decoder.KnownFields(true)
			if err := decoder.Decode(cfg); err != nil {
				return fmt.Errorf("failed to parse YAML from %s: %w", path, err)
			}

			return nil
		}
	}

	// No config file found - this is OK, will use env vars and defaults
	return nil
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

	// Generate RSA key pair for RS256 JWT if not provided
	if cfg.JWT.PrivateKey == "" || cfg.JWT.PublicKey == "" {
		privateKeyPEM, publicKeyPEM, err := generateRSAKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate RSA key pair: %w", err)
		}
		cfg.JWT.PrivateKey = privateKeyPEM
		cfg.JWT.PublicKey = publicKeyPEM
	}

	// Generate KeyID if not provided
	if cfg.JWT.KeyID == "" {
		cfg.JWT.KeyID = fmt.Sprintf("key-%d", time.Now().Unix())
	}

	// Generate legacy JWT secret if not provided (for backward compatibility)
	if cfg.JWT.Secret == "" {
		secret, err := generateRandomSecret(64) // 64 bytes (512 bits) for NIST compliance
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

// generateRSAKeyPair generates an RSA-2048 key pair for JWT RS256 signing
// Returns private key and public key in PEM format
func generateRSAKeyPair() (string, string, error) {
	// Generate 2048-bit RSA private key (minimum per design spec)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM), nil
}

// defaultConfig returns configuration with default values
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "6532",
			ReadTimeout:  15,
			WriteTimeout: 15,
			IdleTimeout:  60,
			Mode:         "debug",
			BaseURL:      "http://localhost:6532", // overwritten by PULSE_SERVER_BASE_URL in production
		},
		DB: DatabaseConfig{
			MaxConnections:  25,
			MinConnections:  2,
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
			AllowedOrigins: "http://localhost:4173,http://localhost:5173,http://localhost:6532",
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
			Secret:                       "", // Will be auto-generated if empty
			AccessTokenExpirationMinutes: 15, // 15 minutes
			RefreshTokenExpirationDays:   7,  // 7 days
			RefreshTokenMaxValidityDays:  30, // 30 days absolute cap
		},
		RateLimit: RateLimitConfig{
			LoginMaxPerMinute:   5,   // 5 login attempts per minute per IP
			LoginMaxPerDay:      100, // 100 login attempts per day per IP
			RefreshMaxPerMinute: 10,  // 10 refresh attempts per minute per token
			RefreshMaxPerDay:    200, // 200 refresh attempts per day per token
			APIKeyMaxPerMinute:  11,  // 11 API key exchanges per minute per key
		},
		Telemetry: TelemetryConfig{
			Enabled:        false,         // opt-in; set PULSE_TELEMETRY_ENABLED=true to activate
			ServiceName:    "pulse",
			ServiceVersion: "unknown",
			Environment:    "development",
			OTLPEndpoint:   "",            // empty → stdout exporter (dev/debug)
			SamplingRate:   1.0,           // trace every request by default
		},
	}
}

// mergeFromEnv overrides configuration with environment variables
func mergeFromEnv(cfg *Config) *Config {
	// Server configuration
	if v := os.Getenv("PULSE_SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("PULSE_SERVER_READ_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.ReadTimeout); i > 0 {
			cfg.Server.ReadTimeout = i
		}
	}
	if v := os.Getenv("PULSE_SERVER_WRITE_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.WriteTimeout); i > 0 {
			cfg.Server.WriteTimeout = i
		}
	}
	if v := os.Getenv("PULSE_SERVER_IDLE_TIMEOUT"); v != "" {
		if i := parseInt(v, cfg.Server.IdleTimeout); i > 0 {
			cfg.Server.IdleTimeout = i
		}
	}
	if v := os.Getenv("PULSE_SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}
	if v := os.Getenv("PULSE_SERVER_BASE_URL"); v != "" {
		cfg.Server.BaseURL = strings.TrimRight(v, "/")
	}

	// Database configuration
	if v := os.Getenv("PULSE_DATABASE_URL"); v != "" {
		cfg.DB.URL = v
	}
	if v := os.Getenv("PULSE_DATABASE_MAX_CONNECTIONS"); v != "" {
		if i := parseInt(v, cfg.DB.MaxConnections); i > 0 {
			cfg.DB.MaxConnections = i
		}
	}
	if v := os.Getenv("PULSE_DATABASE_MIN_CONNECTIONS"); v != "" {
		if i := parseInt(v, cfg.DB.MinConnections); i >= 0 {
			cfg.DB.MinConnections = i
		}
	}
	if v := os.Getenv("PULSE_DATABASE_CONN_MAX_LIFETIME"); v != "" {
		if i := parseInt(v, cfg.DB.ConnMaxLifetime); i > 0 {
			cfg.DB.ConnMaxLifetime = i
		}
	}
	if v := os.Getenv("PULSE_DATABASE_CONN_MAX_IDLE_TIME"); v != "" {
		if i := parseInt(v, cfg.DB.ConnMaxIdleTime); i >= 0 {
			cfg.DB.ConnMaxIdleTime = i
		}
	}

	// Cleanup configuration
	if v := os.Getenv("PULSE_CLEANUP_ENABLED"); v != "" {
		cfg.Cleanup.Enabled = parseBool(v, cfg.Cleanup.Enabled)
	}
	if v := os.Getenv("PULSE_CLEANUP_INTERVAL"); v != "" {
		if i := parseInt(v, cfg.Cleanup.IntervalSeconds); i > 0 {
			cfg.Cleanup.IntervalSeconds = i
		}
	}
	if v := os.Getenv("PULSE_CLEANUP_RETENTION_DAYS"); v != "" {
		if i := parseInt(v, cfg.Cleanup.RetentionDays); i > 0 {
			cfg.Cleanup.RetentionDays = i
		}
	}
	if v := os.Getenv("PULSE_CLEANUP_SLOW_THRESHOLD"); v != "" {
		if i := parseInt64(v, cfg.Cleanup.SlowThresholdMs); i >= 0 {
			cfg.Cleanup.SlowThresholdMs = i
		}
	}

	// Log configuration
	if v := os.Getenv("PULSE_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("PULSE_LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
	if v := os.Getenv("PULSE_LOG_OUTPUT"); v != "" {
		cfg.Log.Output = v
	}

	// CORS configuration
	if v := os.Getenv("PULSE_CORS_ALLOWED_ORIGINS"); v != "" {
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
	}
	if v := os.Getenv("PULSE_ADMIN_PASSWORD"); v != "" {
		cfg.Admin.Password = v
	}

	// Session configuration
	if v := os.Getenv("PULSE_SESSION_SECRET"); v != "" {
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
	}
	if v := os.Getenv("PULSE_JWT_PRIVATE_KEY"); v != "" {
		cfg.JWT.PrivateKey = v
	}
	if v := os.Getenv("PULSE_JWT_PUBLIC_KEY"); v != "" {
		cfg.JWT.PublicKey = v
	}
	if v := os.Getenv("PULSE_JWT_KEY_ID"); v != "" {
		cfg.JWT.KeyID = v
	}
	if v := os.Getenv("PULSE_JWT_ACCESS_TOKEN_EXPIRATION_MINUTES"); v != "" {
		if i := parseInt(v, cfg.JWT.AccessTokenExpirationMinutes); i > 0 {
			cfg.JWT.AccessTokenExpirationMinutes = i
		}
	}
	if v := os.Getenv("PULSE_JWT_REFRESH_TOKEN_EXPIRATION_DAYS"); v != "" {
		if i := parseInt(v, cfg.JWT.RefreshTokenExpirationDays); i > 0 {
			cfg.JWT.RefreshTokenExpirationDays = i
		}
	}
	if v := os.Getenv("PULSE_JWT_REFRESH_TOKEN_MAX_VALIDITY_DAYS"); v != "" {
		if i := parseInt(v, cfg.JWT.RefreshTokenMaxValidityDays); i > 0 {
			cfg.JWT.RefreshTokenMaxValidityDays = i
		}
	}

	// Rate Limit configuration
	if v := os.Getenv("PULSE_RATE_LIMIT_LOGIN_MAX_PER_MINUTE"); v != "" {
		if i := parseInt(v, cfg.RateLimit.LoginMaxPerMinute); i > 0 {
			cfg.RateLimit.LoginMaxPerMinute = i
		}
	}
	if v := os.Getenv("PULSE_RATE_LIMIT_LOGIN_MAX_PER_DAY"); v != "" {
		if i := parseInt(v, cfg.RateLimit.LoginMaxPerDay); i > 0 {
			cfg.RateLimit.LoginMaxPerDay = i
		}
	}
	if v := os.Getenv("PULSE_RATE_LIMIT_REFRESH_MAX_PER_MINUTE"); v != "" {
		if i := parseInt(v, cfg.RateLimit.RefreshMaxPerMinute); i > 0 {
			cfg.RateLimit.RefreshMaxPerMinute = i
		}
	}
	if v := os.Getenv("PULSE_RATE_LIMIT_REFRESH_MAX_PER_DAY"); v != "" {
		if i := parseInt(v, cfg.RateLimit.RefreshMaxPerDay); i > 0 {
			cfg.RateLimit.RefreshMaxPerDay = i
		}
	}
	if v := os.Getenv("PULSE_RATE_LIMIT_APIKEY_MAX_PER_MINUTE"); v != "" {
		if i := parseInt(v, cfg.RateLimit.APIKeyMaxPerMinute); i > 0 {
			cfg.RateLimit.APIKeyMaxPerMinute = i
		}
	}

	// Derive CookieSecure from Mode if not explicitly set
	if !cfg.Session.CookieSecure {
		cfg.Session.CookieSecure = cfg.IsProduction()
	}

	// Telemetry configuration
	if v := os.Getenv("PULSE_TELEMETRY_ENABLED"); v != "" {
		cfg.Telemetry.Enabled = parseBool(v, cfg.Telemetry.Enabled)
	}
	if v := os.Getenv("PULSE_TELEMETRY_SERVICE_NAME"); v != "" {
		cfg.Telemetry.ServiceName = v
	}
	if v := os.Getenv("PULSE_TELEMETRY_SERVICE_VERSION"); v != "" {
		cfg.Telemetry.ServiceVersion = v
	}
	if v := os.Getenv("PULSE_TELEMETRY_ENVIRONMENT"); v != "" {
		cfg.Telemetry.Environment = v
	}
	if v := os.Getenv("PULSE_TELEMETRY_OTLP_ENDPOINT"); v != "" {
		cfg.Telemetry.OTLPEndpoint = v
	}
	if v := os.Getenv("PULSE_TELEMETRY_SAMPLING_RATE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 && f <= 1 {
			cfg.Telemetry.SamplingRate = f
		}
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
	if err := c.RateLimit.Validate(); err != nil {
		return fmt.Errorf("rate_limit config invalid: %w", err)
	}
	if err := c.Telemetry.Validate(); err != nil {
		return fmt.Errorf("telemetry config invalid: %w", err)
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
	// Validate RSA keys for RS256 (preferred)
	if c.PrivateKey != "" || c.PublicKey != "" {
		if c.PrivateKey == "" {
			return fmt.Errorf("jwt private_key must be provided when public_key is set")
		}
		if c.PublicKey == "" {
			return fmt.Errorf("jwt public_key must be provided when private_key is set")
		}
		// Basic PEM format validation
		if !strings.Contains(c.PrivateKey, "BEGIN RSA PRIVATE KEY") && !strings.Contains(c.PrivateKey, "BEGIN PRIVATE KEY") {
			return fmt.Errorf("jwt private_key must be in PEM format")
		}
		if !strings.Contains(c.PublicKey, "BEGIN PUBLIC KEY") {
			return fmt.Errorf("jwt public_key must be in PEM format")
		}
	} else {
		// Fallback to legacy Secret validation for backward compatibility
		if c.Secret == "" {
			return fmt.Errorf("jwt secret or rsa key pair (private_key/public_key) must be provided")
		}
		if len(c.Secret) < 64 {
			return fmt.Errorf("jwt secret must be at least 64 bytes (512 bits) for security, got %d bytes", len(c.Secret))
		}
	}
	if c.AccessTokenExpirationMinutes <= 0 {
		return fmt.Errorf("jwt access_token_expiration_minutes must be positive, got %d", c.AccessTokenExpirationMinutes)
	}
	if c.RefreshTokenExpirationDays <= 0 {
		return fmt.Errorf("jwt refresh_token_expiration_days must be positive, got %d", c.RefreshTokenExpirationDays)
	}
	if c.RefreshTokenMaxValidityDays <= 0 {
		return fmt.Errorf("jwt refresh_token_max_validity_days must be positive, got %d", c.RefreshTokenMaxValidityDays)
	}
	if c.RefreshTokenMaxValidityDays < c.RefreshTokenExpirationDays {
		return fmt.Errorf("jwt refresh_token_max_validity_days (%d) must be >= refresh_token_expiration_days (%d)",
			c.RefreshTokenMaxValidityDays, c.RefreshTokenExpirationDays)
	}
	return nil
}

// Validate validates rate limit configuration
func (c *RateLimitConfig) Validate() error {
	if c.LoginMaxPerMinute <= 0 {
		return fmt.Errorf("rate_limit login_max_per_minute must be positive, got %d", c.LoginMaxPerMinute)
	}
	if c.LoginMaxPerDay <= 0 {
		return fmt.Errorf("rate_limit login_max_per_day must be positive, got %d", c.LoginMaxPerDay)
	}
	if c.RefreshMaxPerMinute <= 0 {
		return fmt.Errorf("rate_limit refresh_max_per_minute must be positive, got %d", c.RefreshMaxPerMinute)
	}
	if c.RefreshMaxPerDay <= 0 {
		return fmt.Errorf("rate_limit refresh_max_per_day must be positive, got %d", c.RefreshMaxPerDay)
	}
	if c.APIKeyMaxPerMinute <= 0 {
		return fmt.Errorf("rate_limit apikey_max_per_minute must be positive, got %d", c.APIKeyMaxPerMinute)
	}
	// Validate that per-day limits are greater than per-minute limits
	if c.LoginMaxPerDay < c.LoginMaxPerMinute {
		return fmt.Errorf("rate_limit login_max_per_day (%d) must be >= login_max_per_minute (%d)",
			c.LoginMaxPerDay, c.LoginMaxPerMinute)
	}
	if c.RefreshMaxPerDay < c.RefreshMaxPerMinute {
		return fmt.Errorf("rate_limit refresh_max_per_day (%d) must be >= refresh_max_per_minute (%d)",
			c.RefreshMaxPerDay, c.RefreshMaxPerMinute)
	}
	return nil
}

// Validate validates telemetry configuration.
func (c *TelemetryConfig) Validate() error {
	if c.SamplingRate < 0 || c.SamplingRate > 1 {
		return fmt.Errorf("telemetry sampling_rate must be between 0.0 and 1.0, got %f", c.SamplingRate)
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
			"JWT{Secret:***REDACTED***,AccessTokenExpirationMinutes:%d,RefreshTokenExpirationDays:%d}}",
		c.Server.Port, c.Server.Mode,
		maskURL(c.DB.URL), c.DB.MaxConnections,
		c.Cleanup.Enabled, c.Log.Level,
		c.CORS.AllowedOrigins, c.Admin.Username,
		c.Session.ExpirationHours,
		c.JWT.AccessTokenExpirationMinutes,
		c.JWT.RefreshTokenExpirationDays,
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

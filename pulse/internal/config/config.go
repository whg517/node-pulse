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
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server" mapstructure:"server"`
	DB        DatabaseConfig  `yaml:"database" mapstructure:"database"`
	Cleanup   CleanupConfig   `yaml:"cleanup" mapstructure:"cleanup"`
	Log       LogConfig       `yaml:"log" mapstructure:"log"`
	CORS      CORSConfig      `yaml:"cors" mapstructure:"cors"`
	Admin     AdminConfig     `yaml:"admin" mapstructure:"admin"`
	Session   SessionConfig   `yaml:"session" mapstructure:"session"`
	JWT       JWTConfig       `yaml:"jwt" mapstructure:"jwt"`
	RateLimit RateLimitConfig `yaml:"rate_limit" mapstructure:"rate_limit"`
	Telemetry TelemetryConfig `yaml:"telemetry" mapstructure:"telemetry"`
	Notify    NotifyConfig    `yaml:"notify" mapstructure:"notify"`
}

// TelemetryConfig holds OpenTelemetry / distributed-tracing configuration.
type TelemetryConfig struct {
	// Enabled controls whether distributed tracing is active.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`

	// ServiceName is the logical service name reported in traces (default: "pulse").
	ServiceName string `yaml:"service_name" mapstructure:"service_name"`

	// ServiceVersion is the deployed version string (default: "unknown").
	ServiceVersion string `yaml:"service_version" mapstructure:"service_version"`

	// Environment is the deployment environment, e.g. "production", "staging", "development".
	Environment string `yaml:"environment" mapstructure:"environment"`

	// OTLPEndpoint is the gRPC address of an OTLP-compatible collector,
	// e.g. "localhost:4317".  When empty, traces are written to stdout.
	OTLPEndpoint string `yaml:"otlp_endpoint" mapstructure:"otlp_endpoint"`

	// SamplingRate controls what fraction of requests are traced (0.0 – 1.0, default 1.0).
	SamplingRate float64 `yaml:"sampling_rate" mapstructure:"sampling_rate"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `yaml:"port" mapstructure:"port"`
	ReadTimeout  int    `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	Mode         string `yaml:"mode" mapstructure:"mode"`
	// BaseURL is the externally reachable URL of this Pulse instance (no trailing
	// slash), used to render absolute links in webhook payloads and exported data.
	// Defaults to http://localhost:<port> for local development.
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`
	// ShutdownTimeoutSeconds is the hard cap for graceful shutdown (O-G5).
	// During this window Pulse flushes the batch writer, stops the scheduler,
	// and drains in-flight HTTP requests. 0 keeps the legacy 10s default.
	ShutdownTimeoutSeconds int `yaml:"shutdown_timeout_seconds" mapstructure:"shutdown_timeout_seconds"`
	// TrustedProxies is a comma-separated list of trusted proxy CIDRs (O-G6).
	// When set, gin trusts only these for X-Forwarded-* header parsing, so
	// c.ClientIP() and audit-log IPs reflect the real client behind a reverse
	// proxy instead of the proxy itself. Empty trusts all (legacy behavior,
	// fine for direct exposure but unsafe behind a proxy you don't control).
	TrustedProxies []string `yaml:"trusted_proxies" mapstructure:"trusted_proxies"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string `yaml:"url" mapstructure:"url"`
	MaxConnections  int    `yaml:"max_connections" mapstructure:"max_connections"`
	MinConnections  int    `yaml:"min_connections" mapstructure:"min_connections"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
}

// CleanupConfig holds cleanup task configuration
type CleanupConfig struct {
	Enabled         bool  `yaml:"enabled" mapstructure:"enabled"`
	IntervalSeconds int   `yaml:"interval_seconds" mapstructure:"interval_seconds"`
	RetentionDays   int   `yaml:"retention_days" mapstructure:"retention_days"`
	SlowThresholdMs int64 `yaml:"slow_threshold_ms" mapstructure:"slow_threshold_ms"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `yaml:"level" mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
	Output string `yaml:"output" mapstructure:"output"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string `yaml:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods string `yaml:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders string `yaml:"allowed_headers" mapstructure:"allowed_headers"`
	MaxAge         int    `yaml:"max_age" mapstructure:"max_age"`
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Secret          string `yaml:"secret" mapstructure:"secret"`
	ExpirationHours int    `yaml:"expiration_hours" mapstructure:"expiration_hours"`
	CookieSecure    bool   `yaml:"cookie_secure" mapstructure:"cookie_secure"`
	CookieSameSite  string `yaml:"cookie_samesite" mapstructure:"cookie_samesite"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret                       string `yaml:"secret" mapstructure:"secret"`
	PrivateKey                   string `yaml:"private_key" mapstructure:"private_key"`
	PublicKey                    string `yaml:"public_key" mapstructure:"public_key"`
	AccessTokenExpirationMinutes int    `yaml:"access_token_expiration_minutes" mapstructure:"access_token_expiration_minutes"`
	RefreshTokenExpirationDays   int    `yaml:"refresh_token_expiration_days" mapstructure:"refresh_token_expiration_days"`
	RefreshTokenMaxValidityDays  int    `yaml:"refresh_token_max_validity_days" mapstructure:"refresh_token_max_validity_days"`
	KeyID                        string `yaml:"key_id" mapstructure:"key_id"`
	// Rotation window (O-G3): when provided, ValidateAccessToken accepts
	// tokens signed by the previous key pair as well as the current one.
	// New tokens are always signed with PrivateKey/KeyID. Set all three
	// during a rotation, then unset them once all old tokens have expired.
	PreviousPrivateKey string `yaml:"previous_private_key" mapstructure:"previous_private_key"`
	PreviousPublicKey  string `yaml:"previous_public_key" mapstructure:"previous_public_key"`
	PreviousKeyID      string `yaml:"previous_key_id" mapstructure:"previous_key_id"`
}

// RateLimitConfig holds rate limiting configuration for auth endpoints
type RateLimitConfig struct {
	LoginMaxPerMinute   int `yaml:"login_max_per_minute" mapstructure:"login_max_per_minute"`
	LoginMaxPerDay      int `yaml:"login_max_per_day" mapstructure:"login_max_per_day"`
	RefreshMaxPerMinute int `yaml:"refresh_max_per_minute" mapstructure:"refresh_max_per_minute"`
	RefreshMaxPerDay    int `yaml:"refresh_max_per_day" mapstructure:"refresh_max_per_day"`
	APIKeyMaxPerMinute  int `yaml:"apikey_max_per_minute" mapstructure:"apikey_max_per_minute"`
}

// NotifyConfig holds outbound-email (SMTP) configuration plus the public
// frontend URL used to build links embedded in emails (password reset, etc.).
// When SMTP.Host is empty the system uses a log-only NoopSender.
type NotifyConfig struct {
	SMTP             SMTPConfig `yaml:"smtp" mapstructure:"smtp"`
	PasswordResetURL string     `yaml:"password_reset_url" mapstructure:"password_reset_url"` // e.g. https://app.example.com/reset-password
}

// SMTPConfig holds SMTP transport settings. Mirrors notify.SMTPConfig but kept
// here so the config layer has no dependency on the notify package.
type SMTPConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	From     string `yaml:"from" mapstructure:"from"`
}

var (
	globalConfig *Config
	configMu     sync.RWMutex
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

// Reload re-reads configuration from disk/env and atomically swaps the global
// config. Used by the SIGHUP hot-reload path (server.reloadConfig) so that
// Get() callers (CORS middleware, logger, etc.) observe the new values on the
// next request without a full restart.
//
// Unlike Load(), this bypasses sync.Once and always re-reads. It validates
// the new config before swapping; on validation error the old config stays
// and the error is returned so the caller can log and keep running.
func Reload() (*Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	configMu.Lock()
	globalConfig = cfg
	configMu.Unlock()
	return cfg, nil
}

// Get returns the global configuration instance.
// Should be called after Load() or MustLoad(). Safe for concurrent use;
// hot-reload callers may see a new *Config between requests, which is the
// intended behavior (each request reads the latest values).
func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	if globalConfig == nil {
		panic("Configuration not loaded. Call Load() or MustLoad() first")
	}
	return globalConfig
}

// loadConfig implements the configuration loading logic.
//
// Pipeline: defaults → config file (via Viper, overrides defaults) → environment
// variables (Viper AutomaticEnv, overrides file) → secret generation → derived
// values → validation. See ADR-004 for the contract.
func loadConfig() (*Config, error) {
	// Start with defaults. Viper merges file + env onto this via Unmarshal.
	cfg := defaultConfig()

	// Locate the config file (if any). A missing file is not an error — env vars
	// and defaults still apply.
	configPath, found := findConfigFile()
	if found {
		// Strict YAML check FIRST: reject unknown fields with a precise yaml.v3
		// error before Viper's lenient parse silently drops them. This preserves
		// the KnownFields(true) behavior callers rely on.
		if err := strictYAMLCheck(configPath); err != nil {
			return nil, fmt.Errorf("failed to parse YAML from %s: %w", configPath, err)
		}
	}

	// Always run Viper so environment variables override defaults even when no
	// config file is present. When a file exists it contributes the middle layer
	// (defaults < file < env).
	if err := loadWithViper(cfg, configPath, found); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Normalize BaseURL (strip trailing slash) regardless of source.
	cfg.Server.BaseURL = strings.TrimRight(cfg.Server.BaseURL, "/")

	// Auto-generate secrets if needed.
	if err := generateSecrets(cfg); err != nil {
		return nil, fmt.Errorf("failed to generate secrets: %w", err)
	}

	// Derive CookieSecure from Mode when not explicitly set. Mirrors the previous
	// mergeFromEnv post-processing step.
	if !cfg.Session.CookieSecure {
		cfg.Session.CookieSecure = cfg.IsProduction()
	}

	// Validate final configuration.
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// findConfigFile resolves the config file path per ADR-004 contract 3:
// PULSE_CONFIG_PATH env var → ./pulse.yaml → /etc/node-pulse/pulse.yaml.
// Returns the path and true when a file exists, "" and false otherwise.
func findConfigFile() (string, bool) {
	candidates := []string{}
	if p := os.Getenv("PULSE_CONFIG_PATH"); p != "" {
		candidates = append(candidates, p)
	}
	candidates = append(candidates, "./pulse.yaml", "/etc/node-pulse/pulse.yaml")
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, true
		}
	}
	return "", false
}

// strictYAMLCheck re-decodes the file with yaml.v3 KnownFields(true) to reject
// unknown/misspelled keys. Viper's own parse is lenient and would silently drop
// them, so this keeps the strict contract (and its tests) intact.
func strictYAMLCheck(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	probe := defaultConfig()
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(probe); err != nil {
		return err
	}
	return nil
}

// loadWithViper wires PULSE_<SECTION>_<FIELD> env-var binding via AutomaticEnv
// and unmarshals onto cfg (which already holds defaults). When hasFile is true
// the YAML file contributes the middle layer (defaults < file < env); when
// false, only defaults and env vars apply.
//
// The defaults are re-applied through v.SetDefault so Viper is aware of every
// key — this is what lets AutomaticEnv override values that are never mentioned
// in the config file (Viper only consults the env for keys it already knows).
func loadWithViper(cfg *Config, configPath string, hasFile bool) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PULSE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Backwards-compat env aliases: the original hand-written binding used a few
	// shortened names that do not match the mapstructure tag (from which
	// AutomaticEnv derives the canonical env var). BindEnv wires the legacy name
	// to the canonical key so existing deployments keep working; the canonical
	// name continues to work too. See ADR-004 contract 2.
	v.BindEnv("cleanup.interval_seconds", "PULSE_CLEANUP_INTERVAL_SECONDS", "PULSE_CLEANUP_INTERVAL")    //nolint:errcheck // alias best-effort
	v.BindEnv("cleanup.slow_threshold_ms", "PULSE_CLEANUP_SLOW_THRESHOLD_MS", "PULSE_CLEANUP_SLOW_THRESHOLD") //nolint:errcheck // alias best-effort

	// Register defaults through Viper so AutomaticEnv can override them. The
	// defaults must arrive as a map keyed by the mapstructure tags.
	defaultMap, err := structToMap(defaultConfig())
	if err != nil {
		return fmt.Errorf("failed to encode config defaults: %w", err)
	}
	for key, value := range defaultMap {
		v.SetDefault(key, value)
	}

	if hasFile {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return err
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

// structToMap converts a *Config into the flat map[string]any form Viper's
// SetDefault expects (lowercased, dot-separated keys, e.g. "server.port").
func structToMap(c *Config) (map[string]any, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	m := map[string]any{}
	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, err
	}
	return m, nil
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
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
			Port:                   "6532",
			ReadTimeout:            15,
			WriteTimeout:           15,
			IdleTimeout:            60,
			Mode:                   "debug",
			BaseURL:                "http://localhost:6532", // overwritten by PULSE_SERVER_BASE_URL in production
			ShutdownTimeoutSeconds: 10,                       // O-G5: legacy default
			TrustedProxies:         nil,                      // O-G6: nil trusts all (legacy)
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
			CookieSecure:    false, // Derived from Mode after load
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
			Enabled:        false, // opt-in; set PULSE_TELEMETRY_ENABLED=true to activate
			ServiceName:    "pulse",
			ServiceVersion: "unknown",
			Environment:    "development",
			OTLPEndpoint:   "",  // empty → stdout exporter (dev/debug)
			SamplingRate:   1.0, // trace every request by default
		},
	}
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
	// shutdown_timeout_seconds == 0 is allowed (falls back to the default), but a
	// negative value is a config mistake. Sanity-bound it; operators wanting an
	// instant kill can set 1.
	if c.ShutdownTimeoutSeconds < 0 {
		return fmt.Errorf("shutdown_timeout_seconds must be >= 0, got %d", c.ShutdownTimeoutSeconds)
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

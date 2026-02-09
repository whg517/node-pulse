package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfig_ValidateJWTConfig tests JWT configuration validation
func TestConfig_ValidateJWTConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      JWTConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid JWT config",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", // 64 bytes
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    30,
			},
			wantErr: false,
		},
		{
			name: "empty secret",
			config: JWTConfig{
				Secret:                         "",
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    30,
			},
			wantErr:     true,
			errContains: "jwt secret cannot be empty",
		},
		{
			name: "secret too short (less than 64 bytes)",
			config: JWTConfig{
				Secret:                         "short", // 5 bytes
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    30,
			},
			wantErr:     true,
			errContains: "jwt secret must be at least 64 bytes",
		},
		{
			name: "secret exactly 64 bytes",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    30,
			},
			wantErr: false,
		},
		{
			name: "negative access token expiration",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				AccessTokenExpirationMinutes:   -1,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    30,
			},
			wantErr:     true,
			errContains: "jwt access_token_expiration_minutes must be positive",
		},
		{
			name: "zero access token expiration",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				AccessTokenExpirationMinutes:   0,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    30,
			},
			wantErr:     true,
			errContains: "jwt access_token_expiration_minutes must be positive",
		},
		{
			name: "negative refresh token expiration",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     -1,
				RefreshTokenMaxValidityDays:    30,
			},
			wantErr:     true,
			errContains: "jwt refresh_token_expiration_days must be positive",
		},
		{
			name: "negative refresh token max validity",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    -1,
			},
			wantErr:     true,
			errContains: "jwt refresh_token_max_validity_days must be positive",
		},
		{
			name: "max validity less than expiration (invalid)",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     30,
				RefreshTokenMaxValidityDays:    7, // Less than expiration
			},
			wantErr:     true,
			errContains: "jwt refresh_token_max_validity_days (7) must be >= refresh_token_expiration_days (30)",
		},
		{
			name: "max validity equal to expiration (valid)",
			config: JWTConfig{
				Secret:                         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				AccessTokenExpirationMinutes:   15,
				RefreshTokenExpirationDays:     7,
				RefreshTokenMaxValidityDays:    7, // Equal to expiration
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Validate should return error")
				assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Validate should not return error")
			}
		})
	}
}

// TestConfig_ValidateRateLimitConfig tests rate limit configuration validation
func TestConfig_ValidateRateLimitConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      RateLimitConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid rate limit config",
			config: RateLimitConfig{
				LoginMaxPerMinute:   5,
				LoginMaxPerDay:      100,
				RefreshMaxPerMinute: 10,
				RefreshMaxPerDay:    200,
				APIKeyMaxPerMinute:  11,
			},
			wantErr: false,
		},
		{
			name: "zero login max per minute",
			config: RateLimitConfig{
				LoginMaxPerMinute:   0,
				LoginMaxPerDay:      100,
				RefreshMaxPerMinute: 10,
				RefreshMaxPerDay:    200,
				APIKeyMaxPerMinute:  11,
			},
			wantErr:     true,
			errContains: "ratelimit login_max_per_minute must be positive",
		},
		{
			name: "negative login max per day",
			config: RateLimitConfig{
				LoginMaxPerMinute:   5,
				LoginMaxPerDay:      -1,
				RefreshMaxPerMinute: 10,
				RefreshMaxPerDay:    200,
				APIKeyMaxPerMinute:  11,
			},
			wantErr:     true,
			errContains: "ratelimit login_max_per_day must be positive",
		},
		{
			name: "zero refresh max per minute",
			config: RateLimitConfig{
				LoginMaxPerMinute:   5,
				LoginMaxPerDay:      100,
				RefreshMaxPerMinute: 0,
				RefreshMaxPerDay:    200,
				APIKeyMaxPerMinute:  11,
			},
			wantErr:     true,
			errContains: "ratelimit refresh_max_per_minute must be positive",
		},
		{
			name: "zero refresh max per day",
			config: RateLimitConfig{
				LoginMaxPerMinute:   5,
				LoginMaxPerDay:      100,
				RefreshMaxPerMinute: 10,
				RefreshMaxPerDay:    0,
				APIKeyMaxPerMinute:  11,
			},
			wantErr:     true,
			errContains: "ratelimit refresh_max_per_day must be positive",
		},
		{
			name: "zero apikey max per minute",
			config: RateLimitConfig{
				LoginMaxPerMinute:   5,
				LoginMaxPerDay:      100,
				RefreshMaxPerMinute: 10,
				RefreshMaxPerDay:    200,
				APIKeyMaxPerMinute:  0,
			},
			wantErr:     true,
			errContains: "ratelimit apikey_max_per_minute must be positive",
		},
		{
			name: "login per day less than per minute (invalid)",
			config: RateLimitConfig{
				LoginMaxPerMinute:   100,
				LoginMaxPerDay:      50, // Less than per minute
				RefreshMaxPerMinute: 10,
				RefreshMaxPerDay:    200,
				APIKeyMaxPerMinute:  11,
			},
			wantErr:     true,
			errContains: "ratelimit login_max_per_day (50) must be >= login_max_per_minute (100)",
		},
		{
			name: "refresh per day less than per minute (invalid)",
			config: RateLimitConfig{
				LoginMaxPerMinute:   5,
				LoginMaxPerDay:      100,
				RefreshMaxPerMinute: 200,
				RefreshMaxPerDay:    100, // Less than per minute
				APIKeyMaxPerMinute:  11,
			},
			wantErr:     true,
			errContains: "ratelimit refresh_max_per_day (100) must be >= refresh_max_per_minute (200)",
		},
		{
			name: "login per day equal to per minute (valid)",
			config: RateLimitConfig{
				LoginMaxPerMinute:   50,
				LoginMaxPerDay:      50, // Equal to per minute
				RefreshMaxPerMinute: 10,
				RefreshMaxPerDay:    200,
				APIKeyMaxPerMinute:  11,
			},
			wantErr: false,
		},
		{
			name: "refresh per day equal to per minute (valid)",
			config: RateLimitConfig{
				LoginMaxPerMinute:   5,
				LoginMaxPerDay:      100,
				RefreshMaxPerMinute: 100,
				RefreshMaxPerDay:    100, // Equal to per minute
				APIKeyMaxPerMinute:  11,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Validate should return error")
				assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Validate should not return error")
			}
		})
	}
}

// TestConfig_ValidateAdminConfig tests admin configuration validation
func TestConfig_ValidateAdminConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      AdminConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid admin config",
			config: AdminConfig{
				Username: "admin",
				Password: "SecurePass123",
			},
			wantErr: false,
		},
		{
			name: "empty username",
			config: AdminConfig{
				Username: "",
				Password: "SecurePass123",
			},
			wantErr:     true,
			errContains: "admin username cannot be empty",
		},
		{
			name: "empty password",
			config: AdminConfig{
				Username: "admin",
				Password: "",
			},
			wantErr:     true,
			errContains: "admin password cannot be empty",
		},
		{
			name: "both empty",
			config: AdminConfig{
				Username: "",
				Password: "",
			},
			wantErr:     true,
			errContains: "admin username cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Validate should return error")
				assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Validate should not return error")
			}
		})
	}
}

// TestConfig_ValidateSessionConfig tests session configuration validation
func TestConfig_ValidateSessionConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      SessionConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid session config",
			config: SessionConfig{
				Secret:          "0123456789abcdef0123456789abcdef0123456789abcdef",
				ExpirationHours: 24,
				CookieSecure:    true,
				CookieSameSite:  "Lax",
			},
			wantErr: false,
		},
		{
			name: "empty secret",
			config: SessionConfig{
				Secret:          "",
				ExpirationHours: 24,
				CookieSecure:    true,
				CookieSameSite:  "Lax",
			},
			wantErr:     true,
			errContains: "session secret cannot be empty",
		},
		{
			name: "zero expiration hours",
			config: SessionConfig{
				Secret:          "0123456789abcdef0123456789abcdef0123456789abcdef",
				ExpirationHours: 0,
				CookieSecure:    true,
				CookieSameSite:  "Lax",
			},
			wantErr:     true,
			errContains: "session expiration_hours must be positive",
		},
		{
			name: "negative expiration hours",
			config: SessionConfig{
				Secret:          "0123456789abcdef0123456789abcdef0123456789abcdef",
				ExpirationHours: -1,
				CookieSecure:    true,
				CookieSameSite:  "Lax",
			},
			wantErr:     true,
			errContains: "session expiration_hours must be positive",
		},
		{
			name: "invalid SameSite value",
			config: SessionConfig{
				Secret:          "0123456789abcdef0123456789abcdef0123456789abcdef",
				ExpirationHours: 24,
				CookieSecure:    true,
				CookieSameSite:  "Invalid",
			},
			wantErr:     true,
			errContains: "cookie_samesite must be one of: Strict, Lax, None",
		},
		{
			name: "valid SameSite Strict",
			config: SessionConfig{
				Secret:          "0123456789abcdef0123456789abcdef0123456789abcdef",
				ExpirationHours: 24,
				CookieSecure:    true,
				CookieSameSite:  "Strict",
			},
			wantErr: false,
		},
		{
			name: "valid SameSite None",
			config: SessionConfig{
				Secret:          "0123456789abcdef0123456789abcdef0123456789abcdef",
				ExpirationHours: 24,
				CookieSecure:    true,
				CookieSameSite:  "None",
			},
			wantErr: false,
		},
		{
			name: "empty SameSite (valid)",
			config: SessionConfig{
				Secret:          "0123456789abcdef0123456789abcdef0123456789abcdef",
				ExpirationHours: 24,
				CookieSecure:    true,
				CookieSameSite:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Validate should return error")
				assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Validate should not return error")
			}
		})
	}
}

// TestConfig_ValidateCleanupConfig tests cleanup configuration validation
func TestConfig_ValidateCleanupConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      CleanupConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid cleanup config (disabled)",
			config: CleanupConfig{
				Enabled:         false,
				IntervalSeconds: 3600,
				RetentionDays:   90,
			},
			wantErr: false,
		},
		{
			name: "valid cleanup config (enabled)",
			config: CleanupConfig{
				Enabled:         true,
				IntervalSeconds: 3600,
				RetentionDays:   90,
			},
			wantErr: false,
		},
		{
			name: "negative interval seconds",
			config: CleanupConfig{
				Enabled:         true,
				IntervalSeconds: -1,
				RetentionDays:   90,
			},
			wantErr:     true,
			errContains: "interval_seconds must be positive",
		},
		{
			name: "zero interval seconds when enabled",
			config: CleanupConfig{
				Enabled:         true,
				IntervalSeconds: 0,
				RetentionDays:   90,
			},
			wantErr:     true,
			errContains: "interval_seconds must be positive",
		},
		{
			name: "negative retention days",
			config: CleanupConfig{
				Enabled:         true,
				IntervalSeconds: 3600,
				RetentionDays:   -1,
			},
			wantErr:     true,
			errContains: "retention_days must be positive",
		},
		{
			name: "zero retention days",
			config: CleanupConfig{
				Enabled:         true,
				IntervalSeconds: 3600,
				RetentionDays:   0,
			},
			wantErr:     true,
			errContains: "retention_days must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Validate should return error")
				assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Validate should not return error")
			}
		})
	}
}

// TestConfig_LoadFromEnv tests loading configuration from environment variables
func TestConfig_LoadFromEnv(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"PULSE_JWT_SECRET",
		"PULSE_RATELIMIT_LOGIN_MAX_PER_MINUTE",
		"PULSE_RATELIMIT_REFRESH_MAX_PER_MINUTE",
	}
	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
	}

	// Clean up after test
	defer func() {
		for envVar, value := range originalEnv {
			if value == "" {
				os.Unsetenv(envVar)
			} else {
				os.Setenv(envVar, value)
			}
		}
		Reset()
	}()

	tests := []struct {
		name     string
		setEnv   map[string]string
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name: "load JWT secret from env",
			setEnv: map[string]string{
				"PULSE_JWT_SECRET": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", cfg.JWT.Secret)
			},
		},
		{
			name: "load rate limit config from env",
			setEnv: map[string]string{
				"PULSE_RATELIMIT_LOGIN_MAX_PER_MINUTE":   "10",
				"PULSE_RATELIMIT_REFRESH_MAX_PER_MINUTE": "20",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 10, cfg.RateLimit.LoginMaxPerMinute)
				assert.Equal(t, 20, cfg.RateLimit.RefreshMaxPerMinute)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and set env vars
			Reset()
			for key, value := range tt.setEnv {
				os.Setenv(key, value)
			}

			// Load config
			cfg, err := Load()
			assert.NoError(t, err, "Load should not return error")

			// Validate
			tt.validate(t, cfg)
		})
	}
}

// TestConfig_GenerateRandomSecret tests secret generation (internal function via load)
func TestConfig_GenerateRandomSecret(t *testing.T) {
	// Reset and set empty JWT secret to trigger auto-generation
	Reset()
	os.Unsetenv("PULSE_JWT_SECRET")

	cfg, err := Load()
	assert.NoError(t, err, "Load should not return error")
	assert.NotEmpty(t, cfg.JWT.Secret, "JWT secret should be auto-generated")
	// JWT secret is 64 bytes (512 bits) encoded as hex = 128 characters
	assert.Equal(t, 128, len(cfg.JWT.Secret), "Generated JWT secret should be 128 hex characters (64 bytes)")

	// Verify it's hexadecimal
	for _, c := range cfg.JWT.Secret {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"Secret should contain only hexadecimal characters")
	}
}

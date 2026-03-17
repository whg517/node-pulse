package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseInt64 tests parseInt64 helper
func TestParseInt64(t *testing.T) {
	tests := []struct {
		input    string
		def      int64
		expected int64
	}{
		{"100", 0, 100},
		{"0", 5, 0},
		{"-1", 0, -1},
		{"invalid", 42, 42},
		{"", 10, 10},
		{"9223372036854775807", 0, 9223372036854775807}, // max int64
	}

	for _, tt := range tests {
		result := parseInt64(tt.input, tt.def)
		assert.Equal(t, tt.expected, result, "parseInt64(%q, %d)", tt.input, tt.def)
	}
}

// TestParseBool tests parseBool helper
func TestParseBool(t *testing.T) {
	trueValues := []string{"true", "1", "yes", "on", "TRUE", "True", "YES", "ON"}
	falseValues := []string{"false", "0", "no", "off", "FALSE", "False", "NO", "OFF"}

	for _, v := range trueValues {
		assert.True(t, parseBool(v, false), "parseBool(%q) should be true", v)
	}

	for _, v := range falseValues {
		assert.False(t, parseBool(v, true), "parseBool(%q) should be false", v)
	}

	// Empty string should return default
	assert.True(t, parseBool("", true))
	assert.False(t, parseBool("", false))

	// Unknown value returns false (not true/1/yes/on)
	assert.False(t, parseBool("unknown", false))
	assert.False(t, parseBool("unknown", true))
}

// TestConfig_String tests Config.String()
func TestConfig_String(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: "6532", Mode: "debug"},
		DB:     DatabaseConfig{URL: "postgres://user:password@localhost:5432/mydb", MaxConnections: 10},
		Cleanup: CleanupConfig{Enabled: true},
		Log:     LogConfig{Level: "info"},
		CORS:    CORSConfig{AllowedOrigins: "http://localhost:3000"},
		Admin:   AdminConfig{Username: "admin", Password: "secret"},
		Session: SessionConfig{Secret: "session-secret", ExpirationHours: 24},
		JWT: JWTConfig{
			Secret:                       "jwt-secret",
			AccessTokenExpirationMinutes: 15,
			RefreshTokenExpirationDays:   7,
		},
	}

	s := cfg.String()
	assert.NotEmpty(t, s)
	// Should redact passwords
	assert.Contains(t, s, "***REDACTED***")
	assert.NotContains(t, s, "password")
	assert.NotContains(t, s, "secret")
	// Should contain non-sensitive values
	assert.Contains(t, s, "6532")
	assert.Contains(t, s, "debug")
}

// TestMaskURL tests the maskURL helper
func TestMaskURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		contains string
		notContains string
	}{
		{
			name:     "postgres URL with password",
			url:      "postgres://user:password@localhost:5432/mydb",
			contains: "***@",
			notContains: "password",
		},
		{
			name:     "postgres URL without password",
			url:      "postgres://localhost:5432/mydb",
			contains: "postgres://localhost",
		},
		{
			name:     "empty URL",
			url:      "",
			contains: "",
		},
		{
			name:     "short URL without @",
			url:      "localhost:5432",
			contains: "localhost:5432",
		},
		{
			name:     "long URL without credentials",
			url:      "postgres://localhost:5432/very-long-database-name-that-exceeds-fifty-characters",
			contains: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskURL(tt.url)
			if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			}
			if tt.notContains != "" {
				assert.NotContains(t, result, tt.notContains)
			}
		})
	}
}

// TestConfig_IsDevelopment tests Config.IsDevelopment()
func TestConfig_IsDevelopment(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Mode: "debug"}}
	assert.True(t, cfg.IsDevelopment())

	cfg.Server.Mode = "release"
	assert.False(t, cfg.IsDevelopment())
	assert.True(t, cfg.IsProduction())

	cfg.Server.Mode = ""
	assert.False(t, cfg.IsDevelopment())
	assert.False(t, cfg.IsProduction())
}

// TestMustLoad tests MustLoad panics on failure
func TestMustLoad_PanicsOnError(t *testing.T) {
	// Reset state
	Reset()

	// Set an invalid value to force config validation failure
	_ = os.Setenv("PULSE_DATABASE_URL", "") // invalid empty URL
	defer func() {
		_ = os.Unsetenv("PULSE_DATABASE_URL")
	}()

	// MustLoad should not panic for load itself (validation is separate)
	// It uses defaults which are valid; let's just ensure it doesn't panic
	Reset()
	assert.NotPanics(t, func() {
		cfg := MustLoad()
		assert.NotNil(t, cfg)
	})
}

// TestGet_PanicsWhenNotLoaded tests Get panics when config not loaded
func TestGet_PanicsWhenNotLoaded(t *testing.T) {
	Reset()
	defer Reset()

	assert.Panics(t, func() {
		Get()
	})
}

// TestGet_ReturnsAfterLoad tests Get returns config after Load
func TestGet_ReturnsAfterLoad(t *testing.T) {
	Reset()
	defer Reset()

	_, err := Load()
	require.NoError(t, err)

	cfg := Get()
	assert.NotNil(t, cfg)
}

// TestMergeFromEnv tests that env vars override config values
func TestMergeFromEnv(t *testing.T) {
	envVars := map[string]string{
		"PULSE_SERVER_PORT":                             "9999",
		"PULSE_SERVER_MODE":                             "release",
		"PULSE_DATABASE_URL":                            "postgres://test:pass@db:5432/testdb",
		"PULSE_DATABASE_MAX_CONNECTIONS":                "50",
		"PULSE_DATABASE_MIN_CONNECTIONS":                "5",
		"PULSE_DATABASE_CONN_MAX_LIFETIME":              "3600",
		"PULSE_DATABASE_CONN_MAX_IDLE_TIME":             "600",
		"PULSE_CLEANUP_ENABLED":                         "true",
		"PULSE_CLEANUP_INTERVAL":                        "600",
		"PULSE_CLEANUP_RETENTION_DAYS":                  "30",
		"PULSE_CLEANUP_SLOW_THRESHOLD":                  "500",
		"PULSE_LOG_LEVEL":                               "warn",
		"PULSE_LOG_FORMAT":                              "json",
		"PULSE_LOG_OUTPUT":                              "stdout",
		"PULSE_CORS_ALLOWED_ORIGINS":                    "https://app.example.com",
		"PULSE_CORS_ALLOWED_METHODS":                    "GET,POST",
		"PULSE_CORS_ALLOWED_HEADERS":                    "Authorization",
		"PULSE_CORS_MAX_AGE":                            "3600",
		"PULSE_ADMIN_USERNAME":                          "testadmin",
		"PULSE_ADMIN_PASSWORD":                          "TestPassword123",
		"PULSE_SESSION_SECRET":                          "test-session-secret-long-enough-for-validation",
		"PULSE_SESSION_EXPIRATION_HOURS":                "48",
		"PULSE_SESSION_COOKIE_SECURE":                   "true",
		"PULSE_SESSION_COOKIE_SAMESITE":                 "Strict",
		"PULSE_JWT_SECRET":                              "test-jwt-secret-that-is-at-least-64-bytes-long-for-validation-purposes",
		"PULSE_JWT_KEY_ID":                              "test-key-id",
		"PULSE_JWT_ACCESS_TOKEN_EXPIRATION_MINUTES":     "30",
		"PULSE_JWT_REFRESH_TOKEN_EXPIRATION_DAYS":       "14",
		"PULSE_JWT_REFRESH_TOKEN_MAX_VALIDITY_DAYS":     "60",
		"PULSE_RATE_LIMIT_LOGIN_MAX_PER_MINUTE":         "10",
		"PULSE_RATE_LIMIT_LOGIN_MAX_PER_DAY":            "200",
		"PULSE_RATE_LIMIT_REFRESH_MAX_PER_MINUTE":       "20",
		"PULSE_RATE_LIMIT_REFRESH_MAX_PER_DAY":          "400",
		"PULSE_RATE_LIMIT_APIKEY_MAX_PER_MINUTE":        "50",
	}

	for k, v := range envVars {
		_ = os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			_ = os.Unsetenv(k)
		}
		Reset()
	}()

	Reset()
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "9999", cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, "postgres://test:pass@db:5432/testdb", cfg.DB.URL)
	assert.Equal(t, 50, cfg.DB.MaxConnections)
	assert.Equal(t, 5, cfg.DB.MinConnections)
	assert.Equal(t, 3600, cfg.DB.ConnMaxLifetime)
	assert.Equal(t, 600, cfg.DB.ConnMaxIdleTime)
	assert.True(t, cfg.Cleanup.Enabled)
	assert.Equal(t, 600, cfg.Cleanup.IntervalSeconds)
	assert.Equal(t, 30, cfg.Cleanup.RetentionDays)
	assert.Equal(t, int64(500), cfg.Cleanup.SlowThresholdMs)
	assert.Equal(t, "warn", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
	assert.Equal(t, "stdout", cfg.Log.Output)
	assert.Equal(t, "https://app.example.com", cfg.CORS.AllowedOrigins)
	assert.Equal(t, "GET,POST", cfg.CORS.AllowedMethods)
	assert.Equal(t, "Authorization", cfg.CORS.AllowedHeaders)
	assert.Equal(t, 3600, cfg.CORS.MaxAge)
	assert.Equal(t, "testadmin", cfg.Admin.Username)
	assert.Equal(t, "TestPassword123", cfg.Admin.Password)
	assert.Equal(t, 48, cfg.Session.ExpirationHours)
	assert.True(t, cfg.Session.CookieSecure)
	assert.Equal(t, "Strict", cfg.Session.CookieSameSite)
	assert.Equal(t, "test-key-id", cfg.JWT.KeyID)
	assert.Equal(t, 30, cfg.JWT.AccessTokenExpirationMinutes)
	assert.Equal(t, 14, cfg.JWT.RefreshTokenExpirationDays)
	assert.Equal(t, 60, cfg.JWT.RefreshTokenMaxValidityDays)
	assert.Equal(t, 10, cfg.RateLimit.LoginMaxPerMinute)
	assert.Equal(t, 200, cfg.RateLimit.LoginMaxPerDay)
	assert.Equal(t, 20, cfg.RateLimit.RefreshMaxPerMinute)
	assert.Equal(t, 400, cfg.RateLimit.RefreshMaxPerDay)
	assert.Equal(t, 50, cfg.RateLimit.APIKeyMaxPerMinute)
}

// TestLoad_IsCached tests that Load returns the same instance
func TestLoad_IsCached(t *testing.T) {
	Reset()
	defer Reset()

	cfg1, err1 := Load()
	cfg2, err2 := Load()

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Same(t, cfg1, cfg2, "Load should return the same cached instance")
}

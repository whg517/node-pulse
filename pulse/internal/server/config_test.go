package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whg517/node-pulse/pulse/internal/config"
)

// newTestConfig builds a *Config populated with known values so getter tests
// are deterministic and do not depend on config.MustLoad (env / files).
func newTestConfig(mode string) *Config {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         "6532",
			ReadTimeout:  15,
			WriteTimeout: 20,
			IdleTimeout:  120,
			Mode:         mode,
		},
		DB: config.DatabaseConfig{
			URL:            "postgres://user:pass@localhost:5432/db?sslmode=disable",
			MaxConnections: 10,
			MinConnections: 2,
		},
		CORS: config.CORSConfig{
			AllowedOrigins: "http://localhost:5173",
			AllowedMethods: "GET,POST",
			AllowedHeaders: "Content-Type",
			MaxAge:         300,
		},
		Admin:   config.AdminConfig{Username: "admin", Password: "secret"},
		Session: config.SessionConfig{Secret: "session-secret", ExpirationHours: 24},
	}
	return &Config{Config: cfg}
}

func TestConfig_GetPort(t *testing.T) {
	c := newTestConfig("debug")
	assert.Equal(t, "6532", c.GetPort())
}

func TestConfig_GetDatabaseURL(t *testing.T) {
	c := newTestConfig("debug")
	assert.Equal(t, "postgres://user:pass@localhost:5432/db?sslmode=disable", c.GetDatabaseURL())
}

func TestConfig_GetTimeouts(t *testing.T) {
	c := newTestConfig("debug")
	assert.Equal(t, 15, c.GetReadTimeout())
	assert.Equal(t, 20, c.GetWriteTimeout())
	assert.Equal(t, 120, c.GetIdleTimeout())
}

func TestConfig_GetCORSConfig(t *testing.T) {
	c := newTestConfig("debug")
	cors := c.GetCORSConfig()
	assert.NotNil(t, cors)
	assert.Equal(t, "http://localhost:5173", cors.AllowedOrigins)
	assert.Equal(t, 300, cors.MaxAge)
}

func TestConfig_GetAdminConfig(t *testing.T) {
	c := newTestConfig("debug")
	admin := c.GetAdminConfig()
	assert.NotNil(t, admin)
	assert.Equal(t, "admin", admin.Username)
	assert.Equal(t, "secret", admin.Password)
}

func TestConfig_GetSessionConfig(t *testing.T) {
	c := newTestConfig("debug")
	session := c.GetSessionConfig()
	assert.NotNil(t, session)
	assert.Equal(t, "session-secret", session.Secret)
	assert.Equal(t, 24, session.ExpirationHours)
}

func TestConfig_IsProduction(t *testing.T) {
	assert.True(t, newTestConfig("release").IsProduction())
	assert.False(t, newTestConfig("debug").IsProduction())
}

func TestConfig_IsDevelopment(t *testing.T) {
	assert.True(t, newTestConfig("debug").IsDevelopment())
	assert.False(t, newTestConfig("release").IsDevelopment())
}

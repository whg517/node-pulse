package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuilder_WithPort verifies the fluent setter updates the underlying config
// and returns the same *Builder for chaining.
func TestBuilder_WithPort(t *testing.T) {
	b := &Builder{config: newTestConfig("debug")}

	ret := b.WithPort("9999")
	require.NotNil(t, ret)
	assert.Same(t, b, ret, "WithPort should return the same builder for chaining")
	assert.Equal(t, "9999", b.config.Server.Port)
}

// TestBuilder_WithDatabase verifies the fluent setter updates the DB URL.
func TestBuilder_WithDatabase(t *testing.T) {
	b := &Builder{config: newTestConfig("debug")}

	const newURL = "postgres://test:test@db.example:5432/nodepulse?sslmode=require"
	ret := b.WithDatabase(newURL)
	require.NotNil(t, ret)
	assert.Same(t, b, ret)
	assert.Equal(t, newURL, b.config.DB.URL)
}

// TestBuilder_ChainedSetters verifies setters compose correctly.
func TestBuilder_ChainedSetters(t *testing.T) {
	b := &Builder{config: newTestConfig("debug")}

	b.WithPort("8080").WithDatabase("postgres://x")

	assert.Equal(t, "8080", b.config.Server.Port)
	assert.Equal(t, "postgres://x", b.config.DB.URL)
}

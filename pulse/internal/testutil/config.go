package testutil

import (
	"os"

	"github.com/whg517/node-pulse/pulse/internal/config"
)

const (
	// defaultTestDBURL is the default database URL for testing
	// This matches the configuration in docker-compose.test.yml
	defaultTestDBURL = "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
)

// GetTestDBURL returns the test database URL.
// Priority:
// 1. TEST_DATABASE_URL environment variable (test convention, kept for compatibility)
// 2. PULSE_DATABASE_URL environment variable
// 3. Default test database URL (docker-compose.test.yml)
func GetTestDBURL() string {
	// Check test-specific convention first
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	// Check unified PULSE_ variable
	if url := os.Getenv("PULSE_DATABASE_URL"); url != "" {
		return url
	}
	// Check legacy DATABASE_URL for backward compatibility
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return defaultTestDBURL
}

// SetupTestConfig initializes configuration for testing
// This resets the global config and sets test environment variables
// Call this at the beginning of your test if you need custom config
//
// Example:
//
//	func TestSomething(t *testing.T) {
//	    testutil.SetupTestConfig()
//	    defer testutil.TeardownTestConfig()
//
//	    // Set test-specific env vars
//	    os.Setenv("PULSE_DATABASE_URL", "postgres://...")
//
//	    // Load config
//	    cfg, err := config.Load()
//	    assert.NoError(t, err)
//	}
func SetupTestConfig() {
	// Reset global config to clean state
	config.Reset()

	// Set test mode by default
	if os.Getenv("PULSE_SERVER_MODE") == "" {
		os.Setenv("PULSE_SERVER_MODE", "test")
	}
}

// TeardownTestConfig cleans up configuration after testing
// Call this in defer after SetupTestConfig
func TeardownTestConfig() {
	// Reset global config
	config.Reset()

	// Clean up test environment variables
	testEnvVars := []string{
		"PULSE_SERVER_MODE",
		"PULSE_DATABASE_URL",
		"PULSE_ADMIN_USERNAME",
		"PULSE_ADMIN_PASSWORD",
		"PULSE_SESSION_SECRET",
		"PULSE_JWT_SECRET",
		"TEST_DATABASE_URL",
	}

	for _, envVar := range testEnvVars {
		os.Unsetenv(envVar)
	}
}

// MustLoadTestConfig loads configuration for testing or panics
// This is a convenience function for tests that need config loaded
//
// Example:
//
//	func TestWithConfig(t *testing.T) {
//	    cfg := testutil.MustLoadTestConfig()
//	    defer testutil.TeardownTestConfig()
//	    // Use cfg...
//	}
func MustLoadTestConfig() *config.Config {
	SetupTestConfig()
	return config.MustLoad()
}

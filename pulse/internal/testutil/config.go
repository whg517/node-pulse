package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/config"
)

const (
	// defaultTestDBURL is the default database URL for testing
	// This matches the configuration in docker-compose.test.yml
	defaultTestDBURL = "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
)

// RequireDB skips the calling test unless a live test database is reachable.
// It is the single entry point integration setup helpers should call first so
// that `go test -short` (and CI unit-test runs without a Postgres service) do
// not FAIL on connection-refused errors but SKIP cleanly instead.
//
// The skip triggers when:
//   - testing.Short() is true (unit/short runs), or
//   - the test database is not reachable within a short probe window.
func RequireDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode (requires database)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping integration test: cannot create pool: %v", err)
		return nil
	}

	// pgxpool is lazy: force a real round-trip so an unreachable DB is detected
	// now rather than failing later as a hard error.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping integration test: database not reachable: %v", err)
		return nil
	}

	return pool
}

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
		_ = os.Setenv("PULSE_SERVER_MODE", "test")
	}
}

// TeardownTestConfig cleans up configuration after testing
// Call this in defer after SetupTestConfig
func TeardownTestConfig() {
	// Reset global config
	config.Reset()

	// Only unset the variable SetupTestConfig itself set. Other PULSE_*
	// variables (e.g. PULSE_DATABASE_URL, TEST_DATABASE_URL, admin credentials)
	// are typically provided by the caller (CI / dev shell) for the whole test
	// run; unsetting them here caused sibling db tests run in the same process
	// to fall back to the 5432 default and skip after the first test cleaned up.
	_ = os.Unsetenv("PULSE_SERVER_MODE")
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

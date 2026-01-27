package testutil

import (
	"os"
)

const (
	// defaultTestDBURL is the default database URL for testing
	// This matches the configuration in docker-compose.test.yml
	defaultTestDBURL = "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
)

// GetTestDBURL returns the test database URL.
// Priority:
// 1. TEST_DATABASE_URL environment variable
// 2. DATABASE_URL environment variable
// 3. Default test database URL (docker-compose.test.yml)
func GetTestDBURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return defaultTestDBURL
}

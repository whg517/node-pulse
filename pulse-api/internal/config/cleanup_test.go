package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCleanupConfig_Defaults(t *testing.T) {
	// Clear environment variables
	clearCleanupEnv()

	cfg, err := LoadCleanupConfig()
	require.NoError(t, err)

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 3600, cfg.IntervalSeconds)
	assert.Equal(t, 7, cfg.RetentionDays)
	assert.Equal(t, int64(30000), cfg.SlowThresholdMs)
}

func TestLoadCleanupConfig_CustomValues(t *testing.T) {
	clearCleanupEnv()

	// Set custom environment variables
	os.Setenv("CLEANUP_ENABLED", "false")
	os.Setenv("CLEANUP_INTERVAL", "7200")
	os.Setenv("CLEANUP_RETENTION_DAYS", "14")
	os.Setenv("CLEANUP_SLOW_THRESHOLD", "60000")

	defer clearCleanupEnv()

	cfg, err := LoadCleanupConfig()
	require.NoError(t, err)

	assert.False(t, cfg.Enabled)
	assert.Equal(t, 7200, cfg.IntervalSeconds)
	assert.Equal(t, 14, cfg.RetentionDays)
	assert.Equal(t, int64(60000), cfg.SlowThresholdMs)
}

func TestLoadCleanupConfig_InvalidInterval(t *testing.T) {
	clearCleanupEnv()
	os.Setenv("CLEANUP_INTERVAL", "0")
	defer clearCleanupEnv()

	cfg, err := LoadCleanupConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "interval_seconds must be positive")
}

func TestLoadCleanupConfig_InvalidRetention(t *testing.T) {
	clearCleanupEnv()
	os.Setenv("CLEANUP_RETENTION_DAYS", "-1")
	defer clearCleanupEnv()

	cfg, err := LoadCleanupConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "retention_days must be positive")
}

func TestLoadCleanupConfig_InvalidSlowThreshold(t *testing.T) {
	clearCleanupEnv()
	os.Setenv("CLEANUP_SLOW_THRESHOLD", "-100")
	defer clearCleanupEnv()

	cfg, err := LoadCleanupConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "slow_threshold_ms cannot be negative")
}

func TestLoadCleanupConfig_BoolParsing(t *testing.T) {
	clearCleanupEnv()

	testCases := []struct {
		value     string
		expected  bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"", true}, // default
	}

	for i, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			os.Setenv("CLEANUP_ENABLED", tc.value)
			cfg, err := LoadCleanupConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.Enabled, "Test case %d failed", i)
		})
	}

	clearCleanupEnv()
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
		SlowThresholdMs: 30000,
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func clearCleanupEnv() {
	os.Unsetenv("CLEANUP_ENABLED")
	os.Unsetenv("CLEANUP_INTERVAL")
	os.Unsetenv("CLEANUP_RETENTION_DAYS")
	os.Unsetenv("CLEANUP_SLOW_THRESHOLD")
}

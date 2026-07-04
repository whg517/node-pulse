package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// newViperWithDefaults builds a Viper instance wired exactly like LoadConfig
// (prefix, replacer, AutomaticEnv, setDefaults) so the bool/zero-value semantics
// tested here match production. Callers may read a YAML string into it.
func newViperWithDefaults(t *testing.T, yaml string) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("BEACON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)
	if yaml != "" {
		if err := v.ReadConfig(strings.NewReader(yaml)); err != nil {
			t.Fatalf("failed to read config: %v", err)
		}
	}
	return v
}

// TestSetDefaults_BoolExplicitFalse verifies that an explicit `false` in the
// config file wins over the SetDefault(true). This is the core invariant the
// old `if !v.IsSet(...)` backfill enforced — the migration to SetDefault must
// not regress it.
func TestSetDefaults_BoolExplicitFalse(t *testing.T) {
	v := newViperWithDefaults(t, `
compression:
  enabled: false
`)
	if v.GetBool("compression.enabled") {
		t.Error("explicit `compression.enabled: false` must win over the SetDefault(true)")
	}
}

// TestSetDefault_BoolAbsent verifies the default applies when the key is absent.
func TestSetDefault_BoolAbsent(t *testing.T) {
	v := newViperWithDefaults(t, `
compression:
  level: 3
`)
	if !v.GetBool("compression.enabled") {
		t.Error("absent compression.enabled must fall back to the SetDefault(true)")
	}
}

// TestSetDefaults_NumericExplicitZero verifies that an explicit `0` is preserved
// (the old `if x == 0` backfill would have wrongly overwritten it with the
// default). metrics_port: 0 is invalid in production (validation catches it),
// but the point is the config layer must not silently mutate it.
func TestSetDefaults_NumericExplicitZero(t *testing.T) {
	v := newViperWithDefaults(t, `metrics_port: 0`)
	if got := v.GetInt("metrics_port"); got != 0 {
		t.Errorf("explicit metrics_port: 0 must be preserved, got %d (default leaked through)", got)
	}
}

// TestSetDefaults_Registered verifies every subsystem default is reachable after
// loading a minimal valid config (only the required fields). This guards against
// accidentally dropping a SetDefault line during refactors.
func TestSetDefaults_AllPresent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	minimal := `
pulse_server: "http://localhost:6532"
node_id: "defaults-test"
node_name: "Defaults Test"
`
	if err := os.WriteFile(configPath, []byte(minimal), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"metrics_port", cfg.MetricsPort, 2112},
		{"metrics_enabled", cfg.MetricsEnabled, true},
		{"metrics_update_seconds", cfg.MetricsUpdateSeconds, 10},
		{"log_level", cfg.LogLevel, "INFO"},
		{"log_file", cfg.LogFile, "/var/log/beacon/beacon.log"},
		{"log_max_size", cfg.LogMaxSize, 10},
		{"log_max_age", cfg.LogMaxAge, 7},
		{"log_max_backups", cfg.LogMaxBackups, 10},
		{"resource_monitor.check_interval_seconds", cfg.ResourceMonitor.CheckIntervalSeconds, 60},
		{"resource_monitor.thresholds.cpu", cfg.ResourceMonitor.Thresholds.CPUMicrocores, 100},
		{"resource_monitor.thresholds.mem", cfg.ResourceMonitor.Thresholds.MemoryMB, 100},
		{"resource_monitor.degraded.cpu", cfg.ResourceMonitor.Degradation.DegradedLevel.CPUMicrocores, 200},
		{"resource_monitor.degraded.mem", cfg.ResourceMonitor.Degradation.DegradedLevel.MemoryMB, 150},
		{"resource_monitor.degraded.mult", cfg.ResourceMonitor.Degradation.DegradedLevel.IntervalMultiplier, 2},
		{"resource_monitor.critical.cpu", cfg.ResourceMonitor.Degradation.CriticalLevel.CPUMicrocores, 300},
		{"resource_monitor.critical.mem", cfg.ResourceMonitor.Degradation.CriticalLevel.MemoryMB, 200},
		{"resource_monitor.critical.mult", cfg.ResourceMonitor.Degradation.CriticalLevel.IntervalMultiplier, 3},
		{"resource_monitor.recovery", cfg.ResourceMonitor.Degradation.Recovery.ConsecutiveNormalChecks, 3},
		{"resource_monitor.alerting", cfg.ResourceMonitor.Alerting.SuppressionWindowSeconds, 300},
		{"mode.mode", cfg.Mode.Mode, ModeRegistered},
		{"mode.config_check", cfg.Mode.ConfigCheckIntervalSeconds, 60},
		{"mode.degraded_threshold", cfg.Mode.DegradedModeThreshold, 3},
		{"compression.enabled", cfg.Compression.Enabled, true},
		{"compression.level", cfg.Compression.Level, 6},
		{"compression.min_size", cfg.Compression.MinSizeBytes, 1024},
		{"resume.enabled", cfg.Resume.Enabled, true},
		{"resume.max_cache", cfg.Resume.MaxCacheSizeBytes, int64(10 * 1024 * 1024)},
		{"resume.cache_path", cfg.Resume.CacheFilePath, "/var/lib/beacon/resume_cache.dat"},
		{"resume.alert_priority", cfg.Resume.AlertPriorityMode, true},
		{"resume.alert_reserve", cfg.Resume.AlertReservePercent, 30},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("default %s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

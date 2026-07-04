package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSaveConfig_Success tests successful config saving
func TestSaveConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "saved.yaml")

	// First create a config file so viper can write to it
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg := &Config{
		PulseServer: "http://test:6532",
		NodeID:      "test-node",
		NodeName:    "Test Node",
	}

	if err := SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

// TestSaveConfig_EmptyPath tests SaveConfig with empty path
func TestSaveConfig_EmptyPath(t *testing.T) {
	cfg := &Config{
		PulseServer: "http://test:6532",
		NodeID:      "test-node",
		NodeName:    "Test Node",
	}

	if err := SaveConfig(cfg, ""); err == nil {
		t.Error("Expected error for empty path")
	}
}

// TestSaveConfig_WithRegion tests SaveConfig with region
func TestSaveConfig_WithRegion(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "saved.yaml")

	// Create initial file
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg := &Config{
		PulseServer: "http://test:6532",
		NodeID:      "test-node",
		NodeName:    "Test Node",
		Region:      "us-east-1",
	}

	if err := SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("SaveConfig with region failed: %v", err)
	}
}

// TestSaveConfig_WithTags tests SaveConfig with tags
func TestSaveConfig_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "saved.yaml")

	// Create initial file
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg := &Config{
		PulseServer: "http://test:6532",
		NodeID:      "test-node",
		NodeName:    "Test Node",
		Tags:        []string{"tag1", "tag2"},
	}

	if err := SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("SaveConfig with tags failed: %v", err)
	}
}

// TestSaveConfig_WithProbes tests SaveConfig with probes
func TestSaveConfig_WithProbes(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "saved.yaml")

	// Create initial file
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg := &Config{
		PulseServer: "http://test:6532",
		NodeID:      "test-node",
		NodeName:    "Test Node",
		Probes: []ProbeConfig{
			{
				Type:           "tcp_ping",
				Target:         "example.com",
				Port:           443,
				Interval:       60,
				TimeoutSeconds: 5,
				Count:          10,
			},
		},
	}

	if err := SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("SaveConfig with probes failed: %v", err)
	}
}

// TestSaveConfig_WithReconnect tests SaveConfig with reconnect config
func TestSaveConfig_WithReconnect(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "saved.yaml")

	// Create initial file
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg := &Config{
		PulseServer: "http://test:6532",
		NodeID:      "test-node",
		NodeName:    "Test Node",
		Reconnect: ReconnectConfig{
			MaxRetries:    3,
			RetryInterval: 5,
			Backoff:       "exponential",
		},
	}

	if err := SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("SaveConfig with reconnect failed: %v", err)
	}
}

// TestValidateMetricsConfig_Valid tests valid metrics config validation
func TestValidateMetricsConfig_Valid(t *testing.T) {
	if err := validateMetricsConfig(2112, 10); err != nil {
		t.Errorf("Expected valid metrics config, got: %v", err)
	}
}

// TestValidateMetricsConfig_InvalidPort tests invalid metrics port
func TestValidateMetricsConfig_InvalidPort(t *testing.T) {
	if err := validateMetricsConfig(0, 10); err == nil {
		t.Error("Expected error for invalid metrics port")
	}
}

// TestValidateMetricsConfig_InvalidUpdateInterval tests invalid update interval
func TestValidateMetricsConfig_InvalidUpdateInterval(t *testing.T) {
	if err := validateMetricsConfig(2112, 0); err == nil {
		t.Error("Expected error for invalid update interval 0")
	}
}

// TestValidateMetricsConfig_UpdateIntervalTooHigh tests update interval too high
func TestValidateMetricsConfig_UpdateIntervalTooHigh(t *testing.T) {
	if err := validateMetricsConfig(2112, 61); err == nil {
		t.Error("Expected error for update interval > 60")
	}
}

// TestValidateLogConfig_Valid tests valid log config
func TestValidateLogConfig_Valid(t *testing.T) {
	if err := validateLogConfig("INFO", "/tmp/beacon.log"); err != nil {
		t.Errorf("Expected valid log config, got: %v", err)
	}
}

// TestValidateLogConfig_InvalidLevel tests invalid log level
func TestValidateLogConfig_InvalidLevel(t *testing.T) {
	if err := validateLogConfig("INVALID", "/tmp/beacon.log"); err == nil {
		t.Error("Expected error for invalid log level")
	}
}

// TestValidateLogConfig_EmptyFile tests empty log file path
func TestValidateLogConfig_EmptyFile(t *testing.T) {
	if err := validateLogConfig("INFO", ""); err == nil {
		t.Error("Expected error for empty log file")
	}
}

// TestValidateLogConfig_WrongExtension tests log file with wrong extension
func TestValidateLogConfig_WrongExtension(t *testing.T) {
	if err := validateLogConfig("INFO", "/tmp/beacon.txt"); err == nil {
		t.Error("Expected error for wrong file extension")
	}
}

// TestConfig_Validate_Valid tests Config.Validate with a valid configuration
func TestConfig_Validate_Valid(t *testing.T) {
cfg := &Config{
PulseServer:          "http://localhost:6532",
NodeID:               "test-node",
NodeName:             "Test Node",
MetricsEnabled:       true,
MetricsPort:          2112,
MetricsUpdateSeconds: 10,
LogLevel:             "INFO",
LogFile:              "/tmp/beacon.log",
LogMaxSize:           10,
LogMaxAge:            7,
LogMaxBackups:        3,
}

if err := cfg.Validate(); err != nil {
t.Errorf("Expected valid config, got: %v", err)
}
}

// TestConfig_Validate_MissingPulseServer tests missing pulse_server
func TestConfig_Validate_MissingPulseServer(t *testing.T) {
cfg := &Config{
NodeID:   "test-node",
NodeName: "Test Node",
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for missing pulse_server")
}
}

// TestConfig_Validate_MissingNodeID tests missing node_id
func TestConfig_Validate_MissingNodeID(t *testing.T) {
cfg := &Config{
PulseServer: "http://localhost:6532",
NodeName:    "Test Node",
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for missing node_id")
}
}

// TestConfig_Validate_MissingNodeName tests missing node_name
func TestConfig_Validate_MissingNodeName(t *testing.T) {
cfg := &Config{
PulseServer: "http://localhost:6532",
NodeID:      "test-node",
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for missing node_name")
}
}

// TestConfig_Validate_InvalidURL tests invalid pulse_server URL
func TestConfig_Validate_InvalidURL(t *testing.T) {
cfg := &Config{
PulseServer: "not-a-valid-url",
NodeID:      "test-node",
NodeName:    "Test Node",
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for invalid URL")
}
}

// TestConfig_Validate_InvalidProbeConfig tests with invalid probe config
func TestConfig_Validate_InvalidProbeConfig(t *testing.T) {
cfg := &Config{
PulseServer: "http://localhost:6532",
NodeID:      "test-node",
NodeName:    "Test Node",
MetricsPort: 2112,
MetricsUpdateSeconds: 10,
LogLevel:    "INFO",
LogFile:     "/tmp/beacon.log",
Probes: []ProbeConfig{
{
Type:   "invalid_type",
Target: "localhost",
Port:   80,
},
},
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for invalid probe config")
}
}

// TestConfig_Validate_InvalidReconnectConfig tests with invalid reconnect config
func TestConfig_Validate_InvalidReconnectConfig(t *testing.T) {
cfg := &Config{
PulseServer: "http://localhost:6532",
NodeID:      "test-node",
NodeName:    "Test Node",
MetricsPort: 2112,
MetricsUpdateSeconds: 10,
LogLevel:    "INFO",
LogFile:     "/tmp/beacon.log",
Reconnect: ReconnectConfig{
MaxRetries:    101, // Invalid: > 100
RetryInterval: 5,
Backoff:       "exponential",
},
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for invalid reconnect config")
}
}

// TestConfig_Validate_InvalidMetricsConfig tests with invalid metrics config
func TestConfig_Validate_InvalidMetricsConfig(t *testing.T) {
cfg := &Config{
PulseServer: "http://localhost:6532",
NodeID:      "test-node",
NodeName:    "Test Node",
MetricsPort: 0, // Invalid
MetricsUpdateSeconds: 10,
LogLevel:    "INFO",
LogFile:     "/tmp/beacon.log",
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for invalid metrics port")
}
}

// TestConfig_Validate_InvalidLogConfig tests with invalid log config
func TestConfig_Validate_InvalidLogConfig(t *testing.T) {
cfg := &Config{
PulseServer: "http://localhost:6532",
NodeID:      "test-node",
NodeName:    "Test Node",
MetricsPort: 2112,
MetricsUpdateSeconds: 10,
	LogLevel:    "INVALID",  // Invalid level
	LogFile:     "/tmp/beacon.log",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for invalid log config")
	}
}

// TestLoadConfig_TelemetryEnvOverrides verifies BEACON_TELEMETRY_* env vars
// override the telemetry section (ADR-004 contract 2). Previously telemetry was
// file-only, which made env-only deployments unable to enable tracing.
func TestLoadConfig_TelemetryEnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	// Minimal valid config; telemetry intentionally absent from the file.
	configContent := `
pulse_server: "http://localhost:6532"
node_id: "tel-test"
node_name: "Telemetry Test"
telemetry:
  enabled: false
  service_name: "from-file"
  sampling_rate: 1.0
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Env overrides (these must win over the file values above).
	t.Setenv("BEACON_TELEMETRY_ENABLED", "true")
	t.Setenv("BEACON_TELEMETRY_SERVICE_NAME", "from-env")
	t.Setenv("BEACON_TELEMETRY_SERVICE_VERSION", "1.2.3")
	t.Setenv("BEACON_TELEMETRY_OTLP_ENDPOINT", "collector:4317")
	t.Setenv("BEACON_TELEMETRY_SAMPLING_RATE", "0.25")

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if !cfg.Telemetry.Enabled {
		t.Error("Expected telemetry.enabled to be overridden to true by env")
	}
	if cfg.Telemetry.ServiceName != "from-env" {
		t.Errorf("Expected telemetry.service_name 'from-env', got %q", cfg.Telemetry.ServiceName)
	}
	if cfg.Telemetry.ServiceVersion != "1.2.3" {
		t.Errorf("Expected telemetry.service_version '1.2.3', got %q", cfg.Telemetry.ServiceVersion)
	}
	if cfg.Telemetry.OTLPEndpoint != "collector:4317" {
		t.Errorf("Expected telemetry.otlp_endpoint 'collector:4317', got %q", cfg.Telemetry.OTLPEndpoint)
	}
	if cfg.Telemetry.SamplingRate != 0.25 {
		t.Errorf("Expected telemetry.sampling_rate 0.25, got %f", cfg.Telemetry.SamplingRate)
	}
}

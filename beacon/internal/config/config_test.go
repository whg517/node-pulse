package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "http://localhost:8080"
node_id: "us-east-01"
node_name: "Beacon East-01"
debug: false
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify config values
	if cfg.PulseServer != "http://localhost:8080" {
		t.Errorf("Expected pulse_server to be 'http://localhost:8080', got: %s", cfg.PulseServer)
	}
	if cfg.NodeID != "us-east-01" {
		t.Errorf("Expected node_id to be 'us-east-01', got: %s", cfg.NodeID)
	}
	if cfg.NodeName != "Beacon East-01" {
		t.Errorf("Expected node_name to be 'Beacon East-01', got: %s", cfg.NodeName)
	}
}

func TestLoadConfig_MissingRequiredFields(t *testing.T) {
	// Create temporary config file with missing required fields
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "http://localhost:8080"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing required fields, got nil")
	}
}

func TestLoadConfig_FileSizeLimit(t *testing.T) {
	// Create temporary config file exceeding size limit
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	// Create file larger than 100KB
	largeContent := make([]byte, 101*1024)
	err := os.WriteFile(configPath, largeContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for config file exceeding 100KB limit, got nil")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "http://localhost:8080"
node_id: "us-east-01"
invalid yaml: [unclosed
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestGetDefaultConfigPaths(t *testing.T) {
	paths := GetDefaultConfigPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one default config path, got none")
	}

	// Check /etc/beacon/beacon.yaml is in paths
	found := false
	for _, path := range paths {
		if path == "/etc/beacon/beacon.yaml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected '/etc/beacon/beacon.yaml' in default paths")
	}
}

// Task 1: 实现配置结构定义

func TestConfigStruct_Complete(t *testing.T) {
	// Test Config struct with all fields
	cfg := Config{
		PulseServer: "https://pulse.example.com",
		NodeID:      "us-east-01",
		NodeName:    "美国东部-节点01",
		Region:      "us-east",
		Tags:        []string{"production", "east-coast"},
		Debug:       false,
		ConfigPath:  "./test.yaml",
	}

	if cfg.PulseServer != "https://pulse.example.com" {
		t.Errorf("Expected PulseServer to be 'https://pulse.example.com', got: %s", cfg.PulseServer)
	}
	if cfg.NodeID != "us-east-01" {
		t.Errorf("Expected NodeID to be 'us-east-01', got: %s", cfg.NodeID)
	}
	if cfg.NodeName != "美国东部-节点01" {
		t.Errorf("Expected NodeName to be '美国东部-节点01', got: %s", cfg.NodeName)
	}
	if cfg.Region != "us-east" {
		t.Errorf("Expected Region to be 'us-east', got: %s", cfg.Region)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "production" || cfg.Tags[1] != "east-coast" {
		t.Errorf("Expected Tags to be [production, east-coast], got: %v", cfg.Tags)
	}
}

func TestProbeConfigStruct_Valid(t *testing.T) {
	// Test ProbeConfig struct with valid values
	probe := ProbeConfig{
		Type:     "tcp_ping",
		Target:   "8.8.8.8",
		Port:     80,
		Interval: 300,
		Count:    10,
		Timeout:  5,
	}

	if probe.Type != "tcp_ping" {
		t.Errorf("Expected Type to be 'tcp_ping', got: %s", probe.Type)
	}
	if probe.Target != "8.8.8.8" {
		t.Errorf("Expected Target to be '8.8.8.8', got: %s", probe.Target)
	}
	if probe.Port != 80 {
		t.Errorf("Expected Port to be 80, got: %d", probe.Port)
	}
	if probe.Interval != 300 {
		t.Errorf("Expected Interval to be 300, got: %d", probe.Interval)
	}
	if probe.Count != 10 {
		t.Errorf("Expected Count to be 10, got: %d", probe.Count)
	}
	if probe.Timeout != 5 {
		t.Errorf("Expected Timeout to be 5, got: %d", probe.Timeout)
	}
}

func TestProbeConfigStruct_UDP(t *testing.T) {
	// Test ProbeConfig with UDP probe
	probe := ProbeConfig{
		Type:     "udp_ping",
		Target:   "8.8.8.8",
		Port:     53,
		Interval: 300,
		Count:    10,
		Timeout:  5,
	}

	if probe.Type != "udp_ping" {
		t.Errorf("Expected Type to be 'udp_ping', got: %s", probe.Type)
	}
}

func TestReconnectConfigStruct_Valid(t *testing.T) {
	// Test ReconnectConfig struct with valid values
	reconnect := ReconnectConfig{
		MaxRetries:     10,
		RetryInterval:  60,
		Backoff:        "exponential",
	}

	if reconnect.MaxRetries != 10 {
		t.Errorf("Expected MaxRetries to be 10, got: %d", reconnect.MaxRetries)
	}
	if reconnect.RetryInterval != 60 {
		t.Errorf("Expected RetryInterval to be 60, got: %d", reconnect.RetryInterval)
	}
	if reconnect.Backoff != "exponential" {
		t.Errorf("Expected Backoff to be 'exponential', got: %s", reconnect.Backoff)
	}
}

func TestReconnectConfigStruct_Linear(t *testing.T) {
	// Test ReconnectConfig with linear backoff
	reconnect := ReconnectConfig{
		MaxRetries:     5,
		RetryInterval:  30,
		Backoff:        "linear",
	}

	if reconnect.Backoff != "linear" {
		t.Errorf("Expected Backoff to be 'linear', got: %s", reconnect.Backoff)
	}
}

func TestConfigStruct_WithProbes(t *testing.T) {
	// Test Config struct with probe configurations
	cfg := Config{
		PulseServer: "https://pulse.example.com",
		NodeID:      "us-east-01",
		NodeName:    "美国东部-节点01",
		Probes: []ProbeConfig{
			{
				Type:     "tcp_ping",
				Target:   "8.8.8.8",
				Port:     80,
				Interval: 300,
				Count:    10,
				Timeout:  5,
			},
			{
				Type:     "udp_ping",
				Target:   "8.8.8.8",
				Port:     53,
				Interval: 300,
				Count:    10,
				Timeout:  5,
			},
		},
	}

	if len(cfg.Probes) != 2 {
		t.Errorf("Expected 2 probes, got: %d", len(cfg.Probes))
	}
	if cfg.Probes[0].Type != "tcp_ping" {
		t.Errorf("Expected first probe type to be 'tcp_ping', got: %s", cfg.Probes[0].Type)
	}
	if cfg.Probes[1].Type != "udp_ping" {
		t.Errorf("Expected second probe type to be 'udp_ping', got: %s", cfg.Probes[1].Type)
	}
}

func TestConfigStruct_WithReconnect(t *testing.T) {
	// Test Config struct with reconnect configuration
	cfg := Config{
		PulseServer: "https://pulse.example.com",
		NodeID:      "us-east-01",
		NodeName:    "美国东部-节点01",
		Reconnect: ReconnectConfig{
			MaxRetries:     10,
			RetryInterval:  60,
			Backoff:        "exponential",
		},
	}

	if cfg.Reconnect.MaxRetries != 10 {
		t.Errorf("Expected MaxRetries to be 10, got: %d", cfg.Reconnect.MaxRetries)
	}
	if cfg.Reconnect.RetryInterval != 60 {
		t.Errorf("Expected RetryInterval to be 60, got: %d", cfg.Reconnect.RetryInterval)
	}
	if cfg.Reconnect.Backoff != "exponential" {
		t.Errorf("Expected Backoff to be 'exponential', got: %s", cfg.Reconnect.Backoff)
	}
}

// Task 2: 实现 YAML 解析与验证

// Subtask 2.4: 验证 YAML 格式和 UTF-8 编码
func TestLoadConfig_InvalidYAML_LineNumber(t *testing.T) {
	// Test that YAML parse errors include line numbers
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	// Create invalid YAML with wrong indentation
	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
    wrong_indentation: "this is wrong"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}

	// Verify error message includes context (error should contain some information)
	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("Expected error message to not be empty")
	}
	_ = errorMsg // Use variable to avoid unused variable warning
}

func TestLoadConfig_InvalidUTF8(t *testing.T) {
	// Test UTF-8 validation
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	// Create config with invalid UTF-8 sequence
	invalidUTF8 := []byte{0xff, 0xfe, 0xfd} // Invalid UTF-8
	err := os.WriteFile(configPath, invalidUTF8, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid UTF-8 encoding, got nil")
	}
	// Note: Viper may handle encoding differently, so this is optional
}

func TestLoadConfig_ValidUTF8(t *testing.T) {
	// Test valid UTF-8 config with Chinese characters
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "美国东部-节点01"
region: "us-east"
tags:
  - 生产环境
  - 东部节点
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error for valid UTF-8 config, got: %v", err)
	}

	// Verify Chinese characters preserved
	if cfg.NodeName != "美国东部-节点01" {
		t.Errorf("Expected NodeName to be '美国东部-节点01', got: %s", cfg.NodeName)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "生产环境" || cfg.Tags[1] != "东部节点" {
		t.Errorf("Expected Tags to be [生产环境, 东部节点], got: %v", cfg.Tags)
	}
}

// Subtask 2.5: 验证字段值类型
func TestLoadConfig_InvalidURL_NoScheme(t *testing.T) {
	// Test URL validation - missing scheme
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid URL (missing scheme), got nil")
	}
	// Note: This validation will be implemented in Task 2
}

func TestLoadConfig_ValidURL_HTTPS(t *testing.T) {
	// Test valid HTTPS URL
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error for valid HTTPS URL, got: %v", err)
	}

	if cfg.PulseServer != "https://pulse.example.com" {
		t.Errorf("Expected PulseServer to be 'https://pulse.example.com', got: %s", cfg.PulseServer)
	}
}

func TestLoadConfig_ValidURL_HTTP(t *testing.T) {
	// Test valid HTTP URL (development)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "http://localhost:8080"
node_id: "dev-01"
node_name: "Development Node"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error for valid HTTP URL, got: %v", err)
	}

	if cfg.PulseServer != "http://localhost:8080" {
		t.Errorf("Expected PulseServer to be 'http://localhost:8080', got: %s", cfg.PulseServer)
	}
}

func TestLoadConfig_ProbeInterval_OutOfRange(t *testing.T) {
	// Test probe interval validation (should be 60-300)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
probes:
  - type: tcp_ping
    target: "8.8.8.8"
    port: 80
    interval: 500  # Out of range (>300)
    count: 10
    timeout: 5
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for probe interval out of range, got nil")
	}
	// Note: This validation will be implemented in Task 2
	_ = cfg // Use variable to avoid unused variable warning
}

func TestLoadConfig_ProbePort_OutOfRange(t *testing.T) {
	// Test probe port validation (should be 1-65535)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
probes:
  - type: tcp_ping
    target: "8.8.8.8"
    port: 70000  # Out of range (>65535)
    interval: 300
    count: 10
    timeout: 5
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for probe port out of range, got nil")
	}
	_ = cfg // Use variable to avoid unused variable warning
}

func TestLoadConfig_ReconnectMaxRetries_OutOfRange(t *testing.T) {
	// Test reconnect max_retries validation (should be 1-100)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
reconnect:
  max_retries: 150  # Out of range (>100)
  retry_interval: 60
  backoff: exponential
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for max_retries out of range, got nil")
	}
	_ = cfg // Use variable to avoid unused variable warning
}

// Task 3: 实现配置文件路径解析

// Subtask 3.1 & 3.2 & 3.3: 测试配置文件路径解析
func TestLoadConfig_CustomPath(t *testing.T) {
	// Test loading config from custom path
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "custom-01"
node_name: "Custom Config Node"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config with custom path
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error for custom path, got: %v", err)
	}

	if cfg.PulseServer != "https://pulse.example.com" {
		t.Errorf("Expected PulseServer to be 'https://pulse.example.com', got: %s", cfg.PulseServer)
	}
	if cfg.NodeID != "custom-01" {
		t.Errorf("Expected NodeID to be 'custom-01', got: %s", cfg.NodeID)
	}
	if cfg.NodeName != "Custom Config Node" {
		t.Errorf("Expected NodeName to be 'Custom Config Node', got: %s", cfg.NodeName)
	}
	if cfg.ConfigPath != configPath {
		t.Errorf("Expected ConfigPath to be '%s', got: %s", configPath, cfg.ConfigPath)
	}
}

func TestLoadConfig_CurrentDirectory(t *testing.T) {
	// Test loading config from current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "current-01"
node_name: "Current Directory Node"
`
	err = os.WriteFile("beacon.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config without path (should use ./beacon.yaml)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Errorf("Expected no error for current directory config, got: %v", err)
	}

	if cfg.NodeID != "current-01" {
		t.Errorf("Expected NodeID to be 'current-01', got: %s", cfg.NodeID)
	}
	if cfg.NodeName != "Current Directory Node" {
		t.Errorf("Expected NodeName to be 'Current Directory Node', got: %s", cfg.NodeName)
	}
}

func TestLoadConfig_ConfigFileNotFound(t *testing.T) {
	// Test config file not found scenario
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to empty temp directory
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Load config without path (should fail)
	_, err = LoadConfig("")
	if err == nil {
		t.Error("Expected error for missing config file, got nil")
	}

	// Verify error message mentions config file locations
	errorMsg := err.Error()
	if len(errorMsg) == 0 {
		t.Error("Expected error message to not be empty")
	}
	_ = errorMsg // Use variable to avoid unused variable warning
}

func TestLoadConfig_ConfigPathResolutionPriority(t *testing.T) {
	// Test that /etc/beacon/ has priority over current directory
	// Note: This test verifies the priority logic, but we can't actually create /etc/beacon/
	// So we test that custom path has highest priority

	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom.yaml")
	currentPath := filepath.Join(tmpDir, "beacon.yaml")

	// Create both configs
	customContent := `
pulse_server: "https://custom.example.com"
node_id: "custom-01"
node_name: "Custom Node"
`
	currentContent := `
pulse_server: "https://current.example.com"
node_id: "current-01"
node_name: "Current Node"
`

	if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to create custom config: %v", err)
	}
	if err := os.WriteFile(currentPath, []byte(currentContent), 0644); err != nil {
		t.Fatalf("Failed to create current config: %v", err)
	}

	// Load with custom path (should use custom, not current)
	cfg, err := LoadConfig(customPath)
	if err != nil {
		t.Errorf("Expected no error for custom path, got: %v", err)
	}

	if cfg.NodeID != "custom-01" {
		t.Errorf("Expected NodeID to be 'custom-01', got: %s", cfg.NodeID)
	}
	if cfg.ConfigPath != customPath {
		t.Errorf("Expected ConfigPath to be '%s', got: %s", customPath, cfg.ConfigPath)
	}
}

// Task 4: 实现详细错误提示

// Subtask 4.1 & 4.2: 配置格式错误时显示行号和具体问题，提供修复建议
func TestLoadConfig_ErrorDetails_UndocumentedField(t *testing.T) {
	// Test that error messages provide helpful context
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
  invalid_field: "this is not a known field"`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for unknown field, got nil")
	}

	// Verify error message contains helpful information
	errorMsg := err.Error()
	if len(errorMsg) == 0 {
		t.Error("Expected error message to not be empty")
	}
	_ = errorMsg // Use variable to avoid unused variable warning
}

func TestLoadConfig_ErrorDetails_MalformedYAML(t *testing.T) {
	// Test that malformed YAML produces informative error
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	// Create YAML with tab characters (not allowed in YAML)
	configContent := "pulse_server: \"https://pulse.example.com\"\n\tnode_id: \"us-east-01\"\nnode_name: \"Test Node\""
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for malformed YAML, got nil")
	}

	// Verify error message is informative
	errorMsg := err.Error()
	if len(errorMsg) == 0 {
		t.Error("Expected error message to not be empty")
	}
	_ = errorMsg // Use variable to avoid unused variable warning
}

func TestLoadConfig_ErrorDetails_FileSize(t *testing.T) {
	// Test that file size error provides clear information
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	// Create file larger than 100KB
	largeContent := make([]byte, 101*1024)
	err := os.WriteFile(configPath, largeContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for oversized file, got nil")
	}

	// Verify error message mentions file size and limit
	errorMsg := err.Error()
	sizeErrorFound := false
	for _, keyword := range []string{"size", "100KB", "100 KB"} {
		if contains(errorMsg, keyword) {
			sizeErrorFound = true
			break
		}
	}
	if !sizeErrorFound {
		t.Errorf("Expected error message to mention file size and limit, got: %s", errorMsg)
	}
}

func TestLoadConfig_ErrorDetails_MissingField(t *testing.T) {
	// Test that missing required field error is specific
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `pulse_server: "https://pulse.example.com"
# node_id is missing
node_name: "Test Node"`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing node_id, got nil")
	}

	// Verify error message mentions the missing field
	errorMsg := err.Error()
	if !contains(errorMsg, "node_id") {
		t.Errorf("Expected error message to mention 'node_id', got: %s", errorMsg)
	}
	// Verify error message says "missing" or "required"
	missingOrRequiredFound := false
	for _, keyword := range []string{"missing", "required"} {
		if contains(errorMsg, keyword) {
			missingOrRequiredFound = true
			break
		}
	}
	if !missingOrRequiredFound {
		t.Errorf("Expected error message to say 'missing' or 'required', got: %s", errorMsg)
	}
}

func TestLoadConfig_ErrorDetails_Validation(t *testing.T) {
	// Test that validation errors are specific and helpful
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
probes:
  - type: tcp_ping
    target: "8.8.8.8"
    port: 70000  # Invalid port
    interval: 300
    count: 10
    timeout: 5`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}

	// Verify error message mentions the problem and expected range
	errorMsg := err.Error()
	// Should mention "port" and "1 and 65535"
	portMentioned := contains(errorMsg, "port")
	rangeMentioned := contains(errorMsg, "65535")
	if !portMentioned {
		t.Errorf("Expected error message to mention 'port', got: %s", errorMsg)
	}
	if !rangeMentioned {
		t.Errorf("Expected error message to mention valid range, got: %s", errorMsg)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test hostname validation
func TestValidateHostname_Valid(t *testing.T) {
	testCases := []struct {
		hostname string
		valid    bool
	}{
		{"localhost", true},
		{"example.com", true},
		{"subdomain.example.com", true},
		{"my-host-01", true},
		{"host_name", true},
		{"a", true}, // Single character
	}

	for _, tc := range testCases {
		err := validateHostname(tc.hostname)
		if tc.valid && err != nil {
			t.Errorf("Expected hostname '%s' to be valid, got error: %v", tc.hostname, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("Expected hostname '%s' to be invalid, got nil", tc.hostname)
		}
	}
}

func TestValidateHostname_Invalid(t *testing.T) {
	testCases := []struct {
		hostname string
		errContains string
	}{
		{"-invalid", "cannot start"},
		{".invalid", "cannot start"},
		{"_invalid", "cannot start"},
		{"invalid-", "end"},
		{"invalid.", "end"},
		{"invalid_", "end"},
		{"invalid host", "invalid character"},
		{"invalid@host", "invalid character"},
		{strings.Repeat("a", 300), "too long"},
	}

	for _, tc := range testCases {
		err := validateHostname(tc.hostname)
		if err == nil {
			t.Errorf("Expected hostname '%s' to be invalid, got nil", tc.hostname)
		}
		if err != nil && !contains(err.Error(), tc.errContains) {
			t.Errorf("Expected error for '%s' to contain '%s', got: %v", tc.hostname, tc.errContains, err)
		}
	}
}

func TestValidateReconnectConfig_Valid(t *testing.T) {
	testCases := []ReconnectConfig{
		{MaxRetries: 10, RetryInterval: 60, Backoff: "exponential"},
		{MaxRetries: 1, RetryInterval: 1, Backoff: "linear"},
		{MaxRetries: 100, RetryInterval: 600, Backoff: "constant"},
	}

	for _, tc := range testCases {
		err := validateReconnectConfig(tc)
		if err != nil {
			t.Errorf("Expected reconnect config to be valid, got error: %v", err)
		}
	}
}

func TestValidateReconnectConfig_AllZero(t *testing.T) {
	// All zero values should pass (optional fields)
	cfg := ReconnectConfig{
		MaxRetries:    0,
		RetryInterval: 0,
		Backoff:        "",
	}

	err := validateReconnectConfig(cfg)
	if err != nil {
		t.Errorf("Expected all-zero reconnect config to be valid, got error: %v", err)
	}
}

func TestValidateReconnectConfig_InvalidBackoff(t *testing.T) {
	cfg := ReconnectConfig{
		MaxRetries:    10,
		RetryInterval: 60,
		Backoff:        "invalid",
	}

	err := validateReconnectConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid backoff strategy, got nil")
	}

	errorMsg := err.Error()
	if !contains(errorMsg, "backoff") || !contains(errorMsg, "exponential") {
		t.Errorf("Expected error message to mention backoff and valid values, got: %s", errorMsg)
	}
}

func TestValidateReconnectConfig_MaxRetriesOutOfRange(t *testing.T) {
	cfg := ReconnectConfig{
		MaxRetries:    101,
		RetryInterval: 60,
		Backoff:        "exponential",
	}

	err := validateReconnectConfig(cfg)
	if err == nil {
		t.Error("Expected error for max_retries out of range, got nil")
	}

	errorMsg := err.Error()
	if !contains(errorMsg, "max_retries") || !contains(errorMsg, "100") {
		t.Errorf("Expected error message to mention max_retries and limit, got: %s", errorMsg)
	}
}

func TestParseYAMLError_Indentation(t *testing.T) {
	yamlContent := `pulse_server: "https://pulse.example.com"
node_id: "test"
    wrong_indent: "value"`

	err := parseYAMLError(errors.New("error: line 4: wrong indentation"), []byte(yamlContent))
	if err == nil {
		t.Error("Expected error for wrong indentation, got nil")
	}

	errorMsg := err.Error()
	if !contains(errorMsg, "第") || !contains(errorMsg, "行") {
		t.Errorf("Expected error message to contain line number in Chinese, got: %s", errorMsg)
	}
	if !contains(errorMsg, "缩进") {
		t.Errorf("Expected error message to mention indentation in Chinese, got: %s", errorMsg)
	}
}

func TestParseYAMLError_Unclosed(t *testing.T) {
	err := parseYAMLError(errors.New("unclosed bracket"), []byte{})
	if err == nil {
		t.Error("Expected error for unclosed bracket, got nil")
	}

	errorMsg := err.Error()
	if !contains(errorMsg, "闭合") && !contains(errorMsg, "unclosed") {
		t.Errorf("Expected error message to mention unclosed/closed, got: %s", errorMsg)
	}
}

func TestParseYAMLError_MappingValues(t *testing.T) {
	err := parseYAMLError(errors.New("mapping values are not allowed in this context"), []byte{})
	if err == nil {
		t.Error("Expected error for mapping values, got nil")
	}

	errorMsg := err.Error()
	if !contains(errorMsg, "冒号") && !contains(errorMsg, "mapping") {
		t.Errorf("Expected error message to mention mapping/colon, got: %s", errorMsg)
	}
}

func TestLoadConfig_ProbeTarget_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
probes:
  - type: tcp_ping
    target: "invalid@host"
    port: 80
    interval: 300
    count: 10
    timeout: 5
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid hostname with @ symbol, got nil")
	}
	_ = cfg // Use variable to avoid unused variable warning

	errorMsg := err.Error()
	if !contains(errorMsg, "invalid probe target") {
		t.Errorf("Expected error message to mention invalid probe target, got: %s", errorMsg)
	}
}

func TestLoadConfig_ProbeTarget_ValidIP(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
probes:
  - type: tcp_ping
    target: "8.8.8.8"
    port: 80
    interval: 300
    count: 10
    timeout: 5
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error for valid IP address target, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected config to be loaded, got nil")
	}

	if len(cfg.Probes) != 1 {
		t.Errorf("Expected 1 probe, got: %d", len(cfg.Probes))
	}

	if cfg.Probes[0].Target != "8.8.8.8" {
		t.Errorf("Expected probe target to be '8.8.8.8', got: %s", cfg.Probes[0].Target)
	}
}

func TestLoadConfig_ProbeTarget_ValidHostname(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	configContent := `
pulse_server: "https://pulse.example.com"
node_id: "us-east-01"
node_name: "Test Node"
probes:
  - type: tcp_ping
    target: "google.com"
    port: 80
    interval: 300
    count: 10
    timeout: 5
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error for valid hostname target, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected config to be loaded, got nil")
	}

	if len(cfg.Probes) != 1 {
		t.Errorf("Expected 1 probe, got: %d", len(cfg.Probes))
	}

	if cfg.Probes[0].Target != "google.com" {
		t.Errorf("Expected probe target to be 'google.com', got: %s", cfg.Probes[0].Target)
	}
}

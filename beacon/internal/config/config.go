package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"unicode/utf8"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the complete Beacon configuration
type Config struct {
	// Required fields
	PulseServer string `mapstructure:"pulse_server" yaml:"pulse_server"`
	NodeID      string `mapstructure:"node_id" yaml:"node_id"`
	NodeName    string `mapstructure:"node_name" yaml:"node_name"`

	// Optional fields
	Region string   `mapstructure:"region" yaml:"region"`
	Tags   []string `mapstructure:"tags" yaml:"tags"`

	// Probe configuration (for Story 3.3)
	Probes []ProbeConfig `mapstructure:"probes" yaml:"probes"`

	// Reconnect configuration (for Story 2.6)
	Reconnect ReconnectConfig `mapstructure:"reconnect" yaml:"reconnect"`

	// Metrics configuration (for Story 3.8)
	MetricsEnabled       bool `mapstructure:"metrics_enabled" yaml:"metrics_enabled"`
	MetricsPort          int  `mapstructure:"metrics_port" yaml:"metrics_port"`
	MetricsUpdateSeconds int  `mapstructure:"metrics_update_seconds" yaml:"metrics_update_seconds"`

	// Logging configuration (for Story 3.9)
	LogLevel      string `mapstructure:"log_level" yaml:"log_level"`                          // DEBUG, INFO, WARN, ERROR
	LogFile       string `mapstructure:"log_file" yaml:"log_file"`                            // /var/log/beacon/beacon.log
	LogMaxSize    int    `mapstructure:"log_max_size" yaml:"log_max_size"`                    // MB
	LogMaxAge     int    `mapstructure:"log_max_age" yaml:"log_max_age"`                      // days
	LogMaxBackups int    `mapstructure:"log_max_backups" yaml:"log_max_backups"`              // number of backups
	LogCompress   bool   `mapstructure:"log_compress" yaml:"log_compress"`                    // compress rotated files
	LogToConsole  bool   `mapstructure:"log_to_console" yaml:"log_to_console"`                // also log to stdout

	// Debug mode configuration (for Story 3.10)
	DebugMode bool `mapstructure:"debug_mode" yaml:"debug_mode"` // Enable debug mode (auto-sets log_level=DEBUG)

	// Resource monitor configuration (for Story 3.11)
	ResourceMonitor ResourceMonitorConfig `mapstructure:"resource_monitor" yaml:"resource_monitor"`

	// Internal fields (not from config file)
	ConfigPath string `mapstructure:"-"`
	Debug      bool   `mapstructure:"debug"`
}

// ProbeConfig represents a single probe configuration
type ProbeConfig struct {
	Type           string `mapstructure:"type" yaml:"type"`
	Target         string `mapstructure:"target" yaml:"target"`
	Port           int    `mapstructure:"port" yaml:"port"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds" yaml:"timeout_seconds"`
	Interval       int    `mapstructure:"interval" yaml:"interval"`
	Count          int    `mapstructure:"count" yaml:"count"`
}

// ReconnectConfig represents connection retry configuration
type ReconnectConfig struct {
	MaxRetries    int    `mapstructure:"max_retries" yaml:"max_retries"`
	RetryInterval int    `mapstructure:"retry_interval" yaml:"retry_interval"`
	Backoff        string `mapstructure:"backoff" yaml:"backoff"`
}

// LoadConfig loads configuration from file with validation
func LoadConfig(configPath string) (*Config, error) {
	// Resolve config file path
	resolvedPath, err := resolveConfigPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check file size (≤100KB)
	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}
	if fileInfo.Size() > 100*1024 {
		return nil, fmt.Errorf("config file size %d exceeds limit of 100KB", fileInfo.Size())
	}

	// Read and validate YAML encoding
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(data) {
		return nil, errors.New("config file contains invalid UTF-8 encoding")
	}

	// Parse YAML with Viper
	v := viper.New()
	v.SetConfigFile(resolvedPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		// Extract line number from YAML parse error for UX-friendly messages
		return nil, parseYAMLError(err, data)
	}

	// Unmarshal to Config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if config.PulseServer == "" {
		return nil, errors.New("required field 'pulse_server' is missing (suggestion: add pulse_server: \"https://pulse.example.com\" to config)")
	}
	if config.NodeID == "" {
		return nil, errors.New("required field 'node_id' is missing (suggestion: add node_id: \"your-node-id\" to config)")
	}
	if config.NodeName == "" {
		return nil, errors.New("required field 'node_name' is missing (suggestion: add node_name: \"Your Node Name\" to config)")
	}

	// Validate URL format
	if _, err := url.ParseRequestURI(config.PulseServer); err != nil {
		return nil, fmt.Errorf("invalid pulse_server URL: %w (suggestion: ensure URL includes scheme like https:// or http://)", err)
	}

	// Validate probe configurations if present
	for i, probe := range config.Probes {
		if err := validateProbeConfig(probe); err != nil {
			return nil, fmt.Errorf("probe %d validation failed: %w", i+1, err)
		}
	}

	// Validate reconnect configuration if present
	if err := validateReconnectConfig(config.Reconnect); err != nil {
		return nil, fmt.Errorf("reconnect configuration validation failed: %w", err)
	}

	// Set default values for metrics configuration (Story 3.8)
	// If metrics_port is not set, use default port and enable metrics
	if config.MetricsPort == 0 {
		config.MetricsPort = 2112 // Default Prometheus port
	}
	// If metrics_enabled is not explicitly set (false), default to true
	if !config.MetricsEnabled && config.MetricsPort != 0 {
		config.MetricsEnabled = true // Default to enabled
	}
	// Fix #4: Set default metrics update interval (10-60 seconds range)
	if config.MetricsUpdateSeconds == 0 {
		config.MetricsUpdateSeconds = 10 // Default 10 seconds
	}

	// Set default values for logging configuration (Story 3.9)
	if config.LogLevel == "" {
		config.LogLevel = "INFO" // Default log level
	}
	if config.LogFile == "" {
		config.LogFile = "/var/log/beacon/beacon.log" // Default log file path
	}
	if config.LogMaxSize == 0 {
		config.LogMaxSize = 10 // Default 10 MB
	}
	if config.LogMaxAge == 0 {
		config.LogMaxAge = 7 // Default 7 days
	}
	if config.LogMaxBackups == 0 {
		config.LogMaxBackups = 10 // Default 10 backups
	}
	// LogCompress and LogToConsole default to false (bool default)

	// Apply debug mode configuration (Story 3.10)
	// When debug_mode is true, automatically set log level to DEBUG
	if config.DebugMode {
		config.LogLevel = "DEBUG"
	}

	// Set default values for resource monitor configuration (Story 3.11)
	if config.ResourceMonitor.CheckIntervalSeconds == 0 {
		config.ResourceMonitor.CheckIntervalSeconds = 60 // Default 60 seconds
	}
	if config.ResourceMonitor.Thresholds.CPUMicrocores == 0 {
		config.ResourceMonitor.Thresholds.CPUMicrocores = 100 // Default 100 microcores
	}
	if config.ResourceMonitor.Thresholds.MemoryMB == 0 {
		config.ResourceMonitor.Thresholds.MemoryMB = 100 // Default 100 MB
	}
	if config.ResourceMonitor.Degradation.DegradedLevel.CPUMicrocores == 0 {
		config.ResourceMonitor.Degradation.DegradedLevel.CPUMicrocores = 200 // Default 200 microcores
	}
	if config.ResourceMonitor.Degradation.DegradedLevel.MemoryMB == 0 {
		config.ResourceMonitor.Degradation.DegradedLevel.MemoryMB = 150 // Default 150 MB
	}
	if config.ResourceMonitor.Degradation.DegradedLevel.IntervalMultiplier == 0 {
		config.ResourceMonitor.Degradation.DegradedLevel.IntervalMultiplier = 2 // Default 2x
	}
	if config.ResourceMonitor.Degradation.CriticalLevel.CPUMicrocores == 0 {
		config.ResourceMonitor.Degradation.CriticalLevel.CPUMicrocores = 300 // Default 300 microcores
	}
	if config.ResourceMonitor.Degradation.CriticalLevel.MemoryMB == 0 {
		config.ResourceMonitor.Degradation.CriticalLevel.MemoryMB = 200 // Default 200 MB
	}
	if config.ResourceMonitor.Degradation.CriticalLevel.IntervalMultiplier == 0 {
		config.ResourceMonitor.Degradation.CriticalLevel.IntervalMultiplier = 3 // Default 3x
	}
	if config.ResourceMonitor.Degradation.Recovery.ConsecutiveNormalChecks == 0 {
		config.ResourceMonitor.Degradation.Recovery.ConsecutiveNormalChecks = 3 // Default 3 checks
	}
	if config.ResourceMonitor.Alerting.SuppressionWindowSeconds == 0 {
		config.ResourceMonitor.Alerting.SuppressionWindowSeconds = 300 // Default 5 minutes
	}

	// Validate metrics configuration
	if err := validateMetricsConfig(config.MetricsPort, config.MetricsUpdateSeconds); err != nil {
		return nil, fmt.Errorf("metrics configuration validation failed: %w", err)
	}

	// Validate logging configuration
	if err := validateLogConfig(config.LogLevel, config.LogFile); err != nil {
		return nil, fmt.Errorf("logging configuration validation failed: %w", err)
	}

	config.ConfigPath = resolvedPath
	return &config, nil
}

// resolveConfigPath resolves config file path with fallback
func resolveConfigPath(customPath string) (string, error) {
	if customPath != "" {
		return customPath, nil
	}

	// Check /etc/beacon/beacon.yaml first
	etcPath := "/etc/beacon/beacon.yaml"
	if _, err := os.Stat(etcPath); err == nil {
		return etcPath, nil
	}

	// Fallback to current directory
	currentPath := "./beacon.yaml"
	if _, err := os.Stat(currentPath); err == nil {
		return currentPath, nil
	}

	return "", errors.New("config file not found (checked /etc/beacon/beacon.yaml and ./beacon.yaml)")
}

// validateProbeConfig validates probe configuration
func validateProbeConfig(probe ProbeConfig) error {
	// Validate type
	if probe.Type != "tcp_ping" && probe.Type != "udp_ping" {
		return fmt.Errorf("invalid probe type '%s', must be 'tcp_ping' or 'udp_ping'", probe.Type)
	}

	// Validate target (IP address or hostname)
	if probe.Target == "" {
		return fmt.Errorf("probe target cannot be empty")
	}
	if net.ParseIP(probe.Target) == nil {
		// Not an IP address, check if it's a valid hostname
		if err := validateHostname(probe.Target); err != nil {
			return fmt.Errorf("invalid probe target '%s': %w", probe.Target, err)
		}
	}

	// Validate port range (1-65535)
	if probe.Port < 1 || probe.Port > 65535 {
		return fmt.Errorf("invalid port %d, must be between 1 and 65535 (suggestion: check port number is valid)", probe.Port)
	}

	// Validate interval range (60-300)
	if probe.Interval < 60 || probe.Interval > 300 {
		return fmt.Errorf("invalid interval %d, must be between 60 and 300 seconds (suggestion: adjust interval to be within range)", probe.Interval)
	}

	// Validate count range (1-100)
	if probe.Count < 1 || probe.Count > 100 {
		return fmt.Errorf("invalid count %d, must be between 1 and 100 (suggestion: adjust probe count to be within range)", probe.Count)
	}

	// Validate timeout range (1-30)
	if probe.TimeoutSeconds < 1 || probe.TimeoutSeconds > 30 {
		return fmt.Errorf("invalid timeout %d, must be between 1 and 30 seconds (suggestion: adjust timeout to be within range)", probe.TimeoutSeconds)
	}

	return nil
}

// validateHostname validates hostname format
func validateHostname(hostname string) error {
	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long (max 253 characters)")
	}

	// Check for invalid characters
	for _, char := range hostname {
		if !isValidHostnameChar(char) {
			return fmt.Errorf("hostname contains invalid character '%c' (suggestion: use only letters, numbers, hyphens, underscores, and dots)", char)
		}
	}

	// Check that hostname doesn't start or end with hyphen/dot/underscore
	if strings.HasPrefix(hostname, "-") || strings.HasPrefix(hostname, ".") || strings.HasPrefix(hostname, "_") ||
		strings.HasSuffix(hostname, "-") || strings.HasSuffix(hostname, ".") || strings.HasSuffix(hostname, "_") {
		return fmt.Errorf("hostname cannot start or end with hyphen, dot, or underscore")
	}

	return nil
}

// isValidHostnameChar checks if a character is valid in hostname
func isValidHostnameChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '.' || r == '_'
}

// validateReconnectConfig validates reconnect configuration
func validateReconnectConfig(reconnect ReconnectConfig) error {
	// Only validate if fields are set (zero values are OK for optional fields)
	if reconnect.MaxRetries == 0 && reconnect.RetryInterval == 0 && reconnect.Backoff == "" {
		return nil // All fields unset, validation passes
	}

	// Validate max_retries range (1-100)
	if reconnect.MaxRetries != 0 && (reconnect.MaxRetries < 1 || reconnect.MaxRetries > 100) {
		return fmt.Errorf("invalid max_retries %d, must be between 1 and 100", reconnect.MaxRetries)
	}

	// Validate retry_interval range (1-600)
	if reconnect.RetryInterval != 0 && (reconnect.RetryInterval < 1 || reconnect.RetryInterval > 600) {
		return fmt.Errorf("invalid retry_interval %d, must be between 1 and 600", reconnect.RetryInterval)
	}

	// Validate backoff type
	if reconnect.Backoff != "" && reconnect.Backoff != "exponential" && reconnect.Backoff != "linear" && reconnect.Backoff != "constant" {
		return fmt.Errorf("invalid backoff '%s', must be 'exponential', 'linear', or 'constant'", reconnect.Backoff)
	}

	return nil
}

// validateMetricsConfig validates metrics configuration (Story 3.8)
func validateMetricsConfig(port int, updateSeconds int) error {
	// Validate metrics port range (1024-65535)
	// Avoid system ports (< 1024) for security
	if port < 1024 || port > 65535 {
		return fmt.Errorf("invalid metrics_port %d, must be between 1024 and 65535 (suggestion: use default port 2112 or choose an available port)", port)
	}

	// Fix #4: Validate metrics update interval (10-60 seconds)
	if updateSeconds < 10 || updateSeconds > 60 {
		return fmt.Errorf("invalid metrics_update_seconds %d, must be between 10 and 60 seconds", updateSeconds)
	}

	return nil
}

// validateLogConfig validates logging configuration (Story 3.9)
func validateLogConfig(logLevel string, logFile string) error {
	// Validate log level
	validLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLevels[logLevel] {
		return fmt.Errorf("invalid log level: %s (must be DEBUG, INFO, WARN, or ERROR)", logLevel)
	}

	// Validate log file path is not empty
	if logFile == "" {
		return errors.New("log file path cannot be empty")
	}

	// Validate log file extension
	if filepath.Ext(logFile) != ".log" {
		return fmt.Errorf("log file must have .log extension, got: %s", logFile)
	}

	return nil
}

// GetDefaultConfigPaths returns possible config file paths
func GetDefaultConfigPaths() []string {
	cwd, _ := os.Getwd()
	return []string{
		"/etc/beacon/beacon.yaml",
		filepath.Join(cwd, "beacon.yaml"),
	}
}

// parseYAMLError parses YAML errors and extracts line numbers with helpful suggestions
func parseYAMLError(err error, data []byte) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Try to extract line number from Viper's error message
	var lineNumber int
	if strings.Contains(errMsg, "line ") {
		parts := strings.Split(errMsg, "line ")
		if len(parts) > 1 {
			numParts := strings.Split(parts[1], " ")
			if len(numParts) > 0 {
				fmt.Sscanf(numParts[0], "%d", &lineNumber)
			}
		}
	}

	// Analyze common YAML errors and provide specific suggestions
	var suggestion string

	// Check for indentation errors
	if strings.Contains(errMsg, "indentation") || strings.Contains(errMsg, "block sequence") {
		lines := strings.Split(string(data), "\n")
		if lineNumber > 0 && lineNumber <= len(lines) {
			lineContent := strings.TrimSpace(lines[lineNumber-1])
			indentCount := len(lines[lineNumber-1]) - len(lineContent)
			suggestion = fmt.Sprintf("第 %d 行缩进错误：当前缩进为 %d 个空格，建议使用 2 个空格", lineNumber, indentCount)
		} else {
			suggestion = "缩进错误：YAML 应使用 2 个空格缩进，不要使用 Tab"
		}
	} else if strings.Contains(errMsg, "unclosed") || strings.Contains(errMsg, "unterminated") {
		suggestion = "语法错误：请检查 YAML 结构是否完整（引号、括号、列表等是否闭合）"
	} else if strings.Contains(errMsg, "mapping values") {
		suggestion = "格式错误：请确保键值对使用冒号 ':' 分隔，且冒号后有空格"
	} else if strings.Contains(errMsg, "could not find expected") {
		suggestion = "格式错误：请检查 YAML 结构，可能缺少换行或缩进不正确"
	} else {
		suggestion = "请检查 YAML 语法是否正确，确保缩进和格式符合规范"
	}

	// Build detailed error message
	if lineNumber > 0 {
		return fmt.Errorf("配置格式错误：第 %d 行 - %s\n%s", lineNumber, suggestion, errMsg)
	}

	return fmt.Errorf("配置格式错误：%s\n%s", suggestion, errMsg)
}

// SaveConfig saves configuration to file (optional feature for MVP)
// Note: MVP saves node_id to memory only. File write is optional for production use.
func SaveConfig(cfg *Config, path string) error {
	if path == "" {
		return errors.New("config path is empty")
	}

	// Create a new Viper instance
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Set all config values
	v.Set("pulse_server", cfg.PulseServer)
	v.Set("node_id", cfg.NodeID)
	v.Set("node_name", cfg.NodeName)
	if cfg.Region != "" {
		v.Set("region", cfg.Region)
	}
	if len(cfg.Tags) > 0 {
		v.Set("tags", cfg.Tags)
	}
	if len(cfg.Probes) > 0 {
		v.Set("probes", cfg.Probes)
	}
	if cfg.Reconnect.MaxRetries > 0 || cfg.Reconnect.RetryInterval > 0 || cfg.Reconnect.Backoff != "" {
		v.Set("reconnect", cfg.Reconnect)
	}

	// Write config to file
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ResourceMonitorConfig represents resource monitoring configuration (Story 3.11)
type ResourceMonitorConfig struct {
	Enabled              bool              `mapstructure:"enabled" yaml:"enabled"`
	CheckIntervalSeconds int               `mapstructure:"check_interval_seconds" yaml:"check_interval_seconds"`
	Thresholds           ThresholdsConfig  `mapstructure:"thresholds" yaml:"thresholds"`
	Degradation          DegradationConfig `mapstructure:"degradation" yaml:"degradation"`
	Alerting             AlertingConfig    `mapstructure:"alerting" yaml:"alerting"`
}

// ThresholdsConfig represents resource threshold configuration
type ThresholdsConfig struct {
	CPUMicrocores int `mapstructure:"cpu_microcores" yaml:"cpu_microcores"`
	MemoryMB      int `mapstructure:"memory_mb" yaml:"memory_mb"`
}

// DegradationConfig represents degradation policy configuration
type DegradationConfig struct {
	DegradedLevel  DegradationLevelConfig `mapstructure:"degraded_level" yaml:"degraded_level"`
	CriticalLevel  DegradationLevelConfig `mapstructure:"critical_level" yaml:"critical_level"`
	Recovery       RecoveryConfig         `mapstructure:"recovery" yaml:"recovery"`
}

// DegradationLevelConfig represents a single degradation level configuration
type DegradationLevelConfig struct {
	CPUMicrocores      int `mapstructure:"cpu_microcores" yaml:"cpu_microcores"`
	MemoryMB           int `mapstructure:"memory_mb" yaml:"memory_mb"`
	IntervalMultiplier int `mapstructure:"interval_multiplier" yaml:"interval_multiplier"`
}

// RecoveryConfig represents auto-recovery configuration
type RecoveryConfig struct {
	ConsecutiveNormalChecks int `mapstructure:"consecutive_normal_checks" yaml:"consecutive_normal_checks"`
}

// AlertingConfig represents alert configuration
type AlertingConfig struct {
	SuppressionWindowSeconds int `mapstructure:"suppression_window_seconds" yaml:"suppression_window_seconds"`
}

package config

import (
	"fmt"
	"sync"
	"time"
)

// OperatingMode represents the beacon's operating mode (FR-4.1.2)
type OperatingMode string

const (
	// ModeStandalone operates independently without Pulse server connection
	// Loads probe tasks from local config, exposes /metrics endpoint
	ModeStandalone OperatingMode = "standalone"

	// ModeRegistered connects to Pulse server, receives probe config
	// Checks for config updates every 60 seconds
	ModeRegistered OperatingMode = "registered"

	// ModeDegraded is entered after 3 consecutive heartbeat failures
	// Uses cached config, exits on successful heartbeat
	ModeDegraded OperatingMode = "degraded"
)

// ConfigSource represents where the current configuration came from
type ConfigSource string

const (
	// SourceLocal config loaded from local file
	SourceLocal ConfigSource = "local"

	// SourceServer config received from Pulse server
	SourceServer ConfigSource = "server"

	// SourceCached config loaded from cache (degraded mode)
	SourceCached ConfigSource = "cached"
)

// ModeConfig contains configuration for dual mode support
type ModeConfig struct {
	// Operating mode: standalone or registered (default: registered)
	Mode OperatingMode `mapstructure:"mode" yaml:"mode"`

	// Config check interval in seconds when in registered mode (default: 60)
	ConfigCheckIntervalSeconds int `mapstructure:"config_check_interval_seconds" yaml:"config_check_interval_seconds"`

	// Number of consecutive failures before entering degraded mode (default: 3)
	DegradedModeThreshold int `mapstructure:"degraded_mode_threshold" yaml:"degraded_mode_threshold"`

	// Node token for registration (required in registered mode)
	NodeToken string `mapstructure:"node_token" yaml:"node_token"`
}

// CompressionConfig contains configuration for data compression (FR-4.1.5)
type CompressionConfig struct {
	// Enable GZIP compression (default: true)
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Compression level 1-9 (default: 6)
	Level int `mapstructure:"level" yaml:"level"`

	// Minimum data size in bytes to trigger compression (default: 1024)
	MinSizeBytes int `mapstructure:"min_size_bytes" yaml:"min_size_bytes"`
}

// ModeManager manages the beacon's operating mode and state transitions
type ModeManager struct {
	mu sync.RWMutex

	// Current operating mode
	currentMode OperatingMode

	// Current config source
	configSource ConfigSource

	// Consecutive heartbeat failure count
	consecutiveFailures int

	// Threshold for entering degraded mode
	degradedThreshold int

	// Config check interval
	configCheckInterval time.Duration

	// Last successful heartbeat time
	lastSuccessTime time.Time

	// Last failure time
	lastFailureTime time.Time

	// Mode change callbacks
	callbacks []func(OperatingMode, OperatingMode)
}

// NewModeManager creates a new mode manager
func NewModeManager(cfg *ModeConfig) *ModeManager {
	if cfg == nil {
		cfg = &ModeConfig{}
	}

	// Set defaults
	mode := cfg.Mode
	if mode == "" {
		mode = ModeRegistered // Default to registered mode
	}

	threshold := cfg.DegradedModeThreshold
	if threshold == 0 {
		threshold = 3 // Default: 3 consecutive failures
	}

	checkInterval := cfg.ConfigCheckIntervalSeconds
	if checkInterval == 0 {
		checkInterval = 60 // Default: 60 seconds
	}

	return &ModeManager{
		currentMode:         mode,
		configSource:        SourceLocal,
		degradedThreshold:   threshold,
		configCheckInterval: time.Duration(checkInterval) * time.Second,
	}
}

// GetMode returns the current operating mode
func (m *ModeManager) GetMode() OperatingMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMode
}

// GetConfigSource returns the current config source
func (m *ModeManager) GetConfigSource() ConfigSource {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configSource
}

// SetConfigSource sets the config source
func (m *ModeManager) SetConfigSource(source ConfigSource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configSource = source
}

// RecordHeartbeatSuccess records a successful heartbeat and handles mode transitions
func (m *ModeManager) RecordHeartbeatSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastSuccessTime = time.Now()
	m.consecutiveFailures = 0

	// Exit degraded mode on successful heartbeat
	if m.currentMode == ModeDegraded {
		oldMode := m.currentMode
		m.currentMode = ModeRegistered
		m.invokeCallbacks(oldMode, m.currentMode)
	}
}

// RecordHeartbeatFailure records a failed heartbeat and handles mode transitions
func (m *ModeManager) RecordHeartbeatFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastFailureTime = time.Now()
	m.consecutiveFailures++

	// Enter degraded mode after threshold failures
	if m.currentMode == ModeRegistered && m.consecutiveFailures >= m.degradedThreshold {
		oldMode := m.currentMode
		m.currentMode = ModeDegraded
		m.configSource = SourceCached
		m.invokeCallbacks(oldMode, m.currentMode)
	}
}

// GetConsecutiveFailures returns the current consecutive failure count
func (m *ModeManager) GetConsecutiveFailures() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveFailures
}

// GetConfigCheckInterval returns the config check interval
func (m *ModeManager) GetConfigCheckInterval() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configCheckInterval
}

// OnModeChange registers a callback for mode changes
func (m *ModeManager) OnModeChange(callback func(oldMode, newMode OperatingMode)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, callback)
}

// invokeCallbacks calls all registered mode change callbacks (must be called with lock held)
func (m *ModeManager) invokeCallbacks(oldMode, newMode OperatingMode) {
	for _, callback := range m.callbacks {
		callback(oldMode, newMode)
	}
}

// IsConnected returns true if beacon is in a mode that connects to server
func (m *ModeManager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMode == ModeRegistered || m.currentMode == ModeDegraded
}

// IsStandalone returns true if beacon is in standalone mode
func (m *ModeManager) IsStandalone() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMode == ModeStandalone
}

// GetLastSuccessTime returns the last successful heartbeat time
func (m *ModeManager) GetLastSuccessTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastSuccessTime
}

// GetLastFailureTime returns the last failed heartbeat time
func (m *ModeManager) GetLastFailureTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastFailureTime
}

// GetStatus returns a summary of the current mode status
func (m *ModeManager) GetStatus() ModeStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return ModeStatus{
		CurrentMode:         string(m.currentMode),
		ConfigSource:        string(m.configSource),
		ConsecutiveFailures: m.consecutiveFailures,
		DegradedThreshold:   m.degradedThreshold,
		LastSuccessTime:     m.lastSuccessTime,
		LastFailureTime:     m.lastFailureTime,
		ConfigCheckInterval: m.configCheckInterval.String(),
	}
}

// ModeStatus represents the current mode status for diagnostics
type ModeStatus struct {
	CurrentMode         string    `json:"current_mode"`
	ConfigSource        string    `json:"config_source"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	DegradedThreshold   int       `json:"degraded_threshold"`
	LastSuccessTime     time.Time `json:"last_success_time"`
	LastFailureTime     time.Time `json:"last_failure_time"`
	ConfigCheckInterval string    `json:"config_check_interval"`
}

// Validate validates the mode configuration
func (c *ModeConfig) Validate() error {
	// Validate mode
	if c.Mode != "" && c.Mode != ModeStandalone && c.Mode != ModeRegistered {
		return fmt.Errorf("invalid mode '%s', must be 'standalone' or 'registered'", c.Mode)
	}

	// Validate config check interval
	if c.ConfigCheckIntervalSeconds != 0 && (c.ConfigCheckIntervalSeconds < 10 || c.ConfigCheckIntervalSeconds > 300) {
		return fmt.Errorf("config_check_interval_seconds must be between 10 and 300, got %d", c.ConfigCheckIntervalSeconds)
	}

	// Validate degraded threshold
	if c.DegradedModeThreshold != 0 && (c.DegradedModeThreshold < 1 || c.DegradedModeThreshold > 10) {
		return fmt.Errorf("degraded_mode_threshold must be between 1 and 10, got %d", c.DegradedModeThreshold)
	}

	return nil
}

// Validate validates the compression configuration
func (c *CompressionConfig) Validate() error {
	// Validate compression level
	if c.Level != 0 && (c.Level < 1 || c.Level > 9) {
		return fmt.Errorf("compression level must be between 1 and 9, got %d", c.Level)
	}

	// Validate min size
	if c.MinSizeBytes != 0 && c.MinSizeBytes < 0 {
		return fmt.Errorf("min_size_bytes cannot be negative, got %d", c.MinSizeBytes)
	}

	return nil
}

// ResumeConfig contains configuration for resume upload feature (FR-4.1.5, FR-4.1.7)
type ResumeConfig struct {
	// Enable resume upload feature (default: true)
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Maximum cache size in bytes for offline data (default: 10MB)
	MaxCacheSizeBytes int64 `mapstructure:"max_cache_size_bytes" yaml:"max_cache_size_bytes"`

	// Cache file path for persistence (default: /var/lib/beacon/resume_cache.dat)
	CacheFilePath string `mapstructure:"cache_file_path" yaml:"cache_file_path"`

	// Enable priority mode for alert data (FR-4.1.7, default: true)
	AlertPriorityMode bool `mapstructure:"alert_priority_mode" yaml:"alert_priority_mode"`

	// Percentage of cache reserved for alert data (FR-4.1.7, default: 30)
	AlertReservePercent int `mapstructure:"alert_reserve_percent" yaml:"alert_reserve_percent"`
}

// Validate validates the resume configuration
func (c *ResumeConfig) Validate() error {
	// Validate max cache size (max 100MB)
	if c.MaxCacheSizeBytes != 0 && (c.MaxCacheSizeBytes < 0 || c.MaxCacheSizeBytes > 100*1024*1024) {
		return fmt.Errorf("max_cache_size_bytes must be between 0 and 100MB, got %d", c.MaxCacheSizeBytes)
	}

	// Validate alert reserve percent (0-100)
	if c.AlertReservePercent != 0 && (c.AlertReservePercent < 0 || c.AlertReservePercent > 100) {
		return fmt.Errorf("alert_reserve_percent must be between 0 and 100, got %d", c.AlertReservePercent)
	}

	return nil
}

package monitor

import (
	"errors"
	"sync"
)

// Errors
var (
	ErrInvalidConfig    = errors.New("invalid monitor configuration")
	ErrMonitorDisabled  = errors.New("resource monitor is disabled")
	ErrAlreadyRunning   = errors.New("monitor is already running")
	ErrNotRunning       = errors.New("monitor is not running")
)

// DegradationLevel represents the resource degradation level
type DegradationLevel int

const (
	DegradationLevelNormal   DegradationLevel = iota
	DegradationLevelDegraded
	DegradationLevelCritical
)

// String returns the string representation of the degradation level
func (d DegradationLevel) String() string {
	switch d {
	case DegradationLevelNormal:
		return "normal"
	case DegradationLevelDegraded:
		return "degraded"
	case DegradationLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ResourceUsage represents current resource usage statistics
type ResourceUsage struct {
	CPUMicrocores float64   `json:"cpu_microcores"`
	MemoryMB      float64   `json:"memory_mb"`
	Timestamp     int64     `json:"timestamp"`
}

// Alert represents a resource alert event
type Alert struct {
	ResourceType string  `json:"resource_type"` // "cpu" or "memory"
	CurrentValue float64 `json:"current_value"`
	Threshold    float64 `json:"threshold"`
	Level        string  `json:"level"` // "degraded" or "critical"
	Timestamp    int64   `json:"timestamp"`
}

// ProbeManager is the interface for updating probe intervals
// This avoids circular dependency with the probe package
type ProbeManager interface {
	UpdateProbeInterval(multiplier int) error
}

// Monitor is the resource monitor interface
type Monitor interface {
	// Start starts the resource monitoring
	Start() error

	// Stop stops the resource monitoring
	Stop()

	// GetDegradationLevel returns the current degradation level
	GetDegradationLevel() DegradationLevel

	// GetResourceUsage returns the latest resource usage
	GetResourceUsage() *ResourceUsage

	// GetAlerts returns the history of alerts
	GetAlerts() []Alert

	// IsRunning returns whether the monitor is running
	IsRunning() bool
}

// monitor implements the Monitor interface
type monitor struct {
	cfg      *ResourceMonitorConfig
	probeMgr ProbeManager
	logger   Logger

	// State management
	mu     sync.RWMutex
	level  DegradationLevel
	currentUsage *ResourceUsage
	alerts       []Alert
	lastAlertTime map[string]int64 // resource_type -> last_alert_timestamp

	// Recovery tracking
	consecutiveNormalChecks int

	// Control
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewMonitor creates a new resource monitor
func NewMonitor(cfg *ResourceMonitorConfig, probeMgr ProbeManager, logger Logger) (Monitor, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}
	if !cfg.Enabled {
		return nil, ErrMonitorDisabled
	}

	return &monitor{
		cfg:            cfg,
		probeMgr:       probeMgr,
		logger:         logger,
		level:          DegradationLevelNormal,
		alerts:         make([]Alert, 0, 100), // Pre-allocate for 100 alerts
		lastAlertTime:  make(map[string]int64),
		stopCh:         make(chan struct{}),
	}, nil
}

// ResourceMonitorConfig represents resource monitoring configuration
type ResourceMonitorConfig struct {
	Enabled              bool                `mapstructure:"enabled" yaml:"enabled"`
	CheckIntervalSeconds int                 `mapstructure:"check_interval_seconds" yaml:"check_interval_seconds"`
	Thresholds           ThresholdsConfig    `mapstructure:"thresholds" yaml:"thresholds"`
	Degradation          DegradationConfig   `mapstructure:"degradation" yaml:"degradation"`
	Alerting             AlertingConfig      `mapstructure:"alerting" yaml:"alerting"`
}

// ThresholdsConfig represents resource threshold configuration
type ThresholdsConfig struct {
	CPUMicrocores int `mapstructure:"cpu_microcores" yaml:"cpu_microcores"`
	MemoryMB      int `mapstructure:"memory_mb" yaml:"memory_mb"`
}

// DegradationConfig represents degradation policy configuration
type DegradationConfig struct {
	DegradedLevel DegradationLevelConfig `mapstructure:"degraded_level" yaml:"degraded_level"`
	CriticalLevel DegradationLevelConfig `mapstructure:"critical_level" yaml:"critical_level"`
	Recovery      RecoveryConfig         `mapstructure:"recovery" yaml:"recovery"`
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

// ConfigFromConfig converts config.ResourceMonitorConfig to monitor.ResourceMonitorConfig
// This is defined here to avoid circular import
func ConfigFromConfig(cfg interface{}) *ResourceMonitorConfig {
	// For now, return nil - conversion will be done by the caller
	return nil
}

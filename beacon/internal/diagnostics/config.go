package diagnostics

import (
	"os"
	"time"
)

// Configuration contains configuration diagnostic information
type Configuration struct {
	ConfigFile     string                 `json:"config_file"`
	ConfigValid    bool                   `json:"config_valid"`
	ConfigVersion  string                 `json:"config_version"`
	LogLevel       string                 `json:"log_level"`
	DebugMode      bool                   `json:"debug_mode"`
	OperatingMode  string                 `json:"operating_mode"`  // FR-4.1.2
	ConfigSource   string                 `json:"config_source"`   // FR-4.1.2
	Compression    CompressionInfo        `json:"compression"`     // FR-4.1.5
	Resume         ResumeInfo             `json:"resume"`          // FR-4.1.5, FR-4.1.7
	ConfigContent  map[string]interface{} `json:"config_content"`
}

// CompressionInfo contains compression diagnostic information
type CompressionInfo struct {
	Enabled    bool    `json:"enabled"`
	Level      int     `json:"level"`
	MinSizeKB  float64 `json:"min_size_kb"`
}

// ResumeInfo contains resume upload diagnostic information
type ResumeInfo struct {
	Enabled              bool    `json:"enabled"`
	MaxCacheSizeMB       float64 `json:"max_cache_size_mb"`
	AlertPriorityMode    bool    `json:"alert_priority_mode"`
	AlertReservePercent  int     `json:"alert_reserve_percent"`
}

// collectConfiguration collects configuration information
func (c *collector) collectConfiguration() (*Configuration, error) {
	configInfo := &Configuration{
		ConfigFile:    c.cfg.ConfigPath,
		ConfigValid:   true,
		ConfigVersion: c.getConfigVersion(),
		LogLevel:      c.cfg.LogLevel,
		DebugMode:     c.cfg.DebugMode,
		OperatingMode: string(c.cfg.Mode.Mode),
		ConfigSource:  string(c.cfg.Mode.Mode), // Will be updated by mode manager at runtime
		Compression: CompressionInfo{
			Enabled:   c.cfg.Compression.Enabled,
			Level:     c.cfg.Compression.Level,
			MinSizeKB: float64(c.cfg.Compression.MinSizeBytes) / 1024,
		},
		Resume: ResumeInfo{
			Enabled:             c.cfg.Resume.Enabled,
			MaxCacheSizeMB:      float64(c.cfg.Resume.MaxCacheSizeBytes) / 1024 / 1024,
			AlertPriorityMode:   c.cfg.Resume.AlertPriorityMode,
			AlertReservePercent: c.cfg.Resume.AlertReservePercent,
		},
		ConfigContent: map[string]interface{}{
			"pulse_server": c.cfg.PulseServer,
			"node_id":      c.cfg.NodeID,
			"node_name":    c.cfg.NodeName,
			"region":       c.cfg.Region,
			"tags":         c.cfg.Tags,
			"log_level":    c.cfg.LogLevel,
			"debug_mode":   c.cfg.DebugMode,
			"mode":         string(c.cfg.Mode.Mode),
		},
	}

	// Add metrics configuration if enabled
	if c.cfg.MetricsEnabled {
		configInfo.ConfigContent["metrics_enabled"] = true
		configInfo.ConfigContent["metrics_port"] = c.cfg.MetricsPort
		configInfo.ConfigContent["metrics_update_seconds"] = c.cfg.MetricsUpdateSeconds
	}

	// Add probe configuration if present
	if len(c.cfg.Probes) > 0 {
		configInfo.ConfigContent["probes"] = c.cfg.Probes
	}

	return configInfo, nil
}

// getConfigVersion gets the configuration version based on file modification time
func (c *collector) getConfigVersion() string {
	if c.cfg.ConfigPath == "" {
		return "unknown"
	}

	fileInfo, err := os.Stat(c.cfg.ConfigPath)
	if err != nil {
		return "unknown"
	}

	return fileInfo.ModTime().Format(time.RFC3339)
}

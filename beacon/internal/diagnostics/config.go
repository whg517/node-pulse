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
	ConfigContent  map[string]interface{} `json:"config_content"`
}

// collectConfiguration collects configuration information
func (c *collector) collectConfiguration() (*Configuration, error) {
	configInfo := &Configuration{
		ConfigFile:    c.cfg.ConfigPath,
		ConfigValid:   true,
		ConfigVersion: c.getConfigVersion(),
		LogLevel:      c.cfg.LogLevel,
		DebugMode:     c.cfg.DebugMode,
		ConfigContent: map[string]interface{}{
			"pulse_server": c.cfg.PulseServer,
			"node_id":      c.cfg.NodeID,
			"node_name":    c.cfg.NodeName,
			"region":       c.cfg.Region,
			"tags":         c.cfg.Tags,
			"log_level":    c.cfg.LogLevel,
			"debug_mode":   c.cfg.DebugMode,
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

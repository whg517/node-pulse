package diagnostics

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// ResourceUsage contains system resource usage information
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryMB      float64 `json:"memory_mb"`
	MemoryPercent float64 `json:"memory_percent"`
}

// ResourceMonitorInfo contains resource monitor information (Story 3.11)
type ResourceMonitorInfo struct {
	Enabled          bool     `json:"enabled"`
	Running          bool     `json:"running"`
	DegradationLevel string   `json:"degradation_level"`
	AlertCount       int      `json:"alert_count"`
	Thresholds       Thresholds `json:"thresholds"`

	// Extended configuration fields (Story 3.11)
	Configured                bool            `json:"configured"`
	CheckIntervalSeconds      int             `json:"check_interval_seconds,omitempty"`
	AlertSuppressionMinutes   int             `json:"alert_suppression_minutes,omitempty"`
	DegradedLevelThreshold    ThresholdConfig `json:"degraded_level_threshold,omitempty"`
	CriticalLevelThreshold    ThresholdConfig `json:"critical_level_threshold,omitempty"`
	RecoveryChecksRequired    int             `json:"recovery_checks_required,omitempty"`
}

// ThresholdConfig represents a degradation level threshold configuration
type ThresholdConfig struct {
	CPUMicrocores      int `json:"cpu_microcores"`
	MemoryMB           int `json:"memory_mb"`
	IntervalMultiplier int `json:"interval_multiplier"`
}

// Thresholds represents resource monitoring thresholds
type Thresholds struct {
	CPUMicrocores int `json:"cpu_microcores"`
	MemoryMB      int `json:"memory_mb"`
}

// collectResourceUsage collects system resource usage information
func (c *collector) collectResourceUsage() (*ResourceUsage, error) {
	// Get CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}

	// Get memory usage
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	usage := &ResourceUsage{
		CPUPercent:    cpuPercent[0],
		MemoryMB:      float64(memStat.Used) / 1024 / 1024,
		MemoryPercent: memStat.UsedPercent,
	}

	return usage, nil
}

// collectResourceMonitorInfo collects resource monitor information (Story 3.11)
func (c *collector) collectResourceMonitorInfo() *ResourceMonitorInfo {
	if !c.cfg.ResourceMonitor.Enabled {
		return &ResourceMonitorInfo{
			Enabled: false,
		}
	}

	// Calculate total suppression window in human-readable format
	suppressionMinutes := c.cfg.ResourceMonitor.Alerting.SuppressionWindowSeconds / 60

	// Note: Runtime status (running, degradation_level, alert_count) is only available
	// when beacon is running. Debug command shows configuration status.
	// For live runtime status, check logs or use monitoring tools.
	info := &ResourceMonitorInfo{
		Enabled: true,
		Running:          false,
		DegradationLevel: "not_applicable",
		AlertCount:       -1, // -1 indicates "not available in debug mode"
		Thresholds: Thresholds{
			CPUMicrocores: c.cfg.ResourceMonitor.Thresholds.CPUMicrocores,
			MemoryMB:      c.cfg.ResourceMonitor.Thresholds.MemoryMB,
		},
		Configured: true,
		CheckIntervalSeconds: c.cfg.ResourceMonitor.CheckIntervalSeconds,
		AlertSuppressionMinutes: suppressionMinutes,
		DegradedLevelThreshold: ThresholdConfig{
			CPUMicrocores: c.cfg.ResourceMonitor.Degradation.DegradedLevel.CPUMicrocores,
			MemoryMB:      c.cfg.ResourceMonitor.Degradation.DegradedLevel.MemoryMB,
			IntervalMultiplier: c.cfg.ResourceMonitor.Degradation.DegradedLevel.IntervalMultiplier,
		},
		CriticalLevelThreshold: ThresholdConfig{
			CPUMicrocores: c.cfg.ResourceMonitor.Degradation.CriticalLevel.CPUMicrocores,
			MemoryMB:      c.cfg.ResourceMonitor.Degradation.CriticalLevel.MemoryMB,
			IntervalMultiplier: c.cfg.ResourceMonitor.Degradation.CriticalLevel.IntervalMultiplier,
		},
		RecoveryChecksRequired: c.cfg.ResourceMonitor.Degradation.Recovery.ConsecutiveNormalChecks,
	}

	return info
}

package diagnostics

import (
	"encoding/json"
	"fmt"
	"time"

	"beacon/internal/config"
)

// DiagnosticInfo contains all diagnostic information
type DiagnosticInfo struct {
	Timestamp   string             `json:"timestamp"`
	Level       string             `json:"level"`
	Message     string             `json:"message"`
	NodeID      string             `json:"node_id,omitempty"`
	NodeName    string             `json:"node_name,omitempty"`
	Diagnostics DiagnosticDetails  `json:"diagnostics"`
}

// DiagnosticDetails contains detailed diagnostic information
type DiagnosticDetails struct {
	NetworkStatus       NetworkStatus       `json:"network_status"`
	Configuration       Configuration       `json:"configuration"`
	ConnectionStatus    ConnectionStatus    `json:"connection_status"`
	ResourceUsage       ResourceUsage       `json:"resource_usage"`
	ResourceMonitor     *ResourceMonitorInfo `json:"resource_monitor,omitempty"`
	ProbeTasks          ProbeTasks          `json:"probe_tasks"`
	PrometheusMetrics   PrometheusMetrics   `json:"prometheus_metrics"`

	// FR-4.1.2 Enhanced mode status
	ModeStatus          *ModeStatusInfo     `json:"mode_status,omitempty"`
}

// ModeStatusInfo contains mode and config source status
type ModeStatusInfo struct {
	CurrentMode         string    `json:"current_mode"`
	ConfigSource        string    `json:"config_source"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	LastSuccessTime     time.Time `json:"last_success_time,omitempty"`
	LastFailureTime     time.Time `json:"last_failure_time,omitempty"`
}

// Collector is the diagnostic information collector interface
type Collector interface {
	Collect() (*DiagnosticInfo, error)
	CollectJSON() ([]byte, error)
	CollectPretty() (string, error)

	// Provider setters for enhanced metrics
	SetModeProvider(provider ModeProvider)
	SetCacheStatsProvider(provider CacheStatsProvider)
	SetCompressionStatsProvider(provider CompressionStatsProvider)
	SetProbeCountProvider(provider ProbeCountProvider)
}

// collector implements the Collector interface
type collector struct {
	cfg       *config.Config
	startTime time.Time

	// Providers for enhanced diagnostics
	modeProvider        ModeProvider
	cacheStatsProvider  CacheStatsProvider
	compressionProvider CompressionStatsProvider
	probeCountProvider  ProbeCountProvider
}

// NewCollector creates a new diagnostic information collector
func NewCollector(cfg *config.Config) Collector {
	return &collector{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// SetModeProvider sets the mode provider for enhanced diagnostics
func (c *collector) SetModeProvider(provider ModeProvider) {
	c.modeProvider = provider
}

// SetCacheStatsProvider sets the cache stats provider
func (c *collector) SetCacheStatsProvider(provider CacheStatsProvider) {
	c.cacheStatsProvider = provider
}

// SetCompressionStatsProvider sets the compression stats provider
func (c *collector) SetCompressionStatsProvider(provider CompressionStatsProvider) {
	c.compressionProvider = provider
}

// SetProbeCountProvider sets the probe count provider
func (c *collector) SetProbeCountProvider(provider ProbeCountProvider) {
	c.probeCountProvider = provider
}

// Collect collects all diagnostic information
func (c *collector) Collect() (*DiagnosticInfo, error) {
	info := &DiagnosticInfo{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "DEBUG",
		Message:   "Beacon diagnostic information",
		NodeID:    c.cfg.NodeID,
		NodeName:  c.cfg.NodeName,
	}

	// Collect network status
	networkStatus, err := c.collectNetworkStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to collect network status: %w", err)
	}
	info.Diagnostics.NetworkStatus = *networkStatus

	// Collect configuration info
	configInfo, err := c.collectConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to collect configuration: %w", err)
	}
	info.Diagnostics.Configuration = *configInfo

	// Collect connection status
	connectionStatus, err := c.collectConnectionStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to collect connection status: %w", err)
	}
	info.Diagnostics.ConnectionStatus = *connectionStatus

	// Collect resource usage
	resourceUsage, err := c.collectResourceUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to collect resource usage: %w", err)
	}
	info.Diagnostics.ResourceUsage = *resourceUsage

	// Collect resource monitor info (Story 3.11)
	resourceMonitorInfo := c.collectResourceMonitorInfo()
	if resourceMonitorInfo != nil {
		info.Diagnostics.ResourceMonitor = resourceMonitorInfo
	}

	// Collect probe tasks status
	probeTasks, err := c.collectProbeTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to collect probe tasks: %w", err)
	}
	info.Diagnostics.ProbeTasks = *probeTasks

	// Collect Prometheus metrics
	promMetrics, err := c.collectPrometheusMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to collect Prometheus metrics: %w", err)
	}
	info.Diagnostics.PrometheusMetrics = *promMetrics

	// Collect mode status (FR-4.1.2)
	modeStatus := c.collectModeStatus()
	if modeStatus != nil {
		info.Diagnostics.ModeStatus = modeStatus
	}

	return info, nil
}

// collectModeStatus collects the current mode status
func (c *collector) collectModeStatus() *ModeStatusInfo {
	if c.modeProvider == nil {
		return nil
	}

	status := &ModeStatusInfo{
		CurrentMode:    string(c.modeProvider.GetMode()),
		ConfigSource:   string(c.modeProvider.GetConfigSource()),
	}

	// If modeProvider also implements ModeManager interface, get additional info
	type modeStatusProvider interface {
		GetConsecutiveFailures() int
		GetLastSuccessTime() time.Time
		GetLastFailureTime() time.Time
	}

	if msp, ok := c.modeProvider.(modeStatusProvider); ok {
		status.ConsecutiveFailures = msp.GetConsecutiveFailures()
		status.LastSuccessTime = msp.GetLastSuccessTime()
		status.LastFailureTime = msp.GetLastFailureTime()
	}

	return status
}

// CollectJSON collects diagnostic information and returns as JSON
func (c *collector) CollectJSON() ([]byte, error) {
	info, err := c.Collect()
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal diagnostic info: %w", err)
	}

	return data, nil
}

// CollectPretty collects diagnostic information and returns as human-readable string
func (c *collector) CollectPretty() (string, error) {
	info, err := c.Collect()
	if err != nil {
		return "", err
	}

	// Format connection status
	connStatus := fmt.Sprintf("Status: %s\n", info.Diagnostics.ConnectionStatus.Status)
	if info.Diagnostics.ConnectionStatus.LastSuccess != nil {
		connStatus += fmt.Sprintf("Last Success: %s\n", info.Diagnostics.ConnectionStatus.LastSuccess.Format(time.RFC3339))
	}
	if info.Diagnostics.ConnectionStatus.LastFailure != nil {
		connStatus += fmt.Sprintf("Last Failure: %s\n", info.Diagnostics.ConnectionStatus.LastFailure.Format(time.RFC3339))
		connStatus += fmt.Sprintf("Failure Reason: %s\n", info.Diagnostics.ConnectionStatus.FailureReason)
	}
	connStatus += fmt.Sprintf("Retry Count: %d\n", info.Diagnostics.ConnectionStatus.RetryCount)
	connStatus += fmt.Sprintf("Backoff: %d seconds\n", info.Diagnostics.ConnectionStatus.BackoffSeconds)
	if info.Diagnostics.ConnectionStatus.NextRetry != nil {
		connStatus += fmt.Sprintf("Next Retry: %s\n", info.Diagnostics.ConnectionStatus.NextRetry.Format(time.RFC3339))
	}
	connStatus += fmt.Sprintf("Queue Size: %d\n", info.Diagnostics.ConnectionStatus.QueueSize)

	// Format probe tasks
	probeInfo := fmt.Sprintf("Total Tasks: %d\n", info.Diagnostics.ProbeTasks.TotalTasks)
	probeInfo += fmt.Sprintf("Running Tasks: %d\n", info.Diagnostics.ProbeTasks.RunningTasks)
	if info.Diagnostics.ProbeTasks.TotalExecs > 0 {
		probeInfo += fmt.Sprintf("Total Executions: %d\n", info.Diagnostics.ProbeTasks.TotalExecs)
		probeInfo += fmt.Sprintf("Successful: %d\n", info.Diagnostics.ProbeTasks.SuccessExecs)
		probeInfo += fmt.Sprintf("Failed: %d\n", info.Diagnostics.ProbeTasks.FailureExecs)
	}

	// Format Prometheus metrics with enhanced metrics
	promInfo := fmt.Sprintf("Beacon Up: %.0f\n", info.Diagnostics.PrometheusMetrics.BeaconUp)
	promInfo += fmt.Sprintf("RTT: %.6f seconds\n", info.Diagnostics.PrometheusMetrics.RTTSeconds)
	promInfo += fmt.Sprintf("Packet Loss: %.4f\n", info.Diagnostics.PrometheusMetrics.PacketLossRate)
	if info.Diagnostics.PrometheusMetrics.JitterMs > 0 {
		promInfo += fmt.Sprintf("Jitter: %.2f ms\n", info.Diagnostics.PrometheusMetrics.JitterMs)
	}
	// Enhanced metrics
	promInfo += fmt.Sprintf("Mode: %s\n", info.Diagnostics.PrometheusMetrics.Mode)
	promInfo += fmt.Sprintf("Config Source: %s\n", info.Diagnostics.PrometheusMetrics.ConfigSource)
	promInfo += fmt.Sprintf("Active Probes: %d\n", info.Diagnostics.PrometheusMetrics.ActiveProbes)
	if info.Diagnostics.PrometheusMetrics.CompressionRatio > 0 {
		promInfo += fmt.Sprintf("Compression Ratio: %.2f%%\n", info.Diagnostics.PrometheusMetrics.CompressionRatio)
	}
	if info.Diagnostics.PrometheusMetrics.CacheSizeBytes > 0 {
		promInfo += fmt.Sprintf("Cache Size: %d bytes\n", info.Diagnostics.PrometheusMetrics.CacheSizeBytes)
	}

	// Format resource monitor (Story 3.11)
	resourceMonitorInfo := "Disabled\n"
	if info.Diagnostics.ResourceMonitor != nil && info.Diagnostics.ResourceMonitor.Enabled {
		if info.Diagnostics.ResourceMonitor.AlertCount >= 0 {
			// Running instance has real data
			resourceMonitorInfo = fmt.Sprintf("Enabled: %v\nRunning: %v\nDegradation Level: %s\nAlert Count: %d\n",
				info.Diagnostics.ResourceMonitor.Enabled,
				info.Diagnostics.ResourceMonitor.Running,
				info.Diagnostics.ResourceMonitor.DegradationLevel,
				info.Diagnostics.ResourceMonitor.AlertCount)
		} else {
			// Debug mode shows configuration
			resourceMonitorInfo = fmt.Sprintf("Enabled: %v\nStatus: configured (debug mode)\nCPU Threshold: %d microcores\nMemory Threshold: %d MB\n",
				info.Diagnostics.ResourceMonitor.Enabled,
				info.Diagnostics.ResourceMonitor.Thresholds.CPUMicrocores,
				info.Diagnostics.ResourceMonitor.Thresholds.MemoryMB)
			if info.Diagnostics.ResourceMonitor.CheckIntervalSeconds > 0 {
				resourceMonitorInfo += fmt.Sprintf("Check Interval: %d seconds\nAlert Suppression: %d minutes\n",
					info.Diagnostics.ResourceMonitor.CheckIntervalSeconds,
					info.Diagnostics.ResourceMonitor.AlertSuppressionMinutes)
			}
		}
	}

	// Format mode status (FR-4.1.2)
	modeStatusInfo := ""
	if info.Diagnostics.ModeStatus != nil {
		modeStatusInfo = fmt.Sprintf("Current Mode: %s\n", info.Diagnostics.ModeStatus.CurrentMode)
		modeStatusInfo += fmt.Sprintf("Config Source: %s\n", info.Diagnostics.ModeStatus.ConfigSource)
		if info.Diagnostics.ModeStatus.ConsecutiveFailures > 0 {
			modeStatusInfo += fmt.Sprintf("Consecutive Failures: %d\n", info.Diagnostics.ModeStatus.ConsecutiveFailures)
		}
		if !info.Diagnostics.ModeStatus.LastSuccessTime.IsZero() {
			modeStatusInfo += fmt.Sprintf("Last Success: %s\n", info.Diagnostics.ModeStatus.LastSuccessTime.Format(time.RFC3339))
		}
		if !info.Diagnostics.ModeStatus.LastFailureTime.IsZero() {
			modeStatusInfo += fmt.Sprintf("Last Failure: %s\n", info.Diagnostics.ModeStatus.LastFailureTime.Format(time.RFC3339))
		}
	} else {
		modeStatusInfo = "Mode: registered (default)\n"
	}

	pretty := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════╗
║           Beacon Diagnostic Information                    ║
╚════════════════════════════════════════════════════════════╝

Timestamp: %s
Node ID: %s
Node Name: %s

─────────────────────────────────────────────────────────────
Operating Mode (FR-4.1.2)
─────────────────────────────────────────────────────────────
%s
─────────────────────────────────────────────────────────────
Network Status
─────────────────────────────────────────────────────────────
Pulse Server: %s
Reachable: %v
RTT (avg): %.2f ms
Packet Loss: %.1f%%

─────────────────────────────────────────────────────────────
Configuration
─────────────────────────────────────────────────────────────
Config File: %s
Config Valid: true
Log Level: %s
Debug Mode: %v

─────────────────────────────────────────────────────────────
Connection Status
─────────────────────────────────────────────────────────────
%s
─────────────────────────────────────────────────────────────
Resource Usage
─────────────────────────────────────────────────────────────
CPU: %.2f%%
Memory: %.2f MB (%.1f%%)

─────────────────────────────────────────────────────────────
Resource Monitor
─────────────────────────────────────────────────────────────
%s
─────────────────────────────────────────────────────────────
Probe Tasks
─────────────────────────────────────────────────────────────
%s
─────────────────────────────────────────────────────────────
Prometheus Metrics
─────────────────────────────────────────────────────────────
%s
`,
		info.Timestamp,
		info.NodeID,
		info.NodeName,
		modeStatusInfo,
		info.Diagnostics.NetworkStatus.PulseServerAddress,
		info.Diagnostics.NetworkStatus.PulseServerReachable,
		info.Diagnostics.NetworkStatus.RTTMs.Avg,
		info.Diagnostics.NetworkStatus.PacketLossRate*100,
		info.Diagnostics.Configuration.ConfigFile,
		info.Diagnostics.Configuration.LogLevel,
		info.Diagnostics.Configuration.DebugMode,
		connStatus,
		info.Diagnostics.ResourceUsage.CPUPercent,
		info.Diagnostics.ResourceUsage.MemoryMB,
		info.Diagnostics.ResourceUsage.MemoryPercent,
		resourceMonitorInfo,
		probeInfo,
		promInfo,
	)

	return pretty, nil
}

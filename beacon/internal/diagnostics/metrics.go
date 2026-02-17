package diagnostics

import (
	"beacon/internal/config"
)

// PrometheusMetrics contains Prometheus metrics summary
type PrometheusMetrics struct {
	BeaconUp       float64 `json:"beacon_up"`
	RTTSeconds     float64 `json:"beacon_rtt_seconds"`
	PacketLossRate float64 `json:"beacon_packet_loss_rate"`
	JitterMs       float64 `json:"beacon_jitter_ms"`

	// NFR-5.5.1 Enhanced metrics
	Mode            string  `json:"beacon_mode"`
	ConfigSource    string  `json:"beacon_config_source"`
	ActiveProbes    int     `json:"beacon_active_probes"`
	CompressionRatio float64 `json:"beacon_compression_ratio"`
	CacheSizeBytes  int64   `json:"beacon_cache_size_bytes"`
	CacheEvictions  int64   `json:"beacon_cache_evictions_total"`
}

// ModeProvider interface for getting mode information
type ModeProvider interface {
	GetMode() config.OperatingMode
	GetConfigSource() config.ConfigSource
}

// CacheStatsProvider interface for getting cache statistics
type CacheStatsProvider interface {
	Size() int64
	Count() int
	Evictions() int64
}

// CompressionStatsProvider interface for getting compression statistics
type CompressionStatsProvider interface {
	GetCompressionRatio() float64
}

// ProbeCountProvider interface for getting probe count
type ProbeCountProvider interface {
	GetProbeCount() int
}

// collectPrometheusMetrics collects Prometheus metrics summary
func (c *collector) collectPrometheusMetrics() (*PrometheusMetrics, error) {
	// Collect network status to derive basic metrics
	networkStatus, err := c.collectNetworkStatus()
	if err != nil {
		return nil, err
	}

	// Convert RTT from milliseconds to seconds for Prometheus format
	rttSeconds := networkStatus.RTTMs.Avg / 1000.0

	metrics := &PrometheusMetrics{
		BeaconUp:       1,                           // Beacon is running
		RTTSeconds:     rttSeconds,                  // Actual RTT from network check
		PacketLossRate: networkStatus.PacketLossRate, // Actual packet loss rate
		JitterMs:       0,                           // Requires multiple samples to calculate
	}

	// Collect enhanced metrics if providers are available
	c.enhanceMetrics(metrics)

	return metrics, nil
}

// enhanceMetrics adds enhanced metrics from various providers
func (c *collector) enhanceMetrics(metrics *PrometheusMetrics) {
	// Get mode info if provider is available
	if c.modeProvider != nil {
		metrics.Mode = string(c.modeProvider.GetMode())
		metrics.ConfigSource = string(c.modeProvider.GetConfigSource())
	} else {
		// Default values
		metrics.Mode = string(config.ModeRegistered)
		metrics.ConfigSource = string(config.SourceLocal)
	}

	// Get probe count if provider is available
	if c.probeCountProvider != nil {
		metrics.ActiveProbes = c.probeCountProvider.GetProbeCount()
	}

	// Get cache stats if provider is available
	if c.cacheStatsProvider != nil {
		metrics.CacheSizeBytes = c.cacheStatsProvider.Size()
		metrics.CacheEvictions = c.cacheStatsProvider.Evictions()
	}

	// Get compression ratio if provider is available
	if c.compressionProvider != nil {
		metrics.CompressionRatio = c.compressionProvider.GetCompressionRatio()
	}
}

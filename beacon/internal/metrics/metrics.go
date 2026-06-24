package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/whg517/node-pulse/beacon/internal/config"
	"github.com/whg517/node-pulse/beacon/internal/logger"
	"github.com/whg517/node-pulse/beacon/internal/probe"
)

// ModeProvider interface for getting current operating mode
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

// Metrics handles Prometheus metrics exposure
type Metrics struct {
	config    *config.Config
	scheduler *probe.ProbeScheduler

	// External providers for enhanced metrics
	modeProvider        ModeProvider
	cacheStatsProvider  CacheStatsProvider
	compressionProvider CompressionStatsProvider

	// Prometheus metrics - Basic
	beaconUp         *prometheus.GaugeVec
	beaconRTTSeconds *prometheus.GaugeVec
	beaconPacketLoss *prometheus.GaugeVec
	beaconJitterMs   *prometheus.GaugeVec

	// Prometheus metrics - NFR-5.5.1 Enhanced
	beaconMode            *prometheus.GaugeVec   // Current operating mode
	beaconConfigSource    *prometheus.GaugeVec   // Config source (local/server/cached)
	beaconConfigVersion   prometheus.Gauge       // Config version number
	beaconActiveProbes    prometheus.Gauge       // Number of active probes
	beaconCompressionRatio prometheus.Gauge      // Compression ratio
	beaconCacheSizeBytes  prometheus.Gauge       // Cache size in bytes
	beaconCacheEvictions  prometheus.Counter     // Total cache evictions
	beaconCompressionCorruption prometheus.Counter // Compression corruption count

	// Probe metrics (NFR-5.5.1)
	beaconProbeTotal        *prometheus.CounterVec   // Total probes by type
	beaconProbeDuration     *prometheus.HistogramVec // Probe duration by type
	beaconProbeFailureTotal *prometheus.CounterVec   // Probe failures by type and error

	// Resource metrics (NFR-5.5.1)
	beaconMemoryUsageBytes prometheus.Gauge // Memory usage in bytes
	beaconCPUUsagePercent  prometheus.Gauge // CPU usage percentage

	// Resume metrics (FR-4.1.5, NFR-5.5.1)
	beaconResumeUploadBytesTotal prometheus.Counter // Total bytes uploaded via resume

	registry *prometheus.Registry
	server   *http.Server

	mu          sync.RWMutex
	running     bool
	stopChan    chan struct{}
	collectorWg sync.WaitGroup
}

// NewMetrics creates a new Metrics handler
func NewMetrics(cfg *config.Config, scheduler *probe.ProbeScheduler) *Metrics {
	registry := prometheus.NewRegistry()

	// Basic metrics
	beaconUp := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "beacon_up",
			Help: "Beacon running status (1=running, 0=stopped)",
		},
		[]string{"node_id", "node_name"},
	)

	beaconRTTSeconds := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "beacon_rtt_seconds",
			Help: "Latest RTT latency in seconds",
		},
		[]string{"node_id", "node_name"},
	)

	beaconPacketLoss := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "beacon_packet_loss_rate",
			Help: "Latest packet loss rate (0-1)",
		},
		[]string{"node_id", "node_name"},
	)

	beaconJitterMs := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "beacon_jitter_ms",
			Help: "Latest jitter in milliseconds",
		},
		[]string{"node_id", "node_name"},
	)

	// NFR-5.5.1 Enhanced metrics
	beaconMode := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "beacon_mode",
			Help: "Current operating mode (0=standalone, 1=registered, 2=degraded)",
		},
		[]string{"node_id", "node_name", "mode"},
	)

	beaconConfigSource := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "beacon_config_source",
			Help: "Current config source (0=local, 1=server, 2=cached)",
		},
		[]string{"node_id", "node_name", "source"},
	)

	beaconActiveProbes := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "beacon_active_probes",
			Help: "Number of active probe tasks",
		},
	)

	beaconCompressionRatio := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "beacon_compression_ratio",
			Help: "Average compression ratio (0-100 percent)",
		},
	)

	beaconCacheSizeBytes := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "beacon_cache_size_bytes",
			Help: "Current cache size in bytes",
		},
	)

	beaconCacheEvictions := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "beacon_cache_evictions_total",
			Help: "Total number of cache evictions",
		},
	)

	beaconCompressionCorruption := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "beacon_compression_corruption_total",
			Help: "Total number of compression corruption detections",
		},
	)

	// Config version metric
	beaconConfigVersion := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "beacon_config_version",
			Help: "Current configuration version number",
		},
	)

	// Probe metrics
	beaconProbeTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "beacon_probe_total",
			Help: "Total number of probes executed by type",
		},
		[]string{"probe_type"},
	)

	beaconProbeDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "beacon_probe_duration_seconds",
			Help:    "Probe execution duration in seconds by type",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"probe_type"},
	)

	beaconProbeFailureTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "beacon_probe_failure_total",
			Help: "Total number of probe failures by type and error",
		},
		[]string{"probe_type", "error_type"},
	)

	// Resource metrics
	beaconMemoryUsageBytes := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "beacon_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)

	beaconCPUUsagePercent := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "beacon_cpu_usage_percent",
			Help: "Current CPU usage percentage",
		},
	)

	// Resume upload metric
	beaconResumeUploadBytesTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "beacon_resume_upload_bytes_total",
			Help: "Total bytes uploaded via resume after reconnection",
		},
	)

	// Register all metrics
	registry.MustRegister(beaconUp)
	registry.MustRegister(beaconRTTSeconds)
	registry.MustRegister(beaconPacketLoss)
	registry.MustRegister(beaconJitterMs)
	registry.MustRegister(beaconMode)
	registry.MustRegister(beaconConfigSource)
	registry.MustRegister(beaconConfigVersion)
	registry.MustRegister(beaconActiveProbes)
	registry.MustRegister(beaconCompressionRatio)
	registry.MustRegister(beaconCacheSizeBytes)
	registry.MustRegister(beaconCacheEvictions)
	registry.MustRegister(beaconCompressionCorruption)
	registry.MustRegister(beaconProbeTotal)
	registry.MustRegister(beaconProbeDuration)
	registry.MustRegister(beaconProbeFailureTotal)
	registry.MustRegister(beaconMemoryUsageBytes)
	registry.MustRegister(beaconCPUUsagePercent)
	registry.MustRegister(beaconResumeUploadBytesTotal)

	return &Metrics{
		config:                   cfg,
		scheduler:                scheduler,
		beaconUp:                 beaconUp,
		beaconRTTSeconds:         beaconRTTSeconds,
		beaconPacketLoss:         beaconPacketLoss,
		beaconJitterMs:           beaconJitterMs,
		beaconMode:               beaconMode,
		beaconConfigSource:       beaconConfigSource,
		beaconConfigVersion:      beaconConfigVersion,
		beaconActiveProbes:       beaconActiveProbes,
		beaconCompressionRatio:   beaconCompressionRatio,
		beaconCacheSizeBytes:     beaconCacheSizeBytes,
		beaconCacheEvictions:     beaconCacheEvictions,
		beaconCompressionCorruption: beaconCompressionCorruption,
		beaconProbeTotal:         beaconProbeTotal,
		beaconProbeDuration:      beaconProbeDuration,
		beaconProbeFailureTotal:  beaconProbeFailureTotal,
		beaconMemoryUsageBytes:   beaconMemoryUsageBytes,
		beaconCPUUsagePercent:    beaconCPUUsagePercent,
		beaconResumeUploadBytesTotal: beaconResumeUploadBytesTotal,
		registry:                 registry,
		stopChan:                 make(chan struct{}),
	}
}

// SetModeProvider sets the mode provider for enhanced metrics
func (m *Metrics) SetModeProvider(provider ModeProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modeProvider = provider
}

// SetCacheStatsProvider sets the cache stats provider for enhanced metrics
func (m *Metrics) SetCacheStatsProvider(provider CacheStatsProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheStatsProvider = provider
}

// SetCompressionStatsProvider sets the compression stats provider for enhanced metrics
func (m *Metrics) SetCompressionStatsProvider(provider CompressionStatsProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compressionProvider = provider
}

// Start starts the metrics server
func (m *Metrics) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.MetricsEnabled {
		logger.WithField("component", "metrics").Info("Metrics server disabled in configuration")
		return nil
	}

	if m.running {
		return fmt.Errorf("metrics server already running")
	}

	// Set beacon_up to 1 (running)
	m.beaconUp.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(1)

	// Initialize basic metrics with default values
	m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
	m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
	m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)

	// Initialize enhanced metrics
	m.initializeEnhancedMetrics()

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))

	addr := fmt.Sprintf(":%d", m.config.MetricsPort)
	m.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Protection against Slowloris attacks
	}

	// Create new stopChan for this start cycle
	m.stopChan = make(chan struct{})

	// Start server in goroutine with error channel
	serverErrChan := make(chan error, 1)
	go func() {
		logger.WithFields(map[string]interface{}{"component": "metrics", "address": addr}).Info("Starting Prometheus metrics server")
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithFields(map[string]interface{}{"component": "metrics", "error": err.Error()}).Error("Metrics server error")
			serverErrChan <- err
		}
	}()

	// Give server a moment to start and check for immediate errors
	select {
	case err := <-serverErrChan:
		// Fix #5: Clean up stopChan on failed start
		close(m.stopChan)
		return fmt.Errorf("failed to start metrics server: %w", err)
	case <-time.After(100 * time.Millisecond):
		// Server started successfully
	}

	// Start metrics collector with WaitGroup
	m.collectorWg.Add(1)
	go m.collectMetrics()

	m.running = true
	logger.WithField("component", "metrics").Info("Prometheus metrics server started successfully")
	return nil
}

// initializeEnhancedMetrics initializes the enhanced metrics with default values
func (m *Metrics) initializeEnhancedMetrics() {
	// Initialize mode metric (default to registered mode if not set)
	mode := config.ModeRegistered
	if m.modeProvider != nil {
		mode = m.modeProvider.GetMode()
	}
	m.updateModeMetric(mode)

	// Initialize config source metric (default to local)
	source := config.SourceLocal
	if m.modeProvider != nil {
		source = m.modeProvider.GetConfigSource()
	}
	m.updateConfigSourceMetric(source)

	// Initialize active probes
	if m.scheduler != nil {
		m.beaconActiveProbes.Set(float64(m.scheduler.GetProbeCount()))
	} else {
		m.beaconActiveProbes.Set(0)
	}

	// Initialize cache and compression metrics
	m.beaconCompressionRatio.Set(0)
	m.beaconCacheSizeBytes.Set(0)
}

// updateModeMetric updates the beacon_mode metric
func (m *Metrics) updateModeMetric(mode config.OperatingMode) {
	// Reset all modes to 0
	m.beaconMode.WithLabelValues(m.config.NodeID, m.config.NodeName, "standalone").Set(0)
	m.beaconMode.WithLabelValues(m.config.NodeID, m.config.NodeName, "registered").Set(0)
	m.beaconMode.WithLabelValues(m.config.NodeID, m.config.NodeName, "degraded").Set(0)

	// Set current mode to 1
	m.beaconMode.WithLabelValues(m.config.NodeID, m.config.NodeName, string(mode)).Set(1)
}

// updateConfigSourceMetric updates the beacon_config_source metric
func (m *Metrics) updateConfigSourceMetric(source config.ConfigSource) {
	// Reset all sources to 0
	m.beaconConfigSource.WithLabelValues(m.config.NodeID, m.config.NodeName, "local").Set(0)
	m.beaconConfigSource.WithLabelValues(m.config.NodeID, m.config.NodeName, "server").Set(0)
	m.beaconConfigSource.WithLabelValues(m.config.NodeID, m.config.NodeName, "cached").Set(0)

	// Set current source to 1
	m.beaconConfigSource.WithLabelValues(m.config.NodeID, m.config.NodeName, string(source)).Set(1)
}

// Stop stops the metrics server
func (m *Metrics) Stop() error {
	m.mu.Lock()

	if !m.running {
		m.mu.Unlock()
		return nil
	}

	// Mark as not running first
	m.running = false

	// Set beacon_up to 0 (stopped)
	m.beaconUp.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)

	// Stop metrics collector by closing channel
	close(m.stopChan)

	m.mu.Unlock()

	// Wait for collector goroutine to finish (outside lock to prevent deadlock)
	m.collectorWg.Wait()

	// Shutdown HTTP server gracefully if it exists
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.server.Shutdown(ctx); err != nil {
			logger.WithFields(map[string]interface{}{"component": "metrics", "error": err.Error()}).Error("Metrics server shutdown error")
			return err
		}
	}

	logger.WithField("component", "metrics").Info("Prometheus metrics server stopped")
	return nil
}

// collectMetrics periodically updates Prometheus metrics from probe results
func (m *Metrics) collectMetrics() {
	defer m.collectorWg.Done()

	// Fix #4: Use configurable update interval from config
	updateInterval := time.Duration(m.config.MetricsUpdateSeconds) * time.Second
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateMetrics()
		case <-m.stopChan:
			return
		}
	}
}

// updateMetrics updates Prometheus metrics from latest probe results
func (m *Metrics) updateMetrics() {
	// Get latest probe results from scheduler
	tcpResults, udpResults := m.scheduler.GetLatestResults()

	totalResults := len(tcpResults) + len(udpResults)
	if totalResults == 0 {
		// No probe results, set metrics to indicate no data
		m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
		m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(1) // 100% loss
		m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
	} else {
		// Aggregate metrics from all probe results (average)
		var totalRTT, totalPacketLoss, totalJitter float64
		count := 0

		// Process TCP probe results
		for _, result := range tcpResults {
			if result != nil && result.Success {
				totalRTT += result.RTTMs
				totalPacketLoss += result.PacketLossRate
				totalJitter += result.JitterMs
				count++
			}
		}

		// Process UDP probe results
		for _, result := range udpResults {
			if result != nil && result.Success {
				totalRTT += result.RTTMs
				totalPacketLoss += result.PacketLossRate
				totalJitter += result.JitterMs
				count++
			}
		}

		if count > 0 {
			// Convert RTT from milliseconds to seconds for Prometheus best practices
			rttSeconds := (totalRTT / float64(count)) / 1000.0
			// Convert packet loss rate from percentage (0-100) to ratio (0-1)
			packetLossRate := (totalPacketLoss / float64(count)) / 100.0
			jitterMs := totalJitter / float64(count)

			m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(rttSeconds)
			m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(packetLossRate)
			m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(jitterMs)
		} else {
			// All probes failed
			m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
			m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(1) // 100% loss
			m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
		}
	}

	// Update enhanced metrics
	m.updateEnhancedMetrics()
}

// updateEnhancedMetrics updates the NFR-5.5.1 enhanced metrics
func (m *Metrics) updateEnhancedMetrics() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Update mode and config source
	if m.modeProvider != nil {
		m.updateModeMetric(m.modeProvider.GetMode())
		m.updateConfigSourceMetric(m.modeProvider.GetConfigSource())
	}

	// Update active probes
	if m.scheduler != nil {
		m.beaconActiveProbes.Set(float64(m.scheduler.GetProbeCount()))
	}

	// Update cache stats
	if m.cacheStatsProvider != nil {
		m.beaconCacheSizeBytes.Set(float64(m.cacheStatsProvider.Size()))

		// Note: Evictions are tracked via counter increments, not here
		// The counter should be incremented by the cache when evictions occur
	}

	// Update compression ratio
	if m.compressionProvider != nil {
		ratio := m.compressionProvider.GetCompressionRatio()
		m.beaconCompressionRatio.Set(ratio)
	}
}

// RecordCacheEviction records a cache eviction event
func (m *Metrics) RecordCacheEviction() {
	m.beaconCacheEvictions.Inc()
}

// RecordCompressionCorruption records a compression corruption event
func (m *Metrics) RecordCompressionCorruption() {
	m.beaconCompressionCorruption.Inc()
}

// IsRunning returns whether the metrics server is running
func (m *Metrics) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// UpdateMode updates the mode metric externally
func (m *Metrics) UpdateMode(mode config.OperatingMode) {
	m.updateModeMetric(mode)
}

// UpdateConfigSource updates the config source metric externally
func (m *Metrics) UpdateConfigSource(source config.ConfigSource) {
	m.updateConfigSourceMetric(source)
}

// UpdateActiveProbes updates the active probes metric
func (m *Metrics) UpdateActiveProbes(count int) {
	m.beaconActiveProbes.Set(float64(count))
}

// UpdateCompressionRatio updates the compression ratio metric
func (m *Metrics) UpdateCompressionRatio(ratio float64) {
	m.beaconCompressionRatio.Set(ratio)
}

// UpdateCacheSize updates the cache size metric
func (m *Metrics) UpdateCacheSize(sizeBytes int64) {
	m.beaconCacheSizeBytes.Set(float64(sizeBytes))
}

// UpdateConfigVersion updates the config version metric
func (m *Metrics) UpdateConfigVersion(version int64) {
	m.beaconConfigVersion.Set(float64(version))
}

// RecordProbe records a probe execution
func (m *Metrics) RecordProbe(probeType string, durationSeconds float64) {
	m.beaconProbeTotal.WithLabelValues(probeType).Inc()
	m.beaconProbeDuration.WithLabelValues(probeType).Observe(durationSeconds)
}

// RecordProbeFailure records a probe failure
func (m *Metrics) RecordProbeFailure(probeType string, errorType string) {
	m.beaconProbeFailureTotal.WithLabelValues(probeType, errorType).Inc()
}

// UpdateMemoryUsage updates the memory usage metric
func (m *Metrics) UpdateMemoryUsage(bytes int64) {
	m.beaconMemoryUsageBytes.Set(float64(bytes))
}

// UpdateCPUUsage updates the CPU usage metric
func (m *Metrics) UpdateCPUUsage(percent float64) {
	m.beaconCPUUsagePercent.Set(percent)
}

// RecordResumeUpload records bytes uploaded via resume
func (m *Metrics) RecordResumeUpload(bytes int64) {
	m.beaconResumeUploadBytesTotal.Add(float64(bytes))
}

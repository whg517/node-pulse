package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/probe"
)

// Metrics handles Prometheus metrics exposure
type Metrics struct {
	config    *config.Config
	scheduler *probe.ProbeScheduler

	// Prometheus metrics
	beaconUp         *prometheus.GaugeVec
	beaconRTTSeconds *prometheus.GaugeVec
	beaconPacketLoss *prometheus.GaugeVec
	beaconJitterMs   *prometheus.GaugeVec

	registry *prometheus.Registry
	server   *http.Server

	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
	collectorWg  sync.WaitGroup
}

// NewMetrics creates a new Metrics handler
func NewMetrics(cfg *config.Config, scheduler *probe.ProbeScheduler) *Metrics {
	registry := prometheus.NewRegistry()

	// Define Prometheus metrics with labels (node_id, node_name)
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

	// Register metrics
	registry.MustRegister(beaconUp)
	registry.MustRegister(beaconRTTSeconds)
	registry.MustRegister(beaconPacketLoss)
	registry.MustRegister(beaconJitterMs)

	return &Metrics{
		config:           cfg,
		scheduler:        scheduler,
		beaconUp:         beaconUp,
		beaconRTTSeconds: beaconRTTSeconds,
		beaconPacketLoss: beaconPacketLoss,
		beaconJitterMs:   beaconJitterMs,
		registry:         registry,
		stopChan:         make(chan struct{}),
	}
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

	// Initialize other metrics with default values (0) so they appear in /metrics
	m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
	m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
	m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))

	addr := fmt.Sprintf(":%d", m.config.MetricsPort)
	m.server = &http.Server{
		Addr:    addr,
		Handler: mux,
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
		return
	}

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

// IsRunning returns whether the metrics server is running
func (m *Metrics) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}


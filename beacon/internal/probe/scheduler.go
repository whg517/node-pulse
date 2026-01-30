package probe

import (
	"fmt"
	"net"
	"sync"
	"time"

	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/models"
)

// ProbeScheduler manages and executes multiple probes
type ProbeScheduler struct {
	tcpPingers    []*TCPPinger
	udpPingers    []*UDPPinger
	interval      time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	running       bool
	mu            sync.RWMutex
	// Cache latest results for heartbeat reporting
	latestTCPResults []*models.TCPProbeResult
	latestUDPResults []*models.UDPProbeResult
	resultsMu        sync.RWMutex
}

// NewProbeScheduler creates a new probe scheduler from configuration
func NewProbeScheduler(probeConfigs []config.ProbeConfig) (*ProbeScheduler, error) {
	scheduler := &ProbeScheduler{
		tcpPingers: make([]*TCPPinger, 0),
		udpPingers: make([]*UDPPinger, 0),
		stopChan:   make(chan struct{}),
		running:    false,
	}

	// Initialize TCP and UDP pingers from config
	for _, cfg := range probeConfigs {
		if cfg.Type == "tcp_ping" {
			tcpConfig := TCPProbeConfig{
				Type:           cfg.Type,
				Target:         cfg.Target,
				Port:           cfg.Port,
				TimeoutSeconds: cfg.TimeoutSeconds,
				Interval:       cfg.Interval,
				Count:          cfg.Count,
			}

			// Validate configuration (includes count ≥ 10 check)
			if err := tcpConfig.Validate(); err != nil {
				return nil, fmt.Errorf("invalid probe config for %s:%d: %w", cfg.Target, cfg.Port, err)
			}

			// Additional count ≥ 10 validation for core metrics
			if cfg.Count < 10 {
				return nil, fmt.Errorf("probe count for %s must be ≥ 10 to calculate core metrics (current: %d)", cfg.Target, cfg.Count)
			}

			pinger := NewTCPPinger(tcpConfig)
			scheduler.tcpPingers = append(scheduler.tcpPingers, pinger)
		} else if cfg.Type == "udp_ping" {
			udpConfig := UDPProbeConfig{
				Type:           cfg.Type,
				Target:         cfg.Target,
				Port:           cfg.Port,
				TimeoutSeconds: cfg.TimeoutSeconds,
				Interval:       cfg.Interval,
				Count:          cfg.Count,
			}

			// Validate configuration (includes count ≥ 10 check)
			if err := udpConfig.Validate(); err != nil {
				return nil, fmt.Errorf("invalid probe config for %s:%d: %w", cfg.Target, cfg.Port, err)
			}

			// Additional count ≥ 10 validation for core metrics
			if cfg.Count < 10 {
				return nil, fmt.Errorf("probe count for %s must be ≥ 10 to calculate core metrics (current: %d)", cfg.Target, cfg.Count)
			}

			pinger := NewUDPPinger(udpConfig)
			scheduler.udpPingers = append(scheduler.udpPingers, pinger)
		}
	}

	return scheduler, nil
}

// Start begins the probe scheduling loop
func (s *ProbeScheduler) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	// Determine interval from first probe config (all probes share the same interval in MVP)
	totalProbes := len(s.tcpPingers) + len(s.udpPingers)
	if totalProbes == 0 {
		logger.Info("No probes configured, scheduler started but will not execute any probes")
		return nil
	}

	interval := time.Duration(60) * time.Second // Default interval
	if len(s.tcpPingers) > 0 {
		interval = time.Duration(s.tcpPingers[0].config.Interval) * time.Second
	} else if len(s.udpPingers) > 0 {
		interval = time.Duration(s.udpPingers[0].config.Interval) * time.Second
	}
	s.interval = interval

	logger.WithFields(map[string]interface{}{
		"component":  "probe",
		"interval":   interval.String(),
		"tcp_count":  len(s.tcpPingers),
		"udp_count":  len(s.udpPingers),
	}).Info("Probe scheduler started")

	// Start scheduling loop in background
	s.wg.Add(1)
	go s.run()

	return nil
}

// run executes the scheduling loop
func (s *ProbeScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Execute probes immediately on start
	s.executeProbes()

	for {
		select {
		case <-ticker.C:
			s.executeProbes()
		case <-s.stopChan:
			logger.WithField("component", "probe").Info("Probe scheduler stopping...")
			return
		}
	}
}

// executeProbes runs all probes concurrently
func (s *ProbeScheduler) executeProbes() {
	logger.WithFields(map[string]interface{}{"component": "probe", "tcp_count": len(s.tcpPingers), "udp_count": len(s.udpPingers)}).Info("Executing probes...")

	var wg sync.WaitGroup

	// Temporary storage for results
	tcpResults := make([]*models.TCPProbeResult, len(s.tcpPingers))
	udpResults := make([]*models.UDPProbeResult, len(s.udpPingers))

	// Execute TCP probes
	for i, pinger := range s.tcpPingers {
		wg.Add(1)
		go func(index int, p *TCPPinger) {
			defer wg.Done()

			// Build target address with IPv6 support
			target := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))
			logger.WithFields(map[string]interface{}{"component": "probe", "probe_type": "tcp_ping", "target": target, "count": p.config.Count}).Debug("Starting TCP probe")

			// Execute batch probes with core metrics calculation
			result, err := p.ExecuteBatch(p.config.Count)
			if err != nil {
				logger.WithFields(map[string]interface{}{"component": "probe", "probe_type": "tcp_ping", "target": target, "error": err}).Error("TCP probe failed")
				return
			}

			// Store result for heartbeat reporting
			tcpResults[index] = result

			// Log core metrics
			logger.WithFields(map[string]interface{}{
				"component":      "probe",
				"probe_type":     "tcp_ping",
				"target":         target,
				"success":        result.Success,
				"sample_count":   result.SampleCount,
				"rtt_ms":         result.RTTMs,
				"rtt_median_ms":  result.RTTMedianMs,
				"jitter_ms":      result.JitterMs,
				"variance_ms":    result.VarianceMs,
				"packet_loss":    result.PacketLossRate,
				"timestamp":      result.Timestamp,
			}).Info("TCP probe completed")
		}(i, pinger)
	}

	// Execute UDP probes
	for i, pinger := range s.udpPingers {
		wg.Add(1)
		go func(index int, p *UDPPinger) {
			defer wg.Done()

			// Build target address with IPv6 support
			target := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))
			logger.WithFields(map[string]interface{}{"component": "probe", "probe_type": "udp_ping", "target": target, "count": p.config.Count}).Debug("Starting UDP probe")

			// Execute batch probes with core metrics calculation
			result, err := p.ExecuteBatch(p.config.Count)
			if err != nil {
				logger.WithFields(map[string]interface{}{"component": "probe", "probe_type": "udp_ping", "target": target, "error": err}).Error("UDP probe failed")
				return
			}

			// Store result for heartbeat reporting
			udpResults[index] = result

			// Log core metrics
			logger.WithFields(map[string]interface{}{
				"component":       "probe",
				"probe_type":      "udp_ping",
				"target":          target,
				"success":         result.Success,
				"sample_count":    result.SampleCount,
				"sent":            result.SentPackets,
				"received":        result.ReceivedPackets,
				"rtt_ms":          result.RTTMs,
				"rtt_median_ms":   result.RTTMedianMs,
				"jitter_ms":       result.JitterMs,
				"variance_ms":     result.VarianceMs,
				"packet_loss":     result.PacketLossRate,
				"timestamp":       result.Timestamp,
			}).Info("UDP probe completed")
		}(i, pinger)
	}

	wg.Wait()

	// Update cached results
	s.resultsMu.Lock()
	s.latestTCPResults = tcpResults
	s.latestUDPResults = udpResults
	s.resultsMu.Unlock()

	logger.WithField("component", "probe").Info("All probes completed")
}

// Stop gracefully stops the scheduler
func (s *ProbeScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	logger.WithField("component", "probe").Info("Stopping probe scheduler...")
	close(s.stopChan)
	s.wg.Wait()
	logger.WithField("component", "probe").Info("Probe scheduler stopped")
}

// IsRunning returns whether the scheduler is running
func (s *ProbeScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetProbeCount returns the number of configured probes
func (s *ProbeScheduler) GetProbeCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tcpPingers) + len(s.udpPingers)
}

// ExecuteProbeNow executes a specific probe immediately (for testing or manual trigger)
func (s *ProbeScheduler) ExecuteProbeNow(index int) (*models.TCPProbeResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index < 0 || index >= len(s.tcpPingers) {
		return nil, fmt.Errorf("probe index %d out of range [0, %d]", index, len(s.tcpPingers)-1)
	}

	pinger := s.tcpPingers[index]
	return pinger.Execute()
}

// GetLatestResults returns the most recent probe results for heartbeat reporting
func (s *ProbeScheduler) GetLatestResults() ([]*models.TCPProbeResult, []*models.UDPProbeResult) {
	s.resultsMu.RLock()
	defer s.resultsMu.RUnlock()

	// Return copies to avoid concurrent access issues
	tcpCopy := make([]*models.TCPProbeResult, len(s.latestTCPResults))
	udpCopy := make([]*models.UDPProbeResult, len(s.latestUDPResults))

	copy(tcpCopy, s.latestTCPResults)
	copy(udpCopy, s.latestUDPResults)

	return tcpCopy, udpCopy
}

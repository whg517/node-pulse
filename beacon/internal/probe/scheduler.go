package probe

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"beacon/internal/config"
	"beacon/internal/models"
)

// ProbeScheduler manages and executes multiple probes
type ProbeScheduler struct {
	tcpPingers []*TCPPinger
	udpPingers []*UDPPinger
	interval   time.Duration
	stopChan   chan struct{}
	wg         sync.WaitGroup
	running    bool
	mu         sync.RWMutex
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
		log.Println("[INFO] No probes configured, scheduler started but will not execute any probes")
		return nil
	}

	interval := time.Duration(60) * time.Second // Default interval
	if len(s.tcpPingers) > 0 {
		interval = time.Duration(s.tcpPingers[0].config.Interval) * time.Second
	} else if len(s.udpPingers) > 0 {
		interval = time.Duration(s.udpPingers[0].config.Interval) * time.Second
	}
	s.interval = interval

	log.Printf("[INFO] Probe scheduler started with interval: %v (%d TCP, %d UDP)", interval, len(s.tcpPingers), len(s.udpPingers))

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
			log.Println("[INFO] Probe scheduler stopping...")
			return
		}
	}
}

// executeProbes runs all probes concurrently
func (s *ProbeScheduler) executeProbes() {
	log.Printf("[INFO] Executing %d TCP and %d UDP probes...", len(s.tcpPingers), len(s.udpPingers))

	var wg sync.WaitGroup

	// Execute TCP probes
	for i, pinger := range s.tcpPingers {
		wg.Add(1)
		go func(index int, p *TCPPinger) {
			defer wg.Done()

			// Build target address with IPv6 support
			target := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))
			log.Printf("[DEBUG] Starting TCP probe #%d to %s (count=%d)", index+1, target, p.config.Count)

			// Execute batch probes with core metrics calculation
			result, err := p.ExecuteBatch(p.config.Count)
			if err != nil {
				log.Printf("[ERROR] TCP probe #%d to %s failed: %v", index+1, target, err)
				return
			}

			// Log core metrics
			successStatus := "failed"
			if result.Success {
				successStatus = "succeeded"
			}

			log.Printf("[INFO] TCP probe #%d to %s completed: %s, samples=%d, RTT=%.2f ms (median=%.2f ms), jitter=%.2f ms, variance=%.2f ms², packet loss=%.2f%%, timestamp=%s",
				index+1, target, successStatus, result.SampleCount, result.RTTMs, result.RTTMedianMs,
				result.JitterMs, result.VarianceMs, result.PacketLossRate, result.Timestamp)

			// TODO: Report results to Pulse (Story 3.7)
		}(i, pinger)
	}

	// Execute UDP probes
	for i, pinger := range s.udpPingers {
		wg.Add(1)
		go func(index int, p *UDPPinger) {
			defer wg.Done()

			// Build target address with IPv6 support
			target := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))
			log.Printf("[DEBUG] Starting UDP probe #%d to %s (count=%d)", index+1, target, p.config.Count)

			// Execute batch probes with core metrics calculation
			result, err := p.ExecuteBatch(p.config.Count)
			if err != nil {
				log.Printf("[ERROR] UDP probe #%d to %s failed: %v", index+1, target, err)
				return
			}

			// Log core metrics
			successStatus := "failed"
			if result.Success {
				successStatus = "succeeded"
			}

			log.Printf("[INFO] UDP probe #%d to %s completed: %s, samples=%d, sent=%d, received=%d, RTT=%.2f ms (median=%.2f ms), jitter=%.2f ms, variance=%.2f ms², packet loss=%.2f%%, timestamp=%s",
				index+1, target, successStatus, result.SampleCount, result.SentPackets, result.ReceivedPackets,
				result.RTTMs, result.RTTMedianMs, result.JitterMs, result.VarianceMs, result.PacketLossRate, result.Timestamp)

			// TODO: Report results to Pulse (Story 3.7)
		}(i, pinger)
	}

	wg.Wait()
	log.Printf("[INFO] All probes completed")
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

	log.Println("[INFO] Stopping probe scheduler...")
	close(s.stopChan)
	s.wg.Wait()
	log.Println("[INFO] Probe scheduler stopped")
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

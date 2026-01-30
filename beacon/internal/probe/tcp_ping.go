package probe

import (
	"fmt"
	"math"
	"net"
	"time"

	"beacon/internal/models"
)

const (
	// rttPrecisionMultiplier is the multiplier for RTT precision (2 decimal places)
	rttPrecisionMultiplier = 100
)

// TCPProbeConfig represents TCP probe configuration
type TCPProbeConfig struct {
	Type           string `yaml:"type" validate:"required,eq=tcp_ping"`
	Target         string `yaml:"target" validate:"required,ip|hostname"`
	Port           int    `yaml:"port" validate:"required,min=1,max=65535"`
	TimeoutSeconds int    `yaml:"timeout" validate:"required,min=1,max=30"`
	Interval       int    `yaml:"interval" validate:"required,min=60,max=300"`
	Count          int    `yaml:"count" validate:"required,min=1,max=100"`
}

// Validate validates the TCP probe configuration
func (c *TCPProbeConfig) Validate() error {
	if c.Type != "tcp_ping" {
		return fmt.Errorf("invalid probe type: %s (must be 'tcp_ping')", c.Type)
	}

	if c.Target == "" {
		return fmt.Errorf("probe target cannot be empty")
	}

	// Validate target is IP or hostname
	if net.ParseIP(c.Target) == nil {
		// Not an IP, check if it's a valid hostname format
		if err := validateHostname(c.Target); err != nil {
			return fmt.Errorf("invalid probe target '%s': %w", c.Target, err)
		}
	}

	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port %d, must be between 1 and 65535", c.Port)
	}

	// Allow timeout=0 (will use default of 5 seconds in Execute)
	if c.TimeoutSeconds < 0 || c.TimeoutSeconds > 30 {
		return fmt.Errorf("invalid timeout %d, must be between 0 and 30 seconds (0 = default 5s)", c.TimeoutSeconds)
	}

	if c.Interval < 60 || c.Interval > 300 {
		return fmt.Errorf("invalid interval %d, must be between 60 and 300 seconds", c.Interval)
	}

	if c.Count < 1 || c.Count > 100 {
		return fmt.Errorf("invalid count %d, must be between 1 and 100", c.Count)
	}

	return nil
}

// validateHostname validates hostname format (basic validation)
func validateHostname(hostname string) error {
	if len(hostname) == 0 || len(hostname) > 253 {
		return fmt.Errorf("invalid hostname length")
	}
	return nil
}

// TCPPinger represents a TCP ping probe engine
type TCPPinger struct {
	config TCPProbeConfig
}

// NewTCPPinger creates a new TCP pinger with the given configuration
func NewTCPPinger(config TCPProbeConfig) *TCPPinger {
	return &TCPPinger{
		config: config,
	}
}

// Execute performs a single TCP probe
func (p *TCPPinger) Execute() (*models.TCPProbeResult, error) {
	// Set default timeout if not configured (5 seconds per AC#3)
	timeout := p.config.TimeoutSeconds
	if timeout == 0 {
		timeout = 5
	}

	// Validate configuration before executing
	if err := p.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	startTime := time.Now()

	// Build target address with IPv6 support
	targetAddr := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))

	// Attempt TCP connection with timeout
	conn, err := net.DialTimeout("tcp", targetAddr, time.Duration(timeout)*time.Second)

	elapsed := time.Since(startTime)

	if err != nil {
		// Connection failed
		return models.NewTCPProbeResult(false, 0, err.Error()), nil
	}

	// Connection succeeded, close immediately
	conn.Close()

	// Calculate RTT in milliseconds (rounded to 2 decimal places)
	rttMs := math.Round(elapsed.Seconds()*1000*rttPrecisionMultiplier) / rttPrecisionMultiplier

	return models.NewTCPProbeResult(true, rttMs, ""), nil
}

// ExecuteBatch performs multiple TCP probes and calculates core metrics
func (p *TCPPinger) ExecuteBatch(count int) (*models.TCPProbeResult, error) {
	if count < 1 || count > 100 {
		return nil, fmt.Errorf("invalid count %d, must be between 1 and 100", count)
	}

	samples := make([]SamplePoint, 0, count)
	sentPackets := 0
	receivedPackets := 0
	var errors []string

	collector := NewCoreMetricsCollector()

	for i := 0; i < count; i++ {
		sentPackets++
		startTime := time.Now()

		// Validate configuration before executing
		if err := p.config.Validate(); err != nil {
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Build target address with IPv6 support
		targetAddr := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))

		// Attempt TCP connection with timeout
		timeout := p.config.TimeoutSeconds
		if timeout == 0 {
			timeout = 5
		}
		conn, err := net.DialTimeout("tcp", targetAddr, time.Duration(timeout)*time.Second)

		elapsed := time.Since(startTime)
		rttMs := math.Round(elapsed.Seconds()*1000*rttPrecisionMultiplier) / rttPrecisionMultiplier

		if err != nil {
			// Connection failed
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Connection succeeded, close immediately
		conn.Close()
		receivedPackets++

		samples = append(samples, SamplePoint{
			RTTMs:     rttMs,
			Timestamp: time.Now().Format(time.RFC3339),
			Success:   true,
		})
	}

	// Calculate core metrics
	metrics := collector.CalculateFromSamples(samples, sentPackets, receivedPackets)

	success := receivedPackets > 0
	errorMessage := ""
	if !success && len(errors) > 0 {
		errorMessage = fmt.Sprintf("probing failed: %d errors", len(errors))
	}

	return models.NewTCPProbeResultWithMetrics(
		success,
		metrics.RTTMs,
		metrics.RTTMedianMs,
		metrics.JitterMs,
		metrics.RTTVarianceMs,
		metrics.PacketLossRate,
		metrics.SampleCount,
		errorMessage,
	), nil
}

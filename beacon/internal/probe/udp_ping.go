package probe

import (
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"beacon/internal/models"
)

// UDPProbeConfig represents UDP probe configuration
type UDPProbeConfig struct {
	Type           string `yaml:"type" validate:"required,eq=udp_ping"`
	Target         string `yaml:"target" validate:"required,ip|hostname"`
	Port           int    `yaml:"port" validate:"required,min=1,max=65535"`
	TimeoutSeconds int    `yaml:"timeout_seconds" validate:"required,min=1,max=30"`
	Interval       int    `yaml:"interval" validate:"required,min=60,max=300"`
	Count          int    `yaml:"count" validate:"required,min=1,max=100"`
}

// Validate validates the UDP probe configuration
func (c *UDPProbeConfig) Validate() error {
	if c.Type != "udp_ping" {
		return fmt.Errorf("invalid probe type: %s (must be 'udp_ping')", c.Type)
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

	if c.TimeoutSeconds < 1 || c.TimeoutSeconds > 30 {
		return fmt.Errorf("invalid timeout %d, must be between 1 and 30 seconds", c.TimeoutSeconds)
	}

	if c.Interval < 60 || c.Interval > 300 {
		return fmt.Errorf("invalid interval %d, must be between 60 and 300 seconds", c.Interval)
	}

	// Validate that interval is sufficient for timeout and count
	minInterval := c.TimeoutSeconds * c.Count
	if c.Interval < minInterval {
		return fmt.Errorf("interval %d seconds insufficient for %d probes with timeout %d seconds (recommended: interval >= timeout * count = %d seconds)",
			c.Interval, c.Count, c.TimeoutSeconds, minInterval)
	}

	if c.Count < 1 || c.Count > 100 {
		return fmt.Errorf("invalid count %d, must be between 1 and 100", c.Count)
	}

	return nil
}

// UDPPinger represents a UDP ping probe engine
type UDPPinger struct {
	config UDPProbeConfig
}

// NewUDPPinger creates a new UDP pinger with the given configuration
func NewUDPPinger(config UDPProbeConfig) *UDPPinger {
	return &UDPPinger{
		config: config,
	}
}

// Execute performs a single UDP probe
func (p *UDPPinger) Execute() (*models.UDPProbeResult, error) {
	// Validate configuration before executing
	if err := p.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	startTime := time.Now()

	// Build target address with IPv6 support
	targetAddr := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))

	// Create UDP connection with timeout
	conn, err := net.DialTimeout("udp", targetAddr, time.Duration(p.config.TimeoutSeconds)*time.Second)

	if err != nil {
		// Connection failed
		return models.NewUDPProbeResult(false, 100.0, 0, 1, 0, err.Error()), nil
	}
	defer conn.Close()

	// Set write deadline
	writeDeadline := time.Now().Add(time.Duration(p.config.TimeoutSeconds) * time.Second)
	err = conn.SetWriteDeadline(writeDeadline)
	if err != nil {
		return models.NewUDPProbeResult(false, 100.0, 0, 1, 0, fmt.Sprintf("set write deadline failed: %v", err)), nil
	}

	// Send test packet
	testPayload := []byte("PING")
	_, err = conn.Write(testPayload)
	if err != nil {
		return models.NewUDPProbeResult(false, 100.0, 0, 1, 0, fmt.Sprintf("send failed: %v", err)), nil
	}

	// Set read deadline
	readDeadline := time.Now().Add(time.Duration(p.config.TimeoutSeconds) * time.Second)
	err = conn.SetReadDeadline(readDeadline)
	if err != nil {
		return models.NewUDPProbeResult(false, 100.0, 0, 1, 0, fmt.Sprintf("set read deadline failed: %v", err)), nil
	}

	// Wait for response
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)

	totalElapsed := time.Since(startTime)

	if err != nil {
		// Timeout or read failure - treat as packet loss
		return models.NewUDPProbeResult(false, 100.0, 0, 1, 0, fmt.Sprintf("no response: %v", err)), nil
	}

	// Success - received response
	rttMs := math.Round(totalElapsed.Seconds()*1000*rttPrecisionMultiplier) / rttPrecisionMultiplier

	return models.NewUDPProbeResult(true, 0.0, rttMs, 1, 1, ""), nil
}

// ExecuteBatch performs multiple UDP probes and calculates core metrics
func (p *UDPPinger) ExecuteBatch(count int) (*models.UDPProbeResult, error) {
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

		startTime := time.Now()

		// Build target address with IPv6 support
		targetAddr := net.JoinHostPort(p.config.Target, fmt.Sprintf("%d", p.config.Port))

		// Create UDP connection with timeout
		conn, err := net.DialTimeout("udp", targetAddr, time.Duration(p.config.TimeoutSeconds)*time.Second)

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

		// Set write deadline
		writeDeadline := time.Now().Add(time.Duration(p.config.TimeoutSeconds) * time.Second)
		err = conn.SetWriteDeadline(writeDeadline)
		if err != nil {
			conn.Close()
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Send test packet
		testPayload := []byte("PING")
		_, err = conn.Write(testPayload)
		if err != nil {
			conn.Close()
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Set read deadline
		readDeadline := time.Now().Add(time.Duration(p.config.TimeoutSeconds) * time.Second)
		err = conn.SetReadDeadline(readDeadline)
		if err != nil {
			conn.Close()
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Wait for response
		buffer := make([]byte, 1024)
		_, err = conn.Read(buffer)
		conn.Close()

		totalElapsed := time.Since(startTime)

		if err != nil {
			// Timeout or read failure - treat as packet loss
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Success - received response
		receivedPackets++
		rttMs := math.Round(totalElapsed.Seconds()*1000*rttPrecisionMultiplier) / rttPrecisionMultiplier

		samples = append(samples, SamplePoint{
			RTTMs:     rttMs,
			Timestamp: time.Now().Format(time.RFC3339),
			Success:   true,
		})
	}

	// Calculate core metrics using CoreMetricsCollector
	metrics := collector.CalculateFromSamples(samples, sentPackets, receivedPackets)

	// Determine success (at least one response received)
	success := receivedPackets > 0

	errorMessage := ""
	if !success && len(errors) > 0 {
		// Provide summary statistics instead of truncating errors
		errorSummary := map[string]int{}
		for _, err := range errors {
			// Group errors by type (timeout, connection refused, etc.)
			if strings.Contains(err, "timeout") || strings.Contains(err, "no response") {
				errorSummary["timeout"]++
			} else if strings.Contains(err, "connection refused") || strings.Contains(err, "connect") {
				errorSummary["connection_refused"]++
			} else if strings.Contains(err, "send failed") {
				errorSummary["send_failed"]++
			} else {
				errorSummary["other"]++
			}
		}

		// Build error summary message
		var summaryParts []string
		for errorType, count := range errorSummary {
			summaryParts = append(summaryParts, fmt.Sprintf("%s: %d", errorType, count))
		}

		errorMessage = fmt.Sprintf("丢包原因统计: %s", strings.Join(summaryParts, ", "))
	}

	return models.NewUDPProbeResultWithMetrics(
		success,
		metrics.PacketLossRate,
		metrics.RTTMs,
		metrics.RTTMedianMs,
		metrics.JitterMs,
		metrics.RTTVarianceMs,
		sentPackets,
		receivedPackets,
		metrics.SampleCount,
		errorMessage,
	), nil
}

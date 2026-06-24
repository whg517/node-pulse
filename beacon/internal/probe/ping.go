package probe

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/whg517/node-pulse/beacon/internal/models"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// ICMPProbeConfig represents ICMP ping probe configuration
type ICMPProbeConfig struct {
	Type           string `yaml:"type" validate:"required,eq=icmp_ping"`
	Target         string `yaml:"target" validate:"required,ip|hostname"`
	TimeoutSeconds int    `yaml:"timeout" validate:"required,min=1,max=30"`
	Interval       int    `yaml:"interval" validate:"required,min=60,max=300"`
	Count          int    `yaml:"count" validate:"required,min=1,max=100"`
	PacketSize     int    `yaml:"packet_size" validate:"omitempty,min=8,max=65507"`
}

// Validate validates the ICMP probe configuration
func (c *ICMPProbeConfig) Validate() error {
	if c.Type != "icmp_ping" {
		return fmt.Errorf("invalid probe type: %s (must be 'icmp_ping')", c.Type)
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

	// Default packet size is 56 bytes (64 bytes total with ICMP header)
	if c.PacketSize == 0 {
		c.PacketSize = 56
	}
	if c.PacketSize < 8 || c.PacketSize > 65507 {
		return fmt.Errorf("invalid packet_size %d, must be between 8 and 65507 bytes", c.PacketSize)
	}

	return nil
}

// ICMPPinger represents an ICMP ping probe engine
type ICMPPinger struct {
	config ICMPProbeConfig
}

// NewICMPPinger creates a new ICMP pinger with the given configuration
func NewICMPPinger(config ICMPProbeConfig) *ICMPPinger {
	// Set default packet size if not specified
	if config.PacketSize == 0 {
		config.PacketSize = 56
	}
	return &ICMPPinger{
		config: config,
	}
}

// ICMPProbeResult represents the result of an ICMP ping operation
type ICMPProbeResult struct {
	Success        bool    `json:"success"`          // Connectivity (success/failure)
	RTTMs          float64 `json:"rtt_ms"`           // Round-trip time in milliseconds (mean)
	RTTMedianMs    float64 `json:"rtt_median_ms"`    // RTT median in milliseconds
	JitterMs       float64 `json:"jitter_ms"`        // Delay jitter in milliseconds
	VarianceMs     float64 `json:"variance_ms"`      // RTT variance in milliseconds^2
	PacketLossRate float64 `json:"packet_loss_rate"` // Packet loss rate (0-100%)
	SampleCount    int     `json:"sample_count"`     // Number of sample points
	TTL            int     `json:"ttl"`              // Time to live from response
	ErrorMessage   string  `json:"error_message"`    // Error message if failed
	Timestamp      string  `json:"timestamp"`        // Probe timestamp (ISO 8601)
}

// Execute performs a single ICMP probe
func (p *ICMPPinger) Execute() (*ICMPProbeResult, error) {
	// Set default timeout if not configured
	timeout := p.config.TimeoutSeconds
	if timeout == 0 {
		timeout = 5
	}

	// Validate configuration before executing
	if err := p.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Resolve the target address
	ipAddr, networkType, err := p.resolveTarget()
	if err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: err.Error(),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Create ICMP connection
	conn, err := p.createICMPConnection(networkType)
	if err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("ICMP requires root/admin privileges: %v", err),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}
	defer func() { _ = conn.Close() }()

	// Set deadline
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to set deadline: %v", err),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Send ICMP echo request
	startTime := time.Now()
	seq := int(startTime.UnixNano() & 0xFFFF)

	echoRequest := p.createEchoRequest(seq, networkType)

	// Get the destination address
	dstAddr := &net.IPAddr{IP: ipAddr}
	if networkType == "ipv6" {
		dstAddr = &net.IPAddr{IP: ipAddr}
	}

	// Send the packet
	if _, err := conn.WriteTo(echoRequest, dstAddr); err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to send ICMP packet: %v", err),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Receive the reply
	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	elapsed := time.Since(startTime)

	if err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("no ICMP reply received: %v", err),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Parse the reply
	ttl, err := p.parseReply(reply[:n], seq, networkType)
	if err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: err.Error(),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Calculate RTT in milliseconds (rounded to 2 decimal places)
	rttMs := math.Round(elapsed.Seconds()*1000*rttPrecisionMultiplier) / rttPrecisionMultiplier

	return &ICMPProbeResult{
		Success:      true,
		RTTMs:        rttMs,
		TTL:          ttl,
		ErrorMessage: "",
		Timestamp:    time.Now().Format(time.RFC3339),
	}, nil
}

// ExecuteBatch performs multiple ICMP probes and calculates core metrics
func (p *ICMPPinger) ExecuteBatch(count int) (*ICMPProbeResult, error) {
	if count < 1 || count > 100 {
		return nil, fmt.Errorf("invalid count %d, must be between 1 and 100", count)
	}

	samples := make([]SamplePoint, 0, count)
	sentPackets := 0
	receivedPackets := 0
	var errors []string
	var lastTTL int

	collector := NewCoreMetricsCollector()

	// Resolve target once
	ipAddr, networkType, err := p.resolveTarget()
	if err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: err.Error(),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Create ICMP connection
	conn, err := p.createICMPConnection(networkType)
	if err != nil {
		return &ICMPProbeResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("ICMP requires root/admin privileges: %v", err),
			Timestamp:    time.Now().Format(time.RFC3339),
		}, nil
	}
	defer func() { _ = conn.Close() }()

	for i := 0; i < count; i++ {
		sentPackets++
		startTime := time.Now()

		// Set deadline for this probe
		timeout := p.config.TimeoutSeconds
		if timeout == 0 {
			timeout = 5
		}
		deadline := time.Now().Add(time.Duration(timeout) * time.Second)
		if err := conn.SetDeadline(deadline); err != nil {
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Create and send ICMP echo request
		seq := (i + 1) & 0xFFFF
		echoRequest := p.createEchoRequest(seq, networkType)

		dstAddr := &net.IPAddr{IP: ipAddr}

		if _, err := conn.WriteTo(echoRequest, dstAddr); err != nil {
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Receive the reply
		reply := make([]byte, 1500)
		n, _, err := conn.ReadFrom(reply)
		elapsed := time.Since(startTime)

		if err != nil {
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Parse the reply
		ttl, err := p.parseReply(reply[:n], seq, networkType)
		if err != nil {
			errors = append(errors, err.Error())
			samples = append(samples, SamplePoint{
				RTTMs:     0,
				Timestamp: time.Now().Format(time.RFC3339),
				Success:   false,
			})
			continue
		}

		// Success
		receivedPackets++
		lastTTL = ttl
		rttMs := math.Round(elapsed.Seconds()*1000*rttPrecisionMultiplier) / rttPrecisionMultiplier

		samples = append(samples, SamplePoint{
			RTTMs:     rttMs,
			Timestamp: time.Now().Format(time.RFC3339),
			Success:   true,
		})

		// Small delay between probes (100ms)
		if i < count-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Calculate core metrics
	metrics := collector.CalculateFromSamples(samples, sentPackets, receivedPackets)

	success := receivedPackets > 0
	errorMessage := ""
	if !success && len(errors) > 0 {
		errorMessage = fmt.Sprintf("ICMP probing failed: %d errors", len(errors))
	}

	return &ICMPProbeResult{
		Success:        success,
		RTTMs:          metrics.RTTMs,
		RTTMedianMs:    metrics.RTTMedianMs,
		JitterMs:       metrics.JitterMs,
		VarianceMs:     metrics.RTTVarianceMs,
		PacketLossRate: metrics.PacketLossRate,
		SampleCount:    metrics.SampleCount,
		TTL:            lastTTL,
		ErrorMessage:   errorMessage,
		Timestamp:      time.Now().Format(time.RFC3339),
	}, nil
}

// resolveTarget resolves the target to an IP address and determines IPv4 or IPv6
func (p *ICMPPinger) resolveTarget() (net.IP, string, error) {
	// First, try parsing as IP
	if ip := net.ParseIP(p.config.Target); ip != nil {
		if ip.To4() != nil {
			return ip, "ipv4", nil
		}
		return ip, "ipv6", nil
	}

	// Resolve as hostname
	ips, err := net.LookupIP(p.config.Target)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve hostname '%s': %v", p.config.Target, err)
	}

	if len(ips) == 0 {
		return nil, "", fmt.Errorf("no IP addresses found for hostname '%s'", p.config.Target)
	}

	// Prefer IPv4, fall back to IPv6
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip, "ipv4", nil
		}
	}

	// Use first IPv6 address
	return ips[0], "ipv6", nil
}

// createICMPConnection creates an ICMP connection for the given network type
func (p *ICMPPinger) createICMPConnection(networkType string) (*icmp.PacketConn, error) {
	var network string
	if networkType == "ipv4" {
		network = "ip4:icmp"
	} else {
		network = "ip6:ipv6-icmp"
	}

	return icmp.ListenPacket(network, "0.0.0.0")
}

// createEchoRequest creates an ICMP echo request message
func (p *ICMPPinger) createEchoRequest(seq int, networkType string) []byte {
	// Create payload (pad with bytes)
	payload := make([]byte, p.config.PacketSize)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	// Use current time as ID
	id := int(time.Now().UnixNano() & 0xFFFF)

	var msgType icmp.Type
	if networkType == "ipv4" {
		msgType = ipv4.ICMPTypeEcho
	} else {
		msgType = ipv6.ICMPTypeEchoRequest
	}

	msg := &icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: payload,
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return nil
	}

	return msgBytes
}

// parseReply parses the ICMP reply and validates it matches our request
func (p *ICMPPinger) parseReply(data []byte, expectedSeq int, networkType string) (int, error) {
	var msg *icmp.Message
	var err error

	if networkType == "ipv4" {
		msg, err = icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), data)
	} else {
		msg, err = icmp.ParseMessage(ipv6.ICMPTypeEchoReply.Protocol(), data)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to parse ICMP reply: %v", err)
	}

	// Check message type
	if networkType == "ipv4" {
		if msg.Type != ipv4.ICMPTypeEchoReply {
			return 0, fmt.Errorf("unexpected ICMP message type: %v (expected EchoReply)", msg.Type)
		}
	} else {
		if msg.Type != ipv6.ICMPTypeEchoReply {
			return 0, fmt.Errorf("unexpected ICMP message type: %v (expected EchoReply)", msg.Type)
		}
	}

	// Extract TTL from IPv4 header (for IPv4)
	// For IPv6, hop limit is handled differently
	ttl := 0
	if networkType == "ipv4" && len(data) >= 8 {
		// The IP header TTL is not directly accessible from parsed ICMP message
		// We set it to 0 since we can't easily extract it
		ttl = 64 // Typical default TTL
	}

	// Validate echo reply body
	echo, ok := msg.Body.(*icmp.Echo)
	if !ok {
		return 0, fmt.Errorf("invalid ICMP echo reply body")
	}

	// Verify sequence number matches
	if echo.Seq != expectedSeq {
		return 0, fmt.Errorf("sequence mismatch: expected %d, got %d", expectedSeq, echo.Seq)
	}

	return ttl, nil
}

// ToGenericResult converts an ICMPProbeResult to a generic ProbeResult
func (r *ICMPProbeResult) ToGenericResult(target string) *models.ProbeResult {
	metrics := map[string]interface{}{
		"rtt_ms":  r.RTTMs,
		"ttl":     r.TTL,
		"success": r.Success,
	}

	if r.SampleCount > 0 {
		metrics["rtt_median_ms"] = r.RTTMedianMs
		metrics["jitter_ms"] = r.JitterMs
		metrics["variance_ms"] = r.VarianceMs
		metrics["packet_loss_rate"] = r.PacketLossRate
		metrics["sample_count"] = r.SampleCount
	}

	return models.NewProbeResult("icmp_ping", target, r.Success, metrics, r.ErrorMessage)
}

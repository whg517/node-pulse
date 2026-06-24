package probe

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/whg517/node-pulse/beacon/internal/models"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// MTRProbeConfig represents MTR probe configuration
type MTRProbeConfig struct {
	Type           string `yaml:"type" validate:"required,eq=mtr"`
	Target         string `yaml:"target" validate:"required,ip|hostname"`
	MaxHops        int    `yaml:"max_hops" validate:"required,min=1,max=64"`
	TimeoutSeconds int    `yaml:"timeout" validate:"required,min=1,max=30"`
	Interval       int    `yaml:"interval" validate:"required,min=60,max=300"`
	Count          int    `yaml:"count" validate:"required,min=1,max=30"`
	PacketSize     int    `yaml:"packet_size" validate:"omitempty,min=64,max=1500"`
}

// Validate validates the MTR probe configuration
func (c *MTRProbeConfig) Validate() error {
	if c.Type != "mtr" {
		return fmt.Errorf("invalid probe type: %s (must be 'mtr')", c.Type)
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

	if c.MaxHops < 1 || c.MaxHops > 64 {
		return fmt.Errorf("invalid max_hops %d, must be between 1 and 64", c.MaxHops)
	}

	if c.TimeoutSeconds < 1 || c.TimeoutSeconds > 30 {
		return fmt.Errorf("invalid timeout %d, must be between 1 and 30 seconds", c.TimeoutSeconds)
	}

	if c.Interval < 60 || c.Interval > 300 {
		return fmt.Errorf("invalid interval %d, must be between 60 and 300 seconds", c.Interval)
	}

	if c.Count < 1 || c.Count > 30 {
		return fmt.Errorf("invalid count %d, must be between 1 and 30", c.Count)
	}

	// Validate packet size if specified
	if c.PacketSize != 0 && (c.PacketSize < 64 || c.PacketSize > 1500) {
		return fmt.Errorf("invalid packet_size %d, must be between 64 and 1500", c.PacketSize)
	}

	return nil
}

// MTRProbe represents an MTR probe engine
type MTRProbe struct {
	config MTRProbeConfig
	destIP net.IP
}

// NewMTRProbe creates a new MTR probe with the given configuration
func NewMTRProbe(config MTRProbeConfig) *MTRProbe {
	return &MTRProbe{
		config: config,
	}
}

// Execute performs the MTR probe and returns the result
func (p *MTRProbe) Execute() (*models.MTRResult, error) {
	// Validate configuration before executing
	if err := p.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Resolve target to IP address
	var err error
	p.destIP, err = p.resolveTarget()
	if err != nil {
		return models.NewMTRResult(p.config.Target, nil, false, fmt.Sprintf("failed to resolve target: %v", err)), nil
	}

	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return models.NewMTRResult(p.config.Target, nil, false, fmt.Sprintf("failed to create ICMP socket: %v", err)), nil
	}
	defer func() { _ = conn.Close() }()

	// Set timeout
	deadline := time.Now().Add(time.Duration(p.config.TimeoutSeconds*len(p.config.Target)+30) * time.Second)
	_ = conn.SetDeadline(deadline)

	// Discover hops using traceroute
	hops := p.discoverHops(conn)

	// Probe each hop multiple times for statistics
	hopsWithData := p.probeHops(conn, hops)

	return models.NewMTRResult(p.config.Target, hopsWithData, true, ""), nil
}

// resolveTarget resolves the target hostname to an IP address
func (p *MTRProbe) resolveTarget() (net.IP, error) {
	// If already an IP, return it
	if ip := net.ParseIP(p.config.Target); ip != nil {
		return ip, nil
	}

	// Resolve hostname
	addrs, err := net.LookupIP(p.config.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname '%s': %w", p.config.Target, err)
	}

	// Find IPv4 address
	for _, addr := range addrs {
		if addr.To4() != nil {
			return addr, nil
		}
	}

	return nil, fmt.Errorf("no IPv4 address found for '%s'", p.config.Target)
}

// discoverHops discovers all hops along the path using traceroute
func (p *MTRProbe) discoverHops(conn *icmp.PacketConn) map[int]string {
	hops := make(map[int]string)
	var mu sync.Mutex

	// Use a map to track which hops we've found
	for ttl := 1; ttl <= p.config.MaxHops; ttl++ {
		// Send single probe to discover hop
		hopIP := p.sendTTLProbe(conn, ttl)
		if hopIP != "" {
			mu.Lock()
			hops[ttl] = hopIP
			mu.Unlock()
		}

		// Check if we've reached the destination
		if hopIP == p.destIP.String() {
			break
		}
	}

	return hops
}

// probeHops probes each discovered hop multiple times to gather statistics
func (p *MTRProbe) probeHops(conn *icmp.PacketConn, hopMap map[int]string) []models.MTRHop {
	var hops []models.MTRHop

	for ttl := 1; ttl <= p.config.MaxHops; ttl++ {
		hopIP, found := hopMap[ttl]
		if !found || hopIP == "" {
			// Hop didn't respond, create entry with no data
			hops = append(hops, models.MTRHop{
				HopNumber: ttl,
				IP:        "*",
				Sent:      p.config.Count,
				Received:  0,
				LossRate:  100.0,
			})
			continue
		}

		// Probe this hop multiple times
		rtts := p.probeHopMultiple(conn, hopIP, ttl)

		hop := models.NewMTRHop(ttl, hopIP, rtts, p.config.Count)

		// Try to resolve hostname (best effort)
		names, _ := net.LookupAddr(hopIP)
		if len(names) > 0 {
			hop.Hostname = names[0]
		}

		hops = append(hops, hop)

		// Stop if we've reached the destination
		if hopIP == p.destIP.String() {
			break
		}
	}

	return hops
}

// probeHopMultiple sends multiple probes to a hop and collects RTTs
func (p *MTRProbe) probeHopMultiple(conn *icmp.PacketConn, hopIP string, ttl int) []float64 {
	var rtts []float64

	for i := 0; i < p.config.Count; i++ {
		rtt := p.probeHop(conn, hopIP, ttl)
		if rtt > 0 {
			rtts = append(rtts, rtt)
		}
		// Small delay between probes
		time.Sleep(50 * time.Millisecond)
	}

	return rtts
}

// probeHop sends a single probe to a hop and returns the RTT
func (p *MTRProbe) probeHop(conn *icmp.PacketConn, _ string, ttl int) float64 {
	// Create ICMP echo request
	echo := icmp.Echo{
		ID:   int(binary.BigEndian.Uint16([]byte{byte(ttl), byte(time.Now().Unix() % 256)})),
		Seq:  1,
		Data: p.createProbeData(),
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &echo,
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return 0
	}

	// Set TTL
	_ = conn.IPv4PacketConn().SetTTL(ttl)

	// Send packet
	destAddr := &net.IPAddr{IP: p.destIP}
	startTime := time.Now()
	_, err = conn.WriteTo(msgBytes, destAddr)
	if err != nil {
		return 0
	}

	// Read response with timeout
	response := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(time.Duration(p.config.TimeoutSeconds) * time.Second))
	if err != nil {
		return 0
	}

	for {
		n, _, err := conn.ReadFrom(response)
		if err != nil {
			return 0
		}

		rtt := float64(time.Since(startTime).Milliseconds())

		// Parse ICMP response
		replyMsg, err := icmp.ParseMessage(1, response[:n])
		if err != nil {
			continue
		}

		switch replyMsg.Type {
		case ipv4.ICMPTypeEchoReply:
			// Reached destination
			return roundRTT(rtt)
		case ipv4.ICMPTypeTimeExceeded:
			// TTL expired at intermediate hop
			return roundRTT(rtt)
		case ipv4.ICMPTypeDestinationUnreachable:
			// Destination unreachable
			return roundRTT(rtt)
		}
	}
}

// sendTTLProbe sends a single probe with given TTL and returns the responding hop IP
func (p *MTRProbe) sendTTLProbe(conn *icmp.PacketConn, ttl int) string {
	// Create ICMP echo request
	echo := icmp.Echo{
		ID:   int(binary.BigEndian.Uint16([]byte{byte(ttl), byte(time.Now().Unix() % 256)})),
		Seq:  1,
		Data: p.createProbeData(),
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &echo,
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return ""
	}

	// Set TTL
	_ = conn.IPv4PacketConn().SetTTL(ttl)

	// Send packet
	destAddr := &net.IPAddr{IP: p.destIP}
	_, err = conn.WriteTo(msgBytes, destAddr)
	if err != nil {
		return ""
	}

	// Read response with timeout
	response := make([]byte, 1500)
	_ = conn.SetReadDeadline(time.Now().Add(time.Duration(p.config.TimeoutSeconds) * time.Second))

	for {
		n, addr, err := conn.ReadFrom(response)
		if err != nil {
			return ""
		}

		// Parse ICMP response
		replyMsg, err := icmp.ParseMessage(1, response[:n])
		if err != nil {
			continue
		}

		switch replyMsg.Type {
		case ipv4.ICMPTypeEchoReply:
			// Reached destination
			return addr.String()
		case ipv4.ICMPTypeTimeExceeded:
			// TTL expired at intermediate hop - return the hop's IP
			ipAddr := addr.String()
			// Strip port if present
			host, _, err := net.SplitHostPort(ipAddr)
			if err != nil {
				return ipAddr
			}
			return host
		case ipv4.ICMPTypeDestinationUnreachable:
			// Destination unreachable
			return addr.String()
		}
	}
}

// createProbeData creates the probe data payload
func (p *MTRProbe) createProbeData() []byte {
	packetSize := p.config.PacketSize
	if packetSize == 0 {
		packetSize = 64
	}

	// ICMP header is 8 bytes, so data is packetSize - 8
	dataSize := packetSize - 8
	if dataSize < 1 {
		dataSize = 56 // Default ping data size
	}

	data := bytes.Repeat([]byte("MTR"), dataSize/3+1)
	return data[:dataSize]
}

// roundRTT rounds RTT to 2 decimal places
func roundRTT(ms float64) float64 {
	return math.Round(ms*100) / 100
}

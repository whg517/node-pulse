package diagnostics

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

// NetworkStatus contains network diagnostic information
type NetworkStatus struct {
	PulseServerReachable bool              `json:"pulse_server_reachable"`
	PulseServerAddress   string            `json:"pulse_server_address"`
	RTTMs                RTTStatistics     `json:"rtt_ms,omitempty"`
	PacketLossRate       float64           `json:"packet_loss_rate"`
	NetworkInterface     InterfaceInfo     `json:"network_interface,omitempty"`
	DNSResolution        DNSInfo           `json:"dns_resolution,omitempty"`
	RecentFailures       []ConnectionFailure `json:"recent_failures,omitempty"` // Recent connection failures
}

// ConnectionFailure represents a single connection failure event
type ConnectionFailure struct {
	Timestamp   time.Time `json:"timestamp"`
	ErrorReason string    `json:"error_reason"`
}

// RTTStatistics contains Round Trip Time statistics
type RTTStatistics struct {
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Samples int     `json:"samples"`
}

// InterfaceInfo contains network interface information
type InterfaceInfo struct {
	IPAddress string `json:"ip_address,omitempty"`
	MACAddress string `json:"mac_address,omitempty"`
}

// DNSInfo contains DNS resolution information
type DNSInfo struct {
	Status     string `json:"status,omitempty"`
	ResolvedIP string `json:"resolved_ip,omitempty"`
}

// collectNetworkStatus collects network status information
func (c *collector) collectNetworkStatus() (*NetworkStatus, error) {
	status := &NetworkStatus{
		PulseServerAddress: c.cfg.PulseServer,
		RTTMs: RTTStatistics{
			Avg:     0,
			Min:     0,
			Max:     0,
			Samples: 0,
		},
		PacketLossRate: 0,
		RecentFailures: []ConnectionFailure{}, // Initialize empty failure list
	}

	// Perform multiple connectivity checks for statistics
	const numPings = 5
	successfulPings := 0
	rttValues := make([]float64, 0, numPings)

	for i := 0; i < numPings; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", extractHost(c.cfg.PulseServer), 3*time.Second)
		if err != nil {
			// Record this failure
			status.RecentFailures = append(status.RecentFailures, ConnectionFailure{
				Timestamp:   time.Now(),
				ErrorReason: err.Error(),
			})
		} else {
			successfulPings++
			rtt := float64(time.Since(start).Microseconds()) / 1000.0 // Convert to ms
			rttValues = append(rttValues, rtt)
			conn.Close()
		}

		// Small delay between pings to avoid overwhelming the server
		if i < numPings-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Calculate statistics based on successful pings
	status.PulseServerReachable = successfulPings > 0
	status.PacketLossRate = float64(numPings-successfulPings) / float64(numPings)

	if successfulPings > 0 {
		// Calculate RTT statistics
		min := rttValues[0]
		max := rttValues[0]
		sum := 0.0
		for _, rtt := range rttValues {
			if rtt < min {
				min = rtt
			}
			if rtt > max {
				max = rtt
			}
			sum += rtt
		}
		status.RTTMs = RTTStatistics{
			Avg:     sum / float64(successfulPings),
			Min:     min,
			Max:     max,
			Samples: successfulPings,
		}
	}

	// Get network interface information
	if iface, err := c.getNetworkInterface(); err == nil {
		status.NetworkInterface = *iface
	}

	// DNS resolution
	if dns, err := c.resolveDNS(c.cfg.PulseServer); err == nil {
		status.DNSResolution = *dns
	}

	return status, nil
}

// getNetworkInterface gets the primary network interface information
func (c *collector) getNetworkInterface() (*InterfaceInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Find first non-loopback interface with an IP address
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			if ip.To4() != nil {
				return &InterfaceInfo{
					IPAddress:  ip.String(),
					MACAddress: iface.HardwareAddr.String(),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid network interface found")
}

// resolveDNS resolves the hostname from the Pulse server URL
func (c *collector) resolveDNS(pulseServer string) (*DNSInfo, error) {
	u, err := url.Parse(pulseServer)
	if err != nil {
		return nil, err
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("no hostname in URL")
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return &DNSInfo{
			Status: "failed",
		}, nil
	}

	// Get first IPv4 address
	for _, ip := range ips {
		if ip.To4() != nil {
			return &DNSInfo{
				Status:     "success",
				ResolvedIP: ip.String(),
			}, nil
		}
	}

	return &DNSInfo{
		Status: "no IPv4 address",
	}, nil
}

// extractHost extracts the host from a URL, adding default port if needed
func extractHost(pulseServer string) string {
	u, err := url.Parse(pulseServer)
	if err != nil {
		return pulseServer
	}

	host := u.Host
	if host == "" {
		return pulseServer
	}

	// If no port, add default based on scheme
	if u.Scheme == "https" && !containsColon(host) {
		return host + ":443"
	} else if u.Scheme == "http" && !containsColon(host) {
		return host + ":80"
	}

	return host
}

// containsColon checks if a string contains a colon (port separator)
func containsColon(s string) bool {
	for _, c := range s {
		if c == ':' {
			return true
		}
	}
	return false
}

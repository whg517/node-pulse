package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// BlockedIPNets contains IP ranges that are blocked for SSRF protection
var BlockedIPNets = []*net.IPNet{
	mustParseCIDR("10.0.0.0/8"),       // RFC 1918 Private
	mustParseCIDR("172.16.0.0/12"),     // RFC 1918 Private
	mustParseCIDR("192.168.0.0/16"),    // RFC 1918 Private
	mustParseCIDR("127.0.0.0/8"),       // Loopback
	mustParseCIDR("169.254.169.254/32"), // Cloud metadata (AWS/GCP/Azure)
	mustParseCIDR("::1/128"),           // IPv6 loopback
	mustParseCIDR("fc00::/7"),          // IPv6 ULA
	mustParseCIDR("fe80::/10"),         // IPv6 link-local
	mustParseCIDR("0.0.0.0/8"),         // Invalid addresses
	mustParseCIDR("224.0.0.0/4"),       // IP multicast
	mustParseCIDR("255.255.255.255/32"), // Broadcast
}

// AllowedDomains is an optional allowlist for webhook URLs (empty = allow all HTTPS)
// This should be configured based on environment requirements
var AllowedDomains []string

// ValidateWebhookURL validates a webhook URL for SSRF protection
// Returns an error if the URL is invalid or potentially malicious
func ValidateWebhookURL(rawURL string) error {
	// 1. Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// 2. Scheme must be HTTPS only (no HTTP, file://, etc.)
	if u.Scheme != "https" {
		return fmt.Errorf("only HTTPS webhooks allowed, got scheme: %s", u.Scheme)
	}

	// 3. Check domain allowlist (if configured)
	if len(AllowedDomains) > 0 {
		allowed := false
		host := strings.ToLower(u.Hostname())
		for _, allowedDomain := range AllowedDomains {
			if strings.HasSuffix(host, strings.ToLower(allowedDomain)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("domain not in allowlist: %s", u.Hostname())
		}
	}

	// 4. Resolve DNS and check ALL resolved IPs against blocklist
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return fmt.Errorf("DNS resolution failed for %s: %w", u.Hostname(), err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses resolved for %s", u.Hostname())
	}

	for _, ip := range ips {
		if IsBlockedIP(ip) {
			return fmt.Errorf("SSRF attempt blocked: private/reserved IP detected (%s) for %s", ip, u.Hostname())
		}
	}

	return nil
}

// IsBlockedIP checks if an IP address is in the blocked range
func IsBlockedIP(ip net.IP) bool {
	for _, blocked := range BlockedIPNets {
		if blocked.Contains(ip) {
			return true
		}
	}
	return false
}

// SetAllowedDomains sets the domain allowlist for webhook URLs
// Pass an empty slice to disable allowlist checking
func SetAllowedDomains(domains []string) {
	AllowedDomains = domains
}

// mustParseCIDR parses a CIDR string and panics on error
// Used for initializing blocked IP nets at package load time
func mustParseCIDR(cidr string) *net.IPNet {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse CIDR %s: %v", cidr, err))
	}
	return ipnet
}

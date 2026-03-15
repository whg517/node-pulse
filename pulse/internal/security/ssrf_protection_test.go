package security

import (
	"net"
	"testing"
)

// canResolveExternalDNS checks whether external DNS resolution is available in the current environment.
func canResolveExternalDNS() bool {
	_, err := net.LookupIP("hooks.slack.com")
	return err == nil
}

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantErr     bool
		errMsg      string
		requiresDNS bool
	}{
		{
			name:        "Valid HTTPS URL",
			url:         "https://hooks.slack.com/services/xxx",
			wantErr:     false,
			requiresDNS: true,
		},
		{
			name:    "HTTP URL blocked",
			url:     "http://example.com/webhook",
			wantErr: true,
			errMsg:  "only HTTPS webhooks allowed",
		},
		{
			name:    "File URL blocked",
			url:     "file:///etc/passwd",
			wantErr: true,
			errMsg:  "only HTTPS webhooks allowed",
		},
		{
			name:    "Loopback IP blocked",
			url:     "https://127.0.0.1/webhook",
			wantErr: true,
			errMsg:  "SSRF attempt blocked",
		},
		{
			name:    "Localhost blocked",
			url:     "https://localhost/webhook",
			wantErr: true,
			errMsg:  "SSRF attempt blocked",
		},
		{
			name:    "RFC 1918 private IP blocked (10.0.0.0/8)",
			url:     "https://10.0.0.1/webhook",
			wantErr: true,
			errMsg:  "SSRF attempt blocked",
		},
		{
			name:    "RFC 1918 private IP blocked (172.16.0.0/12)",
			url:     "https://172.16.0.1/webhook",
			wantErr: true,
			errMsg:  "SSRF attempt blocked",
		},
		{
			name:    "RFC 1918 private IP blocked (192.168.0.0/16)",
			url:     "https://192.168.1.1/webhook",
			wantErr: true,
			errMsg:  "SSRF attempt blocked",
		},
		{
			name:    "Cloud metadata endpoint blocked (AWS)",
			url:     "https://169.254.169.254/latest/meta-data/",
			wantErr: true,
			errMsg:  "SSRF attempt blocked",
		},
		{
			name:    "Invalid URL format",
			url:     "://not-a-url",
			wantErr: true,
			errMsg:  "invalid URL format",
		},
		{
			name:    "IPv6 loopback blocked",
			url:     "https://[::1]/webhook",
			wantErr: true,
			errMsg:  "SSRF attempt blocked",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.requiresDNS && !canResolveExternalDNS() {
				t.Skip("Skipping test: external DNS resolution not available in this environment")
			}
			err := ValidateWebhookURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWebhookURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateWebhookURL() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		blocked  bool
	}{
		{
			name:    "Public IP allowed",
			ip:      "8.8.8.8",
			blocked: false,
		},
		{
			name:    "Loopback blocked",
			ip:      "127.0.0.1",
			blocked: true,
		},
		{
			name:    "RFC 1918 blocked",
			ip:      "192.168.1.1",
			blocked: true,
		},
		{
			name:    "Cloud metadata blocked",
			ip:      "169.254.169.254",
			blocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			blocked := IsBlockedIP(ip)
			if blocked != tt.blocked {
				t.Errorf("IsBlockedIP(%s) = %v, want %v", tt.ip, blocked, tt.blocked)
			}
		})
	}
}

func TestSetAllowedDomains(t *testing.T) {
	// Save original domains
	originalDomains := AllowedDomains
	defer func() {
		AllowedDomains = originalDomains
	}()

	dnsAvailable := canResolveExternalDNS()

	// Test allowlist functionality
	SetAllowedDomains([]string{"hooks.slack.com", "api.pagerduty.com"})

	// Allowed domain should pass (only verify when DNS is available)
	if dnsAvailable {
		err := ValidateWebhookURL("https://hooks.slack.com/services/xxx")
		if err != nil {
			t.Errorf("Allowed domain was blocked: %v", err)
		}
	}

	// Non-allowed domain should fail (this doesn't require DNS resolution since it's blocked by allowlist)
	err := ValidateWebhookURL("https://example.com/webhook")
	if err == nil {
		t.Error("Non-allowed domain was not blocked")
	}

	// Clear allowlist – valid public URL should pass when DNS is available
	SetAllowedDomains([]string{})
	if dnsAvailable {
		err = ValidateWebhookURL("https://example.com/webhook")
		if err != nil {
			t.Errorf("After clearing allowlist, valid URL should pass: %v", err)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

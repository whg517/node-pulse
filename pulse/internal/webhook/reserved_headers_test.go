package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsReservedWebhookHeader covers the reserved-header guard added for J9
// (custom headers). These headers are owned by the HTTP client / Pulse and
// must not be overridable via webhook custom_headers.
func TestIsReservedWebhookHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"Content-Type canonical", "Content-Type", true},
		{"Host", "Host", true},
		{"User-Agent", "User-Agent", true},
		{"Content-Length", "Content-Length", true},
		{"Transfer-Encoding", "Transfer-Encoding", true},
		{"Trailer", "Trailer", true},
		{"Connection", "Connection", true},
		{"custom Authorization", "Authorization", false},
		{"custom X-Tenant", "X-Tenant", false},
		{"custom X-Request-Id", "X-Request-Id", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isReservedWebhookHeader(tt.header))
		})
	}
}

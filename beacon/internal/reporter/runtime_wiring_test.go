package reporter

import (
	"testing"
	"time"
)

// TestComputeBackoff verifies the three reconnect backoff strategies (G19).
func TestComputeBackoff(t *testing.T) {
	r := &HeartbeatReporter{retryBase: 1 * time.Second}

	cases := []struct {
		name     string
		backoff  string
		attempt  int
		expected time.Duration
	}{
		{"exponential-0", "exponential", 0, 1 * time.Second},
		{"exponential-1", "exponential", 1, 2 * time.Second},
		{"exponential-2", "exponential", 2, 4 * time.Second},
		{"linear-0", "linear", 0, 1 * time.Second},
		{"linear-1", "linear", 1, 2 * time.Second},
		{"linear-2", "linear", 2, 3 * time.Second},
		{"constant-0", "constant", 0, 1 * time.Second},
		{"constant-2", "constant", 2, 1 * time.Second},
		{"unknown-falls-back-to-exponential", "", 2, 4 * time.Second},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r.backoff = c.backoff
			got := r.computeBackoff(c.attempt)
			if got != c.expected {
				t.Fatalf("computeBackoff(%q, %d) = %v, want %v", c.backoff, c.attempt, got, c.expected)
			}
		})
	}
}

// TestReporterDefaults verifies that a reporter with no options keeps the legacy
// 3-retry exponential-backoff behavior (G19 backward compatibility).
func TestReporterDefaults(t *testing.T) {
	_, apiClient := createTestClient()
	r := NewHeartbeatReporter(apiClient, &mockProbeScheduler{})
	if r.maxRetries != defaultMaxRetries {
		t.Fatalf("default maxRetries = %d, want %d", r.maxRetries, defaultMaxRetries)
	}
	if r.backoff != "exponential" {
		t.Fatalf("default backoff = %q, want %q", r.backoff, "exponential")
	}
	if _, ok := r.outcomeListener.(emptyOutcomeListener); !ok {
		t.Fatalf("default listener should be the no-op emptyOutcomeListener")
	}
}

// TestWithReconnectConfig verifies reconnect config is applied with zero-value fallback (G19).
func TestWithReconnectConfig(t *testing.T) {
	_, apiClient := createTestClient()
	// Full config
	r := NewHeartbeatReporter(apiClient, &mockProbeScheduler{},
		WithReconnectConfig(5, 10, "linear"))
	if r.maxRetries != 5 || r.retryBase != 10*time.Second || r.backoff != "linear" {
		t.Fatalf("reconnect config not applied: %+v", r)
	}

	// Zero-value config keeps defaults
	r2 := NewHeartbeatReporter(apiClient, &mockProbeScheduler{}, WithReconnectConfig(0, 0, ""))
	if r2.maxRetries != defaultMaxRetries || r2.backoff != "exponential" {
		t.Fatalf("zero-value reconnect should keep defaults: %+v", r2)
	}

	// Invalid backoff string is ignored
	r3 := NewHeartbeatReporter(apiClient, &mockProbeScheduler{}, WithReconnectConfig(0, 0, "bogus"))
	if r3.backoff != "exponential" {
		t.Fatalf("invalid backoff should fall back to exponential, got %q", r3.backoff)
	}
}

// recordingListener captures outcome callbacks for G16 verification.
type recordingListener struct {
	successes int
	failures  int
}

func (r *recordingListener) RecordHeartbeatSuccess() { r.successes++ }
func (r *recordingListener) RecordHeartbeatFailure() { r.failures++ }

// TestOutcomeListenerNotified verifies the degraded-mode listener is wired (G16).
func TestOutcomeListenerNotified(t *testing.T) {
	_, apiClient := createTestClient()
	lis := &recordingListener{}
	r := NewHeartbeatReporter(apiClient, &mockProbeScheduler{}, WithOutcomeListener(lis))
	if r.outcomeListener != lis {
		t.Fatalf("listener not wired")
	}
	r.outcomeListener.RecordHeartbeatSuccess()
	r.outcomeListener.RecordHeartbeatFailure()
	if lis.successes != 1 || lis.failures != 1 {
		t.Fatalf("listener callbacks not recorded: %+v", lis)
	}
}

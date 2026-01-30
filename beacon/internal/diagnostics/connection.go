package diagnostics

import (
	"time"
)

// ConnectionStatus contains connection retry status information
type ConnectionStatus struct {
	Status           string     `json:"status"`            // connected, connecting, disconnected
	LastSuccess      *time.Time `json:"last_success,omitempty"`
	LastFailure      *time.Time `json:"last_failure,omitempty"`
	FailureReason    string     `json:"failure_reason,omitempty"`
	RetryCount       int        `json:"retry_count"`
	BackoffSeconds   int        `json:"backoff_seconds"`
	NextRetry        *time.Time `json:"next_retry,omitempty"`
	QueueSize        int        `json:"queue_size"`
	OldestQueuedItem *time.Time `json:"oldest_queued_item,omitempty"`
}

// collectConnectionStatus collects connection status information
// NOTE: This requires integration with reporter.Reporter to get actual connection status.
// For Story 3.10, this returns placeholder values indicating the feature is partially implemented.
// Future work: Integrate with reporter.Reporter.GetConnectionStatus() method.
func (c *collector) collectConnectionStatus() (*ConnectionStatus, error) {
	status := &ConnectionStatus{
		Status:           "unknown",
		LastSuccess:      nil,
		LastFailure:      nil,
		FailureReason:    "feature requires reporter integration",
		RetryCount:       0,
		BackoffSeconds:   0,
		NextRetry:        nil,
		QueueSize:        0,
		OldestQueuedItem: nil,
	}

	return status, nil
}

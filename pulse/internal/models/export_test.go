package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestExportTask_IsValidFormat covers valid and invalid export formats.
func TestExportTask_IsValidFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   bool
	}{
		{"csv format", "csv", true},
		{"xlsx format", "xlsx", true},
		{"json format unsupported", "json", false},
		{"empty format", "", false},
		{"uppercase not accepted", "CSV", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExportTask{Format: tt.format}
			assert.Equal(t, tt.want, e.IsValidFormat())
		})
	}
}

// TestExportTask_IsValidStatus covers valid and invalid task statuses.
func TestExportTask_IsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"pending", "pending", true},
		{"processing", "processing", true},
		{"completed", "completed", true},
		{"failed", "failed", true},
		{"unknown status", "queued", false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExportTask{Status: tt.status}
			assert.Equal(t, tt.want, e.IsValidStatus())
		})
	}
}

// TestExportTask_CanTransitionTo covers the full status state-machine,
// including that completed/failed are terminal.
func TestExportTask_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus string
		toStatus   string
		want       bool
	}{
		// pending transitions
		{"pending to processing allowed", "pending", "processing", true},
		{"pending to failed allowed", "pending", "failed", true},
		{"pending to completed forbidden", "pending", "completed", false},
		{"pending to pending forbidden", "pending", "pending", false},
		// processing transitions
		{"processing to completed allowed", "processing", "completed", true},
		{"processing to failed allowed", "processing", "failed", true},
		{"processing to pending forbidden", "processing", "pending", false},
		// completed is terminal
		{"completed to anything forbidden", "completed", "processing", false},
		// failed is terminal
		{"failed to anything forbidden", "failed", "processing", false},
		// unknown source status
		{"unknown source forbidden", "queued", "processing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExportTask{Status: tt.fromStatus}
			assert.Equal(t, tt.want, e.CanTransitionTo(tt.toStatus))
		})
	}
}

// TestExportTask_StatusPredicates covers IsCompleted/IsFailed/IsProcessing.
func TestExportTask_StatusPredicates(t *testing.T) {
	e := &ExportTask{}

	e.Status = "completed"
	assert.True(t, e.IsCompleted())
	assert.False(t, e.IsFailed())
	assert.False(t, e.IsProcessing())

	e.Status = "failed"
	assert.False(t, e.IsCompleted())
	assert.True(t, e.IsFailed())
	assert.False(t, e.IsProcessing())

	e.Status = "processing"
	assert.False(t, e.IsCompleted())
	assert.False(t, e.IsFailed())
	assert.True(t, e.IsProcessing())
}

// TestExportTask_GetDuration covers both the completed and not-completed branches.
func TestExportTask_GetDuration(t *testing.T) {
	t.Run("returns duration when completed", func(t *testing.T) {
		created := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		completed := created.Add(5 * time.Minute)
		e := &ExportTask{CreatedAt: created, CompletedAt: &completed}
		assert.Equal(t, 5*time.Minute, e.GetDuration())
	})

	t.Run("returns zero when not completed (nil pointer)", func(t *testing.T) {
		created := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		e := &ExportTask{CreatedAt: created} // CompletedAt nil
		assert.Equal(t, time.Duration(0), e.GetDuration())
	})
}

// TestMetricUnit covers the metric-to-unit lookup and the default branch.
func TestMetricUnit(t *testing.T) {
	tests := []struct {
		name   string
		metric string
		want   string
	}{
		{"latency", "latency", "ms"},
		{"jitter", "jitter", "ms"},
		{"packet_loss_rate", "packet_loss_rate", "%"},
		{"unknown metric returns empty", "cpu_usage", ""},
		{"empty metric returns empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MetricUnit(tt.metric))
		})
	}
}

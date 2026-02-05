package models

import (
	"time"
)

// ExportTask represents a data export task
type ExportTask struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	NodeIDs    []string  `json:"node_ids"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Metrics    []string  `json:"metrics"`
	Format     string    `json:"format"` // csv, xlsx
	Status     string    `json:"status"` // pending, processing, completed, failed
	FilePath   string    `json:"file_path,omitempty"`
	FileSize   int64     `json:"file_size,omitempty"`
	RecordCount int      `json:"record_count,omitempty"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// IsValidFormat checks if the export format is valid
func (e *ExportTask) IsValidFormat() bool {
	switch e.Format {
	case "csv", "xlsx":
		return true
	default:
		return false
	}
}

// IsValidStatus checks if the status is valid
func (e *ExportTask) IsValidStatus() bool {
	switch e.Status {
	case "pending", "processing", "completed", "failed":
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if a status transition is allowed
func (e *ExportTask) CanTransitionTo(newStatus string) bool {
	// Valid transitions:
	// pending -> processing
	// pending -> failed
	// processing -> completed
	// processing -> failed
	switch e.Status {
	case "pending":
		return newStatus == "processing" || newStatus == "failed"
	case "processing":
		return newStatus == "completed" || newStatus == "failed"
	default:
		return false // No transitions from completed or failed
	}
}

// IsCompleted returns true if the export task is completed
func (e *ExportTask) IsCompleted() bool {
	return e.Status == "completed"
}

// IsFailed returns true if the export task has failed
func (e *ExportTask) IsFailed() bool {
	return e.Status == "failed"
}

// IsProcessing returns true if the export is currently being processed
func (e *ExportTask) IsProcessing() bool {
	return e.Status == "processing"
}

// GetDuration returns the duration of the export task
func (e *ExportTask) GetDuration() time.Duration {
	if e.CompletedAt != nil {
		return e.CompletedAt.Sub(e.CreatedAt)
	}
	return 0
}

// ExportMetricsRow represents a single row in the export file
type ExportMetricsRow struct {
	Timestamp string  `json:"timestamp"`
	NodeID    string  `json:"node_id"`
	Region    string  `json:"region"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
}

// MetricUnit returns the unit for a given metric
func MetricUnit(metric string) string {
	switch metric {
	case "latency":
		return "ms"
	case "packet_loss_rate":
		return "%"
	case "jitter":
		return "ms"
	default:
		return ""
	}
}

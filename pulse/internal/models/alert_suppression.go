package models

import "time"

// AlertSuppression represents an alert suppression record
type AlertSuppression struct {
	ID             string    `json:"id" db:"id"`
	NodeID         string    `json:"node_id" db:"node_id"`             // Node to suppress alerts for
	Metric         string    `json:"metric" db:"metric"`                 // latency, packet_loss_rate, jitter
	SuppressedUntil time.Time `json:"suppressed_until" db:"suppressed_until"` // Suppression window end time
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// SuppressionData represents suppression data in response
type SuppressionData struct {
	Suppression *AlertSuppression `json:"suppression"`
}

// CreateSuppressionResponse represents successful suppression creation response
type CreateSuppressionResponse struct {
	Data      SuppressionData `json:"data"`
	Message   string          `json:"message"`
	Timestamp string          `json:"timestamp"`
}

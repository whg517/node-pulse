package models

import "time"

// AlertEvent represents a triggered alert event
type AlertEvent struct {
	ID           string    `json:"id" db:"id"`
	NodeID       string    `json:"node_id" db:"node_id"`
	Metric       string    `json:"metric" db:"metric"`               // latency, packet_loss_rate, jitter
	Threshold    float64   `json:"threshold" db:"threshold"`         // Alert threshold value
	CurrentValue float64   `json:"current_value" db:"current_value"` // Actual metric value
	Level        string    `json:"level" db:"level"`                 // P0, P1, P2
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// AlertEventData represents alert event data in response
type AlertEventData struct {
	AlertEvent *AlertEvent `json:"alert_event"`
}

// CreateAlertEventResponse represents successful alert event creation response
type CreateAlertEventResponse struct {
	Data      AlertEventData `json:"data"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
}

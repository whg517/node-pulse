package models

import "time"

// Alert represents an alert rule in the system
type Alert struct {
	ID        string    `json:"id" db:"id"`
	Metric    string    `json:"metric" db:"metric"`              // latency, packet_loss_rate, jitter
	Threshold float64   `json:"threshold" db:"threshold"`        // Alert threshold value
	Level     string    `json:"level" db:"level"`                // P0, P1, P2
	NodeID    *string   `json:"node_id,omitempty" db:"node_id"`  // NULL for global rules
	Enabled   bool      `json:"enabled" db:"enabled"`            // true/false
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateAlertRequest represents request to create a new alert rule
type CreateAlertRequest struct {
	Metric    string  `json:"metric" binding:"required,oneof=latency packet_loss_rate jitter"`
	Threshold float64 `json:"threshold" binding:"required,gt=0"`
	Level     string  `json:"level" binding:"required,oneof=P0 P1 P2"`
	NodeID    *string `json:"node_id,omitempty"` // NULL for global rules
	Enabled   *bool   `json:"enabled,omitempty"` // Default to true if not provided
}

// UpdateAlertRequest represents request to update an alert rule
type UpdateAlertRequest struct {
	Metric    *string  `json:"metric,omitempty" binding:"omitempty,oneof=latency packet_loss_rate jitter"`
	Threshold *float64 `json:"threshold,omitempty" binding:"omitempty,gt=0"`
	Level     *string  `json:"level,omitempty" binding:"omitempty,oneof=P0 P1 P2"`
	NodeID    *string  `json:"node_id,omitempty"`
	Enabled   *bool    `json:"enabled,omitempty"`
}

// AlertData represents alert data in response
type AlertData struct {
	Alert *Alert `json:"alert"`
}

// CreateAlertResponse represents successful alert creation response
type CreateAlertResponse struct {
	Data      AlertData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
}

// AlertsListData represents list of alerts in response
type AlertsListData struct {
	Alerts []*Alert `json:"alerts"`
}

// GetAlertsResponse represents successful alerts retrieval response
type GetAlertsResponse struct {
	Data      AlertsListData `json:"data"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
}

// GetAlertByIDResponse represents successful single alert retrieval response
type GetAlertByIDResponse struct {
	Data      AlertData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
}

// UpdateAlertResponse represents successful alert update response
type UpdateAlertResponse struct {
	Data      AlertData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
}

// DeleteAlertResponse represents successful alert deletion response
type DeleteAlertResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

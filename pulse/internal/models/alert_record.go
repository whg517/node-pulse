package models

import "time"

// AlertRecord represents a tracked alert with lifecycle status
type AlertRecord struct {
	ID           string      `json:"id"`
	AlertEventID string      `json:"alert_event_id"`
	NodeID       string      `json:"node_id"`
	Metric       string      `json:"metric"`
	Level        string      `json:"level"`
	Status       string      `json:"status"` // pending, in_progress, resolved
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Notes        []AlertNote `json:"notes,omitempty"`
}

// AlertNote represents an operator note attached to an alert record.
type AlertNote struct {
	ID        string    `json:"id"`
	AlertID   string    `json:"alert_id"`
	UserID    string    `json:"user_id,omitempty"`
	UserName  string    `json:"user_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// IsValidStatus checks if the status is valid
func (a *AlertRecord) IsValidStatus() bool {
	switch a.Status {
	case "pending", "in_progress", "resolved":
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if a status transition is allowed
func (a *AlertRecord) CanTransitionTo(newStatus string) bool {
	// Valid transitions:
	// pending → in_progress
	// pending → resolved
	// in_progress → resolved
	// resolved → (no transitions allowed in MVP)
	switch a.Status {
	case "pending":
		return newStatus == "in_progress" || newStatus == "resolved"
	case "in_progress":
		return newStatus == "resolved"
	case "resolved":
		return false // Cannot reopen in MVP
	default:
		return false
	}
}

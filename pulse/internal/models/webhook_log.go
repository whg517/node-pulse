package models

import "time"

// WebhookLog represents a webhook delivery log entry
type WebhookLog struct {
	ID           string    `json:"id" db:"id"`
	WebhookID    string    `json:"webhook_id" db:"webhook_id"`
	AlertEventID string    `json:"alert_event_id" db:"alert_event_id"`
	Status       string    `json:"status" db:"status"`               // success, failure, retrying
	RetryCount   int       `json:"retry_count" db:"retry_count"`     // Number of delivery attempts
	ErrorMessage string    `json:"error_message" db:"error_message"` // Error details if failed
	SentAt       time.Time `json:"sent_at" db:"sent_at"`             // Last delivery attempt time
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // When the log row was written
}

// WebhookLogData represents webhook log data in response
type WebhookLogData struct {
	WebhookLog *WebhookLog `json:"webhook_log"`
}

// CreateWebhookLogResponse represents successful webhook log creation response
type CreateWebhookLogResponse struct {
	Data      WebhookLogData `json:"data"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
}

package models

import "time"

// Webhook represents a webhook configuration in the system
type Webhook struct {
	ID          string         `json:"id" db:"id"`
	URL         string         `json:"url" db:"url"`
	EventFormat map[string]any `json:"event_format,omitempty" db:"event_format"`
	Enabled     bool           `json:"enabled" db:"enabled"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
}

// CreateWebhookRequest represents request to create a new webhook
type CreateWebhookRequest struct {
	URL         string         `json:"url" binding:"required,url"`
	EventFormat map[string]any `json:"event_format,omitempty"`
	Enabled     *bool          `json:"enabled,omitempty"` // Default to true if not provided
}

// UpdateWebhookRequest represents request to update a webhook
type UpdateWebhookRequest struct {
	URL         *string         `json:"url,omitempty" binding:"omitempty,url"`
	EventFormat *map[string]any `json:"event_format,omitempty"`
	Enabled     *bool           `json:"enabled,omitempty"`
}

// PreviewWebhookEventRequest represents a request to render a webhook payload preview.
type PreviewWebhookEventRequest struct {
	EventFormat map[string]any `json:"event_format,omitempty"`
}

// WebhookData represents webhook data in response
type WebhookData struct {
	Webhook *Webhook `json:"webhook"`
}

// WebhookPreviewData represents a rendered webhook payload preview.
type WebhookPreviewData struct {
	Payload map[string]any `json:"payload"`
}

// PreviewWebhookEventResponse represents a successful webhook payload preview response.
type PreviewWebhookEventResponse struct {
	Data      WebhookPreviewData `json:"data"`
	Message   string             `json:"message"`
	Timestamp string             `json:"timestamp"`
}

// CreateWebhookResponse represents successful webhook creation response
type CreateWebhookResponse struct {
	Data      WebhookData `json:"data"`
	Message   string      `json:"message"`
	Timestamp string      `json:"timestamp"`
}

// WebhooksListData represents list of webhooks in response
type WebhooksListData struct {
	Webhooks []*Webhook `json:"webhooks"`
}

// GetWebhooksResponse represents successful webhooks retrieval response
type GetWebhooksResponse struct {
	Data      WebhooksListData `json:"data"`
	Message   string           `json:"message"`
	Timestamp string           `json:"timestamp"`
}

// UpdateWebhookResponse represents successful webhook update response
type UpdateWebhookResponse struct {
	Data      WebhookData `json:"data"`
	Message   string      `json:"message"`
	Timestamp string      `json:"timestamp"`
}

// DeleteWebhookResponse represents successful webhook deletion response
type DeleteWebhookResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// DefaultEventFormat defines the default webhook event format
var DefaultEventFormat = map[string]any{
	"version": "1.0",
	"alert": map[string]any{
		"id":            "{{.AlertID}}",
		"metric":        "{{.Metric}}",
		"threshold":     "{{.Threshold}}",
		"current_value": "{{.CurrentValue}}",
		"level":         "{{.Level}}",
		"node_id":       "{{.NodeID}}",
		"node_name":     "{{.NodeName}}",
		"triggered_at":  "{{.TriggeredAt}}",
	},
	"links": map[string]any{
		"alert_details": "{{.BaseURL}}/nodes/{{.NodeID}}",
		"dashboard":     "{{.BaseURL}}",
	},
}

package models

import "time"

// AlertRoutingRule routes alerts to a specific webhook based on optional metric,
// severity, and node criteria. Null/empty criteria act as wildcards. A webhook
// with no rules keeps the default "receive everything" behavior (ADR-002 Tier-1).
type AlertRoutingRule struct {
	ID          string    `json:"id"`
	OwnerUserID string    `json:"owner_user_id"`
	Name        string    `json:"name"`
	Enabled     bool      `json:"enabled"`
	Metric      string    `json:"metric,omitempty"`     // optional: latency, packet_loss_rate, jitter
	Severities  []string  `json:"severities,omitempty"` // optional: subset of P0/P1/P2
	NodeID      string    `json:"node_id,omitempty"`    // optional: specific node
	WebhookID   string    `json:"webhook_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

package models

import "time"

// HeartbeatRequest represents beacon heartbeat data request
type HeartbeatRequest struct {
	NodeID          string  `json:"node_id" binding:"required"`
	ProbeID         string  `json:"probe_id" binding:"required"`
	LatencyMs       float64 `json:"latency_ms" binding:"required"`
	PacketLossRate  float64 `json:"packet_loss_rate" binding:"required"`
	JitterMs        float64 `json:"jitter_ms" binding:"required"`
	Timestamp       string  `json:"timestamp" binding:"required"`
}

// HeartbeatSuccessResponse represents successful heartbeat response
type HeartbeatSuccessResponse struct {
	Data      HeartbeatData `json:"data"`
	Message   string        `json:"message"`
	Timestamp string        `json:"timestamp"`
}

// HeartbeatData represents heartbeat response data
type HeartbeatData struct {
	Received  bool      `json:"received"`
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
}

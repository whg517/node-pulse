package models

import "time"

// Probe represents a probe configuration for monitoring
type Probe struct {
	ID              string    `json:"id" db:"id"`
	NodeID          string    `json:"node_id" db:"node_id"`
	Type            string    `json:"type" db:"type"`                           // TCP or UDP
	Target          string    `json:"target" db:"target"`                       // IP or domain
	Port            int       `json:"port" db:"port"`                           // 1-65535
	IntervalSeconds int       `json:"interval_seconds" db:"interval_seconds"`   // 60-300
	Count           int       `json:"count" db:"count"`                         // 1-100
	TimeoutSeconds  int       `json:"timeout_seconds" db:"timeout_seconds"`     // 1-30
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProbeRequest represents request to create a new probe
type CreateProbeRequest struct {
	NodeID          string `json:"node_id" binding:"required"`
	Type            string `json:"type" binding:"required"`
	Target          string `json:"target" binding:"required"`
	Port            int    `json:"port" binding:"required,min=1,max=65535"`
	IntervalSeconds int    `json:"interval_seconds" binding:"required,min=60,max=300"`
	Count           int    `json:"count" binding:"required,min=1,max=100"`
	TimeoutSeconds  int    `json:"timeout_seconds" binding:"required,min=1,max=30"`
}

// UpdateProbeRequest represents request to update a probe
type UpdateProbeRequest struct {
	Type            *string `json:"type,omitempty"`
	Target          *string `json:"target,omitempty"`
	Port            *int    `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	IntervalSeconds *int    `json:"interval_seconds,omitempty" binding:"omitempty,min=60,max=300"`
	Count           *int    `json:"count,omitempty" binding:"omitempty,min=1,max=100"`
	TimeoutSeconds  *int    `json:"timeout_seconds,omitempty" binding:"omitempty,min=1,max=30"`
}

// CreateProbeResponse represents successful probe creation response
type CreateProbeResponse struct {
	Data      ProbeData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
}

// ProbeData represents probe data in response
type ProbeData struct {
	Probe *Probe `json:"probe"`
}

// GetProbesResponse represents successful probes list retrieval response
type GetProbesResponse struct {
	Data      ProbesListData `json:"data"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
}

// ProbesListData represents probes data in list response
type ProbesListData struct {
	Probes []*Probe `json:"probes"`
}

// GetProbeResponse represents successful probe retrieval response
type GetProbeResponse struct {
	Data      ProbeData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
}

// UpdateProbeResponse represents successful probe update response
type UpdateProbeResponse struct {
	Data      ProbeData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
}

// DeleteProbeResponse represents successful probe deletion response
type DeleteProbeResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

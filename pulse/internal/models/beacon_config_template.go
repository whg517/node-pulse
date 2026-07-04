package models

import "time"

// BeaconConfigTemplate is a reusable, server-owned beacon probe configuration.
// Templates are owned by a user and applied to beacons via the existing
// POST /beacons/:id/config flow (which produces a new config version).
type BeaconConfigTemplate struct {
	ID              string    `json:"id"`
	OwnerUserID     string    `json:"owner_user_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	Probes          []any     `json:"probes"` // []BeaconProbeConfig, kept as any for json flexibility
	IntervalSeconds int       `json:"interval_seconds"`
	TimeoutSeconds  int       `json:"timeout_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

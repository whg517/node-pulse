package models

import "time"

// ReportSchedule is a server-side recurring report definition (ADR-001).
// The scheduler polls due schedules, generates an export (CSV or PDF), and
// emails it to the owner (or recipient_email when set).
type ReportSchedule struct {
	ID             string     `json:"id"`
	OwnerUserID    string     `json:"owner_user_id"`
	Name           string     `json:"name"`
	Frequency      string     `json:"frequency"` // daily | weekly | monthly
	TimeOfDay      string     `json:"time_of_day"`
	NodeIDs        []string   `json:"node_ids"`
	Metrics        []string   `json:"metrics"`
	Format         string     `json:"format"` // csv | pdf
	RecipientEmail string     `json:"recipient_email,omitempty"`
	Enabled        bool       `json:"enabled"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

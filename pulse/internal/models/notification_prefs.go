package models

import "time"

// NotificationPrefs holds a user's server-side notification preferences
// (F4 Phase 2). The browser-notification preferences (F4 Phase 1) live in
// the frontend settingsStore; these compose with them — the server-side
// floor controls email delivery, the client-side floor controls the desktop
// notification.
type NotificationPrefs struct {
	UserID         string    `json:"user_id"`
	EmailEnabled   bool      `json:"email_enabled"`
	MinAlertLevel  string    `json:"min_alert_level"` // P0 | P1 | P2
	NotifyEmail    *string   `json:"notify_email,omitempty"` // override; nil = use profile email
	UpdatedAt      time.Time `json:"updated_at"`
}

// UpdateNotificationPrefsRequest is the PUT body for /auth/notification-prefs.
// All fields optional; omitting a field keeps the current value.
type UpdateNotificationPrefsRequest struct {
	EmailEnabled  *bool   `json:"email_enabled,omitempty"`
	MinAlertLevel *string `json:"min_alert_level,omitempty"`
	NotifyEmail   *string `json:"notify_email,omitempty"`
}

// DefaultMinAlertLevel is the server-side default floor when a user has no
// row or the field is empty. Matches the F4 Phase 1 client default.
const DefaultMinAlertLevel = "P1"

// IsValidAlertLevel reports whether s is a recognized severity (P0/P1/P2).
func IsValidAlertLevel(s string) bool {
	return s == "P0" || s == "P1" || s == "P2"
}

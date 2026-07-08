package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// NotificationPrefsRepository handles per-user notification preferences (F4 P2).
type NotificationPrefsRepository interface {
	// GetByUserID returns the user's prefs, or a default-valued row when none
	// exists yet (callers should treat absence as "defaults apply" rather than
	// an error).
	GetByUserID(ctx context.Context, userID string) (*models.NotificationPrefs, error)
	// Upsert inserts or updates the user's prefs.
	Upsert(ctx context.Context, userID string, req *models.UpdateNotificationPrefsRequest) (*models.NotificationPrefs, error)
	// ListSubscribersForLevel returns the prefs rows of every user who has
	// email notifications enabled at or below the given severity floor. Used
	// by the alert engine to fan out email on a new alert.
	ListSubscribersForLevel(ctx context.Context, level string) ([]*models.NotificationPrefs, error)
}

type notificationPrefsRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationPrefsRepository constructs the repository.
func NewNotificationPrefsRepository(pool *pgxpool.Pool) NotificationPrefsRepository {
	return &notificationPrefsRepository{pool: pool}
}

// GetByUserID returns the user's prefs or a default row when not yet set.
func (r *notificationPrefsRepository) GetByUserID(ctx context.Context, userID string) (*models.NotificationPrefs, error) {
	p := &models.NotificationPrefs{
		UserID:        userID,
		EmailEnabled:  true,
		MinAlertLevel: models.DefaultMinAlertLevel,
	}
	// COALESCE picks up the column value or falls back to the defaults above.
	query := `
		SELECT email_enabled, COALESCE(min_alert_level, $2), notify_email, updated_at
		FROM user_notification_prefs
		WHERE user_id = $1
	`
	var notifyEmail *string
	err := r.pool.QueryRow(ctx, query, userID, models.DefaultMinAlertLevel).Scan(
		&p.EmailEnabled, &p.MinAlertLevel, &notifyEmail, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No row yet — return defaults (already populated above).
			return p, nil
		}
		return nil, fmt.Errorf("failed to get notification prefs: %w", err)
	}
	p.NotifyEmail = notifyEmail
	return p, nil
}

// Upsert inserts or updates the user's prefs.
func (r *notificationPrefsRepository) Upsert(ctx context.Context, userID string, req *models.UpdateNotificationPrefsRequest) (*models.NotificationPrefs, error) {
	// Start from the current state (or defaults), then apply the patch.
	current, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.EmailEnabled != nil {
		current.EmailEnabled = *req.EmailEnabled
	}
	if req.MinAlertLevel != nil && models.IsValidAlertLevel(*req.MinAlertLevel) {
		current.MinAlertLevel = *req.MinAlertLevel
	}
	if req.NotifyEmail != nil {
		// Empty string clears the override; nil keeps current. To allow
		// clearing we treat "" as "use profile email" (NULL).
		if *req.NotifyEmail == "" {
			current.NotifyEmail = nil
		} else {
			e := *req.NotifyEmail
			current.NotifyEmail = &e
		}
	}

	query := `
		INSERT INTO user_notification_prefs (user_id, email_enabled, min_alert_level, notify_email, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			email_enabled = EXCLUDED.email_enabled,
			min_alert_level = EXCLUDED.min_alert_level,
			notify_email = EXCLUDED.notify_email,
			updated_at = NOW()
		RETURNING email_enabled, min_alert_level, notify_email, updated_at
	`
	var notifyEmail *string
	err = r.pool.QueryRow(ctx, query,
		userID, current.EmailEnabled, current.MinAlertLevel, current.NotifyEmail,
	).Scan(&current.EmailEnabled, &current.MinAlertLevel, &notifyEmail, &current.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert notification prefs: %w", err)
	}
	current.NotifyEmail = notifyEmail
	return current, nil
}

// ListSubscribersForLevel returns prefs for users whose email notifications
// are enabled and whose min_alert_level is at or below the given severity.
// Severity rank: P0 (most severe) < P1 < P2. A user with min_level=P1 wants
// P0 and P1 alerts; an alert of `level` matches when level_rank <= min_rank.
func (r *notificationPrefsRepository) ListSubscribersForLevel(ctx context.Context, level string) ([]*models.NotificationPrefs, error) {
	// Validate the level; default to P1 if unrecognized so we don't spam.
	if !models.IsValidAlertLevel(level) {
		level = models.DefaultMinAlertLevel
	}
	query := `
		SELECT p.user_id, p.email_enabled, p.min_alert_level, p.notify_email, p.updated_at
		FROM user_notification_prefs p
		WHERE p.email_enabled = true
		  AND CASE p.min_alert_level
		      WHEN 'P0' THEN $1 = 'P0'
		      WHEN 'P1' THEN $1 IN ('P0', 'P1')
		      ELSE true
		  END
	`
	rows, err := r.pool.Query(ctx, query, level)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscribers: %w", err)
	}
	defer rows.Close()

	var out []*models.NotificationPrefs
	for rows.Next() {
		p := &models.NotificationPrefs{}
		var notifyEmail *string
		if err := rows.Scan(&p.UserID, &p.EmailEnabled, &p.MinAlertLevel, &notifyEmail, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.NotifyEmail = notifyEmail
		out = append(out, p)
	}
	return out, rows.Err()
}

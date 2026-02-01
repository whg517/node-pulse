package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

var (
	ErrSuppressionNotFound = errors.New("suppression not found")
)

// AlertSuppressionsQuerier defines database operations for alert suppressions
type AlertSuppressionsQuerier interface {
	CheckSuppression(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error)
	CreateOrUpdateSuppression(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error
	DeleteExpiredSuppressions(ctx context.Context) (int64, error)
}

type alertSuppressionsQuerier struct {
	pool *pgxpool.Pool
}

// NewAlertSuppressionsQuerier creates a new AlertSuppressionsQuerier
func NewAlertSuppressionsQuerier(pool *pgxpool.Pool) AlertSuppressionsQuerier {
	return &alertSuppressionsQuerier{pool: pool}
}

// CheckSuppression checks if there's an active suppression for a node and metric
func (q *alertSuppressionsQuerier) CheckSuppression(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error) {
	query := `
		SELECT id, node_id, metric, suppressed_until, created_at, updated_at
		FROM alert_suppressions
		WHERE node_id = $1 AND metric = $2
		ORDER BY suppressed_until DESC
		LIMIT 1
	`

	var suppression models.AlertSuppression
	err := q.pool.QueryRow(ctx, query, nodeID, metric).Scan(
		&suppression.ID,
		&suppression.NodeID,
		&suppression.Metric,
		&suppression.SuppressedUntil,
		&suppression.CreatedAt,
		&suppression.UpdatedAt,
	)

	if err != nil {
		return nil, ErrSuppressionNotFound
	}

	return &suppression, nil
}

// CreateOrUpdateSuppression creates or updates a suppression record (upsert)
func (q *alertSuppressionsQuerier) CreateOrUpdateSuppression(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error {
	query := `
		INSERT INTO alert_suppressions (id, node_id, metric, suppressed_until, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (node_id, metric)
		DO UPDATE SET
			suppressed_until = EXCLUDED.suppressed_until,
			updated_at = NOW()
	`

	id := uuid.New().String()
	_, err := q.pool.Exec(ctx, query, id, nodeID, metric, suppressedUntil)

	if err != nil {
		return errors.New("failed to create or update suppression: " + err.Error())
	}

	return nil
}

// DeleteExpiredSuppressions deletes suppression records that have expired
func (q *alertSuppressionsQuerier) DeleteExpiredSuppressions(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM alert_suppressions
		WHERE suppressed_until < NOW()
	`

	result, err := q.pool.Exec(ctx, query)
	if err != nil {
		return 0, errors.New("failed to delete expired suppressions: " + err.Error())
	}

	return result.RowsAffected(), nil
}

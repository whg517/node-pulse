package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// AlertEventsQuerier defines database operations for alert events
type AlertEventsQuerier interface {
	CreateAlertEvent(ctx context.Context, event *models.AlertEvent) error
}

type alertEventsQuerier struct {
	pool *pgxpool.Pool
}

// NewAlertEventsQuerier creates a new AlertEventsQuerier
func NewAlertEventsQuerier(pool *pgxpool.Pool) AlertEventsQuerier {
	return &alertEventsQuerier{pool: pool}
}

// CreateAlertEvent creates a new alert event in the database
func (q *alertEventsQuerier) CreateAlertEvent(ctx context.Context, event *models.AlertEvent) error {
	event.ID = uuid.New().String()

	query := `
		INSERT INTO alert_events (id, node_id, metric, threshold, current_value, level, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := q.pool.Exec(ctx, query,
		event.ID,
		event.NodeID,
		event.Metric,
		event.Threshold,
		event.CurrentValue,
		event.Level,
		event.CreatedAt,
	)

	if err != nil {
		return errors.New("failed to create alert event: " + err.Error())
	}

	return nil
}

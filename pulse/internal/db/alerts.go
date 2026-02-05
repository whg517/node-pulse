package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// AlertQuerier defines alert database operations
type AlertQuerier interface {
	CreateAlert(ctx context.Context, alert *models.Alert) error
	GetAlerts(ctx context.Context, nodeID *string) ([]*models.Alert, error)
	GetAlertByID(ctx context.Context, id string) (*models.Alert, error)
	UpdateAlert(ctx context.Context, id string, update *models.UpdateAlertRequest) (*models.Alert, error)
	DeleteAlert(ctx context.Context, id string) error
}

type alertQuerier struct {
	pool *pgxpool.Pool
}

// NewAlertQuerier creates a new alert querier
func NewAlertQuerier(pool *pgxpool.Pool) AlertQuerier {
	return &alertQuerier{pool: pool}
}

// CreateAlert creates a new alert rule
func (q *alertQuerier) CreateAlert(ctx context.Context, alert *models.Alert) error {
	alert.ID = uuid.New().String()

	query := `
		INSERT INTO alerts (id, metric, threshold, level, node_id, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at
	`

	err := q.pool.QueryRow(ctx, query,
		alert.ID, alert.Metric, alert.Threshold, alert.Level,
		alert.NodeID, alert.Enabled,
	).Scan(&alert.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetAlerts retrieves all alert rules, optionally filtered by node_id
func (q *alertQuerier) GetAlerts(ctx context.Context, nodeID *string) ([]*models.Alert, error) {
	var query string
	var args []interface{}

	if nodeID != nil {
		query = `
			SELECT id, metric, threshold, level, node_id, enabled, created_at
			FROM alerts
			WHERE node_id = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{*nodeID}
	} else {
		query = `
			SELECT id, metric, threshold, level, node_id, enabled, created_at
			FROM alerts
			ORDER BY created_at DESC
		`
	}

	rows, err := q.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := []*models.Alert{}
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.Metric, &alert.Threshold, &alert.Level,
			&alert.NodeID, &alert.Enabled, &alert.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

// GetAlertByID retrieves a single alert rule by ID
func (q *alertQuerier) GetAlertByID(ctx context.Context, id string) (*models.Alert, error) {
	alert := &models.Alert{}

	query := `
		SELECT id, metric, threshold, level, node_id, enabled, created_at
		FROM alerts
		WHERE id = $1
	`

	err := q.pool.QueryRow(ctx, query, id).Scan(
		&alert.ID, &alert.Metric, &alert.Threshold, &alert.Level,
		&alert.NodeID, &alert.Enabled, &alert.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("alert not found")
		}
		return nil, err
	}

	return alert, nil
}

// UpdateAlert updates an existing alert rule
func (q *alertQuerier) UpdateAlert(ctx context.Context, id string, update *models.UpdateAlertRequest) (*models.Alert, error) {
	// Build dynamic UPDATE query based on provided fields
	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	if update.Metric != nil {
		setClauses = append(setClauses, fmt.Sprintf("metric = $%d", argCount))
		args = append(args, *update.Metric)
		argCount++
	}

	if update.Threshold != nil {
		setClauses = append(setClauses, fmt.Sprintf("threshold = $%d", argCount))
		args = append(args, *update.Threshold)
		argCount++
	}

	if update.Level != nil {
		setClauses = append(setClauses, fmt.Sprintf("level = $%d", argCount))
		args = append(args, *update.Level)
		argCount++
	}

	if update.NodeID != nil {
		setClauses = append(setClauses, fmt.Sprintf("node_id = $%d", argCount))
		args = append(args, *update.NodeID)
		argCount++
	}

	if update.Enabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("enabled = $%d", argCount))
		args = append(args, *update.Enabled)
		argCount++
	}

	if len(setClauses) == 0 {
		return q.GetAlertByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE alerts
		SET %s
		WHERE id = $%d
		RETURNING id, metric, threshold, level, node_id, enabled, created_at
	`, strings.Join(setClauses, ", "), argCount)

	args = append(args, id)

	alert := &models.Alert{}
	err := q.pool.QueryRow(ctx, query, args...).Scan(
		&alert.ID, &alert.Metric, &alert.Threshold, &alert.Level,
		&alert.NodeID, &alert.Enabled, &alert.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("alert not found")
		}
		return nil, err
	}

	return alert, nil
}

// DeleteAlert deletes an alert rule
func (q *alertQuerier) DeleteAlert(ctx context.Context, id string) error {
	query := `DELETE FROM alerts WHERE id = $1`

	result, err := q.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("alert not found")
	}

	return nil
}

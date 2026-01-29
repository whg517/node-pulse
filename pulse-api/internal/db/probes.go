package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

var (
	ErrProbeNotFound      = errors.New("probe not found")
	ErrInvalidNodeType    = errors.New("invalid node type for probe")
)

// ProbesQuerier defines interface for probe database operations
type ProbesQuerier interface {
	CreateProbe(ctx context.Context, probeID uuid.UUID, nodeID uuid.UUID, probeType string, target string, port int, intervalSeconds int, count int, timeoutSeconds int) error
	GetProbes(ctx context.Context) ([]*models.Probe, error)
	GetProbesByNode(ctx context.Context, nodeID uuid.UUID) ([]*models.Probe, error)
	GetProbeByID(ctx context.Context, probeID uuid.UUID) (*models.Probe, error)
	UpdateProbe(ctx context.Context, probeID uuid.UUID, updates map[string]interface{}) error
	DeleteProbe(ctx context.Context, probeID uuid.UUID) error
}

// CreateProbe inserts a new probe into database
func CreateProbe(ctx context.Context, pool *pgxpool.Pool, probeID uuid.UUID, nodeID uuid.UUID, probeType string, target string, port int, intervalSeconds int, count int, timeoutSeconds int) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `
		INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`

	_, err = conn.Exec(ctx, query, probeID, nodeID, probeType, target, port, intervalSeconds, count, timeoutSeconds)
	return err
}

// GetProbes retrieves all probes from database
func GetProbes(ctx context.Context, pool *pgxpool.Pool) ([]*models.Probe, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at
		FROM probes
		ORDER BY created_at DESC
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var probes []*models.Probe
	for rows.Next() {
		var probe models.Probe
		err := rows.Scan(&probe.ID, &probe.NodeID, &probe.Type, &probe.Target, &probe.Port, &probe.IntervalSeconds, &probe.Count, &probe.TimeoutSeconds, &probe.CreatedAt, &probe.UpdatedAt)
		if err != nil {
			return nil, err
		}
		probes = append(probes, &probe)
	}

	return probes, rows.Err()
}

// GetProbesByNode retrieves probes filtered by node
func GetProbesByNode(ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID) ([]*models.Probe, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at
		FROM probes
		WHERE node_id = $1
		ORDER BY created_at DESC
	`

	rows, err := conn.Query(ctx, query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var probes []*models.Probe
	for rows.Next() {
		var probe models.Probe
		err := rows.Scan(&probe.ID, &probe.NodeID, &probe.Type, &probe.Target, &probe.Port, &probe.IntervalSeconds, &probe.Count, &probe.TimeoutSeconds, &probe.CreatedAt, &probe.UpdatedAt)
		if err != nil {
			return nil, err
		}
		probes = append(probes, &probe)
	}

	return probes, rows.Err()
}

// GetProbeByID retrieves a specific probe by ID
func GetProbeByID(ctx context.Context, pool *pgxpool.Pool, probeID uuid.UUID) (*models.Probe, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, node_id, type, target, port, interval_seconds, count, timeout_seconds, created_at, updated_at
		FROM probes
		WHERE id = $1
	`

	var probe models.Probe
	err = conn.QueryRow(ctx, query, probeID).Scan(
		&probe.ID, &probe.NodeID, &probe.Type, &probe.Target, &probe.Port,
		&probe.IntervalSeconds, &probe.Count, &probe.TimeoutSeconds,
		&probe.CreatedAt, &probe.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProbeNotFound
		}
		return nil, err
	}

	return &probe, nil
}

// UpdateProbe updates an existing probe
func UpdateProbe(ctx context.Context, pool *pgxpool.Pool, probeID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Build dynamic update query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 2 // First arg is $1 (probeID)

	if probeType, ok := updates["type"]; ok && probeType != nil {
		setParts = append(setParts, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, probeType)
		argIndex++
	}

	if target, ok := updates["target"]; ok && target != nil {
		setParts = append(setParts, fmt.Sprintf("target = $%d", argIndex))
		args = append(args, target)
		argIndex++
	}

	if port, ok := updates["port"]; ok && port != nil {
		setParts = append(setParts, fmt.Sprintf("port = $%d", argIndex))
		args = append(args, port)
		argIndex++
	}

	if intervalSeconds, ok := updates["interval_seconds"]; ok && intervalSeconds != nil {
		setParts = append(setParts, fmt.Sprintf("interval_seconds = $%d", argIndex))
		args = append(args, intervalSeconds)
		argIndex++
	}

	if count, ok := updates["count"]; ok && count != nil {
		setParts = append(setParts, fmt.Sprintf("count = $%d", argIndex))
		args = append(args, count)
		argIndex++
	}

	if timeoutSeconds, ok := updates["timeout_seconds"]; ok && timeoutSeconds != nil {
		setParts = append(setParts, fmt.Sprintf("timeout_seconds = $%d", argIndex))
		args = append(args, timeoutSeconds)
		argIndex++
	}

	// Always update updated_at
	setParts = append(setParts, "updated_at = NOW()")

	query := "UPDATE probes SET " + setParts[0]
	for _, part := range setParts[1:] {
		query += ", " + part
	}
	query += " WHERE id = $1"

	// Prepend probeID to args
	allArgs := make([]interface{}, len(args)+1)
	allArgs[0] = probeID
	copy(allArgs[1:], args)

	result, err := tx.Exec(ctx, query, allArgs...)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrProbeNotFound
	}

	return tx.Commit(ctx)
}

// DeleteProbe removes a probe from database
func DeleteProbe(ctx context.Context, pool *pgxpool.Pool, probeID uuid.UUID) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := "DELETE FROM probes WHERE id = $1"

	result, err := conn.Exec(ctx, query, probeID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrProbeNotFound
	}

	return nil
}

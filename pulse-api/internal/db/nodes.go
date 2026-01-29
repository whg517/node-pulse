package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

var (
	ErrNodeNotFound = errors.New("node not found")

	// Heartbeat timeout threshold for offline detection
	HeartbeatTimeout = 5 * time.Minute
)

// CalculateNodeStatus determines node status based on last heartbeat time
func CalculateNodeStatus(lastHeartbeat *time.Time) string {
	if lastHeartbeat == nil {
		return "connecting"
	}

	timeSinceHeartbeat := time.Since(*lastHeartbeat)

	// Online if heartbeat within last 5 minutes (≤ 5 min)
	// Offline if heartbeat older than 5 minutes (> 5 min)
	if timeSinceHeartbeat > HeartbeatTimeout {
		return "offline"
	}

	return "online"
}

// NodesQuerier defines interface for node database operations
type NodesQuerier interface {
	CreateNode(ctx context.Context, nodeID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error
	GetNodes(ctx context.Context) ([]*models.Node, error)
	GetNodesByRegion(ctx context.Context, region string) ([]*models.Node, error)
	GetNodeByID(ctx context.Context, nodeID uuid.UUID) (*models.Node, error)
	GetNodeByNameAndIP(ctx context.Context, name string, ip string) (*models.Node, error)
	GetNodeStatus(ctx context.Context, nodeID uuid.UUID) (*models.NodeStatus, error)
	UpdateNode(ctx context.Context, nodeID uuid.UUID, updates map[string]interface{}) error
	DeleteNode(ctx context.Context, nodeID uuid.UUID) error
}

// CreateNode inserts a new node into database
func CreateNode(ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `
		INSERT INTO nodes (id, name, ip, region, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`

	var tagsJSON string
	if tags != nil {
		tagsBytes, err := json.Marshal(tags)
		if err != nil {
			return err
		}
		tagsJSON = string(tagsBytes)
	} else {
		tagsJSON = "{}"
	}

	_, err = conn.Exec(ctx, query, nodeID, name, ip, region, tagsJSON)
	return err
}

// GetNodes retrieves all nodes from database
func GetNodes(ctx context.Context, pool *pgxpool.Pool) ([]*models.Node, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, name, ip, region, tags::text, created_at, updated_at
		FROM nodes
		ORDER BY created_at DESC
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		err := rows.Scan(&node.ID, &node.Name, &node.IP, &node.Region, &node.Tags, &node.CreatedAt, &node.UpdatedAt)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

// GetNodesByRegion retrieves nodes filtered by region
func GetNodesByRegion(ctx context.Context, pool *pgxpool.Pool, region string) ([]*models.Node, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, name, ip, region, tags::text, created_at, updated_at
		FROM nodes
		WHERE region = $1
		ORDER BY created_at DESC
	`

	rows, err := conn.Query(ctx, query, region)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		err := rows.Scan(&node.ID, &node.Name, &node.IP, &node.Region, &node.Tags, &node.CreatedAt, &node.UpdatedAt)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

// GetNodeByNameAndIP retrieves a node by name and IP combination (for duplicate detection)
func GetNodeByNameAndIP(ctx context.Context, pool *pgxpool.Pool, name string, ip string) (*models.Node, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, name, ip, region, tags::text, created_at, updated_at
		FROM nodes
		WHERE name = $1 AND ip = $2
		LIMIT 1
	`

	var node models.Node
	err = conn.QueryRow(ctx, query, name, ip).Scan(
		&node.ID, &node.Name, &node.IP, &node.Region, &node.Tags, &node.CreatedAt, &node.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No duplicate found
		}
		return nil, err
	}

	return &node, nil
}

// GetNodeByID retrieves a specific node by ID
func GetNodeByID(ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID) (*models.Node, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, name, ip, region, tags::text, created_at, updated_at
		FROM nodes
		WHERE id = $1
	`

	var node models.Node
	err = conn.QueryRow(ctx, query, nodeID).Scan(&node.ID, &node.Name, &node.IP, &node.Region, &node.Tags, &node.CreatedAt, &node.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}

	return &node, nil
}

// GetNodeStatus retrieves node status information
func GetNodeStatus(ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID) (*models.NodeStatus, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `
		SELECT id, name, status, last_heartbeat, last_report_time
		FROM nodes
		WHERE id = $1
	`

	var status models.NodeStatus
	var lastHeartbeat *time.Time
	var lastReportTime *time.Time

	err = conn.QueryRow(ctx, query, nodeID).Scan(
		&status.ID,
		&status.Name,
		&status.Status,
		&lastHeartbeat,
		&lastReportTime,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}

	// Calculate status if not set in database
	if status.Status == "" || status.Status == "connecting" {
		status.Status = CalculateNodeStatus(lastHeartbeat)
	}

	status.LastHeartbeat = lastHeartbeat
	status.LastReportTime = lastReportTime

	return &status, nil
}

// UpdateNode updates an existing node
func UpdateNode(ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID, updates map[string]interface{}) error {
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
	argIndex := 2 // First arg is $1 (nodeID)

	if name, ok := updates["name"]; ok && name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, name)
		argIndex++
	}

	if ip, ok := updates["ip"]; ok && ip != nil {
		setParts = append(setParts, fmt.Sprintf("ip = $%d", argIndex))
		args = append(args, ip)
		argIndex++
	}

	if region, ok := updates["region"]; ok && region != nil {
		setParts = append(setParts, fmt.Sprintf("region = $%d", argIndex))
		args = append(args, region)
		argIndex++
	}

	if tags, ok := updates["tags"]; ok && tags != nil {
		tagsBytes, err := json.Marshal(tags)
		if err != nil {
			return err
		}
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, string(tagsBytes))
		argIndex++
	}

	if len(setParts) == 0 {
		tx.Rollback(ctx) // Nothing to update
		return nil
	}

	// Always update updated_at
	setParts = append(setParts, "updated_at = NOW()")

	query := "UPDATE nodes SET " + setParts[0]
	for _, part := range setParts[1:] {
		query += ", " + part
	}
	query += " WHERE id = $1"

	// Prepend nodeID to args
	allArgs := make([]interface{}, len(args)+1)
	allArgs[0] = nodeID
	copy(allArgs[1:], args)

	result, err := tx.Exec(ctx, query, allArgs...)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNodeNotFound
	}

	return tx.Commit(ctx)
}

// DeleteNode removes a node from database
func DeleteNode(ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := "DELETE FROM nodes WHERE id = $1"

	result, err := conn.Exec(ctx, query, nodeID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNodeNotFound
	}

	return nil
}

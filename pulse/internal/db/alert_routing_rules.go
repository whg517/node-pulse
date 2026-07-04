package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// AlertRoutingRulesRepository stores per-webhook routing rules (ADR-002).
type AlertRoutingRulesRepository interface {
	Create(ctx context.Context, r *models.AlertRoutingRule) error
	ListByOwner(ctx context.Context, ownerUserID string) ([]*models.AlertRoutingRule, error)
	Update(ctx context.Context, r *models.AlertRoutingRule) error
	Delete(ctx context.Context, id, ownerUserID string) error
	// ListEnabled returns all enabled rules (used by the dispatch path matcher).
	ListEnabled(ctx context.Context) ([]*models.AlertRoutingRule, error)
}

type alertRoutingRulesRepo struct {
	pool *pgxpool.Pool
}

// NewAlertRoutingRulesRepository creates a repository backed by PostgreSQL.
func NewAlertRoutingRulesRepository(pool *pgxpool.Pool) AlertRoutingRulesRepository {
	return &alertRoutingRulesRepo{pool: pool}
}

func (r *alertRoutingRulesRepo) Create(ctx context.Context, rule *models.AlertRoutingRule) error {
	sev, err := json.Marshal(rule.Severities)
	if err != nil {
		return errors.New("marshal severities: " + err.Error())
	}
	return r.pool.QueryRow(ctx, `
		INSERT INTO alert_routing_rules (owner_user_id, name, enabled, metric, severities, node_id, webhook_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, rule.OwnerUserID, rule.Name, rule.Enabled, nullable(rule.Metric), sev, nullable(rule.NodeID), rule.WebhookID).
		Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (r *alertRoutingRulesRepo) ListByOwner(ctx context.Context, ownerUserID string) ([]*models.AlertRoutingRule, error) {
	return r.query(ctx, `WHERE owner_user_id = $1 ORDER BY created_at DESC`, ownerUserID)
}

func (r *alertRoutingRulesRepo) ListEnabled(ctx context.Context) ([]*models.AlertRoutingRule, error) {
	return r.query(ctx, `WHERE enabled = true`)
}

func (r *alertRoutingRulesRepo) Update(ctx context.Context, rule *models.AlertRoutingRule) error {
	sev, err := json.Marshal(rule.Severities)
	if err != nil {
		return errors.New("marshal severities: " + err.Error())
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE alert_routing_rules SET
			name = $3, enabled = $4, metric = $5, severities = $6, node_id = $7, webhook_id = $8, updated_at = NOW()
		WHERE id = $1 AND owner_user_id = $2
	`, rule.ID, rule.OwnerUserID, rule.Name, rule.Enabled, nullable(rule.Metric), sev, nullable(rule.NodeID), rule.WebhookID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *alertRoutingRulesRepo) Delete(ctx context.Context, id, ownerUserID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM alert_routing_rules WHERE id = $1 AND owner_user_id = $2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *alertRoutingRulesRepo) query(ctx context.Context, whereClause string, args ...any) ([]*models.AlertRoutingRule, error) {
	q := `SELECT id, owner_user_id, name, enabled, COALESCE(metric,''), severities, COALESCE(node_id,''), webhook_id, created_at, updated_at
	      FROM alert_routing_rules ` + whereClause
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.AlertRoutingRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

func scanRule(row rowScanner) (*models.AlertRoutingRule, error) {
	var (
		rule    models.AlertRoutingRule
		sevJSON []byte
	)
	if err := row.Scan(&rule.ID, &rule.OwnerUserID, &rule.Name, &rule.Enabled, &rule.Metric, &sevJSON, &rule.NodeID, &rule.WebhookID, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return nil, err
	}
	if len(sevJSON) > 0 && string(sevJSON) != "null" {
		if err := json.Unmarshal(sevJSON, &rule.Severities); err != nil {
			return nil, errors.New("unmarshal severities: " + err.Error())
		}
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	return &rule, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

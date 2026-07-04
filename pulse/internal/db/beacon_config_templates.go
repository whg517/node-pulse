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

// BeaconConfigTemplatesRepository durably stores reusable beacon config templates.
type BeaconConfigTemplatesRepository interface {
	Create(ctx context.Context, t *models.BeaconConfigTemplate) error
	GetByID(ctx context.Context, id string) (*models.BeaconConfigTemplate, error)
	ListByOwner(ctx context.Context, ownerUserID string) ([]*models.BeaconConfigTemplate, error)
	Update(ctx context.Context, t *models.BeaconConfigTemplate) error
	Delete(ctx context.Context, id, ownerUserID string) error
}

type beaconConfigTemplatesRepo struct {
	pool *pgxpool.Pool
}

// NewBeaconConfigTemplatesRepository creates a new repository backed by PostgreSQL.
func NewBeaconConfigTemplatesRepository(pool *pgxpool.Pool) BeaconConfigTemplatesRepository {
	return &beaconConfigTemplatesRepo{pool: pool}
}

func (r *beaconConfigTemplatesRepo) Create(ctx context.Context, t *models.BeaconConfigTemplate) error {
	probes, err := json.Marshal(t.Probes)
	if err != nil {
		return errors.New("marshal probes: " + err.Error())
	}
	return r.pool.QueryRow(ctx, `
		INSERT INTO beacon_config_templates (owner_user_id, name, description, probes, interval_seconds, timeout_seconds)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, t.OwnerUserID, t.Name, t.Description, probes, t.IntervalSeconds, t.TimeoutSeconds).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *beaconConfigTemplatesRepo) GetByID(ctx context.Context, id string) (*models.BeaconConfigTemplate, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_user_id, name, description, probes, interval_seconds, timeout_seconds, created_at, updated_at
		FROM beacon_config_templates WHERE id = $1
	`, id)
	return scanTemplate(row)
}

func (r *beaconConfigTemplatesRepo) ListByOwner(ctx context.Context, ownerUserID string) ([]*models.BeaconConfigTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_user_id, name, description, probes, interval_seconds, timeout_seconds, created_at, updated_at
		FROM beacon_config_templates
		WHERE owner_user_id = $1
		ORDER BY created_at DESC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.BeaconConfigTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *beaconConfigTemplatesRepo) Update(ctx context.Context, t *models.BeaconConfigTemplate) error {
	probes, err := json.Marshal(t.Probes)
	if err != nil {
		return errors.New("marshal probes: " + err.Error())
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE beacon_config_templates SET
			name = $3, description = $4, probes = $5, interval_seconds = $6, timeout_seconds = $7, updated_at = NOW()
		WHERE id = $1 AND owner_user_id = $2
	`, t.ID, t.OwnerUserID, t.Name, t.Description, probes, t.IntervalSeconds, t.TimeoutSeconds)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *beaconConfigTemplatesRepo) Delete(ctx context.Context, id, ownerUserID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM beacon_config_templates WHERE id = $1 AND owner_user_id = $2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanTemplate(row rowScanner) (*models.BeaconConfigTemplate, error) {
	var (
		t          models.BeaconConfigTemplate
		probesJSON []byte
		desc       *string
	)
	if err := row.Scan(&t.ID, &t.OwnerUserID, &t.Name, &desc, &probesJSON, &t.IntervalSeconds, &t.TimeoutSeconds, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	if desc != nil {
		t.Description = *desc
	}
	if err := json.Unmarshal(probesJSON, &t.Probes); err != nil {
		return nil, errors.New("unmarshal probes: " + err.Error())
	}
	// Normalize zero time to a sensible value for JSON clients.
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	return &t, nil
}

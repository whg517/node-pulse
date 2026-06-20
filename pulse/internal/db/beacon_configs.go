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
)

var ErrBeaconConfigNotFound = errors.New("beacon config not found")

// BeaconProbeConfig is a persisted probe entry assigned to a beacon.
type BeaconProbeConfig struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Target          string `json:"target"`
	Port            int    `json:"port"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	Count           int    `json:"count"`
	MaxHops         int    `json:"max_hops,omitempty"`
	PacketSize      int    `json:"packet_size,omitempty"`
}

// BeaconConfig is the current server-managed config for a beacon.
type BeaconConfig struct {
	BeaconID        uuid.UUID           `json:"beacon_id"`
	Probes          []BeaconProbeConfig `json:"probes"`
	IntervalSeconds int                 `json:"interval_seconds"`
	TimeoutSeconds  int                 `json:"timeout_seconds"`
	Version         int                 `json:"version"`
	LastAckVersion  *int                `json:"last_ack_version,omitempty"`
	LastAckAt       *time.Time          `json:"last_ack_at,omitempty"`
	LastAckStatus   string              `json:"last_ack_status,omitempty"`
	LastAckError    string              `json:"last_ack_error,omitempty"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// BeaconConfigHistoryEntry is a previous config snapshot.
type BeaconConfigHistoryEntry struct {
	Version   int          `json:"version"`
	Config    BeaconConfig `json:"config"`
	ChangedAt time.Time    `json:"changed_at"`
	ChangedBy string       `json:"changed_by"`
}

// BeaconConfigUpdate contains optional config fields to update.
type BeaconConfigUpdate struct {
	Probes          *[]BeaconProbeConfig
	IntervalSeconds *int
	TimeoutSeconds  *int
	ChangedBy       string
}

// BeaconConfigsQuerier defines database operations for server-managed beacon config.
type BeaconConfigsQuerier interface {
	GetBeaconConfig(ctx context.Context, beaconID uuid.UUID) (*BeaconConfig, error)
	UpsertBeaconConfig(ctx context.Context, beaconID uuid.UUID, update BeaconConfigUpdate) (*BeaconConfig, error)
	GetBeaconConfigHistory(ctx context.Context, beaconID uuid.UUID, limit int) ([]BeaconConfigHistoryEntry, error)
	AcknowledgeBeaconConfig(ctx context.Context, beaconID uuid.UUID, version int, status string, errorMessage string) error
}

// GetBeaconConfig returns the persisted beacon config.
func GetBeaconConfig(ctx context.Context, pool *pgxpool.Pool, beaconID uuid.UUID) (*BeaconConfig, error) {
	var cfg BeaconConfig
	var probesJSON []byte

	err := pool.QueryRow(ctx, `
		SELECT beacon_id, probes, interval_seconds, timeout_seconds, version,
			last_ack_version, last_ack_at, COALESCE(last_ack_status, ''), COALESCE(last_ack_error, ''), updated_at
		FROM beacon_configs
		WHERE beacon_id = $1
	`, beaconID).Scan(
		&cfg.BeaconID, &probesJSON, &cfg.IntervalSeconds, &cfg.TimeoutSeconds, &cfg.Version,
		&cfg.LastAckVersion, &cfg.LastAckAt, &cfg.LastAckStatus, &cfg.LastAckError, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBeaconConfigNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(probesJSON, &cfg.Probes); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// UpsertBeaconConfig creates or updates the persisted beacon config and stores the old snapshot.
func UpsertBeaconConfig(ctx context.Context, pool *pgxpool.Pool, beaconID uuid.UUID, update BeaconConfigUpdate) (*BeaconConfig, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	changedBy := update.ChangedBy
	if changedBy == "" {
		changedBy = "system"
	}

	current, err := scanBeaconConfig(ctx, tx, beaconID)
	if err != nil && !errors.Is(err, ErrBeaconConfigNotFound) {
		return nil, err
	}

	if current != nil {
		snapshot, err := json.Marshal(current)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO beacon_config_history (beacon_id, version, config, changed_at, changed_by)
			VALUES ($1, $2, $3, NOW(), $4)
		`, beaconID, current.Version, snapshot, changedBy)
		if err != nil {
			return nil, err
		}
	} else {
		current = &BeaconConfig{
			BeaconID:        beaconID,
			Probes:          []BeaconProbeConfig{},
			IntervalSeconds: 60,
			TimeoutSeconds:  5,
			Version:         0,
		}
	}

	if update.Probes != nil {
		current.Probes = *update.Probes
	}
	if update.IntervalSeconds != nil {
		current.IntervalSeconds = *update.IntervalSeconds
	}
	if update.TimeoutSeconds != nil {
		current.TimeoutSeconds = *update.TimeoutSeconds
	}
	current.Version++

	probesJSON, err := json.Marshal(current.Probes)
	if err != nil {
		return nil, err
	}

	var saved BeaconConfig
	var savedProbesJSON []byte
	err = tx.QueryRow(ctx, `
		INSERT INTO beacon_configs (beacon_id, probes, interval_seconds, timeout_seconds, version, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (beacon_id) DO UPDATE SET
			probes = EXCLUDED.probes,
			interval_seconds = EXCLUDED.interval_seconds,
			timeout_seconds = EXCLUDED.timeout_seconds,
			version = EXCLUDED.version,
			updated_at = NOW()
		RETURNING beacon_id, probes, interval_seconds, timeout_seconds, version,
			last_ack_version, last_ack_at, COALESCE(last_ack_status, ''), COALESCE(last_ack_error, ''), updated_at
	`, beaconID, probesJSON, current.IntervalSeconds, current.TimeoutSeconds, current.Version).
		Scan(
			&saved.BeaconID, &savedProbesJSON, &saved.IntervalSeconds, &saved.TimeoutSeconds, &saved.Version,
			&saved.LastAckVersion, &saved.LastAckAt, &saved.LastAckStatus, &saved.LastAckError, &saved.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(savedProbesJSON, &saved.Probes); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &saved, nil
}

// GetBeaconConfigHistory returns recent config snapshots for a beacon.
func GetBeaconConfigHistory(ctx context.Context, pool *pgxpool.Pool, beaconID uuid.UUID, limit int) ([]BeaconConfigHistoryEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := pool.Query(ctx, `
		SELECT version, config, changed_at, changed_by
		FROM beacon_config_history
		WHERE beacon_id = $1
		ORDER BY changed_at DESC
		LIMIT $2
	`, beaconID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]BeaconConfigHistoryEntry, 0)
	for rows.Next() {
		var entry BeaconConfigHistoryEntry
		var configJSON []byte
		if err := rows.Scan(&entry.Version, &configJSON, &entry.ChangedAt, &entry.ChangedBy); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &entry.Config); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}

	return history, rows.Err()
}

// AcknowledgeBeaconConfig stores a beacon's config apply status.
func AcknowledgeBeaconConfig(ctx context.Context, pool *pgxpool.Pool, beaconID uuid.UUID, version int, status string, errorMessage string) error {
	if version < 1 {
		return fmt.Errorf("version must be >= 1")
	}
	if status != "applied" && status != "failed" {
		return fmt.Errorf("status must be applied or failed")
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO beacon_configs (
			beacon_id, probes, interval_seconds, timeout_seconds, version,
			last_ack_version, last_ack_at, last_ack_status, last_ack_error, updated_at
		)
		VALUES ($1, '[]', 60, 5, GREATEST($2, 1), $2, NOW(), $3, NULLIF($4, ''), NOW())
		ON CONFLICT (beacon_id) DO UPDATE SET
			last_ack_version = EXCLUDED.last_ack_version,
			last_ack_at = EXCLUDED.last_ack_at,
			last_ack_status = EXCLUDED.last_ack_status,
			last_ack_error = EXCLUDED.last_ack_error
	`, beaconID, version, status, errorMessage)
	return err
}

type beaconConfigScanner interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func scanBeaconConfig(ctx context.Context, q beaconConfigScanner, beaconID uuid.UUID) (*BeaconConfig, error) {
	var cfg BeaconConfig
	var probesJSON []byte

	err := q.QueryRow(ctx, `
		SELECT beacon_id, probes, interval_seconds, timeout_seconds, version,
			last_ack_version, last_ack_at, COALESCE(last_ack_status, ''), COALESCE(last_ack_error, ''), updated_at
		FROM beacon_configs
		WHERE beacon_id = $1
		FOR UPDATE
	`, beaconID).Scan(
		&cfg.BeaconID, &probesJSON, &cfg.IntervalSeconds, &cfg.TimeoutSeconds, &cfg.Version,
		&cfg.LastAckVersion, &cfg.LastAckAt, &cfg.LastAckStatus, &cfg.LastAckError, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBeaconConfigNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(probesJSON, &cfg.Probes); err != nil {
		return nil, err
	}
	return &cfg, nil
}

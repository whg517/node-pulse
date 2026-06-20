package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMTRResultNotFound = errors.New("mtr result not found")

// MTRHop is a single hop in a route trace.
type MTRHop struct {
	HopNumber  int     `json:"hop_number"`
	IP         string  `json:"ip"`
	Hostname   string  `json:"hostname,omitempty"`
	ASNumber   string  `json:"as_number,omitempty"`
	Sent       int     `json:"sent"`
	Received   int     `json:"received"`
	LossRate   float64 `json:"loss_rate"`
	LastRTTMs  float64 `json:"last_rtt_ms"`
	AvgRTTMs   float64 `json:"avg_rtt_ms"`
	BestRTTMs  float64 `json:"best_rtt_ms"`
	WorstRTTMs float64 `json:"worst_rtt_ms"`
	StdDevMs   float64 `json:"std_dev_ms"`
	Location   string  `json:"location,omitempty"`
}

// MTRResult is a persisted route trace snapshot.
type MTRResult struct {
	ID           uuid.UUID `json:"id"`
	NodeID       uuid.UUID `json:"node_id"`
	ProbeID      string    `json:"probe_id,omitempty"`
	Target       string    `json:"target"`
	Success      bool      `json:"success"`
	TotalHops    int       `json:"total_hops"`
	Hops         []MTRHop  `json:"hops"`
	CompletedAt  time.Time `json:"completed_at"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// MTRResultInput is the payload needed to persist an MTR result.
type MTRResultInput struct {
	NodeID       uuid.UUID
	ProbeID      string
	Target       string
	Success      bool
	TotalHops    int
	Hops         []MTRHop
	CompletedAt  time.Time
	ErrorMessage string
}

// MTRResultQuery contains filters for listing route trace snapshots.
type MTRResultQuery struct {
	NodeID    uuid.UUID
	Target    string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
}

// MTRResultsQuerier defines MTR result persistence operations.
type MTRResultsQuerier interface {
	SaveMTRResult(ctx context.Context, input MTRResultInput) (*MTRResult, error)
	GetLatestMTRResult(ctx context.Context, nodeID uuid.UUID) (*MTRResult, error)
	GetMTRResults(ctx context.Context, query MTRResultQuery) ([]MTRResult, error)
}

// SaveMTRResult persists an MTR result.
func SaveMTRResult(ctx context.Context, pool *pgxpool.Pool, input MTRResultInput) (*MTRResult, error) {
	hopsJSON, err := json.Marshal(input.Hops)
	if err != nil {
		return nil, err
	}
	if input.TotalHops == 0 {
		input.TotalHops = len(input.Hops)
	}
	if input.CompletedAt.IsZero() {
		input.CompletedAt = time.Now()
	}

	var result MTRResult
	var storedHops []byte
	err = pool.QueryRow(ctx, `
		INSERT INTO mtr_results (
			node_id, probe_id, target, success, total_hops, hops, completed_at, error_message
		)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, NULLIF($8, ''))
		RETURNING id, node_id, COALESCE(probe_id, ''), target, success, total_hops, hops,
			completed_at, COALESCE(error_message, ''), created_at
	`, input.NodeID, input.ProbeID, input.Target, input.Success, input.TotalHops, hopsJSON, input.CompletedAt, input.ErrorMessage).
		Scan(
			&result.ID, &result.NodeID, &result.ProbeID, &result.Target, &result.Success,
			&result.TotalHops, &storedHops, &result.CompletedAt, &result.ErrorMessage, &result.CreatedAt,
		)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(storedHops, &result.Hops); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLatestMTRResult returns the latest MTR result for a node.
func GetLatestMTRResult(ctx context.Context, pool *pgxpool.Pool, nodeID uuid.UUID) (*MTRResult, error) {
	var result MTRResult
	var hopsJSON []byte
	err := pool.QueryRow(ctx, `
		SELECT id, node_id, COALESCE(probe_id, ''), target, success, total_hops, hops,
			completed_at, COALESCE(error_message, ''), created_at
		FROM mtr_results
		WHERE node_id = $1
		ORDER BY completed_at DESC
		LIMIT 1
	`, nodeID).Scan(
		&result.ID, &result.NodeID, &result.ProbeID, &result.Target, &result.Success,
		&result.TotalHops, &hopsJSON, &result.CompletedAt, &result.ErrorMessage, &result.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMTRResultNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(hopsJSON, &result.Hops); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMTRResults returns MTR results for a node ordered from newest to oldest.
func GetMTRResults(ctx context.Context, pool *pgxpool.Pool, query MTRResultQuery) ([]MTRResult, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var startArg any
	if query.StartTime != nil {
		startArg = *query.StartTime
	}
	var endArg any
	if query.EndTime != nil {
		endArg = *query.EndTime
	}

	rows, err := pool.Query(ctx, `
		SELECT id, node_id, COALESCE(probe_id, ''), target, success, total_hops, hops,
			completed_at, COALESCE(error_message, ''), created_at
		FROM mtr_results
		WHERE node_id = $1
		  AND ($2 = '' OR target = $2)
		  AND ($3::timestamptz IS NULL OR completed_at >= $3)
		  AND ($4::timestamptz IS NULL OR completed_at <= $4)
		ORDER BY completed_at DESC
		LIMIT $5
	`, query.NodeID, query.Target, startArg, endArg, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]MTRResult, 0)
	for rows.Next() {
		var result MTRResult
		var hopsJSON []byte
		if err := rows.Scan(
			&result.ID, &result.NodeID, &result.ProbeID, &result.Target, &result.Success,
			&result.TotalHops, &hopsJSON, &result.CompletedAt, &result.ErrorMessage, &result.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(hopsJSON, &result.Hops); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

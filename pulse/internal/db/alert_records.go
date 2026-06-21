package db

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// Custom errors for alert record operations
var (
	ErrAlertRecordNotFound     = errors.New("alert record not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrAlertNoteEmpty          = errors.New("alert note content is empty")
)

// CreateAlertRecord creates a new alert record in the database
func CreateAlertRecord(ctx context.Context, pool *pgxpool.Pool, record *models.AlertRecord) error {
	record.ID = uuid.New().String()
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()

	// Set default status if not provided
	if record.Status == "" {
		record.Status = "pending"
	}

	query := `
		INSERT INTO alert_records (id, alert_event_id, node_id, metric, level, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := pool.QueryRow(ctx, query,
		record.ID,
		record.AlertEventID,
		record.NodeID,
		record.Metric,
		record.Level,
		record.Status,
		record.CreatedAt,
		record.UpdatedAt,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create alert record: %w", err)
	}

	return nil
}

// GetAlertRecords retrieves alert records with optional filtering
func GetAlertRecords(ctx context.Context, pool *pgxpool.Pool, filters AlertRecordFilters) ([]models.AlertRecord, error) {
	// Build query with dynamic filters
	query := `
		SELECT id, alert_event_id, node_id, metric, level, status, created_at, updated_at
		FROM alert_records
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filters.NodeID != nil {
		query += fmt.Sprintf(" AND node_id = $%d", argCount)
		args = append(args, *filters.NodeID)
		argCount++
	}

	if filters.Level != nil {
		query += fmt.Sprintf(" AND level = $%d", argCount)
		args = append(args, *filters.Level)
		argCount++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filters.Status)
		argCount++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.StartTime)
		argCount++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.EndTime)
		argCount++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert records: %w", err)
	}
	defer rows.Close()

	var records []models.AlertRecord
	for rows.Next() {
		var record models.AlertRecord
		err := rows.Scan(
			&record.ID,
			&record.AlertEventID,
			&record.NodeID,
			&record.Metric,
			&record.Level,
			&record.Status,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert records: %w", err)
	}

	return records, nil
}

// GetAlertRecordByID retrieves a single alert record by ID
func GetAlertRecordByID(ctx context.Context, pool *pgxpool.Pool, id string) (*models.AlertRecord, error) {
	query := `
		SELECT id, alert_event_id, node_id, metric, level, status, created_at, updated_at
		FROM alert_records
		WHERE id = $1
	`

	var record models.AlertRecord
	err := pool.QueryRow(ctx, query, id).Scan(
		&record.ID,
		&record.AlertEventID,
		&record.NodeID,
		&record.Metric,
		&record.Level,
		&record.Status,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ErrAlertRecordNotFound, id)
		}
		return nil, fmt.Errorf("failed to get alert record: %w", err)
	}

	return &record, nil
}

// UpdateAlertRecordStatus updates the status of an alert record
func UpdateAlertRecordStatus(ctx context.Context, pool *pgxpool.Pool, id string, newStatus string) error {
	// First, get current record to validate transition
	record, err := GetAlertRecordByID(ctx, pool, id)
	if err != nil {
		return err // GetAlertRecordByID already wraps ErrAlertRecordNotFound
	}

	// Validate status transition
	if !record.CanTransitionTo(newStatus) {
		return fmt.Errorf("%w: %s to %s", ErrInvalidStatusTransition, record.Status, newStatus)
	}

	// Update the status
	query := `
		UPDATE alert_records
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := pool.Exec(ctx, query, newStatus, id)
	if err != nil {
		return fmt.Errorf("failed to update alert record status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrAlertRecordNotFound, id)
	}

	return nil
}

// CreateAlertStatusHistory stores a status transition for an alert record.
func CreateAlertStatusHistory(ctx context.Context, pool *pgxpool.Pool, alertID string, fromStatus string, toStatus string, userID *string) (*models.AlertStatusHistory, error) {
	if _, err := GetAlertRecordByID(ctx, pool, alertID); err != nil {
		return nil, err
	}

	userIDValue, userName := resolveAlertActor(ctx, pool, userID)
	history := &models.AlertStatusHistory{
		AlertID:    alertID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		UserName:   userName,
	}

	query := `
		INSERT INTO alert_status_history (alert_id, from_status, to_status, user_id, user_name)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5)
		RETURNING id, alert_id, COALESCE(from_status, ''), to_status, COALESCE(user_id::text, ''), user_name, created_at
	`

	err := pool.QueryRow(ctx, query, alertID, fromStatus, toStatus, userIDValue, userName).Scan(
		&history.ID,
		&history.AlertID,
		&history.FromStatus,
		&history.ToStatus,
		&history.UserID,
		&history.UserName,
		&history.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert status history: %w", err)
	}

	return history, nil
}

// CreateAlertNote creates an operator note for an alert record.
func CreateAlertNote(ctx context.Context, pool *pgxpool.Pool, alertID string, userID *string, content string) (*models.AlertNote, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrAlertNoteEmpty
	}

	if _, err := GetAlertRecordByID(ctx, pool, alertID); err != nil {
		return nil, err
	}

	userIDValue, userName := resolveAlertActor(ctx, pool, userID)

	note := &models.AlertNote{
		AlertID:  alertID,
		UserName: userName,
		Content:  content,
	}

	query := `
		INSERT INTO alert_notes (alert_id, user_id, user_name, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, alert_id, COALESCE(user_id::text, ''), user_name, content, created_at
	`

	err := pool.QueryRow(ctx, query, alertID, userIDValue, userName, content).Scan(
		&note.ID,
		&note.AlertID,
		&note.UserID,
		&note.UserName,
		&note.Content,
		&note.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert note: %w", err)
	}

	return note, nil
}

// GetAlertTimeline returns a merged lifecycle timeline ordered oldest-first.
func GetAlertTimeline(ctx context.Context, pool *pgxpool.Pool, alertID string) ([]models.AlertTimelineItem, error) {
	record, err := GetAlertRecordByID(ctx, pool, alertID)
	if err != nil {
		return nil, err
	}

	items := []models.AlertTimelineItem{
		{
			ID:        "created-" + record.ID,
			Type:      "created",
			Title:     "Alert created",
			Status:    record.Status,
			CreatedAt: record.CreatedAt,
		},
	}

	historyRows, err := pool.Query(ctx, `
		SELECT id, alert_id, COALESCE(from_status, ''), to_status, COALESCE(user_id::text, ''), user_name, created_at
		FROM alert_status_history
		WHERE alert_id = $1
		ORDER BY created_at ASC, id ASC
	`, alertID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert status history: %w", err)
	}
	defer historyRows.Close()

	for historyRows.Next() {
		var history models.AlertStatusHistory
		if err := historyRows.Scan(
			&history.ID,
			&history.AlertID,
			&history.FromStatus,
			&history.ToStatus,
			&history.UserID,
			&history.UserName,
			&history.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert status history: %w", err)
		}

		items = append(items, models.AlertTimelineItem{
			ID:         history.ID,
			Type:       "status_changed",
			Title:      "Status changed",
			Status:     history.ToStatus,
			FromStatus: history.FromStatus,
			ToStatus:   history.ToStatus,
			UserID:     history.UserID,
			UserName:   history.UserName,
			CreatedAt:  history.CreatedAt,
		})
	}
	if err := historyRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert status history: %w", err)
	}

	notes, err := GetAlertNotes(ctx, pool, alertID)
	if err != nil {
		return nil, err
	}
	for _, note := range notes {
		items = append(items, models.AlertTimelineItem{
			ID:        note.ID,
			Type:      "note",
			Title:     "Note added",
			Content:   note.Content,
			UserID:    note.UserID,
			UserName:  note.UserName,
			CreatedAt: note.CreatedAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items, nil
}

// GetAlertNotes retrieves notes for an alert record ordered oldest-first.
func GetAlertNotes(ctx context.Context, pool *pgxpool.Pool, alertID string) ([]models.AlertNote, error) {
	if _, err := GetAlertRecordByID(ctx, pool, alertID); err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, alert_id, COALESCE(user_id::text, ''), user_name, content, created_at
		FROM alert_notes
		WHERE alert_id = $1
		ORDER BY created_at ASC, id ASC
	`, alertID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert notes: %w", err)
	}
	defer rows.Close()

	notes := make([]models.AlertNote, 0)
	for rows.Next() {
		var note models.AlertNote
		if err := rows.Scan(
			&note.ID,
			&note.AlertID,
			&note.UserID,
			&note.UserName,
			&note.Content,
			&note.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert note: %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert notes: %w", err)
	}

	return notes, nil
}

func resolveAlertActor(ctx context.Context, pool *pgxpool.Pool, userID *string) (interface{}, string) {
	if userID == nil || strings.TrimSpace(*userID) == "" {
		return nil, "System"
	}

	parsedUserID, err := uuid.Parse(strings.TrimSpace(*userID))
	if err != nil {
		return nil, "System"
	}

	var username string
	if err := pool.QueryRow(ctx, `SELECT username FROM users WHERE user_id = $1`, parsedUserID).Scan(&username); err != nil {
		return nil, "System"
	}

	return parsedUserID, username
}

// AlertRecordFilters represents filter parameters for querying alert records
type AlertRecordFilters struct {
	NodeID    *string
	Level     *string
	Status    *string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

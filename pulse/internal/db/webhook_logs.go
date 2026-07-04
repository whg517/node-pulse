package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// WebhookLogsQuerier defines database operations for webhook logs
type WebhookLogsQuerier interface {
	CreateWebhookLog(ctx context.Context, log *models.WebhookLog) error
	CountRecentWebhookLogs(ctx context.Context, totalCount, successCount *int64, limit int) error
	// GetWebhookLogs returns delivery logs for a webhook, newest first, plus the total count.
	GetWebhookLogs(ctx context.Context, webhookID string, limit, offset int) ([]*models.WebhookLog, int, error)
}

type webhookLogsQuerier struct {
	pool *pgxpool.Pool
}

// NewWebhookLogsQuerier creates a new WebhookLogsQuerier
func NewWebhookLogsQuerier(pool *pgxpool.Pool) WebhookLogsQuerier {
	return &webhookLogsQuerier{pool: pool}
}

// CreateWebhookLog creates a new webhook log entry
func (q *webhookLogsQuerier) CreateWebhookLog(ctx context.Context, log *models.WebhookLog) error {
	log.ID = uuid.New().String()
	log.SentAt = time.Now()

	query := `
		INSERT INTO webhook_logs (id, webhook_id, alert_event_id, status, retry_count, error_message, sent_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	_, err := q.pool.Exec(ctx, query,
		log.ID, log.WebhookID, log.AlertEventID, log.Status, log.RetryCount, log.ErrorMessage, log.SentAt)

	if err != nil {
		return errors.New("failed to create webhook log: " + err.Error())
	}

	return nil
}

// CountRecentWebhookLogs counts total and successful webhook logs from recent records
func (q *webhookLogsQuerier) CountRecentWebhookLogs(ctx context.Context, totalCount, successCount *int64, limit int) error {
	query := `
		SELECT
			COUNT(*) as total_count,
			COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) as success_count
		FROM webhook_logs
		WHERE id IN (
			SELECT id FROM webhook_logs
			ORDER BY created_at DESC
			LIMIT $1
		)
	`

	err := q.pool.QueryRow(ctx, query, limit).Scan(totalCount, successCount)
	if err != nil {
		return errors.New("failed to count recent webhook logs: " + err.Error())
	}

	return nil
}

// GetWebhookLogs returns delivery logs for a webhook ordered newest-first with the total count.
func (q *webhookLogsQuerier) GetWebhookLogs(ctx context.Context, webhookID string, limit, offset int) ([]*models.WebhookLog, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM webhook_logs WHERE webhook_id = $1`
	if err := q.pool.QueryRow(ctx, countQuery, webhookID).Scan(&total); err != nil {
		return nil, 0, errors.New("failed to count webhook logs: " + err.Error())
	}

	query := `
		SELECT id, webhook_id, alert_event_id, status, retry_count, error_message, sent_at, created_at
		FROM webhook_logs
		WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := q.pool.Query(ctx, query, webhookID, limit, offset)
	if err != nil {
		return nil, 0, errors.New("failed to query webhook logs: " + err.Error())
	}
	defer rows.Close()

	logs := make([]*models.WebhookLog, 0)
	for rows.Next() {
		log := &models.WebhookLog{}
		if err := rows.Scan(
			&log.ID, &log.WebhookID, &log.AlertEventID, &log.Status,
			&log.RetryCount, &log.ErrorMessage, &log.SentAt, &log.CreatedAt,
		); err != nil {
			return nil, 0, errors.New("failed to scan webhook log: " + err.Error())
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

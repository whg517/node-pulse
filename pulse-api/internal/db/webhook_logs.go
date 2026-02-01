package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

// WebhookLogsQuerier defines database operations for webhook logs
type WebhookLogsQuerier interface {
	CreateWebhookLog(ctx context.Context, log *models.WebhookLog) error
	CountRecentWebhookLogs(ctx context.Context, totalCount, successCount *int64, limit int) error
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

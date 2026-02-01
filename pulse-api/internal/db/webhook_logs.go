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

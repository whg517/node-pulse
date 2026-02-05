package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// WebhookQuerier defines webhook database operations
type WebhookQuerier interface {
	CreateWebhook(ctx context.Context, webhook *models.Webhook) error
	GetWebhooks(ctx context.Context) ([]*models.Webhook, error)
	GetWebhookByID(ctx context.Context, id string) (*models.Webhook, error)
	UpdateWebhook(ctx context.Context, id string, update *models.UpdateWebhookRequest) (*models.Webhook, error)
	DeleteWebhook(ctx context.Context, id string) error
}

type webhookQuerier struct {
	pool *pgxpool.Pool
}

// NewWebhookQuerier creates a new webhook querier
func NewWebhookQuerier(pool *pgxpool.Pool) WebhookQuerier {
	return &webhookQuerier{pool: pool}
}

// CreateWebhook creates a new webhook configuration
func (q *webhookQuerier) CreateWebhook(ctx context.Context, webhook *models.Webhook) error {
	webhook.ID = uuid.New().String()

	// Set default event format if not provided
	if webhook.EventFormat == nil {
		webhook.EventFormat = map[string]interface{}{
			"version": "1.0",
		}
	}

	eventFormatJSON, err := json.Marshal(webhook.EventFormat)
	if err != nil {
		return fmt.Errorf("failed to marshal event format: %w", err)
	}

	query := `
		INSERT INTO webhooks (id, url, event_format, enabled, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at
	`

	err = q.pool.QueryRow(ctx, query,
		webhook.ID, webhook.URL, eventFormatJSON, webhook.Enabled,
	).Scan(&webhook.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetWebhooks retrieves all webhook configurations
func (q *webhookQuerier) GetWebhooks(ctx context.Context) ([]*models.Webhook, error) {
	query := `
		SELECT id, url, event_format, enabled, created_at
		FROM webhooks
		ORDER BY created_at DESC
	`

	rows, err := q.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	webhooks := []*models.Webhook{}
	for rows.Next() {
		webhook := &models.Webhook{}
		var eventFormatJSON []byte

		err := rows.Scan(
			&webhook.ID, &webhook.URL, &eventFormatJSON,
			&webhook.Enabled, &webhook.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(eventFormatJSON) > 0 {
			err = json.Unmarshal(eventFormatJSON, &webhook.EventFormat)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal event format: %w", err)
			}
		}

		webhooks = append(webhooks, webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return webhooks, nil
}

// GetWebhookByID retrieves a single webhook configuration by ID
func (q *webhookQuerier) GetWebhookByID(ctx context.Context, id string) (*models.Webhook, error) {
	webhook := &models.Webhook{}
	var eventFormatJSON []byte

	query := `
		SELECT id, url, event_format, enabled, created_at
		FROM webhooks
		WHERE id = $1
	`

	err := q.pool.QueryRow(ctx, query, id).Scan(
		&webhook.ID, &webhook.URL, &eventFormatJSON,
		&webhook.Enabled, &webhook.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("webhook not found")
		}
		return nil, err
	}

	if len(eventFormatJSON) > 0 {
		err = json.Unmarshal(eventFormatJSON, &webhook.EventFormat)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal event format: %w", err)
		}
	}

	return webhook, nil
}

// UpdateWebhook updates an existing webhook configuration
func (q *webhookQuerier) UpdateWebhook(ctx context.Context, id string, update *models.UpdateWebhookRequest) (*models.Webhook, error) {
	// Build dynamic UPDATE query based on provided fields
	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	if update.URL != nil {
		setClauses = append(setClauses, fmt.Sprintf("url = $%d", argCount))
		args = append(args, *update.URL)
		argCount++
	}

	if update.EventFormat != nil {
		setClauses = append(setClauses, fmt.Sprintf("event_format = $%d", argCount))
		eventFormatJSON, err := json.Marshal(*update.EventFormat)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event format: %w", err)
		}
		args = append(args, eventFormatJSON)
		argCount++
	}

	if update.Enabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("enabled = $%d", argCount))
		args = append(args, *update.Enabled)
		argCount++
	}

	if len(setClauses) == 0 {
		return q.GetWebhookByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE webhooks
		SET %s
		WHERE id = $%d
		RETURNING id, url, event_format, enabled, created_at
	`, strings.Join(setClauses, ", "), argCount)

	args = append(args, id)

	webhook := &models.Webhook{}
	var eventFormatJSON []byte

	err := q.pool.QueryRow(ctx, query, args...).Scan(
		&webhook.ID, &webhook.URL, &eventFormatJSON,
		&webhook.Enabled, &webhook.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("webhook not found")
		}
		return nil, err
	}

	if len(eventFormatJSON) > 0 {
		err = json.Unmarshal(eventFormatJSON, &webhook.EventFormat)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal event format: %w", err)
		}
	}

	return webhook, nil
}

// DeleteWebhook deletes a webhook configuration
func (q *webhookQuerier) DeleteWebhook(ctx context.Context, id string) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	result, err := q.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("webhook not found")
	}

	return nil
}

// MockWebhookQuerier is a mock implementation for testing
type MockWebhookQuerier struct {
	Webhooks map[string]*models.Webhook
}

func (m *MockWebhookQuerier) CreateWebhook(ctx context.Context, webhook *models.Webhook) error {
	if m.Webhooks == nil {
		m.Webhooks = make(map[string]*models.Webhook)
	}
	webhook.ID = "test-" + webhook.ID
	m.Webhooks[webhook.ID] = webhook
	return nil
}

func (m *MockWebhookQuerier) GetWebhooks(ctx context.Context) ([]*models.Webhook, error) {
	webhooks := make([]*models.Webhook, 0, len(m.Webhooks))
	for _, w := range m.Webhooks {
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (m *MockWebhookQuerier) GetWebhookByID(ctx context.Context, id string) (*models.Webhook, error) {
	webhook, exists := m.Webhooks[id]
	if !exists {
		return nil, errors.New("webhook not found")
	}
	return webhook, nil
}

func (m *MockWebhookQuerier) UpdateWebhook(ctx context.Context, id string, update *models.UpdateWebhookRequest) (*models.Webhook, error) {
	webhook, exists := m.Webhooks[id]
	if !exists {
		return nil, errors.New("webhook not found")
	}

	if update.URL != nil {
		webhook.URL = *update.URL
	}
	if update.EventFormat != nil {
		webhook.EventFormat = *update.EventFormat
	}
	if update.Enabled != nil {
		webhook.Enabled = *update.Enabled
	}

	return webhook, nil
}

func (m *MockWebhookQuerier) DeleteWebhook(ctx context.Context, id string) error {
	if _, exists := m.Webhooks[id]; !exists {
		return errors.New("webhook not found")
	}
	delete(m.Webhooks, id)
	return nil
}

package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

func setupWebhookTestDB(t *testing.T) *pgxpool.Pool {
	pool, _ := setupTestDB(t)

	// webhooks table is already created by Migrate in setupTestDB
	return pool
}

func TestCreateWebhook(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	eventFormat := map[string]interface{}{
		"version": "1.0",
		"test":    "data",
	}

	webhook := &models.Webhook{
		URL:         "https://example.com/webhook",
		EventFormat: eventFormat,
		Enabled:     true,
	}

	err := querier.CreateWebhook(ctx, webhook)
	require.NoError(t, err)
	assert.NotEmpty(t, webhook.ID)
	assert.False(t, webhook.CreatedAt.IsZero())
}

func TestCreateWebhookWithDefaultEventFormat(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	webhook := &models.Webhook{
		URL:     "https://example.com/webhook",
		Enabled: true,
	}

	err := querier.CreateWebhook(ctx, webhook)
	require.NoError(t, err)
	assert.NotEmpty(t, webhook.ID)
	assert.NotNil(t, webhook.EventFormat)
}

func TestGetWebhooks(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	// Create multiple webhooks
	webhooks := []*models.Webhook{
		{URL: "https://example.com/hook1", Enabled: true},
		{URL: "https://example.com/hook2", Enabled: false},
	}

	for _, w := range webhooks {
		err := querier.CreateWebhook(ctx, w)
		require.NoError(t, err)
	}

	// Retrieve all webhooks
	retrieved, err := querier.GetWebhooks(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(retrieved), 2)
}

func TestGetWebhookByID(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	eventFormat := map[string]interface{}{
		"version": "1.0",
		"custom":  "field",
	}

	webhook := &models.Webhook{
		URL:         "https://example.com/webhook",
		EventFormat: eventFormat,
		Enabled:     true,
	}

	err := querier.CreateWebhook(ctx, webhook)
	require.NoError(t, err)

	// Retrieve by ID
	retrieved, err := querier.GetWebhookByID(ctx, webhook.ID)
	require.NoError(t, err)
	assert.Equal(t, webhook.ID, retrieved.ID)
	assert.Equal(t, webhook.URL, retrieved.URL)
	assert.Equal(t, webhook.Enabled, retrieved.Enabled)
	assert.NotNil(t, retrieved.EventFormat)
}

func TestGetWebhookByIDNotFound(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	// Try to retrieve non-existent webhook
	_, err := querier.GetWebhookByID(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
	assert.Equal(t, "webhook not found", err.Error())
}

func TestUpdateWebhookURL(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	webhook := &models.Webhook{
		URL:     "https://example.com/old",
		Enabled: true,
	}

	err := querier.CreateWebhook(ctx, webhook)
	require.NoError(t, err)

	// Update URL
	newURL := "https://example.com/new"
	update := &models.UpdateWebhookRequest{
		URL: &newURL,
	}

	updated, err := querier.UpdateWebhook(ctx, webhook.ID, update)
	require.NoError(t, err)
	assert.Equal(t, newURL, updated.URL)
	assert.Equal(t, webhook.Enabled, updated.Enabled)
}

func TestUpdateWebhookEventFormat(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	webhook := &models.Webhook{
		URL:         "https://example.com/webhook",
		EventFormat: models.DefaultEventFormat,
		Enabled:     true,
	}

	err := querier.CreateWebhook(ctx, webhook)
	require.NoError(t, err)

	// Update event format
	newEventFormat := map[string]interface{}{
		"version": "2.0",
		"custom":  "format",
	}
	update := &models.UpdateWebhookRequest{
		EventFormat: &newEventFormat,
	}

	updated, err := querier.UpdateWebhook(ctx, webhook.ID, update)
	require.NoError(t, err)
	assert.NotNil(t, updated.EventFormat)
	assert.Equal(t, "2.0", updated.EventFormat["version"])
}

func TestUpdateWebhookAllFields(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	webhook := &models.Webhook{
		URL:     "https://example.com/old",
		Enabled: true,
	}

	err := querier.CreateWebhook(ctx, webhook)
	require.NoError(t, err)

	// Update all fields
	newURL := "https://example.com/new"
	newEventFormat := map[string]interface{}{"updated": true}
	enabled := false

	update := &models.UpdateWebhookRequest{
		URL:         &newURL,
		EventFormat: &newEventFormat,
		Enabled:     &enabled,
	}

	updated, err := querier.UpdateWebhook(ctx, webhook.ID, update)
	require.NoError(t, err)
	assert.Equal(t, newURL, updated.URL)
	assert.False(t, updated.Enabled)
	assert.NotNil(t, updated.EventFormat)
}

func TestUpdateWebhookNotFound(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	newURL := "https://example.com/new"
	update := &models.UpdateWebhookRequest{
		URL: &newURL,
	}

	// Try to update non-existent webhook
	_, err := querier.UpdateWebhook(ctx, "00000000-0000-0000-0000-000000000000", update)
	assert.Error(t, err)
	assert.Equal(t, "webhook not found", err.Error())
}

func TestDeleteWebhook(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	webhook := &models.Webhook{
		URL:     "https://example.com/webhook",
		Enabled: true,
	}

	err := querier.CreateWebhook(ctx, webhook)
	require.NoError(t, err)

	// Delete webhook
	err = querier.DeleteWebhook(ctx, webhook.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = querier.GetWebhookByID(ctx, webhook.ID)
	assert.Error(t, err)
	assert.Equal(t, "webhook not found", err.Error())
}

func TestDeleteWebhookNotFound(t *testing.T) {
	pool := setupWebhookTestDB(t)
	defer pool.Close()

	querier := NewWebhookQuerier(pool)
	ctx := context.Background()

	// Try to delete non-existent webhook
	err := querier.DeleteWebhook(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
	assert.Equal(t, "webhook not found", err.Error())
}

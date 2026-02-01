package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

func setupAlertEventsTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Connect to test database
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://postgres:postgres@localhost:5432/node_pulse_test?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, testDBURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Run migrations
	err = Migrate(ctx, pool)
	require.NoError(t, err, "Failed to run migrations")

	// Create test node for foreign key reference
	nodeQuerier := NewPoolQuerier(pool)
	testNodeID := uuid.New()
	err = nodeQuerier.CreateNode(ctx, testNodeID, "test-alert-node", "192.168.1.200", "test-region", map[string]interface{}{
		"test": "true",
	})
	require.NoError(t, err, "Failed to create test node")

	// Cleanup function
	cleanup := func() {
		pool.Exec(ctx, "DROP TABLE IF EXISTS alert_events CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS alerts CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS webhooks CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
		pool.Close()
	}

	return pool, cleanup
}

func TestAlertEventsQuerier_CreateAlertEvent(t *testing.T) {
	pool, cleanup := setupAlertEventsTestDB(t)
	defer cleanup()

	ctx := context.Background()
	querier := NewAlertEventsQuerier(pool)

	nodeQuerier := NewPoolQuerier(pool)
	nodes, err := nodeQuerier.GetNodes(ctx)
	require.NoError(t, err)
	require.Greater(t, len(nodes), 0, "Need at least one test node")
	nodeID := nodes[0].ID

	t.Run("Create alert event successfully", func(t *testing.T) {
		event := &models.AlertEvent{
			NodeID:       nodeID,
			Metric:       "latency",
			Threshold:    100.0,
			CurrentValue: 150.0,
			Level:        "P0",
			CreatedAt:    time.Now(),
		}

		err := querier.CreateAlertEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID, "Event ID should be generated")
	})

	t.Run("Create packet loss alert event", func(t *testing.T) {
		event := &models.AlertEvent{
			NodeID:       nodeID,
			Metric:       "packet_loss_rate",
			Threshold:    5.0,
			CurrentValue: 10.5,
			Level:        "P1",
			CreatedAt:    time.Now(),
		}

		err := querier.CreateAlertEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("Create jitter alert event", func(t *testing.T) {
		event := &models.AlertEvent{
			NodeID:       nodeID,
			Metric:       "jitter",
			Threshold:    8.0,
			CurrentValue: 12.3,
			Level:        "P2",
			CreatedAt:    time.Now(),
		}

		err := querier.CreateAlertEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("Create alert event with invalid node ID", func(t *testing.T) {
		event := &models.AlertEvent{
			NodeID:       "00000000-0000-0000-0000-000000000000", // Non-existent node
			Metric:       "latency",
			Threshold:    100.0,
			CurrentValue: 150.0,
			Level:        "P0",
			CreatedAt:    time.Now(),
		}

		err := querier.CreateAlertEvent(ctx, event)
		assert.Error(t, err, "Should fail with foreign key constraint violation")
	})
}

func TestAlertEventsQuerier_MultipleEvents(t *testing.T) {
	pool, cleanup := setupAlertEventsTestDB(t)
	defer cleanup()

	ctx := context.Background()
	querier := NewAlertEventsQuerier(pool)

	nodeQuerier := NewPoolQuerier(pool)
	nodes, err := nodeQuerier.GetNodes(ctx)
	require.NoError(t, err)
	require.Greater(t, len(nodes), 0)
	nodeID := nodes[0].ID

	t.Run("Create multiple alert events for same node", func(t *testing.T) {
		// Create multiple events
		events := []*models.AlertEvent{
			{
				NodeID:       nodeID,
				Metric:       "latency",
				Threshold:    100.0,
				CurrentValue: 150.0,
				Level:        "P0",
				CreatedAt:    time.Now().Add(-2 * time.Minute),
			},
			{
				NodeID:       nodeID,
				Metric:       "latency",
				Threshold:    100.0,
				CurrentValue: 180.0,
				Level:        "P0",
				CreatedAt:    time.Now().Add(-1 * time.Minute),
			},
			{
				NodeID:       nodeID,
				Metric:       "packet_loss_rate",
				Threshold:    5.0,
				CurrentValue: 8.5,
				Level:        "P1",
				CreatedAt:    time.Now(),
			},
		}

		for _, event := range events {
			err := querier.CreateAlertEvent(ctx, event)
			require.NoError(t, err)
			assert.NotEmpty(t, event.ID)
		}

		// Note: We don't have a GetAlertEvents method yet, so we can't query them back
		// This would be added in Story 6.1 (Alert Record Storage API)
	})
}

package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
)

func setupAlertEventsTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	// Setup test config
	testutil.SetupTestConfig()

	// Load config before migrations (seedAdminUser needs config)
	config.MustLoad()

	ctx := context.Background()

	// Connect to test database
	testDBURL := testutil.GetTestDBURL()

	pool, err := pgxpool.New(ctx, testDBURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil, nil
	}

	// Clean up any existing tables to ensure fresh schema
	pool.Exec(ctx, "DROP TABLE IF EXISTS alert_events CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS alerts CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS webhooks CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS webhook_logs CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS alert_suppressions CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")

	// Run migrations to create fresh tables with proper constraints
	err = Migrate(ctx, pool)
	if err != nil {
		pool.Close()
		t.Fatalf("Failed to run migrations: %v", err)
		return nil, nil
	}

	// Create test node for foreign key reference
	testNodeID := uuid.New()
	err = CreateNode(ctx, pool, testNodeID, "test-alert-node", "192.168.1.200", "test-region", nil)
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
		testutil.TeardownTestConfig()
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
		// Use a random UUID that definitely doesn't exist
		invalidNodeID := uuid.New()
		event := &models.AlertEvent{
			NodeID:       invalidNodeID.String(), // Non-existent node
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

	nodes, err := GetNodes(ctx, pool)
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

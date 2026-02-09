package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

func setupAlertTestDB(t *testing.T) *pgxpool.Pool {
	pool, _ := SetupTestDB(t)

	// Clean up before each test
	_, err := pool.Exec(context.Background(), "DELETE FROM alerts WHERE 1=1")
	require.NoError(t, err)

	return pool
}

func TestCreateAlert(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	ctx := context.Background()

	// Test creating a global alert (no node_id)
	alert := &models.Alert{
		Metric:    "latency",
		Threshold: 100.5,
		Level:     "P1",
		NodeID:    nil,
		Enabled:   true,
	}

	err := querier.CreateAlert(ctx, alert)
	require.NoError(t, err)
	assert.NotEmpty(t, alert.ID)
	assert.False(t, alert.CreatedAt.IsZero())

	// Verify the alert was created
	fetched, err := querier.GetAlertByID(ctx, alert.ID)
	require.NoError(t, err)
	assert.Equal(t, alert.Metric, fetched.Metric)
	assert.Equal(t, alert.Threshold, fetched.Threshold)
	assert.Equal(t, alert.Level, fetched.Level)
	assert.Nil(t, fetched.NodeID)
	assert.True(t, fetched.Enabled)
}

func TestCreateAlertWithNode(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	nodeQuerier := NewPoolQuerier(pool)
	ctx := context.Background()

	// Create a test node first
	nodeID := uuid.New()
	err := nodeQuerier.CreateNode(ctx, nodeID, "Test Node", "192.168.1.100", "us-east", map[string]interface{}{})
	require.NoError(t, err)

	nodeIDStr := nodeID.String()

	// Test creating a node-specific alert
	alert := &models.Alert{
		Metric:    "packet_loss_rate",
		Threshold: 5.0,
		Level:     "P0",
		NodeID:    &nodeIDStr,
		Enabled:   true,
	}

	err = querier.CreateAlert(ctx, alert)
	require.NoError(t, err)
	assert.NotNil(t, alert.NodeID)
	assert.Equal(t, nodeIDStr, *alert.NodeID)

	// Verify the alert was created
	fetched, err := querier.GetAlertByID(ctx, alert.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetched.NodeID)
	assert.Equal(t, nodeIDStr, *fetched.NodeID)
}

func TestGetAlerts(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	ctx := context.Background()

	// Create multiple alerts
	alerts := []*models.Alert{
		{Metric: "latency", Threshold: 100.0, Level: "P1", NodeID: nil, Enabled: true},
		{Metric: "packet_loss_rate", Threshold: 5.0, Level: "P0", NodeID: nil, Enabled: true},
		{Metric: "jitter", Threshold: 10.0, Level: "P2", NodeID: nil, Enabled: false},
	}

	for _, alert := range alerts {
		err := querier.CreateAlert(ctx, alert)
		require.NoError(t, err)
	}

	// Test getting all alerts
	fetched, err := querier.GetAlerts(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, fetched, 3)

	// Test filtering by node_id (should return none since we created global alerts)
	nonExistentNodeID := uuid.New().String()
	fetchedFiltered, err := querier.GetAlerts(ctx, &nonExistentNodeID)
	require.NoError(t, err)
	assert.Len(t, fetchedFiltered, 0)
}

func TestGetAlertByID(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	ctx := context.Background()

	// Create an alert
	alert := &models.Alert{
		Metric:    "latency",
		Threshold: 100.0,
		Level:     "P1",
		NodeID:    nil,
		Enabled:   true,
	}
	err := querier.CreateAlert(ctx, alert)
	require.NoError(t, err)

	// Test getting by valid ID
	fetched, err := querier.GetAlertByID(ctx, alert.ID)
	require.NoError(t, err)
	assert.Equal(t, alert.ID, fetched.ID)
	assert.Equal(t, alert.Metric, fetched.Metric)

	// Test getting by invalid ID (valid UUID but doesn't exist)
	nonExistentID := uuid.New().String()
	_, err = querier.GetAlertByID(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateAlert(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	ctx := context.Background()

	// Create an alert
	alert := &models.Alert{
		Metric:    "latency",
		Threshold: 100.0,
		Level:     "P1",
		NodeID:    nil,
		Enabled:   true,
	}
	err := querier.CreateAlert(ctx, alert)
	require.NoError(t, err)

	// Update threshold
	newThreshold := 200.0
	update := &models.UpdateAlertRequest{
		Threshold: &newThreshold,
	}

	updated, err := querier.UpdateAlert(ctx, alert.ID, update)
	require.NoError(t, err)
	assert.Equal(t, newThreshold, updated.Threshold)
	assert.Equal(t, alert.Metric, updated.Metric) // Other fields unchanged

	// Update multiple fields
	newLevel := "P0"
	enabled := false
	update = &models.UpdateAlertRequest{
		Level:   &newLevel,
		Enabled: &enabled,
	}

	updated, err = querier.UpdateAlert(ctx, alert.ID, update)
	require.NoError(t, err)
	assert.Equal(t, newLevel, updated.Level)
	assert.False(t, updated.Enabled)
	assert.Equal(t, newThreshold, updated.Threshold) // Previous update preserved
}

func TestUpdateAlertNotFound(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	ctx := context.Background()

	nonExistentID := uuid.New().String()
	newLevel := "P0"
	update := &models.UpdateAlertRequest{
		Level: &newLevel,
	}

	_, err := querier.UpdateAlert(ctx, nonExistentID, update)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteAlert(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	ctx := context.Background()

	// Create an alert
	alert := &models.Alert{
		Metric:    "latency",
		Threshold: 100.0,
		Level:     "P1",
		NodeID:    nil,
		Enabled:   true,
	}
	err := querier.CreateAlert(ctx, alert)
	require.NoError(t, err)

	// Delete the alert
	err = querier.DeleteAlert(ctx, alert.ID)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = querier.GetAlertByID(ctx, alert.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteAlertNotFound(t *testing.T) {
	pool := setupAlertTestDB(t)
	querier := NewAlertQuerier(pool)
	ctx := context.Background()

	nonExistentID := uuid.New().String()
	err := querier.DeleteAlert(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

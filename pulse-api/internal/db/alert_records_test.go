package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

func setupAlertRecordsTest(t *testing.T) (*pgxpool.Pool, func()) {
	pool, cleanup := setupTestDB(t)

	// Create alert records table
	ctx := context.Background()
	err := createAlertRecordsTable(ctx, pool)
	require.NoError(t, err, "Failed to create alert_records table")

	return pool, func() {
		// Additional cleanup for alert_records
		pool.Exec(ctx, "DROP TABLE IF EXISTS alert_records CASCADE")
		cleanup()
	}
}

func TestCreateAlertRecord(t *testing.T) {
	pool, cleanup := setupAlertRecordsTest(t)
	defer cleanup()

	ctx := context.Background()

	// First create a node and alert event
	nodeID := uuid.New()
	err := CreateNode(ctx, pool, nodeID, "Test Node", "192.168.1.1", "us-east", nil)
	require.NoError(t, err)

	alertEvent := &models.AlertEvent{
		ID:           uuid.New().String(),
		NodeID:       nodeID.String(),
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	// Create alert event
	alertEventsQuerier := NewAlertEventsQuerier(pool)
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	// Create alert record
	record := &models.AlertRecord{
		AlertEventID: alertEvent.ID,
		NodeID:       nodeID.String(),
		Metric:       alertEvent.Metric,
		Level:        alertEvent.Level,
		Status:       "pending",
	}

	err = CreateAlertRecord(ctx, pool, record)
	require.NoError(t, err)

	// Verify record was created
	assert.NotEmpty(t, record.ID, "Record ID should be generated")
	assert.False(t, record.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, record.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.Equal(t, "pending", record.Status, "Status should be pending")
}

func TestGetAlertRecords(t *testing.T) {
	pool, cleanup := setupAlertRecordsTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data
	nodeID := uuid.New()
	err := CreateNode(ctx, pool, nodeID, "Test Node", "192.168.1.1", "us-east", nil)
	require.NoError(t, err)

	// Create multiple alert events and records
	alertEventsQuerier := NewAlertEventsQuerier(pool)
	records := []models.AlertRecord{}

	for i := 0; i < 3; i++ {
		alertEvent := &models.AlertEvent{
			ID:           uuid.New().String(),
			NodeID:       nodeID.String(),
			Metric:       "latency",
			Threshold:    100.0,
			CurrentValue: float64(100 + i*10),
			Level:        "P0",
			CreatedAt:    time.Now(),
		}
		err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
		require.NoError(t, err)

		record := &models.AlertRecord{
			AlertEventID: alertEvent.ID,
			NodeID:       nodeID.String(),
			Metric:       alertEvent.Metric,
			Level:        alertEvent.Level,
			Status:       "pending",
		}
		err = CreateAlertRecord(ctx, pool, record)
		require.NoError(t, err)
		records = append(records, *record)
	}

	// Test getting all records
	filters := AlertRecordFilters{
		Limit:  10,
		Offset: 0,
	}

	results, err := GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.Len(t, results, 3, "Should return 3 records")

	// Test filtering by node ID
	nodeIDStr := nodeID.String()
	filters = AlertRecordFilters{
		NodeID: &nodeIDStr,
		Limit:  10,
		Offset: 0,
	}

	results, err = GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.Len(t, results, 3, "Should return 3 records for node")

	// Test filtering by status
	status := "pending"
	filters = AlertRecordFilters{
		Status: &status,
		Limit:  10,
		Offset: 0,
	}

	results, err = GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3, "Should return at least 3 pending records")

	// Test filtering by level
	level := "P0"
	filters = AlertRecordFilters{
		Level: &level,
		Limit:  10,
		Offset: 0,
	}

	results, err = GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3, "Should return at least 3 P0 records")

	// Test pagination
	filters = AlertRecordFilters{
		Limit:  2,
		Offset: 0,
	}

	results, err = GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.Len(t, results, 2, "Should return 2 records with limit=2")

	// Test offset
	filters = AlertRecordFilters{
		Limit:  2,
		Offset: 1,
	}

	results, err = GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(results), 2, "Should return at most 2 records with offset=1")
}

func TestGetAlertRecordByID(t *testing.T) {
	pool, cleanup := setupAlertRecordsTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data
	nodeID := uuid.New()
	err := CreateNode(ctx, pool, nodeID, "Test Node", "192.168.1.1", "us-east", nil)
	require.NoError(t, err)

	alertEvent := &models.AlertEvent{
		ID:           uuid.New().String(),
		NodeID:       nodeID.String(),
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	alertEventsQuerier := NewAlertEventsQuerier(pool)
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	record := &models.AlertRecord{
		AlertEventID: alertEvent.ID,
		NodeID:       nodeID.String(),
		Metric:       alertEvent.Metric,
		Level:        alertEvent.Level,
		Status:       "pending",
	}
	err = CreateAlertRecord(ctx, pool, record)
	require.NoError(t, err)

	// Test getting record by ID
	result, err := GetAlertRecordByID(ctx, pool, record.ID)
	require.NoError(t, err)
	assert.Equal(t, record.ID, result.ID, "Record ID should match")
	assert.Equal(t, record.AlertEventID, result.AlertEventID, "AlertEventID should match")
	assert.Equal(t, record.NodeID, result.NodeID, "NodeID should match")
	assert.Equal(t, record.Metric, result.Metric, "Metric should match")
	assert.Equal(t, record.Level, result.Level, "Level should match")
	assert.Equal(t, record.Status, result.Status, "Status should match")

	// Test getting non-existent record
	_, err = GetAlertRecordByID(ctx, pool, uuid.New().String())
	assert.Error(t, err, "Should return error for non-existent record")
}

func TestUpdateAlertRecordStatus(t *testing.T) {
	pool, cleanup := setupAlertRecordsTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data
	nodeID := uuid.New()
	err := CreateNode(ctx, pool, nodeID, "Test Node", "192.168.1.1", "us-east", nil)
	require.NoError(t, err)

	alertEvent := &models.AlertEvent{
		ID:           uuid.New().String(),
		NodeID:       nodeID.String(),
		Metric:       "latency",
		Threshold:    100.0,
		CurrentValue: 150.0,
		Level:        "P0",
		CreatedAt:    time.Now(),
	}

	alertEventsQuerier := NewAlertEventsQuerier(pool)
	err = alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
	require.NoError(t, err)

	record := &models.AlertRecord{
		AlertEventID: alertEvent.ID,
		NodeID:       nodeID.String(),
		Metric:       alertEvent.Metric,
		Level:        alertEvent.Level,
		Status:       "pending",
	}
	err = CreateAlertRecord(ctx, pool, record)
	require.NoError(t, err)

	// Test valid transition: pending -> in_progress
	err = UpdateAlertRecordStatus(ctx, pool, record.ID, "in_progress")
	require.NoError(t, err, "Should allow pending -> in_progress transition")

	// Verify status was updated
	updatedRecord, err := GetAlertRecordByID(ctx, pool, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", updatedRecord.Status, "Status should be updated to in_progress")

	// Test valid transition: in_progress -> resolved
	err = UpdateAlertRecordStatus(ctx, pool, record.ID, "resolved")
	require.NoError(t, err, "Should allow in_progress -> resolved transition")

	updatedRecord, err = GetAlertRecordByID(ctx, pool, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "resolved", updatedRecord.Status, "Status should be updated to resolved")

	// Test invalid transition: resolved -> in_progress (not allowed in MVP)
	err = UpdateAlertRecordStatus(ctx, pool, record.ID, "in_progress")
	assert.Error(t, err, "Should not allow resolved -> in_progress transition")
	assert.Contains(t, err.Error(), "invalid status transition", "Error should mention invalid transition")

	// Test updating non-existent record
	err = UpdateAlertRecordStatus(ctx, pool, uuid.New().String(), "resolved")
	assert.Error(t, err, "Should return error for non-existent record")
	assert.Contains(t, err.Error(), "not found", "Error should mention not found")
}

func TestAlertRecordFilters(t *testing.T) {
	pool, cleanup := setupAlertRecordsTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test nodes
	nodes := []models.Node{
		{ID: uuid.New().String(), Name: "Node 1", IP: "192.168.1.1", Region: "us-east"},
		{ID: uuid.New().String(), Name: "Node 2", IP: "192.168.1.2", Region: "eu-west"},
	}
	for _, node := range nodes {
		nodeUUID, _ := uuid.Parse(node.ID)
		err := CreateNode(ctx, pool, nodeUUID, node.Name, node.IP, node.Region, nil)
		require.NoError(t, err)
	}

	alertEventsQuerier := NewAlertEventsQuerier(pool)

	// Create alert records for different nodes and levels
	records := []struct {
		nodeID      string
		level       string
		status      string
		createdAt   time.Time
	}{
		{nodes[0].ID, "P0", "pending", time.Now().Add(-2 * time.Hour)},
		{nodes[0].ID, "P1", "in_progress", time.Now().Add(-1 * time.Hour)},
		{nodes[1].ID, "P0", "resolved", time.Now()},
	}

	for _, r := range records {
		alertEvent := &models.AlertEvent{
			ID:           uuid.New().String(),
			NodeID:       r.nodeID,
			Metric:       "latency",
			Threshold:    100.0,
			CurrentValue: 150.0,
			Level:        r.level,
			CreatedAt:    r.createdAt,
		}
		err := alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
		require.NoError(t, err)

		record := &models.AlertRecord{
			AlertEventID: alertEvent.ID,
			NodeID:       r.nodeID,
			Metric:       alertEvent.Metric,
			Level:        r.level,
			Status:       r.status,
		}
		err = CreateAlertRecord(ctx, pool, record)
		require.NoError(t, err)
	}

	// Test time range filtering
	startTime := time.Now().Add(-90 * time.Minute)
	endTime := time.Now()

	filters := AlertRecordFilters{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     10,
		Offset:    0,
	}

	results, err := GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1, "Should return records within time range")

	// Test combined filters
	nodeID := nodes[0].ID
	status := "pending"
	filters = AlertRecordFilters{
		NodeID: &nodeID,
		Status: &status,
		Limit:  10,
		Offset: 0,
	}

	results, err = GetAlertRecords(ctx, pool, filters)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1, "Should return records matching node and status")
}

package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndGetLatestMTRResult(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	nodeID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO nodes (id, name, ip, region, tags)
		VALUES ($1, 'node-mtr', '192.0.2.10', 'us-east', '{}')
	`, nodeID)
	require.NoError(t, err)

	input := MTRResultInput{
		NodeID:      nodeID,
		ProbeID:     "probe-mtr-1",
		Target:      "example.com",
		Success:     true,
		CompletedAt: time.Now().UTC().Truncate(time.Second),
		Hops: []MTRHop{
			{HopNumber: 1, IP: "192.0.2.1", Sent: 10, Received: 10, LossRate: 0, AvgRTTMs: 1.2},
			{HopNumber: 2, IP: "198.51.100.1", Sent: 10, Received: 9, LossRate: 10, AvgRTTMs: 12.5},
		},
	}

	saved, err := SaveMTRResult(ctx, pool, input)
	require.NoError(t, err)
	assert.Equal(t, 2, saved.TotalHops)

	latest, err := GetLatestMTRResult(ctx, pool, nodeID)
	require.NoError(t, err)
	assert.Equal(t, "example.com", latest.Target)
	assert.Equal(t, "probe-mtr-1", latest.ProbeID)
	require.Len(t, latest.Hops, 2)
	assert.Equal(t, "198.51.100.1", latest.Hops[1].IP)
}

func TestGetMTRResults(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	nodeID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO nodes (id, name, ip, region, tags)
		VALUES ($1, 'node-mtr-history', '192.0.2.20', 'us-east', '{}')
	`, nodeID)
	require.NoError(t, err)

	baseTime := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	inputs := []MTRResultInput{
		{NodeID: nodeID, ProbeID: "probe-1", Target: "example.com", Success: true, CompletedAt: baseTime, Hops: []MTRHop{{HopNumber: 1, IP: "192.0.2.1"}}},
		{NodeID: nodeID, ProbeID: "probe-1", Target: "example.org", Success: true, CompletedAt: baseTime.Add(30 * time.Minute), Hops: []MTRHop{{HopNumber: 1, IP: "192.0.2.2"}}},
		{NodeID: nodeID, ProbeID: "probe-1", Target: "example.com", Success: true, CompletedAt: baseTime.Add(60 * time.Minute), Hops: []MTRHop{{HopNumber: 1, IP: "192.0.2.3"}}},
	}
	for _, input := range inputs {
		_, err := SaveMTRResult(ctx, pool, input)
		require.NoError(t, err)
	}

	latestTwo, err := GetMTRResults(ctx, pool, MTRResultQuery{NodeID: nodeID, Limit: 2})
	require.NoError(t, err)
	require.Len(t, latestTwo, 2)
	assert.Equal(t, "192.0.2.3", latestTwo[0].Hops[0].IP)
	assert.Equal(t, "192.0.2.2", latestTwo[1].Hops[0].IP)

	targetResults, err := GetMTRResults(ctx, pool, MTRResultQuery{NodeID: nodeID, Target: "example.com"})
	require.NoError(t, err)
	require.Len(t, targetResults, 2)
	assert.Equal(t, "example.com", targetResults[0].Target)
	assert.Equal(t, "example.com", targetResults[1].Target)

	startTime := baseTime.Add(20 * time.Minute)
	endTime := baseTime.Add(45 * time.Minute)
	windowResults, err := GetMTRResults(ctx, pool, MTRResultQuery{
		NodeID:    nodeID,
		StartTime: &startTime,
		EndTime:   &endTime,
	})
	require.NoError(t, err)
	require.Len(t, windowResults, 1)
	assert.Equal(t, "example.org", windowResults[0].Target)
}

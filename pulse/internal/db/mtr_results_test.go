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

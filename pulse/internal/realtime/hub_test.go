package realtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	client := hub.Register("user-1", "admin")
	defer hub.Unregister(client)

	hub.Broadcast(Message{
		Type:    EventSystemHeartbeat,
		Payload: map[string]any{"ok": true},
	})

	select {
	case message := <-client.sendCh:
		assert.Equal(t, EventSystemHeartbeat, message.Type)
		assert.False(t, message.Timestamp.IsZero())
	case <-time.After(time.Second):
		t.Fatal("expected broadcast message")
	}
}

func TestHubBroadcastSkipsSlowClients(t *testing.T) {
	hub := NewHub()
	client := hub.Register("user-1", "admin")
	defer hub.Unregister(client)

	for i := 0; i < defaultClientQueueSize; i++ {
		client.sendCh <- Message{Type: EventSystemHeartbeat}
	}

	require.NotPanics(t, func() {
		hub.Broadcast(Message{Type: EventAlertNew})
	})
}

func TestAlertPayloadFromRecord(t *testing.T) {
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	record := &models.AlertRecord{
		ID:           "record-1",
		AlertEventID: "event-1",
		NodeID:       "node-1",
		Metric:       "latency",
		Level:        "P1",
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	payload := AlertPayloadFromRecord(record)

	assert.Equal(t, "record-1", payload.ID)
	assert.Equal(t, "event-1", payload.AlertEventID)
	assert.Equal(t, "node-1", payload.NodeID)
	assert.Equal(t, "latency", payload.Metric)
	assert.Equal(t, "P1", payload.Level)
	assert.Equal(t, "pending", payload.Status)
	assert.Equal(t, now.Format(time.RFC3339), payload.CreatedAt)
}

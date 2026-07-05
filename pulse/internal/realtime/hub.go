package realtime

import (
	"context"
	"sync"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

const (
	EventAlertNew         = "alert:new"
	EventAlertUpdated     = "alert:updated"
	EventAlertResolved    = "alert:resolved"
	EventAlertNoteCreated = "alert:note_created"
	EventNodeOnline       = "node:online"
	EventNodeOffline      = "node:offline"
	EventSystemHeartbeat  = "system:heartbeat"
	EventSystemError      = "system:error"
	EventPong             = "pong"
	defaultClientQueueSize = 64
)

// Message is the websocket envelope consumed by the frontend websocket service.
type Message struct {
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertPayload carries the alert fields the current dashboard notification flow expects.
type AlertPayload struct {
	ID           string  `json:"id"`
	AlertEventID string  `json:"alert_event_id,omitempty"`
	NodeID       string  `json:"node_id"`
	Metric       string  `json:"metric"`
	Level        string  `json:"level"`
	Status       string  `json:"status,omitempty"`
	Threshold    float64 `json:"threshold,omitempty"`
	CurrentValue float64 `json:"current_value,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	UpdatedAt    string  `json:"updated_at,omitempty"`
}

// AlertNotePayload carries the fields needed to append note events to alert timelines.
type AlertNotePayload struct {
	AlertID   string `json:"alert_id"`
	NoteID    string `json:"note_id"`
	UserID    string `json:"user_id,omitempty"`
	UserName  string `json:"user_name"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// NodeStatusPayload carries a node online/offline transition event. The frontend
// useGlobalRealtime hook consumes these to update nodesStore in real time.
type NodeStatusPayload struct {
	NodeID string `json:"node_id"`
	Status string `json:"status"`
}

// Client is a registered realtime connection.
type Client struct {
	id     string
	role   string
	sendCh chan Message
}

// Hub stores active realtime clients and fans out events without blocking callers.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// NewHub creates a realtime event hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
	}
}

// Register adds a client and returns its outbound message channel.
func (h *Hub) Register(userID, role string) *Client {
	client := &Client{
		id:     userID,
		role:   role,
		sendCh: make(chan Message, defaultClientQueueSize),
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	return client
}

// Unregister removes a client and closes its outbound channel.
func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.sendCh)
	}
	h.mu.Unlock()
}

// ClientCount returns the number of active websocket clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Broadcast sends an event to every connected client. Slow clients are skipped
// to keep alert ingestion and operator workflows non-blocking.
func (h *Hub) Broadcast(message Message) {
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now().UTC()
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.sendCh <- message:
		default:
		}
	}
}

// BroadcastAlertRecord sends an alert lifecycle event from a persisted record.
func (h *Hub) BroadcastAlertRecord(eventType string, record *models.AlertRecord) {
	if h == nil || record == nil {
		return
	}

	h.Broadcast(Message{
		Type:    eventType,
		Payload: AlertPayloadFromRecord(record),
	})
}

// BroadcastAlertNote sends a note-created event.
func (h *Hub) BroadcastAlertNote(note *models.AlertNote) {
	if h == nil || note == nil {
		return
	}

	h.Broadcast(Message{
		Type:    EventAlertNoteCreated,
		Payload: AlertNotePayloadFromNote(note),
	})
}

// BroadcastNodeStatus sends a node online/offline transition event. eventType
// must be EventNodeOnline or EventNodeOffline.
func (h *Hub) BroadcastNodeStatus(eventType, nodeID, status string) {
	if h == nil || nodeID == "" {
		return
	}

	h.Broadcast(Message{
		Type: eventType,
		Payload: NodeStatusPayload{
			NodeID: nodeID,
			Status: status,
		},
	})
}

func (c *Client) send(ctx context.Context, message Message) bool {
	select {
	case c.sendCh <- message:
		return true
	case <-ctx.Done():
		return false
	}
}

func AlertPayloadFromEvent(event *models.AlertEvent) AlertPayload {
	if event == nil {
		return AlertPayload{}
	}

	return AlertPayload{
		ID:           event.ID,
		AlertEventID: event.ID,
		NodeID:       event.NodeID,
		Metric:       event.Metric,
		Level:        event.Level,
		Threshold:    event.Threshold,
		CurrentValue: event.CurrentValue,
		CreatedAt:    event.CreatedAt.Format(time.RFC3339),
	}
}

func AlertPayloadFromRecord(record *models.AlertRecord) AlertPayload {
	if record == nil {
		return AlertPayload{}
	}

	return AlertPayload{
		ID:           record.ID,
		AlertEventID: record.AlertEventID,
		NodeID:       record.NodeID,
		Metric:       record.Metric,
		Level:        record.Level,
		Status:       record.Status,
		CreatedAt:    record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    record.UpdatedAt.Format(time.RFC3339),
	}
}

func AlertNotePayloadFromNote(note *models.AlertNote) AlertNotePayload {
	if note == nil {
		return AlertNotePayload{}
	}

	return AlertNotePayload{
		AlertID:   note.AlertID,
		NoteID:    note.ID,
		UserID:    note.UserID,
		UserName:  note.UserName,
		Content:   note.Content,
		CreatedAt: note.CreatedAt.Format(time.RFC3339),
	}
}

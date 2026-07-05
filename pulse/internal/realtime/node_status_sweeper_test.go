package realtime

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/whg517/node-pulse/pulse/internal/db"
)

// fakeRepo is a minimal sweeperRepo for unit-testing the scan/transition logic
// without a database. It records calls so tests can assert behavior.
type fakeRepo struct {
	mu                sync.Mutex
	stale             []db.StaleNode
	staleErr          error
	marked            []string // node IDs marked offline
	markErr           error
}

func (f *fakeRepo) GetStaleNodes(ctx context.Context, timeout time.Duration) ([]db.StaleNode, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stale, f.staleErr
}

func (f *fakeRepo) MarkNodeOfflineByString(ctx context.Context, nodeID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.markErr != nil {
		return f.markErr
	}
	f.marked = append(f.marked, nodeID)
	return nil
}

func (f *fakeRepo) markedIDs() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]string, len(f.marked))
	copy(cp, f.marked)
	return cp
}

// TestSweeper_ScanMarksStaleOffline verifies that scanOnce marks every stale
// node offline and broadcasts a node:offline event per node.
func TestSweeper_ScanMarksStaleOffline(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	repo := &fakeRepo{stale: []db.StaleNode{
		{ID: id1, Name: "node-1", Status: "online"},
		{ID: id2, Name: "node-2", Status: "connecting"},
	}}
	hub := NewHub()
	// Register a client to capture broadcasts.
	client := hub.Register("user-1", "admin")
	s := &NodeStatusSweeper{
		hub:      hub,
		repo:     repo,
		interval: time.Second,
		timeout:  db.HeartbeatTimeout,
		logger:   slog.Default(),
	}

	s.scanOnce(context.Background())

	marked := repo.markedIDs()
	if len(marked) != 2 {
		t.Fatalf("expected 2 nodes marked offline, got %d (%v)", len(marked), marked)
	}

	// Drain the client channel and confirm two node:offline events arrived.
	events := drainEvents(t, client, 2)
	for _, ev := range events {
		if ev.Type != EventNodeOffline {
			t.Errorf("expected event %q, got %q", EventNodeOffline, ev.Type)
		}
		payload, ok := ev.Payload.(NodeStatusPayload)
		if !ok {
			t.Fatalf("expected NodeStatusPayload, got %T", ev.Payload)
		}
		if payload.Status != "offline" {
			t.Errorf("expected status offline, got %q", payload.Status)
		}
	}
}

// TestSweeper_NoStaleNodesIsNoop verifies that an empty stale list produces no
// marks and no broadcasts.
func TestSweeper_NoStaleNodesIsNoop(t *testing.T) {
	repo := &fakeRepo{stale: nil}
	hub := NewHub()
	client := hub.Register("user-1", "admin")
	s := &NodeStatusSweeper{hub: hub, repo: repo, timeout: db.HeartbeatTimeout, logger: slog.Default()}

	s.scanOnce(context.Background())

	if len(repo.markedIDs()) != 0 {
		t.Errorf("expected no marks, got %v", repo.markedIDs())
	}
	if got := drainEvents(t, client, 0); len(got) != 0 {
		t.Errorf("expected no events, got %d", len(got))
	}
}

// TestSweeper_GetStaleErrorDoesNotMark verifies a DB query failure logs and
// skips marking rather than panicking.
func TestSweeper_GetStaleErrorDoesNotMark(t *testing.T) {
	repo := &fakeRepo{staleErr: context.DeadlineExceeded}
	hub := NewHub()
	s := &NodeStatusSweeper{hub: hub, repo: repo, timeout: db.HeartbeatTimeout, logger: slog.Default()}

	s.scanOnce(context.Background()) // must not panic

	if len(repo.markedIDs()) != 0 {
		t.Errorf("expected no marks on query error, got %v", repo.markedIDs())
	}
}

// TestBroadcastNodeStatus_NilHubIsSafe ensures a nil hub (e.g. tests, or when
// realtime is disabled) does not panic.
func TestBroadcastNodeStatus_NilHubIsSafe(t *testing.T) {
	var nilHub *Hub
	nilHub.BroadcastNodeStatus(EventNodeOnline, "x", "online") // must not panic
	nilHub.BroadcastNodeStatus(EventNodeOffline, "", "offline") // empty id, no-op
}

// drainEvents reads up to n messages from a client's channel without blocking.
func drainEvents(t *testing.T, client *Client, n int) []Message {
	t.Helper()
	var got []Message
	deadline := time.After(time.Second)
	for len(got) < n {
		select {
		case msg, ok := <-client.sendCh:
			if !ok {
				return got
			}
			got = append(got, msg)
		case <-deadline:
			return got
		}
	}
	return got
}

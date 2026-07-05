package realtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/whg517/node-pulse/pulse/internal/db"
)

// sweeperRepo adapts db.NodesQuerier to a string-id shape so the sweeper does
// not import uuid directly into its loop logic. GetStaleNodes returns
// db.StaleNode (which carries a uuid.UUID); MarkNodeOfflineByString accepts the
// string form of that id.
type sweeperRepo interface {
	MarkNodeOfflineByString(ctx context.Context, nodeID string) error
	GetStaleNodes(ctx context.Context, timeout time.Duration) ([]db.StaleNode, error)
}

// NodeStatusSweeper periodically marks nodes offline when their last_heartbeat
// exceeds the timeout threshold, and broadcasts node:offline events so the
// frontend updates in real time. Beacons do not report their own offline
// transitions, so without this sweeper a crashed/disconnected node would stay
// 'online' until a human-driven page refresh re-derived the status.
type NodeStatusSweeper struct {
	hub         *Hub
	repo        sweeperRepo
	interval    time.Duration // how often to scan
	timeout     time.Duration // heartbeat staleness threshold
	logger      *slog.Logger
	cancel      context.CancelFunc
	done        chan struct{}
}

// SweeperOption configures a NodeStatusSweeper.
type SweeperOption func(*NodeStatusSweeper)

// WithSweeperInterval overrides the default 60s scan interval.
func WithSweeperInterval(d time.Duration) SweeperOption {
	return func(s *NodeStatusSweeper) { s.interval = d }
}

// WithSweeperLogger overrides the default logger.
func WithSweeperLogger(l *slog.Logger) SweeperOption {
	return func(s *NodeStatusSweeper) { s.logger = l }
}

// sweeperAdapter wraps a db.NodesQuerier into the string-id sweeperRepo shape,
// converting the uuid.UUID returned by GetStaleNodes to a plain string for the
// offline-marking call.
type sweeperAdapter struct {
	q db.NodesQuerier
}

func (a sweeperAdapter) GetStaleNodes(ctx context.Context, timeout time.Duration) ([]db.StaleNode, error) {
	return a.q.GetStaleNodes(ctx, timeout)
}

func (a sweeperAdapter) MarkNodeOfflineByString(ctx context.Context, nodeID string) error {
	id, err := uuid.Parse(nodeID)
	if err != nil {
		return err
	}
	return a.q.MarkNodeOffline(ctx, id)
}

// NewNodeStatusSweeper constructs a sweeper tied to the given hub and node repo.
// timeout is the heartbeat staleness threshold (typically db.HeartbeatTimeout).
func NewNodeStatusSweeper(hub *Hub, nodeRepo db.NodesQuerier, timeout time.Duration, opts ...SweeperOption) *NodeStatusSweeper {
	s := &NodeStatusSweeper{
		hub:      hub,
		repo:     sweeperAdapter{q: nodeRepo},
		interval: 60 * time.Second,
		timeout:  timeout,
		logger:   slog.Default(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start launches the scan loop in a goroutine. It returns immediately. The
// first scan runs after one interval (not immediately) to avoid racing with
// server startup.
func (s *NodeStatusSweeper) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})
	go s.loop(ctx)
}

// Stop signals the loop to exit and blocks until it has, or until the timeout
// elapses (graceful shutdown).
func (s *NodeStatusSweeper) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.done != nil {
		<-s.done
	}
}

func (s *NodeStatusSweeper) loop(ctx context.Context) {
	defer close(s.done)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scanOnce(ctx)
		}
	}
}

// scanOnce queries stale nodes, marks each offline, and broadcasts node:offline.
// Errors are logged but do not stop the loop; a transient DB failure just means
// we retry next tick.
func (s *NodeStatusSweeper) scanOnce(ctx context.Context) {
	stale, err := s.repo.GetStaleNodes(ctx, s.timeout)
	if err != nil {
		s.logger.Warn("node status sweeper: query stale nodes failed", "error", err)
		return
	}
	for _, n := range stale {
		if err := s.repo.MarkNodeOfflineByString(ctx, n.ID.String()); err != nil {
			s.logger.Warn("node status sweeper: mark offline failed",
				"node_id", n.ID, "name", n.Name, "error", err)
			continue
		}
		s.logger.Info("node status sweeper: marked offline",
			"node_id", n.ID, "name", n.Name, "prev_status", n.Status)
		s.hub.BroadcastNodeStatus(EventNodeOffline, n.ID.String(), "offline")
	}
}

package suppression

import (
	"context"
	"log/slog"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/db"
)

const (
	// DefaultSuppressionWindow is the default suppression window duration
	DefaultSuppressionWindow = 5 * time.Minute
)

// Service handles alert suppression logic
type Service struct {
	querier db.AlertSuppressionsQuerier
}

// NewService creates a new SuppressionService
func NewService(querier db.AlertSuppressionsQuerier) *Service {
	return &Service{
		querier: querier,
	}
}

// ShouldSuppress checks if an alert should be suppressed for a node and metric
// Returns true if the alert should be suppressed (within suppression window)
// Returns false if the alert should not be suppressed (no suppression or window expired)
func (s *Service) ShouldSuppress(ctx context.Context, nodeID string, metric string) (bool, error) {
	suppression, err := s.querier.CheckSuppression(ctx, nodeID, metric)
	if err != nil {
		if err == db.ErrSuppressionNotFound {
			// No suppression record exists, don't suppress
			return false, nil
		}
		// Database error, don't suppress to avoid missing alerts
		slog.Error("Failed to check suppression", "node_id", nodeID, "metric", metric, "error", err)
		return false, err
	}

	// Check if still within suppression window
	if time.Now().Before(suppression.SuppressedUntil) {
		slog.Debug("Alert is within suppression window",
			"node_id", nodeID,
			"metric", metric,
			"suppressed_until", suppression.SuppressedUntil.Format(time.RFC3339))
		return true, nil
	}

	// Suppression window has expired
	slog.Debug("Suppression window has expired",
		"node_id", nodeID,
		"metric", metric,
		"suppressed_until", suppression.SuppressedUntil.Format(time.RFC3339))
	return false, nil
}

// RecordSuppression creates or updates a suppression record for a node and metric
func (s *Service) RecordSuppression(ctx context.Context, nodeID string, metric string, window time.Duration) error {
	suppressedUntil := time.Now().Add(window)

	err := s.querier.CreateOrUpdateSuppression(ctx, nodeID, metric, suppressedUntil)
	if err != nil {
		slog.Error("Failed to record suppression",
			"node_id", nodeID,
			"metric", metric,
			"suppressed_until", suppressedUntil.Format(time.RFC3339),
			"error", err)
		return err
	}

	slog.Debug("Suppression recorded",
		"node_id", nodeID,
		"metric", metric,
		"window", window.String(),
		"suppressed_until", suppressedUntil.Format(time.RFC3339))

	return nil
}

// RecordDefaultSuppression records a suppression with the default window (5 minutes)
func (s *Service) RecordDefaultSuppression(ctx context.Context, nodeID string, metric string) error {
	return s.RecordSuppression(ctx, nodeID, metric, DefaultSuppressionWindow)
}

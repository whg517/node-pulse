package suppression

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

// MockAlertSuppressionsQuerier is a mock for testing
type MockAlertSuppressionsQuerier struct {
	checkFunc               func(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error)
	createOrUpdateFunc      func(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error
	deleteExpiredFunc        func(ctx context.Context) (int64, error)
}

func (m *MockAlertSuppressionsQuerier) CheckSuppression(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error) {
	if m.checkFunc != nil {
		return m.checkFunc(ctx, nodeID, metric)
	}
	return nil, db.ErrSuppressionNotFound
}

func (m *MockAlertSuppressionsQuerier) CreateOrUpdateSuppression(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error {
	if m.createOrUpdateFunc != nil {
		return m.createOrUpdateFunc(ctx, nodeID, metric, suppressedUntil)
	}
	return nil
}

func (m *MockAlertSuppressionsQuerier) DeleteExpiredSuppressions(ctx context.Context) (int64, error) {
	if m.deleteExpiredFunc != nil {
		return m.deleteExpiredFunc(ctx)
	}
	return 0, nil
}

func TestService_ShouldSuppress(t *testing.T) {
	t.Run("No suppression record exists", func(t *testing.T) {
		mockQuerier := &MockAlertSuppressionsQuerier{
			checkFunc: func(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error) {
				return nil, db.ErrSuppressionNotFound
			},
		}
		service := NewService(mockQuerier)

		suppressed, err := service.ShouldSuppress(context.Background(), "node-1", "latency")
		require.NoError(t, err)
		assert.False(t, suppressed, "Should not suppress when no record exists")
	})

	t.Run("Suppression window still active", func(t *testing.T) {
		suppression := &models.AlertSuppression{
			SuppressedUntil: time.Now().Add(5 * time.Minute),
		}
		mockQuerier := &MockAlertSuppressionsQuerier{
			checkFunc: func(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error) {
				return suppression, nil
			},
		}
		service := NewService(mockQuerier)

		suppressed, err := service.ShouldSuppress(context.Background(), "node-1", "latency")
		require.NoError(t, err)
		assert.True(t, suppressed, "Should suppress when window is active")
	})

	t.Run("Suppression window has expired", func(t *testing.T) {
		suppression := &models.AlertSuppression{
			SuppressedUntil: time.Now().Add(-1 * time.Minute), // Expired
		}
		mockQuerier := &MockAlertSuppressionsQuerier{
			checkFunc: func(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error) {
				return suppression, nil
			},
		}
		service := NewService(mockQuerier)

		suppressed, err := service.ShouldSuppress(context.Background(), "node-1", "latency")
		require.NoError(t, err)
		assert.False(t, suppressed, "Should not suppress when window has expired")
	})

	t.Run("Database error returns false (fail open)", func(t *testing.T) {
		mockQuerier := &MockAlertSuppressionsQuerier{
			checkFunc: func(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error) {
				return nil, assert.AnError
			},
		}
		service := NewService(mockQuerier)

		suppressed, err := service.ShouldSuppress(context.Background(), "node-1", "latency")
		assert.Error(t, err, "Should return error from database")
		assert.False(t, suppressed, "Should not suppress on database error (fail open)")
	})
}

func TestService_RecordSuppression(t *testing.T) {
	t.Run("Successfully records suppression", func(t *testing.T) {
		createCalled := false
		mockQuerier := &MockAlertSuppressionsQuerier{
			createOrUpdateFunc: func(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error {
				createCalled = true
				assert.Equal(t, "node-1", nodeID)
				assert.Equal(t, "latency", metric)
				assert.True(t, time.Now().Add(5*time.Minute).Sub(suppressedUntil) < time.Second, "suppressed_until should be ~5 minutes from now")
				return nil
			},
		}
		service := NewService(mockQuerier)

		err := service.RecordSuppression(context.Background(), "node-1", "latency", 5*time.Minute)
		require.NoError(t, err)
		assert.True(t, createCalled, "CreateOrUpdate should be called")
	})

	t.Run("Database error on record suppression", func(t *testing.T) {
		mockQuerier := &MockAlertSuppressionsQuerier{
			createOrUpdateFunc: func(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error {
				return assert.AnError
			},
		}
		service := NewService(mockQuerier)

		err := service.RecordSuppression(context.Background(), "node-1", "latency", 5*time.Minute)
		assert.Error(t, err, "Should return error from database")
	})
}

func TestService_RecordDefaultSuppression(t *testing.T) {
	t.Run("Records suppression with default window", func(t *testing.T) {
		windowUsed := 0 * time.Second
		mockQuerier := &MockAlertSuppressionsQuerier{
			createOrUpdateFunc: func(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error {
				windowUsed = 5 * time.Minute
				return nil
			},
		}
		service := NewService(mockQuerier)

		err := service.RecordDefaultSuppression(context.Background(), "node-1", "latency")
		require.NoError(t, err)
		assert.Equal(t, 5*time.Minute, windowUsed, "Should use default 5 minute window")
	})
}

package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

func TestAlertSystemChecker_CheckAlertEngine(t *testing.T) {
	t.Run("Healthy alert engine", func(t *testing.T) {
		mockEngine := &mockAlertEngineStats{
			stats: map[string]interface{}{
				"cached_rules":            15,
				"rule_cache_last_refresh": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
				"metric_channel_depth":    23,
				"metric_channel_capacity": 1000,
			},
		}

		checker := NewAlertSystemChecker(mockEngine, nil, nil)
		status, err := checker.CheckAlertEngine(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "ok", status.Status)
		assert.Equal(t, 15, status.CachedRules)
		assert.Equal(t, 23, status.MetricChannelDepth)
		assert.Equal(t, 1000, status.MetricChannelCapacity)
	})

	t.Run("Stale rule cache", func(t *testing.T) {
		mockEngine := &mockAlertEngineStats{
			stats: map[string]interface{}{
				"cached_rules":            15,
				"rule_cache_last_refresh": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
				"metric_channel_depth":    23,
				"metric_channel_capacity": 1000,
			},
		}

		checker := NewAlertSystemChecker(mockEngine, nil, nil)
		status, err := checker.CheckAlertEngine(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "stale", status.Status)
	})

	t.Run("Full metric channel", func(t *testing.T) {
		mockEngine := &mockAlertEngineStats{
			stats: map[string]interface{}{
				"cached_rules":            15,
				"rule_cache_last_refresh": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
				"metric_channel_depth":    1000,
				"metric_channel_capacity": 1000,
			},
		}

		checker := NewAlertSystemChecker(mockEngine, nil, nil)
		status, err := checker.CheckAlertEngine(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "full", status.Status)
	})
}

func TestAlertSystemChecker_CheckWebhookDelivery(t *testing.T) {
	t.Run("Healthy delivery rate", func(t *testing.T) {
		mockQuerier := &mockWebhookLogsQuerier{
			totalCount:   100,
			successCount: 98,
		}

		checker := NewAlertSystemChecker(nil, mockQuerier, nil)
		status, err := checker.CheckWebhookDelivery(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "healthy", status.Status)
		assert.InDelta(t, 98.0, status.SuccessRate, 0.1)
		assert.Equal(t, 100, status.TotalCount)
		assert.Equal(t, 98, status.SuccessCount)
	})

	t.Run("Degraded delivery rate", func(t *testing.T) {
		mockQuerier := &mockWebhookLogsQuerier{
			totalCount:   100,
			successCount: 85,
		}

		checker := NewAlertSystemChecker(nil, mockQuerier, nil)
		status, err := checker.CheckWebhookDelivery(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "degraded", status.Status)
		assert.InDelta(t, 85.0, status.SuccessRate, 0.1)
	})

	t.Run("Unhealthy delivery rate", func(t *testing.T) {
		mockQuerier := &mockWebhookLogsQuerier{
			totalCount:   100,
			successCount: 75,
		}

		checker := NewAlertSystemChecker(nil, mockQuerier, nil)
		status, err := checker.CheckWebhookDelivery(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "unhealthy", status.Status)
		assert.InDelta(t, 75.0, status.SuccessRate, 0.1)
	})

	t.Run("No webhook logs", func(t *testing.T) {
		mockQuerier := &mockWebhookLogsQuerier{
			totalCount:   0,
			successCount: 0,
		}

		checker := NewAlertSystemChecker(nil, mockQuerier, nil)
		status, err := checker.CheckWebhookDelivery(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "nodata", status.Status)
		assert.Equal(t, 0, status.TotalCount)
	})
}

func TestAlertSystemChecker_CheckAlertSuppression(t *testing.T) {
	t.Run("Active suppressions", func(t *testing.T) {
		mockQuerier := &mockAlertSuppressionsQuerier{
			activeCount: 5,
		}

		checker := NewAlertSystemChecker(nil, nil, mockQuerier)
		status, err := checker.CheckAlertSuppression(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "ok", status.Status)
		assert.Equal(t, int64(5), status.ActiveSuppressionCount)
	})

	t.Run("No active suppressions", func(t *testing.T) {
		mockQuerier := &mockAlertSuppressionsQuerier{
			activeCount: 0,
		}

		checker := NewAlertSystemChecker(nil, nil, mockQuerier)
		status, err := checker.CheckAlertSuppression(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "ok", status.Status)
		assert.Equal(t, int64(0), status.ActiveSuppressionCount)
	})
}

// Mock implementations

type mockAlertEngineStats struct {
	stats map[string]interface{}
}

func (m *mockAlertEngineStats) GetStats() map[string]interface{} {
	return m.stats
}

type mockWebhookLogsQuerier struct {
	totalCount   int64
	successCount int64
}

func (m *mockWebhookLogsQuerier) CountRecentWebhookLogs(ctx context.Context, totalCount, successCount *int64, limit int) error {
	*totalCount = m.totalCount
	*successCount = m.successCount
	return nil
}

func (m *mockWebhookLogsQuerier) CreateWebhookLog(ctx context.Context, log *models.WebhookLog) error {
	return nil
}

type mockAlertSuppressionsQuerier struct {
	activeCount int64
}

func (m *mockAlertSuppressionsQuerier) CheckSuppression(ctx context.Context, nodeID string, metric string) (*models.AlertSuppression, error) {
	return nil, nil
}

func (m *mockAlertSuppressionsQuerier) CreateOrUpdateSuppression(ctx context.Context, nodeID string, metric string, suppressedUntil time.Time) error {
	return nil
}

func (m *mockAlertSuppressionsQuerier) DeleteExpiredSuppressions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockAlertSuppressionsQuerier) CountActiveSuppressions(ctx context.Context) (int64, error) {
	return m.activeCount, nil
}

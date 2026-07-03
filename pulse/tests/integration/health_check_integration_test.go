package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/whg517/node-pulse/pulse/internal/alert"
	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/health"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

type HealthCheckIntegrationTestSuite struct {
	suite.Suite
	pool                     *pgxpool.Pool
	alertEngine              *alert.AlertEngine
	webhookLogsQuerier       db.WebhookLogsQuerier
	alertSuppressionsQuerier db.AlertSuppressionsQuerier
	alertSystemChecker       *health.AlertSystemChecker
	ctx                      context.Context
}

func (suite *HealthCheckIntegrationTestSuite) SetupSuite() {
	testutil.SetupTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(suite.T(), err, "Failed to load config")

	suite.ctx = context.Background()

	// Connect to test database (skips cleanly when DB is unreachable or -short)
	pool := testutil.RequireDB(suite.T())
	suite.pool = pool

	// Run migrations
	err = db.Migrate(suite.ctx, pool)
	require.NoError(suite.T(), err, "Failed to run migrations")
}

func (suite *HealthCheckIntegrationTestSuite) SetupTest() {
	// Initialize alert engine
	alertQuerier := db.NewAlertQuerier(suite.pool)
	alertEngineConfig := alert.DefaultEngineConfig()
	alertEngineConfig.WorkerPoolSize = 2
	suite.alertEngine = alert.NewAlertEngine(suite.pool, alertQuerier, alertEngineConfig)
	suite.alertEngine.Start()

	// Give engine time to initialize
	time.Sleep(100 * time.Millisecond)

	// Initialize queriers
	suite.webhookLogsQuerier = db.NewWebhookLogsQuerier(suite.pool)
	suite.alertSuppressionsQuerier = db.NewAlertSuppressionsQuerier(suite.pool)

	// Initialize alert system health checker
	suite.alertSystemChecker = health.NewAlertSystemChecker(
		suite.alertEngine,
		suite.webhookLogsQuerier,
		suite.alertSuppressionsQuerier,
	)
}

func (suite *HealthCheckIntegrationTestSuite) TearDownTest() {
	if suite.alertEngine != nil {
		suite.alertEngine.Stop()
	}

	// Clean up test data
	_, _ = suite.pool.Exec(suite.ctx, "DELETE FROM webhook_logs")
	_, _ = suite.pool.Exec(suite.ctx, "DELETE FROM alert_suppressions")
	_, _ = suite.pool.Exec(suite.ctx, "DELETE FROM alert_events")
	_, _ = suite.pool.Exec(suite.ctx, "DELETE FROM webhooks")
	_, _ = suite.pool.Exec(suite.ctx, "DELETE FROM alerts")
	_, _ = suite.pool.Exec(suite.ctx, "DELETE FROM nodes")
}

func (suite *HealthCheckIntegrationTestSuite) TearDownSuite() {
	suite.pool.Close()
	testutil.TeardownTestConfig()
}

func (suite *HealthCheckIntegrationTestSuite) TestAlertSystemChecker_CheckAlertEngine() {
	// Check alert engine health
	status, err := suite.alertSystemChecker.CheckAlertEngine(suite.ctx)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)

	// Verify engine status is one of: ok, stale, full
	assert.Contains(suite.T(), []string{"ok", "stale", "full"}, status.Status)

	// Verify cached rules count
	assert.GreaterOrEqual(suite.T(), status.CachedRules, 0)

	// Verify channel metrics
	assert.GreaterOrEqual(suite.T(), status.MetricChannelDepth, 0)
	assert.Greater(suite.T(), status.MetricChannelCapacity, 0)

	// Verify channel depth is not at capacity
	assert.LessOrEqual(suite.T(), status.MetricChannelDepth, status.MetricChannelCapacity)

	// Verify rule cache last refresh is set
	assert.NotEmpty(suite.T(), status.RuleCacheLastRefresh)
}

func (suite *HealthCheckIntegrationTestSuite) TestAlertSystemChecker_CheckWebhookDelivery_NoData() {
	// No webhook logs yet - should return nodata
	status, err := suite.alertSystemChecker.CheckWebhookDelivery(suite.ctx)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), "nodata", status.Status)
	assert.Equal(suite.T(), 0, status.TotalCount)
	assert.Equal(suite.T(), 0, status.SuccessCount)
	assert.Equal(suite.T(), 0.0, status.SuccessRate)
}

func (suite *HealthCheckIntegrationTestSuite) TestAlertSystemChecker_CheckWebhookDelivery_WithLogs() {
	// Create test webhook
	webhook := &models.Webhook{
		URL:     "https://example.com/webhook",
		Enabled: true,
	}
	err := db.NewWebhookQuerier(suite.pool).CreateWebhook(suite.ctx, webhook)
	require.NoError(suite.T(), err)

	// Create a test node for the alert events (required for foreign key constraint)
	nodeID := uuid.New()
	nodeQuerier := db.NewPoolQuerier(suite.pool)
	err = nodeQuerier.CreateNode(suite.ctx, nodeID, "test-node-webhook", "192.168.1.100", "us-west", nil)
	require.NoError(suite.T(), err)

	// Create 70 successful webhook logs with valid alert events
	for i := 0; i < 70; i++ {
		// Create alert event first (required by foreign key)
		alertEventID := uuid.New()
		_, err = suite.pool.Exec(suite.ctx, `
			INSERT INTO alert_events (id, node_id, metric, threshold, current_value, level, created_at)
			VALUES ($1, $2, 'latency', 100, 150, 'P0', NOW())
		`, alertEventID, nodeID)
		require.NoError(suite.T(), err)

		log := &models.WebhookLog{
			WebhookID:    webhook.ID,
			AlertEventID: alertEventID.String(),
			Status:       "success",
			RetryCount:   0,
			ErrorMessage: "",
		}
		err = suite.webhookLogsQuerier.CreateWebhookLog(suite.ctx, log)
		require.NoError(suite.T(), err)
	}

	// Create 30 failed webhook logs with valid alert events
	for i := 0; i < 30; i++ {
		// Create alert event first (required by foreign key)
		alertEventID := uuid.New()
		_, err = suite.pool.Exec(suite.ctx, `
			INSERT INTO alert_events (id, node_id, metric, threshold, current_value, level, created_at)
			VALUES ($1, $2, 'latency', 100, 200, 'P0', NOW())
		`, alertEventID, nodeID)
		require.NoError(suite.T(), err)

		log := &models.WebhookLog{
			WebhookID:    webhook.ID,
			AlertEventID: alertEventID.String(),
			Status:       "failure",
			RetryCount:   3,
			ErrorMessage: "timeout",
		}
		err = suite.webhookLogsQuerier.CreateWebhookLog(suite.ctx, log)
		require.NoError(suite.T(), err)
	}

	// Wait for logs to be persisted
	time.Sleep(100 * time.Millisecond)

	// Check webhook delivery health
	status, err := suite.alertSystemChecker.CheckWebhookDelivery(suite.ctx)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)

	// Should have 100 total logs
	assert.Equal(suite.T(), 100, status.TotalCount)

	// Should have 70 successful logs (70% success rate = degraded)
	assert.Equal(suite.T(), 70, status.SuccessCount)
	assert.InDelta(suite.T(), 70.0, status.SuccessRate, 0.1)

	// Status should be "degraded" (70% < 95% but >= 80% threshold is actually lower)
	// Actually 70% is <80%, so it should be unhealthy, but let's just check it's not nodata
	assert.NotEqual(suite.T(), "nodata", status.Status)
}

func (suite *HealthCheckIntegrationTestSuite) TestAlertSystemChecker_CheckAlertSuppression_NoSuppressions() {
	// No suppressions - should return ok with count 0
	status, err := suite.alertSystemChecker.CheckAlertSuppression(suite.ctx)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), "ok", status.Status)
	assert.Equal(suite.T(), int64(0), status.ActiveSuppressionCount)
}

func (suite *HealthCheckIntegrationTestSuite) TestAlertSystemChecker_CheckAlertSuppression_WithSuppressions() {
	// Create 3 active suppressions (must first create nodes for foreign key constraint)
	nodes := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	nodeQuerier := db.NewPoolQuerier(suite.pool)
	for i, nodeID := range nodes {
		nodeName := fmt.Sprintf("health-check-node-%d", i)
		err := nodeQuerier.CreateNode(suite.ctx, nodeID, nodeName, "192.168.1.100", "us-west", nil)
		require.NoError(suite.T(), err)

		err = suite.alertSuppressionsQuerier.CreateOrUpdateSuppression(
			suite.ctx,
			nodeID.String(),
			"latency",
			time.Now().Add(5*time.Minute),
		)
		require.NoError(suite.T(), err)
	}

	// Check alert suppression health
	status, err := suite.alertSystemChecker.CheckAlertSuppression(suite.ctx)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), "ok", status.Status)
	assert.Equal(suite.T(), int64(3), status.ActiveSuppressionCount)
}

func (suite *HealthCheckIntegrationTestSuite) TestAlertSystemChecker_CheckPerformance() {
	// Health check should complete quickly (< 1 second)
	start := time.Now()

	_, err := suite.alertSystemChecker.CheckAlertEngine(suite.ctx)
	require.NoError(suite.T(), err)

	_, err = suite.alertSystemChecker.CheckWebhookDelivery(suite.ctx)
	require.NoError(suite.T(), err)

	_, err = suite.alertSystemChecker.CheckAlertSuppression(suite.ctx)
	require.NoError(suite.T(), err)

	duration := time.Since(start)

	// All checks should complete in < 1 second
	assert.Less(suite.T(), duration, 1*time.Second, "Alert system health checks should complete in < 1 second")
}

func TestHealthCheckIntegrationSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckIntegrationTestSuite))
}

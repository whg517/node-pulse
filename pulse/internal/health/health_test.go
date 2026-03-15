package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/scheduler"
)

// --- Mocks ---

type mockDBChecker struct {
	err error
}

func (m *mockDBChecker) Check(_ context.Context) error {
	return m.err
}

type mockScheduler struct {
	status *scheduler.TaskStatus
	err    error
}

func (m *mockScheduler) Start(_ context.Context) error     { return nil }
func (m *mockScheduler) Stop() error                       { return nil }
func (m *mockScheduler) RegisterTask(_ scheduler.Task) error { return nil }
func (m *mockScheduler) GetTaskStatus(_ string) (*scheduler.TaskStatus, error) {
	return m.status, m.err
}

// --- Tests ---

func TestNew(t *testing.T) {
	db := &mockDBChecker{}
	sched := &mockScheduler{}
	checker := NewAlertSystemChecker(nil, nil, nil)

	hc := New(db, sched, checker)
	assert.NotNil(t, hc)
	assert.Equal(t, db, hc.db)
	assert.Equal(t, sched, hc.scheduler)
	assert.Equal(t, checker, hc.alertSystemChecker)
}

func TestNew_NilDependencies(t *testing.T) {
	hc := New(nil, nil, nil)
	assert.NotNil(t, hc)
	assert.Nil(t, hc.db)
	assert.Nil(t, hc.scheduler)
	assert.Nil(t, hc.alertSystemChecker)
}

func TestHandler_DatabaseDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hc := New(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "disabled", resp.Checks["database"])
}

func TestHandler_DatabaseHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hc := New(&mockDBChecker{err: nil}, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "ok", resp.Checks["database"])
}

func TestHandler_DatabaseUnhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hc := New(&mockDBChecker{err: errors.New("connection refused")}, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "unhealthy", resp.Status)
}

func TestHandler_SchedulerWithTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	taskStatus := &scheduler.TaskStatus{
		Name:     "metrics-cleanup",
		IsRunning: false,
		LastRun:  time.Now().Add(-5 * time.Minute),
		RunCount: 10,
	}
	sched := &mockScheduler{status: taskStatus, err: nil}
	hc := New(nil, sched, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(t, resp.Scheduler)
	assert.True(t, resp.Scheduler.Running)
	assert.Contains(t, resp.Scheduler.Tasks, "metrics-cleanup")
}

func TestHandler_SchedulerNoTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sched := &mockScheduler{status: nil, err: errors.New("task not found")}
	hc := New(nil, sched, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(t, resp.Scheduler)
	assert.True(t, resp.Scheduler.Running)
	assert.Empty(t, resp.Scheduler.Tasks)
}

func TestHandler_AlertSystemHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockEngine := &mockAlertEngineStats{
		stats: map[string]interface{}{
			"cached_rules":            5,
			"rule_cache_last_refresh": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			"metric_channel_depth":    10,
			"metric_channel_capacity": 1000,
		},
	}
	mockWebhookQuerier := &mockWebhookLogsQuerier{totalCount: 100, successCount: 99}
	mockSuppQuerier := &mockAlertSuppressionsQuerier{activeCount: 2}

	alertChecker := NewAlertSystemChecker(mockEngine, mockWebhookQuerier, mockSuppQuerier)
	hc := New(nil, nil, alertChecker)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(t, resp.AlertSystem)
	assert.Equal(t, "ok", resp.Checks["alert_engine"])
}

func TestHandler_AlertSystemDegraded_FullChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Full metric channel → "full" → unhealthy
	mockEngine := &mockAlertEngineStats{
		stats: map[string]interface{}{
			"cached_rules":            5,
			"rule_cache_last_refresh": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			"metric_channel_depth":    1000,
			"metric_channel_capacity": 1000,
		},
	}
	mockWebhookQuerier := &mockWebhookLogsQuerier{totalCount: 100, successCount: 99}
	mockSuppQuerier := &mockAlertSuppressionsQuerier{activeCount: 0}

	alertChecker := NewAlertSystemChecker(mockEngine, mockWebhookQuerier, mockSuppQuerier)
	hc := New(nil, nil, alertChecker)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	// Full channel makes status "full" → isHealthy=false → http 503
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandler_AlertSystemDegraded_LowSuccessRate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockEngine := &mockAlertEngineStats{
		stats: map[string]interface{}{
			"cached_rules":            5,
			"rule_cache_last_refresh": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			"metric_channel_depth":    0,
			"metric_channel_capacity": 1000,
		},
	}
	// Low success rate (70%) → "unhealthy" webhook delivery → isDegraded=true
	mockWebhookQuerier := &mockWebhookLogsQuerier{totalCount: 100, successCount: 70}
	mockSuppQuerier := &mockAlertSuppressionsQuerier{activeCount: 0}

	alertChecker := NewAlertSystemChecker(mockEngine, mockWebhookQuerier, mockSuppQuerier)
	hc := New(nil, nil, alertChecker)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	hc.Handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "degraded", resp.Status)
}

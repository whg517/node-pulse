package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/pkg/metrics"
)

func TestPerformanceHandler_GetPerformanceData_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create metrics collector
	collector := metrics.NewCollector()
	collector.Start()
	defer collector.Stop()

	handler := NewMetricsHandler(collector)

	router := gin.New()
	router.GET("/performance", handler.GetPerformanceData)

	// Test request - even with no data, should return empty results
	req, _ := http.NewRequest("GET", "/performance?time_range=24h", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, resp, "data")
	assert.Contains(t, resp, "message")
	assert.Contains(t, resp, "timestamp")

	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "metrics")
	assert.Contains(t, data, "trend_data")
	assert.Contains(t, data, "system_health")
	assert.Contains(t, data, "anomalies")
	assert.Contains(t, data, "summary")

	// Verify metrics exist (should have at least the 3 defined target metrics)
	metricsList := data["metrics"].([]interface{})
	assert.GreaterOrEqual(t, len(metricsList), 3)
}

func TestPerformanceHandler_GetPerformanceData_InvalidTimeRange(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	collector := metrics.NewCollector()
	collector.Start()
	defer collector.Stop()

	handler := NewMetricsHandler(collector)

	router := gin.New()
	router.GET("/performance", handler.GetPerformanceData)

	// Test invalid time range
	tests := []struct {
		name       string
		timeRange  string
		wantStatus int
		wantError  string
	}{
		{
			name:       "Invalid format",
			timeRange:  "invalid",
			wantStatus: http.StatusBadRequest,
			wantError:  "INVALID_TIME_RANGE",
		},
		{
			name:       "Empty time range uses default",
			timeRange:  "",
			wantStatus: http.StatusOK, // Empty uses default 24h
			wantError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/performance?time_range="+tt.timeRange, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tt.wantError != "" {
				assert.Equal(t, tt.wantStatus, w.Code)

				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp, "code")
				assert.Equal(t, tt.wantError, resp["code"])
			} else {
				assert.Equal(t, tt.wantStatus, w.Code)
			}
		})
	}
}

func TestPerformanceHandler_GetPerformanceData_SystemHealthCalculation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	collector := metrics.NewCollector()
	collector.Start()
	defer collector.Stop()

	handler := NewMetricsHandler(collector)

	router := gin.New()
	router.GET("/performance", handler.GetPerformanceData)

	// Test request with no data
	req, _ := http.NewRequest("GET", "/performance?time_range=24h", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert - with no data, system should be healthy
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	systemHealth := data["system_health"].(string)

	// Should be healthy with no data
	assert.Equal(t, "healthy", systemHealth)
}

func TestParseTimeRange(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantDur time.Duration
		wantErr bool
	}{
		{
			name:    "24 hours",
			input:   "24h",
			wantDur: 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "7 days",
			input:   "7d",
			wantDur: 7 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "1 hour",
			input:   "1h",
			wantDur: 1 * time.Hour,
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "invalid",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "Too short",
			input:   "h",
			wantDur: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeRange(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantDur, got)
			}
		})
	}
}

func TestEvaluateSystemHealth_WithAnomalies(t *testing.T) {
	// Test severity calculation based on PRD thresholds
	tests := []struct {
		name              string
		metrics           []models.PerformanceMetric
		wantHealth        string
		wantAnomalyCount  int
		wantCriticalCount int
	}{
		{
			name: "No data - all zeros",
			metrics: []models.PerformanceMetric{
				{
					MetricName: "dashboard_load_time",
					CurrentP99: 0,
					CurrentP95: 0,
					TargetP99:  3000,
					TargetP95:  2000,
					Status:     "healthy",
				},
			},
			wantHealth:        "healthy",
			wantAnomalyCount:  0,
			wantCriticalCount: 0,
		},
		{
			name: "Warning level anomaly - exceeds target but below P0",
			metrics: []models.PerformanceMetric{
				{
					MetricName:  "dashboard_load_time",
					DisplayName: "仪表盘加载时间",
					CurrentP99:  3500,
					CurrentP95:  2200,
					TargetP99:   3000,
					TargetP95:   2000,
					Status:      "unhealthy",
					Anomaly:     "P99 超过目标值",
				},
			},
			wantHealth:        "unhealthy",
			wantAnomalyCount:  1,
			wantCriticalCount: 0, // 3500ms < 5000ms threshold
		},
		{
			name: "Critical level anomaly - exceeds P0 threshold",
			metrics: []models.PerformanceMetric{
				{
					MetricName:  "dashboard_load_time",
					DisplayName: "仪表盘加载时间",
					CurrentP99:  5500,
					CurrentP95:  3000,
					TargetP99:   3000,
					TargetP95:   2000,
					Status:      "unhealthy",
					Anomaly:     "P99 超过目标值",
				},
			},
			wantHealth:        "unhealthy",
			wantAnomalyCount:  1,
			wantCriticalCount: 1, // 5500ms > 5000ms threshold
		},
		{
			name: "API response time critical",
			metrics: []models.PerformanceMetric{
				{
					MetricName:  "api_response_time",
					DisplayName: "API 响应时间",
					CurrentP99:  1200,
					CurrentP95:  250,
					TargetP99:   500,
					TargetP95:   200,
					Status:      "unhealthy",
					Anomaly:     "P99 超过目标值",
				},
			},
			wantHealth:        "unhealthy",
			wantAnomalyCount:  1,
			wantCriticalCount: 1, // 1200ms > 1000ms threshold
		},
		{
			name: "Data query time critical",
			metrics: []models.PerformanceMetric{
				{
					MetricName:  "data_query_time",
					DisplayName: "数据查询时间",
					CurrentP99:  700,
					CurrentP95:  250,
					TargetP99:   300,
					TargetP95:   200,
					Status:      "unhealthy",
					Anomaly:     "P99 超过目标值",
				},
			},
			wantHealth:        "unhealthy",
			wantAnomalyCount:  1,
			wantCriticalCount: 1, // 700ms > 600ms threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health, anomalies := evaluateSystemHealth(tt.metrics)

			assert.Equal(t, tt.wantHealth, health)
			assert.Equal(t, tt.wantAnomalyCount, len(anomalies))

			criticalCount := 0
			for _, a := range anomalies {
				if a.Severity == "P0" {
					criticalCount++
				}
			}
			assert.Equal(t, tt.wantCriticalCount, criticalCount)
		})
	}
}

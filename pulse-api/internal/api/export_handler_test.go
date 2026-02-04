package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/export"
	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
)

func TestExportHandler_CreateExportHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	pool := setupTestDB(t)

	exportService := export.NewExportService(pool)
	handler := NewExportHandler(exportService)

	router := gin.New()
	router.POST("/export", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		handler.CreateExportHandler(c)
	})

	// Create test data
	ctx := context.Background()
	nodeID := createTestNode(ctx, t, pool)
	createTestMetrics(ctx, t, pool, nodeID)

	// Test request
	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()

	// Build URL with URL-encoded timestamps
	queryParams := url.Values{}
	queryParams.Set("node_ids", nodeID)
	queryParams.Set("start_time", startTime.Format(time.RFC3339))
	queryParams.Set("end_time", endTime.Format(time.RFC3339))
	queryParams.Set("metrics", "latency")
	queryParams.Set("format", "csv")

	req, _ := http.NewRequest("POST", "/export?"+queryParams.Encode(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp CreateExportResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Export task created successfully", resp.Message)
	assert.NotEmpty(t, resp.Data.ID)
	assert.Equal(t, "pending", resp.Data.Status)
	assert.Equal(t, "csv", resp.Data.Format)
	assert.Equal(t, []string{nodeID}, resp.Data.NodeIDs)
	assert.Equal(t, []string{"latency"}, resp.Data.Metrics)
}

func TestExportHandler_CreateExportHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantError  string
	}{
		{
			name:       "Missing node_ids",
			query:      "?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z&metrics=latency",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request parameters",
		},
		{
			name:       "Missing start_time",
			query:      "?node_ids=node1&end_time=2024-01-02T00:00:00Z&metrics=latency",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request parameters",
		},
		{
			name:       "Missing end_time",
			query:      "?node_ids=node1&start_time=2024-01-01T00:00:00Z&metrics=latency",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request parameters",
		},
		{
			name:       "Missing metrics",
			query:      "?node_ids=node1&start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request parameters",
		},
		{
			name:       "Invalid metric",
			query:      "?node_ids=node1&start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z&metrics=invalid_metric",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid metric",
		},
		{
			name:       "Invalid format",
			query:      "?node_ids=node1&start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z&metrics=latency&format=xlsx",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid format",
		},
		{
			name:       "Invalid time range (end before start)",
			query:      "?node_ids=node1&start_time=2024-01-02T00:00:00Z&end_time=2024-01-01T00:00:00Z&metrics=latency",
			wantStatus: http.StatusBadRequest,
			wantError:  "Failed to create export task",
		},
		{
			name:       "Invalid time format",
			query:      "?node_ids=node1&start_time=invalid&end_time=2024-01-02T00:00:00Z&metrics=latency",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid start_time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			testutil.SetupTestConfig()
			defer testutil.TeardownTestConfig()

			// Load configuration
			_, err := config.Load()
			require.NoError(t, err, "Failed to load config")

			pool := setupTestDB(t)

			exportService := export.NewExportService(pool)
			handler := NewExportHandler(exportService)

			router := gin.New()
			router.POST("/export", func(c *gin.Context) {
				c.Set("user_id", "test-user-id")
				handler.CreateExportHandler(c)
			})

			// Test request - use POST for POST route
			req, _ := http.NewRequest("POST", "/export"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Contains(t, resp["error"], tt.wantError)
		})
	}
}

func TestExportHandler_CreateExportHandler_MaxNodesExceeded(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	pool := setupTestDB(t)

	exportService := export.NewExportService(pool)
	handler := NewExportHandler(exportService)

	router := gin.New()
	router.POST("/export", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		handler.CreateExportHandler(c)
	})

	// Create 51 node IDs (exceeds max of 50)
	nodeIDs := make([]string, 51)
	for i := 0; i < 51; i++ {
		nodeIDs[i] = fmt.Sprintf("node-%d", i)
	}

	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()

	// Build query with 51 node IDs - use url.Values for proper encoding
	queryParams := url.Values{}
	queryParams.Set("start_time", startTime.Format(time.RFC3339))
	queryParams.Set("end_time", endTime.Format(time.RFC3339))
	queryParams.Set("metrics", "latency")
	for _, nodeID := range nodeIDs {
		queryParams.Add("node_ids", nodeID)
	}

	req, _ := http.NewRequest("POST", "/export?"+queryParams.Encode(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "Invalid request parameters")
	assert.Contains(t, resp["details"], "max")
}

func TestExportHandler_GetExportStatusHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	pool := setupTestDB(t)

	exportService := export.NewExportService(pool)
	handler := NewExportHandler(exportService)

	router := gin.New()
	router.GET("/export/:id", handler.GetExportStatusHandler)

	// Create an export
	ctx := context.Background()
	nodeID := createTestNode(ctx, t, pool)
	createTestMetrics(ctx, t, pool, nodeID)

	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()

	req := &export.CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   []string{nodeID},
		StartTime: startTime,
		EndTime:   endTime,
		Metrics:   []string{"latency"},
		Format:    "csv",
	}

	task, err := exportService.CreateExport(ctx, req)
	require.NoError(t, err)

	// Wait a bit for processing to start
	time.Sleep(100 * time.Millisecond)

	// Test request
	getReq, _ := http.NewRequest("GET", "/export/"+task.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, getReq)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp GetExportStatusResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, task.ID, resp.Data.ID)
	assert.NotEmpty(t, resp.Message)
}

func TestExportHandler_GetExportStatusHandler_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	pool := setupTestDB(t)

	exportService := export.NewExportService(pool)
	handler := NewExportHandler(exportService)

	router := gin.New()
	router.GET("/export/:id", handler.GetExportStatusHandler)

	// Test request with non-existent ID
	req, _ := http.NewRequest("GET", "/export/non-existent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Export not found", resp["error"])
}

func TestExportHandler_DownloadExportHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	pool := setupTestDB(t)

	exportService := export.NewExportService(pool)
	handler := NewExportHandler(exportService)

	router := gin.New()
	router.GET("/export/:id/download", handler.DownloadExportHandler)

	// Create an export
	ctx := context.Background()
	nodeID := createTestNode(ctx, t, pool)
	createTestMetrics(ctx, t, pool, nodeID)

	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()

	req := &export.CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   []string{nodeID},
		StartTime: startTime,
		EndTime:   endTime,
		Metrics:   []string{"latency"},
		Format:    "csv",
	}

	task, err := exportService.CreateExport(ctx, req)
	require.NoError(t, err)

	// Wait for export to complete
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Export did not complete in time")
		case <-ticker.C:
			updatedTask, err := exportService.GetExport(task.ID)
			require.NoError(t, err)
			if updatedTask.IsCompleted() {
				goto Download
			}
			if updatedTask.IsFailed() {
				t.Fatalf("Export failed: %s", updatedTask.Error)
			}
		}
	}

Download:
	// Test request
	downloadReq, _ := http.NewRequest("GET", "/export/"+task.ID+"/download", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, downloadReq)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.NotEmpty(t, w.Body.Bytes())
}

func TestExportHandler_DownloadExportHandler_NotReady(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	pool := setupTestDB(t)

	exportService := export.NewExportService(pool)
	handler := NewExportHandler(exportService)

	router := gin.New()
	router.GET("/export/:id/download", handler.DownloadExportHandler)

	// Create an export
	ctx := context.Background()
	nodeID := createTestNode(ctx, t, pool)
	createTestMetrics(ctx, t, pool, nodeID)

	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()

	req := &export.CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   []string{nodeID},
		StartTime: startTime,
		EndTime:   endTime,
		Metrics:   []string{"latency"},
		Format:    "csv",
	}

	task, err := exportService.CreateExport(ctx, req)
	require.NoError(t, err)

	// Try to download immediately (should not be ready yet)
	downloadReq, _ := http.NewRequest("GET", "/export/"+task.ID+"/download", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, downloadReq)

	// Assert - should get an error about not being ready
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Export not ready", resp["error"])
}

func TestExportHandler_Unauthorized(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()
	defer testutil.TeardownTestConfig()

	// Load configuration
	_, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	pool := setupTestDB(t)

	exportService := export.NewExportService(pool)
	handler := NewExportHandler(exportService)

	router := gin.New()
	router.POST("/export", handler.CreateExportHandler)

	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()

	// Test request without user_id in context - use POST for POST route
	req, _ := http.NewRequest("POST", "/export?node_ids=node1&start_time="+startTime.Format(time.RFC3339)+"&end_time="+endTime.Format(time.RFC3339)+"&metrics=latency", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Unauthorized", resp["error"])
}

// Helper functions

func setupTestDB(t *testing.T) *pgxpool.Pool {
	pool, cleanup := setupTestDBWithCleanup(t)
	t.Cleanup(cleanup)
	return pool
}

func setupTestDBWithCleanup(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Use test database from environment or default
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil, func() {}
	}

	// Clean up any existing tables from previous tests
	pool.Exec(ctx, "DROP TABLE IF EXISTS alerts CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS metrics CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS probes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS nodes CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE")
	pool.Exec(ctx, "DROP TABLE IF EXISTS sessions CASCADE")

	// Create all tables
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

func createTestNode(ctx context.Context, t *testing.T, pool *pgxpool.Pool) string {
	nodeID := uuid.New().String()

	query := `
		INSERT INTO nodes (id, name, ip, region)
		VALUES ($1, $2, $3, $4)
	`
	_, err := pool.Exec(ctx, query, nodeID, "test-node", "10.0.0.1", "us-east")
	require.NoError(t, err)

	probeID := uuid.New().String()
	probeQuery := `
		INSERT INTO probes (id, node_id, type, target, port, interval_seconds, count, timeout_seconds)
		VALUES ($1, $2, 'TCP', $3, $4, $5, $6, $7)
	`
	_, err = pool.Exec(ctx, probeQuery, probeID, nodeID, "example.com", 443, 60, 10, 5)
	require.NoError(t, err)

	return nodeID
}

func createTestMetrics(ctx context.Context, t *testing.T, pool *pgxpool.Pool, nodeID string) {
	// Get probe ID for the node
	var probeID string
	err := pool.QueryRow(ctx, "SELECT id FROM probes WHERE node_id = $1 LIMIT 1", nodeID).Scan(&probeID)
	require.NoError(t, err)

	// Insert test metrics
	query := `
		INSERT INTO metrics (node_id, probe_id, timestamp, latency_ms, packet_loss_rate, jitter_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	for i := 0; i < 10; i++ {
		timestamp := now.Add(-time.Duration(i) * 10 * time.Minute)
		latency := 50.0 + float64(i)
		packetLoss := 0.01 + float64(i)*0.001
		jitter := 2.0 + float64(i)*0.5

		_, err := pool.Exec(ctx, query, nodeID, probeID, timestamp, latency, packetLoss, jitter)
		require.NoError(t, err)
	}
}

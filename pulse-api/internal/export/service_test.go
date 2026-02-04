package export

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/config"
	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
)

func TestExportService_CreateExport_Success(t *testing.T) {
	// Setup
	pool := setupTestDB(t)

	service := NewExportService(pool)
	ctx := context.Background()

	// Create test data
	nodeID := createTestNode(ctx, t, pool)
	createTestMetrics(ctx, t, pool, nodeID)

	// Create export request
	req := &CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   []string{nodeID},
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now(),
		Metrics:   []string{"latency"},
		Format:    "csv",
	}

	// Execute
	task, err := service.CreateExport(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "test-user-id", task.UserID)
	assert.Equal(t, []string{nodeID}, task.NodeIDs)
	assert.Equal(t, []string{"latency"}, task.Metrics)
	assert.Equal(t, "csv", task.Format)
	// Status can be "pending" or "processing" since async processing starts immediately
	assert.Contains(t, []string{"pending", "processing"}, task.Status)
	assert.NotZero(t, task.CreatedAt)
}

func TestExportService_CreateExport_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateExportRequest
		wantErr string
	}{
		{
			name: "Empty node IDs",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   []string{},
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now(),
				Metrics:   []string{"latency"},
				Format:    "csv",
			},
			wantErr: "at least one node_id is required",
		},
		{
			name: "Too many nodes",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   make([]string, 51),
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now(),
				Metrics:   []string{"latency"},
				Format:    "csv",
			},
			wantErr: "maximum 50 nodes",
		},
		{
			name: "End time before start time",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   []string{"node1"},
				StartTime: time.Now(),
				EndTime:   time.Now().Add(-1 * time.Hour),
				Metrics:   []string{"latency"},
				Format:    "csv",
			},
			wantErr: "end_time must be after start_time",
		},
		{
			name: "Time range too short",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   []string{"node1"},
				StartTime: time.Now().Add(-30 * time.Minute),
				EndTime:   time.Now(),
				Metrics:   []string{"latency"},
				Format:    "csv",
			},
			wantErr: "time range must be at least",
		},
		{
			name: "Time range too long",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   []string{"node1"},
				StartTime: time.Now().Add(-8 * 24 * time.Hour),
				EndTime:   time.Now(),
				Metrics:   []string{"latency"},
				Format:    "csv",
			},
			wantErr: "time range must be at most",
		},
		{
			name: "Empty metrics",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   []string{"node1"},
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now(),
				Metrics:   []string{},
				Format:    "csv",
			},
			wantErr: "at least one metric is required",
		},
		{
			name: "Invalid metric",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   []string{"node1"},
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now(),
				Metrics:   []string{"invalid_metric"},
				Format:    "csv",
			},
			wantErr: "invalid metric",
		},
		{
			name: "Invalid format",
			req: &CreateExportRequest{
				UserID:    "test-user",
				NodeIDs:   []string{"node1"},
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now(),
				Metrics:   []string{"latency"},
				Format:    "xlsx",
			},
			wantErr: "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			pool := setupTestDB(t)

			service := NewExportService(pool)
			ctx := context.Background()

			// Initialize node IDs slice if needed
			if tt.req.NodeIDs == nil && len(tt.req.NodeIDs) == 0 {
				// For the "too many nodes" test, initialize with dummy IDs
				if tt.name == "Too many nodes" {
					for i := range tt.req.NodeIDs {
						tt.req.NodeIDs[i] = uuid.New().String()
					}
				}
			}

			// Execute
			task, err := service.CreateExport(ctx, tt.req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, task)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestExportService_GetExport_Success(t *testing.T) {
	// Setup
	pool := setupTestDB(t)

	service := NewExportService(pool)
	ctx := context.Background()

	// Create test data
	nodeID := createTestNode(ctx, t, pool)
	createTestMetrics(ctx, t, pool, nodeID)

	// Create export
	req := &CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   []string{nodeID},
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now(),
		Metrics:   []string{"latency"},
		Format:    "csv",
	}

	task, err := service.CreateExport(ctx, req)
	require.NoError(t, err)

	// Execute
	retrievedTask, err := service.GetExport(task.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrievedTask.ID)
	assert.Equal(t, task.UserID, retrievedTask.UserID)
	assert.Equal(t, task.NodeIDs, retrievedTask.NodeIDs)
	assert.Equal(t, task.Metrics, retrievedTask.Metrics)
	assert.Equal(t, task.Format, retrievedTask.Format)
}

func TestExportService_GetExport_NotFound(t *testing.T) {
	// Setup
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewExportService(pool)

	// Execute
	task, err := service.GetExport("non-existent-id")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "export not found")
}

func TestExportService_ProcessExport_CompletesSuccessfully(t *testing.T) {
	// Setup
	pool := setupTestDB(t)

	service := NewExportService(pool)
	ctx := context.Background()

	// Create test data
	nodeID := createTestNode(ctx, t, pool)
	createTestMetrics(ctx, t, pool, nodeID)

	// Create export
	req := &CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   []string{nodeID},
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now(),
		Metrics:   []string{"latency", "packet_loss_rate", "jitter"},
		Format:    "csv",
	}

	task, err := service.CreateExport(ctx, req)
	require.NoError(t, err)

	// Wait for export to complete
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var completedTask *models.ExportTask
	for {
		select {
		case <-timeout:
			t.Fatal("Export did not complete in time")
		case <-ticker.C:
			updatedTask, err := service.GetExport(task.ID)
			require.NoError(t, err)

			if updatedTask.IsCompleted() {
				completedTask = updatedTask
				goto Verify
			}

			if updatedTask.IsFailed() {
				t.Fatalf("Export failed: %s", updatedTask.Error)
			}
		}
	}

Verify:
	// Assert
	assert.NotNil(t, completedTask)
	assert.Equal(t, "completed", completedTask.Status)
	assert.NotEmpty(t, completedTask.FilePath)
	assert.Greater(t, completedTask.FileSize, int64(0))
	assert.Greater(t, completedTask.RecordCount, 0)
	assert.NotNil(t, completedTask.CompletedAt)
	assert.Less(t, completedTask.GetDuration(), 5*time.Second)
}

func TestExportService_ProcessExport_NoData(t *testing.T) {
	// Setup
	pool := setupTestDB(t)

	service := NewExportService(pool)
	ctx := context.Background()

	// Create a node but no metrics
	nodeID := createTestNode(ctx, t, pool)

	// Create export
	req := &CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   []string{nodeID},
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now(),
		Metrics:   []string{"latency"},
		Format:    "csv",
	}

	task, err := service.CreateExport(ctx, req)
	require.NoError(t, err)

	// Wait for export to complete (should succeed even with no data)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Export did not complete in time")
		case <-ticker.C:
			updatedTask, err := service.GetExport(task.ID)
			require.NoError(t, err)

			if updatedTask.IsCompleted() {
				// Should complete with 0 records
				assert.Equal(t, "completed", updatedTask.Status)
				assert.Equal(t, 0, updatedTask.RecordCount)
				return
			}

			if updatedTask.IsFailed() {
				t.Fatalf("Export failed: %s", updatedTask.Error)
			}
		}
	}
}

func TestExportService_ProcessExport_MultipleNodesAndMetrics(t *testing.T) {
	// Setup
	pool := setupTestDB(t)

	service := NewExportService(pool)
	ctx := context.Background()

	// Create test data for multiple nodes
	nodeIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		nodeIDs[i] = createTestNode(ctx, t, pool)
		createTestMetrics(ctx, t, pool, nodeIDs[i])
	}

	// Create export
	req := &CreateExportRequest{
		UserID:    "test-user-id",
		NodeIDs:   nodeIDs,
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now(),
		Metrics:   []string{"latency", "packet_loss_rate", "jitter"},
		Format:    "csv",
	}

	task, err := service.CreateExport(ctx, req)
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
			updatedTask, err := service.GetExport(task.ID)
			require.NoError(t, err)

			if updatedTask.IsCompleted() {
				// Should have more records due to multiple nodes and metrics
				assert.Equal(t, "completed", updatedTask.Status)
				assert.Greater(t, updatedTask.RecordCount, 0)
				return
			}

			if updatedTask.IsFailed() {
				t.Fatalf("Export failed: %s", updatedTask.Error)
			}
		}
	}
}

// Helper functions

func setupTestDB(t *testing.T) *pgxpool.Pool {
	pool, cleanup := setupTestDBWithCleanup(t)
	t.Cleanup(cleanup)
	return pool
}

func setupTestDBWithCleanup(t *testing.T) (*pgxpool.Pool, func()) {
	// Setup test config first
	testutil.SetupTestConfig()

	// Load configuration
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

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
		testutil.TeardownTestConfig()
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

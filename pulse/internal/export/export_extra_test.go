package export

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

func validRequest() *CreateExportRequest {
	now := time.Now()
	return &CreateExportRequest{
		UserID:    "user-1",
		NodeIDs:   []string{"node-1"},
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   now,
		Metrics:   []string{"latency"},
		Format:    "csv",
	}
}

func TestCreateExportRequest_Validate_Valid(t *testing.T) {
	req := validRequest()
	err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateExportRequest_Validate_NoNodeIDs(t *testing.T) {
	req := validRequest()
	req.NodeIDs = []string{}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one node_id")
}

func TestCreateExportRequest_Validate_TooManyNodes(t *testing.T) {
	req := validRequest()
	req.NodeIDs = make([]string, MaxNodes+1)
	for i := range req.NodeIDs {
		req.NodeIDs[i] = "node"
	}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum")
}

func TestCreateExportRequest_Validate_EndBeforeStart(t *testing.T) {
	req := validRequest()
	req.EndTime = req.StartTime.Add(-time.Hour)
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end_time must be after start_time")
}

func TestCreateExportRequest_Validate_TimeRangeTooShort(t *testing.T) {
	now := time.Now()
	req := validRequest()
	req.StartTime = now.Add(-30 * time.Minute)
	req.EndTime = now
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least")
}

func TestCreateExportRequest_Validate_TimeRangeTooLong(t *testing.T) {
	now := time.Now()
	req := validRequest()
	req.StartTime = now.Add(-8 * 24 * time.Hour)
	req.EndTime = now
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at most")
}

func TestCreateExportRequest_Validate_NoMetrics(t *testing.T) {
	req := validRequest()
	req.Metrics = []string{}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one metric")
}

func TestCreateExportRequest_Validate_InvalidMetric(t *testing.T) {
	req := validRequest()
	req.Metrics = []string{"invalid_metric"}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid metric")
}

func TestCreateExportRequest_Validate_DefaultFormat(t *testing.T) {
	req := validRequest()
	req.Format = ""
	err := req.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "csv", req.Format)
}

func TestCreateExportRequest_Validate_UnsupportedFormat(t *testing.T) {
	req := validRequest()
	req.Format = "json"
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestCreateExportRequest_Validate_AllValidMetrics(t *testing.T) {
	validMetrics := []string{"latency", "packet_loss_rate", "jitter"}

	for _, metric := range validMetrics {
		req := validRequest()
		req.Metrics = []string{metric}
		err := req.Validate()
		assert.NoError(t, err, "metric %q should be valid", metric)
	}
}

func TestNewExportService(t *testing.T) {
	svc := NewExportService(nil, nil)
	assert.NotNil(t, svc)

	// Cleanup
	svc.Shutdown()
}

func TestExportService_CreateExport_ValidationError(t *testing.T) {
	svc := NewExportService(nil, nil)
	defer svc.Shutdown()

	req := validRequest()
	req.NodeIDs = []string{} // invalid

	_, err := svc.CreateExport(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestExportService_UpdateTaskStatus(t *testing.T) {
	svc := NewExportService(nil, nil)
	defer svc.Shutdown()

	// Add a task directly without calling CreateExport (to avoid nil pool panic)
	taskID := "test-task-1"
	svc.mu.Lock()
	svc.tasks[taskID] = &models.ExportTask{
		ID:     taskID,
		Status: "pending",
	}
	svc.mu.Unlock()

	// Update task status
	svc.updateTaskStatus(taskID, "processing", "", 0, 0, "")

	found, err := svc.GetExport(taskID)
	require.NoError(t, err)
	assert.Equal(t, "processing", found.Status)
}

func TestExportService_ListExports(t *testing.T) {
	svc := NewExportService(nil, nil)
	defer svc.Shutdown()

	now := time.Now()
	svc.mu.Lock()
	svc.tasks["old"] = &models.ExportTask{
		ID:        "old",
		UserID:    "user-1",
		Status:    "completed",
		CreatedAt: now.Add(-2 * time.Hour),
	}
	svc.tasks["new"] = &models.ExportTask{
		ID:        "new",
		UserID:    "user-1",
		Status:    "failed",
		CreatedAt: now,
	}
	svc.tasks["other-user"] = &models.ExportTask{
		ID:        "other-user",
		UserID:    "user-2",
		Status:    "completed",
		CreatedAt: now.Add(time.Hour),
	}
	svc.mu.Unlock()

	tasks := svc.ListExports("user-1", 10)
	require.Len(t, tasks, 2)
	assert.Equal(t, "new", tasks[0].ID)
	assert.Equal(t, "old", tasks[1].ID)

	limited := svc.ListExports("user-1", 1)
	require.Len(t, limited, 1)
	assert.Equal(t, "new", limited[0].ID)
}

func TestExportService_UpdateTaskStatus_Completed(t *testing.T) {
	svc := NewExportService(nil, nil)
	defer svc.Shutdown()

	taskID := "test-task-2"
	svc.mu.Lock()
	svc.tasks[taskID] = &models.ExportTask{
		ID:     taskID,
		Status: "pending",
	}
	svc.mu.Unlock()

	// Update to completed with file info
	svc.updateTaskStatus(taskID, "completed", "/tmp/export.csv", 1024, 100, "")

	found, err := svc.GetExport(taskID)
	require.NoError(t, err)
	assert.Equal(t, "completed", found.Status)
	assert.Equal(t, "/tmp/export.csv", found.FilePath)
	assert.Equal(t, int64(1024), found.FileSize)
	assert.Equal(t, 100, found.RecordCount)
}

func TestExportService_GetTask_NonExistent(t *testing.T) {
	svc := NewExportService(nil, nil)
	defer svc.Shutdown()

	task := svc.getTask("nonexistent")
	assert.Nil(t, task)
}

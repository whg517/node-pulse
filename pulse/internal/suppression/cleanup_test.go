package suppression

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// mockSuppressionsQuerier is a mock for AlertSuppressionsQuerier
type mockSuppressionsQuerier struct {
	suppression  *models.AlertSuppression
	checkErr     error
	createErr    error
	deletedCount int64
	deleteErr    error
	activeCount  int64
	countErr     error
}

func (m *mockSuppressionsQuerier) CheckSuppression(_ context.Context, _ string, _ string) (*models.AlertSuppression, error) {
	return m.suppression, m.checkErr
}

func (m *mockSuppressionsQuerier) CreateOrUpdateSuppression(_ context.Context, _ string, _ string, _ time.Time) error {
	return m.createErr
}

func (m *mockSuppressionsQuerier) DeleteExpiredSuppressions(_ context.Context) (int64, error) {
	return m.deletedCount, m.deleteErr
}

func (m *mockSuppressionsQuerier) CountActiveSuppressions(_ context.Context) (int64, error) {
	return m.activeCount, m.countErr
}

// ---- CleanupTask tests ----

func TestNewCleanupTask(t *testing.T) {
	querier := &mockSuppressionsQuerier{}
	task := NewCleanupTask(querier)

	assert.NotNil(t, task)
	assert.Equal(t, querier, task.querier)
}

func TestCleanupTask_Name(t *testing.T) {
	task := NewCleanupTask(&mockSuppressionsQuerier{})
	assert.Equal(t, "suppression-cleanup", task.Name())
}

func TestCleanupTask_Interval(t *testing.T) {
	task := NewCleanupTask(&mockSuppressionsQuerier{})
	assert.Equal(t, 1*time.Hour, task.Interval())
}

func TestCleanupTask_Execute_Success_NoDeleted(t *testing.T) {
	querier := &mockSuppressionsQuerier{deletedCount: 0}
	task := NewCleanupTask(querier)

	err := task.Execute(context.Background())
	require.NoError(t, err)
}

func TestCleanupTask_Execute_Success_WithDeleted(t *testing.T) {
	querier := &mockSuppressionsQuerier{deletedCount: 5}
	task := NewCleanupTask(querier)

	err := task.Execute(context.Background())
	require.NoError(t, err)
}

func TestCleanupTask_Execute_Error(t *testing.T) {
	querier := &mockSuppressionsQuerier{
		deletedCount: 0,
		deleteErr:    errors.New("database error"),
	}
	task := NewCleanupTask(querier)

	err := task.Execute(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

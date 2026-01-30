package cleanup

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/config"
)

func TestNewCleanupTask_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
		SlowThresholdMs: 30000,
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "metrics-cleanup", task.Name())
	assert.Equal(t, 3600*time.Second, task.Interval())
}

func TestNewCleanupTask_Disabled(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	cfg := &config.CleanupConfig{
		Enabled: false,
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	assert.NoError(t, err)
	assert.Nil(t, task) // Should return nil when disabled
}

func TestNewCleanupTask_InvalidInterval(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 0, // Invalid
		RetentionDays:   7,
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "invalid interval_seconds")
}

func TestNewCleanupTask_InvalidRetention(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   0, // Invalid
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "invalid retention_days")
}

func TestCleanupTask_Execute_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mock.Close()

	// Expect DELETE SQL execution with parameterized query
	mock.ExpectExec("DELETE FROM metrics WHERE timestamp < NOW\\(\\) - INTERVAL \\$1 \\* INTERVAL '1 day'").
		WithArgs(7). // RetentionDays argument
		WillReturnResult(pgxmock.NewResult("DELETE", 1234)) // Deleted 1234 rows

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
		SlowThresholdMs: 30000,
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = task.Execute(ctx)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanupTask_Execute_DatabaseError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mock.Close()

	// Simulate database error with parameterized query
	mock.ExpectExec("DELETE FROM metrics WHERE timestamp < NOW\\(\\) - INTERVAL \\$1 \\* INTERVAL '1 day'").
		WithArgs(7). // RetentionDays argument
		WillReturnError(&pgconn.PgError{
			Code:    "08006",
			Message: "connection lost",
		})

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = task.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cleanup failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanupTask_Execute_SlowQuery(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mock.Close()

	// Expect DELETE SQL execution with parameterized query
	mock.ExpectExec("DELETE FROM metrics WHERE timestamp < NOW\\(\\) - INTERVAL \\$1 \\* INTERVAL '1 day'").
		WithArgs(7). // RetentionDays argument
		WillReturnResult(pgxmock.NewResult("DELETE", 100))

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
		SlowThresholdMs: 1, // 1ms threshold - will trigger warning
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = task.Execute(ctx)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanupTask_Execute_ZeroRows(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mock.Close()

	// No rows to delete - parameterized query with argument
	mock.ExpectExec("DELETE FROM metrics WHERE timestamp < NOW\\(\\) - INTERVAL \\$1 \\* INTERVAL '1 day'").
		WithArgs(7). // RetentionDays argument
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = task.Execute(ctx)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanupTask_GetStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	cfg := &config.CleanupConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		RetentionDays:   7,
	}

	task, err := NewCleanupTask(cfg, mock, nil)
	require.NoError(t, err)

	status := task.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "metrics-cleanup", status.Name)
	assert.Equal(t, int64(0), status.RunCount)
	assert.False(t, status.IsRunning)
}

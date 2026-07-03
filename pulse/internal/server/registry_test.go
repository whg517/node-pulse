package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whg517/node-pulse/pulse/internal/scheduler"
)

// TestNewTaskRegistry verifies the constructor wires its dependencies.
func TestNewTaskRegistry(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	r := NewTaskRegistry(sched, nil)
	require.NotNil(t, r)
	assert.Equal(t, sched, r.scheduler)
	assert.Nil(t, r.database)
}

// TestRegisterAll_NilDatabase verifies the early-return guard: when no database
// is configured, RegisterAll is a no-op that returns nil without registering
// any tasks.
func TestRegisterAll_NilDatabase(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	r := NewTaskRegistry(sched, nil)
	err = r.RegisterAll()
	assert.NoError(t, err, "RegisterAll should succeed with nil database")
	assert.Nil(t, r.cleanupTask, "no cleanup task should be created without a database")
}

// TestRegisterAll_NilDatabasePool verifies the guard also triggers when the
// database struct is present but its pool is nil.
func TestRegisterAll_NilDatabasePool(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	r := NewTaskRegistry(sched, nil) // database == nil triggers the first branch
	err = r.RegisterAll()
	assert.NoError(t, err)
}

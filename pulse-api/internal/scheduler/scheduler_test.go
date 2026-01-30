package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"
)

// MockTask is a test implementation of Task interface
type MockTask struct {
	name         string
	interval     time.Duration
	executeCount int
	mu           sync.Mutex
	executeFn    func(ctx context.Context) error
	shouldFail   bool
}

func (m *MockTask) Name() string {
	return m.name
}

func (m *MockTask) Interval() time.Duration {
	return m.interval
}

func (m *MockTask) Execute(ctx context.Context) error {
	m.mu.Lock()
	m.executeCount++
	m.mu.Unlock()

	if m.executeFn != nil {
		return m.executeFn(ctx)
	}

	if m.shouldFail {
		return context.Canceled
	}
	return nil
}

func (m *MockTask) GetExecuteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.executeCount
}

func TestNewScheduler(t *testing.T) {
	sched, err := NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	if sched == nil {
		t.Fatal("Scheduler is nil")
	}
}

func TestScheduler_RegisterTask(t *testing.T) {
	sched, err := NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	task := &MockTask{
		name:     "test-task",
		interval: 1 * time.Second,
	}

	err = sched.RegisterTask(task)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	// Try to register same task again (should fail)
	err = sched.RegisterTask(task)
	if err == nil {
		t.Error("Expected error when registering duplicate task, got nil")
	}
}

func TestScheduler_StartAndStop(t *testing.T) {
	sched, err := NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	task := &MockTask{
		name:     "test-task",
		interval: 100 * time.Millisecond,
	}

	err = sched.RegisterTask(task)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	startErr := sched.Start(ctx)
	if startErr != nil {
		t.Fatalf("Failed to start scheduler: %v", startErr)
	}

	// Wait for task to execute at least once
	time.Sleep(150 * time.Millisecond)

	// Stop scheduler
	stopErr := sched.Stop()
	if stopErr != nil {
		t.Fatalf("Failed to stop scheduler: %v", stopErr)
	}

	// Verify task was executed
	count := task.GetExecuteCount()
	if count < 1 {
		t.Errorf("Expected task to execute at least once, got %d", count)
	}
}

func TestScheduler_TaskExecution(t *testing.T) {
	sched, err := NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	task := &MockTask{
		name:     "test-task",
		interval: 50 * time.Millisecond,
	}

	err = sched.RegisterTask(task)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)

	// Wait for task to execute multiple times
	time.Sleep(170 * time.Millisecond)

	// Stop scheduler
	sched.Stop()

	// Verify task executed at least 3 times (0ms, 50ms, 100ms, 150ms)
	count := task.GetExecuteCount()
	if count < 3 {
		t.Errorf("Expected task to execute at least 3 times, got %d", count)
	}
}

func TestScheduler_GracefulStop(t *testing.T) {
	sched, err := NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Task that takes time to complete
	taskRunning := make(chan struct{})
	taskCanFinish := make(chan struct{})

	task := &MockTask{
		name:     "slow-task",
		interval: 100 * time.Millisecond,
		executeFn: func(ctx context.Context) error {
			close(taskRunning)
			<-taskCanFinish // Wait for signal
			return nil
		},
	}

	err = sched.RegisterTask(task)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)

	// Wait for task to start
	<-taskRunning

	// Stop scheduler (should wait for task to complete)
	done := make(chan error)
	go func() {
		done <- sched.Stop()
	}()

	// Verify it doesn't stop immediately (task is still running)
	select {
	case <-done:
		t.Error("Scheduler stopped immediately, should have waited for task")
	case <-time.After(100 * time.Millisecond):
		// Expected - scheduler is waiting
	}

	// Let task finish
	close(taskCanFinish)

	// Now scheduler should stop
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Failed to stop scheduler: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Scheduler did not stop within timeout")
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	sched, err := NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	task := &MockTask{
		name:     "test-task",
		interval: 100 * time.Millisecond,
	}

	err = sched.RegisterTask(task)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go sched.Start(ctx)

	// Wait a bit then cancel context
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Give scheduler time to stop
	time.Sleep(100 * time.Millisecond)

	// Verify task stopped executing (should have executed 0-1 times)
	count := task.GetExecuteCount()
	if count > 2 {
		t.Errorf("Task continued after context cancel, got %d executions", count)
	}
}

func TestScheduler_GetTaskStatus(t *testing.T) {
	sched, err := NewScheduler()
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	task := &MockTask{
		name:     "test-task",
		interval: 100 * time.Millisecond,
	}

	err = sched.RegisterTask(task)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	// Get status of registered task
	status, err := sched.GetTaskStatus("test-task")
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}

	if status.Name != "test-task" {
		t.Errorf("Expected task name 'test-task', got '%s'", status.Name)
	}

	// Try to get status of non-existent task
	_, err = sched.GetTaskStatus("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent task, got nil")
	}
}

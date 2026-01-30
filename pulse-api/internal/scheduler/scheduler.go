package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// taskScheduler implements the Scheduler interface
type taskScheduler struct {
	tasks map[string]*taskState

	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// taskState holds the runtime state of a task
type taskState struct {
	task         Task
	lastRun      time.Time
	lastDuration time.Duration
	lastError    error
	runCount     int64
	isRunning    bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler() (Scheduler, error) {
	return &taskScheduler{
		tasks: make(map[string]*taskState),
	}, nil
}

// Start starts the scheduler and all registered tasks
func (s *taskScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	log.Printf("[Scheduler] Started with %d tasks", len(s.tasks))

	// Start a goroutine for each task
	for _, taskState := range s.tasks {
		s.wg.Add(1)
		go s.runTask(taskState)
	}

	return nil
}

// Stop stops the scheduler and waits for all running tasks to complete
func (s *taskScheduler) Stop() error {
	log.Println("[Scheduler] Stopping...")

	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	// Wait for all tasks to complete
	s.wg.Wait()

	log.Println("[Scheduler] Stopped")
	return nil
}

// RegisterTask registers a new task with the scheduler
func (s *taskScheduler) RegisterTask(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := task.Name()

	if _, exists := s.tasks[name]; exists {
		return fmt.Errorf("task %s already registered", name)
	}

	s.tasks[name] = &taskState{
		task: task,
	}

	log.Printf("[Scheduler] Task registered: %s (interval: %s)", name, task.Interval())

	return nil
}

// GetTaskStatus returns the current status of a task
func (s *taskScheduler) GetTaskStatus(taskName string) (*TaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	taskState, exists := s.tasks[taskName]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskName)
	}

	lastErrMsg := ""
	if taskState.lastError != nil {
		lastErrMsg = taskState.lastError.Error()
	}

	return &TaskStatus{
		Name:         taskName,
		IsRunning:    taskState.isRunning,
		LastRun:      taskState.lastRun,
		NextRun:      taskState.lastRun.Add(taskState.task.Interval()),
		LastDuration: taskState.lastDuration,
		LastError:    lastErrMsg,
		RunCount:     taskState.runCount,
	}, nil
}

// runTask runs a single task in a goroutine
func (s *taskScheduler) runTask(taskState *taskState) {
	defer s.wg.Done()

	task := taskState.task
	interval := task.Interval()

	// Create ticker for periodic execution
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[Scheduler] Task %s: started (interval: %s)", task.Name(), interval)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("[Scheduler] Task %s: stopping", task.Name())
			return

		case <-ticker.C:
			s.executeTask(taskState)
		}
	}
}

// executeTask executes a single task and records its status
func (s *taskScheduler) executeTask(taskState *taskState) {
	task := taskState.task
	start := time.Now()

	taskState.isRunning = true

	log.Printf("[Scheduler] Task %s: executing", task.Name())

	err := task.Execute(s.ctx)

	duration := time.Since(start)
	taskState.isRunning = false
	taskState.lastRun = start
	taskState.lastDuration = duration
	taskState.runCount++
	taskState.lastError = err

	if err != nil {
		log.Printf("[Scheduler] Task %s: failed (duration: %s, error: %v)",
			task.Name(), duration, err)
	} else {
		log.Printf("[Scheduler] Task %s: completed (duration: %s)",
			task.Name(), duration)
	}
}

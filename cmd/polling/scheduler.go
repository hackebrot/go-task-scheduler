package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hackebrot/go-task-scheduler/pkg/scheduler"
)

// pollingScheduler is a task scheduler that checks for ready tasks at fixed intervals.
type pollingScheduler struct {
	tasks         []scheduler.Task
	mu            sync.Mutex
	checkInterval time.Duration
}

// NewScheduler creates a new pollingScheduler instance.
func NewScheduler(checkInterval time.Duration) *pollingScheduler {
	return &pollingScheduler{
		tasks:         make([]scheduler.Task, 0),
		checkInterval: checkInterval,
	}
}

// Schedule adds a task to the scheduler.
func (s *pollingScheduler) Schedule(task scheduler.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, task)
}

// Start runs the scheduler loop, checking for ready tasks at regular intervals until all tasks complete or context is cancelled.
func (s *pollingScheduler) Start(ctx context.Context) {
	slog.Info("starting polling task scheduler", "count_tasks", s.PendingTasksCount())

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		if s.PendingTasksCount() == 0 {
			slog.Info("no remaining tasks, stopping scheduler")
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.executeReadyTasks()
		}
	}
}

// executeReadyTasks identifies and executes all tasks that are ready to run, updating the task list with remaining tasks.
func (s *pollingScheduler) executeReadyTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	remaining := make([]scheduler.Task, 0, len(s.tasks))

	for _, task := range s.tasks {
		if now.Before(task.ExecuteAt()) {
			remaining = append(remaining, task)
		} else {
			go s.executeTask(task)
		}
	}

	s.tasks = remaining
}

// executeTask runs a single task and logs any errors that occur.
func (s *pollingScheduler) executeTask(task scheduler.Task) {
	if err := task.Execute(); err != nil {
		slog.Error("error executing task", "task_id", task.ID(), "error", err)
	}
}

// PendingTasksCount returns the number of tasks waiting to be executed.
func (s *pollingScheduler) PendingTasksCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tasks)
}

var _ scheduler.Scheduler = (*pollingScheduler)(nil)

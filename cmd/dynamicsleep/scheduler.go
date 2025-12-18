package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hackebrot/go-task-scheduler/pkg/scheduler"
)

// defaultCheckInterval is the fallback interval for checking tasks when no tasks are scheduled.
const defaultCheckInterval = 1 * time.Second

// dynamicSleepScheduler is a task scheduler that dynamically calculates sleep duration until the next task.
type dynamicSleepScheduler struct {
	tasks []scheduler.Task
	mu    sync.Mutex
}

// NewScheduler creates a new dynamicSleepScheduler instance.
func NewScheduler() *dynamicSleepScheduler {
	return &dynamicSleepScheduler{
		tasks: make([]scheduler.Task, 0),
	}
}

// batch represents a group of tasks ready for execution and the duration to sleep until the next batch.
type batch struct {
	ready         []scheduler.Task
	sleepDuration time.Duration
}

// Schedule adds a task to the scheduler.
func (s *dynamicSleepScheduler) Schedule(task scheduler.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, task)
}

// nextBatch partitions tasks into ready and remaining, returning ready tasks and sleep duration until next task.
func (s *dynamicSleepScheduler) nextBatch() batch {
	s.mu.Lock()
	defer s.mu.Unlock()

	ready := make([]scheduler.Task, 0, len(s.tasks))
	remaining := make([]scheduler.Task, 0, len(s.tasks))
	var nextTime time.Time
	hasNext := false

	now := time.Now()

	for _, task := range s.tasks {
		if !now.Before(task.ExecuteAt()) {
			ready = append(ready, task)
			continue
		}

		remaining = append(remaining, task)
		if !hasNext {
			nextTime = task.ExecuteAt()
			hasNext = true
		} else if task.ExecuteAt().Before(nextTime) {
			nextTime = task.ExecuteAt()
		}
	}

	s.tasks = remaining

	var sleepDuration time.Duration
	if hasNext {
		sleepDuration = max(time.Until(nextTime), 0)
	} else {
		sleepDuration = defaultCheckInterval
	}

	return batch{
		ready:         ready,
		sleepDuration: sleepDuration,
	}
}

// Start runs the scheduler loop, executing tasks as they become ready until all tasks complete or context is cancelled.
func (s *dynamicSleepScheduler) Start(ctx context.Context) {
	slog.Info("starting dynamic sleep task scheduler", "count_tasks", s.PendingTasksCount())

	for {
		if s.PendingTasksCount() == 0 {
			slog.Info("no remaining tasks, stopping scheduler")
			return
		}

		b := s.nextBatch()

		for _, task := range b.ready {
			if err := task.Execute(); err != nil {
				slog.Error("error executing task", "task_id", task.ID(), "error", err)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(b.sleepDuration):
		}
	}
}

// PendingTasksCount returns the number of tasks waiting to be executed.
func (s *dynamicSleepScheduler) PendingTasksCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tasks)
}

var _ scheduler.Scheduler = (*dynamicSleepScheduler)(nil)

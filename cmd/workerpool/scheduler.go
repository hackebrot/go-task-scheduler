package main

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/hackebrot/go-task-scheduler/pkg/scheduler"
)

const defaultReadyTasksBufSize = 100

// workerpoolScheduler is a task scheduler that uses a pool of workers to execute ready tasks concurrently.
type workerpoolScheduler struct {
	tasks         []scheduler.Task
	mu            sync.Mutex
	checkInterval time.Duration
	workerCount   int
	wg            sync.WaitGroup
	readyTasks    chan scheduler.Task
	results       chan scheduler.TaskResult
}

// NewScheduler creates a new workerpoolScheduler instance.
func NewScheduler(checkInterval time.Duration, workerCount, readyTasksBufSize int) *workerpoolScheduler {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	if readyTasksBufSize <= 0 {
		readyTasksBufSize = defaultReadyTasksBufSize
	}

	return &workerpoolScheduler{
		tasks:         make([]scheduler.Task, 0),
		checkInterval: checkInterval,
		workerCount:   workerCount,
		readyTasks:    make(chan scheduler.Task, readyTasksBufSize),
		results:       make(chan scheduler.TaskResult, readyTasksBufSize),
	}
}

// Schedule adds a task to the scheduler.
func (s *workerpoolScheduler) Schedule(task scheduler.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, task)
}

// Start runs the scheduler loop, executing ready tasks via a worker pool until all tasks complete or context is cancelled.
func (s *workerpoolScheduler) Start(ctx context.Context) {
	slog.Info("starting workerpool task scheduler", "count_tasks", s.PendingTasksCount())

	// Start worker pool
	s.wg.Add(s.workerCount)
	for i := range s.workerCount {
		go s.runWorker(i)
	}

	// Start task scanner
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		if s.PendingTasksCount() == 0 {
			slog.Info("no remaining tasks, stopping scheduler")
			s.shutdown()
			return
		}

		select {
		case <-ctx.Done():
			slog.Info("context cancelled, stopping scheduler")
			s.shutdown()
			return
		case <-ticker.C:
			s.dispatchReadyTasks(ctx)
		}
	}
}

// shutdown gracefully stops the worker pool and closes channels.
func (s *workerpoolScheduler) shutdown() {
	close(s.readyTasks)
	s.wg.Wait()
	close(s.results)
}

// dispatchReadyTasks sends ready tasks to the worker pool and updates the task list with remaining tasks.
func (s *workerpoolScheduler) dispatchReadyTasks(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	remaining := make([]scheduler.Task, 0, len(s.tasks))

	for _, task := range s.tasks {
		if now.Before(task.ExecuteAt()) {
			remaining = append(remaining, task)
		} else {
			select {
			case s.readyTasks <- task:
			case <-ctx.Done():
				remaining = append(remaining, task)
				s.tasks = remaining
				return
			}
		}
	}

	s.tasks = remaining
}

// runWorker processes tasks from the readyTasks channel and sends results.
func (s *workerpoolScheduler) runWorker(id int) {
	defer s.wg.Done()

	for task := range s.readyTasks {
		result := s.executeTask(task, id)
		s.results <- result
	}
}

// executeTask runs a single task and returns a result with timing and any errors that occur.
func (s *workerpoolScheduler) executeTask(task scheduler.Task, workerID int) scheduler.TaskResult {
	startTime := time.Now()
	err := task.Execute()
	endTime := time.Now()

	return scheduler.TaskResult{
		TaskID:    task.ID(),
		Error:     err,
		StartTime: startTime,
		EndTime:   endTime,
		WorkerID:  strconv.Itoa(workerID),
	}
}

// PendingTasksCount returns the number of tasks waiting to be executed.
func (s *workerpoolScheduler) PendingTasksCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tasks)
}

// Results returns a channel that receives task execution results.
// Consumer should read from this channel to collect results.
// The channel is closed when the scheduler stops.
func (s *workerpoolScheduler) Results() <-chan scheduler.TaskResult {
	return s.results
}

var _ scheduler.Scheduler = (*workerpoolScheduler)(nil)

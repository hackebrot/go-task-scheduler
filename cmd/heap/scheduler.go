package main

import (
	"container/heap"
	"context"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/hackebrot/go-task-scheduler/pkg/scheduler"
)

const defaultReadyTasksBufSize = 100

// heapScheduler is a task scheduler that uses a min-heap to efficiently manage task priorities
// and a worker pool to execute ready tasks concurrently.
type heapScheduler struct {
	tasks         *taskHeap
	mu            sync.Mutex
	workerCount   int
	wg            sync.WaitGroup
	readyTasks    chan scheduler.Task
	results       chan scheduler.TaskResult
}

// NewScheduler creates a new heapScheduler instance.
func NewScheduler(workerCount, readyTasksBufSize int) *heapScheduler {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	if readyTasksBufSize <= 0 {
		readyTasksBufSize = defaultReadyTasksBufSize
	}

	h := &taskHeap{}
	heap.Init(h)

	return &heapScheduler{
		tasks:       h,
		workerCount: workerCount,
		readyTasks:  make(chan scheduler.Task, readyTasksBufSize),
		results:     make(chan scheduler.TaskResult, readyTasksBufSize),
	}
}

// Schedule adds a task to the scheduler's heap.
func (s *heapScheduler) Schedule(task scheduler.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	heap.Push(s.tasks, task)
}

// Start runs the scheduler loop, executing ready tasks via a worker pool until all tasks complete or context is cancelled.
func (s *heapScheduler) Start(ctx context.Context) {
	slog.Info("starting heap task scheduler", "count_tasks", s.PendingTasksCount())

	// Start worker pool
	s.wg.Add(s.workerCount)
	for i := range s.workerCount {
		go s.runWorker(i)
	}

	// Start task scanner with dynamic sleep based on next task time
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
		default:
			sleepDuration := s.dispatchReadyTasks(ctx)
			if sleepDuration > 0 {
				timer := time.NewTimer(sleepDuration)
				select {
				case <-ctx.Done():
					timer.Stop()
					slog.Info("context cancelled, stopping scheduler")
					s.shutdown()
					return
				case <-timer.C:
					// Continue to next iteration
				}
			}
		}
	}
}

// shutdown gracefully stops the worker pool and closes channels.
func (s *heapScheduler) shutdown() {
	close(s.readyTasks)
	s.wg.Wait()
	close(s.results)
}

// dispatchReadyTasks sends ready tasks to the worker pool and returns the duration to sleep until the next task.
// Returns the sleep duration until the next task, or 0 if the heap is now empty (allowing immediate shutdown check).
func (s *heapScheduler) dispatchReadyTasks(ctx context.Context) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	dispatched := false

	// The heap keeps tasks sorted by ExecuteAt, so we only need to check the top
	for s.tasks.Len() > 0 {
		nextTask := (*s.tasks)[0]

		if now.Before(nextTask.ExecuteAt()) {
			return nextTask.ExecuteAt().Sub(now)
		}

		task := heap.Pop(s.tasks).(scheduler.Task)
		dispatched = true

		select {
		case s.readyTasks <- task:
		case <-ctx.Done():
			heap.Push(s.tasks, task)
			return 0
		}
	}

	// Heap is empty. Return 0 so Start() can immediately check PendingTasksCount and shutdown.
	if dispatched {
		return 0
	}

	// Defensive: heap was already empty (shouldn't happen since Start checks PendingTasksCount first)
	return 100 * time.Millisecond
}

// runWorker processes tasks from the readyTasks channel and sends results.
func (s *heapScheduler) runWorker(id int) {
	defer s.wg.Done()

	for task := range s.readyTasks {
		result := s.executeTask(task, id)
		s.results <- result
	}
}

// executeTask runs a single task and returns a result with timing and any errors that occur.
func (s *heapScheduler) executeTask(task scheduler.Task, workerID int) scheduler.TaskResult {
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
func (s *heapScheduler) PendingTasksCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tasks.Len()
}

// Results returns a channel that receives task execution results.
// Consumer should read from this channel to collect results.
// The channel is closed when the scheduler stops.
func (s *heapScheduler) Results() <-chan scheduler.TaskResult {
	return s.results
}

var _ scheduler.Scheduler = (*heapScheduler)(nil)

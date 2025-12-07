package scheduler

import "context"

// Scheduler manages and executes tasks at their scheduled times.
type Scheduler interface {
	// Schedule adds a task to the scheduler.
	Schedule(task Task)

	// Start begins the scheduler's background processing.
	// It runs until the provided context is cancelled.
	Start(ctx context.Context)

	// PendingTasksCount returns the number of tasks waiting to be executed.
	PendingTasksCount() int
}

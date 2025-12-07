package scheduler

import "time"

// Task represents a schedulable unit of work with an execution time.
type Task interface {
	// Execute runs the task and returns an error if the execution fails.
	Execute() error

	// ExecuteAt returns the time when this task should be executed.
	ExecuteAt() time.Time

	// ID returns a unique identifier for this task (used for logging).
	ID() string
}

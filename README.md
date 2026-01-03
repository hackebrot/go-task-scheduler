# go-task-scheduler

Task scheduler implementations in Go, from synchronous to heap-based worker pool.

## Installation

```bash
go get github.com/hackebrot/go-task-scheduler
```

## Usage

All schedulers implement a common interface:

```go
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


// Task represents a schedulable unit of work with an execution time.
type Task interface {
	// Execute runs the task and returns an error if the execution fails.
	Execute() error

	// ExecuteAt returns the time when this task should be executed.
	ExecuteAt() time.Time

	// ID returns a unique identifier for this task (used for logging).
	ID() string
}
```

## Schedulers

| Scheduler | Execution Model | Time Complexity | Space Complexity | Best For | Source |
|-----------|----------------|-----------------|------------------|----------|--------|
| Dynamic Sleep | Sequential | O(n) per batch | O(n) | Low concurrency requirements | [cmd/dynamicsleep](cmd/dynamicsleep) |
| Polling | Concurrent (unbounded) | O(n) per interval | O(n) | Low task counts, minimal code complexity | [cmd/polling](cmd/polling) |
| Worker Pool | Concurrent (bounded) | O(n) per interval | O(n) | Controlled concurrency, result tracking | [cmd/workerpool](cmd/workerpool) |
| Heap | Concurrent (bounded) | O(log n) per dispatch | O(n) | High task volumes, performance-critical applications | [cmd/heap](cmd/heap) |

## License

MIT

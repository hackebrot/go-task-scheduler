package fib

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hackebrot/go-fibonacci"
)

// Task computes the nth Fibonacci number using a specified strategy.
type Task struct {
	id        string
	executeAt time.Time
	n         int
	strategy  fibonacci.Strategy
}

// Execute computes the Fibonacci number and logs the result.
func (t *Task) Execute() error {
	slog.Info("starting computation", "task_id", t.id, "n", t.n)

	// Intentional error for demonstration purposes to showcase error handling
	if t.n == 4 {
		return fmt.Errorf("computation failed: n=4 is not supported")
	}

	r := t.strategy.Compute(t.n)
	slog.Info("computation complete", "task_id", t.id, "n", t.n, "result", r)

	return nil
}

// ExecuteAt returns the scheduled execution time for the task.
func (t *Task) ExecuteAt() time.Time {
	return t.executeAt
}

// ID returns the task identifier.
func (t *Task) ID() string {
	return t.id
}

// NewTask creates a new Fibonacci computation task.
func NewTask(id string, n int, strategy fibonacci.Strategy, executeAt time.Time) *Task {
	return &Task{
		id:        id,
		executeAt: executeAt,
		n:         n,
		strategy:  strategy,
	}
}

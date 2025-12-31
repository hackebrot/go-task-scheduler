package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hackebrot/go-fibonacci"
	"github.com/hackebrot/go-task-scheduler/internal/fib"
)

// maxTimeOffset is the maximum number of seconds to randomly offset task execution times.
const maxTimeOffset = 20

// checkInterval is the frequency at which the scheduler polls for ready tasks.
const checkInterval = 100 * time.Millisecond

func main() {
	scheduler := NewScheduler(checkInterval, -1, 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("received shutdown signal")
		cancel()
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		processResults(scheduler.Results())
	}()

	for n := 1; n <= 10; n++ {
		offsetSeconds := rand.Intn(maxTimeOffset)
		offset := time.Now().Add(time.Duration(offsetSeconds) * time.Second)
		taskId := fmt.Sprintf("fib%d-+%ds", n, offsetSeconds)
		task := fib.NewTask(taskId, n, fibonacci.NewRecursive(), offset)
		scheduler.Schedule(task)
	}

	scheduler.Start(ctx)
	<-done
}

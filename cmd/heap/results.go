package main

import (
	"log/slog"
	"slices"
	"time"

	"github.com/hackebrot/go-task-scheduler/pkg/scheduler"
)

// processResults logs task results and computes summary statistics for successful executions.
func processResults(results <-chan scheduler.TaskResult) {
	var durations []time.Duration
	for result := range results {
		duration := result.EndTime.Sub(result.StartTime)
		args := []any{
			"task_id", result.TaskID,
			"worker_id", result.WorkerID,
			"duration_microseconds", duration.Microseconds(),
		}

		if result.Error != nil {
			args = append(args, "error", result.Error)
			slog.Error("error executing task", args...)
		} else {
			durations = append(durations, duration)
			slog.Info("task completed", args...)
		}
	}

	if len(durations) > 0 {
		slices.Sort(durations)

		var total time.Duration
		for _, d := range durations {
			total += d
		}
		// Convert len to time.Duration for division to get mean duration
		mean := total / time.Duration(len(durations))
		median := durations[len(durations)/2]

		args := []any{
			"count", len(durations),
			"mean_microseconds", mean.Microseconds(),
			"median_microseconds", median.Microseconds(),
			"min_microseconds", durations[0].Microseconds(),
			"max_microseconds", durations[len(durations)-1].Microseconds(),
		}

		// Percentiles require sufficient samples to be meaningful (1% of 100 = 1 sample)
		if len(durations) >= 100 {
			p95 := durations[int(float64(len(durations)-1)*0.95)]
			p99 := durations[int(float64(len(durations)-1)*0.99)]
			args = append(args,
				"p95_microseconds", p95.Microseconds(),
				"p99_microseconds", p99.Microseconds(),
			)
		}

		slog.Info("task execution summary", args...)
	}
}

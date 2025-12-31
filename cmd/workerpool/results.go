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
		if result.Error != nil {
			slog.Error(
				"error executing task",
				"task_id", result.TaskID,
				"worker_id", result.WorkerID,
				"duration_microseconds", duration.Microseconds(),
				"error", result.Error,
			)
		} else {
			durations = append(durations, duration)
			slog.Info(
				"task completed",
				"task_id", result.TaskID,
				"worker_id", result.WorkerID,
				"duration_microseconds", duration.Microseconds(),
			)
		}
	}

	if len(durations) > 0 {
		slices.Sort(durations)

		var total time.Duration
		for _, d := range durations {
			total += d
		}
		mean := total / time.Duration(len(durations))
		median := durations[len(durations)/2]
		p95 := durations[int(float64(len(durations))*0.95)]
		p99 := durations[int(float64(len(durations))*0.99)]

		slog.Info(
			"task execution summary",
			"count", len(durations),
			"mean_microseconds", mean.Microseconds(),
			"median_microseconds", median.Microseconds(),
			"p95_microseconds", p95.Microseconds(),
			"p99_microseconds", p99.Microseconds(),
			"min_microseconds", durations[0].Microseconds(),
			"max_microseconds", durations[len(durations)-1].Microseconds(),
		)
	}
}

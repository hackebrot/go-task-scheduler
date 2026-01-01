package main

import "github.com/hackebrot/go-task-scheduler/pkg/scheduler"

// taskHeap implements heap.Interface for scheduler.Task, ordered by ExecuteAt time.
type taskHeap []scheduler.Task

func (h taskHeap) Len() int { return len(h) }

func (h taskHeap) Less(i, j int) bool {
	return h[i].ExecuteAt().Before(h[j].ExecuteAt())
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *taskHeap) Push(x any) {
	*h = append(*h, x.(scheduler.Task))
}

func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	task := old[n-1]
	*h = old[0 : n-1]
	return task
}

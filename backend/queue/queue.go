// Package queue provides a lightweight in-process job queue backed by a
// buffered Go channel.
//
// It is used in dev/single-binary mode so that the API and the embedded build
// worker can exchange deployment jobs without any external infrastructure.
// The payload format is identical to the Redis queue -- raw JSON bytes -- so
// no serialisation changes are needed elsewhere in the codebase.
//
// This package is intentionally NOT under internal/ so that the combined
// cmd/pushpaka binary (a separate module) can import it.
package queue

import (
	"fmt"
	"sync/atomic"
)

// InProcess is a buffered-channel job queue with built-in worker/job counters.
type InProcess struct {
	ch           chan []byte
	totalWorkers atomic.Int32
	activeJobs   atomic.Int32
}

// New returns an InProcess queue with the given buffer capacity.
// A capacity of 100 is more than enough for local development.
func New(capacity int) *InProcess {
	if capacity <= 0 {
		capacity = 100
	}
	return &InProcess{ch: make(chan []byte, capacity)}
}

// Push enqueues a raw JSON job payload.
// Returns an error only if the channel buffer is full (non-blocking).
func (q *InProcess) Push(payload []byte) error {
	select {
	case q.ch <- payload:
		return nil
	default:
		return fmt.Errorf("in-process queue full (capacity %d)", cap(q.ch))
	}
}

// Chan returns the receive-only end of the channel for the worker to consume.
func (q *InProcess) Chan() <-chan []byte {
	return q.ch
}

// WorkerStarted increments the total worker count (call when a worker goroutine starts).
func (q *InProcess) WorkerStarted() { q.totalWorkers.Add(1) }

// WorkerStopped decrements the total worker count (call when a worker goroutine exits).
func (q *InProcess) WorkerStopped() { q.totalWorkers.Add(-1) }

// JobStarted increments the active-job count (call when a job begins processing).
func (q *InProcess) JobStarted() { q.activeJobs.Add(1) }

// JobFinished decrements the active-job count (call when a job finishes processing).
func (q *InProcess) JobFinished() { q.activeJobs.Add(-1) }

// TotalWorkers returns the number of currently running worker goroutines.
func (q *InProcess) TotalWorkers() int { return int(q.totalWorkers.Load()) }

// ActiveJobs returns the number of jobs currently being processed.
func (q *InProcess) ActiveJobs() int { return int(q.activeJobs.Load()) }

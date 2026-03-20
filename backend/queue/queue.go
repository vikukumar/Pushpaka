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
	"sync"
	"sync/atomic"
)

// InProcess is a buffered-channel job queue with built-in worker/job counters.
type InProcess struct {
	channels     map[string]chan []byte
	mu           sync.RWMutex
	capacity     int
	totalWorkers atomic.Int32
	activeJobs   atomic.Int32
	// Worker counts
	syncWorkers  atomic.Int32
	buildWorkers atomic.Int32
	testWorkers  atomic.Int32
	aiWorkers    atomic.Int32
	deployWorkers atomic.Int32
	// Active jobs per role
	syncActive   atomic.Int32
	buildActive  atomic.Int32
	testActive   atomic.Int32
	aiActive     atomic.Int32
	deployActive atomic.Int32
}

// New returns an InProcess queue with the given buffer capacity per role.
func New(capacity int) *InProcess {
	if capacity <= 0 {
		capacity = 100
	}
	return &InProcess{
		channels: make(map[string]chan []byte),
		capacity: capacity,
	}
}

func (q *InProcess) getChannel(role string) chan []byte {
	q.mu.RLock()
	ch, ok := q.channels[role]
	q.mu.RUnlock()
	if ok {
		return ch
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	// Double-check
	if ch, ok = q.channels[role]; ok {
		return ch
	}
	ch = make(chan []byte, q.capacity)
	q.channels[role] = ch
	return ch
}

// Push enqueues a raw JSON job payload for a specific role (e.g. "sync", "build").
func (q *InProcess) Push(role string, payload []byte) error {
	ch := q.getChannel(role)
	select {
	case ch <- payload:
		return nil
	default:
		return fmt.Errorf("in-process queue for role '%s' full (capacity %d)", role, q.capacity)
	}
}

// Chan returns the receive-only end of the channel for the worker role to consume.
func (q *InProcess) Chan(role string) <-chan []byte {
	return q.getChannel(role)
}

// WorkerStarted increments the total worker count and role-specific count.
func (q *InProcess) WorkerStarted(role string) {
	q.totalWorkers.Add(1)
	switch role {
	case "sync", "syncer":
		q.syncWorkers.Add(1)
	case "build", "builder":
		q.buildWorkers.Add(1)
	case "test", "tester":
		q.testWorkers.Add(1)
	case "ai":
		q.aiWorkers.Add(1)
	case "deploy", "deployer":
		q.deployWorkers.Add(1)
	}
}

// WorkerStopped decrements the total worker count and role-specific count.
func (q *InProcess) WorkerStopped(role string) {
	q.totalWorkers.Add(-1)
	switch role {
	case "sync", "syncer":
		q.syncWorkers.Add(-1)
	case "build", "builder":
		q.buildWorkers.Add(-1)
	case "test", "tester":
		q.testWorkers.Add(-1)
	case "ai":
		q.aiWorkers.Add(-1)
	case "deploy", "deployer":
		q.deployWorkers.Add(-1)
	}
}

// JobStarted increments the active-job count.
func (q *InProcess) JobStarted(role string) {
	q.activeJobs.Add(1)
	switch role {
	case "sync", "syncer":
		q.syncActive.Add(1)
	case "build", "builder":
		q.buildActive.Add(1)
	case "test", "tester":
		q.testActive.Add(1)
	case "ai":
		q.aiActive.Add(1)
	case "deploy", "deployer":
		q.deployActive.Add(1)
	}
}

// JobFinished decrements the active-job count.
func (q *InProcess) JobFinished(role string) {
	q.activeJobs.Add(-1)
	switch role {
	case "sync", "syncer":
		q.syncActive.Add(-1)
	case "build", "builder":
		q.buildActive.Add(-1)
	case "test", "tester":
		q.testActive.Add(-1)
	case "ai":
		q.aiActive.Add(-1)
	case "deploy", "deployer":
		q.deployActive.Add(-1)
	}
}

// Getters
func (q *InProcess) TotalWorkers() int { return int(q.totalWorkers.Load()) }
func (q *InProcess) ActiveJobs() int   { return int(q.activeJobs.Load()) }
func (q *InProcess) SyncWorkers() int  { return int(q.syncWorkers.Load()) }
func (q *InProcess) BuildWorkers() int { return int(q.buildWorkers.Load()) }
func (q *InProcess) TestWorkers() int  { return int(q.testWorkers.Load()) }
func (q *InProcess) AIWorkers() int    { return int(q.aiWorkers.Load()) }
func (q *InProcess) DeployWorkers() int { return int(q.deployWorkers.Load()) }

func (q *InProcess) SyncActive() int   { return int(q.syncActive.Load()) }
func (q *InProcess) BuildActive() int  { return int(q.buildActive.Load()) }
func (q *InProcess) TestActive() int   { return int(q.testActive.Load()) }
func (q *InProcess) AIActive() int     { return int(q.aiActive.Load()) }
func (q *InProcess) DeployActive() int { return int(q.deployActive.Load()) }

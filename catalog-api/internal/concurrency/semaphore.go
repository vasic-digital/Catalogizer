package concurrency

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/semaphore"
)

// Semaphore limits concurrent access to a resource using a buffered channel.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new Semaphore that allows up to maxConcurrent
// simultaneous acquisitions.
func NewSemaphore(maxConcurrent int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, maxConcurrent)}
}

// Acquire blocks until a slot is available or the context is cancelled.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a previously acquired slot.
func (s *Semaphore) Release() {
	<-s.ch
}

// TryAcquire attempts to acquire a slot without blocking.
// Returns true if successful, false if the semaphore is full.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// BoundedSemaphore wraps a weighted semaphore with a default timeout.
// It is designed for bounding concurrent work such as parallel
// metadata-provider lookups, where unbounded goroutines would overwhelm
// external APIs or exhaust local resources.
type BoundedSemaphore struct {
	sem     *semaphore.Weighted
	timeout time.Duration
	name    string
	max     int64
}

// NewBoundedSemaphore creates a semaphore that allows up to
// maxConcurrent simultaneous acquisitions with a default acquire
// timeout.
func NewBoundedSemaphore(name string, maxConcurrent int64, timeout time.Duration) *BoundedSemaphore {
	return &BoundedSemaphore{
		sem:     semaphore.NewWeighted(maxConcurrent),
		timeout: timeout,
		name:    name,
		max:     maxConcurrent,
	}
}

// Acquire acquires a semaphore slot, applying the configured default
// timeout on top of the caller's context.  If the slot cannot be
// obtained before the timeout (or the parent context is cancelled),
// an error is returned.
func (bs *BoundedSemaphore) Acquire(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, bs.timeout)
	defer cancel()
	if err := bs.sem.Acquire(ctx, 1); err != nil {
		return fmt.Errorf("semaphore %s acquire timeout: %w", bs.name, err)
	}
	return nil
}

// Release releases a previously acquired semaphore slot.
func (bs *BoundedSemaphore) Release() {
	bs.sem.Release(1)
}

// Name returns the semaphore's descriptive name.
func (bs *BoundedSemaphore) Name() string {
	return bs.name
}

// MaxConcurrent returns the configured maximum concurrency.
func (bs *BoundedSemaphore) MaxConcurrent() int64 {
	return bs.max
}

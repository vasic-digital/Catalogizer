package concurrency

import "context"

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

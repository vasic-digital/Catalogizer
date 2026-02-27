package semaphore

import (
	"context"
	"sync"
)

type Semaphore struct {
	ch     chan struct{}
	mu     sync.RWMutex
	closed bool
}

func New(maxConcurrent int) *Semaphore {
	return &Semaphore{
		ch:     make(chan struct{}, maxConcurrent),
		closed: false,
	}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return ErrSemaphoreClosed
	}
	s.mu.RUnlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.ch <- struct{}{}:
		return nil
	}
}

func (s *Semaphore) TryAcquire() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return false
	}

	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Semaphore) Release() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return
	}

	select {
	case <-s.ch:
	default:
	}
}

func (s *Semaphore) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.ch)
}

func (s *Semaphore) Available() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return 0
	}
	return cap(s.ch) - len(s.ch)
}

func (s *Semaphore) Acquired() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return 0
	}
	return len(s.ch)
}

func (s *Semaphore) Capacity() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cap(s.ch)
}

var ErrSemaphoreClosed = errorf("semaphore is closed")

func errorf(msg string) error {
	return &semaphoreError{msg: msg}
}

type semaphoreError struct {
	msg string
}

func (e *semaphoreError) Error() string {
	return e.msg
}

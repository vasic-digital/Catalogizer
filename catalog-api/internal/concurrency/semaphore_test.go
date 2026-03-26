package concurrency

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSemaphore_LimitsConcurrency(t *testing.T) {
	sem := NewSemaphore(2)
	var running int32
	var maxRunning int32

	done := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func() {
			err := sem.Acquire(context.Background())
			assert.NoError(t, err)
			cur := atomic.AddInt32(&running, 1)
			for {
				old := atomic.LoadInt32(&maxRunning)
				if cur <= old || atomic.CompareAndSwapInt32(&maxRunning, old, cur) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&running, -1)
			sem.Release()
			done <- struct{}{}
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
	assert.LessOrEqual(t, atomic.LoadInt32(&maxRunning), int32(2))
}

func TestSemaphore_ContextCancellation(t *testing.T) {
	sem := NewSemaphore(1)
	_ = sem.Acquire(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := sem.Acquire(ctx)
	assert.Error(t, err)
}

func TestSemaphore_TryAcquire(t *testing.T) {
	sem := NewSemaphore(1)
	assert.True(t, sem.TryAcquire())
	assert.False(t, sem.TryAcquire()) // Already full
	sem.Release()
	assert.True(t, sem.TryAcquire()) // Available again
}

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

// --- BoundedSemaphore tests ---

func TestBoundedSemaphore_LimitsConcurrency(t *testing.T) {
	bs := NewBoundedSemaphore("test", 2, 5*time.Second)
	var running int32
	var maxRunning int32

	done := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func() {
			err := bs.Acquire(context.Background())
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
			bs.Release()
			done <- struct{}{}
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
	assert.LessOrEqual(t, atomic.LoadInt32(&maxRunning), int32(2))
}

func TestBoundedSemaphore_AcquireTimeout(t *testing.T) {
	// Timeout of 50ms; fill the single slot and then try to acquire again.
	bs := NewBoundedSemaphore("timeout-test", 1, 50*time.Millisecond)
	err := bs.Acquire(context.Background())
	assert.NoError(t, err)

	// Second acquire should time out.
	err = bs.Acquire(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "semaphore timeout-test acquire timeout")

	bs.Release()

	// After release, acquire should succeed again.
	err = bs.Acquire(context.Background())
	assert.NoError(t, err)
	bs.Release()
}

func TestBoundedSemaphore_ContextCancellation(t *testing.T) {
	bs := NewBoundedSemaphore("cancel-test", 1, 5*time.Second)
	_ = bs.Acquire(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := bs.Acquire(ctx)
	assert.Error(t, err)
	bs.Release()
}

func TestBoundedSemaphore_NameAndMax(t *testing.T) {
	bs := NewBoundedSemaphore("my-sem", 4, time.Second)
	assert.Equal(t, "my-sem", bs.Name())
	assert.Equal(t, int64(4), bs.MaxConcurrent())
}

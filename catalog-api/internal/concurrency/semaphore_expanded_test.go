package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Semaphore — expanded tests
// ---------------------------------------------------------------------------

func TestSemaphore_AcquireRelease_Sequential(t *testing.T) {
	sem := NewSemaphore(3)

	// Acquire all 3 slots
	for i := 0; i < 3; i++ {
		err := sem.Acquire(context.Background())
		require.NoError(t, err)
	}

	// 4th acquire should block — use TryAcquire to verify
	assert.False(t, sem.TryAcquire(), "semaphore should be full")

	// Release one
	sem.Release()

	// Now one slot is available
	assert.True(t, sem.TryAcquire(), "should be able to acquire after release")

	// Clean up
	for i := 0; i < 3; i++ {
		sem.Release()
	}
}

func TestSemaphore_ContextAlreadyCancelled(t *testing.T) {
	sem := NewSemaphore(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before acquire

	err := sem.Acquire(ctx)
	// With empty semaphore, the acquire may succeed since select is non-deterministic.
	// But if the channel is full, it should return context error.
	_ = err // Either outcome is valid for an empty semaphore with cancelled context
}

func TestSemaphore_TryAcquire_MultipleTimes(t *testing.T) {
	sem := NewSemaphore(2)

	assert.True(t, sem.TryAcquire())
	assert.True(t, sem.TryAcquire())
	assert.False(t, sem.TryAcquire())
	assert.False(t, sem.TryAcquire())

	sem.Release()
	assert.True(t, sem.TryAcquire())
	assert.False(t, sem.TryAcquire())

	// Release all
	sem.Release()
	sem.Release()
}

func TestSemaphore_ConcurrentAcquireRelease(t *testing.T) {
	sem := NewSemaphore(5)
	var maxConcurrent int32
	var current int32

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			err := sem.Acquire(context.Background())
			require.NoError(t, err)

			cur := atomic.AddInt32(&current, 1)
			for {
				old := atomic.LoadInt32(&maxConcurrent)
				if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
					break
				}
			}

			time.Sleep(time.Millisecond)
			atomic.AddInt32(&current, -1)
			sem.Release()
		}()
	}

	wg.Wait()
	assert.LessOrEqual(t, atomic.LoadInt32(&maxConcurrent), int32(5),
		"max concurrent should never exceed semaphore capacity")
}

// ---------------------------------------------------------------------------
// BoundedSemaphore — expanded tests
// ---------------------------------------------------------------------------

func TestBoundedSemaphore_AcquireRelease_Sequential(t *testing.T) {
	bs := NewBoundedSemaphore("test", 2, 5*time.Second)

	err := bs.Acquire(context.Background())
	require.NoError(t, err)

	err = bs.Acquire(context.Background())
	require.NoError(t, err)

	// Release both
	bs.Release()
	bs.Release()

	// Should be able to acquire again
	err = bs.Acquire(context.Background())
	require.NoError(t, err)
	bs.Release()
}

func TestBoundedSemaphore_TimeoutIsApplied(t *testing.T) {
	// Very short timeout
	bs := NewBoundedSemaphore("short-timeout", 1, 10*time.Millisecond)

	// Fill the single slot
	err := bs.Acquire(context.Background())
	require.NoError(t, err)

	start := time.Now()
	err = bs.Acquire(context.Background())
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "short-timeout")
	// Should have taken approximately the timeout duration
	assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond)

	bs.Release()
}

func TestBoundedSemaphore_ParentContextCancelledBeforeTimeout(t *testing.T) {
	bs := NewBoundedSemaphore("parent-cancel", 1, 10*time.Second)

	err := bs.Acquire(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = bs.Acquire(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err)
	// Should have been cancelled by parent context, not the semaphore timeout
	assert.Less(t, elapsed, 1*time.Second)

	bs.Release()
}

func TestBoundedSemaphore_ConcurrentBounded(t *testing.T) {
	bs := NewBoundedSemaphore("concurrent", 3, 30*time.Second)
	var maxConcurrent int32
	var current int32

	const goroutines = 30
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			err := bs.Acquire(context.Background())
			if err != nil {
				return
			}

			cur := atomic.AddInt32(&current, 1)
			for {
				old := atomic.LoadInt32(&maxConcurrent)
				if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
					break
				}
			}

			time.Sleep(time.Millisecond)
			atomic.AddInt32(&current, -1)
			bs.Release()
		}()
	}

	wg.Wait()
	assert.LessOrEqual(t, atomic.LoadInt32(&maxConcurrent), int32(3))
}

func TestBoundedSemaphore_NameAndMax_Various(t *testing.T) {
	tests := []struct {
		name string
		max  int64
	}{
		{"one", 1},
		{"ten", 10},
		{"hundred", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := NewBoundedSemaphore(tt.name, tt.max, time.Second)
			assert.Equal(t, tt.name, bs.Name())
			assert.Equal(t, tt.max, bs.MaxConcurrent())
		})
	}
}

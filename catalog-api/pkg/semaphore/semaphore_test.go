package semaphore

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s := New(5)
	assert.NotNil(t, s)
	assert.Equal(t, 5, s.Capacity())
	assert.Equal(t, 5, s.Available())
	assert.Equal(t, 0, s.Acquired())
}

func TestAcquireAndRelease(t *testing.T) {
	s := New(2)

	err := s.Acquire(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, s.Acquired())
	assert.Equal(t, 1, s.Available())

	err = s.Acquire(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, s.Acquired())
	assert.Equal(t, 0, s.Available())

	s.Release()
	assert.Equal(t, 1, s.Acquired())
	assert.Equal(t, 1, s.Available())

	s.Release()
	assert.Equal(t, 0, s.Acquired())
	assert.Equal(t, 2, s.Available())
}

func TestTryAcquire(t *testing.T) {
	s := New(2)

	assert.True(t, s.TryAcquire())
	assert.True(t, s.TryAcquire())
	assert.False(t, s.TryAcquire())

	s.Release()
	assert.True(t, s.TryAcquire())
}

func TestAcquireWithContextCancellation(t *testing.T) {
	s := New(1)

	require.NoError(t, s.Acquire(context.Background()))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := s.Acquire(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestAcquireWithCancelledContext(t *testing.T) {
	s := New(1)

	require.NoError(t, s.Acquire(context.Background()))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.Acquire(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	s.Release()
}

func TestClose(t *testing.T) {
	s := New(2)

	require.NoError(t, s.Acquire(context.Background()))

	s.Close()

	err := s.Acquire(context.Background())
	assert.Error(t, err)
	assert.Equal(t, ErrSemaphoreClosed, err)

	assert.False(t, s.TryAcquire())

	s.Release()

	assert.Equal(t, 0, s.Available())
	assert.Equal(t, 0, s.Acquired())
}

func TestDoubleClose(t *testing.T) {
	s := New(2)
	s.Close()
	s.Close()
}

func TestConcurrentAcquire(t *testing.T) {
	s := New(3)
	var wg sync.WaitGroup
	acquired := make(chan struct{}, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.TryAcquire() {
				acquired <- struct{}{}
				time.Sleep(10 * time.Millisecond)
				s.Release()
			}
		}()
	}

	wg.Wait()
	close(acquired)

	count := 0
	for range acquired {
		count++
	}

	assert.Equal(t, 3, count)
}

func TestBlockingAcquire(t *testing.T) {
	s := New(1)

	require.NoError(t, s.Acquire(context.Background()))

	done := make(chan error, 1)
	go func() {
		done <- s.Acquire(context.Background())
	}()

	time.Sleep(50 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("Acquire should have blocked")
	default:
	}

	s.Release()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Acquire should have completed after Release")
	}
}

func TestReleaseOnEmpty(t *testing.T) {
	s := New(2)

	s.Release()
	s.Release()
	s.Release()

	assert.Equal(t, 2, s.Available())
}

func TestSemaphoreZero(t *testing.T) {
	s := New(0)

	assert.False(t, s.TryAcquire())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := s.Acquire(ctx)
	assert.Error(t, err)
}

func TestAvailableAfterClose(t *testing.T) {
	s := New(5)
	s.Close()

	assert.Equal(t, 0, s.Available())
	assert.Equal(t, 0, s.Acquired())
	assert.Equal(t, 5, s.Capacity())
}

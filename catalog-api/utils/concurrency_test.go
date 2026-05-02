package utils

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// WorkerPool Tests
// =============================================================================

func TestNewWorkerPool_DefaultValues(t *testing.T) {
	pool := NewWorkerPool(0, 0)
	defer pool.Stop()

	stats := pool.GetStats()
	assert.Greater(t, stats.Workers, 0, "default workers should be > 0")
	assert.Greater(t, stats.QueueCap, 0, "default queue capacity should be > 0")
	assert.False(t, stats.IsStopped)
}

func TestNewWorkerPool_CustomValues(t *testing.T) {
	pool := NewWorkerPool(4, 10)
	defer pool.Stop()

	stats := pool.GetStats()
	assert.Equal(t, 4, stats.Workers)
	assert.Equal(t, 10, stats.QueueCap)
}

func TestWorkerPool_Submit_Success(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	defer pool.Stop()

	var executed atomic.Bool
	done := make(chan struct{})

	ok := pool.Submit(func() {
		executed.Store(true)
		close(done)
	})

	assert.True(t, ok)

	select {
	case <-done:
		assert.True(t, executed.Load())
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for job execution")
	}
}

func TestWorkerPool_Submit_AfterStop(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Stop()

	ok := pool.Submit(func() {})
	assert.False(t, ok, "submit should fail after stop")
}

func TestWorkerPool_SubmitAsync(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	defer pool.Stop()

	var executed atomic.Bool
	done := make(chan struct{})

	pool.SubmitAsync(func() {
		executed.Store(true)
		close(done)
	})

	select {
	case <-done:
		assert.True(t, executed.Load())
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for async job")
	}
}

func TestWorkerPool_SubmitAsync_AfterStop(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Stop()

	// Should not panic
	pool.SubmitAsync(func() {})
}

func TestWorkerPool_Submit_MultipleJobs(t *testing.T) {
	pool := NewWorkerPool(4, 20)
	defer pool.Stop()

	var counter atomic.Int32
	var wg sync.WaitGroup

	jobCount := 10
	wg.Add(jobCount)

	for i := 0; i < jobCount; i++ {
		ok := pool.Submit(func() {
			counter.Add(1)
			wg.Done()
		})
		assert.True(t, ok)
	}

	wg.Wait()
	assert.Equal(t, int32(jobCount), counter.Load())
}

func TestWorkerPool_Stop_Idempotent(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Stop()
	pool.Stop() // second call should be safe

	stats := pool.GetStats()
	assert.True(t, stats.IsStopped)
}

func TestWorkerPool_StopTimeout_Success(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	ok := pool.StopTimeout(2 * time.Second)
	assert.True(t, ok)
}

func TestWorkerPool_StopTimeout_AlreadyStopped(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Stop()

	ok := pool.StopTimeout(1 * time.Second)
	assert.True(t, ok)
}

func TestWorkerPool_GetStats(t *testing.T) {
	pool := NewWorkerPool(3, 8)
	defer pool.Stop()

	stats := pool.GetStats()
	assert.Equal(t, 3, stats.Workers)
	assert.Equal(t, 8, stats.QueueCap)
	assert.Equal(t, 0, stats.QueueSize)
	assert.False(t, stats.IsStopped)
}

// =============================================================================
// Throttler Tests
// =============================================================================

func TestNewThrottler(t *testing.T) {
	throttler := NewThrottler(50 * time.Millisecond)
	defer throttler.Stop()

	ok := throttler.Allow()
	assert.True(t, ok)
}

func TestThrottler_Allow_AfterStop(t *testing.T) {
	throttler := NewThrottler(50 * time.Millisecond)
	throttler.Stop()

	ok := throttler.Allow()
	assert.False(t, ok)
}

func TestThrottler_AllowTimeout_Success(t *testing.T) {
	throttler := NewThrottler(50 * time.Millisecond)
	defer throttler.Stop()

	ok := throttler.AllowTimeout(2 * time.Second)
	assert.True(t, ok)
}

func TestThrottler_AllowTimeout_Expired(t *testing.T) {
	// Create a very slow throttler
	throttler := NewThrottler(10 * time.Second)
	defer throttler.Stop()

	// Drain any initial token
	_ = throttler.AllowTimeout(100 * time.Millisecond)

	// Next one should timeout
	ok := throttler.AllowTimeout(50 * time.Millisecond)
	assert.False(t, ok)
}

func TestThrottler_AllowTimeout_AfterStop(t *testing.T) {
	throttler := NewThrottler(50 * time.Millisecond)
	throttler.Stop()

	ok := throttler.AllowTimeout(1 * time.Second)
	assert.False(t, ok)
}

func TestThrottler_Stop_Idempotent(t *testing.T) {
	throttler := NewThrottler(50 * time.Millisecond)
	throttler.Stop()
	throttler.Stop() // should not panic
}

// =============================================================================
// Debouncer Tests
// =============================================================================

func TestDebouncer_Debounce_ExecutesAfterDelay(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)

	var executed atomic.Bool
	done := make(chan struct{})

	d.Debounce(func() {
		executed.Store(true)
		close(done)
	})

	select {
	case <-done:
		assert.True(t, executed.Load())
	case <-time.After(2 * time.Second):
		t.Fatal("debounced function was not called")
	}
}

func TestDebouncer_Debounce_OnlyLastCallExecutes(t *testing.T) {
	d := NewDebouncer(100 * time.Millisecond)

	var callCount atomic.Int32
	done := make(chan struct{})

	// Rapid-fire debounce calls
	for i := 0; i < 5; i++ {
		d.Debounce(func() {
			callCount.Add(1)
			close(done)
		})
		time.Sleep(10 * time.Millisecond) // shorter than debounce delay
	}

	select {
	case <-done:
		assert.Equal(t, int32(1), callCount.Load(), "only the last call should execute")
	case <-time.After(2 * time.Second):
		t.Fatal("debounced function was not called")
	}
}

func TestDebouncer_Flush(t *testing.T) {
	d := NewDebouncer(10 * time.Second) // long delay

	var executed atomic.Bool

	d.Debounce(func() {
		executed.Store(true)
	})

	// Flush should execute immediately
	d.Flush()
	assert.True(t, executed.Load())
}

func TestDebouncer_Flush_NoPending(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)
	// Flush with nothing pending should not panic
	d.Flush()
}

// =============================================================================
// CircuitBreaker State Tests
// =============================================================================

func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

// =============================================================================
// CircuitBreaker Tests
// =============================================================================

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)
	assert.NotNil(t, cb)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_Allow_Closed(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)
	assert.True(t, cb.Allow())
}

func TestCircuitBreaker_TransitionToOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 1*time.Second)

	// Record enough failures to trip
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	assert.Equal(t, StateOpen, cb.GetState())
	assert.False(t, cb.Allow(), "open circuit should deny requests")
}

func TestCircuitBreaker_TransitionToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 50*time.Millisecond)

	// Trip the circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for timeout to expire
	time.Sleep(100 * time.Millisecond)

	// Should transition to half-open
	assert.True(t, cb.Allow())
	assert.Equal(t, StateHalfOpen, cb.GetState())
}

func TestCircuitBreaker_HalfOpen_ToClose(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 50*time.Millisecond)

	// Trip the circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	time.Sleep(100 * time.Millisecond)
	cb.Allow() // transitions to half-open

	// Record enough successes
	cb.RecordSuccess()
	cb.RecordSuccess()

	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_HalfOpen_BackToOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 50*time.Millisecond)

	// Trip the circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	time.Sleep(100 * time.Millisecond)
	cb.Allow() // transitions to half-open

	// Fail in half-open goes back to open
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_RecordSuccess_Closed_ResetsFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)

	// Record some failures (but not enough to trip)
	cb.RecordFailure()
	cb.RecordFailure()

	// Success resets failure counter
	cb.RecordSuccess()

	// Now it should take 3 more failures to trip
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, StateClosed, cb.GetState())

	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)
	cb.RecordFailure()

	stats := cb.GetStats()
	assert.Equal(t, "closed", stats["state"])
	assert.Equal(t, 1, stats["failures"])
	assert.Equal(t, 0, stats["successes"])
	assert.NotNil(t, stats["last_failure"])
}

func TestCircuitBreaker_Allow_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 50*time.Millisecond)

	// Trip the circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	time.Sleep(100 * time.Millisecond)
	cb.Allow() // transitions to half-open

	// Half-open should allow
	assert.True(t, cb.Allow())
}

// =============================================================================
// Retry Tests
// =============================================================================

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.Delay)
	assert.Equal(t, 5*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.NotNil(t, config.ShouldRetry)
	assert.True(t, config.ShouldRetry(errors.New("test")))
	assert.False(t, config.ShouldRetry(nil))
}

func TestRetry_Success_FirstAttempt(t *testing.T) {
	config := RetryConfig{
		MaxRetries:  3,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return true },
	}

	callCount := 0
	err := Retry(config, func() error {
		callCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetry_Success_AfterRetries(t *testing.T) {
	config := RetryConfig{
		MaxRetries:  3,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return true },
	}

	callCount := 0
	err := Retry(config, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetry_Exhausted(t *testing.T) {
	config := RetryConfig{
		MaxRetries:  2,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return true },
	}

	callCount := 0
	err := Retry(config, func() error {
		callCount++
		return errors.New("persistent error")
	})

	assert.Error(t, err)
	assert.Equal(t, "persistent error", err.Error())
	assert.Equal(t, 3, callCount) // initial + 2 retries
}

func TestRetry_ShouldRetryFalse(t *testing.T) {
	config := RetryConfig{
		MaxRetries:  3,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return false },
	}

	callCount := 0
	err := Retry(config, func() error {
		callCount++
		return errors.New("non-retryable")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetry_DelayCapAtMaxDelay(t *testing.T) {
	config := RetryConfig{
		MaxRetries:  5,
		Delay:       1 * time.Millisecond,
		MaxDelay:    3 * time.Millisecond,
		Multiplier:  10.0,
		ShouldRetry: func(err error) bool { return true },
	}

	start := time.Now()
	_ = Retry(config, func() error {
		return errors.New("error")
	})
	elapsed := time.Since(start)

	// With 5 retries at max 3ms each, should be under 50ms
	assert.Less(t, elapsed, 100*time.Millisecond)
}

// =============================================================================
// RetryWithContext Tests
// =============================================================================

func TestRetryWithContext_Success(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:  3,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return true },
	}

	callCount := 0
	err := RetryWithContext(ctx, config, func() error {
		callCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetryWithContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := RetryConfig{
		MaxRetries:  10,
		Delay:       50 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  1.0,
		ShouldRetry: func(err error) bool { return true },
	}

	callCount := 0
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := RetryWithContext(ctx, config, func() error {
		callCount++
		return errors.New("keep trying")
	})

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestRetryWithContext_ShouldRetryFalse(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:  3,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return false },
	}

	callCount := 0
	err := RetryWithContext(ctx, config, func() error {
		callCount++
		return errors.New("non-retryable")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetryWithContext_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:  5,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return true },
	}

	callCount := 0
	err := RetryWithContext(ctx, config, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetryWithContext_Exhausted(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:  2,
		Delay:       1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: func(err error) bool { return true },
	}

	err := RetryWithContext(ctx, config, func() error {
		return errors.New("persistent")
	})

	assert.Error(t, err)
	assert.Equal(t, "persistent", err.Error())
}

package recovery

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

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.True(t, config.Jitter)
}

func TestRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"retryable error", errors.New("timeout"), true},
		{"non-retryable error", errors.New("auth failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRetryableError(tt.err, tt.retryable)
			assert.Equal(t, tt.err.Error(), re.Error())
			assert.Equal(t, tt.retryable, re.IsRetryable())
		})
	}
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	callCount := 0
	err := Retry(context.Background(), config, func() error {
		callCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetry_SuccessOnSecondAttempt(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	callCount := 0
	err := Retry(context.Background(), config, func() error {
		callCount++
		if callCount < 2 {
			return errors.New("transient failure")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	callCount := 0
	expectedErr := errors.New("persistent failure")
	err := Retry(context.Background(), config, func() error {
		callCount++
		return expectedErr
	})

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 3, callCount)
}

func TestRetry_NonRetryableErrorStopsEarly(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	callCount := 0
	err := Retry(context.Background(), config, func() error {
		callCount++
		return NewRetryableError(errors.New("fatal error"), false)
	})

	assert.Error(t, err)
	assert.Equal(t, 1, callCount) // Should stop after first attempt
}

func TestRetry_ContextCancellation(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   10,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	callCount := 0
	err := Retry(ctx, config, func() error {
		callCount++
		return errors.New("keep failing")
	})

	assert.Error(t, err)
	// Should have been cancelled by context before all 10 attempts
	assert.Less(t, callCount, 10)
}

func TestRetry_ContextAlreadyCancelled(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	callCount := 0
	err := Retry(ctx, config, func() error {
		callCount++
		return errors.New("fail")
	})

	assert.Error(t, err)
	// First attempt runs, then context is checked during delay
	assert.Equal(t, 1, callCount)
	assert.Equal(t, context.Canceled, err)
}

func TestRetryWithCallbacks_OnRetryCallback(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	var retryAttempts []int
	var retryErrors []error

	err := RetryWithCallbacks(
		context.Background(),
		config,
		func(attempt int) error {
			return errors.New("fail")
		},
		func(attempt int, err error, delay time.Duration) {
			retryAttempts = append(retryAttempts, attempt)
			retryErrors = append(retryErrors, err)
		},
		nil,
	)

	assert.Error(t, err)
	// onRetry is called for attempts 1 and 2 (not for the last attempt)
	assert.Equal(t, []int{1, 2}, retryAttempts)
	assert.Len(t, retryErrors, 2)
}

func TestRetryWithCallbacks_OnFinalErrorCallback(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	var finalAttempt int
	var finalErr error

	err := RetryWithCallbacks(
		context.Background(),
		config,
		func(attempt int) error {
			return errors.New("fail")
		},
		nil,
		func(attempt int, err error) {
			finalAttempt = attempt
			finalErr = err
		},
	)

	assert.Error(t, err)
	assert.Equal(t, 2, finalAttempt)
	assert.EqualError(t, finalErr, "fail")
}

func TestRetryWithCallbacks_SuccessNoFinalError(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	}

	finalCalled := false
	err := RetryWithCallbacks(
		context.Background(),
		config,
		func(attempt int) error { return nil },
		nil,
		func(attempt int, err error) { finalCalled = true },
	)

	assert.NoError(t, err)
	assert.False(t, finalCalled)
}

func TestCalculateDelay(t *testing.T) {
	tests := []struct {
		name    string
		config  RetryConfig
		attempt int
		minWant time.Duration
		maxWant time.Duration
	}{
		{
			name: "first attempt no jitter",
			config: RetryConfig{
				InitialDelay:  100 * time.Millisecond,
				BackoffFactor: 2.0,
				MaxDelay:      10 * time.Second,
				Jitter:        false,
			},
			attempt: 0,
			minWant: 100 * time.Millisecond,
			maxWant: 100 * time.Millisecond,
		},
		{
			name: "second attempt no jitter",
			config: RetryConfig{
				InitialDelay:  100 * time.Millisecond,
				BackoffFactor: 2.0,
				MaxDelay:      10 * time.Second,
				Jitter:        false,
			},
			attempt: 1,
			minWant: 200 * time.Millisecond,
			maxWant: 200 * time.Millisecond,
		},
		{
			name: "third attempt no jitter",
			config: RetryConfig{
				InitialDelay:  100 * time.Millisecond,
				BackoffFactor: 2.0,
				MaxDelay:      10 * time.Second,
				Jitter:        false,
			},
			attempt: 2,
			minWant: 400 * time.Millisecond,
			maxWant: 400 * time.Millisecond,
		},
		{
			name: "capped at max delay",
			config: RetryConfig{
				InitialDelay:  100 * time.Millisecond,
				BackoffFactor: 2.0,
				MaxDelay:      300 * time.Millisecond,
				Jitter:        false,
			},
			attempt: 5, // 100ms * 2^5 = 3200ms > 300ms
			minWant: 300 * time.Millisecond,
			maxWant: 300 * time.Millisecond,
		},
		{
			name: "with jitter adds up to 10%",
			config: RetryConfig{
				InitialDelay:  100 * time.Millisecond,
				BackoffFactor: 2.0,
				MaxDelay:      10 * time.Second,
				Jitter:        true,
			},
			attempt: 0,
			minWant: 100 * time.Millisecond,
			maxWant: 110 * time.Millisecond, // 100ms + 10% jitter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateDelay(tt.config, tt.attempt)
			assert.GreaterOrEqual(t, delay, tt.minWant, "delay should be >= min")
			assert.LessOrEqual(t, delay, tt.maxWant, "delay should be <= max")
		})
	}
}

func TestCalculateDelay_JitterDistribution(t *testing.T) {
	config := RetryConfig{
		InitialDelay:  100 * time.Millisecond,
		BackoffFactor: 2.0,
		MaxDelay:      10 * time.Second,
		Jitter:        true,
	}

	// Run multiple times to verify jitter produces varied results
	seen := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		delay := calculateDelay(config, 0)
		seen[delay] = true
		assert.GreaterOrEqual(t, delay, 100*time.Millisecond)
		assert.LessOrEqual(t, delay, 110*time.Millisecond)
	}
	// With 100 samples and random jitter, we should see more than 1 unique value
	assert.Greater(t, len(seen), 1, "jitter should produce varied delays")
}

// ExponentialBackoff tests

func TestNewExponentialBackoff_Defaults(t *testing.T) {
	tests := []struct {
		name   string
		config RetryConfig
	}{
		{
			name:   "zero config gets defaults",
			config: RetryConfig{Logger: newTestLogger()},
		},
		{
			name: "negative values get defaults",
			config: RetryConfig{
				MaxAttempts:   -1,
				InitialDelay:  -1,
				MaxDelay:      -1,
				BackoffFactor: -1,
				Logger:        newTestLogger(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := NewExponentialBackoff(tt.config)
			assert.NotNil(t, eb)
			assert.Equal(t, 3, eb.config.MaxAttempts)
			assert.Equal(t, 1*time.Second, eb.config.InitialDelay)
			assert.Equal(t, 30*time.Second, eb.config.MaxDelay)
			assert.Equal(t, 2.0, eb.config.BackoffFactor)
		})
	}
}

func TestExponentialBackoff_Execute(t *testing.T) {
	eb := NewExponentialBackoff(RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	})

	callCount := 0
	err := eb.Execute(context.Background(), func() error {
		callCount++
		if callCount < 2 {
			return errors.New("transient")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestExponentialBackoff_ExecuteWithCallbacks(t *testing.T) {
	eb := NewExponentialBackoff(RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	})

	var retries []int
	var finalCalled bool

	err := eb.ExecuteWithCallbacks(
		context.Background(),
		func(attempt int) error {
			return errors.New("fail")
		},
		func(attempt int, err error, delay time.Duration) {
			retries = append(retries, attempt)
		},
		func(attempt int, err error) {
			finalCalled = true
		},
	)

	assert.Error(t, err)
	assert.Equal(t, []int{1, 2}, retries)
	assert.True(t, finalCalled)
}

func TestExponentialBackoff_ContextCancellation(t *testing.T) {
	eb := NewExponentialBackoff(RetryConfig{
		MaxAttempts:   10,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Logger:        newTestLogger(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := eb.Execute(ctx, func() error {
		return errors.New("fail")
	})

	assert.Error(t, err)
}

// Bulkhead tests

func TestNewBulkhead_Defaults(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{Logger: newTestLogger()})
	stats := b.GetStats()
	assert.Equal(t, 10, stats["max_concurrent"])
	assert.Equal(t, 100, stats["queue_size"])
	assert.Equal(t, 30*time.Second, stats["timeout"])
}

func TestNewBulkhead_CustomConfig(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 5,
		QueueSize:     50,
		Timeout:       10 * time.Second,
		Logger:        newTestLogger(),
	})
	stats := b.GetStats()
	assert.Equal(t, 5, stats["max_concurrent"])
	assert.Equal(t, 50, stats["queue_size"])
	assert.Equal(t, 10*time.Second, stats["timeout"])
}

func TestBulkhead_Execute_Success(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 2,
		Timeout:       time.Second,
		Logger:        newTestLogger(),
	})

	err := b.Execute(context.Background(), func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestBulkhead_Execute_ReturnsError(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 2,
		Timeout:       time.Second,
		Logger:        newTestLogger(),
	})

	expectedErr := errors.New("operation failed")
	err := b.Execute(context.Background(), func() error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
}

func TestBulkhead_ConcurrentLimit(t *testing.T) {
	maxConcurrent := 3
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: maxConcurrent,
		Timeout:       5 * time.Second,
		Logger:        newTestLogger(),
	})

	var currentConcurrent int64
	var maxObserved int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Execute(context.Background(), func() error {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				defer atomic.AddInt64(&currentConcurrent, -1)

				mu.Lock()
				if cur > maxObserved {
					maxObserved = cur
				}
				mu.Unlock()

				time.Sleep(50 * time.Millisecond)
				return nil
			})
		}()
	}

	wg.Wait()
	assert.LessOrEqual(t, maxObserved, int64(maxConcurrent),
		"concurrent executions should not exceed max")
}

func TestBulkhead_PermitReturnedAfterExecution(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 1,
		Timeout:       time.Second,
		Logger:        newTestLogger(),
	})

	// Execute and complete
	err := b.Execute(context.Background(), func() error { return nil })
	assert.NoError(t, err)

	// Should be able to execute again (permit returned)
	err = b.Execute(context.Background(), func() error { return nil })
	assert.NoError(t, err)
}

func TestBulkhead_PermitReturnedAfterError(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 1,
		Timeout:       time.Second,
		Logger:        newTestLogger(),
	})

	// Execute with error
	_ = b.Execute(context.Background(), func() error {
		return errors.New("fail")
	})

	// Permit should still be returned
	err := b.Execute(context.Background(), func() error { return nil })
	assert.NoError(t, err)
}

func TestBulkhead_ContextCancellation(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 1,
		Timeout:       5 * time.Second,
		Logger:        newTestLogger(),
	})

	// Acquire the only permit
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		_ = b.Execute(context.Background(), func() error {
			close(started)
			<-done // Hold the permit
			return nil
		})
	}()
	<-started

	// Try to execute with an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := b.Execute(ctx, func() error { return nil })
	assert.Equal(t, context.Canceled, err)

	close(done) // Release the permit
}

func TestBulkhead_Timeout(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 1,
		Timeout:       50 * time.Millisecond,
		Logger:        newTestLogger(),
	})

	// Acquire the only permit
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		_ = b.Execute(context.Background(), func() error {
			close(started)
			<-done
			return nil
		})
	}()
	<-started

	// Should timeout waiting for permit
	err := b.Execute(context.Background(), func() error { return nil })
	assert.Error(t, err)

	// The error should be a RetryableError wrapping DeadlineExceeded
	var retryableErr RetryableError
	if errors.As(err, &retryableErr) {
		assert.True(t, retryableErr.IsRetryable())
		assert.Equal(t, context.DeadlineExceeded, retryableErr.Err)
	}

	close(done)
}

func TestBulkhead_GetStats(t *testing.T) {
	b := NewBulkhead(BulkheadConfig{
		MaxConcurrent: 5,
		QueueSize:     20,
		Timeout:       3 * time.Second,
		Logger:        newTestLogger(),
	})

	stats := b.GetStats()
	assert.Equal(t, 5, stats["max_concurrent"])
	assert.Equal(t, 5, stats["available_permits"]) // All permits available
	assert.Equal(t, 20, stats["queue_size"])
	assert.Equal(t, 0, stats["queue_length"])
	assert.Equal(t, 3*time.Second, stats["timeout"])
}

// HealthChecker tests

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())
	assert.NotNil(t, hc)

	status := hc.CheckHealth(context.Background())
	assert.True(t, status.Healthy)
	assert.Empty(t, status.Checks)
}

func TestHealthChecker_AddAndRemoveCheck(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:     "test-check",
		Check:    func(ctx context.Context) error { return nil },
		Critical: true,
	})

	status := hc.CheckHealth(context.Background())
	assert.Len(t, status.Checks, 1)
	assert.Contains(t, status.Checks, "test-check")

	hc.RemoveCheck("test-check")
	status = hc.CheckHealth(context.Background())
	assert.Empty(t, status.Checks)
}

func TestHealthChecker_HealthyCheck(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:     "db",
		Check:    func(ctx context.Context) error { return nil },
		Critical: true,
	})

	status := hc.CheckHealth(context.Background())
	assert.True(t, status.Healthy)
	result := status.Checks["db"]
	assert.True(t, result.Healthy)
	assert.Empty(t, result.Error)
	assert.True(t, result.Critical)
	assert.GreaterOrEqual(t, result.Duration, int64(0))
	assert.False(t, result.Timestamp.IsZero())
}

func TestHealthChecker_UnhealthyCriticalCheck(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:     "db",
		Check:    func(ctx context.Context) error { return errors.New("connection refused") },
		Critical: true,
	})

	status := hc.CheckHealth(context.Background())
	assert.False(t, status.Healthy, "overall health should be false when critical check fails")
	result := status.Checks["db"]
	assert.False(t, result.Healthy)
	assert.Equal(t, "connection refused", result.Error)
}

func TestHealthChecker_UnhealthyNonCriticalCheck(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:     "cache",
		Check:    func(ctx context.Context) error { return errors.New("cache unavailable") },
		Critical: false,
	})

	status := hc.CheckHealth(context.Background())
	assert.True(t, status.Healthy, "overall health should be true when only non-critical checks fail")
	result := status.Checks["cache"]
	assert.False(t, result.Healthy)
	assert.Equal(t, "cache unavailable", result.Error)
}

func TestHealthChecker_MixedChecks(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:     "db",
		Check:    func(ctx context.Context) error { return nil },
		Critical: true,
	})
	hc.AddCheck(HealthCheck{
		Name:     "cache",
		Check:    func(ctx context.Context) error { return errors.New("down") },
		Critical: false,
	})
	hc.AddCheck(HealthCheck{
		Name:     "queue",
		Check:    func(ctx context.Context) error { return nil },
		Critical: true,
	})

	status := hc.CheckHealth(context.Background())
	assert.True(t, status.Healthy)
	assert.Len(t, status.Checks, 3)
	assert.True(t, status.Checks["db"].Healthy)
	assert.False(t, status.Checks["cache"].Healthy)
	assert.True(t, status.Checks["queue"].Healthy)
}

func TestHealthChecker_AllCriticalFailing(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:     "db",
		Check:    func(ctx context.Context) error { return errors.New("down") },
		Critical: true,
	})
	hc.AddCheck(HealthCheck{
		Name:     "api",
		Check:    func(ctx context.Context) error { return errors.New("down") },
		Critical: true,
	})

	status := hc.CheckHealth(context.Background())
	assert.False(t, status.Healthy)
}

func TestHealthChecker_CheckTimeout(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 50*time.Millisecond, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name: "slow-check",
		Check: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		},
		Critical: true,
		// Timeout will be set to the default (50ms) since it's 0
	})

	status := hc.CheckHealth(context.Background())
	assert.False(t, status.Healthy)
	result := status.Checks["slow-check"]
	assert.False(t, result.Healthy)
	assert.Contains(t, result.Error, "context deadline exceeded")
}

func TestHealthChecker_CustomCheckTimeout(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name: "custom-timeout",
		Check: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		},
		Timeout:  50 * time.Millisecond,
		Critical: true,
	})

	status := hc.CheckHealth(context.Background())
	assert.False(t, status.Healthy)
	result := status.Checks["custom-timeout"]
	assert.False(t, result.Healthy)
}

func TestHealthChecker_ContextCancellation(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name: "check",
		Check: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
		Critical: true,
		Timeout:  time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	status := hc.CheckHealth(ctx)
	assert.False(t, status.Healthy)
}

func TestHealthChecker_DefaultTimeout(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 2*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:  "no-timeout-specified",
		Check: func(ctx context.Context) error { return nil },
		// Timeout is 0, should default to hc.timeout (2s)
	})

	// This verifies that AddCheck sets the timeout correctly
	hc.mutex.RLock()
	check := hc.checks["no-timeout-specified"]
	hc.mutex.RUnlock()
	require.Equal(t, 2*time.Second, check.Timeout)
}

func TestHealthChecker_ConcurrentCheckHealth(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, newTestLogger())

	hc.AddCheck(HealthCheck{
		Name:     "concurrent-safe",
		Check:    func(ctx context.Context) error { return nil },
		Critical: true,
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status := hc.CheckHealth(context.Background())
			assert.True(t, status.Healthy)
		}()
	}
	wg.Wait()
}

func TestRetry_NilLogger(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:   2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Logger:        nil, // no logger
	}

	callCount := 0
	err := Retry(context.Background(), config, func() error {
		callCount++
		if callCount < 2 {
			return errors.New("fail")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestHealthChecker_NilLogger(t *testing.T) {
	hc := NewHealthChecker(30*time.Second, 5*time.Second, nil)

	hc.AddCheck(HealthCheck{
		Name:     "test",
		Check:    func(ctx context.Context) error { return nil },
		Critical: true,
	})

	// Should not panic with nil logger
	status := hc.CheckHealth(context.Background())
	assert.True(t, status.Healthy)
}

package recovery

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RetryConfig contains configuration for retry logic
type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	Jitter        bool
	Logger        *zap.Logger
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e RetryableError) Error() string {
	return e.Err.Error()
}

// IsRetryable returns true if the error is retryable
func (e RetryableError) IsRetryable() bool {
	return e.Retryable
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error, retryable bool) RetryableError {
	return RetryableError{
		Err:       err,
		Retryable: retryable,
	}
}

// RetryFunc is a function that can be retried
type RetryFunc func() error

// RetryWithCallback is a function that can be retried with callbacks
type RetryWithCallback func(attempt int) error

// Retry executes a function with retry logic
func Retry(ctx context.Context, config RetryConfig, fn RetryFunc) error {
	return RetryWithCallbacks(ctx, config, func(attempt int) error {
		return fn()
	}, nil, nil)
}

// RetryWithCallbacks executes a function with retry logic and callbacks
func RetryWithCallbacks(
	ctx context.Context,
	config RetryConfig,
	fn RetryWithCallback,
	onRetry func(attempt int, err error, delay time.Duration),
	onFinalError func(attempt int, err error),
) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if config.Logger != nil {
			config.Logger.Debug("Retry attempt",
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", config.MaxAttempts))
		}

		err := fn(attempt)
		if err == nil {
			if config.Logger != nil && attempt > 0 {
				config.Logger.Info("Operation succeeded after retry",
					zap.Int("attempts", attempt+1))
			}
			return nil
		}

		lastErr = err

		// Check if the error is retryable
		if retryableErr, ok := err.(RetryableError); ok && !retryableErr.IsRetryable() {
			if config.Logger != nil {
				config.Logger.Debug("Error is not retryable", zap.Error(err))
			}
			break
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay
		delay := calculateDelay(config, attempt)

		if onRetry != nil {
			onRetry(attempt+1, err, delay)
		}

		if config.Logger != nil {
			config.Logger.Warn("Operation failed, retrying",
				zap.Error(err),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay))
		}

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	if onFinalError != nil {
		onFinalError(config.MaxAttempts, lastErr)
	}

	if config.Logger != nil {
		config.Logger.Error("All retry attempts failed",
			zap.Int("attempts", config.MaxAttempts),
			zap.Error(lastErr))
	}

	return lastErr
}

// calculateDelay calculates the delay for the next retry attempt
func calculateDelay(config RetryConfig, attempt int) time.Duration {
	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt))

	// Apply maximum delay limit
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Apply jitter if enabled
	if config.Jitter {
		jitter := rand.Float64() * 0.1 * delay // 10% jitter
		delay += jitter
	}

	return time.Duration(delay)
}

// ExponentialBackoff implements exponential backoff retry strategy
type ExponentialBackoff struct {
	config RetryConfig
	logger *zap.Logger
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(config RetryConfig) *ExponentialBackoff {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3
	}
	if config.InitialDelay <= 0 {
		config.InitialDelay = 1 * time.Second
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.BackoffFactor <= 0 {
		config.BackoffFactor = 2.0
	}

	return &ExponentialBackoff{
		config: config,
		logger: config.Logger,
	}
}

// Execute executes a function with exponential backoff
func (eb *ExponentialBackoff) Execute(ctx context.Context, fn RetryFunc) error {
	return Retry(ctx, eb.config, fn)
}

// ExecuteWithCallbacks executes a function with exponential backoff and callbacks
func (eb *ExponentialBackoff) ExecuteWithCallbacks(
	ctx context.Context,
	fn RetryWithCallback,
	onRetry func(attempt int, err error, delay time.Duration),
	onFinalError func(attempt int, err error),
) error {
	return RetryWithCallbacks(ctx, eb.config, fn, onRetry, onFinalError)
}

// BulkheadConfig contains configuration for bulkhead pattern
type BulkheadConfig struct {
	MaxConcurrent int
	QueueSize     int
	Timeout       time.Duration
	Logger        *zap.Logger
}

// Bulkhead implements the bulkhead pattern for resource isolation
type Bulkhead struct {
	semaphore chan struct{}
	queue     chan func()
	config    BulkheadConfig
	logger    *zap.Logger
}

// NewBulkhead creates a new bulkhead
func NewBulkhead(config BulkheadConfig) *Bulkhead {
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 10
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 100
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	bulkhead := &Bulkhead{
		semaphore: make(chan struct{}, config.MaxConcurrent),
		queue:     make(chan func(), config.QueueSize),
		config:    config,
		logger:    config.Logger,
	}

	// Pre-fill semaphore
	for i := 0; i < config.MaxConcurrent; i++ {
		bulkhead.semaphore <- struct{}{}
	}

	return bulkhead
}

// Execute executes a function with bulkhead protection
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	// Try to acquire a permit
	select {
	case <-b.semaphore:
		defer func() {
			b.semaphore <- struct{}{} // Return permit
		}()
		return fn()
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(b.config.Timeout):
		if b.logger != nil {
			b.logger.Warn("Bulkhead timeout", zap.Duration("timeout", b.config.Timeout))
		}
		return NewRetryableError(context.DeadlineExceeded, true)
	}
}

// GetStats returns bulkhead statistics
func (b *Bulkhead) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"max_concurrent":    b.config.MaxConcurrent,
		"available_permits": len(b.semaphore),
		"queue_size":        b.config.QueueSize,
		"queue_length":      len(b.queue),
		"timeout":           b.config.Timeout,
	}
}

// HealthChecker provides health checking capabilities
type HealthChecker struct {
	checks   map[string]HealthCheck
	mutex    sync.RWMutex
	logger   *zap.Logger
	interval time.Duration
	timeout  time.Duration
}

// HealthCheck represents a health check function
type HealthCheck struct {
	Name     string
	Check    func(ctx context.Context) error
	Interval time.Duration
	Timeout  time.Duration
	Critical bool
}

// HealthStatus represents the health status
type HealthStatus struct {
	Healthy bool                   `json:"healthy"`
	Checks  map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Healthy   bool      `json:"healthy"`
	Error     string    `json:"error,omitempty"`
	Duration  int64     `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
	Critical  bool      `json:"critical"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval, timeout time.Duration, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		checks:   make(map[string]HealthCheck),
		logger:   logger,
		interval: interval,
		timeout:  timeout,
	}
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	if check.Timeout == 0 {
		check.Timeout = hc.timeout
	}

	hc.checks[check.Name] = check

	if hc.logger != nil {
		hc.logger.Info("Health check added",
			zap.String("name", check.Name),
			zap.Bool("critical", check.Critical))
	}
}

// RemoveCheck removes a health check
func (hc *HealthChecker) RemoveCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.checks, name)

	if hc.logger != nil {
		hc.logger.Info("Health check removed", zap.String("name", name))
	}
}

// CheckHealth performs all health checks
func (hc *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	hc.mutex.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mutex.RUnlock()

	results := make(map[string]CheckResult)
	overall := true

	for name, check := range checks {
		result := hc.runCheck(ctx, check)
		results[name] = result

		if !result.Healthy && check.Critical {
			overall = false
		}
	}

	return HealthStatus{
		Healthy: overall,
		Checks:  results,
	}
}

// runCheck runs a single health check
func (hc *HealthChecker) runCheck(ctx context.Context, check HealthCheck) CheckResult {
	start := time.Now()

	checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
	defer cancel()

	err := check.Check(checkCtx)
	duration := time.Since(start)

	result := CheckResult{
		Healthy:   err == nil,
		Duration:  duration.Milliseconds(),
		Timestamp: time.Now(),
		Critical:  check.Critical,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

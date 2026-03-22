package utils

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPool manages a pool of workers for concurrent task execution
type WorkerPool struct {
	workers  int
	jobQueue chan func()
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	stopped  atomic.Bool
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if queueSize <= 0 {
		queueSize = workers * 2
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:  workers,
		jobQueue: make(chan func(), queueSize),
		ctx:      ctx,
		cancel:   cancel,
	}

	pool.start()
	return pool
}

// start initializes the worker goroutines
func (p *WorkerPool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker processes jobs from the queue
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case job := <-p.jobQueue:
			if job != nil {
				job()
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// Submit adds a job to the pool
func (p *WorkerPool) Submit(job func()) bool {
	if p.stopped.Load() {
		return false
	}

	select {
	case p.jobQueue <- job:
		return true
	case <-p.ctx.Done():
		return false
	default:
		return false // Queue full
	}

}

// SubmitAsync adds a job without blocking
func (p *WorkerPool) SubmitAsync(job func()) {
	if p.stopped.Load() {
		return
	}

	go func() {
		p.Submit(job)
	}()
}

// Stop gracefully shuts down the pool
func (p *WorkerPool) Stop() {
	if p.stopped.CompareAndSwap(false, true) {
		p.cancel()
		p.wg.Wait()
		close(p.jobQueue)
	}
}

// StopTimeout stops with a timeout
func (p *WorkerPool) StopTimeout(timeout time.Duration) bool {
	if p.stopped.CompareAndSwap(false, true) {
		p.cancel()

		done := make(chan struct{})
		go func() {
			p.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			close(p.jobQueue)
			return true
		case <-time.After(timeout):
			return false
		}
	}
	return true
}

// Stats returns pool statistics
type WorkerPoolStats struct {
	Workers   int
	QueueSize int
	QueueCap  int
	IsStopped bool
}

// GetStats returns current pool statistics
func (p *WorkerPool) GetStats() WorkerPoolStats {
	return WorkerPoolStats{
		Workers:   p.workers,
		QueueSize: len(p.jobQueue),
		QueueCap:  cap(p.jobQueue),
		IsStopped: p.stopped.Load(),
	}
}

// Throttler limits the rate of operations
type Throttler struct {
	ticker  *time.Ticker
	limitCh chan struct{}
	stopCh  chan struct{}
	stopped atomic.Bool
}

// NewThrottler creates a new throttler with the specified rate
func NewThrottler(rate time.Duration) *Throttler {
	t := &Throttler{
		ticker:  time.NewTicker(rate),
		limitCh: make(chan struct{}, 1),
		stopCh:  make(chan struct{}),
	}

	go t.run()
	return t
}

// run manages the throttle
func (t *Throttler) run() {
	for {
		select {
		case <-t.ticker.C:
			select {
			case t.limitCh <- struct{}{}:
			default:
			}
		case <-t.stopCh:
			t.ticker.Stop()
			return
		}
	}
}

// Allow waits for the next available slot
func (t *Throttler) Allow() bool {
	if t.stopped.Load() {
		return false
	}

	select {
	case <-t.limitCh:
		return true
	case <-t.stopCh:
		return false
	}
}

// AllowTimeout waits for the next available slot with timeout
func (t *Throttler) AllowTimeout(timeout time.Duration) bool {
	if t.stopped.Load() {
		return false
	}

	select {
	case <-t.limitCh:
		return true
	case <-t.stopCh:
		return false
	case <-time.After(timeout):
		return false
	}
}

// Stop stops the throttler
func (t *Throttler) Stop() {
	if t.stopped.CompareAndSwap(false, true) {
		close(t.stopCh)
	}
}

// Debouncer debounces function calls
type Debouncer struct {
	delay   time.Duration
	timer   *time.Timer
	mu      sync.Mutex
	pending func()
}

// NewDebouncer creates a new debouncer with the specified delay
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay: delay,
	}
}

// Debounce schedules the function to be called after the delay
func (d *Debouncer) Debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.pending = f

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, func() {
		d.mu.Lock()
		fn := d.pending
		d.pending = nil
		d.mu.Unlock()

		if fn != nil {
			fn()
		}
	})
}

// Flush immediately executes the pending function
func (d *Debouncer) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	if d.pending != nil {
		d.pending()
		d.pending = nil
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	state            State
	failures         int
	successes        int
	lastFailureTime  time.Time
	mu               sync.RWMutex
}

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// String returns the state name
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		state:            StateClosed,
	}
}

// Allow checks if the operation is allowed
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.failures = 0
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.successThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
		}
	case StateClosed:
		cb.failures = 0
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateOpen
	case StateClosed:
		if cb.failures >= cb.failureThreshold {
			cb.state = StateOpen
		}
	}
}

// State returns the current state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":        cb.state.String(),
		"failures":     cb.failures,
		"successes":    cb.successes,
		"last_failure": cb.lastFailureTime,
	}
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries  int
	Delay       time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	ShouldRetry func(error) bool
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		Delay:      100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
		ShouldRetry: func(err error) bool {
			return err != nil
		},
	}
}

// Retry executes the function with retry logic
func Retry(config RetryConfig, fn func() error) error {
	var err error
	delay := config.Delay

	for i := 0; i <= config.MaxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !config.ShouldRetry(err) || i == config.MaxRetries {
			return err
		}

		time.Sleep(delay)

		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return err
}

// RetryWithContext executes the function with retry logic and context
func RetryWithContext(ctx context.Context, config RetryConfig, fn func() error) error {
	var err error
	delay := config.Delay

	for i := 0; i <= config.MaxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !config.ShouldRetry(err) || i == config.MaxRetries {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return err
}

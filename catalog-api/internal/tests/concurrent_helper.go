package tests

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// ConcurrentTestRunner provides utilities for running concurrent tests
type ConcurrentTestRunner struct {
	t *testing.T
}

// NewConcurrentTestRunner creates a new concurrent test runner
func NewConcurrentTestRunner(t *testing.T) *ConcurrentTestRunner {
	return &ConcurrentTestRunner{t: t}
}

// Run executes n goroutines running fn concurrently and waits for completion
func (r *ConcurrentTestRunner) Run(n int, fn func(id int)) {
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			defer func() {
				if rec := recover(); rec != nil {
					r.t.Errorf("Goroutine %d panicked: %v", id, rec)
				}
			}()
			fn(id)
		}(i)
	}

	wg.Wait()
}

// RunWithErrors executes n goroutines and collects any errors
func (r *ConcurrentTestRunner) RunWithErrors(n int, fn func(id int) error) []error {
	var (
		mu     sync.Mutex
		errors []error
		wg     sync.WaitGroup
	)

	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			defer func() {
				if rec := recover(); rec != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("goroutine %d panicked: %v", id, rec))
					mu.Unlock()
				}
			}()

			if err := fn(id); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("goroutine %d: %w", id, err))
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return errors
}

// AssertNoErrors fails the test if there are any errors
func (r *ConcurrentTestRunner) AssertNoErrors(errors []error) {
	if len(errors) > 0 {
		var msgs []string
		for _, err := range errors {
			msgs = append(msgs, err.Error())
		}
		r.t.Fatalf("Concurrent test failed with %d error(s):\n%s",
			len(errors), strings.Join(msgs, "\n"))
	}
}

// RaceDetector helps detect race conditions in tests
type RaceDetector struct {
	t      *testing.T
	mu     sync.RWMutex
	values map[string]int
}

// NewRaceDetector creates a new race detector
func NewRaceDetector(t *testing.T) *RaceDetector {
	return &RaceDetector{
		t:      t,
		values: make(map[string]int),
	}
}

// Write simulates a write operation
func (rd *RaceDetector) Write(key string, value int) {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	rd.values[key] = value
}

// Read simulates a read operation
func (rd *RaceDetector) Read(key string) (int, bool) {
	rd.mu.RLock()
	defer rd.mu.RUnlock()
	val, ok := rd.values[key]
	return val, ok
}

// Increment atomically increments a value
func (rd *RaceDetector) Increment(key string) int {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	rd.values[key]++
	return rd.values[key]
}

// LoadTestHelper provides utilities for load testing
type LoadTestHelper struct {
	t         *testing.T
	startTime time.Time
	requests  int
	errors    int
	mu        sync.Mutex
}

// NewLoadTestHelper creates a new load test helper
func NewLoadTestHelper(t *testing.T) *LoadTestHelper {
	return &LoadTestHelper{
		t:         t,
		startTime: time.Now(),
	}
}

// RecordRequest records a successful request
func (lth *LoadTestHelper) RecordRequest() {
	lth.mu.Lock()
	defer lth.mu.Unlock()
	lth.requests++
}

// RecordError records an error
func (lth *LoadTestHelper) RecordError() {
	lth.mu.Lock()
	defer lth.mu.Unlock()
	lth.errors++
}

// GetStats returns test statistics
func (lth *LoadTestHelper) GetStats() (requests int, errors int, duration time.Duration, rps float64) {
	lth.mu.Lock()
	defer lth.mu.Unlock()

	duration = time.Since(lth.startTime)
	rps = float64(lth.requests) / duration.Seconds()

	return lth.requests, lth.errors, duration, rps
}

// PrintStats prints test statistics
func (lth *LoadTestHelper) PrintStats() {
	requests, errors, duration, rps := lth.GetStats()
	lth.t.Logf("Load Test Results:")
	lth.t.Logf("  Total Requests: %d", requests)
	lth.t.Logf("  Errors: %d", errors)
	lth.t.Logf("  Duration: %v", duration)
	lth.t.Logf("  Requests/sec: %.2f", rps)
	lth.t.Logf("  Error Rate: %.2f%%", float64(errors)/float64(requests)*100)
}

// Barrier provides a synchronization barrier for tests
type Barrier struct {
	mu    sync.Mutex
	cond  *sync.Cond
	count int
	total int
}

// NewBarrier creates a new barrier for n goroutines
func NewBarrier(n int) *Barrier {
	b := &Barrier{
		total: n,
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Wait waits for all goroutines to reach the barrier
func (b *Barrier) Wait() {
	b.mu.Lock()
	b.count++

	if b.count >= b.total {
		b.count = 0
		b.cond.Broadcast()
		b.mu.Unlock()
	} else {
		b.cond.Wait()
		b.mu.Unlock()
	}
}

// ResetBarrier resets the barrier for reuse
func (b *Barrier) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.count = 0
}

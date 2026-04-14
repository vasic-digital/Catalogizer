package utils

import (
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkWorkerPool_Submit measures the throughput of synchronous job
// submission to a worker pool. The job function is intentionally trivial
// (atomic increment) so the benchmark isolates channel-send + scheduling
// overhead rather than job execution time.
func BenchmarkWorkerPool_Submit(b *testing.B) {
	pool := NewWorkerPool(4, 1024)
	defer pool.Stop()

	var counter atomic.Int64

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pool.Submit(func() {
			counter.Add(1)
		})
	}
}

// BenchmarkWorkerPool_SubmitAsync measures the throughput of asynchronous
// (goroutine-spawning) job submission. Each call spawns a goroutine tracked
// by the pool's WaitGroup, so this stresses goroutine creation overhead on
// top of the channel send.
func BenchmarkWorkerPool_SubmitAsync(b *testing.B) {
	pool := NewWorkerPool(4, 1024)
	defer pool.Stop()

	var counter atomic.Int64

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pool.SubmitAsync(func() {
			counter.Add(1)
		})
	}
}

// BenchmarkWorkerPool_Submit_Parallel measures Submit throughput under
// concurrent goroutine contention on the job queue channel.
func BenchmarkWorkerPool_Submit_Parallel(b *testing.B) {
	pool := NewWorkerPool(4, 1024)
	defer pool.Stop()

	var counter atomic.Int64

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Submit(func() {
				counter.Add(1)
			})
		}
	})
}

// BenchmarkThrottler_Allow measures the latency of a single throttle-check
// cycle. The throttler is configured with a very short interval (1us) so
// Allow() returns immediately on each iteration, isolating the channel
// receive and atomic load overhead.
func BenchmarkThrottler_Allow(b *testing.B) {
	throttler := NewThrottler(time.Microsecond)
	defer throttler.Stop()

	// Let the ticker fill the channel once before starting.
	time.Sleep(10 * time.Microsecond)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		throttler.Allow()
	}
}

// BenchmarkDebouncer_Debounce measures the cost of scheduling a debounced
// function call. The delay is set long enough (1s) that no timer fires
// during the benchmark — this isolates mutex acquire + timer reset overhead.
func BenchmarkDebouncer_Debounce(b *testing.B) {
	debouncer := NewDebouncer(time.Second)
	noop := func() {}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		debouncer.Debounce(noop)
	}
}

// BenchmarkDebouncer_Debounce_Parallel measures debounce scheduling under
// concurrent goroutine contention on the internal mutex.
func BenchmarkDebouncer_Debounce_Parallel(b *testing.B) {
	debouncer := NewDebouncer(time.Second)
	noop := func() {}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			debouncer.Debounce(noop)
		}
	})
}

// BenchmarkCircuitBreaker_Allow measures the cost of checking whether the
// circuit breaker permits an operation. In the closed state (normal
// operation), this is the hot path and involves a mutex lock + state check.
func BenchmarkCircuitBreaker_Allow(b *testing.B) {
	cb := NewCircuitBreaker(5, 3, 30*time.Second)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.Allow()
	}
}

// BenchmarkCircuitBreaker_Allow_Parallel measures Allow() under concurrent
// contention, which stresses the RWMutex in the circuit breaker.
func BenchmarkCircuitBreaker_Allow_Parallel(b *testing.B) {
	cb := NewCircuitBreaker(5, 3, 30*time.Second)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Allow()
		}
	})
}

// BenchmarkCircuitBreaker_RecordSuccess measures the cost of recording a
// successful operation, which involves a mutex lock + counter update.
func BenchmarkCircuitBreaker_RecordSuccess(b *testing.B) {
	cb := NewCircuitBreaker(5, 3, 30*time.Second)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.RecordSuccess()
	}
}

// BenchmarkRetry_Success measures the overhead of the Retry function when
// the operation succeeds on the first attempt. This is the happy path and
// should show minimal overhead compared to calling the function directly.
func BenchmarkRetry_Success(b *testing.B) {
	config := RetryConfig{
		MaxRetries: 3,
		Delay:      100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
		ShouldRetry: func(err error) bool {
			return err != nil
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Retry(config, func() error {
			return nil
		})
	}
}

// BenchmarkGetStats measures the cost of reading pool statistics, which
// involves atomic loads and channel length queries.
func BenchmarkGetStats(b *testing.B) {
	pool := NewWorkerPool(4, 1024)
	defer pool.Stop()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = pool.GetStats()
	}
}

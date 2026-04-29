package stress

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// STRESS TEST: Goroutine Leak Detection After Operation Cycles
// =============================================================================

func TestNoGoroutineLeaks_AfterOperationCycles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping goroutine leak test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	// Force GC and let goroutines settle
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	// Simulate repeated operation cycles
	for i := 0; i < 50; i++ {
		// Allocate and release resources that might leak goroutines
		ch := make(chan struct{}, 10)
		for j := 0; j < 10; j++ {
			ch <- struct{}{}
		}
		close(ch)

		// Simulate work with a done channel pattern
		done := make(chan struct{})
		go func() {
			defer close(done)
			// Simulate brief work
			runtime.Gosched()
		}()
		<-done
	}

	// Allow goroutines to settle
	runtime.GC()
	time.Sleep(500 * time.Millisecond)
	runtime.GC()

	current := runtime.NumGoroutine()
	leaked := current - baseline

	t.Logf("Goroutine count: baseline=%d, current=%d, delta=%d", baseline, current, leaked)

	// Allow small variance for runtime goroutines
	assert.LessOrEqual(t, leaked, 5,
		"goroutine leak detected: baseline=%d, current=%d, leaked=%d",
		baseline, current, leaked)
}

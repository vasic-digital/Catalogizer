package smb

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestStopCleanupWaitsForLoopExit verifies that after StopCleanup returns,
// calling it again is a fast no-op. The wg.Wait inside StopCleanup guarantees
// the goroutine has exited before the first call returns; the second call
// hits the !isRunning guard and must return immediately.
func TestStopCleanupWaitsForLoopExit(t *testing.T) {
	cfg := DefaultConnectionPoolConfig()
	cfg.HealthCheckInterval = 10 * time.Millisecond
	pool := NewSmbConnectionPoolWithConfig(5, cfg, nil)

	// Let the loop start and run at least one tick.
	time.Sleep(30 * time.Millisecond)

	pool.StopCleanup()

	// A second StopCleanup must return fast (no-op) — if wg semantics were
	// broken we'd potentially block waiting for a goroutine that never exited.
	done := make(chan struct{})
	go func() {
		pool.StopCleanup()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second StopCleanup blocked — wg or guard semantics broken")
	}
}

// TestRestartCleanup verifies StartCleanup after StopCleanup recreates the
// stop channel and does not panic. The pre-fix code reused an already-closed
// channel and the new goroutine would exit immediately.
func TestRestartCleanup(t *testing.T) {
	cfg := DefaultConnectionPoolConfig()
	cfg.HealthCheckInterval = 10 * time.Millisecond
	pool := NewSmbConnectionPoolWithConfig(5, cfg, nil)

	time.Sleep(30 * time.Millisecond)
	pool.StopCleanup()

	// Restart — must not panic, must create a fresh goroutine.
	pool.StartCleanup()

	// Give the restarted loop a chance to run at least one tick without panic.
	time.Sleep(30 * time.Millisecond)

	// And stop it cleanly again.
	pool.StopCleanup()
}

// TestConcurrentStartStop stresses the lifecycle under concurrent calls.
// Must not panic under -race.
func TestConcurrentStartStop(t *testing.T) {
	cfg := DefaultConnectionPoolConfig()
	cfg.HealthCheckInterval = 5 * time.Millisecond
	pool := NewSmbConnectionPoolWithConfig(5, cfg, nil)

	var errs atomic.Int32
	done1 := make(chan struct{})
	done2 := make(chan struct{})

	go func() {
		defer close(done1)
		defer func() {
			if r := recover(); r != nil {
				errs.Add(1)
				t.Errorf("goroutine 1 panicked: %v", r)
			}
		}()
		for i := 0; i < 20; i++ {
			pool.StopCleanup()
			pool.StartCleanup()
		}
	}()
	go func() {
		defer close(done2)
		defer func() {
			if r := recover(); r != nil {
				errs.Add(1)
				t.Errorf("goroutine 2 panicked: %v", r)
			}
		}()
		for i := 0; i < 20; i++ {
			pool.StopCleanup()
			pool.StartCleanup()
		}
	}()

	<-done1
	<-done2
	pool.StopCleanup()

	if errs.Load() > 0 {
		t.Fatalf("got %d panics from concurrent start/stop", errs.Load())
	}
}

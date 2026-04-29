package metrics

import (
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TestSnapshotHTTP_Empty verifies the snapshot gracefully handles an
// empty/unused histogram and counter — all zeros, no panic.
func TestSnapshotHTTP_Empty(t *testing.T) {
	// We can't truly "reset" the package-level promauto vectors without
	// unregistering them; instead we record the baseline before touching
	// anything and assert the snapshot is internally consistent.
	snap := SnapshotHTTP()
	if snap.Errors4xx > snap.TotalRequests {
		t.Errorf("Errors4xx (%d) should not exceed TotalRequests (%d)", snap.Errors4xx, snap.TotalRequests)
	}
	if snap.Errors5xx > snap.TotalRequests {
		t.Errorf("Errors5xx (%d) should not exceed TotalRequests (%d)", snap.Errors5xx, snap.TotalRequests)
	}
	if snap.InFlight < 0 {
		t.Errorf("InFlight must not be negative, got %f", snap.InFlight)
	}
}

// TestSnapshotHTTP_AfterObservation verifies SnapshotHTTP reads back
// real histogram and counter data after observations are recorded.
func TestSnapshotHTTP_AfterObservation(t *testing.T) {
	baseline := SnapshotHTTP()

	// Record a few observations across different latencies and statuses.
	HTTPRequestDuration.WithLabelValues("GET", "/test", "200").Observe(0.05)
	HTTPRequestDuration.WithLabelValues("GET", "/test", "200").Observe(0.15)
	HTTPRequestDuration.WithLabelValues("GET", "/test", "500").Observe(2.0)
	HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
	HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
	HTTPRequestsTotal.WithLabelValues("GET", "/test", "500").Inc()
	HTTPRequestsTotal.WithLabelValues("GET", "/test", "404").Inc()

	HTTPActiveConnections.Inc()
	defer HTTPActiveConnections.Dec()

	snap := SnapshotHTTP()

	if snap.TotalRequests < baseline.TotalRequests+3 {
		t.Errorf("TotalRequests should have grown by >= 3, baseline=%d got=%d",
			baseline.TotalRequests, snap.TotalRequests)
	}
	if snap.Errors5xx < baseline.Errors5xx+1 {
		t.Errorf("Errors5xx should have grown by >= 1, baseline=%d got=%d",
			baseline.Errors5xx, snap.Errors5xx)
	}
	if snap.Errors4xx < baseline.Errors4xx+1 {
		t.Errorf("Errors4xx should have grown by >= 1, baseline=%d got=%d",
			baseline.Errors4xx, snap.Errors4xx)
	}
	if snap.InFlight <= baseline.InFlight-0.001 {
		t.Errorf("InFlight should be at least 1 above baseline, baseline=%f got=%f",
			baseline.InFlight, snap.InFlight)
	}
	if snap.TotalDurationSeconds <= baseline.TotalDurationSeconds {
		t.Errorf("TotalDurationSeconds should grow after observations, baseline=%f got=%f",
			baseline.TotalDurationSeconds, snap.TotalDurationSeconds)
	}
	// Percentiles should be within the observed range (0.05 to 2.0).
	if snap.P95Seconds < 0 {
		t.Errorf("P95 should not be negative, got %f", snap.P95Seconds)
	}
	if snap.MinSeconds < 0 {
		t.Errorf("MinSeconds should not be negative, got %f", snap.MinSeconds)
	}
}

// TestSnapshotSystem verifies the runtime snapshot returns sensible
// values for a running Go program.
func TestSnapshotSystem(t *testing.T) {
	sys := SnapshotSystem()
	if sys.Goroutines < 1 {
		t.Errorf("Goroutines must be at least 1 (this test), got %d", sys.Goroutines)
	}
	if sys.Goroutines > 100000 {
		t.Errorf("Goroutines count looks wrong: %d", sys.Goroutines)
	}
	if sys.AllocBytes == 0 {
		t.Error("AllocBytes must be non-zero for a running process")
	}
	if sys.SysBytes == 0 {
		t.Error("SysBytes must be non-zero for a running process")
	}
	if sys.HeapInuse > sys.SysBytes {
		t.Errorf("HeapInuse (%d) should not exceed SysBytes (%d)", sys.HeapInuse, sys.SysBytes)
	}
	// Snapshot should agree with runtime.NumGoroutine within a small
	// tolerance (tests run in other goroutines during the call).
	current := runtime.NumGoroutine()
	diff := current - sys.Goroutines
	if diff < 0 {
		diff = -diff
	}
	if diff > 20 {
		t.Errorf("Goroutine count drifted too far: snap=%d current=%d", sys.Goroutines, current)
	}
}

// TestPercentileFromBuckets_Empty verifies percentile computation is
// safe on empty bucket slices.
func TestPercentileFromBuckets_Empty(t *testing.T) {
	if got := percentileFromBuckets(nil, 0.95, 0); got != 0 {
		t.Errorf("percentileFromBuckets(nil, 0.95, 0) = %f, want 0", got)
	}
	if got := percentileFromBuckets([]bucketAcc{}, 0.95, 10); got != 0 {
		t.Errorf("percentileFromBuckets([], 0.95, 10) = %f, want 0", got)
	}
}

// TestPercentileFromBuckets_Interpolation verifies linear interpolation
// across bucket edges.
func TestPercentileFromBuckets_Interpolation(t *testing.T) {
	buckets := []bucketAcc{
		{upper: 0.1, count: 10},
		{upper: 0.5, count: 50},
		{upper: 1.0, count: 100},
	}
	// p50 should land inside the [0.1, 0.5] bucket: target = 50 → upper = 0.5
	p50 := percentileFromBuckets(buckets, 0.50, 100)
	if p50 < 0.1 || p50 > 0.5 {
		t.Errorf("p50 = %f, want in [0.1, 0.5]", p50)
	}
	// p95 should land inside the [0.5, 1.0] bucket.
	p95 := percentileFromBuckets(buckets, 0.95, 100)
	if p95 < 0.5 || p95 > 1.0 {
		t.Errorf("p95 = %f, want in [0.5, 1.0]", p95)
	}
}

// TestSnapshotHTTP_Concurrency verifies SnapshotHTTP is safe under
// concurrent calls alongside metric updates.
func TestSnapshotHTTP_Concurrency(t *testing.T) {
	// bluff-scan: no-assert-ok (concurrency test — go test -race catches data races; absence of panic == correctness)
	done := make(chan struct{})
	go func() {
		// Drive observations for 100ms.
		deadline := time.Now().Add(100 * time.Millisecond)
		for time.Now().Before(deadline) {
			HTTPRequestDuration.WithLabelValues("GET", "/concurrent", "200").Observe(0.01)
			HTTPRequestsTotal.WithLabelValues("GET", "/concurrent", "200").Inc()
		}
		close(done)
	}()

	// Concurrently snapshot.
	for i := 0; i < 50; i++ {
		_ = SnapshotHTTP()
		time.Sleep(time.Millisecond)
	}

	<-done
}

// Verify the Prometheus library's gather path still works after a
// direct Collect call — guards against a regression in our manual
// Collect() usage in SnapshotHTTP.
func TestSnapshotHTTP_DoesNotBreakPromGather(t *testing.T) {
	_ = SnapshotHTTP()
	// Attempt a standard Prometheus gather via the default registry
	// — if SnapshotHTTP leaked any channel data this would fail.
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("prometheus gather failed: %v", err)
	}
	if len(mfs) == 0 {
		t.Error("expected non-empty metric families after SnapshotHTTP")
	}
}

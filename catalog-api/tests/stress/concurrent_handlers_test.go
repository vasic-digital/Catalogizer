package stress

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// STRESS TEST: Concurrent HTTP Handlers (50 goroutines, p95 latency)
// =============================================================================

func TestConcurrentHandlers_50Goroutines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	ts := setupStressTestServer(t)
	ltc := newLoadTestContext(ts.URL)
	ltc.authenticate(t)

	t.Run("50ConcurrentHandlerRequests", func(t *testing.T) {
		concurrency := 50
		requestsPerGoroutine := 20

		var wg sync.WaitGroup
		latencies := make([]time.Duration, 0,
			concurrency*requestsPerGoroutine)
		var mu sync.Mutex

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				endpoints := []string{
					"/media?page=1&limit=10",
					"/storage/roots",
					"/analytics/dashboard",
					"/collections",
				}
				for j := 0; j < requestsPerGoroutine; j++ {
					ep := endpoints[j%len(endpoints)]
					resp, latency, err := ltc.makeRequest(
						"GET", ep, nil)
					if err == nil && resp != nil {
						resp.Body.Close()
					}
					mu.Lock()
					latencies = append(latencies, latency)
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()
		ltc.PrintStats(t)

		// Calculate p95 from collected latencies
		mu.Lock()
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})
		p95Idx := int(float64(len(latencies)) * 0.95)
		if p95Idx >= len(latencies) {
			p95Idx = len(latencies) - 1
		}
		p95 := latencies[p95Idx]
		mu.Unlock()

		stats := ltc.GetStats()
		successRate := stats["success_rate"].(float64)

		t.Logf("P95 latency: %v", p95)
		t.Logf("Total requests: %d", stats["requests"])

		assert.Greater(t, successRate, 95.0,
			"Success rate should be >95%% with 50 goroutines")
		assert.Less(t, p95, 500*time.Millisecond,
			"P95 latency should be under 500ms")
		assert.Greater(t, stats["requests"].(int64),
			int64(concurrency*requestsPerGoroutine/2),
			"Should process the majority of requests")
	})
}

func TestConcurrentHandlers_MixedMethodsUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	ts := setupStressTestServer(t)
	ltc := newLoadTestContext(ts.URL)
	ltc.authenticate(t)

	t.Run("GETAndPOSTMixed", func(t *testing.T) {
		concurrency := 50
		requestsPerGoroutine := 15
		var readOps, writeOps int64

		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					if id%3 == 0 {
						// POST (write)
						data := map[string]interface{}{
							"event_type":  "handler_stress",
							"entity_type": "test",
							"entity_id":   id*100 + j,
						}
						resp, _, err := ltc.makeRequest(
							"POST", "/analytics/track", data)
						if err == nil && resp != nil {
							resp.Body.Close()
						}
						atomic.AddInt64(&writeOps, 1)
					} else {
						// GET (read)
						resp, _, err := ltc.makeRequest(
							"GET",
							"/media?page=1&limit=10", nil)
						if err == nil && resp != nil {
							resp.Body.Close()
						}
						atomic.AddInt64(&readOps, 1)
					}
				}
			}(i)
		}

		wg.Wait()
		ltc.PrintStats(t)

		reads := atomic.LoadInt64(&readOps)
		writes := atomic.LoadInt64(&writeOps)
		t.Logf("Read ops: %d, Write ops: %d", reads, writes)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0,
			"Mixed method load should have >90%% success")
		assert.Greater(t, reads, writes,
			"Read operations should outnumber writes")
	})
}

func TestConcurrentHandlers_HealthCheckUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	ts := setupStressTestServer(t)

	t.Run("50ConcurrentHealthChecks", func(t *testing.T) {
		concurrency := 50
		requestsPerGoroutine := 50
		var successCount int64
		latencies := make([]time.Duration, 0,
			concurrency*requestsPerGoroutine)
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := &http.Client{Timeout: 5 * time.Second}
				for j := 0; j < requestsPerGoroutine; j++ {
					start := time.Now()
					resp, err := client.Get(
						ts.URL + "/health")
					elapsed := time.Since(start)

					if err == nil && resp != nil {
						if resp.StatusCode == http.StatusOK {
							atomic.AddInt64(&successCount, 1)
						}
						resp.Body.Close()
					}
					mu.Lock()
					latencies = append(latencies, elapsed)
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		total := int64(concurrency * requestsPerGoroutine)
		success := atomic.LoadInt64(&successCount)
		rate := float64(success) / float64(total) * 100

		// Calculate p95
		mu.Lock()
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})
		p95 := latencies[int(float64(len(latencies))*0.95)]
		mu.Unlock()

		t.Logf("Health check: %d/%d succeeded (%.2f%%), P95: %v",
			success, total, rate, p95)

		assert.Greater(t, rate, 99.0,
			"Health checks should have >99%% success rate")
		assert.Less(t, p95, 200*time.Millisecond,
			"P95 health check latency should be under 200ms")
	})
}

func TestConcurrentHandlers_ErrorResponseConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	ts := setupStressTestServer(t)

	t.Run("Concurrent404Requests", func(t *testing.T) {
		concurrency := 50
		var notFoundCount int64
		var otherCount int64

		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{Timeout: 5 * time.Second}
				path := fmt.Sprintf(
					"/api/v1/nonexistent/%d", id)
				resp, err := client.Get(ts.URL + path)
				if err != nil {
					return
				}
				resp.Body.Close()
				if resp.StatusCode == http.StatusNotFound {
					atomic.AddInt64(&notFoundCount, 1)
				} else {
					atomic.AddInt64(&otherCount, 1)
				}
			}(i)
		}

		wg.Wait()

		notFound := atomic.LoadInt64(&notFoundCount)
		other := atomic.LoadInt64(&otherCount)
		t.Logf("404 responses: %d, Other: %d", notFound, other)

		assert.Equal(t, int64(concurrency), notFound,
			"All requests to non-existent paths should return 404")
		assert.Equal(t, int64(0), other,
			"No other status codes should be returned")
	})
}

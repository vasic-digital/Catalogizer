package stress

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// STRESS TEST: Concurrent API Requests (100+ goroutines)
// =============================================================================

func TestConcurrentAPIStress_100Goroutines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupStressTestServer(t)
	ltc := newLoadTestContext(ts.URL)
	ltc.authenticate(t)

	t.Run("100GoroutinesReadEndpoints", func(t *testing.T) {
		concurrentGoroutines := 100
		requestsPerGoroutine := 20

		var wg sync.WaitGroup
		for i := 0; i < concurrentGoroutines; i++ {
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
					endpoint := endpoints[j%len(endpoints)]
					resp, _, err := ltc.makeRequest("GET", endpoint, nil)
					if err == nil && resp != nil {
						resp.Body.Close()
					}
				}
			}(i)
		}

		wg.Wait()
		ltc.PrintStats(t)

		stats := ltc.GetStats()
		successRate := stats["success_rate"].(float64)
		avgLatency := stats["avg_latency"].(time.Duration)

		assert.Greater(t, successRate, 95.0, "Success rate should be >95% with 100 goroutines")
		assert.Less(t, avgLatency, 500*time.Millisecond, "Average latency should be under 500ms")
		assert.Greater(t, stats["requests"].(int64), int64(1000), "Should have processed at least 1000 requests")
	})
}

func TestConcurrentAPIStress_200Goroutines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupStressTestServer(t)
	ltc := newLoadTestContext(ts.URL)
	ltc.authenticate(t)

	t.Run("200GoroutinesBurst", func(t *testing.T) {
		concurrentGoroutines := 200
		requestsPerGoroutine := 10

		var wg sync.WaitGroup
		for i := 0; i < concurrentGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					resp, _, err := ltc.makeRequest("GET", "/media?page=1&limit=10", nil)
					if err == nil && resp != nil {
						resp.Body.Close()
					}
					time.Sleep(5 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Should handle 200 concurrent goroutines with >90% success")
	})
}

// =============================================================================
// STRESS TEST: Mixed Read/Write Concurrent API Requests
// =============================================================================

func TestConcurrentAPIStress_MixedReadWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupStressTestServer(t)
	ltc := newLoadTestContext(ts.URL)
	ltc.authenticate(t)

	t.Run("MixedReadWrite100Goroutines", func(t *testing.T) {
		concurrentGoroutines := 100
		requestsPerGoroutine := 15

		var readCount int64
		var writeCount int64

		var wg sync.WaitGroup
		for i := 0; i < concurrentGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					if id%4 == 0 {
						// 25% write operations
						eventData := map[string]interface{}{
							"event_type":  "stress_test",
							"entity_type": "test",
							"entity_id":   id*requestsPerGoroutine + j,
						}
						resp, _, err := ltc.makeRequest("POST", "/analytics/track", eventData)
						if err == nil && resp != nil {
							resp.Body.Close()
						}
						atomic.AddInt64(&writeCount, 1)
					} else {
						// 75% read operations
						resp, _, err := ltc.makeRequest("GET", "/media?page=1&limit=10", nil)
						if err == nil && resp != nil {
							resp.Body.Close()
						}
						atomic.AddInt64(&readCount, 1)
					}
					time.Sleep(5 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
		ltc.PrintStats(t)

		t.Logf("Read operations: %d, Write operations: %d", readCount, writeCount)
		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Mixed read/write should maintain >90% success")
		assert.Greater(t, readCount, writeCount, "Read count should exceed write count")
	})
}

// =============================================================================
// STRESS TEST: Concurrent Login Storm
// =============================================================================

func TestConcurrentAPIStress_LoginStorm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupStressTestServer(t)
	ltc := newLoadTestContext(ts.URL)

	t.Run("100ConcurrentLogins", func(t *testing.T) {
		concurrentLogins := 100
		var successCount int64

		var wg sync.WaitGroup
		for i := 0; i < concurrentLogins; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				loginData := map[string]interface{}{
					"username": "admin",
					"password": "admin123",
				}
				resp, _, err := ltc.makeRequest("POST", "/auth/login", loginData)
				if err == nil && resp != nil {
					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&successCount, 1)
					}
					resp.Body.Close()
				}
			}()
		}

		wg.Wait()
		ltc.PrintStats(t)

		assert.Equal(t, int64(concurrentLogins), successCount, "All concurrent logins should succeed")
	})

	t.Run("LoginWithInvalidCredentialsBurst", func(t *testing.T) {
		concurrentLogins := 50
		var failCount int64

		var wg sync.WaitGroup
		for i := 0; i < concurrentLogins; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				loginData := map[string]interface{}{
					"username": fmt.Sprintf("invalid_user_%d", id),
					"password": "wrongpassword",
				}
				resp, _, err := ltc.makeRequest("POST", "/auth/login", loginData)
				if err == nil && resp != nil {
					if resp.StatusCode == http.StatusUnauthorized {
						atomic.AddInt64(&failCount, 1)
					}
					resp.Body.Close()
				}
			}(i)
		}

		wg.Wait()

		assert.Equal(t, int64(concurrentLogins), failCount, "All invalid logins should return 401")
	})
}

// =============================================================================
// STRESS TEST: Endpoint Rotation Under Load
// =============================================================================

func TestConcurrentAPIStress_EndpointRotation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupStressTestServer(t)
	ltc := newLoadTestContext(ts.URL)
	ltc.authenticate(t)

	t.Run("RotateEndpoints150Goroutines", func(t *testing.T) {
		concurrentGoroutines := 150
		duration := 15 * time.Second

		endpoints := []string{
			"/media?page=1&limit=10",
			"/storage/roots",
			"/analytics/dashboard",
			"/collections",
		}

		done := make(chan bool)
		var wg sync.WaitGroup

		for i := 0; i < concurrentGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				requestIdx := 0
				for {
					select {
					case <-done:
						return
					default:
						endpoint := endpoints[requestIdx%len(endpoints)]
						resp, _, err := ltc.makeRequest("GET", endpoint, nil)
						if err == nil && resp != nil {
							resp.Body.Close()
						}
						requestIdx++
						time.Sleep(10 * time.Millisecond)
					}
				}
			}(i)
		}

		time.Sleep(duration)
		close(done)
		wg.Wait()

		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Endpoint rotation should maintain >90% success")
		assert.Greater(t, stats["rps"].(float64), 500.0, "Should sustain >500 RPS across endpoints")
	})
}

// =============================================================================
// STRESS TEST: Rapid Connection Open/Close
// =============================================================================

func TestConcurrentAPIStress_RapidConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupStressTestServer(t)

	t.Run("RapidConnectionCycles", func(t *testing.T) {
		iterations := 200
		var successCount int64
		var errorCount int64

		var wg sync.WaitGroup
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Each goroutine creates its own client (new connection)
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(ts.URL + "/health")
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					return
				}
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}()
		}

		wg.Wait()

		successRate := float64(successCount) / float64(iterations) * 100
		t.Logf("Rapid connections: %d/%d succeeded (%.2f%%)", successCount, iterations, successRate)
		assert.Greater(t, successRate, 95.0, "Rapid connection cycles should have >95% success")
	})
}

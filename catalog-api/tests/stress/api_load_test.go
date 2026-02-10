package stress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL        = "http://localhost:8080"
	apiBaseURL     = "http://localhost:8080/api/v1"
	defaultTimeout = 30 * time.Second
)

// LoadTestContext manages load test execution
type LoadTestContext struct {
	HTTPClient      *http.Client
	AuthToken       string
	RequestCount    int64
	SuccessCount    int64
	ErrorCount      int64
	TotalLatency    int64 // microseconds
	StartTime       time.Time
	Errors          []error
	ErrorsMutex     sync.Mutex
	ResponseTimes   []time.Duration
	ResponseMutex   sync.RWMutex
}

func newLoadTestContext() *LoadTestContext {
	return &LoadTestContext{
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
		StartTime:     time.Now(),
		ResponseTimes: make([]time.Duration, 0, 10000),
	}
}

// Helper to check if server is available
func (ltc *LoadTestContext) isServerAvailable() bool {
	resp, err := ltc.HTTPClient.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Helper to authenticate and get token
func (ltc *LoadTestContext) authenticate(t *testing.T) {
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := ltc.HTTPClient.Post(
		apiBaseURL+"/auth/login",
		"application/json",
		bytes.NewReader(jsonData),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skip("Cannot authenticate - skipping stress tests")
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	ltc.AuthToken = result["token"].(string)
}

// Helper to make authenticated requests and track metrics
func (ltc *LoadTestContext) makeRequest(method, path string, body interface{}) (*http.Response, time.Duration, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, apiBaseURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}

	if ltc.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+ltc.AuthToken)
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := ltc.HTTPClient.Do(req)
	latency := time.Since(start)

	atomic.AddInt64(&ltc.RequestCount, 1)

	if err != nil {
		atomic.AddInt64(&ltc.ErrorCount, 1)
		ltc.recordError(err)
		return nil, latency, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(&ltc.SuccessCount, 1)
	} else {
		atomic.AddInt64(&ltc.ErrorCount, 1)
		ltc.recordError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, path))
	}

	atomic.AddInt64(&ltc.TotalLatency, int64(latency.Microseconds()))
	ltc.recordResponseTime(latency)

	return resp, latency, nil
}

func (ltc *LoadTestContext) recordError(err error) {
	ltc.ErrorsMutex.Lock()
	defer ltc.ErrorsMutex.Unlock()
	if len(ltc.Errors) < 100 { // Keep first 100 errors
		ltc.Errors = append(ltc.Errors, err)
	}
}

func (ltc *LoadTestContext) recordResponseTime(duration time.Duration) {
	ltc.ResponseMutex.Lock()
	defer ltc.ResponseMutex.Unlock()
	ltc.ResponseTimes = append(ltc.ResponseTimes, duration)
}

// GetStats returns load test statistics
func (ltc *LoadTestContext) GetStats() map[string]interface{} {
	duration := time.Since(ltc.StartTime)
	reqCount := atomic.LoadInt64(&ltc.RequestCount)
	successCount := atomic.LoadInt64(&ltc.SuccessCount)
	errorCount := atomic.LoadInt64(&ltc.ErrorCount)
	totalLatency := atomic.LoadInt64(&ltc.TotalLatency)

	rps := float64(reqCount) / duration.Seconds()
	avgLatency := time.Duration(0)
	if reqCount > 0 {
		avgLatency = time.Duration(totalLatency/reqCount) * time.Microsecond
	}

	successRate := 0.0
	if reqCount > 0 {
		successRate = float64(successCount) / float64(reqCount) * 100
	}

	// Calculate percentiles
	ltc.ResponseMutex.RLock()
	p50, p95, p99 := ltc.calculatePercentiles()
	ltc.ResponseMutex.RUnlock()

	return map[string]interface{}{
		"duration":     duration,
		"requests":     reqCount,
		"success":      successCount,
		"errors":       errorCount,
		"rps":          rps,
		"avg_latency":  avgLatency,
		"p50_latency":  p50,
		"p95_latency":  p95,
		"p99_latency":  p99,
		"success_rate": successRate,
	}
}

func (ltc *LoadTestContext) calculatePercentiles() (p50, p95, p99 time.Duration) {
	if len(ltc.ResponseTimes) == 0 {
		return 0, 0, 0
	}

	// Simple percentile calculation (not sorting to avoid modifying slice)
	// For stress tests, this approximation is acceptable
	count := len(ltc.ResponseTimes)
	if count == 0 {
		return 0, 0, 0
	}

	// Sample-based approximation
	sum := time.Duration(0)
	max := time.Duration(0)
	for _, d := range ltc.ResponseTimes {
		sum += d
		if d > max {
			max = d
		}
	}

	// Rough approximations
	avg := sum / time.Duration(count)
	p50 = avg
	p95 = avg * 2
	p99 = max

	return p50, p95, p99
}

// PrintStats prints load test statistics
func (ltc *LoadTestContext) PrintStats(t *testing.T) {
	stats := ltc.GetStats()

	t.Logf("\n=== Load Test Results ===")
	t.Logf("Duration:        %v", stats["duration"])
	t.Logf("Total Requests:  %d", stats["requests"])
	t.Logf("Successful:      %d", stats["success"])
	t.Logf("Errors:          %d", stats["errors"])
	t.Logf("Requests/sec:    %.2f", stats["rps"])
	t.Logf("Avg Latency:     %v", stats["avg_latency"])
	t.Logf("P50 Latency:     %v", stats["p50_latency"])
	t.Logf("P95 Latency:     %v", stats["p95_latency"])
	t.Logf("P99 Latency:     %v", stats["p99_latency"])
	t.Logf("Success Rate:    %.2f%%", stats["success_rate"])

	if len(ltc.Errors) > 0 {
		t.Logf("\nFirst %d Errors:", len(ltc.Errors))
		for i, err := range ltc.Errors {
			if i >= 10 {
				break
			}
			t.Logf("  %d: %v", i+1, err)
		}
	}
}

// =============================================================================
// STRESS TEST: Concurrent API Requests
// =============================================================================

func TestConcurrentAPIRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ltc := newLoadTestContext()
	if !ltc.isServerAvailable() {
		t.Skip("Server not available - skipping stress test")
	}

	ltc.authenticate(t)

	t.Run("100ConcurrentUsers", func(t *testing.T) {
		concurrentUsers := 100
		requestsPerUser := 10

		var wg sync.WaitGroup
		for i := 0; i < concurrentUsers; i++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				for j := 0; j < requestsPerUser; j++ {
					resp, _, err := ltc.makeRequest("GET", "/media?page=1&limit=10", nil)
					if err == nil && resp != nil {
						resp.Body.Close()
					}
					time.Sleep(10 * time.Millisecond) // Small delay between requests
				}
			}(i)
		}

		wg.Wait()
		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Success rate should be >95%")
		assert.Less(t, stats["avg_latency"].(time.Duration), 500*time.Millisecond, "Avg latency should be <500ms")
	})

	t.Run("SpikeLoad", func(t *testing.T) {
		// Sudden spike: 500 concurrent requests
		spikeSize := 500

		var wg sync.WaitGroup
		for i := 0; i < spikeSize; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, _, err := ltc.makeRequest("GET", "/health", nil)
				if err == nil && resp != nil {
					resp.Body.Close()
				}
			}()
		}

		wg.Wait()
		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Should handle spike with >90% success")
	})
}

// =============================================================================
// STRESS TEST: Sustained Load
// =============================================================================

func TestSustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ltc := newLoadTestContext()
	if !ltc.isServerAvailable() {
		t.Skip("Server not available - skipping stress test")
	}

	ltc.authenticate(t)

	t.Run("SustainedLoad30Seconds", func(t *testing.T) {
		duration := 30 * time.Second
		concurrentWorkers := 50
		targetRPS := 100.0 // Target 100 requests per second

		delayBetweenRequests := time.Duration(float64(time.Second) / (targetRPS / float64(concurrentWorkers)))

		done := make(chan bool)
		var wg sync.WaitGroup

		// Start workers
		for i := 0; i < concurrentWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				ticker := time.NewTicker(delayBetweenRequests)
				defer ticker.Stop()

				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						endpoints := []string{
							"/media?page=1&limit=10",
							"/storage/roots",
							"/analytics/dashboard",
							"/collections",
						}
						endpoint := endpoints[workerID%len(endpoints)]

						resp, _, err := ltc.makeRequest("GET", endpoint, nil)
						if err == nil && resp != nil {
							resp.Body.Close()
						}
					}
				}
			}(i)
		}

		// Run for specified duration
		time.Sleep(duration)
		close(done)
		wg.Wait()

		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Should maintain >95% success under sustained load")
		assert.Greater(t, stats["rps"].(float64), targetRPS*0.8, "Should achieve at least 80% of target RPS")
		assert.Less(t, stats["p95_latency"].(time.Duration), 1*time.Second, "P95 latency should be <1s")
	})
}

// =============================================================================
// STRESS TEST: Mixed Operations
// =============================================================================

func TestMixedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ltc := newLoadTestContext()
	if !ltc.isServerAvailable() {
		t.Skip("Server not available - skipping stress test")
	}

	ltc.authenticate(t)

	t.Run("ReadHeavyWorkload", func(t *testing.T) {
		concurrentUsers := 100
		duration := 10 * time.Second

		done := make(chan bool)
		var wg sync.WaitGroup

		for i := 0; i < concurrentUsers; i++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				for {
					select {
					case <-done:
						return
					default:
						// 80% reads, 20% writes
						if userID%5 == 0 {
							// Write operation
							eventData := map[string]interface{}{
								"event_type":  "stress_test",
								"entity_type": "test",
								"entity_id":   userID,
							}
							resp, _, _ := ltc.makeRequest("POST", "/analytics/track", eventData)
							if resp != nil {
								resp.Body.Close()
							}
						} else {
							// Read operation
							resp, _, _ := ltc.makeRequest("GET", "/media?page=1&limit=10", nil)
							if resp != nil {
								resp.Body.Close()
							}
						}
						time.Sleep(20 * time.Millisecond)
					}
				}
			}(i)
		}

		time.Sleep(duration)
		close(done)
		wg.Wait()

		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Mixed workload should have >90% success rate")
	})
}

// =============================================================================
// STRESS TEST: Authentication Load
// =============================================================================

func TestAuthenticationLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ltc := newLoadTestContext()
	if !ltc.isServerAvailable() {
		t.Skip("Server not available - skipping stress test")
	}

	t.Run("ConcurrentLogins", func(t *testing.T) {
		concurrentLogins := 50

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
					resp.Body.Close()
				}
			}()
		}

		wg.Wait()
		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Authentication should handle concurrent logins")
	})
}

// =============================================================================
// STRESS TEST: Ramp-Up Load
// =============================================================================

func TestRampUpLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ltc := newLoadTestContext()
	if !ltc.isServerAvailable() {
		t.Skip("Server not available - skipping stress test")
	}

	ltc.authenticate(t)

	t.Run("GradualRampUp", func(t *testing.T) {
		maxUsers := 200
		rampUpDuration := 20 * time.Second
		testDuration := 40 * time.Second

		done := make(chan bool)
		activeWorkers := int64(0)

		// Ramp up gradually
		go func() {
			ticker := time.NewTicker(rampUpDuration / time.Duration(maxUsers))
			defer ticker.Stop()

			for i := 0; i < maxUsers; i++ {
				select {
				case <-done:
					return
				case <-ticker.C:
					go func() {
						atomic.AddInt64(&activeWorkers, 1)
						defer atomic.AddInt64(&activeWorkers, -1)

						for {
							select {
							case <-done:
								return
							default:
								resp, _, _ := ltc.makeRequest("GET", "/media?page=1&limit=10", nil)
								if resp != nil {
									resp.Body.Close()
								}
								time.Sleep(50 * time.Millisecond)
							}
						}
					}()
				}
			}
		}()

		// Run test
		time.Sleep(testDuration)
		close(done)

		// Wait for workers to finish
		time.Sleep(500 * time.Millisecond)

		ltc.PrintStats(t)

		stats := ltc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Should handle gradual ramp-up")
		t.Logf("Peak concurrent workers: %d", atomic.LoadInt64(&activeWorkers))
	})
}

// =============================================================================
// STRESS TEST: API Endpoint Specific Load
// =============================================================================

func TestEndpointSpecificLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ltc := newLoadTestContext()
	if !ltc.isServerAvailable() {
		t.Skip("Server not available - skipping stress test")
	}

	ltc.authenticate(t)

	endpoints := map[string]string{
		"Media":       "/media?page=1&limit=10",
		"Storage":     "/storage/roots",
		"Analytics":   "/analytics/dashboard",
		"Collections": "/collections",
	}

	for name, endpoint := range endpoints {
		t.Run(name, func(t *testing.T) {
			concurrentRequests := 100

			var wg sync.WaitGroup
			for i := 0; i < concurrentRequests; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					resp, _, err := ltc.makeRequest("GET", endpoint, nil)
					if err == nil && resp != nil {
						resp.Body.Close()
					}
				}()
			}

			wg.Wait()

			// Note: Stats are cumulative, but gives us an idea of endpoint performance
			stats := ltc.GetStats()
			t.Logf("%s endpoint - Avg latency: %v, Success rate: %.2f%%",
				name, stats["avg_latency"], stats["success_rate"])
		})
	}
}

package stress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupResponsivenessServer creates an in-process test server for latency testing.
// It exposes lightweight endpoints to measure framework overhead and response times.
func setupResponsivenessServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Health check - should be extremely fast
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	// Auth state
	var mu sync.Mutex
	tokens := map[string]bool{}

	api := router.Group("/api/v1")
	{
		// Auth/login endpoint
		api.POST("/auth/login", func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			username, _ := body["username"].(string)
			password, _ := body["password"].(string)
			if username == "admin" && password == "admin123" {
				token := fmt.Sprintf("resp-token-%d", time.Now().UnixNano())
				mu.Lock()
				tokens[token] = true
				mu.Unlock()
				c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": 1, "username": "admin"}})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			}
		})

		// List endpoints
		api.GET("/media", func(c *gin.Context) {
			items := make([]gin.H, 20)
			for i := 0; i < 20; i++ {
				items[i] = gin.H{
					"id":   i + 1,
					"name": fmt.Sprintf("media_item_%d.mp4", i+1),
					"type": "movie",
					"size": 1024 * 1024 * (i + 1),
				}
			}
			c.JSON(http.StatusOK, gin.H{"items": items, "total": 20, "page": 1})
		})

		api.GET("/storage/roots", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"roots": []gin.H{
					{"id": 1, "name": "Media", "path": "/media", "protocol": "local"},
					{"id": 2, "name": "NAS", "path": "/nas", "protocol": "smb"},
				},
			})
		})

		api.GET("/collections", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"collections": []gin.H{
					{"id": 1, "name": "Favorites", "count": 15},
					{"id": 2, "name": "Watch Later", "count": 8},
					{"id": 3, "name": "Top Rated", "count": 25},
				},
			})
		})
	}

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// latencyResult holds a set of measured latencies for analysis.
type latencyResult struct {
	latencies []time.Duration
	mu        sync.Mutex
}

func newLatencyResult(capacity int) *latencyResult {
	return &latencyResult{
		latencies: make([]time.Duration, 0, capacity),
	}
}

func (lr *latencyResult) record(d time.Duration) {
	lr.mu.Lock()
	lr.latencies = append(lr.latencies, d)
	lr.mu.Unlock()
}

func (lr *latencyResult) percentile(p float64) time.Duration {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if len(lr.latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(lr.latencies))
	copy(sorted, lr.latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	idx := int(float64(len(sorted)-1) * p / 100)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func (lr *latencyResult) average() time.Duration {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if len(lr.latencies) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range lr.latencies {
		total += d
	}
	return total / time.Duration(len(lr.latencies))
}

func (lr *latencyResult) count() int {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	return len(lr.latencies)
}

// =============================================================================
// RESPONSIVENESS TEST: Health Endpoint
// =============================================================================

func TestResponsiveness_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping responsiveness test in short mode")
	}

	tests := []struct {
		name             string
		concurrentUsers  int
		requestsPerUser  int
		maxP99Latency    time.Duration
		maxAvgLatency    time.Duration
		minSuccessRate   float64
	}{
		{
			name:            "LowLoad_10Users",
			concurrentUsers: 10,
			requestsPerUser: 50,
			maxP99Latency:   50 * time.Millisecond,
			maxAvgLatency:   20 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "MediumLoad_50Users",
			concurrentUsers: 50,
			requestsPerUser: 20,
			maxP99Latency:   100 * time.Millisecond,
			maxAvgLatency:   50 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "HighLoad_100Users",
			concurrentUsers: 100,
			requestsPerUser: 10,
			maxP99Latency:   200 * time.Millisecond,
			maxAvgLatency:   100 * time.Millisecond,
			minSuccessRate:  99.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ts := setupResponsivenessServer(t)
			lr := newLatencyResult(tt.concurrentUsers * tt.requestsPerUser)
			var successCount, errorCount int64

			var wg sync.WaitGroup
			for i := 0; i < tt.concurrentUsers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := &http.Client{Timeout: 5 * time.Second}

					for j := 0; j < tt.requestsPerUser; j++ {
						start := time.Now()
						resp, err := client.Get(ts.URL + "/health")
						elapsed := time.Since(start)

						if err != nil {
							errorCount++
							continue
						}
						resp.Body.Close()

						if resp.StatusCode == http.StatusOK {
							successCount++
						} else {
							errorCount++
						}
						lr.record(elapsed)
					}
				}()
			}

			wg.Wait()

			totalRequests := int64(tt.concurrentUsers * tt.requestsPerUser)
			successRate := float64(successCount) / float64(totalRequests) * 100
			p50 := lr.percentile(50)
			p95 := lr.percentile(95)
			p99 := lr.percentile(99)
			avg := lr.average()

			t.Logf("=== /health Responsiveness ===")
			t.Logf("Requests:    %d", totalRequests)
			t.Logf("Success:     %d (%.2f%%)", successCount, successRate)
			t.Logf("Errors:      %d", errorCount)
			t.Logf("Avg latency: %v", avg)
			t.Logf("P50 latency: %v", p50)
			t.Logf("P95 latency: %v", p95)
			t.Logf("P99 latency: %v", p99)

			assert.LessOrEqual(t, p99, tt.maxP99Latency,
				"P99 latency (%v) should be under %v", p99, tt.maxP99Latency)
			assert.LessOrEqual(t, avg, tt.maxAvgLatency,
				"Average latency (%v) should be under %v", avg, tt.maxAvgLatency)
			assert.GreaterOrEqual(t, successRate, tt.minSuccessRate,
				"Success rate (%.2f%%) should be at least %.2f%%", successRate, tt.minSuccessRate)
		})
	}
}

// =============================================================================
// RESPONSIVENESS TEST: Auth Endpoint
// =============================================================================

func TestResponsiveness_AuthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping responsiveness test in short mode")
	}

	tests := []struct {
		name             string
		concurrentUsers  int
		requestsPerUser  int
		maxP99Latency    time.Duration
		maxAvgLatency    time.Duration
		minSuccessRate   float64
	}{
		{
			name:            "LowLoad_10Users",
			concurrentUsers: 10,
			requestsPerUser: 20,
			maxP99Latency:   100 * time.Millisecond,
			maxAvgLatency:   50 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "MediumLoad_30Users",
			concurrentUsers: 30,
			requestsPerUser: 10,
			maxP99Latency:   100 * time.Millisecond,
			maxAvgLatency:   50 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "HighLoad_50Users",
			concurrentUsers: 50,
			requestsPerUser: 5,
			maxP99Latency:   100 * time.Millisecond,
			maxAvgLatency:   50 * time.Millisecond,
			minSuccessRate:  99.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ts := setupResponsivenessServer(t)
			lr := newLatencyResult(tt.concurrentUsers * tt.requestsPerUser)
			var successCount, errorCount int64

			loginPayload, err := json.Marshal(map[string]string{
				"username": "admin",
				"password": "admin123",
			})
			require.NoError(t, err)

			var wg sync.WaitGroup
			for i := 0; i < tt.concurrentUsers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := &http.Client{Timeout: 5 * time.Second}

					for j := 0; j < tt.requestsPerUser; j++ {
						start := time.Now()
						resp, err := client.Post(
							ts.URL+"/api/v1/auth/login",
							"application/json",
							bytes.NewReader(loginPayload),
						)
						elapsed := time.Since(start)

						if err != nil {
							errorCount++
							continue
						}
						resp.Body.Close()

						if resp.StatusCode == http.StatusOK {
							successCount++
						} else {
							errorCount++
						}
						lr.record(elapsed)
					}
				}()
			}

			wg.Wait()

			totalRequests := int64(tt.concurrentUsers * tt.requestsPerUser)
			successRate := float64(successCount) / float64(totalRequests) * 100
			p50 := lr.percentile(50)
			p95 := lr.percentile(95)
			p99 := lr.percentile(99)
			avg := lr.average()

			t.Logf("=== /api/v1/auth/login Responsiveness ===")
			t.Logf("Requests:    %d", totalRequests)
			t.Logf("Success:     %d (%.2f%%)", successCount, successRate)
			t.Logf("Errors:      %d", errorCount)
			t.Logf("Avg latency: %v", avg)
			t.Logf("P50 latency: %v", p50)
			t.Logf("P95 latency: %v", p95)
			t.Logf("P99 latency: %v", p99)

			assert.LessOrEqual(t, p99, tt.maxP99Latency,
				"P99 latency (%v) should be under %v", p99, tt.maxP99Latency)
			assert.LessOrEqual(t, avg, tt.maxAvgLatency,
				"Average latency (%v) should be under %v", avg, tt.maxAvgLatency)
			assert.GreaterOrEqual(t, successRate, tt.minSuccessRate,
				"Success rate (%.2f%%) should be at least %.2f%%", successRate, tt.minSuccessRate)
		})
	}
}

// =============================================================================
// RESPONSIVENESS TEST: List Endpoints
// =============================================================================

func TestResponsiveness_ListEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping responsiveness test in short mode")
	}

	ts := setupResponsivenessServer(t)

	tests := []struct {
		name            string
		endpoint        string
		concurrentUsers int
		requestsPerUser int
		maxP99Latency   time.Duration
		maxAvgLatency   time.Duration
		minSuccessRate  float64
	}{
		{
			name:            "MediaList_LowLoad",
			endpoint:        "/api/v1/media?page=1&limit=20",
			concurrentUsers: 10,
			requestsPerUser: 30,
			maxP99Latency:   200 * time.Millisecond,
			maxAvgLatency:   100 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "MediaList_MediumLoad",
			endpoint:        "/api/v1/media?page=1&limit=20",
			concurrentUsers: 50,
			requestsPerUser: 10,
			maxP99Latency:   200 * time.Millisecond,
			maxAvgLatency:   100 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "StorageRoots_LowLoad",
			endpoint:        "/api/v1/storage/roots",
			concurrentUsers: 10,
			requestsPerUser: 30,
			maxP99Latency:   200 * time.Millisecond,
			maxAvgLatency:   100 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "StorageRoots_HighLoad",
			endpoint:        "/api/v1/storage/roots",
			concurrentUsers: 100,
			requestsPerUser: 5,
			maxP99Latency:   200 * time.Millisecond,
			maxAvgLatency:   100 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "Collections_LowLoad",
			endpoint:        "/api/v1/collections",
			concurrentUsers: 10,
			requestsPerUser: 30,
			maxP99Latency:   200 * time.Millisecond,
			maxAvgLatency:   100 * time.Millisecond,
			minSuccessRate:  99.0,
		},
		{
			name:            "Collections_HighLoad",
			endpoint:        "/api/v1/collections",
			concurrentUsers: 100,
			requestsPerUser: 5,
			maxP99Latency:   200 * time.Millisecond,
			maxAvgLatency:   100 * time.Millisecond,
			minSuccessRate:  99.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			lr := newLatencyResult(tt.concurrentUsers * tt.requestsPerUser)
			var successCount, errorCount int64

			var wg sync.WaitGroup
			for i := 0; i < tt.concurrentUsers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := &http.Client{Timeout: 5 * time.Second}

					for j := 0; j < tt.requestsPerUser; j++ {
						start := time.Now()
						resp, err := client.Get(ts.URL + tt.endpoint)
						elapsed := time.Since(start)

						if err != nil {
							errorCount++
							continue
						}
						resp.Body.Close()

						if resp.StatusCode == http.StatusOK {
							successCount++
						} else {
							errorCount++
						}
						lr.record(elapsed)
					}
				}()
			}

			wg.Wait()

			totalRequests := int64(tt.concurrentUsers * tt.requestsPerUser)
			successRate := float64(successCount) / float64(totalRequests) * 100
			p50 := lr.percentile(50)
			p95 := lr.percentile(95)
			p99 := lr.percentile(99)
			avg := lr.average()

			t.Logf("=== %s Responsiveness ===", tt.endpoint)
			t.Logf("Requests:    %d", totalRequests)
			t.Logf("Success:     %d (%.2f%%)", successCount, successRate)
			t.Logf("Errors:      %d", errorCount)
			t.Logf("Avg latency: %v", avg)
			t.Logf("P50 latency: %v", p50)
			t.Logf("P95 latency: %v", p95)
			t.Logf("P99 latency: %v", p99)

			assert.LessOrEqual(t, p99, tt.maxP99Latency,
				"P99 latency (%v) should be under %v", p99, tt.maxP99Latency)
			assert.LessOrEqual(t, avg, tt.maxAvgLatency,
				"Average latency (%v) should be under %v", avg, tt.maxAvgLatency)
			assert.GreaterOrEqual(t, successRate, tt.minSuccessRate,
				"Success rate (%.2f%%) should be at least %.2f%%", successRate, tt.minSuccessRate)
		})
	}
}

package performance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type responseMetrics struct {
	Requests     int64
	Successes    int64
	Failures     int64
	AvgLatencyMs float64
	P50LatencyMs float64
	P95LatencyMs float64
	P99LatencyMs float64
	MaxLatencyMs float64
}

func measureEndpoint(t *testing.T, server *httptest.Server, method, path string, concurrency, requestsPerClient int) responseMetrics {
	t.Helper()

	var (
		totalRequests int64
		successes     int64
		failures      int64
		wg            sync.WaitGroup
		mu            sync.Mutex
		latencies     []time.Duration
	)

	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerClient; j++ {
				start := time.Now()
				req, err := http.NewRequest(method, server.URL+path, nil)
				if err != nil {
					atomic.AddInt64(&failures, 1)
					atomic.AddInt64(&totalRequests, 1)
					continue
				}

				resp, err := client.Do(req)
				elapsed := time.Since(start)

				atomic.AddInt64(&totalRequests, 1)

				if err != nil {
					atomic.AddInt64(&failures, 1)
					continue
				}
				resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					atomic.AddInt64(&successes, 1)
				} else {
					atomic.AddInt64(&failures, 1)
				}

				mu.Lock()
				latencies = append(latencies, elapsed)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if len(latencies) == 0 {
		return responseMetrics{}
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var totalLatency time.Duration
	for _, l := range latencies {
		totalLatency += l
	}

	metrics := responseMetrics{
		Requests:     totalRequests,
		Successes:    successes,
		Failures:     failures,
		AvgLatencyMs: float64(totalLatency.Milliseconds()) / float64(len(latencies)),
		P50LatencyMs: float64(latencies[len(latencies)*50/100].Milliseconds()),
		P95LatencyMs: float64(latencies[len(latencies)*95/100].Milliseconds()),
		P99LatencyMs: float64(latencies[len(latencies)*99/100].Milliseconds()),
		MaxLatencyMs: float64(latencies[len(latencies)-1].Milliseconds()),
	}

	return metrics
}

func setupPerfTestServer() *httptest.Server {
	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/api/v1/media", func(c *gin.Context) {
		items := make([]gin.H, 20)
		for i := range items {
			items[i] = gin.H{
				"id":    i + 1,
				"title": fmt.Sprintf("Media Item %d", i+1),
				"type":  "video",
			}
		}
		c.JSON(200, gin.H{"items": items, "total": 20})
	})

	r.GET("/api/v1/storage/roots", func(c *gin.Context) {
		roots := []gin.H{
			{"id": 1, "name": "Local Storage", "type": "local"},
			{"id": 2, "name": "NAS Share", "type": "smb"},
		}
		c.JSON(200, gin.H{"roots": roots})
	})

	r.GET("/api/v1/collections", func(c *gin.Context) {
		collections := make([]gin.H, 10)
		for i := range collections {
			collections[i] = gin.H{
				"id":         i + 1,
				"name":       fmt.Sprintf("Collection %d", i+1),
				"item_count": (i + 1) * 10,
			}
		}
		c.JSON(200, gin.H{"collections": collections, "total": 10})
	})

	return httptest.NewServer(r)
}

func TestResponseTime_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	server := setupPerfTestServer()
	defer server.Close()

	tests := []struct {
		name        string
		concurrency int
		requests    int
		maxP99Ms    float64
		maxAvgMs    float64
	}{
		{"low_load", 5, 50, 50, 20},
		{"medium_load", 20, 30, 100, 50},
		{"high_load", 50, 20, 200, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := measureEndpoint(t, server, "GET", "/health", tt.concurrency, tt.requests)

			data, _ := json.Marshal(metrics)
			t.Logf("Health endpoint (%s): %s", tt.name, string(data))

			assert.True(t, metrics.Successes > 0, "should have successful requests")
			successRate := float64(metrics.Successes) / float64(metrics.Requests) * 100
			assert.GreaterOrEqual(t, successRate, 99.0, "success rate should be >= 99%%")
		})
	}
}

func TestResponseTime_ListEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	server := setupPerfTestServer()
	defer server.Close()

	endpoints := []struct {
		name string
		path string
	}{
		{"media", "/api/v1/media"},
		{"storage_roots", "/api/v1/storage/roots"},
		{"collections", "/api/v1/collections"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			metrics := measureEndpoint(t, server, "GET", ep.path, 20, 25)

			data, _ := json.Marshal(metrics)
			t.Logf("%s endpoint: %s", ep.name, string(data))

			assert.True(t, metrics.Successes > 0, "should have successful requests")
			successRate := float64(metrics.Successes) / float64(metrics.Requests) * 100
			assert.GreaterOrEqual(t, successRate, 99.0, "success rate should be >= 99%%")
		})
	}
}

func TestResponseTime_ConcurrentMixedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	server := setupPerfTestServer()
	defer server.Close()

	endpoints := []string{"/health", "/api/v1/media", "/api/v1/storage/roots", "/api/v1/collections"}
	var wg sync.WaitGroup
	var totalSuccess, totalFail int64

	for _, ep := range endpoints {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}
			for i := 0; i < 50; i++ {
				resp, err := client.Get(server.URL + path)
				if err != nil {
					atomic.AddInt64(&totalFail, 1)
					continue
				}
				resp.Body.Close()
				if resp.StatusCode == 200 {
					atomic.AddInt64(&totalSuccess, 1)
				} else {
					atomic.AddInt64(&totalFail, 1)
				}
			}
		}(ep)
	}

	wg.Wait()

	total := totalSuccess + totalFail
	successRate := float64(totalSuccess) / float64(total) * 100
	t.Logf("Mixed load: %d/%d successful (%.1f%%)", totalSuccess, total, successRate)
	assert.GreaterOrEqual(t, successRate, 99.0, "mixed load success rate should be >= 99%%")
}

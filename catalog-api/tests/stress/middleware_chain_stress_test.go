package stress

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"catalogizer/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// STRESS TEST: Full Middleware Chain Under Concurrent Load
// =============================================================================

func setupMiddlewareStressServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Apply the full middleware stack
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestTimeout(30 * time.Second))
	r.Use(middleware.ConcurrencyLimiter(200))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
	})

	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func TestMiddlewareChainStress_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	router := setupMiddlewareStressServer()

	concurrency := 50
	requestsPerWorker := 100
	var successCount int64
	var failCount int64
	var totalLatency int64
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				reqStart := time.Now()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				req.RemoteAddr = fmt.Sprintf("192.0.2.%d:12345", workerID%254+1)
				router.ServeHTTP(w, req)
				latency := time.Since(reqStart)
				atomic.AddInt64(&totalLatency, int64(latency))

				if w.Code == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
					// Verify security headers are present on every response
					if w.Header().Get("X-Content-Type-Options") == "" {
						atomic.AddInt64(&failCount, 1)
					}
					if w.Header().Get("X-Request-ID") == "" {
						atomic.AddInt64(&failCount, 1)
					}
				} else {
					atomic.AddInt64(&failCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalRequests := int64(concurrency * requestsPerWorker)
	avgLatency := time.Duration(atomic.LoadInt64(&totalLatency) / totalRequests)
	rps := float64(totalRequests) / elapsed.Seconds()
	success := atomic.LoadInt64(&successCount)
	fails := atomic.LoadInt64(&failCount)

	t.Logf("Middleware Chain Stress Results:")
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Success: %d, Failures: %d", success, fails)
	t.Logf("  Success rate: %.2f%%", float64(success)/float64(totalRequests)*100)
	t.Logf("  Average latency: %v", avgLatency)
	t.Logf("  Throughput: %.0f req/s", rps)
	t.Logf("  Elapsed: %v", elapsed)

	assert.Equal(t, int64(0), fails, "No requests should fail through middleware chain")
	assert.Equal(t, totalRequests, success, "All requests should succeed")
	assert.Less(t, avgLatency, 10*time.Millisecond, "Average latency through middleware should be under 10ms")
}

func TestMiddlewareChainStress_SecurityHeadersConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	router := setupMiddlewareStressServer()

	concurrency := 30
	requestsPerWorker := 50
	var missingHeaders int64
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				router.ServeHTTP(w, req)

				headers := []string{
					"X-Content-Type-Options",
					"X-Frame-Options",
					"X-XSS-Protection",
					"Referrer-Policy",
					"Permissions-Policy",
				}
				for _, h := range headers {
					if w.Header().Get(h) == "" {
						atomic.AddInt64(&missingHeaders, 1)
					}
				}
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(0), missingHeaders,
		"All security headers must be present on every response under load")
}

func TestMiddlewareChainStress_RequestIDUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	router := setupMiddlewareStressServer()

	concurrency := 20
	requestsPerWorker := 100
	var mu sync.Mutex
	requestIDs := make(map[string]bool, concurrency*requestsPerWorker)
	var duplicates int64
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localIDs := make([]string, 0, requestsPerWorker)
			for j := 0; j < requestsPerWorker; j++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				router.ServeHTTP(w, req)
				id := w.Header().Get("X-Request-ID")
				if id != "" {
					localIDs = append(localIDs, id)
				}
			}
			// Batch insert to reduce lock contention
			mu.Lock()
			for _, id := range localIDs {
				if requestIDs[id] {
					atomic.AddInt64(&duplicates, 1)
				}
				requestIDs[id] = true
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	totalRequests := concurrency * requestsPerWorker
	assert.Equal(t, totalRequests, len(requestIDs), "Each request should get a unique ID")
	assert.Equal(t, int64(0), duplicates, "No duplicate request IDs should exist")
}

func TestMiddlewareChainStress_ConcurrencyLimiterBackpressure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.ConcurrencyLimiter(5)) // Very low limit

	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	concurrency := 20
	var okCount, rejectedCount int64
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/slow", nil)
			r.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				atomic.AddInt64(&okCount, 1)
			} else if w.Code == http.StatusServiceUnavailable {
				atomic.AddInt64(&rejectedCount, 1)
			}
		}()
	}

	wg.Wait()

	ok := atomic.LoadInt64(&okCount)
	rejected := atomic.LoadInt64(&rejectedCount)
	total := ok + rejected

	t.Logf("Backpressure results: OK=%d, Rejected=%d, Total=%d", ok, rejected, total)
	assert.Equal(t, int64(concurrency), total, "All requests should complete (either OK or rejected)")
	assert.Greater(t, ok, int64(0), "Some requests should succeed")
}

func TestMiddlewareChainStress_TimeoutUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestTimeout(100 * time.Millisecond))

	r.GET("/timeout", func(c *gin.Context) {
		select {
		case <-time.After(200 * time.Millisecond):
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		case <-c.Request.Context().Done():
			return
		}
	})

	r.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	concurrency := 20
	var timedOut, succeeded int64
	var wg sync.WaitGroup

	// Half hit /timeout, half hit /fast
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w := httptest.NewRecorder()
			path := "/fast"
			if id%2 == 0 {
				path = "/timeout"
			}
			req, _ := http.NewRequest("GET", path, nil)
			r.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				atomic.AddInt64(&succeeded, 1)
			} else {
				atomic.AddInt64(&timedOut, 1)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Timeout results: Succeeded=%d, TimedOut=%d", succeeded, timedOut)
	// Fast requests should always succeed
	assert.GreaterOrEqual(t, atomic.LoadInt64(&succeeded), int64(concurrency/2),
		"All fast requests should succeed")
}

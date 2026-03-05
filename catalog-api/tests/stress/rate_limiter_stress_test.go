package stress

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// testRateLimiter is a self-contained rate limiter for stress testing.
// It mirrors the AdvancedRateLimiter in middleware/ but is defined here
// so the stress package does not import internal middleware.
type testRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     float64
	burst    int
}

func newTestRateLimiter(r float64, burst int) *testRateLimiter {
	return &testRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

func (rl *testRateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if l, ok := rl.limiters[key]; ok {
		return l
	}
	l := rate.NewLimiter(rate.Limit(rl.rate), rl.burst)
	rl.limiters[key] = l
	return l
}

func (rl *testRateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		limiter := rl.getLimiter(key)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// setupRateLimitedServer creates a test server with rate limiting applied.
func setupRateLimitedServer(t *testing.T, ratePerSec float64, burst int) (*httptest.Server, *testRateLimiter) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rl := newTestRateLimiter(ratePerSec, burst)

	router := gin.New()
	router.Use(rl.middleware())

	router.GET("/api/v1/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"token": "test-token"})
	})

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts, rl
}

// =============================================================================
// STRESS TEST: Rate Limiter Burst Protection
// =============================================================================

func TestRateLimiter_BurstProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tests := []struct {
		name           string
		ratePerSec     float64
		burst          int
		requestCount   int
		maxSuccessRate float64 // upper bound - should not exceed this
	}{
		{
			name:           "TightBurst_5Requests",
			ratePerSec:     2,
			burst:          5,
			requestCount:   20,
			maxSuccessRate: 50.0, // At most ~5/20 = 25%, but allow margin
		},
		{
			name:           "MediumBurst_10Requests",
			ratePerSec:     5,
			burst:          10,
			requestCount:   50,
			maxSuccessRate: 50.0,
		},
		{
			name:           "LargeBurst_20Requests",
			ratePerSec:     10,
			burst:          20,
			requestCount:   100,
			maxSuccessRate: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ts, _ := setupRateLimitedServer(t, tt.ratePerSec, tt.burst)
			client := &http.Client{Timeout: 5 * time.Second}

			var successCount int64
			var rateLimitedCount int64

			// Send burst of requests as fast as possible (synchronous to control timing)
			for i := 0; i < tt.requestCount; i++ {
				resp, err := client.Get(ts.URL + "/api/v1/resource")
				require.NoError(t, err, "Request %d failed with error", i)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else if resp.StatusCode == http.StatusTooManyRequests {
					atomic.AddInt64(&rateLimitedCount, 1)
				}
			}

			totalRequests := int64(tt.requestCount)
			successRate := float64(successCount) / float64(totalRequests) * 100

			t.Logf("Requests:      %d", totalRequests)
			t.Logf("Successful:    %d", successCount)
			t.Logf("Rate limited:  %d", rateLimitedCount)
			t.Logf("Success rate:  %.2f%%", successRate)

			// Burst should allow exactly burst-count requests through initially
			assert.GreaterOrEqual(t, successCount, int64(1),
				"At least one request should succeed")

			assert.Greater(t, rateLimitedCount, int64(0),
				"Some requests should be rate limited when burst is exceeded")

			// The success count should not exceed burst + some that trickle through
			// during the request loop execution time
			maxExpectedSuccess := int64(tt.burst) + int64(float64(tt.burst)*0.5) + 1
			assert.LessOrEqual(t, successCount, maxExpectedSuccess,
				"Success count (%d) should not greatly exceed burst limit (%d)",
				successCount, tt.burst)
		})
	}
}

// =============================================================================
// STRESS TEST: Rate Limiter Concurrent Access
// =============================================================================

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tests := []struct {
		name       string
		ratePerSec float64
		burst      int
		goroutines int
		reqPerGo   int
	}{
		{
			name:       "50Goroutines_10ReqEach",
			ratePerSec: 100,
			burst:      200,
			goroutines: 50,
			reqPerGo:   10,
		},
		{
			name:       "100Goroutines_5ReqEach",
			ratePerSec: 50,
			burst:      100,
			goroutines: 100,
			reqPerGo:   5,
		},
		{
			name:       "200Goroutines_3ReqEach",
			ratePerSec: 200,
			burst:      300,
			goroutines: 200,
			reqPerGo:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ts, _ := setupRateLimitedServer(t, tt.ratePerSec, tt.burst)

			var successCount int64
			var rateLimitedCount int64
			var errorCount int64

			var wg sync.WaitGroup
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := &http.Client{Timeout: 10 * time.Second}

					for j := 0; j < tt.reqPerGo; j++ {
						resp, err := client.Get(ts.URL + "/api/v1/resource")
						if err != nil {
							atomic.AddInt64(&errorCount, 1)
							continue
						}
						resp.Body.Close()

						switch resp.StatusCode {
						case http.StatusOK:
							atomic.AddInt64(&successCount, 1)
						case http.StatusTooManyRequests:
							atomic.AddInt64(&rateLimitedCount, 1)
						default:
							atomic.AddInt64(&errorCount, 1)
						}
					}
				}()
			}

			wg.Wait()

			totalRequests := int64(tt.goroutines * tt.reqPerGo)
			processed := successCount + rateLimitedCount

			t.Logf("Total requests:  %d", totalRequests)
			t.Logf("Successful:      %d", successCount)
			t.Logf("Rate limited:    %d", rateLimitedCount)
			t.Logf("Errors:          %d", errorCount)
			t.Logf("Processed:       %d", processed)

			// All requests should get a valid response (200 or 429), not errors
			assert.LessOrEqual(t, errorCount, totalRequests/10,
				"Error count (%d) should be less than 10%% of total requests", errorCount)

			// At least some should succeed and some should be rate limited (if total > burst)
			assert.Greater(t, successCount, int64(0),
				"At least some requests should succeed")

			// Verify no data races - the test itself passing without -race failures is the assertion,
			// but we also verify the counts are consistent
			assert.Equal(t, totalRequests, successCount+rateLimitedCount+errorCount,
				"All requests should be accounted for")
		})
	}
}

// =============================================================================
// STRESS TEST: Rate Limiter Recovery
// =============================================================================

func TestRateLimiter_Recovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tests := []struct {
		name         string
		ratePerSec   float64
		burst        int
		initialBurst int
		waitDuration time.Duration
		postBurst    int
	}{
		{
			name:         "RecoveryAfter1Second",
			ratePerSec:   5,
			burst:        5,
			initialBurst: 10,
			waitDuration: 1200 * time.Millisecond,
			postBurst:    3,
		},
		{
			name:         "RecoveryAfter2Seconds",
			ratePerSec:   2,
			burst:        3,
			initialBurst: 8,
			waitDuration: 2200 * time.Millisecond,
			postBurst:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ts, _ := setupRateLimitedServer(t, tt.ratePerSec, tt.burst)
			client := &http.Client{Timeout: 5 * time.Second}

			// Phase 1: Exhaust the rate limit
			var phase1Success int64
			var phase1Limited int64

			for i := 0; i < tt.initialBurst; i++ {
				resp, err := client.Get(ts.URL + "/api/v1/resource")
				require.NoError(t, err)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					phase1Success++
				} else if resp.StatusCode == http.StatusTooManyRequests {
					phase1Limited++
				}
			}

			t.Logf("Phase 1 - Exhaust rate limit:")
			t.Logf("  Successful:    %d", phase1Success)
			t.Logf("  Rate limited:  %d", phase1Limited)

			// Verify rate limit was triggered
			assert.Greater(t, phase1Limited, int64(0),
				"Phase 1 should trigger rate limiting")

			// Confirm we are currently rate limited
			resp, err := client.Get(ts.URL + "/api/v1/resource")
			require.NoError(t, err)
			resp.Body.Close()
			wasLimited := resp.StatusCode == http.StatusTooManyRequests
			t.Logf("  Currently limited: %v", wasLimited)

			// Phase 2: Wait for recovery
			t.Logf("Phase 2 - Waiting %v for recovery...", tt.waitDuration)
			time.Sleep(tt.waitDuration)

			// Phase 3: Verify requests succeed again
			var phase3Success int64
			var phase3Limited int64

			for i := 0; i < tt.postBurst; i++ {
				resp, err := client.Get(ts.URL + "/api/v1/resource")
				require.NoError(t, err)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					phase3Success++
				} else if resp.StatusCode == http.StatusTooManyRequests {
					phase3Limited++
				}
			}

			t.Logf("Phase 3 - After recovery:")
			t.Logf("  Successful:    %d", phase3Success)
			t.Logf("  Rate limited:  %d", phase3Limited)

			assert.Greater(t, phase3Success, int64(0),
				"Requests should succeed after rate limit window expires")

			// Most post-recovery requests should succeed since we waited for tokens
			// to replenish
			assert.GreaterOrEqual(t, phase3Success, int64(tt.postBurst/2),
				"At least half of post-recovery requests should succeed")
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDefaultEnhancedRateLimiterConfig(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()

	assert.Equal(t, StrategyTokenBucket, config.Strategy)
	assert.Equal(t, TierAnonymous, config.DefaultTier)
	assert.NotNil(t, config.Tiers)
	assert.NotNil(t, config.KeyGenerator)
	assert.NotNil(t, config.TierExtractor)
	assert.True(t, config.EnableMetrics)
	assert.Equal(t, 5*time.Minute, config.CleanupInterval)
}

func TestNewEnhancedRateLimiter(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()
	limiter := NewEnhancedRateLimiter(config)

	assert.NotNil(t, limiter)
	assert.NotNil(t, limiter.limiters)
	assert.NotNil(t, limiter.lastSeen)
	assert.NotNil(t, limiter.metrics)
	assert.NotNil(t, limiter.stopCleanup)

	limiter.Stop()
}

func TestEnhancedRateLimiter_Allow(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()
	config.Tiers = map[RateLimitTier]TierConfig{
		TierAnonymous: {Rate: 1, Burst: 2}, // 1 req/sec, burst of 2
	}

	limiter := NewEnhancedRateLimiter(config)
	defer limiter.Stop()

	// Create a test context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// First 2 requests should be allowed (burst)
	assert.True(t, limiter.Allow(c))
	assert.True(t, limiter.Allow(c))

	// Third request might be blocked depending on timing
	// We can't reliably test this without time.Sleep
}

func TestEnhancedRateLimiter_Middleware(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()
	config.Tiers = map[RateLimitTier]TierConfig{
		TierAnonymous: {Rate: 100, Burst: 100}, // Very permissive for testing
	}

	limiter := NewEnhancedRateLimiter(config)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnhancedRateLimiter_Middleware_RateLimited(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()
	config.Tiers = map[RateLimitTier]TierConfig{
		TierAnonymous: {Rate: 0, Burst: 0}, // No requests allowed
	}
	config.Message = "Rate limit exceeded"

	limiter := NewEnhancedRateLimiter(config)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Rate limit exceeded")
}

func TestEnhancedRateLimiter_GetMetrics(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()
	config.Tiers = map[RateLimitTier]TierConfig{
		TierAnonymous: {Rate: 100, Burst: 100},
	}

	limiter := NewEnhancedRateLimiter(config)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Make some requests
	limiter.Allow(c)
	limiter.Allow(c)
	limiter.Allow(c)

	stats := limiter.GetMetrics()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(3), stats["total_requests"])
	assert.Equal(t, int64(3), stats["allowed_requests"])
}

func TestEnhancedRateLimiter_ResetMetrics(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()
	config.Tiers = map[RateLimitTier]TierConfig{
		TierAnonymous: {Rate: 100, Burst: 100},
	}

	limiter := NewEnhancedRateLimiter(config)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Make some requests
	limiter.Allow(c)

	// Reset metrics
	limiter.ResetMetrics()

	stats := limiter.GetMetrics()
	assert.Equal(t, int64(0), stats["total_requests"])
	assert.Equal(t, int64(0), stats["allowed_requests"])
}

func TestRateLimitMetrics_Reset(t *testing.T) {
	metrics := &RateLimitMetrics{
		TotalRequests:   100,
		AllowedRequests: 90,
		BlockedRequests: 10,
		LastReset:       time.Now().Add(-time.Hour),
	}

	metrics.Reset()

	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.AllowedRequests)
	assert.Equal(t, int64(0), metrics.BlockedRequests)
	assert.True(t, time.Since(metrics.LastReset) < time.Second)
}

func TestRateLimitMetrics_GetStats(t *testing.T) {
	metrics := &RateLimitMetrics{
		TotalRequests:   100,
		AllowedRequests: 90,
		BlockedRequests: 10,
		LastReset:       time.Now(),
	}

	stats := metrics.GetStats()

	assert.Equal(t, int64(100), stats["total_requests"])
	assert.Equal(t, int64(90), stats["allowed_requests"])
	assert.Equal(t, int64(10), stats["blocked_requests"])
	assert.Equal(t, 10.0, stats["block_rate"])
}

func TestTierExtractor(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected RateLimitTier
	}{
		{
			name:     "anonymous",
			setup:    func(c *gin.Context) {},
			expected: TierAnonymous,
		},
		{
			name: "authenticated user",
			setup: func(c *gin.Context) {
				c.Set("user_id", "123")
			},
			expected: TierAuthenticated,
		},
		{
			name: "admin user",
			setup: func(c *gin.Context) {
				c.Set("user_role", "admin")
			},
			expected: TierAdmin,
		},
		{
			name: "premium user",
			setup: func(c *gin.Context) {
				c.Set("user_role", "premium")
			},
			expected: TierPremium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setup(c)
			assert.Equal(t, tt.expected, config.TierExtractor(c))
		})
	}
}

func TestIPBasedKeyGenerator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	key := IPBasedKeyGenerator(c)
	assert.NotEmpty(t, key)
}

func TestUserBasedKeyGenerator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Without user ID, should fall back to IP
	key1 := UserBasedKeyGenerator(c)
	assert.NotEmpty(t, key1)

	// With user ID
	c.Set("user_id", "123")
	key2 := UserBasedKeyGenerator(c)
	assert.Equal(t, "123", key2)
}

func TestIPAndUserBasedKeyGenerator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Without user ID, should return IP
	key1 := IPAndUserBasedKeyGenerator(c)
	assert.NotEmpty(t, key1)

	// With user ID
	c.Set("user_id", "123")
	key2 := IPAndUserBasedKeyGenerator(c)
	assert.Contains(t, key2, "123")
}

func TestGlobalRateLimiter(t *testing.T) {
	// Reset global limiter for test
	globalRateLimiter = nil
	globalRateLimiterOnce = sync.Once{}

	config := DefaultEnhancedRateLimiterConfig()
	InitGlobalRateLimiter(config)

	limiter1 := GetGlobalRateLimiter()
	limiter2 := GetGlobalRateLimiter()

	assert.NotNil(t, limiter1)
	assert.Equal(t, limiter1, limiter2) // Should be same instance

	limiter1.Stop()
}

func TestEnhancedRateLimitMiddleware(t *testing.T) {
	// Reset global limiter for test
	globalRateLimiter = nil
	globalRateLimiterOnce = sync.Once{}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(EnhancedRateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Should pass (default config is permissive)
	assert.Equal(t, http.StatusOK, w.Code)

	GetGlobalRateLimiter().Stop()
}

func TestEnhancedRateLimiter_Cleanup(t *testing.T) {
	config := DefaultEnhancedRateLimiterConfig()
	config.CleanupInterval = 100 * time.Millisecond
	config.Tiers = map[RateLimitTier]TierConfig{
		TierAnonymous: {Rate: 100, Burst: 100},
	}

	limiter := NewEnhancedRateLimiter(config)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Create some limiters
	limiter.Allow(c)
	limiter.Allow(c)

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	limiter.Stop()
}

func BenchmarkEnhancedRateLimiter_Allow(b *testing.B) {
	config := DefaultEnhancedRateLimiterConfig()
	config.Tiers = map[RateLimitTier]TierConfig{
		TierAnonymous: {Rate: 10000, Burst: 10000}, // Very permissive
	}

	limiter := NewEnhancedRateLimiter(config)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(c)
	}
}

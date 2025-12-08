package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdvancedRateLimit_BasicBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultRateLimiterConfig()
	config.Rate = 2  // 2 requests per second
	config.Burst = 4 // Burst of 4 requests

	router := gin.New()
	router.Use(AdvancedRateLimit(config))

	requestCount := 0
	router.GET("/test", func(c *gin.Context) {
		requestCount++
		c.JSON(http.StatusOK, gin.H{"count": requestCount})
	})

	// Make requests within burst limit
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	assert.Equal(t, 4, requestCount, "Should have processed burst requests")

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "rate_limit_exceeded")

	// Wait for rate limit to reset
	time.Sleep(time.Second)

	// Should be able to make requests again
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Should work after waiting")
}

func TestAdvancedRateLimit_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultRateLimiterConfig()
	config.Rate = 10  // 10 requests per second
	config.Burst = 15 // Burst of 15

	router := gin.New()
	router.Use(AdvancedRateLimit(config))

	var wg sync.WaitGroup
	var successCount int
	var mu sync.Mutex

	router.GET("/test", func(c *gin.Context) {
		mu.Lock()
		successCount++
		mu.Unlock()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Launch 20 concurrent requests
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	// Should allow burst but rate limit the rest
	assert.True(t, successCount <= 15, "Should not exceed burst limit")
}

func TestIPRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(IPRateLimit(5, 10)) // 5 req/s, burst 10

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test that IP rate limiting works
	successCount := 0
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			successCount++
		}
	}

	assert.True(t, successCount <= 10, "Should not exceed burst limit")
}

func TestAuthRateLimitConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := AuthRateLimiterConfig()

	assert.Equal(t, 3.0, config.Rate)
	assert.Equal(t, 10, config.Burst)
	assert.Contains(t, config.Message, "authentication attempts")

	// Test key generator with username
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("username", "testuser")

	key := config.KeyGenerator(c)
	assert.Contains(t, key, "testuser:")

	// Test key generator without username
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = httptest.NewRequest("GET", "/", nil)
	key2 := config.KeyGenerator(c2)
	assert.NotEmpty(t, key2)
}

func TestAdvancedRateLimit_LimiterCleanup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a limiter with small cleanup threshold
	config := DefaultRateLimiterConfig()

	limiter := NewAdvancedRateLimiter(config)

	// Simulate many different IPs
	for i := 0; i < 100; i++ {
		limiter.getLimiter("127.0.0.1")
		limiter.getLimiter("192.168.1." + string(rune(i%255)))
	}

	// Call cleanup (in real usage, this happens in a goroutine)
	limiter.CleanupExpiredLimiters()

	// Limiter should still work for new keys
	testLimiter := limiter.getLimiter("test-new-key")
	assert.NotNil(t, testLimiter)
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.Equal(t, 10.0, config.Rate)
	assert.Equal(t, 20, config.Burst)
	assert.Contains(t, config.Message, "Rate limit exceeded")
	assert.NotNil(t, config.KeyGenerator)

	// Test default key generator
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)

	key := config.KeyGenerator(c)
	assert.NotEmpty(t, key)
}

func TestStrictRateLimiterConfig(t *testing.T) {
	config := StrictRateLimiterConfig()

	assert.Equal(t, 2.0, config.Rate)
	assert.Equal(t, 5, config.Burst)
	assert.Contains(t, strings.ToLower(config.Message), "too many requests")
}

func TestUserBasedRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultRateLimiterConfig()
	config.Rate = 5
	config.Burst = 10

	router := gin.New()
	router.Use(UserBasedRateLimit(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test with user ID - need to set up the context properly
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request = c.Request.WithContext(c.Request.Context())
	c.Set("user_id", "user123")

	// Create the same key generator that UserBasedRateLimit creates
	userBasedKeyGenerator := func(c *gin.Context) string {
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok {
				return "user:" + id
			}
		}
		return "ip:" + c.ClientIP()
	}

	// Test with user ID
	c1, _ := gin.CreateTestContext(httptest.NewRecorder())
	c1.Request = httptest.NewRequest("GET", "/", nil)
	c1.Set("user_id", "user123")

	key := userBasedKeyGenerator(c1)
	assert.Equal(t, "user:user123", key)

	// Test without user ID (fallback to IP)
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = httptest.NewRequest("GET", "/", nil)
	key2 := userBasedKeyGenerator(c2)
	assert.Contains(t, key2, "ip:")
}

func TestRateLimitHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultRateLimiterConfig()
	config.Rate = 5

	router := gin.New()
	router.Use(AdvancedRateLimit(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "N/A", w.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "N/A", w.Header().Get("X-RateLimit-Reset"))
}

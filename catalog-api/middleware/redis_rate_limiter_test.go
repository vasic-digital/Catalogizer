package middleware

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRedisRateLimiterConfig(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: ":6379"})
	config := DefaultRedisRateLimiterConfig(client)
	
	assert.Equal(t, 100, config.Requests)
	assert.Equal(t, time.Minute, config.Window)
	assert.NotEmpty(t, config.Message)
	assert.NotNil(t, config.Client)
	assert.NotNil(t, config.KeyGenerator)
}

func TestStrictRedisRateLimiterConfig(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: ":6379"})
	config := StrictRedisRateLimiterConfig(client)
	
	assert.Equal(t, 10, config.Requests)
	assert.Equal(t, time.Minute, config.Window)
	assert.NotEmpty(t, config.Message)
	assert.NotNil(t, config.Client)
	assert.NotNil(t, config.KeyGenerator)
}

func TestAuthRedisRateLimiterConfig(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: ":6379"})
	config := AuthRedisRateLimiterConfig(client)
	
	assert.Equal(t, 5, config.Requests)
	assert.Equal(t, time.Minute, config.Window)
	assert.NotEmpty(t, config.Message)
	assert.NotNil(t, config.Client)
	assert.NotNil(t, config.KeyGenerator)
}

func TestUserRedisRateLimiterConfig(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: ":6379"})
	config := UserRedisRateLimiterConfig(client)
	
	assert.Equal(t, 200, config.Requests)
	assert.Equal(t, time.Minute, config.Window)
	assert.NotEmpty(t, config.Message)
	assert.NotNil(t, config.Client)
	assert.NotNil(t, config.KeyGenerator)
}

// Integration test with actual Redis
func TestRedisRateLimit_Integration(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for integration test")
	}
	
	// Clean up after test
	defer client.Del(ctx, "rate_limit:test_ip")
	
	gin.SetMode(gin.TestMode)
	config := DefaultRedisRateLimiterConfig(client)
	config.Requests = 3 // Low limit for testing
	config.Window = time.Second * 2 // Short window for testing
	config.KeyGenerator = func(c *gin.Context) string {
		return "test_ip"
	}
	
	middleware := RedisRateLimit(config)
	
	// First request should pass
	c, _ := gin.CreateTestContext(nil)
	c.Request = createTestRequest()
	c.Set("test_key", "test_value")
	
	middleware(c)
	assert.False(t, c.IsAborted())
	
	// Second request should pass
	c2, _ := gin.CreateTestContext(nil)
	c2.Request = createTestRequest()
	
	middleware(c2)
	assert.False(t, c2.IsAborted())
	
	// Third request should pass
	c3, _ := gin.CreateTestContext(nil)
	c3.Request = createTestRequest()
	
	middleware(c3)
	assert.False(t, c3.IsAborted())
	
	// Fourth request should be rate limited
	c4, _ := gin.CreateTestContext(nil)
	c4.Request = createTestRequest()
	
	middleware(c4)
	assert.True(t, c4.IsAborted())
	assert.Equal(t, 429, c4.Writer.Status())
}

func TestSlidingWindowRedisRateLimit_Integration(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for integration test")
	}
	
	// Clean up after test
	defer client.Del(ctx, "sliding_rate_limit:test_ip")
	
	gin.SetMode(gin.TestMode)
	config := DefaultRedisRateLimiterConfig(client)
	config.Requests = 3
	config.Window = time.Second
	config.KeyGenerator = func(c *gin.Context) string {
		return "test_ip"
	}
	
	middleware := SlidingWindowRedisRateLimit(config)
	
	// Make requests that should pass
	for i := 0; i < 3; i++ {
		c, _ := gin.CreateTestContext(nil)
		c.Request = createTestRequest()
		
		middleware(c)
		assert.False(t, c.IsAborted(), "Request %d should pass", i+1)
		
		// Small delay to spread out requests
		time.Sleep(time.Millisecond * 100)
	}
	
	// Next request should be rate limited
	c, _ := gin.CreateTestContext(nil)
	c.Request = createTestRequest()
	
	middleware(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, 429, c.Writer.Status())
}

func TestTokenBucketRedisRateLimit_Integration(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for integration test")
	}
	
	// Clean up after test
	defer client.Del(ctx, "token_bucket:test_ip")
	
	gin.SetMode(gin.TestMode)
	
	middleware := TokenBucketRedisRateLimit(
		client,
		func(c *gin.Context) string { return "test_ip" },
		5,    // Max tokens
		1,    // Refill rate: 1 token per second
		time.Second, // Refill interval
	)
	
	// Make requests that should consume tokens
	for i := 0; i < 5; i++ {
		c, _ := gin.CreateTestContext(nil)
		c.Request = createTestRequest()
		
		middleware(c)
		assert.False(t, c.IsAborted(), "Request %d should pass", i+1)
	}
	
	// Next request should be rate limited (bucket empty)
	c, _ := gin.CreateTestContext(nil)
	c.Request = createTestRequest()
	
	middleware(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, 429, c.Writer.Status())
	
	// Wait for token refill
	time.Sleep(time.Second * 2)
	
	// Request should pass again
	c2, _ := gin.CreateTestContext(nil)
	c2.Request = createTestRequest()
	
	middleware(c2)
	assert.False(t, c2.IsAborted())
}

func TestRedisRateLimit_ErrorHandling(t *testing.T) {
	// Use invalid Redis client to test error handling
	client := redis.NewClient(&redis.Options{
		Addr: "invalid-host:6379",
	})
	
	gin.SetMode(gin.TestMode)
	config := DefaultRedisRateLimiterConfig(client)
	middleware := RedisRateLimit(config)
	
	// Request should pass even with Redis error (fail open)
	c, _ := gin.CreateTestContext(nil)
	c.Request = createTestRequest()
	
	middleware(c)
	assert.False(t, c.IsAborted()) // Should not abort due to Redis error
}

func TestAuthRateLimiterKeyGeneration(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: ":6379"})
	config := AuthRedisRateLimiterConfig(client)
	
	// Test key generation without username
	c, _ := gin.CreateTestContext(nil)
	c.Request = createTestRequest()
	key := config.KeyGenerator(c)
	assert.Equal(t, "127.0.0.1", key)
	
	// Test key generation with username
	c2, _ := gin.CreateTestContext(nil)
	c2.Request = createTestRequest()
	c2.Set("username", "testuser")
	key2 := config.KeyGenerator(c2)
	assert.Equal(t, "testuser:127.0.0.1", key2)
}

func TestUserRateLimiterKeyGeneration(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: ":6379"})
	config := UserRedisRateLimiterConfig(client)
	
	// Test key generation without user ID (should fall back to IP)
	c, _ := gin.CreateTestContext(nil)
	c.Request = createTestRequest()
	key := config.KeyGenerator(c)
	assert.Equal(t, "ip:127.0.0.1", key)
	
	// Test key generation with user ID
	c2, _ := gin.CreateTestContext(nil)
	c2.Request = createTestRequest()
	c2.Set("user_id", "123")
	key2 := config.KeyGenerator(c2)
	assert.Equal(t, "user:123", key2)
}

// Helper function to create test HTTP request
func createTestRequest() *http.Request {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	return req
}
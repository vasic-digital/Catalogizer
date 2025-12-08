package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestRedisRateLimit_FailClosed demonstrates the fix for security vulnerability
func TestRedisRateLimit_FailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a Redis client that will always fail
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis://invalid-host:6379", // Invalid address to force failures
		Password: "",
		DB:       0,
	})

	config := DefaultRedisRateLimiterConfig(redisClient)
	config.Requests = 5
	config.Window = time.Minute

	router := gin.New()
	router.Use(RedisRateLimit(config))

	requestCount := 0
	router.GET("/test", func(c *gin.Context) {
		requestCount++
		c.JSON(http.StatusOK, gin.H{"count": requestCount})
	})

	// Make a request - should fail with service unavailable due to Redis failure
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The request should NOT be processed (fail closed for security)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code, 
		"Request should be rejected when Redis is unavailable for security")
	assert.Contains(t, w.Body.String(), "Rate limiting service temporarily unavailable")
	
	// No requests should have been processed
	assert.Equal(t, 0, requestCount, "No requests should be processed when Redis fails")
}

// Helper to create a test rate limiter with Redis failure
func createRateLimiterWithFailingRedis() gin.HandlerFunc {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis://invalid-host:6379", // Invalid address to force failures
		Password: "",
		DB:       0,
	})
	config := DefaultRedisRateLimiterConfig(redisClient)
	config.Requests = 5
	config.Window = time.Minute
	return RedisRateLimit(config)
}
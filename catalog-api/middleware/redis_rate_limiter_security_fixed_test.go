package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisRateLimit_FailClosed demonstrates the fix for security vulnerability
func TestRedisRateLimit_FailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a Redis client that will always fail
	redisClient := createFailingRedisClient()

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

// Helper function to create a Redis client that will fail
func createFailingRedisClient() *RedisClient {
	// This is a mock implementation that always fails
	return &RedisClient{}
}

// Mock RedisClient for testing
type RedisClient struct {
	// This would implement the necessary Redis client interface
	// For this test, we just need a type that won't actually connect
}

// Helper to create a test rate limiter with Redis failure
func createRateLimiterWithFailingRedis() gin.HandlerFunc {
	redisClient := createFailingRedisClient()
	config := DefaultRedisRateLimiterConfig(redisClient)
	config.Requests = 5
	config.Window = time.Minute
	return RedisRateLimit(config)
}
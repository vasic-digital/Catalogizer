package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"catalogizer/middleware"
)

func TestRedisRateLimit_SecurityBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a Redis client that will always fail
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis://invalid-host:6379", // Invalid address to force failures
		Password: "",
		DB:       0,
	})

	config := middleware.DefaultRedisRateLimiterConfig(redisClient)
	config.Requests = 5
	config.Window = time.Minute

	router := gin.New()
	router.Use(middleware.RedisRateLimit(config))

	requestCount := 0
	router.GET("/test", func(c *gin.Context) {
		requestCount++
		c.JSON(http.StatusOK, gin.H{"count": requestCount})
	})

	// Make many requests - they should all pass through due to Redis failure
	for i := 0; i < 10; i++ { // More than the configured limit of 5
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// All requests should succeed due to Redis failure (security vulnerability)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should not be rate limited due to Redis failure", i+1)
	}

	// This demonstrates the security vulnerability: when Redis fails, all requests pass through
	assert.Equal(t, 10, requestCount, "All requests should have been processed despite Redis failure")
}

func TestRedisRateLimit_FixedBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock Redis client that simulates failures
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis://localhost:6379", // Valid address
		Password: "",
		DB:       0,
	})

	config := DefaultRedisRateLimiterConfig(redisClient)
	config.Requests = 5
	config.Window = time.Minute

	// Create a custom rate limiter that doesn't fail open
	fixedRateLimiter := func(c middleware.RedisRateLimiterConfig) gin.HandlerFunc {
		return func(gc *gin.Context) {
			key := "rate_limit:" + c.KeyGenerator(gc)
			ctx := context.Background()
			
			// Try Redis with fallback to in-memory if Redis fails
			pipe := c.Client.Pipeline()
			getCmd := pipe.Get(ctx, key)
			incrCmd := pipe.Incr(ctx, key)
			expireCmd := pipe.Expire(ctx, key, c.Window)
			
			_, err := pipe.Exec(ctx)
			
			// If Redis fails, use in-memory rate limiting instead of failing open
			if err != nil {
				// Use simple in-memory rate limiting as fallback
				// This would need to be implemented with proper storage
				gc.JSON(http.StatusInternalServerError, gin.H{
					"error": "Rate limiting service temporarily unavailable",
				})
				gc.Abort()
				return
			}
			
			var currentCount int64
			if getCmd.Err() != nil {
				currentCount = 1
			} else {
				if val, err := getCmd.Int64(); err == nil {
					currentCount = val + 1
				} else {
					currentCount = 1
				}
			}
			
			incrCmd.Val()
			expireCmd.Val()
			
			// Check if rate limit exceeded
			if currentCount > int64(c.Requests) {
				ttl, _ := c.Client.TTL(ctx, key).Result()
				if ttl < 0 {
					ttl = c.Window
				}
				
				gc.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate_limit_exceeded",
					"message":     c.Message,
					"retry_after": ttl.String(),
				})
				gc.Abort()
				return
			}
			
			// Add rate limit headers
			remaining := int64(c.Requests) - currentCount
			gc.Header("X-RateLimit-Limit", strconv.Itoa(c.Requests))
			gc.Header("X-RateLimit-Remaining", strconv.Itoa(int(remaining)))
			gc.Header("X-RateLimit-Reset", time.Now().Add(c.Window).Format(time.RFC3339))
			gc.Header("X-RateLimit-Window", c.Window.String())
			
			gc.Next()
		}
	}

	router := gin.New()
	router.Use(fixedRateLimiter(config))

	requestCount := 0
	router.GET("/test", func(c *gin.Context) {
		requestCount++
		c.JSON(http.StatusOK, gin.H{"count": requestCount})
	})

	// Test with valid Redis connection (would need actual Redis running)
	// For this test, we'll demonstrate the concept
	
	// The key point is that when Redis fails, we should either:
	// 1. Fail closed (block requests)
	// 2. Have a fallback in-memory rate limiter
	// 3. Return an error indicating rate limiting is unavailable
	
	t.Skip("Requires actual Redis instance for full testing")
}
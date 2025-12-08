package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimiterConfig holds configuration for Redis-based rate limiting
type RedisRateLimiterConfig struct {
	// Requests per window
	Requests int
	// Window duration (e.g., time.Minute, time.Hour)
	Window time.Duration
	// Key generator function (e.g., by IP, user ID, etc.)
	KeyGenerator func(*gin.Context) string
	// Message to return when rate limited
	Message string
	// Redis client
	Client *redis.Client
}

// DefaultRedisRateLimiterConfig returns a secure default configuration
func DefaultRedisRateLimiterConfig(client *redis.Client) RedisRateLimiterConfig {
	return RedisRateLimiterConfig{
		Requests: 100, // 100 requests per window
		Window:   time.Minute,
		Message:  "Rate limit exceeded. Please try again later.",
		Client:   client,
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
}

// StrictRedisRateLimiterConfig returns a stricter configuration for sensitive endpoints
func StrictRedisRateLimiterConfig(client *redis.Client) RedisRateLimiterConfig {
	return RedisRateLimiterConfig{
		Requests: 10, // 10 requests per window
		Window:   time.Minute,
		Message:  "Too many requests. Please wait before trying again.",
		Client:   client,
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
}

// AuthRedisRateLimiterConfig returns a rate limiter specifically for auth endpoints
func AuthRedisRateLimiterConfig(client *redis.Client) RedisRateLimiterConfig {
	return RedisRateLimiterConfig{
		Requests: 5, // 5 login attempts per minute
		Window:   time.Minute,
		Message:  "Too many authentication attempts. Account temporarily locked.",
		Client:   client,
		KeyGenerator: func(c *gin.Context) string {
			// Use both IP and username for auth endpoints if available
			key := c.ClientIP()
			if username, exists := c.Get("username"); exists {
				if userStr, ok := username.(string); ok {
					key = userStr + ":" + key
				}
			}
			return key
		},
	}
}

// UserRedisRateLimiterConfig returns user-based rate limiting
func UserRedisRateLimiterConfig(client *redis.Client) RedisRateLimiterConfig {
	return RedisRateLimiterConfig{
		Requests: 200, // 200 requests per minute for authenticated users
		Window:   time.Minute,
		Message:  "Rate limit exceeded. Please try again later.",
		Client:   client,
		KeyGenerator: func(c *gin.Context) string {
			// Try to get user ID from context (set by auth middleware)
			if userID, exists := c.Get("user_id"); exists {
				if id, ok := userID.(string); ok {
					return "user:" + id
				}
			}
			// Fallback to IP for unauthenticated requests
			return "ip:" + c.ClientIP()
		},
	}
}

// RedisRateLimit creates a Gin middleware for Redis-based rate limiting
func RedisRateLimit(config RedisRateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rate_limit:" + config.KeyGenerator(c)
		ctx := context.Background()

		// Use Redis pipeline for atomic operations
		pipe := config.Client.Pipeline()

		// Get current count
		getCmd := pipe.Get(ctx, key)

		// Increment count and set expiration if key doesn't exist
		incrCmd := pipe.Incr(ctx, key)
		expireCmd := pipe.Expire(ctx, key, config.Window)

		// Execute pipeline
		_, err := pipe.Exec(ctx)

		// Handle Redis errors gracefully - fail closed for security
		if err != nil {
			// Log the Redis error
			fmt.Printf("Redis rate limiter error: %v\n", err)

			// When Redis is unavailable, we fail closed for security
			// This prevents bypassing rate limiting when Redis is down
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Rate limiting service temporarily unavailable",
				"message": "Please try again later",
			})
			c.Abort()
			return
		}

		var currentCount int64
		if getCmd.Err() != nil {
			// Key doesn't exist, this is the first request
			currentCount = 1
		} else {
			// Parse existing count
			if val, err := getCmd.Int64(); err == nil {
				currentCount = val + 1
			} else {
				currentCount = 1
			}
		}

		// Wait for increment and expire commands to complete
		incrCmd.Val()
		expireCmd.Val()

		// Check if rate limit exceeded
		if currentCount > int64(config.Requests) {
			// Get TTL for retry-after header
			ttl, _ := config.Client.TTL(ctx, key).Result()
			if ttl < 0 {
				ttl = config.Window
			}

			// Log rate limit attempt
			fmt.Printf("Rate limit exceeded for key: %s, IP: %s, Path: %s\n",
				key, c.ClientIP(), c.Request.URL.Path)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     config.Message,
				"retry_after": ttl.String(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := int64(config.Requests) - currentCount
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(int(remaining)))
		c.Header("X-RateLimit-Reset", time.Now().Add(config.Window).Format(time.RFC3339))
		c.Header("X-RateLimit-Window", config.Window.String())

		c.Next()
	}
}

// SlidingWindowRedisRateLimit implements sliding window rate limiting using Redis
func SlidingWindowRedisRateLimit(config RedisRateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "sliding_rate_limit:" + config.KeyGenerator(c)
		ctx := context.Background()
		now := time.Now().UnixNano()
		windowStart := now - config.Window.Nanoseconds()

		// Clean up old entries and add current request timestamp
		pipe := config.Client.Pipeline()

		// Remove old entries outside the window
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

		// Add current request
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: now,
		})

		// Count current requests in window
		countCmd := pipe.ZCard(ctx, key)

		// Set expiration for the key
		pipe.Expire(ctx, key, config.Window*2)

		// Execute pipeline
		_, err := pipe.Exec(ctx)

		// Handle Redis errors gracefully
		if err != nil {
			fmt.Printf("Redis sliding window rate limiter error: %v\n", err)
			c.Next()
			return
		}

		// Get current count
		count, err := countCmd.Result()
		if err != nil {
			count = 1
		}

		// Check if rate limit exceeded
		if count > int64(config.Requests) {
			// Get TTL for retry-after header
			ttl, _ := config.Client.TTL(ctx, key).Result()
			if ttl < 0 {
				ttl = config.Window
			}

			// Log rate limit attempt
			fmt.Printf("Sliding window rate limit exceeded for key: %s, IP: %s, Path: %s\n",
				key, c.ClientIP(), c.Request.URL.Path)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     config.Message,
				"retry_after": ttl.String(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := int64(config.Requests) - count
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(int(remaining)))
		c.Header("X-RateLimit-Reset", time.Now().Add(config.Window).Format(time.RFC3339))
		c.Header("X-RateLimit-Window", config.Window.String())
		c.Header("X-RateLimit-Algorithm", "sliding-window")

		c.Next()
	}
}

// TokenBucketRedisRateLimit implements token bucket algorithm using Redis
func TokenBucketRedisRateLimit(
	redisClient *redis.Client,
	keyGenerator func(*gin.Context) string,
	tokens int, // Maximum tokens in bucket
	refillRate int, // Tokens to add per second
	refillInterval time.Duration, // How often to add tokens
) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "token_bucket:" + keyGenerator(c)
		ctx := context.Background()

		// Use Lua script for atomic token bucket operations
		luaScript := `
			local key = KEYS[1]
			local current_time = tonumber(ARGV[1])
			local tokens = tonumber(ARGV[2])
			local refill_rate = tonumber(ARGV[3])
			local refill_interval = tonumber(ARGV[4])
			local requested = tonumber(ARGV[5])
			
			local bucket = redis.call('HMGET', key, 'tokens', 'last_refill', 'max_tokens')
			local current_tokens = tonumber(bucket[1]) or tokens
			local last_refill = tonumber(bucket[2]) or current_time
			local max_tokens = tonumber(bucket[3]) or tokens
			
			-- Calculate tokens to add based on elapsed time
			local elapsed = current_time - last_refill
			local tokens_to_add = math.floor(elapsed / refill_interval) * refill_rate
			current_tokens = math.min(max_tokens, current_tokens + tokens_to_add)
			
			-- Check if enough tokens are available
			if current_tokens >= requested then
				-- Consume tokens
				current_tokens = current_tokens - requested
				redis.call('HMSET', key, 'tokens', current_tokens, 'last_refill', current_time, 'max_tokens', max_tokens)
				redis.call('EXPIRE', key, refill_interval * 2)
				return {1, current_tokens, max_tokens}
			else
				-- Not enough tokens, update last_refill but don't consume
				redis.call('HMSET', key, 'tokens', current_tokens, 'last_refill', current_time, 'max_tokens', max_tokens)
				redis.call('EXPIRE', key, refill_interval * 2)
				return {0, current_tokens, max_tokens}
			end
		`

		now := time.Now().Unix()
		result, err := redisClient.Eval(ctx, luaScript, []string{key},
			now, tokens, refillRate, refillInterval.Seconds(), 1).Result()

		if err != nil {
			fmt.Printf("Redis token bucket rate limiter error: %v\n", err)
			c.Next()
			return
		}

		// Parse Lua script result
		if arr, ok := result.([]interface{}); ok && len(arr) >= 3 {
			allowed := arr[0].(int64)
			currentTokens := arr[1].(int64)
			maxTokens := arr[2].(int64)

			if allowed == 0 {
				// Rate limit exceeded
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate_limit_exceeded",
					"message":     "Rate limit exceeded. Please try again later.",
					"retry_after": refillInterval.String(),
				})
				c.Abort()
				return
			}

			// Add rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(int(maxTokens)))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(int(currentTokens)))
			c.Header("X-RateLimit-Reset", time.Now().Add(refillInterval).Format(time.RFC3339))
			c.Header("X-RateLimit-Algorithm", "token-bucket")
		}

		c.Next()
	}
}

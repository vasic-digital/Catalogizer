package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	// Requests per second per client
	Rate float64
	// Burst size allowed
	Burst int
	// Key generator function (e.g., by IP, user ID, etc.)
	KeyGenerator func(*gin.Context) string
	// Message to return when rate limited
	Message string
}

// DefaultRateLimiterConfig returns a secure default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Rate:    10, // 10 requests per second
		Burst:   20, // Allow burst of 20 requests
		Message: "Rate limit exceeded. Please try again later.",
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
}

// StrictRateLimiterConfig returns a stricter configuration for sensitive endpoints
func StrictRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Rate:    2, // 2 requests per second
		Burst:   5, // Burst of 5 requests
		Message: "Too many requests. Please wait before trying again.",
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
}

// AuthRateLimiterConfig returns a rate limiter specifically for auth endpoints
func AuthRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Rate:    3,  // 3 login attempts per second
		Burst:   10, // Burst of 10 attempts
		Message: "Too many authentication attempts. Account temporarily locked.",
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

// AdvancedRateLimiter manages rate limiting for different clients
type AdvancedRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   RateLimiterConfig
}

// NewAdvancedRateLimiter creates a new rate limiter
func NewAdvancedRateLimiter(config RateLimiterConfig) *AdvancedRateLimiter {
	return &AdvancedRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}
}

// getLimiter returns a rate limiter for the given key
func (r *AdvancedRateLimiter) getLimiter(key string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limiter, exists := r.limiters[key]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Limit(r.config.Rate), r.config.Burst)
	r.limiters[key] = limiter
	return limiter
}

// CleanupExpiredLimiters removes limiters that haven't been used recently
func (r *AdvancedRateLimiter) CleanupExpiredLimiters() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// This is a simplified cleanup - in production, you might want to track last access time
	if len(r.limiters) > 10000 { // Prevent memory leaks
		r.limiters = make(map[string]*rate.Limiter)
	}
}

// AdvancedRateLimit creates a Gin middleware for advanced rate limiting
func AdvancedRateLimit(config RateLimiterConfig) gin.HandlerFunc {
	limiter := NewAdvancedRateLimiter(config)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.CleanupExpiredLimiters()
		}
	}()

	return func(c *gin.Context) {
		key := config.KeyGenerator(c)
		clientLimiter := limiter.getLimiter(key)

		if !clientLimiter.Allow() {
			// Log rate limit attempt
			fmt.Printf("Rate limit exceeded for key: %s, IP: %s, Path: %s\n",
				key, c.ClientIP(), c.Request.URL.Path)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     config.Message,
				"retry_after": "5s", // Suggest retry after
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", config.Rate))
		c.Header("X-RateLimit-Remaining", "N/A") // Could track this more precisely
		c.Header("X-RateLimit-Reset", "N/A")

		c.Next()
	}
}

// UserBasedRateLimit creates rate limiting based on authenticated user
func UserBasedRateLimit(config RateLimiterConfig) gin.HandlerFunc {
	userConfig := config
	userConfig.KeyGenerator = func(c *gin.Context) string {
		// Try to get user ID from context (set by auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok {
				return "user:" + id
			}
		}
		// Fallback to IP for unauthenticated requests
		return "ip:" + c.ClientIP()
	}
	return AdvancedRateLimit(userConfig)
}

// IPRateLimit creates IP-based rate limiting
func IPRateLimit(requestsPerSecond float64, burst int) gin.HandlerFunc {
	config := DefaultRateLimiterConfig()
	config.Rate = requestsPerSecond
	config.Burst = burst
	config.KeyGenerator = func(c *gin.Context) string {
		return c.ClientIP()
	}
	return AdvancedRateLimit(config)
}

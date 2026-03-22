package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitStrategy defines different rate limiting strategies
type RateLimitStrategy string

const (
	// StrategyTokenBucket uses token bucket algorithm
	StrategyTokenBucket RateLimitStrategy = "token_bucket"
	// StrategyFixedWindow uses fixed window algorithm
	StrategyFixedWindow RateLimitStrategy = "fixed_window"
	// StrategySlidingWindow uses sliding window algorithm
	StrategySlidingWindow RateLimitStrategy = "sliding_window"
)

// RateLimitTier defines different rate limit tiers
type RateLimitTier string

const (
	// TierAnonymous for unauthenticated users
	TierAnonymous RateLimitTier = "anonymous"
	// TierAuthenticated for authenticated users
	TierAuthenticated RateLimitTier = "authenticated"
	// TierPremium for premium users
	TierPremium RateLimitTier = "premium"
	// TierAdmin for admin users
	TierAdmin RateLimitTier = "admin"
)

// EnhancedRateLimiterConfig holds configuration for enhanced rate limiting
type EnhancedRateLimiterConfig struct {
	// Strategy to use
	Strategy RateLimitStrategy
	// Tier configurations
	Tiers map[RateLimitTier]TierConfig
	// Default tier if not matched
	DefaultTier RateLimitTier
	// Key generator function
	KeyGenerator func(*gin.Context) string
	// Tier extractor function
	TierExtractor func(*gin.Context) RateLimitTier
	// Message to return when rate limited
	Message string
	// Enable metrics collection
	EnableMetrics bool
	// Cleanup interval for stale entries
	CleanupInterval time.Duration
}

// TierConfig holds rate limit configuration for a tier
type TierConfig struct {
	Rate  float64
	Burst int
}

// DefaultEnhancedRateLimiterConfig returns a secure default configuration
func DefaultEnhancedRateLimiterConfig() EnhancedRateLimiterConfig {
	return EnhancedRateLimiterConfig{
		Strategy:    StrategyTokenBucket,
		DefaultTier: TierAnonymous,
		Tiers: map[RateLimitTier]TierConfig{
			TierAnonymous:     {Rate: 10, Burst: 20},   // 10 req/sec
			TierAuthenticated: {Rate: 50, Burst: 100},  // 50 req/sec
			TierPremium:       {Rate: 100, Burst: 200}, // 100 req/sec
			TierAdmin:         {Rate: 200, Burst: 400}, // 200 req/sec
		},
		Message: "Rate limit exceeded. Please try again later.",
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
		TierExtractor: func(c *gin.Context) RateLimitTier {
			// Check for user role in context
			if role, exists := c.Get("user_role"); exists {
				if roleStr, ok := role.(string); ok {
					switch roleStr {
					case "admin":
						return TierAdmin
					case "premium":
						return TierPremium
					case "user":
						return TierAuthenticated
					}
				}
			}
			// Check for user ID (authenticated)
			if _, exists := c.Get("user_id"); exists {
				return TierAuthenticated
			}
			return TierAnonymous
		},
		EnableMetrics:   true,
		CleanupInterval: 5 * time.Minute,
	}
}

// RateLimitMetrics holds metrics for rate limiting
type RateLimitMetrics struct {
	TotalRequests   int64
	AllowedRequests int64
	BlockedRequests int64
	LastReset       time.Time
	mu              sync.RWMutex
}

// Reset resets the metrics
func (m *RateLimitMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests = 0
	m.AllowedRequests = 0
	m.BlockedRequests = 0
	m.LastReset = time.Now()
}

// GetStats returns current metrics
func (m *RateLimitMetrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_requests":   m.TotalRequests,
		"allowed_requests": m.AllowedRequests,
		"blocked_requests": m.BlockedRequests,
		"block_rate":       float64(m.BlockedRequests) / float64(m.TotalRequests) * 100,
		"last_reset":       m.LastReset,
	}
}

// EnhancedRateLimiter manages tiered rate limiting
type EnhancedRateLimiter struct {
	config      EnhancedRateLimiterConfig
	limiters    map[string]*rate.Limiter
	lastSeen    map[string]time.Time
	metrics     *RateLimitMetrics
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(config EnhancedRateLimiterConfig) *EnhancedRateLimiter {
	erl := &EnhancedRateLimiter{
		config:      config,
		limiters:    make(map[string]*rate.Limiter),
		lastSeen:    make(map[string]time.Time),
		metrics:     &RateLimitMetrics{LastReset: time.Now()},
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	go erl.cleanupLoop()

	return erl
}

// Stop gracefully stops the rate limiter
func (er *EnhancedRateLimiter) Stop() {
	close(er.stopCleanup)
}

// cleanupLoop periodically removes stale limiters
func (er *EnhancedRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(er.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			er.cleanupStaleLimiters()
		case <-er.stopCleanup:
			return
		}
	}
}

// cleanupStaleLimiters removes limiters not seen recently
func (er *EnhancedRateLimiter) cleanupStaleLimiters() {
	er.mu.Lock()
	defer er.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-10 * time.Minute)

	for key, lastSeen := range er.lastSeen {
		if lastSeen.Before(cutoff) {
			delete(er.limiters, key)
			delete(er.lastSeen, key)
		}
	}
}

// getLimiter returns a rate limiter for the given key and tier
func (er *EnhancedRateLimiter) getLimiter(key string, tier RateLimitTier) *rate.Limiter {
	er.mu.Lock()
	defer er.mu.Unlock()

	fullKey := string(tier) + ":" + key

	if limiter, exists := er.limiters[fullKey]; exists {
		er.lastSeen[fullKey] = time.Now()
		return limiter
	}

	// Get tier config
	tierConfig, exists := er.config.Tiers[tier]
	if !exists {
		tierConfig = er.config.Tiers[er.config.DefaultTier]
	}

	limiter := rate.NewLimiter(rate.Limit(tierConfig.Rate), tierConfig.Burst)
	er.limiters[fullKey] = limiter
	er.lastSeen[fullKey] = time.Now()

	return limiter
}

// Allow checks if a request is allowed
func (er *EnhancedRateLimiter) Allow(c *gin.Context) bool {
	if er.config.EnableMetrics {
		er.metrics.mu.Lock()
		er.metrics.TotalRequests++
		er.metrics.mu.Unlock()
	}

	key := er.config.KeyGenerator(c)
	tier := er.config.TierExtractor(c)

	limiter := er.getLimiter(key, tier)

	if !limiter.Allow() {
		if er.config.EnableMetrics {
			er.metrics.mu.Lock()
			er.metrics.BlockedRequests++
			er.metrics.mu.Unlock()
		}
		return false
	}

	if er.config.EnableMetrics {
		er.metrics.mu.Lock()
		er.metrics.AllowedRequests++
		er.metrics.mu.Unlock()
	}

	return true
}

// GetMetrics returns current metrics
func (er *EnhancedRateLimiter) GetMetrics() map[string]interface{} {
	if !er.config.EnableMetrics {
		return map[string]interface{}{"metrics_disabled": true}
	}
	return er.metrics.GetStats()
}

// ResetMetrics resets metrics counters
func (er *EnhancedRateLimiter) ResetMetrics() {
	if er.config.EnableMetrics {
		er.metrics.Reset()
	}
}

// Middleware returns the Gin middleware
func (er *EnhancedRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !er.Allow(c) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       er.config.Message,
				"retry_after": "1",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPBasedKeyGenerator generates rate limit keys based on IP address
func IPBasedKeyGenerator(c *gin.Context) string {
	return c.ClientIP()
}

// UserBasedKeyGenerator generates rate limit keys based on user ID
func UserBasedKeyGenerator(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return c.ClientIP()
}

// IPAndUserBasedKeyGenerator generates keys based on both IP and user
func IPAndUserBasedKeyGenerator(c *gin.Context) string {
	ip := c.ClientIP()
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return fmt.Sprintf("%s:%s", id, ip)
		}
	}
	return ip
}

// Global instance
var globalRateLimiter *EnhancedRateLimiter
var globalRateLimiterOnce sync.Once

// InitGlobalRateLimiter initializes the global rate limiter
func InitGlobalRateLimiter(config EnhancedRateLimiterConfig) {
	globalRateLimiterOnce.Do(func() {
		globalRateLimiter = NewEnhancedRateLimiter(config)
	})
}

// GetGlobalRateLimiter returns the global rate limiter instance
func GetGlobalRateLimiter() *EnhancedRateLimiter {
	if globalRateLimiter == nil {
		InitGlobalRateLimiter(DefaultEnhancedRateLimiterConfig())
	}
	return globalRateLimiter
}

// EnhancedRateLimitMiddleware returns the global rate limiter middleware
func EnhancedRateLimitMiddleware() gin.HandlerFunc {
	return GetGlobalRateLimiter().Middleware()
}

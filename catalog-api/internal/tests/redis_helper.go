package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTestHelper provides utilities for testing with Redis
type RedisTestHelper struct {
	client *redis.Client
	t      *testing.T
}

// SetupTestRedis creates a Redis client for testing
// Uses a local Redis instance if available, otherwise skips the test
func SetupTestRedis(t *testing.T) *RedisTestHelper {
	// Try to connect to local Redis
	client := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           15, // Use DB 15 for tests to avoid conflicts
		DialTimeout:  2 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Check if Redis is available
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Skip("Redis not available for testing - skipping test")
		return nil
	}

	helper := &RedisTestHelper{
		client: client,
		t:      t,
	}

	// Clean the test DB
	helper.FlushDB()

	return helper
}

// GetClient returns the Redis client
func (h *RedisTestHelper) GetClient() *redis.Client {
	return h.client
}

// FlushDB flushes the test database
func (h *RedisTestHelper) FlushDB() {
	ctx := context.Background()
	if err := h.client.FlushDB(ctx).Err(); err != nil {
		h.t.Fatalf("Failed to flush Redis test DB: %v", err)
	}
}

// Set sets a key-value pair
func (h *RedisTestHelper) Set(key string, value interface{}, expiration time.Duration) {
	ctx := context.Background()
	if err := h.client.Set(ctx, key, value, expiration).Err(); err != nil {
		h.t.Fatalf("Failed to set Redis key %s: %v", key, err)
	}
}

// Get gets a value by key
func (h *RedisTestHelper) Get(key string) string {
	ctx := context.Background()
	val, err := h.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ""
	}
	if err != nil {
		h.t.Fatalf("Failed to get Redis key %s: %v", key, err)
	}
	return val
}

// Exists checks if a key exists
func (h *RedisTestHelper) Exists(key string) bool {
	ctx := context.Background()
	count, err := h.client.Exists(ctx, key).Result()
	if err != nil {
		h.t.Fatalf("Failed to check Redis key existence %s: %v", key, err)
	}
	return count > 0
}

// Del deletes keys
func (h *RedisTestHelper) Del(keys ...string) {
	ctx := context.Background()
	if err := h.client.Del(ctx, keys...).Err(); err != nil {
		h.t.Fatalf("Failed to delete Redis keys: %v", err)
	}
}

// Keys returns all keys matching a pattern
func (h *RedisTestHelper) Keys(pattern string) []string {
	ctx := context.Background()
	keys, err := h.client.Keys(ctx, pattern).Result()
	if err != nil {
		h.t.Fatalf("Failed to get Redis keys with pattern %s: %v", pattern, err)
	}
	return keys
}

// Incr increments a key
func (h *RedisTestHelper) Incr(key string) int64 {
	ctx := context.Background()
	val, err := h.client.Incr(ctx, key).Result()
	if err != nil {
		h.t.Fatalf("Failed to increment Redis key %s: %v", key, err)
	}
	return val
}

// Expire sets expiration on a key
func (h *RedisTestHelper) Expire(key string, expiration time.Duration) {
	ctx := context.Background()
	if err := h.client.Expire(ctx, key, expiration).Err(); err != nil {
		h.t.Fatalf("Failed to set expiration on Redis key %s: %v", key, err)
	}
}

// TTL gets the time to live for a key
func (h *RedisTestHelper) TTL(key string) time.Duration {
	ctx := context.Background()
	ttl, err := h.client.TTL(ctx, key).Result()
	if err != nil {
		h.t.Fatalf("Failed to get TTL for Redis key %s: %v", key, err)
	}
	return ttl
}

// Close closes the Redis connection
func (h *RedisTestHelper) Close() {
	if h.client != nil {
		h.FlushDB() // Clean up before closing
		h.client.Close()
	}
}

// MockRedisClient creates a mock Redis client for tests that don't need real Redis
// This is useful for unit tests that just need to test logic without Redis
func MockRedisClient() *redis.Client {
	// Return a client that will fail all operations
	// Tests using this should mock the expected behavior
	return redis.NewClient(&redis.Options{
		Addr: "invalid:0",
	})
}

// WaitForKey waits for a key to exist with a timeout
func (h *RedisTestHelper) WaitForKey(key string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for key %s", key)
		case <-ticker.C:
			if h.Exists(key) {
				return nil
			}
		}
	}
}

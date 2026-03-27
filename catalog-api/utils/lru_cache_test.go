package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// LRUCache Constructor Tests
// =============================================================================

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache(100)
	require.NotNil(t, cache)
	assert.Equal(t, 100, cache.Capacity())
	assert.Equal(t, 0, cache.Len())
}

func TestNewLRUCache_ZeroCapacity(t *testing.T) {
	cache := NewLRUCache(0)
	require.NotNil(t, cache)
	assert.Equal(t, 1000, cache.Capacity(), "zero capacity should default to 1000")
}

func TestNewLRUCache_NegativeCapacity(t *testing.T) {
	cache := NewLRUCache(-5)
	require.NotNil(t, cache)
	assert.Equal(t, 1000, cache.Capacity(), "negative capacity should default to 1000")
}

func TestNewLRUCacheWithTTL(t *testing.T) {
	cache := NewLRUCacheWithTTL(50, 5*time.Minute)
	require.NotNil(t, cache)
	assert.Equal(t, 50, cache.Capacity())
}

// =============================================================================
// Get / Set Tests
// =============================================================================

func TestLRUCache_SetAndGet(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Set("key1", "value1")
	val, found := cache.Get("key1")

	assert.True(t, found)
	assert.Equal(t, "value1", val)
}

func TestLRUCache_Get_NotFound(t *testing.T) {
	cache := NewLRUCache(10)

	val, found := cache.Get("nonexistent")
	assert.False(t, found)
	assert.Nil(t, val)
}

func TestLRUCache_Set_UpdateExisting(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Set("key1", "value1")
	cache.Set("key1", "value2")

	val, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value2", val)
	assert.Equal(t, 1, cache.Len(), "should not duplicate entries on update")
}

func TestLRUCache_SetWithTTL(t *testing.T) {
	cache := NewLRUCache(10)

	cache.SetWithTTL("key1", "value1", 50*time.Millisecond)

	val, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", val)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	val, found = cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, val)
}

func TestLRUCache_SetWithTTL_ZeroTTL_NoExpiration(t *testing.T) {
	cache := NewLRUCache(10)

	cache.SetWithTTL("key1", "value1", 0)

	val, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", val)
}

// =============================================================================
// Eviction Tests
// =============================================================================

func TestLRUCache_Eviction_OverCapacity(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	cache.Set("d", 4) // should evict "a"

	assert.Equal(t, 3, cache.Len())

	_, found := cache.Get("a")
	assert.False(t, found, "oldest item should be evicted")

	val, found := cache.Get("d")
	assert.True(t, found)
	assert.Equal(t, 4, val)
}

func TestLRUCache_Eviction_AccessUpdatesOrder(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	// Access "a" so it becomes most recently used
	cache.Get("a")

	cache.Set("d", 4) // should evict "b" (now the oldest)

	_, found := cache.Get("a")
	assert.True(t, found, "recently accessed item should not be evicted")

	_, found = cache.Get("b")
	assert.False(t, found, "least recently used item should be evicted")
}

func TestLRUCache_Eviction_OnEvictCallback(t *testing.T) {
	cache := NewLRUCache(2)

	var evictedKey string
	var evictedValue interface{}

	cache.SetOnEvict(func(key string, value interface{}) {
		evictedKey = key
		evictedValue = value
	})

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3) // should evict "a"

	assert.Equal(t, "a", evictedKey)
	assert.Equal(t, 1, evictedValue)
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestLRUCache_Delete_Existing(t *testing.T) {
	cache := NewLRUCache(10)
	cache.Set("key1", "value1")

	deleted := cache.Delete("key1")
	assert.True(t, deleted)
	assert.Equal(t, 0, cache.Len())

	_, found := cache.Get("key1")
	assert.False(t, found)
}

func TestLRUCache_Delete_NonExistent(t *testing.T) {
	cache := NewLRUCache(10)

	deleted := cache.Delete("nonexistent")
	assert.False(t, deleted)
}

func TestLRUCache_Delete_TriggersOnEvict(t *testing.T) {
	cache := NewLRUCache(10)

	var evictedKey string
	cache.SetOnEvict(func(key string, value interface{}) {
		evictedKey = key
	})

	cache.Set("key1", "value1")
	cache.Delete("key1")

	assert.Equal(t, "key1", evictedKey)
}

// =============================================================================
// Contains Tests
// =============================================================================

func TestLRUCache_Contains_Exists(t *testing.T) {
	cache := NewLRUCache(10)
	cache.Set("key1", "value1")

	assert.True(t, cache.Contains("key1"))
}

func TestLRUCache_Contains_NotExists(t *testing.T) {
	cache := NewLRUCache(10)
	assert.False(t, cache.Contains("nonexistent"))
}

func TestLRUCache_Contains_Expired(t *testing.T) {
	cache := NewLRUCache(10)
	cache.SetWithTTL("key1", "value1", 50*time.Millisecond)

	assert.True(t, cache.Contains("key1"))

	time.Sleep(100 * time.Millisecond)

	assert.False(t, cache.Contains("key1"))
}

// =============================================================================
// Peek Tests
// =============================================================================

func TestLRUCache_Peek_DoesNotUpdateOrder(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	// Peek at "a" -- should NOT update order
	val, found := cache.Peek("a")
	assert.True(t, found)
	assert.Equal(t, 1, val)

	// Adding a new item should evict "a" since Peek didn't update order
	cache.Set("d", 4)

	_, found = cache.Get("a")
	assert.False(t, found, "peeked item should still be evicted as LRU")
}

func TestLRUCache_Peek_NotFound(t *testing.T) {
	cache := NewLRUCache(10)

	val, found := cache.Peek("nonexistent")
	assert.False(t, found)
	assert.Nil(t, val)
}

func TestLRUCache_Peek_Expired(t *testing.T) {
	cache := NewLRUCache(10)
	cache.SetWithTTL("key1", "value1", 50*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	val, found := cache.Peek("key1")
	assert.False(t, found)
	assert.Nil(t, val)
}

// =============================================================================
// Len / Capacity Tests
// =============================================================================

func TestLRUCache_Len(t *testing.T) {
	cache := NewLRUCache(10)
	assert.Equal(t, 0, cache.Len())

	cache.Set("a", 1)
	assert.Equal(t, 1, cache.Len())

	cache.Set("b", 2)
	assert.Equal(t, 2, cache.Len())

	cache.Delete("a")
	assert.Equal(t, 1, cache.Len())
}

func TestLRUCache_Capacity(t *testing.T) {
	cache := NewLRUCache(42)
	assert.Equal(t, 42, cache.Capacity())
}

// =============================================================================
// Clear Tests
// =============================================================================

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10)
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	cache.Clear()

	assert.Equal(t, 0, cache.Len())
	_, found := cache.Get("a")
	assert.False(t, found)
}

func TestLRUCache_Clear_TriggersOnEvict(t *testing.T) {
	cache := NewLRUCache(10)

	evictedKeys := make([]string, 0)
	cache.SetOnEvict(func(key string, value interface{}) {
		evictedKeys = append(evictedKeys, key)
	})

	cache.Set("a", 1)
	cache.Set("b", 2)

	cache.Clear()

	assert.Len(t, evictedKeys, 2)
}

func TestLRUCache_Clear_NoOnEvict(t *testing.T) {
	cache := NewLRUCache(10)
	cache.Set("a", 1)
	cache.Set("b", 2)

	// Should not panic without OnEvict callback
	cache.Clear()
	assert.Equal(t, 0, cache.Len())
}

// =============================================================================
// Keys Tests
// =============================================================================

func TestLRUCache_Keys(t *testing.T) {
	cache := NewLRUCache(10)
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	keys := cache.Keys()
	assert.Len(t, keys, 3)
	// Keys should be in MRU to LRU order
	assert.Equal(t, "c", keys[0])
	assert.Equal(t, "b", keys[1])
	assert.Equal(t, "a", keys[2])
}

func TestLRUCache_Keys_Empty(t *testing.T) {
	cache := NewLRUCache(10)
	keys := cache.Keys()
	assert.Len(t, keys, 0)
}

func TestLRUCache_Keys_SkipsExpired(t *testing.T) {
	cache := NewLRUCache(10)
	cache.SetWithTTL("expired", "val", 50*time.Millisecond)
	cache.Set("alive", "val")

	time.Sleep(100 * time.Millisecond)

	keys := cache.Keys()
	assert.Len(t, keys, 1)
	assert.Equal(t, "alive", keys[0])
}

// =============================================================================
// CleanupExpired Tests
// =============================================================================

func TestLRUCache_CleanupExpired(t *testing.T) {
	cache := NewLRUCache(10)
	cache.SetWithTTL("expired1", "val1", 50*time.Millisecond)
	cache.SetWithTTL("expired2", "val2", 50*time.Millisecond)
	cache.Set("alive", "val3")

	time.Sleep(100 * time.Millisecond)

	removed := cache.CleanupExpired()
	assert.Equal(t, 2, removed)
	assert.Equal(t, 1, cache.Len())
}

func TestLRUCache_CleanupExpired_NoneExpired(t *testing.T) {
	cache := NewLRUCache(10)
	cache.Set("a", 1)
	cache.Set("b", 2)

	removed := cache.CleanupExpired()
	assert.Equal(t, 0, removed)
	assert.Equal(t, 2, cache.Len())
}

// =============================================================================
// GetStats Tests
// =============================================================================

func TestLRUCache_GetStats(t *testing.T) {
	cache := NewLRUCacheWithTTL(100, 50*time.Millisecond)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.SetWithTTL("c", 3, 50*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	stats := cache.GetStats()
	assert.Equal(t, 100, stats.Capacity)
	assert.Equal(t, 3, stats.Size)
	assert.GreaterOrEqual(t, stats.Expired, 1) // at least "c" expired, maybe also "a" and "b"
	assert.Greater(t, stats.Load, 0.0)
}

func TestLRUCacheStats_String(t *testing.T) {
	stats := LRUCacheStats{
		Capacity: 100,
		Size:     50,
		Expired:  5,
		Load:     50.0,
	}

	str := stats.String()
	assert.Contains(t, str, "100")
	assert.Contains(t, str, "50")
	assert.Contains(t, str, "50.0%")
}

// =============================================================================
// SafeCache Tests
// =============================================================================

func TestNewSafeCache(t *testing.T) {
	sc := NewSafeCache(10)
	require.NotNil(t, sc)
}

func TestSafeCache_SetAndGet(t *testing.T) {
	sc := NewSafeCache(10)

	sc.Set("key1", "value1")
	val, found := sc.Get("key1")

	assert.True(t, found)
	assert.Equal(t, "value1", val)
}

func TestSafeCache_Get_NotFound(t *testing.T) {
	sc := NewSafeCache(10)

	val, found := sc.Get("nonexistent")
	assert.False(t, found)
	assert.Nil(t, val)
}

func TestSafeCache_Delete(t *testing.T) {
	sc := NewSafeCache(10)

	sc.Set("key1", "value1")
	deleted := sc.Delete("key1")

	assert.True(t, deleted)

	_, found := sc.Get("key1")
	assert.False(t, found)
}

func TestSafeCache_Len(t *testing.T) {
	sc := NewSafeCache(10)
	assert.Equal(t, 0, sc.Len())

	sc.Set("a", 1)
	assert.Equal(t, 1, sc.Len())
}

func TestSafeCache_Clear(t *testing.T) {
	sc := NewSafeCache(10)
	sc.Set("a", 1)
	sc.Set("b", 2)

	sc.Clear()
	assert.Equal(t, 0, sc.Len())
}

// =============================================================================
// Concurrency Safety Tests
// =============================================================================

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache := NewLRUCache(100)

	var wg sync.WaitGroup
	goroutines := 10
	opsPerGoroutine := 100

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("key-%d-%d", id, i)
				cache.Set(key, i)
				cache.Get(key)
				cache.Contains(key)
				cache.Peek(key)
				if i%10 == 0 {
					cache.Delete(key)
				}
			}
		}(g)
	}

	wg.Wait()
	// No panics or data races means success
}

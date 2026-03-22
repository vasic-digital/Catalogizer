package utils

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// LRUCache implements a thread-safe Least Recently Used cache with TTL support
type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
	mu       sync.RWMutex
	onEvict  func(key string, value interface{})
	ttl      time.Duration
}

// cacheEntry represents a single cache entry
type cacheEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// NewLRUCache creates a new LRU cache with the specified capacity
func NewLRUCache(capacity int) *LRUCache {
	return NewLRUCacheWithTTL(capacity, 0)
}

// NewLRUCacheWithTTL creates a new LRU cache with TTL support
func NewLRUCacheWithTTL(capacity int, ttl time.Duration) *LRUCache {
	if capacity <= 0 {
		capacity = 1000 // Default capacity
	}

	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
		ttl:      ttl,
	}
}

// SetOnEvict sets a callback function called when an entry is evicted
func (c *LRUCache) SetOnEvict(onEvict func(key string, value interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = onEvict
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry)

		// Check if expired
		if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
			c.removeElement(elem)
			return nil, false
		}

		// Move to front (most recently used)
		c.order.MoveToFront(elem)
		return entry.value, true
	}

	return nil, false
}

// Set adds or updates a value in the cache
func (c *LRUCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL adds or updates a value with a specific TTL
func (c *LRUCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate expiration
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	// Update existing entry
	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry)
		entry.value = value
		entry.expiresAt = expiresAt
		c.order.MoveToFront(elem)
		return
	}

	// Add new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}
	elem := c.order.PushFront(entry)
	c.items[key] = elem

	// Evict oldest if over capacity
	if c.order.Len() > c.capacity {
		c.evictOldest()
	}
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
		return true
	}

	return false
}

// Contains checks if a key exists in the cache (without updating LRU order)
func (c *LRUCache) Contains(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry)
		// Check if expired
		if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
			return false
		}
		return true
	}

	return false
}

// Peek retrieves a value without updating LRU order
func (c *LRUCache) Peek(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry)
		if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
			return nil, false
		}
		return entry.value, true
	}

	return nil, false
}

// Len returns the number of items in the cache
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// Capacity returns the cache capacity
func (c *LRUCache) Capacity() int {
	return c.capacity
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvict != nil {
		for elem := c.order.Back(); elem != nil; elem = elem.Prev() {
			entry := elem.Value.(*cacheEntry)
			c.onEvict(entry.key, entry.value)
		}
	}

	c.items = make(map[string]*list.Element, c.capacity)
	c.order.Init()
}

// Keys returns all keys in the cache (from most to least recently used)
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, c.order.Len())
	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*cacheEntry)
		if entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt) {
			keys = append(keys, entry.key)
		}
	}

	return keys
}

// removeElement removes an element from the cache
func (c *LRUCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*cacheEntry)
	delete(c.items, entry.key)
	c.order.Remove(elem)

	if c.onEvict != nil {
		c.onEvict(entry.key, entry.value)
	}
}

// evictOldest removes the oldest entry from the cache
func (c *LRUCache) evictOldest() {
	elem := c.order.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// CleanupExpired removes all expired entries
func (c *LRUCache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	for elem := c.order.Back(); elem != nil; {
		prev := elem.Prev()
		entry := elem.Value.(*cacheEntry)

		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			c.removeElement(elem)
			removed++
		}

		elem = prev
	}

	return removed
}

// GetStats returns cache statistics
func (c *LRUCache) GetStats() LRUCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	expired := 0

	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*cacheEntry)
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			expired++
		}
	}

	return LRUCacheStats{
		Capacity: c.capacity,
		Size:     c.order.Len(),
		Expired:  expired,
		Load:     float64(c.order.Len()) / float64(c.capacity) * 100,
	}
}

// LRUCacheStats contains cache statistics
type LRUCacheStats struct {
	Capacity int
	Size     int
	Expired  int
	Load     float64
}

// String returns a string representation of the stats
func (s LRUCacheStats) String() string {
	return fmt.Sprintf("LRUCache{capacity=%d, size=%d, expired=%d, load=%.1f%%}",
		s.Capacity, s.Size, s.Expired, s.Load)
}

// SafeCache provides a type-safe wrapper around LRUCache
type SafeCache struct {
	cache *LRUCache
}

// NewSafeCache creates a new type-safe cache
func NewSafeCache(capacity int) *SafeCache {
	return &SafeCache{
		cache: NewLRUCache(capacity),
	}
}

// Get retrieves a typed value from the cache
func (sc *SafeCache) Get(key string) (interface{}, bool) {
	return sc.cache.Get(key)
}

// Set stores a typed value in the cache
func (sc *SafeCache) Set(key string, value interface{}) {
	sc.cache.Set(key, value)
}

// Delete removes a value from the cache
func (sc *SafeCache) Delete(key string) bool {
	return sc.cache.Delete(key)
}

// Len returns the number of items in the cache
func (sc *SafeCache) Len() int {
	return sc.cache.Len()
}

// Clear removes all items from the cache
func (sc *SafeCache) Clear() {
	sc.cache.Clear()
}

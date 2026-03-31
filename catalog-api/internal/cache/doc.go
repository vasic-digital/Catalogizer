// Package cache provides an in-memory caching layer with TTL-based expiration and automatic
// cleanup for the Catalogizer API.
//
// The CacheService manages cached entries with configurable time-to-live values and spawns
// a background goroutine for periodic cleanup of expired entries. It uses sync.Once for safe
// shutdown and must be closed via Close() in tests and during application shutdown to prevent
// goroutine leaks.
package cache

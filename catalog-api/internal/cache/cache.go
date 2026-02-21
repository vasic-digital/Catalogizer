// Package cache exposes the digital.vasic.cache types for use within Catalogizer.
//
// This package re-exports the core Cache interface and related types from
// digital.vasic.cache/pkg/cache as type aliases so all Catalogizer code can
// reference a single import path. No adapters are needed — Go type aliases are
// transparent to the compiler.
//
// Design patterns applied:
//   - Facade: single import point for all cache types
//   - Adapter: re-exports vasic cache types for Catalogizer callers
package cache

import (
	vasicache "digital.vasic.cache/pkg/cache"
)

// Cache is the core interface that all cache backends must implement.
// Backed by digital.vasic.cache/pkg/cache.Cache.
type Cache = vasicache.Cache

// Config holds general cache configuration.
type Config = vasicache.Config

// Stats captures runtime cache statistics.
type Stats = vasicache.Stats

// EvictionPolicy determines how entries are evicted when the cache reaches capacity.
type EvictionPolicy = vasicache.EvictionPolicy

const (
	// LRU evicts the least recently used entry.
	LRU = vasicache.LRU
	// LFU evicts the least frequently used entry.
	LFU = vasicache.LFU
	// FIFO evicts the oldest entry (first-in, first-out).
	FIFO = vasicache.FIFO
)

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return vasicache.DefaultConfig()
}

// NewTypedCache wraps an existing Cache with typed JSON Get/Set methods.
func NewTypedCache[T any](c Cache) *vasicache.TypedCache[T] {
	return vasicache.NewTypedCache[T](c)
}

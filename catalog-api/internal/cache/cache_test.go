package cache

import (
	"context"
	"testing"
	"time"

	vasicache "digital.vasic.cache/pkg/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCache is a minimal in-memory Cache implementation used for testing
// the facade wiring without pulling in the full memory backend.
type mockCache struct {
	store  map[string][]byte
	closed bool
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string][]byte)}
}

func (m *mockCache) Get(_ context.Context, key string) ([]byte, error) {
	v, ok := m.store[key]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	m.store[key] = value
	return nil
}

func (m *mockCache) Delete(_ context.Context, key string) error {
	delete(m.store, key)
	return nil
}

func (m *mockCache) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.store[key]
	return ok, nil
}

func (m *mockCache) Close() error {
	m.closed = true
	return nil
}

// ---------------------------------------------------------------------------
// DefaultConfig
// ---------------------------------------------------------------------------

func TestDefaultConfig(t *testing.T) {
	t.Run("returns non-nil config", func(t *testing.T) {
		cfg := DefaultConfig()
		require.NotNil(t, cfg)
	})

	t.Run("has sensible defaults", func(t *testing.T) {
		cfg := DefaultConfig()

		assert.Greater(t, cfg.DefaultTTL, time.Duration(0), "DefaultTTL should be positive")
		assert.Greater(t, cfg.MaxSize, 0, "MaxSize should be positive")
		assert.Equal(t, LRU, cfg.EvictionPolicy, "default eviction policy should be LRU")
	})

	t.Run("matches upstream DefaultConfig", func(t *testing.T) {
		facade := DefaultConfig()
		upstream := vasicache.DefaultConfig()

		assert.Equal(t, upstream.DefaultTTL, facade.DefaultTTL)
		assert.Equal(t, upstream.MaxSize, facade.MaxSize)
		assert.Equal(t, upstream.EvictionPolicy, facade.EvictionPolicy)
	})
}

// ---------------------------------------------------------------------------
// EvictionPolicy constants
// ---------------------------------------------------------------------------

func TestEvictionPolicyConstants(t *testing.T) {
	t.Run("constants are distinct", func(t *testing.T) {
		policies := []EvictionPolicy{LRU, LFU, FIFO}
		seen := make(map[EvictionPolicy]bool, len(policies))
		for _, p := range policies {
			assert.False(t, seen[p], "duplicate eviction policy value: %v", p)
			seen[p] = true
		}
	})

	t.Run("constants match upstream values", func(t *testing.T) {
		tests := []struct {
			name     string
			facade   EvictionPolicy
			upstream vasicache.EvictionPolicy
		}{
			{name: "LRU", facade: LRU, upstream: vasicache.LRU},
			{name: "LFU", facade: LFU, upstream: vasicache.LFU},
			{name: "FIFO", facade: FIFO, upstream: vasicache.FIFO},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.upstream, tt.facade)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// Type alias wiring
// ---------------------------------------------------------------------------

func TestTypeAliasWiring(t *testing.T) {
	t.Run("Cache alias accepts upstream implementation", func(t *testing.T) {
		mock := newMockCache()
		// The mock satisfies vasicache.Cache; it must also satisfy our alias.
		var c Cache = mock
		require.NotNil(t, c)
	})

	t.Run("Config alias is interchangeable with upstream", func(t *testing.T) {
		var cfg Config
		cfg.DefaultTTL = 5 * time.Minute
		cfg.MaxSize = 500
		cfg.EvictionPolicy = LFU

		// Assign to upstream type variable — compiles only if the alias is correct.
		var upstream vasicache.Config = cfg
		assert.Equal(t, cfg.DefaultTTL, upstream.DefaultTTL)
		assert.Equal(t, cfg.MaxSize, upstream.MaxSize)
		assert.Equal(t, cfg.EvictionPolicy, upstream.EvictionPolicy)
	})

	t.Run("Stats alias is interchangeable with upstream", func(t *testing.T) {
		var s Stats
		s.Hits = 100
		s.Misses = 25

		var upstream vasicache.Stats = s
		assert.Equal(t, s.Hits, upstream.Hits)
		assert.Equal(t, s.Misses, upstream.Misses)
	})
}

// ---------------------------------------------------------------------------
// NewTypedCache
// ---------------------------------------------------------------------------

type testItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestNewTypedCache(t *testing.T) {
	t.Run("returns non-nil wrapper", func(t *testing.T) {
		mock := newMockCache()
		tc := NewTypedCache[testItem](mock)
		require.NotNil(t, tc)
	})

	t.Run("set and get round-trip", func(t *testing.T) {
		ctx := context.Background()
		mock := newMockCache()
		tc := NewTypedCache[testItem](mock)

		item := testItem{Name: "widget", Value: 42}
		err := tc.Set(ctx, "item:1", item, time.Minute)
		require.NoError(t, err)

		got, found, err := tc.Get(ctx, "item:1")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, item, got)
	})

	t.Run("get returns zero value on miss", func(t *testing.T) {
		ctx := context.Background()
		mock := newMockCache()
		tc := NewTypedCache[testItem](mock)

		got, found, err := tc.Get(ctx, "nonexistent")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Equal(t, testItem{}, got)
	})

	t.Run("delete removes entry", func(t *testing.T) {
		ctx := context.Background()
		mock := newMockCache()
		tc := NewTypedCache[testItem](mock)

		require.NoError(t, tc.Set(ctx, "k", testItem{Name: "x", Value: 1}, 0))

		require.NoError(t, tc.Delete(ctx, "k"))

		_, found, err := tc.Get(ctx, "k")
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("exists reports presence", func(t *testing.T) {
		ctx := context.Background()
		mock := newMockCache()
		tc := NewTypedCache[testItem](mock)

		exists, err := tc.Exists(ctx, "k")
		require.NoError(t, err)
		assert.False(t, exists)

		require.NoError(t, tc.Set(ctx, "k", testItem{Name: "y"}, 0))

		exists, err = tc.Exists(ctx, "k")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("close delegates to underlying cache", func(t *testing.T) {
		mock := newMockCache()
		tc := NewTypedCache[testItem](mock)

		require.NoError(t, tc.Close())
		assert.True(t, mock.closed)
	})

	t.Run("works with primitive types", func(t *testing.T) {
		ctx := context.Background()
		mock := newMockCache()

		tests := []struct {
			name string
			fn   func(t *testing.T)
		}{
			{
				name: "string",
				fn: func(t *testing.T) {
					tc := NewTypedCache[string](mock)
					require.NoError(t, tc.Set(ctx, "s", "hello", 0))
					got, found, err := tc.Get(ctx, "s")
					require.NoError(t, err)
					assert.True(t, found)
					assert.Equal(t, "hello", got)
				},
			},
			{
				name: "int",
				fn: func(t *testing.T) {
					tc := NewTypedCache[int](mock)
					require.NoError(t, tc.Set(ctx, "i", 99, 0))
					got, found, err := tc.Get(ctx, "i")
					require.NoError(t, err)
					assert.True(t, found)
					assert.Equal(t, 99, got)
				},
			},
			{
				name: "slice",
				fn: func(t *testing.T) {
					tc := NewTypedCache[[]string](mock)
					val := []string{"a", "b", "c"}
					require.NoError(t, tc.Set(ctx, "sl", val, 0))
					got, found, err := tc.Get(ctx, "sl")
					require.NoError(t, err)
					assert.True(t, found)
					assert.Equal(t, val, got)
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, tt.fn)
		}
	})
}

package providers

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"catalogizer/internal/media/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubProvider is a minimal MetadataProvider for testing.
type stubProvider struct {
	name    string
	enabled bool
}

func (s *stubProvider) GetName() string { return s.name }
func (s *stubProvider) Search(_ context.Context, _ string, _ string, _ *int) ([]SearchResult, error) {
	return nil, nil
}
func (s *stubProvider) GetDetails(_ context.Context, _ string) (*models.ExternalMetadata, error) {
	return nil, nil
}
func (s *stubProvider) IsEnabled() bool { return s.enabled }

// ---------------------------------------------------------------------------
// NewLazyProvider
// ---------------------------------------------------------------------------

func TestNewLazyProvider_ReturnsNonNil(t *testing.T) {
	lp := NewLazyProvider(func() MetadataProvider {
		return &stubProvider{name: "test", enabled: true}
	})
	require.NotNil(t, lp)
}

// ---------------------------------------------------------------------------
// Get — basic initialization
// ---------------------------------------------------------------------------

func TestLazyProvider_Get_InitializesOnFirstCall(t *testing.T) {
	var callCount int32
	lp := NewLazyProvider(func() MetadataProvider {
		atomic.AddInt32(&callCount, 1)
		return &stubProvider{name: "lazy-test", enabled: true}
	})

	provider := lp.Get()
	require.NotNil(t, provider)
	assert.Equal(t, "lazy-test", provider.GetName())
	assert.True(t, provider.IsEnabled())
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestLazyProvider_Get_ReturnsCachedInstance(t *testing.T) {
	var callCount int32
	lp := NewLazyProvider(func() MetadataProvider {
		atomic.AddInt32(&callCount, 1)
		return &stubProvider{name: "cached", enabled: true}
	})

	p1 := lp.Get()
	p2 := lp.Get()
	p3 := lp.Get()

	// Factory should be called exactly once
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount),
		"factory should be called exactly once")

	// All calls return the same instance
	assert.Same(t, p1, p2)
	assert.Same(t, p2, p3)
}

// ---------------------------------------------------------------------------
// Get — concurrent access
// ---------------------------------------------------------------------------

func TestLazyProvider_Get_ConcurrentAccess(t *testing.T) {
	var callCount int32
	lp := NewLazyProvider(func() MetadataProvider {
		atomic.AddInt32(&callCount, 1)
		return &stubProvider{name: "concurrent", enabled: true}
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	providers := make([]MetadataProvider, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			providers[idx] = lp.Get()
		}(i)
	}

	wg.Wait()

	// Factory called exactly once
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount),
		"factory must be called exactly once under concurrent access")

	// All goroutines got the same instance
	for i := 1; i < goroutines; i++ {
		assert.Same(t, providers[0], providers[i],
			"all concurrent Get() calls must return the same instance")
	}
}

// ---------------------------------------------------------------------------
// Get — nil provider
// ---------------------------------------------------------------------------

func TestLazyProvider_Get_NilFactory(t *testing.T) {
	lp := NewLazyProvider(func() MetadataProvider {
		return nil
	})

	provider := lp.Get()
	assert.Nil(t, provider,
		"Get() should return nil if factory returns nil")

	// Second call should also return nil (cached)
	provider2 := lp.Get()
	assert.Nil(t, provider2)
}

// ---------------------------------------------------------------------------
// Get — disabled provider
// ---------------------------------------------------------------------------

func TestLazyProvider_Get_DisabledProvider(t *testing.T) {
	lp := NewLazyProvider(func() MetadataProvider {
		return &stubProvider{name: "disabled", enabled: false}
	})

	provider := lp.Get()
	require.NotNil(t, provider)
	assert.False(t, provider.IsEnabled())
	assert.Equal(t, "disabled", provider.GetName())
}

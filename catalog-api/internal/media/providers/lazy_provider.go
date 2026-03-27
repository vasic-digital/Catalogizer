package providers

import (
	"sync"
)

// LazyProvider wraps a provider factory function for deferred
// initialization.  The underlying MetadataProvider is created on the
// first call to Get() and reused for all subsequent calls.  This
// avoids paying the cost of provider construction (env-var lookups,
// HTTP client setup) for providers that may never be queried during a
// given session.
type LazyProvider struct {
	factory  func() MetadataProvider
	once     sync.Once
	provider MetadataProvider
}

// NewLazyProvider creates a provider that initializes on first use.
func NewLazyProvider(factory func() MetadataProvider) *LazyProvider {
	return &LazyProvider{factory: factory}
}

// Get returns the underlying MetadataProvider, initializing it on first
// call.  Subsequent calls return the cached instance.
func (lp *LazyProvider) Get() MetadataProvider {
	lp.once.Do(func() {
		lp.provider = lp.factory()
	})
	return lp.provider
}

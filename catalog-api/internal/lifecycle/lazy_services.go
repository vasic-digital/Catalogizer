// Package lifecycle provides on-demand service initialization using lazy loading.
//
// LazyServiceRegistry wraps sync.Once to guarantee each registered service
// initializer runs exactly once, even under concurrent access. This enables
// deferred startup costs for services that may not be needed on every run.
package lifecycle

import (
	"fmt"
	"sync"
)

// InitFunc creates a service on first access.
type InitFunc func() (interface{}, error)

// LazyServiceRegistry manages lazily-initialized services.
// Services are registered with a name and an initialization function.
// The init function is called at most once, on first Get() for that name.
type LazyServiceRegistry struct {
	mu       sync.RWMutex
	inits    map[string]InitFunc
	values   map[string]interface{}
	initOnce map[string]*sync.Once
}

// NewLazyServiceRegistry creates a new empty registry.
func NewLazyServiceRegistry() *LazyServiceRegistry {
	return &LazyServiceRegistry{
		inits:    make(map[string]InitFunc),
		values:   make(map[string]interface{}),
		initOnce: make(map[string]*sync.Once),
	}
}

// Register adds a named service with its initialization function.
// If a service with the same name was already registered, it is replaced
// and any cached value is discarded.
func (r *LazyServiceRegistry) Register(name string, init InitFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inits[name] = init
	r.initOnce[name] = &sync.Once{}
	delete(r.values, name) // clear stale cached value on re-register
}

// Get returns the service for the given name, initializing it on first call.
// Returns an error if the name is not registered or if initialization fails.
// After a failed initialization the error is permanent for that registration;
// call Register again to retry with a new init function.
func (r *LazyServiceRegistry) Get(name string) (interface{}, error) {
	r.mu.RLock()
	initFn, ok := r.inits[name]
	once, _ := r.initOnce[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("service %q not registered", name)
	}

	var initErr error
	once.Do(func() {
		val, err := initFn()
		if err != nil {
			initErr = err
			return
		}
		r.mu.Lock()
		r.values[name] = val
		r.mu.Unlock()
	})

	if initErr != nil {
		return nil, fmt.Errorf("service %q initialization failed: %w", name, initErr)
	}

	r.mu.RLock()
	val := r.values[name]
	r.mu.RUnlock()
	return val, nil
}

// Names returns the list of registered service names.
func (r *LazyServiceRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.inits))
	for name := range r.inits {
		names = append(names, name)
	}
	return names
}

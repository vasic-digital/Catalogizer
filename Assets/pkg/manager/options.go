package manager

import (
	"io"

	"digital.vasic.assets/pkg/defaults"
	"digital.vasic.assets/pkg/event"
	"digital.vasic.assets/pkg/resolver"
	"digital.vasic.assets/pkg/store"
)

// Option configures the Manager.
type Option func(*Manager)

// WithStore sets the content store.
func WithStore(s store.Store) Option {
	return func(m *Manager) { m.store = s }
}

// WithResolver sets the asset resolver.
func WithResolver(r resolver.Resolver) Option {
	return func(m *Manager) { m.resolver = r }
}

// WithEventBus sets the event bus for lifecycle notifications.
func WithEventBus(bus event.EventBus) Option {
	return func(m *Manager) { m.eventBus = bus }
}

// WithDefaults sets the default content provider.
func WithDefaults(p defaults.Provider) Option {
	return func(m *Manager) { m.defaults = p }
}

// WithWorkers sets the number of background resolution workers.
func WithWorkers(n int) Option {
	return func(m *Manager) { m.workerCount = n }
}

// WithLogger sets the logger for the manager.
func WithLogger(w io.Writer) Option {
	return func(m *Manager) { m.logger = w }
}

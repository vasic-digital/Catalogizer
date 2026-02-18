package event

import (
	"sync"
)

// InMemoryBus is an in-memory EventBus implementation.
type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[int]EventHandler
	nextID   int
}

// NewInMemoryBus creates a new in-memory event bus.
func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[int]EventHandler),
	}
}

// Publish sends an event to all subscribed handlers.
func (b *InMemoryBus) Publish(evt Event) {
	b.mu.RLock()
	handlers := make([]EventHandler, 0, len(b.handlers))
	for _, h := range b.handlers {
		handlers = append(handlers, h)
	}
	b.mu.RUnlock()

	for _, h := range handlers {
		h(evt)
	}
}

// Subscribe registers an event handler and returns an unsubscribe function.
func (b *InMemoryBus) Subscribe(handler EventHandler) func() {
	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.handlers[id] = handler
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		delete(b.handlers, id)
		b.mu.Unlock()
	}
}

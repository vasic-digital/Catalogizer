package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"digital.vasic.assets/pkg/resolver"
)

// CircuitBreakerResolver wraps any inner resolver.Resolver with a
// minimal in-process circuit breaker. P2 fix
// (docs/nexus/remaining-work.md): a slow upstream provider
// (Fanart.tv, IGDB, CAA, LLM, etc.) used to stall the whole cover
// chain because every request re-tried the broken endpoint in
// lockstep. The breaker opens after a configurable consecutive-error
// count and rejects new requests for an open window, letting the
// chain advance to the next provider without waiting for a new
// timeout on the sick one.
//
// The implementation is deliberately self-contained — it does not
// import the shared digital.vasic.recovery breaker to keep
// catalog-api's build graph small. Operators that want the full
// Recovery breaker (half-open probes, failure-rate windows, etc.)
// can swap this wrapper for their own decorator; both satisfy the
// resolver.Resolver interface.
type CircuitBreakerResolver struct {
	inner resolver.Resolver

	mu           sync.Mutex
	consecFails  int
	openedAt     time.Time
	failureLimit int
	openFor      time.Duration
}

// CircuitBreakerOption tunes the breaker at construction time.
type CircuitBreakerOption func(*CircuitBreakerResolver)

// WithCircuitBreakerFailureLimit configures how many consecutive
// Resolve failures trip the breaker. Zero / negative values fall
// back to the default (5).
func WithCircuitBreakerFailureLimit(n int) CircuitBreakerOption {
	return func(c *CircuitBreakerResolver) {
		if n > 0 {
			c.failureLimit = n
		}
	}
}

// WithCircuitBreakerOpenFor configures how long the breaker stays
// open after tripping. Zero / negative values fall back to 30s.
func WithCircuitBreakerOpenFor(d time.Duration) CircuitBreakerOption {
	return func(c *CircuitBreakerResolver) {
		if d > 0 {
			c.openFor = d
		}
	}
}

// NewCircuitBreakerResolver wraps inner with a circuit breaker.
func NewCircuitBreakerResolver(inner resolver.Resolver, opts ...CircuitBreakerOption) *CircuitBreakerResolver {
	c := &CircuitBreakerResolver{
		inner:        inner,
		failureLimit: 5,
		openFor:      30 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Name composes with the inner resolver.
func (c *CircuitBreakerResolver) Name() string {
	if c.inner == nil {
		return "circuit_breaker"
	}
	return "circuit_breaker(" + c.inner.Name() + ")"
}

// Priority defers to the inner.
func (c *CircuitBreakerResolver) Priority() int {
	if c.inner == nil {
		return 0
	}
	return c.inner.Priority()
}

// CanResolve returns false while the breaker is open so the chain
// advances past this provider immediately.
func (c *CircuitBreakerResolver) CanResolve(ctx context.Context, req *resolver.ResolveRequest) bool {
	if c.inner == nil {
		return false
	}
	if c.isOpen() {
		return false
	}
	return c.inner.CanResolve(ctx, req)
}

// ErrCircuitBreakerOpen is returned from Resolve while the breaker
// is in its open window.
var ErrCircuitBreakerOpen = errors.New("circuit breaker open")

// Resolve forwards to the inner resolver while the breaker is
// closed; tracks consecutive failures + trips the breaker after
// the configured threshold.
func (c *CircuitBreakerResolver) Resolve(ctx context.Context, req *resolver.ResolveRequest) (*resolver.ResolveResult, error) {
	if c.inner == nil {
		return nil, errors.New("circuit breaker: nil inner resolver")
	}
	if c.isOpen() {
		return nil, fmt.Errorf("%w: provider %s cooling down", ErrCircuitBreakerOpen, c.inner.Name())
	}
	result, err := c.inner.Resolve(ctx, req)
	c.recordOutcome(err)
	return result, err
}

func (c *CircuitBreakerResolver) isOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.openedAt.IsZero() {
		return false
	}
	if time.Since(c.openedAt) < c.openFor {
		return true
	}
	// Cooldown elapsed — move to half-open by clearing state. The
	// next Resolve() call gets a fresh attempt; a single failure
	// re-trips the breaker immediately (zero failures needed).
	c.openedAt = time.Time{}
	c.consecFails = 0
	return false
}

func (c *CircuitBreakerResolver) recordOutcome(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		c.consecFails = 0
		c.openedAt = time.Time{}
		return
	}
	c.consecFails++
	if c.consecFails >= c.failureLimit {
		c.openedAt = time.Now()
	}
}

// Verify interface compliance at compile time.
var _ resolver.Resolver = (*CircuitBreakerResolver)(nil)

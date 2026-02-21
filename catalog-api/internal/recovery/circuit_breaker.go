// Package recovery provides fault-tolerance primitives for Catalogizer services.
//
// CircuitBreaker and CircuitBreakerManager wrap digital.vasic.concurrency/pkg/breaker,
// adding Catalogizer-specific features: zap logger integration, named circuit breakers,
// state change callbacks, and a centralized manager.
package recovery

import (
	"sync"
	"time"

	vasicbreaker "digital.vasic.concurrency/pkg/breaker"
	"go.uber.org/zap"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal operation — requests pass through
	StateHalfOpen                     // Probing — limited requests pass through
	StateOpen                         // Failing — requests are rejected immediately
)

// String returns a human-readable state name.
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// mapBreakState translates vasicbreaker.State to CircuitState.
func mapBreakState(s vasicbreaker.State) CircuitState {
	switch s {
	case vasicbreaker.Closed:
		return StateClosed
	case vasicbreaker.HalfOpen:
		return StateHalfOpen
	case vasicbreaker.Open:
		return StateOpen
	default:
		return StateClosed
	}
}

// CircuitBreakerConfig contains configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	Name         string
	MaxFailures  int
	ResetTimeout time.Duration
	Logger       *zap.Logger
}

// CircuitBreaker wraps digital.vasic.concurrency/pkg/breaker.CircuitBreaker
// with Catalogizer-specific features: logger integration, named identification,
// and state change callbacks.
//
// Design patterns applied:
//   - Decorator: adds logging/callbacks to the base vasic breaker
//   - Facade: simplifies the breaker API surface for Catalogizer callers
type CircuitBreaker struct {
	name          string
	inner         *vasicbreaker.CircuitBreaker
	logger        *zap.Logger
	onStateChange func(string, CircuitState, CircuitState)
	maxFailures   int           // stored for GetStats
	resetTimeout  time.Duration // stored for GetStats
}

// NewCircuitBreaker creates a new circuit breaker backed by
// digital.vasic.concurrency/pkg/breaker.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	maxFailures := config.MaxFailures
	if maxFailures <= 0 {
		maxFailures = 5
	}
	resetTimeout := config.ResetTimeout
	if resetTimeout <= 0 {
		resetTimeout = 60 * time.Second
	}

	cfg := &vasicbreaker.Config{
		MaxFailures:      maxFailures,
		Timeout:          resetTimeout,
		HalfOpenRequests: 1,
	}

	return &CircuitBreaker{
		name:         config.Name,
		inner:        vasicbreaker.New(cfg),
		logger:       config.Logger,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// SetStateChangeCallback registers a callback invoked on every state transition.
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(string, CircuitState, CircuitState)) {
	cb.onStateChange = callback
}

// Execute wraps fn with circuit breaker protection, delegating to the
// digital.vasic.concurrency breaker engine and adding logging and callbacks.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	prevState := mapBreakState(cb.inner.State())

	err := cb.inner.Execute(fn)

	newState := mapBreakState(cb.inner.State())

	if cb.logger != nil {
		if err != nil {
			cb.logger.Warn("Circuit breaker recorded failure",
				zap.String("name", cb.name),
				zap.Int("failures", cb.inner.Failures()),
				zap.String("state", newState.String()))
		} else {
			cb.logger.Debug("Circuit breaker recorded success",
				zap.String("name", cb.name),
				zap.String("state", newState.String()))
		}
	}

	if prevState != newState {
		if cb.logger != nil {
			cb.logger.Info("Circuit breaker state changed",
				zap.String("name", cb.name),
				zap.String("old_state", prevState.String()),
				zap.String("new_state", newState.String()))
		}
		if cb.onStateChange != nil {
			cb.onStateChange(cb.name, prevState, newState)
		}
	}

	return err
}

// GetState returns the current circuit breaker state.
func (cb *CircuitBreaker) GetState() CircuitState {
	return mapBreakState(cb.inner.State())
}

// GetFailures returns the current consecutive failure count.
func (cb *CircuitBreaker) GetFailures() int {
	return cb.inner.Failures()
}

// GetStats returns circuit breaker statistics as a map.
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"name":          cb.name,
		"state":         mapBreakState(cb.inner.State()).String(),
		"failures":      cb.inner.Failures(),
		"max_failures":  cb.maxFailures,
		"reset_timeout": cb.resetTimeout,
	}
}

// Reset forces the circuit breaker back to the closed state.
func (cb *CircuitBreaker) Reset() {
	cb.inner.Reset()
	if cb.logger != nil {
		cb.logger.Info("Circuit breaker manually reset", zap.String("name", cb.name))
	}
}

// CircuitBreakerManager manages a named registry of circuit breakers.
//
// Design pattern: Registry — centralized lookup and creation of named breakers.
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
	logger   *zap.Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager.
func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate retrieves an existing circuit breaker by name, or creates a new one.
func (m *CircuitBreakerManager) GetOrCreate(name string, config CircuitBreakerConfig) *CircuitBreaker {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if cb, exists := m.breakers[name]; exists {
		return cb
	}

	config.Name = name
	if config.Logger == nil {
		config.Logger = m.logger
	}

	cb := NewCircuitBreaker(config)
	m.breakers[name] = cb

	if m.logger != nil {
		m.logger.Info("Created new circuit breaker", zap.String("name", name))
	}
	return cb
}

// Get retrieves a circuit breaker by name. Returns nil if not found.
func (m *CircuitBreakerManager) Get(name string) *CircuitBreaker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.breakers[name]
}

// GetAll returns a snapshot of all managed circuit breakers.
func (m *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*CircuitBreaker, len(m.breakers))
	for name, cb := range m.breakers {
		result[name] = cb
	}
	return result
}

// GetStats returns aggregated statistics for all managed circuit breakers.
func (m *CircuitBreakerManager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{}, len(m.breakers))
	for name, cb := range m.breakers {
		stats[name] = cb.GetStats()
	}
	return stats
}

// Reset resets all managed circuit breakers to the closed state.
func (m *CircuitBreakerManager) Reset() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, cb := range m.breakers {
		cb.Reset()
	}

	if m.logger != nil {
		m.logger.Info("All circuit breakers reset")
	}
}

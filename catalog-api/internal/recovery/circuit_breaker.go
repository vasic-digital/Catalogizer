package recovery

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

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

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	name            string
	maxFailures     int
	resetTimeout    time.Duration
	state           CircuitState
	failures        int
	lastFailureTime time.Time
	nextAttempt     time.Time
	mutex           sync.RWMutex
	logger          *zap.Logger
	onStateChange   func(string, CircuitState, CircuitState)
}

// CircuitBreakerConfig contains configuration for a circuit breaker
type CircuitBreakerConfig struct {
	Name         string
	MaxFailures  int
	ResetTimeout time.Duration
	Logger       *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 5
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 60 * time.Second
	}

	return &CircuitBreaker{
		name:         config.Name,
		maxFailures:  config.MaxFailures,
		resetTimeout: config.ResetTimeout,
		state:        StateClosed,
		logger:       config.Logger,
	}
}

// SetStateChangeCallback sets a callback for state changes
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(string, CircuitState, CircuitState)) {
	cb.onStateChange = callback
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return errors.New("circuit breaker is open")
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// allowRequest determines if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return time.Now().After(cb.nextAttempt)
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.setState(StateOpen)
			cb.nextAttempt = time.Now().Add(cb.resetTimeout)
		}
	case StateHalfOpen:
		cb.setState(StateOpen)
		cb.nextAttempt = time.Now().Add(cb.resetTimeout)
	}

	cb.logger.Warn("Circuit breaker recorded failure",
		zap.String("name", cb.name),
		zap.Int("failures", cb.failures),
		zap.String("state", cb.state.String()))
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateHalfOpen:
		cb.failures = 0
		cb.setState(StateClosed)
	case StateClosed:
		cb.failures = 0
	}

	cb.logger.Debug("Circuit breaker recorded success",
		zap.String("name", cb.name),
		zap.String("state", cb.state.String()))
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState

	cb.logger.Info("Circuit breaker state changed",
		zap.String("name", cb.name),
		zap.String("old_state", oldState.String()),
		zap.String("new_state", newState.String()))

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, newState)
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failures
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"name":              cb.name,
		"state":             cb.state.String(),
		"failures":          cb.failures,
		"max_failures":      cb.maxFailures,
		"last_failure_time": cb.lastFailureTime,
		"next_attempt":      cb.nextAttempt,
		"reset_timeout":     cb.resetTimeout,
	}
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	cb.setState(StateClosed)

	cb.logger.Info("Circuit breaker manually reset", zap.String("name", cb.name))
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
	logger   *zap.Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (m *CircuitBreakerManager) GetOrCreate(name string, config CircuitBreakerConfig) *CircuitBreaker {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if breaker, exists := m.breakers[name]; exists {
		return breaker
	}

	config.Name = name
	if config.Logger == nil {
		config.Logger = m.logger
	}

	breaker := NewCircuitBreaker(config)
	m.breakers[name] = breaker

	m.logger.Info("Created new circuit breaker", zap.String("name", name))
	return breaker
}

// Get retrieves a circuit breaker by name
func (m *CircuitBreakerManager) Get(name string) *CircuitBreaker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.breakers[name]
}

// GetAll returns all circuit breakers
func (m *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*CircuitBreaker)
	for name, breaker := range m.breakers {
		result[name] = breaker
	}
	return result
}

// GetStats returns statistics for all circuit breakers
func (m *CircuitBreakerManager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, breaker := range m.breakers {
		stats[name] = breaker.GetStats()
	}
	return stats
}

// Reset resets all circuit breakers
func (m *CircuitBreakerManager) Reset() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, breaker := range m.breakers {
		breaker.Reset()
	}

	m.logger.Info("All circuit breakers reset")
}

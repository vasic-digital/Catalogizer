package middleware

import "sync"

// Package-level shutdown registry.
//
// Several middleware factories (e.g. RateLimiter) spawn
// background cleanup goroutines but return only a gin.HandlerFunc, so the
// caller has no handle to stop them. The registry gives main.go a single
// entry point to stop all of them at once during graceful shutdown.
var (
	stopMu       sync.Mutex
	stopRegistry []func()
)

// registerStop adds a stop function to the registry. Called by middleware
// factories that spawn background goroutines.
func registerStop(f func()) {
	stopMu.Lock()
	defer stopMu.Unlock()
	stopRegistry = append(stopRegistry, f)
}

// StopAll invokes every registered stop function exactly once and empties
// the registry. Safe to call multiple times — subsequent calls are no-ops.
func StopAll() {
	stopMu.Lock()
	fns := stopRegistry
	stopRegistry = nil
	stopMu.Unlock()
	for _, f := range fns {
		f()
	}
}

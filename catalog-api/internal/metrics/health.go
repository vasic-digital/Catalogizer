package metrics

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
}

// HealthCheckResponse represents the overall health check response
type HealthCheckResponse struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     string                     `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker performs health checks on various components
type HealthChecker struct {
	db        *sql.DB
	startTime time.Time
	version   string
	mu        sync.RWMutex
	checks    map[string]func(context.Context) ComponentHealth
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB, version string) *HealthChecker {
	hc := &HealthChecker{
		db:        db,
		startTime: time.Now(),
		version:   version,
		checks:    make(map[string]func(context.Context) ComponentHealth),
	}

	// Register default checks
	hc.RegisterCheck("database", hc.checkDatabase)

	return hc
}

// RegisterCheck registers a custom health check
func (hc *HealthChecker) RegisterCheck(name string, check func(context.Context) ComponentHealth) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// Check performs all health checks and returns the overall health status
func (hc *HealthChecker) Check(ctx context.Context) HealthCheckResponse {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	components := make(map[string]ComponentHealth)
	overallStatus := HealthStatusHealthy

	// Run all checks
	for name, checkFunc := range hc.checks {
		componentHealth := checkFunc(ctx)
		components[name] = componentHealth

		// Determine overall status
		if componentHealth.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if componentHealth.Status == HealthStatusDegraded && overallStatus != HealthStatusUnhealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return HealthCheckResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    hc.version,
		Uptime:     time.Since(hc.startTime).String(),
		Components: components,
	}
}

// checkDatabase checks database connectivity
func (hc *HealthChecker) checkDatabase(ctx context.Context) ComponentHealth {
	if hc.db == nil {
		return ComponentHealth{
			Status:  HealthStatusUnhealthy,
			Message: "Database not configured",
		}
	}

	start := time.Now()

	// Create context with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := hc.db.PingContext(pingCtx); err != nil {
		return ComponentHealth{
			Status:  HealthStatusUnhealthy,
			Message: "Database ping failed: " + err.Error(),
		}
	}

	latency := time.Since(start)

	// Check connection pool stats
	stats := hc.db.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections-1 {
		return ComponentHealth{
			Status:  HealthStatusDegraded,
			Message: "Database connection pool near limit",
			Latency: latency.String(),
		}
	}

	return ComponentHealth{
		Status:  HealthStatusHealthy,
		Latency: latency.String(),
	}
}

// LivenessProbe returns a simple liveness check (is the service running?)
func (hc *HealthChecker) LivenessProbe() int {
	// Always return 200 if the service is running
	return http.StatusOK
}

// ReadinessProbe returns a readiness check (is the service ready to serve traffic?)
func (hc *HealthChecker) ReadinessProbe(ctx context.Context) int {
	health := hc.Check(ctx)

	if health.Status == HealthStatusHealthy || health.Status == HealthStatusDegraded {
		return http.StatusOK
	}

	return http.StatusServiceUnavailable
}

// StartupProbe returns a startup check (has the service finished initializing?)
func (hc *HealthChecker) StartupProbe(ctx context.Context) int {
	health := hc.Check(ctx)

	if health.Status != HealthStatusUnhealthy {
		return http.StatusOK
	}

	return http.StatusServiceUnavailable
}

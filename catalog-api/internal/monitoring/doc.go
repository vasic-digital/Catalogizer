// Package monitoring provides health monitoring and readiness checks for the Catalogizer API.
//
// It implements health check endpoints that report the status of critical dependencies
// including database connectivity, filesystem availability, and external service reachability.
// The monitoring system supports both liveness probes (is the process running) and readiness
// probes (is the service ready to handle requests) for container orchestration compatibility.
package monitoring

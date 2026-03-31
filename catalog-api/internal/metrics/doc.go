// Package metrics provides Prometheus metrics collection and exposition for the Catalogizer API.
//
// It defines and registers application-specific metrics including request counters, latency
// histograms, active connection gauges, and scan operation metrics. The collected metrics
// are exposed via the /metrics HTTP endpoint for consumption by Prometheus and visualization
// in monitoring dashboards.
package metrics

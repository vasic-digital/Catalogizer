// Package smb provides SMB resilience mechanisms including a circuit breaker, offline cache,
// and exponential backoff retry for the Catalogizer API.
//
// It wraps SMB connection operations with fault tolerance patterns that handle the inherently
// unreliable nature of network file shares. The circuit breaker prevents repeated connection
// attempts to unavailable shares, the offline cache serves stale data during outages, and
// exponential backoff provides controlled retry behavior for transient failures.
package smb

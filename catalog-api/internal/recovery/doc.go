// Package recovery provides fault tolerance mechanisms including panic recovery and graceful
// degradation for the Catalogizer API.
//
// It implements recovery middleware that catches panics in HTTP handlers and converts them to
// 500 responses with structured error logging, preventing individual request failures from
// crashing the entire server. The package also provides circuit breaker patterns for
// protecting the system from cascading failures in external service calls.
package recovery

// Package httpclient provides a pooled HTTP client with connection reuse, configurable
// timeouts, and retry logic for the Catalogizer API.
//
// It wraps the standard net/http client with sensible defaults for connection pooling,
// keep-alive management, and exponential backoff retries. This package is used by metadata
// providers, external API integrations, and the challenge system to make reliable outbound
// HTTP requests without exhausting system resources.
package httpclient

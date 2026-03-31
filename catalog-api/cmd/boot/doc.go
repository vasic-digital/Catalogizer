// Package main is the application entry point and bootstrap for the Catalogizer API server.
//
// It initializes configuration, database connections, service registries, and the Gin HTTP
// router with all middleware and route handlers. The bootstrap process sets up HTTP/3 (QUIC)
// with self-signed TLS certificates, dynamic port binding, and graceful shutdown handling
// including cleanup of WebSocket connections and cache services.
package main

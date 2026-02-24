# ADR-003: HTTP/3 (QUIC) with Brotli Compression Requirement

## Status
Accepted (2026-02-23)

## Context

Catalogizer is a media collection manager that serves large file catalogs across multiple protocols (SMB, FTP, NFS, WebDAV, local). The API handles:

- Browsing catalogs with potentially hundreds of thousands of file entries
- Streaming media content to web, desktop, mobile, and TV clients
- Real-time scan progress updates over WebSocket connections
- Cover art and thumbnail asset delivery
- Multi-client concurrent access during scan operations

HTTP/1.1 has fundamental performance limitations for these workloads:
- Head-of-line blocking means a slow response blocks all subsequent responses on the same connection
- TCP connection setup adds latency for each new connection, especially over high-latency NAS networks
- No multiplexing forces clients to open multiple TCP connections for parallel requests
- TCP slow-start reduces throughput for bursty media browsing patterns

HTTP/2 solves multiplexing but still suffers from TCP-level head-of-line blocking: a single lost TCP packet stalls all multiplexed streams.

HTTP/3 (QUIC) eliminates these issues by running over UDP with per-stream flow control, meaning a lost packet only affects its own stream. This is particularly valuable for:
- Media streaming where packet loss should not stall metadata API responses
- Browsing APIs where multiple parallel requests (entity list + types + stats) should be independent
- High-latency NAS connections where TCP handshake overhead is amplified

Additionally, JSON API responses are highly compressible. Brotli compression typically achieves 15-25% better compression ratios than gzip for JSON payloads, directly reducing bandwidth consumption for clients browsing large catalogs.

## Decision

All Catalogizer network communication uses HTTP/3 (QUIC) as the primary transport protocol, with HTTP/2 + gzip as the fallback, and HTTP/1.1 as the last resort. Brotli compression is the preferred encoding for all API responses.

### Server Implementation (catalog-api)

The API server starts three listeners concurrently:

1. **HTTP/1.1 server** on the configured port (default 8080) for backward compatibility and health checks
2. **HTTPS/HTTP2 server** on port 8443 with TLS for HTTP/2 clients
3. **HTTP/3 (QUIC) server** on UDP port 8443 using `github.com/quic-go/quic-go/http3`

```go
// TLS config supports h3, h2, and http/1.1 via ALPN negotiation
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    NextProtos:   []string{"h3", "h2", "http/1.1"},
}

// HTTP/3 server on UDP
http3Server = &http3.Server{
    Addr:      ":8443",
    Handler:   router,
    TLSConfig: tlsConfig,
}

// Alt-Svc header advertises HTTP/3 availability
router.Use(func(c *gin.Context) {
    c.Header("Alt-Svc", `h3=":8443"; ma=86400`)
    c.Next()
})
```

### TLS Certificate Management

For development, the server generates a self-signed TLS certificate at startup:
- RSA 2048-bit key
- Valid for 1 year
- Covers `localhost` and `127.0.0.1`
- No external certificate management required for development

Production deployments should provide proper certificates via configuration or a reverse proxy.

### Compression Middleware

Brotli compression is applied via `middleware.CompressionMiddleware()` using `github.com/andybalholm/brotli`:
- Brotli is used when the client sends `Accept-Encoding: br`
- gzip is used as fallback when Brotli is not supported
- Compression level is configurable via `middleware.DefaultCompressionConfig()`
- Static assets are pre-compressed where possible

### Client Requirements

- **catalog-web**: Served via HTTP/3-capable reverse proxy; Brotli-compressed static assets built by Vite
- **Tauri desktop apps**: HTTP/3 client for API communication
- **Android apps**: OkHttp with HTTP/3 (Cronet) + Brotli
- **catalogizer-api-client**: HTTP/3-capable fetch with Brotli `Accept-Encoding` header

## Consequences

### Positive

- **Reduced head-of-line blocking**: Independent QUIC streams prevent a slow media download from stalling catalog browsing responses.
- **Faster connection establishment**: QUIC's 0-RTT handshake reduces latency for returning clients, important for mobile and TV apps that frequently reconnect.
- **Better compression**: Brotli's superior compression ratio reduces bandwidth for large JSON responses (entity lists, search results, statistics).
- **Connection migration**: QUIC connections survive network changes (e.g., Wi-Fi to cellular on Android), important for mobile clients.
- **Future-proof**: HTTP/3 is the current standard and browser support is universal.

### Negative

- **Self-signed certificates in development**: Browsers and clients must be configured to trust self-signed certificates during development, adding initial setup friction.
- **UDP firewall rules**: Some corporate networks block UDP traffic, which would prevent HTTP/3. The HTTP/2 fallback handles this case, but at reduced performance.
- **Library maturity**: `quic-go/http3` is a Go implementation of HTTP/3 that, while widely used, has different characteristics than kernel-level TCP stacks. Performance under extreme load may differ from HTTP/2.
- **Three listeners**: Running HTTP, HTTPS, and HTTP/3 servers simultaneously increases resource usage slightly, though the overhead is minimal compared to the database and scanner workloads.
- **Port complexity**: The API is accessible on two ports (HTTP on dynamic port, HTTPS/QUIC on 8443), requiring service discovery via the `.service-port` file and proper documentation of port assignments.

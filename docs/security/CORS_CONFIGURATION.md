# CORS Configuration

Cross-Origin Resource Sharing is handled by the `CORS()` middleware defined in two locations:

- `catalog-api/middleware/request.go` (domain-level middleware)
- `catalog-api/internal/middleware/middleware.go` (infrastructure-level middleware)

Both implementations share the same logic.

## Default Allowed Origins

When `CORS_ALLOWED_ORIGINS` is not set, the following origins are allowed:

```
http://localhost:5173
http://localhost:3000
```

Port 5173 is the Vite dev server default. Port 3000 is the catalog-web dev server port.

## Environment Variable Configuration

Set `CORS_ALLOWED_ORIGINS` to a comma-separated list of allowed origins:

```bash
export CORS_ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"
```

Or in `catalog-api/.env`:

```env
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
```

Leading and trailing whitespace around each origin is trimmed automatically.

## Response Headers

On every request, these headers are set:

| Header | Value |
|--------|-------|
| `Access-Control-Allow-Headers` | `Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With` |
| `Access-Control-Allow-Methods` | `POST, OPTIONS, GET, PUT, DELETE` |

Conditionally (only when the request `Origin` header matches an allowed origin):

| Header | Value |
|--------|-------|
| `Access-Control-Allow-Origin` | The matched origin (not `*`) |
| `Access-Control-Allow-Credentials` | `true` |

## Origin Matching

The middleware performs exact string matching against the allowed origins list. If the request `Origin` header does not match any allowed origin, the `Access-Control-Allow-Origin` header is not set, causing the browser to reject the request.

Wildcard (`*`) is not used because `Access-Control-Allow-Credentials: true` requires a specific origin.

## Preflight Requests

`OPTIONS` requests are handled inline by the CORS middleware:

1. The CORS headers are set as described above
2. The middleware responds with HTTP 204 (No Content)
3. `c.AbortWithStatus(204)` prevents further middleware and handlers from executing

No `Access-Control-Max-Age` header is set, so browsers use their default preflight cache duration (typically 5 seconds for Chrome, 24 hours for Firefox).

## Production Configuration

For production deployments behind nginx, CORS can be handled at the reverse proxy layer instead. The nginx config in `config/nginx.conf` should include appropriate CORS headers. When using the nginx layer, the API middleware CORS can be made a no-op by setting `CORS_ALLOWED_ORIGINS` to the exact production origin.

Recommended production setup:

```env
CORS_ALLOWED_ORIGINS=https://catalogizer.example.com
```

## Domain-Level vs Infrastructure-Level

The two CORS implementations are nearly identical. The domain-level version (`middleware/request.go`) includes `PATCH` in `Access-Control-Allow-Methods` in some routes; the infrastructure version does not. Both use the same environment variable for configuration.

## Source

- `catalog-api/middleware/request.go` (lines 90-119)
- `catalog-api/internal/middleware/middleware.go` (lines 15-44)

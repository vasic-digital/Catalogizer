# Performance Tuning Guide

## Go Runtime

### GOMAXPROCS

```bash
GOMAXPROCS=3 ./catalog-api  # Limit to 3 OS threads
```

For dedicated servers, leave unset (defaults to `runtime.NumCPU()`).

### Test Parallelism

```bash
GOMAXPROCS=3 go test ./... -p 2 -parallel 2
```

## Database

### Connection Pool

| Setting | Default | Dev | Prod |
|---------|---------|-----|------|
| MaxOpenConns | 25 | 10 | 50 |
| MaxIdleConns | 5 | 5 | 25 |
| ConnMaxLifetime | 5m | 5m | 15m |

### SQLite WAL Mode

Enabled via `PRAGMA journal_mode=WAL` after connection. Allows concurrent reads during writes.

### PostgreSQL

- Use PgBouncer for > 100 concurrent connections
- `VACUUM ANALYZE` after large batch operations
- v9 migration adds performance indexes

## HTTP/3 and Compression

### QUIC Protocol

catalog-api uses HTTP/3 via `quic-go/http3` with self-signed TLS certs. Fallback: HTTP/2 + gzip.

### Brotli Compression

Enabled via `andybalholm/brotli` middleware at level 6. Fallback: gzip.

## Cache Configuration

### In-Memory Cache

| Setting | Default |
|---------|---------|
| TTL | 5m |
| MaxSize | 1000 entries |
| CleanupInterval | 1m |

### Redis (Optional)

```env
REDIS_HOST=localhost
REDIS_PORT=6379
```

## Container Resource Limits

| Container | CPU | Memory |
|-----------|-----|--------|
| API | 2 | 4 GB |
| Web | 1 | 2 GB |
| PostgreSQL | 1 | 2 GB |
| Builder | 3 | 8 GB |

Total budget: 4 CPUs, 8 GB RAM across all running containers.

## Monitoring

Use `/metrics` (Prometheus) to track request latency, goroutine count, DB pool utilization, cache hit/miss ratios.

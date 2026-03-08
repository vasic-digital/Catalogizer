# Concurrency Control Architecture

## Overview

Catalogizer uses multiple concurrency patterns across its Go backend to safely handle parallel operations: mutex-based synchronization, semaphores, circuit breakers, and connection pooling.

## Patterns Used

### sync.Mutex / sync.RWMutex

Standard Go mutexes protect shared state throughout the codebase:

| Location | Protected State | Type |
|----------|----------------|------|
| `Lazy/pkg/lazy/` | Value[T] and Service[T] cached results | `sync.Mutex` |
| `internal/cache/` | In-memory cache entries | `sync.RWMutex` |
| `internal/media/realtime/` | WebSocket client registry | `sync.RWMutex` |
| `internal/smb/resilience.go` | Circuit breaker state | `sync.Mutex` |
| `internal/metrics/` | Prometheus counter maps | `sync.Mutex` |

### sync.Once

Used for one-time initialization:

- `Lazy/pkg/lazy/Value[T]` — loader runs exactly once per instance
- `Lazy/pkg/lazy/Service[T]` — service init runs exactly once
- Database connection pool initialization
- Prometheus metrics registration

### Semaphore Pattern (Concurrency Module)

The `Concurrency/` module provides a weighted semaphore:

```go
sem := concurrency.NewSemaphore(maxWorkers)
sem.Acquire()
defer sem.Release()
```

Used in scanner parallel file processing, batch metadata enrichment, concurrent SMB traversal.

### Circuit Breaker (Recovery Module)

States: Closed -> Open -> Half-Open -> Closed

- **Closed**: Requests pass through; failures counted
- **Open**: Requests fail immediately; timer starts
- **Half-Open**: Limited requests to test recovery

Used in `internal/smb/resilience.go` for NAS connectivity.

### Connection Pooling

- **Database**: `database/sql` built-in pool with `SetMaxOpenConns`
- **SMB**: Custom `SmbConnectionPool` with per-host reuse
- **HTTP**: `http.Client` with `Transport.MaxIdleConnsPerHost`

## Resource Limits

| Resource | Limit |
|----------|-------|
| `GOMAXPROCS` | 3 |
| Test `-p` | 2 |
| Test `-parallel` | 2 |
| Container CPU total | 4 CPUs |
| Container memory total | 8 GB |

## Key Constraints

1. No global locks — each component manages its own synchronization
2. Prefer RWMutex for read-heavy workloads (cache, metrics)
3. Always use defer for unlock/release to prevent deadlocks
4. Context propagation — all long-running operations accept `context.Context`
5. Graceful shutdown — `wg.Wait()` with timeout (10s) for cleanup

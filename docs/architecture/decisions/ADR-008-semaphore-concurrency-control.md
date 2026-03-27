# ADR-008: Semaphore-Based Concurrency Control

## Status
Accepted (2026-03-18)

## Context

Catalogizer performs many operations that are naturally concurrent: scanning files across multiple storage roots, querying metadata providers (TMDB, OMDB, OpenLibrary, MusicBrainz) for enrichment, converting media formats, downloading subtitles from multiple providers simultaneously, and processing detection pipeline stages in parallel.

Without concurrency limits, these parallel operations cause several critical problems:

1. **External API rate limiting**: Metadata providers enforce rate limits (TMDB: 40 requests per 10 seconds, MusicBrainz: 1 request per second for unauthenticated clients). Unbounded goroutines fire hundreds of requests simultaneously, triggering rate limits, temporary bans, and cascading retry storms that make the situation worse.

2. **Local resource exhaustion**: File scanning with unbounded parallelism can open thousands of network connections (especially over SMB/NFS), exhaust file descriptors, saturate NAS device CPU, and cause connection timeouts that appear as scan failures.

3. **Host resource limits**: The mandatory 30-40% host resource cap (documented in CLAUDE.md) means Catalogizer cannot afford to spike CPU or memory usage. Unbounded goroutines can easily push past this limit, freezing the host machine.

4. **Unpredictable behavior**: Without concurrency control, system behavior varies dramatically between a scan of 100 files and 100,000 files. Small tests pass but production workloads fail -- a classic "works on my machine" problem.

5. **No backpressure**: When the system is overloaded, new operations are accepted without limit, compounding the overload rather than applying backpressure to callers.

## Decision

We implement a `BoundedSemaphore` primitive in `internal/concurrency/` that provides configurable, timeout-aware concurrency limiting for all parallel operations. Every subsystem that performs concurrent work must acquire a semaphore permit before proceeding.

### Architecture

```
Caller (handler/service)
    |
    | semaphore.Acquire(ctx)  -- blocks until permit available or context canceled
    v
BoundedSemaphore (channel-based, configurable capacity)
    |
    | permit granted
    v
Concurrent Operation (API call, file scan, conversion)
    |
    | defer semaphore.Release()
    v
Permit returned to pool
```

### Implementation

```go
type BoundedSemaphore struct {
    permits chan struct{}
    limit   int
    timeout time.Duration
}

func NewBoundedSemaphore(limit int, opts ...Option) *BoundedSemaphore {
    s := &BoundedSemaphore{
        permits: make(chan struct{}, limit),
        limit:   limit,
        timeout: 30 * time.Second, // default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

func (s *BoundedSemaphore) Acquire(ctx context.Context) error {
    select {
    case s.permits <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(s.timeout):
        return ErrAcquireTimeout
    }
}

func (s *BoundedSemaphore) Release() {
    <-s.permits
}
```

### Semaphore Allocation by Subsystem

Each subsystem has its own semaphore with limits tuned to its characteristics:

| Subsystem | Default Limit | Rationale |
|-----------|--------------|-----------|
| Metadata providers (TMDB) | 4 | Stay well under TMDB's 40/10s rate limit |
| Metadata providers (MusicBrainz) | 1 | MusicBrainz enforces 1 req/s for unauthenticated |
| File scanning (per root) | 8 | Balance scan speed vs NAS connection limits |
| Subtitle download | 3 | OpenSubtitles rate limits, multiple providers |
| Format conversion | 2 | CPU-intensive, respect host resource cap |
| HTTP client pool (outbound) | 16 | Global cap on concurrent outbound connections |

### Configuration

Semaphore limits are configurable via `config.json` and overridable via environment variables, allowing operators to tune for their deployment:

```json
{
  "concurrency": {
    "metadata_tmdb": 4,
    "metadata_musicbrainz": 1,
    "file_scan": 8,
    "subtitle_download": 3,
    "format_conversion": 2,
    "http_outbound": 16
  }
}
```

Environment variable override example: `CONCURRENCY_FILE_SCAN=16` doubles the scan parallelism for a deployment with a high-performance NAS.

### Key Design Choices

- **Channel-based, not sync.Mutex**: Channels provide natural `select`-based timeout and cancellation support via `context.Context`. Mutex-based approaches require manual timeout logic.
- **Per-subsystem isolation**: Each subsystem has its own semaphore rather than a global limit. This prevents a slow metadata provider from blocking file scans.
- **Timeout with error**: `Acquire()` returns `ErrAcquireTimeout` rather than blocking forever. Callers can handle backpressure (return 503, queue the request, log and skip).
- **Context-aware cancellation**: If a request's context is canceled (client disconnect, HTTP timeout), the `Acquire()` call unblocks immediately rather than holding a goroutine indefinitely.

## Consequences

### Positive

- **Predictable resource usage**: Every parallel operation has a known upper bound. System behavior is consistent whether processing 100 or 100,000 files.
- **Respects host resource limits**: The 30-40% host resource cap is enforceable because concurrent operations are bounded. `podman stats` shows stable, predictable resource consumption.
- **Prevents cascading failures**: When a metadata provider is slow, the semaphore limits the number of in-flight requests. Other subsystems continue operating normally because they have independent semaphores.
- **Configurable per deployment**: A Raspberry Pi can set `CONCURRENCY_FILE_SCAN=2` while a rack server can set `CONCURRENCY_FILE_SCAN=32`. The same binary adapts to the deployment environment.
- **Backpressure signaling**: Callers receive explicit errors (`ErrAcquireTimeout`, `context.Canceled`) when the system is overloaded, enabling graceful degradation rather than silent queuing.

### Negative

- **Reduced throughput at default settings**: Conservative default limits (e.g., 4 concurrent TMDB requests) may underutilize fast network connections and powerful hardware. Operators must tune limits for their environment.
- **Additional complexity**: Every concurrent operation must acquire/release a semaphore. Forgetting `defer semaphore.Release()` causes a permit leak that progressively reduces available concurrency until the system deadlocks.
- **Timeout tuning**: The default 30-second acquire timeout may be too short for deployments with very slow storage (degraded NAS, high-latency WAN) or too long for interactive requests where the user expects a fast response.
- **Per-subsystem limits don't account for aggregate load**: A deployment might be within limits for each individual subsystem but still exceed the host resource cap when all subsystems are active simultaneously. The per-subsystem approach assumes subsystems rarely peak concurrently.

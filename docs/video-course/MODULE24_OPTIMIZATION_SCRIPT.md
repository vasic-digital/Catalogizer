# Module 24: Advanced Optimization -- Video Script

**Duration**: 55 minutes
**Prerequisites**: Module 2 (Backend Development), Module 5 (Frontend Development), Module 15 (Concurrency Patterns)

---

## Video 24.1: Go Lazy Loading with LazyServiceRegistry (12 min)

### Opening

Welcome to Module 24, our advanced optimization module. Performance is not an afterthought in Catalogizer -- it is a design constraint enforced by the 30-40% host resource limit. This module covers lazy loading on both backend and frontend, semaphore-based concurrency control, non-blocking health checks, virtual scrolling, image lazy loading, and a methodology for measuring performance with k6.

### The Problem: Startup Cost

**[Visual: Timeline showing a naive startup that initializes all services eagerly]**

**Narrator**: A naive application initializes every service at startup: database connections, Redis clients, metadata providers, subtitle providers, cache layers, and WebSocket handlers. On Catalogizer's host, this spike consumes too many resources at once and delays the first request by several seconds.

**[Visual: Timeline showing deferred initialization with LazyServiceRegistry]**

**Narrator**: The `LazyServiceRegistry` in `internal/lifecycle/` solves this. Services register their initialization functions but do not execute them until first use. This spreads the startup cost across the first few requests instead of concentrating it at boot.

### LazyServiceRegistry Design

**[Visual: Open `catalog-api/internal/lifecycle/lazy_registry.go`]**

**Narrator**: The registry maps service names to lazy initializers. Each initializer wraps a `sync.Once` to guarantee exactly-once execution, even under concurrent access.

```go
// catalog-api/internal/lifecycle/lazy_registry.go
type LazyServiceRegistry struct {
    mu       sync.RWMutex
    services map[string]*lazyEntry
    logger   *zap.Logger
    order    []string // dependency ordering
}

type lazyEntry struct {
    once     sync.Once
    init     func() (interface{}, error)
    instance interface{}
    err      error
    deps     []string
}

func NewLazyServiceRegistry(logger *zap.Logger) *LazyServiceRegistry {
    return &LazyServiceRegistry{
        services: make(map[string]*lazyEntry),
        logger:   logger,
    }
}
```

### Registration and Resolution

**[Visual: Show service registration in `main.go`]**

**Narrator**: Services are registered during application setup with their dependencies declared explicitly. The registry resolves dependencies in topological order when a service is first requested.

```go
// catalog-api/main.go (simplified)
func setupServices(registry *lifecycle.LazyServiceRegistry, db *database.DB) {
    registry.Register("cache", nil, func() (interface{}, error) {
        return services.NewCacheService(db, logger), nil
    })

    registry.Register("subtitle", []string{"cache"}, func() (interface{}, error) {
        cache := registry.MustGet("cache").(*services.CacheService)
        return services.NewSubtitleService(db, cache, logger), nil
    })

    registry.Register("aggregation", []string{"cache"}, func() (interface{}, error) {
        cache := registry.MustGet("cache").(*services.CacheService)
        return services.NewAggregationService(db, cache, logger), nil
    })
}
```

**[Visual: Show the `Get` method resolving a service lazily]**

**Narrator**: When a handler needs a service, it calls `registry.Get("subtitle")`. The first call triggers initialization of the subtitle service and its dependency (cache). Subsequent calls return the cached instance immediately.

```go
// catalog-api/internal/lifecycle/lazy_registry.go
func (r *LazyServiceRegistry) Get(name string) (interface{}, error) {
    r.mu.RLock()
    entry, ok := r.services[name]
    r.mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("service not registered: %s", name)
    }

    entry.once.Do(func() {
        // Resolve dependencies first
        for _, dep := range entry.deps {
            if _, err := r.Get(dep); err != nil {
                entry.err = fmt.Errorf("dependency %s: %w", dep, err)
                return
            }
        }

        r.logger.Info("Initializing service", zap.String("name", name))
        entry.instance, entry.err = entry.init()
    })

    return entry.instance, entry.err
}

func (r *LazyServiceRegistry) MustGet(name string) interface{} {
    svc, err := r.Get(name)
    if err != nil {
        panic(fmt.Sprintf("service %s: %v", name, err))
    }
    return svc
}
```

**Narrator**: The `sync.Once` inside each entry means that even if 100 goroutines call `Get("cache")` simultaneously, the cache is initialized exactly once. All callers block until initialization completes, then receive the same instance.

---

## Video 24.2: React.lazy and Code Splitting (10 min)

### Frontend Bundle Problem

**[Visual: Webpack bundle analyzer showing a monolithic 2MB JavaScript bundle]**

**Narrator**: Without code splitting, the frontend ships a single JavaScript bundle containing every page, component, and library. Users pay the full download and parse cost even if they only visit the login page.

### Route-Level Code Splitting

**[Visual: Show Vite chunk configuration in `vite.config.ts`]**

**Narrator**: Catalogizer splits the frontend into route-level chunks using `React.lazy` and dynamic imports. Each page is a separate chunk loaded on demand when the user navigates to it.

```typescript
// catalog-web/src/App.tsx
import { lazy, Suspense } from 'react';
import { Routes, Route } from 'react-router-dom';
import LoadingSpinner from '@/components/LoadingSpinner';

const Dashboard = lazy(() => import('@/pages/Dashboard'));
const EntityBrowser = lazy(() => import('@/pages/EntityBrowser'));
const EntityDetail = lazy(() => import('@/pages/EntityDetail'));
const Collections = lazy(() => import('@/pages/Collections'));
const Settings = lazy(() => import('@/pages/Settings'));
const AIMetadataDashboard = lazy(() => import('@/pages/AIMetadataDashboard'));
const SubtitleManager = lazy(() => import('@/pages/SubtitleManager'));
const ConversionJobs = lazy(() => import('@/pages/ConversionJobs'));

export default function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/browse" element={<EntityBrowser />} />
        <Route path="/entity/:id" element={<EntityDetail />} />
        <Route path="/collections" element={<Collections />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/dashboard/ai" element={<AIMetadataDashboard />} />
        <Route path="/subtitles" element={<SubtitleManager />} />
        <Route path="/conversions" element={<ConversionJobs />} />
      </Routes>
    </Suspense>
  );
}
```

### Vendor Chunk Strategy

**[Visual: Show the `manualChunks` configuration in `vite.config.ts`]**

**Narrator**: Beyond route-level splitting, Catalogizer extracts vendor libraries into separate chunks that cache independently. When application code changes, users only re-download the app chunks -- the vendor chunks remain cached.

```typescript
// catalog-web/vite.config.ts (build.rollupOptions.output.manualChunks)
manualChunks(id) {
  if (id.includes('node_modules')) {
    if (id.includes('react') || id.includes('react-dom')) {
      return 'vendor';
    }
    if (id.includes('react-router')) {
      return 'router';
    }
    if (id.includes('@radix-ui') || id.includes('class-variance-authority')) {
      return 'ui';
    }
    if (id.includes('recharts') || id.includes('d3')) {
      return 'charts';
    }
    if (id.includes('date-fns') || id.includes('zod') || id.includes('zustand')) {
      return 'utils';
    }
  }
}
```

**[Visual: Show network tab with chunks loading on navigation]**

**Narrator**: The result: the initial page load downloads only the vendor chunk and the current route's chunk. Navigating to the entity browser triggers a small network request for its chunk. The total JavaScript parsed on first load drops from 2MB to under 400KB.

### Prefetching

**[Visual: Show link prefetch hints]**

**Narrator**: For routes the user is likely to visit next, Catalogizer prefetches chunks on hover or during idle time. This eliminates the loading spinner for common navigation paths.

```typescript
// catalog-web/src/components/NavLink.tsx
function NavLink({ to, children }: NavLinkProps) {
  const prefetch = () => {
    // Trigger dynamic import to start chunk download
    const routes: Record<string, () => Promise<unknown>> = {
      '/browse': () => import('@/pages/EntityBrowser'),
      '/collections': () => import('@/pages/Collections'),
      '/dashboard/ai': () => import('@/pages/AIMetadataDashboard'),
    };
    routes[to]?.();
  };

  return (
    <Link to={to} onMouseEnter={prefetch}>
      {children}
    </Link>
  );
}
```

---

## Video 24.3: Semaphore-Based Concurrency with BoundedSemaphore (10 min)

### Why Bound Parallelism?

**[Visual: Show system load spiking when 50 concurrent scans run unbounded]**

**Narrator**: Unbounded parallelism is the fastest way to freeze a host machine. When the scanner processes 50 storage roots simultaneously, CPU usage hits 100%, memory fills, and the OOM killer starts terminating processes. The `BoundedSemaphore` prevents this.

### BoundedSemaphore from digital.vasic.concurrency

**[Visual: Open `Concurrency/pkg/semaphore/bounded.go`]**

**Narrator**: The `BoundedSemaphore` is a counting semaphore with a fixed capacity. It limits the number of concurrent operations to a configurable maximum. The implementation uses a buffered channel, which is the idiomatic Go approach.

```go
// Concurrency/pkg/semaphore/bounded.go
type BoundedSemaphore struct {
    tokens chan struct{}
    size   int
}

func NewBoundedSemaphore(maxConcurrent int) *BoundedSemaphore {
    return &BoundedSemaphore{
        tokens: make(chan struct{}, maxConcurrent),
        size:   maxConcurrent,
    }
}

func (s *BoundedSemaphore) Acquire(ctx context.Context) error {
    select {
    case s.tokens <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (s *BoundedSemaphore) Release() {
    <-s.tokens
}

func (s *BoundedSemaphore) TryAcquire() bool {
    select {
    case s.tokens <- struct{}{}:
        return true
    default:
        return false
    }
}

func (s *BoundedSemaphore) Available() int {
    return s.size - len(s.tokens)
}
```

### Usage in Catalogizer

**[Visual: Show semaphore usage in the scanner, asset manager, and HTTP middleware]**

**Narrator**: The semaphore appears in three critical paths.

**Narrator**: First, the `UniversalScanner` limits concurrent scan operations. The `maxConcurrentScans` value defaults to 2, meaning at most two storage roots are scanned simultaneously.

```go
// catalog-api/internal/services/universal_scanner.go
func (s *UniversalScanner) processJob(ctx context.Context, job ScanJob) {
    if err := s.scanSem.Acquire(ctx); err != nil {
        s.logger.Warn("Scan cancelled while waiting for semaphore",
            zap.String("job_id", job.ID))
        return
    }
    defer s.scanSem.Release()

    // Scan proceeds with bounded concurrency
    s.executeScan(ctx, job)
}
```

**Narrator**: Second, the asset manager limits concurrent resolution workers to 4, preventing a burst of thumbnail requests from overwhelming external APIs.

**Narrator**: Third, the `ConcurrencyLimiter` middleware caps in-flight HTTP requests at 100. When the limit is reached, new requests receive a 503 Service Unavailable response with a Retry-After header.

```go
// catalog-api/middleware/concurrency.go
func ConcurrencyLimiter(maxConcurrent int) gin.HandlerFunc {
    sem := semaphore.NewBoundedSemaphore(maxConcurrent)

    return func(c *gin.Context) {
        if !sem.TryAcquire() {
            c.Header("Retry-After", "5")
            c.AbortWithStatusJSON(503, gin.H{
                "error": "server at capacity, retry later",
            })
            return
        }
        defer sem.Release()
        c.Next()
    }
}
```

---

## Video 24.4: Non-Blocking Health Checks (8 min)

### The Health Check Problem

**[Visual: Show a load balancer health check timing out because the database query is slow]**

**Narrator**: Health checks must respond instantly. If a health check queries the database and the database is under heavy load, the check times out, the load balancer marks the instance as unhealthy, and traffic is diverted -- making the overload worse.

### Catalogizer's Health Check Design

**[Visual: Open `catalog-api/internal/handlers/health_handler.go`]**

**Narrator**: Catalogizer's health endpoint returns cached status rather than performing live checks. A background goroutine probes the database, Redis, and external services every 10 seconds, updating an atomic status value. The health handler reads this value without any I/O.

```go
// catalog-api/internal/handlers/health_handler.go
type HealthHandler struct {
    status atomic.Value // stores *HealthStatus
    stopCh chan struct{}
    wg     sync.WaitGroup
    db     *database.DB
    logger *zap.Logger
}

type HealthStatus struct {
    Status    string            `json:"status"` // healthy, degraded, unhealthy
    Uptime    time.Duration     `json:"uptime"`
    Checks    map[string]string `json:"checks"`
    Timestamp time.Time         `json:"timestamp"`
}

func (h *HealthHandler) GetHealth(c *gin.Context) {
    status := h.status.Load().(*HealthStatus)
    if status.Status == "unhealthy" {
        c.JSON(503, status)
        return
    }
    c.JSON(200, status)
}
```

**[Visual: Show the background probe goroutine]**

**Narrator**: The probe goroutine runs individual checks with tight timeouts. A database ping that takes longer than 2 seconds is marked as degraded. A failed ping is marked as unhealthy. The results are stored atomically, so the health handler never blocks.

```go
// catalog-api/internal/handlers/health_handler.go
func (h *HealthHandler) probeLoop() {
    defer h.wg.Done()
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-h.stopCh:
            return
        case <-ticker.C:
            status := h.runChecks()
            h.status.Store(status)
        }
    }
}

func (h *HealthHandler) runChecks() *HealthStatus {
    checks := make(map[string]string)

    // Database check with 2-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if err := h.db.PingContext(ctx); err != nil {
        checks["database"] = "unhealthy"
    } else {
        checks["database"] = "healthy"
    }

    // Determine overall status
    overall := "healthy"
    for _, v := range checks {
        if v == "unhealthy" {
            overall = "unhealthy"
            break
        }
    }

    return &HealthStatus{
        Status:    overall,
        Checks:    checks,
        Timestamp: time.Now(),
    }
}
```

---

## Video 24.5: Virtual Scrolling and Image Lazy Loading (7 min)

### The Rendering Problem

**[Visual: Browser showing jank when rendering 10,000 entity cards]**

**Narrator**: A large media library can contain tens of thousands of entities. Rendering all of them into the DOM at once causes multi-second layout calculations, excessive memory usage, and visible jank during scrolling.

### Virtual Scrolling

**[Visual: Show the entity browser with virtual scrolling enabled, DOM node count staying constant]**

**Narrator**: Virtual scrolling renders only the items visible in the viewport plus a small buffer. As the user scrolls, items entering the viewport are rendered and items leaving are unmounted. The DOM node count stays constant regardless of list size.

```typescript
// catalog-web/src/components/VirtualEntityGrid.tsx
import { useVirtualizer } from '@tanstack/react-virtual';

export function VirtualEntityGrid({ items }: { items: MediaItem[] }) {
  const parentRef = useRef<HTMLDivElement>(null);
  const columnCount = useResponsiveColumns(); // 2-6 based on viewport

  const rowCount = Math.ceil(items.length / columnCount);

  const virtualizer = useVirtualizer({
    count: rowCount,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 320, // estimated row height in pixels
    overscan: 3, // render 3 extra rows above/below viewport
  });

  return (
    <div ref={parentRef} className="h-full overflow-auto">
      <div style={{ height: virtualizer.getTotalSize() }}>
        {virtualizer.getVirtualItems().map((virtualRow) => (
          <div
            key={virtualRow.key}
            style={{
              position: 'absolute',
              top: virtualRow.start,
              height: virtualRow.size,
              width: '100%',
            }}
          >
            <div className="grid grid-cols-responsive gap-4">
              {items
                .slice(
                  virtualRow.index * columnCount,
                  (virtualRow.index + 1) * columnCount
                )
                .map((item) => (
                  <EntityCard key={item.id} item={item} />
                ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
```

### Image Lazy Loading

**[Visual: Network tab showing images loading only as they scroll into view]**

**Narrator**: Cover art images use the native `loading="lazy"` attribute combined with an Intersection Observer for browsers that need additional control. A low-quality placeholder is shown immediately, and the full image loads when the card enters the viewport.

```typescript
// catalog-web/src/components/LazyImage.tsx
export function LazyImage({ src, alt, className }: LazyImageProps) {
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState(false);
  const imgRef = useRef<HTMLImageElement>(null);

  return (
    <div className={cn('relative overflow-hidden bg-gray-200', className)}>
      {!loaded && !error && (
        <div className="absolute inset-0 animate-pulse bg-gray-300" />
      )}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center">
          <FilmIcon className="h-8 w-8 text-gray-400" />
        </div>
      )}
      <img
        ref={imgRef}
        src={src}
        alt={alt}
        loading="lazy"
        onLoad={() => setLoaded(true)}
        onError={() => setError(true)}
        className={cn(
          'h-full w-full object-cover transition-opacity duration-300',
          loaded ? 'opacity-100' : 'opacity-0'
        )}
      />
    </div>
  );
}
```

**Narrator**: The combination of virtual scrolling and image lazy loading means that a library of 50,000 entities renders as smoothly as a library of 50. The browser only holds DOM nodes and images for the visible viewport.

---

## Video 24.6: Performance Tuning Methodology with k6 (8 min)

### The Methodology

**[Visual: Flowchart: Baseline -> Identify Bottleneck -> Optimize -> Verify -> Repeat]**

**Narrator**: Performance tuning without measurement is guesswork. Catalogizer uses a structured methodology: establish a baseline with k6 load tests, identify bottlenecks with profiling, apply targeted optimizations, verify improvement with the same k6 tests, and repeat until targets are met.

### k6 Load Test Suite

**[Visual: Open `tests/k6/load_test.js`]**

**Narrator**: The k6 test suite has three profiles. The load test ramps to 50 virtual users and verifies that the 95th percentile response time stays below 500 milliseconds. The stress test ramps to 300 users to find the breaking point. The soak test runs 20 users for 30 minutes to detect memory leaks.

```javascript
// tests/k6/load_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const entityListDuration = new Trend('entity_list_duration');

export const options = {
  stages: [
    { duration: '2m', target: 10 },  // ramp up
    { duration: '5m', target: 50 },  // sustain
    { duration: '2m', target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95th percentile < 500ms
    errors: ['rate<0.01'],              // error rate < 1%
  },
};

export default function () {
  // Entity list endpoint (most common operation)
  const listRes = http.get('http://localhost:8080/api/v1/entities?limit=50');
  entityListDuration.add(listRes.timings.duration);
  check(listRes, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'has items': (r) => JSON.parse(r.body).items.length > 0,
  });

  // Entity detail endpoint
  const detailRes = http.get('http://localhost:8080/api/v1/entities/1');
  check(detailRes, {
    'detail status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
```

### Running k6

**[Visual: Terminal showing k6 running in a container]**

**Narrator**: k6 runs in a container to avoid installing it on the host. The test scripts are mounted as a volume.

```bash
# Load test
podman run --rm --network host \
    -v $(pwd)/tests/k6:/scripts \
    docker.io/grafana/k6:latest \
    run /scripts/load_test.js

# Stress test (find breaking point)
podman run --rm --network host \
    -v $(pwd)/tests/k6:/scripts \
    docker.io/grafana/k6:latest \
    run /scripts/stress_test.js

# Soak test (detect memory leaks, 30 minutes)
podman run --rm --network host \
    -v $(pwd)/tests/k6:/scripts \
    docker.io/grafana/k6:latest \
    run /scripts/soak_test.js
```

### Interpreting Results

**[Visual: k6 output showing request durations, error rates, and threshold pass/fail]**

**Narrator**: k6 outputs a summary with request counts, durations (min, median, p90, p95, max), error rates, and threshold results. A red threshold means the optimization target was missed. The key metrics to watch are:

| Metric | Target | Why |
|--------|--------|-----|
| `http_req_duration p(95)` | < 500ms | User-perceived responsiveness |
| `http_req_failed` | < 1% | Reliability under load |
| `http_reqs` | > 100/s | Throughput capacity |
| `iteration_duration` | stable over time | No degradation (soak test) |

**Narrator**: After each optimization -- adding an index, enabling caching, reducing a query -- re-run the same k6 profile. Compare the p95 duration before and after. If it improved, keep the change. If not, revert and investigate further.

---

## Key Code Examples

### Optimization Checklist
```
Backend:
  [ ] LazyServiceRegistry for deferred initialization
  [ ] BoundedSemaphore on all concurrent operations
  [ ] Non-blocking health checks with atomic status
  [ ] Database connection pool tuning (MaxOpen=25, MaxIdle=10)
  [ ] SQL query optimization with EXPLAIN ANALYZE
  [ ] Response caching via CacheService (TTL per resource type)

Frontend:
  [ ] React.lazy for route-level code splitting
  [ ] Vendor chunk extraction (react, router, ui, charts, utils)
  [ ] Virtual scrolling for large lists
  [ ] Image lazy loading with loading="lazy"
  [ ] React Query stale-while-revalidate for API responses
  [ ] Prefetching on hover for likely navigation targets

Measurement:
  [ ] k6 load test: p95 < 500ms at 50 users
  [ ] k6 stress test: identify breaking point
  [ ] k6 soak test: no memory leaks over 30 minutes
  [ ] Prometheus metrics for runtime monitoring
```

---

## Key Files Referenced

- `catalog-api/internal/lifecycle/lazy_registry.go` -- LazyServiceRegistry
- `catalog-api/main.go` -- Service registration and startup
- `catalog-api/middleware/concurrency.go` -- ConcurrencyLimiter middleware
- `catalog-api/internal/handlers/health_handler.go` -- Non-blocking health checks
- `Concurrency/pkg/semaphore/bounded.go` -- BoundedSemaphore implementation
- `catalog-web/vite.config.ts` -- Code splitting and chunk configuration
- `catalog-web/src/App.tsx` -- React.lazy route definitions
- `catalog-web/src/components/VirtualEntityGrid.tsx` -- Virtual scrolling
- `catalog-web/src/components/LazyImage.tsx` -- Image lazy loading
- `tests/k6/load_test.js` -- k6 load test script
- `tests/k6/stress_test.js` -- k6 stress test script
- `tests/k6/soak_test.js` -- k6 soak test script

---

## Exercises

1. Add a new service to the `LazyServiceRegistry` that depends on two other services. Write a test verifying that dependency resolution works correctly and that initialization happens exactly once even under concurrent access.
2. Modify the `ConcurrencyLimiter` middleware to track the number of rejected requests as a Prometheus counter and add a Grafana dashboard panel for it.
3. Implement a prefetching strategy that preloads the entity detail page chunk when the user hovers over an entity card for more than 200 milliseconds.
4. Write a k6 custom metric that measures the time from scan start to entity appearing in the API response (end-to-end latency).
5. Profile the entity list endpoint with `go tool pprof` and identify the top 3 CPU consumers. Propose optimizations for each.

---

## Quiz Questions

1. How does the `LazyServiceRegistry` ensure thread-safe, exactly-once initialization?
   **Answer**: Each service entry contains a `sync.Once`. When `Get()` is called, the `sync.Once.Do()` executes the initialization function exactly once. If multiple goroutines call `Get()` concurrently, all but the first block until initialization completes, then all receive the same instance.

2. Why does the health check use `atomic.Value` instead of querying the database directly?
   **Answer**: A direct database query could timeout under load, causing the load balancer to mark the instance as unhealthy. Using `atomic.Value`, the health handler reads a pre-computed status with zero I/O. A background goroutine probes the database every 10 seconds with a 2-second timeout and updates the atomic value.

3. What is the benefit of virtual scrolling over pagination?
   **Answer**: Virtual scrolling provides a seamless, continuous scroll experience without page boundaries. The DOM node count stays constant regardless of total item count, maintaining consistent performance. Pagination requires user interaction to advance and can lose scroll context.

4. What are the three k6 test profiles, and what does each measure?
   **Answer**: Load test: ramps to 50 users, verifies p95 response time below 500ms (normal capacity). Stress test: ramps to 300 users, finds the breaking point where errors start (maximum capacity). Soak test: maintains 20 users for 30 minutes, detects memory leaks and performance degradation over time (stability).

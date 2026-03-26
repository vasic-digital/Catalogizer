# Catalogizer Project Completion - Master Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete all unfinished work across the Catalogizer project — fix stubs, harden security, maximize test coverage, optimize performance, and finalize all documentation, courses, and website content.

**Architecture:** 10 phases executed sequentially, each producing working, tested software. Phases are ordered by dependency: code fixes first (Phase 1-4), then testing (Phase 5-6), then documentation (Phase 7-9), then final validation (Phase 10). Each phase has verification gates.

**Tech Stack:** Go 1.24/Gin, React 18/TypeScript/Vite, Kotlin/Compose, Tauri/Rust, Podman, SonarQube, Snyk, Trivy, Semgrep, k6, Playwright, Vitest

**Constraints:**
- Podman only (no Docker)
- Host resource limits: 30-40% max (GOMAXPROCS=3, container budget: 4 CPUs / 8 GB)
- No GitHub Actions
- No interactive processes (no sudo, no root prompts)
- HTTP/3 (QUIC) + Brotli mandatory
- All builds/QA in containers

---

## Phase 1: Code Fixes — Stubs, Placeholders, Dead Code Activation

**Objective:** Eliminate all stub implementations, placeholder code, hardcoded values, and activate unused submodules.

### Task 1.1: Fix Hardcoded URLs in Frontend

**Files:**
- Modify: `catalog-web/src/lib/collectionsApi.ts:145,328`
- Test: `catalog-web/src/lib/__tests__/collectionsApi.test.ts`

- [ ] **Step 1: Write failing test for share URL generation**

In `catalog-web/src/lib/__tests__/collectionsApi.test.ts`, add:

```typescript
describe('generateShareUrl', () => {
  it('should use window.location.origin instead of hardcoded localhost', () => {
    // Mock window.location
    const originalLocation = window.location;
    Object.defineProperty(window, 'location', {
      value: { ...originalLocation, origin: 'https://my-catalogizer.example.com' },
      writable: true,
    });

    const api = new CollectionsApi();
    const result = api.generateShareUrl('test-id-123');
    expect(result).toContain('https://my-catalogizer.example.com/shared/');
    expect(result).not.toContain('localhost');

    // Restore
    Object.defineProperty(window, 'location', { value: originalLocation, writable: true });
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd catalog-web && npm run test -- --run src/lib/__tests__/collectionsApi.test.ts`
Expected: FAIL — share URL contains `localhost:3006`

- [ ] **Step 3: Fix hardcoded URLs**

In `catalog-web/src/lib/collectionsApi.ts`, replace both occurrences (lines 145 and 328):

```typescript
// OLD (line 145):
share_url: `http://localhost:3006/shared/share_${id}_${Date.now()}`
// NEW:
share_url: `${window.location.origin}/shared/share_${id}_${Date.now()}`
```

Apply same fix at line 328.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd catalog-web && npm run test -- --run src/lib/__tests__/collectionsApi.test.ts`
Expected: PASS

- [ ] **Step 5: Fix `any` return type on getCollectionItems**

In `catalog-web/src/lib/collectionsApi.ts:79`, replace:

```typescript
// OLD:
async getCollectionItems(id: string, page = 1, limit = 50): Promise<any>
// NEW:
async getCollectionItems(id: string, page = 1, limit = 50): Promise<CollectionItemsResponse>
```

Add the interface at the top of the file (after existing interfaces):

```typescript
interface CollectionItemsResponse {
  items: CollectionItem[];
  total: number;
  page: number;
  limit: number;
}
```

- [ ] **Step 6: Run type-check and tests**

Run: `cd catalog-web && npm run type-check && npm run test -- --run src/lib/__tests__/collectionsApi.test.ts`
Expected: PASS (no type errors, tests pass)

- [ ] **Step 7: Commit**

```bash
cd catalog-web
git add src/lib/collectionsApi.ts src/lib/__tests__/collectionsApi.test.ts
git commit -m "fix: replace hardcoded localhost URLs with window.location.origin, type collectionsApi"
```

---

### Task 1.2: Fix Android TV API Configuration

**Files:**
- Modify: `catalogizer-androidtv/app/build.gradle.kts:52-63`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/models/Settings.kt:25`

- [ ] **Step 1: Fix build.gradle.kts debug endpoint**

In `catalogizer-androidtv/app/build.gradle.kts`, change debug buildType (around line 52):

```kotlin
// OLD:
buildConfigField("String", "API_BASE_URL", "\"https://catalogizer.dev\"")
// NEW:
buildConfigField("String", "API_BASE_URL", "\"http://10.0.2.2:8080\"")
```

Keep release as `https://catalogizer.dev`.

- [ ] **Step 2: Update Settings.kt default**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/models/Settings.kt:25`:

```kotlin
// OLD:
const val DEFAULT_SERVER_URL = "https://catalogizer.dev"
// NEW:
const val DEFAULT_SERVER_URL = BuildConfig.API_BASE_URL
```

- [ ] **Step 3: Commit**

```bash
cd catalogizer-androidtv
git add app/build.gradle.kts app/src/main/java/com/catalogizer/androidtv/data/models/Settings.kt
git commit -m "fix: use emulator localhost for debug builds, BuildConfig for default URL"
```

---

### Task 1.3: Add Missing vitest.config.ts to 3 TS Submodules

**Files:**
- Create: `WebSocket-Client-TS/vitest.config.ts`
- Create: `Media-Types-TS/vitest.config.ts`
- Create: `Catalogizer-API-Client-TS/vitest.config.ts`

- [ ] **Step 1: Create vitest.config.ts for WebSocket-Client-TS**

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    include: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      exclude: ['node_modules', 'dist', '**/*.d.ts', '**/*.test.*'],
    },
  },
});
```

- [ ] **Step 2: Create vitest.config.ts for Media-Types-TS**

Same content as Step 1, but with `environment: 'node'` (no DOM needed for type definitions).

- [ ] **Step 3: Create vitest.config.ts for Catalogizer-API-Client-TS**

Same content as Step 1, but with `environment: 'node'`.

- [ ] **Step 4: Verify all 3 modules' tests still pass**

```bash
cd WebSocket-Client-TS && npm test
cd ../Media-Types-TS && npm test
cd ../Catalogizer-API-Client-TS && npm test
```

Expected: All PASS

- [ ] **Step 5: Commit each**

```bash
for dir in WebSocket-Client-TS Media-Types-TS Catalogizer-API-Client-TS; do
  cd "$dir" && git add vitest.config.ts && git commit -m "chore: add explicit vitest.config.ts for consistency" && cd ..
done
```

---

### Task 1.4: Integrate Lazy/Memory/Recovery Submodules into catalog-api

**Files:**
- Modify: `catalog-api/main.go`
- Create: `catalog-api/internal/lifecycle/lazy_services.go`
- Create: `catalog-api/internal/lifecycle/lazy_services_test.go`

- [ ] **Step 1: Write failing test for lazy service initialization**

Create `catalog-api/internal/lifecycle/lazy_services_test.go`:

```go
package lifecycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLazyServiceRegistry_Get_InitializesOnFirstAccess(t *testing.T) {
	callCount := 0
	registry := NewLazyServiceRegistry()
	registry.Register("test-service", func() (interface{}, error) {
		callCount++
		return "initialized", nil
	})

	// First access should initialize
	val, err := registry.Get("test-service")
	assert.NoError(t, err)
	assert.Equal(t, "initialized", val)
	assert.Equal(t, 1, callCount)

	// Second access should return cached
	val2, err := registry.Get("test-service")
	assert.NoError(t, err)
	assert.Equal(t, "initialized", val2)
	assert.Equal(t, 1, callCount) // NOT re-initialized
}

func TestLazyServiceRegistry_Get_UnknownService(t *testing.T) {
	registry := NewLazyServiceRegistry()
	_, err := registry.Get("nonexistent")
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd catalog-api && go test -v -run TestLazyServiceRegistry ./internal/lifecycle/`
Expected: FAIL — package does not exist

- [ ] **Step 3: Implement lazy service registry**

Create `catalog-api/internal/lifecycle/lazy_services.go`:

```go
package lifecycle

import (
	"fmt"
	"sync"

	lazypkg "digital.vasic.lazy/pkg/lazy"
)

// LazyServiceRegistry manages lazily-initialized services.
type LazyServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]*lazypkg.Lazy[interface{}]
}

func NewLazyServiceRegistry() *LazyServiceRegistry {
	return &LazyServiceRegistry{
		services: make(map[string]*lazypkg.Lazy[interface{}]),
	}
}

func (r *LazyServiceRegistry) Register(name string, init func() (interface{}, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = lazypkg.New(init)
}

func (r *LazyServiceRegistry) Get(name string) (interface{}, error) {
	r.mu.RLock()
	svc, ok := r.services[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("service %q not registered", name)
	}
	return svc.Get()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd catalog-api && go test -v -run TestLazyServiceRegistry ./internal/lifecycle/`
Expected: PASS

- [ ] **Step 5: Wire lazy initialization for heavyweight services in main.go**

In `catalog-api/main.go`, convert ProviderManager, MediaAnalyzer, and SMBDiscoveryService to lazy initialization. These are heavyweight and not needed until first use:

```go
import (
	"catalog-api/internal/lifecycle"
	recoverypkg "digital.vasic.recovery/pkg/circuitbreaker"
	memorypkg "digital.vasic.memory/pkg/tracker"
)

// After database setup, before handler creation:
lazyRegistry := lifecycle.NewLazyServiceRegistry()

// Register heavyweight services for lazy init
lazyRegistry.Register("provider-manager", func() (interface{}, error) {
	return providers.NewProviderManager(logger), nil
})

lazyRegistry.Register("smb-discovery", func() (interface{}, error) {
	return internal_services.NewSMBDiscoveryService(db, logger), nil
})

// Create memory tracker for development/test builds
memTracker := memorypkg.NewTracker(memorypkg.Config{
	Enabled:       gin.Mode() != gin.ReleaseMode,
	CheckInterval: 60,
})
defer memTracker.Stop()

// Create circuit breaker for external APIs
apiBreaker := recoverypkg.NewCircuitBreaker(recoverypkg.Config{
	MaxFailures: 5,
	Timeout:     30,
})
```

- [ ] **Step 6: Run all existing tests to verify no regressions**

Run: `cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1`
Expected: All packages PASS

- [ ] **Step 7: Commit**

```bash
cd catalog-api
git add internal/lifecycle/ main.go
git commit -m "feat: integrate Lazy/Memory/Recovery submodules, lazy-init heavyweight services"
```

---

### Task 1.5: Add Semaphore-Based Concurrency Control

**Files:**
- Create: `catalog-api/internal/concurrency/semaphore.go`
- Create: `catalog-api/internal/concurrency/semaphore_test.go`
- Modify: `catalog-api/internal/services/universal_scanner.go` (wrap scan operations)

- [ ] **Step 1: Write failing test for scan semaphore**

Create `catalog-api/internal/concurrency/semaphore_test.go`:

```go
package concurrency

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScanSemaphore_LimitsConcurrency(t *testing.T) {
	sem := NewSemaphore(2) // max 2 concurrent
	var running int32
	var maxRunning int32

	done := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func() {
			err := sem.Acquire(context.Background())
			assert.NoError(t, err)
			cur := atomic.AddInt32(&running, 1)
			for {
				old := atomic.LoadInt32(&maxRunning)
				if cur <= old || atomic.CompareAndSwapInt32(&maxRunning, old, cur) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&running, -1)
			sem.Release()
			done <- struct{}{}
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
	assert.LessOrEqual(t, atomic.LoadInt32(&maxRunning), int32(2))
}

func TestScanSemaphore_ContextCancellation(t *testing.T) {
	sem := NewSemaphore(1)
	_ = sem.Acquire(context.Background()) // Take the only slot

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := sem.Acquire(ctx)
	assert.Error(t, err) // Should fail — slot occupied, context expired
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd catalog-api && go test -v -run TestScanSemaphore ./internal/concurrency/`
Expected: FAIL — package does not exist

- [ ] **Step 3: Implement semaphore**

Create `catalog-api/internal/concurrency/semaphore.go`:

```go
package concurrency

import "context"

// Semaphore limits concurrent access to a resource.
type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(maxConcurrent int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, maxConcurrent)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) Release() {
	<-s.ch
}

func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd catalog-api && go test -v -run TestScanSemaphore ./internal/concurrency/`
Expected: PASS

- [ ] **Step 5: Apply semaphore to universal scanner**

In `catalog-api/internal/services/universal_scanner.go`, add a semaphore field and use it to limit concurrent scan operations:

```go
// In the struct definition, add:
scanSemaphore *concurrency.Semaphore

// In the constructor, add:
scanSemaphore: concurrency.NewSemaphore(3), // Max 3 concurrent scans

// In the scan method, wrap the operation:
if err := s.scanSemaphore.Acquire(ctx); err != nil {
    return fmt.Errorf("scan semaphore: %w", err)
}
defer s.scanSemaphore.Release()
```

- [ ] **Step 6: Run all tests to verify no regressions**

Run: `cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
cd catalog-api
git add internal/concurrency/semaphore.go internal/concurrency/semaphore_test.go internal/services/universal_scanner.go
git commit -m "feat: add semaphore-based concurrency control for scan operations"
```

---

### Task 1.6: Implement Free-API Metadata Providers (OpenLibrary, MusicBrainz, IGDB)

**Files:**
- Modify: `catalog-api/internal/media/providers/providers.go:458-645`
- Modify: `catalog-api/internal/media/providers/providers_test.go`

- [ ] **Step 1: Write failing tests for OpenLibrary provider**

In `catalog-api/internal/media/providers/providers_test.go`, add:

```go
func TestOpenLibraryProvider_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/search.json")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"docs": []map[string]interface{}{
				{
					"key":              "/works/OL123W",
					"title":            "Test Book",
					"first_publish_year": 2020,
					"author_name":      []string{"Test Author"},
					"cover_i":          12345,
				},
			},
		})
	}))
	defer server.Close()

	logger := testLogger()
	provider := &OpenLibraryProvider{
		NewBaseProvider("openlibrary", server.URL, "", &http.Client{}, logger),
	}

	results, err := provider.Search(context.Background(), "Test Book", "book", nil)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Test Book", results[0].Title)
	assert.Equal(t, 2020, results[0].Year)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd catalog-api && go test -v -run TestOpenLibraryProvider_Search ./internal/media/providers/`
Expected: FAIL — Search returns empty (stub)

- [ ] **Step 3: Implement OpenLibrary Search and GetDetails**

Replace the stub `Search()` and `GetDetails()` methods for OpenLibraryProvider in `providers.go`:

```go
func (p *OpenLibraryProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		return nil, nil
	}
	url := fmt.Sprintf("%s/search.json?q=%s&limit=10", p.baseURL, neturl.QueryEscape(query))
	if year != nil {
		url += fmt.Sprintf("&first_publish_year=%d", *year)
	}

	resp, err := p.makeRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("openlibrary search: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Docs []struct {
			Key              string   `json:"key"`
			Title            string   `json:"title"`
			FirstPublishYear int      `json:"first_publish_year"`
			AuthorName       []string `json:"author_name"`
			CoverI           int      `json:"cover_i"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("openlibrary decode: %w", err)
	}

	results := make([]SearchResult, 0, len(data.Docs))
	for _, doc := range data.Docs {
		coverURL := ""
		if doc.CoverI > 0 {
			coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
		}
		results = append(results, SearchResult{
			ExternalID: doc.Key,
			Title:      doc.Title,
			Year:       doc.FirstPublishYear,
			CoverURL:   coverURL,
			Provider:   p.GetName(),
			Relevance:  0.8,
		})
	}
	return results, nil
}

func (p *OpenLibraryProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		return nil, fmt.Errorf("openlibrary provider not enabled")
	}
	url := fmt.Sprintf("%s%s.json", p.baseURL, externalID)
	resp, err := p.makeRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("openlibrary details: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Title       string `json:"title"`
		Description interface{} `json:"description"`
		Covers      []int  `json:"covers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("openlibrary decode: %w", err)
	}

	description := ""
	switch v := data.Description.(type) {
	case string:
		description = v
	case map[string]interface{}:
		if val, ok := v["value"].(string); ok {
			description = val
		}
	}

	coverURL := ""
	if len(data.Covers) > 0 {
		coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", data.Covers[0])
	}

	return &models.ExternalMetadata{
		ExternalID:  externalID,
		Provider:    "openlibrary",
		Title:       data.Title,
		Description: description,
		CoverURL:    coverURL,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd catalog-api && go test -v -run TestOpenLibraryProvider ./internal/media/providers/`
Expected: PASS

- [ ] **Step 5: Implement MusicBrainz provider (free API, no key needed)**

Same pattern — implement `Search()` using `https://musicbrainz.org/ws/2/recording?query={query}&fmt=json` and `GetDetails()` using `/ws/2/recording/{id}?inc=artists+releases&fmt=json`. Write tests with httptest mock server first.

- [ ] **Step 6: Implement remaining providers with graceful degradation**

For providers requiring API keys (IMDB, TVDB, Spotify, LastFM, IGDB, Steam, etc.), implement the actual API integration but return empty results with a logged warning when no API key is configured:

```go
func (p *IMDBProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("IMDB provider not enabled (no API key)", zap.String("query", query))
		return nil, nil
	}
	// Full implementation...
}
```

- [ ] **Step 7: Run all provider tests**

Run: `cd catalog-api && go test -v ./internal/media/providers/ -count=1`
Expected: All PASS

- [ ] **Step 8: Commit**

```bash
cd catalog-api
git add internal/media/providers/
git commit -m "feat: implement metadata providers (OpenLibrary, MusicBrainz full; others with graceful degradation)"
```

---

### Task 1.7: Configure Redis Connection Pooling

**Files:**
- Modify: `catalog-api/main.go` (Redis client initialization, around line 439)

- [ ] **Step 1: Add explicit pool configuration**

```go
// OLD:
redisClient := redis.NewClient(&redis.Options{
	Addr: redisAddr,
})

// NEW:
redisClient := redis.NewClient(&redis.Options{
	Addr:         redisAddr,
	PoolSize:     10,
	MinIdleConns: 3,
	MaxRetries:   3,
	DialTimeout:  5 * time.Second,
	ReadTimeout:  3 * time.Second,
	WriteTimeout: 3 * time.Second,
	PoolTimeout:  4 * time.Second,
})
```

- [ ] **Step 2: Run tests to verify no regressions**

Run: `cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1`
Expected: All PASS

- [ ] **Step 3: Commit**

```bash
cd catalog-api && git add main.go && git commit -m "perf: configure explicit Redis connection pooling"
```

---

## Phase 2: Security Scanning & Remediation

**Objective:** Execute all configured security scanners, analyze findings, and remediate all critical/high issues.

### Task 2.1: Run SonarQube Scan

**Files:**
- Modify: `docker-compose.security.yml` (if needed)
- Create: `scripts/run-sonarqube-scan.sh`

- [ ] **Step 1: Create SonarQube scanner script**

Create `scripts/run-sonarqube-scan.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "=== Starting SonarQube Infrastructure ==="
podman-compose -f docker-compose.security.yml up -d sonarqube sonarqube-db

echo "=== Waiting for SonarQube to be healthy ==="
for i in $(seq 1 60); do
    if curl -sf http://localhost:9000/api/system/status | grep -q '"status":"UP"'; then
        echo "SonarQube is ready"
        break
    fi
    echo "Waiting... ($i/60)"
    sleep 5
done

echo "=== Generating Go coverage report ==="
cd catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -coverprofile=coverage.out -count=1 || true
go test ./... -json > test-results.json 2>/dev/null || true
cd ..

echo "=== Generating Frontend coverage report ==="
cd catalog-web
npm run test:coverage || true
cd ..

echo "=== Running SonarQube Scanner ==="
podman run --rm --network host \
    -v "$(pwd):/usr/src" \
    -w /usr/src \
    docker.io/sonarsource/sonar-scanner-cli:latest \
    -Dsonar.host.url=http://localhost:9000 \
    -Dsonar.login=admin \
    -Dsonar.password=admin

echo "=== SonarQube scan complete ==="
echo "View results at: http://localhost:9000/dashboard?id=catalogizer"
```

- [ ] **Step 2: Make executable and run**

```bash
chmod +x scripts/run-sonarqube-scan.sh
```

- [ ] **Step 3: Commit script**

```bash
git add scripts/run-sonarqube-scan.sh
git commit -m "feat: add SonarQube scanner execution script"
```

---

### Task 2.2: Add Semgrep to Security Scanning

**Files:**
- Modify: `docker-compose.security.yml`
- Modify: `scripts/security-scan.sh`

- [ ] **Step 1: Add Semgrep service to docker-compose.security.yml**

Add after the trivy-scanner service:

```yaml
  # Semgrep SAST Scanner
  semgrep-scanner:
    image: docker.io/semgrep/semgrep:latest
    container_name: catalogizer-semgrep
    volumes:
      - ..:/project:ro
      - ./reports:/reports
    working_dir: /project
    command: [
      "semgrep", "scan",
      "--config", "auto",
      "--json",
      "--output", "/reports/semgrep-results.json",
      "--exclude", "node_modules",
      "--exclude", "vendor",
      "--exclude", "dist",
      "--exclude", "build",
      "--exclude", "target",
      "--exclude", "releases",
      "--severity", "WARNING",
      "."
    ]
    profiles:
      - semgrep-scan
    networks:
      - security-testing-network
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

- [ ] **Step 2: Run Semgrep scan**

```bash
podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner
```

- [ ] **Step 3: Analyze and remediate findings**

Read `reports/semgrep-results.json`, categorize by severity, fix all HIGH and CRITICAL findings.

- [ ] **Step 4: Commit**

```bash
git add docker-compose.security.yml
git commit -m "feat: add Semgrep SAST scanner to security compose"
```

---

### Task 2.3: Run Snyk + Trivy Scans and Remediate

- [ ] **Step 1: Run Snyk comprehensive scan**

```bash
podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-cli
```

- [ ] **Step 2: Run Trivy filesystem scan**

```bash
podman-compose -f docker-compose.security.yml --profile trivy-scan run --rm trivy-scanner
```

- [ ] **Step 3: Analyze reports**

Read `reports/snyk-dependencies-results.json`, `reports/snyk-code-results.json`, and `reports/trivy-results.json`.

- [ ] **Step 4: Remediate all critical/high vulnerabilities**

For each finding: update dependency version, apply code fix, or add suppression with documented justification.

- [ ] **Step 5: Re-run scans to verify remediation**

- [ ] **Step 6: Generate consolidated security report**

Create `docs/security/SECURITY_SCAN_REPORT_2026-03-26.md` with scan dates, tool versions, findings, remediations, and remaining accepted risks.

- [ ] **Step 7: Commit**

```bash
git add docs/security/ reports/
git commit -m "security: remediate all critical/high findings from Snyk, Trivy, Semgrep scans"
```

---

## Phase 3: Memory Leak, Race Condition & Deadlock Safety

**Objective:** Comprehensive audit and hardening against concurrency issues.

### Task 3.1: Run Go Race Detector on All Packages

- [ ] **Step 1: Run race detector**

```bash
cd catalog-api && GOMAXPROCS=3 go test -race ./... -p 1 -parallel 1 -count=1 -short 2>&1 | tee race-report.txt
```

- [ ] **Step 2: Analyze race report**

Search for `WARNING: DATA RACE` in output. For each race:
- Identify the goroutines involved
- Add proper synchronization (mutex, atomic, or channel)

- [ ] **Step 3: Fix any races found**

Apply fixes with proper locking patterns. Always use `defer mu.Unlock()` after `mu.Lock()`.

- [ ] **Step 4: Re-run race detector to verify**

```bash
cd catalog-api && GOMAXPROCS=3 go test -race ./... -p 1 -parallel 1 -count=1 -short
```

Expected: 0 races

- [ ] **Step 5: Commit**

```bash
git commit -am "safety: fix all data races detected by Go race detector"
```

---

### Task 3.2: Run Race Detector on All Go Submodules

- [ ] **Step 1: Run race detector on each module**

```bash
for mod in Auth Cache Challenges Concurrency Config Containers Database Discovery Entities EventBus Filesystem Lazy Media Memory Middleware Observability RateLimiter Recovery Security Storage Streaming Watcher; do
    echo "=== Testing $mod ==="
    cd "$mod" && go test -race ./... -count=1 -short 2>&1 | tee "../reports/${mod}-race.txt" && cd ..
done
```

- [ ] **Step 2: Fix any races found per module**

- [ ] **Step 3: Commit fixes per module**

---

### Task 3.3: Add Goroutine Leak Detection Test

**Files:**
- Create: `catalog-api/tests/stress/goroutine_leak_test.go`

- [ ] **Step 1: Write goroutine leak detection test**

```go
package stress

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoGoroutineLeaks_AfterServiceLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping goroutine leak test in short mode")
	}

	// Baseline goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	// Simulate 100 service creation/destruction cycles
	for i := 0; i < 100; i++ {
		// Create and immediately close cache service, websocket handler, etc.
		// (Use test helpers to create lightweight instances)
	}

	// Allow goroutines to settle
	runtime.GC()
	time.Sleep(500 * time.Millisecond)

	current := runtime.NumGoroutine()
	leaked := current - baseline

	// Allow small variance (runtime goroutines)
	assert.LessOrEqual(t, leaked, 5,
		"goroutine leak detected: baseline=%d, current=%d, leaked=%d", baseline, current, leaked)
}
```

- [ ] **Step 2: Run the test**

Run: `cd catalog-api && go test -v -run TestNoGoroutineLeaks ./tests/stress/ -count=1`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd catalog-api && git add tests/stress/goroutine_leak_test.go
git commit -m "test: add goroutine leak detection stress test"
```

---

### Task 3.4: Add Memory Pressure Monitoring Test

**Files:**
- Modify: `catalog-api/tests/stress/memory_pressure_test.go`

- [ ] **Step 1: Enhance existing memory pressure test with heap growth tracking**

Add to the existing file:

```go
func TestMemoryStability_UnderSustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory stability test in short mode")
	}

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Simulate sustained operations
	for round := 0; round < 10; round++ {
		// Create temporary large allocations, then release
		data := make([]byte, 1<<20) // 1MB
		_ = data
	}

	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	runtime.ReadMemStats(&memAfter)

	heapGrowthMB := float64(memAfter.HeapAlloc-memBefore.HeapAlloc) / (1 << 20)
	t.Logf("Heap growth: %.2f MB (before: %d, after: %d)", heapGrowthMB,
		memBefore.HeapAlloc, memAfter.HeapAlloc)

	// Heap should not grow more than 10MB after GC
	assert.Less(t, heapGrowthMB, 10.0, "excessive heap growth indicates memory leak")
}
```

- [ ] **Step 2: Run test**

Run: `cd catalog-api && go test -v -run TestMemoryStability ./tests/stress/ -count=1`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd catalog-api && git add tests/stress/memory_pressure_test.go
git commit -m "test: enhance memory stability test with heap growth tracking"
```

---

## Phase 4: Performance Optimization

**Objective:** Implement lazy loading, connection pooling, and non-blocking patterns throughout.

### Task 4.1: Add HTTP Client Connection Pooling

**Files:**
- Modify: `catalog-api/main.go`
- Create: `catalog-api/internal/httpclient/pool.go`
- Create: `catalog-api/internal/httpclient/pool_test.go`

- [ ] **Step 1: Write test for pooled HTTP client**

```go
package httpclient

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPooledClient_HasCorrectTransportSettings(t *testing.T) {
	client := NewPooledClient()
	transport, ok := client.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.Equal(t, 100, transport.MaxIdleConns)
	assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
	assert.Equal(t, 30, int(transport.IdleConnTimeout.Seconds()))
}
```

- [ ] **Step 2: Implement pooled client**

```go
package httpclient

import (
	"net"
	"net/http"
	"time"
)

func NewPooledClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:  10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}
```

- [ ] **Step 3: Wire into main.go and ProviderManager**

Replace ad-hoc `http.Client` creation in ProviderManager with the pooled client.

- [ ] **Step 4: Run tests, commit**

---

### Task 4.2: Add k6 Spike Test Scenario

**Files:**
- Create: `tests/k6/spike_test.js`

- [ ] **Step 1: Create spike test**

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const apiLatency = new Trend('api_latency');

const API_URL = __ENV.API_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '10s', target: 5 },    // Normal load
    { duration: '5s', target: 200 },   // Spike!
    { duration: '30s', target: 200 },  // Stay at spike
    { duration: '5s', target: 5 },     // Recovery
    { duration: '30s', target: 5 },    // Verify recovery
    { duration: '5s', target: 0 },     // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<2000'], // Spikes tolerate higher latency
    'errors': ['rate<0.20'],             // Allow up to 20% errors during spike
  },
};

export default function () {
  const endpoints = [
    '/api/v1/health',
    '/api/v1/stats',
    '/api/v1/storage-roots',
  ];

  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const res = http.get(`${API_URL}${endpoint}`, { timeout: '10s' });

  apiLatency.add(res.timings.duration);
  errorRate.add(res.status >= 400);

  check(res, {
    'status is not 5xx': (r) => r.status < 500,
  });

  sleep(0.1);
}
```

- [ ] **Step 2: Commit**

```bash
git add tests/k6/spike_test.js
git commit -m "test: add k6 spike test scenario for sudden traffic surges"
```

---

## Phase 5: Test Coverage Expansion

**Objective:** Increase test coverage to theoretical maximum across all components.

### Task 5.1: Add challenge.go Handler Tests

**Files:**
- Create: `catalog-api/handlers/challenge_dedicated_test.go`

- [ ] **Step 1: Write comprehensive handler tests**

Test all challenge endpoints: list, get, run, run-all, get-status, get-results. Use httptest with gin engine.

- [ ] **Step 2: Verify they pass**

- [ ] **Step 3: Commit**

---

### Task 5.2: Expand ReportingService Private Method Coverage

**Files:**
- Modify: `catalog-api/services/reporting_service_test.go`

- [ ] **Step 1: Add tests for format methods**

Test `formatAsMarkdown`, `formatAsHTML`, and `formatAsPDF` through the public `GenerateReport` method with different format parameters.

- [ ] **Step 2: Add edge case tests**

Empty data, missing fields, very large reports, special characters in content.

- [ ] **Step 3: Verify, commit**

---

### Task 5.3: Expand AnalyticsService Private Method Coverage

**Files:**
- Modify: `catalog-api/services/analytics_service_test.go`

- [ ] **Step 1: Add tests that exercise private analysis helpers**

Through the public API: `GetUserAnalytics`, `GetSystemAnalytics`, `GetMediaAnalytics` — these internally call the private helper functions. Test with various data scenarios.

- [ ] **Step 2: Verify, commit**

---

### Task 5.4: Expand SyncService Private Method Coverage

**Files:**
- Modify: `catalog-api/services/sync_service_test.go`

- [ ] **Step 1: Test sync operations through StartSync**

Create test scenarios for each sync type (mirror, incremental, bidirectional) with mock filesystem operations.

- [ ] **Step 2: Verify, commit**

---

### Task 5.5: Add Integration Tests for Provider Pipeline

**Files:**
- Create: `catalog-api/tests/integration/provider_pipeline_test.go`

- [ ] **Step 1: Test full provider → analyzer → entity creation pipeline**

Use httptest servers to mock external APIs and verify the full flow from provider search to entity creation.

- [ ] **Step 2: Verify, commit**

---

### Task 5.6: Add Frontend Accessibility Tests

**Files:**
- Create: `catalog-web/e2e/tests/accessibility.spec.ts`

- [ ] **Step 1: Write a11y test using @axe-core/playwright**

```typescript
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Accessibility', () => {
  test('dashboard has no critical a11y violations', async ({ page }) => {
    await page.goto('/');
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze();
    expect(results.violations.filter(v => v.impact === 'critical')).toEqual([]);
  });
});
```

- [ ] **Step 2: Run and fix any violations**

- [ ] **Step 3: Commit**

---

### Task 5.7: Generate and Review Coverage Reports

- [ ] **Step 1: Generate Go coverage**

```bash
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -coverprofile=coverage.out -count=1
go tool cover -func=coverage.out | tail -1
```

- [ ] **Step 2: Generate Frontend coverage**

```bash
cd catalog-web && npm run test:coverage
```

- [ ] **Step 3: Identify any remaining gaps below 80%**

Review coverage output, write additional tests for any package/file below 80%.

- [ ] **Step 4: Commit all new tests**

---

## Phase 6: New Challenges

**Objective:** Add challenges for all new functionality added in Phases 1-5.

### Task 6.1: Add Provider Verification Challenges

**Files:**
- Create: `catalog-api/challenges/provider_verification.go`
- Modify: `catalog-api/challenges/register.go`

- [ ] **Step 1: Create challenges for provider functionality**

```go
// CH-051: Verify OpenLibrary provider returns results
// CH-052: Verify MusicBrainz provider returns results
// CH-053: Verify provider graceful degradation (no API key)
// CH-054: Verify provider circuit breaker activation
// CH-055: Verify lazy service initialization
```

- [ ] **Step 2: Register in register.go**

- [ ] **Step 3: Run challenges to verify**

- [ ] **Step 4: Commit**

---

### Task 6.2: Add Performance Verification Challenges

**Files:**
- Create: `catalog-api/challenges/performance_verification.go`
- Modify: `catalog-api/challenges/register.go`

- [ ] **Step 1: Create performance challenges**

```go
// CH-056: Verify scan semaphore limits concurrency
// CH-057: Verify Redis connection pooling
// CH-058: Verify HTTP client connection reuse
// CH-059: Verify no goroutine leaks after 100 operations
// CH-060: Verify API response time < 500ms under load
```

- [ ] **Step 2: Register, run, verify, commit**

---

## Phase 7: Documentation Completion

**Objective:** Complete all 11 incomplete docs, create MODULE 9-10 scripts, update all existing documentation.

### Task 7.1: Complete Incomplete Documentation Files

For each of the 11 incomplete files, update with current accurate content:

- [ ] **Step 1: Update docs/deployment/PRODUCTION_DEPLOYMENT_GUIDE.md**
  - Replace TODO for Prometheus metrics with actual `/metrics` endpoint documentation
  - Complete environment procedures with actual container commands

- [ ] **Step 2: Update docs/testing/TEST_IMPLEMENTATION_SUMMARY.md**
  - Remove Android TODO items — document current state (JDK 17 requirement, 166 test files)

- [ ] **Step 3: Update planning docs to reflect completed work**

For each planning document (docs/status/COMPREHENSIVE_STATUS_AND_IMPLEMENTATION_PLAN.md, docs/MASTER_AUDIT_AND_IMPLEMENTATION_PLAN.md, docs/plans/*.md, docs/COMPREHENSIVE_IMPLEMENTATION_PLAN.md, docs/UNFINISHED_WORK_COMPREHENSIVE_REPORT.md):
  - Add "SUPERSEDED" header with reference to this new plan
  - Or update status to reflect current completion state

- [ ] **Step 4: Update security scan report**

Regenerate `docs/security/SECURITY_SCAN_REPORT_2026-03-26.md` from Phase 2 results.

- [ ] **Step 5: Commit all doc updates**

```bash
git add docs/
git commit -m "docs: complete all incomplete documentation, update status reports"
```

---

### Task 7.2: Create Course Module 9 and 10 Scripts

**Files:**
- Create: `docs/courses/scripts/MODULE_9_ADVANCED_FEATURES.md`
- Create: `docs/courses/scripts/MODULE_10_TROUBLESHOOTING.md`

- [ ] **Step 1: Write MODULE_9 script**

Cover: AI analytics, smart collections, playlist automation, subtitle management, recommendation engine, media entity system. Follow format of existing MODULE_1-8 scripts (18-24KB each, instructor narration format).

- [ ] **Step 2: Write MODULE_10 script**

Cover: Troubleshooting common issues, debugging network protocols, database recovery, log analysis, performance profiling, challenge system for verification. Follow existing format.

- [ ] **Step 3: Commit**

```bash
git add docs/courses/scripts/
git commit -m "docs: add MODULE 9 and 10 course scripts (advanced features, troubleshooting)"
```

---

### Task 7.3: Update All Existing Documentation

- [ ] **Step 1: Update CLAUDE.md with new features**

Add sections for: lazy service registry, semaphore concurrency control, new providers, circuit breaker integration.

- [ ] **Step 2: Update AGENTS.md with new commands**

Add: SonarQube scan script, Semgrep scan command, spike test command.

- [ ] **Step 3: Update docs/architecture/ARCHITECTURE.md**

Add: lazy loading architecture diagram, circuit breaker flow, semaphore patterns.

- [ ] **Step 4: Update docs/api/API_DOCUMENTATION.md**

Verify all 65+ endpoints are documented with request/response examples.

- [ ] **Step 5: Update SQL migration docs**

Update `docs/architecture/SQL_MIGRATIONS.md` with any new migrations.

- [ ] **Step 6: Update each Go module's CLAUDE.md**

For modules that gained new functionality (Lazy, Memory, Recovery — now used in catalog-api), update their CLAUDE.md with usage examples.

- [ ] **Step 7: Commit**

```bash
git add CLAUDE.md AGENTS.md docs/
git commit -m "docs: comprehensive documentation update for all new features and patterns"
```

---

## Phase 8: User Manuals & Video Courses

**Objective:** Extend all user-facing documentation.

### Task 8.1: Update User Guides for Each Platform

- [ ] **Step 1: Update docs/guides/WEB_APP_GUIDE.md**

Add: new collection sharing (with correct non-localhost URLs), accessibility features, provider-enriched media details.

- [ ] **Step 2: Update docs/guides/ANDROID_TV_GUIDE.md**

Add: server discovery configuration (debug vs release), media browsing with enriched metadata.

- [ ] **Step 3: Update docs/guides/DESKTOP_GUIDE.md and INSTALLER_WIZARD_GUIDE.md**

Verify accuracy, add any new features.

- [ ] **Step 4: Update docs/tutorials/***

Add tutorial for: configuring metadata providers, running security scans, interpreting challenge results.

- [ ] **Step 5: Commit**

```bash
git add docs/guides/ docs/tutorials/
git commit -m "docs: update all user guides and tutorials with new features"
```

---

### Task 8.2: Update Course Slides

- [ ] **Step 1: Update MODULE_9_SLIDES.md and MODULE_10_SLIDES.md**

Align with new full scripts created in Task 7.2.

- [ ] **Step 2: Review and update MODULE_1-8 slides for accuracy**

Check each slide deck against current feature set. Update screenshots, commands, and feature descriptions.

- [ ] **Step 3: Commit**

```bash
git add docs/courses/slides/
git commit -m "docs: update all course slide decks to reflect current features"
```

---

## Phase 9: Website Content Update

**Objective:** Update all website pages to reflect current project state.

### Task 9.1: Update Website Content

**Files:**
- Modify: `Website/features.md`
- Modify: `Website/changelog.md`
- Modify: `Website/faq.md`
- Modify: `Website/documentation.md`
- Modify: `Website/course.md`
- Modify: `Website/download.md`

- [ ] **Step 1: Update features.md**

Add: metadata provider integration (13 providers), lazy loading, circuit breaker resilience, semaphore concurrency, enhanced security scanning (SonarQube, Semgrep, Snyk, Trivy).

- [ ] **Step 2: Update changelog.md**

Add entry for this release with all improvements categorized.

- [ ] **Step 3: Update faq.md**

Add FAQs for: configuring metadata providers, running security scans, container resource limits, HTTP/3 requirements.

- [ ] **Step 4: Update documentation.md**

Ensure all doc links are valid and point to current documents.

- [ ] **Step 5: Update course.md**

Add MODULE 9 and 10 descriptions.

- [ ] **Step 6: Update download.md**

Verify download instructions and version numbers.

- [ ] **Step 7: Commit**

```bash
cd Website && git add . && git commit -m "content: update website for project completion release"
```

---

## Phase 10: Final Validation & Comprehensive Testing

**Objective:** Run every test type, verify zero failures, generate final reports.

### Task 10.1: Run Complete Test Suite

- [ ] **Step 1: Go unit tests with race detector**

```bash
cd catalog-api && GOMAXPROCS=3 go test -race ./... -p 2 -parallel 2 -count=1
```

- [ ] **Step 2: Go submodule tests**

```bash
for mod in Auth Cache Challenges Concurrency Config Containers Database Discovery Entities EventBus Filesystem Lazy Media Memory Middleware Observability RateLimiter Recovery Security Storage Streaming Watcher; do
    echo "=== $mod ===" && cd "$mod" && go test ./... -count=1 -short && cd ..
done
```

- [ ] **Step 3: Frontend unit tests**

```bash
cd catalog-web && npm run test
```

- [ ] **Step 4: Frontend type check + lint**

```bash
cd catalog-web && npm run type-check && npm run lint
```

- [ ] **Step 5: TS submodule tests**

```bash
for dir in WebSocket-Client-TS UI-Components-React Media-Types-TS Catalogizer-API-Client-TS Auth-Context-React Media-Browser-React Media-Player-React Collection-Manager-React Dashboard-Analytics-React; do
    echo "=== $dir ===" && cd "$dir" && npm test && cd ..
done
```

- [ ] **Step 6: Desktop + Installer tests**

```bash
cd catalogizer-desktop && npm test && cd ..
cd installer-wizard && npm test && cd ..
```

Expected: ALL PASS with 0 failures.

---

### Task 10.2: Run Security Scans

- [ ] **Step 1: govulncheck**

```bash
cd catalog-api && govulncheck ./...
```

- [ ] **Step 2: npm audit**

```bash
cd catalog-web && npm audit --production
```

- [ ] **Step 3: Snyk scan**

```bash
podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-cli
```

Expected: 0 critical vulnerabilities

---

### Task 10.3: Run Load Tests

- [ ] **Step 1: k6 load test**

```bash
podman run --rm --network host -v $(pwd)/tests/k6:/scripts docker.io/grafana/k6:latest run /scripts/load_test.js
```

Expected: p95 < 500ms, error rate < 5%

- [ ] **Step 2: k6 stress test**

```bash
podman run --rm --network host -v $(pwd)/tests/k6:/scripts docker.io/grafana/k6:latest run /scripts/stress_test.js
```

Expected: p99 < 2s, error rate < 15%

- [ ] **Step 3: k6 spike test**

```bash
podman run --rm --network host -v $(pwd)/tests/k6:/scripts docker.io/grafana/k6:latest run /scripts/spike_test.js
```

Expected: Recovery within 30s, p95 < 2s

- [ ] **Step 4: k6 soak test**

```bash
podman run --rm --network host -v $(pwd)/tests/k6:/scripts docker.io/grafana/k6:latest run /scripts/soak_test.js
```

Expected: p95 < 500ms after 30 minutes, no memory leak

---

### Task 10.4: Run Challenge Suite

- [ ] **Step 1: Start services in containers**

```bash
podman-compose -f docker-compose.dev.yml up -d
```

- [ ] **Step 2: Run all challenges**

Via the running API: trigger RunAll through the challenge endpoint.

- [ ] **Step 3: Verify all 239+ challenges pass**

Expected: All original (CH-001 to CH-050), userflow (UF-*), module (MOD-*), and new (CH-051 to CH-060) challenges PASS.

- [ ] **Step 4: Stop services**

```bash
podman-compose -f docker-compose.dev.yml down
```

---

### Task 10.5: Generate Final Completion Report

**Files:**
- Create: `docs/status/FINAL_COMPLETION_REPORT_2026-03-26.md`

- [ ] **Step 1: Write comprehensive report**

Include:
- Summary of all work completed per phase
- Test results: unit counts, coverage percentages, challenge pass rates
- Security scan results: vulnerabilities found/fixed
- Performance metrics: k6 results, response times
- Documentation inventory: all docs created/updated
- Known limitations and accepted risks

- [ ] **Step 2: Commit final report**

```bash
git add docs/status/FINAL_COMPLETION_REPORT_2026-03-26.md
git commit -m "docs: final project completion report with all verification results"
```

---

### Task 10.6: Push to All Remotes

- [ ] **Step 1: Push main repo**

```bash
ssh-keyscan github.com gitlab.com gitflic.ru >> ~/.ssh/known_hosts 2>/dev/null
ssh-keyscan -p 2222 gitverse.ru >> ~/.ssh/known_hosts 2>/dev/null
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

- [ ] **Step 2: Push updated submodules**

Push each submodule that was modified to its remotes.

---

## Phase Summary Table

| Phase | Objective | Tasks | Estimated Steps |
|-------|-----------|-------|-----------------|
| 1 | Code Fixes (stubs, placeholders, lazy loading, semaphores) | 7 | 45 |
| 2 | Security Scanning & Remediation | 3 | 18 |
| 3 | Memory/Race/Deadlock Safety | 4 | 16 |
| 4 | Performance Optimization | 2 | 10 |
| 5 | Test Coverage Expansion | 7 | 25 |
| 6 | New Challenges | 2 | 8 |
| 7 | Documentation Completion | 3 | 18 |
| 8 | User Manuals & Video Courses | 2 | 8 |
| 9 | Website Content Update | 1 | 7 |
| 10 | Final Validation | 6 | 20 |
| **TOTAL** | | **37 tasks** | **175 steps** |

## Verification Gates

After each phase, run:

```bash
# Gate check: all existing tests still pass
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1
cd catalog-web && npm run test && npm run type-check
```

If any test fails, stop and fix before proceeding to the next phase.

# Master Completion Plan - Full Project Audit & Implementation

> **STATUS: COMPLETE** -- All 11 phases executed. See `docs/status/FINAL_COMPLETION_REPORT_2026-03-30.md` for details.

**Date**: 2026-03-30
**Scope**: Complete audit of all unfinished work + phased implementation plan to 100% completion
**Supersedes**: All prior remediation and completion plans

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Full Audit Report](#2-full-audit-report)
3. [Phase 1: Dead Code Elimination & Stub Implementation](#phase-1)
4. [Phase 2: Wire Unused Modules](#phase-2)
5. [Phase 3: Concurrency Safety & Memory Leak Fixes](#phase-3)
6. [Phase 4: Lazy Loading, Semaphores & Non-Blocking Patterns](#phase-4)
7. [Phase 5: Security Scanning & Remediation](#phase-5)
8. [Phase 6: Test Coverage to Theoretical Maximum](#phase-6)
9. [Phase 7: Stress, Integration & Monitoring Tests](#phase-7)
10. [Phase 8: Challenge Bank Expansion](#phase-8)
11. [Phase 9: Documentation Completion](#phase-9)
12. [Phase 10: Video Course & Website Update](#phase-10)
13. [Phase 11: Final Verification & Release](#phase-11)
14. [Constraints & Rules](#constraints)
15. [Risk Matrix](#risk-matrix)

---

## 1. Executive Summary

### Current State (2026-03-30)

| Metric | Value | Target |
|--------|-------|--------|
| Go backend test coverage | 65.1% | 95%+ |
| Frontend (catalog-web) test coverage | ~46% | 90%+ |
| Desktop app test coverage | ~3.8% | 80%+ |
| Installer wizard test coverage | ~9.4% | 80%+ |
| Android app test coverage | ~4.8% | 80%+ |
| Android TV test coverage | ~3.2% | 80%+ |
| API client test coverage | ~0.6% | 90%+ |
| Registered challenges | 249 | 350+ |
| Stub/placeholder endpoints | 10 | 0 |
| Unused Go modules (replace directives) | 12 | 0 |
| Unused TypeScript packages (file://) | 6 | 0 |
| Disabled/commented-out middleware | 3 | 0 |
| Concurrency safety issues | 8 critical | 0 |
| React memory leak patterns | 6 | 0 |
| Missing CLAUDE.md in main components | 4 | 0 |
| Missing AGENTS.md in submodules | 8 | 0 |
| Missing docs/ in main components | 4 | 0 |

### What This Plan Delivers

- **Zero dead code** -- every function, handler, service, and module is wired and operational
- **Zero stubs** -- all placeholder endpoints fully implemented or removed
- **Zero concurrency issues** -- all goroutine leaks, race conditions, deadlocks fixed
- **Zero memory leaks** -- all unbounded maps, unclosed resources, React leaks fixed
- **95%+ test coverage** across Go backend; 80-90%+ across all other components
- **350+ challenges** covering all test types and all platforms
- **Complete documentation** -- every module, every API, every workflow documented
- **Updated video course** -- all 12 modules current with new features
- **Updated website** -- all pages reflecting current state
- **Security-hardened** -- Snyk, SonarQube, Semgrep all passing with zero critical findings

---

## 2. Full Audit Report

### 2.1 Dead Code & Stub Implementations

#### 2.1.1 Stub Handler Endpoints (catalog-api)

| Endpoint | Handler | File | Status |
|----------|---------|------|--------|
| `GET /api/v1/media/recent` | `GetRecentMedia()` | `handlers/stub_handler.go:21` | Returns empty `[]` |
| `GET /api/v1/media/popular` | `GetPopularMedia()` | `handlers/stub_handler.go:30` | Returns empty `[]` |
| `GET /api/v1/media/by-path` | `GetMediaByPath()` | `handlers/stub_handler.go:39` | Returns empty `[]` |
| `POST /api/v1/media/analyze` | `AnalyzeMedia()` | `handlers/stub_handler.go:66` | Returns "not yet implemented" |
| `POST /api/v1/media/:id/refresh` | `RefreshMediaMetadata()` | `handlers/stub_handler.go:48` | Returns "not yet implemented" |
| `GET /api/v1/media/:id/quality` | `GetMediaQuality()` | `handlers/stub_handler.go:57` | Returns "not yet implemented" |
| `POST /api/v1/auth/change-password` | `ChangePassword()` | `handlers/stub_handler.go:75` | Returns "not yet implemented" |

#### 2.1.2 Admin Handler Stubs

| Endpoint | Handler | File | Status |
|----------|---------|------|--------|
| `GET /api/v1/admin/backups` | `GetBackups()` | `handlers/admin_handler.go:300` | Returns empty list |
| `POST /api/v1/admin/backups` | `CreateBackup()` | `handlers/admin_handler.go:312` | Acknowledges but no-ops |
| `POST /api/v1/admin/backups/:id/restore` | `RestoreBackup()` | `handlers/admin_handler.go:334` | Acknowledges but no-ops |
| `POST /api/v1/admin/storage/scan` | `ScanStorage()` | `handlers/admin_handler.go:350` | Returns "pending" |

#### 2.1.3 Commented-Out Middleware (main.go:685-687)

```go
// router.Use(root_middleware.AdvancedRateLimit(...))
// router.Use(root_middleware.UserBasedRateLimit(...))
// router.Use(root_middleware.IPRateLimit(10, 20))
```

#### 2.1.4 Unused Go Modules (12 replace directives, never imported)

| Module | Path | Action Required |
|--------|------|-----------------|
| `digital.vasic.database` | `../Database` | Wire or remove |
| `digital.vasic.lazy` | `../Lazy` | Wire or remove |
| `digital.vasic.media` | `../Media` | Wire or remove |
| `digital.vasic.memory` | `../Memory` | Wire or remove |
| `digital.vasic.middleware` | `../Middleware` | Wire or remove |
| `digital.vasic.observability` | `../Observability` | Wire or remove |
| `digital.vasic.ratelimiter` | `../RateLimiter` | Wire or remove |
| `digital.vasic.recovery` | `../Recovery` | Wire or remove |
| `digital.vasic.security` | `../Security` | Wire or remove |
| `digital.vasic.storage` | `../Storage` | Wire or remove |
| `digital.vasic.streaming` | `../Streaming` | Wire or remove |
| `digital.vasic.watcher` | `../Watcher` | Wire or remove |

#### 2.1.5 Unused TypeScript Packages (6 local, never imported)

| Package | Location | Used? |
|---------|----------|-------|
| `@vasic-digital/auth-context` | `../Auth-Context-React` | No |
| `@vasic-digital/catalogizer-api-client` | `../Catalogizer-API-Client-TS` | No |
| `@vasic-digital/collection-manager` | `../Collection-Manager-React` | No |
| `@vasic-digital/dashboard-analytics` | `../Dashboard-Analytics-React` | No |
| `@vasic-digital/media-browser` | `../Media-Browser-React` | No |
| `@vasic-digital/media-player` | `../Media-Player-React` | No |

#### 2.1.6 Lazy Service Registry (Dead Reference)

- `main.go:336-337`: `lazyServices` initialized but assigned to `_` (unused)

### 2.2 Concurrency & Memory Safety Issues

#### 2.2.1 Critical: Unbounded Goroutines Without Cleanup

| File | Line | Issue | Severity |
|------|------|-------|----------|
| `services/sync_service.go` | 213 | `go s.performSync()` -- no context cancellation | CRITICAL |
| `services/error_reporting_service.go` | 112,116,149,153 | 4x fire-and-forget goroutines | CRITICAL |
| `services/log_management_service.go` | 127,316 | Streaming goroutines without lifecycle | CRITICAL |

#### 2.2.2 High: Cleanup Goroutines Without Cancellation

| File | Line | Issue | Severity |
|------|------|-------|----------|
| `middleware/request.go` | 39-52 | Rate limiter cleanup runs forever | HIGH |
| `middleware/enhanced_rate_limiter.go` | 159-180 | Cleanup only on explicit Stop() | HIGH |
| `internal/smb/resilience.go` | 225-228 | Unbounded retry goroutines | HIGH |
| `internal/media/realtime/enhanced_watcher.go` | 82,135 | Unbounded worker/monitor goroutines | HIGH |

#### 2.2.3 Memory Leak Patterns

| File | Line | Issue | Severity |
|------|------|-------|----------|
| `internal/smb/resilience.go` | 163 | Bounded event channel, no drain on Stop() | HIGH |
| `middleware/request.go` | 36 | IP bucket map grows with rotating IPs | MEDIUM |
| `middleware/enhanced_rate_limiter.go` | 142-143 | Limiter maps grow without cleanup failure handling | MEDIUM |
| `services/log_management_service.go` | 132-150 | Unbounded slice append in log collection | MEDIUM |
| `Assets/pkg/event/bus.go` | 36-48 | Event handlers stored indefinitely | MEDIUM |

#### 2.2.4 React/TypeScript Memory Leaks

| File | Line | Issue | Severity |
|------|------|-------|----------|
| `components/collections/CollectionRealTime.tsx` | 266-272 | No AbortController for fetch | HIGH |
| `components/collections/CollectionRealTime.tsx` | 311-316 | Notification/Audio objects not cleaned | MEDIUM |
| `components/collections/PerformanceOptimizer.tsx` | 55-79 | IntersectionObserver recreated on every loadedItems change | MEDIUM |
| `components/collections/PerformanceOptimizer.tsx` | 26-27 | Map/Set state grows unbounded | MEDIUM |

### 2.3 Test Coverage Gaps

#### 2.3.1 Go Source Files Without Tests (94 files)

- **challenges/**: 67 challenge implementation files with no unit tests
- **database/**: 6 migration files untested (migrations_postgres.go, migrations_sqlite.go, v9, v10, v11, v13)
- **handlers/**: 3 files (admin_handler.go, playlist_handler.go, stub_handler.go)
- **repository/**: 1 file (playlist_repository.go)
- **services/**: 1 file (playlist_service.go)
- **database/**: 1 file (tx_helpers.go)

#### 2.3.2 TypeScript Source Files Without Tests

- **catalog-web/src/**: ~125 files untested (58 components, 8 hooks, 13 pages, 2 contexts, 22 utils)
- **catalogizer-desktop/**: ~1,584 files untested (3.8% coverage)
- **installer-wizard/**: ~2,007 files untested (9.4% coverage)
- **Catalogizer-API-Client-TS/**: ~508 files untested (0.6% coverage)

#### 2.3.3 Android/Kotlin Test Gaps

- **catalogizer-android/**: ~894 files untested (4.8% coverage)
- **catalogizer-androidtv/**: ~823 files untested (3.2% coverage)

#### 2.3.4 Missing Test Types

| Test Type | Go Backend | Frontend | Desktop | Android | Status |
|-----------|-----------|----------|---------|---------|--------|
| Unit | Good | Partial | Minimal | Minimal | Expand |
| Integration | Good | Minimal | Missing | Missing | Create |
| E2E | N/A | Good (26 Playwright) | Missing | Missing | Create |
| Stress | 54 (skipped in short) | Missing | Missing | Missing | Enable + Create |
| Load (k6) | 4 tests | Missing | N/A | N/A | Expand |
| Benchmark | 33 tests | Missing | N/A | N/A | Expand |
| Race condition | 77 tests | N/A | N/A | N/A | Maintain |
| Fuzz | 8 tests | Missing | N/A | N/A | Expand |
| Security | Partial | Missing | Missing | Missing | Create |
| Visual regression | N/A | Missing | Missing | N/A | Create |
| Accessibility | N/A | 1 test | Missing | N/A | Expand |
| Contract | Partial | Missing | N/A | N/A | Create |

### 2.4 Documentation Gaps

| Component | CLAUDE.md | AGENTS.md | docs/ | README |
|-----------|-----------|-----------|-------|--------|
| catalog-api | MISSING | Present | Partial (5 files) | Present |
| catalog-web | MISSING | Present | MISSING | Present |
| catalogizer-android | MISSING | Present | MISSING | Present |
| catalogizer-androidtv | MISSING | Present | MISSING | Present |
| DocProcessor | Present | Present | MISSING | Present |
| LLMOrchestrator | Present | Present | MISSING | Present |
| VisionEngine | Present | Present | MISSING | Present |
| LLMProvider | Present | Present | MISSING | MISSING |

**8 submodules missing AGENTS.md**: Entities, Media-Types-TS, Catalogizer-API-Client-TS, Auth-Context-React, Media-Browser-React, Dashboard-Analytics-React, Media-Player-React, Collection-Manager-React

### 2.5 Infrastructure Gaps

| Area | Status | Gap |
|------|--------|-----|
| AlertManager | Email only | Missing Slack/webhook |
| Semgrep rules | Auto only | No custom org rules |
| Distributed tracing | Missing | No OpenTelemetry |
| Log aggregation | Missing | No Loki |

---

## Phase 1: Dead Code Elimination & Stub Implementation {#phase-1}

**Duration**: 2-3 sessions | **Risk**: LOW | **Dependencies**: None

### Step 1.1: Implement All Stub Endpoints

Replace every stub in `handlers/stub_handler.go` with real implementations:

1. **`GetRecentMedia()`** -- Query `media_items` ordered by `updated_at DESC`, limit 20. Join `media_files` for file info. Return paginated JSON.
2. **`GetPopularMedia()`** -- Query `media_items` ordered by view count or favorite count. Use `favorites` table join. Return paginated JSON.
3. **`GetMediaByPath()`** -- Accept `?path=` query param. Query `files` table with path prefix match. Return matching media items via `media_files` join.
4. **`RefreshMediaMetadata()`** -- Accept media item ID. Trigger re-aggregation via `AggregationService`. Queue metadata provider refresh for the entity. Return 202 Accepted with job ID.
5. **`GetMediaQuality()`** -- Accept media item ID. Query file metadata (codec, bitrate, resolution from `files` table). Return quality analysis JSON.
6. **`AnalyzeMedia()`** -- Accept file path or media item ID. Run detection pipeline (`detector` -> `analyzer` -> `providers`). Return analysis results.
7. **`ChangePassword()`** -- Accept old password + new password. Validate old password via `AuthService`. Hash new password with bcrypt. Update `users` table. Invalidate existing JWT tokens.

**Files to modify**:
- `catalog-api/handlers/stub_handler.go` -- Replace all 7 stubs
- `catalog-api/main.go` -- Update route registrations if handler signatures change
- `catalog-api/services/` -- Add any new service methods needed

### Step 1.2: Implement Admin Backup Endpoints

1. **`GetBackups()`** -- List backup files from configured backup directory. Return metadata (filename, size, date, type).
2. **`CreateBackup()`** -- Export SQLite database file (copy) or PostgreSQL `pg_dump`. Store in backup directory with timestamp. Return 202 with backup ID.
3. **`RestoreBackup()`** -- Validate backup file integrity. For SQLite: replace DB file. For PostgreSQL: `pg_restore`. Return 202 with restore job ID.
4. **`ScanStorage()`** -- Delegate to `UniversalScanner.StartScan()`. Return scan job status with progress.

**Files to modify**:
- `catalog-api/handlers/admin_handler.go` -- Replace 4 stubs
- `catalog-api/services/backup_service.go` -- **NEW**: Backup/restore logic
- `catalog-api/repository/backup_repository.go` -- **NEW**: Backup metadata storage

### Step 1.3: Enable Commented-Out Middleware

1. Uncomment the 3 rate limiting middleware lines in `main.go:685-687`
2. Verify `AdvancedRateLimit`, `UserBasedRateLimit`, `IPRateLimit` are properly configured
3. Test that all existing endpoints still work with rate limiting enabled
4. Add appropriate rate limit bypass for health/metrics endpoints

**Files to modify**:
- `catalog-api/main.go:685-687` -- Uncomment middleware

### Step 1.4: Remove LazyServiceRegistry Dead Reference

1. Either wire `lazyServices` to actual service initialization or remove the unused variable
2. If LazyServiceRegistry provides value, integrate it into the service startup sequence

**Files to modify**:
- `catalog-api/main.go:336-337` -- Wire or remove

### Step 1.5: Tests for Phase 1

- Unit tests for all 7 former stub handlers (table-driven, test happy path + error cases)
- Unit tests for backup service (mock filesystem, mock DB)
- Integration test for rate limiting middleware chain
- Update existing Playwright E2E tests to cover new endpoints

**New test files**:
- `catalog-api/handlers/stub_handler_test.go`
- `catalog-api/handlers/admin_handler_backup_test.go`
- `catalog-api/services/backup_service_test.go`
- `catalog-api/middleware/rate_limit_integration_test.go`

### Step 1.6: Challenge Coverage for Phase 1

Create challenges verifying the formerly-stubbed endpoints work end-to-end:

| Challenge ID | Name | Validates |
|-------------|------|-----------|
| CH-141 | Recent Media API | GET /api/v1/media/recent returns real data |
| CH-142 | Popular Media API | GET /api/v1/media/popular returns sorted results |
| CH-143 | Media By Path API | GET /api/v1/media/by-path filters correctly |
| CH-144 | Media Analysis | POST /api/v1/media/analyze returns analysis |
| CH-145 | Metadata Refresh | POST /api/v1/media/:id/refresh triggers re-aggregation |
| CH-146 | Media Quality | GET /api/v1/media/:id/quality returns codec/bitrate info |
| CH-147 | Change Password | POST /api/v1/auth/change-password works with valid credentials |
| CH-148 | Backup Create | POST /api/v1/admin/backups creates backup file |
| CH-149 | Backup List | GET /api/v1/admin/backups lists available backups |
| CH-150 | Backup Restore | POST /api/v1/admin/backups/:id/restore restores DB |
| CH-151 | Storage Scan Admin | POST /api/v1/admin/storage/scan triggers scan |
| CH-152 | Rate Limiting | Verify rate limits enforce on repeated requests |

---

## Phase 2: Wire Unused Modules {#phase-2}

**Duration**: 3-4 sessions | **Risk**: MEDIUM | **Dependencies**: Phase 1

### Step 2.1: Audit Each Module's API Surface

For each of the 12 unused modules, determine:
1. What functionality does it provide?
2. Where in catalog-api should it be integrated?
3. Does it duplicate existing internal packages?

### Step 2.2: Wire Modules Into catalog-api

| Module | Integration Point | What It Replaces/Augments |
|--------|-------------------|--------------------------|
| `digital.vasic.database` | `database/` package | Augments dialect abstraction, adds connection pooling utilities |
| `digital.vasic.lazy` | `internal/lifecycle/` | Replaces internal lazy loading; use for all service initialization |
| `digital.vasic.media` | `internal/media/` | Augments media detection pipeline with shared types |
| `digital.vasic.memory` | `internal/monitoring/` | Adds memory leak detection, heap profiling |
| `digital.vasic.middleware` | `middleware/` | Provides reusable middleware (logging, tracing, recovery) |
| `digital.vasic.observability` | `internal/metrics/` | Augments Prometheus metrics with structured logging, tracing |
| `digital.vasic.ratelimiter` | `middleware/` | Replaces internal rate limiter with module's implementation |
| `digital.vasic.recovery` | `internal/recovery/` | Provides circuit breaker, retry with exponential backoff |
| `digital.vasic.security` | `middleware/`, `internal/auth/` | Adds security headers, CORS, input validation utilities |
| `digital.vasic.storage` | `internal/services/` | Provides unified storage abstraction |
| `digital.vasic.streaming` | `internal/media/realtime/` | Provides streaming utilities for media playback |
| `digital.vasic.watcher` | `internal/media/realtime/` | Replaces internal file watcher with module's implementation |

### Step 2.3: Wire TypeScript Packages Into catalog-web

| Package | Integration Point | What Changes |
|---------|-------------------|-------------|
| `@vasic-digital/auth-context` | `src/contexts/AuthContext.tsx` | Replace internal auth context with shared package |
| `@vasic-digital/catalogizer-api-client` | `src/lib/api.ts` | Replace internal API client calls with shared client |
| `@vasic-digital/collection-manager` | `src/components/collections/` | Replace internal collection components with shared ones |
| `@vasic-digital/dashboard-analytics` | `src/components/dashboard/` | Replace internal dashboard components |
| `@vasic-digital/media-browser` | `src/components/media/` | Replace internal media browser components |
| `@vasic-digital/media-player` | `src/components/media/` | Replace internal media player components |

### Step 2.4: Tests for Phase 2

- Integration tests verifying each module works correctly when wired
- Regression tests ensuring no existing functionality breaks
- Module-level tests for any new adapter code

### Step 2.5: Challenge Coverage for Phase 2

| Challenge ID | Name | Validates |
|-------------|------|-----------|
| CH-153 | Module Integration - Database | Database module wired and functional |
| CH-154 | Module Integration - Lazy | Lazy loading module replaces internal |
| CH-155 | Module Integration - Memory | Memory leak detection active |
| CH-156 | Module Integration - Middleware | Shared middleware active |
| CH-157 | Module Integration - Observability | Tracing and structured logging |
| CH-158 | Module Integration - RateLimiter | Rate limiter module active |
| CH-159 | Module Integration - Recovery | Circuit breaker active |
| CH-160 | Module Integration - Security | Security module headers active |
| CH-161 | Module Integration - Storage | Storage abstraction wired |
| CH-162 | Module Integration - Streaming | Streaming utilities active |
| CH-163 | Module Integration - Watcher | File watcher module active |
| CH-164 | Module Integration - Media | Media types shared |

---

## Phase 3: Concurrency Safety & Memory Leak Fixes {#phase-3}

**Duration**: 2-3 sessions | **Risk**: HIGH (touching concurrent code) | **Dependencies**: Phase 2

### Step 3.1: Fix Critical Goroutine Leaks

#### 3.1.1: SyncService (services/sync_service.go:213)

```go
// BEFORE: go s.performSync(session, endpoint)
// AFTER:
ctx, cancel := context.WithTimeout(parentCtx, s.syncTimeout)
defer cancel()
s.wg.Add(1)
go func() {
    defer s.wg.Done()
    s.performSync(ctx, session, endpoint)
}()
```

Add `Stop()` method with `s.wg.Wait()` and context cancellation.

#### 3.1.2: ErrorReportingService (services/error_reporting_service.go:112-153)

Add WaitGroup tracking for all 4 fire-and-forget goroutines. Add `Close()` method:
```go
func (s *ErrorReportingService) Close() {
    s.cancel()
    s.wg.Wait()
}
```

#### 3.1.3: LogManagementService (services/log_management_service.go:127,316)

Add context propagation and WaitGroup tracking. Ensure streaming goroutines respect context cancellation.

### Step 3.2: Fix Cleanup Goroutine Leaks

#### 3.2.1: Rate Limiter Cleanup (middleware/request.go:39-52)

Add stop channel:
```go
type ipRateLimiter struct {
    // ...existing fields
    stopChan chan struct{}
}

// In cleanup goroutine:
select {
case <-ticker.C:
    // cleanup logic
case <-rl.stopChan:
    return
}
```

#### 3.2.2: Enhanced Rate Limiter (middleware/enhanced_rate_limiter.go)

Ensure `Stop()` is called in all code paths (add to main.go shutdown sequence).

#### 3.2.3: SMB Resilience (internal/smb/resilience.go:225-228)

Implement goroutine pool with semaphore limiting max concurrent reconnections:
```go
reconSem := semaphore.NewWeighted(int64(maxConcurrentReconnections))
```

#### 3.2.4: Enhanced Watcher (internal/media/realtime/enhanced_watcher.go:82,135)

Implement bounded worker pool. Use semaphore to limit concurrent workers.

### Step 3.3: Fix Memory Leak Patterns

#### 3.3.1: Event Channel Drain (internal/smb/resilience.go:163)

Add drain mechanism in `Stop()`:
```go
func (m *ResilientSMBManager) Stop() {
    close(m.stopChan)
    // Drain event channel
    for {
        select {
        case <-m.eventChannel:
        default:
            return
        }
    }
}
```

#### 3.3.2: IP Bucket Map Growth (middleware/request.go:36)

Add max size check in cleanup:
```go
if len(buckets) > maxBuckets {
    // Evict oldest entries
}
```

#### 3.3.3: Log Collection Unbounded Slice (services/log_management_service.go:132-150)

Implement batch processing with max batch size:
```go
const maxLogBatchSize = 10000
if len(allEntries) >= maxLogBatchSize {
    // Flush batch
}
```

#### 3.3.4: Event Bus Handler Accumulation (Assets/pkg/event/bus.go:36-48)

Add TTL or max handler count with LRU eviction.

### Step 3.4: Fix React Memory Leaks

#### 3.4.1: AbortController (CollectionRealTime.tsx:266-272)

```tsx
useEffect(() => {
    const controller = new AbortController()
    connectWebSocket()
    return () => {
        controller.abort()
        disconnectWebSocket()
    }
}, [connectWebSocket, disconnectWebSocket])
```

#### 3.4.2: Notification Cleanup (CollectionRealTime.tsx:311-316)

Store Notification ref and close on cleanup.

#### 3.4.3: IntersectionObserver (PerformanceOptimizer.tsx:55-79)

Remove `loadedItems` from dependency array. Use callback ref pattern.

#### 3.4.4: Unbounded Map/Set State (PerformanceOptimizer.tsx:26-27)

Add max size eviction policy.

### Step 3.5: Tests for Phase 3

- **Race condition tests**: Run all affected packages with `-race` flag
- **Goroutine leak tests**: Verify goroutine count returns to baseline after operations
- **Memory pressure tests**: Allocate under load, verify no unbounded growth
- **Stress tests**: High-concurrency scenarios targeting all fixed areas

**New test files**:
- `catalog-api/services/sync_service_race_test.go`
- `catalog-api/services/error_reporting_race_test.go`
- `catalog-api/middleware/rate_limiter_leak_test.go`
- `catalog-api/internal/smb/resilience_leak_test.go`
- `catalog-web/src/components/collections/CollectionRealTime.test.tsx`
- `catalog-web/src/components/collections/PerformanceOptimizer.test.tsx`

### Step 3.6: Challenge Coverage for Phase 3

| Challenge ID | Name | Validates |
|-------------|------|-----------|
| CH-165 | Goroutine Leak Detection | Zero goroutine leaks under sustained load |
| CH-166 | Memory Pressure Stability | Memory stays bounded under 1000 concurrent requests |
| CH-167 | Rate Limiter Cleanup | Rate limiter maps don't grow unbounded |
| CH-168 | WebSocket Cleanup | WebSocket connections fully cleaned after disconnect |
| CH-169 | SMB Reconnection Stability | SMB reconnection uses bounded goroutine pool |
| CH-170 | Event Bus Safety | Event subscriptions cleaned properly |

---

## Phase 4: Lazy Loading, Semaphores & Non-Blocking Patterns {#phase-4}

**Duration**: 2-3 sessions | **Risk**: MEDIUM | **Dependencies**: Phase 3

### Step 4.1: Expand Lazy Initialization

1. **LazyServiceRegistry** -- Wire to actual service initialization (currently dead code)
2. **Database connections** -- Lazy connect on first query (not at startup)
3. **Metadata providers** -- Already lazy (verify with tests)
4. **Redis connection** -- Lazy connect on first cache operation
5. **File watcher** -- Lazy start when first storage root is added

### Step 4.2: Add Semaphore Controls

1. **File I/O within scans** -- Add semaphore limiting concurrent file reads during universal scan (currently only scan jobs are bounded, not I/O within jobs)
2. **Database query concurrency** -- Add configurable max concurrent query semaphore
3. **HTTP client requests** -- Verify pooled HTTP client has proper connection limits
4. **WebSocket broadcast** -- Semaphore on concurrent broadcast writes
5. **Backup operations** -- Semaphore ensuring only 1 backup at a time

### Step 4.3: Non-Blocking Patterns

1. **Metadata enrichment** -- Already async (verify with tests). Make sure goroutine has proper lifecycle.
2. **Scan progress updates** -- Use non-blocking channel sends for progress reporting
3. **Event bus publish** -- Non-blocking send with overflow logging
4. **Log streaming** -- Non-blocking log channel writes

### Step 4.4: Frontend Non-Blocking

1. **React.lazy()** -- Verify all route-level components use React.lazy
2. **Intersection Observer** -- Lazy load media thumbnails and list items
3. **Virtual scrolling** -- Verify react-window used for large lists
4. **Service Worker** -- Cache API responses for offline browsing
5. **Web Workers** -- Move heavy computation (search, sorting) to workers

### Step 4.5: Tests for Phase 4

- Benchmark tests measuring initialization time before/after lazy loading
- Stress tests verifying semaphore limits hold under load
- Responsiveness tests measuring API response times under concurrent load
- Frontend performance tests measuring Time to Interactive

**New test files**:
- `catalog-api/internal/lifecycle/lazy_registry_test.go`
- `catalog-api/tests/stress/semaphore_stress_test.go`
- `catalog-api/tests/stress/nonblocking_test.go`
- `catalog-web/src/__tests__/performance/lazy-loading.test.tsx`
- `tests/k6/semaphore_stress_test.js`

### Step 4.6: Challenge Coverage for Phase 4

| Challenge ID | Name | Validates |
|-------------|------|-----------|
| CH-171 | Lazy Service Initialization | Services initialize on first use, not startup |
| CH-172 | Scan Semaphore Enforcement | Concurrent file I/O bounded during scan |
| CH-173 | Database Query Limiting | Max concurrent queries enforced |
| CH-174 | Non-Blocking Event Bus | Event publish never blocks caller |
| CH-175 | API Responsiveness Under Load | p95 < 200ms with 50 concurrent users |
| CH-176 | Frontend Lazy Loading | Route components load on demand |

---

## Phase 5: Security Scanning & Remediation {#phase-5}

**Duration**: 2-3 sessions | **Risk**: LOW-MEDIUM | **Dependencies**: Phase 4

### Step 5.1: Run Snyk Scan

```bash
podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-scanner
```

1. Parse JSON output from `reports/snyk-*.json`
2. Categorize findings by severity (critical, high, medium, low)
3. Fix all critical and high findings
4. Document medium/low with justification if not fixing

### Step 5.2: Run SonarQube Scan

```bash
./scripts/sonarqube-scan.sh
```

1. Review quality gate results
2. Fix all bugs, vulnerabilities, and code smells rated Blocker or Critical
3. Reduce technical debt to < 5 days
4. Achieve 80%+ SonarQube coverage gate

### Step 5.3: Run Semgrep Scan

```bash
podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner
```

1. Parse SARIF/JSON output
2. Fix all WARNING+ findings
3. Add custom Semgrep rules for project-specific patterns:
   - No hardcoded credentials
   - No SQL string concatenation
   - No unvalidated user input in file paths
   - No unescaped template rendering

### Step 5.4: Run govulncheck

```bash
cd catalog-api && govulncheck ./...
```

Fix any reported Go stdlib or dependency vulnerabilities.

### Step 5.5: Run npm audit

```bash
cd catalog-web && npm audit --production
cd catalogizer-desktop && npm audit --production
cd installer-wizard && npm audit --production
```

Fix all critical and high npm vulnerabilities.

### Step 5.6: Create Custom Semgrep Rules

**New file**: `config/semgrep-rules.yml`

Rules for:
- No `fmt.Sprintf` in SQL queries
- No `os.Exec` with user-provided arguments
- No missing error checks on `db.Exec/Query`
- Mandatory `defer rows.Close()` after Query
- No `http.DefaultClient` (use pooled client)

### Step 5.7: Tests for Phase 5

- Security test suite validating all OWASP Top 10 mitigations
- Input validation tests for all API endpoints
- SQL injection tests for all repository methods
- XSS tests for all frontend rendered content

**New test files**:
- `catalog-api/middleware/security_comprehensive_test.go`
- `catalog-api/tests/security/owasp_top10_test.go`
- `catalog-api/tests/security/sql_injection_test.go`
- `catalog-web/src/__tests__/security/xss.test.tsx`

### Step 5.8: Challenge Coverage for Phase 5

| Challenge ID | Name | Validates |
|-------------|------|-----------|
| CH-177 | Snyk Zero Critical | Snyk scan returns 0 critical vulnerabilities |
| CH-178 | SonarQube Quality Gate | SonarQube quality gate PASSES |
| CH-179 | Semgrep Clean | Semgrep returns 0 WARNING+ findings |
| CH-180 | govulncheck Clean | govulncheck returns 0 vulnerabilities |
| CH-181 | npm Audit Clean | npm audit returns 0 production vulnerabilities |
| CH-182 | SQL Injection Prevention | All repository methods resist SQL injection |
| CH-183 | XSS Prevention | All rendered content escapes user input |
| CH-184 | CSRF Protection | All state-changing endpoints have CSRF protection |

---

## Phase 6: Test Coverage to Theoretical Maximum {#phase-6}

**Duration**: 5-7 sessions | **Risk**: LOW | **Dependencies**: Phases 1-5

### Step 6.1: Go Backend Coverage (65.1% -> 95%+)

#### Priority 1: Untested Source Files

| File | Tests Needed |
|------|-------------|
| `handlers/admin_handler.go` | CRUD, auth, error cases |
| `handlers/playlist_handler.go` | CRUD, pagination, error cases |
| `handlers/stub_handler.go` | (Will be replaced in Phase 1) |
| `repository/playlist_repository.go` | DB operations, edge cases |
| `services/playlist_service.go` | Business logic, validation |
| `database/tx_helpers.go` | Transaction helpers |
| `database/migrations_*.go` | Migration up/down |

#### Priority 2: Challenge File Unit Tests

Create unit tests for all 67 challenge implementation files. Test:
- Challenge initialization
- Execute() method logic
- Error handling
- Assertion validation

#### Priority 3: Coverage of All Branches

For every Go source file, ensure:
- All exported functions have tests
- All error branches are tested
- All edge cases (nil input, empty collections, max values) are tested

### Step 6.2: Frontend Coverage (46% -> 90%+)

#### Priority 1: Component Tests (58 untested components)

Test every React component with:
- Render test (renders without crash)
- Props validation test
- User interaction test (clicks, form submission)
- State change test
- Error boundary test
- Accessibility test (aria attributes)

**Key untested components**:
- `AdminPanel.tsx`
- `LoginForm.tsx`, `RegisterForm.tsx`, `ProtectedRoute.tsx`
- All 14 collection components
- All 7 playlist components
- All 5 media components
- All 4 entity components
- All 3 layout components
- All 13 UI components
- `SplashScreen.tsx`, `ErrorBoundary.tsx`

#### Priority 2: Hook Tests (8 untested hooks)

Test each custom hook with `renderHook()`:
- Initial state
- State transitions
- Cleanup on unmount
- Error handling

#### Priority 3: Page Tests (13 untested pages)

Test each page component:
- Route matching
- Data loading
- Navigation
- Protected route behavior
- Error states

#### Priority 4: Context & Utility Tests

- Auth context provider tests
- WebSocket context tests
- All utility function tests
- All lib/ function tests

### Step 6.3: Desktop App Coverage (3.8% -> 80%+)

Focus on testable TypeScript code:
- Tauri IPC command handlers
- State management
- UI components
- Form validation
- Configuration management

### Step 6.4: Installer Wizard Coverage (9.4% -> 80%+)

Focus on:
- Installation workflow state machine
- System requirement validation
- Configuration wizard steps
- Database initialization logic
- Network setup validation

### Step 6.5: Android Coverage (4.8% -> 80%+)

Focus on:
- ViewModel unit tests (StateFlow, business logic)
- Repository tests (Room, Retrofit mocking)
- Compose UI tests (semantics, interaction)
- Navigation tests

### Step 6.6: Android TV Coverage (3.2% -> 80%+)

Focus on:
- TV-specific navigation (D-pad)
- Focus management
- Media playback controls
- Banner/card components

### Step 6.7: API Client Coverage (0.6% -> 90%+)

Focus on:
- Every API client method
- Request/response serialization
- Error handling
- Authentication flows
- WebSocket operations
- Type safety validation

### Step 6.8: Challenge Coverage for Phase 6

| Challenge ID | Name | Validates |
|-------------|------|-----------|
| CH-185 | Go Coverage Gate | Go test coverage >= 95% |
| CH-186 | Frontend Coverage Gate | Frontend test coverage >= 90% |
| CH-187 | Desktop Coverage Gate | Desktop test coverage >= 80% |
| CH-188 | Installer Coverage Gate | Installer test coverage >= 80% |
| CH-189 | Android Coverage Gate | Android test coverage >= 80% |
| CH-190 | Android TV Coverage Gate | Android TV test coverage >= 80% |
| CH-191 | API Client Coverage Gate | API client test coverage >= 90% |

---

## Phase 7: Stress, Integration & Monitoring Tests {#phase-7}

**Duration**: 3-4 sessions | **Risk**: LOW | **Dependencies**: Phase 6

### Step 7.1: Enable All Stress Tests

The 54+ stress tests in `catalog-api/tests/stress/` are currently skipped in short mode. Create a dedicated stress test runner:

```bash
# New script: scripts/run-stress-tests.sh
cd catalog-api && GOMAXPROCS=3 go test ./tests/stress/... -v -count=1 -timeout 30m -p 1
```

Ensure all stress tests pass:
- `repository_stress_test.go` (6 tests)
- `database_stress_test.go` (6 tests)
- `cache_stress_test.go` (4 tests)
- `api_load_test.go` (6 tests)
- `concurrent_api_stress_test.go` (6 tests)
- `middleware_chain_stress_test.go` (5 tests)
- `memory_pressure_test.go` (4 tests)
- `concurrent_handlers_test.go` (4 tests)
- `rate_limiter_stress_test.go` (3 tests)
- `websocket_stress_test.go` (4 tests)
- `goroutine_leak_test.go` (1 test)
- `db_responsiveness_test.go` (2 tests)
- `responsiveness_test.go` (3 tests)

### Step 7.2: Create New Integration Tests

| Test Suite | File | What It Validates |
|-----------|------|-------------------|
| Full API Flow | `tests/integration/full_api_flow_test.go` | Register -> Login -> Create Root -> Scan -> Browse -> Search -> Collect -> Play |
| Backup/Restore | `tests/integration/backup_restore_test.go` | Create backup -> Modify data -> Restore -> Verify original state |
| Multi-Protocol | `tests/integration/multi_protocol_test.go` | SMB + FTP + NFS + WebDAV scanning in same session |
| Entity Pipeline | `tests/integration/entity_pipeline_complete_test.go` | Scan -> Aggregate -> Enrich -> Search -> Browse full cycle |
| WebSocket Events | `tests/integration/websocket_events_test.go` | Subscribe -> Trigger event -> Receive notification |
| Auth Lifecycle | `tests/integration/auth_lifecycle_test.go` | Register -> Login -> Refresh -> Change Password -> Login Again |
| Concurrent Users | `tests/integration/concurrent_users_test.go` | 10 users performing operations simultaneously |

### Step 7.3: Create Monitoring & Metrics Tests

| Test | File | What It Validates |
|------|------|-------------------|
| Prometheus Metrics | `tests/monitoring/prometheus_test.go` | All custom metrics emit correctly |
| Health Endpoint | `tests/monitoring/health_test.go` | /health returns correct component status |
| Memory Metrics | `tests/monitoring/memory_metrics_test.go` | Memory metrics accurate under load |
| Goroutine Metrics | `tests/monitoring/goroutine_metrics_test.go` | Goroutine count accurate |
| Cache Metrics | `tests/monitoring/cache_metrics_test.go` | Cache hit/miss ratios accurate |
| Latency Metrics | `tests/monitoring/latency_metrics_test.go` | Response time histograms accurate |

### Step 7.4: Expand k6 Load Tests

| Test | File | What It Validates |
|------|------|-------------------|
| Auth Flow Load | `tests/k6/auth_load_test.js` | 100 concurrent logins |
| Scan Load | `tests/k6/scan_load_test.js` | Concurrent scan operations |
| Search Load | `tests/k6/search_load_test.js` | 200 concurrent searches |
| WebSocket Load | `tests/k6/websocket_load_test.js` | 100 concurrent WebSocket connections |
| Entity Browse Load | `tests/k6/entity_browse_load_test.js` | 100 concurrent browse operations |
| Mixed Workload | `tests/k6/mixed_workload_test.js` | Realistic mixed read/write patterns |

### Step 7.5: Create Frontend Performance Tests

| Test | What It Validates |
|------|-------------------|
| Bundle Size Test | Total bundle < 500KB gzipped |
| First Paint Test | FCP < 1.5s |
| Interactive Test | TTI < 3s |
| Lighthouse Score | Performance > 90 |
| Component Render | No component renders > 16ms |

### Step 7.6: Responsiveness Optimization Based on Metrics

After running all monitoring tests:
1. Identify endpoints with p95 > 200ms
2. Add database query optimization (indexes, query plans)
3. Add caching for frequently-accessed read endpoints
4. Optimize serialization for large response payloads
5. Verify HTTP/3 + Brotli compression active on all responses

### Step 7.7: Challenge Coverage for Phase 7

| Challenge ID | Name | Validates |
|-------------|------|-----------|
| CH-192 | Stress Tests Pass | All 54+ stress tests pass |
| CH-193 | Integration Full Flow | Complete API flow works end-to-end |
| CH-194 | Concurrent Users Stable | 10 concurrent users, zero errors |
| CH-195 | k6 Load Test Pass | p95 < 500ms at 50 users |
| CH-196 | k6 Stress Test Pass | System stable at 200 users |
| CH-197 | k6 Soak Test Pass | Zero memory leaks in 30min soak |
| CH-198 | Prometheus Metrics Valid | All custom metrics emit correctly |
| CH-199 | Health Check Complete | /health validates all components |
| CH-200 | Frontend Performance | Lighthouse score > 90 |

---

## Phase 8: Challenge Bank Expansion {#phase-8}

**Duration**: 2-3 sessions | **Risk**: LOW | **Dependencies**: Phases 1-7

### Step 8.1: New Challenge Categories

**Total target**: 350+ challenges (currently 249)

| Category | ID Range | Count | Description |
|----------|----------|-------|-------------|
| Stub Implementation | CH-141 to CH-152 | 12 | Phase 1 endpoints |
| Module Integration | CH-153 to CH-164 | 12 | Phase 2 modules |
| Concurrency Safety | CH-165 to CH-170 | 6 | Phase 3 fixes |
| Performance Patterns | CH-171 to CH-176 | 6 | Phase 4 patterns |
| Security Scanning | CH-177 to CH-184 | 8 | Phase 5 scanning |
| Coverage Gates | CH-185 to CH-191 | 7 | Phase 6 coverage |
| Stress & Monitoring | CH-192 to CH-200 | 9 | Phase 7 tests |
| Playlist Feature | CH-201 to CH-210 | 10 | Playlist CRUD, ordering, sharing |
| Backup Feature | CH-211 to CH-220 | 10 | Backup/restore full lifecycle |
| Change Password | CH-221 to CH-225 | 5 | Password change workflow |
| Media Analysis | CH-226 to CH-235 | 10 | Quality analysis, format detection |
| Multi-Platform | CH-236 to CH-250 | 15 | Cross-platform consistency |

**New total**: 249 existing + 110 new = **359 challenges**

### Step 8.2: Register All New Challenges

Update `catalog-api/challenges/register.go` with all new challenge registrations.

### Step 8.3: Create Challenge Bank Definitions

Create JSON/YAML definitions in `challenges/config/` for each new challenge with:
- ID, Name, Description
- Category, Severity
- Prerequisites (other challenges that must pass first)
- Expected assertions
- Timeout

### Step 8.4: Verify All 359 Challenges

Run full challenge suite and verify 100% pass rate:
```bash
# Via catalog-api service (MANDATORY: compiled binary only)
curl -X POST http://localhost:8080/api/v1/challenges/run-all
```

---

## Phase 9: Documentation Completion {#phase-9}

**Duration**: 3-4 sessions | **Risk**: LOW | **Dependencies**: Phases 1-8

### Step 9.1: Create Missing CLAUDE.md Files

| Component | File | Content |
|-----------|------|---------|
| catalog-api | `catalog-api/CLAUDE.md` | Go architecture, handler/service/repo pattern, test conventions, database dialect |
| catalog-web | `catalog-web/CLAUDE.md` | React architecture, component patterns, state management, testing with Vitest |
| catalogizer-android | `catalogizer-android/CLAUDE.md` | MVVM, Compose, Room, Hilt, testing |
| catalogizer-androidtv | `catalogizer-androidtv/CLAUDE.md` | TV-specific navigation, D-pad, Leanback |

### Step 9.2: Create Missing AGENTS.md Files

Create for 8 submodules: Entities, Media-Types-TS, Catalogizer-API-Client-TS, Auth-Context-React, Media-Browser-React, Dashboard-Analytics-React, Media-Player-React, Collection-Manager-React

### Step 9.3: Create Missing docs/ Directories

| Component | Directory | Contents |
|-----------|-----------|----------|
| catalog-web | `catalog-web/docs/` | ARCHITECTURE.md, COMPONENTS.md, TESTING.md, SETUP.md |
| catalogizer-android | `catalogizer-android/docs/` | ARCHITECTURE.md, SETUP.md, TESTING.md, TROUBLESHOOTING.md |
| catalogizer-androidtv | `catalogizer-androidtv/docs/` | ARCHITECTURE.md, SETUP.md, TESTING.md, TV_SPECIFIC.md |
| DocProcessor | `DocProcessor/docs/` | ARCHITECTURE.md, USAGE.md, API.md |
| LLMOrchestrator | `LLMOrchestrator/docs/` | ARCHITECTURE.md, USAGE.md, AGENTS.md |
| VisionEngine | `VisionEngine/docs/` | ARCHITECTURE.md, USAGE.md, PROVIDERS.md |
| LLMProvider | `LLMProvider/docs/` | ARCHITECTURE.md, PROVIDERS.md, API.md + README.md |

### Step 9.4: Update OpenAPI Specification

Update `docs/api/openapi.yaml` with:
- All formerly-stubbed endpoints now fully documented
- Backup/restore endpoints
- Playlist endpoints
- Change password endpoint
- Response schemas for all endpoints
- Error response schemas
- Authentication flows

### Step 9.5: Update Data Dictionary

Update `docs/DATA_DICTIONARY.md` with:
- Any new tables from Phase 1-2 (backups, playlists if new)
- Updated column descriptions
- New indexes from optimization
- Relationship diagrams updated

### Step 9.6: Update SQL Definitions

Update `docs/SQL_COMPLETE_SCHEMA.md` and `docs/SQL_MIGRATIONS.md` with:
- All new migrations
- Complete current schema DDL
- Index definitions
- Constraint documentation

### Step 9.7: Update Architecture Diagrams

Update all diagrams in `docs/diagrams/`:
- System architecture SVG -- add newly wired modules
- Entity-relationship SVG -- add any new tables
- Sequence diagrams -- add backup/restore, password change flows
- Component interaction -- add module integration points

### Step 9.8: Update Architecture Documentation

Update these key documents:
- `docs/ARCHITECTURE.md` -- Add module integration architecture
- `docs/CONCURRENCY_PATTERNS.md` -- Add new semaphore and non-blocking patterns
- `docs/LAZY_LOADING.md` -- Add new lazy initialization patterns
- `docs/OPTIMIZATION_GUIDE.md` -- Add monitoring-driven optimizations
- `docs/SECURITY_HEADERS.md` -- Add custom Semgrep rules

### Step 9.9: Update User Guides

| Guide | Updates Needed |
|-------|---------------|
| `docs/guides/WEB_APP_GUIDE.md` | Password change, playlist management, media quality view |
| `docs/guides/ANDROID_GUIDE.md` | Any new features exposed in mobile |
| `docs/guides/ANDROID_TV_GUIDE.md` | Any new features on TV |
| `docs/guides/DESKTOP_GUIDE.md` | Backup/restore from desktop |
| `docs/ADMIN_GUIDE.md` | Backup management, security scanning, monitoring |
| `docs/TROUBLESHOOTING.md` | New error scenarios, new debugging steps |
| `docs/DEVELOPER_GUIDE.md` | Module integration, testing guide updates |

### Step 9.10: Update Deployment Documentation

| Document | Updates Needed |
|----------|---------------|
| `docs/DEPLOYMENT_GUIDE.md` | Add module wiring verification steps |
| `docs/PRODUCTION_DEPLOYMENT_GUIDE.md` | Add monitoring verification |
| `docs/PRODUCTION_RUNBOOK.md` | Add backup/restore runbook |
| `docs/MONITORING_GUIDE.md` | Add new metrics and alerts |
| `docs/BACKUP_AND_RECOVERY.md` | Complete rewrite with real implementation |

### Step 9.11: Add Go Package Documentation

For every Go package in catalog-api:
- Add package-level doc comment (`// Package xxx provides...`)
- Document all exported functions with godoc comments
- Document all exported types and interfaces

### Step 9.12: Add TypeScript JSDoc/TSDoc

For all public exports in catalog-web and TypeScript packages:
- Add JSDoc comments to exported functions
- Add `@param`, `@returns`, `@throws` annotations
- Add `@example` code snippets for complex APIs

---

## Phase 10: Video Course & Website Update {#phase-10}

**Duration**: 2-3 sessions | **Risk**: LOW | **Dependencies**: Phase 9

### Step 10.1: Update Video Course Outline

Update `docs/courses/COURSE_OUTLINE.md` with new modules:

| Module | Updates |
|--------|---------|
| Module 2: Getting Started | Add backup/restore workflow |
| Module 3: Media Management | Add media quality analysis, media analysis |
| Module 4: Collections & Playlists | Add playlist management |
| Module 5: Administration | Add backup management, security scanning |
| Module 6: Advanced Features | Add password change, module integration |
| Module 7: Testing | Add stress testing, monitoring tests, challenge expansion |
| Module 8: Production | Add Snyk/SonarQube/Semgrep scanning |
| Module 9: Architecture | Add module decoupling, lazy loading, semaphores |
| Module 10: Security | Add custom Semgrep rules, OWASP compliance |
| Module 11: Monitoring | Add Prometheus metrics, Grafana dashboards, alerting |
| Module 12: Multi-Platform | Add cross-platform consistency, platform-specific features |

### Step 10.2: Update Course Exercises

Update `docs/courses/EXERCISES.md` with hands-on exercises for:
- Backup/restore workflow
- Running security scans
- Creating custom challenges
- Monitoring with Grafana
- Load testing with k6
- Stress testing Go services

### Step 10.3: Update Course Assessments

Update `docs/courses/ASSESSMENT.md` with new questions covering:
- Module integration architecture
- Concurrency safety patterns
- Security scanning tools
- Monitoring and alerting
- Performance optimization

### Step 10.4: Create Course Slide Decks

Create/update slide content in `docs/courses/slides/` for each module.

### Step 10.5: Update Website Content

Update all files in `Website/`:

| File | Updates |
|------|---------|
| `index.md` | Update feature highlights, add new capabilities |
| `features.md` | Add backup/restore, playlists, media quality, security scanning |
| `getting-started.md` | Update quick start with new features |
| `download.md` | Update version to reflect new build |
| `faq.md` | Add FAQ entries for new features |
| `documentation.md` | Update documentation index with new docs |
| `changelog.md` | Add changelog entries for all phases |
| `course.md` | Update course information with new modules |
| `support.md` | Update support channels |

### Step 10.6: Update Website Guides

| Guide | Updates |
|-------|---------|
| `guides/android.md` | New features, updated screenshots |
| `guides/android-tv.md` | New features, updated screenshots |
| `guides/configuration.md` | New configuration options |
| `guides/desktop.md` | Backup/restore from desktop |
| `guides/monitoring.md` | New metrics, dashboards |
| `guides/security.md` | Security scanning, Semgrep rules |
| `guides/web-app.md` | New features, updated screenshots |

### Step 10.7: Update Website Developer Docs

| File | Updates |
|------|---------|
| `docs/developer-guide/api-reference.md` | All new endpoints |
| `docs/developer-guide/monitoring.md` | New metrics and alerts |
| `docs/developer-guide/security.md` | Custom Semgrep rules, scanning |
| `docs/developer-guide/testing-strategy.md` | Stress tests, monitoring tests |

---

## Phase 11: Final Verification & Release {#phase-11}

**Duration**: 2-3 sessions | **Risk**: LOW | **Dependencies**: All previous phases

### Step 11.1: Run Complete Test Suite

```bash
# Go backend (all test types)
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -race -count=1

# Frontend
cd catalog-web && npm run test && npm run test:e2e

# Desktop
cd catalogizer-desktop && npm run test

# Installer
cd installer-wizard && npm run test

# API client
cd catalogizer-api-client && npm run test

# Android (requires JDK 17)
cd catalogizer-android && ./gradlew test
cd catalogizer-androidtv && ./gradlew test
```

### Step 11.2: Run Stress Tests

```bash
cd catalog-api && go test ./tests/stress/... -v -count=1 -timeout 30m -p 1
```

### Step 11.3: Run Security Scans

```bash
# All scans
./scripts/security-scan.sh

# SonarQube
./scripts/sonarqube-scan.sh

# Semgrep
podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner

# Snyk
podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-scanner
```

### Step 11.4: Run Load Tests

```bash
podman run --rm --network host \
  -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest run /scripts/load_test.js

podman run --rm --network host \
  -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest run /scripts/stress_test.js

podman run --rm --network host \
  -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest run /scripts/soak_test.js
```

### Step 11.5: Run All Challenges

```bash
# Start services
podman-compose -f docker-compose.dev.yml up -d

# Run all 359 challenges via API
curl -X POST http://localhost:8080/api/v1/challenges/run-all
```

Verify 100% pass rate.

### Step 11.6: Verify Coverage Gates

| Component | Gate | Verified? |
|-----------|------|-----------|
| catalog-api | >= 95% | |
| catalog-web | >= 90% | |
| catalogizer-desktop | >= 80% | |
| installer-wizard | >= 80% | |
| catalogizer-android | >= 80% | |
| catalogizer-androidtv | >= 80% | |
| catalogizer-api-client | >= 90% | |

### Step 11.7: Verify Zero Issues

| Check | Expected | Verified? |
|-------|----------|-----------|
| Go test failures | 0 | |
| Frontend test failures | 0 | |
| Race conditions detected | 0 | |
| Goroutine leaks | 0 | |
| Memory leaks | 0 | |
| Security vulnerabilities (critical) | 0 | |
| SonarQube quality gate | PASS | |
| Semgrep findings (WARNING+) | 0 | |
| Console errors (browser) | 0 | |
| Console warnings (browser) | 0 | |
| Failed network requests | 0 | |
| Dead code / stubs | 0 | |
| Unused modules | 0 | |
| Missing documentation | 0 | |

### Step 11.8: Build Release

```bash
./scripts/release-build.sh --container --force
```

All 7 components must build successfully.

### Step 11.9: Update versions.json

Bump version to reflect the comprehensive update:
- `1.0.0 build 12` -> `2.0.0 build 13`

### Step 11.10: Final Documentation

- Write `docs/status/FINAL_COMPLETION_REPORT_2026-03-30.md` with all metrics
- Update `README.md` with current state
- Update `CLAUDE.md` with any new patterns/conventions
- Update `AGENTS.md` with any new commands

### Step 11.11: Commit & Push

```bash
# Commit all changes across all submodules and main repo
# Push to all 6 remotes
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

---

## Constraints & Rules {#constraints}

All work MUST comply with these constraints from CLAUDE.md and AGENTS.md:

### Non-Negotiable Rules

1. **Container-only builds (Podman)** -- No bare metal builds/services in production/QA
2. **No GitHub Actions** -- All CI/CD runs locally
3. **No API keys in git** -- .env.example with placeholders only
4. **HTTP/3 (QUIC) + Brotli** -- All network communication
5. **Host resource limits** -- 30-40% max (GOMAXPROCS=3, container CPU/mem limits)
6. **SQLite WAL mode** -- Explicit PRAGMA after connection
7. **Zero warning/error policy** -- No console errors, no failed requests
8. **HelixQA fully autonomous** -- No hardcoded flows
9. **Challenge ops by compiled binaries only** -- No curl/scripts in challenges
10. **No interactive processes** -- No sudo/root password prompts during execution

### Code Standards

- **Go**: NewService constructor injection, error wrapping, table-driven tests
- **TypeScript**: PascalCase components, camelCase functions, Zod validation
- **Kotlin**: MVVM, Result sealed classes, Room for offline
- **Config**: env vars > .env > config.json > defaults
- **PostCSS**: CommonJS (module.exports) for Node 18 compat

### Resource Budgets

| Resource | Limit |
|----------|-------|
| Go tests | `GOMAXPROCS=3 -p 2 -parallel 2` |
| PostgreSQL | `--cpus=1 --memory=2g` |
| API container | `--cpus=2 --memory=4g` |
| Web container | `--cpus=1 --memory=2g` |
| Builder container | `--cpus=3 --memory=8g` |
| Total containers | max 4 CPUs, 8 GB RAM |
| Challenges | Sequential only, never parallel |

---

## Risk Matrix {#risk-matrix}

| Phase | Risk Level | Primary Risk | Mitigation |
|-------|-----------|-------------|------------|
| Phase 1 (Stubs) | LOW | New endpoint bugs | Table-driven tests, challenge verification |
| Phase 2 (Modules) | MEDIUM | Breaking existing functionality | Run full test suite after each module integration |
| Phase 3 (Concurrency) | HIGH | Introducing new race conditions | -race flag on all tests, stress testing |
| Phase 4 (Lazy/Semaphore) | MEDIUM | Deadlocks from new semaphores | Careful ordering, timeout-based deadlock detection |
| Phase 5 (Security) | LOW | False positives in scans | Manual review of all findings |
| Phase 6 (Coverage) | LOW | Test maintenance burden | Focus on behavior tests, not implementation tests |
| Phase 7 (Stress) | LOW | Flaky stress tests | Deterministic test design, adequate timeouts |
| Phase 8 (Challenges) | LOW | Challenge registration errors | Automated verification of all registrations |
| Phase 9 (Docs) | LOW | Stale documentation | Generate docs from code where possible |
| Phase 10 (Content) | LOW | Content inconsistency | Cross-reference with implementation |
| Phase 11 (Verify) | LOW | Missed issues | Comprehensive checklist verification |

---

## Summary of Deliverables

| Phase | New Files | Modified Files | New Tests | New Challenges |
|-------|-----------|---------------|-----------|----------------|
| Phase 1 | ~5 | ~8 | ~50 | 12 (CH-141 to CH-152) |
| Phase 2 | ~12 | ~25 | ~60 | 12 (CH-153 to CH-164) |
| Phase 3 | ~8 | ~15 | ~40 | 6 (CH-165 to CH-170) |
| Phase 4 | ~6 | ~12 | ~30 | 6 (CH-171 to CH-176) |
| Phase 5 | ~5 | ~8 | ~30 | 8 (CH-177 to CH-184) |
| Phase 6 | ~200+ | ~50 | ~500+ | 7 (CH-185 to CH-191) |
| Phase 7 | ~25 | ~10 | ~60 | 9 (CH-192 to CH-200) |
| Phase 8 | ~15 | ~5 | ~20 | 60 (CH-201 to CH-250) |
| Phase 9 | ~40+ | ~30 | 0 | 0 |
| Phase 10 | ~10 | ~25 | 0 | 0 |
| Phase 11 | ~3 | ~10 | 0 | 0 |
| **TOTAL** | **~329+** | **~198** | **~790+** | **120 (new)** |

**Final totals**:
- **Challenges**: 249 + 110 = **359**
- **New test files**: ~790+
- **Coverage targets**: Go 95%+, Frontend 90%+, All others 80%+
- **Security gates**: Snyk, SonarQube, Semgrep, govulncheck, npm audit all PASS
- **Zero**: dead code, stubs, goroutine leaks, memory leaks, race conditions, console errors

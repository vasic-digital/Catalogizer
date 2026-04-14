# Comprehensive Completion Plan v4 -- Full Project Audit & Implementation Roadmap

**Date**: 2026-04-14
**Version**: 2.3.0 (Build 22) -> Target: 2.4.0 (Build 23)
**Scope**: All 7 applications, 32 submodules, all test categories, all documentation, website
**Supersedes**: All previous plans in `docs/plans/`

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Full Audit Report](#2-full-audit-report)
3. [Phased Implementation Plan](#3-phased-implementation-plan)
4. [Test Coverage Matrix](#4-test-coverage-matrix)
5. [Documentation Deliverables](#5-documentation-deliverables)
6. [Challenge & Bank Expansion](#6-challenge--bank-expansion)
7. [Risk Register](#7-risk-register)

---

## 1. Executive Summary

### Current State

The Catalogizer project is a mature multi-platform media collection manager with 7 application components, 32 Git submodules, and extensive infrastructure. A comprehensive 7-axis audit was performed covering:

1. **Unfinished code** -- 2 stub features found (Android TV episode navigation)
2. **Test coverage** -- Critical gaps: Rust/Tauri 0%, Android 5.3%, Android TV 10.8%
3. **Dead code** -- 5 unused hooks, 4 disconnected features, 3 unused middleware, 2 mock APIs
4. **Concurrency safety** -- 29 issues: 4 critical resource leaks, 14 high-severity race conditions
5. **Documentation** -- 6 missing AGENTS.md, website underdeveloped, Article V gaps
6. **Security scanning** -- Infrastructure mature; missing Hadolint, custom Semgrep rules, govulncheck integration
7. **Build configuration** -- 3 critical version mismatches (Go, Android SDK, versions.json)

### Severity Summary

| Severity | Count | Category |
|----------|-------|----------|
| CRITICAL | 7 | 3 resource leaks, 1 goroutine leak, 3 build version mismatches |
| HIGH | 18 | 5 race conditions, 3 deadlock risks, 4 memory leaks, 6 coverage gaps |
| MEDIUM | 14 | Config issues, partial implementations, documentation gaps |
| LOW | 8 | Dead code cleanup, naming inconsistencies |

### Target Outcome

Zero unfinished work. Zero broken tests. Zero disabled features. 100% test coverage across all 10 Constitution Article V categories. Complete documentation. Updated website. Extended video courses. All security scans passing.

---

## 2. Full Audit Report

### 2.1 Unfinished Code

| ID | Component | File | Issue | Severity |
|----|-----------|------|-------|----------|
| UC-001 | catalogizer-androidtv | `ui/player/VLCPlayerActivity.kt:348` | ~~`playNextEpisode()` returns toast stub~~ **FIXED** | RESOLVED |
| UC-002 | catalogizer-androidtv | `ui/player/VLCPlayerActivity.kt:359` | ~~`playPreviousEpisode()` returns toast stub~~ **FIXED** | RESOLVED |

**Resolution**: Both now use `/api/v1/entities/:id/children` API to resolve sibling episodes. Implemented in Phase 3.

### 2.2 Test Coverage Gaps

| ID | Component | Source Files | Test Files | Coverage | Gap |
|----|-----------|-------------|------------|----------|-----|
| TC-001 | catalogizer-desktop (Rust) | 46 | 0 | 0% | CRITICAL |
| TC-002 | installer-wizard (Rust) | 47 | 0 | 0% | CRITICAL |
| TC-003 | catalogizer-android | 1053 | 56 | 5.3% | CRITICAL |
| TC-004 | catalogizer-androidtv | 436 | 47 | 10.8% | CRITICAL |
| TC-005 | catalogizer-api-client | 9 | 7 | 77.8% | HIGH |
| TC-006 | Auth-Context-React | 3 | 1 | 33% | HIGH |
| TC-007 | Media-Player-React | 5 | 3 | 60% | HIGH |
| TC-008 | Collection-Manager-React | 6 | 4 | 67% | HIGH |
| TC-009 | Dashboard-Analytics-React | 6 | 4 | 67% | HIGH |
| TC-010 | catalog-web (6 skipped tests) | -- | -- | -- | MEDIUM |
| TC-011 | installer-wizard (10 step components) | 10 | 0 | 0% | HIGH |
| TC-012 | catalog-api | 337 | 334 | 99% | LOW |
| TC-013 | catalog-web | 137 | 130 | 95% | LOW |

**Missing E2E suites**: catalogizer-desktop, installer-wizard, catalogizer-android, catalogizer-androidtv.
**Missing integration tests**: catalogizer-desktop, installer-wizard.

### 2.3 Dead Code & Disconnected Features

#### 2.3.1 Unused React Hooks (catalog-web)

| ID | File | Hook | Status |
|----|------|------|--------|
| DC-001 | `src/hooks/useDebounce.ts` | useDebounce | Never imported |
| DC-002 | `src/hooks/useLazyImage.ts` | useLazyImage | Never imported |
| DC-003 | `src/hooks/usePlayerState.tsx` | usePlayerState | Superseded |
| DC-004 | `src/hooks/usePlaylistReorder.tsx` | usePlaylistReorder | Not wired to UI |
| DC-005 | `src/hooks/useVirtualScroll.ts` | useVirtualScroll | Not wired to UI |

#### 2.3.2 Disconnected Features (catalog-web)

| ID | File | Feature | Status |
|----|------|---------|--------|
| DC-006 | `src/components/collections/CollectionSharing.tsx` | Collection sharing | Not routed |
| DC-007 | `src/components/collections/CollectionRealTime.tsx` | Real-time sync | Not imported |
| DC-008 | `src/components/collections/PerformanceOptimizer.tsx` | Performance monitor | Not imported |
| DC-009 | `src/components/collections/ExternalIntegrations.tsx` | External integrations | Partial impl |

#### 2.3.3 Mock API Stubs

| ID | File | Issue |
|----|------|-------|
| DC-010 | `src/lib/conversionApi.ts` | Falls back to mock data on 404 |
| DC-011 | `src/lib/mockCollectionsApi.ts` | Full mock API -- backend incomplete |

#### 2.3.4 Unused Go Code

| ID | File | Issue |
|----|------|-------|
| DC-012 | `middleware/advanced_rate_limiter.go` | Never applied in main.go |
| DC-013 | `middleware/enhanced_rate_limiter.go` | Never instantiated |
| DC-014 | `middleware/redis_rate_limiter.go` | Documented but never used |
| DC-015 | `handlers/stub_handler.go:638` | Deprecated `NewStubHandler()` |
| DC-016 | `config/config.go` | 8 config fields parsed but never read |
| DC-017 | `.env.example` | 10 unused AI/Vision env vars |

### 2.4 Concurrency & Safety Issues

#### 2.4.1 CRITICAL (Must Fix Immediately)

| ID | File | Issue | Impact |
|----|------|-------|--------|
| CS-001 | `utils/concurrency.go:85` | `SubmitAsync` goroutine leak -- no WaitGroup tracking | Unbounded goroutine growth |
| CS-002 | `internal/media/providers/providers.go:375` | Missing `resp.Body.Close()` on ReadAll error | File descriptor leak |
| CS-003 | `filesystem/webdav_client.go:266` | Missing `resp.Body.Close()` on ReadAll error | File descriptor leak |
| CS-004 | `handlers/admin_handler.go:260` | Missing `rows.Close()` after Query | DB connection pool exhaustion |

#### 2.4.2 HIGH (Fix Before Release)

| ID | File | Issue | Impact |
|----|------|-------|--------|
| CS-005 | `utils/concurrency.go:160` | Throttler goroutine not tracked, double-close panic | Goroutine leak |
| CS-006 | `internal/media/realtime/watcher.go:133` | `monitorPath` adds untracked goroutines | Unbounded growth |
| CS-007 | `handlers/websocket_handler.go:237` | Lock released then reacquired -- race on capacity | Over-admission |
| CS-008 | `utils/concurrency.go:233` | Debouncer: pending fn can re-enter causing deadlock | Deadlock |
| CS-009 | `internal/media/realtime/watcher.go:31` | `changeQueue` channel never closed in Stop() | Worker goroutine leak |
| CS-010 | `services/cache_service.go:134` | Context from Background(), not cancellable on shutdown | Delayed shutdown |
| CS-011 | `catalog-web: WebSocketContext.tsx:42` | `webSocket` in useEffect deps causes reconnect storms | Memory leak |
| CS-012 | `catalog-web: lib/websocket.ts:129` | dispose() may not stop reconnection attempts | Connection leak |
| CS-013 | `catalogizer-android: SyncManager.kt:99` | No cancellation on logout | Background work after logout |
| CS-014 | `catalogizer-android: MediaRepository.kt:20` | Flow collection without lifecycle awareness | Memory leak |

#### 2.4.3 MEDIUM

| ID | File | Issue |
|----|------|-------|
| CS-015 | `catalog-web: UploadManager.tsx:40` | Timer not cleared on rapid calls |
| CS-016 | `catalog-web: AuthContext.tsx:78` | Duplicate permissions useEffect |
| CS-017 | `catalogizer-android: CatalogizerDatabase.kt:55` | Room DB singleton never closed |
| CS-018 | `catalogizer-android: MediaRepository.kt:110` | `Flow.first()` without timeout |
| CS-019 | `catalogizer-android: AuthRepository.kt:39` | Multiple DataStore collectors without sharing |

### 2.5 Build Configuration Issues

| ID | Issue | Current | Expected | Severity |
|----|-------|---------|----------|----------|
| BC-001 | Dockerfile Go version mismatch | 1.26.1 | 1.25.7 | CRITICAL |
| BC-002 | Android compileSdk vs Dockerfile SDK | 35 vs 34 | Must match | CRITICAL |
| BC-003 | versions.json vs package.json | 2.3.0 vs 2.2.0 | Must match | HIGH |
| BC-004 | Go submodule versions (LLM*) | 1.24.x | 1.25.x | HIGH |
| BC-005 | Desktop build stale | Build 17 | Build 22+ | MEDIUM |
| BC-006 | Tauri version inconsistency | >=2.0.0 vs 2.8.0 | Pin both | LOW |
| BC-007 | API client naming | Two names | One canonical | LOW |
| BC-008 | TS submodule versions | 0.1.0 | 1.x.x | LOW |

### 2.6 Security Scanning Gaps

| ID | Tool | Status | Gap |
|----|------|--------|-----|
| SS-001 | Semgrep | Auto-mode only | No custom `.semgrep.yml` rules |
| SS-002 | Hadolint | Missing | No Dockerfile linting |
| SS-003 | govulncheck | Not in main scan script | Only in QA scripts |
| SS-004 | SonarQube | Always-on (no profile) | Wastes resources |
| SS-005 | Snyk | Configured but SNYK_TOKEN undocumented | Needs .env.example entry |
| SS-006 | License scanning | Missing | No SBOM generation |

### 2.7 Documentation Gaps

| ID | Gap | Impact |
|----|-----|--------|
| DG-001 | 6 missing AGENTS.md (web, android, androidtv, desktop, api-client, wizard) | No AI agent guidance |
| DG-002 | Website has only 9 pages + 4 dev-guide files | Underdeveloped public docs |
| DG-003 | No video courses for core components (api, web, android, desktop) | Training gap |
| DG-004 | No SQL migration reference doc | Schema change tracking |
| DG-005 | No performance tuning guide | Ops gap |
| DG-006 | Multiple overlapping plans (13 files) | Confusion on current plan |
| DG-007 | Missing component API_REFERENCE.md, CONTRIBUTING.md | Contributor barrier |
| DG-008 | Constitution Article V: missing DDoS/benchmarking test bank coverage | Compliance gap |

---

## 3. Phased Implementation Plan

### Phase Overview

| Phase | Name | Scope | Status |
|-------|------|-------|--------|
| 1 | Critical Safety Fixes | CS-001..004, BC-001..002 | COMPLETE |
| 2 | Resource Leak & Concurrency Hardening | CS-005..019, BC-003..004 | COMPLETE |
| 3 | Dead Code Wiring & Feature Completion | DC-001..017, UC-001..002 | COMPLETE |
| 4 | Test Coverage: Go & TypeScript | TC-005..013, skipped tests | IN PROGRESS |
| 5 | Test Coverage: Rust/Tauri | TC-001..002 | IN PROGRESS |
| 6 | Test Coverage: Android & Android TV | TC-003..004 | REMAINING |
| 7 | Security Scanning & Hardening | SS-001..006, Snyk, SonarQube runs | COMPLETE |
| 8 | Stress, DDoS, Benchmarking Tests | Article V categories 5,7,8 | COMPLETE (benchmarks) |
| 9 | Challenge & Bank Expansion | All 10 categories, fixes-validation | IN PROGRESS |
| 10 | Lazy Loading, Semaphores, Non-Blocking | Performance optimization | COMPLETE |
| 11 | Documentation Completion | AGENTS.md, guides, SQL, API refs | COMPLETE |
| 12 | Video Courses & Training | Course scripts for all components | COMPLETE |
| 13 | Website Update & Content Migration | VitePress pages, developer guide | IN PROGRESS |
| 14 | Version Alignment & Release Build | Version bump, container build, final QA | PARTIAL (versions aligned) |

---

### Phase 1: Critical Safety Fixes

**Goal**: Fix all CRITICAL issues that could cause crashes, data loss, or build failures.

#### Step 1.1: Fix Dockerfile Go Version (BC-001)
- **File**: `docker/Dockerfile.builder`
- **Action**: Change Go version from 1.26.1 to 1.25.7
- **Verify**: `podman build --network host -f docker/Dockerfile.builder -t catalogizer-builder:latest .`

#### Step 1.2: Fix Android SDK Version (BC-002)
- **File**: `docker/Dockerfile.builder`
- **Action**: Add `sdkmanager "platforms;android-35"` to Dockerfile
- **Alternative**: Downgrade `catalogizer-android/app/build.gradle.kts` compileSdk from 35 to 34
- **Decision**: Prefer upgrading Dockerfile (keeps app on latest SDK)

#### Step 1.3: Fix SubmitAsync Goroutine Leak (CS-001)
- **File**: `catalog-api/utils/concurrency.go:85-93`
- **Action**: Add WaitGroup tracking to SubmitAsync goroutines
- **Pattern**: `p.wg.Add(1); go func() { defer p.wg.Done(); p.Submit(job) }()`
- **Test**: Add `TestSubmitAsync_GoroutineTracking` verifying no goroutine leak after Stop()

#### Step 1.4: Fix HTTP Response Body Leaks (CS-002, CS-003)
- **File**: `catalog-api/internal/media/providers/providers.go:375`
- **Action**: Add `defer resp.Body.Close()` before `io.ReadAll`
- **File**: `catalog-api/filesystem/webdav_client.go:266`
- **Action**: Add `defer resp.Body.Close()` before `io.ReadAll`
- **Test**: Add `TestProviderHTTP_BodyClosedOnError` and `TestWebDAVClient_BodyClosedOnError`

#### Step 1.5: Fix Database rows.Close() Leak (CS-004)
- **File**: `catalog-api/handlers/admin_handler.go:260`
- **Action**: Add `defer rows.Close()` after Query call
- **Audit**: Grep all `db.Query(` and `db.QueryRow(` calls for missing Close()
- **Test**: Add integration test verifying connection pool recovery under load

#### Step 1.6: Verify Fixes
- Run: `cd catalog-api && go vet ./... && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -race`
- Run: `cd catalog-web && npm run lint && npm run type-check && npm run test`
- Confirm zero new warnings/errors

---

### Phase 2: Resource Leak & Concurrency Hardening

**Goal**: Fix all HIGH-severity concurrency issues across Go, TypeScript, and Kotlin.

#### Step 2.1: Fix Throttler Goroutine Tracking (CS-005)
- **File**: `catalog-api/utils/concurrency.go:160`
- **Action**: Add WaitGroup to Throttler; use `CompareAndSwap` for Stop() to prevent double-close
- **Test**: `TestThrottler_StopWaitsForGoroutine`, `TestThrottler_DoubleStopSafe`

#### Step 2.2: Fix SMBChangeWatcher Issues (CS-006, CS-009)
- **File**: `catalog-api/internal/media/realtime/watcher.go`
- **Action**: Cap concurrent monitorPath goroutines; close changeQueue in Stop()
- **Test**: `TestSMBChangeWatcher_StopClosesChannel`, `TestSMBChangeWatcher_BoundedGoroutines`

#### Step 2.3: Fix WebSocketHandler Lock Ordering (CS-007)
- **File**: `catalog-api/handlers/websocket_handler.go:237-281`
- **Action**: Hold lock through capacity check AND client registration atomically
- **Test**: `TestWebSocketHandler_ConcurrentConnections_NoOverAdmission`

#### Step 2.4: Fix Debouncer Deadlock Risk (CS-008)
- **File**: `catalog-api/utils/concurrency.go:233-252`
- **Action**: Copy pending function before releasing lock; execute outside lock
- **Test**: `TestDebouncer_ReentrantCallNoPanic`

#### Step 2.5: Fix CacheService Context Propagation (CS-010)
- **File**: `catalog-api/services/cache_service.go:134`
- **Action**: Use service's own context (derived from shutdown) instead of `context.Background()`
- **Test**: `TestCacheService_ShutdownCancelsCleanup`

#### Step 2.6: Fix React WebSocket Memory Leak (CS-011, CS-012)
- **File**: `catalog-web/src/contexts/WebSocketContext.tsx:42-55`
- **Action**: Remove `webSocket` from useEffect dependency array; use ref instead
- **File**: `catalog-web/src/lib/websocket.ts:129-134`
- **Action**: Ensure dispose() cancels all reconnection timers and removes all event listeners
- **Test**: `WebSocketContext.reconnect.test.tsx`

#### Step 2.7: Fix React AuthContext Duplicate Effect (CS-016)
- **File**: `catalog-web/src/contexts/AuthContext.tsx:78-92`
- **Action**: Remove duplicate permissions useEffect; consolidate into single effect
- **Test**: Existing auth tests should cover

#### Step 2.8: Fix React UploadManager Timer Leak (CS-015)
- **File**: `catalog-web/src/components/upload/UploadManager.tsx:40-46`
- **Action**: Clear previous timer before setting new one; add unmount guard for async ops
- **Test**: `UploadManager.timer.test.tsx`

#### Step 2.9: Fix Android SyncManager Cancellation (CS-013)
- **File**: `catalogizer-android/app/.../data/sync/SyncManager.kt:99`
- **Action**: Accept `CoroutineContext` parameter; check `ensureActive()` at sync boundaries
- **Action**: Cancel sync on logout via `SyncManager.cancelSync()` in AuthRepository.logout()
- **Test**: `SyncManagerTest.kt: testSyncCancelledOnLogout`

#### Step 2.10: Fix Android Flow Lifecycle (CS-014, CS-018, CS-019)
- **File**: `catalogizer-android/app/.../data/repository/MediaRepository.kt`
- **Action**: Add `withTimeoutOrNull(5000)` around `Flow.first()` calls
- **File**: `catalogizer-android/app/.../data/repository/AuthRepository.kt`
- **Action**: Share DataStore Flow via `shareIn(scope, SharingStarted.Lazily, 1)`
- **Test**: `MediaRepositoryTest.kt: testFlowFirstWithTimeout`, `AuthRepositoryTest.kt: testSharedDataStoreFlow`

#### Step 2.11: Fix Android Room DB Lifecycle (CS-017)
- **File**: `catalogizer-android/app/.../data/local/CatalogizerDatabase.kt:55`
- **Action**: Add `close()` method that nullifies INSTANCE; call from Application.onTerminate()
- **Test**: `CatalogizerDatabaseTest.kt: testDatabaseCloseAndReopen`

#### Step 2.12: Run Full Test Suites
- Go: `cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -race`
- Frontend: `cd catalog-web && npm run test && npm run lint && npm run type-check`
- Android: `cd catalogizer-android && ./gradlew test`
- Android TV: `cd catalogizer-androidtv && ./gradlew test`

---

### Phase 3: Dead Code Wiring & Feature Completion

**Goal**: Connect all disconnected features, remove dead code, complete partial implementations.

#### Step 3.1: Wire Unused React Hooks or Remove (DC-001..005)
- **Decision per hook**:
  - `useDebounce` -- WIRE: useful for search inputs. Connect to MediaSearch component
  - `useLazyImage` -- WIRE: useful for media grid. Connect to MediaGrid component
  - `usePlayerState` -- REMOVE: superseded by current player implementation
  - `usePlaylistReorder` -- WIRE: connect to PlaylistView drag-drop
  - `useVirtualScroll` -- WIRE: connect to large media lists for performance

#### Step 3.2: Complete Collection Features (DC-006..009)
- `CollectionSharing.tsx` -- Add route in App.tsx; implement sharing API endpoint in catalog-api
- `CollectionRealTime.tsx` -- Import in Collections page; connect to WebSocket events
- `PerformanceOptimizer.tsx` -- Wire as dev-mode overlay component
- `ExternalIntegrations.tsx` -- Complete implementation; add settings UI

#### Step 3.3: Replace Mock APIs with Real Endpoints (DC-010..011)
- `conversionApi.ts` -- Implement `/api/v1/conversion` endpoints in catalog-api
- `mockCollectionsApi.ts` -- Remove mock; collections API already exists, just needs wiring
- **Verify**: Zero 404 responses; zero mock data fallbacks

#### Step 3.4: Clean Up Unused Go Code (DC-012..017)
- `advanced_rate_limiter.go` -- DELETE (functionality covered by per-route rate limiting)
- `enhanced_rate_limiter.go` -- DELETE (duplicate)
- `redis_rate_limiter.go` -- KEEP but add integration path (Redis rate limiting for production)
- `NewStubHandler()` -- DELETE deprecated function
- Config fields -- Remove unparsed fields OR wire them to actual usage
- `.env.example` -- Remove unused AI/Vision vars OR add implementation stubs

#### Step 3.5: Implement Episode Navigation (UC-001..002)
- **Backend**: Add `/api/v1/entities/:id/episodes` endpoint returning ordered episodes for a series
- **Backend**: Add `GetAdjacentEpisodes(entityID)` to `MediaItemRepository`
- **Android TV**: Wire `playNextEpisode()`/`playPreviousEpisode()` to real API calls
- **Test**: E2E test for episode navigation flow

#### Step 3.6: Verify No Dead Code Remains
- Run: `go vet ./...` (catches unused exports in main package)
- Run: `npm run lint` with no-unused-vars rule
- Manual review of all changes

---

### Phase 4: Test Coverage -- Go & TypeScript

**Goal**: Achieve 100% test coverage for catalog-api, catalog-web, API client, and all TS submodules.

#### Step 4.1: Fix 6 Skipped Tests in catalog-web (TC-010)
- Files: CollectionTemplates, MediaFilters, MediaPlayer, MemoCache, PlaylistPlayer, usePlayerState
- For each: investigate skip reason, implement proper test, remove `.skip`

#### Step 4.2: Complete API Client Tests (TC-005)
- Add tests for: `index.ts` (exports), `http.ts` (retry logic, error handling), `websocket.ts` (connection, reconnection, message handling)
- Target: 100% coverage

#### Step 4.3: Complete Auth-Context-React Tests (TC-006)
- Add tests for: AuthProvider state management, token refresh, logout cleanup, permission resolution
- Target: 100% coverage (3 source files, need 3 test files)

#### Step 4.4: Complete Media-Player-React Tests (TC-007)
- Add tests for: player controls, seek, volume, fullscreen, error states
- Target: 100% coverage (5 source files, need 5 test files)

#### Step 4.5: Complete Collection-Manager-React Tests (TC-008)
- Add tests for: collection CRUD, item ordering, filtering, sharing
- Target: 100% coverage (6 source files, need 6 test files)

#### Step 4.6: Complete Dashboard-Analytics-React Tests (TC-009)
- Add tests for: chart rendering, data transformation, time range selection, export
- Target: 100% coverage (6 source files, need 6 test files)

#### Step 4.7: Complete Installer Wizard Step Tests (TC-011)
- All 10 configuration step components need tests:
  - ConfigurationManagementStep, FTPConfigurationStep, LocalConfigurationStep
  - NFSConfigurationStep, NetworkScanStep, ProtocolSelectionStep
  - SMBConfigurationStep, SummaryStep, WebDAVConfigurationStep, WelcomeStep
- Test: validation logic, state transitions, error handling for each

#### Step 4.8: Fill catalog-web Remaining Gaps
- Add tests for: HistoryDrawer, ProgressBadge, playbackApi, playback types
- Target: 100% component coverage

#### Step 4.9: Fill catalog-api Remaining Gaps (TC-012)
- Audit the 3 untested packages; add tests
- Target: 100% package coverage

#### Step 4.10: Run Full Coverage Reports
- Go: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Web: `npm run test:coverage`
- Review coverage percentages; iterate until 100%

---

### Phase 5: Test Coverage -- Rust/Tauri

**Goal**: Add comprehensive Rust test coverage to catalogizer-desktop and installer-wizard.

#### Step 5.1: catalogizer-desktop Rust Tests (TC-001)
- Add `#[cfg(test)]` modules to all source files in `src-tauri/src/`
- Test categories:
  - **IPC commands**: All Tauri command handlers
  - **Configuration**: Config loading, validation, defaults
  - **File operations**: File system interactions
  - **State management**: App state transitions
  - **Error handling**: All error paths
- Target: 46 source files, minimum 46 test modules

#### Step 5.2: installer-wizard Rust Tests (TC-002)
- Add `#[cfg(test)]` modules to all source files in `src-tauri/src/`
- Test categories:
  - **Wizard flow**: Step progression, back/forward, validation
  - **Protocol configuration**: SMB/FTP/NFS/WebDAV/Local config validation
  - **System detection**: Network scan, service discovery
  - **Installation**: Config writing, service registration
- Target: 47 source files, minimum 47 test modules

#### Step 5.3: Integration Tests
- Add `tests/` directory in each Tauri project
- Test IPC round-trips with mock frontend
- Test configuration persistence

#### Step 5.4: Run Rust Tests
- `cd catalogizer-desktop/src-tauri && cargo test`
- `cd installer-wizard/src-tauri && cargo test`
- Verify zero failures

---

### Phase 6: Test Coverage -- Android & Android TV

**Goal**: Raise Android from 5.3% to 100% and Android TV from 10.8% to 100%.

#### Step 6.1: catalogizer-android Unit Tests (TC-003)
- **Priority 1 -- ViewModels**: All ViewModel classes need tests with MockK
- **Priority 2 -- Repositories**: All Repository classes with mocked DAOs and APIs
- **Priority 3 -- Services**: SyncManager, AuthManager, MediaPlayer service
- **Priority 4 -- Utils**: All utility classes and extension functions
- **Priority 5 -- UI**: Compose UI tests with ComposeTestRule
- Target: ~200 new test files covering 1053 source files

#### Step 6.2: catalogizer-android Instrumented Tests
- Add `androidTest/` suite for:
  - Room database operations (all DAOs)
  - DataStore read/write
  - Network integration (with MockWebServer)
  - Navigation flow
  - Compose UI rendering

#### Step 6.3: catalogizer-androidtv Unit Tests (TC-004)
- **Priority 1 -- TV-specific**: TvChannelRepository, WatchNextManager, ChannelProgramMapper
- **Priority 2 -- ViewModels**: All ViewModel classes
- **Priority 3 -- Repositories**: All Repository classes
- **Priority 4 -- Player**: VLCPlayerActivity, player controls
- **Priority 5 -- Navigation**: DPAD navigation, focus management
- Target: ~100 new test files covering 436 source files

#### Step 6.4: catalogizer-androidtv Instrumented Tests
- Add `androidTest/` directory (currently missing)
- Test: TV channel operations, deep link handling, DPAD navigation, player integration

#### Step 6.5: Run Android Test Suites
- `cd catalogizer-android && ./gradlew test`
- `cd catalogizer-androidtv && ./gradlew test`
- Target: zero failures across all tests

---

### Phase 7: Security Scanning & Hardening

**Goal**: Complete security scanning infrastructure; run all scanners; fix all findings.

#### Step 7.1: Add Hadolint Service (SS-002)
- **File**: `docker-compose.security.yml`
- Add hadolint service with profile `hadolint`
- Run against all Dockerfiles
- Fix all findings

#### Step 7.2: Create Custom Semgrep Rules (SS-001)
- **File**: `.semgrep.yml`
- Rules for:
  - Go: SQL injection, command injection, SSRF, path traversal
  - TypeScript: XSS, prototype pollution, unsafe eval
  - Kotlin: intent injection, WebView security
- Run: `podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner`

#### Step 7.3: Integrate govulncheck (SS-003)
- **File**: `scripts/security-scan.sh`
- Add `govulncheck ./...` to main security scan script
- Add to `scripts/install-security-tools.sh`

#### Step 7.4: Fix SonarQube Profile (SS-004)
- **File**: `docker-compose.security.yml`
- Move sonarqube + sonarqube-db to `sonarqube` profile
- Run: `./scripts/run-sonarqube-scan.sh`
- Analyze findings; fix all code quality issues

#### Step 7.5: Document Snyk Setup (SS-005)
- **File**: `.env.example`
- Add `SNYK_TOKEN=YOUR_SNYK_TOKEN_HERE`
- Run: Snyk scans via compose profile
- Fix all findings

#### Step 7.6: Run Comprehensive Security Scan
- `./scripts/security-scan.sh`
- govulncheck: 0 vulnerabilities
- npm audit: 0 critical/high
- Semgrep: 0 high/critical findings
- Gosec: 0 high findings
- Trivy: 0 critical/high CVEs
- Generate consolidated report to `docs/security/`

---

### Phase 8: Stress, DDoS, Benchmarking Tests

**Goal**: Complete Constitution Article V categories 5 (stress), 7 (DDoS/rate-limit), 8 (benchmarking).

#### Step 8.1: Expand Stress Tests (Category 5)
- **Location**: `tests/k6/` and `catalog-api/tests/stress/`
- New scenarios:
  - Concurrent file scanning (100+ simultaneous scans)
  - Large payload handling (1GB+ files)
  - Long session stability (24h continuous operation)
  - Database connection pool saturation
  - WebSocket connection storm (1000+ simultaneous)
  - Memory pressure under sustained load
  - Disk I/O saturation during bulk operations

#### Step 8.2: DDoS/Rate-Limit Tests (Category 7)
- **Location**: `tests/k6/ddos_ratelimit_test.js` (extend)
- New scenarios:
  - HTTP flood (10K+ req/s sustained)
  - Slowloris attack simulation
  - Connection exhaustion (max open connections)
  - Auth endpoint brute force (verify rate limiter)
  - API key rotation under load
  - Circuit breaker activation and recovery
  - Rate limit header verification (`X-RateLimit-*`)
  - Burst traffic patterns (spike from 0 to 1000 req/s)
- Verify: system recovers within 30s after attack stops

#### Step 8.3: Benchmarking Tests (Category 8)
- **Location**: `tests/benchmarks/`
- Establish baselines for:
  - API response latency (p50, p95, p99) per endpoint
  - File scan throughput (files/second)
  - Database query latency (reads, writes, aggregations)
  - WebSocket message latency
  - Memory usage per concurrent user
  - Startup time (cold start, warm start)
  - Build time per component
- Implement regression detection:
  - Compare against stored baselines
  - Fail if p99 latency increases >10%
  - Fail if throughput decreases >5%
- Output: `reports/benchmarks/baseline-YYYY-MM-DD.json`

#### Step 8.4: Go Benchmark Tests
- Add `_bench_test.go` files for all hot paths:
  - `BenchmarkMediaDetection`
  - `BenchmarkTitleParsing`
  - `BenchmarkDatabaseQuery`
  - `BenchmarkJWTValidation`
  - `BenchmarkRateLimiter`
  - `BenchmarkWebSocketBroadcast`
  - `BenchmarkCacheOperations`
  - `BenchmarkFileSystemListing`

---

### Phase 9: Challenge & Bank Expansion

**Goal**: Cover all 10 Constitution Article V categories in challenge banks.

#### Step 9.1: Audit Current Challenge Coverage
- Current: 507 challenges (50 original + 174 userflow + 15 module + others)
- Map each challenge to its Article V category
- Identify categories with <100% coverage

#### Step 9.2: Add Security Challenges (Category 6)
- SQL injection attempts on all text inputs
- XSS payload injection in all string fields
- SSRF via URL-accepting endpoints
- Authentication bypass attempts
- Authorization escalation attempts
- JWT manipulation challenges
- CSRF verification challenges
- Path traversal on file endpoints

#### Step 9.3: Add DDoS/Rate-Limit Challenges (Category 7)
- Rate limiter verification per endpoint
- Circuit breaker activation challenges
- Connection pool exhaustion + recovery
- Auth brute force resistance
- WebSocket flood resistance

#### Step 9.4: Add Benchmarking Challenges (Category 8)
- Response time SLA challenges (p95 < 500ms)
- Throughput minimum challenges (>100 req/s sustained)
- Memory ceiling challenges (<500MB under normal load)
- Startup time challenges (<5s cold start)

#### Step 9.5: Expand HelixQA Banks
- **Full QA banks** (extend existing):
  - `full-qa-api.yaml`: Add security, DDoS, benchmarking entries
  - `full-qa-web.yaml`: Add accessibility, performance, security entries
  - `full-qa-androidtv.yaml`: Add DPAD edge cases, channel management, deep link entries
  - `full-qa-android.yaml`: Add offline mode, sync, biometric entries
- **New bank**: `fixes-validation.yaml` -- regression tests for all bugs fixed in this plan
- **New bank**: `full-qa-desktop.yaml` -- desktop-specific QA entries
- **New bank**: `full-qa-wizard.yaml` -- wizard-specific QA entries

#### Step 9.6: Convert Banks to JSON
```bash
for f in HelixQA/banks/*.yaml; do
  python3 -c "import yaml,json; json.dump(yaml.safe_load(open('$f')), open('${f%.yaml}.json','w'))"
done
```

---

### Phase 10: Lazy Loading, Semaphores, Non-Blocking Mechanisms

**Goal**: Optimize responsiveness across all components.

#### Step 10.1: Go Backend -- Lazy Initialization
- Extend `LazyServiceRegistry` usage in `main.go`
- Services to lazy-load:
  - Metadata providers (TMDB, OMDB, OpenLibrary, MusicBrainz) -- load on first request
  - Redis cache -- connect on first cache operation
  - WebSocket handler -- initialize on first connection
  - Subtitle service -- already lazy, verify
- Use `digital.vasic.lazy` module for generic lazy loading

#### Step 10.2: Go Backend -- Semaphore Controls
- Use `digital.vasic.concurrency` semaphore for:
  - Concurrent file scanning (cap at `GOMAXPROCS`)
  - Concurrent metadata fetches (cap at 5)
  - Concurrent WebSocket broadcasts (cap at runtime.NumCPU())
  - Database write operations (cap at MaxOpen/2)
- Add metrics: semaphore utilization, wait time, rejection count

#### Step 10.3: Go Backend -- Non-Blocking Patterns
- All I/O operations must use `context.Context` with timeouts
- Replace any synchronous HTTP calls with async patterns where appropriate
- Ensure WebSocket broadcast is non-blocking (use buffered channels)
- File scanning: use worker pool with bounded queue

#### Step 10.4: React Frontend -- Lazy Loading
- React.lazy() for all route-level components (pages)
- Dynamic imports for heavy components:
  - MediaPlayer -- load only when playback requested
  - Charts/Analytics -- load only on dashboard page
  - CollectionManager -- load only when accessed
  - UploadManager -- load only when upload initiated
- Verify Vite chunk splitting aligns with lazy boundaries

#### Step 10.5: React Frontend -- Virtual Scrolling
- Wire `useVirtualScroll` hook to:
  - Media grid (large collections)
  - File browser listing
  - Search results
  - Collection item lists
- Use `react-window` or `@tanstack/virtual` for implementation

#### Step 10.6: Android -- Lazy Loading
- Verify all Hilt modules use `@Provides` with lazy initialization
- Paging 3 already implements lazy loading for lists
- Lazy image loading via Coil (already configured)
- Lazy Room database initialization (already singleton pattern)

#### Step 10.7: Monitoring & Metrics
- Add lazy init metrics to Prometheus:
  - `lazy_init_duration_seconds` -- time to initialize each lazy service
  - `semaphore_utilization_ratio` -- current/max for each semaphore
  - `semaphore_wait_duration_seconds` -- time spent waiting for semaphore
  - `nonblocking_queue_depth` -- current queue depth for async operations
- Add Grafana dashboard panel for these metrics

#### Step 10.8: Performance Validation
- Run k6 load tests before and after optimizations
- Compare: response latency, throughput, memory usage
- Document improvements in performance report

---

### Phase 11: Documentation Completion

**Goal**: Complete all documentation to nano detail. No gaps.

#### Step 11.1: Create Missing AGENTS.md Files (DG-001)
- Create for: catalog-web, catalogizer-android, catalogizer-androidtv, catalogizer-desktop, catalogizer-api-client, installer-wizard
- Content: component architecture, AI agent constraints, build/test commands, key files

#### Step 11.2: Create Missing API_REFERENCE.md Files (DG-007)
- Create for all components without one
- Content: all public APIs, parameters, return types, error codes, examples

#### Step 11.3: Create Missing CONTRIBUTING.md Files (DG-007)
- Create for all components without one
- Content: setup instructions, coding standards, PR process, test requirements

#### Step 11.4: Create SQL Migration Reference (DG-004)
- **File**: `docs/database/SQL_MIGRATION_REFERENCE.md`
- Content: all 15 migrations documented with schema changes, rationale, rollback procedures
- Include ERD updates per migration

#### Step 11.5: Create Performance Tuning Guide (DG-005)
- **File**: `docs/guides/PERFORMANCE_TUNING.md`
- Content: database tuning, connection pool sizing, cache configuration, resource limits, monitoring setup

#### Step 11.6: Consolidate Plans (DG-006)
- Archive old plans to `docs/plans/archive/`
- THIS document becomes the single authoritative plan
- Update status references to point here

#### Step 11.7: Update All Existing Documentation
- Review and update every file in `docs/`:
  - USER_GUIDE.md -- add new features from v2.3.0
  - INSTALLATION_GUIDE.md -- update for current versions
  - ADMIN_GUIDE.md -- add new admin endpoints
  - DEPLOYMENT_GUIDE.md -- update container instructions
  - TESTING_GUIDE.md -- add all 10 Article V categories
  - ARCHITECTURE_DIAGRAMS.md -- add new diagrams for lazy loading, semaphores
  - DATA_DICTIONARY.md -- add any new tables/fields
  - CHANGELOG.md -- update with all changes from this plan

#### Step 11.8: Update All Mermaid Diagrams
- Review and update all 20+ `.mmd` files
- Add new diagrams:
  - `lazy-initialization-flow.mmd`
  - `semaphore-control-flow.mmd`
  - `security-scanning-pipeline.mmd`
  - `test-coverage-matrix.mmd`

#### Step 11.9: Update Component Documentation
- Every submodule CLAUDE.md -- verify accuracy
- Every submodule ARCHITECTURE.md -- verify accuracy
- Every submodule README.md -- verify setup instructions work

---

### Phase 12: Video Courses & Training

**Goal**: Comprehensive video course scripts covering all components.

#### Step 12.1: Core Video Course Scripts
- **File**: `docs/video-course/catalog-api-course.md`
  - Module 1: Architecture overview (handlers, services, repositories)
  - Module 2: Database layer (dual-dialect, migrations, transactions)
  - Module 3: Media detection pipeline (scanner, detector, analyzer, providers)
  - Module 4: Authentication & authorization (JWT, roles, permissions)
  - Module 5: WebSocket real-time events
  - Module 6: File system protocols (SMB, FTP, NFS, WebDAV, Local)
  - Module 7: Challenge system
  - Module 8: Monitoring & metrics (Prometheus, Grafana)
  - Module 9: Security hardening
  - Module 10: Performance optimization (lazy loading, semaphores)

#### Step 12.2: Frontend Video Course
- **File**: `docs/video-course/catalog-web-course.md`
  - Module 1: React architecture (providers, routing, state)
  - Module 2: Media browsing & search
  - Module 3: Collection management
  - Module 4: Media playback
  - Module 5: Settings & configuration
  - Module 6: WebSocket integration
  - Module 7: Performance (virtual scroll, lazy loading, code splitting)

#### Step 12.3: Mobile Video Courses
- **File**: `docs/video-course/android-course.md`
  - Module 1: MVVM architecture
  - Module 2: Compose UI
  - Module 3: Offline-first with Room
  - Module 4: Media playback
  - Module 5: Sync & background work

- **File**: `docs/video-course/androidtv-course.md`
  - Module 1: Leanback architecture
  - Module 2: DPAD navigation
  - Module 3: Home screen channels
  - Module 4: Media playback (VLC)
  - Module 5: Deep linking

#### Step 12.4: Desktop & Wizard Video Courses
- **File**: `docs/video-course/desktop-course.md`
  - Module 1: Tauri architecture (Rust + React)
  - Module 2: IPC commands
  - Module 3: VLC integration
  - Module 4: System tray & notifications

- **File**: `docs/video-course/wizard-course.md`
  - Module 1: Installation wizard flow
  - Module 2: Protocol configuration
  - Module 3: Network scanning
  - Module 4: Summary & deployment

#### Step 12.5: Update Existing Video Courses
- Update: DocProcessor, LLMOrchestrator, VisionEngine, HelixQA, Containers courses
- Add modules for new features and changes from this plan

---

### Phase 13: Website Update & Content Migration

**Goal**: Comprehensive, up-to-date website with all documentation accessible.

#### Step 13.1: Website Structure Expansion
- **New pages**:
  - `architecture.md` -- System architecture with Mermaid diagrams
  - `security.md` -- Security posture and scanning
  - `api-reference.md` -- Link to OpenAPI spec
  - `testing.md` -- Testing strategy and Article V compliance
  - `platforms.md` -- Platform-specific guides (web, desktop, android, tv)
  - `contributing.md` -- How to contribute
  - `roadmap.md` -- Project roadmap

#### Step 13.2: Developer Guide Expansion
- **New developer guide pages**:
  - `database.md` -- Database architecture, migrations, ERD
  - `protocols.md` -- File system protocol integration
  - `media-pipeline.md` -- Media detection and enrichment
  - `websockets.md` -- Real-time event system
  - `challenges.md` -- Challenge framework usage
  - `deployment.md` -- Deployment procedures
  - `performance.md` -- Performance optimization guide

#### Step 13.3: Getting Started Expansion
- **New getting started pages**:
  - `quick-start.md` -- 5-minute setup guide
  - `configuration.md` -- Detailed configuration reference
  - `first-scan.md` -- Running your first media scan
  - `troubleshooting.md` -- Common issues and solutions

#### Step 13.4: Content Migration
- Migrate relevant content from `docs/` to website
- Ensure website links to detailed docs where appropriate
- Add download links for all platforms

#### Step 13.5: Website Verification
- Build website: `cd Website && npm run build`
- Verify all pages render correctly
- Verify all internal links work
- Verify all diagrams render

---

### Phase 14: Version Alignment & Release Build

**Goal**: Align all versions, build all components, final QA pass.

#### Step 14.1: Version Alignment (BC-003, BC-005..008)
- Update all `package.json` files to version 2.4.0
- Update all `build.gradle.kts` files to version 2.4.0
- Update `versions.json` to 2.4.0 build 23
- Bump TS submodule versions from 0.1.0 to 1.0.0
- Align Go submodule versions to 1.25.7
- Standardize API client name to `@catalogizer/api-client`

#### Step 14.2: Pre-Build Verification
- Run all test suites across all components
- Run security scans (all tools)
- Verify zero warnings, zero errors
- Verify zero console errors in browser
- Verify zero failed network requests

#### Step 14.3: Container Release Build
```bash
./scripts/release-build.sh --container --force
```
- All 7 components must build successfully
- All build artifacts in `releases/`

#### Step 14.4: Final QA Campaign
- Run HelixQA orchestrator across all platforms:
```bash
./scripts/helixqa-orchestrator.sh
```
- Complete the mandatory retesting loop until clean
- Generate final report to `docs/reports/qa-sessions/qa-session-2026-04-14/`

#### Step 14.5: Git Commit & Push
- Single comprehensive commit with all changes
- Push to all 6 remotes
- Update submodules

---

## 4. Test Coverage Matrix

### Constitution Article V -- 10 Categories x 10 Platforms

| Category | catalog-api | catalog-web | desktop | wizard | android | androidtv | api-client | Go subs | TS subs | HelixQA |
|----------|------------|-------------|---------|--------|---------|-----------|------------|---------|---------|---------|
| 1. Unit | 99% | 95% | 92%R/0%Rust | 100%R/0%Rust | 5.3% | 10.8% | 77.8% | 100%+ | 60-100% | N/A |
| 2. Integration | Partial | Partial | None | None | None | None | None | Some | None | N/A |
| 3. E2E | Via challenges | Playwright | None | None | None | None | None | N/A | N/A | HelixQA |
| 4. Full automation | Challenges | Playwright CI | None | None | None | None | None | CI-ready | CI-ready | Autonomous |
| 5. Stress | k6 scripts | None | None | None | None | None | None | None | None | Edge banks |
| 6. Security | govulncheck | npm audit | None | None | None | None | None | govulncheck | npm audit | Security bank |
| 7. DDoS/Rate | k6 ddos test | None | N/A | N/A | N/A | N/A | N/A | N/A | N/A | None |
| 8. Benchmarking | None | None | None | None | None | None | None | None | None | None |
| 9. Challenges | 507 reg'd | Via userflow | Via userflow | Via userflow | Via userflow | Via userflow | None | MOD-* | None | N/A |
| 10. HelixQA | API bank | Web bank | None | None | Android bank | AndroidTV bank | N/A | N/A | N/A | Self-test |

**Legend**: "None" = gap to fill in this plan. Greyed cells = not applicable.

### Target After Plan Execution

All cells should show "100%" or "Complete" or "N/A" (genuinely not applicable).

---

## 5. Documentation Deliverables

### New Documents to Create

| Document | Location | Phase |
|----------|----------|-------|
| 6x AGENTS.md | Component roots | 11 |
| 6x API_REFERENCE.md | Component roots | 11 |
| 6x CONTRIBUTING.md | Component roots | 11 |
| SQL_MIGRATION_REFERENCE.md | `docs/database/` | 11 |
| PERFORMANCE_TUNING.md | `docs/guides/` | 11 |
| catalog-api-course.md | `docs/video-course/` | 12 |
| catalog-web-course.md | `docs/video-course/` | 12 |
| android-course.md | `docs/video-course/` | 12 |
| androidtv-course.md | `docs/video-course/` | 12 |
| desktop-course.md | `docs/video-course/` | 12 |
| wizard-course.md | `docs/video-course/` | 12 |
| 7x website pages | `Website/` | 13 |
| 7x developer guide pages | `Website/docs/developer-guide/` | 13 |
| 4x getting started pages | `Website/docs/getting-started/` | 13 |
| 4x new Mermaid diagrams | `docs/architecture/` | 11 |

### Documents to Update

| Document | Updates Needed | Phase |
|----------|---------------|-------|
| CLAUDE.md | Add Phase 10 patterns, new endpoints | 11 |
| AGENTS.md | Add new constraints from safety fixes | 11 |
| USER_GUIDE.md | Add v2.4.0 features | 11 |
| INSTALLATION_GUIDE.md | Update versions | 11 |
| ADMIN_GUIDE.md | New admin endpoints | 11 |
| DEPLOYMENT_GUIDE.md | Container updates | 11 |
| TESTING_GUIDE.md | All 10 Article V categories | 11 |
| DATA_DICTIONARY.md | New tables/fields | 11 |
| CHANGELOG.md | v2.4.0 changes | 14 |
| ARCHITECTURE_DIAGRAMS.md | New diagrams | 11 |
| versions.json | 2.4.0 build 23 | 14 |
| 5x existing video courses | New modules | 12 |
| 20+ Mermaid diagrams | Accuracy review | 11 |
| 46x CLAUDE.md files | Accuracy review | 11 |

---

## 6. Challenge & Bank Expansion

### New Challenges to Register

| Category | Count | Description |
|----------|-------|-------------|
| Security | 20 | SQL injection, XSS, SSRF, auth bypass, JWT manipulation |
| DDoS/Rate-limit | 10 | Flood resistance, circuit breaker, burst handling |
| Benchmarking | 15 | Latency SLA, throughput minimum, memory ceiling |
| Episode navigation | 5 | Next/previous episode, series listing, ordering |
| Collection sharing | 5 | Share, unshare, permission check, link generation |
| Conversion | 5 | File format conversion endpoints |
| Desktop E2E | 10 | Desktop application flows |
| Wizard E2E | 10 | Installation wizard flows |

**Total new challenges**: ~80
**New total**: ~587 registered challenges

### New Bank Entries

| Bank | New Entries | Description |
|------|------------|-------------|
| `fixes-validation.yaml` | 29 | One per concurrency fix from Phase 1-2 |
| `full-qa-desktop.yaml` | 50 | Desktop application QA |
| `full-qa-wizard.yaml` | 40 | Installation wizard QA |
| `security-comprehensive.yaml` | 30 | Deep security testing |
| `ddos-ratelimit-comprehensive.yaml` | 20 | DDoS and rate limit testing |
| `benchmarking-baselines.yaml` | 15 | Performance regression tests |

**Total new bank entries**: ~184
**New total**: ~1,784 HelixQA test cases

---

## 7. Risk Register

| Risk | Impact | Mitigation |
|------|--------|------------|
| Android test expansion breaks existing tests | HIGH | Run existing tests after each batch of new tests |
| Rust test additions require Tauri test infrastructure | MEDIUM | Use `#[cfg(test)]` modules with mocked IPC |
| Concurrency fixes introduce regressions | HIGH | Run with `-race` flag; add fixes-validation bank |
| Security scan findings require dependency updates | MEDIUM | Pin versions; test after each update |
| Version alignment breaks cross-component compatibility | HIGH | Build all components after version bump |
| Docker Go version change affects build cache | LOW | Force rebuild with `--no-cache` |
| Website build fails with new content | LOW | Build and test locally before commit |
| k6 stress tests exceed host resource limits | MEDIUM | Enforce 30-40% resource cap per CLAUDE.md |
| Challenge count increase slows RunAll | MEDIUM | Progress-based liveness; 5min stale threshold |
| Documentation volume overwhelms review | LOW | Review per-phase, not all at once |

---

## Appendix A: File Change Summary

### Files to Create (~90)
- 6 AGENTS.md, 6 API_REFERENCE.md, 6 CONTRIBUTING.md
- 6 video course scripts
- 18 website pages
- 4 Mermaid diagrams
- ~50 new test files (Rust, Android, TS)
- `.semgrep.yml`
- `docs/database/SQL_MIGRATION_REFERENCE.md`
- `docs/guides/PERFORMANCE_TUNING.md`

### Files to Modify (~200)
- All concurrency fix files (~15 Go, ~5 TS, ~5 Kotlin)
- All dead code cleanup files (~10 Go, ~5 TS)
- All documentation updates (~50 markdown files)
- All version files (versions.json, package.json x5, build.gradle.kts x2)
- Docker/Dockerfile.builder
- docker-compose.security.yml
- scripts/security-scan.sh
- All challenge bank files (~25 YAML/JSON)
- All existing test files with .skip markers (~6)

### Files to Delete (~5)
- `middleware/advanced_rate_limiter.go`
- `middleware/enhanced_rate_limiter.go`
- `handlers/stub_handler.go` (NewStubHandler function only)
- `hooks/usePlayerState.tsx` (superseded)

---

## Appendix B: Execution Order Dependencies

```
Phase 1 (Critical Safety) ──────────────┐
                                         │
Phase 2 (Concurrency Hardening) ─────────┤
                                         │
Phase 3 (Dead Code / Feature Completion) ┤
                                         ├──> Phase 7 (Security Scanning)
Phase 4 (Tests: Go/TS) ─────────────────┤        │
                                         │        ├──> Phase 9 (Challenges)
Phase 5 (Tests: Rust) ──────────────────┤        │         │
                                         │        │         ├──> Phase 14 (Release)
Phase 6 (Tests: Android) ───────────────┘        │         │
                                                  │         │
Phase 8 (Stress/DDoS/Bench) ─────────────────────┘         │
                                                             │
Phase 10 (Lazy/Semaphore/NonBlock) ──────────────────────────┤
                                                              │
Phase 11 (Documentation) ────────────────────────────────────┤
                                                              │
Phase 12 (Video Courses) ────────────────────────────────────┤
                                                              │
Phase 13 (Website) ──────────────────────────────────────────┘
```

Phases 1-6 are sequential (each builds on prior fixes).
Phases 7-13 can partially overlap but should complete before Phase 14.
Phase 14 is the final gate.

---

**End of Plan**
**Author**: Claude Code (Opus 4.6)
**Reviewed**: Pending human review

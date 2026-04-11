# Comprehensive Completion: Audit Report & Master Roadmap

> **For agentic workers:** REQUIRED SUB-SKILL: This is a master roadmap. Each phase has its own bite-sized sub-plan in `docs/plans/2026-04-11-phase-N-*.md`. Use superpowers:subagent-driven-development or superpowers:executing-plans to execute each phase. Phase 1 sub-plan exists; Phases 2–12 will be generated on demand once their predecessors complete.

**Goal:** Eliminate every unfinished, broken, disabled, undocumented, untested, leaky, or unsafe artifact across the entire Catalogizer monorepo (catalog-api + catalog-web + Tauri + Android + Android TV + 41 submodules), bring test coverage to its theoretical maximum, complete all documentation surfaces (manuals, video courses, website, diagrams, SQL), and prove correctness via stress / integration / security / performance test campaigns — without breaking any existing working functionality.

**Architecture:** Phased completion. Each phase produces independently shippable improvements with green tests + green challenges before the next phase begins. Phases 1–3 are remediation (fix what's broken), Phases 4–7 are hardening (test + secure + observe), Phases 8–10 are documentation (write everything down), Phases 11–12 are validation (prove it).

**Tech Stack:** Go 1.25 + Gin, React 18 + TS + Vite + Vitest, Tauri + Rust, Kotlin + Compose, PostgreSQL/SQLite (dual dialect), Podman, k6, Snyk + SonarQube + Semgrep + govulncheck, HelixQA (LLM-vision-driven autonomous testing), VitePress.

**Constraints (verbatim, non-negotiable):**
- Zero unfinished work — no TODO/FIXME/NOTE-as-deferral, no stubs, no skips without justification.
- All operations local-user (no `sudo`, no `root`, no interactive prompts).
- Containers via Podman. Tests run with `GOMAXPROCS=3 -p 2 -parallel 2`. Container budget: 4 CPUs / 8 GB RAM total.
- Universal solution principle: never modify the app under test to make tests work — fix the test infra.
- HelixQA is the sole UI/UX testing tool, fully LLM-vision-driven, no hardcoded coordinates.
- HTTP/3 + Brotli mandatory in production.
- Changes must be rock-solid: every fix needs a regression test in `banks/fixes-validation.yaml` AND a Go/TS/Kotlin unit test that proves the fix.
- All work respects GitSpec constitution + AGENTS.md + CLAUDE.md.

---

## Section A — Audit Report (current state, real findings)

This section catalogs every concrete unfinished/broken/risky/missing artifact found by parallel research agents on 2026-04-11. Each finding is a remediation target consumed by a phase task.

### A.1 — Code Completeness Gaps (stub / no-op / placeholder implementations)

| ID | File:line | Symbol | Severity | Description |
|---|---|---|---|---|
| CC-01 | `catalog-api/services/reporting_service.go:~210` | `calculateGrowthRate()` | High | Returns hardcoded `0.0` with NOTE deferring to "future enhancement when multi-period analytics are implemented" |
| CC-02 | `catalog-api/services/reporting_service.go:~280` | `calculatePerformanceMetrics()` | High | Returns empty `models.PerformanceMetrics{}` — requires middleware instrumentation that doesn't exist yet |
| CC-03 | `catalog-api/services/reporting_service.go:~305` | `calculateSystemLoad()` | High | Returns empty `models.SystemLoad{}` — requires metrics-collection daemon |
| CC-04 | `catalog-api/services/reporting_service.go:~310` | `calculateErrorRates()` | High | Returns empty `models.ErrorRates{}` — requires error-logging middleware |
| CC-05 | `catalog-api/services/reporting_service.go:~190` | vulnerability tracking | High | Returns hardcoded `0` — requires `security_events` table |
| CC-06 | `catalog-api/services/reporting_service.go:~290` | response-time tracking | High | Returns empty `models.ResponseTimes{}` — requires HTTP middleware |
| CC-07 | `catalog-api/internal/services/playback_position_service.go:522-527` | `SyncAcrossDevices()` | High | No-op stub returning `nil` immediately — feature is shipped but does nothing |
| CC-08 | `catalog-api/internal/services/cache_service.go:942-946` | `Warmup()` | High | Logs "Starting cache warmup" then returns `nil` without doing any warmup |
| CC-09 | `catalog-api/internal/services/media_recognition_service.go:1230-1275` | 5 cover-art / metadata fetch methods | High | All 5 (`searchLocalCoverArt`, `fetchTMDBCoverArt`, `fetchMusicBrainzCoverArt`, `fetchTMDBMetadata`, `fetchMusicBrainzMetadata`) return empty results — comment: "In production, this would call the TMDB/MusicBrainz API" |
| CC-10 | `catalog-api/internal/services/smb.go:423` | `Connect()` | High | Documented stub — only checks hostname presence in config; doesn't establish an SMB session |
| CC-11 | `catalog-api/challenges/ch044_websocket_latency.go:102` | challenge body | Medium | Returns `"WebSocket not available; challenge passes as stub"` instead of failing or skipping with reason |
| CC-12 | `catalog-api/challenges/ch081_088.go:96` | challenge body | Medium | Same stub-pass pattern as CC-11 |
| CC-13 | `catalog-api/challenges/browsing_web_app.go:317` | media-stats validation | Medium | Returns error `"endpoint was a stub?"` — indicates upstream unfinished work in `/api/v1/media/stats` |
| CC-14 | `LLMProvider/pkg/providers/gemini/gemini_cli_stub.go` | all methods | Low | Intentional fallback when CLI module unavailable — keep but document |
| CC-15 | `LLMProvider/pkg/providers/junie/junie_cli_stub.go` | all methods | Low | Same as CC-14 |

### A.2 — Concurrency Safety (memory leaks, races, deadlocks, panics)

| ID | File:line | Severity | Description |
|---|---|---|---|
| CS-01 | `catalog-api/smb/types.go:170` | **CRITICAL** | Reading from a closed channel (`<-p.cleanupDone`) after `close(p.cleanupDone)` at line 166 — will panic. Cleanup loop also has `case <-p.cleanupDone` from same closed channel — race + panic |
| CS-02 | `catalog-api/smb/types.go:156-170` | High | Potential AB-BA deadlock: `StopCleanup()` holds `p.mu.Lock()` while closing `p.cleanupDone`, then waits on the channel. Cleanup loop calls `cleanupIdleConnections()` which acquires `p.mu` |
| CS-03 | `catalog-api/middleware/request.go:45-78` | High | Rate limiter cleanup goroutine has no shutdown mechanism — blocks on `ticker.C` until process exit |
| CS-04 | `catalog-api/middleware/advanced_rate_limiter.go:122-133` | High | Cleanup goroutine has no cancellation signal — relies on process exit |
| CS-05 | `catalog-api/handlers/websocket_handler.go:88-106` | High | `connCount` field incremented without explicit lock in `HandleConnection`, decremented under `RLock` in `cleanupStaleConnections` — race |
| CS-06 | `catalog-api/internal/services/cache_service.go:127-152` | Medium | Race window between `wg.Add(1)` and `Close()` checking shutdown channel — `closeMu` protects the check but `wg.Add()` can still race |
| CS-07 | `catalog-api/services/log_management_service.go:584-617` | Medium | `channel <- entry` send can block if receiver is slow — no timeout, no done-channel check — goroutine leak risk |
| CS-08 | `catalog-api/handlers/websocket_handler.go:132-152` | Medium | Ticker goroutine relies on tests calling `handler.Stop()` — not enforced |
| CS-09 | `catalog-api/handlers/media_entity_handler.go:1099-1106` | Low | Raw `h.db.ExecContext()` calls — must verify they go through `database.DB` wrapper for dialect rewriting, otherwise breaks on PostgreSQL |
| CS-10 | `catalogizer-desktop/src/components/VLCPlayer.tsx:81` | Low | `.catch(() => {})` silently swallows watch-progress update errors |
| CS-11 | `catalogizer-desktop/src/hooks/useVLCPlayer.ts:96` | Low | `.catch(() => {})` silently swallows VLC stop errors on unmount |

**Kotlin / Android:** Zero coroutine issues found — `GlobalScope.launch` absent, `runBlocking` absent from UI, Flow patterns clean. No remediation needed.

### A.3 — Test Coverage Gaps

| Area | Current | Target | Gap |
|---|---|---|---|
| `catalog-api/challenges/` | 14 test files for 89 source files (0.16 ratio) | ≥1 test file per challenge group + per-challenge unit assertions | **MAJOR** — 75 source files lack co-located tests |
| `catalog-api/cmd/` | 0 test files | Smoke test for `main()` boot path | Small but missing |
| `catalog-api/utils/` | 0 test files | Unit coverage on every util | Small but missing |
| `catalog-web` line coverage | 70.14% | ≥95% (theoretical max 100% minus generated code) | ~25 percentage points |
| `catalog-web` branch coverage | 62.23% | ≥90% | ~28 percentage points |
| `catalog-web` function coverage | 59.42% | ≥95% | ~36 percentage points |
| Skipped tests in short mode (`-short`) | 130+ across catalog-api stress/integration/performance | Always-run subset added to CI; long-run subset documented and gated by `STRESS=1` | Skipped tests are invisible to CI |
| Submodule benchmark skips | 40+ across Auth/Cache/Concurrency/EventBus/Memory/Observability/Security/Storage/Streaming benchmarks | Always-run smoke benchmark per submodule | Benchmarks never run in default CI |
| Filesystem integration tests | FTP/NFS/SMB integration tests skipped (no local servers) | Containerized FTP/NFS/SMB stacks via `docker-compose.test.yml` profiles | Real-protocol coverage missing |

### A.4 — Database Dialect Gaps

| ID | Issue | Severity |
|---|---|---|
| DB-01 | Migrations 000002, 000003, 014, 015 lack `.sqlite.up.sql` variants — only 000001 has dual-dialect (`sqlite.up.sql`). PostgreSQL-only DDL will fail SQLite test runs | High |
| DB-02 | `media_entity_handler.go:1099-1106` raw `ExecContext` — needs verification that it goes through `database.DB` wrapper | Low |

### A.5 — Disabled Services / Features

| ID | Service | File | Status |
|---|---|---|---|
| DS-01 | `prometheus` | `deployment/docker-compose.yml:166` (profile: monitoring) | Not started by default |
| DS-02 | `grafana` | `deployment/docker-compose.yml:188` (profile: monitoring) | Not started by default |
| DS-03 | `backup` | `deployment/docker-compose.yml:207` (profile: backup) | Not started by default |
| DS-04 | Test FTP/NFS/SMB stacks | (do not exist) | Need creation |

### A.6 — Documentation Gaps

| ID | Surface | Issue | Severity |
|---|---|---|---|
| DOC-01 | `catalog-api/AGENTS.md` | Missing — every other submodule has it | Medium |
| DOC-02 | `Database/ARCHITECTURE.md` | Missing — every other Go submodule has it | Medium |
| DOC-03 | `docs/guides/USER_MANUAL.md` | Header reads "v2.2.0" — current is v2.3.0 | Medium |
| DOC-04 | `docs/guides/PERFORMANCE_TUNING.md` | "Applies to: v2.2.0+" | Low |
| DOC-05 | `docs/deployment/KUBERNETES_DEPLOYMENT.md` | "Applies to: v2.2.0+" | Low |
| DOC-06 | `docs/api/API_DOCUMENTATION.md` | Response examples show `"version": "2.1.0"` | Medium |
| DOC-07 | `Website/changelog.md` | Latest entry: v2.2.0 (2026-04-03) — needs v2.3.0 + future versions | Medium |
| DOC-08 | `Website/features.md` | No v2.3 section | Low |
| DOC-09 | `Website/faq.md` | References v2.1.0 features | Low |
| DOC-10 | `Website/course.md` | Mentions v2.1.0 in course outline | Low |
| DOC-11 | `.vitepress/config.ts` | `ignoreDeadLinks: true` — masking sidebar links to non-existent `/docs/developer-guide/*` | Medium |
| DOC-12 | Diagrams | Only ASCII/markdown diagrams (`docs/diagrams/*.md`) — no `.mmd`, `.puml`, `.drawio`, `.svg` source files | Medium |
| DOC-13 | Video courses | 30 module scripts present but no per-module README, storyboard, or producer notes | Low |
| DOC-14 | API endpoints doc | OpenAPI spec exists (`openapi.yaml`, 24 paths) but no narrative `docs/api/endpoints.md` listing every `/api/v1/*` route | Low |

### A.7 — Security Scan Status

Per CLAUDE.md: `govulncheck` clean, `npm audit` clean per the last QA campaign. **Snyk and SonarQube via Compose are NOT confirmed wired** — `docker-compose.security.yml` has Snyk/Trivy/Gosec but no SonarQube container. Sonarqube scan script (`scripts/run-sonarqube-scan.sh`) exists but its execution environment must be containerized and proven not to require sudo/interactive auth.

| ID | Tool | Status | Action |
|---|---|---|---|
| SEC-01 | `govulncheck` | Clean (per memory) | Re-run baseline this campaign |
| SEC-02 | `npm audit` | Clean (per memory) | Re-run baseline this campaign |
| SEC-03 | Snyk | Not in `docker-compose.security.yml` as runnable profile | Add containerized Snyk profile + token via env var |
| SEC-04 | SonarQube | `run-sonarqube-scan.sh` exists; container wiring unverified | Add SonarQube + scanner containers to `docker-compose.security.yml`, rootless |
| SEC-05 | Semgrep | In `docker-compose.security.yml --profile semgrep-scan` | Run baseline + fix all findings |
| SEC-06 | Trivy | In `docker-compose.security.yml` | Run on built images + fix |
| SEC-07 | Gosec | In `docker-compose.security.yml` | Run + fix |

### A.8 — Performance / Observability Gaps

| ID | Gap | Impact |
|---|---|---|
| PO-01 | No HTTP middleware emitting per-route latency, error count, in-flight count to Prometheus | CC-02, CC-04, CC-06 stay stubbed |
| PO-02 | No metrics-collection daemon for system load (CPU/memory/FD) | CC-03 stays stubbed |
| PO-03 | Prometheus + Grafana behind `monitoring` profile — not started by default | DS-01, DS-02 |
| PO-04 | No always-on stress/soak run as part of nightly CI | Regressions invisible until release campaign |
| PO-05 | No baseline of lazy-loading / semaphore use across services — opportunities unmeasured | Need audit + targets |

---

## Section B — Master Phased Roadmap

12 phases. Each phase has: **Scope** (what's in), **Deliverables** (concrete artifacts), **Exit criteria** (how we prove it's done), **Dependencies** (what must be done first), **Sub-plan reference** (where the bite-sized tasks live).

### Phase 0 — Audit & Baseline (this document) ✅

**Status:** Complete (this document is the artifact).

**Deliverables:** Section A above. The remaining phases consume A.1–A.8 as their inputs.

---

### Phase 1 — Concurrency Safety & Data-Race Hardening

**Sub-plan:** `docs/plans/2026-04-11-phase-1-concurrency-hardening.md`

**Scope:** Fix every CS-* finding from A.2. Highest priority because CS-01 is a guaranteed panic and CS-03/CS-04/CS-05 are real races.

**Deliverables:**
1. `catalog-api/smb/types.go` — `StopCleanup()` rewritten with `sync.Once` and a separate `done` channel; cleanup loop uses `select` with `<-ctx.Done()`; no double-close, no read-after-close.
2. `catalog-api/middleware/request.go` + `middleware/advanced_rate_limiter.go` — both cleanup goroutines accept a stop signal; both have `Stop()` methods called from `main.go` shutdown.
3. `catalog-api/handlers/websocket_handler.go` — `connCount` migrated to `atomic.Int64`; `connCount` reads/writes audited.
4. `catalog-api/internal/services/cache_service.go` — `wg.Add` moved before goroutine launch under `closeMu`; `Close()` checks shutdown atomically.
5. `catalog-api/services/log_management_service.go` — channel send wrapped in `select` with `<-ctx.Done()` and a timeout.
6. `catalog-api/handlers/media_entity_handler.go` — raw `ExecContext` calls verified or wrapped through `database.DB`.
7. `catalogizer-desktop/src/components/VLCPlayer.tsx` + `useVLCPlayer.ts` — `.catch(() => {})` replaced with logging via the existing logger module.
8. **Regression tests** for every CS-* finding under `catalog-api/internal/tests/concurrency_regression_test.go` (Go) and `catalogizer-desktop/src/__tests__/vlc-player.test.tsx` (TS). Each test reproduces the original race / panic / leak before the fix and passes after.
9. **Race-detector challenge** — new challenge `ch_concurrency_smoke` registered in `RegisterConcurrencyHardeningChallenges()` that runs `go test -race ./...` on a representative subset and asserts zero races.

**Exit criteria:**
- `GOMAXPROCS=3 go test -race ./... -p 2 -parallel 2` is green for catalog-api and every Go submodule (zero data races).
- All 11 CS-* findings have a test that fails on the pre-fix code and passes on the post-fix code.
- `main.go` shutdown sequence stops every long-lived goroutine before `srv.Shutdown(ctx)`.
- New regression tests added to `banks/fixes-validation.yaml`.
- Phase 1 commit pushed to all 6 remotes.

**Dependencies:** Phase 0.

---

### Phase 2 — Stub Implementation Completion

**Sub-plan:** `docs/plans/2026-04-11-phase-2-stub-completion.md` (to be generated when Phase 1 completes)

**Scope:** Fix every CC-* finding from A.1.

**Deliverables:**
1. **`ReportingService` (CC-01..06)** — implement real growth-rate calculation by querying historical aggregates from a new `reporting_snapshots` table; implement `calculatePerformanceMetrics`/`calculateSystemLoad`/`calculateErrorRates`/`calculateResponseTimes`/`vulnerability_count` by reading from the Prometheus metrics registry exposed at `/metrics` (Phase 6 will populate it). For now, ReportingService reads whatever metrics already exist plus database-derived values. No method may return empty/zero — if data is unavailable, return `error`.
2. **`PlaybackPositionService.SyncAcrossDevices()` (CC-07)** — real implementation: read latest playback position per `(user_id, media_id)`, broadcast via the existing WebSocket event bus to all connected clients of the same user, persist sync timestamp to a new `playback_sync_log` table.
3. **`CacheService.Warmup()` (CC-08)** — read configured warmup keys from `config.json`, fetch them via the underlying cache backend, log throughput.
4. **`MediaRecognitionService` cover-art / metadata (CC-09)** — wire the existing `internal/media/providers/tmdb.go` and `internal/media/providers/musicbrainz.go` clients into the 5 placeholder methods. Each method returns a real result OR a typed error (`ErrProviderUnavailable`); never empty.
5. **`internal/services/smb.go` `Connect()` (CC-10)** — replace stub with a real SMB session handshake using the existing `internal/smb` package; return connection on success, error on failure.
6. **Challenges CC-11..13** — replace `"passes as stub"` returns with proper assertions; if the challenge truly cannot run in the current env, mark it `Skipped` with structured reason, never `Passed`.
7. **Schema migrations** — add `reporting_snapshots` and `playback_sync_log` migrations (with both PostgreSQL and SQLite variants, addressing DB-01 partially).
8. **Unit tests** for every newly-implemented method, table-driven, edge cases included. Tests use `database.WrapDB()` for in-memory SQLite.
9. **Documentation update** for every behavior change in the relevant `docs/architecture/*.md`.

**Exit criteria:**
- Zero stub methods remain. Every method either does real work or returns a typed error.
- `go test ./services/... ./internal/services/...` 100% pass.
- New methods covered ≥95%.
- ReportingService returns data for every field requested by the existing reports (verified by a new integration test).

**Dependencies:** Phase 1 complete.

---

### Phase 3 — Database Dialect & Migration Completeness

**Sub-plan:** `docs/plans/2026-04-11-phase-3-database-dialect.md`

**Scope:** Address DB-01, DB-02. Make every migration dual-dialect, audit every raw SQL access path.

**Deliverables:**
1. `database/migrations/000002_conversion_jobs.sqlite.up.sql` (+ down)
2. `database/migrations/000003_add_user_tables.sqlite.up.sql` (+ down)
3. `database/migrations/014_create_subtitle_tables.sqlite.up.sql` (+ down)
4. `database/migrations/015_fix_subtitle_foreign_keys.sqlite.up.sql` (+ down)
5. Migrations from Phase 2 (`reporting_snapshots`, `playback_sync_log`) — both dialects shipped together.
6. **Audit script** `scripts/audit-raw-sql.sh` (no sudo) — greps every Go file for `db.Exec(`, `db.Query(`, `db.QueryRow(`, `*sql.DB.Exec`, `ExecContext`, `QueryContext` and reports any not going through `database.DB` wrapper.
7. **Wrapper migration** for every offending file found by step 6.
8. **Migration parity test** — new test `database/migrations_parity_test.go` that asserts every migration has both PostgreSQL and SQLite variants (or is explicitly marked PostgreSQL-only with a documented reason).
9. **Cross-dialect integration test** — one-shot test that boots an in-memory SQLite, runs all migrations, then boots a PostgreSQL container, runs all migrations, and asserts the resulting `information_schema` is functionally equivalent (same tables, same columns, same indexes).

**Exit criteria:**
- Every migration has both dialects.
- `scripts/audit-raw-sql.sh` reports zero offenders.
- Migration parity test green on both dialects.
- New tests added to `banks/fixes-validation.yaml`.

**Dependencies:** Phase 2 (because Phase 2 adds new migrations).

---

### Phase 4 — Test Coverage Maximization

**Sub-plan:** `docs/plans/2026-04-11-phase-4-test-coverage.md`

**Scope:** Push every component to its theoretical maximum coverage. Address every gap in A.3.

**Deliverables:**
1. **`catalog-api/challenges/` unit tests** — for every challenge file (89 source files), write a co-located `*_test.go` that:
   - Constructs the challenge in isolation,
   - Mocks its dependencies (HTTP clients, database) via the existing test helpers,
   - Asserts on success and failure paths,
   - Asserts on the `Result` schema.
2. **`catalog-api/cmd/` smoke test** — boot main with a temp config, assert no panic, assert HTTP server binds, assert clean shutdown.
3. **`catalog-api/utils/` unit tests** — one test per util function, edge cases.
4. **`catalog-web` coverage push** — for every untested or under-tested file under `src/`, add Vitest tests until line coverage ≥95%, branch ≥90%, function ≥95%. Use existing patterns from `__tests__/` siblings.
5. **Submodule benchmark always-run subset** — for each `tests/benchmark/*.go` that uses `b.Skip("...short mode...")`, add a smoke variant that runs in `-short` mode with reduced N.
6. **Filesystem integration tests** — add `docker-compose.test.yml` profiles for `vsftpd`, `nfs-ganesha`, `samba`. Replace `t.Skip("Skipping integration test")` with `if !envServerRunning() { t.Skip(...) }` BUT add a CI job that brings up the containers and runs the integration tests.
7. **Coverage gate challenges** — extend `RegisterCoverageGateChallenges()` so every package has a per-package coverage assertion (`≥95%` line).
8. **Coverage report aggregator** — `scripts/coverage-aggregate.sh` (no sudo) that reads all coverage outputs, produces `docs/reports/coverage-latest.md` with per-package + total numbers, and fails if any package is below the per-package gate.

**Exit criteria:**
- catalog-api line coverage ≥95% per package.
- catalog-web line coverage ≥95%, branch ≥90%, function ≥95%.
- Zero `t.Skip` without a documented reason and a corresponding "always-run" sibling test.
- All coverage challenges green.

**Dependencies:** Phase 3 (because new code from Phases 1–3 needs to be in coverage scope).

---

### Phase 5 — Stress, Soak & Integration Test Suite

**Sub-plan:** `docs/plans/2026-04-11-phase-5-stress-integration.md`

**Scope:** Build the always-on validation suite that proves the system is "responsive like the flash and not possible to overload or break".

**Deliverables:**
1. **k6 test expansion** — `tests/k6/`:
   - `load_test.js` — already exists, verify p95 < 500 ms at 50 VU.
   - `stress_test.js` — already exists, find break point.
   - `soak_test.js` — already exists, 30 min memory-leak detection.
   - **NEW**: `spike_test.js` — instant 0→500 VU, verify recovery.
   - **NEW**: `breakpoint_test.js` — find max RPS with constant ramp.
   - **NEW**: `endurance_test.js` — 4-hour run, hourly assertions.
   - **NEW**: `concurrent_writers_test.js` — N writers vs N readers on the same media items.
2. **Integration test stack** — `docker-compose.test-full.yml` boots: catalog-api + catalog-web + Postgres + Redis + Prometheus + vsftpd + nfs-ganesha + samba + Playwright + k6. Single-command boot.
3. **Always-on stress challenges** — register `ch_stress_*` challenges that run k6 scripts via the existing `userflow-runner` and assert thresholds.
4. **Race detector CI job** — `scripts/run-race-detector.sh` runs `GOMAXPROCS=3 go test -race -p 2 -parallel 2 ./...` on every Go submodule + catalog-api, fails on any race. Wired into release-build.
5. **HelixQA stress bank** — new `banks/full-qa-stress.yaml` with rapid-action sequences, network interruption mid-stream, login storms, etc.

**Exit criteria:**
- Every k6 script meets its threshold on the test stack.
- Race detector green across the entire codebase.
- New integration tests added to release-build pipeline.
- HelixQA stress bank passes 100% on at least one Android TV device + web + desktop.

**Dependencies:** Phase 4.

---

### Phase 6 — Performance Monitoring, Metrics & Optimization

**Sub-plan:** `docs/plans/2026-04-11-phase-6-perf-monitoring.md`

**Scope:** Address PO-01..05. Build the metrics infrastructure that ReportingService depends on (closes the loop on Phase 2's CC-02..06). Audit and improve lazy loading + semaphore use.

**Deliverables:**
1. **HTTP metrics middleware** — `internal/middleware/metrics.go` registering Prometheus counters/histograms per route: `http_requests_total`, `http_request_duration_seconds`, `http_in_flight`, `http_errors_total`. Wired in `main.go` BEFORE all routes.
2. **System metrics collector** — `internal/services/system_metrics.go` that periodically samples `runtime.MemStats`, goroutine count, GC pause, file descriptors, and exposes via Prometheus.
3. **Prometheus + Grafana on by default** — remove `profiles: monitoring` from `deployment/docker-compose.yml` so they're started by default. Add resource limits per CLAUDE.md.
4. **Grafana dashboards** — provision JSON dashboards under `monitoring/grafana/dashboards/` for: API latency (p50/p95/p99), error rate by route, cache hit rate, goroutine count, DB pool stats, WebSocket connection count.
5. **ReportingService re-test** — after metrics are populated, re-run Phase 2's ReportingService tests against a real metrics endpoint and assert non-empty data.
6. **Lazy loading audit** — `scripts/audit-lazy-init.sh` walks every `New*` constructor in catalog-api + Go submodules and reports services that:
   - Acquire DB connections at construction (eager) → candidates for lazy
   - Pre-populate caches at boot → candidates for lazy
   - Spawn goroutines at construction → already lazy via `LazyServiceRegistry` but verify
7. **Semaphore audit** — `scripts/audit-semaphores.sh` finds every parallel-execution site and verifies it uses `internal/concurrency/semaphore.go` for bounded concurrency. Add semaphores to any unbounded `for ... { go func() {...}() }` patterns.
8. **Non-blocking audit** — find every blocking call on the request hot path (sync DB query that could be cached, sync HTTP call that could be batched) and add async/cache layers where safe.
9. **Performance regression challenges** — `ch_perf_baseline_*` that runs k6 against the test stack and asserts p95 < threshold. Failure halts the build.
10. **Optimization changelog** — `docs/reports/optimizations-2026-04-11.md` listing every optimization applied with before/after numbers.

**Exit criteria:**
- Prometheus + Grafana running by default.
- All ReportingService stubs from Phase 2 returning real numbers in production.
- Lazy/semaphore audits show zero unbounded patterns.
- Perf baseline challenges green; before/after optimizations recorded.

**Dependencies:** Phase 5.

---

### Phase 7 — Security Scanning & Resolution

**Sub-plan:** `docs/plans/2026-04-11-phase-7-security.md`

**Scope:** Address SEC-01..07. Wire Snyk + SonarQube via Compose, run all scans, fix every finding.

**Deliverables:**
1. **Snyk Compose profile** — extend `docker-compose.security.yml` with a `snyk` service (image: `snyk/snyk:linux`) under `--profile snyk-scan`. Token via `SNYK_TOKEN` env var (read from local `.env`, never committed). Scans Go + npm + container images. Output to `docs/reports/security/snyk-<date>.json`.
2. **SonarQube Compose profile** — extend `docker-compose.security.yml` with `sonarqube` (server) + `sonar-scanner` (scanner) services under `--profile sonarqube-scan`. Volume-mounted PostgreSQL data dir. Rootless. Output to `docs/reports/security/sonarqube-<date>.html`.
3. **Run all scans** — `./scripts/security-scan-all.sh` runs `govulncheck`, `npm audit`, `snyk test`, `semgrep scan`, `trivy image`, `gosec ./...`, `sonar-scanner`. Aggregates into `docs/reports/security/CONSOLIDATED-<date>.md`.
4. **Resolution loop** — for every High/Critical finding, write a fix + a regression test. Track in `docs/reports/security/findings-tracker.md`.
5. **Security challenges** — `RegisterSecurityScanChallenges()` runs the consolidated scan and asserts zero High/Critical.
6. **Pre-commit hook** — extend `.pre-commit-config.yaml` to run `govulncheck` and `gosec` on changed files (fast subset).

**Exit criteria:**
- All 7 scanners run via Compose with zero sudo / interactive auth.
- Zero High or Critical findings in any scanner.
- Pre-commit hook blocks new High/Critical findings.
- Security challenges green.

**Dependencies:** Phase 6 (so optimized code is what gets scanned).

---

### Phase 8 — Documentation Sync & Extension

**Sub-plan:** `docs/plans/2026-04-11-phase-8-documentation.md`

**Scope:** Address every DOC-* finding from A.6 plus extend everything documented to date.

**Deliverables:**
1. **Missing files** — `catalog-api/AGENTS.md`, `Database/ARCHITECTURE.md` (DOC-01, DOC-02).
2. **Version bumps** — every doc mentioning v2.x older than current updated to v2.3.0+ (DOC-03..06).
3. **API_DOCUMENTATION.md** rebuild — auto-generate from `openapi.yaml` via `redocly` or `widdershins`, plus a hand-written narrative section per route group. Examples reference current version.
4. **Endpoints reference** — `docs/api/endpoints.md` listing every `/api/v1/*` route extracted from `main.go` (DOC-14).
5. **Graphical diagrams** — `docs/diagrams/sources/`:
   - `architecture.mmd` (Mermaid C4 component diagram)
   - `er-model.dbml` (DBML for the entity model)
   - `auth-flow.mmd` (sequence diagram)
   - `media-aggregation.mmd` (pipeline diagram)
   - `helixqa-pipeline.mmd` (Learn → Plan → Execute → Curiosity → Analyze)
   - `tv-channels-flow.mmd`
   - `database-dialect-rewriting.mmd`
   - Render to SVG via `mermaid-cli` in a container.
6. **SQL schema doc rebuild** — `docs/sql/SCHEMA.md` auto-generated from migration files, includes ER diagram + per-table column descriptions + dual-dialect notes.
7. **Per-submodule deep-dives** — for each Go submodule, extend its existing `ARCHITECTURE.md` with: data flow diagram, public API table, internal package map, common pitfalls.
8. **Concurrency cookbook** — `docs/architecture/CONCURRENCY_PATTERNS.md` documenting `sync.Once`, semaphore, lazy registry, atomic counters, channel-based shutdown — referenced by Phase 1's regression tests.
9. **Operations runbooks** — `docs/operations/RUNBOOK_*.md` for: starting full stack, restoring backup, rotating JWT secret, rotating API keys, applying a hotfix, recovering from corrupted DB.
10. **Documentation challenges** — extend `RegisterDocumentationChallenges()` to assert: every Go submodule has README+CLAUDE.md+AGENTS.md+ARCHITECTURE.md, every public Go package has a `doc.go`, every API endpoint has OpenAPI coverage, every diagram source renders.

**Exit criteria:**
- Zero stale version mentions.
- All diagram source files render to SVG and are committed alongside.
- Documentation challenges green.
- `Database/ARCHITECTURE.md` and `catalog-api/AGENTS.md` exist.

**Dependencies:** Phase 7 (because new Phase 7 workflows need to be documented).

---

### Phase 9 — Video Course Extension & Update

**Sub-plan:** `docs/plans/2026-04-11-phase-9-video-courses.md`

**Scope:** Extend the existing 30 video-course modules to cover Phases 1–8 work; add per-module README, storyboard, producer notes; add new modules.

**Deliverables:**
1. **Per-module README** — `docs/video-course/MODULE_N/README.md` for each existing module covering: learning objectives, prerequisites, runtime, assets, exercise spec.
2. **Storyboards** — `docs/video-course/MODULE_N/STORYBOARD.md` shot-by-shot.
3. **Updated existing modules** — modules touching: backend (1, 2), concurrency (15), security scanning (16, 26), test coverage (27), perf monitoring (28), HelixQA (22), module architecture (29) — refresh scripts to reflect Phase 1–8 changes.
4. **New modules:**
   - MODULE31: "Database Dialect Rewriting & Migration Parity" (Phase 3)
   - MODULE32: "Concurrency Hardening: From Race to Atomic" (Phase 1)
   - MODULE33: "Stress, Soak & Spike Testing with k6" (Phase 5)
   - MODULE34: "Wiring SonarQube + Snyk via Compose" (Phase 7)
   - MODULE35: "Lazy Loading & Semaphore Patterns" (Phase 6)
   - MODULE36: "Building Universal Test Infrastructure with HelixQA" (HelixQA mandate)
5. **Course outline** — `docs/video-course/COURSE_OUTLINE.md` updated to include modules 31–36 with prerequisite graph.
6. **Assessment update** — `docs/courses/ASSESSMENT.md` adds questions for new modules.
7. **Course challenge** — extend `RegisterDocumentationChallenges()` to assert every module has script + README + storyboard.

**Exit criteria:**
- Every module has script + README + storyboard.
- 6 new modules complete.
- Course outline reflects current state.
- Course challenges green.

**Dependencies:** Phase 8 (course content references docs that must exist).

---

### Phase 10 — Website Refresh

**Sub-plan:** `docs/plans/2026-04-11-phase-10-website.md`

**Scope:** Address DOC-07..11. Bring `Website/` to v2.3.0+ and extend.

**Deliverables:**
1. `Website/changelog.md` — add v2.3.0 entry + every Phase 1–9 milestone.
2. `Website/features.md` — add v2.3 section.
3. `Website/faq.md` — refresh version refs.
4. `Website/course.md` — update outline to reflect modules 31–36.
5. `Website/.vitepress/config.ts` — fix every broken sidebar link, remove `ignoreDeadLinks: true`, fail build on dead links.
6. **New website pages:**
   - `Website/architecture.md` — embeds Phase 8 diagrams
   - `Website/security.md` — references Phase 7 reports
   - `Website/performance.md` — references Phase 6 dashboards
   - `Website/contributing.md` — links to GitSpec constitution + CLAUDE.md + AGENTS.md
7. **Build verification** — `npm run build` in `Website/` passes with `ignoreDeadLinks: false`.
8. **Website challenge** — `ch_website_links` runs `vitepress build` and asserts zero dead links.

**Exit criteria:**
- Website builds clean with `ignoreDeadLinks: false`.
- All version refs current.
- New pages live.
- Website challenge green.

**Dependencies:** Phase 9 (course updates referenced from website).

---

### Phase 11 — HelixQA Bank Expansion & Full Validation Run

**Sub-plan:** `docs/plans/2026-04-11-phase-11-helixqa-validation.md`

**Scope:** Run the complete HelixQA campaign on the completed system. Expand banks to cover every new feature from Phases 1–10.

**Deliverables:**
1. **Bank expansion:**
   - `banks/full-qa-api.yaml` — add cases for new ReportingService endpoints, SyncAcrossDevices, system metrics endpoint.
   - `banks/full-qa-web.yaml` — add cases for new website pages, Grafana embedding, performance dashboards.
   - `banks/full-qa-androidtv.yaml` + `full-qa-android.yaml` — add cases exercising new sync features.
   - `banks/full-qa-stress.yaml` — already added in Phase 5.
   - `banks/fixes-validation.yaml` — must contain a regression case for every CC-* and CS-* finding.
2. **Convert all YAML banks to JSON** (per HelixQA mandate).
3. **Full QA campaign** via `./scripts/helixqa-orchestrator.sh` against all 4 platforms (web, desktop, Android TV, Android phone).
4. **Real-time log monitoring** active for every session.
5. **Video recording** for every device session (mandatory per CLAUDE.md).
6. **Iterative test-fix-rebuild loop** until all banks pass.
7. **Final QA report** at `docs/reports/qa-sessions/qa-session-2026-04-11/FINAL-REPORT.md`.

**Exit criteria:**
- Every bank passes 100% on its target platforms.
- Zero outstanding tickets in `docs/reports/qa-sessions/qa-session-2026-04-11/tickets/`.
- All Phase 1–10 deliverables observably present in the QA evidence.

**Dependencies:** Phase 10.

---

### Phase 12 — Final Verification & Sign-Off

**Sub-plan:** `docs/plans/2026-04-11-phase-12-final-verification.md`

**Scope:** Prove all 11 phases are done. No new development, only verification + commit + push.

**Deliverables:**
1. **Full release build** via `./scripts/release-build.sh --container --force` (no `--skip-tests`).
2. **All-tests run** via `./scripts/run-all-tests.sh` — every Go submodule, catalog-api, catalog-web, all TS submodules, both Android apps, all Tauri apps. Zero failures, zero warnings.
3. **All-challenges run** via the catalog-api `RunAll` endpoint. Assert every challenge in every group passes.
4. **All-security run** via Phase 7's `security-scan-all.sh`. Zero High/Critical.
5. **Coverage final report** committed to `docs/reports/coverage-final-2026-04-11.md`.
6. **Master completion report** at `docs/reports/MASTER-COMPLETION-2026-04-11.md` summarizing every phase's deliverables, exit criteria status, before/after metrics, and links to evidence.
7. **Version bump** — `versions.json` to v2.4.0 build N.
8. **Final commit** + push to all 6 remotes.
9. **Memory update** — append session entry to `MEMORY.md` index.

**Exit criteria:**
- Release build green.
- All tests green.
- All challenges green.
- All scanners green.
- Master completion report committed.
- Pushed to all remotes.

**Dependencies:** Phase 11.

---

## Section C — Cross-Phase Conventions

### C.1 — Test Discipline (TDD per phase)

Every phase follows the same loop per task:

1. Write the failing test that reproduces the bug / proves the missing feature.
2. Run it. Confirm it fails for the expected reason.
3. Implement the minimal fix.
4. Run it. Confirm it passes.
5. Run the full package test (`go test ./pkg/...`) to verify no regressions.
6. Commit with conventional commit message.

### C.2 — Commit Conventions

```
<type>(<scope>): <subject>

<body>

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
```

Types: `fix`, `feat`, `test`, `docs`, `refactor`, `perf`, `chore`, `security`.

Each phase produces ≥1 commit per logical task — frequent commits per CLAUDE.md.

### C.3 — Resource Discipline (per CLAUDE.md)

- All Go test runs: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- All container runs: `--cpus=N --memory=Ng` per CLAUDE.md budget
- Total budget: 4 CPUs, 8 GB RAM
- Monitor: `podman stats --no-stream` between phases

### C.4 — No-Sudo Discipline

Every script created must:
- Run as the local user
- Use rootless Podman
- Read tokens from `.env` (gitignored), never prompt
- Have a `# REQUIRES: no sudo` comment in the header

### C.5 — Universal Solution Discipline

No phase modifies an app under test to make tests work. Test infrastructure improvements go in test infra (HelixQA, k6 scripts, test stacks), never in production code.

### C.6 — Per-Phase Push

Every phase ends with a push to all 6 remotes via:

```bash
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

### C.7 — Phase Sub-Plan Generation

Each phase's bite-sized sub-plan (`docs/plans/2026-04-11-phase-N-*.md`) is generated **after** its predecessor completes and before its tasks begin. This avoids planning against stale state — Phase 1's fixes will change the file:line numbers Phase 2 references.

The sub-plan generator: dispatch a `superpowers:writing-plans` invocation pointed at this master roadmap's relevant phase section + the latest audit findings for that phase's scope.

---

## Section D — Risk Register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Phase 1 race fixes break existing functionality | Medium | High | Full test run + race detector after every commit; per-fix regression test |
| Phase 2 stub completions require new schema migrations that conflict with existing data | Medium | Medium | All migrations dual-dialect from the start; tested on copy of production data |
| Phase 6 metrics middleware adds latency | Low | Medium | Histogram observation overhead measured; <0.5% target |
| Phase 7 SonarQube container resource use exceeds budget | High | Low | SonarQube run isolated from other services; explicit `--cpus=2 --memory=4g` |
| HelixQA campaign in Phase 11 finds new Phase 1–10 regressions | High | Medium | Iterative test-fix-rebuild loop is built into Phase 11 by design |
| Documentation drift between commits | High | Medium | Documentation challenges enforce per-commit; pre-commit hook runs them |
| Scope creep into "while we're here" cleanups | High | Medium | Each phase has fixed scope; out-of-scope cleanups go into a backlog file, not the current phase |

---

## Section E — How to Execute This Plan

This is a master roadmap. **Do not execute it directly.** Each phase has its own bite-sized sub-plan with the exact files, code, commands, and TDD steps needed to do the work. Execution flow:

1. Read this document end-to-end.
2. Open `docs/plans/2026-04-11-phase-1-concurrency-hardening.md`.
3. Execute Phase 1 via `superpowers:executing-plans` (inline) or `superpowers:subagent-driven-development` (parallel review).
4. When Phase 1 exit criteria are met, dispatch a fresh `writing-plans` invocation to generate `docs/plans/2026-04-11-phase-2-stub-completion.md`.
5. Execute Phase 2.
6. Repeat for Phases 3–12.

Phases 2–12 sub-plans are intentionally not generated yet because each one needs to plan against the post-previous-phase state of the code, not the current state.

**Total estimated effort:** 6–10 weeks of focused work for one engineer. 2–4 weeks with parallel subagent dispatch.

---

## Section F — Acceptance Criteria for the Whole Initiative

When this plan is complete, the following statements must all be true:

1. ✅ Zero stub/no-op/placeholder methods in production code (every CC-* fixed).
2. ✅ Zero data races detected by `go test -race ./...` across the monorepo (every CS-* fixed).
3. ✅ Every package has ≥95% line coverage; catalog-web ≥95% line / ≥90% branch.
4. ✅ Every migration has both PostgreSQL and SQLite variants; cross-dialect parity test green.
5. ✅ Every k6 stress, soak, spike, and breakpoint test meets its threshold.
6. ✅ Prometheus + Grafana running by default; ReportingService returns real numbers from real metrics.
7. ✅ All 7 security scanners (govulncheck, npm audit, snyk, semgrep, trivy, gosec, sonarqube) report zero High/Critical.
8. ✅ Every doc surface (catalog-api+submodule docs, manuals, video courses, website, diagrams, SQL) reflects v2.4.0 (post-completion bump).
9. ✅ Website builds with `ignoreDeadLinks: false`.
10. ✅ HelixQA full-qa-* + fixes-validation banks pass 100% on web, desktop, Android, Android TV.
11. ✅ All 6 git remotes are in sync at the post-Phase-12 commit.
12. ✅ The next engineer reading CLAUDE.md and this roadmap can understand the system and contribute on day one.

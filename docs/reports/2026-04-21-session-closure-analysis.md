# 2026-04-21 — Session Closure & Open-Items Analysis

**Scope:** Article VII Master Cycle (2026-04-20 T22:05) plus follow-up
Q-cycle, R-cycle, and S-cycle (2026-04-21). This document summarises
what landed, what's deferred with pointers to specific tickets, and
what's blocked on external inputs — so a future session can pick up
any item without re-deriving context.

**Authoritative session archive:** `docs/reports/qa-sessions/2026-04-20-T22-05/`
(FINAL-REPORT.md, tickets/, analysis/parse-runall-log.py, logs/).

---

## 1. Fixes landed this session

All tests green (catalog-api 45/45 packages; HelixQA e2e + vision
packages; catalog-web 131 files / 2318 tests). All commits pushed to
every configured upstream.

| ID | Area | Commit(s) | Summary |
|---|---|---|---|
| `FIX-QA-2026-04-20-001` | HelixQA / `tests/e2e/pipeline_test.go` | HelixQA `dfa8562` | TestFullPipeline false-positive: featureless gradient fed to contour detector + `assert.NotEmpty` (non-fatal) paired with unconditional success log. Routed feature-rich `createTestImageWithText` fixture + converted to `require.NotEmpty`. |
| `FIX-QA-2026-04-20-002` | HelixQA / same file | HelixQA `55037b6` | Swept 6 more `assert.* + success-log` sites (`TestPerformance`, `TestDistributedState`, `TestWebRTCSignaling`, `TestHostDiscovery`, `TestVisionOCRIntegration`, `TestGStreamerPipeline`, `TestConcurrentProcessing`) to `require.*`. Raised `TestPerformance` latency budget 100 → 500 ms to tolerate full-suite CPU contention (isolated run 1-65 µs/frame; full-suite 150-250 ms/frame). Guarded latent FPS div-by-zero. |
| `FIX-QA-2026-04-21-001` | catalog-api / `handlers/media_handler.go` + `database/` | main `e8231aab` | `PUT /api/v1/media/:id/favorite` returned 500 on every call — handler wrote to `media_items.is_favorite` + `media_items.updated_at`, neither column existed in production schema (only in the parallel `internal/tests/test_helper.go` schema). Added **migration v18** (`add_media_items_favorite_column`) with SQLite `PRAGMA table_info` probe + `updated_at` backfill from `last_updated`. Added `TestUpdateFavoriteStatus_HappyPath` and `TestUpdateFavoriteStatus_MediaNotFound` — before this, every favorite test was a 400-path branch, so the real UPDATE SQL was never exercised. |
| `FIX-QA-2026-04-21-002` | catalog-api / `main.go` | main `e8231aab` | `/api/v1/health` returned 404 — canonical path was `/health` only. Bank + external monitors probed both. Added shared-handler closure mirroring the route under `api` group root (kept unauthenticated). |
| `FIX-QA-2026-04-21-003` | catalog-api / `main.go` | main `2b851e24` | `/api/v1/admin/{config,errors,health,logs}` each 404×4. Added 4 alias GETs delegating to the canonical `ConfigurationHandler.GetConfiguration`, `ErrorReportingHandler.ListErrorReports` / `GetSystemHealth`, `LogManagementHandler.ListLogCollections`. Hoisted `wrap := root_handlers.WrapHTTPHandler` above the admin group. Smoke-tested live: all 4 return 200. |
| `FIX-QA-2026-04-21-004` | catalog-api / `handlers/challenge.go` | main `2d026db0` | Partial mitigation of DEFER-001: `GET /api/v1/challenges/results` hung past the 60s `RequestTimeout` after a 508-challenge RunAll because the full results slice is tens of MB. GetResults now takes `?limit=N` (default 100, last-N semantic; 0 = unlimited) and reports `total_count`. New `TestChallengeHandler_GetResults_LimitTruncates` seeds 250 mock results and verifies the three modes. |
| Schema-drift sweep (Q2) | catalog-api / `database/migrations_v18_media_favorite.go` | main `2b851e24` | Audit of production SQL across handlers/services/repository found one more drift: `playlist_service.go:695` `SUM(mi.duration)` references `media_items.duration` (test-only column). Extended migration v18 atomically. |
| TestE2E_Quick* normalization | HelixQA / same file | HelixQA `c62fc90` | Three `TestE2E_Quick*` functions used `assert.*` for their final checks with no success log (anti-pattern didn't strictly apply, but was inconsistent with the rest of the suite after FIX-QA-2026-04-20-001/002). All converted to `require.*`; unused `testify/assert` import removed. |
| End-to-end smoke verification (S-cycle) | catalog-api | main `1b79fbf4` | Booted the rebuilt binary on port 8083; every landed fix verified live — `/health`, `/api/v1/health`, all 7 `/admin/*` paths, `/media/1/favorite` happy path (200, UPDATE hit the migration-v18 columns), `/challenges/results` 4-key shape. `go vet ./...` zero warnings. |

**Commit map (remotes pushed):**

- HelixQA: `dfa8562` → `55037b6` → `c62fc90` → `763f7dc` (4 upstreams each — GitHub×2, GitLab×2).
- Main repo: `d6832993` → `3df7d9dc` → `e8231aab` → `2b851e24` → `2d026db0` → `1b79fbf4` (6 upstreams each — GitHub×2, GitLab×2, GitFlic, GitVerse:2222).

---

## 2. Deferred with tickets (workable in a focused cycle, not this one)

### `DEFER-QA-2026-04-21-001` — full ctx-threading through Challenges runner

- **File:** `docs/reports/qa-sessions/2026-04-20-T22-05/tickets/DEFER-QA-2026-04-21-001-challenges-results-hang.md`.
- **Scope:** thread `ctx` end-to-end through `Challenges/pkg/runner/runner.go`
  and every challenge's `Execute()` so handlers can pass `c.Request.Context()`
  and respect client disconnect. Currently handlers pass `context.Background()`
  by design (comment in `handlers/challenge.go:100` says "RunAll is
  long-running and must not be cancelled by HTTP write timeouts").
- **Status:** R-cycle landed a partial mitigation (GetResults `?limit`
  default) that removes the observed symptom. Full refactor still
  deferred — touches ~508 challenges in the Challenges submodule.
- **Unblocks:** production would respect client disconnect on
  `/:id/run` POSTs; no more zombie challenge goroutines after curl
  timeouts. Measurement: re-run the 2026-04-20 RunAll and verify no
  post-RunAll activity after the HTTP 200 returns.

### `DEFER-QA-2026-04-21-002` — pprof-driven memory-burst review

- **File:** `docs/reports/qa-sessions/2026-04-20-T22-05/tickets/DEFER-QA-2026-04-21-002-memory-alerts.md`.
- **Scope:** 26 × "MEMORY ALERT: potential leak detected" log lines
  during the 16-minute RunAll, peak `heap_growth_ratio` 53.5× (alert
  threshold is 3×). Goroutine count stayed flat (32-40) — consistent
  with accumulating `ChallengeService.results` rather than a true leak,
  but needs a pprof profile to confirm.
- **Status:** alert is informational, no FATAL triggered. Not a
  user-visible regression. Fix decisions (raise threshold in
  `modules/registry.go:170`, cap `ChallengeService.results`, or stream
  results to disk) need pprof data.
- **Unblocks:** silence expected bursts and catch real leaks.

---

## 3. Blocked on external inputs (not workable without user action)

### Full production compose-stack boot

- **Observed:** `podman-compose -f docker-compose.dev.yml up postgres redis`
  failed with `rootlessport listen tcp 0.0.0.0:5432: bind: address already in use`
  and `:6379: bind: address already in use`. Pre-existing containers
  `helixagent-postgres` and `helixagent-redis` claim those ports.
- **Unblocks:** operator decision — either
  (a) `podman stop helixagent-postgres helixagent-redis`, or
  (b) override `POSTGRES_PORT` / `REDIS_PORT` in `.env` to different
  host ports for the Catalogizer stack.
- **Compose files themselves validate cleanly** (every
  `docker-compose*.yml` parses via `podman-compose config --quiet`).

### HelixQA autonomous device QA

- Article VII mandates autonomous QA per app/platform with video +
  screenshot review, but every HelixQA session needs a connected
  device pool via `.devconnect` + hardware.
- **Unblocks:** operator provides device list (Android TV, Android
  phone, desktop targets); run `./scripts/devconnect.sh`; then
  `./scripts/helixqa-orchestrator.sh`.

### Operator-action credentials

- `docs/OPEN_POINTS_CLOSURE.md` §1 lists 8 operator-only credentials
  (Fanart.tv, IGDB/Twitch, TMDB v3+v4, OMDB, Astica.AI, Gemini/OpenAI/
  Anthropic/Kimi, Semgrep app token, Sentry DSN). Claude refuses
  production secrets by constitution.
- **Unblocks:** operator rotates keys every 90 days, stores in vault,
  re-verifies `.gitignore` coverage of `.env` in every submodule.

### Full RunAll response-body capture

- The 2026-04-20 RunAll returned HTTP 200 at 16m0s but the response
  body was lost to a `tee | head -c` pipe buffer race
  (`analysis/post-mortem.md`). `parse-runall-log.py` reconstructs the
  matrix from the raw server log, which is acceptable for triage but
  not a substitute for the true per-challenge pass/fail data.
- **Unblocks:** next RunAll invocation uses `curl -o FILE` (direct
  file write, no pipe); snapshot `/api/v1/challenges/results` from a
  parallel thread mid-run.

---

## 4. Known anti-patterns to watch for (operator reminders)

1. **`assert.*` (non-fatal testify) immediately followed by an
   unconditional success log.** Every such site is a false-positive
   waiting to happen. Swept in HelixQA `tests/e2e/pipeline_test.go`
   this session — sweep other suites too.

2. **Test/production schema drift** via parallel CREATE TABLE
   statements in `internal/tests/test_helper.go` that no migration
   catches up with. `media_items` had ~40 test-only columns; only two
   (`is_favorite`, `updated_at`) actually broke handlers in the wild,
   plus `duration` found by audit. Long-term fix: have `test_helper`
   call `RunMigrations()` on the in-memory SQLite, not maintain a
   separate schema. Short-term: audit every handler UPDATE/INSERT.

3. **Handler SQL written against columns that the production
   migration never declared.** Every new handler with a real SQL
   mutation MUST have a happy-path test exercising the mutation
   against the migration schema (not test_helper), otherwise column
   drift slips through.

4. **Shell pipe buffer races on long-running command output.** Use
   `curl -o FILE` instead of `curl -s | tee FILE | head -c N` when
   the response body matters.

5. **Handlers that spawn work and don't check client disconnect.**
   `ChallengeHandler.RunChallenge` still uses `context.Background()`
   — a curl that times out leaves the challenge running server-side,
   which in turn saturates the concurrency limiter. See DEFER-001.

---

## 5. Quick-reference: commands that matter

```bash
# Reproduce a full Article VII cycle
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOTOOLCHAIN=local GOMAXPROCS=3 go build -o catalog-api .
GOTOOLCHAIN=local GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1 -timeout=600s

cd ../HelixQA
GOTOOLCHAIN=local GOMAXPROCS=3 go test -mod=vendor -count=1 -timeout=600s ./...

cd ../catalog-web
npm run test -- --run

# Re-create the server log parser table
python3 docs/reports/qa-sessions/2026-04-20-T22-05/analysis/parse-runall-log.py \
  docs/reports/qa-sessions/2026-04-20-T22-05/logs/catalog-api-server.log

# Run a single bug's regression tests
cd catalog-api
GOTOOLCHAIN=local go test -count=1 -run "TestUpdateFavoriteStatus" ./handlers/...
GOTOOLCHAIN=local go test -count=1 -run "TestChallengeHandler_GetResults_LimitTruncates" ./handlers/...
GOTOOLCHAIN=local go test -count=1 -run "TestMigrationSequence_AllVersionsApplied" ./database/...
```

---

**Last updated:** 2026-04-21. Update this file when a deferred ticket
closes or a blocked item is unblocked by an operator action.

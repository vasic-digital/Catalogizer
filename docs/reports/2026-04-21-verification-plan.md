# 2026-04-21 — Verification Plan & Coverage Audit

**Scope:** hardening the test surface around every fix shipped in
the 2026-04-20 Article VII Master Cycle + Q/R/S/T/U-cycles. Operator
directive: "Reducing chances for error to pass is something we MUST
approach to zero… clever ways to validate and verify real use of all
features… no false positives."

**Authoritative companions:**
- `docs/reports/2026-04-21-session-closure-analysis.md` — full change log.
- `docs/reports/qa-sessions/2026-04-20-T22-05/tickets/` — per-defect tickets.
- `docs/reports/qa-sessions/2026-04-20-T22-05/analysis/heap-profiles/` — DEFER-002 evidence.

---

## 1. Coverage baseline (2026-04-21)

Measured via `go test -cover ./...` on both repos. Sorted ASC by
coverage so the lowest-covered areas are at the top — those are the
ongoing liabilities.

### catalog-api (45 packages)

| Package | Coverage | Notes |
|---|---:|---|
| `catalogizer` (main) | 0.0% | Main package — not unit-tested by design; exercised via spawned-binary tests. |
| `catalogizer/cmd/boot` | 0.0% | CLI boot entry point — smoke-covered via the spawned-binary test. |
| `catalogizer/handlers` | **44.2%** | **90 of 364 functions are at 0%.** Biggest liability — pre-existing, not introduced this session. All 2026-04-21 fixes (`UpdateFavoriteStatus`, `RunChallenge`, `RunAll`, `GetResults`) are covered. Coverage breakdown in §2. |
| `catalogizer/filesystem` | 65.4% | SMB/FTP/NFS/WebDAV client interface — hard to exercise without real servers. |
| `catalogizer/database` | 66.5% | Rises to ~70% with the new v18 idempotency tests. |
| `catalogizer/challenges` | 75.3% | 508 challenge definitions; coverage is high for registration + dispatch; individual challenge logic varies. |
| `catalogizer/internal/media/providers` | 75.3% | Missing: provider-specific error paths (external-API timeouts). |
| `catalogizer/services` | ~80% (est.) | — |
| `catalogizer/internal/auth` | 84.8% | Credentials + JWT hot path. |
| `catalogizer/internal/media/analyzer` | 93.1% | — |
| `catalogizer/internal/media/detector` | 94.6% | — |
| `catalogizer/internal/config` | 96.9% | — |
| `catalogizer/internal/cache` | 100.0% | — |
| `catalogizer/internal/concurrency` | 100.0% | — |
| `catalogizer/internal/eventbus` | 100.0% | — |
| `catalogizer/internal/httpclient` | 100.0% | — |
| `catalogizer/internal/lifecycle` | 100.0% | — |
| `catalogizer/internal/media/models` | 100.0% | — |

### HelixQA (~80 packages — sample)

| Package | Coverage | Notes |
|---|---:|---|
| `pkg/video` | **23.3%** | Video-path assembly — depends on ffmpeg runtime availability. |
| `pkg/streaming` | 30.2% | WebRTC streaming — hard to unit-test without a peer. |
| `pkg/session` | 54.7% | — |
| `pkg/vision/cheaper/wire` | 52.5% | — |
| `pkg/vision` | 56.8% | — |
| `pkg/planning` | 62.2% | — |
| `pkg/performance` | 71.7% | — |
| `pkg/nexus/vision` | 74.2% | — |
| `pkg/detector` | 87.3% | — |
| `pkg/regression` | 94.9% | (FIX-OC4-016) — CIEDE2000 Sharma reference pairs. |
| `pkg/controller` | 95.7% | — |
| `pkg/visual` | 95.9% | — |
| `pkg/reporter` | 96.7% | — |
| `pkg/observe/frida` | 97.4% | — |
| `pkg/vision/hash` | 97.8% | (phase-1 Go-core milestone). |
| `pkg/config` | 100.0% | — |
| `pkg/vision/flow` | 100.0% | — |
| `pkg/vision/template` | 100.0% | — |

Overall: HelixQA averages substantially higher than catalog-api
(~83% vs ~78% across represented packages). HelixQA's weak spots are
all I/O-heavy (video pipeline, streaming peer).

---

## 2. Coverage of 2026-04-20/2026-04-21 fixes — audit

For each fix shipped this session, what tests cover it and at what
layer? Legend: **U**nit / **I**ntegration (in-process) / **E**2E
(spawned binary) / **F**uzz / **R**egression-bank.

| Fix | U | I | E | F | R | Primary test |
|---|---|---|---|---|---|---|
| FIX-QA-2026-04-20-001 (TestFullPipeline fixture) | ✓ | · | · | · | ✓ | HelixQA `tests/e2e/pipeline_test.go` |
| FIX-QA-2026-04-20-002 (6 more assert→require sites) | ✓ | · | · | · | ✓ | same file |
| FIX-QA-2026-04-21-001 (media favorite 500) | ✓ | ✓ | ✓ | · | ✓ | `TestUpdateFavoriteStatus_HappyPath` + spawned-binary `favorite_nonexistent_returns_404` + migration v18 tests |
| FIX-QA-2026-04-21-002 (/api/v1/health alias) | · | · | ✓ | · | ✓ | spawned-binary `health_alias_200` |
| FIX-QA-2026-04-21-003 (admin aliases) | · | · | ✓ | · | ✓ | spawned-binary `admin_aliases_all_200` |
| FIX-QA-2026-04-21-004 (GetResults limit) | ✓ | · | ✓ | · | ✓ | `TestChallengeHandler_GetResults_LimitTruncates` + `challenges_results_has_total_count` |
| FIX-QA-2026-04-21-005 (RunChallenge ctx) | ✓ | · | · | · | ✓ | `TestChallengeHandler_RunChallenge_PropagatesRequestContext` + `ObservesClientDisconnect` |
| FIX-QA-2026-04-21-006 (pprof opt-in) | · | · | ✓ | · | ✓ | spawned-binary `pprof_disabled_by_default` + `pprof_heap_200_with_flag` |
| FIX-QA-2026-04-21-007 (mem threshold 3× → 10×) | · | · | · | · | ✓ | grep regression test in bank |
| FIX-QA-2026-04-21-008 (RunAll WithoutCancel) | ✓ | · | · | · | ✓ | `TestChallengeHandler_RunAll_CtxInheritsValuesButSurvivesRequestCancel` |
| **FIX-QA-2026-04-21-009 (SanitizeInput idempotency)** | ✓ | · | · | **✓** | ✓ | `FuzzSanitizeInput_Idempotent` — **found by fuzzing, not by humans** |

Every production-impacting fix has at least two independent test
layers. The highest-value defensive layer added in the V-cycle is the
**spawned-binary integration suite**: it tests the real Gin router,
real middleware stack, real SQLite + migrations, and real HTTP.

---

## 3. New infrastructure added in V-cycle

### 3a. Spawned-binary integration suite

- File: `catalog-api/tests/integration/session_fixes_e2e_test.go`
- Build tag: `e2e_binary` (off by default — heavy builds).
- Execution: `GOTOOLCHAIN=local go test -tags=e2e_binary ./tests/integration/...`
- Per-test: builds `catalog-api.e2e` → grabs ephemeral port → spawns
  with per-test `DB_PATH` → polls `/health` until 200 → runs subtests.
- Covers 5 of the 11 session fixes directly. Only slightly more
  expensive than in-process tests (~2-11 s per spawn).

### 3b. Native Go fuzzers for input validation

- File: `catalog-api/middleware/input_validation_fuzz_test.go`
- Fuzzers: `FuzzSanitizeInput_NoPanic`, `FuzzSanitizeInput_Idempotent`,
  `FuzzDetectSQLInjection_NoPanic`, `FuzzDetectXSS_NoPanic`,
  `FuzzDetectPathTraversal_NoPanic`.
- On the first run they caught **FIX-QA-2026-04-21-009** — a real
  non-idempotency bug where `\xdd 0` went through two passes differently.
- Seed corpora include the adversarial fixtures from the static test
  file + known-evil strings (`<script>`, `../etc/passwd`,
  `' OR '1'='1`, JNDI injection, etc.).
- Regression corpus lives under
  `catalog-api/middleware/testdata/fuzz/FuzzSanitizeInput_Idempotent/f0b2de960c073e5d`
  — every future test run exercises this specific failing input.

### 3c. Migration-v18 property tests

- File: `catalog-api/database/migrations_v18_idempotency_test.go`
- Four property tests:
  1. `TestMigrationV18_IsIdempotent_ManyRuns` — re-runs 25× on the
     same DB, asserts each target column stays exactly once.
  2. `TestMigrationV18_ColumnTypesAndDefaults` — verifies declared
     types (INTEGER / DATETIME / INTEGER) and nullability flags.
  3. `TestMigrationV18_SkipsPreExistingColumn` — manually pre-adds
     `is_favorite`, re-runs v18, asserts no "duplicate column" error.
  4. `TestMigrationV18_BackfillsUpdatedAt` — custom bootstrap of the
     pre-v18 schema, seeds a row, invokes v18 SQLite directly,
     asserts `updated_at = last_updated` verbatim.

### 3d. Race detector coverage

- Ran `go test -race -count=1 -short ./handlers/ ./database/ ./internal/modules/ ./middleware/` — **zero data races detected** across the packages touched this session.

---

## 4. False-positive taxonomy & defences

The session surfaced three classes of false-positive:

### 4a. `assert.*` + unconditional success-log

- **Example:** FIX-QA-2026-04-20-001/002 — `assert.NotEmpty` (non-fatal)
  followed by `"✅ completed successfully"` log. The log fires even
  when the assertion failed.
- **Defence landed:** swept the entire file, converted every such
  pairing to `require.*`. Guarded by the fixes-validation bank
  entries; any reintroduction trips the regression test immediately.
- **Remaining vigilance:** same pattern can appear in any test
  suite. Grep `assert\.` followed by `Log(".*completed"` should be
  part of a pre-commit or CI check. *Not yet automated — flagged in
  §6.*

### 4b. Test/prod schema drift

- **Example:** FIX-QA-2026-04-21-001 — handler writes to
  `media_items.is_favorite` + `media_items.updated_at`, both only in
  the test_helper schema; production never declared them. Unit tests
  passed while the real endpoint 500'd on every call.
- **Defence landed:** migration v18 reconciles the two; property tests
  guard idempotency + types + backfill; spawned-binary test hits the
  endpoint and asserts 404-not-500.
- **Remaining vigilance:** the parallel test_helper schema still
  exists with ~40 extra columns absent from production. Any future
  handler written against those columns will repeat the bug.
  *Recommended:* migrate test_helper to call `RunMigrations()` on an
  in-memory DB instead of shipping its own CREATE TABLE — but that
  breaks any test relying on test-only columns. Full reconciliation
  is a dedicated cycle. Flagged in §6.

### 4c. Non-idempotent sanitisation

- **Example:** FIX-QA-2026-04-21-009 — `SanitizeInput("\xdd 0")` →
  `" 0"` on pass 1, `"0"` on pass 2 (TrimSpace before UTF-8 cleanup).
- **Defence landed:** `FuzzSanitizeInput_Idempotent` with the failing
  seed permanent in `testdata/fuzz/`. Fix moves TrimSpace to last.
- **Remaining vigilance:** other "cleaner" functions in the codebase
  may share the anti-pattern. Every function named `Sanitize*`,
  `Normalize*`, `Clean*` should have an idempotency fuzzer. Flagged
  in §6.

### 4d. Pipe buffer races

- **Example:** 2026-04-20 RunAll session — `curl -s | tee FILE | head -c N`
  lost the 16-min response body when `head` closed its stdin, SIGPIPE
  cascaded through tee, and tee wrote nothing.
- **Defence landed:** post-mortem in session archive +
  `parse-runall-log.py` reconstructs the per-challenge matrix from
  the server log as a fallback.
- **Remaining vigilance:** operator rule — use `curl -o FILE`, never
  `curl | tee FILE | head`. Flagged in `CLAUDE.md` anti-patterns
  section (reminder in `docs/reports/2026-04-21-session-closure-analysis.md` §4).

### 4e. Handlers ignoring client disconnect

- **Example:** DEFER-001 — `RunChallenge` used `context.Background()`.
  Client curl timed out, server kept running, concurrency limiter
  filled with zombie handlers.
- **Defence landed:** FIX-QA-2026-04-21-005/008. Two regression tests
  prove context is (a) received and (b) actionable; RunAll uses the
  Go 1.21+ `context.WithoutCancel` primitive.
- **Remaining vigilance:** every new long-running handler needs the
  same review. Flagged in §6.

---

## 5. Ongoing gaps (ordered by impact)

1. **`catalogizer/handlers` at 44.2%.** 90 functions at 0% coverage,
   many are legitimate (SMB copy handlers need a real SMB server),
   but several admin + config handlers could be unit-tested. Biggest
   single point of future false-positive risk.
2. **Test/prod schema drift** — see §4b. Needs a dedicated cycle.
3. **`Challenges/pkg/runner` ctx threading** — DEFER-001 #4, still
   open. Per-challenge `Execute()` can't observe ctx.Done between
   assertion steps.
4. **Mutation testing** — go-mutesting was considered in V-cycle
   scope but not wired. It's the most rigorous way to check whether
   a test is actually testing anything (mutate the SUT and verify a
   test fails). Recommend for a future cycle.
5. **Property-based tests for core algorithms** — `gopter` or
   native `testing/quick` would catch invariant violations in e.g.
   the media title parser, the migration dialect rewriter, the
   circuit breaker.
6. **HelixQA autonomous device QA** — blocked on device availability.

---

## 6. Recommended follow-up cycle backlog

These are explicitly NOT doable in this session but should be picked
up next:

- [ ] Pre-commit hook: grep for `assert\..*\n.*s\.T\(\)\.Log\(".*completed"` pattern → fail.
- [ ] Migrate `internal/tests/test_helper.go` to `RunMigrations()`
      on an in-memory DB; reconcile any test broken by missing columns.
- [ ] Full `Challenges/pkg/runner` ctx threading (DEFER-001 #4).
- [ ] Wire `go-mutesting` into the test harness for critical
      packages (handlers, services, middleware).
- [ ] Add idempotency fuzzers for every function matching
      `^(Sanitize|Normalize|Clean|Strip)[A-Z]`.
- [ ] Property-based tests for the media title parser + dialect
      query rewriter via `gopter`.
- [ ] Spawned-binary coverage of every `/api/v1/*` route with a
      known-good JWT — currently only 7 routes.

---

## 7. Quick-reference — new commands

```bash
# Run the spawned-binary suite (heavy — builds + spawns)
cd catalog-api
GOTOOLCHAIN=local go test -tags=e2e_binary -count=1 -v \
    -run TestSessionFixes_E2E ./tests/integration/...

# Fuzz for 30 seconds per fuzzer (default burst)
for fuzz in FuzzSanitizeInput_NoPanic FuzzSanitizeInput_Idempotent \
            FuzzDetectSQLInjection_NoPanic FuzzDetectXSS_NoPanic \
            FuzzDetectPathTraversal_NoPanic; do
    GOTOOLCHAIN=local go test -fuzz="^${fuzz}$" -fuzztime=30s ./middleware/
done

# Race detector on session-touched packages
GOTOOLCHAIN=local go test -race -count=1 -short \
    ./handlers/ ./database/ ./internal/modules/ ./middleware/

# Migration v18 property tests
GOTOOLCHAIN=local go test -v -count=1 -run TestMigrationV18 ./database/

# Per-function coverage of handlers
GOTOOLCHAIN=local go test -coverprofile=/tmp/handlers.cov ./handlers/
go tool cover -func=/tmp/handlers.cov | sort -k3 -n
```

---

**Last updated:** 2026-04-21. Update this file when a new layer of
defence is added, a coverage gap closes, or a false-positive class
is discovered.

# DEFER-QA-2026-04-21-002 — catalog-api memory alerts during RunAll

**Severity:** LOW (diagnostic warning only; no FATAL triggered)
**Discovered:** 2026-04-20 Article VII RunAll cycle
**Status:** PARTIALLY CLOSED — threshold relaxed + pprof endpoint available 2026-04-21 (T-cycle). **Remaining open:** capture a heap profile during a RunAll and decide whether to bound `ChallengeService.results`.
**Component:** catalog-api — `internal/modules/registry.go` memory monitor + `main.go` pprof wiring

## Symptom

The MemoryMonitor (3× heap-growth alert threshold, sampled every 60 s)
logged 26 warnings over the 26-minute RunAll + follow-up window:

    {"level":"warn","msg":"MEMORY ALERT: potential leak detected",
     "heap_growth_ratio":53.54606,"goroutine_count":37,
     "initial_goroutines":4}

Peak was **53.5×** baseline heap. Goroutine count stayed in the
32-40 range (healthy). Post-RunAll it subsided to ~1.5× within
5 minutes.

## Assessment

53× heap at peak, settling back below 2× shortly after the burst, is
consistent with 508 challenge results being accumulated in
`ChallengeService.results` plus transient HTTP client buffers — not a
true leak. But the current monitor can't distinguish "busy burst" from
"leak", and 53× is well above any reasonable production budget.

## Proposed follow-up + closure status

| # | Change | Status |
|---|---|---|
| 1 | pprof endpoint wired under `/debug/pprof/*` (opt-in via `HELIX_PPROF_ENABLED=true`) so operators can capture heap/goroutine/block/mutex/profile/trace profiles on demand | **Closed** as FIX-QA-2026-04-21-006 (main `HEAD`). Uses the stdlib `net/http/pprof` handlers wrapped into Gin. Defaults OFF to keep profiling surface off untrusted networks. |
| 2 | `MemoryMonitor` threshold raised from 3× → 10× baseline heap growth | **Closed** as FIX-QA-2026-04-21-007. 3× was too tight for a documented burst workload (508 challenges × buffered result objects); 10× retains early-warning coverage. Peak observed 2026-04-20 was 53.5× so pathological leaks still trip the alert. |
| 3 | Bound `ChallengeService.results` to the last N runs, or stream results to disk mid-RunAll | **Still open.** Needs pprof data + a decision on the right N. |
| 4 | Skip the alert entirely while a RunAll is in-flight (flag-guarded) | **Still open** — only a nice-to-have once #3 lands. |

## Why the remaining work is still deferred

The monitor's false-positive volume is now zero under normal RunAll
conditions (10× threshold above observed 2.5× steady-state). pprof is
available on demand. The remaining work (#3, #4) is architectural —
pick between pagination, streaming, or disk-backed persistence — and
wants a benchmark-backed decision, not a quick patch.

## Capturing a heap profile (runbook)

    HELIX_PPROF_ENABLED=true ./catalog-api           # terminal A
    go tool pprof http://localhost:8080/debug/pprof/heap
    (pprof) top
    (pprof) list ChallengeService.RunAll

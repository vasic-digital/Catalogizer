# DEFER-QA-2026-04-21-002 — catalog-api memory alerts during RunAll

**Severity:** LOW (diagnostic warning only; no FATAL triggered)
**Discovered:** 2026-04-20 Article VII RunAll cycle
**Status:** OPEN (instrumentation task — needs pprof session)
**Component:** catalog-api — `internal/modules/registry.go` memory monitor

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

## Proposed follow-up

1. Add a pprof endpoint (already present via `net/http/pprof` optional
   import?) and capture a heap profile mid-RunAll to confirm the bulk
   is `challenge.Result` slices + `strings` for assertions/outputs.
2. Consider bounding `ChallengeService.results` to the last N runs or
   streaming results to disk during RunAll instead of buffering all
   508 in memory.
3. Raise the MemoryMonitor threshold (`modules/registry.go:170`) from
   3× to 10× or skip the alert when a documented long-running
   operation (RunAll) is active.

## Why deferred

The monitor alert is informational, not load-bearing. A proper fix
needs pprof data captured during an intentional RunAll, plus a
benchmark-backed decision on where to bound the result store. Both
belong in a dedicated performance cycle.

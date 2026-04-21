# DEFER-QA-2026-04-21-001 — /api/v1/challenges/results hangs after RunAll

**Severity:** MEDIUM (post-RunAll diagnostic gap, not a user-path regression)
**Discovered:** 2026-04-20 Article VII RunAll cycle
**Status:** PARTIALLY CLOSED — handler-boundary mitigation landed 2026-04-21 as FIX-QA-2026-04-21-004 (GetResults `?limit`) + FIX-QA-2026-04-21-005 (RunChallenge ctx propagation). **Remaining open:** full ctx-threading through the Challenges submodule runner (`Challenges/pkg/runner/runner.go`) and every challenge's `Execute()`.
**Component:** catalog-api — `handlers/challenge.go` + `services/challenge_service.go` (closed); `Challenges/pkg/runner/` + per-challenge Execute (open)

## Symptom

After a 16-minute RunAll returned HTTP 200 at 22:39:36, subsequent
`GET /api/v1/challenges/results` and `GET /api/v1/challenges` calls
blocked for >120 s with 0 bytes received. Server `/health` kept
responding normally throughout.

## Analysis

Not a service mutex — `ChallengeService.GetResults()` holds only a
short `mu.RLock` for a slice copy. The hang is driven by:

1. **Global concurrency limiter saturation.** `ConcurrencyLimiter(100)`
   in `main.go:883` serialises at a 5s semaphore-acquire timeout. After
   RunAll, several single-challenge `POST /:id/run` calls I'd issued
   during the cycle were still executing server-side (handler never
   checks `c.Request.Context().Done()`) and kept semaphore slots for
   minutes.
2. **No handler-level client-disconnect check.** The challenge POST
   handlers run to completion regardless of whether the curl that
   triggered them has gone away.

## Proposed fix + closure status

| # | Change | Status |
|---|---|---|
| 1 | `GetResults` accepts `?limit=N` (default 100, last-N) + reports `total_count` | **Closed** as `FIX-QA-2026-04-21-004` (main `2d026db0`). Removes the hang symptom by bounding the JSON payload size. |
| 2 | `RunChallenge` handler passes `c.Request.Context()` instead of `context.Background()` so single-challenge runs respect client disconnect | **Closed** as `FIX-QA-2026-04-21-005`. Regression test `TestChallengeHandler_RunChallenge_PropagatesRequestContext` seeds a request with a `context.WithDeadline` and asserts the mock receives a ctx with that exact deadline. |
| 3 | `RunAll` / `RunByCategory` handlers use `context.WithoutCancel(c.Request.Context())` — keeps request-scoped values (request_id, trace, auth subject) without propagating the outer `RequestTimeout(60*time.Second)` cancel | **Closed** as FIX-QA-2026-04-21-008 (U-cycle, 2026-04-21). Regression test `TestChallengeHandler_RunAll_CtxInheritsValuesButSurvivesRequestCancel` cancels the request ctx BEFORE serving and asserts the mock's ctx (a) inherits the custom probe value, (b) is NOT Done, (c) has nil `Err()`. |
| 4 | `Challenges/pkg/runner/runner.go` threads `ctx` through `executeChallenge` to every challenge's `Execute()`; per-challenge work respects `ctx.Done()` between assertion steps | **Still open.** Submodule refactor — touches 508 challenges. |
| 5 | A test that proves the handler + service layer honour client disconnect — mock observes `ctx.Done()` when the request ctx is cancelled mid-call | **Closed** as part of U-cycle: `TestChallengeHandler_RunChallenge_ObservesClientDisconnect` blocks the mock on `<-ctx.Done()` with a 5 s safety timeout; asserts Done fires within 500 ms of the client-side cancel and `ctx.Err() == context.Canceled`. |

## Why the remaining work is still deferred

Only item #4 remains — threading `ctx` end-to-end through the
Challenges submodule runner so individual challenge steps can
co-operatively observe progress. That's a submodule refactor across
508 challenges and wants a dedicated cycle. Every handler-boundary
problem is closed; the observable symptom of DEFER-001 (RunAll
results endpoint hang + zombie single-challenge handlers) is gone.

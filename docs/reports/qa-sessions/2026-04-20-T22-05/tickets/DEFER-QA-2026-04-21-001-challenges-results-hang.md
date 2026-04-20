# DEFER-QA-2026-04-21-001 — /api/v1/challenges/results hangs after RunAll

**Severity:** MEDIUM (post-RunAll diagnostic gap, not a user-path regression)
**Discovered:** 2026-04-20 Article VII RunAll cycle
**Status:** OPEN (analysis complete; fix deferred — requires handler refactor)
**Component:** catalog-api — `handlers/challenge.go` + `services/challenge_service.go`

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

## Proposed fix

Two orthogonal changes:

1. In every challenge handler (`RunAll`, `RunChallenge`, `RunByCategory`,
   `GetResults`), pass `c.Request.Context()` instead of
   `context.Background()` so the runner can short-circuit on client
   disconnect.
2. Add a test: fire a POST `/:id/run` with a short-deadline client
   context, then immediately fire `GET /challenges` and assert the GET
   returns within 500 ms even while a long-running challenge is still
   spawning goroutines.

## Why deferred

The fix requires threading `ctx` end-to-end through
`Challenges/pkg/runner/runner.go` and every challenge's `Execute()`.
That's a meaningful refactor of a submodule and needs a dedicated
cycle to avoid destabilising the 508-challenge bank.

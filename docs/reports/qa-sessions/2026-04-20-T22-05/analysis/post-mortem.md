# Post-mortem — RunAll response capture failure

## What happened

The `POST /api/v1/challenges/run` call to the running catalog-api binary
executed successfully — the server log confirms `200 | 16m0s` at 22:39:36 for
the call that started at ~22:23:36. But the tee'd response body ended up empty
(`challenges/run-all-raw.json` — 0 bytes). A follow-up `GET /api/v1/challenges/results`
hung for 120+ s with 0 bytes received.

## Root causes

1. **Pipe buffer race.** The curl command used
   `curl -s --max-time 1800 ... | tee FILE | head -c 300`. `head -c 300` closes
   its stdin after 300 bytes, which can propagate SIGPIPE back through tee,
   and in some shells collapses the tee write to the file before stderr/exit
   propagation completes. Net result: `EXIT: 0` shown but FILE is 0 bytes.
2. **Post-RunAll lock retention.** The `/api/v1/challenges/results` GET hung,
   suggesting a global challenge lock (documented in CLAUDE.md — "RunAll is
   synchronous/blocking — no other challenge can run until it finishes") was
   still held by some post-processing step even after the 200 return.

## Mitigations for next cycle

- Use `curl -o FILE` (direct file write, no pipe) for large responses.
- Snapshot `/api/v1/challenges/results` mid-run from a parallel thread — do
  not wait until RunAll returns.
- Consider the "single-challenge run" approach: loop over each challenge ID
  and `POST /:id/run` individually so a single hung challenge doesn't block
  the whole capture.
- Lift the challenge write_timeout to cover RunAll fully (CLAUDE.md already
  documents the 900s write_timeout rule; this must be re-audited).

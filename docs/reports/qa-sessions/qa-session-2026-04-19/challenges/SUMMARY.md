# Phase-4 Challenge Runs — 2026-04-19

## Aggregate

| Bucket | Result |
|---|---|
| **Category runs (dep-clean)** | **27 / 27 ✅** |
| - module-verification | 21 / 21 ✅ |
| - environment | 2 / 2 ✅ |
| - build | 4 / 4 ✅ |
| **Individual dep-free challenges** | **10 / 11 ✅, 1 timeout** |

### Individual leaf challenges (no dependencies)

| ID | Category | Result | Latency |
|---|---|---|---|
| api-docs | documentation | ✅ | <1 s |
| browsing-api-health | e2e | ✅ | <1 s |
| config-docs | documentation | ✅ | <1 s |
| database-docs | documentation | ✅ | 155 s (slow) |
| database-connectivity | e2e | ❌ timeout | >30 s, 2× |
| first-catalog-smb-connect | e2e | ✅ | <1 s |
| grafana-dashboard-renders | observability | ✅ | <1 s |
| musicbrainz-provider-search | provider-verification | ✅ | ~1 s |
| openlibrary-provider-search | provider-verification | ✅ | <1 s |
| provider-graceful-degradation | provider-verification | ✅ | <1 s |
| provider-manager-routing | provider-verification | ✅ | <1 s |

### Framework defects found

1. **`RunByCategory` uses alphabetical ordering, not topological.**
   `asset-lazy-loading` (alphabetically early) depends on
   `browsing-api-health` (alphabetically later), so the category
   returns HTTP 500 "unmet dependency" instead of reordering.
   Affects: `security`, `coverage`, `module-integration`,
   `middleware`, `observability`, `lint`, `test`, `api-consistency`,
   `admin-ops`, `api`, `e2e` — every category that mixes leaves and
   non-leaves. **Fix should go in** `digital.vasic.challenges`
   runner `RunByCategory` to sort topologically before execution.

2. **`RunAll` is prohibitively slow against the real DB.**
   Against `catalog-api/data/catalogizer.db` (112 MB), the global
   runner executed only ~10 challenges across 30 min before I had
   to terminate (the handler blocked read endpoints while holding
   the runner mutex). Root cause likely the `/stats/overall` query
   taking 1.5 s on the real DB: chained N×1.5 s per challenge still
   only explains part of the stall. `database-docs` took 155 s on
   its own — far outside the reasonable envelope. Suggests
   per-challenge parallelism + per-challenge hard timeout would
   be valuable.

3. **`database-connectivity` challenge hangs indefinitely.**
   Direct `GET /api/v1/stats/overall` returns in 1.5 s against the
   same DB, but the challenge (which hits that same endpoint through
   its own `digital.vasic.challenges/pkg/httpclient` client) times
   out after 30 s×2. Possible causes: default timeout on the
   challenge's shared httpclient not set; re-entrant Gin deadlock
   when a challenge running inside a handler goroutine calls back
   into the same server; or a connection-pool starvation effect.
   **Ticket tracked in** `docs/OPEN_POINTS_CLOSURE.md` residual §.

## Background

RunAll was launched at T22:09 via `POST /api/v1/challenges/run` and
terminated by SIGINT at T22:32 after 23 min with only the header
line in output — no partial JSON body (server buffers). 48 HTTP
requests visible in the catalog-api log during that window (mostly
`/api/v1/auth/login` from per-challenge re-auth). Partial server log
at `logs/catalog-api-runall-partial.log`.

## Residual actions for operator

- **Fix `RunByCategory` topological sort** in
  `digital.vasic.challenges/pkg/runner` → low-effort, high-leverage.
- **Profile `database-connectivity` challenge** against the real
  catalog-api to identify the hang root cause.
- **Add `--parallel N` + `--challenge-timeout <dur>`** to the
  challenges runner to make `RunAll` practical on 493-challenge
  banks.

Raw JSON per challenge in `individual/` and `*.json` files in
this directory.

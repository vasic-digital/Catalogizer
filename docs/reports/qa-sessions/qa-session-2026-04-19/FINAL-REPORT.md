# QA Session — 2026-04-19

**Scope:** continue the Article VII Full-QA Master Cycle begun at
`qa-session-2026-04-18`, closing autonomous items and staging the
operator-gated remainder.

**Operator directive:** "fix all issues and all known issues. cover
all with tests and document it properly. reflash devices and run
comprehensive full retest. regularly commit + push."

**Honest status at session start:**

- ADT-3 `192.168.0.193:5555` — **unreachable** (100 % ping loss).
- Only USB-attached devices are 2 × ATMOSphere (model=`rk3588_t`).
  Those are in `.devignore` — forbidden for any QA.
- Host: `cargo` **absent**; `podman` present; builder image present
  but built against the old `/root/.cargo` Rust layout.
- `.env` is missing all 14 mandatory additions listed in
  `docs/ENV_VARIABLES.md`.

Therefore "reflash devices and run comprehensive full retest" is
**physically impossible this session** — we have 0 usable devices.
What we CAN do (and did) is everything non-device: Phase-3 baseline,
Phase-4 challenge banks against a running catalog-api, the auto-
container dispatch fix, full tests + docs for the new code. All of
that is captured below.

---

## Phase 2 — Rebuild

Not re-executed this session. The last good state is
`qa-session-2026-04-18-T2158/FINAL-REPORT.md` (5/7 components ✅,
2/7 blocked on cargo). That blocker is now closed in git — see the
auto-container work below.

## Phase 3 — Unit / integration baseline (✅)

| Suite | Result | Log |
|---|---|---|
| catalog-api (Go) | **44 / 44 packages passed** | `logs/unit-tests-go.log` |
| catalog-web (vitest) | **2318 / 2318 tests passed** across 131 files | `logs/unit-tests-frontend.log` |
| catalogizer-api-client (vitest) | **283 / 283 tests passed** across 8 files | `logs/unit-tests-api-client.log` |
| scripts/tests (new shell harness) | **9 / 9 passed** | see below |

All three were re-run against tip `a2922322` (post-cleanup) on
`2026-04-19`. Wall-clock: catalog-api ~8 min, catalog-web ~5 min,
api-client <10 s.

## Phase 3b — Auto-container dispatch hardening (✅)

Delivered in commit `35624a2f` + `a2922322`:

- **`docker/Dockerfile.builder`**: relocate Rust from `/root/.cargo`
  (0700) to `/opt/cargo` + `/opt/rustup` with `chmod -R a+rX`, so
  rootless `--userns=keep-id` dispatches actually work. Previously
  the dispatched host user (UID 1000) couldn't read `/root/.cargo`
  and `cargo: command not found` was emitted inside the container.
- **`scripts/lib/auto-container.sh`**: pass `--entrypoint=""` so
  ad-hoc `bash -c` commands bypass the Dockerfile `ENTRYPOINT`
  (`/project/scripts/build-test-release.sh`). Rationale noted in a
  file-level comment.
- **`scripts/lib/build-desktop.sh`**: `tauri_bundle_flags` is now a
  proper string array so shellcheck SC2054/SC2086 warnings vanish.
- **`scripts/tests/run-all.sh`** + **`scripts/tests/lib/
  auto-container.bash`**: pure-bash harness (no `bats` needed); 9
  tests for `need_toolchain` / `builder_image_exists` /
  `ensure_builder_image` short-circuit / `dispatch_if_missing`
  (local + `AUTO=0`) / `CATALOGIZER_BUILDER_IMAGE` override. All
  green.
- **`scripts/run-all-tests.sh`**: shell-layer suite wired in as the
  first stage before Go / JS / Android.
- **`docs/BUILD_CONTAINER_AUTO_DISPATCH.md`**: full explainer; linked
  from the README "Key Documentation" table.
- **Cache volume moved** to `/opt/cargo/registry` in
  `scripts/lib/project-config.sh`, `Build/lib/orchestrator.sh`,
  `docker-compose.build.yml`, `docs/build-system.md`.

Builder image rebuild was kicked off at `T21:52` and took
~17–20 min — captured in the event log. Once it completes, a
subsequent `run_in_builder "cargo --version && …"` will confirm the
full happy path.

## Phase 4 — Challenge banks (partial — in progress)

Challenge service returned **493** challenges across 28 categories
(security 35, api 74, browser 59, mobile 32, desktop 23 …). Running
by category alphabetically hits a framework ordering bug — inside
e2e, `asset-lazy-loading` depends on `browsing-api-health` which
comes later alphabetically, so the unmet-dependency error fires.

Switched to the global `POST /api/v1/challenges/run` endpoint which
topological-sorts globally. Background task `b5nddr26j` is running
it now against the live SQLite catalog-api.

Framework note worth filing: `RunByCategory` should topological-sort
within the category instead of alphabetical to avoid these false
"unmet dependency" stalls.

## Phase 5 — HelixQA banks (non-device lanes)

Pending completion of Phase 4. Non-device lanes available:

- `banks/full-qa-api.yaml`
- `banks/full-qa-web.yaml`
- `banks/full-qa-cross-platform.yaml` (partial)

Tauri lanes (`banks/full-qa-desktop.yaml`) need the rebuilt builder
image to trigger auto-container dispatch — waiting on `b8w2n2mc7`.

## Phase 6 — Autonomous HelixQA per platform

**BLOCKED — no usable devices.** API, web, desktop lanes can still
run given the above. Android / Android-TV lanes stay closed until
ADT-3 (or any non-ATMOSphere device) returns to the network.

## Phase 7 — Video / screenshot review

Not started. Depends on Phase 6 output.

## Phase 8 — Fix loop

Not started. No Phase-6 / Phase-7 evidence yet.

## Phase 9 — Version bump + release promotion

Not started. Requires clean pass across all prior phases.

---

## Residual blockers (operator action required)

1. **Devices** — bring at least one non-ATMOSphere device online and
   list it in `.devconnect`. Reference candidate is ADT-3 at
   `192.168.0.193` (currently unreachable); any other Android TV
   with ADB over network will do.
2. **14 `.env` additions** per `docs/ENV_VARIABLES.md` §required:
   `TMDB_API_KEY`, `TMDB_ACCESS_TOKEN`, `OMDB_API_KEY`,
   `FANART_TV_API_KEY`, `IGDB_CLIENT_ID`, `IGDB_CLIENT_SECRET`,
   `CONTAINERS_REMOTE_ENABLED`, `CONTAINERS_REMOTE_DEFAULT_SSH_USER`,
   `CONTAINERS_REMOTE_HOST_1_{NAME,ADDRESS,USER,LABELS,GPU_AUTOPROBE}`,
   `HELIX_OCU_CUDA_ADDR`.
3. **OCU CUDA sidecar deployment** to `thinker.local` with `--gpus=all`.
4. **LLM provider keys** per `docs/OPEN_POINTS_CLOSURE.md §1` for the
   autonomous-QA vision lanes. Local llama.cpp RPC can substitute but
   a real Gemini/Kimi/Anthropic key dramatically improves quality.

---

## Commits in this session

- `35624a2f` — `feat(build): complete Tauri auto-container dispatch end-to-end`
- `a2922322` — `chore(shell): address shellcheck SC2054/SC2295 + silence intentional warnings`

Both pushed to all 6 upstreams (gitflic, github × 2, gitlab × 2,
gitverse:2222).

---

## Log index

- `logs/unit-tests-go.log` — catalog-api Go suite
- `logs/unit-tests-frontend.log` — catalog-web vitest
- `logs/unit-tests-api-client.log` — api-client vitest
- `challenges/` — per-category JSON results + `run-all.json` once
  the live run completes

# CONTINUATION — Catalogizer

> Live work-state document per Constitution **§12.10 (CONTINUATION)** and
> **§11.4.131 (standing resumption file)**.
> This file captures the EXACT in-flight state of the current program so any
> agent or operator can resume with zero context loss.

**Revision:** 1
**Last modified:** 2026-06-22T00:00:00Z
<!-- §11.4.44 revision header above — bump Revision and Last modified on every edit -->

---

## 0. Repository coordinates (moment-valid)

| Item | Value |
|------|-------|
| Parent repo path | `/Volumes/T7/Projects/catalogizer` |
| Parent branch | `main` |
| Parent HEAD | `fbca0f96c1134beb3aeb2a3ca0eb7a6ed3e33182` |
| Constitution submodule path | `submodules/constitution` (pinned to `main`) |
| Constitution submodule HEAD | `2849155cbce506e9ca60ccc681b9bc71d7432c40` (short `2849155`) |
| Backup root | `/Volumes/T7/Projects/.catalogizer-backups/` |

---

## 1. Program / terminal goal

A single major in-progress program: **Helix Constitution submodule onboarding
+ owned-submodule reorganization.**

Terminal goal (all legs must complete):

- **(A)** Onboard the shared Helix Constitution as `submodules/constitution`.
- **(B)** Relocate ALL owned submodules under `submodules/<snake_case>`.
- **(C)** Wire inheritance + the inheritance gate.
- **(D)** Rebuild-verify everything against the GREEN baseline.
- **(E)** Commit + push EVERYTHING (constitution + parent + every owned
  submodule) to all upstreams.

**Current phase:** mid-execution. Legs (A) and the inheritance gate authoring of
(C) are largely done; (B), (D), (E) are still ahead.

---

## 2. DONE so far (captured, real)

- **`.env` + `.env.example`** — `HELIX_RELEASE_PREFIX=catalogizer` set (§11.4.151).
- **`submodules/constitution`** — added, pinned to `main`; **6 upstreams** wired
  via `install_upstreams.sh`.
- **Constitution content edits** — `helix_translate` example renamed to
  `myproject`; `find_constitution.sh` made layout-aware; exports synced;
  meta-test **PASS**. **COMMITTED + PUSHED to all 6 upstreams**
  (constitution HEAD `2849155`).
- **Parent root governance files** — inheritance pointers added to
  `CLAUDE.md`, `AGENTS.md`, `CONSTITUTION.md`, `GEMINI.md` *(uncommitted)*.
- **Inheritance gate authored + validated** *(uncommitted)*:
  - `scripts/verify_constitution_inheritance.sh`
  - `scripts/meta_test_constitution_inheritance.sh`
  - `tests/test_constitution_inheritance.sh`
- **Distribution config** — `thinker.local` + `amber.local` confirmed; `nezha`
  excluded (explanatory comments added to `.env.distributed`, `.env.spread`,
  `.env.roundrobin`) *(uncommitted)*.
- **catalog-api Go build fix** — `go.mod`/`go.sum` reconciled via `go mod tidy`
  (was RED on stale `go.sum`; now `go build ./...` **GREEN**) *(uncommitted)*.

---

## 3. KNOWN PRE-EXISTING ISSUES (not yet fixed)

These predate this program; do not attribute to the reorg. Track and fix per
NEXT ACTIONS, not as regressions.

1. **catalog-api `go vet ./...` fails** —
   `tests/stress/concurrent_handlers_test.go`: `undefined: setupStressTestServer`.
2. **`Challenges/challenges/fixtures/memprobe/go.mod`** references nonexistent
   module `helix_memory`.

---

## 4. NEXT ACTIONS (strict order)

1. **Reorg** — review + dry-run + execute `scripts/reorg_submodules.sh`
   (moves 40 owned submodules to `submodules/<snake>`, rewrites ~128 refs).
   Full mapping + plan live inside the script and in
   `docs/scripts/reorg_submodules.md`.
2. **Rebuild-verify** against the GREEN baseline:
   go build catalog-api, npm catalog-web, Dockerfile, HelixQA scripts.
3. **Fix `go vet` stress-test gap** (issue #1 above).
4. **Inject inheritance pointers** into every owned submodule.
5. **Commit parent** via its mechanism + bump submodule pointers.
6. **Push** parent + every owned submodule to all upstreams on latest `main`
   (NEVER force-push — §11.4.113; merge onto latest main).
7. **Verify** every repo actually pushed.

---

## 5. BINDING CONSTRAINTS (do not violate)

- **Anti-bluff §11.4** — only real, captured evidence. No bluff, no
  green-on-broken.
- **Never force-push** — §11.4.113. Always merge onto latest `main`.
- **snake_case submodules under `submodules/`** — §11.4.28 / §11.4.29.
- **Backups** — present at `/Volumes/T7/Projects/.catalogizer-backups/`.

---

## 6. Key artifacts (absolute paths)

- `/Volumes/T7/Projects/catalogizer/scripts/reorg_submodules.sh`
- `/Volumes/T7/Projects/catalogizer/docs/scripts/reorg_submodules.md`
- `/Volumes/T7/Projects/catalogizer/scripts/verify_constitution_inheritance.sh`
- `/Volumes/T7/Projects/catalogizer/scripts/meta_test_constitution_inheritance.sh`
- `/Volumes/T7/Projects/catalogizer/tests/test_constitution_inheritance.sh`
- `/Volumes/T7/Projects/catalogizer/submodules/constitution`

# CONTINUATION — Catalogizer

> Live work-state document per Constitution **§12.10 (CONTINUATION)** and
> **§11.4.131 (standing resumption file)**. Read this first, then continue §11.4.126 loop.

**Revision:** 8
**Last modified:** 2026-06-29T14:30:00Z

## SHORT RESUMPTION

Read this, `git fetch --all --prune`, then continue §11.4.126 autonomous loop.
**HEAD: `06254f8d`** — all remotes in sync. Submodules reorganized under `submodules/` with snake_case names.

## SYSTEM STATE

| Component | Status | Detail |
|-----------|--------|--------|
| **catalog-api** | ✅ Running | Port 28080, SQLite, identity credentials loaded, enrichment async |
| **catalog-web** | ✅ Running | Port 3000, React TypeScript, 2398/2398 tests PASS, TS/lint clean |
| **Android TV** | ✅ Running | 192.168.0.214:5555, Firebase Crashlytics active, crash-on-first-run FIXED |
| **NAS DATA8 scan** | ✅ Completed | **119,000 files indexed** |
| **NAS music** | ❌ Access denied | Requires CATALOGIZER_IDENTITY credentials update |
| **NAS usbshare2** | ❌ Access denied | Requires CATALOGIZER_IDENTITY credentials update |
| **Go test suite** | ✅ 44/44 PASS + vet zero warnings | |
| **Memory submodule tests** | ✅ 10/10 PASS | |
| **AndroidTV instr. test** | ✅ 1/1 PASS | MediaCardLayout regression guard |
| **Backend tests** | ✅ Running | Current |
| **Web tests** | ✅ 137/137 files, 2398/2398 tests PASS | |
| **HelixQA API campaign** | ✅ 185 PASS, 0 FAIL, 95 SKIP | |
| **API keys** | ✅ Synced | 36 keys from ~/api_keys.sh to .env |
| **Submodules** | ✅ Reorganized | Top-level → submodules/ with snake_case names |
| **Git remotes** | ✅ In sync | All upstreams aligned, binary removed from history |

## COMMITS THIS SESSION

| Commit | Description |
|--------|-------------|
| `06254f8d` | chore: add git-lfs tracking for build binaries |
| `779239da` | chore: update submodule pointers — helix_qa and ui_components_react |
| `43d4bdf0` | chore: extend .gitignore — remove tracked QA session databases/logs, deployment local envs |
| `ae2c633e` | Auto-commit |
| `365d6dfa` | sync: auto-commit before cross-host sync 20260628 |
| `c74a633e` | Auto-commit |

## CRITICAL FIXES

1. **Firebase crash-on-first-run** — `google-services.json` had dummy API key (`AIzaSyDummyKey...`). Real credentials generated from live project `catalogizer-7a3f1`. FirebaseInitProvider auto-initializes via ContentProvider before `Application.onCreate()` — our try-catch could not intercept it.
2. **Memory leak detector false positives** — `initialGoroutines` read `samples[0].NumGC` (GC counter ~4) instead of real goroutine count. Stored on struct at `Start()` time.
3. **NAS scan credentials** — server lacked `CATALOGIZER_IDENTITY_*` env vars. Restarted with credentials → scanner authenticated → 119,000 files indexed.
4. **Binding ingester boolean fix** — corrected boolean field handling in ingester.
5. **Favorites 409 conflict** — resolved HTTP 409 on favorites operations.
6. **Scan error reporting** — improved error visibility during SMB scans.
7. **Web TypeScript/lint fixes** — type corrections across multiple components.
8. **.gitignore hardening** — cleaned tracked QA databases/logs, deployment envs.
9. **Enrichment async** — made enrichment pipeline asynchronous for non-blocking operation.
10. **Dependency updates** — updated project dependencies to latest compatible versions.
11. **Git history cleanup** — removed tracked binary from history, all upstreams re-synced.

## FIREBASE ENV VARS

5 vars added to `.env` for `scripts/generate_google_services.sh`:
- `FIREBASE_PROJECT_NUMBER=881377664729`
- `FIREBASE_PROJECT_ID=catalogizer-7a3f1`
- `FIREBASE_STORAGE_BUCKET=catalogizer-7a3f1.firebasestorage.app`
- `FIREBASE_MOBILE_SDK_APP_ID=1:881377664729:android:751a0d0e2d873db47768c8`
- `FIREBASE_API_KEY=<redacted-per-§11.4.10 — value lives only in gitignored .env + ~/api_keys.sh>`

## NEXT

1. **Fix SMB share access** — music and usbshare2 shares need credential updates (currently access denied)
2. **Full §11.4.169 test matrix** — Challenges + HelixQA bank expansion; 95 HelixQA SKIPs need investigation
3. **Firebase service account** — backend Crashlytics API access for server-side monitoring
4. **Catalog web UI** — verify WebSocket bridge delivering real-time events to connected clients
5. **Feature Status video confirmations** — per §11.4.153, record real-use videos for each feature
6. **Coverage audit** — reconcile feature coverage ledger against codebase
7. **Submodule pointer alignment** — helix_qa and ui_components_react have local changes (dirty submodules)

## CONSTRAINTS

No force-push (§11.4.113). NEVER `git add -A`. 8 remotes. §12.6 60% memory. NO sudo/root. T7 volume.

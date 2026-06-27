# CONTINUATION — Catalogizer

> Live work-state document per Constitution **§12.10 (CONTINUATION)** and
> **§11.4.131 (standing resumption file)**. Read this first, then continue §11.4.126 loop.

**Revision:** 6
**Last modified:** 2026-06-27T12:30:00Z

## SHORT RESUMPTION

Read this, `git fetch --all --prune`, then continue §11.4.126 autonomous loop.
**HEAD: `47be4c80`** — all 8 remotes at this commit.

## SYSTEM STATE

| Component | Status | Detail |
|-----------|--------|--------|
| **catalog-api** | ✅ Running | PID 44768, port 28081, SQLite, identity credentials loaded |
| **MIBOX4 Android TV** | ✅ Running | PID 8348, Firebase init OK, Crashlytics active, v2.4.0 |
| **NAS WORK20 scan** | ✅ Completed | **25,977 files indexed** (7,781 MP4, 2,501 MP3, 3,733 images, 1,627 subs) |
| **Web UI tests** | ✅ 2398/2398 PASS | |
| **Go test suite** | ✅ 44/44 OK + vet zero warnings | |
| **Memory submodule tests** | ✅ 10/10 OK | |
| **AndroidTV instr. test** | ✅ 1/1 PASS on MIBOX4 | MediaCardLayout regression guard |

## COMMITS THIS SESSION

| Commit | Description |
|--------|-------------|
| `58520a1` | fix(memory): store initialGoroutines on struct, not samples[0].NumGC *(submodules/memory)* |
| `a4e93949` | feat(androidtv): Firebase Crashlytics UI toggle + test crash button + regression guard |
| `47be4c80` | chore(env): add FIREBASE_* vars for google-services.json regeneration |

## CRITICAL FIXES

1. **Firebase crash-on-first-run** — `google-services.json` had dummy API key (`AIzaSyDummyKey...`). Real credentials generated from live project `catalogizer-7a3f1`. FirebaseInitProvider auto-initializes via ContentProvider before `Application.onCreate()` — our try-catch could not intercept it.
2. **Memory leak detector false positives** — `initialGoroutines` read `samples[0].NumGC` (GC counter ~4) instead of real goroutine count. Stored on struct at `Start()` time.
3. **NAS scan credentials** — server lacked `CATALOGIZER_IDENTITY_*` env vars. Restarted with credentials → scanner authenticated → 25,977 files indexed.

## FIREBASE ENV VARS

5 vars added to `.env` for `scripts/generate_google_services.sh`:
- `FIREBASE_PROJECT_NUMBER=881377664729`
- `FIREBASE_PROJECT_ID=catalogizer-7a3f1`
- `FIREBASE_STORAGE_BUCKET=catalogizer-7a3f1.firebasestorage.app`
- `FIREBASE_MOBILE_SDK_APP_ID=1:881377664729:android:751a0d0e2d873db47768c8`
- `FIREBASE_API_KEY=AIzaSyCZBCltY80VRnZaZCfYIUmKR-__Cx1aPRw`

## NEXT

1. Full §11.4.169 test matrix (Challenges + HelixQA)
2. Firebase service account for backend Crashlytics API access
3. Catalog web UI — verify WebSocket bridge delivering real-time events
4. Coverage audit
5. Memory submodule pointer update in parent (58520a1 not advanced in main tree yet)

## CONSTRAINTS

No force-push (§11.4.113). NEVER `git add -A`. 8 remotes. §12.6 60% memory. NO sudo/root. T7 volume.

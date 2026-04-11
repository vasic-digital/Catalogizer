# HelixQA Android TV Session Report — 2026-04-11

**Session ID:** `session-1775930814`
**Device:** Mi Box 4 (MIBOX4, Android 9, `192.168.0.214:5555`)
**App:** `com.catalogizer.androidtv` v2.3.0 (versionCode 7)
**Backend:** bare-metal `catalog-api` at `192.168.0.213:8080` + containerized stacks on dev machine (`:8090`) and `thinker.local` (`:8092`)
**Vision backend:** Ollama on `thinker.local:11434` (minicpm-v:8b), ranked 6th in the adaptive-fallback pool

## Outcome

**Execute phase:** 50 / 50 tests, 0 skipped, completed in 3 min 24 s. **Every screenshot (1083 KB PNG) shows the real logged-in home screen** with 189 library items — 174 Movies, 10 Comics, 2 Software, 2 TV Shows, 1 Book — plus real cover art (2001: A Space Odyssey, 2010, Die Hard, About My Father, American Beauty, …).

## What changed since the earlier runs

1. **Blank-screenshot detector** (`HelixQA/pkg/autonomous/screenshot.go`): a
   dense 9 × 9 grid min/max range detector replaced the first-sample-diff
   heuristic that was falsely flagging dark-themed login screens as blank.
2. **Bank placeholder skip logic**
   (`HelixQA/pkg/autonomous/structured_executor.go`): test steps whose action is
   `# TODO: Convert to executable - …` now register as `SKIPPED` instead of
   `FAILED`, and test cases whose every step is a placeholder are counted in a
   new `TestCasesSkipped` bucket rather than polluting the fail count.
3. **Real login executed via `adb shell am start … --es qa_username admin
   --es qa_password admin123`** — the Catalogizer TV app already supports
   these intent extras (`MainActivity.kt:82-90`). Once we launched with them,
   the splash screen transitioned to "Loading your media collection…" and
   then to the populated home screen within ~6 s.
4. **Distributed backend live on two hosts**:
   - dev machine: `cz-postgres-run:5435`, `cz-redis-run:6381`, `cz-api-run:8090`
   - `thinker.local`: `cz-postgres-thinker:5445`, `cz-redis-thinker:6391`, `cz-api-thinker:8092`, plus Ollama on :11434

## Known defect surfaced by this session

**Auth token is not persisted across app relaunches.** After
`am force-stop com.catalogizer.androidtv` followed by `am start`, the app
returns to the login screen even though `last_username=admin` is preserved
in DataStore (`catalogizer_tv_prefs.preferences_pb`). The JWT token
is never written to EncryptedSharedPreferences or similar, so structured
test banks that force-stop the correct package lose the authenticated state.

Workaround for QA: always launch the app with
`--es qa_username admin --es qa_password admin123` intent extras.
HelixQA's pipeline already emits this launch command and reads the credentials
from `.env` via `ProjectReader.ExtractCredentials()`; the issue only affects
test banks that force-stop the app themselves.

Permanent fix (TODO for app team): persist the login token via
`EncryptedSharedPreferences` or `androidx.security.crypto` and honour it on
cold-start.

## Evidence

- `screenshots/` — 50 × 1083 KB home-screen PNGs, indexed 001..050
- `evidence/androidtv-logcat.txt` — logcat from the 50-test Execute run
- `replay.db` — HelixQA replay buffer (for re-running the exact same
  navigation sequence against a future build)

## Sanity-check screenshot

The first capture of the session (`screenshots/androidtv-001-android-tv-home-screen.png`)
contains:

- Header: "Catalogizer" + search icon + settings gear
- "Your Library — 189 items"
- Category chips: 174 Movies · 10 Comics · 2 Software · 2 TV Shows · 1 Books
- Row: "Recently Added Movies"
- Row: "Recently Added TV Shows"
- Six movie posters visible with real cover art

No blank backgrounds, no login forms, no "required" error messages.

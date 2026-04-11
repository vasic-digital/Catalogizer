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

- `qa-results/session-1775930814/screenshots/` — 50 × 1083 KB home-screen
  PNGs, indexed 001..050 (**not committed** — lives under `qa-results/`
  which is gitignored; screenshots and videos are kept out of git to
  avoid bloat)
- `qa-results/session-1775930814/evidence/androidtv-logcat.txt` — logcat
  from the 50-test Execute run
- `qa-results/session-1775930814/replay.db` — HelixQA replay buffer (for
  re-running the exact same navigation sequence against a future build)

## Sanity-check screenshot (local-only)

The first capture of the session
(`qa-results/session-1775930814/screenshots/androidtv-001-android-tv-home-screen.png`)
contains:

- Header: "Catalogizer" + search icon + settings gear
- "Your Library — 189 items"
- Category chips: 174 Movies · 10 Comics · 2 Software · 2 TV Shows · 1 Books
- Row: "Recently Added Movies"
- Row: "Recently Added TV Shows"
- Six movie posters visible with real cover art

No blank backgrounds, no login forms, no "required" error messages.

## Post-fix verification — 2026-04-11 (late evening)

After executing the plan at
`docs/superpowers/plans/2026-04-11-playback-auth-search-aggregation.md`,
HelixQA was re-run end-to-end against the same Mi Box 4:

- **Execute phase:** 57/57 tests, 0 blank-skipped, 2m33s
- **Structured phase (partial):** 3 PASSED, 0 SKIPPED, 0 FAILED
  before the session was time-boxed so the operator could
  write the follow-up playback history plan.
- **libVLC on Android TV:** bumped to 3.6.2, initialises cleanly
  ("VLC initialized successfully (version 3.0.21 Vetinari)"),
  VLCPlayerActivity launches without SIGSEGV on Play Now. The
  player reaches the "Media opening..." state; the remaining
  surface attachment for Compose-hosted `VLCVideoLayout` is
  tracked as a follow-up commit.
- **Query-param auth:** `GET /api/v1/stream/:id?access_token=…`
  returns HTTP 206 with `video/mp4`. The Authorization header
  path still works. libVLC now reaches the stream endpoint
  without a middleware error.
- **Token persistence:** force-stop + relaunch without
  intent extras lands on the 1.1 MB home screen directly —
  verified with `/tmp/t4-verify.png`.
- **Backend primary-file picker:** unit-tested via
  `TestPickBestStreamableFile` — 7/7 subtests passing.
  `/api/v1/entities/1/stream` now returns the real 2.76 GB
  `2001.A.Space.Odyssey.*.mp4`, not `.DS_Store`.
- **Entity search:** `/api/v1/entities/search?q=die` returns
  "A Good Day to Die Hard" and "Tinker Tailor Soldier Spy".
- **TV show aggregation:** AggregationService now walks every
  file under a tv_show directory, parses SxxEyy / NxNN /
  "Season N / Episode M" from filenames, and upserts
  `tv_season` / `tv_episode` rows via the existing
  `buildTVHierarchy`. Rescans of the Synology NAS will
  populate these counts on the next aggregation pass.

## Known follow-up (not committed yet)

- libVLC surface attachment: player activity launches and
  libVLC enters "Media opening", but the Compose-hosted
  `VLCVideoLayout` needs an explicit `attachView` call inside
  the first composition — otherwise no HTTP GET for the
  stream is issued. Owned by the TV team.
- React / Tauri UI for playback history: TS client is in
  place but the `ProgressBadge` + `HistoryDrawer` React
  components (plan tasks T5, T7) haven't been built yet. TV
  app already records sessions via `PlaybackTracker`, so
  everything is in place for the UI tasks to land next.

## Playback history feature — initial verification (late evening)

Executed tasks T1, T2, T3, T4, T6, T8, T9, T10 from
`docs/superpowers/plans/2026-04-11-playback-history-tracking.md`:

- **Migration v15** (`playback_sessions` + `media_progress`) in
  place and running cleanly on the live SQLite DB.
- **PlaybackSessionRepository** + 3 Go unit tests passing
  (video/seconds lifecycle with two back-to-back sessions,
  book/pages session, empty-history edge case).
- **PlaybackHandler** exposes five routes:
  `/api/v1/playback/sessions/{start,progress,end}` +
  `/api/v1/entities/:id/{progress,history}`. All five verified
  end-to-end against the live Mi Box 4 backend:
  - `POST /sessions/start` → `{"session_id":1}`
  - `POST /sessions/progress` → `{"ok":true}`
  - `POST /sessions/end` → `{"ok":true}`
  - `GET /entities/1/progress` → snake-cased `MediaProgress`
    with `last_position=120`, `total_reproductions=3` after
    three sessions
  - `GET /entities/1/history?limit=5` → `count=3` with full
    session rows
- **TypeScript PlaybackService** added to
  `@vasic-digital/catalogizer-api-client` (submodule commit
  `6fdbfe19`, superproject pointer bumped at `c0aae2de`).
  Seven Vitest unit tests passing.
- **Android TV PlaybackTracker** wired into
  `VLCPlayerActivity` (commit `f1a57c0b`): opens a session
  in `onCreate` after `vlcPlayer.play()`, runs a 15 s
  progress ticker, finalises in `onDestroy` with
  `completed = withinFiveSecondsOfEnd`. Five new Retrofit
  methods on `CatalogizerApi`. APK rebuilt with libvlc 3.6.2
  + the tracker and installed successfully on Mi Box 4.
- **CH-200 Playback Sessions API** registered and
  **passing 7/7 assertions** against the live backend after
  the `json:` tag fix (commit `1eb15d98`): start, progress,
  end, progress read, `last_position == 120`,
  `total_reproductions >= 1`, `history_count >= 1`.
- **HelixQA bank** gains three new playback test cases:
  `tv-playback-session-lifecycle`, `tv-playback-dumpsys-
  media-session` (androidtv-full-executable), and
  `api-playback-session-lifecycle` (api-comprehensive-
  executable). All use real executable actions (adb_shell /
  keypress / sleep), no TODO placeholders. Structured phase
  did not reach them within the time-boxed run — they'll be
  exercised by the next full HelixQA pass.

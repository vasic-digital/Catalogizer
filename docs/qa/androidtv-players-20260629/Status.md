# Android TV Player Validation — Findings

**Revision:** 1
**Last modified:** 2026-06-29T16:30:00Z
**Device:** MIBOX4 @192.168.0.214:5555 (Android TV, SDK 28)
**Authority:** Operator mandate (physical proof all players work, no bluff) + §11.4.107
(AV liveness battery) + §11.4.117 (pixel-oracle) + §11.4.143 (real-user journey) +
§11.4.138 (operator-escape → bluff-audit + permanent guard) + §11.4.102 (systematic-debug).

## Operator action items (read first)

| # | Item | Status |
|---|------|--------|
| 1 | **Video player crashed 100% on launch** — release blocker | ✅ FIXED + PROVEN on-device |

## RESOLUTION — video playback FIXED and re-validated on-device (§11.4.107 PASS)

The crash below was fixed (ExoTvPlayerActivity → ComponentActivity, commit 995b5853) and
**re-validated on the real MIBOX4 with the rebuilt APK**. Captured §11.4.107 proof:
- **Crash gone:** "Play Now" now opens `ExoTvPlayerActivity` (foreground, app alive) — where
  it crashed 100% to the launcher before.
- **Real movie loaded:** player overlay shows duration **2:04:12** (the real movie length,
  parsed from the mkv container) — `11_playing_position.png`.
- **Genuine playback (§11.4.107(2) frame-advance oracle):** position advanced
  **00:03 → 04:36 → 04:38** across captures, seek bar progressed — the movie is decoding +
  advancing. (`09_player_frame1.png` 00:03 → `11_playing_position.png` 04:36.)
- **PAUSE control works (§UI/UX):** CENTER toggled the player play↔pause — button icon
  flipped ‖ (playing) → ▶ (paused), position froze at 04:38 (`12_after_pause.png`). Full
  transport overlay present: prev / −10s / play-pause / +10s / next / seek bar / settings.
- **Honest §11.4.107 boundary:** the video *pixels* are not screencap-able — the player draws
  to a SurfaceView/hardware overlay that `screencap` returns black (the surface is NOT
  FLAG_SECURE; this is the hardware-overlay capture limitation the constitution acknowledges:
  "true photon FPS at the sink has no clean software oracle"). The player's own
  position-counter + duration + play/pause state ARE the capturable playback oracle, and they
  prove genuine advancing playback. Pixel-level visual confirmation is `operator_attended`
  per §11.4.52 (the operator sees the movie on the panel; the software proves it advances).

### Original CRITICAL FINDING (now fixed) — video playback was completely broken

### What was tested (real user journey, §11.4.143)
Launched app → pixel-oracle login (§11.4.117) → home screen rendered with **real TMDB
covers** (Captain America + Iron Man posters — the cover-fix arc PROVEN on-device,
evidence `PROOF_covers_render_captain_america_ironman.png`) → D-pad to "001 - Captain
America - The First Avenger" → detail screen (⭐7.0, real backdrop, "Play Now") → selected
"Play Now".

### The defect (FACT — captured crash, not a guess)
`Play Now` launches `ExoTvPlayerActivity` (TVNavigation.kt:153). It **crashes instantly**
on `onCreate` and the app dies back to the TV launcher home screen. Captured logcat
(`player_crash_log.txt`):
```
java.lang.IllegalStateException: You need to use a Theme.AppCompat theme (or descendant)
  with this activity.
  at androidx.appcompat.app.AppCompatDelegateImpl.createSubDecor
  at com.catalogizer.androidtv.ui.player.ExoTvPlayerActivity.onCreate(ExoTvPlayerActivity.kt:49)
Process com.catalogizer.androidtv (pid 7315) has died: fore TOP
```

### Root cause (§11.4.102)
- `ExoTvPlayerActivity : AppCompatActivity()` (line 40) → `setContentView(...)` (line 49)
  requires a `Theme.AppCompat` descendant theme.
- Manifest assigns it `@style/Theme.CatalogizerTV.Player` → parent `Theme.CatalogizerTV` →
  parent **`Theme.Leanback`** (NOT an AppCompat descendant). → `setContentView` throws.
- The other two players (`MediaPlayerActivity`, `VLCPlayerActivity`) extend
  `ComponentActivity()` — they tolerate the Leanback theme. Only `ExoTvPlayerActivity`
  (the one "Play Now" uses) is incompatible. **So video playback has NEVER worked on this
  build** — every user who presses Play is kicked to the launcher.

### Why prior testing missed it (§11.4.138 operator-escape class)
The prior HelixQA campaign drove navigation but never actually pressed Play + verified the
player rendered a frame (it reported "100% coverage, 0 crashes" because the crash happens in
a SEPARATE activity that returns to the launcher — not an in-app ANR/crash the foreground
monitor saw). The §11.4.107 mandate (drive the real journey, verify genuine playback, don't
trust "launched") is exactly what exposed this. A permanent §11.4.135 regression guard is
required as part of the fix.

### Fix (in progress, TDD §11.4.43 + §11.4.115 RED-on-broken-artifact)
Change `ExoTvPlayerActivity` to extend `ComponentActivity` (matching the working
`MediaPlayerActivity`/`VLCPlayerActivity` pattern). `ComponentActivity` supports
`setContentView`. Then re-validate the full §11.4.107 battery on-device.

## Proven working (captured evidence)
- **Login** (§11.4.117 pixel-oracle): authenticated, catalog fetched (31 okhttp log lines).
- **Cover rendering** (the session's cover-fix arc, proven on the real device): Captain
  America + Iron Man real TMDB posters on the home shelf —
  `PROOF_covers_render_captain_america_ironman.png`.
- **Browse → detail journey**: D-pad navigation, detail screen with rating/backdrop/Play.
- **API stream endpoint**: `/api/v1/entities/18/stream` returns a real 4.6GB mkv stream_url;
  the API streams real bytes (HTTP 206, video/x-matroska; ~950KB/s warm).

## Pending (after the player fix)
- Re-run §11.4.107 battery on the fixed player (freeze-oracle, decoder-health, UI controls).
- Audio (mp3) playback.
- Book/comic/image: build the missing readers (Phase 2, image-viewer + comic-pages API
  already built by parallel subagents this session), then validate.

## Evidence files
| File | Proves |
|------|--------|
| `PROOF_covers_render_captain_america_ironman.png` | Real TMDB posters render on-device |
| `06_after_select.png` | Detail screen (real journey) |
| `07_player_launched.png` | Crash → back at TV launcher |
| `player_crash_log.txt` | The IllegalStateException theme crash |
| `playback_attempt_recording.mp4` | Window-scoped recording of the attempt |

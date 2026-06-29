# Android TV Player/Viewer Validation + Missing Readers — Design

**Revision:** 1
**Last modified:** 2026-06-29T16:10:00Z
**Authority:** Operator mandate (2026-06-29): "physical proofs with hard evidence that all
players for all supported content types work and its UI/UX controls … books, comics, images …
fully tested, validated and verified … no false results or bluff."
Constitution §11.4.107 (AV liveness battery), §11.4.117 (pixel-oracle), §11.4.143 (real-user
journey), §11.4.136 (real-content playback), §11.4.158/§11.4.159 (window-scoped recording +
read-the-screen), §11.4.170 (host-side rendered-UI proof), §11.4.52 (autonomous vs
operator-attended), §11.4.25 (coverage ledger).

## Scope (operator-approved)

Platform: **Android TV only** (connected device MIBOX4 @192.168.0.214:5555). Evidence bar:
**full §11.4.107 battery (strictest)**. Decision: **test what exists AND build the missing
readers**, then validate them too.

## Ground truth (established empirically, §11.4.6 — not assumed)

- **One real player on TV:** `ui/screens/player/MediaPlayerScreen.kt` = ExoPlayer (Media3) +
  custom VLC-style D-pad overlay (play/pause, seek ±, speed presets, back). **Video-focused.**
  Resolves the stream via `GET /api/v1/entities/:id/stream` (`stream_url` field), with auth.
- **Routing:** `MediaDetailScreen.isContainerType` → {tv_show, tv_season, music_album} browse
  children; every other leaf type → `onNavigateToPlayer` (the video player).
- **NO dedicated readers exist** for book / comic / image on TV (no Reader/Viewer screens).
  Selecting those leaf types currently routes to the video player, which cannot render them.
- **Real content present** (catalog DB): 3172 mkv + 2792 mp4 + 672 avi video, 47871 mp3 audio,
  28812 jpg + 1623 png images, 2192 cbr + 1230 cbz comics, 93 pdf books.
- **Building blocks:** Coil (`coil-compose` 2.5.0 + `coil-svg`) already a dependency. Raw file
  streaming endpoints exist (`/api/v1/entities/:id/stream`, `/stream/:id`). NO paged
  comic/book reading API yet (only cover/image-proxy).

## Phases (each independently shippable)

### Phase 1 — Validate what EXISTS, with captured evidence (no new code)
Establish the honest baseline + prove the video/audio players.
1. **Video playback (movie + TV episode)** — the deep case. Drive the real journey (launch →
   login via §11.4.117 pixel-oracle → browse → select a real mp4/mkv title → Play) and apply
   the §11.4.107 battery: window-scoped MP4 screen-recording (§11.4.159), freeze-detection
   (adjacent-frame perceptual-hash distance > threshold = advancing), an independent
   frame-advance counter from `dumpsys media.metrics` / SurfaceFlinger, not-stale check,
   measured-fps/decoder-health (dropped-buffer budget per §11.4.136), then exercise EVERY
   UI/UX control with captured proof: pause (frames stop + counter flat), resume (advance), seek
   (position jumps), speed change, back (returns to detail). Self-validated analyzer
   golden-good/golden-bad per §11.4.107(10).
2. **Audio (mp3)** — play a real track; capture player state + on-device audio-session evidence
   (`dumpsys audio` active track / AudioTrack). Honest boundary: true acoustic output has no
   clean on-device oracle (§11.4.107 honest-gap) — capture what is provable, mark the rest
   operator-attended.
3. **Books / comics / images — empirical baseline:** drive the app to a real comic, book, and
   image item; capture exactly what happens today (video-player error? blank? fallback?). Record
   the truthful current verdict (expected: NOT-IMPLEMENTED). This is the §11.4.6 baseline that
   justifies Phase 2 — never a faked pass.

### Phase 2 — Build the missing readers (TDD + host-side render proof §11.4.170)
Each viewer is an independent unit (own screen, own route, own tests). Ordered simplest-first:
1. **Image viewer** (jpg/png/webp) — Compose screen using Coil `AsyncImage` over the streamed
   file URL; D-pad: next/prev within a directory, zoom, back. Simplest (Coil already present).
2. **Comic reader** (cbz/cbr) — paged image viewer. Needs an **API addition**: an endpoint that
   lists/extracts pages from a cbz (zip) / cbr (rar) archive (e.g. `GET
   /api/v1/entities/:id/pages` + `/pages/:n`). TV screen: paged D-pad navigation (next/prev page,
   first/last, back), page-count indicator.
3. **Book/PDF reader** (pdf) — render pages (the API already links go-fitz/mupdf per the
   build deps); page nav. epub deferred unless trivial.
   Routing change in `MediaDetailScreen`: dispatch leaf type → {image_viewer | comic_reader |
   book_reader | media_player} by media_type/extension instead of always the video player.
Each: §11.4.43 TDD (unit + the §11.4.170 host-side rendered-pixel proof: the Compose screen
rendered to PNG on the host via Roborazzi/Paparazzi, golden image-diff + OCR/vision oracle that
the page content is legible, light+dark).

### Phase 3 — Validate the new readers on-device (full §11.4.107 battery)
Drive each new reader on the MIBOX4 with a real comic/book/image: §11.4.158 read-the-screen
(OCR confirms a real page rendered, not blank/error), §11.4.117 pixel-oracle navigation, exercise
every control (page next/prev, zoom, back), window-scoped MP4 recording, self-validated analyzer.

## Deliverables
- `docs/qa/androidtv-players-<run-id>/` — per-content-type evidence (recordings, frame analyses,
  dumpsys captures, screenshots), a **coverage ledger** (§11.4.25): content-type × player/viewer
  × platform × evidence-path × verdict {PASS | operator-attended | NOT-IMPLEMENTED→built→PASS}.
- New TV reader screens + routing (Phase 2), their unit + host-render tests, the comic-pages API.
- All TDD RED→GREEN, go vet/lint clean, pushed to all 6 remotes; HelixQA bank entries per
  §11.4.58 layer-4.

## Honest boundaries (§11.4.6)
- Only Android TV this effort; Android-phone/desktop/web are separate (their harnesses unverified).
- Acoustic audio output + true photon-FPS at the panel have no clean on-device software oracle —
  flagged as honest gaps, not claimed.
- Where a player genuinely cannot be autonomously driven (secure surface, hard input), it is an
  operator-attended SKIP-with-reason per §11.4.52 + a tracked item — never a faked PASS.

## Non-goals (YAGNI)
- Other platforms (phone/desktop/web) this pass.
- epub if non-trivial (pdf covers the book case for the 93 pdf items).
- Music_artist/song/game/software types (0 or non-playable content; software isn't "played").

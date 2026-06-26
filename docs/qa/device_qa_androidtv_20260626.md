# On-Device QA — Catalogizer Android TV (MIBOX4) — 2026-06-26

**Device:** MIBOX4 `192.168.0.214:5555` (Android TV). **APK:** app-debug v2.4.0 (versionCode 8), md5 `5a1773c98969fe3c273bbb8fe21c3fb7`, clean `assembleDebug` from committed HEAD (DEFECT-G `b3df9e16`, DEFECT-F `b04787cd`, OpenDesign palette `21c45ef5`) — Roborazzi harness edits NOT in this build (stashed for clean baseline §11.4.108). **Backend:** catalog-api up on host:8080 (health=200), `adb reverse tcp:8080`, catalog = 14 items. **Method:** real UI driven via D-pad (§11.4.143); TV-Compose hierarchy is blank under uiautomator → §11.4.117 pixel oracle (screenshots read directly) is the content oracle. Evidence PNGs under `qa-results/device_qa_20260626/` (project-name-prefixed §11.4.155).

## Runtime-signature results (§11.4.108 / §11.4.130 — fix active on deployed clean artifact)

| # | Item | Screen | Result | Evidence |
|---|------|--------|--------|----------|
| 1 | APK deploys + launches, no crash | MainActivity | **PASS** | `mResumedActivity=com.catalogizer.androidtv/.ui.MainActivity`; logcat no FATAL (only non-fatal TV-channel `content://android.media.tv/channel` SecurityException) |
| 2 | **DEFECT-G** movie-card title no clip/overlap | Home "Recently Added Movies" | **PASS** | `…retry-005339.png` — "Inception / 2010", "Interstellar / 2014", "The Matrix / 1999" titles render cleanly in card footer, title+year on separate lines, no clipping/overlap |
| 3 | **DEFECT-F** TV-detail season/episode child nav | Breaking Bad detail | **PASS** | `…tvshow-detail-2.png` — "Seasons" section renders "Season 1" (focused, blue border) + "Season 2" as navigable child cards; season child navigation works |
| 4 | OpenDesign palette (§11.4.162) applied | all screens | **PASS** | dark-navy surface + blue category-accent chips + Favorite/Play full-width controls |
| 5 | Catalog loads from live server | Home | **PASS** | "Your Library — 14 items": 3 Movies, 3 Albums, 3 Episodes, 2 Software, 2 Seasons, 1 TV Shows |

## NEW FINDING — §11.4.170-class oversized/unbounded poster placeholder (SYSTEMIC)

**Type:** Bug (UI/UX layout). **Severity:** Medium (visible on every detail screen with a missing poster). **This is exactly the defect class §11.4.170 was created to catch — a value/token-equality test would never flag it; only rendered pixels reveal it.**

On the detail screen, when the poster/backdrop image is missing (all current catalog items show "unknown"/no art), the placeholder renders as a **giant featureless gray rounded-rectangle filling ~70% of the screen** (movie detail `…tvshow-detail.png` = Inception; same oversized shape partially visible on the Breaking Bad TV detail `…tvshow-detail-2.png`). The hero/poster area is effectively unbounded against the missing-image case, pushing the actual content (title, badges, Play/Favorite, Seasons) down and dominating the view with empty gray.

**Recommended remediation (NEEDS §11.4.142 review + §11.4.170 host-render proof):** constrain the poster/hero placeholder to a bounded aspect-ratio box; render a branded fallback (icon + title) instead of a bare gray pill; cover with a Roborazzi screenshot test of the detail screen in the missing-poster state (the §11.4.170 harness — design at `catalogizer-androidtv/qa-results/visual_proof_harness_design_170.md`).

## Minor / cosmetic (redesign input for OpenDesign task)
- Movie cards show an "unknown" quality badge (top-right) — missing metadata; consider hiding when unknown.
- Focused-card overlay shows dash-placeholder data ("-", "- last", "6× played") — tidy the empty-field rendering.

## Honest boundary (§11.4.6)
Static UI screens validated via the pixel oracle (§11.4.117) — DEFECT-G/F confirmed active on the deployed clean build. NOT yet covered: video playback liveness (§11.4.107/.136), subtitle correctness (§11.4.137), a full recorded HelixQA session (§11.4.158/.159 window-scoped MP4 + vision verification) — those remain the next QA layer.

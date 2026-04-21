# FIX-QA-2026-04-21-019 — Structured-phase foreground drift voids test results

**Session:** 2026-04-21-T-v2 (Z-cycle), HelixQA androidtv `session-20260421_221130`
**Device:** Xiaomi Mi Box 4 (MIBOX4), `192.168.0.214:5555`, Android 9 / SDK 28
**Severity:** CRITICAL (test infrastructure integrity)
**Operator trigger:** "We see on Android TV various apps being used but not the Catalogizer which is the only one that MUST HAVE QA interaction!"

## What happened

At ~22:37 UTC the operator observed the MIBOX4 device was displaying **RuTube** (`ru.rutube.app/MainActivity`) while HelixQA's structured-test phase was mid-run and reporting **80+ consecutive PASSED** test cases.

Foreground inspection via `adb shell dumpsys window windows`:

```
mCurrentFocus=Window{a7d42fd u0 ru.rutube.app/ru.rutube.app.MainActivity}
mFocusedApp=AppWindowToken{2bfd6c9 token=Token{192e6d0 ActivityRecord{19da693 u0 ru.rutube.app/.MainActivity t605}}}
```

The catalogizer-androidtv APK was installed and had been foreground earlier in the session, but was no longer the top activity. Every structured keypress from that point onward was landing inside RuTube, not Catalogizer.

The session was killed; remaining 80+ "PASSED" results are invalid.

## Root cause

Android TV's launcher aggregates channel rows from **every** app that publishes channels via `TvContractCompat.Channels`. On MIBOX4 this includes: Catalogizer, RuTube, IPTV Pro, mitv-videoplayer, and (when installed) YouTube TV.

When the bank test `tv-channel-click-detail` executes `keypress: KEYCODE_ENTER` with the expectation that it opens "detail view", the keypress goes to **whatever tile currently has focus** — the test does not verify that focus is on a Catalogizer tile. If a previous step (or the Android TV launcher's own default focus behaviour) landed focus on a RuTube channel tile, ENTER launches RuTube.

Once RuTube is foreground:
- All subsequent DPAD keypresses navigate RuTube's UI
- LLM vision verification sees "Android TV app UI with rows of content" → marks `tv-channel-dpad-browse` as PASS because it genuinely is a TV app with rows
- Screenshots saved with generic names (`android-tv-home-screen.png`) that match the expected outcome text
- No per-step guard verifies the foreground package is still `com.catalogizer.androidtv`

The pipeline's existing foreground guard (pipeline.go:1894-1926) only runs inside the curiosity phase, **not** the structured phase.

## Why HelixQA's other guards did not catch it

| Guard | Why it missed |
|---|---|
| Screen-state stagnation (FIX-018) | RuTube's UI changes normally on DPAD input, so the stagnation detector never trips |
| Vision outcome verification | RuTube IS a "home screen with rows" — the prompt answered TRUE truthfully to "is a home screen visible" |
| Device state preservation (Article VIII) | Only checks sensitive settings keys, not foreground package |
| Orchestrator exit-code pipe fix (FIX-011) | No error propagated because no underlying command failed |
| Curiosity foreground guard (pipeline.go:1894) | Only active in curiosity phase, not structured |

## Fix

**HelixQA/pkg/autonomous/structured_executor.go:**

1. **Preflight force-stop** of commonly-observed channel-publishing apps in `Execute()` before any test runs (`preflightStopCompetingApps`). Configurable via `PipelineConfig.CompetingAppPackages` so the library stays project-agnostic; defaults to an empirically-observed list.
2. **Per-step foreground guard** in `executeStep()` (`ensureAppForeground`). Before every step's `performAction`, runs `adb -s <device> shell dumpsys window windows` and verifies `config.AndroidPackage` appears in the focus window. On drift:
   - Emits a **CRITICAL** `AnalysisFinding` (Category: Functional, Severity: Critical) so session analysis reports it
   - Force-launches the target activity with `--es qa_username/password` extras (matching the curiosity-phase recovery pattern)
   - Sleeps 3s and continues — the current step's result is no longer trusted but subsequent tests can rerun cleanly

**HelixQA/pkg/autonomous/pipeline.go:**

- New field `CompetingAppPackages []string` on `PipelineConfig`, documented as consumer-owned to respect HelixQA Constitution §1 (no project-specific data baked into the library).

**HelixQA/cmd/helixqa/main.go:**

- Reads `HELIX_COMPETING_APP_PACKAGES` (comma-separated) into `cfg.CompetingAppPackages`.

**`.env` + `HelixQA/.env`:**

- Export `HELIX_COMPETING_APP_PACKAGES=ru.rutube.app,ru.iptvremote.android.iptv.pro,com.mitv.videoplayer,com.google.android.youtube.tv,com.google.android.youtube.tvmusic,com.xiaomi.mitv.updateservice`.

## Validation artefacts (to be added to banks/fixes-validation.yaml)

```yaml
- id: fix-qa-2026-04-21-019-foreground-drift
  name: "Structured phase guards against foreground drift"
  priority: critical
  platform: androidtv
  steps:
    - name: "Launch RuTube to simulate pre-existing drift"
      action: "adb_shell: am start -n ru.rutube.app/.MainActivity"
      expected: "RuTube in foreground"
    - name: "Run HelixQA structured phase"
      action: "run_structured_phase: bank=tv-channels"
      expected: "CRITICAL finding 'Foreground Drift During Structured Test' emitted; target app relaunched"
    - name: "Verify target app back in foreground"
      action: "adb_shell: dumpsys window windows | grep mCurrentFocus"
      expected: "Contains com.catalogizer.androidtv"
```

## Files touched

- `HelixQA/pkg/autonomous/structured_executor.go` (+135 lines — preflight + guard + helper)
- `HelixQA/pkg/autonomous/pipeline.go` (+15 lines — `CompetingAppPackages` field)
- `HelixQA/cmd/helixqa/main.go` (+11 lines — env parser)
- `.env` (+5 lines — operator-supplied list)
- `HelixQA/.env` (+8 lines — same)

## Constitution implication

This is the second foreground-drift-class bug this cycle (after FIX-015 device-state preservation). Consider amending **Article IX — HelixQA Tool Hygiene** with a `§9.4 Foreground fidelity` clause:

> Every automated step whose correctness depends on the app under test being in the foreground MUST verify `dumpsys window windows` before the step executes. A step that runs while the target package is not focused is an infrastructure failure and its result MUST NOT be used to satisfy Article V coverage requirements.

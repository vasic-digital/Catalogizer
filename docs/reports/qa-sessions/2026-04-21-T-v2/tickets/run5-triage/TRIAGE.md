# Run5 Tickets Triage — 2026-04-22

> **Source session:** `qa-results/session-20260422_000101/helixqa/session-1776805261/`
> **Pipeline report:** 70 raw findings → 36 unique tickets (dedup rate 48.6 %)
> **Orchestrator log:** `docs/reports/qa-sessions/2026-04-21-T-v2/logs/helixqa-orchestrator-run5.log`

## 1. Executive Summary

Run5's Analyze → Reproduce pipeline produced 3 vision-analyzed screenshots
yielding a total of 12 raw findings, which the ticket generator (together
with per-test failure findings and foreground-drift findings from
structured phase) consolidated to 36 unique tickets. The reproduce phase
re-attempted the first 38 per-step findings; outcomes:

| Reproduce outcome | Count | Meaning |
|---|---:|---|
| **CONFIRMED** | **5** | Bug reproduced in ≤3 retries → **actionable, top priority** |
| not reproduced | 7 | Retry couldn't trigger the failure — flaky, transient, or already-fixed; investigate if repeated |
| context deadline exceeded | 27 | Reproduce attempt timed out (90s per-step budget exhausted) → inconclusive; will retry next cycle |

## 2. Top-Priority Actionable Tickets (5 CONFIRMED)

| # | Finding ID | Platform | Context (from log) |
|---|---|---|---|
| 1 | `finding-1-androidtv` | androidtv | First reproduce, 1 attempt — highest-confidence real bug |
| 2 | `finding-4-androidtv` | androidtv | 2 attempts to reproduce |
| 3 | `finding-6-androidtv` | androidtv | 3 attempts — edge of the confirm threshold |
| 4 | `finding-7-androidtv` | androidtv | 2 attempts |
| 5 | `finding-8-androidtv` | androidtv | 1 attempt |

### ⚠ HelixQA persistence gap

The orchestrator's ticket generator emits `tickets_created: 36` as a scalar
in `pipeline-report.json`, but **does not write per-ticket markdown files to
the session directory**. The vision-analyzed finding descriptions, screenshot
references, and suggested remediations live only in transient in-memory
state during the session and are not recoverable after the pipeline exits.

**Follow-up ticket for HelixQA:** `FIX-QA-2026-04-22-001 — Ticket
generator must persist each issue to `session-<ts>/tickets/FIX-*.md`
with Category / Severity / Title / Description / Screenshot /
Reproduction / AcceptanceCriteria so downstream triage is not forced
to reconstruct from the orchestrator log.

## 3. Vision Analysis Screenshots

| Screenshot | Findings | Path |
|---|---:|---|
| `androidtv-001-android-tv-home-screen.png` | 7 | `qa-results/session-20260422_000101/helixqa/session-1776805261/screenshots/` |
| `androidtv-033-android-tv-home-screen.png` | 0 | same |
| `androidtv-curiosity-002-after.png` | 5 | same |

These are the ground-truth frames the LLM analyzed. Any real UI bugs
are visible in the home-screen and post-curiosity captures. The
curiosity-002-after finding is likely related to the cover-tile blank
issue (FIX-QA-2026-04-21-COVERS, prior-cycle ticket): post-launch home
screen with 189 items loaded but cover tiles rendering blank due to
client-side Coil SVG decoder absence.

## 4. Triage Plan for Next Cycle

### 4.1 First wave — confirmed findings

1. **Pull the actual screenshots** referenced by each
   `finding-{1,4,6,7,8}-androidtv` from
   `qa-results/session-20260422_000101/helixqa/session-1776805261/screenshots/`
   and the paired `frames/androidtv-session/` extracts.
2. **Review the 2.4 MB session video** at
   `qa-results/session-20260422_000101/helixqa/session-1776805261/videos/`
   — look for frozen frames, unresponsive DPAD, cover-load gaps.
3. **File concrete FIX-QA tickets** under
   `docs/reports/qa-sessions/2026-04-21-T-v2/tickets/` following the
   `templates/BUG_RETROSPECTIVE.md` schema.
4. **Land fixes** with the 4-artefact requirement (unit test +
   integration test + `banks/fixes-validation.yaml` entry + new
   challenge) per Article VII.

### 4.2 Second wave — inconclusive findings

- The 27 context-deadline-exceeded findings will be retested in Run6
  with a longer reproduce timeout (current budget 90s per-step is tight
  for vision+action+verification chains). Propose 180s default.
- Audit whether the 7 "not reproduced" findings indicate real flakes
  (bad vision call, transient LLM provider) vs. already-fixed issues;
  if flaky, add stagnation + drift counters to the ticket payload so
  triage can distinguish.

### 4.3 HelixQA code fix priority

`FIX-QA-2026-04-22-001` (ticket persistence) is the prerequisite for
every future triage. Without it, each cycle loses the semantic content
of its findings the moment the pipeline exits. Should land before the
next HelixQA campaign.

## 5. Deferred / Watch List

Nothing from Run5 indicates an infrastructure bug in HelixQA itself:
- Foreground-drift guard: 11 drifts, all recovered ✅
- Device state preservation: `font_scale=1.0` restored ✅
- Vision verification: tri-state (VERIFIED yes / ambiguous / error) working ✅
- Segment-video recorder: clean 1h 25m MP4 ✅
- Stagnation detector: no false stagnations reported ✅

The remaining infrastructure gap is ticket persistence (§4.3).

## 6. Links

- Full session: `qa-results/session-20260422_000101/`
- HelixQA artefacts: `qa-results/session-20260422_000101/helixqa/session-1776805261/`
- Orchestrator log: `docs/reports/qa-sessions/2026-04-21-T-v2/logs/helixqa-orchestrator-run5.log`
- Cycle final report: `docs/reports/qa-sessions/2026-04-21-T-v2/FINAL-REPORT.md`

# Run5 Tickets Triage — 2026-04-22 (CORRECTED)

> **Source session:** `qa-results/session-20260422_000101/helixqa/session-1776805261/`
> **Pipeline report:** 70 raw findings → 36 unique tickets (dedup rate 48.6 %)
> **Ticket files:** `docs/issues/HELIX-{145..180}-*.md` (36 files, all created between 01:00 and 01:30 on 2026-04-22)
> **Orchestrator log:** `docs/reports/qa-sessions/2026-04-21-T-v2/logs/helixqa-orchestrator-run5.log`

## 1. Correction to earlier draft

An earlier revision of this document incorrectly claimed HelixQA
does not persist ticket markdown files. **That was wrong.** The
FindingsBridge (see `HelixQA/pkg/autonomous/findings_bridge.go`)
writes a markdown file per finding to `$PROJECT/docs/issues/` — the
canonical cross-session issues directory. Configured by
`cmd/helixqa/main.go:603 IssuesDir: filepath.Join(*project, "docs",
"issues")`. All 36 Run5 tickets are at
`docs/issues/HELIX-{145..180}-*.md` and were written at their
pipeline-emit timestamps.

Any follow-up concerns about ticket persistence are invalid; the
filesystem has the full content.

## 2. Run5 Distribution

### By severity
| Severity | Count |
|---|---:|
| critical | 16 |
| high | 10 |
| medium | 6 |
| low | 4 |
| **Total** | **36** |

### By category
| Category | Count |
|---|---:|
| functional | 27 |
| ux | 6 |
| accessibility | 1 |
| content | 1 |
| performance | 1 |

## 3. Critical Findings — Already-Known / Our Own Fix Firing

7 of the 16 critical findings are **emitted by our own
FIX-QA-2026-04-21-019 foreground-drift guard** (Constitution Article
IX). They are not new product bugs — they are the guard correctly
identifying cases where the `tv-voice-search-*`, `tv-channel-*`, or
`tv-watch-next-*` structured bank tests navigated away from
Catalogizer (voice overlay, IPTV Pro, RuTube, etc.) and the guard
recovered by force-stopping the hijacker.

| ID | Test | Drift target |
|---|---|---|
| HELIX-146 | voice-search / speak-query step 1 | `ihq` voice-input overlay |
| HELIX-147 | voice-search-results display correctly step 1 | `ihq` |
| HELIX-148 | manual server-url entry step 3 | IPTV Pro |
| HELIX-149 | tv-watch-next resume-playback step 3 | IPTV Pro |
| HELIX-150 | tv-player video-start verification step 2 | RuTube |
| HELIX-163 | voice-search-failure graceful-handling step 1 | `ihq` |
| HELIX-164 | voice-search-failure graceful-handling step 2 | Google Katniss |
| HELIX-166 | home-button during playback saves-state step 1 | `ihq` |

**Triage action:** close these 8 as "not-a-product-bug, guard working
as designed". Follow-up work is in the bank, not the app:
- Add `allow_foreground_leave: true` to voice-search tests so the
  guard silences drift findings when the test is knowingly
  exercising a system overlay.
- Similarly flag `tv-channel-*` and `tv-watch-next-*` tests that
  legitimately visit the launcher.

## 4. Critical Findings — Actual Product Bugs (7)

These 7 are genuine product-side issues that need developer attention.

| ID | Title | Notes |
|---|---|---|
| HELIX-145 | tv-cold-start launch step 1 failed | Cold-start flow regression — blocker if reproducible |
| HELIX-152 | memory-pressure kill-and-restore step 3 failed | OS-level kill + restore doesn't restore full state |
| HELIX-154 | focus-lost-after-dialog-dismissal step 2 failed | Compose focus-recovery after AlertDialog dismiss |
| HELIX-157 | session-expires-during-active-media-playback step 2 failed | JWT expiry mid-stream; app should refresh token mid-playback |
| HELIX-168 | focus-indicator-visibility on all backgrounds step 2 failed | Focus-ring contrast against certain backgrounds insufficient |
| HELIX-169 | HelixQA stagnation: device did not react to 8 consecutive actions | Either product freeze OR HelixQA's stagnation detector firing legitimately |
| HELIX-179 | missing search functionality | Likely vision misread on a loading state; investigate |
| HELIX-175 | lack of color contrast | WCAG issue — matches Phase 7 a11y bucket |

**Triage action:** these become the working set for the next
Article VII fix cycle. Each needs:
1. Root-cause investigation (replay the screenshot + video frames)
2. Unit + integration + E2E + HelixQA bank + challenge fixture
3. fixes-validation entry
4. Rebuild + re-run against fixed build

## 5. High-Severity Findings (10)

Not all individually inspected; bucketed from the MD file titles:

| ID range | Theme |
|---|---|
| HELIX-151, 153, 156, 158, 159, 160, 162, 171, 172, 178 | Focus-chain traversal, D-pad edge cases, dialog handling, text-input boundary |

Triage action: review as a group; many may consolidate to a single
focus-management refactor.

## 6. Medium + Low (10)

UX polish items:
- HELIX-155, 161, 165, 167 (focus-trapped-on-invisible-element,
  various invisible-focusable placements)
- HELIX-170, 173, 174, 176, 177, 180 (unclear navigation, inconsistent
  button styles, small font size, unclear error messages, lack of
  call-to-action, missing content information)

These fold naturally into Phase 7 frontend polish + Phase 9 TV UX
pass.

## 7. Reproduce-phase Outcomes (log-sourced)

| Outcome | Count |
|---|---:|
| CONFIRMED (reproduced on retry) | 5 |
| Not reproduced (retry ran cleanly) | 7 |
| Context deadline exceeded | 27 |

The 5 CONFIRMED findings from the reproduce phase are
finding-{1,4,6,7,8}-androidtv in the orchestrator log — they may or
may not map 1:1 to the HELIX-NNN ticket IDs (ticket IDs are
allocated sequentially by the memory store's NextFindingID, not by
reproduce order). A proper mapping would require HelixQA to include
the HELIX-NNN IDs in reproduce-phase log lines; filed as
**FIX-QA-2026-04-22-002** (non-blocking).

## 8. Summary for the next cycle

- 8 foreground-drift tickets: **close as not-a-bug + update banks
  with `allow_foreground_leave: true`** on legitimate launcher-
  visiting tests
- 7 genuine critical bugs (HELIX-145/152/154/157/168/175/179 +
  HELIX-169 stagnation): **primary fix queue** — these need Article
  VII 4-artefact closure each
- 10 high-severity focus/dialog findings: **consolidate into a
  single focus-management refactor PR**
- 10 medium/low UX findings: **batch into Phase 7 polish pass**

Next HelixQA campaign (Run6) should:
- Raise reproduce per-step timeout from 90 s → 180 s (catches the
  27 deadline-exceeded)
- Emit HELIX-NNN IDs in reproduce-phase lines
- Mark voice-search + channel-* tests as `allow_foreground_leave`

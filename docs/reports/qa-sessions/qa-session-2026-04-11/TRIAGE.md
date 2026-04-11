# Open Issue Triage — 2026-04-12

This document consolidates every open HelixQA ticket, known issue, and
unfinished-work marker discovered during Phase B of the Constitution
Article V remediation initiative. Each row lists current status as
determined by cross-referencing recent commits (post 2026-04-04).

Status legend:
- **CLOSED** — fix committed and verified by a test or subsequent session
- **NEEDS VERIFICATION** — fix committed but not yet re-run through HelixQA
- **OPEN** — no fix found in git log
- **NEW** — surfaced after 2026-04-04, no ticket file yet

## HelixQA tickets from session 2026-04-04

| # | Severity | Platform | Title | Status | Commit(s) | Regression test |
|---|----------|----------|-------|--------|-----------|-----------------|
| 1 | MAJOR | Android TV | Search: text accumulates instead of replacing | NEEDS VERIFICATION | `91637738 fix(tv/search): open IME on DPAD_CENTER, submit on second press` | Needs HelixQA bank entry |
| 2 | MAJOR | Android TV | Black screen during app transition | OPEN | — | — |
| 3 | MINOR | Android TV | Settings screen visible behind loading spinner | OPEN | — | — |
| 4 | MINOR | Android TV | Entity title truncated on detail screen | OPEN | — | — |
| 5 | CRITICAL | Web | Playwright sees blank white page | NEEDS VERIFICATION | Memory notes "Playwright pre-login via API token injection (web QA unblocked)" | Needs HelixQA bank entry |
| 6 | MAJOR | Web | HelixQA unable to navigate past login page | NEEDS VERIFICATION | Same as #5 | Needs bank entry |
| 7 | CRITICAL | Web | Playwright cannot authenticate | NEEDS VERIFICATION | Same as #5 | Needs bank entry |
| 8 | MAJOR | Web | Dashboard "Something went wrong" after login | OPEN | — | — |
| 9 | MAJOR | All | Media playback not verified during QA | NEEDS VERIFICATION | `fb787d85 feat(tv/player): libVLC 3.6.2`, `0e059db8 fix: TV Play Now no longer hands libVLC a metadata file`, `f1a57c0b feat(tv/playback): record every VLC session via PlaybackTracker`, `e8233a77 test(challenges): CH-200 playback sessions lifecycle` | CH-200 + `tv-playback-session-lifecycle` bank entry exist |

## Issues surfaced after 2026-04-04 (no ticket files yet)

| # | Severity | Platform | Title | Status | Evidence |
|---|----------|----------|-------|--------|----------|
| N1 | MAJOR | Android TV | Auth token not persisted across relaunch | CLOSED | `7d206ae9 TokenStore persists JWT in EncryptedSharedPreferences`, `07d1c8c4 hydrate persisted token on cold start` — verified in `qa-session-2026-04-11/FINAL-REPORT.md` "Post-fix verification" section |
| N2 | MAJOR | Android TV | libVLC surface attachment for Compose-hosted `VLCVideoLayout` | OPEN | FINAL-REPORT "Known follow-up" — player enters "Media opening" but no HTTP GET for stream; owned by TV team |
| N3 | MAJOR | Backend | entity search + TV aggregation | CLOSED | `15b78883 feat(api): add /api/v1/entities/search`, `4e180a90 fix(api): build tv_season/tv_episode hierarchy` |
| N4 | INFO | All | Playback history UI (ProgressBadge + HistoryDrawer) | CLOSED | Completed this session: commits `7a9d6cac` / `0f447e22` / `8595b013` / `2a73e02e` (TV), `9f411a4b` (phone), `71a68e6d` (web), `22380e77` (desktop). No regression test in fixes-validation bank yet. |

## Unfinished-work markers across the repo

Scanned with `git grep -nE 'TODO\|FIXME\|XXX'` at this commit:
— pending manual sweep (to be added in next Phase B iteration if any markers slipped past the pre-commit hook).

## Constitution Article V coverage gaps (per-platform)

The ten mandatory test categories are: unit, integration, e2e, full
automation, stress, security, ddos, benchmarking, challenges, HelixQA.

Initial gap assessment (deeper audit happens during each platform's
Phase E task):

| Platform | Unit | Integration | E2E | Automation | Stress | Security | DDoS | Benchmark | Challenges | HelixQA |
|----------|------|-------------|-----|------------|--------|----------|------|-----------|------------|---------|
| catalog-api | partial | partial | partial | partial | needs audit | partial (`govulncheck`) | **missing** | partial (k6) | partial (50 CH + 174 UF + 15 MOD) | partial |
| catalog-web | partial | partial | partial | partial | needs audit | partial (`npm audit`) | **missing** | needs audit | via userflow | partial |
| catalogizer-desktop | partial (new today) | needs audit | needs audit | needs audit | **missing** | needs audit | **missing** | **missing** | via userflow | needs audit |
| installer-wizard | needs audit | needs audit | needs audit | needs audit | **missing** | needs audit | **missing** | **missing** | via userflow | needs audit |
| catalogizer-android | partial | needs audit | needs audit | needs audit | **missing** | needs audit | **missing** | **missing** | via userflow | partial |
| catalogizer-androidtv | partial | needs audit | needs audit | needs audit | **missing** | needs audit | **missing** | **missing** | via userflow | partial |
| catalogizer-api-client | partial | needs audit | n/a | n/a | **missing** | needs audit | **missing** | **missing** | n/a | n/a |

"Needs audit" means no systematic enumeration has been done yet.
"Missing" means the category has no identifiable test files at all.

## Next steps (Phase C onward)

1. Add HelixQA fixes-validation bank entries for every NEEDS VERIFICATION
   row so a single HelixQA pass confirms them in one sweep.
2. Fix T2 (black screen on app transition) and T8 (web dashboard
   "Something went wrong") at root cause.
3. Fix T3 and T4 (minor androidtv UX).
4. Fix N2 (libVLC surface attachment) — hand-off to TV team marker;
   we'll implement the `attachView` in the first composition.
5. Add the playback-history UI regression test entry (FIX-003).
6. Begin Phase D infrastructure scaffolding for the missing categories
   (DDoS, stress, benchmark) across platforms.
7. Execute Phase E1 (catalog-api 100% pass), then E2..E9 sequentially.

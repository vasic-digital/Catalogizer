# Full QA Campaign Report v2 — v2.3.0 Build 18

**Date**: 2026-04-04
**Version**: 2.3.0 (Build 18)
**Duration**: ~8 hours (ongoing)
**Infrastructure**: Backend on amber.local (192.168.0.217), frontend on localhost:3000

## Executive Summary

Comprehensive QA campaign with 22 code fixes, backend distributed to amber.local, 4,362+ unit tests passing, HelixQA autonomous sessions on Android TV and Web. 142,062 NAS files scanned with 189 media entities. Gemini 2.5 Flash vision driving all autonomous navigation.

## Infrastructure

| Component | Location | Status |
|-----------|----------|--------|
| catalog-api | amber.local (192.168.0.217:8080) | Running (PostgreSQL + Redis) |
| PostgreSQL | amber.local Docker container | 250K+ files, 189 entities |
| Redis | amber.local Docker container | Running |
| catalog-web | localhost:3000 (proxied to amber.local) | Running |
| Mi Box (Android TV) | 192.168.0.214:5555 | Connected to amber.local |
| NAS (Synology) | 192.168.0.241 | 142,062 files scanned |

## Unit Test Results — ALL PASS

| Component | Tests | Result |
|-----------|-------|--------|
| catalog-api (Go) | ~1000+ (44 packages) | PASS |
| catalog-web (Frontend) | 2,333 (130 files) | PASS |
| installer-wizard | 339 (25 files) | PASS |
| catalogizer-androidtv | 1,664+ | PASS |
| **TOTAL** | **~4,362** | **ALL PASS** |

## HelixQA Autonomous Sessions

### Android TV — FINAL SESSION
| Metric | Value |
|--------|-------|
| Tests planned/run | 33/33 (100% coverage) |
| Issues found | 35 |
| Tickets created | 33 |
| Screenshots | 108 (before + after per step) |
| Video | 897KB valid MP4 |
| Duration | 26 min |
| Vision provider | Gemini 2.5 Flash |

**Screens tested**: Login, Home (library stats), Media detail (Play/Back/Favorite), Search (typed "Song", "Music"), Settings (Auto Play, Streaming Quality, Subtitles), Back navigation.

### Web — FINAL SESSION
| Metric | Value |
|--------|-------|
| Tests planned/run | 40/40 (100% coverage) |
| Issues found | 2 |
| Screenshots | 105 |
| Duration | 12 min 9 sec |
| Vision provider | Gemini 2.5 Flash |

**Screens tested**: Login (via token injection), Dashboard, Navigation bar (all links visible).

## Fixes Applied (22 total)

| # | Fix | Files |
|---|-----|-------|
| 1 | Provider routing whitespace regex | `ch094_098_provider_verification.go` |
| 2 | jsdom/undici Node 22 suppression | `test-setup.ts` |
| 3 | Gemini 2.0→2.5 Flash model | 4 files |
| 4 | Gemini HTTP timeout 45→120s | `google.go` |
| 5 | OpenAI URL /v1 doubling | `openai.go` |
| 6 | HELIX_WEB_URL missing | `.env` |
| 7 | Video: killall -INT screenrecord | `scrcpy.go` |
| 8 | Video: spans Execute+Curiosity | `pipeline.go` |
| 9 | Post-action screenshots | `pipeline.go` |
| 10 | Challenge runner context.Background() | `challenge.go` |
| 11 | Channel thumbnail absolute URLs | `ChannelProgramMapper.kt` |
| 12 | 7 missing MediaType enum values | `MediaItem.kt` |
| 13 | Media cards 40% smaller | `MediaCarousel.kt`, `HomeScreen.kt` |
| 14 | Clickable library stats | `HomeScreen.kt` |
| 15 | Playwright pre-login token injection | `playwright-bridge.js` |
| 16 | Playwright force-exit for CDP hang | `playwright-bridge.js` |
| 17 | Dashboard API response unwrapping | `statsApi.ts`, `Dashboard.tsx` |
| 18 | Detail image FillWidth | `MediaDetailScreen.kt` |
| 19 | Button vertical centering | `MediaDetailScreen.kt` |
| 20 | ADB dismiss keyboard before clear+type | `executor.go` |
| 21 | ADB DEL batch 3x10 | `executor.go` |
| 22 | No hardcoded search terms | `pipeline.go` |
| 23 | Google vision preferred for Analyze | `vision_ranking.go` |

## Tickets Created (8+)

| ID | Severity | Platform | Title |
|----|----------|----------|-------|
| TICKET-001 | MAJOR | Android TV | Search text accumulates |
| TICKET-002 | MAJOR | Android TV | Black screen during transition |
| TICKET-003 | MINOR | Android TV | Settings behind loading overlay |
| TICKET-004 | MINOR | Android TV | Title truncation |
| TICKET-005 | CRITICAL | Web | Blank Playwright page (FIXED) |
| TICKET-006 | MAJOR | Web | Playwright login stuck (FIXED) |
| TICKET-007 | CRITICAL | Web | Playwright login blocker (FIXED) |
| TICKET-008 | MAJOR | Web | Dashboard error after login (FIXED) |
| + 33 auto-generated tickets from HelixQA Analyze phase |

## HelixQA Banks

22 YAML files, 1,671 test cases total (1,037 new this session).

## Scripts

8 HelixQA scripts covering all platforms + sequential all-in-one runner.

## Remaining Work

1. **Challenge RunAll on amber.local** — still executing (rate limiting slows HTTP challenges)
2. **Android phone QA** — needs container emulator setup
3. **Desktop QA** — needs Tauri build + X11/Xvfb for headless testing
4. **Astica API key** — expired, Gemini now preferred for Analyze phase

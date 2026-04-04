# Full QA Campaign v2.3.0 — Design Spec

**Date**: 2026-04-04
**Version**: 2.3.0 (Build 18 target)
**Scope**: Clean-slate rebuild, comprehensive testing, HelixQA autonomous QA, iterative fix loop

## Objective

Execute a complete quality assurance campaign across all Catalogizer platforms: rebuild all binaries and containers from clean state, run all existing tests and challenges, extend HelixQA with comprehensive test suites covering all features with varied data sets, run full autonomous QA sessions with video recording, analyze all evidence, fix all discovered issues, and iterate until zero defects remain.

## Target Platforms & Devices

| Platform | Target | Testing Method |
|----------|--------|----------------|
| catalog-api (Go/Gin) | Host + Container | Unit tests, Challenges, HelixQA bank |
| catalog-web (React/TS) | Container | Unit tests, Playwright E2E, HelixQA autonomous |
| catalogizer-androidtv | Mi Box 192.168.0.214:5555 | HelixQA autonomous + video recording |
| catalogizer-android | Container emulator | HelixQA autonomous + video recording |
| catalogizer-desktop | Host (Tauri) | HelixQA autonomous |
| installer-wizard | Host (Tauri) | Unit tests |
| catalogizer-api-client | Host | Unit tests |

**Excluded**: ATMOSphere devices (in `.devignore`). No other ADB devices.

## Phase 0: Constitution & Documentation Updates

Update CLAUDE.md and AGENTS.md with mandatory constraints:
- Iterative test-fix-rebuild loop requirement
- Live monitoring during all test execution
- Video recording mandatory for all device/emulator sessions
- Complete logging and archival in docs/reports/
- All new tests must be persisted in bank system
- Rebuild before each test iteration
- Fixes must include validation test suite entries

## Phase 1: Clean-Slate Rebuild

1. Clean old build artifacts
2. Build catalog-api binary (`go build -o catalog-api`)
3. Build catalog-web (`npm run build`)
4. Build catalogizer-api-client (`npm run build`)
5. Build Android TV APK (`./gradlew assembleDebug`)
6. Container build: `./scripts/release-build.sh --container --force --skip-tests`
7. Deploy APK to Mi Box: `adb -s 192.168.0.214:5555 install -r`
8. Version bump to build 18

## Phase 2: All Unit Tests + Challenges

### Unit Tests
- Go: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2` (44 packages)
- Frontend: `npm run test` in catalog-web (130+ files)
- API Client: `npm run test` in catalogizer-api-client
- Installer Wizard: `npm run test` in installer-wizard
- Android TV: `./gradlew test` in catalogizer-androidtv (if JDK 17 available)

### Challenges
- Start services via containers (catalog-api + catalog-web + PostgreSQL)
- Execute ALL registered challenges (500+) via REST API sequentially
- Log each challenge: ID, name, status, duration, assertions
- Capture all results to `docs/reports/qa-sessions/qa-session-2026-04-04/challenges/`

## Phase 3: HelixQA Bank Tests

Execute all 16 existing YAML banks:
```
helixqa run --banks banks/*.yaml --platform all
```
Results captured with full evidence (screenshots, logs).

## Phase 4: HelixQA Extension — Comprehensive Test Suites

Create new comprehensive bank files:

### Full QA Test Suite Banks (new)
1. `banks/full-qa-api.yaml` — All API endpoints with positive/negative/edge data
2. `banks/full-qa-web.yaml` — All web pages, components, interactions
3. `banks/full-qa-androidtv.yaml` — All TV screens, DPAD navigation, channels
4. `banks/full-qa-android.yaml` — All phone screens, touch interactions
5. `banks/full-qa-cross-platform.yaml` — Cross-platform data consistency

### Data Sets
- Known media titles from catalog (movies, TV shows, music, books, comics, games, software)
- Invalid/malformed data (SQL injection strings, XSS payloads, empty strings, Unicode edge cases)
- Boundary data (max-length strings, zero values, negative IDs)
- Cyrillic text (matching NAS content paths)

### Coverage Requirements
Every test bank MUST cover:
- All happy paths for the platform
- All screens/pages/views with UI element validation
- All CRUD operations for each entity type
- Search with real content terms + edge cases
- Authentication flows (login, logout, session expiry)
- Navigation (forward, back, deep linking)
- Media playback (video, audio, images)
- Settings and configuration
- Error states and recovery

## Phase 5: Full Autonomous QA

### Web Platform
```
helixqa autonomous --project . --platforms web --env .env --timeout 2h
```
- Playwright browser automation
- Video recording via Playwright `--video on`

### Android TV (Mi Box)
```
helixqa autonomous --project . --platforms androidtv --env .env --timeout 2h
```
- ADB at 192.168.0.214:5555
- Video recording via `adb shell screenrecord`
- ADB reverse proxy for API access

### Android Phone (Container Emulator)
```
podman-compose -f docker-compose.test.yml --profile android up -d
helixqa autonomous --project . --platforms android --env .env --timeout 2h
```
- Container-based emulator
- Video recording via ADB screenrecord

### Live Monitoring Requirements
During all autonomous sessions:
- Real-time display: platform, app, current test case, progress percentage
- Per-test-case: short description, status (running/pass/fail), duration
- Aggregate: total tests, passed, failed, skipped, warnings
- All output logged to `docs/reports/qa-sessions/qa-session-2026-04-04/logs/`

## Phase 6: Video Analysis & Ticket Creation

### Analysis Checklist (per recording)
- [ ] Every screen visited matches expected layout
- [ ] No visual glitches, clipped text, wrong colors, missing assets
- [ ] All interactive elements respond correctly
- [ ] Data displays match API/database content
- [ ] Animations smooth, no frozen frames
- [ ] Loading states appropriate duration
- [ ] Brand compliance (Vasic Digital logo, color scheme)
- [ ] No unexpected crashes or restarts

### Ticket Format
```markdown
# [SEVERITY] Short description

**Platform**: web/androidtv/android
**Screen**: Screen name
**Timestamp**: MM:SS in video / screenshot filename
**Reproduction**: Step-by-step
**Expected**: What should happen
**Actual**: What happened
**Evidence**: Screenshot/video frame attachment
```

Tickets stored in `docs/reports/qa-sessions/qa-session-2026-04-04/tickets/`

## Phase 7: Fix Loop

```
while issues_remain:
    investigate_root_cause(issue)
    implement_fix(issue)
    create_validation_test(issue)  # Added to "Fixes Validation" bank
    rebuild_affected_components()
    rerun_failed_tests()
    if all_pass:
        mark_resolved(issue)
```

### Exit Conditions
- All tests pass (unit, challenges, HelixQA bank, autonomous)
- All tickets resolved or downgraded to "known issue" with justification
- FATAL BLOCKER encountered (system crash, hardware failure)
- Nothing left to test, fix, or polish

## Phase 8: Final Report

`docs/reports/qa-sessions/qa-session-2026-04-04/FINAL-REPORT.md`:
- Executive summary
- Per-platform test results table (success/warning/error counts)
- All challenge results
- All HelixQA bank results
- Autonomous QA session summaries
- Issues found and resolved
- Issues remaining (if any)
- Video analysis findings
- Performance metrics
- Suggestions for further improvement

## Phase 9: Enterprise Enhancement Research

Research and evaluate:
- Cutting-edge QA frameworks and tools
- Visual regression testing platforms
- AI-driven test generation
- Performance profiling integration
- Accessibility testing tools

Integrate applicable solutions into Challenges + HelixQA submodules.

## Resource Constraints

- **Host**: 30-40% max CPU/RAM (GOMAXPROCS=3, container limits enforced)
- **Containers**: max 4 CPUs, 8 GB RAM total
- **Podman only** — no Docker
- **Sequential challenge execution** — never parallel
- **No CI/CD** — all local execution

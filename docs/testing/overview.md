# User Flow Testing Overview

## Purpose

The user flow testing suite provides exhaustive end-to-end validation of every Catalogizer application across all 6 platforms. Unlike the existing challenge suite (CH-001 through CH-035) which validates API endpoints and system integration, the user flow challenges (UF-*) simulate real user interactions: clicking buttons, filling forms, navigating screens, and verifying visual outcomes.

## Architecture

User flow challenges are built on the generic `pkg/userflow/` framework in the Challenges submodule (`digital.vasic.challenges`). This framework provides:

- **6 adapter interfaces** -- BrowserAdapter, MobileAdapter, DesktopAdapter, APIAdapter, BuildAdapter, ProcessAdapter
- **CLI implementations** -- PlaywrightCLIAdapter, ADBCLIAdapter, TauriCLIAdapter, HTTPAPIAdapter, GradleCLIAdapter, CargoCLIAdapter, NPMCLIAdapter, GoCLIAdapter
- **Challenge templates** -- BuildChallenge, UnitTestChallenge, LintChallenge, APIHealthChallenge, APIFlowChallenge, BrowserFlowChallenge, MobileLaunchChallenge, MobileFlowChallenge, DesktopLaunchChallenge, DesktopFlowChallenge, DesktopIPCChallenge, InstrumentedTestChallenge, EnvironmentSetupChallenge, EnvironmentTeardownChallenge
- **12 assertion evaluators** -- build_succeeds, all_tests_pass, lint_passes, app_launches, app_stable, status_code, response_contains, response_not_empty, json_field_equals, screenshot_exists, flow_completes, within_duration

All Catalogizer-specific challenges live in `catalog-api/challenges/userflow_*.go` and are registered via `register.go`.

## Challenge Distribution

| Platform | File | Count | Description |
|----------|------|-------|-------------|
| Environment | `userflow_api.go` | 2 | Setup and teardown bookends |
| API | `userflow_api.go` | 49 | REST API flows across all endpoints |
| Web | `userflow_web.go` | 59 | Browser automation via Playwright |
| Desktop | `userflow_desktop.go` | 18 | Tauri desktop app (catalogizer-desktop) |
| Wizard | `userflow_desktop.go` | 10 | Tauri installer wizard (installer-wizard) |
| Android | `userflow_mobile.go` | 22 | Android phone app via ADB |
| Android TV | `userflow_mobile.go` | 16 | Android TV app via ADB with D-pad |
| **Total** | | **174** (+ 2 env) | |

## Platform Details

### API (49 challenges)

Validates all REST API endpoints using the `HTTPAPIAdapter`. Challenges are organized into these categories:

| Category | Count | Challenge IDs |
|----------|-------|---------------|
| Health | 3 | UF-API-HEALTH, UF-API-HEALTH-VERSION, UF-API-HEALTH-METRICS |
| Auth | 5 | UF-API-AUTH-LOGIN, UF-API-AUTH-REGISTER, UF-API-AUTH-INVALID, UF-API-AUTH-TOKEN-REFRESH, UF-API-AUTH-LOGOUT |
| Media | 10 | UF-API-MEDIA-LIST, UF-API-MEDIA-GET, UF-API-MEDIA-SEARCH, UF-API-MEDIA-TYPES, UF-API-MEDIA-ENTITY, UF-API-MEDIA-HIERARCHY, UF-API-MEDIA-COVER, UF-API-MEDIA-METADATA, UF-API-MEDIA-RECENT, UF-API-MEDIA-STATS |
| Collections | 5 | UF-API-COLL-LIST, UF-API-COLL-CREATE, UF-API-COLL-ADD, UF-API-COLL-REMOVE, UF-API-COLL-DELETE |
| Storage | 5 | UF-API-STORAGE-LIST, UF-API-STORAGE-ADD, UF-API-STORAGE-SCAN, UF-API-STORAGE-STATUS, UF-API-STORAGE-FILES |
| Admin | 5 | UF-API-ADMIN-USERS, UF-API-ADMIN-CONFIG, UF-API-ADMIN-LOGS, UF-API-ADMIN-STATS, UF-API-ADMIN-SESSIONS |
| Downloads | 3 | UF-API-DL-REQUEST, UF-API-DL-STATUS, UF-API-DL-STREAM |
| Favorites | 3 | UF-API-FAV-ADD, UF-API-FAV-LIST, UF-API-FAV-REMOVE |
| WebSocket | 2 | UF-API-WS-CONNECT, UF-API-WS-EVENTS |
| Errors | 3 | UF-API-ERR-404, UF-API-ERR-401, UF-API-ERR-400 |
| Security | 3 | UF-API-SEC-CORS, UF-API-SEC-RATE, UF-API-SEC-HEADERS |

### Web (59 challenges)

Browser automation using `PlaywrightCLIAdapter` connecting to a Playwright container via CDP (Chrome DevTools Protocol) at `ws://localhost:9222`. Default browser: Chromium, headless, 1920x1080.

| Category | Count | Prefix |
|----------|-------|--------|
| Auth | 5 | UF-WEB-AUTH-* |
| Dashboard | 5 | UF-WEB-DASH-* |
| Media Browser | 8 | UF-WEB-BROWSE-* |
| Collections | 6 | UF-WEB-COLL-* |
| Player | 4 | UF-WEB-PLAYER-* |
| Admin | 5 | UF-WEB-ADMIN-* |
| Subtitles | 4 | UF-WEB-SUB-* |
| Conversion | 3 | UF-WEB-CONV-* |
| Analytics | 3 | UF-WEB-ANALYTICS-* |
| Favorites | 3 | UF-WEB-FAV-* |
| Playlists | 4 | UF-WEB-PLAYLIST-* |
| Responsive | 3 | UF-WEB-RESP-* |
| Error Handling | 3 | UF-WEB-ERR-* |
| Accessibility | 3 | UF-WEB-A11Y-* |

### Desktop (18 challenges)

Tauri WebDriver automation of the `catalogizer-desktop` application using `TauriCLIAdapter` and `CargoCLIAdapter` for build verification.

| Category | Count | Prefix |
|----------|-------|--------|
| Build | 3 | UF-DESKTOP-BUILD, UF-DESKTOP-TEST, UF-DESKTOP-LINT |
| Launch | 3 | UF-DESKTOP-LAUNCH, UF-DESKTOP-STABLE, UF-DESKTOP-SCREENSHOT |
| Auth | 3 | UF-DESKTOP-AUTH-LOGIN, UF-DESKTOP-AUTH-PERSIST, UF-DESKTOP-AUTH-LOGOUT |
| Browse | 4 | UF-DESKTOP-BROWSE-LOAD, UF-DESKTOP-BROWSE-SEARCH, UF-DESKTOP-BROWSE-DETAIL, UF-DESKTOP-BROWSE-FILTER |
| IPC | 3 | UF-DESKTOP-IPC-VERSION, UF-DESKTOP-IPC-CONFIG, UF-DESKTOP-IPC-SETTINGS |
| Settings | 2 | UF-DESKTOP-SETTINGS-LOAD, UF-DESKTOP-SETTINGS-SAVE |

### Wizard (10 challenges)

Tauri WebDriver automation of the `installer-wizard` application. Tests the complete setup wizard flow including protocol selection, server configuration, and validation.

| Category | Count | Prefix |
|----------|-------|--------|
| Build | 2 | UF-WIZARD-BUILD, UF-WIZARD-TEST |
| Flow | 5 | UF-WIZARD-LAUNCH, UF-WIZARD-WELCOME, UF-WIZARD-PROTOCOL, UF-WIZARD-SERVER, UF-WIZARD-COMPLETE |
| Validation | 3 | UF-WIZARD-VALIDATE-EMPTY, UF-WIZARD-VALIDATE-IP, UF-WIZARD-VALIDATE-PATH |

### Android (22 challenges)

ADB automation of the `catalogizer-android` application using `ADBCLIAdapter` and `GradleCLIAdapter`. Interactions use coordinate-based taps (1080x1920 resolution) and keycode events.

| Category | Count | Prefix |
|----------|-------|--------|
| Build | 3 | UF-ANDROID-BUILD, UF-ANDROID-TEST, UF-ANDROID-LINT |
| Launch | 3 | UF-ANDROID-LAUNCH, UF-ANDROID-STABLE, UF-ANDROID-SCREENSHOT |
| Auth | 3 | UF-ANDROID-AUTH-LOGIN, UF-ANDROID-AUTH-INVALID, UF-ANDROID-AUTH-LOGOUT |
| Browse | 4 | UF-ANDROID-BROWSE-LOAD, UF-ANDROID-BROWSE-SEARCH, UF-ANDROID-BROWSE-DETAIL, UF-ANDROID-BROWSE-SCROLL |
| Playback | 3 | UF-ANDROID-PLAY-START, UF-ANDROID-PLAY-CONTROLS, UF-ANDROID-PLAY-SEEK |
| Settings | 2 | UF-ANDROID-SETTINGS-LOAD, UF-ANDROID-SETTINGS-SERVER |
| Offline | 2 | UF-ANDROID-OFFLINE-BANNER, UF-ANDROID-OFFLINE-CACHE |
| Instrumented | 2 | UF-ANDROID-INSTR-UI, UF-ANDROID-INSTR-NAV |

### Android TV (16 challenges)

ADB automation of the `catalogizer-androidtv` application. Uses D-pad keycode events (DPAD_UP, DPAD_DOWN, DPAD_LEFT, DPAD_RIGHT, DPAD_CENTER) for all navigation since TV devices have no touchscreen.

| Category | Count | Prefix |
|----------|-------|--------|
| Build | 3 | UF-ANDROIDTV-BUILD, UF-ANDROIDTV-TEST, UF-ANDROIDTV-LINT |
| Launch | 2 | UF-ANDROIDTV-LAUNCH, UF-ANDROIDTV-STABLE |
| Navigation | 3 | UF-ANDROIDTV-NAV-DPAD, UF-ANDROIDTV-NAV-SELECT, UF-ANDROIDTV-NAV-BACK |
| Browse | 3 | UF-ANDROIDTV-BROWSE-LOAD, UF-ANDROIDTV-BROWSE-ROW, UF-ANDROIDTV-BROWSE-DETAIL |
| Playback | 3 | UF-ANDROIDTV-PLAY-START, UF-ANDROIDTV-PLAY-CONTROLS, UF-ANDROIDTV-PLAY-DPAD |
| Settings | 2 | UF-ANDROIDTV-SETTINGS-LOAD, UF-ANDROIDTV-SETTINGS-SERVER |

## Dependency Model

All challenges form a directed acyclic graph (DAG) rooted at `UF-ENV-SETUP`:

```
UF-ENV-SETUP
  |
  +-- UF-API-HEALTH
  |     |
  |     +-- UF-API-AUTH-LOGIN --> all other API challenges
  |     +-- UF-API-AUTH-INVALID (no login required)
  |     +-- UF-API-ERR-401 (no login required)
  |
  +-- UF-WEB-AUTH-LOGIN --> all other web challenges
  |
  +-- UF-DESKTOP-BUILD --> UF-DESKTOP-LAUNCH --> UF-DESKTOP-AUTH-LOGIN --> browse/settings
  +-- UF-WIZARD-BUILD --> UF-WIZARD-LAUNCH --> UF-WIZARD-WELCOME --> flow/validation
  |
  +-- UF-ANDROID-BUILD --> UF-ANDROID-LAUNCH --> UF-ANDROID-AUTH-LOGIN --> browse/play/settings/offline
  +-- UF-ANDROIDTV-BUILD --> UF-ANDROIDTV-LAUNCH --> nav/browse/play/settings
  |
  +-- UF-ENV-TEARDOWN (depends on all other challenges)
```

## Resource Budget

All test execution respects the host's 30-40% resource limit (4 CPU / 8 GB RAM total). Platform groups run sequentially:

| Group | Containers | CPU Budget | Memory Budget |
|-------|-----------|------------|---------------|
| API | catalog-api | 1 | 2 GB |
| Web | catalog-api, catalog-web, playwright | 2.5 | 5 GB |
| Desktop | catalog-api, tauri-desktop | 1.5 | 3 GB |
| Wizard | tauri-wizard | 0.5 | 1 GB |
| Android | catalog-api, android-emulator | 3 | 6 GB |
| TV | catalog-api, android-emulator (TV) | 3 | 6 GB |

Groups start and stop their containers independently to stay within budget.

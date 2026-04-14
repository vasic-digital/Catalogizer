# User Flow Automation Design

**Date:** 2026-02-24
**Status:** Approved
**Scope:** Extend Challenges module with generic `pkg/userflow/` framework; create exhaustive Catalogizer user flow challenges across all 6 applications

## Overview

Refactor the Challenges submodule's project-specific `pkg/yole/` into a fully generic `pkg/userflow/` package. This universal framework provides adapter interfaces for 6 platform concerns (browser, mobile, desktop, API, build, process), pre-built challenge templates, assertion evaluators, and container infrastructure integration via the Containers module. Zero project-specific references in the Challenges module.

Catalogizer-specific challenges (~200-265) live in `catalog-api/challenges/` and import `pkg/userflow/`.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Scope | All 6 apps simultaneously | Exhaustive coverage required |
| Architecture | Adapter-per-platform | Clean separation, independently testable, extensible |
| Generic package name | `pkg/userflow/` | Describes what it does (user flow automation) |
| Project-specific location | `catalog-api/challenges/` | Keeps Challenges module universal |
| Yole references | Remove completely | Module must be fully universal |
| Execution environment | Containerized (Podman) always | Reproducible, self-contained |
| Container orchestration | Containers submodule (`digital.vasic.containers`) | Already has Podman runtime, compose, health checks, service registry |
| Test frameworks | Playwright (web), Robolectric + UIAutomator (Android), Tauri WebDriver (desktop), HTTP client (API) | State-of-the-art, container-friendly, already aligned with Challenges module interfaces |
| Coverage level | Exhaustive | Every button, form field, error state, permission |

## Architecture

### Package Structure: `Challenges/pkg/userflow/`

```
pkg/userflow/
├── types.go                # TestResult, BuildResult, LintResult, TestSuite, TestCase, TestFailure
├── plugin.go               # UserFlowPlugin (registers evaluators with assertion engine)
├── evaluators.go           # 12 generic evaluators
├── options.go              # Functional options (WithProjectRoot, WithContainerRuntime, etc.)
├── result_parser.go        # JUnit XML / Go JSON / Cargo JSON result normalization
├── container_infra.go      # TestEnvironment (bridges to Containers module)
│
├── adapter_browser.go      # BrowserAdapter interface
├── adapter_mobile.go       # MobileAdapter interface
├── adapter_desktop.go      # DesktopAdapter interface
├── adapter_api.go          # APIAdapter interface
├── adapter_build.go        # BuildAdapter interface
├── adapter_process.go      # ProcessAdapter interface
│
├── playwright_adapter.go   # BrowserAdapter impl (Playwright via CDP in container)
├── adb_adapter.go          # MobileAdapter impl (ADB CLI, configurable package/activity)
├── tauri_adapter.go        # DesktopAdapter impl (Tauri WebDriver protocol)
├── http_api_adapter.go     # APIAdapter impl (wraps existing httpclient)
├── gradle_adapter.go       # BuildAdapter impl (Gradle)
├── npm_adapter.go          # BuildAdapter impl (npm/Vitest/Playwright)
├── go_adapter.go           # BuildAdapter impl (go build/test)
├── cargo_adapter.go        # BuildAdapter impl (cargo build/test)
├── process_adapter.go      # ProcessAdapter impl (generic binary lifecycle)
│
├── challenge_build.go      # BuildChallenge template
├── challenge_unit_test.go  # UnitTestChallenge template
├── challenge_lint.go       # LintChallenge template
├── challenge_api_health.go # APIHealthChallenge template
├── challenge_api_flow.go   # APIFlowChallenge template (multi-step API flows)
├── challenge_browser.go    # BrowserFlowChallenge template
├── challenge_mobile.go     # MobileLaunchChallenge, MobileFlowChallenge, InstrumentedTestChallenge
├── challenge_desktop.go    # DesktopLaunchChallenge, DesktopFlowChallenge, DesktopIPCChallenge
├── challenge_env.go        # EnvironmentSetupChallenge, EnvironmentTeardownChallenge
│
├── flow_api.go             # APIFlow, APIStep, StepAssertion types
├── flow_browser.go         # BrowserFlow, BrowserStep types
├── flow_mobile.go          # MobileFlow, MobileStep types
└── flow_ipc.go             # IPCCommand type
```

### Adapter Interfaces

**6 adapter interfaces** covering all platform concerns:

1. **BrowserAdapter** — Web testing via Playwright (Navigate, Click, Fill, SelectOption, IsVisible, WaitForSelector, GetText, GetAttribute, Screenshot, EvaluateJS, NetworkIntercept, Close)
2. **MobileAdapter** — Android device/emulator testing via ADB (InstallApp, LaunchApp, StopApp, TakeScreenshot, Tap, SendKeys, PressKey, WaitForApp, RunInstrumentedTests, Close)
3. **DesktopAdapter** — Tauri desktop testing via WebDriver (LaunchApp, Navigate, Click, Fill, IsVisible, Screenshot, InvokeCommand, WaitForWindow, Close)
4. **APIAdapter** — REST/WebSocket testing via HTTP client (Login, LoginWithRetry, Get, GetRaw, GetArray, PostJSON, PutJSON, Delete, WebSocketConnect, SetToken)
5. **BuildAdapter** — Build system abstraction (Build, RunTests, Lint) with implementations for Gradle, npm, Go, Cargo
6. **ProcessAdapter** — Application process lifecycle (Launch, IsRunning, WaitForReady, Stop)

All adapters have `Available(ctx) bool` for capability detection.

### CLI Implementations

All adapters work via subprocess execution — no Go bindings to Playwright/ADB/etc.:

- **PlaywrightCLIAdapter** — Connects to Playwright container via CDP, executes Node.js scripts via `runtime.Exec()`
- **ADBCLIAdapter** — Fully configurable package/activity names (no hardcoded app references), `adb` CLI calls
- **TauriCLIAdapter** — Launches Tauri binary with `TAURI_AUTOMATION=true`, communicates via WebDriver HTTP protocol
- **HTTPAPIAdapter** — Wraps existing `Challenges/pkg/httpclient` (zero duplication), adds WebSocket via gorilla/websocket
- **GradleCLIAdapter** — `./gradlew` or `podman-compose run`, JUnit XML parsing
- **NPMCLIAdapter** — `npm run test --reporter=junit`, JUnit XML parsing
- **GoCLIAdapter** — `go test -json`, JSON output parsing
- **CargoCLIAdapter** — `cargo test --format=json`, JSON output parsing
- **ProcessCLIAdapter** — Generic binary launch, SIGTERM/SIGKILL shutdown

### Common Result Types

All build/test adapters normalize to:

```go
type TestResult struct {
    Suites       []TestSuite
    TotalTests   int
    TotalFailed  int
    TotalErrors  int
    TotalSkipped int
    Duration     time.Duration
    Output       string
}

type BuildResult struct {
    Target    string
    Success   bool
    Duration  time.Duration
    Output    string
    Artifacts []string
}
```

### Assertion Evaluators (12)

Registered via `UserFlowPlugin`:

| Evaluator | Input | Pass Condition |
|-----------|-------|----------------|
| `build_succeeds` | bool | value == true |
| `all_tests_pass` | int (failures) | value == 0 |
| `lint_passes` | bool | value == true |
| `app_launches` | bool | value == true |
| `app_stable` | bool | value == true |
| `status_code` | int | value == expected |
| `response_contains` | string | contains expected |
| `response_not_empty` | string | len > 0 |
| `json_field_equals` | any | field == expected |
| `screenshot_exists` | []byte | len > 0 |
| `flow_completes` | bool | all steps passed |
| `within_duration` | int (ms) | value <= threshold |

### Container Infrastructure (TestEnvironment)

Bridges to Containers module (`digital.vasic.containers`):

```go
type TestEnvironment struct {
    runtime    containers.ContainerRuntime   // Auto-detected (Podman-first)
    compose    containers.ComposeOrchestrator
    health     containers.HealthChecker
    registry   containers.ServiceRegistry
    eventBus   containers.EventBus
    groups     []PlatformGroup
}
```

**Lifecycle:** Setup → Health Check → Register Services → Run Challenges → Teardown

**Platform Groups** (sequential, respecting 4 CPU / 8 GB budget):

| Group | Containers | Budget |
|-------|-----------|--------|
| api | catalog-api | 1 CPU, 2g |
| web | catalog-api, catalog-web, playwright | 2.5 CPU, 5g |
| desktop | catalog-api, tauri-desktop | 1.5 CPU, 3g |
| wizard | tauri-wizard | 0.5 CPU, 1g |
| android | catalog-api, android-emulator | 3 CPU, 6g |
| tv | catalog-api, android-emulator (TV profile) | 3 CPU, 6g |

Groups run sequentially; containers started/stopped per group.

## Catalogizer-Specific Challenges

### Location

All in `catalog-api/challenges/userflow_*.go`, registered via updated `register.go`.

### Challenge Count by Platform

| Platform | Categories | Est. Challenges |
|----------|-----------|-----------------|
| Environment | Setup/Teardown | 2 |
| API | Auth, Media, Entities, Collections, Scanning, Admin, Downloads, Subtitles, Conversion, Stats, Errors, Logs, Recommendations, SMB, WebSocket, Stress, Security | 80-100 |
| Web | Auth, Dashboard, Browse, Search, Collections, Player, Admin, Subtitles, Conversion, Analytics, Favorites, Responsive, Errors, Accessibility | 60-80 |
| Desktop | Setup, Auth, Browse, Settings, IPC | 15-20 |
| Wizard | Flow, Protocols, Validation | 10-15 |
| Android | Build, Launch, Auth, Browse, Playback, Settings, Offline | 20-30 |
| Android TV | Build, Launch, Nav, Browse, Playback, Settings | 15-20 |
| **Total** | | **~200-265** |

### Dependency Graph

```
ENV-SETUP (root)
├── API tier (depend on env-setup)
│   ├── API-AUTH → all other API challenges
│   └── API-STRESS, API-SECURITY (last in API tier)
├── WEB tier (depend on env-setup + API-AUTH)
│   ├── WEB-AUTH → all other web challenges
│   └── WEB-ACCESSIBILITY, WEB-RESPONSIVE (last)
├── DESKTOP tier (depend on env-setup + API-AUTH)
│   ├── DESKTOP-SETUP → DESKTOP-AUTH → others
├── WIZARD tier (depend on env-setup)
│   ├── WIZARD-FLOW → WIZARD-PROTOCOLS → WIZARD-VALIDATION
├── ANDROID tier (depend on env-setup)
│   ├── ANDROID-BUILD → ANDROID-LAUNCH → ANDROID-AUTH → others
├── ANDROID-TV tier (depend on env-setup)
│   ├── ANDROIDTV-BUILD → ANDROIDTV-LAUNCH → others
└── ENV-TEARDOWN (depends on everything)
```

## Documentation

### Challenges Submodule (`Challenges/docs/userflow/`)

13 documentation files covering: README, all 6 adapters, challenge templates, evaluators, container integration, writing challenges guide, writing adapters guide, architecture overview.

### Catalogizer Repo (`docs/testing/`)

6 documentation files covering: overview, running tests, container setup, challenge map, adding challenges, troubleshooting.

## Execution

```bash
# All platforms (sequential groups)
curl -X POST localhost:8080/api/v1/challenges/run/category/userflow

# Single platform
curl -X POST localhost:8080/api/v1/challenges/run/category/userflow-web

# Standalone CLI
go run ./cmd/userflow-runner/ --platform=all --report=html

# Containerized CI
podman run --network host catalogizer-test-runner:latest --platform=web
```

## Changes Required

### Challenges Submodule

1. Rename `pkg/yole/` → `pkg/userflow/` (remove all Yole references)
2. Rename `cmd/yole-challenges/` → `cmd/userflow-runner/`
3. Generalize all adapter implementations (configurable package names, etc.)
4. Add new adapters: PlaywrightCLIAdapter, TauriCLIAdapter, NPMCLIAdapter, GoCLIAdapter, CargoCLIAdapter
5. Add new challenge templates: APIFlowChallenge, BrowserFlowChallenge, DesktopFlowChallenge, etc.
6. Add TestEnvironment with Containers module integration
7. Add 7 new evaluators
8. Write 13 documentation files
9. Update go.mod with Containers module dependency

### Catalogizer Repo (catalog-api)

1. Update go.mod replace directive for Challenges submodule
2. Add ~200-265 userflow challenge files in `catalog-api/challenges/`
3. Update `register.go` to register new challenges
4. Add `docker-compose.test.yml` for test container stack
5. Write 6 documentation files
6. Add `cmd/userflow-runner/` entry point for standalone execution

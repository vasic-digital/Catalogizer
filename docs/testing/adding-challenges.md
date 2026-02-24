# Adding New User Flow Challenges

## Overview

All Catalogizer-specific user flow challenges live in `catalog-api/challenges/userflow_*.go`. These files import the generic `pkg/userflow/` framework from the Challenges submodule and create platform-specific challenge instances.

## File Organization

| File | Purpose |
|------|---------|
| `userflow_api.go` | API challenges (HTTP flows, env setup/teardown) |
| `userflow_web.go` | Web browser challenges (Playwright-based) |
| `userflow_desktop.go` | Desktop and wizard challenges (Tauri-based) |
| `userflow_mobile.go` | Android and Android TV challenges (ADB-based) |

Choose the file that matches your target platform. If a new platform is needed, create a new `userflow_<platform>.go` file.

## Step 1: Choose the Right Challenge Template

The `pkg/userflow/` package provides these challenge templates:

| Template | Use Case | Adapter |
|----------|----------|---------|
| `NewAPIHealthChallenge` | Simple HTTP health check (single endpoint, single status code) | APIAdapter |
| `NewAPIFlowChallenge` | Multi-step API flow (login, sequence of requests, assertions) | APIAdapter |
| `NewBrowserFlowChallenge` | Browser automation flow (navigate, click, fill, assert) | BrowserAdapter |
| `NewBuildChallenge` | Build verification (compile, check output) | BuildAdapter |
| `NewUnitTestChallenge` | Run unit tests and verify results | BuildAdapter |
| `NewLintChallenge` | Run linter and verify clean output | BuildAdapter |
| `NewMobileLaunchChallenge` | Install and launch a mobile app | MobileAdapter |
| `NewMobileFlowChallenge` | Multi-step mobile interaction (tap, swipe, keys) | MobileAdapter |
| `NewInstrumentedTestChallenge` | Run instrumented tests on device | MobileAdapter |
| `NewDesktopLaunchChallenge` | Launch a desktop app and verify window | DesktopAdapter |
| `NewDesktopFlowChallenge` | Multi-step desktop browser flow | DesktopAdapter |
| `NewDesktopIPCChallenge` | Invoke Tauri IPC commands | DesktopAdapter |
| `NewEnvironmentSetupChallenge` | Pre-test environment verification | (custom func) |
| `NewEnvironmentTeardownChallenge` | Post-test cleanup | (custom func) |

## Step 2: Write the Challenge

### Example: Adding a new API challenge

Add the challenge to the `registerUserFlowAPIChallenges()` function in `userflow_api.go`:

```go
challenges = append(challenges,
    userflow.NewAPIFlowChallenge(
        "UF-API-MYFEATURE-CHECK",              // Unique ID (UF-<PLATFORM>-<CATEGORY>-<NAME>)
        "API My Feature Check",                 // Human-readable name
        "Verify the my-feature endpoint works", // Description
        []challenge.ID{"UF-API-AUTH-LOGIN"},    // Dependencies (must pass first)
        adapter,                                // HTTPAPIAdapter instance
        userflow.APIFlow{
            Credentials: creds,                 // Will auto-login before steps
            Steps: []userflow.APIStep{
                {
                    Name:           "check-my-feature",
                    Method:         "GET",
                    Path:           "/api/v1/my-feature",
                    ExpectedStatus: 200,
                    Assertions: []userflow.StepAssertion{
                        {
                            Type:    "not_empty",
                            Target:  "my_feature_body",
                            Message: "my-feature returns data",
                        },
                    },
                },
            },
        },
    ),
)
```

### Example: Adding a new web browser challenge

Add to one of the `register*Challenges()` helper functions in `userflow_web.go`:

```go
challenges = append(challenges,
    userflow.NewBrowserFlowChallenge(
        "UF-WEB-MYPAGE-LOAD",
        "Web My Page Load",
        "Navigate to my-page and verify content loads",
        authDep,                                // Depends on auth login
        adapter,                                // PlaywrightCLIAdapter
        cfg,                                    // BrowserConfig (chromium, headless, 1920x1080)
        userflow.BrowserFlow{
            Name:        "my-page-load",
            Description: "Load the my-page route",
            StartURL:    "http://localhost:3000/my-page",
            Steps: []userflow.BrowserStep{
                {
                    Name:     "wait-for-content",
                    Action:   "wait",
                    Selector: "[data-testid='my-page-content']",
                    Timeout:  10 * time.Second,
                    Assertions: []userflow.StepAssertion{
                        {
                            Type:    "flow_completes",
                            Target:  "my_page_loaded",
                            Message: "my-page content loaded",
                        },
                    },
                },
            },
        },
    ),
)
```

### Example: Adding a new mobile challenge

Add to `registerUserFlowMobileChallenges()` in `userflow_mobile.go`:

```go
challenges = append(challenges,
    userflow.NewMobileFlowChallenge(
        "UF-ANDROID-MYSCREEN-LOAD",
        "Android My Screen Load",
        "Navigate to my screen and verify it loads",
        androidAuthDeps,
        androidADB,
        userflow.MobileFlow{
            Name:        "android-myscreen-load",
            Description: "Open my screen on Android",
            Config:      androidMobileConfig(),
            AppPath:     androidAPKPath(),
            Steps: []userflow.MobileStep{
                {
                    Name:   "wait-for-app",
                    Action: "wait",
                },
                {
                    Name:   "tap-my-tab",
                    Action: "tap",
                    X:      540,          // Screen center for 1080px width
                    Y:      1800,         // Bottom nav bar area
                },
                {
                    Name:   "wait-screen-loaded",
                    Action: "wait",
                    Assertions: []userflow.StepAssertion{
                        {
                            Type:    "app_stable",
                            Target:  "my_screen_loaded",
                            Message: "my screen loaded successfully",
                        },
                    },
                },
            },
        },
    ),
)
```

## Step 3: Register the Challenge

If you added the challenge to an existing `register*()` function, it is already registered -- the function is called from `RegisterAll()` in `register.go`.

If you created a new registration function (e.g., for a new platform), you must add the call to `register.go`:

```go
// In register.go, inside RegisterAll():
RegisterUserFlowMyPlatformChallenges(svc)  // New platform
```

And create the corresponding public registration wrapper:

```go
func RegisterUserFlowMyPlatformChallenges(
    svc interface {
        Register(challenge.Challenge) error
    },
) {
    for _, ch := range registerUserFlowMyPlatformChallenges() {
        _ = svc.Register(ch)
    }
}
```

## Challenge ID Naming Convention

All user flow challenge IDs follow this pattern:

```
UF-<PLATFORM>-<CATEGORY>-<NAME>
```

| Component | Values |
|-----------|--------|
| `UF` | User Flow prefix (always) |
| `PLATFORM` | ENV, API, WEB, DESKTOP, WIZARD, ANDROID, ANDROIDTV |
| `CATEGORY` | AUTH, MEDIA, COLL, STORAGE, ADMIN, DL, FAV, WS, ERR, SEC, HEALTH, DASH, BROWSE, PLAYER, SUB, CONV, ANALYTICS, PLAYLIST, RESP, A11Y, BUILD, LAUNCH, IPC, SETTINGS, NAV, PLAY, OFFLINE, INSTR, VALIDATE |
| `NAME` | Short descriptive name (LOGIN, LIST, CREATE, LOAD, etc.) |

## Assertion Types

Available assertion evaluators:

| Type | Input | Pass Condition |
|------|-------|----------------|
| `build_succeeds` | bool | value == true |
| `all_tests_pass` | int (failures) | value == 0 |
| `lint_passes` | bool | value == true |
| `app_launches` | bool | value == true |
| `app_stable` | bool | value == true |
| `status_code` | int | value == expected |
| `response_contains` | string | contains expected |
| `not_empty` | string | len > 0 |
| `json_field_equals` | any | field == expected |
| `screenshot_exists` | []byte | len > 0 |
| `flow_completes` | bool | all steps passed |
| `within_duration` | int (ms) | value <= threshold |

## Browser Step Actions

For web and desktop browser flow challenges:

| Action | Description | Required Fields |
|--------|-------------|----------------|
| `navigate` | Navigate to a URL | Value (URL) |
| `click` | Click an element | Selector |
| `fill` | Fill a form field | Selector, Value |
| `wait` | Wait for element to be visible | Selector, Timeout |
| `assert_visible` | Assert element is visible | Selector |
| `select_option` | Select dropdown option | Selector, Value |
| `screenshot` | Take a screenshot | (none) |

## Mobile Step Actions

For Android and Android TV challenges:

| Action | Description | Required Fields |
|--------|-------------|----------------|
| `wait` | Wait for app to stabilize | (none) |
| `tap` | Tap at coordinates | X, Y |
| `send_keys` | Type text into focused field | Value |
| `press_key` | Press a keycode | Value (e.g., KEYCODE_ENTER) |
| `screenshot` | Take a device screenshot | (none) |
| `assert_running` | Assert app is still running | (none) |

## Best Practices

1. **Always specify dependencies.** Every challenge should depend on at least one prerequisite (typically auth login or build). Only root challenges (ENV-SETUP, BUILD) can have nil dependencies.

2. **Use descriptive assertion messages.** The message field is shown in test reports and helps diagnose failures.

3. **Set appropriate timeouts.** Default is 5 minutes per challenge via `challenge.NewConfig()`. For long-running builds, zero the timeout to use the runner's timeout.

4. **Use `ExpectedStatus: 0`** when the response code may vary (e.g., 200 or 201 for creation, 200 or 409 for registration).

5. **Use `ExtractTo`** in API steps to capture response fields for use in subsequent steps. Template variables use `{{variable_name}}` syntax in Path fields.

6. **Keep challenges focused.** Each challenge should test one logical user action or assertion. Prefer many small challenges over few large ones.

7. **Test error states.** Include challenges for invalid input, unauthorized access, and not-found cases alongside happy-path flows.

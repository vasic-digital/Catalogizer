# Module 14: Challenge System Deep Dive - Script

**Duration**: 45 minutes
**Module**: 14 - Challenge System Deep Dive

---

## Scene 1: Challenge Framework Architecture (0:00 - 15:00)

**[Visual: Challenge framework package diagram showing pkg/challenge, pkg/registry, pkg/runner, pkg/assertion, pkg/report]**

**Narrator**: Welcome to Module 14. The challenge system is the backbone of Catalogizer's quality assurance. It is built on the `digital.vasic.challenges` module -- a generic, reusable Go framework for defining, registering, executing, and reporting on structured test scenarios. Let us explore its architecture.

**[Code: Show the Challenge interface from pkg/challenge/challenge.go]**

```go
type Challenge interface {
    ID() ID
    Name() string
    Description() string
    Category() string
    Dependencies() []ID
    Configure(config *Config) error
    Validate(ctx context.Context) error
    Execute(ctx context.Context) (*Result, error)
    Cleanup(ctx context.Context) error
}
```

**[Visual: Show the lifecycle diagram: Configure -> Validate -> Execute -> Cleanup]**

**Narrator**: Every challenge implements this interface. The lifecycle is strict: Configure applies runtime settings, Validate checks preconditions, Execute runs the actual test, and Cleanup releases resources. Dependencies between challenges are expressed via ID references and resolved by the runner using topological sorting.

**[Code: Show the BaseChallenge struct]**

**Narrator**: The `BaseChallenge` struct provides a template method implementation. Concrete challenges embed it and override `Execute()`:

```go
type HealthChallenge struct {
    challenge.BaseChallenge
}

func NewHealthChallenge() *HealthChallenge {
    return &HealthChallenge{
        BaseChallenge: *challenge.NewBaseChallenge(
            "CH-001", "Health Check",
            "Verify API health endpoint responds",
            "integration",
        ),
    }
}

func (c *HealthChallenge) Execute(ctx context.Context) (
    *challenge.Result, error) {
    // ... test logic ...
    return &challenge.Result{
        Status: challenge.StatusPassed,
        Assertions: assertions,
        Metrics: metrics,
    }, nil
}
```

**[Visual: Show the Result struct with Status, Assertions, Metrics, Duration]**

**Narrator**: The `Result` struct carries the outcome: a status (passed, failed, error, skipped, stuck, timed_out), a list of assertion results, custom metrics, and execution duration. This structured output feeds into reporting and monitoring.

**[Visual: Show the assertion engine with 16 built-in evaluators]**

**Narrator**: The assertion engine in `pkg/assertion` provides 16 built-in evaluators: `not_empty`, `not_mock`, `contains`, `contains_any`, `min_length`, `quality_score`, `reasoning_present`, `code_valid`, `min_count`, `exact_count`, `max_latency`, `all_valid`, `no_duplicates`, `all_pass`, `no_mock_responses`, and `min_score`. Custom evaluators can be registered at runtime.

---

## Scene 2: Writing Custom Challenges (15:00 - 25:00)

**[Visual: Show catalog-api/challenges/ directory listing]**

**Narrator**: Catalogizer defines its challenges in `catalog-api/challenges/`. The `register.go` file wires everything together via `RegisterAll()`, which loads challenge definitions from the bank configuration and registers all challenge suites.

**[Code: Show RegisterAll from register.go]**

```go
func RegisterAll(svc *services.ChallengeService) error {
    // Load challenge bank definitions from config
    // Register 50 original challenges (CH-001 to CH-050)
    // Register module verification (MOD-001 to MOD-015)
    RegisterUserFlowAPIChallenges(svc)     // 49 API challenges
    RegisterUserFlowWebChallenges(svc)     // 59 web browser challenges
    RegisterUserFlowDesktopChallenges(svc) // 28 desktop + wizard challenges
    RegisterUserFlowMobileChallenges(svc)  // 38 Android + TV challenges
}
```

**[Visual: Show challenge count breakdown]**

**Narrator**: In total, Catalogizer has 239 registered challenges:

- 50 original challenges (CH-001 to CH-050): core integration tests
- 15 module verification challenges (MOD-001 to MOD-015): validate decoupled Go modules
- 174 user flow challenges across 4 platform groups

**[Code: Show a custom challenge Execute method]**

**Narrator**: To write a custom challenge, embed `BaseChallenge`, implement `Execute()`, and use the assertion engine:

```go
func (c *BrowsingChallenge) Execute(ctx context.Context) (
    *challenge.Result, error) {
    client := httpclient.NewClient(c.BaseURL)
    resp, _ := client.LoginWithRetry(ctx, user, pass, 5)
    resp, _ = client.Get(ctx, "/api/v1/browse/roots")

    assertions := []assertion.Result{
        engine.Evaluate(assertion.Def{
            Type: "not_empty", Target: "roots",
        }, resp.Body),
    }
    return &challenge.Result{
        Status: challenge.StatusPassed, Assertions: assertions,
    }, nil
}
```

**[Visual: Show the challenge bank JSON configuration]**

**Narrator**: Challenge definitions can also be loaded from JSON files in `challenges/config/`. The bank system supports declarative metadata, assertions, and endpoint configurations.

**[Demo: Create a new challenge, register it, and run it via the API]**

---

## Scene 3: Running Challenges (25:00 - 35:00)

**[Visual: Show the challenge API routes]**

**Narrator**: Challenges are exposed via the REST API under `/api/v1/challenges`:

- `GET /challenges` -- List all registered challenges
- `GET /challenges/:id` -- Get challenge details
- `POST /challenges/:id/run` -- Run a single challenge
- `POST /challenges/run` -- Run all challenges (blocking)
- `POST /challenges/run/category/:category` -- Run by category
- `GET /challenges/results` -- Get historical results

**[Visual: Warning icon with RunAll constraints]**

**Narrator**: A critical constraint: `RunAll` is synchronous and blocking. No other challenge can execute until it finishes. For a full Catalogizer suite, this can take 25 minutes or more if it includes NAS scanning. The `config.json` `write_timeout` must be set to 900 seconds to prevent premature HTTP timeout.

**[Code: Show progress-based liveness detection]**

```go
// Runner configuration with liveness detection:
runner.NewRunner(
    runner.WithTimeout(72*time.Hour),
    runner.WithStaleThreshold(5*time.Minute),
)

// In Execute(), report progress to avoid "stuck" status:
c.ReportProgress("scanning", map[string]any{
    "files_processed": i, "total_files": len(files),
})
```

**[Visual: Show StatusStuck vs StatusTimedOut distinction]**

**Narrator**: The framework distinguishes between stuck and timed out. If a challenge reports no progress for 5 minutes, it is declared stuck and cancelled. A hard timeout is a generous upper bound for legitimately long operations. The `ProgressReporter` is automatically attached to any challenge embedding `BaseChallenge`.

**[Demo: Run a single challenge via curl, then run a category via the API]**

---

## Scene 4: User Flow Automation and Module Verification (35:00 - 45:00)

**[Visual: Show the pkg/userflow/ package structure with adapters and templates]**

**Narrator**: The user flow automation framework in `Challenges/pkg/userflow/` is a generic, multi-platform test execution engine with 8 adapter interfaces and 21 implementations across browsers, mobile, desktop, APIs, gRPC, and WebSocket.

**[Visual: Show the Catalogizer user flow challenge files]**

**Narrator**: Key adapters include `BrowserAdapter` (Playwright, Selenium, Cypress, Puppeteer), `MobileAdapter` (ADB, Appium, Maestro, Espresso), `DesktopAdapter` (Tauri WebDriver), `APIAdapter` (HTTP), and `BuildAdapter` (Gradle, Cargo, npm). Catalogizer wires these into 174 challenges:

| File | Platform | Count |
|------|----------|-------|
| `userflow_api.go` | Go API (HTTP) | 49 |
| `userflow_web.go` | React web (Playwright) | 59 |
| `userflow_desktop.go` | Tauri desktop + wizard | 28 |
| `userflow_mobile.go` | Android + Android TV | 38 |

**[Code: Show a user flow challenge template]**

```go
// API flow challenge using the HTTPAPIAdapter
apiChallenge := userflow.NewAPIFlowChallenge(
    "UF-API-001", "Login Flow",
    "Verify complete login flow",
    adapter,
    func(ctx context.Context, api userflow.APIAdapter) error {
        resp, err := api.Post(ctx, "/api/v1/auth/login", loginBody)
        if err != nil { return err }
        if resp.StatusCode != 200 { return fmt.Errorf("expected 200") }
        return nil
    },
)
```

**[Visual: Show MOD-001 to MOD-015 in register.go]**

**Narrator**: The framework includes 12 evaluators (`http_status_ok`, `browser_element_visible`, `build_success`, `test_pass_rate`, etc.). Module verification challenges (MOD-001 to MOD-015) validate the 15 decoupled Go modules -- compilation, tests, API stability, and integration through `replace` directives.

**Narrator**: The test stack (`docker-compose.test.yml`) runs catalog-api, catalog-web, and Playwright with `network_mode: host`. Reports are generated in Markdown, JSON, and HTML formats.

**[Demo: Run the full challenge suite, review the generated report]**

---

## Key Code Examples

### List All Challenges
```bash
curl http://localhost:8080/api/v1/challenges \
  -H "Authorization: Bearer $TOKEN"
```

### Run a Single Challenge
```bash
curl -X POST http://localhost:8080/api/v1/challenges/CH-001/run \
  -H "Authorization: Bearer $TOKEN"
```

### Run by Category and CLI
```bash
curl -X POST http://localhost:8080/api/v1/challenges/run/category/integration \
  -H "Authorization: Bearer $TOKEN"

# CLI runner
./userflow-runner --platform api --report markdown --timeout 1h --output reports/
```

---

## Quiz Questions

1. What is the difference between StatusStuck and StatusTimedOut?
   **Answer**: `StatusStuck` means no progress was reported within the stale threshold (5 minutes), indicating a deadlock. `StatusTimedOut` means the hard timeout was exceeded. The distinction helps diagnose broken vs slow challenges.

2. Why is RunAll blocking, and what configuration is needed to prevent HTTP timeouts?
   **Answer**: RunAll is blocking because challenges declare dependencies resolved via topological sort. The `config.json` `write_timeout` must be 900 seconds to prevent HTTP timeout during long executions.

3. How do MOD-* challenges differ from UF-* challenges?
   **Answer**: MOD challenges validate individual Go modules (compilation, tests, API stability). UF challenges validate end-to-end user workflows across 4 platforms using real adapters (HTTP, Playwright, Tauri, ADB).

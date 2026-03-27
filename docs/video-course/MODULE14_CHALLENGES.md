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

---

## Addendum: New Test Banks -- Entity, Admin, Security, and Performance

**[Visual: Test bank directory listing showing YAML files organized by domain]**

**Narrator**: The challenge system's power comes from its test bank -- structured definitions that specify inputs, expected outputs, and validation criteria. Four new test banks have been added, significantly expanding coverage into entity detection, administration, security, and performance verification.

### Entity Test Bank

**[Visual: Open `challenges/config/testbank_entity.yaml`]**

**Narrator**: The entity test bank contains 127 test cases covering the entire media detection and aggregation pipeline. Each case specifies a filename, a directory path, and the expected media entity output including type, title, year, and hierarchy position.

```yaml
# challenges/config/testbank_entity.yaml (excerpt)
- id: ENT-001
  name: "Simple movie detection"
  category: entity_detection
  input:
    filename: "The Matrix (1999).mkv"
    path: "/movies/"
  expected:
    media_type: movie
    title: "The Matrix"
    year: 1999

- id: ENT-058
  name: "Music album with multi-disc structure"
  category: entity_hierarchy
  input:
    filename: "01 - Track One.flac"
    path: "/music/Pink Floyd/The Wall/Disc 1/"
  expected:
    media_type: song
    title: "Track One"
    track_number: 1
    hierarchy:
      - type: music_artist
        title: "Pink Floyd"
      - type: music_album
        title: "The Wall"

- id: ENT-112
  name: "Anime with Japanese naming convention"
  category: entity_detection
  input:
    filename: "[SubGroup] Steins;Gate - 01 [1080p].mkv"
    path: "/anime/"
  expected:
    media_type: tv_episode
    title: "Steins;Gate"
    episode: 1
```

**Narrator**: The entity test bank is consumed by the title parser's table-driven unit tests and by the `ENT-*` challenge series. When a new naming convention is encountered in the wild, it is added to the test bank first, then the parser is updated to handle it.

### Admin Test Bank

**[Visual: Open `challenges/config/testbank_admin.yaml`]**

**Narrator**: The admin test bank contains 68 test cases covering all administration endpoints: system info retrieval, user management, storage operations, and backup/restore workflows.

```yaml
# challenges/config/testbank_admin.yaml (excerpt)
- id: ADM-001
  name: "System info returns runtime metrics"
  category: admin_system
  endpoint: GET /api/v1/admin/system-info
  requires_role: admin
  expected:
    status_code: 200
    body_contains:
      - version
      - goroutines
      - memory_alloc_mb
      - db_open_connections

- id: ADM-023
  name: "Admin cannot lock own account"
  category: admin_users
  endpoint: PUT /api/v1/admin/users/{self_id}
  requires_role: admin
  request_body:
    is_locked: true
  expected:
    status_code: 400
    error_contains: "cannot modify own account"

- id: ADM-045
  name: "Backup creates valid SQLite file"
  category: admin_backup
  endpoint: POST /api/v1/admin/backups
  requires_role: admin
  expected:
    status_code: 201
    body_contains:
      - filename
      - size
      - created_at
  post_validation:
    - backup_file_exists
    - backup_file_is_valid_sqlite
```

**Narrator**: The admin test bank verifies both positive and negative cases. `ADM-023` specifically tests the self-lockout prevention -- an admin trying to lock their own account must receive a 400 error. `ADM-045` goes beyond HTTP response verification: the `post_validation` step checks that the backup file actually exists on disk and is a valid SQLite database.

### Security Test Bank

**[Visual: Open `challenges/config/testbank_security.yaml`]**

**Narrator**: The security test bank contains 94 test cases organized into authentication, authorization, input validation, and injection prevention categories.

```yaml
# challenges/config/testbank_security.yaml (excerpt)
- id: SEC-001
  name: "Unauthenticated access to protected endpoint"
  category: auth_required
  endpoint: GET /api/v1/admin/users
  headers: {}  # No Authorization header
  expected:
    status_code: 401

- id: SEC-034
  name: "SQL injection via search query"
  category: injection_prevention
  endpoint: GET /api/v1/search
  params:
    query: "'; DROP TABLE files; --"
  expected:
    status_code: 200
    assertion: response_is_valid_json
    negative_assertion: no_sql_error_in_response

- id: SEC-067
  name: "Path traversal in file download"
  category: path_traversal
  endpoint: GET /api/v1/download/file
  params:
    path: "../../../etc/passwd"
  expected:
    status_code: 400
    error_contains: "invalid path"

- id: SEC-088
  name: "JWT with expired token"
  category: token_validation
  endpoint: GET /api/v1/auth/me
  headers:
    Authorization: "Bearer <expired_token>"
  expected:
    status_code: 401
```

**Narrator**: The security test bank is particularly important because security regressions are invisible until exploited. Each case documents a specific attack vector -- SQL injection, path traversal, expired tokens, privilege escalation -- and the expected defensive response. The challenge runner executes these against the live API, ensuring defenses are not accidentally removed during refactoring.

### Performance Test Bank

**[Visual: Open `challenges/config/testbank_performance.yaml`]**

**Narrator**: The performance test bank contains 52 test cases that define response time and throughput SLOs for critical endpoints.

```yaml
# challenges/config/testbank_performance.yaml (excerpt)
- id: PERF-001
  name: "Health endpoint responds within 50ms"
  category: latency
  endpoint: GET /health
  expected:
    max_latency_ms: 50
    status_code: 200

- id: PERF-012
  name: "Deep health check responds within 200ms"
  category: latency
  endpoint: GET /health/deep
  expected:
    max_latency_ms: 200
    status_code: 200

- id: PERF-028
  name: "Entity search responds within 2000ms"
  category: latency
  endpoint: GET /api/v1/entities/search?q=matrix
  expected:
    max_latency_ms: 2000
    status_code: 200

- id: PERF-041
  name: "File list pagination does not degrade with offset"
  category: scalability
  endpoint: GET /api/v1/files?limit=100&offset=10000
  expected:
    max_latency_ms: 500
    status_code: 200
```

**Narrator**: Performance test cases use the `max_latency` assertion evaluator. The challenge runner measures actual response time and fails the case if it exceeds the threshold. `PERF-041` is noteworthy -- it verifies that paginating deep into large result sets does not cause query time to grow linearly, which would indicate a missing index or inefficient `OFFSET` implementation.

### Test Bank Summary

| Bank | Cases | Categories | Key Focus |
|------|-------|------------|-----------|
| Entity | 127 | detection, hierarchy, edge_cases, internationalization | Title parsing accuracy |
| Admin | 68 | system, users, storage, backup | Admin endpoint correctness |
| Security | 94 | auth, authorization, injection, traversal, tokens | Attack surface coverage |
| Performance | 52 | latency, throughput, scalability, resource_usage | SLO enforcement |
| **Total** | **341** | | |

**Narrator**: Combined with the existing API contract (89 cases), conversion (43 cases), and search (44 cases) test banks, Catalogizer now has 517 structured test case definitions. These feed into the challenge system, the unit test generators, and the CI verification pipeline. Every new feature or bug fix starts with a test bank entry, ensuring the specification stays ahead of the implementation.

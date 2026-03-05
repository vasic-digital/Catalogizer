# Module 10: Testing and Quality - Script

**Duration**: 45 minutes
**Module**: 10 - Testing and Quality

---

## Scene 1: Unit Testing Patterns (0:00 - 15:00)

**[Visual: Test coverage summary: Go 38/38 packages, Frontend 101 files/1623 tests, Installer 19 files/178 tests]**

**Narrator**: Welcome to Module 10. Catalogizer maintains zero test failures across all platforms. Go backend: 38 packages with zero race conditions. Frontend: 101 test files with 1623 tests. Installer wizard: 19 files with 178 tests. Let us examine how this level of quality is achieved.

**[Visual: Open a Go table-driven test file]**

**Narrator**: Go tests follow the table-driven pattern. Each test case is a struct with a name, inputs, and expected outputs. A single test function iterates over all cases, running each as a subtest. This pattern is used throughout the codebase.

```go
// catalog-api/database/dialect_test.go
func TestRewritePlaceholders(t *testing.T) {
    tests := []struct {
        name     string
        dialect  DialectType
        input    string
        expected string
    }{
        {
            name:     "sqlite passthrough",
            dialect:  DialectSQLite,
            input:    "SELECT * FROM files WHERE id = ?",
            expected: "SELECT * FROM files WHERE id = ?",
        },
        {
            name:     "postgres single placeholder",
            dialect:  DialectPostgres,
            input:    "SELECT * FROM files WHERE id = ?",
            expected: "SELECT * FROM files WHERE id = $1",
        },
        {
            name:     "postgres multiple placeholders",
            dialect:  DialectPostgres,
            input:    "INSERT INTO files (name, size) VALUES (?, ?)",
            expected: "INSERT INTO files (name, size) VALUES ($1, $2)",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            d := &Dialect{Type: tt.dialect}
            result := d.RewritePlaceholders(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**[Visual: Show test helper for database setup]**

**Narrator**: The test helper in `internal/tests/test_helper.go` provides a `SetupTestDB` function that creates an in-memory SQLite database wrapped with the dialect-aware `database.DB`. This gives every test a clean, isolated database with zero external dependencies.

```go
// catalog-api/internal/tests/test_helper.go
func SetupTestDB(t *testing.T) *database.DB {
    sqlDB, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    db := database.WrapDB(sqlDB, database.DialectSQLite)
    // Run migrations
    require.NoError(t, db.RunMigrations(context.Background()))
    return db
}
```

**[Visual: Show mock vs real database in tests]**

**Narrator**: Catalogizer prefers real in-memory SQLite over mocks for database tests. The dialect abstraction means the same SQL runs against both SQLite and PostgreSQL. In-memory SQLite tests run in milliseconds and catch real SQL bugs that mocks would miss.

**[Visual: Show fuzz tests]**

**Narrator**: The codebase also includes fuzz tests. `dialect_fuzz_test.go` fuzzes the SQL rewriter with random inputs to find edge cases. `title_parser_fuzz_test.go` fuzzes the media title parser. `factory_fuzz_test.go` fuzzes the protocol factory. Fuzz tests catch inputs that unit tests miss.

```go
// catalog-api/database/dialect_fuzz_test.go
func FuzzRewritePlaceholders(f *testing.F) {
    f.Add("SELECT * FROM ? WHERE id = ?")
    f.Add("")
    f.Add("'don''t touch this ?'")
    f.Fuzz(func(t *testing.T, query string) {
        d := &Dialect{Type: DialectPostgres}
        result := d.RewritePlaceholders(query)
        // Should not panic, should not produce empty output for non-empty input
        if query != "" && result == "" {
            t.Errorf("empty result for non-empty input: %q", query)
        }
    })
}
```

**[Visual: Show resource-limited test execution]**

**Narrator**: Tests run with strict resource limits. Go tests use `GOMAXPROCS=3`, 2 parallel packages, and 2 parallel tests per package. This ensures the test suite does not consume more than 30-40% of host resources.

```bash
GOMAXPROCS=3 go test ./... -p 2 -parallel 2
```

---

## Scene 2: Integration Testing (15:00 - 30:00)

**[Visual: Show `docker-compose.test.yml` configuration]**

**Narrator**: Integration tests use a container stack defined in `docker-compose.test.yml`. Three containers run on the host network: the catalog-api, catalog-web, and a Playwright container for browser automation.

**[Visual: Show API contract testing pattern]**

**Narrator**: API contract tests verify that every endpoint the frontend calls exists, returns valid responses, and matches the expected shape. This is the Zero Warning / Zero Error policy: no console errors, no failed network requests, no missing endpoints.

**[Visual: Show integration test files]**

**Narrator**: Integration tests live in `services/services_integration_test.go` and `internal/services/services_integration_test.go`. These tests spin up the full service stack with a real database and test complete workflows: create a storage root, trigger a scan, wait for completion, and verify entities were created.

**[Visual: Show database fixtures]**

**Narrator**: Test fixtures seed the database with known data. A fixture might create a storage root, insert specific files, and then verify that the aggregation service produces the correct entities. Fixtures are defined inline in the test, not in separate files, keeping the test self-contained.

**[Visual: Show end-to-end flow test]**

**Narrator**: End-to-end tests exercise the full stack. A typical test: POST to `/api/v1/auth/login`, save the token, POST to create a storage root, POST to start a scan, poll until complete, GET the file list, GET the entity list, and verify counts and types.

**[Visual: Show Playwright E2E tests]**

**Narrator**: Frontend E2E tests use Playwright to automate a real browser. They click through the UI, fill forms, trigger actions, and verify visual state. The test stack runs in containers with `network_mode: host` so the browser can reach both the API and the web server.

---

## Scene 3: Challenge System (30:00 - 45:00)

**[Visual: Challenge system architecture diagram]**

**Narrator**: The challenge system is Catalogizer's unique quality assurance framework. 209 challenges -- 35 original challenges plus 174 user flow challenges -- verify the entire system end-to-end with 406 assertions. All 209 pass.

**[Visual: Show challenge registration in `catalog-api/challenges/register.go`]**

**Narrator**: Challenges are registered in `register.go`. The `RegisterAll` function loads endpoint configuration, creates challenge instances for each configured storage root, and registers browsing, asset, and populate challenges.

```go
// catalog-api/challenges/register.go
func RegisterAll(svc *services.ChallengeService) error {
    cfg, err := LoadEndpointConfig(DefaultConfigPath())
    if err != nil {
        if os.IsNotExist(err) { return nil }
        return nil
    }

    for _, ep := range cfg.Endpoints {
        svc.Register(NewSMBConnectivityChallenge(&endpoint))
        svc.Register(NewDirectoryDiscoveryChallenge(&endpoint))
        // ... per-directory content type challenges
    }

    svc.Register(NewFirstCatalogPopulateChallenge())
    svc.Register(NewBrowsingAPIHealthChallenge())
    svc.Register(NewBrowsingAPICatalogChallenge())
    // ...
}
```

**[Visual: Show challenge struct pattern]**

**Narrator**: Each challenge is a Go struct embedding `challenge.BaseChallenge` from the Challenges submodule. The `Execute()` method contains the test logic. Challenges interact with the running API through HTTP calls -- never through direct database access or internal function calls.

**[Visual: Show the Challenges submodule structure]**

**Narrator**: The `Challenges/` submodule provides the generic framework. The `pkg/userflow/` package defines multi-platform automation with adapter interfaces for Browser, Mobile, Desktop, API, Build, and Process operations.

**[Visual: Show userflow challenge distribution]**

**Narrator**: 174 user flow challenges span four platforms:

- 49 API challenges in `userflow_api.go` -- HTTP endpoint verification
- 59 Web challenges in `userflow_web.go` -- Playwright browser automation
- 28 Desktop challenges in `userflow_desktop.go` -- Tauri desktop + wizard
- 38 Mobile challenges in `userflow_mobile.go` -- Android + Android TV

**[Visual: Show challenge execution constraints]**

**Narrator**: Critical constraint: `RunAll` is synchronous and blocking. No other challenge can run until it finishes. A 5-minute stale threshold kills stuck challenges. Progress reporting happens every 5 seconds. The entire suite must be executed through the running catalog-api service -- never through scripts or curl commands.

**[Visual: Show challenge API endpoints]**

**Narrator**: Challenges are exposed via REST at `/api/v1/challenges`. You can list all challenges, run a specific one by ID, run all challenges, and retrieve results. The challenge bank definitions live in `challenges/config/`.

**[Visual: Show the CLI runner]**

**Narrator**: The CLI runner at `Challenges/cmd/userflow-runner/` provides command-line execution with flags for platform, report format, compose file, root directory, timeout, and output directory.

```bash
# Run all API challenges
cd Challenges && go run cmd/userflow-runner/main.go --platform api --root /path/to/catalogizer

# Run web challenges with Playwright
go run cmd/userflow-runner/main.go --platform web --compose docker-compose.test.yml
```

**[Visual: Course title card]**

**Narrator**: Testing at this scale requires discipline: table-driven unit tests, in-memory database fixtures, container-based integration tests, Playwright E2E automation, and 209 challenge verifications. The Zero Warning / Zero Error policy ensures no regression goes undetected. In Module 11, we harden the system with security tools and monitoring.

---

## Key Code Examples

### Running All Tests
```bash
# Backend (resource-limited)
cd catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2

# Frontend
cd catalog-web
npm run test           # 1623 tests, single run
npm run test:coverage  # with coverage report

# Installer wizard
cd installer-wizard
npm run test           # 178 tests

# E2E
cd catalog-web
npm run test:e2e       # Playwright
```

### Challenge Configuration
```json
// challenges/config/endpoints.json
{
  "endpoints": [
    {
      "name": "Synology NAS",
      "protocol": "smb",
      "host": "synology.local",
      "directories": [
        { "path": "/music", "content_type": "music" },
        { "path": "/movies", "content_type": "movie" },
        { "path": "/tv", "content_type": "tv_show" }
      ]
    }
  ]
}
```

### Test Summary
```
Go:        38/38 packages, 0 race conditions
Frontend:  101/101 files, 1623 tests, 0 failures
Installer: 19/19 files, 178 tests, 0 failures
Challenges: 209/209 PASSED, 406/406 assertions
Security:  govulncheck 0 vulns, npm audit 0 critical
```

---

## Quiz Questions

1. Why does Catalogizer prefer in-memory SQLite over mocks for database tests?
   **Answer**: In-memory SQLite runs real SQL queries through the dialect abstraction layer, catching actual SQL bugs (syntax errors, incorrect joins, missing columns) that mocks would miss. Tests run in milliseconds because SQLite in-memory is fast. The dialect abstraction ensures the same SQL works on PostgreSQL, so SQLite tests are a valid proxy for production behavior.

2. What is the Zero Warning / Zero Error policy?
   **Answer**: All components must run with zero console warnings, zero console errors, and zero failed network requests in every environment. Every API endpoint the frontend calls must exist and return valid responses. No framework deprecation warnings. No WebSocket connection failures. If a feature is not implemented, stub endpoints return valid empty responses. The 209-challenge suite enforces this end-to-end.

3. How does the challenge system differ from traditional integration tests?
   **Answer**: Challenges execute against the running system through HTTP API calls and browser automation -- exactly as an end user would. They never access databases directly or call internal functions. They verify the complete stack: API, frontend, desktop, and mobile. Traditional integration tests often bypass the HTTP layer. Challenges also have progress reporting, timeout handling, and a bank/registry pattern.

4. What are the resource constraints for running the Go test suite?
   **Answer**: `GOMAXPROCS=3` limits Go to 3 OS threads. `-p 2` runs at most 2 packages in parallel. `-parallel 2` limits per-package test parallelism to 2. This ensures tests use no more than 30-40% of host resources, preventing system freezes on the development machine.

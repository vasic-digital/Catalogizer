# Module 7: Testing and Quality Assurance - Slide Outlines

---

## Slide 7.0.1: Title Slide

**Title**: Testing and Quality Assurance

**Subtitle**: Go Tests, Frontend Tests, E2E Testing, Challenge Framework, and Security Scanning

**Speaker Notes**: This advanced module covers the comprehensive testing strategy used across all Catalogizer components. Students should have completed Module 6 (at least Lessons 6.1-6.2) before starting this module. By the end, students will understand how to write, run, and interpret tests at every level of the stack.

---

## Slide 7.1.1: Go Testing Conventions

**Title**: Test Files Beside Source Files

**Bullet Points**:
- Every test file: `*_test.go` placed beside the file it tests
- Example: `catalog.go` tested by `catalog_test.go`
- Table-driven test pattern: slice of structs with `t.Run` loop
- Three test layers: handlers, services, repositories
- Test helper: `internal/tests/test_helper.go` provides in-memory SQLite via `database.WrapDB()`

**Visual**: Directory listing showing source files paired with their test files

**Speaker Notes**: Go testing conventions are strict in this project. No exceptions to the naming convention. Table-driven tests make it trivial to add new test cases -- each case is a single struct in a slice. The test helper eliminates the need for an external database during testing.

---

## Slide 7.1.2: Table-Driven Test Pattern

**Title**: The Standard Test Structure

**Visual**: Code example showing the table-driven pattern:

```go
tests := []struct {
    name     string
    input    string
    expected int
    wantErr  bool
}{
    {"valid input", "test", 200, false},
    {"empty input", "", 400, true},
    {"special chars", "a&b", 200, false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := myFunc(tt.input)
        // assertions
    })
}
```

**Speaker Notes**: This pattern is used everywhere in the Go codebase. Each test case has a descriptive name that appears in the test output. Adding a new edge case means adding one struct to the slice. The t.Run wrapper gives each case its own sub-test with independent pass/fail status.

---

## Slide 7.1.3: Test Layers

**Title**: Handler, Service, and Repository Tests

| Layer | Example Files | What It Tests |
|-------|--------------|---------------|
| Handler | `auth_handler_test.go`, `media_handler_test.go` | HTTP contract: status codes, response format, headers |
| Service | `catalog_test.go`, `favorites_service_test.go`, `playlist_service_test.go` | Business logic: inputs produce expected outputs |
| Repository | `user_repository_test.go`, `favorites_repository_test.go`, `sync_repository_test.go` | Database operations: CRUD, queries, constraints |

**Speaker Notes**: Each layer tests a specific concern. Handler tests verify the HTTP interface. Service tests verify business rules. Repository tests verify data persistence. Testing all three layers catches bugs at the narrowest possible scope.

---

## Slide 7.1.4: Running Go Tests

**Title**: Commands and Resource Limits

**Bullet Points**:
- All tests: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- Single test: `go test -v -run TestName ./path/to/pkg/`
- With coverage: `go test -cover ./...`
- HTML coverage report: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- Race detector: `go test -race ./...`
- **Resource limits are mandatory**: 30-40% max host usage

**Visual**: Terminal output showing 38/38 packages passing

**Speaker Notes**: The resource limits are non-negotiable. GOMAXPROCS=3 limits OS threads. The -p 2 limits parallel package testing. The -parallel 2 limits parallelism within a package. Exceeding these can freeze the host machine. Current status: 38 packages, 0 failures, 0 race conditions.

---

## Slide 7.2.1: Frontend Testing with Vitest

**Title**: React and TypeScript Testing

**Bullet Points**:
- Framework: Vitest (Jest-compatible API, native Vite support)
- Location: `__tests__/` directories alongside components
- Naming: `ComponentName.test.tsx` or `filename.test.ts`
- Rendering: React Testing Library (`render`, `screen.getByText`, `screen.getByRole`)
- Mocking: Vitest built-in mocking for API calls and modules
- Hooks: `renderHook` for testing custom React hooks

**Speaker Notes**: Vitest was chosen over Jest for its native Vite integration, which gives faster startup and built-in TypeScript support. The testing patterns mirror Jest exactly, so anyone familiar with Jest is immediately productive.

---

## Slide 7.2.2: Frontend Test Categories

**Title**: What Gets Tested

| Category | Examples | Count |
|----------|---------|-------|
| UI Components | Badge, Progress, Select, Switch, Tabs, Textarea | Low-level rendering |
| Page Components | Collections, Favorites, Playlists, Admin, AIDashboard | Full page behavior |
| API Layer | api.test.ts, collectionsApi.test.ts, favoritesApi.test.ts | Request/response |
| Hooks | Custom hook tests with renderHook | State logic |
| Installer Wizard | Wizard components, flow tests | 19 files, 178 tests |

**Bullet Points**:
- catalog-web: 101 test files, 1623 tests
- installer-wizard: 19 test files, 178 tests
- All tests pass with zero failures
- Zero warnings policy enforced

**Speaker Notes**: The test suite is comprehensive. UI components verify rendering with specific props. Page components verify full user workflows including API interactions. API tests verify request construction without needing a running backend. The installer wizard has its own independent test suite.

---

## Slide 7.2.3: Frontend Quality Commands

**Title**: Lint, Type Check, and Test

| Command | Purpose | Expected Result |
|---------|---------|----------------|
| `npm run test` | Run all Vitest tests (single run) | 1623 tests pass |
| `npm run test:watch` | Continuous testing during development | Reruns on file change |
| `npm run test:coverage` | Tests with coverage report | Coverage percentages |
| `npm run lint` | ESLint across codebase | 0 errors |
| `npm run type-check` | TypeScript compilation check | 0 errors |
| `npm run test:e2e` | Playwright end-to-end tests | All scenarios pass |

**Speaker Notes**: All of these must pass before merging any changes. The lint and type-check commands catch issues that tests do not -- unused imports, type mismatches, style violations. Running all three is a habit, not an option.

---

## Slide 7.3.1: Playwright End-to-End Testing

**Title**: Testing from the User's Perspective

**Bullet Points**:
- Real browser automation: navigates pages, clicks buttons, fills forms
- Test stack: `docker-compose.test.yml` with `network_mode: host`
- Services: catalog-api + catalog-web + Playwright in coordinated containers
- Catches integration issues unit tests miss
- Screenshots on failure for debugging

**Visual**: Diagram: Playwright Browser -> catalog-web (port 3000) -> catalog-api (port 8080) -> Database

**Speaker Notes**: Playwright tests are the final verification before deployment. They exercise the entire stack as a real user would. If a handler works but the frontend sends the wrong request format, Playwright catches it. If CSS hides a critical button, Playwright catches it.

---

## Slide 7.3.2: Challenge Framework Overview

**Title**: 209 Structured Test Scenarios

**Bullet Points**:
- **35 original challenges** (CH-001 to CH-035): core functionality, 117 assertions
- **174 userflow challenges** (UF prefix): multi-platform automation
  - 49 API challenges (HTTP endpoints)
  - 59 Web challenges (Playwright browser)
  - 28 Desktop challenges (Tauri applications)
  - 38 Mobile challenges (Android apps)
- Registered in `catalog-api/challenges/register.go` via `RegisterAll()`
- Exposed via REST: `/api/v1/challenges`

**Visual**: Table showing the breakdown: 35 + 49 + 59 + 28 + 38 = 209

**Speaker Notes**: The challenge system is unique to Catalogizer. Each challenge is a Go struct that tests real system behavior against a running instance. Unlike unit tests which test code in isolation, challenges test the deployed system end-to-end. The 209 challenges with 406 assertions provide comprehensive verification.

---

## Slide 7.3.3: Challenge Execution Constraints

**Title**: Running Challenges Safely

**Bullet Points**:
- **RunAll is synchronous/blocking**: no other challenge runs until it finishes
- **Stale threshold**: 5 minutes -- kills stuck challenges with no progress
- **Progress reporting**: challenges report progress to avoid stale detection
- **Default timeout**: `challenge.NewConfig()` sets 5 min -- zero it for runner's timeout
- **config.json write_timeout**: must be 900 (not 30) for RunAll
- All operations MUST run through system deliverables (compiled binaries)

**Speaker Notes**: These constraints exist for stability. RunAll being synchronous prevents resource contention. The stale threshold prevents runaway challenges from blocking the system. Progress reporting is the escape valve for long-running operations. Never use custom scripts or curl commands within challenge execution.

---

## Slide 7.3.4: Userflow Automation Framework

**Title**: Testing Across All Platforms

| Component | Count |
|-----------|-------|
| Adapter Interfaces | 6: Browser, Mobile, Desktop, API, Build, Process |
| CLI Implementations | 9: Playwright, ADB, Tauri, HTTP, Gradle, npm, Go, Cargo, Process |
| Challenge Templates | 13: Env, Build, UnitTest, Lint, APIHealth, APIFlow, BrowserFlow, etc. |
| Evaluators | 12: build_succeeds, all_tests_pass, lint_passes, app_launches, etc. |

**Bullet Points**:
- Framework in `Challenges/pkg/userflow/` -- zero project-specific references
- CLI runner: `Challenges/cmd/userflow-runner/`
- Flags: `--platform`, `--report`, `--compose`, `--root`, `--timeout`, `--output`
- 209 tests across 51 test files in the framework itself

**Speaker Notes**: The userflow framework is generic and reusable. It has no Catalogizer-specific code. The adapters and templates can test any application that has a web frontend, API backend, mobile app, or desktop app. The Catalogizer-specific challenges are defined in catalog-api/challenges/.

---

## Slide 7.4.1: Security Scanning

**Title**: Zero Vulnerability Approach

| Tool | Scope | Command | Current Status |
|------|-------|---------|----------------|
| govulncheck | Go dependencies | `govulncheck ./...` | 0 vulnerabilities |
| npm audit | npm packages | `npm audit --production` | 0 critical production vulns |
| gosec | Go static analysis | `gosec ./...` | Security issues flagged |
| security-test.sh | Comprehensive | `scripts/security-test.sh` | Full security suite |

**Bullet Points**:
- govulncheck: call-graph analysis -- only reports reachable vulnerabilities
- npm audit --production: excludes dev-only dependencies
- docker-compose.security.yml: dedicated security scanning environment
- Run security scans weekly, not just at release time

**Speaker Notes**: The zero vulnerability target is aspirational but tracked rigorously. govulncheck's call-graph analysis is important -- it tells you which vulnerabilities actually affect your code, not just which ones exist in your dependency tree. This reduces false positives significantly.

---

## Slide 7.4.2: Quality Metrics and Benchmarks

**Title**: Measuring and Maintaining Quality

**Bullet Points**:
- **Zero warning/zero error policy**: no console errors, no failed network requests, no deprecation warnings
- **Benchmarks**: `providers_bench_test.go`, `auth_service_bench_test.go`
  - Run with `go test -bench=. ./path/`
  - Output: ns/op, B/op, allocs/op
- **Stress testing**: `stress_test_service.go` and `stress_test_handler.go`
- **Memory leak detection**: `scripts/memory-leak-check.sh`
- **Coverage validation**: `scripts/validate-coverage.sh`

**Speaker Notes**: Quality is not just about tests passing. The zero warning policy means every API endpoint the frontend calls must exist and return valid responses. Benchmarks catch performance regressions before they reach production. Memory leak detection prevents gradual degradation.

---

## Slide 7.4.3: Complete Testing Pipeline

**Title**: scripts/run-all-tests.sh

**Visual**: Pipeline flowchart:

```
Go Tests (38 packages) -> Frontend Tests (1623 tests) -> Installer Tests (178 tests)
     |                          |                              |
     v                          v                              v
Race Detection         Lint + Type Check               Security Scans
     |                          |                              |
     v                          v                              v
                    Coverage Validation
                           |
                           v
                    PASS / FAIL
```

**Bullet Points**:
- Runs all component test suites in sequence
- Includes security scans and coverage validation
- Must pass before any code is merged
- Current totals: 38 Go packages + 101 frontend files + 19 installer files = all green

**Speaker Notes**: This script is the single source of truth for code quality. If it passes, the code is ready. If it fails, something needs attention. Run it before submitting any changes.

---

## Slide 7.4.4: Module 7 Summary

**Title**: What We Covered

**Bullet Points**:
- Go testing: table-driven tests, handler/service/repository layers, resource-limited execution
- Frontend testing: Vitest with React Testing Library, 1623 tests across 101 files
- E2E testing: Playwright with docker-compose.test.yml test stack
- Challenge framework: 209 challenges with 406 assertions across 4 platforms
- Security: govulncheck (0 vulns), npm audit (0 critical), gosec, dedicated security environment
- Quality metrics: zero warning policy, benchmarks, stress testing, memory leak detection

**Speaker Notes**: Testing in Catalogizer is not an afterthought -- it is a first-class concern at every layer. The combination of unit tests, integration tests, E2E tests, and the challenge framework provides confidence that the system works correctly across all components and platforms.

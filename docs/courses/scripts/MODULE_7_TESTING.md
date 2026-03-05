# Module 7: Testing and Quality Assurance - Video Scripts

---

## Lesson 7.1: Go Backend Testing

**Duration**: 18 minutes

### Narration

Welcome to Module 7 on Testing and Quality Assurance. This is an advanced module that builds on the developer foundations from Module 6. In this first lesson, we are going to dive deep into Go backend testing.

Catalogizer follows strict Go testing conventions. Every test file is named with a _test.go suffix and placed beside the source file it tests. This is not a project convention -- it is the Go language convention, and this project follows it without exception.

The dominant pattern throughout the codebase is table-driven tests. Let me show you what this looks like. You define a slice of anonymous structs, each representing a test case with a name, inputs, and expected outputs. Then you loop through the cases with t.Run, which gives each case its own sub-test. This approach has several advantages: adding a new test case is a single struct addition, test names appear in output making failures easy to locate, and the pattern is instantly recognizable to any Go developer.

Let me walk through the test layers. At the handler level, we test HTTP request handling. Open auth_handler_test.go. You can see tests that create a Gin test context with httptest.NewRecorder, set up the request with the appropriate method, headers, and body, call the handler function, and then assert on the response status code and body. Handler tests verify that the HTTP contract is correct -- the right status codes for success and error cases, the right response format, the right headers.

At the service level, we test business logic. Open catalog_test.go. Service tests focus on the logic: given these inputs, do we get the expected output? Do error conditions produce the right error types? Services are where most of the complexity lives, so these tests tend to be the most thorough.

Repository tests verify database interactions. Open user_repository_test.go. These tests use the test helper in internal/tests/test_helper.go, which provides an in-memory SQLite database via database.WrapDB. This function wraps a standard sql.DB with the dialect abstraction layer, so your repository code works identically whether talking to SQLite or PostgreSQL. The test helper sets up the schema and returns a ready-to-use database connection. No external database required.

To run all backend tests, use the resource-limited command: GOMAXPROCS=3 go test ./... -p 2 -parallel 2. The GOMAXPROCS limits the number of OS threads. The -p flag limits the number of packages tested in parallel. The -parallel flag limits parallelism within a single package. These limits are mandatory because the host machine runs other critical processes, and exceeding 30-40% resource usage can freeze the system.

To run a specific test, use go test -v -run TestName ./path/to/pkg/. The -v flag gives verbose output with each sub-test's pass or fail status. The -run flag accepts a regex, so TestCatalog would match TestCatalogList, TestCatalogSearch, and any other test starting with TestCatalog.

For measuring coverage, add the -cover flag: go test -cover ./.... For a detailed HTML report, use go test -coverprofile=coverage.out ./... followed by go tool cover -html=coverage.out. The HTML report highlights which lines are covered and which are not.

The race detector is another essential tool. Run tests with -race to detect data races: go test -race ./.... The Catalogizer test suite reports zero race conditions with 38 packages all passing.

### On-Screen Actions

- [00:00] Show title: "Go Backend Testing"
- [00:30] Open catalog-api directory -- show *_test.go files beside source files
- [01:00] Open a test file -- highlight the _test.go naming convention
- [01:30] Show a table-driven test pattern: struct slice with test cases
- [02:30] Walk through the t.Run loop executing each case
- [03:30] Open handlers/auth_handler_test.go -- show handler test structure
- [04:00] Highlight httptest.NewRecorder and Gin test context setup
- [04:30] Show assertion on response status code and body
- [05:00] Open services/catalog_test.go -- show service-level tests
- [05:30] Show business logic test cases
- [06:30] Open repository/user_repository_test.go -- show repository tests
- [07:00] Open internal/tests/test_helper.go -- show database.WrapDB usage
- [07:30] Explain in-memory SQLite for testing
- [08:00] Open a terminal and run: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- [09:00] Show all 38 packages passing
- [09:30] Run a single test: `go test -v -run TestCatalog ./internal/services/`
- [10:00] Show verbose output with sub-test names
- [10:30] Run with coverage: `go test -cover ./...`
- [11:00] Generate HTML coverage report: `go test -coverprofile=coverage.out ./...`
- [11:30] Open the HTML coverage report in a browser
- [12:00] Show covered and uncovered lines highlighted in green and red
- [12:30] Run with race detector: `go test -race ./...`
- [13:00] Show zero race conditions detected
- [13:30] Open services/favorites_service_test.go -- show another service test
- [14:00] Open services/playlist_service_test.go -- show playlist tests
- [14:30] Open services/subtitle_service_test.go -- show subtitle tests
- [15:00] Open repository/favorites_repository_test.go -- show repository pattern
- [15:30] Open repository/sync_repository_test.go -- show sync tests
- [16:00] Show the test helper creating clean database state for each test
- [16:30] Explain GOMAXPROCS, -p, and -parallel resource limits
- [17:00] Recap: 38 packages, 0 failures, 0 race conditions

### Key Points

- Test files: *_test.go placed beside source files (Go convention)
- Table-driven tests: slice of structs with t.Run loop for each case
- Handler tests: httptest.NewRecorder with Gin test context
- Service tests: business logic verification with expected inputs and outputs
- Repository tests: in-memory SQLite via internal/tests/test_helper.go and database.WrapDB
- Resource limits mandatory: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- Coverage: `-cover` flag or `-coverprofile` for HTML reports
- Race detector: `-race` flag detects concurrent access issues
- Current status: 38/38 packages pass, 0 race conditions

### Tips

> **Tip**: When writing a new test, copy the structure from an existing test in the same package. The patterns are consistent, and matching them ensures your test integrates cleanly with the suite.

> **Tip**: Use the race detector in CI but not during rapid development iteration -- it slows tests significantly. Run it before committing.

---

## Lesson 7.2: Frontend and Client Testing

**Duration**: 15 minutes

### Narration

Now let us look at testing on the frontend side. Catalogizer uses Vitest as its test framework for the React frontend and TypeScript projects. Vitest is compatible with Jest's API but built on top of Vite, giving it fast startup and native TypeScript support.

Frontend tests live in __tests__ directories alongside the components they test. The naming convention is ComponentName.test.tsx for React component tests and filename.test.ts for utility and library tests.

Let me show you the different categories of frontend tests. First, UI component tests. Open the __tests__ directory under components. You will find tests for Badge, Progress, Select, Switch, Tabs, Textarea, and more. These tests render the component with specific props and assert on the rendered output. They use React Testing Library's render function and queries like getByText, getByRole, and getByTestId.

Page-level component tests are more involved. Open the tests for Collections, Favorites, Playlists, Admin Panel, and AI Dashboard. These tests verify that entire page components render correctly, handle user interactions, and display the right data. They often mock API calls using Vitest's mocking capabilities to provide controlled test data.

Library and API tests verify the data fetching layer. Open api.test.ts, collectionsApi.test.ts, and favoritesApi.test.ts. These test the functions that make HTTP requests to the backend, verifying request construction, parameter handling, and response parsing.

To run all frontend tests, use npm run test from the catalog-web directory. This runs Vitest in single-run mode. For development, use npm run test:watch to get continuous test execution as you change files. For coverage reports, use npm run test:coverage.

Linting and type checking are equally important. Run npm run lint to execute ESLint across the codebase. This catches style issues, unused variables, missing imports, and React-specific problems. Run npm run type-check to verify TypeScript compilation without emitting files. Both must pass with zero errors.

The API client library has its own test suite in catalogizer-api-client. Navigate to that directory, run npm install, then npm run build to compile, and npm run test to execute. These tests verify the client's authentication flow, API method signatures, request construction, and response handling.

The installer wizard also has a comprehensive test suite. Navigate to installer-wizard and run npm run test. It covers 19 test files with 178 tests for the wizard components and flow.

The current test count across all frontend projects: catalog-web has 101 test files with 1623 individual tests. The installer wizard adds 19 files with 178 tests. All pass with zero failures.

### On-Screen Actions

- [00:00] Show title: "Frontend and Client Testing"
- [00:30] Open catalog-web/src -- show __tests__ directories
- [01:00] Open a component test: Badge.test.tsx or Progress.test.tsx
- [01:30] Show the test structure: describe block, test cases, render, assertions
- [02:00] Show React Testing Library usage: render, screen.getByText, expect
- [02:30] Open a page-level test: Collections.test.tsx
- [03:00] Show API mocking with Vitest
- [03:30] Show assertions on rendered page content
- [04:00] Open api.test.ts -- show API layer tests
- [04:30] Show request construction verification
- [05:00] Show collectionsApi.test.ts and favoritesApi.test.ts
- [05:30] Run `npm run test` in catalog-web
- [06:00] Show test output: 101 files, 1623 tests, all passing
- [06:30] Run `npm run test:watch` -- show watch mode with file change detection
- [07:00] Run `npm run test:coverage` -- show coverage output
- [07:30] Run `npm run lint` -- show ESLint output with zero errors
- [08:00] Run `npm run type-check` -- show TypeScript verification
- [08:30] Navigate to catalogizer-api-client
- [09:00] Run `npm run build && npm run test`
- [09:30] Show API client tests passing
- [10:00] Navigate to installer-wizard
- [10:30] Run `npm run test` -- show 19 files, 178 tests passing
- [11:00] Return to catalog-web and open a hook test
- [11:30] Show testing custom React hooks with renderHook
- [12:00] Open the Vitest configuration in vite.config.ts
- [12:30] Show the test configuration: environment, globals, setup files
- [13:00] Show the path aliases working in tests: @/components, @/hooks, @/lib
- [13:30] Explain the zero-warning policy for tests
- [14:00] Recap all test counts and verification steps

### Key Points

- Framework: Vitest for all TypeScript/React testing (Jest-compatible API, Vite-native)
- Convention: __tests__/ directories with ComponentName.test.tsx naming
- UI component tests: render with props, assert with React Testing Library queries
- Page-level tests: mock API calls, verify rendering and user interactions
- API tests: verify request construction, parameter handling, response parsing
- Run commands: `npm run test` (single), `npm run test:watch` (dev), `npm run test:coverage`
- Linting: `npm run lint` (ESLint) and `npm run type-check` (TypeScript) must pass with zero errors
- API client: `npm run build && npm run test` in catalogizer-api-client
- Current totals: 101 files / 1623 tests (catalog-web) + 19 files / 178 tests (installer-wizard)

### Tips

> **Tip**: When a component test breaks, check whether the component's API changed. Update the test to match the new API rather than changing the component to make old tests pass. Tests should reflect intended behavior.

> **Tip**: Use npm run test:watch during active development. It only reruns tests affected by your changes, giving fast feedback without running the entire suite.

---

## Lesson 7.3: End-to-End and Challenge Framework Testing

**Duration**: 15 minutes

### Narration

Beyond unit tests, Catalogizer has two powerful system-level testing mechanisms: Playwright end-to-end tests and the challenge framework.

Playwright tests verify the application from the user's perspective. They launch a real browser, navigate pages, click buttons, fill forms, and assert on visible results. This catches integration issues that unit tests miss -- a handler might work perfectly in isolation, but the frontend might send the wrong request format, or a CSS class might hide a critical button.

To run the Playwright E2E tests: npm run test:e2e from the catalog-web directory. The test configuration is in the Playwright config file. Tests use the docker-compose.test.yml stack which runs catalog-api, catalog-web, and Playwright all with network_mode: host so they can communicate directly.

Now let me explain the challenge framework, which is unique to Catalogizer. The Challenges submodule at the project root provides a structured test scenario system. Challenges are Go structs that embed challenge.BaseChallenge and implement a custom Execute method. They test real system behavior against a running Catalogizer instance.

There are 209 total challenges. The original 35 challenges (CH-001 through CH-035) test core functionality: service startup, storage root creation, scanning, media detection, entity aggregation, favorites, collections, playlists, API endpoints, and more. These produce 117 assertions that all must pass.

The 174 userflow challenges (UF prefix) test across 4 platform groups: 49 API challenges test HTTP endpoints directly, 59 web challenges use Playwright to test the browser interface, 28 desktop challenges test the Tauri applications, and 38 mobile challenges test the Android apps.

Challenges are registered in catalog-api/challenges/register.go via RegisterAll(). They are exposed via REST endpoints under /api/v1/challenges. You can list challenges, run individual challenges, or run all challenges. An important constraint: RunAll is synchronous and blocking. No other challenge can execute until it finishes. The runner has a stale threshold of 5 minutes -- if a challenge reports no progress for 5 minutes, it is killed as stuck.

The challenge system uses the progress-based liveness detection. Each challenge can report progress during execution. The populate challenge, for example, reports progress every 5 seconds during scan polling. This prevents the runner from killing long-running but healthy challenges.

To run the challenges, start the Catalogizer backend and then use the challenge API:

```
POST /api/v1/challenges/run-all
GET /api/v1/challenges/status
GET /api/v1/challenges/results
```

All challenge operations must be executed by system deliverables -- the running catalog-api service. Never use custom scripts or curl commands to trigger API endpoints within challenge execution.

The userflow automation framework in Challenges/pkg/userflow/ provides 6 adapter interfaces (Browser, Mobile, Desktop, API, Build, Process) and 9 CLI adapter implementations (Playwright, ADB, Tauri, HTTP, Gradle, npm, Go, Cargo, Process). This framework enables testing across all platforms through a unified interface.

### On-Screen Actions

- [00:00] Show title: "End-to-End and Challenge Framework Testing"
- [00:30] Open catalog-web -- show Playwright configuration
- [01:00] Show docker-compose.test.yml with network_mode: host
- [01:30] Run `npm run test:e2e` -- show Playwright launching a browser
- [02:00] Show a Playwright test navigating and interacting with the UI
- [02:30] Show test results with screenshots of failures
- [03:00] Navigate to the Challenges submodule
- [03:30] Open Challenges/pkg/challenge/ -- show BaseChallenge struct
- [04:00] Open a challenge implementation -- show Execute method
- [04:30] Open catalog-api/challenges/register.go -- show RegisterAll
- [05:00] Show the 35 original challenges listed in registration
- [05:30] Open catalog-api/challenges/userflow_api.go -- show 49 API challenges
- [06:00] Open catalog-api/challenges/userflow_web.go -- show 59 web challenges
- [06:30] Open catalog-api/challenges/userflow_desktop.go -- show 28 desktop challenges
- [07:00] Open catalog-api/challenges/userflow_mobile.go -- show 38 mobile challenges
- [07:30] Show the challenge REST endpoints in the API
- [08:00] Start the backend and trigger a challenge run via the API
- [08:30] Show challenge progress being reported in real time
- [09:00] Show challenge results: 209/209 passed, 406/406 assertions
- [09:30] Open the Challenges/pkg/userflow/ directory
- [10:00] Show the 6 adapter interfaces: Browser, Mobile, Desktop, API, Build, Process
- [10:30] Show the 9 CLI adapter implementations
- [11:00] Show the 13 challenge templates
- [11:30] Open Challenges/cmd/userflow-runner -- show the CLI runner
- [12:00] Show runner flags: --platform, --report, --compose, --timeout
- [12:30] Explain the stale threshold and progress-based liveness detection
- [13:00] Explain RunAll is synchronous and blocking
- [13:30] Show challenge.NewConfig() and the 5-minute default timeout
- [14:00] Recap: 209 challenges, 406 assertions, zero failures

### Key Points

- Playwright E2E tests: real browser testing via `npm run test:e2e`
- Test stack: docker-compose.test.yml with network_mode: host for all services
- Challenge framework: 209 total challenges (35 original + 174 userflow)
- Original challenges (CH-001 to CH-035): 117 assertions testing core functionality
- Userflow challenges: 49 API + 59 web + 28 desktop + 38 mobile
- Challenges registered in catalog-api/challenges/register.go
- REST endpoints: /api/v1/challenges for listing, running, and viewing results
- RunAll is synchronous/blocking -- one at a time
- Stale threshold: 5 minutes with progress-based liveness detection
- Userflow framework: 6 adapter interfaces, 9 CLI implementations, 13 templates
- All operations must run through system deliverables, never custom scripts

### Tips

> **Tip**: Run the original 35 challenges (CH-001 to CH-035) first before attempting RunAll. They validate the core system and finish much faster than the full 209-challenge suite.

> **Tip**: If a challenge appears stuck, check the stale threshold. The challenge may have hit the 5-minute progress reporting deadline. Ensure long-running operations report progress regularly.

---

## Lesson 7.4: Security Testing and Quality Metrics

**Duration**: 12 minutes

### Narration

The final piece of the testing strategy is security scanning and quality metrics. Catalogizer takes a zero-vulnerability approach to security.

For Go dependencies, the primary tool is govulncheck. Run it from the catalog-api directory: govulncheck ./.... This tool checks all Go module dependencies against the Go vulnerability database. It reports only vulnerabilities that actually affect your code -- not just every vulnerability in every dependency, but specifically the ones reachable from your call graph. The current status is zero vulnerabilities.

For frontend dependencies, npm audit checks the npm package tree. Run npm audit --production from the catalog-web directory. The --production flag limits the scan to production dependencies, excluding dev-only tools. The current status is zero critical production vulnerabilities.

Beyond dependency scanning, there are dedicated security test scripts. The scripts/security-test.sh script runs a comprehensive security test suite. It is part of the docker-compose.security.yml environment, which provides a dedicated container setup for security scanning.

Static analysis with gosec catches common security issues in Go code: SQL injection risks, hardcoded credentials, insecure TLS configurations, and more. Run it with gosec ./... from the catalog-api directory.

For code quality metrics, the project tracks several indicators. Test coverage across all components. The zero warning/zero error policy means every component must run with zero console warnings, zero console errors, and zero failed network requests. Every API endpoint the frontend calls must exist, return valid 2xx responses, and match the expected shape. No framework deprecation warnings. No WebSocket connection failures.

Performance testing is available through the stress test service. The stress_test_service.go and stress_test_handler.go provide endpoints for load testing. Benchmark tests in providers_bench_test.go and auth_service_bench_test.go measure the performance of critical code paths.

Memory leak detection is available via scripts/memory-leak-check.sh. This monitors the application's memory usage over time to detect gradual increases that indicate leaks.

The scripts/validate-coverage.sh script checks that test coverage meets minimum thresholds. And scripts/run-all-tests.sh ties everything together by running Go tests, frontend tests, security scans, and validation in sequence. This is the script that should pass before any code is merged.

### On-Screen Actions

- [00:00] Show title: "Security Testing and Quality Metrics"
- [00:30] Run `govulncheck ./...` in catalog-api
- [01:00] Show output: 0 vulnerabilities found
- [01:30] Explain how govulncheck traces the call graph
- [02:00] Run `npm audit --production` in catalog-web
- [02:30] Show output: 0 critical production vulnerabilities
- [03:00] Open scripts/security-test.sh -- show what it executes
- [03:30] Open docker-compose.security.yml -- show security scanning containers
- [04:00] Run `gosec ./...` in catalog-api -- show static analysis output
- [04:30] Show any findings and how to address them
- [05:00] Explain the zero warning/zero error policy
- [05:30] Show a browser with zero console errors and zero failed network requests
- [06:00] Open internal/services/stress_test_service.go -- show load testing
- [06:30] Open internal/handlers/stress_test_handler.go -- show the API endpoints
- [07:00] Open internal/media/providers/providers_bench_test.go -- show benchmarks
- [07:30] Run a benchmark: `go test -bench=. ./internal/media/providers/`
- [08:00] Show benchmark output with ns/op measurements
- [08:30] Open services/auth_service_bench_test.go -- show auth benchmarks
- [09:00] Show scripts/memory-leak-check.sh
- [09:30] Show scripts/validate-coverage.sh
- [10:00] Open scripts/run-all-tests.sh -- show the complete pipeline
- [10:30] Run the complete test suite and show summary output
- [11:00] Recap all quality metrics and their current status

### Key Points

- govulncheck: Go dependency vulnerability scanning with call-graph analysis (0 vulns)
- npm audit --production: frontend dependency scanning (0 critical production vulns)
- gosec: Go static security analysis for SQL injection, hardcoded credentials, etc.
- scripts/security-test.sh + docker-compose.security.yml: dedicated security testing environment
- Zero warning/zero error policy: no console errors, no failed requests, no deprecation warnings
- Stress testing: stress_test_service.go and stress_test_handler.go for load testing
- Benchmarks: providers_bench_test.go and auth_service_bench_test.go for performance measurement
- Memory leak detection: scripts/memory-leak-check.sh
- Coverage validation: scripts/validate-coverage.sh
- Complete pipeline: scripts/run-all-tests.sh runs everything in sequence

### Tips

> **Tip**: Run govulncheck weekly even if no code changed. New vulnerabilities are discovered in existing dependencies regularly. Catching them early gives you time to update before they become urgent.

> **Tip**: The zero warning/zero error policy is enforced by the challenge suite (CH-001 to CH-020+). If you introduce a new API endpoint, make sure the frontend handles it correctly -- a missing endpoint or wrong response shape will fail a challenge.

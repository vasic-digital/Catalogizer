# Module 10: Advanced Testing - Video Scripts

---

## Lesson 10.1: Fuzz Testing with Go

**Duration**: 18 minutes

### Narration

Welcome to Module 10, the final advanced module. This module covers testing techniques that go beyond standard unit and integration tests -- fuzz testing, property-based testing, stress and chaos testing, security testing, and visual regression testing. These techniques collectively contribute to the 239 registered challenges that validate Catalogizer end to end.

In this first lesson, we focus on fuzz testing. Go has built-in fuzz testing support since version 1.18. Fuzz testing works by providing seed inputs, then automatically mutating those inputs to discover edge cases, crashes, and unexpected behavior. Unlike unit tests that check specific input/output pairs, fuzz tests explore the input space algorithmically.

A fuzz function has the signature func FuzzXxx(f *testing.F). The testing.F type provides methods for adding seed corpus entries with f.Add() and defining the fuzz target with f.Fuzz(). The fuzzer starts with your seeds and mutates them -- flipping bits, inserting bytes, removing characters, combining seeds -- to generate thousands of test inputs.

When the fuzzer discovers an input that causes a crash, it saves that input to the testdata/fuzz/ directory. This crash input becomes a permanent regression test. Every subsequent test run replays all saved corpus entries, ensuring the bug never reappears.

Catalogizer has fuzz tests for several critical components. Let me walk through each one.

The title parser fuzz test is in internal/services/title_parser_fuzz_test.go. The title parser extracts structured information from filenames -- movie titles, TV show seasons and episodes, music artist and album names. Arbitrary filenames can contain any character sequence, making this an ideal fuzz target. The fuzz test seeds include standard naming patterns like "Movie.Name.2024.1080p.mkv" and "Show.S01E02.Episode.Name.mp4", then lets the fuzzer generate thousands of variations. The test verifies that the parser never panics on any input, always returns a valid result structure, and handles edge cases like empty strings, extremely long strings, and unicode characters.

The dialect rewriter fuzz test is in database/dialect_fuzz_test.go. The dialect abstraction rewrites SQL queries between SQLite and PostgreSQL formats: converting ? placeholders to $1, $2 numbered parameters, rewriting INSERT OR IGNORE to ON CONFLICT DO NOTHING, and converting boolean literals. The fuzz test ensures that the rewriter never produces invalid SQL, never panics on malformed input, and handles all combinations of placeholders and SQL keywords.

The filesystem factory fuzz test is in filesystem/factory_fuzz_test.go. The factory creates protocol-specific clients based on connection strings. Fuzzing ensures it handles malformed URLs, unusual protocol prefixes, and edge cases in credential parsing without panicking.

The download handler fuzz test is in internal/handlers/download_fuzz_test.go. Download paths can contain path traversal attempts, null bytes, and encoding tricks. The fuzz test verifies that the handler correctly rejects all forms of path traversal.

The input sanitizer fuzz test is in middleware/fuzz_test.go. The sanitization middleware processes user input before it reaches handlers. Fuzzing ensures it handles all forms of potentially dangerous input without crashing.

When running fuzz tests, always respect the project resource limits. Use GOMAXPROCS=3 go test -fuzz=FuzzXxx -p 2 -parallel 2 to stay within the 30-40% host resource budget. The -fuzztime flag controls how long the fuzzer runs: -fuzztime=60s for a quick check, -fuzztime=300s for a more thorough exploration. Longer runs explore more of the input space.

Saved crash inputs in testdata/fuzz/ are committed to version control. They serve as regression tests -- the go test command replays all corpus entries even without the -fuzz flag. This means every CI run verifies that all previously discovered crashes remain fixed.

### On-Screen Actions

- [00:00] Show title: "Fuzz Testing with Go"
- [00:30] Explain the concept: seed inputs, mutation, crash discovery
- [01:00] Show the testing.F API: f.Add() for seeds, f.Fuzz() for the target
- [01:30] Open internal/services/title_parser_fuzz_test.go
- [02:00] Show seed corpus entries: standard naming patterns
- [02:30] Show the fuzz target function: parse and verify no panic
- [03:00] Run: `GOMAXPROCS=3 go test -fuzz=FuzzTitleParser -fuzztime=30s ./internal/services/ -p 2 -parallel 2`
- [03:30] Show the fuzzer running: iterations per second, total inputs tested
- [04:00] Show a crash being discovered (or explain what happens when one is found)
- [04:30] Show testdata/fuzz/ directory -- saved crash inputs
- [05:00] Open database/dialect_fuzz_test.go
- [05:30] Show SQL rewriting fuzz target: verify output is valid SQL
- [06:00] Run the dialect fuzz test for 30 seconds
- [06:30] Show test output: iterations, coverage
- [07:00] Open filesystem/factory_fuzz_test.go
- [07:30] Show connection string fuzzing: malformed URLs, unusual protocols
- [08:00] Open internal/handlers/download_fuzz_test.go
- [08:30] Show path traversal fuzzing: null bytes, encoding tricks, ../ sequences
- [09:00] Open middleware/fuzz_test.go
- [09:30] Show input sanitization fuzzing
- [10:00] Show the saved corpus in middleware/testdata/fuzz/FuzzSanitizeInput/
- [10:30] Run a normal test (without -fuzz): show corpus entries being replayed
- [11:00] Explain resource limits: GOMAXPROCS=3, -p 2, -parallel 2
- [11:30] Show `cat /proc/loadavg` during fuzz testing -- verify within limits
- [12:00] Discuss -fuzztime options: 60s quick, 300s thorough, longer for CI
- [12:30] Explain crash-to-regression workflow: discover, save, commit, verify
- [13:00] Show how to add a new fuzz test: choose a target, write seeds, define invariants
- [13:30] Discuss what makes a good fuzz target: parsers, validators, serializers
- [14:00] Show edge cases the fuzzer commonly finds: buffer boundaries, unicode, encoding
- [14:30] Discuss integrating fuzz tests into the development workflow
- [15:00] Show running all fuzz corpus entries as part of normal test suite
- [15:30] Explain the value: catches inputs you never considered
- [16:00] Show all five fuzz test files side by side
- [16:30] Discuss coverage-guided fuzzing: Go fuzzer uses coverage to guide mutations
- [17:00] Recap fuzz testing: targets, seeds, mutations, crashes, regression

### Key Points

- Go native fuzzing since 1.18: testing.F with f.Add() seeds and f.Fuzz() targets
- Catalogizer fuzz targets: title parser, dialect rewriter, filesystem factory, download handler, input sanitizer
- Crash inputs saved to testdata/fuzz/ -- become permanent regression tests
- Resource limits: `GOMAXPROCS=3 go test -fuzz=FuzzXxx -p 2 -parallel 2`
- -fuzztime controls duration: 60s quick check, 300s thorough exploration
- Coverage-guided: Go fuzzer uses code coverage to guide input mutations
- Good fuzz targets: parsers, validators, serializers -- anything processing arbitrary input
- Common discoveries: buffer boundaries, unicode edge cases, encoding issues
- Saved corpus replayed in normal test runs (without -fuzz flag)
- All fuzz tests committed to version control as regression protection

### Tips

> **Tip**: Start with short fuzz runs (30-60 seconds) during development. Use longer runs (5+ minutes) before releases. The fuzzer finds most low-hanging bugs quickly, but deeper issues require more exploration time.

> **Tip**: The title parser is the highest-value fuzz target because it processes arbitrary filenames from the filesystem. Any filename that crashes the parser would crash the scanner. Fuzz it thoroughly.

### Quiz Questions

1. **Q**: What happens when the fuzzer discovers an input that crashes the code?
   **A**: The crash input is saved to testdata/fuzz/ and becomes a permanent regression test that is replayed in every subsequent test run.

2. **Q**: Why must fuzz tests be run with resource limits in Catalogizer?
   **A**: The host machine runs other mission-critical processes, and all workloads must be limited to 30-40% of total resources. Use GOMAXPROCS=3, -p 2, -parallel 2.

3. **Q**: What makes a good fuzz target?
   **A**: Functions that process arbitrary input: parsers, validators, serializers, and sanitizers. The title parser and dialect rewriter are ideal because they handle user-provided or filesystem-derived strings.

---

## Lesson 10.2: Property-Based Testing

**Duration**: 14 minutes

### Narration

Property-based testing is a complementary technique to both example-based unit tests and fuzz testing. Instead of testing specific input/output pairs, you define properties -- invariants that must hold for all valid inputs -- and let the testing framework generate random inputs to verify them.

The key question in property-based testing is: what must always be true about this function, regardless of input? For example, sorting a list must always produce a result with the same length as the input. Parsing and serializing a value must produce the original value. Searching with no filters must return all results.

Go's standard library includes testing/quick, a basic property testing package. For more sophisticated needs, the gopter library provides custom generators, shrinkers, and configurable test parameters. On the TypeScript side, fast-check provides similar capabilities for frontend testing.

Let me walk through properties that are valuable for Catalogizer.

The dialect rewriting idempotency property states: rewriting a query twice produces the same result as rewriting it once. If you take a SQLite query, rewrite it for PostgreSQL, and then rewrite the result again for PostgreSQL, the output should be identical to the first rewrite. This catches cases where the rewriter might double-transform a value -- for example, converting $1 back to ? and then to $1 again.

The pagination completeness property states: the union of page 1 and page 2 results covers all results without overlap when page size equals total divided by two. This verifies that pagination does not skip or duplicate items at page boundaries.

The search superset property states: searching with no filters returns a superset of results from searching with any filter applied. Adding a filter can only reduce results, never add new ones that were not in the unfiltered set.

The hierarchy consistency property states: every entity with a parent must have a parent that exists in the database. No orphaned children allowed. Additionally, the root of every hierarchy chain must be a type with no natural parent (tv_show, music_artist, movie, game, software, book, comic).

The boolean literal rewriting property states: for all known boolean columns, the rewriter converts 0 to FALSE and 1 to TRUE when targeting PostgreSQL. This must hold for every column name in the boolean column list, and must not affect non-boolean columns.

To implement these with testing/quick:

```go
func TestDialectRewriteIdempotent(t *testing.T) {
    f := func(query string) bool {
        first := RewritePlaceholders(query, DialectPostgres)
        second := RewritePlaceholders(first, DialectPostgres)
        return first == second
    }
    if err := quick.Check(f, nil); err != nil {
        t.Error(err)
    }
}
```

The quick.Check function generates random strings and verifies the property holds for each one. If it finds a counterexample, it reports the failing input. The default runs 100 iterations, configurable via quick.Config.

For the TypeScript frontend, fast-check works similarly:

```typescript
import fc from 'fast-check';

test('search filter reduces results', () => {
    fc.assert(fc.property(
        fc.string(),
        (query) => {
            const unfiltered = search({ q: query });
            const filtered = search({ q: query, type: 'movie' });
            return filtered.length <= unfiltered.length;
        }
    ));
});
```

Property-based tests complement unit tests by exploring the boundaries between test cases. A unit test checks that sorting [3, 1, 2] produces [1, 2, 3]. A property test verifies that sorting any list produces a sorted, same-length result. The unit test catches specific bugs; the property test catches categories of bugs.

The combination of fuzz testing and property-based testing is powerful. Fuzz testing finds crash-inducing inputs. Property testing finds logical invariant violations. Together, they cover both stability and correctness across a wide input space.

### On-Screen Actions

- [00:00] Show title: "Property-Based Testing"
- [00:30] Explain the concept: invariants that hold for all inputs
- [01:00] Compare: unit test checks specific input/output, property test checks invariants
- [01:30] Show Go's testing/quick package
- [02:00] Write the dialect rewriting idempotency test
- [02:30] Run the test -- show 100 random queries being verified
- [03:00] Show a failing case (if found) or explain how failures are reported
- [03:30] Show the pagination completeness property
- [04:00] Write the test: page 1 + page 2 covers all results without overlap
- [04:30] Run the test -- show pagination being verified across random data
- [05:00] Show the search superset property
- [05:30] Write the test: no-filter results superset of filtered results
- [06:00] Show the hierarchy consistency property
- [06:30] Write the test: all parents exist, all roots are valid types
- [07:00] Show the boolean literal rewriting property
- [07:30] Write the test: known boolean columns rewritten correctly
- [08:00] Show quick.Config for adjusting iteration count
- [08:30] Increase to 1000 iterations for more thorough testing
- [09:00] Show TypeScript fast-check for frontend property tests
- [09:30] Write a fast-check property test for search filtering
- [10:00] Run the frontend property test with vitest
- [10:30] Discuss shrinkers: when a property fails, the framework minimizes the counterexample
- [11:00] Show a shrunk counterexample -- minimal failing input
- [11:30] Discuss choosing properties: what must always be true?
- [12:00] Compare fuzz testing vs property testing: crashes vs invariant violations
- [12:30] Discuss combining both techniques for comprehensive coverage
- [13:00] Recap property-based testing: invariants, generators, counterexamples

### Key Points

- Property-based testing verifies invariants that must hold for all valid inputs
- Go: testing/quick (stdlib) and gopter for advanced needs; TypeScript: fast-check
- Key properties for Catalogizer: dialect rewrite idempotency, pagination completeness, search superset, hierarchy consistency, boolean rewriting
- Default: 100 random iterations per property, configurable for more thorough runs
- Shrinkers minimize counterexamples to the smallest failing input
- Complements unit tests (specific cases) and fuzz tests (crash discovery)
- Unit tests check input/output pairs; property tests check invariant categories
- Combination of fuzz + property testing covers stability and correctness

### Tips

> **Tip**: The hardest part of property-based testing is identifying good properties. Start by asking: what must always be true? What relationship must hold between input and output? What should never change regardless of input?

> **Tip**: Run property tests with higher iteration counts before releases. The default 100 iterations catches obvious violations, but 10,000 iterations explores the input space more thoroughly and can find subtle edge cases.

### Quiz Questions

1. **Q**: What is the difference between a unit test and a property test?
   **A**: A unit test checks a specific input/output pair. A property test defines an invariant that must hold for all valid inputs and verifies it with randomly generated inputs.

2. **Q**: What is the dialect rewriting idempotency property?
   **A**: Rewriting a SQL query twice for PostgreSQL produces the same result as rewriting it once. This ensures the rewriter does not double-transform values.

3. **Q**: What are shrinkers in property-based testing?
   **A**: When a property fails, shrinkers minimize the counterexample to the smallest input that still causes the failure, making it easier to diagnose the bug.

---

## Lesson 10.3: Stress Testing and Chaos Testing

**Duration**: 16 minutes

### Narration

Stress testing and chaos testing validate Catalogizer under extreme conditions. Stress testing pushes the system to its limits with high concurrency, large data volumes, and sustained load. Chaos testing deliberately injects failures to verify recovery mechanisms. Together, they ensure the system degrades gracefully under pressure and recovers automatically from failures.

Catalogizer has dedicated stress tests in the tests/stress/ directory. The concurrent API stress test in tests/stress/concurrent_api_stress_test.go sends many simultaneous requests to API endpoints, testing the Gin server's ability to handle concurrent connections. The database stress test in tests/stress/database_stress_test.go performs concurrent reads and writes against the database, testing WAL mode behavior under contention. The API load test in tests/stress/api_load_test.go simulates sustained request patterns over time.

For external load testing, k6 test scripts live in tests/k6/. The load test at tests/k6/load_test.js ramps up to 50 virtual users and verifies that the 95th percentile response time stays below 500 milliseconds. The stress test at tests/k6/stress_test.js ramps up to 300 virtual users to find the breaking point -- the concurrency level where response times degrade unacceptably. The soak test at tests/k6/soak_test.js runs 20 virtual users for 30 minutes to detect memory leaks and resource exhaustion over time.

k6 runs in a container to maintain isolation:

```bash
podman run --rm --network host \
  -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest \
  run /scripts/load_test.js
```

The stress test challenge in challenges/stress_test_challenge.go integrates stress testing into the challenge system. It can be triggered via the challenge API and reports results through the standard challenge infrastructure.

All stress tests must respect the 30-40% host resource limit. This is not just a guideline -- exceeding it can freeze the host machine. For Go stress tests, always use GOMAXPROCS=3 with -p 2 -parallel 2. For k6, configure the virtual user count to stay within the resource budget. For container-based tests, set --cpus and --memory limits.

Chaos testing validates the recovery mechanisms built into Catalogizer. The circuit breaker in internal/smb/ is the primary chaos testing target. When an SMB share becomes unreachable, the circuit breaker opens after repeated connection failures. During the open state, the offline cache serves previously cached file listings and metadata. When the share becomes available again, the circuit breaker transitions to half-open, tests the connection, and closes if successful.

To chaos test the circuit breaker: disconnect a NAS from the network (or block its IP), observe the circuit breaker opening, verify cached data is served, reconnect the NAS, and verify automatic recovery. The circuit breaker metrics (available via Prometheus) show state transitions, failure counts, and recovery times.

Database chaos testing validates SQLite WAL mode under adverse conditions. Kill database connections mid-query and verify the WAL journal allows recovery. Simulate disk full conditions and verify graceful error handling. The explicit PRAGMA journal_mode=WAL in database/connection.go ensures WAL mode is active even with the go-sqlcipher driver which ignores connection string pragmas.

WebSocket connection chaos verifies that the real-time event system handles mass disconnections and reconnections. Disconnect all clients simultaneously, reconnect them, and verify event delivery resumes without data loss. The sync.Once pattern in WebSocketHandler ensures safe shutdown even under concurrent close attempts.

Goroutine leak detection is critical for long-running processes. The CacheService and WebSocketHandler both spawn background goroutines. Tests must call defer service.Close() and handler.Stop() respectively. The Memory submodule provides leak detection utilities that track goroutine counts before and after test execution, flagging any that were not properly cleaned up.

### On-Screen Actions

- [00:00] Show title: "Stress Testing and Chaos Testing"
- [00:30] Open tests/stress/ directory -- show the three stress test files
- [01:00] Open concurrent_api_stress_test.go -- show concurrent request generation
- [01:30] Open database_stress_test.go -- show concurrent read/write patterns
- [02:00] Run a stress test with resource limits: `GOMAXPROCS=3 go test ./tests/stress/ -p 2 -parallel 2`
- [02:30] Show test output: concurrent operations, timing, error rates
- [03:00] Open tests/k6/ directory -- show the three k6 scripts
- [03:30] Open load_test.js -- show ramping to 50 users, p95 < 500ms threshold
- [04:00] Open stress_test.js -- show ramping to 300 users for breaking point
- [04:30] Open soak_test.js -- show 20 users for 30 minutes
- [05:00] Run k6 load test in a Podman container
- [05:30] Show k6 output: request rate, response times, error percentages
- [06:00] Show p95 and p99 latency metrics
- [06:30] Open challenges/stress_test_challenge.go -- show challenge integration
- [07:00] Show `podman stats --no-stream` -- verify resource usage within limits
- [07:30] Show `cat /proc/loadavg` during stress testing
- [08:00] Explain chaos testing: deliberate failure injection
- [08:30] Show the SMB circuit breaker in internal/smb/
- [09:00] Simulate NAS disconnection -- show circuit breaker opening
- [09:30] Show offline cache serving cached data during outage
- [10:00] Reconnect NAS -- show circuit breaker recovery: half-open, then closed
- [10:30] Show circuit breaker metrics in Prometheus
- [11:00] Show database chaos: WAL mode recovery after interrupted writes
- [11:30] Show the PRAGMA journal_mode=WAL in database/connection.go
- [12:00] Show WebSocket chaos: mass disconnect and reconnect
- [12:30] Show sync.Once pattern in WebSocketHandler for safe shutdown
- [13:00] Show goroutine leak detection: before/after goroutine counts
- [13:30] Show CacheService.Close() and WebSocketHandler.Stop() cleanup
- [14:00] Discuss the Recovery submodule: circuit breaker patterns
- [14:30] Show resource budget: max 4 CPUs, 8 GB RAM across all containers
- [15:00] Recap stress testing and chaos testing approaches

### Key Points

- Stress tests in tests/stress/: concurrent API, database contention, sustained load
- k6 scripts in tests/k6/: load (50 users, p95 < 500ms), stress (300 users), soak (30 min)
- k6 runs in Podman: `podman run --rm --network host -v tests/k6:/scripts docker.io/grafana/k6:latest`
- Resource limits mandatory: GOMAXPROCS=3, -p 2, -parallel 2 for Go; --cpus/--memory for containers
- Chaos testing: circuit breaker validation via NAS disconnection and recovery
- SMB circuit breaker: open on failure, offline cache serves data, auto-recovery on reconnection
- Database chaos: WAL mode ensures recovery after interrupted writes
- WebSocket chaos: mass disconnect/reconnect, sync.Once ensures safe shutdown
- Goroutine leak detection: track counts before/after, flag uncleaned goroutines
- All tests constrained to 30-40% of host resources to prevent system freeze

### Tips

> **Tip**: Run the soak test (30 minutes, 20 users) before any production deployment. Memory leaks and resource exhaustion often only manifest under sustained load over time, not during short burst tests.

> **Tip**: The circuit breaker is your primary resilience mechanism. Chaos test it regularly by simulating NAS outages. Verify not just that it opens, but that it recovers automatically and that cached data is served correctly during the outage.

### Quiz Questions

1. **Q**: What are the three k6 test scenarios and what do they test?
   **A**: Load test (50 users, verifies p95 < 500ms), stress test (300 users, finds breaking point), and soak test (20 users for 30 minutes, detects memory leaks).

2. **Q**: How does the SMB circuit breaker handle a NAS outage?
   **A**: It opens after repeated failures, serves cached data from the offline cache, transitions to half-open when connectivity returns, tests the connection, and closes if successful.

3. **Q**: Why is the 30-40% resource limit critical during stress testing?
   **A**: The host machine runs other mission-critical processes. Exceeding this limit can freeze the entire system. All tests must use GOMAXPROCS=3, -p 2, -parallel 2 or equivalent container limits.

---

## Lesson 10.4: Security Testing Patterns

**Duration**: 15 minutes

### Narration

Security testing in Catalogizer follows a zero-vulnerability policy enforced in builds. Every build must pass dependency scanning with zero known vulnerabilities, static analysis with zero security anti-patterns, and authorization boundary testing for every protected endpoint.

The primary security scanning tool for Go dependencies is govulncheck. It checks all Go module dependencies against the Go vulnerability database and reports any known vulnerabilities. Catalogizer maintains zero vulnerabilities -- any new vulnerability discovered in a dependency must be resolved before the next release.

For the frontend, npm audit checks all Node.js dependencies. The policy requires zero critical vulnerabilities in production dependencies. Development-only vulnerabilities are tracked but do not block builds.

Static analysis uses gosec for Go code. Gosec scans for common security anti-patterns: hardcoded credentials, SQL injection via string concatenation, unvalidated redirects, weak cryptographic usage, and more. The dialect abstraction in database/dialect.go prevents SQL injection by design -- all queries use parameterized placeholders that are rewritten per dialect, never string concatenation.

The security scanning script at scripts/security-scan.sh automates the complete scanning pipeline. It runs govulncheck, npm audit, and optionally Semgrep, Snyk, Trivy, and Gosec via the docker-compose.security.yml configuration.

Semgrep provides pattern-based static analysis. It can detect complex multi-file vulnerability patterns that simpler tools miss. The Semgrep configuration targets Go and TypeScript patterns specific to web applications: SSRF, XSS, CSRF, insecure deserialization, and more.

Snyk and Trivy provide comprehensive dependency and container image scanning. They detect vulnerabilities in both application dependencies and the base container images used for deployment. The docker-compose.security.yml file defines these tools as Compose services for easy execution.

Authorization boundary testing is where security testing intersects with the challenge system. The 49 API user flow challenges include tests for every protected endpoint. Each test verifies three boundaries: accessing without authentication returns 401, accessing with the wrong role returns 403, and accessing another user's resources returns 403.

The middleware layer enforces these boundaries. The JWT authentication middleware in middleware/auth.go validates tokens on every request. The rate limiter middleware applies strict limits (5 per minute) on login and registration endpoints to prevent brute force attacks, and default limits (100 per minute) on other endpoints.

XSS prevention relies on React's default output escaping in the frontend and Content-Security-Policy headers set by the Nginx reverse proxy. The input sanitization middleware in the backend strips potentially dangerous characters before they reach handlers.

For database security, SQLCipher provides encryption at rest for SQLite databases. The DB_ENCRYPTION_KEY environment variable controls the encryption key. For PostgreSQL, use the database's native encryption features and ensure connections use TLS.

The security test pipeline runs as part of the full test suite via scripts/run-all-tests.sh. It also runs independently via scripts/security-scan.sh for focused security review. Results are logged and any finding above the threshold blocks the build.

### On-Screen Actions

- [00:00] Show title: "Security Testing Patterns"
- [00:30] Explain the zero-vulnerability policy
- [01:00] Run `govulncheck ./...` in the catalog-api directory
- [01:30] Show output: 0 vulnerabilities found
- [02:00] Run `npm audit` in the catalog-web directory
- [02:30] Show output: 0 critical production vulnerabilities
- [03:00] Open scripts/security-scan.sh -- show the scanning pipeline
- [03:30] Run the security scan script
- [04:00] Show gosec scanning Go code for anti-patterns
- [04:30] Show a gosec finding example (or clean output)
- [05:00] Open database/dialect.go -- show parameterized queries preventing SQL injection
- [05:30] Show how RewritePlaceholders converts ? to $1, $2 -- no string concatenation
- [06:00] Open docker-compose.security.yml -- show Semgrep, Snyk, Trivy services
- [06:30] Run Semgrep scan via Podman
- [07:00] Show Semgrep results: pattern-based findings
- [07:30] Run Trivy container image scan
- [08:00] Show Trivy output: base image vulnerabilities
- [08:30] Discuss authorization boundary testing
- [09:00] Show a user flow challenge testing 401/403 boundaries
- [09:30] Show the JWT middleware in middleware/auth.go
- [10:00] Show rate limiter: 5/min on login, 100/min on other endpoints
- [10:30] Demonstrate: request without token returns 401
- [11:00] Demonstrate: request with wrong role returns 403
- [11:30] Show XSS prevention: React escaping, Content-Security-Policy headers
- [12:00] Show input sanitization middleware
- [12:30] Show SQLCipher encryption: DB_ENCRYPTION_KEY configuration
- [13:00] Show scripts/run-all-tests.sh including security scans
- [13:30] Discuss the security review workflow: scan, triage, remediate, verify
- [14:00] Recap security testing: dependency scanning, static analysis, boundary testing

### Key Points

- Zero-vulnerability policy: govulncheck 0 vulns, npm audit 0 critical production vulns
- Static analysis: gosec for Go anti-patterns, Semgrep for complex multi-file patterns
- Dependency scanning: Snyk and Trivy for application and container image vulnerabilities
- SQL injection prevented by design: dialect abstraction uses parameterized queries, never string concatenation
- Authorization boundaries: 49 API challenges test 401 (no auth), 403 (wrong role), 403 (wrong user)
- Rate limiting: 5/min on login/register (brute force prevention), 100/min default
- XSS prevention: React default escaping + Content-Security-Policy headers + input sanitization
- Database encryption: SQLCipher with DB_ENCRYPTION_KEY for SQLite at-rest encryption
- Automated pipeline: scripts/security-scan.sh runs all tools; findings above threshold block builds
- Security tools: docker-compose.security.yml defines Semgrep, Snyk, Trivy as Compose services

### Tips

> **Tip**: Run govulncheck and npm audit after every dependency update. A single vulnerable transitive dependency can compromise the entire application. Automate this in your development workflow.

> **Tip**: The dialect abstraction is your strongest SQL injection defense. Never bypass it by writing raw SQL with string concatenation. If you need a query the abstraction does not support, extend the abstraction rather than working around it.

### Quiz Questions

1. **Q**: How does the dialect abstraction prevent SQL injection?
   **A**: It uses parameterized placeholders (? for SQLite, $1/$2 for PostgreSQL) that are rewritten per dialect. Queries never use string concatenation for user input.

2. **Q**: What are the three authorization boundaries tested by the API challenges?
   **A**: No authentication (expects 401), wrong role (expects 403), and accessing another user's resources (expects 403).

3. **Q**: What rate limits are applied to login endpoints?
   **A**: Strict rate limiting of 5 requests per minute on login and registration endpoints to prevent brute force attacks. Other endpoints use the default 100 per minute.

---

## Lesson 10.5: Visual Regression Testing

**Duration**: 12 minutes

### Narration

Visual regression testing catches changes to the user interface that functional tests miss. A button that still works but moved 50 pixels to the left, a color that changed from the design spec, a layout that breaks on a specific screen size -- these are all visual regressions that pass functional tests but degrade the user experience.

Catalogizer uses Playwright for visual regression testing. Playwright captures screenshots at key interaction points and compares them against baseline images. The test environment uses docker-compose.test.yml with network_mode: host for all services: catalog-api, catalog-web, and the Playwright runner.

The 59 web user flow challenges include visual verification steps. Each challenge navigates to a specific page or state, performs interactions, and captures screenshots. These screenshots serve dual purposes: they verify visual consistency when compared to baselines, and they provide documentation of the application state at each step.

The screenshot comparison process works as follows. The first time a test runs, it captures a baseline screenshot and stores it in version control. On subsequent runs, the test captures a new screenshot and performs pixel-level comparison against the baseline. If the difference exceeds a configurable threshold, the test fails and produces a diff image highlighting the changed pixels.

To set up visual regression testing, you need the test container stack running:

```bash
podman-compose -f docker-compose.test.yml up
```

This starts the API server, web frontend, and Playwright runner, all with network_mode: host so they can communicate on localhost.

The key pages to capture for visual regression are: the login page (brand identity, form layout), the dashboard (stats panels, charts, navigation), the media browser (grid view, list view, filters), the entity detail page (metadata layout, poster placement), the collections page (card layout, action buttons), and the admin panel (user table, configuration forms).

Baseline screenshots must be regenerated when intentional UI changes are made. If you redesign the dashboard, the old baseline is no longer valid. The workflow is: make the UI change, verify it visually, regenerate baselines, commit the new baselines to version control.

The zero console error policy extends to visual testing. During every screenshot capture, the test also checks the browser console for errors and warnings. Any console error or failed network request is a defect. Every API endpoint the frontend calls must exist, return valid 2xx responses, and match the expected response shape.

Screen size testing captures screenshots at multiple viewport sizes: desktop (1920x1080), tablet (768x1024), and mobile (375x812). This verifies responsive design works correctly and catches layout regressions specific to certain screen sizes.

The resource budget applies to visual testing as well. The test containers are limited to max 4 CPUs and 8 GB RAM total. Playwright runs one browser instance at a time to stay within limits. Screenshots are captured sequentially, not in parallel.

### On-Screen Actions

- [00:00] Show title: "Visual Regression Testing"
- [00:30] Explain the concept: catching visual changes that functional tests miss
- [01:00] Show examples: button moved, color changed, layout broken
- [01:30] Open docker-compose.test.yml -- show the test stack with network_mode: host
- [02:00] Start the test stack: `podman-compose -f docker-compose.test.yml up`
- [02:30] Show three containers running: API, web, Playwright
- [03:00] Show a web user flow challenge with screenshot capture
- [03:30] Show the screenshot capture code in Playwright
- [04:00] Show baseline screenshots stored in version control
- [04:30] Show pixel-level comparison: baseline vs new capture
- [05:00] Show a diff image with changed pixels highlighted
- [05:30] Show the configurable threshold: how much difference triggers failure
- [06:00] Capture screenshots at key pages: login, dashboard, media browser
- [06:30] Show entity detail page screenshot with metadata layout
- [07:00] Show collections page and admin panel screenshots
- [07:30] Demonstrate responsive screenshots: desktop, tablet, mobile viewports
- [08:00] Show the zero console error policy: checking console during captures
- [08:30] Open browser DevTools -- show zero errors during screenshot sequence
- [09:00] Demonstrate baseline regeneration workflow
- [09:30] Make an intentional UI change -- show baseline mismatch
- [10:00] Regenerate baseline -- commit new screenshots
- [10:30] Show resource limits: max 4 CPUs, 8 GB RAM for test containers
- [11:00] Recap visual regression testing: baselines, comparison, responsive, console errors

### Key Points

- Playwright captures screenshots at key interaction points for pixel-level comparison
- Test stack: docker-compose.test.yml with network_mode: host (API, web, Playwright)
- 59 web user flow challenges include visual verification steps
- Baseline screenshots stored in version control; differences trigger test failure
- Configurable threshold for pixel-level comparison sensitivity
- Key pages: login, dashboard, media browser, entity detail, collections, admin panel
- Responsive testing: desktop (1920x1080), tablet (768x1024), mobile (375x812)
- Zero console error policy: any browser console error or failed network request is a defect
- Baseline regeneration required after intentional UI changes
- Resource budget: max 4 CPUs, 8 GB RAM across all test containers, sequential captures

### Tips

> **Tip**: Keep your baseline screenshots up to date. Outdated baselines create false positives that erode trust in the visual regression suite. When you make an intentional UI change, update baselines as part of the same commit.

> **Tip**: The zero console error policy is one of the most effective quality gates. A single failed network request or unhandled error means either a frontend bug or a missing backend endpoint. Fix it immediately -- these issues compound quickly.

### Quiz Questions

1. **Q**: What does visual regression testing catch that functional tests do not?
   **A**: Layout changes, color differences, element positioning shifts, and responsive design regressions -- changes that do not affect functionality but degrade the visual user experience.

2. **Q**: When must baseline screenshots be regenerated?
   **A**: After intentional UI changes. The old baseline is no longer valid, so the new screenshots must be captured, verified visually, and committed to version control.

3. **Q**: What does the zero console error policy enforce during visual testing?
   **A**: Every screenshot capture also checks the browser console. Any console error, warning, or failed network request is treated as a defect that must be fixed.

---

## Lesson 10.6: Challenge System for End-to-End Verification

**Duration**: 15 minutes

### Narration

The challenge system ties all testing techniques together into a comprehensive end-to-end verification framework. Catalogizer has 239 registered challenges organized into three groups: 50 original challenges covering core functionality, 174 user flow challenges covering multi-platform end-to-end scenarios, and 15 module verification challenges covering the submodule architecture.

The challenge framework is implemented in the Challenges submodule at Challenges/. Each challenge is a Go struct that embeds challenge.BaseChallenge and implements an Execute() method. Challenges are registered in catalog-api/challenges/register.go via RegisterAll(), and exposed through REST endpoints at /api/v1/challenges.

The original 50 challenges (CH-001 through CH-050) verify core system behavior: storage root creation, file scanning, media detection, metadata enrichment, search, browse, collections, favorites, playlists, subtitles, user management, security, and monitoring. These challenges run against a live Catalogizer instance and validate real behavior, not mocked interfaces.

The 174 user flow challenges cover four platform groups. The 49 API challenges in challenges/userflow_api.go verify every REST endpoint via HTTP. The 59 web challenges in challenges/userflow_web.go verify the React frontend via Playwright. The 28 desktop challenges in challenges/userflow_desktop.go verify the Tauri desktop app and installer wizard. The 38 mobile challenges in challenges/userflow_mobile.go verify the Android and Android TV apps.

The 15 module verification challenges (MOD-001 through MOD-015) verify that each independent submodule works correctly when integrated into the main application.

The user flow framework in Challenges/pkg/userflow/ provides a generic automation layer with adapter interfaces for each platform: Browser (Playwright), Mobile (ADB), Desktop (Tauri), API (HTTP), Build (Gradle, npm, Go, Cargo), and Process management. Challenge templates standardize common patterns like environment setup, build verification, API health checks, browser flows, and mobile launches.

Running challenges is straightforward. The challenge API endpoint GET /api/v1/challenges lists all registered challenges. POST /api/v1/challenges/:id/run executes a single challenge. POST /api/v1/challenges/run-all executes all challenges sequentially. Results are returned with pass/fail status, execution time, assertion details, and any error messages.

There are critical constraints for challenge execution. RunAll is synchronous and blocking -- no other challenge can run until it finishes. The progress-based liveness detection has a 5-minute stale threshold: if a challenge stops reporting progress for 5 minutes, it is killed as stuck. Challenge configurations default to a 5-minute timeout, but this can be zeroed to use the runner's timeout instead. The config.json write_timeout must be set to 900 seconds (not the default 30) for long-running challenge suites.

All challenge operations must be executed by the system's compiled binaries -- the catalog-api service and other Catalogizer applications. Never use custom scripts, curl commands, or third-party tools to trigger API endpoints within challenge execution. This ensures challenges test the system exactly as an end user would use it.

The challenge runner CLI at Challenges/cmd/userflow-runner/ supports flags for platform selection, report format, compose file, project root, timeout, and output directory. It can run challenges against the containerized test stack defined in docker-compose.test.yml.

Challenges are run sequentially, never in parallel, to avoid resource contention and ensure deterministic results. Monitor resource usage with podman stats --no-stream and cat /proc/loadavg during challenge execution.

### On-Screen Actions

- [00:00] Show title: "Challenge System for End-to-End Verification"
- [00:30] Show the challenge breakdown: 50 original + 174 user flow + 15 module = 239 total
- [01:00] Open catalog-api/challenges/register.go -- show RegisterAll()
- [01:30] Show challenge registration: original, user flow, module verification
- [02:00] Open a sample challenge -- show BaseChallenge embedding and Execute()
- [02:30] Show the Challenges/ submodule directory structure
- [03:00] Open challenges/userflow_api.go -- show API challenges
- [03:30] Show the 49 REST endpoint verification challenges
- [04:00] Open challenges/userflow_web.go -- show web challenges
- [04:30] Show the 59 Playwright-based browser challenges
- [05:00] Open challenges/userflow_desktop.go and userflow_mobile.go
- [05:30] Show desktop (28) and mobile (38) challenges
- [06:00] Open Challenges/pkg/userflow/ -- show the generic automation framework
- [06:30] Show adapter interfaces: Browser, Mobile, Desktop, API
- [07:00] Show GET /api/v1/challenges -- list all 239 challenges
- [07:30] Run a single challenge: POST /challenges/:id/run
- [08:00] Show challenge result: pass/fail, execution time, assertions
- [08:30] Explain RunAll constraints: synchronous, blocking, 5-min stale threshold
- [09:00] Show config.json write_timeout=900 requirement
- [09:30] Run a subset of challenges via the API
- [10:00] Show challenge progress reporting: updates every 5 seconds
- [10:30] Open Challenges/cmd/userflow-runner/ -- show the CLI runner
- [11:00] Show CLI flags: --platform, --report, --compose, --root, --timeout
- [11:30] Run the CLI runner against docker-compose.test.yml
- [12:00] Show challenge results: pass counts, failure details
- [12:30] Show `podman stats --no-stream` during challenge execution
- [13:00] Discuss the end-to-end verification philosophy: test as the user would
- [13:30] Show how challenges combine: fuzz findings become challenge assertions
- [14:00] Recap the challenge system: 239 challenges, 4 platforms, automated verification

### Key Points

- 239 registered challenges: 50 original (CH-001 to CH-050) + 174 user flow (UF-*) + 15 module (MOD-*)
- User flow platforms: 49 API (HTTP), 59 web (Playwright), 28 desktop (Tauri), 38 mobile (ADB)
- Framework: Challenges/ submodule with generic adapters for Browser, Mobile, Desktop, API, Build
- Challenge API: GET /challenges (list), POST /challenges/:id/run (single), POST /challenges/run-all (all)
- RunAll is synchronous/blocking: no other challenge can run until it completes
- 5-minute stale threshold: stuck challenges are killed if no progress reported
- config.json write_timeout must be 900 (not 30) for long-running challenge suites
- All operations must use compiled system binaries, never custom scripts or curl
- CLI runner: Challenges/cmd/userflow-runner/ with platform, report, compose, timeout flags
- Sequential execution only: never parallel, monitor resources during runs

### Tips

> **Tip**: Run the original 50 challenges (CH-001 to CH-050) first as a quick validation. They cover core functionality and run faster than the full 239-challenge suite. If the core challenges pass, proceed to the full suite.

> **Tip**: When a challenge fails, check the assertion details and error messages carefully. Challenges test real system behavior, so a failure indicates a genuine issue that would affect users. The 5-minute stale threshold helps identify challenges that hang, typically due to connectivity or resource issues.

### Quiz Questions

1. **Q**: How many total challenges does Catalogizer have and how are they organized?
   **A**: 239 total: 50 original (CH-001 to CH-050) for core functionality, 174 user flow for multi-platform end-to-end scenarios across 4 platforms, and 15 module verification for submodule integration.

2. **Q**: Why must challenge operations use compiled system binaries instead of custom scripts?
   **A**: To ensure challenges test the system exactly as an end user would use it. Custom scripts or curl commands would bypass the application's normal request processing pipeline.

3. **Q**: What happens if a challenge stops reporting progress for 5 minutes?
   **A**: The progress-based liveness detection kills it as stuck, with a status of "stuck" distinct from "timed_out". This prevents hung challenges from blocking the entire suite.

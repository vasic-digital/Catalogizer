# Module 10: Advanced Testing - Slide Outlines

---

## Slide 10.0.1: Title Slide

**Title**: Advanced Testing

**Subtitle**: Fuzz Testing, Property-Based Testing, Stress Testing, Security Scanning, Visual Regression, and Challenge Verification

**Speaker Notes**: This module covers advanced testing techniques beyond standard unit and integration tests. By the end, students will be able to apply fuzz testing, chaos testing, security scanning (SonarQube, Semgrep, Snyk, Trivy), k6 load testing, visual regression testing, and challenge-based end-to-end verification to Go and TypeScript projects.

---

## Slide 10.1.1: Fuzz Testing with Go

**Title**: Go 1.18+ Native Fuzzing

**Bullet Points**:
- Built-in `testing.F` type for fuzz targets since Go 1.18
- Fuzz functions: `func FuzzXxx(f *testing.F)`
- Seed corpus: `f.Add(seedValue)` provides initial inputs
- Fuzzer mutates seeds to discover edge cases and crashes
- Run: `go test -fuzz=FuzzXxx -fuzztime=60s ./path/to/pkg/`
- Crash inputs saved to `testdata/fuzz/` for regression testing

**Speaker Notes**: Fuzz targets are valuable for the title parser, MIME type detector, and SQL dialect rewriter. Each crash discovered by the fuzzer becomes a permanent regression test via the saved corpus.

---

## Slide 10.1.2: Fuzz Targets in Catalogizer

**Title**: What to Fuzz

**Bullet Points**:
- Title parser: ensure no panics on arbitrary filenames
- MIME detector: verify graceful handling of corrupt file headers
- Dialect rewriter: confirm SQL rewriting never produces invalid SQL
- Path sanitizer: verify no path traversal on crafted inputs
- Seed with happy path, empty strings, long strings, and unicode
- Resource limits: `GOMAXPROCS=3 go test -fuzz=... -p 2 -parallel 2`

**Speaker Notes**: The fuzzer discovers inputs you never considered -- buffer boundaries, unicode edge cases, and encoding issues. Always run fuzz tests with the project resource limits (30-40% of host resources).

---

## Slide 10.2.1: Property-Based Testing

**Title**: Testing Invariants, Not Examples

**Bullet Points**:
- Property: a statement that must hold for all valid inputs
- Libraries: `testing/quick` (stdlib), `gopter` (Go), `fast-check` (TypeScript)
- Properties for Catalogizer:
  - Pagination: page 1 + page 2 covers all results without overlap
  - Dialect rewriting is idempotent (rewriting twice equals rewriting once)
  - Searching with no filters returns all results

**Speaker Notes**: Property-based testing complements example-based tests. Instead of checking specific input/output pairs, you verify invariants across all inputs. The idempotency property for dialect rewriting is especially powerful.

---

## Slide 10.3.1: Stress and Chaos Testing

**Title**: Testing Under Extreme Conditions

**Bullet Points**:
- **Stress testing**: concurrent scanner sessions, rapid search queries, many WebSocket connections
- **Chaos testing**: kill database connections mid-query, simulate NAS network partitions, corrupt cache entries
- SMB circuit breaker designed for chaos: auto-opens, serves cached data, recovers when NAS returns
- SQLite WAL mode: explicit `PRAGMA journal_mode=WAL` in `database/connection.go`
- Goroutine leak detection: track counts before/after test execution via Memory submodule
- All tests limited to 30-40% host resources

**Speaker Notes**: Stress testing reveals performance cliffs and resource leaks. Chaos testing validates recovery paths. The circuit breaker is validated by injecting NAS failures and verifying automatic recovery. WebSocket chaos testing verifies mass disconnect/reconnect with sync.Once ensuring safe shutdown. CacheService and WebSocketHandler must be properly closed in tests.

---

## Slide 10.3.2: k6 Load Testing Scenarios

**Title**: External Load Testing with k6

**Bullet Points**:
- **Load test** (`tests/k6/load_test.js`): ramp to 50 users, verify p95 < 500ms
- **Stress test** (`tests/k6/stress_test.js`): ramp to 300 users, find breaking point
- **Soak test** (`tests/k6/soak_test.js`): 20 users for 30 minutes, detect memory leaks
- **Spike test** (`tests/k6/spike_test.js`): sudden burst to high concurrency, verify recovery
- Run in Podman: `podman run --rm --network host -v tests/k6:/scripts docker.io/grafana/k6:latest run /scripts/load_test.js`
- Stress test challenge integrates k6 results into the challenge system

**Speaker Notes**: The spike test simulates sudden traffic bursts -- for example, many users launching the app simultaneously. It verifies the system recovers gracefully after the spike subsides, with response times returning to normal. All k6 scenarios respect the container resource budget: max 4 CPUs, 8 GB RAM.

---

## Slide 10.4.1: Security Testing Patterns

**Title**: Automated Security Verification

**Bullet Points**:
- **Static analysis**: `gosec` for Go security anti-patterns
- **Dependency scanning**: `govulncheck` (0 vulns), `npm audit` (0 critical)
- **SQL injection**: parameterized queries enforced by dialect abstraction
- **XSS prevention**: React default escaping + Content-Security-Policy headers
- **Auth boundaries**: test every endpoint without auth (401), wrong role (403), other user (403)
- Rate limiting: 5/min on login/register (brute force prevention), 100/min default
- Zero-vulnerability policy enforced in builds

**Speaker Notes**: The dialect abstraction prevents SQL injection by design. Manual security review focuses on authorization logic and access control boundaries. The 49 API user flow challenges include auth boundary tests for all protected endpoints. Database encryption via SQLCipher with DB_ENCRYPTION_KEY for SQLite at-rest protection.

---

## Slide 10.4.2: SonarQube Code Quality Scanning

**Title**: SonarQube for Code Quality and Security

**Bullet Points**:
- SonarQube scanner analyzes Go and TypeScript code for bugs, vulnerabilities, and code smells
- Runs via `scripts/sonarqube-scan.sh` using containerized SonarQube
- Detects: cognitive complexity, duplicated blocks, security hotspots, reliability issues
- Quality gates enforce minimum standards before release
- Results viewable in SonarQube web dashboard
- Complements gosec with broader code quality analysis beyond security

**Speaker Notes**: SonarQube provides a broader view than security-focused tools. It catches maintainability issues, code duplication, and complexity hotspots that accumulate over time. The scanner runs in a Podman container to maintain the containerized build policy.

---

## Slide 10.4.3: Semgrep SAST Analysis

**Title**: Semgrep Static Application Security Testing

**Bullet Points**:
- Pattern-based static analysis for Go and TypeScript
- Detects complex multi-file vulnerability patterns that simpler tools miss
- Targets: SSRF, XSS, CSRF, insecure deserialization, injection flaws
- Configured via `docker-compose.security.yml` as a Compose service
- Also available: Snyk (dependency + container scanning) and Trivy (image vulnerabilities)
- Full pipeline: `scripts/security-scan.sh` runs all tools; findings above threshold block builds

**Speaker Notes**: Semgrep excels at detecting patterns across multiple files -- for example, a user input flowing through several functions to reach a dangerous API. Combined with Snyk for dependency scanning and Trivy for container image scanning, the security pipeline covers code, dependencies, and infrastructure.

---

## Slide 10.5.1: Visual Regression Testing

**Title**: Catching UI Changes with Screenshots

**Bullet Points**:
- Playwright captures screenshots at key interaction points
- Baseline images stored in version control
- Pixel-level comparison with configurable threshold
- Test environment: `docker-compose.test.yml` with `network_mode: host`
- 59 web user flow challenges include visual verification steps
- Zero console error policy: every failed network request is a defect

**Speaker Notes**: Visual regression catches CSS changes, layout shifts, and rendering bugs that functional tests miss. Baseline screenshots must be regenerated when intentional UI changes are made. The resource budget applies: max 4 CPUs, 8 GB RAM across test containers.

---

## Slide 10.6.1: Challenge System for End-to-End Verification

**Title**: 239 Challenges Across 4 Platforms

**Bullet Points**:
- **50 original** (CH-001 to CH-050): core functionality -- scanning, detection, enrichment, search, browse, collections
- **174 user flow**: 49 API (HTTP) + 59 web (Playwright) + 28 desktop (Tauri) + 38 mobile (ADB)
- **15 module verification** (MOD-001 to MOD-015): submodule integration checks
- Challenge API: `GET /challenges` (list), `POST /challenges/:id/run` (single), `POST /challenges/run-all` (all)
- RunAll is synchronous/blocking; 5-minute stale threshold kills stuck challenges
- CLI runner: `Challenges/cmd/userflow-runner/` with `--platform`, `--report`, `--timeout` flags

**Speaker Notes**: The challenge system ties all testing techniques together. Challenges run against live Catalogizer instances and validate real behavior, not mocked interfaces. All operations must use compiled system binaries -- never custom scripts or curl. Sequential execution only, with config.json write_timeout set to 900 for long suites.

---

## Slide 10.6.2: Challenge Framework Architecture

**Title**: Generic Automation with Platform Adapters

**Bullet Points**:
- Framework in `Challenges/pkg/userflow/` -- zero project-specific references
- 6 adapter interfaces: Browser (Playwright), Mobile (ADB), Desktop (Tauri), API (HTTP), Build, Process
- 13 challenge templates: Env setup/teardown, Build, UnitTest, Lint, APIHealth, APIFlow, BrowserFlow, MobileLaunch, etc.
- 12 evaluators: build_succeeds, all_tests_pass, status_code, response_contains, flow_completes, within_duration
- Test stack: `docker-compose.test.yml` with `network_mode: host` (API, web, Playwright)
- Registered in `catalog-api/challenges/register.go` via `RegisterAll()`

**Speaker Notes**: The user flow framework is generic and reusable beyond Catalogizer. Adapter interfaces abstract platform differences while challenge templates standardize common verification patterns. The test stack runs all services on host networking for reliable inter-service communication.

---

## Slide 10.7.1: Module 10 Summary

**Title**: What We Covered

**Bullet Points**:
- Fuzz testing: Go `testing.F` for parser, detector, and rewriter verification
- Property-based testing: invariant verification with `testing/quick`
- Stress and chaos testing: concurrent load, failure injection, circuit breaker validation
- k6 load testing: load, stress, soak, and spike test scenarios in Podman containers
- Security scanning: SonarQube (code quality), Semgrep (SAST), Snyk/Trivy (dependencies/images), gosec (Go)
- Visual regression: Playwright screenshots, pixel comparison, recorded challenges
- Challenge system: 239 challenges across 4 platforms for end-to-end verification
- Resource limits: all tests constrained to 30-40% of host resources

**Speaker Notes**: These techniques contribute to the 239 registered challenges that validate Catalogizer end to end. Fuzz testing finds unexpected inputs. Property testing verifies invariants. Stress, chaos, and k6 testing validate resilience. SonarQube, Semgrep, Snyk, and Trivy enforce zero-vulnerability policy. Visual regression catches UI drift. The challenge system ties everything together into automated, repeatable verification.

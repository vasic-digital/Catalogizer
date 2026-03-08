# Module 10: Advanced Testing - Slide Outlines

---

## Slide 10.0.1: Title Slide

**Title**: Advanced Testing

**Subtitle**: Fuzz Testing, Property-Based Testing, Stress Testing, Security Testing, and Visual Regression

**Speaker Notes**: This module covers advanced testing techniques beyond standard unit and integration tests. By the end, students will be able to apply fuzz testing, chaos testing, security testing, and visual regression testing to Go and TypeScript projects.

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
- All tests limited to 30-40% host resources

**Speaker Notes**: Stress testing reveals performance cliffs and resource leaks. Chaos testing validates recovery paths. The circuit breaker is validated by injecting NAS failures and verifying automatic recovery.

---

## Slide 10.4.1: Security Testing Patterns

**Title**: Automated Security Verification

**Bullet Points**:
- **Static analysis**: `gosec` for Go security anti-patterns
- **Dependency scanning**: `govulncheck` (0 vulns), `npm audit` (0 critical)
- **SQL injection**: parameterized queries enforced by dialect abstraction
- **XSS prevention**: React default escaping + Content-Security-Policy headers
- **Auth boundaries**: test every endpoint without auth (401), wrong role (403), other user (403)
- Zero-vulnerability policy enforced in builds

**Speaker Notes**: The dialect abstraction prevents SQL injection by design. Manual security review focuses on authorization logic and access control boundaries. The 49 API user flow challenges include auth boundary tests for all protected endpoints.

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

## Slide 10.6.1: Module 10 Summary

**Title**: What We Covered

**Bullet Points**:
- Fuzz testing: Go `testing.F` for parser, detector, and rewriter verification
- Property-based testing: invariant verification with `testing/quick`
- Stress and chaos testing: concurrent load, failure injection, circuit breaker validation
- Security testing: static analysis, dependency scanning, auth boundary testing
- Visual regression: Playwright screenshots, pixel comparison, recorded challenges
- Resource limits: all tests constrained to 30-40% of host resources

**Speaker Notes**: These techniques contribute to the 239 registered challenges that validate Catalogizer end to end. Fuzz testing finds unexpected inputs. Property testing verifies invariants. Stress and chaos testing validate resilience. Security testing enforces zero-vulnerability policy. Visual regression catches UI drift.

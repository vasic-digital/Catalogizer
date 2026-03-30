# Module 7: Testing and Quality Assurance - Slide Deck Outline

**Total Slides**: 12
**Estimated Duration**: 60 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Testing and Quality Assurance

- Test strategy, writing tests, running suites, coverage, security scanning
- Prerequisites: Module 6 completed (Lessons 6.1-6.2 minimum)
- By the end: write, run, and interpret tests at every level of the stack

---

## Slide 2: Testing Philosophy (4 min)

**Title**: Zero Warning / Zero Error Policy

- All components must run with zero console warnings and zero errors
- Every failed network request is a defect
- Stub endpoints return valid empty responses for unimplemented features
- Challenge suite (CH-001 to CH-050+) enforces this end-to-end
- Resource limits: GOMAXPROCS=3, -p 2, -parallel 2

---

## Slide 3: Go Backend Testing (5 min)

**Title**: Table-Driven Tests With In-Memory SQLite

- *_test.go files placed beside their source files
- Table-driven pattern: slice of structs with t.Run loop
- Test helper: database.WrapDB(sqlDB, DialectSQLite) for in-memory database
- Three layers tested: handlers, services, repositories
- Run single test: go test -v -run TestName ./path/to/pkg/
- Exercise reference: Exercise 7.1 -- write a table-driven service test

---

## Slide 4: Handler and Middleware Testing (5 min)

**Title**: Testing HTTP Handlers With Gin Test Context

- httptest.NewRecorder for capturing responses
- gin.CreateTestContext for isolated handler testing
- Mock service injection via constructor
- Middleware tests: auth verification, rate limiting, metrics
- Demo: walk through auth_handler_test.go

---

## Slide 5: Frontend Unit Testing (5 min)

**Title**: Vitest With React Testing Library

- Vitest with jsdom environment for DOM simulation
- Setup file: src/test-setup.ts
- Test files use .test.ts/.test.tsx extension
- 101 test files, 1623 tests across all frontend packages
- npm run test (single run), npm run test:watch, npm run test:coverage
- Exercise reference: Exercise 7.2 -- write a component test with Vitest

---

## Slide 6: E2E Testing With Playwright (5 min)

**Title**: End-to-End Tests for User Flows

- Playwright spec files in catalog-web/e2e/tests/
- Fixtures: auth.ts (mock auth), api-mocks.ts (mock API responses)
- Pattern: mockAuthEndpoints -> loginAs -> page.goto -> assertions
- npm run test:e2e (headless), npm run test:e2e:headed (visual)
- Demo: run auth.spec.ts and observe browser automation

---

## Slide 7: API Client and Installer Tests (4 min)

**Title**: Testing TypeScript Libraries

- catalogizer-api-client: npm run build && npm run test
- installer-wizard: 19 test files, 178 tests
- Vitest for both libraries with the same patterns as catalog-web
- Type checking: npm run type-check (tsc --noEmit)
- Linting: npm run lint with --max-warnings 0

---

## Slide 8: Android Testing (5 min)

**Title**: Unit Tests and Instrumented Tests

- ./gradlew test for unit tests (JUnit, Mockito)
- ./gradlew connectedAndroidTest for instrumented tests
- MVVM testing: test ViewModels with fake repositories
- Room database tests with in-memory instances
- Requires jvmToolchain(17) and --add-opens JVM args

---

## Slide 9: Security Scanning (5 min)

**Title**: Finding Vulnerabilities Before They Ship

- govulncheck ./... for Go dependency vulnerabilities
- npm audit --production for frontend dependency issues
- Semgrep SAST via docker-compose.security.yml
- SonarQube: ./scripts/run-sonarqube-scan.sh
- Snyk and Trivy for container image scanning
- Exercise reference: Exercise 7.3 -- run govulncheck and npm audit

---

## Slide 10: Performance Testing (5 min)

**Title**: Load, Stress, and Soak Tests With k6

- tests/k6/load_test.js: ramp to 50 users, verify p95 < 500ms
- tests/k6/stress_test.js: ramp to 300 users, find breaking point
- tests/k6/soak_test.js: 20 users for 30 minutes, detect memory leaks
- Run via Podman: podman run --rm --network host grafana/k6
- scripts/memory-leak-check.sh for automated memory profiling

---

## Slide 11: Coverage and Quality Gates (4 min)

**Title**: Measuring and Maintaining Coverage

- npm run test:coverage generates coverage via @vitest/coverage-v8
- scripts/validate-coverage.sh enforces minimum thresholds
- scripts/run-all-tests.sh runs all tests across all components
- Challenge suite provides end-to-end validation
- Exercise reference: Exercise 7.4 -- achieve 80%+ coverage on a module

---

## Slide 12: Module Summary and Next Steps (3 min)

**Title**: What We Covered

- Zero warning/zero error policy enforced by challenges
- Go table-driven tests with in-memory SQLite
- Frontend unit tests with Vitest, E2E with Playwright
- Security scanning: govulncheck, npm audit, Semgrep, SonarQube
- Performance testing with k6 load/stress/soak tests
- Next module: Deployment and Production

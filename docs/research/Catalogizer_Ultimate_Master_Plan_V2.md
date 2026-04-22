# Catalogizer: Ultimate Comprehensive Implementation & Fix Plan

## TL;DR

This document provides a complete, line-by-line plan to bring the Catalogizer multi-platform media collection management system to **100% working status across all 41 submodules and all client applications**. The plan addresses the universal "green tests, broken product" phenomenon by implementing a multi-layered verification system with institutional knowledge documentation, self-correcting prompt chains, LLM-as-Judge gates, and per-submodule fix strategies. Every submodule — from HelixQA to LLMsVerifier, from catalog-api to catalogizer-androidtv — is covered with specific actions, verification commands, and success criteria.

---

## 1. Project Overview & Current State Analysis

### 1.1 What is Catalogizer?

Catalogizer is an advanced multi-protocol media collection management system that automatically detects, categorizes, and organizes media files across SMB, FTP, NFS, WebDAV, and local filesystem protocols. The system consists of a Go-based REST API backend (catalog-api), a React TypeScript web frontend (catalog-web), Android mobile and TV applications (catalogizer-android, catalogizer-androidtv), a desktop application via Tauri (catalogizer-desktop), an installation wizard (installer-wizard), and a cross-platform API client library (catalogizer-api-client). The entire ecosystem is supported by 41 independent git submodules providing reusable functionality under the `digital.vasic.*` Go module convention and `@vasic-digital/*` TypeScript package namespace.

The project has been developed with extensive AI assistance, resulting in high unit test coverage (approaching 100% in many modules) but persistent issues where the actual product fails during manual testing. This document addresses that gap comprehensively.

### 1.2 Current Repository State

The main repository at `git@github.com:vasic-digital/Catalogizer.git` contains 1,040+ commits and is organized as a monorepo with git submodules. The codebase spans **six programming languages**: Go (59.3%), TypeScript (21.2%), Kotlin (10.7%), Shell (6.1%), Rust (1.1%), and JavaScript (0.9%). The project has 7 open issues tracking phased implementation work, with Issues #2, #7, and #3 partially completed and Issues #4, #5, #6, and #8 representing remaining documentation, testing, and integration work.

The architecture follows a modular design with strict separation of concerns. The backend uses `Handler → Service → Repository → Database` layering with dual SQLite/PostgreSQL support. The frontend uses `AuthProvider → WebSocketProvider → Router` with React Query for server state and Zustand for client state. All components communicate via REST APIs and WebSocket connections, with HTTP/3 (QUIC) + Brotli compression as the production transport standard.

### 1.3 The Core Problem: Green Tests, Broken Product

The project exhibits the classic AI-assisted development failure mode where generated code passes isolated unit tests but fails in real-world integration. Root causes include: semantic violations of undocumented institutional knowledge, missing integration between individually-tested components, API contract mismatches between frontend and backend, platform-specific edge cases not covered by generic tests, and environmental assumptions that don't hold across development and production contexts.

The solution requires shifting from trusting AI output to rigorously verifying it at every stage, with a focus on the parts of the system that AI cannot see: runtime behavior, cross-component interactions, protocol-level edge cases, and hardware-specific quirks.

---

## 2. Complete Submodule Inventory (All 41 Modules)

### 2.1 Core Go Submodules (21 Modules)

| # | Submodule | Package | Purpose | Status |
|---|-----------|---------|---------|--------|
| 1 | **Auth** | `digital.vasic.auth` | JWT authentication, bcrypt password helpers | Needs integration testing |
| 2 | **Cache** | `digital.vasic.cache` | Redis-backed caching with TTL management | Core functionality stable |
| 3 | **Challenges** | `digital.vasic.challenges` | Structured test scenario framework | 35/35 passing |
| 4 | **Concurrency** | `digital.vasic.concurrency` | Retry with backoff, offline cache patterns | Needs stress testing |
| 5 | **Config** | `digital.vasic.config` | Configuration management (env, file, validation) | Stable |
| 6 | **Database** | `digital.vasic.database` | Migration patterns, dual SQLite/PostgreSQL | 15 PostgreSQL bugs found |
| 7 | **Discovery** | `digital.vasic.discovery` | Network/service discovery (SMB, mDNS) | Needs real-network testing |
| 8 | **EventBus** | `digital.vasic.eventbus` | Typed event channels and pub/sub | Needs integration tests |
| 9 | **Filesystem** | `digital.vasic.filesystem` | Unified multi-protocol client | Critical: SMB resilience |
| 10 | **Lazy** | `digital.vasic.lazy` | Lazy initialization patterns | Stable |
| 11 | **Media** | `digital.vasic.media` | Media detection, analysis, metadata extraction | Core feature |
| 12 | **Middleware** | `digital.vasic.middleware` | HTTP middleware (CORS, logging, recovery) | Stable |
| 13 | **Observability** | `digital.vasic.observability` | Prometheus metrics, OpenTelemetry | Needs verification |
| 14 | **RateLimiter** | `digital.vasic.ratelimiter` | Pluggable rate limiting | Phase 4.2 hardening done |
| 15 | **Recovery** | `digital.vasic.recovery` | Panic recovery patterns | Stable |
| 16 | **Security** | `digital.vasic.security` | CORS config, CSP headers, sanitization | 24 HIGH gosec items |
| 17 | **Storage** | `digital.vasic.storage` | Object storage abstraction (MinIO/S3) | Needs integration |
| 18 | **Streaming** | `digital.vasic.streaming` | WebSocket hub with room/topic support | Needs E2E testing |
| 19 | **Watcher** | `digital.vasic.watcher` | Filesystem watcher with debouncing | Needs long-running test |
| 20 | **Entities** | (package not specified) | Entity definitions and core types | Stable |
| 21 | **Memory** | (package not specified) | Memory management utilities | Stable |

### 2.2 AI/QA Stack Submodules (9 Modules)

| # | Submodule | Package | Purpose | Status |
|---|-----------|---------|---------|--------|
| 22 | **HelixQA** | (external org) | AI-driven autonomous QA orchestration | Critical: Phase 3 in progress |
| 23 | **LLMsVerifier** | `digital.vasic.llmsverifier` | Strategy-based LLM verification, scoring, ranking | Phase 1 completed |
| 24 | **LLMOrchestrator** | (external org) | Multi-provider LLM orchestration | Building |
| 25 | **LLMProvider** | (external org) | Unified LLM provider interface | Building |
| 26 | **DocProcessor** | (external org) | Document format processing (ADOC, RST) | Stable |
| 27 | **VisionEngine** | (external org) | Vision model backend (llama.cpp RPC) | Needs calibration |
| 28 | **ReplayBuffer** | (external org) | Session replay storage | Stable |
| 29 | **ScreenDiff** | (external org) | Screenshot comparison for QA | Needs tuning |
| 30 | **TrainingCollector** | (external org) | Training data collection | Stable |

### 2.3 React/TypeScript Submodules (6 Modules)

| # | Submodule | Package | Purpose | Status |
|---|-----------|---------|---------|--------|
| 31 | **Auth-Context-React** | `@vasic-digital/auth-context` | React authentication context | Stable |
| 32 | **Catalogizer-API-Client-TS** | `@vasic-digital/api-client` | TypeScript API client | 14-phase completion |
| 33 | **Collection-Manager-React** | `@vasic-digital/collection-manager` | Collection management UI | Stable |
| 34 | **Dashboard-Analytics-React** | `@vasic-digital/dashboard` | Analytics dashboard components | Stable |
| 35 | **Media-Browser-React** | `@vasic-digital/media-browser` | Media browsing UI | Stable |
| 36 | **Media-Player-React** | `@vasic-digital/media-player` | Media player components | Stable |

### 2.4 Utility Submodules (5 Modules)

| # | Submodule | Package | Purpose | Status |
|---|-----------|---------|---------|--------|
| 37 | **Media-Types-TS** | `@vasic-digital/media-types` | Shared TypeScript media type definitions | Stable |
| 38 | **UI-Components-React** | `@vasic-digital/ui-components` | Reusable React UI component library | Stable |
| 39 | **WebSocket-Client-TS** | `@vasic-digital/websocket-client` | Generic WebSocket client with React hooks | Stable |
| 40 | **VisualRegression** | (utility) | Visual regression testing utilities | Stable |
| 41 | **Assets** | (shared) | Shared media assets and resources | Stable |

### 2.5 Main Application Components (7 Components)

| # | Component | Tech Stack | Purpose | Status |
|---|-----------|------------|---------|--------|
| 42 | **catalog-api** | Go 1.25, Gin, SQLCipher | REST API server | 15 PostgreSQL bugs |
| 43 | **catalog-web** | React 18, TypeScript, Vite | Web frontend | 2318 tests pass |
| 44 | **catalogizer-android** | Kotlin, Compose, MVVM | Mobile Android app | Phase 8 instrumented |
| 45 | **catalogizer-androidtv** | Kotlin, Leanback, TVProvider | Android TV app | HELIX-152/154/157/168/175 |
| 46 | **catalogizer-desktop** | Tauri, Rust, React | Desktop app (Win/Mac/Linux) | X-Cover-Quality badge |
| 47 | **installer-wizard** | Tauri, React | Cross-platform setup wizard | 93% coverage |
| 48 | **catalogizer-api-client** | TypeScript | npm package API client | v2.4.0 |

### 2.6 Infrastructure Components (4 Components)

| # | Component | Purpose | Status |
|---|-----------|---------|--------|
| 49 | **OCU-CUDA-Sidecar** | GPU sidecar for llama.cpp RPC on thinker.local | Dockerfile ready |
| 50 | **Build** | Generic shell-based build framework | Orchestrator complete |
| 51 | **Containers** | Container boot system with builder image | Auto-dispatch working |
| 52 | **Upstreams** | Multi-remote git push configuration (6 targets) | Configured |

---

## 3. Critical Documents to Create/Maintain

### 3.1 CLAUDE.md (Primary AI Instruction File)

This file already exists and contains the primary operating manual for Claude Code. It must be kept updated with these sections: Project Overview & Architecture with explicit tech stack versions; Development Commands with exact syntax for each module; Institutional Knowledge with "landmine" rules; Verification Protocol with the self-correction loop; Code Style & Anti-Patterns; and the "Challenge Me" section. The file is located at the repository root and every major submodule has its own CLAUDE.md override.

### 3.2 CONSTITUTION.md (Non-Negotiable Rules)

This critical document contains the project constitution with articles that must never be violated. Article V mandates 100% test coverage across 10 categories (unit, integration, E2E, full automation, stress, security, DDoS/rate-limit, benchmarking, challenges, HelixQA). Article VI establishes the Open-Points Closure Brief as the single source of truth for operator-action items. Article VII defines the Full-QA Master Cycle as the mandatory verification loop. Article VIII mandates device state preservation during QA sessions. Article IX establishes HelixQA tool hygiene requirements.

### 3.3 docs/LANDMINES.md (Production Landmines & Semantic Rules)

This document captures the hidden rules that cause production failures even when tests pass. Each rule must have: a unique ID (e.g., RULE-GO-001), context explaining why the rule exists, detection criteria for the LLM-as-Judge, and the fix pattern. Current rules that must be documented include: the Legacy Billing Service 1MB payload limit, the Android TV HTTP/1.1 requirement (HTTP/2 crashes on MTK chipsets), the analytics non-blocking requirement, the ProGuard `-keep` rule requirement for Android model classes, the WebSocket connection reservation pattern, the cache cleanup goroutine lifecycle, and the SQLite WAL mode explicit pragma requirement.

### 3.4 docs/OPEN_POINTS_CLOSURE.md (Operator Action Tracker)

This is the single source of truth for every outstanding operator-action item. It must be consulted before any work and updated in the same commit that changes an item's state. Deleting an unclosed item is a constitutional violation. The document must track: credential setup status for each external API (TMDB, OMDB, Spotify, Steam, etc.); hardware item availability (Android TV test device, GPU sidecar host); infrastructure tasks (container registry setup, SSL certificate provisioning); and environment configuration status for each deployment target.

---

## 4. The 10-Category Testing Framework

### 4.1 Category 1: Unit Tests (Per-Module)

Every public function must have table-driven tests covering happy path, error path, edge cases, and adversarial inputs. Tests must use `database.WrapDB()` for in-memory SQLite, `testify/suite` for complex suites, and `testify/mock` for mocks. Every submodule must run its unit tests in isolation before integration.

**Verification command per Go submodule:**
```bash
cd <submodule> && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1
```

**Verification command per TypeScript submodule:**
```bash
cd <submodule> && npm run test -- --run
```

**Verification command for Android:**
```bash
cd catalogizer-android && ./gradlew testDebugUnitTest
```

### 4.2 Category 2: Integration Tests (Cross-Module)

Integration tests verify that modules work together correctly. The primary integration points are: catalog-api ↔ Database (dual-dialect SQL rewriting), catalog-api ↔ external metadata providers (TMDB, OMDB, etc.), catalog-web ↔ catalog-api (API contract compliance), catalogizer-android ↔ catalog-api (network + auth flows), and HelixQA ↔ all application surfaces.

**Verification command:**
```bash
cd catalog-api && go test -tags=integration ./... -count=1
podman-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### 4.3 Category 3: E2E Tests (Full User Flows)

End-to-end tests verify complete user workflows across the full stack. The catalog-web uses Playwright for browser automation. The catalog-api uses the Challenges framework's userflow tests (174 challenges across 4 platforms). The Android apps use instrumented tests with Espresso.

**Verification commands:**
```bash
cd catalog-web && npm run test:e2e           # Playwright E2E
cd catalog-api && ./run_challenges.sh         # API userflow challenges
```

### 4.4 Category 4: Full Automation (CI/CD Pipeline)

All tests must run automatically in the local CI pipeline (GitHub Actions are disabled per project requirements). The pipeline runs: code formatting (`go fmt`, `eslint`), static analysis (`go vet`, `golangci-lint`, `tsc --noEmit`), unit tests, integration tests, security scans (govulncheck, npm audit, Semgrep, Snyk, Trivy, Gosec), challenge execution, and HelixQA autonomous testing.

**Verification command:**
```bash
./scripts/run-all-tests.sh
```

### 4.5 Category 5: Stress Tests

Stress tests find the breaking point under extreme load. k6 scripts are provided for load testing (ramp to 50 users, p95 < 500ms), stress testing (ramp to 300 users, find breaking point), and soak testing (20 users for 30 minutes, detect memory leaks).

**Verification command:**
```bash
podman run --rm --network host -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest run /scripts/stress_test.js
```

### 4.6 Category 6: Security Tests

Security testing uses multiple tools: `govulncheck` for Go vulnerability scanning, `npm audit` for JavaScript dependency vulnerabilities, Semgrep with custom rules in `.semgrep.yml`, Snyk for dependency and container scanning, Trivy for filesystem and container scanning, and Gosec for Go security analysis. SonarQube Community Edition provides static code quality analysis.

**Verification command:**
```bash
./scripts/security-scan.sh
```

### 4.7 Category 7: DDoS/Rate-Limit Tests

The rate limiter must withstand spoofed `X-Forwarded-For` headers and distributed attack patterns. Phase 4.2 hardened the rate limiter against `X-Forwarded-For` spoofing but this needs continuous verification.

**Verification command:**
```bash
cd tests && go test -tags=rate_limit ./... -v -run TestDDoS
```

### 4.8 Category 8: Benchmark Tests

Benchmark tests measure performance characteristics and detect regressions. Key benchmarks include: database query performance (SQLite vs PostgreSQL), API endpoint latency (p50, p95, p99), media detection throughput, WebSocket message propagation delay, and file scan performance across protocols.

### 4.9 Category 9: Challenge Tests (174 User Flows)

The Challenges framework provides 174 Catalogizer-specific challenges testing real user flows across all platforms: 49 API challenges (HTTP), 59 web challenges (Playwright), 28 desktop/wizard challenges, and 38 mobile challenges (Android + Android TV). All 35 challenge IDs in `challenge_ids.txt` must pass.

**Verification command:**
```bash
cd catalog-api && ./run_challenges.sh        # Must show 35/35 passing
cd Challenges/cmd/userflow-runner && go run main.go --platform all --report
```

### 4.10 Category 10: HelixQA Autonomous Testing

HelixQA is the sole authorized tool for all automated UI/UX testing. It operates via a 5-phase pipeline: Learn → Plan → Execute → Curiosity → Analyze. Banks live in `banks/full-qa-{api,web,androidtv,android,cross-platform}.yaml` plus `banks/fixes-validation.yaml`. The orchestrator script runs the full pipeline.

**Verification command:**
```bash
./scripts/helixqa-orchestrator.sh            # All platforms
./scripts/helixqa-orchestrator.sh androidtv  # Single platform
```

---

## 5. Per-Submodule Detailed Fix Plans

### 5.1 HelixQA (Submodule #22 — Critical Priority)

**Current State:** Phase 3 (Enhanced Autonomous Session) is in progress. The submodule is from an external organization (HelixDevelopment) and is the sole authorized UI automation tool.

**Complete Fix Plan:**

1. **Complete Phase 3 Implementation (P3-001 through P3-016):** Finish the LLM-powered navigation engine (`llm_navigator.go`), enhance the NavigationGraph with LLM inference, implement the LLM analyzer for issue detection, expand issue categories and severity levels, create the enhanced recorder with timeline support, implement annotated screenshot capture, create the enhanced ticket generator with templates, and wire all components into the SessionCoordinator.

2. **Implement Phase 4 (Open-Points Closure):** Create the mapper for feature-to-test mapping, add DocProcessor format support for ADOC and RST documents, and write all 16 planned unit and integration tests.

3. **Bank Format Standardization:** Ensure all bank files are convertible from YAML to JSON at runtime using `python3 -c "import yaml,json; json.dump(yaml.safe_load(open('bank.yaml')), open('bank.json','w'))"`. Verify all 5 bank files (full-qa-api, full-qa-web, full-qa-androidtv, full-qa-android, full-qa-cross-platform) plus fixes-validation.yaml load without errors.

4. **Vision Pipeline Calibration:** Verify llama.cpp RPC connectivity on thinker.local, calibrate Astica.AI / Gemini / OpenAI fallback providers, and confirm the three strategy types (NavigationStrategy for Execute/Curiosity, AnalysisStrategy for Analyze, PlanningStrategy for Learn/Plan) select appropriate models.

5. **Device State Preservation (Constitution Article VIII):** Verify the deferred cleanup mechanism restores all device settings (`font_scale`, `wm density`, brightness, rotation) after each session. Add any missing preservation entries.

6. **Video Recording Pipeline:** Confirm 16 Mbps minimum bitrate at 1920×1080 resolution. Verify frame extraction works for post-analysis. Test the screenshot-to-video fallback for Android 15 (SDK 35) where `screenrecord` fails with `Encoder failed (err=-38)`.

7. **Curiosity Phase Implementation:** Ensure the "curiosity" phase genuinely explores the app rather than following predetermined paths. Verify it detects UI anomalies, missing assets, and navigation dead-ends.

**Verification:**
```bash
cd HelixQA && go test ./... -count=1
./scripts/helixqa-orchestrator.sh --dry-run
```

### 5.2 LLMsVerifier (Submodule #23 — High Priority)

**Current State:** Phase 1 completed. Strategy pattern implemented with 7 built-in strategies. Recipe builder with fluent API available.

**Complete Fix Plan:**

1. **Verify All 7 Strategies Work End-to-End:** Test each strategy (Default, Vision, Navigation, Analysis, Planning, QA, Catalogizer) with real model lists. Confirm `Rank()` produces consistent orderings and `Select()` respects `Requirements` constraints.

2. **Implement Bridged CLI Discovery:** Verify discovery of Claude Code, Qwen Coder, and OpenCode on PATH. Test the bridged model scoring integration.

3. **Local Model Support via SSH:** Test Ollama instance probing across distributed hosts. Confirm the SSH tunnel establishment and model enumeration work.

4. **Recipe Builder Validation:** Test all predefined recipes for HelixQA use cases. Verify the fluent API chain produces valid configurations.

5. **Wire into HelixQA go.mod:** Ensure the `replace` directive in HelixQA's go.mod points to the correct LLMsVerifier path. Verify version compatibility.

**Verification:**
```bash
cd LLMsVerifier && go test ./... -count=1
cd LLMsVerifier && go test -v -run TestStrategy ./pkg/strategy/
cd LLMsVerifier && go test -v -run TestRecipe ./pkg/recipe/
```

### 5.3 LLMOrchestrator (Submodule #24 — High Priority)

**Current State:** Building. Provides multi-provider LLM orchestration.

**Complete Fix Plan:**

1. **Multi-Provider Pool Verification:** Test the multi_pool.go implementation with multiple concurrent providers. Verify provider failover when one is unavailable.

2. **Headless Mode Handlers:** Complete and test the headless mode handlers for all supported CLI agents: OpenCode, Claude Code, Gemini, Junie, and Qwen Code.

3. **Agent Selector Strategies:** Implement and verify the agent selection strategies (cost-optimized, speed-optimized, quality-optimized, balanced).

4. **Integration with HelixQA:** Verify LLMOrchestrator correctly routes HelixQA requests to the appropriate LLM based on the phase requirements.

**Verification:**
```bash
cd LLMOrchestrator && go test ./... -count=1
```

### 5.4 catalog-api (Component #42 — Critical Priority)

**Current State:** 15 PostgreSQL bugs identified. 35/35 challenges passing for core features. Dynamic port binding, HTTP/3, dual database support, and WebSocket hub all implemented.

**Complete Fix Plan:**

1. **Fix 15 PostgreSQL Bugs:** The specific bugs were identified during a clean-slate retest. Each must be fixed with: the SQL dialect rewriting fix, a unit test using `database.WrapDB()` with PostgreSQL dialect, an integration test against a real PostgreSQL container, and an entry in the fixes-validation challenge bank.

2. **Verify Dual-Dialect SQL Rewriting:** Test all three rewrite functions: `RewritePlaceholders()` (`?` → `$1, $2, ...`), `RewriteInsertOrIgnore()` (`INSERT OR IGNORE` → `ON CONFLICT DO NOTHING`), and `BooleanLiterals()` (`= 0/1` → `= FALSE/TRUE`). Create a comprehensive test matrix covering every SQL statement in the codebase.

3. **SMB Resilience Layer:** Verify the circuit breaker pattern, exponential backoff retry, offline cache, and automatic reconnection all function correctly under real network failure conditions. Test with an actual SMB server and simulate disconnections.

4. **Media Detection Pipeline:** Verify the full pipeline: `UniversalScanner` → `AggregationService.AggregateAfterScan()` → title parser → `MediaItem` creation → hierarchy builder → duplicate detection. Confirm all 11 media types are correctly identified and linked to external metadata.

5. **WebSocket Hub:** Verify the connection reservation pattern prevents race conditions between capacity check and registration. Test the cleanup goroutine lifecycle. Confirm real-time updates propagate to all connected clients.

6. **Lifecycle Management:** Verify `LazyServiceRegistry` correctly orders service initialization. Test the graceful shutdown sequence: `wsHandler.Stop()` → `cacheService.Close()` → HTTP server shutdown.

7. **Rate Limiter Hardening:** Verify Phase 4.2 `X-Forwarded-For` spoofing protection. Test with spoofed headers, distributed attacks, and legitimate forwarded requests.

**Verification:**
```bash
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1
cd catalog-api && go test -tags=integration ./... -count=1
cd catalog-api && go test -tags=postgres ./... -count=1
./scripts/run_challenges.sh                    # 35/35 must pass
podman-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### 5.5 catalog-web (Component #43 — High Priority)

**Current State:** 2,318 tests passing. Lint and TypeScript type-check clean. Build successful. Lighthouse CI configured. Playwright E2E tests present.

**Complete Fix Plan:**

1. **Verify API Contract Compliance:** Compare every API endpoint call in the frontend against the actual catalog-api handler implementations. Confirm request shapes, response shapes, error handling, and status codes match exactly. Document any discrepancies as API contract violations.

2. **WebSocket Integration:** Verify the WebSocket client correctly handles connection establishment, reconnection after disconnect, message parsing, and error recovery. Test with the actual catalog-api running.

3. **Zero Console Warnings/Errors:** Run the full application in development mode and verify zero browser console warnings and zero failed network requests. Any warnings must be eliminated, not suppressed.

4. **E2E Test Execution:** Run all Playwright E2E tests against a running catalog-api instance. Verify all user flows work: authentication, catalog browsing, media search, entity detail views, collection management, and settings.

5. **Production Build Verification:** Verify `npm run build` produces a clean production build with no errors. Confirm all static assets are correctly generated and the build output can be served by nginx.

6. **Performance Budget:** Run Lighthouse CI and verify all performance budgets are met. Check Core Web Vitals (LCP, FID, CLS) against targets.

**Verification:**
```bash
cd catalog-web && npm run lint && npm run type-check   # Must be clean
cd catalog-web && npm run test -- --run                  # 2318 tests
cd catalog-web && npm run test:e2e                       # Playwright
cd catalog-web && npm run build                          # Production build
```

### 5.6 catalogizer-android (Component #44 — High Priority)

**Current State:** Phase 8 instrumented tests added. Build script `build-fixed.sh` works. Unit tests compile and pass.

**Complete Fix Plan:**

1. **Complete Instrumented Tests:** Finish all pending instrumented test files. Ensure Compose UI tests work with the test runner. Fix any remaining test compilation errors.

2. **Server Discovery & URL Persistence:** Verify the server discovery mechanism finds catalog-api instances on the local network. Confirm URL persistence works across app restarts.

3. **Offline Support:** Verify Room database caching works correctly. Test the sync mechanism when the app comes back online. Confirm data integrity during conflict resolution.

4. **ProGuard Rules:** Verify all model classes used in Retrofit/Gson have explicit `-keep` rules in `proguard-rules.pro`. Test a minified release build.

**Verification:**
```bash
cd catalogizer-android && ./gradlew testDebugUnitTest
cd catalogizer-android && ./gradlew connectedDebugAndroidTest  # Requires emulator/device
```

### 5.7 catalogizer-androidtv (Component #45 — Critical Priority)

**Current State:** Multiple HELIX tickets open: HELIX-145 (tv-cold-start), HELIX-154 (TV focus), HELIX-157, HELIX-168, HELIX-175. HTTP/1.1 protocol enforced. Foreground guard implemented.

**Complete Fix Plan:**

1. **Fix HELIX-145 (TV Cold Start):** Investigate and fix the slow cold start issue. This may involve lazy loading of components, optimizing initialization order, or reducing the initial data fetch.

2. **Fix HELIX-154 (TV Focus):** Fix D-Pad navigation focus issues. Every new UI element must implement `Focusable` and handle D-Pad navigation correctly. A blank screen with passing tests indicates a focus problem.

3. **Fix HELIX-157, HELIX-168, HELIX-175:** Address each specific issue documented in the HelixQA tickets. Each fix requires: root cause analysis, code fix, unit test, integration test, HelixQA bank entry, and challenge entry.

4. **Android TV Home Screen Channels:** Verify `androidx.tvprovider` integration. Confirm the default "Catalogizer Picks" channel auto-creates on launch. Test per-category dynamic channels. Verify the Watch Next row integration for partially-watched items. Test deep linking via `catalogizer://media/{id}?type={type}`.

5. **WorkManager Sync:** Verify the 6-hour periodic sync works correctly. Test the app-launch sync trigger and the SyncService manual trigger. Confirm full cleanup on logout.

6. **HTTP/1.1 Protocol Enforcement:** Verify the OkHttp client explicitly sets `protocols(listOf(Protocol.HTTP_1_1))`. Confirm HTTP/2 is never used on TV endpoints.

**Verification:**
```bash
cd catalogizer-androidtv && ./gradlew testTvDebugUnitTest
cd catalogizer-androidtv && ./gradlew lintTvDebug -PabortOnError=true
./scripts/helixqa-orchestrator.sh androidtv
```

### 5.8 catalogizer-desktop (Component #46 — Medium Priority)

**Current State:** X-Cover-Quality debug badge surfaced. Tauri auto-container dispatch works end-to-end.

**Complete Fix Plan:**

1. **Verify Tauri Auto-Container Dispatch:** Confirm builds transparently route through the `catalogizer-builder` container when the host lacks cargo. Verify no sudo/root is required.

2. **X-Cover-Quality Badge:** Ensure the debug badge correctly displays cover image quality metrics. Remove or hide the badge in production builds.

3. **Cross-Platform Builds:** Verify builds succeed for all three targets: Windows, macOS, and Linux. Test the AppImage packaging on Linux with `APPIMAGE_EXTRACT_AND_RUN=1`.

**Verification:**
```bash
cd catalogizer-desktop && npm run tauri:build
```

### 5.9 OCU-CUDA-Sidecar (Component #49 — Medium Priority)

**Current State:** Go source and Dockerfile present. gRPC protocol definitions in `proto/` directory.

**Complete Fix Plan:**

1. **Verify gRPC Server:** Build and run the sidecar. Confirm it accepts gRPC requests from HelixQA/LLMOrchestrator.

2. **GPU Detection:** Verify CUDA GPU detection works. Confirm the sidecar reports GPU availability and memory to the orchestrator.

3. **llama.cpp RPC Proxy:** Test proxying of llama.cpp RPC calls through the sidecar. Verify model loading and inference work.

4. **Container Build:** Build the Dockerfile and verify the container runs correctly with GPU passthrough.

**Verification:**
```bash
cd OCU-CUDA-Sidecar && go build ./...
podman build -f OCU-CUDA-Sidecar/Dockerfile -t ocu-cuda-sidecar:latest .
```

### 5.10 Remaining Go Submodules (#1-#21, excluding already covered)

For each of the remaining Go submodules, the following verification must pass:

1. **Auth:** Verify JWT token generation, validation, and refresh. Test bcrypt password hashing. Confirm role-based access control works.
2. **Cache:** Test Redis connection, TTL management, cache hit/miss patterns. Verify cleanup goroutine lifecycle.
3. **Concurrency:** Test retry with backoff under various failure modes. Verify offline cache pattern works correctly.
4. **Config:** Test configuration loading from env vars, files, and defaults. Verify validation rules.
5. **Database:** Run all migrations against both SQLite and PostgreSQL. Verify `InsertReturningID()` works on both dialects. Confirm WAL mode is active.
6. **Discovery:** Test SMB and mDNS discovery on a real network. Verify service announcement and discovery.
7. **EventBus:** Test typed event channels, pub/sub patterns, and subscriber cleanup.
8. **Filesystem:** Test all 5 protocol implementations (SMB, FTP, NFS, WebDAV, local). Verify the `UnifiedClient` interface is fully implemented by each.
9. **Lazy:** Test lazy initialization patterns and thread safety.
10. **Media:** Test media detection for all 50+ supported types. Verify metadata extraction and external provider integration.
11. **Middleware:** Test all middleware (CORS, logging, recovery, request ID) in isolation and in chains.
12. **Observability:** Verify Prometheus metrics are correctly emitted. Test OpenTelemetry integration.
13. **RateLimiter:** Test memory, Redis, and sliding window implementations. Verify rate limit enforcement.
14. **Recovery:** Test panic recovery patterns and graceful degradation.
15. **Security:** Fix all 24 HIGH items from gosec. Test CORS configuration and CSP headers.
16. **Storage:** Test MinIO/S3-compatible storage abstraction. Verify upload, download, and delete operations.
17. **Streaming:** Test WebSocket hub with multiple rooms and topics. Verify message broadcasting.
18. **Watcher:** Test filesystem watching with debouncing and filtering. Verify cross-platform behavior.
19. **Entities & Memory:** Verify all entity definitions and memory management utilities.

**Verification per submodule:**
```bash
cd <submodule> && go test ./... -count=1
cd <submodule> && go vet ./...
cd <submodule> && golangci-lint run ./...
```

### 5.11 TypeScript/React Submodules (#31-#41)

For each TypeScript/React submodule:

1. **Build Verification:** `npm run build` must succeed with zero errors.
2. **Lint Verification:** `npm run lint` must report zero warnings (with `--max-warnings 0`).
3. **Type Checking:** `npx tsc --noEmit` must report zero type errors.
4. **Test Execution:** All unit tests must pass.

**Verification per submodule:**
```bash
cd <submodule> && npm run lint && npx tsc --noEmit && npm run test -- --run
```

---

## 6. The Full-QA Master Cycle (Constitution Article VII)

### 6.1 Cycle Definition

The Full-QA Master Cycle is a rigid verification loop that must be executed for every release. The cycle consists of: (1) Clean rebuild of all components, (2) All tests across all 10 categories, (3) All Challenges execution (174 user flows), (4) All HelixQA bank execution (all platforms), (5) Autonomous QA per app/platform, (6) Video and screenshot review, (7) Ticket generation for any failures, (8) Root-cause fix with 4 artifacts (unit test + fixes-validation entry + HelixQA bank entry + challenge), (9) Rebuild, and (10) Repeat until clean pass.

The cycle stops only on three conditions: FATAL BLOCKER (the system cannot be tested further), SYSTEM BREAKS (the fix would require architectural changes), or NOTHING LEFT (all tests pass with zero tickets).

### 6.2 Session Archive Layout

Every QA session must archive to `docs/reports/qa-sessions/<YYYY-MM-DD-THH-MM>/` with the following structure: `FINAL-REPORT.md` (comprehensive summary), `logs/` (all application logs), `challenges/` (challenge execution results), `helixqa/` (HelixQA output), `videos/` (screen recordings), `screenshots/` (captured screenshots), `tickets/` (generated tickets), and `analysis/` (root cause analysis).

On clean pass, a version bump is performed and the session is archived to `releases/<platform>/<app>/<version>/`.

### 6.3 Self-Correcting Prompt Chain for AI Agents

When assigning work to Claude Code or other AI agents, use this exact prompt structure:

**Task Assignment:**
```
[SYSTEM DIRECTIVE: VERIFICATION MODE ENABLED]
TASK: Implement [FEATURE NAME] in [MODULE NAME].
REQUIREMENTS:
1. Read docs/LANDMINES.md and confirm you understand the constraints for [MODULE NAME].
2. Read the existing code for similar implementations. Do not write new code until you have found and understood 2+ similar existing implementations.
3. Write the code.
4. Run the verification command: [INSERT COMMAND FROM CLAUDE.md].
CONSTRAINT: You are FORBIDDEN from claiming the task is complete until the verification command exits with code 0.
If the verification fails: Step A: Read the error log. Step B: Fix the error. Step C: Rerun verification. Step D: If it fails again, STOP and output "VERIFICATION BLOCKED: [error]".
BEGIN TASK.
```

**Intermediate Correction:**
```
[CONTINUE VERIFICATION LOOP]
The previous attempt failed with the following output:
--- BEGIN ERROR LOG ---
[Paste actual terminal error]
--- END ERROR LOG ---
ANALYSIS REQUIREMENT: Explain *why* the code caused this error.
ACTION: Fix the code and re-run the verification command.
```

**Final Handoff:**
```
[FINAL REVIEW CHECKLIST]
The verification passed. Before I merge, generate a "Deployment Impact Summary" covering:
1. Which specific Landmine Rules (from docs/LANDMINES.md) were relevant to this change?
2. Did this change modify any public API contract? (Yes/No + Diff).
3. What is the expected behavior of this change in a low-bandwidth network environment?
4. What tests were added to prevent regression?
```

---

## 7. LLM-as-Judge Pre-Merge Gate

### 7.1 Judge Prompt

Before merging any AI-generated code, run this review using a separate Claude Code session:

```
ROLE: Senior Software Architect & QA Gatekeeper.
CONTEXT: We are about to merge a pull request written by a Junior Developer (AI Agent). Your job is to find the "unknown unknowns."
INPUT:
- PR Diff: [PASTE GIT DIFF HERE]
- Landmine Rules: [PASTE CONTENTS OF docs/LANDMINES.md HERE]
- API Contracts: [PASTE RELEVANT SECTIONS OF API_CONTRACTS.md HERE]
TASK: Review the diff against the Landmine Rules, API contracts, and general software best practices for our stack (Go, Kotlin, TypeScript/React, Rust).
OUTPUT FORMAT (JSON):
{
  "veto": true/false,
  "severity": "BLOCKER" / "WARNING" / "INFO",
  "violations": [
    {
      "rule": "RULE-GO-001",
      "description": "Description of violation",
      "fix_suggestion": "How to fix"
    }
  ],
  "risk_assessment": "Low / Medium / High - Explanation.",
  "api_contract_impact": "None / Modified / New endpoint - Details.",
  "test_coverage_gaps": ["List of untested edge cases"]
}
CRITICAL: If you are less than 95% confident that this code will work in production, you MUST set veto: true and severity: BLOCKER.
```

### 7.2 Veto Override Policy

A veto can only be overridden by a human engineer who: (1) Reads the judge's full analysis, (2) Explicitly acknowledges each violation in writing, (3) Provides a documented reason for accepting the risk, and (4) Adds a TODO ticket to address the violation in the next cycle.

---

## 8. Operational Guardrails & Environment Setup

### 8.1 Container Runtime (Podman Only)

All production builds, service deployments, and QA executions must use Podman (never Docker). Mandatory resource limits: Go tests use `GOMAXPROCS=3 -p 2 -parallel 2`, PostgreSQL containers use `--cpus=1 --memory=2g`, catalog-api containers use `--cpus=2 --memory=4g`, catalog-web containers use `--cpus=1 --memory=2g`, builder containers use `--cpus=3 --memory=8g`. Total budget across all containers: 4 CPUs / 8 GB RAM.

Always use `podman build --network host` and `podman run --network host` to avoid SSL issues with external package repositories. Use fully qualified image names (`docker.io/library/...`). Set `GOTOOLCHAIN=local` to prevent Go from auto-downloading newer toolchains. Set `APPIMAGE_EXTRACT_AND_RUN=1` for Tauri AppImage bundling.

### 8.2 Multi-Remote Git Configuration

The project pushes to 6 targets on `origin`: 2× GitHub, 2× GitLab, GitFlic, and GitVerse (port 2222). Configure SSH known hosts:
```bash
ssh-keyscan github.com gitlab.com gitflic.ru >> ~/.ssh/known_hosts
ssh-keyscan -p 2222 gitverse.ru >> ~/.ssh/known_hosts
```

Always push with: `GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main`

### 8.3 Environment Variable Security

Never commit real `.env` files. Use `.env.example` with `YOUR_API_KEY_HERE`. Verify `.gitignore` covers `.env` in the project root and every submodule. Immediately rotate any leaked keys.

### 8.4 `.devignore` Device Exclusion

The `.devignore` file lists device models that must never be used for testing (e.g., ATMOSphere). Before any ADB operation, check:
```bash
DEVICE_MODEL=$(adb -s $DEVICE shell getprop ro.product.model)
grep -qi "$DEVICE_MODEL" .devignore && { echo "❌ in .devignore"; exit 1; }
```

### 8.5 `.devconnect` Android TV Auto-Connect

The `.devconnect` file (gitignored) contains one IP per line for Android TV devices. Run `./scripts/devconnect.sh` before HelixQA sessions. Do not use inline `# comments` — trailing comments get concatenated into the device ID.

---

## 9. Issue Resolution Plan (All 7 Open Issues)

### 9.1 Issue #2: Phase 1 — LLMsVerifier Strategy Pattern Extension

**Status:** Core implementation completed. All 12 tasks (P1-001 through P1-012) finished. Building successfully.

**Remaining work:** Final integration testing with HelixQA. Verify the wiring into HelixQA's go.mod works correctly. Run the full strategy selection pipeline with real model configurations.

### 9.2 Issue #7: Phase 2 — OpenCode Headless Integration

**Status:** Completed per comments ("Phase 7 COMPLETED" was posted here, likely a copy-paste error). Headless mode handlers for OpenCode, Claude Code, Gemini, Junie, and Qwen Code should all be implemented.

**Remaining work:** Verify all 5 headless adapters work with real CLI agent installations. Test the multi-provider pool with multiple concurrent agents. Verify the LLMOrchestrator API correctly routes requests.

### 9.3 Issue #3: Phase 3 — Enhanced Autonomous Session

**Status:** In progress. Dependencies on Phase 1 and Phase 2.

**Remaining work:** Complete all 16 tasks (P3-001 through P3-016). This is the largest remaining work item at 76 estimated hours. Priority order: P3-003 (llm_navigator.go), P3-005 (llm_analyzer.go), P3-007 (enhanced recorder), P3-011 (SessionCoordinator wiring), then all test tasks (P3-013 through P3-016).

### 9.4 Issue #6: Phase 4 — Configuration & Environment

**Status:** Open. No tasks defined in the issue description.

**Required work:** Create comprehensive configuration documentation. Verify all `.env.example` files are complete and accurate. Test configuration loading across all modules. Document every environment variable with its purpose, valid values, and default.

### 9.5 Issue #5: Phase 5 — Comprehensive Testing

**Status:** Open. No tasks defined in the issue description.

**Required work:** Execute all 10 testing categories for all modules. Fill any coverage gaps. Create missing integration tests. Execute the full Challenges suite (174 user flows). Run HelixQA on all platforms. Document all results.

### 9.6 Issue #4: Phase 6 — Documentation

**Status:** Open. 16 tasks defined (P6-001 through P6-016).

**Required work:** Write all user guides (getting started, configuration, running sessions, understanding reports, troubleshooting). Write all developer guides (architecture, extending HelixQA, adding platforms). Create API reference documentation. Create all diagrams (architecture, sequence, flowcharts). Record all 5 video course modules. Estimated: 66 hours.

### 9.7 Issue #8: Phase 7 — Final Integration & Deployment

**Status:** Open. 7 tasks defined (P7-001 through P7-007).

**Required work:** Final code review (8h), fix all remaining issues (8h), run full test suite (4h), create release notes (2h), tag releases in all repos (2h), sync GitHub/GitLab (2h), create project completion report (4h). This is the final phase that can only begin after all other phases complete.

---

## 10. Dependency Map & Critical Path

### 10.1 Build Order

The following order must be respected for builds: (1) All independent Go submodules (Auth, Cache, Config, etc.), (2) LLMsVerifier (depends on independent modules), (3) LLMOrchestrator and LLMProvider (depend on LLMsVerifier), (4) HelixQA (depends on LLMOrchestrator, LLMProvider, DocProcessor, VisionEngine), (5) catalog-api (depends on all Go submodules), (6) catalog-web (depends on catalog-api + TypeScript submodules), (7) catalogizer-android and catalogizer-androidtv (depend on catalog-api), (8) catalogizer-desktop and installer-wizard (depend on catalog-api + Build framework).

### 10.2 Test Order

Tests must run in this order: (1) Unit tests for all submodules in build order, (2) Integration tests for module groups, (3) catalog-api full test suite, (4) catalog-web test suite + E2E, (5) Challenge execution (all 35), (6) HelixQA autonomous testing (all platforms), (7) Security scan suite.

### 10.3 Critical Path to 100% Working Status

The critical path is: Complete HelixQA Phase 3 (P3-003, P3-005, P3-007, P3-011) → Fix catalog-api PostgreSQL bugs (all 15) → Run Full-QA Master Cycle → Fix any tickets generated → Run cycle again until clean pass. This path must be executed sequentially with no parallel work that could introduce new issues.

---

## 11. Execution Timeline & Resource Estimates

### 11.1 Phase Timeline (Conservative Estimates)

| Phase | Work Items | Estimated Hours | Prerequisites |
|-------|-----------|-----------------|---------------|
| HelixQA P3 completion | P3-003, P3-005, P3-007, P3-011 | 30h | Phase 1, Phase 2 |
| HelixQA P3 tests | P3-013 through P3-016 | 20h | Core P3 code |
| catalog-api PG bugs | All 15 PostgreSQL bugs | 24h | None |
| catalog-web E2E | Playwright test execution + fixes | 16h | catalog-api stable |
| Android TV fixes | HELIX-145, 154, 157, 168, 175 | 20h | HelixQA stable |
| Android instrumented | Complete instrumented tests | 16h | None |
| Desktop verification | Cross-platform build verification | 8h | None |
| OCU-CUDA-Sidecar | GPU sidecar verification | 8h | None |
| Remaining Go modules | Test + fix all 21 modules | 32h | None |
| Remaining TS modules | Build + lint + test all 6 | 8h | None |
| Documentation (P6) | All docs + videos | 66h | All code stable |
| Final Integration (P7) | Review, tag, sync, report | 30h | All above |

### 11.2 Total Estimated Effort

**Code work: ~230 hours. Documentation: ~66 hours. Final integration: ~30 hours. Total: ~326 hours.**

This estimate assumes AI-assisted development with the Full-QA Master Cycle running continuously. With the self-correcting prompt chain and LLM-as-Judge gates, the effective human oversight time is estimated at ~80 hours (25% of total).

---

## 12. Success Criteria (100% Definition of Done)

The project is considered 100% complete when ALL of the following conditions are met:

1. **All 7 open GitHub issues are closed** with completed checklists and linked commits.
2. **All 15 PostgreSQL bugs are fixed** with regression tests and fixes-validation entries.
3. **All 35 challenge IDs pass** (challenge_ids.txt shows 35/35).
4. **All 174 userflow challenges pass** across all 4 platforms (API, Web, Desktop, Mobile).
5. **HelixQA executes a full clean pass** on all 5 platforms (api, web, androidtv, android, cross-platform) with zero tickets generated.
6. **Zero console warnings/errors** in catalog-web development mode.
7. **Zero lint warnings** across all TypeScript/React modules (`--max-warnings 0`).
8. **Zero `go vet` issues** across all Go modules.
9. **Zero security scan failures** (govulncheck, npm audit, Semgrep, Snyk, Trivy, Gosec all clean).
10. **All 6 git push targets** receive the final release tag.
11. **All documentation** (user guides, developer guides, API reference, diagrams) is complete.
12. **5 video course modules** are recorded and published.
13. **No TODO, FIXME, or empty implementations** remain in any codebase (enforced by pre-commit hooks).
14. **No open entries** in docs/OPEN_POINTS_CLOSURE.md.
15. **Session archived** to `docs/reports/qa-sessions/<YYYY-MM-DD-THH-MM>/` with FINAL-REPORT.md and all artifacts.

---

## 13. Risk Mitigation

### 13.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| HelixQA vision model unavailability | Medium | High | Fallback to API providers (Astica.AI, Gemini, OpenAI). Skip vision phases if all unavailable. |
| PostgreSQL dialect bugs are symptoms | Medium | High | Each fix includes root cause analysis. If bugs indicate architectural issue, escalate to human. |
| Android TV hardware dependency | Medium | Medium | Use emulator as primary target. Reserve physical device for final verification only. |
| GPU sidecar network issues | Low | Medium | Container includes CPU fallback mode. Mark GPU as optional in orchestrator. |
| Submodule circular dependencies | Low | High | Build order validation in `scripts/release-build.sh` detects cycles before build. |

### 13.2 Process Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| AI generates fixes that break other modules | High | High | Full-QA Master Cycle catches regressions. Self-correcting prompt chain requires passing all tests. |
| Documentation drifts from implementation | Medium | Medium | Documentation tasks (P6) scheduled after code stabilization. AGENTS.md and CLAUDE.md updated in same commits as code changes. |
| Scope creep from new features | Medium | Medium | CONSTITUTION.md Article VI mandates Open-Points tracking. New features require new issues and cannot bypass the cycle. |

---

## 14. Tools & Scripts Reference

### 14.1 Essential Scripts (in `scripts/`)

| Script | Purpose | When to Run |
|--------|---------|-------------|
| `run-all-tests.sh` | Full test suite + security scans | Every cycle iteration |
| `helixqa-orchestrator.sh` | HelixQA autonomous testing | Every cycle iteration |
| `release-build.sh` | Build all 7 components | On clean pass |
| `security-scan.sh` | govulncheck + npm audit + Semgrep + Snyk + Trivy + Gosec | Every cycle iteration |
| `run-sonarqube-scan.sh` | SonarQube static analysis | Before release |
| `services-up.sh` / `services-down.sh` | Start/stop all services | Development/QA sessions |
| `devconnect.sh` | Connect Android TV devices | Before HelixQA TV sessions |
| `setup-submodule.sh` | Initialize new submodules | When adding modules |

### 14.2 Verification Command Quick Reference

```bash
# Full stack verification (run this after every change cycle)
./scripts/run-all-tests.sh && \
  ./scripts/helixqa-orchestrator.sh && \
  echo "✅ ALL SYSTEMS PASS"

# Backend only
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1 && go vet ./...

# Frontend only
cd catalog-web && npm run lint && npm run type-check && npm run test -- --run

# Android only
cd catalogizer-android && ./gradlew testDebugUnitTest lintDebug

# Security only
./scripts/security-scan.sh

# Challenges only
cd catalog-api && ./run_challenges.sh
```

---

## 15. Conclusion

This plan provides a complete, mechanical, and verifiable path to bring Catalogizer to 100% working status across all 52 submodules and components. The key principles are: **trust nothing, verify everything** — every change must pass the Full-QA Master Cycle before being considered complete; **institutional knowledge is code** — every hidden rule is documented in LANDMINES.md and checked by the LLM-as-Judge; **regressions are unacceptable** — every fix includes 4 artifacts (test, validation entry, HelixQA entry, challenge); and **the cycle is the product** — shipping is prohibited while any test category is incomplete or any ticket is open.

Execute this plan sequentially, one submodule at a time, with the self-correcting prompt chain and LLM-as-Judge gates at every merge point. The 326-hour estimate is conservative; with effective AI assistance, the actual time will likely be shorter. What matters is not speed but correctness — every feature must work 100% with no exceptions.

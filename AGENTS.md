# AGENTS.md - Catalogizer Development Guide

## ⚠️ CRITICAL CONSTRAINTS (Non-negotiable)

**Every agent MUST follow these rules:**

- **ZERO UNFINISHED WORK**: No TODOs, FIXMEs, or known issues allowed in codebase. Fix all discovered issues immediately.
- **100% TEST COVERAGE**: Every component needs unit, integration, E2E, full automation, stress, security, DDoS/rate-limit, benchmark, challenges, and HelixQA tests.
- **HELIXQA ONLY**: All automated UI/UX testing MUST use HelixQA. Never write custom ADB scripts or use third-party tools.
- **REAL-TIME LOGS**: Monitor logs in real-time during ALL QA sessions (Android logcat, browser console, backend services). Pause session on ANR/crash.
- **NO SUDO/ROOT**: ALL operations run at local user level. Use containers for system-level dependencies.
- **SSH GIT ONLY**: Use `git@github.com:` URLs, never HTTPS. Configure remotes to use SSH.
- **SECRETS**: Never commit .env files or hardcode API keys. Use .env.example with placeholders.
- **CONTAINERS**: All builds, services, and QA testing MUST use Podman containers. Never run directly on host.
- **HTTP/3+BROTLI**: All network communication MUST use HTTP/3 (QUIC) with Brotli compression.
- **RESOURCE LIMITS**: Limit workloads to 30-40% of host resources. Use `GOMAXPROCS=3` for Go tests.
- **UNIVERSAL SOLUTIONS**: Fixes must work with ANY application. Never add test-only code to target app.
- **DEVICE RULES**: Check `.devignore` before ADB use. Run `./scripts/devconnect.sh` before HelixQA.
- **FULL-QA CYCLE**: Execute: clean rebuild → tests → Challenges → HelixQA → tickets → fixes → validation tests → repeat until clean pass.

## 🔧 ESSENTIAL COMMANDS

**Backend (catalog-api):**
```bash
go run main.go          # dev server (writes .service-port)
GOMAXPROCS=3 go test ./... -p 2 -parallel 2  # resource-limited tests
go test -v -run ^TestName$ ./path/to/pkg/    # single test
go test -v -run ^TestSuiteName/TestSubtest$ ./path/  # single subtest
```

**Frontend (catalog-web):**
```bash
npm run dev                 # dev server (port 3000, proxies /api)
npm run test:e2e            # Playwright E2E tests
npm run test:e2e -- --grep "test title"  # single E2E test
npm run lint                # ESLint (--max-warnings 0 enforced)
npm run type-check          # tsc --noEmit
```

**Android (catalogizer-android/tv):**
```bash
./gradlew test              # all unit tests
./gradlew test --tests "*TestClassName.testMethod"  # single test method
./gradlew assembleDebug     # debug APK
```

**Full System:**
```bash
./scripts/run-all-tests.sh  # all tests + security scans
podman-compose -f docker-compose.dev.yml up  # dev environment
./scripts/services-up.sh    # start all services
```

## 🏗️ ARCHITECTURE NOTES

**Submodules:** 41 independent git submodules under `digital.vasic.*` and `@vasic-digital/*`
- Go modules: wired via `replace` in `catalog-api/go.mod` (23 modules)
- TS/React modules: linked via `file:../` in `catalog-web/package.json` (9 modules)
- HelixQA/AI: 9 modules for QA/testing

**Key Files:**
- `catalog-api/main.go` → API entry point, route registration
- `catalog-api/database/dialect.go` → dual-dialect SQL rewriting (SQLite/PostgreSQL)
- `catalog-api/filesystem/interface.go` → `UnifiedClient` protocol abstraction
- `catalog-web/src/App.tsx` → React root (AuthProvider → WebSocketProvider → Router)
- `catalog-web/vite.config.ts` → path aliases, API proxy (reads `../catalog-api/.service-port`)

**Media Entity System:**
Scanned files → UniversalScanner → AggregationService → MediaItem/MediaFile linking → Hierarchy builder (TV: show→season→episode) → Entity API → Entity Browser UI

**Dynamic Port Binding:** API writes chosen port to `.service-port` on startup. Frontend reads this for proxy target.

## 🧪 TESTING QUIRKS

**Validation Requirement:** Every bug fix MUST include a bank test entry in `fixes-validation.yaml` before closing ticket.

**QA Campaign Protocol:** Rebuild → Execute all tests → Analyze results → Create tickets → Fix root causes → Create validation tests → Repeat. Loop stops only on: all pass, fatal blocker, or nothing left.

**Video Recording:** ALL device/emulator QA sessions MUST record video. Analyze for: visual glitches, UI/UX issues, content gaps, brand compliance, performance, crashes.

**HelixQA Bank Format:** Requires JSON. Convert YAML: `python3 -c "import yaml,json; json.dump(yaml.safe_load(open('bank.yaml')), open('bank.json','w'))"`

**Catalog Population:** QA tests require populated database. Run populate challenge or configure SMB storage first.

## ⚙️ SETUP REQUIREMENTS

**Container Runtime:**
- Always use Podman (no Docker)
- `podman build --network host` and `podman run --network host` for builds
- Set `GOTOOLCHAIN=local` to prevent Go auto-downloading newer toolchain versions
- Use fully qualified image names (`docker.io/library/...`)
- Set `APPIMAGE_EXTRACT_AND_RUN=1` for Tauri AppImage bundling (no FUSE)
- API container needs `--add-host=synology.local:192.168.0.241` for NAS access

**Environment:**
- Config precedence: `env vars > .env > config.json > defaults`
- `.env` at project root required with all LLM provider keys wired
- Android SDK builds MUST use `catalogizer-builder` container via Containers submodule
- `.devconnect` lists IPs to auto-connect via `adb connect` (run before HelixQA)

**Prerequisites:**
- All git submodules initialized: `git submodule init && git submodule update --recursive`
- Builder container image: `localhost/catalogizer-builder:latest` (build if missing)


## Universal Mandatory Constraints

These rules are inherited from the cross-project Universal Mandatory Development Constraints (canonical source: `/tmp/UNIVERSAL_MANDATORY_RULES.md`, derived from the HelixAgent root `CLAUDE.md`). They are non-negotiable across every project, submodule, and sibling repository. Project-specific addenda are welcome but cannot weaken or override these.

### Hard Stops (permanent, non-negotiable)

1. **NO CI/CD pipelines.** No `.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`, `.travis.yml`, `.circleci/`, or any automated pipeline. No Git hooks either. All builds and tests run manually or via Makefile / script targets.
2. **NO HTTPS for Git.** SSH URLs only (`git@github.com:…`, `git@gitlab.com:…`, etc.) for clones, fetches, pushes, and submodule operations. Including for public repos. SSH keys are configured on every service.
3. **NO manual container commands.** Container orchestration is owned by the project's binary / orchestrator (e.g. `make build` → `./bin/<app>`). Direct `docker`/`podman start|stop|rm` and `docker-compose up|down` are prohibited as workflows. The orchestrator reads its configured `.env` and brings up everything.

### Mandatory Development Standards

1. **100% Test Coverage.** Every component MUST have unit, integration, E2E, automation, security/penetration, and benchmark tests. No false positives. Mocks/stubs ONLY in unit tests; all other test types use real data and live services.
2. **Challenge Coverage.** Every component MUST have Challenge scripts (`./challenges/scripts/`) validating real-life use cases. No false success — validate actual behavior, not return codes.
3. **Real Data.** Beyond unit tests, all components MUST use actual API calls, real databases, live services. No simulated success. Fallback chains tested with actual failures.
4. **Health & Observability.** Every service MUST expose health endpoints. Circuit breakers for all external dependencies. Prometheus / OpenTelemetry integration where applicable.
5. **Documentation & Quality.** Update `CLAUDE.md`, `AGENTS.md`, and relevant docs alongside code changes. Pass language-appropriate format/lint/security gates. Conventional Commits: `<type>(<scope>): <description>`.
6. **Validation Before Release.** Pass the project's full validation suite (`make ci-validate-all`-equivalent) plus all challenges (`./challenges/scripts/run_all_challenges.sh`).
7. **No Mocks or Stubs in Production.** Mocks, stubs, fakes, placeholder classes, TODO implementations are STRICTLY FORBIDDEN in production code. All production code is fully functional with real integrations. Only unit tests may use mocks/stubs.
8. **Comprehensive Verification.** Every fix MUST be verified from all angles: runtime testing (actual HTTP requests / real CLI invocations), compile verification, code structure checks, dependency existence checks, backward compatibility, and no false positives in tests or challenges. Grep-only validation is NEVER sufficient.
9. **Resource Limits for Tests & Challenges (CRITICAL).** ALL test and challenge execution MUST be strictly limited to 30-40% of host system resources. Use `GOMAXPROCS=2`, `nice -n 19`, `ionice -c 3`, `-p 1` for `go test`. Container limits required. The host runs mission-critical processes — exceeding limits causes system crashes.
10. **Bugfix Documentation.** All bug fixes MUST be documented in `docs/issues/fixed/BUGFIXES.md` (or the project's equivalent) with root cause analysis, affected files, fix description, and a link to the verification test/challenge.
11. **Real Infrastructure for All Non-Unit Tests.** Mocks/fakes/stubs/placeholders MAY be used ONLY in unit tests (files ending `_test.go` run under `go test -short`, equivalent for other languages). ALL other test types — integration, E2E, functional, security, stress, chaos, challenge, benchmark, runtime verification — MUST execute against the REAL running system with REAL containers, REAL databases, REAL services, and REAL HTTP calls. Non-unit tests that cannot connect to real services MUST skip (not fail).
12. **Reproduction-Before-Fix (CONST-032 — MANDATORY).** Every reported error, defect, or unexpected behavior MUST be reproduced by a Challenge script BEFORE any fix is attempted. Sequence: (1) Write the Challenge first. (2) Run it; confirm fail (it reproduces the bug). (3) Then write the fix. (4) Re-run; confirm pass. (5) Commit Challenge + fix together. The Challenge becomes the regression guard for that bug forever.
13. **Concurrent-Safe Containers (Go-specific, where applicable).** Any struct field that is a mutable collection (map, slice) accessed concurrently MUST use `safe.Store[K,V]` / `safe.Slice[T]` from `digital.vasic.concurrency/pkg/safe` (or the project's equivalent primitives). Bare `sync.Mutex + map/slice` combinations are prohibited for new code.

### Definition of Done (universal)

A change is NOT done because code compiles and tests pass. "Done" requires pasted terminal output from a real run, produced in the same session as the change.

- **No self-certification.** Words like *verified, tested, working, complete, fixed, passing* are forbidden in commits/PRs/replies unless accompanied by pasted output from a command that ran in that session.
- **Demo before code.** Every task begins by writing the runnable acceptance demo (exact commands + expected output).
- **Real system, every time.** Demos run against real artifacts.
- **Skips are loud.** `t.Skip` / `@Ignore` / `xit` / `describe.skip` without a trailing `SKIP-OK: #<ticket>` comment break validation.
- **Evidence in the PR.** PR bodies must contain a fenced `## Demo` block with the exact command(s) run and their output.

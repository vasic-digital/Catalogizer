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
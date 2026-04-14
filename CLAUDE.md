# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-platform media collection manager. Detects, categorizes, and organizes media across SMB, FTP, NFS, WebDAV, and local filesystems. Components: **catalog-api** (Go 1.25/Gin backend), **catalog-web** (React 18/TS/Vite frontend), **catalogizer-desktop** & **installer-wizard** (Tauri/Rust+React), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

## ⚠️ Mandatory Constraints (Non-Negotiable)

### Zero Unfinished Work Policy

**No TODOs, FIXMEs, empty implementations, silent error swallows, fake data, panic-prone `unwrap()`, or empty catch blocks may be committed.** Pre-commit hooks block them; CI fails on them.

When an issue is found, fix **all** instances — not just the reported one. Definition of "done": compiles without warnings, all bugs fixed, all error cases handled, all tests pass with real assertions.

### No Sudo / No Root

All operations run at local-user level only. Never `sudo`, never `root`, never elevated privileges. Use rootless Podman, user systemd, user-writable directories. If a command requires elevation, find a user-level alternative.

### API Keys & Secrets

Never commit `.env` files with real keys, never hardcode secrets in source / CLAUDE.md / AGENTS.md. Use `.env.example` with `YOUR_API_KEY_HERE` placeholders. Verify `.gitignore` covers `.env` before every commit. Rotate any leaked key immediately. All submodules must have `.env` in their `.gitignore`.

### Git Access via SSH Only

Always SSH (`git@github.com:...`), never HTTPS. Submodules in `.gitmodules` must use SSH URLs. CI/CD uses SSH deploy keys. GitVerse uses port 2222.

### GitHub Actions Permanently Disabled

Do not create files in `.github/workflows/`. CI/CD runs locally only.

### Containers Mandatory for Builds, Services, QA

All builds (`./scripts/release-build.sh --container`), services (`podman-compose`), and QA testing run in containers. Bare-metal `go run` / `npm run dev` is acceptable only for rapid local iteration, never production or QA.

### .devignore — Devices Forbidden for Testing

Devices listed in `.devignore` (e.g. ATMOSphere) **must never** be used for any testing, QA, or app deployment. Before any ADB operation, check device model against `.devignore` and abort if matched. If no valid devices are connected, abort the session — never fall back to excluded devices.

```bash
DEVICE_MODEL=$(adb -s $DEVICE shell getprop ro.product.model)
if grep -qi "$DEVICE_MODEL" .devignore; then
  echo "❌ Device $DEVICE_MODEL is in .devignore - CANNOT USE"; exit 1
fi
```

### .devconnect — Auto-Connect Devices

`.devconnect` (gitignored) lists IPs of Android TV devices to auto-connect before HelixQA runs. Format: one IP per line (`192.168.0.214` or `192.168.0.214:5555`). Run `./scripts/devconnect.sh` to validate reachability and connect (idempotent).

### Universal Solution Principle

All fixes and testing infrastructure must be **universal** — working with any application, not just Catalogizer. Never add test-only code to the app under test (no `QAInputReceiver`, no test endpoints, no bypasses). Fix detection / input / parsing in HelixQA itself, not in the app. Any solution that modifies the app under test is invalid.

### Real-Time Log Monitoring

Every QA session must stream logs in real time: `adb logcat` (Android), browser console (web), application + system logs (desktop), service logs (backend). ANR/crash/fatal-exception detection must pause the session immediately. No QA session is valid without live log monitoring.

### Host Resource Limits (30–40% Maximum)

The host runs other mission-critical processes. Workloads must stay under 30–40% of total resources or the system can freeze.

- **Go tests**: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- **Container limits** (mandatory `podman run` flags):
  - PostgreSQL: `--cpus=1 --memory=2g`
  - catalog-api: `--cpus=2 --memory=4g`
  - catalog-web: `--cpus=1 --memory=2g`
  - Builder: `--cpus=3 --memory=8g`
- **Total budget**: max 4 CPUs, 8 GB RAM across all running containers
- **Challenges**: run sequentially via the API, never in parallel
- **Monitor**: `podman stats --no-stream`, `cat /proc/loadavg`

### HTTP/3 (QUIC) + Brotli Mandatory

All network communication uses HTTP/3 with Brotli compression. Fallback: HTTP/2 + gzip. Never HTTP/1.1 in production. catalog-api uses `quic-go/http3` + `andybalholm/brotli`. Android uses OkHttp + Cronet + Brotli.

### Zero Warning / Zero Error Policy

All components run with zero console warnings, zero console errors, and zero failed network requests in every environment. Every failed network request is a defect. If a feature is not yet implemented, provide a stub endpoint that returns a valid empty response. The challenge suite enforces this end-to-end.

### 100% Test Coverage Across All Categories (Constitution Article V)

**Every component must maintain no-less-than 100% coverage in every one of these ten categories** — none may be skipped, deferred, or partially covered:

1. Unit — pure logic, individual functions / classes
2. Integration — cross-module, DB, cache, queues, filesystems
3. E2E — full user journeys through the live system
4. Full automation — unattended, reproducible, CI-runnable E2E
5. Stress — saturation, concurrency, large payloads, long sessions
6. Security — authn/z, injection, SSRF, secrets, CVE scans (`govulncheck`, `npm audit`, Semgrep, Gosec, Trivy)
7. DDoS / rate-limit — floods, bursts, slowloris, connection exhaustion, rejection + recovery verification
8. Benchmarking — latency / throughput / memory baselines with regression detection
9. Challenges — registered `digital.vasic.challenges` entry per feature
10. HelixQA — autonomous bank + session entry per screen, flow, and adversarial case

"100%" means every branch (happy, error, edge, adversarial) of every public function / endpoint / UI component is exercised, every feature has an E2E flow + challenge + HelixQA bank entry, and every fix has a regression test in the `fixes-validation` bank **before** the ticket is closed.

**Mandatory retesting loop** (rebuild → execute all categories → analyze → ticket → fix + regression → repeat) runs until a full pass is clean. **Shipping is prohibited** while any category is incomplete or any ticket is open.

Coverage is achieved **sequentially, one platform at a time**, across all services and applications. See `CONSTITUTION.md` Article V for the full text.

## HelixQA: Autonomous LLM-Driven Testing

HelixQA is the **sole authorized tool** for all automated UI/UX testing across Android TV, Android phone, web, and desktop. Pipeline: **Learn → Plan → Execute → Curiosity → Analyze**. Run via `helixqa autonomous --platforms androidtv` or use the orchestrator script (below). Configuration in `HelixQA/.env.example`.

### Absolute Rules

- **No manual UI scripts.** No custom ADB tap sequences, coordinate-based automation, shell scripts for UI interactions, Appium scripts outside HelixQA, or standalone Playwright scripts. HelixQA is it.
- **Vision-driven only.** Every action is `screenshot → LLM analysis → action decision`. No hardcoded coordinates, no sleep timers, no keystroke sequences, no fallback hardcoded actions. If vision providers are unavailable, the phase **skips** — never fakes results.
- **Fix in HelixQA, not the app.** Parsing bugs, retry logic, prompt issues — all fixed in HelixQA Go code. Never work around with shell scripts.
- **All connected ADB devices tested** (except `.devignore` entries). ADB reverse proxy auto-configured.
- **QA priority order**: (1) happy paths (login, browse, play), (2) standard flows, (3) edge cases, (4) adversarial.
- **Never type credentials into non-login fields** — the LLM must understand which screen it's on.

### Vision Architecture

Phase-specific model selection via LLMsVerifier strategies:
- **NavigationStrategy** (Execute / Curiosity): JSON-action models
- **AnalysisStrategy** (Analyze): rich-description models
- **PlanningStrategy** (Learn / Plan): reasoning models

**llama.cpp RPC distributed inference** is the primary local backend. Cloud providers (Astica.AI, Gemini, OpenAI) complement it. Models are scored dynamically per phase — no hardcoded preferences. Bridged CLI models discovered via `pkg/bridge/`.

### Screen Recognition & Action Verification

A QA system that cannot detect "stuck on same screen" is useless. HelixQA must:

1. **Track screen state** — compare current vs previous; report stagnation (>10s identical frames) as a critical issue.
2. **Verify actions executed** — confirm screen state changed as expected after each action; fail the test if no change.
3. **Use executable actions in test banks**, never prose descriptions:
   ```yaml
   - name: Type username
     action: "adb_shell: input text admin"
     expected: "Username field populated"
   ```
   not:
   ```json
   { "action": "Enter admin/admin123 credentials", "expected": "Home screen loads" }
   ```
4. **Frame-by-frame video analysis** — extract frames at 1s intervals, compare N to N+1, report stagnation if identical for >5s after an action.
5. **Auto-report**: stuck screens, login never attempted, home never reached, blank/black screens, ANR/crash indicators.

QA reports showing "100% pass" when the app never progressed past login are fraudulent and unacceptable.

### Mandatory Video Recording (Device / Emulator)

Every Android device or emulator session **must** record video.

- **Android 9 and below** (Mi Box, emulators): `adb shell screenrecord --bit-rate 4000000 /sdcard/qa_session.mp4`
- **Android 10+**: rapid `adb shell screencap` assembled into video via ffmpeg
- **Android 15 (SDK 35)**: `screenrecord` fails with `Encoder failed (err=-38)` — use screenshot-to-video
- **Web**: Playwright `--video on` or ffmpeg x11grab
- **Desktop (Tauri)**: ffmpeg x11grab or Xvfb
- **Recording quality**: 16 Mbps minimum, 1920×1080 for frame extraction
- **Frame extraction**: `ExtractFrameAt(timestamp)` and `ExtractLatestFrame()` from video — higher quality and more reliable than direct screenshots
- **Pull videos after recording**: `adb pull /sdcard/qa_session.mp4`
- **Output**: `qa-results/video-sessions-<timestamp>/` with per-device frames and analysis markdown

Every video must be analyzed for: misaligned UI, clipped text, wrong colors, missing assets, unresponsive buttons, broken animations, empty screens with backend data, brand compliance (Vasic Digital logo in rounded square with red border), jank, frozen frames, app restarts.

### Evidence Validation

- Visually inspect every screenshot to verify expected screen state.
- Login verification: UI dump must **not** contain "Sign In" after login attempt — if it does, login failed.
- Compare API responses against on-screen content. Empty screens with backend data = bug.
- Verify every phase transition before proceeding to the next.
- Cross-reference screen content against codebase logic and database state.

### Iterative Test-Fix-Rebuild Loop

QA campaigns must follow this loop until all pass / fatal blocker / nothing left:

1. **Rebuild** affected binaries, containers, deployments
2. **Execute** all tests (unit, challenges, HelixQA bank, autonomous)
3. **Analyze** results, videos, screenshots, logs
4. **Create tickets** for every defect (severity, evidence, repro steps)
5. **Fix** root causes (never workarounds); add a regression test to the **Fixes Validation** bank
6. **Repeat** from step 1

Validation tests are permanent — they persist across all future QA campaigns.

### Live Monitoring & Reporting Layout

Every test run captures real-time platform / test-case-ID / pass-fail-skip / aggregate stats and archives everything to:

```
docs/reports/qa-sessions/qa-session-YYYY-MM-DD/
├── FINAL-REPORT.md
├── logs/                  # unit-tests-go.log, unit-tests-frontend.log, challenges.log,
│                          # helixqa-bank.log, helixqa-autonomous-<platform>.log
├── challenges/            # JSON results + summary
├── helixqa/               # bank-results/, autonomous/
├── videos/
├── screenshots/
├── tickets/
└── analysis/
```

### Comprehensive Bank Coverage

Test banks must cover all features, all screens, all use cases, all 11 media types, all CRUD operations, search (Cyrillic / special chars / SQL injection / XSS payloads), auth flows, navigation (incl. TV channels + DPAD), media interaction, settings, edge cases, negative data. Data sets include real catalog content from NAS, known TMDB / OpenLibrary / MusicBrainz entries, boundary values, and internationalized content.

Bank files: `banks/full-qa-{api,web,androidtv,android,cross-platform}.yaml`, `banks/fixes-validation.yaml`. **Bank format is JSON** — convert YAML with `python3 -c "import yaml,json; json.dump(yaml.safe_load(open('bank.yaml')), open('bank.json','w'))"`.

### One-Command Execution

```bash
./scripts/helixqa-orchestrator.sh [platforms]   # all platforms by default
./scripts/helixqa-orchestrator.sh android       # Android TV only
./scripts/helixqa-orchestrator.sh web           # web only
```

Phases: env validation → device connect (`.devconnect`) → APK install (builds if needed) → background health monitoring → autonomous testing → report generation. Output: `qa-results/session-<timestamp>/`.

## Submodule Architecture

41 independent git submodules under the `vasic-digital` org. Each has its own GitHub + GitLab repo, tests, docs, and `Upstreams/` for multi-remote push.

### Go Modules (wired via `replace` in `catalog-api/go.mod`)

| Module | Path | Description |
|--------|------|-------------|
| `digital.vasic.challenges` | `Challenges/` | Challenge framework |
| `digital.vasic.assets` | `Assets/` | Asset management |
| `digital.vasic.containers` | `Containers/` | Container discovery + service ports |
| `digital.vasic.concurrency` | `Concurrency/` | Concurrency utilities |
| `digital.vasic.config` | `Config/` | Configuration management |
| `digital.vasic.filesystem` | `Filesystem/` | Unified filesystem protocol abstraction |
| `digital.vasic.auth` | `Auth/` | Authentication primitives |
| `digital.vasic.cache` | `Cache/` | Caching layer |
| `digital.vasic.entities` | `Entities/` | Entity model definitions |
| `digital.vasic.eventbus` | `EventBus/` | Pub/sub event bus |
| `digital.vasic.database` | `Database/` | Connection, dialect abstraction, migrations |
| `digital.vasic.discovery` | `Discovery/` | Service discovery |
| `digital.vasic.lazy` | `Lazy/` | Generic lazy loading |
| `digital.vasic.media` | `Media/` | Media detection + analysis pipeline |
| `digital.vasic.memory` | `Memory/` | Memory leak detection |
| `digital.vasic.middleware` | `Middleware/` | HTTP middleware |
| `digital.vasic.observability` | `Observability/` | Metrics, logging, tracing |
| `digital.vasic.ratelimiter` | `RateLimiter/` | Rate limiting |
| `digital.vasic.recovery` | `Recovery/` | Circuit breaker + recovery patterns |
| `digital.vasic.security` | `Security/` | Security utilities |
| `digital.vasic.storage` | `Storage/` | Storage abstraction |
| `digital.vasic.streaming` | `Streaming/` | Media streaming |
| `digital.vasic.watcher` | `Watcher/` | Filesystem watcher |

### TypeScript / React Modules (linked via `file:../` in `catalog-web/package.json`)

| Module | Path | Description |
|--------|------|-------------|
| `@vasic-digital/websocket-client` | `WebSocket-Client-TS/` | WebSocket client + React hooks |
| `@vasic-digital/ui-components` | `UI-Components-React/` | React UI library |
| `@vasic-digital/media-types` | `Media-Types-TS/` | Shared media type definitions |
| `@vasic-digital/catalogizer-api-client` | `Catalogizer-API-Client-TS/` | TypeScript API client |
| `@vasic-digital/auth-context` | `Auth-Context-React/` | Auth context provider |
| `@vasic-digital/media-browser` | `Media-Browser-React/` | Media browsing components |
| `@vasic-digital/media-player` | `Media-Player-React/` | Playback components |
| `@vasic-digital/collection-manager` | `Collection-Manager-React/` | Collection management UI |
| `@vasic-digital/dashboard-analytics` | `Dashboard-Analytics-React/` | Dashboard + analytics |

### HelixQA / AI Submodules

`HelixQA/`, `DocProcessor/`, `LLMOrchestrator/`, `LLMProvider/`, `VisionEngine/`, `ReplayBuffer/`, `ScreenDiff/`, `TrainingCollector/`, `VisualRegression/`.

### Submodule Commands

```bash
git submodule update --init --recursive               # after cloning
git submodule update --remote --recursive             # sync to latest tracked branches
./scripts/setup-submodule.sh ModuleName [--create-repos] [--go|--ts|--kotlin]
cd SubmoduleName && commit "message"                  # push to all upstreams
cd SubmoduleName && install_upstreams                 # install upstream remotes
```

## Commands

### Backend (catalog-api)

```bash
cd catalog-api
go run main.go                                   # dev server (dynamic port → .service-port)
go build -o catalog-api                          # build binary
GOMAXPROCS=3 go test ./... -p 2 -parallel 2      # all tests (resource-limited)
go test -v -run TestName ./path/to/pkg/          # single test
go test -v -run TestSuite/TestSubtest ./pkg/     # subtest
go fmt ./... && go vet ./...                     # format + vet
```

Test helper at `catalog-api/internal/tests/test_helper.go` provides SQLite test DB via `database.WrapDB()`.

### Frontend (catalog-web) — port 3000, proxies `/api` to catalog-api

```bash
cd catalog-web
npm run dev                                      # dev server
npm run test                                     # vitest (single run)
npm run test:watch                               # watch mode
npm run test:coverage                            # with coverage
npx vitest run path/to/file.test.ts              # single file
npx vitest run -t "test name pattern"            # single test by name
npm run test:e2e                                 # Playwright E2E
npm run build                                    # tsc + vite production build
npm run lint && npm run type-check               # ESLint --max-warnings 0 + tsc --noEmit
```

### Desktop / Wizard (Tauri)

```bash
cd catalogizer-desktop          # or installer-wizard
npm run tauri:dev
npm run tauri:build
```

### API Client

```bash
cd catalogizer-api-client
npm run build && npm run test
```

### Android

```bash
cd catalogizer-android   # or catalogizer-androidtv
./gradlew test                                                 # all unit tests
./gradlew :app:testDebugUnitTest --tests ClassName             # single class
./gradlew :app:testDebugUnitTest --tests ClassName.methodName  # single method
./gradlew assembleDebug                                        # debug APK
```

### Full System

```bash
podman-compose -f docker-compose.dev.yml up                    # dev env
./scripts/services-up.sh                                       # start all services
./scripts/services-down.sh                                     # stop all services
./scripts/run-all-tests.sh                                     # all tests + security
./scripts/release-build.sh --container --force --skip-tests    # release build (all 7 components)
./scripts/run-sonarqube-scan.sh                                # SonarQube
./scripts/security-scan.sh                                     # govulncheck + npm audit + Semgrep
```

### Pre-Commit Checklist

```bash
cd catalog-api && go fmt ./... && go vet ./...
cd catalog-web && npm run lint && npm run type-check
pre-commit run --all-files
```

Verify zero browser console warnings/errors. Verify `.gitignore` covers `.env`.

## Architecture

### catalog-api (Go / Gin)

`Handler → Service → Repository → SQLite/PostgreSQL`. Routes mounted under `/api/v1` in `main.go`.

- **Dual package layout**: top-level `handlers/`, `repository/`, `services/`, `middleware/` for domain logic; `internal/handlers/`, `internal/services/`, `internal/middleware/` for infrastructure.
- `filesystem/interface.go` defines `UnifiedClient`; `filesystem/factory.go` builds per-protocol clients. New protocol = implement the interface.
- `internal/smb/`: circuit breaker + offline cache + exponential-backoff retry.
- `internal/media/detector/` → `analyzer/` → `providers/` (TMDB, OMDB, OpenLibrary, MusicBrainz, …): detection pipeline.
- `internal/media/realtime/`: event bus → WebSocket → clients.
- `internal/auth/` + `middleware/`: JWT auth with role-based access.
- `internal/metrics/`: Prometheus metrics on `/metrics`.
- **Dynamic port binding**: writes the bound port to `.service-port` at startup; the frontend reads this for proxy target.
- **HTTP/3 (QUIC)**: `quic-go/http3` server with self-signed TLS certs generated at startup.
- **Redis**: optional cache via `go-redis/v9`.
- **Version injection**: `Version`, `BuildNumber`, `BuildDate` via `-ldflags`.
- `internal/lifecycle/`: `LazyServiceRegistry` — deferred service init with dependency ordering.
- `internal/concurrency/`: semaphore-based parallelism control.
- `internal/httpclient/`: pooled HTTP client with reuse, timeouts, retries.

### Database Layer

Dual-dialect abstraction supporting SQLite (dev) and PostgreSQL (production).

- `database/dialect.go`: `DialectType` (DialectSQLite | DialectPostgres) with query rewriting:
  - `RewritePlaceholders()` — `?` → `$1, $2, …` for PostgreSQL
  - `RewriteInsertOrIgnore()` — `INSERT OR IGNORE` → `ON CONFLICT DO NOTHING`
  - `BooleanLiterals()` — `= 0/1` → `= FALSE/TRUE` for known boolean columns
- `database.DB` wraps `*sql.DB` with shadowed `Exec()`, `Query()`, `QueryRow()` that auto-rewrite SQL.
- `InsertReturningID()` and `TxInsertReturningID()` replace `LastInsertId()` (PostgreSQL uses `RETURNING id`).
- `database.WrapDB(sqlDB, DialectSQLite)` for unit tests (in-memory SQLite).
- Migrations in `database/migrations/` — separate SQLite and PostgreSQL variants per migration.
- SQLCipher imported for encrypted SQLite.
- **SQLite WAL mode**: explicit `PRAGMA journal_mode=WAL` after connection in `database/connection.go` — go-sqlcipher ignores connection-string pragmas.

### catalog-web (React / TypeScript / Vite)

`AuthProvider → WebSocketProvider → Router`. React Query (server state), Zustand (client state), Tailwind CSS, React Hook Form + Zod, framer-motion, Vitest, Playwright.

- Auth-gated routes via `ProtectedRoute`.
- Path aliases in `vite.config.ts`: `@/components`, `@/hooks`, `@/lib`, `@/types`, `@/services`, `@/store`, `@/pages`, `@/assets`.
- API proxy reads `../catalog-api/.service-port` at dev-server startup (falls back to 8080).
- Build chunks: `vendor` (react), `router`, `ui`, `charts`, `utils`.

### Other Components

**Android**: MVVM — Compose UI → ViewModel (StateFlow) → Repository → Room + Retrofit. Hilt DI. Requires `jvmToolchain(17)` and `--add-opens` JVM args for kapt + JDK 21 compat.

**Android TV Home Screen Channels** (catalogizer-androidtv v2.3.0): full integration with `androidx.tvprovider`. Default "Catalogizer Picks" channel auto-created on launch. Per-category dynamic channels. System Watch Next row for partially-watched items + auto-next-episode. Deep linking via `catalogizer://media/{id}?type={type}` with per-category launch behavior in Settings. `WorkManager` periodic sync (6h) + app-launch + SyncService triggers. Full cleanup on logout. Files: `data/tv/TvChannelRepository.kt`, `data/tv/ChannelProgramMapper.kt`, `data/tv/WatchNextManager.kt`, `data/tv/TvChannelSyncWorker.kt`, `ui/ChannelDeepLinkActivity.kt`.

**Tauri apps**: React frontend ↔ Rust backend via IPC commands/events.

### Build Framework

`Build/` submodule provides a generic shell-based build framework. `scripts/release-build.sh` orchestrates all 7 components using per-component builders in `scripts/lib/build-*.sh`.

```bash
source Build/lib/common.sh        # logging, container runtime detection, git
source Build/lib/version.sh       # semantic versioning via versions.json
source Build/lib/hash.sh          # SHA256 change detection (skip unchanged components)
source Build/lib/orchestrator.sh  # CLI parsing, build loop
```

Projects must define `BUILD_COMPONENTS`, `BUILD_COMPONENT_PATTERNS`, and `build_single_component()`.

### Challenge System

`digital.vasic.challenges` framework via `Challenges/` submodule. Challenges are Go structs embedding `challenge.BaseChallenge` with a custom `Execute()`. Registered in `catalog-api/challenges/register.go` via `RegisterAll()`. Exposed via `/api/v1/challenges` REST endpoints. Bank definitions in `challenges/config/`.

**All challenge operations must be executed exclusively by system deliverables (the catalog-api binary and other Catalogizer apps).** Never use shell scripts, curl, or third-party tools to drive API endpoints during challenge execution. Scanning, storage-root creation, and every other operation must go through the running services exactly as an end user would.

Constraints:
- `RunAll` is synchronous / blocking — no other challenge can run until it finishes.
- Progress-based liveness: 5-minute stale threshold kills stuck challenges.
- `challenge.NewConfig()` defaults `Timeout=5min` — zero it to use the runner's timeout.
- `config.json` `write_timeout` must be 900 (not 30) for long-running RunAll.

### User Flow Automation

Multi-platform automation via `Challenges/pkg/userflow/`. 174 Catalogizer-specific challenges in `catalog-api/challenges/userflow_*.go`:

| File | Platform | Challenges |
|------|----------|-----------|
| `userflow_api.go` | Go API (HTTP) | 49 |
| `userflow_web.go` | React web (Playwright) | 59 |
| `userflow_desktop.go` | Tauri desktop + wizard | 28 |
| `userflow_mobile.go` | Android + Android TV | 38 |

Registered via `RegisterUserFlow{API,Web,Desktop,Mobile}Challenges()` in `register.go`. CLI runner: `Challenges/cmd/userflow-runner` — flags `--platform`, `--report`, `--compose`, `--root`, `--timeout`, `--output`. Container test stack: `docker-compose.test.yml` (`network_mode: host`).

## Media Entity System

Scanned files become structured entities via a post-scan aggregation pipeline:

```
UniversalScanner (scan completes)
       ↓ (post-scan hook)
AggregationService.AggregateAfterScan()
  ├── Title parser (regex: movie, TV, music, game, software)
  ├── MediaItem creation/update (media_items)
  ├── MediaFile linking (media_files junction)
  ├── Hierarchy builder (TV: show→season→episode, Music: artist→album→song)
  └── Duplicate detection (same title + type + year)
       ↓
Entity API (/api/v1/entities) → Entity Browser UI (/browse, /entity/:id)
```

**11 media types** seeded in `media_types`: movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic.

**Entity tables**: media_types, media_items (parent_id self-ref), media_files (junction), media_collections, media_collection_items, external_metadata, user_metadata, directory_analyses, detection_rules.

API routes: `handlers/media_entity_handler.go`. Key files: `repository/media_item_repository.go` (CRUD, search, hierarchy), `internal/services/aggregation_service.go` (post-scan), `internal/services/title_parser.go` (regex parsers).

**Metadata providers**: OpenLibrary (books) and MusicBrainz (music) fully implemented; TMDB and OMDB cover movies/TV. Other providers (IGDB, GiantBomb, …) degrade gracefully — missing keys never block the pipeline.

All scanned files must be associated with a recognized entity after aggregation.

## Container Runtime (Podman Only)

Project uses Podman exclusively (no Docker). All commands use `podman` / `podman-compose`.

```bash
podman-compose -f docker-compose.dev.yml up         # dev env
podman-compose -f docker-compose.yml config --quiet # validate
```

Critical notes:
- `podman build --network host` and `podman run --network host` — default container networking has SSL issues with dl.google.com, crates.io, etc.
- `GOTOOLCHAIN=local` to prevent Go auto-downloading newer toolchain versions.
- Use fully qualified image names (`docker.io/library/...`) — short names fail without TTY.
- `APPIMAGE_EXTRACT_AND_RUN=1` in containers for Tauri AppImage bundling (no FUSE).
- catalog-api container needs `--add-host=synology.local:192.168.0.241` for NAS access.
- Builder image: `localhost/catalogizer-builder:latest`. Rebuild with `podman build -f docker/Dockerfile.builder -t catalogizer-builder:latest .`

### Android APK Builds

All APK builds must use the `catalogizer-builder` container.

```bash
# Builder infrastructure (PostgreSQL, Redis, builder)
cd Containers && ./bin/boot --project /path/to/catalogizer

# Or directly via compose
podman-compose -f docker-compose.build.yml up --build --abort-on-container-exit

# Direct one-off build
podman run --rm --entrypoint="" \
  -v /path/to/project:/project \
  -w /project/catalogizer-androidtv \
  -e ANDROID_HOME=/opt/android-sdk \
  -e JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64 \
  localhost/catalogizer-builder:latest \
  ./gradlew assembleDebug --no-daemon
```

### Docker Compose Files

| File | Purpose |
|---|---|
| `docker-compose.yml` | Production stack |
| `docker-compose.dev.yml` | Development environment |
| `docker-compose.build.yml` | Containerized build pipeline |
| `docker-compose.test.yml` | Test stack (`network_mode: host`) |
| `docker-compose.test-infra.yml` | Test infra services |
| `docker-compose.security.yml` | Security scanning tools |
| `docker-compose.qa.yml` | QA environment |
| `docker-compose.qa-robot.yml` | QA robot configuration |

## Root Directory Structure

New files **must** go in the correct directory. Do not add files to the project root unless they are conventional root files (README, LICENSE, .gitignore, docker-compose, …).

| Directory | Purpose |
|---|---|
| `challenges/` | Challenge bank definitions and runtime results |
| `config/` | Infrastructure config (nginx.conf, redis.conf) — Compose mounts these |
| `scripts/` | Shell scripts (install, setup, CI/CD, test runners) |
| `scripts/lib/` | Per-component build scripts (`build-*.sh`) |
| `tests/` | Standalone / integration test files |
| `docs/` | All markdown documentation, organized by subdirectory |
| `Assets/` | Static assets (also a Go submodule) |
| `Build/` | Generic build framework submodule |
| `build/` | Build output and container build context |
| `deployment/` | Deployment configurations |
| `monitoring/` | Monitoring and observability configs |
| `tools/` | Development tooling |
| `Upstreams/` | Git upstream remote configurations for submodules |

## Local Development Setup

### Database

- **SQLite (dev)**: no setup — catalog-api creates `catalogizer.db` automatically.
- **PostgreSQL (production)**: set `DB_TYPE=postgres`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`. Container port mapping 5432 → 5433.

### Environment Variables

Create `.env` in `catalog-api/`. **Env vars override `config.json`.**

```env
PORT=8080
GIN_MODE=debug
DB_TYPE=sqlite
JWT_SECRET=your-dev-secret-key
ADMIN_PASSWORD=admin123
TMDB_API_KEY=your_tmdb_key     # optional
OMDB_API_KEY=your_omdb_key     # optional
```

### Running the Full Stack

```bash
ss -tlnp | grep :3000                            # kill anything on 3000 first (e.g. Bear Messenger)

cd catalog-api && go run main.go                 # Terminal 1 — backend writes .service-port
cd catalog-web && npm install && npm run dev     # Terminal 2 — frontend reads .service-port
# Access: http://localhost:3000 (frontend) / http://localhost:8080 (API)
```

## Concurrency & Lifecycle Patterns

- **CacheService**: spawns cleanup goroutine in `NewCacheService()`. Cleanup context respects shutdown signal (cancels in-flight cleanup on `Close()`). Tests must `defer service.Close()`. Uses `sync.Once` for safe double-close.
- **WebSocketHandler**: spawns cleanup goroutine in constructor. Uses `sync.Once` for safe `Stop()`. Connection admission uses reservation pattern (pre-increment connCount, decrement on upgrade failure) to prevent race between capacity check and registration. Tests must call `handler.Stop()` before `server.Close()` to unblock `readPump`.
- **WorkerPool**: `SubmitAsync()` goroutines are tracked via WaitGroup -- `Stop()` waits for all in-flight submissions.
- **Throttler**: run goroutine is tracked via WaitGroup. `Stop()` uses `CompareAndSwap` for safe double-close and waits for goroutine exit.
- **Debouncer**: `Flush()` copies the pending function and clears it under the lock, then executes outside the lock to prevent deadlock if the function calls `Debounce()` recursively.
- **SMBChangeWatcher**: `Stop()` drains all pending debounce timers before closing stop channel to prevent `wg.Add(1)` after `wg.Wait()` has started.
- **Production shutdown**: `main.go` calls `wsHandler.Stop()` and `cacheService.Close()` before HTTP server shutdown.
- **Database pool defaults**: MaxOpen=25, MaxIdle=10, MaxLifetime=5m, MaxIdleTime=3m. Overridable via config.
- **Race safety**: `connCount` reads in WebSocketHandler protected by mutex. `SyncService.StartSync()` and `LogManagementService.CollectLogs()` return copies to prevent shared-pointer races.

## Load Testing

k6 scripts in `tests/k6/`:
- `load_test.js` — ramp to 50 users, verify p95 < 500 ms
- `stress_test.js` — ramp to 300 users, find breaking point
- `soak_test.js` — 20 users for 30 min, detect memory leaks

```bash
podman run --rm --network host -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest run /scripts/load_test.js
```

## Security Scanning

`./scripts/security-scan.sh` runs all of: `govulncheck`, `npm audit`, Semgrep (custom rules in `.semgrep.yml` + auto), Snyk, Trivy, Gosec, Hadolint.

```bash
./scripts/security-scan.sh                                     # all tools
podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner  # Semgrep only
podman-compose -f docker-compose.security.yml --profile hadolint run --rm hadolint             # Dockerfile lint
podman-compose -f docker-compose.security.yml --profile sonarqube up                           # SonarQube (profile-gated)
```

## Git

6 push targets on `origin` (2× GitHub, 2× GitLab, GitFlic, GitVerse). GitVerse uses port 2222.

```bash
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
ssh-keyscan github.com gitlab.com gitflic.ru >> ~/.ssh/known_hosts
ssh-keyscan -p 2222 gitverse.ru >> ~/.ssh/known_hosts
```

`releases/` and `reports/` are gitignored — build artifacts are not version-controlled.

## Conventions

**Config precedence**: env vars > `.env` > `config.json` > defaults

**PostCSS**: `postcss.config.js` must use `module.exports` (CommonJS) for Node 18 compat.

### Go Backend Style

- **Naming**: PascalCase exported, camelCase unexported. Interfaces: `Reader`, `Writer`, `Service` suffixes.
- **Receivers**: single-letter (`s *Service`, `h *Handler`, `r *Repository`).
- **Imports**: three groups separated by blank lines — stdlib, third-party, local:
  ```go
  import (
      "encoding/json"
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/stretchr/testify/assert"

      "catalogizer/database"
      "catalogizer/models"
  )
  ```
- **Constructors**: `NewService(dep Dependency) *Service` with DI.
- **Errors**: wrap with `fmt.Errorf("context: %w", err)`. Use `errors.New` for static. Never expose internal details to clients.
- **Testing**: table-driven with `t.Run`. `testify/suite` for complex suites, `testify/mock` for mocks. `*_test.go` beside source. Use `database.WrapDB()` for in-memory SQLite.
- **Concurrency**: services spawning goroutines use `sync.Once` for cleanup; tests must `defer service.Close()` / `handler.Stop()`.
- **Database**: `?` placeholders (auto-converted to `$1, $2…`); use `InsertReturningID()` instead of `LastInsertId()`.

### TypeScript / React Frontend Style

- **Naming**: PascalCase components/interfaces, camelCase functions/variables, SCREAMING_SNAKE_CASE constants.
- **Imports**: three groups — React, third-party, local path aliases.
- **Formatting**: Prettier. Tailwind classes via `cn()` from `@/lib/utils`.
- **Linting**: ESLint with `@typescript-eslint`, `--max-warnings 0` enforced.
- **State**: React Query (server), Zustand (client).
- **Forms**: React Hook Form + Zod.
- **Testing**: Vitest + React Testing Library; Playwright for E2E.

### Kotlin / Android Style

- **Architecture**: MVVM — Compose UI → ViewModel (StateFlow) → Repository → Room + Retrofit.
- **DI**: Hilt.
- **Async**: `suspend` functions, `Flow` / `StateFlow`, Paging 3.
- **Errors**: sealed `Result` classes for operation outcomes.
- **Testing**: JUnit 4 + MockK / Mockito; coroutines via `kotlinx-coroutines-test`.
- **Build**: JDK 21 with `--add-opens` JVM args for kapt compatibility.

## Key Files

- `catalog-api/main.go` — API entry, route registration
- `catalog-api/database/dialect.go` — dual-dialect SQL rewriting
- `catalog-api/database/connection.go` — explicit WAL pragma
- `catalog-api/filesystem/interface.go` — `UnifiedClient` protocol abstraction
- `catalog-api/challenges/register.go` — challenge registration
- `catalog-web/src/App.tsx` — React root (`AuthProvider → WebSocketProvider → Router`)
- `catalog-web/vite.config.ts` — path aliases, API proxy
- `versions.json` — version tracking for all components
- `.env.example` — environment variable template

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-platform media collection manager. Detects, categorizes, and organizes media across SMB, FTP, NFS, WebDAV, and local filesystems. Components: **catalog-api** (Go 1.25/Gin backend), **catalog-web** (React 18/TS/Vite frontend), **catalogizer-desktop** & **installer-wizard** (Tauri/Rust+React), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

## Companion Docs

- **`AGENTS.md`** — autonomous-agent constraints and coding conventions (read alongside this file).
- **`MEMORY.md`** — auto-memory index for persistent Claude session state.
- **`CONSTITUTION.md`** — non-negotiable program rules. Article V (100% test coverage), Article VI (Open-Points brief), Article VII (Full-QA Master Cycle).
- **`docs/OPEN_POINTS_CLOSURE.md`** — single source of truth for every operator-action item.
- **`README.md`** — extensive architecture narrative, provider details, deployment guidance.
- **`versions.json`** — authoritative version for every app/service.

## ⚠️ Mandatory Constraints (Non-Negotiable)

The full text of each constraint below lives in `CONSTITUTION.md`. These are invariants — do not paraphrase into code comments or other docs.

- **Full-QA Master Cycle** (Article VII). Clean rebuild → all tests → all Challenges → all HelixQA banks → autonomous QA per app/platform → video+screenshot review → tickets → root-cause fix with 4 artefacts (unit/integration test + `fixes-validation` entry + HelixQA bank entry + challenge) → rebuild → repeat until clean pass. Stop only on FATAL BLOCKER / SYSTEM BREAKS / NOTHING LEFT. Every session archives to `docs/reports/qa-sessions/<YYYY-MM-DD-THH-MM>/` (FINAL-REPORT.md, logs/, challenges/, helixqa/, videos/, screenshots/, tickets/, analysis/). On clean pass: version-bump + archive to `releases/<platform>/<app>/<version>/`.
- **100% Test Coverage Across 10 Categories** (Article V). Unit, integration, E2E, full automation, stress, security, DDoS/rate-limit, benchmarking, challenges, HelixQA. Every branch (happy / error / edge / adversarial) of every public function, endpoint, UI component. Shipping is prohibited while any category is incomplete or any ticket is open. Achieved sequentially, one platform at a time.
- **Open-Points Closure Brief** (Article VI). `docs/OPEN_POINTS_CLOSURE.md` is the single source of truth for operator-action items. Consult before any work; update in the **same commit** that changes an item's state (tick checkbox + refresh "Last refresh" date). Deleting an unclosed item is an Article VI violation.
- **Zero Unfinished Work.** No TODOs, FIXMEs, empty implementations, silent error swallows, fake data, panic-prone `unwrap()`, or empty catch blocks. Pre-commit hooks + CI enforce this. When you find one instance, fix **all** instances.
- **Zero Warning / Zero Error.** Every component runs with zero console warnings/errors and zero failed network requests in every environment. Missing features: return a valid empty stub response, not a 404/500.
- **No Sudo / No Root.** All operations local-user only: rootless Podman, user systemd, user-writable dirs. If something seems to need elevation, use containers via the Containers submodule (https://github.com/vasic-digital/Containers) or find a user-level alternative. Never use `sudo` or `su`.
- **API Keys & Secrets.** Never commit real `.env` files or hardcode secrets anywhere (source, CLAUDE.md, AGENTS.md). Use `.env.example` with `YOUR_API_KEY_HERE`. Verify `.gitignore` covers `.env` in the project root **and every submodule** before committing. Rotate leaked keys immediately.
- **Git via SSH Only.** `git@github.com:…`, never HTTPS. Submodules in `.gitmodules` use SSH URLs. GitVerse uses port 2222.
- **GitHub Actions Disabled.** Do not create files in `.github/workflows/`. CI/CD runs locally.
- **Containers Mandatory for Builds, Services, QA.** Bare-metal `go run` / `npm run dev` is acceptable for rapid local iteration only, never for production or QA. See [Container Runtime](#container-runtime-podman-only) for resource limits and flags.
- **HTTP/3 (QUIC) + Brotli.** Fallback HTTP/2 + gzip. Never HTTP/1.1 in production. catalog-api uses `quic-go/http3` + `andybalholm/brotli`; Android uses OkHttp + Cronet + Brotli.
- **`.devignore`** (e.g. ATMOSphere): never used for any testing/QA/deployment. Check device model before any ADB operation; abort if matched. Never fall back to an excluded device.
  ```bash
  DEVICE_MODEL=$(adb -s $DEVICE shell getprop ro.product.model)
  grep -qi "$DEVICE_MODEL" .devignore && { echo "❌ in .devignore"; exit 1; }
  ```
- **`.devconnect`** (gitignored): one IP per line for Android TV auto-connect. Run `./scripts/devconnect.sh` (idempotent) before HelixQA. IP lines MUST NOT carry inline `# comments` — trailing-comment gets concatenated into the device ID by the orchestrator's `head -1 | grep -v ^#` parser.
- **Device state preservation** (Constitution Article VIII). A QA session MUST NOT leave the device with changed `font_scale`, `wm density`, brightness, rotation, or any other `settings put …` value. HelixQA snapshots sensitive keys at session start and restores them via deferred cleanup — see `HelixQA/pkg/autonomous/device_preserve.go`. If a session leaves a device polluted, it's both a product bug (curiosity should not navigate into Settings → Accessibility) and a missing preservation entry to patch.
- **HelixQA tool hygiene** (Constitution Article IX). No manual `adb shell screenrecord` workaround scripts; no `tee`-style exit-code laundering; no "✓ completed successfully" reports gated by anything other than real pass assertions. If the autonomous pipeline produces broken output, fix the Go code.

## HelixQA: Autonomous LLM-Driven Testing

HelixQA is the **sole authorized tool** for all automated UI/UX testing across Android TV, Android phone, web, and desktop. Pipeline: **Learn → Plan → Execute → Curiosity → Analyze**. Configuration in `HelixQA/.env.example`. Full details in `HelixQA/README.md`.

### Invariants (must never be violated)

- **HelixQA-only for UI automation.** No custom ADB tap sequences, no coordinate-based shell scripts, no standalone Playwright/Appium outside HelixQA.
- **Vision-driven only.** Every action is `screenshot → LLM analysis → action decision`. No hardcoded coordinates, no sleep timers, no keystroke sequences, no fallback hardcoded actions. If vision providers are unavailable the phase **skips** — never fakes results.
- **Universal Solution Principle.** Fix detection / input / parsing bugs in **HelixQA itself**, never in the app under test. No `QAInputReceiver`, no test endpoints, no bypasses — solutions that modify the app are invalid.
- **Live log monitoring.** Every QA session streams `adb logcat` (Android), browser console (web), app/system logs (desktop), service logs (backend). ANR/crash/fatal-exception pauses the session immediately.
- **Screen-state tracking.** Compare frame N to N+1. Stagnation (identical >10s or no change after an action) is a critical failure — a "100% pass" report when the app never progressed past login is fraudulent.
- **Executable actions in banks**, never prose. `action: "adb_shell: input text admin"`, not `"Enter credentials"`.
- **Credential safety.** Never type credentials into non-login fields — the LLM must recognize the current screen.
- **Video mandatory for every device/emulator session.** 16 Mbps min, 1920×1080, pulled after recording; frames extracted for post-analysis (misaligned UI, clipped text, wrong colors, missing assets, unresponsive buttons, jank, brand compliance).
- **Evidence validation.** Post-login UI dump must not contain "Sign In". Empty screens when backend has data = bug. Compare API responses against on-screen content.
- **Validation tests are permanent.** Every fix adds a regression entry to `banks/fixes-validation.yaml`; it persists across all future campaigns.

### Vision Architecture

Phase-specific model selection via LLMsVerifier strategies — NavigationStrategy (Execute/Curiosity: JSON-action models), AnalysisStrategy (Analyze: rich-description models), PlanningStrategy (Learn/Plan: reasoning models). **llama.cpp RPC** is the primary local backend; Astica.AI / Gemini / OpenAI complement it. Models scored dynamically per phase; bridged CLI models discovered via `pkg/bridge/`.

### Bank Coverage & Execution

Banks live in `banks/full-qa-{api,web,androidtv,android,cross-platform}.yaml` + `banks/fixes-validation.yaml`. **Bank format is JSON at runtime** — convert with `python3 -c "import yaml,json; json.dump(yaml.safe_load(open('bank.yaml')), open('bank.json','w'))"`. Coverage spans all 11 media types, all CRUD, all screens, all auth/nav flows, adversarial search (Cyrillic, SQL injection, XSS), boundary values, and internationalized content.

```bash
./scripts/helixqa-orchestrator.sh [platforms]   # all platforms by default
./scripts/helixqa-orchestrator.sh androidtv     # one platform
```

Phases: env validation → device connect (`.devconnect`) → APK install (builds if needed) → background health monitoring → autonomous testing → report generation. Output: `qa-results/session-<timestamp>/`. Session archive layout is mandated by Article VII — see Full-QA Master Cycle.

### Platform-Specific Recording Notes

- **Android 9 and below** (Mi Box, emulators): `adb shell screenrecord --bit-rate 4000000 /sdcard/qa_session.mp4`.
- **Android 10+**: rapid `screencap` frames assembled into video via ffmpeg.
- **Android 15 (SDK 35)**: `screenrecord` fails with `Encoder failed (err=-38)` → use screenshot-to-video.
- **Web**: Playwright `--video on` or `ffmpeg x11grab`.
- **Desktop (Tauri)**: `ffmpeg x11grab` or `Xvfb`.

## Submodule Architecture

41 independent git submodules under the `vasic-digital` org. Each has its own GitHub + GitLab repos, tests, docs, and `Upstreams/` config for multi-remote push. Wiring:

- **Go modules** (`digital.vasic.*`) are wired into `catalog-api/go.mod` via `replace` directives — inspect `go.mod` for the authoritative list.
- **TypeScript/React packages** (`@vasic-digital/*`) are linked in `catalog-web/package.json` via `file:../` — inspect `package.json`.
- **HelixQA / AI stack**: `HelixQA/`, `DocProcessor/`, `LLMOrchestrator/`, `LLMProvider/`, `VisionEngine/`, `ReplayBuffer/`, `ScreenDiff/`, `TrainingCollector/`, `VisualRegression/`.

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

**Android TV Home Screen Channels** (`catalogizer-androidtv`; current version in `versions.json`): `androidx.tvprovider` integration. Default "Catalogizer Picks" channel auto-created on launch; per-category dynamic channels; system Watch Next row for partially-watched items + auto-next-episode. Deep linking via `catalogizer://media/{id}?type={type}`. `WorkManager` periodic sync (6h) + app-launch + SyncService triggers. Full cleanup on logout. Files: `data/tv/TvChannelRepository.kt`, `data/tv/ChannelProgramMapper.kt`, `data/tv/WatchNextManager.kt`, `data/tv/TvChannelSyncWorker.kt`, `ui/ChannelDeepLinkActivity.kt`.

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

Project uses Podman exclusively (no Docker). All commands use `podman` / `podman-compose`. Compose files at the project root (`docker-compose*.yml`) are self-describing — `podman-compose -f <file> config --quiet` validates a file without running it.

**Host resource limits (30–40% max).** The host runs other mission-critical processes. Exceeding the budget can freeze the system.

- **Go tests**: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- **Container flags** (mandatory on `podman run`):
  - PostgreSQL: `--cpus=1 --memory=2g`
  - catalog-api: `--cpus=2 --memory=4g`
  - catalog-web: `--cpus=1 --memory=2g`
  - Builder: `--cpus=3 --memory=8g`
- **Total budget**: 4 CPUs / 8 GB RAM across all running containers.
- **Challenges run sequentially** via the API — never in parallel.
- **Monitor**: `podman stats --no-stream`, `cat /proc/loadavg`.

**Networking & image hygiene:**
- `podman build --network host` / `podman run --network host` — default container networking has SSL issues with `dl.google.com`, `crates.io`, etc.
- `GOTOOLCHAIN=local` prevents Go from auto-downloading newer toolchains.
- Fully qualified image names (`docker.io/library/...`) — short names fail without TTY.
- `APPIMAGE_EXTRACT_AND_RUN=1` for Tauri AppImage bundling in containers (no FUSE).
- catalog-api container needs `--add-host=synology.local:192.168.0.241` for NAS access.

**Builder image**: `localhost/catalogizer-builder:latest`. Rebuild with `podman build -f docker/Dockerfile.builder -t catalogizer-builder:latest .`

### Android APK Builds (container only)

```bash
# Full builder infrastructure (PostgreSQL, Redis, builder)
cd Containers && ./bin/boot --project /path/to/catalogizer

# Or build via compose
podman-compose -f docker-compose.build.yml up --build --abort-on-container-exit

# Direct one-off
podman run --rm --entrypoint="" \
  -v /path/to/project:/project \
  -w /project/catalogizer-androidtv \
  -e ANDROID_HOME=/opt/android-sdk \
  -e JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64 \
  localhost/catalogizer-builder:latest \
  ./gradlew assembleDebug --no-daemon
```

## Directory Conventions

New files go in existing purpose-specific directories — do not add files to the project root unless they are conventional root files (README, LICENSE, `.gitignore`, compose files, etc.). Notable top-level dirs: `challenges/`, `config/`, `scripts/` (+ `scripts/lib/` for per-component build scripts), `tests/`, `docs/`, `Build/`, `build/`, `deployment/`, `monitoring/`, `tools/`, `Upstreams/`.

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

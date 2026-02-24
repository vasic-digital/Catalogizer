# Running User Flow Challenges

## Prerequisites

1. The catalog-api server must be running (for API and web challenges)
2. The catalog-web dev server must be running (for web challenges)
3. A Playwright container must be running with CDP exposed on port 9222 (for web challenges)
4. ADB must be available with a connected device or emulator (for Android challenges)
5. Tauri desktop binaries must be built (for desktop and wizard challenges)

## Method 1: Via API Endpoint

The simplest way to run user flow challenges is through the running catalog-api server.

### Run all user flow challenges (all platforms, sequential)

```bash
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow
```

### Run a single platform group

```bash
# API challenges only (49 challenges)
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow-api

# Web challenges only (59 challenges)
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow-web

# Desktop challenges only (18 challenges)
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow-desktop

# Wizard challenges only (10 challenges)
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow-wizard

# Android challenges only (22 challenges)
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow-android

# Android TV challenges only (16 challenges)
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow-androidtv
```

### Run a single challenge by ID

```bash
curl -X POST http://localhost:8080/api/v1/challenges/run/UF-API-AUTH-LOGIN
```

### Check challenge status

```bash
# Get status of all challenges
curl http://localhost:8080/api/v1/challenges

# Get status of a specific challenge
curl http://localhost:8080/api/v1/challenges/UF-API-AUTH-LOGIN

# Get challenge results
curl http://localhost:8080/api/v1/challenges/results
```

### Important: RunAll is blocking

The challenge runner executes synchronously. When you trigger a run via the API, no other challenge can execute until the current run completes. Monitor progress via:

```bash
curl http://localhost:8080/api/v1/challenges/progress
```

The stale threshold is 5 minutes -- if no progress is reported within 5 minutes, the stuck challenge is killed.

## Method 2: Standalone CLI Runner

For CI/CD or local development, use the standalone runner:

```bash
cd catalog-api

# Run all platforms
go run ./cmd/userflow-runner/ --platform=all --report=html

# Run a single platform
go run ./cmd/userflow-runner/ --platform=web --report=json

# Run with verbose output
go run ./cmd/userflow-runner/ --platform=api --verbose

# Generate HTML report
go run ./cmd/userflow-runner/ --platform=all --report=html --output=reports/userflow-report.html
```

### CLI flags

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--platform` | all, api, web, desktop, wizard, android, androidtv | all | Platform group to run |
| `--report` | json, html, text | text | Report output format |
| `--output` | file path | stdout | Write report to file |
| `--verbose` | (flag) | false | Verbose logging |
| `--timeout` | duration | 5m per challenge | Per-challenge timeout |
| `--parallel` | (flag) | false | Run independent challenges in parallel (within a platform) |

## Method 3: Containerized Execution

Use the test container stack for reproducible, isolated execution.

### Start the test stack

```bash
podman-compose -f docker-compose.test.yml up -d
```

### Run challenges inside the container

```bash
# Run via the API endpoint (catalog-api is running in the container)
curl -X POST http://localhost:8080/api/v1/challenges/run/category/userflow

# Or run the CLI runner in the catalog-api container
podman exec catalogizer-api go run ./cmd/userflow-runner/ --platform=all
```

### Stop the test stack

```bash
podman-compose -f docker-compose.test.yml down
```

See [container-setup.md](./container-setup.md) for full container configuration details.

## Environment Variables

All adapters are configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `BROWSING_API_URL` | `http://localhost:8080` | API server base URL |
| `ADMIN_USERNAME` | `admin` | Admin username for auth challenges |
| `ADMIN_PASSWORD` | (empty) | Admin password for auth challenges |
| `DESKTOP_PROJECT_ROOT` | `../catalogizer-desktop` | Desktop app project root |
| `DESKTOP_BINARY_PATH` | `../catalogizer-desktop/src-tauri/target/debug/catalogizer-desktop` | Desktop binary path |
| `WIZARD_PROJECT_ROOT` | `../installer-wizard` | Wizard app project root |
| `WIZARD_BINARY_PATH` | `../installer-wizard/src-tauri/target/debug/installer-wizard` | Wizard binary path |
| `ANDROID_PROJECT_ROOT` | `../catalogizer-android` | Android project root |
| `ANDROID_APK_PATH` | `../catalogizer-android/app/build/outputs/apk/debug/app-debug.apk` | Android APK path |
| `ANDROIDTV_PROJECT_ROOT` | `../catalogizer-androidtv` | Android TV project root |
| `ANDROIDTV_APK_PATH` | `../catalogizer-androidtv/app/build/outputs/apk/debug/app-debug.apk` | Android TV APK path |
| `ANDROID_DEVICE_SERIAL` | (empty) | ADB device serial for Android phone |
| `ANDROIDTV_DEVICE_SERIAL` | (empty) | ADB device serial for Android TV |

## Interpreting Results

Challenge results follow the standard Challenges module format:

```json
{
  "id": "UF-API-AUTH-LOGIN",
  "name": "API Auth Login",
  "status": "passed",
  "duration_ms": 245,
  "assertions": [
    {
      "type": "not_empty",
      "target": "auth_me_body",
      "passed": true,
      "message": "auth/me returns user data"
    }
  ]
}
```

Possible statuses: `passed`, `failed`, `skipped` (dependency not met), `stuck` (no progress for 5 minutes), `timed_out`.

## Resource Limits

Always respect the host resource budget when running tests:

```bash
# Go tests (if running alongside challenges)
GOMAXPROCS=3 go test ./... -p 2 -parallel 2

# Monitor resource usage
podman stats --no-stream
cat /proc/loadavg
```

Total container budget: max 4 CPUs, 8 GB RAM across all running containers. Platform groups run sequentially, never in parallel, to stay within this budget.

# Containerized Build System

## Overview

The Catalogizer containerized build system runs all tests and builds inside Docker/Podman containers, providing a reproducible, isolated environment with integration services (PostgreSQL, Redis).

### Architecture

```
docker-compose.build.yml
├── postgres              PostgreSQL 15 (integration test database)
├── redis                 Redis 7 (integration test cache)
├── catalogizer-builder   Ubuntu 22.04 multi-toolchain container
│   ├── Go 1.24          (catalog-api)
│   ├── Node.js 18       (catalog-web, desktop apps, api-client)
│   ├── Rust stable      (Tauri desktop/installer builds)
│   ├── JDK 17           (Android apps)
│   ├── Android SDK 34   (Android builds)
│   └── Playwright       (E2E browser tests)
└── android-emulator     (optional, requires KVM)
```

A **single builder container** is used because the `catalogizer-api-client` is a `file:` dependency of `installer-wizard`, requiring a shared filesystem. Shared caches (npm, Go modules, Gradle, Cargo) also work naturally in a single container.

## Quick Start

```bash
# Default build (v1.0.0)
./scripts/container-build.sh

# Build with specific version
./scripts/container-build.sh 2.0.0

# Skip E2E tests
./scripts/container-build.sh 1.0.0 --skip-e2e

# Enable Android emulator tests (requires KVM)
./scripts/container-build.sh 1.0.0 --with-emulator
```

## Prerequisites

- **Podman** (preferred) or **Docker** with compose support
- Podman: `podman` + `podman-compose` (`pip3 install podman-compose`)
- Docker: `docker` + `docker-compose` or `docker compose` plugin
- **KVM** (optional): Required only for Android emulator testing (`/dev/kvm`)
- **Disk space**: ~10GB for builder image, ~5GB for caches

## Build Phases

The pipeline runs 10 sequential phases inside the builder container:

| Phase | Description | Outputs |
|-------|-------------|---------|
| 0 | Generate signing keys | `docker/signing/catalogizer-debug.keystore` |
| 1 | Infrastructure health checks | Validates PostgreSQL + Redis connectivity |
| 2 | API Client (test + build) | `catalogizer-api-client/dist/` |
| 3 | Backend API (test + build) | Go binaries for Linux/Windows/macOS |
| 4 | Web Frontend (test + build) | Production static files |
| 5 | Desktop Apps (test + build) | .deb, .AppImage, binary |
| 6 | Android Apps (test + build) | Signed APKs |
| 7 | Emulator smoke tests | App launch verification (optional) |
| 8 | Coverage validation | `reports/validation-report.json` |
| 9 | Artifact collection | `releases/MANIFEST.json`, `SHA256SUMS.txt` |
| 10 | Report generation | `reports/build-report.html` |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BUILD_VERSION` | `1.0.0` | Version string for built artifacts |
| `SKIP_EMULATOR_TESTS` | `true` | Skip Android emulator smoke tests |
| `SKIP_E2E_TESTS` | `false` | Skip Playwright E2E browser tests |
| `SKIP_SECURITY_TESTS` | `true` | Skip security scanning (SonarQube, Snyk) |
| `POSTGRES_HOST` | `postgres` | PostgreSQL host (set by compose) |
| `REDIS_HOST` | `redis` | Redis host (set by compose) |

## Output Structure

```
releases/
├── MANIFEST.json                    Build metadata
├── SHA256SUMS.txt                   Integrity checksums
├── linux/
│   ├── catalog-api/                 Linux AMD64 binary
│   ├── catalog-web/                 Static web files
│   ├── catalogizer-desktop/         .deb, .AppImage, binary
│   └── installer-wizard/            .deb, .AppImage, binary
├── windows/
│   └── catalog-api/                 Windows AMD64 .exe
├── macos/
│   ├── amd64/catalog-api/           macOS Intel binary
│   └── arm64/catalog-api/           macOS Apple Silicon binary
├── android/
│   ├── catalogizer-android/         Signed APK
│   └── catalogizer-androidtv/       Signed APK
└── lib/
    └── catalogizer-api-client/      TypeScript library

reports/
├── build-report.html                Build summary report
├── build-report.json                Machine-readable build report
├── validation-report.html           Coverage validation report
├── validation-report.json           Machine-readable validation
├── go-coverage.html                 Go coverage HTML
├── go-coverage.out                  Go coverage data
└── *-test.log                       Per-component test logs
```

## Signing Keys

The build system generates debug/CI signing keys automatically. These are **not suitable for production releases**.

**Generated files:**
- `docker/signing/catalogizer-debug.keystore` - Android debug keystore
- `docker/signing/signing.properties` - Gradle signing configuration

**For production releases:**
1. Replace the keystore with your production signing key
2. Update `signing.properties` with production credentials
3. Or set environment variables to override paths

The `.gitignore` already excludes `*.keystore` and `*.jks` files.

## Coverage Validation

The `scripts/validate-coverage.sh` script performs:

1. **Go coverage parsing** - Reads `coverage.out`, reports per-package coverage
2. **JS/TS coverage parsing** - Reads Vitest/Jest coverage summaries
3. **Android/JaCoCo parsing** - Reads XML reports for instruction coverage
4. **Assertion density check** - Flags test files with no assertions
5. **Artifact smoke validation** - Verifies built artifacts are valid executables/archives
6. **Skipped test detection** - Scans logs for silently skipped tests

## Caching

Named volumes persist build caches between runs:

| Volume | Purpose |
|--------|---------|
| `go-cache` | Go module cache (`/root/go`) |
| `gradle-cache` | Gradle dependencies (`/root/.gradle`) |
| `npm-cache` | npm cache (`/root/.npm`) |
| `cargo-cache` | Rust/Cargo registry (`/root/.cargo/registry`) |

To clear caches:
```bash
podman volume rm catalogizer_go-cache catalogizer_gradle-cache catalogizer_npm-cache catalogizer_cargo-cache
```

## Troubleshooting

### Builder image fails to build

Tauri dependencies (webkit2gtk) require Ubuntu 22.04. If the base image pull fails, check your network/proxy settings.

### Android SDK download fails

The SDK cmdline-tools download URL may change. Update the URL in `docker/Dockerfile.builder` if needed.

### Tests fail with PostgreSQL connection errors

Ensure the `postgres` service is healthy before the builder starts. The `depends_on` with `condition: service_healthy` handles this automatically.

### Desktop (Tauri) build fails

Tauri requires system libraries (webkit2gtk, libappindicator). These are installed in the Dockerfile. If builds fail, check the Tauri build log in `reports/desktop-build.log`.

### Out of disk space

The builder image is large (~5GB). Clear Docker/Podman cache:
```bash
podman system prune -a    # or: docker system prune -a
```

### Emulator tests fail

Android emulator requires KVM. Verify with:
```bash
ls -la /dev/kvm
```
If KVM is not available, use `--skip-emulator` (the default).

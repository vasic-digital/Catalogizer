# ADR-005: Container-Only Build and Runtime Policy

## Status
Accepted (2026-02-23)

## Context

Catalogizer targets 7 platforms (Go backend, React web, Tauri desktop, Tauri installer wizard, Android, Android TV, TypeScript API client) with complex build dependencies:

- Go 1.24 for the backend
- Node 18 for the web frontend and TypeScript libraries
- Rust toolchain for Tauri desktop and installer wizard apps
- JDK 17 + Android SDK 34 for Android and Android TV apps
- Playwright for E2E tests

Setting up this toolchain on a developer machine is error-prone, version-sensitive, and non-reproducible. Different developers with different OS versions, package manager states, or library versions produce different build outputs. CI/CD environments compound this problem since GitHub Actions are permanently disabled for this project.

Additionally, the host machine runs other mission-critical processes, and build workloads must be constrained to 30-40% of total host resources to prevent system freezes.

A separate but related concern: the project discovered through hard experience that Podman's default container networking has SSL issues with external registries (dl.google.com, crates.io, etc.), and that several container-specific workarounds are required for reliable builds.

## Decision

All builds and services use containers exclusively, with Podman as the sole container runtime (no Docker). The containerized build pipeline ensures reproducible builds across all 7 components, and container resource limits enforce the 30-40% host resource constraint.

### Container Runtime: Podman

Podman is chosen over Docker because:
- Rootless containers by default (no daemon running as root)
- Daemonless architecture (no single point of failure)
- Compatible with Docker CLI commands and Dockerfiles
- Available on the target Linux platform (Podman 5.7.1 + podman-compose 1.5.0)

### Build Pipeline

The builder image (4.82 GB, Ubuntu 22.04) contains all 7 toolchains pre-installed. Builds are triggered via `./scripts/release-build.sh --container --force --skip-tests` or `podman-compose -f docker-compose.build.yml`.

### Critical Container Constraints

These constraints were discovered through debugging real build failures:

1. **`podman build --network host`**: Default container networking causes SSL certificate verification failures when downloading dependencies from Google (dl.google.com for Go toolchain, Android SDK) and Rust registries (crates.io). Using `--network host` bypasses the container's network namespace and uses the host's DNS and routing directly.

2. **`podman run --network host`**: Same SSL issue applies at runtime for single-container builds. Do NOT use docker-compose for builds involving network-dependent operations like dependency downloads; the orchestrated networking adds another layer of potential failure.

3. **`GOTOOLCHAIN=local`**: Without this environment variable, the Go toolchain inside the container auto-downloads newer toolchain versions when it detects a `go.mod` requiring a newer patch version. This fails in restricted network environments and produces non-reproducible builds.

4. **Fully qualified image names**: Podman without a TTY (e.g., in scripts, CI) does not support short image names like `golang:1.24`. Use `docker.io/library/golang:1.24` instead.

5. **`APPIMAGE_EXTRACT_AND_RUN=1`**: Tauri's AppImage bundling requires FUSE, which is not available inside containers. This environment variable tells the AppImage tools to extract and run directly instead.

6. **`postcss.config.js` must use CommonJS**: Node 18 inside the container requires `module.exports` syntax, not `export default` (ESM).

7. **Go tarball pre-download**: The Go toolchain tarball must be pre-downloaded and `COPY`'d into the Docker image because network issues with Google URLs are intermittent and unreliable during image builds.

### Resource Limits

Container resource limits are mandatory and enforced via CLI flags:

| Container | CPU Limit | Memory Limit |
|-----------|-----------|--------------|
| PostgreSQL | 1 CPU | 2 GB |
| API Server | 2 CPUs | 4 GB |
| Web Frontend | 1 CPU | 2 GB |
| Builder | 3 CPUs | 8 GB |

**Total container budget**: Maximum 4 CPUs and 8 GB RAM across all running containers at any time.

**Go test limits**: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2` constrains test parallelism.

**Monitoring**: `podman stats --no-stream` and `cat /proc/loadavg` are used to verify compliance.

### Service Orchestration

Development services are managed via `podman-compose`:

```bash
podman-compose -f docker-compose.dev.yml up       # Development environment
podman-compose -f docker-compose.yml config --quiet  # Validate configuration
```

Individual containers use `podman run` with `--network host` for network-dependent operations.

## Consequences

### Positive

- **Reproducible builds**: Every developer and build environment produces identical artifacts from identical source, eliminating "works on my machine" issues.
- **Single builder image**: All 7 toolchains in one image (4.82 GB) means no per-platform build environment maintenance.
- **Resource safety**: Container CPU and memory limits prevent build workloads from starving the host system's mission-critical processes.
- **No Docker dependency**: Podman's rootless, daemonless architecture is more secure and eliminates Docker daemon maintenance.
- **Build output**: All 7 components build successfully in approximately 17 minutes, which is acceptable for release builds.

### Negative

- **Large builder image**: The 4.82 GB builder image takes significant time and disk space. Layer caching mitigates rebuild time but initial pull is slow.
- **Container networking complexity**: The `--network host` requirement and SSL workarounds add operational knowledge that must be documented and maintained.
- **No native development builds**: Developers who want fast iteration must still use local toolchains for development (`go run`, `npm run dev`). The container-only policy applies to release builds and CI, not the development inner loop.
- **Podman-specific knowledge**: The project requires Podman-specific flags and workarounds that differ from Docker, which may be unfamiliar to developers accustomed to Docker.
- **Resource constraints slow builds**: The 30-40% resource limit means builds take longer than they would on a dedicated build machine. This is an intentional trade-off for host stability.

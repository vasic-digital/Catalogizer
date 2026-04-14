# Distributed Build Guide

## Overview

The distributed build system extends Catalogizer's build pipeline to distribute component builds across multiple hosts based on real-time resource availability. It uses the Containers submodule's scheduler, SSH executor, and remote execution infrastructure.

## Architecture

```
release-build.sh --distributed
       |
       v
Go binary (Containers/cmd/distributed-build)
  1. Load host config from .env or env vars
  2. Probe all hosts for CPU/Memory/Disk usage
  3. Schedule components via resource-aware algorithm
  4. Sync source code to remote hosts via SCP
  5. Launch builder containers on each assigned host
  6. Collect build artifacts back to local releases/
```

## Host Requirements

Each build host must have:
- Passwordless SSH access from the build orchestrator
- Podman or Docker installed
- Go 1.25+, Node.js 18+, JDK 21 (for Android builds)
- At least 4GB free disk space in `/tmp`
- Network access to the project directory (for source sync)

## Configuration

### Environment Variables

Add to `.env` or set as environment variables:

```env
CONTAINERS_REMOTE_HOST_1_NAME=thinker
CONTAINERS_REMOTE_HOST_1_ADDRESS=thinker.local
CONTAINERS_REMOTE_HOST_1_USER=milosvasic
CONTAINERS_REMOTE_HOST_1_KEY_PATH=~/.ssh/id_ed25519
CONTAINERS_REMOTE_HOST_1_RUNTIME=podman
CONTAINERS_REMOTE_HOST_1_LABELS=go=true,npm=true,jdk=true,rust=true
```

Hosts are numbered 1-100. Discovery stops at the first gap.

### Labels

Labels control which hosts can build which components:
- `go=true` — Required for catalog-api
- `npm=true` — Required for catalog-web, desktop, installer, API client
- `jdk=true` — Required for Android and Android TV
- `rust=true` — Required for desktop (Tauri) and installer

### Scheduling Strategies

| Strategy | Description |
|----------|-------------|
| `resource_aware` (default) | Picks host with best CPU/Memory score |
| `round_robin` | Rotates through hosts evenly |
| `spread` | Picks host with fewest active builds |
| `bin_pack` | Fills most-used host first |

## Usage

### Check host prerequisites

```bash
./scripts/prepare-build-host.sh milosvasic@thinker.local
```

### Dry run (show plan only)

```bash
go run Containers/cmd/distributed-build --project . --dry-run
```

### Build all components distributed

```bash
go run Containers/cmd/distributed-build --project . --force --skip-tests
```

### Build single component on best available host

```bash
go run Containers/cmd/distributed-build --project . --component catalog-api --force
```

### Use specific strategy

```bash
go run Containers/cmd/distributed-build --project . --strategy spread --force
```

## Resource Allocation

The scheduler respects a 30-40% resource limit per host:

| Component | CPU | Memory | Disk |
|-----------|-----|--------|------|
| catalog-api | 2 cores | 2 GB | 1 GB |
| catalog-web | 1 core | 1 GB | 512 MB |
| catalogizer-android | 3 cores | 4 GB | 2 GB |
| catalogizer-androidtv | 3 cores | 4 GB | 2 GB |
| catalogizer-desktop | 3 cores | 4 GB | 2 GB |
| installer-wizard | 3 cores | 4 GB | 2 GB |
| catalogizer-api-client | 1 core | 1 GB | 512 MB |

## Troubleshooting

### Host unreachable

Check SSH connectivity:
```bash
ssh thinker.local echo ok
```

### Build timeout

Increase with `--timeout` (minutes):
```bash
go run Containers/cmd/distributed-build --project . --timeout 60 --force
```

### Source sync failures

Ensure rsync is installed on both hosts:
```bash
rsync --version
```

# Build - Generic Build Framework

Reusable, containerized build system with automatic versioning, change detection, and multi-component orchestration.

## Features

- **Semantic versioning** with build numbers via `versions.json`
- **Change detection** using SHA256 source hashes (skip unchanged components)
- **Container runtime detection** (Podman/Docker)
- **Artifact generation** with BUILD_INFO.json and SHA256SUM checksums
- **CLI interface** with `--dry-run`, `--force`, `--component`, `--bump`, etc.
- **Generic framework** - adapts to any multi-component project

## Quick Start

### 1. Add to your project

Copy or add as a submodule:
```bash
git submodule add git@github.com:vasic-digital/Build.git Build
```

### 2. Create project config

Create a project-specific config file that defines your components:

```bash
# scripts/lib/project-config.sh
BUILD_PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

BUILD_COMPONENTS=(
    "my-api"
    "my-web"
    "my-mobile"
)

declare -A BUILD_COMPONENT_PATTERNS=(
    ["my-api"]="*.go go.mod go.sum"
    ["my-web"]="*.ts *.tsx *.js *.json *.html *.css"
    ["my-mobile"]="*.kt *.java *.xml *.gradle.kts *.properties"
)
```

### 3. Implement component builders

```bash
# scripts/lib/build-my-api.sh
build_my_api() {
    local version="$1" build_number="$2" version_string="$3" source_hash="$4"
    local release_dir
    release_dir="$(create_release_dir "my-api" "linux-amd64" "$version_string")"

    cd "$BUILD_PROJECT_ROOT/my-api"
    go build -ldflags "-X main.Version=$version" -o "$release_dir/my-api" .

    generate_checksums "$release_dir"
    generate_build_info "$release_dir" "my-api" "linux-amd64" \
        "$version" "$build_number" "$version_string" "$source_hash"
}
```

### 4. Create entry point

```bash
#!/usr/bin/env bash
# scripts/release-build.sh
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Source project config
source "$SCRIPT_DIR/lib/project-config.sh"

# Source Build framework
source "$BUILD_PROJECT_ROOT/Build/lib/common.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/version.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/hash.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/orchestrator.sh"

# Source component builders
source "$SCRIPT_DIR/lib/build-my-api.sh"

# Dispatch function called by orchestrator
build_single_component() {
    local component="$1" version="$2" build_number="$3"
    local version_string="$4" source_hash="$5"
    case "$component" in
        my-api) build_my_api "$version" "$build_number" "$version_string" "$source_hash" ;;
        *) log_error "No builder for: $component"; return 1 ;;
    esac
}

build_main "$@"
```

### 5. Run

```bash
./scripts/release-build.sh                         # Build changed components
./scripts/release-build.sh --force                  # Rebuild everything
./scripts/release-build.sh --component my-api       # Build one component
./scripts/release-build.sh --dry-run                # Preview
./scripts/release-build.sh --bump patch             # Bump version
./scripts/release-build.sh --status                 # Show change status
```

## Framework Files

| File | Purpose |
|------|---------|
| `lib/common.sh` | Logging, runtime detection, git helpers, artifact generation |
| `lib/version.sh` | Semantic version management via `versions.json` |
| `lib/hash.sh` | Source hash computation and change detection |
| `lib/orchestrator.sh` | CLI parsing, build loop, component dispatching |

## versions.json Format

```json
{
  "schema_version": 1,
  "global": {
    "major": 1, "minor": 0, "patch": 0,
    "build_number": 5
  },
  "components": {
    "my-api": {
      "last_build_number": 5,
      "last_build_date": "2026-02-17T14:30:00Z",
      "last_source_hash": "a1b2c3d4...",
      "last_git_commit": "c5a4a4b7"
    }
  }
}
```

## Requirements

- Bash 4.0+ (for associative arrays)
- Python 3 (for JSON manipulation)
- sha256sum (coreutils)
- Git

## License

MIT

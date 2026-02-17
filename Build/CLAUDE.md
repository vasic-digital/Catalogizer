# CLAUDE.md - Build Module

## Overview

`digital.vasic.build` is a generic, reusable shell-based build framework providing automatic versioning, change detection, and multi-component build orchestration.

**Type**: Shell (Bash 4.0+)

## Usage

```bash
# Source the framework in your build script
source Build/lib/common.sh
source Build/lib/version.sh
source Build/lib/hash.sh
source Build/lib/orchestrator.sh
```

## Framework Structure

| File | Purpose |
|------|---------|
| `lib/common.sh` | Logging, container runtime detection (Podman/Docker), git helpers, artifact generation |
| `lib/version.sh` | Semantic versioning via `versions.json` (read/write/bump) |
| `lib/hash.sh` | SHA256 source hash computation and change detection |
| `lib/orchestrator.sh` | CLI parsing, build loop, component dispatching |

## Key Functions

- `log_info/success/warn/error/step/header` - Colored logging
- `detect_runtime/detect_compose` - Container runtime detection
- `git_short_commit/git_branch/git_is_dirty` - Git helpers
- `get_version/get_version_string/bump_version` - Version management
- `compute_source_hash/needs_rebuild/show_hash_status` - Change detection
- `create_release_dir/generate_checksums/generate_build_info` - Artifact generation
- `build_main/parse_build_args` - CLI orchestration

## Integration

Projects must define:
- `BUILD_COMPONENTS` - Array of component names
- `BUILD_COMPONENT_PATTERNS` - Associative array mapping component -> file patterns
- `build_single_component()` - Function to build a single component

## Dependencies

- Bash 4.0+ (associative arrays)
- Python 3 (JSON manipulation)
- sha256sum (coreutils)
- Git

## Commit Style

Conventional Commits: `feat(hash): add custom exclude patterns`

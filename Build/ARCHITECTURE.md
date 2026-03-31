# Architecture -- Build

## Purpose

Generic, reusable shell-based build framework providing automatic semantic versioning, SHA256 change detection, container runtime detection, and multi-component build orchestration. Designed to be included as a submodule in any multi-component project.

## Structure

```
lib/
  common.sh        Logging, container runtime detection (Podman/Docker), git helpers, artifact generation
  version.sh       Semantic versioning via versions.json (read/write/bump major/minor/patch)
  hash.sh          SHA256 source hash computation and change detection (skip unchanged components)
  orchestrator.sh  CLI parsing, build loop, component dispatching, --dry-run/--force/--component flags
```

## Key Components

- **`common.sh`** -- Colored logging (`log_info`, `log_error`, `log_step`), `detect_runtime`/`detect_compose` for Podman/Docker, git helpers (`git_short_commit`, `git_is_dirty`), artifact generation (`create_release_dir`, `generate_checksums`, `generate_build_info`)
- **`version.sh`** -- Reads/writes `versions.json` with Python JSON manipulation. Functions: `get_version`, `get_version_string`, `bump_version`, `get_build_number`
- **`hash.sh`** -- Computes SHA256 hashes of source file patterns per component. Functions: `compute_source_hash`, `needs_rebuild`, `show_hash_status`
- **`orchestrator.sh`** -- Parses CLI arguments, iterates over `BUILD_COMPONENTS`, calls `build_single_component()` for each changed component. Functions: `build_main`, `parse_build_args`

## Data Flow

```
build_main(args) -> parse_build_args -> for each component in BUILD_COMPONENTS:
    compute_source_hash(component, patterns)
        |
    needs_rebuild? (compare with versions.json last_source_hash)
        |
    build_single_component(component, version, build_number, version_string, hash)
        |
    generate_build_info + generate_checksums -> release directory
        |
    update versions.json (new hash, build number, date, git commit)
```

## Dependencies

- Bash 4.0+ (for associative arrays)
- Python 3 (for JSON manipulation of versions.json)
- sha256sum (coreutils)
- Git

## Testing Strategy

The framework is tested via integration in the Catalogizer project's `scripts/release-build.sh`. The `--dry-run` flag allows verification without building. `--status` shows change detection state for all components.

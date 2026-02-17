#!/usr/bin/env bash
# release-build.sh - Master orchestrator for Catalogizer release builds
# Host-side entry point: CLI parsing, hash comparison, version bumping,
# optional container launch, and post-build bookkeeping
#
# Uses the generic Build framework (Build/ submodule) with
# Catalogizer-specific component builders (scripts/lib/)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ─── Source project config ───────────────────────────────────────────
source "$SCRIPT_DIR/lib/project-config.sh"

# ─── Source Build framework (generic) ────────────────────────────────
source "$BUILD_PROJECT_ROOT/Build/lib/common.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/version.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/hash.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/orchestrator.sh"

# ─── Source per-component builders (project-specific) ────────────────
source "$SCRIPT_DIR/lib/build-catalog-api.sh"
source "$SCRIPT_DIR/lib/build-catalog-web.sh"
source "$SCRIPT_DIR/lib/build-api-client.sh"
source "$SCRIPT_DIR/lib/build-desktop.sh"
source "$SCRIPT_DIR/lib/build-installer.sh"
source "$SCRIPT_DIR/lib/build-android.sh"
source "$SCRIPT_DIR/lib/build-androidtv.sh"

# ─── Component dispatch (called by orchestrator) ────────────────────

build_single_component() {
    local component="$1"
    local version="$2"
    local build_number="$3"
    local version_string="$4"
    local source_hash="$5"

    case "$component" in
        catalog-api)
            build_catalog_api "$version" "$build_number" "$version_string" "$source_hash"
            ;;
        catalog-web)
            build_catalog_web "$version" "$build_number" "$version_string" "$source_hash"
            ;;
        catalogizer-api-client)
            build_api_client "$version" "$build_number" "$version_string" "$source_hash"
            ;;
        catalogizer-desktop)
            build_desktop "$version" "$build_number" "$version_string" "$source_hash"
            ;;
        installer-wizard)
            build_installer "$version" "$build_number" "$version_string" "$source_hash"
            ;;
        catalogizer-android)
            build_android "$version" "$build_number" "$version_string" "$source_hash"
            ;;
        catalogizer-androidtv)
            build_androidtv "$version" "$build_number" "$version_string" "$source_hash"
            ;;
        *)
            log_error "No builder for component: $component"
            return 1
            ;;
    esac
}

# ─── Run ─────────────────────────────────────────────────────────────
build_main "$@"

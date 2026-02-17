#!/usr/bin/env bash
# build-component.sh - Container entry point for building Catalogizer components
# Called inside the builder container by docker-compose.build.yml
#
# Environment variables:
#   BUILD_VERSION      - Version string (e.g., "1.0.0")
#   BUILD_NUMBER       - Build number
#   BUILD_COMPONENTS   - Space-separated list of components to build (or "all")
#   FORCE_BUILD        - Set to "true" to force rebuild
#   SKIP_TESTS         - Set to "true" to skip tests

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Source project config and Build framework
source "$SCRIPT_DIR/project-config.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/common.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/version.sh"
source "$BUILD_PROJECT_ROOT/Build/lib/hash.sh"

# Source all component builders
source "$SCRIPT_DIR/build-catalog-api.sh"
source "$SCRIPT_DIR/build-catalog-web.sh"
source "$SCRIPT_DIR/build-api-client.sh"
source "$SCRIPT_DIR/build-desktop.sh"
source "$SCRIPT_DIR/build-installer.sh"
source "$SCRIPT_DIR/build-android.sh"
source "$SCRIPT_DIR/build-androidtv.sh"

# Read environment variables
BUILD_SKIP_TESTS="${SKIP_TESTS:-false}"
FORCE="${FORCE_BUILD:-false}"
COMPONENTS="${BUILD_COMPONENTS_ENV:-all}"

log_header "Container Build Entry Point"
log_info "Version: ${BUILD_VERSION:-unknown}"
log_info "Build number: ${BUILD_NUMBER:-0}"
log_info "Components: $COMPONENTS"
log_info "Skip tests: $BUILD_SKIP_TESTS"
log_info "Force: $FORCE"

# Determine which components to build
if [[ "$COMPONENTS" == "all" ]]; then
    COMPONENTS_TO_BUILD=("${BUILD_COMPONENTS[@]}")
else
    IFS=' ' read -ra COMPONENTS_TO_BUILD <<< "$COMPONENTS"
fi

# Build each component
build_ok=true
for component in "${COMPONENTS_TO_BUILD[@]}"; do
    log_header "Building: $component"

    local_version="${BUILD_VERSION:-$(get_version)}"
    local_build_number="${BUILD_NUMBER:-$(get_build_number)}"
    local_version_string="v${local_version}-build.${local_build_number}"
    local_source_hash="$(compute_source_hash "$component")"

    case "$component" in
        catalog-api)
            build_catalog_api "$local_version" "$local_build_number" "$local_version_string" "$local_source_hash" || build_ok=false
            ;;
        catalog-web)
            build_catalog_web "$local_version" "$local_build_number" "$local_version_string" "$local_source_hash" || build_ok=false
            ;;
        catalogizer-api-client)
            build_api_client "$local_version" "$local_build_number" "$local_version_string" "$local_source_hash" || build_ok=false
            ;;
        catalogizer-desktop)
            build_desktop "$local_version" "$local_build_number" "$local_version_string" "$local_source_hash" || build_ok=false
            ;;
        installer-wizard)
            build_installer "$local_version" "$local_build_number" "$local_version_string" "$local_source_hash" || build_ok=false
            ;;
        catalogizer-android)
            build_android "$local_version" "$local_build_number" "$local_version_string" "$local_source_hash" || build_ok=false
            ;;
        catalogizer-androidtv)
            build_androidtv "$local_version" "$local_build_number" "$local_version_string" "$local_source_hash" || build_ok=false
            ;;
        *)
            log_error "Unknown component: $component"
            build_ok=false
            ;;
    esac
done

if [[ "$build_ok" == "true" ]]; then
    log_success "All container builds completed successfully."
else
    log_error "Some container builds failed."
    exit 1
fi

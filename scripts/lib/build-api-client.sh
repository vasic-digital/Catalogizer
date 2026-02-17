#!/usr/bin/env bash
# build-api-client.sh - Builder for catalogizer-api-client (TypeScript library)
# Produces compiled dist/ with package.json

set -euo pipefail

build_api_client() {
    local version="$1"
    local build_number="$2"
    local version_string="$3"
    local source_hash="$4"
    local skip_tests="${BUILD_SKIP_TESTS:-false}"

    local comp_dir="$BUILD_PROJECT_ROOT/catalogizer-api-client"

    # Install dependencies
    log_step "Installing catalogizer-api-client dependencies..."
    (cd "$comp_dir" && npm ci --prefer-offline 2>&1) || {
        log_error "npm ci failed for catalogizer-api-client"
        return 1
    }

    # Run tests (unless skipped)
    if [[ "$skip_tests" != "true" ]]; then
        log_step "Running catalogizer-api-client tests..."
        if (cd "$comp_dir" && npm run test 2>&1); then
            log_success "Tests passed"
        else
            log_warn "Tests failed or not configured for catalogizer-api-client, continuing..."
        fi
    fi

    # Build
    log_step "Building catalogizer-api-client..."
    if (cd "$comp_dir" && npm run build 2>&1); then
        local release_dir
        release_dir="$(create_release_dir "catalogizer-api-client" "npm" "$version_string")"

        # Copy build output and package.json
        if [[ -d "$comp_dir/dist" ]]; then
            cp -r "$comp_dir/dist" "$release_dir/"
        fi
        cp "$comp_dir/package.json" "$release_dir/"

        generate_checksums "$release_dir"
        generate_build_info "$release_dir" "catalogizer-api-client" "npm" \
            "$version" "$build_number" "$version_string" "$source_hash"
        log_success "catalogizer-api-client -> $release_dir"
    else
        log_error "Build failed for catalogizer-api-client"
        return 1
    fi
}

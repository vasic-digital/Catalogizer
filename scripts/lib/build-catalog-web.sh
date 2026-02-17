#!/usr/bin/env bash
# build-catalog-web.sh - Builder for catalog-web (React/TypeScript frontend)
# Produces optimized static dist/ bundle

set -euo pipefail

build_catalog_web() {
    local version="$1"
    local build_number="$2"
    local version_string="$3"
    local source_hash="$4"
    local skip_tests="${BUILD_SKIP_TESTS:-false}"

    local comp_dir="$BUILD_PROJECT_ROOT/catalog-web"

    # Install dependencies
    log_step "Installing catalog-web dependencies..."
    (cd "$comp_dir" && npm ci --prefer-offline 2>&1) || {
        log_error "npm ci failed for catalog-web"
        return 1
    }

    # Run tests (unless skipped)
    if [[ "$skip_tests" != "true" ]]; then
        log_step "Running catalog-web tests..."
        if (cd "$comp_dir" && npm run test -- --run 2>&1); then
            log_success "Tests passed"
        else
            log_error "Tests failed for catalog-web"
            return 1
        fi
    fi

    # Build production bundle
    log_step "Building catalog-web production bundle..."
    if (cd "$comp_dir" && VITE_APP_VERSION="$version_string" npm run build 2>&1); then
        local release_dir
        release_dir="$(create_release_dir "catalog-web" "web" "$version_string")"

        # Copy dist contents
        cp -r "$comp_dir/dist" "$release_dir/"

        generate_checksums "$release_dir"
        generate_build_info "$release_dir" "catalog-web" "web" \
            "$version" "$build_number" "$version_string" "$source_hash"
        log_success "catalog-web -> $release_dir"
    else
        log_error "Build failed for catalog-web"
        return 1
    fi
}

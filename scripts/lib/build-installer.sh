#!/usr/bin/env bash
# build-installer.sh - Builder for installer-wizard (Tauri/Rust+React)
# Produces Linux artifacts (AppImage, .deb)
#
# Auto-container dispatch: if cargo is absent on the host, the build
# re-routes through the catalogizer-builder container (Constitution
# Article VII §7.1 — no sudo/root).

set -euo pipefail

# shellcheck source=/dev/null
source "$BUILD_PROJECT_ROOT/scripts/lib/auto-container.sh"

build_installer() {
    local version="$1"
    local build_number="$2"
    local version_string="$3"
    local source_hash="$4"
    local skip_tests="${BUILD_SKIP_TESTS:-false}"

    local comp_dir="$BUILD_PROJECT_ROOT/installer-wizard"

    # Auto-dispatch to builder container when host lacks Rust.
    if ! is_container && ! need_toolchain cargo; then
        log_info "installer-wizard: cargo missing on host — auto-dispatching to builder container"
        run_in_builder "source /workspace/scripts/lib/project-config.sh && source /workspace/Build/lib/common.sh && source /workspace/scripts/lib/build-installer.sh && build_installer '$version' '$build_number' '$version_string' '$source_hash'"
        return $?
    fi

    # Container: install deps and enable AppImage extraction (no FUSE in containers)
    if is_container; then
        if command -v apt-get &>/dev/null; then
            apt-get install -y --no-install-recommends xdg-utils >/dev/null 2>&1 || true
        fi
        export APPIMAGE_EXTRACT_AND_RUN=1
    fi

    # Install frontend dependencies
    log_step "Installing installer-wizard dependencies..."
    (cd "$comp_dir" && npm ci --prefer-offline 2>&1) || {
        log_error "npm ci failed for installer-wizard"
        return 1
    }

    # Run tests (unless skipped)
    if [[ "$skip_tests" != "true" ]]; then
        log_step "Running installer-wizard tests..."
        (cd "$comp_dir" && npm run test -- --run 2>&1) || {
            log_warn "Tests failed or not configured for installer-wizard, continuing..."
        }
    fi

    # Build Tauri application
    log_step "Building installer-wizard Tauri app..."
    if (cd "$comp_dir" && npm run tauri:build 2>&1); then
        local release_dir
        release_dir="$(create_release_dir "installer-wizard" "linux" "$version_string")"

        # Copy Tauri build artifacts
        local tauri_out="$comp_dir/src-tauri/target/release/bundle"
        if [[ -d "$tauri_out/appimage" ]]; then
            cp "$tauri_out"/appimage/*.AppImage "$release_dir/" 2>/dev/null || true
        fi
        if [[ -d "$tauri_out/deb" ]]; then
            cp "$tauri_out"/deb/*.deb "$release_dir/" 2>/dev/null || true
        fi
        if [[ -d "$tauri_out/rpm" ]]; then
            cp "$tauri_out"/rpm/*.rpm "$release_dir/" 2>/dev/null || true
        fi

        generate_checksums "$release_dir"
        generate_build_info "$release_dir" "installer-wizard" "linux" \
            "$version" "$build_number" "$version_string" "$source_hash"
        log_success "installer-wizard -> $release_dir"
    else
        log_error "Build failed for installer-wizard"
        return 1
    fi
}

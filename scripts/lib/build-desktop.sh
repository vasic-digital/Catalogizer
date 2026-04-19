#!/usr/bin/env bash
# build-desktop.sh - Builder for catalogizer-desktop (Tauri/Rust+React)
# Produces Linux artifacts (AppImage, .deb)
#
# Auto-container dispatch: if cargo/rustc are absent on the host, the
# build re-routes through the catalogizer-builder container per
# Constitution Article VII §7.1 (no sudo/root). Operator never needs
# to install Rust locally.

set -euo pipefail

# shellcheck source=/dev/null
source "$BUILD_PROJECT_ROOT/scripts/lib/auto-container.sh"

build_desktop() {
    local version="$1"
    local build_number="$2"
    local version_string="$3"
    local source_hash="$4"
    local skip_tests="${BUILD_SKIP_TESTS:-false}"

    local comp_dir="$BUILD_PROJECT_ROOT/catalogizer-desktop"

    # If we're NOT already inside the builder container AND cargo is
    # missing locally, re-dispatch the entire desktop build into the
    # catalogizer-builder image. The container has Rust + Node + VLC
    # headers pre-installed per docker/Dockerfile.builder.
    if ! is_container && ! need_toolchain cargo; then
        log_info "catalogizer-desktop: cargo missing on host — auto-dispatching to builder container"
        run_in_builder "source /workspace/scripts/lib/project-config.sh && source /workspace/Build/lib/common.sh && source /workspace/scripts/lib/build-desktop.sh && build_desktop '$version' '$build_number' '$version_string' '$source_hash'"
        return $?
    fi

    # Container: configure build environment
    local -a tauri_bundle_flags=()
    if is_container; then
        export APPIMAGE_EXTRACT_AND_RUN=1
        # Ensure pkg-config can find VLC libs when vlc-player feature is enabled
        export PKG_CONFIG_PATH="/usr/lib/x86_64-linux-gnu/pkgconfig:${PKG_CONFIG_PATH:-}"
        # AppImage bundling requires xdg-open + FUSE which are unreliable in containers.
        # Produce .deb and .rpm only — AppImage can be built on the host.
        tauri_bundle_flags=("--" "--bundles" "deb,rpm")
    fi

    # Install frontend dependencies
    log_step "Installing catalogizer-desktop dependencies..."
    (cd "$comp_dir" && npm ci --prefer-offline 2>&1) || {
        log_error "npm ci failed for catalogizer-desktop"
        return 1
    }

    # Run tests (unless skipped)
    if [[ "$skip_tests" != "true" ]]; then
        log_step "Running catalogizer-desktop tests..."
        (cd "$comp_dir" && npm run test -- --run 2>&1) || {
            log_warn "Tests failed or not configured for catalogizer-desktop, continuing..."
        }
    fi

    # Build Tauri application
    log_step "Building catalogizer-desktop Tauri app..."
    if (cd "$comp_dir" && npm run tauri:build "${tauri_bundle_flags[@]}" 2>&1); then
        local release_dir
        release_dir="$(create_release_dir "catalogizer-desktop" "linux" "$version_string")"

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
        generate_build_info "$release_dir" "catalogizer-desktop" "linux" \
            "$version" "$build_number" "$version_string" "$source_hash"
        log_success "catalogizer-desktop -> $release_dir"
    else
        log_error "Build failed for catalogizer-desktop"
        return 1
    fi
}

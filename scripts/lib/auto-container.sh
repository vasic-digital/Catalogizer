#!/usr/bin/env bash
# auto-container.sh — automatic container dispatch for host-builds that
# require toolchains (Rust, etc.) not installed on the host. Enforces
# Constitution Article VII §7.1 ("no-sudo / no-root") by routing through
# rootless podman (or docker) and the catalogizer-builder image.
#
# Usage (sourced by scripts/lib/build-*.sh):
#   source "$BUILD_PROJECT_ROOT/scripts/lib/auto-container.sh"
#   need_toolchain cargo || run_in_builder "cd /workspace/catalogizer-desktop && npm run tauri:build -- --bundles deb,rpm"
#
# Env overrides:
#   CATALOGIZER_BUILDER_IMAGE — builder image tag (default: localhost/catalogizer-builder:latest)
#   CATALOGIZER_BUILDER_AUTO=0 — disable auto-dispatch (force local)

set -euo pipefail

: "${CATALOGIZER_BUILDER_IMAGE:=localhost/catalogizer-builder:latest}"
: "${CATALOGIZER_BUILDER_AUTO:=1}"

# need_toolchain NAME — returns 0 if NAME is in PATH, else 1.
need_toolchain() {
    command -v "$1" >/dev/null 2>&1
}

# builder_image_exists — returns 0 if the builder image is already built
# locally.
builder_image_exists() {
    local runtime
    runtime="$(detect_runtime)"
    "$runtime" image exists "$CATALOGIZER_BUILDER_IMAGE" 2>/dev/null
}

# ensure_builder_image — builds the catalogizer-builder image if it is
# missing locally. Uses rootless podman/docker.
ensure_builder_image() {
    if builder_image_exists; then
        return 0
    fi
    local runtime
    runtime="$(detect_runtime)"
    log_step "Building $CATALOGIZER_BUILDER_IMAGE (first use; ~8-15 min)…"
    (cd "$BUILD_PROJECT_ROOT" && "$runtime" build \
        --network host \
        -f docker/Dockerfile.builder \
        -t "$CATALOGIZER_BUILDER_IMAGE" \
        . 2>&1) || {
            log_error "Failed to build $CATALOGIZER_BUILDER_IMAGE"
            return 1
    }
    log_success "Builder image ready: $CATALOGIZER_BUILDER_IMAGE"
}

# run_in_builder CMD — executes the given shell command inside a
# rootless container, with the project mounted at /workspace. Returns
# the command's exit code. Automatically ensures the builder image
# exists.
run_in_builder() {
    local cmd="$1"
    if [[ "$CATALOGIZER_BUILDER_AUTO" != "1" ]]; then
        log_error "auto-container: CATALOGIZER_BUILDER_AUTO=0 but toolchain missing"
        return 1
    fi
    ensure_builder_image || return 1
    local runtime
    runtime="$(detect_runtime)"
    log_step "Dispatching to $CATALOGIZER_BUILDER_IMAGE…"
    "$runtime" run --rm \
        --network host \
        --userns=keep-id \
        -v "$BUILD_PROJECT_ROOT:/workspace:z" \
        -w /workspace \
        -e APPIMAGE_EXTRACT_AND_RUN=1 \
        -e GOTOOLCHAIN=local \
        "$CATALOGIZER_BUILDER_IMAGE" \
        bash -c "$cmd"
}

# dispatch_if_missing TOOL CMD — if TOOL is not available locally,
# runs CMD inside the builder container. Otherwise runs CMD locally
# via bash -c. Designed to be a drop-in replacement for ad-hoc
# toolchain-dependent invocations.
dispatch_if_missing() {
    local tool="$1"
    local cmd="$2"
    if need_toolchain "$tool"; then
        bash -c "$cmd"
    else
        log_info "Toolchain '$tool' missing locally — auto-dispatching to builder container"
        run_in_builder "$cmd"
    fi
}

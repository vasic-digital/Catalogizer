#!/usr/bin/env bash
# container-runtime.sh - Dynamic container runtime detection library
# Detects available container runtime (Podman > Docker > nerdctl) and provides
# unified wrapper functions for container operations.
#
# Usage: source this file, then call detect_container_runtime()
#   source "$(dirname "$0")/lib/container-runtime.sh"
#   detect_container_runtime
#
# Globals set by detect_container_runtime():
#   CONTAINER_CMD      - Container CLI command (e.g., "podman", "docker", "nerdctl")
#   COMPOSE_CMD        - Compose CLI command (e.g., "podman-compose", "docker-compose", "docker compose")
#   CONTAINER_RUNTIME  - Human-readable runtime name ("podman", "docker", "nerdctl")
#
# No side effects on source -- only function definitions.

# ─── Internal helpers ───────────────────────────────────────────────

_cr_log_info() {
    echo -e "\033[0;34m[INFO]\033[0m $*"
}

_cr_log_warn() {
    echo -e "\033[1;33m[WARN]\033[0m $*"
}

_cr_log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $*" >&2
}

# ─── Detection ──────────────────────────────────────────────────────

# Detect the compose companion for a given container runtime.
# Args: $1 = runtime name ("podman", "docker", "nerdctl")
# Prints the compose command string or returns 1.
_detect_compose_for() {
    local runtime="$1"
    case "$runtime" in
        podman)
            if command -v podman-compose &>/dev/null; then
                echo "podman-compose"
                return 0
            fi
            _cr_log_warn "podman-compose not found -- install with: pip install podman-compose"
            return 1
            ;;
        docker)
            if command -v docker-compose &>/dev/null; then
                echo "docker-compose"
                return 0
            elif docker compose version &>/dev/null 2>&1; then
                echo "docker compose"
                return 0
            fi
            _cr_log_warn "docker-compose not found"
            return 1
            ;;
        nerdctl)
            if nerdctl compose version &>/dev/null 2>&1; then
                echo "nerdctl compose"
                return 0
            fi
            _cr_log_warn "nerdctl compose not available"
            return 1
            ;;
    esac
    return 1
}

# Detect and set container runtime globals.
# Priority: Podman > Docker > nerdctl
# Returns 0 on success, 1 if no runtime is found.
detect_container_runtime() {
    local runtimes=("podman" "docker" "nerdctl")

    for rt in "${runtimes[@]}"; do
        if command -v "$rt" &>/dev/null; then
            CONTAINER_CMD="$rt"
            CONTAINER_RUNTIME="$rt"

            if COMPOSE_CMD="$(_detect_compose_for "$rt")"; then
                _cr_log_info "Container runtime: ${CONTAINER_RUNTIME} (compose: ${COMPOSE_CMD})"
            else
                COMPOSE_CMD=""
                _cr_log_info "Container runtime: ${CONTAINER_RUNTIME} (no compose companion found)"
            fi

            export CONTAINER_CMD COMPOSE_CMD CONTAINER_RUNTIME
            return 0
        fi
    done

    _cr_log_error "No container runtime found. Install one of: podman (preferred), docker, nerdctl."
    return 1
}

# ─── Public API ─────────────────────────────────────────────────────

# Check whether any supported container runtime is available.
# Returns 0 if a runtime can be found, 1 otherwise. Does NOT set globals.
container_available() {
    command -v podman &>/dev/null ||
    command -v docker &>/dev/null ||
    command -v nerdctl &>/dev/null
}

# Run a container. Passes all arguments to the detected runtime.
# Usage: container_run --rm --network host my-image:latest
container_run() {
    if [[ -z "${CONTAINER_CMD:-}" ]]; then
        _cr_log_error "container_run: runtime not detected -- call detect_container_runtime first"
        return 1
    fi
    $CONTAINER_CMD run "$@"
}

# Build a container image. Passes all arguments to the detected runtime.
# Usage: container_build -f Dockerfile.api -t catalogizer/api:latest .
container_build() {
    if [[ -z "${CONTAINER_CMD:-}" ]]; then
        _cr_log_error "container_build: runtime not detected -- call detect_container_runtime first"
        return 1
    fi
    $CONTAINER_CMD build "$@"
}

# Run a compose command. Passes all arguments to the detected compose companion.
# Usage: container_compose -f docker-compose.dev.yml up -d
container_compose() {
    if [[ -z "${COMPOSE_CMD:-}" ]]; then
        _cr_log_error "container_compose: compose command not available -- call detect_container_runtime first"
        return 1
    fi
    $COMPOSE_CMD "$@"
}

#!/bin/bash
# ============================================================
# Catalogizer - Containerized Build Entry Point
# Detects container runtime and launches the build pipeline
# Usage: ./scripts/container-build.sh [version] [--skip-emulator] [--skip-e2e] [--with-emulator]
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Defaults
VERSION="1.0.0"
SKIP_EMULATOR="true"
SKIP_E2E="false"
EXTRA_PROFILES=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[BUILD]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[BUILD]${NC} $1"; }
log_error() { echo -e "${RED}[BUILD]${NC} $1"; }

usage() {
    echo "Usage: $0 [version] [options]"
    echo ""
    echo "Arguments:"
    echo "  version              Build version (default: 1.0.0)"
    echo ""
    echo "Options:"
    echo "  --skip-emulator      Skip Android emulator tests (default)"
    echo "  --with-emulator      Enable Android emulator tests (requires /dev/kvm)"
    echo "  --skip-e2e           Skip Playwright E2E tests"
    echo "  --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 1.0.0"
    echo "  $0 2.0.0 --with-emulator"
    echo "  $0 1.0.0 --skip-e2e"
}

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --skip-emulator)
            SKIP_EMULATOR="true"
            shift
            ;;
        --with-emulator)
            SKIP_EMULATOR="false"
            EXTRA_PROFILES="--profile emulator"
            shift
            ;;
        --skip-e2e)
            SKIP_E2E="true"
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        -*)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

echo -e "${CYAN}"
echo "============================================="
echo "  Catalogizer Containerized Build Pipeline"
echo "  Version: $VERSION"
echo "============================================="
echo -e "${NC}"

# ============================================================
# Detect container runtime
# ============================================================
CONTAINER_CMD=""
COMPOSE_CMD=""

if command -v podman &>/dev/null; then
    CONTAINER_CMD="podman"
    if command -v podman-compose &>/dev/null; then
        COMPOSE_CMD="podman-compose"
    else
        log_error "podman-compose is not installed"
        log_info "Install with: pip3 install podman-compose"
        exit 1
    fi
elif command -v docker &>/dev/null; then
    CONTAINER_CMD="docker"
    if command -v docker-compose &>/dev/null; then
        COMPOSE_CMD="docker-compose"
    elif docker compose version &>/dev/null 2>&1; then
        COMPOSE_CMD="docker compose"
    else
        log_error "docker-compose is not installed"
        exit 1
    fi
else
    log_error "Neither podman nor docker is installed"
    log_info "Install podman: https://podman.io/getting-started/installation"
    log_info "Install docker: https://docs.docker.com/engine/install/"
    exit 1
fi

log_info "Container runtime: $CONTAINER_CMD"
log_info "Compose tool: $COMPOSE_CMD"

# ============================================================
# Check emulator prerequisites
# ============================================================
if [ "$SKIP_EMULATOR" = "false" ]; then
    if [ -e /dev/kvm ]; then
        log_info "KVM available - Android emulator will be started"
    else
        log_warn "/dev/kvm not available - disabling Android emulator"
        SKIP_EMULATOR="true"
        EXTRA_PROFILES=""
    fi
fi

# ============================================================
# Generate signing keys (before container build)
# ============================================================
log_info "Checking signing keys..."
if [ -x "$PROJECT_ROOT/docker/signing/generate-keys.sh" ]; then
    "$PROJECT_ROOT/docker/signing/generate-keys.sh" || log_warn "Signing key generation had warnings"
fi

# ============================================================
# Validate compose file
# ============================================================
log_info "Validating docker-compose.build.yml..."
cd "$PROJECT_ROOT"
$COMPOSE_CMD -f docker-compose.build.yml config --quiet 2>/dev/null || {
    # Some compose versions don't support --quiet
    $COMPOSE_CMD -f docker-compose.build.yml config >/dev/null 2>&1 || {
        log_error "docker-compose.build.yml validation failed"
        exit 1
    }
}
log_info "Compose file is valid"

# ============================================================
# Run the build pipeline
# ============================================================
log_info "Starting containerized build pipeline..."
log_info "Build version: $VERSION"
log_info "Skip emulator: $SKIP_EMULATOR"
log_info "Skip E2E: $SKIP_E2E"

export BUILD_VERSION="$VERSION"
export SKIP_EMULATOR_TESTS="$SKIP_EMULATOR"
export SKIP_E2E_TESTS="$SKIP_E2E"

# Build and run
$COMPOSE_CMD -f docker-compose.build.yml $EXTRA_PROFILES up --build --abort-on-container-exit
EXIT_CODE=$?

# ============================================================
# Cleanup containers
# ============================================================
log_info "Stopping services..."
$COMPOSE_CMD -f docker-compose.build.yml $EXTRA_PROFILES down --volumes 2>/dev/null || true

# ============================================================
# Print results summary
# ============================================================
echo ""
echo -e "${CYAN}=============================================${NC}"
echo -e "${CYAN}  Build Pipeline Results${NC}"
echo -e "${CYAN}=============================================${NC}"

if [ -f "$PROJECT_ROOT/releases/MANIFEST.json" ]; then
    log_info "MANIFEST.json:"
    cat "$PROJECT_ROOT/releases/MANIFEST.json" 2>/dev/null || true
    echo ""
fi

if [ -f "$PROJECT_ROOT/releases/SHA256SUMS.txt" ]; then
    log_info "SHA256 checksums:"
    cat "$PROJECT_ROOT/releases/SHA256SUMS.txt" 2>/dev/null || true
    echo ""
fi

# Count artifacts
if [ -d "$PROJECT_ROOT/releases" ]; then
    ARTIFACT_COUNT=$(find "$PROJECT_ROOT/releases" -type f \( -name "*.exe" -o -name "*.AppImage" \
        -o -name "*.deb" -o -name "*.apk" -o -name "catalog-api*" \) 2>/dev/null | wc -l || echo 0)
    log_info "Total release artifacts: $ARTIFACT_COUNT"
fi

if [ -d "$PROJECT_ROOT/reports" ]; then
    log_info "Reports directory: $PROJECT_ROOT/reports/"
    ls -la "$PROJECT_ROOT/reports/"*.html 2>/dev/null || true
fi

echo ""
if [ "$EXIT_CODE" -eq 0 ]; then
    echo -e "${GREEN}Build pipeline completed successfully!${NC}"
else
    echo -e "${RED}Build pipeline failed with exit code $EXIT_CODE${NC}"
fi

exit $EXIT_CODE

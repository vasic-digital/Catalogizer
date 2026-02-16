#!/bin/bash
# Generate default signing keys for containerized builds
# These are DEBUG/CI keys only - do NOT use for production releases

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYSTORE_FILE="$SCRIPT_DIR/catalogizer-debug.keystore"
SIGNING_PROPS="$SCRIPT_DIR/signing.properties"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[SIGNING]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[SIGNING]${NC} $1"; }

# Generate Android debug keystore
generate_android_keystore() {
    if [ -f "$KEYSTORE_FILE" ]; then
        log_info "Android keystore already exists: $KEYSTORE_FILE"
        return 0
    fi

    log_info "Generating Android debug keystore..."

    if ! command -v keytool &>/dev/null; then
        log_warn "keytool not found - skipping Android keystore generation"
        return 1
    fi

    keytool -genkeypair -v \
        -keystore "$KEYSTORE_FILE" \
        -alias catalogizer \
        -keyalg RSA \
        -keysize 2048 \
        -validity 10000 \
        -storepass catalogizer123 \
        -keypass catalogizer123 \
        -dname "CN=Catalogizer Dev, OU=Dev, O=Catalogizer, L=Dev, ST=Dev, C=US"

    log_info "Android keystore generated: $KEYSTORE_FILE"
}

# Generate signing.properties for Gradle
generate_signing_properties() {
    if [ -f "$SIGNING_PROPS" ]; then
        log_info "signing.properties already exists: $SIGNING_PROPS"
        return 0
    fi

    log_info "Generating signing.properties..."

    cat > "$SIGNING_PROPS" << EOF
# Auto-generated signing properties for containerized builds
# DEBUG/CI use only - do NOT use for production releases
storeFile=$KEYSTORE_FILE
storePassword=catalogizer123
keyAlias=catalogizer
keyPassword=catalogizer123
EOF

    log_info "signing.properties generated: $SIGNING_PROPS"
}

# Generate Tauri updater signing key (optional)
generate_tauri_key() {
    local tauri_key_file="$SCRIPT_DIR/tauri-updater.key"

    if [ -f "$tauri_key_file" ]; then
        log_info "Tauri signing key already exists: $tauri_key_file"
        return 0
    fi

    if command -v cargo &>/dev/null && cargo tauri --version &>/dev/null 2>&1; then
        log_info "Generating Tauri updater signing key..."
        # Generate key pair; tauri signer writes to stdout
        TAURI_SIGNING_PRIVATE_KEY_PASSWORD="" cargo tauri signer generate \
            -w "$tauri_key_file" 2>/dev/null || {
            log_warn "Tauri signer not available - skipping Tauri key generation"
            return 0
        }
        log_info "Tauri signing key generated: $tauri_key_file"
    else
        log_warn "Tauri CLI not available - skipping Tauri key generation"
    fi
}

# Main
main() {
    log_info "=== Signing Key Generation ==="
    log_info "Output directory: $SCRIPT_DIR"

    generate_android_keystore
    generate_signing_properties
    generate_tauri_key

    log_info "=== Signing key generation complete ==="
}

main "$@"

#!/bin/bash

# Wrapper script for Android development environment setup
# Usage: ./scripts/setup-android.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Run the Android setup script
log_info() { echo -e "\033[0;34m[INFO]\033[0m $1"; }
log_success() { echo -e "\033[0;32m[SUCCESS]\033[0m $1"; }

log_info "Starting Android development environment setup..."
log_info "This will download JDK 21 with jmods and Android SDK into tools/ directory"
log_info "Approximate download size: ~2GB (JDK: ~200MB, Android SDK: ~1.8GB)"
echo ""

# Check for sufficient disk space (minimum 5GB free)
if command -v df &>/dev/null; then
    available_gb=$(df -B1G "$PROJECT_ROOT" | awk 'NR==2 {print $4}' | tr -d 'G')
    if [[ $available_gb -lt 5 ]]; then
        echo -e "\033[1;33m[WARNING]\033[0m Low disk space: ${available_gb}GB available, recommend at least 5GB"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
fi

# Run the setup script
"$SCRIPT_DIR/android/setup.sh"

log_success "Android development environment setup complete!"
echo ""
echo "To activate the environment for your current shell, run:"
echo "  source tools/android-env.sh"
echo ""
echo "To build the Android project, run:"
echo "  cd catalogizer-android && ./gradlew assembleDebug"
echo ""
echo "To run unit tests, run:"
echo "  cd catalogizer-android && ./gradlew test"
echo ""
echo "Note: All downloaded components are in tools/ directory (ignored by git)"
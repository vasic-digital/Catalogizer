#!/usr/bin/env bash
# build-android.sh - Builder for catalogizer-android (Kotlin/Compose)
# Produces release APK (falls back to debug if signing is unavailable)

set -euo pipefail

build_android() {
    local version="$1"
    local build_number="$2"
    local version_string="$3"
    local source_hash="$4"
    local skip_tests="${BUILD_SKIP_TESTS:-false}"

    local comp_dir="$BUILD_PROJECT_ROOT/catalogizer-android"

    # Ensure ANDROID_HOME is set
    if [[ -z "${ANDROID_HOME:-}" ]]; then
        if [[ -d "$HOME/Android/Sdk" ]]; then
            export ANDROID_HOME="$HOME/Android/Sdk"
        elif [[ -d "/opt/android-sdk" ]]; then
            export ANDROID_HOME="/opt/android-sdk"
        fi
    fi

    # Run tests (unless skipped)
    if [[ "$skip_tests" != "true" ]]; then
        log_step "Running catalogizer-android tests..."
        if (cd "$comp_dir" && ./gradlew test 2>&1); then
            log_success "Tests passed"
        else
            log_warn "Tests failed or not configured for catalogizer-android, continuing..."
        fi
    fi

    # Try release first, fall back to debug if signing fails
    log_step "Building catalogizer-android APK..."
    local build_type="release"
    local gradle_task="assembleRelease"

    if ! (cd "$comp_dir" && ./gradlew "$gradle_task" \
        -PversionName="$version" \
        -PversionCode="$build_number" 2>&1); then
        log_warn "Release build failed (likely missing keystore). Falling back to debug build..."
        build_type="debug"
        gradle_task="assembleDebug"
        if ! (cd "$comp_dir" && ./gradlew "$gradle_task" \
            -PversionName="$version" \
            -PversionCode="$build_number" 2>&1); then
            log_error "Build failed for catalogizer-android"
            return 1
        fi
    fi

    local release_dir
    release_dir="$(create_release_dir "catalogizer-android" "android" "$version_string")"

    # Find and copy APK
    local apk
    apk=$(find "$comp_dir" -path "*/${build_type}/*.apk" -type f 2>/dev/null | head -1)
    if [[ -z "$apk" ]]; then
        apk=$(find "$comp_dir" -name "*.apk" -type f 2>/dev/null | head -1)
    fi
    if [[ -n "$apk" ]]; then
        cp "$apk" "$release_dir/catalogizer-android.apk"
    else
        log_warn "No APK found after build"
    fi

    generate_checksums "$release_dir"
    generate_build_info "$release_dir" "catalogizer-android" "android" \
        "$version" "$build_number" "$version_string" "$source_hash"
    log_success "catalogizer-android ($build_type) -> $release_dir"
}

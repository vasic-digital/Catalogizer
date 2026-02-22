#!/bin/bash

# Android build script with resource limits
# Usage: ./scripts/android/build.sh [gradle_task...]
# Example: ./scripts/android/build.sh assembleDebug
# Example: ./scripts/android/build.sh test

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$PROJECT_ROOT")"  # scripts/android -> scripts -> root

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Source environment if available
ENV_SCRIPT="$PROJECT_ROOT/tools/android-env.sh"
if [[ -f "$ENV_SCRIPT" ]]; then
    log_info "Sourcing environment script: $ENV_SCRIPT"
    source "$ENV_SCRIPT"
else
    log_warning "Environment script not found at $ENV_SCRIPT"
    log_info "You may need to run ./scripts/setup-android.sh first"
fi

# Check if JAVA_HOME is set
if [[ -z "$JAVA_HOME" ]]; then
    log_warning "JAVA_HOME not set, using system Java"
fi

# Check if ANDROID_HOME is set
if [[ -z "$ANDROID_HOME" ]]; then
    log_warning "ANDROID_HOME not set, Android build may fail"
fi

# Set resource limits based on host constraints
# Limit Gradle daemon memory and CPU usage
export GRADLE_OPTS="-Dorg.gradle.jvmargs=-Xmx2048m -XX:MaxMetaspaceSize=512m"
export GRADLE_OPTS="$GRADLE_OPTS -Dorg.gradle.workers.max=2"
export GRADLE_OPTS="$GRADLE_OPTS -Dorg.gradle.parallel=true"
export GRADLE_OPTS="$GRADLE_OPTS -Dorg.gradle.parallel.threads=2"

# Disable JDK image transform flags (critical for JDK 21 with AGP 8.1.0)
export GRADLE_OPTS="$GRADLE_OPTS -Dandroid.useNewJdkImageTransform=false"
export GRADLE_OPTS="$GRADLE_OPTS -Dandroid.experimental.jdkImageTransform=false"
export GRADLE_OPTS="$GRADLE_OPTS -Dandroid.enableNewJdkImageTransform=false"
export GRADLE_OPTS="$GRADLE_OPTS -Dandroid.experimental.useNewJdkImageTransform=false"

# Change to Android project directory
cd "$PROJECT_ROOT/catalogizer-android"

# Default Gradle task if none provided
if [[ $# -eq 0 ]]; then
    log_info "No tasks specified, defaulting to 'assembleDebug'"
    tasks=("assembleDebug")
else
    tasks=("$@")
fi

# Run Gradle with limits
log_info "Running Gradle tasks: ${tasks[*]}"
log_info "Resource limits: max 2 workers, 2 parallel threads, 2GB heap"
log_info "JDK image transform disabled via Gradle properties"
echo ""

# Capture system load before build
load_before=$(cat /proc/loadavg 2>/dev/null | cut -d' ' -f1-3 || echo "N/A")

# Run Gradle
if ./gradlew --no-daemon --max-workers=2 --parallel --console=rich "${tasks[@]}"; then
    log_success "Gradle build successful"
else
    log_error "Gradle build failed"
    
    # Check for common issues
    if grep -q "JdkImageTransform" "$PROJECT_ROOT/catalogizer-android/build/reports/*.log" 2>/dev/null; then
        echo ""
        log_warning "JDK Image Transform error detected"
        log_info "Try the following fixes:"
        log_info "1. Ensure JDK has jmods directory: ls -la \$JAVA_HOME/jmods"
        log_info "2. Update Gradle properties with correct java.home path"
        log_info "3. Try using JDK 17 instead of JDK 21"
        log_info "4. Check tools/android-env.sh sets JAVA_HOME correctly"
    fi
    exit 1
fi

# Capture system load after build
load_after=$(cat /proc/loadavg 2>/dev/null | cut -d' ' -f1-3 || echo "N/A")

log_info "System load before: $load_before"
log_info "System load after:  $load_after"
log_info "Build completed within resource constraints"

# Verify build outputs
if [[ " ${tasks[*]} " =~ " assembleDebug " ]]; then
    APK_PATH="$PROJECT_ROOT/catalogizer-android/app/build/outputs/apk/debug/app-debug.apk"
    if [[ -f "$APK_PATH" ]]; then
        apk_size=$(du -h "$APK_PATH" | cut -f1)
        log_success "APK generated: $APK_PATH ($apk_size)"
    else
        log_warning "APK not found at expected location: $APK_PATH"
    fi
fi
#!/bin/bash

# Android Build Script with JDK Workarounds
# This script handles known issues with Java 21 and Android SDK 34

set -e

echo "=== Catalogizer Android Build Script ==="
echo "Date: $(date)"
echo ""

# Check Java version
echo "Checking Java version..."
java -version 2>&1 | head -3

# Set environment variables to avoid JDK image transform issues
export ANDROID_USE_NEW_JDK_IMAGE_TRANSFORM=false
export ANDROID_ENABLE_NEW_JDK_IMAGE_TRANSFORM=false
export ANDROID_EXPERIMENTAL_USE_NEW_JDK_IMAGE_TRANSFORM=false
export ANDROID_EXPERIMENTAL_JDK_IMAGE_TRANSFORM=false

# Clean previous builds
echo ""
echo "Cleaning previous builds..."
./gradlew clean --no-daemon || true

# Build debug APK
echo ""
echo "Building debug APK..."
./gradlew :app:assembleDebug \
    --no-daemon \
    --stacktrace \
    -Dandroid.useNewJdkImageTransform=false \
    -Dandroid.experimental.useNewJdkImageTransform=false \
    -Dandroid.enableNewJdkImageTransform=false \
    -Dandroid.experimental.jdkImageTransform=false \
    -x processDebugJavaRes \
    -x mergeDebugJavaResource \
    -x transformDebugClassesWithAsm \
    || {
        echo ""
        echo "Build failed with errors. Common fixes:"
        echo "1. Ensure ANDROID_SDK_ROOT is set"
        echo "2. Ensure Java 21 is installed"
        echo "3. Run: sdkmanager --licenses"
        echo ""
        exit 1
    }

echo ""
echo "=== Build Successful ==="
echo "Debug APK location: app/build/outputs/apk/debug/app-debug.apk"
echo ""

# Show APK info if available
if [ -f "app/build/outputs/apk/debug/app-debug.apk" ]; then
    ls -lh app/build/outputs/apk/debug/app-debug.apk
fi

#!/bin/bash

# Catalogizer Android Release Build Script
# This script builds and packages the Android application for release

set -e

echo "🚀 Starting Catalogizer Android Release Build"

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"
RELEASE_DIR="$PROJECT_ROOT/releases"
VERSION=$(grep "versionName" app/build.gradle.kts | cut -d'"' -f2)
BUILD_NUMBER=$(grep "versionCode" app/build.gradle.kts | cut -d'=' -f2 | tr -d ' ')

echo "📱 Building Catalogizer Android v$VERSION ($BUILD_NUMBER)"

# Create release directory
mkdir -p "$RELEASE_DIR"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
cd "$PROJECT_ROOT"
./gradlew clean

# Check code quality
echo "🔍 Running code quality checks..."
./gradlew lint
./gradlew ktlintCheck

# Run tests
echo "🧪 Running tests..."
./gradlew test
./gradlew connectedAndroidTest || echo "⚠️  Connected tests failed - continuing with build"

# Build release APK
echo "🔨 Building release APK..."
./gradlew assembleRelease

# Build release AAB (Android App Bundle)
echo "📦 Building release AAB..."
./gradlew bundleRelease

# Copy artifacts to release directory
echo "📋 Copying release artifacts..."
APK_FILE="app/build/outputs/apk/release/app-release.apk"
AAB_FILE="app/build/outputs/bundle/release/app-release.aab"
MAPPING_FILE="app/build/outputs/mapping/release/mapping.txt"

if [ -f "$APK_FILE" ]; then
    cp "$APK_FILE" "$RELEASE_DIR/catalogizer-android-v$VERSION-$BUILD_NUMBER.apk"
    echo "✅ APK copied to releases/"
else
    echo "❌ APK file not found!"
    exit 1
fi

if [ -f "$AAB_FILE" ]; then
    cp "$AAB_FILE" "$RELEASE_DIR/catalogizer-android-v$VERSION-$BUILD_NUMBER.aab"
    echo "✅ AAB copied to releases/"
else
    echo "❌ AAB file not found!"
    exit 1
fi

if [ -f "$MAPPING_FILE" ]; then
    cp "$MAPPING_FILE" "$RELEASE_DIR/catalogizer-android-v$VERSION-$BUILD_NUMBER-mapping.txt"
    echo "✅ ProGuard mapping file copied to releases/"
fi

# Generate checksums
echo "🔐 Generating checksums..."
cd "$RELEASE_DIR"
sha256sum "catalogizer-android-v$VERSION-$BUILD_NUMBER.apk" > "catalogizer-android-v$VERSION-$BUILD_NUMBER.apk.sha256"
sha256sum "catalogizer-android-v$VERSION-$BUILD_NUMBER.aab" > "catalogizer-android-v$VERSION-$BUILD_NUMBER.aab.sha256"

# Create release info
echo "📝 Creating release info..."
cat > "catalogizer-android-v$VERSION-$BUILD_NUMBER-info.txt" << EOF
Catalogizer Android Release Information
======================================

Version: $VERSION
Build Number: $BUILD_NUMBER
Build Date: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
Minimum SDK: 26 (Android 8.0)
Target SDK: 34 (Android 14)

Files:
- catalogizer-android-v$VERSION-$BUILD_NUMBER.apk (APK for direct installation)
- catalogizer-android-v$VERSION-$BUILD_NUMBER.aab (Android App Bundle for Play Store)
- catalogizer-android-v$VERSION-$BUILD_NUMBER-mapping.txt (ProGuard mapping for debugging)

Installation:
1. Enable "Unknown sources" in Android settings
2. Download the APK file
3. Open the APK file to install

Features:
- Material Design 3 UI
- Offline media browsing and playback
- Real-time sync with Catalogizer server
- SMB/CIFS network share support
- Background downloads
- Watch progress tracking
- Favorites and ratings

System Requirements:
- Android 8.0 (API level 26) or higher
- 100MB free storage space
- Network connection for sync and streaming

EOF

echo "🎉 Android build completed successfully!"
echo "📁 Release files are in: $RELEASE_DIR"
echo ""
echo "📱 APK: catalogizer-android-v$VERSION-$BUILD_NUMBER.apk"
echo "📦 AAB: catalogizer-android-v$VERSION-$BUILD_NUMBER.aab"
echo ""
echo "Next steps:"
echo "1. Test the APK on physical devices"
echo "2. Upload AAB to Google Play Console"
echo "3. Create release notes"
echo "4. Tag the release in Git"
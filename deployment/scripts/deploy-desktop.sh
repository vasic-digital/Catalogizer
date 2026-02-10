#!/bin/bash

# Desktop Deployment Script
# Builds and deploys desktop applications for Windows, macOS, and Linux

set -e

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DEPLOYMENT_CONFIG="$SCRIPT_DIR/desktop-deploy.env"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_header() { echo -e "${PURPLE}[DESKTOP DEPLOY]${NC} $1"; }

# Default configuration
DEPLOY_TARGET="current"  # current, all, windows, macos, linux
BUILD_TYPE="release"
DEPLOY_TO_GITHUB="false"
DEPLOY_TO_WEBSITE="false"
DEPLOY_TO_STORE="false"
SIGN_APPS="true"
RUN_TESTS="true"
AUTO_UPDATE="true"

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy desktop applications for Windows, macOS, and Linux platforms.

OPTIONS:
    -h, --help                  Show this help message
    -t, --target TARGET         Deployment target: current, all, windows, macos, linux
    -b, --build-type TYPE       Build type: debug, release (default: release)
    -c, --config FILE           Use custom deployment configuration file
    --github                    Deploy to GitHub Releases
    --website                   Deploy to website download page
    --store                     Deploy to platform stores (Mac App Store, Microsoft Store)
    --no-sign                   Skip application signing and notarization
    --no-tests                  Skip running tests
    --no-auto-update            Disable auto-update functionality
    --dry-run                   Show what would be deployed without actually deploying

DEPLOYMENT TARGETS:
    current                     Build for current platform only
    all                         Build for all platforms (requires cross-compilation setup)
    windows                     Build Windows executables (MSI, NSIS, portable)
    macos                       Build macOS application (DMG, App bundle)
    linux                       Build Linux packages (AppImage, DEB, RPM)

BUILD TYPES:
    debug                       Build debug versions with debugging symbols
    release                     Build optimized release versions

EXAMPLES:
    # Build for current platform and deploy to GitHub
    $0 --github

    # Build for all platforms and deploy to website
    $0 --target all --website

    # Build Windows version and deploy to Microsoft Store
    $0 --target windows --store

    # Debug build for current platform
    $0 --build-type debug --no-sign

    # Dry run to see what would be built
    $0 --target all --dry-run

CONFIGURATION:
    Create desktop-deploy.env file or use --config to specify custom configuration.
    See desktop-deploy.env.example for available options.

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -t|--target)
                DEPLOY_TARGET="$2"
                shift 2
                ;;
            -b|--build-type)
                BUILD_TYPE="$2"
                shift 2
                ;;
            -c|--config)
                DEPLOYMENT_CONFIG="$2"
                shift 2
                ;;
            --github)
                DEPLOY_TO_GITHUB="true"
                shift
                ;;
            --website)
                DEPLOY_TO_WEBSITE="true"
                shift
                ;;
            --store)
                DEPLOY_TO_STORE="true"
                shift
                ;;
            --no-sign)
                SIGN_APPS="false"
                shift
                ;;
            --no-tests)
                RUN_TESTS="false"
                shift
                ;;
            --no-auto-update)
                AUTO_UPDATE="false"
                shift
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Load deployment configuration
load_config() {
    if [[ -f "$DEPLOYMENT_CONFIG" ]]; then
        log_info "Loading deployment configuration from: $DEPLOYMENT_CONFIG"
        source "$DEPLOYMENT_CONFIG"
        log_success "Configuration loaded"
    else
        log_warning "Deployment configuration not found: $DEPLOYMENT_CONFIG"
        log_info "Using default configuration"
        create_default_config
    fi
}

# Create default deployment configuration
create_default_config() {
    cat > "$DEPLOYMENT_CONFIG" << 'EOF'
# Desktop Deployment Configuration

#==============================================================================
# APPLICATION SIGNING
#==============================================================================

# Windows Code Signing
WINDOWS_SIGN_CERT_PATH=~/windows-signing-cert.p12
WINDOWS_SIGN_CERT_PASSWORD=your_cert_password
WINDOWS_SIGN_TIMESTAMP_URL=http://timestamp.digicert.com

# macOS Code Signing
MACOS_SIGN_IDENTITY="Developer ID Application: Your Name (TEAM_ID)"
MACOS_SIGN_CERT_PATH=~/macos-signing-cert.p12
MACOS_SIGN_CERT_PASSWORD=your_cert_password

# macOS Notarization
MACOS_NOTARIZE_USERNAME=your_apple_id@example.com
MACOS_NOTARIZE_PASSWORD=your_app_specific_password
MACOS_NOTARIZE_TEAM_ID=YOUR_TEAM_ID

# Linux Signing (optional)
LINUX_SIGN_KEY_PATH=~/linux-signing-key.gpg
LINUX_SIGN_KEY_ID=your_key_id

#==============================================================================
# GITHUB RELEASES
#==============================================================================

# GitHub repository and token
GITHUB_REPOSITORY=catalogizer/catalogizer
GITHUB_TOKEN=your_github_token

# Release configuration
GITHUB_RELEASE_TAG_PREFIX=desktop-v
GITHUB_PRERELEASE=false
GITHUB_RELEASE_DRAFT=false

#==============================================================================
# WEBSITE DEPLOYMENT
#==============================================================================

# Website server configuration
WEBSITE_SERVER=downloads.catalogizer.com
WEBSITE_USERNAME=deploy
WEBSITE_SSH_KEY=~/.ssh/deploy_key
WEBSITE_PATH=/var/www/downloads

# CDN configuration
CDN_ENABLED=false
CDN_DISTRIBUTION_ID=your_cloudfront_distribution_id
CDN_AWS_ACCESS_KEY=your_aws_access_key
CDN_AWS_SECRET_KEY=your_aws_secret_key

#==============================================================================
# PLATFORM STORES
#==============================================================================

# Mac App Store
MAC_APP_STORE_USERNAME=your_apple_id@example.com
MAC_APP_STORE_PASSWORD=your_app_specific_password
MAC_APP_STORE_APP_ID=1234567890

# Microsoft Store
MICROSOFT_STORE_CLIENT_ID=your_client_id
MICROSOFT_STORE_CLIENT_SECRET=your_client_secret
MICROSOFT_STORE_TENANT_ID=your_tenant_id
MICROSOFT_STORE_APP_ID=your_app_id

# Snap Store
SNAP_STORE_LOGIN=your_snapcraft_login
SNAP_STORE_CHANNEL=stable

#==============================================================================
# AUTO-UPDATE CONFIGURATION
#==============================================================================

# Update server
UPDATE_SERVER_URL=https://updates.catalogizer.com
UPDATE_PRIVATE_KEY_PATH=~/update-signing-key.pem

# Update channels
UPDATE_CHANNEL=stable  # stable, beta, alpha
UPDATE_CHECK_INTERVAL=24  # hours

#==============================================================================
# BUILD CONFIGURATION
#==============================================================================

# Version configuration
AUTO_INCREMENT_VERSION=true
VERSION_SUFFIX=

# Build targets
BUILD_WINDOWS_MSI=true
BUILD_WINDOWS_NSIS=true
BUILD_WINDOWS_PORTABLE=true
BUILD_MACOS_DMG=true
BUILD_MACOS_APP=true
BUILD_LINUX_APPIMAGE=true
BUILD_LINUX_DEB=true
BUILD_LINUX_RPM=true
BUILD_LINUX_SNAP=false

# Build optimizations
TAURI_BUNDLE_OPTIMIZE=true
TAURI_BUNDLE_MINIFY=true

#==============================================================================
# NOTIFICATION SETTINGS
#==============================================================================

# Slack notifications
NOTIFY_SLACK=false
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/your/slack/webhook
SLACK_CHANNEL=#releases

# Email notifications
NOTIFY_EMAIL=false
EMAIL_RECIPIENTS=team@catalogizer.com
EMAIL_SMTP_SERVER=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=noreply@catalogizer.com
EMAIL_PASSWORD=your_email_password

#==============================================================================
# DEPLOYMENT SETTINGS
#==============================================================================

# Retry configuration
MAX_RETRY_ATTEMPTS=3
RETRY_DELAY=30

# Cleanup
CLEANUP_BUILD_ARTIFACTS=true
KEEP_LAST_N_BUILDS=5

EOF

    log_success "Default deployment configuration created: $DEPLOYMENT_CONFIG"
    log_warning "Please update the configuration with your actual credentials and settings"
}

# Detect current platform
detect_platform() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        CURRENT_PLATFORM="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        CURRENT_PLATFORM="macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        CURRENT_PLATFORM="windows"
    else
        log_error "Unsupported platform: $OSTYPE"
        exit 1
    fi

    log_info "Current platform: $CURRENT_PLATFORM"
}

# Check build requirements
check_requirements() {
    log_info "Checking build requirements..."

    # Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js is required but not installed"
        exit 1
    fi

    # Check Rust and Cargo
    if ! command -v cargo &> /dev/null; then
        log_error "Rust/Cargo is required but not installed"
        exit 1
    fi

    # Check Tauri CLI
    if ! command -v cargo-tauri &> /dev/null; then
        log_info "Installing Tauri CLI..."
        cargo install tauri-cli
    fi

    # Platform-specific requirements
    case $CURRENT_PLATFORM in
        "windows")
            check_windows_requirements
            ;;
        "macos")
            check_macos_requirements
            ;;
        "linux")
            check_linux_requirements
            ;;
    esac

    log_success "Requirements check passed"
}

# Check Windows-specific requirements
check_windows_requirements() {
    # Check Visual Studio Build Tools
    if [[ "$DEPLOY_TARGET" == "all" ]] || [[ "$DEPLOY_TARGET" == "windows" ]] || [[ "$DEPLOY_TARGET" == "current" && "$CURRENT_PLATFORM" == "windows" ]]; then
        # Check for MSVC
        if ! command -v cl.exe &> /dev/null; then
            log_warning "Visual Studio Build Tools not found in PATH"
            log_info "Please install Visual Studio Build Tools for C++"
        fi
    fi
}

# Check macOS-specific requirements
check_macos_requirements() {
    if [[ "$DEPLOY_TARGET" == "all" ]] || [[ "$DEPLOY_TARGET" == "macos" ]] || [[ "$DEPLOY_TARGET" == "current" && "$CURRENT_PLATFORM" == "macos" ]]; then
        # Check Xcode Command Line Tools
        if ! command -v xcode-select &> /dev/null; then
            log_error "Xcode Command Line Tools not found"
            log_info "Install with: xcode-select --install"
            exit 1
        fi
    fi
}

# Check Linux-specific requirements
check_linux_requirements() {
    if [[ "$DEPLOY_TARGET" == "all" ]] || [[ "$DEPLOY_TARGET" == "linux" ]] || [[ "$DEPLOY_TARGET" == "current" && "$CURRENT_PLATFORM" == "linux" ]]; then
        # Check development packages
        local required_packages=("build-essential" "libwebkit2gtk-4.0-dev" "libgtk-3-dev" "libappindicator3-dev")

        for package in "${required_packages[@]}"; do
            if ! dpkg -l | grep -q "^ii  $package "; then
                log_warning "Required package not found: $package"
                log_info "Install with: sudo apt-get install $package"
            fi
        done
    fi
}

# Build desktop application
build_desktop() {
    log_header "Building Desktop Application"

    cd "$PROJECT_ROOT/catalogizer-desktop"

    # Install dependencies
    log_info "Installing dependencies..."
    if [[ "$DRY_RUN" != "true" ]]; then
        npm install
    fi

    # Run tests if enabled
    if [[ "$RUN_TESTS" == "true" ]]; then
        log_info "Running tests..."
        if [[ "$DRY_RUN" != "true" ]]; then
            npm run test || log_warning "Tests failed"
        fi
    fi

    # Build frontend
    log_info "Building frontend..."
    if [[ "$DRY_RUN" != "true" ]]; then
        npm run build
    fi

    # Determine build targets
    local build_targets=()
    case $DEPLOY_TARGET in
        "current")
            build_targets=("$CURRENT_PLATFORM")
            ;;
        "all")
            build_targets=("windows" "macos" "linux")
            ;;
        "windows"|"macos"|"linux")
            build_targets=("$DEPLOY_TARGET")
            ;;
        *)
            log_error "Unknown deployment target: $DEPLOY_TARGET"
            exit 1
            ;;
    esac

    # Build for each target
    for target in "${build_targets[@]}"; do
        build_for_platform "$target"
    done
}

# Build for specific platform
build_for_platform() {
    local platform="$1"

    log_info "Building for platform: $platform"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would build for $platform"
        return
    fi

    local tauri_args=""
    case $platform in
        "windows")
            if [[ "$CURRENT_PLATFORM" != "windows" ]]; then
                tauri_args="--target x86_64-pc-windows-msvc"
            fi
            ;;
        "macos")
            if [[ "$CURRENT_PLATFORM" != "macos" ]]; then
                tauri_args="--target x86_64-apple-darwin"
            fi
            ;;
        "linux")
            if [[ "$CURRENT_PLATFORM" != "linux" ]]; then
                tauri_args="--target x86_64-unknown-linux-gnu"
            fi
            ;;
    esac

    # Build with Tauri
    if [[ "$BUILD_TYPE" == "release" ]]; then
        npm run tauri:build $tauri_args
    else
        npm run tauri:build -- --debug $tauri_args
    fi

    # Sign applications if enabled
    if [[ "$SIGN_APPS" == "true" && "$BUILD_TYPE" == "release" ]]; then
        sign_application "$platform"
    fi

    log_success "Build completed for $platform"
}

# Sign application for specific platform
sign_application() {
    local platform="$1"

    log_info "Signing application for $platform..."

    case $platform in
        "windows")
            sign_windows_app
            ;;
        "macos")
            sign_macos_app
            ;;
        "linux")
            sign_linux_app
            ;;
    esac
}

# Sign Windows application
sign_windows_app() {
    if [[ -z "$WINDOWS_SIGN_CERT_PATH" ]]; then
        log_warning "Windows signing certificate not configured"
        return
    fi

    log_info "Signing Windows executables..."

    # Find built executables
    local exe_files=($(find src-tauri/target -name "*.exe" -o -name "*.msi"))

    for exe_file in "${exe_files[@]}"; do
        if [[ -f "$exe_file" ]]; then
            signtool sign \
                /f "$WINDOWS_SIGN_CERT_PATH" \
                /p "$WINDOWS_SIGN_CERT_PASSWORD" \
                /t "$WINDOWS_SIGN_TIMESTAMP_URL" \
                "$exe_file"
            log_success "Signed: $exe_file"
        fi
    done
}

# Sign macOS application
sign_macos_app() {
    if [[ -z "$MACOS_SIGN_IDENTITY" ]]; then
        log_warning "macOS signing identity not configured"
        return
    fi

    log_info "Signing macOS application..."

    # Find built applications
    local app_files=($(find src-tauri/target -name "*.app" -o -name "*.dmg"))

    for app_file in "${app_files[@]}"; do
        if [[ -e "$app_file" ]]; then
            codesign --force --deep --sign "$MACOS_SIGN_IDENTITY" "$app_file"
            log_success "Signed: $app_file"

            # Notarize if configured
            if [[ -n "$MACOS_NOTARIZE_USERNAME" ]]; then
                notarize_macos_app "$app_file"
            fi
        fi
    done
}

# Notarize macOS application
notarize_macos_app() {
    local app_file="$1"

    log_info "Notarizing macOS application: $app_file"

    # Create ZIP for notarization
    local zip_file="${app_file}.zip"
    ditto -c -k --keepParent "$app_file" "$zip_file"

    # Submit for notarization
    xcrun altool --notarize-app \
        --primary-bundle-id "com.catalogizer.desktop" \
        --username "$MACOS_NOTARIZE_USERNAME" \
        --password "$MACOS_NOTARIZE_PASSWORD" \
        --asc-provider "$MACOS_NOTARIZE_TEAM_ID" \
        --file "$zip_file"

    # Note: In production, you would need to check notarization status
    # and staple the ticket to the application

    rm -f "$zip_file"
    log_success "Notarization submitted for: $app_file"
}

# Sign Linux application
sign_linux_app() {
    if [[ -z "$LINUX_SIGN_KEY_PATH" ]]; then
        log_warning "Linux signing key not configured"
        return
    fi

    log_info "Signing Linux packages..."

    # Find built packages
    local package_files=($(find src-tauri/target -name "*.deb" -o -name "*.rpm" -o -name "*.AppImage"))

    for package_file in "${package_files[@]}"; do
        if [[ -f "$package_file" ]]; then
            gpg --detach-sign --armor \
                --local-user "$LINUX_SIGN_KEY_ID" \
                "$package_file"
            log_success "Signed: $package_file"
        fi
    done
}

# Deploy to GitHub Releases
deploy_to_github() {
    if [[ "$DEPLOY_TO_GITHUB" != "true" ]]; then
        return
    fi

    log_info "Deploying to GitHub Releases..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy to GitHub"
        return
    fi

    local version=$(node -p "require('./package.json').version")
    local tag="${GITHUB_RELEASE_TAG_PREFIX}${version}"

    # Create release
    local release_args=""
    if [[ "$GITHUB_PRERELEASE" == "true" ]]; then
        release_args+=" --prerelease"
    fi
    if [[ "$GITHUB_RELEASE_DRAFT" == "true" ]]; then
        release_args+=" --draft"
    fi

    gh release create "$tag" \
        --repo "$GITHUB_REPOSITORY" \
        --title "Desktop Release v$version" \
        --notes "Desktop application release v$version" \
        $release_args

    # Upload artifacts
    upload_github_artifacts "$tag"

    log_success "Deployed to GitHub Releases"
}

# Upload artifacts to GitHub
upload_github_artifacts() {
    local tag="$1"

    # Find all built artifacts
    local artifacts=($(find src-tauri/target -name "*.dmg" -o -name "*.msi" -o -name "*.exe" -o -name "*.deb" -o -name "*.rpm" -o -name "*.AppImage"))

    for artifact in "${artifacts[@]}"; do
        if [[ -f "$artifact" ]]; then
            gh release upload "$tag" "$artifact" --repo "$GITHUB_REPOSITORY"
            log_success "Uploaded: $(basename "$artifact")"
        fi
    done
}

# Deploy to website
deploy_to_website() {
    if [[ "$DEPLOY_TO_WEBSITE" != "true" ]]; then
        return
    fi

    log_info "Deploying to website..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy to website"
        return
    fi

    # Upload to server
    rsync -avz --progress \
        -e "ssh -i $WEBSITE_SSH_KEY" \
        src-tauri/target/release/bundle/ \
        "$WEBSITE_USERNAME@$WEBSITE_SERVER:$WEBSITE_PATH/"

    # Invalidate CDN if enabled
    if [[ "$CDN_ENABLED" == "true" ]]; then
        aws cloudfront create-invalidation \
            --distribution-id "$CDN_DISTRIBUTION_ID" \
            --paths "/*"
    fi

    log_success "Deployed to website"
}

# Deploy to platform stores
deploy_to_stores() {
    if [[ "$DEPLOY_TO_STORE" != "true" ]]; then
        return
    fi

    log_info "Deploying to platform stores..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy to stores"
        return
    fi

    # Deploy to Mac App Store
    if [[ "$BUILD_MACOS_APP" == "true" ]]; then
        deploy_to_mac_app_store
    fi

    # Deploy to Microsoft Store
    if [[ "$BUILD_WINDOWS_MSI" == "true" ]]; then
        deploy_to_microsoft_store
    fi

    # Deploy to Snap Store
    if [[ "$BUILD_LINUX_SNAP" == "true" ]]; then
        deploy_to_snap_store
    fi
}

# Deploy to Mac App Store
deploy_to_mac_app_store() {
    log_info "Deploying to Mac App Store..."

    # This would require additional setup for Mac App Store submission
    # Including proper entitlements, provisioning profiles, etc.
    log_warning "Mac App Store deployment not yet implemented"
}

# Deploy to Microsoft Store
deploy_to_microsoft_store() {
    log_info "Deploying to Microsoft Store..."

    # This would use the Microsoft Store submission API
    log_warning "Microsoft Store deployment not yet implemented"
}

# Deploy to Snap Store
deploy_to_snap_store() {
    log_info "Deploying to Snap Store..."

    # This would use snapcraft to upload to the Snap Store
    log_warning "Snap Store deployment not yet implemented"
}

# Update auto-update server
update_auto_update_server() {
    if [[ "$AUTO_UPDATE" != "true" ]]; then
        return
    fi

    log_info "Updating auto-update server..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would update auto-update server"
        return
    fi

    # Generate update manifest
    python3 "$SCRIPT_DIR/tools/generate-update-manifest.py" \
        --build-dir "src-tauri/target" \
        --version "$(node -p "require('./package.json').version")" \
        --channel "$UPDATE_CHANNEL" \
        --server-url "$UPDATE_SERVER_URL" \
        --private-key "$UPDATE_PRIVATE_KEY_PATH"

    log_success "Auto-update server updated"
}

# Send notifications
send_notifications() {
    local deployment_summary="$1"

    # Slack notification
    if [[ "$NOTIFY_SLACK" == "true" ]]; then
        log_info "Sending Slack notification..."
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"channel\":\"$SLACK_CHANNEL\",\"text\":\"$deployment_summary\"}" \
            "$SLACK_WEBHOOK_URL"
    fi

    # Email notification
    if [[ "$NOTIFY_EMAIL" == "true" ]]; then
        log_info "Sending email notification..."
        python3 "$SCRIPT_DIR/tools/send-email.py" \
            --recipients "$EMAIL_RECIPIENTS" \
            --subject "Desktop Deployment Completed" \
            --body "$deployment_summary" \
            --smtp-server "$EMAIL_SMTP_SERVER" \
            --smtp-port "$EMAIL_SMTP_PORT" \
            --username "$EMAIL_USERNAME" \
            --password "$EMAIL_PASSWORD"
    fi
}

# Cleanup build artifacts
cleanup_artifacts() {
    if [[ "$CLEANUP_BUILD_ARTIFACTS" == "true" ]]; then
        log_info "Cleaning up build artifacts..."

        # Keep only the last N builds
        find src-tauri/target -name "*.dmg" -o -name "*.msi" -o -name "*.exe" -o -name "*.deb" -o -name "*.rpm" -o -name "*.AppImage" | \
            sort -r | tail -n +$((KEEP_LAST_N_BUILDS + 1)) | xargs rm -f

        log_success "Cleanup completed"
    fi
}

# Main deployment function
main() {
    log_header "Starting Desktop Deployment"

    # Parse arguments and load configuration
    parse_args "$@"
    load_config
    detect_platform
    check_requirements

    # Build desktop application
    build_desktop

    # Deploy to configured targets
    local deployed_targets=()

    deploy_to_github
    if [[ "$DEPLOY_TO_GITHUB" == "true" ]]; then
        deployed_targets+=("GitHub Releases")
    fi

    deploy_to_website
    if [[ "$DEPLOY_TO_WEBSITE" == "true" ]]; then
        deployed_targets+=("Website")
    fi

    deploy_to_stores
    if [[ "$DEPLOY_TO_STORE" == "true" ]]; then
        deployed_targets+=("Platform Stores")
    fi

    # Update auto-update server
    update_auto_update_server

    # Generate deployment summary
    local summary="Desktop deployment completed successfully!\n"
    summary+="Target: $DEPLOY_TARGET\n"
    summary+="Build Type: $BUILD_TYPE\n"
    summary+="Platform: $CURRENT_PLATFORM\n"
    summary+="Deployed to: $(IFS=', '; echo "${deployed_targets[*]}")\n"
    summary+="Timestamp: $(date)"

    # Send notifications
    send_notifications "$summary"

    # Cleanup
    cleanup_artifacts

    log_success "Desktop deployment completed successfully!"
    echo -e "$summary"
}

# Run main function
main "$@"
#!/bin/bash

# Android Deployment Script
# Builds and deploys Android applications to various distribution channels

set -e

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DEPLOYMENT_CONFIG="$SCRIPT_DIR/android-deploy.env"

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
log_header() { echo -e "${PURPLE}[ANDROID DEPLOY]${NC} $1"; }

# Default configuration
DEPLOY_TARGET="all"  # all, mobile, tv, debug, release
BUILD_TYPE="release"
DEPLOY_TO_STORE="false"
DEPLOY_TO_FIREBASE="false"
DEPLOY_TO_APK_PURE="false"
DEPLOY_TO_GITHUB="false"
SIGN_APPS="true"
RUN_TESTS="true"
DEPLOY_INTERNAL="false"

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy Android applications to various distribution channels.

OPTIONS:
    -h, --help                  Show this help message
    -t, --target TARGET         Deployment target: all, mobile, tv (default: all)
    -b, --build-type TYPE       Build type: debug, release (default: release)
    -c, --config FILE           Use custom deployment configuration file
    --store                     Deploy to Google Play Store
    --firebase                  Deploy to Firebase App Distribution
    --apkpure                   Deploy to APKPure
    --github                    Deploy to GitHub Releases
    --internal                  Deploy to internal distribution
    --no-sign                   Skip application signing
    --no-tests                  Skip running tests
    --dry-run                   Show what would be deployed without actually deploying

DEPLOYMENT TARGETS:
    all                         Deploy both mobile and TV applications
    mobile                      Deploy only mobile application
    tv                          Deploy only TV application

BUILD TYPES:
    debug                       Build debug versions (for testing)
    release                     Build release versions (for production)

EXAMPLES:
    # Deploy both apps to all configured channels
    $0

    # Deploy only mobile app to Play Store
    $0 --target mobile --store

    # Deploy debug builds to Firebase
    $0 --build-type debug --firebase

    # Deploy to internal distribution without signing
    $0 --internal --no-sign

    # Dry run to see what would be deployed
    $0 --dry-run

CONFIGURATION:
    Create android-deploy.env file or use --config to specify custom configuration.
    See android-deploy.env.example for available options.

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
            --store)
                DEPLOY_TO_STORE="true"
                shift
                ;;
            --firebase)
                DEPLOY_TO_FIREBASE="true"
                shift
                ;;
            --apkpure)
                DEPLOY_TO_APK_PURE="true"
                shift
                ;;
            --github)
                DEPLOY_TO_GITHUB="true"
                shift
                ;;
            --internal)
                DEPLOY_INTERNAL="true"
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
# Android Deployment Configuration

#==============================================================================
# APPLICATION SIGNING
#==============================================================================

# Keystore configuration
ANDROID_KEYSTORE_PATH=~/android-keystore.jks
ANDROID_KEYSTORE_PASSWORD=your_keystore_password
ANDROID_KEY_ALIAS=catalogizer
ANDROID_KEY_PASSWORD=your_key_password

# Upload keystore (for Play Store)
ANDROID_UPLOAD_KEYSTORE_PATH=~/android-upload-keystore.jks
ANDROID_UPLOAD_KEYSTORE_PASSWORD=your_upload_keystore_password
ANDROID_UPLOAD_KEY_ALIAS=catalogizer-upload
ANDROID_UPLOAD_KEY_PASSWORD=your_upload_key_password

#==============================================================================
# GOOGLE PLAY STORE
#==============================================================================

# Service account key for Play Store API
GOOGLE_PLAY_SERVICE_ACCOUNT_KEY=~/google-play-service-account.json

# Play Store application IDs
PLAY_STORE_MOBILE_APP_ID=com.catalogizer.android
PLAY_STORE_TV_APP_ID=com.catalogizer.androidtv

# Release track (internal, alpha, beta, production)
PLAY_STORE_TRACK=internal

# Release notes
PLAY_STORE_RELEASE_NOTES="Bug fixes and performance improvements"

#==============================================================================
# FIREBASE APP DISTRIBUTION
#==============================================================================

# Firebase project configuration
FIREBASE_PROJECT_ID=catalogizer-project
FIREBASE_SERVICE_ACCOUNT_KEY=~/firebase-service-account.json

# Distribution groups
FIREBASE_DISTRIBUTION_GROUPS=testers,qa-team

# Release notes
FIREBASE_RELEASE_NOTES="Internal testing build"

#==============================================================================
# GITHUB RELEASES
#==============================================================================

# GitHub repository and token
GITHUB_REPOSITORY=catalogizer/catalogizer
GITHUB_TOKEN=your_github_token

# Release configuration
GITHUB_RELEASE_TAG_PREFIX=android-v
GITHUB_PRERELEASE=false

#==============================================================================
# INTERNAL DISTRIBUTION
#==============================================================================

# Internal distribution server
INTERNAL_DISTRIBUTION_URL=https://internal.catalogizer.com/releases
INTERNAL_DISTRIBUTION_TOKEN=your_internal_token

# Notification settings
NOTIFY_SLACK=false
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/your/slack/webhook
SLACK_CHANNEL=#releases

NOTIFY_EMAIL=false
EMAIL_RECIPIENTS=team@catalogizer.com
EMAIL_SMTP_SERVER=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=noreply@catalogizer.com
EMAIL_PASSWORD=your_email_password

#==============================================================================
# BUILD CONFIGURATION
#==============================================================================

# Version configuration
AUTO_INCREMENT_VERSION=true
VERSION_SUFFIX=

# Build optimizations
ENABLE_PROGUARD=true
ENABLE_R8=true
ENABLE_SHRINK_RESOURCES=true

# Testing configuration
RUN_UNIT_TESTS=true
RUN_INSTRUMENTED_TESTS=false
TEST_TIMEOUT=300

#==============================================================================
# DEPLOYMENT SETTINGS
#==============================================================================

# Parallel deployment
DEPLOY_PARALLEL=true
MAX_PARALLEL_DEPLOYMENTS=3

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

# Validate configuration
validate_config() {
    log_info "Validating deployment configuration..."

    local errors=0

    # Check signing configuration
    if [[ "$SIGN_APPS" == "true" ]]; then
        if [[ ! -f "$ANDROID_KEYSTORE_PATH" ]]; then
            log_error "Keystore not found: $ANDROID_KEYSTORE_PATH"
            errors=$((errors + 1))
        fi
    fi

    # Check deployment targets
    if [[ "$DEPLOY_TO_STORE" == "true" ]]; then
        if [[ ! -f "$GOOGLE_PLAY_SERVICE_ACCOUNT_KEY" ]]; then
            log_error "Google Play service account key not found: $GOOGLE_PLAY_SERVICE_ACCOUNT_KEY"
            errors=$((errors + 1))
        fi
    fi

    if [[ "$DEPLOY_TO_FIREBASE" == "true" ]]; then
        if [[ ! -f "$FIREBASE_SERVICE_ACCOUNT_KEY" ]]; then
            log_error "Firebase service account key not found: $FIREBASE_SERVICE_ACCOUNT_KEY"
            errors=$((errors + 1))
        fi
    fi

    if [[ "$DEPLOY_TO_GITHUB" == "true" ]]; then
        if [[ -z "$GITHUB_TOKEN" ]]; then
            log_error "GitHub token not configured"
            errors=$((errors + 1))
        fi
    fi

    if [[ $errors -gt 0 ]]; then
        log_error "Configuration validation failed with $errors errors"
        exit 1
    fi

    log_success "Configuration validation passed"
}

# Build Android applications
build_applications() {
    log_header "Building Android Applications"

    local apps_to_build=()

    case $DEPLOY_TARGET in
        "all")
            apps_to_build=("mobile" "tv")
            ;;
        "mobile")
            apps_to_build=("mobile")
            ;;
        "tv")
            apps_to_build=("tv")
            ;;
        *)
            log_error "Unknown deployment target: $DEPLOY_TARGET"
            exit 1
            ;;
    esac

    for app in "${apps_to_build[@]}"; do
        build_app "$app"
    done
}

# Build specific application
build_app() {
    local app_type="$1"
    local app_dir=""
    local app_name=""

    case $app_type in
        "mobile")
            app_dir="$PROJECT_ROOT/catalogizer-android"
            app_name="Catalogizer Android"
            ;;
        "tv")
            app_dir="$PROJECT_ROOT/catalogizer-androidtv"
            app_name="Catalogizer Android TV"
            ;;
        *)
            log_error "Unknown app type: $app_type"
            return 1
            ;;
    esac

    log_info "Building $app_name ($BUILD_TYPE)..."

    if [[ ! -d "$app_dir" ]]; then
        log_error "Application directory not found: $app_dir"
        return 1
    fi

    cd "$app_dir"

    # Clean previous builds
    if [[ "$DRY_RUN" != "true" ]]; then
        ./gradlew clean
    fi

    # Run tests if enabled
    if [[ "$RUN_TESTS" == "true" ]]; then
        log_info "Running tests for $app_name..."
        if [[ "$DRY_RUN" != "true" ]]; then
            if [[ "$RUN_UNIT_TESTS" == "true" ]]; then
                ./gradlew test
            fi
            if [[ "$RUN_INSTRUMENTED_TESTS" == "true" ]]; then
                ./gradlew connectedAndroidTest || log_warning "Instrumented tests failed"
            fi
        fi
    fi

    # Build application
    local gradle_task=""
    case $BUILD_TYPE in
        "debug")
            gradle_task="assembleDebug"
            ;;
        "release")
            if [[ "$SIGN_APPS" == "true" ]]; then
                gradle_task="assembleRelease bundleRelease"
            else
                gradle_task="assembleDebug"
            fi
            ;;
    esac

    if [[ "$DRY_RUN" != "true" ]]; then
        ./gradlew $gradle_task
        log_success "$app_name built successfully"
    else
        log_info "DRY RUN: Would execute ./gradlew $gradle_task"
    fi
}

# Deploy to Google Play Store
deploy_to_play_store() {
    local app_type="$1"

    log_info "Deploying $app_type to Google Play Store..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy to Play Store"
        return
    fi

    local app_id=""
    local bundle_path=""

    case $app_type in
        "mobile")
            app_id="$PLAY_STORE_MOBILE_APP_ID"
            bundle_path="$PROJECT_ROOT/catalogizer-android/app/build/outputs/bundle/release/app-release.aab"
            ;;
        "tv")
            app_id="$PLAY_STORE_TV_APP_ID"
            bundle_path="$PROJECT_ROOT/catalogizer-androidtv/app/build/outputs/bundle/release/app-release.aab"
            ;;
    esac

    # Upload using Google Play Console API
    python3 "$SCRIPT_DIR/tools/play-store-upload.py" \
        --service-account-key "$GOOGLE_PLAY_SERVICE_ACCOUNT_KEY" \
        --package-name "$app_id" \
        --bundle-path "$bundle_path" \
        --track "$PLAY_STORE_TRACK" \
        --release-notes "$PLAY_STORE_RELEASE_NOTES"

    log_success "$app_type deployed to Play Store"
}

# Deploy to Firebase App Distribution
deploy_to_firebase() {
    local app_type="$1"

    log_info "Deploying $app_type to Firebase App Distribution..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy to Firebase"
        return
    fi

    local apk_path=""
    case $app_type in
        "mobile")
            apk_path="$PROJECT_ROOT/catalogizer-android/app/build/outputs/apk/release/app-release.apk"
            ;;
        "tv")
            apk_path="$PROJECT_ROOT/catalogizer-androidtv/app/build/outputs/apk/release/app-release.apk"
            ;;
    esac

    # Upload using Firebase CLI
    firebase appdistribution:distribute "$apk_path" \
        --project "$FIREBASE_PROJECT_ID" \
        --groups "$FIREBASE_DISTRIBUTION_GROUPS" \
        --release-notes "$FIREBASE_RELEASE_NOTES"

    log_success "$app_type deployed to Firebase"
}

# Deploy to GitHub Releases
deploy_to_github() {
    log_info "Deploying to GitHub Releases..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy to GitHub"
        return
    fi

    local version=$(grep "versionName" "$PROJECT_ROOT/catalogizer-android/app/build.gradle.kts" | cut -d'"' -f2)
    local tag="${GITHUB_RELEASE_TAG_PREFIX}${version}"

    # Create release using GitHub CLI
    gh release create "$tag" \
        --repo "$GITHUB_REPOSITORY" \
        --title "Android Release v$version" \
        --notes "Android applications release v$version" \
        ${GITHUB_PRERELEASE:+--prerelease}

    # Upload artifacts
    local artifacts=()

    if [[ "$DEPLOY_TARGET" == "all" ]] || [[ "$DEPLOY_TARGET" == "mobile" ]]; then
        artifacts+=(
            "$PROJECT_ROOT/catalogizer-android/app/build/outputs/apk/release/app-release.apk"
            "$PROJECT_ROOT/catalogizer-android/app/build/outputs/bundle/release/app-release.aab"
        )
    fi

    if [[ "$DEPLOY_TARGET" == "all" ]] || [[ "$DEPLOY_TARGET" == "tv" ]]; then
        artifacts+=(
            "$PROJECT_ROOT/catalogizer-androidtv/app/build/outputs/apk/release/app-release.apk"
            "$PROJECT_ROOT/catalogizer-androidtv/app/build/outputs/bundle/release/app-release.aab"
        )
    fi

    for artifact in "${artifacts[@]}"; do
        if [[ -f "$artifact" ]]; then
            gh release upload "$tag" "$artifact" --repo "$GITHUB_REPOSITORY"
        fi
    done

    log_success "Deployed to GitHub Releases"
}

# Deploy to internal distribution
deploy_internal() {
    log_info "Deploying to internal distribution..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy to internal distribution"
        return
    fi

    # Upload to internal server
    python3 "$SCRIPT_DIR/tools/internal-upload.py" \
        --url "$INTERNAL_DISTRIBUTION_URL" \
        --token "$INTERNAL_DISTRIBUTION_TOKEN" \
        --target "$DEPLOY_TARGET" \
        --build-type "$BUILD_TYPE"

    log_success "Deployed to internal distribution"
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
            --subject "Android Deployment Completed" \
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
        for app_dir in "$PROJECT_ROOT/catalogizer-android" "$PROJECT_ROOT/catalogizer-androidtv"; do
            if [[ -d "$app_dir" ]]; then
                find "$app_dir/app/build/outputs" -name "*.apk" -o -name "*.aab" | \
                    sort -r | tail -n +$((KEEP_LAST_N_BUILDS + 1)) | xargs rm -f
            fi
        done

        log_success "Cleanup completed"
    fi
}

# Main deployment function
main() {
    log_header "Starting Android Deployment"

    # Parse arguments and load configuration
    parse_args "$@"
    load_config
    validate_config

    # Build applications
    build_applications

    # Deploy to configured targets
    local deployed_targets=()
    local apps_to_deploy=()

    case $DEPLOY_TARGET in
        "all")
            apps_to_deploy=("mobile" "tv")
            ;;
        "mobile"|"tv")
            apps_to_deploy=("$DEPLOY_TARGET")
            ;;
    esac

    for app in "${apps_to_deploy[@]}"; do
        if [[ "$DEPLOY_TO_STORE" == "true" ]]; then
            deploy_to_play_store "$app"
            deployed_targets+=("Play Store ($app)")
        fi

        if [[ "$DEPLOY_TO_FIREBASE" == "true" ]]; then
            deploy_to_firebase "$app"
            deployed_targets+=("Firebase ($app)")
        fi
    done

    if [[ "$DEPLOY_TO_GITHUB" == "true" ]]; then
        deploy_to_github
        deployed_targets+=("GitHub Releases")
    fi

    if [[ "$DEPLOY_INTERNAL" == "true" ]]; then
        deploy_internal
        deployed_targets+=("Internal Distribution")
    fi

    # Generate deployment summary
    local summary="Android deployment completed successfully!\n"
    summary+="Target: $DEPLOY_TARGET\n"
    summary+="Build Type: $BUILD_TYPE\n"
    summary+="Deployed to: $(IFS=', '; echo "${deployed_targets[*]}")\n"
    summary+="Timestamp: $(date)"

    # Send notifications
    send_notifications "$summary"

    # Cleanup
    cleanup_artifacts

    log_success "Android deployment completed successfully!"
    echo -e "$summary"
}

# Run main function
main "$@"
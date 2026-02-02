#!/bin/bash

# Catalogizer Installation Script
# This script installs and configures the complete Catalogizer ecosystem

set -e

# Script information
SCRIPT_VERSION="1.0.0"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_header() { echo -e "${PURPLE}[CATALOGIZER]${NC} $1"; }

# Configuration
DEFAULT_ENV_FILE="$PROJECT_ROOT/.env"
CUSTOM_ENV_FILE=""
INSTALL_MODE="full"  # full, server-only, clients-only
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yml"
OVERRIDE_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.override.yml"

# Container runtime variables (set by detect_container_runtime)
CONTAINER_CMD=""
COMPOSE_CMD=""

# Container runtime detection - prefer podman over docker
detect_container_runtime() {
    if command -v podman &>/dev/null; then
        CONTAINER_CMD="podman"
        if command -v podman-compose &>/dev/null; then
            COMPOSE_CMD="podman-compose"
        else
            log_warning "podman found but podman-compose is not installed"
            COMPOSE_CMD=""
        fi
    elif command -v docker &>/dev/null; then
        CONTAINER_CMD="docker"
        if command -v docker-compose &>/dev/null; then
            COMPOSE_CMD="docker-compose"
        elif docker compose version &>/dev/null 2>&1; then
            COMPOSE_CMD="docker compose"
        else
            log_warning "docker found but docker-compose is not installed"
            COMPOSE_CMD=""
        fi
    else
        CONTAINER_CMD=""
        COMPOSE_CMD=""
    fi
}

# Show banner
show_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
   ____      _        _             _
  / ___|__ _| |_ __ _| | ___   __ _(_)_______ _ __
 | |   / _` | __/ _` | |/ _ \ / _` | |_  / _ \ '__|
 | |__| (_| | || (_| | | (_) | (_| | |/ /  __/ |
  \____\__,_|\__\__,_|_|\___/ \__, |_/___\___|_|
                              |___/
EOF
    echo -e "${NC}"
    echo -e "${PURPLE}Catalogizer Installation Script v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}Complete media management ecosystem installer${NC}"
    echo ""
}

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -e, --env-file FILE     Use custom environment file (default: .env)
    -m, --mode MODE         Installation mode: full, server-only, clients-only (default: full)
    -v, --version           Show version information
    --dry-run               Show what would be installed without actually installing
    --skip-docker-check     Skip Docker/Podman installation check
    --skip-deps             Skip dependency installation
    --force                 Force installation even if components exist

INSTALLATION MODES:
    full                    Install server + build and package all clients
    server-only             Install only the server components (Docker)
    clients-only            Build and package only client applications
    development             Install in development mode with hot-reload

EXAMPLES:
    # Full installation with default configuration
    $0

    # Server-only installation with custom environment
    $0 --mode server-only --env-file production.env

    # Development setup
    $0 --mode development

    # Build only client applications
    $0 --mode clients-only

ENVIRONMENT FILE:
    Create a .env file or specify a custom one with --env-file.
    See .env.example for available configuration options.

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
            -e|--env-file)
                CUSTOM_ENV_FILE="$2"
                shift 2
                ;;
            -m|--mode)
                INSTALL_MODE="$2"
                shift 2
                ;;
            -v|--version)
                echo "Catalogizer Installer v$SCRIPT_VERSION"
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --skip-docker-check)
                SKIP_DOCKER_CHECK=true
                shift
                ;;
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --force)
                FORCE_INSTALL=true
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

# Load environment configuration
load_environment() {
    local env_file="${CUSTOM_ENV_FILE:-$DEFAULT_ENV_FILE}"

    if [[ -f "$env_file" ]]; then
        log_info "Loading environment from: $env_file"
        set -a  # Automatically export all variables
        source "$env_file"
        set +a
        log_success "Environment loaded successfully"
    else
        log_warning "Environment file not found: $env_file"
        log_info "Creating default environment file..."
        create_default_env
    fi
}

# Create default environment file
create_default_env() {
    cat > "$DEFAULT_ENV_FILE" << 'EOF'
# Catalogizer Environment Configuration
# Copy this file and modify as needed for your environment

#==============================================================================
# GENERAL SETTINGS
#==============================================================================

# Environment mode (development, staging, production)
CATALOGIZER_ENV=production

# Application version
CATALOGIZER_VERSION=1.0.0

# Installation directory
INSTALL_DIR=/opt/catalogizer

# Data directory for media files and databases
DATA_DIR=/var/lib/catalogizer

# Logs directory
LOGS_DIR=/var/log/catalogizer

#==============================================================================
# SERVER CONFIGURATION
#==============================================================================

# Server hostname/IP
CATALOGIZER_HOST=localhost

# API server port
CATALOGIZER_PORT=8080

# WebSocket port
CATALOGIZER_WS_PORT=8081

# Admin interface port
CATALOGIZER_ADMIN_PORT=9090

# Database configuration
DATABASE_TYPE=postgres  # postgres, mysql, sqlite
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=catalogizer
DATABASE_USER=catalogizer
DATABASE_PASSWORD=catalogizer_password

# Redis configuration (for caching and sessions)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

#==============================================================================
# SECURITY SETTINGS
#==============================================================================

# JWT Secret for authentication (change this!)
JWT_SECRET=your-super-secret-jwt-key-change-this

# Admin user credentials (change these!)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin_password
ADMIN_EMAIL=admin@example.com

# SSL/TLS Configuration
SSL_ENABLED=false
SSL_CERT_PATH=/etc/catalogizer/ssl/cert.pem
SSL_KEY_PATH=/etc/catalogizer/ssl/key.pem

#==============================================================================
# MEDIA CONFIGURATION
#==============================================================================

# Default media directories (comma-separated)
MEDIA_DIRECTORIES=/mnt/media,/storage/media

# SMB/CIFS default configuration
SMB_WORKGROUP=WORKGROUP
SMB_DEFAULT_PORT=445

# Media processing
TRANSCODE_ENABLED=true
TRANSCODE_QUALITY=medium  # low, medium, high
THUMBNAIL_GENERATION=true

#==============================================================================
# DOCKER CONFIGURATION
#==============================================================================

# Docker network name
DOCKER_NETWORK=catalogizer-network

# Docker volumes
DOCKER_DATA_VOLUME=catalogizer-data
DOCKER_MEDIA_VOLUME=catalogizer-media
DOCKER_CONFIG_VOLUME=catalogizer-config

# Container restart policy
RESTART_POLICY=unless-stopped

#==============================================================================
# CLIENT BUILD CONFIGURATION
#==============================================================================

# Android build configuration
ANDROID_COMPILE_SDK=34
ANDROID_MIN_SDK=26
ANDROID_TARGET_SDK=34

# Desktop build targets (comma-separated)
DESKTOP_TARGETS=windows,macos,linux

# Build output directory
BUILD_OUTPUT_DIR=./releases

#==============================================================================
# DEPLOYMENT CONFIGURATION
#==============================================================================

# Git repository for updates
GIT_REPOSITORY=https://github.com/catalogizer/catalogizer.git
GIT_BRANCH=main

# Backup configuration
BACKUP_ENABLED=true
BACKUP_SCHEDULE="0 2 * * *"  # Daily at 2 AM
BACKUP_RETENTION_DAYS=30

# Monitoring
MONITORING_ENABLED=false
METRICS_PORT=9091

# Update check
AUTO_UPDATE_CHECK=true
UPDATE_CHANNEL=stable  # stable, beta, nightly

EOF

    log_success "Default environment file created: $DEFAULT_ENV_FILE"
    log_warning "Please review and modify the environment file before continuing"
    log_info "Key settings to update:"
    echo "  - Database credentials"
    echo "  - Admin credentials"
    echo "  - JWT secret"
    echo "  - Media directories"
    echo "  - Host/port configuration"
    echo ""
    read -p "Press Enter to continue after reviewing the configuration..."
}

# Check system requirements
check_requirements() {
    log_info "Checking system requirements..."

    # Check OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        log_success "Operating System: Linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        log_success "Operating System: macOS"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        OS="windows"
        log_success "Operating System: Windows"
    else
        log_error "Unsupported operating system: $OSTYPE"
        exit 1
    fi

    # Check container runtime (for server installation)
    if [[ "$INSTALL_MODE" == "full" ]] || [[ "$INSTALL_MODE" == "server-only" ]] || [[ "$INSTALL_MODE" == "development" ]]; then
        if [[ "$SKIP_DOCKER_CHECK" != "true" ]]; then
            detect_container_runtime

            if [[ -z "$CONTAINER_CMD" ]]; then
                log_error "A container runtime is required but not installed"
                log_info "Please install Docker or Podman"
                log_info "  Docker: https://docs.docker.com/get-docker/"
                log_info "  Podman: https://podman.io/getting-started/installation"
                exit 1
            fi

            if [[ -z "$COMPOSE_CMD" ]]; then
                log_error "A compose tool is required but not installed"
                log_info "Please install docker-compose or podman-compose"
                log_info "  Docker Compose: https://docs.docker.com/compose/install/"
                log_info "  Podman Compose: https://github.com/containers/podman-compose"
                exit 1
            fi

            log_success "Container runtime available: $CONTAINER_CMD (compose: $COMPOSE_CMD)"
        fi
    fi

    # Check build dependencies (for client builds)
    if [[ "$INSTALL_MODE" == "full" ]] || [[ "$INSTALL_MODE" == "clients-only" ]]; then
        check_build_dependencies
    fi
}

# Check build dependencies
check_build_dependencies() {
    log_info "Checking client build dependencies..."

    # Node.js (for desktop app and API client)
    if ! command -v node &> /dev/null; then
        log_warning "Node.js not found - required for desktop and API client builds"
        if [[ "$SKIP_DEPS" != "true" ]]; then
            install_nodejs
        fi
    else
        NODE_VERSION=$(node --version)
        log_success "Node.js found: $NODE_VERSION"
    fi

    # NPM
    if ! command -v npm &> /dev/null; then
        log_error "NPM not found - required for builds"
        exit 1
    fi

    # Java (for Android builds)
    if ! command -v java &> /dev/null; then
        log_warning "Java not found - required for Android builds"
        if [[ "$SKIP_DEPS" != "true" ]]; then
            install_java
        fi
    else
        JAVA_VERSION=$(java -version 2>&1 | head -1)
        log_success "Java found: $JAVA_VERSION"
    fi

    # Rust (for desktop Tauri builds)
    if ! command -v cargo &> /dev/null; then
        log_warning "Rust not found - required for desktop builds"
        if [[ "$SKIP_DEPS" != "true" ]]; then
            install_rust
        fi
    else
        RUST_VERSION=$(cargo --version)
        log_success "Rust found: $RUST_VERSION"
    fi
}

# Install Node.js
install_nodejs() {
    log_info "Installing Node.js..."

    if [[ "$OS" == "linux" ]]; then
        # Install Node.js via NodeSource repository
        curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
        sudo apt-get install -y nodejs
    elif [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            brew install node
        else
            log_error "Homebrew not found. Please install Node.js manually"
            exit 1
        fi
    else
        log_error "Please install Node.js manually for Windows"
        exit 1
    fi
}

# Install Java
install_java() {
    log_info "Installing Java..."

    if [[ "$OS" == "linux" ]]; then
        sudo apt-get update
        sudo apt-get install -y openjdk-17-jdk
    elif [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            brew install openjdk@17
        else
            log_error "Homebrew not found. Please install Java manually"
            exit 1
        fi
    else
        log_error "Please install Java manually for Windows"
        exit 1
    fi
}

# Install Rust
install_rust() {
    log_info "Installing Rust..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source ~/.cargo/env
}

# Install server components
install_server() {
    log_header "Installing Catalogizer Server"

    # Create directories
    create_directories

    # Generate Docker Compose configuration
    generate_docker_compose

    # Start services
    start_services

    # Wait for services to be ready
    wait_for_services

    # Run initial setup
    run_initial_setup
}

# Create necessary directories
create_directories() {
    log_info "Creating directories..."

    local dirs=(
        "$DATA_DIR"
        "$LOGS_DIR"
        "$INSTALL_DIR"
        "$INSTALL_DIR/config"
        "$INSTALL_DIR/ssl"
        "$INSTALL_DIR/backups"
    )

    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            sudo mkdir -p "$dir"
            log_success "Created directory: $dir"
        fi
    done

    # Set permissions
    sudo chown -R $USER:$USER "$INSTALL_DIR" 2>/dev/null || true
}

# Generate Docker Compose configuration
generate_docker_compose() {
    log_info "Generating Docker Compose configuration..."

    # Copy Docker Compose files
    cp "$PROJECT_ROOT/deployment/docker-compose.yml" "$DOCKER_COMPOSE_FILE"

    if [[ -f "$PROJECT_ROOT/deployment/docker-compose.override.yml" ]]; then
        cp "$PROJECT_ROOT/deployment/docker-compose.override.yml" "$OVERRIDE_COMPOSE_FILE"
    fi

    log_success "Docker Compose configuration ready"
}

# Start Docker services
start_services() {
    log_info "Starting Catalogizer services..."

    cd "$PROJECT_ROOT"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would execute: $COMPOSE_CMD up -d"
        return
    fi

    # Pull latest images
    $COMPOSE_CMD pull

    # Start services
    $COMPOSE_CMD up -d

    log_success "Services started successfully"
}

# Wait for services to be ready
wait_for_services() {
    log_info "Waiting for services to be ready..."

    local max_attempts=30
    local attempt=0

    while [[ $attempt -lt $max_attempts ]]; do
        if curl -s "http://${CATALOGIZER_HOST}:${CATALOGIZER_PORT}/health" > /dev/null 2>&1; then
            log_success "Server is ready!"
            return
        fi

        attempt=$((attempt + 1))
        log_info "Waiting for server... (attempt $attempt/$max_attempts)"
        sleep 10
    done

    log_error "Server failed to start within expected time"
    log_info "Check logs with: $COMPOSE_CMD logs"
    exit 1
}

# Run initial setup
run_initial_setup() {
    log_info "Running initial server setup..."

    # Create admin user
    log_info "Creating admin user..."
    $COMPOSE_CMD exec catalogizer-server /app/scripts/create-admin-user.sh \
        "$ADMIN_USERNAME" "$ADMIN_PASSWORD" "$ADMIN_EMAIL"

    # Import initial data if available
    if [[ -f "$PROJECT_ROOT/data/initial-data.sql" ]]; then
        log_info "Importing initial data..."
        $COMPOSE_CMD exec database psql -U "$DATABASE_USER" -d "$DATABASE_NAME" \
            -f /docker-entrypoint-initdb.d/initial-data.sql
    fi

    log_success "Initial setup completed"
}

# Build and package clients
build_clients() {
    log_header "Building Client Applications"

    # API Client Library
    build_api_client

    # Android Apps
    build_android_clients

    # Desktop App
    build_desktop_client

    # Create unified release package
    create_unified_release
}

# Build API client library
build_api_client() {
    log_info "Building API Client Library..."

    cd "$PROJECT_ROOT/catalogizer-api-client"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would build API client library"
        return
    fi

    # Run build script
    if [[ -f "build-scripts/build-release.sh" ]]; then
        ./build-scripts/build-release.sh
        log_success "API Client Library built successfully"
    else
        log_error "API Client build script not found"
    fi
}

# Build Android clients
build_android_clients() {
    log_info "Building Android Applications..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would build Android applications"
        return
    fi

    # Android Mobile
    if [[ -d "$PROJECT_ROOT/catalogizer-android" ]]; then
        log_info "Building Android Mobile App..."
        cd "$PROJECT_ROOT/catalogizer-android"
        if [[ -f "build-scripts/build-release.sh" ]]; then
            ./build-scripts/build-release.sh
            log_success "Android Mobile App built successfully"
        fi
    fi

    # Android TV
    if [[ -d "$PROJECT_ROOT/catalogizer-androidtv" ]]; then
        log_info "Building Android TV App..."
        cd "$PROJECT_ROOT/catalogizer-androidtv"
        if [[ -f "build-scripts/build-release.sh" ]]; then
            ./build-scripts/build-release.sh
            log_success "Android TV App built successfully"
        fi
    fi
}

# Build desktop client
build_desktop_client() {
    log_info "Building Desktop Application..."

    cd "$PROJECT_ROOT/catalogizer-desktop"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would build desktop application"
        return
    fi

    # Run build script
    if [[ -f "build-scripts/build-release.sh" ]]; then
        ./build-scripts/build-release.sh
        log_success "Desktop Application built successfully"
    else
        log_error "Desktop build script not found"
    fi
}

# Create unified release package
create_unified_release() {
    log_info "Creating unified release package..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create unified release package"
        return
    fi

    # Run the unified build script
    if [[ -f "$PROJECT_ROOT/build-scripts/build-all.sh" ]]; then
        cd "$PROJECT_ROOT"
        ./build-scripts/build-all.sh
        log_success "Unified release package created"
    else
        log_error "Unified build script not found"
    fi
}

# Show installation summary
show_summary() {
    echo ""
    log_header "Installation Summary"
    echo ""

    case $INSTALL_MODE in
        "full")
            echo "✅ Server components installed and running"
            echo "✅ Client applications built and packaged"
            echo ""
            echo "🌐 Server Access:"
            echo "   - Web Interface: http://${CATALOGIZER_HOST}:${CATALOGIZER_PORT}"
            echo "   - Admin Panel: http://${CATALOGIZER_HOST}:${CATALOGIZER_ADMIN_PORT}"
            echo "   - API Endpoint: http://${CATALOGIZER_HOST}:${CATALOGIZER_PORT}/api"
            echo ""
            echo "📱 Client Downloads:"
            echo "   - Check the releases/ directory for built applications"
            ;;
        "server-only")
            echo "✅ Server components installed and running"
            echo ""
            echo "🌐 Server Access:"
            echo "   - Web Interface: http://${CATALOGIZER_HOST}:${CATALOGIZER_PORT}"
            echo "   - Admin Panel: http://${CATALOGIZER_HOST}:${CATALOGIZER_ADMIN_PORT}"
            echo "   - API Endpoint: http://${CATALOGIZER_HOST}:${CATALOGIZER_PORT}/api"
            ;;
        "clients-only")
            echo "✅ Client applications built and packaged"
            echo ""
            echo "📱 Client Downloads:"
            echo "   - Check the releases/ directory for built applications"
            ;;
        "development")
            echo "✅ Development environment set up"
            echo ""
            echo "🔧 Development Access:"
            echo "   - Server: http://${CATALOGIZER_HOST}:${CATALOGIZER_PORT}"
            echo "   - Hot reload enabled for client development"
            ;;
    esac

    echo ""
    echo "📋 Next Steps:"
    echo "1. Review the configuration in the admin panel"
    echo "2. Add your media directories and SMB shares"
    echo "3. Install client applications on your devices"
    echo "4. Start organizing your media library!"
    echo ""
    echo "📚 Documentation:"
    echo "   - Installation Guide: ./docs/INSTALLATION.md"
    echo "   - User Manual: ./docs/USER_GUIDE.md"
    echo "   - API Documentation: ./docs/API.md"
    echo ""
    echo "🆘 Support:"
    echo "   - Issues: https://github.com/catalogizer/catalogizer/issues"
    echo "   - Community: https://discord.gg/catalogizer"
    echo ""
}

# Main installation function
main() {
    show_banner
    parse_args "$@"

    log_info "Starting Catalogizer installation..."
    log_info "Installation mode: $INSTALL_MODE"

    # Load configuration
    load_environment

    # Check requirements
    check_requirements

    # Execute based on mode
    case $INSTALL_MODE in
        "full")
            install_server
            build_clients
            ;;
        "server-only")
            install_server
            ;;
        "clients-only")
            build_clients
            ;;
        "development")
            install_server
            log_info "Development mode: Server started with hot-reload"
            ;;
        *)
            log_error "Unknown installation mode: $INSTALL_MODE"
            exit 1
            ;;
    esac

    # Show summary
    show_summary

    log_success "Catalogizer installation completed successfully!"
}

# Run main function with all arguments
main "$@"
#!/bin/bash

# Server Deployment Script
# Deploys Catalogizer server using Docker Compose with various deployment strategies

set -e

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DEPLOYMENT_CONFIG="$SCRIPT_DIR/server-deploy.env"

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
log_header() { echo -e "${PURPLE}[SERVER DEPLOY]${NC} $1"; }

# Container runtime detection - prefer podman over docker
if command -v podman &>/dev/null; then
    CONTAINER_CMD="podman"
    if command -v podman-compose &>/dev/null; then
        COMPOSE_CMD="podman-compose"
    else
        COMPOSE_CMD=""
    fi
elif command -v docker &>/dev/null; then
    CONTAINER_CMD="docker"
    if command -v docker-compose &>/dev/null; then
        COMPOSE_CMD="docker-compose"
    elif docker compose version &>/dev/null 2>&1; then
        COMPOSE_CMD="docker compose"
    else
        COMPOSE_CMD=""
    fi
else
    CONTAINER_CMD=""
    COMPOSE_CMD=""
fi

# Default configuration
DEPLOY_TARGET="production"  # production, staging, development
DEPLOYMENT_STRATEGY="rolling"  # rolling, blue-green, recreate
UPDATE_STRATEGY="pull"  # pull, build, hybrid
BACKUP_BEFORE_DEPLOY="true"
RUN_MIGRATIONS="true"
HEALTH_CHECK="true"
ROLLBACK_ON_FAILURE="true"

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy Catalogizer server using Docker Compose.

OPTIONS:
    -h, --help                  Show this help message
    -t, --target TARGET         Deployment target: production, staging, development
    -s, --strategy STRATEGY     Deployment strategy: rolling, blue-green, recreate
    -u, --update UPDATE         Update strategy: pull, build, hybrid
    -c, --config FILE           Use custom deployment configuration file
    --no-backup                 Skip database backup before deployment
    --no-migrations             Skip running database migrations
    --no-health-check           Skip health checks after deployment
    --no-rollback               Don't rollback on deployment failure
    --force                     Force deployment even if health checks fail
    --dry-run                   Show what would be deployed without actually deploying

DEPLOYMENT TARGETS:
    production                  Production environment with full monitoring
    staging                     Staging environment for testing
    development                 Development environment with debugging

DEPLOYMENT STRATEGIES:
    rolling                     Rolling update with zero downtime
    blue-green                  Blue-green deployment with traffic switching
    recreate                    Stop old, start new (brief downtime)

UPDATE STRATEGIES:
    pull                        Pull latest images from registry
    build                       Build images locally
    hybrid                      Pull for stable, build for development

EXAMPLES:
    # Production deployment with rolling update
    $0 --target production --strategy rolling

    # Staging deployment with local build
    $0 --target staging --update build

    # Development deployment without health checks
    $0 --target development --no-health-check

    # Blue-green production deployment
    $0 --target production --strategy blue-green

    # Dry run to see what would be deployed
    $0 --dry-run

CONFIGURATION:
    Create server-deploy.env file or use --config to specify custom configuration.
    See server-deploy.env.example for available options.

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
            -s|--strategy)
                DEPLOYMENT_STRATEGY="$2"
                shift 2
                ;;
            -u|--update)
                UPDATE_STRATEGY="$2"
                shift 2
                ;;
            -c|--config)
                DEPLOYMENT_CONFIG="$2"
                shift 2
                ;;
            --no-backup)
                BACKUP_BEFORE_DEPLOY="false"
                shift
                ;;
            --no-migrations)
                RUN_MIGRATIONS="false"
                shift
                ;;
            --no-health-check)
                HEALTH_CHECK="false"
                shift
                ;;
            --no-rollback)
                ROLLBACK_ON_FAILURE="false"
                shift
                ;;
            --force)
                FORCE_DEPLOY="true"
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
    # Load base environment
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        source "$PROJECT_ROOT/.env"
    fi

    # Load deployment-specific configuration
    if [[ -f "$DEPLOYMENT_CONFIG" ]]; then
        log_info "Loading deployment configuration from: $DEPLOYMENT_CONFIG"
        source "$DEPLOYMENT_CONFIG"
        log_success "Configuration loaded"
    else
        log_warning "Deployment configuration not found: $DEPLOYMENT_CONFIG"
        log_info "Using default configuration"
        create_default_config
    fi

    # Load target-specific environment
    local target_env="$PROJECT_ROOT/.env.$DEPLOY_TARGET"
    if [[ -f "$target_env" ]]; then
        log_info "Loading target-specific environment: $target_env"
        source "$target_env"
    fi
}

# Create default deployment configuration
create_default_config() {
    cat > "$DEPLOYMENT_CONFIG" << 'EOF'
# Server Deployment Configuration

#==============================================================================
# DEPLOYMENT SETTINGS
#==============================================================================

# Docker registry configuration
DOCKER_REGISTRY=catalogizer
DOCKER_REGISTRY_USERNAME=
DOCKER_REGISTRY_PASSWORD=
DOCKER_REGISTRY_EMAIL=

# Image tags
SERVER_IMAGE_TAG=latest
WEB_IMAGE_TAG=latest
TRANSCODER_IMAGE_TAG=latest

# Container restart policy
RESTART_POLICY=unless-stopped

#==============================================================================
# HEALTH CHECK CONFIGURATION
#==============================================================================

# Health check settings
HEALTH_CHECK_INTERVAL=30
HEALTH_CHECK_TIMEOUT=10
HEALTH_CHECK_RETRIES=3
HEALTH_CHECK_START_PERIOD=60

# Service health check URLs
SERVER_HEALTH_URL=http://localhost:8080/health
WEB_HEALTH_URL=http://localhost/health

#==============================================================================
# BACKUP CONFIGURATION
#==============================================================================

# Database backup settings
BACKUP_RETENTION_DAYS=30
BACKUP_COMPRESSION=true
BACKUP_ENCRYPTION=false
BACKUP_ENCRYPTION_KEY=

# Backup storage
BACKUP_STORAGE_TYPE=local  # local, s3, gcs, azure
BACKUP_S3_BUCKET=
BACKUP_S3_REGION=
BACKUP_S3_ACCESS_KEY=
BACKUP_S3_SECRET_KEY=

#==============================================================================
# MONITORING AND ALERTING
#==============================================================================

# Monitoring configuration
ENABLE_METRICS=true
METRICS_PORT=9091

# Alerting
ENABLE_ALERTS=true
ALERT_WEBHOOK_URL=
ALERT_SLACK_CHANNEL=#alerts
ALERT_EMAIL_RECIPIENTS=ops@catalogizer.com

#==============================================================================
# LOAD BALANCER CONFIGURATION
#==============================================================================

# Load balancer settings (for blue-green deployments)
LB_HEALTH_CHECK_PATH=/health
LB_HEALTH_CHECK_INTERVAL=5
LB_DRAIN_TIMEOUT=60

# Traffic routing
BLUE_GREEN_SWITCH_DELAY=30
CANARY_PERCENTAGE=10

#==============================================================================
# ROLLBACK CONFIGURATION
#==============================================================================

# Rollback settings
ROLLBACK_ON_HEALTH_FAILURE=true
ROLLBACK_TIMEOUT=300
KEEP_PREVIOUS_VERSIONS=3

#==============================================================================
# NOTIFICATION SETTINGS
#==============================================================================

# Slack notifications
NOTIFY_SLACK=false
SLACK_WEBHOOK_URL=
SLACK_CHANNEL=#deployments

# Email notifications
NOTIFY_EMAIL=false
EMAIL_RECIPIENTS=team@catalogizer.com
EMAIL_SMTP_SERVER=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=noreply@catalogizer.com
EMAIL_PASSWORD=

# Discord notifications
NOTIFY_DISCORD=false
DISCORD_WEBHOOK_URL=

EOF

    log_success "Default deployment configuration created: $DEPLOYMENT_CONFIG"
    log_warning "Please update the configuration with your actual settings"
}

# Validate deployment environment
validate_environment() {
    log_info "Validating deployment environment..."

    # Check container runtime (docker or podman)
    if [[ -z "$CONTAINER_CMD" ]]; then
        log_error "A container runtime is required but neither docker nor podman is installed"
        exit 1
    fi

    if [[ -z "$COMPOSE_CMD" ]]; then
        log_error "A compose tool is required but neither docker-compose, docker compose, nor podman-compose is installed"
        exit 1
    fi

    # Check if container runtime is accessible
    if ! $CONTAINER_CMD info &> /dev/null; then
        log_error "Cannot connect to $CONTAINER_CMD daemon"
        exit 1
    fi

    log_info "Using container runtime: $CONTAINER_CMD"
    log_info "Using compose command: $COMPOSE_CMD"

    # Validate configuration
    local errors=0

    # Check required environment variables
    local required_vars=("CATALOGIZER_HOST" "CATALOGIZER_PORT" "DATABASE_PASSWORD" "JWT_SECRET")
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var}" ]]; then
            log_error "Required environment variable not set: $var"
            errors=$((errors + 1))
        fi
    done

    if [[ $errors -gt 0 ]]; then
        log_error "Environment validation failed with $errors errors"
        exit 1
    fi

    log_success "Environment validation passed"
}

# Backup database before deployment
backup_database() {
    if [[ "$BACKUP_BEFORE_DEPLOY" != "true" ]]; then
        return
    fi

    log_info "Creating database backup before deployment..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create database backup"
        return
    fi

    local backup_timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_filename="catalogizer_backup_${backup_timestamp}.sql"
    local backup_path="/tmp/$backup_filename"

    # Create backup
    $COMPOSE_CMD exec -T database pg_dump \
        -U "$DATABASE_USER" \
        -d "$DATABASE_NAME" \
        --clean --if-exists > "$backup_path"

    # Compress if enabled
    if [[ "$BACKUP_COMPRESSION" == "true" ]]; then
        gzip "$backup_path"
        backup_path="${backup_path}.gz"
        backup_filename="${backup_filename}.gz"
    fi

    # Store backup based on storage type
    case "$BACKUP_STORAGE_TYPE" in
        "local")
            mkdir -p "$PROJECT_ROOT/backups"
            mv "$backup_path" "$PROJECT_ROOT/backups/"
            ;;
        "s3")
            aws s3 cp "$backup_path" "s3://$BACKUP_S3_BUCKET/backups/$backup_filename"
            rm "$backup_path"
            ;;
        # Add other storage types as needed
    esac

    log_success "Database backup created: $backup_filename"
}

# Pull or build Docker images
update_images() {
    log_info "Updating Docker images using strategy: $UPDATE_STRATEGY"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would update images"
        return
    fi

    case "$UPDATE_STRATEGY" in
        "pull")
            pull_images
            ;;
        "build")
            build_images
            ;;
        "hybrid")
            if [[ "$DEPLOY_TARGET" == "development" ]]; then
                build_images
            else
                pull_images
            fi
            ;;
        *)
            log_error "Unknown update strategy: $UPDATE_STRATEGY"
            exit 1
            ;;
    esac
}

# Pull images from registry
pull_images() {
    log_info "Pulling latest images from registry..."

    # Login to registry if credentials provided
    if [[ -n "$DOCKER_REGISTRY_USERNAME" ]]; then
        echo "$DOCKER_REGISTRY_PASSWORD" | $CONTAINER_CMD login "$DOCKER_REGISTRY" \
            --username "$DOCKER_REGISTRY_USERNAME" --password-stdin
    fi

    # Pull images
    $COMPOSE_CMD pull

    log_success "Images pulled successfully"
}

# Build images locally
build_images() {
    log_info "Building images locally..."

    $COMPOSE_CMD build --pull

    log_success "Images built successfully"
}

# Deploy using specified strategy
deploy_application() {
    log_header "Deploying Application"

    case "$DEPLOYMENT_STRATEGY" in
        "rolling")
            deploy_rolling
            ;;
        "blue-green")
            deploy_blue_green
            ;;
        "recreate")
            deploy_recreate
            ;;
        *)
            log_error "Unknown deployment strategy: $DEPLOYMENT_STRATEGY"
            exit 1
            ;;
    esac
}

# Rolling deployment
deploy_rolling() {
    log_info "Performing rolling deployment..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would perform rolling deployment"
        return
    fi

    # Get list of services to update
    local services=("catalogizer-server" "web" "transcoder")

    for service in "${services[@]}"; do
        log_info "Updating service: $service"

        # Update service one by one
        $COMPOSE_CMD up -d --no-deps "$service"

        # Wait for service to be healthy
        wait_for_service_health "$service"

        log_success "Service $service updated successfully"
    done
}

# Blue-green deployment
deploy_blue_green() {
    log_info "Performing blue-green deployment..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would perform blue-green deployment"
        return
    fi

    # Create green environment
    log_info "Creating green environment..."
    $COMPOSE_CMD -f docker-compose.yml -f docker-compose.green.yml up -d

    # Wait for green environment to be healthy
    wait_for_environment_health "green"

    # Switch traffic to green
    log_info "Switching traffic to green environment..."
    switch_traffic_to_green

    # Wait for traffic switch delay
    sleep "$BLUE_GREEN_SWITCH_DELAY"

    # Remove blue environment
    log_info "Removing blue environment..."
    $COMPOSE_CMD -f docker-compose.yml down

    # Rename green to blue
    $COMPOSE_CMD -f docker-compose.green.yml down
    $COMPOSE_CMD up -d

    log_success "Blue-green deployment completed"
}

# Recreate deployment (with downtime)
deploy_recreate() {
    log_info "Performing recreate deployment..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would perform recreate deployment"
        return
    fi

    # Stop all services
    $COMPOSE_CMD down

    # Start all services with new images
    $COMPOSE_CMD up -d

    # Wait for all services to be healthy
    wait_for_environment_health "main"

    log_success "Recreate deployment completed"
}

# Wait for service health
wait_for_service_health() {
    local service="$1"
    local max_attempts=30
    local attempt=0

    log_info "Waiting for $service to be healthy..."

    while [[ $attempt -lt $max_attempts ]]; do
        if $COMPOSE_CMD ps "$service" | grep -q "healthy\|Up"; then
            log_success "$service is healthy"
            return 0
        fi

        attempt=$((attempt + 1))
        log_info "Waiting for $service health... (attempt $attempt/$max_attempts)"
        sleep 10
    done

    log_error "$service failed to become healthy"
    return 1
}

# Wait for environment health
wait_for_environment_health() {
    local environment="$1"

    if [[ "$HEALTH_CHECK" != "true" ]]; then
        log_info "Health checks disabled, skipping..."
        return 0
    fi

    log_info "Performing health checks for $environment environment..."

    # Check server health
    check_service_health "$SERVER_HEALTH_URL" "Server"

    # Check web health
    check_service_health "$WEB_HEALTH_URL" "Web"

    log_success "All health checks passed"
}

# Check individual service health
check_service_health() {
    local url="$1"
    local service_name="$2"
    local max_attempts="$HEALTH_CHECK_RETRIES"
    local attempt=0

    while [[ $attempt -lt $max_attempts ]]; do
        if curl -f -s "$url" > /dev/null 2>&1; then
            log_success "$service_name health check passed"
            return 0
        fi

        attempt=$((attempt + 1))
        log_info "$service_name health check failed (attempt $attempt/$max_attempts)"
        sleep "$HEALTH_CHECK_INTERVAL"
    done

    log_error "$service_name health check failed after $max_attempts attempts"
    return 1
}

# Run database migrations
run_migrations() {
    if [[ "$RUN_MIGRATIONS" != "true" ]]; then
        return
    fi

    log_info "Running database migrations..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would run database migrations"
        return
    fi

    # Wait for database to be ready
    wait_for_database

    # Run migrations
    $COMPOSE_CMD exec catalogizer-server /app/scripts/migrate.sh

    log_success "Database migrations completed"
}

# Wait for database to be ready
wait_for_database() {
    local max_attempts=30
    local attempt=0

    log_info "Waiting for database to be ready..."

    while [[ $attempt -lt $max_attempts ]]; do
        if $COMPOSE_CMD exec database pg_isready -U "$DATABASE_USER" > /dev/null 2>&1; then
            log_success "Database is ready"
            return 0
        fi

        attempt=$((attempt + 1))
        log_info "Waiting for database... (attempt $attempt/$max_attempts)"
        sleep 5
    done

    log_error "Database failed to become ready"
    return 1
}

# Switch traffic to green environment (for blue-green)
switch_traffic_to_green() {
    # This would depend on your load balancer configuration
    # Example for nginx:
    # docker-compose exec nginx nginx -s reload

    # Example for HAProxy:
    # echo "enable server green/server1" | socat stdio tcp4-connect:127.0.0.1:9999

    log_info "Traffic switching implementation depends on your load balancer"
}

# Rollback deployment
rollback_deployment() {
    log_warning "Rolling back deployment..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would rollback deployment"
        return
    fi

    # Get previous version
    local previous_version=$($CONTAINER_CMD images --format "table {{.Repository}}\t{{.Tag}}" | \
        grep "catalogizer/server" | grep -v "latest" | head -1 | awk '{print $2}')

    if [[ -n "$previous_version" ]]; then
        log_info "Rolling back to version: $previous_version"

        # Update image tags to previous version
        export SERVER_IMAGE_TAG="$previous_version"
        export WEB_IMAGE_TAG="$previous_version"

        # Redeploy with previous version
        $COMPOSE_CMD up -d

        # Wait for health checks
        wait_for_environment_health "rollback"

        log_success "Rollback completed successfully"
    else
        log_error "No previous version found for rollback"
        return 1
    fi
}

# Send deployment notifications
send_notifications() {
    local status="$1"
    local message="$2"

    # Slack notification
    if [[ "$NOTIFY_SLACK" == "true" ]]; then
        send_slack_notification "$status" "$message"
    fi

    # Email notification
    if [[ "$NOTIFY_EMAIL" == "true" ]]; then
        send_email_notification "$status" "$message"
    fi

    # Discord notification
    if [[ "$NOTIFY_DISCORD" == "true" ]]; then
        send_discord_notification "$status" "$message"
    fi
}

# Send Slack notification
send_slack_notification() {
    local status="$1"
    local message="$2"

    local color="good"
    if [[ "$status" == "failure" ]]; then
        color="danger"
    elif [[ "$status" == "warning" ]]; then
        color="warning"
    fi

    local payload=$(cat << EOF
{
    "channel": "$SLACK_CHANNEL",
    "attachments": [
        {
            "color": "$color",
            "title": "Catalogizer Deployment $status",
            "text": "$message",
            "fields": [
                {
                    "title": "Environment",
                    "value": "$DEPLOY_TARGET",
                    "short": true
                },
                {
                    "title": "Strategy",
                    "value": "$DEPLOYMENT_STRATEGY",
                    "short": true
                },
                {
                    "title": "Timestamp",
                    "value": "$(date)",
                    "short": false
                }
            ]
        }
    ]
}
EOF
    )

    curl -X POST -H 'Content-type: application/json' \
        --data "$payload" \
        "$SLACK_WEBHOOK_URL"
}

# Send email notification
send_email_notification() {
    local status="$1"
    local message="$2"

    python3 "$SCRIPT_DIR/tools/send-email.py" \
        --recipients "$EMAIL_RECIPIENTS" \
        --subject "Catalogizer Deployment $status" \
        --body "$message" \
        --smtp-server "$EMAIL_SMTP_SERVER" \
        --smtp-port "$EMAIL_SMTP_PORT" \
        --username "$EMAIL_USERNAME" \
        --password "$EMAIL_PASSWORD"
}

# Send Discord notification
send_discord_notification() {
    local status="$1"
    local message="$2"

    local color=65280  # Green
    if [[ "$status" == "failure" ]]; then
        color=16711680  # Red
    elif [[ "$status" == "warning" ]]; then
        color=16776960  # Yellow
    fi

    local payload=$(cat << EOF
{
    "embeds": [
        {
            "title": "Catalogizer Deployment $status",
            "description": "$message",
            "color": $color,
            "fields": [
                {
                    "name": "Environment",
                    "value": "$DEPLOY_TARGET",
                    "inline": true
                },
                {
                    "name": "Strategy",
                    "value": "$DEPLOYMENT_STRATEGY",
                    "inline": true
                }
            ],
            "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)"
        }
    ]
}
EOF
    )

    curl -X POST -H 'Content-type: application/json' \
        --data "$payload" \
        "$DISCORD_WEBHOOK_URL"
}

# Cleanup old images and containers
cleanup_deployment() {
    log_info "Cleaning up old images and containers..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would cleanup old resources"
        return
    fi

    # Remove unused images
    $CONTAINER_CMD image prune -f

    # Remove old versions (keep last N)
    local images_to_remove=$($CONTAINER_CMD images --format "table {{.Repository}}\t{{.Tag}}" | \
        grep "catalogizer/" | grep -v "latest" | tail -n +$((KEEP_PREVIOUS_VERSIONS + 1)) | \
        awk '{print $1":"$2}')

    if [[ -n "$images_to_remove" ]]; then
        echo "$images_to_remove" | xargs $CONTAINER_CMD rmi
    fi

    log_success "Cleanup completed"
}

# Main deployment function
main() {
    log_header "Starting Server Deployment"

    # Parse arguments and load configuration
    parse_args "$@"
    load_config
    validate_environment

    # Start deployment process
    local deployment_start_time=$(date +%s)
    local deployment_success=true

    # Backup database
    backup_database

    # Update images
    update_images

    # Deploy application
    if deploy_application; then
        # Run migrations
        run_migrations

        # Final health check
        if ! wait_for_environment_health "final"; then
            if [[ "$ROLLBACK_ON_FAILURE" == "true" && "$FORCE_DEPLOY" != "true" ]]; then
                rollback_deployment
                deployment_success=false
            fi
        fi
    else
        deployment_success=false
        if [[ "$ROLLBACK_ON_FAILURE" == "true" ]]; then
            rollback_deployment
        fi
    fi

    local deployment_end_time=$(date +%s)
    local deployment_duration=$((deployment_end_time - deployment_start_time))

    # Generate deployment summary
    local status="success"
    local summary="Server deployment completed successfully!"

    if [[ "$deployment_success" != "true" ]]; then
        status="failure"
        summary="Server deployment failed!"
    fi

    summary+="\nTarget: $DEPLOY_TARGET"
    summary+="\nStrategy: $DEPLOYMENT_STRATEGY"
    summary+="\nDuration: ${deployment_duration}s"
    summary+="\nTimestamp: $(date)"

    # Send notifications
    send_notifications "$status" "$summary"

    # Cleanup
    cleanup_deployment

    if [[ "$deployment_success" == "true" ]]; then
        log_success "Server deployment completed successfully!"
        exit 0
    else
        log_error "Server deployment failed!"
        exit 1
    fi
}

# Run main function
main "$@"
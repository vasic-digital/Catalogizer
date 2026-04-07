#!/bin/bash
#
# helixqa-orchestrator.sh - Master orchestration script for HelixQA testing
#
# Wires everything together:
#   1. Validates environment (API, services)
#   2. Connects devices from .devconnect
#   3. Installs/updates APKs on devices
#   4. Runs HelixQA autonomous testing
#   5. Monitors progress in real-time
#   6. Generates consolidated reports
#
# Usage: ./scripts/helixqa-orchestrator.sh [platforms]
#   platforms: android,web,desktop,all (default: all)
#

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HELIXQA_BIN="${PROJECT_ROOT}/HelixQA/bin/helixqa"
DEVCONNECT_SCRIPT="${PROJECT_ROOT}/scripts/devconnect.sh"
QA_RESULTS_DIR="${PROJECT_ROOT}/qa-results"
LOG_DIR="${QA_RESULTS_DIR}/logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SESSION_DIR="${QA_RESULTS_DIR}/session-${TIMESTAMP}"

# Platforms to test
PLATFORMS="${1:-all}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Logging
LOG_FILE="${SESSION_DIR}/orchestrator.log"

log() {
    local level=$1
    local message=$2
    local color=$3
    echo -e "${color}[${level}]${NC} ${message}"
    echo "[$(date -Iseconds)] [${level}] ${message}" >> "$LOG_FILE" 2>/dev/null || true
}

log_info() { log "INFO" "$1" "$GREEN"; }
log_warn() { log "WARN" "$1" "$YELLOW"; }
log_error() { log "ERROR" "$1" "$RED"; }
log_step() { log "STEP" "$1" "$CYAN"; }
log_monitor() { log "MONITOR" "$1" "$BLUE"; }

# Create session directory
init_session() {
    mkdir -p "$SESSION_DIR"/{logs,reports,screenshots}
    mkdir -p "$LOG_DIR"
    log_info "Session directory: $SESSION_DIR"
}

# Validate environment
validate_environment() {
    log_step "=== Phase 1: Environment Validation ==="
    
    # Check API is running
    log_info "Checking API (localhost:8080)..."
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        log_info "✓ API is healthy"
    else
        log_error "✗ API not responding on localhost:8080"
        log_info "Starting API..."
        (cd "$PROJECT_ROOT/catalog-api" && HOST=0.0.0.0 go run main.go > /tmp/api.log 2>&1 &)
        sleep 10
        if curl -s http://localhost:8080/health >/dev/null 2>&1; then
            log_info "✓ API started successfully"
        else
            log_error "✗ Failed to start API"
            exit 1
        fi
    fi
    
    # Check Web is accessible
    log_info "Checking Web (localhost:3000)..."
    if curl -s http://localhost:3000 >/dev/null 2>&1; then
        log_info "✓ Web is accessible"
    else
        log_warn "⚠ Web not responding on localhost:3000"
    fi
    
    # Check HelixQA binary
    log_info "Checking HelixQA..."
    if [[ -x "$HELIXQA_BIN" ]]; then
        log_info "✓ HelixQA found"
    else
        log_error "✗ HelixQA not found at $HELIXQA_BIN"
        exit 1
    fi
    
    # Check .devconnect exists
    if [[ -f "$PROJECT_ROOT/.devconnect" ]]; then
        log_info "✓ .devconnect file found"
    else
        log_warn "⚠ .devconnect not found - no devices will be auto-connected"
    fi
    
    log_info "✓ Environment validation complete"
}

# Connect devices
connect_devices() {
    log_step "=== Phase 2: Device Connection ==="
    
    if [[ ! -f "$PROJECT_ROOT/.devconnect" ]]; then
        log_warn "No .devconnect file, skipping device connection"
        return 0
    fi
    
    log_info "Running devconnect script..."
    if bash "$DEVCONNECT_SCRIPT" 2>&1 | tee -a "$LOG_FILE"; then
        log_info "✓ Device connection complete"
    else
        log_error "✗ Device connection failed"
        return 1
    fi
    
    # Verify devices
    log_info "Connected ADB devices:"
    adb devices | grep -v "List" | grep "device$" | while read -r line; do
        log_info "  - $line"
    done
}

# Install APKs on Android devices
install_apks() {
    log_step "=== Phase 3: APK Installation ==="
    
    if [[ "$PLATFORMS" != "all" && "$PLATFORMS" != "android" ]]; then
        log_info "Skipping APK installation (platforms: $PLATFORMS)"
        return 0
    fi
    
    # Check APKs exist
    TV_APK="$PROJECT_ROOT/catalogizer-androidtv/app/build/outputs/apk/debug/app-debug.apk"
    MOBILE_APK="$PROJECT_ROOT/catalogizer-android/app/build/outputs/apk/debug/app-debug.apk"
    
    if [[ ! -f "$TV_APK" ]]; then
        log_error "✗ Android TV APK not found: $TV_APK"
        log_info "Building Android TV APK..."
        (cd "$PROJECT_ROOT" && podman run --rm --network host --entrypoint="" \
            -v "$PROJECT_ROOT:/project" -w /project/catalogizer-androidtv \
            -e ANDROID_HOME=/opt/android-sdk \
            localhost/catalogizer-builder:latest \
            ./gradlew assembleDebug --no-daemon -x test 2>&1 | tail -20)
    fi
    
    # Get list of connected devices
    local devices
    devices=$(adb devices | grep -v "List" | grep "device$" | awk '{print $1}')
    
    if [[ -z "$devices" ]]; then
        log_warn "No Android devices connected for APK installation"
        return 0
    fi
    
    # Install on each device
    for device in $devices; do
        log_info "Installing APKs on device: $device"
        
        # Check if device is in .devignore
        local model
        model=$(adb -s "$device" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
        if grep -qi "$model" "$PROJECT_ROOT/.devignore" 2>/dev/null; then
            log_warn "  Skipping $device - device $model is in .devignore"
            continue
        fi
        
        # Uninstall old versions
        adb -s "$device" uninstall com.catalogizer.androidtv 2>/dev/null || true
        adb -s "$device" uninstall com.catalogizer.android 2>/dev/null || true
        
        # Install new versions
        if [[ -f "$TV_APK" ]]; then
            if adb -s "$device" install "$TV_APK" 2>&1 | tee -a "$LOG_FILE"; then
                log_info "  ✓ Android TV APK installed"
            else
                log_error "  ✗ Android TV APK install failed"
            fi
        fi
        
        if [[ -f "$MOBILE_APK" ]]; then
            if adb -s "$device" install "$MOBILE_APK" 2>&1 | tee -a "$LOG_FILE"; then
                log_info "  ✓ Android APK installed"
            else
                log_error "  ✗ Android APK install failed"
            fi
        fi
    done
}

# Start monitoring in background
start_monitoring() {
    log_step "=== Phase 4: Starting Monitoring ==="
    
    # Monitor script
    cat > "$SESSION_DIR/monitor.sh" << 'MONITOR_EOF'
#!/bin/bash
SESSION_DIR="$1"
LOG_FILE="$SESSION_DIR/monitor.log"

echo "[$(date -Iseconds)] Monitor started" >> "$LOG_FILE"

while true; do
    # Check API health
    api_status=$(curl -s http://localhost:8080/health 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4 || echo "down")
    echo "[$(date -Iseconds)] API: $api_status" >> "$LOG_FILE"
    
    # Check ADB devices
    adb devices | grep -v "List" | grep "device$" | while read -r device _; do
        echo "[$(date -Iseconds)] Device: $device connected" >> "$LOG_FILE"
    done
    
    # Check HelixQA process
    if pgrep -f "helixqa" > /dev/null; then
        echo "[$(date -Iseconds)] HelixQA: running" >> "$LOG_FILE"
    else
        echo "[$(date -Iseconds)] HelixQA: not running" >> "$LOG_FILE"
    fi
    
    sleep 30
done
MONITOR_EOF
    chmod +x "$SESSION_DIR/monitor.sh"
    
    # Start monitor in background
    nohup bash "$SESSION_DIR/monitor.sh" "$SESSION_DIR" > /dev/null 2>&1 &
    local monitor_pid=$!
    echo $monitor_pid > "$SESSION_DIR/monitor.pid"
    
    log_info "✓ Monitor started (PID: $monitor_pid)"
    log_info "  Log: $SESSION_DIR/monitor.log"
}

# Run HelixQA
run_helixqa() {
    log_step "=== Phase 5: Running HelixQA ==="
    
    local platform_arg="$PLATFORMS"
    if [[ "$PLATFORMS" == "all" ]]; then
        platform_arg="android,web"
    fi
    
    log_info "Platforms: $platform_arg"
    log_info "Timeout: 2h"
    log_info "Output: $SESSION_DIR/helixqa"
    
    mkdir -p "$SESSION_DIR/helixqa"
    
    # Export Android device for HelixQA
    export ANDROID_DEVICE=""
    if [[ -f "$PROJECT_ROOT/.devconnect" ]]; then
        ANDROID_DEVICE=$(grep -v "^#" "$PROJECT_ROOT/.devconnect" | grep -v "^$" | head -1)
        export ANDROID_DEVICE
        log_info "Android device: $ANDROID_DEVICE"
    fi
    
    # Run HelixQA
    log_info "Starting HelixQA autonomous session..."
    
    if "$HELIXQA_BIN" autonomous \
        -platforms "$platform_arg" \
        -project "$PROJECT_ROOT" \
        -output "$SESSION_DIR/helixqa" \
        -timeout 2h \
        -verbose 2>&1 | tee -a "$LOG_FILE"; then
        log_info "✓ HelixQA completed successfully"
    else
        log_error "✗ HelixQA failed or timed out"
        return 1
    fi
}

# Generate consolidated report
generate_report() {
    log_step "=== Phase 6: Generating Report ==="
    
    local report_file="$SESSION_DIR/report.md"
    
    cat > "$report_file" << EOF
# HelixQA Orchestration Report

**Session ID:** $TIMESTAMP  
**Date:** $(date -Iseconds)  
**Platforms:** $PLATFORMS  

## Summary

$(grep -c "INFO\|ERROR\|WARN" "$LOG_FILE" 2>/dev/null | head -1) log entries

## Devices

$(adb devices | grep -v "List" | grep "device$" 2>/dev/null || echo "No devices")

## HelixQA Results

$(ls -la "$SESSION_DIR/helixqa/" 2>/dev/null | tail -n +4 || echo "No results")

## Links

- Session Directory: \`$SESSION_DIR\`
- Full Log: \`$LOG_FILE\`
- Monitor Log: \`$SESSION_DIR/monitor.log\`

## Next Steps

1. Review HelixQA findings in \`$SESSION_DIR/helixqa/\`
2. Check generated issue tickets
3. Address critical issues first
EOF
    
    log_info "✓ Report generated: $report_file"
}

# Stop monitoring
stop_monitoring() {
    if [[ -f "$SESSION_DIR/monitor.pid" ]]; then
        local pid
        pid=$(cat "$SESSION_DIR/monitor.pid")
        if kill "$pid" 2>/dev/null; then
            log_info "✓ Monitor stopped (PID: $pid)"
        fi
    fi
}

# Cleanup on exit
cleanup() {
    log_info "Cleaning up..."
    stop_monitoring
}

trap cleanup EXIT

# Initialize session
init_session

# Main execution
main() {
    echo ""
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║        HelixQA Master Orchestrator                        ║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    log_info "Starting HelixQA orchestration..."
    log_info "Project: $PROJECT_ROOT"
    log_info "Platforms: $PLATFORMS"
    log_info "Session: $TIMESTAMP"
    
    # Run all phases
    init_session
    validate_environment
    connect_devices
    install_apks
    start_monitoring
    run_helixqa
    generate_report
    
    log_step "=== Orchestration Complete ==="
    log_info "Session directory: $SESSION_DIR"
    log_info "Report: $SESSION_DIR/report.md"
    log_info "All phases completed successfully!"
    
    echo ""
    echo -e "${GREEN}✓ HelixQA orchestration finished${NC}"
    echo "  Results: $SESSION_DIR"
    echo ""
}

# Show usage if --help
if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    echo "Usage: $0 [platforms]"
    echo ""
    echo "Platforms:"
    echo "  all       - Test all platforms (default)"
    echo "  android   - Test Android TV apps only"
    echo "  web       - Test Web app only"
    echo "  desktop   - Test Desktop app only"
    echo ""
    echo "Examples:"
    echo "  $0              # Run all tests"
    echo "  $0 android      # Test Android TV only"
    echo "  $0 web          # Test Web only"
    exit 0
fi

main "$@"

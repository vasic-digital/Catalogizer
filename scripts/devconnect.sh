#!/bin/bash
#
# devconnect.sh - Auto-connect Android devices for HelixQA testing
#
# Reads .devconnect file and ensures listed devices are connected via ADB.
# Validates device reachability before attempting connection.
# 
# Usage: ./scripts/devconnect.sh
#

set -e

DEVCONNECT_FILE="${DEVCONNECT_FILE:-.devconnect}"
ADB_TIMEOUT=5
PING_TIMEOUT=2

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if device is reachable via ping
check_reachable() {
    local host=$1
    ping -c 1 -W "$PING_TIMEOUT" "$host" >/dev/null 2>&1
}

# Check if device is already connected via ADB
is_adb_connected() {
    local device=$1
    adb devices | grep -q "^${device}.*device$"
}

# Connect device via ADB
connect_device() {
    local device=$1
    log_info "Connecting to $device..."
    
    if adb -s "$device" reconnect 2>/dev/null || adb connect "$device" 2>/dev/null; then
        sleep 1
        if is_adb_connected "$device"; then
            log_info "✓ Successfully connected to $device"
            return 0
        else
            log_error "✗ Failed to connect to $device (not in device list)"
            return 1
        fi
    else
        log_error "✗ ADB connect failed for $device"
        return 1
    fi
}

# Process a single device
process_device() {
    local device=$1
    local host
    local port
    
    # Parse host:port format
    if [[ "$device" =~ ^([^:]+):([0-9]+)$ ]]; then
        host="${BASH_REMATCH[1]}"
        port="${BASH_REMATCH[2]}"
    else
        host="$device"
        port="5555"
    fi
    
    local full_address="${host}:${port}"
    
    echo ""
    log_info "Processing device: $full_address"
    
    # Check if already connected
    if is_adb_connected "$full_address"; then
        log_info "✓ Device $full_address is already connected"
        
        # Verify it's responsive
        if adb -s "$full_address" shell echo "ping" >/dev/null 2>&1; then
            log_info "✓ Device $full_address is responsive"
            return 0
        else
            log_warn "Device $full_address is connected but not responsive, reconnecting..."
        fi
    fi
    
    # Check reachability
    if ! check_reachable "$host"; then
        log_error "✗ Device $host is not reachable (ping failed)"
        return 1
    fi
    log_info "✓ Device $host is reachable"
    
    # Connect the device
    connect_device "$full_address"
}

# Main execution
main() {
    log_info "Device Connect Script for HelixQA"
    log_info "=================================="
    
    # Check if .devconnect file exists
    if [[ ! -f "$DEVCONNECT_FILE" ]]; then
        log_warn ".devconnect file not found at $DEVCONNECT_FILE"
        log_info "No devices to auto-connect. Skipping."
        exit 0
    fi
    
    # Check if adb is available
    if ! command -v adb >/dev/null 2>&1; then
        log_error "ADB not found in PATH"
        exit 1
    fi
    
    log_info "Reading device list from: $DEVCONNECT_FILE"
    
    local success_count=0
    local fail_count=0
    
    # Read and process each device
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip empty lines and comments
        line=$(echo "$line" | tr -d '[:space:]')
        [[ -z "$line" ]] && continue
        [[ "$line" =~ ^# ]] && continue
        
        if process_device "$line"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
        
    done < "$DEVCONNECT_FILE"
    
    echo ""
    log_info "=================================="
    log_info "Device Connect Summary:"
    log_info "  Success: $success_count"
    log_info "  Failed:  $fail_count"
    
    # List all connected devices
    echo ""
    log_info "Currently connected ADB devices:"
    adb devices | grep -v "List" | grep "device$" || log_warn "No devices connected"
    
    if [[ $fail_count -gt 0 ]]; then
        exit 1
    fi
    
    exit 0
}

main "$@"

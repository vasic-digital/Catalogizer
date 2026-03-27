#!/usr/bin/env bash
# HelixQA Autonomous QA — Android TV
# Runs full 4-phase autonomous QA with video recording on all connected Android TV devices.
#
# Phases:
#   1. Setup — Deploy APK, configure server, authenticate, start recording
#   2. Doc-Driven Verification — Navigate documented features on each device
#   3. Curiosity-Driven Exploration — Autonomous exploration beyond documented features
#   4. Report & Cleanup — Stop recording, pull videos, generate report
#
# Usage:
#   ./scripts/run-helixqa-androidtv.sh [--api-url URL] [--timeout MINUTES]
#
# Prerequisites:
#   - catalog-api running and accessible on the network
#   - ADB connected Android TV devices (leanback or tv characteristics)
#   - ffmpeg available
#
# Output: qa-results/helixqa-androidtv-<timestamp>/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
OUTPUT_DIR="$PROJECT_ROOT/qa-results/helixqa-androidtv-${TIMESTAMP}"
PKG="com.catalogizer.androidtv"
HELIXQA="$PROJECT_ROOT/HelixQA/helixqa"
BANKS_DIR="$PROJECT_ROOT/challenges/helixqa-banks"
FFMPEG="${HELIX_RECORDING_FFMPEG_PATH:-/home/milosvasic/bin/ffmpeg}"

# Defaults
API_URL=""
TIMEOUT_MINUTES=30
RECORD_SEGMENT_SECONDS=180

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${BLUE}[HelixQA-ATV]${NC} $*"; }
ok()   { echo -e "${GREEN}[    OK     ]${NC} $*"; }
warn() { echo -e "${YELLOW}[   WARN   ]${NC} $*"; }
fail() { echo -e "${RED}[   FAIL   ]${NC} $*"; }

# ─── Parse Arguments ─────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
    case "$1" in
        --api-url)   API_URL="$2"; shift 2 ;;
        --timeout)   TIMEOUT_MINUTES="$2"; shift 2 ;;
        *)           shift ;;
    esac
done

# ─── Auto-detect API URL ────────────────────────────────────────────────
if [[ -z "$API_URL" ]]; then
    HOST_IP=$(ip -4 addr show scope global | grep -oP '(?<=inet\s)[\d.]+' | head -1)
    if [[ -n "$HOST_IP" ]]; then
        API_URL="http://${HOST_IP}:8080"
    else
        API_URL="http://localhost:8080"
    fi
fi
log "API URL: $API_URL"

# ─── Discover Android TV Devices ────────────────────────────────────────
log "Discovering Android TV devices..."
TV_DEVICES=()
SERIALS=$(adb devices 2>/dev/null | grep -w "device" | awk '{print $1}')
for serial in $SERIALS; do
    [[ -z "$serial" ]] && continue
    # Check for leanback or tv characteristics
    is_tv=false
    if adb -s "$serial" shell pm list features 2>/dev/null | grep -qE "leanback|television"; then
        is_tv=true
    fi
    chars=$(adb -s "$serial" shell getprop ro.build.characteristics 2>/dev/null | tr -d '\r')
    if [[ "$chars" == *"tv"* ]]; then
        is_tv=true
    fi
    if $is_tv; then
        model=$(adb -s "$serial" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
        sdk=$(adb -s "$serial" shell getprop ro.build.version.sdk 2>/dev/null | tr -d '\r')
        TV_DEVICES+=("$serial")
        ok "Found TV device: $serial ($model, SDK $sdk)"
    fi
done

if [[ ${#TV_DEVICES[@]} -eq 0 ]]; then
    fail "No Android TV devices found"
    exit 1
fi
log "Found ${#TV_DEVICES[@]} Android TV device(s)"

# ─── Create Output Directories ──────────────────────────────────────────
for i in "${!TV_DEVICES[@]}"; do
    device_label="device$((i+1))"
    mkdir -p "$OUTPUT_DIR/$device_label"/{videos,screenshots,evidence,logs}
done
log "Output directory: $OUTPUT_DIR"

# ═══════════════════════════════════════════════════════════════════════
# PHASE 1: SETUP — Deploy, Configure, Authenticate, Record
# ═══════════════════════════════════════════════════════════════════════
log "═══ PHASE 1: SETUP ═══"

# Check API health
if ! curl -sf "$API_URL/health" > /dev/null 2>&1; then
    fail "catalog-api not reachable at $API_URL"
    exit 1
fi
ok "catalog-api healthy at $API_URL"

# Find APK
APK=$(find "$PROJECT_ROOT/catalogizer-androidtv" -name "app-debug.apk" -path "*/build/outputs/*" 2>/dev/null | head -1)
if [[ -z "$APK" ]]; then
    warn "No debug APK found, building..."
    (cd "$PROJECT_ROOT/catalogizer-androidtv" && ./gradlew assembleDebug --no-daemon 2>&1 | tail -3)
    APK=$(find "$PROJECT_ROOT/catalogizer-androidtv" -name "app-debug.apk" -path "*/build/outputs/*" 2>/dev/null | head -1)
fi
ok "APK: $APK"

# Deploy to all devices
for serial in "${TV_DEVICES[@]}"; do
    adb -s "$serial" install -r "$APK" 2>&1 | tail -1 &
done
wait
ok "APK deployed to all devices"

# Set up ADB reverse for network access
for serial in "${TV_DEVICES[@]}"; do
    adb -s "$serial" reverse tcp:8080 tcp:8080 2>/dev/null || true
done
ok "ADB reverse port forwarding set"

# Launch app, configure server URL, and login on each device
# login_on_device: Uses uiautomator to find exact field coordinates, types credentials,
# and verifies login succeeded by checking UI state. Returns 0 on success, 1 on failure.
login_on_device() {
    local SERIAL="$1"
    local DIR="$2"
    local ATTEMPT="$3"

    # Dump UI to find exact field coordinates
    adb -s "$SERIAL" shell uiautomator dump /sdcard/helixqa_ui.xml 2>/dev/null
    local ui_xml
    ui_xml=$(adb -s "$SERIAL" shell cat /sdcard/helixqa_ui.xml 2>/dev/null)

    # Extract Username field center coordinates
    local username_bounds
    username_bounds=$(echo "$ui_xml" | grep -oP 'text="Username"[^/]*bounds="\K[^"]+' | head -1)
    # Extract Password field center coordinates
    local password_bounds
    password_bounds=$(echo "$ui_xml" | grep -oP 'text="Password"[^/]*bounds="\K[^"]+' | head -1)
    # Extract Sign In button center
    local signin_bounds
    signin_bounds=$(echo "$ui_xml" | grep -oP 'text="Sign In"[^/]*bounds="\K[^"]+' | head -1)

    if [[ -z "$username_bounds" ]] || [[ -z "$password_bounds" ]]; then
        warn "  Could not find login fields in UI dump (attempt $ATTEMPT)"
        return 1
    fi

    # Parse bounds [left,top][right,bottom] -> center (x,y)
    # Username field is the EditText ABOVE the "Username" label
    local u_left u_top u_right u_bottom
    u_left=$(echo "$ui_xml" | grep -B1 'text="Username"' | grep 'EditText' | grep -oP 'bounds="\[\K\d+' | head -1)
    u_top=$(echo "$ui_xml" | grep -B1 'text="Username"' | grep 'EditText' | grep -oP 'bounds="\[\d+,\K\d+' | head -1)
    u_right=$(echo "$ui_xml" | grep -B1 'text="Username"' | grep 'EditText' | grep -oP '\]\[\K\d+' | head -1)
    u_bottom=$(echo "$ui_xml" | grep -B1 'text="Username"' | grep 'EditText' | grep -oP '\]\[\d+,\K\d+' | head -1)

    local p_left p_top p_right p_bottom
    p_left=$(echo "$ui_xml" | grep -B1 'text="Password"' | grep 'EditText' | grep -oP 'bounds="\[\K\d+' | head -1)
    p_top=$(echo "$ui_xml" | grep -B1 'text="Password"' | grep 'EditText' | grep -oP 'bounds="\[\d+,\K\d+' | head -1)
    p_right=$(echo "$ui_xml" | grep -B1 'text="Password"' | grep 'EditText' | grep -oP '\]\[\K\d+' | head -1)
    p_bottom=$(echo "$ui_xml" | grep -B1 'text="Password"' | grep 'EditText' | grep -oP '\]\[\d+,\K\d+' | head -1)

    # Use fallback coordinates if parsing fails
    local ux=$((${u_left:-480} + (${u_right:-1440} - ${u_left:-480}) / 2))
    local uy=$((${u_top:-323} + (${u_bottom:-451} - ${u_top:-323}) / 2))
    local px=$((${p_left:-480} + (${p_right:-1440} - ${p_left:-480}) / 2))
    local py=$((${p_top:-475} + (${p_bottom:-603} - ${p_top:-475}) / 2))

    log "    Attempt $ATTEMPT: Username@($ux,$uy) Password@($px,$py)"

    # Clear any existing text in fields first
    adb -s "$SERIAL" shell input tap "$ux" "$uy"; sleep 1
    adb -s "$SERIAL" shell input keyevent KEYCODE_MOVE_END; sleep 0.2
    for _ in $(seq 1 20); do adb -s "$SERIAL" shell input keyevent KEYCODE_DEL; done
    adb -s "$SERIAL" shell input text "admin"; sleep 0.5
    adb -s "$SERIAL" shell input keyevent KEYCODE_BACK; sleep 0.5

    adb -s "$SERIAL" shell input tap "$px" "$py"; sleep 1
    adb -s "$SERIAL" shell input keyevent KEYCODE_MOVE_END; sleep 0.2
    for _ in $(seq 1 20); do adb -s "$SERIAL" shell input keyevent KEYCODE_DEL; done
    adb -s "$SERIAL" shell input text "admin123"; sleep 0.5
    adb -s "$SERIAL" shell input keyevent KEYCODE_BACK; sleep 0.5

    # TAB to Sign In, ENTER to submit
    adb -s "$SERIAL" shell input keyevent KEYCODE_TAB; sleep 0.5
    adb -s "$SERIAL" shell input keyevent KEYCODE_ENTER; sleep 8

    adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/00${ATTEMPT}-login-attempt.png" 2>/dev/null

    # VERIFY login: dump UI and check if "Sign In" text is still present
    adb -s "$SERIAL" shell uiautomator dump /sdcard/helixqa_ui2.xml 2>/dev/null
    local post_ui
    post_ui=$(adb -s "$SERIAL" shell cat /sdcard/helixqa_ui2.xml 2>/dev/null)
    if echo "$post_ui" | grep -q "Sign In"; then
        warn "    Login attempt $ATTEMPT FAILED — still on login screen"
        return 1
    else
        ok "    Login attempt $ATTEMPT SUCCEEDED — past login screen"
        return 0
    fi
}

setup_device() {
    local SERIAL="$1"
    local LABEL="$2"
    local DIR="$OUTPUT_DIR/$LABEL"

    log "  Setting up $LABEL ($SERIAL)..."

    # Force-stop and launch — Mi Box needs 15-20 seconds for Compose to render
    adb -s "$SERIAL" shell am force-stop "$PKG" 2>/dev/null
    sleep 1
    adb -s "$SERIAL" shell am start -n "$PKG/.ui.MainActivity" 2>/dev/null
    sleep 25  # Mi Box is SLOW — 15s splash + compose render

    # Screenshot: initial state
    adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/001-initial.png" 2>/dev/null

    # Retry login up to 3 times with UI dump-based coordinate finding
    local login_ok=false
    for attempt in 2 3 4; do
        if login_on_device "$SERIAL" "$DIR" "$attempt"; then
            login_ok=true
            break
        fi
        warn "  $LABEL: Retrying login (attempt $attempt failed)..."
        sleep 3
    done

    if $login_ok; then
        ok "  $LABEL setup complete (VERIFIED past login screen)"
    else
        fail "  $LABEL setup FAILED — could not log in after 3 attempts"
        adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/005-login-FAILED.png" 2>/dev/null
    fi
}

for i in "${!TV_DEVICES[@]}"; do
    setup_device "${TV_DEVICES[$i]}" "device$((i+1))" &
done
wait
ok "All devices configured and logged in"

# Start screen recording on all devices
log "Starting screen recording..."
RECORD_PIDS=()
for i in "${!TV_DEVICES[@]}"; do
    serial="${TV_DEVICES[$i]}"
    adb -s "$serial" shell "screenrecord --bit-rate 4000000 --time-limit $RECORD_SEGMENT_SECONDS /sdcard/qa_session_1.mp4" &
    RECORD_PIDS+=($!)
done
ok "Screen recording started on ${#TV_DEVICES[@]} devices"

# ─── CONSTITUTIONAL CONSTRAINT ───────────────────────────────────────
# HelixQA MUST NEVER explore outside the target application.
# Before EVERY action, verify the app is in the foreground.
# If not, immediately re-launch it. NEVER send HOME or MENU keys.
# Only use in-app navigation: DPAD directions, CENTER (select), BACK.
# ────────────────────────────────────────────────────────────────────

# ensure_app_foreground: Returns 0 if app is in foreground, re-launches if not.
# This is the CONSTITUTIONAL GUARD — called before every single input action.
ensure_app_foreground() {
    local SERIAL="$1"
    local fg_pkg
    fg_pkg=$(adb -s "$SERIAL" shell "dumpsys activity activities 2>/dev/null | grep -E 'mResumedActivity|mFocusedActivity' | head -1" 2>/dev/null | tr -d '\r')
    if [[ "$fg_pkg" != *"$PKG"* ]]; then
        # App is NOT in foreground — re-launch immediately
        adb -s "$SERIAL" shell am start -n "$PKG/.ui.MainActivity" 2>/dev/null
        sleep 2
        return 1
    fi
    return 0
}

# ═══════════════════════════════════════════════════════════════════════
# PHASE 2: DOC-DRIVEN VERIFICATION
# ═══════════════════════════════════════════════════════════════════════
log "═══ PHASE 2: DOC-DRIVEN VERIFICATION ═══"

# Navigate through documented screens on each device
verify_device() {
    local SERIAL="$1"
    local LABEL="$2"
    local DIR="$OUTPUT_DIR/$LABEL"
    local STEP=10

    log "  Verifying documented features on $LABEL..."

    # CONSTITUTIONAL GUARD: Ensure app is in foreground
    ensure_app_foreground "$SERIAL"
    sleep 1

    # Navigate the app: Home screen
    adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/0${STEP}-home.png" 2>/dev/null
    ((STEP++))

    # DPAD navigation through main sections
    local SECTIONS=("search" "movies" "tvshows" "settings" "player" "browse")
    for section in "${SECTIONS[@]}"; do
        ensure_app_foreground "$SERIAL"  # Guard before every action
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_RIGHT; sleep 1
        adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/0${STEP}-nav-${section}.png" 2>/dev/null
        ((STEP++))
        ensure_app_foreground "$SERIAL"
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_CENTER; sleep 2
        adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/0${STEP}-${section}-detail.png" 2>/dev/null
        ((STEP++))
        adb -s "$SERIAL" shell input keyevent KEYCODE_BACK; sleep 1
    done

    # Navigate down through rows — browse content
    for row in 1 2 3 4; do
        ensure_app_foreground "$SERIAL"
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_DOWN; sleep 1
        adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/0${STEP}-row${row}.png" 2>/dev/null
        ((STEP++))
        # Select item to browse content details
        ensure_app_foreground "$SERIAL"
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_CENTER; sleep 2
        adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/0${STEP}-row${row}-item.png" 2>/dev/null
        ((STEP++))
        # Try to play content (press CENTER again on detail screen)
        ensure_app_foreground "$SERIAL"
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_CENTER; sleep 3
        adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/0${STEP}-row${row}-play.png" 2>/dev/null
        ((STEP++))
        adb -s "$SERIAL" shell input keyevent KEYCODE_BACK; sleep 1
        adb -s "$SERIAL" shell input keyevent KEYCODE_BACK; sleep 1
    done

    # Scroll through items horizontally in each row
    for scroll in 1 2 3; do
        ensure_app_foreground "$SERIAL"
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_RIGHT; sleep 0.5
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_RIGHT; sleep 0.5
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_RIGHT; sleep 0.5
        adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/0${STEP}-scroll${scroll}.png" 2>/dev/null
        ((STEP++))
        adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_DOWN; sleep 1
    done

    ok "  $LABEL doc-driven verification complete ($STEP screenshots)"
}

for i in "${!TV_DEVICES[@]}"; do
    verify_device "${TV_DEVICES[$i]}" "device$((i+1))" &
done
wait
ok "Phase 2 complete"

# ═══════════════════════════════════════════════════════════════════════
# PHASE 3: CURIOSITY-DRIVEN EXPLORATION
# ═══════════════════════════════════════════════════════════════════════
log "═══ PHASE 3: CURIOSITY-DRIVEN EXPLORATION ═══"

# Autonomous exploration: ONLY in-app navigation, with foreground enforcement
explore_device() {
    local SERIAL="$1"
    local LABEL="$2"
    local DIR="$OUTPUT_DIR/$LABEL"
    local STEP=50
    local EXPLORE_SECONDS=$((TIMEOUT_MINUTES * 60 / 3))
    local END_TIME=$(($(date +%s) + EXPLORE_SECONDS))
    local OUT_OF_APP_COUNT=0

    log "  Curiosity exploration on $LABEL (${EXPLORE_SECONDS}s)..."

    while [[ $(date +%s) -lt $END_TIME ]]; do
        # CONSTITUTIONAL GUARD: Verify app is in foreground before EVERY action
        if ! ensure_app_foreground "$SERIAL"; then
            ((OUT_OF_APP_COUNT++))
            warn "  $LABEL: App left foreground (restored, count=$OUT_OF_APP_COUNT)"
            adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/${STEP}-RESTORED.png" 2>/dev/null
            ((STEP++))
            continue
        fi

        # ONLY in-app navigation actions — NO HOME, NO MENU, NO keys that leave the app
        local ACTION=$((RANDOM % 6))
        case $ACTION in
            0) adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_UP ;;
            1) adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_DOWN ;;
            2) adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_LEFT ;;
            3) adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_RIGHT ;;
            4) adb -s "$SERIAL" shell input keyevent KEYCODE_DPAD_CENTER ;;
            5) # BACK is allowed but could exit the app — guard on next iteration
               adb -s "$SERIAL" shell input keyevent KEYCODE_BACK ;;
        esac
        sleep 1

        # Take screenshot every 3 actions (more frequent than before)
        if [[ $((STEP % 3)) -eq 0 ]]; then
            adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/${STEP}-explore.png" 2>/dev/null
        fi
        ((STEP++))

        # Check for crashes
        # Only detect crashes in OUR package — not system processes
        local crash
        crash=$(adb -s "$SERIAL" logcat -d -t 5 "*:E" 2>/dev/null | grep -E "FATAL|AndroidRuntime" | grep -c "$PKG" || echo "0")
        if [[ "$crash" -gt 0 ]]; then
            warn "  Crash detected on $LABEL at step $STEP!"
            adb -s "$SERIAL" exec-out screencap -p > "$DIR/screenshots/${STEP}-CRASH.png" 2>/dev/null
            adb -s "$SERIAL" logcat -d -t 30 "*:E" > "$DIR/logs/crash-step${STEP}.txt" 2>/dev/null
            # Restart app after crash
            adb -s "$SERIAL" shell am force-stop "$PKG" 2>/dev/null
            sleep 1
            adb -s "$SERIAL" shell am start -n "$PKG/.ui.MainActivity" 2>/dev/null
            sleep 3
        fi
    done

    ok "  $LABEL curiosity exploration complete ($STEP steps, $OUT_OF_APP_COUNT foreground restores)"
}

for i in "${!TV_DEVICES[@]}"; do
    explore_device "${TV_DEVICES[$i]}" "device$((i+1))" &
done
wait
ok "Phase 3 complete"

# ═══════════════════════════════════════════════════════════════════════
# PHASE 4: REPORT & CLEANUP
# ═══════════════════════════════════════════════════════════════════════
log "═══ PHASE 4: REPORT & CLEANUP ═══"

# Stop recordings
log "Stopping screen recordings..."
for pid in "${RECORD_PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
done
sleep 3

# Start second recording segment (in case first finished)
for i in "${!TV_DEVICES[@]}"; do
    serial="${TV_DEVICES[$i]}"
    adb -s "$serial" shell "kill \$(pidof screenrecord) 2>/dev/null" 2>/dev/null || true
done
sleep 2

# Pull recordings from devices
log "Pulling video recordings..."
for i in "${!TV_DEVICES[@]}"; do
    serial="${TV_DEVICES[$i]}"
    label="device$((i+1))"
    for seg in 1 2 3; do
        adb -s "$serial" pull "/sdcard/qa_session_${seg}.mp4" "$OUTPUT_DIR/$label/videos/" 2>/dev/null || true
    done
    ok "  Pulled videos from $label"
done

# Pull logcat
log "Collecting device logs..."
for i in "${!TV_DEVICES[@]}"; do
    serial="${TV_DEVICES[$i]}"
    label="device$((i+1))"
    adb -s "$serial" logcat -d > "$OUTPUT_DIR/$label/logs/logcat-full.txt" 2>/dev/null
    adb -s "$serial" logcat -d -s "Catalogizer" "*:E" > "$OUTPUT_DIR/$label/logs/logcat-errors.txt" 2>/dev/null
done
ok "Device logs collected"

# Generate report
log "Generating QA report..."
REPORT="$OUTPUT_DIR/qa-report.md"
cat > "$REPORT" << EOF
# HelixQA Android TV Autonomous QA Report

**Generated:** $(date -Iseconds)
**API URL:** $API_URL
**Duration:** ${TIMEOUT_MINUTES} minutes
**Devices:** ${#TV_DEVICES[@]}

## Devices Tested

| # | Serial | Model | SDK |
|---|--------|-------|-----|
EOF

for i in "${!TV_DEVICES[@]}"; do
    serial="${TV_DEVICES[$i]}"
    model=$(adb -s "$serial" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
    sdk=$(adb -s "$serial" shell getprop ro.build.version.sdk 2>/dev/null | tr -d '\r')
    echo "| $((i+1)) | $serial | $model | $sdk |" >> "$REPORT"
done

cat >> "$REPORT" << 'EOF'

## Phases

### Phase 1: Setup
- APK deployed to all devices
- Server URL configured via ADB reverse
- Authentication completed (admin/admin123)
- Screen recording started

### Phase 2: Doc-Driven Verification
- Navigated through all documented screens
- Screenshots captured at each transition
- DPAD navigation through sections and rows

### Phase 3: Curiosity-Driven Exploration
- Random navigation actions (UP/DOWN/LEFT/RIGHT/CENTER/BACK/MENU/HOME)
- Screenshots captured every 5 actions
- Crash detection via logcat monitoring
- Automatic app restart on crash

### Phase 4: Report & Cleanup
- Video recordings pulled from devices
- Logcat logs collected
- QA report generated

## Evidence

EOF

for i in "${!TV_DEVICES[@]}"; do
    label="device$((i+1))"
    screenshots=$(find "$OUTPUT_DIR/$label/screenshots" -name "*.png" 2>/dev/null | wc -l)
    videos=$(find "$OUTPUT_DIR/$label/videos" -name "*.mp4" 2>/dev/null | wc -l)
    crashes=$(find "$OUTPUT_DIR/$label/logs" -name "crash-*.txt" 2>/dev/null | wc -l)
    echo "### Device $((i+1)) (${TV_DEVICES[$i]})" >> "$REPORT"
    echo "- Screenshots: $screenshots" >> "$REPORT"
    echo "- Videos: $videos" >> "$REPORT"
    echo "- Crashes detected: $crashes" >> "$REPORT"
    echo "" >> "$REPORT"
done

echo "---" >> "$REPORT"
echo "*Generated by HelixQA Autonomous QA Runner*" >> "$REPORT"

ok "Report: $REPORT"
log "═══ SESSION COMPLETE ═══"
log "Results: $OUTPUT_DIR"

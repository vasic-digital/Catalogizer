#!/usr/bin/env bash
# HelixQA Autonomous QA — Android (Phone/Tablet)
# Runs full 4-phase autonomous QA with video recording on all connected Android phone devices.
#
# Usage:
#   ./scripts/run-helixqa-android.sh [--api-url URL] [--timeout MINUTES]
#
# Prerequisites:
#   - catalog-api running and accessible
#   - ADB connected Android phone/tablet devices (non-TV)
#
# Output: qa-results/helixqa-android-<timestamp>/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
OUTPUT_DIR="$PROJECT_ROOT/qa-results/helixqa-android-${TIMESTAMP}"
PKG="com.catalogizer.android"
FFMPEG="${HELIX_RECORDING_FFMPEG_PATH:-/home/milosvasic/bin/ffmpeg}"

# Defaults
API_URL=""
TIMEOUT_MINUTES=30

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${BLUE}[HelixQA-Android]${NC} $*"; }
ok()   { echo -e "${GREEN}[      OK       ]${NC} $*"; }
warn() { echo -e "${YELLOW}[     WARN     ]${NC} $*"; }
fail() { echo -e "${RED}[     FAIL     ]${NC} $*"; }

while [[ $# -gt 0 ]]; do
    case "$1" in
        --api-url)   API_URL="$2"; shift 2 ;;
        --timeout)   TIMEOUT_MINUTES="$2"; shift 2 ;;
        *)           shift ;;
    esac
done

if [[ -z "$API_URL" ]]; then
    HOST_IP=$(ip -4 addr show scope global | grep -oP '(?<=inet\s)[\d.]+' | head -1)
    API_URL="http://${HOST_IP:-localhost}:8080"
fi
log "API URL: $API_URL"

# ─── Discover Android Phone Devices (exclude TV) ────────────────────────
log "Discovering Android phone devices..."
PHONE_DEVICES=()
while IFS= read -r serial; do
    [[ -z "$serial" ]] && continue
    features=$(adb -s "$serial" shell pm list features 2>/dev/null | grep -cE "leanback|television" || true)
    chars=$(adb -s "$serial" shell getprop ro.build.characteristics 2>/dev/null | tr -d '\r')
    if [[ "$features" -eq 0 ]] && [[ "$chars" != *"tv"* ]]; then
        model=$(adb -s "$serial" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
        sdk=$(adb -s "$serial" shell getprop ro.build.version.sdk 2>/dev/null | tr -d '\r')
        PHONE_DEVICES+=("$serial")
        ok "Found phone: $serial ($model, SDK $sdk)"
    fi
done < <(adb devices 2>/dev/null | grep -w "device" | awk '{print $1}')

if [[ ${#PHONE_DEVICES[@]} -eq 0 ]]; then
    fail "No Android phone devices found"
    exit 1
fi
log "Found ${#PHONE_DEVICES[@]} Android phone device(s)"

for i in "${!PHONE_DEVICES[@]}"; do
    mkdir -p "$OUTPUT_DIR/phone$((i+1))"/{videos,screenshots,evidence,logs}
done

# ═══ PHASE 1: SETUP ═════════════════════════════════════════════════════
log "═══ PHASE 1: SETUP ═══"

if ! curl -sf "$API_URL/health" > /dev/null 2>&1; then
    fail "catalog-api not reachable at $API_URL"; exit 1
fi
ok "catalog-api healthy"

APK=$(find "$PROJECT_ROOT/catalogizer-android" -name "app-debug.apk" -path "*/build/outputs/*" 2>/dev/null | head -1)
if [[ -z "$APK" ]]; then
    warn "No debug APK found, building..."
    (cd "$PROJECT_ROOT/catalogizer-android" && ./gradlew assembleDebug --no-daemon 2>&1 | tail -3)
    APK=$(find "$PROJECT_ROOT/catalogizer-android" -name "app-debug.apk" -path "*/build/outputs/*" 2>/dev/null | head -1)
fi

for serial in "${PHONE_DEVICES[@]}"; do
    adb -s "$serial" install -r "$APK" 2>&1 | tail -1 &
    adb -s "$serial" reverse tcp:8080 tcp:8080 2>/dev/null || true
done
wait
ok "APK deployed, ADB reverse set"

# Launch and setup each device
for i in "${!PHONE_DEVICES[@]}"; do
    serial="${PHONE_DEVICES[$i]}"
    label="phone$((i+1))"
    dir="$OUTPUT_DIR/$label"

    adb -s "$serial" shell am force-stop "$PKG" 2>/dev/null
    adb -s "$serial" shell am start -n "$PKG/.ui.MainActivity" 2>/dev/null
    sleep 4
    adb -s "$serial" exec-out screencap -p > "$dir/screenshots/001-launch.png" 2>/dev/null
done
ok "All phones launched"

# Start recording (SDK 35 uses screenshot-to-video approach)
RECORD_PIDS=()
for i in "${!PHONE_DEVICES[@]}"; do
    serial="${PHONE_DEVICES[$i]}"
    sdk=$(adb -s "$serial" shell getprop ro.build.version.sdk 2>/dev/null | tr -d '\r')
    label="phone$((i+1))"
    if [[ "$sdk" -ge 35 ]]; then
        # SDK 35: screenrecord may fail, use rapid screenshots
        (
            dir="$OUTPUT_DIR/$label/videos/frames"
            mkdir -p "$dir"
            frame=0
            end=$(($(date +%s) + TIMEOUT_MINUTES * 60))
            while [[ $(date +%s) -lt $end ]]; do
                adb -s "$serial" exec-out screencap -p > "$dir/frame_$(printf '%05d' $frame).png" 2>/dev/null
                ((frame++))
                sleep 2
            done
        ) &
        RECORD_PIDS+=($!)
    else
        adb -s "$serial" shell "screenrecord --bit-rate 4000000 --time-limit 180 /sdcard/qa_session_1.mp4" &
        RECORD_PIDS+=($!)
    fi
done
ok "Recording started"

# ═══ PHASE 2: DOC-DRIVEN VERIFICATION ═══════════════════════════════════
log "═══ PHASE 2: DOC-DRIVEN VERIFICATION ═══"

for i in "${!PHONE_DEVICES[@]}"; do
    serial="${PHONE_DEVICES[$i]}"
    label="phone$((i+1))"
    dir="$OUTPUT_DIR/$label"
    step=10

    # Navigate through app screens via swipes and taps
    for action in "swipe 500 1500 500 500" "tap 540 960" "keyevent KEYCODE_BACK" \
                  "swipe 500 500 500 1500" "tap 540 300" "keyevent KEYCODE_BACK" \
                  "swipe 900 960 100 960" "keyevent KEYCODE_BACK" \
                  "swipe 100 960 900 960" "keyevent KEYCODE_BACK"; do
        adb -s "$serial" shell input $action 2>/dev/null
        sleep 1.5
        adb -s "$serial" exec-out screencap -p > "$dir/screenshots/0${step}-verify.png" 2>/dev/null
        ((step++))
    done &
done
wait
ok "Phase 2 complete"

# ═══ PHASE 3: CURIOSITY-DRIVEN EXPLORATION ══════════════════════════════
log "═══ PHASE 3: CURIOSITY-DRIVEN EXPLORATION ═══"

for i in "${!PHONE_DEVICES[@]}"; do
    serial="${PHONE_DEVICES[$i]}"
    label="phone$((i+1))"
    dir="$OUTPUT_DIR/$label"
    (
        step=50
        end=$(($(date +%s) + TIMEOUT_MINUTES * 60 / 3))
        while [[ $(date +%s) -lt $end ]]; do
            case $((RANDOM % 6)) in
                0) adb -s "$serial" shell input swipe 500 1500 500 500 300 ;;
                1) adb -s "$serial" shell input swipe 500 500 500 1500 300 ;;
                2) adb -s "$serial" shell input tap $((RANDOM % 1080)) $((RANDOM % 1920)) ;;
                3) adb -s "$serial" shell input keyevent KEYCODE_BACK ;;
                4) adb -s "$serial" shell input swipe 900 960 100 960 300 ;;
                5) adb -s "$serial" shell input swipe 100 960 900 960 300 ;;
            esac
            sleep 1
            if [[ $((step % 5)) -eq 0 ]]; then
                adb -s "$serial" exec-out screencap -p > "$dir/screenshots/${step}-explore.png" 2>/dev/null
            fi
            # Only detect crashes in OUR package — not system processes
            crash=$(adb -s "$serial" logcat -d -t 3 "*:E" 2>/dev/null | grep -E "FATAL|AndroidRuntime" | grep -c "$PKG" || echo "0")
            if [[ "$crash" -gt 0 ]]; then
                adb -s "$serial" exec-out screencap -p > "$dir/screenshots/${step}-CRASH.png" 2>/dev/null
                adb -s "$serial" logcat -d -t 30 "*:E" > "$dir/logs/crash-step${step}.txt" 2>/dev/null
                adb -s "$serial" shell am force-stop "$PKG" 2>/dev/null; sleep 1
                adb -s "$serial" shell am start -n "$PKG/.ui.MainActivity" 2>/dev/null; sleep 3
            fi
            ((step++))
        done
    ) &
done
wait
ok "Phase 3 complete"

# ═══ PHASE 4: REPORT & CLEANUP ══════════════════════════════════════════
log "═══ PHASE 4: REPORT & CLEANUP ═══"

for pid in "${RECORD_PIDS[@]}"; do kill "$pid" 2>/dev/null || true; done
sleep 2

for i in "${!PHONE_DEVICES[@]}"; do
    serial="${PHONE_DEVICES[$i]}"
    label="phone$((i+1))"
    adb -s "$serial" pull "/sdcard/qa_session_1.mp4" "$OUTPUT_DIR/$label/videos/" 2>/dev/null || true
    adb -s "$serial" logcat -d > "$OUTPUT_DIR/$label/logs/logcat-full.txt" 2>/dev/null
done

REPORT="$OUTPUT_DIR/qa-report.md"
cat > "$REPORT" << EOF
# HelixQA Android Phone Autonomous QA Report

**Generated:** $(date -Iseconds)
**Devices:** ${#PHONE_DEVICES[@]}
**Duration:** ${TIMEOUT_MINUTES} minutes

## Devices
EOF

for i in "${!PHONE_DEVICES[@]}"; do
    serial="${PHONE_DEVICES[$i]}"
    model=$(adb -s "$serial" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
    label="phone$((i+1))"
    screenshots=$(find "$OUTPUT_DIR/$label/screenshots" -name "*.png" 2>/dev/null | wc -l)
    crashes=$(find "$OUTPUT_DIR/$label/logs" -name "crash-*.txt" 2>/dev/null | wc -l)
    echo "- **$label** ($serial): $model — $screenshots screenshots, $crashes crashes" >> "$REPORT"
done

echo -e "\n---\n*Generated by HelixQA*" >> "$REPORT"

ok "Report: $REPORT"
log "═══ SESSION COMPLETE ═══"
log "Results: $OUTPUT_DIR"

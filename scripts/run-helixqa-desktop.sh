#!/usr/bin/env bash
# HelixQA Autonomous QA — Desktop (Tauri)
# Runs full 4-phase autonomous QA with video recording against the desktop app.
#
# Usage:
#   ./scripts/run-helixqa-desktop.sh [--api-url URL] [--timeout MINUTES]
#
# Prerequisites:
#   - catalog-api running on port 8080
#   - catalogizer-desktop built (AppImage or binary)
#   - X11 display available
#   - ffmpeg, xdotool installed
#
# Output: qa-results/helixqa-desktop-<timestamp>/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
OUTPUT_DIR="$PROJECT_ROOT/qa-results/helixqa-desktop-${TIMESTAMP}"
FFMPEG="${HELIX_RECORDING_FFMPEG_PATH:-/home/milosvasic/bin/ffmpeg}"
DISPLAY="${DISPLAY:-:0}"

API_URL="http://localhost:8080"
TIMEOUT_MINUTES=20

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${BLUE}[HelixQA-Desktop]${NC} $*"; }
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

mkdir -p "$OUTPUT_DIR"/{videos,screenshots,evidence,logs}

# ═══ PHASE 1: SETUP ═════════════════════════════════════════════════════
log "═══ PHASE 1: SETUP ═══"

if ! curl -sf "$API_URL/health" > /dev/null 2>&1; then
    fail "catalog-api not running at $API_URL"; exit 1
fi
ok "catalog-api healthy"

# Find desktop binary
DESKTOP_BIN=""
for candidate in \
    "$PROJECT_ROOT/catalogizer-desktop/src-tauri/target/release/catalogizer-desktop" \
    "$PROJECT_ROOT/catalogizer-desktop/src-tauri/target/debug/catalogizer-desktop" \
    "$PROJECT_ROOT/releases/linux/catalogizer-desktop/"*.AppImage; do
    if [[ -x "$candidate" ]]; then
        DESKTOP_BIN="$candidate"
        break
    fi
done

if [[ -z "$DESKTOP_BIN" ]]; then
    fail "No desktop binary found. Build with: cd catalogizer-desktop && npm run tauri:build"
    exit 1
fi
ok "Desktop binary: $DESKTOP_BIN"

# Start screen recording
if command -v "$FFMPEG" &>/dev/null; then
    "$FFMPEG" -y -f x11grab -framerate 10 -video_size 1920x1080 -i "$DISPLAY" \
        -c:v libx264 -preset ultrafast -crf 28 \
        "$OUTPUT_DIR/videos/desktop-session.mp4" 2>/dev/null &
    FFMPEG_PID=$!
    ok "Screen recording started"
else
    FFMPEG_PID=""
    warn "ffmpeg not found, skipping recording"
fi

# Launch desktop app
APPIMAGE_EXTRACT_AND_RUN=1 "$DESKTOP_BIN" &
APP_PID=$!
sleep 5
ok "Desktop app launched (PID=$APP_PID)"

# Initial screenshot
import -window root "$OUTPUT_DIR/screenshots/001-launch.png" 2>/dev/null || \
    scrot "$OUTPUT_DIR/screenshots/001-launch.png" 2>/dev/null || \
    warn "Could not capture screenshot (install scrot or imagemagick)"

# ═══ PHASE 2: DOC-DRIVEN VERIFICATION ═══════════════════════════════════
log "═══ PHASE 2: DOC-DRIVEN VERIFICATION ═══"

STEP=10
if command -v xdotool &>/dev/null; then
    # Wait for window to appear
    sleep 3
    WID=$(xdotool search --name "Catalogizer" 2>/dev/null | head -1 || echo "")

    if [[ -n "$WID" ]]; then
        xdotool windowactivate "$WID" 2>/dev/null || true
        sleep 1

        # Navigate through tabs/sections using keyboard
        KEYS=("Tab" "Tab" "Return" "Escape" "Tab" "Tab" "Return" "Escape" \
              "Tab" "Tab" "Return" "Escape" "ctrl+f" "Escape")

        for key in "${KEYS[@]}"; do
            xdotool key "$key" 2>/dev/null || true
            sleep 1.5
            scrot "$OUTPUT_DIR/screenshots/0${STEP}-nav.png" 2>/dev/null || true
            ((STEP++))
        done
        ok "Keyboard navigation complete"
    else
        warn "Could not find Catalogizer window"
    fi
else
    warn "xdotool not installed, skipping keyboard navigation"
fi
ok "Phase 2 complete"

# ═══ PHASE 3: CURIOSITY-DRIVEN EXPLORATION ══════════════════════════════
log "═══ PHASE 3: CURIOSITY-DRIVEN EXPLORATION ═══"

EXPLORE_END=$(($(date +%s) + TIMEOUT_MINUTES * 60 / 3))
while [[ $(date +%s) -lt $EXPLORE_END ]]; do
    if command -v xdotool &>/dev/null; then
        case $((RANDOM % 5)) in
            0) xdotool key Tab 2>/dev/null ;;
            1) xdotool key Return 2>/dev/null ;;
            2) xdotool key Escape 2>/dev/null ;;
            3) xdotool key "ctrl+f" 2>/dev/null ;;
            4) xdotool key "alt+F4" 2>/dev/null; sleep 1
               APPIMAGE_EXTRACT_AND_RUN=1 "$DESKTOP_BIN" &
               APP_PID=$!; sleep 3 ;;
        esac
    fi
    sleep 2

    if [[ $((STEP % 5)) -eq 0 ]]; then
        scrot "$OUTPUT_DIR/screenshots/${STEP}-explore.png" 2>/dev/null || true
    fi
    ((STEP++))

    # Check if app is still running
    if ! kill -0 "$APP_PID" 2>/dev/null; then
        warn "App crashed at step $STEP, restarting..."
        echo "Crash at step $STEP, $(date -Iseconds)" >> "$OUTPUT_DIR/logs/crashes.txt"
        APPIMAGE_EXTRACT_AND_RUN=1 "$DESKTOP_BIN" &
        APP_PID=$!; sleep 3
    fi
done
ok "Phase 3 complete"

# ═══ PHASE 4: REPORT & CLEANUP ══════════════════════════════════════════
log "═══ PHASE 4: REPORT & CLEANUP ═══"

kill "$APP_PID" 2>/dev/null || true
if [[ -n "${FFMPEG_PID:-}" ]]; then
    kill "$FFMPEG_PID" 2>/dev/null || true
    wait "$FFMPEG_PID" 2>/dev/null || true
fi

SCREENSHOTS=$(find "$OUTPUT_DIR/screenshots" -name "*.png" 2>/dev/null | wc -l)
VIDEOS=$(find "$OUTPUT_DIR/videos" -name "*.mp4" 2>/dev/null | wc -l)
CRASHES=$(grep -c "Crash" "$OUTPUT_DIR/logs/crashes.txt" 2>/dev/null || echo "0")

cat > "$OUTPUT_DIR/qa-report.md" << EOF
# HelixQA Desktop Autonomous QA Report

**Generated:** $(date -Iseconds)
**API URL:** $API_URL
**Binary:** $DESKTOP_BIN
**Duration:** ${TIMEOUT_MINUTES} minutes

## Summary
- **Screenshots:** $SCREENSHOTS
- **Videos:** $VIDEOS
- **Crashes detected:** $CRASHES

---
*Generated by HelixQA*
EOF

ok "Report: $OUTPUT_DIR/qa-report.md"
log "═══ SESSION COMPLETE ═══"
log "Results: $OUTPUT_DIR"

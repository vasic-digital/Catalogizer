#!/usr/bin/env bash
# HelixQA Full QA — ALL Platforms (Sequential)
# Runs full autonomous QA sequentially: API → Web → Android TV → Android → Desktop
# Each platform completes before the next starts (resource-safe).
#
# Usage:
#   ./scripts/run-helixqa-all.sh [--timeout MINUTES] [--skip PLATFORM]
#
# Prerequisites:
#   - catalog-api running on port 8080
#   - catalog-web running on port 3000
#   - ADB devices connected (for Android/Android TV)
#   - HelixQA binary built

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HELIXQA_BIN="$PROJECT_ROOT/submodules/helix_qa/helixqa"
ENV_FILE="$PROJECT_ROOT/submodules/helix_qa/.env"
OUTPUT_DIR="$PROJECT_ROOT/qa-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SESSION_DATE=$(date +%Y-%m-%d)
LOG_DIR="$PROJECT_ROOT/docs/reports/qa-sessions/qa-session-$SESSION_DATE/logs"
REPORT_DIR="$PROJECT_ROOT/docs/reports/qa-sessions/qa-session-$SESSION_DATE"

TIMEOUT="${TIMEOUT:-60m}"
CURIOSITY_TIMEOUT="${CURIOSITY_TIMEOUT:-20m}"
SKIP_PLATFORMS="${SKIP_PLATFORMS:-}"

mkdir -p "$LOG_DIR" "$REPORT_DIR"

echo "================================================================"
echo "  HelixQA FULL QA — All Platforms (Sequential)"
echo "================================================================"
echo "  Date:              $(date)"
echo "  Timeout/platform:  $TIMEOUT"
echo "  Curiosity timeout: $CURIOSITY_TIMEOUT"
echo "  Output:            $OUTPUT_DIR"
echo "  Reports:           $REPORT_DIR"
echo "  Skip:              ${SKIP_PLATFORMS:-none}"
echo "================================================================"
echo ""

TOTAL_PASS=0
TOTAL_FAIL=0
TOTAL_ISSUES=0
RESULTS=""

run_platform() {
    local PLATFORM="$1"
    local EXTRA_ARGS="${2:-}"

    if echo "$SKIP_PLATFORMS" | grep -qi "$PLATFORM"; then
        echo "[$PLATFORM] SKIPPED (in skip list)"
        RESULTS="$RESULTS\n  $PLATFORM: SKIPPED"
        return 0
    fi

    echo "=============================================="
    echo "  Platform: $PLATFORM"
    echo "  Start:    $(date)"
    echo "=============================================="

    local LOG_FILE="$LOG_DIR/helixqa-autonomous-$PLATFORM-$TIMESTAMP.log"
    local START_TIME=$(date +%s)

    if "$HELIXQA_BIN" autonomous \
        --project "$PROJECT_ROOT" \
        --platforms "$PLATFORM" \
        --env "$ENV_FILE" \
        --timeout "$TIMEOUT" \
        --curiosity-timeout "$CURIOSITY_TIMEOUT" \
        --coverage-target 0.9 \
        --output "$OUTPUT_DIR" \
        --verbose \
        $EXTRA_ARGS \
        2>&1 | tee "$LOG_FILE"; then

        local END_TIME=$(date +%s)
        local DURATION=$(( END_TIME - START_TIME ))
        echo ""
        echo "[$PLATFORM] COMPLETED in ${DURATION}s"
        RESULTS="$RESULTS\n  $PLATFORM: COMPLETED (${DURATION}s)"
        TOTAL_PASS=$((TOTAL_PASS + 1))
    else
        local END_TIME=$(date +%s)
        local DURATION=$(( END_TIME - START_TIME ))
        echo ""
        echo "[$PLATFORM] FAILED after ${DURATION}s"
        RESULTS="$RESULTS\n  $PLATFORM: FAILED (${DURATION}s)"
        TOTAL_FAIL=$((TOTAL_FAIL + 1))
    fi

    # Copy evidence to report directory
    local LATEST_SESSION=$(ls -td "$OUTPUT_DIR"/session-* 2>/dev/null | head -1)
    if [ -n "$LATEST_SESSION" ] && [ -d "$LATEST_SESSION/screenshots" ]; then
        mkdir -p "$REPORT_DIR/screenshots/$PLATFORM"
        cp -r "$LATEST_SESSION/screenshots/"* "$REPORT_DIR/screenshots/$PLATFORM/" 2>/dev/null || true
    fi
    if [ -n "$LATEST_SESSION" ] && [ -d "$LATEST_SESSION/videos" ]; then
        mkdir -p "$REPORT_DIR/videos/$PLATFORM"
        cp -r "$LATEST_SESSION/videos/"* "$REPORT_DIR/videos/$PLATFORM/" 2>/dev/null || true
    fi

    echo ""
}

# Sequential execution — one platform at a time
echo ">>> Phase 1/5: API"
run_platform "api"

echo ">>> Phase 2/5: Web"
run_platform "web"

echo ">>> Phase 3/5: Android TV"
run_platform "androidtv"

echo ">>> Phase 4/5: Android"
run_platform "android"

echo ">>> Phase 5/5: Desktop"
run_platform "desktop"

# Summary
echo ""
echo "================================================================"
echo "  HelixQA FULL QA — Summary"
echo "================================================================"
echo "  Completed: $TOTAL_PASS"
echo "  Failed:    $TOTAL_FAIL"
echo -e "  Results:$RESULTS"
echo "  Finished:  $(date)"
echo "================================================================"

# Write summary to report
cat > "$REPORT_DIR/helixqa-summary-$TIMESTAMP.md" << EOF
# HelixQA Full QA Summary — $SESSION_DATE

| Metric | Value |
|--------|-------|
| Date | $(date) |
| Platforms completed | $TOTAL_PASS |
| Platforms failed | $TOTAL_FAIL |
| Timeout/platform | $TIMEOUT |

## Platform Results
$(echo -e "$RESULTS")

## Session Logs
$(ls "$LOG_DIR"/helixqa-autonomous-*-$TIMESTAMP.log 2>/dev/null | while read f; do echo "- $(basename "$f")"; done)
EOF

echo ""
echo "Full report: $REPORT_DIR/helixqa-summary-$TIMESTAMP.md"

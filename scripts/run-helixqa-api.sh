#!/usr/bin/env bash
# HelixQA Autonomous QA — API (catalog-api)
# Runs full autonomous QA against the REST API endpoints only (no UI).
#
# Usage:
#   ./scripts/run-helixqa-api.sh [--api-url URL] [--timeout MINUTES]
#
# Prerequisites:
#   - catalog-api running and accessible (default: http://localhost:8080)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HELIXQA_BIN="$PROJECT_ROOT/submodules/helix_qa/helixqa"
ENV_FILE="$PROJECT_ROOT/submodules/helix_qa/.env"
OUTPUT_DIR="$PROJECT_ROOT/qa-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_DIR="$PROJECT_ROOT/docs/reports/qa-sessions/qa-session-$(date +%Y-%m-%d)/logs"

API_URL="${1:-http://localhost:8080}"
TIMEOUT="${2:-30m}"

mkdir -p "$LOG_DIR"

echo "=============================================="
echo "  HelixQA API QA Session"
echo "=============================================="
echo "  API URL:    $API_URL"
echo "  Timeout:    $TIMEOUT"
echo "  Output:     $OUTPUT_DIR"
echo "  Log:        $LOG_DIR/helixqa-api-$TIMESTAMP.log"
echo "=============================================="

# Verify API is reachable
if ! curl -s --max-time 5 "$API_URL/health" > /dev/null 2>&1; then
    echo "ERROR: API not reachable at $API_URL/health"
    exit 1
fi
echo "API health check: OK"

# Run HelixQA bank tests for API platform
echo ""
echo "--- Running API bank tests ---"
"$HELIXQA_BIN" run \
    --banks "$PROJECT_ROOT/submodules/helix_qa/banks/full-qa-api.yaml" \
    --platform api \
    2>&1 | tee "$LOG_DIR/helixqa-api-$TIMESTAMP.log"

echo ""
echo "API QA session complete at $(date)"

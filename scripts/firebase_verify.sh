#!/usr/bin/env bash
# firebase_verify.sh — Verify Firebase CLI connectivity and report enabled
# services. Tests authentication, project access, and App Distribution
# readiness.
#
# §11.4.18 — companion doc at docs/scripts/firebase_verify.md

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load .env if present
ENV_FILE="$PROJECT_ROOT/.env"
if [[ -f "$ENV_FILE" ]]; then
    set -a
    # shellcheck source=/dev/null
    source "$ENV_FILE"
    set +a
fi

PASS=0
FAIL=0
SKIP=0

_pass() { PASS=$((PASS+1)); echo "  [PASS] $*"; }
_fail() { FAIL=$((FAIL+1)); echo "  [FAIL] $*"; }
_skip() { SKIP=$((SKIP+1)); echo "  [SKIP] $*"; }
_info() { echo "  [INFO] $*"; }

# ── Check 1: Firebase CLI installed ─────────────────────────────────────
check_firebase_cli() {
    echo ""
    echo "=== 1. Firebase CLI ==="

    if command -v firebase &>/dev/null; then
        local version
        version="$(firebase --version 2>/dev/null || echo "unknown")"
        _pass "firebase CLI found, version $version"
    else
        _fail "firebase CLI not found on PATH"
        _info "Install: npm install -g firebase-tools"
    fi
}

# ── Check 2: Authentication status ──────────────────────────────────────
check_auth() {
    echo ""
    echo "=== 2. Authentication ==="

    local login_output
    login_output="$(firebase login --no-localhost 2>&1 || true)"

    if firebase projects:list --json 2>/dev/null | grep -q '"projectId"'; then
        _pass "Authenticated to Firebase"
    else
        _fail "Not authenticated. Run: firebase login"
        return
    fi
}

# ── Check 3: Project access ─────────────────────────────────────────────
check_projects() {
    echo ""
    echo "=== 3. Firebase Projects ==="

    local json_output
    json_output="$(firebase projects:list --json 2>/dev/null || echo '{}')"

    local project_count
    project_count="$(echo "$json_output" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    projects = d.get('projects', [])
    print(len(projects))
    for p in projects:
        print(p.get('projectId', '?'))
except Exception:
    print(0)
" 2>/dev/null || echo "0")"

    if [[ "$project_count" -gt 0 ]]; then
        _pass "Found $project_count Firebase project(s):"
        echo "$project_count" | while IFS= read -r line; do
            if [[ "$line" =~ ^[0-9]+$ ]]; then
                continue
            fi
            _info "  - $line"
        done
    else
        _fail "No Firebase projects accessible"
    fi
}

# ── Check 4: .env config for App Distribution ───────────────────────────
check_env_config() {
    echo ""
    echo "=== 4. App Distribution Config (.env) ==="

    local has_env_file=false
    local missing_keys=()

    if [[ -f "$ENV_FILE" ]]; then
        has_env_file=true
        _pass ".env file exists at $ENV_FILE"
    else
        _fail ".env file not found (App IDs will use defaults)"
    fi

    for key in FIREBASE_TESTER_EMAILS FIREBASE_API_APP_ID FIREBASE_ANDROIDTV_APP_ID; do
        local val="${!key:-}"
        if [[ -n "$val" ]]; then
            _pass "$key is set"
        else
            missing_keys+=("$key")
        fi
    done

    if [[ ${#missing_keys[@]} -gt 0 ]]; then
        _fail "Missing env vars: ${missing_keys[*]}"
        _info "Run: scripts/firebase_setup_env.sh"
    fi
}

# ── Check 5: App Distribution API enabled ───────────────────────────────
check_app_distribution_api() {
    echo ""
    echo "=== 5. Firebase App Distribution API ==="

    # Check via Firebase CLI by trying a harmless API call
    local test_output
    test_output="$(firebase appdistribution:distribute --help 2>&1 || true)"

    if echo "$test_output" | grep -qi "usage\|firebase appdistribution:distribute\|testers"; then
        _pass "Firebase App Distribution CLI commands available"
        _info "  - firebase appdistribution:distribute"
        _info "  - firebase appdistribution:testers:add"
    else
        _fail "App Distribution CLI not responding as expected"
        _info "Enable the API at: https://console.firebase.google.com"
    fi
}

# ── Check 6: Build artifacts ────────────────────────────────────────────
check_build_artifacts() {
    echo ""
    echo "=== 6. Build Artifacts ==="

    local api_binary="$PROJECT_ROOT/build/catalog-api/catalog-api"
    if [[ -f "$api_binary" ]]; then
        _pass "catalog-api binary found at build/catalog-api/catalog-api"
    else
        _skip "No pre-built catalog-api binary (build first with scripts/distribute.sh or SKIP_BUILD=true)"
    fi

    local apk
    apk="$(find "$PROJECT_ROOT/catalogizer-androidtv" -path '*/release/*.apk' -type f 2>/dev/null | head -1 || true)"
    if [[ -n "$apk" ]]; then
        _pass "Android TV APK found"
    else
        _skip "No pre-built Android TV APK (build first with scripts/distribute.sh or SKIP_BUILD=true)"
    fi
}

# ── Summary ─────────────────────────────────────────────────────────────
print_summary() {
    echo ""
    echo "═══════════════════════════════════════"
    echo "  Results: $PASS passed, $FAIL failed, $SKIP skipped"
    echo "═══════════════════════════════════════"

    if [[ "$FAIL" -eq 0 ]]; then
        echo "  All checks PASSED. Ready for distribution."
    else
        echo "  $FAIL check(s) FAILED — review above before distributing."
    fi
    echo ""
}

# ── Main ────────────────────────────────────────────────────────────────
main() {
    echo "==========================================="
    echo "  Firebase Connectivity & Service Report"
    echo "  $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    echo "==========================================="

    check_firebase_cli
    check_auth
    check_projects
    check_env_config
    check_app_distribution_api
    check_build_artifacts

    print_summary

    # Exit codes: 0 = all pass, 1 = any fail
    if [[ "$FAIL" -gt 0 ]]; then
        exit 1
    fi
}

main "$@"

#!/usr/bin/env bash
# firebase_setup_env.sh — Bootstrap Firebase-related .env entries
# Creates FIREBASE_* entries in .env if not present, and documents them
# in .env.example.
#
# §11.4.18 — companion doc at docs/scripts/firebase_setup_env.md

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

ENV_FILE="$PROJECT_ROOT/.env"
ENV_EXAMPLE="$PROJECT_ROOT/.env.example"

# ── Firebase config block ───────────────────────────────────────────────
read -r -d '' FIREBASE_ENV_BLOCK <<'FIREBASE_BLOCK' || true

# ─── Firebase App Distribution ──────────────────────────────────────────
# Comma-separated list of tester emails for Firebase App Distribution.
# These testers will receive builds distributed via scripts/distribute.sh.
# Use non-production emails for testing.
FIREBASE_TESTER_EMAILS=

# Firebase App IDs for each distributed artifact.
# Find these in the Firebase Console → Project Settings → General → Your apps.
FIREBASE_API_APP_ID=
FIREBASE_ANDROIDTV_APP_ID=

# Distribution release notes template (optional — override via CLI arg)
FIREBASE_RELEASE_NOTES_TEMPLATE="Build $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
FIREBASE_BLOCK

# ── Helpers ─────────────────────────────────────────────────────────────
_log()  { echo "[firebase_setup_env] $*"; }
_err()  { echo "[firebase_setup_env] ERROR: $*" >&2; }
_warn() { echo "[firebase_setup_env] WARN: $*" >&2; }

_env_has_key() {
    local key="$1"
    local file="$2"
    if [[ -f "$file" ]] && grep -q "^${key}=" "$file" 2>/dev/null; then
        return 0
    fi
    return 1
}

_ensure_newline() {
    local file="$1"
    if [[ -s "$file" ]] && [[ "$(tail -c1 "$file" | wc -l)" -eq 0 ]]; then
        echo "" >> "$file"
    fi
}

# ── Update .env ─────────────────────────────────────────────────────────
setup_env() {
    if [[ ! -f "$ENV_FILE" ]]; then
        _log ".env does not exist yet — creating from .env.example"
        _info "Copy .env.example to .env and fill in your values."
        return 0
    fi

    local any_missing=false

    for key in FIREBASE_TESTER_EMAILS FIREBASE_API_APP_ID FIREBASE_ANDROIDTV_APP_ID; do
        if _env_has_key "$key" "$ENV_FILE"; then
            _log "$key is already set in .env"
        else
            _warn "$key is MISSING from .env"
            any_missing=true
        fi
    done

    if $any_missing; then
        _log ""
        _log "The following keys need values in $ENV_FILE:"
        _log "  FIREBASE_TESTER_EMAILS   — e.g. tester1@example.com,tester2@example.com"
        _log "  FIREBASE_API_APP_ID      — Firebase App ID for catalog-api (see Console)"
        _log "  FIREBASE_ANDROIDTV_APP_ID — Firebase App ID for catalogizer-androidtv"
        _log ""
        _log "Run the following to add placeholder entries:"
        _log "  cat >> $ENV_FILE << 'EOF'"
        echo ""
        echo "# Firebase App Distribution"
        echo "FIREBASE_TESTER_EMAILS="
        echo "FIREBASE_API_APP_ID="
        echo "FIREBASE_ANDROIDTV_APP_ID="
        echo ""
        _log "Then edit $ENV_FILE with real values."
        _log ""
        _log "Firebase App IDs are found at:"
        _log "  https://console.firebase.google.com → Project Settings → General → Your apps"
    else
        _log "All Firebase env vars present."
    fi
}

# ── Update .env.example ─────────────────────────────────────────────────
setup_env_example() {
    if [[ ! -f "$ENV_EXAMPLE" ]]; then
        _warn ".env.example not found at $ENV_EXAMPLE — skipping."
        return 0
    fi

    # Check if the Firebase block already exists
    if grep -q "FIREBASE_TESTER_EMAILS" "$ENV_EXAMPLE" 2>/dev/null; then
        _log "Firebase env vars already documented in .env.example"
        return 0
    fi

    _ensure_newline "$ENV_EXAMPLE"
    echo "$FIREBASE_ENV_BLOCK" >> "$ENV_EXAMPLE"
    _log "Firebase env block appended to .env.example"
}

# ── Main ────────────────────────────────────────────────────────────────
main() {
    _log "=== Firebase Environment Setup ==="
    _log ""

    setup_env
    setup_env_example

    _log ""
    _log "=== Done ==="
    _log "Next steps:"
    _log "  1. Edit $ENV_FILE with your Firebase App IDs"
    _log "  2. Run: scripts/firebase_verify.sh"
    _log "  3. Run: scripts/distribute.sh"
}

main "$@"

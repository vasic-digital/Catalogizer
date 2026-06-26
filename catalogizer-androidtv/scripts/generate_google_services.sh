#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# §11.4.10 — Generate google-services.json from .env vars at build time.
#
# Purpose:
#   Generates app/google-services.json from project environment variables
#   defined in the project's .env file, so the real credentials are never
#   committed to version control (per §11.4.30).
#
# Usage:
#   bash scripts/generate_google_services.sh
#
# Inputs (read from .env — all must be set):
#   FIREBASE_PROJECT_NUMBER
#   FIREBASE_PROJECT_ID
#   FIREBASE_STORAGE_BUCKET
#   FIREBASE_MOBILE_SDK_APP_ID
#   FIREBASE_API_KEY
#
# Outputs:
#   app/google-services.json — overwritten atomically via temp file + rename.
#
# Side-effects:
#   Creates app/google-services.json. Exits 1 if any required var is empty.
#
# Dependencies:
#   envsubst (gettext suite) or a POSIX-compatible sed chain.
#
# Cross-references:
#   .gitignore (§11.4.30 — google-services.json is git-ignored)
#   app/google-services.json.example (template without real values)
# ──────────────────────────────────────────────────────────────────────────────

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load .env if present
ENV_FILE="$PROJECT_ROOT/.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

# Required variables (all must be non-empty)
REQUIRED_VARS=(
  FIREBASE_PROJECT_NUMBER
  FIREBASE_PROJECT_ID
  FIREBASE_STORAGE_BUCKET
  FIREBASE_MOBILE_SDK_APP_ID
  FIREBASE_API_KEY
)

MISSING=false
for var in "${REQUIRED_VARS[@]}"; do
  if [ -z "${!var:-}" ]; then
    echo "ERROR: \$FIREBASE_* variables not set in .env; required=$var"
    MISSING=true
  fi
done

if [ "$MISSING" = true ]; then
  echo ""
  echo "FATAL: Cannot generate google-services.json without all $FIREBASE_* variables."
  echo "  Add them to $PROJECT_ROOT/.env:"
  echo ""
  echo "  FIREBASE_PROJECT_NUMBER=123456789"
  echo "  FIREBASE_PROJECT_ID=my-project-id"
  echo "  FIREBASE_STORAGE_BUCKET=my-project-id.appspot.com"
  echo "  FIREBASE_MOBILE_SDK_APP_ID=1:123456789:android:abcdef123456"
  echo "  FIREBASE_API_KEY=AIzaSy..."
  echo ""
  echo "  Obtain these from the Firebase Console: Project Settings > General > Your apps."
  exit 1
fi

TEMPLATE="$PROJECT_ROOT/app/google-services.json.example"
if [ ! -f "$TEMPLATE" ]; then
  echo "FATAL: Template not found at $TEMPLATE"
  exit 1
fi

# Use envsubst if available (gettext suite), otherwise chain sed
if command -v envsubst &>/dev/null; then
  export FIREBASE_PROJECT_NUMBER FIREBASE_PROJECT_ID FIREBASE_STORAGE_BUCKET FIREBASE_MOBILE_SDK_APP_ID FIREBASE_API_KEY
  envsubst < "$TEMPLATE" > "$PROJECT_ROOT/app/google-services.json.tmp"
else
  # Fallback: sed replacements one by one
  cp "$TEMPLATE" "$PROJECT_ROOT/app/google-services.json.tmp"
  sed -i '' "s|PASTE_YOUR_PROJECT_NUMBER|$FIREBASE_PROJECT_NUMBER|g" "$PROJECT_ROOT/app/google-services.json.tmp"
  sed -i '' "s|PASTE_YOUR_PROJECT_ID|$FIREBASE_PROJECT_ID|g" "$PROJECT_ROOT/app/google-services.json.tmp"
  sed -i '' "s|PASTE_YOUR_STORAGE_BUCKET|$FIREBASE_STORAGE_BUCKET|g" "$PROJECT_ROOT/app/google-services.json.tmp"
  sed -i '' "s|PASTE_YOUR_MOBILE_SDK_APP_ID|$FIREBASE_MOBILE_SDK_APP_ID|g" "$PROJECT_ROOT/app/google-services.json.tmp"
  sed -i '' "s|PASTE_YOUR_API_KEY|$FIREBASE_API_KEY|g" "$PROJECT_ROOT/app/google-services.json.tmp"
fi

# Atomic rename to avoid partial read
mv "$PROJECT_ROOT/app/google-services.json.tmp" "$PROJECT_ROOT/app/google-services.json"
echo "OK: google-services.json generated at app/google-services.json"

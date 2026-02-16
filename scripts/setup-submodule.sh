#!/bin/bash

# setup-submodule.sh - Automates submodule setup for Catalogizer project
#
# Usage:
#   ./scripts/setup-submodule.sh <ModuleName> [--create-repos] [--go|--ts|--kotlin]
#
# Examples:
#   ./scripts/setup-submodule.sh Auth                     # Add existing repo as submodule
#   ./scripts/setup-submodule.sh Filesystem --create-repos --go  # Create new repos + add as submodule
#   ./scripts/setup-submodule.sh WebSocket-Client-TS --create-repos --ts

set -euo pipefail

MODULE_NAME="${1:-}"
CREATE_REPOS=false
MODULE_TYPE="go"

if [ -z "$MODULE_NAME" ]; then
  echo "ERROR: Module name required"
  echo "Usage: $0 <ModuleName> [--create-repos] [--go|--ts|--kotlin]"
  exit 1
fi

shift
while [ $# -gt 0 ]; do
  case "$1" in
    --create-repos) CREATE_REPOS=true ;;
    --go) MODULE_TYPE="go" ;;
    --ts) MODULE_TYPE="ts" ;;
    --kotlin) MODULE_TYPE="kotlin" ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
  shift
done

# Derive names
MODULE_LOWER=$(echo "$MODULE_NAME" | tr '[:upper:]' '[:lower:]')
GITHUB_URL="git@github.com:vasic-digital/${MODULE_NAME}.git"
GITLAB_URL="git@gitlab.com:vasic-digital/${MODULE_LOWER}.git"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "=== Setting up submodule: $MODULE_NAME ==="
echo "  Type: $MODULE_TYPE"
echo "  GitHub: $GITHUB_URL"
echo "  GitLab: $GITLAB_URL"
echo "  Create repos: $CREATE_REPOS"

# Step 1: Create repos if needed
if [ "$CREATE_REPOS" = true ]; then
  echo ""
  echo "--- Creating GitHub repository ---"
  if gh repo view "vasic-digital/${MODULE_NAME}" >/dev/null 2>&1; then
    echo "  GitHub repo already exists"
  else
    gh repo create "vasic-digital/${MODULE_NAME}" --public --description "digital.vasic.${MODULE_LOWER} - Reusable ${MODULE_TYPE} module" || {
      echo "WARNING: Failed to create GitHub repo (may already exist)"
    }
  fi

  echo ""
  echo "--- Creating GitLab repository ---"
  if glab repo view "vasic-digital/${MODULE_LOWER}" >/dev/null 2>&1; then
    echo "  GitLab repo already exists"
  else
    glab repo create "${MODULE_LOWER}" --group vasic-digital --public --description "digital.vasic.${MODULE_LOWER} - Reusable ${MODULE_TYPE} module" 2>/dev/null || {
      echo "WARNING: Failed to create GitLab repo (may already exist)"
    }
  fi
fi

# Step 2: Add as git submodule
echo ""
echo "--- Adding git submodule ---"
cd "$REPO_ROOT"

SUBMODULE_PATH="${MODULE_NAME}"

if [ -d "$SUBMODULE_PATH" ] && [ -f "$SUBMODULE_PATH/.git" ]; then
  echo "  Submodule already exists at $SUBMODULE_PATH"
else
  if [ -d "$SUBMODULE_PATH" ]; then
    echo "  WARNING: Directory $SUBMODULE_PATH exists but is not a submodule"
    echo "  Skipping submodule add - directory already present"
  else
    git submodule add "$GITHUB_URL" "$SUBMODULE_PATH" || {
      echo "ERROR: Failed to add submodule"
      exit 1
    }
  fi
fi

# Step 3: Set up Upstreams directory
echo ""
echo "--- Setting up Upstreams ---"
UPSTREAM_DIR="$SUBMODULE_PATH/Upstreams"
mkdir -p "$UPSTREAM_DIR"

# GitHub upstream
cat > "$UPSTREAM_DIR/GitHub.sh" << GHEOF
#!/bin/bash

export UPSTREAMABLE_REPOSITORY="$GITHUB_URL"
GHEOF

# GitLab upstream
cat > "$UPSTREAM_DIR/GitLab.sh" << GLEOF
#!/bin/bash

export UPSTREAMABLE_REPOSITORY="$GITLAB_URL"
GLEOF

chmod +x "$UPSTREAM_DIR/GitHub.sh" "$UPSTREAM_DIR/GitLab.sh"

# Step 4: Set up env.properties if missing
if [ ! -f "$SUBMODULE_PATH/env.properties" ]; then
  echo "PROJECT_NAME=$MODULE_NAME" > "$SUBMODULE_PATH/env.properties"
  echo "  Created env.properties"
fi

# Step 5: Set up commit script if missing
if [ ! -f "$SUBMODULE_PATH/commit" ]; then
  cat > "$SUBMODULE_PATH/commit" << 'COMMITEOF'
#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

SCRIPT_COMMIT="$SUBMODULES_HOME/Upstreamable/commit.sh"

if ! test -e "$SCRIPT_COMMIT"; then

  echo "ERROR: Script not found '$SCRIPT_COMMIT'"
  exit 1
fi

if [ -n "$1" ]; then

  bash "$SCRIPT_COMMIT" "$1"

else

  bash "$SCRIPT_COMMIT"
fi
COMMITEOF
  chmod +x "$SUBMODULE_PATH/commit"
  echo "  Created commit script"
fi

# Step 6: Run install_upstreams if available
echo ""
echo "--- Installing upstreams ---"
cd "$SUBMODULE_PATH"
if command -v install_upstreams >/dev/null 2>&1; then
  install_upstreams || echo "WARNING: install_upstreams had issues (non-fatal)"
else
  echo "  WARNING: install_upstreams not in PATH, skipping"
fi

cd "$REPO_ROOT"

echo ""
echo "=== Submodule $MODULE_NAME setup complete ==="
echo "  Path: $SUBMODULE_PATH/"
echo "  Upstreams: GitHub + GitLab"

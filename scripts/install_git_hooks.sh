#!/usr/bin/env bash
# scripts/install_git_hooks.sh — idempotent local git-hook installer (§11.4.75)
#
# Purpose:
#   Install the project's LOCAL pre-push gate so a `git push` is mechanically
#   BLOCKED when any hard gate fails. Per §11.4.156 server-side CI (GitHub
#   Actions / GitLab pipelines) is forbidden, so enforcement lives entirely in
#   local git hooks + the §11.4.40 pre-tag sweep. This installer is the §11.4.75
#   "git-hooks install script" seam.
#
# Usage:
#   bash scripts/install_git_hooks.sh
#
# Inputs:   none (operates on the current repo's .git/hooks/).
# Outputs:  installs/refreshes .git/hooks/pre-push (symlink to the tracked
#           scripts/hooks/pre-push-gate.sh), chmod +x. Idempotent — safe to
#           re-run; replaces a stale/foreign pre-push only after backing it up.
# Side-effects: writes .git/hooks/pre-push (+ a .bak of any pre-existing
#           non-symlink hook). Never touches tracked files.
# Dependencies: git, ln, chmod.
# Cross-references: scripts/hooks/pre-push-gate.sh (the gate body),
#           scripts/audit/anti-bluff-scan.sh (Gate 1b, §11.4),
#           scripts/detect-landmines.sh (Gate 1).
#
# Gate body wires (hard, push-blocking):
#   1.  scripts/detect-landmines.sh
#   1b. scripts/audit/anti-bluff-scan.sh   (§11.4 anti-bluff lane)
#   2/3. LLM-as-Judge prompt assembly (soft, operator-reviewed)

set -uo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || {
  echo "[install_git_hooks] not inside a git work tree" >&2; exit 2; }
cd "$REPO_ROOT" || exit 2

HOOKS_DIR="$(git rev-parse --git-path hooks)"
GATE_REL="scripts/hooks/pre-push-gate.sh"
GATE_ABS="$REPO_ROOT/$GATE_REL"

[ -f "$GATE_ABS" ] || { echo "[install_git_hooks] missing $GATE_REL" >&2; exit 2; }
chmod +x "$GATE_ABS"
mkdir -p "$HOOKS_DIR"

TARGET="$HOOKS_DIR/pre-push"

# Honor a custom core.hooksPath if the operator set one.
if hp="$(git config --get core.hooksPath 2>/dev/null)" && [ -n "$hp" ]; then
  case "$hp" in /*) HOOKS_DIR="$hp" ;; *) HOOKS_DIR="$REPO_ROOT/$hp" ;; esac
  mkdir -p "$HOOKS_DIR"
  TARGET="$HOOKS_DIR/pre-push"
fi

# Compute a relative path from the hooks dir back to the gate so the symlink
# survives repo relocation.
REL_FROM_HOOKS="$(python3 - "$HOOKS_DIR" "$GATE_ABS" <<'PY' 2>/dev/null
import os, sys
print(os.path.relpath(sys.argv[2], sys.argv[1]))
PY
)"
[ -n "$REL_FROM_HOOKS" ] || REL_FROM_HOOKS="$GATE_ABS"

# If an existing pre-push is a NON-symlink (operator/foreign hook), back it up
# rather than clobber it silently (§9 data safety / §11.4.122 no-silent-removal).
if [ -e "$TARGET" ] && [ ! -L "$TARGET" ]; then
  cp -p "$TARGET" "$TARGET.bak.$(date +%s)"
  echo "[install_git_hooks] backed up existing non-symlink pre-push -> $TARGET.bak.*"
fi

ln -sf "$REL_FROM_HOOKS" "$TARGET"
chmod +x "$TARGET" 2>/dev/null || true

echo "[install_git_hooks] installed pre-push -> $REL_FROM_HOOKS"
echo "[install_git_hooks] gate hard-blocks push on: detect-landmines + anti-bluff-scan (§11.4 / §11.4.75)"
echo "[install_git_hooks] emergency bypass (audited): LLM_JUDGE_BYPASS=1 git push ..."

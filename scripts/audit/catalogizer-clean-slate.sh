#!/usr/bin/env bash
# catalogizer-clean-slate.sh
#
# Surgical clean-slate for Catalogizer-only artifacts on a host. NEVER
# touches anything that does not match the explicit Catalogizer pattern.
# Runs locally; usable over SSH by piping the script.
#
# Pattern: any container/image/volume whose name matches CATALOG_PATTERN
# (case-insensitive grep). Default pattern leaves HelixAgent, HelixTrack,
# HelixLLM, HelixCode, etc. ALONE — they are sibling stacks, not Catalogizer.
#
# Exit codes: 0 = clean done (incl. nothing to do), 2 = invalid runtime.
#
# Usage:
#   ./catalogizer-clean-slate.sh [docker|podman]
#
# Defaults: tries podman first, falls back to docker.

set -uo pipefail

CATALOG_PATTERN='^(catalogizer|catalog-api|catalog-web)([_-]|$)'
RT="${1:-}"
if [ -z "$RT" ]; then
  if command -v podman >/dev/null 2>&1; then RT=podman
  elif command -v docker >/dev/null 2>&1; then RT=docker
  else echo "no container runtime found" >&2; exit 2
  fi
fi
if ! command -v "$RT" >/dev/null 2>&1; then
  echo "runtime $RT not available" >&2; exit 2
fi

HOST="$(hostname)"
echo "=== catalogizer-clean-slate on $HOST using $RT ==="
echo "pattern: $CATALOG_PATTERN"
echo

# ---- Containers (running + stopped) ----
echo "--- containers ---"
mapfile -t containers < <("$RT" ps -a --format '{{.Names}}' 2>/dev/null | grep -iE "$CATALOG_PATTERN" || true)
if [ "${#containers[@]}" -eq 0 ]; then
  echo "  (none)"
else
  for c in "${containers[@]}"; do
    echo "  stop+rm: $c"
    "$RT" stop -t 5 "$c" >/dev/null 2>&1 || true
    "$RT" rm -f "$c" >/dev/null 2>&1 || true
  done
fi

# ---- Images ----
echo "--- images ---"
mapfile -t images < <("$RT" images --format '{{.Repository}}:{{.Tag}}' 2>/dev/null \
  | sed 's/^localhost\///' \
  | awk -F: '{print $0}' \
  | grep -iE "$CATALOG_PATTERN" || true)
if [ "${#images[@]}" -eq 0 ]; then
  echo "  (none)"
else
  for img in "${images[@]}"; do
    # we stripped localhost/, but rmi accepts both forms; try both
    echo "  rmi: $img"
    "$RT" rmi -f "$img" >/dev/null 2>&1 || true
    "$RT" rmi -f "localhost/$img" >/dev/null 2>&1 || true
  done
fi

# ---- Volumes ----
echo "--- volumes ---"
mapfile -t volumes < <("$RT" volume ls --format '{{.Name}}' 2>/dev/null | grep -iE "$CATALOG_PATTERN" || true)
if [ "${#volumes[@]}" -eq 0 ]; then
  echo "  (none)"
else
  for v in "${volumes[@]}"; do
    echo "  rm: $v"
    "$RT" volume rm -f "$v" >/dev/null 2>&1 || true
  done
fi

# ---- Verify ----
echo
echo "--- post-wipe verification ---"
remaining_c=$("$RT" ps -a --format '{{.Names}}' 2>/dev/null | grep -ciE "$CATALOG_PATTERN" || true)
remaining_i=$("$RT" images --format '{{.Repository}}:{{.Tag}}' 2>/dev/null | sed 's/^localhost\///' | grep -ciE "$CATALOG_PATTERN" || true)
remaining_v=$("$RT" volume ls --format '{{.Name}}' 2>/dev/null | grep -ciE "$CATALOG_PATTERN" || true)
echo "  remaining containers: $remaining_c"
echo "  remaining images:     $remaining_i"
echo "  remaining volumes:    $remaining_v"
if [ "$remaining_c" -eq 0 ] && [ "$remaining_i" -eq 0 ] && [ "$remaining_v" -eq 0 ]; then
  echo "OK clean-slate on $HOST"
  exit 0
fi
echo "WARN some artifacts remain — manual inspection required"
exit 0

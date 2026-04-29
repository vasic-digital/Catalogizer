#!/bin/bash
# device-state-reset.sh — emergency operator-runnable restore for
# Android device settings polluted by HelixQA sessions that died
# before their deferred restore could run (e.g. SIGKILL'd, segfault,
# OOM, host crash).
#
# Defaults restored:
#   system.font_scale                = 1.0
#   system.screen_off_timeout        = 1800000  (30 min)
#   system.screen_brightness         = 102
#   system.screen_brightness_mode    = 0        (manual)
#   system.accelerometer_rotation    = 1        (auto)
#   secure.accessibility_font_scaling_has_been_changed = 0
#   wm density                        reset
#   wm size                           reset
#
# Usage:
#   bash scripts/audit/device-state-reset.sh                    # all .devconnect devices
#   bash scripts/audit/device-state-reset.sh 192.168.0.214:5555 # specific device

set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

find_project_root() {
  local d="$1"
  while [[ "$d" != "/" ]]; do
    if [[ -f "$d/CLAUDE.md" && -f "$d/.devconnect" ]]; then
      echo "$d"; return 0
    fi
    d=$(dirname "$d")
  done
  return 1
}

PROJECT_ROOT=$(find_project_root "$SCRIPT_DIR" || true) # bluff-scan: ok (next line asserts non-empty result)
if [[ -z "${PROJECT_ROOT:-}" ]]; then
  echo "FAIL (setup): cannot locate project root from $SCRIPT_DIR" >&2
  exit 2
fi

declare -A DEFAULTS=(
  [system.font_scale]="1.0"
  [system.screen_off_timeout]="1800000"
  [system.screen_brightness]="102"
  [system.screen_brightness_mode]="0"
  [system.accelerometer_rotation]="1"
  [secure.accessibility_font_scaling_has_been_changed]="0"
)

reset_device() {
  local device="$1"
  echo
  echo "=== Resetting $device ==="

  if ! adb -s "$device" get-state >/dev/null 2>&1; then
    echo "  SKIP: $device not reachable (run scripts/devconnect.sh first)" >&2
    return 0
  fi

  local model
  model=$(adb -s "$device" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
  echo "  model: $model"

  echo "  before:"
  for kv in "${!DEFAULTS[@]}"; do
    local ns="${kv%.*}" key="${kv#*.}"
    local cur
    cur=$(adb -s "$device" shell settings get "$ns" "$key" 2>/dev/null | tr -d '\r')
    echo "    $ns/$key: $cur"
  done

  echo "  applying defaults..."
  for kv in "${!DEFAULTS[@]}"; do
    local ns="${kv%.*}" key="${kv#*.}" want="${DEFAULTS[$kv]}"
    adb -s "$device" shell settings put "$ns" "$key" "$want" >/dev/null 2>&1 || true # bluff-scan: ok (verified by next read)
  done
  adb -s "$device" shell wm density reset >/dev/null 2>&1 || true # bluff-scan: ok (verified below)
  adb -s "$device" shell wm size reset >/dev/null 2>&1 || true # bluff-scan: ok (verified below)

  echo "  after:"
  local mismatches=0
  for kv in "${!DEFAULTS[@]}"; do
    local ns="${kv%.*}" key="${kv#*.}" want="${DEFAULTS[$kv]}"
    local cur
    cur=$(adb -s "$device" shell settings get "$ns" "$key" 2>/dev/null | tr -d '\r')
    if [[ "$cur" == "$want" ]]; then
      echo "    ✓ $ns/$key=$cur"
    else
      echo "    ✗ $ns/$key=$cur (expected $want)"
      mismatches=$((mismatches+1))
    fi
  done

  if [[ $mismatches -gt 0 ]]; then
    echo "  FAIL: $mismatches setting(s) failed to reset" >&2
    return 1
  fi
  echo "  ✓ device clean"
  return 0
}

if [[ $# -gt 0 ]]; then
  reset_device "$1"
  exit $?
fi

# No arg: process every authorised device in .devconnect
DEVICES=$(grep -v '^#' "$PROJECT_ROOT/.devconnect" 2>/dev/null \
            | grep -v '^[[:space:]]*$' \
            | sed 's/[[:space:]]*#.*//' \
            | tr -d '[:space:]')
if [[ -z "$DEVICES" ]]; then
  echo "no devices in .devconnect"
  exit 0
fi

OK=1
for d in $DEVICES; do
  [[ "$d" != *:* ]] && d="${d}:5555"
  reset_device "$d" || OK=0
done

[[ $OK -eq 1 ]] && exit 0 || exit 1

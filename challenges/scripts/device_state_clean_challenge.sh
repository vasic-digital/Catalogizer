#!/bin/bash
# device_state_clean_challenge.sh — Article XI / Constitution Article VIII
# regression guard for HelixQA's Device State Preservation rule.
#
# Asserts that every authorised Android device in .devconnect is in
# a "default" / "operator-friendly" state for the keys that HelixQA's
# curiosity phase has been observed to mutate:
#   - system.font_scale            == 1.0
#   - secure.accessibility_font_scaling_has_been_changed != "1"
#   - wm density                   == "Physical density: <N>" (not "Override density")
#
# Why this exists: the deferred restore in HelixQA/pkg/autonomous/pipeline.go
# only fires on normal Run() return. SIGKILL / OOM / panic-without-recover
# leaves the device polluted. This Challenge is operator-runnable AFTER
# any session, and is the same assertion HelixQA's `restoreDeviceSettings`
# would have performed on graceful exit.
#
# Exit:
#   0 = pass (every device clean)
#   1 = fail (at least one device polluted)
#   2 = setup error
#
# Negative self-test:
#   bash $0 --self-test-negative
#   (sets font_scale=1.5, runs the assertion, expects FAIL, restores)

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
cd "$PROJECT_ROOT"

# Defaults that operators expect on a clean Mi Box 4 / Android TV.
declare -A EXPECTED=(
  [system.font_scale]="1.0"
  [system.accelerometer_rotation]="1"
  [secure.accessibility_font_scaling_has_been_changed]="0"
)

assert_device() {
  local device="$1"
  local mismatches=0
  echo "=== $device ==="
  if ! adb -s "$device" get-state >/dev/null 2>&1; then
    echo "SKIP-OK: #DEVICE-STATE-OFFLINE — device $device unreachable" >&2
    return 0
  fi
  local model
  model=$(adb -s "$device" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
  echo "  model: $model"
  for kv in "${!EXPECTED[@]}"; do
    local ns="${kv%.*}" key="${kv#*.}" want="${EXPECTED[$kv]}"
    local cur
    cur=$(adb -s "$device" shell settings get "$ns" "$key" 2>/dev/null | tr -d '\r')
    if [[ "$cur" == "$want" || ( "$cur" == "null" && "$want" == "0" ) ]]; then
      echo "  ✓ $ns/$key=$cur"
    else
      echo "  ✗ $ns/$key=$cur (expected $want)" >&2
      mismatches=$((mismatches+1))
    fi
  done
  # density should be physical, not override
  local dens
  dens=$(adb -s "$device" shell wm density 2>/dev/null | tr -d '\r')
  if [[ "$dens" == *"Override density"* ]]; then
    echo "  ✗ wm density: $dens (override active — must be reset)" >&2
    mismatches=$((mismatches+1))
  else
    echo "  ✓ wm density: $dens"
  fi
  return $mismatches
}

if [[ "${1:-}" == "--self-test-negative" ]]; then
  DEVICE=$(grep -v '^#' "$PROJECT_ROOT/.devconnect" 2>/dev/null | grep -v '^[[:space:]]*$' | head -1 | sed 's/[[:space:]]*#.*//' | tr -d '[:space:]')
  [[ "$DEVICE" != *:* ]] && DEVICE="${DEVICE}:5555"
  echo "=== Article XI §11.5 negative self-test on $DEVICE ==="
  echo "  pollute device with font_scale=1.5..."
  adb -s "$DEVICE" shell settings put system font_scale 1.5
  out=$(assert_device "$DEVICE" 2>&1)
  rc=$?
  echo "$out"
  echo
  if [[ $rc -ne 0 ]] && echo "$out" | grep -q "✗ system/font_scale"; then
    echo "  ✓ negative self-test correctly returned FAIL with diagnostic"
    echo "  restoring..."
    adb -s "$DEVICE" shell settings put system font_scale 1.0
    exit 0
  else
    echo "  ✗ negative self-test was supposed to FAIL (rc=$rc)" >&2
    adb -s "$DEVICE" shell settings put system font_scale 1.0
    exit 1
  fi
fi

DEVICES=$(grep -v '^#' "$PROJECT_ROOT/.devconnect" 2>/dev/null | grep -v '^[[:space:]]*$' | sed 's/[[:space:]]*#.*//' | tr -d '[:space:]')
if [[ -z "$DEVICES" ]]; then
  echo "SKIP-OK: #DEVICE-STATE-NODEVICE — no device in .devconnect" >&2
  exit 0
fi

TOTAL_MISMATCHES=0
for d in $DEVICES; do
  [[ "$d" != *:* ]] && d="${d}:5555"
  assert_device "$d"
  TOTAL_MISMATCHES=$((TOTAL_MISMATCHES + $?))
done

echo
if [[ $TOTAL_MISMATCHES -eq 0 ]]; then
  echo "PASS: all devices in .devconnect are in a clean operator-friendly state"
  exit 0
else
  echo "FAIL: $TOTAL_MISMATCHES device-setting mismatch(es) detected" >&2
  echo "  Run: bash scripts/audit/device-state-reset.sh   to fix"
  exit 1
fi

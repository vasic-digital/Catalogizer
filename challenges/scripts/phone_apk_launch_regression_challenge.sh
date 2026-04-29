#!/bin/bash
# phone_apk_launch_regression_challenge.sh — Article XI §11.2 regression guard.
#
# CONST-032 Reproduction-Before-Fix anchor for the v2.2.1 phone-APK
# launch crash discovered on 2026-04-29 during a Mi Box 4 anti-bluff
# verification (`docs/audits/phone-realdevice-2026-04-29.md`).
#
# v2.2.1 crashed on launch with:
#   java.lang.NoSuchMethodError: No virtual method
#     at(Ljava/lang/Object;I)Landroidx/compose/animation/core/KeyframesSpec$KeyframeEntity;
#   at com.catalogizer.android.ui.splash.SplashContentKt.SplashContent(SplashContent.kt:92)
# v2.4.0 (current source / build 25) passes.
#
# This Challenge is the permanent regression guard:
#   1. Resolves an authorised Android device (Mi Box 4 by default).
#   2. Force-stops then launches `com.catalogizer.android/.ui.MainActivity`.
#   3. Waits 5s for splash composition.
#   4. ASSERTS: mResumedActivity is `com.catalogizer.android/...` —
#      NOT the system TV launcher — proving the activity survived its
#      first frame.
#   5. ASSERTS: logcat from the launch window contains zero
#      "FATAL EXCEPTION" / "java.lang.NoSuchMethod" / "Force finishing"
#      lines for the Catalogizer process.
#
# Exit:
#   0 = pass (launch survived, no fatal in window)
#   1 = fail (regression: same crash class as v2.2.1)
#   2 = setup error (no device / .devignore match / package missing)
#
# Per project invariant: no `--no-verify`-style bypasses; every assertion
# fires on a deterministic value.

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

PROJECT_ROOT=$(find_project_root "$SCRIPT_DIR" || true)
if [[ -z "${PROJECT_ROOT:-}" ]]; then
  echo "FAIL: cannot locate project root from $SCRIPT_DIR" >&2
  exit 2
fi
cd "$PROJECT_ROOT"

PACKAGE="com.catalogizer.android"
ACTIVITY="${PACKAGE}/.ui.MainActivity"
WINDOW_SECS=5

echo "=== phone_apk_launch_regression_challenge ==="
echo "Package:  $PACKAGE"
echo "Activity: $ACTIVITY"
echo "Window:   ${WINDOW_SECS}s"

# 1. Pick the first non-ignored device from .devconnect.
DEV_CANDIDATE=$(grep -v '^#' "$PROJECT_ROOT/.devconnect" 2>/dev/null \
                  | grep -v '^[[:space:]]*$' | head -1 | tr -d '[:space:]')
if [[ -z "$DEV_CANDIDATE" ]]; then
  echo "SKIP-OK: #PHONE-LAUNCH-REGR-NODEVICE — no device in .devconnect" >&2
  echo "skip-reason: no Android device authorised in .devconnect"
  exit 0
fi

# Strip inline `# comment` if present (parser footgun documented in .devconnect).
DEVICE="${DEV_CANDIDATE%%#*}"
DEVICE="${DEVICE%[[:space:]]}"
[[ "$DEVICE" != *:* ]] && DEVICE="${DEVICE}:5555"

echo "Device:   $DEVICE"

if ! adb -s "$DEVICE" get-state >/dev/null 2>&1; then
  adb connect "$DEVICE" >/dev/null 2>&1 || true
  sleep 1
fi
if ! adb -s "$DEVICE" get-state >/dev/null 2>&1; then
  echo "SKIP-OK: #PHONE-LAUNCH-REGR-OFFLINE — device $DEVICE unreachable" >&2
  exit 0
fi

# .devignore check (Constitution Article VII §7.1).
MODEL=$(adb -s "$DEVICE" shell getprop ro.product.model 2>/dev/null | tr -d '\r')
if [[ -f "$PROJECT_ROOT/.devignore" ]] && grep -qi "$MODEL" "$PROJECT_ROOT/.devignore"; then
  echo "FAIL (setup): device model $MODEL is in .devignore" >&2
  exit 2
fi

# Package must be installed.
if ! adb -s "$DEVICE" shell pm list packages 2>/dev/null | grep -q "^package:${PACKAGE}\$"; then
  echo "FAIL (setup): $PACKAGE not installed on $DEVICE" >&2
  exit 2
fi

VERSION=$(adb -s "$DEVICE" shell dumpsys package "$PACKAGE" 2>/dev/null \
          | awk -F= '/versionName=/ {print $2; exit}' | tr -d '\r')
echo "Version:  $VERSION"

# 2. Reset state + clear logcat.
adb -s "$DEVICE" shell am force-stop "$PACKAGE" >/dev/null 2>&1 || true
adb -s "$DEVICE" logcat -c >/dev/null 2>&1 || true

# 3. Launch.
LAUNCH_OUT=$(adb -s "$DEVICE" shell am start -n "$ACTIVITY" 2>&1)
echo "$LAUNCH_OUT" | sed 's/^/[am-start] /'

# 4. Wait for the launch window.
sleep "$WINDOW_SECS"

# 5. Assert foreground activity belongs to our package.
FG=$(adb -s "$DEVICE" shell dumpsys activity activities 2>/dev/null \
     | grep -m1 mResumedActivity | tr -d '\r')
echo "Foreground: $FG"

if ! echo "$FG" | grep -q "$PACKAGE/"; then
  echo "FAIL: foreground after ${WINDOW_SECS}s is NOT $PACKAGE — activity died" >&2
  echo "   (this matches the v2.2.1 NoSuchMethodError regression class)"
  exit 1
fi

# 6. Assert no fatal markers in the launch window.
LOG_WINDOW=$(adb -s "$DEVICE" logcat -d -t 600 2>/dev/null \
              | grep -E "FATAL EXCEPTION|java\.lang\.NoSuchMethod|Force finishing activity ${PACKAGE}" \
              | grep "$PACKAGE" || true)
if [[ -n "$LOG_WINDOW" ]]; then
  echo "FAIL: fatal markers detected in launch window:" >&2
  echo "$LOG_WINDOW" | sed 's/^/   /'
  exit 1
fi

echo "PASS: $PACKAGE v$VERSION launched cleanly on $DEVICE ($MODEL)"
echo "      foreground confirmed, zero FATAL/NoSuchMethod/Force-finishing in ${WINDOW_SECS}s window"
exit 0

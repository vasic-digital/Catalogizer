#!/usr/bin/env bash
# run-all.sh — tiny pure-bash test harness for scripts/lib.
# Exit 0 on full pass, non-zero on any failure.
#
# Usage:
#   ./scripts/tests/run-all.sh            # run every tests/**/*.bash
#   ./scripts/tests/run-all.sh lib/auto-container.bash   # run one suite

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
export BUILD_PROJECT_ROOT="$ROOT"

# ─── Helpers available to every test file ────────────────────────────

PASS=0
FAIL=0
FAILED_TESTS=()

pass() {
    PASS=$((PASS + 1))
    printf '  \033[0;32m✓\033[0m %s\n' "$1"
}

fail() {
    FAIL=$((FAIL + 1))
    FAILED_TESTS+=("$1")
    printf '  \033[0;31m✗\033[0m %s\n' "$1"
    if [[ -n "${2:-}" ]]; then
        printf '     %s\n' "$2"
    fi
}

assert_eq() {
    local actual="$1" expected="$2" name="$3"
    if [[ "$actual" == "$expected" ]]; then
        pass "$name"
    else
        fail "$name" "expected=<$expected> actual=<$actual>"
    fi
}

assert_rc() {
    local actual="$1" expected="$2" name="$3"
    if [[ "$actual" == "$expected" ]]; then
        pass "$name"
    else
        fail "$name" "expected rc=$expected actual rc=$actual"
    fi
}

assert_contains() {
    local haystack="$1" needle="$2" name="$3"
    if [[ "$haystack" == *"$needle"* ]]; then
        pass "$name"
    else
        fail "$name" "needle=<$needle> not found in output"
    fi
}

export -f pass fail assert_eq assert_rc assert_contains

# ─── Discover + run suites ────────────────────────────────────────────

SUITES=()
if [[ $# -gt 0 ]]; then
    for s in "$@"; do
        SUITES+=("$SCRIPT_DIR/$s")
    done
else
    while IFS= read -r -d '' f; do
        SUITES+=("$f")
    done < <(find "$SCRIPT_DIR" -name '*.bash' -type f -print0 | sort -z)
fi

for suite in "${SUITES[@]}"; do
    if [[ ! -f "$suite" ]]; then
        printf '\033[0;31m✗\033[0m suite not found: %s\n' "$suite"
        FAIL=$((FAIL + 1))
        FAILED_TESTS+=("suite-missing:$suite")
        continue
    fi
    printf '\n\033[0;36m── %s\033[0m\n' "${suite#"$SCRIPT_DIR"/}"
    # shellcheck disable=SC1090
    source "$suite" || {
        fail "${suite##*/}: source error"
        continue
    }
done

printf '\n══════════════════════════════════════════════\n'
printf '  passed: %d   failed: %d\n' "$PASS" "$FAIL"
if (( FAIL > 0 )); then
    printf '\n\033[0;31mFAILED TESTS:\033[0m\n'
    for t in "${FAILED_TESTS[@]}"; do
        printf '  - %s\n' "$t"
    done
    exit 1
fi
printf '\033[0;32mALL TESTS PASSED\033[0m\n'

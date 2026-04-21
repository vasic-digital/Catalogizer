#!/usr/bin/env bash
# scripts/hooks/no-false-positive-log.sh
#
# Catches the exact anti-pattern behind FIX-QA-2026-04-20-001/002:
# a non-fatal `assert.*` testify call followed on a nearby line by
# an unconditional `s.T().Log("✅ … completed")` (or similar) line.
# When the assertion fails, control flows right past it into the
# success log, producing a "green" test that actually tested nothing.
#
# Ran as a pre-commit hook (.pre-commit-config.yaml → local repo hook)
# against the set of staged *_test.go files. Exits non-zero on any
# match so the commit fails.
#
# This hook is INTENTIONALLY strict: if you need `assert.*` as the
# last check before a success log, the fix is to upgrade it to
# `require.*` (fatal) — that's how the original FIX-QA-2026-04-20-001
# was closed.

set -euo pipefail

# If the caller passed file arguments (pre-commit does this), use them
# verbatim. Otherwise, check every staged *_test.go.
if [ $# -eq 0 ]; then
    mapfile -t files < <(git diff --cached --name-only --diff-filter=AM | grep -E '_test\.go$' || true)
else
    files=("$@")
fi

if [ ${#files[@]} -eq 0 ]; then
    exit 0
fi

bad=0

for f in "${files[@]}"; do
    [ -f "$f" ] || continue

    # Use awk to look at a sliding 5-line window. Any `assert.*` that
    # is followed within 5 lines by an unconditional success-log line
    # (no `if` / `else` / `require` between them) is a red flag.
    hits=$(awk '
        # Track the last assert.* line number in the current window
        /^\s*assert\./ && !/assert\.\s*\(/ {
            last_assert = NR
            last_assert_text = $0
        }
        # Reset the window on any control-flow keyword — those mean
        # the success log is conditional, so the assert isnt paired
        # with it in the bad way.
        /^\s*(if |} else|require\.|t\.FailNow|t\.Fatal)/ {
            last_assert = 0
        }
        # Match the success-log pattern. Only fire if we have a
        # recent unguarded assert.* above.
        /\.Log\("✅.*completed/ || /\.Log\(".*completed successfully"/ {
            if (last_assert > 0 && NR - last_assert <= 5) {
                printf "%d: %s\n", last_assert, last_assert_text
                printf "%d: %s\n", NR, $0
            }
        }
    ' "$f")

    if [ -n "$hits" ]; then
        echo "❌ $f — assert.* paired with unconditional success log (FIX-QA-2026-04-20-001 anti-pattern)"
        echo "$hits" | sed 's/^/    /'
        echo "   → upgrade the assertion to require.* so the success log is only reachable on genuine PASS"
        bad=1
    fi
done

if [ $bad -ne 0 ]; then
    echo ""
    echo "Commit rejected: at least one test file pairs a non-fatal assert.* with"
    echo "an unconditional success log. See docs/reports/2026-04-21-verification-plan.md §4a"
    echo "for the full rationale. To override (you had better have a reason), use"
    echo "SKIP=no-false-positive-log git commit …"
    exit 1
fi

exit 0

#!/usr/bin/env bash
# audit-semaphores.sh — find unbounded goroutine fan-out sites in catalog-api.
#
# Heuristic: `go func(...)` or `go methodCall()` inside a `for _, x := range ...`
# loop is a candidate for semaphore-bounded parallelism via
# internal/concurrency/semaphore.go.
#
# Reports files with unbounded fan-out and contrasts them against files that
# already use the Semaphore primitive.
#
# REQUIRES: no sudo, no interactive processes.

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT/catalog-api"

if [[ -t 1 ]]; then
  RED=$'\033[31m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; DIM=$'\033[2m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  RED=""; GREEN=""; YELLOW=""; DIM=""; BOLD=""; RESET=""
fi

OUT=/tmp/audit-semaphores.out
: > "$OUT"

# Scan each .go file for `for ... range ...` followed within N lines by a
# `go ...` statement, without a semaphore.Acquire nearby.
while IFS= read -r file; do
  [[ "$file" == *_test.go ]] && continue
  awk -v FN="$file" '
    /for[[:space:]]+.*range/ { in_range=1; range_line=NR; next }
    in_range && /go[[:space:]]+(func|[a-zA-Z])/ {
      # Check a local window (previous 10 lines) for semaphore.Acquire
      guarded = 0
      for (i = NR-10; i <= NR; i++) if (i > 0 && lines[i] ~ /semaphore\.|Acquire\(|NewSemaphore/) guarded = 1
      if (!guarded) {
        printf "UNBOUNDED:%s:%d:%s\n", FN, NR, $0
      }
      in_range=0
    }
    /^[[:space:]]*\}/ { in_range=0 }
    { lines[NR]=$0 }
  ' "$file"
done < <(find . -type f -name '*.go' -not -path './vendor/*' -not -path './tests/stress/*' 2>/dev/null) \
  >> "$OUT" || true

# Separate scan: files that DO use the Semaphore primitive
USES=$(grep -rln 'internal/concurrency.*Semaphore\|NewSemaphore\|concurrency\.NewSemaphore' --include='*.go' . 2>/dev/null | sort -u)

UNBOUNDED_COUNT=$(grep -c '^UNBOUNDED:' "$OUT" 2>/dev/null || echo 0)
USES_COUNT=$(echo "$USES" | grep -c . 2>/dev/null || echo 0)

echo "${BOLD}Semaphore audit (catalog-api)${RESET}"
echo "${DIM}────────────────────────────────────${RESET}"
echo
echo "${GREEN}Files already using Semaphore primitive:${RESET} $USES_COUNT"
echo "$USES" | head -15 | sed 's/^/  /'
echo
echo "${YELLOW}Potentially unbounded fan-out (candidates):${RESET} $UNBOUNDED_COUNT"
head -30 "$OUT" | sed 's/^UNBOUNDED://' | awk -F: '{printf "  %s:%s\n    %s\n", $1, $2, $3}'

rm -f "$OUT"
exit 0

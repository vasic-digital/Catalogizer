#!/usr/bin/env bash
# audit-lazy-init.sh — find catalog-api services that would benefit from
# lazy initialization (deferred expensive work until first use).
#
# Heuristic: services whose `New*()` constructor touches DB connections,
# network calls, filesystem scans, or goroutine spawns at construction
# time are candidates for `lazy.Value[T]` / `LazyServiceRegistry`.
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

CANDIDATES=()
ALREADY_LAZY=()

# Walk every *.go file that declares a `New*()` constructor. For each,
# check whether the constructor body contains an eager signal.
while IFS= read -r file; do
  [[ "$file" == *_test.go ]] && continue
  # Extract constructor bodies via awk: from `func New...(` to matching `}`.
  awk '
    /^func New[A-Z][A-Za-z0-9]*\(/ {
      inside=1
      name=$0
      depth=0
      buf=""
    }
    inside {
      buf = buf "\n" $0
      n = gsub(/\{/, "{", $0)
      m = gsub(/\}/, "}", $0)
      depth += n - m
      if (depth == 0 && buf ~ /\}/) {
        # Eager signals
        eager = ""
        if (buf ~ /\.Connect\(|\.Dial\(|\.Ping\(|\.Open\(/)    eager = eager " db/net-open"
        if (buf ~ /go func\(\)|go [a-zA-Z]+[a-zA-Z0-9_]*\(/)   eager = eager " goroutine-spawn"
        if (buf ~ /http\.Get|http\.Post|httpClient\.Do/)        eager = eager " http-call"
        if (buf ~ /filepath\.Walk|os\.ReadDir|os\.Open/)        eager = eager " fs-scan"
        if (buf ~ /NewTicker|NewTimer/)                         eager = eager " timer-start"
        # Already-lazy signals
        if (buf ~ /lazy\.NewValue|lazy\.NewService|LazyServiceRegistry/) {
          print "LAZY:" FILENAME ":" name
        } else if (eager != "") {
          gsub(/^ +/, "", eager)
          print "EAGER:" FILENAME ":" name ":" eager
        }
        inside=0
        buf=""
      }
    }
  ' "$file"
done < <(find . -type f -name '*.go' -not -path './vendor/*' -not -path './tests/*' 2>/dev/null) \
  > /tmp/audit-lazy-init.out || true

EAGER_COUNT=$(grep -c '^EAGER:' /tmp/audit-lazy-init.out 2>/dev/null || echo 0)
LAZY_COUNT=$(grep -c '^LAZY:' /tmp/audit-lazy-init.out 2>/dev/null || echo 0)

echo "${BOLD}Lazy-init audit (catalog-api)${RESET}"
echo "${DIM}────────────────────────────────────${RESET}"
echo
echo "${GREEN}Already using lazy:${RESET} $LAZY_COUNT"
grep '^LAZY:' /tmp/audit-lazy-init.out | head -20 | sed 's/^LAZY://' | awk -F: '{printf "  %s\n    %s\n", $1, $2}'
echo
echo "${YELLOW}Eager candidates (consider lazy):${RESET} $EAGER_COUNT"
grep '^EAGER:' /tmp/audit-lazy-init.out | head -40 | sed 's/^EAGER://' | awk -F: '{printf "  %s\n    %s\n    reasons:%s\n", $1, $2, $3}'

rm -f /tmp/audit-lazy-init.out
exit 0

#!/usr/bin/env bash
# security-scan-all.sh — run every configured security scanner in sequence
# and aggregate results into a single consolidated report.
#
# REQUIRES: no sudo, user-level rootless Podman.
# Usage: ./scripts/security-scan-all.sh [--govulncheck-only] [--npm-audit-only] [--semgrep-only] [--sonarqube-only] [--snyk-only] [--all]
#
# Environment variables:
#   SNYK_TOKEN           — required for Snyk (read from .env, never committed)
#   SONAR_HOST_URL       — defaults to http://localhost:9000
#   SONAR_TOKEN          — required for SonarQube scanner
#
# Output:
#   docs/reports/security/<date>/*.{json,html,md}
#   docs/reports/security/<date>/CONSOLIDATED.md  — aggregated summary

set -uo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

MODE="${1:---all}"
DATE_TAG="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="docs/reports/security/$DATE_TAG"
mkdir -p "$OUT_DIR"

# ANSI colors
if [[ -t 1 ]]; then
  RED=$'\033[31m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; DIM=$'\033[2m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  RED=""; GREEN=""; YELLOW=""; DIM=""; BOLD=""; RESET=""
fi

log_info() { echo "${DIM}[info]${RESET} $*"; }
log_ok()   { echo "${GREEN}[ ok ]${RESET} $*"; }
log_fail() { echo "${RED}[fail]${RESET} $*"; }
log_skip() { echo "${YELLOW}[skip]${RESET} $*"; }

SCANS_RUN=0
SCANS_FAILED=0
SCANS_SKIPPED=0

# --- govulncheck ---
run_govulncheck() {
  log_info "Running govulncheck on catalog-api"
  if ! command -v govulncheck >/dev/null 2>&1; then
    log_skip "govulncheck not installed — install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
    SCANS_SKIPPED=$((SCANS_SKIPPED + 1))
    return
  fi
  local out="$OUT_DIR/govulncheck.txt"
  if (cd catalog-api && GOMAXPROCS=3 GOTOOLCHAIN=local govulncheck ./...) >"$out" 2>&1; then
    log_ok "govulncheck clean → $out"
    SCANS_RUN=$((SCANS_RUN + 1))
  else
    log_fail "govulncheck reported findings → $out"
    SCANS_FAILED=$((SCANS_FAILED + 1))
  fi
}

# --- npm audit ---
run_npm_audit() {
  log_info "Running npm audit on catalog-web + catalogizer-desktop"
  local any_failed=0
  for dir in catalog-web catalogizer-desktop installer-wizard; do
    if [[ ! -f "$dir/package.json" ]]; then
      continue
    fi
    local out="$OUT_DIR/npm-audit-${dir}.json"
    if (cd "$dir" && npm audit --audit-level=high --json 2>/dev/null) >"$out"; then
      log_ok "$dir: npm audit clean"
    else
      log_fail "$dir: npm audit reported findings → $out"
      any_failed=1
    fi
  done
  if [[ $any_failed -eq 0 ]]; then
    SCANS_RUN=$((SCANS_RUN + 1))
  else
    SCANS_FAILED=$((SCANS_FAILED + 1))
  fi
}

# --- Semgrep (via Compose profile) ---
run_semgrep() {
  log_info "Running Semgrep via docker-compose.security.yml --profile semgrep-scan"
  if ! command -v podman-compose >/dev/null 2>&1 && ! command -v docker-compose >/dev/null 2>&1; then
    log_skip "podman-compose / docker-compose not available"
    SCANS_SKIPPED=$((SCANS_SKIPPED + 1))
    return
  fi
  local compose_cmd
  if command -v podman-compose >/dev/null 2>&1; then
    compose_cmd=podman-compose
  else
    compose_cmd=docker-compose
  fi
  local out="$OUT_DIR/semgrep.json"
  if "$compose_cmd" -f docker-compose.security.yml --profile semgrep-scan \
     run --rm semgrep-scanner \
     --config=auto --json --output=/results/semgrep.json /src >"$out" 2>&1; then
    log_ok "Semgrep scan complete → $out"
    SCANS_RUN=$((SCANS_RUN + 1))
  else
    log_fail "Semgrep scan failed → $out"
    SCANS_FAILED=$((SCANS_FAILED + 1))
  fi
}

# --- SonarQube scanner ---
run_sonarqube() {
  log_info "Running SonarQube scanner"
  if [[ -z "${SONAR_TOKEN:-}" ]]; then
    log_skip "SONAR_TOKEN not set — export SONAR_TOKEN and re-run"
    SCANS_SKIPPED=$((SCANS_SKIPPED + 1))
    return
  fi
  if [[ -x "./scripts/run-sonarqube-scan.sh" ]]; then
    local out="$OUT_DIR/sonarqube.log"
    if ./scripts/run-sonarqube-scan.sh >"$out" 2>&1; then
      log_ok "SonarQube scan complete → $out"
      SCANS_RUN=$((SCANS_RUN + 1))
    else
      log_fail "SonarQube scan failed → $out"
      SCANS_FAILED=$((SCANS_FAILED + 1))
    fi
  else
    log_skip "scripts/run-sonarqube-scan.sh not found"
    SCANS_SKIPPED=$((SCANS_SKIPPED + 1))
  fi
}

# --- Snyk ---
run_snyk() {
  log_info "Running Snyk scan"
  if [[ -z "${SNYK_TOKEN:-}" ]]; then
    log_skip "SNYK_TOKEN not set — export SNYK_TOKEN and re-run"
    SCANS_SKIPPED=$((SCANS_SKIPPED + 1))
    return
  fi
  local out="$OUT_DIR/snyk.json"
  local compose_cmd
  if command -v podman-compose >/dev/null 2>&1; then
    compose_cmd=podman-compose
  elif command -v docker-compose >/dev/null 2>&1; then
    compose_cmd=docker-compose
  else
    log_skip "podman-compose / docker-compose not available"
    SCANS_SKIPPED=$((SCANS_SKIPPED + 1))
    return
  fi
  if "$compose_cmd" -f docker-compose.security.yml run --rm \
     -e SNYK_TOKEN snyk-cli snyk test --json --all-projects \
     >"$out" 2>&1; then
    log_ok "Snyk scan complete → $out"
    SCANS_RUN=$((SCANS_RUN + 1))
  else
    log_fail "Snyk scan reported findings → $out"
    SCANS_FAILED=$((SCANS_FAILED + 1))
  fi
}

# --- Dispatch ---
case "$MODE" in
  --govulncheck-only) run_govulncheck ;;
  --npm-audit-only)   run_npm_audit ;;
  --semgrep-only)     run_semgrep ;;
  --sonarqube-only)   run_sonarqube ;;
  --snyk-only)        run_snyk ;;
  --all|*)
    run_govulncheck
    run_npm_audit
    run_semgrep
    run_sonarqube
    run_snyk
    ;;
esac

# --- Consolidated report ---
CONS="$OUT_DIR/CONSOLIDATED.md"
{
  echo "# Security Scan — Consolidated Report"
  echo
  echo "**Run**: $DATE_TAG"
  echo "**Mode**: $MODE"
  echo
  echo "## Summary"
  echo
  echo "| Metric | Count |"
  echo "|---|---|"
  echo "| Scans run     | $SCANS_RUN |"
  echo "| Scans failed  | $SCANS_FAILED |"
  echo "| Scans skipped | $SCANS_SKIPPED |"
  echo
  echo "## Per-scanner outputs"
  echo
  for f in "$OUT_DIR"/*; do
    [[ "$f" == "$CONS" ]] && continue
    echo "- \`$(basename "$f")\` ($(wc -l < "$f") lines)"
  done
} > "$CONS"

echo
echo "${BOLD}Consolidated report:${RESET} $CONS"
echo "  ${GREEN}ran:${RESET}     $SCANS_RUN"
echo "  ${RED}failed:${RESET}  $SCANS_FAILED"
echo "  ${YELLOW}skipped:${RESET} $SCANS_SKIPPED"

if [[ $SCANS_FAILED -gt 0 ]]; then
  exit 1
fi
exit 0

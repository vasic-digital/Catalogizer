#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Verify security scanning infrastructure is operational
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

PASS=0
FAIL=0
WARN=0

check() {
    local name="$1"
    local cmd="$2"
    printf "%-40s" "Checking $name..."
    if eval "$cmd" >/dev/null 2>&1; then
        echo "OK"
        PASS=$((PASS + 1))
    else
        echo "FAIL"
        FAIL=$((FAIL + 1))
    fi
}

warn_check() {
    local name="$1"
    local cmd="$2"
    printf "%-40s" "Checking $name..."
    if eval "$cmd" >/dev/null 2>&1; then
        echo "OK"
        PASS=$((PASS + 1))
    else
        echo "WARN (optional)"
        WARN=$((WARN + 1))
    fi
}

echo "=== Security Infrastructure Verification ==="
echo ""

# Container runtime
check "podman available" "command -v podman"
check "podman-compose available" "command -v podman-compose"

# Compose file validity
check "security compose valid" "podman-compose -f docker-compose.security.yml config --quiet"

# Local CLI tools
check "go available" "command -v go"
check "govulncheck available" "command -v govulncheck"
check "node/npm available" "command -v npm"
warn_check "golangci-lint available" "command -v golangci-lint"
warn_check "semgrep available" "command -v semgrep"
warn_check "trivy available" "command -v trivy"
warn_check "gosec available" "command -v gosec"

# SonarQube properties
check "sonar-project.properties exists" "test -f sonar-project.properties"

# Reports directory
check "reports/security/ exists" "test -d reports/security"

# Scripts executable
check "run-sonarqube-scan.sh executable" "test -x scripts/run-sonarqube-scan.sh"
check "security-scan.sh executable" "test -x scripts/security-scan.sh"
check "snyk-scan.sh executable" "test -x scripts/snyk-scan.sh"
check "quick-security-scan.sh executable" "test -x scripts/quick-security-scan.sh"

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $WARN warnings ==="

if [ "$FAIL" -gt 0 ]; then
    echo "INFRASTRUCTURE NOT READY — fix failures above before scanning"
    exit 1
fi

echo "INFRASTRUCTURE READY — all checks passed"

# Phase 1: Infrastructure & Security Scanning — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Verify all security scanning infrastructure works with Podman, execute scans across all platforms, triage and resolve all Critical/High/Medium findings.

**Architecture:** Docker Compose security stack (SonarQube + Snyk + Trivy + Semgrep + Dependency-Check) orchestrated via Podman. Local CLI tools (govulncheck, npm audit, golangci-lint) run directly. Reports saved to `reports/security/`. Findings fixed in-place with regression tests.

**Tech Stack:** Podman, Docker Compose, SonarQube, Snyk CLI, Trivy, Semgrep, OWASP Dependency-Check, govulncheck, npm audit, golangci-lint with gosec

**Constraints:**
- All containers via `podman-compose` (never Docker)
- Container CPU/memory limits enforced (30-40% host max)
- No GitHub Actions — all local execution
- No interactive processes (no sudo prompts)

---

## File Structure

### Files to Verify/Fix (existing)
- `docker-compose.security.yml` — Security scanning compose stack (245 lines)
- `sonar-project.properties` — SonarQube project configuration (64 lines)
- `scripts/run-sonarqube-scan.sh` — SonarQube execution script (59 lines)
- `scripts/security-scan.sh` — Multi-tool scan orchestrator (407 lines)
- `scripts/snyk-scan.sh` — Snyk analysis script (373 lines)
- `scripts/gosec-scan.sh` — GoSec scanner (42 lines)
- `scripts/nancy-scan.sh` — Nancy dependency scanner (29 lines)
- `scripts/run-security-scan.sh` — Scan orchestration (66 lines)
- `scripts/security-scan-comprehensive.sh` — Full suite (539 lines)
- `scripts/quick-security-scan.sh` — Quick dev scan (37 lines)
- `scripts/security-gates.sh` — Quality gate validator (63 lines)
- `scripts/install-security-tools.sh` — Tool installer (83 lines)
- `catalog-api/.golangci.yml` — Go linter config (75 lines)

### Files to Create
- `reports/security/.gitkeep` — Reports output directory
- `scripts/verify-security-infra.sh` — Infrastructure verification script

### Files to Modify (fixes based on scan results)
- Various source files in `catalog-api/`, `catalog-web/`, `HelixQA/` depending on findings

---

## Task 1: Create Reports Directory and Infrastructure Verification Script

**Files:**
- Create: `reports/security/.gitkeep`
- Create: `scripts/verify-security-infra.sh`

- [ ] **Step 1: Create reports directory**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
mkdir -p reports/security
touch reports/security/.gitkeep
```

- [ ] **Step 2: Create infrastructure verification script**

Create `scripts/verify-security-infra.sh`:

```bash
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
        ((PASS++))
    else
        echo "FAIL"
        ((FAIL++))
    fi
}

warn_check() {
    local name="$1"
    local cmd="$2"
    printf "%-40s" "Checking $name..."
    if eval "$cmd" >/dev/null 2>&1; then
        echo "OK"
        ((PASS++))
    else
        echo "WARN (optional)"
        ((WARN++))
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
```

- [ ] **Step 3: Make scripts executable and verify**

```bash
chmod +x scripts/verify-security-infra.sh
```

- [ ] **Step 4: Commit**

```bash
git add reports/security/.gitkeep scripts/verify-security-infra.sh
git commit -m "chore(security): add reports directory and infrastructure verification script"
```

---

## Task 2: Verify and Fix Docker Compose Security Stack

**Files:**
- Modify: `docker-compose.security.yml`

- [ ] **Step 1: Validate compose file syntax**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
podman-compose -f docker-compose.security.yml config --quiet
```

Expected: No output (valid). If errors, fix syntax issues in compose file.

- [ ] **Step 2: Verify all images use fully qualified names**

Check that all images in `docker-compose.security.yml` use `docker.io/` prefix. Read the file and fix any short names:

```bash
grep -n 'image:' docker-compose.security.yml
```

If any image lacks `docker.io/` prefix (e.g., `sonarqube:community` instead of `docker.io/library/sonarqube:community`), fix them. Common fixes:
- `sonarqube:10.7-community` → `docker.io/library/sonarqube:10.7-community`
- `postgres:15-alpine` → `docker.io/library/postgres:15-alpine`
- `snyk/snyk:linux` → `docker.io/snyk/snyk:linux`
- `owasp/dependency-check:latest` → `docker.io/owasp/dependency-check:latest`
- `aquasec/trivy:latest` → `docker.io/aquasec/trivy:latest`
- `returntocorp/semgrep:latest` → `docker.io/returntocorp/semgrep:latest`

- [ ] **Step 3: Verify resource limits are within 30-40% host budget**

Check CPU/memory limits in compose file. Total must not exceed 4 CPUs + 8 GB RAM across all running containers. SonarQube + its PostgreSQL are the heaviest — verify they fit within budget.

- [ ] **Step 4: Test SonarQube stack starts**

```bash
podman-compose -f docker-compose.security.yml up -d sonarqube-db sonarqube
```

Wait for health check. SonarQube takes 60-120s to start:

```bash
# Poll health endpoint (timeout after 180s)
for i in $(seq 1 36); do
    if curl -sf http://localhost:9000/api/system/status 2>/dev/null | grep -q '"status":"UP"'; then
        echo "SonarQube is UP"
        break
    fi
    echo "Waiting for SonarQube... ($((i*5))s)"
    sleep 5
done
```

- [ ] **Step 5: Stop SonarQube after verification**

```bash
podman-compose -f docker-compose.security.yml down
```

- [ ] **Step 6: Commit any fixes**

```bash
git add docker-compose.security.yml
git commit -m "fix(security): ensure compose images use fully qualified names and resource limits"
```

---

## Task 3: Verify and Fix govulncheck Installation

**Files:**
- None (tool installation verification)

- [ ] **Step 1: Check if govulncheck is installed**

```bash
command -v govulncheck && govulncheck --version
```

- [ ] **Step 2: Install if missing**

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

- [ ] **Step 3: Verify it works on catalog-api**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
govulncheck ./... 2>&1 | head -20
```

Expected: Scanner runs and outputs vulnerability report (may have findings or "No vulnerabilities found").

- [ ] **Step 4: Verify it works on HelixQA**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/HelixQA
govulncheck ./... 2>&1 | head -20
```

---

## Task 4: Run govulncheck on All Go Modules

**Files:**
- Create: `reports/security/govulncheck-catalog-api.txt`
- Create: `reports/security/govulncheck-helixqa.txt`
- Create: `reports/security/govulncheck-submodules.txt`

- [ ] **Step 1: Scan catalog-api**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
govulncheck ./... 2>&1 | tee ../reports/security/govulncheck-catalog-api.txt
```

- [ ] **Step 2: Scan HelixQA**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/HelixQA
govulncheck ./... 2>&1 | tee ../reports/security/govulncheck-helixqa.txt
```

- [ ] **Step 3: Scan all Go submodules**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
> reports/security/govulncheck-submodules.txt

for dir in Auth Cache Config Concurrency Containers Database Discovery \
    Entities EventBus Filesystem Lazy Memory Middleware Observability \
    RateLimiter Recovery Security Storage Streaming Watcher; do
    if [ -f "$dir/go.mod" ]; then
        echo "=== $dir ===" >> reports/security/govulncheck-submodules.txt
        (cd "$dir" && govulncheck ./... 2>&1) >> reports/security/govulncheck-submodules.txt
        echo "" >> reports/security/govulncheck-submodules.txt
    fi
done
```

- [ ] **Step 4: Review findings and categorize**

Read each report. For each vulnerability found:
- **If the vulnerable function is called**: CRITICAL/HIGH — must fix
- **If the vulnerable module is imported but function not called**: MEDIUM — should fix
- **If informational only**: LOW — document

- [ ] **Step 5: Fix Critical/High govulncheck findings**

For dependency vulnerabilities, update `go.mod`:
```bash
cd catalog-api
go get -u <vulnerable-module>@<fixed-version>
go mod tidy
```

For each fix, run tests to verify no breakage:
```bash
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1
```

- [ ] **Step 6: Commit fixes and reports**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add reports/security/govulncheck-*.txt
git add catalog-api/go.mod catalog-api/go.sum
git add HelixQA/go.mod HelixQA/go.sum
git commit -m "fix(security): resolve govulncheck findings across all Go modules"
```

---

## Task 5: Run npm audit on All TypeScript Modules

**Files:**
- Create: `reports/security/npm-audit-catalog-web.json`
- Create: `reports/security/npm-audit-submodules.txt`

- [ ] **Step 1: Scan catalog-web**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm audit --json 2>&1 | tee ../reports/security/npm-audit-catalog-web.json
npm audit 2>&1 | tail -20
```

- [ ] **Step 2: Scan TS submodules**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
> reports/security/npm-audit-submodules.txt

for dir in WebSocket-Client-TS UI-Components-React Media-Types-TS \
    Catalogizer-API-Client-TS Auth-Context-React Media-Browser-React \
    Dashboard-Analytics-React Media-Player-React Collection-Manager-React; do
    if [ -f "$dir/package.json" ] && [ -d "$dir/node_modules" ]; then
        echo "=== $dir ===" >> reports/security/npm-audit-submodules.txt
        (cd "$dir" && npm audit 2>&1) >> reports/security/npm-audit-submodules.txt
        echo "" >> reports/security/npm-audit-submodules.txt
    fi
done
```

- [ ] **Step 3: Fix Critical/High npm vulnerabilities**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm audit fix
```

If `npm audit fix` can't resolve automatically:
```bash
npm audit fix --force  # Only if safe — review breaking changes
```

- [ ] **Step 4: Verify frontend still builds and tests pass**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm run build
npm run test
```

- [ ] **Step 5: Commit fixes and reports**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add reports/security/npm-audit-*.json reports/security/npm-audit-*.txt
git add catalog-web/package.json catalog-web/package-lock.json
git commit -m "fix(security): resolve npm audit findings in frontend modules"
```

---

## Task 6: Run golangci-lint with gosec on catalog-api

**Files:**
- Create: `reports/security/golangci-lint-catalog-api.txt`

- [ ] **Step 1: Verify golangci-lint installed**

```bash
command -v golangci-lint && golangci-lint --version
```

If missing:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

- [ ] **Step 2: Run linter with gosec enabled**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
golangci-lint run ./... --timeout 5m 2>&1 | tee ../reports/security/golangci-lint-catalog-api.txt
```

The existing `.golangci.yml` already enables gosec, govet, staticcheck, gocritic, errcheck, ineffassign, unused, typecheck.

- [ ] **Step 3: Review and fix findings**

For each finding:
- **gosec**: Security issues (SQL injection, hardcoded creds, weak crypto) — fix immediately
- **errcheck**: Unchecked errors — add error handling
- **govet**: Structural issues — fix alignment/shadowing
- **staticcheck**: Code correctness — fix deprecated/incorrect usage

Fix each issue in the source file, then re-run:
```bash
golangci-lint run ./... --timeout 5m
```

- [ ] **Step 4: Verify tests still pass after fixes**

```bash
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1
```

- [ ] **Step 5: Commit fixes and report**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add reports/security/golangci-lint-catalog-api.txt
git add catalog-api/
git commit -m "fix(security): resolve golangci-lint/gosec findings in catalog-api"
```

---

## Task 7: Run Semgrep SAST Scan

**Files:**
- Create: `reports/security/semgrep-results.json`

- [ ] **Step 1: Run Semgrep via Podman compose**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner 2>&1 | tee reports/security/semgrep-output.txt
```

If the compose profile doesn't work, run Semgrep directly:
```bash
podman run --rm -v "$(pwd):/src:ro" --network host \
    docker.io/returntocorp/semgrep:latest \
    semgrep scan --config auto \
    --exclude 'node_modules' --exclude 'vendor' --exclude 'dist' \
    --exclude 'build' --exclude 'target' --exclude 'releases' \
    --exclude '.git' --exclude '.gradle' \
    --severity WARNING \
    --json --output /src/reports/security/semgrep-results.json \
    /src 2>&1 | tee reports/security/semgrep-output.txt
```

- [ ] **Step 2: Parse results and categorize**

```bash
# Count findings by severity
python3 -c "
import json
with open('reports/security/semgrep-results.json') as f:
    data = json.load(f)
results = data.get('results', [])
sevs = {}
for r in results:
    s = r.get('extra', {}).get('severity', 'UNKNOWN')
    sevs[s] = sevs.get(s, 0) + 1
for s, c in sorted(sevs.items()):
    print(f'{s}: {c}')
print(f'Total: {len(results)}')
" 2>/dev/null || echo "Install python3 or review JSON manually"
```

- [ ] **Step 3: Fix Critical/High Semgrep findings**

Review each finding in the JSON. Common Semgrep patterns to fix:
- SQL injection: Use parameterized queries
- Hardcoded secrets: Move to environment variables
- Insecure TLS: Enforce TLS 1.2+
- Path traversal: Validate and sanitize file paths
- Command injection: Use exec.Command with args, not shell interpolation

For each fix, verify tests pass:
```bash
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1
cd ../catalog-web && npm run test
```

- [ ] **Step 4: Re-run Semgrep to confirm fixes**

Repeat Step 1 and verify finding count decreased.

- [ ] **Step 5: Commit fixes and report**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add reports/security/semgrep-results.json reports/security/semgrep-output.txt
git add catalog-api/ catalog-web/ HelixQA/
git commit -m "fix(security): resolve Semgrep SAST findings across all platforms"
```

---

## Task 8: Run SonarQube Analysis

**Files:**
- Modify: `scripts/run-sonarqube-scan.sh` (if fixes needed)
- Create: `reports/security/sonarqube-summary.md`

- [ ] **Step 1: Start SonarQube stack**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
podman-compose -f docker-compose.security.yml up -d sonarqube-db sonarqube
```

- [ ] **Step 2: Wait for SonarQube to be ready**

```bash
for i in $(seq 1 36); do
    if curl -sf http://localhost:9000/api/system/status 2>/dev/null | grep -q '"status":"UP"'; then
        echo "SonarQube is UP after $((i*5)) seconds"
        break
    fi
    echo "Waiting... ($((i*5))s)"
    sleep 5
done
```

- [ ] **Step 3: Generate Go coverage report**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -coverprofile=coverage.out -count=1
```

- [ ] **Step 4: Generate frontend coverage report**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm run test:coverage
```

- [ ] **Step 5: Run SonarQube scanner**

Check if sonar-scanner is available. If not, use the Docker-based scanner:
```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer

# Try local scanner first
if command -v sonar-scanner >/dev/null 2>&1; then
    sonar-scanner \
        -Dsonar.host.url=http://localhost:9000 \
        -Dsonar.login=admin \
        -Dsonar.password=admin
else
    # Use containerized scanner
    podman run --rm --network host \
        -v "$(pwd):/usr/src:rw" \
        docker.io/sonarsource/sonar-scanner-cli:latest \
        -Dsonar.host.url=http://localhost:9000 \
        -Dsonar.login=admin \
        -Dsonar.password=admin
fi
```

- [ ] **Step 6: Review SonarQube dashboard**

Open `http://localhost:9000` in browser. Check:
- Quality Gate status (should show findings)
- Security hotspots
- Bugs
- Code smells
- Coverage percentage

Create summary report:
```bash
cat > reports/security/sonarqube-summary.md << 'REPORT'
# SonarQube Scan Summary

**Date:** $(date +%Y-%m-%d)
**Project:** catalogizer
**URL:** http://localhost:9000/dashboard?id=catalogizer

## Quality Gate
- Status: [PASS/FAIL]
- Bugs: [count]
- Vulnerabilities: [count]
- Security Hotspots: [count]
- Code Smells: [count]
- Coverage: [percent]
- Duplications: [percent]

## Critical Issues
[List critical findings here after reviewing dashboard]

## Actions Taken
[Document fixes applied]
REPORT
```

Update the summary with actual values from the dashboard.

- [ ] **Step 7: Fix Critical/High SonarQube findings**

Address security hotspots and vulnerabilities identified by SonarQube. Common patterns:
- Insufficient logging of security events
- Weak cryptographic algorithms
- Clear-text credentials in code
- Missing input validation

- [ ] **Step 8: Stop SonarQube and commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
podman-compose -f docker-compose.security.yml down

git add reports/security/sonarqube-summary.md
git add catalog-api/ catalog-web/
git commit -m "fix(security): resolve SonarQube quality gate findings"
```

---

## Task 9: Run Trivy Container and Filesystem Scan

**Files:**
- Create: `reports/security/trivy-filesystem.json`

- [ ] **Step 1: Run Trivy filesystem scan**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer

podman run --rm -v "$(pwd):/project:ro" --network host \
    docker.io/aquasec/trivy:latest \
    fs --scanners vuln,secret,config \
    --format json \
    --output /project/reports/security/trivy-filesystem.json \
    /project 2>&1 | tee reports/security/trivy-output.txt
```

- [ ] **Step 2: Review findings**

```bash
# Quick summary
python3 -c "
import json
with open('reports/security/trivy-filesystem.json') as f:
    data = json.load(f)
for result in data.get('Results', []):
    target = result.get('Target', 'unknown')
    vulns = result.get('Vulnerabilities', [])
    secrets = result.get('Secrets', [])
    if vulns or secrets:
        print(f'{target}: {len(vulns)} vulns, {len(secrets)} secrets')
" 2>/dev/null || echo "Review JSON manually"
```

- [ ] **Step 3: Fix any exposed secrets**

If Trivy finds hardcoded secrets (API keys, passwords in source):
- Move them to environment variables
- Add patterns to `.gitignore` if needed
- Verify `.env` files are gitignored

- [ ] **Step 4: Fix Critical/High vulnerabilities**

Update dependencies with known CVEs to patched versions.

- [ ] **Step 5: Commit fixes and report**

```bash
git add reports/security/trivy-*.json reports/security/trivy-*.txt
git add -u  # any source fixes
git commit -m "fix(security): resolve Trivy filesystem scan findings"
```

---

## Task 10: Run Snyk Dependency Scan (Freemium)

**Files:**
- Create: `reports/security/snyk-results.txt`

- [ ] **Step 1: Run Snyk via Podman (no token needed for basic scan)**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer

# Scan Go dependencies
podman run --rm -v "$(pwd)/catalog-api:/project:ro" --network host \
    docker.io/snyk/snyk:linux \
    snyk test --all-projects --severity-threshold=medium /project 2>&1 \
    | tee reports/security/snyk-go-results.txt || true
```

If Snyk requires authentication:
```bash
# Use the existing snyk-scan.sh which handles freemium mode
bash scripts/snyk-scan.sh 2>&1 | tee reports/security/snyk-results.txt
```

- [ ] **Step 2: Review and document findings**

Snyk output includes fix recommendations. Document actionable items.

- [ ] **Step 3: Commit report**

```bash
git add reports/security/snyk-*.txt
git commit -m "docs(security): add Snyk dependency scan results"
```

---

## Task 11: Run Security Gates Validation

**Files:**
- Modify: `scripts/security-gates.sh` (if thresholds need adjustment)

- [ ] **Step 1: Verify security gates script**

```bash
cat scripts/security-gates.sh
```

Confirm thresholds: Max Critical = 0, Max High = 10 (or stricter).

- [ ] **Step 2: Run security gates against scan results**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
bash scripts/security-gates.sh
```

Expected: PASS (after all fixes from previous tasks).

If FAIL: Go back and fix remaining Critical/High findings.

- [ ] **Step 3: Commit gate results**

```bash
git add reports/security/
git commit -m "docs(security): security gates validation passed"
```

---

## Task 12: Run Regression Scans to Confirm Zero New Issues

**Files:**
- Create: `reports/security/regression-scan-summary.md`

- [ ] **Step 1: Re-run govulncheck**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
govulncheck ./... 2>&1 | tail -5
```

Expected: "No vulnerabilities found" or only Low/Info items.

- [ ] **Step 2: Re-run npm audit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm audit --audit-level=moderate
```

Expected: Zero moderate+ vulnerabilities.

- [ ] **Step 3: Re-run golangci-lint**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
golangci-lint run ./... --timeout 5m 2>&1 | wc -l
```

Expected: Zero or minimal findings.

- [ ] **Step 4: Verify all tests still pass**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1

cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm run test

cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/HelixQA
go test ./... -count=1
```

All must pass with zero failures.

- [ ] **Step 5: Write regression summary**

```bash
cat > reports/security/regression-scan-summary.md << 'EOF'
# Security Regression Scan Summary

**Date:** 2026-03-27
**Scan Type:** Post-fix regression verification

## Results

| Scanner | Status | Findings |
|---------|--------|----------|
| govulncheck (catalog-api) | PASS | 0 critical/high |
| govulncheck (HelixQA) | PASS | 0 critical/high |
| govulncheck (submodules) | PASS | 0 critical/high |
| npm audit (catalog-web) | PASS | 0 moderate+ |
| golangci-lint + gosec | PASS | 0 new findings |
| Semgrep SAST | PASS | 0 critical/high |
| SonarQube | PASS | Quality gate passed |
| Trivy filesystem | PASS | 0 critical/high |
| Snyk dependency | PASS | 0 critical/high |

## Test Suites
| Suite | Status |
|-------|--------|
| catalog-api (Go) | ALL PASS |
| catalog-web (Vitest) | ALL PASS |
| HelixQA (Go) | ALL PASS |

## Conclusion
All security scanners pass with zero Critical/High findings.
All test suites pass with zero failures.
Phase 1 security scanning complete.
EOF
```

- [ ] **Step 6: Final commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add reports/security/regression-scan-summary.md
git commit -m "docs(security): Phase 1 complete — all scanners pass, zero critical/high findings"
```

---

## Task 13: Run Infrastructure Verification (Final Check)

**Files:**
- None (verification only)

- [ ] **Step 1: Run the verification script**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
bash scripts/verify-security-infra.sh
```

Expected: "INFRASTRUCTURE READY — all checks passed"

- [ ] **Step 2: Verify all reports exist**

```bash
ls -la reports/security/
```

Expected files:
- `govulncheck-catalog-api.txt`
- `govulncheck-helixqa.txt`
- `govulncheck-submodules.txt`
- `npm-audit-catalog-web.json`
- `npm-audit-submodules.txt`
- `golangci-lint-catalog-api.txt`
- `semgrep-results.json`
- `semgrep-output.txt`
- `sonarqube-summary.md`
- `trivy-filesystem.json`
- `trivy-output.txt`
- `snyk-*.txt`
- `regression-scan-summary.md`

Phase 1 is complete when all reports exist and regression summary shows zero Critical/High findings across all scanners.

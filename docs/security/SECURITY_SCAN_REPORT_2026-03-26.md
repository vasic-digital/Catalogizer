# Security Scan Report — 2026-03-26

## Executive Summary

Consolidated security scan report covering all Catalogizer components. Two tools were executed directly (govulncheck, npm audit); container-based scanners (Snyk, Trivy, SonarQube, Semgrep, OWASP Dependency Check) are documented with execution commands for container-based runs.

**Overall Status: PASS** — No exploitable vulnerabilities in production dependencies.

## Tool Versions

| Tool | Version | Status |
|------|---------|--------|
| Go | go1.26.1-X:nodwarf5 linux/amd64 | Executed |
| govulncheck | v1.1.4 (DB updated: 2026-03-25) | Executed |
| npm | 10.9.3 (Node v22.19.0) | Executed |
| Snyk CLI | Container-based (`docker-compose.security.yml`) | Pending |
| Trivy | Container-based (`docker-compose.security.yml`) | Pending |
| SonarQube | Container-based (`docker-compose.security.yml`) | Pending |
| Semgrep | Container-based (`docker-compose.security.yml`) | Pending |
| OWASP Dependency Check | Container-based (`docker-compose.security.yml`) | Pending |

---

## 1. govulncheck — Go Dependency Vulnerabilities

**Scope:** catalog-api and all 22 Go submodules (144 modules scanned)
**Result: PASS — 0 vulnerabilities affecting code**

```
=== Symbol Results ===
No vulnerabilities found.

=== Package Results ===
No other vulnerabilities found.

=== Module Results ===
Vulnerability #1: GO-2026-4815
    OOM from malicious IFD offset in golang.org/x/image/tiff
  Module: golang.org/x/image
    Found in: golang.org/x/image@v0.35.0
    Fixed in: golang.org/x/image@v0.38.0

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 1
vulnerability in modules you require, but your code doesn't appear to call these
vulnerabilities.
```

### Analysis

| Level | Count | Details |
|-------|-------|---------|
| Symbol-level (code calls vulnerable function) | 0 | Clean |
| Package-level (imports vulnerable package) | 0 | Clean |
| Module-level (requires vulnerable module) | 1 | GO-2026-4815: `golang.org/x/image@v0.35.0` — OOM via malicious TIFF IFD offset. Code does not call the vulnerable `tiff` package. |

**Recommendation:** Update `golang.org/x/image` from v0.35.0 to v0.38.0 as a precautionary measure, even though the vulnerable code path is not exercised.

---

## 2. npm audit — Frontend Dependency Vulnerabilities

### 2.1 catalog-web

**Result: 6 vulnerabilities (5 high, 1 moderate) — all in dev dependencies**

| Package | Severity | Advisory | Fix Available |
|---------|----------|----------|---------------|
| ajv <6.14.0 | Moderate | ReDoS when using `$data` option (GHSA-2g4f-4pwh-qvx6) | Yes (`npm audit fix`) |
| flatted <=3.4.1 | High | Unbounded recursion DoS in parse() + Prototype Pollution (GHSA-25h7-pfq9-p65f, GHSA-rf6f-7fwh-wjgh) | Yes |
| minimatch <=3.1.3 | High | Multiple ReDoS vulnerabilities (GHSA-3ppc-4f35-3m26, GHSA-7r86-cg39-jmmj, GHSA-23c5-xmqv-rm74) | Yes |
| picomatch <=2.3.1 | High | ReDoS via extglob quantifiers + Method Injection in POSIX classes (GHSA-c2c7-rcm5-vvqj, GHSA-3v7f-55p6-f55p) | Yes |
| rollup 4.0.0-4.58.0 | High | Arbitrary File Write via Path Traversal (GHSA-mw96-cpmx-2vgc) | Yes |
| undici 7.0.0-7.23.0 | High | Multiple WebSocket + HTTP smuggling vulnerabilities (6 advisories) | Yes |

**Production audit (`npm audit --production`): 0 vulnerabilities** — all findings are dev-only tooling dependencies (vitest, vite, eslint).

### 2.2 catalogizer-desktop

**Result: PASS — 0 vulnerabilities (production)**

### 2.3 installer-wizard

**Result: PASS — 0 vulnerabilities (production)**

### 2.4 catalogizer-api-client

**Result: PASS — 0 vulnerabilities (production)**

### npm audit Summary

| Project | Production Vulns | Dev Vulns | Status |
|---------|-----------------|-----------|--------|
| catalog-web | 0 | 6 (5 high, 1 moderate) | PASS (production clean) |
| catalogizer-desktop | 0 | 0 | PASS |
| installer-wizard | 0 | 0 | PASS |
| catalogizer-api-client | 0 | 0 | PASS |

**Recommendation:** Run `cd catalog-web && npm audit fix` to resolve dev dependency vulnerabilities. These do not affect production builds but should be updated to maintain hygiene.

---

## 3. Snyk — Comprehensive Vulnerability Scanner

**Status: Pending container execution**

Snyk provides dependency scanning, SAST code analysis, container image scanning, and IaC configuration scanning.

### How to Run

```bash
# Set Snyk token (free tier available at https://snyk.io)
export SNYK_TOKEN=your-token-here

# Run via docker-compose.security.yml
podman-compose -f docker-compose.security.yml --profile snyk-scan up snyk-cli

# Reports output to: reports/
#   - snyk-dependencies-results.json
#   - snyk-code-results.json
#   - snyk-container-results.json
#   - snyk-iac-results.json
```

### Previous Results (2026-02-10)

From archived JSON reports in `docs/security/`:
- `snyk-go-20260210_172319.json` — Go dependencies scanned
- `snyk-catalog-web-20260210_172319.json` — catalog-web scanned
- `snyk-catalogizer-desktop-20260210_172319.json` — Desktop scanned
- `snyk-catalogizer-api-client-20260210_172319.json` — API client scanned
- `snyk-installer-wizard-20260210_172319.json` — Wizard scanned

---

## 4. Trivy — Filesystem and Container Scanner

**Status: Pending container execution**

Trivy scans for vulnerabilities, secrets, and misconfigurations across the entire filesystem.

### How to Run

```bash
# Run via docker-compose.security.yml
podman-compose -f docker-compose.security.yml --profile trivy-scan up trivy-scanner

# Or run directly
podman run --rm \
  -v /run/media/milosvasic/DATA4TB/Projects/Catalogizer:/project:ro \
  docker.io/aquasec/trivy:latest fs \
  --scanners vuln,secret,config \
  --severity HIGH,CRITICAL \
  --format json \
  --output /dev/stdout \
  /project

# Reports output to: reports/trivy-results.json
```

---

## 5. SonarQube — Static Analysis and Code Quality

**Status: Pending container execution**

SonarQube provides deep static analysis, code smells, security hotspots, and technical debt tracking.

### How to Run

```bash
# Start SonarQube server
podman-compose -f docker-compose.security.yml up -d sonarqube sonarqube-db

# Wait for SonarQube to be ready (usually ~60s)
# Access UI at http://localhost:9000 (default: admin/admin)

# Run scanner (requires sonar-scanner or scripts/sonarqube-scan.sh)
./scripts/sonarqube-scan.sh

# Stop SonarQube
podman-compose -f docker-compose.security.yml down
```

---

## 6. Semgrep — SAST Scanner

**Status: Pending container execution**

Semgrep performs pattern-based static analysis across Go, TypeScript, and Kotlin code.

### How to Run

```bash
# Run via docker-compose.security.yml
podman-compose -f docker-compose.security.yml --profile semgrep-scan up semgrep-scanner

# Or run directly
podman run --rm \
  -v /run/media/milosvasic/DATA4TB/Projects/Catalogizer:/project:ro \
  docker.io/semgrep/semgrep:latest semgrep scan \
  --config auto \
  --json \
  --exclude node_modules \
  --exclude vendor \
  --exclude dist \
  --exclude build \
  --exclude target \
  --exclude releases \
  --exclude .git \
  --severity WARNING \
  /project

# Reports output to: reports/semgrep-results.json
```

---

## 7. OWASP Dependency Check

**Status: Pending container execution**

OWASP Dependency Check identifies known vulnerable components using the NVD database.

### How to Run

```bash
# Run via docker-compose.security.yml
podman-compose -f docker-compose.security.yml --profile dependency-check up dependency-check

# Reports output to: reports/dependency-check/ (HTML, XML, JSON)
```

---

## 8. gosec — Go Static Security Analysis

**Status: Not run (available via `go install`)**

### Previous Results (2026-03-04)

From the March 4 security scan report:
- 525 total findings: 29 HIGH, 166 MEDIUM, 324 LOW (454 production, 71 test-only)
- All HIGH findings addressed (path traversal fixed, integer overflow fixed, remainder are false positives)
- All MEDIUM/LOW findings reviewed and classified as acceptable patterns

### How to Run

```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run
cd catalog-api && gosec -fmt json -out gosec-results.json -exclude-generated ./...
```

---

## Security Infrastructure Summary

### Available Container Scanners (docker-compose.security.yml)

| Service | Image | Profile | Resource Limits |
|---------|-------|---------|----------------|
| SonarQube | `sonarqube:community` | (default) | 2 CPU / 2 GB |
| SonarQube DB | `postgres:15-alpine` | (default) | 1 CPU / 1 GB |
| Snyk CLI | `node:18-alpine` | `snyk-scan` | 2 CPU / 2 GB |
| Trivy | `aquasec/trivy:latest` | `trivy-scan` | 1 CPU / 1 GB |
| Semgrep | `semgrep/semgrep:latest` | `semgrep-scan` | 2 CPU / 2 GB |
| OWASP Dep Check | `owasp/dependency-check:latest` | `dependency-check` | 2 CPU / 4 GB |

### Local Tools (no containers needed)

| Tool | Command | Status |
|------|---------|--------|
| govulncheck | `cd catalog-api && govulncheck ./...` | Installed, v1.1.4 |
| npm audit | `cd catalog-web && npm audit` | Installed, npm 10.9.3 |
| gosec | `go install github.com/securego/gosec/v2/cmd/gosec@latest` | Available |

### Automated Scanning Script

```bash
# Run all available local scanners
./scripts/security-scan.sh

# Output: docs/security/security-scan-YYYYMMDD_HHMMSS.md + JSON reports
```

---

## Recommendations

### Immediate Actions

1. Run `cd catalog-web && npm audit fix` to resolve 6 dev dependency vulnerabilities
2. Update `golang.org/x/image` to v0.38.0 in `catalog-api/go.mod` (precautionary; code does not call vulnerable path)

### Scheduled Container Scans

Run the following container-based scans when container infrastructure is available:

```bash
# Full security scan suite
podman-compose -f docker-compose.security.yml --profile snyk-scan up snyk-cli
podman-compose -f docker-compose.security.yml --profile trivy-scan up trivy-scanner
podman-compose -f docker-compose.security.yml --profile semgrep-scan up semgrep-scanner
podman-compose -f docker-compose.security.yml --profile dependency-check up dependency-check
```

### Continuous Security

1. Run `govulncheck` and `npm audit` before each release
2. Schedule weekly container-based Snyk/Trivy scans
3. Run SonarQube analysis quarterly for code quality and security hotspots
4. Re-run gosec after any security-sensitive code changes
5. Monitor Go vulnerability database updates at https://vuln.go.dev

---

## Comparison with Previous Report (2026-03-04)

| Area | 2026-03-04 | 2026-03-26 | Change |
|------|-----------|-----------|--------|
| govulncheck (symbol) | 0 vulns | 0 vulns | No change |
| govulncheck (module) | 0 vulns | 1 vuln (x/image, not called) | New advisory published |
| npm audit (production) | 0 vulns (all projects) | 0 vulns (all projects) | No change |
| npm audit (dev, catalog-web) | 0 vulns | 6 vulns (5 high, 1 moderate) | New advisories published |
| gosec | 525 findings (all addressed) | Not re-run | Pending |

---

**End of Report**

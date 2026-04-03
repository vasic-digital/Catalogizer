# Security Scan Consolidated Report

**Date:** 2026-04-03
**Version:** v2.2.0
**Previous Scan:** 2026-02-10

---

## Executive Summary

| Tool | Critical | High | Medium | Low | Status |
|------|----------|------|--------|-----|--------|
| govulncheck (Go) | 0 | 0 | 0 | 0 | PASS |
| npm audit (production) | 0 | 0 | 0 | 0 | PASS |
| SonarQube | - | - | - | - | Pending (requires container) |
| Snyk | - | - | - | - | Pending (requires container) |
| Semgrep | - | - | - | - | Pending (requires container) |
| OWASP Dep-Check | - | - | - | - | Pending (requires container) |
| Trivy | - | - | - | - | Pending (requires container) |

**Overall Status:** PASS for all locally-runnable scans. Container-based scans require `podman-compose -f docker-compose.security.yml` execution.

---

## Scan Results

### 1. govulncheck (Go Dependencies)

**Command:** `cd catalog-api && govulncheck ./...`
**Result:** No vulnerabilities found.
**Coverage:** All Go standard library and third-party dependencies.

### 2. npm audit (Frontend Dependencies)

**Command:** `cd catalog-web && npm audit --omit=dev`
**Result:** 0 vulnerabilities (after remediation).

**Remediation Applied:**
- `lodash` upgraded from 4.17.23 to 4.18.1
  - Fixed: GHSA-r5fr-rjxr-66jc (Code Injection via `_.template`, HIGH)
  - Fixed: GHSA-f23m-r3pf-42rh (Prototype Pollution via `_.unset`/`_.omit`, HIGH)

### 3. SonarQube (Static Analysis)

**Status:** Infrastructure ready via `docker-compose.security.yml`
**Configuration:** `sonar-project.properties` created in project root
**Execution command:**
```bash
./scripts/run-sonarqube-scan.sh
```
**Note:** Requires SonarQube container to be running. First-time startup takes ~3-5 minutes.

### 4. Snyk (Dependency + Code + Container + IaC)

**Status:** Configured in `docker-compose.security.yml` (profile: snyk-scan)
**Execution command:**
```bash
podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-cli
```
**Note:** Requires SNYK_TOKEN environment variable for authenticated scans.

### 5. Semgrep (SAST)

**Status:** Configured in `docker-compose.security.yml` (profile: semgrep-scan)
**Execution command:**
```bash
podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner
```
**Coverage:** Auto-configured rules for Go, TypeScript, JavaScript.

### 6. OWASP Dependency Check

**Status:** Configured in `docker-compose.security.yml` (profile: dependency-check)
**Execution command:**
```bash
podman-compose -f docker-compose.security.yml --profile dependency-check run --rm dependency-check
```

### 7. Trivy (Filesystem + Secrets + Config)

**Status:** Configured in `docker-compose.security.yml` (profile: trivy-scan)
**Execution command:**
```bash
podman-compose -f docker-compose.security.yml --profile trivy-scan run --rm trivy
```

---

## Comparison with Previous Scan (2026-02-10)

| Metric | 2026-02-10 | 2026-04-03 | Delta |
|--------|-----------|-----------|-------|
| Go vulnerabilities | 0 | 0 | = |
| npm critical/high (prod) | 0 | 0 | = |
| lodash version | 4.17.23 | 4.18.1 | Upgraded |

---

## Security Hardening Applied in v2.2.0

1. **Lodash upgraded** — Fixed code injection and prototype pollution vulnerabilities
2. **Dynamic container runtime** — scripts/lib/container-runtime.sh detects Podman > Docker > nerdctl
3. **SonarQube project configuration** — sonar-project.properties added for automated scanning
4. **Rate limiting** — Strict (5/min) on auth endpoints, default (100/min) on others
5. **JWT validation** — Token expiration enforced, refresh mechanism active
6. **CORS configuration** — Disabled by default in production, configurable via env vars
7. **HTTP/3 (QUIC)** — TLS with self-signed certs for dev, proper certs for production
8. **Brotli compression** — All API responses compressed

---

## Recommendations

1. **Run container-based scans** when infrastructure is available to complete the audit
2. **Schedule regular scans** — monthly cadence recommended
3. **Monitor lodash usage** — consider replacing with lodash-es for tree-shaking
4. **Keep dependencies updated** — run `govulncheck` and `npm audit` before each release

---

## How to Run All Scans

```bash
# Local scans (no containers needed)
cd catalog-api && govulncheck ./...
cd catalog-web && npm audit --omit=dev

# Container-based scans (requires Podman)
# Start SonarQube infrastructure
./scripts/run-sonarqube-scan.sh

# Run individual scanning profiles
podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-cli
podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner
podman-compose -f docker-compose.security.yml --profile dependency-check run --rm dependency-check
podman-compose -f docker-compose.security.yml --profile trivy-scan run --rm trivy
```

# Module 16: Security Scanning -- Video Script

**Duration**: 40 minutes
**Prerequisites**: Module 11 (Security and Monitoring), familiarity with vulnerability databases (CVE, NVD)

---

## Video 16.1: Security Scanning Overview (10 min)

### Opening

Welcome to Module 16. Security is not a one-time activity -- it requires continuous scanning of both your own code and your dependencies. This module covers the six security scanning tools integrated into Catalogizer and how they fit together.

### The Security Scanning Stack

Catalogizer uses a layered approach to security scanning:

| Tool | Target | Scope |
|------|--------|-------|
| govulncheck | Go dependencies | Known CVEs in imported packages |
| npm audit | Node.js dependencies | Known vulnerabilities in npm packages |
| Semgrep | Source code | Static analysis patterns (injection, XSS, etc.) |
| SonarQube | Source code | Code quality, security hotspots, complexity |
| Snyk | Dependencies + containers | Dependency and container image vulnerabilities |
| Trivy | Container images | OS package and application dependency vulnerabilities |

Each tool catches a different category of vulnerability. Using all six provides defense in depth.

### Running All Scans

The project includes scripts for automated scanning:

```bash
# Run all security tests
./scripts/run-all-tests.sh

# Run security-specific tests
./scripts/security-test.sh

# Run individual scans
./scripts/snyk-scan.sh
./scripts/sonarqube-scan.sh
```

The Docker Compose security profile provides containerized scanning:

```bash
podman-compose -f docker-compose.security.yml up
```

---

## Video 16.2: govulncheck for Go (10 min)

### What govulncheck Does

`govulncheck` is Go's official vulnerability scanner. Unlike generic dependency scanners, it performs call graph analysis -- it only reports vulnerabilities in functions your code actually calls, not just functions that exist in imported packages.

### Running govulncheck

```bash
cd catalog-api

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan all packages
govulncheck ./...
```

### Interpreting Results

```
Vulnerability #1: GO-2024-XXXX
    Description: Buffer overflow in example/package
    Found in: example/package@v1.2.3
    Fixed in: example/package@v1.2.4
    Call stacks:
        main.go:45 -> services.NewScanner -> example/package.Parse
```

The output shows:
- The vulnerability ID (linked to the Go vulnerability database)
- The affected package and version
- The fixed version you need to upgrade to
- The exact call stack from your code to the vulnerable function

### Fixing Vulnerabilities

```bash
# Update the vulnerable dependency
go get example/package@v1.2.4

# Verify the fix
govulncheck ./...

# Run tests to confirm no regressions
GOMAXPROCS=3 go test ./... -p 2 -parallel 2
```

### Current Status

Catalogizer maintains zero known vulnerabilities as verified by govulncheck. This is checked before every release.

---

## Video 16.3: Semgrep Static Analysis (10 min)

### What Semgrep Does

Semgrep is a static analysis tool that matches code patterns. It catches security issues that dependency scanners miss because they are in your own code, not in libraries.

### Key Rules for Catalogizer

Semgrep checks for:
- **SQL injection**: String concatenation in SQL queries (Catalogizer uses parameterized queries via the `database.DB` wrapper)
- **Path traversal**: Unsanitized file paths in handlers
- **Hardcoded secrets**: API keys, passwords, JWT secrets in source code
- **Insecure cryptography**: Weak hash algorithms, insecure random number generation
- **XSS vulnerabilities**: Unescaped user input in responses

### Running Semgrep

```bash
# Install Semgrep
pip install semgrep

# Run with auto-configuration
semgrep --config auto catalog-api/

# Run with specific rulesets
semgrep --config p/golang catalog-api/
semgrep --config p/javascript catalog-web/
semgrep --config p/typescript catalog-web/

# Run with OWASP rules
semgrep --config p/owasp-top-ten catalog-api/
```

### Interpreting Results

Semgrep categorizes findings by severity:

- **ERROR**: Must fix before deployment (e.g., SQL injection)
- **WARNING**: Should fix (e.g., missing error check)
- **INFO**: Best practice suggestion (e.g., prefer constants over magic numbers)

### How Catalogizer Prevents SQL Injection

The dialect abstraction layer is the primary defense:

```go
// The DB wrapper rewrites all queries
// Application code writes:
db.QueryRow("SELECT * FROM users WHERE id = ?", userID)

// For PostgreSQL, this becomes:
// SELECT * FROM users WHERE id = $1
```

All queries use `?` placeholders that are rewritten by `RewritePlaceholders()`. No raw string concatenation is used in SQL construction. Semgrep rules verify this pattern is maintained.

---

## Video 16.4: npm audit, Snyk, and Trivy (10 min)

### npm audit for Frontend Dependencies

```bash
cd catalog-web

# Run the audit
npm audit

# Fix automatically where possible
npm audit fix

# Check for production-only vulnerabilities
npm audit --omit=dev
```

Catalogizer maintains zero critical and zero production vulnerabilities in npm dependencies.

### Snyk for Comprehensive Scanning

Snyk scans both dependencies and container images:

```bash
# Authenticate
snyk auth

# Scan Go dependencies
cd catalog-api && snyk test

# Scan Node.js dependencies
cd catalog-web && snyk test

# Scan a container image
snyk container test catalogizer-api:latest

# Monitor for new vulnerabilities
snyk monitor
```

### Trivy for Container Security

Trivy scans container images for OS package vulnerabilities:

```bash
# Scan the API container image
trivy image catalogizer-api:latest

# Scan with severity filter
trivy image --severity HIGH,CRITICAL catalogizer-api:latest

# Scan a filesystem
trivy fs catalog-api/

# Generate JSON report
trivy image --format json --output trivy-report.json catalogizer-api:latest
```

### SonarQube for Code Quality

SonarQube provides deeper static analysis with security hotspot detection:

```bash
# Start SonarQube (containerized)
podman-compose -f docker-compose.security.yml up sonarqube

# Run the scanner
./scripts/sonarqube-scan.sh
```

SonarQube identifies:
- Security hotspots (code that needs security review)
- Code smells (maintainability issues)
- Bugs (likely runtime errors)
- Technical debt estimates

### Integrating Scans into the Workflow

Since GitHub Actions are permanently disabled for this project, scans run locally:

```bash
# Pre-release security check
cd catalog-api && govulncheck ./...
cd catalog-web && npm audit --omit=dev
semgrep --config auto catalog-api/ catalog-web/
trivy image catalogizer-api:latest
```

This sequence is part of the release build pipeline and runs before any deployment.

---

## Exercises

1. Run `govulncheck ./...` in the `catalog-api` directory and interpret the output
2. Write a custom Semgrep rule that detects any use of `fmt.Sprintf` to construct SQL queries
3. Run `npm audit` on `catalog-web` and investigate any findings
4. Build the API container image and scan it with Trivy

---

## Key Files Referenced

- `scripts/security-test.sh` -- Security test runner
- `scripts/snyk-scan.sh` -- Snyk scanning script
- `scripts/sonarqube-scan.sh` -- SonarQube scanning script
- `docker-compose.security.yml` -- Security scanning container stack
- `catalog-api/database/dialect.go` -- SQL injection prevention via parameterized queries
- `catalog-api/middleware/input_validation.go` -- Input validation middleware
- `catalog-api/middleware/security_headers.go` -- Security headers middleware

---

## Addendum: Phase 1 Scanning Methodology and CVE Fixes

**[Visual: Timeline showing Phase 1 security audit: scan -> triage -> fix -> verify -> document]**

**Narrator**: The initial security scanning effort -- Phase 1 -- established the baseline methodology and resolved the first wave of findings. This addendum documents the process, the specific CVE fixes applied, and the ongoing scanning discipline.

### Phase 1 Scanning Methodology

**[Visual: Flowchart: Inventory -> Scan -> Triage -> Prioritize -> Fix -> Verify -> Report]**

**Narrator**: The Phase 1 approach follows a structured seven-step methodology:

1. **Inventory**: Enumerate all dependencies across Go modules (`go.sum`), Node.js packages (`package-lock.json`), container base images, and system libraries.

2. **Scan**: Run all six tools in sequence:
   ```bash
   # Go dependencies
   cd catalog-api && govulncheck ./...

   # Node.js dependencies
   cd catalog-web && npm audit --omit=dev
   cd installer-wizard && npm audit --omit=dev

   # Static analysis
   semgrep --config auto --config p/owasp-top-ten catalog-api/ catalog-web/

   # Container images
   trivy image --severity HIGH,CRITICAL catalogizer-api:latest
   trivy image --severity HIGH,CRITICAL catalogizer-web:latest

   # Code quality and security hotspots
   ./scripts/run-sonarqube-scan.sh
   ```

3. **Triage**: Classify each finding by severity (critical, high, medium, low), exploitability (is the vulnerable code path reachable?), and impact (data exposure, denial of service, remote code execution).

4. **Prioritize**: Fix critical and high findings immediately. Medium findings are scheduled for the next sprint. Low findings are documented and tracked.

5. **Fix**: Apply the fix -- dependency upgrade, code change, or configuration adjustment.

6. **Verify**: Re-run the specific scanner that found the issue to confirm the fix.

7. **Report**: Document the finding, fix, and verification in `docs/security/`.

### CVE Fixes Applied in Phase 1

**[Visual: Table of resolved CVEs with severity badges]**

The following categories of vulnerabilities were identified and resolved during Phase 1:

#### Go Dependency Upgrades

| Category | Action | Impact |
|----------|--------|--------|
| HTTP/2 rapid reset (CVE class) | Upgraded `golang.org/x/net` | Prevented DoS via HTTP/2 stream reset floods |
| QUIC handshake vulnerability | Upgraded `quic-go` to latest | Fixed TLS handshake edge case in HTTP/3 |
| Crypto timing side-channel | Upgraded `golang.org/x/crypto` | Eliminated timing leak in bcrypt comparison |
| YAML parsing DoS | Upgraded `gopkg.in/yaml.v3` | Fixed billion-laughs-style entity expansion |

```bash
# Typical fix workflow for Go dependencies
cd catalog-api
go get golang.org/x/net@latest
go get golang.org/x/crypto@latest
go mod tidy
govulncheck ./...     # Verify: 0 vulnerabilities
GOMAXPROCS=3 go test ./... -p 2 -parallel 2  # Verify: 0 failures
```

#### Node.js Dependency Upgrades

| Category | Action | Impact |
|----------|--------|--------|
| Prototype pollution in dev deps | Upgraded transitive dependencies | Eliminated prototype pollution vectors |
| ReDoS in validation libraries | Upgraded `zod`, `semver` | Fixed regex denial-of-service patterns |
| Path traversal in build tools | Upgraded `vite`, `esbuild` | Prevented directory escape in dev server |

```bash
# Typical fix workflow for Node.js dependencies
cd catalog-web
npm audit fix
npm audit --omit=dev  # Verify: 0 vulnerabilities in production deps
npm run test          # Verify: 1623 tests pass
npm run build         # Verify: clean build
```

#### Static Analysis Fixes (Semgrep)

| Finding | Severity | Fix |
|---------|----------|-----|
| Hardcoded JWT secret in test file | Medium | Moved to environment variable, added `.env.test` |
| Missing error check on `db.Close()` | Low | Added deferred error check with logging |
| Potential path traversal in file handler | High | Added `filepath.Clean()` and base directory validation |
| Insecure TLS minimum version | Medium | Set `tls.Config.MinVersion = tls.VersionTLS13` |

### Safety Improvements from Phase 1

**[Visual: Before/after comparison of safety patterns]**

Beyond CVE fixes, Phase 1 established several defensive patterns that prevent future vulnerabilities:

**Database query timeout (30s default):** Every database query now has a 30-second context timeout applied at the `database.DB` wrapper level. This prevents a single slow query from exhausting the connection pool.

**Redis middleware fail-open (500ms timeout):** The Redis rate limiter uses a 500ms timeout. If Redis is slow or down, requests pass through rather than blocking. This is a deliberate availability-over-strictness trade-off -- rate limiting is a best-effort defense, not a hard security boundary.

**Goroutine lifecycle cleanup:** All background goroutines (`CacheService`, `WebSocketHandler`, `SyncService`) now have explicit shutdown paths with `sync.Once`-guarded `Close()`/`Stop()` methods. This prevents goroutine leaks that could lead to memory exhaustion (a denial-of-service vector).

**Shared-pointer race elimination:** `SyncService.StartSync()` and `LogManagementService.CollectLogs()` now return deep copies instead of pointers to internal state, eliminating data races that could cause inconsistent security decisions.

### Ongoing Scanning Discipline

**[Visual: Checklist showing pre-release security gate]**

**Narrator**: Phase 1 established the scanning cadence that all subsequent releases must follow:

```bash
# Pre-release security gate (mandatory before any deployment)
cd catalog-api && govulncheck ./...                          # 0 findings
cd catalog-web && npm audit --omit=dev                       # 0 critical/high
semgrep --config auto --severity ERROR catalog-api/          # 0 errors
trivy image --severity CRITICAL catalogizer-api:latest       # 0 critical

# Periodic deep scan (weekly)
./scripts/run-sonarqube-scan.sh                              # Review hotspots
podman-compose -f docker-compose.security.yml \
  --profile semgrep-scan run --rm semgrep-scanner            # Full ruleset
```

**Narrator**: The pre-release gate runs four fast scans that complete in under 2 minutes. The weekly deep scan runs SonarQube and the full Semgrep ruleset, which takes 10-15 minutes but catches subtler issues like code complexity, maintainability risks, and security hotspots that require human review.

**Key takeaways:**
- Phase 1 resolved all critical and high CVEs across Go, Node.js, and container images.
- Safety patterns (query timeout, Redis fail-open, goroutine lifecycle) prevent new vulnerability classes.
- The pre-release security gate is a mandatory 2-minute check before any deployment.
- Weekly deep scans with SonarQube and Semgrep catch issues that fast scans miss.

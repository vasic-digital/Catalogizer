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

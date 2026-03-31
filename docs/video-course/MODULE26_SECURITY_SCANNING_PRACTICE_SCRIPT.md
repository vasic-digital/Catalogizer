# Module 26: Security Scanning in Practice

## Video Script — Running & Interpreting Security Scans

### Duration: ~20 minutes

---

### Scene 1: Introduction (2 min)

"Security scanning is not optional — it's a continuous quality gate. In this module, we'll run every scanner configured in Catalogizer, interpret the results, and understand what to fix vs. what to accept."

**Tools covered:** govulncheck, npm audit, Semgrep (built-in + custom rules), SonarQube, Snyk

---

### Scene 2: govulncheck — Go Vulnerability Scanning (3 min)

```bash
cd catalog-api && govulncheck ./...
```

"govulncheck checks your Go dependencies and stdlib usage against the Go vulnerability database. It only reports vulnerabilities in code paths you actually call."

**Interpreting results:**
- "No vulnerabilities found" = CLEAN
- If findings: check if the vulnerable function is in your call path
- Fix: update dependency (`go get -u module@latest`)

---

### Scene 3: npm audit — Frontend Vulnerabilities (3 min)

```bash
cd catalog-web && npm audit --omit=dev
```

"Focus on production dependencies only (`--omit=dev`). Dev-only vulnerabilities don't reach users."

**Severity levels:** critical > high > moderate > low
**Fix:** `npm audit fix` or manual dependency update

---

### Scene 4: Semgrep with Custom Rules (5 min)

```bash
# Built-in rules
semgrep --config auto catalog-api/

# Custom project rules
semgrep --config config/semgrep-rules.yml catalog-api/
```

**Our 8 custom rules:**
1. `no-sql-string-concat` (ERROR) — prevents SQL injection
2. `no-hardcoded-credentials` (WARNING) — detects hardcoded secrets
3. `no-os-exec-user-input` (WARNING) — flags command injection risk
4. `missing-rows-close` (ERROR) — catches unclosed database rows
5. `no-default-http-client` (WARNING) — enforces pooled client
6. `no-fmt-errorf-without-wrap` (INFO) — encourages error chain preservation
7. `react-missing-key-prop` (WARNING) — catches missing React keys
8. `no-any-type` (WARNING) — discourages TypeScript any

**Understanding false positives:**
- `missing-rows-close` may flag QueryContext patterns — verify manually
- `no-hardcoded-credentials` may flag field type constants like `FieldTypePassword = "password"`

---

### Scene 5: SonarQube Code Quality (4 min)

```bash
# Start SonarQube
podman-compose -f docker-compose.security.yml up -d sonarqube sonarqube-db

# Run scan
./scripts/run-sonarqube-scan.sh
```

**Quality gate checks:** bugs, vulnerabilities, code smells, coverage, duplication
**Dashboard:** `http://localhost:9000` — project overview, hotspots, issues

---

### Scene 6: Snyk Vulnerability Scanning (3 min)

```bash
podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-scanner
```

"Snyk scans Go modules, npm packages, Dockerfiles, and IaC configurations."

**Freemium usage:** Works without token for basic scanning. Token enables monitoring and CI integration.

---

### Summary

| Scanner | What It Checks | Command |
|---------|---------------|---------|
| govulncheck | Go stdlib + deps | `govulncheck ./...` |
| npm audit | Node.js deps | `npm audit --omit=dev` |
| Semgrep | Code patterns | `semgrep --config config/semgrep-rules.yml` |
| SonarQube | Quality gate | `./scripts/run-sonarqube-scan.sh` |
| Snyk | Multi-language | `snyk test --all-projects` |

**Rule:** Run all scanners before every release. Zero critical findings.

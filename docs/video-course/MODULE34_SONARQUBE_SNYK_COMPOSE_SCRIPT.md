# Module 34 — Wiring SonarQube + Snyk via Compose

**Duration:** 16 minutes
**Prerequisites:** Module 16 (Security Scanning), Module 26 (Security Scanning Practice)

## Learning objectives

1. Run SonarQube (server + database) inside rootless Podman.
2. Run Snyk against Go + npm + container images without embedding tokens in source.
3. Orchestrate `govulncheck`, `npm audit`, Semgrep, Trivy, SonarQube, and Snyk from a single script.
4. Aggregate results into a single consolidated report.

## Segment 1 — Why these tools (0:00 – 3:00)

- **SonarQube** — long-term quality tracking: code smells, bugs, coverage trend, technical debt.
- **Snyk** — SCA (software composition analysis) across Go dependencies, npm, Dockerfile/container images. Cloud-assisted (needs a token).
- **govulncheck** — Go stdlib + module CVE check. Local, no token.
- **npm audit** — npm registry CVE check. Local, no token.
- **Semgrep** — static analysis rules (OWASP, security anti-patterns). Local via container.
- **Trivy** — container image CVE + misconfig scanner. Local via container.

No single tool covers everything; run them all and aggregate.

## Segment 2 — SonarQube in compose (3:00 – 7:00)

**Show on screen:** `docker-compose.security.yml::sonarqube` + `sonarqube-db`.

Key configuration:
- `sonar.forceAuthentication: false` — dev convenience, do NOT use in production.
- `SONAR_ES_BOOTSTRAP_CHECKS_DISABLE: true` — skips Elasticsearch vm.max_map_count check (required for rootless containers).
- JVM memory caps: 1 GB web, 1 GB compute engine (stays under the 8 GB project-wide budget).
- PostgreSQL 15 as the backing DB, isolated on `security-testing-network`.

Launch:
```bash
podman-compose -f docker-compose.security.yml up -d sonarqube sonarqube-db
# Wait for http://localhost:9000/api/system/status == UP
```

## Segment 3 — Snyk with token injection (7:00 – 10:00)

Store `SNYK_TOKEN` in `.env` at the repo root (gitignored):
```
SNYK_TOKEN=your-real-token-here
```

Run via compose — token passed via environment, never written to disk inside the container:
```bash
podman-compose -f docker-compose.security.yml run --rm \
  -e SNYK_TOKEN snyk-cli snyk test --json --all-projects
```

**Never commit the token.** Pre-commit hook should scan for `SNYK_TOKEN=` patterns.

## Segment 4 — The `security-scan-all.sh` orchestrator (10:00 – 14:00)

**Show on screen:** `scripts/security-scan-all.sh`.

Script runs each scanner in sequence, writes per-scanner output to `docs/reports/security/<date>/`, and produces a single `CONSOLIDATED.md`. Modes:

```bash
./scripts/security-scan-all.sh --all             # every scanner
./scripts/security-scan-all.sh --govulncheck-only
./scripts/security-scan-all.sh --snyk-only
./scripts/security-scan-all.sh --sonarqube-only
```

Per-scanner error handling:
- Missing tool (e.g., govulncheck not installed) → SKIP with install hint.
- Missing token (SNYK_TOKEN / SONAR_TOKEN) → SKIP with hint.
- Scanner failure → FAIL, script exits non-zero.
- Scanner success → OK, output written to the per-date directory.

## Segment 5 — Integration with CI (14:00 – 16:00)

The project's policy forbids GitHub Actions (see CLAUDE.md). Integration is manual:
- Pre-release: run `./scripts/security-scan-all.sh --all` and commit the consolidated report.
- Pre-commit hook: run `govulncheck` and `gosec` on changed files only (fast subset).
- Per-PR reviewer: review the consolidated report as part of the PR checklist.

## Exercise

1. Export `SNYK_TOKEN` from your `.env` (`set -a; source .env; set +a`).
2. Run `./scripts/security-scan-all.sh --all`.
3. Open the resulting `CONSOLIDATED.md` and fix any High/Critical finding reported.

## Assessment

1. Why is `--network host` needed for some scanners? Answer: they proxy to the compose service network; host mode lets them reach `localhost:9000` / `localhost:8080` transparently.
2. What should you do if Snyk reports a false positive? Answer: add a `.snyk` ignore entry with the CVE ID, severity, and expiry date.

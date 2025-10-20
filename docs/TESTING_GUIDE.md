# Testing Guide

## Overview

Catalogizer employs a comprehensive testing strategy that includes unit tests, integration tests, security testing, and quality assurance. This guide covers all testing aspects including the mandatory security scanning requirements.

## Test Categories

### 1. Unit Tests
- **Go Backend**: `go test ./...` - Tests individual functions and methods
- **React Frontend**: `npm test` - Jest-based component and utility tests
- **Android Apps**: `./gradlew test` - Unit tests for Kotlin code
- **Coverage**: Minimum 80% coverage required for all modules

### 2. Integration Tests
- **API Integration**: End-to-end API testing with real database
- **Cross-platform**: Tests between different client applications
- **Database**: Tests data persistence and migration
- **File System**: Tests multi-protocol file operations

### 3. Security Testing (Mandatory)

#### SonarQube Code Quality Analysis
**Purpose**: Static code analysis for bugs, vulnerabilities, and code smells

**Setup**:
```bash
# Set environment variable
export SONAR_TOKEN=your_sonar_token_here

# Run analysis
./scripts/sonarqube-scan.sh
```

**Requirements**:
- Quality gate must pass
- No critical or blocker issues
- Coverage minimum 80%
- Code smell density < 5%

**Reports**: `reports/sonarqube-report.json`

#### Snyk Security Scanning (Freemium)
**Purpose**: Dependency vulnerability scanning and Static Application Security Testing (SAST)

**Freemium Benefits**:
- Unlimited private repositories
- Unlimited developers
- 200 tests per month for public repos
- Basic vulnerability remediation guidance

**Setup**:
```bash
# 1. Sign up for free account at https://snyk.io
# 2. Get your token from https://snyk.io/account
# 3. Set environment variables
export SNYK_TOKEN=your_snyk_token_here
export SNYK_ORG=your_org_name  # Optional

# 4. Run scanning
./scripts/snyk-scan.sh
```

**Requirements**:
- No high or critical severity vulnerabilities
- Dependencies must be regularly updated
- Security policies must be enforced

**Reports**: `reports/snyk-*-results.json`

#### Additional Security Tools

**Trivy Vulnerability Scanner**:
```bash
# Scan filesystem
docker-compose -f docker-compose.security.yml run --rm trivy-scanner
```

**OWASP Dependency Check**:
```bash
# Check dependencies
docker-compose -f docker-compose.security.yml run --rm dependency-check
```

## Running All Tests

#### Initial Setup (Freemium Accounts)
```bash
# Setup your freemium security testing accounts
./scripts/setup-freemium-tokens.sh
```

#### Full Test Suite (Including Security)
```bash
# Run all tests including security scans
./scripts/security-test.sh

# Or run individual security scans
./scripts/sonarqube-scan.sh  # Requires SONAR_TOKEN
./scripts/snyk-scan.sh       # Requires SNYK_TOKEN
```

This script will:
1. Start security services (SonarQube, etc.)
2. Run all unit and integration tests
3. Perform security scans
4. Generate comprehensive reports
5. Stop security services

### Prerequisites
- Docker and Docker Compose installed
- Environment variables set for security tools
- All dependencies installed in project modules

## Test Reports

All test results are stored in the `reports/` directory:

- `comprehensive-security-report.html` - Main security report
- `sonarqube-report.json` - Code quality analysis
- `snyk-*-results.json` - Vulnerability scans per module
- `trivy-results.json` - Container vulnerability scan
- `dependency-check/` - OWASP dependency analysis

## Quality Gates

### Mandatory Requirements
- ✅ All unit tests pass (100% success rate)
- ✅ All integration tests pass
- ✅ SonarQube quality gate passes
- ✅ No high/critical Snyk vulnerabilities
- ✅ Minimum 80% test coverage
- ✅ No broken modules or features

### Zero-Defect Policy
Catalogizer follows a zero-defect policy where:
- All tests must pass before deployment
- Security scans must pass with no critical issues
- Code quality metrics must meet standards
- No module can be left broken or disabled

## CI/CD Integration

### GitHub Actions
The security testing is integrated into the QA pipeline:

```yaml
- name: 🔍 Run SonarQube Analysis
  if: env.SONAR_TOKEN != ''
  env:
    SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
  run: ./scripts/sonarqube-scan.sh

- name: 🔒 Run Snyk Security Analysis
  if: env.SNYK_TOKEN != ''
  env:
    SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
  run: ./scripts/snyk-scan.sh
```

### Required Secrets
- `SONAR_TOKEN`: SonarQube authentication token
- `SNYK_TOKEN`: Snyk API token
- `SNYK_ORG`: Snyk organization name (optional)

## Troubleshooting

### Common Issues

#### SonarQube Connection Failed
```bash
# Check if SonarQube is running
curl -f http://localhost:9000/api/system/status

# Restart services
docker-compose -f docker-compose.security.yml down
docker-compose -f docker-compose.security.yml up -d sonarqube
```

#### Snyk Authentication Failed
```bash
# Verify token
snyk auth $SNYK_TOKEN

# Check organization
snyk orgs
```

#### Docker Issues
```bash
# Clean up containers
docker system prune -a

# Restart Docker service
sudo systemctl restart docker
```

### Test Failures

#### Integration Tests Failing
- Check database connection
- Verify API server is running
- Check environment configuration

#### Security Scans Failing
- Update dependencies to latest secure versions
- Review and fix code quality issues
- Address security hotspots

## Development Testing

### Pre-commit Testing
```bash
# Run tests before committing
./scripts/run-tests.sh

# Run security checks
./scripts/security-test.sh
```

### Module-specific Testing

#### Go Backend
```bash
cd catalog-api
go test -v -race -cover ./...
```

#### React Frontend
```bash
cd catalog-web
npm test -- --coverage
```

#### Android Apps
```bash
cd catalogizer-android
./gradlew testDebugUnitTest
```

## Performance Testing

### Load Testing
```bash
# Using Artillery (if configured)
cd qa-ai-system/scripts/ci-cd
artillery run performance-test.yml
```

### Memory and CPU Profiling
```bash
# Go profiling
cd catalog-api
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
```

## Security Testing Best Practices

### Code Review Checklist
- [ ] Security-sensitive functions reviewed
- [ ] Input validation implemented
- [ ] Authentication/authorization checked
- [ ] SQL injection prevention verified
- [ ] XSS prevention implemented
- [ ] CSRF protection in place

### Dependency Management
- [ ] Dependencies regularly updated
- [ ] Security advisories monitored
- [ ] License compliance checked
- [ ] Unused dependencies removed

### Infrastructure Security
- [ ] Container images scanned
- [ ] Secrets properly managed
- [ ] Network security configured
- [ ] Access controls implemented

## Reporting Issues

### Test Failures
1. Check test logs in `reports/` directory
2. Identify root cause
3. Fix the issue
4. Re-run tests
5. Update documentation if needed

### Security Vulnerabilities
1. Assess severity and impact
2. Implement fix or mitigation
3. Update dependencies if applicable
4. Re-scan to verify resolution
5. Document the resolution

## Continuous Improvement

### Metrics Tracking
- Test coverage trends
- Security scan results over time
- Performance benchmarks
- Code quality metrics

### Regular Updates
- Security tools updated quarterly
- Dependencies reviewed monthly
- Test suites expanded with new features
- Documentation kept current

## Support

For testing-related issues:
- Check this guide first
- Review test logs in `reports/`
- Check GitHub Issues for similar problems
- Contact the development team

## Appendix

### Test Scripts Reference

| Script | Purpose | Requirements |
|--------|---------|--------------|
| `security-test.sh` | Full security testing | Docker, tokens |
| `sonarqube-scan.sh` | Code quality analysis | SONAR_TOKEN |
| `snyk-scan.sh` | Vulnerability scanning | SNYK_TOKEN |
| `run-tests.sh` | Unit and integration tests | None |

### Environment Variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `SONAR_TOKEN` | SonarQube authentication | Yes for SonarQube |
| `SNYK_TOKEN` | Snyk authentication | Yes for Snyk |
| `SNYK_ORG` | Snyk organization | Optional |
| `SONAR_HOST_URL` | SonarQube server URL | Optional (defaults to localhost) |

### File Structure

```
reports/
├── comprehensive-security-report.html
├── sonarqube-report.json
├── snyk-api-results.json
├── snyk-web-results.json
├── trivy-results.json
└── dependency-check/
    ├── dependency-check-report.html
    ├── dependency-check-report.json
    └── dependency-check-report.xml
```
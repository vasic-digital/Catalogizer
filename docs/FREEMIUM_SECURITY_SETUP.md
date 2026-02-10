# Freemium Security Testing Setup Guide

## Overview

Catalogizer uses **freemium versions** of industry-standard security tools to provide comprehensive security testing without licensing costs. All tools are either completely free/open source or offer generous free tiers suitable for most development teams.

## Freemium Tools Used

### 🔍 SonarQube Community Edition (Free)
- **Cost**: Completely free
- **Features**: Code quality analysis, bug detection, security hotspots
- **Limits**: None for private repositories
- **Setup**: https://sonarcloud.io (cloud-based, no server needed)

### 🔒 Snyk Free Tier (Freemium)
- **Cost**: Free for unlimited private repos
- **Features**: Dependency scanning, SAST, container scanning
- **Limits**: 200 tests/month for public repos (unlimited for private)
- **Setup**: https://snyk.io (free account)

### 🛡️ OWASP Dependency Check (Free/Open Source)
- **Cost**: Completely free
- **Features**: Third-party dependency vulnerability scanning
- **Limits**: None
- **Setup**: Docker-based, no account needed

### 🐳 Trivy (Free/Open Source)
- **Cost**: Completely free
- **Features**: Container and filesystem vulnerability scanning
- **Limits**: None
- **Setup**: Docker-based, no account needed

## Quick Setup (3 Steps)

### Step 1: Setup Accounts
```bash
# Interactive setup script
./scripts/setup-freemium-tokens.sh
```

### Step 2: Verify Setup
```bash
# Check if everything is configured
./scripts/verify-freemium-setup.sh
```

### Step 3: Run Security Tests
```bash
# Run all security tests
./scripts/security-test.sh
```

## Manual Setup

### SonarQube Setup
1. Go to https://sonarcloud.io
2. Sign up for free account
3. Create organization or use personal account
4. Go to Account → Security → Generate Token
5. Set environment variable:
   ```bash
   export SONAR_TOKEN=your_token_here
   ```

### Snyk Setup
1. Go to https://snyk.io
2. Sign up for free account
3. Verify email address
4. Go to Account → General → API Token
5. Set environment variables:
   ```bash
   export SNYK_TOKEN=your_token_here
   export SNYK_ORG=your_org_name  # Optional
   ```

## Environment Variables

Create a `.env.security` file or export these variables:

```bash
# SonarQube (Required for code quality)
SONAR_TOKEN=your_sonar_token

# Snyk (Required for vulnerability scanning)
SNYK_TOKEN=your_snyk_token
SNYK_ORG=catalogizer  # Optional
SNYK_SEVERITY_THRESHOLD=medium  # Optional

# Docker (Optional, for full testing)
COMPOSE_PROJECT_NAME=catalogizer-security
```

## Docker vs CLI Approach

### Docker Approach (Recommended)
- Uses containerized versions of all tools
- Consistent environment across machines
- Requires Docker and Docker Compose
- Full feature set available

```bash
# Run with Docker
docker-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-cli
```

### CLI Approach (Fallback)
- Uses native CLI tools
- Faster startup
- Less resource intensive
- May require local tool installation

```bash
# Run with CLI
./scripts/snyk-scan.sh
```

## Testing Modes

### Full Security Testing (Docker Required)
```bash
./scripts/security-test.sh
```
Includes: SonarQube, Snyk, Trivy, OWASP Dependency Check

### Individual Tool Testing
```bash
# SonarQube only
./scripts/sonarqube-scan.sh

# Snyk only
./scripts/snyk-scan.sh

# Docker-based tools
docker-compose -f docker-compose.security.yml run --rm trivy-scanner
docker-compose -f docker-compose.security.yml run --rm dependency-check
```

## Freemium Limitations & Workarounds

### Snyk Free Tier Limits
- **Public repos**: 200 tests/month
- **Private repos**: Unlimited
- **Workaround**: Use private repositories or upgrade to paid tier for unlimited public repos

### SonarQube Cloud Limits
- **Private projects**: Unlimited
- **Public projects**: Unlimited
- **Analysis time**: May have queue times during peak hours
- **Workaround**: Use during off-peak hours or self-host Community Edition

## Cost Analysis

| Tool | Free Tier | Paid Tier | Our Usage |
|------|-----------|-----------|-----------|
| SonarQube | Unlimited private | $15/user/month | ✅ Free |
| Snyk | Unlimited private | $3.02/user/month | ✅ Free |
| OWASP Dep Check | Free | N/A | ✅ Free |
| Trivy | Free | N/A | ✅ Free |

**Total Cost**: $0/month for comprehensive security testing

## Troubleshooting

### Token Issues
```bash
# Verify Snyk token
snyk whoami

# Test SonarQube token
curl -H "Authorization: Bearer $SONAR_TOKEN" https://sonarcloud.io/api/user_tokens/search
```

### Docker Issues
```bash
# Check Docker status
docker --version
docker-compose --version

# Clean up containers
docker system prune -a
```

### Network Issues
- Ensure outbound HTTPS access to:
  - sonarcloud.io
  - snyk.io
  - docker.io (for images)

## CI/CD Integration

### Local Security Scanning

> **Note:** GitHub Actions are permanently disabled for this project. Run security scans locally using the commands below.

```bash
# SonarQube analysis
export SONAR_TOKEN=your_sonar_token
./scripts/sonarqube-scan.sh

# Snyk security analysis
export SNYK_TOKEN=your_snyk_token
./scripts/snyk-scan.sh

# Full test suite including security
./scripts/run-all-tests.sh
```

### Required Environment Variables
- `SONAR_TOKEN`: From SonarCloud account
- `SNYK_TOKEN`: From Snyk account

## Security Reports

All reports are generated in the `reports/` directory:

- `sonarqube-report.json` - Code quality analysis
- `snyk-*-results.json` - Vulnerability scans
- `trivy-results.json` - Container scans
- `dependency-check/` - OWASP analysis
- `comprehensive-security-report.html` - Combined report

## Support

### Getting Help
- **SonarQube**: https://community.sonarsource.com
- **Snyk**: https://support.snyk.io
- **Documentation**: docs/TESTING_GUIDE.md

### Common Issues
- **Token expired**: Regenerate tokens from respective platforms
- **Rate limits**: Wait and retry, or upgrade to paid tier
- **Docker issues**: Ensure Docker Desktop is running

## Migration from Paid Tools

If you're migrating from paid security tools:

1. **SonarQube Server** → **SonarCloud**: Export quality profiles and rules
2. **Other SAST tools** → **Snyk Code**: Similar vulnerability detection
3. **Commercial scanners** → **Trivy + OWASP**: Comprehensive coverage

The freemium setup provides equivalent security coverage at zero cost!
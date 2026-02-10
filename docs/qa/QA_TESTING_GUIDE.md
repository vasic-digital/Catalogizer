# Catalogizer QA Testing Guide

![Test Status](https://img.shields.io/badge/tests-automated-brightgreen)
![CI/CD](https://img.shields.io/badge/CI/CD-enabled-blue)
![Coverage](https://img.shields.io/badge/coverage-tracking-yellow)

## 📋 Overview

Catalogizer has a comprehensive quality assurance system with real, automated tests that run both locally and in CI/CD pipelines. Our testing approach ensures code quality through multiple levels of validation.

## ✅ Testing Architecture

### Components

- **API Tests** - Go unit tests, integration tests, and benchmarks
- **Android Tests** - Kotlin unit tests and Gradle build validation
- **Database Tests** - Schema validation and migration testing
- **Integration Tests** - Cross-component workflow testing
- **Security Tests** - Static analysis with gosec and vulnerability scanning
- **Performance Tests** - Benchmarks and build performance metrics

### CI/CD Integration

> **Note:** GitHub Actions are permanently disabled for this project. All tests must be run locally.

Run the full test suite with:
```bash
./scripts/run-all-tests.sh
```

## 🚀 Quick Start Commands

### 1. Quick Validation (Fast Feedback)
```bash
./qa-ai-system/scripts/run-qa-tests.sh quick
```

**What it runs:**
- Pre-commit style validation
- Code formatting checks (Go, Android)
- Merge conflict detection
- Debug statement scanning
- Go vet and linting

**Duration:** 10-30 seconds

### 2. Standard Testing (Development)
```bash
./qa-ai-system/scripts/run-qa-tests.sh standard
```

**What it runs:**
- All quick checks
- Go API unit tests with coverage
- Android unit tests
- Database tests
- Integration tests
- Go build validation

**Duration:** 5-15 minutes

### 3. Complete Testing (Pre-Production)
```bash
./qa-ai-system/scripts/run-qa-tests.sh complete
```

**What it runs:**
- All standard tests
- Security scanning (gosec, vulnerability checks)
- Performance benchmarks
- Comprehensive code analysis

**Duration:** 15-30 minutes

## 🧪 Test Coverage

### Current Coverage Metrics

The project tracks test coverage for:

- **Go API:** `go test -coverprofile=coverage.out ./...`
- **Frontend (catalog-web):** 80% minimum threshold enforced
- **Database:** Migration and schema validation tests

View coverage reports:
```bash
# Go API coverage
cd catalog-api && go test -cover ./...
go tool cover -html=coverage.out

# Frontend coverage
cd catalog-web && npm run test:coverage
```

### Coverage Goals

| Component | Current | Target |
|-----------|---------|--------|
| Go API | Measured | 70%+ |
| Frontend | 80%+ | 80%+ |
| Integration | Growing | 60%+ |

## 🔧 Component-Specific Testing

Test individual components to save time during development:

```bash
# API only
./qa-ai-system/scripts/run-qa-tests.sh standard api

# Android only
./qa-ai-system/scripts/run-qa-tests.sh standard android

# Database only
./qa-ai-system/scripts/run-qa-tests.sh standard database

# Integration tests only
./qa-ai-system/scripts/run-qa-tests.sh standard integration

# Security scan only
./qa-ai-system/scripts/run-qa-tests.sh complete security

# Performance tests only
./qa-ai-system/scripts/run-qa-tests.sh complete performance
```

## 📊 Understanding Test Results

### Success Output
```
🎉 ALL QA TESTS PASSED!

✅ Test suites executed: 6
✅ Overall result: SUCCESS
✅ Quality level: standard validation completed
```

### Failure Output
```
❌ QA TESTS FAILED

📊 Test suites executed: 6
❌ Overall result: FAILED
🔍 Please review failed components above
```

### Test Logs

Each run generates a timestamped log file:
```
qa-tests-20241020-142305.log
```

Review logs for detailed error information and stack traces.

## 🛠️ Advanced Usage

### Dry Run Mode
Preview what would be tested without running:
```bash
./qa-ai-system/scripts/run-qa-tests.sh standard --dry-run
```

### Verbose Output
Get detailed execution information:
```bash
./qa-ai-system/scripts/run-qa-tests.sh standard --verbose
```

### Help
View all available options:
```bash
./qa-ai-system/scripts/run-qa-tests.sh --help
```

## 🔄 Recommended Development Workflow

### During Active Development
1. Make code changes
2. Run quick validation frequently:
   ```bash
   ./qa-ai-system/scripts/run-qa-tests.sh quick
   ```
3. Fix any formatting or linting issues immediately

### Before Committing
1. Run standard testing:
   ```bash
   ./qa-ai-system/scripts/run-qa-tests.sh standard
   ```
2. Ensure all tests pass
3. Review coverage reports if you added new code
4. Commit changes

### Before Pull Requests
1. Run complete testing:
   ```bash
   ./qa-ai-system/scripts/run-qa-tests.sh complete
   ```
2. Review all test results
3. Check security scan results
4. Verify performance benchmarks
5. Create pull request

### CI/CD Will Automatically
1. Run all tests on your PR
2. Report results as checks
3. Block merge if tests fail
4. Generate coverage reports

## 🏗️ Running Tests Manually

### Go API Tests
```bash
cd catalog-api

# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./internal/handlers/...

# Run benchmarks
go test -bench=. -benchmem ./...
```

### Android Tests
```bash
cd catalogizer-android

# Run unit tests
./gradlew testDebugUnitTest

# Build APK
./gradlew assembleDebug

# Run linting
./gradlew ktlintCheck
```

### Frontend Tests
```bash
cd catalog-web

# Run tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test:watch
```

## 🔐 Security Testing

Security tests run as part of the complete test suite and include:

### Static Analysis
- **gosec** - Go security scanner for common vulnerabilities
- Pattern matching for hardcoded secrets
- SQL injection vulnerability detection

### Dependency Scanning
- **govulncheck** - Check for known vulnerabilities in Go dependencies
- **Trivy** - Container and filesystem vulnerability scanner (in CI/CD)

### Manual Security Testing
```bash
# Install security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run gosec
cd catalog-api && gosec ./...

# Run govulncheck
cd catalog-api && govulncheck ./...
```

## ⚡ Performance Testing

Performance tests measure:
- Go benchmark execution time and memory allocation
- Build time (should be <30s for good, <60s for acceptable)
- Binary size
- API response times (in integration tests)

Run performance tests:
```bash
cd catalog-api

# Run all benchmarks
go test -bench=. -benchmem ./...

# Benchmark specific package
go test -bench=. -benchmem ./internal/services/...

# Compare benchmarks
go test -bench=. -benchmem ./... > old.txt
# Make changes
go test -bench=. -benchmem ./... > new.txt
benchstat old.txt new.txt
```

## 🐳 Docker-Based Testing

Run tests in a production-like environment using Docker:

```bash
# Start development environment with PostgreSQL and Redis
docker-compose -f docker-compose.dev.yml up -d

# Run tests against Docker services
DATABASE_URL="postgres://catalogizer:dev_password_change_me@localhost:5432/catalogizer_dev" \
REDIS_URL="redis://localhost:6379" \
go test ./...

# Clean up
docker-compose -f docker-compose.dev.yml down
```

See `../deployment/DOCKER_SETUP.md` for more details.

## 📈 Test Coverage Reporting

### Generate Coverage Reports

```bash
# Go API
cd catalog-api
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Frontend
cd catalog-web
npm run test:coverage
```

### Coverage Badges

Add coverage badges to README (requires CI/CD integration):

```markdown
![API Coverage](https://img.shields.io/badge/API%20Coverage-75%25-yellow)
![Frontend Coverage](https://img.shields.io/badge/Frontend%20Coverage-80%25-green)
```

## 🚨 Troubleshooting

### Tests Fail Locally

1. Check Go version: `go version` (should be 1.21+)
2. Check dependencies: `go mod tidy`
3. Clear test cache: `go clean -testcache`
4. Check environment variables

### Database Tests Fail

1. Ensure SQLite is installed: `sqlite3 --version`
2. Check database file permissions
3. For Postgre SQL tests, ensure Docker is running
4. Verify DATABASE_URL environment variable

### Android Tests Fail

1. Check Java version: `java -version` (should be 17)
2. Ensure ANDROID_HOME is set
3. Clear Gradle cache: `./gradlew clean`
4. Re-sync Gradle: `./gradlew --refresh-dependencies`

### Permission Issues

```bash
chmod +x qa-ai-system/scripts/run-qa-tests.sh
```

## 💡 Best Practices

1. **Run quick tests frequently** during development
2. **Always run standard tests** before committing
3. **Review coverage reports** when adding new features
4. **Fix security issues immediately** - don't ignore warnings
5. **Keep tests fast** - slow tests won't be run
6. **Write tests for new code** - maintain or improve coverage
7. **Use component-specific tests** when working on isolated features

## 📚 Additional Resources

- **Database Migrations:** `catalog-api/database/migrations/README.md`
- **Docker Setup:** `../deployment/DOCKER_SETUP.md`
- **CI/CD (Local):** `./scripts/run-all-tests.sh`
- **API Testing:** `catalog-api/docs/TESTING.md`

## 🎯 Quality Standards

Our quality gate requires:
- ✅ All unit tests passing
- ✅ All integration tests passing
- ✅ No critical security vulnerabilities
- ✅ Coverage threshold met (80% for frontend)
- ✅ Build succeeds
- ✅ Linting passes
- ✅ No merge conflicts

---

**The Catalogizer QA system provides real, comprehensive testing that ensures code quality at every level!** 🚀

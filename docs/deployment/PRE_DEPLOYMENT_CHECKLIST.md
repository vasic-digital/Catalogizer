# Pre-Deployment Checklist

**Project**: Catalogizer
**Version**: 1.0.0
**Target**: Production Deployment
**Date**: 2026-02-10

---

## 🔴 CRITICAL - MUST COMPLETE BEFORE DEPLOYMENT

### 1. Upgrade Go Runtime ⚠️ **REQUIRED**

**Current**: Go 1.25.5
**Required**: Go 1.25.7+
**Reason**: Fixes 3 critical standard library vulnerabilities

**Steps**:
```bash
# Download Go 1.25.7 or later
wget https://go.dev/dl/go1.25.7.linux-amd64.tar.gz

# Remove old version (if using system Go)
sudo rm -rf /usr/local/go

# Extract new version
sudo tar -C /usr/local -xzf go1.25.7.linux-amd64.tar.gz

# Verify
go version  # Should show go1.25.7 or higher
```

**Validation**:
- [ ] Go version shows 1.25.7+
- [ ] All tests pass: `go test ./...`
- [ ] Build succeeds: `go build`

---

### 2. Upgrade Dependencies ⚠️ **REQUIRED**

**Vulnerability**: HTTP/3 QPACK DoS in quic-go

**Steps**:
```bash
cd catalog-api
go get github.com/quic-go/quic-go@v0.57.0
go mod tidy
go test ./...
```

**Validation**:
- [ ] quic-go version is v0.57.0+
- [ ] All tests pass after upgrade
- [ ] No new vulnerabilities: `govulncheck ./...`

---

### 3. Run Complete Test Suite ⚠️ **REQUIRED**

**Backend Tests**:
```bash
cd catalog-api
go test ./... -race -coverprofile=coverage.out
# Expected: 100% pass rate, no race conditions
```

**Frontend Tests**:
```bash
cd catalog-web
npm run test:coverage
# Expected: 75%+ coverage, all tests passing
```

**Android Tests**:
```bash
cd catalogizer-android
./gradlew test
# Expected: 95%+ pass rate

cd catalogizer-androidtv
./gradlew test
# Expected: 100% pass rate
```

**Validation**:
- [ ] Backend: 100% pass rate, no races
- [ ] Frontend: 100% pass rate, 75%+ coverage
- [ ] Android: 95%+ pass rate
- [ ] AndroidTV: 100% pass rate

---

### 4. Security Validation ⚠️ **REQUIRED**

**Run All Security Scans**:
```bash
# Go vulnerabilities
govulncheck ./...
# Expected: 0 vulnerabilities

# Static analysis
staticcheck ./...
# Review findings (low priority)

# Security issues
gosec ./...
# Review critical/high issues only

# npm vulnerabilities
cd catalog-web && npm audit
cd catalogizer-api-client && npm audit
# Expected: 0 production vulnerabilities
```

**Validation**:
- [ ] govulncheck: 0 vulnerabilities
- [ ] npm audit: 0 production vulnerabilities
- [ ] No critical gosec issues

---

## 🟡 HIGH PRIORITY - RECOMMENDED BEFORE DEPLOYMENT

### 5. Database Migrations

**Verify**:
- [ ] All migrations applied successfully
- [ ] Database indexes created
- [ ] Foreign keys properly configured
- [ ] Backup strategy in place

**Commands**:
```bash
# Test migrations on staging
cd catalog-api
go run cmd/migrate/main.go up

# Verify schema
sqlite3 catalogizer.db ".schema"
```

---

### 6. Environment Configuration

**Production Environment Variables**:
- [ ] JWT_SECRET set to strong random value (32+ chars)
- [ ] ADMIN_PASSWORD set to secure password
- [ ] DB_TYPE configured (sqlite/postgres)
- [ ] Database credentials secured
- [ ] API keys for TMDB/OMDB configured
- [ ] CORS origins configured correctly
- [ ] GIN_MODE=release

**Example `.env.production`**:
```env
PORT=8080
GIN_MODE=release
JWT_SECRET=<STRONG-RANDOM-SECRET-32-CHARS>
ADMIN_PASSWORD=<SECURE-ADMIN-PASSWORD>
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=catalogizer_prod
DB_USER=catalogizer
DB_PASSWORD=<SECURE-DB-PASSWORD>
TMDB_API_KEY=<YOUR-KEY>
OMDB_API_KEY=<YOUR-KEY>
CORS_ORIGINS=https://yourdomain.com
```

**Validation**:
- [ ] All required env vars set
- [ ] Secrets are strong (checked with password strength tool)
- [ ] No sensitive data in git

---

### 7. Build Production Artifacts

**Backend Build**:
```bash
cd catalog-api
CGO_ENABLED=1 go build -ldflags="-s -w" -o catalog-api-linux-amd64 main.go
# Size should be ~50-80MB
```

**Frontend Build**:
```bash
cd catalog-web
npm run build
# Output: dist/ directory with optimized assets
# Verify: dist/index.html exists, bundle size < 500KB
```

**Desktop Apps** (if deploying):
```bash
cd catalogizer-desktop
npm run tauri:build
# Output: src-tauri/target/release/
```

**Validation**:
- [ ] Backend binary created and executable
- [ ] Frontend dist/ directory created
- [ ] Bundle sizes within limits
- [ ] No build errors or warnings

---

### 8. Docker Images (if using containers)

**Build Images**:
```bash
docker build -t catalogizer-api:1.0.0 -f Dockerfile.api .
docker build -t catalogizer-web:1.0.0 -f Dockerfile.web .
```

**Security Scan**:
```bash
trivy image catalogizer-api:1.0.0
trivy image catalogizer-web:1.0.0
# Expected: No HIGH or CRITICAL vulnerabilities
```

**Validation**:
- [ ] Images build successfully
- [ ] Images scanned for vulnerabilities
- [ ] Image sizes reasonable (<500MB)

---

## 🟢 RECOMMENDED - BEST PRACTICES

### 9. Performance Testing

**Load Testing**:
```bash
# Run load tests (if available)
cd catalog-api/tests/stress
go test -v
# Expected: Handle 10K concurrent users
```

**API Performance**:
- [ ] p95 response time < 200ms
- [ ] p99 response time < 500ms
- [ ] No memory leaks under sustained load

**Frontend Performance**:
- [ ] Lighthouse score > 90
- [ ] First Contentful Paint < 2s
- [ ] Largest Contentful Paint < 2.5s

---

### 10. Monitoring & Observability

**Set Up Monitoring**:
- [ ] Prometheus metrics endpoint exposed
- [ ] Grafana dashboards configured
- [ ] Alert rules defined
- [ ] Log aggregation configured
- [ ] Error tracking (Sentry/similar) configured

**Health Checks**:
- [ ] `/health` endpoint returns 200
- [ ] `/ready` endpoint returns 200
- [ ] Database connection validated

---

### 11. Backup & Recovery

**Backup Strategy**:
- [ ] Database backup cron job configured
- [ ] Backup retention policy defined
- [ ] Backup restoration tested
- [ ] Backup encryption enabled

**Disaster Recovery**:
- [ ] Recovery procedure documented
- [ ] RTO/RPO defined
- [ ] Failover tested (if multi-region)

---

### 12. Documentation Review

**Verify Complete**:
- [ ] API documentation updated
- [ ] Deployment guide reviewed
- [ ] Operations runbook created
- [ ] Troubleshooting guide available
- [ ] Security policies documented

---

## 🔒 Security Hardening

### 13. Security Configuration

**Application Security**:
- [ ] HTTPS only (no HTTP)
- [ ] TLS 1.2+ enforced
- [ ] Security headers configured (CSP, HSTS, X-Frame-Options)
- [ ] Rate limiting enabled
- [ ] CORS properly configured
- [ ] File upload limits set
- [ ] SQL injection protection verified
- [ ] XSS protection verified

**Server Security**:
- [ ] Firewall configured
- [ ] Only required ports open
- [ ] SSH key-only authentication
- [ ] OS patches up to date
- [ ] Non-root user for application

---

### 14. Access Control

**User Management**:
- [ ] Default admin password changed
- [ ] User roles properly configured
- [ ] Password policy enforced
- [ ] Session timeout configured
- [ ] Multi-factor authentication (if required)

**API Security**:
- [ ] API keys rotated
- [ ] Token expiration configured
- [ ] Refresh token rotation enabled

---

## 📊 Final Validation

### Pre-Deployment Sign-Off

**Technical Lead**:
- [ ] Code review complete
- [ ] All tests passing
- [ ] Security audit passed
- [ ] Performance validated

**DevOps**:
- [ ] Infrastructure ready
- [ ] Monitoring configured
- [ ] Backup strategy in place
- [ ] Rollback plan ready

**Security**:
- [ ] Security scans clean
- [ ] Compliance verified
- [ ] Penetration test complete (if required)

---

## 🚀 Deployment Steps

### Production Deployment Procedure

1. **Pre-Deployment**:
   - [ ] Announce maintenance window
   - [ ] Create database backup
   - [ ] Tag release in git
   - [ ] Build and test artifacts

2. **Deployment**:
   - [ ] Deploy database migrations
   - [ ] Deploy backend service
   - [ ] Deploy frontend assets
   - [ ] Verify health checks

3. **Post-Deployment**:
   - [ ] Run smoke tests
   - [ ] Verify all services healthy
   - [ ] Check monitoring dashboards
   - [ ] Announce deployment complete

4. **Rollback Plan** (if issues):
   - [ ] Revert to previous version
   - [ ] Restore database backup
   - [ ] Notify stakeholders

---

## 📝 Deployment Checklist Summary

### Critical (MUST DO)
- ✅ Upgrade Go to 1.25.7+
- ✅ Upgrade quic-go to v0.57.0
- ✅ Run all tests (100% pass rate required)
- ✅ Security scans clean

### High Priority (SHOULD DO)
- ⚠️ Database migrations verified
- ⚠️ Environment variables configured
- ⚠️ Production builds created
- ⚠️ Docker images built and scanned

### Recommended (NICE TO HAVE)
- 📋 Performance testing complete
- 📋 Monitoring configured
- 📋 Backup strategy in place
- 📋 Documentation reviewed

---

## ⚠️ Known Issues & Mitigations

### Non-Blocking Issues

1. **1 Android Test Failure** (Low Priority)
   - Issue: AuthViewModelTest mock timing
   - Impact: None on production (test issue only)
   - Mitigation: Investigate post-deployment

2. **2 npm Dev Dependencies** (Low Priority)
   - Issue: esbuild/vite dev server vulnerabilities
   - Impact: None (dev dependencies only)
   - Mitigation: None required for production

3. **381 Unhandled Errors** (Code Quality)
   - Issue: gosec findings
   - Impact: Potential hidden errors
   - Mitigation: Monitor logs, add handling post-deployment

---

## ✅ Sign-Off

**Date**: _______________

**Technical Lead**: _______________ ☐ Approved ☐ Rejected

**DevOps Lead**: _______________ ☐ Approved ☐ Rejected

**Security Lead**: _______________ ☐ Approved ☐ Rejected

**Product Owner**: _______________ ☐ Approved ☐ Rejected

---

## 📞 Emergency Contacts

**On-Call Engineer**: _______________
**DevOps Lead**: _______________
**Database Admin**: _______________

---

**Last Updated**: 2026-02-10
**Version**: 1.0
**Status**: Ready for review after critical fixes

**End of Pre-Deployment Checklist**

# Module 5: Administration - Slide Outlines

---

## Slide 5.0.1: Title Slide

**Title**: Administration & Configuration

**Subtitle**: User Management, Security, Monitoring, Backup, and Troubleshooting

**Speaker Notes**: This module is for anyone responsible for running a Catalogizer instance. We cover the full administration lifecycle: creating users, securing the system, monitoring health, backing up data, and diagnosing problems.

---

## Slide 5.1.1: User Management

**Title**: Managing Users and Roles

**Bullet Points**:
- Admin Panel (`AdminPanel.tsx`) provides centralized user management
- Create, edit, deactivate user accounts
- Role-based access control: Admin, User, Viewer
- Password requirements: 8+ characters, uppercase, lowercase, numbers, special characters
- Passwords hashed before storage -- never stored in plain text

**Visual**: Screenshot of the Admin Panel user list with role badges

**Speaker Notes**: Start by creating a few test accounts with different roles to demonstrate the permission differences. Admin sees everything. User can manage their own content. Viewer has read-only access.

---

## Slide 5.1.2: Session Management

**Title**: Controlling Active Sessions

**Bullet Points**:
- View all active sessions for any user
- See login device, location, and time
- Force logout individual sessions
- Force logout all sessions for a user
- Essential for security when devices are lost or compromised
- Auth code: `internal/auth/service.go`, `middleware.go`, `models.go`

**Speaker Notes**: Session management is a security feature. If a user reports a stolen laptop, you can immediately terminate all their sessions from the Admin Panel, preventing unauthorized access.

---

## Slide 5.2.1: Security Layers

**Title**: Defense in Depth

**Bullet Points**:
- **Layer 1**: JWT authentication with configurable expiry (JWT_SECRET, JWT_EXPIRY_HOURS, REFRESH_TOKEN_EXPIRY_HOURS)
- **Layer 2**: SQLCipher database encryption (AES-256, DB_ENCRYPTION_KEY -- exactly 32 characters)
- **Layer 3**: CORS and rate limiting middleware
- **Layer 4**: Two-factor authentication with authenticator app support
- **Layer 5**: Security scanning (Snyk, SonarQube)

**Visual**: Layered security diagram showing each protection layer

**Speaker Notes**: Security is not a single feature but a layered approach. Even if one layer is bypassed, the others provide protection. JWT handles authentication, encryption protects data at rest, middleware prevents abuse, 2FA adds identity verification, and scanning catches vulnerabilities before they are exploited.

---

## Slide 5.2.2: Security Testing Tools

**Title**: Automated Security Verification

**Bullet Points**:
- `scripts/security-test.sh`: Automated security checks
- `docker-compose.security.yml`: Security-focused testing environment
- `scripts/snyk-scan.sh`: Dependency vulnerability scanning (Go + npm)
- `scripts/sonarqube-scan.sh`: Static code analysis for bugs and security flaws
- `dependency-check-suppressions.xml`: Manage false positives
- Run scans before every deployment

**Speaker Notes**: These tools should be part of your regular maintenance routine. New vulnerabilities in dependencies are discovered daily. Running Snyk weekly and SonarQube before each deployment catches issues early.

---

## Slide 5.3.1: Monitoring Stack

**Title**: Prometheus + Grafana for Observability

**Bullet Points**:
- Backend metrics defined in `internal/metrics/metrics.go`
- Automatic HTTP instrumentation via `internal/metrics/middleware.go`
- Prometheus configuration: `monitoring/prometheus.yml`
- Pre-built Grafana dashboards: `monitoring/grafana/`
- Dashboards: API performance, detection pipeline throughput, storage health

**Visual**: Screenshot of a Grafana dashboard showing API request rates and latencies

**Speaker Notes**: The monitoring stack gives you visibility into how Catalogizer is performing. Are API response times increasing? Is the detection pipeline keeping up with new files? Are any storage sources having connectivity issues? All answerable from the dashboards.

---

## Slide 5.3.2: SMB Connection Monitoring

**Title**: Storage Source Health

**Bullet Points**:
- Circuit breaker states: Closed (healthy), Open (disconnected), Half-Open (testing)
- Offline cache metrics: hit rate, cached items count
- Retry metrics: attempt counts, backoff delays
- Source health indicators: Green (healthy), Yellow (degraded), Red (offline/cached)
- Configure Grafana alerts for state transitions

**Visual**: Health status indicator panel: source names with colored status dots

**Speaker Notes**: SMB monitoring is especially important because network shares are inherently unreliable. The traffic light indicator system makes it easy to spot problems at a glance. Set up alerts so you are notified when a source goes red rather than discovering it when users complain.

---

## Slide 5.4.1: Backup Strategy

**Title**: What to Back Up

**Bullet Points**:
- **Critical**: SQLCipher database file (encrypted; requires DB_ENCRYPTION_KEY to restore)
- **Critical**: `.env` files with server configuration and API keys
- **Important**: `config/` directory (nginx.conf, redis.conf)
- **Recommended**: Export favorites and collections as JSON
- **Optional**: Cloud sync to S3, Google Cloud Storage, or local folders

**Visual**: Priority matrix showing backup targets by criticality

**Speaker Notes**: The database is the most important backup target. It contains all metadata, user accounts, collections, favorites, and settings. Without it, you would need to rescan every source from scratch. Without the encryption key, the database is unreadable -- store them separately.

---

## Slide 5.4.2: Three-Tier Backup Strategy

**Title**: Daily, Weekly, Monthly

**Bullet Points**:
- **Daily**: Automated database backup (cron job or scheduled task)
- **Weekly**: Full configuration export (.env, config/, generated reports)
- **Monthly**: Restore verification (test that backups are actually restorable)
- Store DB_ENCRYPTION_KEY separately from database backups
- Use the reporting service for point-in-time PDF snapshots

**Speaker Notes**: A backup that has never been tested is not a backup. The monthly verification step is the most important part of this strategy. Actually restore from a backup to a test environment and verify that everything works.

---

## Slide 5.4.3: Restore Procedures

**Title**: Recovery Scenarios

| Scenario | Procedure |
|----------|-----------|
| Database corruption | Stop server, replace DB file from backup, verify encryption key matches, restart |
| Configuration loss | Restore .env and config/ files, restart services |
| Media source failure | Catalog metadata preserved; reconnect source after recovery, rescan |
| Full disaster | Restore database + configuration, reconnect sources, verify |

**Speaker Notes**: Walk through each scenario. Emphasize that Catalogizer's metadata is independent of the media files themselves. Even if a NAS dies, all the metadata, collections, favorites, and organizational work is preserved in the database backup.

---

## Slide 5.5.1: Troubleshooting SMB Issues

**Title**: Diagnosing Connection Problems

**Bullet Points**:
- Set `LOG_LEVEL=debug` for detailed logging
- Circuit breaker transitions in logs: closed -> open (disconnection detected)
- Open -> half-open (reconnection attempt), half-open -> closed (recovery)
- Check: SMB_RETRY_ATTEMPTS, SMB_RETRY_DELAY_SECONDS, SMB_HEALTH_CHECK_INTERVAL
- Offline cache serves data during outages

**Visual**: Log excerpt showing circuit breaker state transitions

**Speaker Notes**: When troubleshooting SMB, start with the logs. The circuit breaker transitions tell you exactly what happened and when. If you see rapid closed -> open cycling, the share is unstable. Increase the health check interval to reduce network overhead.

---

## Slide 5.5.2: Common Issues Checklist

**Title**: Quick Diagnosis Guide

**Bullet Points**:
- **No real-time updates**: Check WebSocket connection in browser dev tools; check firewall/proxy
- **Slow detection**: Tune MAX_CONCURRENT_ANALYSIS and ANALYSIS_TIMEOUT_MINUTES
- **Missing metadata**: Verify external API keys (TMDB_API_KEY, etc.); check rate limits
- **Login failures**: Verify JWT_SECRET matches across restarts; check token expiry settings
- **High error rate**: Check metrics dashboard; review backend logs for 500 errors
- **Crash recovery**: `internal/recovery/` restores state from persistent storage on restart

**Speaker Notes**: This checklist covers the most common issues. For each one, the resolution is straightforward. The key is knowing where to look. Logs for backend issues, browser dev tools for frontend issues, and the metrics dashboard for performance issues.

---

## Slide 5.5.3: Module 5 Summary

**Title**: What We Covered

**Bullet Points**:
- User management with role-based access control and session management
- Multi-layer security: JWT, encryption, CORS, 2FA, scanning
- Prometheus + Grafana monitoring with pre-built dashboards
- Three-tier backup strategy: daily database, weekly configuration, monthly verification
- Cloud sync to S3 or Google Cloud Storage
- Troubleshooting with debug logs, circuit breaker state tracking, and recovery mechanisms

**Next Steps**: Module 6 -- Developer Guide (Architecture, Dev Environment, Features, Submodules, Build Pipeline)

**Speaker Notes**: Administrators should now be confident in managing a Catalogizer deployment end-to-end. From user onboarding to disaster recovery, all the tools and procedures have been covered. Module 6 is for those who want to extend or contribute to Catalogizer itself.

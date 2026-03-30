# Module 5: Administration and Configuration - Slide Deck Outline

**Total Slides**: 13
**Estimated Duration**: 65 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Administration and Configuration

- User management, security, monitoring, backups, troubleshooting
- Prerequisites: Module 2 completed, admin account access
- By the end: manage users, configure security, and restore from backups

---

## Slide 2: Admin Panel Overview (4 min)

**Title**: The Admin Panel

- Access AdminPanel.tsx from the profile menu (admin role required)
- Sections: Users, Security, Monitoring, Backups, System Settings
- Role-based access control: admin vs user capabilities
- Demo: navigate the admin panel sections

---

## Slide 3: User Management (5 min)

**Title**: Creating and Managing User Accounts

- Create, edit, and deactivate user accounts
- JWT-based authentication via internal/auth/service.go
- Role assignment: admin or user
- Manage active sessions and force logout when needed
- Exercise reference: Exercise 5.1 -- create a user and assign a role

---

## Slide 4: Security Configuration (5 min)

**Title**: JWT, Encryption, and Rate Limiting

- JWT_SECRET: signs authentication tokens (32+ random characters)
- JWT_EXPIRY_HOURS and REFRESH_TOKEN_EXPIRY_HOURS for token lifetime
- DB_ENCRYPTION_KEY: encrypts database at rest (exactly 32 characters)
- Auth rate limiter: strict 5/min on login/register, default 100/min elsewhere
- Demo: configure JWT secret and verify token-based access

---

## Slide 5: Two-Factor Authentication (4 min)

**Title**: Adding a Second Layer of Security

- Enable 2FA for individual user accounts
- TOTP-based (time-based one-time password) implementation
- Recovery codes generated during setup
- Mandatory 2FA enforcement option for all users
- Exercise reference: Exercise 5.2 -- enable 2FA for the admin account

---

## Slide 6: Monitoring and Metrics (5 min)

**Title**: Prometheus Metrics and Health Checks

- internal/metrics/metrics.go exposes metrics at /metrics
- Prometheus scraping configured in monitoring/prometheus.yml
- Key metrics: request count, latency, error rate, active connections
- Health check endpoint: /api/v1/health
- Demo: query Prometheus for API latency percentiles

---

## Slide 7: Grafana Dashboards (5 min)

**Title**: Visualizing System Health

- Pre-built dashboards in monitoring/grafana/ and config/grafana-dashboards/
- Panels: API latency, request throughput, error rates, database queries
- SMB circuit breaker status and offline cache metrics
- Media detection pipeline throughput
- Exercise reference: Exercise 5.3 -- deploy Grafana and explore dashboards

---

## Slide 8: Backup Management (5 min)

**Title**: Creating and Restoring Backups

- POST /api/v1/admin/backup to create a database backup
- GET /api/v1/admin/backups to list available backups
- POST /api/v1/admin/backup/restore to restore from a backup
- Backup semaphore prevents concurrent backup operations
- Path traversal protection on restore for security
- Demo: create a backup and list it via the API

---

## Slide 9: Cloud Sync and Archiving (5 min)

**Title**: Syncing to S3, GCS, and Local Folders

- Synchronize files with Amazon S3 or Google Cloud Storage
- Configure automatic archiving rules for storage management
- Generate PDF reports with charts and analytics
- Three-tier strategy: daily database, weekly config, monthly verification
- Exercise reference: Exercise 5.4 -- set up a local backup schedule

---

## Slide 10: Disaster Recovery (5 min)

**Title**: Restoring After Data Loss

- Restore database from backup file
- Rebuild configuration from config.json template
- Re-scan storage sources after restoration
- Handle media source failure: circuit breaker + offline cache
- Exercise reference: Exercise 5.5 -- simulate and recover from database loss

---

## Slide 11: Troubleshooting SMB Connections (5 min)

**Title**: Diagnosing Network Storage Issues

- Circuit breaker states: closed (normal), open (failed), half-open (testing)
- Exponential backoff retry: configurable attempts and delay
- Offline cache serves data when storage is unavailable
- Configuration: SMB_RETRY_ATTEMPTS, SMB_RETRY_DELAY_SECONDS, SMB_HEALTH_CHECK_INTERVAL
- Demo: disconnect a share and observe circuit breaker behavior

---

## Slide 12: Log Management (4 min)

**Title**: Reviewing Logs and Adjusting Verbosity

- LOG_LEVEL environment variable: debug, info, warn, error
- Structured JSON logging for machine parsing
- Log rotation policies for production environments
- Recovery mechanisms in internal/recovery/ for crash recovery
- Exercise reference: Exercise 5.6 -- set debug logging and trace an issue

---

## Slide 13: Module Summary and Next Steps (3 min)

**Title**: What We Covered

- User management with JWT auth and role-based access
- Security: 2FA, encryption, rate limiting
- Monitoring with Prometheus and Grafana dashboards
- Backup, restore, and disaster recovery procedures
- SMB troubleshooting with circuit breaker diagnostics
- Next module: Developer Guide and API (for technical users)

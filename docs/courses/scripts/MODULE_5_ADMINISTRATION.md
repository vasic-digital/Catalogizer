# Module 5: Administration - Video Scripts

---

## Lesson 5.1: User Management & Roles

**Duration**: 12 minutes

### Narration

As an administrator, user management is one of your primary responsibilities. The Admin page, implemented in Admin.tsx, provides the interface for managing users, while the AdminPanel component in catalog-web/src/components/admin/AdminPanel.tsx provides the detailed administration controls.

User management starts with the authentication system. Catalogizer uses JWT-based authentication, implemented in internal/auth/. The auth service (service.go) handles user creation, credential validation, and token generation. The middleware (middleware.go) validates tokens on every protected request.

When creating a new user, you specify a username, password, email, and role. The password must meet security requirements: minimum 8 characters with uppercase, lowercase, numbers, and special characters. Passwords are hashed before storage -- never stored in plain text.

Roles control what users can do. The role-based access control system assigns permissions based on user roles. An admin has full access to all features including user management and system configuration. Regular users can browse, search, manage their own favorites and collections, and play media. Viewer roles might have read-only access.

From the Admin Panel, you can view all users, their roles, and their account status. You can edit user details, change roles, reset passwords, and deactivate accounts. Deactivated accounts cannot log in but their data is preserved.

Session management is also available. You can see active sessions for any user -- which devices they are logged in from -- and force logout individual sessions or all sessions for a user. This is important for security when a device is lost or compromised.

The auth models (models.go) define the data structures for users, tokens, and sessions. Token integration tests (token_integration_test.go) verify that the complete authentication flow works correctly.

### On-Screen Actions

- [00:00] Navigate to the Admin page in the web UI
- [00:30] Show the AdminPanel with the user list
- [01:00] Click "Create User" -- fill in the form: username, email, password, role
- [02:00] Show password validation requirements
- [02:30] Save the new user
- [03:00] Edit an existing user: change their role
- [03:30] Show the different role options and what each permits
- [04:30] View a user's active sessions
- [05:00] Force logout a specific session
- [05:30] Deactivate a user account
- [06:00] Show that the deactivated user cannot log in
- [06:30] Reactivate the account
- [07:00] Open internal/auth/service.go -- show user creation and token generation
- [07:30] Open internal/auth/middleware.go -- show token validation logic
- [08:30] Open internal/auth/models.go -- show user and token structures
- [09:00] Show middleware_test.go and service_test.go
- [09:30] Show token_integration_test.go
- [10:00] Open Admin.tsx and AdminPanel.tsx in the code
- [10:30] Show the role-based route protection in the frontend
- [11:00] Final overview of user management

### Key Points

- Admin Panel (AdminPanel.tsx) provides centralized user management
- JWT authentication: service.go handles tokens, middleware.go validates on every request
- Role-based access control: admin, user, viewer roles with different permissions
- Session management: view active sessions, force logout individual or all sessions
- Password hashing -- never stored in plain text; minimum complexity requirements enforced
- Auth code lives in internal/auth/: service.go, middleware.go, models.go

### Tips

> **Tip**: Regularly review active sessions in the Admin Panel. Unexpected sessions from unfamiliar locations or devices may indicate compromised credentials.

---

## Lesson 5.2: Security Configuration

**Duration**: 12 minutes

### Narration

Security is built into Catalogizer at multiple levels. Let us configure and verify each layer.

The first layer is JWT authentication. The JWT_SECRET environment variable must be a strong, random string. This secret signs all authentication tokens. JWT_EXPIRY_HOURS controls how long an access token is valid -- default 24 hours. REFRESH_TOKEN_EXPIRY_HOURS controls the refresh token -- default 168 hours (one week). Shorter expiry times are more secure but require more frequent re-authentication.

The second layer is database encryption. SQLCipher encrypts the entire database at rest using AES-256. The DB_ENCRYPTION_KEY must be exactly 32 characters. Without this key, the database file is unreadable. This protects sensitive metadata, user credentials, and session data even if someone gains access to the server filesystem.

For security testing, Catalogizer includes dedicated tools. The security test script at scripts/security-test.sh runs automated security checks. There is also a docker-compose.security.yml that sets up a security-focused testing environment.

Snyk scanning (scripts/snyk-scan.sh) checks your dependencies for known vulnerabilities. It scans both the Go backend and the Node.js frontend. SonarQube (scripts/sonarqube-scan.sh) performs static analysis to detect code quality issues, potential security flaws, and code smells. The sonar-project.properties file at the root configures the SonarQube analysis.

The dependency-check-suppressions.xml file manages false positives from dependency scanning, ensuring your reports focus on actual issues.

Two-factor authentication adds another security layer for user accounts. When enabled, users must provide a verification code from an authenticator app in addition to their password. This significantly reduces the risk of compromised credentials.

### On-Screen Actions

- [00:00] Open the .env file and highlight JWT configuration
- [01:00] Explain JWT_SECRET strength requirements
- [01:30] Show JWT_EXPIRY_HOURS and REFRESH_TOKEN_EXPIRY_HOURS settings
- [02:00] Highlight DB_ENCRYPTION_KEY and explain AES-256 encryption
- [02:30] Show what happens when trying to open the DB without the key
- [03:30] Open scripts/security-test.sh in the editor
- [04:00] Run the security test script -- show output
- [05:00] Open docker-compose.security.yml and explain its purpose
- [05:30] Run `docker-compose -f docker-compose.security.yml up`
- [06:00] Open scripts/snyk-scan.sh -- explain dependency vulnerability scanning
- [06:30] Run the Snyk scan -- show results
- [07:30] Open scripts/sonarqube-scan.sh
- [08:00] Show sonar-project.properties configuration
- [08:30] Open dependency-check-suppressions.xml -- explain false positive management
- [09:00] Demonstrate enabling two-factor authentication for a user
- [09:30] Show the QR code scan flow with an authenticator app
- [10:00] Log in with 2FA: enter password then verification code
- [10:30] Show backup codes
- [11:00] Recap all security layers

### Key Points

- JWT authentication: strong JWT_SECRET, configurable expiry for access and refresh tokens
- SQLCipher database encryption: AES-256 with 32-character DB_ENCRYPTION_KEY
- Security testing: scripts/security-test.sh and docker-compose.security.yml
- Dependency scanning: Snyk (scripts/snyk-scan.sh) for known vulnerabilities
- Static analysis: SonarQube (scripts/sonarqube-scan.sh) for code quality and security
- Two-factor authentication with authenticator app support

### Tips

> **Tip**: Run Snyk and SonarQube scans before every deployment. New vulnerabilities in dependencies are discovered regularly, and catching them early prevents security incidents.

> **Tip**: Store your DB_ENCRYPTION_KEY and JWT_SECRET in a secure secrets manager, not in version control. Losing the DB encryption key means losing access to your entire database.

---

## Lesson 5.3: Monitoring & Metrics

**Duration**: 15 minutes

### Narration

Catalogizer includes a comprehensive monitoring stack based on Prometheus and Grafana. Understanding your system's health is essential for maintaining reliable service.

The metrics system starts in the backend. The internal/metrics/metrics.go file defines the metrics that Catalogizer exposes. These include request counts, response times, error rates, and custom metrics for media detection and scanning.

The metrics middleware (internal/metrics/middleware.go) automatically instruments every HTTP request. It records the endpoint, method, status code, and latency. This gives you visibility into API performance without any manual instrumentation.

Prometheus is configured via monitoring/prometheus.yml. This file tells Prometheus where to scrape metrics from Catalogizer. The scrape interval, target endpoints, and any label additions are configured here.

Grafana dashboards live in monitoring/grafana/. These pre-built dashboards visualize the most important operational metrics. You will find dashboards for API request rates and latencies, media detection pipeline throughput, storage source connection health, and system resource usage.

For SMB specifically, the monitoring is detailed. The circuit breaker in internal/smb/ tracks connection state transitions: closed (healthy), open (disconnected), and half-open (testing reconnection). The offline cache metrics show how many items are being served from cache during disconnections. Exponential backoff retry metrics show reconnection attempts.

Health check status is visible for each connected source. Green means healthy with active connection. Yellow means degraded with some retries happening. Red means offline with data being served from cache.

Setting up the monitoring stack is straightforward with Docker. The prometheus.yml and grafana configuration files are ready to use. Just include them in your deployment and access Grafana through its web interface.

### On-Screen Actions

- [00:00] Open internal/metrics/metrics.go -- show defined metrics
- [01:00] Open internal/metrics/middleware.go -- show automatic instrumentation
- [02:00] Show metrics_test.go and middleware_test.go
- [02:30] Open monitoring/prometheus.yml -- show scrape configuration
- [03:30] Open the monitoring/grafana/ directory and show dashboard files
- [04:30] Open Grafana in a browser -- show the dashboard list
- [05:00] Open the API performance dashboard: request rate, latency histograms
- [06:00] Open the media detection dashboard: pipeline throughput, detection counts by type
- [07:00] Open the storage health dashboard: connection status per source
- [07:30] Show SMB-specific monitoring: circuit breaker state, retry counts
- [08:30] Show offline cache metrics: cache hit rate, cached items count
- [09:00] Show the source health indicators: green, yellow, red
- [09:30] Demonstrate a simulated disconnection: watch metrics change
- [10:30] Show the circuit breaker transitioning from closed to open
- [11:00] Show offline cache serving data during the disconnection
- [11:30] Reconnect and show recovery: circuit breaker goes to half-open then closed
- [12:30] Open Prometheus directly -- show raw metric queries
- [13:00] Write a custom PromQL query for request latency
- [13:30] Show alert configuration possibilities
- [14:00] Final overview of the monitoring stack

### Key Points

- Backend metrics defined in internal/metrics/metrics.go with automatic HTTP instrumentation via middleware
- Prometheus configured via monitoring/prometheus.yml for metric scraping
- Pre-built Grafana dashboards in monitoring/grafana/ for API, detection, and storage health
- SMB monitoring: circuit breaker states, offline cache metrics, retry counts
- Source health indicators: green (healthy), yellow (degraded), red (offline/cached)
- Full Prometheus + Grafana stack deployable via Docker

### Tips

> **Tip**: Set up Grafana alerts for circuit breaker state changes. Getting notified when a storage source goes offline lets you investigate quickly rather than discovering issues when users report them.

> **Tip**: Monitor the media detection pipeline throughput after connecting new large sources. A sudden drop in throughput might indicate resource constraints that need attention.

---

## Lesson 5.4: Backup, Restore & Cloud Sync

**Duration**: 14 minutes

### Narration

Protecting your Catalogizer data requires a solid backup strategy. Let us cover the key areas you need to back up and the tools available.

The most critical item is your database. The SQLCipher-encrypted SQLite file contains all your media metadata, user accounts, collections, playlists, favorites, and settings. Regular backups of this file are essential. Remember that backups of an encrypted database are only useful if you also have the DB_ENCRYPTION_KEY.

Configuration files are the next priority. Your .env files contain server settings, API keys, and connection parameters. The config directory has nginx.conf and redis.conf. Losing these means manual reconfiguration.

Catalogizer supports cloud storage synchronization. You can sync files with Amazon S3, Google Cloud Storage, or local backup folders. This feature, described in the README, allows you to push media or metadata to cloud storage for redundancy.

The advanced reporting system can generate professional PDF reports with charts and analytics. These reports capture the state of your library at a point in time, serving as both documentation and a form of metadata backup.

Storage management tools help you maintain a healthy library. Archive settings let you configure automatic archiving rules -- for example, moving media older than a certain date to an archive source. Cleanup tools identify and remove duplicate or unwanted files.

For restore scenarios, the process depends on what failed. If the database is corrupted, restore from your latest backup file and ensure the encryption key matches. If configuration is lost, restore .env and config/ files. If media files are lost on a source, the catalog still has the metadata -- reconnect the source after recovery and Catalogizer rescans.

I recommend implementing a three-tier backup strategy. Daily automated database backups. Weekly full configuration exports. Monthly verification that backups are restorable.

### On-Screen Actions

- [00:00] Show the database file location (DB_PATH from .env)
- [00:30] Demonstrate creating a manual database backup (copy the file)
- [01:30] Show how to automate backups with a cron job
- [02:00] List all files that need backup: database, .env, config/
- [03:00] Show cloud sync configuration for S3
- [04:00] Configure Google Cloud Storage sync
- [05:00] Set up local folder backup destination
- [05:30] Trigger a sync and show files being uploaded
- [06:00] Show the advanced reporting feature: generate a PDF report
- [06:30] Open the generated PDF: show charts, statistics, library summary
- [07:30] Demonstrate storage management: view storage usage
- [08:00] Show cleanup tools: duplicate detection
- [08:30] Configure automatic archiving rules
- [09:00] Simulate a database restore: stop the server, replace the DB file, restart
- [09:30] Verify all data is intact after restore
- [10:00] Simulate a configuration restore: restore .env and config/ files
- [10:30] Verify services start correctly with restored configuration
- [11:00] Demonstrate the three-tier backup strategy
- [12:00] Show a backup verification checklist
- [12:30] Discuss disaster recovery scenarios
- [13:00] Final overview

### Key Points

- Critical backup targets: SQLCipher database (with encryption key), .env files, config/ directory
- Cloud sync: Amazon S3, Google Cloud Storage, or local backup folders
- Advanced reporting: PDF reports with charts and analytics for point-in-time snapshots
- Storage management: cleanup tools, duplicate detection, automatic archiving rules
- Restore: replace database file and/or configuration, restart services
- Three-tier strategy: daily DB backup, weekly config export, monthly restore verification

### Tips

> **Tip**: Always test your backup restoration process before you need it. A backup that cannot be restored is not a backup at all.

> **Tip**: Store the DB_ENCRYPTION_KEY separately from the database backup. If both are lost together in a security incident, the backup is useless. If the key is stored separately, the encrypted backup remains secure.

---

## Lesson 5.5: Troubleshooting & Resilience

**Duration**: 12 minutes

### Narration

Even well-configured systems encounter issues. Let us go through common problems and how to diagnose and resolve them.

The most frequent issue is SMB connection problems. Catalogizer is built with extensive resilience for this. The internal/smb/ directory contains the circuit breaker pattern, offline cache, and exponential backoff retry logic.

When an SMB source becomes unreachable, the circuit breaker opens. This prevents the system from repeatedly trying a connection that is known to be down. Instead, data is served from the offline cache. The retry mechanism uses exponential backoff -- starting at SMB_RETRY_DELAY_SECONDS and doubling with each attempt, up to SMB_RETRY_ATTEMPTS times.

To diagnose SMB issues, check the logs first. Set LOG_LEVEL to debug for detailed output. Look for circuit breaker state transitions: closed to open means a disconnection was detected. Open to half-open means a reconnection attempt is being made. Half-open to closed means recovery succeeded.

The recovery package (internal/recovery/) provides crash recovery mechanisms. If the application terminates unexpectedly, recovery routines restore state from persistent storage on the next startup.

For API issues, the metrics middleware logs every request with status codes and latencies. A spike in 500 errors points to backend problems. Slow response times might indicate database issues or overwhelmed detection pipeline.

Frontend issues are usually related to WebSocket connectivity. If real-time updates stop, check the browser console for WebSocket errors. The WebSocketContext automatically attempts reconnection, but firewall or proxy changes might block WebSocket traffic.

For media detection problems, check the detection pipeline configuration. MAX_CONCURRENT_ANALYSIS limits parallelism -- too low and scanning is slow, too high and the system becomes overloaded. ANALYSIS_TIMEOUT_MINUTES prevents individual items from blocking the pipeline.

External provider issues -- TMDB, Spotify, etc. -- are usually due to expired or rate-limited API keys. Check the provider-specific error logs and verify your API keys are valid.

### On-Screen Actions

- [00:00] Show a healthy Catalogizer installation with all sources green
- [00:30] Simulate an SMB disconnection (disconnect a network share)
- [01:00] Show the circuit breaker opening in the logs
- [01:30] Show the offline cache serving data
- [02:00] Show retry attempts with exponential backoff in the logs
- [02:30] Reconnect the share and show recovery
- [03:00] Open internal/smb/ directory -- show circuit breaker code
- [03:30] Open the retry logic with exponential backoff
- [04:00] Show the offline cache implementation
- [04:30] Show internal/recovery/ -- crash recovery mechanisms
- [05:00] Change LOG_LEVEL to debug and show verbose output
- [05:30] Demonstrate diagnosing an API error from logs
- [06:00] Show metrics indicating a problem: high error rate or latency
- [06:30] Open browser console and show WebSocket connection status
- [07:00] Simulate a WebSocket disconnection and reconnection
- [07:30] Show detection pipeline configuration: MAX_CONCURRENT_ANALYSIS, ANALYSIS_TIMEOUT_MINUTES
- [08:00] Diagnose a slow detection issue by adjusting parallelism
- [08:30] Show an external provider error in logs (expired API key)
- [09:00] Update the API key and show recovery
- [09:30] Walk through the install script: scripts/install.sh
- [10:00] Show the setup implementation script: scripts/setup-implementation.sh
- [10:30] Common issues checklist and resolution steps
- [11:00] Final overview of troubleshooting approach

### Key Points

- SMB resilience: circuit breaker (prevents storm of failed retries), offline cache (serves data during outage), exponential backoff (gradually retries)
- Diagnose with LOG_LEVEL=debug for detailed logging
- Circuit breaker states: closed (healthy) -> open (disconnected) -> half-open (testing) -> closed (recovered)
- Recovery package (internal/recovery/) handles crash recovery on restart
- Detection pipeline tuning: MAX_CONCURRENT_ANALYSIS and ANALYSIS_TIMEOUT_MINUTES
- External provider issues typically involve API key expiry or rate limiting
- Scripts available: install.sh, setup-implementation.sh for environment setup

### Tips

> **Tip**: When troubleshooting, always start with the logs. Set LOG_LEVEL to debug temporarily, reproduce the issue, then examine the log output. Remember to set it back to info when done to avoid excessive log volume.

> **Tip**: If SMB sources frequently disconnect, increase SMB_HEALTH_CHECK_INTERVAL to reduce network overhead, but decrease SMB_RETRY_DELAY_SECONDS for faster recovery when disconnections do occur.

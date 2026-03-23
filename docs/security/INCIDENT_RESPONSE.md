# Security Incident Response Plan

This document defines the security incident response procedures for Catalogizer deployments. It covers detection, containment, eradication, recovery, and post-incident review.

---

## Table of Contents

1. [Incident Classification](#incident-classification)
2. [Roles and Responsibilities](#roles-and-responsibilities)
3. [Phase 1: Detection and Identification](#phase-1-detection-and-identification)
4. [Phase 2: Containment](#phase-2-containment)
5. [Phase 3: Eradication](#phase-3-eradication)
6. [Phase 4: Recovery](#phase-4-recovery)
7. [Phase 5: Post-Incident Review](#phase-5-post-incident-review)
8. [Incident Types and Playbooks](#incident-types-and-playbooks)
9. [Communication Templates](#communication-templates)

---

## Incident Classification

### Severity Levels

| Level | Name | Description | Response Time |
|-------|------|-------------|---------------|
| P1 | Critical | Active data breach, system compromise, or complete service outage | Immediate (< 15 min) |
| P2 | High | Unauthorized access detected, vulnerability actively exploited | < 1 hour |
| P3 | Medium | Suspicious activity, failed attack attempts, non-critical vulnerability found | < 4 hours |
| P4 | Low | Policy violation, informational security event | < 24 hours |

### Incident Categories

- **Authentication Breach**: Compromised credentials, JWT token theft, session hijacking
- **Data Exposure**: Unauthorized access to media metadata, user data, or configuration
- **Service Disruption**: DDoS, resource exhaustion, crash exploitation
- **Malware/Injection**: SQL injection, XSS, command injection
- **Configuration Exposure**: Leaked API keys, database credentials, JWT secrets
- **Insider Threat**: Unauthorized actions by authenticated users

---

## Roles and Responsibilities

| Role | Responsibilities |
|------|-----------------|
| Incident Commander | Coordinates response, makes escalation decisions, communicates with stakeholders |
| Security Analyst | Investigates the incident, performs forensic analysis, identifies scope |
| System Administrator | Executes containment and recovery actions on infrastructure |
| Database Administrator | Investigates database-level compromise, restores data |
| Communications Lead | Manages internal and external notifications |

---

## Phase 1: Detection and Identification

### Detection Sources

1. **Application Logs**: Structured Zap logs in `catalog-api`
2. **Auth Audit Log**: `auth_audit_log` table records all authentication events
3. **Prometheus Metrics**: Anomalous request rates, error spikes at `/metrics`
4. **Grafana Alerts**: Pre-configured alerts for latency, error rates, availability
5. **Security Scans**: govulncheck, Semgrep, SonarQube, Snyk, Trivy findings
6. **User Reports**: Reports from users about unexpected behavior

### Detection Procedures

```bash
# Check for unusual authentication patterns
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT event_type, COUNT(*), ip_address
FROM auth_audit_log
WHERE created_at > NOW() - INTERVAL '1 hour'
GROUP BY event_type, ip_address
ORDER BY COUNT(*) DESC
LIMIT 20;"

# Check for brute force attempts
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT u.username, u.failed_login_attempts, u.is_locked, u.last_login_ip
FROM users u
WHERE u.failed_login_attempts > 5
ORDER BY u.failed_login_attempts DESC;"

# Check for unusual API request patterns
curl -s http://localhost:8080/metrics | grep 'http_requests_total'

# Review recent error logs
journalctl -u catalogizer --since "1 hour ago" | grep -i -E "error|unauthorized|forbidden|inject"

# Check active sessions for anomalies
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT u.username, s.ip_address, s.user_agent, s.created_at, s.last_activity_at
FROM user_sessions s
JOIN users u ON s.user_id = u.id
WHERE s.is_active = TRUE
ORDER BY s.last_activity_at DESC;"
```

### Identification Checklist

- [ ] What type of incident is this? (See categories above)
- [ ] When did it start? (Earliest evidence in logs)
- [ ] What systems are affected?
- [ ] How was access gained? (If applicable)
- [ ] What data may be exposed?
- [ ] Is the incident ongoing?
- [ ] Assign severity level (P1-P4)

---

## Phase 2: Containment

### Immediate Containment (Short-Term)

**For Authentication Breach:**

```bash
# Lock compromised user accounts
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
UPDATE users SET is_locked = TRUE, locked_until = NOW() + INTERVAL '24 hours'
WHERE username IN ('compromised_user');"

# Invalidate all sessions for compromised users
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
UPDATE user_sessions SET is_active = FALSE
WHERE user_id IN (SELECT id FROM users WHERE username IN ('compromised_user'));"

# If JWT_SECRET is compromised, rotate it immediately
# Update .env with new JWT_SECRET and restart
# This invalidates ALL active tokens system-wide
sudo systemctl restart catalogizer
```

**For Active Exploitation:**

```bash
# Block attacking IP at the firewall
sudo iptables -A INPUT -s <attacker_ip> -j DROP

# If running behind nginx, add to deny list
echo "deny <attacker_ip>;" >> /etc/nginx/conf.d/block.conf
sudo systemctl reload nginx

# Rate limit is already in place (5/min auth, 100/min general)
# If insufficient, temporarily reduce limits
```

**For Data Exposure:**

```bash
# Rotate all exposed credentials immediately
# 1. Database password
PGPASSWORD=$OLD_PASSWORD psql -h $DB_HOST -U postgres -c \
  "ALTER USER catalogizer_user PASSWORD 'new_secure_password';"

# 2. JWT secret (invalidates all sessions)
# Update DATABASE_PASSWORD and JWT_SECRET in .env
sudo systemctl restart catalogizer

# 3. External API keys (TMDB, OMDB, etc.)
# Revoke and regenerate keys from each provider's dashboard
```

### Long-Term Containment

```bash
# Preserve evidence before cleanup
# Create forensic copies of logs and database
mkdir -p /opt/catalogizer/incident/$(date +%Y%m%d)
journalctl -u catalogizer --since "24 hours ago" > /opt/catalogizer/incident/$(date +%Y%m%d)/service.log
cp catalog-api/data/catalogizer.db /opt/catalogizer/incident/$(date +%Y%m%d)/database.db 2>/dev/null

# Export auth audit log
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c \
  "\COPY auth_audit_log TO '/opt/catalogizer/incident/$(date +%Y%m%d)/audit.csv' CSV HEADER;"
```

---

## Phase 3: Eradication

### Root Cause Analysis

1. **Review the auth_audit_log** for the full timeline of malicious actions
2. **Analyze application logs** for injection attempts, unusual request patterns
3. **Check for persistence mechanisms**: unauthorized user accounts, modified configurations
4. **Review database for tampering**: unexpected data modifications

```bash
# Check for unauthorized admin accounts
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT id, username, email, role_id, created_at, last_login_at
FROM users WHERE role_id = 1
ORDER BY created_at;"

# Check for modified configurations
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT * FROM roles WHERE is_system = FALSE;"

# Check storage roots for unauthorized additions
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT id, name, protocol, host, created_at FROM storage_roots ORDER BY created_at DESC;"
```

### Remediation Actions

1. Remove unauthorized accounts and sessions
2. Restore modified data from clean backups
3. Patch the vulnerability that was exploited
4. Update all credentials (database, JWT secret, API keys, admin password)
5. Run security scans to confirm no remaining vulnerabilities:

```bash
# Go vulnerability scan
cd catalog-api && govulncheck ./...

# Dependency audit
cd catalog-web && npm audit

# Static analysis
# semgrep --config auto catalog-api/
```

---

## Phase 4: Recovery

### Service Restoration

```bash
# 1. Verify all credentials have been rotated
# 2. Verify patches have been applied
# 3. Restart all services

sudo systemctl restart catalogizer

# 4. Verify health
curl -s http://localhost:8080/health | jq

# 5. Verify authentication works
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"new_password"}'

# 6. Run the challenge suite to verify system integrity
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"new_password"}' | jq -r '.token')

curl -s -X POST http://localhost:8080/api/v1/challenges/run \
  -H "Authorization: Bearer $TOKEN"
```

### Enhanced Monitoring

After recovery, temporarily increase monitoring sensitivity:

```bash
# Enable debug logging
echo "GIN_MODE=debug" >> catalog-api/.env
sudo systemctl restart catalogizer

# Monitor in real-time
journalctl -u catalogizer -f | grep -i -E "auth|login|error|denied"

# Watch Prometheus metrics for anomalies
curl -s http://localhost:8080/metrics | grep -E "http_requests|auth_"
```

---

## Phase 5: Post-Incident Review

### Review Meeting Agenda

Schedule within 72 hours of incident resolution. Document the following:

1. **Timeline**: Exact sequence of events from detection to resolution
2. **Root Cause**: What vulnerability or weakness was exploited
3. **Impact Assessment**: Data exposed, services disrupted, users affected
4. **Response Evaluation**: What went well, what could be improved
5. **Action Items**: Specific improvements with owners and deadlines

### Post-Incident Report Template

```
Incident Report: [INCIDENT-YYYY-MM-DD-NNN]
Date: [Date]
Severity: [P1/P2/P3/P4]
Status: [Resolved/Monitoring]

1. Summary
   [One-paragraph summary of the incident]

2. Timeline
   [Timestamp] - [Event description]
   [Timestamp] - [Event description]
   ...

3. Root Cause
   [Description of the vulnerability or failure]

4. Impact
   - Users affected: [count]
   - Data exposed: [description]
   - Service downtime: [duration]

5. Containment Actions
   [List of actions taken]

6. Remediation
   [Patches applied, credentials rotated, etc.]

7. Lessons Learned
   [What went well and what needs improvement]

8. Action Items
   - [ ] [Action] - Owner: [name] - Due: [date]
   - [ ] [Action] - Owner: [name] - Due: [date]
```

### Preventive Measures Checklist

- [ ] Update security scanning schedules
- [ ] Review and strengthen rate limiting configuration
- [ ] Audit all user accounts and permissions
- [ ] Verify backup integrity and test restore procedures
- [ ] Update this incident response plan based on lessons learned
- [ ] Conduct security awareness training if insider threat was involved
- [ ] Review and update firewall rules
- [ ] Schedule regular penetration testing

---

## Incident Types and Playbooks

### Playbook: JWT Token Compromise

1. Rotate `JWT_SECRET` in `.env` (invalidates all tokens)
2. Restart the service
3. Force all users to re-authenticate
4. Investigate how the token was obtained (logs, XSS, etc.)
5. Patch the leak vector

### Playbook: SQL Injection Attempt

1. Review application logs for the injection payload
2. Verify the input validation middleware caught it
3. If it bypassed validation, identify the vulnerable endpoint
4. The dialect layer uses parameterized queries (`?` placeholders); verify no raw string concatenation exists
5. Run Semgrep with SQL injection rules on the codebase

### Playbook: Brute Force Attack

1. Check `auth_audit_log` for repeated failed login attempts
2. Verify rate limiting is active (5 requests/min on auth endpoints)
3. Lock affected accounts if not already locked (auto-lock after threshold)
4. Block the attacking IP
5. Consider implementing CAPTCHA for future prevention

### Playbook: Dependency Vulnerability

1. Run `govulncheck ./...` (Go) and `npm audit` (TypeScript)
2. Assess whether the vulnerability is exploitable in Catalogizer's usage
3. Update the affected dependency: `go get <module>@latest` or `npm update <package>`
4. Run the full test suite to verify no regressions
5. Deploy the patched version

---

## Communication Templates

### Internal Notification (P1/P2)

```
Subject: [SECURITY] P[1/2] Incident Detected - [Brief Description]

A security incident has been detected in the Catalogizer deployment.

Severity: P[1/2]
Detected: [timestamp]
Status: [Investigating/Containing/Resolved]
Impact: [Brief impact description]

Incident Commander: [name]
Next update: [time]

Do not discuss details outside the incident response team.
```

### User Notification (if data exposed)

```
Subject: Security Notification - Action Required

We detected unauthorized access to the Catalogizer system on [date].
As a precaution, we have:

- Reset all active sessions (you will need to log in again)
- Rotated security credentials

Recommended actions:
1. Change your Catalogizer password
2. If you used the same password elsewhere, change those passwords too

We are investigating the scope of this incident and will provide updates.
For questions, contact: [email]
```

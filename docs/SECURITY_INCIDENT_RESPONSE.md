# Security Incident Response Plan

**Document Version:** 1.0  
**Last Updated:** April 6, 2026  
**Classification:** Internal Use Only  
**Owner:** Security Team  
**Review Cycle:** Quarterly  

---

## Table of Contents

1. [Purpose & Scope](#1-purpose--scope)
2. [Incident Classification](#2-incident-classification)
3. [Response Team (CSIRT)](#3-response-team-csirt)
4. [Detection & Reporting](#4-detection--reporting)
5. [Response Procedures](#5-response-procedures)
6. [Containment Procedures](#6-containment-procedures)
7. [Eradication Procedures](#7-eradication-procedures)
8. [Recovery Procedures](#8-recovery-procedures)
9. [Post-Incident Analysis](#9-post-incident-analysis)
10. [Communication Templates](#10-communication-templates)

---

## 1. Purpose & Scope

### 1.1 Purpose

This Security Incident Response Plan (SIRP) establishes procedures for detecting, responding to, and recovering from security incidents affecting the Catalogizer platform and its users.

### 1.2 Scope

This plan covers:
- Application security (catalog-api, catalog-web, mobile apps)
- Infrastructure security (servers, containers, networks)
- Data security (user data, media metadata)
- Third-party integrations
- Physical security (if applicable)

### 1.3 Compliance

This plan aligns with:
- GDPR Article 33 (Breach Notification)
- CCPA Section 1798.82
- ISO 27035 (Incident Management)
- NIST Cybersecurity Framework

---

## 2. Incident Classification

### 2.1 Severity Levels

| Level | Description | Examples | Response Time |
|-------|-------------|----------|---------------|
| **Critical** | Active exploitation, massive data breach | Ransomware, complete database compromise | Immediate (15 min) |
| **High** | Confirmed breach, significant impact | Unauthorized admin access, API key leak | 1 hour |
| **Medium** | Potential breach, limited impact | Suspicious login attempts, minor vulnerability | 4 hours |
| **Low** | Policy violation, no immediate risk | Misconfiguration, failed compliance check | 24 hours |

### 2.2 Incident Categories

#### Data Breach
- Unauthorized access to user data
- Accidental data exposure
- Insider threat

#### Malware/Infection
- Ransomware
- Cryptominers
- Trojans/Backdoors

#### Service Disruption
- DDoS attack
- Resource exhaustion
- Sabotage

#### Vulnerability Exploitation
- SQL injection
- XSS/CSRF attacks
- Authentication bypass

#### Insider Threat
- Data theft
- Privilege abuse
- Intentional damage

---

## 3. Response Team (CSIRT)

### 3.1 Team Structure

```
CSIRT Commander (CISO/CTO)
    ├── Technical Lead (Senior Security Engineer)
    │       ├── Incident Handlers (2-3)
    │       ├── Forensics Specialist
    │       └── Malware Analyst
    ├── Communications Lead (PR/Legal)
    │       ├── Internal Communications
    │       └── External Communications
    └── Business Lead (Product/Operations)
            ├── Business Impact Assessment
            └── Recovery Coordination
```

### 3.2 Roles & Responsibilities

| Role | Responsibilities | Primary | Escalation |
|------|-----------------|---------|------------|
| **CSIRT Commander** | Overall command, external communication | CISO | CTO → CEO |
| **Technical Lead** | Technical investigation, containment | SecEng Lead | Engineering Manager |
| **Incident Handler** | Incident execution, documentation | Security Engineers | Technical Lead |
| **Forensics Specialist** | Evidence collection, analysis | Senior SecEng | External consultant |
| **Legal Counsel** | Regulatory compliance, legal advice | General Counsel | External law firm |
| **Communications Lead** | PR management, user communication | PR Manager | CMO |

### 3.3 Contact Information

| Role | Name | Primary Contact | Secondary Contact |
|------|------|-----------------|-------------------|
| CSIRT Commander | [Name] | [Phone] | [Email] |
| Technical Lead | [Name] | [Phone] | [Email] |
| Legal Counsel | [Name] | [Phone] | [Email] |
| External Forensics | [Firm] | [24/7 Hotline] | [Email] |

---

## 4. Detection & Reporting

### 4.1 Detection Sources

| Source | Tool/Method | Detection Type |
|--------|-------------|----------------|
| Security Scanning | Snyk, Trivy, Gosec | Vulnerabilities |
| Intrusion Detection | WAF, IDS | Attacks, anomalies |
| Log Analysis | ELK Stack, Grafana | Suspicious activity |
| User Reports | Support tickets | Account issues |
| Threat Intelligence | External feeds | Known threats |
| Monitoring | Prometheus, Alertmanager | Anomalies |

### 4.2 Automated Alerting

```yaml
# Critical Security Alerts
alerts:
  - name: UnauthorizedAdminAccess
    condition: admin_login_from_unknown_ip
    severity: critical
    notify: [security-team, on-call]
    
  - name: DatabaseUnusualQuery
    condition: query_volume > 1000% baseline
    severity: high
    notify: [security-team, dba]
    
  - name: APIKeyAbuse
    condition: api_requests > 10000/min from single key
    severity: high
    notify: [security-team, api-team]
    
  - name: PrivilegeEscalation
    condition: user_role_changed_without_ticket
    severity: critical
    notify: [security-team, management]
```

### 4.3 Incident Reporting

#### Internal Reporting

```
To: security@catalogizer.local
Subject: [SECURITY INCIDENT] - Brief Description

Required Information:
1. Date/Time of detection
2. Reporter name/contact
3. Affected systems/data
4. Description of incident
5. Evidence (logs, screenshots)
6. Current status (ongoing/contained)
```

#### External Reporting

| Entity | Timing | Method | Required Info |
|--------|--------|--------|---------------|
| GDPR Authority | Within 72 hours | Online form | Nature, data subjects, measures |
| Affected Users | Without undue delay | Email/Site notice | What, when, steps to take |
| Law Enforcement | If criminal activity | Police report | Evidence, timeline |
| Cyber Insurance | Per policy terms | Phone/Email | Incident summary, damages |

---

## 5. Response Procedures

### 5.1 Initial Response (First 15 Minutes)

```
□ 1. Alert received and acknowledged
□ 2. CSIRT Commander notified
□ 3. Initial severity assessment
□ 4. Incident handler assigned
□ 5. Incident ticket created
□ 6. Evidence preservation started
□ 7. Log collection initiated
```

### 5.2 Investigation Phase

```bash
#!/bin/bash
# Incident investigation checklist

INCIDENT_ID="INC-$(date +%Y%m%d-%H%M%S)"
EVIDENCE_DIR="/incidents/$INCIDENT_ID"
mkdir -p "$EVIDENCE_DIR"

# 1. Collect system logs
echo "Collecting system logs..."
journalctl --since "1 hour ago" > "$EVIDENCE_DIR/system.log"

# 2. Collect application logs
echo "Collecting application logs..."
cp /var/log/catalogizer/*.log "$EVIDENCE_DIR/"

# 3. Collect database logs
echo "Collecting database logs..."
cp /var/log/postgresql/*.log "$EVIDENCE_DIR/"

# 4. Network connections
echo "Capturing network state..."
netstat -tunap > "$EVIDENCE_DIR/network.txt"

# 5. Running processes
echo "Capturing process list..."
ps aux > "$EVIDENCE_DIR/processes.txt"

# 6. Container state
echo "Capturing container state..."
podman ps -a > "$EVIDENCE_DIR/containers.txt"

# 7. Create evidence hash
echo "Creating evidence integrity hash..."
find "$EVIDENCE_DIR" -type f -exec sha256sum {} \; > "$EVIDENCE_DIR/hashes.txt"

echo "Evidence collected in $EVIDENCE_DIR"
```

### 5.3 Severity Assessment Matrix

| Factor | Low (1) | Medium (2) | High (3) | Critical (4) |
|--------|---------|------------|----------|--------------|
| **Data Exposure** | No data | Public data | User PII | Financial/medical |
| **System Impact** | Single service | Multiple services | Complete outage | Infrastructure |
| **User Impact** | < 100 users | 100-1000 users | 1000-10000 | > 10000 users |
| **Active Exploitation** | None | Attempted | Confirmed | Widespread |
| **Public Awareness** | None | Internal only | Partners notified | Media coverage |

**Severity = Sum of factors**
- 5-8: Low
- 9-12: Medium
- 13-16: High
- 17-20: Critical

---

## 6. Containment Procedures

### 6.1 Short-term Containment

#### Isolate Affected System

```bash
# 1. Isolate container
podman network disconnect catalogizer-network affected-container

# 2. Create firewall rule to block traffic
iptables -A INPUT -s <attacker_ip> -j DROP

# 3. Disable compromised account
psql -c "ALTER USER compromised_user WITH LOGIN NOCREATEUSER NOCREATEDB;"

# 4. Revoke API keys
curl -X POST http://localhost:8080/api/v1/admin/revoke-key \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"key_id": "compromised-key-id"}'
```

#### Preserve Evidence

```bash
# Create forensic snapshot
podman commit affected-container incident-evidence:$INCIDENT_ID
podman save -o "$EVIDENCE_DIR/container-snapshot.tar" incident-evidence:$INCIDENT_ID

# Snapshot database
pg_dump catalogizer > "$EVIDENCE_DIR/database-snapshot.sql"

# Capture memory dump (if applicable)
gcore -o "$EVIDENCE_DIR/process-dump" <pid>
```

### 6.2 Long-term Containment

```bash
# Deploy isolated analysis environment
podman-compose -f docker-compose.isolated.yml up -d

# Redirect traffic to maintenance page
kubectl patch ingress catalogizer-web -p '{"spec":{"rules":[{"http":{"paths":[{"path":"/","pathType":"Prefix","backend":{"service":{"name":"maintenance-page","port":{"number":80}}}}}]}}]}}'

# Enable enhanced monitoring
./scripts/enable-forensic-logging.sh
```

---

## 7. Eradication Procedures

### 7.1 Malware Removal

```bash
# 1. Stop affected services
podman-compose stop

# 2. Scan for malware
clamscan -r /opt/catalogizer --log=/var/log/clamav-scan.log

# 3. Remove infected files
# (Based on scan results)
rm -f /opt/catalogizer/infected-file

# 4. Rebuild clean containers
podman-compose build --no-cache

# 5. Restore from known-good backup
./scripts/restore-clean-state.sh --date=2026-04-05
```

### 7.2 Vulnerability Remediation

```bash
# 1. Identify vulnerable component
cd catalog-api
govulncheck ./...

# 2. Update dependencies
go get -u ./...
go mod tidy

# 3. Rebuild and test
go build -o catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2

# 4. Deploy patched version
podman-compose up -d --build api
```

### 7.3 Account Compromise

```sql
-- Reset compromised passwords
UPDATE users SET password_hash = NULL, password_reset_required = true 
WHERE id IN (SELECT user_id FROM compromised_accounts);

-- Invalidate all sessions
DELETE FROM user_sessions WHERE user_id IN (SELECT user_id FROM compromised_accounts);

-- Force 2FA setup
UPDATE users SET two_factor_required = true 
WHERE role IN ('admin', 'moderator');
```

---

## 8. Recovery Procedures

### 8.1 System Restoration

```bash
# 1. Verify threat elimination
./scripts/verify-system-clean.sh

# 2. Restore from backup (if needed)
./scripts/restore-database.sh --verify

# 3. Restart services
podman-compose up -d

# 4. Verify functionality
./scripts/smoke-tests.sh

# 5. Monitor for recurrence
./scripts/enhanced-monitoring.sh --duration=48h
```

### 8.2 Service Restoration Phases

| Phase | Services | Verification |
|-------|----------|--------------|
| 1 | Database, Core API | Health checks pass |
| 2 | Authentication, Web UI | Login/logout works |
| 3 | Media operations | Browse, search, play |
| 4 | Full functionality | All features tested |

### 8.3 User Communication

```markdown
**Security Incident Update**

We have resolved the security incident affecting Catalogizer. 

**What happened:** [Brief description]
**What we did:** [Remediation steps]
**What you should do:** [User actions required]

Your security is our priority. We sincerely apologize for any inconvenience.
```

---

## 9. Post-Incident Analysis

### 9.1 Timeline Documentation

```yaml
incident: INC-2026-0406-001
timeline:
  - time: "2026-04-06 14:30:00"
    event: "Alert triggered - unusual database activity"
    actor: "Automated monitoring"
    
  - time: "2026-04-06 14:32:00"
    event: "Incident acknowledged"
    actor: "Security Engineer"
    
  - time: "2026-04-06 14:35:00"
    event: "CSIRT activated"
    actor: "CSIRT Commander"
    
  - time: "2026-04-06 14:45:00"
    event: "Affected systems isolated"
    actor: "Incident Handler"
    
  - time: "2026-04-06 15:30:00"
    event: "Threat eliminated"
    actor: "Incident Handler"
    
  - time: "2026-04-06 16:00:00"
    event: "Services restored"
    actor: "Operations Team"
    
  - time: "2026-04-06 17:00:00"
    event: "Post-incident review completed"
    actor: "CSIRT Commander"
```

### 9.2 Root Cause Analysis

**5 Whys Template:**

1. **Why did the incident occur?**
   - Unauthorized access gained via SQL injection

2. **Why was SQL injection possible?**
   - Input validation missing on search endpoint

3. **Why was validation missing?**
   - Code review did not catch the vulnerability

4. **Why wasn't it caught in review?**
   - Security check not included in review checklist

5. **Why wasn't security check included?**
   - Review process didn't mandate security verification

**Root Cause:** Security review process gap

### 9.3 Lessons Learned

| Lesson | Action Item | Owner | Due Date |
|--------|-------------|-------|----------|
| Security reviews incomplete | Update review checklist | SecEng Lead | +1 week |
| Detection too slow | Tune alert thresholds | DevOps | +2 weeks |
| Response delays | Conduct DR drill | CSIRT | +1 month |

---

## 10. Communication Templates

### 10.1 Initial Internal Notification

```
Subject: [SECURITY] Incident INC-XXXX - Initial Report

INCIDENT SUMMARY:
- ID: INC-YYYY-MM-DD-NNN
- Detected: [Timestamp]
- Severity: [Critical/High/Medium/Low]
- Status: [Investigating/Contained/Resolved]

DESCRIPTION:
[Brief description of the incident]

IMMEDIATE ACTIONS:
1. [Action 1]
2. [Action 2]

NEXT UPDATE: [Time]

CSIRT Commander: [Name]
Technical Lead: [Name]
```

### 10.2 User Notification

```
Subject: Important Security Update - Catalogizer

Dear Catalogizer User,

We are writing to inform you of a security incident that may have affected your account.

WHAT HAPPENED:
On [Date], we detected unauthorized access to our systems. We immediately took action to secure our platform and investigate the incident.

WHAT INFORMATION WAS INVOLVED:
[List affected data types]

WHAT WE ARE DOING:
- Secured the vulnerability
- Engaged security experts
- Notified relevant authorities
- Enhanced monitoring

WHAT YOU SHOULD DO:
- Change your password
- Enable two-factor authentication
- Review account activity
- Report suspicious activity

We sincerely apologize for this incident and any inconvenience it may cause.

For questions: security@catalogizer.local

Catalogizer Security Team
```

### 10.3 Regulatory Notification (GDPR)

```
To: [Supervisory Authority]
Subject: Data Breach Notification - Article 33

1. Nature of breach: [Description]
2. Categories of data: [Personal data types]
3. Approximate number of affected individuals: [Number]
4. Likely consequences: [Potential impact]
5. Measures taken: [Remediation steps]
6. Contact details: [DPO contact]
```

---

## Appendix A: Incident Response Checklist

### Immediate Response (0-1 hour)
- [ ] Alert acknowledged
- [ ] CSIRT activated
- [ ] Evidence preserved
- [ ] Initial containment
- [ ] Legal counsel notified (if needed)

### Investigation (1-4 hours)
- [ ] Scope determined
- [ ] Root cause identified
- [ ] Impact assessed
- [ ] Evidence collected

### Containment (2-8 hours)
- [ ] Systems isolated
- [ ] Threat contained
- [ ] Chain of custody established

### Eradication (4-24 hours)
- [ ] Malware removed
- [ ] Vulnerabilities patched
- [ ] Compromised accounts secured

### Recovery (8-48 hours)
- [ ] Systems restored
- [ ] Services verified
- [ ] Monitoring enhanced

### Post-Incident (48+ hours)
- [ ] Report completed
- [ ] Lessons learned documented
- [ ] Improvements implemented

---

**Document Control:**
- Version: 1.0
- Classification: Internal Use Only
- Approved by: [CISO]
- Date approved: April 6, 2026
- Next review: July 6, 2026


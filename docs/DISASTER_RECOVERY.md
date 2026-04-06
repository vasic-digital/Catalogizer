# Disaster Recovery Plan

**Document Version:** 1.0  
**Last Updated:** April 6, 2026  
**Owner:** DevOps Team  
**Review Cycle:** Quarterly  

---

## Table of Contents

1. [Overview & Objectives](#1-overview--objectives)
2. [Disaster Types & Scenarios](#2-disaster-types--scenarios)
3. [Recovery Team & Responsibilities](#3-recovery-team--responsibilities)
4. [Backup Procedures](#4-backup-procedures)
5. [Recovery Procedures](#5-recovery-procedures)
6. [Recovery Time Objectives (RTO)](#6-recovery-time-objectives-rto)
7. [Recovery Point Objectives (RPO)](#7-recovery-point-objectives-rpo)
8. [Testing & Validation](#8-testing--validation)
9. [Communication Plan](#9-communication-plan)

---

## 1. Overview & Objectives

### 1.1 Purpose

This Disaster Recovery Plan (DRP) provides a comprehensive framework for responding to and recovering from disruptive events that impact the Catalogizer media management system. The plan ensures business continuity and minimizes data loss.

### 1.2 Scope

This plan covers:
- **catalog-api** (Go backend API)
- **catalog-web** (React frontend)
- **catalogizer-android** (Android mobile app)
- **catalogizer-androidtv** (Android TV app)
- **catalogizer-desktop** (Tauri desktop app)
- **Database systems** (SQLite/PostgreSQL)
- **Supporting infrastructure** (Redis, monitoring, etc.)

### 1.3 Objectives

| Objective | Target |
|-----------|--------|
| Maximum Acceptable Downtime | 4 hours (RTO) |
| Maximum Data Loss | 15 minutes (RPO) |
| Recovery Success Rate | 99.9% |
| DR Testing Frequency | Quarterly |

---

## 2. Disaster Types & Scenarios

### 2.1 Disaster Classification

| Level | Description | Examples |
|-------|-------------|----------|
| **Critical** | Complete system failure, data loss | Database corruption, total infrastructure failure |
| **High** | Major component failure | API cluster down, storage failure |
| **Medium** | Partial functionality loss | Single node failure, network partition |
| **Low** | Minor impact | Performance degradation, non-critical service down |

### 2.2 Specific Scenarios

#### Scenario 1: Database Corruption
**Impact:** High - All media metadata inaccessible  
**Detection:** Database health checks, error logs  
**Response Time:** Immediate

#### Scenario 2: Complete Infrastructure Failure
**Impact:** Critical - Entire system unavailable  
**Detection:** Monitoring alerts, user reports  
**Response Time:** Immediate

#### Scenario 3: Data Center Outage
**Impact:** Critical - Multi-region failure  
**Detection:** Infrastructure monitoring  
**Response Time:** 15 minutes

#### Scenario 4: Ransomware/Malware Attack
**Impact:** Critical - Data integrity compromised  
**Detection:** Security monitoring, unusual file activity  
**Response Time:** Immediate

#### Scenario 5: Accidental Data Deletion
**Impact:** High - User or system data lost  
**Detection:** Audit logs, user reports  
**Response Time:** 1 hour

---

## 3. Recovery Team & Responsibilities

### 3.1 Team Structure

| Role | Responsibility | Primary | Secondary |
|------|---------------|---------|-----------|
| **DR Coordinator** | Overall DR process management | DevOps Lead | Engineering Manager |
| **Technical Lead** | Technical recovery decisions | Senior Backend Dev | Senior DevOps |
| **Database Admin** | Database recovery | DBA | Senior Backend Dev |
| **Infrastructure** | Infrastructure restoration | DevOps Engineer | SRE |
| **Communications** | Stakeholder communication | Product Manager | Engineering Manager |
| **QA Validation** | Post-recovery testing | QA Lead | Senior QA |

### 3.2 Escalation Matrix

| Time Elapsed | Action | Notify |
|--------------|--------|--------|
| 0-15 min | Initial assessment | DR Coordinator, Technical Lead |
| 15-30 min | DR activation decision | Engineering Manager, CTO |
| 30-60 min | External communication | Product Manager, Support Lead |
| 1-4 hours | Hourly status updates | All stakeholders |
| 4+ hours | Executive briefing | CEO, CTO, VP Engineering |

---

## 4. Backup Procedures

### 4.1 Database Backups

#### PostgreSQL (Production)

```bash
# Automated backup script
#!/bin/bash
# /opt/catalogizer/scripts/backup-database.sh

BACKUP_DIR="/backup/postgres/$(date +%Y%m%d)"
RETENTION_DAYS=30
DB_NAME="catalogizer"
DB_USER="postgres"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Full database dump
pg_dump -h localhost -U "$DB_USER" -Fc "$DB_NAME" > "$BACKUP_DIR/catalogizer_$(date +%H%M%S).dump"

# Verify backup
if pg_restore --list "$BACKUP_DIR"/catalogizer_*.dump > /dev/null 2>&1; then
    echo "Backup verified: $(date)" >> /var/log/catalogizer/backup.log
else
    echo "BACKUP FAILED: $(date)" | mail -s "CRITICAL: Backup Failure" alerts@catalogizer.local
fi

# Clean old backups
find /backup/postgres -type d -mtime +$RETENTION_DAYS -exec rm -rf {} \;
```

**Schedule:**
- Full backup: Daily at 02:00 UTC
- Incremental backup: Every 4 hours
- WAL archiving: Continuous

**Retention:**
- Daily backups: 30 days
- Weekly backups: 12 weeks
- Monthly backups: 12 months

#### SQLite (Development/Single-user)

```bash
# SQLite backup
sqlite3 catalogizer.db ".backup '/backup/sqlite/catalogizer_$(date +%Y%m%d_%H%M%S).db'"
```

### 4.2 File System Backups

#### Media Storage

```bash
# Rsync-based backup
rsync -avz --delete \
    /media/storage/ \
    backup-server:/backup/media/storage/ \
    --log-file=/var/log/catalogizer/media-backup.log
```

**Schedule:**
- Incremental: Every 6 hours
- Full sync: Weekly

### 4.3 Configuration Backups

```bash
# Configuration backup
CONFIG_BACKUP="/backup/config/$(date +%Y%m%d)"
mkdir -p "$CONFIG_BACKUP"

cp -r /etc/catalogizer/* "$CONFIG_BACKUP/"
cp /opt/catalogizer/.env "$CONFIG_BACKUP/"
cp /opt/catalogizer/config.json "$CONFIG_BACKUP/"

# Docker Compose files
cp /opt/catalogizer/docker-compose*.yml "$CONFIG_BACKUP/"

# Kubernetes manifests
cp -r /opt/catalogizer/k8s/* "$CONFIG_BACKUP/"
```

### 4.4 Backup Verification

```bash
# Monthly backup verification script
#!/bin/bash

# Test database restore
gpg --decrypt /backup/postgres/latest.dump.gpg | pg_restore --list > /dev/null

# Test file restore
rsync --dry-run backup-server:/backup/media/storage/ /tmp/restore-test/

# Generate verification report
echo "Backup verification completed: $(date)" >> /var/log/catalogizer/backup-verify.log
```

---

## 5. Recovery Procedures

### 5.1 Database Recovery

#### PostgreSQL Point-in-Time Recovery

```bash
# 1. Stop application
podman-compose -f docker-compose.yml stop api

# 2. Restore from backup
pg_restore -h localhost -U postgres -d catalogizer --clean --create \
    /backup/postgres/20250406/catalogizer_020000.dump

# 3. Replay WAL logs (for PITR)
pg_waldump /var/lib/postgresql/wal/ > /dev/null

# 4. Verify database
psql -h localhost -U postgres -d catalogizer -c "SELECT COUNT(*) FROM media_items;"

# 5. Restart application
podman-compose -f docker-compose.yml start api
```

#### SQLite Recovery

```bash
# Simple file restore
cp /backup/sqlite/catalogizer_20250406_020000.db catalogizer.db

# Verify
sqlite3 catalogizer.db "PRAGMA integrity_check;"
```

### 5.2 Complete System Recovery

```bash
#!/bin/bash
# Complete disaster recovery script

set -e

RECOVERY_LOG="/var/log/catalogizer/recovery-$(date +%Y%m%d-%H%M%S).log"
exec > >(tee -a "$RECOVERY_LOG")
exec 2>&1

echo "=== Catalogizer Disaster Recovery Started: $(date) ==="

# Step 1: Infrastructure setup
echo "[1/7] Setting up infrastructure..."
podman-compose -f docker-compose.yml pull

# Step 2: Database recovery
echo "[2/7] Restoring database..."
./scripts/restore-database.sh

# Step 3: Media storage mount
echo "[3/7] Mounting media storage..."
mount -t nfs nas.local:/volume1/media /media/storage

# Step 4: Configuration restore
echo "[4/7] Restoring configuration..."
cp /backup/config/latest/.env /opt/catalogizer/
cp /backup/config/latest/config.json /opt/catalogizer/

# Step 5: Start services
echo "[5/7] Starting services..."
podman-compose -f docker-compose.yml up -d

# Step 6: Health checks
echo "[6/7] Running health checks..."
sleep 30
./scripts/health-check.sh

# Step 7: Validation
echo "[7/7] Validating recovery..."
./scripts/validate-recovery.sh

echo "=== Disaster Recovery Completed: $(date) ==="
```

### 5.3 Kubernetes Recovery

```bash
# Restore from manifests
kubectl apply -f /backup/config/latest/k8s/

# Verify pods
kubectl wait --for=condition=ready pod -l app=catalogizer-api --timeout=300s

# Check services
kubectl get svc

# Verify ingress
kubectl get ingress
```

### 5.4 Application-Specific Recovery

#### Mobile Apps (Android/Android TV)

- App binaries: Restore from CI/CD artifacts
- Signing keys: Retrieve from secure vault
- App Store: Resubmit if necessary

#### Desktop App (Tauri)

```bash
# Rebuild from source
git clone https://github.com/catalogizer/catalogizer.git
cd catalogizer/catalogizer-desktop
npm install
npm run tauri:build
```

---

## 6. Recovery Time Objectives (RTO)

### 6.1 Component RTOs

| Component | RTO | Recovery Method |
|-----------|-----|-----------------|
| API Service | 30 minutes | Container restart/redployment |
| Web Frontend | 30 minutes | CDN/cache refresh |
| Database | 2 hours | Point-in-time restore |
| Media Storage | 4 hours | Failover to replica |
| Monitoring | 1 hour | Container restart |
| Mobile Apps | N/A | Client-side, no server impact |

### 6.2 Service Level Tiers

| Tier | Services | RTO |
|------|----------|-----|
| **Tier 1** | API, Database, Auth | 1 hour |
| **Tier 2** | Web UI, Caching | 2 hours |
| **Tier 3** | Analytics, Reporting | 4 hours |
| **Tier 4** | Development, Testing | 24 hours |

---

## 7. Recovery Point Objectives (RPO)

### 7.1 Data Loss Tolerance

| Data Type | RPO | Backup Frequency |
|-----------|-----|------------------|
| User Data | 15 minutes | Continuous WAL + hourly backup |
| Media Metadata | 15 minutes | Continuous WAL + hourly backup |
| Configuration | 24 hours | Daily backup |
| Media Files | 6 hours | Incremental backup |
| Logs | 24 hours | Daily archive |

### 7.2 Zero Data Loss Strategy

For critical transactions:
```sql
-- Enable synchronous replication
ALTER SYSTEM SET synchronous_commit = 'remote_apply';
ALTER SYSTEM SET synchronous_standby_names = 'replica_1, replica_2';
```

---

## 8. Testing & Validation

### 8.1 DR Testing Schedule

| Test Type | Frequency | Scope |
|-----------|-----------|-------|
| Tabletop Exercise | Monthly | Process review |
| Backup Restoration | Weekly | Automated test |
| Full DR Drill | Quarterly | Complete failover |
| Chaos Engineering | Monthly | Random failures |

### 8.2 Automated DR Tests

```bash
#!/bin/bash
# Weekly DR validation

# Test backup integrity
pg_restore --list latest.dump > /dev/null || exit 1

# Test service startup
podman-compose up -d --scale api=0
sleep 5
podman-compose up -d --scale api=1

# Run smoke tests
curl -f http://localhost:8080/health || exit 1

echo "DR test passed: $(date)"
```

### 8.3 Post-Recovery Validation

```bash
#!/bin/bash
# Recovery validation checklist

echo "=== Post-Recovery Validation ==="

# 1. API health
curl -s http://localhost:8080/health | grep -q "healthy" && echo "✓ API healthy" || echo "✗ API unhealthy"

# 2. Database connectivity
psql -h localhost -U postgres -c "SELECT 1" > /dev/null && echo "✓ Database accessible" || echo "✗ Database inaccessible"

# 3. Media count verification
ACTUAL=$(psql -h localhost -U postgres -t -c "SELECT COUNT(*) FROM media_items")
EXPECTED=$(cat /backup/expected-counts.txt | grep media_items | cut -d: -f2)
[ "$ACTUAL" -eq "$EXPECTED" ] && echo "✓ Media count matches" || echo "⚠ Media count differs"

# 4. WebSocket functionality
./scripts/test-websocket.sh && echo "✓ WebSocket working" || echo "✗ WebSocket failed"

# 5. File access
ls /media/storage > /dev/null && echo "✓ Storage accessible" || echo "✗ Storage inaccessible"

echo "=== Validation Complete ==="
```

---

## 9. Communication Plan

### 9.1 Internal Communication

| Stakeholder | Notification Channel | Timing |
|-------------|---------------------|--------|
| Engineering Team | Slack #incidents | Immediate |
| Management | Email + Phone | 15 minutes |
| Executive Team | Phone bridge | 1 hour (if ongoing) |
| All Staff | Company-wide Slack | 30 minutes |

### 9.2 External Communication

| Audience | Channel | Timing | Owner |
|----------|---------|--------|-------|
| Customers | Status page | 30 minutes | Support Lead |
| Customers | Email | 1 hour (if >2hr outage) | Product Manager |
| Partners | Direct contact | 1 hour | Business Dev |
| Media | Press release | If newsworthy | Marketing |

### 9.3 Status Page Updates

```markdown
**Template: Incident Update**

**[Investigating]** Service Disruption - Catalogizer API

We are currently investigating reports of service disruption affecting the Catalogizer API. 

**Impact:** Media browsing and search are unavailable
**Started:** 2026-04-06 14:30 UTC
**Status:** Our engineering team is actively working on resolution

We will provide updates every 30 minutes until resolved.

---

**[Resolved]** Service Restored - Catalogizer API

All services have been restored and are operating normally.

**Duration:** 45 minutes
**Root Cause:** Database connection pool exhaustion
**Resolution:** Restarted database connections, increased pool size

We apologize for any inconvenience caused.
```

### 9.4 Communication Templates

See `docs/communication/` for:
- Initial incident notification
- Status update templates
- Resolution notification
- Post-incident report

---

## Appendix A: Emergency Contacts

| Role | Name | Phone | Email |
|------|------|-------|-------|
| DR Coordinator | [Name] | [Phone] | [Email] |
| Technical Lead | [Name] | [Phone] | [Email] |
| Infrastructure | [Name] | [Phone] | [Email] |
| Database Admin | [Name] | [Phone] | [Email] |
| Hosting Provider | [Provider] | [Support Line] | [Email] |
| Domain Registrar | [Registrar] | [Support Line] | [Email] |

## Appendix B: Recovery Resources

### Backup Locations

| Resource | Primary | Secondary | Tertiary |
|----------|---------|-----------|----------|
| Database | Local NAS | Cloud S3 | Off-site |
| Config | GitHub | GitLab | Local |
| Media | Local NAS | Cloud | Off-site |

### Critical Credentials

Stored in: `/secure/vault/credentials.yml` (encrypted)

### Network Information

| Component | IP/URL | Notes |
|-----------|--------|-------|
| Primary API | api.catalogizer.local | Load balancer |
| Database Master | db-master.local | PostgreSQL |
| Database Replica | db-replica.local | Read replica |
| Storage NAS | nas.local | NFS mount |
| Backup Server | backup.local | Rsync target |

---

**Document Control:**
- Version: 1.0
- Approved by: [Name]
- Date approved: April 6, 2026
- Next review: July 6, 2026


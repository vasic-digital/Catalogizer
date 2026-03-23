# Catalogizer Disaster Recovery Guide

This document defines backup procedures, restore steps, data verification, and backup scheduling for Catalogizer deployments.

---

## Table of Contents

1. [Backup Strategy Overview](#backup-strategy-overview)
2. [SQLite Backup Procedures](#sqlite-backup-procedures)
3. [PostgreSQL Backup Procedures](#postgresql-backup-procedures)
4. [Configuration Backup](#configuration-backup)
5. [Asset and Cache Backup](#asset-and-cache-backup)
6. [Restore Procedures](#restore-procedures)
7. [Data Verification](#data-verification)
8. [Backup Scheduling](#backup-scheduling)
9. [Disaster Recovery Scenarios](#disaster-recovery-scenarios)

---

## Backup Strategy Overview

Catalogizer stores data in three categories that must be backed up:

| Category | Location | Priority | Description |
|----------|----------|----------|-------------|
| Database | `catalog-api/data/catalogizer.db` (SQLite) or PostgreSQL server | Critical | All metadata, user accounts, media entities, scan history |
| Configuration | `catalog-api/.env`, `catalog-api/config.json`, `config/` | High | Server settings, API keys, nginx/redis configs |
| Assets | `catalog-api/cache/assets/`, `catalog-api/cache/cover_art/` | Medium | Cover art, thumbnails, cached images (regenerable) |

Media files themselves are stored on external storage roots (SMB, NFS, etc.) and are not managed by Catalogizer backups.

### Backup Frequency Recommendations

| Component | Frequency | Retention |
|-----------|-----------|-----------|
| Database (full) | Daily | 30 days |
| Database (incremental) | Hourly (PostgreSQL WAL) | 7 days |
| Configuration | On change | 90 days |
| Assets | Weekly | 14 days |

---

## SQLite Backup Procedures

### Online Backup (Recommended)

SQLite with WAL mode supports online backups without stopping the service:

```bash
#!/bin/bash
# backup-sqlite.sh -- Online SQLite backup

BACKUP_DIR="/opt/catalogizer/backups"
DB_PATH="/opt/catalogizer/catalog-api/data/catalogizer.db"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/catalogizer_${TIMESTAMP}.db"

mkdir -p "$BACKUP_DIR"

# Use SQLite .backup command for a consistent snapshot
sqlite3 "$DB_PATH" ".backup '${BACKUP_FILE}'"

# Compress the backup
gzip "$BACKUP_FILE"

echo "Backup created: ${BACKUP_FILE}.gz"
echo "Size: $(du -h "${BACKUP_FILE}.gz" | cut -f1)"
```

### Offline Backup

If the service can be stopped:

```bash
# Stop the service
sudo systemctl stop catalogizer

# Copy the database file and WAL/SHM files
cp catalog-api/data/catalogizer.db /opt/catalogizer/backups/catalogizer_$(date +%Y%m%d).db
cp catalog-api/data/catalogizer.db-wal /opt/catalogizer/backups/catalogizer_$(date +%Y%m%d).db-wal 2>/dev/null
cp catalog-api/data/catalogizer.db-shm /opt/catalogizer/backups/catalogizer_$(date +%Y%m%d).db-shm 2>/dev/null

# Restart the service
sudo systemctl start catalogizer
```

### Encrypted Database Backup

If SQLCipher encryption is enabled, the backup file is also encrypted. Store the `DB_ENCRYPTION_KEY` separately from the backup for security:

```bash
# The .backup command preserves encryption
sqlite3 "$DB_PATH" ".backup '${BACKUP_FILE}'"

# Store the encryption key in a separate, secured location
echo "$DB_ENCRYPTION_KEY" > /opt/catalogizer/keys/db_key_${TIMESTAMP}.txt
chmod 600 /opt/catalogizer/keys/db_key_${TIMESTAMP}.txt
```

---

## PostgreSQL Backup Procedures

### Logical Backup (pg_dump)

```bash
#!/bin/bash
# backup-postgres.sh -- PostgreSQL logical backup

BACKUP_DIR="/opt/catalogizer/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/catalogizer_pg_${TIMESTAMP}.dump"

mkdir -p "$BACKUP_DIR"

# Custom format backup (compressed, supports selective restore)
PGPASSWORD="$DATABASE_PASSWORD" pg_dump \
  -h "${DATABASE_HOST:-localhost}" \
  -p "${DATABASE_PORT:-5432}" \
  -U "${DATABASE_USER:-catalogizer_user}" \
  -d "${DATABASE_NAME:-catalogizer}" \
  -Fc \
  -f "$BACKUP_FILE"

echo "Backup created: ${BACKUP_FILE}"
echo "Size: $(du -h "${BACKUP_FILE}" | cut -f1)"
```

### SQL Text Backup

For portability and readability:

```bash
PGPASSWORD="$DATABASE_PASSWORD" pg_dump \
  -h "$DATABASE_HOST" \
  -U "$DATABASE_USER" \
  -d "$DATABASE_NAME" \
  --no-owner \
  --no-privileges \
  -f "${BACKUP_DIR}/catalogizer_pg_${TIMESTAMP}.sql"

gzip "${BACKUP_DIR}/catalogizer_pg_${TIMESTAMP}.sql"
```

### Containerized PostgreSQL Backup

```bash
# Using podman exec
podman exec catalogizer-db pg_dump \
  -U catalogizer_user \
  -d catalogizer \
  -Fc > "/opt/catalogizer/backups/catalogizer_pg_$(date +%Y%m%d_%H%M%S).dump"
```

### Continuous Archiving (WAL)

For point-in-time recovery, configure PostgreSQL WAL archiving:

```ini
# postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'cp %p /opt/catalogizer/backups/wal/%f'
```

---

## Configuration Backup

```bash
#!/bin/bash
# backup-config.sh -- Configuration backup

BACKUP_DIR="/opt/catalogizer/backups/config"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "${BACKUP_DIR}/${TIMESTAMP}"

# Back up configuration files
cp catalog-api/.env "${BACKUP_DIR}/${TIMESTAMP}/.env" 2>/dev/null
cp catalog-api/config.json "${BACKUP_DIR}/${TIMESTAMP}/config.json" 2>/dev/null
cp -r config/ "${BACKUP_DIR}/${TIMESTAMP}/config/" 2>/dev/null

# Back up docker-compose files
for f in docker-compose*.yml; do
  cp "$f" "${BACKUP_DIR}/${TIMESTAMP}/" 2>/dev/null
done

# Create archive
tar czf "${BACKUP_DIR}/config_${TIMESTAMP}.tar.gz" -C "${BACKUP_DIR}" "${TIMESTAMP}"
rm -rf "${BACKUP_DIR}/${TIMESTAMP}"

echo "Config backup: ${BACKUP_DIR}/config_${TIMESTAMP}.tar.gz"
```

---

## Asset and Cache Backup

Assets (cover art, thumbnails) are stored in the filesystem and can be regenerated from external providers. Back them up to avoid re-fetching:

```bash
#!/bin/bash
# backup-assets.sh -- Asset backup

BACKUP_DIR="/opt/catalogizer/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

tar czf "${BACKUP_DIR}/assets_${TIMESTAMP}.tar.gz" \
  catalog-api/cache/assets/ \
  catalog-api/cache/cover_art/ \
  catalog-api/cache/tls/ \
  2>/dev/null

echo "Asset backup: ${BACKUP_DIR}/assets_${TIMESTAMP}.tar.gz"
```

---

## Restore Procedures

### Restore SQLite Database

```bash
# Stop the service
sudo systemctl stop catalogizer

# Restore from backup
gunzip -k /opt/catalogizer/backups/catalogizer_20260323_020000.db.gz
cp /opt/catalogizer/backups/catalogizer_20260323_020000.db catalog-api/data/catalogizer.db

# Remove WAL and SHM files (will be recreated)
rm -f catalog-api/data/catalogizer.db-wal catalog-api/data/catalogizer.db-shm

# Start the service
sudo systemctl start catalogizer
```

### Restore PostgreSQL Database

```bash
# Drop and recreate the database
PGPASSWORD="$DATABASE_PASSWORD" dropdb -h "$DATABASE_HOST" -U "$DATABASE_USER" "$DATABASE_NAME"
PGPASSWORD="$DATABASE_PASSWORD" createdb -h "$DATABASE_HOST" -U "$DATABASE_USER" -O "$DATABASE_USER" "$DATABASE_NAME"

# Restore from custom format dump
PGPASSWORD="$DATABASE_PASSWORD" pg_restore \
  -h "$DATABASE_HOST" \
  -U "$DATABASE_USER" \
  -d "$DATABASE_NAME" \
  --no-owner \
  --no-privileges \
  /opt/catalogizer/backups/catalogizer_pg_20260323_020000.dump
```

### Restore from SQL Text Backup

```bash
gunzip -k /opt/catalogizer/backups/catalogizer_pg_20260323_020000.sql.gz

PGPASSWORD="$DATABASE_PASSWORD" psql \
  -h "$DATABASE_HOST" \
  -U "$DATABASE_USER" \
  -d "$DATABASE_NAME" \
  -f /opt/catalogizer/backups/catalogizer_pg_20260323_020000.sql
```

### Restore Configuration

```bash
tar xzf /opt/catalogizer/backups/config_20260323_020000.tar.gz -C /tmp/
cp /tmp/20260323_020000/.env catalog-api/.env
cp /tmp/20260323_020000/config.json catalog-api/config.json
cp -r /tmp/20260323_020000/config/ config/
```

### Restore Assets

```bash
tar xzf /opt/catalogizer/backups/assets_20260323_020000.tar.gz -C /opt/catalogizer/
```

---

## Data Verification

After restoring, verify data integrity:

### SQLite Verification

```bash
# Check database integrity
sqlite3 catalog-api/data/catalogizer.db "PRAGMA integrity_check;"

# Verify row counts
sqlite3 catalog-api/data/catalogizer.db "
SELECT 'storage_roots' as tbl, COUNT(*) FROM storage_roots
UNION ALL SELECT 'files', COUNT(*) FROM files
UNION ALL SELECT 'users', COUNT(*) FROM users
UNION ALL SELECT 'media_items', COUNT(*) FROM media_items
UNION ALL SELECT 'media_files', COUNT(*) FROM media_files
UNION ALL SELECT 'scan_history', COUNT(*) FROM scan_history;"

# Check foreign key integrity
sqlite3 catalog-api/data/catalogizer.db "PRAGMA foreign_key_check;"

# Verify migration state
sqlite3 catalog-api/data/catalogizer.db "SELECT * FROM migrations ORDER BY version;"
```

### PostgreSQL Verification

```bash
# Verify table access
PGPASSWORD="$DATABASE_PASSWORD" psql -h "$DATABASE_HOST" -U "$DATABASE_USER" -d "$DATABASE_NAME" -c "
SELECT 'storage_roots' as tbl, COUNT(*) FROM storage_roots
UNION ALL SELECT 'files', COUNT(*) FROM files
UNION ALL SELECT 'users', COUNT(*) FROM users
UNION ALL SELECT 'media_items', COUNT(*) FROM media_items
UNION ALL SELECT 'media_files', COUNT(*) FROM media_files
UNION ALL SELECT 'scan_history', COUNT(*) FROM scan_history;"

# Verify migration state
PGPASSWORD="$DATABASE_PASSWORD" psql -h "$DATABASE_HOST" -U "$DATABASE_USER" -d "$DATABASE_NAME" -c \
  "SELECT * FROM migrations ORDER BY version;"

# Check for constraint violations
PGPASSWORD="$DATABASE_PASSWORD" psql -h "$DATABASE_HOST" -U "$DATABASE_USER" -d "$DATABASE_NAME" -c "
SELECT conname, conrelid::regclass
FROM pg_constraint
WHERE contype = 'f'
  AND NOT convalidated;"
```

### Application-Level Verification

```bash
# Health check
curl -s http://localhost:8080/health | jq

# Login as admin
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# Verify storage roots exist
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/storage-roots | jq '.[] | .name'

# Verify media entity counts
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/entities/stats | jq

# Run the challenge suite
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/challenges/run
```

---

## Backup Scheduling

### Cron-Based Scheduling

```bash
# /etc/cron.d/catalogizer-backup

# Daily full database backup at 2:00 AM
0 2 * * * root /opt/catalogizer/scripts/backup-postgres.sh >> /var/log/catalogizer-backup.log 2>&1

# Hourly configuration backup
0 * * * * root /opt/catalogizer/scripts/backup-config.sh >> /var/log/catalogizer-backup.log 2>&1

# Weekly asset backup on Sundays at 3:00 AM
0 3 * * 0 root /opt/catalogizer/scripts/backup-assets.sh >> /var/log/catalogizer-backup.log 2>&1

# Daily cleanup of backups older than 30 days
0 4 * * * root find /opt/catalogizer/backups -name "*.dump" -mtime +30 -delete
0 4 * * * root find /opt/catalogizer/backups -name "*.db.gz" -mtime +30 -delete
```

### Systemd Timer (Alternative)

```ini
# /etc/systemd/system/catalogizer-backup.service
[Unit]
Description=Catalogizer Database Backup

[Service]
Type=oneshot
ExecStart=/opt/catalogizer/scripts/backup-postgres.sh
User=catalogizer

# /etc/systemd/system/catalogizer-backup.timer
[Unit]
Description=Run Catalogizer backup daily

[Timer]
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

```bash
sudo systemctl enable catalogizer-backup.timer
sudo systemctl start catalogizer-backup.timer
```

### Backup Rotation

Implement a retention policy to manage disk space:

```bash
#!/bin/bash
# rotate-backups.sh

BACKUP_DIR="/opt/catalogizer/backups"

# Keep daily backups for 30 days
find "$BACKUP_DIR" -name "catalogizer_pg_*.dump" -mtime +30 -delete
find "$BACKUP_DIR" -name "catalogizer_*.db.gz" -mtime +30 -delete

# Keep config backups for 90 days
find "$BACKUP_DIR" -name "config_*.tar.gz" -mtime +90 -delete

# Keep asset backups for 14 days
find "$BACKUP_DIR" -name "assets_*.tar.gz" -mtime +14 -delete

# Report remaining backups
echo "Remaining backups:"
du -sh "$BACKUP_DIR"
ls -lh "$BACKUP_DIR"
```

---

## Disaster Recovery Scenarios

### Scenario 1: Database Corruption

1. Stop the service: `sudo systemctl stop catalogizer`
2. Identify the most recent clean backup
3. Restore from backup (see Restore Procedures above)
4. Verify data integrity
5. Restart: `sudo systemctl start catalogizer`
6. Re-scan storage roots if the backup is stale: trigger scans via the API

### Scenario 2: Complete Server Loss

1. Provision a new server with the same OS
2. Install prerequisites: Go 1.24+, PostgreSQL 15+, Podman
3. Clone the repository: `git clone --recurse-submodules <repo-url>`
4. Restore configuration from backup
5. Restore database from backup
6. Restore assets from backup (optional -- they regenerate)
7. Build and start: `cd catalog-api && go build -o catalogizer && ./catalogizer`
8. Verify all endpoints

### Scenario 3: Storage Root Unavailable

If a storage root goes offline, Catalogizer continues to serve cached metadata. When the root comes back online:

1. Verify connectivity via `POST /api/v1/smb/test`
2. Trigger a rescan: `POST /api/v1/scans` with the storage root ID
3. The scanner detects added, modified, and deleted files

### Scenario 4: Accidental Data Deletion

If rows were accidentally deleted from the database:

1. Stop the service
2. Restore the database from the most recent backup
3. If the backup is older than the deletion, you may need to rescan storage roots to re-detect files
4. Media entities will be re-aggregated during the post-scan pipeline

### Scenario 5: Encryption Key Loss

If the SQLCipher `DB_ENCRYPTION_KEY` is lost:

1. The encrypted database file is unrecoverable
2. Restore from the most recent backup where the key is known
3. If no keyed backup exists, rescan all storage roots from scratch
4. User accounts and settings will need to be recreated

Prevention: Store the encryption key in at least two separate secure locations (e.g., password manager and sealed envelope).

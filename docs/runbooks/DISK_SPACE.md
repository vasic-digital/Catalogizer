# Runbook: Disk Space Issues

**Alert:** DiskSpaceLow / DiskSpaceCritical  
**Severity:** Warning (80%) / Critical (90%)  
**Service:** System  

---

## Summary

This alert fires when disk usage exceeds thresholds (80% warning, 90% critical).

---

## Immediate Assessment

```bash
# Check disk usage
df -h

# Find largest directories
du -h / | sort -rh | head -20

# Check inode usage
df -i
```

---

## Common Causes & Solutions

### Log Files

```bash
# Check log sizes
find /var/log -type f -size +100M -ls

# Rotate logs manually
logrotate -f /etc/logrotate.conf

# Clear old logs (be careful!)
find /var/log/catalogizer -name "*.log.*" -mtime +30 -delete
```

### Database

```bash
# Check PostgreSQL log size
ls -lh /var/log/postgresql/

# Check WAL archive size
du -sh /var/lib/postgresql/archive/

# Vacuum and clean up
psql -c "VACUUM FULL;"
```

### Application Data

```bash
# Check catalogizer data
du -sh /opt/catalogizer/data/

# Clean up temporary files
find /opt/catalogizer/temp -type f -mtime +7 -delete

# Clean up old releases
ls -lt /opt/catalogizer/releases/ | tail -n +6 | xargs rm -rf
```

### Docker/Podman

```bash
# Check container disk usage
podman system df

# Clean up unused containers
podman container prune -f

# Clean up unused images
podman image prune -a -f

# Clean up volumes
podman volume prune -f
```

---

## Emergency Cleanup

If critical (>90%):

```bash
# 1. Clear package cache
apt-get clean
yum clean all

# 2. Clear thumbnail cache
rm -rf ~/.cache/thumbnails/*

# 3. Clear old kernels (keep current + 1)
# (distribution-specific)

# 4. Clear logs older than 7 days
find /var/log -type f -name "*.log" -mtime +7 -delete

# 5. Truncate large active logs
: > /var/log/large-file.log
```

---

## Verification

```bash
# Check disk usage again
df -h

# Set up monitoring
watch -n 60 'df -h'
```

---

## Prevention

- Set up log rotation
- Monitor disk usage trends
- Set up automatic cleanup
- Use separate partitions for logs and data
- Set up alerts at 70%, 80%, 90%

---

*Last Updated: April 10, 2026*

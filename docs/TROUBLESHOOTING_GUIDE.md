# Catalogizer v3.0 - Troubleshooting and Maintenance Guide

## Table of Contents
1. [Overview](#overview)
2. [Quick Diagnostics](#quick-diagnostics)
3. [Common Issues](#common-issues)
4. [Service Management](#service-management)
5. [Database Issues](#database-issues)
6. [Storage and Media Issues](#storage-and-media-issues)
7. [Performance Issues](#performance-issues)
8. [Network and Connectivity](#network-and-connectivity)
9. [Authentication and Authorization](#authentication-and-authorization)
10. [Configuration Problems](#configuration-problems)
11. [Log Analysis](#log-analysis)
12. [Debug Tools and Utilities](#debug-tools-and-utilities)
13. [Maintenance Procedures](#maintenance-procedures)
14. [Recovery Procedures](#recovery-procedures)
15. [Prevention and Monitoring](#prevention-and-monitoring)

## Overview

This guide provides comprehensive troubleshooting procedures and maintenance tasks for Catalogizer v3.0. It covers common issues, diagnostic procedures, and preventive maintenance to ensure optimal system operation.

### Emergency Contacts

- **System Administrator**: admin@yourcompany.com
- **Database Administrator**: dba@yourcompany.com
- **Network Operations**: noc@yourcompany.com
- **Emergency Hotline**: +1-XXX-XXX-XXXX

### Support Resources

- **Documentation**: https://docs.catalogizer.com
- **Issue Tracker**: https://github.com/your-org/catalogizer/issues
- **Community Forum**: https://community.catalogizer.com
- **Status Page**: https://status.catalogizer.com

## Quick Diagnostics

### Health Check Commands

```bash
# Quick system health check
curl -s http://localhost:8080/health | jq

# Service status
sudo systemctl status catalogizer

# Resource usage
top -p $(pgrep catalogizer)
free -h
df -h

# Log tail for immediate issues
sudo tail -f /opt/catalogizer/logs/catalogizer.log

# Database connectivity
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"
```

### System Information Script

```bash
#!/bin/bash
# system_info.sh - Quick system diagnostics

echo "=== Catalogizer System Information ==="
echo "Date: $(date)"
echo "Hostname: $(hostname)"
echo "Uptime: $(uptime)"
echo

echo "=== Service Status ==="
sudo systemctl status catalogizer --no-pager
echo

echo "=== Resource Usage ==="
echo "Memory:"
free -h
echo
echo "Disk:"
df -h /opt/catalogizer
echo
echo "CPU Load:"
cat /proc/loadavg
echo

echo "=== Network Status ==="
netstat -tulpn | grep :8080
echo

echo "=== Recent Errors ==="
sudo tail -20 /opt/catalogizer/logs/catalogizer.log | grep -i error
echo

echo "=== Database Status ==="
if command -v psql >/dev/null 2>&1; then
    PGPASSWORD=$DB_PASSWORD psql -h ${DB_HOST:-localhost} -U ${DB_USER:-catalogizer_user} -d ${DB_NAME:-catalogizer} -c "SELECT count(*) as total_users FROM users;" 2>/dev/null || echo "Database connection failed"
else
    echo "PostgreSQL client not installed"
fi
```

### Quick Fix Commands

```bash
# Restart service
sudo systemctl restart catalogizer

# Clear logs (if disk space issue)
sudo truncate -s 0 /opt/catalogizer/logs/catalogizer.log

# Check and fix permissions
sudo chown -R catalogizer:catalogizer /opt/catalogizer
sudo chmod 755 /opt/catalogizer/bin/catalogizer

# Reload configuration
curl -X POST http://localhost:8080/api/config/reload

# Clear cache (if using Redis)
redis-cli FLUSHDB
```

## Common Issues

### Issue: Service Won't Start

**Symptoms:**
- Service fails to start
- Error: "Job for catalogizer.service failed"
- Application exits immediately

**Diagnosis:**
```bash
# Check service status
sudo systemctl status catalogizer -l

# Check logs
sudo journalctl -u catalogizer -f

# Test configuration
/opt/catalogizer/bin/catalogizer --config /opt/catalogizer/config/config.json --validate

# Check file permissions
ls -la /opt/catalogizer/bin/catalogizer
ls -la /opt/catalogizer/config/config.json
```

**Solutions:**

1. **Configuration Issues:**
```bash
# Validate configuration
/opt/catalogizer/bin/catalogizer --validate-config

# Check for syntax errors
cat /opt/catalogizer/config/config.json | jq
```

2. **Permission Issues:**
```bash
sudo chown catalogizer:catalogizer /opt/catalogizer/bin/catalogizer
sudo chmod +x /opt/catalogizer/bin/catalogizer
sudo chown -R catalogizer:catalogizer /opt/catalogizer/config
```

3. **Port Already in Use:**
```bash
# Check what's using the port
sudo netstat -tulpn | grep :8080
sudo lsof -i :8080

# Kill conflicting process
sudo kill $(sudo lsof -t -i:8080)
```

4. **Missing Dependencies:**
```bash
# Check Go installation
go version

# Check required libraries
ldd /opt/catalogizer/bin/catalogizer
```

### Issue: Database Connection Failed

**Symptoms:**
- "Connection refused" errors
- "Authentication failed" messages
- Timeout errors

**Diagnosis:**
```bash
# Test database connectivity
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();"

# Check database service status
sudo systemctl status postgresql

# Check database logs
sudo tail -f /var/log/postgresql/postgresql-14-main.log

# Verify connection parameters
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "User: $DB_USER"
echo "Database: $DB_NAME"
```

**Solutions:**

1. **Database Not Running:**
```bash
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

2. **Incorrect Credentials:**
```bash
# Reset password
sudo -u postgres psql -c "ALTER USER catalogizer_user PASSWORD 'new_password';"

# Update configuration
sudo nano /opt/catalogizer/config/config.json
```

3. **Connection Limit Reached:**
```bash
# Check active connections
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT count(*) as active_connections, setting as max_connections
FROM pg_stat_activity, pg_settings
WHERE name='max_connections';"

# Kill idle connections
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle' AND state_change < now() - interval '1 hour';"
```

4. **Network Issues:**
```bash
# Check firewall rules
sudo ufw status
sudo iptables -L

# Test network connectivity
telnet $DB_HOST $DB_PORT
```

### Issue: High Memory Usage

**Symptoms:**
- System becomes unresponsive
- Out of memory errors
- Swap usage is high

**Diagnosis:**
```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Check Catalogizer memory usage
ps aux | grep catalogizer

# Check for memory leaks
cat /proc/$(pgrep catalogizer)/status | grep -i vm

# Generate memory profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**Solutions:**

1. **Increase System Memory:**
```bash
# Add swap space (temporary solution)
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

2. **Optimize Application:**
```bash
# Restart service to clear memory
sudo systemctl restart catalogizer

# Adjust memory limits
sudo systemctl edit catalogizer
# Add:
# [Service]
# MemoryLimit=2G
```

3. **Configuration Tuning:**
```json
{
  "performance": {
    "worker_count": 2,
    "cache_size": "256MB",
    "queue_size": 500
  }
}
```

### Issue: Slow Performance

**Symptoms:**
- High response times
- Timeouts
- Users complaining about slowness

**Diagnosis:**
```bash
# Check CPU usage
top -p $(pgrep catalogizer)

# Check I/O wait
iostat -x 1 5

# Check network latency
ping -c 5 $DB_HOST

# Generate CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Check database performance
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT query, mean_time, calls
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;"
```

**Solutions:**

1. **Database Optimization:**
```sql
-- Update statistics
ANALYZE;

-- Check for missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public'
ORDER BY n_distinct DESC;

-- Add indexes for slow queries
CREATE INDEX CONCURRENTLY idx_media_items_user_created
ON media_items(user_id, created_at);
```

2. **Application Tuning:**
```json
{
  "database": {
    "max_connections": 20,
    "max_idle_connections": 10,
    "connection_lifetime": "1h"
  },
  "performance": {
    "cache": {
      "enabled": true,
      "size": "512MB",
      "ttl": "1h"
    }
  }
}
```

3. **System Optimization:**
```bash
# Increase file descriptors
echo "catalogizer soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "catalogizer hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Optimize kernel parameters
echo "net.core.somaxconn = 65535" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

### Issue: SSL/TLS Certificate Problems

**Symptoms:**
- Browser warnings about insecure connections
- Certificate expired errors
- SSL handshake failures

**Diagnosis:**
```bash
# Check certificate validity
openssl x509 -in /etc/ssl/certs/catalogizer.crt -text -noout

# Test SSL connection
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com

# Check certificate expiry
echo | openssl s_client -connect yourdomain.com:443 2>/dev/null | openssl x509 -noout -dates

# Verify certificate chain
curl -I https://yourdomain.com
```

**Solutions:**

1. **Renew Let's Encrypt Certificate:**
```bash
sudo certbot renew --dry-run
sudo certbot renew
sudo systemctl reload nginx
```

2. **Update Certificate Files:**
```bash
# Copy new certificate
sudo cp new_certificate.crt /etc/ssl/certs/catalogizer.crt
sudo cp new_private.key /etc/ssl/private/catalogizer.key

# Set permissions
sudo chmod 644 /etc/ssl/certs/catalogizer.crt
sudo chmod 600 /etc/ssl/private/catalogizer.key

# Restart web server
sudo systemctl restart nginx
```

3. **Fix Certificate Chain:**
```bash
# Concatenate intermediate certificates
cat domain.crt intermediate.crt > /etc/ssl/certs/catalogizer.crt
```

## Service Management

### Systemd Service Operations

```bash
# Service status and control
sudo systemctl status catalogizer
sudo systemctl start catalogizer
sudo systemctl stop catalogizer
sudo systemctl restart catalogizer
sudo systemctl reload catalogizer

# Enable/disable service
sudo systemctl enable catalogizer
sudo systemctl disable catalogizer

# View service logs
sudo journalctl -u catalogizer -f
sudo journalctl -u catalogizer --since "1 hour ago"
sudo journalctl -u catalogizer --since "2024-01-01"

# Check service configuration
sudo systemctl cat catalogizer
sudo systemctl show catalogizer
```

### Process Management

```bash
# Find Catalogizer processes
ps aux | grep catalogizer
pgrep -f catalogizer

# Kill processes
sudo pkill -f catalogizer
sudo kill -TERM $(pgrep catalogizer)
sudo kill -KILL $(pgrep catalogizer)

# Check process tree
pstree -p $(pgrep catalogizer)

# Monitor process resources
watch -n 1 'ps aux | grep catalogizer'
```

### Configuration Reload

```bash
# Graceful configuration reload
curl -X POST http://localhost:8080/api/config/reload
kill -HUP $(pgrep catalogizer)

# Full restart with new configuration
sudo systemctl restart catalogizer

# Validate configuration before reload
/opt/catalogizer/bin/catalogizer --validate-config
```

## Database Issues

### Connection Pool Management

```bash
# Monitor connection pool
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT count(*) filter (where state = 'active') as active,
       count(*) filter (where state = 'idle') as idle,
       count(*) as total
FROM pg_stat_activity
WHERE usename = 'catalogizer_user';"

# Kill idle connections
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
  AND state_change < now() - interval '1 hour'
  AND usename = 'catalogizer_user';"
```

### Database Maintenance

```bash
# Vacuum and analyze tables
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
VACUUM ANALYZE users;
VACUUM ANALYZE media_items;
VACUUM ANALYZE analytics_events;
VACUUM ANALYZE log_entries;"

# Check table sizes
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"

# Check index usage
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT schemaname, tablename, indexname, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_tup_read DESC;"
```

### Database Performance Issues

```sql
-- Find slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Find blocking queries
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_statement,
       blocking_activity.query AS current_statement_in_blocking_process
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- Check for table bloat
SELECT schemaname, tablename,
       round(100 * pg_relation_size(schemaname||'.'||tablename) / pg_total_relation_size(schemaname||'.'||tablename)) AS table_pct,
       round(100 * (pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) / pg_total_relation_size(schemaname||'.'||tablename)) AS index_pct
FROM pg_tables
WHERE schemaname = 'public';
```

## Storage and Media Issues

### Disk Space Management

```bash
# Check disk usage
df -h /opt/catalogizer/media
du -sh /opt/catalogizer/media/*

# Find large files
find /opt/catalogizer/media -type f -size +100M -exec ls -lh {} \; | sort -k5 -hr

# Clean up temporary files
find /opt/catalogizer/media/temp -type f -mtime +7 -delete

# Check inode usage
df -i /opt/catalogizer/media
```

### Media File Issues

```bash
# Check for corrupted files
find /opt/catalogizer/media -type f -name "*.jpg" -exec jpeginfo -c {} \; | grep -v "OK"
find /opt/catalogizer/media -type f -name "*.png" -exec pngcheck {} \; | grep -v "OK"

# Fix permissions
sudo chown -R catalogizer:catalogizer /opt/catalogizer/media
find /opt/catalogizer/media -type f -exec chmod 644 {} \;
find /opt/catalogizer/media -type d -exec chmod 755 {} \;

# Rebuild media index
curl -X POST http://localhost:8080/api/admin/rebuild-index \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Check for orphaned files
curl -X GET http://localhost:8080/api/admin/orphaned-files \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Storage Configuration Issues

```bash
# Test storage connectivity (S3)
aws s3 ls s3://your-bucket-name/ --profile catalogizer

# Test WebDAV connectivity
curl -X PROPFIND http://webdav.example.com/catalogizer/ \
  -u username:password \
  -H "Depth: 1"

# Check local storage permissions
namei -l /opt/catalogizer/media/uploads/
```

## Performance Issues

### CPU Performance

```bash
# Monitor CPU usage
top -p $(pgrep catalogizer)
htop -p $(pgrep catalogizer)

# Check CPU-intensive processes
ps aux --sort=-%cpu | head -10

# Generate CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof catalogizer cpu.prof

# Check system load
uptime
cat /proc/loadavg
```

### Memory Performance

```bash
# Check memory usage patterns
vmstat 1 10
sar -r 1 10

# Check for memory leaks
valgrind --tool=memcheck --leak-check=full /opt/catalogizer/bin/catalogizer

# Monitor garbage collection
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof catalogizer heap.prof
```

### I/O Performance

```bash
# Monitor disk I/O
iostat -x 1 5
iotop -o

# Check disk performance
dd if=/dev/zero of=/opt/catalogizer/test_file bs=1M count=1000 oflag=direct
rm /opt/catalogizer/test_file

# Monitor file descriptor usage
lsof -p $(pgrep catalogizer) | wc -l
cat /proc/$(pgrep catalogizer)/limits | grep "Max open files"
```

## Network and Connectivity

### Network Diagnostics

```bash
# Check listening ports
netstat -tulpn | grep catalogizer
ss -tulpn | grep catalogizer

# Test connectivity
curl -v http://localhost:8080/health
wget --spider -S http://localhost:8080/health

# Check DNS resolution
nslookup yourdomain.com
dig yourdomain.com

# Test external connectivity
curl -I https://api.github.com
```

### Load Balancer Issues

```bash
# Check Nginx status
sudo systemctl status nginx
sudo nginx -t

# Check Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# Test backend connectivity
curl -H "Host: yourdomain.com" http://localhost:8080/health

# Check upstream status
curl http://localhost/nginx_status
```

### Firewall Issues

```bash
# Check firewall rules
sudo ufw status verbose
sudo iptables -L -n

# Test port accessibility
nc -zv localhost 8080
telnet localhost 8080

# Check SELinux (if applicable)
getenforce
sudo sealert -a /var/log/audit/audit.log
```

## Authentication and Authorization

### JWT Token Issues

```bash
# Validate JWT token
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/user/profile

# Check token expiration
echo $TOKEN | cut -d. -f2 | base64 -d | jq .exp

# Generate new token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### User Access Issues

```bash
# Check user status
curl -X GET http://localhost:8080/api/admin/users/123 \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Reset user password
curl -X POST http://localhost:8080/api/admin/users/123/reset-password \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Check user permissions
curl -X GET http://localhost:8080/api/admin/users/123/permissions \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Configuration Problems

### Configuration Validation

```bash
# Validate configuration syntax
cat /opt/catalogizer/config/config.json | jq .

# Test configuration
/opt/catalogizer/bin/catalogizer --config /opt/catalogizer/config/config.json --validate

# Check environment variables
env | grep CATALOGIZER

# Compare with default configuration
diff /opt/catalogizer/config/config.json /opt/catalogizer/config/config.json.example
```

### Configuration Recovery

```bash
# Backup current configuration
cp /opt/catalogizer/config/config.json /opt/catalogizer/config/config.json.backup.$(date +%Y%m%d_%H%M%S)

# Restore from backup
curl -X GET http://localhost:8080/api/config/backups \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -X POST http://localhost:8080/api/config/backups/123/restore \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Reset to defaults
cp /opt/catalogizer/config/config.json.default /opt/catalogizer/config/config.json
```

## Log Analysis

### Log File Locations

```bash
# Application logs
/opt/catalogizer/logs/catalogizer.log
/opt/catalogizer/logs/error.log
/opt/catalogizer/logs/access.log

# System logs
/var/log/syslog
/var/log/messages
/var/log/daemon.log

# Service logs
sudo journalctl -u catalogizer
sudo journalctl -u nginx
sudo journalctl -u postgresql
```

### Log Analysis Commands

```bash
# Search for errors
grep -i error /opt/catalogizer/logs/catalogizer.log | tail -20

# Count error types
grep -i error /opt/catalogizer/logs/catalogizer.log | \
  awk -F'"' '{print $4}' | sort | uniq -c | sort -nr

# Analyze response times
awk '/response_time/ {
  sum += $NF; count++
} END {
  print "Average response time:", sum/count "ms"
}' /opt/catalogizer/logs/catalogizer.log

# Find slow requests
awk '/response_time/ && $NF > 1000 {print}' /opt/catalogizer/logs/catalogizer.log

# Check for memory issues
grep -i "out of memory\|memory\|gc" /opt/catalogizer/logs/catalogizer.log

# Monitor logs in real-time
tail -f /opt/catalogizer/logs/catalogizer.log | grep -i error
```

### Log Rotation Issues

```bash
# Check logrotate configuration
sudo cat /etc/logrotate.d/catalogizer

# Test logrotate
sudo logrotate -d /etc/logrotate.d/catalogizer
sudo logrotate -f /etc/logrotate.d/catalogizer

# Manual log rotation
sudo service catalogizer stop
sudo mv /opt/catalogizer/logs/catalogizer.log /opt/catalogizer/logs/catalogizer.log.$(date +%Y%m%d)
sudo service catalogizer start
```

## Debug Tools and Utilities

### Built-in Debug Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Runtime statistics
curl http://localhost:8080/debug/vars | jq

# Memory profiling
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# CPU profiling
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Goroutine analysis
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof

# Full goroutine dump
curl http://localhost:8080/debug/pprof/goroutine?debug=2
```

### System Debugging Tools

```bash
# Process monitoring
strace -p $(pgrep catalogizer)
ltrace -p $(pgrep catalogizer)

# Network monitoring
tcpdump -i any port 8080
netstat -continuous

# File system monitoring
inotify-tools
watch -n 1 'lsof -p $(pgrep catalogizer) | wc -l'

# System call analysis
perf record -p $(pgrep catalogizer) sleep 10
perf report
```

### Custom Debug Scripts

```bash
#!/bin/bash
# debug_collector.sh - Collect debugging information

DEBUG_DIR="/tmp/catalogizer_debug_$(date +%Y%m%d_%H%M%S)"
mkdir -p $DEBUG_DIR

echo "Collecting debug information..."

# System information
uname -a > $DEBUG_DIR/system_info.txt
uptime >> $DEBUG_DIR/system_info.txt
free -h >> $DEBUG_DIR/system_info.txt
df -h >> $DEBUG_DIR/system_info.txt

# Service status
systemctl status catalogizer > $DEBUG_DIR/service_status.txt

# Configuration
cp /opt/catalogizer/config/config.json $DEBUG_DIR/

# Logs
tail -1000 /opt/catalogizer/logs/catalogizer.log > $DEBUG_DIR/app_logs.txt
journalctl -u catalogizer --since "1 hour ago" > $DEBUG_DIR/systemd_logs.txt

# Network status
netstat -tulpn > $DEBUG_DIR/network_status.txt
ss -tulpn >> $DEBUG_DIR/network_status.txt

# Process information
ps aux | grep catalogizer > $DEBUG_DIR/process_info.txt
pmap $(pgrep catalogizer) > $DEBUG_DIR/memory_map.txt

# Create archive
tar -czf catalogizer_debug_$(date +%Y%m%d_%H%M%S).tar.gz -C /tmp $(basename $DEBUG_DIR)
echo "Debug information collected: catalogizer_debug_$(date +%Y%m%d_%H%M%S).tar.gz"
```

## Maintenance Procedures

### Regular Maintenance Tasks

#### Daily Tasks

```bash
#!/bin/bash
# daily_maintenance.sh

# Check service status
systemctl is-active catalogizer || echo "ALERT: Catalogizer service is down"

# Check disk space
DISK_USAGE=$(df /opt/catalogizer | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "WARNING: Disk usage is ${DISK_USAGE}%"
fi

# Check memory usage
MEMORY_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
if [ $MEMORY_USAGE -gt 80 ]; then
    echo "WARNING: Memory usage is ${MEMORY_USAGE}%"
fi

# Check for errors in logs
ERROR_COUNT=$(grep -c "ERROR" /opt/catalogizer/logs/catalogizer.log)
if [ $ERROR_COUNT -gt 10 ]; then
    echo "WARNING: ${ERROR_COUNT} errors found in logs"
fi

# Check database connectivity
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;" >/dev/null 2>&1 || echo "ALERT: Database connection failed"
```

#### Weekly Tasks

```bash
#!/bin/bash
# weekly_maintenance.sh

# Database maintenance
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "VACUUM ANALYZE;"

# Clean up old logs
find /opt/catalogizer/logs -name "*.log.*" -mtime +7 -delete

# Check certificate expiry
if [ -f /etc/ssl/certs/catalogizer.crt ]; then
    CERT_EXPIRY=$(openssl x509 -in /etc/ssl/certs/catalogizer.crt -noout -enddate | cut -d= -f2)
    CERT_EXPIRY_EPOCH=$(date -d "$CERT_EXPIRY" +%s)
    CURRENT_EPOCH=$(date +%s)
    DAYS_TO_EXPIRY=$(( ($CERT_EXPIRY_EPOCH - $CURRENT_EPOCH) / 86400 ))

    if [ $DAYS_TO_EXPIRY -lt 30 ]; then
        echo "WARNING: SSL certificate expires in $DAYS_TO_EXPIRY days"
    fi
fi

# Update system packages
sudo apt update && sudo apt list --upgradable
```

#### Monthly Tasks

```bash
#!/bin/bash
# monthly_maintenance.sh

# Full database backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > /opt/catalogizer/backups/monthly_backup_$(date +%Y%m).sql.gz

# System security updates
sudo apt update && sudo apt upgrade -y

# Clean up old backups
find /opt/catalogizer/backups -name "*.sql.gz" -mtime +90 -delete

# Performance analysis
curl http://localhost:8080/debug/pprof/profile?seconds=60 > /tmp/monthly_profile_$(date +%Y%m).prof

# Generate monthly report
echo "Monthly System Report - $(date +%Y-%m)" > /tmp/monthly_report.txt
echo "======================================" >> /tmp/monthly_report.txt
echo "Uptime: $(uptime)" >> /tmp/monthly_report.txt
echo "Disk Usage: $(df -h /opt/catalogizer)" >> /tmp/monthly_report.txt
echo "Memory Usage: $(free -h)" >> /tmp/monthly_report.txt
echo "Top Processes: $(ps aux --sort=-%cpu | head -5)" >> /tmp/monthly_report.txt
```

### Automated Monitoring

```bash
#!/bin/bash
# monitoring_script.sh

ALERT_EMAIL="admin@yourcompany.com"
LOG_FILE="/var/log/catalogizer_monitoring.log"

log_message() {
    echo "$(date): $1" >> $LOG_FILE
}

send_alert() {
    echo "$1" | mail -s "Catalogizer Alert" $ALERT_EMAIL
    log_message "ALERT: $1"
}

# Check service status
if ! systemctl is-active --quiet catalogizer; then
    send_alert "Catalogizer service is not running"
fi

# Check response time
RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' http://localhost:8080/health)
if (( $(echo "$RESPONSE_TIME > 5.0" | bc -l) )); then
    send_alert "High response time: ${RESPONSE_TIME}s"
fi

# Check disk space
DISK_USAGE=$(df /opt/catalogizer | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 90 ]; then
    send_alert "Critical disk usage: ${DISK_USAGE}%"
fi

# Check memory usage
MEMORY_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
if [ $MEMORY_USAGE -gt 90 ]; then
    send_alert "Critical memory usage: ${MEMORY_USAGE}%"
fi

log_message "Monitoring check completed"
```

## Recovery Procedures

### Service Recovery

```bash
#!/bin/bash
# service_recovery.sh

echo "Starting Catalogizer service recovery..."

# Stop all related services
sudo systemctl stop catalogizer
sudo systemctl stop nginx

# Check for hung processes
if pgrep catalogizer; then
    echo "Killing hung processes..."
    sudo pkill -9 catalogizer
fi

# Check disk space
DISK_USAGE=$(df /opt/catalogizer | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 95 ]; then
    echo "Critical disk space - cleaning logs..."
    sudo find /opt/catalogizer/logs -name "*.log.*" -delete
    sudo truncate -s 100M /opt/catalogizer/logs/catalogizer.log
fi

# Verify configuration
if ! /opt/catalogizer/bin/catalogizer --validate-config; then
    echo "Configuration invalid - restoring backup..."
    sudo cp /opt/catalogizer/config/config.json.backup /opt/catalogizer/config/config.json
fi

# Check database connectivity
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;" >/dev/null 2>&1; then
    echo "Database connection failed - checking database service..."
    sudo systemctl restart postgresql
    sleep 10
fi

# Start services
sudo systemctl start catalogizer
sleep 5

# Verify service is running
if systemctl is-active --quiet catalogizer; then
    echo "Service recovery successful"
    sudo systemctl start nginx
else
    echo "Service recovery failed - manual intervention required"
    exit 1
fi
```

### Database Recovery

```bash
#!/bin/bash
# database_recovery.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

echo "Starting database recovery from $BACKUP_FILE..."

# Stop application
sudo systemctl stop catalogizer

# Create recovery database
PGPASSWORD=$DB_PASSWORD createdb -h $DB_HOST -U $DB_USER catalogizer_recovery

# Restore from backup
if [[ $BACKUP_FILE == *.gz ]]; then
    gunzip -c $BACKUP_FILE | PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER catalogizer_recovery
else
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER catalogizer_recovery < $BACKUP_FILE
fi

# Verify restoration
ROW_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER catalogizer_recovery -t -c "SELECT count(*) FROM users;")
echo "Restored database contains $ROW_COUNT users"

# Swap databases
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -c "
ALTER DATABASE catalogizer RENAME TO catalogizer_old;
ALTER DATABASE catalogizer_recovery RENAME TO catalogizer;
"

# Start application
sudo systemctl start catalogizer

# Test application
sleep 10
if curl -f http://localhost:8080/health; then
    echo "Database recovery successful"
    # Clean up old database after verification
    # PGPASSWORD=$DB_PASSWORD dropdb -h $DB_HOST -U $DB_USER catalogizer_old
else
    echo "Application failed to start - rolling back..."
    sudo systemctl stop catalogizer
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -c "
    ALTER DATABASE catalogizer RENAME TO catalogizer_failed;
    ALTER DATABASE catalogizer_old RENAME TO catalogizer;
    "
    sudo systemctl start catalogizer
fi
```

## Prevention and Monitoring

### Proactive Monitoring Setup

```bash
# Install monitoring tools
sudo apt install prometheus-node-exporter
sudo systemctl enable prometheus-node-exporter
sudo systemctl start prometheus-node-exporter

# Configure log monitoring
sudo apt install logwatch
echo "logwatch --detail high --mailto admin@yourcompany.com --service catalogizer" | sudo tee /etc/cron.daily/catalogizer-logwatch
```

### Health Check Monitoring

```bash
#!/bin/bash
# health_monitor.sh

ENDPOINTS=(
    "http://localhost:8080/health"
    "http://localhost:8080/api/health"
    "http://localhost:8080/metrics"
)

for endpoint in "${ENDPOINTS[@]}"; do
    if ! curl -f -s "$endpoint" >/dev/null; then
        echo "ALERT: $endpoint is not responding"
        # Send alert notification
    fi
done
```

### Preventive Measures

1. **Regular Backups**: Automated daily backups with verification
2. **Monitoring**: Comprehensive monitoring of all system metrics
3. **Updates**: Regular security updates and patches
4. **Testing**: Regular disaster recovery testing
5. **Documentation**: Keep troubleshooting procedures updated
6. **Training**: Ensure team is familiar with recovery procedures

### Alerting Configuration

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@yourcompany.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  email_configs:
  - to: 'admin@yourcompany.com'
    subject: 'Catalogizer Alert: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      Alert: {{ .Annotations.summary }}
      Description: {{ .Annotations.description }}
      {{ end }}
```

For additional help with specific issues not covered in this guide, please:

1. Check the [Configuration Guide](CONFIGURATION_GUIDE.md) for setup issues
2. Review the [Deployment Guide](DEPLOYMENT_GUIDE.md) for infrastructure problems
3. Consult the [API Documentation](api/API_DOCUMENTATION.md) for integration issues
4. Contact support with the debug information collected using the scripts in this guide
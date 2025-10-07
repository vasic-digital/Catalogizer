# Catalogizer Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Catalogizer system, including multi-protocol storage resilience, database problems, and performance issues.

## Table of Contents

- [Storage Connection Issues](#storage-connection-issues)
- [Database Problems](#database-problems)
- [Authentication Issues](#authentication-issues)
- [Performance Problems](#performance-problems)
- [Frontend Issues](#frontend-issues)
- [API Errors](#api-errors)
- [Deployment Issues](#deployment-issues)
- [Monitoring & Diagnostics](#monitoring--diagnostics)

## Storage Connection Issues

### Symptoms
- Media files not being detected
- "Storage source offline" notifications
- Connection timeouts in logs
- Circuit breaker activation

### Diagnostic Steps

#### 1. Check Storage Source Status
```bash
# API endpoint to check all storage sources
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/storage/sources/status
```

Expected response for healthy sources:
```json
{
  "sources": {
    "smb_123": {
      "name": "Media Server 1",
      "protocol": "smb",
      "path": "smb://server1/media",
      "state": "connected",
      "last_connected": "2024-01-15T10:30:00Z",
      "retry_attempts": 0,
      "is_enabled": true
    },
    "ftp_456": {
      "name": "FTP Server",
      "protocol": "ftp",
      "path": "ftp://ftp.example.com/media",
      "state": "connected",
      "last_connected": "2024-01-15T10:25:00Z",
      "retry_attempts": 0,
      "is_enabled": true
    }
  },
  "summary": {
    "total": 3,
    "connected": 3,
    "disconnected": 0,
    "offline": 0
  }
}
```

#### 2. Test SMB Connectivity Manually
```bash
# Test basic SMB connection
smbclient -L //your-server -U your-username

# Test file listing
smbclient //your-server/share -U your-username -c "ls"

# Test with different authentication
smbclient //your-server/share -U domain/username%password -c "ls"
```

#### 3. Check Network Connectivity
```bash
# Test network connectivity to SMB server
ping your-server

# Test specific SMB ports
nc -zv your-server 445  # SMB over TCP
nc -zv your-server 139  # NetBIOS

# Check firewall rules
sudo iptables -L | grep 445
```

### Common Solutions

#### 1. Authentication Issues
```env
# Update credentials in .env file
SMB_USERNAME=correct-username
SMB_PASSWORD=correct-password
SMB_DOMAIN=your-domain

# For workgroup environments, try without domain
SMB_DOMAIN=
```

#### 2. Network/Firewall Issues
```bash
# Allow SMB ports through firewall
sudo ufw allow 445
sudo ufw allow 139

# Check if SMB services are running on server
sudo systemctl status smbd
sudo systemctl status nmbd
```

#### 3. Circuit Breaker Reset
```bash
# Force reconnection via API
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/smb/sources/smb_123/reconnect
```

#### 4. SMB Protocol Version Issues
```bash
# Try different SMB protocol versions in smb.conf
echo "client min protocol = SMB2" >> /etc/samba/smb.conf
echo "client max protocol = SMB3" >> /etc/samba/smb.conf
```

### SMB Resilience Features

#### Offline Cache Behavior
When SMB sources become unavailable:
1. System switches to offline mode
2. Cached metadata serves user requests
3. Background reconnection attempts continue
4. Once reconnected, cached changes are synchronized

#### Circuit Breaker States
- **Closed**: Normal operation, all requests pass through
- **Open**: Too many failures, requests fail fast
- **Half-Open**: Testing if service has recovered

#### Recovery Process
1. Detect connection failure
2. Activate circuit breaker
3. Enable offline cache
4. Start exponential backoff retry
5. Test connection in half-open state
6. Resume normal operation when stable

## Database Problems

### Symptoms
- "Database locked" errors
- Slow query performance
- Connection pool exhausted
- Data corruption warnings

### Diagnostic Steps

#### 1. Check Database Health
```bash
# Test basic connectivity
sqlite3 /var/lib/catalogizer/catalogizer.db ".tables"

# Check database integrity
sqlite3 /var/lib/catalogizer/catalogizer.db "PRAGMA integrity_check;"

# Check database size and page count
sqlite3 /var/lib/catalogizer/catalogizer.db ".schema" | head -20
```

#### 2. Monitor Database Performance
```sql
-- Check slow queries
EXPLAIN QUERY PLAN SELECT * FROM media_items WHERE title LIKE '%matrix%';

-- Check index usage
.indices media_items

-- Analyze database statistics
ANALYZE;
```

### Common Solutions

#### 1. Database Locks
```bash
# Kill any processes holding locks
fuser /var/lib/catalogizer/catalogizer.db

# Check for long-running transactions
sqlite3 /var/lib/catalogizer/catalogizer.db "PRAGMA busy_timeout = 30000;"
```

#### 2. Performance Optimization
```sql
-- Rebuild indexes
REINDEX;

-- Update statistics
ANALYZE;

-- Vacuum database to reclaim space
VACUUM;
```

#### 3. Connection Pool Issues
```go
// Adjust connection pool settings in code
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

#### 4. Backup and Recovery
```bash
# Create backup before fixes
sqlite3 /var/lib/catalogizer/catalogizer.db ".backup backup.db"

# Restore from backup if needed
cp backup.db /var/lib/catalogizer/catalogizer.db
```

## Authentication Issues

### Symptoms
- "Invalid or expired token" errors
- Login failures with correct credentials
- Permission denied errors
- Session timeouts

### Diagnostic Steps

#### 1. Check JWT Token
```bash
# Decode JWT token (without verification)
echo "your-jwt-token" | cut -d. -f2 | base64 -d | jq .

# Check token expiration
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/auth/status
```

#### 2. Verify User Credentials
```bash
# Check user exists in database
sqlite3 /var/lib/catalogizer/catalogizer.db \
  "SELECT username, role, is_active FROM users WHERE username = 'admin';"
```

### Common Solutions

#### 1. Token Refresh
```bash
# Use refresh token to get new access token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "your-refresh-token"}'
```

#### 2. Password Reset
```go
// Reset admin password via CLI tool
go run cmd/admin/main.go reset-password admin newpassword123
```

#### 3. Permission Issues
```sql
-- Check user permissions
SELECT u.username, u.role, GROUP_CONCAT(p.name) as permissions
FROM users u
LEFT JOIN user_permissions up ON u.id = up.user_id
LEFT JOIN permissions p ON up.permission_id = p.id
WHERE u.username = 'your-username';
```

## Performance Problems

### Symptoms
- Slow page loading
- High CPU/memory usage
- Database query timeouts
- Unresponsive UI

### Diagnostic Steps

#### 1. Monitor System Resources
```bash
# Check CPU and memory usage
top -p $(pgrep catalogizer)

# Monitor I/O usage
iotop

# Check disk space
df -h /var/lib/catalogizer
```

#### 2. Profile Application Performance
```bash
# Enable Go pprof endpoint
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze memory usage
curl http://localhost:8080/debug/pprof/heap > heap.prof

# View profiles
go tool pprof cpu.prof
go tool pprof heap.prof
```

#### 3. Database Performance Analysis
```sql
-- Find slow queries
.timer on
SELECT COUNT(*) FROM media_items;

-- Check query plans
EXPLAIN QUERY PLAN
SELECT * FROM media_items
WHERE media_type = 'movie'
ORDER BY updated_at DESC
LIMIT 20;
```

### Common Solutions

#### 1. Database Optimization
```sql
-- Add missing indexes
CREATE INDEX IF NOT EXISTS idx_media_type_updated
ON media_items(media_type, updated_at DESC);

-- Optimize queries
CREATE INDEX IF NOT EXISTS idx_media_search_fts
ON media_items(title, description);
```

#### 2. Caching Improvements
```go
// Implement query result caching
cache := make(map[string]interface{})
if result, exists := cache[key]; exists {
    return result
}
```

#### 3. Resource Limits
```bash
# Increase file descriptor limits
echo "catalogizer soft nofile 65536" >> /etc/security/limits.conf
echo "catalogizer hard nofile 65536" >> /etc/security/limits.conf

# Optimize Go garbage collector
export GOGC=100
export GOMEMLIMIT=2GiB
```

## Frontend Issues

### Symptoms
- White screen on load
- API call failures
- WebSocket disconnections
- UI components not updating

### Diagnostic Steps

#### 1. Check Browser Console
```javascript
// Check for JavaScript errors
console.log("Catalogizer debug info");

// Check API connectivity
fetch('/api/v1/health')
  .then(r => r.json())
  .then(console.log)
  .catch(console.error);
```

#### 2. Network Issues
```bash
# Check if API is accessible
curl http://localhost:8080/api/v1/health

# Test WebSocket connection
wscat -c ws://localhost:8080/ws
```

### Common Solutions

#### 1. API Configuration
```env
# Check environment variables
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
```

#### 2. CORS Issues
```go
// Add CORS middleware in Go API
func CORSMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })
}
```

#### 3. Build Issues
```bash
# Clear build cache
rm -rf node_modules/.vite
npm run build

# Check for dependency conflicts
npm audit fix
```

## API Errors

### Common Error Codes and Solutions

#### 401 Unauthorized
```json
{"error": "Invalid or expired token"}
```

**Solution:**
```bash
# Refresh token or re-login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'
```

#### 403 Forbidden
```json
{"error": "Insufficient permissions"}
```

**Solution:**
```sql
-- Check user permissions
SELECT role, permissions FROM users WHERE username = 'your-user';

-- Update user role if needed
UPDATE users SET role = 'admin' WHERE username = 'your-user';
```

#### 500 Internal Server Error
```json
{"error": "Internal server error"}
```

**Solution:**
```bash
# Check server logs
journalctl -u catalogizer-api -f

# Check disk space
df -h

# Restart service if needed
sudo systemctl restart catalogizer-api
```

#### 503 Service Unavailable
```json
{"error": "Circuit breaker is open"}
```

**Solution:**
```bash
# Check service health
curl http://localhost:8080/health

# Reset circuit breaker
curl -X POST http://localhost:8080/api/v1/system/circuit-breaker/reset
```

## Deployment Issues

### Docker Deployment Problems

#### Container Won't Start
```bash
# Check container logs
docker logs catalogizer-api

# Check container status
docker ps -a

# Inspect container configuration
docker inspect catalogizer-api
```

#### Volume Mount Issues
```bash
# Check volume permissions
ls -la /path/to/volume

# Fix permissions
sudo chown -R 1000:1000 /path/to/volume
```

### Kubernetes Deployment Problems

#### Pod Crashes
```bash
# Check pod status
kubectl get pods -n catalogizer

# View pod logs
kubectl logs -f deployment/catalogizer-api -n catalogizer

# Describe pod for events
kubectl describe pod <pod-name> -n catalogizer
```

#### Service Discovery Issues
```bash
# Test service connectivity
kubectl exec -it <pod-name> -n catalogizer -- curl catalogizer-api:8080/health

# Check service endpoints
kubectl get endpoints -n catalogizer
```

## Monitoring & Diagnostics

### Health Check Endpoints

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed status
curl http://localhost:8080/api/v1/status

# SMB sources health
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/smb/health

# Database health
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/db/health
```

### Log Analysis

#### Application Logs
```bash
# Follow application logs
tail -f /opt/catalogizer/logs/catalogizer.log

# Search for errors
grep -i error /opt/catalogizer/logs/catalogizer.log

# Filter by level
grep -i "level=error" /opt/catalogizer/logs/catalogizer.log
```

#### System Logs
```bash
# Service logs
journalctl -u catalogizer-api -f

# System messages
tail -f /var/log/syslog | grep catalogizer

# Nginx logs
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
```

### Performance Monitoring

#### Metrics Collection
```bash
# CPU and memory usage
ps aux | grep catalogizer

# Network connections
netstat -tulpn | grep :8080

# File descriptors
lsof -p $(pgrep catalogizer) | wc -l
```

#### Database Monitoring
```sql
-- Active connections
.show

-- Table sizes
SELECT name, COUNT(*) FROM sqlite_master
WHERE type='table'
GROUP BY name;

-- Index usage statistics
PRAGMA index_list('media_items');
```

### Emergency Recovery

#### Service Recovery
```bash
# Quick restart
sudo systemctl restart catalogizer-api

# Full system restart
sudo systemctl stop catalogizer-api
sudo systemctl start catalogizer-api

# Reset to defaults
rm /var/lib/catalogizer/catalogizer.db
go run cmd/migrate/main.go
```

#### Data Recovery
```bash
# Restore from backup
sudo systemctl stop catalogizer-api
cp /opt/catalogizer/backup/latest.db /var/lib/catalogizer/catalogizer.db
sudo systemctl start catalogizer-api

# Emergency admin user creation
go run cmd/admin/main.go create-admin emergency-admin secure-password-123
```

## Getting Help

### Debug Information Collection

When reporting issues, please collect the following information:

```bash
#!/bin/bash
# debug-info.sh - Collect debug information

echo "=== Catalogizer Debug Information ===" > debug-info.txt
echo "Date: $(date)" >> debug-info.txt
echo "Hostname: $(hostname)" >> debug-info.txt
echo "" >> debug-info.txt

echo "=== System Information ===" >> debug-info.txt
uname -a >> debug-info.txt
cat /etc/os-release >> debug-info.txt
echo "" >> debug-info.txt

echo "=== Service Status ===" >> debug-info.txt
systemctl status catalogizer-api >> debug-info.txt
echo "" >> debug-info.txt

echo "=== Recent Logs ===" >> debug-info.txt
tail -100 /opt/catalogizer/logs/catalogizer.log >> debug-info.txt
echo "" >> debug-info.txt

echo "=== SMB Status ===" >> debug-info.txt
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/smb/sources/status >> debug-info.txt
echo "" >> debug-info.txt

echo "=== Database Info ===" >> debug-info.txt
sqlite3 /var/lib/catalogizer/catalogizer.db ".schema" | head -50 >> debug-info.txt
echo "" >> debug-info.txt

echo "Debug information collected in debug-info.txt"
```

### Support Channels

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check README.md and ARCHITECTURE.md
- **Community**: Discord/Slack for community support
- **Enterprise**: Professional support options

### Useful Resources

- [Architecture Documentation](ARCHITECTURE.md)
- [Deployment Guide](DEPLOYMENT.md)
- [API Documentation](API.md)
- [SMB Resilience Guide](SMB_RESILIENCE.md)

This troubleshooting guide covers the most common issues you may encounter. For specific problems not covered here, please check the logs and consider opening a GitHub issue with the debug information collected using the script above.
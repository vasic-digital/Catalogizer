# Runbook: API Unresponsive / High Error Rate

**Alert:** APIDown / HighErrorRate / CriticalErrorRate / HighLatency  
**Severity:** Critical / High  
**Service:** catalogizer-api  

---

## Summary

This alert fires when the API is down, returning errors, or responding slowly.

---

## Immediate Assessment (1 minute)

### 1. Check API Health
```bash
# Basic health check
curl -f http://localhost:8080/health || echo "API DOWN"

# Check metrics endpoint
curl http://localhost:8080/metrics > /dev/null && echo "API UP"
```

### 2. Check Process Status
```bash
# Is the process running?
ps aux | grep catalogizer-api

# Check systemd status
systemctl status catalogizer-api

# Check container status
podman ps | grep catalogizer-api
```

### 3. Check Logs
```bash
# Recent errors
journalctl -u catalogizer-api -n 50 --no-pager

# Or container logs
podman logs --tail 50 catalogizer-api
```

---

## Diagnosis Steps

### Scenario A: API Process Not Running

```bash
# Try to start the service
systemctl start catalogizer-api

# Check for startup errors
journalctl -u catalogizer-api -f
```

**Common Causes:**
- OOM Killer terminated process
- Panic/crash
- Dependency failure (database, Redis)
- Configuration error

### Scenario B: API Running But Unresponsive

```bash
# Check if port is listening
ss -tuln | grep 8080
netstat -tlnp | grep 8080

# Check connection count
ss -tan | grep :8080 | wc -l

# Check goroutine count (if debug endpoint available)
curl http://localhost:8080/debug/pprof/goroutine?debug=1 | head -5
```

**Common Causes:**
- Deadlock
- Goroutine leak (too many concurrent connections)
- Database connection pool exhausted
- External dependency timeout

### Scenario C: High Error Rate

```bash
# Check error logs
grep "ERROR" /var/log/catalogizer/api.log | tail -50

# Check specific error types
grep -E "(panic|fatal|ERROR)" /var/log/catalogizer/api.log | tail -20
```

**Common Causes:**
- Database connection failures
- External API failures (TMDB, etc.)
- File system permission issues
- Memory exhaustion

### Scenario D: High Latency

```bash
# Check pprof for blocking calls
curl http://localhost:8080/debug/pprof/block?debug=1

# Check mutex contention
curl http://localhost:8080/debug/pprof/mutex?debug=1
```

**Common Causes:**
- Slow database queries
- External API latency
- Lock contention
- Resource exhaustion

---

## Resolution Steps

### Immediate Actions

1. **Restart Service** (if down or unresponsive)
   ```bash
   # Graceful restart
   systemctl restart catalogizer-api
   
   # Or if systemd not working
   kill -TERM $(pgrep -f catalogizer-api)
   sleep 5
   /usr/local/bin/catalogizer-api &
   ```

2. **Clear Connection Pool** (if connection issues)
   ```bash
   # Restart to clear stuck connections
   systemctl restart catalogizer-api
   ```

3. **Scale Horizontally** (if load-related)
   ```bash
   # Add more instances
   podman-compose up -d --scale catalogizer-api=3
   ```

### Database-Related Issues

1. **Check Database Connectivity**
   ```bash
   # Test database connection
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"
   ```

2. **Check for Locks**
   ```sql
   -- PostgreSQL
   SELECT * FROM pg_locks WHERE NOT granted;
   
   -- Check for blocked queries
   SELECT * FROM pg_stat_activity WHERE wait_event_type = 'Lock';
   ```

3. **Restart Database if Necessary**
   ```bash
   # Only if absolutely necessary
   systemctl restart postgresql
   ```

### Dependency Issues

1. **Check Redis** (if using)
   ```bash
   redis-cli ping
   redis-cli info stats
   ```

2. **Check External APIs**
   ```bash
   # Test TMDB connectivity
   curl -I https://api.themoviedb.org/3
   ```

---

## Verification

After taking action:

```bash
# 1. Check health endpoint
curl -f http://localhost:8080/health

# 2. Check metrics
curl http://localhost:8080/metrics | grep -E "(http_requests_total|up)"

# 3. Run smoke tests
./scripts/smoke-test.sh

# 4. Monitor for 5 minutes
watch -n 10 'curl -s http://localhost:8080/health'
```

---

## Prevention

- Implement health checks with detailed status
- Set up circuit breakers for external dependencies
- Use connection pooling with proper limits
- Implement graceful shutdown handling
- Set up automated restarts (systemd restart=always)
- Regular load testing

---

## Escalation

If API remains down after 10 minutes:
1. Escalate to Engineering Manager
2. Activate incident response
3. Consider failover to backup systems
4. Notify users via status page

---

**Related Runbooks:**
- [DATABASE_CONNECTION.md](DATABASE_CONNECTION.md)
- [HIGH_CPU.md](HIGH_CPU.md)
- [MEMORY_LEAK.md](MEMORY_LEAK.md)

# Runbook: Memory Leak / High Memory Usage

**Alert:** HighMemoryUsage / CriticalMemoryUsage  
**Severity:** Warning (80%) / Critical (95%)  
**Service:** System / Application  

---

## Summary

This alert fires when memory usage exceeds defined thresholds (80% for warning, 95% for critical) for a sustained period.

---

## Initial Assessment (2 minutes)

### 1. Verify Memory Usage
```bash
# Check current memory usage
free -h
vmstat -s | head -5

# Check if swap is being used
swapon -s
```

### 2. Identify Memory-Heavy Processes
```bash
# Top memory consumers
ps aux --sort=-%mem | head -10

# Check Catalogizer specifically
pmap $(pgrep -f catalogizer-api) | tail -20
```

### 3. Check for OOM Events
```bash
# Check system logs for OOM killer
dmesg | grep -i "out of memory"
journalctl -k | grep -i "oom"
```

---

## Diagnosis Steps

### For Catalogizer API

```bash
# Check Go memory stats
curl -s http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Inside pprof:
# (pprof) top 20
# (pprof) list <function-name>

# Get memory statistics
curl http://localhost:8080/metrics | grep -E "(go_memstats|process_resident)"
```

**Key Metrics to Check:**
- `go_memstats_heap_inuse_bytes` - Current heap usage
- `go_memstats_heap_alloc_bytes` - Allocated heap
- `go_memstats_sys_bytes` - Total system memory
- `process_resident_memory_bytes` - RSS

### Common Memory Leak Patterns

1. **Goroutine Leak**
   ```bash
   # Check goroutine count
   curl http://localhost:8080/debug/pprof/goroutine?debug=1
   ```

2. **Unclosed Resources**
   - Database connections not released
   - File handles not closed
   - HTTP response bodies not closed

3. **Cache Growth**
   ```bash
   # Check cache size
   curl http://localhost:8080/metrics | grep cache
   ```

---

## Resolution Steps

### Immediate Actions

1. **Graceful Restart** (if critical)
   ```bash
   # Send SIGTERM for graceful shutdown
   kill -TERM $(pgrep -f catalogizer-api)
   
   # Or use systemd
   systemctl restart catalogizer-api
   ```

2. **Free System Cache** (temporary relief)
   ```bash
   # Clear pagecache (safe)
   echo 1 > /proc/sys/vm/drop_caches
   
   # Clear dentries and inodes
   echo 2 > /proc/sys/vm/drop_caches
   
   # Clear all
   echo 3 > /proc/sys/vm/drop_caches
   ```

### Long-Term Fixes

1. **Analyze Heap Profile**
   ```bash
   # Get heap profile over time
   curl http://localhost:8080/debug/pprof/heap > heap-$(date +%s).prof
   
   # Compare profiles
   go tool pprof --diff_base=heap-old.prof heap-new.prof
   ```

2. **Check for Connection Leaks**
   ```bash
   # Database connections
   netstat -an | grep :5432 | wc -l
   
   # Open file descriptors
   ls /proc/$(pgrep -f catalogizer-api)/fd | wc -l
   ```

3. **Enable GC Logs** (if needed)
   ```bash
   # Set environment variable
   export GODEBUG=gctrace=1
   # Restart service
   ```

---

## Verification

After restart or fix:

```bash
# Monitor memory for 10 minutes
watch -n 10 'free -h; echo "---"; ps aux | grep catalogizer-api'

# Check if memory stabilizes
# In Prometheus, watch:
# process_resident_memory_bytes{job="catalogizer-api"}
```

---

## Prevention

### Code Practices
- Always `defer rows.Close()` for database queries
- Use `defer resp.Body.Close()` for HTTP responses
- Implement request timeouts
- Use connection pooling with limits
- Set cache size limits with TTL

### Monitoring
- Set up memory growth alerts
- Monitor goroutine count
- Track connection pool usage
- Set up OOM prevention (systemd OOM score)

### Infrastructure
- Configure memory limits in containers
```yaml
# docker-compose.yml
deploy:
  resources:
    limits:
      memory: 2G
    reservations:
      memory: 512M
```

---

## Escalation

If memory continues to grow after restart:
1. Collect heap dumps every 5 minutes
2. Escalate to Engineering Manager
3. Consider code rollback if recent deployment
4. Enable detailed profiling

---

**Related Runbooks:**
- [HIGH_CPU.md](HIGH_CPU.md)
- [DATABASE_CONNECTION.md](DATABASE_CONNECTION.md)
- [API_UNRESPONSIVE.md](API_UNRESPONSIVE.md)

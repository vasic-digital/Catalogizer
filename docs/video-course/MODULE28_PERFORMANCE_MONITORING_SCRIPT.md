# Module 28: Performance Monitoring

## Video Script — Prometheus, Grafana, k6 & Alerting

### Duration: ~25 minutes

---

### Scene 1: Introduction (2 min)

"You can't optimize what you can't measure. This module covers the full monitoring stack: Prometheus for metrics collection, Grafana for visualization, k6 for load testing, and Alertmanager for notifications."

---

### Scene 2: Prometheus Metrics in catalog-api (5 min)

**File:** `internal/metrics/`

"catalog-api exposes custom Prometheus metrics at `/metrics`."

**Custom metrics:**
- `catalogizer_http_requests_total` — request count by method, path, status
- `catalogizer_http_request_duration_seconds` — response time histogram
- `catalogizer_db_query_duration_seconds` — database query latency
- `catalogizer_active_goroutines` — current goroutine count
- `catalogizer_cache_hits_total` / `catalogizer_cache_misses_total`
- `catalogizer_scan_files_processed_total`
- `catalogizer_websocket_connections_active`

```bash
curl http://localhost:8080/metrics | grep catalogizer_
```

---

### Scene 3: Prometheus Configuration (3 min)

**File:** `monitoring/prometheus/prometheus.yml`

```yaml
scrape_configs:
  - job_name: 'catalog-api'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

**Alert rules** (`monitoring/prometheus/alerts.yml`):
- HighErrorRate: >5% errors for 5min (CRITICAL)
- HighLatency: p95 > 1s for 5min (WARNING)
- LowCacheHitRatio: <50% for 10min (WARNING)
- HighMemoryUsage: >4GB for 5min (WARNING)
- TooManyGoroutines: >1000 for 5min (WARNING)
- DatabaseErrors: >0.1/sec for 2min (CRITICAL)

---

### Scene 4: Grafana Dashboard (4 min)

**File:** `monitoring/grafana/dashboards/catalogizer-overview.json`

"The pre-built dashboard shows request rates, latency percentiles, error rates, cache efficiency, goroutine counts, and memory usage."

```bash
# Start monitoring stack
podman-compose -f docker-compose.yml up -d prometheus grafana

# Access Grafana
open http://localhost:3001  # default admin/admin
```

**Dashboard panels:** Request rate, p50/p95/p99 latency, error %, cache hit ratio, goroutine trend, memory usage, active WebSocket connections, scan progress.

---

### Scene 5: k6 Load Testing (5 min)

**Files:** `tests/k6/`

"k6 simulates realistic traffic patterns to validate performance under load."

**Test types:**

| Test | Users | Duration | Validates |
|------|-------|----------|-----------|
| `load_test.js` | 10→50 | 5 min | p95 < 500ms |
| `stress_test.js` | 50→300 | 10 min | Breaking point |
| `soak_test.js` | 20 | 30 min | No memory leaks |
| `spike_test.js` | 10→200→10 | 5 min | Recovery speed |
| `auth_load_test.js` | 50 | 5 min | Auth throughput |
| `mixed_workload_test.js` | 30 | 10 min | Realistic patterns |

```bash
podman run --rm --network host \
  -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest run /scripts/load_test.js
```

---

### Scene 6: Alertmanager (3 min)

**File:** `monitoring/alertmanager/alertmanager.yml`

"Alertmanager routes alerts from Prometheus to email, webhooks, or Slack."

**Configuration:**
- Critical alerts → immediate email + webhook to `/api/v1/admin/webhooks/alerts`
- Warning alerts → grouped, batched delivery
- Inhibit rules suppress lower-severity duplicates

---

### Scene 7: Monitoring-Driven Optimization (3 min)

"After collecting metrics, optimize based on data — not guesses."

**Workflow:**
1. Run k6 load test
2. Check Grafana for slow endpoints (p95 > 200ms)
3. Add database indexes for slow queries
4. Add caching for frequently-accessed read endpoints
5. Re-run load test to verify improvement

**Challenge:** CH-290 validates all Prometheus metrics emit correctly.

---

### Summary

- Prometheus collects metrics from `/metrics` endpoint
- Grafana visualizes with pre-built dashboard
- k6 validates performance: load, stress, soak, spike
- Alertmanager sends notifications on threshold violations
- Optimize based on measured data, not guesses

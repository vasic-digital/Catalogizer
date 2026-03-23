# Module 18: Monitoring and Observability -- Video Script

**Duration**: 45 minutes
**Prerequisites**: Module 11 (Security and Monitoring), Module 17 (Load Testing)

---

## Video 18.1: Prometheus Metrics (15 min)

### Opening

Welcome to Module 18, our final module. Monitoring answers the question "is my system healthy right now?" while observability answers "why is it behaving this way?" Catalogizer implements both through Prometheus metrics, Grafana dashboards, structured logging, and runtime metrics collection.

### Metrics Architecture

Catalogizer exposes metrics at the `/metrics` endpoint using the Prometheus client library. The metrics middleware (`internal/metrics/`) instruments every HTTP request automatically.

```
Gin Request
    |
    v
metrics.GinMiddleware()  -->  Prometheus Registry
    |                              |
    v                              v
Handler                     /metrics endpoint
                                   |
                                   v
                           Prometheus Server
                                   |
                                   v
                           Grafana Dashboard
```

### Built-In Metrics

The `GinMiddleware()` in `catalog-api/internal/metrics/` registers these metrics automatically:

```go
// HTTP request metrics
http_requests_total{method, path, status}        // Counter: total requests
http_request_duration_seconds{method, path}       // Histogram: request latency
http_request_size_bytes{method, path}             // Histogram: request body size
http_response_size_bytes{method, path}            // Histogram: response body size
```

### Custom Application Metrics

Beyond HTTP metrics, Catalogizer registers domain-specific metrics:

```go
// Scan metrics
scan_operations_total{storage_root, status}       // Counter: scan operations
scan_files_processed_total{storage_root}          // Counter: files processed
scan_duration_seconds{storage_root}               // Histogram: scan duration

// Media entity metrics
media_entities_total{type}                         // Gauge: entities by type
media_detection_duration_seconds                   // Histogram: detection time

// WebSocket metrics
websocket_connections_active                       // Gauge: active connections
websocket_messages_sent_total                      // Counter: messages broadcast
```

### Querying Metrics

```bash
# View raw metrics
curl -s http://localhost:8080/metrics

# Filter for specific metrics
curl -s http://localhost:8080/metrics | grep http_request_duration

# Check request rate (requires Prometheus)
# In Prometheus UI: rate(http_requests_total[5m])
```

### Metric Types Explained

| Type | Use Case | Example |
|------|----------|---------|
| Counter | Values that only increase | `http_requests_total` |
| Gauge | Values that go up and down | `websocket_connections_active` |
| Histogram | Distribution of values | `http_request_duration_seconds` |
| Summary | Pre-calculated quantiles | (not used in Catalogizer) |

Histograms are preferred over summaries because they can be aggregated across instances and time ranges.

---

## Video 18.2: Runtime Metrics Collector (10 min)

### The Runtime Collector

Catalogizer starts a background collector that periodically samples Go runtime statistics:

```go
// In main.go
metrics.StartRuntimeCollector(15 * time.Second)
```

This collector runs every 15 seconds and exports:

```go
// Goroutine metrics
go_goroutines                                     // Current goroutine count
go_threads                                        // OS threads created

// Memory metrics
go_memstats_alloc_bytes                           // Current heap allocation
go_memstats_sys_bytes                             // Total memory from OS
go_memstats_heap_objects                          // Number of heap objects
go_memstats_gc_duration_seconds                   // GC pause duration

// GC metrics
go_gc_duration_seconds                            // GC cycle duration histogram
go_memstats_last_gc_time_seconds                  // Time of last GC
```

### What to Monitor

**Goroutine count** (`go_goroutines`): Should be stable under constant load. A steadily increasing count indicates a goroutine leak. Normal values for Catalogizer under moderate load are 50-200 goroutines.

**Heap allocation** (`go_memstats_alloc_bytes`): Shows a sawtooth pattern as the garbage collector runs. The peaks should not trend upward over hours -- that indicates a memory leak.

**GC pause time** (`go_gc_duration_seconds`): Typically under 1ms for Catalogizer. Long GC pauses (> 10ms) indicate too many heap allocations.

### Detecting Problems

```bash
# Quick health check via metrics
curl -s http://localhost:8080/metrics | grep -E "go_goroutines|go_memstats_alloc_bytes|go_gc_duration"

# Example output:
# go_goroutines 87
# go_memstats_alloc_bytes 4.2e+07
# go_gc_duration_seconds{quantile="0.5"} 0.000234
```

If `go_goroutines` is above 500 in a quiet system, investigate using pprof:

```bash
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

---

## Video 18.3: Grafana Dashboards (10 min)

### Dashboard Setup

Catalogizer includes pre-built Grafana dashboard configurations in the `monitoring/` directory. The monitoring stack runs via Docker Compose:

```bash
# Start Prometheus + Grafana
podman-compose -f docker-compose.dev.yml up prometheus grafana
```

Grafana is accessible at `http://localhost:3001` (default credentials: admin/admin).

### Dashboard Panels

The main Catalogizer dashboard includes these panels:

**Row 1: Request Overview**
- Request Rate: `rate(http_requests_total[5m])`
- Error Rate: `rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])`
- Active Requests: `sum(http_requests_in_flight)`

**Row 2: Latency**
- p50 Latency: `histogram_quantile(0.5, rate(http_request_duration_seconds_bucket[5m]))`
- p95 Latency: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- p99 Latency: `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))`

**Row 3: Go Runtime**
- Goroutines: `go_goroutines`
- Heap Allocation: `go_memstats_alloc_bytes`
- GC Pause: `go_gc_duration_seconds{quantile="0.99"}`

**Row 4: Application**
- WebSocket Connections: `websocket_connections_active`
- Scan Operations: `rate(scan_operations_total[5m])`
- Media Entities: `media_entities_total`

### Creating Custom Panels

To add a panel for endpoint-specific latency:

1. Click "Add panel" in Grafana
2. Set the query:
```promql
histogram_quantile(0.95,
  rate(http_request_duration_seconds_bucket{path="/api/v1/entities"}[5m])
)
```
3. Set the legend to `{{path}}` for label display
4. Set the Y-axis unit to "seconds"

### Prometheus Configuration

The Prometheus configuration in `monitoring/prometheus.yml` scrapes the Catalogizer API:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'catalogizer-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

---

## Video 18.4: Alerting and Log Aggregation (10 min)

### Alerting Rules

Prometheus alerting rules detect problems before users report them. Define rules in `monitoring/alerts.yml`:

```yaml
groups:
  - name: catalogizer
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"

      # High latency
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High p95 latency"
          description: "95th percentile latency is {{ $value }}s"

      # Goroutine leak
      - alert: GoroutineLeak
        expr: go_goroutines > 500
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Possible goroutine leak"
          description: "Goroutine count is {{ $value }}, sustained for 10 minutes"

      # Memory growth
      - alert: HighMemoryUsage
        expr: go_memstats_alloc_bytes > 1e9
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Heap allocation is {{ $value | humanize1024 }}B"

      # Service down
      - alert: ServiceDown
        expr: up{job="catalogizer-api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Catalogizer API is down"
```

### Alertmanager Configuration

Route alerts to email, Slack, or webhooks:

```yaml
# monitoring/alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@catalogizer.local'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'email'
  routes:
    - match:
        severity: critical
      receiver: 'email'
      repeat_interval: 15m

receivers:
  - name: 'email'
    email_configs:
      - to: 'admin@catalogizer.local'
        subject: 'Catalogizer Alert: {{ .GroupLabels.alertname }}'
```

### Structured Logging

Catalogizer uses `go.uber.org/zap` for structured JSON logging:

```go
logger.Info("Scan completed",
    zap.String("storage_root", root.Name),
    zap.Int("files_processed", stats.FilesProcessed),
    zap.Duration("duration", elapsed),
    zap.Error(err),
)
```

This produces structured log lines:

```json
{
  "level": "info",
  "ts": 1711152000.123,
  "caller": "services/universal_scanner.go:245",
  "msg": "Scan completed",
  "storage_root": "NAS-Media",
  "files_processed": 85432,
  "duration": "25m14.3s"
}
```

Structured logs can be aggregated and searched with tools like Loki, Elasticsearch, or even `jq`:

```bash
# Search for errors in structured logs
journalctl -u catalogizer --output=json | jq 'select(.PRIORITY == "3")'

# Find slow scans
journalctl -u catalogizer --output=json | jq 'select(.msg == "Scan completed") | select(.duration | test("^[0-9]{2,}m"))'
```

### The Log Management API

Catalogizer includes a built-in log management system accessible via the API:

```bash
# Create a log collection
curl -X POST http://localhost:8080/api/v1/logs/collect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "debug-session", "level": "debug", "duration": "1h"}'

# List collections
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/logs/collections

# Analyze logs for patterns
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/logs/collections/1/analyze

# Stream logs in real-time
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/logs/stream
```

### Monitoring Checklist for Production

Before deploying to production, verify:

- [ ] Prometheus is scraping the `/metrics` endpoint
- [ ] Grafana dashboards are imported and accessible
- [ ] Alert rules are configured for error rate, latency, memory, and availability
- [ ] Alertmanager routes are configured with valid notification targets
- [ ] Structured logging is writing to a persistent location
- [ ] Log rotation is configured to prevent disk exhaustion
- [ ] The runtime metrics collector is running (15-second interval)
- [ ] Resource limits are enforced (30-40% of host resources)

---

## Exercises

1. Set up Prometheus and Grafana using the monitoring stack and import the Catalogizer dashboard
2. Create an alert rule that fires when the WebSocket connection count drops to zero during business hours
3. Run a soak test (Module 17) while watching the Grafana dashboard and correlate the metrics
4. Write a PromQL query that shows the top 5 slowest API endpoints by p99 latency

---

## Key Files Referenced

- `catalog-api/internal/metrics/` -- Prometheus metrics registration and GinMiddleware
- `catalog-api/internal/middleware/` -- Logger middleware (Zap structured logging)
- `catalog-api/main.go` -- `metrics.StartRuntimeCollector(15 * time.Second)` initialization
- `monitoring/` -- Prometheus, Grafana, and Alertmanager configurations
- `docker-compose.dev.yml` -- Development stack including monitoring services
- `docs/deployment/MONITORING_GUIDE.md` -- Detailed monitoring setup instructions

# Monitoring Guide

Catalogizer provides built-in observability through Prometheus metrics, health endpoints, real-time WebSocket events, and pre-configured Grafana dashboards.

---

## Prometheus Metrics Endpoint

The API server exposes metrics at `/metrics` in Prometheus text format. This endpoint does not require authentication.

```
GET http://localhost:8080/metrics
```

Scrape configuration for `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'catalogizer-api'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

---

## Available Metrics

### HTTP Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests by method, path, and status code |
| `http_request_duration_seconds` | Histogram | Request latency distribution by method and path |
| `http_requests_in_flight` | Gauge | Currently active requests |
| `http_response_size_bytes` | Histogram | Response body size distribution |

### Database Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `db_query_duration_seconds` | Histogram | SQL query execution time by operation type |
| `db_connections_open` | Gauge | Current open database connections |
| `db_connections_idle` | Gauge | Current idle database connections |
| `db_errors_total` | Counter | Database errors by operation |

### Runtime Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `go_goroutines` | Gauge | Number of active goroutines |
| `go_memstats_alloc_bytes` | Gauge | Current heap memory allocation |
| `go_memstats_sys_bytes` | Gauge | Total memory obtained from the OS |
| `go_gc_duration_seconds` | Summary | GC pause duration |

### Application Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `scan_files_total` | Counter | Total files scanned by storage root |
| `scan_duration_seconds` | Histogram | Scan completion time |
| `media_entities_total` | Gauge | Total media entities by type |
| `websocket_connections_active` | Gauge | Currently connected WebSocket clients |

---

## Health Endpoint

The health endpoint provides a quick liveness check:

```
GET http://localhost:8080/health
```

Response:

```json
{
  "status": "ok",
  "version": "3.0.0",
  "uptime": "4h32m15s"
}
```

Use this endpoint for container health checks:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

---

## Grafana Dashboard Setup

Pre-configured Grafana dashboards are available in the `monitoring/` directory.

### Quick Start

1. Start the monitoring stack:

```bash
podman-compose -f docker-compose.dev.yml up prometheus grafana
```

2. Open Grafana at `http://localhost:3001` (default credentials: `admin`/`admin`).

3. The Prometheus data source is auto-provisioned. Import dashboards from `monitoring/dashboards/`.

### Dashboard Panels

The default Catalogizer dashboard includes:

- **Request Rate**: HTTP requests per second by status code
- **Latency Percentiles**: p50, p90, p99 request duration
- **Error Rate**: 4xx and 5xx responses over time
- **Database Performance**: Query duration and connection pool usage
- **Goroutines and Memory**: Runtime resource consumption
- **Scan Activity**: Active scans, files processed, scan duration
- **WebSocket Connections**: Connected client count over time

---

## WebSocket Real-Time Events

The WebSocket endpoint at `/ws` streams real-time operational events. Connect with a valid JWT token:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=<jwt_token>');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // data.type: event type string
  // data.payload: event-specific data
};
```

### Monitoring-Relevant Events

| Event | Description |
|-------|-------------|
| `scan.started` | A storage root scan has begun |
| `scan.progress` | Scan progress with file count and percentage |
| `scan.completed` | Scan finished with summary statistics |
| `source.connected` | Storage source came online |
| `source.disconnected` | Storage source went offline |
| `source.recovered` | Storage source reconnected after failure |

These events enable building live monitoring dashboards in the web UI without polling.

---

## Alerting Configuration

Configure Prometheus alerting rules for critical conditions:

```yaml
groups:
  - name: catalogizer
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High 5xx error rate"

      - alert: SlowRequests
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "p99 latency exceeds 5 seconds"

      - alert: DatabaseConnectionExhaustion
        expr: db_connections_open > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connections nearing pool limit"
```

---

## Resource Limits

The host machine runs other processes. All Catalogizer containers must stay within these budgets:

| Container | CPU Limit | Memory Limit |
|-----------|-----------|-------------|
| catalog-api | 2 CPUs | 4 GB |
| catalog-web | 1 CPU | 2 GB |
| PostgreSQL | 1 CPU | 2 GB |
| Prometheus | 0.5 CPU | 1 GB |
| Grafana | 0.5 CPU | 1 GB |
| **Total** | **5 CPUs** | **10 GB** |

Monitor container resource usage:

```bash
podman stats --no-stream
cat /proc/loadavg
```

Keep total CPU usage under 40% of host capacity to avoid impacting other processes.

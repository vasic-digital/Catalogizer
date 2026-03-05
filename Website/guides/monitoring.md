---
title: Monitoring Guide
description: Setting up monitoring for Catalogizer with Prometheus, Grafana, and built-in health endpoints
---

# Monitoring Guide

Catalogizer exposes Prometheus-compatible metrics and includes pre-configured Grafana dashboards for monitoring API performance, media detection, and system health.

---

## Health Endpoints

The API provides built-in health check endpoints that require no additional setup.

### `/health`

Returns the server's health status and basic information.

```bash
curl http://localhost:8080/health
```

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "48h32m15s"
}
```

Use this endpoint for container health checks, load balancer probes, and uptime monitoring.

### `/metrics`

Exposes Prometheus-compatible metrics for scraping.

```bash
curl http://localhost:8080/metrics
```

This endpoint returns metrics in Prometheus text exposition format, ready for scraping by a Prometheus server.

---

## Prometheus Setup

### Configuration

The project includes a Prometheus configuration file at `monitoring/prometheus.yml`.

```yaml
# monitoring/prometheus.yml
scrape_configs:
  - job_name: 'catalogizer-api'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

### Running Prometheus

Start Prometheus using containers:

```bash
podman run -d \
  --name prometheus \
  --network host \
  -v ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro \
  docker.io/prom/prometheus:latest
```

Prometheus is accessible at `http://localhost:9090`.

### Available Metrics

Catalogizer exposes the following metric categories:

**HTTP Metrics**
- `http_requests_total` -- Total request count by method, path, and status code
- `http_request_duration_seconds` -- Request latency histogram
- `http_requests_in_flight` -- Currently active requests

**Application Metrics**
- `catalogizer_scans_total` -- Total scan operations
- `catalogizer_scan_duration_seconds` -- Scan duration histogram
- `catalogizer_files_detected_total` -- Total files detected by media type
- `catalogizer_storage_sources_active` -- Number of active storage sources
- `catalogizer_websocket_connections` -- Active WebSocket connections

**Go Runtime Metrics**
- `go_goroutines` -- Number of goroutines
- `go_memstats_alloc_bytes` -- Allocated memory
- `go_gc_duration_seconds` -- Garbage collection duration

---

## Grafana Dashboards

### Setup

Start Grafana using containers:

```bash
podman run -d \
  --name grafana \
  --network host \
  -v ./monitoring/grafana:/etc/grafana/provisioning:ro \
  docker.io/grafana/grafana:latest
```

Grafana is accessible at `http://localhost:3001` (default credentials: `admin` / `admin`).

### Pre-Built Dashboards

The `monitoring/grafana/` directory contains dashboard JSON files that are automatically provisioned.

**API Performance Dashboard**
- Request rate by endpoint
- Response time percentiles (p50, p95, p99)
- Error rate by status code
- Requests in flight

**Media Detection Dashboard**
- Files detected per scan
- Detection throughput over time
- Media type distribution
- Scan duration trends

**System Health Dashboard**
- Go runtime memory usage
- Goroutine count
- GC pause duration
- Active storage source count
- WebSocket connection count

---

## Resource Monitoring

Catalogizer enforces resource limits to prevent overloading the host machine.

### Container Resource Limits

| Container | Max CPU | Max Memory |
|-----------|---------|------------|
| PostgreSQL | 1 core | 2 GB |
| catalog-api | 2 cores | 4 GB |
| catalog-web | 1 core | 2 GB |

### Monitoring Host Resources

```bash
# Container resource usage
podman stats --no-stream

# System load average
cat /proc/loadavg

# Total resource budget: max 4 CPUs, 8 GB RAM across all containers
```

---

## Alerting

Configure alerts in Grafana to notify you of issues before they impact users.

### Recommended Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| High error rate | `http_requests_total{status=~"5.."}` > 5% of total | Critical |
| Slow response time | p95 latency > 2 seconds | Warning |
| Storage source down | Source health check fails for 5 minutes | Warning |
| High memory usage | Container memory > 80% of limit | Warning |
| No active WebSocket connections | `catalogizer_websocket_connections` = 0 for 10 min | Info |

Configure notification channels (email, Slack, webhook) in Grafana's Alerting section.

---

## Log Monitoring

Set the `LOG_LEVEL` environment variable to control log verbosity.

| Level | Description |
|-------|-------------|
| `debug` | Verbose output including request details and internal state |
| `info` | Standard operational messages |
| `warn` | Potential issues that do not prevent operation |
| `error` | Failures requiring attention |

For production, use `info` level. Switch to `debug` when diagnosing specific issues.

```bash
# View container logs
podman logs catalog-api --tail 100 -f

# Filter for errors
podman logs catalog-api 2>&1 | grep -i error
```

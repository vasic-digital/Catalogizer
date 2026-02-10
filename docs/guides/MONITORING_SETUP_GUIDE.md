# Monitoring Setup Guide

## Overview

Catalogizer includes comprehensive monitoring using Prometheus for metrics collection and Grafana for visualization. This guide covers setup, configuration, and usage of the monitoring stack.

## Architecture

```
┌──────────────┐
│  Catalogizer │ ──metrics──> ┌────────────┐
│     API      │               │ Prometheus │
└──────────────┘               └─────┬──────┘
                                     │
                                     │ scrapes
                                     │
                               ┌─────▼──────┐
                               │  Grafana   │
                               └────────────┘
```

## Metrics Exposed

### HTTP Metrics
- `catalogizer_http_requests_total` - Total HTTP requests by method, path, status
- `catalogizer_http_request_duration_seconds` - Request duration histogram
- `catalogizer_http_active_connections` - Current active connections

### Database Metrics
- `catalogizer_db_queries_total` - Total database queries by operation and table
- `catalogizer_db_query_duration_seconds` - Query duration histogram
- `catalogizer_db_connections_active` - Active database connections
- `catalogizer_db_connections_idle` - Idle database connections

### Media Processing Metrics
- `catalogizer_media_files_scanned_total` - Total files scanned
- `catalogizer_media_files_analyzed_total` - Total files analyzed
- `catalogizer_media_analysis_duration_seconds` - Analysis duration histogram
- `catalogizer_media_by_type` - Media count by type (movie, tv, music, etc.)

### External API Metrics
- `catalogizer_external_api_calls_total` - API calls by provider and status
- `catalogizer_external_api_call_duration_seconds` - API call duration

### Cache Metrics
- `catalogizer_cache_hits_total` - Cache hits by cache type
- `catalogizer_cache_misses_total` - Cache misses by cache type
- `catalogizer_cache_size_bytes` - Cache size in bytes

### WebSocket Metrics
- `catalogizer_websocket_connections_active` - Active WebSocket connections
- `catalogizer_websocket_messages_total` - Messages by direction (sent/received)

### Filesystem Metrics
- `catalogizer_filesystem_operations_total` - Operations by protocol, operation, status
- `catalogizer_filesystem_operation_duration_seconds` - Operation duration

### Storage Metrics
- `catalogizer_storage_roots_total` - Storage roots by protocol and status
- `catalogizer_storage_space_used_bytes` - Space used by root

### Authentication Metrics
- `catalogizer_auth_attempts_total` - Auth attempts by method and status
- `catalogizer_active_sessions` - Current active sessions

### Error Metrics
- `catalogizer_errors_total` - Errors by component and type

### Runtime Metrics
- `catalogizer_runtime_goroutines` - Number of goroutines
- `catalogizer_runtime_memory_alloc_bytes` - Allocated memory
- `catalogizer_runtime_memory_sys_bytes` - System memory
- `catalogizer_runtime_memory_heap_inuse_bytes` - Heap memory in use

### System Metrics
- `catalogizer_uptime_seconds` - Application uptime

## Quick Start

### 1. Configure Prometheus

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'catalogizer'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### 2. Start Prometheus

```bash
# Using Docker
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Using Podman
podman run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

### 3. Start Grafana

```bash
# Using Docker
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana

# Using Podman
podman run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana
```

### 4. Configure Grafana Data Source

1. Open Grafana: http://localhost:3000 (admin/admin)
2. Go to Configuration → Data Sources
3. Add Prometheus data source
4. Set URL to `http://localhost:9090`
5. Click "Save & Test"

### 5. Import Dashboard

1. Go to Dashboards → Import
2. Upload `config/grafana-dashboards/catalogizer-overview.json`
3. Select Prometheus data source
4. Click Import

## Health Checks

Catalogizer exposes three health check endpoints:

### Liveness Probe
**Endpoint:** `GET /health/live`

Checks if the service is running. Always returns 200 if the process is alive.

**Usage:**
```bash
curl http://localhost:8080/health/live
```

**Kubernetes:**
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

### Readiness Probe
**Endpoint:** `GET /health/ready`

Checks if the service is ready to serve traffic (database connected, etc.).

**Usage:**
```bash
curl http://localhost:8080/health/ready
```

**Kubernetes:**
```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

### Startup Probe
**Endpoint:** `GET /health/startup`

Checks if the service has completed initialization.

**Usage:**
```bash
curl http://localhost:8080/health/startup
```

**Kubernetes:**
```yaml
startupProbe:
  httpGet:
    path: /health/startup
    port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
  failureThreshold: 30
```

### Detailed Health Check
**Endpoint:** `GET /health`

Returns detailed health status of all components.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "components": {
    "database": {
      "status": "healthy",
      "latency": "2.5ms"
    }
  }
}
```

## Production Deployment

### Docker Compose

```yaml
version: '3.8'

services:
  catalogizer:
    image: catalogizer:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://...
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health/live"]
      interval: 30s
      timeout: 10s
      retries: 3

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./config/grafana-dashboards:/etc/grafana/provisioning/dashboards
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false

volumes:
  prometheus-data:
  grafana-data:
```

### Kubernetes

```yaml
apiVersion: v1
kind: Service
metadata:
  name: catalogizer-metrics
  labels:
    app: catalogizer
spec:
  selector:
    app: catalogizer
  ports:
    - name: metrics
      port: 8080
      targetPort: 8080
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: catalogizer
  labels:
    app: catalogizer
spec:
  selector:
    matchLabels:
      app: catalogizer
  endpoints:
    - port: metrics
      path: /metrics
      interval: 30s
```

## Alerting

### Example Prometheus Alert Rules

Create `alerts.yml`:

```yaml
groups:
  - name: catalogizer_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(catalogizer_errors_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors/sec"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(catalogizer_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "P95 response time is {{ $value }} seconds"

      - alert: DatabaseDown
        expr: up{job="catalogizer"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Catalogizer is down"
          description: "Catalogizer has been down for more than 1 minute"

      - alert: HighMemoryUsage
        expr: catalogizer_runtime_memory_alloc_bytes > 1e9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value | humanize }}B"

      - alert: LowCacheHitRate
        expr: sum(rate(catalogizer_cache_hits_total[5m])) / (sum(rate(catalogizer_cache_hits_total[5m])) + sum(rate(catalogizer_cache_misses_total[5m]))) < 0.5
        for: 10m
        labels:
          severity: info
        annotations:
          summary: "Low cache hit rate"
          description: "Cache hit rate is {{ $value | humanizePercentage }}"
```

Update `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

rule_files:
  - 'alerts.yml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

scrape_configs:
  - job_name: 'catalogizer'
    static_configs:
      - targets: ['catalogizer:8080']
```

## Dashboard Features

The included Grafana dashboard provides:

1. **Request Rate** - Real-time requests per second
2. **Uptime** - Application uptime
3. **HTTP Requests by Method** - GET, POST, PUT, DELETE breakdown
4. **Response Time** - P50 and P95 latency by endpoint
5. **Memory Usage** - Heap allocation and system memory
6. **Connections** - Goroutines, database connections, WebSocket connections
7. **Media by Type** - Distribution of media items
8. **Cache Hit Rate** - Cache performance over time

## Best Practices

### Metrics Collection

1. **Use Labels Wisely**
   - Keep cardinality low (avoid user IDs in labels)
   - Use meaningful label names
   - Group similar metrics

2. **Set Appropriate Buckets**
   - Histogram buckets should cover expected range
   - Use exponential buckets for wide ranges

3. **Monitor Resource Usage**
   - Track memory, CPU, connections
   - Set alerts for threshold violations

### Dashboard Design

1. **Organize by Use Case**
   - Create separate dashboards for different audiences
   - Group related metrics together

2. **Use Consistent Time Ranges**
   - Default to recent data (last hour)
   - Allow flexible time selection

3. **Add Context**
   - Include panel titles and descriptions
   - Use units and formatting

### Alerting

1. **Avoid Alert Fatigue**
   - Only alert on actionable issues
   - Use appropriate severity levels

2. **Set Meaningful Thresholds**
   - Base on actual usage patterns
   - Review and adjust regularly

3. **Include Helpful Context**
   - Provide clear descriptions
   - Link to runbooks

## Troubleshooting

### Metrics Not Showing

1. **Check if Prometheus is scraping:**
   ```bash
   curl http://localhost:9090/targets
   ```

2. **Verify metrics endpoint:**
   ```bash
   curl http://localhost:8080/metrics
   ```

3. **Check Prometheus logs:**
   ```bash
   docker logs prometheus
   ```

### High Memory Usage

Monitor these metrics:
- `catalogizer_runtime_memory_alloc_bytes`
- `catalogizer_runtime_goroutines`
- `catalogizer_db_connections_active`

Use Go profiling for detailed analysis:
```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Slow Queries

Monitor these metrics:
- `catalogizer_db_query_duration_seconds`
- `catalogizer_http_request_duration_seconds`

Enable query logging in database and analyze slow queries.

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)

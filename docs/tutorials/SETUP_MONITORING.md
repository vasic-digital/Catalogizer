# Set Up Monitoring

This tutorial walks through enabling Catalogizer's monitoring stack with Prometheus and Grafana, exploring dashboards, and creating alerts.

## Prerequisites

- Catalogizer running via Docker Compose (see [Quick Start](QUICK_START.md))
- Docker Compose with profile support (Docker Compose v2+)

## Overview

Catalogizer includes an optional monitoring profile in the Docker Compose configuration that deploys:

- **Prometheus** for metrics collection and storage (port 9090)
- **Grafana** for visualization and alerting (port 3001)
- Pre-configured dashboards for Catalogizer API metrics

## Step 1: Enable the Monitoring Profile

Start the monitoring services alongside the core stack:

```bash
# If starting fresh (all services)
docker compose --profile monitoring up -d

# If core services are already running, add monitoring
docker compose --profile monitoring up -d prometheus grafana
```

**Expected result:** Two additional containers start: `catalogizer-prometheus` and `catalogizer-grafana`.

Verify both are running:

```bash
docker compose --profile monitoring ps
```

## Step 2: Verify Prometheus Is Scraping Metrics

Open Prometheus at http://localhost:9090 or verify from the command line:

```bash
# Check Prometheus health
curl -sf http://localhost:9090/-/healthy && echo "Prometheus is healthy"

# Check that the catalog-api target is being scraped
curl -s http://localhost:9090/api/v1/targets | python3 -m json.tool
```

**Expected result:** Prometheus reports as healthy. The targets endpoint shows the `catalog-api` job with `health: "up"`.

You can also verify the API exposes metrics:

```bash
curl http://localhost:8080/metrics
```

This returns Prometheus-format metrics including HTTP request counts, latencies, Go runtime stats, and custom Catalogizer metrics.

## Step 3: Access Grafana

Open Grafana at http://localhost:3001.

Default credentials:
- **Username:** admin
- **Password:** admin (or the value of `GRAFANA_PASSWORD` in your `.env`)

You will be prompted to change the password on first login.

**Expected result:** The Grafana login page loads. After logging in, you see the Grafana home screen.

## Step 4: Explore Pre-Configured Dashboards

Catalogizer ships with a pre-provisioned dashboard. Navigate to it:

1. Click the hamburger menu (three lines) in the top-left corner
2. Click **Dashboards**
3. Look for the **Catalogizer Overview** dashboard in the provisioned folder

The overview dashboard includes panels for:
- API request rate and latency
- HTTP status code distribution
- Active connections and resource utilization
- Go runtime metrics (goroutines, memory, GC)

**Expected result:** The dashboard loads with live data from Prometheus. Panels populate within the scrape interval (15 seconds by default).

## Step 5: Explore Available Metrics

In Grafana, use the **Explore** view (compass icon in the sidebar) to query metrics directly.

Useful queries to try:

```promql
# Total HTTP requests per endpoint
rate(http_requests_total[5m])

# Average request latency (95th percentile)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Current number of goroutines
go_goroutines

# Memory usage
process_resident_memory_bytes
```

**Expected result:** Graphs showing real-time metric data from the Catalogizer API.

## Step 6: Create a Custom Alert

Set up an alert for high API latency:

1. In Grafana, go to **Alerting** > **Alert rules** (bell icon in the sidebar)
2. Click **New alert rule**
3. Configure the rule:
   - **Name:** High API Latency
   - **Query A:** `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
   - **Condition:** Is above 2 (seconds)
   - **Evaluate every:** 1m
   - **For:** 5m
4. Add labels and annotations as needed
5. Click **Save rule and exit**

**Expected result:** The alert rule appears in the alert rules list. When the 95th percentile latency exceeds 2 seconds for 5 minutes, the alert fires.

Additional alert ideas:
- **High error rate:** `rate(http_requests_total{status=~"5.."}[5m]) > 0.1`
- **API down:** `up{job="catalog-api"} == 0`
- **High memory usage:** `process_resident_memory_bytes > 1.5e9`

## Step 7: Configure Alert Notifications (Optional)

To receive alert notifications via email, Slack, or other channels:

1. Go to **Alerting** > **Contact points**
2. Click **New contact point**
3. Select the integration type (Email, Slack, Discord, PagerDuty, etc.)
4. Configure the connection details
5. Click **Save contact point**
6. Go to **Alerting** > **Notification policies** and route alerts to the new contact point

**Expected result:** When alerts fire, notifications are sent to the configured channel.

## Monitoring Architecture

```
Catalogizer API (/metrics)
        |
        v
  Prometheus (scrapes every 15s)
        |
        v
  Grafana (queries Prometheus)
        |
        v
  Dashboards + Alerts
```

Configuration files in the repository:
- `monitoring/prometheus.yml` - Prometheus scrape configuration
- `monitoring/grafana/provisioning/datasources/` - Grafana datasource configuration
- `monitoring/grafana/provisioning/dashboards/` - Dashboard provisioning
- `monitoring/grafana/dashboards/catalogizer-overview.json` - The pre-built dashboard

## Environment Variables

Configure monitoring ports and credentials in `.env`:

```env
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001
GRAFANA_USER=admin
GRAFANA_PASSWORD=your_secure_password
```

## Troubleshooting

### Prometheus shows target as "down"

- Verify the API container is running: `docker compose ps api`
- Check that the API exposes metrics: `curl http://localhost:8080/metrics`
- Ensure both containers are on the same Docker network (`catalogizer-network`)
- Check Prometheus logs: `docker compose logs prometheus`

### Grafana dashboard shows "No data"

- Verify Prometheus is scraping successfully (Step 2)
- Check the time range selector in Grafana (top-right corner) -- data may not exist for the selected range
- Wait at least 30 seconds after starting monitoring for data to appear
- In Explore view, verify you can query `up` and get results

### Cannot access Grafana at port 3001

- Verify the Grafana container is running: `docker compose --profile monitoring ps`
- Check for port conflicts: `ss -tlnp | grep 3001`
- Review Grafana logs: `docker compose logs grafana`

### Prometheus storage growing too large

- Default retention is 200 hours (about 8 days)
- Adjust in `docker-compose.yml` under the Prometheus service: `--storage.tsdb.retention.time=200h`
- To reduce storage, lower the retention period or reduce the scrape interval

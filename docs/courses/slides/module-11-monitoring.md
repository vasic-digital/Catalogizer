# Module 11: Monitoring and Observability - Slide Deck Outline

**Total Slides**: 10
**Estimated Duration**: 45 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Monitoring and Observability

- Prometheus metrics, Grafana dashboards, memory monitoring, health aggregation
- Prerequisites: Module 8 completed
- By the end: configure metrics, detect memory leaks, set up health aggregation

---

## Slide 2: Prometheus Metrics Architecture (5 min)

**Title**: What Gets Measured

- internal/metrics/metrics.go defines all custom metrics
- Middleware records: request count, latency histogram, status code distribution
- Database metrics: query count, connection pool utilization, migration status
- Scan metrics: files processed, scan duration, entities created
- Exposed at /metrics endpoint in Prometheus text format
- Exercise reference: Exercise 11.1 -- query Prometheus for API latency

---

## Slide 3: Configuring Prometheus Scraping (5 min)

**Title**: monitoring/prometheus.yml

- Scrape interval and evaluation interval configuration
- Target: catalog-api /metrics endpoint
- Label rewriting for multi-instance deployments
- Alert rules for service down, high error rate, latency breach
- Demo: start Prometheus container and verify target is up

---

## Slide 4: Memory Monitor Module (5 min)

**Title**: Heap Tracking and Goroutine Monitoring

- digital.vasic.memory module: heap snapshots, goroutine count tracking
- Stability tests compare heap snapshots over time to detect growth
- Memory leak protections: rate limiter bucket cap, log entry cap, event channel drain
- scripts/memory-leak-check.sh for automated profiling
- Exercise reference: Exercise 11.2 -- run a memory stability test

---

## Slide 5: Go Runtime Memory Metrics (4 min)

**Title**: Interpreting Memory Data at /metrics

- go_memstats_heap_alloc_bytes: current heap allocation
- go_memstats_gc_pause_total_ns: GC pause duration
- go_goroutines: number of active goroutines
- process_resident_memory_bytes: total process memory
- Threshold alerts: heap growth > 20% over 10 minutes, goroutines > 1000

---

## Slide 6: Health Aggregator (5 min)

**Title**: Composite Service Health Status

- Individual health checks: database, cache, filesystem, WebSocket
- Composite health endpoint rolls up all checks into single status
- Database pool health: MaxOpen=25, MaxIdle=10, MaxLifetime=5m
- Circuit breaker states for SMB: closed (ok), open (failed), half-open (testing)
- Redis cache connection pool metrics
- Demo: query health endpoint and inspect each subsystem

---

## Slide 7: Grafana Dashboards (5 min)

**Title**: Pre-Built Visualization Panels

- Deploy from monitoring/grafana/ and config/grafana-dashboards/
- API Overview: request rate, error rate, latency percentiles
- Database: query throughput, connection pool usage, slow queries
- Media Pipeline: scan progress, entity creation rate, provider latency
- Infrastructure: CPU, memory, disk, network per container
- Exercise reference: Exercise 11.3 -- deploy Grafana and import dashboards

---

## Slide 8: Alerting With Grafana and AlertManager (4 min)

**Title**: Configuring Alerts for Production

- Grafana alert rules for visual threshold-based alerting
- AlertManager routes alerts to email and webhook receivers
- Severity levels: critical (immediate), warning (next business day)
- Inhibit rules: critical suppresses matching warning alerts
- Repeat interval: 12 hours to balance awareness and fatigue

---

## Slide 9: Load Test Metrics Validation (4 min)

**Title**: Using k6 to Generate Controlled Load

- tests/k6/load_test.js: verify p95 < 500ms under 50 concurrent users
- tests/k6/stress_test.js: find the breaking point at 300 users
- tests/k6/soak_test.js: 30-minute soak test for memory leak detection
- Run via Podman: podman run --rm --network host grafana/k6
- Correlate k6 results with Prometheus/Grafana metrics during the test
- Exercise reference: Exercise 11.4 -- run load test and review dashboard

---

## Slide 10: Module Summary (3 min)

**Title**: What We Covered

- Prometheus metrics: request, database, scan, and runtime metrics
- Memory module: heap tracking, goroutine monitoring, leak detection
- Health aggregator: composite status across all subsystems
- Grafana dashboards with pre-built panels for all components
- k6 load testing correlated with monitoring data
- Next module: Certification and Assessment Preparation

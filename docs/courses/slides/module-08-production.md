# Module 8: Deployment and Production - Slide Deck Outline

**Total Slides**: 13
**Estimated Duration**: 55 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Deployment and Production

- Container deployment, infrastructure, security scanning, monitoring, maintenance
- Prerequisites: Modules 5 and 6 completed
- By the end: deploy to production with monitoring and security scanning

---

## Slide 2: Production Container Deployment (5 min)

**Title**: Deploying With Podman Compose

- docker-compose.yml for production with resource limits and health checks
- Required env vars: POSTGRES_PASSWORD, JWT_SECRET, DB_ENCRYPTION_KEY
- Container resource budget: max 4 CPUs, 8 GB RAM total
- podman-compose up -d and verify service health
- Critical: use --network host for builds, fully qualified image names
- Exercise reference: Exercise 8.1 -- deploy the production stack

---

## Slide 3: Docker Compose Files (4 min)

**Title**: Choosing the Right Compose Configuration

- docker-compose.yml: production stack
- docker-compose.dev.yml: development with hot reloading
- docker-compose.build.yml: containerized build pipeline
- docker-compose.test.yml: test stack (network_mode: host)
- docker-compose.security.yml: security scanning tools
- Demo: compare resource limits across configurations

---

## Slide 4: Nginx Reverse Proxy (5 min)

**Title**: Configuring Nginx for Production

- config/nginx.conf and config/nginx/catalogizer.prod.conf
- TLS/SSL termination at the Nginx layer
- HTTP/3 (QUIC) with Brotli compression (mandatory)
- Proxy /api to catalog-api, serve catalog-web static assets
- Docker Compose volume mounts reference config/ directory
- Exercise reference: Exercise 8.2 -- configure TLS with a self-signed cert

---

## Slide 5: Redis Caching (4 min)

**Title**: Configuring Redis for Performance

- config/redis.conf for cache settings
- Optional caching layer via go-redis/v9
- Container limits: --cpus=1 --memory=2g
- Cache invalidation on data mutations
- Connection pool health monitoring

---

## Slide 6: PostgreSQL Production Setup (4 min)

**Title**: Database Configuration for Production

- DB_TYPE=postgres with host, port, name, user, password env vars
- Container port mapping: 5432 -> 5433
- Connection pool: MaxOpen=25, MaxIdle=10, MaxLifetime=5m, MaxIdleTime=3m
- Migrations run automatically on startup
- Encrypted at rest with DB_ENCRYPTION_KEY (SQLCipher for SQLite)

---

## Slide 7: Systemd Services (4 min)

**Title**: Bare-Metal Deployment With Systemd

- config/systemd/catalogizer-api.service for service management
- Automatic restart on failure with configurable delay
- Environment file for secrets (not in unit file)
- Journal logging integration for log aggregation
- Exercise reference: Exercise 8.3 -- create a systemd service unit

---

## Slide 8: Security Scanning Pipeline (5 min)

**Title**: Automated Security Scanning

- govulncheck for Go stdlib/dependency vulnerabilities
- npm audit for frontend dependency issues
- Semgrep SAST: podman-compose -f docker-compose.security.yml
- SonarQube: ./scripts/run-sonarqube-scan.sh with quality gate
- Snyk and Trivy for container image scanning
- Demo: run the full security scan and interpret results

---

## Slide 9: Production Monitoring (5 min)

**Title**: Prometheus and Grafana in Production

- Prometheus scraping configured in monitoring/prometheus.yml
- Grafana dashboards in monitoring/grafana/ and config/grafana-dashboards/
- Key alerts: service down, disk space low, error rate spike, latency breach
- SMB circuit breaker state monitoring
- Host resource monitoring: podman stats, /proc/loadavg
- Exercise reference: Exercise 8.4 -- set up alerting for API errors

---

## Slide 10: AlertManager Configuration (4 min)

**Title**: Routing Alerts to Email and Webhooks

- AlertManager groups alerts by alertname and severity
- Email notifications for all alerts (default receiver)
- Webhook notifications for critical alerts (API integration)
- Inhibit rules: critical alerts suppress matching warnings
- Repeat interval: 12 hours to avoid alert fatigue

---

## Slide 11: Resource Limits (4 min)

**Title**: Host Resource Protection (30-40% Maximum)

- Total host resource cap: 30-40% of CPU and memory
- PostgreSQL: --cpus=1 --memory=2g
- API: --cpus=2 --memory=4g
- Web: --cpus=1 --memory=2g
- Builder: --cpus=3 --memory=8g
- Challenges run sequentially, never in parallel

---

## Slide 12: Maintenance and Upgrades (4 min)

**Title**: Zero-Downtime Upgrades and Disaster Recovery

- Container rolling updates for zero-downtime deploys
- Three-tier backup: daily database, weekly config, monthly verification
- Database migrations handled by the migrations system
- Disaster recovery: database corruption, config loss, media source failure
- Exercise reference: Exercise 8.5 -- perform a rolling upgrade

---

## Slide 13: Module Summary and Next Steps (3 min)

**Title**: What We Covered

- Production deployment with Podman Compose and resource limits
- Nginx reverse proxy with HTTP/3 and Brotli compression
- Redis caching and PostgreSQL production configuration
- Security scanning pipeline: govulncheck, Semgrep, SonarQube, Snyk
- Monitoring with Prometheus, Grafana, and AlertManager
- Next module: Architecture Deep Dive

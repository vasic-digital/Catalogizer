# Support

Get help with Catalogizer through the following channels and resources.

---

## Documentation

The best starting point for resolving issues is the project documentation.

### User Documentation
- [Quick Start Guide](../tutorials/QUICK_START.md) -- Install and browse your catalog in 10 minutes
- [User Guide](../USER_GUIDE.md) -- Comprehensive guide to the web interface
- [Configuration Guide](../CONFIGURATION_GUIDE.md) -- All configuration options and environment variables

### Platform Guides
- [Web Application Guide](../guides/WEB_APP_GUIDE.md) -- Using the React web interface
- [Android Guide](../guides/ANDROID_GUIDE.md) -- Android mobile app
- [Android TV Guide](../guides/ANDROID_TV_GUIDE.md) -- Android TV app
- [Desktop Guide](../guides/DESKTOP_GUIDE.md) -- Desktop application
- [Installer Wizard Guide](../guides/INSTALLER_WIZARD_GUIDE.md) -- Setup wizard

### Administration
- [Installation Guide](../INSTALLATION_GUIDE.md) -- Detailed installation for all components
- [Deployment Guide](../deployment/DEPLOYMENT.md) -- Production deployment strategies
- [Monitoring Guide](../deployment/MONITORING_GUIDE.md) -- Prometheus and Grafana setup
- [Backup and Recovery](../deployment/BACKUP_AND_RECOVERY.md) -- Data protection procedures

### Developer Resources
- [Architecture Overview](../architecture/ARCHITECTURE.md) -- System design and components
- [API Documentation](../api/API_DOCUMENTATION.md) -- REST API endpoint reference
- [WebSocket Events](../api/WEBSOCKET_EVENTS.md) -- Real-time event documentation
- [Contributing Guide](../CONTRIBUTING.md) -- How to contribute to the project

---

## Troubleshooting

### Guides
- [Troubleshooting Guide](../TROUBLESHOOTING_GUIDE.md) -- Common issues and step-by-step solutions
- [Troubleshooting Reference](../guides/TROUBLESHOOTING.md) -- Additional troubleshooting tips

### Quick Fixes for Common Issues

**Cannot connect to the backend**
- Verify the backend is running: `curl http://localhost:8080/api/v1/health`
- Check the PORT setting in your `.env` file
- Ensure no firewall is blocking port 8080

**Frontend shows "connection refused"**
- Verify `VITE_API_BASE_URL` in `.env.local` points to the correct backend URL
- If using containers, ensure the backend container is healthy: `podman ps` or `docker ps`

**No media detected after connecting a source**
- Check the backend logs for scanning errors: set `LOG_LEVEL=debug` in `.env`
- Verify the storage source is accessible with the provided credentials
- Ensure the scan has completed -- large sources may take time on first scan

**SMB share keeps disconnecting**
- Review circuit breaker logs for state transitions
- Adjust resilience parameters: `SMB_RETRY_ATTEMPTS`, `SMB_RETRY_DELAY_SECONDS`, `SMB_HEALTH_CHECK_INTERVAL`
- Verify network stability between the Catalogizer server and the SMB share

**Missing metadata for detected media**
- Verify external API keys are configured: `TMDB_API_KEY`, `OMDB_API_KEY`
- Check for rate limiting from external providers in the backend logs
- Some media types may not have metadata available from external providers

**Login fails after server restart**
- Verify `JWT_SECRET` has not changed between restarts
- Existing tokens are invalidated if the JWT secret changes
- Users need to log in again after a secret rotation

**Docker/Podman containers fail to start**
- Verify `POSTGRES_PASSWORD` is set in `.env` (required)
- For Podman: use fully qualified image names (e.g., `docker.io/library/postgres:15-alpine`)
- Check container logs: `podman logs <container-name>` or `docker logs <container-name>`

---

## Reporting Issues

### Before Reporting

1. Check the [Troubleshooting Guide](../TROUBLESHOOTING_GUIDE.md) for known solutions
2. Search existing issues to avoid duplicates
3. Try the latest version -- the issue may already be fixed

### What to Include in a Report

When reporting an issue, include the following information to help with diagnosis:

- **Component**: Which component is affected (catalog-api, catalog-web, Android, desktop, etc.)
- **Version**: The version or commit hash you are running
- **Environment**: Operating system, container runtime (Podman/Docker), browser version
- **Steps to reproduce**: Minimal steps to trigger the issue
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Logs**: Relevant log output (set `LOG_LEVEL=debug` for detailed logs)
- **Configuration**: Relevant `.env` settings (redact secrets)

### Log Collection

To collect detailed logs for a bug report:

```bash
# Backend logs with debug level
# Set LOG_LEVEL=debug in catalog-api/.env, then restart

# Container logs
podman logs catalog-api 2>&1 | tail -200
podman logs catalog-web 2>&1 | tail -200

# Frontend: Open browser Developer Tools -> Console tab
# Copy any error messages or warnings
```

---

## Self-Hosted Diagnostics

### Health Check Endpoints

```bash
# Backend health
curl http://localhost:8080/api/v1/health

# Database connectivity (via backend logs)
# Set LOG_LEVEL=debug and check startup messages

# Container health status
podman ps --format "table {{.Names}}\t{{.Status}}"
```

### Monitoring

For production deployments, set up the included monitoring stack:

- **Prometheus**: Collects metrics from `monitoring/prometheus.yml`
- **Grafana**: Visualizes metrics with pre-built dashboards in `monitoring/grafana/`
- Dashboards cover: API performance, media detection throughput, storage source health

See the [Monitoring Setup Tutorial](../tutorials/SETUP_MONITORING.md) for step-by-step instructions.

### Performance Diagnostics

```bash
# Run performance tests
cd catalog-api && go test -bench=. ./internal/media/providers/

# Check memory usage
scripts/memory-leak-check.sh

# Run stress tests
scripts/performance-test.sh
```

---

## Security Issues

If you discover a security vulnerability, do not report it publicly. Instead:

1. Review the [Security Testing Guide](../SECURITY_TESTING_GUIDE.md) to confirm the issue
2. Report the vulnerability privately with a detailed description
3. Include steps to reproduce and potential impact assessment
4. Allow time for a fix before any public disclosure

---

## Video Course

For structured learning, the Catalogizer video course covers six modules:

| Module | Topic | Duration |
|--------|-------|----------|
| Module 1 | Installation and Setup | ~45 minutes |
| Module 2 | Getting Started | ~75 minutes |
| Module 3 | Media Management | ~75 minutes |
| Module 4 | Multi-Platform | ~55 minutes |
| Module 5 | Administration | ~65 minutes |
| Module 6 | Developer Guide | ~70 minutes |

Course materials including scripts, slide outlines, and exercises are available in `docs/courses/`.

---

## Additional Resources

- [Changelog](CHANGELOG.md) -- Version history and recent changes
- [FAQ](FAQ.md) -- Answers to frequently asked questions
- [Features](FEATURES.md) -- Complete feature list
- [Download](DOWNLOAD.md) -- Installation options per platform
- [Documentation Hub](DOCUMENTATION.md) -- All documentation organized by audience

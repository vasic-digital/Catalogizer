# Changelog

All notable changes to the Catalogizer project are documented here. This page provides a user-facing summary of version history. For the full technical changelog, see the [development changelog](../CHANGELOG.md).

---

## Version 1.1.0 -- March 23, 2026

A comprehensive remediation and documentation release spanning 12 phases, covering concurrency safety, security hardening, performance optimization, monitoring, and complete documentation coverage.

### Concurrency and Reliability

- Goroutine lifecycle management with context-based cancellation across all background workers
- Bounded parallelism using semaphore patterns in scanner, asset manager, and middleware
- ConcurrencyLimiter middleware caps in-flight HTTP requests at 100
- Idempotent CacheService Close() pattern prevents goroutine leaks on shutdown
- Lazy initialization via the digital.vasic.lazy module for deferred expensive operations
- Memory leak detection via the digital.vasic.memory module
- Circuit breaker recovery patterns via the digital.vasic.recovery module

### Security Scanning

- Six-tool security scanning stack: govulncheck, Semgrep, SonarQube, Snyk, Trivy, npm audit
- Zero known vulnerabilities in Go dependencies (govulncheck verified)
- Zero critical or production vulnerabilities in npm dependencies
- Input validation middleware with configurable sanitization rules
- Security headers middleware (X-Frame-Options, CSP, HSTS)

### Performance

- Database connection pooling with configurable MaxOpen (25), MaxIdle (10), ConnMaxLifetime (5 min)
- SQLite WAL mode with explicit PRAGMA (go-sqlcipher ignores connection string pragmas)
- Migration v9: performance indexes on files (file_type, extension, name), media_items (title+type compound), user_metadata (user+watched_status compound)
- Unique index on media_files(media_item_id, file_id) with automatic deduplication
- k6 load test suite: load, stress, soak, and spike test scenarios

### Monitoring and Observability

- Runtime metrics collector sampling goroutines, heap, and GC every 15 seconds
- Pre-built Grafana dashboard panels for request rate, latency, Go runtime, and application metrics
- Prometheus alerting rules for error rate, latency, goroutine leak, memory growth, and service down
- Structured JSON logging via Zap with field-level search support
- Built-in log management API: collection, analysis, sharing, real-time streaming

### Documentation

- Complete data dictionary documenting all 32 database tables with columns, types, constraints, relationships, and indexes
- API changelog cataloging all REST endpoints (120+) grouped by domain with HTTP methods and descriptions
- SQLite-to-PostgreSQL migration guide with step-by-step export/import procedures
- Disaster recovery guide with backup procedures (SQLite, PostgreSQL), restore steps, verification, and scheduling
- Security incident response plan: detection, containment, eradication, recovery, post-incident review
- Troubleshooting guide expanded with goroutine leak detection, database lock issues, WebSocket problems, cache invalidation, and container networking
- Four new video course modules (15-18): Concurrency Patterns, Security Scanning, Load Testing, Monitoring
- Updated website features, changelog, and FAQ pages

### Infrastructure

- 32 git submodules: 29 original + 3 new (Lazy, Memory, Recovery) from decoupling refactoring
- 22 replace directives in go.mod wiring reusable modules
- Container build pipeline producing all 7 components in approximately 17 minutes
- Host resource limits enforced: 30-40% maximum across all workloads

---

## Version 1.0.0 -- February 2, 2026

The first stable release of Catalogizer, delivering a complete multi-platform media collection management system.

### Highlights

- Full multi-protocol support: SMB/CIFS, FTP/FTPS, NFS, WebDAV, and local filesystem
- Automated media detection pipeline identifying 50+ media types
- Metadata enrichment from TMDB, IMDB, TVDB, MusicBrainz, Spotify, and Steam
- Seven platform components: backend API, web frontend, desktop app, installer wizard, Android app, Android TV app, API client library
- Real-time updates via WebSocket across all connected clients

### Multi-Platform Clients

- **catalog-api**: Go/Gin backend with REST API, JWT authentication, and SQLite/PostgreSQL support
- **catalog-web**: React/TypeScript web frontend with real-time updates, analytics dashboard, and responsive design
- **catalogizer-desktop**: Cross-platform native app built with Tauri (Rust + React) for Windows, macOS, and Linux
- **installer-wizard**: Guided setup tool with network discovery and connection validation
- **catalogizer-android**: Android mobile app with Kotlin/Compose, MVVM architecture, and offline mode
- **catalogizer-androidtv**: Android TV app with leanback UI and D-pad navigation
- **catalogizer-api-client**: TypeScript API client library for custom integrations

### Storage and Detection

- Five storage protocol support through a unified client interface
- SMB resilience: circuit breaker, exponential backoff retry, and offline caching
- Automatic SMB share discovery on the local network
- Quality analysis: resolution, codec, bitrate detection with version tracking
- Duplicate detection across storage sources

### Media Player

- Built-in video and audio playback in the browser
- Playback position tracking with cross-device sync
- Subtitle management: SRT, ASS, SSA, VTT with auto-matching
- Lyrics display during music playback
- Cover art fetching from multiple providers
- Deep linking to specific playback positions

### Organization

- Favorites with JSON and CSV import/export
- Collections: Manual, Smart (rule-based auto-population), and Dynamic
- Playlists with drag-and-drop reordering and auto-advancement
- Format conversion with batch queue and real-time progress

### Security

- JWT authentication with configurable token expiry
- Role-based access control (Admin, Moderator, User, Viewer)
- SQLCipher database encryption (AES-256)
- Two-factor authentication support
- Security scanning with Snyk and SonarQube

### Monitoring and Analytics

- Prometheus metrics with automatic HTTP instrumentation
- Pre-built Grafana dashboards for API performance and storage health
- Analytics dashboard with library composition, growth trends, and quality analysis
- AI Dashboard for intelligent insights

### Documentation

- Architecture guides for all components
- API documentation with endpoint reference
- Platform-specific guides for web, desktop, Android, and Android TV
- Deployment, monitoring, backup, and troubleshooting guides
- Video course with six modules covering installation through development
- Contributing guide with code standards and testing requirements

### Infrastructure

- Docker and Podman support with development, production, and build compose configurations
- Nginx reverse proxy configuration for production
- Redis caching with custom configuration
- Systemd service file for bare-metal deployment
- Submodule architecture with 19 Go, 2 TypeScript, and 1 Kotlin reusable modules

---

## Upcoming

Features planned for future releases:

- iOS application
- Apple TV application
- Expanded cloud storage provider integrations
- Machine learning-based media classification
- Collaborative collections with shared editing
- Plugin system for community extensions

For the full roadmap, see the [project roadmap](../roadmap/CATALOGIZER_ROADMAP_V3.md).

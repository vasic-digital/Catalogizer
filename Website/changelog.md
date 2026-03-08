---
title: Changelog
description: Version history and notable changes to the Catalogizer project
---

# Changelog

All notable changes to the Catalogizer project are documented here. This page provides a user-facing summary of version history.

---

## Version 1.1.0 -- March 8, 2026

Comprehensive remediation, security hardening, and feature expansion.

### New Features

- **Search API**: Full-text search with advanced filters, duplicate detection, and paginated results
- **Browse API**: Storage root browsing and directory listing with content type detection
- **Cloud Sync API**: Synchronization with Amazon S3 and Google Cloud Storage
- **Prometheus Metrics**: HTTP request metrics, DB query duration, runtime metrics, Grafana dashboard
- **28 new challenges** (CH-061 to CH-088): Feature validation, security, performance, resilience, observability
- **6 module functional challenges** (MOD-016 to MOD-021): Lazy, Recovery, Memory module verification

### Security

- Security headers validation (X-Frame-Options, X-Content-Type-Options, CSP, HSTS)
- CORS origin validation and rejection of unauthorized origins
- Input validation rejecting SQL injection, XSS, and path traversal
- Rate limiting on authentication endpoints
- JWT token lifecycle validation
- File upload magic bytes verification

### Performance

- API response latency benchmarks
- Concurrent request handling validation
- Graceful degradation under load
- Memory stability during load testing
- Database connection pool recovery

### Architecture

- 3 new Go modules: Lazy (generic lazy loading), Memory (leak detection), Recovery (circuit breaker)
- Total Go modules: 29 (up from 19)
- Module functional verification challenges validate specific capabilities
- 285+ registered challenges (up from 249)

### Documentation

- 10 new documentation files: API reference (Search, Browse, Sync), Security (headers, CORS, secrets), Architecture (lazy loading, concurrency), Guides (performance tuning), Testing (stress results)
- 4 new video course modules (13-14) with slide decks (9-10)
- 11 CLAUDE.md files for TS/React and other submodules
- Comprehensive remediation report

### Test Coverage Improvements

- Database coverage: 61.9% → 90.8% (+28.9%)
- Config coverage: 73.8% → 92.9% (+19.1%)
- Auth coverage: 74.4% → 84.8% (+10.4%)
- Internal/handlers coverage: 48.9% → 66.5% (+17.6%)
- 38/38 Go packages pass, 0 failures, 0 races
- 102 frontend test files, 1795 tests pass

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
- Expanded cloud storage provider integrations (beyond S3 and GCS)
- Machine learning-based media classification
- Collaborative collections with shared editing
- Plugin system for community extensions

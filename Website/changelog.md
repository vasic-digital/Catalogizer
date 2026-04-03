---
title: Changelog
description: Version history and notable changes to the Catalogizer project
---

# Changelog

All notable changes to the Catalogizer project are documented here. This page provides a user-facing summary of version history.

---

## v2.2.0 (2026-04-03)

### New Features
- Dynamic container runtime detection: automatically selects Podman, Docker, or other available runtimes with graceful fallback
- Android offline mode activation with Room database caching and automatic background sync
- Android biometric authentication (fingerprint and face unlock) on supported devices
- Android WebSocket real-time updates: live scan progress, new media notifications, and source status changes
- Panoptic cloud analytics implementation for aggregated usage metrics and fleet-wide insights
- 4 new k6 stress test scripts: load, stress, soak, and spike scenarios
- 10 Mermaid architecture diagrams covering system topology, data flow, media pipeline, and protocol abstraction

### Challenge System
- 557+ challenges registered (stress, integration, security, documentation)
- Expanded coverage for all platform targets (API, web, desktop, mobile)

### Video Course
- 17 modules: 7 user/admin modules plus 10 developer modules
- Complete restructuring with user course, advanced modules, and developer course tracks

### Security
- Comprehensive security scanning pipeline: SonarQube, Snyk, Semgrep SAST, Trivy, OWASP Dependency Check
- govulncheck: 0 vulnerabilities
- npm audit: 0 vulnerabilities across all components
- Semgrep custom rules: 0 findings

### Test Coverage
- Test coverage maximization across all modules (>85% target)
- Go 44/44 packages pass, Frontend 130/130 files pass
- HelixQA: 1228+ total test cases including negative and cross-platform scenarios

### Infrastructure
- Dockerfile updated: Go 1.25, 13 submodule COPYs, proper user/group creation
- PostgreSQL datetime handling fixes, migration column gap corrections
- Build 16 verified across all 7 components

---

## v2.1.0 (2026-03-31)

### New Features
- 152 new challenges (CH-251 to CH-401): concurrency hardening, coverage gates, stress validation, documentation completeness, accessibility, API contract compliance
- Total challenge bank: 400+ (from 249)
- 6 new video course modules (25-30): Concurrency Hardening, Security Scanning in Practice, Test Coverage Mastery, Performance Monitoring, Module Architecture Deep Dive, Cross-Platform Consistency
- Real media quality breakdown by file extension (`/api/v1/media/stats` `by_quality` field)

### Test Coverage Expansion
- 82+ new frontend component tests (auth, layout, UI base, media, entity, dashboard, playlists, collections, pages, hooks)
- 30+ new TypeScript submodule tests across all 9 modules
- Go doc.go files for all 41 packages
- ARCHITECTURE.md for all submodules
- `.env` protection in `.gitignore` for all submodules

### Improvements
- `rows.Err()` check after database row iteration in media stats handler
- Removed `console.debug()` from production code (MemoCache.tsx)
- Replaced `value: any` with proper union type in CollectionRule
- Type-safe Select/Input value narrowing in SmartCollectionBuilder
- typeof guard for Date.parse in collection rule validation

### Security
- govulncheck: 0 vulnerabilities (verified with scan artifacts)
- npm audit: 0 vulnerabilities across all components
- Semgrep custom rules (8 rules): 0 real security findings
- Removed 2 `fmt.Printf("DEBUG: ...")` statements from challenge code
- Fixed stale test expecting panic on graceful nil-DB handling

### Documentation
- Video course expanded to 30 modules (was 24)
- Course outline updated to 16-18 hours
- Master Completion Plan v3: 13-phase hardening roadmap
- Session report with full verification results

---

## v2.0.0 (2026-03-30)

### New Features
- Backup management (create, list, restore database backups)
- Password change API
- Media quality analysis endpoint
- Media analysis and metadata refresh endpoints
- Popular media (sorted by favorites)
- Recent media browsing
- Media search by file path

### Improvements
- 12 Go modules integrated (database, lazy, media, memory, middleware, observability, ratelimiter, recovery, security, storage, streaming, watcher)
- Backup operation semaphore (prevents concurrent backups)
- Goroutine lifecycle management with WaitGroup + Close() methods
- Memory leak protections (rate limiter bucket cap, log entry cap, event channel drain)
- React Audio ref reuse, IntersectionObserver optimization
- Non-blocking event patterns verified across all subsystems

### Security
- Zero vulnerabilities (govulncheck, npm audit, all components)
- Path traversal protection on backup restore

### Documentation
- CLAUDE.md for catalog-api, catalog-web, catalogizer-android, catalogizer-androidtv
- AGENTS.md for 8 submodules
- Architecture docs for VisionEngine, LLMProvider, DocProcessor, LLMOrchestrator
- 42+ new unit tests

---

## 2026-03-26 -- Project Completion Release

### New Features
- OpenLibrary and MusicBrainz metadata providers (fully implemented, free APIs)
- Lazy service initialization via LazyServiceRegistry
- Semaphore-based concurrency control for scan operations
- HTTP connection pooling for external API calls
- k6 spike test scenario

### Security
- Semgrep SAST scanner added to security compose
- SonarQube scanner execution script
- Consolidated security scan report
- 2 data races fixed (analyzer, SMB resilience)

### Testing
- Challenge handler: 39 new tests
- ReportingService: 55+ new tests
- AnalyticsService: 30 new tests (120 subtests)
- SyncService: 38 new tests
- Goroutine leak detection test
- Memory stability test with heap tracking
- Playwright accessibility tests (axe-core)

### Documentation
- Course modules 9 (Advanced Features) and 10 (Troubleshooting)
- All incomplete docs completed or marked SUPERSEDED
- Updated platform guides and tutorials

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

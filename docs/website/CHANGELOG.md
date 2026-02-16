# Changelog

All notable changes to the Catalogizer project are documented here. This page provides a user-facing summary of version history. For the full technical changelog, see the [development changelog](../CHANGELOG.md).

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

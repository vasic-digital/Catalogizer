# Changelog

All notable changes to the Catalogizer project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- (No unreleased changes yet)

### Changed

### Deprecated

### Removed

### Fixed

### Security

---

## [1.0.0] - 2026-02-02

### Added

#### Multi-Platform Support
- Go/Gin backend API (`catalog-api`) with Handler, Service, and Repository layers
- React/TypeScript web frontend (`catalog-web`) with React Query state management
- Tauri/Rust desktop application (`catalogizer-desktop`) with IPC commands/events
- Tauri-based installer wizard (`installer-wizard`) for guided setup
- Kotlin/Compose Android application (`catalogizer-android`) with MVVM architecture
- Kotlin/Compose Android TV application (`catalogizer-androidtv`) with lean-back UI
- TypeScript API client library (`catalogizer-api-client`)

#### Protocol Support
- SMB/CIFS network share scanning with circuit breaker and offline cache
- FTP file system support
- NFS network filesystem support
- WebDAV protocol support with sync and backup capabilities
- Local filesystem scanning

#### Media Detection and Management
- Automated media detection pipeline: detector, analyzer, and metadata providers
- TMDB, IMDB, and other metadata provider integrations
- Comprehensive media type classification (movies, TV shows, music, games, software, and 30+ additional types)
- Duplicate detection with configurable hash algorithms (MD5, SHA256, SHA1, BLAKE3, quick hash)
- Quality profile comparison and ranking (4K/UHD, 1080p, 720p, DVD, etc.)
- Media collection grouping for series, albums, and franchises
- File rename detection system
- Advanced AI features for media analysis (Phase 3.2.8)
- Performance optimization system (Phase 3.2.7)
- Advanced collection management system (Phase 3.2.5-3.2.6)

#### Media Player Features
- Subtitle track management with multi-language support
- Subtitle sync verification and offset adjustment
- Subtitle download tracking and caching
- Audio track management for multi-language audio
- Chapter support for video bookmarks
- Cover art fetching from MusicBrainz, Last.fm, Spotify, iTunes, Discogs
- Lyrics data with synced lyrics support
- Playback session tracking with resume capability
- Playlist management including smart playlists
- Translation cache for subtitles and lyrics
- Language preference management per content type

#### Real-Time Features
- WebSocket-based real-time event bus for live updates
- Real-time media change notifications
- File system monitoring and change tracking

#### Authentication and Authorization
- JWT-based authentication with session management
- Role-based access control (Admin, Moderator, User, Viewer)
- User permission management with custom permission grants
- Authentication audit logging
- Account lockout protection with failed login tracking
- Session management with device tracking

#### Media Conversion
- Format conversion job queue with priority support
- Support for video, audio, image, and document formats
- Batch conversion with progress tracking
- Quality settings per conversion job

#### Database and Storage
- SQLite database with automatic migrations on startup
- PostgreSQL support for production deployments
- 6-version migration system with tracking table
- Virtual paths for unified file access across protocols
- Scan history tracking with detailed statistics
- External API response caching

#### Administration and Deployment
- Docker and Docker Compose support for development and production
- Comprehensive configuration via environment variables, .env files, and config.json
- Installation wizard with step-by-step guided setup
- Backup and recovery documentation
- Scaling guide for production environments
- Production runbook for operations teams

#### Documentation
- Architecture documentation for all components (Go backend, React frontend, Android, Tauri IPC)
- API documentation with endpoint reference
- WebSocket events documentation
- Database schema documentation
- User guides for web, desktop, Android, and Android TV
- Deployment guide with Docker setup
- Monitoring guide with metrics
- Contributing guide with code standards
- Configuration guide
- Troubleshooting guide
- Security testing guide
- QA testing guide and execution reports
- Test strategy documentation

#### Testing
- Comprehensive test suite across all modules (~1000+ tests across phases)
- Table-driven tests for Go backend
- Benchmark tests for performance-critical paths
- Security scanning and remediation (Phase 6)
- Cross-cutting error handling tests
- WebSocket integration tests
- Auth flow tests
- Accessibility tests for frontend

### Changed
- Migrated storage model from SMB-only (`smb_roots`) to multi-protocol (`storage_roots`)
- Updated media items schema for Android TV compatibility (migration 006)
- Refactored subtitle foreign keys from `media_items` to `files` table reference
- Reorganized root directory structure: moved ~65 files into config/, scripts/, tests/, docs/ subdirectories

### Fixed
- Critical build blockers and security issues (Phase 0)
- Memory leaks and race conditions (Phase 0)
- Subtitle foreign key references pointing to wrong table (migration 015/v6)
- Disconnected features across all modules (Phase 2)

### Security
- Security hardening and stability fixes (Phase 1)
- Security scanning and remediation across all modules (Phase 6)
- JWT token management with secure session handling
- Password hashing with salt
- Input validation with Zod on frontend
- CORS and rate limiting middleware
- Freemium security setup documentation

---

## Release History

This changelog was created as part of the Phase 8 documentation completion effort. Prior development was tracked through auto-commits and phase-based commit messages. Future releases will follow this structured format.

[Unreleased]: https://github.com/catalogizer/catalogizer/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/catalogizer/catalogizer/releases/tag/v1.0.0

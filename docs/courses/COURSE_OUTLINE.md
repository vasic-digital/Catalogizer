# Catalogizer Video Course - Complete Outline

## Course Title: Mastering Catalogizer - Multi-Protocol Media Collection Management

**Total Estimated Duration**: ~9 hours 45 minutes
**Target Audience**: End users, system administrators, and developers
**Prerequisites**: Basic familiarity with web applications and file management concepts
**Certification**: Complete all modules, exercises, and pass the final assessment with 80% or higher

---

## Module 1: Introduction & Installation

**Module Duration**: ~45 minutes
**Description**: Understand what Catalogizer is, its architecture, and how to get it running on your system.
**Prerequisites**: None

### Learning Objectives

By the end of this module, students will be able to:
- Explain the purpose and value proposition of Catalogizer
- Identify all major system components and their roles
- Install Catalogizer using either containers or manual setup
- Complete first-time configuration and verify the installation

### Lesson 1.1: What is Catalogizer?

- **Duration**: 8 minutes
- **Learning Objectives**:
  - Understand the purpose and core value proposition of Catalogizer
  - Identify the main components: catalog-api (Go), catalog-web (React), desktop apps (Tauri), mobile apps (Kotlin/Compose), and the API client library
  - Recognize supported storage protocols: SMB, FTP, NFS, WebDAV, and local filesystem
- **Prerequisites**: None

### Lesson 1.2: System Requirements & Architecture Overview

- **Duration**: 10 minutes
- **Learning Objectives**:
  - Identify hardware and software prerequisites (Go 1.21+, Node.js 18+, Rust, Android SDK)
  - Understand the layered architecture: Handler -> Service -> Repository -> SQLite
  - Recognize the role of external providers (TMDB, IMDB, TVDB, MusicBrainz, Spotify, Steam)
  - Understand the WebSocket-based real-time update system
  - Understand the submodule architecture for reusable components
- **Prerequisites**: None

### Lesson 1.3: Container Installation (Podman/Docker)

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Set up Catalogizer using Podman Compose or Docker Compose with the production configuration
  - Configure PostgreSQL, Redis, and Nginx services via docker-compose.yml
  - Understand volume mounts, health checks, and resource limits
  - Start the development environment with docker-compose.dev.yml
- **Prerequisites**: Podman or Docker installed

### Lesson 1.4: Manual Installation (Backend & Frontend)

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Clone the repository and initialize submodules
  - Install Go dependencies and build the catalog-api backend
  - Install Node.js dependencies and start the catalog-web frontend
  - Initialize the SQLite database and verify the installation
- **Prerequisites**: Go 1.21+, Node.js 18+, Git

---

## Module 2: Getting Started with Media Management

**Module Duration**: ~75 minutes
**Description**: Navigate the web interface, connect storage sources, browse your catalog, use search, and understand analytics.
**Prerequisites**: Module 1 completed; Catalogizer running

### Learning Objectives

By the end of this module, students will be able to:
- Navigate the full web interface confidently
- Connect and manage multiple storage protocol sources
- Browse, filter, sort, and search media effectively
- Interpret analytics data and growth trends
- Use the AI Dashboard for intelligent insights

### Lesson 2.1: Web UI Overview & Dashboard

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Navigate the main layout: Dashboard, Media, Collections, Search, Profile menu
  - Read the Quick Stats panel: total media count, collections, favorites, storage used
  - Review recent activity, collection updates, and system notifications
  - Use quick actions: Upload Media, Create Collection, Import from Cloud, View Analytics
- **Prerequisites**: Module 1 completed; Catalogizer running

### Lesson 2.2: Connecting Storage Sources

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Add an SMB/CIFS share with credentials and domain settings
  - Connect FTP and WebDAV sources through the unified client interface
  - Configure NFS mounts using the NFS client
  - Add local filesystem paths for direct access
  - Understand the UnifiedClient interface that abstracts all protocols (filesystem/interface.go)
  - Use SMB discovery to auto-detect available network shares
- **Prerequisites**: Lesson 2.1 completed; access to at least one network share

### Lesson 2.3: Browsing & Navigating the Catalog

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Switch between Grid view and List view in the Media Browser
  - Apply filters by type (Images, Videos, Documents, Audio), date range, and file size
  - Sort items by name, date, size, type, or relevance
  - Open the Media Detail Modal to view metadata, quality info, and external provider data
  - Understand real-time updates delivered via WebSocket (WebSocketContext)
- **Prerequisites**: Lesson 2.2 completed; media sources connected

### Lesson 2.4: Search & Discovery

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Use the advanced search with filters, tags, and metadata queries
  - Filter results by media type and category (movies, TV shows, music, games, software, documentaries)
  - Leverage external metadata from TMDB, IMDB, TVDB, MusicBrainz, and Spotify for enriched search results
  - Use the recommendation service for content discovery
  - Identify duplicates across storage sources
- **Prerequisites**: Lesson 2.3 completed

### Lesson 2.5: Analytics Dashboard

- **Duration**: 9 minutes
- **Learning Objectives**:
  - Access the Analytics page for comprehensive library statistics
  - Interpret growth trends and quality analysis data
  - Understand the AI Dashboard features for intelligent insights
  - Export analytics data for external reporting
- **Prerequisites**: Lesson 2.3 completed

### Lesson 2.6: Localization & Multi-Language Support

- **Duration**: 10 minutes
- **Learning Objectives**:
  - Configure language preferences for the interface
  - Understand the localization service (localization_service.go) and translation service (translation_service.go)
  - Search and browse media in multiple languages
  - Set up multi-language metadata enrichment
- **Prerequisites**: Lesson 2.1 completed

---

## Module 3: Advanced Media Features

**Module Duration**: ~75 minutes
**Description**: Organize your library with favorites, collections, playlists, subtitles, format conversion, and the built-in media player.
**Prerequisites**: Module 2 completed

### Learning Objectives

By the end of this module, students will be able to:
- Organize media using favorites, collections, and playlists
- Manage subtitles across their video library
- Convert media between formats using built-in tools
- Use the media player with all its features including lyrics, position tracking, and deep linking

### Lesson 3.1: Favorites & Bookmarks

- **Duration**: 10 minutes
- **Learning Objectives**:
  - Add and remove items from Favorites using the Favorites page and useFavorites hook
  - Export favorites to JSON and CSV formats with full metadata
  - Import favorites from exported files
  - Organize favorites for quick access to frequently used media
- **Prerequisites**: Module 2 completed

### Lesson 3.2: Collections & Organization

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Create Manual, Smart, and Dynamic collections via the Collections page
  - Drag and drop media items into collections; use bulk selection for batch operations
  - Configure Smart collections with automatic filter-based population rules
  - Set access permissions (Public, Private, Friends Only) on collections
- **Prerequisites**: Lesson 3.1 completed

### Lesson 3.3: Playlists

- **Duration**: 10 minutes
- **Learning Objectives**:
  - Create and manage playlists through the Playlists page
  - Reorder playlist items using drag-and-drop (usePlaylistReorder hook)
  - Understand the playlist service backend (playlist_service.go)
  - Play playlists sequentially through the built-in media player
- **Prerequisites**: Lesson 3.2 completed

### Lesson 3.4: Subtitle Management

- **Duration**: 10 minutes
- **Learning Objectives**:
  - Access the Subtitle Manager page for centralized subtitle operations
  - Upload and associate subtitle files with video media
  - Use the subtitle service backend (subtitle_service.go) for automatic subtitle matching
  - Switch between multiple subtitle tracks during playback
- **Prerequisites**: Lesson 3.3 completed

### Lesson 3.5: Format Conversion & PDF Tools

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Use the Conversion Tools page for media format conversion
  - Convert PDF documents to images, text, or HTML using the PDF conversion service
  - Monitor conversion progress and manage conversion queue
  - Understand supported format conversions for images, video, audio, and documents
- **Prerequisites**: Lesson 3.1 completed

### Lesson 3.6: Built-in Media Player

- **Duration**: 16 minutes
- **Learning Objectives**:
  - Use the MediaPlayer component for video and audio playback
  - Control playback with the usePlayerState hook (play, pause, seek, volume)
  - Resume playback from saved positions (playback_position_service.go)
  - Use the video player service and music player service for specialized playback
  - Access lyrics display during music playback (lyrics_service.go)
  - Stream media from remote protocol sources via the media player handlers
  - Share specific moments using deep linking (deep_linking_service.go)
- **Prerequisites**: Lessons 3.3 and 3.4 completed

---

## Module 4: Multi-Platform Experience

**Module Duration**: ~55 minutes
**Description**: Use Catalogizer on Android, Android TV, Desktop, and through the API client library.
**Prerequisites**: Module 2 completed

### Learning Objectives

By the end of this module, students will be able to:
- Build and install the Android mobile and TV applications
- Set up and use the Tauri desktop application
- Understand the Installer Wizard for guided setup
- Use the API client library for custom integrations

### Lesson 4.1: Android Mobile App

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Install and configure the catalogizer-android app
  - Understand the MVVM architecture: Compose UI, ViewModel with StateFlow, Repository, Room + Retrofit
  - Browse your catalog, manage favorites, and search media on mobile
  - Use offline mode with Room database local caching
  - Understand Hilt dependency injection used throughout the app
- **Prerequisites**: Module 2 completed; Android 8.0+ device

### Lesson 4.2: Android TV App

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Install and set up the catalogizer-androidtv app on your TV device
  - Navigate the leanback UI optimized for large screens and remote control
  - Browse collections, search, and play media directly on TV
  - Understand shared architecture with the mobile app (Kotlin/Compose, MVVM)
- **Prerequisites**: Module 2 completed; Android TV device

### Lesson 4.3: Desktop Application

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Build and install the catalogizer-desktop Tauri application
  - Understand the Tauri architecture: React frontend communicating with Rust backend via IPC commands and events
  - Use the Installer Wizard for guided first-time desktop setup
  - Access native OS features: file system integration, system tray, notifications
  - Use dev mode (npm run tauri:dev) and production builds (npm run tauri:build)
- **Prerequisites**: Module 2 completed; Rust toolchain and Node.js installed

### Lesson 4.4: API Client Library

- **Duration**: 13 minutes
- **Learning Objectives**:
  - Install and configure the catalogizer-api-client TypeScript library
  - Authenticate and make API calls using the client services
  - Understand the client type system and utility functions
  - Build custom integrations and automations using the API client
  - Run the client test suite (npm run build && npm run test)
- **Prerequisites**: Module 2 completed; TypeScript/Node.js knowledge

---

## Module 5: Administration & Configuration

**Module Duration**: ~65 minutes
**Description**: Manage users, configure roles, monitor system health, perform backups, and troubleshoot issues.
**Prerequisites**: Module 2 completed; admin account access

### Learning Objectives

By the end of this module, students will be able to:
- Create and manage user accounts with role-based access control
- Configure security layers including JWT, database encryption, and 2FA
- Set up and interpret monitoring dashboards with Prometheus and Grafana
- Implement a reliable backup and restore strategy
- Diagnose and resolve common operational issues

### Lesson 5.1: User Management & Roles

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Access the Admin Panel (AdminPanel.tsx) for user administration
  - Create, edit, and deactivate user accounts
  - Understand JWT-based authentication (internal/auth/service.go, middleware.go)
  - Configure role-based access control and permissions
  - Manage active sessions and force logout when needed
- **Prerequisites**: Module 2 completed; admin account access

### Lesson 5.2: Security Configuration

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Configure JWT secrets and token expiry (JWT_SECRET, JWT_EXPIRY_HOURS, REFRESH_TOKEN_EXPIRY_HOURS)
  - Understand database encryption (DB_ENCRYPTION_KEY)
  - Set up the security testing pipeline (scripts/security-test.sh, docker-compose.security.yml)
  - Run Snyk and SonarQube scans (scripts/snyk-scan.sh, scripts/sonarqube-scan.sh)
  - Configure two-factor authentication for users
- **Prerequisites**: Lesson 5.1 completed

### Lesson 5.3: Monitoring & Metrics

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Understand the metrics system (internal/metrics/metrics.go, middleware.go)
  - Configure Prometheus scraping using monitoring/prometheus.yml
  - Set up Grafana dashboards from monitoring/grafana/ and config/grafana-dashboards/
  - Monitor SMB connection health with circuit breaker status and offline cache metrics
  - Track media detection pipeline performance and analysis throughput
- **Prerequisites**: Lesson 5.1 completed

### Lesson 5.4: Backup, Restore & Cloud Sync

- **Duration**: 14 minutes
- **Learning Objectives**:
  - Back up the database and configuration files
  - Synchronize files with Amazon S3, Google Cloud Storage, or local backup folders
  - Generate professional PDF reports with charts and analytics (advanced reporting)
  - Configure automatic archiving rules for storage management
  - Restore from backups after data loss
- **Prerequisites**: Lesson 5.1 completed

### Lesson 5.5: Troubleshooting & Resilience

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Diagnose SMB disconnection issues using circuit breaker logs and exponential backoff retry metrics
  - Understand offline caching behavior when network storage is unavailable
  - Configure SMB resilience parameters: retry attempts, delay, health check interval, connection timeout
  - Use recovery mechanisms (internal/recovery/) for crash recovery
  - Review log files and adjust LOG_LEVEL for debugging
- **Prerequisites**: Module 5 lessons completed

---

## Module 6: Developer Guide & API

**Module Duration**: ~70 minutes
**Description**: Understand the codebase architecture, set up a development environment, add features, use the submodule system, and contribute.
**Prerequisites**: Familiarity with Go and TypeScript

### Learning Objectives

By the end of this module, students will be able to:
- Trace request flow through all backend layers
- Set up a complete development environment for any component
- Add new features following established patterns
- Work with the submodule architecture for reusable components
- Build and package all platform targets

### Lesson 6.1: Architecture Deep Dive

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Trace the request flow: Handler (handlers/) -> Service (services/) -> Repository (repository/) -> SQLite
  - Understand the media detection pipeline: detector/ -> analyzer/ -> providers/ (TMDB, IMDB, etc.)
  - Map the real-time event system: event bus -> WebSocket -> clients (internal/media/realtime/)
  - Understand the filesystem abstraction: UnifiedClient interface -> protocol-specific clients (SMB, FTP, NFS, WebDAV, local)
  - Review the frontend architecture: AuthProvider -> WebSocketProvider -> Router with ProtectedRoute
- **Prerequisites**: Familiarity with Go and TypeScript

### Lesson 6.2: Setting Up the Development Environment

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Clone the monorepo, initialize submodules, and understand the directory structure
  - Start the backend in dev mode (go run main.go) and frontend (npm run dev on port 5173)
  - Launch the dev container environment (podman-compose -f docker-compose.dev.yml up)
  - Set up the desktop app dev environment (npm run tauri:dev) with Rust toolchain
  - Configure Android development with Gradle (./gradlew assembleDebug)
- **Prerequisites**: Go 1.21+, Node.js 18+, Rust, Android SDK

### Lesson 6.3: Adding New Features

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Add a new storage protocol by implementing the UnifiedClient interface (filesystem/interface.go)
  - Create a new API endpoint: handler in handlers/, service in services/, route in main.go
  - Add a new external metadata provider in internal/media/providers/
  - Build a new frontend page with React Query for server state and protected routing
  - Add Tauri IPC commands for desktop-specific features
- **Prerequisites**: Lesson 6.1 completed

### Lesson 6.4: Submodule Architecture & Reusable Components

- **Duration**: 14 minutes
- **Learning Objectives**:
  - Understand the submodule architecture: Go, TypeScript, and Kotlin modules
  - Use the setup-submodule.sh script to create new submodules
  - Work with upstream remotes for multi-remote push (GitHub + GitLab)
  - Integrate submodules into the main application (Auth, Cache, Database, Concurrency, etc.)
  - Manage submodule dependencies and versioning
- **Prerequisites**: Lesson 6.1 completed

### Lesson 6.5: Build Pipeline & Packaging

- **Duration**: 14 minutes
- **Learning Objectives**:
  - Use the build scripts in scripts/ for automated builds
  - Build the containerized pipeline with scripts/container-build.sh
  - Create release artifacts for all platforms with scripts/build-all-releases.sh
  - Deploy to production using docker-compose.yml with proper environment configuration
  - Configure Nginx reverse proxy (config/nginx.conf) and Redis cache (config/redis.conf)
  - Set up systemd services using config/systemd/catalogizer-api.service
- **Prerequisites**: Lesson 6.4 completed

---

## Module 7: Testing & Quality Assurance

**Module Duration**: ~60 minutes
**Description**: Comprehensive testing strategies, writing effective tests, security scanning, performance testing, and maintaining code quality across all components.
**Prerequisites**: Module 6 completed (at least Lessons 6.1-6.2)

### Learning Objectives

By the end of this module, students will be able to:
- Write effective tests for all components using established patterns
- Run the complete test suite and interpret results
- Perform security scanning with Snyk and SonarQube
- Execute performance and stress tests
- Validate test coverage and maintain quality standards

### Lesson 7.1: Go Backend Testing

- **Duration**: 18 minutes
- **Learning Objectives**:
  - Write table-driven tests following Go conventions with *_test.go files beside source
  - Test handlers (auth_handler_test.go, media_handler_test.go, subtitle_handler_test.go, etc.)
  - Test services (catalog_test.go, favorites_service_test.go, playlist_service_test.go, etc.)
  - Test repositories (user_repository_test.go, favorites_repository_test.go, sync_repository_test.go, etc.)
  - Use the test helper (internal/tests/test_helper.go) for database setup
  - Run targeted tests with `go test -v -run TestName ./path/`
  - Measure and interpret test coverage
- **Prerequisites**: Module 6 Lessons 6.1-6.2 completed

### Lesson 7.2: Frontend & Client Testing

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Write React component tests in __tests__/ directories using Vitest
  - Test UI components (Badge, Progress, Select, Switch, Tabs, Textarea)
  - Test page-level components (Collections, Favorites, Playlists, Admin, AIDashboard)
  - Test library functions (api.test.ts, collectionsApi.test.ts, favoritesApi.test.ts)
  - Run linting (npm run lint) and type checking (npm run type-check)
  - Test the API client library (catalogizer-api-client)
  - Test the installer wizard components
- **Prerequisites**: Module 6 Lessons 6.1-6.2 completed

### Lesson 7.3: Security Scanning & Vulnerability Management

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Run the security test suite (scripts/security-test.sh)
  - Use the security Docker Compose environment (docker-compose.security.yml)
  - Perform dependency scanning with Snyk (scripts/snyk-scan.sh)
  - Execute static analysis with SonarQube (scripts/sonarqube-scan.sh)
  - Manage false positives with dependency-check-suppressions.xml
  - Run the comprehensive security scan (scripts/security-scan.sh)
  - Interpret and act on security scan results
- **Prerequisites**: Lesson 7.1 completed

### Lesson 7.4: Performance Testing & Stress Testing

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Run performance tests with scripts/performance-test.sh
  - Execute memory leak detection with scripts/memory-leak-check.sh
  - Use the stress test service (stress_test_service.go) and handler (stress_test_handler.go)
  - Interpret benchmark results from providers_bench_test.go and auth_service_bench_test.go
  - Validate test coverage with scripts/validate-coverage.sh
  - Run the complete test suite with scripts/run-all-tests.sh
- **Prerequisites**: Lesson 7.2 completed

---

## Module 8: Deployment & Production

**Module Duration**: ~55 minutes
**Description**: Deploy Catalogizer to production environments, configure infrastructure, set up monitoring, and maintain the system.
**Prerequisites**: Modules 5 and 6 completed

### Learning Objectives

By the end of this module, students will be able to:
- Deploy Catalogizer using containers with production-grade configuration
- Configure Nginx reverse proxy and Redis caching
- Set up systemd services for bare-metal deployment
- Implement production monitoring with Prometheus and Grafana
- Establish operational procedures for maintenance and upgrades

### Lesson 8.1: Production Container Deployment

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Configure docker-compose.yml for production with proper resource limits and health checks
  - Set required environment variables: POSTGRES_PASSWORD, JWT_SECRET, DB_ENCRYPTION_KEY
  - Understand container networking, volume persistence, and restart policies
  - Use Podman as an alternative to Docker for rootless container deployment
  - Deploy with `podman-compose up -d` and verify service health
  - Use the build container (docker-compose.build.yml) for reproducible builds
- **Prerequisites**: Module 5 and Module 6 completed

### Lesson 8.2: Infrastructure Configuration

- **Duration**: 14 minutes
- **Learning Objectives**:
  - Configure Nginx as a reverse proxy using config/nginx.conf and config/nginx/catalogizer.prod.conf
  - Set up Redis caching with config/redis.conf
  - Configure PostgreSQL for production workloads
  - Set up systemd services using config/systemd/catalogizer-api.service for bare-metal deployment
  - Configure TLS/SSL termination at the Nginx layer
  - Tune resource limits for database, cache, and application services
- **Prerequisites**: Lesson 8.1 completed

### Lesson 8.3: Production Monitoring & Alerting

- **Duration**: 14 minutes
- **Learning Objectives**:
  - Deploy Prometheus with monitoring/prometheus.yml for metric collection
  - Set up Grafana dashboards from monitoring/grafana/ and config/grafana-dashboards/
  - Configure alerts for service health, disk space, error rates, and latency
  - Monitor SMB connection resilience: circuit breaker states, offline cache, retry metrics
  - Track media detection pipeline throughput in production
  - Set up log aggregation and rotation policies
- **Prerequisites**: Lesson 8.1 completed

### Lesson 8.4: Maintenance, Upgrades & Disaster Recovery

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Plan and execute zero-downtime upgrades with container rolling updates
  - Implement the three-tier backup strategy: daily database, weekly config, monthly verification
  - Perform database migrations using the migrations system (database/migrations/)
  - Handle disaster recovery scenarios: database corruption, configuration loss, media source failure
  - Establish a runbook for common operational procedures
  - Plan capacity based on analytics data and growth trends
- **Prerequisites**: Module 8 lessons completed

---

## Summary

| Module | Lessons | Duration |
|--------|---------|----------|
| Module 1: Introduction & Installation | 4 | ~45 min |
| Module 2: Getting Started with Media Management | 6 | ~75 min |
| Module 3: Advanced Media Features | 6 | ~75 min |
| Module 4: Multi-Platform Experience | 4 | ~55 min |
| Module 5: Administration & Configuration | 5 | ~65 min |
| Module 6: Developer Guide & API | 5 | ~70 min |
| Module 7: Testing & Quality Assurance | 4 | ~60 min |
| Module 8: Deployment & Production | 4 | ~55 min |
| **Total** | **38** | **~9h 45min** |

---

## Certification Path

1. **Complete all 8 modules** including video lessons and reading materials
2. **Complete all hands-on exercises** (see EXERCISES.md) with at least 80% completion
3. **Pass the final assessment** (see ASSESSMENT.md) with a score of 80% or higher
4. **Complete one capstone project** from the exercises document

Certification levels:
- **Catalogizer User** -- Modules 1-4 completed with exercises and assessment
- **Catalogizer Administrator** -- Modules 1-5 completed with exercises and assessment
- **Catalogizer Developer** -- All 8 modules completed with exercises, assessment, and capstone project

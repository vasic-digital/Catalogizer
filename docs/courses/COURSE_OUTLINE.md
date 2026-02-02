# Catalogizer Video Course - Complete Outline

## Course Title: Mastering Catalogizer - Multi-Protocol Media Collection Management

**Total Estimated Duration**: ~6 hours 30 minutes
**Target Audience**: End users, system administrators, and developers
**Prerequisites**: Basic familiarity with web applications and file management concepts

---

## Module 1: Introduction & Installation

**Module Duration**: ~55 minutes
**Description**: Understand what Catalogizer is, its architecture, and how to get it running on your system.

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
  - Identify hardware and software prerequisites (Go 1.21+, Node.js 18+, SQLCipher)
  - Understand the layered architecture: Handler, Service, Repository, SQLite
  - Recognize the role of external providers (TMDB, IMDB, TVDB, MusicBrainz, Spotify, Steam)
  - Understand the WebSocket-based real-time update system
- **Prerequisites**: None

### Lesson 1.3: Docker Installation

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Set up Catalogizer using Docker Compose with the production configuration
  - Configure PostgreSQL, Redis, and Nginx services via docker-compose.yml
  - Understand volume mounts, health checks, and resource limits
  - Start the development environment with docker-compose.dev.yml
- **Prerequisites**: Docker and Docker Compose installed

### Lesson 1.4: Manual Installation (Backend & Frontend)

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Clone the repository and install Go dependencies for catalog-api
  - Build and run the backend API server on port 8080
  - Install Node.js dependencies and start the catalog-web frontend on port 5173
  - Initialize the SQLCipher-encrypted database
- **Prerequisites**: Go 1.21+, Node.js 18+, SQLCipher, Git

### Lesson 1.5: First-Time Configuration

- **Duration**: 10 minutes
- **Learning Objectives**:
  - Configure backend environment variables (.env): database path, JWT secrets, SMB sources, external API keys
  - Configure frontend environment variables (.env.local): API base URL, WebSocket URL, feature flags
  - Set up SMB resilience parameters (retry attempts, health check interval, offline cache size)
  - Verify the installation by accessing the web UI and API documentation at /swagger/index.html
- **Prerequisites**: Lessons 1.3 or 1.4 completed

---

## Module 2: Getting Started

**Module Duration**: ~60 minutes
**Description**: Navigate the web interface, connect storage sources, browse your catalog, and use search.

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
  - Mount NFS shares (macOS) using the NFS client
  - Add local filesystem paths for direct access
  - Understand the UnifiedClient interface that abstracts all protocols (filesystem/interface.go)
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
  - Save and reuse search queries for recurring workflows
- **Prerequisites**: Lesson 2.3 completed

### Lesson 2.5: Analytics Dashboard

- **Duration**: 9 minutes
- **Learning Objectives**:
  - Access the Analytics page for comprehensive library statistics
  - Interpret growth trends and quality analysis data
  - Understand the AI Dashboard features for intelligent insights
  - Export analytics data for external reporting
- **Prerequisites**: Lesson 2.3 completed

---

## Module 3: Media Management

**Module Duration**: ~70 minutes
**Description**: Organize your library with favorites, collections, playlists, subtitles, and the built-in media player.

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
  - Use the useCollections hook to understand how collection state is managed
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
- **Prerequisites**: Lessons 3.3 and 3.4 completed

---

## Module 4: Multi-Platform Usage

**Module Duration**: ~55 minutes
**Description**: Use Catalogizer on Android, Android TV, Desktop, and through the API client library.

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

## Module 5: Administration

**Module Duration**: ~65 minutes
**Description**: Manage users, configure roles, monitor system health, perform backups, and troubleshoot issues.

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
  - Understand SQLCipher database encryption (DB_ENCRYPTION_KEY)
  - Set up the security testing pipeline (scripts/security-test.sh, docker-compose.security.yml)
  - Run Snyk and SonarQube scans (scripts/snyk-scan.sh, scripts/sonarqube-scan.sh)
  - Configure two-factor authentication for users
- **Prerequisites**: Lesson 5.1 completed

### Lesson 5.3: Monitoring & Metrics

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Understand the metrics system (internal/metrics/metrics.go, middleware.go)
  - Configure Prometheus scraping using monitoring/prometheus.yml
  - Set up Grafana dashboards from monitoring/grafana/ for visual monitoring
  - Monitor SMB connection health with circuit breaker status and offline cache metrics
  - Track media detection pipeline performance and analysis throughput
- **Prerequisites**: Lesson 5.1 completed

### Lesson 5.4: Backup, Restore & Cloud Sync

- **Duration**: 14 minutes
- **Learning Objectives**:
  - Back up the SQLCipher database and configuration files
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

## Module 6: Developer Guide

**Module Duration**: ~65 minutes
**Description**: Understand the codebase architecture, set up a development environment, add features, write tests, and contribute.

### Lesson 6.1: Architecture Deep Dive

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Trace the request flow: Handler (internal/handlers/) -> Service (internal/services/) -> Repository -> SQLite
  - Understand the media detection pipeline: detector/ -> analyzer/ -> providers/ (TMDB, IMDB, etc.)
  - Map the real-time event system: event bus -> WebSocket -> clients (internal/media/realtime/)
  - Understand the filesystem abstraction: UnifiedClient interface -> protocol-specific clients (SMB, FTP, NFS, WebDAV, local)
  - Review the frontend architecture: AuthProvider -> WebSocketProvider -> Router with ProtectedRoute
- **Prerequisites**: Familiarity with Go and TypeScript

### Lesson 6.2: Setting Up the Development Environment

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Clone the monorepo and understand the directory structure across all components
  - Start the backend in dev mode (go run main.go) and frontend (npm run dev on port 5173)
  - Launch the dev Docker environment (docker-compose -f docker-compose.dev.yml up)
  - Set up the desktop app dev environment (npm run tauri:dev) with Rust toolchain
  - Configure Android development with Gradle (./gradlew assembleDebug)
- **Prerequisites**: Go 1.21+, Node.js 18+, Rust, Android SDK

### Lesson 6.3: Adding New Features

- **Duration**: 15 minutes
- **Learning Objectives**:
  - Add a new storage protocol by implementing the UnifiedClient interface (filesystem/interface.go)
  - Create a new API endpoint: handler in internal/handlers/, service in internal/services/, route in main.go
  - Add a new external metadata provider in internal/media/providers/
  - Build a new frontend page with React Query for server state and protected routing
  - Add Tauri IPC commands for desktop-specific features
- **Prerequisites**: Lesson 6.1 completed

### Lesson 6.4: Testing Strategy

- **Duration**: 12 minutes
- **Learning Objectives**:
  - Write Go table-driven tests with *_test.go files beside source
  - Run backend tests (go test ./...) and frontend tests (npm run test)
  - Execute the full test suite including security tests (scripts/run-all-tests.sh)
  - Run frontend linting and type checking (npm run lint && npm run type-check)
  - Execute Android unit tests (./gradlew test) and API client tests (npm run build && npm run test)
  - Understand integration tests in the tests/ directory
- **Prerequisites**: Lesson 6.2 completed

### Lesson 6.5: CI/CD, Security Scanning & Deployment

- **Duration**: 11 minutes
- **Learning Objectives**:
  - Understand the build scripts in build-scripts/ for automated builds
  - Run security scans: Snyk (scripts/snyk-scan.sh), SonarQube (scripts/sonarqube-scan.sh)
  - Use the security Docker Compose environment (docker-compose.security.yml)
  - Deploy to production using docker-compose.yml with proper environment configuration
  - Configure Nginx reverse proxy and Redis cache from config/ directory
- **Prerequisites**: Lesson 6.4 completed

---

## Summary

| Module | Lessons | Duration |
|--------|---------|----------|
| Module 1: Introduction & Installation | 5 | ~55 min |
| Module 2: Getting Started | 5 | ~60 min |
| Module 3: Media Management | 6 | ~70 min |
| Module 4: Multi-Platform Usage | 4 | ~55 min |
| Module 5: Administration | 5 | ~65 min |
| Module 6: Developer Guide | 5 | ~65 min |
| **Total** | **30** | **~6h 10min** |

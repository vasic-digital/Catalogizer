# Catalogizer Features

Catalogizer is a multi-protocol media collection management system that detects, categorizes, and organizes media across all your storage. Below is a comprehensive overview of its capabilities.

---

## Multi-Protocol Storage

Connect to media stored anywhere on your network or cloud.

- **SMB/CIFS**: Windows and Samba file shares with automatic reconnection, circuit breaker pattern, and offline caching
- **FTP/FTPS**: Standard and secure File Transfer Protocol access
- **NFS**: Network File System with automatic mounting support
- **WebDAV**: HTTP-based file access for web storage services
- **Local Filesystem**: Direct access to locally attached storage
- **Cloud Storage Sync**: Synchronize files with Amazon S3, Google Cloud Storage, or local folders

All protocols share a common UnifiedClient interface, making it easy to manage media across different storage backends from a single catalog. A factory pattern creates the appropriate client based on the protocol, so the rest of the application is protocol-agnostic.

### Network Resilience

Catalogizer is designed for unreliable network environments:

- **Circuit Breaker**: Prevents repeated connection attempts to downed servers, preserving system resources
- **Exponential Backoff Retry**: Gradually retries failed connections with increasing delays
- **Offline Cache**: Serves previously loaded metadata during storage outages so users can continue browsing
- **SMB Discovery**: Auto-detects available SMB shares on the local network for simplified setup

## Media Detection and Analysis

Automatically identify and categorize your media collection.

- **50+ media types detected**: Movies, TV shows, music, games, software, documentaries, and more
- **Quality analysis**: Automatic resolution, codec, and bitrate detection with version tracking
- **External metadata integration**: Enriches your catalog with data from TMDB, IMDB, TVDB, MusicBrainz, Spotify, and Steam
- **Real-time monitoring**: Continuously watches storage sources for new, changed, or removed files
- **Media detection pipeline**: Detector identifies file types, analyzer extracts quality metadata, providers fetch external information

## Subtitle Management

Comprehensive subtitle support for your video collection.

- **Multi-provider search**: Search subtitles across OpenSubtitles, SubDB, Yify Subtitles, Subscene, and Addic7ed
- **Hash-based matching**: Match subtitles precisely using file hash and size
- **Subtitle translation**: Translate subtitles between languages with configurable translation providers
- **Synchronization verification**: Check and adjust subtitle timing against video files
- **Custom upload**: Upload your own subtitle files in SRT, ASS, SSA, VTT, and SUB formats

## Security

Enterprise-grade security for your media catalog.

- **JWT authentication**: Token-based auth with configurable expiry and refresh tokens
- **Role-based access control**: Define user roles and permissions
- **SQLCipher encrypted database**: Media metadata is stored in an encrypted SQLite database
- **CORS configuration**: Configurable cross-origin resource sharing for web deployments
- **Security testing**: Built-in security testing suite via Docker Compose security profile

## Monitoring and Analytics

Track your catalog's health and growth.

- **Prometheus metrics**: The API exposes a `/metrics` endpoint with HTTP request rates, latencies, and custom application metrics
- **Grafana dashboards**: Pre-configured dashboard for API performance, resource utilization, and Go runtime statistics
- **Collection analytics**: Total files, storage usage, quality distribution, growth trends, and source reliability
- **Real-time status**: WebSocket-based live updates for connection health, scan progress, and new media notifications
- **Alerting**: Configure alerts in Grafana for API latency, error rates, and availability

## Multi-Platform Clients

Access your catalog from any device.

### Web Application (catalog-web)
- Modern React TypeScript interface with Tailwind CSS
- Real-time updates via WebSocket integration
- Advanced search with full-text search, filters, and multiple view modes (grid, list, detail)
- Analytics dashboard with collection statistics and growth charts
- Responsive design for desktop and mobile browsers

### Desktop Application (catalogizer-desktop)
- Cross-platform native app built with Tauri (Rust + React)
- Builds for Windows, macOS, and Linux
- System tray integration and native performance

### Android App (catalogizer-android)
- MVVM architecture with Jetpack Compose UI
- Offline mode with Room database and automatic sync
- Configurable caching with Wi-Fi-only and storage limit options
- Material Design 3 components

### Android TV App (catalogizer-androidtv)
- Leanback UI optimized for TV screens
- D-pad and remote control navigation
- Google Assistant voice search
- Android TV recommendations integration

### Installation Wizard (installer-wizard)
- Desktop setup tool built with Tauri
- Automatic network discovery for SMB devices
- Visual configuration with real-time connection testing
- Exports configuration files for the main system

### TypeScript API Client (catalogizer-api-client)
- Typed client library for integrating Catalogizer into other applications
- Media search, metadata retrieval, and source management
- Publishable as an npm package or usable via local linking

## Built-in Media Player

Play video and audio directly in the browser without external software.

- **Universal playback**: Stream from any connected storage source regardless of protocol
- **Playback position tracking**: Resume where you left off, synced across all devices
- **Subtitle support**: SRT, ASS, SSA, and VTT formats with auto-matching and in-player track selection
- **Music features**: Lyrics display, cover art fetching from MusicBrainz, Last.fm, Spotify, iTunes, and Discogs
- **Deep linking**: Share links to specific playback positions with other users
- **Playlist playback**: Auto-advancement through ordered sequences with seamless transitions

## Organization and Library Management

Keep your library structured with powerful organizational tools.

- **Favorites**: Quick bookmarking with JSON and CSV export/import; matching uses metadata so imports work across instances
- **Collections**: Manual (hand-picked), Smart (rule-based auto-population), and Dynamic (adaptive criteria) collection types
- **Playlists**: Ordered sequences for sequential playback with drag-and-drop reordering
- **Access permissions**: Collections support Public, Private, and shared-with-specific-users visibility
- **Bulk operations**: Select multiple items for batch actions across the catalog

## Format Conversion

Transform media between formats without leaving the application.

- **Video**: Convert between containers and codecs
- **Audio**: MP3, FLAC, WAV, AAC, and more
- **PDF**: Convert to images (thumbnails), text (search indexing), or HTML (web display)
- **Batch queue**: Queue multiple conversions with real-time progress via WebSocket
- **Automatic cataloging**: Converted files appear in the catalog alongside originals

## Localization

Full multi-language support for both the interface and media metadata.

- **Interface localization**: Translates UI labels, messages, and system text
- **Media metadata translation**: Displays titles and descriptions in the user's preferred language
- **TMDB multi-language**: Fetches metadata in dozens of languages
- **Cross-language search**: Find media using translated titles
- **Extensible**: Add new languages through translation files without code changes

## Advanced Reporting

Generate professional reports and analytics exports.

- **PDF reports**: Charts, statistics, and library summaries in professional format
- **Analytics export**: Export data for external reporting tools
- **Point-in-time snapshots**: Reports capture the state of your library at generation time
- **Growth analysis**: Track how your library has changed over time

## Modular Architecture

Built for extensibility with a submodule-based architecture.

- **19 Go modules**: Auth, Cache, Database, Concurrency, Storage, EventBus, Streaming, Security, Observability, Formatters, Plugins, Challenges, Filesystem, RateLimiter, Config, Discovery, Media, Middleware, Watcher
- **2 TypeScript modules**: WebSocket-Client, UI-Components-React
- **1 Kotlin module**: Android-Toolkit
- Each module is an independent git repository with its own tests and documentation
- Shared across projects for consistent behavior and reduced duplication

## Additional Features

- **Duplicate detection**: Identify the same content across different storage sources using hash-based matching
- **Recommendation engine**: Suggests media based on user interaction patterns
- **AI Dashboard**: Intelligent insights derived from library patterns and usage data
- **WebSocket Event Bus**: Real-time event system connecting backend changes to all connected clients
- **Connection Pooling**: Managed connection pools for storage protocols
- **Crash Recovery**: Automatic state restoration from persistent storage after unexpected termination

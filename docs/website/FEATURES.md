# Catalogizer Features

## Multi-Protocol Storage

Connect to media stored anywhere on your network or cloud.

- **SMB/CIFS**: Windows and Samba file shares with automatic reconnection, circuit breaker pattern, and offline caching
- **FTP/FTPS**: Standard and secure File Transfer Protocol access
- **NFS**: Network File System with automatic mounting support
- **WebDAV**: HTTP-based file access for web storage services
- **Local Filesystem**: Direct access to locally attached storage
- **Cloud Storage Sync**: Synchronize files with Amazon S3, Google Cloud Storage, or local folders

All protocols share a common interface, making it easy to manage media across different storage backends from a single catalog.

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

## Additional Features

- **PDF Conversion Service**: Convert PDF documents to images, text, or HTML formats
- **Favorites Export/Import**: Export and import favorites in JSON and CSV formats with metadata
- **Advanced Reporting**: Generate professional PDF reports with charts and analytics
- **Resilient Architecture**: Circuit breaker, exponential backoff retry, and offline caching for network storage protocols
- **WebSocket Event Bus**: Real-time event system connecting backend changes to all connected clients
- **Connection Pooling**: Managed connection pools for storage protocols

---
title: Frequently Asked Questions
description: Answers to common questions about Catalogizer - installation, storage protocols, media detection, security, and more
---

# Frequently Asked Questions

---

## General

### What is Catalogizer?

Catalogizer is a multi-protocol media collection management system. It connects to your media stored across SMB/CIFS, FTP/FTPS, NFS, WebDAV, and local filesystems, automatically detects and categorizes over 50 media types, and enriches items with metadata from external providers like TMDB, IMDB, TVDB, MusicBrainz, Spotify, and Steam. You access your unified catalog from a web browser, desktop app, Android phone, or Android TV.

### Is Catalogizer free?

Catalogizer is a self-hosted application. You run it on your own hardware or server. There are no subscription fees, cloud dependencies, or usage limits.

### What platforms does Catalogizer run on?

The backend (catalog-api) runs on any platform that supports Go: Linux, macOS, and Windows. The web frontend runs in any modern browser (Chrome 90+, Firefox 88+, Safari 14+, Edge 90+). Native apps are available for Windows, macOS, Linux (desktop), Android 8.0+ (phone), and Android TV.

### Does Catalogizer store my media files?

No. Catalogizer stores only metadata about your media files (titles, descriptions, quality information, organizational data). Your actual media files remain on their original storage locations. Catalogizer reads from those locations but does not copy or move files unless you explicitly use conversion or copy features.

---

## Installation

### What are the minimum system requirements?

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4 cores |
| RAM | 4 GB | 8 GB |
| Storage | 20 GB | 50 GB+ |
| Network | 100 Mbps | 1 Gbps |

### Should I use Docker/Podman or manual installation?

Container installation (Docker or Podman) is recommended for most users. It handles all dependencies automatically and provides consistent behavior. Manual installation is better for development or when you need fine-grained control over each component. If Docker is not available, Podman is fully supported as an alternative container runtime.

### Do I need PostgreSQL?

For development and small deployments, SQLite is the default database and requires no additional setup. For production deployments with multiple concurrent users, PostgreSQL 15+ is recommended. The Docker Compose configuration includes PostgreSQL automatically.

### Can I run Catalogizer behind a reverse proxy?

Yes. The project includes Nginx configuration files in `config/nginx.conf` and `config/nginx/catalogizer.prod.conf`. These handle TLS termination, request routing, and static file serving. You can adapt the configuration for other reverse proxies like Caddy or Traefik.

### How do I update Catalogizer?

Pull the latest code, rebuild the components, and restart. For container deployments: pull the latest images and run `podman-compose up -d` or `docker-compose up -d`. Database migrations run automatically on startup. Always back up your database before updating.

---

## Storage and Protocols

### Which storage protocols are supported?

Catalogizer supports five protocols through its UnifiedClient interface:
- **SMB/CIFS**: Windows and Samba file shares
- **FTP/FTPS**: Standard and secure FTP servers
- **NFS**: Network File System (Linux and macOS)
- **WebDAV**: HTTP-based file access
- **Local Filesystem**: Directly attached storage

### Can I add multiple storage sources?

Yes. You can connect as many storage sources as you need, using any combination of protocols. All sources are unified into a single catalog with a consistent browsing and search experience.

### What happens when a network storage source goes offline?

Catalogizer uses a resilience system for network storage, particularly SMB. A circuit breaker prevents repeated connection attempts to a downed server. An offline cache serves previously loaded metadata so users can continue browsing. Exponential backoff retry gradually attempts reconnection. When the source comes back online, the system automatically recovers and resumes normal operation.

### Can Catalogizer discover shares on my network?

Yes. The SMB discovery feature can auto-detect available SMB shares on your local network. This is useful for finding shares you did not know existed or for simplifying initial setup.

### Does Catalogizer modify files on my storage sources?

Catalogizer operates in read-only mode by default. It reads files to detect types and extract metadata but does not modify, move, or delete files on storage sources unless you explicitly initiate a file operation (such as copy, rename, or conversion).

---

## Media Detection and Metadata

### How many media types does Catalogizer detect?

Catalogizer detects over 50 media types, including movies, TV shows, music albums, individual tracks, games, software, documentaries, audiobooks, podcasts, ebooks, and more. The detection pipeline uses filename analysis, path structure, and file extension matching.

### Where does the metadata come from?

Metadata is fetched from external providers:
- **TMDB** (The Movie Database): Movies and TV shows
- **IMDB**: Movies and TV shows
- **TVDB**: TV series information
- **MusicBrainz**: Music albums and tracks
- **Spotify**: Music metadata and album art
- **Steam**: Games and software

You need API keys for some providers (TMDB is free). Metadata is cached locally to minimize external API calls.

### What if metadata is wrong for a file?

The detection pipeline uses heuristics and may occasionally misidentify a file. You can manually correct metadata through the web interface. Corrected metadata is stored locally and takes precedence over auto-detected information.

### Does Catalogizer detect video quality?

Yes. The analyzer component of the detection pipeline extracts technical metadata including resolution (720p, 1080p, 4K/UHD), codec (H.264, H.265, VP9), bitrate, and container format. Quality profiles allow comparison and ranking across versions of the same content.

---

## Media Player and Playback

### Can I play media directly in Catalogizer?

Yes. Catalogizer includes a built-in media player that handles video and audio playback in the browser. The player streams content from storage sources regardless of protocol. It supports playback controls, subtitle selection, and fullscreen mode.

### Does playback position sync across devices?

Yes. The playback position service saves your position whenever you pause or close the player. When you return to the same item from any device (web, desktop, or Android), playback resumes from where you left off.

### What subtitle formats are supported?

Catalogizer supports SRT, ASS, SSA, and VTT subtitle formats. Subtitles are automatically matched with videos based on naming conventions. You can also manually associate subtitle files and upload new ones through the Subtitle Manager page.

### Can I share a specific moment in a video?

Yes. The deep linking service generates links that include both the media item and an exact playback position. Recipients who open the link jump directly to the specified moment.

---

## Organization

### What is the difference between collections and playlists?

Collections organize media thematically. They can be Manual (hand-picked items), Smart (auto-populated based on rules), or Dynamic (adaptive criteria). Playlists are ordered sequences designed for sequential playback, with auto-advancement when one item finishes.

### How do Smart collections work?

You define rules such as "all movies from 2024" or "all 4K videos" when creating a Smart collection. Catalogizer automatically adds matching media to the collection. When new media matching the rules is detected in the future, it is added automatically without any manual intervention.

### Can I export my favorites?

Yes. Favorites can be exported to JSON or CSV format with full metadata. You can import favorites from these files on another Catalogizer instance. Matching uses metadata rather than file paths, so imports work even when files are on different storage sources.

---

## Security

### How is authentication handled?

Catalogizer uses JWT (JSON Web Token) authentication. Users log in with credentials and receive access and refresh tokens. The access token is included with every API request and validated by the auth middleware. Tokens have configurable expiry times.

### Is the database encrypted?

Yes. SQLCipher provides AES-256 encryption for the SQLite database at rest. The encryption key is configured via the DB_ENCRYPTION_KEY environment variable (exactly 32 characters). Without this key, the database file is unreadable.

### Does Catalogizer support two-factor authentication?

Yes. Administrators can enable 2FA for user accounts. Users scan a QR code with an authenticator app and provide a verification code during login in addition to their password.

### How do I run security scans?

Catalogizer includes scripts for automated security testing:
- `scripts/security-test.sh` for security-focused tests
- `scripts/snyk-scan.sh` for dependency vulnerability scanning
- `scripts/sonarqube-scan.sh` for static code analysis

---

## Multi-Platform

### Do mobile apps work offline?

Yes. The Android app uses Room database for local caching. Previously loaded metadata is available offline. When connectivity returns, the app syncs with the server automatically.

### Can I build the desktop app for my platform?

Yes. The desktop app uses Tauri and produces native installers:
- Windows: MSI installer
- macOS: DMG image
- Linux: AppImage or .deb package

Build with `npm run tauri:build` in the `catalogizer-desktop/` directory.

### What is the API client library for?

The TypeScript API client library (`catalogizer-api-client`) provides programmatic access to the Catalogizer API. Use it to build custom dashboards, automation scripts, batch processing tools, or entirely new client applications.

---

## Administration

### How do I back up Catalogizer?

Back up three things:
1. The SQLCipher database file (most critical -- contains all metadata and user data)
2. The `.env` configuration files (server settings and API keys)
3. The `config/` directory (nginx.conf, redis.conf)

Store the DB_ENCRYPTION_KEY separately from the database backup for security. Test your restore procedure monthly.

### How do I monitor Catalogizer in production?

Catalogizer exposes Prometheus-compatible metrics. The project includes a pre-configured Prometheus configuration and Grafana dashboards in the `monitoring/` directory. Dashboards cover API performance, media detection throughput, and storage source health.

### How do I troubleshoot connection issues?

Set `LOG_LEVEL=debug` in your `.env` file and restart the backend. Check the logs for circuit breaker state transitions, retry attempts, and error details. The Troubleshooting Guide in the documentation provides step-by-step diagnosis procedures for common issues.

---

## Development

### How do I contribute?

See the Contributing Guide (`docs/CONTRIBUTING.md`) for detailed instructions. The basic workflow is: set up the development environment, make changes following existing code patterns, write tests, and submit a pull request. Run `scripts/run-all-tests.sh` before submitting.

### How do I add support for a new storage protocol?

Implement the UnifiedClient interface defined in `filesystem/interface.go`. Create your client file (e.g., `filesystem/s3_client.go`), update the factory in `filesystem/factory.go` to recognize your protocol, and write tests. No other application code needs to change.

### How do I add a new metadata provider?

Study the provider interface in `internal/media/providers/providers.go` and use existing providers like `movie_recognition_provider.go` as a reference. Create your provider, implement the interface, and register it in the provider system. Handle API authentication, rate limiting, and error cases.

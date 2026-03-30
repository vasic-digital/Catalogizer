# Module 2: Getting Started with Media Management - Slide Deck Outline

**Total Slides**: 12
**Estimated Duration**: 75 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Getting Started with Media Management

- Navigate the web UI, connect storage, browse, search, and analyze
- Prerequisites: Module 1 completed, Catalogizer running
- By the end: confidently navigate the full web interface

---

## Slide 2: Web UI Overview (5 min)

**Title**: Dashboard and Navigation Layout

- Main navigation: Dashboard, Media, Collections, Search, Profile menu
- Quick Stats panel: total media, collections, favorites, storage used
- Recent activity feed and system notifications
- Quick actions: Upload Media, Create Collection, Import, View Analytics
- Demo: walk through each navigation section

---

## Slide 3: Connecting SMB Storage (6 min)

**Title**: Adding an SMB/CIFS Share

- Navigate to Storage Sources in the admin panel
- Enter server address, share name, credentials, domain settings
- UnifiedClient interface abstracts all protocol differences
- SMB discovery auto-detects available network shares
- Circuit breaker + offline cache for network resilience
- Exercise reference: Exercise 2.1 -- connect a network share

---

## Slide 4: Other Protocols (5 min)

**Title**: FTP, NFS, WebDAV, and Local Sources

- FTP/FTPS: hostname, port, credentials, passive mode
- NFS: server:/export path with mount options
- WebDAV: HTTPS URL with basic or digest authentication
- Local filesystem: direct path to any accessible directory
- Each protocol uses the same UnifiedClient interface

---

## Slide 5: Triggering a Scan (5 min)

**Title**: Scanning Storage Sources

- Initiate scan from the storage root management panel
- Universal Scanner crawls connected sources recursively
- Post-scan aggregation: title parsing, entity creation, hierarchy building
- Real-time progress via WebSocket push events
- Demo: trigger a scan and observe the progress in the UI

---

## Slide 6: Browsing the Catalog (6 min)

**Title**: Grid View, List View, and Filters

- Switch between Grid and List views in the Media Browser
- Filter by type: Images, Videos, Documents, Audio
- Filter by date range, file size, storage source
- Sort by name, date, size, type, or relevance
- Open Media Detail Modal for metadata and quality info
- Demo: browse and filter a populated catalog

---

## Slide 7: Media Entity Browser (6 min)

**Title**: Structured Media Entities

- 11 media types: movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic
- Hierarchical browsing: TV Show -> Seasons -> Episodes
- Entity detail page with associated files and metadata
- Duplicate detection across storage sources
- Exercise reference: Exercise 2.2 -- browse entities by type

---

## Slide 8: Search and Discovery (6 min)

**Title**: Finding What You Need

- Advanced search with filters, tags, and metadata queries
- Filter results by media type and category
- External metadata from TMDB, MusicBrainz enriches search
- Recommendation service for content discovery
- Demo: search for a movie and inspect its metadata

---

## Slide 9: Analytics Dashboard (5 min)

**Title**: Library Statistics and Insights

- Access Analytics page for comprehensive library statistics
- Growth trends: new files over time, storage consumption
- Quality analysis: resolution distribution, codec breakdown
- AI Dashboard for intelligent insights
- Exercise reference: Exercise 2.3 -- review analytics after scanning

---

## Slide 10: Real-Time Updates (5 min)

**Title**: Live Updates Without Refreshing

- Event bus captures file changes, metadata updates, scan events
- WebSocket server pushes to all connected clients
- AuthContext and WebSocketContext distribute events in React
- No polling: instant updates across all connected devices
- Demo: drop a file on a share and watch it appear live

---

## Slide 11: Localization (5 min)

**Title**: Multi-Language Support

- Configure language preferences in profile settings
- Localization service and translation service in the backend
- Search and browse media in multiple languages
- Multi-language metadata enrichment from providers

---

## Slide 12: Module Summary and Next Steps (4 min)

**Title**: What We Covered

- Navigated the dashboard, connected storage sources, scanned media
- Browsed the catalog with filters, sorting, and entity hierarchy
- Used advanced search and analytics
- Experienced real-time WebSocket updates
- Next module: Advanced Media Features (favorites, collections, playlists)
- Homework: connect a second storage source and compare scan results

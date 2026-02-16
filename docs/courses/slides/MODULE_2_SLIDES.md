# Module 2: Getting Started - Slide Outlines

---

## Slide 2.1.1: Title Slide

**Title**: Getting Started with Media Management

**Subtitle**: Navigating the Web Interface, Connecting Sources, and Discovering Your Media

**Speaker Notes**: This module assumes a running Catalogizer instance from Module 1. We cover the web UI, connecting storage, browsing, searching, analytics, and localization.

---

## Slide 2.1.2: The Web Interface

**Title**: Dashboard -- Your Command Center

**Bullet Points**:
- Main navigation: Dashboard, Media, Collections, Search, Profile
- Quick Stats panel: total media count, collections, favorites, storage used
- Recent Activity: uploads, collection updates, shared items, system notifications
- Quick Actions: Upload Media, Create Collection, Import from Cloud, View Analytics
- All data updates in real-time via WebSocket

**Visual**: Annotated screenshot of the Dashboard page

**Speaker Notes**: Point out each UI element. Emphasize that the Quick Stats numbers update live. Demonstrate by adding a file and watching the count increase.

---

## Slide 2.1.3: Authentication and Context Architecture

**Title**: How the Frontend Is Organized

**Bullet Points**:
- **AuthContext**: Manages JWT tokens, session state, and login/logout
- **WebSocketContext**: Establishes real-time connection, distributes events
- **ProtectedRoute**: Guards authenticated pages; redirects to login if unauthenticated
- **React Query**: Manages server state with automatic caching and revalidation

**Visual**: Context hierarchy diagram: AuthProvider -> WebSocketProvider -> Router -> ProtectedRoute -> Page Components

**Speaker Notes**: These two contexts wrap the entire application. Every component has access to auth state and live events. This is a clean architecture pattern that separates concerns.

---

## Slide 2.2.1: Connecting Storage Sources

**Title**: Five Protocols, One Interface

**Bullet Points**:
- **SMB/CIFS**: Windows/Samba shares; server, share name, credentials, domain
- **FTP/FTPS**: File Transfer Protocol; server, port, credentials
- **NFS**: Network File System; automatic mounting; platform-specific (Linux/macOS)
- **WebDAV**: HTTP-based; URL and credentials; cloud storage compatible
- **Local**: Direct filesystem path; simplest option

**Visual**: Protocol comparison table with icons

**Speaker Notes**: Start with local or SMB as they are the most common. Each protocol uses the same UnifiedClient interface internally, so behavior is consistent.

---

## Slide 2.2.2: The UnifiedClient Interface

**Title**: One Interface for All Protocols

**Bullet Points**:
- Defined in `filesystem/interface.go`
- Common methods: list files, read metadata, connect, disconnect
- Factory pattern in `filesystem/factory.go` creates the right client per protocol
- Application code is protocol-agnostic
- Adding a new protocol means implementing this single interface

**Visual**: Code snippet showing the UnifiedClient interface methods

**Speaker Notes**: This is a key architectural decision. The rest of Catalogizer does not know or care whether a file is on an SMB share or a local disk. The factory examines the URL scheme and creates the appropriate client.

---

## Slide 2.2.3: SMB Resilience

**Title**: Built for Unreliable Networks

**Bullet Points**:
- **Circuit Breaker**: Prevents repeated connections to a downed server
- **Exponential Backoff Retry**: Doubles delay between reconnection attempts
- **Offline Cache**: Serves previously loaded data during outages
- **SMB Discovery**: Auto-detects available shares on the network
- Universal Scanner begins cataloging immediately after connection

**Visual**: State diagram: Closed (Healthy) -> Open (Disconnected) -> Half-Open (Testing) -> Closed (Recovered)

**Speaker Notes**: SMB connections are inherently unreliable. The circuit breaker pattern is borrowed from distributed systems engineering. It prevents wasted resources while maintaining user experience through caching.

---

## Slide 2.3.1: Browsing Your Catalog

**Title**: Grid View and List View

**Bullet Points**:
- **Grid View**: Thumbnail cards with title, type, and quick-action buttons (MediaGrid.tsx, MediaCard.tsx)
- **List View**: Tabular format with filename, type, size, and date columns
- **Filters**: By type (Images, Videos, Documents, Audio), date range, file size
- **Sorting**: Name, date added, size, type, relevance
- Real-time updates keep the view current without refresh

**Visual**: Side-by-side screenshot of Grid and List views

**Speaker Notes**: Grid view is best for visual browsing of movies and photos. List view works better for comparing file sizes or when you need dense information. Switch between them based on your current task.

---

## Slide 2.3.2: Media Detail Modal

**Title**: Rich Metadata at a Glance

**Bullet Points**:
- Click any item to open the MediaDetailModal
- Shows: filename, path, source protocol, detected media type, file size, dates
- External metadata: poster, synopsis, cast, ratings (from TMDB, IMDB)
- Quality information: resolution, codec, bitrate
- Version tracking across sources

**Visual**: Screenshot of MediaDetailModal showing a movie with TMDB metadata

**Speaker Notes**: The detail modal is where the metadata enrichment really shines. A file that was just "Inception.mkv" now has a poster, plot summary, cast list, and quality details. This is all fetched automatically from external providers.

---

## Slide 2.4.1: Search and Discovery

**Title**: Finding Media Across Your Entire Library

**Bullet Points**:
- Full-text search across filenames, titles, tags, and all metadata
- Category filters: movies, TV shows, music, games, software, documentaries
- Search by metadata not in the filename: director, actor, genre, year
- Recommendation service suggests related content based on patterns
- Duplicate detection identifies the same content across sources

**Visual**: Screenshot of search results with category filters applied

**Speaker Notes**: Because Catalogizer fetches metadata from TMDB and other providers, you can search for a director's name and find all their movies, even if the filenames are just random abbreviations. This is the power of metadata enrichment combined with search.

---

## Slide 2.4.2: Recommendations and Deduplication

**Title**: Smart Discovery Features

**Bullet Points**:
- **Recommendation Service** (`recommendation_service.go`): Suggests media based on interaction patterns
- **Duplicate Detection** (`duplicate_detection_service.go`): Flags identical content across sources
- Helps identify redundant copies for storage cleanup
- Recommendations improve as you use the system more

**Speaker Notes**: Recommendations are opt-in and based on your own usage patterns. Duplicate detection is especially useful if you have the same movies on multiple drives and want to consolidate.

---

## Slide 2.5.1: Analytics Dashboard

**Title**: Understanding Your Library

**Bullet Points**:
- **Library Composition**: Breakdown by media type (movies, music, documents, etc.)
- **Growth Trends**: How your library has expanded over time
- **Quality Analysis**: Video resolution and codec distribution
- **Storage Statistics**: Space usage per source and media type
- **AI Dashboard** (`AIDashboard.tsx`): Intelligent insights and pattern-based suggestions

**Visual**: Chart showing media type distribution and growth trend

**Speaker Notes**: Analytics help you make informed decisions about storage planning. If you see that 60% of your storage is video files, and most are below 1080p, you might plan for upgrades. The AI Dashboard takes this further with automated insights.

---

## Slide 2.5.2: Reporting and Export

**Title**: Taking Data Beyond the Dashboard

**Bullet Points**:
- Export analytics data for external reporting tools
- Reporting service (`reporting_service.go`) generates professional PDF reports
- Analytics repository and service manage data aggregation
- Prometheus metrics feed into Grafana for operational monitoring
- Reports serve as point-in-time library snapshots

**Speaker Notes**: Export capabilities let you integrate Catalogizer data with other business intelligence tools. PDF reports are useful for documentation or presentations about your media infrastructure.

---

## Slide 2.6.1: Localization and Multi-Language Support

**Title**: Catalogizer Speaks Your Language

**Bullet Points**:
- **Localization Service** (`localization_service.go`): Translates UI strings and labels
- **Translation Service** (`translation_service.go`): Translates media metadata
- Language preference configurable per user in profile settings
- TMDB and other providers supply metadata in multiple languages
- Search works across translated metadata
- New languages added through translation files without code changes

**Speaker Notes**: If you have a multilingual household or team, each user can set their own language preference. Media metadata from TMDB is available in dozens of languages, so a French user sees French titles and synopses.

---

## Slide 2.6.2: Module 2 Summary

**Title**: What We Covered

**Bullet Points**:
- Dashboard provides real-time overview with Quick Stats, Activity, and Quick Actions
- Five storage protocols connected through a unified interface
- Grid and List views with filtering, sorting, and real-time updates
- Full-text search with category filters and metadata-enriched discovery
- Analytics dashboard with growth trends, quality analysis, and AI insights
- Localization supports multiple languages for both UI and media metadata

**Next Steps**: Module 3 -- Media Management (Favorites, Collections, Playlists, Subtitles, Conversion, Player)

**Speaker Notes**: Ensure students are comfortable navigating the UI and have at least one storage source connected. Module 3 builds on this foundation with advanced organizational and playback features.

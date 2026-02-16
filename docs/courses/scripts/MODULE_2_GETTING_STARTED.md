# Module 2: Getting Started with Media Management - Video Scripts

---

## Lesson 2.1: Web UI Overview & Dashboard

**Duration**: 12 minutes

### Narration

Now that Catalogizer is installed and configured, let us explore the web interface. The frontend is a React TypeScript application using Tailwind CSS for styling and React Query for server state management.

When you first log in, you land on the Dashboard. This is your command center. At the top, you see the main navigation bar with links to Dashboard, Media, Collections, and Search. On the right side is your profile menu for account settings and logout.

The Dashboard page -- implemented in Dashboard.tsx -- shows a Quick Stats panel at the top. This displays four key numbers: your total media count, number of collections, total favorites, and storage used. These numbers update in real-time via WebSocket.

Below the stats, you will find Recent Activity showing your latest uploads, collection updates, shared items, and system notifications. This gives you a quick pulse on what has changed in your library.

The Quick Actions section provides one-click access to the most common tasks: Upload Media, Create Collection, Import from Cloud, and View Analytics.

The application uses two important React contexts. AuthContext handles authentication state, JWT tokens, and session management. WebSocketContext establishes and maintains the real-time connection to the backend, distributing events to all listening components. These wrap the entire application, so every page has access to authentication and live updates.

Protected routes ensure that unauthenticated users are redirected to the login page. The ProtectedRoute component checks your auth state before rendering any page content.

### On-Screen Actions

- [00:00] Open browser to the Catalogizer login page
- [00:30] Log in with credentials
- [01:00] Show the Dashboard page loading with Quick Stats panel
- [02:00] Point to each navigation element: Dashboard, Media, Collections, Search, Profile
- [03:00] Highlight the Quick Stats panel: total media, collections, favorites, storage
- [04:00] Scroll to Recent Activity section
- [05:00] Demonstrate Quick Actions: click each one briefly to show what opens
- [06:00] Open browser dev tools and show WebSocket connection in Network tab
- [07:00] Show a real-time update arriving: add a file to a storage source and watch it appear
- [08:00] Open AuthContext.tsx briefly in the code editor to show structure
- [09:00] Open WebSocketContext.tsx to show event handling
- [10:00] Demonstrate logging out and being redirected to login
- [10:30] Show the ProtectedRoute component in the code
- [11:00] Log back in and return to Dashboard

### Key Points

- Dashboard provides Quick Stats (media count, collections, favorites, storage), Recent Activity, and Quick Actions
- Main navigation: Dashboard, Media, Collections, Search, Profile menu
- AuthContext manages JWT-based authentication; WebSocketContext manages real-time updates
- ProtectedRoute gates all authenticated pages
- All data updates arrive in real-time via WebSocket -- no manual refresh needed

### Tips

> **Tip**: Keep an eye on the Quick Stats panel after connecting new storage sources. Watching the media count increase in real-time confirms that detection is working correctly.

### Quiz Questions

1. **Q**: What are the four Quick Stats displayed on the Dashboard?
   **A**: Total media count, number of collections, total favorites, and storage used.

2. **Q**: What are the two main React contexts that wrap the entire application?
   **A**: AuthContext (authentication state and JWT tokens) and WebSocketContext (real-time connection to backend).

3. **Q**: What happens when an unauthenticated user tries to access a protected page?
   **A**: The ProtectedRoute component redirects them to the login page.

---

## Lesson 2.2: Connecting Storage Sources

**Duration**: 15 minutes

### Narration

Catalogizer supports five storage protocols, all abstracted behind a UnifiedClient interface. Let us connect each type and understand how they work.

Starting with SMB, the most common protocol for home and office networks. SMB sources are configured through environment variables or the admin interface. You provide the server address, share name, username, password, and optionally a domain. Behind the scenes, the smb_client.go implementation handles the connection with built-in resilience -- circuit breaker, exponential backoff retry, and offline caching.

For FTP connections, you specify the server address, port, username, and password. The ftp_client.go handles both FTP and FTPS connections. This is useful for accessing remote file servers over the internet.

NFS is supported with automatic mounting capabilities. The implementation varies by platform -- there is nfs_client_darwin.go for macOS and nfs_client.go for the general implementation. NFS is ideal for Linux and macOS environments where you need high-performance local network access.

WebDAV provides HTTP-based file access, implemented in webdav_client.go. This is useful for connecting to cloud storage services that support WebDAV, or for accessing files through web servers.

Local filesystem access is the simplest, handled by local_client.go. Just point Catalogizer to a directory and it will catalog everything inside.

The factory pattern in filesystem/factory.go examines the protocol in the source URL and creates the appropriate client. This means the rest of the application does not care which protocol is used -- it works with the same UnifiedClient interface regardless.

When you add a source, Catalogizer immediately begins scanning. The universal scanner in universal_scanner.go crawls the file tree, and the media detection pipeline identifies each item. SMB discovery (smb_discovery.go) can also auto-discover available shares on your network.

### On-Screen Actions

- [00:00] Navigate to source management in the web UI
- [01:00] Add an SMB source: enter server address, share, credentials
- [02:30] Show the connection establishing and initial scan starting
- [03:30] Open the code: filesystem/interface.go -- highlight UnifiedClient interface methods
- [04:30] Open filesystem/smb_client.go -- show connection logic
- [05:30] Add an FTP source: enter server, port, credentials
- [06:30] Open filesystem/ftp_client.go briefly
- [07:00] Explain NFS setup: show nfs_client.go and the platform-specific darwin variant
- [08:00] Add a WebDAV source: enter URL and credentials
- [08:30] Open filesystem/webdav_client.go
- [09:00] Add a local filesystem path
- [09:30] Open filesystem/local_client.go
- [10:00] Open filesystem/factory.go -- show how it selects the right client
- [10:30] Show universal_scanner.go beginning a scan
- [11:30] Open smb_discovery.go -- demonstrate network share discovery
- [12:30] Return to the web UI and show media appearing from all connected sources
- [13:30] Show the media browser with items from different protocols
- [14:00] Demonstrate filtering by source

### Key Points

- Five protocols supported: SMB/CIFS, FTP/FTPS, NFS, WebDAV, local filesystem
- UnifiedClient interface (filesystem/interface.go) abstracts all protocols behind common methods
- Factory pattern (filesystem/factory.go) creates the right client based on protocol
- SMB has built-in resilience: circuit breaker, retry with exponential backoff, offline caching
- SMB discovery can auto-detect available network shares
- Universal scanner crawls all connected sources and feeds the media detection pipeline

### Tips

> **Tip**: Start with a small directory when testing a new storage source. Once you confirm it works, expand to larger directories. This prevents long initial scan times during setup.

> **Tip**: Use SMB discovery to find available shares on your network before manually configuring them.

### Quiz Questions

1. **Q**: What interface do all storage protocol clients implement?
   **A**: The UnifiedClient interface defined in filesystem/interface.go.

2. **Q**: What three resilience mechanisms does the SMB client use?
   **A**: Circuit breaker, exponential backoff retry, and offline caching.

3. **Q**: What component crawls the file tree after a source is connected?
   **A**: The universal scanner (universal_scanner.go).

---

## Lesson 2.3: Browsing & Navigating the Catalog

**Duration**: 12 minutes

### Narration

With sources connected and media detected, let us explore how to browse your catalog effectively.

The Media Browser page, implemented in MediaBrowser.tsx, is your primary interface for viewing media. It supports two display modes: Grid view shows thumbnails in a responsive card layout using MediaGrid.tsx and MediaCard.tsx components. List view shows a detailed table with filename, type, size, and date columns.

Each media item is rendered as a MediaCard component. Cards show a thumbnail or icon, the title, media type, and quick-action buttons for favorites and playback.

Filtering is handled by the MediaFilters component. You can filter by type -- images, videos, documents, audio -- or by date range and file size. Sorting options include name, date added, file size, type, and relevance.

Clicking on any media item opens the MediaDetailModal. This modal shows the full metadata: filename, path, source protocol, detected media type and category, file size, and dates. If external metadata was fetched -- for example, from TMDB for a movie -- you will see the poster image, synopsis, cast information, ratings, and more.

Quality information is also displayed. Catalogizer performs automatic quality detection and version tracking, so you can see whether a video file is 720p, 1080p, or 4K, and what codec it uses.

All of this updates in real-time. The WebSocketContext delivers events to the media browser, so if a new file appears on a connected source, you will see it pop into view without refreshing. If a file is deleted or moved, it disappears.

### On-Screen Actions

- [00:00] Navigate to the Media page
- [00:30] Show the Grid view with thumbnail cards
- [01:30] Switch to List view and show the tabular format
- [02:30] Click on a MediaCard to show hover effects and quick actions
- [03:00] Open MediaFilters: filter by Video type
- [04:00] Clear filter and apply Date filter: "This Week"
- [04:30] Sort by Size descending to find largest files
- [05:00] Click on a movie item to open MediaDetailModal
- [05:30] Show TMDB metadata: poster, synopsis, cast, ratings
- [06:30] Click on a music file to show MusicBrainz/Spotify metadata
- [07:30] Show quality information for a video: resolution, codec, bitrate
- [08:30] Demonstrate real-time update: add a file to a connected source and watch it appear
- [09:30] Show the MediaGrid.tsx component structure in the code
- [10:00] Show MediaCard.tsx and its props
- [10:30] Show MediaDetailModal.tsx
- [11:00] Show MediaFilters.tsx filter options
- [11:30] Return to the web UI for a final overview

### Key Points

- Media Browser supports Grid and List views (MediaGrid.tsx, MediaCard.tsx)
- MediaFilters allows filtering by type, date range, and file size with multiple sort options
- MediaDetailModal shows full metadata including external provider enrichment (TMDB, IMDB, etc.)
- Quality detection shows resolution, codec, and bitrate for video files
- Real-time WebSocket updates keep the browser in sync without manual refresh

### Tips

> **Tip**: Use the Grid view for visual browsing of movies and photos. Switch to List view when you need to compare file sizes or sort by specific attributes.

### Quiz Questions

1. **Q**: What two display modes does the Media Browser support?
   **A**: Grid view (thumbnail cards) and List view (tabular format).

2. **Q**: What information is shown in the MediaDetailModal?
   **A**: Full metadata including filename, path, source protocol, media type, file size, dates, and external provider enrichment (posters, synopsis, cast, ratings).

3. **Q**: How does the browser stay up to date without manual refresh?
   **A**: WebSocketContext delivers real-time events from the backend to the media browser.

---

## Lesson 2.4: Search & Discovery

**Duration**: 12 minutes

### Narration

Catalogizer offers advanced search capabilities that go beyond simple filename matching. The search system leverages all the metadata collected during detection and enrichment.

You can search by filename, title, tags, or any metadata field. The search bar accepts natural queries -- type a movie name, an actor name, a genre, or even a year, and Catalogizer matches against the full metadata.

Results can be filtered by media category. The detection engine categorizes media into types like movies, TV shows, music, games, software, and documentaries. Applying a category filter narrows your results immediately.

External metadata makes search particularly powerful. Because Catalogizer fetches data from TMDB, IMDB, TVDB, MusicBrainz, and Spotify, you can search for things that are not in the filename at all. Search for a director name, and you will find their movies. Search for a music genre, and albums tagged with that genre appear.

The recommendation service, implemented in recommendation_service.go with a simple variant in simple_recommendation_handler.go, can suggest media based on your viewing and interaction patterns. When you view search results, you may see recommended items that match your interests.

The duplicate detection service (duplicate_detection_service.go) also plays a role in search. If your results include duplicate files across different sources, Catalogizer can identify and flag them, helping you clean up your library.

### On-Screen Actions

- [00:00] Click on Search in the navigation
- [00:30] Type a movie title and show results with TMDB metadata
- [01:30] Search for an actor name -- show movies matching that actor
- [02:30] Search for a music genre and show matching albums
- [03:30] Apply a category filter: Movies only
- [04:30] Change category filter to TV Shows
- [05:00] Clear filters and search for a file size range
- [05:30] Search by year to find media from a specific period
- [06:30] Show a recommended item appearing in results
- [07:30] Demonstrate duplicate detection: search for a title that exists on multiple sources
- [08:30] Show the recommendation_service.go in the code editor
- [09:00] Show duplicate_detection_service.go
- [09:30] Show the search handler (search.go) in handlers/
- [10:30] Return to the UI for a final demonstration
- [11:00] Save a search query for reuse

### Key Points

- Full-text search across filenames, titles, tags, and all metadata fields
- Category filters: movies, TV shows, music, games, software, documentaries, and more
- External metadata enables searching by director, actor, genre, and other enriched fields
- Recommendation service suggests related media based on user patterns
- Duplicate detection identifies the same content across different storage sources

### Tips

> **Tip**: When searching for movies, try the original language title as well as the English title. TMDB metadata often includes both, giving you more chances to find what you are looking for.

### Quiz Questions

1. **Q**: Can you search for metadata that is not in the filename?
   **A**: Yes. Because Catalogizer fetches data from TMDB, IMDB, and other providers, you can search by director, actor, genre, and other enriched fields.

2. **Q**: What service helps identify the same content across different storage sources?
   **A**: The duplicate detection service (duplicate_detection_service.go).

3. **Q**: What service provides content suggestions based on user patterns?
   **A**: The recommendation service (recommendation_service.go).

---

## Lesson 2.5: Analytics Dashboard

**Duration**: 9 minutes

### Narration

The Analytics page gives you deep insight into your media collection. Accessible from the navigation, the Analytics page presents charts, trends, and statistics about your library.

You will see your total library composition broken down by media type. How many movies, how many music files, how many documents -- all displayed in clear charts. Growth trends show how your library has expanded over time, which is useful for understanding storage planning needs.

Quality analysis breaks down your video collection by resolution and codec. You can see what percentage of your movies are 4K, 1080p, 720p, or lower. This helps identify candidates for upgrade or cleanup.

Storage statistics show how much space each source consumes, how much is used versus available, and which media types consume the most storage.

The AI Dashboard page, implemented in AIDashboard.tsx, takes analytics a step further. It provides intelligent insights derived from your media patterns, potentially flagging anomalies, suggesting organizational improvements, or predicting storage needs.

You can export analytics data for external use -- for example, generating spreadsheets or feeding data into other reporting tools. The reporting service (reporting_service.go) in the backend handles report generation.

The analytics repository (analytics_repository.go) manages the data storage for all analytics, while the analytics service (analytics_service.go) processes and aggregates the data for presentation.

The metrics backend (internal/metrics/metrics.go) also exposes Prometheus-compatible metrics, which ties into the Grafana dashboards for real-time operational monitoring. We will cover that in more detail in Module 5.

### On-Screen Actions

- [00:00] Navigate to Analytics in the web UI
- [00:30] Show the library composition chart by media type
- [01:30] Show growth trends over time
- [02:30] Display quality analysis: resolution breakdown for video files
- [03:30] Show storage statistics by source and media type
- [04:30] Navigate to the AI Dashboard page
- [05:00] Show intelligent insights and suggestions
- [06:00] Demonstrate exporting analytics data
- [06:30] Briefly show analytics_repository.go and analytics_service.go
- [07:00] Show reporting_service.go for report generation
- [07:30] Show a Grafana dashboard if available
- [08:00] Return to the analytics page for a recap

### Key Points

- Analytics page shows library composition, growth trends, quality analysis, and storage statistics
- AI Dashboard provides intelligent insights and pattern-based suggestions
- Quality analysis breaks down video files by resolution and codec
- Analytics data can be exported for external reporting
- Reporting service (reporting_service.go) generates reports
- Analytics repository and service manage data aggregation
- Prometheus-compatible metrics feed into Grafana dashboards for operational monitoring

### Tips

> **Tip**: Check the Analytics page regularly after connecting new sources. It helps verify that detection is categorizing your media correctly and that metadata enrichment is working.

### Quiz Questions

1. **Q**: What types of analysis does the Analytics page provide?
   **A**: Library composition by media type, growth trends, quality analysis (resolution/codec breakdown), and storage statistics.

2. **Q**: What does the AI Dashboard provide beyond standard analytics?
   **A**: Intelligent insights derived from media patterns, including anomaly detection, organizational suggestions, and storage need predictions.

3. **Q**: What backend service handles report generation?
   **A**: The reporting service (reporting_service.go).

---

## Lesson 2.6: Localization & Multi-Language Support

**Duration**: 10 minutes

### Narration

Catalogizer supports multiple languages for both its interface and media metadata. This lesson covers how to configure and use the localization features.

The localization system is implemented through two backend services. The localization_service.go manages the interface localization -- translating UI labels, messages, and system text into the user's preferred language. The translation_service.go handles media metadata translation, allowing media titles, descriptions, and other metadata to be displayed in different languages.

On the frontend, localization is integrated into the component hierarchy. The localization handlers in the backend (localization_handlers.go in internal/handlers/) expose API endpoints for fetching translated strings and managing language preferences.

When you set your language preference, the frontend requests all UI strings in that language from the backend. The system supports adding new languages by providing translation files without modifying the application code.

For media metadata, external providers like TMDB support multiple languages. When Catalogizer fetches metadata for a movie, it can request the title, synopsis, and other details in your preferred language. If a translation is not available, it falls back to the original language.

Search also respects language settings. You can search in your preferred language, and results will include matches from translated metadata. This makes it possible to find media even if you do not know its original title.

### On-Screen Actions

- [00:00] Navigate to the profile settings
- [00:30] Show the language preference setting
- [01:00] Change the language and show the interface updating
- [02:00] Navigate to the Media Browser and show translated category labels
- [03:00] Open a movie detail and show metadata in the selected language
- [04:00] Switch back to the original language and compare
- [05:00] Search for a movie using a translated title
- [05:30] Show results matching the translated metadata
- [06:00] Open localization_service.go in the code editor
- [06:30] Show translation_service.go
- [07:00] Show localization_handlers.go in internal/handlers/
- [07:30] Show localization_service_test.go to verify the tests
- [08:00] Demonstrate adding a new language translation
- [09:00] Show how TMDB metadata supports multiple languages
- [09:30] Final overview of localization capabilities

### Key Points

- Two backend services: localization_service.go (UI strings) and translation_service.go (media metadata)
- Localization handlers (internal/handlers/) expose API endpoints for language management
- Interface language can be changed in user preferences
- External providers like TMDB supply metadata in multiple languages
- Search works across translated metadata in the user's preferred language
- New languages can be added through translation files without code changes

### Tips

> **Tip**: If you have a multilingual media library, configure your preferred language in TMDB and other provider settings. This ensures metadata is fetched in the language you prefer by default.

> **Tip**: The translation service caches translated strings for performance. If you update translations, the cache refreshes automatically on the next request cycle.

### Quiz Questions

1. **Q**: What two services handle localization in the backend?
   **A**: localization_service.go handles UI string translations, and translation_service.go handles media metadata translations.

2. **Q**: Can you search for media using translated titles?
   **A**: Yes. Search respects language settings and includes matches from translated metadata.

3. **Q**: How do you add support for a new language?
   **A**: By providing translation files -- no application code changes are required.

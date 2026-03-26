# Module 9: Search & Sync - Video Scripts

---

## Lesson 9.1: Search API Deep Dive

**Duration**: 18 minutes

### Narration

Welcome to Module 9. This module covers the search, browse, and synchronization subsystems that form the backbone of how users find, navigate, and distribute their media across Catalogizer.

In this first lesson, we are going to explore the Search API in depth. Catalogizer provides three levels of search: file-level search, entity-level search, and advanced multi-field queries. Each operates on a different data model and serves a different use case.

The file-level search operates on raw scanned files. The primary endpoint is GET /api/v1/search/files. This endpoint accepts query parameters for full-text search, extension filtering, file type filtering, size range, date range, and pagination. The query parameter q performs full-text matching against filenames and paths. Extension and file_type parameters provide exact-match filtering. Size constraints use min_size and max_size in bytes. Date filtering uses modified_after and modified_before in RFC3339 format. Pagination is controlled by page and limit parameters, with defaults of page 1 and limit 100.

Pagination is not optional for production catalogs. A NAS with 85,000 files would overwhelm the frontend without it. The API enforces a maximum limit to prevent accidental full-catalog downloads.

The duplicate detection endpoint at GET /api/v1/search/files/duplicates identifies files that appear in multiple storage roots. This is valuable when you have the same media copied across SMB shares, local disks, and NFS mounts. Duplicates are identified by content hash, filename, or a combination of metadata attributes.

The advanced search endpoint at POST /api/v1/search/advanced accepts a JSON body for complex multi-field queries. This allows combining multiple conditions with AND/OR logic, nested filters, array-based inclusion and exclusion lists, and sorting by multiple fields. The JSON body supports structured queries that cannot be expressed through simple query parameters.

Let me show the actual handler. SearchHandler in handlers/search.go processes these requests. It validates input parameters, delegates to the file repository for database queries, and returns paginated results with total count metadata. The repository layer translates search criteria into SQL queries with proper parameterization -- the dialect abstraction ensures queries work identically on both SQLite and PostgreSQL.

Beyond file-level search, the entity-level search operates on structured media items. The endpoint GET /api/v1/media/search queries the media_items table, which contains aggregated entities: movies, TV shows, music albums, games, and software. Entity search is hierarchy-aware -- searching for a TV show returns the show with its seasons and episodes as children.

The SearchHandler supports multiple sort options: by name, date, size, type, or relevance. Relevance sorting uses a scoring algorithm that weighs title matches higher than path matches, exact matches higher than partial matches, and recent items higher than older ones.

All search endpoints require JWT authentication. The middleware validates the token, extracts the user ID, and ensures the user has permission to access the requested resources. Search results are scoped to storage roots the user has access to.

### On-Screen Actions

- [00:00] Show title: "Search API Deep Dive"
- [00:30] Open a browser and navigate to the Catalogizer web UI search page
- [01:00] Show the search bar and filter panel
- [01:30] Type a search query and show results populating in real time
- [02:00] Open catalog-api/main.go -- show search route registrations at lines 747-752
- [02:30] Show GET /search/files with query parameters in the browser
- [03:00] Demonstrate extension filter: search for `.mkv` files only
- [03:30] Demonstrate size filter: find files larger than 1 GB
- [04:00] Demonstrate date filter: find files modified in the last 7 days
- [04:30] Show pagination: page through results with page and limit parameters
- [05:00] Open handlers/search.go -- show the SearchFiles handler
- [05:30] Trace from handler to repository -- show SQL query construction
- [06:00] Show dialect abstraction: same query works on SQLite and PostgreSQL
- [06:30] Demonstrate GET /search/files/duplicates -- show duplicate results
- [07:00] Explain duplicate detection logic: content hash and filename matching
- [07:30] Open a REST client and send a POST /search/advanced request with JSON body
- [08:00] Show the JSON body structure with nested conditions
- [08:30] Show the response with matching results
- [09:00] Demonstrate GET /media/search for entity-level search
- [09:30] Search for a TV show and show hierarchy results
- [10:00] Open handlers/media_browse.go -- show the SearchMedia handler
- [10:30] Show media_item_repository.go -- show entity search queries
- [11:00] Demonstrate sorting: sort by name, date, size, relevance
- [11:30] Show relevance scoring: title match vs path match
- [12:00] Open middleware/auth.go -- show JWT validation on search endpoints
- [12:30] Demonstrate unauthorized access: search without a token returns 401
- [13:00] Show search results scoped to user-accessible storage roots
- [13:30] Open the web UI search page -- show all filters working together
- [14:00] Demonstrate combining query text, extension, size, and date filters
- [14:30] Show the URL parameters generated by the frontend filters
- [15:00] Discuss performance: indexed columns, query optimization
- [15:30] Show explain plan for a search query on SQLite
- [16:00] Discuss the difference between file search and entity search use cases
- [16:30] Show React Query integration: useQuery hooks for search
- [17:00] Recap search API endpoints and filtering capabilities

### Key Points

- Three search endpoint groups: /search/files, /search/files/duplicates, /search/advanced
- Entity search: /media/search queries aggregated media items with hierarchy awareness
- Query parameters: q (text), extension, file_type, min_size, max_size, modified_after, modified_before
- Pagination: page and limit parameters, default page 1 limit 100, enforced maximum
- Advanced search: POST with JSON body for complex multi-field queries with AND/OR logic
- Duplicate detection: identifies files across multiple storage roots by hash and filename
- Relevance sorting: title matches scored higher than path matches, exact over partial, recent over old
- All endpoints require JWT authentication; results scoped to user-accessible storage roots
- Dialect abstraction ensures identical queries on SQLite and PostgreSQL

### Tips

> **Tip**: Use entity search (/media/search) when you want structured results with hierarchy. Use file search (/search/files) when you need raw file-level results. They serve different use cases and query different tables.

> **Tip**: For large catalogs, always use pagination. The default limit of 100 is a sensible starting point. Increase it only when you know the client can handle the response size.

### Quiz Questions

1. **Q**: What are the three file-level search endpoints?
   **A**: GET /search/files (full-text), GET /search/files/duplicates (duplicate detection), and POST /search/advanced (complex queries).

2. **Q**: Why is pagination important for search results?
   **A**: A NAS with 85,000 files would overwhelm the frontend without pagination. The API enforces a maximum limit to prevent accidental full-catalog downloads.

3. **Q**: How does entity search differ from file search?
   **A**: Entity search queries the media_items table for aggregated entities (movies, TV shows, etc.) with hierarchy awareness. File search queries raw scanned files.

---

## Lesson 9.2: Browse API and Storage Roots

**Duration**: 15 minutes

### Narration

The Browse API provides filesystem-like navigation through your catalog. Unlike search, which finds items matching criteria, browsing lets you drill down through the directory structure exactly as it exists on your storage sources.

All browse endpoints operate on indexed data, not the live filesystem. When you browse a directory, Catalogizer queries the database for files and subdirectories that were discovered during the last scan. This means browsing is fast regardless of the underlying storage protocol -- an SMB share on a remote NAS browses just as quickly as a local directory, because both are queries against the same database.

The tradeoff is that you see the catalog state at the last scan time, not the live filesystem. If a file was added to an SMB share after the last scan, it will not appear in browse results until the next scan completes. The watcher system and scheduled scans mitigate this by keeping the index fresh.

Storage roots are the entry points for navigation. A storage root represents a configured mount point or share. It can be an SMB share, an NFS export, an FTP server, a WebDAV collection, or a local directory. The endpoint GET /api/v1/browse/roots returns all configured storage roots that the current user has access to. The frontend renders these as top-level nodes in a file tree component.

Directory browsing uses GET /api/v1/browse/directory/*path. The path parameter is the full path within the catalog, starting from the storage root. The response includes files and subdirectories with their metadata: name, size, modification time, type, and whether the item is a directory.

File info retrieval uses GET /api/v1/browse/file-info/*path. This returns detailed metadata for a single file, including all detected media attributes, external metadata from providers, and quality analysis results.

Directory size aggregation is available via GET /api/v1/browse/directory-sizes/*path. This returns the total size of each subdirectory, which is useful for understanding storage distribution. For large directories, this is computed from indexed data rather than scanning the filesystem, making it performant even for deep hierarchies.

Duplicate detection within a subtree uses GET /api/v1/browse/duplicates/*path. This finds duplicate files within a specific directory tree, which is more focused than the global duplicate search in the Search API.

There is also entity-level browsing via GET /api/v1/entities/browse/:type. This endpoint groups media items by type -- you can browse all movies, all TV shows, all music albums, and so on. This provides a media-centric view rather than a filesystem-centric view.

The BrowseHandler in handlers/browse.go processes these requests. It validates paths, prevents path traversal attacks, queries the file repository, and enriches results with type information. The handler supports both JSON and directory listing formats.

On the frontend, the browse experience uses a tree component from the Media-Browser-React submodule. React Query manages server state with automatic caching and background refetching. Zustand stores UI state like the currently expanded directories and selected items.

### On-Screen Actions

- [00:00] Show title: "Browse API and Storage Roots"
- [00:30] Open the web UI and navigate to the Browse page
- [01:00] Show storage roots rendered as top-level nodes
- [01:30] Click a storage root to expand it -- show subdirectories loading
- [02:00] Open catalog-api/main.go -- show browse route registrations at lines 983-990
- [02:30] Show GET /browse/roots in a REST client
- [03:00] Show the storage root response: name, path, protocol, status
- [03:30] Show GET /browse/directory/*path -- browse into a directory
- [04:00] Show the directory listing: files and subdirectories with metadata
- [04:30] Drill deeper: navigate through multiple directory levels
- [05:00] Show GET /browse/file-info/*path for a specific file
- [05:30] Show detailed metadata: media type, quality, external metadata
- [06:00] Show GET /browse/directory-sizes/*path
- [06:30] Show directory size aggregation results
- [07:00] Show GET /browse/duplicates/*path for subtree duplicates
- [07:30] Open handlers/browse.go -- show the BrowseDirectory handler
- [08:00] Show path validation and traversal prevention logic
- [08:30] Trace to the file repository -- show indexed data queries
- [09:00] Show GET /entities/browse/:type -- entity-level browsing
- [09:30] Browse movies -- show media entities grouped by type
- [10:00] Browse TV shows -- show hierarchy: show -> seasons -> episodes
- [10:30] Open the web UI browse page -- show the tree component
- [11:00] Expand and collapse directories in the tree
- [11:30] Show React Query caching: previously loaded directories load instantly
- [12:00] Select a file in the tree -- show detail panel with metadata
- [12:30] Show entity browse in the web UI -- grid/list view toggle
- [13:00] Explain the difference between filesystem browse and entity browse
- [13:30] Show storage root management: adding and removing roots
- [14:00] Recap the Browse API endpoints and storage root system

### Key Points

- Browse operates on indexed data, not live filesystem -- fast regardless of protocol
- Tradeoff: shows catalog state at last scan time, not live filesystem
- GET /browse/roots: all configured storage roots the user can access
- GET /browse/directory/*path: directory listing with files and subdirectories
- GET /browse/file-info/*path: detailed metadata for a single file
- GET /browse/directory-sizes/*path: subdirectory size aggregation from indexed data
- GET /browse/duplicates/*path: duplicate detection within a subtree
- GET /entities/browse/:type: media-centric browsing by entity type (movie, TV show, etc.)
- Frontend: tree component from Media-Browser-React, React Query for caching, Zustand for UI state
- Path validation prevents traversal attacks in the browse handler

### Tips

> **Tip**: Entity browsing is better for media consumption -- finding a movie to watch or an album to listen to. Filesystem browsing is better for organization tasks -- understanding storage layout, finding large directories, and identifying duplicates.

> **Tip**: Storage roots are the bridge between physical storage and the catalog. Adding a new root triggers an initial scan. Removing a root removes its entries from the catalog but does not delete any files.

### Quiz Questions

1. **Q**: Why does the Browse API operate on indexed data instead of querying the live filesystem?
   **A**: For performance -- it makes browsing fast regardless of the underlying storage protocol (SMB, NFS, FTP, etc.) because all queries hit the local database.

2. **Q**: What is the difference between filesystem browsing and entity browsing?
   **A**: Filesystem browsing navigates the directory structure as it exists on storage. Entity browsing groups media items by type (movies, TV shows, music) with hierarchy awareness.

3. **Q**: How does the browse handler prevent path traversal attacks?
   **A**: It validates paths in the BrowseDirectory handler, checking for and rejecting path traversal sequences before querying the repository.

---

## Lesson 9.3: Cloud Synchronization

**Duration**: 16 minutes

### Narration

The synchronization system extends Catalogizer beyond local access to cloud storage. It lets you replicate your catalog data to remote destinations for backup, distribution, or multi-site access.

The sync system follows the same Handler-Service-Repository pattern as every other Catalogizer service. The core concepts are sync endpoints, sync sessions, and sync schedules. A sync endpoint defines a connection to a remote storage destination -- its type, URL, credentials, and sync direction. A sync session tracks a single execution of a sync operation with file-level progress counters. A sync schedule enables recurring synchronization with cron-like timing.

Supported sync providers include Amazon S3, Google Cloud Storage, WebDAV, and local directory targets. Each provider implements the same interface, so the sync logic is protocol-agnostic.

Sync endpoints are created with POST /api/v1/sync/endpoints. The request body includes the endpoint type (s3, gcs, webdav, or local), the connection URL or bucket name, credentials, and the sync direction: push, pull, or bidirectional. Push copies data from Catalogizer to the remote destination. Pull copies from the remote to Catalogizer. Bidirectional synchronizes both ways, handling conflicts based on modification time.

The endpoint management API provides full CRUD operations. GET /api/v1/sync/endpoints lists all endpoints for the current user. PUT /api/v1/sync/endpoints/:id updates an endpoint. DELETE /api/v1/sync/endpoints/:id removes it. The system validates connectivity when an endpoint is created or updated -- if the remote destination is unreachable, the API returns a 502 error.

To execute a sync, use POST /api/v1/sync/endpoints/:id/sync. This starts an asynchronous sync operation and returns a session ID. The session tracks progress: total files to sync, files completed, files failed, bytes transferred, and the current status (pending, in_progress, completed, failed).

Session history is available via GET /api/v1/sync/sessions. Each session records its start time, end time, status, and detailed file-level results. This provides a complete audit trail of all sync operations. Individual session details are at GET /api/v1/sync/sessions/:id.

For recurring synchronization, POST /api/v1/sync/schedules creates a schedule. You specify the endpoint, the cron expression for timing, and whether it should be active. The schedule tracks last_run and next_run timestamps. Aggregate statistics across all sessions are available at GET /api/v1/sync/statistics.

The SyncService in services/sync_service.go handles the business logic. It validates credentials, manages connection pooling, handles retries on transient failures, and coordinates file transfers. The service is user-scoped -- each user manages their own sync endpoints independently.

Credentials are stored securely. The password field is excluded from JSON serialization, so it never appears in API responses. Connection validation happens server-side before persisting endpoint configuration.

The sync repository in repository/sync_repository.go manages persistence. It stores endpoints, sessions, schedules, and statistics in the database. The dual-dialect abstraction ensures this works on both SQLite and PostgreSQL.

Session cleanup is important for long-running installations. POST /api/v1/sync/cleanup removes old completed sessions based on a retention period. This prevents the session history from growing unbounded.

On the frontend, the sync management interface provides a dashboard showing all configured endpoints, their last sync status, and scheduled runs. Users can trigger manual syncs, view session history, and monitor progress in real time via WebSocket updates.

### On-Screen Actions

- [00:00] Show title: "Cloud Synchronization"
- [00:30] Open the web UI sync management page
- [01:00] Show configured sync endpoints with status indicators
- [01:30] Open catalog-api/main.go -- show sync route registrations at lines 993-1006
- [02:00] Show the SyncEndpoint model: type, URL, credentials, direction
- [02:30] Create a new sync endpoint: POST /sync/endpoints with JSON body
- [03:00] Show endpoint types: s3, gcs, webdav, local
- [03:30] Show sync directions: push, pull, bidirectional
- [04:00] Demonstrate connectivity validation: create endpoint with bad URL, show 502 error
- [04:30] Create a valid local sync endpoint for demonstration
- [05:00] Show GET /sync/endpoints -- list all user endpoints
- [05:30] Trigger a sync: POST /sync/endpoints/:id/sync
- [06:00] Show the session ID returned
- [06:30] Monitor progress: GET /sync/sessions/:id with status updates
- [07:00] Show file-level counters: total, completed, failed, bytes transferred
- [07:30] Open services/sync_service.go -- show the sync execution logic
- [08:00] Show retry handling for transient failures
- [08:30] Show credential security: password excluded from JSON serialization
- [09:00] Show GET /sync/sessions -- session history with audit trail
- [09:30] Show session details: start time, end time, status, file results
- [10:00] Create a sync schedule: POST /sync/schedules with cron expression
- [10:30] Show the schedule response: last_run, next_run, active status
- [11:00] Show GET /sync/statistics -- aggregate statistics
- [11:30] Show session cleanup: POST /sync/cleanup
- [12:00] Open repository/sync_repository.go -- show database operations
- [12:30] Show dialect-agnostic queries working on both SQLite and PostgreSQL
- [13:00] Open the web UI -- show real-time sync progress via WebSocket
- [13:30] Demonstrate manual sync trigger from the UI
- [14:00] Show sync history and session details in the UI
- [14:30] Discuss use cases: backup to S3, multi-site distribution, offline copies
- [15:00] Recap the sync system: endpoints, sessions, schedules, statistics

### Key Points

- Sync endpoints: connection to remote storage (S3, GCS, WebDAV, local) with push/pull/bidirectional
- Full CRUD: POST/GET/PUT/DELETE for endpoints, plus sync execution and session history
- Session tracking: total files, completed, failed, bytes transferred, start/end times
- Sync schedules: cron-like recurring sync with last_run/next_run tracking
- Connectivity validation: 502 error if remote destination is unreachable
- Credential security: password field excluded from JSON serialization
- User-scoped: each user manages their own sync endpoints independently
- Session cleanup: POST /sync/cleanup removes old completed sessions
- Aggregate statistics: GET /sync/statistics for overview metrics
- Real-time progress: WebSocket updates during active sync operations

### Tips

> **Tip**: Start with a local sync endpoint for testing before configuring cloud providers. This lets you verify sync logic without network variables.

> **Tip**: Use bidirectional sync cautiously. Conflict resolution is based on modification time, which can produce unexpected results if clocks are not synchronized across systems. For backups, prefer push direction.

### Quiz Questions

1. **Q**: What are the three sync directions supported?
   **A**: Push (local to remote), pull (remote to local), and bidirectional (both ways with timestamp-based conflict resolution).

2. **Q**: How does the sync system protect credentials?
   **A**: The password field is excluded from JSON serialization, so it never appears in API responses. Connection validation happens server-side before persisting.

3. **Q**: What happens when a sync endpoint is created with an unreachable URL?
   **A**: The API validates connectivity and returns a 502 error if the remote destination is unreachable.

---

## Lesson 9.4: Media Entity System and Metadata Enrichment

**Duration**: 18 minutes

### Narration

The media entity system is the intelligence layer of Catalogizer. While the scanner discovers raw files, the entity system transforms them into structured, enrichable media objects with hierarchical relationships.

When a scan completes, the post-scan aggregation pipeline runs automatically. The AggregationService in internal/services/aggregation_service.go orchestrates this process. First, the title parser extracts structured information from filenames using regex patterns -- movie titles with years, TV show names with season and episode numbers, music with artist and album, game titles, and software names. Second, the system creates or updates MediaItem records in the media_items table. Third, it links files to entities via the media_files junction table. Fourth, it builds hierarchies: TV shows contain seasons which contain episodes, music artists contain albums which contain songs. Finally, it detects duplicates -- items with the same title, type, and year.

Catalogizer defines 11 media types, seeded in the media_types table: movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, and comic. Every scanned file must be associated with a recognized media entity after aggregation.

The entity hierarchy uses a parent_id self-reference in the media_items table. A TV show has no parent. A TV season has the show as its parent. An episode has the season as its parent. Similarly, a music artist has no parent, an album references the artist, and a song references the album. This allows navigating the full hierarchy with GET /api/v1/entities/:id/children.

The entity API provides comprehensive access. GET /api/v1/entities lists all entities with pagination and filtering. GET /api/v1/entities/types returns the 11 media types. GET /api/v1/entities/stats provides aggregate statistics. GET /api/v1/entities/:id returns a single entity with all metadata. GET /api/v1/entities/:id/children returns child entities. GET /api/v1/entities/:id/files returns associated files. GET /api/v1/entities/:id/metadata returns external metadata. GET /api/v1/entities/:id/duplicates shows duplicate entries.

Metadata enrichment is where the entity system becomes powerful. Catalogizer integrates with external metadata providers to augment entity records. POST /api/v1/entities/enrich triggers enrichment for all entities. POST /api/v1/entities/:id/metadata/refresh refreshes metadata for a single entity.

The external metadata providers include TMDB (The Movie Database) for movies and TV shows, providing plot summaries, cast lists, ratings, poster images, and trailers. OpenLibrary provides book metadata including author, publisher, cover art, and subject classifications. MusicBrainz provides music metadata including recording details, artist information, and release data. Additional providers handle games and software.

The enrichment pipeline in internal/media/ works as follows. The detector identifies the media type. The analyzer extracts additional attributes. The providers fetch external metadata and store it in the external_metadata table. User-generated metadata -- ratings, tags, notes -- is stored separately in the user_metadata table and managed via PUT /api/v1/entities/:id/user-metadata.

The media entity handler in handlers/media_entity_handler.go processes all entity requests. It coordinates between the media item repository, media file repository, external metadata repository, and user metadata repository. The handler supports streaming and downloading entities directly via GET /api/v1/entities/:id/stream and GET /api/v1/entities/:id/download.

Entity browsing via GET /api/v1/entities/browse/:type provides a type-filtered view. Browse movies to see all movies. Browse tv_show to see all TV shows (each expandable to seasons and episodes). This is the primary interface for media consumption, as opposed to the filesystem browser which shows raw directory structure.

### On-Screen Actions

- [00:00] Show title: "Media Entity System and Metadata Enrichment"
- [00:30] Open the web UI entity browser
- [01:00] Show entities organized by type: movies, TV shows, music, games
- [01:30] Click a movie -- show metadata: title, year, plot, cast, rating, poster
- [02:00] Open internal/services/aggregation_service.go -- show post-scan pipeline
- [02:30] Show title parser: regex extracting structured data from filenames
- [03:00] Open internal/services/title_parser.go -- show regex patterns
- [03:30] Show movie pattern: "Title (Year)" format
- [04:00] Show TV pattern: "Show S01E02" format
- [04:30] Show music pattern: "Artist - Album - Track" format
- [05:00] Show the media_types table: 11 seeded types
- [05:30] Show media_items table with parent_id hierarchy
- [06:00] Show a TV show hierarchy: show -> season -> episode
- [06:30] Show the media_files junction table linking files to entities
- [07:00] Open catalog-api/main.go -- show entity route registrations at lines 933-953
- [07:30] Show GET /entities -- list all entities with pagination
- [08:00] Show GET /entities/types -- the 11 media types
- [08:30] Show GET /entities/stats -- aggregate statistics
- [09:00] Show GET /entities/:id -- detailed entity view
- [09:30] Show GET /entities/:id/children -- child hierarchy
- [10:00] Show GET /entities/:id/files -- associated files
- [10:30] Show GET /entities/:id/metadata -- external metadata
- [11:00] Show POST /entities/enrich -- trigger enrichment for all entities
- [11:30] Show metadata appearing: TMDB data for movies, MusicBrainz for music
- [12:00] Open the enrichment pipeline code: detector -> analyzer -> providers
- [12:30] Show TMDB provider: API key configuration, data fetching
- [13:00] Show OpenLibrary provider: book metadata fetching
- [13:30] Show MusicBrainz provider: music metadata fetching
- [14:00] Show the external_metadata table structure
- [14:30] Show user metadata: PUT /entities/:id/user-metadata
- [15:00] Demonstrate adding user ratings and tags to an entity
- [15:30] Show entity streaming: GET /entities/:id/stream
- [16:00] Show entity browsing: GET /entities/browse/:type
- [16:30] Show the web UI entity detail page with all metadata combined
- [17:00] Recap the entity system: aggregation, hierarchy, enrichment

### Key Points

- Post-scan aggregation: title parser -> MediaItem creation -> file linking -> hierarchy building -> duplicate detection
- 11 media types: movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic
- Hierarchy: parent_id self-reference (TV: show -> season -> episode, Music: artist -> album -> song)
- Entity API: 13 endpoints covering listing, details, children, files, metadata, streaming, downloading
- Metadata enrichment: TMDB (movies/TV), OpenLibrary (books), MusicBrainz (music)
- Pipeline: detector identifies type -> analyzer extracts attributes -> providers fetch external metadata
- External metadata stored in external_metadata table; user metadata in user_metadata table
- Entity browsing: GET /entities/browse/:type for media-centric navigation
- Every scanned file must be associated with a recognized media entity after aggregation

### Tips

> **Tip**: Configure your TMDB API key before running the first scan. This way, entities are enriched immediately during the post-scan aggregation, rather than requiring a separate enrichment pass.

> **Tip**: The title parser uses regex patterns that match common naming conventions. If your files use unusual naming, some may not be correctly categorized on the first pass. Use user metadata to manually correct or augment entity records.

### Quiz Questions

1. **Q**: What are the five stages of the post-scan aggregation pipeline?
   **A**: Title parsing, MediaItem creation/update, file linking via junction table, hierarchy building, and duplicate detection.

2. **Q**: How does the entity hierarchy work for TV shows?
   **A**: TV shows use parent_id self-reference: show (no parent) -> season (parent = show) -> episode (parent = season).

3. **Q**: What external metadata providers does Catalogizer integrate with?
   **A**: TMDB for movies and TV shows, OpenLibrary for books, MusicBrainz for music, plus additional providers for games and software.

---

## Lesson 9.5: Advanced Search and Filtering in the Frontend

**Duration**: 13 minutes

### Narration

In this final lesson, we bring the search, browse, and entity APIs together in the frontend experience. The catalog-web frontend provides a unified interface for finding and navigating media that combines all the backend capabilities we have covered.

The search experience starts with the global search bar, accessible from every page. Typing a query triggers a debounced search that hits the /search/files endpoint. As results appear, the frontend categorizes them by media type and presents them in a dropdown with quick-access links. Pressing enter opens the full search results page with all filter controls.

The full search page uses React Query for server state management. The useQuery hook fetches search results, with automatic caching and background refetching. When you change a filter, React Query invalidates the previous query and fetches new results. The query key includes all active filters, so each unique filter combination is cached independently.

Filter state is managed by Zustand, the client-side state management library. This separates UI concerns (which filters are open, which view mode is active) from server state (the actual search results). Filter changes update the Zustand store, which triggers a React Query refetch.

The advanced filter panel provides controls for media type selection (checkboxes for each of the 11 types), file extension filter, size range slider, date range picker, and sorting options. These map directly to the query parameters we covered in the Search API lesson. The panel can be expanded or collapsed, and its state persists across navigation.

Entity browsing in the frontend provides a grid or list view of media items. The grid view shows poster images, titles, and ratings -- ideal for visual browsing of movies and TV shows. The list view shows detailed metadata in rows -- better for music albums and software where metadata density matters more than visual appeal. Users can toggle between views with a single click.

The entity detail page combines all metadata into a comprehensive view. The top section shows the poster or cover art, title, type, year, and primary metadata. Below that, external metadata from providers (TMDB plots, MusicBrainz recording details) appears in formatted sections. User metadata (ratings, tags, custom notes) is editable inline. Child entities (seasons, episodes, tracks) appear in a collapsible hierarchy.

The file tree component from the Media-Browser-React submodule provides the filesystem browsing experience. It renders storage roots at the top level, with directories expandable to reveal their contents. Lazy loading ensures only visible directories are fetched, keeping the interface responsive even for catalogs with tens of thousands of files. Selection state tracks which items are checked for batch operations.

React Hook Form with Zod validation handles the search form submission. Form state is validated client-side before making API calls, providing immediate feedback on invalid inputs. Zod schemas define the valid shape of search parameters, including type constraints and value ranges.

Real-time updates via WebSocket keep search results and browse views current. When a scan completes and new files are indexed, the WebSocket event triggers React Query to refetch active queries. This means if you are browsing a directory and a new file appears from a scan, the view updates automatically without manual refresh.

### On-Screen Actions

- [00:00] Show title: "Advanced Search and Filtering in the Frontend"
- [00:30] Open the web UI -- show the global search bar
- [01:00] Type a query -- show debounced dropdown results
- [01:30] Press enter -- navigate to full search results page
- [02:00] Show the filter panel: type, extension, size, date, sort
- [02:30] Toggle media type checkboxes -- show results filtering in real time
- [03:00] Adjust size range slider -- show results updating
- [03:30] Set a date range -- show date-filtered results
- [04:00] Change sort order: name, date, size, relevance
- [04:30] Open browser DevTools -- show React Query cache entries
- [05:00] Show each filter combination cached independently
- [05:30] Navigate away and back -- show cached results loading instantly
- [06:00] Open the entity browser in grid view
- [06:30] Show movie posters with titles and ratings
- [07:00] Switch to list view -- show detailed metadata rows
- [07:30] Click an entity -- open the entity detail page
- [08:00] Show combined metadata: poster, title, year, plot, cast, rating
- [08:30] Show user metadata editing: add a rating and tags inline
- [09:00] Show child entity hierarchy: expand seasons, show episodes
- [09:30] Open the filesystem browser -- show the file tree component
- [10:00] Expand directories -- show lazy loading in action
- [10:30] Select multiple files -- show batch operation controls
- [11:00] Demonstrate real-time update: trigger a scan, watch results appear via WebSocket
- [11:30] Show React Hook Form + Zod validation on the search form
- [12:00] Recap the frontend search and browse experience

### Key Points

- Global search bar with debounced dropdown results, accessible from every page
- React Query for server state: caching, background refetching, query key per filter combination
- Zustand for client state: filter panel state, view mode, UI preferences
- Advanced filters: media type, extension, size range, date range, sorting -- map to API query parameters
- Entity browse: grid view (visual, posters) and list view (metadata-dense, detailed rows)
- Entity detail: combined external and user metadata, editable inline, collapsible child hierarchy
- File tree: lazy loading, storage roots at top level, selection state for batch operations
- React Hook Form + Zod: client-side validation before API calls
- Real-time updates: WebSocket events trigger React Query refetch for live results

### Tips

> **Tip**: Use grid view for movie and TV show browsing where visual recognition from posters helps. Switch to list view for music, games, and software where reading metadata is more important than seeing cover art.

> **Tip**: If search results seem stale, check the last scan time. Browse results reflect the catalog state at last scan. You can trigger a manual rescan from the storage root management page.

### Quiz Questions

1. **Q**: What libraries manage server state and client state in the frontend?
   **A**: React Query manages server state (search results, entity data). Zustand manages client state (filter panel, view mode, UI preferences).

2. **Q**: How does the file tree component handle large catalogs?
   **A**: Through lazy loading -- only visible directories are fetched from the API. This keeps the interface responsive even for catalogs with tens of thousands of files.

3. **Q**: How do real-time updates work in the search and browse views?
   **A**: WebSocket events from scan completions trigger React Query to refetch active queries, so results update automatically without manual refresh.

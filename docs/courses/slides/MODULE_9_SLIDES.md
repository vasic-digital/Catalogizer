# Module 9: Search & Sync - Slide Outlines

---

## Slide 9.0.1: Title Slide

**Title**: Search & Sync

**Subtitle**: Search API, Browse API, Cloud Sync, Media Entities, and Metadata Enrichment

**Speaker Notes**: This module covers the subsystems that let users find, navigate, synchronize, and enrich their media. By the end, students will understand the search, browse, sync, entity, and metadata enrichment APIs.

---

## Slide 9.1.1: Search API Overview

**Title**: Three Search Endpoints

**Bullet Points**:
- `GET /api/v1/search/files` -- Full-text search with query parameters
- `GET /api/v1/search/files/duplicates` -- Duplicate file detection
- `POST /api/v1/search/advanced` -- Complex multi-field queries via JSON body
- Additional: `GET /api/v1/media/search` for media entity search
- All endpoints require JWT authentication

**Speaker Notes**: The search system operates at two levels: file-level (raw files) and entity-level (structured media items). File search is metadata based. Entity search operates on aggregated media items with hierarchy awareness.

---

## Slide 9.1.2: Search Filters

**Title**: SearchHandler Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Full-text query (filename + path) |
| `extension` / `file_type` | string | Type filters (exact match) |
| `min_size` / `max_size` | int | Size range in bytes |
| `modified_after` / `modified_before` | string | Date range (RFC3339) |
| `page` / `limit` | int | Pagination (default: page 1, limit 100) |

**Speaker Notes**: All filters are optional and combinable. Pagination is mandatory for large catalogs -- a NAS with 85,000 files would overwhelm the frontend without it. The advanced search endpoint accepts a JSON body for complex multi-field queries with arrays and nested objects.

---

## Slide 9.2.1: Browse API Architecture

**Title**: Directory Navigation Endpoints

**Bullet Points**:
- `GET /browse/roots` -- List all configured storage roots
- `GET /browse/directory/*path` -- List directory contents
- `GET /browse/file-info/*path` -- Single file metadata
- `GET /browse/directory-sizes/*path` -- Subdirectory size aggregation
- `GET /browse/duplicates/*path` -- Duplicates in subtree
- All operations query indexed data, not the live filesystem

**Speaker Notes**: The browse API is read-only and operates on indexed data. Browsing is fast regardless of storage protocol. The tradeoff is showing catalog state at last scan time, not live filesystem state.

---

## Slide 9.2.2: Storage Roots

**Title**: Entry Points for Navigation

**Bullet Points**:
- A storage root represents a configured mount point or share
- Types: SMB share, NFS export, FTP server, WebDAV collection, local directory
- Frontend renders roots as top-level nodes in a file tree component
- Entity browse: `GET /entities/browse/:type` for media-type navigation

**Speaker Notes**: Storage roots are created during setup or via the management API. The browse API queries the database regardless of source protocol, providing a unified view across all storage backends.

---

## Slide 9.3.1: Sync System Overview

**Title**: Cloud Synchronization Architecture

**Bullet Points**:
- Handler-Service-Repository pattern (same as all Catalogizer services)
- `SyncEndpoint`: configuration for a cloud destination (type, URL, credentials, direction)
- `SyncSession`: per-execution progress tracking with file-level counters
- `SyncSchedule`: cron-like recurring sync with last/next run tracking
- Supported providers: S3, GCS, WebDAV, local

**Speaker Notes**: The sync system is user-scoped -- each user manages their own endpoints. The service layer validates connectivity before persisting. Sessions provide audit trail and progress visibility.

---

## Slide 9.3.2: Sync API Endpoints

**Title**: Full CRUD Plus Execution

| Method | Path | Description |
|--------|------|-------------|
| POST | /sync/endpoints | Create endpoint |
| GET | /sync/endpoints | List user endpoints |
| PUT | /sync/endpoints/:id | Update endpoint |
| DELETE | /sync/endpoints/:id | Remove endpoint |
| POST | /sync/endpoints/:id/sync | Start sync |
| GET | /sync/sessions | List sessions |
| POST | /sync/schedules | Schedule recurring sync |
| GET | /sync/statistics | Aggregate statistics |

**Speaker Notes**: Sync directions are push (local to remote), pull (remote to local), or bidirectional. Credentials are stored securely with the password field excluded from JSON serialization. Connection validation returns 502 if unreachable.

---

## Slide 9.4.1: Media Entity System

**Title**: Post-Scan Aggregation Pipeline

**Bullet Points**:
- Scan completes -> AggregationService runs automatically
- Title parser extracts structured info from filenames (regex: movie, TV, music, game, software)
- MediaItem creation/update in `media_items` table with parent_id hierarchy
- File linking via `media_files` junction table
- Hierarchy: TV show -> season -> episode; Music artist -> album -> song
- Duplicate detection: same title + type + year

**Speaker Notes**: The entity system transforms raw scanned files into structured, enrichable media objects. Every scanned file must be associated with a recognized media entity after aggregation. The parent_id self-reference enables hierarchical navigation via GET /entities/:id/children.

---

## Slide 9.4.2: 11 Media Types and Entity API

**Title**: Entity Types and Endpoints

**Bullet Points**:
- 11 media types (seeded in `media_types`): movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic
- `GET /entities` -- list with pagination and filtering
- `GET /entities/types` -- the 11 media types
- `GET /entities/:id` -- single entity with all metadata
- `GET /entities/:id/children` -- child hierarchy
- `GET /entities/:id/files` -- associated files
- `GET /entities/:id/stream` -- stream media directly
- `GET /entities/browse/:type` -- type-filtered view for media consumption

**Speaker Notes**: The entity API provides 13 endpoints covering listing, details, hierarchy, files, metadata, streaming, and downloading. Entity browsing by type is the primary interface for media consumption, as opposed to the filesystem browser which shows raw directory structure.

---

## Slide 9.4.3: Metadata Enrichment Pipeline

**Title**: External Provider Integration

**Bullet Points**:
- Pipeline: detector identifies type -> analyzer extracts attributes -> providers fetch external metadata
- **TMDB** (The Movie Database): movies and TV shows -- plot summaries, cast, ratings, posters, trailers
- **OpenLibrary**: books -- author, publisher, cover art, subject classifications
- **MusicBrainz**: music -- recording details, artist information, release data
- `POST /entities/enrich` -- trigger enrichment for all entities
- `POST /entities/:id/metadata/refresh` -- refresh metadata for single entity
- External metadata stored in `external_metadata` table; user metadata in `user_metadata` table

**Speaker Notes**: Metadata enrichment is where the entity system becomes powerful. TMDB, OpenLibrary, and MusicBrainz provide rich external data. User-generated metadata (ratings, tags, notes) is stored separately and managed via PUT /entities/:id/user-metadata. Configure your TMDB API key before the first scan for immediate enrichment.

---

## Slide 9.5.1: Frontend Search and Browse Experience

**Title**: Unified Media Discovery in the Frontend

**Bullet Points**:
- Global search bar with debounced dropdown results on every page
- React Query for server state: caching, background refetching, query key per filter combination
- Zustand for client state: filter panel, view mode, UI preferences
- Advanced filters: media type, extension, size range, date range, sorting
- Entity browse: grid view (posters, visual) and list view (metadata-dense, detailed)
- File tree from Media-Browser-React: lazy loading, storage roots, batch selection
- Real-time updates: WebSocket events trigger React Query refetch on scan completion

**Speaker Notes**: The frontend combines all backend APIs into a unified experience. Grid view is ideal for movies and TV shows where poster recognition helps. List view is better for music and software where metadata density matters. WebSocket events keep views current as new scans complete.

---

## Slide 9.6.1: Module 9 Summary

**Title**: What We Covered

**Bullet Points**:
- Search API: three endpoint groups with rich filtering and pagination
- Browse API: storage root navigation, directory listing, size analysis
- Sync system: endpoint management, session tracking, scheduling
- Cloud providers: S3, GCS, WebDAV, local with push/pull/bidirectional modes
- Media entity system: post-scan aggregation, 11 types, parent_id hierarchy
- Metadata enrichment: TMDB (movies/TV), OpenLibrary (books), MusicBrainz (music)
- Frontend integration: React Query, Zustand, tree components, grid/list views

**Speaker Notes**: Search, browse, sync, and entity enrichment form the user-facing core of Catalogizer. Search finds media across all sources. Browse provides filesystem-like navigation. Sync extends to cloud storage. The entity system with metadata enrichment transforms raw files into rich, navigable media objects.

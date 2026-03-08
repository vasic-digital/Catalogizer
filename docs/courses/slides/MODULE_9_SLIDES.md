# Module 9: Search & Sync - Slide Outlines

---

## Slide 9.0.1: Title Slide

**Title**: Search & Sync

**Subtitle**: Search API, Browse API, Cloud Sync, and Provider Integration

**Speaker Notes**: This module covers the three subsystems that let users find, navigate, and synchronize their media. By the end, students will understand the search, browse, and sync APIs.

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

## Slide 9.4.1: Module 9 Summary

**Title**: What We Covered

**Bullet Points**:
- Search API: three endpoint groups with rich filtering and pagination
- Browse API: storage root navigation, directory listing, size analysis
- Sync system: endpoint management, session tracking, scheduling
- Cloud providers: S3, GCS, WebDAV, local with push/pull/bidirectional modes
- Frontend integration: React Query, Zustand, tree components

**Speaker Notes**: Search, browse, and sync form the user-facing core of Catalogizer. Search finds media across all sources. Browse provides filesystem-like navigation. Sync extends the system to cloud storage for backup and distribution.

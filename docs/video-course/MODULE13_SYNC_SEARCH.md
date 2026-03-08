# Module 13: Search, Browse & Cloud Sync - Script

**Duration**: 45 minutes
**Module**: 13 - Search, Browse & Cloud Sync

---

## Scene 1: Search API (0:00 - 15:00)

**[Visual: Search architecture diagram showing SearchHandler -> FileRepository -> Database]**

**Narrator**: Welcome to Module 13. Catalogizer exposes a comprehensive Search API that supports full-text search, multi-field filtering, pagination, and duplicate detection. In this module we will explore the search, browse, and cloud sync subsystems end to end.

**[Visual: Show handlers/search.go with SearchHandler struct]**

**Narrator**: The `SearchHandler` in `handlers/search.go` wraps a `FileRepository` and exposes three endpoint groups under `/api/v1/search`:

- `GET /search/files` -- Full-text search with filters
- `GET /search/files/duplicates` -- Duplicate file detection
- `POST /search/advanced` -- Complex multi-field queries via JSON body

**[Code: Show the SearchFiles query parameters]**

```go
// Query parameters for GET /search/files
q           string  // Search query (filename and path)
path        string  // Path filter (partial match)
name        string  // Name filter (partial match)
extension   string  // File extension (exact match)
file_type   string  // File type (exact match)
mime_type   string  // MIME type (exact match)
smb_roots   string  // SMB roots (comma-separated)
min_size    int     // Minimum file size in bytes
max_size    int     // Maximum file size in bytes
modified_after  string  // RFC3339 date
modified_before string  // RFC3339 date
include_deleted bool    // Include soft-deleted files
only_duplicates bool    // Show only duplicates
page        int     // Page number (default: 1)
limit       int     // Items per page (default: 100)
```

**[Visual: Show a search request and response in the browser]**

**Narrator**: Every search response includes pagination metadata: total count, current page, page size, and total pages. The frontend uses React Query to cache results and provide instant re-renders when filters change.

**[Code: Show the AdvancedSearch JSON body]**

```json
{
  "query": "vacation",
  "extensions": ["mp4", "mkv"],
  "file_types": ["video"],
  "min_size": 1048576,
  "max_size": 10737418240,
  "date_range": {
    "after": "2025-01-01T00:00:00Z",
    "before": "2025-12-31T23:59:59Z"
  },
  "sort_by": "size",
  "sort_order": "desc",
  "page": 1,
  "limit": 50
}
```

**[Visual: Show the media search endpoint]**

**Narrator**: In addition to file-level search, the `MediaBrowseHandler` provides `/api/v1/media/search` for searching media entities and `/api/v1/media/stats` for aggregate statistics. Entity search operates on the `media_items` table, returning structured results with hierarchy information -- so searching for a TV show returns the show, its seasons, and its episodes.

**[Demo: Execute a search in the web application, showing filters being applied]**

---

## Scene 2: Browse API (15:00 - 25:00)

**[Visual: Browse API route diagram showing /browse/roots, /browse/directory/*, /browse/file-info/*]**

**Narrator**: The Browse API provides directory navigation across all storage protocols. The `BrowseHandler` in `handlers/browse.go` exposes five endpoints under `/api/v1/browse`:

- `GET /browse/roots` -- List all configured storage roots
- `GET /browse/directory/*path` -- List files in a directory
- `GET /browse/file-info/*path` -- Get metadata for a single file
- `GET /browse/directory-sizes/*path` -- Get subdirectory sizes
- `GET /browse/duplicates/*path` -- Find duplicates within a directory tree

**[Code: Show the GetStorageRoots handler]**

```go
func (h *BrowseHandler) GetStorageRoots(c *gin.Context) {
    ctx := c.Request.Context()
    roots, err := h.fileRepo.GetStorageRoots(ctx)
    if err != nil {
        utils.SendErrorResponse(c, http.StatusInternalServerError,
            "Failed to get storage roots", err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    roots,
    })
}
```

**[Visual: Show directory browsing in the web UI with breadcrumb navigation]**

**Narrator**: Storage roots represent the top-level mount points: an SMB share, an NFS export, a WebDAV collection, or a local directory. The frontend renders them as the root nodes of a file tree. Clicking a root navigates into its directory structure.

**[Visual: Show directory size calculations]**

**Narrator**: The `directory-sizes` endpoint calculates cumulative sizes for each subdirectory. This powers the storage analytics dashboard, showing which directories consume the most space. The calculation is performed on the indexed data, not by querying the remote filesystem, so it is fast even for NAS shares with hundreds of thousands of files.

**[Visual: Show the entity browse endpoint]**

**Narrator**: The Entity API also provides browsing via `GET /api/v1/entities/browse/:type`, which lists all media entities of a given type -- movies, TV shows, music albums -- with pagination and sorting.

**[Demo: Navigate the file tree from a storage root through several directory levels]**

---

## Scene 3: Cloud Sync (25:00 - 45:00)

**[Visual: Sync architecture diagram showing SyncHandler -> SyncService -> SyncRepository]**

**Narrator**: The sync subsystem enables remote synchronization between the Catalogizer catalog and cloud storage providers. It follows the same Handler-Service-Repository pattern as the rest of the backend.

**[Visual: Show the SyncEndpoint model]**

**Narrator**: A `SyncEndpoint` represents a configured cloud destination. It stores the provider type, URL, credentials, sync direction, local and remote paths, and current status.

```go
type SyncEndpoint struct {
    ID            int        `json:"id"`
    UserID        int        `json:"user_id"`
    Name          string     `json:"name"`
    Type          string     `json:"type"`         // s3, gcs, webdav, local
    URL           string     `json:"url"`
    SyncDirection string     `json:"sync_direction"` // push, pull, bidirectional
    LocalPath     string     `json:"local_path"`
    RemotePath    string     `json:"remote_path"`
    Status        string     `json:"status"`
    LastSyncAt    *time.Time `json:"last_sync_at,omitempty"`
}
```

**[Visual: Show the sync API routes]**

**Narrator**: The Sync API under `/api/v1/sync` provides full CRUD for endpoints plus sync execution:

- `POST /sync/endpoints` -- Create a sync endpoint
- `GET /sync/endpoints` -- List user endpoints
- `GET /sync/endpoints/:id` -- Get endpoint details
- `PUT /sync/endpoints/:id` -- Update endpoint configuration
- `DELETE /sync/endpoints/:id` -- Remove an endpoint
- `POST /sync/endpoints/:id/sync` -- Start a sync session
- `GET /sync/sessions` -- List sync sessions
- `GET /sync/sessions/:id` -- Get session details
- `POST /sync/schedules` -- Schedule recurring sync
- `GET /sync/statistics` -- Get aggregate sync statistics
- `POST /sync/cleanup` -- Clean up old session records

**[Code: Show the CreateEndpoint handler]**

```go
func (h *SyncHandler) CreateEndpoint(c *gin.Context) {
    currentUser, err := h.getCurrentUser(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false, "error": "Unauthorized"})
        return
    }
    var endpoint models.SyncEndpoint
    if err := c.ShouldBindJSON(&endpoint); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false, "error": "Invalid request body"})
        return
    }
    created, err := h.syncService.CreateSyncEndpoint(
        currentUser.ID, &endpoint)
    // ... error handling ...
    c.JSON(http.StatusCreated, gin.H{
        "success": true, "data": created})
}
```

**[Visual: Show the SyncSession model with progress tracking fields]**

**Narrator**: Each sync execution produces a `SyncSession` with detailed progress tracking: total files, synced files, failed files, skipped files, duration, and error messages. The frontend polls session status to show a progress bar during sync.

```go
type SyncSession struct {
    ID           int            `json:"id"`
    EndpointID   int            `json:"endpoint_id"`
    Status       string         `json:"status"`
    SyncType     string         `json:"sync_type"`
    TotalFiles   int            `json:"total_files"`
    SyncedFiles  int            `json:"synced_files"`
    FailedFiles  int            `json:"failed_files"`
    SkippedFiles int            `json:"skipped_files"`
    Duration     *time.Duration `json:"duration,omitempty"`
}
```

**[Visual: Show the SyncSchedule model]**

**Narrator**: Scheduled syncs use the `SyncSchedule` model with frequency settings and active/inactive toggling. The scheduler tracks last run time and calculates next run time automatically.

**[Demo: Create a WebDAV sync endpoint, trigger a sync, and observe the session progress]**

**Narrator**: To set up sync in the web application, navigate to Settings, then Sync. Add an endpoint by choosing a provider type, entering connection details, and selecting sync direction. The service validates connectivity before saving. Once configured, you can trigger manual syncs or set up a schedule.

**[Visual: Show statistics endpoint response]**

**Narrator**: The `/sync/statistics` endpoint provides aggregate data: total sessions, success rate, average duration, and total files synced. The cleanup endpoint removes old session records.

---

## Key Code Examples

### Search with Filters
```bash
# Full-text search with extension filter and pagination
curl "http://localhost:8080/api/v1/search/files?q=vacation&extension=mp4&page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

### Browse and Sync
```bash
# List all storage roots
curl http://localhost:8080/api/v1/browse/roots \
  -H "Authorization: Bearer $TOKEN"

# Create a sync endpoint
curl -X POST http://localhost:8080/api/v1/sync/endpoints \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Cloud Backup","type":"webdav","url":"https://cloud.example.com/dav/media","sync_direction":"push"}'
```

---

## Quiz Questions

1. What is the difference between `GET /search/files` and `POST /search/advanced`?
   **Answer**: `GET /search/files` uses query parameters for simple searches. `POST /search/advanced` accepts a JSON body with nested date ranges, multiple extensions, and structured sort options for complex queries.

2. How does the Browse API handle directory sizes without querying the remote filesystem?
   **Answer**: The `directory-sizes` endpoint aggregates stored file sizes by directory path from the database, not from live filesystem queries. This is fast even for NAS shares with hundreds of thousands of files.

3. What happens when a sync endpoint fails connectivity validation during creation?
   **Answer**: The handler returns HTTP 502 Bad Gateway. The endpoint is not saved to the database.

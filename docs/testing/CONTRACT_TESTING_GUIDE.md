# API Contract Testing Guide

Contract tests verify that API responses match the shapes expected by consumers, primarily the TypeScript API client (`catalogizer-api-client`). They protect against unintentional breaking changes in API contracts.

## Location

All contract tests are in `catalog-api/tests/integration/contract_test.go`.

## Running Contract Tests

```bash
cd catalog-api

# Run all contract tests
go test -v -run TestContract ./tests/integration/

# Run a specific contract test
go test -v -run TestContract_HealthResponse ./tests/integration/

# Contract tests are skipped in short mode
go test -short ./tests/integration/   # skips contract tests

# With race detection
go test -race -v -run TestContract ./tests/integration/
```

## How Contract Tests Work

Each contract test:

1. Creates an in-memory SQLite database with the production schema
2. Seeds representative test data
3. Sets up a Gin router with endpoint handlers
4. Makes HTTP requests via `httptest.NewRecorder`
5. Validates the response JSON structure matches the expected contract

### Test Infrastructure

The `contractTestServer()` helper function creates a complete test environment:

```go
func contractTestServer(t *testing.T) (*gin.Engine, *sql.DB) {
    db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=1")
    // ... creates schema, seeds data, sets up routes ...
    return router, db
}
```

### Field Validation

The `assertJSONHasField()` helper validates JSON structure by checking both existence and type:

```go
assertJSONHasField(t, resp, "status", "string")    // field exists and is string
assertJSONHasField(t, resp, "total", "number")      // field exists and is number
assertJSONHasField(t, resp, "enabled", "boolean")   // field exists and is boolean
assertJSONHasField(t, resp, "files", "array")        // field exists and is array
```

## Contract Test Inventory

### TestContract_HealthResponse

Endpoint: `GET /api/v1/health`

Expected shape:

```json
{
  "status": "healthy|unhealthy",
  "timestamp": "2026-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### TestContract_StorageRootsResponse

Endpoint: `GET /api/v1/storage-roots`

Expected shape:

```json
{
  "storage_roots": [
    {
      "id": 1,
      "name": "test-root",
      "protocol": "smb|ftp|nfs|webdav|local",
      "enabled": true
    }
  ],
  "total": 1
}
```

### TestContract_FilesListResponse

Endpoint: `GET /api/v1/catalog/files`

Expected shape:

```json
{
  "files": [
    {
      "id": 1,
      "name": "file.mkv",
      "path": "/media/file.mkv",
      "size": 1048576,
      "is_directory": false,
      "file_type": "video|audio|image|subtitle|document|metadata|archive|other"
    }
  ],
  "total": 5,
  "page": 1,
  "per_page": 50
}
```

### TestContract_EntitiesResponse

Endpoint: `GET /api/v1/entities`

Expected shape:

```json
{
  "entities": [
    {
      "id": 1,
      "title": "Test Movie",
      "media_type": "movie|tv_show|tv_season|tv_episode|music_artist|music_album|song|game|software|book|comic",
      "year": 2025
    }
  ],
  "total": 1
}
```

### TestContract_ScanHistoryResponse

Endpoint: `GET /api/v1/scans`

Expected shape:

```json
{
  "scans": [
    {
      "id": 1,
      "storage_root_id": 1,
      "scan_type": "full",
      "status": "pending|running|completed|failed|cancelled",
      "files_processed": 100
    }
  ],
  "total": 1
}
```

### TestContract_ErrorResponse

Validates all error responses follow `{ "error": "message" }` and never contain stack traces or goroutine info.

### TestContract_PaginationResponse

Validates paginated endpoints include `total` (>= 0), `page` (>= 1), and `per_page` (> 0).

### TestContract_ContentType

Validates all endpoints return `Content-Type: application/json`.

## Writing a New Contract Test

When adding a new API endpoint, add a corresponding contract test:

```go
func TestContract_NewEndpointResponse(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping contract test in short mode")
    }

    router, db := contractTestServer(t)
    defer db.Close()

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/new-endpoint", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var resp map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)

    // Validate the contract
    assertJSONHasField(t, resp, "data", "array")
    assertJSONHasField(t, resp, "total", "number")
}
```

Key rules:

1. Always check `testing.Short()` and skip if true
2. Use `contractTestServer()` for consistent test infrastructure
3. Validate field existence AND type using `assertJSONHasField()`
4. Validate enum values (protocols, media types, statuses) against known lists
5. Add the corresponding route handler to `contractTestServer()` if needed

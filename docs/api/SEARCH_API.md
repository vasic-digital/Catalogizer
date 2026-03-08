# Search API Reference

Base path: `/api/v1`

All search endpoints return JSON with `{ "success": true, "data": ... }` envelope.

## Endpoints

### GET /api/v1/search/files

Full-text file search with filters, pagination, and sorting.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | string | | Search query (matches filename and path) |
| `path` | string | | Path filter (partial match) |
| `name` | string | | Name filter (partial match) |
| `extension` | string | | File extension (exact match, e.g. `mp4`) |
| `file_type` | string | | File type (exact match) |
| `mime_type` | string | | MIME type (exact match) |
| `smb_roots` | string | | Comma-separated storage root names |
| `min_size` | int | | Minimum file size in bytes |
| `max_size` | int | | Maximum file size in bytes |
| `modified_after` | string | | RFC 3339 datetime lower bound |
| `modified_before` | string | | RFC 3339 datetime upper bound |
| `include_deleted` | bool | false | Include soft-deleted files |
| `only_duplicates` | bool | false | Return only duplicate files |
| `exclude_duplicates` | bool | false | Exclude duplicates from results |
| `include_directories` | bool | true | Include directory entries |
| `page` | int | 1 | Page number (min 1) |
| `limit` | int | 100 | Items per page (1-1000, clamped) |
| `sort_by` | string | name | Sort field: `name`, `size`, `modified_at`, `created_at`, `path`, `extension` |
| `sort_order` | string | asc | `asc` or `desc` |

**Response 200:**

```json
{
  "success": true,
  "data": {
    "files": [
      {
        "id": 1,
        "name": "movie.mp4",
        "path": "/media/movies/movie.mp4",
        "size": 1073741824,
        "extension": "mp4",
        "file_type": "video",
        "mime_type": "video/mp4",
        "modified_at": "2026-01-15T10:30:00Z"
      }
    ],
    "total_count": 42,
    "page": 1,
    "limit": 100,
    "total_pages": 1
  }
}
```

**Errors:** 400 (invalid date format), 500 (search failure).

### GET /api/v1/search/files/duplicates

Find duplicate files across storage roots. Automatically sets `only_duplicates=true` and `include_directories=false`.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `smb_roots` | string | | Comma-separated storage root names |
| `min_size` | int | | Minimum file size in bytes |
| `max_size` | int | | Maximum file size in bytes |
| `file_type` | string | | File type filter |
| `extension` | string | | Extension filter |
| `page` | int | 1 | Page number |
| `limit` | int | 100 | Items per page (1-1000) |
| `sort_by` | string | name | `name`, `size`, `modified_at`, `path` |
| `sort_order` | string | asc | `asc` or `desc` |

Response format is identical to `/search/files`.

### POST /api/v1/search/advanced

Complex search using a JSON request body. Supports the same filters as the GET endpoint but accepts them structured in a JSON body.

**Request body:**

```json
{
  "filter": {
    "query": "vacation",
    "extension": "jpg",
    "min_size": 1048576,
    "storage_roots": ["Media", "Photos"],
    "modified_after": "2025-06-01T00:00:00Z"
  },
  "page": 1,
  "limit": 50,
  "sort_by": "size",
  "sort_order": "desc"
}
```

The `filter` object maps to the `SearchFilter` model with the same fields as the GET query parameters.

**Response:** Same paginated `SearchResult` format as GET endpoints.

**Errors:** 400 (invalid JSON body), 500 (search failure).

## Pagination

All search endpoints use offset-based pagination:

- `page` starts at 1 (values < 1 are clamped to 1)
- `limit` is clamped to range 1-1000 (out-of-range values default to 100)
- Response includes `total_count`, `total_pages`, current `page`, and `limit`

## Sorting

Valid sort fields: `name`, `size`, `modified_at`, `created_at`, `path`, `extension`.
Invalid sort fields silently fall back to `name`. Invalid sort orders fall back to `asc`.

Duplicate search supports a subset: `name`, `size`, `modified_at`, `path`.

## Source

- Handler: `catalog-api/handlers/search.go`
- Model: `catalog-api/models/file.go` (SearchFilter, SearchResult, PaginationOptions, SortOptions)
- Repository: `catalog-api/repository/file_repository.go`

# Catalogizer REST API Reference

Complete REST API documentation for the Catalogizer backend (`catalog-api`). All endpoints are served under the base URL `http://localhost:8080`.

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
   - [POST /api/v1/auth/login](#post-apiv1authlogin)
   - [POST /api/v1/auth/register](#post-apiv1authregister)
   - [POST /api/v1/auth/refresh](#post-apiv1authrefresh)
   - [POST /api/v1/auth/logout](#post-apiv1authlogout)
   - [GET /api/v1/auth/me](#get-apiv1authme)
3. [Catalog Browsing](#catalog-browsing)
   - [GET /api/v1/catalog](#get-apiv1catalog)
   - [GET /api/v1/catalog/{path}](#get-apiv1catalogpath)
   - [GET /api/v1/catalog-info/{path}](#get-apiv1catalog-infopath)
4. [Search](#search)
   - [GET /api/v1/search](#get-apiv1search)
   - [GET /api/v1/search/duplicates](#get-apiv1searchduplicates)
5. [Download](#download)
   - [GET /api/v1/download/file/{id}](#get-apiv1downloadfileid)
   - [GET /api/v1/download/directory/{path}](#get-apiv1downloaddirectorypath)
   - [POST /api/v1/download/archive](#post-apiv1downloadarchive)
6. [File Copy Operations](#file-copy-operations)
   - [POST /api/v1/copy/storage](#post-apiv1copystorage)
   - [POST /api/v1/copy/local](#post-apiv1copylocal)
   - [POST /api/v1/copy/upload](#post-apiv1copyupload)
7. [Media Operations](#media-operations)
   - [GET /api/v1/media/{id}](#get-apiv1mediaid)
   - [PUT /api/v1/media/{id}/progress](#put-apiv1mediaidprogress)
   - [PUT /api/v1/media/{id}/favorite](#put-apiv1mediaidfavorite)
8. [Recommendations](#recommendations)
   - [GET /api/v1/recommendations/similar/{media_id}](#get-apiv1recommendationssimilarmedia_id)
   - [GET /api/v1/recommendations/trending](#get-apiv1recommendationstrending)
   - [GET /api/v1/recommendations/personalized/{user_id}](#get-apiv1recommendationspersonalizeduser_id)
   - [GET /api/v1/recommendations/test](#get-apiv1recommendationstest)
9. [Subtitles](#subtitles)
   - [GET /api/v1/subtitles/search](#get-apiv1subtitlessearch)
   - [POST /api/v1/subtitles/download](#post-apiv1subtitlesdownload)
   - [GET /api/v1/subtitles/media/{media_id}](#get-apiv1subtitlesmediamedia_id)
   - [GET /api/v1/subtitles/{subtitle_id}/verify-sync/{media_id}](#get-apiv1subtitlessubtitle_idverify-syncmedia_id)
   - [POST /api/v1/subtitles/translate](#post-apiv1subtitlestranslate)
   - [POST /api/v1/subtitles/upload](#post-apiv1subtitlesupload)
   - [GET /api/v1/subtitles/languages](#get-apiv1subtitleslanguages)
   - [GET /api/v1/subtitles/providers](#get-apiv1subtitlesproviders)
10. [Storage](#storage)
    - [GET /api/v1/storage/roots](#get-apiv1storageroots)
    - [GET /api/v1/storage/list/{path}](#get-apiv1storagelistpath)
11. [Statistics](#statistics)
    - [GET /api/v1/stats/directories/by-size](#get-apiv1statsdirectoriesby-size)
    - [GET /api/v1/stats/duplicates/count](#get-apiv1statsduplicatescount)
    - [GET /api/v1/stats/overall](#get-apiv1statsoverall)
    - [GET /api/v1/stats/smb/{smb_root}](#get-apiv1statssmb_root)
    - [GET /api/v1/stats/filetypes](#get-apiv1statsfiletypes)
    - [GET /api/v1/stats/sizes](#get-apiv1statssizes)
    - [GET /api/v1/stats/duplicates](#get-apiv1statsduplicates)
    - [GET /api/v1/stats/duplicates/groups](#get-apiv1statsduplicatesgroups)
    - [GET /api/v1/stats/access](#get-apiv1statsaccess)
    - [GET /api/v1/stats/growth](#get-apiv1statsgrowth)
    - [GET /api/v1/stats/scans](#get-apiv1statsscans)
12. [SMB Discovery](#smb-discovery)
    - [POST /api/v1/smb/discover](#post-apiv1smbdiscover)
    - [GET /api/v1/smb/discover](#get-apiv1smbdiscover)
    - [POST /api/v1/smb/test](#post-apiv1smbtest)
    - [GET /api/v1/smb/test](#get-apiv1smbtest)
    - [POST /api/v1/smb/browse](#post-apiv1smbbrowse)
13. [Conversion](#conversion)
    - [POST /api/v1/conversion/jobs](#post-apiv1conversionjobs)
    - [GET /api/v1/conversion/jobs](#get-apiv1conversionjobs)
    - [GET /api/v1/conversion/jobs/{id}](#get-apiv1conversionjobsid)
    - [POST /api/v1/conversion/jobs/{id}/cancel](#post-apiv1conversionjobsidcancel)
    - [GET /api/v1/conversion/formats](#get-apiv1conversionformats)
14. [User Management](#user-management)
    - [POST /api/v1/users](#post-apiv1users)
    - [GET /api/v1/users](#get-apiv1users)
    - [GET /api/v1/users/{id}](#get-apiv1usersid)
    - [PUT /api/v1/users/{id}](#put-apiv1usersid)
    - [DELETE /api/v1/users/{id}](#delete-apiv1usersid)
    - [POST /api/v1/users/{id}/reset-password](#post-apiv1usersidreset-password)
    - [POST /api/v1/users/{id}/lock](#post-apiv1usersidlock)
    - [POST /api/v1/users/{id}/unlock](#post-apiv1usersidunlock)
15. [Role Management](#role-management)
    - [POST /api/v1/roles](#post-apiv1roles)
    - [GET /api/v1/roles](#get-apiv1roles)
    - [GET /api/v1/roles/{id}](#get-apiv1rolesid)
    - [PUT /api/v1/roles/{id}](#put-apiv1rolesid)
    - [DELETE /api/v1/roles/{id}](#delete-apiv1rolesid)
    - [GET /api/v1/roles/permissions](#get-apiv1rolespermissions)
16. [Configuration](#configuration)
    - [GET /api/v1/configuration](#get-apiv1configuration)
    - [POST /api/v1/configuration/test](#post-apiv1configurationtest)
    - [GET /api/v1/configuration/status](#get-apiv1configurationstatus)
    - [GET /api/v1/configuration/wizard/step/{step_id}](#get-apiv1configurationwizardstepstep_id)
    - [POST /api/v1/configuration/wizard/step/{step_id}/validate](#post-apiv1configurationwizardstepstep_idvalidate)
    - [POST /api/v1/configuration/wizard/step/{step_id}/save](#post-apiv1configurationwizardstepstep_idsave)
    - [GET /api/v1/configuration/wizard/progress](#get-apiv1configurationwizardprogress)
    - [POST /api/v1/configuration/wizard/complete](#post-apiv1configurationwizardcomplete)
17. [Error Reporting](#error-reporting)
    - [POST /api/v1/errors/report](#post-apiv1errorsreport)
    - [POST /api/v1/errors/crash](#post-apiv1errorscrash)
    - [GET /api/v1/errors/reports](#get-apiv1errorsreports)
    - [GET /api/v1/errors/reports/{id}](#get-apiv1errorsreportsid)
    - [PUT /api/v1/errors/reports/{id}/status](#put-apiv1errorsreportsidstatus)
    - [GET /api/v1/errors/crashes](#get-apiv1errorscrashes)
    - [GET /api/v1/errors/crashes/{id}](#get-apiv1errorscrashesid)
    - [PUT /api/v1/errors/crashes/{id}/status](#put-apiv1errorscrashesidstatus)
    - [GET /api/v1/errors/statistics](#get-apiv1errorsstatistics)
    - [GET /api/v1/errors/crash-statistics](#get-apiv1errorscrash-statistics)
    - [GET /api/v1/errors/health](#get-apiv1errorshealth)
18. [Log Management](#log-management)
    - [POST /api/v1/logs/collect](#post-apiv1logscollect)
    - [GET /api/v1/logs/collections](#get-apiv1logscollections)
    - [GET /api/v1/logs/collections/{id}](#get-apiv1logscollectionsid)
    - [GET /api/v1/logs/collections/{id}/entries](#get-apiv1logscollectionsidentries)
    - [POST /api/v1/logs/collections/{id}/export](#post-apiv1logscollectionsidexport)
    - [GET /api/v1/logs/collections/{id}/analyze](#get-apiv1logscollectionsidanalyze)
    - [POST /api/v1/logs/share](#post-apiv1logsshare)
    - [GET /api/v1/logs/share/{token}](#get-apiv1logssharetoken)
    - [DELETE /api/v1/logs/share/{id}](#delete-apiv1logsshareid)
    - [GET /api/v1/logs/stream](#get-apiv1logsstream)
    - [GET /api/v1/logs/statistics](#get-apiv1logsstatistics)
19. [Health and Metrics](#health-and-metrics)
    - [GET /health](#get-health)
    - [GET /metrics](#get-metrics)
20. [Global Middleware](#global-middleware)
21. [Error Handling](#error-handling)
22. [Rate Limiting](#rate-limiting)

---

## Overview

| Property | Value |
|---|---|
| Base URL | `http://localhost:8080` |
| API Version | `v1` |
| Auth Method | JWT Bearer Token |
| Content Type | `application/json` |
| Database | SQLite (dev) / PostgreSQL (prod) |

All API routes (except `/health`, `/metrics`, and `/api/v1/auth/*`) require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

---

## Authentication

Authentication endpoints are under `/api/v1/auth`. These endpoints have stricter rate limiting (5 requests/minute) and do not require a JWT token (except `/me`).

### POST /api/v1/auth/login

Authenticate a user and receive JWT tokens.

| Property | Value |
|---|---|
| Auth Required | No |
| Rate Limit | 5/min |
| Permission | None |

**Request Body:**

```json
{
  "username": "admin",
  "password": "securepassword",
  "device_info": {
    "device_type": "desktop",
    "platform": "linux",
    "app_version": "3.0.0"
  },
  "remember_me": true
}
```

**Success Response (200):**

```json
{
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "role_id": 1,
    "role": {
      "id": 1,
      "name": "admin",
      "permissions": ["*"]
    },
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "session_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2024-01-02T00:00:00Z"
}
```

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Invalid request format"}` | Malformed JSON |
| 401 | `{"error": "invalid credentials"}` | Wrong username/password |

---

### POST /api/v1/auth/register

Register a new user account.

| Property | Value |
|---|---|
| Auth Required | No |
| Rate Limit | 5/min |
| Permission | None |

**Request Body:**

```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

| Field | Type | Required | Validation |
|---|---|---|---|
| `username` | string | Yes | - |
| `email` | string | Yes | Valid email |
| `password` | string | Yes | Min 8 characters |
| `first_name` | string | Yes | - |
| `last_name` | string | Yes | - |

**Success Response (201):**

Returns the created `User` object (see [API_SCHEMAS.md](./API_SCHEMAS.md#user)).

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "<validation details>"}` | Validation failure |
| 409 | `{"error": "Username already exists"}` | Duplicate username |
| 409 | `{"error": "Email already exists"}` | Duplicate email |

---

### POST /api/v1/auth/refresh

Refresh an expired access token using a refresh token.

| Property | Value |
|---|---|
| Auth Required | No |
| Rate Limit | 5/min |

**Request Body:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Success Response (200):**

```json
{
  "session_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2024-01-02T00:00:00Z"
}
```

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Invalid request format"}` | Malformed JSON |
| 401 | `{"error": "invalid refresh token"}` | Expired or invalid token |

---

### POST /api/v1/auth/logout

Invalidate the current session token.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 5/min |

**Success Response (200):**

```json
{
  "message": "Logged out successfully"
}
```

---

### GET /api/v1/auth/me

Get the currently authenticated user's profile.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 5/min |

**Success Response (200):**

Returns the current `User` object (see [API_SCHEMAS.md](./API_SCHEMAS.md#user)).

---

## Catalog Browsing

Browse the file catalog across all configured storage roots (SMB, FTP, NFS, WebDAV, local).

### GET /api/v1/catalog

List all available storage root directories.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Success Response (200):**

```json
{
  "roots": [
    "nas-media",
    "nas-backup",
    "local-storage"
  ]
}
```

---

### GET /api/v1/catalog/{path}

List files and directories at the specified path.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `sort_by` | string | `name` | Sort field: `name`, `size`, `modified` |
| `sort_order` | string | `asc` | Sort order: `asc`, `desc` |
| `limit` | int | `100` | Max results to return |
| `offset` | int | `0` | Pagination offset |

**Example Request:**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/catalog/nas-media/movies?sort_by=size&sort_order=desc&limit=50"
```

**Success Response (200):**

```json
{
  "files": [
    {
      "id": 1024,
      "storage_root_id": 1,
      "storage_root_name": "nas-media",
      "path": "movies/The Matrix (1999)",
      "name": "The Matrix (1999)",
      "size": 4294967296,
      "is_directory": true,
      "mime_type": null,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 50,
  "offset": 0
}
```

---

### GET /api/v1/catalog-info/{path}

Get detailed information about a specific file or directory.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Success Response (200):**

Returns a `FileInfo` object (see [API_SCHEMAS.md](./API_SCHEMAS.md#fileinfo)).

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Path is required"}` | Empty path |
| 404 | `{"error": "File not found"}` | Path does not exist |

---

## Search

### GET /api/v1/search

Search for files and directories using various criteria.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `query` | string | Yes | - | Search term (filename match) |
| `path` | string | No | - | Path filter |
| `extension` | string | No | - | File extension filter |
| `mime_type` | string | No | - | MIME type filter |
| `min_size` | int | No | - | Minimum file size (bytes) |
| `max_size` | int | No | - | Maximum file size (bytes) |
| `smb_roots` | string | No | - | Comma-separated storage root names |
| `is_directory` | bool | No | - | Filter by directory status |
| `sort_by` | string | No | `name` | Sort field |
| `sort_order` | string | No | `asc` | Sort direction |
| `limit` | int | No | `100` | Max results |
| `offset` | int | No | `0` | Pagination offset |

**Example Request:**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/search?query=matrix&extension=mkv&smb_roots=nas-media"
```

**Success Response (200):**

```json
{
  "files": [...],
  "total": 15,
  "count": 15,
  "limit": 100,
  "offset": 0
}
```

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Search query is required"}` | Missing query parameter |

---

### GET /api/v1/search/duplicates

Find groups of duplicate files within a storage root.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `smb_root` | string | Yes | - | Storage root name to search |
| `min_count` | int | No | `2` | Minimum duplicates per group |
| `limit` | int | No | `50` | Max groups to return |

**Success Response (200):**

```json
{
  "groups": [
    {
      "id": 42,
      "file_count": 3,
      "total_size": 12884901888,
      "created_at": "2024-01-10T08:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

---

## Download

### GET /api/v1/download/file/{id}

Download a single file by its database ID.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |
| Response Type | `application/octet-stream` |

**Path Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `id` | int | File ID |

**Response Headers:**

```
Content-Disposition: attachment; filename="movie.mkv"
Content-Type: application/octet-stream
Content-Length: 4294967296
```

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Invalid file ID"}` | Non-numeric ID |
| 400 | `{"error": "Cannot download directory as single file"}` | Path is a directory |
| 404 | `{"error": "File not found"}` | File does not exist |

---

### GET /api/v1/download/directory/{path}

Download a directory as a compressed archive.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |
| Response Type | `application/zip` or `application/gzip` |

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `format` | string | `zip` | Archive format: `zip`, `tar`, `tar.gz` |

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Invalid format. Supported: zip, tar, tar.gz"}` | Unsupported format |
| 400 | `{"error": "Directory too large for download", "total_size": ..., "max_size": ...}` | Exceeds max size |
| 404 | `{"error": "Directory not found or empty"}` | Path does not exist |

---

### POST /api/v1/download/archive

Create and download an archive from multiple specified file paths.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Request Body:**

```json
{
  "paths": [
    "movies/The Matrix (1999)/The.Matrix.mkv",
    "movies/Inception (2010)/Inception.mkv"
  ],
  "format": "zip",
  "smb_root": "nas-media"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `paths` | string[] | Yes | Array of file paths to include |
| `format` | string | No | `zip` (default), `tar`, `tar.gz` |
| `smb_root` | string | No | Storage root name |

---

## File Copy Operations

### POST /api/v1/copy/storage

Copy a file to a storage location.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Request Body:**

```json
{
  "source_path": "/tmp/upload/document.pdf",
  "dest_path": "/documents/archive/document.pdf",
  "storage_id": "local"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `source_path` | string | Yes | Source file path |
| `dest_path` | string | Yes | Destination file path |
| `storage_id` | string | Yes | Target storage root ID |

**Success Response (200):**

```json
{
  "message": "File copied to storage successfully",
  "source": "/tmp/upload/document.pdf",
  "destination": "/documents/archive/document.pdf",
  "storage_id": "local"
}
```

---

### POST /api/v1/copy/local

Copy a file from a remote storage (SMB) to local filesystem.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Request Body:**

```json
{
  "source_path": "nas-media:movies/movie.mkv",
  "destination_path": "/local/downloads/movie.mkv",
  "overwrite": false
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `source_path` | string | Yes | Source in `host:path` format |
| `destination_path` | string | Yes | Local destination path |
| `overwrite` | bool | No | Overwrite existing files (default: false) |

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Invalid source format. Use 'host:path'"}` | Bad source format |
| 409 | `{"error": "Destination file already exists"}` | File exists and overwrite=false |

---

### POST /api/v1/copy/upload

Upload a file from local filesystem to SMB storage.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |
| Content-Type | `multipart/form-data` |

**Form Fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `file` | file | Yes | File to upload |
| `destination` | string | Yes | Destination in `host:path` format |
| `overwrite` | string | No | `"true"` to overwrite existing files |

**Success Response (200):**

```json
{
  "message": "File uploaded successfully",
  "filename": "document.pdf",
  "destination": "nas-media:documents/document.pdf",
  "size": 1048576
}
```

---

## Media Operations

### GET /api/v1/media/{id}

Get detailed media item information by ID.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Success Response (200):**

```json
{
  "id": 42,
  "title": "The Matrix (1999)",
  "media_type": "video",
  "year": 1999,
  "description": "A computer hacker learns about the true nature of reality.",
  "cover_image": "https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg",
  "rating": 8.7,
  "quality": "1080p",
  "file_size": 4294967296,
  "duration": 8160,
  "directory_path": "/movies/The Matrix (1999)",
  "smb_path": "//nas/media/movies/The Matrix (1999)",
  "created_at": "2024-01-15 10:30:00",
  "updated_at": "2024-01-15 10:30:00",
  "external_metadata": [],
  "versions": [],
  "is_favorite": true,
  "watch_progress": 0.75,
  "last_watched": "2024-01-20 20:00:00",
  "is_downloaded": false
}
```

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Invalid media ID"}` | Non-numeric ID |
| 404 | `{"error": "Media not found"}` | Media does not exist |

---

### PUT /api/v1/media/{id}/progress

Update watch progress for a media item.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Request Body:**

```json
{
  "progress": 0.75
}
```

| Field | Type | Required | Validation |
|---|---|---|---|
| `progress` | float | Yes | 0.0 to 1.0 |

**Success Response (200):**

```json
{
  "success": true,
  "message": "Watch progress updated successfully"
}
```

---

### PUT /api/v1/media/{id}/favorite

Update the favorite status for a media item.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Request Body:**

```json
{
  "favorite": true
}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Favorite status updated successfully"
}
```

---

## Recommendations

### GET /api/v1/recommendations/similar/{media_id}

Get media items similar to a given media item.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Path Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `media_id` | int | Media item ID |

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `max_local_items` | int | `10` | Max local results |
| `max_external_items` | int | `5` | Max external results |
| `include_external` | bool | `false` | Include external API results |
| `similarity_threshold` | float | `0.3` | Minimum similarity score |

**Success Response (200):**

```json
{
  "media_id": "42",
  "local_items": [...],
  "external_items": [...],
  "total_local": 8,
  "total_external": 5
}
```

---

### GET /api/v1/recommendations/trending

Get trending media items based on recent activity.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `media_type` | string | - | Filter by type: `video`, `audio`, etc. |
| `limit` | int | `20` | Max results |
| `time_range` | string | `week` | Time range: `day`, `week`, `month`, `year` |

**Success Response (200):**

```json
{
  "items": [
    {
      "id": 100,
      "title": "Trending Movie 1",
      "media_type": "video",
      "rating": 8.2,
      "is_favorite": false,
      "watch_progress": 0,
      "is_downloaded": true
    }
  ],
  "media_type": "video",
  "time_range": "week",
  "generated_at": "2024-01-20T12:00:00Z"
}
```

---

### GET /api/v1/recommendations/personalized/{user_id}

Get personalized recommendations based on viewing history.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `limit` | int | `20` | Max results |

**Success Response (200):**

```json
{
  "user_id": 1,
  "items": [...],
  "generated_at": "2024-01-20T12:00:00Z"
}
```

---

### GET /api/v1/recommendations/test

Simple test endpoint for the recommendation system.

**Success Response (200):**

```json
{
  "message": "Simple recommendation works!"
}
```

---

## Subtitles

### GET /api/v1/subtitles/search

Search for subtitles across multiple providers.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Query Parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `media_path` | string | Yes | Path to the media file |
| `title` | string | No | Media title override |
| `year` | int | No | Release year |
| `season` | int | No | TV season number |
| `episode` | int | No | TV episode number |
| `languages` | string | No | Comma-separated language codes |
| `providers` | string | No | Comma-separated provider names |

**Available Providers:** `opensubtitles`, `subdb`, `yifysubtitles`, `subscene`, `addic7ed`

**Success Response (200):**

```json
{
  "success": true,
  "results": [
    {
      "id": "os-12345",
      "title": "The Matrix",
      "language": "English",
      "language_code": "en",
      "provider": "opensubtitles",
      "format": "srt",
      "rating": 9.2,
      "download_count": 150000
    }
  ],
  "count": 1
}
```

---

### POST /api/v1/subtitles/download

Download a specific subtitle by result ID.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Rate Limit | 100/min |

**Request Body:**

```json
{
  "media_item_id": 42,
  "result_id": "os-12345",
  "language": "en"
}
```

**Success Response (200):**

```json
{
  "success": true,
  "track": {
    "id": "sub-001",
    "language": "English",
    "language_code": "en",
    "format": "srt",
    "is_default": false,
    "is_forced": false
  }
}
```

---

### GET /api/v1/subtitles/media/{media_id}

Get all subtitle tracks for a media item.

**Success Response (200):**

```json
{
  "success": true,
  "subtitles": [
    {
      "id": "sub-001",
      "language": "English",
      "language_code": "en",
      "format": "srt"
    }
  ],
  "media_item_id": 42
}
```

---

### GET /api/v1/subtitles/{subtitle_id}/verify-sync/{media_id}

Verify if a subtitle is properly synchronized with its media.

**Success Response (200):**

```json
{
  "success": true,
  "sync_result": {
    "is_synced": true,
    "offset_ms": 150,
    "confidence": 0.95
  }
}
```

---

### POST /api/v1/subtitles/translate

Translate a subtitle to another language.

**Request Body:**

```json
{
  "subtitle_id": "sub-001",
  "source_language": "en",
  "target_language": "es"
}
```

**Success Response (200):**

```json
{
  "success": true,
  "translated_track": {
    "id": "sub-002",
    "language": "Spanish",
    "language_code": "es",
    "format": "srt"
  }
}
```

---

### POST /api/v1/subtitles/upload

Upload a subtitle file for a media item.

| Property | Value |
|---|---|
| Content-Type | `multipart/form-data` |

**Form Fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `media_item_id` | int | Yes | Media item ID |
| `language` | string | Yes | Language name (e.g., "English") |
| `language_code` | string | Yes | ISO 639-1 code (e.g., "en") |
| `file` | file | Yes | Subtitle file (.srt, .vtt, .ass, .txt) |

---

### GET /api/v1/subtitles/languages

Get the list of supported subtitle languages.

**Success Response (200):**

```json
{
  "success": true,
  "languages": [
    {"code": "en", "name": "English", "native_name": "English"},
    {"code": "es", "name": "Spanish", "native_name": "Espanol"},
    {"code": "fr", "name": "French", "native_name": "Francais"},
    {"code": "de", "name": "German", "native_name": "Deutsch"},
    {"code": "ru", "name": "Russian", "native_name": "Russkij"}
  ],
  "count": 19
}
```

---

### GET /api/v1/subtitles/providers

Get the list of supported subtitle providers.

**Success Response (200):**

```json
{
  "success": true,
  "providers": [
    {"provider": "opensubtitles", "name": "OpenSubtitles", "description": "Large subtitle database with multiple languages", "supported": true},
    {"provider": "subdb", "name": "SubDB", "description": "Hash-based subtitle matching", "supported": true},
    {"provider": "yifysubtitles", "name": "YIFY Subtitles", "description": "Subtitles for YIFY movie releases", "supported": true},
    {"provider": "subscene", "name": "Subscene", "description": "Community-driven subtitle site", "supported": true},
    {"provider": "addic7ed", "name": "Addic7ed", "description": "TV show subtitles with translations", "supported": true}
  ],
  "count": 5
}
```

---

## Storage

### GET /api/v1/storage/roots

Get all available storage root configurations.

**Success Response (200):**

```json
{
  "roots": [
    {"id": "local", "name": "Local Storage", "path": "/data/storage"},
    {"id": "smb", "name": "SMB Storage", "path": "smb://server/share"}
  ]
}
```

---

### GET /api/v1/storage/list/{path}

List files in a storage path.

**Query Parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `storage_id` | string | Yes | Storage root ID |

**Success Response (200):**

```json
{
  "path": "/documents",
  "storage_id": "local",
  "files": [...]
}
```

---

## Statistics

### GET /api/v1/stats/directories/by-size

Get directories sorted by total size.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `smb_root` | string | Yes | - | Storage root name |
| `limit` | int | No | `50` | Max results |

**Success Response (200):**

```json
{
  "directories": [
    {
      "path": "/movies/4K",
      "name": "4K",
      "storage_root_name": "nas-media",
      "file_count": 150,
      "total_size": 1099511627776
    }
  ],
  "count": 10
}
```

---

### GET /api/v1/stats/duplicates/count

Get statistics about duplicate files.

**Query Parameters:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `smb_root` | string | No | Storage root name |

**Success Response (200):**

```json
{
  "duplicate_groups": 42,
  "total_duplicates": 125,
  "total_wasted_space": 53687091200,
  "smb_root": "nas-media"
}
```

---

### GET /api/v1/stats/overall

Get comprehensive catalog statistics.

**Success Response (200):**

```json
{
  "success": true,
  "data": {
    "total_files": 25000,
    "total_directories": 3200,
    "total_size": 5497558138880,
    "total_duplicates": 125,
    "duplicate_groups": 42,
    "storage_roots_count": 5,
    "active_storage_roots": 4,
    "last_scan_time": 1705312200
  }
}
```

---

### GET /api/v1/stats/smb/{smb_root}

Get statistics for a specific storage root.

**Success Response (200):**

```json
{
  "success": true,
  "data": {
    "name": "nas-media",
    "total_files": 15000,
    "total_directories": 2000,
    "total_size": 3298534883328,
    "duplicate_files": 80,
    "duplicate_groups": 25,
    "last_scan_time": 1705312200,
    "is_online": true
  }
}
```

---

### GET /api/v1/stats/filetypes

Get file type distribution statistics.

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `smb_root` | string | - | Storage root filter |
| `limit` | int | `50` | Max results (max 1000) |

**Success Response (200):**

```json
{
  "success": true,
  "data": [
    {"file_type": "video", "extension": ".mkv", "count": 5000, "total_size": 2748779069440, "average_size": 549755813},
    {"file_type": "audio", "extension": ".flac", "count": 12000, "total_size": 549755813888, "average_size": 45812984}
  ]
}
```

---

### GET /api/v1/stats/sizes

Get file size distribution.

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `smb_root` | string | Storage root filter |

**Success Response (200):**

```json
{
  "success": true,
  "data": {
    "tiny": 500,
    "small": 3000,
    "medium": 8000,
    "large": 10000,
    "huge": 3000,
    "massive": 500
  }
}
```

Size buckets: tiny (<1KB), small (1KB-1MB), medium (1MB-10MB), large (10MB-100MB), huge (100MB-1GB), massive (>1GB).

---

### GET /api/v1/stats/duplicates

Get duplicate file statistics.

**Success Response (200):**

```json
{
  "success": true,
  "data": {
    "total_duplicates": 125,
    "duplicate_groups": 42,
    "wasted_space": 53687091200,
    "largest_duplicate_group": 8,
    "average_group_size": 2.97
  }
}
```

---

### GET /api/v1/stats/duplicates/groups

Get the largest duplicate groups.

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `sort_by` | string | `count` | `count` or `size` |
| `limit` | int | `20` | Max results (max 100) |
| `smb_root` | string | - | Storage root filter |

---

### GET /api/v1/stats/access

Get file access pattern statistics.

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `smb_root` | string | - | Storage root filter |
| `days` | int | `30` | Analysis period (max 365) |

---

### GET /api/v1/stats/growth

Get storage growth trends over time.

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `smb_root` | string | - | Storage root filter |
| `months` | int | `12` | Analysis period (max 60) |

**Success Response (200):**

```json
{
  "success": true,
  "data": {
    "monthly_growth": [
      {"month": "2024-01", "files_added": 500, "size_added": 274877906944, "total_files": 25000, "total_size": 5497558138880}
    ],
    "total_growth_rate": 12.5,
    "file_growth_rate": 8.3,
    "size_growth_rate": 15.2
  }
}
```

---

### GET /api/v1/stats/scans

Get scan operation history.

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `smb_root` | string | - | Storage root filter |
| `limit` | int | `50` | Max results (max 1000) |
| `offset` | int | `0` | Pagination offset |

**Success Response (200):**

```json
{
  "success": true,
  "data": {
    "scans": [
      {
        "id": 23,
        "storage_root_id": 1,
        "scan_type": "full",
        "status": "completed",
        "start_time": "2024-01-15T10:00:00Z",
        "end_time": "2024-01-15T10:45:00Z",
        "files_processed": 25000,
        "files_added": 150,
        "files_updated": 30,
        "files_deleted": 5,
        "error_count": 0
      }
    ],
    "total_count": 100,
    "limit": 50,
    "offset": 0
  }
}
```

---

## SMB Discovery

### POST /api/v1/smb/discover

Discover available SMB shares on a host.

**Request Body:**

```json
{
  "host": "192.168.1.100",
  "username": "user",
  "password": "password",
  "domain": "WORKGROUP"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `host` | string | Yes | SMB host address |
| `username` | string | Yes | Authentication username |
| `password` | string | Yes | Authentication password |
| `domain` | string | No | Windows domain |

**Success Response (200):**

Returns an array of `SMBShareInfo` objects.

---

### GET /api/v1/smb/discover

Discover SMB shares using query parameters (for testing).

**Query Parameters:**

| Parameter | Type | Required |
|---|---|---|
| `host` | string | Yes |
| `username` | string | Yes |
| `password` | string | Yes |
| `domain` | string | No |

---

### POST /api/v1/smb/test

Test connectivity to an SMB share.

**Request Body:**

```json
{
  "host": "192.168.1.100",
  "port": 445,
  "share": "media",
  "username": "user",
  "password": "password",
  "domain": "WORKGROUP"
}
```

**Success Response (200):**

```json
{
  "success": true,
  "host": "192.168.1.100",
  "share": "media",
  "username": "user",
  "connection": true
}
```

---

### GET /api/v1/smb/test

Test SMB connection using query parameters.

**Query Parameters:**

| Parameter | Type | Required | Default |
|---|---|---|---|
| `host` | string | Yes | - |
| `share` | string | Yes | - |
| `username` | string | Yes | - |
| `password` | string | Yes | - |
| `domain` | string | No | - |
| `port` | int | No | `445` |

---

### POST /api/v1/smb/browse

Browse files and directories in an SMB share.

**Request Body:**

```json
{
  "host": "192.168.1.100",
  "port": 445,
  "share": "media",
  "username": "user",
  "password": "password",
  "domain": "WORKGROUP",
  "path": "movies"
}
```

**Success Response (200):**

Returns an array of `SMBFileEntry` objects.

---

## Conversion

### POST /api/v1/conversion/jobs

Create a new media format conversion job.

| Property | Value |
|---|---|
| Auth Required | Bearer Token |
| Permission | `conversion.create` |

**Request Body:**

```json
{
  "source_path": "/media/video.mp4",
  "target_path": "/media/video.mp3",
  "source_format": "mp4",
  "target_format": "mp3",
  "conversion_type": "audio",
  "quality": "high",
  "priority": 1,
  "settings": "{\"bitrate\": \"320k\"}",
  "scheduled_for": "2024-01-20T10:00:00Z"
}
```

**Success Response (200):**

Returns a `ConversionJob` object (see [API_SCHEMAS.md](./API_SCHEMAS.md#conversionjob)).

---

### GET /api/v1/conversion/jobs

List conversion jobs for the current user.

| Property | Value |
|---|---|
| Permission | `conversion.view` |

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `status` | string | - | Filter by status: `pending`, `running`, `completed`, `failed`, `cancelled` |
| `limit` | int | `50` | Max results (max 100) |
| `offset` | int | `0` | Pagination offset |

**Success Response (200):**

Returns an array of `ConversionJob` objects.

---

### GET /api/v1/conversion/jobs/{id}

Get a specific conversion job by ID.

**Error Responses:**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "Invalid job ID"}` | Non-numeric ID |
| 404 | `{"error": "Job not found"}` | Job does not exist |

---

### POST /api/v1/conversion/jobs/{id}/cancel

Cancel a running conversion job.

| Property | Value |
|---|---|
| Permission | `conversion.manage` |

**Success Response (200):**

```json
{
  "message": "Job cancelled successfully"
}
```

---

### GET /api/v1/conversion/formats

Get all supported conversion formats.

| Property | Value |
|---|---|
| Permission | `conversion.view` |

**Success Response (200):**

```json
{
  "video": {
    "input": ["mp4", "mkv", "avi", "mov", "wmv", "flv", "webm"],
    "output": ["mp4", "mkv", "avi", "webm"]
  },
  "audio": {
    "input": ["mp3", "flac", "wav", "aac", "ogg", "wma", "m4a"],
    "output": ["mp3", "flac", "wav", "aac", "ogg"]
  },
  "document": {
    "input": ["pdf", "doc", "docx", "txt", "rtf"],
    "output": ["pdf", "txt"]
  },
  "image": {
    "input": ["jpg", "png", "gif", "bmp", "webp", "tiff"],
    "output": ["jpg", "png", "webp"]
  }
}
```

---

## User Management

All user management endpoints require a JWT token and appropriate permissions.

### POST /api/v1/users

Create a new user.

| Property | Value |
|---|---|
| Permission | `user.create` |

**Request Body:**

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword123",
  "role_id": 2,
  "first_name": "John",
  "last_name": "Doe",
  "display_name": "John D.",
  "time_zone": "America/New_York",
  "language": "en",
  "is_active": true
}
```

**Success Response (201):**

Returns the created `User` object.

**Error Responses:**

| Status | Condition |
|---|---|
| 400 | Password validation failure |
| 409 | Username or email already exists |

---

### GET /api/v1/users

List all users with pagination.

| Property | Value |
|---|---|
| Permission | `user.view` |

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `limit` | int | `50` | Max results (max 100) |
| `offset` | int | `0` | Pagination offset |

**Success Response (200):**

```json
{
  "users": [...],
  "total_count": 150,
  "limit": 50,
  "offset": 0
}
```

---

### GET /api/v1/users/{id}

Get a specific user by ID. Users can view their own profile; viewing other users requires `user.view` permission.

---

### PUT /api/v1/users/{id}

Update a user's information. Users can update their own profile; updating other users requires `user.update` permission. Changing `role_id` or `is_active` requires `user.manage` permission.

**Request Body (all fields optional):**

```json
{
  "username": "newusername",
  "email": "newemail@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "display_name": "John D.",
  "avatar_url": "https://example.com/avatar.jpg",
  "time_zone": "America/New_York",
  "language": "en",
  "role_id": 2,
  "is_active": true,
  "settings": {"theme": "dark"}
}
```

---

### DELETE /api/v1/users/{id}

Delete a user account.

| Property | Value |
|---|---|
| Permission | `user.delete` |

**Success Response:** `204 No Content`

**Error Responses:**

| Status | Condition |
|---|---|
| 400 | Attempting to delete own account |

---

### POST /api/v1/users/{id}/reset-password

Reset a user's password (admin operation).

| Property | Value |
|---|---|
| Permission | `user.manage` |

**Request Body:**

```json
{
  "new_password": "newSecurePassword123"
}
```

---

### POST /api/v1/users/{id}/lock

Lock a user account until a specified time.

| Property | Value |
|---|---|
| Permission | `user.manage` |

**Request Body:**

```json
{
  "lock_until": "2024-02-01T00:00:00Z"
}
```

---

### POST /api/v1/users/{id}/unlock

Unlock a locked user account.

| Property | Value |
|---|---|
| Permission | `user.manage` |

**Success Response (200):**

```json
{
  "message": "Account unlocked successfully"
}
```

---

## Role Management

All role management endpoints require `system.admin` permission.

### POST /api/v1/roles

Create a new role.

**Request Body:**

```json
{
  "name": "editor",
  "description": "Content editor with media management access",
  "permissions": ["media.view", "media.edit", "media.upload"]
}
```

**Success Response (201):**

Returns the created `Role` object.

---

### GET /api/v1/roles

List all roles.

**Success Response (200):**

Returns an array of `Role` objects.

---

### GET /api/v1/roles/{id}

Get a specific role by ID.

---

### PUT /api/v1/roles/{id}

Update a role. System roles cannot be modified.

**Request Body:**

```json
{
  "name": "editor",
  "description": "Updated description",
  "permissions": ["media.view", "media.edit", "media.upload", "media.delete"]
}
```

---

### DELETE /api/v1/roles/{id}

Delete a role. System roles and roles assigned to users cannot be deleted.

**Success Response:** `204 No Content`

---

### GET /api/v1/roles/permissions

Get the complete permission catalog organized by category.

**Success Response (200):**

```json
{
  "user_management": {
    "create_user": "user.create",
    "view_user": "user.view",
    "edit_user": "user.update",
    "delete_user": "user.delete",
    "manage_users": "user.manage"
  },
  "media_management": {
    "view_media": "media.view",
    "upload_media": "media.upload",
    "edit_media": "media.edit",
    "delete_media": "media.delete"
  },
  "share_management": {
    "view_shares": "share.view",
    "create_shares": "share.create",
    "edit_shares": "share.edit",
    "delete_shares": "share.delete"
  },
  "system": {
    "system_admin": "system.admin",
    "view_analytics": "analytics.view",
    "export_data": "analytics.export",
    "manage_settings": "system.configure"
  }
}
```

---

## Configuration

### GET /api/v1/configuration

Get the current system configuration schema.

| Property | Value |
|---|---|
| Permission | `system.configure` |

---

### POST /api/v1/configuration/test

Test a configuration without applying it.

| Property | Value |
|---|---|
| Permission | `system.admin` |

**Request Body:** A `Configuration` object.

**Success Response (200):**

```json
{
  "is_valid": true,
  "errors": [],
  "warnings": ["SMTP not configured"]
}
```

---

### GET /api/v1/configuration/status

Get system component health status.

| Property | Value |
|---|---|
| Permission | `system.configure` |

**Success Response (200):**

```json
{
  "status": "healthy",
  "version": "3.0.0",
  "uptime": "24h 30m",
  "components": {
    "database": "healthy",
    "storage": "healthy",
    "authentication": "healthy",
    "media_conversion": "healthy",
    "sync": "healthy"
  }
}
```

---

### GET /api/v1/configuration/wizard/step/{step_id}

Get a specific setup wizard step definition.

### POST /api/v1/configuration/wizard/step/{step_id}/validate

Validate data for a specific wizard step.

### POST /api/v1/configuration/wizard/step/{step_id}/save

Save progress for a specific wizard step.

### GET /api/v1/configuration/wizard/progress

Get the current wizard progress for the authenticated user.

### POST /api/v1/configuration/wizard/complete

Complete the setup wizard and apply the configuration.

---

## Error Reporting

### POST /api/v1/errors/report

Submit an error report.

| Property | Value |
|---|---|
| Permission | `report.create` |

**Request Body:**

```json
{
  "level": "error",
  "message": "Failed to process media file",
  "error_code": "MEDIA_PROCESS_ERROR",
  "component": "media_processor",
  "stack_trace": "goroutine 1 [running]:...",
  "context": {"file_id": 123, "operation": "thumbnail_generation"},
  "user_agent": "Mozilla/5.0...",
  "url": "/api/v1/media/123/thumbnail"
}
```

**Success Response (200):**

Returns the created `ErrorReport` object.

---

### POST /api/v1/errors/crash

Submit a crash report.

| Property | Value |
|---|---|
| Permission | `report.create` |

**Request Body:**

```json
{
  "signal": "SIGSEGV",
  "message": "Segmentation fault in media decoder",
  "stack_trace": "...",
  "context": {"media_id": 42}
}
```

---

### GET /api/v1/errors/reports

List error reports with filtering.

| Property | Value |
|---|---|
| Permission | `report.view` |

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `level` | string | Filter by level: `debug`, `info`, `warning`, `error`, `fatal` |
| `component` | string | Filter by component |
| `status` | string | Filter by status: `new`, `in_progress`, `resolved`, `ignored` |
| `start_date` | string | Start date (YYYY-MM-DD) |
| `end_date` | string | End date (YYYY-MM-DD) |
| `limit` | int | Max results |
| `offset` | int | Pagination offset |

---

### GET /api/v1/errors/reports/{id}

Get a specific error report.

### PUT /api/v1/errors/reports/{id}/status

Update error report status.

**Request Body:**

```json
{
  "status": "resolved"
}
```

### GET /api/v1/errors/crashes

List crash reports with filtering (same parameters as error reports, with `signal` instead of `level`/`component`).

### GET /api/v1/errors/crashes/{id}

Get a specific crash report.

### PUT /api/v1/errors/crashes/{id}/status

Update crash report status.

### GET /api/v1/errors/statistics

Get error reporting statistics.

**Success Response (200):**

```json
{
  "total_errors": 150,
  "errors_by_level": {"error": 100, "warning": 40, "fatal": 10},
  "errors_by_component": {"media_processor": 80, "auth": 30, "storage": 40},
  "recent_errors": 15,
  "resolved_errors": 120,
  "avg_resolution_time": 3600.5
}
```

### GET /api/v1/errors/crash-statistics

Get crash reporting statistics.

### GET /api/v1/errors/health

Get system health based on error and crash data.

| Property | Value |
|---|---|
| Permission | `system.admin` |

---

## Log Management

All log management endpoints require `system.admin` permission.

### POST /api/v1/logs/collect

Create a new log collection.

**Request Body:**

```json
{
  "name": "Debug Session 2024-01-20",
  "description": "Investigating media processing issue",
  "components": ["api", "media_processor", "storage"],
  "log_level": "debug",
  "start_time": "2024-01-20T00:00:00Z",
  "end_time": "2024-01-20T23:59:59Z",
  "filters": {"include_stack_traces": true}
}
```

---

### GET /api/v1/logs/collections

List log collections.

**Query Parameters:**

| Parameter | Type | Default |
|---|---|---|
| `limit` | int | `20` |
| `offset` | int | `0` |

---

### GET /api/v1/logs/collections/{id}

Get a specific log collection.

### GET /api/v1/logs/collections/{id}/entries

Get log entries for a collection.

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `level` | string | Filter by level |
| `component` | string | Filter by component |
| `search` | string | Full-text search |
| `start_time` | string | ISO 8601 start time |
| `end_time` | string | ISO 8601 end time |
| `limit` | int | Max results |
| `offset` | int | Pagination offset |

### POST /api/v1/logs/collections/{id}/export

Export log collection data.

**Query Parameters:**

| Parameter | Type | Default | Options |
|---|---|---|---|
| `format` | string | `json` | `json`, `csv`, `txt`, `zip` |

### GET /api/v1/logs/collections/{id}/analyze

Analyze a log collection for patterns and insights.

**Success Response (200):**

```json
{
  "collection_id": 5,
  "total_entries": 5000,
  "entries_by_level": {"error": 100, "warning": 500, "info": 3000, "debug": 1400},
  "entries_by_component": {"api": 2000, "storage": 1500, "media_processor": 1500},
  "error_patterns": {"connection_timeout": 45, "file_not_found": 30},
  "time_range": {"start": "2024-01-20T00:00:00Z", "end": "2024-01-20T23:59:59Z"},
  "insights": ["Error rate increased 40% between 14:00-16:00", "Storage component shows timeout pattern"]
}
```

### POST /api/v1/logs/share

Create a shareable link for a log collection.

**Request Body:**

```json
{
  "collection_id": 5,
  "share_type": "private",
  "expires_at": "2024-02-01T00:00:00Z",
  "permissions": ["read"],
  "recipients": ["dev@example.com"]
}
```

### GET /api/v1/logs/share/{token}

Access a shared log collection via share token.

### DELETE /api/v1/logs/share/{id}

Revoke a log share.

### GET /api/v1/logs/stream

Stream live logs via Server-Sent Events (SSE).

| Property | Value |
|---|---|
| Content-Type | `text/event-stream` |

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `level` | string | Filter by level |
| `component` | string | Filter by component |
| `search` | string | Search term |

**SSE Format:**

```
data: {"id":1,"timestamp":"2024-01-20T10:00:00Z","level":"error","component":"api","message":"Request timeout","context":{}}

data: {"id":2,"timestamp":"2024-01-20T10:00:01Z","level":"info","component":"storage","message":"File scan completed","context":{"files":500}}
```

### GET /api/v1/logs/statistics

Get log management statistics.

**Success Response (200):**

```json
{
  "total_collections": 25,
  "total_entries": 150000,
  "active_shares": 3,
  "collections_by_status": {"completed": 20, "in_progress": 3, "failed": 2},
  "recent_collections": 5
}
```

---

## Health and Metrics

### GET /health

Simple health check endpoint. No authentication required.

**Success Response (200):**

```json
{
  "status": "healthy",
  "time": "2024-01-20T12:00:00Z"
}
```

---

### GET /metrics

Prometheus metrics endpoint. Returns metrics in Prometheus exposition format. No authentication required.

Tracked metrics include HTTP request durations, request counts, response sizes, active goroutines, and memory usage.

---

## Global Middleware

All requests pass through these middleware layers:

| Middleware | Description |
|---|---|
| CORS | Cross-Origin Resource Sharing headers |
| Prometheus Metrics | Request duration and count tracking |
| Logger | Structured request logging (zap) |
| Error Handler | Consistent error response formatting |
| Request ID | Unique `X-Request-ID` header per request |
| Input Validation | Request body sanitization and validation |
| JWT Auth | Token validation on `/api/v1/*` routes (except auth) |
| Rate Limiting | Per-user request throttling |

---

## Error Handling

All error responses follow a consistent format:

```json
{
  "error": "Human-readable error message"
}
```

Or the structured format used by subtitle and recommendation handlers:

```json
{
  "success": false,
  "error": "Human-readable error message",
  "code": "MACHINE_READABLE_CODE"
}
```

### HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | Success |
| 201 | Created (user registration, resource creation) |
| 204 | No Content (successful deletion) |
| 400 | Bad Request (validation error, invalid parameters) |
| 401 | Unauthorized (missing or invalid JWT) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found (resource does not exist) |
| 409 | Conflict (duplicate resource) |
| 500 | Internal Server Error |

---

## Rate Limiting

Rate limiting is applied per-user based on the JWT token.

| Endpoint Group | Limit |
|---|---|
| `/api/v1/auth/*` | 5 requests/minute |
| All other `/api/v1/*` | 100 requests/minute |

When rate limited, the server returns `429 Too Many Requests`.

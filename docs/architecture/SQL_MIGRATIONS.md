# SQL Migration Documentation

This document provides a comprehensive reference for all database migrations in the Catalogizer project, including their purpose, the tables they create or modify, dialect differences, and a complete schema reference.

## Table of Contents

- [Overview](#overview)
- [Dialect Abstraction](#dialect-abstraction)
- [Migration Versions](#migration-versions)
- [Complete Schema Reference](#complete-schema-reference)
- [Index Definitions](#index-definitions)
- [Foreign Key Relationships](#foreign-key-relationships)
- [Column Descriptions](#column-descriptions)
- [Sample Queries](#sample-queries)
- [How Migrations Are Applied](#how-migrations-are-applied)
- [How to Create a New Migration](#how-to-create-a-new-migration)
- [Troubleshooting](#troubleshooting)
- [Additional Schema Files](#additional-schema-files)
- [Related Documentation](#related-documentation)

## Overview

Catalogizer uses a **dual-dialect migration system** supporting SQLite (development) and PostgreSQL (production). Migrations are Go functions in the `catalog-api/database/` package that execute dialect-specific DDL at application startup. Each migration is registered with a version number and is applied exactly once, tracked by a `migrations` table.

**Current schema version: 13** (`create_playlist_tables`)

The system consists of:

1. **Go-based programmatic migrations** (primary) -- Defined across multiple files in `catalog-api/database/`. Each migration version has separate SQLite and PostgreSQL implementations dispatched by the dialect. Migrations run automatically on application startup via `RunMigrations()`.

2. **SQL file migrations** (reference/CLI) -- Stored in `catalog-api/database/migrations/` as `.up.sql` and `.down.sql` files. These can be applied manually using the `golang-migrate` CLI tool and serve as a reference for the first few programmatic migrations.

### Migration Source Files

| File | Contents |
|------|----------|
| `database/migrations.go` | `RunMigrations()`, `Migration` struct, dialect dispatch functions for v1-v8 |
| `database/migrations_sqlite.go` | SQLite implementations for v1-v8 |
| `database/migrations_postgres.go` | PostgreSQL implementations for v1-v8 |
| `database/migrations_v9_performance.go` | Performance indexes (v9) -- both dialects |
| `database/migrations_v10_sync_tables.go` | Sync tables (v10) -- both dialects |
| `database/migrations_v11_service_tables.go` | Service tables (v11), column fixes (v12) -- both dialects |
| `database/migrations_v13_playlist_tables.go` | Playlist tables (v13) -- both dialects |

### Migrations Tracking Table

```sql
-- SQLite
CREATE TABLE IF NOT EXISTS migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- PostgreSQL
CREATE TABLE IF NOT EXISTS migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Registered Migrations

| Version | Name | Description |
|---------|------|-------------|
| 1 | `create_initial_tables` | Storage roots, files, file metadata, duplicate groups, virtual paths, scan history |
| 2 | `migrate_smb_to_storage_roots` | Data migration from legacy `smb_roots` table |
| 3 | `create_auth_tables` | Users, roles, sessions, permissions, audit log |
| 4 | `create_conversion_jobs_table` | Media format conversion job tracking |
| 5 | `create_subtitle_tables` | Subtitle tracks, sync status, cache, downloads, media-subtitle association |
| 6 | `fix_subtitle_foreign_keys` | Corrects subtitle FK references from `media_items` to `files` |
| 7 | `create_assets_table` | Asset management (cover art, thumbnails, etc.) |
| 8 | `create_media_entity_tables` | Media types (11 seeded), media items, media files, collections, external metadata, user metadata, directory analyses, detection rules |
| 9 | `create_performance_indexes` | Performance-critical indexes on files, media_items, user_metadata, media_files |
| 10 | `create_sync_tables` | Sync endpoints, sync sessions, sync schedules |
| 11 | `create_service_tables` | Favorites, analytics, error/crash reports, log management, cache, wizard progress |
| 12 | `fix_service_table_columns` | Adds missing columns to v11 tables for pre-existing databases |
| 13 | `create_playlist_tables` | Playlists and playlist items |

## Dialect Abstraction

The dialect abstraction layer in `database/dialect.go` and `database/connection.go` allows application code to write SQLite-style SQL that is automatically rewritten for PostgreSQL at execution time. The `database.DB` wrapper shadows `Exec`, `ExecContext`, `Query`, `QueryContext`, `QueryRow`, and `QueryRowContext` to apply these transformations transparently.

### Rewrite Pipeline

Every query passes through `rewriteQuery()` before reaching the database driver:

```go
func (db *DB) rewriteQuery(query string) string {
    query = db.dialect.RewritePlaceholders(query)
    if db.dialect.IsPostgres() {
        query = db.dialect.RewriteInsertOrIgnore(query)
        query = db.dialect.RewriteInsertOrReplace(query)
        query = db.dialect.RewriteBooleanLiterals(query)
    }
    return query
}
```

### RewritePlaceholders

Converts `?` positional placeholders to PostgreSQL's `$1, $2, ...` format. Respects single-quoted string literals (does not rewrite `?` inside strings).

```
SQLite:     SELECT * FROM files WHERE storage_root_id = ? AND path = ?
PostgreSQL: SELECT * FROM files WHERE storage_root_id = $1 AND path = $2
```

### RewriteInsertOrIgnore

Converts SQLite's `INSERT OR IGNORE INTO` to PostgreSQL's `INSERT INTO ... ON CONFLICT DO NOTHING`.

```
SQLite:     INSERT OR IGNORE INTO roles (id, name) VALUES (?, ?)
PostgreSQL: INSERT INTO roles (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING
```

### RewriteInsertOrReplace

Converts `INSERT OR REPLACE INTO` to plain `INSERT INTO` for PostgreSQL. Callers must add their own `ON CONFLICT ... DO UPDATE` clause for PostgreSQL upsert behavior.

### RewriteBooleanLiterals

Converts integer boolean comparisons to PostgreSQL `TRUE`/`FALSE` for known boolean columns. Matched column names:

`is_active`, `is_locked`, `is_system`, `is_default`, `is_forced`, `is_duplicate`, `is_directory`, `deleted`, `enabled`, `verified_sync`, `is_favorite`, `is_public`, `is_smart`, `shuffle_enabled`, `hdr`, `dolby_vision`, `dolby_atmos`, `is_synced`

```
SQLite:     WHERE is_active = 1 AND deleted = 0
PostgreSQL: WHERE is_active = TRUE AND deleted = FALSE
```

### Dialect Helper Methods

| Method | SQLite | PostgreSQL |
|--------|--------|------------|
| `AutoIncrement()` | `INTEGER PRIMARY KEY AUTOINCREMENT` | `SERIAL PRIMARY KEY` |
| `TimestampType()` | `DATETIME` | `TIMESTAMP` |
| `BooleanDefault(true)` | `DEFAULT 1` | `DEFAULT TRUE` |
| `BooleanDefault(false)` | `DEFAULT 0` | `DEFAULT FALSE` |
| `CurrentTimestamp()` | `CURRENT_TIMESTAMP` | `CURRENT_TIMESTAMP` |

### InsertReturningID

A dialect-aware helper that returns the newly inserted row's ID:

- **PostgreSQL**: Appends `RETURNING id` to the query and uses `QueryRow().Scan()`
- **SQLite**: Uses `Exec()` followed by `result.LastInsertId()`

Also available as `TxInsertReturningID()` for use within transactions.

## Migration Versions

### Version 1: create_initial_tables

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`
**SQL Reference**: `database/migrations/000001_initial_schema.up.sql` (PostgreSQL), `000001_initial_schema.sqlite.up.sql` (SQLite)

Creates the foundational schema for file cataloging and storage management.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `storage_roots` | Network/local storage endpoints (SMB, FTP, NFS, WebDAV, local). Stores connection details, credentials, scan configuration, and filtering patterns. |
| `files` | Cataloged files with path, size, type, timestamps, and multiple hash columns (MD5, SHA256, SHA1, BLAKE3, quick_hash) for duplicate detection. Has a UNIQUE constraint on `(storage_root_id, path)`. |
| `file_metadata` | Key-value metadata pairs associated with files. Supports typed values (string default). |
| `duplicate_groups` | Groups of duplicate files identified by hash matching. Tracks count and total size. |
| `virtual_paths` | Unified path mappings across protocols. Maps a virtual path to a target entity (type + ID). |
| `scan_history` | Audit log of storage root scans. Records files processed, added, updated, deleted, and errors. |

**Dialect differences**:
- PostgreSQL: `SERIAL PRIMARY KEY`, `BOOLEAN DEFAULT TRUE/FALSE`, `TIMESTAMP`, `BIGINT` for sizes. Creates `duplicate_groups` before `files` to satisfy FK ordering.
- SQLite: `INTEGER PRIMARY KEY AUTOINCREMENT`, `BOOLEAN DEFAULT 0/1`, `DATETIME`, `INTEGER` for sizes. Inline FK on `duplicate_groups` in `files` table.

**Indexes**: `idx_files_storage_root_path` (UNIQUE), `idx_files_parent_id`, `idx_files_duplicate_group`, `idx_files_deleted`, `idx_file_metadata_file_id`, `idx_scan_history_storage_root`

**Rollback**: `000001_initial_schema.down.sql` drops all tables and indexes in reverse dependency order.

---

### Version 2: migrate_smb_to_storage_roots

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`

Data migration that converts legacy `smb_roots` table data into the new multi-protocol `storage_roots` format. This migration is **idempotent** -- it checks if the `smb_roots` table exists before attempting anything.

**Operations**:
1. Checks if the legacy `smb_roots` table exists
2. Copies SMB root entries into `storage_roots` with `protocol = 'smb'`, mapping `share` to `path`
3. (SQLite only) Updates `files.storage_root_id` and `scan_history.storage_root_id` to reference the new IDs

**No tables created. No rollback file** -- this is a one-way data migration.

---

### Version 3: create_auth_tables

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`
**SQL Reference**: `database/migrations/000003_add_user_tables.up.sql`

Creates the authentication and authorization schema with role-based access control.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `users` | User accounts with username, email, password hash + salt, profile fields, activity tracking, and lockout protection. |
| `roles` | Named roles with JSON permission arrays. |
| `user_sessions` | Active sessions with tokens, device info, IP tracking, and expiration. |
| `permissions` | Named permissions with resource/action pairs. |
| `user_permissions` | Junction table for custom per-user permission grants. Composite PK `(user_id, permission_id)`. |
| `auth_audit_log` | Audit trail for authentication events (login, logout, failed login, password change). |

**Seed Data**:
- Admin role (id=1): `permissions = '["*"]'`, `is_system = TRUE`
- User role (id=2): `permissions = '["media.view", "media.download"]'`, `is_system = TRUE`

**Dialect differences**:
- PostgreSQL: Uses `ON CONFLICT (id) DO NOTHING` for seed inserts. Resets the `roles_id_seq` sequence after seeding. Each statement executed individually.
- SQLite: Uses `INSERT OR IGNORE` for seed inserts. All statements executed as a single batch.

**Indexes**: `idx_users_username`, `idx_users_email`, `idx_users_role_id`, `idx_users_is_active`, `idx_user_sessions_user_id`, `idx_user_sessions_token`, `idx_user_sessions_expires_at`

---

### Version 4: create_conversion_jobs_table

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`
**SQL Reference**: `database/migrations/000002_conversion_jobs.up.sql`

Adds media format conversion job tracking.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `conversion_jobs` | Tracks media conversion tasks with source/target paths and formats, conversion type, quality settings, priority, scheduling, status, duration, and error reporting. |

**Indexes**: `idx_conversion_jobs_user_id`, `idx_conversion_jobs_status`, `idx_conversion_jobs_created_at`

**Note**: The Go migration schema uses `source_path`/`target_path` and includes `conversion_type`, `settings`, `priority`, `scheduled_for`, `duration` columns. The SQL reference file uses slightly different column names (`source_file_path`/`target_file_path`, `quality_level`, `progress`). The Go version is authoritative.

---

### Version 5: create_subtitle_tables

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`
**SQL Reference**: `database/migrations/014_create_subtitle_tables.up.sql`

Creates comprehensive subtitle management tables for multi-language media support.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `subtitle_tracks` | Individual subtitle tracks with language, format (SRT default), encoding, sync offset, and verified status. Initially references `media_items(id)`. |
| `subtitle_sync_status` | Tracks subtitle operations (download, upload, sync, verify) with progress and error reporting. |
| `subtitle_cache` | Temporary cache for subtitle search results from external providers. Keyed by `cache_key` (UNIQUE). |
| `subtitle_downloads` | Download history with provider, language, file details, and sync verification status. |
| `media_subtitles` | Many-to-many association between media items and subtitle tracks. UNIQUE on `(media_item_id, subtitle_track_id)`. |

**Triggers** (SQLite uses `AFTER UPDATE` triggers; PostgreSQL uses `BEFORE UPDATE` trigger functions):
- `update_subtitle_tracks_updated_at` -- Auto-updates `updated_at` on subtitle track changes
- `update_subtitle_sync_status_updated_at` -- Auto-updates `updated_at` on sync status changes
- `set_subtitle_sync_status_completed_at` -- Sets `completed_at` when status transitions to 'completed'

**Dialect differences**:
- PostgreSQL: Uses `CREATE OR REPLACE FUNCTION ... RETURNS TRIGGER` with `BEFORE UPDATE` triggers. No FK constraints on `media_items` since it does not exist yet. Subtitle tables are created without FK on `media_item_id`.
- SQLite: Uses `AFTER UPDATE` triggers with `BEGIN...END` blocks. FK references `media_items(id)`.

---

### Version 6: fix_subtitle_foreign_keys

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`
**SQL Reference**: `database/migrations/015_fix_subtitle_foreign_keys.up.sql`

Corrects foreign key references in subtitle tables from `media_items(id)` to `files(id)`.

**SQLite** (backup-recreate-restore pattern):
1. Creates backup copies of all 4 subtitle tables
2. Drops the original tables
3. Recreates tables with `FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE`
4. Restores data from backup tables
5. Drops backup tables
6. Recreates all triggers and indexes

**PostgreSQL**: No-op. The v5 migration for PostgreSQL already creates the tables without FK constraints on `media_items`, so no correction is needed.

---

### Version 7: create_assets_table

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`

Creates the assets management table for cover art, thumbnails, and other media assets.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `assets` | Asset records with TEXT primary key (`id`), type, status, content type, size, source hint, entity linkage (`entity_type`/`entity_id`), metadata (JSON), local path, and timestamp fields including `resolved_at` and `expires_at`. |

**Dialect differences**:
- PostgreSQL: `BIGINT` for `size`
- SQLite: `INTEGER` for `size`

**Indexes**: `idx_assets_entity` (entity_type, entity_id), `idx_assets_status`

---

### Version 8: create_media_entity_tables

**Source**: `migrations_sqlite.go` / `migrations_postgres.go`

Creates the core media entity system. This is the largest single migration, establishing the entity model that transforms scanned files into structured media content.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `media_types` | Catalog of 11 media type definitions with detection patterns and metadata provider configuration. |
| `media_items` | Core entity table with title, year, genre, director, cast, rating, runtime, language, country, status, and self-referencing `parent_id` for hierarchy. |
| `media_files` | Junction table linking `media_items` to `files`. Includes quality info, language, and primary flag. |
| `media_collections` | Named collections (franchise, series, soundtrack) with total item count and external IDs. |
| `media_collection_items` | Junction table linking media items to collections with ordering (sequence, season, release order). |
| `external_metadata` | Provider-specific metadata (TMDB, OMDB, MusicBrainz, OpenLibrary, etc.) with external IDs, ratings, URLs. |
| `user_metadata` | Per-user media metadata: ratings, watched status, notes, tags, favorite flag. |
| `directory_analyses` | Directory-level detection results with confidence scores and analysis data. |
| `detection_rules` | Configurable media type detection patterns with confidence weights and priority. |

**Seeded Media Types** (11 types):

| Name | Description |
|------|-------------|
| `movie` | Feature films and standalone movies |
| `tv_show` | Television series |
| `tv_season` | Season of a TV show |
| `tv_episode` | Episode of a TV season |
| `music_artist` | Music artist or band |
| `music_album` | Music album |
| `song` | Individual music track |
| `game` | Video games |
| `software` | Software applications and utilities |
| `book` | Books and e-books |
| `comic` | Comics and graphic novels |

**Dialect differences**:
- PostgreSQL: Uses batch `INSERT INTO ... VALUES (...), (...) ON CONFLICT (name) DO NOTHING` for seed data. `is_primary BOOLEAN DEFAULT FALSE` for `media_files`. `favorite BOOLEAN DEFAULT FALSE` for `user_metadata`. `BIGINT` for `total_size` in `directory_analyses`. `enabled BOOLEAN DEFAULT TRUE` for `detection_rules`.
- SQLite: Uses individual `INSERT OR IGNORE INTO` statements for each media type. `is_primary INTEGER DEFAULT 0`. `favorite INTEGER DEFAULT 0`. `INTEGER` for all sizes. `enabled INTEGER DEFAULT 1`.

**Indexes**: `idx_media_items_type`, `idx_media_items_parent`, `idx_media_items_title`, `idx_media_files_item`, `idx_media_files_file`, `idx_external_metadata_item`, `idx_external_metadata_provider` (provider, external_id), `idx_user_metadata_item`, `idx_user_metadata_user`, `idx_directory_analyses_path`, `idx_detection_rules_type`, `idx_media_collection_items_collection`, `idx_media_collection_items_item`

---

### Version 9: create_performance_indexes

**Source**: `migrations_v9_performance.go`

Adds performance-critical indexes identified from repository query patterns. This migration is idempotent (`IF NOT EXISTS`). Also deduplicates the `media_files` junction table and creates a UNIQUE compound index.

**Indexes Created**:

| Index | Table | Columns | Purpose |
|-------|-------|---------|---------|
| `idx_files_file_type` | `files` | `file_type` | Filtered in SearchFiles and stats queries |
| `idx_files_extension` | `files` | `extension` | Extension-based filtering |
| `idx_files_is_directory` | `files` | `is_directory` | Directory listing queries |
| `idx_files_name` | `files` | `name` | Name-based search |
| `idx_media_items_title_type` | `media_items` | `title, media_type_id` | Compound for GetByTitle and GetDuplicates |
| `idx_media_items_status` | `media_items` | `status` | Status filtering and duplicate detection |
| `idx_media_items_year` | `media_items` | `year` | Year-based filtering |
| `idx_user_metadata_user_watched` | `user_metadata` | `user_id, watched_status` | Watched media queries |
| `idx_media_files_item_file` (UNIQUE) | `media_files` | `media_item_id, file_id` | Prevents duplicate file-entity links |

**Deduplication** (runs before creating the UNIQUE index):
- SQLite: `DELETE FROM media_files WHERE rowid NOT IN (SELECT MIN(rowid) FROM media_files GROUP BY media_item_id, file_id)`
- PostgreSQL: `DELETE FROM media_files a USING media_files b WHERE a.ctid > b.ctid AND a.media_item_id = b.media_item_id AND a.file_id = b.file_id`

---

### Version 10: create_sync_tables

**Source**: `migrations_v10_sync_tables.go`

Creates tables for managing remote synchronization endpoints and scheduling.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `sync_endpoints` | Remote sync endpoint definitions (FTP, WebDAV, etc.) with connection details, sync direction, paths, settings, and status. |
| `sync_sessions` | Individual sync execution records with progress tracking (total, synced, failed, skipped files) and duration. |
| `sync_schedules` | Recurring sync schedule configuration with frequency, last/next run timestamps, and active flag. |

**Dialect differences**:
- PostgreSQL: `BOOLEAN DEFAULT TRUE` for `is_active`
- SQLite: `BOOLEAN DEFAULT 1` for `is_active`

**Indexes**: `idx_sync_endpoints_user_id`, `idx_sync_endpoints_status`, `idx_sync_endpoints_type`, `idx_sync_sessions_endpoint_id`, `idx_sync_sessions_user_id`, `idx_sync_sessions_status`, `idx_sync_sessions_started_at`, `idx_sync_schedules_endpoint_id`, `idx_sync_schedules_user_id`, `idx_sync_schedules_is_active`, `idx_sync_schedules_next_run`

---

### Version 11: create_service_tables

**Source**: `migrations_v11_service_tables.go`

Creates tables for analytics, favorites, error reporting, crash reporting, log management, caching, and configuration wizard. These tables were previously created lazily by individual services but are now part of the migration to ensure they exist on fresh databases.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `favorites` | User favorites with entity type/ID, category, notes, tags, public flag. UNIQUE on `(user_id, entity_type, entity_id)`. |
| `favorite_categories` | User-defined favorite categories with color, icon, sort order. UNIQUE on `(user_id, name)`. |
| `analytics_events` | User activity events with event type, data (JSON), media ID, session ID, device info, geolocation, timestamps. |
| `media_access_logs` | Media access tracking with action type, duration, playback duration, device info, location. |
| `error_reports` | Error reports with level, message, error code, component, stack trace, context, system info, fingerprint, status. |
| `crash_reports` | Crash reports with signal, crash type, message, stack trace, context, system/device info, app/OS version, fingerprint, status. |
| `log_collections` | Log collection sessions with components, log level, time range, entry count, filters. |
| `log_shares` | Shared log collection access with share token (UNIQUE), permissions, recipients, expiration. |
| `cache_entries` | General cache entries with key (UNIQUE), value, type, provider, expiration. |
| `api_cache` | API response cache with key (UNIQUE), value, expiration. |
| `media_metadata_cache` | Media metadata cache with key (UNIQUE), value, provider, expiration. |
| `analytics_reports` | Generated analytics reports with type, title, data (JSON), format. |
| `wizard_progress` | Configuration wizard progress tracking with step ID, step data (JSON), all data (JSON), completion flag. |

**Dialect differences**:
- PostgreSQL: Uses `JSONB` for JSON columns (`event_data`, `device_info`, `filters`, `report_data`, `step_data`, `all_data`). Uses `BOOLEAN` for booleans.
- SQLite: Uses `TEXT DEFAULT '{}'` for JSON columns. Uses `INTEGER DEFAULT 0/1` for booleans.

**Indexes**: `idx_favorites_user`, `idx_favorites_entity`, `idx_analytics_events_user`, `idx_analytics_events_type`, `idx_analytics_events_date`, `idx_media_access_user`, `idx_media_access_media`, `idx_error_reports_user`, `idx_error_reports_status`, `idx_error_reports_level`, `idx_crash_reports_user`, `idx_log_collections_user`, `idx_cache_entries_key`, `idx_cache_entries_expires`

---

### Version 12: fix_service_table_columns

**Source**: `migrations_v11_service_tables.go`

Adds missing columns to tables created by v11 for databases that had pre-existing (incomplete) versions of these tables. Uses `ALTER TABLE ADD COLUMN` with error suppression for "duplicate column" / "already exists" errors.

**Columns added** (if missing):

| Table | Columns |
|-------|---------|
| `error_reports` | `level`, `error_code`, `component`, `context`, `user_agent`, `url`, `reported_at`, `resolved_at` |
| `crash_reports` | `signal`, `crash_type`, `context`, `device_info`, `app_version`, `os_version`, `reported_at`, `resolved_at` |
| `log_shares` | `user_id`, `share_type`, `accessed_at`, `is_active`, `permissions`, `recipients` |
| `wizard_progress` | `current_step`, `all_data` |
| `media_access_logs` | `action`, `playback_duration`, `access_time`, `device_info` |
| `analytics_events` | `timestamp`, `device_info`, `event_category`, `access_count`, `file_type`, `session_start`, `session_end`, `data`, `duration_seconds`, `country`, `city`, `latitude`, `longitude` |
| `log_collections` | `description`, `components`, `log_level`, `start_time`, `end_time`, `completed_at`, `entry_count`, `filters` |
| `favorites` | `tags`, `is_public`, `updated_at` |

---

### Version 13: create_playlist_tables

**Source**: `migrations_v13_playlist_tables.go`

Creates user-facing playlist management tables for the `/api/v1/playlists` endpoint.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `playlists` | User playlists with name, description, public flag, timestamps. |
| `playlist_items` | Items in a playlist with entity reference (`entity_id`, `entity_type`), position ordering, and add timestamp. |

**Dialect differences**:
- PostgreSQL: `SERIAL PRIMARY KEY`, `BOOLEAN DEFAULT FALSE`, `TIMESTAMP`
- SQLite: `INTEGER PRIMARY KEY AUTOINCREMENT`, `INTEGER DEFAULT 0`, `DATETIME`

**Indexes**: `idx_playlists_user_id`, `idx_playlist_items_playlist_id`, `idx_playlist_items_entity` (entity_id, entity_type)

---

## Complete Schema Reference

This section documents the final state of every table after all 13 migrations have been applied. SQLite syntax is used as the canonical reference; see the [Dialect Abstraction](#dialect-abstraction) section for PostgreSQL equivalents.

### migrations

```sql
CREATE TABLE IF NOT EXISTS migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### storage_roots

```sql
CREATE TABLE IF NOT EXISTS storage_roots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    protocol TEXT NOT NULL,           -- 'smb', 'ftp', 'nfs', 'webdav', 'local'
    host TEXT,
    port INTEGER,
    path TEXT,
    username TEXT,
    password TEXT,
    domain TEXT,                       -- SMB domain
    mount_point TEXT,
    options TEXT,                      -- JSON protocol-specific options
    url TEXT,                          -- Full URL (for WebDAV, etc.)
    enabled BOOLEAN DEFAULT 1,
    max_depth INTEGER DEFAULT 10,
    enable_duplicate_detection BOOLEAN DEFAULT 1,
    enable_metadata_extraction BOOLEAN DEFAULT 1,
    include_patterns TEXT,             -- Glob patterns for file inclusion
    exclude_patterns TEXT,             -- Glob patterns for file exclusion
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_scan_at DATETIME
);
```

### files

```sql
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storage_root_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL,
    extension TEXT,
    mime_type TEXT,
    file_type TEXT,                     -- 'video', 'audio', 'image', 'document', etc.
    size INTEGER NOT NULL,             -- BIGINT in PostgreSQL
    is_directory BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME NOT NULL,
    accessed_at DATETIME,
    deleted BOOLEAN DEFAULT 0,
    deleted_at DATETIME,
    last_scan_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_verified_at DATETIME,
    md5 TEXT,
    sha256 TEXT,
    sha1 TEXT,
    blake3 TEXT,
    quick_hash TEXT,                   -- Fast partial-file hash for initial dedup
    is_duplicate BOOLEAN DEFAULT 0,
    duplicate_group_id INTEGER,
    parent_id INTEGER,                 -- Self-reference for directory hierarchy
    UNIQUE(storage_root_id, path),
    FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id),
    FOREIGN KEY (parent_id) REFERENCES files(id),
    FOREIGN KEY (duplicate_group_id) REFERENCES duplicate_groups(id)
);
```

### file_metadata

```sql
CREATE TABLE IF NOT EXISTS file_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    data_type TEXT DEFAULT 'string',   -- 'string', 'integer', 'float', 'boolean', 'json'
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);
```

### duplicate_groups

```sql
CREATE TABLE IF NOT EXISTS duplicate_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_count INTEGER DEFAULT 0,
    total_size INTEGER DEFAULT 0,      -- BIGINT in PostgreSQL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### virtual_paths

```sql
CREATE TABLE IF NOT EXISTS virtual_paths (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    target_type TEXT NOT NULL,          -- Entity type (e.g., 'file', 'directory')
    target_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### scan_history

```sql
CREATE TABLE IF NOT EXISTS scan_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storage_root_id INTEGER NOT NULL,
    scan_type TEXT NOT NULL,            -- 'full', 'incremental', 'quick'
    status TEXT NOT NULL,               -- 'running', 'completed', 'failed', 'cancelled'
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    files_processed INTEGER DEFAULT 0,
    files_added INTEGER DEFAULT 0,
    files_updated INTEGER DEFAULT 0,
    files_deleted INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    error_message TEXT,
    FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
);
```

### users

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    salt TEXT NOT NULL,
    role_id INTEGER NOT NULL,
    first_name TEXT,
    last_name TEXT,
    display_name TEXT,
    avatar_url TEXT,
    time_zone TEXT,
    language TEXT,
    settings TEXT DEFAULT '{}',        -- JSON user settings
    is_active INTEGER DEFAULT 1,       -- BOOLEAN DEFAULT TRUE in PostgreSQL
    is_locked INTEGER DEFAULT 0,       -- BOOLEAN DEFAULT FALSE in PostgreSQL
    locked_until DATETIME,
    failed_login_attempts INTEGER DEFAULT 0,
    last_login_at DATETIME,
    last_login_ip TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### roles

```sql
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT DEFAULT '[]',      -- JSON array of permission strings
    is_system INTEGER DEFAULT 0,       -- BOOLEAN DEFAULT FALSE in PostgreSQL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### user_sessions

```sql
CREATE TABLE IF NOT EXISTS user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    refresh_token TEXT,
    device_info TEXT,
    ip_address TEXT,
    user_agent TEXT,
    is_active INTEGER DEFAULT 1,       -- BOOLEAN DEFAULT TRUE in PostgreSQL
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### permissions

```sql
CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT
);
```

### user_permissions

```sql
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    granted_by INTEGER,
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id)
);
```

### auth_audit_log

```sql
CREATE TABLE IF NOT EXISTS auth_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    event_type TEXT NOT NULL,           -- 'login', 'logout', 'failed_login', 'password_change', etc.
    ip_address TEXT,
    user_agent TEXT,
    details TEXT,                       -- JSON additional details
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### conversion_jobs

```sql
CREATE TABLE IF NOT EXISTS conversion_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    source_path TEXT NOT NULL,
    target_path TEXT NOT NULL,
    source_format TEXT NOT NULL,
    target_format TEXT NOT NULL,
    conversion_type TEXT NOT NULL,      -- 'transcode', 'remux', 'extract', etc.
    quality TEXT DEFAULT 'medium',
    settings TEXT,                      -- JSON conversion parameters
    priority INTEGER DEFAULT 0,
    status TEXT DEFAULT 'pending',      -- 'pending', 'running', 'completed', 'failed', 'cancelled'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    scheduled_for DATETIME,
    duration INTEGER,                   -- Processing duration in seconds
    error_message TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### subtitle_tracks

After v6 fix, foreign key references `files(id)` instead of `media_items(id)`.

```sql
CREATE TABLE subtitle_tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,    -- Despite the name, references files(id) after v6
    language TEXT NOT NULL,
    language_code TEXT NOT NULL,        -- ISO 639-1 code (e.g., 'en', 'fr')
    source TEXT NOT NULL DEFAULT 'downloaded', -- 'embedded', 'downloaded', 'uploaded'
    format TEXT NOT NULL DEFAULT 'srt', -- 'srt', 'vtt', 'ass', 'ssa', 'sub'
    path TEXT,                          -- File path on disk
    content TEXT,                       -- Inline subtitle content
    is_default BOOLEAN DEFAULT FALSE,
    is_forced BOOLEAN DEFAULT FALSE,
    encoding TEXT DEFAULT 'utf-8',
    sync_offset REAL DEFAULT 0.0,      -- Timing offset in seconds
    verified_sync BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE
);
```

### subtitle_sync_status

```sql
CREATE TABLE subtitle_sync_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    subtitle_id TEXT NOT NULL,
    operation TEXT NOT NULL,            -- 'download', 'upload', 'sync', 'verify'
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'failed'
    progress INTEGER DEFAULT 0,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE
);
```

### subtitle_cache

```sql
CREATE TABLE IF NOT EXISTS subtitle_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT UNIQUE NOT NULL,
    result_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    title TEXT,
    language TEXT,
    language_code TEXT,
    download_url TEXT,
    format TEXT,
    encoding TEXT,
    upload_date DATETIME,
    downloads INTEGER,
    rating REAL,
    comments INTEGER,
    match_score REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    data TEXT                           -- JSON blob for additional data
);
```

### subtitle_downloads

```sql
CREATE TABLE subtitle_downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    result_id TEXT NOT NULL,
    subtitle_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    language TEXT NOT NULL,
    file_path TEXT,
    file_size INTEGER,
    download_url TEXT,
    download_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    verified_sync BOOLEAN DEFAULT FALSE,
    sync_offset REAL DEFAULT 0.0,
    FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE
);
```

### media_subtitles

```sql
CREATE TABLE media_subtitles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    subtitle_track_id INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (subtitle_track_id) REFERENCES subtitle_tracks(id) ON DELETE CASCADE,
    UNIQUE(media_item_id, subtitle_track_id)
);
```

### assets

```sql
CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,               -- UUID string, not auto-increment
    type TEXT NOT NULL,                 -- 'cover', 'thumbnail', 'backdrop', etc.
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'resolved', 'failed'
    content_type TEXT,                  -- MIME type (e.g., 'image/jpeg')
    size INTEGER DEFAULT 0,            -- BIGINT in PostgreSQL
    source_hint TEXT,                   -- URL or path hint for resolution
    entity_type TEXT,                   -- 'movie', 'tv_show', 'music_album', etc.
    entity_id TEXT,                     -- ID of the associated entity
    metadata TEXT,                      -- JSON metadata
    local_path TEXT,                    -- Resolved file path on disk
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    expires_at TIMESTAMP
);
```

### media_types

```sql
CREATE TABLE IF NOT EXISTS media_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    detection_patterns TEXT,            -- JSON array of patterns
    metadata_providers TEXT,            -- JSON array of provider names
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Seeded with 11 types: movie, tv_show, tv_season, tv_episode,
-- music_artist, music_album, song, game, software, book, comic
```

### media_items

```sql
CREATE TABLE IF NOT EXISTS media_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_type_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    original_title TEXT,
    year INTEGER,
    description TEXT,
    genre TEXT,                         -- Comma-separated or JSON
    director TEXT,
    cast_crew TEXT,                     -- JSON array
    rating REAL,                        -- Aggregate rating (0.0-10.0)
    runtime INTEGER,                   -- Duration in minutes
    language TEXT,
    country TEXT,
    status TEXT NOT NULL DEFAULT 'detected', -- 'detected', 'confirmed', 'manual'
    parent_id INTEGER,                 -- Self-reference for hierarchy
    season_number INTEGER,             -- For tv_season
    episode_number INTEGER,            -- For tv_episode
    track_number INTEGER,              -- For songs
    first_detected DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_type_id) REFERENCES media_types(id),
    FOREIGN KEY (parent_id) REFERENCES media_items(id) ON DELETE CASCADE
);
```

### media_files

```sql
CREATE TABLE IF NOT EXISTS media_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    quality_info TEXT,                  -- JSON quality metadata (resolution, bitrate, codec)
    language TEXT,
    is_primary INTEGER DEFAULT 0,      -- BOOLEAN DEFAULT FALSE in PostgreSQL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);
-- UNIQUE INDEX idx_media_files_item_file ON media_files(media_item_id, file_id) -- added in v9
```

### media_collections

```sql
CREATE TABLE IF NOT EXISTS media_collections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    collection_type TEXT NOT NULL,      -- 'franchise', 'series', 'soundtrack', 'box_set'
    description TEXT,
    total_items INTEGER DEFAULT 0,
    external_ids TEXT,                  -- JSON map of provider IDs
    cover_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### media_collection_items

```sql
CREATE TABLE IF NOT EXISTS media_collection_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_id INTEGER NOT NULL,
    media_item_id INTEGER NOT NULL,
    sequence_number INTEGER,           -- Position within the collection
    season_number INTEGER,             -- For TV series collections
    release_order INTEGER,             -- Chronological release order
    FOREIGN KEY (collection_id) REFERENCES media_collections(id) ON DELETE CASCADE,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);
```

### external_metadata

```sql
CREATE TABLE IF NOT EXISTS external_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    provider TEXT NOT NULL,            -- 'tmdb', 'omdb', 'musicbrainz', 'openlibrary', etc.
    external_id TEXT NOT NULL,         -- Provider-specific ID
    data TEXT,                         -- JSON full metadata from provider
    rating REAL,                       -- Provider-specific rating
    review_url TEXT,
    cover_url TEXT,
    trailer_url TEXT,
    last_fetched DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);
```

### user_metadata

```sql
CREATE TABLE IF NOT EXISTS user_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    user_rating REAL,                  -- User's personal rating (0.0-10.0)
    watched_status TEXT,               -- 'unwatched', 'watching', 'watched', 'on_hold', 'dropped'
    watched_date DATETIME,
    personal_notes TEXT,
    tags TEXT,                         -- JSON array of user tags
    favorite INTEGER DEFAULT 0,        -- BOOLEAN DEFAULT FALSE in PostgreSQL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### directory_analyses

```sql
CREATE TABLE IF NOT EXISTS directory_analyses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    directory_path TEXT NOT NULL,
    smb_root TEXT,
    media_item_id INTEGER,
    confidence_score REAL DEFAULT 0,   -- 0.0-1.0 detection confidence
    detection_method TEXT,             -- 'title_parse', 'nfo_file', 'metadata', etc.
    analysis_data TEXT,                -- JSON analysis results
    last_analyzed DATETIME DEFAULT CURRENT_TIMESTAMP,
    files_count INTEGER DEFAULT 0,
    total_size INTEGER DEFAULT 0,      -- BIGINT in PostgreSQL
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE SET NULL
);
```

### detection_rules

```sql
CREATE TABLE IF NOT EXISTS detection_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_type_id INTEGER NOT NULL,
    rule_name TEXT NOT NULL,
    rule_type TEXT NOT NULL,           -- 'regex', 'extension', 'directory', 'nfo'
    pattern TEXT NOT NULL,
    confidence_weight REAL DEFAULT 1.0,
    enabled INTEGER DEFAULT 1,         -- BOOLEAN DEFAULT TRUE in PostgreSQL
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_type_id) REFERENCES media_types(id) ON DELETE CASCADE
);
```

### sync_endpoints

```sql
CREATE TABLE IF NOT EXISTS sync_endpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,                -- 'ftp', 'webdav', 'sftp', 's3', etc.
    url TEXT NOT NULL,
    username TEXT,
    password TEXT,
    sync_direction TEXT DEFAULT 'bidirectional', -- 'upload', 'download', 'bidirectional'
    local_path TEXT,
    remote_path TEXT,
    sync_settings TEXT,                -- JSON configuration
    status TEXT DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_sync_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### sync_sessions

```sql
CREATE TABLE IF NOT EXISTS sync_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    endpoint_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    status TEXT DEFAULT 'running',     -- 'running', 'completed', 'failed', 'cancelled'
    sync_type TEXT,                     -- 'full', 'incremental', 'delta'
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    duration INTEGER,                  -- Seconds
    total_files INTEGER DEFAULT 0,
    synced_files INTEGER DEFAULT 0,
    failed_files INTEGER DEFAULT 0,
    skipped_files INTEGER DEFAULT 0,
    error_message TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### sync_schedules

```sql
CREATE TABLE IF NOT EXISTS sync_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    endpoint_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    frequency TEXT NOT NULL,           -- Cron expression or named interval
    last_run DATETIME,
    next_run DATETIME,
    is_active BOOLEAN DEFAULT 1,       -- BOOLEAN DEFAULT TRUE in PostgreSQL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### favorites

```sql
CREATE TABLE IF NOT EXISTS favorites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    entity_type TEXT NOT NULL,         -- 'media_item', 'file', 'collection', etc.
    entity_id INTEGER NOT NULL,
    category TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    tags TEXT,                          -- JSON array
    is_public INTEGER DEFAULT 0,       -- BOOLEAN DEFAULT FALSE in PostgreSQL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, entity_type, entity_id)
);
```

### favorite_categories

```sql
CREATE TABLE IF NOT EXISTS favorite_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    color TEXT DEFAULT '',             -- Hex color code
    icon TEXT DEFAULT '',              -- Icon name/path
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);
```

### analytics_events

```sql
CREATE TABLE IF NOT EXISTS analytics_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    event_type TEXT NOT NULL,           -- 'play', 'pause', 'search', 'browse', etc.
    event_data TEXT DEFAULT '{}',       -- JSONB in PostgreSQL
    media_id INTEGER,
    session_id TEXT,
    device_info TEXT DEFAULT '{}',      -- JSONB in PostgreSQL
    ip_address TEXT,
    user_agent TEXT,
    device_type TEXT,
    location TEXT,
    event_category TEXT DEFAULT '',
    access_count INTEGER DEFAULT 0,
    file_type TEXT DEFAULT '',
    data TEXT DEFAULT '{}',
    duration_seconds INTEGER DEFAULT 0,
    country TEXT DEFAULT '',
    city TEXT DEFAULT '',
    latitude REAL DEFAULT 0,
    longitude REAL DEFAULT 0,
    session_start DATETIME,
    session_end DATETIME,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### media_access_logs

```sql
CREATE TABLE IF NOT EXISTS media_access_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    media_id INTEGER NOT NULL,
    action TEXT NOT NULL DEFAULT 'view',
    access_type TEXT NOT NULL DEFAULT 'view', -- 'view', 'download', 'stream'
    duration INTEGER DEFAULT 0,
    playback_duration INTEGER DEFAULT 0,
    ip_address TEXT,
    user_agent TEXT,
    device_type TEXT,
    device_info TEXT DEFAULT '{}',
    location TEXT,
    access_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### error_reports

```sql
CREATE TABLE IF NOT EXISTS error_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    level TEXT NOT NULL DEFAULT 'error', -- 'error', 'warning', 'critical'
    message TEXT,
    error_code TEXT DEFAULT '',
    component TEXT DEFAULT '',          -- System component that generated the error
    stack_trace TEXT,
    context TEXT DEFAULT '{}',          -- JSON contextual data
    system_info TEXT DEFAULT '{}',      -- JSON system information
    user_agent TEXT DEFAULT '',
    url TEXT DEFAULT '',
    fingerprint TEXT DEFAULT '',        -- Deduplication fingerprint
    status TEXT DEFAULT 'new',         -- 'new', 'acknowledged', 'resolved', 'ignored'
    reported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME
);
```

### crash_reports

```sql
CREATE TABLE IF NOT EXISTS crash_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    signal TEXT NOT NULL DEFAULT 'crash',
    crash_type TEXT NOT NULL DEFAULT 'crash',
    message TEXT,
    stack_trace TEXT,
    context TEXT DEFAULT '{}',
    system_info TEXT DEFAULT '{}',
    device_info TEXT DEFAULT '{}',
    app_version TEXT DEFAULT '',
    os_version TEXT DEFAULT '',
    fingerprint TEXT DEFAULT '',
    status TEXT DEFAULT 'new',
    reported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME
);
```

### log_collections

```sql
CREATE TABLE IF NOT EXISTS log_collections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    components TEXT DEFAULT '[]',       -- JSON array of component names
    log_level TEXT DEFAULT 'info',
    start_time DATETIME,
    end_time DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    status TEXT DEFAULT 'active',
    entry_count INTEGER DEFAULT 0,
    filters TEXT DEFAULT '{}'           -- JSONB in PostgreSQL
);
```

### log_shares

```sql
CREATE TABLE IF NOT EXISTS log_shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL DEFAULT 0,
    share_token TEXT NOT NULL UNIQUE,
    share_type TEXT DEFAULT 'link',
    created_by INTEGER NOT NULL DEFAULT 0,
    can_read INTEGER DEFAULT 1,        -- BOOLEAN DEFAULT TRUE in PostgreSQL
    can_write INTEGER DEFAULT 0,       -- BOOLEAN DEFAULT FALSE in PostgreSQL
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    accessed_at DATETIME,
    is_active INTEGER DEFAULT 1,       -- BOOLEAN DEFAULT TRUE in PostgreSQL
    permissions TEXT DEFAULT '{}',
    recipients TEXT DEFAULT '[]',
    FOREIGN KEY (collection_id) REFERENCES log_collections(id) ON DELETE CASCADE
);
```

### cache_entries

```sql
CREATE TABLE IF NOT EXISTS cache_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT NOT NULL UNIQUE,
    cache_value TEXT,
    cache_type TEXT DEFAULT 'general',
    provider TEXT DEFAULT '',
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### api_cache

```sql
CREATE TABLE IF NOT EXISTS api_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT NOT NULL UNIQUE,
    cache_value TEXT,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### media_metadata_cache

```sql
CREATE TABLE IF NOT EXISTS media_metadata_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT NOT NULL UNIQUE,
    cache_value TEXT,
    provider TEXT DEFAULT '',
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### analytics_reports

```sql
CREATE TABLE IF NOT EXISTS analytics_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    report_type TEXT NOT NULL,
    title TEXT,
    report_data TEXT DEFAULT '{}',      -- JSONB in PostgreSQL
    format TEXT DEFAULT 'json',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### wizard_progress

```sql
CREATE TABLE IF NOT EXISTS wizard_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    current_step TEXT NOT NULL DEFAULT '',
    step_id TEXT NOT NULL DEFAULT '',
    step_data TEXT DEFAULT '{}',        -- JSONB in PostgreSQL
    all_data TEXT DEFAULT '{}',         -- JSONB in PostgreSQL
    completed INTEGER DEFAULT 0,       -- BOOLEAN DEFAULT FALSE in PostgreSQL
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### playlists

```sql
CREATE TABLE IF NOT EXISTS playlists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    is_public INTEGER DEFAULT 0,       -- BOOLEAN DEFAULT FALSE in PostgreSQL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### playlist_items

```sql
CREATE TABLE IF NOT EXISTS playlist_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    playlist_id INTEGER NOT NULL,
    entity_id INTEGER NOT NULL,
    entity_type TEXT NOT NULL,          -- 'media_item', 'file', etc.
    position INTEGER DEFAULT 0,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
);
```

## Index Definitions

### V1 Indexes (Initial Schema)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_files_storage_root_path` | `files` | `storage_root_id, path` | UNIQUE |
| `idx_files_parent_id` | `files` | `parent_id` | Standard |
| `idx_files_duplicate_group` | `files` | `duplicate_group_id` | Standard |
| `idx_files_deleted` | `files` | `deleted` | Standard |
| `idx_file_metadata_file_id` | `file_metadata` | `file_id` | Standard |
| `idx_scan_history_storage_root` | `scan_history` | `storage_root_id` | Standard |

### V3 Indexes (Auth)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_users_username` | `users` | `username` | Standard |
| `idx_users_email` | `users` | `email` | Standard |
| `idx_users_role_id` | `users` | `role_id` | Standard |
| `idx_users_is_active` | `users` | `is_active` | Standard |
| `idx_user_sessions_user_id` | `user_sessions` | `user_id` | Standard |
| `idx_user_sessions_token` | `user_sessions` | `session_token` | Standard |
| `idx_user_sessions_expires_at` | `user_sessions` | `expires_at` | Standard |

### V4 Indexes (Conversion Jobs)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_conversion_jobs_user_id` | `conversion_jobs` | `user_id` | Standard |
| `idx_conversion_jobs_status` | `conversion_jobs` | `status` | Standard |
| `idx_conversion_jobs_created_at` | `conversion_jobs` | `created_at` | Standard |

### V5 Indexes (Subtitles)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_subtitle_tracks_media_item_id` | `subtitle_tracks` | `media_item_id` | Standard |
| `idx_subtitle_tracks_language` | `subtitle_tracks` | `language` | Standard |
| `idx_subtitle_tracks_language_code` | `subtitle_tracks` | `language_code` | Standard |
| `idx_subtitle_tracks_source` | `subtitle_tracks` | `source` | Standard |
| `idx_subtitle_sync_status_media_item_id` | `subtitle_sync_status` | `media_item_id` | Standard |
| `idx_subtitle_sync_status_status` | `subtitle_sync_status` | `status` | Standard |
| `idx_subtitle_sync_status_operation` | `subtitle_sync_status` | `operation` | Standard |
| `idx_subtitle_cache_cache_key` | `subtitle_cache` | `cache_key` | Standard |
| `idx_subtitle_cache_expires_at` | `subtitle_cache` | `expires_at` | Standard |
| `idx_subtitle_downloads_media_item_id` | `subtitle_downloads` | `media_item_id` | Standard |
| `idx_subtitle_downloads_result_id` | `subtitle_downloads` | `result_id` | Standard |
| `idx_subtitle_downloads_subtitle_id` | `subtitle_downloads` | `subtitle_id` | Standard |
| `idx_subtitle_downloads_provider` | `subtitle_downloads` | `provider` | Standard |
| `idx_subtitle_downloads_language` | `subtitle_downloads` | `language` | Standard |
| `idx_subtitle_downloads_download_date` | `subtitle_downloads` | `download_date` | Standard |
| `idx_media_subtitles_media_item_id` | `media_subtitles` | `media_item_id` | Standard |
| `idx_media_subtitles_subtitle_track_id` | `media_subtitles` | `subtitle_track_id` | Standard |
| `idx_media_subtitles_is_active` | `media_subtitles` | `is_active` | Standard |

### V7 Indexes (Assets)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_assets_entity` | `assets` | `entity_type, entity_id` | Standard |
| `idx_assets_status` | `assets` | `status` | Standard |

### V8 Indexes (Media Entities)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_media_items_type` | `media_items` | `media_type_id` | Standard |
| `idx_media_items_parent` | `media_items` | `parent_id` | Standard |
| `idx_media_items_title` | `media_items` | `title` | Standard |
| `idx_media_files_item` | `media_files` | `media_item_id` | Standard |
| `idx_media_files_file` | `media_files` | `file_id` | Standard |
| `idx_external_metadata_item` | `external_metadata` | `media_item_id` | Standard |
| `idx_external_metadata_provider` | `external_metadata` | `provider, external_id` | Standard |
| `idx_user_metadata_item` | `user_metadata` | `media_item_id` | Standard |
| `idx_user_metadata_user` | `user_metadata` | `user_id` | Standard |
| `idx_directory_analyses_path` | `directory_analyses` | `directory_path` | Standard |
| `idx_detection_rules_type` | `detection_rules` | `media_type_id` | Standard |
| `idx_media_collection_items_collection` | `media_collection_items` | `collection_id` | Standard |
| `idx_media_collection_items_item` | `media_collection_items` | `media_item_id` | Standard |

### V9 Indexes (Performance)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_files_file_type` | `files` | `file_type` | Standard |
| `idx_files_extension` | `files` | `extension` | Standard |
| `idx_files_is_directory` | `files` | `is_directory` | Standard |
| `idx_files_name` | `files` | `name` | Standard |
| `idx_media_items_title_type` | `media_items` | `title, media_type_id` | Standard |
| `idx_media_items_status` | `media_items` | `status` | Standard |
| `idx_media_items_year` | `media_items` | `year` | Standard |
| `idx_user_metadata_user_watched` | `user_metadata` | `user_id, watched_status` | Standard |
| `idx_media_files_item_file` | `media_files` | `media_item_id, file_id` | UNIQUE |

### V10 Indexes (Sync)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_sync_endpoints_user_id` | `sync_endpoints` | `user_id` | Standard |
| `idx_sync_endpoints_status` | `sync_endpoints` | `status` | Standard |
| `idx_sync_endpoints_type` | `sync_endpoints` | `type` | Standard |
| `idx_sync_sessions_endpoint_id` | `sync_sessions` | `endpoint_id` | Standard |
| `idx_sync_sessions_user_id` | `sync_sessions` | `user_id` | Standard |
| `idx_sync_sessions_status` | `sync_sessions` | `status` | Standard |
| `idx_sync_sessions_started_at` | `sync_sessions` | `started_at` | Standard |
| `idx_sync_schedules_endpoint_id` | `sync_schedules` | `endpoint_id` | Standard |
| `idx_sync_schedules_user_id` | `sync_schedules` | `user_id` | Standard |
| `idx_sync_schedules_is_active` | `sync_schedules` | `is_active` | Standard |
| `idx_sync_schedules_next_run` | `sync_schedules` | `next_run` | Standard |

### V11 Indexes (Service Tables)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_favorites_user` | `favorites` | `user_id` | Standard |
| `idx_favorites_entity` | `favorites` | `entity_type, entity_id` | Standard |
| `idx_analytics_events_user` | `analytics_events` | `user_id` | Standard |
| `idx_analytics_events_type` | `analytics_events` | `event_type` | Standard |
| `idx_analytics_events_date` | `analytics_events` | `created_at` | Standard |
| `idx_media_access_user` | `media_access_logs` | `user_id` | Standard |
| `idx_media_access_media` | `media_access_logs` | `media_id` | Standard |
| `idx_error_reports_user` | `error_reports` | `user_id` | Standard |
| `idx_error_reports_status` | `error_reports` | `status` | Standard |
| `idx_error_reports_level` | `error_reports` | `level` | Standard |
| `idx_crash_reports_user` | `crash_reports` | `user_id` | Standard |
| `idx_log_collections_user` | `log_collections` | `user_id` | Standard |
| `idx_cache_entries_key` | `cache_entries` | `cache_key` | Standard |
| `idx_cache_entries_expires` | `cache_entries` | `expires_at` | Standard |

### V13 Indexes (Playlists)

| Index | Table | Columns | Type |
|-------|-------|---------|------|
| `idx_playlists_user_id` | `playlists` | `user_id` | Standard |
| `idx_playlist_items_playlist_id` | `playlist_items` | `playlist_id` | Standard |
| `idx_playlist_items_entity` | `playlist_items` | `entity_id, entity_type` | Standard |

## Foreign Key Relationships

```
storage_roots
    ^
    |  storage_root_id
    |
files ----parent_id----> files (self-ref)
    |  \
    |   \--duplicate_group_id--> duplicate_groups
    |
    +---file_id----------> file_metadata
    |
    +---file_id----------> media_files ------media_item_id----> media_items
    |                                                              |   ^
    +---media_item_id----> subtitle_tracks                         |   |
    +---media_item_id----> subtitle_sync_status     media_type_id--+   |
    +---media_item_id----> subtitle_downloads                      |   |
    +---media_item_id----> media_subtitles ---subtitle_track_id--> subtitle_tracks
                                                                   |
                                                  parent_id--------+ (self-ref)
                                                                   |
media_items <--media_item_id-- external_metadata                   |
media_items <--media_item_id-- user_metadata ---user_id--> users   |
media_items <--media_item_id-- directory_analyses                  |
media_items <--media_item_id-- media_collection_items              |
                                  |                                |
                                  +--collection_id--> media_collections
                                                                   |
media_types <--media_type_id-- media_items                         |
media_types <--media_type_id-- detection_rules

users
    ^
    |  user_id
    +----------- user_sessions
    +----------- user_permissions --permission_id--> permissions
    +----------- auth_audit_log
    +----------- conversion_jobs
    +----------- sync_endpoints
    +----------- sync_sessions
    +----------- sync_schedules
    +----------- favorites
    +----------- playlists ----> playlist_items

roles <--role_id-- users

log_collections <--collection_id-- log_shares

scan_history --storage_root_id--> storage_roots
```

### Key Relationship Details

**Media Entity Hierarchy** (`media_items.parent_id` self-reference):
- TV Show -> TV Seasons -> TV Episodes
- Music Artist -> Music Albums -> Songs
- Other types (movie, game, software, book, comic) are typically root-level entities

**File-to-Entity Link** (`media_files` junction):
- A media item can have multiple files (different quality versions, formats)
- A file can belong to multiple media items (e.g., a multi-episode file)
- The `is_primary` flag marks the preferred version
- UNIQUE constraint on `(media_item_id, file_id)` prevents duplicate links

**Subtitle FK Note**: After migration v6, subtitle tables reference `files(id)` through the `media_item_id` column (the column name is historical; it now points to files, not media_items).

## Column Descriptions

### Common Patterns

| Column Pattern | Description |
|----------------|-------------|
| `created_at` | Row creation timestamp, defaults to `CURRENT_TIMESTAMP` |
| `updated_at` | Last modification timestamp, defaults to `CURRENT_TIMESTAMP` (auto-updated by triggers on subtitle tables) |
| `status` | State machine column -- values depend on context (see table-specific documentation) |
| `is_active` / `enabled` | Soft-enable/disable flag |
| `deleted` / `deleted_at` | Soft-delete pattern |

### Hash Columns (files table)

| Column | Algorithm | Purpose |
|--------|-----------|---------|
| `md5` | MD5 | Legacy compatibility, fast but not collision-resistant |
| `sha256` | SHA-256 | Primary dedup hash, cryptographically strong |
| `sha1` | SHA-1 | Secondary hash, widely supported |
| `blake3` | BLAKE3 | High-performance hash for large files |
| `quick_hash` | Partial file hash | Fast initial dedup (hashes first/last N bytes + size) |

### Status Values

| Table | Column | Values |
|-------|--------|--------|
| `scan_history` | `status` | `running`, `completed`, `failed`, `cancelled` |
| `conversion_jobs` | `status` | `pending`, `running`, `completed`, `failed`, `cancelled` |
| `subtitle_sync_status` | `status` | `pending`, `in_progress`, `completed`, `failed` |
| `media_items` | `status` | `detected`, `confirmed`, `manual` |
| `assets` | `status` | `pending`, `resolved`, `failed` |
| `sync_sessions` | `status` | `running`, `completed`, `failed`, `cancelled` |
| `error_reports` | `status` | `new`, `acknowledged`, `resolved`, `ignored` |
| `crash_reports` | `status` | `new`, `acknowledged`, `resolved`, `ignored` |
| `log_collections` | `status` | `active`, `completed`, `archived` |

## Sample Queries

### Search files by name and type

```sql
SELECT f.id, f.name, f.path, f.file_type, f.size, sr.name AS root_name
FROM files f
JOIN storage_roots sr ON f.storage_root_id = sr.id
WHERE f.name LIKE '%avengers%'
  AND f.file_type = 'video'
  AND f.deleted = 0
ORDER BY f.modified_at DESC;
```

### Browse media entities by type

```sql
SELECT mi.id, mi.title, mi.year, mi.rating, mt.name AS type_name
FROM media_items mi
JOIN media_types mt ON mi.media_type_id = mt.id
WHERE mt.name = 'movie'
  AND mi.status = 'confirmed'
ORDER BY mi.title;
```

### TV show hierarchy traversal (show -> seasons -> episodes)

```sql
-- Get a TV show with its seasons and episodes
SELECT
    show.title AS show_title,
    season.season_number,
    ep.episode_number,
    ep.title AS episode_title,
    ep.runtime
FROM media_items show
JOIN media_items season ON season.parent_id = show.id
JOIN media_items ep ON ep.parent_id = season.id
JOIN media_types mt ON show.media_type_id = mt.id
WHERE mt.name = 'tv_show'
  AND show.title = 'Breaking Bad'
ORDER BY season.season_number, ep.episode_number;
```

### Music hierarchy traversal (artist -> albums -> songs)

```sql
SELECT
    artist.title AS artist_name,
    album.title AS album_title,
    album.year,
    song.title AS song_title,
    song.track_number
FROM media_items artist
JOIN media_items album ON album.parent_id = artist.id
JOIN media_items song ON song.parent_id = album.id
JOIN media_types mt ON artist.media_type_id = mt.id
WHERE mt.name = 'music_artist'
  AND artist.title = 'Pink Floyd'
ORDER BY album.year, song.track_number;
```

### Find files linked to a media item

```sql
SELECT f.name, f.path, f.size, f.extension, mf.quality_info, mf.is_primary
FROM media_files mf
JOIN files f ON mf.file_id = f.id
WHERE mf.media_item_id = ?
ORDER BY mf.is_primary DESC, f.size DESC;
```

### Get user's watched media with ratings

```sql
SELECT mi.title, mi.year, mt.name AS type_name,
       um.user_rating, um.watched_status, um.watched_date
FROM user_metadata um
JOIN media_items mi ON um.media_item_id = mi.id
JOIN media_types mt ON mi.media_type_id = mt.id
WHERE um.user_id = ?
  AND um.watched_status = 'watched'
ORDER BY um.watched_date DESC;
```

### Find duplicate files

```sql
SELECT dg.id AS group_id, dg.file_count, dg.total_size,
       f.name, f.path, f.size, sr.name AS root_name
FROM duplicate_groups dg
JOIN files f ON f.duplicate_group_id = dg.id
JOIN storage_roots sr ON f.storage_root_id = sr.id
WHERE dg.file_count > 1
ORDER BY dg.total_size DESC, dg.id, f.name;
```

### Get external metadata for a media item

```sql
SELECT em.provider, em.external_id, em.rating, em.cover_url, em.trailer_url,
       em.last_fetched
FROM external_metadata em
WHERE em.media_item_id = ?
ORDER BY em.last_fetched DESC;
```

### Storage root scan statistics

```sql
SELECT sr.name, sr.protocol, sr.last_scan_at,
       sh.scan_type, sh.status, sh.files_processed,
       sh.files_added, sh.files_updated, sh.files_deleted,
       sh.error_count
FROM scan_history sh
JOIN storage_roots sr ON sh.storage_root_id = sr.id
ORDER BY sh.start_time DESC
LIMIT 20;
```

### User playlist with items

```sql
SELECT p.name AS playlist_name, p.description,
       pi.position, pi.entity_type, pi.entity_id,
       mi.title, mt.name AS media_type
FROM playlists p
JOIN playlist_items pi ON pi.playlist_id = p.id
LEFT JOIN media_items mi ON pi.entity_id = mi.id AND pi.entity_type = 'media_item'
LEFT JOIN media_types mt ON mi.media_type_id = mt.id
WHERE p.user_id = ?
ORDER BY p.name, pi.position;
```

## How Migrations Are Applied

### Automatic (Default)

Migrations run automatically when the application starts via `main.go`:

```go
log.Println("Running database migrations...")
if err := databaseDB.RunMigrations(ctx); err != nil {
    log.Fatal("Failed to run database migrations:", err)
}
log.Println("Database migrations completed successfully")
```

The `RunMigrations()` function:
1. Creates the `migrations` tracking table if it does not exist
2. Iterates through all 13 registered migrations in order
3. For each migration, checks if the version exists in the `migrations` table
4. Skips already-applied migrations
5. Executes the migration's `Up` function (dialect-dispatched)
6. Records the version and name in the `migrations` table

### Manual (CLI)

For manual control with the SQL reference files in `database/migrations/`:

```bash
# Install the CLI tool
go install -tags 'postgres sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Apply all pending migrations (SQLite)
migrate -path catalog-api/database/migrations \
    -database "sqlite3://./catalogizer.db" up

# Apply all pending migrations (PostgreSQL)
migrate -path catalog-api/database/migrations \
    -database "postgres://catalogizer:password@localhost:5432/catalogizer?sslmode=disable" up

# Roll back one migration
migrate -path catalog-api/database/migrations \
    -database "sqlite3://./catalogizer.db" down 1

# Check current version
migrate -path catalog-api/database/migrations \
    -database "sqlite3://./catalogizer.db" version
```

**Note**: The CLI migration files only cover versions 1-3 and subtitles (014/015). The Go-based programmatic migrations (v1-v13) are the authoritative source.

### Docker/Podman

Migrations run automatically when the container starts, as part of the normal application startup sequence.

## How to Create a New Migration

### Step 1: Add dialect-specific implementations

Create implementation functions in the appropriate file. For a new migration, either add to an existing `migrations_v*.go` file or create a new one (e.g., `migrations_v14_my_feature.go`):

```go
package database

import (
    "context"
    "fmt"
)

func (db *DB) createMyNewTables(ctx context.Context) error {
    if db.dialect.IsPostgres() {
        return db.createMyNewTablesPostgres(ctx)
    }
    return db.createMyNewTablesSQLite(ctx)
}

func (db *DB) createMyNewTablesSQLite(ctx context.Context) error {
    schema := `
    CREATE TABLE IF NOT EXISTS my_table (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        is_active INTEGER DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_my_table_name ON my_table(name);
    `
    _, err := db.ExecContext(ctx, schema)
    return err
}

func (db *DB) createMyNewTablesPostgres(ctx context.Context) error {
    statements := []string{
        `CREATE TABLE IF NOT EXISTS my_table (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            is_active BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE INDEX IF NOT EXISTS idx_my_table_name ON my_table(name)`,
    }
    for _, stmt := range statements {
        if _, err := db.ExecContext(ctx, stmt); err != nil {
            return fmt.Errorf("failed: %w", err)
        }
    }
    return nil
}
```

### Step 2: Register the migration

Add the migration to the `migrations` slice in `RunMigrations()` in `migrations.go`:

```go
migrations := []Migration{
    // ... existing migrations ...
    {Version: 14, Name: "create_my_new_tables", Up: db.createMyNewTables},
}
```

### Step 3: Follow migration guidelines

1. **Always use `IF NOT EXISTS`** / `IF EXISTS` for idempotency
2. **Never modify existing migrations** that have been deployed
3. **Test on both SQLite and PostgreSQL**
4. **Use `SERIAL` / `TIMESTAMP` / `BOOLEAN` for PostgreSQL** and `INTEGER PRIMARY KEY AUTOINCREMENT` / `DATETIME` / `INTEGER` for SQLite
5. **Document complex migrations** with Go doc comments
6. **For PostgreSQL triggers**, use `CREATE OR REPLACE FUNCTION ... RETURNS TRIGGER` (not SQLite's `CREATE TRIGGER ... BEGIN ... END`)
7. **For column additions to existing tables**, suppress "already exists" errors (see v12 pattern)

### SQLite vs PostgreSQL Syntax Summary

| Feature | PostgreSQL | SQLite |
|---------|------------|--------|
| Auto-increment | `SERIAL PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| Boolean | `BOOLEAN DEFAULT TRUE/FALSE` | `INTEGER DEFAULT 1/0` |
| Timestamp | `TIMESTAMP` | `DATETIME` |
| Big integers | `BIGINT` | `INTEGER` |
| JSON columns | `JSONB` | `TEXT` (stored as JSON string) |
| ALTER TABLE | Full support | Limited (no DROP COLUMN, no ALTER CONSTRAINT) |
| Triggers | `CREATE FUNCTION` + `CREATE TRIGGER ... EXECUTE FUNCTION` | `CREATE TRIGGER ... BEGIN ... END` |
| FK change | `ALTER TABLE ... DROP/ADD CONSTRAINT` | Backup-recreate-restore pattern |
| Upsert seed | `ON CONFLICT (col) DO NOTHING` | `INSERT OR IGNORE` |
| Sequence reset | `SELECT setval('table_id_seq', ...)` | Not needed |

## Troubleshooting

### Dirty Database State

If a migration fails mid-way using the CLI tool:

```bash
migrate -path catalog-api/database/migrations \
    -database "sqlite3://./catalogizer.db" force VERSION
```

The Go-based migration system does not have a "dirty" state concept -- it simply checks whether a version number exists in the `migrations` table.

### Reset Database (Development Only)

```bash
# SQLite: delete the file and restart
rm catalogizer.db
cd catalog-api && go run main.go

# PostgreSQL: drop and recreate
psql -U catalogizer -c "DROP DATABASE catalogizer; CREATE DATABASE catalogizer;"
cd catalog-api && go run main.go
```

**WARNING**: This deletes all data. Only use in development environments.

### Checking Applied Migrations

```sql
SELECT version, name, applied_at FROM migrations ORDER BY version;
```

Expected output for a fully migrated database:

| version | name | applied_at |
|---------|------|------------|
| 1 | create_initial_tables | (timestamp) |
| 2 | migrate_smb_to_storage_roots | (timestamp) |
| 3 | create_auth_tables | (timestamp) |
| 4 | create_conversion_jobs_table | (timestamp) |
| 5 | create_subtitle_tables | (timestamp) |
| 6 | fix_subtitle_foreign_keys | (timestamp) |
| 7 | create_assets_table | (timestamp) |
| 8 | create_media_entity_tables | (timestamp) |
| 9 | create_performance_indexes | (timestamp) |
| 10 | create_sync_tables | (timestamp) |
| 11 | create_service_tables | (timestamp) |
| 12 | fix_service_table_columns | (timestamp) |
| 13 | create_playlist_tables | (timestamp) |

### SQLite WAL Mode

SQLite connections explicitly set `PRAGMA journal_mode=WAL` after opening the connection in `database/connection.go`. This is necessary because go-sqlcipher ignores `_journal_mode=WAL` in the connection string. The connection string also sets `_busy_timeout=30000`, `_synchronous=NORMAL`, and `_foreign_keys=1`.

### Connection Pool Defaults

| Setting | Default | Config Key |
|---------|---------|------------|
| MaxOpenConns | 25 | `max_open_connections` |
| MaxIdleConns | 10 | `max_idle_connections` |
| ConnMaxLifetime | 5 min | `conn_max_lifetime` |
| ConnMaxIdleTime | 3 min | `conn_max_idle_time` |

## Additional Schema Files

### `catalog-api/migrations/005_media_player_features.sql`

Standalone migration defining an extended media player schema. Not part of the programmatic migration system. Includes tables for audio tracks, chapters, cover art, lyrics, playback sessions, playlists, user preferences, translation cache, external API cache, media analysis queue, language preferences, and supported languages (20 pre-loaded). Also defines views `media_with_metadata` and `playlists_with_items`.

### `catalog-api/migrations/006_media_items_schema_update.sql`

Updates `media_items` for Android TV compatibility. Adds columns (`directory_path`, `smb_path`, `external_metadata`, `versions`, `watch_progress`, `last_watched`, `is_downloaded`), recreates the table with a simplified schema, and migrates existing data.

### `catalog-api/internal/media/database/schema.sql`

Media detection and metadata database schema used by the detection pipeline. Defines its own set of tables including `media_types` (40+ classifications), `media_items`, `external_metadata`, `directory_analysis`, `media_files`, `quality_profiles`, `change_log`, `media_collections`, `media_collection_items`, `user_metadata`, and `detection_rules`. Also defines views `media_overview` and `duplicate_media`.

## Related Documentation

- [Database Schema](DATABASE_SCHEMA.md) -- Complete table and index reference
- [Architecture Overview](ARCHITECTURE.md) -- System design and component interactions
- [Auth Flow](AUTH_FLOW.md) -- Authentication system details
- [Concurrency Patterns](CONCURRENCY_PATTERNS.md) -- Database concurrency and connection pool usage

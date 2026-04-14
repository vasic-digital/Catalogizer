# SQL Migration Reference

**Document Version:** 1.0
**Last Updated:** April 14, 2026
**Applies to:** Catalogizer v2.3.0+ (catalog-api)

---

## Table of Contents

1. [Migration Versioning Scheme](#1-migration-versioning-scheme)
2. [Migration Registry](#2-migration-registry)
3. [Dialect Abstraction Layer](#3-dialect-abstraction-layer)
4. [Migration Details (v1 through v15)](#4-migration-details)
5. [Current Complete Schema Summary](#5-current-complete-schema-summary)
6. [Database ERD (Mermaid)](#6-database-erd)
7. [Migration Procedures](#7-migration-procedures)
8. [Rollback Considerations](#8-rollback-considerations)

---

## 1. Migration Versioning Scheme

Catalogizer uses a **dual migration system**:

### 1.1 Go-Based Migrations (Primary)

The primary migration engine is embedded in the Go application code. Migrations are defined as functions in `database/migrations.go` and dispatched to dialect-specific implementations in `migrations_sqlite.go` and `migrations_postgres.go`. These run automatically at application startup.

Each migration is registered in the `RunMigrations()` function as a `Migration` struct:

```go
type Migration struct {
    Version int
    Name    string
    Up      func(context.Context) error
}
```

The `migrations` table tracks which versions have been applied:

| Column | Type | Description |
|--------|------|-------------|
| version | INTEGER (PK) | Migration version number |
| name | TEXT | Human-readable migration name |
| applied_at | DATETIME/TIMESTAMP | When the migration was applied |

Before executing a migration, the runner checks `SELECT COUNT(*) FROM migrations WHERE version = ?`. If the version already exists, the migration is skipped. This makes migrations idempotent and safe for repeated startup cycles.

### 1.2 SQL File Migrations (Reference/CLI)

SQL migration files in `database/migrations/` follow the naming convention:

```
{version}_{name}.up.sql          -- PostgreSQL up migration
{version}_{name}.down.sql        -- PostgreSQL/SQLite down migration
{version}_{name}.sqlite.up.sql   -- SQLite-specific up migration
```

These can be run via the `golang-migrate` CLI tool but are primarily used as reference. The Go-based migrations are the authoritative source.

### 1.3 Version Numbering

Versions are sequential integers starting at 1. The current latest version is **15**. There are no gaps in the version sequence.

---

## 2. Migration Registry

The full migration registry from `database/migrations.go`:

| Version | Name | Source File(s) | Description |
|---------|------|----------------|-------------|
| 1 | create_initial_tables | `migrations_sqlite.go`, `migrations_postgres.go` | Base schema: storage_roots, files, file_metadata, duplicate_groups, virtual_paths, scan_history |
| 2 | migrate_smb_to_storage_roots | `migrations_sqlite.go`, `migrations_postgres.go` | Data migration from legacy smb_roots table to unified storage_roots |
| 3 | create_auth_tables | `migrations_sqlite.go`, `migrations_postgres.go` | Users, roles, sessions, permissions, audit log; seeds Admin and User roles |
| 4 | create_conversion_jobs_table | `migrations_sqlite.go`, `migrations_postgres.go` | Media format conversion job tracking |
| 5 | create_subtitle_tables | `migrations_sqlite.go`, `migrations_postgres.go` | Subtitle tracks, sync status, cache, downloads, association table; SQLite triggers |
| 6 | fix_subtitle_foreign_keys | `migrations_sqlite.go`, `migrations_postgres.go` | Changes subtitle FKs from media_items to files (SQLite uses backup/recreate; PostgreSQL no-op) |
| 7 | create_assets_table | `migrations_sqlite.go`, `migrations_postgres.go` | Asset management table (covers, thumbnails, etc.) |
| 8 | create_media_entity_tables | `migrations_sqlite.go`, `migrations_postgres.go` | Media entity system: media_types (seeded with 11 types), media_items, media_files, collections, external_metadata, user_metadata, directory_analyses, detection_rules |
| 9 | create_performance_indexes | `migrations_v9_performance.go` | Performance-critical indexes on files, media_items, user_metadata; deduplicates media_files and creates unique compound index |
| 10 | create_sync_tables | `migrations_v10_sync_tables.go` | Sync endpoints, sessions, and schedules for remote synchronization |
| 11 | create_service_tables | `migrations_v11_service_tables.go` | Favorites, analytics, media access logs, error/crash reports, log management, cache tables, wizard progress |
| 12 | fix_service_table_columns | `migrations_v11_service_tables.go` | Adds missing columns to existing service tables (idempotent ALTER TABLE statements) |
| 13 | create_playlist_tables | `migrations_v13_playlist_tables.go` | User playlists and playlist items |
| 14 | create_additional_indexes | `migrations_v14_additional_indexes.go` | Additional indexes for files, file_metadata, analytics, scan_history, media_items, user_sessions |
| 15 | create_playback_session_tables | `migrations_v15_playback_sessions.go` | Playback sessions and media progress tracking with position units (seconds/pages/events) |

---

## 3. Dialect Abstraction Layer

The `database/dialect.go` file provides cross-database SQL compatibility through the `Dialect` struct. The `database.DB` wrapper shadows `Exec()`, `Query()`, and `QueryRow()` to automatically apply all rewrites before executing SQL.

### 3.1 Automatic Query Rewrites

| Rewrite | SQLite Input | PostgreSQL Output | Method |
|---------|-------------|-------------------|--------|
| Placeholders | `WHERE id = ?` | `WHERE id = $1` | `RewritePlaceholders()` |
| Insert or Ignore | `INSERT OR IGNORE INTO ...` | `INSERT INTO ... ON CONFLICT DO NOTHING` | `RewriteInsertOrIgnore()` |
| Insert or Replace | `INSERT OR REPLACE INTO ...` | `INSERT INTO ...` | `RewriteInsertOrReplace()` |
| Boolean literals | `is_active = 1` | `is_active = TRUE` | `RewriteBooleanLiterals()` |

### 3.2 Known Boolean Columns

The boolean rewriter recognizes these column names:

`is_active`, `is_locked`, `is_system`, `is_default`, `is_forced`, `is_duplicate`, `is_directory`, `deleted`, `enabled`, `verified_sync`, `is_favorite`, `is_public`, `is_smart`, `shuffle_enabled`, `hdr`, `dolby_vision`, `dolby_atmos`, `is_synced`

### 3.3 Type Differences

| Feature | PostgreSQL | SQLite |
|---------|-----------|--------|
| Auto-increment PK | `SERIAL PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| Boolean type | `BOOLEAN` (TRUE/FALSE) | `INTEGER` (0/1) |
| Timestamp type | `TIMESTAMP` | `DATETIME` |
| Large integer | `BIGINT` | `INTEGER` |
| JSON type | `JSONB` (v11 tables) | `TEXT` |
| ID generation | `INSERT ... RETURNING id` | `LastInsertId()` |

### 3.4 Helper Functions

- `InsertReturningID(ctx, query, args...)` -- abstracts ID retrieval after INSERT across dialects
- `TxInsertReturningID(tx, query, args...)` -- transaction-aware variant
- `TableExists(ctx, tableName)` -- checks for table existence using dialect-appropriate system catalogs
- `WrapDB(sqlDB, dialectType)` -- wraps a raw `*sql.DB` for testing (typically with in-memory SQLite)

---

## 4. Migration Details

### 4.1 Version 1: Initial Schema (create_initial_tables)

Creates the foundational tables for file cataloging.

#### Tables Created

**storage_roots** -- Storage endpoint definitions (SMB, FTP, NFS, WebDAV, local)

| Column | SQLite Type | PostgreSQL Type | Default | Notes |
|--------|------------|-----------------|---------|-------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- | |
| name | TEXT NOT NULL UNIQUE | TEXT NOT NULL UNIQUE | -- | |
| protocol | TEXT NOT NULL | TEXT NOT NULL | -- | smb, ftp, nfs, webdav, local |
| host | TEXT | TEXT | -- | |
| port | INTEGER | INTEGER | -- | |
| path | TEXT | TEXT | -- | |
| username | TEXT | TEXT | -- | |
| password | TEXT | TEXT | -- | |
| domain | TEXT | TEXT | -- | SMB domain |
| mount_point | TEXT | TEXT | -- | |
| options | TEXT | TEXT | -- | |
| url | TEXT | TEXT | -- | |
| enabled | BOOLEAN (0/1) | BOOLEAN | 1/TRUE | |
| max_depth | INTEGER | INTEGER | 10 | |
| enable_duplicate_detection | BOOLEAN (0/1) | BOOLEAN | 1/TRUE | |
| enable_metadata_extraction | BOOLEAN (0/1) | BOOLEAN | 1/TRUE | |
| include_patterns | TEXT | TEXT | -- | |
| exclude_patterns | TEXT | TEXT | -- | |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP | |
| last_scan_at | DATETIME | TIMESTAMP | -- | |

**files** -- Scanned file entries

| Column | SQLite Type | PostgreSQL Type | Default | Notes |
|--------|------------|-----------------|---------|-------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- | |
| storage_root_id | INTEGER NOT NULL | INTEGER NOT NULL | -- | FK -> storage_roots(id) |
| path | TEXT NOT NULL | TEXT NOT NULL | -- | |
| name | TEXT NOT NULL | TEXT NOT NULL | -- | |
| extension | TEXT | TEXT | -- | |
| mime_type | TEXT | TEXT | -- | |
| file_type | TEXT | TEXT | -- | |
| size | INTEGER NOT NULL | BIGINT NOT NULL | -- | |
| is_directory | BOOLEAN (0) | BOOLEAN (FALSE) | 0/FALSE | |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP | |
| modified_at | DATETIME NOT NULL | TIMESTAMP NOT NULL | -- | |
| accessed_at | DATETIME | TIMESTAMP | -- | |
| deleted | BOOLEAN (0) | BOOLEAN (FALSE) | 0/FALSE | |
| deleted_at | DATETIME | TIMESTAMP | -- | |
| last_scan_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP | |
| last_verified_at | DATETIME | TIMESTAMP | -- | |
| md5 | TEXT | TEXT | -- | |
| sha256 | TEXT | TEXT | -- | |
| sha1 | TEXT | TEXT | -- | |
| blake3 | TEXT | TEXT | -- | |
| quick_hash | TEXT | TEXT | -- | |
| is_duplicate | BOOLEAN (0) | BOOLEAN (FALSE) | 0/FALSE | |
| duplicate_group_id | INTEGER | INTEGER | -- | FK -> duplicate_groups(id) |
| parent_id | INTEGER | INTEGER | -- | FK -> files(id), self-ref |

Unique constraint: `UNIQUE(storage_root_id, path)` (enforced via unique index)

**duplicate_groups** -- Groups of duplicate files

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| file_count | INTEGER | INTEGER | 0 |
| total_size | INTEGER | BIGINT | 0 |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

**file_metadata** -- Key-value metadata for files

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| file_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| key | TEXT NOT NULL | TEXT NOT NULL | -- |
| value | TEXT NOT NULL | TEXT NOT NULL | -- |
| data_type | TEXT | TEXT | 'string' |

FK: `file_id -> files(id) ON DELETE CASCADE`

**virtual_paths** -- Virtual path mappings

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| path | TEXT NOT NULL UNIQUE | TEXT NOT NULL UNIQUE | -- |
| target_type | TEXT NOT NULL | TEXT NOT NULL | -- |
| target_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

**scan_history** -- Scan execution records

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| storage_root_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| scan_type | TEXT NOT NULL | TEXT NOT NULL | -- |
| status | TEXT NOT NULL | TEXT NOT NULL | -- |
| start_time | DATETIME NOT NULL | TIMESTAMP NOT NULL | -- |
| end_time | DATETIME | TIMESTAMP | -- |
| files_processed | INTEGER | INTEGER | 0 |
| files_added | INTEGER | INTEGER | 0 |
| files_updated | INTEGER | INTEGER | 0 |
| files_deleted | INTEGER | INTEGER | 0 |
| error_count | INTEGER | INTEGER | 0 |
| error_message | TEXT | TEXT | -- |

FK: `storage_root_id -> storage_roots(id)`

#### Indexes Created (v1)

| Index | Table | Column(s) | Type |
|-------|-------|-----------|------|
| idx_files_storage_root_path | files | (storage_root_id, path) | UNIQUE |
| idx_files_parent_id | files | parent_id | BTREE |
| idx_files_duplicate_group | files | duplicate_group_id | BTREE |
| idx_files_deleted | files | deleted | BTREE |
| idx_file_metadata_file_id | file_metadata | file_id | BTREE |
| idx_scan_history_storage_root | scan_history | storage_root_id | BTREE |

---

### 4.2 Version 2: SMB Migration (migrate_smb_to_storage_roots)

A **data migration** (no DDL). Migrates rows from the legacy `smb_roots` table into `storage_roots` with `protocol = 'smb'`. Updates `files.storage_root_id` and `scan_history.storage_root_id` to point to the new storage_roots rows. Skips gracefully if `smb_roots` does not exist.

---

### 4.3 Version 3: Authentication Tables (create_auth_tables)

#### Tables Created

**users** -- User accounts

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| username | TEXT NOT NULL UNIQUE | TEXT NOT NULL UNIQUE | -- |
| email | TEXT NOT NULL UNIQUE | TEXT NOT NULL UNIQUE | -- |
| password_hash | TEXT NOT NULL | TEXT NOT NULL | -- |
| salt | TEXT NOT NULL | TEXT NOT NULL | -- |
| role_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| first_name | TEXT | TEXT | -- |
| last_name | TEXT | TEXT | -- |
| display_name | TEXT | TEXT | -- |
| avatar_url | TEXT | TEXT | -- |
| time_zone | TEXT | TEXT | -- |
| language | TEXT | TEXT | -- |
| settings | TEXT | TEXT | '{}' |
| is_active | INTEGER (1) | BOOLEAN (TRUE) | 1/TRUE |
| is_locked | INTEGER (0) | BOOLEAN (FALSE) | 0/FALSE |
| locked_until | DATETIME | TIMESTAMP | -- |
| failed_login_attempts | INTEGER | INTEGER | 0 |
| last_login_at | DATETIME | TIMESTAMP | -- |
| last_login_ip | TEXT | TEXT | -- |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

**roles** -- Role definitions

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| name | TEXT NOT NULL UNIQUE | TEXT NOT NULL UNIQUE | -- |
| description | TEXT | TEXT | -- |
| permissions | TEXT | TEXT | '[]' |
| is_system | INTEGER (0) | BOOLEAN (FALSE) | 0/FALSE |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

Seeded data:
- Role 1: Admin -- permissions `["*"]`, is_system=true
- Role 2: User -- permissions `["media.view", "media.download"]`, is_system=true

PostgreSQL resets the `roles_id_seq` sequence after seeding.

**user_sessions** -- Active sessions

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| session_token | TEXT NOT NULL UNIQUE | TEXT NOT NULL UNIQUE | -- |
| refresh_token | TEXT | TEXT | -- |
| device_info | TEXT | TEXT | -- |
| ip_address | TEXT | TEXT | -- |
| user_agent | TEXT | TEXT | -- |
| is_active | INTEGER (1) | BOOLEAN (TRUE) | 1/TRUE |
| expires_at | DATETIME NOT NULL | TIMESTAMP NOT NULL | -- |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| last_activity_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

FK: `user_id -> users(id) ON DELETE CASCADE`

**permissions** -- Permission definitions

| Column | Type | Notes |
|--------|------|-------|
| id | PK | Auto-increment |
| name | TEXT NOT NULL UNIQUE | |
| resource | TEXT NOT NULL | |
| action | TEXT NOT NULL | |
| description | TEXT | |

**user_permissions** -- User-permission junction (composite PK)

| Column | Type | Notes |
|--------|------|-------|
| user_id | INTEGER NOT NULL | FK -> users(id) ON DELETE CASCADE |
| permission_id | INTEGER NOT NULL | FK -> permissions(id) ON DELETE CASCADE |
| granted_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| granted_by | INTEGER | FK -> users(id) |

**auth_audit_log** -- Authentication event logging

| Column | Type | Notes |
|--------|------|-------|
| id | PK | Auto-increment |
| user_id | INTEGER | FK -> users(id) |
| event_type | TEXT NOT NULL | |
| ip_address | TEXT | |
| user_agent | TEXT | |
| details | TEXT | |
| created_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |

#### Indexes Created (v3)

| Index | Table | Column(s) |
|-------|-------|-----------|
| idx_users_username | users | username |
| idx_users_email | users | email |
| idx_users_role_id | users | role_id |
| idx_users_is_active | users | is_active |
| idx_user_sessions_user_id | user_sessions | user_id |
| idx_user_sessions_token | user_sessions | session_token |
| idx_user_sessions_expires_at | user_sessions | expires_at |

---

### 4.4 Version 4: Conversion Jobs (create_conversion_jobs_table)

**conversion_jobs** -- Media format conversion tracking

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| source_path | TEXT NOT NULL | TEXT NOT NULL | -- |
| target_path | TEXT NOT NULL | TEXT NOT NULL | -- |
| source_format | TEXT NOT NULL | TEXT NOT NULL | -- |
| target_format | TEXT NOT NULL | TEXT NOT NULL | -- |
| conversion_type | TEXT NOT NULL | TEXT NOT NULL | -- |
| quality | TEXT | TEXT | 'medium' |
| settings | TEXT | TEXT | -- |
| priority | INTEGER | INTEGER | 0 |
| status | TEXT | TEXT | 'pending' |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| started_at | DATETIME | TIMESTAMP | -- |
| completed_at | DATETIME | TIMESTAMP | -- |
| scheduled_for | DATETIME | TIMESTAMP | -- |
| duration | INTEGER | INTEGER | -- |
| error_message | TEXT | TEXT | -- |

FK: `user_id -> users(id) ON DELETE CASCADE`

Indexes: `idx_conversion_jobs_user_id`, `idx_conversion_jobs_status`, `idx_conversion_jobs_created_at`

---

### 4.5 Version 5: Subtitle Tables (create_subtitle_tables)

#### Tables Created

**subtitle_tracks** -- Subtitle track definitions

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| id | PK | -- | |
| media_item_id | INTEGER NOT NULL | -- | FK -> media_items(id) ON DELETE CASCADE |
| language | TEXT NOT NULL | -- | |
| language_code | TEXT NOT NULL | -- | |
| source | TEXT NOT NULL | 'downloaded' | |
| format | TEXT NOT NULL | 'srt' | |
| path | TEXT | -- | |
| content | TEXT | -- | |
| is_default | BOOLEAN | FALSE | |
| is_forced | BOOLEAN | FALSE | |
| encoding | TEXT | 'utf-8' | |
| sync_offset | REAL | 0.0 | |
| verified_sync | BOOLEAN | FALSE | |
| created_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP | |
| updated_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP | |

**subtitle_sync_status** -- Sync operation tracking

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| id | PK | -- | |
| media_item_id | INTEGER NOT NULL | -- | FK -> media_items(id) |
| subtitle_id | TEXT NOT NULL | -- | |
| operation | TEXT NOT NULL | -- | 'download', 'upload', 'sync', 'verify' |
| status | TEXT NOT NULL | 'pending' | 'pending', 'in_progress', 'completed', 'failed' |
| progress | INTEGER | 0 | |
| error_message | TEXT | -- | |
| created_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP | |
| updated_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP | |
| completed_at | DATETIME/TIMESTAMP | -- | |

**subtitle_cache** -- Search result caching

| Column | Type | Default |
|--------|------|---------|
| id | PK | -- |
| cache_key | TEXT UNIQUE NOT NULL | -- |
| result_id | TEXT NOT NULL | -- |
| provider | TEXT NOT NULL | -- |
| title | TEXT | -- |
| language | TEXT | -- |
| language_code | TEXT | -- |
| download_url | TEXT | -- |
| format | TEXT | -- |
| encoding | TEXT | -- |
| upload_date | DATETIME/TIMESTAMP | -- |
| downloads | INTEGER | -- |
| rating | REAL | -- |
| comments | INTEGER | -- |
| match_score | REAL | -- |
| created_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| expires_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| data | TEXT | -- |

**subtitle_downloads** -- Download history

| Column | Type | Default |
|--------|------|---------|
| id | PK | -- |
| media_item_id | INTEGER NOT NULL | -- |
| result_id | TEXT NOT NULL | -- |
| subtitle_id | TEXT NOT NULL | -- |
| provider | TEXT NOT NULL | -- |
| language | TEXT NOT NULL | -- |
| file_path | TEXT | -- |
| file_size | INTEGER | -- |
| download_url | TEXT | -- |
| download_date | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| verified_sync | BOOLEAN | FALSE |
| sync_offset | REAL | 0.0 |

**media_subtitles** -- Many-to-many junction

| Column | Type | Default |
|--------|------|---------|
| id | PK | -- |
| media_item_id | INTEGER NOT NULL | -- |
| subtitle_track_id | INTEGER NOT NULL | -- |
| is_active | BOOLEAN | TRUE |
| added_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |

Unique constraint: `UNIQUE(media_item_id, subtitle_track_id)`

#### Triggers (SQLite)

- `update_subtitle_tracks_updated_at` -- Updates `updated_at` after any UPDATE on subtitle_tracks
- `update_subtitle_sync_status_updated_at` -- Updates `updated_at` after any UPDATE on subtitle_sync_status
- `set_subtitle_sync_status_completed_at` -- Sets `completed_at` when status changes to 'completed'

PostgreSQL uses trigger functions (`BEFORE UPDATE ... EXECUTE FUNCTION`) instead of `AFTER UPDATE BEGIN...END` blocks.

---

### 4.6 Version 6: Fix Subtitle Foreign Keys (fix_subtitle_foreign_keys)

Changes FK references in all subtitle tables from `media_items(id)` to `files(id)`.

- **SQLite**: Uses the backup/recreate pattern (CREATE backup -> DROP original -> CREATE new with correct FK -> INSERT from backup -> DROP backup -> recreate triggers)
- **PostgreSQL**: No-op (tables were created correctly in the first place, or are fresh installations)

---

### 4.7 Version 7: Assets Table (create_assets_table)

**assets** -- Managed assets (covers, thumbnails, etc.)

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | TEXT PK | TEXT PK | -- |
| type | TEXT NOT NULL | TEXT NOT NULL | -- |
| status | TEXT NOT NULL | TEXT NOT NULL | 'pending' |
| content_type | TEXT | TEXT | -- |
| size | INTEGER | BIGINT | 0 |
| source_hint | TEXT | TEXT | -- |
| entity_type | TEXT | TEXT | -- |
| entity_id | TEXT | TEXT | -- |
| metadata | TEXT | TEXT | -- |
| local_path | TEXT | TEXT | -- |
| created_at | TIMESTAMP | TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | TIMESTAMP | TIMESTAMP | CURRENT_TIMESTAMP |
| resolved_at | TIMESTAMP | TIMESTAMP | -- |
| expires_at | TIMESTAMP | TIMESTAMP | -- |

Note: The `id` column is `TEXT`, not auto-increment (assets use generated UUIDs).

Indexes: `idx_assets_entity(entity_type, entity_id)`, `idx_assets_status(status)`

---

### 4.8 Version 8: Media Entity Tables (create_media_entity_tables)

The core media entity system that structures scanned files into recognized media items.

#### Tables Created

**media_types** -- 11 seeded media type definitions

| Column | Type | Default |
|--------|------|---------|
| id | PK | Auto-increment |
| name | TEXT NOT NULL UNIQUE | -- |
| description | TEXT | -- |
| detection_patterns | TEXT | -- |
| metadata_providers | TEXT | -- |
| created_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |

Seeded values: `movie`, `tv_show`, `tv_season`, `tv_episode`, `music_artist`, `music_album`, `song`, `game`, `software`, `book`, `comic`

**media_items** -- Core entity table (self-referential hierarchy)

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| id | PK | Auto-increment | |
| media_type_id | INTEGER NOT NULL | -- | FK -> media_types(id) |
| title | TEXT NOT NULL | -- | |
| original_title | TEXT | -- | |
| year | INTEGER | -- | |
| description | TEXT | -- | |
| genre | TEXT | -- | |
| director | TEXT | -- | |
| cast_crew | TEXT | -- | |
| rating | REAL | -- | |
| runtime | INTEGER | -- | |
| language | TEXT | -- | |
| country | TEXT | -- | |
| status | TEXT NOT NULL | 'detected' | |
| parent_id | INTEGER | -- | FK -> media_items(id) ON DELETE CASCADE |
| season_number | INTEGER | -- | |
| episode_number | INTEGER | -- | |
| track_number | INTEGER | -- | |
| first_detected | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP | |
| last_updated | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP | |

**media_files** -- Junction table linking files to media items

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| media_item_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| file_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| quality_info | TEXT | TEXT | -- |
| language | TEXT | TEXT | -- |
| is_primary | INTEGER (0) | BOOLEAN (FALSE) | 0/FALSE |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

FKs: `media_item_id -> media_items(id) ON DELETE CASCADE`, `file_id -> files(id) ON DELETE CASCADE`

**media_collections** -- Named collections

| Column | Type | Default |
|--------|------|---------|
| id | PK | Auto-increment |
| name | TEXT NOT NULL | -- |
| collection_type | TEXT NOT NULL | -- |
| description | TEXT | -- |
| total_items | INTEGER | 0 |
| external_ids | TEXT | -- |
| cover_url | TEXT | -- |
| created_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |

**media_collection_items** -- Collection membership

| Column | Type | Default |
|--------|------|---------|
| id | PK | Auto-increment |
| collection_id | INTEGER NOT NULL | -- |
| media_item_id | INTEGER NOT NULL | -- |
| sequence_number | INTEGER | -- |
| season_number | INTEGER | -- |
| release_order | INTEGER | -- |

FKs: `collection_id -> media_collections(id) ON DELETE CASCADE`, `media_item_id -> media_items(id) ON DELETE CASCADE`

**external_metadata** -- Third-party metadata (TMDB, OMDB, OpenLibrary, MusicBrainz)

| Column | Type | Default |
|--------|------|---------|
| id | PK | Auto-increment |
| media_item_id | INTEGER NOT NULL | -- |
| provider | TEXT NOT NULL | -- |
| external_id | TEXT NOT NULL | -- |
| data | TEXT | -- |
| rating | REAL | -- |
| review_url | TEXT | -- |
| cover_url | TEXT | -- |
| trailer_url | TEXT | -- |
| last_fetched | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |

FK: `media_item_id -> media_items(id) ON DELETE CASCADE`

**user_metadata** -- Per-user media metadata (ratings, watched status)

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| media_item_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| user_rating | REAL | REAL | -- |
| watched_status | TEXT | TEXT | -- |
| watched_date | DATETIME | TIMESTAMP | -- |
| personal_notes | TEXT | TEXT | -- |
| tags | TEXT | TEXT | -- |
| favorite | INTEGER (0) | BOOLEAN (FALSE) | 0/FALSE |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

FKs: `media_item_id -> media_items(id) ON DELETE CASCADE`, `user_id -> users(id) ON DELETE CASCADE`

**directory_analyses** -- Directory-level media detection results

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| directory_path | TEXT NOT NULL | TEXT NOT NULL | -- |
| smb_root | TEXT | TEXT | -- |
| media_item_id | INTEGER | INTEGER | -- |
| confidence_score | REAL | REAL | 0 |
| detection_method | TEXT | TEXT | -- |
| analysis_data | TEXT | TEXT | -- |
| last_analyzed | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| files_count | INTEGER | INTEGER | 0 |
| total_size | INTEGER | BIGINT | 0 |

FK: `media_item_id -> media_items(id) ON DELETE SET NULL`

**detection_rules** -- Configurable detection patterns

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| media_type_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| rule_name | TEXT NOT NULL | TEXT NOT NULL | -- |
| rule_type | TEXT NOT NULL | TEXT NOT NULL | -- |
| pattern | TEXT NOT NULL | TEXT NOT NULL | -- |
| confidence_weight | REAL | REAL | 1.0 |
| enabled | INTEGER (1) | BOOLEAN (TRUE) | 1/TRUE |
| priority | INTEGER | INTEGER | 0 |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

FK: `media_type_id -> media_types(id) ON DELETE CASCADE`

#### Indexes Created (v8)

| Index | Table | Column(s) |
|-------|-------|-----------|
| idx_media_items_type | media_items | media_type_id |
| idx_media_items_parent | media_items | parent_id |
| idx_media_items_title | media_items | title |
| idx_media_files_item | media_files | media_item_id |
| idx_media_files_file | media_files | file_id |
| idx_external_metadata_item | external_metadata | media_item_id |
| idx_external_metadata_provider | external_metadata | (provider, external_id) |
| idx_user_metadata_item | user_metadata | media_item_id |
| idx_user_metadata_user | user_metadata | user_id |
| idx_directory_analyses_path | directory_analyses | directory_path |
| idx_detection_rules_type | detection_rules | media_type_id |
| idx_media_collection_items_collection | media_collection_items | collection_id |
| idx_media_collection_items_item | media_collection_items | media_item_id |

---

### 4.9 Version 9: Performance Indexes (create_performance_indexes)

Adds performance-critical indexes identified from repository query patterns. Also deduplicates `media_files` rows before creating a unique compound index.

#### Indexes Added

| Index | Table | Column(s) | Type |
|-------|-------|-----------|------|
| idx_files_file_type | files | file_type | BTREE |
| idx_files_extension | files | extension | BTREE |
| idx_files_is_directory | files | is_directory | BTREE |
| idx_files_name | files | name | BTREE |
| idx_media_items_title_type | media_items | (title, media_type_id) | COMPOUND |
| idx_media_items_status | media_items | status | BTREE |
| idx_media_items_year | media_items | year | BTREE |
| idx_user_metadata_user_watched | user_metadata | (user_id, watched_status) | COMPOUND |
| idx_media_files_item_file | media_files | (media_item_id, file_id) | UNIQUE |

#### Dialect Difference

- **SQLite**: Deduplicates using `DELETE FROM media_files WHERE rowid NOT IN (SELECT MIN(rowid) ...)`
- **PostgreSQL**: Deduplicates using `DELETE FROM media_files a USING media_files b WHERE a.ctid > b.ctid AND ...`

---

### 4.10 Version 10: Sync Tables (create_sync_tables)

#### Tables Created

**sync_endpoints** -- Remote sync endpoint definitions

| Column | Type | Default |
|--------|------|---------|
| id | PK | Auto-increment |
| user_id | INTEGER NOT NULL | -- |
| name | TEXT NOT NULL | -- |
| type | TEXT NOT NULL | -- |
| url | TEXT NOT NULL | -- |
| username | TEXT | -- |
| password | TEXT | -- |
| sync_direction | TEXT | 'bidirectional' |
| local_path | TEXT | -- |
| remote_path | TEXT | -- |
| sync_settings | TEXT | -- |
| status | TEXT | 'active' |
| created_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| last_sync_at | DATETIME/TIMESTAMP | -- |

FK: `user_id -> users(id)`

**sync_sessions** -- Individual sync execution records

| Column | Type | Default |
|--------|------|---------|
| id | PK | Auto-increment |
| endpoint_id | INTEGER NOT NULL | -- |
| user_id | INTEGER NOT NULL | -- |
| status | TEXT | 'running' |
| sync_type | TEXT | -- |
| started_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |
| completed_at | DATETIME/TIMESTAMP | -- |
| duration | INTEGER | -- |
| total_files | INTEGER | 0 |
| synced_files | INTEGER | 0 |
| failed_files | INTEGER | 0 |
| skipped_files | INTEGER | 0 |
| error_message | TEXT | -- |
| updated_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |

FKs: `endpoint_id -> sync_endpoints(id)`, `user_id -> users(id)`

**sync_schedules** -- Recurring sync configuration

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| endpoint_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| frequency | TEXT NOT NULL | TEXT NOT NULL | -- |
| last_run | DATETIME | TIMESTAMP | -- |
| next_run | DATETIME | TIMESTAMP | -- |
| is_active | BOOLEAN (1) | BOOLEAN (TRUE) | 1/TRUE |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

#### Indexes Created (v10)

11 indexes covering user_id, status, type, endpoint_id, started_at, is_active, next_run across all three sync tables.

---

### 4.11 Version 11: Service Tables (create_service_tables)

Creates tables for analytics, favorites, error/crash reporting, log management, caching, and wizard progress.

#### Tables Created

**favorites** -- User favorites with categories

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| entity_type | TEXT NOT NULL | TEXT NOT NULL | -- |
| entity_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| category | TEXT | TEXT | '' |
| notes | TEXT | TEXT | '' |
| tags | TEXT | TEXT | -- |
| is_public | INTEGER (0) | BOOLEAN (FALSE) | 0/FALSE |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

Unique constraint: `UNIQUE(user_id, entity_type, entity_id)`

**favorite_categories** -- Custom favorite categories per user

**analytics_events** -- User activity events (uses `JSONB` for event_data and device_info in PostgreSQL, `TEXT` in SQLite)

**media_access_logs** -- Media access tracking

**error_reports** -- Client error reporting

**crash_reports** -- Application crash reporting

**log_collections** -- Log collection sessions (uses `JSONB` for filters in PostgreSQL)

**log_shares** -- Shared log collection access tokens

**cache_entries** -- General cache storage

**api_cache** -- API response cache

**media_metadata_cache** -- Metadata provider cache

**analytics_reports** -- Generated analytics reports (uses `JSONB` for report_data in PostgreSQL)

**wizard_progress** -- Configuration wizard state tracking (uses `JSONB` for step_data and all_data in PostgreSQL)

---

### 4.12 Version 12: Fix Service Table Columns (fix_service_table_columns)

Adds missing columns to service tables created before v11 had complete column definitions. Uses `ALTER TABLE ADD COLUMN` with graceful error handling -- "duplicate column" or "already exists" errors are silently ignored. This ensures idempotency for both fresh and upgraded databases.

Affected tables: `error_reports`, `crash_reports`, `log_shares`, `wizard_progress`, `media_access_logs`, `analytics_events`, `log_collections`, `favorites`

---

### 4.13 Version 13: Playlist Tables (create_playlist_tables)

**playlists** -- User-created playlists

| Column | SQLite Type | PostgreSQL Type | Default |
|--------|------------|-----------------|---------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- |
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- |
| name | TEXT NOT NULL | TEXT NOT NULL | -- |
| description | TEXT | TEXT | '' |
| is_public | INTEGER (0) | BOOLEAN (FALSE) | 0/FALSE |
| created_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |
| updated_at | DATETIME | TIMESTAMP | CURRENT_TIMESTAMP |

FK: `user_id -> users(id)`

**playlist_items** -- Items within playlists

| Column | Type | Default |
|--------|------|---------|
| id | PK | Auto-increment |
| playlist_id | INTEGER NOT NULL | -- |
| entity_id | INTEGER NOT NULL | -- |
| entity_type | TEXT NOT NULL | -- |
| position | INTEGER | 0 |
| added_at | DATETIME/TIMESTAMP | CURRENT_TIMESTAMP |

FK: `playlist_id -> playlists(id) ON DELETE CASCADE`

Indexes: `idx_playlists_user_id`, `idx_playlist_items_playlist_id`, `idx_playlist_items_entity(entity_id, entity_type)`

---

### 4.14 Version 14: Additional Indexes (create_additional_indexes)

Adds indexes for time-series queries, common filter patterns, and cleanup operations.

| Index | Table | Column(s) |
|-------|-------|-----------|
| idx_files_created_at | files | created_at |
| idx_files_modified_at | files | modified_at |
| idx_files_size | files | size |
| idx_files_is_directory | files | is_directory |
| idx_files_storage_root_created | files | (storage_root_id, created_at) |
| idx_file_metadata_key | file_metadata | key |
| idx_file_metadata_key_value | file_metadata | (key, value) |
| idx_analytics_events_time | analytics_events | timestamp |
| idx_analytics_events_user | analytics_events | (user_id, timestamp) |
| idx_analytics_events_type | analytics_events | (event_type, timestamp) |
| idx_scan_history_start_time | scan_history | start_time |
| idx_scan_history_status | scan_history | status |
| idx_media_items_type | media_items | media_type_id |
| idx_media_items_title | media_items | title |
| idx_media_items_status | media_items | status |
| idx_user_sessions_is_active | user_sessions | is_active |

---

### 4.15 Version 15: Playback Session Tables (create_playback_session_tables)

Tracks media reproduction history and progress across all media types.

**playback_sessions** -- Individual playback events

| Column | SQLite Type | PostgreSQL Type | Default | Notes |
|--------|------------|-----------------|---------|-------|
| id | INTEGER PK AUTOINCREMENT | SERIAL PK | -- | |
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- | |
| media_item_id | INTEGER NOT NULL | INTEGER NOT NULL | -- | FK -> media_items(id) ON DELETE CASCADE |
| file_id | INTEGER | INTEGER | -- | |
| started_at | DATETIME NOT NULL | TIMESTAMP NOT NULL | -- | |
| ended_at | DATETIME | TIMESTAMP | -- | |
| position_unit | TEXT NOT NULL | TEXT NOT NULL | -- | CHECK IN ('seconds','pages','events') |
| start_position | INTEGER NOT NULL | BIGINT NOT NULL | 0 | |
| end_position | INTEGER | BIGINT | -- | |
| total_amount | INTEGER NOT NULL | BIGINT NOT NULL | 0 | |
| completed | INTEGER NOT NULL | BOOLEAN NOT NULL | 0/FALSE | |

**media_progress** -- Denormalized latest progress snapshot (composite PK)

| Column | SQLite Type | PostgreSQL Type | Default | Notes |
|--------|------------|-----------------|---------|-------|
| user_id | INTEGER NOT NULL | INTEGER NOT NULL | -- | Part of composite PK |
| media_item_id | INTEGER NOT NULL | INTEGER NOT NULL | -- | Part of composite PK, FK -> media_items(id) ON DELETE CASCADE |
| position_unit | TEXT NOT NULL | TEXT NOT NULL | -- | |
| duration_total | INTEGER | BIGINT | -- | |
| last_position | INTEGER NOT NULL | BIGINT NOT NULL | 0 | |
| last_session_amount | INTEGER NOT NULL | BIGINT NOT NULL | 0 | |
| total_reproductions | INTEGER NOT NULL | BIGINT NOT NULL | 0 | |
| aggregate_amount | INTEGER NOT NULL | BIGINT NOT NULL | 0 | |
| last_session_ended_at | DATETIME | TIMESTAMP | -- | |
| updated_at | DATETIME NOT NULL | TIMESTAMP NOT NULL | -- | |

Position units: `seconds` (video/audio), `pages` (books/comics), `events` (games/software launches)

Indexes: `idx_playback_sessions_item(media_item_id, started_at DESC)`, `idx_playback_sessions_user(user_id, started_at DESC)`

---

## 5. Current Complete Schema Summary

### 5.1 All Tables (37 total)

| # | Table | Created in | Purpose |
|---|-------|-----------|---------|
| 1 | migrations | Bootstrap | Migration version tracking |
| 2 | storage_roots | v1 | Storage endpoint definitions |
| 3 | files | v1 | Scanned file entries |
| 4 | duplicate_groups | v1 | Duplicate file groups |
| 5 | file_metadata | v1 | Key-value file metadata |
| 6 | virtual_paths | v1 | Virtual path mappings |
| 7 | scan_history | v1 | Scan execution records |
| 8 | users | v3 | User accounts |
| 9 | roles | v3 | Role definitions (Admin, User) |
| 10 | user_sessions | v3 | Active sessions |
| 11 | permissions | v3 | Permission definitions |
| 12 | user_permissions | v3 | User-permission junction |
| 13 | auth_audit_log | v3 | Authentication events |
| 14 | conversion_jobs | v4 | Media format conversion |
| 15 | subtitle_tracks | v5/v6 | Subtitle definitions |
| 16 | subtitle_sync_status | v5/v6 | Subtitle sync operations |
| 17 | subtitle_cache | v5 | Subtitle search cache |
| 18 | subtitle_downloads | v5/v6 | Subtitle download history |
| 19 | media_subtitles | v5/v6 | Media-subtitle junction |
| 20 | assets | v7 | Managed assets (UUID PK) |
| 21 | media_types | v8 | 11 seeded media types |
| 22 | media_items | v8 | Core media entities |
| 23 | media_files | v8/v9 | File-entity junction |
| 24 | media_collections | v8 | Named collections |
| 25 | media_collection_items | v8 | Collection membership |
| 26 | external_metadata | v8 | Provider metadata |
| 27 | user_metadata | v8 | User ratings/watched status |
| 28 | directory_analyses | v8 | Directory detection results |
| 29 | detection_rules | v8 | Configurable detection patterns |
| 30 | sync_endpoints | v10 | Remote sync endpoints |
| 31 | sync_sessions | v10 | Sync execution records |
| 32 | sync_schedules | v10 | Recurring sync config |
| 33 | favorites | v11 | User favorites |
| 34 | favorite_categories | v11 | Custom favorite categories |
| 35 | analytics_events | v11 | User activity events |
| 36 | media_access_logs | v11 | Media access tracking |
| 37 | error_reports | v11 | Client error reports |
| 38 | crash_reports | v11 | Application crash reports |
| 39 | log_collections | v11 | Log collection sessions |
| 40 | log_shares | v11 | Shared log access tokens |
| 41 | cache_entries | v11 | General cache |
| 42 | api_cache | v11 | API response cache |
| 43 | media_metadata_cache | v11 | Metadata provider cache |
| 44 | analytics_reports | v11 | Generated reports |
| 45 | wizard_progress | v11 | Config wizard state |
| 46 | playlists | v13 | User playlists |
| 47 | playlist_items | v13 | Playlist items |
| 48 | playback_sessions | v15 | Playback events |
| 49 | media_progress | v15 | Latest progress snapshot |

### 5.2 Key Relationships

- **files** -> storage_roots (many-to-one via storage_root_id)
- **files** -> files (self-referential via parent_id)
- **files** -> duplicate_groups (many-to-one via duplicate_group_id)
- **media_items** -> media_types (many-to-one via media_type_id)
- **media_items** -> media_items (self-referential hierarchy via parent_id: show->season->episode, artist->album->song)
- **media_files** -> media_items + files (junction table, many-to-many)
- **media_collection_items** -> media_collections + media_items (junction)
- **external_metadata** -> media_items (many-to-one)
- **user_metadata** -> media_items + users (per-user per-media)
- **subtitle_tracks** -> files (many-to-one via media_item_id, post v6 FK fix)
- **media_subtitles** -> media_items + subtitle_tracks (junction)
- **users** -> roles (many-to-one via role_id)
- **user_sessions** -> users (many-to-one)
- **user_permissions** -> users + permissions (junction)
- **conversion_jobs** -> users (many-to-one)
- **sync_endpoints** -> users (many-to-one)
- **sync_sessions** -> sync_endpoints + users
- **sync_schedules** -> sync_endpoints + users
- **playlists** -> users (many-to-one)
- **playlist_items** -> playlists (many-to-one)
- **playback_sessions** -> media_items (many-to-one)
- **media_progress** -> media_items (composite PK: user_id + media_item_id)
- **log_shares** -> log_collections (many-to-one)

---

## 6. Database ERD

```mermaid
erDiagram
    storage_roots ||--o{ files : "contains"
    files ||--o{ files : "parent_id"
    files ||--o{ file_metadata : "has"
    duplicate_groups ||--o{ files : "groups"

    media_types ||--o{ media_items : "classifies"
    media_items ||--o{ media_items : "parent_id (hierarchy)"
    media_items ||--o{ media_files : "linked via"
    files ||--o{ media_files : "linked via"

    media_collections ||--o{ media_collection_items : "contains"
    media_items ||--o{ media_collection_items : "member of"

    media_items ||--o{ external_metadata : "enriched by"
    media_items ||--o{ user_metadata : "rated by"
    users ||--o{ user_metadata : "rates"

    media_items ||--o{ directory_analyses : "detected in"
    media_types ||--o{ detection_rules : "configured by"

    files ||--o{ subtitle_tracks : "has subtitles"
    media_items ||--o{ media_subtitles : "associated"
    subtitle_tracks ||--o{ media_subtitles : "associated"
    media_items ||--o{ subtitle_sync_status : "sync status"
    media_items ||--o{ subtitle_downloads : "downloaded"

    roles ||--o{ users : "assigned to"
    users ||--o{ user_sessions : "has sessions"
    users ||--o{ user_permissions : "granted"
    permissions ||--o{ user_permissions : "assigned"
    users ||--o{ auth_audit_log : "audited"

    users ||--o{ conversion_jobs : "owns"

    users ||--o{ sync_endpoints : "configures"
    sync_endpoints ||--o{ sync_sessions : "executes"
    sync_endpoints ||--o{ sync_schedules : "scheduled"
    users ||--o{ sync_sessions : "runs"

    users ||--o{ favorites : "favorites"
    users ||--o{ favorite_categories : "categorizes"

    users ||--o{ playlists : "creates"
    playlists ||--o{ playlist_items : "contains"

    media_items ||--o{ playback_sessions : "played in"
    media_items ||--o{ media_progress : "progress tracked"

    users ||--o{ analytics_events : "generates"
    users ||--o{ error_reports : "reports"
    users ||--o{ crash_reports : "reports"
    users ||--o{ log_collections : "collects"
    log_collections ||--o{ log_shares : "shared via"

    storage_roots {
        int id PK
        text name UK
        text protocol
        text host
        int port
        text path
        boolean enabled
        int max_depth
    }

    files {
        int id PK
        int storage_root_id FK
        text path
        text name
        text extension
        bigint size
        boolean is_directory
        boolean deleted
        int parent_id FK
        int duplicate_group_id FK
    }

    media_types {
        int id PK
        text name UK
        text description
    }

    media_items {
        int id PK
        int media_type_id FK
        text title
        int year
        text status
        int parent_id FK
        int season_number
        int episode_number
        int track_number
    }

    media_files {
        int id PK
        int media_item_id FK
        int file_id FK
        boolean is_primary
    }

    users {
        int id PK
        text username UK
        text email UK
        text password_hash
        int role_id FK
        boolean is_active
        boolean is_locked
    }

    roles {
        int id PK
        text name UK
        text permissions
        boolean is_system
    }

    playlists {
        int id PK
        int user_id FK
        text name
        boolean is_public
    }

    playback_sessions {
        int id PK
        int user_id
        int media_item_id FK
        text position_unit
        bigint start_position
        bigint end_position
        boolean completed
    }

    media_progress {
        int user_id PK
        int media_item_id PK
        bigint last_position
        bigint total_reproductions
    }
```

---

## 7. Migration Procedures

### 7.1 Automatic Migrations (Application Startup)

Migrations run automatically when the application starts via `db.RunMigrations(ctx)` called from `main.go`. No manual intervention is required for standard deployments.

```bash
# Start the API server -- migrations run automatically
cd catalog-api && go run main.go
```

### 7.2 Manual Migration via golang-migrate CLI

Install the CLI tool:

```bash
go install -tags 'postgres sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Run migrations against PostgreSQL:

```bash
migrate -path database/migrations \
  -database "postgres://catalogizer:password@localhost:5433/catalogizer?sslmode=disable" up
```

Run migrations against SQLite:

```bash
migrate -path database/migrations \
  -database "sqlite3://./catalogizer.db" up
```

Check current version:

```bash
migrate -path database/migrations \
  -database "postgres://..." version
```

### 7.3 Creating a New Migration

1. Determine the next version number (currently 15, so next is 16)
2. Add the migration function to the appropriate Go files:
   - Create `migrations_v16_<name>.go` with both `create<Name>SQLite()` and `create<Name>Postgres()` functions
   - Register in `migrations.go` by appending to the `migrations` slice
3. Follow the dialect patterns:
   - SQLite: `INTEGER PRIMARY KEY AUTOINCREMENT`, `DATETIME`, `INTEGER` for booleans (0/1)
   - PostgreSQL: `SERIAL PRIMARY KEY`, `TIMESTAMP`, `BOOLEAN` (TRUE/FALSE), `BIGINT` for large sizes, `JSONB` where appropriate
4. All DDL should use `IF NOT EXISTS` for idempotency
5. Add indexes for columns used in WHERE clauses
6. Optionally create corresponding `.sql` files in `database/migrations/` for reference

### 7.4 Migration Best Practices

- Always test migrations on both SQLite and PostgreSQL
- Make migrations idempotent (use `IF NOT EXISTS`, `ON CONFLICT DO NOTHING`)
- Never modify existing migrations that have been deployed -- create a new version instead
- Use `ALTER TABLE ADD COLUMN` with duplicate-column error handling for additive schema changes (see v12 pattern)
- For SQLite FK changes, use the backup/recreate pattern (see v6)
- Seed data with `INSERT OR IGNORE` (SQLite) / `ON CONFLICT DO NOTHING` (PostgreSQL)

---

## 8. Rollback Considerations

### 8.1 General Approach

The Go-based migration system does not include automatic rollback. The SQL file migrations in `database/migrations/` include `.down.sql` files for versions 1-3 and the subtitle tables, which can be used with the `golang-migrate` CLI:

```bash
# Rollback the last migration
migrate -path database/migrations -database "postgres://..." down 1

# Rollback to a specific version
migrate -path database/migrations -database "postgres://..." goto 8
```

### 8.2 Available Down Migrations

| Version | Down Migration | What It Does |
|---------|---------------|--------------|
| 1 | 000001_initial_schema.down.sql | Drops all v1 tables and indexes |
| 2 | 000002_conversion_jobs.down.sql | Drops conversion_jobs |
| 3 | 000003_add_user_tables.down.sql | Drops user_sessions, users, roles |
| 14 | 014_create_subtitle_tables.down.sql | Drops subtitle tables and triggers |
| 15 | 015_fix_subtitle_foreign_keys.down.sql | Reverts subtitle FKs back to media_items |

### 8.3 Manual Rollback

For migrations without `.down.sql` files, manual rollback steps are:

1. Identify the tables and indexes created by the migration from this document
2. Drop them in reverse dependency order (children before parents)
3. Delete the migration version from the `migrations` table:
   ```sql
   DELETE FROM migrations WHERE version = <version_number>;
   ```

### 8.4 Dirty State Recovery

If a migration fails partway through:

```bash
# Force the migration version (marks it as clean at that version)
migrate -path database/migrations -database "..." force <VERSION>
```

For the Go-based system, manually delete the partially-applied version from the `migrations` table and restart the application.

### 8.5 Development Reset

For a complete database reset in development:

```bash
# SQLite: simply delete the database file
rm catalog.db

# PostgreSQL: drop and recreate
psql -U catalogizer -c "DROP DATABASE catalogizer;"
psql -U catalogizer -c "CREATE DATABASE catalogizer;"

# Then restart the application to re-run all migrations
```

# Catalogizer Data Dictionary

This document describes every table, column, constraint, relationship, and index in the Catalogizer database schema. The schema supports two dialects -- SQLite (development) and PostgreSQL (production) -- managed through the dual-dialect abstraction in `catalog-api/database/dialect.go`. Migrations are applied sequentially from version 1 through version 10 by `catalog-api/database/migrations.go`.

---

## Table of Contents

1. [migrations](#migrations)
2. [storage_roots](#storage_roots)
3. [files](#files)
4. [file_metadata](#file_metadata)
5. [duplicate_groups](#duplicate_groups)
6. [virtual_paths](#virtual_paths)
7. [scan_history](#scan_history)
8. [users](#users)
9. [roles](#roles)
10. [user_sessions](#user_sessions)
11. [permissions](#permissions)
12. [user_permissions](#user_permissions)
13. [auth_audit_log](#auth_audit_log)
14. [conversion_jobs](#conversion_jobs)
15. [subtitle_tracks](#subtitle_tracks)
16. [subtitle_sync_status](#subtitle_sync_status)
17. [subtitle_cache](#subtitle_cache)
18. [subtitle_downloads](#subtitle_downloads)
19. [media_subtitles](#media_subtitles)
20. [assets](#assets)
21. [media_types](#media_types)
22. [media_items](#media_items)
23. [media_files](#media_files)
24. [media_collections](#media_collections)
25. [media_collection_items](#media_collection_items)
26. [external_metadata](#external_metadata)
27. [user_metadata](#user_metadata)
28. [directory_analyses](#directory_analyses)
29. [detection_rules](#detection_rules)
30. [sync_endpoints](#sync_endpoints)
31. [sync_sessions](#sync_sessions)
32. [sync_schedules](#sync_schedules)

---

## Migration History

| Version | Name | Description |
|---------|------|-------------|
| 1 | create_initial_tables | Core tables: storage_roots, files, file_metadata, duplicate_groups, virtual_paths, scan_history |
| 2 | migrate_smb_to_storage_roots | Migrates legacy smb_roots table data into storage_roots |
| 3 | create_auth_tables | Authentication: users, roles, user_sessions, permissions, user_permissions, auth_audit_log |
| 4 | create_conversion_jobs_table | Media format conversion job queue |
| 5 | create_subtitle_tables | Subtitle management: subtitle_tracks, subtitle_sync_status, subtitle_cache, subtitle_downloads, media_subtitles |
| 6 | fix_subtitle_foreign_keys | Fixes FK references in subtitle tables (SQLite backup/recreate; PostgreSQL no-op) |
| 7 | create_assets_table | Asset management for cover art, thumbnails, and other media assets |
| 8 | create_media_entity_tables | Media entity system: media_types, media_items, media_files, media_collections, external_metadata, user_metadata, directory_analyses, detection_rules |
| 9 | create_performance_indexes | Performance indexes for files, media_items, user_metadata, media_files; deduplication of media_files |
| 10 | create_sync_tables | Remote synchronization: sync_endpoints, sync_sessions, sync_schedules |

---

## migrations

Tracks which database migrations have been applied. Created before any versioned migration runs.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| version | INTEGER | INTEGER | NO | -- | PRIMARY KEY |
| name | TEXT | TEXT | NO | -- | |
| applied_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

---

## storage_roots

Defines storage locations that Catalogizer scans. Each root represents a connection to a protocol-specific storage backend.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| name | TEXT | TEXT | NO | -- | UNIQUE |
| protocol | TEXT | TEXT | NO | -- | Values: smb, ftp, nfs, webdav, local |
| host | TEXT | TEXT | YES | -- | |
| port | INTEGER | INTEGER | YES | -- | |
| path | TEXT | TEXT | YES | -- | |
| username | TEXT | TEXT | YES | -- | |
| password | TEXT | TEXT | YES | -- | |
| domain | TEXT | TEXT | YES | -- | SMB domain |
| mount_point | TEXT | TEXT | YES | -- | NFS/local mount point |
| options | TEXT | TEXT | YES | -- | JSON protocol-specific options |
| url | TEXT | TEXT | YES | -- | WebDAV URL |
| enabled | BOOLEAN | BOOLEAN | YES | 1 / TRUE | |
| max_depth | INTEGER | INTEGER | YES | 10 | Maximum scan directory depth |
| enable_duplicate_detection | BOOLEAN | BOOLEAN | YES | 1 / TRUE | |
| enable_metadata_extraction | BOOLEAN | BOOLEAN | YES | 1 / TRUE | |
| include_patterns | TEXT | TEXT | YES | -- | JSON array of glob patterns |
| exclude_patterns | TEXT | TEXT | YES | -- | JSON array of glob patterns |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| last_scan_at | DATETIME | TIMESTAMP | YES | -- | |

---

## files

Stores metadata for every scanned file and directory. Central table of the catalog.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| storage_root_id | INTEGER | INTEGER | NO | -- | FK -> storage_roots(id) |
| path | TEXT | TEXT | NO | -- | UNIQUE with storage_root_id |
| name | TEXT | TEXT | NO | -- | |
| extension | TEXT | TEXT | YES | -- | |
| mime_type | TEXT | TEXT | YES | -- | |
| file_type | TEXT | TEXT | YES | -- | Detected media type category |
| size | INTEGER | BIGINT | NO | -- | File size in bytes |
| is_directory | BOOLEAN | BOOLEAN | YES | 0 / FALSE | |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| modified_at | DATETIME | TIMESTAMP | NO | -- | File modification timestamp |
| accessed_at | DATETIME | TIMESTAMP | YES | -- | |
| deleted | BOOLEAN | BOOLEAN | YES | 0 / FALSE | Soft delete flag |
| deleted_at | DATETIME | TIMESTAMP | YES | -- | |
| last_scan_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| last_verified_at | DATETIME | TIMESTAMP | YES | -- | |
| md5 | TEXT | TEXT | YES | -- | MD5 hash |
| sha256 | TEXT | TEXT | YES | -- | SHA-256 hash |
| sha1 | TEXT | TEXT | YES | -- | SHA-1 hash |
| blake3 | TEXT | TEXT | YES | -- | BLAKE3 hash |
| quick_hash | TEXT | TEXT | YES | -- | Fast partial hash for dedup |
| is_duplicate | BOOLEAN | BOOLEAN | YES | 0 / FALSE | |
| duplicate_group_id | INTEGER | INTEGER | YES | -- | FK -> duplicate_groups(id) |
| parent_id | INTEGER | INTEGER | YES | -- | FK -> files(id), self-referential |

**Unique Constraint:** (storage_root_id, path)

**Indexes:**
| Index Name | Columns | Type |
|------------|---------|------|
| idx_files_storage_root_path | storage_root_id, path | UNIQUE |
| idx_files_parent_id | parent_id | |
| idx_files_duplicate_group | duplicate_group_id | |
| idx_files_deleted | deleted | |
| idx_files_file_type | file_type | v9 performance |
| idx_files_extension | extension | v9 performance |
| idx_files_is_directory | is_directory | v9 performance |
| idx_files_name | name | v9 performance |

---

## file_metadata

Key-value metadata store for files. Stores extracted technical metadata such as resolution, codec, bitrate, and other properties.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| file_id | INTEGER | INTEGER | NO | -- | FK -> files(id) ON DELETE CASCADE |
| key | TEXT | TEXT | NO | -- | Metadata key name |
| value | TEXT | TEXT | NO | -- | Metadata value |
| data_type | TEXT | TEXT | YES | 'string' | Type hint: string, int, float, bool, json |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_file_metadata_file_id | file_id |

---

## duplicate_groups

Groups files that are duplicates of each other based on hash matching.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| file_count | INTEGER | INTEGER | YES | 0 | |
| total_size | INTEGER | BIGINT | YES | 0 | Sum of all file sizes in group |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

---

## virtual_paths

Maps virtual paths to physical storage targets. Provides user-friendly path aliases.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| path | TEXT | TEXT | NO | -- | UNIQUE |
| target_type | TEXT | TEXT | NO | -- | Target entity type |
| target_id | INTEGER | INTEGER | NO | -- | Target entity ID |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

---

## scan_history

Records every scan operation performed on storage roots.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| storage_root_id | INTEGER | INTEGER | NO | -- | FK -> storage_roots(id) |
| scan_type | TEXT | TEXT | NO | -- | Values: full, incremental, quick |
| status | TEXT | TEXT | NO | -- | Values: running, completed, failed, cancelled |
| start_time | DATETIME | TIMESTAMP | NO | -- | |
| end_time | DATETIME | TIMESTAMP | YES | -- | |
| files_processed | INTEGER | INTEGER | YES | 0 | |
| files_added | INTEGER | INTEGER | YES | 0 | |
| files_updated | INTEGER | INTEGER | YES | 0 | |
| files_deleted | INTEGER | INTEGER | YES | 0 | |
| error_count | INTEGER | INTEGER | YES | 0 | |
| error_message | TEXT | TEXT | YES | -- | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_scan_history_storage_root | storage_root_id |

---

## users

User accounts for authentication and authorization.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| username | TEXT | TEXT | NO | -- | UNIQUE |
| email | TEXT | TEXT | NO | -- | UNIQUE |
| password_hash | TEXT | TEXT | NO | -- | bcrypt hash |
| salt | TEXT | TEXT | NO | -- | Password salt |
| role_id | INTEGER | INTEGER | NO | -- | FK -> roles(id) |
| first_name | TEXT | TEXT | YES | -- | |
| last_name | TEXT | TEXT | YES | -- | |
| display_name | TEXT | TEXT | YES | -- | |
| avatar_url | TEXT | TEXT | YES | -- | |
| time_zone | TEXT | TEXT | YES | -- | |
| language | TEXT | TEXT | YES | -- | |
| settings | TEXT | TEXT | YES | '{}' | JSON user preferences |
| is_active | INTEGER/BOOLEAN | BOOLEAN | YES | 1 / TRUE | |
| is_locked | INTEGER/BOOLEAN | BOOLEAN | YES | 0 / FALSE | |
| locked_until | DATETIME | TIMESTAMP | YES | -- | |
| failed_login_attempts | INTEGER | INTEGER | YES | 0 | |
| last_login_at | DATETIME | TIMESTAMP | YES | -- | |
| last_login_ip | TEXT | TEXT | YES | -- | |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_users_username | username |
| idx_users_email | email |
| idx_users_role_id | role_id |
| idx_users_is_active | is_active |

---

## roles

Defines authorization roles with associated permissions.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| name | TEXT | TEXT | NO | -- | UNIQUE |
| description | TEXT | TEXT | YES | -- | |
| permissions | TEXT | TEXT | YES | '[]' | JSON array of permission strings |
| is_system | INTEGER/BOOLEAN | BOOLEAN | YES | 0 / FALSE | System roles cannot be deleted |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Seed Data:**

| id | name | permissions | is_system |
|----|------|-------------|-----------|
| 1 | Admin | ["*"] | true |
| 2 | User | ["media.view", "media.download"] | true |

---

## user_sessions

Tracks active JWT sessions per user and device.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| user_id | INTEGER | INTEGER | NO | -- | FK -> users(id) ON DELETE CASCADE |
| session_token | TEXT | TEXT | NO | -- | UNIQUE |
| refresh_token | TEXT | TEXT | YES | -- | |
| device_info | TEXT | TEXT | YES | -- | |
| ip_address | TEXT | TEXT | YES | -- | |
| user_agent | TEXT | TEXT | YES | -- | |
| is_active | INTEGER/BOOLEAN | BOOLEAN | YES | 1 / TRUE | |
| expires_at | DATETIME | TIMESTAMP | NO | -- | |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| last_activity_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_user_sessions_user_id | user_id |
| idx_user_sessions_token | session_token |
| idx_user_sessions_expires_at | expires_at |

---

## permissions

Defines granular permissions that can be assigned to users.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| name | TEXT | TEXT | NO | -- | UNIQUE |
| resource | TEXT | TEXT | NO | -- | Target resource (e.g., media, users) |
| action | TEXT | TEXT | NO | -- | Permitted action (e.g., view, edit, delete) |
| description | TEXT | TEXT | YES | -- | |

---

## user_permissions

Junction table linking users to individually granted permissions.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| user_id | INTEGER | INTEGER | NO | -- | PK, FK -> users(id) ON DELETE CASCADE |
| permission_id | INTEGER | INTEGER | NO | -- | PK, FK -> permissions(id) ON DELETE CASCADE |
| granted_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| granted_by | INTEGER | INTEGER | YES | -- | FK -> users(id) |

**Primary Key:** (user_id, permission_id)

---

## auth_audit_log

Records authentication events for security auditing.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| user_id | INTEGER | INTEGER | YES | -- | FK -> users(id) |
| event_type | TEXT | TEXT | NO | -- | Values: login, logout, failed_login, password_change, etc. |
| ip_address | TEXT | TEXT | YES | -- | |
| user_agent | TEXT | TEXT | YES | -- | |
| details | TEXT | TEXT | YES | -- | JSON event details |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

---

## conversion_jobs

Queue for media format conversion jobs.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| user_id | INTEGER | INTEGER | NO | -- | FK -> users(id) ON DELETE CASCADE |
| source_path | TEXT | TEXT | NO | -- | |
| target_path | TEXT | TEXT | NO | -- | |
| source_format | TEXT | TEXT | NO | -- | |
| target_format | TEXT | TEXT | NO | -- | |
| conversion_type | TEXT | TEXT | NO | -- | Values: video, audio, document |
| quality | TEXT | TEXT | YES | 'medium' | Values: low, medium, high, lossless |
| settings | TEXT | TEXT | YES | -- | JSON conversion settings |
| priority | INTEGER | INTEGER | YES | 0 | Higher = more urgent |
| status | TEXT | TEXT | YES | 'pending' | Values: pending, running, completed, failed, cancelled |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| started_at | DATETIME | TIMESTAMP | YES | -- | |
| completed_at | DATETIME | TIMESTAMP | YES | -- | |
| scheduled_for | DATETIME | TIMESTAMP | YES | -- | |
| duration | INTEGER | INTEGER | YES | -- | Duration in seconds |
| error_message | TEXT | TEXT | YES | -- | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_conversion_jobs_user_id | user_id |
| idx_conversion_jobs_status | status |
| idx_conversion_jobs_created_at | created_at |

---

## subtitle_tracks

Stores subtitle tracks associated with media items.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> files(id) ON DELETE CASCADE |
| language | TEXT | TEXT | NO | -- | Full language name |
| language_code | TEXT | TEXT | NO | -- | ISO 639-1 code |
| source | TEXT | TEXT | NO | 'downloaded' | Values: downloaded, uploaded, embedded, translated |
| format | TEXT | TEXT | NO | 'srt' | Values: srt, ass, ssa, vtt, sub |
| path | TEXT | TEXT | YES | -- | File path on disk |
| content | TEXT | TEXT | YES | -- | Inline subtitle content |
| is_default | BOOLEAN | BOOLEAN | YES | FALSE | |
| is_forced | BOOLEAN | BOOLEAN | YES | FALSE | |
| encoding | TEXT | TEXT | YES | 'utf-8' | Character encoding |
| sync_offset | REAL | REAL | YES | 0.0 | Timing offset in seconds |
| verified_sync | BOOLEAN | BOOLEAN | YES | FALSE | |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | Auto-updated by trigger |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_subtitle_tracks_media_item_id | media_item_id |
| idx_subtitle_tracks_language | language |
| idx_subtitle_tracks_language_code | language_code |
| idx_subtitle_tracks_source | source |

**Triggers:** `update_subtitle_tracks_updated_at` -- sets `updated_at` to CURRENT_TIMESTAMP on every UPDATE.

---

## subtitle_sync_status

Tracks ongoing subtitle synchronization and processing operations.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> files(id) ON DELETE CASCADE |
| subtitle_id | TEXT | TEXT | NO | -- | External subtitle identifier |
| operation | TEXT | TEXT | NO | -- | Values: download, translate, sync_check |
| status | TEXT | TEXT | NO | 'pending' | Values: pending, running, completed, failed |
| progress | INTEGER | INTEGER | YES | 0 | Percentage 0-100 |
| error_message | TEXT | TEXT | YES | -- | |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | Auto-updated by trigger |
| completed_at | DATETIME | TIMESTAMP | YES | -- | Set by trigger when status becomes 'completed' |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_subtitle_sync_status_media_item_id | media_item_id |
| idx_subtitle_sync_status_status | status |
| idx_subtitle_sync_status_operation | operation |

**Triggers:**
- `update_subtitle_sync_status_updated_at` -- sets `updated_at` on UPDATE.
- `set_subtitle_sync_status_completed_at` -- sets `completed_at` when status transitions to 'completed'.

---

## subtitle_cache

Caches subtitle search results from external providers to reduce API calls.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| cache_key | TEXT | TEXT | NO | -- | UNIQUE |
| result_id | TEXT | TEXT | NO | -- | Provider-specific result identifier |
| provider | TEXT | TEXT | NO | -- | Provider name (e.g., opensubtitles, subdb) |
| title | TEXT | TEXT | YES | -- | |
| language | TEXT | TEXT | YES | -- | |
| language_code | TEXT | TEXT | YES | -- | |
| download_url | TEXT | TEXT | YES | -- | |
| format | TEXT | TEXT | YES | -- | |
| encoding | TEXT | TEXT | YES | -- | |
| upload_date | DATETIME | TIMESTAMP | YES | -- | |
| downloads | INTEGER | INTEGER | YES | -- | Download count from provider |
| rating | REAL | REAL | YES | -- | User rating from provider |
| comments | INTEGER | INTEGER | YES | -- | Comment count |
| match_score | REAL | REAL | YES | -- | Confidence score 0.0-1.0 |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| expires_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | Cache TTL |
| data | TEXT | TEXT | YES | -- | JSON supplementary data |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_subtitle_cache_cache_key | cache_key |
| idx_subtitle_cache_expires_at | expires_at |

---

## subtitle_downloads

Records individual subtitle file downloads for tracking and deduplication.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> files(id) ON DELETE CASCADE |
| result_id | TEXT | TEXT | NO | -- | Cache result identifier |
| subtitle_id | TEXT | TEXT | NO | -- | External subtitle identifier |
| provider | TEXT | TEXT | NO | -- | Provider name |
| language | TEXT | TEXT | NO | -- | |
| file_path | TEXT | TEXT | YES | -- | Local file path |
| file_size | INTEGER | INTEGER | YES | -- | Size in bytes |
| download_url | TEXT | TEXT | YES | -- | Source URL |
| download_date | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| verified_sync | BOOLEAN | BOOLEAN | YES | FALSE | |
| sync_offset | REAL | REAL | YES | 0.0 | Timing offset |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_subtitle_downloads_media_item_id | media_item_id |
| idx_subtitle_downloads_result_id | result_id |
| idx_subtitle_downloads_subtitle_id | subtitle_id |
| idx_subtitle_downloads_provider | provider |
| idx_subtitle_downloads_language | language |
| idx_subtitle_downloads_download_date | download_date |

---

## media_subtitles

Junction table linking media items to their active subtitle tracks.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> files(id) ON DELETE CASCADE |
| subtitle_track_id | INTEGER | INTEGER | NO | -- | FK -> subtitle_tracks(id) ON DELETE CASCADE |
| is_active | BOOLEAN | BOOLEAN | YES | TRUE | |
| added_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Unique Constraint:** (media_item_id, subtitle_track_id)

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_media_subtitles_media_item_id | media_item_id |
| idx_media_subtitles_subtitle_track_id | subtitle_track_id |
| idx_media_subtitles_is_active | is_active |

---

## assets

Manages media assets such as cover art, thumbnails, and promotional images. Uses a text-based primary key for UUID-style identifiers.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | TEXT | TEXT | NO | -- | PRIMARY KEY (UUID) |
| type | TEXT | TEXT | NO | -- | Values: cover_art, thumbnail, backdrop, logo |
| status | TEXT | TEXT | NO | 'pending' | Values: pending, resolving, ready, failed |
| content_type | TEXT | TEXT | YES | -- | MIME type (e.g., image/jpeg) |
| size | INTEGER | BIGINT | YES | 0 | File size in bytes |
| source_hint | TEXT | TEXT | YES | -- | Hint for resolution (e.g., provider URL) |
| entity_type | TEXT | TEXT | YES | -- | Associated entity type |
| entity_id | TEXT | TEXT | YES | -- | Associated entity identifier |
| metadata | TEXT | TEXT | YES | -- | JSON metadata |
| local_path | TEXT | TEXT | YES | -- | Path in the asset store |
| created_at | TIMESTAMP | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| resolved_at | TIMESTAMP | TIMESTAMP | YES | -- | When the asset was successfully resolved |
| expires_at | TIMESTAMP | TIMESTAMP | YES | -- | Cache expiration |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_assets_entity | entity_type, entity_id |
| idx_assets_status | status |

---

## media_types

Lookup table for the 11 supported media types. Seeded during migration v8.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| name | TEXT | TEXT | NO | -- | UNIQUE |
| description | TEXT | TEXT | YES | -- | |
| detection_patterns | TEXT | TEXT | YES | -- | JSON detection patterns |
| metadata_providers | TEXT | TEXT | YES | -- | JSON provider configuration |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Seed Data:**

| name | description |
|------|-------------|
| movie | Feature films and standalone movies |
| tv_show | Television series |
| tv_season | Season of a TV show |
| tv_episode | Episode of a TV season |
| music_artist | Music artist or band |
| music_album | Music album |
| song | Individual music track |
| game | Video games |
| software | Software applications and utilities |
| book | Books and e-books |
| comic | Comics and graphic novels |

---

## media_items

Core entity table for structured media. Supports hierarchical relationships via parent_id self-reference (TV show -> seasons -> episodes; music artist -> albums -> songs).

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_type_id | INTEGER | INTEGER | NO | -- | FK -> media_types(id) |
| title | TEXT | TEXT | NO | -- | |
| original_title | TEXT | TEXT | YES | -- | Original language title |
| year | INTEGER | INTEGER | YES | -- | Release year |
| description | TEXT | TEXT | YES | -- | |
| genre | TEXT | TEXT | YES | -- | JSON array of genres |
| director | TEXT | TEXT | YES | -- | |
| cast_crew | TEXT | TEXT | YES | -- | JSON cast and crew data |
| rating | REAL | REAL | YES | -- | Average rating |
| runtime | INTEGER | INTEGER | YES | -- | Runtime in minutes |
| language | TEXT | TEXT | YES | -- | Primary language |
| country | TEXT | TEXT | YES | -- | Country of origin |
| status | TEXT | TEXT | NO | 'detected' | Values: detected, confirmed, manual, archived |
| parent_id | INTEGER | INTEGER | YES | -- | FK -> media_items(id) ON DELETE CASCADE |
| season_number | INTEGER | INTEGER | YES | -- | For tv_season and tv_episode |
| episode_number | INTEGER | INTEGER | YES | -- | For tv_episode |
| track_number | INTEGER | INTEGER | YES | -- | For song |
| first_detected | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| last_updated | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns | Type |
|------------|---------|------|
| idx_media_items_type | media_type_id | |
| idx_media_items_parent | parent_id | |
| idx_media_items_title | title | |
| idx_media_items_title_type | title, media_type_id | v9 compound |
| idx_media_items_status | status | v9 performance |
| idx_media_items_year | year | v9 performance |

---

## media_files

Junction table linking media entities to their physical files. A media item can have multiple files (e.g., different quality versions), and a file can belong to multiple media items.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> media_items(id) ON DELETE CASCADE |
| file_id | INTEGER | INTEGER | NO | -- | FK -> files(id) ON DELETE CASCADE |
| quality_info | TEXT | TEXT | YES | -- | JSON quality metadata |
| language | TEXT | TEXT | YES | -- | File language |
| is_primary | INTEGER/BOOLEAN | BOOLEAN | YES | 0 / FALSE | Primary file for the entity |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns | Type |
|------------|---------|------|
| idx_media_files_item | media_item_id | |
| idx_media_files_file | file_id | |
| idx_media_files_item_file | media_item_id, file_id | UNIQUE (v9) |

---

## media_collections

User-created or auto-generated collections of media items.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| name | TEXT | TEXT | NO | -- | |
| collection_type | TEXT | TEXT | NO | -- | Values: manual, smart, dynamic |
| description | TEXT | TEXT | YES | -- | |
| total_items | INTEGER | INTEGER | YES | 0 | |
| external_ids | TEXT | TEXT | YES | -- | JSON external identifiers |
| cover_url | TEXT | TEXT | YES | -- | Collection cover image URL |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

---

## media_collection_items

Junction table linking media items to collections with ordering.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| collection_id | INTEGER | INTEGER | NO | -- | FK -> media_collections(id) ON DELETE CASCADE |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> media_items(id) ON DELETE CASCADE |
| sequence_number | INTEGER | INTEGER | YES | -- | Position in collection |
| season_number | INTEGER | INTEGER | YES | -- | Season grouping |
| release_order | INTEGER | INTEGER | YES | -- | Chronological order |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_media_collection_items_collection | collection_id |
| idx_media_collection_items_item | media_item_id |

---

## external_metadata

Stores metadata fetched from external providers (TMDB, IMDB, MusicBrainz, etc.).

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> media_items(id) ON DELETE CASCADE |
| provider | TEXT | TEXT | NO | -- | Values: tmdb, imdb, tvdb, musicbrainz, spotify, steam |
| external_id | TEXT | TEXT | NO | -- | Provider-specific identifier |
| data | TEXT | TEXT | YES | -- | JSON provider response data |
| rating | REAL | REAL | YES | -- | Provider rating |
| review_url | TEXT | TEXT | YES | -- | Link to reviews |
| cover_url | TEXT | TEXT | YES | -- | Cover art URL |
| trailer_url | TEXT | TEXT | YES | -- | Trailer URL |
| last_fetched | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_external_metadata_item | media_item_id |
| idx_external_metadata_provider | provider, external_id |

---

## user_metadata

Per-user metadata for media items: personal ratings, watch status, notes, tags.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_item_id | INTEGER | INTEGER | NO | -- | FK -> media_items(id) ON DELETE CASCADE |
| user_id | INTEGER | INTEGER | NO | -- | FK -> users(id) ON DELETE CASCADE |
| user_rating | REAL | REAL | YES | -- | User's personal rating |
| watched_status | TEXT | TEXT | YES | -- | Values: unwatched, watching, watched |
| watched_date | DATETIME | TIMESTAMP | YES | -- | |
| personal_notes | TEXT | TEXT | YES | -- | |
| tags | TEXT | TEXT | YES | -- | JSON array of user tags |
| favorite | INTEGER/BOOLEAN | BOOLEAN | YES | 0 / FALSE | |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns | Type |
|------------|---------|------|
| idx_user_metadata_item | media_item_id | |
| idx_user_metadata_user | user_id | |
| idx_user_metadata_user_watched | user_id, watched_status | v9 compound |

---

## directory_analyses

Stores results of directory-level media detection analysis.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| directory_path | TEXT | TEXT | NO | -- | |
| smb_root | TEXT | TEXT | YES | -- | Legacy SMB root reference |
| media_item_id | INTEGER | INTEGER | YES | -- | FK -> media_items(id) ON DELETE SET NULL |
| confidence_score | REAL | REAL | YES | 0 | Detection confidence 0.0-1.0 |
| detection_method | TEXT | TEXT | YES | -- | Method used for detection |
| analysis_data | TEXT | TEXT | YES | -- | JSON analysis details |
| last_analyzed | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| files_count | INTEGER | INTEGER | YES | 0 | Number of files in directory |
| total_size | INTEGER | BIGINT | YES | 0 | Total size of directory contents |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_directory_analyses_path | directory_path |

---

## detection_rules

Configurable rules for media type detection. Rules can be enabled/disabled and prioritized.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| media_type_id | INTEGER | INTEGER | NO | -- | FK -> media_types(id) ON DELETE CASCADE |
| rule_name | TEXT | TEXT | NO | -- | |
| rule_type | TEXT | TEXT | NO | -- | Values: regex, extension, path_pattern |
| pattern | TEXT | TEXT | NO | -- | Pattern string |
| confidence_weight | REAL | REAL | YES | 1.0 | Weight applied to confidence score |
| enabled | INTEGER/BOOLEAN | BOOLEAN | YES | 1 / TRUE | |
| priority | INTEGER | INTEGER | YES | 0 | Higher = evaluated first |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_detection_rules_type | media_type_id |

---

## sync_endpoints

Defines remote synchronization endpoints for file sync operations.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| user_id | INTEGER | INTEGER | NO | -- | FK -> users(id) |
| name | TEXT | TEXT | NO | -- | Endpoint display name |
| type | TEXT | TEXT | NO | -- | Values: s3, gcs, webdav, local |
| url | TEXT | TEXT | NO | -- | Endpoint URL or path |
| username | TEXT | TEXT | YES | -- | |
| password | TEXT | TEXT | YES | -- | |
| sync_direction | TEXT | TEXT | YES | 'bidirectional' | Values: push, pull, bidirectional |
| local_path | TEXT | TEXT | YES | -- | Local directory path |
| remote_path | TEXT | TEXT | YES | -- | Remote directory path |
| sync_settings | TEXT | TEXT | YES | -- | JSON sync configuration |
| status | TEXT | TEXT | YES | 'active' | Values: active, paused, error |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| last_sync_at | DATETIME | TIMESTAMP | YES | -- | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_sync_endpoints_user_id | user_id |
| idx_sync_endpoints_status | status |
| idx_sync_endpoints_type | type |

---

## sync_sessions

Records individual sync execution runs with progress tracking.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| endpoint_id | INTEGER | INTEGER | NO | -- | FK -> sync_endpoints(id) |
| user_id | INTEGER | INTEGER | NO | -- | FK -> users(id) |
| status | TEXT | TEXT | YES | 'running' | Values: running, completed, failed, cancelled |
| sync_type | TEXT | TEXT | YES | -- | Type of sync operation |
| started_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |
| completed_at | DATETIME | TIMESTAMP | YES | -- | |
| duration | INTEGER | INTEGER | YES | -- | Duration in seconds |
| total_files | INTEGER | INTEGER | YES | 0 | |
| synced_files | INTEGER | INTEGER | YES | 0 | |
| failed_files | INTEGER | INTEGER | YES | 0 | |
| skipped_files | INTEGER | INTEGER | YES | 0 | |
| error_message | TEXT | TEXT | YES | -- | |
| updated_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_sync_sessions_endpoint_id | endpoint_id |
| idx_sync_sessions_user_id | user_id |
| idx_sync_sessions_status | status |
| idx_sync_sessions_started_at | started_at |

---

## sync_schedules

Configures recurring synchronization schedules.

| Column | SQLite Type | PostgreSQL Type | Nullable | Default | Constraints |
|--------|-------------|-----------------|----------|---------|-------------|
| id | INTEGER | SERIAL | NO | AUTOINCREMENT | PRIMARY KEY |
| endpoint_id | INTEGER | INTEGER | NO | -- | FK -> sync_endpoints(id) |
| user_id | INTEGER | INTEGER | NO | -- | FK -> users(id) |
| frequency | TEXT | TEXT | NO | -- | Cron expression or interval string |
| last_run | DATETIME | TIMESTAMP | YES | -- | |
| next_run | DATETIME | TIMESTAMP | YES | -- | |
| is_active | BOOLEAN | BOOLEAN | YES | 1 / TRUE | |
| created_at | DATETIME | TIMESTAMP | YES | CURRENT_TIMESTAMP | |

**Indexes:**
| Index Name | Columns |
|------------|---------|
| idx_sync_schedules_endpoint_id | endpoint_id |
| idx_sync_schedules_user_id | user_id |
| idx_sync_schedules_is_active | is_active |
| idx_sync_schedules_next_run | next_run |

---

## Entity Relationship Summary

```
storage_roots 1--* files
files 1--* file_metadata
files *--1 duplicate_groups
files *--1 files (parent_id self-reference)
files 1--* media_files
media_items 1--* media_files
media_items *--1 media_types
media_items *--1 media_items (parent_id hierarchy)
media_items 1--* external_metadata
media_items 1--* user_metadata
media_items *--* media_collections (via media_collection_items)
media_items 1--* directory_analyses
media_types 1--* detection_rules
users 1--* user_sessions
users 1--* user_permissions
users *--1 roles
users 1--* conversion_jobs
users 1--* user_metadata
users 1--* sync_endpoints
users 1--* auth_audit_log
sync_endpoints 1--* sync_sessions
sync_endpoints 1--* sync_schedules
files 1--* subtitle_tracks
files 1--* subtitle_sync_status
files 1--* subtitle_downloads
files 1--* media_subtitles
subtitle_tracks 1--* media_subtitles
```

---

## Dialect Differences

| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| Auto-increment PK | `INTEGER PRIMARY KEY AUTOINCREMENT` | `SERIAL PRIMARY KEY` |
| Boolean type | `INTEGER` (0/1) or `BOOLEAN` | `BOOLEAN` (TRUE/FALSE) |
| Large integers | `INTEGER` | `BIGINT` (for size columns) |
| Timestamp type | `DATETIME` | `TIMESTAMP` |
| Upsert syntax | `INSERT OR IGNORE` | `ON CONFLICT DO NOTHING` |
| Placeholders | `?` | `$1, $2, ...` (auto-rewritten) |
| Boolean literals | `= 0` / `= 1` | `= FALSE` / `= TRUE` (auto-rewritten) |
| Last insert ID | `LastInsertId()` | `RETURNING id` (via `InsertReturningID()`) |
| Trigger syntax | `AFTER UPDATE ... BEGIN ... END` | `BEFORE UPDATE ... EXECUTE FUNCTION` |
| FK recreation | Backup/recreate pattern | `ALTER TABLE` (or no-op if correct) |

The `database.DB` wrapper automatically rewrites SQL using `RewritePlaceholders()`, `RewriteInsertOrIgnore()`, and `BooleanLiterals()` so application code writes SQLite-style SQL and PostgreSQL queries are generated transparently.

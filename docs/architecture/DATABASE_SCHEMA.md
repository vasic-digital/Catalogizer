# Catalogizer Database Schema

This document describes the complete database schema for the Catalogizer system. The database uses SQLite (with optional SQLCipher encryption) for development and PostgreSQL for production. Schema is managed through versioned migrations defined in `catalog-api/database/migrations.go`.

## Table of Contents

- [ER Diagram](#er-diagram)
- [Core File Catalog Tables](#core-file-catalog-tables)
- [Authentication and Authorization Tables](#authentication-and-authorization-tables)
- [Media Detection Tables](#media-detection-tables)
- [Subtitle Management Tables](#subtitle-management-tables)
- [Media Conversion Tables](#media-conversion-tables)
- [Multi-User v3 Tables](#multi-user-v3-tables)
- [System Tables](#system-tables)
- [Index Documentation](#index-documentation)
- [Triggers](#database-triggers)
- [Views](#database-views)
- [Migration History](#migration-history)

---

## ER Diagram

```mermaid
erDiagram
    storage_roots {
        INTEGER id PK
        TEXT name UK
        TEXT protocol
        TEXT host
        INTEGER port
        TEXT path
        TEXT username
        TEXT password
        TEXT domain
        TEXT mount_point
        TEXT options
        TEXT url
        BOOLEAN enabled
        INTEGER max_depth
        BOOLEAN enable_duplicate_detection
        BOOLEAN enable_metadata_extraction
        TEXT include_patterns
        TEXT exclude_patterns
        DATETIME created_at
        DATETIME updated_at
        DATETIME last_scan_at
    }

    files {
        INTEGER id PK
        INTEGER storage_root_id FK
        TEXT path
        TEXT name
        TEXT extension
        TEXT mime_type
        TEXT file_type
        INTEGER size
        BOOLEAN is_directory
        DATETIME created_at
        DATETIME modified_at
        DATETIME accessed_at
        BOOLEAN deleted
        DATETIME deleted_at
        DATETIME last_scan_at
        DATETIME last_verified_at
        TEXT md5
        TEXT sha256
        TEXT sha1
        TEXT blake3
        TEXT quick_hash
        BOOLEAN is_duplicate
        INTEGER duplicate_group_id FK
        INTEGER parent_id FK
    }

    file_metadata {
        INTEGER id PK
        INTEGER file_id FK
        TEXT key
        TEXT value
        TEXT data_type
    }

    duplicate_groups {
        INTEGER id PK
        INTEGER file_count
        INTEGER total_size
        DATETIME created_at
        DATETIME updated_at
    }

    virtual_paths {
        INTEGER id PK
        TEXT path UK
        TEXT target_type
        INTEGER target_id
        DATETIME created_at
    }

    scan_history {
        INTEGER id PK
        INTEGER storage_root_id FK
        TEXT scan_type
        TEXT status
        DATETIME start_time
        DATETIME end_time
        INTEGER files_processed
        INTEGER files_added
        INTEGER files_updated
        INTEGER files_deleted
        INTEGER error_count
        TEXT error_message
    }

    users {
        INTEGER id PK
        TEXT username UK
        TEXT email UK
        TEXT password_hash
        TEXT salt
        INTEGER role_id FK
        TEXT first_name
        TEXT last_name
        TEXT display_name
        TEXT avatar_url
        TEXT time_zone
        TEXT language
        TEXT settings
        INTEGER is_active
        INTEGER is_locked
        DATETIME locked_until
        INTEGER failed_login_attempts
        DATETIME last_login_at
        TEXT last_login_ip
        DATETIME created_at
        DATETIME updated_at
    }

    roles {
        INTEGER id PK
        TEXT name UK
        TEXT description
        TEXT permissions
        INTEGER is_system
        DATETIME created_at
        DATETIME updated_at
    }

    user_sessions {
        INTEGER id PK
        INTEGER user_id FK
        TEXT session_token UK
        TEXT refresh_token
        TEXT device_info
        TEXT ip_address
        TEXT user_agent
        INTEGER is_active
        DATETIME expires_at
        DATETIME created_at
        DATETIME last_activity_at
    }

    permissions {
        INTEGER id PK
        TEXT name UK
        TEXT resource
        TEXT action
        TEXT description
    }

    user_permissions {
        INTEGER user_id FK
        INTEGER permission_id FK
        DATETIME granted_at
        INTEGER granted_by FK
    }

    auth_audit_log {
        INTEGER id PK
        INTEGER user_id FK
        TEXT event_type
        TEXT ip_address
        TEXT user_agent
        TEXT details
        DATETIME created_at
    }

    conversion_jobs {
        INTEGER id PK
        INTEGER user_id FK
        TEXT source_path
        TEXT target_path
        TEXT source_format
        TEXT target_format
        TEXT conversion_type
        TEXT quality
        TEXT settings
        INTEGER priority
        TEXT status
        DATETIME created_at
        DATETIME started_at
        DATETIME completed_at
        DATETIME scheduled_for
        INTEGER duration
        TEXT error_message
    }

    subtitle_tracks {
        INTEGER id PK
        INTEGER media_item_id FK
        TEXT language
        TEXT language_code
        TEXT source
        TEXT format
        TEXT path
        TEXT content
        BOOLEAN is_default
        BOOLEAN is_forced
        TEXT encoding
        REAL sync_offset
        BOOLEAN verified_sync
        DATETIME created_at
        DATETIME updated_at
    }

    subtitle_sync_status {
        INTEGER id PK
        INTEGER media_item_id FK
        TEXT subtitle_id
        TEXT operation
        TEXT status
        INTEGER progress
        TEXT error_message
        DATETIME created_at
        DATETIME updated_at
        DATETIME completed_at
    }

    subtitle_cache {
        INTEGER id PK
        TEXT cache_key UK
        TEXT result_id
        TEXT provider
        TEXT title
        TEXT language
        TEXT language_code
        TEXT download_url
        TEXT format
        TEXT encoding
        DATETIME upload_date
        INTEGER downloads
        REAL rating
        INTEGER comments
        REAL match_score
        DATETIME created_at
        DATETIME expires_at
        TEXT data
    }

    subtitle_downloads {
        INTEGER id PK
        INTEGER media_item_id FK
        TEXT result_id
        TEXT subtitle_id
        TEXT provider
        TEXT language
        TEXT file_path
        INTEGER file_size
        TEXT download_url
        DATETIME download_date
        BOOLEAN verified_sync
        REAL sync_offset
    }

    media_subtitles {
        INTEGER id PK
        INTEGER media_item_id FK
        INTEGER subtitle_track_id FK
        BOOLEAN is_active
        DATETIME added_at
    }

    media_types {
        INTEGER id PK
        TEXT name UK
        TEXT description
        TEXT detection_patterns
        TEXT metadata_providers
        DATETIME created_at
        DATETIME updated_at
    }

    media_items {
        INTEGER id PK
        INTEGER media_type_id FK
        TEXT title
        TEXT original_title
        INTEGER year
        TEXT description
        TEXT genre
        TEXT director
        TEXT cast_crew
        REAL rating
        INTEGER runtime
        TEXT language
        TEXT country
        TEXT status
        DATETIME first_detected
        DATETIME last_updated
    }

    external_metadata {
        INTEGER id PK
        INTEGER media_item_id FK
        TEXT provider
        TEXT external_id
        TEXT data
        REAL rating
        TEXT review_url
        TEXT cover_url
        TEXT trailer_url
        DATETIME last_fetched
    }

    directory_analysis {
        INTEGER id PK
        TEXT directory_path UK
        TEXT smb_root
        INTEGER media_item_id FK
        REAL confidence_score
        TEXT detection_method
        TEXT analysis_data
        DATETIME last_analyzed
        INTEGER files_count
        INTEGER total_size
    }

    media_files {
        INTEGER id PK
        INTEGER media_item_id FK
        TEXT file_path
        TEXT smb_root
        TEXT filename
        INTEGER file_size
        TEXT file_extension
        TEXT quality_info
        TEXT language
        TEXT subtitle_tracks_json
        TEXT audio_tracks
        INTEGER duration
        TEXT checksum
        TEXT virtual_smb_link
        TEXT direct_smb_link
        DATETIME last_verified
        DATETIME created_at
    }

    quality_profiles {
        INTEGER id PK
        TEXT name UK
        INTEGER resolution_width
        INTEGER resolution_height
        INTEGER min_bitrate
        INTEGER max_bitrate
        TEXT preferred_codecs
        INTEGER quality_score
        DATETIME created_at
    }

    change_log {
        INTEGER id PK
        TEXT entity_type
        TEXT entity_id
        TEXT change_type
        TEXT old_data
        TEXT new_data
        DATETIME detected_at
        BOOLEAN processed
    }

    media_collections {
        INTEGER id PK
        TEXT name
        TEXT collection_type
        TEXT description
        INTEGER total_items
        TEXT external_ids
        TEXT cover_url
        DATETIME created_at
        DATETIME updated_at
    }

    media_collection_items {
        INTEGER id PK
        INTEGER collection_id FK
        INTEGER media_item_id FK
        INTEGER sequence_number
        INTEGER season_number
        INTEGER release_order
    }

    user_metadata {
        INTEGER id PK
        INTEGER media_item_id FK
        REAL user_rating
        TEXT watched_status
        DATETIME watched_date
        TEXT personal_notes
        TEXT tags
        BOOLEAN favorite
        DATETIME created_at
        DATETIME updated_at
    }

    detection_rules {
        INTEGER id PK
        INTEGER media_type_id FK
        TEXT rule_name
        TEXT rule_type
        TEXT pattern
        REAL confidence_weight
        BOOLEAN enabled
        INTEGER priority
        DATETIME created_at
    }

    media_access_logs {
        INTEGER id PK
        INTEGER user_id FK
        INTEGER media_id FK
        TEXT action
        REAL location_latitude
        REAL location_longitude
        REAL location_accuracy
        TEXT location_address
        TEXT device_info
        TEXT session_id
        INTEGER duration_seconds
        INTEGER position_seconds
        TEXT quality_level
        TEXT metadata_json
        DATETIME created_at
    }

    favorites {
        INTEGER id PK
        INTEGER user_id FK
        TEXT entity_type
        INTEGER entity_id
        INTEGER category_id FK
        TEXT notes
        INTEGER sort_order
        BOOLEAN is_pinned
        DATETIME created_at
    }

    favorite_categories {
        INTEGER id PK
        INTEGER user_id FK
        TEXT name
        TEXT description
        TEXT color
        TEXT icon
        INTEGER parent_id FK
        INTEGER sort_order
        DATETIME created_at
    }

    analytics_events {
        INTEGER id PK
        INTEGER user_id FK
        TEXT session_id
        TEXT event_type
        TEXT event_category
        TEXT event_action
        TEXT event_label
        REAL event_value
        TEXT entity_type
        INTEGER entity_id
        TEXT properties
        REAL location_latitude
        REAL location_longitude
        TEXT device_info
        TEXT user_agent
        TEXT ip_address
        DATETIME created_at
    }

    analytics_reports {
        INTEGER id PK
        TEXT name
        TEXT description
        TEXT report_type
        TEXT parameters
        TEXT schedule_expression
        TEXT output_format
        INTEGER created_by FK
        BOOLEAN is_active
        DATETIME last_generated_at
        DATETIME next_generation_at
        DATETIME created_at
        DATETIME updated_at
    }

    generated_reports {
        INTEGER id PK
        INTEGER report_id FK
        TEXT file_path
        INTEGER file_size
        TEXT output_format
        REAL generation_time_seconds
        TEXT parameters_used
        INTEGER row_count
        DATETIME created_at
        DATETIME expires_at
    }

    conversion_profiles {
        INTEGER id PK
        TEXT name
        TEXT description
        TEXT input_formats
        TEXT output_format
        TEXT parameters
        BOOLEAN is_system_profile
        INTEGER created_by FK
        BOOLEAN is_active
        INTEGER usage_count
        DATETIME created_at
    }

    sync_endpoints {
        INTEGER id PK
        INTEGER user_id FK
        TEXT name
        TEXT endpoint_type
        TEXT url
        TEXT credentials
        TEXT settings_json
        BOOLEAN is_active
        TEXT sync_direction
        DATETIME last_sync_at
        DATETIME last_successful_sync_at
        TEXT sync_status
        TEXT error_message
        INTEGER files_synced
        INTEGER bytes_synced
        DATETIME created_at
        DATETIME updated_at
    }

    sync_history {
        INTEGER id PK
        INTEGER endpoint_id FK
        TEXT sync_type
        TEXT direction
        TEXT status
        INTEGER files_processed
        INTEGER files_successful
        INTEGER files_failed
        INTEGER bytes_transferred
        REAL duration_seconds
        DATETIME started_at
        DATETIME completed_at
        TEXT error_summary
        TEXT details
    }

    sync_conflicts {
        INTEGER id PK
        INTEGER endpoint_id FK
        TEXT local_file_path
        TEXT remote_file_path
        TEXT conflict_type
        TEXT local_file_info
        TEXT remote_file_info
        TEXT resolution
        INTEGER resolved_by FK
        DATETIME resolved_at
        DATETIME created_at
    }

    error_reports {
        INTEGER id PK
        INTEGER user_id FK
        TEXT session_id
        TEXT error_type
        TEXT error_level
        TEXT error_message
        TEXT error_code
        TEXT stack_trace
        TEXT context
        TEXT device_info
        TEXT app_version
        TEXT os_version
        TEXT user_agent
        TEXT url
        TEXT user_feedback
        BOOLEAN is_crash
        BOOLEAN is_resolved
        INTEGER resolved_by FK
        DATETIME resolved_at
        TEXT resolution_notes
        DATETIME created_at
    }

    system_logs {
        INTEGER id PK
        TEXT log_level
        TEXT component
        TEXT message
        TEXT context
        INTEGER user_id FK
        TEXT session_id
        TEXT file_name
        INTEGER line_number
        TEXT function_name
        TEXT request_id
        TEXT ip_address
        DATETIME created_at
    }

    log_exports {
        INTEGER id PK
        INTEGER requested_by FK
        TEXT export_type
        TEXT filters
        TEXT file_path
        INTEGER file_size
        TEXT status
        TEXT privacy_level
        DATETIME expires_at
        DATETIME created_at
        DATETIME completed_at
    }

    system_config {
        INTEGER id PK
        TEXT config_key UK
        TEXT config_value
        TEXT data_type
        TEXT description
        TEXT category
        BOOLEAN is_system_config
        BOOLEAN requires_restart
        BOOLEAN is_secret
        INTEGER updated_by FK
        DATETIME created_at
        DATETIME updated_at
    }

    user_preferences {
        INTEGER id PK
        INTEGER user_id FK
        TEXT preference_key
        TEXT preference_value
        TEXT data_type
        DATETIME created_at
        DATETIME updated_at
    }

    performance_metrics {
        INTEGER id PK
        TEXT metric_type
        TEXT metric_name
        REAL metric_value
        TEXT unit
        TEXT context
        INTEGER user_id FK
        TEXT session_id
        DATETIME created_at
    }

    health_checks {
        INTEGER id PK
        TEXT check_name
        TEXT status
        REAL response_time_ms
        TEXT details
        DATETIME checked_at
    }

    schema_migrations {
        TEXT version PK
        DATETIME applied_at
    }

    migrations {
        INTEGER version PK
        TEXT name
        DATETIME applied_at
    }

    %% Core File Catalog Relationships
    storage_roots ||--o{ files : "contains"
    storage_roots ||--o{ scan_history : "tracked by"
    files ||--o{ file_metadata : "has metadata"
    files ||--o{ files : "parent_id"
    files }o--o| duplicate_groups : "belongs to"

    %% Subtitle Relationships (FK to files after migration 6)
    files ||--o{ subtitle_tracks : "has subtitles"
    files ||--o{ subtitle_sync_status : "sync tracked"
    files ||--o{ subtitle_downloads : "downloaded for"
    files ||--o{ media_subtitles : "associated via"
    subtitle_tracks ||--o{ media_subtitles : "linked via"

    %% Auth Relationships
    roles ||--o{ users : "assigned to"
    users ||--o{ user_sessions : "has sessions"
    users ||--o{ user_permissions : "has custom perms"
    users ||--o{ auth_audit_log : "audit trail"
    users ||--o{ conversion_jobs : "creates"
    permissions ||--o{ user_permissions : "granted as"

    %% Media Detection Relationships
    media_types ||--o{ media_items : "classifies"
    media_types ||--o{ detection_rules : "has rules"
    media_items ||--o{ external_metadata : "has external data"
    media_items ||--o{ media_files : "has files"
    media_items ||--o{ directory_analysis : "analyzed as"
    media_items ||--o{ user_metadata : "user prefs"
    media_items ||--o{ media_collection_items : "in collections"
    media_collections ||--o{ media_collection_items : "contains"

    %% Multi-User v3 Relationships
    users ||--o{ media_access_logs : "accesses media"
    files ||--o{ media_access_logs : "accessed by"
    users ||--o{ favorites : "favorites"
    users ||--o{ favorite_categories : "organizes"
    favorite_categories ||--o{ favorites : "categorizes"
    favorite_categories ||--o{ favorite_categories : "parent hierarchy"
    users ||--o{ analytics_events : "generates events"
    users ||--o{ analytics_reports : "creates reports"
    analytics_reports ||--o{ generated_reports : "generates"
    users ||--o{ sync_endpoints : "syncs via"
    sync_endpoints ||--o{ sync_history : "has history"
    sync_endpoints ||--o{ sync_conflicts : "has conflicts"
    users ||--o{ error_reports : "reports errors"
    users ||--o{ system_logs : "logged for"
    users ||--o{ log_exports : "exports logs"
    users ||--o{ system_config : "updates config"
    users ||--o{ user_preferences : "has preferences"
    users ||--o{ performance_metrics : "measured for"
    users ||--o{ conversion_profiles : "creates profiles"
```

---

## Core File Catalog Tables

### storage_roots

Unified storage root configuration supporting multiple protocols (SMB, FTP, NFS, WebDAV, Local). Replaces the legacy `smb_roots` table.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Unique human-readable name |
| protocol | TEXT | No | - | Protocol type: `smb`, `ftp`, `nfs`, `webdav`, `local` |
| host | TEXT | Yes | NULL | Server hostname/IP (not used for local) |
| port | INTEGER | Yes | NULL | Server port (protocol-specific default if omitted) |
| path | TEXT | Yes | NULL | Share name (SMB), remote path (FTP/NFS/WebDAV), base path (local) |
| username | TEXT | Yes | NULL | Authentication username |
| password | TEXT | Yes | NULL | Authentication password |
| domain | TEXT | Yes | NULL | SMB domain/workgroup |
| mount_point | TEXT | Yes | NULL | NFS local mount point |
| options | TEXT | Yes | NULL | Protocol-specific options (e.g., NFS mount options) |
| url | TEXT | Yes | NULL | WebDAV endpoint URL |
| enabled | BOOLEAN | No | 1 | Whether this root is active for scanning |
| max_depth | INTEGER | No | 10 | Maximum directory traversal depth |
| enable_duplicate_detection | BOOLEAN | No | 1 | Run hash-based duplicate detection |
| enable_metadata_extraction | BOOLEAN | No | 1 | Extract file metadata during scanning |
| include_patterns | TEXT | Yes | NULL | Glob patterns for files to include |
| exclude_patterns | TEXT | Yes | NULL | Glob patterns for files to exclude |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Record creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |
| last_scan_at | DATETIME | Yes | NULL | Last successful scan time |

### files

Central file catalog storing all discovered files across all storage roots.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| storage_root_id | INTEGER | No | - | FK to storage_roots |
| path | TEXT | No | - | Relative path within storage root |
| name | TEXT | No | - | File or directory name |
| extension | TEXT | Yes | NULL | File extension (without dot) |
| mime_type | TEXT | Yes | NULL | Detected MIME type |
| file_type | TEXT | Yes | NULL | Categorized type: video, audio, image, text, book, game, other |
| size | INTEGER | No | - | File size in bytes |
| is_directory | BOOLEAN | No | 0 | Whether this entry is a directory |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Record creation time |
| modified_at | DATETIME | No | - | File last modification time |
| accessed_at | DATETIME | Yes | NULL | File last access time |
| deleted | BOOLEAN | No | 0 | Soft-delete flag |
| deleted_at | DATETIME | Yes | NULL | When the file was marked deleted |
| last_scan_at | DATETIME | No | CURRENT_TIMESTAMP | Last time file was verified during scan |
| last_verified_at | DATETIME | Yes | NULL | Last integrity verification time |
| md5 | TEXT | Yes | NULL | MD5 hash |
| sha256 | TEXT | Yes | NULL | SHA-256 hash |
| sha1 | TEXT | Yes | NULL | SHA-1 hash |
| blake3 | TEXT | Yes | NULL | BLAKE3 hash |
| quick_hash | TEXT | Yes | NULL | Partial/fast hash for quick comparison |
| is_duplicate | BOOLEAN | No | 0 | Whether this file has known duplicates |
| duplicate_group_id | INTEGER | Yes | NULL | FK to duplicate_groups |
| parent_id | INTEGER | Yes | NULL | FK to files (parent directory) |

### file_metadata

Key-value metadata store for file attributes discovered during scanning.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| file_id | INTEGER | No | - | FK to files (CASCADE delete) |
| key | TEXT | No | - | Metadata key name |
| value | TEXT | No | - | Metadata value |
| data_type | TEXT | No | 'string' | Value type: string, integer, float, boolean, json |

### duplicate_groups

Groups of files identified as duplicates based on hash matching.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| file_count | INTEGER | No | 0 | Number of files in group |
| total_size | INTEGER | No | 0 | Combined size of all duplicates |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | When the group was first identified |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

### virtual_paths

Virtual file system paths that map to files or storage roots for unified navigation.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| path | TEXT | No | - | Virtual path (unique) |
| target_type | TEXT | No | - | Target entity type |
| target_id | INTEGER | No | - | Target entity ID |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |

### scan_history

Audit trail of all scan operations performed on storage roots.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| storage_root_id | INTEGER | No | - | FK to storage_roots |
| scan_type | TEXT | No | - | Type of scan: full, incremental, verify |
| status | TEXT | No | - | Status: running, completed, failed |
| start_time | DATETIME | No | - | Scan start time |
| end_time | DATETIME | Yes | NULL | Scan completion time |
| files_processed | INTEGER | No | 0 | Total files examined |
| files_added | INTEGER | No | 0 | New files discovered |
| files_updated | INTEGER | No | 0 | Files with changed metadata |
| files_deleted | INTEGER | No | 0 | Files marked as deleted |
| error_count | INTEGER | No | 0 | Number of errors during scan |
| error_message | TEXT | Yes | NULL | Error details if failed |

---

## Authentication and Authorization Tables

### users

User accounts with profile information, security settings, and login tracking.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| username | TEXT | No | - | Unique username |
| email | TEXT | No | - | Unique email address |
| password_hash | TEXT | No | - | bcrypt password hash |
| salt | TEXT | No | - | Password salt |
| role_id | INTEGER | No | - | FK to roles |
| first_name | TEXT | Yes | NULL | First name |
| last_name | TEXT | Yes | NULL | Last name |
| display_name | TEXT | Yes | NULL | Display name |
| avatar_url | TEXT | Yes | NULL | Profile avatar URL |
| time_zone | TEXT | Yes | NULL | User timezone |
| language | TEXT | Yes | NULL | Preferred language |
| settings | TEXT | No | '{}' | JSON user settings/preferences |
| is_active | INTEGER | No | 1 | Account active flag |
| is_locked | INTEGER | No | 0 | Account locked flag |
| locked_until | DATETIME | Yes | NULL | Lock expiration time |
| failed_login_attempts | INTEGER | No | 0 | Consecutive failed logins |
| last_login_at | DATETIME | Yes | NULL | Last successful login time |
| last_login_ip | TEXT | Yes | NULL | IP address of last login |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Account creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last profile update time |

### roles

Role definitions with associated permission sets. System roles (Admin, User) are seeded automatically.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Unique role name |
| description | TEXT | Yes | NULL | Role description |
| permissions | TEXT | No | '[]' | JSON array of permission strings |
| is_system | INTEGER | No | 0 | System role (cannot be deleted) |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

Default roles seeded at migration 3:
- **Admin** (id=1): `["*"]` -- full wildcard permissions
- **User** (id=2): `["media.view", "media.download"]` -- basic access

Extended roles from v3 multiuser schema:
- **Super Administrator**: `["*"]` -- full system access
- **Manager**: `["user.view", "media.manage", "share.create", "analytics.view"]`
- **Guest**: `["media.view"]` -- read-only access

### user_sessions

Active user sessions tracking device info, tokens, and activity.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | No | - | FK to users (CASCADE delete) |
| session_token | TEXT | No | - | JWT access token (unique) |
| refresh_token | TEXT | Yes | NULL | Refresh token for re-authentication |
| device_info | TEXT | Yes | NULL | JSON device metadata |
| ip_address | TEXT | Yes | NULL | Client IP address |
| user_agent | TEXT | Yes | NULL | Client user agent string |
| is_active | INTEGER | No | 1 | Whether session is valid |
| expires_at | DATETIME | No | - | Token expiration time |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Session creation time |
| last_activity_at | DATETIME | No | CURRENT_TIMESTAMP | Last request time |

### permissions

Granular permission definitions using resource:action naming convention.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Unique permission identifier (e.g., `read:media`) |
| resource | TEXT | No | - | Resource name (media, catalog, users, etc.) |
| action | TEXT | No | - | Action type (read, write, delete, manage, etc.) |
| description | TEXT | Yes | NULL | Human-readable description |

### user_permissions

Junction table for assigning individual permissions to users beyond their role.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| user_id | INTEGER | No | - | FK to users (CASCADE delete, part of composite PK) |
| permission_id | INTEGER | No | - | FK to permissions (CASCADE delete, part of composite PK) |
| granted_at | DATETIME | No | CURRENT_TIMESTAMP | When the permission was granted |
| granted_by | INTEGER | Yes | NULL | FK to users (who granted this) |

### auth_audit_log

Audit trail for all authentication-related events.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | Yes | NULL | FK to users (NULL for failed attempts with unknown user) |
| event_type | TEXT | No | - | Event: login_success, failed_login, logout, password_changed, etc. |
| ip_address | TEXT | Yes | NULL | Client IP address |
| user_agent | TEXT | Yes | NULL | Client user agent |
| details | TEXT | Yes | NULL | JSON additional context |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Event timestamp |

---

## Media Detection Tables

These tables are part of the media detection and metadata subsystem (`catalog-api/internal/media/database/schema.sql`). They use a separate encrypted SQLite database (SQLCipher).

### media_types

Enumeration of all supported media categories with detection configuration.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Unique type name (movie, tv_show, music, game, software, etc.) |
| description | TEXT | Yes | NULL | Human-readable description |
| detection_patterns | TEXT | Yes | NULL | JSON array of file glob patterns for detection |
| metadata_providers | TEXT | Yes | NULL | JSON array of external provider names |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

Pre-seeded with 40+ media types including: movie, tv_show, anime, documentary, concert, music, podcast, audiobook, pc_game, console_game, software, os, training, ebook, comic, youtube_video, archive, and more.

### media_items

Detected media items representing aggregated content entities.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_type_id | INTEGER | No | - | FK to media_types |
| title | TEXT | No | - | Media title |
| original_title | TEXT | Yes | NULL | Title in original language |
| year | INTEGER | Yes | NULL | Release year |
| description | TEXT | Yes | NULL | Synopsis/description |
| genre | TEXT | Yes | NULL | JSON array of genres |
| director | TEXT | Yes | NULL | Director name |
| cast_crew | TEXT | Yes | NULL | JSON object with cast and crew |
| rating | REAL | Yes | NULL | Aggregate rating |
| runtime | INTEGER | Yes | NULL | Runtime in minutes |
| language | TEXT | Yes | NULL | Primary language |
| country | TEXT | Yes | NULL | Country of origin |
| status | TEXT | No | 'active' | Status: active, archived, missing |
| first_detected | DATETIME | No | CURRENT_TIMESTAMP | First detection time |
| last_updated | DATETIME | No | CURRENT_TIMESTAMP | Last metadata update |

### external_metadata

Metadata fetched from external providers (IMDB, TMDB, MusicBrainz, IGDB, etc.).

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_item_id | INTEGER | No | - | FK to media_items |
| provider | TEXT | No | - | Provider name (imdb, tmdb, tvdb, musicbrainz, igdb, etc.) |
| external_id | TEXT | No | - | ID on the external provider |
| data | TEXT | No | - | JSON blob of all metadata from provider |
| rating | REAL | Yes | NULL | Provider rating |
| review_url | TEXT | Yes | NULL | Link to reviews |
| cover_url | TEXT | Yes | NULL | Cover art/poster URL |
| trailer_url | TEXT | Yes | NULL | Trailer URL |
| last_fetched | DATETIME | No | CURRENT_TIMESTAMP | Last fetch time |

Unique constraint: (media_item_id, provider)

### directory_analysis

Results of directory-level content analysis and detection.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| directory_path | TEXT | No | - | Unique directory path |
| smb_root | TEXT | No | - | Storage root identifier |
| media_item_id | INTEGER | Yes | NULL | FK to media_items (if matched) |
| confidence_score | REAL | No | - | Detection confidence 0.0 to 1.0 |
| detection_method | TEXT | No | - | Method: filename, structure, metadata, hybrid |
| analysis_data | TEXT | Yes | NULL | JSON with detection details |
| last_analyzed | DATETIME | No | CURRENT_TIMESTAMP | Analysis timestamp |
| files_count | INTEGER | No | 0 | Number of files in directory |
| total_size | INTEGER | No | 0 | Total size of directory contents |

### media_files

Individual file versions and quality variants for media items.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_item_id | INTEGER | No | - | FK to media_items |
| file_path | TEXT | No | - | Full file path |
| smb_root | TEXT | No | - | Storage root identifier |
| filename | TEXT | No | - | File name |
| file_size | INTEGER | No | - | File size in bytes |
| file_extension | TEXT | Yes | NULL | File extension |
| quality_info | TEXT | Yes | NULL | JSON: resolution, bitrate, codec, etc. |
| language | TEXT | Yes | NULL | Audio language |
| subtitle_tracks | TEXT | Yes | NULL | JSON array of subtitle tracks |
| audio_tracks | TEXT | Yes | NULL | JSON array of audio tracks |
| duration | INTEGER | Yes | NULL | Duration in seconds |
| checksum | TEXT | Yes | NULL | File checksum |
| virtual_smb_link | TEXT | Yes | NULL | Generated virtual SMB link |
| direct_smb_link | TEXT | Yes | NULL | Direct SMB path |
| last_verified | DATETIME | No | CURRENT_TIMESTAMP | Last verification time |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |

### quality_profiles

Quality tier definitions for comparing media file quality.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Unique profile name (4K, 1080p, 720p, etc.) |
| resolution_width | INTEGER | Yes | NULL | Width in pixels |
| resolution_height | INTEGER | Yes | NULL | Height in pixels |
| min_bitrate | INTEGER | Yes | NULL | Minimum bitrate in kbps |
| max_bitrate | INTEGER | Yes | NULL | Maximum bitrate in kbps |
| preferred_codecs | TEXT | Yes | NULL | JSON array of preferred codecs |
| quality_score | INTEGER | No | - | Score for ranking (higher = better) |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |

Pre-seeded profiles: 4K/UHD (100), 1080p (80), 720p (60), 480p/DVD (40), 360p (20), Audio_Lossless (90), Audio_320k (70), Audio_128k (30).

### media_collections

Groups of related media items (series, franchises, discographies).

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Collection name |
| collection_type | TEXT | No | - | Type: tv_series, movie_franchise, album_discography, game_series |
| description | TEXT | Yes | NULL | Collection description |
| total_items | INTEGER | No | 0 | Item count |
| external_ids | TEXT | Yes | NULL | JSON object with provider IDs |
| cover_url | TEXT | Yes | NULL | Cover image URL |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

### media_collection_items

Junction table linking media items to collections with ordering.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| collection_id | INTEGER | No | - | FK to media_collections |
| media_item_id | INTEGER | No | - | FK to media_items |
| sequence_number | INTEGER | Yes | NULL | Episode/track number |
| season_number | INTEGER | Yes | NULL | Season number (TV shows) |
| release_order | INTEGER | Yes | NULL | Release chronology order |

Unique constraint: (collection_id, media_item_id)

### user_metadata

Per-user preferences and watch status for media items.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_item_id | INTEGER | No | - | FK to media_items |
| user_rating | REAL | Yes | NULL | User's personal rating |
| watched_status | TEXT | Yes | NULL | Status: unwatched, watching, completed, dropped |
| watched_date | DATETIME | Yes | NULL | When the user watched it |
| personal_notes | TEXT | Yes | NULL | User notes |
| tags | TEXT | Yes | NULL | JSON array of user tags |
| favorite | BOOLEAN | No | FALSE | Favorite flag |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

### detection_rules

Configurable rules for media type detection.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_type_id | INTEGER | No | - | FK to media_types |
| rule_name | TEXT | No | - | Rule identifier |
| rule_type | TEXT | No | - | Type: filename_pattern, directory_structure, file_analysis |
| pattern | TEXT | No | - | Regex or JSON structure pattern |
| confidence_weight | REAL | No | 1.0 | Weight for confidence scoring |
| enabled | BOOLEAN | No | TRUE | Whether rule is active |
| priority | INTEGER | No | 0 | Execution priority (higher runs first) |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |

### change_log

File system change event log used by real-time watchers.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| entity_type | TEXT | No | - | Entity type: directory, file, metadata |
| entity_id | TEXT | No | - | Entity identifier (path or ID) |
| change_type | TEXT | No | - | Operation: created, updated, deleted, moved |
| old_data | TEXT | Yes | NULL | JSON previous state |
| new_data | TEXT | Yes | NULL | JSON new state |
| detected_at | DATETIME | No | CURRENT_TIMESTAMP | Detection time |
| processed | BOOLEAN | No | FALSE | Whether change has been processed |

---

## Subtitle Management Tables

### subtitle_tracks

Subtitle track records associated with media files. After migration 6, the `media_item_id` column references `files(id)` rather than `media_items(id)`.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_item_id | INTEGER | No | - | FK to files (CASCADE delete) |
| language | TEXT | No | - | Full language name |
| language_code | TEXT | No | - | ISO language code (e.g., en, fr, de) |
| source | TEXT | No | 'downloaded' | Source: downloaded, uploaded, embedded, generated |
| format | TEXT | No | 'srt' | Format: srt, ass, ssa, vtt |
| path | TEXT | Yes | NULL | File path on disk |
| content | TEXT | Yes | NULL | Inline subtitle content |
| is_default | BOOLEAN | No | FALSE | Default subtitle track |
| is_forced | BOOLEAN | No | FALSE | Forced subtitle flag |
| encoding | TEXT | No | 'utf-8' | Text encoding |
| sync_offset | REAL | No | 0.0 | Timing offset in seconds |
| verified_sync | BOOLEAN | No | FALSE | Whether sync has been verified |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update (via trigger) |

### subtitle_sync_status

Tracks asynchronous subtitle processing operations.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_item_id | INTEGER | No | - | FK to files (CASCADE delete) |
| subtitle_id | TEXT | No | - | External subtitle identifier |
| operation | TEXT | No | - | Operation type: download, upload, sync, verify |
| status | TEXT | No | 'pending' | Status: pending, in_progress, completed, failed |
| progress | INTEGER | No | 0 | Completion percentage (0-100) |
| error_message | TEXT | Yes | NULL | Error details |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Operation start |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last status update (via trigger) |
| completed_at | DATETIME | Yes | NULL | Completion time (set via trigger) |

### subtitle_cache

Temporary cache for subtitle search results to reduce external API calls.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| cache_key | TEXT | No | - | Unique cache key |
| result_id | TEXT | No | - | External result identifier |
| provider | TEXT | No | - | Provider name (opensubtitles, etc.) |
| title | TEXT | Yes | NULL | Subtitle title |
| language | TEXT | Yes | NULL | Language name |
| language_code | TEXT | Yes | NULL | ISO language code |
| download_url | TEXT | Yes | NULL | Download URL |
| format | TEXT | Yes | NULL | Subtitle format |
| encoding | TEXT | Yes | NULL | Text encoding |
| upload_date | DATETIME | Yes | NULL | When uploaded to provider |
| downloads | INTEGER | Yes | NULL | Download count |
| rating | REAL | Yes | NULL | User rating |
| comments | INTEGER | Yes | NULL | Comment count |
| match_score | REAL | Yes | NULL | Relevance score |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Cache entry time |
| expires_at | DATETIME | No | CURRENT_TIMESTAMP | Cache expiration |
| data | TEXT | Yes | NULL | JSON additional data |

### subtitle_downloads

History of subtitle downloads for tracking and deduplication.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_item_id | INTEGER | No | - | FK to files (CASCADE delete) |
| result_id | TEXT | No | - | External result ID |
| subtitle_id | TEXT | No | - | External subtitle ID |
| provider | TEXT | No | - | Provider name |
| language | TEXT | No | - | Language name |
| file_path | TEXT | Yes | NULL | Local file path |
| file_size | INTEGER | Yes | NULL | File size in bytes |
| download_url | TEXT | Yes | NULL | Source download URL |
| download_date | DATETIME | No | CURRENT_TIMESTAMP | Download time |
| verified_sync | BOOLEAN | No | FALSE | Sync verification status |
| sync_offset | REAL | No | 0.0 | Applied timing offset |

### media_subtitles

Many-to-many association between media files and subtitle tracks.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| media_item_id | INTEGER | No | - | FK to files (CASCADE delete) |
| subtitle_track_id | INTEGER | No | - | FK to subtitle_tracks (CASCADE delete) |
| is_active | BOOLEAN | No | TRUE | Whether this association is active |
| added_at | DATETIME | No | CURRENT_TIMESTAMP | Association time |

Unique constraint: (media_item_id, subtitle_track_id)

---

## Media Conversion Tables

### conversion_jobs

Media format conversion job queue and tracking.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | No | - | FK to users (CASCADE delete) |
| source_path | TEXT | No | - | Source file path |
| target_path | TEXT | No | - | Output file path |
| source_format | TEXT | No | - | Input format (mp4, avi, mp3, etc.) |
| target_format | TEXT | No | - | Output format |
| conversion_type | TEXT | No | - | Type: video, audio, document, image |
| quality | TEXT | No | 'medium' | Quality preset: low, medium, high |
| settings | TEXT | Yes | NULL | JSON additional conversion settings |
| priority | INTEGER | No | 0 | Job priority (higher = more urgent) |
| status | TEXT | No | 'pending' | Status: pending, running, completed, failed, cancelled |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Job creation time |
| started_at | DATETIME | Yes | NULL | Processing start time |
| completed_at | DATETIME | Yes | NULL | Processing completion time |
| scheduled_for | DATETIME | Yes | NULL | Deferred execution time |
| duration | INTEGER | Yes | NULL | Processing duration in seconds |
| error_message | TEXT | Yes | NULL | Error details if failed |

---

## Multi-User v3 Tables

These tables are defined in `database/schema_v3_multiuser.sql` and extend the core schema with multi-user support, analytics, favorites, sync, error reporting, logging, configuration, and performance monitoring.

### media_access_logs

Detailed media access tracking with optional location data.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | No | - | FK to users (CASCADE delete) |
| media_id | INTEGER | No | - | FK to files (CASCADE delete) |
| action | TEXT | No | - | Action: view, play, pause, stop, download, share |
| location_latitude | REAL | Yes | NULL | GPS latitude |
| location_longitude | REAL | Yes | NULL | GPS longitude |
| location_accuracy | REAL | Yes | NULL | GPS accuracy in meters |
| location_address | TEXT | Yes | NULL | Reverse-geocoded address |
| device_info | TEXT | Yes | NULL | JSON device information |
| session_id | TEXT | Yes | NULL | Session identifier |
| duration_seconds | INTEGER | Yes | NULL | For playback actions |
| position_seconds | INTEGER | Yes | NULL | Playback position |
| quality_level | TEXT | Yes | NULL | Playback quality |
| metadata | TEXT | Yes | NULL | JSON additional metadata |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Event timestamp |

### favorites

Generic favorites table supporting any entity type.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | No | - | FK to users (CASCADE delete) |
| entity_type | TEXT | No | - | Entity type: media, share, playlist, user, etc. |
| entity_id | INTEGER | No | - | ID of the favorited entity |
| category_id | INTEGER | Yes | NULL | FK to favorite_categories (SET NULL on delete) |
| notes | TEXT | Yes | NULL | User notes |
| sort_order | INTEGER | No | 0 | Custom sort order |
| is_pinned | BOOLEAN | No | 0 | Pin to top |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |

Unique constraint: (user_id, entity_type, entity_id)

### favorite_categories

Hierarchical categories for organizing favorites.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | No | - | FK to users (CASCADE delete) |
| name | TEXT | No | - | Category name |
| description | TEXT | Yes | NULL | Category description |
| color | TEXT | Yes | NULL | Hex color code |
| icon | TEXT | Yes | NULL | Icon identifier |
| parent_id | INTEGER | Yes | NULL | FK to favorite_categories (self-referencing, CASCADE delete) |
| sort_order | INTEGER | No | 0 | Display order |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |

Unique constraint: (user_id, name)

### analytics_events

Comprehensive event tracking for user behavior analytics.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | Yes | NULL | FK to users (SET NULL on delete) |
| session_id | TEXT | Yes | NULL | Session identifier |
| event_type | TEXT | No | - | Event type identifier |
| event_category | TEXT | Yes | NULL | Event category |
| event_action | TEXT | Yes | NULL | Event action |
| event_label | TEXT | Yes | NULL | Event label |
| event_value | REAL | Yes | NULL | Numeric event value |
| entity_type | TEXT | Yes | NULL | Related entity type |
| entity_id | INTEGER | Yes | NULL | Related entity ID |
| properties | TEXT | Yes | NULL | JSON properties |
| location_latitude | REAL | Yes | NULL | GPS latitude |
| location_longitude | REAL | Yes | NULL | GPS longitude |
| device_info | TEXT | Yes | NULL | JSON device information |
| user_agent | TEXT | Yes | NULL | Client user agent |
| ip_address | TEXT | Yes | NULL | Client IP address |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Event timestamp |

### analytics_reports

Scheduled and on-demand report definitions.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Report name |
| description | TEXT | Yes | NULL | Report description |
| report_type | TEXT | No | - | Type: usage, performance, user_behavior, content |
| parameters | TEXT | Yes | NULL | JSON report parameters |
| schedule_expression | TEXT | Yes | NULL | Cron expression for scheduling |
| output_format | TEXT | No | 'html' | Format: html, pdf, markdown, json |
| created_by | INTEGER | No | - | FK to users |
| is_active | BOOLEAN | No | 1 | Whether report is active |
| last_generated_at | DATETIME | Yes | NULL | Last generation time |
| next_generation_at | DATETIME | Yes | NULL | Next scheduled generation |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

### generated_reports

Generated report output files.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| report_id | INTEGER | No | - | FK to analytics_reports (CASCADE delete) |
| file_path | TEXT | No | - | Output file path |
| file_size | INTEGER | Yes | NULL | File size in bytes |
| output_format | TEXT | Yes | NULL | Output format used |
| generation_time_seconds | REAL | Yes | NULL | Time to generate |
| parameters_used | TEXT | Yes | NULL | JSON snapshot of parameters |
| row_count | INTEGER | Yes | NULL | Number of data rows |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Generation time |
| expires_at | DATETIME | Yes | NULL | Expiration time |

### conversion_profiles

Reusable conversion presets.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| name | TEXT | No | - | Profile name |
| description | TEXT | Yes | NULL | Profile description |
| input_formats | TEXT | Yes | NULL | JSON array of supported input formats |
| output_format | TEXT | No | - | Target output format |
| parameters | TEXT | No | - | JSON conversion parameters |
| is_system_profile | BOOLEAN | No | 0 | System-provided preset |
| created_by | INTEGER | Yes | NULL | FK to users |
| is_active | BOOLEAN | No | 1 | Whether profile is active |
| usage_count | INTEGER | No | 0 | Usage counter |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |

### sync_endpoints

Remote synchronization endpoint configuration.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | No | - | FK to users (CASCADE delete) |
| name | TEXT | No | - | Endpoint name |
| endpoint_type | TEXT | No | - | Type: webdav, ftp, cloud_storage |
| url | TEXT | No | - | Endpoint URL |
| credentials | TEXT | Yes | NULL | JSON encrypted credentials |
| settings | TEXT | Yes | NULL | JSON sync settings |
| is_active | BOOLEAN | No | 1 | Whether endpoint is active |
| sync_direction | TEXT | No | 'both' | Direction: upload, download, both |
| last_sync_at | DATETIME | Yes | NULL | Last sync attempt |
| last_successful_sync_at | DATETIME | Yes | NULL | Last successful sync |
| sync_status | TEXT | No | 'idle' | Status: idle, syncing, error |
| error_message | TEXT | Yes | NULL | Last error message |
| files_synced | INTEGER | No | 0 | Total files synced |
| bytes_synced | INTEGER | No | 0 | Total bytes synced |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

### sync_history

Historical record of sync operations.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| endpoint_id | INTEGER | No | - | FK to sync_endpoints (CASCADE delete) |
| sync_type | TEXT | No | - | Type: manual, scheduled, auto |
| direction | TEXT | No | - | Direction: upload, download, both |
| status | TEXT | No | - | Status: success, failed, partial |
| files_processed | INTEGER | No | 0 | Files processed |
| files_successful | INTEGER | No | 0 | Files succeeded |
| files_failed | INTEGER | No | 0 | Files failed |
| bytes_transferred | INTEGER | No | 0 | Bytes transferred |
| duration_seconds | REAL | Yes | NULL | Operation duration |
| started_at | DATETIME | No | CURRENT_TIMESTAMP | Start time |
| completed_at | DATETIME | Yes | NULL | Completion time |
| error_summary | TEXT | Yes | NULL | Error summary |
| details | TEXT | Yes | NULL | JSON detailed information |

### sync_conflicts

Sync conflict records requiring resolution.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| endpoint_id | INTEGER | No | - | FK to sync_endpoints (CASCADE delete) |
| local_file_path | TEXT | No | - | Local file path |
| remote_file_path | TEXT | No | - | Remote file path |
| conflict_type | TEXT | No | - | Type: modified_both, deleted_local, deleted_remote |
| local_file_info | TEXT | Yes | NULL | JSON local file metadata |
| remote_file_info | TEXT | Yes | NULL | JSON remote file metadata |
| resolution | TEXT | Yes | NULL | Resolution: local_wins, remote_wins, manual, skip |
| resolved_by | INTEGER | Yes | NULL | FK to users |
| resolved_at | DATETIME | Yes | NULL | Resolution time |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Detection time |

### error_reports

Comprehensive error and crash reporting.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | Yes | NULL | FK to users (SET NULL on delete) |
| session_id | TEXT | Yes | NULL | Session identifier |
| error_type | TEXT | No | - | Error type classification |
| error_level | TEXT | No | 'error' | Level: debug, info, warning, error, critical |
| error_message | TEXT | No | - | Error message |
| error_code | TEXT | Yes | NULL | Error code |
| stack_trace | TEXT | Yes | NULL | Stack trace |
| context | TEXT | Yes | NULL | JSON error context |
| device_info | TEXT | Yes | NULL | JSON device information |
| app_version | TEXT | Yes | NULL | Application version |
| os_version | TEXT | Yes | NULL | OS version |
| user_agent | TEXT | Yes | NULL | Client user agent |
| url | TEXT | Yes | NULL | Request URL |
| user_feedback | TEXT | Yes | NULL | User-submitted feedback |
| is_crash | BOOLEAN | No | 0 | Whether this was a crash |
| is_resolved | BOOLEAN | No | 0 | Resolution status |
| resolved_by | INTEGER | Yes | NULL | FK to users |
| resolved_at | DATETIME | Yes | NULL | Resolution time |
| resolution_notes | TEXT | Yes | NULL | Resolution details |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Report time |

### system_logs

Structured system logging.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| log_level | TEXT | No | - | Level: debug, info, warning, error, critical |
| component | TEXT | No | - | Component: api, android, sync, conversion, etc. |
| message | TEXT | No | - | Log message |
| context | TEXT | Yes | NULL | JSON additional context |
| user_id | INTEGER | Yes | NULL | FK to users (SET NULL on delete) |
| session_id | TEXT | Yes | NULL | Session identifier |
| file_name | TEXT | Yes | NULL | Source file name |
| line_number | INTEGER | Yes | NULL | Source line number |
| function_name | TEXT | Yes | NULL | Source function name |
| request_id | TEXT | Yes | NULL | Request trace ID |
| ip_address | TEXT | Yes | NULL | Client IP address |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Log timestamp |

### log_exports

Log export requests and status.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| requested_by | INTEGER | No | - | FK to users (CASCADE delete) |
| export_type | TEXT | No | - | Type: email, download, external_app |
| filters | TEXT | Yes | NULL | JSON export filters |
| file_path | TEXT | Yes | NULL | Export file path |
| file_size | INTEGER | Yes | NULL | File size |
| status | TEXT | No | 'pending' | Status: pending, processing, completed, failed |
| privacy_level | TEXT | No | 'sanitized' | Level: full, sanitized, minimal |
| expires_at | DATETIME | Yes | NULL | Expiration time |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Request time |
| completed_at | DATETIME | Yes | NULL | Completion time |

### system_config

System-wide configuration key-value store.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| config_key | TEXT | No | - | Unique configuration key |
| config_value | TEXT | Yes | NULL | Configuration value |
| data_type | TEXT | No | 'string' | Value type: string, integer, boolean, json |
| description | TEXT | Yes | NULL | Human-readable description |
| category | TEXT | Yes | NULL | Configuration category |
| is_system_config | BOOLEAN | No | 0 | System-level config (read-only for users) |
| requires_restart | BOOLEAN | No | 0 | Requires app restart to take effect |
| is_secret | BOOLEAN | No | 0 | Sensitive value (masked in UI) |
| updated_by | INTEGER | Yes | NULL | FK to users |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

### user_preferences

Per-user configuration key-value store.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| user_id | INTEGER | No | - | FK to users (CASCADE delete) |
| preference_key | TEXT | No | - | Preference key |
| preference_value | TEXT | Yes | NULL | Preference value |
| data_type | TEXT | No | 'string' | Value type |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Creation time |
| updated_at | DATETIME | No | CURRENT_TIMESTAMP | Last update time |

Unique constraint: (user_id, preference_key)

### performance_metrics

Performance measurement data points.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| metric_type | TEXT | No | - | Type: api_response_time, db_query_time, conversion_time |
| metric_name | TEXT | No | - | Metric identifier |
| metric_value | REAL | No | - | Measured value |
| unit | TEXT | Yes | NULL | Unit: ms, seconds, bytes, percentage |
| context | TEXT | Yes | NULL | JSON additional context |
| user_id | INTEGER | Yes | NULL | FK to users (SET NULL on delete) |
| session_id | TEXT | Yes | NULL | Session identifier |
| created_at | DATETIME | No | CURRENT_TIMESTAMP | Measurement time |

### health_checks

System health check results.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER | No | AUTOINCREMENT | Primary key |
| check_name | TEXT | No | - | Check identifier |
| status | TEXT | No | - | Status: healthy, warning, critical |
| response_time_ms | REAL | Yes | NULL | Check response time |
| details | TEXT | Yes | NULL | JSON health check details |
| checked_at | DATETIME | No | CURRENT_TIMESTAMP | Check timestamp |

---

## System Tables

### migrations

Schema version tracking table used by the migration system in `catalog-api/database/migrations.go`.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| version | INTEGER | No | - | Migration version number (PK) |
| name | TEXT | No | - | Migration name |
| applied_at | DATETIME | No | CURRENT_TIMESTAMP | When the migration was applied |

### schema_migrations

Schema version tracking for the v3 multiuser schema.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| version | TEXT | No | - | Version string (PK) |
| applied_at | DATETIME | No | CURRENT_TIMESTAMP | When applied |

---

## Index Documentation

### File Catalog Indexes

| Index Name | Table | Columns | Purpose |
|------------|-------|---------|---------|
| idx_files_storage_root_path | files | storage_root_id, path | Fast file lookup by storage root and path |
| idx_files_parent_id | files | parent_id | Efficient directory listing |
| idx_files_duplicate_group | files | duplicate_group_id | Duplicate group member lookup |
| idx_files_deleted | files | deleted | Filter active vs soft-deleted files |
| idx_files_name | files | name | File name search |
| idx_files_extension | files | extension | File type filtering |
| idx_files_file_type | files | file_type | Category filtering |
| idx_file_metadata_file_id | file_metadata | file_id | Metadata lookup by file |
| idx_scan_history_storage_root | scan_history | storage_root_id | Scan history by storage root |

### Authentication Indexes

| Index Name | Table | Columns | Purpose |
|------------|-------|---------|---------|
| idx_users_username | users | username | Fast username lookup during login |
| idx_users_email | users | email | Fast email lookup during login |
| idx_users_role_id | users | role_id | Users by role queries |
| idx_users_is_active | users | is_active | Active user filtering |
| idx_user_sessions_user_id | user_sessions | user_id | Sessions by user lookup |
| idx_user_sessions_token | user_sessions | session_token | Token validation |
| idx_user_sessions_expires_at | user_sessions | expires_at | Session expiration cleanup |

### Conversion Job Indexes

| Index Name | Table | Columns | Purpose |
|------------|-------|---------|---------|
| idx_conversion_jobs_user_id | conversion_jobs | user_id | Jobs by user |
| idx_conversion_jobs_status | conversion_jobs | status | Job queue processing |
| idx_conversion_jobs_created_at | conversion_jobs | created_at | Chronological ordering |

### Subtitle Indexes

| Index Name | Table | Columns | Purpose |
|------------|-------|---------|---------|
| idx_subtitle_tracks_media_item_id | subtitle_tracks | media_item_id | Subtitles by media file |
| idx_subtitle_tracks_language | subtitle_tracks | language | Language filtering |
| idx_subtitle_tracks_language_code | subtitle_tracks | language_code | ISO code filtering |
| idx_subtitle_tracks_source | subtitle_tracks | source | Source type filtering |
| idx_subtitle_sync_status_media_item_id | subtitle_sync_status | media_item_id | Sync status by media |
| idx_subtitle_sync_status_status | subtitle_sync_status | status | Pending operation queries |
| idx_subtitle_sync_status_operation | subtitle_sync_status | operation | Operation type filtering |
| idx_subtitle_cache_cache_key | subtitle_cache | cache_key | Cache lookup |
| idx_subtitle_cache_expires_at | subtitle_cache | expires_at | Cache expiration cleanup |
| idx_subtitle_downloads_media_item_id | subtitle_downloads | media_item_id | Downloads by media |
| idx_subtitle_downloads_result_id | subtitle_downloads | result_id | Deduplication check |
| idx_subtitle_downloads_subtitle_id | subtitle_downloads | subtitle_id | Subtitle lookup |
| idx_subtitle_downloads_provider | subtitle_downloads | provider | Provider statistics |
| idx_subtitle_downloads_language | subtitle_downloads | language | Language statistics |
| idx_subtitle_downloads_download_date | subtitle_downloads | download_date | Chronological ordering |
| idx_media_subtitles_media_item_id | media_subtitles | media_item_id | Association lookup |
| idx_media_subtitles_subtitle_track_id | media_subtitles | subtitle_track_id | Reverse association lookup |
| idx_media_subtitles_is_active | media_subtitles | is_active | Active subtitle filtering |

### Media Detection Indexes

| Index Name | Table | Columns | Purpose |
|------------|-------|---------|---------|
| idx_media_items_type | media_items | media_type_id | Media items by type |
| idx_media_items_title | media_items | title | Title search |
| idx_media_items_year | media_items | year | Year filtering |
| idx_external_metadata_provider | external_metadata | provider | Provider filtering |
| idx_media_files_size | media_files | file_size | Size-based queries |
| idx_media_files_extension | media_files | file_extension | Extension filtering |
| idx_directory_path | directory_analysis | directory_path | Directory lookup |
| idx_smb_root | directory_analysis | smb_root | Root-based queries |
| idx_media_item | directory_analysis | media_item_id | Analysis by media item |
| idx_media_item_files | media_files | media_item_id | Files by media item |
| idx_file_path | media_files | file_path | Path-based lookup |
| idx_change_type | change_log | change_type | Change type filtering |
| idx_detected_at | change_log | detected_at | Chronological queries |
| idx_processed | change_log | processed | Unprocessed change queries |

### Multi-User v3 Indexes

| Index Name | Table | Columns | Purpose |
|------------|-------|---------|---------|
| idx_media_access_logs_user_id | media_access_logs | user_id | Access logs by user |
| idx_media_access_logs_media_id | media_access_logs | media_id | Access logs by media |
| idx_media_access_logs_created_at | media_access_logs | created_at | Chronological access logs |
| idx_media_access_logs_action | media_access_logs | action | Action type filtering |
| idx_analytics_events_user_id | analytics_events | user_id | Events by user |
| idx_analytics_events_event_type | analytics_events | event_type | Event type filtering |
| idx_analytics_events_created_at | analytics_events | created_at | Chronological events |
| idx_analytics_events_entity | analytics_events | entity_type, entity_id | Entity-based event lookup |
| idx_system_logs_level | system_logs | log_level | Log level filtering |
| idx_system_logs_component | system_logs | component | Component filtering |
| idx_system_logs_created_at | system_logs | created_at | Chronological logs |
| idx_system_logs_user_id | system_logs | user_id | Logs by user |
| idx_performance_metrics_type | performance_metrics | metric_type | Metric type filtering |
| idx_performance_metrics_name | performance_metrics | metric_name | Metric name lookup |
| idx_performance_metrics_created_at | performance_metrics | created_at | Chronological metrics |

---

## Database Triggers

| Trigger Name | Table | Event | Description |
|--------------|-------|-------|-------------|
| update_subtitle_tracks_updated_at | subtitle_tracks | AFTER UPDATE | Auto-update the `updated_at` timestamp |
| update_subtitle_sync_status_updated_at | subtitle_sync_status | AFTER UPDATE | Auto-update the `updated_at` timestamp |
| set_subtitle_sync_status_completed_at | subtitle_sync_status | AFTER UPDATE | Set `completed_at` when status changes to 'completed' |
| update_users_timestamp | users | AFTER UPDATE | Auto-update `updated_at` (v3 schema) |
| update_sync_endpoints_timestamp | sync_endpoints | AFTER UPDATE | Auto-update `updated_at` (v3 schema) |
| update_system_config_timestamp | system_config | AFTER UPDATE | Auto-update `updated_at` (v3 schema) |

---

## Database Views

### media_overview (Media Detection DB)

Aggregated view of media items with file counts and quality information.

```sql
SELECT mi.id, mi.title, mi.year, mt.name as media_type,
       COUNT(mf.id) as file_count, SUM(mf.file_size) as total_size,
       MAX(mf.last_verified) as last_verified,
       GROUP_CONCAT(DISTINCT substr(mf.quality_info, 1, 20)) as available_qualities
FROM media_items mi
JOIN media_types mt ON mi.media_type_id = mt.id
LEFT JOIN media_files mf ON mi.id = mf.media_item_id
GROUP BY mi.id, mi.title, mi.year, mt.name;
```

### duplicate_media (Media Detection DB)

Identifies media items that appear to be duplicates based on title, year, and type.

### user_summary (v3 Multi-User DB)

Aggregated user information with role, access counts, and favorite counts.

### media_usage_stats (v3 Multi-User DB)

Per-media access statistics including total accesses, unique users, average duration, play counts, and download counts.

### popular_content (v3 Multi-User DB)

Ranked view of popular content combining access statistics with favorite counts.

---

## Migration History

| Version | Name | Description |
|---------|------|-------------|
| 1 | create_initial_tables | Core schema: storage_roots, files, file_metadata, duplicate_groups, virtual_paths, scan_history |
| 2 | migrate_smb_to_storage_roots | Migrates legacy smb_roots data to the unified storage_roots table |
| 3 | create_auth_tables | Authentication: users, roles, user_sessions, permissions, user_permissions, auth_audit_log |
| 4 | create_conversion_jobs_table | Media conversion job queue |
| 5 | create_subtitle_tables | Subtitle management: tracks, sync status, cache, downloads, associations |
| 6 | fix_subtitle_foreign_keys | Corrects subtitle table foreign keys to reference files instead of media_items |

For detailed migration documentation including how to create new migrations, rollback procedures, and troubleshooting, see the [SQL Migrations Guide](SQL_MIGRATIONS.md).

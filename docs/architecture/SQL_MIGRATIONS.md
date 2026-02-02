# SQL Migration Documentation

This document provides a comprehensive reference for all database migrations in the Catalogizer project, including their purpose, the tables they create or modify, and instructions for creating new migrations.

## Overview

Catalogizer uses two complementary migration systems:

1. **Go-based programmatic migrations** (primary) - Defined in `catalog-api/database/migrations.go`, these run automatically on application startup. Each migration is a Go function registered in the `RunMigrations()` method.

2. **SQL file migrations** (reference/CLI) - Stored in `catalog-api/database/migrations/` as `.up.sql` and `.down.sql` files. These can be applied manually using the `golang-migrate` CLI tool and serve as a reference for the programmatic migrations.

Additionally, there are standalone SQL schema files in `catalog-api/migrations/` and `catalog-api/internal/media/database/schema.sql` that define extended media detection schemas.

## Current Schema Version

**Version: 6** (`fix_subtitle_foreign_keys`)

The `migrations` table tracks which versions have been applied:

```sql
CREATE TABLE IF NOT EXISTS migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Migration Versions

### Version 1: `create_initial_tables`

**Source**: `catalog-api/database/migrations.go` function `createInitialTables()`
**SQL Reference**: `catalog-api/database/migrations/000001_initial_schema.up.sql`
**SQLite Variant**: `catalog-api/database/migrations/000001_initial_schema.sqlite.up.sql`

**Purpose**: Creates the foundational schema for file cataloging and storage management.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `storage_roots` | Network/local storage endpoints (SMB, FTP, NFS, WebDAV, local). Stores connection details, credentials, scan configuration, and filtering patterns. |
| `files` | Cataloged files with path, size, type, timestamps, and multiple hash columns (MD5, SHA256, SHA1, BLAKE3, quick_hash) for duplicate detection. |
| `file_metadata` | Key-value metadata pairs associated with files. Supports typed values (string default). |
| `duplicate_groups` | Groups of duplicate files identified by hash matching. Tracks count and total size. |
| `virtual_paths` | Unified path mappings across protocols. Maps a virtual path to a target entity (type + ID). |
| `scan_history` | Audit log of storage root scans. Records files processed, added, updated, deleted, and errors. |

**Indexes Created**:
- `idx_files_storage_root_path` (storage_root_id, path)
- `idx_files_parent_id` (parent_id)
- `idx_files_duplicate_group` (duplicate_group_id)
- `idx_files_deleted` (deleted)
- `idx_file_metadata_file_id` (file_id)
- `idx_scan_history_storage_root` (storage_root_id)

**Rollback**: `000001_initial_schema.down.sql` drops all tables and indexes in reverse dependency order.

---

### Version 2: `migrate_smb_to_storage_roots`

**Source**: `catalog-api/database/migrations.go` function `migrateSMBToStorageRoots()`

**Purpose**: Data migration that converts legacy `smb_roots` table data into the new multi-protocol `storage_roots` format. This migration is idempotent and skips if no `smb_roots` table exists.

**Operations**:
1. Checks if the legacy `smb_roots` table exists
2. Copies SMB root entries into `storage_roots` with `protocol = 'smb'`
3. Updates `files.storage_root_id` to reference the new storage root IDs
4. Updates `scan_history.storage_root_id` to reference the new storage root IDs

**Tables Modified**: `storage_roots`, `files`, `scan_history`

**No rollback file** - this is a one-way data migration.

---

### Version 3: `create_auth_tables`

**Source**: `catalog-api/database/migrations.go` function `createAuthTables()`
**SQL Reference**: `catalog-api/database/migrations/000003_add_user_tables.up.sql`

**Purpose**: Creates the authentication and authorization schema with role-based access control.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `users` | User accounts with username, email, password hash + salt, profile fields, activity tracking, and lockout protection. |
| `roles` | Named roles with JSON permission arrays. Seeded with Admin (`["*"]`) and User (`["media.view", "media.download"]`). |
| `user_sessions` | Active sessions with tokens, device info, IP tracking, and expiration. |
| `permissions` | Named permissions with resource/action pairs. |
| `user_permissions` | Junction table for custom per-user permission grants beyond their role. |
| `auth_audit_log` | Audit trail for authentication events (login, logout, failed login, password change). |

**Seed Data**:
- Admin role (id=1): Full permissions
- User role (id=2): View and download permissions

**Indexes Created**:
- `idx_users_username`, `idx_users_email`, `idx_users_role_id`, `idx_users_is_active`
- `idx_user_sessions_user_id`, `idx_user_sessions_token`, `idx_user_sessions_expires_at`

**Rollback**: `000003_add_user_tables.down.sql` drops `user_sessions`, `users`, and `roles`.

**Note**: The `AuthService` in `catalog-api/internal/auth/service.go` also creates its own version of auth tables (`users`, `roles`, `sessions`, `permissions`, `user_permissions`, `auth_audit_log`) with a slightly different schema. Both use `CREATE TABLE IF NOT EXISTS` to avoid conflicts.

---

### Version 4: `create_conversion_jobs_table`

**Source**: `catalog-api/database/migrations.go` function `createConversionJobsTable()`
**SQL Reference**: `catalog-api/database/migrations/000002_conversion_jobs.up.sql`

**Purpose**: Adds media format conversion job tracking.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `conversion_jobs` | Tracks media conversion tasks with source/target formats, quality settings, priority, scheduling, status, and error reporting. References `users(id)`. |

**Indexes Created**:
- `idx_conversion_jobs_user_id` (user_id)
- `idx_conversion_jobs_status` (status)
- `idx_conversion_jobs_created_at` (created_at)

**Rollback**: `000002_conversion_jobs.down.sql` drops the `conversion_jobs` table.

---

### Version 5: `create_subtitle_tables`

**Source**: `catalog-api/database/migrations.go` function `createSubtitleTables()`
**SQL Reference**: `catalog-api/database/migrations/014_create_subtitle_tables.up.sql`

**Purpose**: Creates comprehensive subtitle management tables for multi-language media support.

**Tables Created**:

| Table | Description |
|-------|-------------|
| `subtitle_tracks` | Individual subtitle tracks with language, format (SRT, VTT, ASS, etc.), encoding, sync offset, and verified status. |
| `subtitle_sync_status` | Tracks subtitle operations (download, upload, sync, verify) with progress and error reporting. |
| `subtitle_cache` | Temporary cache for subtitle search results from external providers. |
| `subtitle_downloads` | Download history with provider, language, file details, and sync verification status. |
| `media_subtitles` | Many-to-many association between media items and subtitle tracks. |

**Triggers Created**:
- `update_subtitle_tracks_updated_at` - Auto-updates `updated_at` on subtitle track changes
- `update_subtitle_sync_status_updated_at` - Auto-updates `updated_at` on sync status changes
- `set_subtitle_sync_status_completed_at` - Sets `completed_at` when status transitions to 'completed'

**Rollback**: `014_create_subtitle_tables.down.sql` drops triggers, then tables in reverse dependency order.

---

### Version 6: `fix_subtitle_foreign_keys`

**Source**: `catalog-api/database/migrations.go` function `fixSubtitleForeignKeys()`
**SQL Reference**: `catalog-api/database/migrations/015_fix_subtitle_foreign_keys.up.sql`

**Purpose**: Corrects foreign key references in subtitle tables. The original version 5 migration referenced `media_items(id)`, but the correct reference should be `files(id)`. Since SQLite does not support `ALTER CONSTRAINT`, this migration recreates all subtitle tables.

**Operations**:
1. Creates backup copies of all subtitle tables
2. Drops the original tables
3. Recreates tables with `FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE`
4. Restores data from backup tables
5. Drops backup tables
6. Recreates all triggers and indexes

**Rollback**: `015_fix_subtitle_foreign_keys.down.sql` reverses the FK change back to `media_items(id)`.

---

## Additional Schema Files

### `catalog-api/migrations/005_media_player_features.sql`

Standalone migration defining the media player schema. Not part of the programmatic migration system. Creates:

- `media_items` - Enhanced media items with playback metadata
- `subtitle_tracks` - Subtitle tracks (text PK variant)
- `audio_tracks` - Multi-language audio track management
- `chapters` - Video chapters and bookmarks
- `cover_art` - Cover art from multiple sources (embedded, MusicBrainz, Last.fm, Spotify, etc.)
- `lyrics_data` - Lyrics with sync data and translations
- `playback_sessions` - Current playback state tracking
- `playlists` and `playlist_items` - Playlist management
- `user_preferences` - Key-value user preferences
- `translation_cache` - Translated text caching
- `external_api_cache` - External API response caching
- `media_analysis_queue` - Background processing queue
- `language_preferences` - Per-content-type language settings
- `supported_languages` - Language registry with 20 pre-loaded languages
- Views: `media_with_metadata`, `playlists_with_items`

### `catalog-api/migrations/006_media_items_schema_update.sql`

Updates `media_items` for Android TV compatibility. Adds columns (`directory_path`, `smb_path`, `external_metadata`, `versions`, `watch_progress`, `last_watched`, `is_downloaded`), recreates the table with a simplified schema, and migrates existing data.

### `catalog-api/internal/media/database/schema.sql`

Media detection and metadata database schema. Defines:

- `media_types` - 40+ media type classifications with detection patterns
- `media_items` - Detected media with external metadata references
- `external_metadata` - Provider-specific metadata (IMDB, TMDB, TVDB, etc.)
- `directory_analysis` - Directory scanning results with confidence scores
- `media_files` - Individual file versions and quality info
- `quality_profiles` - Quality ranking definitions (4K through Audio_128k)
- `change_log` - Real-time change tracking
- `media_collections` and `media_collection_items` - Series/franchise grouping
- `user_metadata` - User ratings, watch status, tags, favorites
- `detection_rules` - Configurable media type detection patterns
- Views: `media_overview`, `duplicate_media`

## How Migrations Are Applied

### Automatic (Default)

Migrations run automatically when the application starts via `main.go`:

```go
// In catalog-api/main.go
log.Println("Running database migrations...")
if err := databaseDB.RunMigrations(ctx); err != nil {
    log.Fatal("Failed to run database migrations:", err)
}
log.Println("Database migrations completed successfully")
```

The `RunMigrations()` function in `catalog-api/database/migrations.go`:
1. Creates the `migrations` tracking table if it does not exist
2. Iterates through all registered migrations (versions 1-6)
3. For each migration, checks if the version exists in the `migrations` table
4. Skips already-applied migrations
5. Executes the migration's `Up` function
6. Records the version and name in the `migrations` table

### Manual (CLI)

For manual control, use the `golang-migrate` CLI with the SQL files:

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

# Force a specific version (for recovering from dirty state)
migrate -path catalog-api/database/migrations \
    -database "sqlite3://./catalogizer.db" force VERSION
```

### Docker

Migrations run automatically when the Docker container starts, as part of the normal application startup sequence.

## How to Create a New Migration

### Step 1: Define the Go migration function

Add a new function to `catalog-api/database/migrations.go`:

```go
func (db *DB) createMyNewTable(ctx context.Context) error {
    schema := `
    CREATE TABLE IF NOT EXISTS my_new_table (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_my_new_table_name ON my_new_table(name);
    `
    if _, err := db.ExecContext(ctx, schema); err != nil {
        return fmt.Errorf("failed to create my_new_table: %w", err)
    }
    return nil
}
```

### Step 2: Register the migration

Add the migration to the `migrations` slice in `RunMigrations()`:

```go
migrations := []Migration{
    // ... existing migrations ...
    {
        Version: 7,  // Next version number
        Name:    "create_my_new_table",
        Up:      db.createMyNewTable,
    },
}
```

### Step 3: Create SQL reference files (optional but recommended)

```bash
cd catalog-api/database/migrations

# Create up migration
touch 000007_create_my_new_table.up.sql

# Create down migration (rollback)
touch 000007_create_my_new_table.down.sql

# If SQLite syntax differs from PostgreSQL
touch 000007_create_my_new_table.sqlite.up.sql
```

### Step 4: Follow migration guidelines

1. **Always use `IF NOT EXISTS`** / `IF EXISTS` for idempotency
2. **Create both up and down migrations** - The Go function handles "up"; write corresponding `.down.sql` for CLI rollback
3. **Never modify existing migrations** that have been deployed
4. **Use transactions where possible** for atomicity
5. **Test on both SQLite and PostgreSQL** if supporting both
6. **Document complex migrations** with SQL comments
7. **Test rollbacks** before deploying

### SQLite vs PostgreSQL Differences

| Feature | PostgreSQL | SQLite |
|---------|------------|--------|
| Auto-increment | `SERIAL` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| Boolean | `BOOLEAN` | `INTEGER` (0/1) |
| Timestamp | `TIMESTAMP` | `DATETIME` |
| Big integers | `BIGINT` | `INTEGER` |
| ALTER TABLE | Full support | Limited (no DROP COLUMN, no ALTER CONSTRAINT) |
| Current time | `CURRENT_TIMESTAMP` | `CURRENT_TIMESTAMP` |

When SQLite limitations require a different approach (e.g., changing foreign keys), use the backup-recreate-restore pattern as demonstrated in migration version 6.

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
# Using CLI
migrate -path catalog-api/database/migrations \
    -database "sqlite3://./catalogizer.db" drop
migrate -path catalog-api/database/migrations \
    -database "sqlite3://./catalogizer.db" up

# Or simply delete the SQLite file and restart the application
rm catalogizer.db
go run main.go
```

**WARNING**: This deletes all data. Only use in development environments.

### Checking Applied Migrations

Query the migrations table directly:

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

## Related Documentation

- [Database Schema](DATABASE_SCHEMA.md) - Complete table and index reference
- [Architecture Overview](ARCHITECTURE.md) - System design and component interactions
- [Auth Flow](AUTH_FLOW.md) - Authentication system details
- [Deployment Guide](../DEPLOYMENT_GUIDE.md) - Production deployment instructions
- [Backup and Recovery](../deployment/BACKUP_AND_RECOVERY.md) - Database backup procedures

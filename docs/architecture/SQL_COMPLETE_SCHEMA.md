# Complete SQL Schema Reference

This document contains the complete CREATE TABLE statements for all Catalogizer database tables, organized by database engine (SQLite and PostgreSQL) with all migrations applied in order.

## Table of Contents

- [SQLite Complete Schema](#sqlite-complete-schema)
- [PostgreSQL Complete Schema](#postgresql-complete-schema)
- [Migration Application Order](#migration-application-order)

---

## SQLite Complete Schema

This is the final-state SQLite schema after all 6 migrations have been applied, plus the media detection database and v3 multiuser schema.

### Migration Tracking

```sql
-- Migration version tracking (created before any migrations run)
CREATE TABLE IF NOT EXISTS migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Migration 1: Core File Catalog

```sql
-- Storage roots table (replaces legacy smb_roots)
CREATE TABLE IF NOT EXISTS storage_roots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    protocol TEXT NOT NULL,
    host TEXT,
    port INTEGER,
    path TEXT,
    username TEXT,
    password TEXT,
    domain TEXT,
    mount_point TEXT,
    options TEXT,
    url TEXT,
    enabled BOOLEAN DEFAULT 1,
    max_depth INTEGER DEFAULT 10,
    enable_duplicate_detection BOOLEAN DEFAULT 1,
    enable_metadata_extraction BOOLEAN DEFAULT 1,
    include_patterns TEXT,
    exclude_patterns TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_scan_at DATETIME
);

-- Duplicate groups table (must be created before files due to FK)
CREATE TABLE IF NOT EXISTS duplicate_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_count INTEGER DEFAULT 0,
    total_size INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Files table
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storage_root_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL,
    extension TEXT,
    mime_type TEXT,
    file_type TEXT,
    size INTEGER NOT NULL,
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
    quick_hash TEXT,
    is_duplicate BOOLEAN DEFAULT 0,
    duplicate_group_id INTEGER,
    parent_id INTEGER,
    FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id),
    FOREIGN KEY (parent_id) REFERENCES files(id),
    FOREIGN KEY (duplicate_group_id) REFERENCES duplicate_groups(id)
);

-- File metadata table
CREATE TABLE IF NOT EXISTS file_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    data_type TEXT DEFAULT 'string',
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Virtual paths table
CREATE TABLE IF NOT EXISTS virtual_paths (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Scan history table
CREATE TABLE IF NOT EXISTS scan_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storage_root_id INTEGER NOT NULL,
    scan_type TEXT NOT NULL,
    status TEXT NOT NULL,
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

-- Indexes for migration 1
CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path);
CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id);
CREATE INDEX IF NOT EXISTS idx_files_duplicate_group ON files(duplicate_group_id);
CREATE INDEX IF NOT EXISTS idx_files_deleted ON files(deleted);
CREATE INDEX IF NOT EXISTS idx_files_name ON files(name);
CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension);
CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type);
CREATE INDEX IF NOT EXISTS idx_file_metadata_file_id ON file_metadata(file_id);
CREATE INDEX IF NOT EXISTS idx_scan_history_storage_root ON scan_history(storage_root_id);
```

### Migration 2: SMB to Storage Roots Data Migration

```sql
-- No new tables. Migrates data from legacy smb_roots to storage_roots.
-- See catalog-api/database/migrations.go for the data migration logic.
```

### Migration 3: Authentication Tables

```sql
-- Users table
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
    settings TEXT DEFAULT '{}',
    is_active INTEGER DEFAULT 1,
    is_locked INTEGER DEFAULT 0,
    locked_until DATETIME,
    failed_login_attempts INTEGER DEFAULT 0,
    last_login_at DATETIME,
    last_login_ip TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT DEFAULT '[]',
    is_system INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    refresh_token TEXT,
    device_info TEXT,
    ip_address TEXT,
    user_agent TEXT,
    is_active INTEGER DEFAULT 1,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT
);

-- User permissions junction table
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

-- Authentication audit log
CREATE TABLE IF NOT EXISTS auth_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    event_type TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    details TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Seed default roles
INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
VALUES (1, 'Admin', 'Administrator role with all permissions', '["*"]', 1);

INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
VALUES (2, 'User', 'Standard user role', '["media.view", "media.download"]', 1);

-- Indexes for migration 3
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
```

### Migration 4: Conversion Jobs

```sql
-- Conversion jobs table
CREATE TABLE IF NOT EXISTS conversion_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    source_path TEXT NOT NULL,
    target_path TEXT NOT NULL,
    source_format TEXT NOT NULL,
    target_format TEXT NOT NULL,
    conversion_type TEXT NOT NULL,
    quality TEXT DEFAULT 'medium',
    settings TEXT,
    priority INTEGER DEFAULT 0,
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    scheduled_for DATETIME,
    duration INTEGER,
    error_message TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for migration 4
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_user_id ON conversion_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_status ON conversion_jobs(status);
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_created_at ON conversion_jobs(created_at);
```

### Migration 5 + 6: Subtitle Tables (Final State After FK Fix)

```sql
-- Subtitle tracks (FK references files after migration 6)
CREATE TABLE subtitle_tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    language TEXT NOT NULL,
    language_code TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'downloaded',
    format TEXT NOT NULL DEFAULT 'srt',
    path TEXT,
    content TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    is_forced BOOLEAN DEFAULT FALSE,
    encoding TEXT DEFAULT 'utf-8',
    sync_offset REAL DEFAULT 0.0,
    verified_sync BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Subtitle sync status (FK references files after migration 6)
CREATE TABLE subtitle_sync_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    subtitle_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (media_item_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Subtitle search result cache
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
    data TEXT
);

-- Subtitle download history (FK references files after migration 6)
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

-- Media-subtitle association (FK references files after migration 6)
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

-- Subtitle indexes
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item_id ON subtitle_tracks(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language_code ON subtitle_tracks(language_code);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_source ON subtitle_tracks(source);
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_media_item_id ON subtitle_sync_status(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_status ON subtitle_sync_status(status);
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_operation ON subtitle_sync_status(operation);
CREATE INDEX IF NOT EXISTS idx_subtitle_cache_cache_key ON subtitle_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_subtitle_cache_expires_at ON subtitle_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_media_item_id ON subtitle_downloads(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_result_id ON subtitle_downloads(result_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_subtitle_id ON subtitle_downloads(subtitle_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_provider ON subtitle_downloads(provider);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_language ON subtitle_downloads(language);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_download_date ON subtitle_downloads(download_date);
CREATE INDEX IF NOT EXISTS idx_media_subtitles_media_item_id ON media_subtitles(media_item_id);
CREATE INDEX IF NOT EXISTS idx_media_subtitles_subtitle_track_id ON media_subtitles(subtitle_track_id);
CREATE INDEX IF NOT EXISTS idx_media_subtitles_is_active ON media_subtitles(is_active);

-- Subtitle triggers
CREATE TRIGGER IF NOT EXISTS update_subtitle_tracks_updated_at
    AFTER UPDATE ON subtitle_tracks
    FOR EACH ROW
BEGIN
    UPDATE subtitle_tracks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_subtitle_sync_status_updated_at
    AFTER UPDATE ON subtitle_sync_status
    FOR EACH ROW
BEGIN
    UPDATE subtitle_sync_status SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS set_subtitle_sync_status_completed_at
    AFTER UPDATE ON subtitle_sync_status
    FOR EACH ROW
    WHEN NEW.status = 'completed' AND OLD.status != 'completed'
BEGIN
    UPDATE subtitle_sync_status SET completed_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
```

### Media Detection Database (Separate Encrypted SQLCipher DB)

```sql
PRAGMA foreign_keys = ON;

-- Media types enumeration
CREATE TABLE IF NOT EXISTS media_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    detection_patterns TEXT,
    metadata_providers TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Detected media items
CREATE TABLE IF NOT EXISTS media_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_type_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    original_title TEXT,
    year INTEGER,
    description TEXT,
    genre TEXT,
    director TEXT,
    cast_crew TEXT,
    rating REAL,
    runtime INTEGER,
    language TEXT,
    country TEXT,
    status TEXT DEFAULT 'active',
    first_detected DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_type_id) REFERENCES media_types(id)
);

-- External provider metadata
CREATE TABLE IF NOT EXISTS external_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    provider TEXT NOT NULL,
    external_id TEXT NOT NULL,
    data TEXT NOT NULL,
    rating REAL,
    review_url TEXT,
    cover_url TEXT,
    trailer_url TEXT,
    last_fetched DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id),
    UNIQUE(media_item_id, provider)
);

-- Directory analysis results
CREATE TABLE IF NOT EXISTS directory_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    directory_path TEXT NOT NULL UNIQUE,
    smb_root TEXT NOT NULL,
    media_item_id INTEGER,
    confidence_score REAL NOT NULL,
    detection_method TEXT NOT NULL,
    analysis_data TEXT,
    last_analyzed DATETIME DEFAULT CURRENT_TIMESTAMP,
    files_count INTEGER DEFAULT 0,
    total_size INTEGER DEFAULT 0,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id)
);

-- Individual media file versions
CREATE TABLE IF NOT EXISTS media_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    smb_root TEXT NOT NULL,
    filename TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    file_extension TEXT,
    quality_info TEXT,
    language TEXT,
    subtitle_tracks TEXT,
    audio_tracks TEXT,
    duration INTEGER,
    checksum TEXT,
    virtual_smb_link TEXT,
    direct_smb_link TEXT,
    last_verified DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id)
);

-- Quality profile definitions
CREATE TABLE IF NOT EXISTS quality_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    resolution_width INTEGER,
    resolution_height INTEGER,
    min_bitrate INTEGER,
    max_bitrate INTEGER,
    preferred_codecs TEXT,
    quality_score INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Change event log
CREATE TABLE IF NOT EXISTS change_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    change_type TEXT NOT NULL,
    old_data TEXT,
    new_data TEXT,
    detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT FALSE
);

-- Media collections
CREATE TABLE IF NOT EXISTS media_collections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    collection_type TEXT NOT NULL,
    description TEXT,
    total_items INTEGER DEFAULT 0,
    external_ids TEXT,
    cover_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Collection membership
CREATE TABLE IF NOT EXISTS media_collection_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_id INTEGER NOT NULL,
    media_item_id INTEGER NOT NULL,
    sequence_number INTEGER,
    season_number INTEGER,
    release_order INTEGER,
    FOREIGN KEY (collection_id) REFERENCES media_collections(id),
    FOREIGN KEY (media_item_id) REFERENCES media_items(id),
    UNIQUE(collection_id, media_item_id)
);

-- User preferences per media item
CREATE TABLE IF NOT EXISTS user_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    user_rating REAL,
    watched_status TEXT,
    watched_date DATETIME,
    personal_notes TEXT,
    tags TEXT,
    favorite BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id)
);

-- Detection rules
CREATE TABLE IF NOT EXISTS detection_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_type_id INTEGER NOT NULL,
    rule_name TEXT NOT NULL,
    rule_type TEXT NOT NULL,
    pattern TEXT NOT NULL,
    confidence_weight REAL DEFAULT 1.0,
    enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_type_id) REFERENCES media_types(id)
);

-- Media detection indexes
CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(media_type_id);
CREATE INDEX IF NOT EXISTS idx_media_items_title ON media_items(title);
CREATE INDEX IF NOT EXISTS idx_media_items_year ON media_items(year);
CREATE INDEX IF NOT EXISTS idx_external_metadata_provider ON external_metadata(provider);
CREATE INDEX IF NOT EXISTS idx_media_files_size ON media_files(file_size);
CREATE INDEX IF NOT EXISTS idx_media_files_extension ON media_files(file_extension);

-- Views
CREATE VIEW IF NOT EXISTS media_overview AS
SELECT
    mi.id, mi.title, mi.year, mt.name as media_type,
    COUNT(mf.id) as file_count, SUM(mf.file_size) as total_size,
    MAX(mf.last_verified) as last_verified,
    GROUP_CONCAT(DISTINCT substr(mf.quality_info, 1, 20)) as available_qualities
FROM media_items mi
JOIN media_types mt ON mi.media_type_id = mt.id
LEFT JOIN media_files mf ON mi.id = mf.media_item_id
GROUP BY mi.id, mi.title, mi.year, mt.name;

CREATE VIEW IF NOT EXISTS duplicate_media AS
SELECT
    mi1.title, mi1.year, mt.name as media_type,
    COUNT(*) as duplicate_count,
    GROUP_CONCAT(mi1.id) as media_item_ids
FROM media_items mi1
JOIN media_types mt ON mi1.media_type_id = mt.id
WHERE EXISTS (
    SELECT 1 FROM media_items mi2
    WHERE mi2.title = mi1.title AND mi2.year = mi1.year
    AND mi2.media_type_id = mi1.media_type_id AND mi2.id != mi1.id
)
GROUP BY mi1.title, mi1.year, mi1.media_type_id
HAVING COUNT(*) > 1;
```

### v3 Multi-User Schema Extensions

```sql
PRAGMA foreign_keys = ON;

-- Extended roles (supplements migration 3 roles)
INSERT OR IGNORE INTO roles (name, display_name, description, permissions, is_system_role) VALUES
('super_admin', 'Super Administrator', 'Full system access with all permissions', '["*"]', 1),
('manager', 'Manager', 'Manage users and content', '["user.view", "media.manage", "share.create", "analytics.view"]', 1),
('guest', 'Guest User', 'Limited read-only access', '["media.view"]', 1);

-- Media access logs
CREATE TABLE IF NOT EXISTS media_access_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    media_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    location_latitude REAL,
    location_longitude REAL,
    location_accuracy REAL,
    location_address TEXT,
    device_info TEXT,
    session_id VARCHAR(255),
    duration_seconds INTEGER,
    position_seconds INTEGER,
    quality_level VARCHAR(20),
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (media_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Favorite categories (must be before favorites due to FK)
CREATE TABLE IF NOT EXISTS favorite_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7),
    icon VARCHAR(50),
    parent_id INTEGER,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES favorite_categories(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

-- Favorites
CREATE TABLE IF NOT EXISTS favorites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    category_id INTEGER,
    notes TEXT,
    sort_order INTEGER DEFAULT 0,
    is_pinned BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES favorite_categories(id) ON DELETE SET NULL,
    UNIQUE(user_id, entity_type, entity_id)
);

-- Analytics events
CREATE TABLE IF NOT EXISTS analytics_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    session_id VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50),
    event_action VARCHAR(50),
    event_label VARCHAR(100),
    event_value REAL,
    entity_type VARCHAR(50),
    entity_id INTEGER,
    properties TEXT,
    location_latitude REAL,
    location_longitude REAL,
    device_info TEXT,
    user_agent TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Analytics reports
CREATE TABLE IF NOT EXISTS analytics_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    report_type VARCHAR(50) NOT NULL,
    parameters TEXT,
    schedule_expression VARCHAR(100),
    output_format VARCHAR(20) DEFAULT 'html',
    created_by INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    last_generated_at TIMESTAMP,
    next_generation_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Generated reports
CREATE TABLE IF NOT EXISTS generated_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER,
    output_format VARCHAR(20),
    generation_time_seconds REAL,
    parameters_used TEXT,
    row_count INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    FOREIGN KEY (report_id) REFERENCES analytics_reports(id) ON DELETE CASCADE
);

-- Conversion profiles
CREATE TABLE IF NOT EXISTS conversion_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    input_formats TEXT,
    output_format VARCHAR(50) NOT NULL,
    parameters TEXT NOT NULL,
    is_system_profile BOOLEAN DEFAULT 0,
    created_by INTEGER,
    is_active BOOLEAN DEFAULT 1,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Sync endpoints
CREATE TABLE IF NOT EXISTS sync_endpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name VARCHAR(200) NOT NULL,
    endpoint_type VARCHAR(50) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    credentials TEXT,
    settings TEXT,
    is_active BOOLEAN DEFAULT 1,
    sync_direction VARCHAR(20) DEFAULT 'both',
    last_sync_at TIMESTAMP,
    last_successful_sync_at TIMESTAMP,
    sync_status VARCHAR(50) DEFAULT 'idle',
    error_message TEXT,
    files_synced INTEGER DEFAULT 0,
    bytes_synced INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Sync history
CREATE TABLE IF NOT EXISTS sync_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    endpoint_id INTEGER NOT NULL,
    sync_type VARCHAR(20) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    files_processed INTEGER DEFAULT 0,
    files_successful INTEGER DEFAULT 0,
    files_failed INTEGER DEFAULT 0,
    bytes_transferred INTEGER DEFAULT 0,
    duration_seconds REAL,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    error_summary TEXT,
    details TEXT,
    FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id) ON DELETE CASCADE
);

-- Sync conflicts
CREATE TABLE IF NOT EXISTS sync_conflicts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    endpoint_id INTEGER NOT NULL,
    local_file_path VARCHAR(1000) NOT NULL,
    remote_file_path VARCHAR(1000) NOT NULL,
    conflict_type VARCHAR(50) NOT NULL,
    local_file_info TEXT,
    remote_file_info TEXT,
    resolution VARCHAR(50),
    resolved_by INTEGER,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id) ON DELETE CASCADE,
    FOREIGN KEY (resolved_by) REFERENCES users(id)
);

-- Error reports
CREATE TABLE IF NOT EXISTS error_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    session_id VARCHAR(255),
    error_type VARCHAR(100) NOT NULL,
    error_level VARCHAR(20) DEFAULT 'error',
    error_message TEXT NOT NULL,
    error_code VARCHAR(50),
    stack_trace TEXT,
    context TEXT,
    device_info TEXT,
    app_version VARCHAR(50),
    os_version VARCHAR(50),
    user_agent TEXT,
    url VARCHAR(1000),
    user_feedback TEXT,
    is_crash BOOLEAN DEFAULT 0,
    is_resolved BOOLEAN DEFAULT 0,
    resolved_by INTEGER,
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (resolved_by) REFERENCES users(id)
);

-- System logs
CREATE TABLE IF NOT EXISTS system_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_level VARCHAR(20) NOT NULL,
    component VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    context TEXT,
    user_id INTEGER,
    session_id VARCHAR(255),
    file_name VARCHAR(255),
    line_number INTEGER,
    function_name VARCHAR(255),
    request_id VARCHAR(255),
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Log exports
CREATE TABLE IF NOT EXISTS log_exports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    requested_by INTEGER NOT NULL,
    export_type VARCHAR(50) NOT NULL,
    filters TEXT,
    file_path VARCHAR(500),
    file_size INTEGER,
    status VARCHAR(20) DEFAULT 'pending',
    privacy_level VARCHAR(20) DEFAULT 'sanitized',
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (requested_by) REFERENCES users(id) ON DELETE CASCADE
);

-- System configuration
CREATE TABLE IF NOT EXISTS system_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key VARCHAR(200) NOT NULL UNIQUE,
    config_value TEXT,
    data_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    category VARCHAR(100),
    is_system_config BOOLEAN DEFAULT 0,
    requires_restart BOOLEAN DEFAULT 0,
    is_secret BOOLEAN DEFAULT 0,
    updated_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (updated_by) REFERENCES users(id)
);

-- User preferences
CREATE TABLE IF NOT EXISTS user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    preference_key VARCHAR(200) NOT NULL,
    preference_value TEXT,
    data_type VARCHAR(20) DEFAULT 'string',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, preference_key)
);

-- Performance metrics
CREATE TABLE IF NOT EXISTS performance_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type VARCHAR(100) NOT NULL,
    metric_name VARCHAR(200) NOT NULL,
    metric_value REAL NOT NULL,
    unit VARCHAR(20),
    context TEXT,
    user_id INTEGER,
    session_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Health checks
CREATE TABLE IF NOT EXISTS health_checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    check_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    response_time_ms REAL,
    details TEXT,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- v3 indexes
CREATE INDEX IF NOT EXISTS idx_media_access_logs_user_id ON media_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_media_access_logs_media_id ON media_access_logs(media_id);
CREATE INDEX IF NOT EXISTS idx_media_access_logs_created_at ON media_access_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_media_access_logs_action ON media_access_logs(action);
CREATE INDEX IF NOT EXISTS idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_event_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at);
CREATE INDEX IF NOT EXISTS idx_analytics_events_entity ON analytics_events(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_system_logs_level ON system_logs(log_level);
CREATE INDEX IF NOT EXISTS idx_system_logs_component ON system_logs(component);
CREATE INDEX IF NOT EXISTS idx_system_logs_created_at ON system_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_system_logs_user_id ON system_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_type ON performance_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_name ON performance_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_created_at ON performance_metrics(created_at);

-- v3 triggers
CREATE TRIGGER IF NOT EXISTS update_users_timestamp
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_sync_endpoints_timestamp
AFTER UPDATE ON sync_endpoints
BEGIN
    UPDATE sync_endpoints SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_system_config_timestamp
AFTER UPDATE ON system_config
BEGIN
    UPDATE system_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- v3 views
CREATE VIEW IF NOT EXISTS user_summary AS
SELECT u.id, u.username, u.email, u.display_name, r.name as role_name,
       r.display_name as role_display_name, u.is_active, u.last_login_at,
       COUNT(DISTINCT mal.id) as total_media_accesses,
       COUNT(DISTINCT f.id) as total_favorites, u.created_at
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
LEFT JOIN media_access_logs mal ON u.id = mal.user_id
LEFT JOIN favorites f ON u.id = f.user_id
GROUP BY u.id;

CREATE VIEW IF NOT EXISTS media_usage_stats AS
SELECT mal.media_id, COUNT(*) as total_accesses,
       COUNT(DISTINCT mal.user_id) as unique_users,
       AVG(mal.duration_seconds) as avg_duration,
       MAX(mal.created_at) as last_accessed,
       COUNT(CASE WHEN mal.action = 'play' THEN 1 END) as play_count,
       COUNT(CASE WHEN mal.action = 'download' THEN 1 END) as download_count
FROM media_access_logs mal
GROUP BY mal.media_id;

CREATE VIEW IF NOT EXISTS popular_content AS
SELECT mus.media_id, mus.total_accesses, mus.unique_users,
       mus.play_count, mus.last_accessed,
       COUNT(f.id) as favorite_count
FROM media_usage_stats mus
LEFT JOIN favorites f ON f.entity_type = 'media' AND f.entity_id = mus.media_id
GROUP BY mus.media_id
ORDER BY mus.total_accesses DESC, mus.unique_users DESC;

INSERT OR IGNORE INTO schema_migrations (version) VALUES ('3.0.0_multiuser_complete');
```

---

## PostgreSQL Complete Schema

The PostgreSQL schema uses `SERIAL` instead of `AUTOINCREMENT`, `TIMESTAMP` instead of `DATETIME`, `TRUE/FALSE` instead of `1/0`, and supports `ALTER TABLE ... ADD CONSTRAINT` for foreign keys.

### Core Tables (PostgreSQL)

```sql
-- Storage roots
CREATE TABLE IF NOT EXISTS storage_roots (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    protocol TEXT NOT NULL,
    host TEXT,
    port INTEGER,
    path TEXT,
    username TEXT,
    password TEXT,
    domain TEXT,
    mount_point TEXT,
    options TEXT,
    url TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    max_depth INTEGER DEFAULT 10,
    enable_duplicate_detection BOOLEAN DEFAULT TRUE,
    enable_metadata_extraction BOOLEAN DEFAULT TRUE,
    include_patterns TEXT,
    exclude_patterns TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_scan_at TIMESTAMP
);

-- Duplicate groups
CREATE TABLE IF NOT EXISTS duplicate_groups (
    id SERIAL PRIMARY KEY,
    file_count INTEGER DEFAULT 0,
    total_size BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Files
CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    storage_root_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL,
    extension TEXT,
    mime_type TEXT,
    file_type TEXT,
    size BIGINT NOT NULL,
    is_directory BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP NOT NULL,
    accessed_at TIMESTAMP,
    deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP,
    last_scan_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_verified_at TIMESTAMP,
    md5 TEXT,
    sha256 TEXT,
    sha1 TEXT,
    blake3 TEXT,
    quick_hash TEXT,
    is_duplicate BOOLEAN DEFAULT FALSE,
    duplicate_group_id INTEGER,
    parent_id INTEGER,
    FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id),
    FOREIGN KEY (parent_id) REFERENCES files(id)
);

ALTER TABLE files ADD CONSTRAINT fk_duplicate_group
    FOREIGN KEY (duplicate_group_id) REFERENCES duplicate_groups(id);

-- File metadata
CREATE TABLE IF NOT EXISTS file_metadata (
    id SERIAL PRIMARY KEY,
    file_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    data_type TEXT DEFAULT 'string',
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Virtual paths
CREATE TABLE IF NOT EXISTS virtual_paths (
    id SERIAL PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Scan history
CREATE TABLE IF NOT EXISTS scan_history (
    id SERIAL PRIMARY KEY,
    storage_root_id INTEGER NOT NULL,
    scan_type TEXT NOT NULL,
    status TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    files_processed INTEGER DEFAULT 0,
    files_added INTEGER DEFAULT 0,
    files_updated INTEGER DEFAULT 0,
    files_deleted INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    error_message TEXT,
    FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
);

-- Roles
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT DEFAULT '[]',
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    salt TEXT NOT NULL,
    role_id INTEGER NOT NULL REFERENCES roles(id),
    first_name TEXT,
    last_name TEXT,
    display_name TEXT,
    avatar_url TEXT,
    time_zone TEXT,
    language TEXT,
    settings TEXT DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    is_locked BOOLEAN DEFAULT FALSE,
    locked_until TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    last_login_at TIMESTAMP,
    last_login_ip TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User sessions
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token TEXT NOT NULL UNIQUE,
    refresh_token TEXT,
    device_info TEXT,
    ip_address TEXT,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Permissions
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT
);

-- User permissions
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by INTEGER REFERENCES users(id),
    PRIMARY KEY (user_id, permission_id)
);

-- Auth audit log
CREATE TABLE IF NOT EXISTS auth_audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    event_type TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Conversion jobs
CREATE TABLE IF NOT EXISTS conversion_jobs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_path TEXT NOT NULL,
    target_path TEXT NOT NULL,
    source_format TEXT NOT NULL,
    target_format TEXT NOT NULL,
    conversion_type TEXT NOT NULL,
    quality TEXT DEFAULT 'medium',
    settings TEXT,
    priority INTEGER DEFAULT 0,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    scheduled_for TIMESTAMP,
    duration INTEGER,
    error_message TEXT
);

-- Subtitle tracks
CREATE TABLE IF NOT EXISTS subtitle_tracks (
    id SERIAL PRIMARY KEY,
    media_item_id INTEGER NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    language TEXT NOT NULL,
    language_code TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'downloaded',
    format TEXT NOT NULL DEFAULT 'srt',
    path TEXT,
    content TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    is_forced BOOLEAN DEFAULT FALSE,
    encoding TEXT DEFAULT 'utf-8',
    sync_offset REAL DEFAULT 0.0,
    verified_sync BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- PostgreSQL indexes (same as SQLite)
CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path);
CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id);
CREATE INDEX IF NOT EXISTS idx_files_duplicate_group ON files(duplicate_group_id);
CREATE INDEX IF NOT EXISTS idx_files_deleted ON files(deleted);
CREATE INDEX IF NOT EXISTS idx_files_name ON files(name);
CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension);
CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type);
CREATE INDEX IF NOT EXISTS idx_file_metadata_file_id ON file_metadata(file_id);
CREATE INDEX IF NOT EXISTS idx_scan_history_storage_root ON scan_history(storage_root_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_user_id ON conversion_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_status ON conversion_jobs(status);
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_created_at ON conversion_jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item_id ON subtitle_tracks(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language);

-- PostgreSQL updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subtitle_tracks_updated_at BEFORE UPDATE ON subtitle_tracks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

## Migration Application Order

The migrations are applied in strict version order by `catalog-api/database/migrations.go`:

| Order | Version | Name | Tables Created |
|-------|---------|------|----------------|
| 1 | 1 | create_initial_tables | storage_roots, files, file_metadata, duplicate_groups, virtual_paths, scan_history |
| 2 | 2 | migrate_smb_to_storage_roots | (data migration only) |
| 3 | 3 | create_auth_tables | users, roles, user_sessions, permissions, user_permissions, auth_audit_log |
| 4 | 4 | create_conversion_jobs_table | conversion_jobs |
| 5 | 5 | create_subtitle_tables | subtitle_tracks, subtitle_sync_status, subtitle_cache, subtitle_downloads, media_subtitles |
| 6 | 6 | fix_subtitle_foreign_keys | (recreates subtitle tables with corrected FKs) |

Additional schemas applied separately:
- **Media Detection DB**: `catalog-api/internal/media/database/schema.sql` (10 tables, 2 views)
- **v3 Multi-User Schema**: `database/schema_v3_multiuser.sql` (18 tables, 3 views)

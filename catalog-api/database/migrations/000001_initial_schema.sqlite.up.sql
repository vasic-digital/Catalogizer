-- SQLite-specific initial schema migration

-- Storage roots table
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

-- Duplicate groups table
CREATE TABLE IF NOT EXISTS duplicate_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_count INTEGER DEFAULT 0,
    total_size INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

-- Indexes
CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path);
CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id);
CREATE INDEX IF NOT EXISTS idx_files_duplicate_group ON files(duplicate_group_id);
CREATE INDEX IF NOT EXISTS idx_files_deleted ON files(deleted);
CREATE INDEX IF NOT EXISTS idx_file_metadata_file_id ON file_metadata(file_id);
CREATE INDEX IF NOT EXISTS idx_scan_history_storage_root ON scan_history(storage_root_id);
CREATE INDEX IF NOT EXISTS idx_files_name ON files(name);
CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension);
CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type);

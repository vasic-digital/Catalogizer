-- Initial schema migration
-- This migration creates the base tables for Catalogizer

-- Storage roots table (replaces smb_roots)
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

-- Files table
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

-- Duplicate groups table
CREATE TABLE IF NOT EXISTS duplicate_groups (
    id SERIAL PRIMARY KEY,
    file_count INTEGER DEFAULT 0,
    total_size BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key for duplicate_group_id
ALTER TABLE files ADD CONSTRAINT fk_duplicate_group
    FOREIGN KEY (duplicate_group_id) REFERENCES duplicate_groups(id);

-- File metadata table
CREATE TABLE IF NOT EXISTS file_metadata (
    id SERIAL PRIMARY KEY,
    file_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    data_type TEXT DEFAULT 'string',
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Virtual paths table
CREATE TABLE IF NOT EXISTS virtual_paths (
    id SERIAL PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Scan history table
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

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path);
CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id);
CREATE INDEX IF NOT EXISTS idx_files_duplicate_group ON files(duplicate_group_id);
CREATE INDEX IF NOT EXISTS idx_files_deleted ON files(deleted);
CREATE INDEX IF NOT EXISTS idx_file_metadata_file_id ON file_metadata(file_id);
CREATE INDEX IF NOT EXISTS idx_scan_history_storage_root ON scan_history(storage_root_id);
CREATE INDEX IF NOT EXISTS idx_files_name ON files(name);
CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension);
CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type);

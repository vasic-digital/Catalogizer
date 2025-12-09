-- Fix foreign key references in subtitle tables to use files table instead of media_items

-- Drop foreign key constraints (SQLite doesn't support ALTER CONSTRAINT, so we need to recreate tables)
-- First, create backup tables
CREATE TABLE subtitle_tracks_backup AS SELECT * FROM subtitle_tracks;
CREATE TABLE subtitle_sync_status_backup AS SELECT * FROM subtitle_sync_status;
CREATE TABLE subtitle_downloads_backup AS SELECT * FROM subtitle_downloads;
CREATE TABLE media_subtitles_backup AS SELECT * FROM media_subtitles;

-- Drop the tables
DROP TABLE IF EXISTS media_subtitles;
DROP TABLE IF EXISTS subtitle_downloads;
DROP TABLE IF EXISTS subtitle_sync_status;
DROP TABLE IF EXISTS subtitle_tracks;

-- Recreate subtitle_tracks table with correct foreign key
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

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item_id ON subtitle_tracks(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language_code ON subtitle_tracks(language_code);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_source ON subtitle_tracks(source);

-- Recreate subtitle_sync_status table with correct foreign key
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

-- Create indexes for sync status
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_media_item_id ON subtitle_sync_status(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_status ON subtitle_sync_status(status);
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_operation ON subtitle_sync_status(operation);

-- Recreate subtitle_downloads table with correct foreign key
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

-- Create indexes for downloads
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_media_item_id ON subtitle_downloads(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_result_id ON subtitle_downloads(result_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_subtitle_id ON subtitle_downloads(subtitle_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_provider ON subtitle_downloads(provider);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_language ON subtitle_downloads(language);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_download_date ON subtitle_downloads(download_date);

-- Recreate media_subtitles table with correct foreign key
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

-- Create indexes for association
CREATE INDEX IF NOT EXISTS idx_media_subtitles_media_item_id ON media_subtitles(media_item_id);
CREATE INDEX IF NOT EXISTS idx_media_subtitles_subtitle_track_id ON media_subtitles(subtitle_track_id);
CREATE INDEX IF NOT EXISTS idx_media_subtitles_is_active ON media_subtitles(is_active);

-- Restore data from backup tables (if any)
INSERT INTO subtitle_tracks SELECT * FROM subtitle_tracks_backup;
INSERT INTO subtitle_sync_status SELECT * FROM subtitle_sync_status_backup;
INSERT INTO subtitle_downloads SELECT * FROM subtitle_downloads_backup;
INSERT INTO media_subtitles SELECT * FROM media_subtitles_backup;

-- Drop backup tables
DROP TABLE IF EXISTS subtitle_tracks_backup;
DROP TABLE IF EXISTS subtitle_sync_status_backup;
DROP TABLE IF EXISTS subtitle_downloads_backup;
DROP TABLE IF EXISTS media_subtitles_backup;

-- Recreate triggers
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
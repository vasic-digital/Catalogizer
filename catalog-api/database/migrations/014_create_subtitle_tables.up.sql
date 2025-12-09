-- Create subtitle_tracks table
CREATE TABLE IF NOT EXISTS subtitle_tracks (
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
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item_id ON subtitle_tracks(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language_code ON subtitle_tracks(language_code);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_source ON subtitle_tracks(source);

-- Create subtitle_sync_status table for tracking sync operations
CREATE TABLE IF NOT EXISTS subtitle_sync_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    subtitle_id TEXT NOT NULL,
    operation TEXT NOT NULL, -- 'download', 'upload', 'sync', 'verify'
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'failed'
    progress INTEGER DEFAULT 0,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for sync status
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_media_item_id ON subtitle_sync_status(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_status ON subtitle_sync_status(status);
CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_operation ON subtitle_sync_status(operation);

-- Create subtitle_cache table for temporary caching
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
    data TEXT -- JSON blob for additional data
);

-- Create indexes for cache
CREATE INDEX IF NOT EXISTS idx_subtitle_cache_cache_key ON subtitle_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_subtitle_cache_expires_at ON subtitle_cache(expires_at);

-- Create subtitle_downloads table for tracking download history
CREATE TABLE IF NOT EXISTS subtitle_downloads (
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
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for downloads
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_media_item_id ON subtitle_downloads(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_result_id ON subtitle_downloads(result_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_subtitle_id ON subtitle_downloads(subtitle_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_provider ON subtitle_downloads(provider);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_language ON subtitle_downloads(language);
CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_download_date ON subtitle_downloads(download_date);

-- Create media_subtitles association table for many-to-many relationship
CREATE TABLE IF NOT EXISTS media_subtitles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    subtitle_track_id INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
    FOREIGN KEY (subtitle_track_id) REFERENCES subtitle_tracks(id) ON DELETE CASCADE,
    UNIQUE(media_item_id, subtitle_track_id)
);

-- Create indexes for association
CREATE INDEX IF NOT EXISTS idx_media_subtitles_media_item_id ON media_subtitles(media_item_id);
CREATE INDEX IF NOT EXISTS idx_media_subtitles_subtitle_track_id ON media_subtitles(subtitle_track_id);
CREATE INDEX IF NOT EXISTS idx_media_subtitles_is_active ON media_subtitles(is_active);

-- Create trigger to update updated_at timestamp
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
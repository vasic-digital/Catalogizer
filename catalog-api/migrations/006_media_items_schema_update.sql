-- Migration 006: Media Items Schema Update for Android TV Compatibility
-- Updates media_items table to match Android TV app expectations with additional fields

-- Add missing columns to media_items table
ALTER TABLE media_items ADD COLUMN directory_path TEXT;
ALTER TABLE media_items ADD COLUMN smb_path TEXT;
ALTER TABLE media_items ADD COLUMN external_metadata TEXT DEFAULT '[]';
ALTER TABLE media_items ADD COLUMN versions TEXT DEFAULT '[]';
ALTER TABLE media_items ADD COLUMN watch_progress REAL DEFAULT 0.0;
ALTER TABLE media_items ADD COLUMN last_watched TIMESTAMP;
ALTER TABLE media_items ADD COLUMN is_downloaded BOOLEAN DEFAULT FALSE;

-- Update existing records to set directory_path from path if null
UPDATE media_items SET directory_path = substr(path, 1, length(path) - length(filename) - 1) 
WHERE directory_path IS NULL AND path IS NOT NULL AND filename IS NOT NULL;

-- Rename path column to file_path for clarity
ALTER TABLE media_items RENAME TO media_items_old;

-- Create new media_items table with correct schema
CREATE TABLE media_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    media_type TEXT NOT NULL,
    year INTEGER,
    description TEXT,
    cover_image TEXT,
    rating REAL,
    quality TEXT,
    file_size INTEGER,
    duration INTEGER,
    directory_path TEXT NOT NULL,
    smb_path TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    external_metadata TEXT DEFAULT '[]',
    versions TEXT DEFAULT '[]',
    is_favorite BOOLEAN DEFAULT FALSE,
    watch_progress REAL DEFAULT 0.0,
    last_watched TIMESTAMP,
    is_downloaded BOOLEAN DEFAULT FALSE
);

-- Migrate data from old table
INSERT INTO media_items (
    id, title, media_type, year, description, rating, 
    file_size, duration, directory_path, created_at, updated_at,
    is_favorite, watch_progress, last_watched, is_downloaded,
    quality, cover_image
)
SELECT 
    id, 
    COALESCE(title, filename) as title,
    media_type,
    year,
    description,
    CAST(CASE WHEN rating BETWEEN 1 AND 5 THEN rating * 2.0 ELSE NULL END AS REAL), -- Convert 1-5 to 2-10 scale
    size as file_size,
    CAST(duration AS INTEGER),
    directory_path,
    created_at,
    updated_at,
    is_favorite,
    COALESCE(last_position / 300.0, 0.0) as watch_progress, -- Convert position to progress (5min default duration)
    last_played as last_watched,
    FALSE as is_downloaded, -- Default to false for existing records
    resolution as quality,
    -- Try to extract cover art from cover_art table if exists
    (SELECT local_path FROM cover_art ca WHERE ca.media_item_id = media_items_old.id AND ca.is_default = 1 LIMIT 1)
FROM media_items_old;

-- Create indexes for new table
CREATE INDEX idx_media_items_type ON media_items(media_type);
CREATE INDEX idx_media_items_year ON media_items(year);
CREATE INDEX idx_media_items_rating ON media_items(rating);
CREATE INDEX idx_media_items_favorite ON media_items(is_favorite);
CREATE INDEX idx_media_items_watch_progress ON media_items(watch_progress);
CREATE INDEX idx_media_items_last_watched ON media_items(last_watched);
CREATE INDEX idx_media_items_downloaded ON media_items(is_downloaded);
CREATE INDEX idx_media_items_quality ON media_items(quality);
CREATE INDEX idx_media_items_directory_path ON media_items(directory_path);

-- Drop old table
DROP TABLE media_items_old;

-- Create trigger for updating timestamps
CREATE TRIGGER IF NOT EXISTS update_media_items_timestamp
    AFTER UPDATE ON media_items
    BEGIN
        UPDATE media_items SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Insert some sample data for testing if table is empty
INSERT OR IGNORE INTO media_items (title, media_type, directory_path, created_at, updated_at)
VALUES 
    ('Sample Movie', 'movie', '/movies/sample', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Sample TV Show', 'tv_show', '/tv/shows/sample', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Sample Documentary', 'documentary', '/docs/sample', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

PRAGMA foreign_keys = ON;
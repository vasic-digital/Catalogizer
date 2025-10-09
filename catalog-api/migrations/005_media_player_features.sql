-- Migration 005: Media Player Features
-- Adds comprehensive media player functionality including subtitles, lyrics, cover art, and translations

-- Media items table with enhanced metadata
CREATE TABLE IF NOT EXISTS media_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    filename TEXT NOT NULL,
    title TEXT NOT NULL,
    media_type TEXT NOT NULL CHECK (media_type IN ('music', 'video', 'game', 'software', 'ebook', 'document')),
    mime_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    duration REAL, -- Duration in seconds
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Music-specific metadata
    artist TEXT,
    album TEXT,
    album_artist TEXT,
    genre TEXT,
    year INTEGER,
    track_number INTEGER,
    disc_number INTEGER,

    -- Video-specific metadata
    video_codec TEXT,
    audio_codec TEXT,
    resolution TEXT, -- e.g., "1920x1080"
    framerate REAL,
    bitrate INTEGER,

    -- TV Show/Series metadata
    series_title TEXT,
    season INTEGER,
    episode INTEGER,
    episode_title TEXT,

    -- Additional metadata
    description TEXT,
    language TEXT,

    -- Playback metadata
    last_position REAL, -- Last playback position in seconds
    play_count INTEGER DEFAULT 0,
    last_played TIMESTAMP,
    is_favorite BOOLEAN DEFAULT FALSE,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5), -- 1-5 stars

    -- Indexing for performance
    FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
);

-- Create indexes for media items
CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(media_type);
CREATE INDEX IF NOT EXISTS idx_media_items_artist ON media_items(artist);
CREATE INDEX IF NOT EXISTS idx_media_items_album ON media_items(album);
CREATE INDEX IF NOT EXISTS idx_media_items_genre ON media_items(genre);
CREATE INDEX IF NOT EXISTS idx_media_items_year ON media_items(year);
CREATE INDEX IF NOT EXISTS idx_media_items_series ON media_items(series_title, season, episode);
CREATE INDEX IF NOT EXISTS idx_media_items_play_count ON media_items(play_count);
CREATE INDEX IF NOT EXISTS idx_media_items_last_played ON media_items(last_played);
CREATE INDEX IF NOT EXISTS idx_media_items_rating ON media_items(rating);

-- Subtitle tracks table
CREATE TABLE IF NOT EXISTS subtitle_tracks (
    id TEXT PRIMARY KEY,
    media_item_id INTEGER NOT NULL,
    language TEXT NOT NULL,
    language_code TEXT NOT NULL, -- ISO 639-1 code
    source TEXT NOT NULL CHECK (source IN ('embedded', 'external', 'downloaded', 'translated')),
    format TEXT NOT NULL CHECK (format IN ('srt', 'vtt', 'ass', 'ssa', 'sub', 'idx')),
    path TEXT, -- Path to external subtitle file
    content TEXT, -- Inline subtitle content
    is_default BOOLEAN DEFAULT FALSE,
    is_forced BOOLEAN DEFAULT FALSE,
    encoding TEXT DEFAULT 'utf-8',
    sync_offset REAL DEFAULT 0.0, -- Milliseconds offset for sync adjustment
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified_sync BOOLEAN DEFAULT FALSE, -- Whether sync has been verified

    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for subtitle tracks
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item ON subtitle_tracks(media_item_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language_code);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_default ON subtitle_tracks(is_default);

-- Audio tracks table (for multi-language audio)
CREATE TABLE IF NOT EXISTS audio_tracks (
    id TEXT PRIMARY KEY,
    media_item_id INTEGER NOT NULL,
    language TEXT NOT NULL,
    language_code TEXT NOT NULL,
    codec TEXT NOT NULL,
    channels INTEGER NOT NULL,
    bitrate INTEGER,
    sample_rate INTEGER,
    is_default BOOLEAN DEFAULT FALSE,
    title TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for audio tracks
CREATE INDEX IF NOT EXISTS idx_audio_tracks_media_item ON audio_tracks(media_item_id);
CREATE INDEX IF NOT EXISTS idx_audio_tracks_language ON audio_tracks(language_code);

-- Chapters table (for video chapters and bookmarks)
CREATE TABLE IF NOT EXISTS chapters (
    id TEXT PRIMARY KEY,
    media_item_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    start_time REAL NOT NULL, -- Start time in seconds
    end_time REAL, -- End time in seconds (optional)
    thumbnail TEXT, -- Path to thumbnail image
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for chapters
CREATE INDEX IF NOT EXISTS idx_chapters_media_item ON chapters(media_item_id);
CREATE INDEX IF NOT EXISTS idx_chapters_start_time ON chapters(start_time);

-- Cover art table
CREATE TABLE IF NOT EXISTS cover_art (
    id TEXT PRIMARY KEY,
    media_item_id INTEGER NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('embedded', 'local', 'musicbrainz', 'lastfm', 'spotify', 'itunes', 'discogs')),
    url TEXT, -- Remote URL
    local_path TEXT, -- Local file path
    width INTEGER,
    height INTEGER,
    format TEXT NOT NULL CHECK (format IN ('jpeg', 'png', 'webp', 'gif')),
    size INTEGER, -- File size in bytes
    quality TEXT NOT NULL CHECK (quality IN ('thumbnail', 'medium', 'high', 'original')),
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cached_at TIMESTAMP,

    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for cover art
CREATE INDEX IF NOT EXISTS idx_cover_art_media_item ON cover_art(media_item_id);
CREATE INDEX IF NOT EXISTS idx_cover_art_quality ON cover_art(quality);
CREATE INDEX IF NOT EXISTS idx_cover_art_default ON cover_art(is_default);

-- Lyrics data table
CREATE TABLE IF NOT EXISTS lyrics_data (
    id TEXT PRIMARY KEY,
    media_item_id INTEGER NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('embedded', 'genius', 'musixmatch', 'azlyrics', 'lyricfind', 'translated')),
    language TEXT NOT NULL,
    language_code TEXT NOT NULL,
    content TEXT NOT NULL,
    is_synced BOOLEAN DEFAULT FALSE,
    sync_data TEXT, -- JSON array of synchronized lyrics lines
    translations TEXT, -- JSON object of language_code -> translated content
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cached_at TIMESTAMP,

    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for lyrics data
CREATE INDEX IF NOT EXISTS idx_lyrics_data_media_item ON lyrics_data(media_item_id);
CREATE INDEX IF NOT EXISTS idx_lyrics_data_language ON lyrics_data(language_code);
CREATE INDEX IF NOT EXISTS idx_lyrics_data_synced ON lyrics_data(is_synced);

-- Playback sessions table (for tracking current playback state)
CREATE TABLE IF NOT EXISTS playback_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    media_item_id INTEGER NOT NULL,
    playlist_id TEXT,
    current_position REAL NOT NULL DEFAULT 0.0,
    state TEXT NOT NULL CHECK (state IN ('playing', 'paused', 'stopped', 'loading', 'error')),
    volume REAL NOT NULL DEFAULT 1.0 CHECK (volume >= 0.0 AND volume <= 1.0),
    playback_rate REAL NOT NULL DEFAULT 1.0,
    repeat_mode TEXT NOT NULL DEFAULT 'off' CHECK (repeat_mode IN ('off', 'one', 'all', 'random')),
    shuffle_enabled BOOLEAN DEFAULT FALSE,
    current_subtitle_id TEXT,
    current_audio_id TEXT,
    player_settings TEXT, -- JSON object for player-specific settings
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
    FOREIGN KEY (current_subtitle_id) REFERENCES subtitle_tracks(id),
    FOREIGN KEY (current_audio_id) REFERENCES audio_tracks(id)
);

-- Create indexes for playback sessions
CREATE INDEX IF NOT EXISTS idx_playback_sessions_user ON playback_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_playback_sessions_media_item ON playback_sessions(media_item_id);
CREATE INDEX IF NOT EXISTS idx_playback_sessions_updated ON playback_sessions(updated_at);

-- Playlists table
CREATE TABLE IF NOT EXISTS playlists (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    is_smart BOOLEAN DEFAULT FALSE, -- Smart playlists based on criteria
    smart_criteria TEXT, -- JSON object for smart playlist rules
    cover_art_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (cover_art_id) REFERENCES cover_art(id)
);

-- Create indexes for playlists
CREATE INDEX IF NOT EXISTS idx_playlists_user ON playlists(user_id);
CREATE INDEX IF NOT EXISTS idx_playlists_public ON playlists(is_public);
CREATE INDEX IF NOT EXISTS idx_playlists_smart ON playlists(is_smart);

-- Playlist items table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS playlist_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    playlist_id TEXT NOT NULL,
    media_item_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    added_by TEXT, -- User who added the item

    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
    UNIQUE(playlist_id, position)
);

-- Create indexes for playlist items
CREATE INDEX IF NOT EXISTS idx_playlist_items_playlist ON playlist_items(playlist_id);
CREATE INDEX IF NOT EXISTS idx_playlist_items_media_item ON playlist_items(media_item_id);
CREATE INDEX IF NOT EXISTS idx_playlist_items_position ON playlist_items(playlist_id, position);

-- User preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    preference_key TEXT NOT NULL,
    preference_value TEXT NOT NULL, -- JSON value
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, preference_key)
);

-- Create indexes for user preferences
CREATE INDEX IF NOT EXISTS idx_user_preferences_user ON user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_user_preferences_key ON user_preferences(preference_key);

-- Translation cache table
CREATE TABLE IF NOT EXISTS translation_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT NOT NULL UNIQUE,
    source_text TEXT NOT NULL,
    source_language TEXT NOT NULL,
    target_language TEXT NOT NULL,
    translated_text TEXT NOT NULL,
    provider TEXT NOT NULL,
    confidence REAL NOT NULL,
    context_type TEXT, -- 'lyrics', 'subtitle', 'general'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    access_count INTEGER DEFAULT 1
);

-- Create indexes for translation cache
CREATE INDEX IF NOT EXISTS idx_translation_cache_key ON translation_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_translation_cache_languages ON translation_cache(source_language, target_language);
CREATE INDEX IF NOT EXISTS idx_translation_cache_accessed ON translation_cache(accessed_at);

-- External API cache table (for cover art, lyrics, etc.)
CREATE TABLE IF NOT EXISTS external_api_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT NOT NULL UNIQUE,
    provider TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    request_params TEXT, -- JSON parameters
    response_data TEXT NOT NULL, -- JSON response
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    access_count INTEGER DEFAULT 1
);

-- Create indexes for external API cache
CREATE INDEX IF NOT EXISTS idx_external_api_cache_key ON external_api_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_external_api_cache_provider ON external_api_cache(provider);
CREATE INDEX IF NOT EXISTS idx_external_api_cache_expires ON external_api_cache(expires_at);

-- Media analysis queue table (for background processing)
CREATE TABLE IF NOT EXISTS media_analysis_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL,
    analysis_type TEXT NOT NULL CHECK (analysis_type IN ('metadata', 'thumbnail', 'chapters', 'audio_analysis')),
    priority INTEGER DEFAULT 5, -- 1 (highest) to 10 (lowest)
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    progress INTEGER DEFAULT 0, -- 0-100 percentage
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
);

-- Create indexes for media analysis queue
CREATE INDEX IF NOT EXISTS idx_media_analysis_queue_status ON media_analysis_queue(status);
CREATE INDEX IF NOT EXISTS idx_media_analysis_queue_priority ON media_analysis_queue(priority, created_at);
CREATE INDEX IF NOT EXISTS idx_media_analysis_queue_media_item ON media_analysis_queue(media_item_id);

-- Language preferences table
CREATE TABLE IF NOT EXISTS language_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    content_type TEXT NOT NULL CHECK (content_type IN ('subtitle', 'audio', 'lyrics', 'ui')),
    languages TEXT NOT NULL, -- JSON array of preferred language codes in order
    auto_translate BOOLEAN DEFAULT FALSE,
    fallback_language TEXT DEFAULT 'en',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, content_type)
);

-- Create indexes for language preferences
CREATE INDEX IF NOT EXISTS idx_language_preferences_user ON language_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_language_preferences_type ON language_preferences(content_type);

-- Create triggers for updating timestamps
CREATE TRIGGER IF NOT EXISTS update_media_items_timestamp
    AFTER UPDATE ON media_items
    BEGIN
        UPDATE media_items SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_playback_sessions_timestamp
    AFTER UPDATE ON playback_sessions
    BEGIN
        UPDATE playback_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_playlists_timestamp
    AFTER UPDATE ON playlists
    BEGIN
        UPDATE playlists SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_user_preferences_timestamp
    AFTER UPDATE ON user_preferences
    BEGIN
        UPDATE user_preferences SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_language_preferences_timestamp
    AFTER UPDATE ON language_preferences
    BEGIN
        UPDATE language_preferences SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Create trigger to update translation cache access statistics
CREATE TRIGGER IF NOT EXISTS update_translation_cache_access
    AFTER UPDATE ON translation_cache
    BEGIN
        UPDATE translation_cache
        SET accessed_at = CURRENT_TIMESTAMP, access_count = access_count + 1
        WHERE id = NEW.id;
    END;

-- Create trigger to update external API cache access statistics
CREATE TRIGGER IF NOT EXISTS update_external_api_cache_access
    AFTER UPDATE ON external_api_cache
    BEGIN
        UPDATE external_api_cache
        SET accessed_at = CURRENT_TIMESTAMP, access_count = access_count + 1
        WHERE id = NEW.id;
    END;

-- Insert default language preferences for system
INSERT OR IGNORE INTO language_preferences (user_id, content_type, languages, auto_translate, fallback_language)
VALUES
    ('system', 'subtitle', '["en", "es", "fr", "de"]', TRUE, 'en'),
    ('system', 'audio', '["en", "es", "fr", "de"]', FALSE, 'en'),
    ('system', 'lyrics', '["en", "es", "fr", "de"]', TRUE, 'en'),
    ('system', 'ui', '["en"]', FALSE, 'en');

-- Insert popular supported languages
CREATE TABLE IF NOT EXISTS supported_languages (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    native_name TEXT NOT NULL,
    flag TEXT, -- Unicode flag emoji
    direction TEXT NOT NULL DEFAULT 'ltr' CHECK (direction IN ('ltr', 'rtl')),
    is_popular BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO supported_languages (code, name, native_name, flag, direction, is_popular)
VALUES
    ('en', 'English', 'English', '🇺🇸', 'ltr', TRUE),
    ('es', 'Spanish', 'Español', '🇪🇸', 'ltr', TRUE),
    ('fr', 'French', 'Français', '🇫🇷', 'ltr', TRUE),
    ('de', 'German', 'Deutsch', '🇩🇪', 'ltr', TRUE),
    ('it', 'Italian', 'Italiano', '🇮🇹', 'ltr', TRUE),
    ('pt', 'Portuguese', 'Português', '🇵🇹', 'ltr', TRUE),
    ('ru', 'Russian', 'Русский', '🇷🇺', 'ltr', TRUE),
    ('ja', 'Japanese', '日本語', '🇯🇵', 'ltr', TRUE),
    ('ko', 'Korean', '한국어', '🇰🇷', 'ltr', TRUE),
    ('zh', 'Chinese', '中文', '🇨🇳', 'ltr', TRUE),
    ('ar', 'Arabic', 'العربية', '🇸🇦', 'rtl', TRUE),
    ('hi', 'Hindi', 'हिन्दी', '🇮🇳', 'ltr', TRUE),
    ('th', 'Thai', 'ไทย', '🇹🇭', 'ltr', FALSE),
    ('vi', 'Vietnamese', 'Tiếng Việt', '🇻🇳', 'ltr', FALSE),
    ('tr', 'Turkish', 'Türkçe', '🇹🇷', 'ltr', FALSE),
    ('pl', 'Polish', 'Polski', '🇵🇱', 'ltr', FALSE),
    ('nl', 'Dutch', 'Nederlands', '🇳🇱', 'ltr', FALSE),
    ('sv', 'Swedish', 'Svenska', '🇸🇪', 'ltr', FALSE),
    ('da', 'Danish', 'Dansk', '🇩🇰', 'ltr', FALSE),
    ('no', 'Norwegian', 'Norsk', '🇳🇴', 'ltr', FALSE);

-- Create views for common queries
CREATE VIEW IF NOT EXISTS media_with_metadata AS
SELECT
    m.*,
    GROUP_CONCAT(DISTINCT s.language_code) as available_subtitles,
    GROUP_CONCAT(DISTINCT a.language_code) as available_audio,
    ca.local_path as cover_art_path,
    ca.url as cover_art_url,
    l.content as lyrics_content,
    l.is_synced as has_synced_lyrics
FROM media_items m
LEFT JOIN subtitle_tracks s ON m.id = s.media_item_id
LEFT JOIN audio_tracks a ON m.id = a.media_item_id
LEFT JOIN cover_art ca ON m.id = ca.media_item_id AND ca.is_default = 1
LEFT JOIN lyrics_data l ON m.id = l.media_item_id
GROUP BY m.id;

-- Create view for playlist with items
CREATE VIEW IF NOT EXISTS playlists_with_items AS
SELECT
    p.*,
    COUNT(pi.id) as item_count,
    SUM(m.duration) as total_duration
FROM playlists p
LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
LEFT JOIN media_items m ON pi.media_item_id = m.id
GROUP BY p.id;

PRAGMA foreign_keys = ON;
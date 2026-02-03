package tests

import (
	"database/sql"
	"testing"

	// Import SQLite driver once for all tests
	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run schema migrations
	if err := runTestMigrations(db); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	return db
}

// SetupTestDBWithoutMigrations creates an in-memory SQLite database without running migrations
// Use this when you want to test against an empty database
func SetupTestDBWithoutMigrations(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	return db
}

// runTestMigrations creates all necessary tables for testing
func runTestMigrations(db *sql.DB) error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			salt TEXT,
			role_id INTEGER NOT NULL DEFAULT 1,
			first_name TEXT,
			last_name TEXT,
			display_name TEXT,
			avatar_url TEXT,
			time_zone TEXT,
			language TEXT,
			settings TEXT DEFAULT '{}',
			is_active BOOLEAN DEFAULT 1,
			is_locked BOOLEAN DEFAULT 0,
			locked_until DATETIME,
			failed_login_attempts INTEGER DEFAULT 0,
			last_login_at DATETIME,
			last_login_ip TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Storage roots table
		`CREATE TABLE IF NOT EXISTS storage_roots (
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
		)`,

		// Albums table
		`CREATE TABLE IF NOT EXISTS albums (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			artist TEXT NOT NULL,
			album_artist TEXT,
			genre TEXT,
			year INTEGER,
			total_tracks INTEGER DEFAULT 0,
			total_discs INTEGER DEFAULT 1,
			cover_art_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Media items table - comprehensive schema for all services
		// Note: Many string columns have DEFAULT '' because VideoPlayerService doesn't use sql.NullString
		`CREATE TABLE IF NOT EXISTS media_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			filename TEXT DEFAULT '',
			title TEXT NOT NULL,
			original_title TEXT DEFAULT '',
			type TEXT NOT NULL,
			media_type TEXT DEFAULT '',
			mime_type TEXT DEFAULT '',
			size INTEGER DEFAULT 0,
			file_size INTEGER DEFAULT 0,
			file_path TEXT DEFAULT '',
			duration INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_added DATETIME DEFAULT CURRENT_TIMESTAMP,
			artist TEXT DEFAULT '',
			album TEXT DEFAULT '',
			album_id INTEGER,
			album_artist TEXT DEFAULT '',
			genre TEXT DEFAULT '',
			genres TEXT DEFAULT '[]',
			year INTEGER DEFAULT 0,
			track_number INTEGER DEFAULT 0,
			disc_number INTEGER DEFAULT 0,
			video_codec TEXT DEFAULT '',
			audio_codec TEXT DEFAULT '',
			codec TEXT DEFAULT '',
			resolution TEXT DEFAULT '',
			aspect_ratio TEXT DEFAULT '',
			framerate REAL DEFAULT 0.0,
			frame_rate REAL DEFAULT 0.0,
			bitrate INTEGER DEFAULT 0,
			format TEXT DEFAULT '',
			sample_rate INTEGER DEFAULT 0,
			channels INTEGER DEFAULT 0,
			bpm INTEGER,
			key TEXT,
			hdr BOOLEAN DEFAULT FALSE,
			dolby_vision BOOLEAN DEFAULT FALSE,
			dolby_atmos BOOLEAN DEFAULT FALSE,
			series_title TEXT DEFAULT '',
			season INTEGER DEFAULT 0,
			episode INTEGER DEFAULT 0,
			episode_title TEXT DEFAULT '',
			description TEXT DEFAULT '',
			language TEXT DEFAULT '',
			country TEXT DEFAULT '',
			directors TEXT DEFAULT '[]',
			actors TEXT DEFAULT '[]',
			writers TEXT DEFAULT '[]',
			imdb_id TEXT DEFAULT '',
			tmdb_id TEXT DEFAULT '',
			release_date DATETIME,
			last_position REAL DEFAULT 0.0,
			play_count INTEGER DEFAULT 0,
			last_played DATETIME,
			is_favorite BOOLEAN DEFAULT FALSE,
			rating REAL DEFAULT 0.0,
			user_rating INTEGER DEFAULT 0,
			watched_percentage REAL DEFAULT 0.0,
			user_id INTEGER,
			storage_root_id INTEGER,
			FOREIGN KEY (album_id) REFERENCES albums(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
		)`,

		// Subtitle tracks table
		`CREATE TABLE IF NOT EXISTS subtitle_tracks (
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
		)`,

		// Audio tracks table
		`CREATE TABLE IF NOT EXISTS audio_tracks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER NOT NULL,
			language TEXT NOT NULL,
			language_code TEXT NOT NULL,
			codec TEXT NOT NULL,
			channels INTEGER NOT NULL,
			bitrate INTEGER,
			sample_rate INTEGER,
			is_default BOOLEAN DEFAULT FALSE,
			title TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
		)`,

		// Cover art table
		`CREATE TABLE IF NOT EXISTS cover_art (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER NOT NULL,
			source TEXT NOT NULL,
			url TEXT,
			local_path TEXT,
			width INTEGER,
			height INTEGER,
			format TEXT NOT NULL DEFAULT 'jpeg',
			size INTEGER,
			quality TEXT NOT NULL DEFAULT 'medium',
			is_default BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			cached_at DATETIME,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
		)`,

		// Lyrics data table
		`CREATE TABLE IF NOT EXISTS lyrics_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER NOT NULL,
			source TEXT NOT NULL,
			language TEXT NOT NULL,
			language_code TEXT NOT NULL,
			content TEXT NOT NULL,
			is_synced BOOLEAN DEFAULT FALSE,
			sync_data TEXT,
			translations TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			cached_at DATETIME,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
		)`,

		// Playlists table
		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			is_public BOOLEAN DEFAULT FALSE,
			is_smart BOOLEAN DEFAULT FALSE,
			smart_criteria TEXT,
			cover_art_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (cover_art_id) REFERENCES cover_art(id)
		)`,

		// Playlist items table
		`CREATE TABLE IF NOT EXISTS playlist_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER NOT NULL,
			media_item_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			added_by INTEGER,
			FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
			UNIQUE(playlist_id, position)
		)`,

		// Playback sessions table
		`CREATE TABLE IF NOT EXISTS playback_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			media_item_id INTEGER NOT NULL,
			playlist_id INTEGER,
			current_position REAL NOT NULL DEFAULT 0.0,
			state TEXT NOT NULL DEFAULT 'stopped',
			volume REAL NOT NULL DEFAULT 1.0,
			playback_rate REAL NOT NULL DEFAULT 1.0,
			repeat_mode TEXT NOT NULL DEFAULT 'off',
			shuffle_enabled BOOLEAN DEFAULT FALSE,
			current_subtitle_id INTEGER,
			current_audio_id INTEGER,
			player_settings TEXT,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
			FOREIGN KEY (playlist_id) REFERENCES playlists(id),
			FOREIGN KEY (current_subtitle_id) REFERENCES subtitle_tracks(id),
			FOREIGN KEY (current_audio_id) REFERENCES audio_tracks(id)
		)`,

		// Playback positions table (for tracking playback progress)
		`CREATE TABLE IF NOT EXISTS playback_positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			media_item_id INTEGER NOT NULL,
			position REAL NOT NULL DEFAULT 0.0,
			duration REAL,
			progress REAL DEFAULT 0.0,
			device_id TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
			UNIQUE(user_id, media_item_id, device_id)
		)`,

		// Translation cache table
		`CREATE TABLE IF NOT EXISTS translation_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cache_key TEXT NOT NULL UNIQUE,
			source_text TEXT NOT NULL,
			source_language TEXT NOT NULL,
			target_language TEXT NOT NULL,
			translated_text TEXT NOT NULL,
			provider TEXT NOT NULL,
			confidence REAL NOT NULL,
			context_type TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			access_count INTEGER DEFAULT 1
		)`,

		// External API cache table
		`CREATE TABLE IF NOT EXISTS external_api_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cache_key TEXT NOT NULL UNIQUE,
			provider TEXT NOT NULL,
			endpoint TEXT NOT NULL,
			request_params TEXT,
			response_data TEXT NOT NULL,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			access_count INTEGER DEFAULT 1
		)`,

		// Recommendations table
		`CREATE TABLE IF NOT EXISTS recommendations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			media_item_id INTEGER NOT NULL,
			source_item_id INTEGER,
			score REAL NOT NULL DEFAULT 0.0,
			reason TEXT,
			recommendation_type TEXT NOT NULL,
			is_viewed BOOLEAN DEFAULT FALSE,
			is_dismissed BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
			FOREIGN KEY (source_item_id) REFERENCES media_items(id)
		)`,

		// Deep links table
		`CREATE TABLE IF NOT EXISTS deep_links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			short_code TEXT UNIQUE NOT NULL,
			target_type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			user_id INTEGER,
			expires_at DATETIME,
			click_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		// Duplicate groups table
		`CREATE TABLE IF NOT EXISTS duplicate_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_count INTEGER DEFAULT 0,
			total_size INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Files table
		`CREATE TABLE IF NOT EXISTS files (
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
		)`,

		// Create indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(type)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_album_id ON media_items(album_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item ON subtitle_tracks(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audio_tracks_media_item ON audio_tracks(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cover_art_media_item ON cover_art(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_lyrics_data_media_item ON lyrics_data(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_playlist_items_playlist ON playlist_items(playlist_id)`,
		`CREATE INDEX IF NOT EXISTS idx_playback_sessions_user ON playback_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_playback_positions_user_media ON playback_positions(user_id, media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_recommendations_user ON recommendations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_deep_links_short_code ON deep_links(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path)`,

		// Insert default test user
		`INSERT OR IGNORE INTO users (id, username, email, is_active) VALUES (1, 'testuser', 'test@example.com', 1)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}

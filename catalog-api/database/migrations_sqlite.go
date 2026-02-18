package database

import (
	"context"
	"fmt"
)

// createMigrationsTableSQLite creates the migrations tracking table for SQLite.
func (db *DB) createMigrationsTableSQLite(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// createInitialTablesSQLite creates the initial database schema for SQLite.
func (db *DB) createInitialTablesSQLite(ctx context.Context) error {
	queries := []string{
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

		`CREATE TABLE IF NOT EXISTS file_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			data_type TEXT DEFAULT 'string',
			FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS duplicate_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_count INTEGER DEFAULT 0,
			total_size INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS virtual_paths (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			target_type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS scan_history (
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
		)`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path)`,
		`CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_files_duplicate_group ON files(duplicate_group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_files_deleted ON files(deleted)`,
		`CREATE INDEX IF NOT EXISTS idx_file_metadata_file_id ON file_metadata(file_id)`,
		`CREATE INDEX IF NOT EXISTS idx_scan_history_storage_root ON scan_history(storage_root_id)`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// migrateSMBToStorageRootsSQLite migrates SMB root data for SQLite.
func (db *DB) migrateSMBToStorageRootsSQLite(ctx context.Context) error {
	var exists int
	err := db.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='smb_roots'").Scan(&exists)
	if err != nil {
		return err
	}

	if exists == 0 {
		return nil
	}

	query := `
		INSERT INTO storage_roots (
			name, protocol, host, port, path, username, password, domain,
			enabled, max_depth, enable_duplicate_detection, enable_metadata_extraction,
			include_patterns, exclude_patterns, created_at, updated_at, last_scan_at
		)
		SELECT
			name, 'smb', host, port, share, username, password, domain,
			enabled, max_depth, enable_duplicate_detection, enable_metadata_extraction,
			include_patterns, exclude_patterns, created_at, updated_at, last_scan_at
		FROM smb_roots
		WHERE NOT EXISTS (
			SELECT 1 FROM storage_roots WHERE name = smb_roots.name AND protocol = 'smb'
		)
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to migrate SMB roots: %w", err)
	}

	updateQuery := `
		UPDATE files
		SET storage_root_id = (
			SELECT sr.id
			FROM storage_roots sr
			JOIN smb_roots old_sr ON sr.name = old_sr.name AND sr.protocol = 'smb'
			WHERE files.smb_root_id = old_sr.id
		)
		WHERE storage_root_id IS NULL OR storage_root_id = 0
	`
	if _, err := db.ExecContext(ctx, updateQuery); err != nil {
		return fmt.Errorf("failed to update files storage_root_id: %w", err)
	}

	scanUpdateQuery := `
		UPDATE scan_history
		SET storage_root_id = (
			SELECT sr.id
			FROM storage_roots sr
			JOIN smb_roots old_sr ON sr.name = old_sr.name AND sr.protocol = 'smb'
			WHERE scan_history.smb_root_id = old_sr.id
		)
		WHERE storage_root_id IS NULL OR storage_root_id = 0
	`
	if _, err := db.ExecContext(ctx, scanUpdateQuery); err != nil {
		return fmt.Errorf("failed to update scan_history storage_root_id: %w", err)
	}

	return nil
}

// createAuthTablesSQLite creates authentication tables for SQLite.
func (db *DB) createAuthTablesSQLite(ctx context.Context) error {
	schema := `
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

	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		permissions TEXT DEFAULT '[]',
		is_system INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

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

	CREATE TABLE IF NOT EXISTS permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		resource TEXT NOT NULL,
		action TEXT NOT NULL,
		description TEXT
	);

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

	INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
	VALUES (1, 'Admin', 'Administrator role with all permissions', '["*"]', 1);

	INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
	VALUES (2, 'User', 'Standard user role', '["media.view", "media.download"]', 1);

	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
	CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
	CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
	CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create auth tables: %w", err)
	}

	return nil
}

// createConversionJobsTableSQLite creates the conversion_jobs table for SQLite.
func (db *DB) createConversionJobsTableSQLite(ctx context.Context) error {
	query := `
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
		)
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create conversion_jobs table: %w", err)
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_conversion_jobs_user_id ON conversion_jobs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_conversion_jobs_status ON conversion_jobs(status)",
		"CREATE INDEX IF NOT EXISTS idx_conversion_jobs_created_at ON conversion_jobs(created_at)",
	}

	for _, indexQuery := range indexes {
		if _, err := db.ExecContext(ctx, indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", indexQuery, err)
		}
	}

	return nil
}

// createSubtitleTablesSQLite creates subtitle tables for SQLite.
func (db *DB) createSubtitleTablesSQLite(ctx context.Context) error {
	schema := `
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

	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item_id ON subtitle_tracks(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language);
	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language_code ON subtitle_tracks(language_code);
	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_source ON subtitle_tracks(source);

	CREATE TABLE IF NOT EXISTS subtitle_sync_status (
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
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_media_item_id ON subtitle_sync_status(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_status ON subtitle_sync_status(status);
	CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_operation ON subtitle_sync_status(operation);

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

	CREATE INDEX IF NOT EXISTS idx_subtitle_cache_cache_key ON subtitle_cache(cache_key);
	CREATE INDEX IF NOT EXISTS idx_subtitle_cache_expires_at ON subtitle_cache(expires_at);

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

	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_media_item_id ON subtitle_downloads(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_result_id ON subtitle_downloads(result_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_subtitle_id ON subtitle_downloads(subtitle_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_provider ON subtitle_downloads(provider);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_language ON subtitle_downloads(language);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_download_date ON subtitle_downloads(download_date);

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

	CREATE INDEX IF NOT EXISTS idx_media_subtitles_media_item_id ON media_subtitles(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_media_subtitles_subtitle_track_id ON media_subtitles(subtitle_track_id);
	CREATE INDEX IF NOT EXISTS idx_media_subtitles_is_active ON media_subtitles(is_active);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create subtitle tables: %w", err)
	}

	triggers := `
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
	`

	if _, err := db.ExecContext(ctx, triggers); err != nil {
		return fmt.Errorf("failed to create subtitle triggers: %w", err)
	}

	return nil
}

// fixSubtitleForeignKeysSQLite fixes FK references using SQLite backup/recreate.
func (db *DB) fixSubtitleForeignKeysSQLite(ctx context.Context) error {
	migration := `
	CREATE TABLE subtitle_tracks_backup AS SELECT * FROM subtitle_tracks;
	CREATE TABLE subtitle_sync_status_backup AS SELECT * FROM subtitle_sync_status;
	CREATE TABLE subtitle_downloads_backup AS SELECT * FROM subtitle_downloads;
	CREATE TABLE media_subtitles_backup AS SELECT * FROM media_subtitles;

	DROP TABLE IF EXISTS media_subtitles;
	DROP TABLE IF EXISTS subtitle_downloads;
	DROP TABLE IF EXISTS subtitle_sync_status;
	DROP TABLE IF EXISTS subtitle_tracks;

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

	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item_id ON subtitle_tracks(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language);
	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language_code ON subtitle_tracks(language_code);
	CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_source ON subtitle_tracks(source);

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

	CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_media_item_id ON subtitle_sync_status(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_status ON subtitle_sync_status(status);
	CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_operation ON subtitle_sync_status(operation);

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

	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_media_item_id ON subtitle_downloads(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_result_id ON subtitle_downloads(result_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_subtitle_id ON subtitle_downloads(subtitle_id);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_provider ON subtitle_downloads(provider);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_language ON subtitle_downloads(language);
	CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_download_date ON subtitle_downloads(download_date);

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

	CREATE INDEX IF NOT EXISTS idx_media_subtitles_media_item_id ON media_subtitles(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_media_subtitles_subtitle_track_id ON media_subtitles(subtitle_track_id);
	CREATE INDEX IF NOT EXISTS idx_media_subtitles_is_active ON media_subtitles(is_active);

	INSERT INTO subtitle_tracks SELECT * FROM subtitle_tracks_backup;
	INSERT INTO subtitle_sync_status SELECT * FROM subtitle_sync_status_backup;
	INSERT INTO subtitle_downloads SELECT * FROM subtitle_downloads_backup;
	INSERT INTO media_subtitles SELECT * FROM media_subtitles_backup;

	DROP TABLE IF EXISTS subtitle_tracks_backup;
	DROP TABLE IF EXISTS subtitle_sync_status_backup;
	DROP TABLE IF EXISTS subtitle_downloads_backup;
	DROP TABLE IF EXISTS media_subtitles_backup;

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
	`

	if _, err := db.ExecContext(ctx, migration); err != nil {
		return fmt.Errorf("failed to fix subtitle foreign keys: %w", err)
	}

	return nil
}

// createAssetsTableSQLite creates the assets table for SQLite.
func (db *DB) createAssetsTableSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS assets (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		content_type TEXT,
		size INTEGER DEFAULT 0,
		source_hint TEXT,
		entity_type TEXT,
		entity_id TEXT,
		metadata TEXT,
		local_path TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		resolved_at TIMESTAMP,
		expires_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_assets_entity ON assets(entity_type, entity_id);
	CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create assets table: %w", err)
	}

	return nil
}

// createMediaEntityTablesSQLite creates the media entity tables for SQLite (migration v8).
func (db *DB) createMediaEntityTablesSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS media_types (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		detection_patterns TEXT,
		metadata_providers TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('movie', 'Feature films and standalone movies');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('tv_show', 'Television series');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('tv_season', 'Season of a TV show');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('tv_episode', 'Episode of a TV season');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('music_artist', 'Music artist or band');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('music_album', 'Music album');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('song', 'Individual music track');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('game', 'Video games');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('software', 'Software applications and utilities');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('book', 'Books and e-books');
	INSERT OR IGNORE INTO media_types (name, description) VALUES
		('comic', 'Comics and graphic novels');

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
		status TEXT NOT NULL DEFAULT 'detected',
		parent_id INTEGER,
		season_number INTEGER,
		episode_number INTEGER,
		track_number INTEGER,
		first_detected DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_type_id) REFERENCES media_types(id),
		FOREIGN KEY (parent_id) REFERENCES media_items(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS media_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_item_id INTEGER NOT NULL,
		file_id INTEGER NOT NULL,
		quality_info TEXT,
		language TEXT,
		is_primary INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
		FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
	);

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

	CREATE TABLE IF NOT EXISTS media_collection_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		collection_id INTEGER NOT NULL,
		media_item_id INTEGER NOT NULL,
		sequence_number INTEGER,
		season_number INTEGER,
		release_order INTEGER,
		FOREIGN KEY (collection_id) REFERENCES media_collections(id) ON DELETE CASCADE,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS external_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_item_id INTEGER NOT NULL,
		provider TEXT NOT NULL,
		external_id TEXT NOT NULL,
		data TEXT,
		rating REAL,
		review_url TEXT,
		cover_url TEXT,
		trailer_url TEXT,
		last_fetched DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS user_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_item_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		user_rating REAL,
		watched_status TEXT,
		watched_date DATETIME,
		personal_notes TEXT,
		tags TEXT,
		favorite INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS directory_analyses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		directory_path TEXT NOT NULL,
		smb_root TEXT,
		media_item_id INTEGER,
		confidence_score REAL DEFAULT 0,
		detection_method TEXT,
		analysis_data TEXT,
		last_analyzed DATETIME DEFAULT CURRENT_TIMESTAMP,
		files_count INTEGER DEFAULT 0,
		total_size INTEGER DEFAULT 0,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS detection_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type_id INTEGER NOT NULL,
		rule_name TEXT NOT NULL,
		rule_type TEXT NOT NULL,
		pattern TEXT NOT NULL,
		confidence_weight REAL DEFAULT 1.0,
		enabled INTEGER DEFAULT 1,
		priority INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_type_id) REFERENCES media_types(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(media_type_id);
	CREATE INDEX IF NOT EXISTS idx_media_items_parent ON media_items(parent_id);
	CREATE INDEX IF NOT EXISTS idx_media_items_title ON media_items(title);
	CREATE INDEX IF NOT EXISTS idx_media_files_item ON media_files(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_media_files_file ON media_files(file_id);
	CREATE INDEX IF NOT EXISTS idx_external_metadata_item ON external_metadata(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_external_metadata_provider ON external_metadata(provider, external_id);
	CREATE INDEX IF NOT EXISTS idx_user_metadata_item ON user_metadata(media_item_id);
	CREATE INDEX IF NOT EXISTS idx_user_metadata_user ON user_metadata(user_id);
	CREATE INDEX IF NOT EXISTS idx_directory_analyses_path ON directory_analyses(directory_path);
	CREATE INDEX IF NOT EXISTS idx_detection_rules_type ON detection_rules(media_type_id);
	CREATE INDEX IF NOT EXISTS idx_media_collection_items_collection ON media_collection_items(collection_id);
	CREATE INDEX IF NOT EXISTS idx_media_collection_items_item ON media_collection_items(media_item_id);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create media entity tables: %w", err)
	}

	return nil
}

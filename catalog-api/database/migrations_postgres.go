package database

import (
	"context"
	"fmt"
)

// createMigrationsTablePostgres creates the migrations tracking table for PostgreSQL.
func (db *DB) createMigrationsTablePostgres(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// createInitialTablesPostgres creates the initial database schema for PostgreSQL.
// Tables are ordered so that FK references are satisfied: duplicate_groups before files.
func (db *DB) createInitialTablesPostgres(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS storage_roots (
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
		)`,

		// duplicate_groups must come before files (files has FK to duplicate_groups)
		`CREATE TABLE IF NOT EXISTS duplicate_groups (
			id SERIAL PRIMARY KEY,
			file_count INTEGER DEFAULT 0,
			total_size BIGINT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS files (
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
			FOREIGN KEY (parent_id) REFERENCES files(id),
			FOREIGN KEY (duplicate_group_id) REFERENCES duplicate_groups(id)
		)`,

		`CREATE TABLE IF NOT EXISTS file_metadata (
			id SERIAL PRIMARY KEY,
			file_id INTEGER NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			data_type TEXT DEFAULT 'string',
			FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS virtual_paths (
			id SERIAL PRIMARY KEY,
			path TEXT NOT NULL UNIQUE,
			target_type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS scan_history (
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
			return fmt.Errorf("failed to execute query: %s, error: %w", query[:min(80, len(query))], err)
		}
	}

	return nil
}

// migrateSMBToStorageRootsPostgres migrates SMB root data for PostgreSQL.
func (db *DB) migrateSMBToStorageRootsPostgres(ctx context.Context) error {
	// Check if old smb_roots table exists
	exists, err := db.TableExists(ctx, "smb_roots")
	if err != nil {
		return err
	}
	if !exists {
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

	return nil
}

// createAuthTablesPostgres creates authentication tables for PostgreSQL.
func (db *DB) createAuthTablesPostgres(ctx context.Context) error {
	// Execute each statement individually for PostgreSQL
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
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
			is_active BOOLEAN DEFAULT TRUE,
			is_locked BOOLEAN DEFAULT FALSE,
			locked_until TIMESTAMP,
			failed_login_attempts INTEGER DEFAULT 0,
			last_login_at TIMESTAMP,
			last_login_ip TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS roles (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			permissions TEXT DEFAULT '[]',
			is_system BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS user_sessions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			session_token TEXT NOT NULL UNIQUE,
			refresh_token TEXT,
			device_info TEXT,
			ip_address TEXT,
			user_agent TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS permissions (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			resource TEXT NOT NULL,
			action TEXT NOT NULL,
			description TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS user_permissions (
			user_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			granted_by INTEGER,
			PRIMARY KEY (user_id, permission_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
			FOREIGN KEY (granted_by) REFERENCES users(id)
		)`,

		`CREATE TABLE IF NOT EXISTS auth_audit_log (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			event_type TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			details TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`INSERT INTO roles (id, name, description, permissions, is_system)
		 VALUES (1, 'Admin', 'Administrator role with all permissions', '["*"]', TRUE)
		 ON CONFLICT (id) DO NOTHING`,

		`INSERT INTO roles (id, name, description, permissions, is_system)
		 VALUES (2, 'User', 'Standard user role', '["media.view", "media.download"]', TRUE)
		 ON CONFLICT (id) DO NOTHING`,

		// Reset the sequence to be after the seeded IDs
		`SELECT setval('roles_id_seq', GREATEST((SELECT MAX(id) FROM roles), 2))`,

		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token)`,
		`CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at)`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute auth migration: %w", err)
		}
	}

	return nil
}

// createConversionJobsTablePostgres creates the conversion_jobs table for PostgreSQL.
func (db *DB) createConversionJobsTablePostgres(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS conversion_jobs (
			id SERIAL PRIMARY KEY,
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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			scheduled_for TIMESTAMP,
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

// createSubtitleTablesPostgres creates subtitle tables for PostgreSQL.
func (db *DB) createSubtitleTablesPostgres(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS subtitle_tracks (
			id SERIAL PRIMARY KEY,
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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_media_item_id ON subtitle_tracks(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language ON subtitle_tracks(language)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_language_code ON subtitle_tracks(language_code)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_source ON subtitle_tracks(source)`,

		`CREATE TABLE IF NOT EXISTS subtitle_sync_status (
			id SERIAL PRIMARY KEY,
			media_item_id INTEGER NOT NULL,
			subtitle_id TEXT NOT NULL,
			operation TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			progress INTEGER DEFAULT 0,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_media_item_id ON subtitle_sync_status(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_status ON subtitle_sync_status(status)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_sync_status_operation ON subtitle_sync_status(operation)`,

		`CREATE TABLE IF NOT EXISTS subtitle_cache (
			id SERIAL PRIMARY KEY,
			cache_key TEXT UNIQUE NOT NULL,
			result_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			title TEXT,
			language TEXT,
			language_code TEXT,
			download_url TEXT,
			format TEXT,
			encoding TEXT,
			upload_date TIMESTAMP,
			downloads INTEGER,
			rating REAL,
			comments INTEGER,
			match_score REAL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			data TEXT
		)`,

		`CREATE INDEX IF NOT EXISTS idx_subtitle_cache_cache_key ON subtitle_cache(cache_key)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_cache_expires_at ON subtitle_cache(expires_at)`,

		`CREATE TABLE IF NOT EXISTS subtitle_downloads (
			id SERIAL PRIMARY KEY,
			media_item_id INTEGER NOT NULL,
			result_id TEXT NOT NULL,
			subtitle_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			language TEXT NOT NULL,
			file_path TEXT,
			file_size INTEGER,
			download_url TEXT,
			download_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			verified_sync BOOLEAN DEFAULT FALSE,
			sync_offset REAL DEFAULT 0.0
		)`,

		`CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_media_item_id ON subtitle_downloads(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_result_id ON subtitle_downloads(result_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_subtitle_id ON subtitle_downloads(subtitle_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_provider ON subtitle_downloads(provider)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_language ON subtitle_downloads(language)`,
		`CREATE INDEX IF NOT EXISTS idx_subtitle_downloads_download_date ON subtitle_downloads(download_date)`,

		`CREATE TABLE IF NOT EXISTS media_subtitles (
			id SERIAL PRIMARY KEY,
			media_item_id INTEGER NOT NULL,
			subtitle_track_id INTEGER NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(media_item_id, subtitle_track_id)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_media_subtitles_media_item_id ON media_subtitles(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_subtitles_subtitle_track_id ON media_subtitles(subtitle_track_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_subtitles_is_active ON media_subtitles(is_active)`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to create subtitle tables: %w", err)
		}
	}

	// PostgreSQL triggers use functions
	triggerStatements := []string{
		`CREATE OR REPLACE FUNCTION update_subtitle_tracks_timestamp()
		 RETURNS TRIGGER AS $$
		 BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		 END;
		 $$ LANGUAGE plpgsql`,

		`DROP TRIGGER IF EXISTS update_subtitle_tracks_updated_at ON subtitle_tracks`,
		`CREATE TRIGGER update_subtitle_tracks_updated_at
			BEFORE UPDATE ON subtitle_tracks
			FOR EACH ROW
			EXECUTE FUNCTION update_subtitle_tracks_timestamp()`,

		`CREATE OR REPLACE FUNCTION update_subtitle_sync_status_timestamp()
		 RETURNS TRIGGER AS $$
		 BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
				NEW.completed_at = CURRENT_TIMESTAMP;
			END IF;
			RETURN NEW;
		 END;
		 $$ LANGUAGE plpgsql`,

		`DROP TRIGGER IF EXISTS update_subtitle_sync_status_updated_at ON subtitle_sync_status`,
		`CREATE TRIGGER update_subtitle_sync_status_updated_at
			BEFORE UPDATE ON subtitle_sync_status
			FOR EACH ROW
			EXECUTE FUNCTION update_subtitle_sync_status_timestamp()`,
	}

	for _, stmt := range triggerStatements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to create subtitle triggers: %w", err)
		}
	}

	return nil
}

// fixSubtitleForeignKeysPostgres fixes FK references using ALTER TABLE.
func (db *DB) fixSubtitleForeignKeysPostgres(ctx context.Context) error {
	// PostgreSQL can drop and add constraints directly.
	// Since the FK constraints might not have explicit names, we skip this
	// for fresh installations (the v5 migration already creates them without
	// the FK to media_items for PG, since media_items doesn't exist).
	// This migration is only needed for SQLite backup/recreate pattern.
	// For PostgreSQL, we treat this as a no-op since the tables were
	// created correctly in the first place (or are fresh).
	return nil
}

// createAssetsTablePostgres creates the assets table for PostgreSQL.
func (db *DB) createAssetsTablePostgres(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS assets (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			content_type TEXT,
			size BIGINT DEFAULT 0,
			source_hint TEXT,
			entity_type TEXT,
			entity_id TEXT,
			metadata TEXT,
			local_path TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			resolved_at TIMESTAMP,
			expires_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_entity ON assets(entity_type, entity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status)`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to create assets table: %w", err)
		}
	}

	return nil
}

// createMediaEntityTablesPostgres creates the media entity tables for PostgreSQL (migration v8).
func (db *DB) createMediaEntityTablesPostgres(ctx context.Context) error {
	statements := []string{
		// media_types seed table
		`CREATE TABLE IF NOT EXISTS media_types (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			detection_patterns TEXT,
			metadata_providers TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Seed media types
		`INSERT INTO media_types (name, description) VALUES
			('movie', 'Feature films and standalone movies'),
			('tv_show', 'Television series'),
			('tv_season', 'Season of a TV show'),
			('tv_episode', 'Episode of a TV season'),
			('music_artist', 'Music artist or band'),
			('music_album', 'Music album'),
			('song', 'Individual music track'),
			('game', 'Video games'),
			('software', 'Software applications and utilities'),
			('book', 'Books and e-books'),
			('comic', 'Comics and graphic novels')
		ON CONFLICT (name) DO NOTHING`,

		// media_items core entity table
		`CREATE TABLE IF NOT EXISTS media_items (
			id SERIAL PRIMARY KEY,
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
			first_detected TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_type_id) REFERENCES media_types(id),
			FOREIGN KEY (parent_id) REFERENCES media_items(id) ON DELETE CASCADE
		)`,

		// media_files junction table linking files to media_items
		`CREATE TABLE IF NOT EXISTS media_files (
			id SERIAL PRIMARY KEY,
			media_item_id INTEGER NOT NULL,
			file_id INTEGER NOT NULL,
			quality_info TEXT,
			language TEXT,
			is_primary BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
			FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
		)`,

		// media_collections table
		`CREATE TABLE IF NOT EXISTS media_collections (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			collection_type TEXT NOT NULL,
			description TEXT,
			total_items INTEGER DEFAULT 0,
			external_ids TEXT,
			cover_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// media_collection_items table
		`CREATE TABLE IF NOT EXISTS media_collection_items (
			id SERIAL PRIMARY KEY,
			collection_id INTEGER NOT NULL,
			media_item_id INTEGER NOT NULL,
			sequence_number INTEGER,
			season_number INTEGER,
			release_order INTEGER,
			FOREIGN KEY (collection_id) REFERENCES media_collections(id) ON DELETE CASCADE,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
		)`,

		// external_metadata table
		`CREATE TABLE IF NOT EXISTS external_metadata (
			id SERIAL PRIMARY KEY,
			media_item_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			external_id TEXT NOT NULL,
			data TEXT,
			rating REAL,
			review_url TEXT,
			cover_url TEXT,
			trailer_url TEXT,
			last_fetched TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
		)`,

		// user_metadata table
		`CREATE TABLE IF NOT EXISTS user_metadata (
			id SERIAL PRIMARY KEY,
			media_item_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			user_rating REAL,
			watched_status TEXT,
			watched_date TIMESTAMP,
			personal_notes TEXT,
			tags TEXT,
			favorite BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// directory_analyses table
		`CREATE TABLE IF NOT EXISTS directory_analyses (
			id SERIAL PRIMARY KEY,
			directory_path TEXT NOT NULL,
			smb_root TEXT,
			media_item_id INTEGER,
			confidence_score REAL DEFAULT 0,
			detection_method TEXT,
			analysis_data TEXT,
			last_analyzed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			files_count INTEGER DEFAULT 0,
			total_size BIGINT DEFAULT 0,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE SET NULL
		)`,

		// detection_rules table
		`CREATE TABLE IF NOT EXISTS detection_rules (
			id SERIAL PRIMARY KEY,
			media_type_id INTEGER NOT NULL,
			rule_name TEXT NOT NULL,
			rule_type TEXT NOT NULL,
			pattern TEXT NOT NULL,
			confidence_weight REAL DEFAULT 1.0,
			enabled BOOLEAN DEFAULT TRUE,
			priority INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_type_id) REFERENCES media_types(id) ON DELETE CASCADE
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(media_type_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_parent ON media_items(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_title ON media_items(title)`,
		`CREATE INDEX IF NOT EXISTS idx_media_files_item ON media_files(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_files_file ON media_files(file_id)`,
		`CREATE INDEX IF NOT EXISTS idx_external_metadata_item ON external_metadata(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_external_metadata_provider ON external_metadata(provider, external_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_metadata_item ON user_metadata(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_metadata_user ON user_metadata(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_directory_analyses_path ON directory_analyses(directory_path)`,
		`CREATE INDEX IF NOT EXISTS idx_detection_rules_type ON detection_rules(media_type_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_collection_items_collection ON media_collection_items(collection_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_collection_items_item ON media_collection_items(media_item_id)`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to create media entity tables: %w", err)
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

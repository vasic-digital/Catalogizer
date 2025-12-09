package database

import (
	"context"
	"fmt"
)

// RunMigrations runs database migrations
func (db *DB) RunMigrations(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := db.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run migrations
	migrations := []Migration{
		{
			Version: 1,
			Name:    "create_initial_tables",
			Up:      db.createInitialTables,
		},
		{
			Version: 2,
			Name:    "migrate_smb_to_storage_roots",
			Up:      db.migrateSMBToStorageRoots,
		},
		{
			Version: 3,
			Name:    "create_auth_tables",
			Up:      db.createAuthTables,
		},
		{
			Version: 4,
			Name:    "create_conversion_jobs_table",
			Up:      db.createConversionJobsTable,
		},
		{
			Version: 5,
			Name:    "create_subtitle_tables",
			Up:      db.createSubtitleTables,
		},
		{
			Version: 6,
			Name:    "fix_subtitle_foreign_keys",
			Up:      db.fixSubtitleForeignKeys,
		},
	}

	for _, migration := range migrations {
		if err := db.runMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Name, err)
		}
	}

	return nil
}

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      func(context.Context) error
}

// createMigrationsTable creates the migrations tracking table
func (db *DB) createMigrationsTable(ctx context.Context) error {
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

// runMigration runs a single migration if it hasn't been applied
func (db *DB) runMigration(ctx context.Context, migration Migration) error {
	// Check if migration has already been applied
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE version = ?", migration.Version).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Migration already applied
	}

	// Run migration
	if err := migration.Up(ctx); err != nil {
		return err
	}

	// Record migration as applied
	_, err = db.ExecContext(ctx, "INSERT INTO migrations (version, name) VALUES (?, ?)", migration.Version, migration.Name)
	return err
}

// createInitialTables creates the initial database schema
func (db *DB) createInitialTables(ctx context.Context) error {
	queries := []string{
		// Storage roots table (replaces smb_roots)
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

		// File metadata table
		`CREATE TABLE IF NOT EXISTS file_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			data_type TEXT DEFAULT 'string',
			FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
		)`,

		// Duplicate groups table
		`CREATE TABLE IF NOT EXISTS duplicate_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_count INTEGER DEFAULT 0,
			total_size INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Virtual paths table
		`CREATE TABLE IF NOT EXISTS virtual_paths (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			target_type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Scan history table
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

		// Indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path)`,
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

// migrateSMBToStorageRoots migrates existing SMB root data to the new storage roots format
func (db *DB) migrateSMBToStorageRoots(ctx context.Context) error {
	// Check if old smb_roots table exists
	var exists int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='smb_roots'").Scan(&exists)
	if err != nil {
		return err
	}

	if exists == 0 {
		return nil // No old table to migrate
	}

	// Migrate SMB roots to storage roots
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

	// Update files table to use storage_root_id instead of smb_root_id
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

	// Update scan_history table
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

// createConversionJobsTable creates the conversion_jobs table
func (db *DB) createConversionJobsTable(ctx context.Context) error {
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

	// Create indexes
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

// createAuthTables creates authentication-related tables
func (db *DB) createAuthTables(ctx context.Context) error {
	schema := `
	-- Users table with all required columns
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

	-- User permissions (for custom permissions beyond role)
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

	-- Audit log for authentication events
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

	-- Insert default admin role
	INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
	VALUES (1, 'Admin', 'Administrator role with all permissions', '["*"]', 1);

	-- Insert default user role
	INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
	VALUES (2, 'User', 'Standard user role', '["media.view", "media.download"]', 1);

	-- Indexes for users table
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
	CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

	-- Indexes for sessions table
	CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
	CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create auth tables: %w", err)
	}

	return nil
}

// createSubtitleTables creates subtitle-related tables
func (db *DB) createSubtitleTables(ctx context.Context) error {
	schema := `
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
		operation TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
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
		data TEXT
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
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create subtitle tables: %w", err)
	}

	// Create triggers
	triggers := `
	-- Trigger to update updated_at timestamp
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

// fixSubtitleForeignKeys fixes foreign key references to use files table instead of media_items
func (db *DB) fixSubtitleForeignKeys(ctx context.Context) error {
	migration := `
	-- Fix foreign key references in subtitle tables to use files table instead of media_items

	-- Drop foreign key constraints (SQLite doesn't support ALTER CONSTRAINT, so we need to recreate tables)
	-- First, create backup tables
	CREATE TABLE subtitle_tracks_backup AS SELECT * FROM subtitle_tracks;
	CREATE TABLE subtitle_sync_status_backup AS SELECT * FROM subtitle_sync_status;
	CREATE TABLE subtitle_downloads_backup AS SELECT * FROM subtitle_downloads;
	CREATE TABLE media_subtitles_backup AS SELECT * FROM media_subtitles;

	-- Drop tables
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
	`

	if _, err := db.ExecContext(ctx, migration); err != nil {
		return fmt.Errorf("failed to fix subtitle foreign keys: %w", err)
	}

	return nil
}

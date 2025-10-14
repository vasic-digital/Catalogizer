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

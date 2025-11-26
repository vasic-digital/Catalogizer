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
			source_file_path TEXT NOT NULL,
			target_file_path TEXT NOT NULL,
			source_format TEXT NOT NULL,
			target_format TEXT NOT NULL,
			quality_level TEXT DEFAULT 'medium',
			status TEXT DEFAULT 'pending',
			progress INTEGER DEFAULT 0,
			error_message TEXT,
			started_at DATETIME,
			completed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		is_active BOOLEAN NOT NULL DEFAULT 1,
		last_login DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Roles table
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		permissions TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Sessions table
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ip_address TEXT,
		user_agent TEXT,
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

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	CREATE INDEX IF NOT EXISTS idx_auth_audit_user_id ON auth_audit_log(user_id);
	CREATE INDEX IF NOT EXISTS idx_auth_audit_event_type ON auth_audit_log(event_type);
	CREATE INDEX IF NOT EXISTS idx_auth_audit_created_at ON auth_audit_log(created_at);

	-- Insert default roles
	INSERT OR IGNORE INTO roles (name, description, permissions) VALUES
	('admin', 'System Administrator', '["admin:system", "manage:users", "manage:roles", "read:media", "write:media", "delete:media", "read:catalog", "write:catalog", "delete:catalog", "trigger:analysis", "view:analysis", "view:logs", "access:api", "write:api"]'),
	('moderator', 'Content Moderator', '["read:media", "write:media", "read:catalog", "write:catalog", "trigger:analysis", "view:analysis", "access:api", "write:api"]'),
	('user', 'Regular User', '["read:media", "write:media", "read:catalog", "write:catalog", "view:analysis", "access:api"]'),
	('viewer', 'Read-only Viewer', '["read:media", "read:catalog", "view:analysis", "access:api"]');

	-- Insert default permissions
	INSERT OR IGNORE INTO permissions (name, resource, action, description) VALUES
	('read:media', 'media', 'read', 'View media items and metadata'),
	('write:media', 'media', 'write', 'Create and update media items'),
	('delete:media', 'media', 'delete', 'Delete media items'),
	('read:catalog', 'catalog', 'read', 'Browse file catalog'),
	('write:catalog', 'catalog', 'write', 'Modify file catalog'),
	('delete:catalog', 'catalog', 'delete', 'Delete from catalog'),
	('trigger:analysis', 'analysis', 'trigger', 'Start media analysis'),
	('view:analysis', 'analysis', 'view', 'View analysis results'),
	('manage:users', 'users', 'manage', 'Create, update, delete users'),
	('manage:roles', 'roles', 'manage', 'Create, update, delete roles'),
	('view:logs', 'logs', 'view', 'View system logs'),
	('admin:system', 'system', 'admin', 'Full system administration'),
	('access:api', 'api', 'access', 'Access API endpoints'),
	('write:api', 'api', 'write', 'Modify data via API');
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create auth tables: %w", err)
	}
	
	return nil
}

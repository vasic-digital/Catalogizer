package database

import (
	"context"
	"fmt"
)

// createSyncTables creates the sync_endpoints, sync_sessions, and sync_schedules
// tables. These tables support the sync service for managing remote synchronization
// endpoints, tracking sync session progress, and scheduling recurring syncs.
//
// Tables:
//   - sync_endpoints: remote sync endpoint definitions (FTP, WebDAV, etc.)
//   - sync_sessions: individual sync execution records with progress tracking
//   - sync_schedules: recurring sync schedule configuration
func (db *DB) createSyncTables(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createSyncTablesPostgres(ctx)
	}
	return db.createSyncTablesSQLite(ctx)
}

func (db *DB) createSyncTablesSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS sync_endpoints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		url TEXT NOT NULL,
		username TEXT,
		password TEXT,
		sync_direction TEXT DEFAULT 'bidirectional',
		local_path TEXT,
		remote_path TEXT,
		sync_settings TEXT,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_sync_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS sync_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		endpoint_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		status TEXT DEFAULT 'running',
		sync_type TEXT,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		duration INTEGER,
		total_files INTEGER DEFAULT 0,
		synced_files INTEGER DEFAULT 0,
		failed_files INTEGER DEFAULT 0,
		skipped_files INTEGER DEFAULT 0,
		error_message TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS sync_schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		endpoint_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		frequency TEXT NOT NULL,
		last_run DATETIME,
		next_run DATETIME,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_sync_endpoints_user_id ON sync_endpoints(user_id);
	CREATE INDEX IF NOT EXISTS idx_sync_endpoints_status ON sync_endpoints(status);
	CREATE INDEX IF NOT EXISTS idx_sync_endpoints_type ON sync_endpoints(type);
	CREATE INDEX IF NOT EXISTS idx_sync_sessions_endpoint_id ON sync_sessions(endpoint_id);
	CREATE INDEX IF NOT EXISTS idx_sync_sessions_user_id ON sync_sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sync_sessions_status ON sync_sessions(status);
	CREATE INDEX IF NOT EXISTS idx_sync_sessions_started_at ON sync_sessions(started_at);
	CREATE INDEX IF NOT EXISTS idx_sync_schedules_endpoint_id ON sync_schedules(endpoint_id);
	CREATE INDEX IF NOT EXISTS idx_sync_schedules_user_id ON sync_schedules(user_id);
	CREATE INDEX IF NOT EXISTS idx_sync_schedules_is_active ON sync_schedules(is_active);
	CREATE INDEX IF NOT EXISTS idx_sync_schedules_next_run ON sync_schedules(next_run);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create sync tables: %w", err)
	}

	return nil
}

func (db *DB) createSyncTablesPostgres(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sync_endpoints (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			url TEXT NOT NULL,
			username TEXT,
			password TEXT,
			sync_direction TEXT DEFAULT 'bidirectional',
			local_path TEXT,
			remote_path TEXT,
			sync_settings TEXT,
			status TEXT DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_sync_at TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`CREATE TABLE IF NOT EXISTS sync_sessions (
			id SERIAL PRIMARY KEY,
			endpoint_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status TEXT DEFAULT 'running',
			sync_type TEXT,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			duration INTEGER,
			total_files INTEGER DEFAULT 0,
			synced_files INTEGER DEFAULT 0,
			failed_files INTEGER DEFAULT 0,
			skipped_files INTEGER DEFAULT 0,
			error_message TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`CREATE TABLE IF NOT EXISTS sync_schedules (
			id SERIAL PRIMARY KEY,
			endpoint_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			frequency TEXT NOT NULL,
			last_run TIMESTAMP,
			next_run TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sync_endpoints_user_id ON sync_endpoints(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_endpoints_status ON sync_endpoints(status)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_endpoints_type ON sync_endpoints(type)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_sessions_endpoint_id ON sync_sessions(endpoint_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_sessions_user_id ON sync_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_sessions_status ON sync_sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_sessions_started_at ON sync_sessions(started_at)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_schedules_endpoint_id ON sync_schedules(endpoint_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_schedules_user_id ON sync_schedules(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_schedules_is_active ON sync_schedules(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_schedules_next_run ON sync_schedules(next_run)`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to create sync tables: %w", err)
		}
	}

	return nil
}

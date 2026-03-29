package database

import (
	"context"
	"fmt"
)

// createPlaylistTables creates the playlists and playlist_items tables
// for the user-facing playlist API (/api/v1/playlists).
func (db *DB) createPlaylistTables(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createPlaylistTablesPostgres(ctx)
	}
	return db.createPlaylistTablesSQLite(ctx)
}

func (db *DB) createPlaylistTablesSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		is_public INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_playlists_user_id ON playlists(user_id);

	CREATE TABLE IF NOT EXISTS playlist_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		playlist_id INTEGER NOT NULL,
		entity_id INTEGER NOT NULL,
		entity_type TEXT NOT NULL,
		position INTEGER DEFAULT 0,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_playlist_items_playlist_id ON playlist_items(playlist_id);
	CREATE INDEX IF NOT EXISTS idx_playlist_items_entity ON playlist_items(entity_id, entity_type);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create playlist tables (SQLite): %w", err)
	}

	return nil
}

func (db *DB) createPlaylistTablesPostgres(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS playlists (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		is_public BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_playlists_user_id ON playlists(user_id);

	CREATE TABLE IF NOT EXISTS playlist_items (
		id SERIAL PRIMARY KEY,
		playlist_id INTEGER NOT NULL,
		entity_id INTEGER NOT NULL,
		entity_type TEXT NOT NULL,
		position INTEGER DEFAULT 0,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_playlist_items_playlist_id ON playlist_items(playlist_id);
	CREATE INDEX IF NOT EXISTS idx_playlist_items_entity ON playlist_items(entity_id, entity_type);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create playlist tables (PostgreSQL): %w", err)
	}

	return nil
}

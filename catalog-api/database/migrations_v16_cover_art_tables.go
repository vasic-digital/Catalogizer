package database

import (
	"context"
	"fmt"
)

// createCoverArtTables creates the cover_art and cover_art_cache tables
// used by the cover art service to persist downloaded and generated
// cover images locally.
func (db *DB) createCoverArtTables(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createCoverArtTablesPostgres(ctx)
	}
	return db.createCoverArtTablesSQLite(ctx)
}

func (db *DB) createCoverArtTablesSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS cover_art (
		id TEXT PRIMARY KEY,
		media_item_id INTEGER NOT NULL,
		source TEXT NOT NULL,
		url TEXT,
		local_path TEXT,
		width INTEGER,
		height INTEGER,
		format TEXT,
		size INTEGER,
		quality TEXT,
		is_default INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		cached_at DATETIME,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_cover_art_media_item
		ON cover_art(media_item_id, is_default, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_cover_art_source
		ON cover_art(source);

	CREATE TABLE IF NOT EXISTS cover_art_cache (
		id TEXT PRIMARY KEY,
		cache_key TEXT NOT NULL UNIQUE,
		provider TEXT NOT NULL,
		title TEXT NOT NULL,
		artist TEXT,
		album TEXT,
		url TEXT NOT NULL,
		thumbnail_url TEXT,
		width INTEGER,
		height INTEGER,
		format TEXT,
		quality TEXT,
		size INTEGER,
		match_score REAL,
		source TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_cover_art_cache_key
		ON cover_art_cache(cache_key);
	CREATE INDEX IF NOT EXISTS idx_cover_art_cache_created
		ON cover_art_cache(created_at);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create cover art tables (SQLite): %w", err)
	}
	return nil
}

func (db *DB) createCoverArtTablesPostgres(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS cover_art (
		id TEXT PRIMARY KEY,
		media_item_id BIGINT NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
		source TEXT NOT NULL,
		url TEXT,
		local_path TEXT,
		width INTEGER,
		height INTEGER,
		format TEXT,
		size BIGINT,
		quality TEXT,
		is_default BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		cached_at TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_cover_art_media_item
		ON cover_art(media_item_id, is_default, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_cover_art_source
		ON cover_art(source);

	CREATE TABLE IF NOT EXISTS cover_art_cache (
		id TEXT PRIMARY KEY,
		cache_key TEXT NOT NULL UNIQUE,
		provider TEXT NOT NULL,
		title TEXT NOT NULL,
		artist TEXT,
		album TEXT,
		url TEXT NOT NULL,
		thumbnail_url TEXT,
		width INTEGER,
		height INTEGER,
		format TEXT,
		quality TEXT,
		size BIGINT,
		match_score DOUBLE PRECISION,
		source TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_cover_art_cache_key
		ON cover_art_cache(cache_key);
	CREATE INDEX IF NOT EXISTS idx_cover_art_cache_created
		ON cover_art_cache(created_at);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create cover art tables (PostgreSQL): %w", err)
	}
	return nil
}

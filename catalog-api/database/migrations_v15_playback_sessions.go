package database

import (
	"context"
	"fmt"
)

// createPlaybackSessionTables creates playback_sessions and
// media_progress tables used by the playback history tracking
// feature. Every time a user reproduces a media item (plays a
// video, audio, reads a book, etc.) a row is appended to
// playback_sessions; media_progress is a denormalised snapshot
// of the latest state per (user, media_item) that the cards UI
// reads to show duration / current / last session / total
// reproduction count at a glance without walking the session
// history.
//
// position_unit is an enum ('seconds' | 'pages' | 'events') so
// the same table serves video/audio (seconds), books/comics
// (pages), and games/software (events = launches). Units are
// stored as BIGINT; callers convert to the appropriate display
// format on the client.
func (db *DB) createPlaybackSessionTables(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createPlaybackSessionTablesPostgres(ctx)
	}
	return db.createPlaybackSessionTablesSQLite(ctx)
}

func (db *DB) createPlaybackSessionTablesSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS playback_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		media_item_id INTEGER NOT NULL,
		file_id INTEGER,
		started_at DATETIME NOT NULL,
		ended_at DATETIME,
		position_unit TEXT NOT NULL CHECK (position_unit IN ('seconds','pages','events')),
		start_position INTEGER NOT NULL DEFAULT 0,
		end_position INTEGER,
		total_amount INTEGER NOT NULL DEFAULT 0,
		completed INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_playback_sessions_item
		ON playback_sessions(media_item_id, started_at DESC);
	CREATE INDEX IF NOT EXISTS idx_playback_sessions_user
		ON playback_sessions(user_id, started_at DESC);

	CREATE TABLE IF NOT EXISTS media_progress (
		user_id INTEGER NOT NULL,
		media_item_id INTEGER NOT NULL,
		position_unit TEXT NOT NULL,
		duration_total INTEGER,
		last_position INTEGER NOT NULL DEFAULT 0,
		last_session_amount INTEGER NOT NULL DEFAULT 0,
		total_reproductions INTEGER NOT NULL DEFAULT 0,
		aggregate_amount INTEGER NOT NULL DEFAULT 0,
		last_session_ended_at DATETIME,
		updated_at DATETIME NOT NULL,
		PRIMARY KEY (user_id, media_item_id),
		FOREIGN KEY (media_item_id) REFERENCES media_items(id) ON DELETE CASCADE
	);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create playback session tables (SQLite): %w", err)
	}
	return nil
}

func (db *DB) createPlaybackSessionTablesPostgres(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS playback_sessions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		media_item_id INTEGER NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
		file_id INTEGER,
		started_at TIMESTAMP NOT NULL,
		ended_at TIMESTAMP,
		position_unit TEXT NOT NULL CHECK (position_unit IN ('seconds','pages','events')),
		start_position BIGINT NOT NULL DEFAULT 0,
		end_position BIGINT,
		total_amount BIGINT NOT NULL DEFAULT 0,
		completed BOOLEAN NOT NULL DEFAULT FALSE
	);
	CREATE INDEX IF NOT EXISTS idx_playback_sessions_item
		ON playback_sessions(media_item_id, started_at DESC);
	CREATE INDEX IF NOT EXISTS idx_playback_sessions_user
		ON playback_sessions(user_id, started_at DESC);

	CREATE TABLE IF NOT EXISTS media_progress (
		user_id INTEGER NOT NULL,
		media_item_id INTEGER NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
		position_unit TEXT NOT NULL,
		duration_total BIGINT,
		last_position BIGINT NOT NULL DEFAULT 0,
		last_session_amount BIGINT NOT NULL DEFAULT 0,
		total_reproductions BIGINT NOT NULL DEFAULT 0,
		aggregate_amount BIGINT NOT NULL DEFAULT 0,
		last_session_ended_at TIMESTAMP,
		updated_at TIMESTAMP NOT NULL,
		PRIMARY KEY (user_id, media_item_id)
	);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create playback session tables (PostgreSQL): %w", err)
	}
	return nil
}

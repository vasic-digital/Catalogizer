package database

import (
	"context"
	"fmt"
)

// addMediaItemsFavoriteColumn adds the is_favorite + updated_at columns
// to media_items.
//
// FIX-QA-2026-04-21-001: the /api/v1/media/:id/favorite handler
// (AndroidTVMediaHandler.UpdateFavoriteStatus) writes
// `is_favorite = ?, updated_at = CURRENT_TIMESTAMP`, but the production
// schema in migrations_sqlite.go never declared either column — every
// call returned 500 ("Failed to update favorite status"), observed 4×
// in the 2026-04-20 Article VII RunAll log analysis.
//
// The test_helper schema (internal/tests/test_helper.go) does declare
// both, which is why the handler's unit tests pass. This migration
// brings production schema in line with the test-schema assumption the
// handler was written against.
//
// The per-user favorite flag in user_metadata.favorite is preserved
// (it serves a different use case); this is the global/Android-TV
// style favorite flag the handler targets.
func (db *DB) addMediaItemsFavoriteColumn(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.addMediaItemsFavoriteColumnPostgres(ctx)
	}
	return db.addMediaItemsFavoriteColumnSQLite(ctx)
}

func (db *DB) addMediaItemsFavoriteColumnSQLite(ctx context.Context) error {
	has, err := db.columnExistsSQLite(ctx, "media_items", "is_favorite")
	if err != nil {
		return fmt.Errorf("sqlite: probe media_items.is_favorite: %w", err)
	}
	if !has {
		if _, err := db.ExecContext(ctx, `ALTER TABLE media_items ADD COLUMN is_favorite INTEGER NOT NULL DEFAULT 0`); err != nil {
			return fmt.Errorf("sqlite: add media_items.is_favorite: %w", err)
		}
	}

	has, err = db.columnExistsSQLite(ctx, "media_items", "updated_at")
	if err != nil {
		return fmt.Errorf("sqlite: probe media_items.updated_at: %w", err)
	}
	if !has {
		// SQLite refuses non-constant defaults on ALTER TABLE, so we
		// add the column nullable and backfill existing rows with the
		// last_updated timestamp (semantically equivalent for existing
		// data — the handler overwrites it on every update anyway).
		if _, err := db.ExecContext(ctx, `ALTER TABLE media_items ADD COLUMN updated_at DATETIME`); err != nil {
			return fmt.Errorf("sqlite: add media_items.updated_at: %w", err)
		}
		if _, err := db.ExecContext(ctx, `UPDATE media_items SET updated_at = COALESCE(last_updated, CURRENT_TIMESTAMP) WHERE updated_at IS NULL`); err != nil {
			return fmt.Errorf("sqlite: backfill media_items.updated_at: %w", err)
		}
	}
	return nil
}

func (db *DB) addMediaItemsFavoriteColumnPostgres(ctx context.Context) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE media_items ADD COLUMN IF NOT EXISTS is_favorite BOOLEAN NOT NULL DEFAULT FALSE`); err != nil {
		return fmt.Errorf("postgres: add media_items.is_favorite: %w", err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE media_items ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`); err != nil {
		return fmt.Errorf("postgres: add media_items.updated_at: %w", err)
	}
	return nil
}

// columnExistsSQLite asks SQLite's schema catalog whether the given
// column exists on the given table. SQLite lacks an "ADD COLUMN IF NOT
// EXISTS" form, so we probe before the ALTER.
func (db *DB) columnExistsSQLite(ctx context.Context, table, column string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid      int
			name     string
			ctype    string
			notnull  int
			dfltVal  any
			pk       int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltVal, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, rows.Err()
		}
	}
	return false, rows.Err()
}

package database

import (
	"context"
	"fmt"
)

// createPerformanceIndexes adds performance-critical indexes identified from
// repository query patterns. This migration is idempotent (uses IF NOT EXISTS).
//
// Targeted tables and columns:
//   - files: file_type, extension, is_directory, name — frequently filtered in
//     SearchFiles, directory listing, and stats queries.
//   - media_items: (title, media_type_id) compound — used by GetByTitle and
//     GetDuplicates which always query both columns together.
//   - media_items: status, year — filtered in search and duplicate detection.
//   - user_metadata: (user_id, watched_status) compound — filtered in watched
//     media queries.
//   - media_files: (media_item_id, file_id) UNIQUE — prevents duplicate
//     file-to-entity links and accelerates junction lookups.
func (db *DB) createPerformanceIndexes(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createPerformanceIndexesPostgres(ctx)
	}
	return db.createPerformanceIndexesSQLite(ctx)
}

func (db *DB) createPerformanceIndexesSQLite(ctx context.Context) error {
	indexes := []string{
		// files table: columns frequently used in WHERE clauses
		`CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type)`,
		`CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension)`,
		`CREATE INDEX IF NOT EXISTS idx_files_is_directory ON files(is_directory)`,
		`CREATE INDEX IF NOT EXISTS idx_files_name ON files(name)`,

		// media_items: compound index for title+type lookups (GetByTitle, GetDuplicates)
		`CREATE INDEX IF NOT EXISTS idx_media_items_title_type ON media_items(title, media_type_id)`,

		// media_items: status and year columns used in filtering and duplicate detection
		`CREATE INDEX IF NOT EXISTS idx_media_items_status ON media_items(status)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_year ON media_items(year)`,

		// user_metadata: compound index for user+watched_status queries
		`CREATE INDEX IF NOT EXISTS idx_user_metadata_user_watched ON user_metadata(user_id, watched_status)`,
	}

	for _, query := range indexes {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to create performance index: %s, error: %w", query, err)
		}
	}

	// media_files: unique compound index prevents duplicate file-entity links.
	// Remove duplicates first (keep lowest rowid per pair) then create the index.
	if _, err := db.ExecContext(ctx,
		`DELETE FROM media_files WHERE rowid NOT IN (
			SELECT MIN(rowid) FROM media_files GROUP BY media_item_id, file_id
		)`); err != nil {
		return fmt.Errorf("failed to deduplicate media_files: %w", err)
	}
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_media_files_item_file ON media_files(media_item_id, file_id)`); err != nil {
		return fmt.Errorf("failed to create media_files unique index: %w", err)
	}

	return nil
}

func (db *DB) createPerformanceIndexesPostgres(ctx context.Context) error {
	indexes := []string{
		// files table: columns frequently used in WHERE clauses
		`CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type)`,
		`CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension)`,
		`CREATE INDEX IF NOT EXISTS idx_files_is_directory ON files(is_directory)`,
		`CREATE INDEX IF NOT EXISTS idx_files_name ON files(name)`,

		// media_items: compound index for title+type lookups (GetByTitle, GetDuplicates)
		`CREATE INDEX IF NOT EXISTS idx_media_items_title_type ON media_items(title, media_type_id)`,

		// media_items: status and year columns used in filtering and duplicate detection
		`CREATE INDEX IF NOT EXISTS idx_media_items_status ON media_items(status)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_year ON media_items(year)`,

		// user_metadata: compound index for user+watched_status queries
		`CREATE INDEX IF NOT EXISTS idx_user_metadata_user_watched ON user_metadata(user_id, watched_status)`,
	}

	for _, stmt := range indexes {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to create performance index: %s, error: %w", stmt, err)
		}
	}

	// media_files: unique compound index prevents duplicate file-entity links.
	// Remove duplicates first (keep lowest ctid per pair) then create the index.
	if _, err := db.ExecContext(ctx,
		`DELETE FROM media_files a USING media_files b
		 WHERE a.ctid > b.ctid
		   AND a.media_item_id = b.media_item_id
		   AND a.file_id = b.file_id`); err != nil {
		return fmt.Errorf("failed to deduplicate media_files: %w", err)
	}
	if _, err := db.ExecContext(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_media_files_item_file ON media_files(media_item_id, file_id)`); err != nil {
		return fmt.Errorf("failed to create media_files unique index: %w", err)
	}

	return nil
}

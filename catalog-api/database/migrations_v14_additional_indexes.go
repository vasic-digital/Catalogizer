package database

import (
	"context"
	"fmt"
)

// createAdditionalIndexes adds indexes for frequently queried columns
// to improve query performance for common operations.
func (db *DB) createV14AdditionalIndexes(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createV14AdditionalIndexesPostgres(ctx)
	}
	return db.createV14AdditionalIndexesSQLite(ctx)
}

func (db *DB) createV14AdditionalIndexesSQLite(ctx context.Context) error {
	indexes := []string{
		// Files table - common query patterns
		`CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_files_modified_at ON files(modified_at)`,
		`CREATE INDEX IF NOT EXISTS idx_files_size ON files(size)`,
		`CREATE INDEX IF NOT EXISTS idx_files_is_directory ON files(is_directory)`,
		`CREATE INDEX IF NOT EXISTS idx_files_storage_root_created ON files(storage_root_id, created_at)`,

		// File metadata - fast lookups
		`CREATE INDEX IF NOT EXISTS idx_file_metadata_key ON file_metadata(key)`,
		`CREATE INDEX IF NOT EXISTS idx_file_metadata_key_value ON file_metadata(key, value)`,

		// Analytics - time-series queries
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_time ON analytics_events(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_user ON analytics_events(user_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type, timestamp)`,

		// Scan history - recent scans
		`CREATE INDEX IF NOT EXISTS idx_scan_history_start_time ON scan_history(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_scan_history_status ON scan_history(status)`,

		// Media items
		`CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(media_type_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_title ON media_items(title)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_status ON media_items(status)`,

		// User sessions - cleanup queries
		`CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active)`,
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}

func (db *DB) createV14AdditionalIndexesPostgres(ctx context.Context) error {
	indexes := []string{
		// Files table - common query patterns
		`CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_files_modified_at ON files(modified_at)`,
		`CREATE INDEX IF NOT EXISTS idx_files_size ON files(size)`,
		`CREATE INDEX IF NOT EXISTS idx_files_is_directory ON files(is_directory)`,
		`CREATE INDEX IF NOT EXISTS idx_files_storage_root_created ON files(storage_root_id, created_at)`,

		// File metadata - fast lookups
		`CREATE INDEX IF NOT EXISTS idx_file_metadata_key ON file_metadata(key)`,
		`CREATE INDEX IF NOT EXISTS idx_file_metadata_key_value ON file_metadata(key, value)`,

		// Analytics - time-series queries
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_time ON analytics_events(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_user ON analytics_events(user_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type, timestamp)`,

		// Scan history - recent scans
		`CREATE INDEX IF NOT EXISTS idx_scan_history_start_time ON scan_history(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_scan_history_status ON scan_history(status)`,

		// Media items
		`CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(media_type_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_title ON media_items(title)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_status ON media_items(status)`,

		// User sessions - cleanup queries
		`CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active)`,
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}

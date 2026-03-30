package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// createServiceTables (v11) — SQLite
// ---------------------------------------------------------------------------

func TestCreateServiceTables_SQLite(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Auth tables needed for FK references (users)
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createServiceTables(ctx)
	require.NoError(t, err)

	expectedTables := []string{
		"favorites",
		"favorite_categories",
		"analytics_events",
		"media_access_logs",
		"error_reports",
		"crash_reports",
		"log_collections",
		"log_shares",
		"cache_entries",
		"api_cache",
		"media_metadata_cache",
		"analytics_reports",
		"wizard_progress",
	}

	for _, table := range expectedTables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		assert.NoError(t, err, "checking table %s", table)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}
}

func TestCreateServiceTables_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createServiceTables(ctx)
	require.NoError(t, err)

	// Second call should succeed (IF NOT EXISTS)
	err = db.createServiceTables(ctx)
	assert.NoError(t, err)
}

func TestCreateServiceTables_Indexes(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createServiceTables(ctx)
	require.NoError(t, err)

	expectedIndexes := []string{
		"idx_favorites_user",
		"idx_favorites_entity",
		"idx_analytics_events_user",
		"idx_analytics_events_type",
		"idx_analytics_events_date",
		"idx_media_access_user",
		"idx_media_access_media",
		"idx_error_reports_user",
		"idx_error_reports_status",
		"idx_error_reports_level",
		"idx_crash_reports_user",
		"idx_log_collections_user",
		"idx_cache_entries_key",
		"idx_cache_entries_expires",
	}

	for _, idx := range expectedIndexes {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&count)
		assert.NoError(t, err, "checking index %s", idx)
		assert.Equal(t, 1, count, "index %s should exist", idx)
	}
}

func TestCreateServiceTables_CanInsertFavorite(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createServiceTables(ctx)
	require.NoError(t, err)

	// Insert a user
	_, err = db.ExecContext(ctx,
		"INSERT INTO users (username, email, password_hash, salt, role_id) VALUES (?, ?, ?, ?, ?)",
		"favuser", "fav@example.com", "hash", "salt", 1)
	require.NoError(t, err)

	// Insert a favorite
	id, err := db.InsertReturningID(ctx,
		"INSERT INTO favorites (user_id, entity_type, entity_id) VALUES (?, ?, ?)",
		1, "movie", 42)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Query it back
	var entityType string
	err = db.QueryRowContext(ctx, "SELECT entity_type FROM favorites WHERE id = ?", id).Scan(&entityType)
	assert.NoError(t, err)
	assert.Equal(t, "movie", entityType)
}

func TestCreateServiceTables_CanInsertCacheEntry(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createServiceTables(ctx)
	require.NoError(t, err)

	id, err := db.InsertReturningID(ctx,
		"INSERT INTO cache_entries (cache_key, cache_value, cache_type) VALUES (?, ?, ?)",
		"test-key", `{"data": "value"}`, "general")
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

// ---------------------------------------------------------------------------
// fixServiceTableColumns (v12)
// ---------------------------------------------------------------------------

func TestFixServiceTableColumns_AfterV11(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createServiceTables(ctx)
	require.NoError(t, err)

	// fixServiceTableColumns should not fail when columns already exist from v11
	err = db.fixServiceTableColumns(ctx)
	assert.NoError(t, err)
}

func TestFixServiceTableColumns_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createServiceTables(ctx)
	require.NoError(t, err)

	// Call twice; should be safe
	err = db.fixServiceTableColumns(ctx)
	assert.NoError(t, err)

	err = db.fixServiceTableColumns(ctx)
	assert.NoError(t, err)
}

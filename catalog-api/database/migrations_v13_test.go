package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// createPlaylistTables (v13) — SQLite
// ---------------------------------------------------------------------------

func TestCreatePlaylistTables_SQLite(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Auth tables needed for FK references (users)
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createPlaylistTables(ctx)
	require.NoError(t, err)

	expectedTables := []string{"playlists", "playlist_items"}
	for _, table := range expectedTables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		assert.NoError(t, err, "checking table %s", table)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}
}

func TestCreatePlaylistTables_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createPlaylistTables(ctx)
	require.NoError(t, err)

	// Second call should succeed (IF NOT EXISTS)
	err = db.createPlaylistTables(ctx)
	assert.NoError(t, err)
}

func TestCreatePlaylistTables_Indexes(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createPlaylistTables(ctx)
	require.NoError(t, err)

	expectedIndexes := []string{
		"idx_playlists_user_id",
		"idx_playlist_items_playlist_id",
		"idx_playlist_items_entity",
	}

	for _, idx := range expectedIndexes {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&count)
		assert.NoError(t, err, "checking index %s", idx)
		assert.Equal(t, 1, count, "index %s should exist", idx)
	}
}

func TestCreatePlaylistTables_CanInsertAndQuery(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createPlaylistTables(ctx)
	require.NoError(t, err)

	// Insert a user for FK
	_, err = db.ExecContext(ctx,
		"INSERT INTO users (username, email, password_hash, salt, role_id) VALUES (?, ?, ?, ?, ?)",
		"pluser", "pl@example.com", "hash", "salt", 1)
	require.NoError(t, err)

	// Insert a playlist
	plID, err := db.InsertReturningID(ctx,
		"INSERT INTO playlists (user_id, name, description) VALUES (?, ?, ?)",
		1, "Test Playlist", "A description")
	require.NoError(t, err)
	assert.Greater(t, plID, int64(0))

	// Insert a playlist item
	itemID, err := db.InsertReturningID(ctx,
		"INSERT INTO playlist_items (playlist_id, entity_id, entity_type, position) VALUES (?, ?, ?, ?)",
		plID, 42, "movie", 0)
	require.NoError(t, err)
	assert.Greater(t, itemID, int64(0))

	// Query back
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM playlists WHERE id = ?", plID).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "Test Playlist", name)

	var entityType string
	err = db.QueryRowContext(ctx, "SELECT entity_type FROM playlist_items WHERE id = ?", itemID).Scan(&entityType)
	assert.NoError(t, err)
	assert.Equal(t, "movie", entityType)
}

func TestCreatePlaylistTables_ForeignKeyConstraint(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	// Enable foreign keys for this test
	_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	err = db.createPlaylistTables(ctx)
	require.NoError(t, err)

	// Insert a user
	_, err = db.ExecContext(ctx,
		"INSERT INTO users (username, email, password_hash, salt, role_id) VALUES (?, ?, ?, ?, ?)",
		"fkuser", "fk@example.com", "hash", "salt", 1)
	require.NoError(t, err)

	// Insert a playlist
	plID, err := db.InsertReturningID(ctx,
		"INSERT INTO playlists (user_id, name) VALUES (?, ?)", 1, "FK Test")
	require.NoError(t, err)

	// Insert an item referencing the playlist — should succeed
	_, err = db.InsertReturningID(ctx,
		"INSERT INTO playlist_items (playlist_id, entity_id, entity_type) VALUES (?, ?, ?)",
		plID, 1, "song")
	assert.NoError(t, err)

	// Insert an item referencing a non-existent playlist — should fail with FK enabled
	_, err = db.InsertReturningID(ctx,
		"INSERT INTO playlist_items (playlist_id, entity_id, entity_type) VALUES (?, ?, ?)",
		9999, 1, "song")
	assert.Error(t, err, "foreign key constraint should prevent insert with invalid playlist_id")
}

// ---------------------------------------------------------------------------
// Full migration includes v13
// ---------------------------------------------------------------------------

func TestFullMigration_IncludesPlaylistTables(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Verify playlist tables exist after full migration
	for _, table := range []string{"playlists", "playlist_items"} {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		assert.NoError(t, err, "checking table %s", table)
		assert.Equal(t, 1, count, "table %s should exist after full migration", table)
	}

	// Verify v13 migration was recorded
	var exists int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE version = 13").Scan(&exists)
	assert.NoError(t, err)
	assert.Equal(t, 1, exists, "migration v13 should be recorded")
}

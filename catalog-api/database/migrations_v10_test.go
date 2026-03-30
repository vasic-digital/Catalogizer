package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// createSyncTables (v10) — SQLite
// ---------------------------------------------------------------------------

func TestCreateSyncTables_SQLite(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	// Need users table for foreign key references
	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createSyncTables(ctx)
	require.NoError(t, err)

	// Verify tables exist
	tables := []string{"sync_endpoints", "sync_sessions", "sync_schedules"}
	for _, table := range tables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		assert.NoError(t, err, "checking table %s", table)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}
}

func TestCreateSyncTables_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createSyncTables(ctx)
	require.NoError(t, err)

	// Second call should succeed (IF NOT EXISTS)
	err = db.createSyncTables(ctx)
	assert.NoError(t, err)
}

func TestCreateSyncTables_Indexes(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createSyncTables(ctx)
	require.NoError(t, err)

	// Verify indexes
	expectedIndexes := []string{
		"idx_sync_endpoints_user_id",
		"idx_sync_endpoints_status",
		"idx_sync_endpoints_type",
		"idx_sync_sessions_endpoint_id",
		"idx_sync_sessions_user_id",
		"idx_sync_sessions_status",
		"idx_sync_sessions_started_at",
		"idx_sync_schedules_endpoint_id",
		"idx_sync_schedules_user_id",
		"idx_sync_schedules_is_active",
		"idx_sync_schedules_next_run",
	}

	for _, idx := range expectedIndexes {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&count)
		assert.NoError(t, err, "checking index %s", idx)
		assert.Equal(t, 1, count, "index %s should exist", idx)
	}
}

func TestCreateSyncTables_CanInsertAndQuery(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createSyncTables(ctx)
	require.NoError(t, err)

	// Insert a user first (for FK)
	_, err = db.ExecContext(ctx,
		"INSERT INTO users (username, email, password_hash, salt, role_id) VALUES (?, ?, ?, ?, ?)",
		"syncuser", "sync@example.com", "hash", "salt", 1)
	require.NoError(t, err)

	// Insert a sync endpoint
	id, err := db.InsertReturningID(ctx,
		"INSERT INTO sync_endpoints (user_id, name, type, url) VALUES (?, ?, ?, ?)",
		1, "My NAS", "smb", "smb://nas.local/share")
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Query it back
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM sync_endpoints WHERE id = ?", id).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "My NAS", name)
}
